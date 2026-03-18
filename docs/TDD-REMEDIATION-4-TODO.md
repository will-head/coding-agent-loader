# TDD Remediation 4 — Arrange/Act/Assert Consistency and Table-Driven Test Elimination

> Items 1–5 and 8 complete as of 2026-03-18. See [TDD-REMEDIATION-4-DONE.md](TDD-REMEDIATION-4-DONE.md) for details.

---

## Item 9 — Rewrite Table-Driven Inner Loop in `TestClearCacheByType` (FAIL)

**File:** `internal/isolation/cache_test.go` (line 1334)

**Problem:** The `"when cache type is valid should clear that cache type"` subtest contains an inner `for _, cacheType := range testCases` loop over 4 values (`"homebrew"`, `"npm"`, `"go"`, `"git"`). Convention requires individual `t.Run("when...should...", ...)` blocks — not table-driven loops.

**Action:** Replace the inner `testCases` slice + `for` loop with 4 individual `t.Run` blocks, each with `// Arrange`, `// Act`, `// Assert` sections.

Run `go test ./internal/isolation/... -run TestClearCacheByType` — all subtests pass. Test count unchanged.

---

## Completion Criteria

- [ ] `TestClearCacheByType` inner loop in `cache_test.go` replaced with 4 individual `t.Run` blocks
- [ ] `go test ./...` passes (test count ≥ 208)
- [ ] `staticcheck ./...` passes with no warnings
