package llm

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	. "q/types"
)

func TestCreateRequest_Headers(t *testing.T) {
	tests := []struct {
		name         string
		cfg          ModelConfig
		expectAuth   string
		expectAPIKey string
		expectOrg    string
	}{
		{
			name: "openai_with_org",
			cfg: ModelConfig{
				ModelName: "gpt-5",
				Endpoint:  "https://api.openai.com/v1/chat/completions",
				Auth:      "SECRET",
				OrgID:     "ORG",
			},
			expectAuth:   "Bearer SECRET",
			expectAPIKey: "",
			expectOrg:    "ORG",
		},
		{
			name: "openai_no_org",
			cfg: ModelConfig{
				ModelName: "gpt-5",
				Endpoint:  "https://api.openai.com/v1/chat/completions",
				Auth:      "SECRET",
			},
			expectAuth:   "Bearer SECRET",
			expectAPIKey: "",
			expectOrg:    "",
		},
		{
			name: "azure_api_key",
			cfg: ModelConfig{
				ModelName: "my-deployment",
				Endpoint:  "https://foo.openai.azure.com/openai/deployments/my-deployment/chat/completions?api-version=2024-05-01-preview",
				Auth:      "AZSECRET",
			},
			expectAuth:   "",
			expectAPIKey: "AZSECRET",
			expectOrg:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewLLMClient(tt.cfg)
			req, err := c.createRequest(Payload{Model: tt.cfg.ModelName, Messages: []Message{}})
			if err != nil {
				t.Fatalf("createRequest: %v", err)
			}
			if got := req.Header.Get("Authorization"); got != tt.expectAuth {
				t.Fatalf("Authorization header mismatch: got %q want %q", got, tt.expectAuth)
			}
			if got := req.Header.Get("Api-Key"); got != tt.expectAPIKey {
				t.Fatalf("Api-Key header mismatch: got %q want %q", got, tt.expectAPIKey)
			}
			if got := req.Header.Get("OpenAI-Organization"); got != tt.expectOrg {
				t.Fatalf("OpenAI-Organization mismatch: got %q want %q", got, tt.expectOrg)
			}
		})
	}
}

func TestCreateRequest_BodyIncludesModelAndMessages(t *testing.T) {
	cfg := ModelConfig{ModelName: "gpt-5", Endpoint: "https://api.openai.com/v1/chat/completions"}
	c := NewLLMClient(cfg)
	p := Payload{Model: cfg.ModelName, Messages: []Message{{Role: "user", Content: "hi"}}}
	req, err := c.createRequest(p)
	if err != nil {
		t.Fatalf("createRequest: %v", err)
	}
	body, _ := ioutil.ReadAll(req.Body)
	req.Body.Close()
	var roundtrip Payload
	if err := json.Unmarshal(body, &roundtrip); err != nil {
		t.Fatalf("body not JSON: %v; %s", err, string(body))
	}
	if roundtrip.Model != "gpt-5" || len(roundtrip.Messages) != 1 || roundtrip.Messages[0].Role != "user" {
		t.Fatalf("unexpected payload: %+v", roundtrip)
	}
	if !strings.Contains(req.Header.Get("Content-Type"), "application/json") {
		t.Fatalf("missing Content-Type: application/json")
	}
}

type roundTripFunc func(*http.Request) *http.Response

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func newTestHTTPClient(fn roundTripFunc) *http.Client {
	return &http.Client{Transport: fn}
}

func TestStreamingBehavior(t *testing.T) {
	t.Run("aggregates_and_updates_history", func(t *testing.T) {
		// Prepare a fake streaming response body
		stream := strings.Join([]string{
			"data: {\"id\":\"1\",\"object\":\"chat.completion.chunk\",\"model\":\"x\",\"choices\":[{\"delta\":{\"content\":\"he\"},\"index\":0}]}",
			"data: {\"id\":\"1\",\"object\":\"chat.completion.chunk\",\"model\":\"x\",\"choices\":[{\"delta\":{\"content\":\"llo\"},\"index\":0}]}",
			"data: [DONE]",
			"",
		}, "\n")

		var capturedBody []byte
		transport := newTestHTTPClient(roundTripFunc(func(r *http.Request) *http.Response {
			b, _ := ioutil.ReadAll(r.Body)
			r.Body.Close()
			capturedBody = b
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(stream)),
				Header:     make(http.Header),
			}
		}))

		cfg := ModelConfig{
			ModelName: "gpt-5",
			Endpoint:  "https://api.openai.com/v1/chat/completions",
			Prompt:    []Message{{Role: "system", Content: "S"}},
		}
		c := NewLLMClient(cfg)
		c.httpClient = transport
		callbackCalls := 0
		c.StreamCallback = func(s string, err error) { callbackCalls++ }

		out, err := c.Query("hi")
		if err != nil {
			t.Fatalf("Query: %v", err)
		}
		if out != "hello" {
			t.Fatalf("unexpected output: %q", out)
		}
		if len(c.messages) == 0 || c.messages[len(c.messages)-1].Role != "assistant" || c.messages[len(c.messages)-1].Content != "hello" {
			t.Fatalf("assistant message not appended: %+v", c.messages)
		}
		if callbackCalls == 0 {
			t.Fatalf("expected StreamCallback to be invoked")
		}

		var sent Payload
		if err := json.Unmarshal(capturedBody, &sent); err != nil {
			t.Fatalf("sent body not JSON: %v; %s", err, string(capturedBody))
		}
		if len(sent.Messages) < 2 || sent.Messages[0].Role != "system" || sent.Messages[1].Role != "user" {
			t.Fatalf("unexpected sent messages: %+v", sent.Messages)
		}
	})

	t.Run("ignores_newline_heavy_initial_chunks", func(t *testing.T) {
		// First chunk has newline(s) and should be ignored while counter < 2
		lines := []string{
			"data: {\"choices\":[{\"delta\":{\"content\":\"line1\\n\"},\"index\":0}]}",
			"data: {\"choices\":[{\"delta\":{\"content\":\"ok\"},\"index\":0}]}",
			"data: [DONE]",
			"",
		}
		body := strings.Join(lines, "\n")
		c := NewLLMClient(ModelConfig{Endpoint: "https://api.openai.com/v1/chat/completions"})
		c.StreamCallback = func(string, error) {}
		resp := &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewBufferString(body))}
		out, err := c.processStream(resp)
		if err != nil {
			t.Fatalf("processStream: %v", err)
		}
		if out != "ok" {
			t.Fatalf("expected 'ok', got %q", out)
		}
	})
}

func TestCallStream_Non200ReturnsError(t *testing.T) {
	transport := newTestHTTPClient(roundTripFunc(func(r *http.Request) *http.Response {
		return &http.Response{StatusCode: 500, Body: io.NopCloser(strings.NewReader("")), Header: make(http.Header)}
	}))
	c := NewLLMClient(ModelConfig{Endpoint: "https://api.openai.com/v1/chat/completions"})
	c.httpClient = transport
	_, err := c.callStream(Payload{Model: "gpt-5", Messages: []Message{}})
	if err == nil {
		t.Fatalf("expected error for non-200 status")
	}
}
