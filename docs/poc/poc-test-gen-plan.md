# PoC Test Gen Plan

## Target Setup

- **Test subject:** fork of [gin-gonic/gin](https://github.com/gin-gonic/gin) (Go)
- **Validation strategy:** manually introduce a known regression in a branch, open a PR, verify jitgen catches it
- **LLM:** Mistral Codestral via API key
- **Execution command:** `go test ./...`
- **Implementation language:** Go (single static binary, packaged as a Docker-based GitHub Action)

---

## Phase 1 — Core Loop (local script, no GitHub)

**Goal:** validate that the LLM can generate a test that passes on the parent and fails on the child.

**Inputs:**
- A git diff (the suspected regression)
- A few existing `_test.go` files from gin as style context

**Steps:**
1. Call Codestral with a catching-test prompt (diff + context → generate test)
2. Check out parent commit, inject generated test file, run `go test`, record result
3. Check out child commit, run `go test` again, record result
4. Output: **candidate catch** (pass on parent + fail on child) or **discarded**

**Success criteria:** at least one valid catch generated against the known regression.

---

## Phase 2 — GitHub Actions + Draft PR Flow

**Goal:** wrap Phase 1 in GitHub Actions and implement the full engineer-facing UX.

**Steps:**
1. Wrap the Phase 1 script in a GitHub Actions workflow triggered on `pull_request`
2. Create temp branch `jitgen/pr-{number}-{hash}` off the feature branch
3. Push the generated test file to the temp branch
4. Open a draft PR (temp branch → feature branch)
5. Post a short comment on the original PR linking to the draft PR

**Dismiss/confirm:**
- Close draft PR → dismissed (false positive)
- Merge draft PR → test lands permanently on the feature branch

**Success criteria:** full flow works on the gin fork — catch reported, engineer can dismiss or confirm via the draft PR.
