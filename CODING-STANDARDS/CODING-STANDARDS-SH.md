# Shell Coding Standards

Shell-specific mandatory standards for CALF development.

See [CODING-STANDARDS.md](CODING-STANDARDS.md) for shared standards that also apply.

---

## Dependency Management

```bash
for tool in jq gh curl; do
    if ! command -v "$tool" &>/dev/null; then
        echo "Error: Required tool '$tool' is not installed"
        exit 1
    fi
done
```

---

## Error Handling

Never redirect errors to `/dev/null`. Log errors even when not shown to users. Provide actionable messages.

**Wrong:**
```bash
git clone "$repo" &>/dev/null
```

**Correct:**
```bash
if ! git clone "$repo" 2>&1; then
    echo "Error: Failed to clone $repo. Check network and SSH keys."
    return 1
fi
```

---

## Proactive Validation

Validate auth and connectivity before performing remote operations.

```bash
if ! ssh -T git@github.com 2>&1 | grep -q "successfully authenticated"; then
    echo "Error: GitHub SSH authentication failed"
    echo "Run: ssh-keygen && gh auth login"
    return 1
fi
```

---

## Security Practices

Never use `eval`. Quote all variables. Sanitize external input before use in commands.

**Wrong:**
```bash
eval echo "$some_path"
```

**Correct:**
```bash
target_dir="${HOME}/code"
```

---

## `trap` Handlers

Each new `trap ... EXIT` call **replaces** the previous one — it does not chain. When a script creates multiple temporary resources at different points, every subsequent `trap` must include all earlier resources.

**Wrong — TMP leaks if script exits after TMP2 is created:**
```bash
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT
# ... later ...
TMP2=$(mktemp -d)
trap 'rm -rf "$TMP2"' EXIT   # silently drops TMP cleanup
```

**Correct — update the handler to cover all live resources:**
```bash
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT
# ... later ...
TMP2=$(mktemp -d)
trap 'rm -rf "$TMP" "$TMP2"' EXIT
```

---

## Mandatory Test Scenarios

Valid inputs · invalid inputs · missing dependencies · auth failures · existing state · network failures.

---

## Code Review Checklist — Shell

In addition to shared checklist items:

- [ ] No `eval`; all shell variables quoted
- [ ] Each `trap EXIT` update covers all previously registered resources — no silent drops
- [ ] All six mandatory test scenarios covered
