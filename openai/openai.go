package openai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Payload struct {
	Model       string    `json:"model"`
	Prompt      string    `json:"prompt,omitempty"`
	MaxTokens   int       `json:"max_tokens"`
	Temperature float32   `json:"temperature"`
	Stop        []string  `json:"stop,omitempty"`
	Messages    []Message `json:"messages"`
	Stream      bool      `json:"stream"`
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

type Choice struct {
	Delta struct {
		Content string `json:"content"`
	} `json:"delta"`
	Index        int    `json:"index"`
	FinishReason string `json:"finish_reason"`
}

type ResponseData struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int      `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
}

type OpenAIClient struct {
	apiKey    string
	url       string
	model     string
	maxTokens int

	messages []Message

	StreamCallback func(string, error)

	httpClient *http.Client
}

type SSEMessage struct {
	content string
	isDone  bool
}

func NewClient(apiKey string, model string, url string, model_prompt []Message) *OpenAIClient {
	return &OpenAIClient{
		apiKey:    apiKey,
		url:       url,
		model:     model,
		maxTokens: 256,

		messages: model_prompt,

		httpClient: &http.Client{
			Timeout: time.Second * 120,
		},
	}
}

func (c *OpenAIClient) createRequest(payload Payload) (*http.Request, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}
	req, err := http.NewRequest("POST", c.url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func (c *OpenAIClient) Query(query string) (string, error) {
	messages := c.messages
	messages = append(messages, Message{Role: "user", Content: query})

	payload := Payload{
		Model:       c.model,
		Messages:    messages,
		Temperature: 0,
		MaxTokens:   c.maxTokens,
		Stream:      true,
	}
	message, err := c.callStream(payload)
	if err != nil {
		return "", err
	}
	c.messages = append(c.messages, message)
	return message.Content, nil
}

func (c *OpenAIClient) processStream(resp *http.Response) (string, error) {
	counter := 0
	streamReader := bufio.NewReader(resp.Body)
	totalData := ""
	for {
		line, err := streamReader.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)
		if line == "data: [DONE]" {
			break
		}
		if strings.HasPrefix(line, "data:") {
			payload := strings.TrimPrefix(line, "data:")

			var responseData ResponseData
			err = json.Unmarshal([]byte(payload), &responseData)
			if err != nil {
				fmt.Println("Error parsing data:", err)
				continue
			}
			if len(responseData.Choices) == 0 {
				continue
			}
			content := responseData.Choices[0].Delta.Content
			if counter < 2 && strings.Count(content, "\n") > 0 {
				continue
			}
			totalData += content
			c.StreamCallback(totalData, nil)
			counter++
		}
	}
	return totalData, nil
}

func (c *OpenAIClient) callStream(payload Payload) (Message, error) {
	req, err := c.createRequest(payload)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Message{}, fmt.Errorf("failed to make the API request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return Message{}, fmt.Errorf("API request failed: %s", resp.Status)
	}
	content, err := c.processStream(resp)
	return Message{Role: "assistant", Content: content}, nil
}
