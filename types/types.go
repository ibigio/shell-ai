package types

type ModelConfig struct {
	ModelName string    `yaml:"name"`
	Provider  string    `yaml:"provider"`
	Endpoint  string    `yaml:"endpoint"`
	Auth      string    `yaml:"auth_env_var"`
	OrgID     string    `yaml:"org_env_var,omitempty"`
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
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float32   `json:"temperature,omitempty"`
	Messages    []Message `json:"messages"`
	Stream      bool      `json:"stream,omitempty"`
}

type OpenAIResponseData struct {
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

type GeminiPayload struct {
	Contents         []GeminiContent        `json:"contents"`
	GenerationConfig GeminiGenerationConfig `json:"generationConfig,omitempty"`
	SafetySettings   []GeminiSafetySetting  `json:"safetySettings,omitempty"`
}

type GeminiContent struct {
	Role  string       `json:"role"`
	Parts []GeminiPart `json:"parts"`
}

type GeminiPart struct {
	Text string `json:"text"`
}

type GeminiGenerationConfig struct {
	Temperature     float32 `json:"temperature,omitempty"`
	MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
	TopP            float32 `json:"topP,omitempty"`
	TopK            int     `json:"topK,omitempty"`
}

type GeminiSafetySetting struct {
	Category  string `json:"category"`
	Threshold string `json:"threshold"`
}

type GeminiStreamResponse struct {
	Candidates []GeminiCandidate `json:"candidates"`
}

type GeminiCandidate struct {
	Content       GeminiContent       `json:"content"`
	FinishReason  string              `json:"finishReason"`
	SafetyRatings []GeminiSafetyRating `json:"safetyRatings,omitempty"`
}

type GeminiSafetyRating struct {
	Category    string `json:"category"`
	Probability string `json:"probability"`
}
