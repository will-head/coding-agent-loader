# Phase 1 (CLI Foundation) - Completed Items

> [← Back to PLAN.md](../PLAN.md)

**Status:** In Progress

**Goal:** Replace manual Tart commands with `calf isolation` CLI.

**Deliverable:** Working `calf isolation` CLI that wraps Tart operations.

---

## Critical Issue #5: No-Network SMB Bypass (Host Credentials) — ✅ COMPLETED (2026-02-10)

**Problem:** `--no-network` mode blocks local network IPs via softnet but still allowed SMB access to the host gateway (`192.168.64.1`). A VM running in "isolated" mode could mount the host's filesystem if valid credentials were provided — a full security bypass.

**Investigation:** Previous approach attempted to patch softnet source to add `--block-tcp-ports` / `--block-udp-ports` flags. All compiled softnet versions (patched or not) failed VM initialization. Homebrew softnet v0.18.0 works; compiled-from-source v0.1.0 (git main) does not. Investigation doc: [`docs/softnet-port-blocking-investigation.md`](softnet-port-blocking-investigation.md)

**Solution: Host-side pf anchor** — Standard Homebrew tart/softnet only. Block SMB/NetBIOS on the HOST using macOS `pf` with a temporary named anchor. No patched binaries. No changes to `/etc/pf.conf`.

**Architecture:**
- Anchor `com.apple/calf.smb-block` fits under existing `anchor "com.apple/*"` wildcard in `/etc/pf.conf`
- Rules: `block in quick proto tcp from $VM_IP to any port {445, 139}` + UDP 137/138
- Rules are in-memory only — no disk changes; removed on session end
- `--gui` mode: background watcher polls `kill -0 $TART_PID` every 5s; removes rules when VM process exits
- One-time sudoers drop-in `/etc/sudoers.d/calf-pfctl` enables NOPASSWD for load/flush/show on this anchor only
- SMB block failure = hard stop (VM killed); no-network mode without blocking is meaningless

**Security design fixes applied during implementation:**
1. **Security window eliminated** — `setup_smb_block_permissions()` runs BEFORE VM starts in all flows. Any sudo password prompt appears before the VM is running.
2. **NOPASSWD covers load** — `-f -` (rule load) added to sudoers alongside `-F all` (flush) and `-sr` (show). Without it, the most common operation still prompted.
3. **`pfctl -e` removed** — always a no-op on macOS 10.15+; NOPASSWD commands don't cache the sudo timestamp so the following non-NOPASSWD `pfctl -e` caused a spurious bare `Password:` prompt.
4. **Idempotency check via `sudo -n`** — replaced `sudo grep` on root-owned file with `sudo -n pfctl -sr` test. No bare prompt on re-runs.

**Key implementation details:**
- `~/.calf-vm-no-network` (host-side) persists mode across reboots; all subsequent `--run`/`--gui` enforce blocking even without flags
- `~/.calf-vm-config` (inside VM) stores `NO_NETWORK=true` for VM-side detection
- `disown $! 2>/dev/null || true` required for background watcher — `set -e` + non-interactive zsh makes `disown` fail
- Background watcher I/O: `</dev/null >>"$CALF_LOG" 2>&1 &` — writing to terminal from a background process while user is at prompt corrupts shell display

**Files changed:**
- `scripts/calf-bootstrap` — removed `resolve_patched_tart()`, simplified `resolve_softnet_path()`, added `start_smb_block()`, `stop_smb_block()`, `setup_smb_block_permissions()`, updated all VM start flows, added `--no-smb-block` / `--clear-smb-block` / `--remove-smb-permissions` flags, help text

**Verified (smoke tests 2026-02-10):**
- ✅ SMB TCP 445 blocked from inside VM (`nc` timeout)
- ✅ SMB TCP 139 blocked
- ✅ Internet works (github.com HTTP 200)
- ✅ No host mounts (`/Volumes/` = Macintosh HD only)
- ✅ VNC Finder confirms "Connection Failed" on host SMB
- ✅ `--init --safe-mode` + `--gui` (no flags) both enforce blocking via `~/.calf-vm-no-network` marker
- ✅ Rules removed when VM stops (watcher log confirms)
- ✅ No spurious password prompts after one-time NOPASSWD setup

**Reference:** `docs/softnet-port-blocking-investigation.md` (historical) • [PLAN-PHASE-01-TODO.md § 5a](PLAN-PHASE-01-TODO.md) (future Go helper daemon)

