package prompt

import "testing"

func TestExtractGoCode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "fenced go block",
			input:    "Here is the test:\n```go\npackage foo\n\nfunc TestX(t *testing.T) {}\n```\nDone.",
			expected: "package foo\n\nfunc TestX(t *testing.T) {}",
		},
		{
			name:     "bare fenced block",
			input:    "```\npackage foo\n```",
			expected: "package foo",
		},
		{
			name:     "no code block",
			input:    "package foo\n\nfunc TestX(t *testing.T) {}",
			expected: "package foo\n\nfunc TestX(t *testing.T) {}",
		},
		{
			name:     "multiple blocks returns first",
			input:    "```go\nfirst\n```\n```go\nsecond\n```",
			expected: "first",
		},
		{
			name:     "empty response",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractGoCode(tt.input)
			if got != tt.expected {
				t.Errorf("got %q, want %q", got, tt.expected)
			}
		})
	}
}
