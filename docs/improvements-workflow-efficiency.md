# Workflow Efficiency Recommendations

> Recommendations for reducing token cost and improving speed of agent sessions

**Created:** 2026-02-11
**Status:** Recommendations (not yet implemented)

---

## Recommendations Summary

| # | Recommendation | Priority | Token Impact | Effort |
|---|---------------|----------|-------------|--------|
| 1 | [Conditional doc loading](#1-conditional-document-loading) | High | ~40% reduction | Low |
| 2 | [Session start ceremony reduction](#2-session-start-ceremony-reduction) | High | ~15% reduction | Low |
| 3 | [Auto-memory for session context](#3-auto-memory-for-session-context) | Medium | ~10% reduction | Low |
| 4 | [WORKFLOWS.md progressive disclosure](#4-workflowsmd-progressive-disclosure) | Medium | ~8% reduction | Medium |
| 5 | [Phase TODO restructuring](#5-phase-todo-restructuring) | Medium | ~10% reduction | Low |
| 6 | [Workflow file deduplication](#6-workflow-file-deduplication) | Low | ~5% reduction | Medium |
| 7 | [Workflows as Claude Code skills](#7-workflows-as-claude-code-skills) | Low | Structural | High |
| 8 | [Git worktrees for PR workflows](#8-git-worktrees-for-pr-workflows) | Low | Speed gain | Medium |
| 9 | [TDD skill (rejected)](#9-tdd-skill-rejected) | N/A | N/A | N/A |

**Impact estimates are relative to session start overhead, not total session cost.**

---

## Current State Analysis

### Session Start Cost

Every session begins by reading these files before any productive work starts:

| File | Lines | Notes |
|------|-------|-------|
| CLAUDE.md | 179 | Loaded as system prompt (always) |
| WORKFLOWS.md | 505 | Read every session per Session Start Procedure |
| WORKFLOW-*.md | 177-411 | Specific workflow file (varies) |
| PLAN.md | 142 | Phase overview |
| PLAN-PHASE-01-TODO.md | 689 | Active phase tasks |
| STATUS.md | 69 | PR/TODO queue status |
| **Total** | **~1,761-1,995** | **Before any work begins** |

Additional files read for code-producing workflows:

| File | Lines | Notes |
|------|-------|-------|
| CODING_STANDARDS.md | 274 | Only needed for workflows 1, 3, 5-7 |

**Estimated token cost of session start: ~4,000-5,000 tokens** (at ~2.5 lines per token for markdown).

### Documentation Scale

- 57 files in `docs/` directory
- 9 workflow files totaling 2,691 lines
- ~658 lines of duplicated content across workflow files (~25%)
- 4 completed items still present in PLAN-PHASE-01-TODO.md (adding ~200 lines of noise)

### Key Observations

1. **WORKFLOWS.md is read every session** but only ~150 of its 505 lines (Quick Reference table + Shared Conventions headers) are needed during session start. The remaining ~350 lines (workflow summaries, selection guide, PR cycle diagram) are reference material.

2. **Phase TODO files contain completed items** that should have been moved to DONE files. PLAN-PHASE-01-TODO.md has 4 completed sections consuming space.

3. **CODING_STANDARDS.md is read for all workflows** but is only relevant to code-producing ones (not Documentation, Refine, Test PR, or Merge PR).

4. **Each workflow file repeats** session start references, checklists, test/build commands, and related documentation sections.

---

## Session Optimization

### 1. Conditional Document Loading

**Problem:** Every session reads CODING_STANDARDS.md (274 lines) and the full workflow file regardless of whether the workflow needs them.

**Recommendation:** Update CLAUDE.md Session Start to specify which files each workflow requires:

| Workflow | Skip CODING_STANDARDS | Skip Workflow File |
|----------|----------------------|-------------------|
| 2 - Documentation | Yes | No |
| 4 - Refine | Yes | No |
| 8 - Test PR | Yes | No |
| 9 - Merge PR | Yes | No |

**Implementation:** Add a "Required Reading" table to CLAUDE.md Session Start that maps workflow numbers to required files. The agent reads only what's listed for the chosen workflow.

**Savings:** ~274 lines (~110 tokens) for 4 of 9 workflows.

### 2. Session Start Ceremony Reduction

**Problem:** The Session Start Procedure requires the agent to read the workflow file and then "reiterate to user" and "confirm understanding" before doing anything. This adds a full conversational round-trip plus output tokens for the summary.

**Recommendation:**
- Remove the "reiterate to user" and "confirm understanding" steps from Session Start Procedure
- Replace with a one-line status message: `"Starting [Workflow Name] workflow on [branch]"`
- The agent still reads the workflow file internally — it just doesn't narrate it back

**Why this is safe:** The workflow file serves as the agent's instructions. Summarizing it back to the user doesn't improve execution quality — the user already knows which workflow they selected.

**Savings:** ~200-400 output tokens per session, plus one fewer round-trip.

### 3. Auto-Memory for Session Context

**Problem:** Every session re-reads PLAN.md and the phase TODO file to discover what's already known: current phase, active tasks, recent completions. This is the same information session after session until a phase transition occurs.

**Recommendation:** Use Claude Code's auto-memory (`MEMORY.md`) to cache:
- Current active phase number and goal
- Active TODO count and key items
- Recent PR queue state (from STATUS.md)
- Last session's workflow and outcome

**How it works:** At session end, write a 10-20 line summary to MEMORY.md. At session start, read MEMORY.md first. If the cached phase matches PLAN.md's current phase, skip reading the full TODO file and just read STATUS.md for queue changes.

**Savings:** ~689 lines (~275 tokens) when phase hasn't changed (most sessions).

**Trade-off:** Stale cache risk if another agent/user makes changes between sessions. Mitigated by always reading PLAN.md (142 lines) to detect phase changes, and falling back to full reads when cache is stale.

---

## Documentation Structure

### 4. WORKFLOWS.md Progressive Disclosure

**Problem:** WORKFLOWS.md is 505 lines and read every session, but most content is reference material not needed during session start.

**Current structure (all read every session):**
- Quick Reference table (22 lines) — needed
- Default Workflow routing (12 lines) — in CLAUDE.md already
- Workflow Summaries (150 lines) — rarely needed (agent reads the specific workflow file)
- PR Workflow Cycle diagram (30 lines) — rarely needed
- STATUS.md Sections table (12 lines) — useful for PR workflows only
- Workflow Selection Guide (30 lines) — in CLAUDE.md already
- Shared Conventions (185 lines) — needed
- Related Documentation (20 lines) — rarely needed

**Recommendation:** Split WORKFLOWS.md into two files:

1. **WORKFLOWS.md** (slim, ~220 lines) — Quick Reference + Shared Conventions only
2. **WORKFLOW-REFERENCE.md** (~285 lines) — Summaries, diagrams, selection guide, STATUS sections

Only WORKFLOWS.md gets read every session. WORKFLOW-REFERENCE.md is consulted when the agent needs to help the user choose a workflow (the `.` command) or understand the PR pipeline.

**Savings:** ~285 lines (~114 tokens) per session.

### 5. Phase TODO Restructuring

**Problem:** PLAN-PHASE-01-TODO.md is 689 lines but contains 4 completed sections (~200 lines) that should be in the DONE file. The TODO file should only contain open work.

**Recommendation:**
1. **Immediate:** Move the 4 completed items from PLAN-PHASE-01-TODO.md to PLAN-PHASE-01-DONE.md
2. **Ongoing:** Enforce the existing TODO-to-DONE movement convention more strictly — completed items should be moved at merge time, not left with "COMPLETED" markers

**Savings:** ~200 lines (~80 tokens) per session, immediately.

### 6. Workflow File Deduplication

**Problem:** ~658 lines of duplicated content across the 9 workflow files (~25% of total). Major sources:

| Duplication Pattern | Lines | Files |
|-------------------|-------|-------|
| Pre-commit/review checklists | ~110 | 6 |
| Test/build command sequences | ~120 | 5 |
| Read-from-queue patterns | ~95 | 7 |
| Key Principles content | ~80 | 8 |
| Related Documentation footers | ~72 | 9 |
| Session Start Procedure sections | ~54 | 9 |
| Command Execution Policy | ~22 | 2 |
| Other (prohibitions, cross-refs) | ~105 | Various |

**Recommendation:** Extract common patterns into Shared Conventions in WORKFLOWS.md, then reference them from workflow files:

1. **Shared Pre-Commit Checklist** — one checklist in WORKFLOWS.md with per-workflow additions
2. **Shared Test/Build Step** — single description referenced by step number
3. **Shared Queue Reading Pattern** — template for "read STATUS.md section X" steps
4. **Remove Related Documentation footers** — the agent doesn't need these; CLAUDE.md already lists all docs

**Savings:** ~200-300 lines across all workflow files. Per-session savings depend on which workflow is active (~20-40 lines per workflow file).

**Trade-off:** Workflow files become less self-contained. An agent must read both WORKFLOWS.md and the specific workflow file (which it already does).

---

## Architecture

### 7. Workflows as Claude Code Skills

**Problem:** Workflows are currently plain markdown documents. The agent reads them as text and follows instructions. There's no programmatic dispatch, parameter passing, or context isolation.

**What Claude Code skills offer:**
- Invoked via `/skill-name` (already how users invoke workflows via numbers)
- Can include structured prompts with parameters
- Execute within the conversation context
- Can be defined in `.claude/skills/` directory

**Recommendation:** Convert each workflow into a Claude Code skill that:
1. Encodes the workflow steps as a structured prompt
2. Loads only the documents that workflow needs (solves conditional loading)
3. Provides workflow-specific context without reading the full WORKFLOWS.md

**Example structure:**
```
.claude/skills/
  workflow-interactive.md
  workflow-documentation.md
  workflow-create-pr.md
  ...
```

**Why "Low" priority:** This is a structural change that requires rethinking how workflows are defined and maintained. The current system works — skills would make it more efficient but require significant migration effort. Best attempted after the simpler optimizations (1-5) are in place.

**Prerequisite:** Understand Claude Code skill limitations — can skills call other skills? Can they read files? How do they interact with CLAUDE.md instructions?

### 8. Git Worktrees for PR Workflows

**Problem:** PR workflows (5-9) require switching branches, which means:
- Stashing or committing uncommitted work
- Running `git checkout` + `git pull` (takes time)
- Reading STATUS.md/PLAN.md from main after switching (they may be stale on feature branches)
- Switching back after the workflow completes

**What git worktrees offer:**
- Multiple branches checked out simultaneously in separate directories
- No branch switching needed — just `cd` to the worktree
- Main branch always available for reading STATUS.md/PLAN.md

**Recommendation:** Use worktrees for PR workflows:

```bash
# One-time setup
git worktree add ../calf-pr-work create-pr/feature-name

# PR workflow operates in ../calf-pr-work/
# Main branch files always readable from ./
```

**Benefits:**
- Eliminates branch switching overhead
- STATUS.md/PLAN.md always readable from main worktree
- Multiple PRs can be in progress simultaneously

**Why "Low" priority:** The current branch-switching approach works and the overhead is small (a few seconds per switch). Worktrees add complexity (two directories to manage, potential confusion about which directory to operate in). Most valuable if running multiple PR workflows in rapid succession.

**Trade-off:** Worktrees consume additional disk space and add cognitive overhead. The `CALF_VM=true` environment already auto-approves the `git checkout` commands, so the main cost is time, not interaction overhead.

### 9. TDD Skill (Rejected)

**Investigated:** Whether a dedicated TDD skill would improve test-first development in Create PR and Bug Cleanup workflows.

**Finding:** TDD is already well-integrated into the existing workflows:
- Create PR workflow (Step 3) explicitly requires TDD: "Write tests first, then implement"
- Bug Cleanup workflow requires reproduction tests before fixes
- [CODING_STANDARDS.md](../CODING_STANDARDS.md) specifies test coverage requirements

**Why rejected:**
- TDD is a development practice, not a discrete operation that benefits from skill isolation
- The test-write-test cycle is inherently interleaved with implementation — extracting it into a separate skill would fragment the workflow
- The existing workflow instructions adequately enforce TDD discipline
- A skill can't easily maintain the iterative red-green-refactor context across multiple tool calls

**Recommendation:** No action needed. Keep TDD as an inline requirement within existing workflows.

---

## Implementation Roadmap

### Phase A: Quick Wins — DONE (2026-02-11)

1. **Move completed items from TODO to DONE** (Recommendation 5) — done
   - Removed 4 completed sections (109 lines) from PLAN-PHASE-01-TODO.md
   - Items were already documented in PLAN-PHASE-01-DONE.md

2. **Replace session start narration with brief confirmation** (Recommendation 2) — done
   - Updated Session Start Procedure in WORKFLOWS.md and CLAUDE.md
   - Replaced "reiterate to user" + "confirm understanding" with one-line confirmation
   - Agent still reads the workflow file; just doesn't narrate it back

3. **Add conditional doc loading table** (Recommendation 1) — done
   - Added Required Reading table to CLAUDE.md Session Start (step 3)
   - Workflows 2, 4, 8, 9 skip CODING_STANDARDS.md

### Phase B: Documentation Restructuring — DONE (2026-02-11)

4. **Split WORKFLOWS.md** (Recommendation 4) — done
   - Moved Workflow Summaries, PR Workflow Cycle, STATUS.md Sections, and Workflow Selection Guide to new WORKFLOW-REFERENCE.md
   - WORKFLOWS.md reduced from 503 to ~350 lines (Quick Reference + Shared Conventions)
   - Updated Quick Reference table links to point directly to workflow files
   - Added WORKFLOW-REFERENCE.md to Related Documentation

5. **Deduplicate workflow files** (Recommendation 6) — done
   - Removed Related Documentation footers from all 9 workflow files + WORKFLOW-TODOS.md (~65 lines total)
   - Extracted Command Execution Policy from WORKFLOW-INTERACTIVE.md and WORKFLOW-BUG-CLEANUP.md into Shared Conventions (~22 lines deduplicated)
   - Both files now reference the shared convention instead of duplicating the policy

6. **Set up auto-memory caching** (Recommendation 3) — done
   - Created MEMORY.md template in auto-memory directory with session context cache
   - Added Session Start step 1: check auto-memory, skip TODO file read if cache is current
   - Added Session End section to CLAUDE.md/AGENTS.md for memory updates
   - Stale-cache detection: always read PLAN.md to verify phase hasn't changed

### Phase C: Architectural Changes (High effort, evaluate after A+B)

7. **Evaluate Claude Code skills** (Recommendation 7)
   - Prototype one workflow as a skill
   - Measure token savings vs. current approach
   - Decide whether to migrate remaining workflows

8. **Evaluate git worktrees** (Recommendation 8)
   - Test worktree setup in calf-dev VM
   - Measure time savings for PR workflow cycles
   - Document worktree conventions if adopted

---

## Related Documentation

- [WORKFLOWS.md](WORKFLOWS.md) — Workflow index and shared conventions
- [CLAUDE.md](../CLAUDE.md) — Agent instructions and session start procedure
- [CODING_STANDARDS.md](../CODING_STANDARDS.md) — Code quality standards
- [PLAN.md](../PLAN.md) — Project plan and phase status
