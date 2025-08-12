package ollama

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	types "q/types"
)

type Adapter struct{}

func (a Adapter) Name() string { return "ollama" }

func (a Adapter) BuildRequest(cfg types.ResolvedModelConfig, messages []types.Message) (*http.Request, error) {
	// Use chat API shape to align with messages
	payload := map[string]interface{}{
		"model":    cfg.Name,
		"messages": messages,
		"stream":   true,
		"options": map[string]interface{}{
			"temperature": cfg.Temperature,
		},
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}
	req, err := http.NewRequest("POST", cfg.Endpoint, bytes.NewBuffer(b))
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range cfg.Headers {
		req.Header.Set(k, v)
	}
	return req, nil
}

// Ollama streams NDJSON lines where each line contains partial content. Stop on done: true
func (a Adapter) ParseStream(resp *http.Response, onDelta func(string) error) (string, error) {
	scanner := bufio.NewScanner(resp.Body)
	// increase buffer for long tokens
	buf := make([]byte, 0, 1024*64)
	scanner.Buffer(buf, 1024*1024)
	total := ""
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var evt map[string]interface{}
		if err := json.Unmarshal(line, &evt); err != nil {
			continue
		}
		if done, ok := evt["done"].(bool); ok && done {
			break
		}
		if msg, ok := evt["message"].(map[string]interface{}); ok {
			if c, ok := msg["content"].(string); ok {
				total += c
			}
		} else if c, ok := evt["response"].(string); ok {
			total += c
		}
		if onDelta != nil {
			if err := onDelta(total); err != nil {
				return total, err
			}
		}
	}
	return total, nil
}
