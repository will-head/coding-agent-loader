# TDD Remediation 3 — Private Interface Access and Test Isolation

> **Source:** coops-tdd audit 2026-03-17
>
> **Scope:** Fix all FAIL and WARN violations from the Remediation 3 audit. Items ordered smallest-to-largest: naming/style fixes first (no logic change), then test-only changes, then production+test changes requiring design work.
>
> **Reference assessment:** coops-tdd audit report, session history 2026-03-17
>
> **Workflow:** Always follow the `coops-tdd` skill throughout this remediation. Every item — even pure test restructuring — must go through the coops-tdd process before changes are made.
>
> **Style reference:** Use `cmd/calf/config_test.go` and `internal/isolation/tart_test.go` as style guides. Each scenario is an individual `t.Run("when...should...", ...)` block with explicit `// Arrange`, `// Act`, `// Assert` sections — not table-driven loops.

---

## Summary of Violations

| # | File | Severity | Violation |
|---|------|----------|-----------|
| 1 | `internal/isolation/tart_test.go` | WARN | `TestVMStateString` subtests named after data values, not behaviour |
| 2 | `internal/isolation/tart_test.go` | WARN | `TestCloneWhenTart*` flat functions have no `t.Run("when...should...")` |
| 3 | `internal/isolation/tart_test.go` | FAIL | `createTestClient` writes unexported fields: `tartPath`, `pollInterval`, `pollTimeout`, `runCommand` |
| 4 | `internal/isolation/tart_test.go` | FAIL | `TestCloneWhenTart*` tests write unexported fields: `lookPath`, `stdinReader`, `runBrewCommand`, `runCommand` |
| 5 | `internal/isolation/tart_test.go` | FAIL | `client.pollTimeout` written directly at line 281 in `TestIP` |
| 6 | `internal/isolation/tart_test.go` | FAIL | `makeInstallingRunCommand` calls unexported method `client.ensureInstalled()` |
| 7 | `internal/isolation/tart_test.go` | WARN | `cacheDirMount` unexported constant referenced directly in tests |
| 8 | `cmd/calf/config_test.go` | FAIL | `configShowCmd` unexported var accessed in `setupConfigShow` for flag reset |
| 9 | `cmd/calf/main_test.go` | WARN | Shared mutable `rootCmd` — tests share global state |
| 10 | `internal/config/config_test.go` | WARN | Error message prefix assertions couple tests to internal string wording |

Items 8 and 9 share the same root cause and fix (Item 10 below).
Item 10 from the summary is assessed as a non-issue (see Item 11 below).

---

## Item 2 — Wrap `TestCloneWhenTart*` Flat Tests in `t.Run` (WARN)

**File:** `internal/isolation/tart_test.go` (lines 630–755)

**Problem:** Five top-level functions contain all assertions at the top level without a `t.Run("when...should...", ...)` subtest. Project convention is to use `t.Run` even for single-scenario tests.

**Action:** Wrap each function body in `t.Run("when [condition] should [outcome]", func(t *testing.T) {...})`. No logic changes — only restructuring.

| Function | `t.Run` name to use |
|----------|---------------------|
| `TestCloneWhenTartIsInstalled` | `"when tart is installed should dispatch clone command"` |
| `TestCloneWhenTartIsNotInstalledAndUserDeclines` | `"when tart is not installed and user declines should return cancelled error"` |
| `TestCloneWhenTartIsNotInstalledAndUserConfirmsAndBrewSucceeds` | `"when tart is not installed and user confirms and brew succeeds should clone successfully"` |
| `TestCloneWhenTartIsNotInstalledAndBrewFails` | `"when tart is not installed and brew fails should return install error"` |
| `TestCloneWhenBrewIsNotAvailableAndTartNotFound` | `"when neither tart nor brew is available should return error without prompting"` |

Example of the wrapping pattern — apply to all five:

```go
func TestCloneWhenTartIsInstalled(t *testing.T) {
    t.Run("when tart is installed should dispatch clone command", func(t *testing.T) {
        // existing body here, unchanged
    })
}
```

Run `go test ./internal/isolation/... -run TestCloneWhen` — all 5 pass.

---

## Item 5 — Add Functional Options to `TartClient` (FAIL — production change, requires Item 4)

**File:** `internal/isolation/tart.go`

