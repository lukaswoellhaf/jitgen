# jitgen

**Just-in-time regression test generation using LLMs.**

jitgen is a GitHub Action that automatically generates *catching tests* on pull requests. A catching test is one that **passes on the parent** (base branch) and **fails on the child** (PR head), surfacing potential regressions before they're merged.

Inspired by [*"Just-in-Time Catching Test Generation at Meta"*](https://arxiv.org/abs/2601.22832) (Becker, Chen, Cochran et al., 2026), which demonstrated that differential LLM-generated tests can catch real regressions at scale.

> **Note:** jitgen is currently a proof of concept. It works end-to-end but targets Go projects only and has known limitations (see [Current Limitations](#current-limitations)).

## How it works

When a pull request is opened, jitgen runs as a GitHub Action and:

1. **Diffs** the parent and child refs to identify changed code
2. **Gathers context** from existing test files in the affected packages
3. **Prompts an LLM** to generate a test that asserts the parent's behavior
4. **Runs the test differentially** — on both the parent and child versions using git worktrees
5. **Evaluates the result:**
   - **CATCH** — test passes on parent, fails on child → potential regression found
   - **DISCARD** — test doesn't differentiate the two versions → no signal
6. If CATCH: **pushes a branch** with the test, **opens a draft PR**, and **comments on the original PR**

The developer can then review the draft PR and either merge it (to keep the test) or close it (to dismiss as a false positive).

## Usage

Add a workflow to your repository:

```yaml
name: jitgen
on:
  pull_request:
    types: [opened, synchronize]

permissions:
  contents: write
  pull-requests: write

jobs:
  jitgen:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: lukaswoellhaf/jitgen@main
        with:
          parent-ref: ${{ github.event.pull_request.base.sha }}
          child-ref: ${{ github.event.pull_request.head.sha }}
          pr-number: ${{ github.event.pull_request.number }}
          llm-api-key: ${{ secrets.LLM_API_KEY }}
          llm-endpoint: ${{ vars.LLM_ENDPOINT }}
          llm-model: ${{ vars.LLM_MODEL }}
```

### Inputs

| Input | Required | Description |
|---|---|---|
| `parent-ref` | Yes | Parent git ref (base branch) |
| `child-ref` | Yes | Child git ref (PR head) |
| `pr-number` | Yes | Pull request number to comment on |
| `llm-api-key` | Yes | API key for the LLM provider |
| `llm-endpoint` | Yes | LLM API endpoint (OpenAI-compatible) |
| `llm-model` | Yes | Model name (e.g. `codestral-latest`) |
| `github-token` | No | Defaults to `${{ github.token }}` |

### LLM compatibility

jitgen uses the OpenAI chat completions API format (`/v1/chat/completions`). Any provider offering a compatible endpoint should work. Tested with Mistral (Codestral).

## Current Limitations

- **Go only** — test generation and execution currently targets Go projects
- **Single LLM attempt** — one completion per run, no retries or candidate selection
- **Package detection** — uses the first non-test `.go` file in the diff to determine the target package
- **Test context** — gathers sibling test files by naming convention (`utils.go` → `utils_test.go`), not by semantic relevance

## Project structure

```
cmd/jitgen/          CLI entry point
docs/                Design documents and research notes
internal/
  diff/              Git diff extraction and file parsing
  github/            GitHub API client (draft PRs, comments)
  llm/               LLM HTTP client (OpenAI-compatible)
  orchestrator/      Pipeline coordination
  prompt/            Prompt construction and response extraction
  runner/            Differential test execution via git worktrees
```

## License

[MIT](LICENSE)
