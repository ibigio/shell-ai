package provider

import (
	"net/http"
	. "q/types"
)

type Provider interface {
	CreateRequest(payload Payload) (*http.Request, error)
	ProcessStream(resp *http.Response) (string, error)
	GetProviderName() string
	SetStreamCallback(callback func(string, error))
}

func Factory(config ModelConfig) Provider {
	provider := config.Provider
	if provider == "" {
		provider = "openai"
	}
	
	switch provider {
	case "gemini":
		return NewGeminiProvider(config)
	default:
		return NewOpenAIProvider(config)
	}
}