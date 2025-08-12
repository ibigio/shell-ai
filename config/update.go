package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// UpdateDefaultModel updates the default model in the current config file
// preserving the existing config format (v1 or v2) and all other fields.
func UpdateDefaultModel(model string) error {
	abs, err := FullFilePath(configFilePath)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(abs)
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}
	var v struct {
		Version string `yaml:"config_format_version"`
	}
	if err := yaml.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("yaml parse: %w", err)
	}
	if v.Version == "2" {
		var cfg AppConfigV2
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return fmt.Errorf("v2 unmarshal: %w", err)
		}
		cfg.Preferences.Model = model
		return writeYAML(abs, cfg)
	}
	// default to v1
	var cfg AppConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("v1 unmarshal: %w", err)
	}
	cfg.Preferences.DefaultModel = model
	return writeYAML(abs, cfg)
}

func writeYAML(path string, v interface{}) error {
	b, err := yaml.Marshal(v)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, b, 0644)
}
