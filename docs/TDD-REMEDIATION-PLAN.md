# TDD Remediation Plan

This plan brings the entire codebase into compliance with `coops-tdd` as if the tests had been written first from the beginning. Every item follows the Red → Green → Refactor cycle. The plan is self-contained: an agent working through it sequentially should reach full compliance without further clarification.

---

## Guiding Principles

These govern every item in the plan. Read them before starting.

**Test behaviors, not implementations.**
The trigger for a new test is a new behavior or requirement — never a new method or class. Tests describe what the system does, not how it does it. Private functions and internal helpers are covered by the tests that exercise the behavior that caused them to exist.

**Test through the public API.**
Tests go through the stable, exported contract of each package. Do not access unexported fields or methods except where they are the explicit injection mechanism (e.g., `runCommand` on `TartClient` is the documented I/O injection point).

**Test doubles for I/O only.**
Replace subprocess calls, filesystem access, and stdin with in-memory fakes. Do not mock collaborating objects to isolate a class — let real code run end-to-end within the package.

**Naming convention.**
Top-level test functions: `TestBehaviorDescription` (Go PascalCase).
Sub-tests: `"when [condition] should [outcome]"` — descriptive, readable as a specification.

**Arrange / Act / Assert.**
Separate these three phases with blank lines. Use comments only where the separation is not visually obvious.

**Isolation.**
Every test is independent. Each test that needs a filesystem creates its own `t.TempDir()`. No shared mutable state between sub-tests. Sub-tests do not depend on execution order.

**The red phase must be genuine.**
Before making a test pass, verify it fails for the right reason — the behavior is absent, not that there is an import error or syntax problem. Never skip the red phase.

**Seeing os.Exit red.**
For tests against commands that currently call `os.Exit`, the red phase is acknowledged differently: write the test in terms of the correct behavior (command returns an error), note that the current production code cannot satisfy it without killing the test process, then make the production fix. The production fix is the direct consequence of what the test requires.

---

## Pre-Work: Fix go.mod Version

**No TDD cycle required — this is a configuration correction.**

The `go.mod` file specifies `go 1.25.6`, which is not a valid Go release. Before any other work:

1. Run `go version` to determine the actual installed Go version.
2. Update `go.mod` line 3 to match the installed version (e.g., `go 1.23.0`).
3. Run `go build ./...` and `go test ./...` to confirm nothing breaks.
4. Commit: `fix: correct go.mod version to match installed toolchain`.

---

## Item 0: Rename Existing Tests to Behavioral Convention

**Files:** `internal/config/config_test.go`, `internal/isolation/tart_test.go`, `internal/isolation/cache_test.go`

**Context:** All existing passing tests were written before coops-tdd was adopted. Their names follow a method-centric or action-centric style rather than the behavioral `when [condition] should [outcome]` specification format. This item renames them to match the convention — no behavioral changes, no production code changes.

**Before starting:** Run `go test ./...` and confirm all tests pass. Record the count. After renaming, the count must be identical and all must still pass.

### Naming rules to apply

**Top-level test functions** group related behaviors. In Go the convention is `TestXxx` — rename to group by **behavior area**, not by method name. Examples:
- `TestTartClient_Clone_Success` + `TestTartClient_Clone_Error` → `TestClone` (with sub-tests)
- `TestTartClient_Stop_Normal` + `TestTartClient_Stop_Force` → `TestStop`
- `TestVMState_String` → `TestVMStateString` (fine as-is, it describes the behavior of the `String()` method)

When a top-level function already contains `t.Run` sub-tests, the top-level name is just a grouping namespace — keep it concise. When a top-level function has no sub-tests, it should read as a behavior specification itself: `TestNewTartClientSetsDefaults`.

**Sub-tests** (strings passed to `t.Run`) must follow `"when [condition] should [outcome]"`. Every existing sub-test string must be rewritten to this format. Examples:
- `"missing config file uses defaults"` → `"when config file is missing should use default values"`
- `"valid config file loads correctly"` → `"when valid config file exists should load all fields correctly"`
- `"SetupHomebrewCache creates host cache directory"` → `"when home dir is available should create homebrew cache directory"`
- `"already exists returns false"` → `"when repo already cached should return false"`
- `"non-git directory is skipped"` → `"when directory is not a git repo should skip without error"`

