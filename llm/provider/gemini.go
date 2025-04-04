package provider

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	. "q/types"
	"strings"
)

type GeminiProvider struct {
	config     ModelConfig
	httpClient *http.Client
	streamFunc func(string, error)
}

func NewGeminiProvider(config ModelConfig) *GeminiProvider {
	return &GeminiProvider{
		config: config,
	}
}

func (p *GeminiProvider) SetStreamCallback(callback func(string, error)) {
	p.streamFunc = callback
}

func (p *GeminiProvider) CreateRequest(payload Payload) (*http.Request, error) {
	geminiContents := make([]GeminiContent, 0, len(payload.Messages))
	var systemPromptContent strings.Builder
	userMessageFound := false

	for _, msg := range payload.Messages {
		role := msg.Role
		content := msg.Content

		if role == "system" && !userMessageFound {
			if systemPromptContent.Len() > 0 {
				systemPromptContent.WriteString("\n\n")
			}
			systemPromptContent.WriteString(content)
			continue
		} else if role == "user" {
			userMessageFound = true
			if systemPromptContent.Len() > 0 {
				content = systemPromptContent.String() + "\n\n" + content
				systemPromptContent.Reset()
			}
		} else if role == "assistant" {
			role = "model"
		} else if role == "system" && userMessageFound {
			fmt.Printf("Warning: Ignoring system message after user message: %s\n", content)
			continue
		}

		if len(geminiContents) > 0 && geminiContents[len(geminiContents)-1].Role == role {
			lastContent := &geminiContents[len(geminiContents)-1]
			if len(lastContent.Parts) > 0 {
				lastPart := &lastContent.Parts[0]
				lastPart.Text += "\n\n" + content
				continue
			}
		}

		geminiContents = append(geminiContents, GeminiContent{
			Role:  role,
			Parts: []GeminiPart{{Text: content}},
		})
	}

	geminiPayload := GeminiPayload{
		Contents: geminiContents,
		GenerationConfig: GeminiGenerationConfig{
			Temperature: payload.Temperature,
		},
	}

	payloadBytes, err := json.Marshal(geminiPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Gemini payload: %w", err)
	}

	endpoint := p.config.Endpoint
	streamSuffix := ":generateContent"
	if payload.Stream {
		streamSuffix = ":streamGenerateContent"
	}
	if strings.HasSuffix(endpoint, ":generateContent") {
		endpoint = strings.TrimSuffix(endpoint, ":generateContent") + streamSuffix
	} else if strings.HasSuffix(endpoint, ":streamGenerateContent") {
		endpoint = strings.TrimSuffix(endpoint, ":streamGenerateContent") + streamSuffix
	} else {
		endpoint += streamSuffix
	}

	queryParams := url.Values{}
	apiKey := os.Getenv(p.config.Auth)
	if apiKey == "" {
		return nil, fmt.Errorf("API key environment variable '%s' not set", p.config.Auth)
	}
	queryParams.Set("key", apiKey)

	if payload.Stream {
		queryParams.Set("alt", "sse")
	}

	finalURL := endpoint + "?" + queryParams.Encode()

	req, err := http.NewRequest("POST", finalURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func (p *GeminiProvider) ProcessStream(resp *http.Response) (string, error) {
	streamReader := bufio.NewReader(resp.Body)
	totalData := ""

	for {
		line, err := streamReader.ReadString('\n')
		if err != nil {
			break
		}

		line = strings.TrimSpace(line)
		if line == "data: [DONE]" || line == "data: " {
			break
		}

		if strings.HasPrefix(line, "data: ") {
			payload := strings.TrimPrefix(line, "data: ")

			var responseData GeminiStreamResponse
			err = json.Unmarshal([]byte(payload), &responseData)
			if err != nil {
				fmt.Println("Error parsing Gemini data:", err)
				continue
			}

			if len(responseData.Candidates) > 0 &&
				len(responseData.Candidates[0].Content.Parts) > 0 {
				content := responseData.Candidates[0].Content.Parts[0].Text
				totalData += content
				if p.streamFunc != nil {
					p.streamFunc(totalData, nil)
				}
			}
		}
	}

	return totalData, nil
}

func (p *GeminiProvider) GetProviderName() string {
	return "gemini"
} 