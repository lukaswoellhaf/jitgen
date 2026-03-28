package main

import (
	"flag"
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

var dbg = log.New(os.Stderr, "[debug] ", 0)

func main() {
	repo := flag.String("repo", ".", "path to the git repository")
	parent := flag.String("parent", "", "parent git ref (required)")
	child := flag.String("child", "", "child git ref (required)")
	output := flag.String("output", "", "output file path (default: stdout)")
	debug := flag.Bool("debug", false, "show prompts, LLM response, and test output")
	flag.Parse()

	if !*debug {
		dbg.SetOutput(io.Discard)
	}

	if *parent == "" || *child == "" {
		fmt.Fprintln(os.Stderr, "error: --parent and --child are required")
		flag.Usage()
		os.Exit(1)
	}

	apiKey := os.Getenv("LLM_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "error: LLM_API_KEY environment variable is not set")
		os.Exit(1)
	}
	endpoint := os.Getenv("LLM_ENDPOINT")
	if endpoint == "" {
		fmt.Fprintln(os.Stderr, "error: LLM_ENDPOINT environment variable is not set")
		os.Exit(1)
	}
	model := os.Getenv("LLM_MODEL")
	if model == "" {
		fmt.Fprintln(os.Stderr, "error: LLM_MODEL environment variable is not set")
		os.Exit(1)
	}

	// 1. Extract the diff
	fmt.Fprintf(os.Stderr, "extracting diff: %s..%s\n", *parent, *child)
	diffContent, err := diff.Extract(*repo, *parent, *child)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if diffContent == "" {
		fmt.Fprintln(os.Stderr, "no diff found between refs")
		os.Exit(0)
	}
	touchedFiles := diff.ParseTouchedFiles(diffContent)
	dbg.Printf("diff size: %d bytes, touched files: %s", len(diffContent), strings.Join(touchedFiles, ", "))

	// 2. Gather test context from sibling test files
	fmt.Fprintln(os.Stderr, "gathering test context...")
	testContexts, err := diff.GatherTestContext(*repo, diffContent, 3)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: gathering context: %v\n", err)
	}
	fmt.Fprintf(os.Stderr, "found %d test file(s) as context\n", len(testContexts))

	// 3. Build prompt and call LLM
	systemPrompt, userPrompt := prompt.BuildCatchingTestPrompt(diffContent, testContexts)

	dbg.Printf("=== SYSTEM PROMPT ===\n%s", systemPrompt)
	dbg.Printf("=== USER PROMPT ===\n%s", userPrompt)

	client := llm.NewClient(apiKey, endpoint, model)
	fmt.Fprintf(os.Stderr, "calling %s (%s)...\n", client.Model, client.Endpoint)
	response, err := client.Generate(systemPrompt, userPrompt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: LLM call failed: %v\n", err)
		os.Exit(1)
	}

	dbg.Printf("=== RAW LLM RESPONSE ===\n%s", response)

	// 4. Extract Go code from response
	code := prompt.ExtractGoCode(response)

	dbg.Printf("=== EXTRACTED TEST CODE ===\n%s", code)

	// 5. Run differential test (parent must pass, child must fail)
	fmt.Fprintln(os.Stderr, "running differential test...")
	result, err := runner.Run(*repo, *parent, *child, code, diffContent)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: runner failed: %v\n", err)
		os.Exit(1)
	}

	// 6. Report result
	dbg.Printf("=== PARENT TEST OUTPUT ===\n%s", result.ParentOutput)
	if result.ParentPassed {
		fmt.Fprintln(os.Stderr, "parent: PASS")
	} else {
		fmt.Fprintln(os.Stderr, "parent: FAIL")
	}

	dbg.Printf("=== CHILD TEST OUTPUT ===\n%s", result.ChildOutput)
	if result.ChildFailed {
		fmt.Fprintln(os.Stderr, "child:  FAIL")
	} else {
		fmt.Fprintln(os.Stderr, "child:  PASS")
	}

	if result.IsCatch {
		fmt.Fprintln(os.Stderr, "result: CATCH")
	} else {
		fmt.Fprintln(os.Stderr, "result: DISCARD")
	}

	// 7. If CATCH, create a temp branch with the test in the target repo
	if result.IsCatch {
		branch, err := runner.CreateCatchBranch(*repo, *child, code, diffContent)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not create catch branch: %v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "catch branch: %s\n", branch)
		}
	}

	// 8. Output generated test if requested
	if *output != "" {
		if err := os.WriteFile(*output, []byte(code+"\n"), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "error writing output: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "wrote generated test to %s\n", *output)
	} else {
		fmt.Println(code)
	}
}
