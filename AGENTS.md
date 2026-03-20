# Agent Context

> **Source of truth for all agent instructions.** This codebase is agent-agnostic — AGENTS.md is the single file to edit.
> - `CLAUDE.md` must only ever contain `@AGENTS.md` (Claude Code import). Do not edit it directly.
> - `.cursorrules` is a symlink to this file. Do not replace it with a copy.

## Project

**CALF** (Coding Agent Loader) - VM-based sandbox for AI coding agents in Tart macOS VMs.

**Stack:** Go + Charm (bubbletea/lipgloss/bubbles) + Cobra + Viper

**Structure:**
```
cmd/calf/main.go          # Entry point
internal/                  # tui/, isolation/, agent/, env/, github/, config/
scripts/                   # Shell scripts (calf-bootstrap, vm-setup, vm-auth)
```

**Commands:** `go build -o calf ./cmd/calf` | `go test ./...` | `staticcheck ./...`

---

## Core Rules

### CALF_VM Auto-Approve

The agent **MUST** verify VM status at session start by running `echo $CALF_VM` before any approval-gated step:
- `CALF_VM=true` → Display "Running in calf-dev VM (isolated environment)" → auto-approve enabled
- Any other value (empty, unset, `false`, etc.) → Display "Running on HOST machine (not isolated)" → require all approvals
- **Fail-safe:** If the check cannot be performed or returns unexpected output, default to HOST (require approval)
- **Never assume VM status** — always verify explicitly

When `CALF_VM=true` (confirmed via explicit check):
- All operations proceed without user confirmation
- EXCEPTION: Destructive remote git operations always require approval:
  - `push --force` (overwrites remote history)
  - `push --delete` / deleting remote branches
- Local-only operations (reset, checkout, clean, etc.) are allowed — GitHub is the restore point
- This applies to ALL workflows (Interactive, Bug Cleanup, Documentation, Implement, Review, Integrate, etc.)

When `CALF_VM` is not true (running on HOST):
- Standard workflow approvals apply as documented
- **When in doubt, require approval**

### Don't Jump Ahead - Always Ask Permission

**The agent must NEVER proactively correct, undo, or fix things without explicit user permission.**

When the agent realizes:
- A mistake was made
- Something needs to be corrected or undone
- An issue should be fixed
- A different approach should be taken

**STOP and ask the user first:**
- Explain what you noticed
- Describe what you think should be done
- **Wait for explicit permission** before taking action

