# Workflow Redesign Plan

> Standalone plan for completing the worktree-based workflow redesign.
> Can be picked up in a fresh session without prior context.

---

## What Was Done

Redesigned the workflow system from a PR-based pipeline to a worktree-based async pipeline.

**Key design decisions:**
- No GitHub PRs — git history on main is sufficient
- Implement and Review agents run on the same machine, sharing the filesystem
- Worktrees at `.claude/worktrees/<name>` are the shared artifact between agents
- Agents stay busy by looping to next queued item rather than waiting
- User participates only in Test workflow (7), run when they have time
- `EnterWorktree`/`ExitWorktree` used only by Implement agent for new work
- Review, Test, Integrate agents access worktrees via file paths + `git -C`

**9 workflows → 8 workflows:**

| # | Old | New |
|---|-----|-----|
| 1 | Interactive | Interactive (unchanged) |
| 2 | Documentation | Documentation (unchanged) |
| 3 | Bug Cleanup | Bug Cleanup (unchanged) |
| 4 | Refine | Refine (minor update) |
| 5 | Create PR | **Implement** (new) |
| 6 | Review & Fix PR | **Review** (new) |
| 7 | Update PR | **Test** (new) |
| 8 | Test PR | **Integrate** (new) |
| 9 | Merge PR | _(removed — folded into Integrate)_ |

**Files created:**
- `docs/WORKFLOW-IMPLEMENT.md` — autonomous, 9-step
- `docs/WORKFLOW-REVIEW.md` — autonomous, 8-step
- `docs/WORKFLOW-TEST.md` — user-driven, 7-step
- `docs/WORKFLOW-INTEGRATE.md` — autonomous, 8-step

**Files rewritten:**
- `docs/WORKFLOWS.md` — async pipeline diagram, new STATUS.md sections, worktree conventions

**Files updated (minor):**
- `CLAUDE.md` — routing rules (1-8), workflow list, docs table, TODOs section, session end
- `docs/WORKFLOW-INTERACTIVE.md` — next workflow reference updated
- `docs/WORKFLOW-REFINE.md` — STATUS.md entry format updated, "Create PR" reference removed

**Files deleted:**
- `docs/WORKFLOW-CREATE-PR.md`
- `docs/WORKFLOW-REVIEW-PR.md`
- `docs/WORKFLOW-UPDATE-PR.md`
- `docs/WORKFLOW-TEST-PR.md`
- `docs/WORKFLOW-MERGE-PR.md`
- `docs/PR-WORKFLOW-DIAGRAM.md`
- `docs/WORKFLOW-REFERENCE.md`

---

## New Pipeline Overview

```
Refined
  ↓ Implement agent (autonomous, workflow 5)
Needs Review
  ↓ Review agent (autonomous, workflow 6)
Needs Testing ←──────────────────────────────────┐
  ↓ USER (workflow 7, when ready)                 │
Needs Integration    Needs Rework ────────────────┘
  ↓ Integrate agent (autonomous, workflow 8)
Done
```

**STATUS.md entry format (canonical):**
```
## Refined
- <feature-name> | docs/PLAN-PHASE-XX-TODO.md § X.X | <description> | refined: YYYY-MM-DD

## Implementing / Reviewing  (claim sections — agent moves item here before starting)
- <feature-name> | .claude/worktrees/<name> | <description>

## Needs Review / Needs Testing / Needs Integration
- <feature-name> | .claude/worktrees/<name> | <description>

## Needs Testing (with test scope)
- <feature-name> | .claude/worktrees/<name> | <description> | Tests: <brief scope>

## Needs Rework (with failure type)
- <feature-name> | .claude/worktrees/<name> | <description> | <Review|Build|Test|User>: <feedback>

## Done
- <feature-name> | <description> | merged: YYYY-MM-DD

## Closed
- <feature-name> | <description> | closed: <reason>
```

---

## Outstanding Issues — Fix These

These are clear bugs introduced by the redesign. Fix all before considering the redesign complete.

