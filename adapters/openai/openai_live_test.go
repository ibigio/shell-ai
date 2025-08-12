package openai

import (
	"net/http"
	"os"
	"strings"
	"testing"

	types "q/types"
)

func liveClientOrSkip(t *testing.T) (string, string) {
	t.Helper()
	key := os.Getenv("OPENAI_API_KEY")
	if key == "" {
		t.Skip("OPENAI_API_KEY not set; skipping live test")
	}
	org := os.Getenv("OPENAI_ORG_ID")
	return key, org
}

func TestLive_OpenAI_CodeBlockVerbatim(t *testing.T) {
	key, org := liveClientOrSkip(t)
	a := Adapter{}
	cfg := types.ResolvedModelConfig{
		Name:         "gpt-4.1",
		Endpoint:     "https://api.openai.com/v1/chat/completions",
		PayloadStyle: "openai",
		Headers:      map[string]string{"Authorization": "Bearer " + key},
		Temperature:  0,
	}
	if org != "" {
		cfg.Headers["OpenAI-Organization"] = org
	}
	prompt := "Return exactly the following content verbatim, no commentary, no surrounding text: ```bash\nls -a\n```"
	msgs := []types.Message{{Role: "user", Content: prompt}}

	req, err := a.BuildRequest(cfg, msgs)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("status: %s", resp.Status)
	}
	out, err := a.ParseStream(resp, nil)
	if err != nil {
		t.Fatalf("parse stream: %v", err)
	}
	if !strings.Contains(out, "```bash\nls -a\n```") {
		t.Fatalf("expected code block in output, got: %q", out)
	}
}

func TestLive_OpenAI_EchoSimple(t *testing.T) {
	key, org := liveClientOrSkip(t)
	a := Adapter{}
	cfg := types.ResolvedModelConfig{
		Name:         "gpt-4.1",
		Endpoint:     "https://api.openai.com/v1/chat/completions",
		PayloadStyle: "openai",
		Headers:      map[string]string{"Authorization": "Bearer " + key},
		Temperature:  0,
	}
	if org != "" {
		cfg.Headers["OpenAI-Organization"] = org
	}
	prompt := "Repeat exactly this: HELLO_WORLD_12345"
	msgs := []types.Message{{Role: "user", Content: prompt}}

	req, err := a.BuildRequest(cfg, msgs)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("status: %s", resp.Status)
	}
	out, err := a.ParseStream(resp, nil)
	if err != nil {
		t.Fatalf("parse stream: %v", err)
	}
	if !strings.Contains(out, "HELLO_WORLD_12345") {
		t.Fatalf("expected echo token in output, got: %q", out)
	}
}

func TestLive_OpenAI_ReasoningModelEcho(t *testing.T) {
	key, org := liveClientOrSkip(t)
	model := os.Getenv("OPENAI_REASONING_MODEL")
	if model == "" {
		model = "gpt-5" // expected reasoning-capable
	}
	a := Adapter{}
	cfg := types.ResolvedModelConfig{
		Name:         model,
		Endpoint:     "https://api.openai.com/v1/chat/completions",
		PayloadStyle: "openai",
		Headers:      map[string]string{"Authorization": "Bearer " + key},
		Temperature:  0,
	}
	if org != "" {
		cfg.Headers["OpenAI-Organization"] = org
	}
	prompt := "Return exactly the following content verbatim, no commentary, no surrounding text: ```bash\necho hello world\n```"
	msgs := []types.Message{{Role: "user", Content: prompt}}

	req, err := a.BuildRequest(cfg, msgs)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		// Best-effort read (no ioutil)
		var sb strings.Builder
		buf := make([]byte, 4096)
		for {
			n, _ := resp.Body.Read(buf)
			if n <= 0 {
				break
			}
			sb.Write(buf[:n])
			if n < len(buf) {
				break
			}
		}
		t.Fatalf("reasoning model call failed: %s body=%s", resp.Status, sb.String())
	}
	out, err := a.ParseStream(resp, nil)
	if err != nil {
		t.Fatalf("parse stream: %v", err)
	}
	if !strings.Contains(out, "```bash\necho hello world\n```") {
		t.Fatalf("expected code block in output, got: %q", out)
	}
}
