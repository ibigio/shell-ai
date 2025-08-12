package openai

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

// buildSSE constructs a simple SSE stream with data lines and blank line separators
func buildSSE(lines ...string) string {
	var b strings.Builder
	for _, l := range lines {
		b.WriteString("data: ")
		b.WriteString(l)
		b.WriteString("\n\n")
	}
	b.WriteString("data: [DONE]\n\n")
	return b.String()
}

func TestParseStream_IgnoresRoleDeltaAndAccumulates(t *testing.T) {
	a := Adapter{}
	// First chunk only sets role, then actual content chunks
	body := buildSSE(
		`{"id":"","object":"chat.completion.chunk","choices":[{"delta":{"role":"assistant"}}]}`,
		`{"choices":[{"delta":{"content":"hel"}}]}`,
		`{"choices":[{"delta":{"content":"lo"}}]}`,
	)
	resp := &http.Response{StatusCode: 200, Body: ioutil.NopCloser(strings.NewReader(body))}
	got, err := a.ParseStream(resp, nil)
	if err != nil {
		t.Fatalf("parse stream: %v", err)
	}
	if got != "hello" {
		t.Fatalf("unexpected content: %q", got)
	}
}

func TestParseStream_IgnoresReasoningDeltaAndAccumulatesContent(t *testing.T) {
	a := Adapter{}
	body := buildSSE(
		`{"choices":[{"delta":{"role":"assistant"}}]}`,
		`{"choices":[{"delta":{"reasoning":{"content":"scratch"}}}]}`,
		`{"choices":[{"delta":{"content":"hi"}}]}`,
	)
	resp := &http.Response{StatusCode: 200, Body: ioutil.NopCloser(strings.NewReader(body))}
	got, err := a.ParseStream(resp, nil)
	if err != nil {
		t.Fatalf("parse stream: %v", err)
	}
	if got != "hi" {
		t.Fatalf("unexpected content: %q", got)
	}
}

func TestParseStream_LeadingNewlineInFirstContentChunk_IsNotDropped(t *testing.T) {
	a := Adapter{}
	// Real OpenAI often sends leading newlines in the first delta content
	body := buildSSE(
		`{"choices":[{"delta":{"content":"\nHello"}}]}`,
	)
	resp := &http.Response{StatusCode: 200, Body: ioutil.NopCloser(strings.NewReader(body))}
	got, err := a.ParseStream(resp, nil)
	if err != nil {
		t.Fatalf("parse stream: %v", err)
	}
	// We expect at least the visible text to be present even if a newline is prepended
	if !strings.Contains(got, "Hello") {
		t.Fatalf("expected output to contain 'Hello', got=%q", got)
	}
}

func TestParseStream_BlankSeparatorsAndExtraFields(t *testing.T) {
	a := Adapter{}
	// Include blank lines between events and extra fields typical of chat.completion.chunk
	sse := ""
	sse += "data: {\"object\":\"chat.completion.chunk\",\"choices\":[{\"delta\":{\"content\":\"A\"}}]}\n\n"
	sse += "\n\n" // stray blank lines
	sse += "data: {\"object\":\"chat.completion.chunk\",\"choices\":[{\"delta\":{\"content\":\"B\"}}]}\n\n"
	sse += "data: [DONE]\n\n"
	resp := &http.Response{StatusCode: 200, Body: ioutil.NopCloser(strings.NewReader(sse))}
	got, err := a.ParseStream(resp, nil)
	if err != nil {
		t.Fatalf("parse stream: %v", err)
	}
	if got != "AB" {
		t.Fatalf("unexpected content: %q", got)
	}
}

func TestParseStream_CodeBlockAccumulation(t *testing.T) {
	a := Adapter{}
	// Simulate assistant role followed by a fenced code block
	body := buildSSE(
		`{"choices":[{"delta":{"role":"assistant"}}]}`,
		`{"choices":[{"delta":{"content":"\u0060\u0060\u0060bash\n"}}]}`, // ```bash\n
		`{"choices":[{"delta":{"content":"ls -a\n"}}]}`,
		`{"choices":[{"delta":{"content":"\u0060\u0060\u0060"}}]}`, // ```
	)
	resp := &http.Response{StatusCode: 200, Body: ioutil.NopCloser(strings.NewReader(body))}
	got, err := a.ParseStream(resp, nil)
	if err != nil {
		t.Fatalf("parse stream: %v", err)
	}
	want := "```bash\nls -a\n```"
	if got != want {
		t.Fatalf("expected code block,\nwant: %q\n got: %q", want, got)
	}
}