**Top-level functions with no sub-tests** that test a single scenario should be renamed to describe the behavior:
- `TestTartClient_NewTartClient` → `TestNewTartClientSetsDefaults`
- `TestTartClient_Constants` → `TestCacheDirMountConstant` (or split into per-constant tests if there are multiple)
- `TestNewCacheManager` → `TestNewCacheManagerInitialisesFields`

### Tests to skip (being replaced in later items)

Do not rename these — they are deleted and replaced in Item 2:
- `TestTartClient_Run_Headless`
- `TestTartClient_Run_VNC_UsesExperimental`
- `TestTartClient_RunWithCacheDirs_AcceptsCacheDirs`

### Pass 1: Rename

Work through one file at a time:

1. Read the file in full.
2. For each top-level function and each sub-test string, apply the naming rules above.
3. After renaming within a file, run `go test ./...` — all tests must pass before moving to the next file.

Do not change any test logic, assertions, or setup — only the names. If a rename makes an existing test's intent ambiguous (the new name doesn't accurately describe what the test actually does), fix the name to match what the test actually asserts — still no logic changes.

### Pass 2: Arrange / Act / Assert structure

Once all files are renamed and green, make a second pass over the same three files to enforce AAA structure within each test body.

Rules:
- The three phases — Arrange (setup), Act (call the system under test), Assert (verify outcome) — must be visually separated by a blank line.
- Add a comment (`// Arrange`, `// Act`, `// Assert`) only where the separation is not immediately obvious to a reader. If blank lines make it clear, comments are unnecessary.
- Combine `// Act` and `// Assert` into `// Act / Assert` only when the act and assert are a single statement (e.g., `assert.NoError(t, doThing())`).
- Do not change test logic, add assertions, or remove setup — structure only.

After each file, run `go test ./...`. All tests must pass.

---

## Item 1: internal/config — GetDefaultConfigPath and GetVMConfigPath

**File:** `internal/config/config_test.go` (extend existing file)

**Context:** `GetDefaultConfigPath` and `GetVMConfigPath` are exported functions with no direct tests. They are exercised indirectly through `LoadConfig` but their specific contracts (path structure, error on missing home dir) are unverified.

**Before starting:** Read `internal/config/config.go` to confirm the exact signatures and home-dir dependency of both functions.

### Behaviors to specify

Add to the existing `config_test.go` file. Each behavior is a sub-test inside a descriptive top-level function.

**`TestGetDefaultConfigPath`**
- `"when home dir is available should return path ending in .calf/config.yaml"`
- `"when home dir is available should return an absolute path"`

**`TestGetVMConfigPath`**
- `"when vm name is provided should return path containing the vm name"`
- `"when vm name is provided should return path ending in config.yaml"`
- `"when vm name is provided should return an absolute path"`

**Notes on error cases:** If either function accepts a home-dir override or has an injectable dependency for `os.UserHomeDir`, add:
- `"when home dir is unavailable should return error"` for each function.

If the functions call `os.UserHomeDir` directly with no injection point, do not add a fake — note in a code comment that the error path is untestable without refactoring, and defer to a later refactor item. Do not add injection points speculatively; only add them when a failing test demands it.

### Red → Green → Refactor

Red: write all sub-tests. They should compile and run. Most will pass immediately (the code exists). Any that fail reveal a gap to fix.

Green: if any fail, make the minimal change to the production code to pass them.

Refactor: none expected here unless a structural improvement is obviously needed.

---

## Item 2: internal/isolation/tart.go — Run and RunWithCacheDirs

**Files:** `internal/isolation/tart.go` (production change), `internal/isolation/tart_test.go` (replace no-op tests, add behavior tests)

**Context:** `Run` and `RunWithCacheDirs` call `exec.Command` directly, bypassing the injectable `runCommand` field used by all other methods. The existing tests for these methods (`TestTartClient_Run_Headless`, `TestTartClient_Run_VNC_UsesExperimental`, `TestTartClient_RunWithCacheDirs_AcceptsCacheDirs`) are no-ops: they use `"echo"` as a stub binary, ignore all errors, and assert nothing about command construction.

