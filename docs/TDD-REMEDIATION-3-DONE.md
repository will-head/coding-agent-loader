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

## Item 4 — Move `ensureInstalled` to Public Methods (2026-03-18)

**File:** `internal/isolation/tart.go`

Removed `c.ensureInstalled()` from `runTartCommand` (lines 151–153). Added the call at the top of each public method that dispatches commands: `Clone`, `Set`, `RunWithCacheDirs`, `Stop`, `Delete`, `List`, `IP`. `Run` delegates to `RunWithCacheDirs` so no guard needed there. `Get`, `IsRunning`, `Exists`, `GetState` delegate to `List` so no guard needed there either.

Error wrapping for the `ensureInstalled` call in `RunWithCacheDirs` was kept consistent with all other methods (raw error, no wrapping).

**Completion criteria met:**
- [x] `ensureInstalled` removed from `runTartCommand`; added to `Clone`, `Set`, `RunWithCacheDirs`, `Stop`, `Delete`, `List`, `IP`
- [x] `go test ./...` passes (208 tests)
- [x] `staticcheck ./...` clean
