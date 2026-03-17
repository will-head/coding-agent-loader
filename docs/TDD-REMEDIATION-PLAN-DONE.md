# TDD Remediation Plan — Completed Items

> **Source:** [TDD-REMEDIATION-PLAN.md](TDD-REMEDIATION-PLAN.md) items completed 2026-03-16 – 2026-03-17

---

## Pre-work, Items 0–10 (2026-03-16 – 2026-03-17)

### Pre-work: Fix go.mod Version

Corrected `go.mod` from invalid version `go 1.25.6` to match the installed Go toolchain. `go build ./...` and `go test ./...` confirmed unaffected.

### Item 0: Rename Existing Tests to Behavioral Convention

Renamed all existing tests across `internal/config/config_test.go`, `internal/isolation/tart_test.go`, and `internal/isolation/cache_test.go` to the `"when [condition] should [outcome]"` sub-test format and concise PascalCase top-level grouping names. Second pass enforced Arrange / Act / Assert structure with blank-line separation. No test logic changed; all tests remain green.

### Item 1: internal/config — GetDefaultConfigPath and GetVMConfigPath

Added direct tests for the two path-returning functions in `internal/config/config_test.go`:

**`TestGetDefaultConfigPath`**
- `"when home dir is available should return path ending in .calf/config.yaml"`
- `"when home dir is available should return an absolute path"`

**`TestGetVMConfigPath`**
- `"when vm name is provided should return path containing the vm name"`
- `"when vm name is provided should return path ending in vm.yaml"`
- `"when vm name is provided should return an absolute path"`

No production code changes. Error-path tests deferred — both functions call `os.UserHomeDir()` directly with no injection point; adding one would be speculative.

All 29 tests in `internal/config` pass.

### Item 2: internal/isolation/tart.go — Run and RunWithCacheDirs (2026-03-16)

Deleted 6 no-op and implementation-detail tests:
- `TestTartClientConstants` — asserted value of unexported `cacheDirMount` constant
- `TestTartClient_Run_Headless` — no-op; asserted nothing about command construction
- `TestTartClient_Run_VNC_UsesExperimental` — no-op; signature check only
- `TestRunCommandConstruction` — no-op; only checked `client != nil`
- `TestCacheSharingAlwaysAdded` — redundantly asserted `cacheDirMount` constant value
- `TestTartClient_RunWithCacheDirs_AcceptsCacheDirs` — no-op; nil check on initialised slice (SA4031)

Added 9 behavioral sub-tests:

**`TestRun`**
- `"when called with headless true should pass --headless flag to tart run"`
- `"when called with headless false should not pass --headless flag"`
- `"when called with vnc true should pass --vnc-experimental flag"`
- `"when called with vnc false should not pass --vnc-experimental flag"`
- `"when called should pass vm name as argument"`

**`TestRunWithCacheDirs`**
- `"when called with cache dirs should include --dir flag for each directory"`
- `"when called should always include the cache sharing directory"`
- `"when called with empty cache dirs should still include cache sharing directory"`
- `"when called should pass vm name as argument"`

Production change: `RunWithCacheDirs` previously called `exec.Command` directly, bypassing the injectable `runCommand` field. Replaced with `c.runCommand(args...)` and removed the explicit `ensureInstalled()` call (production `runTartCommand` handles it). `Run` already delegated to `RunWithCacheDirs` — no change needed there. Added `sliceContains` helper for partial argument matching in tests.

All 162 tests pass. `go vet` clean.

### Item 4: internal/isolation/cache.go — CacheManager Injectable Constructor (2026-03-16)

Added `NewCacheManagerWithDirs(homeDir, cacheBaseDir string) *CacheManager` constructor to `internal/isolation/cache.go`. `NewCacheManager` now delegates to it, eliminating the duplicate struct literal.

Replaced all 34 `&CacheManager{homeDir: ..., cacheBaseDir: ...}` struct literal constructions in `cache_test.go` with `NewCacheManagerWithDirs(...)` calls. No test logic changed — construction mechanism only.

Added 1 new behavioral test:

**`TestNewCacheManagerWithDirs`**
- `"when dirs provided should initialise with given home and cache base dirs"`

