# Coding Standards

Mandatory standards for CALF development, derived from code review findings.

Standards that apply to all languages are here. Language-specific standards are in separate files.

---

## Code Duplication

Never leave copy-paste artifacts. Extract repeated logic into functions. Use `git diff` before committing to catch unintended duplications.

---

## Dependency Management

Verify external tools exist before use and give clear error messages when missing. See language-specific files for idiomatic patterns.

---

## Documentation Accuracy

- Verify code implements what docs describe; update docs immediately when behaviour changes
- Include `TODO:` status for planned features not yet implemented
- Review PR descriptions against actual code changes
- Allowed to document default dev credentials — label as "default" and note how to change

---

## Error Handling

Never redirect errors to `/dev/null`. Log errors even when not shown to users. Provide actionable messages. See language-specific files for idiomatic patterns.

---

## Proactive Validation

Validate preconditions before attempting operations: auth before network, permissions before filesystem, connectivity before remote.

---

## Security Practices

Never use `eval`. Quote all variables in shell scripts. Sanitize external input before use in commands.

---

## Testing Requirements

### Invoke `coops-tdd` skill before any code change

The `coops-tdd` skill is mandatory before writing any code. It covers: `when...should...` naming, Arrange/Act/Assert structure, public-interface-only rule, mock rules, and Red/Green/Refactor cycle. Code written without invoking it must not be committed.

### Mandatory test scenarios — all languages

Every change must cover: success path · all error return paths · edge/boundary conditions · component interactions (where applicable).

---

## Code Review Checklist

Before submitting code for review, verify:

- [ ] `coops-tdd` skill invoked before writing any code
- [ ] No duplicate code blocks or copy-paste errors
- [ ] All external dependencies checked before use
- [ ] Documentation matches implementation; planned features marked `TODO:`
- [ ] Errors never silently suppressed
- [ ] Preconditions validated before operations
- [ ] All test scenarios executed; language linters/build pass

See language-specific files for additional checklist items.

---

## Language Index

| Language | Standards | Patterns |
|----------|-----------|----------|
| Go | [CODING-STANDARDS-GO.md](CODING-STANDARDS-GO.md) | [CODING-STANDARDS-GO-PATTERNS.md](CODING-STANDARDS-GO-PATTERNS.md) |
| Shell | [CODING-STANDARDS-SH.md](CODING-STANDARDS-SH.md) | standards pending |
