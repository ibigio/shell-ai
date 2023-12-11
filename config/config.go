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
var configFilepath string = ".shell-ai"

func FullFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error getting home directory: %s", err)
	}
	configFilepath := filepath.Join(homeDir, configFilepath)
	return configFilepath, nil
}

func LoadAppConfig() (config AppConfig, err error) {
	filepath, err := FullFilePath()
	if err != nil {
		return config, fmt.Errorf("error getting config file path: %s", err)
	}

	// if file doesn't exist, create it with defaults
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return createConfigWithDefaults(filepath)
	}
	return loadExistingConfig(filepath)
}

func createConfigWithDefaults(filepath string) (AppConfig, error) {
	config := AppConfig{}
	err := yaml.Unmarshal(embeddedConfigFile, &config)
	if err != nil {
		return config, fmt.Errorf("error unmarshalling embedded config: %s", err)
	}

	modelOverride := os.Getenv("OPENAI_MODEL_OVERRIDE")
	if modelOverride == "" {
		modelOverride = "gpt-4"
	}
	config.Preferences.DefaultModel = modelOverride

	modifiedConfig, err := yaml.Marshal(config)
	if err != nil {
		return config, fmt.Errorf("error marshalling modified config: %s", err)
	}

	err = os.WriteFile(filepath, modifiedConfig, 0644)
	if err != nil {
		return config, fmt.Errorf("error writing modified config to file: %s", err)
	}

	return config, nil
}

func loadExistingConfig(filepath string) (AppConfig, error) {
	config := AppConfig{}
	yamlFile, err := os.ReadFile(filepath)
	if err != nil {
		return config, fmt.Errorf("error reading config file: %s", err)
	}
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return config, fmt.Errorf("error unmarshalling config file: %s", err)
	}

	return config, nil
}