**Problem:** All `TartClient` fields are unexported. Tests write them directly to configure the client. This violates "test public interface only."

**Action:** Add a `TartClientOption` exported type and option functions. Update `NewTartClient` to accept options.

**Add after the `commandRunner` type declaration (after line 60):**

```go
// TartClientOption configures a TartClient.
type TartClientOption func(*TartClient)

// WithRunCommand overrides the command runner used to dispatch tart commands.
// Intended for use in tests.
func WithRunCommand(fn commandRunner) TartClientOption {
    return func(c *TartClient) { c.runCommand = fn }
}

// WithPollInterval overrides the IP polling interval.
// Intended for use in tests.
func WithPollInterval(d time.Duration) TartClientOption {
    return func(c *TartClient) { c.pollInterval = d }
}

// WithPollTimeout overrides the IP polling timeout.
// Intended for use in tests.
func WithPollTimeout(d time.Duration) TartClientOption {
    return func(c *TartClient) { c.pollTimeout = d }
}

// WithTartPath sets the tart binary path, skipping ensureInstalled discovery.
// Intended for use in tests.
func WithTartPath(path string) TartClientOption {
    return func(c *TartClient) { c.tartPath = path }
}

// WithLookPath overrides the exec.LookPath function used to locate binaries.
// Intended for use in tests.
func WithLookPath(fn func(string) (string, error)) TartClientOption {
    return func(c *TartClient) { c.lookPath = fn }
}

// WithStdinReader overrides the stdin reader for install confirmation prompts.
// Intended for use in tests.
func WithStdinReader(r io.Reader) TartClientOption {
    return func(c *TartClient) { c.stdinReader = r }
}

// WithBrewRunner overrides the brew command runner used during tart installation.
// Intended for use in tests.
func WithBrewRunner(fn commandRunner) TartClientOption {
    return func(c *TartClient) { c.runBrewCommand = fn }
}
```

**Update `NewTartClient` signature** to accept variadic options, applied after defaults:

```go
// NewTartClient creates a new TartClient with optional configuration overrides.
func NewTartClient(opts ...TartClientOption) *TartClient {
    client := &TartClient{
        installPrompt: TartInstallPrompt,
        outputWriter:  os.Stdout,
        errorWriter:   os.Stderr,
        pollInterval:  defaultPollInterval,
        pollTimeout:   defaultPollTimeout,
        stdinReader:   os.Stdin,
        lookPath:      exec.LookPath,
    }
    client.runCommand = client.runTartCommand
    client.runBrewCommand = func(args ...string) (string, error) {
        brewPath, err := client.lookPath("brew")
        if err != nil {
            return "", fmt.Errorf("brew not found: %w", err)
        }
        cmd := exec.Command(brewPath, args...)
        cmd.Stdout = client.outputWriter
        cmd.Stderr = client.errorWriter
        if err := cmd.Run(); err != nil {
            return "", err
        }
        return "", nil
    }
    for _, opt := range opts {
        opt(client)
    }
    return client
}
```

The variadic signature is backwards-compatible — all existing `NewTartClient()` calls continue to compile without changes.

Run `go build ./...` and `go test ./...` — all pass. `staticcheck ./...` — clean.

---

## Item 6 — Rewrite `createTestClient` Using Options (FAIL — test change, requires Item 5)

**File:** `internal/isolation/tart_test.go` (lines 48–61)

**Problem:** `createTestClient` writes four unexported fields directly: `tartPath`, `pollInterval`, `pollTimeout`, `runCommand`.

**Action:** Replace with options-based construction using the options added in Item 5.

**Replace `createTestClient` with:**

```go
// createTestClient creates a TartClient configured for testing using public options.
func createTestClient(mock *mockCommandRunner) *TartClient {
    return NewTartClient(
        WithTartPath("/usr/local/bin/tart"),
        WithPollInterval(10*time.Millisecond),
        WithPollTimeout(100*time.Millisecond),
        WithRunCommand(func(args ...string) (string, error) {
            return mock.runCommand("tart", args...)
        }),
    )
}
```

Run `go test ./internal/isolation/... -run TestClone|TestSet|TestStop|TestDelete|TestList|TestIP|TestGet|TestIsRunning|TestExists|TestGetState|TestRun` — all pass.

---

