package anthropic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	types "q/types"
)

// Minimal stub adapter for Anthropic Messages API. Streaming event parsing can be added later.
type Adapter struct{}

func (a Adapter) Name() string { return "anthropic" }

func (a Adapter) BuildRequest(cfg types.ResolvedModelConfig, messages []types.Message) (*http.Request, error) {
	// Anthropic uses a different payload; map simply for now
	conv := make([]map[string]string, 0, len(messages))
	for _, m := range messages {
		conv = append(conv, map[string]string{"role": m.Role, "content": m.Content})
	}
	payload := map[string]interface{}{
		"model":    cfg.Name,
		"messages": conv,
		"stream":   true,
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

func (a Adapter) ParseStream(resp *http.Response, onDelta func(string) error) (string, error) {
	// TODO: Implement Anthropic event stream parsing. For now, return error to signal unsupported.
	return "", fmt.Errorf("anthropic streaming not implemented")
}
