# TDD Remediation 2 — Completed Items

> **Source:** coops-tdd audit 2026-03-17

---

## Item 3 — Delete `TestNewCacheManagerInitialisesFields` (2026-03-17)

**File:** `internal/isolation/cache_test.go`

Deleted `TestNewCacheManagerInitialisesFields` entirely. The test only asserted that unexported fields (`homeDir`, `cacheBaseDir`) were non-empty after `NewCacheManager()` — testing implementation, not behaviour. The observable outcome (that the manager works correctly) is already covered by `TestHomebrewCacheSetup` and other setup tests.

**Completion criteria met:**
- [x] `TestNewCacheManagerInitialisesFields` deleted (no replacement needed)
- [x] `go test ./...` passes (205 tests)

---

## Item 1 — Delete `TestNewTartClientSetsDefaults` (2026-03-17)

**File:** `internal/isolation/tart_test.go`

Deleted `TestNewTartClientSetsDefaults` entirely. The test only asserted that unexported fields (`installPrompt`, `pollInterval`, `pollTimeout`) were non-zero after `NewTartClient()` — testing implementation, not behaviour. The behaviour proof (that defaults are actually used) is covered by `Clone`, `IP`, and other operation tests.

**Completion criteria met:**
- [x] `TestNewTartClientSetsDefaults` deleted (no replacement needed)
- [x] `go test ./...` passes (204 tests)

---

## Item 7 — Fix `vmName` cleanup in config_test.go (2026-03-17)

**File:** `cmd/calf/config_test.go`

Replaced `t.Cleanup(func() { vmName = "" })` with `t.Cleanup(func() { _ = configShowCmd.Flags().Set("vm", "") })`. The old form only zeroed the bound Go variable, leaving cobra's internal `changed` flag as `true` and risking cross-test contamination. The new form resets both the value and cobra's state atomically via the cobra API.

**Completion criteria met:**
- [x] `vmName = ""` cleanup investigated — reset IS needed (cobra does not auto-reset between executions)
- [x] Replaced with cobra API reset: `configShowCmd.Flags().Set("vm", "")`
- [x] `go test ./...` passes (207 tests)

---

## Item 5 — Remove `cm.cacheBaseDir` accesses from cache_test.go (2026-03-17)

**File:** `internal/isolation/cache_test.go`

Replaced all `cm.cacheBaseDir` field accesses throughout `cache_test.go` with calls to the public `GetXxxCacheInfo()` methods, using `info.Path` for directory paths. Three patterns applied:
- **Pattern A** (verify dir created): `os.Stat(GetXxxCacheInfo().Path)`
- **Pattern B** (create fixture files): `filepath.Join(GetXxxCacheInfo().Path, "subdir")`
- **Pattern C** (Go cache subdirs): replaced `pkg/mod` + `pkg/sumdb` subdir assertions with `GetGoCacheInfo().Available`

Affected: `TestHomebrewCacheSetup`, `TestGetHomebrewCacheInfo`, `TestNpmCacheSetup`, `TestGetNpmCacheInfo`, `TestGoCacheSetup`, `TestGetGoCacheInfo`, `TestGitCacheSetup`, `TestGetGitCacheInfo`, `TestGetCachedGitRepos`, `TestCacheGitRepo`, `TestUpdateGitRepos`, `TestClearCache`, `TestCacheManagerWriterInjection`.

`TestNewCacheManagerWithDirs` (lines 179–180) retained for Item 4 which will rewrite that function entirely.

**Completion criteria met:**
- [x] All `cm.cacheBaseDir` accesses removed except `TestNewCacheManagerWithDirs` (Item 4)
- [x] `go test ./...` passes (204 tests)
- [x] `staticcheck ./...` clean

---

## Item 4 — Rewrite `TestNewCacheManagerWithDirs` (2026-03-17)

**File:** `internal/isolation/cache_test.go`

Rewrote `TestNewCacheManagerWithDirs` to test observable behaviour instead of unexported fields. The old test asserted `cm.homeDir` and `cm.cacheBaseDir` directly. The new test calls `SetupHomebrewCache()` and uses `os.Stat` to confirm the homebrew cache directory was created under the given `cacheBaseDir`. This proves both that `homeDir` was stored (non-empty → proceeds rather than no-ops) and that `cacheBaseDir` was stored (directory created at the expected path).

**Completion criteria met:**
- [x] `TestNewCacheManagerWithDirs` rewritten as behaviour test via `SetupHomebrewCache()` + `os.Stat`
- [x] No unexported fields accessed
- [x] `go test ./...` passes (204 tests)

---

## Item 6 — Remove shell script text assertions from `SetupVM*Cache` tests (2026-03-17)

**File:** `internal/isolation/cache_test.go`

