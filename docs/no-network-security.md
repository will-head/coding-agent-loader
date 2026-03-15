# No-Network Mode Security Investigation (SMB Bypass)

Updated: 2026-02-10 (resolved via host-side pf anchor)
Workflow: Interactive (1)
Branch: main
Status: RESOLVED — SMB blocked via pf anchor (com.apple/calf.smb-block)

---

## Critical Security Issue

User can still access host Mac from VM using username/password.

Evidence observed:
- VM Finder sidebar shows "Network"
- Host network share is visible
- VM can browse network shares despite `--net-softnet --net-softnet-block=224.0.0.0/4`

### What We Tested vs Reality

| Test | Result | Actual Security |
|------|--------|-----------------|
| mDNS discovery blocked | FAIL (leaks through) | Host services discoverable |
| SMB port 445 reachable | OK (open, auth fails for guest/admin:admin) | WARNING: real user credentials work |
| SMB mount with VM creds | OK (blocked) | OK (VM creds don't work on host) |
| SMB mount with HOST creds | NOT TESTED | FAIL (works - security bypass) |

### Root Cause

Tests only checked:
1. Guest access (no credentials) - blocked
2. VM credentials (admin:admin) - blocked

Host user credentials were not tested.

The security model assumed:
"AI coding agents don't know host credentials, so even if port 445 is open, they can't authenticate"

This assumption is false when:
- User runs AI agent on their main Mac (not in VM)
- AI agent can access user's keychain/password manager or prompt user
- User accidentally provides credentials when prompted
- Social engineering attacks

---

## What Actually Works

All tests pass (14/14) but test the wrong threat model.

Blocked:
- Direct IP access to local network (192.168.x.x, 10.x.x.x)
- Guest SMB access
- SMB access using VM credentials
- No host filesystem mounts

Works as intended:
- Internet access via NAT
- DNS resolution
- HTTPS to github.com
- SSH from host to VM

Not blocked (security hole):
- Network browsing in Finder
- SMB connection using host credentials
- Mounting host shares with real username/password

---

## Implementation Status (as of 2026-02-08)

### Completed Features

1) `--no-network` flag
   - Uses `--net-softnet --net-softnet-block=224.0.0.0/4`
   - Blocks direct local network IP access
   - Allows internet via NAT
   - One-way init option (cannot be changed after VM creation)
   - Host marker: `~/.calf-vm-no-network`
   - VM config: `NO_NETWORK=true`

2) `--safe-mode` flag
   - Combines `--no-mount` + `--no-network`
   - Single command for maximum isolation

3) Softnet dependency management
   - Auto-detects installation
   - Offers to install via Homebrew
   - Checks and sets SUID bit
   - Interactive setup during init

4) Status display
   - Shows isolation mode in `--status`

5) Config file persistence
   - `vm-setup.sh` now updates in place (doesn't overwrite)
   - Preserves `NO_MOUNT` and `NO_NETWORK` settings

### Code Changes

Modified:
- `scripts/calf-bootstrap` - `--no-network`, `--safe-mode`, softnet integration
- `scripts/vm-setup.sh` - config update-not-overwrite

New test scripts:
- `scripts/test-safe-mode.sh` (v3.0) - smoke tests (14 tests, all pass)
- `scripts/test-softnet-block.sh` - softnet validation
- `scripts/test-no-network.sh` - feature tests
- `scripts/test-smb-mount.sh` - mount bypass tests
- `scripts/test-smb-firewall.sh` - obsolete (pf approach failed)

### Test Results

Smoke test: 14/14 PASS
- VM prerequisites
- Host marker files exist
- Status shows safe mode
- SSH connectivity
- Internet access (DNS, HTTPS)
- Local network blocked (IP-level)
- No host mounts
- VM config correct (NO_MOUNT=true, NO_NETWORK=true)
- Persistence across restart

Security test: FAIL
- SMB access with host credentials still works

---

## Technical Analysis

### Why Softnet Doesn't Block SMB Completely

1) Softnet blocks at IP/network layer:
   - Blocks connections to 192.168.x.x, 10.x.x.x (local network IPs)
   - Does NOT block connections to gateway (192.168.64.1)

