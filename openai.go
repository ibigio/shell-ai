package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Payload struct {
	Model       string    `json:"model"`
	Prompt      string    `json:"prompt,omitempty"`
	MaxTokens   int       `json:"max_tokens"`
	Temperature float32   `json:"temperature"`
	Stop        []string  `json:"stop,omitempty"`
	Messages    []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Response struct {
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
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
		Index        int    `json:"index"`
	} `json:"choices"`
}

type OpenAIClient struct {
	apiKey    string
	url       string
	model     string
	maxTokens int

	messages []Message

	httpClient *http.Client
}

func promptForModel(model string) []Message {
	switch model {
	case "gpt-4":
		return []Message{
			{Role: "system", Content: "You are a terminal assistant. Turn the natural language instructions into a terminal command. By default always only output code, and in a code block. However, if the user is clearly asking a question then answer it very briefly and well."},
			{Role: "user", Content: "print hi"},
			{Role: "assistant", Content: "```bash\necho \"hi\"\n```"},
		}
	}
	// default for gpt-3.5-turbo
	return []Message{
		{Role: "system", Content: "You are a terminal assistant. Turn the natural language instructions into a terminal command. By default always only output code, and in a code block. DO NOT OUTPUT ADDITIONAL REMARKS ABOUT THE CODE YOU OUTPUT. Do not repeat the question the users asks. Do not add explanations for your code. Do not output any non-code words at all. Just output the code. Short is better. However, if the user is clearly asking a general question then answer it very briefly and well."},
		{Role: "user", Content: "get the current time from some website"},
		{Role: "assistant", Content: "```bash\ncurl -s http://worldtimeapi.org/api/ip | jq '.datetime'\n```"},
		{Role: "user", Content: "print hi"},
		{Role: "assistant", Content: "```bash\necho \"hi\"\n```"},
	}
}

func NewClient(apiKey string, modelOverride string) *OpenAIClient {
	model := "gpt-3.5-turbo"
	if modelOverride != "" {
		model = modelOverride
	}

	return &OpenAIClient{
		apiKey:    apiKey,
		url:       "https://api.openai.com/v1/chat/completions",
		model:     model,
		maxTokens: 256,

		messages: promptForModel(model),

		httpClient: &http.Client{},
	}
}

func (c *OpenAIClient) Query(query string) (string, error) {
	messages := c.messages
	messages = append(messages, Message{Role: "user", Content: query})

	payload := Payload{
		Model:       c.model,
		Messages:    messages,
		Temperature: 0,
		MaxTokens:   c.maxTokens,
	}
	message, err := c.call(payload)
	if err != nil {
		return "", err
	}
	c.messages = append(c.messages, message)
	return message.Content, nil
}

func (c *OpenAIClient) call(payload Payload) (Message, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return Message{}, fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", c.url, bytes.NewReader(payloadBytes))
	if err != nil {
		return Message{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Message{}, fmt.Errorf("failed to make the API request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Message{}, fmt.Errorf("API request failed: %s", resp.Status)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Message{}, fmt.Errorf("failed to read response body: %w", err)
	}

	var response Response
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return Message{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(response.Choices) == 0 {
		return Message{}, fmt.Errorf("no completions found in response")
	}

	return response.Choices[0].Message, nil
}