All 168 tests pass. `staticcheck` clean (pre-existing BUG-010 in `config.go` unrelated).

---

### Item 3: internal/isolation/tart.go — ensureInstalled Homebrew Branch (2026-03-16)

Added 3 injectable fields to `TartClient`:
- `stdinReader io.Reader` — defaults to `os.Stdin`; used in `ensureInstalled` instead of hardcoded `bufio.NewReader(os.Stdin)`
- `lookPath func(string) (string, error)` — defaults to `exec.LookPath`; replaces direct `exec.LookPath` calls for both `"tart"` and `"brew"` checks
- `runBrewCommand commandRunner` — defaults to a closure that resolves brew via `c.lookPath` (fixes silent failure on Apple Silicon where brew is at `/opt/homebrew/bin/brew`); replaces direct `exec.Command("brew", ...)` call

Added 4 behavioral sub-tests to `TestEnsureInstalled`:
- `"when tart is found on path should set tart path without prompting"`
- `"when tart is not found and user declines install should return error"`
- `"when tart is not found and user confirms install and brew succeeds should update tart path"`
- `"when tart is not found and user confirms install and brew fails should return error"`

All 137 tests pass. `go vet` clean.

### Item 5: internal/isolation/cache.go — Fix Sub-Test Isolation in TestClearCache (2026-03-16)

- [x] Fix `TestCacheManager_Clear` sub-test isolation (shared state) (completed 2026-03-16)

Restructured `TestClearCache` so every sub-test creates its own `t.TempDir()` and `NewCacheManagerWithDirs(...)`. Previously, a single `cm` and `tmpDir` were shared across all seven sub-tests, creating implicit ordering dependencies and shared mutable state.

Changes (test-only — no production code):
- Each sub-test now fully self-contained with its own dirs and `CacheManager`
- Symlink sub-test no longer borrows `tmpDir` from outer scope
- `switch cacheType` dispatch replaced with `map[string]func() error` (removes 4 near-identical branches)
- Added `else t.Fatalf` guard so mismatched `testCases`/`setupFuncs` fails loudly
- Removed redundant `os.Stat` before `os.ReadDir` in symlink sub-test (TOCTOU pre-check)
- Removed dead `os.Chmod(readOnlyFile)` from defer (file already deleted by `Clear`)
- Checked error from `filepath.EvalSymlinks` (was silently discarded)
- Verified every sub-test passes independently via `-run TestClearCache/<name>`

All 168 tests pass.

### Item 6: internal/isolation/cache.go — Eliminate GetXxxCacheInfo Duplication (2026-03-17)

- [x] Extract `getCacheInfo` helper to eliminate 4× duplication (completed 2026-03-17)

Extracted private `getCacheInfo(cachePath string) (*CacheInfo, error)` helper encapsulating the shared pattern across `GetHomebrewCacheInfo`, `GetNpmCacheInfo`, `GetGoCacheInfo`, and `GetGitCacheInfo`. Each public method now delegates in a single line. Also removed orphaned GoDoc fragment from the previously deleted `getSharedVolumeCachePath` function.

Net change: -113 lines / +12 lines. All 168 tests pass.

### Item 7: internal/isolation/cache.go — UpdateGitRepos Error Contract (2026-03-17)

- [x] Fix `UpdateGitRepos` to surface errors to caller instead of always returning `nil` (completed 2026-03-17)

Added `rev-parse --git-dir` pre-check to skip non-git directories silently (preserving existing behaviour). Added `failed` counter; when any valid git repo fails to fetch, returns `fmt.Errorf("failed to update %d of %d repos", failed, updated+failed)` while still processing remaining repos and returning the partial success count.

Added 4 sub-tests to `TestUpdateGitRepos` (merged to 3 after code review): success path, single-failure error return, and combined partial-count + continues-past-failure. Extracted `makeBadGitRepo` test helper to eliminate duplicate arrange blocks.

Net change: +10 lines production / +91 lines test. All 171 tests pass.

### Item 8: cmd/calf/config.go — runConfigShow RunE Conversion (2026-03-17)

- [x] Add tests for `cmd/calf/config.go`; convert `os.Exit` to `RunE` (completed 2026-03-17)

