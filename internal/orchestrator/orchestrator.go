package orchestrator

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/lukaswoellhaf/jitgen/internal/diff"
	"github.com/lukaswoellhaf/jitgen/internal/llm"
	"github.com/lukaswoellhaf/jitgen/internal/prompt"
	"github.com/lukaswoellhaf/jitgen/internal/runner"
)

// Config holds all parameters for a jitgen run.
type Config struct {
	Repo              string
	Parent            string
	Child             string
	Output            string
	ApiKey            string
	Endpoint          string
	Model             string
	Debug             bool
	GitHubToken       string
	Owner             string
	RepoName          string
	PullRequestNumber int
	HeadRef           string
}

var dbg = log.New(io.Discard, "[debug] ", 0)

// Run executes the full jitgen pipeline: diff → prompt → LLM → test → branch → publish.
func Run(cfg Config) error {
	if cfg.Debug {
		dbg.SetOutput(os.Stderr)
	}

	// 1. Extract the diff
	fmt.Fprintf(os.Stderr, "extracting diff: %s..%s\n", cfg.Parent, cfg.Child)
	diffContent, err := diff.Extract(cfg.Repo, cfg.Parent, cfg.Child)
	if err != nil {
		return fmt.Errorf("diff: %w", err)
	}
	if diffContent == "" {
		fmt.Fprintln(os.Stderr, "no diff found between refs")
		return nil
	}
	touchedFiles := diff.ParseTouchedFiles(diffContent)
	dbg.Printf("diff size: %d bytes, touched files: %s", len(diffContent), strings.Join(touchedFiles, ", "))

	// 2. Gather test context from sibling test files
	fmt.Fprintln(os.Stderr, "gathering test context...")
	testContexts, err := diff.GatherTestContext(cfg.Repo, diffContent, 3)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: gathering context: %v\n", err)
	}
	fmt.Fprintf(os.Stderr, "found %d test file(s) as context\n", len(testContexts))

	// 3. Build prompt and call LLM
	systemPrompt, userPrompt := prompt.BuildCatchingTestPrompt(diffContent, testContexts)

	dbg.Printf("=== SYSTEM PROMPT ===\n%s", systemPrompt)
	dbg.Printf("=== USER PROMPT ===\n%s", userPrompt)

	client := llm.NewClient(cfg.ApiKey, cfg.Endpoint, cfg.Model)
	fmt.Fprintf(os.Stderr, "calling %s (%s)...\n", client.Model, client.Endpoint)
	response, err := client.Generate(systemPrompt, userPrompt)
	if err != nil {
		return fmt.Errorf("LLM call failed: %w", err)
	}

	dbg.Printf("=== RAW LLM RESPONSE ===\n%s", response)

	// 4. Extract Go code from response
	code := prompt.ExtractGoCode(response)

	dbg.Printf("=== EXTRACTED TEST CODE ===\n%s", code)

	// 5. Run differential test (parent must pass, child must fail)
	fmt.Fprintln(os.Stderr, "running differential test...")
	result, err := runner.Run(cfg.Repo, cfg.Parent, cfg.Child, code, diffContent)
	if err != nil {
		return fmt.Errorf("runner: %w", err)
	}

	// 6. Report result
	dbg.Printf("=== PARENT TEST OUTPUT ===\n%s", result.ParentOutput)
	dbg.Printf("=== CHILD TEST OUTPUT ===\n%s", result.ChildOutput)

	if result.ParentPassed {
		fmt.Fprintln(os.Stderr, "parent: PASS")
	} else {
		fmt.Fprintln(os.Stderr, "parent: FAIL")
	}

	if result.ChildPassed {
		fmt.Fprintln(os.Stderr, "child:  PASS")
	} else {
		fmt.Fprintln(os.Stderr, "child:  FAIL")
	}

	if result.ParentPassed && !result.ChildPassed {
		fmt.Fprintln(os.Stderr, "result: CATCH")
	} else {
		fmt.Fprintln(os.Stderr, "result: DISCARD")
	}

	// 7. If CATCH, create a branch and publish
	if result.IsCatch {
		branch, err := runner.CreateCatchBranch(cfg.Repo, cfg.Child, code, diffContent)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not create catch branch: %v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "catch branch: %s\n", branch)
			publishCatch(cfg, branch)
		}
	}

	// 8. Output generated test
	if cfg.Output != "" {
		if err := os.WriteFile(cfg.Output, []byte(code+"\n"), 0644); err != nil {
			return fmt.Errorf("writing output: %w", err)
		}
		fmt.Fprintf(os.Stderr, "wrote generated test to %s\n", cfg.Output)
	} else {
		fmt.Println(code)
	}

	return nil
}
