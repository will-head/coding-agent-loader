# How to Update Coding Standards

Standards are managed via the `coding-standards` skill and the `CODING-STANDARDS/` directory in the project root. The process is designed to promote patterns organically from code-review findings rather than requiring manual updates.

---

## Directory Structure

```
CODING-STANDARDS/
├── CODING-STANDARDS.md                  # Shared standards + language index
├── CODING-STANDARDS-GO.md               # Go-specific standards
├── CODING-STANDARDS-SH.md               # Shell-specific standards
├── CODING-STANDARDS-GO-PATTERNS.md      # Pattern tracking for Go
└── CODING-STANDARDS-SH-PATTERNS.md      # Pattern tracking for Shell
```

New language files are created on demand when a pattern for that language is first observed.

---

## Normal Flow — Patterns Promoted Automatically

### 1. code-review writes patterns

After each code review, the `code-review` skill writes observed issues to `CODING-STANDARDS-[LANG]-PATTERNS.md`:
- Semantically similar issues are merged (count incremented + new example added)
- New issues start at count 1

### 2. coding-standards promotes at threshold 3

When a pattern reaches count ≥ 3, invoke the `coding-standards` skill explicitly:

```
Use the coding-standards skill to promote patterns
```

The skill will:
1. Read all PATTERNS files
2. Promote any count ≥ 3 pattern to `CODING-STANDARDS-[LANG].md`
3. Mark promoted patterns in the PATTERNS file
4. Run a file audit

---

## Manual Addition

For a new standard that doesn't come from a review finding (e.g. a new tool or framework requirement):

1. Add it directly to the relevant `CODING-STANDARDS-[LANG].md` using the entry format:
   ```markdown
   ## [Rule name]
   **Rule:** One-sentence imperative statement.
   **Why:** Brief rationale.
   **Wrong:** ...
   **Correct:** ...
   ```
2. If it's shared across all languages, add it to `CODING-STANDARDS.md`
3. Update the index table in `CODING-STANDARDS.md` if a new language file was created

---

## Identifying Good Patterns

Extract the underlying pattern from specific findings:

- **Too specific:** "Line 182 in vm-auth.sh doesn't check for jq"
- **Good pattern:** "Scripts use external tools without checking if they're installed"

Questions to ask:
- Could this happen in other parts of the codebase?
- Is this a symptom of a larger anti-pattern?
- Does this apply to one language or all code types?

---

## File Size and Token Audit

Run `Use the coding-standards skill to audit file sizes` periodically or after promotion. Target: shared + one language file + patterns file under ~600 lines combined.

---

## References

- [CODING-STANDARDS/](../CODING-STANDARDS/CODING-STANDARDS.md) — The standards directory
- [AGENTS.md](../AGENTS.md) — Agent instructions
- [WORKFLOWS.md](WORKFLOWS.md) — Code review and workflow procedures