## Item 7 — Fix `client.pollTimeout` Direct Write at Line 281 (FAIL — test change, requires Item 5)

**File:** `internal/isolation/tart_test.go` (lines 279–281)

**Problem:** `client.pollTimeout = 50 * time.Millisecond` — direct unexported field write in the test body after `createTestClient` has already set it to 100ms. The test needs a shorter timeout to make the timeout case fast.

**Action:** Build the client directly with `WithPollTimeout(50*time.Millisecond)` instead of using `createTestClient` and then overriding.

**Replace lines 279–281:**
```go
// Before:
client := createTestClient(mock)
client.pollTimeout = 50 * time.Millisecond

// After:
client := NewTartClient(
    WithTartPath("/usr/local/bin/tart"),
    WithPollInterval(10*time.Millisecond),
    WithPollTimeout(50*time.Millisecond),
    WithRunCommand(func(args ...string) (string, error) {
        return mock.runCommand("tart", args...)
    }),
)
```

Run `go test ./internal/isolation/... -run TestIP` — both subtests pass.

---

## Item 8 — Rewrite `TestCloneWhenTart*` Using Options (FAIL — test change, requires Items 4 and 5)

**File:** `internal/isolation/tart_test.go` (lines 630–755)

**Problem:** Five `TestCloneWhenTart*` functions write unexported fields `lookPath`, `stdinReader`, `runBrewCommand`, `runCommand` directly.

**After Item 4**, `Clone` calls `ensureInstalled` directly before dispatching `runCommand`. This means tests only need to inject `lookPath`, `stdinReader`, `runBrewCommand` (which `ensureInstalled` uses) and `runCommand` (for the actual tart dispatch). No test helper needs to call `ensureInstalled` manually.

**Replace direct field assignments in all five functions with options. Apply changes inside the `t.Run` wrapper added in Item 2.**

`TestCloneWhenTartIsInstalled`:
```go
// Arrange
mock := newMockCommandRunner()
mock.addOutput("clone test-image test-vm", "")
client := NewTartClient(
    WithLookPath(func(file string) (string, error) {
        if file == "tart" {
            return "/usr/local/bin/tart", nil
        }
        return "", fmt.Errorf("not found")
    }),
    WithRunCommand(func(args ...string) (string, error) {
        return mock.runCommand("tart", args...)
    }),
)
```

`TestCloneWhenTartIsNotInstalledAndUserDeclines`:
```go
// Arrange
client := NewTartClient(
    WithLookPath(func(file string) (string, error) {
        if file == "brew" {
            return "/usr/local/bin/brew", nil
        }
        return "", fmt.Errorf("not found")
    }),
    WithStdinReader(strings.NewReader("n\n")),
)
```

`TestCloneWhenTartIsNotInstalledAndUserConfirmsAndBrewSucceeds`:
```go
// Arrange
lookPathCalls := 0
mock := newMockCommandRunner()
mock.addOutput("clone test-image test-vm", "")
client := NewTartClient(
    WithLookPath(func(file string) (string, error) {
        if file == "brew" {
            return "/usr/local/bin/brew", nil
        }
        if file == "tart" {
            lookPathCalls++
            if lookPathCalls > 1 {
                return "/usr/local/bin/tart", nil
            }
        }
        return "", fmt.Errorf("not found")
    }),
    WithStdinReader(strings.NewReader("y\n")),
    WithBrewRunner(func(args ...string) (string, error) {
        return "", nil
    }),
    WithRunCommand(func(args ...string) (string, error) {
        return mock.runCommand("tart", args...)
    }),
)
```

`TestCloneWhenTartIsNotInstalledAndBrewFails`:
```go
// Arrange
client := NewTartClient(
    WithLookPath(func(file string) (string, error) {
        if file == "brew" {
            return "/usr/local/bin/brew", nil
        }
        return "", fmt.Errorf("not found")
    }),
    WithStdinReader(strings.NewReader("y\n")),
    WithBrewRunner(func(args ...string) (string, error) {
        return "", fmt.Errorf("brew install failed")
    }),
)
```

`TestCloneWhenBrewIsNotAvailableAndTartNotFound`:
```go
// Arrange
client := NewTartClient(
    WithLookPath(func(file string) (string, error) {
        return "", fmt.Errorf("not found")
    }),
    WithStdinReader(strings.NewReader("")),
)
```

