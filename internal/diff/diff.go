package diff

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Extract runs git diff between parentRef and childRef in the given repo.
func Extract(repoPath, parentRef, childRef string) (string, error) {
	cmd := exec.Command("git", "diff", parentRef, childRef)
	cmd.Dir = repoPath
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git diff failed: %w", err)
	}
	return string(out), nil
}

// GatherTestContext finds _test.go files that are siblings of files touched
// by the diff, reads their contents, and returns them as context strings.
// Returns at most maxFiles test files.
func GatherTestContext(repoPath, diffOutput string, maxFiles int) ([]string, error) {
	touched := parseTouchedFiles(diffOutput)
	seen := map[string]bool{}
	var contexts []string

	for _, f := range touched {
		dir := filepath.Dir(f)
		base := filepath.Base(f)

		// Skip test files themselves
		if strings.HasSuffix(base, "_test.go") {
			continue
		}

		// Look for sibling _test.go files
		testFile := strings.TrimSuffix(base, ".go") + "_test.go"
		testPath := filepath.Join(dir, testFile)
		absPath := filepath.Join(repoPath, testPath)

		if seen[testPath] {
			continue
		}

		content, err := os.ReadFile(absPath)
		if err != nil {
			continue // test file doesn't exist, skip
		}

		seen[testPath] = true
		contexts = append(contexts, string(content))

		if len(contexts) >= maxFiles {
			break
		}
	}

	return contexts, nil
}

// parseTouchedFiles extracts file paths from diff --- a/... and +++ b/... headers.
func parseTouchedFiles(diffOutput string) []string {
	seen := map[string]bool{}
	var files []string

	scanner := bufio.NewScanner(strings.NewReader(diffOutput))
	for scanner.Scan() {
		line := scanner.Text()
		var path string
		if strings.HasPrefix(line, "+++ b/") {
			path = strings.TrimPrefix(line, "+++ b/")
		} else if strings.HasPrefix(line, "--- a/") {
			path = strings.TrimPrefix(line, "--- a/")
		}
		if path != "" && path != "/dev/null" && !seen[path] {
			seen[path] = true
			files = append(files, path)
		}
	}

	return files
}
