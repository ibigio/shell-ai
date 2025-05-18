package llm

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	. "q/types"
	"strings"
	"time"
)

type LLMClient struct {
	config   ModelConfig
	messages []Message

	StreamCallback func(string, error)

	httpClient *http.Client
}

func NewLLMClient(config ModelConfig) *LLMClient {
	return &LLMClient{
		config:   config,
		messages: append([]Message(nil), config.Prompt...),

		httpClient: &http.Client{
			Timeout: time.Second * 120,
		},
	}
}

func (c *LLMClient) createRequest(payload Payload) (*http.Request, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}
	req, err := http.NewRequest("POST", c.config.Endpoint, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	if strings.Contains(c.config.Endpoint, "openai.azure.com") {
		req.Header.Set("Api-Key", c.config.Auth)
	} else if strings.Contains(c.config.Endpoint, "api.deepseek.com") {
		req.Header.Set("Authorization", "Bearer "+c.config.Auth)
	} else {
		req.Header.Set("Authorization", "Bearer "+c.config.Auth)
	}
	if c.config.OrgID != "" {
		req.Header.Set("OpenAI-Organization", c.config.OrgID)
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func (c *LLMClient) Query(query string) (string, error) {
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

func (c *LLMClient) processStream(resp *http.Response) (string, error) {
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

func (c *LLMClient) callStream(payload Payload) (Message, error) {
	req, err := c.createRequest(payload)
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
	content, err := c.processStream(resp)
	return Message{Role: "assistant", Content: content}, err
}
