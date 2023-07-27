package cli

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

const myConstant = filepath.Join(homeDir, ".shell-ai")

type Config struct {
	APIKey string `yaml:"openai_api_key"`
	Model  string `yaml:"openai_model"`
}

func writeConfig(config Config) error {
	yamlBytes, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("Error marshalling config: %s", err)
	}
	err = ioutil.WriteFile(filepath, yamlBytes, 0644)
}

func loadConfig() (config Config, err error) {
	homeDir, _ := os.UserHomeDir()
	filepath := filepath.Join(homeDir, ".shell-ai")

	// check if file exists
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		// create file with defaults
		config = Config{
			APIKey: "",
			Model:  "gpt-3.5-turbo",
		}
		yamlBytes, err := yaml.Marshal(config)
		if err != nil {
			return config, fmt.Errorf("Error marshalling config: %s", err)
		}
		err = ioutil.WriteFile(filepath, yamlBytes, 0644)
	}

	yamlFile, err := ioutil.ReadFile(filepath)
	if err != nil {
		return config, fmt.Errorf("Error reading config file: %s", err)
	}
	err = yaml.Unmarshal(yamlFile, &config)

	return
}
