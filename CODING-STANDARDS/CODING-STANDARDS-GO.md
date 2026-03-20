# Go Coding Standards

Go-specific mandatory standards for CALF development.

See [CODING-STANDARDS.md](CODING-STANDARDS.md) for shared standards that also apply.

---

## Dependency Management

```go
if _, err := exec.LookPath("tart"); err != nil {
    return fmt.Errorf("required command 'tart' not found in PATH")
}
```

---

## Stdlib Over Custom Implementations

Use `strings`, `filepath`, `slices`, etc. before writing custom helpers.

**Wrong:**
```go
func contains(s, substr string) bool {
    for i := 0; i <= len(s)-len(substr); i++ {
        if s[i:i+len(substr)] == substr { return true }
    }
    return false
}
```

**Correct:**
```go
strings.Contains(s, substr)
```

---

## GoDoc on All Exported Identifiers

All exported types, functions, constants, and variables must have GoDoc comments starting with the identifier name.

```go
// Config represents the top-level CAL configuration structure.
type Config struct { ... }

// LoadConfig loads configuration from global and per-VM paths.
// Returns error if files exist but cannot be read or parsed.
func LoadConfig(globalPath, vmPath string) (*Config, error) { ... }
```

---

## Test Style

Naming, Arrange/Act/Assert, and public-interface-only rules come from the `coops-tdd` skill. The rules below cover Go-specific patterns the skill does not address.

### No Table-Driven Loops

Never use `for _, tt := range tests { t.Run(tt.name, ...) }`. Each scenario must be its own `t.Run` block with its own Arrange/Act/Assert.

**Correct:**
```go
t.Run("when vm is running should return true", func(t *testing.T) {
    // Arrange
    mock := newMockCommandRunner()
    mock.addOutput("list --format json", `[{"name":"test-vm","state":"running"}]`)
    client := createTestClient(mock)
    // Act
    got := client.IsRunning("test-vm")
    // Assert
    if !got {
        t.Errorf("IsRunning() = false, want true")
    }
})
```

**Wrong:**
```go
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) { ... })
}
```

### Fresh Instance Per Test — No Shared State

Each subtest creates its own instance. Never share a constructed instance across subtests.

**Commands (cobra):** wrap a factory function in a setup helper:
```go
func setupRootCmd(t *testing.T, args ...string) (*cobra.Command, *bytes.Buffer, *bytes.Buffer) {
    t.Helper()
    out, errOut := &bytes.Buffer{}, &bytes.Buffer{}
    cmd := newRootCmd("test")
    cmd.SetOut(out); cmd.SetErr(errOut); cmd.SetArgs(args)
    return cmd, out, errOut
}
```

**Structs with injectable deps:** use functional options — never write unexported fields directly:
```go
// Correct
client := NewTartClient(
    WithLookPath(func(file string) (string, error) { ... }),
    WithRunCommand(func(args ...string) (string, error) { return mock.runCommand("tart", args...) }),
)

// Wrong
client.lookPath = func(...) { ... }  // unexported field
```

### `t.TempDir()` Over `os.MkdirTemp`

```go
homeDir := t.TempDir()                                            // Correct — auto-cleanup
tmpDir, _ := os.MkdirTemp("", "x"); defer os.RemoveAll(tmpDir)   // Wrong
```

### `t.Helper()` in Test Helpers

Call `t.Helper()` first in any function that calls `t.Fatal`/`t.Error` — failure output points to the calling test, not the helper.

### Mandatory Test Scenarios

Every change must cover: success path · all error return paths · edge/boundary conditions · component interactions (where applicable).

### Canonical Reference Files

| File | Demonstrates |
|------|-------------|
| `cmd/calf/config_test.go` | Factory via `newRootCmd()`, `t.Setenv` for env isolation |
| `cmd/calf/main_test.go` | Shared setup helper, fresh cmd per test |
| `cmd/calf/cache_test.go` | File-system assertions, confirm/decline flows |
| `internal/isolation/tart_test.go` (Clone tests) | Functional options, multi-option client construction |
| `internal/isolation/cache_test.go` (`TestClearCache`) | Per-subtest `t.TempDir()` + factory |

---

## Code Review Checklist — Go

In addition to shared checklist items:

- [ ] `go test ./...` and `staticcheck ./...` pass
- [ ] `go build ./...` succeeds
- [ ] Stdlib used over custom implementations
- [ ] All exported identifiers have GoDoc comments
- [ ] No table-driven `for _, tt := range tests` loops — each scenario is its own `t.Run` block
- [ ] Each subtest creates its own fresh instance — no shared state between subtests
- [ ] Temporary directories use `t.TempDir()` not `os.MkdirTemp` + `defer os.RemoveAll`
- [ ] Test helpers call `t.Helper()` as their first statement
- [ ] Injectable deps use functional options or exported constructors — no unexported field writes
