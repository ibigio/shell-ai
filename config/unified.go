package config

import (
	"fmt"
	"net/url"
	"os"
	types "q/types"
	"strings"

	"gopkg.in/yaml.v2"
)

// V2 schema types (YAML-facing)
type PreferencesV2 struct {
	Model           string `yaml:"model"`
	Config          string `yaml:"config"`
	ReasoningEffort string `yaml:"reasoning_effort,omitempty"`
}

type AppConfigV2 struct {
	Preferences  PreferencesV2                     `yaml:"preferences"`
	Models       []types.ModelDefinition           `yaml:"models"`
	Configs      map[string]types.ProviderSettings `yaml:"configs"`
	Version      string                            `yaml:"config_format_version"`
	SystemPrompt string                            `yaml:"system_prompt"`
}

// ResolveSelectedModel reads the current config file (v1 or v2) and returns a unified, runtime-ready model config.
func ResolveSelectedModel() (types.ResolvedModelConfig, error) {
	filePath, err := FullFilePath(configFilePath)
	if err != nil {
		return types.ResolvedModelConfig{}, fmt.Errorf("config path: %w", err)
	}
	// If file doesn't exist, create defaults using existing helper
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if _, err := createConfigWithDefaults(filePath); err != nil {
			return types.ResolvedModelConfig{}, err
		}
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		return types.ResolvedModelConfig{}, fmt.Errorf("read config: %w", err)
	}
	// Detect version
	var v struct {
		Version string `yaml:"config_format_version"`
	}
	if err := yaml.Unmarshal(data, &v); err != nil {
		return types.ResolvedModelConfig{}, fmt.Errorf("yaml parse: %w", err)
	}
	if strings.TrimSpace(v.Version) == "2" {
		var cfg AppConfigV2
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return types.ResolvedModelConfig{}, fmt.Errorf("v2 unmarshal: %w", err)
		}
		return resolveFromV2(cfg)
	}
	// Default to v1
	var cfg AppConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return types.ResolvedModelConfig{}, fmt.Errorf("v1 unmarshal: %w", err)
	}
	return resolveFromV1(cfg)
}

func resolveFromV1(cfg AppConfig) (types.ResolvedModelConfig, error) {
	// pick default model
	var mdl *types.ModelConfig
	for i := range cfg.Models {
		if cfg.Models[i].ModelName == cfg.Preferences.DefaultModel {
			mdl = &cfg.Models[i]
			break
		}
	}
	if mdl == nil {
		if len(cfg.Models) == 0 {
			return types.ResolvedModelConfig{}, fmt.Errorf("no models configured")
		}
		mdl = &cfg.Models[0]
	}
	token := os.Getenv(mdl.Auth)
	if token == "" {
		return types.ResolvedModelConfig{}, fmt.Errorf("%s environment variable not set", mdl.Auth)
	}
	headers := map[string]string{}
	// bearer token default
	headers["Authorization"] = "Bearer " + token
	if mdl.OrgID != "" {
		if org := os.Getenv(mdl.OrgID); org != "" {
			headers["OpenAI-Organization"] = org
		}
	}
	return types.ResolvedModelConfig{
		Name:         mdl.ModelName,
		Endpoint:     mdl.Endpoint,
		PayloadStyle: "openai",
		Headers:      headers,
		QueryParams:  map[string]string{},
		Temperature:  0,
		Prompt:       append([]types.Message(nil), mdl.Prompt...),
	}, nil
}

