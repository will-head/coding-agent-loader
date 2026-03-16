# Integrate Workflow (9-Step)

> Autonomous merge of tested worktree branches into main with cleanup

**Use When:** Items in STATUS.md "Needs Integration" section

**Key Principles:**
- **No permission needed** — fully autonomous except `git push --delete` (always requires approval)
- **Claim before starting** — move item to "Integrating" in STATUS.md on main before merging
- **Merge commit always** — use `--no-ff` to preserve branch history; never squash or rebase
- **Post-merge verification** — run tests and build on main after merge before pushing
- **Conflicts go back** — merge conflicts cannot be auto-resolved; send back to Needs Rework
- **Full cleanup** — remove worktree directory and remote branch after successful merge
- **Keep busy** — after completing an item, check for more before exiting

---

## Overview

The Integrate workflow merges tested worktree branches into main, verifies the result, cleans up worktrees and branches, updates documentation, and moves TODOs to DONE. It is the final step in the autonomous pipeline and runs without user input.

**Target:** main branch (integration)
**Approvals:** Not required (autonomous), except `git push --delete` (always requires approval)
**Steps:** 9

---

## Step-by-Step Process

### Step 1: Read Queue

On main, read STATUS.md "Needs Integration" section.

**If queue empty:** Report "Nothing to integrate — Needs Integration queue is empty." Suggest next workflow per [Next Workflow Guidance](WORKFLOWS.md#next-workflow-guidance). Exit.

**If multiple items:** Process all items in sequence. Present them using [Numbered Choice Presentation](WORKFLOWS.md#numbered-choice-presentation) if user input is available; otherwise process first item and continue.

### Step 2: Claim Item

On main, move the item from "Needs Integration" to "Integrating" in STATUS.md. Commit and push immediately:

```bash
git add STATUS.md
git commit -m "Claim <feature-name> for integration"
git push
```

**If push fails** (race condition — another agent claimed first): pull, find next unclaimed item, repeat from Step 1.

### Step 3: Pull Latest Main

Ensure main is current before merging:

```bash
git checkout main
git pull
```

This reduces the likelihood of conflicts introduced by concurrent main commits.

### Step 4: Check for Conflicts (Dry Run)

Before merging, check whether a clean merge is possible:

```bash
git merge --no-commit --no-ff implement/<feature-name>
git merge --abort
```

**If dry run reports conflicts:** Do NOT attempt the real merge. Update STATUS.md: move item from "Integrating" to "Needs Rework" with feedback:

```
## Needs Rework
- <feature-name> | .claude/worktrees/<name> | <description> | Review: merge conflict with main — rebase or resolve conflicts in branch
```

Commit and push STATUS.md on main. Proceed to next item or exit.

**If dry run is clean:** Proceed to Step 5.

### Step 5: Merge Branch

```bash
git merge --no-ff implement/<feature-name> -m "$(cat <<'EOF'
Merge implement/<feature-name>

<Description of what this implements>

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>
EOF
)"
```

`--no-ff` ensures a merge commit is always created, preserving the full branch history.

**Never use `--squash` or `--rebase`.**

### Step 6: Post-Merge Verification

Run tests and build on main after the merge:

```bash
go test ./...
go build -o calf ./cmd/calf
which staticcheck >/dev/null 2>&1 || { echo "staticcheck not installed. Install with: go install honnef.co/go/tools/cmd/staticcheck@latest"; exit 1; }
staticcheck ./...
```

**If staticcheck is not installed:** Stop. Report the error and ask the user to install it. Do not push or clean up — leave main in its post-merge state until resolved.

**If either test, build, or staticcheck fails:** The merge introduced a regression. Revert it:

```bash
git reset --hard HEAD~1
```

Update STATUS.md: move item from "Integrating" to "Needs Rework":

```
## Needs Rework
- <feature-name> | .claude/worktrees/<name> | <description> | Test: post-merge regression — <test failure summary>
```

Commit and push STATUS.md on main. The worktree and branch are preserved for the Implement agent. Proceed to next item or exit.

**If both pass:** Proceed to Step 7.

### Step 7: Push Main

```bash
git push
```

If push fails (another agent pushed to main concurrently): pull, re-run tests, then push again.

### Step 8: Clean Up Worktree and Branch

**Remove local worktree:**

```bash
git worktree remove .claude/worktrees/<name>
```

If worktree has uncommitted changes (should not happen at this stage but may if something went wrong): investigate before forcing removal.

**Delete remote branch (requires approval — destructive remote operation):**

Ask user approval (auto-approved when `CALF_VM=true` does NOT apply here — `git push --delete` always requires explicit approval per [CALF_VM Auto-Approve](WORKFLOWS.md#calf_vm-auto-approve)):

```bash
git push origin --delete implement/<feature-name>
```

**If user declines branch deletion:** Leave the remote branch. Note it in STATUS.md Done entry as "remote branch not deleted". Continue.

**Delete local branch reference:**

```bash
git branch -d implement/<feature-name>
```

### Step 9: Update Documentation

**Update STATUS.md on main:**
- Remove item from "Integrating"
- Add to "Done" with merge date

```
## Done
- <feature-name> | <description> | merged: YYYY-MM-DD
```

**Invoke the `update-docs` skill** to move completed TODOs to DONE, add any new TODOs, and update PLAN.md phase status.

**Commit all documentation changes:**

```bash
git add STATUS.md PLAN.md docs/PLAN-PHASE-*.md
git commit -m "$(cat <<'EOF'
Integrate <feature-name>: update documentation

Move TODO to DONE, update STATUS.md, update PLAN.md.

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>
EOF
)"
git push
```

**Check queue for more items.** If unclaimed items exist in "Needs Integration", loop back to Step 1.

---

## Pre-Integration Checklist

- [ ] Item claimed in STATUS.md ("Integrating") before merging
- [ ] Latest main pulled before merging
- [ ] Dry-run merge confirmed no conflicts
- [ ] Branch merged with `--no-ff` (merge commit)
- [ ] Post-merge tests pass (`go test ./...`)
- [ ] Post-merge build succeeds (`go build -o calf ./cmd/calf`)
- [ ] Post-merge staticcheck passes (`staticcheck ./...`)
- [ ] Main pushed
- [ ] Local worktree removed (`git worktree remove`)
- [ ] Remote branch deleted (with approval) or noted as skipped
- [ ] Local branch reference deleted
- [ ] STATUS.md updated (moved to Done)
- [ ] Completed TODO moved from phase TODO to phase DONE file
- [ ] PLAN.md phase status updated if applicable
- [ ] Documentation changes committed and pushed

---

## Edge Cases

**Dry-run shows conflicts:** Always send back to Needs Rework. Never attempt manual conflict resolution during integration — that is a design decision best made by the Implement agent with full context.

**Post-merge regression:** Revert immediately with `git reset --hard HEAD~1`. Do not push a broken main. The worktree is preserved so the Implement agent can fix the issue.

**Worktree directory missing at cleanup:** If `.claude/worktrees/<name>` no longer exists, skip `git worktree remove` — it may have already been cleaned up. Continue with branch deletion and documentation steps.

**`git worktree remove` fails due to uncommitted changes:** Investigate. If changes are the review fixes (should be committed), something went wrong. Report to user. Do not force-remove.

**`git branch -d` fails (unmerged):** This means the merge did not include all commits on the branch. Investigate before using `-D`. This should not happen after a successful `git merge --no-ff`.

**Multiple items, one conflicts:** Skip the conflicting item (send to Needs Rework), continue with the remaining items.

**Phase complete after integration:** If moving the TODO to DONE results in all items being done, update PLAN.md phase status to "Complete" and note it in the commit message.

**Remote branch deletion declined:** Acceptable. Add a note to the Done entry. Stale remote branches are low risk and can be cleaned up in batch with `git remote prune origin` later.
