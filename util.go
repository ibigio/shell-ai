package main

import "strings"

func extractCodeBlock(s string) (string, bool) {
	trimmed := strings.TrimSpace(s)
	if strings.HasPrefix(trimmed, "```") && strings.HasSuffix(trimmed, "```") {
		// There might be a language hint after the first ```
		// Example: ```go
		// We should remove this if it's present
		content := strings.TrimPrefix(trimmed, "```")
		content = strings.TrimSuffix(content, "```")
		// Find newline after the first ```
		newlinePos := strings.Index(content, "\n")
		if newlinePos != -1 {
			// Check if there's a word immediately after the first ```
			if content[0:newlinePos] == strings.TrimSpace(content[0:newlinePos]) {
				// If so, remove that part from the content
				content = content[newlinePos+1:]
			}
		}
		// Strip the final newline, if present
		if content[len(content)-1] == '\n' {
			content = content[:len(content)-1]
		}
		return content, true
	}
	return s, false
}
