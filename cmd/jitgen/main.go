package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/lukaswoellhaf/jitgen/internal/diff"
	"github.com/lukaswoellhaf/jitgen/internal/llm"
	"github.com/lukaswoellhaf/jitgen/internal/prompt"
)

func main() {
	repo := flag.String("repo", ".", "path to the git repository")
	parent := flag.String("parent", "", "parent git ref (required)")
	child := flag.String("child", "", "child git ref (required)")
	output := flag.String("output", "", "output file path (default: stdout)")
	flag.Parse()

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

	// 2. Gather test context from sibling test files
	fmt.Fprintln(os.Stderr, "gathering test context...")
	testContexts, err := diff.GatherTestContext(*repo, diffContent, 3)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: gathering context: %v\n", err)
	}
	fmt.Fprintf(os.Stderr, "found %d test file(s) as context\n", len(testContexts))

	// 3. Build prompt and call LLM
	systemPrompt, userPrompt := prompt.BuildCatchingTestPrompt(diffContent, testContexts)

	client := llm.NewClient(apiKey, endpoint, model)
	fmt.Fprintf(os.Stderr, "calling %s (%s)...\n", client.Model, client.Endpoint)
	response, err := client.Generate(systemPrompt, userPrompt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: LLM call failed: %v\n", err)
		os.Exit(1)
	}

	// 4. Extract Go code from response
	code := prompt.ExtractGoCode(response)

	// 5. Output
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
