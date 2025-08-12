package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v2"
)

func TestUpdateDefaultModel_V2_PreservesShape(t *testing.T) {
	_, restore := withTempConfigPaths(t)
	defer restore()

	yamlText := `
system_prompt: |
  Hello.
preferences:
  model: gpt-5
models:
  - name: gpt-5
  - name: gpt-4.1
config_format_version: "2"
`
	abs, _ := FullFilePath(configFilePath)
	os.MkdirAll(filepath.Dir(abs), 0o755)
	if err := ioutil.WriteFile(abs, []byte(yamlText), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := UpdateDefaultModel("gpt-4.1"); err != nil {
		t.Fatalf("UpdateDefaultModel: %v", err)
	}
	var got AppConfigV2
	b, _ := ioutil.ReadFile(abs)
	if err := yaml.Unmarshal(b, &got); err != nil {
		t.Fatalf("unmarshal back: %v", err)
	}
	if got.Preferences.Model != "gpt-4.1" || got.SystemPrompt == "" || len(got.Models) != 2 || got.Version != "2" {
		t.Fatalf("unexpected v2 after update: %+v", got)
	}
}

func TestUpdateDefaultModel_V1_PreservesShape(t *testing.T) {
	_, restore := withTempConfigPaths(t)
	defer restore()

	data, err := ioutil.ReadFile("testdata/v1_ok_minimal.yaml")
	if err != nil {
		t.Skip("fixture not available")
	}
	abs, _ := FullFilePath(configFilePath)
	os.MkdirAll(filepath.Dir(abs), 0o755)
	ioutil.WriteFile(abs, data, 0644)

	if err := UpdateDefaultModel("gpt-4.1"); err != nil {
		t.Fatalf("UpdateDefaultModel: %v", err)
	}
	var got AppConfig
	b, _ := ioutil.ReadFile(abs)
	if err := yaml.Unmarshal(b, &got); err != nil {
		t.Fatalf("unmarshal back: %v", err)
	}
	if got.Preferences.DefaultModel != "gpt-4.1" || got.Version != "1" {
		t.Fatalf("unexpected v1 after update: %+v", got)
	}
}
