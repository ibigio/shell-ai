package types

import (
	"fmt"
	"os"
)

type ValueOrVar string

func (v ValueOrVar) Raw() string {
	return string(v)
}

func (v ValueOrVar) Resolve() string {
	return os.ExpandEnv(string(v))
}

type ModelConfig struct {
	ModelName string     `yaml:"name"`
	Endpoint  string     `yaml:"endpoint"`
	ApiKey    ValueOrVar `yaml:"api_key,omitempty"`
	OrgID     ValueOrVar `yaml:"org_id,omitempty"`
	ProjectID ValueOrVar `yaml:"project_id,omitempty"`
	Prompt    []Message  `yaml:"prompt"`
	// deprecated var-only keys
	V1_Auth      string `yaml:"auth_env_var,omitempty"`
	V1_OrgID     string `yaml:"org_env_var,omitempty"`
	V1_ProjectID string `yaml:"project_env_var,omitempty"`
}

func (c ModelConfig) Migrate() ModelConfig {
	c.ApiKey = ValueOrVar(fmt.Sprintf("${%s}", c.V1_Auth))
	c.V1_Auth = ""

	c.OrgID = ValueOrVar(fmt.Sprintf("${%s}", c.V1_OrgID))
	c.V1_OrgID = ""

	c.ProjectID = ValueOrVar("${OPENAI_PROJECT_ID}")

	return c
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
