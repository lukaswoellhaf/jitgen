package prompt

import (
	"fmt"
	"strings"
)

const systemPrompt = `You are a regression test engineer. Your job is to write a catching test: a test that PASSES on the parent (pre-diff) version of the code and FAILS on the child (post-diff) version.

Rules:
- The parent version is correct. The diff is a suspected regression.
- Your test must assert the PARENT's behavior — the behavior BEFORE the diff was applied.
- Generate a single, complete, runnable Go test file. Include package declaration and all imports.
- Use the same package name, assertion library, and style as the existing tests provided.
- The test function name must start with "TestJitGen" to be identifiable.
- Do NOT test the new (child) behavior. Test the OLD (parent) behavior that the diff breaks.
- Output ONLY the Go code inside a single fenced code block. No explanation outside the code block.`

// BuildCatchingTestPrompt constructs system and user prompts for catching test generation.
func BuildCatchingTestPrompt(diffContent string, testContexts []string) (string, string) {
	var b strings.Builder

	b.WriteString("## Git Diff (suspected regression)\n\n```diff\n")
	b.WriteString(diffContent)
	b.WriteString("\n```\n\n")

	if len(testContexts) > 0 {
		b.WriteString("## Existing test files (use as style reference)\n\n")
		for _, ctx := range testContexts {
			b.WriteString(fmt.Sprintf("```go\n%s\n```\n\n", ctx))
		}
	}

	b.WriteString("Generate a catching test that passes on the parent (before the diff) and fails on the child (after the diff).")

	return systemPrompt, b.String()
}