**Before starting:** Read `internal/isolation/tart.go` in full to understand the current `Run`, `RunWithCacheDirs`, and `runTartCommand` implementations.

### Behaviors to specify

**Delete the three existing no-op tests** before writing the replacements:
- `TestTartClient_Run_Headless`
- `TestTartClient_Run_VNC_UsesExperimental`
- `TestTartClient_RunWithCacheDirs_AcceptsCacheDirs`

**Also delete the `cacheDirMount` constant test** (named `TestTartClient_Constants` or similar after Item 0 renaming). This test asserts the value of an unexported constant — an implementation detail. The behavior it intends to protect ("cache sharing dir is always included") is fully covered by the `TestRunWithCacheDirs` sub-test `"when called should always include the cache sharing directory"` written below. Once that behavioral test exists, the constant test is redundant and structure-sensitive.

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

### Red phase

Write all sub-tests using the existing `mockCommandRunner` pattern. The mock intercepts `c.runCommand` calls and captures the arguments.

**Why these are red:** `Run` and `RunWithCacheDirs` call `exec.Command` directly — the mock receives zero calls. All argument assertions fail.

### Production change required

Route `Run` and `RunWithCacheDirs` through `c.runCommand` instead of `exec.Command` directly. The mock will then capture the calls and argument assertions will pass.

Specifically:
- `Run` should build its argument slice (vm name, `--headless`, `--vnc-experimental` as appropriate) and call `c.runCommand("tart", "run", ...args)` or equivalent.
- `RunWithCacheDirs` should do the same, appending `--dir` flags for each cache directory plus the always-present cache sharing dir.

The method signatures, return types, and observable behavior (which flags appear under which conditions) must remain identical to the current implementation.

### Green → Refactor

After routing through `runCommand`, all sub-tests pass. Refactor only if duplication between `Run` and `RunWithCacheDirs` can be cleanly eliminated — `RunWithCacheDirs` may simply call `Run` with the cache dirs appended, or they may share an internal arg-builder. Run tests after each structural change.

---

## Item 3: internal/isolation/tart.go — ensureInstalled Homebrew Branch

**Files:** `internal/isolation/tart.go` (production change), `internal/isolation/tart_test.go` (add behaviors)

**Context:** `ensureInstalled` creates `bufio.NewReader(os.Stdin)` directly. This makes the user-prompt and Homebrew install branch untestable. The existing test (`TestTartClient_ensureInstalled_ChecksPath`) only covers the happy path where tart is already on the system path; it does this by setting `client.tartPath` directly.

**Before starting:** Read the full `ensureInstalled` implementation in `tart.go`, including the Homebrew branch.

### Behaviors to specify

**`TestEnsureInstalled`**
- `"when tart is found on path should set tart path without prompting"` — verify `client.tartPath` is set, no stdin consumed
- `"when tart is not found and user declines install should return error"`
- `"when tart is not found and user confirms install and brew succeeds should update tart path"`
- `"when tart is not found and user confirms install and brew fails should return error"`

### Red phase

Write all sub-tests. The declining/confirming sub-tests cannot be run without injectable stdin — they will hang waiting on `os.Stdin`. This confirms the red phase: the behavior is unspecifiable without a production change.

### Production change required

Add a `stdinReader io.Reader` field to `TartClient`. In `NewTartClient`, default it to `os.Stdin`. In `ensureInstalled`, read from `c.stdinReader` instead of creating a new `bufio.NewReader(os.Stdin)`.

If `ensureInstalled` calls `exec.Command` directly for the `brew install` step, route that call through `c.runCommand` as well, so the mock can simulate success or failure.

In tests, inject `strings.NewReader("y\n")` for confirm, `strings.NewReader("n\n")` for decline.

### Green → Refactor

All four sub-tests pass. The existing `TestTartClient_ensureInstalled_ChecksPath` should still pass — confirm it does. Refactor any duplication in argument setup.

---

## Item 4: internal/isolation/cache.go — CacheManager Injectable Constructor

**Files:** `internal/isolation/cache.go` (production change), `internal/isolation/cache_test.go` (update all construction sites)

