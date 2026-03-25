---
name: review-changes
description: Review current code changes for quality, reuse, and efficiency against coding standards, then record any new patterns found to CODING-STANDARDS-[LANG]-PATTERNS.md (if coding-standards is installed). Invoke at code review checkpoints before committing.
context: fork
---

# Review Changes

Review current code changes for quality, reuse, and efficiency. Always load relevant coding standards before reviewing. Document findings with file and line references, severity ratings (critical / major / minor), and specific recommendations.

## Change Scope

Identify current changes — `git diff` shows unstaged, `git diff --cached` shows staged:

```bash
git diff && git diff --cached
```

## Step 1: Load Standards

If the `coding-standards` skill is available, invoke it to load the relevant standards for the language(s) being reviewed. This loads `CODING-STANDARDS/CODING-STANDARDS.md` (shared) and any language-specific file (e.g. `CODING-STANDARDS-GO.md`). This re-loads standards to ensure freshness even if already loaded at session start. If unavailable, proceed to Step 2 without pre-loaded standards.

Note: `coding-standards` is invoked again in Step 4, but for a different purpose — Step 1 is Load Mode (read standards to inform the review); Step 4 is Code-Review Mode (record findings back as patterns). Both invocations are intentional.

## Step 2: Run simplify

Invoke the `simplify` skill.

If `simplify` is unavailable, skip to Step 3 and apply the checklist manually.

## Step 3: Apply Checklist

Using the loaded standards (or general best practices if `coding-standards` was unavailable), verify every item in the Code Review Checklist against the current changes. Also check:

- **Code quality** — Readability, maintainability, modularity
- **Test coverage** — All scenarios tested (valid inputs, invalid inputs, errors, edge cases)
- **Security** — Input validation, no injection risks, proper error handling
- **Performance** — Efficient algorithms, no unnecessary operations
- **Language conventions** — Apply the loaded standards for the project's language(s)

## Step 4: Record Patterns

If the `coding-standards` skill is available, invoke it in Code-Review Mode. It will run its own independent review of the diff against the loaded standards and record patterns from its findings. This is intentional — it complements Step 3 rather than consuming its output, applying a standards-specific lens alongside the broader checklist review. Both sets of findings feed into pattern tracking, which may produce richer patterns than either review alone. It owns all pattern tracking and promotion. If unavailable, skip this step.

## Step 5: Present Findings

Present a structured report:

```
## Code Review Findings

### Issues
- [file:line] [critical/major/minor] — [description] — [recommendation]

_If no issues: "No issues found."_

### Patterns Recorded
- [LANG] [pattern-slug] — count now N[, PROMOTED if count reached 3]

_If none: "No new patterns recorded."_

### Summary
[Overall assessment and readiness for next step]
```

## Step 6: Fix Issues

If critical or major issues were found, invoke `coops-tdd-auto` to fix them. This skill runs as a fork (isolated subagent) so `coops-tdd` cannot be used — its approval gates require a live user back-channel that is not available in a forked context.

If `coops-tdd-auto` is unavailable, include the issues in the findings report with a note that they require manual remediation.

If no critical or major issues were found, the review is complete — proceed to the next workflow step.
