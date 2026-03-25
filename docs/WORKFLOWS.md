# CALF Workflows

> Index of all workflows with quick reference

**Purpose:** This document serves as the index for all CALF workflows. Each workflow has detailed documentation in its own file.

---

## Quick Reference

| # | Workflow | Mode | Steps | Target | Use Case |
|---|----------|------|-------|--------|----------|
| 1 | [Interactive](WORKFLOW-INTERACTIVE.md) | User-gated | 10 | main branch | Default for direct code changes |
| 2 | [Documentation](WORKFLOW-DOCUMENTATION.md) | User-gated | 3 | main branch | Docs-only changes |
| 3 | [Bug Cleanup](WORKFLOW-BUG-CLEANUP.md) | User-gated | 11 | main branch | Fix tracked bugs from BUGS.md |
| 4 | [Refine](WORKFLOW-REFINE.md) | User-gated | 6 | main branch | Refine PLAN.md TODOs and bugs |
| 5 | [Implement](WORKFLOW-IMPLEMENT.md) | Autonomous | 9 | worktree branch | Implement refined TODOs |
| 6 | [Review](WORKFLOW-REVIEW.md) | Autonomous | 8 | worktree branch | Automated code review + fixes |
| 7 | [Test](WORKFLOW-TEST.md) | User-driven | 7 | worktree branch | Manual testing gate |
| 8 | [Integrate](WORKFLOW-INTEGRATE.md) | Autonomous | 9 | main branch | Merge tested work to main |

### Number Shortcuts

Users can enter a workflow number (1-8) at the start of a session to skip the menu and launch that workflow directly.

When the user enters `.`, present the numbered workflow list and wait for selection.

---

## Default Workflow

**Interactive** is the default workflow unless:
- User specifies "bug cleanup" → use Bug Cleanup workflow
- User specifies "refine" or "refinement" → use Refine workflow
- User specifies "implement" → use Implement workflow
- User specifies "review" → use Review workflow
- User specifies "test" → use Test workflow
- User specifies "integrate" → use Integrate workflow
- Changes are documentation-only → use Documentation workflow

**If unclear, ask user explicitly which workflow to use.**

---

## Async Pipeline

Workflows 5–8 form an autonomous pipeline. Agents stay busy by moving to the next queued item rather than waiting for the user. The user participates only in the Test workflow (7), which they run when ready.

```
Refined
  ↓ Implement agent (autonomous)
Needs Review
  ↓ Review agent (autonomous: code review + automated tests)
Needs Testing ←──────────────────────────────────┐
  ↓ USER (workflow 7, when ready)                 │
Needs Integration    Needs Rework ────────────────┘
  ↓ Integrate agent (autonomous)     ↓ Implement agent (loop)
Done
```

**Agent behaviour:** On completing an item, always check the queue for another item before exiting. Multiple items can flow through the pipeline concurrently (each in its own worktree).

---

## STATUS.md Sections

| Section | Set by | Next Workflow |
|---------|--------|---------------|
| **Refined** | Refine (4) | Implement (5) |
| **Implementing** | Implement (5) — claim | Implement (5) in progress |
| **Needs Review** | Implement (5) | Review (6) |
| **Reviewing** | Review (6) — claim | Review (6) in progress |
| **Needs Testing** | Review (6) | Test (7) |
| **Needs Integration** | Test (7) | Integrate (8) |
| **Integrating** | Integrate (8) — claim | Integrate (8) in progress |
| **Needs Rework** | Review (6) or Test (7) | Implement (5) |
| **Done / Closed** | Integrate (8) or any workflow | Appended to [STATUS-MERGED.md](../STATUS-MERGED.md) |

### STATUS.md Entry Format

```
## Refined
- <feature-name> | docs/PLAN-PHASE-XX-TODO.md § X.X | <description> | refined: YYYY-MM-DD

## Implementing
- <feature-name> | .claude/worktrees/<name> | <description>

## Needs Review
- <feature-name> | .claude/worktrees/<name> | <description>

## Reviewing
- <feature-name> | .claude/worktrees/<name> | <description>

## Needs Testing
- <feature-name> | .claude/worktrees/<name> | <description> | Tests: <brief test scope>

## Needs Integration
- <feature-name> | .claude/worktrees/<name> | <description>

## Integrating
- <feature-name> | .claude/worktrees/<name> | <description>

## Needs Rework
- <feature-name> | .claude/worktrees/<name> | <description> | <Review|Build|Test|User>: <feedback>
```

