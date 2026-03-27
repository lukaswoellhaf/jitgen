# Baseline Questions

## Core Architecture
- What is my target language/ecosystem (Java, Python, JS, etc.)? The approach differs significantly per language.
  > The LLM is language-agnostic, but everything around it is not. Test execution (pytest, jest, go test, etc.), scaffolding (valid file structure per language), environment isolation, and dependency resolution are all language-specific. The LLM writes the test logic; running it is an entirely different concern.

- How will I detect which tests pass/fail on the parent vs. child version of the code?
  > A generated test is only a valid catch if it **passes on the parent AND fails on the child**. Running against both versions is the core proof mechanism — a test that just fails on the child could be broken, testing something always broken, or have a bad import. The differential result is what makes it a catch. Run on parent first: if it fails, discard. If it passes, run on child: if it fails there, it's a candidate catch.

- Do I run tests against two versions simultaneously, or sequentially? How do I isolate environments?
  > Sequentially. Checkout parent, run test, record result. Checkout child, run test, record result. Isolation means the two runs don't interfere (separate working directories, virtualenvs, or containers).

- What is my unit of work — a git diff, a PR, a commit?
  > A **git diff** is the right primitive — it's the smallest unit that captures a behavioral change. The tool can be invoked at any level (PR, branch comparison, commit, pre-commit hook); the diff is just the input format.

## Test Generation
- Which LLM(s) will I use for test generation, and via which API (OpenAI, Anthropic, local models)?
- What context do I feed the LLM? (diff, surrounding code, existing tests, call graph, commit message?)
- How do I prompt the LLM to generate a catching test rather than a normal passing test?
  > When you normally ask an LLM to write a test, it produces one that verifies the *current* (post-diff) behavior — it passes on the child and is useless as a catch. The prompt must instead instruct the LLM to: (1) identify what behavior the diff *changed*, (2) anchor the expected values to the *parent* behavior, and (3) frame the diff as a suspected regression. A naively prompted LLM will describe the new behavior. The prompt has to explicitly treat the diff as the suspect.
- Do I implement the dodgy-diff workflow, the intent-aware workflow, or both?
- For intent inference: how do I extract developer intent reliably from a diff + commit message?
- How do I handle test scaffolding — does the LLM produce a complete runnable test file, or just assertions?

## Diff Risk Scoring
- Do I test every diff or only "high-risk" ones? If the latter, how do I define and compute a risk score?
- What signals indicate risk? (file churn, code complexity, touched files with no existing tests, etc.)

## CI/CD Integration
- How does this plug into existing pipelines? (GitHub Actions, GitLab CI, Jenkins, a git hook?)
  > GitHub Actions, triggered on `pull_request` events. No external CI system or server required.

- Is this a blocking gate (prevents merge) or advisory (posts a warning)?
  > Advisory. The original PR is never blocked. The engineer decides the outcome by merging (confirm) or closing (dismiss) the draft PR.
- How do I present results to the engineer? (PR comment, CLI output, dashboard?)
  > The tool creates a temp branch `jitgen/pr-{number}-{hash}` off the feature branch, pushes the generated test file there, and opens a **draft PR** (temp branch → feature branch). A short comment is posted on the original PR linking to the draft PR. The engineer sees the test as proper code with syntax highlighting, can run it locally via `git checkout`, and CI runs on the draft PR independently.

- How do I let engineers dismiss false positives with low friction?
  > The draft PR itself is the interaction surface — no custom commands or webhooks needed:
  > - **Dismiss (false positive):** close the draft PR without merging.
  > - **Confirm (true positive / keep the test):** merge the draft PR — the test lands permanently on the feature branch.
  >
  > Temp branches are cleaned up automatically via a `pull_request` event (`types: [closed]`) on the original PR. All tooling is standard `GITHUB_TOKEN` — no infrastructure beyond GitHub Actions required.

## Infrastructure & Compute

- How do I run two versions of the codebase in isolation cheaply? (Docker, ephemeral VMs, sandboxing?)
- Where does this run — developer machine, CI server, or a dedicated service?
  > GitHub Actions runner. No dedicated service or developer machine setup required.
- How do I handle compute cost? LLM calls per diff can be expensive at scale.
- Do I need to cache anything (e.g., parent-version test results)?

## Open Source Concerns
- What is the scope of the tool? A CLI? A GitHub Action? A standalone service?
  > A GitHub Action. No separate server, CLI, or infrastructure required. Installed by adding a workflow file to any repository.
- How do I make it language-agnostic vs. focusing on one ecosystem first?
- How do I handle LLM API key configuration securely?
- What's the minimum viable version that demonstrates value to early adopters?
- How do I write docs/examples so users can plug in their own test runners and LLMs?

## Validation
- How do I distinguish "the test correctly caught a bug" from "the test was just brittle"?
