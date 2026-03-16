# Test Workflow (7-Step)

> User-driven manual testing of items ready in the Needs Testing queue

**Use When:** Items in STATUS.md "Needs Testing" section, when user is available to test

**Key Principles:**
- **User drives the pace** — run this workflow when you have time to test; agents handle the rest
- **Sequential presentation** — one test step at a time, never a batch list
- **User approval ALWAYS required** — even when `CALF_VM=true`; agents cannot perform manual tests
- **Failure feedback feeds back** — failed items return to Needs Rework with specific notes for the Implement agent
- **Partial testing is fine** — test what you can; untested items stay in Needs Testing
- **Async-friendly** — STATUS.md is updated after each item so agents can act immediately on results

---

## Overview

The user runs this workflow when they are ready to test. It presents each item in "Needs Testing" and walks through test steps one at a time. Passing items advance to "Needs Integration" for the Integrate agent. Failing items return to "Needs Rework" with specific feedback. The user can test as many or as few items as they wish in one session.

**Target:** Needs Testing → Needs Integration or Needs Rework
**Approvals:** Always required (user performs the testing)
**Steps:** 7

---

## Step-by-Step Process

### Step 1: Read Test Queue

Read STATUS.md "Needs Testing" section.

**If queue empty:** Report "Nothing to test — Needs Testing queue is empty." Suggest next workflow per [Next Workflow Guidance](WORKFLOWS.md#next-workflow-guidance). Exit.

**If one item:** Confirm with user before starting: "Ready to test `<feature-name>` (<description>)?"

**If multiple items:** Present using [Numbered Choice Presentation](WORKFLOWS.md#numbered-choice-presentation). User can select a specific item or say "all" to work through them all.

### Step 2: Build Worktree for Testing

Before presenting test steps, ensure the worktree build is current. Ask user approval (even in CALF_VM — user may want to do this themselves), then:

```bash
go build -C .claude/worktrees/<name> -o calf ./cmd/calf
```

**If build fails:** Do not proceed with testing. Move item to "Needs Rework" immediately:

```
## Needs Rework
- <feature-name> | .claude/worktrees/<name> | <description> | Build: build failed before testing — <error summary>
```

Commit and push on main. Move to next item or exit.

### Step 3: Present Test Instructions (Sequential)

Follow [Sequential Question and Test Presentation](WORKFLOWS.md#sequential-question-and-test-presentation) strictly:

1. Announce the item being tested: "Testing `<feature-name>`: <description>"
2. Present **test step #1 only** with:
   - What to do (specific commands or actions)
   - What to look for (expected outcome)
3. STOP and wait for the user to run the test and report pass or fail
4. Handle the result before moving to the next step

**Do NOT present all test steps at once.**

### Step 4: Handle Each Test Result

After the user reports a result for each step:

**Pass:**
- Confirm: "Step N passed."
- Present the next step (return to Step 3)
- If all steps passed, proceed to Step 5

**Fail:**
- Ask user to describe what went wrong (if not already clear)
- Offer three options:
  1. **Fix now** — user wants to investigate and fix immediately; offer to help diagnose (read logs, grep code, suggest causes); wait for the user to indicate the fix is ready; then re-present the same test step from scratch
  2. **Add as TODO** — add to phase TODO file, continue testing remaining steps (final result still "passed with known issues")
  3. **Send back** — this failure blocks acceptance; item goes to Needs Rework

**For option 3 (send back):** Note the failed step number and the user's description. Collect failure details for the Needs Rework entry. Skip remaining test steps for this item and proceed to Step 6 (failure path).

**Unexpected behaviour not in test steps:**
- User reports something unexpected that wasn't a test step: ask whether it's a blocker (option 3) or a TODO (option 2).

### Step 5: All Steps Passed — Advance to Needs Integration

On main, update STATUS.md:
- Remove item from "Needs Testing"
- Add to "Needs Integration"

```
## Needs Integration
- <feature-name> | .claude/worktrees/<name> | <description>
```

Commit and push on main immediately so the Integrate agent can act on it.

If the user chose option 2 for any steps (non-blocking failures), invoke the `update-docs` skill to add each as a TODO to the appropriate phase TODO file, then commit that too.

### Step 6: Test Failed — Return to Needs Rework

On main, update STATUS.md:
- Remove item from "Needs Testing"
- Add to "Needs Rework" with specific feedback

```
## Needs Rework
- <feature-name> | .claude/worktrees/<name> | <description> | User: <step N failed — description>
```

Be specific: include the step number, what was done, and what went wrong. The Implement agent will read this to understand what to fix.

Commit and push on main.

### Step 7: Check for More Items

Ask user: "There are N more items in Needs Testing. Would you like to continue?" (or proceed automatically if user said "all" in Step 1).

When done testing, read STATUS.md and suggest next workflow per [Next Workflow Guidance](WORKFLOWS.md#next-workflow-guidance). If items moved to "Needs Integration", the Integrate workflow (8) can run immediately.

---

## Pre-Session Checklist

- [ ] STATUS.md "Needs Testing" section checked
- [ ] Item selected and confirmed with user
- [ ] Build succeeded in worktree before testing
- [ ] Test steps presented one at a time (not as a batch)
- [ ] Each step result confirmed before presenting the next
- [ ] Failures handled (fix/TODO/send back) before moving on
- [ ] STATUS.md updated after each item (don't batch updates)
- [ ] New TODOs added to phase file if option 2 chosen
- [ ] STATUS.md changes committed and pushed after each item

---

## Edge Cases

**User is not available to test a specific item:** Leave it in Needs Testing. Move to the next item if they want to test others, or exit.

**User starts testing, then has to stop:** Items already evaluated are updated in STATUS.md immediately (Step 5 or 6 after each item). Items not yet tested remain in "Needs Testing" unaffected.

**Worktree missing:** If the worktree path no longer exists on disk, the item cannot be tested. Move to "Needs Rework" with note "worktree missing — needs re-implementation".

**User finds a pre-existing bug unrelated to the feature:** This is a new bug, not a test failure. Add it to `docs/BUGS.md` and/or the phase TODO file. Do not send the feature back to Needs Rework for an unrelated issue.

**Build succeeds but binary behaves unexpectedly:** Treat as a test failure. Report the specific behaviour observed.

**Test instructions are ambiguous:** Ask the user to clarify what they did and what they saw before recording a pass or fail.

**Item has no test instructions in STATUS.md:** Ask the user what they want to verify. Record what was tested in the Needs Integration entry for traceability.
