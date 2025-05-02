package provider

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	. "q/types"
	"strings"
)

type OpenAIProvider struct {
	config      ModelConfig
	httpClient  *http.Client
	streamFunc  func(string, error)
}

func NewOpenAIProvider(config ModelConfig) *OpenAIProvider {
	return &OpenAIProvider{
		config: config,
	}
}

func (p *OpenAIProvider) SetStreamCallback(callback func(string, error)) {
	p.streamFunc = callback
}

func (p *OpenAIProvider) CreateRequest(payload Payload) (*http.Request, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", p.config.Endpoint, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	apiKey := os.Getenv(p.config.Auth)
	if apiKey == "" {
		return nil, fmt.Errorf("API key environment variable '%s' not set", p.config.Auth)
	}

	if strings.Contains(p.config.Endpoint, "openai.azure.com") {
		req.Header.Set("Api-Key", apiKey)
	} else {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	if p.config.OrgID != "" {
		orgID := os.Getenv(p.config.OrgID)
		if orgID != "" {
			req.Header.Set("OpenAI-Organization", orgID)
		}
	}

	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func (p *OpenAIProvider) ProcessStream(resp *http.Response) (string, error) {
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

			var responseData OpenAIResponseData
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
			if p.streamFunc != nil {
				p.streamFunc(totalData, nil)
			}
			counter++
		}
	}

	return totalData, nil
}

func (p *OpenAIProvider) GetProviderName() string {
	return "openai"
} 