2) Gateway is the SMB server:
   - macOS host runs SMB service on gateway IP (192.168.64.1)
   - Gateway must be allowed for NAT/internet
   - Softnet has no port-level filtering (only IP CIDR)

3) mDNS/Bonjour leaks through:
   - Multicast (224.0.0.0/4) block is unreliable
   - May operate at L2 (Ethernet) before IP filtering
   - Finder still discovers network shares

### Why We Can't Block Gateway

From testing:
- Blocking gateway (192.168.64.1) breaks SSH
- Host IS the gateway on vmnet
- SSH response packets (VM -> host) are blocked
- No stateful/connection-tracking in softnet

### Alternative Approaches Already Tried

1) Host-level pf firewall - Softnet bypasses kernel pf
2) VM-level pf firewall - VM has sudo, can disable
3) Softnet port filtering - Not supported (IP CIDR only)
4) Block gateway IP - Breaks SSH (host is gateway)
5) Block multicast - Unreliable, doesn't prevent SMB auth

---

## Options Going Forward (original evaluation)

### Option 1: Accept Current State (Recommended for Phase 1)

Rationale:
- AI coding agents running inside the VM don't have host credentials
- Threat model is "untrusted code in VM accessing host"
- Code in VM can't mount host shares without credentials
- For AI agents on host Mac, VM isolation doesn't apply anyway

Residual risk:
- Malicious code could prompt user for credentials
- Social engineering: "Enter your password to access project files"
- If user enters host password, shares can be mounted

Mitigation:
- Document limitation clearly
- Educate users not to enter host passwords in VM
- Consider acceptable for initial release

Action items:
- Update documentation with security limitations
- Add warning during `--init` about credential prompts
- Consider detection/warning if SMB mount attempted

### Option 2: Disable SMB on Host (User Responsibility)

Approach:
- Document how to disable File Sharing in System Settings
- Make it user's responsibility to disable SMB
- Check during init if SMB is enabled, warn user

Pros:
- Simple, effective
- User controls their own security posture

Cons:
- User may need File Sharing for other purposes
- Requires manual configuration
- Can be re-enabled accidentally

### Option 3: Block SMB in Guest OS (Fragile)

Approach:
- Add pf rules in VM to block outbound port 445, 139
- Runs during VM boot via LaunchDaemon

Pros:
- Would block SMB even with credentials

Cons:
- VM has sudo - malicious code can disable
- Not a reliable security boundary
- Adds complexity

### Option 4: Request Softnet Enhancement (Long-term)

Approach:
- File feature request with Cirrus Labs
- Request port-level filtering in softnet
- `--net-softnet-block-port=445,139`

Pros:
- Proper solution at right layer
- Benefits other softnet users

Cons:
- Requires upstream changes
- Timeline uncertain
- May not be accepted

### Option 5: Shelve Feature (Not Recommended)

Remove `--no-network` entirely until a complete solution exists.

Cons:
- Wastes work already done
- Current implementation still provides value
- Perfect is enemy of good

---

## Recommended Path Forward (original)

Immediate:
1) Document the limitation in:
   - `--init` warning message
   - `--help` text
   - README / security docs
   - ADR documenting the security boundary
2) Update warning during init:
   - Clearly state SMB access possible with host credentials
   - Warn users not to enter host Mac password in VM
3) Add security note to status output

Short-term:
1) Add SMB detection - warn if user mounts network shares
2) Consider VM-level pf as defense-in-depth
3) File softnet feature request for port filtering

Long-term:
1) Evaluate if limitation is acceptable based on user feedback
2) Consider alternative VM platforms if this becomes critical
3) Implement additional monitoring/detection

---

## Files Status (original)

