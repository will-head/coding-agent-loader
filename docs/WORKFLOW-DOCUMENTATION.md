# Documentation Workflow (4-Step)

> Simplified Interactive workflow for documentation-only changes on main branch

**Use When:** Making changes **only** to `.md` files or code comments

**Key Principles:**
- **Always on main branch** - direct commits, no PRs
- **User approval required on HOST** - must approve before commit (auto-approved when `CALF_VM=true`; see [CALF_VM Auto-Approve](WORKFLOWS.md#calf_vm-auto-approve))
- **Skip tests, build, and code review** - not needed for docs
- **Simplified Interactive** - like Interactive workflow but only 4 steps

---

## Overview

The Documentation workflow is a simplified version of the Interactive workflow for documentation-only changes. It commits directly to main with user approval on HOST (auto-approved when `CALF_VM=true`) but skips automated testing, build verification, and code review steps since these aren't applicable to markdown files or comments.

**Target:** main branch (direct commits)
**Approvals:** Required on HOST (auto-approved when `CALF_VM=true`)
**Steps:** 4 (simplified)

---

## Session Start Procedure

Follow [Session Start Procedure](WORKFLOWS.md#session-start-procedure) from Shared Conventions, highlighting:
- This is the Documentation workflow (simplified 4-step)
- Key principles: docs-only, main branch, user approval required on HOST (auto-approved when `CALF_VM=true`), skip tests/build/review
- 4 steps: Make Changes → Ask Approval → Update TODOs → Commit and Push
- Clarify what counts as documentation-only vs. not

---

## When to Use

Use Documentation workflow for changes **exclusively** to:
- Markdown files (`.md`)
- Code comments (inline documentation)
- README files
- Documentation in `docs/` folder

**Do NOT use for:**
- Code changes (even with documentation)
- Script changes (even minor)
- Configuration file changes
- Build file changes

---

## Step-by-Step Process

### Step 1: Make Changes

Edit documentation files:
- Update markdown files
- Fix typos and formatting
- Improve clarity and examples
- Add new sections or documentation

**Ensure:**
- Proper markdown formatting
- Internal links work
- Code examples are accurate
- Consistent style with existing docs

### Step 2: Ask Approval

Present changes to user:
- Summarize what was changed
- Explain why changes were made
- List affected files

**Wait for explicit approval** before committing (auto-approved when `CALF_VM=true`).

### Step 3: Update TODOs

**Invoke the `update-docs` skill** to move completed TODOs to DONE and add any new TODOs discovered.

### Step 4: Commit and Push

**Ask user approval** (auto-approved when `CALF_VM=true`), then commit using [Commit Message Format](WORKFLOWS.md#commit-message-format) from Shared Conventions. Push after commit.

**Done!** No tests, build, or code review needed.

---

## Documentation Quality Checklist

Before committing:
- [ ] Spelling and grammar correct
- [ ] Markdown formatting proper
- [ ] Code examples tested and accurate
- [ ] Internal links work
- [ ] External links valid
- [ ] Consistent style with project
- [ ] Clear and concise language
- [ ] Appropriate level of detail
- [ ] No sensitive information exposed

---

## Important Notes

### What Counts as Documentation-Only

**Documentation-only means:**
- ✅ Markdown file changes
- ✅ Comment changes in code
- ✅ README updates
- ✅ Example updates (if only docs)

**NOT documentation-only:**
- ❌ Code changes with updated comments
- ❌ Script changes (even small)
- ❌ Configuration updates
- ❌ Example code that runs

### When in Doubt

If you're unsure whether changes are documentation-only:
- **Use full Interactive workflow** instead
- Better safe than sorry
- Tests and build won't hurt

### PLAN.md and Phase TODO Updates

Even for docs-only changes, invoke the `update-docs` skill to move completed TODOs to DONE and add any new TODOs discovered.

---

## Examples

### Example 1: Fix Typos

**Changes:**
- Fixed typos in README.md
- Updated broken link in docs/architecture.md

**Workflow:**
1. Fix typos in both files
2. Ask user approval
3. Commit: "Fix typos and broken link in documentation"
4. Push

**Skipped:** Tests, build, code review

### Example 2: Add New Doc Section

**Changes:**
- Added "Troubleshooting" section to docs/bootstrap.md

**Workflow:**
1. Write new troubleshooting section
2. Ask user approval
3. Commit: "Add troubleshooting section to bootstrap guide"
4. Push

**Skipped:** Tests, build, code review

### Example 3: Update Command Examples

**Changes:**
- Updated CLI examples in docs/cli.md

**Workflow:**
1. Update command examples
2. Verify examples are accurate
3. Ask user approval
4. Commit: "Update CLI examples to match current syntax"
5. Push

**Skipped:** Tests, build, code review

