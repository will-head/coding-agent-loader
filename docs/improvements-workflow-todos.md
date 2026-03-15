# Workflow Improvement TODOs

> Meta-level improvements to the CALF project workflows and documentation management

This document tracks improvements to the workflow system itself, not feature implementation. These are tools and processes that help manage the project more effectively.

---

## Project Management Skills

### 1. Skill: `continue later`

**Goal:** Save complete session state for easy resumption across agent sessions.

**Purpose:** Enable pausing work and resuming later with full context preserved.

**Requirements:**
- Save complete state of current work
- Document everything achieved in current session
- Capture any issues or blockers encountered
- Create resumption context for next session
- **Cross-agent compatible format** - any agent can pick up where another left off
- Store state in standardized location
- Include timestamp and agent metadata

**Use Cases:**
- Ending a work session with incomplete work
- Need to pause and resume later (hours/days)
- Switching between different tasks or workflows
- Preserve context across different agent sessions
- Handoff between agents (e.g., start work in one agent, continue in another)

**Implementation Notes:**
- Consider JSON or YAML format for state storage
- Include: current workflow, current step, context, achievements, blockers, next steps
- Location: project-specific (e.g., `.calf/session-state.json`)
- Should be git-ignored (local state only)

---

### 2. Skill: `update docs`

**Goal:** Ensure all project documentation stays current and accurate.

**Purpose:** Automatically detect and update outdated documentation after code changes.

**Requirements:**
- Scan for documentation that may be outdated
- Check if recent changes need doc updates
- Verify consistency across documentation files
- Update affected documentation files
- **NEVER touch ADRs** (immutable historical records)
- Show diff preview before applying changes
- Allow selective application of updates

**Use Cases:**
- After completing significant features
- When code changes affect documented behavior
- Periodically to maintain doc accuracy
- Before major releases or milestones
- After merging PRs that change behavior

**Implementation Notes:**
- Check git history for recent changes
- Map code changes to documentation files
- Verify examples in docs still work
- Update command references, file paths, etc.
- Exclude: `docs/adr/*`, `docs/prd/*`

---

### 3. Skill: `refactor docs`

**Goal:** Optimize documentation for progressive disclosure and token efficiency.

**Purpose:** Restructure documentation to minimize token usage during normal operations while preserving deep technical content for research.

**Requirements:**
- Refactor docs for progressive disclosure (high-level → details)
- Improve information architecture
- Reduce token usage during normal operations
- Move deep technical details to appropriate locations
- **NEVER touches ADRs** - these are immutable and only read for:
  - Deep research tasks
  - Implementation planning
  - NOT during normal operation
- Preserve all content (no deletion, only reorganization)
- Maintain all links and cross-references

**Principles:**
- **Progressive disclosure:** Start with essentials, link to details
- **Token efficiency:** Most common operations use fewest tokens
- **Deep content preservation:** Technical details moved to specialized files, not deleted
- **ADR protection:** ADRs remain untouched as historical records
- **Layered information:** Quick reference → How-to → Deep dive → Historical context

**Use Cases:**
- Documentation has grown large or complex
- Finding information takes too many token reads
- New users struggle with information overload
- Periodic maintenance to improve usability
- After major feature additions that bloat docs

**Implementation Notes:**
- Analyze token costs of common operations
- Identify content read frequently vs. rarely
- Create summary/overview files
- Move details to appendices or separate files
- Update cross-references and navigation
- Test: Can agent complete common tasks with fewer reads?

---

## Implementation Priority

1. **`continue later`** - Highest priority (enables better work management)
2. **`update docs`** - Medium priority (quality of life improvement)
3. **`refactor docs`** - Lower priority (optimization, do when docs become unwieldy)

