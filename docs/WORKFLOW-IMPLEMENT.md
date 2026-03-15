# Implement Workflow (9-Step)

> Autonomous implementation of refined TODOs and rework feedback in isolated worktrees

**Use When:** Items in STATUS.md "Needs Rework" or "Refined" sections

**Key Principles:**
- **Needs Rework first** — always drain the rework queue before picking up new work
- **No permission needed** — fully autonomous operation for tests, builds, and commits
- **Worktree isolation** — all changes made in `implement/<feature-name>` worktree, never on main
- **Claim before starting** — move item to "Implementing" in STATUS.md on main before entering worktree
- **TDD required** — write test first, then implementation
- **Self-review before handoff** — review against requirements before moving to Needs Review
- **Keep busy** — after completing an item, check for more before exiting

---

## Overview

The Implement workflow picks up refined TODOs and rework items, implements them with TDD in an isolated worktree, self-reviews, and hands off to the Review workflow via STATUS.md. It never touches main directly. STATUS.md claim/update steps happen on main before and after worktree work.

**Target:** `implement/<feature-name>` worktree branch
**Approvals:** Not required (autonomous)
**Steps:** 9

---

## Queue Priority

Always process **Needs Rework before Refined**:

1. Check STATUS.md "Needs Rework" — process first if any unclaimed items
2. Check STATUS.md "Refined" — process if Needs Rework is empty

This ensures in-flight work is never stalled by new work.

---

## Step-by-Step Process

### Step 1: Read Queue

On main, read STATUS.md. Find the first unclaimed item:
- "Needs Rework" first (any item not in "Implementing")
- "Refined" second (any item not in "Implementing")

**If queue empty:** Report "Nothing to implement — Needs Rework and Refined queues are empty." Suggest workflow 4 (Refine). Exit.