**Examples of jumping ahead (DON'T do this):**
- ❌ "I made a mistake, let me undo that change..."
- ❌ "I should revert that commit, running git reset..."
- ❌ "That file needs updating, let me fix it..."

**Correct behavior (DO this):**
- ✅ "I notice I made a mistake in X. Should I revert that change?"
- ✅ "That file might need updating. Would you like me to update it?"
- ✅ "I could fix Y. Do you want me to proceed?"

**Exception:** If the user explicitly instructs you to fix issues autonomously (e.g., "fix any problems you find"), then you may proceed without asking each time.

### Workflow Modes

**Interactive** is the default workflow unless user specifies otherwise or changes are docs-only.

Routing rules:
- Number `1`-`8` → Launch that workflow directly (see [Quick Reference](docs/WORKFLOWS.md#quick-reference))
- "bug cleanup" → 3. Bug Cleanup workflow
- "refine" / "refinement" → 4. Refine workflow
- "implement" → 5. Implement workflow
- "review" → 6. Review workflow
- "test" → 7. Test workflow
- "integrate" → 8. Integrate workflow
- Documentation-only changes → 2. Documentation workflow

**If unclear, ask user explicitly which workflow to use.**

See [WORKFLOWS.md](docs/WORKFLOWS.md) for complete index, quick reference table, shared conventions, and detailed procedures.

### TODOs
- **`PLAN.md` and phase TODO files are the single source of truth** for all TODOs
- Phase overview in `PLAN.md`, detailed TODOs in `docs/PLAN-PHASE-XX-TODO.md`
- **Completed items must be moved from TODO to DONE files** (e.g., `PLAN-PHASE-XX-TODO.md` → `PLAN-PHASE-XX-DONE.md`)
  - On integrate: move with feature name and date
  - On closure: move with closure reason
  - On direct implementation (Interactive workflow): move with completion date
- Phase complete only when ALL items moved from TODO to DONE
- Code TODOs must also be tracked in phase TODO files

### ADRs and PRDs
**Never modify `docs/adr/*` or `docs/prd/*`** - ADRs and PRDs are immutable historical records.
Create new ADR/PRD to supersede if needed.

### coops-tdd Skills — Mandatory for All Code Changes

**A coops-tdd skill MUST be invoked before writing any code — no exceptions.**

Two variants exist — choose based on whether a human is in the loop:

| Skill | When to use |
|-------|-------------|
| `coops-tdd` | Human in the loop (Interactive, Bug Cleanup workflows) |
| `coops-tdd-auto` | Automated agents from a task list (Implement, Review workflows) |

This applies to:
- New functions, methods, classes, or files
- Bug fixes
- Feature implementation
- Refactoring that changes behaviour
- Any other change to code or scripts

**Skip only for:** explaining concepts, reading/summarising code, documentation edits, config-only changes, or setup questions.

The skill structures all implementation via test-first development: write a failing test → write minimum code to pass → refactor.

**Violation of this rule is treated the same as skipping code review — the change must not be committed.**

### Coding Standards
**All code must meet mandatory quality standards.** Common errors to avoid:
- **Code duplication** - Never leave copy-paste artifacts
- **Missing dependency checks** - Always verify external tools before use
- **Documentation mismatches** - Code must match what docs claim
- **Silent error suppression** - Never hide errors with `&>/dev/null`
- **Missing validation** - Check preconditions before operations
- **Dangerous constructs** - Avoid `eval` and injection risks
- **Go: Custom implementations** - Use stdlib over custom helpers (e.g., `strings.Contains`)
- **Go: Missing GoDoc** - All exported identifiers must have documentation

**Must test all scenarios:** valid inputs, invalid inputs, missing dependencies, auth failures, existing state, network failures.

See [`CODING-STANDARDS/`](CODING-STANDARDS/CODING-STANDARDS.md) for complete requirements and patterns (shared + language-specific files).

---

## Prohibitions

**Never:**
- Run commands without user approval (Interactive workflow — unless `CALF_VM=true`, see [CALF_VM Auto-Approve](#calf_vm-auto-approve))
- Commit without user approval (Interactive workflow — unless `CALF_VM=true`)
- Commit to main branch (Implement, Review workflows — all changes go via worktree branches)
- Perform destructive remote git operations without approval (`push --force`, `push --delete` — even when `CALF_VM=true`)
- Commit with failing tests or build
- Write or modify code without first invoking the appropriate coops-tdd skill (`coops-tdd` for interactive workflows, `coops-tdd-auto` for automated workflows)
- Skip code review for code/script changes (Interactive workflow)
- Modify ADR or PRD files
- Mark TODOs as `[x]` in TODO files - always move completed items to DONE files
- Mark phase complete with items remaining in TODO file
- Work around a hook-blocked operation — if a tool call is blocked by a hook, stop and ask the user to perform the operation manually

---

## Quick Workflow Selection

**When user enters a single `.` as their prompt:**

1. Read `docs/WORKFLOWS.md`
2. Present numbered workflow list (see [WORKFLOWS.md Quick Reference](docs/WORKFLOWS.md#quick-reference))
3. Wait for user to select a number
4. Run the chosen workflow following its standard procedure

**When user enters a number (1-8) as their prompt:**

Skip the menu and launch that workflow directly (same as selecting it from the `.` menu).

---

## Session Start

1. **Check auto-memory** - Read `MEMORY.md` from the auto-memory directory. If the cached active phase matches PLAN.md's current phase (verified in step 7), you can skip reading the full phase TODO file and rely on the cache for context. Always read PLAN.md itself to detect phase changes.
2. **Determine workflow** - If user entered a number (1-8), use that workflow directly. Otherwise, if unclear, ask user (see [Workflow Modes](#workflow-modes) routing rules)
3. **Read workflow and confirm** - Read the workflow file, then confirm briefly: `"Read [Workflow Name] workflow ([N]-step). Proceeding with session start."` — do NOT summarize or reiterate the full steps
4. **Read required docs only** - Load documents based on the workflow:

   | Always read | Workflows 1, 3, 5, 6 only |
   |-------------|-------------------------------|
   | `docs/WORKFLOWS.md` | `CODING-STANDARDS/CODING-STANDARDS.md` + relevant language file |
   | `PLAN.md` | |
   | Active phase TODO file | |
   | `STATUS.md` (for workflows 4-8) | |

   Skip `CODING-STANDARDS/` for workflows 2 (Documentation), 4 (Refine), 7 (Test), 8 (Integrate) — they don't produce code. Use the `coding-standards` skill to load the right files.

5. **Check environment** - Run `echo $CALF_VM` (must happen before any approval-gated step):
   - `CALF_VM=true`: Display "Running in calf-dev VM (isolated environment)" — approvals auto-granted
   - Any other value (empty, unset, etc.): Display "Running on HOST machine (not isolated)" — approvals required
   - If check fails: default to HOST (require approval)
6. Run `git status` to see current branch, then **switch to main if not already there** with `git checkout main && git pull` (ask approval on HOST; auto-approved when `CALF_VM=true`)
7. Run `git fetch` to get latest remote state
8. **Read required docs** from the table in step 4 (always read from main branch). If auto-memory cache is current (phase matches), skip reading the full phase TODO file.
9. Report status and suggest next steps using [Numbered Choice Presentation](docs/WORKFLOWS.md#numbered-choice-presentation)

**Why main branch?** STATUS.md and PLAN.md are only updated on main (per [Documentation Updates on Main](docs/WORKFLOWS.md#documentation-updates-on-main)). Reading from feature branches may show stale data.

**Note:** Only read the active phase TODO file. Do not read future phase files until the current phase is complete.

---

## Session End

At the end of each session, update `MEMORY.md` in the auto-memory directory with:
- Current active phase number and goal
- Open TODO count and key items worked on
- Pipeline queue snapshot (from STATUS.md if updated — Needs Review / Needs Testing / Needs Integration counts)
- Last session date and workflow used
- Any key patterns or conventions discovered

This enables the next session to skip redundant file reads when the phase hasn't changed.

---

## Documentation

**Planning:** [PLAN.md](PLAN.md) **(source of truth)** | [STATUS.md](STATUS.md) | Phase TODO/DONE files in `docs/PLAN-PHASE-XX-{TODO,DONE}.md` (phases 00-05)

**Important:** Read only the active phase TODO file per session.

**Operational:** [ADR-002](docs/adr/ADR-002-tart-vm-operational-guide.md) | [ADR-003](docs/adr/ADR-003-package-download-caching.md) | [bootstrap.md](docs/bootstrap.md)

**Reference:** [WORKFLOWS.md](docs/WORKFLOWS.md) (index + shared conventions) | [WORKFLOW-*.md](docs/) | [architecture.md](docs/architecture.md) | [cli.md](docs/cli.md) | [SPEC.md](docs/SPEC.md) | [CODING-STANDARDS/](CODING-STANDARDS/CODING-STANDARDS.md)

**Historical (immutable):** [ADR-001](docs/adr/ADR-001-cal-isolation.md) | [ADR-002](docs/adr/ADR-002-tart-vm-operational-guide.md) | [PRD-001](docs/prd/prd-001-tart-vm-gui-access.md)
