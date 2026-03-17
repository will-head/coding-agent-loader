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
| 1 | `internal/isolation/tart_test.go` | High | `TestNewTartClientSetsDefaults` tests unexported fields only |
| 2 | `internal/isolation/tart_test.go` | High | `TestEnsureInstalled` tests unexported method `ensureInstalled()` |
| 3 | `internal/isolation/cache_test.go` | High | `TestNewCacheManagerInitialisesFields` tests unexported fields only |
| 4 | `internal/isolation/cache_test.go` | High | `TestNewCacheManagerWithDirs` tests unexported fields only |
| 5 | `internal/isolation/cache_test.go` | High | Many tests access `cm.cacheBaseDir` (unexported) to build fixture paths |
| 6 | `internal/isolation/cache_test.go` | Medium | `SetupVM*Cache` tests assert on exact shell script text |
| 7 | `cmd/calf/config_test.go` | Low | `t.Cleanup` directly resets package-level `vmName` variable by name |
| 8 | `internal/config/config_test.go` | Low | ~~Fragile `err.Error()[:len(expectedMsg)]` string slicing~~ ✅ Done |

---

## Item 1 — Delete `TestNewTartClientSetsDefaults`

**File:** `internal/isolation/tart_test.go` (lines 83–97)

**Problem:** The test only asserts that unexported fields are non-zero after construction:
```go
if client.installPrompt == "" { ... }
if client.pollInterval == 0 { ... }
if client.pollTimeout == 0 { ... }
```
This is not a behaviour test. If any of these field names change, the test breaks for the wrong reason. The behaviour proof is that operations like `Clone` and `IP` actually use the defaults (covered by their own tests with overrides).

**Action:** Delete `TestNewTartClientSetsDefaults` entirely. No replacement needed.

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

## Item 3 — Delete `TestNewCacheManagerInitialisesFields`

**File:** `internal/isolation/cache_test.go` (lines 153–168)

**Problem:** The test only asserts that unexported fields are non-empty after `NewCacheManager()`:
```go
if cm.homeDir == "" { ... }
if cm.cacheBaseDir == "" { ... }
```
This tests implementation, not behaviour. The behaviour proof is that `SetupHomebrewCache()` and friends succeed, which is already covered elsewhere.

**Action:** Delete `TestNewCacheManagerInitialisesFields` entirely. No replacement needed.

---

## Item 4 — Rewrite `TestNewCacheManagerWithDirs`

**File:** `internal/isolation/cache_test.go` (lines 171–191)

**Problem:** Asserts that injected values were stored in unexported fields:
```go
if cm.homeDir != homeDir { ... }
if cm.cacheBaseDir != cacheBaseDir { ... }
```

**Action:** Replace with a behaviour test that proves the injected dirs are *used*:

```
TestNewCacheManagerWithDirs
  when dirs provided should set up homebrew cache under the provided cache base dir
  - Arrange: homeDir = t.TempDir(), cacheBaseDir = filepath.Join(homeDir, "cache")
  - Act: cm = NewCacheManagerWithDirs(homeDir, cacheBaseDir)
  - Act: err = cm.SetupHomebrewCache()
  - Assert: err is nil
  - Act: info, err = cm.GetHomebrewCacheInfo()
  - Assert: err is nil
  - Assert: info.Path contains cacheBaseDir (proves the injected dir is used, not the default)
  - Assert: info.Available is true (proves cache was created)
```

---

## Item 5 — Remove access to `cm.cacheBaseDir` throughout cache_test.go

**File:** `internal/isolation/cache_test.go`

**Problem:** Many tests access the unexported `cm.cacheBaseDir` to construct paths for creating fixture files or asserting directory existence. Full list of affected locations:

| Test function | Use of `cm.cacheBaseDir` | Purpose |
|---|---|---|
| `TestHomebrewCacheSetup` (line 29) | `filepath.Join(cm.cacheBaseDir, "homebrew")` | Verify dir created |
| `TestGetHomebrewCacheInfo` (line 97) | `filepath.Join(cm.cacheBaseDir, "homebrew")` | Create fixture file |
| `TestNpmCacheSetup` (line 366) | `filepath.Join(cm.cacheBaseDir, "npm")` | Verify dir created |
| `TestGetNpmCacheInfo` (line 431) | `filepath.Join(cm.cacheBaseDir, "npm")` | Create fixture file |
| `TestGoCacheSetup` (line 522) | `filepath.Join(cm.cacheBaseDir, "go")` | Verify dir created |
| `TestGetGoCacheInfo` (lines 598, 598) | `filepath.Join(cm.cacheBaseDir, "go", "pkg", "mod")` | Create fixture file |
| `TestGitCacheSetup` (line 692) | `filepath.Join(cm.cacheBaseDir, "git")` | Verify dir created |
| `TestGetGitCacheInfo` (lines 758, 762) | `filepath.Join(cm.cacheBaseDir, "git", ...)` | Create fixture file |
| `TestGetCachedGitRepos` (lines 860, 863) | `filepath.Join(cm.cacheBaseDir, "git", ...)` | Create fixture dirs |
| `TestCacheGitRepo` (line 910) | `filepath.Join(cm.cacheBaseDir, "git", ...)` | Create fixture dir |
| `TestUpdateGitRepos` (lines 976, 1002, 1044) | `filepath.Join(cm.cacheBaseDir, "git", ...)` | Create fixture dirs |
| `TestClearCache` (lines 1154, 1173, 1189) | `filepath.Join(cm.cacheBaseDir, "homebrew")` | Create fixture file, assert dir state |
| `TestCacheManagerWriterInjection` (line 1126) | `filepath.Join(localCm.cacheBaseDir, "git", ...)` | Create fixture dir |

