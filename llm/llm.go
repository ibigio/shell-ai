package llm

import (
	"fmt"
	"net/http"
	"q/llm/provider"
	. "q/types"
	"time"
)

type LLMClient struct {
	config   ModelConfig
	messages []Message
	provider provider.Provider

	StreamCallback func(string, error)

	httpClient *http.Client
}

func NewLLMClient(config ModelConfig) *LLMClient {
	client := &LLMClient{
		config:   config,
		messages: append([]Message(nil), config.Prompt...),
		httpClient: &http.Client{
			Timeout: time.Second * 120,
		},
	}
	
	client.provider = provider.Factory(config)
	
	return client
}

func (c *LLMClient) Query(query string) (string, error) {
	if c.StreamCallback != nil {
		c.provider.SetStreamCallback(c.StreamCallback)
	}

	messages := c.messages
	messages = append(messages, Message{Role: "user", Content: query})

	payload := Payload{
		Model:       c.config.ModelName,
		Messages:    messages,
		Temperature: 0,
		Stream:      true,
	}

	message, err := c.callStream(payload)
	if err != nil {
		return "", err
	}
	c.messages = append(c.messages, message)
	return message.Content, nil
}

func (c *LLMClient) callStream(payload Payload) (Message, error) {
	req, err := c.provider.CreateRequest(payload)
	if err != nil {
		return Message{}, fmt.Errorf("failed to create the request: %w", err)
	}
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Message{}, fmt.Errorf("failed to make the API request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return Message{}, fmt.Errorf("API request failed: %s", resp.Status)
	}
	
	content, err := c.provider.ProcessStream(resp)
	return Message{Role: "assistant", Content: content}, err
}

// returns the underlying provider instance
func (c *LLMClient) GetProvider() provider.Provider {
	return c.provider
}