---

## Critical Issue #1: CLI Command Name Collision — ✅ COMPLETED (2026-02-07)

**Problem:** The `cal` command clashed with the system calendar command, requiring users to use `./cal` instead.

**Solution:** Renamed to `calf` (**C**oding **A**gent **L**oader **F**oundation) per [ADR-005](adr/ADR-005-cli-rename-cal-to-calf.md).

**Implementation Completed:**
- ✅ **1.1 Go Source Code** (9 files) - cmd/cal → cmd/calf, all references updated
- ✅ **1.2 Shell Scripts** (7 files) - calf-bootstrap, all VM scripts updated
- ✅ **1.3 Config/Flag File Paths** - Documented runtime path changes
- ✅ **1.4 Environment Variables** - CAL_VM → CALF_VM, CAL_LOG → CALF_LOG, etc.
- ✅ **1.5 Build System** - Makefile updated (go build -o calf)
- ✅ **1.6 Documentation** (51 files) - All .md files updated
- ✅ **1.7 Testing** - All tests pass, binary functional, user testing verified

**Results:**
- 68 files updated across entire codebase
- All Go tests pass (config: 0.300s, isolation: 0.274s)
- Binary builds and functions correctly (`calf --help`, `calf cache status`, `calf config show`)
- User testing validated: correct paths (.calf-cache), correct branding (CALF Configuration)

**Follow-up Tasks:**
- Some VM-side script references still need fixing (tracked in separate TODO)
- Smoke tests pending after VM script fixes complete

**Reference:** [ADR-005](adr/ADR-005-cli-rename-cal-to-calf.md)

---

## Critical Issue #2: Cache Clear Confirmation UX — ✅ COMPLETED (2026-02-07)

**Problem:** `calf cache clear --all` cleared all caches without any confirmation, creating risk of accidental data loss.

**Solution:** Added final y/N confirmation prompt and new `--force` flag for automation.

**Implementation Completed:**
- ✅ Added `--force` flag (`-f`) to skip all confirmations
- ✅ Updated `--all` flag description to reflect new behavior
- ✅ Added confirmation prompt logic with safe defaults (abort on anything except 'y')
- ✅ Updated command help text to document all usage modes
- ✅ Updated cli.md documentation with cache commands section

**Behavior:**
- `calf cache clear` → prompts for each cache individually (existing behavior)
- `calf cache clear --all` → shows warning + final y/N confirmation (NEW)
- `calf cache clear --all --force` → skips all confirmations for automation (NEW)
- `calf cache clear --dry-run` → previews what would be cleared without prompts

**Code Changes:**
- `cmd/calf/cache.go` - Added force flag, confirmation logic, updated descriptions
- `docs/cli.md` - Added Cache section documenting all cache commands

**Testing:**
- ✅ All unit tests pass (config, isolation packages)
- ✅ Build succeeds without errors
- ✅ Manual testing confirmed all three usage modes work correctly
- ✅ Confirmation accepts 'y' and proceeds
- ✅ Confirmation rejects 'N' and aborts
- ✅ Force flag skips all prompts
- ✅ Dry-run shows preview without confirmation

**Security:**
- Safe defaults: aborts on EOF, aborts on anything except explicit 'y'
- Case-insensitive comparison for user convenience
- Clear warning message explains impact
- Force flag requires explicit use for automation

**Impact:** Medium - prevents accidental data loss while maintaining automation support

---

## 1.1 **REFINED:** Project Scaffolding (PR #3, merged 2026-02-01)

**Tasks:**
1. Initialize Go module
   ```bash
   go mod init github.com/will-head/coding-agent-launcher
   ```
   - Use full repository path as module name
   - Enables internal imports like `import "github.com/will-head/coding-agent-launcher/internal/config"`

2. Create directory structure (directories only, add .go files when implementing features):
   ```
   cmd/cal/
   internal/
     config/
     isolation/
     agent/
     tui/
   ```
   - Test files will be added alongside code as features are implemented (e.g., `config_test.go` next to `config.go`)

3. Create `cmd/cal/main.go` with minimal Cobra root command:
   - Basic cobra setup with root command
   - Add version flag (`--version`)
   - Ready to add subcommands in later TODOs
   - Should be executable and respond to `calf --version`

