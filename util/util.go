package util

import (
	"os/exec"
	"runtime"
	"strings"

	"github.com/mattn/go-tty"
)

const (
	TermMaxWidth        = 100
	TermSafeZonePadding = 10
)

func StartsWithCodeBlock(s string) bool {
	if len(s) <= 3 {
		return strings.Repeat("`", len(s)) == s
	}
	return strings.HasPrefix(s, "```")
}

func ExtractFirstCodeBlock(s string) (content string, isOnlyCode bool) {
	isOnlyCode = true
	if len(s) <= 3 {
		return "", false
	}
	start := strings.Index(s, "```")
	if start == -1 {
		return "", false
	}
	if start != 0 {
		isOnlyCode = false
	}
	fromStart := s[start:]
	content = strings.TrimPrefix(fromStart, "```")
	// Find newline after the first ```
	newlinePos := strings.Index(content, "\n")
	if newlinePos != -1 {
		// Check if there's a word immediately after the first ```
		if content[0:newlinePos] == strings.TrimSpace(content[0:newlinePos]) {
			// If so, remove that part from the content
			content = content[newlinePos+1:]
		}
	}
	// Strip final ``` if present
	end := strings.Index(content, "```")
	if end < len(content)-3 {
		isOnlyCode = false
	}
	if end != -1 {
		content = content[:end]
	}
	if len(content) == 0 {
		return "", false
	}
	// Strip the final newline, if present
	if content[len(content)-1] == '\n' {
		content = content[:len(content)-1]
	}
	return
}

func GetTermSafeMaxWidth() int {
	maxWidth := TermMaxWidth
	termWidth, err := getTermWidth()
	if err != nil || termWidth < maxWidth {
		maxWidth = termWidth - TermSafeZonePadding
	}
	return maxWidth
}

func getTermWidth() (width int, err error) {
	t, err := tty.Open()
	if err != nil {
		return 0, err
	}
	defer t.Close()
	width, _, err = t.Size()
	return width, err
}

func IsLikelyBillingError(s string) bool {
	return strings.Contains(s, "429 Too Many Requests")
}

func OpenBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default: // For Linux or anything else
		cmd = exec.Command("xdg-open", url)
	}

	return cmd.Start()
}
