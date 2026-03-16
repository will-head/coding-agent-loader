# Interactive Workflow (10-Step)

> Default workflow for direct code changes with user approvals at each step

**Use When:** Making code changes directly to the main branch with interactive approvals

**Key Principles:**
- **User approval required on HOST** - ask permission before ALL commands (auto-approved when `CALF_VM=true`; see [CALF_VM Auto-Approve](WORKFLOWS.md#calf_vm-auto-approve))
- **Blocking checkpoints** - each step must complete before proceeding
- **Code review mandatory** - all code/script changes reviewed before commit
- **Documentation-only exception** - skip tests/build/review for `.md` files only

---

## Overview

The Interactive workflow is the default workflow for making code changes directly to the main branch. On HOST, it requires explicit user approval before running any commands (git, build, tests, installs), ensuring the user maintains full control over all operations. When `CALF_VM=true`, approvals are auto-granted (see [CALF_VM Auto-Approve](WORKFLOWS.md#calf_vm-auto-approve)).

**Target:** main branch (direct commits)
**Approvals:** Required on HOST for all commands (auto-approved when `CALF_VM=true`)
**Steps:** 10 (full workflow) or simplified for docs-only

---

## Session Start Procedure

Follow [Session Start Procedure](WORKFLOWS.md#session-start-procedure) from Shared Conventions, highlighting:
- This is the Interactive workflow (10-step with approvals)
- Key principles: user approval required on HOST (auto-approved when `CALF_VM=true`), blocking checkpoints, code review mandatory
- 10 steps for code or simplified 3-step for docs-only
- Present available TODOs using [Numbered Choice Presentation](WORKFLOWS.md#numbered-choice-presentation)

---

## Documentation-Only Changes

For changes **only** to `.md` files or code comments:

1. Make changes
2. Ask user approval to commit (auto-approved when `CALF_VM=true`)
3. Commit and push

**Skip:** tests, build, and code review for docs-only changes.

---

## Code/Script Changes (Full 10-Step Workflow)

**Each step is a blocking checkpoint.**

### Step 1: Implement

- **Invoke the `coops-tdd` skill first** — this is mandatory before writing any code, no exceptions
- The skill structures implementation via TDD: write failing test → implement minimum code → refactor
- Follow Go conventions and shell script best practices
- Make minimum changes needed to accomplish the goal
- Avoid over-engineering or adding unnecessary features

**Exception:** Read/Grep/Glob tools for searching code do not require approval.

### Step 2: Test

- **Ask user approval** before running (auto-approved when `CALF_VM=true`)
- Execute: `go test ./...`
- **Stop if tests fail** - fix issues before proceeding

All tests must pass to continue.

### Step 3: Build

- **Ask user approval** before running (auto-approved when `CALF_VM=true`)
- Execute: `go build -o calf ./cmd/calf`
- **Stop if build fails** - fix issues before proceeding

Build must succeed to continue.

### Step 4: Code Review

- **Invoke the `simplify` skill** — reviews changed code for reuse, quality, and efficiency

### Step 5: Present Review

- Present review findings to user
- Fix all issues found before proceeding
- Note any new TODOs discovered — must add to PLAN.md
- **STOP and wait for explicit user approval** (auto-approved when `CALF_VM=true`)
- User responses like "approved", "looks good", "proceed" = approved
- Do not proceed without approval on HOST

### Step 6: Present User Testing Instructions

Present testing instructions to the user **one by one** (not as a batch list):
- Present the first test instruction and wait for the user to confirm pass/fail
- Only after the current test is resolved, present the next test instruction
- If a test fails, the user can choose to:
  - **Fix it now** - address the issue before continuing
  - **Add as a TODO** - add it to the appropriate phase TODO file for later
  - **Accept as known issue** - acknowledge and proceed
- Continue until all user testing instructions have been presented

**STOP and wait for user confirmation** on each test before presenting the next.

### Step 7: Present Final Code Review

- **Invoke the `simplify` skill** again as a final check after user testing
- Confirm all tests still pass after any fixes made during user testing

**STOP and wait for explicit user approval** before proceeding (auto-approved when `CALF_VM=true`).

### Step 8: Update Documentation

Update affected documentation files:
- `README.md` - if user-facing changes
- `CLAUDE.md` (AGENTS.md) - if workflow or rules changed
- `docs/SPEC.md` - if technical spec changed
- `docs/architecture.md` - if architecture changed
- `docs/cli.md` - if CLI commands changed
- `docs/bootstrap.md` - if setup changed
- `docs/plugins.md` - if plugins affected
- `docs/roadmap.md` - if roadmap changed
- Inline comments in changed code files

**Never modify `docs/adr/*` or `docs/prd/*`** - ADRs and PRDs are immutable historical records.

**Always update PLAN.md and phase TODO files** - follow [TODO → DONE Movement](WORKFLOWS.md#todo--done-movement) rules from Shared Conventions. Also add new TODOs discovered during implementation and update PLAN.md phase status.

### Step 9: Commit and Push

- **Ask user approval** before committing (auto-approved when `CALF_VM=true`)
- Follow [Commit Message Format](WORKFLOWS.md#commit-message-format) from Shared Conventions
- Execute only after all previous steps complete successfully

### Step 10: Complete

Report completion status and suggest next steps using [Next Workflow Guidance](WORKFLOWS.md#next-workflow-guidance).

---

## Pre-Commit Checklist

Before every commit:
- [ ] `coops-tdd` skill invoked before writing any code
- [ ] Tests pass (`go test ./...`)
- [ ] Build succeeds (`go build -o calf ./cmd/calf`)
- [ ] Code review presented and user approved (for code changes)
- [ ] User testing instructions presented one by one and resolved
- [ ] Final code review presented and user approved
- [ ] Documentation updated (affected files)
- [ ] PLAN.md updated with current project status
- [ ] Completed TODOs moved from phase TODO file to phase DONE file
- [ ] New TODOs added to appropriate phase TODO file
- [ ] User approved commit operation

---

## Important Notes

### Command Execution Policy

See [Command Execution Policy](WORKFLOWS.md#command-execution-policy) in Shared Conventions.

### TODO Tracking

See [PLAN.md is Source of Truth](WORKFLOWS.md#planmd-is-source-of-truth) and [TODO → DONE Movement](WORKFLOWS.md#todo--done-movement) in Shared Conventions.