### Fix 1 — CODING_STANDARDS table contradiction (CLAUDE.md + AGENTS.md)

**Location:** `CLAUDE.md` line ~156, `AGENTS.md` same line

**Problem:** The table header says `CODING_STANDARDS.md` is needed for "Workflows 1, 3, 5, 6, 8 only" but the prose below it says "Skip for workflows 2, 4, 7, **8** (Integrate)". Direct contradiction — Integrate does not produce code.

**Fix:** Change table header column from `Workflows 1, 3, 5, 6, 8 only` to `Workflows 1, 3, 5, 6 only`. Apply to both CLAUDE.md and AGENTS.md.

---

### Fix 2 — `go test -C` flag invalid (WORKFLOW-IMPLEMENT.md)

**Location:** `docs/WORKFLOW-IMPLEMENT.md` Step 6

**Problem:** The text offers `go test -C .claude/worktrees/<name> ./...` as an alternative. The `-C` flag is not valid for `go test` (only for `go build` since Go 1.21).

**Fix:** Remove the `-C` option. Replace with: "Run from within the worktree using the Bash tool with the worktree directory as CWD, or use `cd .claude/worktrees/<name> && go test ./...`."

---

### Fix 3 — Broken anchor in WORKFLOW-INTEGRATE.md

**Location:** `docs/WORKFLOW-INTEGRATE.md` Step 7 (near `git push --delete` instruction)

**Problem:** Links to `WORKFLOWS.md#cal_vm-auto-approve` — missing the `F`. Correct anchor is `#calf_vm-auto-approve`.

**Fix:** Change `#cal_vm-auto-approve` to `#calf_vm-auto-approve`.

Note: the same broken anchor exists pre-existing in CLAUDE.md and AGENTS.md at the Prohibitions section — fix those too while here.

---

### Fix 4 — WORKFLOW-REFINE.md STATUS.md entry format is wrong

**Location:** `docs/WORKFLOW-REFINE.md` Step 4 examples

**Problem:** Still uses the old pipe-table format with 5 columns. Implement agents reading STATUS.md will fail to parse table-format entries as bullet-format entries.

**Fix:** Replace the Step 4 STATUS.md example with the canonical bullet format:
```
- ssh-management | docs/PLAN-PHASE-01-TODO.md § 1.5 | Add SSH connection retry logic | refined: 2026-03-15
```
Remove the table header and the extra "Notes" column. The feature-name field is important — it becomes the worktree name used by the Implement agent.

---

### Fix 5 — WORKFLOWS.md Session Start missing STATUS.md read

**Location:** `docs/WORKFLOWS.md` — Shared Conventions → Session Start Procedure

**Problem:** The shared session start procedure (which all workflow files defer to) does not include reading STATUS.md. CLAUDE.md's session start does include it (for workflows 4-8). Agents following WORKFLOWS.md's procedure would miss the STATUS.md read.

**Fix:** Add to the Session Start Procedure bullet list, after the `git fetch` step:
```
- Read STATUS.md (for workflows 4–8) to see current pipeline queue
```

---

### Fix 6 — "Phase complete" rule contradicts TODO→DONE movement

**Location:** `docs/WORKFLOWS.md` — Shared Conventions → PLAN.md is Source of Truth

**Problem:** States "Phase complete only when ALL checkboxes `[x]`". This directly contradicts the TODO→DONE Movement rules which say completed items must be *moved* to DONE files and must **never** be marked `[x]` in TODO files.

**Fix:** Replace "Phase complete only when ALL checkboxes `[x]`" with "Phase complete only when ALL items moved from TODO file to DONE file (TODO file is empty)".

---

### Fix 7 — AGENTS.md not updated

**Location:** `AGENTS.md` (byte-for-byte copy of CLAUDE.md)

**Problem:** AGENTS.md was not updated when CLAUDE.md was updated during the redesign. Currently out of sync.

