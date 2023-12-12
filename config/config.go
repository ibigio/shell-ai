package config

import (
	"fmt"
	"os"
	"path/filepath"
	. "q/types"

	_ "embed"

	"gopkg.in/yaml.v2"
)

type AppConfig struct {
	Models      []ModelConfig `yaml:"models"`
	Preferences Preferences   `yaml:"preferences"`
	Version     string        `yaml:"version"`
}

//go:embed config.yaml
var embeddedConfigFile []byte
var configFilePath string = ".shell-ai/config.yaml"

func FullFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error getting home directory: %s", err)
	}
	configFilePath := filepath.Join(homeDir, configFilePath)
	return configFilePath, nil
}

func LoadAppConfig() (config AppConfig, err error) {
	filePath, err := FullFilePath()
	if err != nil {
		return config, fmt.Errorf("error getting config file path: %s", err)
	}

	// if file doesn't exist, create it with defaults
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return createConfigWithDefaults(filePath)
	}
	return loadExistingConfig(filePath)
}

func createConfigWithDefaults(filePath string) (AppConfig, error) {
	config := AppConfig{}
	err := yaml.Unmarshal(embeddedConfigFile, &config)
	if err != nil {
		return config, fmt.Errorf("error unmarshalling embedded config: %s", err)
	}

	config.Preferences.DefaultModel = "gpt-4"

	// set default model to legacy option (for backwards compat)
	modelOverride := os.Getenv("OPENAI_MODEL_OVERRIDE")
	if modelOverride != "" {
		config.Preferences.DefaultModel = modelOverride
	}

	modifiedConfig, err := yaml.Marshal(config)
	if err != nil {
		return config, fmt.Errorf("error marshalling modified config: %s", err)
	}

	// Create all directories in the filepath
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return config, fmt.Errorf("error creating directories: %s", err)
	}

	err = os.WriteFile(filePath, modifiedConfig, 0644)
	if err != nil {
		return config, fmt.Errorf("error writing modified config to file: %s", err)
	}

	return config, nil
}

func loadExistingConfig(filePath string) (AppConfig, error) {
	config := AppConfig{}
	yamlFile, err := os.ReadFile(filePath)
	if err != nil {
		return config, fmt.Errorf("error reading config file: %s", err)
	}
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return config, fmt.Errorf("error unmarshalling config file: %s", err)
	}

	return config, nil
}