Removed `commandsStr := strings.Join(commands, " ")` and all `strings.Contains(commandsStr, ...)` assertions from the "when host cache exists should return setup commands" subtest in `TestVMHomebrewCacheSetup`, `TestVMNpmCacheSetup`, `TestVMGoCacheSetup`, and `TestVMGitCacheSetup`. These were binding tests to exact shell command strings (e.g. `mount | grep -q`, `touch ~/.zshrc`, `HOMEBREW_CACHE`). Structural assertions (`commands != nil`, `len(commands) > 0`) retained. Nil-return subtests unchanged.

**Completion criteria met:**
- [x] All `SetupVM*Cache` shell script text assertions removed (13 lines deleted across 4 functions)
- [x] `go test ./...` passes (204 tests)
- [x] `staticcheck ./...` clean

---

## Item 8 — Fix fragile string slicing in config_test.go (2026-03-17)

**File:** `internal/config/config_test.go`

Replaced all 7 occurrences of `err.Error()[:len(expectedMsg)] != expectedMsg` with `!strings.HasPrefix(err.Error(), expectedMsg)`. Eliminates potential panics when error strings are shorter than expected, and uses stdlib over manual slicing.

**Completion criteria met:**
- [x] All `err.Error()[:len(x)]` slicing replaced with `strings.HasPrefix`
- [x] `go test ./...` passes (207 tests)
- [x] `go build` succeeds

---

## Item 2 — Replace `TestEnsureInstalled` with public Clone interface tests (2026-03-17)

**File:** `internal/isolation/tart_test.go`

Deleted `TestEnsureInstalled` (5 subtests calling unexported `client.ensureInstalled()` and asserting internal `client.tartPath` state). Also deleted the orphaned `// Integration tests` comment block.

Added 5 replacement test functions exercising the same scenarios through the public `Clone` method:
- `TestCloneWhenTartIsInstalled` — tart found via lookPath; dispatches "clone" command
- `TestCloneWhenTartIsNotInstalledAndUserDeclines` — user declines install; error contains "cancelled"
- `TestCloneWhenTartIsNotInstalledAndUserConfirmsAndBrewSucceeds` — brew install succeeds; Clone succeeds
- `TestCloneWhenTartIsNotInstalledAndBrewFails` — brew install fails; error contains "failed to install"
- `TestCloneWhenBrewIsNotAvailableAndTartNotFound` — no tart, no brew; returns error

Success-path tests override `runCommand` with a closure that calls `ensureInstalled` then delegates to `mockCommandRunner` — preserving real dispatch behaviour without exec. Failure-path tests rely on default `runTartCommand` returning early when `ensureInstalled` errors.

Code review simplified the tests by extracting the duplicated `runCommand` closure into `makeInstallingRunCommand(client, mock)` helper.

Also resolves BUG-011 (SA4031 nil check on initialised slice in `TestTartClient_RunWithCacheDirs_AcceptsCacheDirs`) — that test was already deleted in TDD-R2 Item 2 (internal/isolation/tart.go).

**TDD Remediation 2 is COMPLETE.** All items done (7: Items 7, 8, 3, 1, 5, 4, 6 + Item 2 today). 203 tests pass, `staticcheck` clean.

Net change: -131 lines test / +79 lines test (5 new functions + helper).

**Completion criteria met:**
- [x] `TestEnsureInstalled` deleted and replaced with 5 public-interface tests via `Clone`
- [x] `makeInstallingRunCommand` helper extracted
- [x] `go test ./...` passes (203 tests)
- [x] `staticcheck ./...` clean

---

## BUG-011: Nil Check on Initialised Slice (SA4031) — ✅ COMPLETED (2026-03-17)

- [x] BUG-011: SA4031 nil check on initialised slice in `TestTartClient_RunWithCacheDirs_AcceptsCacheDirs` (completed 2026-03-17)

Resolved as part of TDD-R2 Item 2 — the test containing this violation was deleted and replaced with behaviour-based tests. See Item 2 entry above.

---

## BUG-012: CacheManager Injectable Writer — ✅ COMPLETED (2026-03-17)

- [x] BUG-012: CacheManager writes directly to os.Stderr (untestable warnings) (completed 2026-03-17)

Added `writer io.Writer` field to `CacheManager`. `NewCacheManager()` and `NewCacheManagerWithDirs()` both default to `os.Stderr` (no production behaviour change). New `NewCacheManagerWithWriter(homeDir, cacheBaseDir, w)` constructor enables tests to inject `&bytes.Buffer{}`. All five `fmt.Fprintf(os.Stderr, ...)` calls in `SetupHomebrewCache`, `SetupNpmCache`, `SetupGoCache`, `SetupGitCache`, and `UpdateGitRepos` replaced with `fmt.Fprintf(c.writer, ...)`. `NewCacheManagerWithDirs` now delegates to `NewCacheManagerWithWriter` (eliminates struct literal duplication).

Added 6 new sub-tests in `TestCacheManagerWriterInjection` covering all four cache type warnings and the `UpdateGitRepos` per-repo warning.

Net change: +79 lines test / +14 lines production. All 207 tests pass.