Merged and closed entries go to **STATUS-MERGED.md** (not STATUS.md). See [STATUS-MERGED.md](../STATUS-MERGED.md).

---

## Workflow Selection Guide

### Ask Yourself:

1. **Does a TODO or bug need clarification?**
   - → Use **Refine** workflow (4)

2. **Are you making docs-only changes?**
   - → Use **Documentation** workflow (2)

3. **Is there an active bug in BUGS.md to fix?**
   - → Use **Bug Cleanup** workflow (3)

4. **Is there a refined TODO in STATUS.md ready for implementation?**
   - → Use **Implement** workflow (5)

5. **Is there an item in "Needs Review"?**
   - → Use **Review** workflow (6)

6. **Is there an item in "Needs Testing"?**
   - → Use **Test** workflow (7)

7. **Is there an item in "Needs Integration"?**
   - → Use **Integrate** workflow (8)

8. **Are you making direct code changes to main?**
   - → Use **Interactive** workflow (1)

---

## Shared Conventions

These conventions apply across all workflows. Individual workflow files reference this section rather than repeating these patterns.

### Session Start Procedure

Every workflow follows this procedure at session start:

1. **Read the workflow file** - Read the appropriate `docs/WORKFLOW-*.md`
2. **Confirm briefly** - One line: `"Read [Workflow Name] workflow ([N]-step). Proceeding with session start."` — do NOT summarise or reiterate the full steps
3. **Proceed with standard session start:**
   - Run `echo $CALF_VM` to check environment (must happen before any approval-gated step)
   - Run `git status` to see current branch
   - **CRITICAL:** If not on main branch, switch to main with `git checkout main && git pull` before reading STATUS.md or PLAN.md
   - Run `git fetch` to get latest remote state
   - Read STATUS.md (for workflows 4–8) to see current pipeline queue
   - **Check for orphaned worktrees** — list `.claude/worktrees/` if it exists; cross-reference with ALL active STATUS.md sections (Implementing, Reviewing, Needs Review, Needs Testing, Needs Integration, Integrating); any worktree directory with no STATUS.md entry is an orphan — report to user
   - Read PLAN.md for overview and current phase status
   - Read active phase TODO file for current tasks
   - Report status and suggest next steps

**Why main branch first?** STATUS.md and PLAN.md are the source of truth and only updated on main. Reading them from a feature branch may show stale data.

### CALF_VM Auto-Approve

#### VM Verification

The agent **MUST** verify VM status at session start by running `echo $CALF_VM`:
- `CALF_VM=true` → Display "Running in calf-dev VM (isolated environment)" → auto-approve enabled
- Any other value (empty, unset, `false`, etc.) → Display "Running on HOST machine (not isolated)" → require all approvals
- **Fail-safe:** If the check cannot be performed or returns unexpected output, default to HOST (require approval)
- **Never assume VM status** — always verify explicitly

#### Approval Behaviour

When `CALF_VM=true` (confirmed via explicit check), individual workflow approval steps are skipped — operations proceed automatically without user confirmation.

**Exceptions — always require approval regardless of CALF_VM:**
- `git push --force` (overwrites remote history)
- `git push --delete` / deleting remote branches

**When in doubt, require approval.**

