# TDD Remediation 3 — Completed Items

> **Source:** coops-tdd audit 2026-03-17

---

## Item 1 — Rename `TestVMStateString` Subtests (2026-03-17)

**File:** `internal/isolation/tart_test.go`

Replaced table-driven `TestVMStateString` with individual `t.Run("when...should...", ...)` blocks, each with explicit `// Arrange / // Act / // Assert` sections. Added separate `want` field; removed the dual-purpose `name` field that doubled as expected string.

Also applied as part of code review: replaced hand-rolled `equalStringSlices` and `sliceContains` helpers with stdlib `slices.Equal` and `slices.Contains` (Go 1.21+); removed both helpers.

**Completion criteria met:**
- [x] `TestVMStateString` subtests renamed to `when...should...` with separate `want` field
- [x] `go test ./...` passes (203 tests)
- [x] `staticcheck ./...` clean

---

## Item 2 — Wrap `TestCloneWhenTart*` in `t.Run` Subtests (2026-03-17)

**File:** `internal/isolation/tart_test.go`

Wrapped each of the five `TestCloneWhenTart*` flat function bodies in a `t.Run("when...should...", ...)` subtest block. No logic changes — purely structural.

**Completion criteria met:**
- [x] All five `TestCloneWhenTart*` functions wrap bodies in `t.Run("when...should...", ...)`
- [x] `go test ./internal/isolation/... -run TestCloneWhen` — all 5 pass

---

## Item 3 — `cacheDirMount` Constant Coupling (2026-03-18)

**File:** `internal/isolation/tart_test.go`

**Assessment:** No code change made. Assessed two approaches:

1. **Export `cacheDirMount` as `CacheDirMount`** (original plan) — moves the coupling from an unexported to an exported name; tests still reference an implementation detail. Rejected.
2. **Use literal string** `"--dir=tart-cache:~/.tart/cache:ro"` in tests — decouples tests from the constant name, but code review flagged as stringly-typed duplication that could silently diverge from the constant.

**Decision:** Retain existing `fmt.Sprintf("--dir=%s", cacheDirMount)` pattern. The coupling is temporary — Items 5–6 introduce functional options (`WithTartPath`, `WithRunCommand`) which allow `createTestClient` to be rewritten without referencing any unexported fields or constants. At that point this test will be rewritten and the constant reference removed naturally.

**Completion criteria met:**
- [x] Item assessed; no code change warranted; coupling addressed by Items 5–6

---

## Item 5 — Add Functional Options to `TartClient` (2026-03-18)

**File:** `internal/isolation/tart.go`

Added `TartClientOption` exported type and 7 option functions: `WithRunCommand`, `WithPollInterval`, `WithPollTimeout`, `WithTartPath`, `WithLookPath`, `WithStdinReader`, `WithBrewRunner`. Updated `NewTartClient` to accept `...TartClientOption` — variadic, fully backwards-compatible. Options are applied after defaults so they correctly override any default value.

**Completion criteria met:**
- [x] `TartClientOption` type exported; 7 option functions added
- [x] `NewTartClient` accepts `...TartClientOption` (backwards-compatible)
- [x] `go build ./...` and `go test ./...` pass (208 tests)
- [x] `staticcheck ./...` clean

---

## Item 6 — Rewrite `createTestClient` Using Options (2026-03-18)

**File:** `internal/isolation/tart_test.go`

Replaced direct unexported field assignments (`tartPath`, `pollInterval`, `pollTimeout`, `runCommand`) with a single `NewTartClient(...)` call using functional options. Added variadic `extra ...TartClientOption` parameter so callers can supply per-test overrides (e.g. `WithPollTimeout`) without post-construction field mutation. GoDoc comment updated to document the override semantics.

**Completion criteria met:**
- [x] `createTestClient` uses `WithTartPath`, `WithPollInterval`, `WithPollTimeout`, `WithRunCommand` — no unexported field writes
- [x] `go test ./...` passes (208 tests)
- [x] `staticcheck ./...` clean

---

## Item 7 — Fix `client.pollTimeout` Direct Write in `TestIP` (2026-03-18)

**File:** `internal/isolation/tart_test.go`

Replaced `client.pollTimeout = 50 * time.Millisecond` (direct unexported field write after `createTestClient`) with `createTestClient(mock, WithPollTimeout(50*time.Millisecond))`. The variadic `extra` parameter added in Item 6 enables this one-liner without duplicating the full `NewTartClient(...)` call.