4. Add initial dependencies (cobra/viper for CLI foundation):
   ```bash
   go get github.com/spf13/cobra
   go get github.com/spf13/viper
   ```
   - Add remaining dependencies (bubbletea, lipgloss, bubbles, ssh, yaml) incrementally as features are implemented
   - TUI libraries added when implementing TUI features (Phase 2)
   - SSH library added when implementing SSH management (1.5)

5. Create `.gitignore` (comprehensive):
   - Standard Go ignores: `cal` binary, `*.out`, `coverage.txt`, `vendor/`, build artifacts
   - IDE/editor files: `.vscode/`, `.idea/`, `*.swp`, `.DS_Store`
   - Local config/test files: `*.local`, `tmp/`, `test-output/`

6. Create `Makefile` with build automation:
   - `build`: Compile binary to `./cal` using `go build -o cal ./cmd/cal`
   - `test`: Run all tests with `go test ./...` (may be empty initially)
   - `lint`: Run `staticcheck ./...` for code quality checks
   - `install`: Install binary to `$(GOPATH)/bin` or `/usr/local/bin`
   - `clean`: Remove binary and test artifacts

**Acceptance Criteria:**
- [x] Project builds successfully: `go build ./cmd/cal` completes without errors
- [x] `make build` and `make test` execute successfully
- [x] `calf --version` runs and displays version information
- [x] All standard Go project files present: `go.mod`, `.gitignore`, `Makefile`, directory structure created
- [x] No placeholder .go files in internal/ (just directories)

**Constraints:**
- Use staticcheck for linting, not golangci-lint
- Keep scaffolding minimal - add files and dependencies incrementally
- Directory structure only - actual implementation files added in subsequent TODOs

**Estimated files:** 4 new files (`go.mod`, `.gitignore`, `Makefile`, `cmd/cal/main.go`) + directory structure

---

## 1.2 **REFINED:** Configuration Management (PR #4, merged 2026-02-01)

**Design Decisions:**
- **Precedence:** Per-VM config overrides global config overrides hard-coded defaults
- **Scope:** YAML configs only (`~/.calf/config.yaml` and per-VM `vm.yaml`). Other subsystems manage their own files (proxy module handles `~/.calf-proxy-config`, lifecycle handles flags, etc.)
- **Missing config:** Use hard-coded defaults silently (no auto-create, no errors)
- **Validation:** Error out immediately with clear messages including invalid value, expected range/format, and file path
- **Validation rules:** Strict validation using Tart-documented ranges:
  - CPU: Valid range from Tart documentation
  - Memory: Valid range from Tart documentation (MB)
  - Disk size: Valid range from Tart documentation (GB)
  - Proxy mode: Must be one of `auto`, `on`, `off`
  - Base image: String validation (non-empty)
- **Config inspection:** `calf config show [--vm name]` displays effective merged configuration

**Tasks:**
1. Define config structs in `internal/config/config.go` with schema version support
2. Implement config loading from `~/.calf/config.yaml` (optional file, silent fallback to defaults)
3. Implement per-VM config from `~/.calf/isolation/vms/{name}/vm.yaml` (optional)
4. Implement config merging logic: hard-coded defaults → global config → per-VM config
5. Add config validation with strict ranges from Tart documentation
6. Add hard-coded config defaults in code
7. Implement `calf config show [--vm name]` command to display effective merged config
8. Add clear error messages (format: "Invalid {field} '{value}' in {path}: must be {expected}")

**Config schema (from ADR):**
```yaml
version: 1
isolation:
  defaults:
    vm:
      cpu: 4
      memory: 8192
      disk_size: 80
      base_image: "ghcr.io/cirruslabs/macos-sequoia-base:latest"
    github:
      default_branch_prefix: "agent/"
    output:
      sync_dir: "~/calf-output"
    proxy:
      mode: "auto"  # auto, on, off
```

**Per-VM config example (`~/.calf/isolation/vms/heavy-build/vm.yaml`):**
```yaml
# Only specify fields to override from global config
cpu: 8
memory: 16384
# Other fields inherit from global config or defaults
```

**Config loading order:**
1. Load hard-coded defaults
2. Merge global config from `~/.calf/config.yaml` (if exists)
3. Merge per-VM config from `~/.calf/isolation/vms/{name}/vm.yaml` (if exists)
4. Result: Per-VM values override global values override defaults

