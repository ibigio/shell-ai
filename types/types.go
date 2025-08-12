package types

type ModelConfig struct {
	ModelName string    `yaml:"name"`
	Endpoint  string    `yaml:"endpoint"`
	Auth      string    `yaml:"auth_env_var"`
	OrgID     string    `yaml:"org_env_var,omitempty"`
	Prompt    []Message `yaml:"prompt"`
}

type Message struct {
	Role    string `yaml:"role" json:"role"`
	Content string `yaml:"content" json:"content"`
}

type Preferences struct {
	DefaultModel string `yaml:"default_model"`
}

type Payload struct {
	Model       string    `json:"model"`
	Prompt      string    `json:"prompt,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float32   `json:"temperature,omitempty"`
	Messages    []Message `json:"messages"`
	Stream      bool      `json:"stream,omitempty"`
}

type ResponseData struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	Model   string `json:"model"`
	Usage   struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		Index        int    `json:"index"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

// === Modern v2 config and unified runtime types ===

// AuthSettings represents provider-level auth resolution for v2.
type AuthSettings struct {
	Type      string `yaml:"type"` // "bearer" or "header"
	Header    string `yaml:"header,omitempty"`
	TokenEnv  string `yaml:"token_env,omitempty"`
	OrgEnv    string `yaml:"org_env,omitempty"`
	OrgHeader string `yaml:"org_header,omitempty"`
}

// ProviderSettings captures provider endpoints and behavior for v2.
type ProviderSettings struct {
	PayloadStyle string            `yaml:"payload_style"`
	Endpoint     string            `yaml:"endpoint"`
	Auth         AuthSettings      `yaml:"auth"`
	Headers      map[string]string `yaml:"headers"`
	QueryParams  map[string]string `yaml:"query_params"`
}

// ModelDefinition represents a model entry in v2 configs.
type ModelDefinition struct {
	Name            string    `yaml:"name"`
	Config          string    `yaml:"config,omitempty"` // provider key override
	Temperature     float32   `yaml:"temperature,omitempty"`
	ReasoningEffort string    `yaml:"reasoning_effort,omitempty"`
	Prompt          []Message `yaml:"prompt"`
}

// ResolvedModelConfig is the unified runtime configuration used by the LLM client and adapters.
type ResolvedModelConfig struct {
	Name            string
	Endpoint        string
	PayloadStyle    string
	Headers         map[string]string
	QueryParams     map[string]string
	Temperature     float32
	ReasoningEffort string
	Prompt          []Message
}
