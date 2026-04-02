# Improvements

Tracking future improvements beyond the PoC scope.

---

## 1. Test context gathering limited to Go unit test conventions

**Current behavior:** `GatherTestContext` finds sibling test files by naming convention — `utils.go` → `utils_test.go`. This works because Go unit tests follow the `<file>_test.go` pattern.

**Problem:** Integration tests, end-to-end tests, or tests organized in separate directories (e.g. `test/`, `integration/`, `e2e/`) won't be discovered. Projects that group tests by feature rather than by source file will also be missed.

**Ideas:**
- Scan the entire package directory for any `_test.go` files, not just the name-matched one
- Support configurable test directory patterns (e.g. `test/`, `integration/`)
- Language-agnostic: when extending beyond Go, each language will need its own test discovery strategy

---

## 2. Test file selection lacks semantic relevance

**Current behavior:** We take the first 3 test files that match by filename. If the diff touches many files, we may pick up test files that aren't relevant to the actual behavioral change.

**Problem:** The LLM gets test context that doesn't help it understand the regression. For example, a diff in `router.go` might benefit more from `routergroup_test.go` (which tests related routing behavior) than from `router_test.go` (which might only test initialization).

**Ideas:**
- **Call graph analysis:** Parse the diff to extract changed/called function names, then search test files for references to those functions. Prioritize test files that exercise the affected call paths.
- **Symbol-aware ranking:** Use Go's AST parser to find which functions are modified in the diff, then rank test files by how many of those functions they reference.
- **Weighted context window:** Instead of a hard cap of 3 files, use a token budget and prioritize by relevance score.
- **Include the changed function's callers:** Walk up the call stack to find which higher-level functions depend on the changed code, and include tests for those too.

---

## 3. Single LLM completion — no retry or candidate selection

**Current behavior:** We request a single completion from the LLM (`n=1` default) and use whatever it returns. If the generated test doesn't catch the regression (DISCARD), the run is over.

**Problem:** LLM output is non-deterministic. A single attempt may produce a weak or incorrect test even when the model is capable of generating a good one. One unlucky generation means a missed catch.

**Ideas:**
- **Multiple candidates (`n > 1`):** Request several completions in a single API call, run the differential test on each, and keep whichever catches. The API supports this via the `n` parameter in the request body.
- **Sequential retries with feedback:** If the first attempt is a DISCARD, retry with the test output appended to the prompt so the LLM can learn from its failure.
- **Temperature variation:** Try different temperature settings across attempts to balance creativity vs. precision.
- **Cost/latency tradeoff:** More candidates = higher API cost and longer runtime. Could make the number of attempts configurable (e.g. `--attempts 3`).

---

## 4. Generated test always goes into a separate file

**Current behavior:** The catching test is always written to its own file (`jitgen_catch_test.go`). This creates a new file in the package directory regardless of whether a suitable test file already exists.

**Problem:** In most projects, adding a regression test to an existing test file is more natural and maintainable. A separate file feels foreign and may be dismissed by reviewers. It also misses the opportunity to place the test near related test cases, making it easier to understand the context.

**Ideas:**
- Identify the most relevant existing test file (related to improvement #2) and append the generated test function to it instead of creating a new file.
- Fall back to a separate file only when no suitable test file exists.
- When appending, strip the `package` declaration and `import` block from the generated code, merging imports into the existing file's imports.

---

## 5. Target package directory based on first file only

**Current behavior:** `targetPkgDir` returns the directory of the first non-test `.go` file touched by the diff. This determines where the test file is placed and where `go test` runs.

**Problem:** If the diff touches files across multiple packages, we only consider the first one. The most important behavioral change might be in a different package than the first file listed in the diff. This is related to improvement #2 — better semantic relevance would also improve target directory selection.

**Ideas:**
- If the diff spans multiple packages, run the catching test in each affected package separately.
- Rank packages by the size/significance of the changes (e.g. lines changed, functions modified) and prioritize the most impacted one.
- Let the LLM indicate which package it's targeting in its response, and validate that it matches a touched package.