**Note on `WithRunCommand` in success-path tests:** After `ensureInstalled` sets `tartPath`, `Clone` calls `c.runCommand("clone", ...)`. Without `WithRunCommand`, this would call `runTartCommand` which tries to exec the real tart binary. `WithRunCommand` injects a mock for the actual tart dispatch — this is legitimate: the tart binary is external slow I/O and should be mocked.

Run `go test ./internal/isolation/... -run TestCloneWhen` — all 5 pass.

---

## Item 9 — Delete `makeInstallingRunCommand` (FAIL — test change, requires Items 4 and 8)

**File:** `internal/isolation/tart_test.go` (lines 757–766)

**Problem:** `makeInstallingRunCommand` calls `client.ensureInstalled()` — an unexported method. After Item 4, `Clone` calls `ensureInstalled` directly. After Item 8, no test uses `makeInstallingRunCommand`.

**Action:** Delete the `makeInstallingRunCommand` function (lines 757–766). Verify no remaining references with `grep -n makeInstallingRunCommand internal/isolation/tart_test.go`.

Run `go build ./...` — no compilation errors. Run `go test ./...` — all tests pass.

---

## Item 10 — Add `newRootCmd()` Factory to Fix Shared Command State (FAIL + WARN)

**Files:** `cmd/calf/main.go`, `cmd/calf/config.go`, `cmd/calf/main_test.go`, `cmd/calf/config_test.go`

**Problem:**
- `configShowCmd` (unexported global var) is accessed in `config_test.go` line 30 to reset cobra flag state between tests — a Rule 2 violation.
- `rootCmd` (unexported global var) is shared across all tests in `main_test.go` — tests mutate shared state.

**Root cause:** Command tree wired in `init()` produces global cobra instances. Subsequent test `Execute()` calls share the same instance, requiring cleanup hacks.

**Design:** Introduce `newConfigCmd()` and `newRootCmd()` factory functions. Tests call `newRootCmd()` to get a fresh, fully wired command tree per test — no shared state, no cleanup needed. This is the same pattern already used by `newCacheCmd()` in `cache.go`.

### Step 1 — Refactor `config.go`

Remove global vars `configCmd`, `configShowCmd`, `vmName` and the `init()` function. `runConfigShow` takes `vmName` as a parameter instead of closing over a package-level var.

```go
package main

import (
    "fmt"

    "github.com/spf13/cobra"
    "github.com/will-head/coding-agent-launcher/internal/config"
)

// newConfigCmd creates and wires the "calf config" subcommand tree.
func newConfigCmd() *cobra.Command {
    configCmd := &cobra.Command{
        Use:   "config",
        Short: "Manage CALF configuration",
    }

    configShowCmd := &cobra.Command{
        Use:   "show",
        Short: "Display effective configuration",
    }

    var vmName string
    configShowCmd.Flags().StringVarP(&vmName, "vm", "v", "", "VM name to show config for")
    configShowCmd.RunE = func(cmd *cobra.Command, args []string) error {
        return runConfigShow(cmd, vmName)
    }

    configCmd.AddCommand(configShowCmd)
    return configCmd
}

func runConfigShow(cmd *cobra.Command, vmName string) error {
    globalConfigPath, err := config.GetDefaultConfigPath()
    if err != nil {
        return fmt.Errorf("getting default config path: %w", err)
    }

    var vmConfigPath string
    if vmName != "" {
        vmConfigPath, err = config.GetVMConfigPath(vmName)
        if err != nil {
            return fmt.Errorf("getting VM config path: %w", err)
        }
    }

    cfg, err := config.LoadConfig(globalConfigPath, vmConfigPath)
    if err != nil {
        return fmt.Errorf("loading configuration: %w", err)
    }

    out := cmd.OutOrStdout()

    fmt.Fprintln(out, "CALF Configuration")
    fmt.Fprintln(out, "=================")
    fmt.Fprintln(out)
    fmt.Fprintln(out, "VM Defaults:")
    fmt.Fprintf(out, "  CPU: %d cores\n", cfg.Isolation.Defaults.VM.CPU)
    fmt.Fprintf(out, "  Memory: %d MB\n", cfg.Isolation.Defaults.VM.Memory)
    fmt.Fprintf(out, "  Disk Size: %d GB\n", cfg.Isolation.Defaults.VM.DiskSize)
    fmt.Fprintf(out, "  Base Image: %s\n", cfg.Isolation.Defaults.VM.BaseImage)
    fmt.Fprintln(out)
    fmt.Fprintln(out, "GitHub:")
    fmt.Fprintf(out, "  Default Branch Prefix: %s\n", cfg.Isolation.Defaults.GitHub.DefaultBranchPrefix)
    fmt.Fprintln(out)
    fmt.Fprintln(out, "Output:")
    fmt.Fprintf(out, "  Sync Directory: %s\n", cfg.Isolation.Defaults.Output.SyncDir)
    fmt.Fprintln(out)
    fmt.Fprintln(out, "Proxy:")
    fmt.Fprintf(out, "  Mode: %s\n", cfg.Isolation.Defaults.Proxy.Mode)
    fmt.Fprintln(out)

    if vmName != "" {
        fmt.Fprintf(out, "(Showing config for VM: %s)\n", vmName)
    } else {
        fmt.Fprintln(out, "(Showing global config)")
    }

    return nil
}
```