**If multiple items:** Present using [Numbered Choice Presentation](WORKFLOWS.md#numbered-choice-presentation). Pick first.

### Step 2: Claim Item

On main, move the item from its current section to "Implementing" in STATUS.md. Commit and push this change immediately:

```bash
# Example: claiming a Refined item
# Before: entry in "## Refined"
# After:  entry in "## Implementing" with worktree path added

git add STATUS.md
git commit -m "Claim <feature-name> for implementation"
git push
```

**If push fails** (race condition — another agent claimed first): pull, find next unclaimed item, repeat from Step 1.

### Step 3: Read Requirements and Standards

Two sub-steps:

1. **Read requirements** — for Refined items: read the full TODO from the phase TODO file. For Rework items: read the Needs Rework feedback from STATUS.md plus the original TODO. Note acceptance criteria and constraints.

2. **Read coding standards** — read `CODING_STANDARDS.md` before implementing.

### Step 4: Enter or Resume Worktree

**For Refined (new work):**
```
EnterWorktree implement/<feature-name>
```
If name conflicts with existing worktree, use `implement/<feature-name>-2` etc. Verify worktree created successfully before proceeding.

**For Needs Rework (existing worktree):**
Do NOT call `EnterWorktree` — the worktree already exists. Work directly in `.claude/worktrees/<name>` using file paths. Read/Edit/Grep tools use the full path. Git commands use `git -C .claude/worktrees/<name>`.

**Edge case — rework worktree missing:** If STATUS.md references a worktree path that no longer exists on disk, stop. Update STATUS.md: move item to "Needs Rework" with note "worktree missing — needs re-implementation from scratch". Pick up the next item.

### Step 5: Implement (TDD)

1. Write failing test first (red)
2. Implement minimum code to pass test (green)
3. Refactor if needed without breaking tests
4. For rework: address each piece of feedback from STATUS.md specifically
5. For rework: do not change behaviour beyond what the feedback requests

**If approach needs to change mid-implementation:** Stop. Do not proceed with a wrong approach. Update STATUS.md: move to "Needs Rework" with note describing the blocker. Exit worktree. Suggest workflow 4 (Refine) to clarify requirements.

### Step 6: Test

```bash
go test ./...
```

**For new work (inside EnterWorktree session):** Run `go test ./...` directly — the CWD is already the worktree.

**For Needs Rework (file path access, no EnterWorktree):** Run from main using `cd .claude/worktrees/<name> && go test ./...`.

Must pass before proceeding. Test all scenarios: valid inputs, invalid inputs, missing dependencies, auth failures, existing state, network failures, edge cases.

**If tests fail:** Fix within this step. Do not move forward with failing tests. If unable to fix, update STATUS.md: move to "Needs Rework" with note "Tests failing after implementation: [details]". Exit worktree.

### Step 7: Build and Static Analysis

```bash
go build -o calf ./cmd/calf
```

**Run from within the worktree.** Must succeed before proceeding.

**If build fails:** Fix within this step. Do not move forward with a broken build.

Then run static analysis:

```bash
which staticcheck >/dev/null 2>&1 || { echo "staticcheck not installed. Install with: go install honnef.co/go/tools/cmd/staticcheck@latest"; exit 1; }
staticcheck ./...
```

**If staticcheck is not installed:** Stop. Report the error and ask the user to install it. Do not proceed without it.

**If staticcheck reports errors:** Fix all errors before proceeding. Do not move forward with staticcheck failures.

### Step 8: Self-Review

Review implementation against the original requirements from Step 3 across 10 quality areas:

1. **Code Quality** — readability, modularity, naming, comments
2. **Architecture** — patterns, separation of concerns, interface design
3. **Correctness** — logic errors, edge cases, race conditions, requirements match
4. **Error Handling** — propagation, messages, recovery, wrapping
5. **Security** — input validation, injection vulnerabilities, OWASP Top 10
6. **Performance** — complexity, resource usage, caching opportunities
7. **Testing** — coverage, quality, missing scenarios
8. **Documentation** — accuracy, completeness, user-facing changes
9. **Language Conventions** — Go idioms, shell best practices, style
10. **Dependencies** — tool availability checks, version management

Fix all issues found. Re-run tests and build after fixes. Self-review is complete when all 10 areas assessed, all issues resolved, tests pass, build succeeds.

### Step 9: Commit, Exit Worktree, Update STATUS.md

**For new work (entered via EnterWorktree):**

1. Commit all changes on the feature branch:
```bash
git add -A
git commit -m "$(cat <<'EOF'
Implement <feature-name>

<Summary of what was implemented>

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>
EOF
)"
git push -u origin implement/<feature-name>
```

2. Exit the worktree:
```
ExitWorktree keep
```

**For rework (file path access, no EnterWorktree):**

1. Commit from the worktree directory using `git -C`:
```bash
git -C .claude/worktrees/<name> add -A
git -C .claude/worktrees/<name> commit -m "$(cat <<'EOF'
Address review feedback for <feature-name>

<Summary of what was changed and why>

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>
EOF
)"
git -C .claude/worktrees/<name> push
```

**Update STATUS.md on main (both cases):**

Move item from "Implementing" to "Needs Review":

```bash
git add STATUS.md
git commit -m "Move <feature-name> to Needs Review"
git push
```

**Check queue for more items.** If unclaimed items exist in Needs Rework or Refined, loop back to Step 1.

---

## Pre-Handoff Checklist

- [ ] Item claimed in STATUS.md before starting work
- [ ] Requirements read from phase TODO file
- [ ] Coding standards reviewed
- [ ] Tests written first (TDD)
- [ ] Tests pass (`go test ./...`)
- [ ] Build succeeds (`go build -o calf ./cmd/calf`)
- [ ] staticcheck passes (`staticcheck ./...`)
- [ ] Self-review completed (all 10 areas)
- [ ] All self-review issues fixed and re-tested
- [ ] Changes committed and pushed on feature branch
- [ ] ExitWorktree keep called (for new work)
- [ ] STATUS.md updated to "Needs Review" on main
- [ ] STATUS.md changes committed and pushed

---

## Edge Cases

**Name conflict on EnterWorktree:** Use `implement/<name>-2`, `-3` etc. Update STATUS.md worktree path accordingly.

**Rework worktree missing:** Report and move to "Needs Rework" with note. Do not attempt re-implementation without clarification.

**Tests fail after rework:** Do not move to Needs Review. Update STATUS.md "Needs Rework" with new feedback note. Exit. Let Review agent or user triage next.

**Approach changed mid-implementation:** Stop, update STATUS.md with blocker note, suggest Refine workflow.

**Queue claimed by another agent between check and push:** Pull, re-read STATUS.md, pick next unclaimed item.
