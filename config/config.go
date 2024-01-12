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
	Version     string        `yaml:"config_format_version"`
}

// //go:embed config.yaml
// var embeddedConfigFile []byte

//go:embed config.yaml
var embeddedConfigFile []byte
var configFilePath string = ".shell-ai/config.yaml"
var backupConfigFilePath string = ".shell-ai/.backup-config.yaml"

func FullFilePath(relativeFilePath string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error getting home directory: %s", err)
	}
	configFilePath := filepath.Join(homeDir, relativeFilePath)
	return configFilePath, nil
}

func LoadAppConfig() (config AppConfig, err error) {
	filePath, err := FullFilePath(configFilePath)
	if err != nil {
		return config, fmt.Errorf("error getting config file path: %s", err)
	}

	// if file doesn't exist, create it with defaults
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return createConfigWithDefaults(filePath)
	}
	return loadExistingConfig(filePath)
}

func SaveAppConfig(config AppConfig) error {
	return writeConfigToFile(config)
}

func ResetAppConfigToDefault() error {
	_, err := createConfigWithDefaults(configFilePath)
	return err
}

func RevertAppConfigToBackup() error {
	fullConfigPath, _ := FullFilePath(configFilePath)
	fullBackupConfigPath, _ := FullFilePath(backupConfigFilePath)

	// delete the file if it exists
	if err := os.Remove(fullConfigPath); !os.IsNotExist(err) && err != nil {
		return err
	}
	config, err := loadExistingConfig(fullBackupConfigPath)
	if err != nil {
		return err
	}
	return writeConfigToFile(config)
}

func createConfigWithDefaults(filePath string) (AppConfig, error) {
	config := AppConfig{}
	err := yaml.Unmarshal(embeddedConfigFile, &config)
	if err != nil {
		return config, fmt.Errorf("error unmarshalling embedded config: %s", err)
	}
	// set default model to legacy option (for backwards compat)
	modelOverride := os.Getenv("OPENAI_MODEL_OVERRIDE")
	if modelOverride != "" {
		config.Preferences.DefaultModel = modelOverride
	}

	return config, writeConfigToFile(config)
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

func SaveBackupConfig(config AppConfig) error {
	filePath, err := FullFilePath(backupConfigFilePath)
	if err != nil {
		return err
	}
	configData, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("error marshalling config: %s", err)
	}

	err = os.WriteFile(filePath, configData, 0644)
	if err != nil {
		return fmt.Errorf("error writing config to file: %s", err)
	}
	return nil
}

func writeConfigToFile(config AppConfig) error {
	filePath, _ := FullFilePath(configFilePath)
	// Create all directories in the filepath
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creating directories: %s", err)
	}
	configData, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("error marshalling config: %s", err)
	}

	err = os.WriteFile(filePath, configData, 0644)
	if err != nil {
		return fmt.Errorf("error writing config to file: %s", err)
	}
	SaveBackupConfig(config)
	return nil
}
