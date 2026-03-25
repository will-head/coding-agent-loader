---
name: update-docs
description: This skill should be used when the user asks to "update documentation", "update docs", "update affected docs", "move completed TODO items to DONE", "update PLAN.md", or when a workflow step requires documentation and TODO tracking updates after code changes are complete (Interactive step 8, Bug Cleanup step 9, Integrate step 9).
---

# Update Documentation

Apply after code changes to update stale documentation and track TODO progress. Covers Interactive, Bug Cleanup, Integrate, Review, and Implement workflows.

## Step 1: Update Affected Documentation

Review files edited this session (run `git diff` or check session history) and update every doc that is now stale.

**Consider updating:**
- `README.md` — if user-facing behaviour changed
- `AGENTS.md` (via `CLAUDE.md`) — if workflows, rules, or agent behaviour changed
- `docs/SPEC.md` — if technical spec changed
- `docs/architecture.md` — if architecture changed
- `docs/cli.md` — if CLI commands or flags changed
- `docs/bootstrap.md` — if setup or install steps changed
- `docs/plugins.md` — if plugin integration changed
- `docs/roadmap.md` — if roadmap changed
- Comments in changed source files — only if they clarify intent; do not use source comments to track TODOs

**Bug Cleanup workflow only — also update:**
1. `docs/bugs/BUG-NNN-slug.md` — set Status to "Resolved", add resolution details and date
2. `docs/BUGS.md` — remove the row from the active bugs table
3. `docs/bugs/README.md` — change status to "Resolved", add resolved date

**Never modify `docs/adr/*` or `docs/prd/*`** — ADRs and PRDs are immutable historical records. Create a new ADR/PRD to supersede if needed.

## Step 2: Move Completed TODOs to DONE

Skip this step if no TODO items were completed this session.

**Never mark `[x]` in the TODO file — always physically move the item.**

1. Identify which TODO items in the active phase TODO file (`docs/PLAN-PHASE-XX-TODO.md`) were completed this session. Cross-reference the bug report with the phase TODO file if working on a bug fix.
2. Cut each completed item from the TODO file.
3. Paste into the corresponding DONE file (`docs/PLAN-PHASE-XX-DONE.md`) with completion context, using today's date:

   | Workflow | Format |
   |----------|--------|
   | Interactive (direct commit) | `- [x] Item text (completed <today's date>)` |
   | Integrate (merged PR) | `- [x] Item text (feature-branch-name, merged <today's date>)` |
   | Closure (abandoned) | `- [x] Item text (closed - reason)` |

4. Confirm the TODO file no longer contains those items.
5. Phase is only complete when the TODO file is empty — check this before reporting phase status in Step 4.
6. **Clear STATUS.md Refined entries:** If any completed TODOs were listed in the Refined section of `STATUS.md`, remove those rows now. This applies especially to the Interactive workflow, where items bypass the async pipeline and are never automatically removed from Refined.

## Step 3: Add New TODOs

Skip this step if no new TODOs were discovered.

Add any work items identified during this session to the active phase TODO file. TODOs belong in the phase TODO file, not as inline source code comments.

- New work items found during implementation
- Follow-up issues noted but not addressed

Format: `- [ ] Description of work needed`

If a TODO clearly belongs to a future phase rather than the current one, add it to that phase's TODO file instead.

## Step 4: Update PLAN.md

**CRITICAL: Switch to `main` branch before editing PLAN.md or STATUS.md.** If currently in a worktree, commit the feature branch first, then exit the worktree (`ExitWorktree`) and switch to main. Pipeline section moves (Needs Review, Needs Testing, Integrating, etc.) are the responsibility of the calling workflow — do not move items between pipeline sections here. The Refined cleanup in Step 2.6 is the only STATUS.md update this skill performs directly.

Update `PLAN.md` to reflect current project status:
- Phase progress (items remaining after TODO → DONE movement)
- Phase status change if the TODO file is now empty
- Any new work added this session

## Quick Reference

| Workflow | Docs (Step 1) | Bug lifecycle | PLAN.md/TODOs |
|----------|---------------|---------------|---------------|
| Interactive | Affected docs | — | Always |
| Bug Cleanup | Affected docs | BUG file, BUGS.md, bugs/README.md | If bug was a tracked TODO |
| Integrate | — | — | Always (move on merge) |
| Implement / Review | Affected docs | — | Always |