**Acceptance criteria:**
- [x] Config loads from global and per-VM files with correct precedence
- [x] Missing config files handled gracefully (silent fallback to defaults)
- [x] Invalid config values rejected with clear error messages showing value, expected range, and file path
- [x] `calf config show` displays effective merged configuration for default VM
- [x] `calf config show --vm <name>` displays effective merged configuration for specific VM
- [x] Validation uses Tart-documented ranges (research Tart docs during implementation)
- [x] Other subsystems manage their own config files independently (config module doesn't touch them)

**Constraints:**
- YAML format only for Phase 1
- Must research Tart documentation for accurate validation ranges
- Error messages must include: field name, invalid value, expected range/format, file path where set
- Config module does NOT manage other VM files (listed below for reference only)

**Other VM files (NOT managed by config module - for reference only):**
- `~/.calf-vm-info` - VM metadata (managed by VM lifecycle subsystem)
- `~/.calf-vm-config` - VM password (managed by lifecycle subsystem, mode 600)
- `~/.calf-proxy-config` - Proxy settings (managed by proxy subsystem)
- `~/.calf-auth-needed` / `~/.calf-first-run` - Lifecycle flags (managed by lifecycle subsystem)
- `~/.tmux.conf` - tmux configuration (managed by SSH subsystem)
- `~/.zshrc` - Shell configuration (managed by lifecycle subsystem)
- `~/.zlogout` - Logout git status check (managed by git safety subsystem)

**Future enhancements (tracked as separate TODOs below):**
- Interactive config fixing on validation errors
- Environment variable overrides (e.g., `CAL_VM_CPU=8`)
- `calf config validate` command
- Config schema migration strategy for version changes
- `calf config show --defaults` to display hard-coded defaults

---

## 1.3 Tart Wrapper (PR #5, merged 2026-02-03)

**File:** `internal/isolation/tart.go`

**Implementation:**
- Implemented `TartClient` struct that wraps all Tart CLI operations
- Methods: Clone, Set, Run (headless/VNC), Stop, Delete, List, IP, Get
- JSON parsing using Go's `encoding/json` (no jq dependency)
- Auto-install Tart via Homebrew with interactive user prompt
- IP polling with progress indicator (2s interval, 60s timeout)
- VNC experimental mode (`--vnc-experimental`) by default for better UX
- Cache sharing (`--dir=tart-cache:~/.tart/cache:ro`) on all VM starts
- VM state tracking (running/stopped/not_found) with fresh queries
- Error wrapping with operation context for clear failure messages
- Comprehensive unit tests (27 tests) covering all methods and error paths

**Acceptance Criteria Met:**
- [x] All Tart operations wrapped with clear Go API
- [x] Errors include helpful context and operation details
- [x] IP polling shows progress and completes within 60s or fails clearly
- [x] Auto-install prompts user and handles missing Homebrew gracefully
- [x] Cache sharing enabled on all runs without user configuration
- [x] Unit tests cover command generation and error handling
- [x] No external dependencies (jq not required)

---

## 1.1.1 Homebrew Package Download Cache (PR #6, merged 2026-02-03)

**Cache Location:**
- **Host:** `~/.calf-cache/homebrew/` (persistent across VM operations)
- **VM:** Symlink `~/.calf-cache/homebrew/` → `/Volumes/My Shared Files/cal-cache/homebrew/`
- **Pattern:** Same as Tart cache sharing in section 1.9 (proven approach)

**Implementation Details:**

1. **Code Location:** `internal/isolation/cache.go`
   - New `CacheManager` struct with methods for setup, status
   - Integration point: Called from VM init/setup process
   - Follows existing isolation subsystem patterns

2. **Host Cache Setup:**
   - Create `~/.calf-cache/homebrew/` on host if doesn't exist
   - No host environment configuration needed (host uses default Homebrew cache)
   - Host directory structure: `~/.calf-cache/homebrew/downloads/`, `~/.calf-cache/homebrew/Cask/`

3. **VM Cache Passthrough (Symlink Approach):**
   - Create Tart shared directory: Ensure `/Volumes/My Shared Files/cal-cache/` exists
   - Copy host cache to shared volume: `rsync -a ~/.calf-cache/homebrew/ "/Volumes/My Shared Files/cal-cache/homebrew/"`
   - Create symlink in VM: `ln -sf "/Volumes/My Shared Files/cal-cache/homebrew" ~/.calf-cache/homebrew`
   - Configure in VM: `export HOMEBREW_CACHE=~/.calf-cache/homebrew` (add to `.zshrc`)
   - Verify symlink writable from VM

4. **Error Handling (Graceful Degradation):**
   - If symlink creation fails: Log warning, continue without cache
   - If shared volume unavailable: Log warning, continue without cache
   - Bootstrap still works, just slower (no hard failure)
   - Consistent with Tart cache sharing pattern in section 1.9

5. **Cache Status Command:** `calf cache status`
   - Display information:
     - Cache sizes per package manager (e.g., "Homebrew: 450 MB")
     - Cache location path (e.g., "~/.calf-cache/homebrew/")
     - Cache availability status (✓ Homebrew cache ready, ✗ npm cache not configured)
     - Last access time (from filesystem mtime)
   - Output format: Human-readable table or list
   - Implementation: `internal/isolation/cache.go` → `Status()` method

6. **Cache Invalidation Strategy:**
   - **Let package managers handle it** - no manual invalidation logic
   - Homebrew validates cache integrity and checksums automatically
   - Invalid or outdated cache entries are re-downloaded by Homebrew
   - Simplest approach: just set `HOMEBREW_CACHE` and let Homebrew manage lifetime

**Benefits:**
- **Speed:** Saves ~5-10 minutes per bootstrap (biggest single win)
- **Reliability:** Reduces network dependency, fewer timeout failures
- **Bandwidth:** Saves hundreds of MB per bootstrap iteration
- **Development:** Faster snapshot/restore testing cycles

**Constraints:**
- Requires Tart shared directories feature (graceful degradation if unavailable)
- Disk space: ~500-800 MB for Homebrew cache
- Cache persists across VM operations (intended behavior)

**Testing Strategy:**
- **Unit Tests:** Cache setup logic, symlink creation, graceful degradation paths
- **Integration Tests (with mocks):** Mock filesystem operations, verify environment configuration
- **Manual Testing:**
  - First bootstrap: Download everything, populate cache
  - Second bootstrap: Verify cache used, measure time improvement
  - Snapshot/restore: Verify cache persists and remains functional
  - Symlink failure: Verify graceful degradation (bootstrap completes without cache)

**Acceptance Criteria:**
- [x] Homebrew cache directory created on host
- [x] Symlink created in VM pointing to shared cache
- [x] `HOMEBREW_CACHE` environment variable set in VM
- [x] `calf cache status` shows Homebrew cache info (size, location, availability, last access)
- [x] Bootstrap time reduced by at least 30% on second run (Homebrew portion)
- [x] Graceful degradation works if symlink fails
- [x] Unit and integration tests pass
- [x] Documentation updated in ADR-002

**Related:**
- Section 1.9: VM lifecycle automation (Tart cache sharing pattern reference)
- BUG-006: Network timeout during bootstrap (Homebrew cache will help prevent)

---

## 1.1.2 npm Package Download Cache (PR #7, merged 2026-02-03)

**Cache Location:**
- **Host:** `~/.calf-cache/npm/` (persistent across VM operations)
- **VM:** Symlink `~/.calf-cache/npm/` → `/Volumes/My Shared Files/cal-cache/npm/`
- **Pattern:** Same as Phase 1.1.1 (proven approach)

**Implementation Details:**

1. **Code Location:** `internal/isolation/cache.go` (extend existing `CacheManager`)
   - Add npm-specific setup method
   - Integrate into VM init/setup process

2. **Host Cache Setup:**
   - Create `~/.calf-cache/npm/` on host if doesn't exist
   - No host environment configuration needed

3. **VM Cache Passthrough:**
   - Create Tart shared directory: `/Volumes/My Shared Files/cal-cache/npm/`
   - Copy host cache: `rsync -a ~/.calf-cache/npm/ "/Volumes/My Shared Files/cal-cache/npm/"`
   - Create symlink in VM: `ln -sf "/Volumes/My Shared Files/cal-cache/npm" ~/.calf-cache/npm`
   - Configure in VM: `npm config set cache ~/.calf-cache/npm` (run during vm-setup.sh)
   - Verify symlink writable

4. **Error Handling:** Graceful degradation (same as Phase 1.1.1)

5. **Cache Status:** Update `calf cache status` to include npm cache info

6. **Cache Invalidation:** Let npm handle it (validates cache metadata automatically)

**Benefits:**
- **Speed:** Saves ~2-3 minutes per bootstrap
- **Bandwidth:** Saves ~50-100 MB per bootstrap
- **Packages:** claude, agent, ccs, codex CLI tools

**Constraints:**
- Disk space: ~100-200 MB for npm cache

**Testing Strategy:**
- Unit tests for npm cache setup logic
- Integration tests with mocks
- Manual: Bootstrap twice, verify npm uses cache

**Acceptance Criteria:**
- [x] npm cache directory created on host
- [x] Symlink created in VM
- [x] `npm config get cache` returns `~/.calf-cache/npm` in VM
- [x] `calf cache status` shows npm cache info
- [x] Bootstrap time reduced by additional 15-20% with npm cache
- [x] Graceful degradation works
- [x] Tests pass

---

## 1.1.3 Go Modules Cache (PR #8, merged 2026-02-03)

**Dependencies:** Phase 1.1.2 (npm cache) must be complete first.

**Status:** Merged

**Cache Location:**
- **Host:** `~/.calf-cache/go/pkg/mod/` (persistent across VM operations)
- **VM:** Symlink `~/.calf-cache/go/` → `/Volumes/My Shared Files/cal-cache/go/`
- **Pattern:** Same as Phases 1.1.1 and 1.1.2

**Implementation Details:**

1. **Code Location:** `internal/isolation/cache.go` (extend existing `CacheManager`)
   - Add Go-specific setup method

2. **Host Cache Setup:**
   - Create `~/.calf-cache/go/pkg/mod/` on host if doesn't exist
   - Create `~/.calf-cache/go/pkg/sumdb/` for checksum database

3. **VM Cache Passthrough:**
   - Create Tart shared directory: `/Volumes/My Shared Files/cal-cache/go/`
   - Copy host cache: `rsync -a ~/.calf-cache/go/ "/Volumes/My Shared Files/cal-cache/go/"`
   - Create symlink in VM: `ln -sf "/Volumes/My Shared Files/cal-cache/go" ~/.calf-cache/go`
   - Configure in VM: `export GOMODCACHE=~/.calf-cache/go/pkg/mod` (add to `.zshrc`)
   - Verify symlink writable

4. **Error Handling:** Graceful degradation (same as previous phases)

5. **Cache Status:** Update `calf cache status` to include Go cache info

6. **Cache Invalidation:** Let Go handle it (uses `go.sum` checksums for validation)

**Benefits:**
- **Speed:** Saves ~1-2 minutes per bootstrap
- **Bandwidth:** Saves ~20-50 MB per bootstrap
- **Modules:** staticcheck, goimports, delve, mockgen, air

**Constraints:**
- Disk space: ~50-150 MB for Go module cache

**Testing Strategy:**
- Unit tests for Go cache setup logic
- Integration tests with mocks
- Manual: Bootstrap twice, verify Go uses cache

**Acceptance Criteria:**
- [x] Go cache directory created on host
- [x] Symlink created in VM
- [x] `go env GOMODCACHE` returns `~/.calf-cache/go/pkg/mod` in VM
- [x] `calf cache status` shows Go cache info
- [x] Bootstrap time reduced by additional 10-15% with Go cache
- [x] Graceful degradation works
- [x] Tests pass

## 1.1.4 **REFINED:** Git Clones Cache (PR #9, merged 2026-02-03)

**Dependencies:** Phase 1.1.3 (Go modules cache) complete.

**Status:** Merged with complete cache integration

**Cache Location:**
- **Host:** `~/.calf-cache/git/<repo-name>/` (persistent across VM operations)
- **VM:** Shared via `/Volumes/My Shared Files/cal-cache/git/<repo-name>/`
- **Pattern:** Selective caching for frequently cloned repos (TPM)

**Implementation Highlights:**

1. **Code Location:** `internal/isolation/cache.go`
   - Extended `CacheManager` with git cache methods
   - `SetupGitCache()`, `GetGitCacheInfo()`, `CacheGitRepo()`, etc.
   - Unit tests with full coverage

2. **Bootstrap Integration:**
   - `calf-bootstrap`: Cache directory creation during --init
   - `vm-setup.sh`: VM cache configuration (permanent)
   - `vm-tmux-resurrect.sh`: TPM caching from shared host cache
   - Host cache temporary (script-only), VM cache permanent

3. **Cache Population:**
   - TPM cloned from GitHub on first install
   - TPM cached to shared volume for future bootstraps
   - Cache updated with `git fetch` before use
   - Graceful fallback to GitHub if cache unavailable

4. **Additional Improvements:**
   - Homebrew/npm/Go cache integration into bootstrap
   - Cursor CLI via Homebrew Cask (now cacheable)
   - Complete package manager cache configuration
   - 100% of downloads now cached

**Testing Results:**
- ✅ All manual tests passed (git cache, TPM caching, offline bootstrap)
- ✅ Cache sharing verified working
- ✅ Offline capability confirmed
- ✅ Unit tests: 330 tests passing

**Benefits Realized:**
- **Speed:** ~30-60 seconds saved per bootstrap (TPM)
- **Offline:** Works without network after first bootstrap
- **Total Cache:** ~20-30GB (all package managers + git repos)
- **Integration:** All package managers use shared cache

**Documentation:**
- docs/PR-9-TEST-RESULTS.md - Complete test results
- docs/PR-9-INIT-REVIEW.md - Init integration review
- docs/CACHE-INTEGRATION.md - Integration design and details

**Acceptance Criteria Met:**
- [x] Git cache directory created on host
- [x] TPM cached and used during bootstrap
- [x] Cache updated with `git fetch` before use
- [x] `calf cache status` shows cached git repos
- [x] Bootstrap works offline with cached repos
- [x] Graceful degradation when cache unavailable
- [x] All tests pass (unit + manual)
- [x] Bootstrap integration complete

---

## 1.1.5 Cache Clear Command (PR #10, merged 2026-02-04)

**Dependencies:** Phases 1.1.1-1.1.4 must be complete first (all caches implemented).

**Command:** `calf cache clear`

**Implementation Details:**

1. **Code Location:** `internal/isolation/cache.go` (extend existing `CacheManager`)
   - Add `Clear()` method with per-cache confirmation

2. **Behavior:**
   - Prompt user to confirm clearing each cache type individually
   - Example flow:
     ```
     Clear Homebrew cache (450 MB)? [y/N]: y
     Clearing Homebrew cache...
     Clear npm cache (120 MB)? [y/N]: n
     Skipping npm cache
     Clear Go modules cache (80 MB)? [y/N]: y
     Clearing Go modules cache...
     Clear git clones cache (25 MB)? [y/N]: y
     Clearing git clones cache...

     Summary: Cleared 555 MB (3/4 caches)
     ```

3. **Implementation:**
   - For each cache type (Homebrew, npm, Go, Git):
     - Calculate cache size: `du -sh <cache-dir>`
     - Prompt user: `Clear <type> cache (<size>)? [y/N]:`
     - If confirmed: `rm -rf <cache-dir>` and recreate empty directory
     - Track cleared caches for summary
   - Display summary of total space freed

4. **Flags:**
   - `--all` or `-a`: Clear all caches without prompting (dangerous)
   - `--dry-run`: Show what would be cleared without actually clearing

5. **Safety:**
   - Default to "No" for each prompt (require explicit "y")
   - Warn if clearing will slow down next bootstrap
   - Suggest alternatives: "Consider clearing individual caches if low on disk space"

**Benefits:**
- **Disk Management:** Users can reclaim 1-2 GB when needed
- **Troubleshooting:** Clear corrupted caches
- **Flexibility:** Per-cache granularity with confirmation

**Constraints:**
- Clearing cache means next bootstrap will be slow again
- No undo (must re-download everything)

**Testing Strategy:**
- Unit tests for clear logic with mocks
- Integration tests for confirmation prompts
- Manual: Test clearing each cache individually and all together

**Acceptance Criteria Met:**
- [x] `calf cache clear` prompts for each cache individually
- [x] Each cache cleared only after user confirms "y"
- [x] Summary shows total space freed
- [x] `--all` flag clears all without prompting
- [x] `--dry-run` shows what would be cleared
- [x] Graceful handling if cache doesn't exist
- [x] Tests pass

---

## Critical Issue #3: Cache Mount Architecture (completed 2026-02-07)

**Problem:** Symlink-based cache architecture was fragile:
- Easily deleted during cache clear operations with `rm -rf`
- Not automatically repaired if broken
- Confusing for testing and troubleshooting

**Solution:** Direct virtio-fs mounting with custom tags per [ADR-004](adr/ADR-004-cache-mount-architecture.md)

**Implementation:**

### Files Created
1. **scripts/calf-mount-shares.sh** - Mount script with retry logic
   - Mounts virtio-fs shares to target locations
   - Retry logic for boot timing (5 attempts, 2s delay)
   - Logging to `/tmp/cal-mount.log`
   - Called by LaunchDaemon at boot
   - Extensible for future mounts (iOS signing, etc.)

2. **scripts/com.cal.mount-shares.plist** - LaunchDaemon for boot-time persistence
   - RunAtLoad + KeepAlive on failure
   - Sets HOME=/Users/admin and USER=admin environment
   - Logs to `/tmp/cal-mount.log`

### Files Modified
1. **scripts/calf-bootstrap** (3 changes)
   - Line 241 & 1747: Updated Tart `--dir` flag from old format to new: `--dir "${HOME}/.calf-cache:tag=cal-cache"`
   - Line 484: Added `calf-mount-shares.sh` and `com.cal.mount-shares.plist` to deployment array

2. **scripts/calf-mount-shares.sh** - macOS compatibility + simplification
   - Replaced Linux `mountpoint -q` with macOS `mount | grep -q` (2 occurrences)
   - Removed migration logic (11 lines) - simplified architecture

3. **scripts/vm-setup.sh** - macOS compatibility
   - Replaced `mountpoint -q` with macOS-compatible `mount | grep -q` (2 occurrences)
   - Lines 52: Direct check
   - Line 74: In .zshrc heredoc for self-healing

4. **internal/isolation/cache.go** - macOS compatibility + flexibility
   - Updated 4 methods: `SetupVMHomebrewCache()`, `SetupVMNpmCache()`, `SetupVMGoCache()`, `SetupVMGitCache()`
   - Changed from `mountpoint -q ~/.calf-cache` to `mount | grep -q " on $HOME/.calf-cache "`
   - Uses `$HOME` variable instead of hardcoded `/Users/admin` path

5. **internal/isolation/cache_test.go** - Test assertions updated
   - Updated 4 test assertions to match new mount verification commands
   - Changed expected strings from `mountpoint -q` to `mount | grep -q`

### Architecture Changes

**Before (Symlinks):**
```
Host: --dir calf-cache:~/.calf-cache:rw,tag=com.apple.virtio-fs.automount
VM:   /Volumes/My Shared Files/cal-cache/
      ~/.calf-cache/homebrew -> /Volumes/.../homebrew (symlink)
      ~/.calf-cache/npm -> /Volumes/.../npm (symlink)
```

**After (Direct Mounts):**
```
Host: --dir ${HOME}/.calf-cache:tag=cal-cache
VM:   ~/.calf-cache (direct mount via calf-mount-shares.sh)
      ├── homebrew/ (directly accessible)
      ├── npm/
      ├── go/
      └── git/
```

### Key Improvements
- ✅ **Robustness:** Mount cannot be deleted with `rm -rf` (returns "Resource busy")
- ✅ **Self-healing:** LaunchDaemon + .zshrc fallback automatically restore mounts
- ✅ **macOS Compatible:** Uses `mount | grep` instead of Linux-only `mountpoint -q`
- ✅ **Persistent:** Survives reboot via LaunchDaemon
- ✅ **Snapshot-safe:** Mount infrastructure survives snapshot/restore operations
- ✅ **Simplified:** Removed unnecessary migration logic
- ✅ **Flexible:** Uses `$HOME` variable for portability

### Testing Results
All manual tests passed (8/8):
- ✅ Test 1: Fresh Init - Script Deployment
- ✅ Test 2: Mount Functionality
- ✅ Test 3: Reboot Persistence (LaunchDaemon)
- ✅ Test 4: Self-Healing (.zshrc fallback)
- ✅ Test 5: Robustness (`rm -rf` protection)
- ✅ Test 6: Cache Functionality
- ⏭️ Test 7: Migration (skipped - migration logic removed)
- ✅ Test 8: Snapshot/Restore

Go unit tests: All passed (`go test ./internal/isolation/... -v`)

### Future Work
See PLAN-PHASE-01-TODO.md § 6 for Go code parity cleanup (low priority):
- Remove dead code (`sharedCacheMount` constant, `GetSharedCacheMount()` method)
- Update comments referencing old architecture
- Verify symlink handling logic still needed

**Reference:** [ADR-004](adr/ADR-004-cache-mount-architecture.md) for complete architecture decision record

