package github

import (
	"context"

	"github.com/google/go-github/v72/github"
)

// Client wraps the go-github client for jitgen operations.
type Client struct {
	gh    *github.Client
	owner string
	repo  string
}

func New(token, owner, repo string) *Client {
	return &Client{
		gh:    github.NewClient(nil).WithAuthToken(token),
		owner: owner,
		repo:  repo,
	}
}

// CreateDraftPR creates a draft pull request and returns its number and HTML URL.
func (c *Client) CreateDraftPR(ctx context.Context, head, base, title, body string) (int, string, error) {
	pr, _, err := c.gh.PullRequests.Create(ctx, c.owner, c.repo, &github.NewPullRequest{
		Title: &title,
		Head:  &head,
		Base:  &base,
		Body:  &body,
		Draft: github.Ptr(true),
	})
	if err != nil {
		return 0, "", err
	}
	return pr.GetNumber(), pr.GetHTMLURL(), nil
}

// CreateComment posts a comment on a pull request.
func (c *Client) CreateComment(ctx context.Context, prNumber int, body string) error {
	_, _, err := c.gh.Issues.CreateComment(ctx, c.owner, c.repo, prNumber, &github.IssueComment{
		Body: &body,
	})
	return err
}
