# Refine Workflow (6-Step)

> Refine TODOs and bugs to ensure they are implementation-ready with user approvals

**Use When:** Clarifying and detailing TODOs or bugs before implementation begins

**Key Principles:**
- **Defaults to active phase** - refine TODOs in current active phase unless user specifies different phase
- **Also offers active bugs** - presents bugs from `docs/BUGS.md` alongside phase TODOs
- **Approval required on HOST** - user must approve changes before committing to main (auto-approved when `CALF_VM=true`; see [CALF_VM Auto-Approve](WORKFLOWS.md#calf_vm-auto-approve))
- **Target main branch** - updates phase TODO file, bug reports, and STATUS.md directly on main
- **Comprehensive requirements** - gather all details needed for implementation
- **Track refinement** - both prefix TODO in phase file and add to STATUS.md

---

## Overview

The Refine workflow ensures TODOs and bugs are implementation-ready by gathering complete requirements through clarifying questions. Once refined, TODOs are prefixed with "REFINED" in the phase TODO file and tracked in STATUS.md's "Refined" section. Bugs are refined by updating their bug report with additional detail.

**Default Behavior:** Offers TODOs from the **current active phase** and active bugs from `docs/BUGS.md` unless user specifies a different phase or specific item.

**Target:** main branch (direct updates)
**Approvals:** Required on HOST (auto-approved when `CALF_VM=true`)
**Steps:** 6 (thorough refinement with tracking)

---

## Session Start Procedure

Follow [Session Start Procedure](WORKFLOWS.md#session-start-procedure) from Shared Conventions, highlighting:
- This is the Refine workflow (6-step for refining phase TODOs and bugs)
- Key principles: defaults to active phase, also offers active bugs, approval required on HOST (auto-approved when `CALF_VM=true`), main branch, comprehensive requirements, track refinement
- 6 steps: Read PLAN.md & Phase TODO & BUGS.md → Ask Questions → Update Phase TODO/Bug Report → Update STATUS.md → Ask Approval → Commit
- Explain the REFINED prefix and STATUS.md tracking
- Defaults to active phase TODOs and active bugs unless user specifies different phase or item

---

## When to Use

Use Refine workflow when:
- A TODO in a phase TODO file lacks sufficient detail for implementation
- A bug in `docs/BUGS.md` needs more investigation or detail before fixing
- Requirements are unclear or ambiguous
- Multiple implementation approaches are possible
- User input is needed to define acceptance criteria
- Technical decisions require user preferences

**Defaults to active phase TODOs and active bugs** but user can specify any phase to refine.

**Do NOT use for:**
- Simple, self-explanatory TODOs or well-documented bugs
- TODOs with complete requirements already documented
- Implementation work (use Interactive, Bug Cleanup, or Implement workflows instead)

---

## Step-by-Step Process

### Step 1: Read PLAN.md, Phase TODO File, and BUGS.md

**Note:** Session Start Procedure ensures you're on main branch before this step (PLAN.md is only updated on main).

**First, read `PLAN.md`** to determine the current active phase:
- Check "Current Status" section to identify active phase (e.g., "Phase 0 (Bootstrap): Mostly Complete")
- Note the active phase TODO file (e.g., `docs/PLAN-PHASE-00-TODO.md`)
- Verify the phase status

**Determine target phase:**
- **Default:** Use the active phase unless user specifies otherwise
- **User-specified:** If user mentions a specific phase (e.g., "refine Phase 1 TODO"), use that phase instead
- Read the appropriate phase TODO file based on target phase

**Also read `docs/BUGS.md`** to get the list of active bugs.

**Then, identify the item needing refining** from both phase TODOs and active bugs:

**Identify:**
- Which TODO or bug the user wants refined
- Current description and context
- Related TODOs or dependencies within the same phase
- Section location within the phase file (for TODOs) or bug report path (for bugs)

If user hasn't specified which item, present candidates from **both** phase TODOs and active bugs using [Numbered Choice Presentation](WORKFLOWS.md#numbered-choice-presentation) so the user can reply with just a number. Group them with clear headings (e.g., "Phase TODOs:" and "Active Bugs:").

### Step 2: Ask Clarifying Questions

Ask questions **one by one** following the [Sequential Question and Test Presentation](WORKFLOWS.md#sequential-question-and-test-presentation) convention. Wait for the user's full response to each question before asking the next.

Ask the user to clarify:

**Requirements:**
- What is the desired outcome?
- What are the acceptance criteria?
- Are there constraints or limitations?
- What edge cases need handling?

**Implementation Details:**
- What approach should be used?
- Are there preferred tools or libraries?
- Should existing patterns be followed?
- What testing is required?

**User Preferences:**
- UI/UX decisions
- Configuration options
- Error handling behavior
- Performance vs. simplicity trade-offs

**Continue asking until:**
- All ambiguity is resolved
- Implementation path is clear
- Acceptance criteria are defined
- User confirms completeness

### Step 3: Update Phase TODO File or Bug Report

**For TODOs:** Update the TODO in the phase TODO file (e.g., `docs/PLAN-PHASE-00-TODO.md`):

1. **Prefix with "REFINED"** at the start of the TODO line
2. **Expand description** with gathered requirements
3. **Add sub-items** with implementation details if helpful
4. **Include acceptance criteria** clearly stated
5. **Note any constraints** or special considerations

**For Bugs:** Update the bug report file (e.g., `docs/bugs/BUG-NNN-slug.md`):

1. Expand root cause analysis with gathered information
2. Update or add resolution path details
3. Add acceptance criteria for the fix
4. Update workarounds if new ones discovered
5. Note any constraints or dependencies

**Example transformation (in `docs/PLAN-PHASE-00-TODO.md`):**

Before:
```markdown
- [ ] Add option to sync git repos on init
```

After:
```markdown
- [ ] **REFINED:** Add option to sync git repos on init
  - Prompt user during --init to enter repo names (format: owner/repo)
  - Clone using `gh repo clone` to ~/code/github.com/owner/repo
  - Support multiple repos (comma-separated input)
  - Skip if gh auth not configured (show warning)
  - Acceptance criteria: User can specify repos during init and they are cloned successfully
  - Constraints: Must handle gh auth failures gracefully
```

### Step 4: Update STATUS.md

Add entry to `STATUS.md` under the "Refined" section:

**Entry format:**
```
- <feature-name> | docs/PLAN-PHASE-XX-TODO.md § X.X | <description> | refined: YYYY-MM-DD
```

**Include:**
- Feature name: lowercase, hyphens, max ~30 chars (used as worktree name by Implement agent)
- Location in phase TODO file (e.g., `docs/PLAN-PHASE-01-TODO.md § 1.5`)
- Concise description of what was refined
- Date refined (use YYYY-MM-DD format)

### Step 5: Ask Approval

Present changes to user for review:

1. **Show phase TODO file changes** - highlight refined TODO with full details (e.g., in `docs/PLAN-PHASE-00-TODO.md`)
2. **Show STATUS.md entry** - display new tracking entry
3. **Summarize refining** - explain what was clarified
4. **List affected files** - phase TODO file and STATUS.md

**Wait for explicit user approval** before committing (auto-approved when `CALF_VM=true`).

If user requests changes, return to Step 2 or Step 3 as needed.

### Step 6: Commit and Push

After approval (user approval on HOST; auto-approved when `CALF_VM=true`), stage `docs/PLAN-PHASE-XX-TODO.md` and `STATUS.md`, then commit using [Commit Message Format](WORKFLOWS.md#commit-message-format). Include "Refine TODO:" prefix and list key requirements in the body. Push after commit.

**Done!** TODO is now implementation-ready.

**Suggest next workflow** by checking STATUS.md — see [Next Workflow Guidance](WORKFLOWS.md#next-workflow-guidance).

---

## Refine Quality Checklist

Before presenting for approval:
- [ ] All ambiguity removed from TODO
- [ ] Implementation approach is clear
- [ ] Acceptance criteria are defined
- [ ] Edge cases are considered
- [ ] Constraints and limitations documented
- [ ] User preferences captured
- [ ] Correct phase TODO file updated (active phase or user-specified)
- [ ] TODO prefixed with "REFINED" in phase TODO file
- [ ] Entry added to STATUS.md "Refined" section with correct location format
- [ ] Related TODOs considered for dependencies (within same phase)

---

## Important Notes

### Phase Selection

**Default behavior:**
- Refine TODOs in the current active phase
- Most TODOs should be refined in the active phase

**User can specify different phase:**
- User may want to refine future phase TODOs for planning purposes
- If user specifies a phase (e.g., "refine the config file TODO in Phase 1"), use that phase
- Useful for planning ahead or clarifying dependencies

### What Makes a TODO "Refined"

A refined TODO should:
- Be actionable without further clarification
- Have clear acceptance criteria
- Include implementation guidance
- Note any constraints or gotchas
- Specify testing requirements
- Be ready for immediate implementation

### When to Stop Asking Questions

Stop asking questions when:
- User confirms "that's everything"
- Implementation path is unambiguous
- All decision points have answers
- Further details would be over-specification

### Multiple TODOs

If multiple related TODOs need refining:
- Refine one at a time
- Note dependencies between them
- Ensure consistency across refined TODOs
- Can run workflow multiple times

### STATUS.md Tracking

The "Refined" section in STATUS.md:
- Provides quick overview of refined TODOs
- Helps avoid duplicate refine work
- Tracks when refining occurred
- Links refined items to phase TODO file location (e.g., `PLAN-PHASE-00-TODO.md § 0.10`)

---

## Examples

### Example 1: Vague TODO

**Original (in `docs/PLAN-PHASE-01-TODO.md`):**
```markdown
- [ ] Improve error messages
```

**After Refining (in `docs/PLAN-PHASE-01-TODO.md`):**
```markdown
- [ ] **REFINED:** Improve error messages in calf-bootstrap script
  - Add context to all error messages (what failed, why, what to do)
  - Use consistent format: "ERROR: [what failed]. [why]. [action]"
  - Replace generic "Command failed" with specific operation names
  - Add suggestions for common failures (Tart not installed, VM not found, etc.)
  - Acceptance criteria: All error messages follow format and provide actionable guidance
  - Testing: Trigger each error condition and verify message quality
```

**STATUS.md entry:**
```markdown
- improve-error-messages | docs/PLAN-PHASE-01-TODO.md § 1.2 | Standardize error format with context and suggestions | refined: 2026-01-23
```

### Example 2: Implementation Choice

**Original (in `docs/PLAN-PHASE-01-TODO.md`):**
```markdown
- [ ] Add configuration file support
```

**After Refining (in `docs/PLAN-PHASE-01-TODO.md`):**
```markdown
- [ ] **REFINED:** Add configuration file support for calf-bootstrap
  - File location: ~/.config/calf/config.yaml (XDG standard)
  - Format: YAML with sections for vm_defaults, proxy, snapshots
  - Supported options: default_cpu, default_memory, default_disk, proxy_mode, auto_snapshot
  - Fallback to built-in defaults if file missing (no error)
  - Validation: Check types and ranges, show warnings for invalid values
  - Acceptance criteria: Config file overrides defaults, validation works, errors are clear
  - Implementation: Use yq or pure bash parsing (decided: pure bash for zero dependencies)
```

**STATUS.md entry:**
```markdown
- add-config-file-support | docs/PLAN-PHASE-01-TODO.md § 1.2 | YAML config at ~/.config/calf/config.yaml with validation | refined: 2026-01-23
```

### Example 3: Feature with Dependencies

**Original (in `docs/PLAN-PHASE-00-TODO.md`):**
```markdown
- [ ] Auto-sync repos on VM start
```

**After Refining (in `docs/PLAN-PHASE-00-TODO.md`):**
```markdown
- [ ] **REFINED:** Auto-sync repos on VM start in calf-bootstrap
  - Prerequisites: Requires "Add git repo sync on init" TODO to be completed first
  - Behavior: On `--run`, check ~/code for git repos, fetch updates, show status if behind
  - User prompt: If repos are behind, ask "Pull updates? [Y/n]"
  - Default action: Pull all repos if user confirms
  - Configurable: Add AUTO_SYNC=true/false to config file (default: true)
  - Skip conditions: No network, no repos found, AUTO_SYNC=false
  - Acceptance criteria: Updates are detected and user can choose to pull automatically
  - Dependencies: Blocks on git repo sync TODO (must have repos to sync)
```

**STATUS.md entry:**
```markdown
- auto-sync-repos | docs/PLAN-PHASE-00-TODO.md § 0.10 | Fetch and optionally pull repo updates on --run | refined: 2026-01-23
```

