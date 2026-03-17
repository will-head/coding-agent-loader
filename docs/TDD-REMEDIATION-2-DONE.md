# TDD Remediation 2 — Completed Items

> **Source:** coops-tdd audit 2026-03-17

---

## Item 7 — Fix `vmName` cleanup in config_test.go (2026-03-17)

**File:** `cmd/calf/config_test.go`

Replaced `t.Cleanup(func() { vmName = "" })` with `t.Cleanup(func() { _ = configShowCmd.Flags().Set("vm", "") })`. The old form only zeroed the bound Go variable, leaving cobra's internal `changed` flag as `true` and risking cross-test contamination. The new form resets both the value and cobra's state atomically via the cobra API.

**Completion criteria met:**
- [x] `vmName = ""` cleanup investigated — reset IS needed (cobra does not auto-reset between executions)
- [x] Replaced with cobra API reset: `configShowCmd.Flags().Set("vm", "")`
- [x] `go test ./...` passes (207 tests)

---

## Item 8 — Fix fragile string slicing in config_test.go (2026-03-17)

**File:** `internal/config/config_test.go`

Replaced all 7 occurrences of `err.Error()[:len(expectedMsg)] != expectedMsg` with `!strings.HasPrefix(err.Error(), expectedMsg)`. Eliminates potential panics when error strings are shorter than expected, and uses stdlib over manual slicing.

**Completion criteria met:**
- [x] All `err.Error()[:len(x)]` slicing replaced with `strings.HasPrefix`
- [x] `go test ./...` passes (207 tests)
- [x] `go build` succeeds
