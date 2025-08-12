package llm

import (
	"fmt"
	"net/http"
	"time"

	"q/adapters"
	anthropic "q/adapters/anthropic"
	ollama "q/adapters/ollama"
	openai "q/adapters/openai"
	. "q/types"
)

type LLMClient struct {
	cfg      ResolvedModelConfig
	adapter  adapters.Adapter
	messages []Message

	StreamCallback func(string, error)

	httpClient *http.Client
}

func NewLLMClient(cfg ResolvedModelConfig) *LLMClient {
	return &LLMClient{
		cfg:      cfg,
		adapter:  pickAdapter(cfg.PayloadStyle),
		messages: append([]Message(nil), cfg.Prompt...),
		httpClient: &http.Client{
			Timeout: time.Second * 120,
		},
	}
}

func pickAdapter(style string) adapters.Adapter {
	switch style {
	case "anthropic":
		return anthropic.Adapter{}
	case "ollama":
		return ollama.Adapter{}
	default:
		return openai.Adapter{}
	}
}

func (c *LLMClient) Query(query string) (string, error) {
	msgs := append([]Message(nil), c.messages...)
	msgs = append(msgs, Message{Role: "user", Content: query})

	req, err := c.adapter.BuildRequest(c.cfg, msgs)
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make the API request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("API request failed: %s", resp.Status)
	}
	content, err := c.adapter.ParseStream(resp, func(total string) error {
		if c.StreamCallback != nil {
			c.StreamCallback(total, nil)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	c.messages = append(c.messages, Message{Role: "assistant", Content: content})
	return content, nil
}
