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

## Item 8 — Fix fragile string slicing in config_test.go (2026-03-17)

**File:** `internal/config/config_test.go`

Replaced all 7 occurrences of `err.Error()[:len(expectedMsg)] != expectedMsg` with `!strings.HasPrefix(err.Error(), expectedMsg)`. Eliminates potential panics when error strings are shorter than expected, and uses stdlib over manual slicing.

**Completion criteria met:**
- [x] All `err.Error()[:len(x)]` slicing replaced with `strings.HasPrefix`
- [x] `go test ./...` passes (207 tests)
- [x] `go build` succeeds
