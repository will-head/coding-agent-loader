# Phase 1 (CLI Foundation) - TODOs

> [← Back to PLAN.md](../PLAN.md)

**Status:** In Progress

**Goal:** Replace manual Tart commands with `calf isolation` CLI.

**Deliverable:** Working `calf isolation` CLI that wraps Tart operations.

**Reference:** [ADR-002](adr/ADR-002-tart-vm-operational-guide.md) § Phase 1 Readiness for complete operational requirements.

---

## Critical Issues - HIGHEST PRIORITY

### 0. **BLOCKER:** TDD Remediation — Bring Codebase into coops-tdd Compliance

**No new code may be written until this is complete.**

The codebase was implemented before the `coops-tdd` skill was adopted. Existing code must be brought into full compliance as if it had been written test-first from the beginning.

**Full plan:** [`docs/TDD-REMEDIATION-PLAN.md`](TDD-REMEDIATION-PLAN.md)

**Summary of work (execute in order):**
- Pre-work: Fix invalid `go.mod` version
- Item 0: Rename all existing tests to `when [condition] should [outcome]` convention + AAA structure pass
- Item 1: Add direct tests for `GetDefaultConfigPath` / `GetVMConfigPath`
- Item 2: Fix `Run`/`RunWithCacheDirs` to route through injectable `runCommand`; delete no-op tests and `cacheDirMount` constant test; add behavioral replacements
- Item 3: Make `ensureInstalled` stdin injectable; test Homebrew install branch
- Item 4: Add `NewCacheManagerWithDirs` constructor; replace struct literals in tests
- Item 5: Fix `TestCacheManager_Clear` sub-test isolation (shared state)
- Item 6: Extract `getCacheInfo` helper to eliminate 4× duplication
- Item 7: Fix `UpdateGitRepos` to surface errors to caller instead of always returning `nil`
- Item 8: Add tests for `cmd/calf/config.go`; convert `os.Exit` to `RunE`
- Item 9: Add tests for `cmd/calf/cache.go`; make stdin injectable
- Item 10: Add root command dispatch tests for `cmd/calf/main.go`

**Done when:** `go test ./...` passes, `staticcheck ./...` passes, all tests satisfy Kent Beck's Test Desiderata (isolated, deterministic, fast, behavioral, structure-insensitive, readable).

---

### 4. **REFINED:** Bootstrap Init Logic - Update vs Full Recreate Behavior

**Problem:** When both calf-dev and calf-init exist, `calf-bootstrap --init` offers to update calf-init from calf-dev. If user declines, script aborts completely (exit 0). User cannot proceed with full fresh init even if desired.

**Current Behavior (lines 1436-1506 in calf-bootstrap):**
```
Do you want to replace calf-init with current calf-dev? (y/N)
  → If yes: Updates calf-init, exits
  → If no: "Aborted. Existing VMs not modified." exits with code 0  ← dead end
```

**Required Behavior:**
```
Do you want to replace calf-init with current calf-dev? (y/N)
  → If yes: Updates calf-init from calf-dev (existing behavior, unchanged)
  → If no: fall through to step 2

Delete calf-dev and calf-init, then re-initialize? (y/N)
  → If yes: git safety check on calf-dev → delete both VMs → fresh init from scratch
  → If no: "Aborted. Existing VMs not modified." exits with code 0
```

**Implementation:**
- `calf-bootstrap`: Remove `exit 0` at line 1505 (the `else` branch on decline) so it falls through to the existing full-init flow already at line 1510. The full-init flow already handles git safety checks, deletion prompts, and fresh init — no new logic needed.
- `calf isolation init` (Go): Implement the same two-step flow when both VMs exist.

**Acceptance Criteria:**
- Declining the update offer presents the full-reinit option instead of aborting
- Full reinit runs git safety checks before deleting VMs
- Confirming full reinit deletes both calf-dev and calf-init, then starts fresh init
- Declining full reinit exits cleanly with "Aborted. Existing VMs not modified."
- `--yes` flag skips both prompts and proceeds directly with full reinit (delete both, reinit)
- Go `calf isolation init` mirrors the same two-step flow

**Constraints:**
- The "replace calf-init with calf-dev" shortcut path (Y on first prompt) is unchanged
- No new flags needed
- calf-bootstrap fix is a minimal one-line change (remove erroneous exit 0)

