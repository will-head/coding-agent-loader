# Review Workflow (8-Step)

> Autonomous code review of worktree branches with direct issue resolution and automated testing

**Use When:** Items in STATUS.md "Needs Review" section

**Key Principles:**
- **No permission needed** — fully autonomous operation
- **Work in existing worktree** — review and fix files directly at `.claude/worktrees/<name>`; never `git checkout`
- **Claim before starting** — move item to "Reviewing" in STATUS.md on main before touching the worktree
- **Fix directly** — resolve minor and major issues on the worktree branch; only send back architectural issues
- **Automated tests gate handoff** — tests and build must pass before moving to Needs Testing
- **Keep busy** — after completing an item, check for more before exiting

---

## Overview

The Review workflow picks up items from "Needs Review", reviews and fixes code directly in the existing worktree, runs tests and build, then either advances to "Needs Testing" (passing) or returns to "Needs Rework" (failing or architectural). It never touches main code — only STATUS.md is updated on main.

**Target:** Existing `implement/<feature-name>` worktree
**Approvals:** Not required (autonomous)
**Steps:** 8

---

## What to Fix vs. What to Send Back

### Fix Directly (in worktree)
- Code quality issues (naming, duplication, organisation)
- Missing error handling or validation
- Security vulnerabilities
- Test gaps
- Documentation mismatches
- Performance issues
- Style and convention violations
- Missing dependency checks
- Minor logic errors with clear fixes

### Send Back to Implement (Needs Rework)
- Fundamental design approach is wrong
- Major structural reorganisation needed
- Requirements misunderstood
- Breaking changes to public interfaces needing design discussion
- Trade-offs requiring clarification that could change scope

**Rule of thumb:** If you can fix it confidently without changing the overall design, fix it. If fixing it would change the fundamental approach, send back.

---

## Step-by-Step Process

### Step 1: Read Queue

On main, read STATUS.md "Needs Review" section. Find the first unclaimed item (not currently in "Reviewing").