### Step 2 — Refactor `main.go`

Add `newRootCmd()`. Keep a package-level `rootCmd` for `main()`:

```go
package main

import (
    "fmt"
    "os"

    "github.com/spf13/cobra"
)

var Version = "dev"

// newRootCmd creates a fully wired root command tree.
func newRootCmd() *cobra.Command {
    root := &cobra.Command{
        Use:   "calf",
        Short: "CALF - Coding Agent Loader Foundation",
        Long: `CALF (Coding Agent Loader Foundation) - VM-based sandbox for AI coding agents.

CALF provides isolated macOS VMs (via Tart) for running AI coding agents safely,
with automated setup, snapshot management, and GitHub workflow integration.`,
        Version: Version,
    }
    root.AddCommand(newConfigCmd())
    root.AddCommand(newCacheCmd(os.UserHomeDir, os.Stdin))
    return root
}

var rootCmd = newRootCmd()

func main() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

**Note:** `newCacheCmd` already exists in `cache.go`. Check its current signature by reading `cache.go` and ensure the arguments passed here (`os.UserHomeDir`, `os.Stdin`) match. Adjust if the signature differs.

### Step 3 — Rewrite `config_test.go`

Remove the `configShowCmd.Flags().Set("vm", "")` cleanup. Use `newRootCmd()` for a fresh instance per test. Update `setupConfigShow` to return the fresh `cmd` and have tests call `cmd.Execute()`:

```go
// setupConfigShow wires a fresh root command for "calf config show [extraArgs...]" in an
// isolated temp HOME. Returns the home dir, captured stdout/stderr buffers, and the command.
func setupConfigShow(t *testing.T, extraArgs ...string) (home string, out, errOut *bytes.Buffer, cmd *cobra.Command) {
    t.Helper()
    home = t.TempDir()
    t.Setenv("HOME", home)
    out = &bytes.Buffer{}
    errOut = &bytes.Buffer{}
    cmd = newRootCmd()
    cmd.SetOut(out)
    cmd.SetErr(errOut)
    cmd.SetArgs(append([]string{"config", "show"}, extraArgs...))
    return home, out, errOut, cmd
}
```

Update the import to add `"github.com/spf13/cobra"` if not already present.

Update all test bodies that call `setupConfigShow` to use the returned `cmd`:

```go
// Before:
_, out, _ := setupConfigShow(t)
err := rootCmd.Execute()

// After:
_, out, _, cmd := setupConfigShow(t)
err := cmd.Execute()
```

Apply this pattern to all 7 subtests in `TestConfigShow`.

Remove `configShowCmd` from the import scope — it is no longer referenced.

### Step 4 — Rewrite `main_test.go`

Replace all `rootCmd.Execute()` calls with a fresh `newRootCmd()` instance per test. Remove shared `t.Cleanup` blocks that reset `rootCmd` output writers.

Pattern to apply throughout `main_test.go`:

```go
// Before:
out := &bytes.Buffer{}
rootCmd.SetOut(out)
rootCmd.SetErr(&bytes.Buffer{})
t.Cleanup(func() {
    rootCmd.SetOut(nil)
    rootCmd.SetErr(nil)
})
rootCmd.SetArgs(...)
t.Cleanup(func() { rootCmd.SetArgs(nil) })
err := rootCmd.Execute()

