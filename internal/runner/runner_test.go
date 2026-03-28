package runner

import "testing"

func TestTargetPkgDir(t *testing.T) {
	tests := []struct {
		name     string
		diff     string
		expected string
	}{
		{
			name: "root level go file",
			diff: `--- a/utils.go
+++ b/utils.go`,
			expected: ".",
		},
		{
			name: "nested go file",
			diff: `--- a/internal/foo/bar.go
+++ b/internal/foo/bar.go`,
			expected: "internal/foo",
		},
		{
			name: "skips test files",
			diff: `--- a/pkg/foo_test.go
+++ b/pkg/foo_test.go
--- a/pkg/foo.go
+++ b/pkg/foo.go`,
			expected: "pkg",
		},
		{
			name: "no go files falls back to dot",
			diff: `--- a/README.md
+++ b/README.md`,
			expected: ".",
		},
		{
			name:     "empty diff",
			diff:     "",
			expected: ".",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := targetPkgDir(tt.diff)
			if got != tt.expected {
				t.Errorf("got %q, want %q", got, tt.expected)
			}
		})
	}
}