**If queue empty:** Report "Nothing to review — Needs Review queue is empty." Suggest next workflow per [Next Workflow Guidance](WORKFLOWS.md#next-workflow-guidance). Exit.

**If multiple items:** Present using [Numbered Choice Presentation](WORKFLOWS.md#numbered-choice-presentation).

### Step 2: Claim Item

On main, move item from "Needs Review" to "Reviewing". Commit and push immediately:

```bash
git add STATUS.md
git commit -m "Claim <feature-name> for review"
git push
```

**If push fails** (race condition): pull, find next unclaimed item, repeat from Step 1.

### Step 3: Read Requirements and Standards

1. **Read original requirements** — find and read the full TODO from the phase TODO file referenced in STATUS.md. Understand what was asked for, acceptance criteria, and constraints. This anchors the review against intent, not just code quality.

2. **Load coding standards** — invoke the `coding-standards` skill to load relevant files from `CODING-STANDARDS/` as the review baseline.

### Step 4: Review Code

Review all changed files in the worktree comprehensively across 10 areas. Use Read/Grep tools with the worktree path:

1. **Code Quality** — readability, modularity, naming, comments, duplication
2. **Architecture** — patterns, separation of concerns, interface design, requirements match
3. **Correctness** — logic errors, edge cases, race conditions, off-by-ones
4. **Error Handling** — propagation, messages, recovery, wrapping
5. **Security** — input validation, injection vulnerabilities, OWASP Top 10
6. **Performance** — algorithmic complexity, resource usage, unnecessary allocations
7. **Testing** — coverage of valid inputs, invalid inputs, error scenarios, edge cases
8. **Documentation** — GoDoc on exported identifiers, accuracy, completeness
9. **Language Conventions** — Go idioms, shell best practices, style consistency
10. **Dependencies** — availability checks before use, version management

Document findings with:
- File path and line number (relative to worktree root)
- Severity: critical / major / minor
- Classification: fixable / architectural
- Recommendation

### Step 5: Fix Issues

For all fixable issues, edit files directly in the worktree using full paths. Fix methodically, one at a time.

**Invoke the `coops-tdd-auto` skill before writing any new or modified code — no exceptions.** This applies to all fixes, including test additions and corrections.

If a "fixable" issue turns out to require rethinking the design: reclassify as architectural, document it, and proceed to Step 7 (architectural path).

Track what was fixed for the commit message.

### Step 6: Test and Build

Run from the worktree directory:

```bash
# From within worktree, or use -C flag from main:
go test ./...
go build -o calf ./cmd/calf
which staticcheck >/dev/null 2>&1 || { echo "staticcheck not installed. Install with: go install honnef.co/go/tools/cmd/staticcheck@latest"; exit 1; }
staticcheck ./...
```

**If staticcheck is not installed:** Stop. Report the error and ask the user to install it.

**If staticcheck reports errors:** Fix all errors before proceeding. Treat staticcheck failures the same as test failures.

**If tests fail after fixes:** Attempt to fix within this step. If unable to fix within reasonable effort, this is a "Build/Test" rework. Proceed to Step 7 (rework path) with failure details.

**If all pass:** Commit fixes on the worktree branch:

```bash
git -C .claude/worktrees/<name> add -A
git -C .claude/worktrees/<name> commit -m "$(cat <<'EOF'
Address review findings for <feature-name>

- <list key fixes>

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>
EOF
)"
git -C .claude/worktrees/<name> push
```

### Step 7: Evaluate and Update STATUS.md

**Path A — All issues fixed, tests and build pass:**

On main, update STATUS.md:
- Move item from "Reviewing" to "Needs Testing"
- Add brief test scope note (what the tester should exercise)

```
## Needs Testing
- <feature-name> | .claude/worktrees/<name> | <description> | Tests: <brief scope>
```

Commit and push on main.

**Path B — Architectural issues or test/build failures remain:**

On main, update STATUS.md:
- Move item from "Reviewing" to "Needs Rework"
- Include failure type prefix and clear feedback

```
## Needs Rework
- <feature-name> | .claude/worktrees/<name> | <description> | Review: <what needs redesigning>
- <feature-name> | .claude/worktrees/<name> | <description> | Build: go build fails — <error summary>
- <feature-name> | .claude/worktrees/<name> | <description> | Test: TestFoo fails — <error summary>
```

Commit and push on main.

**Record new patterns** if this review surfaced recurring issues — write them to `CODING-STANDARDS/CODING-STANDARDS-[LANG]-PATTERNS.md`. If any pattern reaches count ≥ 3, invoke the `coding-standards` skill to promote it. Commit on main.

### Step 8: Check Queue for More

Read STATUS.md "Needs Review" again. If unclaimed items exist, loop back to Step 1 to pick up the next item. Otherwise suggest next workflow per [Next Workflow Guidance](WORKFLOWS.md#next-workflow-guidance).

---

## Pre-Handoff Checklist

- [ ] Item claimed in STATUS.md before starting review
- [ ] Original requirements read from phase TODO file
- [ ] Coding standards reviewed
- [ ] Comprehensive review completed (all 10 areas)
- [ ] Findings documented with file path, severity, classification
- [ ] `coops-tdd-auto` skill invoked before writing any fixes
- [ ] All fixable issues resolved directly in worktree
- [ ] Tests pass after fixes (`go test ./...`)
- [ ] Build succeeds after fixes (`go build -o calf ./cmd/calf`)
- [ ] staticcheck passes after fixes (`staticcheck ./...`)
- [ ] Fixes committed and pushed to worktree branch
- [ ] Recurring patterns recorded to CODING-STANDARDS-[LANG]-PATTERNS.md; promotion candidates flagged
- [ ] STATUS.md updated (Needs Testing or Needs Rework) on main
- [ ] STATUS.md changes committed and pushed

---

## Edge Cases

**Worktree path missing:** The worktree directory referenced in STATUS.md doesn't exist. Report to user. Move item to "Needs Rework" with note "worktree missing — needs re-implementation". Do not attempt to re-create the worktree.

**Review fixes break unrelated tests:** Investigate whether the failure is pre-existing (check git log in worktree) or introduced by fixes. If pre-existing, document as separate issue. If introduced by fixes, resolve before proceeding.

**Main has diverged significantly:** Note in the Needs Testing entry that the feature branch may need to be rebased before integration. Do not rebase during review — that is the Integrate workflow's concern.

**Fixable issue becomes architectural mid-fix:** Stop fixing. Revert partial changes if they make the code worse. Document as architectural in Step 7 rework path.

**All issues are architectural:** No commits needed. Move directly to Needs Rework with detailed feedback.

**go test hangs:** Kill after a reasonable timeout. Document as "Test: go test ./... hangs — possible deadlock in [package]" in Needs Rework.
