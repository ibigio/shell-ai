package config

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	ollama "q/adapters/ollama"
	openai "q/adapters/openai"
)

func TestResolveSelectedModel_V1(t *testing.T) {
	_, restore := withTempConfigPaths(t)
	defer restore()

	// write v1 fixture to expected path
	data, err := ioutil.ReadFile("testdata/v1.yaml")
	if err != nil {
		t.Fatalf("read v1.yaml: %v", err)
	}
	abs, _ := FullFilePath(configFilePath)
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := ioutil.WriteFile(abs, data, 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	os.Setenv("OPENAI_API_KEY", "tok")
	os.Setenv("OPENAI_ORG_ID", "org")

	got, err := ResolveSelectedModel()
	if err != nil {
		t.Fatalf("resolve v1: %v", err)
	}
	if got.Name == "" || got.Endpoint == "" || got.PayloadStyle == "" {
		t.Fatalf("unexpected resolved: %+v", got)
	}
	if got.Headers["Authorization"] == "" {
		t.Fatalf("expected Authorization header")
	}
}

func TestResolveSelectedModel_V1_MissingEnv(t *testing.T) {
	_, restore := withTempConfigPaths(t)
	defer restore()

	data, _ := ioutil.ReadFile("testdata/v1.yaml")
	abs, _ := FullFilePath(configFilePath)
	os.MkdirAll(filepath.Dir(abs), 0o755)
	ioutil.WriteFile(abs, data, 0644)

	os.Unsetenv("OPENAI_API_KEY")

	_, err := ResolveSelectedModel()
	if err == nil || !strings.Contains(err.Error(), "OPENAI_API_KEY") {
		t.Fatalf("expected missing env var error, got %v", err)
	}
}

func TestResolveSelectedModel_V2_OpenAI(t *testing.T) {
	_, restore := withTempConfigPaths(t)
	defer restore()

	yaml := `
preferences:
  model: gpt-5
  config: openai
models:
  - name: gpt-5
configs:
  openai:
    payload_style: openai
    endpoint: https://api.openai.com/v1/chat/completions
    auth:
      type: bearer
      token_env: OPENAI_API_KEY
      org_env: OPENAI_ORG_ID
      org_header: OpenAI-Organization
    headers: {}
    query_params: {}
config_format_version: "2"
`
	abs, _ := FullFilePath(configFilePath)
	os.MkdirAll(filepath.Dir(abs), 0o755)
	ioutil.WriteFile(abs, []byte(yaml), 0644)

	os.Setenv("OPENAI_API_KEY", "tok")
	os.Setenv("OPENAI_ORG_ID", "org")

	got, err := ResolveSelectedModel()
	if err != nil {
		t.Fatalf("resolve v2: %v", err)
	}
	if got.Headers["Authorization"] == "" {
		t.Fatalf("expected bearer auth header for openai")
	}
}

func TestResolveSelectedModel_V2_MinimalDefaultsOpenAI(t *testing.T) {
	_, restore := withTempConfigPaths(t)
	defer restore()

	yaml := `
preferences:
  model: gpt-4.1
models:
  - name: gpt-4.1
config_format_version: "2"
`
	abs, _ := FullFilePath(configFilePath)
	os.MkdirAll(filepath.Dir(abs), 0o755)
	ioutil.WriteFile(abs, []byte(yaml), 0644)

	os.Setenv("OPENAI_API_KEY", "tok")
	os.Setenv("OPENAI_ORG_ID", "org")

	got, err := ResolveSelectedModel()
	if err != nil {
		t.Fatalf("resolve v2 minimal: %v", err)
	}
	if got.Endpoint != "https://api.openai.com/v1/chat/completions" || got.Headers["Authorization"] == "" {
		t.Fatalf("expected openai defaults, got endpoint=%s headers=%v", got.Endpoint, got.Headers)
	}
}

func TestResolveSelectedModel_V2_SystemPromptMapped(t *testing.T) {
	_, restore := withTempConfigPaths(t)
	defer restore()

	yaml := `
system_prompt: |
  You are a test system.
preferences:
  model: gpt-4.1
models:
  - name: gpt-4.1
config_format_version: "2"
`
	abs, _ := FullFilePath(configFilePath)
	os.MkdirAll(filepath.Dir(abs), 0o755)
	ioutil.WriteFile(abs, []byte(yaml), 0644)

	os.Setenv("OPENAI_API_KEY", "tok")
	os.Setenv("OPENAI_ORG_ID", "org")

	got, err := ResolveSelectedModel()
	if err != nil {
		t.Fatalf("resolve v2 system prompt: %v", err)
	}
	if len(got.Prompt) == 0 || got.Prompt[0].Role != "system" || !strings.Contains(got.Prompt[0].Content, "You are a test system.") {
		t.Fatalf("expected system prompt mapped, got: %+v", got.Prompt)
	}
}

func TestResolveSelectedModel_V2_AzureHeader(t *testing.T) {
	_, restore := withTempConfigPaths(t)
	defer restore()

	yaml := `
preferences:
  model: gpt-4.1
  config: azure-openai
models:
  - name: gpt-4.1
configs:
  azure-openai:
    payload_style: openai
    endpoint: https://example.azure.com/openai/deployments/{name}/chat/completions
    auth:
      type: header
      header: Api-Key
      token_env: AZURE_OPENAI_API_KEY
    headers: {}
    query_params:
      api-version: "2024-05-01-preview"
config_format_version: "2"
`
	abs, _ := FullFilePath(configFilePath)
	os.MkdirAll(filepath.Dir(abs), 0o755)
	ioutil.WriteFile(abs, []byte(yaml), 0644)
	os.Setenv("AZURE_OPENAI_API_KEY", "tok")

	got, err := ResolveSelectedModel()
	if err != nil {
		t.Fatalf("resolve azure: %v", err)
	}
	if got.Headers["Api-Key"] != "tok" {
		t.Fatalf("expected Api-Key header, got headers=%v", got.Headers)
	}
	if !strings.Contains(got.Endpoint, "deployments/gpt-4.1/chat/completions") || !strings.Contains(got.Endpoint, "api-version=") {
		t.Fatalf("interpolation or query params missing: %s", got.Endpoint)
	}
}

func TestOpenAIAdapter_ParseStream(t *testing.T) {
	a := openai.Adapter{}
	body := "data: {\"choices\":[{\"delta\":{\"content\":\"hel\"}}]}\n" +
		"data: {\"choices\":[{\"delta\":{\"content\":\"lo\"}}]}\n" +
		"data: [DONE]\n"
	resp := &http.Response{StatusCode: 200, Body: ioutil.NopCloser(strings.NewReader(body))}
	var last string
	got, err := a.ParseStream(resp, func(s string) error { last = s; return nil })
	if err != nil {
		t.Fatalf("parse stream: %v", err)
	}
	if got != "hello" || last != "hello" {
		t.Fatalf("unexpected content: got=%q last=%q", got, last)
	}
}

func TestOllamaAdapter_ParseStream(t *testing.T) {
	a := ollama.Adapter{}
	body := "{\"response\":\"hel\"}\n{\"response\":\"lo\"}\n{\"done\":true}\n"
	resp := &http.Response{StatusCode: 200, Body: ioutil.NopCloser(strings.NewReader(body))}
	got, err := a.ParseStream(resp, nil)
	if err != nil {
		t.Fatalf("ollama parse: %v", err)
	}
	if got != "hello" {
		t.Fatalf("unexpected content: %q", got)
	}
}
