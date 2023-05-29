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

func QueryOpenAIAssistant(apiKey string, messages []Message) (string, error) {
	url := "https://api.openai.com/v1/chat/completions"

	payload := Payload{
		Model: "gpt-4",
		// Model:       "gpt-3.5-turbo",
		Messages:    messages,
		Temperature: 0,
		MaxTokens:   256,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewReader(payloadBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make the API request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed: %s", resp.Status)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	completions, ok := response["choices"].([]interface{})
	if !ok || len(completions) == 0 {
		return "", fmt.Errorf("no completions found in response")
	}

	lastCompletion, ok := completions[len(completions)-1].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("failed to parse last completion in response")
	}

	message, ok := lastCompletion["message"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("failed to parse message in last completion")
	}

	content, ok := message["content"].(string)
	if !ok {
		return "", fmt.Errorf("failed to parse content in last completion")
	}

	return content, nil
}
