package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	_ "embed"

	"gopkg.in/yaml.v2"
)

//go:embed config.yaml
var embeddedConfigFile []byte
var configFilepath string = ".shell-ai"

type Config struct {
	APIKeys     map[string]string `yaml:"api_keys"`
	Models      []Model           `yaml:"models"`
	Preferences Preferences       `yaml:"preferences"`
}

type Model struct {
	Name     string    `yaml:"name"`
	Endpoint string    `yaml:"endpoint"`
	APIKey   string    `yaml:"api_key"`
	Prompt   []Message `yaml:"prompt"`
}

type Message struct {
	Role    string `yaml:"role"`
	Content string `yaml:"content"`
}

type Preferences struct {
	DefaultModel string `yaml:"default_model"`
}

func fullFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("Error getting home directory: %s", err)
	}
	configFilepath := filepath.Join(homeDir, configFilepath)
	return configFilepath, nil
}

func LoadConfig() (config Config, err error) {
	filepath, err := fullFilePath()
	if err != nil {
		return config, fmt.Errorf("Error getting config file path: %s", err)
	}

	// if file doesn't exist, create it with defaults
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		err = ioutil.WriteFile(filepath, embeddedConfigFile, 0644)
	}

	yamlFile, err := ioutil.ReadFile(filepath)
	if err != nil {
		return config, fmt.Errorf("Error reading config file: %s", err)
	}
	err = yaml.Unmarshal(yamlFile, &config)
	return
}