**Context:** Tests in `cache_test.go` construct `CacheManager` using struct literals with unexported fields (`homeDir`, `cacheBaseDir`). This makes every test in the file structure-sensitive: renaming an internal field breaks tests even though no behavior changed. This violates Beck's "structure-insensitive" desideratum.

If written test-first, the test would have demanded a constructor that accepts the home directory and base directory as parameters. The current `NewCacheManager()` takes no arguments and derives paths from the environment.

**Before starting:** Read `internal/isolation/cache.go` to confirm `NewCacheManager`'s current signature and how `homeDir` and `cacheBaseDir` are set. Read `cache_test.go` to count all struct literal construction sites.

### Production change required

Add a constructor suitable for testing:

```go
// NewCacheManagerWithDirs creates a CacheManager rooted at the given home and
// cache base directories. Intended for use in tests.
func NewCacheManagerWithDirs(homeDir, cacheBaseDir string) *CacheManager
```

This constructor sets the same fields as the struct literals currently used in tests. The existing `NewCacheManager()` remains unchanged and continues to be used in production code.

### Test change required

Replace every struct literal `CacheManager{homeDir: ..., cacheBaseDir: ...}` in `cache_test.go` with a call to `NewCacheManagerWithDirs(...)`. No test logic changes — only the construction mechanism.

### Red → Green → Refactor

Red: write a new sub-test (or compile check) that constructs `CacheManager` via `NewCacheManagerWithDirs` — it fails because the constructor doesn't exist yet.

Green: add `NewCacheManagerWithDirs` to `cache.go`. Update all struct literal construction sites in `cache_test.go`.

Verify: run `go test ./...` — all tests pass. Then confirm that no struct literals with unexported fields remain in `cache_test.go`.

---

## Item 5: internal/isolation/cache.go — Fix Sub-Test Isolation in TestCacheManager_Clear

**File:** `internal/isolation/cache_test.go` (test-only change)

**Context:** `TestCacheManager_Clear` creates a single `CacheManager` instance and reuses it across many sub-tests. This violates the isolation requirement — sub-tests have implicit ordering dependencies and share mutable state.

**Before starting:** Read `TestCacheManager_Clear` in full to understand the current structure and which sub-tests share state.

### Change required

Restructure `TestCacheManager_Clear` so that every sub-test:
1. Creates its own `t.TempDir()` for home and cache base directories.
2. Constructs its own `CacheManager` pointing at those directories.
3. Sets up only the state it needs (calls `SetupXxxCache` explicitly within the sub-test).
4. Performs its act and assert steps.
5. Does not depend on any prior sub-test having run.

This is a test-only change. No production code changes. All sub-tests must pass after restructuring.

**Verify:** Run the sub-tests in isolation (e.g., with `-run TestCacheManager_Clear/subtest_name`) to confirm each one passes independently.

---

## Item 6: internal/isolation/cache.go — Eliminate GetXxxCacheInfo Duplication

**File:** `internal/isolation/cache.go` (production refactor), `internal/isolation/cache_test.go` (verify coverage before refactor)

**Context:** `GetHomebrewCacheInfo`, `GetNpmCacheInfo`, `GetGoCacheInfo`, and `GetGitCacheInfo` follow an identical ~30-line pattern: get cache path → resolve real path → choose path for size → stat → getDiskUsage → return CacheInfo. This violates CODING_STANDARDS.md ("Code duplication — Never leave copy-paste artifacts").

**Before starting:** Read all four `GetXxxCacheInfo` functions and confirm that `cache_test.go` has sub-tests covering each one's behavior (existence check, size reporting, unavailable cache returns zero). The existing tests must all be green before the refactor begins.

### Red → Green → Refactor sequence

This item is **refactor-only** — the behaviors already have test coverage. The sequence is:

1. Confirm all existing `GetXxxCacheInfo` tests pass (green baseline).
2. Extract a private `getCacheInfo(cachePath string) (CacheInfo, error)` helper that encapsulates the shared pattern.
3. Refactor each of the four public functions to call the helper.
4. Run all tests. They must remain green.

Do not add new tests for the private helper — it is covered by the tests for the four public functions.

If any test fails after the refactor, the refactor changed behavior. Revert the change, identify the behavioral difference, and fix it before re-extracting.

