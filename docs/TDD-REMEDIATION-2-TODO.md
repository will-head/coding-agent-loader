# TDD Remediation 2 — Behaviour vs Implementation

> **Source:** coops-tdd audit 2026-03-17
>
> **Scope:** Fix all test violations where tests assert on internal state, access unexported fields, or test private methods rather than observable behaviour. Tests may be rewritten or deleted entirely.
>
> **Reference assessment:** coops-tdd report in session history (2026-03-17)

---

## Summary of Violations

| # | File | Severity | Violation |
|---|------|----------|-----------|
| 2 | `internal/isolation/tart_test.go` | High | `TestEnsureInstalled` tests unexported method `ensureInstalled()` |

---

## Item 2 — Rewrite `TestEnsureInstalled`

**File:** `internal/isolation/tart_test.go` (lines 651–781)

**Problem:** All 5 subtests call the unexported method `client.ensureInstalled()` directly and then inspect internal state (`client.tartPath`):
```go
err := client.ensureInstalled()
// ...
if client.tartPath == "" { ... }
```

**Action:** Replace all 5 subtests with tests that exercise the same conditions through public methods. The injectable fields (`lookPath`, `stdinReader`, `runBrewCommand`) are already available on `TartClient` — use them with `Clone` as the entry point instead.

**New tests to write (replace the entire `TestEnsureInstalled` function):**

```
TestCloneWhenTartIsInstalled
  when tart is found via lookPath should execute the clone command
  - Arrange: client with lookPath that returns "/usr/local/bin/tart" for "tart"
  - Arrange: mock.addOutput("clone test-image test-vm", "")
  - Act: client.Clone("test-image", "test-vm")
  - Assert: err is nil
  - Assert: mock.commands[0] contains "clone" (tart was dispatched, not just path set)

TestCloneWhenTartIsNotInstalledAndUserDeclines
  when tart is not found and user declines install should return error
  - Arrange: client with lookPath that returns error for all files
  - Arrange: stdinReader = strings.NewReader("n\n")
  - Act: client.Clone("test-image", "test-vm")
  - Assert: err is non-nil
  - Assert: err.Error() contains "cancelled"

TestCloneWhenTartIsNotInstalledAndUserConfirmsAndBrewSucceeds
  when tart is not found and user confirms and brew install succeeds should execute clone
  - Arrange: lookPath returns error for "tart" on first call, success on second call
  - Arrange: stdinReader = strings.NewReader("y\n")
  - Arrange: runBrewCommand returns nil (success)
  - Arrange: mock.addOutput("clone test-image test-vm", "")
  - Act: client.Clone("test-image", "test-vm")
  - Assert: err is nil

TestCloneWhenTartIsNotInstalledAndBrewFails
  when tart is not found and brew install fails should return install error
  - Arrange: lookPath returns error for all files
  - Arrange: stdinReader = strings.NewReader("y\n")
  - Arrange: runBrewCommand returns fmt.Errorf("brew install failed")
  - Act: client.Clone("test-image", "test-vm")
  - Assert: err is non-nil
  - Assert: err.Error() contains "failed to install"

TestCloneWhenBrewIsNotAvailableAndTartNotFound
  when neither tart nor brew is found should return error without prompting
  - Arrange: lookPath returns error for all files including "brew"
  - Arrange: stdinReader = strings.NewReader("") (no input should be consumed)
  - Act: client.Clone("test-image", "test-vm")
  - Assert: err is non-nil
```

**Note on `createTestClient` helper:** The existing helper pre-sets `client.tartPath` to skip `ensureInstalled` entirely. Keep this as-is for all tests that are already working correctly — the new `TestClone*` tests above must NOT use `createTestClient`, they must use `NewTartClient()` directly with only the injectable fields set.

---

## Execution Order

Work through items in this order to keep the test suite green throughout:

1. **Item 5** — Refactor cache_test.go fixture setup (no logic change, just path sourcing). Run `go test ./...` after each test function updated.
2. **Item 4** — Rewrite `TestNewCacheManagerWithDirs`. Run `go test ./...` after.
3. **Item 6** — Remove shell script text assertions. Run `go test ./...` after.
4. **Item 2** — Rewrite `TestEnsureInstalled` as public-interface tests. This is the most significant change. Run `go test ./...` after each new test written.

After all items complete, run `go test ./...` and `staticcheck ./...` to confirm clean.

---

## Completion Criteria

- [x] `TestEnsureInstalled` replaced with 5 public-interface tests via `Clone` (completed 2026-03-17)
- [x] All `SetupVM*Cache` shell script text assertions replaced with structural assertions (completed 2026-03-17)
- [x] `go test ./...` passes with no failures (203 tests pass)
- [x] `staticcheck ./...` passes with no warnings

**TDD Remediation 2 is COMPLETE (2026-03-17).**
