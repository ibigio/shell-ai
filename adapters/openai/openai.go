package openai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	types "q/types"
	"strings"
)

type Adapter struct{}

func (a Adapter) Name() string { return "openai-compatible" }

func (a Adapter) BuildRequest(cfg types.ResolvedModelConfig, messages []types.Message) (*http.Request, error) {
	payload := map[string]interface{}{
		"model":    cfg.Name,
		"messages": messages,
		"stream":   true,
	}
	if cfg.Temperature > 0 {
		payload["temperature"] = cfg.Temperature
	}
	// Note: Chat Completions currently rejects unknown parameters like reasoning or reasoning_effort.
	// If support for reasoning-specific knobs is needed, a separate adapter using the Responses API should be used.
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}
	req, err := http.NewRequest("POST", cfg.Endpoint, bytes.NewBuffer(b))
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	// headers
	req.Header.Set("Content-Type", "application/json")
	for k, v := range cfg.Headers {
		req.Header.Set(k, v)
	}
	return req, nil
}

func (a Adapter) ParseStream(resp *http.Response, onDelta func(string) error) (string, error) {
	reader := bufio.NewReader(resp.Body)
	total := ""
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)
		if line == "data: [DONE]" {
			break
		}
		if strings.HasPrefix(line, "data:") {
			payload := strings.TrimPrefix(line, "data:")
			var rd types.ResponseData
			if err := json.Unmarshal([]byte(payload), &rd); err != nil {
				// ignore malformed lines
				continue
			}
			if len(rd.Choices) == 0 {
				continue
			}
			content := rd.Choices[0].Delta.Content
			if content == "" {
				// ignore role-only or empty deltas
				continue
			}
			total += content
			if onDelta != nil {
				if err := onDelta(total); err != nil {
					return total, err
				}
			}
		}
	}
	return total, nil
}