// After:
out := &bytes.Buffer{}
cmd := newRootCmd()
cmd.SetOut(out)
cmd.SetErr(&bytes.Buffer{})
cmd.SetArgs(...)
err := cmd.Execute()
```

No `t.Cleanup` needed — `cmd` is local to the test.

**Verification:**
```
go test ./cmd/calf/... -v
go test -count=2 ./cmd/calf/...   # proves no shared state leakage
go test ./...
staticcheck ./...
```

All must pass.

---

## Item 11 — Error Message Coupling in `config_test.go` (WARN — no action required)

**File:** `internal/config/config_test.go`

**Assessment:** Tests use `strings.HasPrefix(err.Error(), "invalid CPU '0'")` to verify validation errors. This was introduced in TDD-R2 as the correct form (replacing fragile slice indexing). For a configuration package, verifying that validation errors report the field name and invalid value IS the observable behaviour — the error message is the user-facing output. This is acceptable behaviour testing.

**Action:** No change required. Resolved as non-issue.

---

## Execution Order

Work through items strictly in this order to keep the test suite green throughout:

1. ~~**Item 1** — `TestVMStateString` rename. DONE.~~
2. ~~**Item 2** — `TestCloneWhenTart*` `t.Run` wrapping. DONE.~~
3. ~~**Item 3** — `cacheDirMount` coupling. DONE (no change — coupling resolved by Items 5–6).~~
4. ~~**Item 4** — Move `ensureInstalled` to public methods (production change). DONE.~~
5. **Item 5** — Add functional options + update `NewTartClient` (production change). Run `go test ./...`.
6. **Item 6** — Rewrite `createTestClient` using options. Run `go test ./...`.
7. **Item 7** — Fix `client.pollTimeout` direct write. Run `go test ./internal/isolation/... -run TestIP`.
8. **Item 8** — Rewrite `TestCloneWhenTart*` using options. Run `go test ./internal/isolation/...`.
9. **Item 9** — Delete `makeInstallingRunCommand`. Run `go test ./...`.
10. **Item 10** — `newRootCmd()` and `newConfigCmd()` factories + fix `config_test.go` and `main_test.go`. Run `go test ./...`.

Final check: `go test ./...` and `staticcheck ./...` must both pass clean.

---

## Completion Criteria

- [x] `TestCloneWhenTart*` functions wrap bodies in `t.Run("when...should...", ...)`
- [x] `cacheDirMount` coupling assessed; no export needed — addressed by Items 5–6
- [x] `ensureInstalled` removed from `runTartCommand`; added to `Clone`, `Set`, `RunWithCacheDirs`, `Stop`, `Delete`, `List`, `IP`
- [ ] `TartClientOption` type exported; 7 option functions added: `WithRunCommand`, `WithPollInterval`, `WithPollTimeout`, `WithTartPath`, `WithLookPath`, `WithStdinReader`, `WithBrewRunner`
- [ ] `NewTartClient` accepts `...TartClientOption` (backwards-compatible)
- [ ] `createTestClient` uses `WithTartPath`, `WithPollInterval`, `WithPollTimeout`, `WithRunCommand` — no unexported field writes
- [ ] `client.pollTimeout = ...` at line 281 replaced with `NewTartClient(WithPollTimeout(...))`
- [ ] `TestCloneWhenTart*` uses `WithLookPath`, `WithStdinReader`, `WithBrewRunner`, `WithRunCommand` — no unexported field writes
- [ ] `makeInstallingRunCommand` deleted; no references remain
- [ ] `newConfigCmd()` factory added to `config.go`; global `configCmd`, `configShowCmd`, `vmName` vars and `init()` removed
- [ ] `newRootCmd()` factory added to `main.go`; wires `newConfigCmd()` and `newCacheCmd()`
- [ ] `config_test.go`: `setupConfigShow` returns fresh `cmd` via `newRootCmd()`; no `configShowCmd` reference; no flag cleanup
- [ ] `main_test.go`: all tests use `newRootCmd()` per test; no shared `rootCmd` mutations
- [ ] `go test ./...` passes (test count ≥ 203)
- [ ] `go test -count=2 ./cmd/calf/...` passes (proves no shared state leakage)
- [ ] `staticcheck ./...` passes with no warnings
