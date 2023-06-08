package main

import "strings"

func startsWithCodeBlock(s string) bool {
	if len(s) <= 3 {
		return strings.Repeat("`", len(s)) == s
	}
	return strings.HasPrefix(s, "```")
}

func extractFirstCodeBlock(s string) string {
	if len(s) <= 3 {
		return ""
	}
	start := strings.Index(s, "```")
	if start == -1 {
		return ""
	}
	fromStart := s[start:]
	content := strings.TrimPrefix(fromStart, "```")
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
	if end != -1 {
		content = content[:end]
	}
	if len(content) == 0 {
		return ""
	}
	// Strip the final newline, if present
	if content[len(content)-1] == '\n' {
		content = content[:len(content)-1]
	}
	return content
}
