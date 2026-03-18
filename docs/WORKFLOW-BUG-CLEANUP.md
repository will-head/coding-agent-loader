# Bug Cleanup Workflow (11-Step)

> Interactive workflow for resolving tracked bugs from BUGS.md

**Use When:** Fixing bugs tracked in `docs/BUGS.md`

**Key Principles:**
- **No quick fixes** - every solution must be thoroughly analyzed and well-considered. No hacks or shortcuts
- **Understand first, act second** - fully understand the problem before proposing anything. Ask questions if there is any ambiguity
- **Propose before implementing** - present the solution and wait for explicit approval before writing any code. This applies at session start AND at every subsequent interaction (e.g., if new issues surface during the fix)
- **Prove before testing** - before asking the user to test, demonstrate the fix is sound: tests must pass, provide evidence and reasoning. Never run anything that could modify the host system — ask if unsure
- **Value user time** - don't ask the user to test prematurely. Get it right first
- **Bug-driven** - work items come from `docs/BUGS.md`, not phase TODO files
- **User approval required on HOST** - ask permission before ALL commands (auto-approved when `CALF_VM=true`; see [CALF_VM Auto-Approve](WORKFLOWS.md#calf_vm-auto-approve))
- **Blocking checkpoints** - each step must complete before proceeding
- **Code review mandatory** - all code/script changes reviewed before commit
- **Bug lifecycle** - resolved bugs move from `BUGS.md` to `bugs/README.md`

---

## Overview

The Bug Cleanup workflow is a variant of the Interactive workflow where work items are sourced from `docs/BUGS.md` instead of phase TODO files. It adds a dedicated analysis and proposal step before any implementation begins, ensuring solutions are well-considered and user-approved. User approvals are required at each checkpoint on HOST (auto-approved when `CALF_VM=true`; see [CALF_VM Auto-Approve](WORKFLOWS.md#calf_vm-auto-approve)).

**Target:** main branch (direct commits)
**Approvals:** Required on HOST for all commands (auto-approved when `CALF_VM=true`)
**Steps:** 11 (Interactive + analysis/proposal step)

---

## Session Start Procedure

Follow [Session Start Procedure](WORKFLOWS.md#session-start-procedure) from Shared Conventions, highlighting:
- This is the Bug Cleanup workflow (11-step, Interactive variant for bug fixes)
- Key principles: no quick fixes, understand first, propose before implementing, prove before testing, value user time
- Work items sourced from `docs/BUGS.md`

**Then:**
1. Read `docs/BUGS.md` to get list of active bugs
2. Present active bugs using [Numbered Choice Presentation](WORKFLOWS.md#numbered-choice-presentation)
3. Wait for user to select a bug
4. Read the full bug report (e.g., `docs/bugs/BUG-NNN-slug.md`) for the selected bug
5. **Analyze thoroughly** — Read all related code, surrounding context, and any referenced documentation. Understand the full picture before forming a solution
6. **Ask clarifying questions** — If there is ANY ambiguity about the bug, its scope, expected behavior, or constraints, ask questions (one at a time per [Sequential Question Presentation](WORKFLOWS.md#sequential-question-and-test-presentation)). Do not guess or assume
7. **Present proposed solution** — Explain the proposed approach with clear rationale: what will change, why it addresses the root cause, and any trade-offs. Do NOT write any code yet
8. **STOP and wait for user approval** of the approach before proceeding to implementation. The user may have further questions or request a different approach

---

## Documentation-Only Changes

For bug fixes that only affect `.md` files or code comments:

1. Make changes
2. Ask user approval to commit (auto-approved when `CALF_VM=true`)
3. Commit and push

**Skip:** tests, build, and code review for docs-only changes.

---

## Code/Script Changes (Full 11-Step Workflow)

**Each step is a blocking checkpoint.**

### Step 1: Analyze and Propose

- **Thoroughly analyze** the bug report, all related code, and surrounding context
- Understand the root cause before proposing anything — no quick fixes or hacks
- **Ask clarifying questions** if there is ANY ambiguity about the bug, its scope, expected behavior, or constraints (one at a time per [Sequential Question Presentation](WORKFLOWS.md#sequential-question-and-test-presentation))
- **Present a proposed solution** with clear rationale:
  - What will change (files, functions, approach)
  - Why it addresses the root cause (not just the symptom)
  - Any trade-offs or risks
  - How it will be tested
- **Do NOT write any code** until the approach is approved

**STOP and wait for explicit user approval** before proceeding. The user may have further questions, request clarification, or prefer a different approach.

**This principle applies throughout the entire session** — if new issues surface during implementation, testing, or review, analyze and propose a solution before acting. Never apply quick patches without user approval.

**Exception:** Read/Grep/Glob tools for searching code do not require approval.

### Step 2: Implement

- Only proceed after approach is approved in Step 1
- **Read `CODING_STANDARDS.md`** before writing any code
- **Invoke the `coops-tdd` skill first** — this is mandatory before writing any code, no exceptions
- The skill structures implementation via TDD: write failing test that reproduces the bug → implement fix → verify test passes
- Follow Go conventions and shell script best practices
- Make minimum changes needed to fix the bug
- Avoid over-engineering or adding unnecessary features
- If implementation reveals the approach needs to change, **stop and re-propose** — do not silently deviate from the approved plan

**Exception:** Read/Grep/Glob tools for searching code do not require approval.

### Step 3: Test

- **Ask user approval** before running (auto-approved when `CALF_VM=true`)
- Execute: `go test ./...`
- **Stop if tests fail** - fix issues before proceeding

All tests must pass to continue.

### Step 4: Build

- **Ask user approval** before running (auto-approved when `CALF_VM=true`)
- Execute: `go build -o calf ./cmd/calf`
- **Stop if build fails** - fix issues before proceeding

Build must succeed to continue.

### Step 5: Code Review

**Invoke the `code-review` skill.** Also verify these bug-fix-specific criteria:
- **Bug fix correctness** — Does the fix address the root cause documented in the bug report?
- **Regression risk** — Could this fix break other functionality?

### Step 6: Present Review

- Always present review findings to user
- **STOP and wait for explicit user approval** (auto-approved when `CALF_VM=true`)
- User responses like "approved", "looks good", "proceed" = approved
- Do not proceed without approval on HOST

### Step 7: Prove Fix and Present User Testing

**Before involving the user, prove the solution is sound:**

1. **Confirm all automated tests pass** — tests must be green before proceeding
2. **Provide evidence the fix addresses the root cause:**
   - Explain why the root cause is resolved (not just the symptom)
   - Show before/after analysis where applicable
   - Reference the specific code changes and how they prevent recurrence
3. **Never run anything that could modify the host system** — if unsure whether a verification step is safe, ask the user first

**Only after demonstrating soundness**, present testing instructions to the user **one by one** (not as a batch list):
- Include specific steps to verify the bug is fixed (from the bug report's "Steps to Reproduce")
- Present each test instruction and wait for the user to confirm pass/fail
- If a test fails, the user can choose to:
  - **Fix it now** - address the issue before continuing
  - **Add as a TODO** - add it to the appropriate phase TODO file for later
  - **Accept as known issue** - acknowledge and proceed
- Continue until all user testing instructions have been presented

**STOP and wait for user confirmation** on each test before presenting the next.

### Step 8: Present Final Code Review

**Invoke the `code-review` skill** as a final check after user testing. Also confirm:
- The bug is fixed (root cause addressed)
- All tests still pass after any fixes made during user testing

**STOP and wait for explicit user approval** before proceeding (auto-approved when `CALF_VM=true`).

### Step 9: Update Documentation

**Invoke the `update-docs` skill.**

### Step 10: Commit and Push

- **Ask user approval** before committing (auto-approved when `CALF_VM=true`)
- Follow [Commit Message Format](WORKFLOWS.md#commit-message-format) from Shared Conventions
- Reference the bug ID in the commit message (e.g., "Fix BUG-001: ...")
- Execute only after all previous steps complete successfully

### Step 11: Complete

Report completion status:
- Confirm bug status updated in all tracking files
- Suggest next bug from `docs/BUGS.md` if any remain
- Or suggest next steps from PLAN.md

---

## Pre-Commit Checklist

Before every commit:
- [ ] Solution proposed and user approved approach (Step 1)
- [ ] `coops-tdd` skill invoked before writing any code
- [ ] Tests pass (`go test ./...`)
- [ ] Build succeeds (`go build -o calf ./cmd/calf`)
- [ ] Code review presented and user approved (for code changes)
- [ ] Fix proven sound with evidence and reasoning before user testing
- [ ] User testing instructions presented one by one and resolved
- [ ] Final code review presented and user approved
- [ ] Bug report updated with resolution details
- [ ] Bug removed from `docs/BUGS.md` (active bugs)
- [ ] Bug status updated in `docs/bugs/README.md` (all bugs index)
- [ ] Other documentation updated (affected files)
- [ ] PLAN.md updated if bug relates to a tracked TODO
- [ ] User approved commit operation

---

## Bug Lifecycle

```
Open (in BUGS.md + bugs/README.md)
    ↓
Bug Cleanup workflow
    ↓
Resolved (removed from BUGS.md, updated in bugs/README.md)
```

**When a bug is resolved:**
1. Update the individual bug report with resolution details
2. Remove the entry from `docs/BUGS.md` (active bugs only)
3. Update status in `docs/bugs/README.md` (complete index)

---

## Important Notes

### Command Execution Policy

See [Command Execution Policy](WORKFLOWS.md#command-execution-policy) in Shared Conventions.

### Upstream Bugs

Some bugs may have resolution paths that require upstream fixes (e.g., reporting to external projects). For these:
- Document the upstream report in the bug report
- Update status to "Blocked" if waiting on upstream
- Keep in `docs/BUGS.md` until resolved
- Workarounds can be implemented as separate fixes