---

## Item 7: internal/isolation/cache.go — UpdateGitRepos Error Contract

**Files:** `internal/isolation/cache.go` (production change), `internal/isolation/cache_test.go` (add behaviors)

**Context:** `UpdateGitRepos` has the signature `(int, error)` but always returns `nil` for the error — individual repo fetch failures are written to stderr as warnings and swallowed. A caller checking the error return cannot distinguish "all repos updated" from "some failed". This is a broken contract: the signature promises error information that the implementation never delivers.

If written test-first, the test `"when a repo fails to update should return error"` would have driven the correct implementation from the start.

**Before starting:** Read the full `UpdateGitRepos` implementation and the existing tests for it in `cache_test.go`.

### Behaviors to specify

Add to `TestUpdateGitRepos` (the existing top-level function, renamed in Item 0):

- `"when all repos update successfully should return updated count and nil error"`
- `"when a repo fetch fails should return error"`
- `"when one repo fails should still attempt remaining repos"`
- `"when one repo fails should return count of successfully updated repos"`

### Red phase

The sub-tests `"when a repo fetch fails should return error"` and related failure cases are red immediately: the current implementation returns `nil` error regardless of what happens to individual repos.

### Production change required

When a repo fetch fails, accumulate the error rather than discarding it. On return, if any repos failed, return a descriptive error (e.g., `fmt.Errorf("failed to update %d of %d repos", failCount, total)`). The int return value should reflect the count of successfully updated repos, not the total attempted.

The warning to stderr may be retained alongside the error return — do not remove diagnostic output, only ensure the error is also surfaced to the caller.

### Green → Refactor

All sub-tests pass. Refactor the error accumulation if needed (e.g., collecting individual errors with `errors.Join`). Run `go test ./...`.

---

## Item 8: cmd/calf/config.go — runConfigShow

**Files:** `cmd/calf/config.go` (production change), `cmd/calf/config_test.go` (create new)

**Context:** `runConfigShow` calls `os.Exit(1)` directly in error paths instead of returning an error via `RunE`. This prevents the error paths from being tested without killing the test process. The function has no tests at all.

**Before starting:** Read `cmd/calf/config.go` in full. Note every `os.Exit` call and every code path.

### Behaviors to specify

Create `cmd/calf/config_test.go`. Use `package calf` (same package, to access command constructors).

Each test builds the command, attaches a `bytes.Buffer` via `cmd.SetOut`/`cmd.SetErr`, sets args via `cmd.SetArgs`, and calls `cmd.Execute()`. Use `t.TempDir()` with real config files on disk for filesystem I/O.

**`TestConfigShow`**
- `"when valid config file exists should output base image field"`
- `"when valid config file exists should output cpu count field"`
- `"when valid config file exists should output memory size field"`
- `"when config file is missing should output default values"`
- `"when vm name flag provided and vm config exists should output vm-specific values"`
- `"when vm name flag provided and vm config missing should fall back to global config"`
- `"when config file path is invalid should return error not exit process"`

### Red phase for the error test

The sub-test `"when config file path is invalid should return error not exit process"` is the critical one. Write it to call `Execute()` and assert `err != nil`. With `os.Exit` in place, the process exits instead of returning — the test cannot assert anything. This is the genuine red: the current code cannot satisfy the test's contract.

### Production change required

Convert `runConfigShow` from a `Run func(cmd *cobra.Command, args []string)` handler to a `RunE func(cmd *cobra.Command, args []string) error` handler. Replace every `os.Exit(1)` with `return fmt.Errorf(...)` (or `return err`). The command's successful output path remains unchanged.

### Green → Refactor

After converting to `RunE`, all sub-tests pass. Refactor only if there is obvious cleanup (e.g., repeated error-wrapping patterns).

---

## Item 9: cmd/calf/cache.go — runCacheClear

**Files:** `cmd/calf/cache.go` (production change), `cmd/calf/cache_test.go` (create new)

**Context:** `runCacheClear` (~135 lines) is entirely untested. It reads from `os.Stdin` directly in two places (global confirmation and per-type confirmation), making the interactive confirmation flows untestable.

