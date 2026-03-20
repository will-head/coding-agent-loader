---
name: code-review
description: Review current code changes for quality, reuse, and efficiency against CODING-STANDARDS/, then record any new patterns found to CODING-STANDARDS-[LANG]-PATTERNS.md. Use at code review checkpoints in the Interactive workflow (steps 4 and 7) and Bug Cleanup workflow (steps 5 and 8). Invoke whenever a workflow step requires a structured code review before committing.
context: fork
---

# Code Review

Review current code changes for quality, reuse, and efficiency. Always load relevant files from `CODING-STANDARDS/` before reviewing. Document findings with file and line references, severity ratings (critical / major / minor), and specific recommendations.

## Change Scope

Identify current changes — `git diff` shows unstaged, `git diff --cached` shows staged:

```bash
git diff && git diff --cached
```

## Step 1: Load Standards

Invoke the `coding-standards` skill to load the relevant standards for the language(s) being reviewed. This loads `CODING-STANDARDS/CODING-STANDARDS.md` (shared) and any language-specific file (e.g. `CODING-STANDARDS-GO.md`).

## Step 2: Run simplify

Invoke the `simplify` skill.

If `simplify` is unavailable, skip to Step 3 and apply the checklist manually.

## Step 3: Apply Checklist

Using the loaded standards, verify every item in the Code Review Checklist against the current changes. Also check:

- **Code quality** — Readability, maintainability, modularity
- **Test coverage** — All scenarios tested (valid inputs, invalid inputs, errors, edge cases)
- **Security** — Input validation, no injection risks, proper error handling
- **Performance** — Efficient algorithms, no unnecessary operations
- **Go conventions** — Idiomatic Go, stdlib over custom implementations, GoDoc on all exported identifiers; run `staticcheck ./...` and `go test ./...`
- **Shell script best practices** — Proper quoting, dependency checks, no `eval`, errors never suppressed

## Step 4: Record Patterns

After completing the review, for each issue found:

1. Identify the language (`GO`, `SH`, etc.)
2. Open `CODING-STANDARDS/CODING-STANDARDS-[LANG]-PATTERNS.md` (create it if it doesn't exist)
3. Check whether a semantically similar pattern already exists — "avoid eval" and "don't use eval in scripts" are the same pattern; "avoid eval" and "sanitise before eval" are not
   - If match found: increment its count and add a new example with today's date and file:line
   - If no match: add a new entry (count starts at 1)
4. If any pattern's count has reached 3, flag it in the report — the `coding-standards` skill should be invoked to promote it

## Step 5: Present Findings

Present a structured report:

```
## Code Review Findings

### Issues
- [file:line] [critical/major/minor] — [description] — [recommendation]

_If no issues: "No issues found."_

### Patterns Recorded
- [LANG] [pattern-slug] — count now N[, PROMOTION CANDIDATE if count ≥ 3]

_If none: "No new patterns recorded."_

### Summary
[Overall assessment and readiness for next step]
```
