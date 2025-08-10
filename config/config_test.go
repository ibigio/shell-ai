package config

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"gopkg.in/yaml.v2"
)

func TestV1_Unmarshal_PrimaryFixture(t *testing.T) {
	data, err := ioutil.ReadFile("testdata/v1.yaml")
	if err != nil {
		t.Fatalf("read v1.yaml: %v", err)
	}
	var cfg AppConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("unmarshal v1: %v", err)
	}

	if got, want := cfg.Version, "1"; got != want {
		t.Fatalf("version mismatch: got %q want %q", got, want)
	}
	if got, want := cfg.Preferences.DefaultModel, "gpt-5"; got != want {
		t.Fatalf("default model mismatch: got %q want %q", got, want)
	}
	if len(cfg.Models) < 2 {
		t.Fatalf("expected at least 2 models, got %d", len(cfg.Models))
	}
	if cfg.Models[0].ModelName == "" {
		t.Fatalf("first model name empty")
	}
	if len(cfg.Models[0].Prompt) == 0 || cfg.Models[0].Prompt[0].Role != "system" {
		b, _ := json.Marshal(cfg.Models[0])
		t.Fatalf("expected first model to have system prompt, got: %s", string(b))
	}
}

func TestUnifiedParser_V1(t *testing.T) {
	t.Skip("TODO: implement unified v1->modern parser and unskip")
}

func TestUnifiedParser_V2(t *testing.T) {
	t.Skip("no v2 tests yet")
}

func TestV1_Unmarshal_Variants(t *testing.T) {
	cases := []struct {
		name    string
		file    string
		wantErr bool
	}{
		{"ok_minimal", "testdata/v1_ok_minimal.yaml", false},
		{"ok_no_prompt", "testdata/v1_ok_no_prompt.yaml", false},
		{"fail_prompt_scalar", "testdata/v1_fail_prompt_scalar.yaml", true},
		{"fail_prompt_map", "testdata/v1_fail_prompt_map.yaml", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := ioutil.ReadFile(tc.file)
			if err != nil {
				t.Fatalf("read %s: %v", tc.file, err)
			}
			var cfg AppConfig
			err = yaml.Unmarshal(data, &cfg)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for %s, got nil", tc.name)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for %s: %v", tc.name, err)
			}
		})
	}
}
