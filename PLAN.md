# CALF Implementation Plan

> 🎯 **THIS IS THE SINGLE SOURCE OF TRUTH FOR PROJECT STATUS AND TODOS**
>
> - Phase completion is determined by checkboxes in phase TODO files
> - All TODOs must be tracked in phase files (code TODOs should reference this file)
> - Operational guide: [ADR-002](docs/adr/ADR-002-tart-vm-operational-guide.md)
> - Cache architecture: [ADR-003](docs/adr/ADR-003-package-download-caching.md) (symlinks) → [ADR-004](docs/adr/ADR-004-cache-mount-architecture.md) (direct mounts)
> - Original design: [ADR-001](docs/adr/ADR-001-cal-isolation.md) (immutable)

## Current Status

**Phase 0 (Bootstrap): Complete** - Core functionality documented in [ADR-002](docs/adr/ADR-002-tart-vm-operational-guide.md). Tmux session persistence (section 0.11) implemented and tested. Post-integration bug fixes (BUG-005 edge cases, BUG-008, BUG-009) all resolved — see [ADR-003 § Bug Fixes Post-Integration](docs/adr/ADR-003-package-download-caching.md#bug-fixes-post-integration). All Phase 0 tasks complete. No active bugs remaining.

Three-tier VM architecture (calf-clean, calf-dev, calf-init), automated setup via calf-bootstrap, transparent proxy for network reliability, comprehensive git safety checks, VM detection capabilities, first-run/logout git automation, Tart cache sharing for nested VMs, Ghostty terminal emulator, and tmux session persistence are all operational.

**Active Work:** Phase 1 (CLI Foundation) in progress. Project scaffolding (PR #3), Configuration Management (PR #4), Tart Wrapper (PR #5), Homebrew Cache (PR #6), npm Cache (PR #7), Go Modules Cache (PR #8), Git Cache with complete cache integration (PR #9), and Cache Clear command (PR #10) all merged. **Package download caching (1.1) complete** — Homebrew, npm, Go, and Git caches all implemented with full bootstrap integration and cache clear command. **Cache Mount Architecture (Critical Issue #3) complete** — Replaced fragile symlinks with robust direct virtio-fs mounts (LaunchDaemon + self-healing), macOS-compatible mount verification, cannot be deleted with `rm -rf`. Cache architecture evolution: [ADR-003](docs/adr/ADR-003-package-download-caching.md) (symlinks) → [ADR-004](docs/adr/ADR-004-cache-mount-architecture.md) (direct mounts). **Security (Critical Issue #5) complete** — `--no-network` SMB bypass fixed via host-side pf anchor (`com.apple/calf.smb-block`). Standard Homebrew tart/softnet; no patched binaries. NOPASSWD sudoers drop-in for passwordless operation; background watcher for `--gui` cleanup. Full smoke test verified: SMB blocked, internet works, marker file persists mode across sessions. See [PLAN-PHASE-01-DONE.md § Critical Issue #5](docs/PLAN-PHASE-01-DONE.md). **TDD Remediation complete** — All items (Pre-work, 0–10) done. **BUG-012 complete** — CacheManager now has injectable `io.Writer`; all warning output testable. **TDD Remediation 2 complete** — All 8 items done. 203 tests pass, `staticcheck` clean. BUG-011 and BUG-012 resolved. **TDD Remediation 3 complete** — All 10 items done (TestVMStateString rename; TestCloneWhenTart* t.Run wrapping; Item 3 no-change; ensureInstalled moved to public methods; TartClientOption functional options added; createTestClient rewritten using options; TestIP pollTimeout direct write replaced with WithPollTimeout; TestCloneWhenTart* rewritten using functional options; makeInstallingRunCommand deleted; newRootCmd/newConfigCmd factories replacing global vars and init()). 208 tests pass, staticcheck clean. **TDD Remediation 4 in progress** — Item 1 done (TDD-R3 doc drift closed). Items 2–5 remaining: Arrange/Act/Assert markers for config_test.go and tart_test.go; table-driven test rewrites in tart_test.go; cache_test.go t.TempDir() + per-subtest isolation. Next: Critical Issue #4 (bootstrap init fall-through), or begin Snapshot Management (1.4) / SSH Management (1.5), or Critical Issue #6 (repository rename).

---

## Phases Overview

| Phase | Name | Status | Outstanding TODOs | Detail Files |
|-------|------|--------|-------------------|--------------|
| 0 | Bootstrap | Complete | 0 | [TODO](docs/PLAN-PHASE-00-TODO.md) • [DONE](docs/PLAN-PHASE-00-DONE.md) |
| 1 | CLI Foundation | In Progress | ~44 | [TODO](docs/PLAN-PHASE-01-TODO.md) • [DONE](docs/PLAN-PHASE-01-DONE.md) |
| 2 | Agent Integration & UX | Not Started | ~20 | [TODO](docs/PLAN-PHASE-02-TODO.md) • [DONE](docs/PLAN-PHASE-02-DONE.md) |
| 3 | GitHub Workflow | Not Started | ~16 | [TODO](docs/PLAN-PHASE-03-TODO.md) • [DONE](docs/PLAN-PHASE-03-DONE.md) |
| 4 | Environment Plugin System | Not Started | ~25 | [TODO](docs/PLAN-PHASE-04-TODO.md) • [DONE](docs/PLAN-PHASE-04-DONE.md) |
| 5 | TUI & Polish | Not Started | ~15 | [TODO](docs/PLAN-PHASE-05-TODO.md) • [DONE](docs/PLAN-PHASE-05-DONE.md) |

### Phase Goals

**Phase 0 (Bootstrap):** Manual VM setup and bootstrap automation
**Deliverable:** Three-tier VM architecture with automated setup and safety features

**Phase 1 (CLI Foundation):** Replace manual Tart commands with `calf isolation` CLI
**Deliverable:** Working `calf isolation` CLI that wraps Tart operations

**Phase 2 (Agent Integration & UX):** Seamless agent launching with safety UI
**Deliverable:** `calf isolation run <workspace>` launches agent with full UX

**Phase 3 (GitHub Workflow):** Complete git workflow from VM
**Deliverable:** Clone → Edit → Commit → PR workflow working

**Phase 4 (Environment Plugin System):** Pluggable development environments
**Deliverable:** Multi-platform development with pluggable environments

**Phase 5 (TUI & Polish):** Full terminal UI experience
**Deliverable:** Complete TUI for CALF

---

## Recommended Order of Implementation

### Immediate (Phase 0)
1. Complete manual VM setup following bootstrap guide
2. Verify all agents work
3. Create safety snapshot
4. Begin using for development
5. Complete outstanding Phase 0 enhancements

### Short-term (Phase 1)
1. **Package download caching (1.1) - HIGHEST PRIORITY**
2. Project scaffolding (1.1) - DONE
3. Configuration management (1.2) - DONE
4. Tart wrapper (1.3)
5. Snapshot management (1.4)
6. SSH management (1.5)
7. CLI commands (1.6)
8. Git safety checks (1.7)
9. Proxy management (1.8)
10. VM lifecycle automation (1.9)
11. Helper script deployment (1.10)

### Medium-term (Phases 2-3)
1. Status banner (2.2)
2. Launch confirmation (2.3)
3. Agent management (2.4)
4. GitHub auth (3.1)
5. Clone/commit/PR (3.2-3.4)
6. TUI framework (2.1)
7. SSH tunnel with overlay (2.5)

### Long-term (Phases 4-5)
1. Plugin system (4.1-4.4)
2. Core plugins (4.3)
3. Templates (4.5)
4. Full TUI (5.1-5.4)

---

## Testing Strategy

### Unit Tests
- Configuration parsing
- Git command generation
- SSH command building

### Integration Tests
- VM lifecycle
- Snapshot operations
- SSH connectivity

### End-to-End Tests
- Full workflow: init → start → clone → run → commit → pr
- Agent installation verification
- Rollback verification

### Manual Testing
- Destructive operation containment
- Network interruption recovery
- VM crash recovery

---

## Dependencies to Install (for development)

```bash
# On host (for building CAL)
brew install go
go install honnef.co/go/tools/cmd/staticcheck@latest

# Verify
go version
staticcheck --version
```

---

## Next Action

**Phase 0 is complete.** All operational learnings captured in [ADR-002](docs/adr/ADR-002-tart-vm-operational-guide.md).

**Begin Phase 1 (CLI Foundation):**
1. Project scaffolding and Go module setup (1.1)
2. Configuration management with VM config file awareness (1.2)
3. Tart wrapper with cache sharing, VNC experimental, BSD awk compat (1.3)
4. Build out remaining subsystems per ADR-002 operational requirements

Read [PLAN-PHASE-01-TODO.md](docs/PLAN-PHASE-01-TODO.md) for detailed tasks.
See [ADR-002 § Phase 1 Readiness](docs/adr/ADR-002-tart-vm-operational-guide.md) for command mapping and operational requirements.