**Fix:** Apply all the same changes made to CLAUDE.md to AGENTS.md:
- Routing rules: update 1-9 → 1-8, replace PR workflow names with new names
- CALF_VM workflow list: add Test (7) alongside Interactive, Bug Cleanup, Documentation, Implement, Review, Integrate
- Prohibitions: "Commit to main branch" wording update
- CODING_STANDARDS table (same as Fix 1)
- Broken anchor fix (same as Fix 3 note)
- Session Start number (1-9 → 1-8, twice)
- TODO movement: "On merge: move with PR number and date" → "On integrate: move with feature name and date"
- Session End: "PR queue snapshot" → "Pipeline queue snapshot"

---

## Discussion Items — Decide Before Finalising

These require a judgment call before implementation.

### Discussion A — Should `staticcheck` be a required gate?

`CLAUDE.md` lists `staticcheck ./...` as a project command alongside `go test` and `go build`. It is not currently mentioned in Implement (Step 6-7), Review (Step 6), or Integrate (Step 5) test/build gates.

**Options:**
1. Add `staticcheck ./...` to the test/build steps of workflows 5, 6, 8 — makes it a hard gate
2. Leave it out of workflow gates — it's a developer tool, not a CI gate
3. Add it as an advisory step with a note ("fix any errors; warnings are informational")

---

### Discussion B — Should "keep busy" be in the Integrate checklist?

WORKFLOW-IMPLEMENT.md has "keep busy / check queue for more" as a named Key Principle and checklist item. WORKFLOW-INTEGRATE.md mentions it only in Step 8 prose and not in the Pre-Integration Checklist.

**Options:**
1. Add "check queue for more items" to Integrate's checklist — consistency with Implement
2. Leave as-is — Integrate's Step 1 and Step 8 prose both describe it; checklist is optional

---

### Discussion C — Orphaned worktree check: WORKFLOWS.md or CLAUDE.md?

The orphaned worktree check ("list `.claude/worktrees/`, cross-reference with STATUS.md") is in WORKFLOWS.md's Session Start but not in CLAUDE.md's Session Start. They describe the same procedure from different angles.

**Options:**
1. Add orphaned worktree check to CLAUDE.md's Session Start — both documents complete
2. Leave in WORKFLOWS.md only — agents read WORKFLOWS.md at session start anyway
3. Make CLAUDE.md's session start explicitly defer to WORKFLOWS.md for the procedure details

---

### Discussion D — Command Execution Policy and the `push --delete` exception

WORKFLOWS.md "Command Execution Policy" says "Autonomous workflows (5, 6, 8) do not require approval." This is correct but incomplete — it doesn't mention that `git push --delete` always requires approval even in autonomous workflows.

**Options:**
1. Add a note: "Exception: `git push --delete` always requires approval — see CALF_VM Auto-Approve"
2. Leave as-is — the CALF_VM Auto-Approve section is the authoritative place for this rule

---

## Pre-Existing Issues (Not Introduced by This Redesign)

These existed before the redesign and are out of scope for this plan, but noted for awareness:

- **Broken anchor in CLAUDE.md/AGENTS.md:** `#cal_vm-auto-approve` should be `#calf_vm-auto-approve` (also in Fix 3 above)
- **`AskUserQuestion` tool reference in WORKFLOW-REFINE.md** — refers to a non-existent tool; Refine asks questions through normal conversation
- **Session Start duplication in WORKFLOW-REFINE.md** — re-lists key principles inline rather than deferring cleanly to shared convention
- **WORKFLOW-TODOS.md and WORKFLOW-EFFICIENCY.md** — exist in `docs/` but not referenced anywhere; status unknown
- **AGENTS.md/CLAUDE.md sync** — two identical files maintained manually; risk of divergence on every edit

---

## How to Resume This Work

1. Read this file
2. Fix issues 1-7 in order (Fix 7 last, since it depends on knowing all CLAUDE.md changes)
3. For each Discussion item A-D, ask the user for their preference before implementing
4. Run a final consistency check across all workflow files once fixes are applied
5. Delete or update this file when complete

---

_Created: 2026-03-15 | Conversation: worktree workflow redesign_
