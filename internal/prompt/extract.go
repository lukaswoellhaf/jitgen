package prompt

import "strings"

// ExtractGoCode pulls the first fenced ```go code block from the LLM response.
// Falls back to the raw response if no code block is found.
func ExtractGoCode(response string) string {
	// Try ```go ... ```
	start := strings.Index(response, "```go")
	if start == -1 {
		// Try bare ``` ... ```
		start = strings.Index(response, "```")
		if start == -1 {
			return strings.TrimSpace(response)
		}
		start += len("```")
	} else {
		start += len("```go")
	}

	// Skip to next line
	if idx := strings.Index(response[start:], "\n"); idx != -1 {
		start += idx + 1
	}

	end := strings.Index(response[start:], "```")
	if end == -1 {
		return strings.TrimSpace(response[start:])
	}

	return strings.TrimSpace(response[start : start+end])
}
