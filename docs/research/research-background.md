# Research-Background: Just-in-Time Catching Test Generation at Meta

> Becker, Chen, Cochran, Ghasemi, Gulati, Harman, Haluza, Honarkhah, Robert et al.
> *"Just-in-Time Catching Test Generation at Meta"*, arXiv:2601.22832, January 2026
> https://arxiv.org/abs/2601.22832
> Submitted to FSE 2026 industry track

---

## The Core Idea

Traditional automated tests are *hardening* tests: they are written, verified to pass on the current codebase, and kept in the suite permanently to catch future regressions. This paper introduces a different class — **catching tests** — which are **differential tests**: they run against both the parent (pre-diff) and child (post-diff) version of the code, and their job is to **pass on the parent but fail on the child**.

| | Hardening test | JIT catching test |
|---|---|---|
| **When generated** | Any time | At diff submission |
| **Expected result at generation** | Pass on current code | Pass on parent, fail on child |
| **Lifetime** | Permanent, in test suite | Temporary, discarded after review |
| **Purpose** | Prevent future regressions | Block a bad diff *right now* |

The key insight: the **parent is treated as ground truth** — it has already passed CI and review, so its behavior is assumed correct. The diff is the suspect. A catching test doesn't need to encode what "correct behavior" is in any general sense; it only needs to detect that the child behaves differently from the parent in a way that wasn't intended. Because the test is thrown away after the diff is reviewed, it doesn't matter if the assertion is brittle or narrow — it only has to be valid for this one moment.

---

## Results (from the paper)

Meta analyzed **22,126 generated tests** across a codebase of hundreds of millions of lines of code:

- **4×** more candidate catches than hardening-style generation
- **20×** more than coincidentally failing (random) tests
- **70%** reduction in human review load via AI-based false positive filtering
- Of **41 reported catches**, 8 confirmed true bugs — **4 would have caused serious production failures**

---

## How It Works

### 1. Trigger on high-risk diffs
Not every diff is tested. A **Diff Risk Score** prioritizes changes most likely to introduce severe regressions, keeping compute costs manageable.

### 2. Two test generation workflows

**Dodgy Diff (intent-unaware)**
The diff is treated as a suspected mutant. Tests are generated to "kill" it — produce a different observable result on parent vs. child code. No attempt to understand *why* the diff was written. Simple, high-volume, higher false positive rate (4.0% weak catches per diff).

**Intent-Aware**
An LLM infers the developer's intent from the diff, commit message, and discussion thread. It generates mutants representing specific risks ("what if this change accidentally breaks X?"). Fewer tests, higher quality — **7.9% weak catches per diff**.

### 3. AI-based false positive filtering

Two complementary assessors run before any human sees a result:

- **Rule-based (RubFake):** Fast, deterministic detection of common false positive patterns — broken mocks, type mismatches, flaky test signatures.
- **LLM-as-judge ensemble:** Multiple LLMs classify each failure with confidence scores and explanations. **98% precision** in identifying false positives.

### 4. Human review
Engineers receive only survivors of both filters. Takes a few minutes per case. Unexpected behavior → investigate. Expected behavior → dismiss.

---

## Key Distinction from Mutation Testing

Classical mutation testing evaluates *test suite quality* — existing tests are run against artificial mutants. JIT catching inverts this: the diff is the suspected mutant, and new tests are generated specifically to expose it. The goal is not coverage — it's blocking a specific bad change before merge.