Converted `configShowCmd` from `Run:` to `RunE:`. Replaced all three `os.Exit(1)` calls with `return fmt.Errorf(...)`. Changed all output from bare `fmt.Println`/`fmt.Printf` to `fmt.Fprintln`/`fmt.Fprintf` through `cmd.OutOrStdout()` so tests can capture output via `rootCmd.SetOut`.

Created `cmd/calf/config_test.go` with 7 sub-tests in `TestConfigShow`. Extracted `setupConfigShow` helper to eliminate repeated rootCmd wiring. Tests use `t.Setenv("HOME", t.TempDir())` to isolate config paths without mocking. Key discovery: cobra's `ExecuteC()` on a child command delegates to `Root().ExecuteC()` — tests must call `rootCmd.Execute()`, not `configShowCmd.Execute()`.

Net change: +1 new test file (165 lines) / +8 lines production. All 179 tests pass.

### Item 9: cmd/calf/cache.go — runCacheClear stdin injectable (2026-03-17)

- [x] Add tests for `cmd/calf/cache.go`; make stdin injectable (completed 2026-03-17)

Replaced package-level `var cacheCmd/cacheClearCmd/clearAll/force/dryRun` globals with a `newCacheCmd(stdin io.Reader, homeDir string) *cobra.Command` factory. Added `newCacheManager(homeDir string)` helper that delegates to `NewCacheManagerWithDirs` when a non-empty homeDir is provided, enabling tests to point the cache manager at `t.TempDir()` directories. All `fmt.Print*` calls replaced with `cmd.OutOrStdout()` / `cmd.ErrOrStderr()`.

Added per-type flags `--homebrew`, `--npm`, `--go`, `--git` (new behavior) driven by test requirements. Added `cacheTypeFlags` struct with `anySet()` helper to cleanly encapsulate per-type flag state. Added `GoDoc` comment to `runCacheClear`.

Created `cmd/calf/cache_test.go` with 12 sub-tests in `TestCacheClear` covering: force, dry-run (output and file preservation), all+confirm, all+decline, per-type flags (×4), per-type+dry-run, all+force, all+dry-run. Extracted `setupCacheCmd` and `setupCacheCmdWithDirs` helpers.

Net change: +1 new test file (270 lines) / +80 lines production. All 192 tests pass.

### Item 10: cmd/calf/main.go — Root Command Dispatch Tests (2026-03-17)

- [x] Add root command dispatch tests for `cmd/calf/main.go` (completed 2026-03-17)

Created `cmd/calf/main_test.go` with 5 sub-tests across 3 top-level functions covering the outermost public surface of the CLI:

**`TestRootCommand`**
- `"when no args provided should print usage information"`
- `"when help flag provided should print help text"`
- `"when unknown subcommand provided should return error"`

**`TestConfigSubcommand`**
- `"when config subcommand provided should be recognized"`

**`TestCacheSubcommand`**
- `"when cache subcommand provided should be recognized"`

Extracted `setupRootCmd` helper following the same pattern as `setupConfigShow` in `config_test.go`. All tests went green immediately — dispatch was already correctly wired. No production code changes.

Code review (run during session) also fixed two issues in `cache.go` (hoisted single `bufio.NewReader`; replaced repeated `strings.ToLower` with pre-computed `clearKey` field) and extracted `writeGlobalConfig` helper in `config_test.go` to remove 6× boilerplate.

BUG-012 added to TODO: `CacheManager` writes warnings directly to `os.Stderr` — add injectable `writer io.Writer` field.

**TDD Remediation is COMPLETE.** All items (Pre-work, 0–10) done. New code may now be written.

Net change: +1 new test file (107 lines). All 200 tests pass.

### BUG-010: Capitalized Error Strings in config.go (2026-03-17)

- [x] Lowercase capitalized error strings in `internal/config/config.go` lines 280, 282 (completed 2026-03-17)

Fixed `staticcheck` ST1005 violation: lowercased `"Invalid ..."` → `"invalid ..."` in the `validationError` helper (2 lines). Updated 7 corresponding test assertions in `config_test.go` to assert correct lowercase behavior (tests had been written against the buggy capitalized strings). All 200 tests pass, `staticcheck` clean.