**Completion criteria met:**
- [x] `client.pollTimeout = ...` at line 281 replaced with `WithPollTimeout` via `createTestClient` extra arg
- [x] `go test ./internal/isolation/... -run TestIP` — both subtests pass

---

## Item 8 — Rewrite `TestCloneWhenTart*` Using Options (2026-03-18)

**File:** `internal/isolation/tart_test.go`

Replaced direct unexported field assignments (`lookPath`, `stdinReader`, `runBrewCommand`, `runCommand`) in all five `TestCloneWhenTart*` functions with `NewTartClient(...)` calls using functional options (`WithLookPath`, `WithStdinReader`, `WithBrewRunner`, `WithRunCommand`). Applied inside the `t.Run("when...should...", ...)` wrappers added in Item 2. Commit `4b0ebe9`.

**Completion criteria met:**
- [x] `TestCloneWhenTart*` uses `WithLookPath`, `WithStdinReader`, `WithBrewRunner`, `WithRunCommand` — no unexported field writes
- [x] `go test ./internal/isolation/... -run TestCloneWhen` — all 5 pass

---

## Item 9 — Delete `makeInstallingRunCommand` (2026-03-18)

**File:** `internal/isolation/tart_test.go`

Deleted `makeInstallingRunCommand` helper (lines 757–766) which called `client.ensureInstalled()` — an unexported method. After Item 4 moved `ensureInstalled` into public methods and Item 8 rewrote all callers, no test referenced `makeInstallingRunCommand`. Verified no remaining references. Commit `4b0ebe9`.

**Completion criteria met:**
- [x] `makeInstallingRunCommand` deleted; no references remain
- [x] `go build ./...` succeeds; `go test ./...` passes (208 tests)

---

## Item 10 — `newRootCmd()` / `newConfigCmd()` Factories (2026-03-18)

**Files:** `cmd/calf/main.go`, `cmd/calf/config.go`, `cmd/calf/main_test.go`, `cmd/calf/config_test.go`

Added `newConfigCmd()` factory to `config.go`; removed global `configCmd`, `configShowCmd`, `vmName` vars and `init()`. Added `newRootCmd()` factory to `main.go` wiring `newConfigCmd()` and `newCacheCmd()`. Rewrote `config_test.go` `setupConfigShow` to return a fresh `cmd` via `newRootCmd()` — no `configShowCmd` reference, no flag cleanup needed. Rewrote all `main_test.go` tests to use `newRootCmd()` per test — no shared `rootCmd` mutations, no `t.Cleanup` teardown. Commit `2c17cf4`.

**Completion criteria met:**
- [x] `newConfigCmd()` factory added to `config.go`; global `configCmd`, `configShowCmd`, `vmName` vars and `init()` removed
- [x] `newRootCmd()` factory added to `main.go`; wires `newConfigCmd()` and `newCacheCmd()`
- [x] `config_test.go`: `setupConfigShow` returns fresh `cmd` via `newRootCmd()`; no `configShowCmd` reference; no flag cleanup
- [x] `main_test.go`: all tests use `newRootCmd()` per test; no shared `rootCmd` mutations
- [x] `go test ./...` passes (208 tests)
- [x] `go test -count=2 ./cmd/calf/...` passes (proves no shared state leakage)
- [x] `staticcheck ./...` passes with no warnings

---

## Item 4 — Move `ensureInstalled` to Public Methods (2026-03-18)

**File:** `internal/isolation/tart.go`

Removed `c.ensureInstalled()` from `runTartCommand` (lines 151–153). Added the call at the top of each public method that dispatches commands: `Clone`, `Set`, `RunWithCacheDirs`, `Stop`, `Delete`, `List`, `IP`. `Run` delegates to `RunWithCacheDirs` so no guard needed there. `Get`, `IsRunning`, `Exists`, `GetState` delegate to `List` so no guard needed there either.

Error wrapping for the `ensureInstalled` call in `RunWithCacheDirs` was kept consistent with all other methods (raw error, no wrapping).

**Completion criteria met:**
- [x] `ensureInstalled` removed from `runTartCommand`; added to `Clone`, `Set`, `RunWithCacheDirs`, `Stop`, `Delete`, `List`, `IP`
- [x] `go test ./...` passes (208 tests)
- [x] `staticcheck ./...` clean