This applies to ALL workflows. See [CLAUDE.md § CALF_VM Auto-Approve](../CLAUDE.md#calf_vm-auto-approve) for the authoritative definition.

### Worktree Conventions

These conventions apply to all workflows that use worktrees (5–8).

#### Paths

- Worktrees are created at `.claude/worktrees/<feature-name>`
- Feature name: lowercase, hyphens, derived from TODO (e.g., "SSH management" → `ssh-management`)
- Branch name matches feature name: `implement/<feature-name>`
- If name already taken, append `-2`, `-3`, etc.

#### Git Operations in a Worktree

When working in a worktree from a separate session (Review, Integrate), use the `-C` flag to run git commands in the worktree directory without changing the session's CWD:

```bash
git -C .claude/worktrees/<name> add -A
git -C .claude/worktrees/<name> commit -m "..."
git -C .claude/worktrees/<name> push
```

Or use absolute paths in Read/Edit/Grep/Write tools to access worktree files directly.

#### Orphaned Worktrees

A worktree is orphaned if its STATUS.md entry is missing or inconsistent. At session start:
1. List `.claude/worktrees/` contents
2. Cross-reference with ALL active STATUS.md sections: Implementing, Reviewing, Needs Review, Needs Testing, Needs Integration, Integrating
3. Any directory with no matching STATUS.md entry is an orphan — report to user; do NOT silently delete

#### Claim Mechanism

Before entering or working in a worktree, agents must claim the item in STATUS.md to prevent two agents picking up the same work:
1. Move item from its current section (e.g., "Refined" → "Implementing", "Needs Review" → "Reviewing", "Needs Integration" → "Integrating")
2. Commit and push this STATUS.md change on main **before** starting work
3. If push fails (race condition — another agent claimed first), stop and pick the next unclaimed item

#### Worktree Lifecycle

| Stage | Action | By |
|-------|--------|----|
| New work | `EnterWorktree implement/<name>` | Implement agent |
| In progress | `ExitWorktree keep` | Implement agent |
| Review/Test | File paths into `.claude/worktrees/<name>` | Review, Test agents |
| Claim for merge | Move to "Integrating" in STATUS.md | Integrate agent |
| Merge | `git merge implement/<name>` from main | Integrate agent |
| Cleanup | `git worktree remove .claude/worktrees/<name>` | Integrate agent |

### Numbered Choice Presentation

When presenting items for user selection (TODOs, worktrees, tasks), **always use a numbered list** so users can reply with just a number:

```
Available items:

1. ssh-management | SSH management (1.5)
2. snapshot-create | Snapshot create command (1.4)
3. cache-parity | Go cache parity cleanup (1.7)

Enter number:
```

This applies to:
- TODO selection (Interactive, Refine workflows)
- Queue selection (Implement, Review, Test, Integrate workflows)
- Next step suggestions at workflow completion
- Any time user must choose between multiple items

### Next Workflow Guidance

At the end of workflows 4–8, read STATUS.md and suggest the next workflow based on what's actually queued. Check sections in **priority order** (items further along the pipeline should be completed first):

| Priority | STATUS.md Section | Suggested Workflow |
|----------|-------------------|--------------------|
| 1 (highest) | Needs Integration (has entries) | **8** (Integrate) |
| 2 | Needs Testing (has entries) | **7** (Test) |
| 3 | Needs Rework (has entries) | **5** (Implement) |
| 4 | Needs Review (has entries) | **6** (Review) |
| 5 | Refined (has entries) | **5** (Implement) |
| 6 (lowest) | Nothing queued | **4** (Refine) to prepare more TODOs |

**How to apply:**

1. Read STATUS.md after the workflow completes (already on main branch at this point)
2. Find the highest-priority non-empty section from the table above
3. Display: `Next: run workflow X (Workflow Name) — N items in queue`
4. If multiple sections have entries, mention them: `Also: N items in Needs Review, N refined TODOs ready`

**Example output:**
```
Next: run workflow 7 (Test) — 2 items in Needs Testing
Also: 1 item in Needs Review
```

### Sequential Question and Test Presentation

When gathering information or presenting testing instructions, **present items one by one** rather than as a batch list.

**For multi-part questions:**
1. Present question #1 only
2. Wait for user's answer
3. Ask any follow-up questions needed to fully understand
4. When fully satisfied with #1, present question #2
5. Repeat until all questions answered

**For manual test instructions:**
1. Present test step #1 only
2. Wait for user to confirm pass/fail
3. If failed: user can choose to fix now, add as TODO, or accept as known issue
4. When step #1 is resolved (passed or handled), present step #2
5. Repeat until all tests complete

**Never present a batch list like:**
- ❌ "Run these tests: 1. Test A, 2. Test B, 3. Test C"

**Always present sequentially:**
- ✅ "Test 1: Do A" → wait for result → handle → "Test 2: Do B" → etc.

### Commit Message Format

Use imperative mood with Co-Authored-By. Always use heredoc for multi-line:

```bash
git commit -m "$(cat <<'EOF'
Brief summary (imperative mood)

Detailed description of what changed and why.

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>
EOF
)"
```

### Documentation Updates on Main

**CRITICAL:** PLAN.md and STATUS.md updates must ALWAYS be done on main branch:
- All STATUS.md claim/update steps happen on main before or after worktree work
- Do NOT update STATUS.md from inside a worktree session
- Switch back to main: `git checkout main` (or use ExitWorktree first)
- Commit and push STATUS.md/PLAN.md changes on main

### PLAN.md is Source of Truth

- All TODOs must be tracked in PLAN.md
- Phase complete only when ALL items moved from TODO file to DONE file (TODO file is empty)
- Update PLAN.md in every workflow
- Code TODOs must reference PLAN.md

### TODO → DONE Movement

**Completed items must be moved from TODO files to DONE files:**

**When to move:**
- **On integrate** (Integrate workflow) — most common
  - Cut TODO from `PLAN-PHASE-XX-TODO.md`
  - Paste into `PLAN-PHASE-XX-DONE.md` with feature name and date
  - Example: `- [x] Add SSH retry logic (ssh-management, merged 2026-03-15)`

- **On closure** (any workflow where work is abandoned)
  - Move to DONE file with closure reason
  - Example: `- [x] Add SSH retry logic (closed - filed as known issue)`

- **On direct implementation** (Interactive workflow)
  - Move to DONE file after successful commit
  - Example: `- [x] Add SSH retry logic (completed 2026-03-15)`

**Never mark as `[x]` in TODO file — always move to DONE file when complete.**

### coops-tdd Skill Requirement

**A coops-tdd skill MUST be invoked before writing any code in any workflow — no exceptions.**

Two variants exist — choose based on whether a human is in the loop:

| Skill | Workflows |
|-------|-----------|
| `coops-tdd` | Interactive (1) — Step 1 (Implement); Bug Cleanup (3) — Step 2 (Implement) |
| `coops-tdd-auto` | Implement (5) — Step 5 (Implement TDD); Review (6) — Step 5 (Fix Issues) |

Code changes made without invoking the appropriate skill must not be committed. This rule is enforced in all pre-commit checklists and the Prohibitions section of AGENTS.md.

### Command Execution Policy

**Ask user approval before running ANY command** (auto-approved when `CALF_VM=true`), including:
- Git operations (commit, push, branch, merge)
- Build commands
- Test commands
- Script execution
- Package installs
- Any destructive operations

**Exception:** Read/Grep/Glob tools for code searching do not require approval.

This policy applies to Interactive and Bug Cleanup workflows. Autonomous workflows (5, 6, 8) do not require approval.

**Exception:** `git push --delete` always requires approval in all workflows, including autonomous ones — see [CALF_VM Auto-Approve](#calf_vm-auto-approve).

---

## Related Documentation

**Core Documentation:**
- [CLAUDE.md](../CLAUDE.md) - Agent instructions and core rules

**Workflow Detail Files:**
- [WORKFLOW-INTERACTIVE.md](WORKFLOW-INTERACTIVE.md) - Interactive workflow (10-step)
- [WORKFLOW-DOCUMENTATION.md](WORKFLOW-DOCUMENTATION.md) - Documentation workflow (3-step)
- [WORKFLOW-BUG-CLEANUP.md](WORKFLOW-BUG-CLEANUP.md) - Bug Cleanup workflow (11-step)
- [WORKFLOW-REFINE.md](WORKFLOW-REFINE.md) - Refine workflow (6-step)
- [WORKFLOW-IMPLEMENT.md](WORKFLOW-IMPLEMENT.md) - Implement workflow (9-step, autonomous)
- [WORKFLOW-REVIEW.md](WORKFLOW-REVIEW.md) - Review workflow (8-step, autonomous)
- [WORKFLOW-TEST.md](WORKFLOW-TEST.md) - Test workflow (7-step, user-driven)
- [WORKFLOW-INTEGRATE.md](WORKFLOW-INTEGRATE.md) - Integrate workflow (9-step, autonomous)

**Project Management:**
- [PLAN.md](../PLAN.md) - TODOs and implementation tasks (source of truth)
- [STATUS.md](../STATUS.md) - Project status tracking (pipeline queue)
- [CODING-STANDARDS/](../CODING-STANDARDS/CODING-STANDARDS.md) - Code quality standards (shared + language-specific)
