package diff

import "testing"

func TestParseTouchedFiles(t *testing.T) {
	tests := []struct {
		name     string
		diff     string
		expected []string
	}{
		{
			name: "single file",
			diff: `diff --git a/utils.go b/utils.go
--- a/utils.go
+++ b/utils.go
@@ -1,3 +1,3 @@
 package gin
-func old() {}
+func new() {}`,
			expected: []string{"utils.go"},
		},
		{
			name: "multiple files",
			diff: `diff --git a/foo.go b/foo.go
--- a/foo.go
+++ b/foo.go
@@ -1 +1 @@
-a
+b
diff --git a/bar.go b/bar.go
--- a/bar.go
+++ b/bar.go
@@ -1 +1 @@
-c
+d`,
			expected: []string{"foo.go", "bar.go"},
		},
		{
			name: "new file skips /dev/null",
			diff: `diff --git a/new.go b/new.go
--- /dev/null
+++ b/new.go
@@ -0,0 +1 @@
+package x`,
			expected: []string{"new.go"},
		},
		{
			name:     "empty diff",
			diff:     "",
			expected: nil,
		},
		{
			name: "no duplicates",
			diff: `--- a/same.go
+++ b/same.go`,
			expected: []string{"same.go"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseTouchedFiles(tt.diff)
			if len(got) != len(tt.expected) {
				t.Fatalf("got %d files %v, want %d files %v", len(got), got, len(tt.expected), tt.expected)
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("file[%d] = %q, want %q", i, got[i], tt.expected[i])
				}
			}
		})
	}
}
