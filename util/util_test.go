package util

import "testing"

func TestStartsWithCodeBlock(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"", true},    // len<=3: all backticks equals input (empty)
		{"`", true},   // single backtick
		{"``", true},  // double backtick
		{"```", true}, // triple
		{"```bash", true},
		{"not code", false},
		{"`` code", false},
	}
	for _, tc := range cases {
		if got := StartsWithCodeBlock(tc.in); got != tc.want {
			t.Fatalf("StartsWithCodeBlock(%q)=%v want %v", tc.in, got, tc.want)
		}
	}
}

func TestExtractFirstCodeBlock(t *testing.T) {
	// Only code block at start with language
	content, only := ExtractFirstCodeBlock("```bash\necho hi\n```")
	if content != "echo hi" || !only {
		t.Fatalf("unexpected: content=%q only=%v", content, only)
	}

	// Embedded code block, should flip isOnlyCode to false
	s := "hello before\n```bash\ncmd\n```\nafter"
	content, only = ExtractFirstCodeBlock(s)
	if content != "cmd" || only {
		t.Fatalf("embedded: content=%q only=%v", content, only)
	}

	// Missing closing fence: still returns content after opening fence; only=false due to logic
	content, only = ExtractFirstCodeBlock("```bash\nls -la")
	if content != "ls -la" || only {
		t.Fatalf("missing footer: content=%q only=%v", content, only)
	}

	// No code block present
	content, only = ExtractFirstCodeBlock("no fences here")
	if content != "" || only {
		t.Fatalf("no fences: content=%q only=%v", content, only)
	}
}

func TestIsLikelyBillingError(t *testing.T) {
	if !IsLikelyBillingError("429 Too Many Requests: quota exceeded") {
		t.Fatalf("expected true for 429 message")
	}
	if IsLikelyBillingError("some other error") {
		t.Fatalf("expected false for unrelated error")
	}
}