Ready to commit:
- `scripts/calf-bootstrap` - core functionality works
- `scripts/vm-setup.sh` - config update fix
- `scripts/test-safe-mode.sh` - all tests pass

Needs documentation updates:
- Update `--init` warning with SMB limitation
- Update `--help` text
- Update `--status` output
- Create ADR documenting security boundary

To delete:
- `scripts/test-smb-firewall.sh` - obsolete (pf approach)
- `scripts/test-no-network.sh.bak` - backup file

To keep (reference/validation):
- `scripts/test-softnet-block.sh` - validation tool
- `scripts/test-smb-mount.sh` - mount bypass tests
- `scripts/test-no-network.sh` - feature tests

---

## Security Boundary Definition (original)

What `--no-network` protects against:
- Untrusted code in VM accessing local network via direct IP
- Accidental network access by misconfigured tools
- Guest/anonymous SMB access
- Network access using VM credentials

What `--no-network` does NOT protect against:
- User entering host credentials when prompted
- Social engineering attacks
- Attacks that obtain host credentials via other means
- mDNS/Bonjour service discovery (partial/unreliable blocking)

Primary threat model:
AI coding agent executing untrusted code in VM should not be able to:
- Access host filesystem (solved by `--no-mount`)
- Access local network devices/servers by IP (solved by `--no-network`)
- Access host SMB shares without user intervention (solved - credentials required)

Out of scope:
- Protecting against user errors (entering passwords)
- Preventing all network discovery (mDNS leak accepted)
- Blocking SMB with valid credentials (requires port filtering)

---

## Next Steps (original)

User decision required:
Is the current security boundary acceptable for initial release?

If YES:
1) Update documentation with limitations
2) Add warnings to init/status
3) Commit changes
4) File softnet feature request
5) Move forward with other features

If NO:
1) Discuss alternative approaches
2) Consider shelving feature
3) Investigate other VM platforms

---

## Quick Reference (original)

Test command:
```
./scripts/test-safe-mode.sh
```

Init commands:
```
./calf-bootstrap --init --safe-mode --yes
./calf-bootstrap --init --no-network --yes
./calf-bootstrap --init --no-mount --yes
```

Check status:
```
./calf-bootstrap --status
```

Marker files:
- Host: `~/.calf-vm-no-mount`, `~/.calf-vm-no-network`
- VM: `~/.calf-vm-config` (contains `NO_MOUNT=true`, `NO_NETWORK=true`)

---

## Resolution (2026-02-10) — ✅ COMPLETED

The patched Softnet/Tart approach was abandoned. All compiled-from-source softnet versions failed VM initialization; only Homebrew softnet v0.18.0 works. See `docs/softnet-port-blocking-investigation.md` for full investigation.

**Solution: Host-side pf anchor** — Standard Homebrew tart/softnet only. Block SMB/NetBIOS on the HOST using macOS `pf` with a temporary named anchor. No patched binaries. No changes to `/etc/pf.conf`.

Architecture:
- Anchor `com.apple/calf.smb-block` under existing `anchor "com.apple/*"` wildcard
- Rules: block TCP 445/139 + UDP 137/138 from VM IP to any destination
- In-memory only — no disk changes; removed when VM stops
- One-time sudoers drop-in `/etc/sudoers.d/calf-pfctl` for NOPASSWD pfctl on this anchor
- `--gui` mode: background watcher polls `kill -0 $TART_PID` every 5s; removes rules when VM exits
- `setup_smb_block_permissions()` runs BEFORE VM starts — eliminates security window

Verified smoke tests (all pass):
- ✅ SMB TCP 445/139 blocked from inside VM
- ✅ Internet works (github.com HTTP 200)
- ✅ No host mounts in /Volumes/
- ✅ Rules removed on VM stop
- ✅ No spurious password prompts after one-time setup

See `docs/PLAN-PHASE-01-DONE.md § Critical Issue #5` for full implementation details.
