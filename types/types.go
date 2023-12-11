package types

type ModelConfig struct {
	ModelName string    `yaml:"name"`
	Endpoint  string    `yaml:"endpoint"`
	Auth      string    `yaml:"auth"`
	Prompt    []Message `yaml:"prompt"`
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
	MaxTokens   int       `json:"max_tokens"`
	Temperature float32   `json:"temperature"`
	Stop        []string  `json:"stop,omitempty"`
	Messages    []Message `json:"messages"`
	Stream      bool      `json:"stream"`
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