**Fix strategy — two patterns:**

**Pattern A — Verifying directory creation:** Instead of constructing the path from `cm.cacheBaseDir` and calling `os.Stat`, use the public info method:
```go
// Before:
hostCacheDir := filepath.Join(cm.cacheBaseDir, "homebrew")
info, err := os.Stat(hostCacheDir)
if !info.IsDir() { ... }

// After:
info, err := cm.GetHomebrewCacheInfo()
if err != nil { t.Fatalf(...) }
if info.Path == "" { t.Fatalf("expected non-empty path") }
// Verify the directory exists via the public path:
stat, err := os.Stat(info.Path)
if err != nil || !stat.IsDir() { t.Fatalf("expected cache directory to exist and be a dir") }
```

**Pattern B — Creating fixture files for size tests:** Get the path from the info method after setup:
```go
// Before:
hostCacheDir := filepath.Join(cm.cacheBaseDir, "homebrew")
testFile := filepath.Join(hostCacheDir, "test-file.bin")

// After:
info, err := cm.GetHomebrewCacheInfo()
if err != nil { t.Fatalf(...) }
testFile := filepath.Join(info.Path, "test-file.bin")
```

**Pattern C — Verifying subdirectory creation (TestGoCacheSetup):**
Currently checks that `pkg/mod` and `pkg/sumdb` subdirs exist under `cacheBaseDir`. These are internal details of Go cache structure. Replace with a single outcome assertion: after `SetupGoCache()`, `GetGoCacheInfo().Available` is true and `Path` is non-empty. The specific subdirectory layout is an implementation detail.

**Apply all three patterns throughout the file.** Each affected test function must be updated to eliminate all references to `cm.cacheBaseDir`.

---

## Item 6 — Remove shell script text assertions from `SetupVM*Cache` tests

**File:** `internal/isolation/cache_test.go`

**Affected test functions:**
- `TestVMHomebrewCacheSetup` (lines 237–286)
- `TestVMNpmCacheSetup` (lines 452–505)
- `TestVMGoCacheSetup` (lines 619–675)
- `TestVMGitCacheSetup` (lines 782–832)

**Problem:** Each "when host cache exists should return setup commands" subtest joins all commands into a string and then asserts on specific shell fragments:
```go
commandsStr := strings.Join(commands, " ")
if !strings.Contains(commandsStr, "mount | grep -q \" on $HOME/.calf-cache \"") { ... }
if !strings.Contains(commandsStr, "touch ~/.zshrc") { ... }
if !strings.Contains(commandsStr, "HOMEBREW_CACHE") { ... }
```

These bind the test to the exact shell implementation. Any refactoring of the shell commands (even if the VM behaviour is identical) will break these tests.

**Action:** Replace the specific string assertions with structural assertions only:
```go
// Before (remove all of these):
commandsStr := strings.Join(commands, " ")
if !strings.Contains(commandsStr, "mount | grep -q ...") { ... }
if !strings.Contains(commandsStr, "test -d") { ... }
if !strings.Contains(commandsStr, "HOMEBREW_CACHE") { ... }

// After (keep only structural assertions):
if len(commands) == 0 {
    t.Fatalf("expected at least one setup command, got empty slice")
}
```

The nil/non-nil and nil/empty-list tests ("when home dir is unavailable should return nil", "when host cache does not exist should return nil") are correct behaviour tests — keep them unchanged.

---

## Execution Order

Work through items in this order to keep the test suite green throughout:

1. ~~**Item 8** — Mechanical find-and-replace, lowest risk.~~ ✅ Done
2. **Item 7** — Investigate cobra flag reset, make minimal change. Run `go test ./...` after.
3. **Item 3** — Delete one test function. Run `go test ./...` after.
4. **Item 1** — Delete one test function. Run `go test ./...` after.
5. **Item 5** — Refactor cache_test.go fixture setup (no logic change, just path sourcing). Run `go test ./...` after each test function updated.
6. **Item 4** — Rewrite `TestNewCacheManagerWithDirs`. Run `go test ./...` after.
7. **Item 6** — Remove shell script text assertions. Run `go test ./...` after.
8. **Item 2** — Rewrite `TestEnsureInstalled` as public-interface tests. This is the most significant change. Run `go test ./...` after each new test written.

After all items complete, run `go test ./...` and `staticcheck ./...` to confirm clean.

---

## Completion Criteria

- [ ] `TestNewTartClientSetsDefaults` deleted
- [ ] `TestEnsureInstalled` replaced with 5 public-interface tests via `Clone`
- [ ] `TestNewCacheManagerInitialisesFields` deleted
- [ ] `TestNewCacheManagerWithDirs` rewritten as behaviour test using `GetHomebrewCacheInfo`
- [ ] All `cm.cacheBaseDir` accesses removed from cache_test.go
- [ ] All `SetupVM*Cache` shell script text assertions replaced with structural assertions
- [x] All `err.Error()[:len(x)]` slicing replaced with `strings.HasPrefix` — see DONE file
- [ ] `go test ./...` passes with no failures
- [ ] `staticcheck ./...` passes with no warnings
