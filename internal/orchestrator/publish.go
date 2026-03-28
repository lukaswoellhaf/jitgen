package orchestrator

import (
	"context"
	"fmt"
	"os"

	"github.com/lukaswoellhaf/jitgen/internal/github"
	"github.com/lukaswoellhaf/jitgen/internal/runner"
)

// publishCatch pushes the catch branch and opens a draft PR if running in GitHub Actions.
func publishCatch(cfg Config, branch string) {
	fmt.Fprintln(os.Stderr, "pushing catch branch...")
	if err := runner.PushCatchBranch(cfg.Repo, branch); err != nil {
		fmt.Fprintf(os.Stderr, "warning: push failed: %v\n", err)
		return
	}
	fmt.Fprintf(os.Stderr, "pushed %s to origin\n", branch)

	ctx := context.Background()
	gh := github.New(cfg.GitHubToken, cfg.Owner, cfg.RepoName)

	title := fmt.Sprintf("jitgen: catching test for #%d", cfg.PullRequestNumber)
	body := fmt.Sprintf("Auto-generated catching test from jitgen.\n\n"+
		"**Close** this PR to dismiss (false positive).\n"+
		"**Merge** to keep the test on `%s`.", cfg.HeadRef)

	draftNum, draftURL, err := gh.CreateDraftPR(ctx, branch, cfg.HeadRef, title, body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not create draft PR: %v\n", err)
		return
	}
	fmt.Fprintf(os.Stderr, "draft PR: %s\n", draftURL)

	comment := fmt.Sprintf(
		"**jitgen** found a potential regression and generated a catching test.\n\n"+
			"Review it here: %s (draft PR #%d)\n\n"+
			"- **Close** the draft PR to dismiss\n"+
			"- **Merge** the draft PR to keep the test",
		draftURL, draftNum)
	if err := gh.CreateComment(ctx, cfg.PullRequestNumber, comment); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not comment on PR: %v\n", err)
	}
}