**Related:** May interact with `--no-mount` implementation (New Feature #5)

---

### 4b. **REFINED:** Git Safety Check: Worktree Awareness (calf-bootstrap)

**Problem:** `check_vm_git_changes` in `calf-bootstrap` silently misses uncommitted or unpushed work in git worktrees, risking silent data loss when the VM is deleted or reset.

**Root Cause:**
1. `find -name ".git" -type d` skips worktree `.git` entries — they are files, not directories.
2. `git status --porcelain` run in the main checkout only reports that checkout's working tree; changes in linked worktrees are invisible.

**Affected operations in calf-bootstrap:**
- `--init` (delete calf-dev before reinit)
- `--snapshot restore` (replace calf-dev with snapshot)
- `--snapshot delete` (delete VM)

**Approach:** For each main `.git` dir found, use `git worktree list --porcelain` to enumerate linked worktrees, then run both checks in each linked worktree.

**Implementation — extend `check_vm_git_changes` (lines 750–811):**

The SSH command that finds repos and checks `git status --porcelain` must be extended to also loop over linked worktrees for each found repo. Apply the same extension to the unpushed-commits SSH command.

Pattern for uncommitted check:
```bash
for gitdir in $(find ~/workspace ~/projects ~/repos ~/code 2>/dev/null -name ".git" -type d; \
                find ~ -maxdepth 2 -name ".git" -type d 2>/dev/null | sort -u); do
  dir=$(dirname "$gitdir")
  # check main checkout
  (cd "$dir" && [ -n "$(git status --porcelain 2>/dev/null)" ] && echo "$dir")
  # check each linked worktree
  git -C "$dir" worktree list --porcelain 2>/dev/null \
    | awk '/^worktree /{print $2}' \
    | grep -v "^${dir}$" \
    | while read -r wt_dir; do
        (cd "$wt_dir" 2>/dev/null && [ -n "$(git status --porcelain 2>/dev/null)" ] && echo "$wt_dir")
      done
done
```

Apply the same worktree loop to the unpushed-commits check (use `git log "@{u}.." --oneline` in each worktree that has an upstream set).

**Acceptance Criteria:**
- Uncommitted changes in any linked worktree (e.g., `.claude/worktrees/implement/foo`) trigger the same warning as changes in the main checkout
- Unpushed commits in any linked worktree are also detected
- Repos with no worktrees: no behaviour change (worktree list returns only the main entry, loop body never executes)
- `--force` flag continues to skip all git checks as before
- All three affected operations (`--init`, `--snapshot restore`, `--snapshot delete`) benefit automatically (they all call `check_vm_git_changes`)

**Constraints:**
- `git worktree list` available in Git 2.5+ (Homebrew default is well above this)
- Change is purely additive — existing main-checkout check is unchanged; worktree loop is appended
- Go `CheckGitChanges()` in `safety.go` (1.7) is a separate TODO — see note in 1.7

**Testing:**
- Repo with no worktrees: existing behaviour unchanged
- Repo with one linked worktree with uncommitted changes: warning fires
- Repo with one linked worktree with unpushed commits: warning fires
- `--force` flag: no git checks run at all
- Repo where worktree directory no longer exists: graceful skip (no error)

---

### 5a. Future: Go Privileged Helper Daemon for pf Management

**Context:** The current `--gui` background watcher uses a shell loop + sudoers NOPASSWD entry
(`/etc/sudoers.d/calf-pfctl`) to remove pf rules when the VM stops. This works but has a
5-second poll delay and requires a one-time sudoers setup.

**Goal:** Replace the shell watcher with a small Go `LaunchDaemon` that manages pf rule
lifetime using `kqueue EVFILT_PROC NOTE_EXIT` — the same pattern used by Docker Desktop
(`com.docker.vmnetd`) and Mullvad VPN.

**Design:**
- Small Go binary (`calf-netd`) running as a root `LaunchDaemon`
- Listens on a Unix socket; accepts two messages: `load-anchor <vm-ip>` and `unload-anchor`
- Uses `kqueue EVFILT_PROC NOTE_EXIT` to watch the tart PID — no polling
- Calls `pfctl` to load/flush the anchor atomically
- Crash-safe: if tart dies unexpectedly, kqueue fires immediately and rules are removed
- Removes need for `/etc/sudoers.d/calf-pfctl` entirely

**Why deferred:**
- Requires Go binary + `LaunchDaemon` plist installation (higher setup cost)
- Current sudoers approach works well enough for Phase 1
- Natural fit for Phase 1 CLI work when the Go `calf isolation` command is built

**Implementation notes:**
- See research findings: macOS pf has NO automatic anchor cleanup on process exit (confirmed via `xnu/bsd/net/pf_ioctl.c`)
- The pf token system (`-E`/`-X` flags, `DIOCSTARTREF`/`DIOCSTOPREF`) controls only whether pf is enabled — does NOT affect anchor lifetime
- `kqueue EVFILT_PROC` with `NOTE_EXIT` is the correct kernel mechanism for PID death watching
- Reference implementations: `mullvad/pfctl-rs`, Docker Desktop `vmnetd`

---

### 6. Repository Rename - launcher → loader (HIGH PRIORITY)

**Problem:** Repository name `coding-agent-launcher` doesn't match the CALF acronym (Coding Agent **L**oader).

**Scope:**
1. **GitHub rename:** `coding-agent-launcher` → `coding-agent-loader`
2. **go.mod:** Update module path to `github.com/will-head/coding-agent-loader`
3. **Import paths:** Update all `import` statements in Go files (2 files: `cmd/calf/cache.go`, `cmd/calf/config.go`)
4. **STATUS.md:** Update all PR links (10 URLs) to new repo name
5. **Local filesystem:** Rename working directory from `coding-agent-launcher` to `coding-agent-loader`
6. **git remote:** Update origin URL after local rename

**Out of Scope (immutable):**
- `docs/adr/ADR-005-*.md` — references old name (historical record)
- `docs/prd/prd-001-*.md` — references old name (historical record)

**Verification:**
- `go build ./...` succeeds
- `go test ./...` passes
- `git push` works

**Impact:** High — blocks other work if delayed (more commits = more PR links to update)

---

## New Features - Normal Priority

### 4. CLI Proxy Utility for VM↔Host Command Transport

**Goal:** Enable VM-based applications to execute CLI commands on the host transparently.

**Use Case:** 1Password's `op` CLI requires communication with the 1Password desktop app, which runs on the host and is not accessible from within the VM.

**Concept:**
1. Alias commands in VM (e.g., `op` → `cli-proxy op`)
2. `cli-proxy` securely transports command and arguments from VM to host
3. Host executes actual command against host resources (e.g., 1Password desktop app)
4. Results returned securely to VM
5. Transparent to user - feels like native command execution

**Requirements:**
- Secure transport mechanism (SSH-based, encrypted)
- Verbatim command/response passing
- Low latency for interactive commands
- Error handling and exit code preservation
- Support for stdin/stdout/stderr
- Configurable command allowlist for security

**Potential Implementation:**
- SSH-based command forwarding
- Host-side daemon/service to receive and execute commands
- VM-side client wrapper (`cli-proxy`)
- Configuration file for allowed commands

---

### 5. No-Mount Mode for Secure Isolated VMs

**Status:** ✅ **Implemented in calf-bootstrap** (2026-02-07)

**Goal:** Enable creation of fully isolated VMs with no host filesystem mounts for maximum security.

**Use Case:** Secure locked-down VM with zero risk of host filesystem disruption. Useful for untrusted code execution or high-security development environments.

**Implementation (Completed in calf-bootstrap):**
1. ✅ Add `--no-mount` flag to `calf-bootstrap --init` command
2. ✅ When set, VM:
   - Does NOT mount `calf-cache` from host
   - Does NOT mount `tart-cache` from host
   - Creates local `~/.calf-cache` folder inside VM for package caching
   - Uses VM-local Tart cache
3. ✅ Setting is permanent and enforced for VM lifetime:
   - Can only be set at VM creation time (`init`)
   - Stored in `~/.calf-vm-config` with `NO_MOUNT=true` inside VM
   - Host marker file `~/.calf-vm-no-mount` tracks mode for subsequent operations
   - All subsequent operations (start, restart, gui) respect this setting
   - Cannot be changed after creation (VM must be destroyed and recreated)
4. ✅ Updated `calf-bootstrap` script with full support
5. ✅ Updated `calf-mount-shares.sh` to check flag and create local dirs when `NO_MOUNT=true`
6. ✅ Added permanent setting warning with Y/n confirmation
7. ✅ Added mount mode to `--status` output
8. ✅ Updated documentation (bootstrap.md)

**Remaining Work:**
- [ ] Add `--no-mount` support to Go implementation (`calf isolation init`)
- [ ] Testing and validation in VM environment

**Impact:** Medium - enhances security options for sensitive workloads

---

### 6. Screenshot Drag-and-Drop Support for VM-based Coding Agents

**Goal:** Enable drag-and-drop of screenshots from host into coding agents running in the VM.

**Current Limitation:** On the host system, users can drag and drop screenshots directly into coding agents. This functionality doesn't work when the coding agent runs in the VM.

**Required Investigation:**
1. How do coding agents currently receive drag-and-drop screenshots?
   - Clipboard integration?
   - File path passing?
   - Direct image data?
2. What's the technical barrier in the VM?
   - Clipboard isolation?
   - File system isolation?
   - GUI application integration?
3. Potential solutions:
   - Shared clipboard between host and VM
   - Automatic screenshot sync to VM filesystem
   - VNC/remote desktop integration improvements
   - Custom bridge application

**Acceptance Criteria:**
- User can drag screenshot from host desktop
- Screenshot appears in coding agent running in VM
- Works with common coding agents (Claude Code, Cursor, etc.)
- Minimal latency (feels instant)

---

### 7. Go Code Parity with Updated Cache Mount Architecture

**Goal:** Update Go implementation (`internal/isolation/cache.go` and `internal/isolation/tart.go`) to match the new direct virtio-fs mount architecture implemented in calf-bootstrap and scripts.

**Background:** Critical Issue #3 updated shell scripts to use direct mounts instead of symlinks, with macOS-compatible mount verification. The Go code has legacy/dead code and outdated patterns that need cleanup for consistency.

**Required Changes:**

#### 6.1 Remove Dead Code
- [ ] Remove `sharedCacheMount` constant (cache.go:47) - unused, references old mount format
- [ ] Remove `GetSharedCacheMount()` method (cache.go:90-93) - only used in tests, never in production
- [ ] Remove test `TestCacheManager_SharedCacheMount` (cache_test.go:289-296) - tests dead code
- [ ] Update `GetHomebrewCacheHostPath()` (cache.go:95-98) if it references old format

#### 6.2 Verify Mount Specification Format
- [ ] Check if tart.go needs cal-cache mount support (currently only has tart-cache mount)
- [ ] If adding cal-cache mount to tart.go, use new format: `${HOME}/.calf-cache:tag=cal-cache`
- [ ] Ensure consistency with calf-bootstrap lines 241 & 1747

#### 6.3 Update Comments and Documentation
- [ ] Update comment "Mount is handled by calf-mount-shares.sh via LaunchDaemon" (appears 4x) - verify accuracy
- [ ] Review symlink-related comments - some may reference old architecture
- [ ] Update package-level documentation if it references symlink-based caching

#### 6.4 Verify Symlink Handling Logic
- [ ] Review `resolveRealCachePath()` (cache.go:390-429) - confirm still needed for backwards compat
- [ ] Review symlink preservation in `Clear()` (cache.go:744-792) - confirm still needed
- [ ] Document if/when symlinks are still used vs. direct mounts

**Impact:** Low urgency - Go code works correctly with new architecture, this is cleanup/consistency

**Testing:**
- Unit tests already pass (confirmed 2026-02-07)
- No functional impact - purely cleanup

**Reference:**
- ADR-004 for mount architecture
- calf-bootstrap lines 241, 1747 for mount specification format
- Code review findings from Critical Issue #3 implementation

---

## 1.4 Snapshot Management

**File:** `internal/isolation/snapshot.go`

**Tasks:**
1. Implement `SnapshotManager` struct
2. Methods:
   - `Create(name)` - create snapshot via `tart clone` (stop VM first)
   - `Restore(name)` - restore from snapshot (check git, delete calf-dev, clone)
   - `List()` - list snapshots with sizes (JSON format)
   - `Delete(names, force)` - delete one or more snapshots
   - `Cleanup(olderThan, autoOnly)` - cleanup old snapshots
3. Auto-snapshot on session start (configurable)
4. Snapshot naming: user-provided exact names (no prefix)

**Key learnings from Phase 0 (ADR-002):**
- Tart "snapshots" are actually clones (copy-on-write)
- Restore must work even if calf-dev doesn't exist (create from snapshot)
- Delete supports multiple VM names in one command
- `--force` flag skips git checks and avoids booting VM (for unresponsive VMs)
- Git safety checks must run before restore and delete (see 1.7)
- Don't stop a running VM before git check - use it while running
- Snapshot names are case-sensitive, no prefix required
- **Filesystem sync before snapshot creation** (BUG-009): Call `sync && sleep 2` via SSH before stopping VM for snapshot. Without this, data written by SSH operations (e.g., repo clones during vm-auth) may be lost due to unflushed filesystem buffers

---

## 1.5 SSH Management

**File:** `internal/isolation/ssh.go`

**Tasks:**
1. Implement `SSHClient` struct using `golang.org/x/crypto/ssh`
2. Methods:
   - `Connect(host, user, keyPath)` - establish connection (key-based auth)
   - `ConnectPassword(host, user, password)` - password auth (initial setup)
   - `Run(command)` - execute command
   - `Shell()` - interactive shell via tmux-wrapper.sh
   - `CopyFiles(localPaths, remotePath)` - SCP equivalent
   - `Close()` - close connection
3. Connection retry logic (VM may be booting, up to 60s)
4. Key setup automation (generate ed25519, copy to VM)

**Key learnings from Phase 0 (ADR-002):**
- SSH key auth preferred after initial setup (password for bootstrap only)
- Default credentials: admin/admin
- SSH options for automation: `StrictHostKeyChecking=no`, `UserKnownHostsFile=/dev/null`, `ConnectTimeout=2`, `BatchMode=yes`
- tmux sessions via `~/scripts/tmux-wrapper.sh new-session -A -s calf`
- TERM handling: never set TERM explicitly in command environment (opencode hangs). Use tmux-wrapper.sh which sets TERM in script environment
- Helper script deployment: copy vm-setup.sh, vm-auth.sh, vm-first-run.sh, tmux-wrapper.sh to ~/scripts/
- Must check scp exit codes for all file copy operations

**Conditional tmux auto-restore (from Phase 0.11):**
- Check `~/.calf-first-run` flag before starting tmux
- If flag exists: use `tmux new-session -s calf` (fresh session, no auto-restore)
- If flag absent: use `tmux new-session -A -s calf` (attach existing or create with auto-restore)
- This prevents vm-auth authentication screen from appearing inside restored tmux session on first boot
- Flag check must happen in `calf isolation start`, `calf isolation init`, and `calf isolation restart`

---

## 1.6 CLI Commands (Cobra)

**File:** `cmd/calf/main.go` + `cmd/calf/isolation.go`

**Tasks:**
1. Root command `calf`
2. Subcommand group `calf isolation` (alias: `calf iso`)
3. Implement commands (mapped from calf-bootstrap per ADR-002):

| Command | Maps from | Description |
|---------|-----------|-------------|
| `calf isolation init [--proxy auto\|on\|off] [--yes]` | `calf-bootstrap --init` | Full VM creation and setup |
| `calf isolation start [--headless]` | `calf-bootstrap --run` | Start VM and SSH in with tmux |
| `calf isolation stop [--force]` | `calf-bootstrap --stop` | Stop calf-dev |
| `calf isolation restart` | `calf-bootstrap --restart` | Restart VM and reconnect |
| `calf isolation gui` | `calf-bootstrap --gui` | Launch with VNC experimental mode |
| `calf isolation ssh [command]` | Direct SSH | Run command or interactive shell |
| `calf isolation status` | `calf-bootstrap --status` | Show VM state, IP, size, and context-appropriate commands |
| `calf isolation destroy` | N/A | Delete VMs with safety checks |
| `calf isolation snapshot list` | `calf-bootstrap -S list` | List snapshots with sizes |
| `calf isolation snapshot create <name>` | `calf-bootstrap -S create` | Create snapshot |
| `calf isolation snapshot restore <name>` | `calf-bootstrap -S restore` | Restore snapshot (git safety) |
| `calf isolation snapshot delete <names...> [--force]` | `calf-bootstrap -S delete` | Delete snapshots |
| `calf isolation rollback` | N/A | Restore to session start |

4. Global flags: `--yes` / `-y` (skip confirmations), `--proxy auto|on|off`, `--clean` (force full script deployment)
5. Isolation flags for `init` and `start`/`gui`:
   - `--no-network` — enable network isolation (softnet + host-side pf SMB block)
   - `--safe-mode` — enable both `--no-network` and `--no-mount`
   - `--no-smb-block` — disable SMB blocking (testing only)
   - `--clear-smb-block` — emergency: flush stuck pf rules
   - `--remove-smb-permissions` — remove `/etc/sudoers.d/calf-pfctl`

**Key learnings for Go implementation (from calf-bootstrap pf work):**
- `setup_smb_block_permissions()` must run BEFORE VM starts — eliminates the security window where VM runs with no block; NOPASSWD allows password-free operation after one-time setup
- NOPASSWD must cover ALL pfctl operations on the anchor: load (`-f -`), flush (`-F all`), show-rules (`-sr`). Covering only flush/show-rules means the load still prompts — the most common operation
- Use `sudo -n pfctl -a <anchor> -sr` to test NOPASSWD idempotency — avoids `sudo grep` on a root-owned file (which prompts even for reads)
- `pfctl -e` is always a no-op on macOS 10.15+ (pf always enabled) and must NOT be called — NOPASSWD commands don't cache the sudo timestamp, so the following non-NOPASSWD `pfctl -e` would prompt unexpectedly
- pf anchor `com.apple/calf.smb-block` fits under the existing `anchor "com.apple/*"` wildcard — no `/etc/pf.conf` modifications needed
- pf rules are NOT automatically cleaned up when the loading process exits (confirmed via xnu source `bsd/net/pf_ioctl.c`) — cleanup is always manual; for `--gui`, use a background watcher
- `disown $! 2>/dev/null || true` required in `set -e` zsh scripts — job control is not fully active so `disown $!` may fail with "job not found", triggering exit
- Background watcher I/O must be redirected away from terminal: `</dev/null >>"$LOG" 2>&1 &` — writing to terminal from a background process while user is at prompt corrupts the shell display
- `~/.calf-vm-no-network` is a HOST-ONLY marker file (not inside VM). Inside VM, isolation config is in `~/.calf-vm-config` (`NO_NETWORK=true`). Both are needed.
- Go `calf isolation gui` will need the Go helper daemon (see 5a below) since it can't use shell `disown`/background subshells reliably for cleanup

---

## 1.7 Git Safety Checks

**File:** `internal/isolation/safety.go`

**Tasks:**
1. Implement reusable `CheckGitChanges(sshClient)` function
2. Scan VM directories for uncommitted/unpushed changes:
   - `~/workspace`, `~/projects`, `~/repos`, `~/code`
   - `~` (home directory, depth 2 only)
3. Display warnings with affected repository paths
4. Prompt for confirmation before destructive operations
5. Integration points:
   - `calf isolation init` (before deleting existing VMs)
   - `calf isolation snapshot restore` (before replacing calf-dev, skip if calf-dev doesn't exist)
   - `calf isolation snapshot delete` (before deleting, skip with `--force`)
   - `calf isolation destroy` (before deleting)

**Key learnings from Phase 0 (ADR-002):**
- Start VM if not running to perform check via SSH, stop after if it wasn't running before
- Unpushed commit detection requires upstream tracking (`git branch -u origin/main`)
- Use `git status --porcelain` for uncommitted changes
- Use `git log @{u}.. --oneline` for unpushed commits
- `--force` flag on delete skips git checks entirely (for unresponsive VMs)
- Single confirmation prompt per operation (avoid double/triple prompts)

**Worktree awareness (required — see Critical Issue 4b for bash reference implementation):**
- `find -name ".git" -type d` does NOT find linked worktree `.git` entries (they are files)
- `git status --porcelain` in a main checkout does NOT report changes in linked worktrees
- For each repo found, call `git worktree list` to enumerate linked worktrees and run both checks (uncommitted + unpushed) in each linked worktree
- Repos with no worktrees: no behaviour change

---

## 1.8 Proxy Management

**File:** `internal/isolation/proxy.go`

**Tasks:**
1. Implement proxy mode management (auto, on, off)
2. Network connectivity testing (`curl -s --connect-timeout 2 -I https://github.com`)
3. Bootstrap SOCKS proxy (SSH -D 1080) for init phase before sshuttle installed
4. sshuttle transparent proxy lifecycle (start, stop, restart, status)
5. VM→Host SSH key setup for proxy
6. Proxy auto-start on shell initialization

**Key learnings from Phase 0 (ADR-002):**
- Bootstrap proxy solves chicken-and-egg: need network to install sshuttle
- Use `socks5h://` (not `socks5://`) for DNS resolution through proxy
- sshuttle excludes: `-x HOST_GATEWAY/32 -x 192.168.64.0/24`
- Host requirements: SSH server enabled, Python installed
- Auto-start errors suppressed to avoid spamming shell startup
- Proxy config stored in `~/.calf-proxy-config`
- Proxy logs in `~/.calf-proxy.log`, PID in `~/.calf-proxy.pid`

---

## 1.9 VM Lifecycle Automation

**Tasks:**
1. Keychain auto-unlock setup during init
   - Save VM password to `~/.calf-vm-config` (mode 600)
   - Configure `.zshrc` keychain unlock block
   - `CALF_SESSION_INITIALIZED` guard to prevent re-execution on logout cancel
2. First-run flag system
   - `~/.calf-auth-needed` flag triggers vm-auth.sh during init
   - `~/.calf-first-run` flag triggers vm-first-run.sh after restore
   - Call `sync` after creating flag files (filesystem sync timing)
3. Logout git status check
   - Configure `~/.zlogout` to scan ~/code for uncommitted/unpushed changes
   - Cancel logout starts new login shell with session flag preserved
4. Tmux session persistence on logout
   - Add tmux session save to `~/.zlogout` hook (before git status check)
   - **Must gate save on first-run flag** (BUG-005): only save if `~/.calf-first-run` does NOT exist
   - Prevents capturing auth screens during --init into session data
   - Run `tmux run-shell ~/.tmux/plugins/tmux-resurrect/scripts/save.sh` on logout
   - Ensures session state is captured even between auto-save intervals
5. Filesystem sync before VM stop (BUG-009)
   - Call `sync && sleep 2` via SSH after vm-auth or any critical write operation completes
   - Must happen before VM stop or snapshot creation
   - Prevents data loss from unflushed filesystem buffers
   - Mirrors existing sync pattern used for flag files
6. Delay before VM stop for session save (BUG-005)
   - Add 10-second delay before `tart stop` in stop/restart/gui operations
   - Allows detach hook saves to complete before VM shutdown
   - **Do NOT add explicit tmux saves** in stop/restart/gui — detach hook already saves
   - Explicit background saves can corrupt save files if VM stops mid-write
7. VM detection setup
   - Create `~/.calf-vm-info` with VM metadata
   - Add `CALF_VM=true` to `.zshrc`
   - Install helper functions (`is-calf-vm`, `calf-vm-info`)
8. Tart cache sharing setup
   - Create symlink `~/.tart/cache -> /Volumes/My Shared Files/tart-cache`
   - Idempotent (safe to run multiple times)
   - Graceful degradation if sharing not available

**Key learnings from Phase 0 (ADR-002):**
- First-run flag reliability: set in calf-dev (running, known IP) → clone to calf-init → remove from calf-dev
- Session guard (`CALF_SESSION_INITIALIZED`) persists through `exec zsh -l` (environment variable)
- vm-first-run.sh only checks for updates, doesn't auto-pull (avoids surprise merge conflicts)

**Key learnings from Phase 0.11 (Tmux Session Persistence):**
- Session name must be `calf` (not `calf-dev`) for `calf isolation` commands
- Auto-restore on tmux start via tmux-continuum (no manual intervention needed)
- Auto-save every 15 minutes via tmux-continuum, plus manual save on logout
- Pane contents (scrollback) preserved with 50,000 line limit
- Resurrect data stored in `~/.local/share/tmux/resurrect/` (tmux-resurrect default) — survives VM restarts and snapshot/restore
- Manual save (`Ctrl+b Ctrl+s`) runs silently without confirmation message
- Manual restore keybinding: `Ctrl+b Ctrl+r`

---

## 1.10 Helper Script Deployment

**Tasks:**
1. Deploy helper scripts to VM `~/scripts/` directory (idempotent)
2. Scripts to deploy:
   - `vm-setup.sh` - Tool installation and configuration (calls vm-tmux-resurrect.sh during --init)
     - Must create `alias agent='cursor-agent'` in ~/.zshrc (idempotent check) (BUG-008)
   - `vm-auth.sh` - Interactive agent authentication **ONLY** (no state management) (BUG-008)
     - Creates agent alias directly if `agent` command missing (not by sourcing ~/.zshrc — avoids side effects like early tmux-resurrect loading)
     - Does NOT manage first-run flag or tmux session state
   - `vm-first-run.sh` - Post-restore initialization (BUG-005/BUG-008 architecture)
     - Checks git repositories for updates
     - Loads TPM to enable tmux session persistence: `~/.tmux/plugins/tpm/tpm`
     - Removes `~/.calf-first-run` flag AFTER tmux history is enabled
     - Session persistence only starts on first user login, never during --init
   - `tmux-wrapper.sh` - TERM compatibility wrapper for tmux
   - `vm-tmux-resurrect.sh` - Tmux session persistence setup (Phase 0.11)
3. Deploy comprehensive tmux.conf with:
   - **PATH environment for plugin scripts** (required for tmux-resurrect - see note below)
   - **Conditional TPM loading based on first-run flag** (prevents auth screen capture)
   - Session persistence via tmux-resurrect and tmux-continuum plugins
   - Auto-save every 15 minutes, auto-restore on tmux start
   - Pane contents (scrollback) capture with 50,000 line limit
   - Client-detached hook for save on `Ctrl+b d` — **must gate on first-run flag** (BUG-005): `if [ ! -f ~/.calf-first-run ]; then save; fi`
   - Keybindings: `Ctrl+b R` reload config, `Ctrl+b r` resize pane to 67%
   - Split bindings: `Ctrl+b |` horizontal, `Ctrl+b -` vertical
4. Add `~/scripts` to PATH in `.zshrc`
5. Verify deployment after SCP (check exit codes)
6. **Implement checksum-based deployment optimization:**
   - Compare MD5 checksums between host and VM scripts before copying
   - Only copy scripts that are new or changed (skip unchanged)
   - Visual feedback: `↻` (unchanged/skipped), `↑` (updated), `+` (new)
   - `--clean` flag forces full deployment (bypasses checksum optimization)
   - Saves ~2 seconds per `--run`/`--restart` when scripts are current

**Key learnings from Phase 0.11 (Tmux Session Persistence):**
- vm-tmux-resurrect.sh installs tmux-resurrect and tmux-continuum plugins via TPM
- Must be integrated into vm-setup.sh `--init` path for fresh installations
- tmux.conf is the single source for all tmux configuration and keybindings
- Session name `calf` used by `tmux-wrapper.sh new-session -A -s calf`
- Session data stored in `~/.local/share/tmux/resurrect/` (tmux-resurrect default location)
- Manual save (`Ctrl+b Ctrl+s`) runs silently; manual restore (`Ctrl+b Ctrl+r`)
- **Mouse mode must be enabled by default** (`set -g mouse on`) for tmux right-click menu functionality
  - `mouse on` = tmux context menu (Swap, Kill, Respawn, Mark, Rename, etc.)
  - `mouse off` = terminal app menu (Copy, Paste, Split, etc.)
  - See BUG-004 for regression details

**Critical: PATH requirement for tmux-resurrect (BUG-005):**
- tmux-resurrect scripts run via `tmux run-shell` which has minimal PATH (`/usr/bin:/bin:/usr/sbin:/sbin`)
- This PATH doesn't include Homebrew directories where tmux is installed (`/opt/homebrew/bin`)
- Without proper PATH, save files contain only "state state state" instead of actual session data
- **Solution:** Add `set-environment -g PATH` in tmux.conf to include Homebrew paths
- Example: `set-environment -g PATH "/opt/homebrew/bin:/opt/homebrew/sbin:/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin"`

**TPM installation reliability (BUG-006):**
- TPM (Tmux Plugin Manager) installation can fail due to network issues
- **Solution:** Implement retry logic with 3 attempts and 5-second delay between retries
- Cache cloned TPM repository (`~/.tmux/plugins/tpm`) for reuse
- Clear error messages if all attempts fail
- Explicit cleanup on init failure (delete calf-dev to allow clean retry)

**Conditional TPM loading for first-run:**
- TPM must NOT load during first-run (while `~/.calf-first-run` flag exists)
- Prevents tmux-resurrect from capturing the vm-auth authentication screen
- Use conditional in tmux.conf: `if-shell '[ ! -f ~/.calf-first-run ]' 'run ~/.tmux/plugins/tpm/tpm'`
- After first-run completes and flag is removed, session persistence works normally

---

## 1.11 Configuration Enhancements (Future)

**Tasks (deferred to future phases):**
1. **Interactive config fixing** - When validation fails, prompt user to fix config interactively
   - Detect invalid values and offer to correct them on the spot
   - Show valid ranges and let user enter new value
   - Write corrected config back to file
2. **Environment variable overrides** - Support env vars overriding config values
   - Example: `CALF_VM_CPU=8 calf isolation init` overrides config CPU setting
   - Follow 12-factor app pattern for configuration hierarchy
   - Priority: env vars > per-VM config > global config > defaults
3. **Config validation command** - `calf config validate` to check config without running
   - Parse and validate config files
   - Report all errors (don't stop at first error)
   - Exit 0 if valid, non-zero if invalid
4. **Config schema migration** - Strategy for handling config version changes
   - Detect old config versions and migrate automatically
   - Backup old config before migration
   - Clear migration messages to user
5. **Default values documentation** - `calf config show --defaults` flag
   - Display hard-coded default values
   - Help users understand what they get without a config file
   - Show which values are from defaults vs. config files
6. **Tmux session save feedback** - Improve discoverability of tmux-resurrect functionality
   - Current: `Ctrl+b Ctrl+s` saves silently with no confirmation
   - Enhancement: Display brief confirmation message when session is saved
   - Consider: `tmux display-message "Session saved to ~/.local/share/tmux/resurrect/"` after save
   - Also consider: Status bar indicator showing last save time
   - Trade-off: More feedback vs. silent operation preference

**Notes:**
- These enhancements improve UX but are not critical for Phase 1 functionality
- Can be prioritized based on user feedback after Phase 1 completion
- Interactive fixing and env var overrides are most valuable for daily use
- Tmux save feedback is low priority (auto-save works, manual save is advanced feature)

---

## Testing Requirements

**Unit Tests:**
- Configuration parsing and validation
- Tart command generation
- SSH command building
- Git safety check logic
- Proxy mode detection

**Integration Tests:**
- VM lifecycle (create, start, stop, delete)
- Snapshot operations (create, restore, list, delete)
- SSH connectivity and command execution
- SCP file transfer
- Git change detection in VM

**Key testing lessons from Phase 0 (ADR-002):**
- BSD awk incompatibility: test on macOS, not just Linux
- `shift || true` errors in zsh: use `[[ $# -gt 0 ]] && shift`
- Double/triple confirmation prompts: ensure single prompt per operation
- scp error handling: always check exit codes
- Filesystem sync: flag files need `sync` before VM reboot
- macOS `timeout` command unavailable: use built-in timeouts

**Additional testing lessons from Phase 0.11 (ADR-002):**
- **tmux-resurrect PATH:** Scripts run via `tmux run-shell` have minimal PATH - must set `set-environment -g PATH` in tmux.conf to include Homebrew directories
- **TPM installation network failures:** Use retry logic (3 attempts, 5s delay) with caching for reliability; clear error messages on failure
- **First-run flag and session restore:** Must check flag before tmux start; use `-A` flag only when flag absent to prevent auth screen in restored session
- **Arithmetic in set -e:** `((counter++))` fails with `set -e` - use `counter=$((counter + 1))` instead
- **Tmux capturing auth screen:** Conditionally load TPM only after first-run completes; clear session data after auth if needed

**Key testing lessons from SMB pf blocking (Critical Issue #5):**
- **Security window testing:** Verify that any privilege setup (sudoers, pf rules) completes BEFORE the VM is running. Any password prompt AFTER a VM has an IP = security window. Test by checking which message appears first in output.
- **NOPASSWD scope:** Test ALL operations that need sudo, not just the obvious ones. Load (`-f -`), flush (`-F all`), show (`-sr`) may all be needed; missing any one causes a password prompt. Use `sudo -n <cmd>` to verify each operation is truly passwordless.
- **pf rule verification from inside VM:** `nc -w3 -z <host-ip> 445` (TCP) and `nc -w3 -z <host-ip> 139` are the correct tools. `smbutil view` can be used but requires credential prompts. `nc` exit code 1 = blocked.
- **Internet not broken by port blocking:** Always test `curl -s https://api.github.com` after enabling pf rules to confirm HTTPS (port 443) is not affected.
- **Background watcher terminal isolation:** Any background process that completes while the user is at a prompt must have ALL I/O redirected: `</dev/null >>"$LOG" 2>&1 &`. Even a single `echo` to stdout from a background process corrupts the terminal display.
- **sudo NOPASSWD + timestamp caching:** NOPASSWD commands do NOT update the sudo timestamp. A sequence of `sudo NOPASSWD-cmd` followed by `sudo non-NOPASSWD-cmd` will always prompt for the second command, even milliseconds later. Never assume a prior NOPASSWD command caches credentials.
- **pf cleanup is NOT automatic:** pf rules survive process exit. Test by killing the script mid-run and checking `sudo pfctl -a com.apple/calf.smb-block -sr` — rules will still be there. Cleanup must always be explicit.
- **Smoke test for isolated VMs:** Test matrix — (1) SMB blocked, (2) internet works, (3) no host mounts in `/Volumes/`, (4) `~/.calf-vm-config` shows isolation mode, (5) tools installed. All five must pass.

**Key testing lessons from post-cache-integration bugs (ADR-003 § Bug Fixes):**
- **Agent alias (BUG-008):** Never source ~/.zshrc in scripts — causes side effects (tmux-resurrect loading early). Create aliases directly instead
- **Filesystem sync timing (BUG-009):** Always call `sync && sleep 2` via SSH after operations that write data, before VM stop or snapshot creation. Silent data loss occurs without this
- **Save hook gating (BUG-005):** All tmux save triggers (detach hook, .zlogout, auto-save) must check `~/.calf-first-run` flag. Ungated saves capture auth screens during --init
- **No explicit saves before VM stop (BUG-005):** Rely on detach hook saves only. Explicit background saves (`tmux run-shell -b`) can be killed mid-write by VM stop, corrupting save files
- **Delay before VM stop (BUG-005):** 10-second delay before `tart stop` lets detach hook saves complete. Without delay, save may be killed mid-write
- **Script architecture (BUG-008):** vm-auth.sh = authentication only; vm-first-run.sh = state management (TPM loading, flag removal). Mixing concerns caused cascading issues