**Before starting:** Read `cmd/calf/cache.go` in full. Map every code path: flag combinations (`--all`, `--force`, `--dry-run`, per-type flags), confirmation prompts, and output messages.

### Behaviors to specify

Create `cmd/calf/cache_test.go`. Use `t.TempDir()` for all cache directories. Inject stdin via `strings.NewReader` for interactive paths.

**`TestCacheClear`**
- `"when force flag set should clear without prompting for confirmation"`
- `"when dry run flag set should not delete any files"`
- `"when dry run flag set should output what would be cleared"`
- `"when all flag set and user confirms should clear all cache types"`
- `"when all flag set and user declines should abort without deleting files"`
- `"when homebrew type flag set should clear only homebrew cache"`
- `"when npm type flag set should clear only npm cache"`
- `"when go type flag set should clear only go cache"`
- `"when git type flag set should clear only git cache"`
- `"when all and force flags set should clear all types without prompting"`
- `"when all and dry run flags set should report all types without deleting"`

### Red phase

Write all sub-tests. Tests that exercise the confirmation prompt paths will hang on `os.Stdin` — confirming the red phase. Tests that use `--force` (no prompt) may pass or partially pass depending on how the cache path is constructed.

### Production change required

The `runCacheClear` function (or the command setup that creates it) needs injectable stdin. The cleanest approach:

- Add an `io.Reader` parameter to a new `newCacheCmd(stdin io.Reader) *cobra.Command` constructor.
- The command's run function closes over `stdin` instead of reading from `os.Stdin`.
- In `main.go`, call `newCacheCmd(os.Stdin)`.
- In tests, call `newCacheCmd(strings.NewReader("y\n"))` or `newCacheCmd(strings.NewReader("n\n"))`.

If `runCacheClear` already uses a `CacheManager`, ensure the test can point the manager at `t.TempDir()` directories — either via a constructor parameter or by setting the manager's fields directly (acceptable since `CacheManager` is in the same module).

### Green → Refactor

All sub-tests pass. Refactor the large `runCacheClear` function if sub-functions would make it more readable — but only if the tests remain green and no new behavior is introduced.

---

## Item 10: cmd/calf/main.go — Root Command Integration Tests

**Files:** `cmd/calf/main_test.go` (create new)

**Context:** The root command setup and top-level dispatch have no tests. This is the outermost public surface of the CLI.

**Before starting:** Read `cmd/calf/main.go` to understand `rootCmd` setup, registered sub-commands, and any persistent flags.

### Behaviors to specify

Create `cmd/calf/main_test.go`. Tests use `cmd.SetArgs`, `cmd.SetOut`, `cmd.SetErr`, and `cmd.Execute()`. No real config files needed for these dispatch-level tests.

**`TestRootCommand`**
- `"when no args provided should print usage information"`
- `"when help flag provided should print help text"`
- `"when unknown subcommand provided should return error"`

**`TestConfigSubcommand`**
- `"when config subcommand provided should be recognized"`

**`TestCacheSubcommand`**
- `"when cache subcommand provided should be recognized"`

Note: these are dispatch tests only — they verify routing, not the full behavior of each sub-command (which is covered in Items 6 and 7). Do not duplicate the sub-command behavior tests here.

### Red → Green → Refactor

Most of these will go green immediately since the commands are already wired up. Any that fail reveal a dispatch problem to fix. No production changes are expected here.

---

## Completion Checklist

Work through items in order. Do not skip items or reorder them — later items depend on production changes from earlier items (e.g., Item 10 depends on the `newCacheCmd` constructor from Item 9).

After all items are complete, run the full test suite:

```
go test ./...
```

All tests must pass. Then run:

```
staticcheck ./...
```

All static analysis checks must pass with no issues.

Finally, verify the test suite satisfies Kent Beck's Test Desiderata:
- **Isolated**: each test can run alone without other tests having run first
- **Deterministic**: `go test ./...` produces the same result on every run
- **Fast**: the full suite completes in under 30 seconds (no real subprocess calls, no sleep)
- **Behavioral**: tests describe what the system does, not how
- **Structure-insensitive**: renaming an unexported function does not break any test
- **Readable**: each sub-test name reads as a specification of the system's behavior

If any item in this checklist fails, fix it before marking the remediation complete.
