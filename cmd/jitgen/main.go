package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/lukaswoellhaf/jitgen/internal/orchestrator"
)

func main() {
	repo := flag.String("repo", ".", "path to the git repository")
	parent := flag.String("parent", "", "parent git ref (required)")
	child := flag.String("child", "", "child git ref (required)")
	output := flag.String("output", "", "output file path (default: stdout)")
	debug := flag.Bool("debug", false, "show prompts, LLM response, and test output")
	flag.Parse()

	if *parent == "" || *child == "" {
		fmt.Fprintln(os.Stderr, "error: --parent and --child are required")
		flag.Usage()
		os.Exit(1)
	}

	ghRepo := requireEnv("GITHUB_REPOSITORY")
	parts := strings.SplitN(ghRepo, "/", 2)
	if len(parts) != 2 {
		fmt.Fprintf(os.Stderr, "error: invalid GITHUB_REPOSITORY: %s\n", ghRepo)
		os.Exit(1)
	}

	prNumber, err := strconv.Atoi(requireEnv("GITHUB_PR_NUMBER"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: invalid GITHUB_PR_NUMBER: %v\n", err)
		os.Exit(1)
	}

	cfg := orchestrator.Config{
		Repo:     *repo,
		Parent:   *parent,
		Child:    *child,
		Output:   *output,
		Debug:    *debug,
		ApiKey:   requireEnv("LLM_API_KEY"),
		Endpoint: requireEnv("LLM_ENDPOINT"),
		Model:    requireEnv("LLM_MODEL"),

		GitHubToken:       requireEnv("GITHUB_TOKEN"),
		Owner:             parts[0],
		RepoName:          parts[1],
		PullRequestNumber: prNumber,
		HeadRef:           os.Getenv("GITHUB_HEAD_REF"),
	}

	if err := orchestrator.Run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		fmt.Fprintf(os.Stderr, "error: %s environment variable is not set\n", key)
		os.Exit(1)
	}
	return v
}