func resolveFromV2(cfg AppConfigV2) (types.ResolvedModelConfig, error) {
	// select model
	var m *types.ModelDefinition
	for i := range cfg.Models {
		if cfg.Models[i].Name == cfg.Preferences.Model {
			m = &cfg.Models[i]
			break
		}
	}
	if m == nil {
		if len(cfg.Models) == 0 {
			return types.ResolvedModelConfig{}, fmt.Errorf("no models configured")
		}
		// fallback to first
		m = &cfg.Models[0]
	}
	// select provider config (default to openai if unspecified)
	providerKey := m.Config
	if providerKey == "" {
		providerKey = cfg.Preferences.Config
	}
	if providerKey == "" {
		providerKey = "openai"
	}
	// Start from built-in preset, merge user overrides if any
	preset, hasPreset := providerPreset(providerKey)
	var provider types.ProviderSettings
	if cfg.Configs != nil {
		if custom, ok := cfg.Configs[providerKey]; ok {
			if hasPreset {
				provider = mergeProviderSettings(preset, custom)
			} else {
				provider = custom
			}
		} else if hasPreset {
			provider = preset
		}
	} else if hasPreset {
		provider = preset
	}
	if provider.Endpoint == "" {
		return types.ResolvedModelConfig{}, fmt.Errorf("provider %s not found and no preset available", providerKey)
	}
	// resolve auth
	headers := map[string]string{}
	for k, v := range provider.Headers {
		headers[k] = v
	}
	token := ""
	if provider.Auth.TokenEnv != "" {
		token = os.Getenv(provider.Auth.TokenEnv)
		if token == "" {
			return types.ResolvedModelConfig{}, fmt.Errorf("%s environment variable not set", provider.Auth.TokenEnv)
		}
	}
	switch strings.ToLower(provider.Auth.Type) {
	case "bearer":
		if token != "" {
			headers["Authorization"] = "Bearer " + token
		}
	case "header":
		if provider.Auth.Header == "" {
			return types.ResolvedModelConfig{}, fmt.Errorf("header auth requires header name")
		}
		if token != "" {
			headers[provider.Auth.Header] = token
		}
	case "":
		// no auth
	default:
		return types.ResolvedModelConfig{}, fmt.Errorf("unknown auth type: %s", provider.Auth.Type)
	}
	if provider.Auth.OrgEnv != "" {
		if org := os.Getenv(provider.Auth.OrgEnv); org != "" {
			name := provider.Auth.OrgHeader
			if name == "" {
				name = "OpenAI-Organization"
			}
			headers[name] = org
		}
	}
	// endpoint interpolation and query params
	endpoint := provider.Endpoint
	endpoint = strings.ReplaceAll(endpoint, "{name}", url.PathEscape(m.Name))
	if len(provider.QueryParams) > 0 {
		// append query params
		sep := "?"
		if strings.Contains(endpoint, "?") {
			sep = "&"
		}
		q := url.Values{}
		for k, v := range provider.QueryParams {
			q.Set(k, v)
		}
		endpoint = endpoint + sep + q.Encode()
	}
	// Build prompt: optional top-level system prompt + model-level prompt
	prompt := make([]types.Message, 0, len(m.Prompt)+1)
	if strings.TrimSpace(cfg.SystemPrompt) != "" {
		prompt = append(prompt, types.Message{Role: "system", Content: cfg.SystemPrompt})
	}
	if len(m.Prompt) > 0 {
		prompt = append(prompt, m.Prompt...)
	}

	return types.ResolvedModelConfig{
		Name:            m.Name,
		Endpoint:        endpoint,
		PayloadStyle:    strings.ToLower(provider.PayloadStyle),
		Headers:         headers,
		QueryParams:     provider.QueryParams,
		Temperature:     m.Temperature,
		ReasoningEffort: m.ReasoningEffort,
		Prompt:          prompt,
	}, nil
}

// providerPreset returns built-in provider defaults for common providers.
func providerPreset(name string) (types.ProviderSettings, bool) {
	switch strings.ToLower(name) {
	case "openai":
		return types.ProviderSettings{
			PayloadStyle: "openai",
			Endpoint:     "https://api.openai.com/v1/chat/completions",
			Auth: types.AuthSettings{
				Type:      "bearer",
				TokenEnv:  "OPENAI_API_KEY",
				OrgEnv:    "OPENAI_ORG_ID",
				OrgHeader: "OpenAI-Organization",
			},
			Headers:     map[string]string{},
			QueryParams: map[string]string{},
		}, true
	case "anthropic":
		return types.ProviderSettings{
			PayloadStyle: "anthropic",
			Endpoint:     "https://api.anthropic.com/v1/messages",
			Auth: types.AuthSettings{
				Type:     "header",
				Header:   "x-api-key",
				TokenEnv: "ANTHROPIC_API_KEY",
			},
			Headers:     map[string]string{},
			QueryParams: map[string]string{},
		}, true
	case "ollama":
		return types.ProviderSettings{
			PayloadStyle: "ollama",
			Endpoint:     "http://localhost:11434/api/chat",
			Headers:      map[string]string{},
			QueryParams:  map[string]string{},
		}, true
	case "azure-openai":
		return types.ProviderSettings{
			PayloadStyle: "openai",
			Endpoint:     "https://example.openai.azure.com/openai/deployments/{name}/chat/completions",
			Auth: types.AuthSettings{
				Type:     "header",
				Header:   "Api-Key",
				TokenEnv: "AZURE_OPENAI_API_KEY",
			},
			Headers:     map[string]string{},
			QueryParams: map[string]string{"api-version": "2024-05-01-preview"},
		}, true
	default:
		return types.ProviderSettings{}, false
	}
}

func mergeProviderSettings(base, override types.ProviderSettings) types.ProviderSettings {
	out := base
	if override.PayloadStyle != "" {
		out.PayloadStyle = override.PayloadStyle
	}
	if override.Endpoint != "" {
		out.Endpoint = override.Endpoint
	}
	// Auth overrides: non-empty fields replace
	if override.Auth.Type != "" {
		out.Auth.Type = override.Auth.Type
	}
	if override.Auth.Header != "" {
		out.Auth.Header = override.Auth.Header
	}
	if override.Auth.TokenEnv != "" {
		out.Auth.TokenEnv = override.Auth.TokenEnv
	}
	if override.Auth.OrgEnv != "" {
		out.Auth.OrgEnv = override.Auth.OrgEnv
	}
	if override.Auth.OrgHeader != "" {
		out.Auth.OrgHeader = override.Auth.OrgHeader
	}
	// Merge maps
	if out.Headers == nil {
		out.Headers = map[string]string{}
	}
	for k, v := range override.Headers {
		out.Headers[k] = v
	}
	if out.QueryParams == nil {
		out.QueryParams = map[string]string{}
	}
	for k, v := range override.QueryParams {
		out.QueryParams[k] = v
	}
	return out
}
