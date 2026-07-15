---
title: Security Notes
description: "MoAI-ADK v2.20.0-rc1 security hardening — CWE-732/214/345 mappings and self-audit procedures"
weight: 72
draft: false
tags: ["security", "cwe", "audit"]
---

An agentic harness is a system that hands execution authority to agents. The more authority a system delegates, the more the security of its credentials and update paths forms the floor of harness trust. This page summarizes the **user-visible security changes** introduced as of MoAI-ADK v2.20.0-rc1. Each item includes its CWE mapping, the changed behavior, and self-audit commands.

## Why — Why This Page Exists

`SPEC-V3R5-SECURITY-CRIT-001` (PR #1032, merge commit `03a2552a2`) corrected **3 P0 release-blocker security defects** found in code review between v2.14.0 and v2.20.0-rc1. This page codifies that correction, and the procedures for users to verify the new protections in their own environments, as official 4-locale guidance.

All three defects relate to the GLM integration + auto-update paths.

- **CWE-732 / CWE-552** — `.claude/settings.local.json` file mode enforced to `0o600` (owner-only read/write)
- **CWE-214** — `moai cg` tmux environment-variable injection goes via source-file instead of argv (GLM token no longer visible in argv)
- **CWE-345** — `moai update` checksum verification is mandatory (update refused on download failure)

Each item is locked by regression tests, blocking future regressions.

## CWE-732 — settings.local.json Permission Hardening {#cwe-732}

### What Changed

When `.claude/settings.local.json` is created or updated, its file permission is enforced to **`0o600`** (owner-only read/write). Previously it was created as `0o644` (owner read/write + group/world read), so on multi-user workstations other local users could read sensitive credentials such as `ANTHROPIC_AUTH_TOKEN`.

### Threat Model

- **Attacker**: a low-privilege local user on the same host
- **Attack surface**: the group/world read permission of `.claude/settings.local.json`
- **Leaked information**: GLM API token (`ANTHROPIC_AUTH_TOKEN`), OAuth refresh token, other `settings.Env` values
- **CWE mapping**: CWE-732 (Incorrect Permission Assignment for Critical Resource), CWE-552 (Files or Directories Accessible to External Parties)

### Implementation Locations

- `internal/hook/settings_io.go` — the `secureSettingsMode os.FileMode = 0o600` constant + the `writeSettingsSecure` helper
- `internal/hook/session_start.go` — all `settings.local.json` writers, including `ensureGLMCredentials`, `ensureClaudeEnvFile`
- `internal/hook/session_end.go` — GLM keys write-back path

### Self-Audit

Check the permission of your existing `settings.local.json`.

```bash
# Linux
stat -c '%a' .claude/settings.local.json
# Expected: 600

# macOS
stat -f '%A' .claude/settings.local.json
# Expected: 600
```

If the permission shows `644` or any looser value, MoAI-ADK automatically corrects it to `0o600` at the next session start. To correct it immediately:

```bash
chmod 0600 .claude/settings.local.json
```

### Impact (Trade-off)

Workflows that expect `group-readable` (the very rare scenario of a separate OS user reading the same project directory) may break. This trade-off is intentional; the security recovery is the clear priority.

## CWE-214 — Blocking tmux IPC Token argv Exposure {#cwe-214}

### What Changed

When `moai cg` (CG mode) injects the GLM token (`ANTHROPIC_AUTH_TOKEN`) into the tmux session environment, it uses the **source-file channel** (`tmux source-file <tmp>`) instead of the **argv channel** (`tmux set-environment <KEY> <VALUE>`). The token is no longer exposed in plaintext to `ps auxe`, `/proc/<pid>/cmdline`, auditd logs, sysmon traces, or crash dumps.

Since CG mode is the key savings mechanism of tokenomics (Claude leader + GLM workers, 60-70% savings), the security of its credential path is especially important.

### Implementation Flow

1. A temp file is created under `~/.moai/run/` with `mkstemp` (mode `0o600` automatic + explicit `chmod 0o600`)
2. A single `set-environment -t <session> <KEY> <VALUE>` line is written to the temp file
3. `tmux source-file <tmp>` has tmux read that file and inject it into the environment
4. Immediately after injection, the temp file is unlinked with `os.Remove`

Only the temp file path is exposed in argv; the token itself is not.

### Threat Model

- **Attacker**: a local user on the same host + system log collection (`ps`, `/proc`, auditd, sysmon)
- **Attack surface**: the argv channel of tmux env injection
- **Leaked information**: momentary visibility of the GLM API token
- **CWE mapping**: CWE-214 (Invocation of Process Using Visible Sensitive Information)

### Implementation Locations

- `internal/tmux/session.go` — the `InjectSensitiveEnv` method, `sensitiveTempDir = ".moai/run"`, `mkstemp` + `chmod 0o600` + `tmux source-file` + `os.Remove`
- `internal/tmux/errors.go` — the `ErrTmuxSensitiveInjectFailed` sentinel
- `internal/hook/glm_tmux.go` — in `ensureTmuxGLMEnv`, only `ANTHROPIC_AUTH_TOKEN` branches to the sensitive path (other non-sensitive values such as URLs and model names keep the existing argv path)

### Non-sensitive Values Stay on argv

Values that are not tokens — `CLAUDE_CONFIG_DIR` (a directory path), `ANTHROPIC_BASE_URL` (a URL), `ANTHROPIC_DEFAULT_*_MODEL` (model names) — keep the argv path. This is explicit intent and unrelated to token leakage risk.

### Behavior on Failure

If the source-file injection fails (disk full, tmux source-file failure, etc.), it does **not fall back to argv and leak** — it returns the `ErrTmuxSensitiveInjectFailed` sentinel error and aborts the injection itself. Not falling back to convenience on failure is the heart of this design.

### Self-Audit

Check whether the token is exposed in argv while CG mode is running.

```bash
# Inside a new tmux session after running moai cg
ps auxe | grep -i 'tmux set-environment.*ANTHROPIC_AUTH_TOKEN'
# Expected: 0 matches (token is not present in argv)
```

Check that the temp file is properly unlinked.

```bash
ls -la ~/.moai/run/ 2>/dev/null
# Expected: empty directory or no stale files
```

If residual files remain in `~/.moai/run/` after session end, they can be removed manually (not a security threat — the files were already subject to an unlink attempt).

### User Responsibility

The `~/.moai/.env.glm` source file must keep `0o600` permission in your environment. The `moai glm` command sets this automatically.

```bash
stat -c '%a' ~/.moai/.env.glm    # Linux: 600
stat -f '%A' ~/.moai/.env.glm    # macOS: 600
```

Details: [CG Mode](/en/multi-llm/cg-mode/)

## CWE-345 — Mandatory Checksum Verification in the Update Flow {#cwe-345}

### What Changed

The `moai update` auto-update flow **cannot bypass checksum verification**. If downloading or parsing the release's `checksums.txt` fails, it returns the sentinel error `ErrChecksumUnavailable` and **aborts** the update flow — the binary download is not attempted.

### Retry Policy

The `checksums.txt` download is retried **3 times** with exponential backoff.

| Attempt | Wait time |
|------|-----------|
| 1st (immediate) | 0s |
| 2nd retry | 2s wait |
| 3rd retry | 4s wait |
| No further retries | Fails after ~6s total wait |

(Internal implementation: base delay 2s × 2^(attempt-1) exponential backoff)

If all retries fail, it terminates with the `ErrChecksumUnavailable` sentinel. **No bypass option such as `--skip-checksum` exists**.

### Defense-in-depth

If `downloadAndVerify` is reached with the `version.Checksum` field as an empty string, the binary download does not proceed and `ErrChecksumUnavailable` is returned. Dual protection (checker stage + updater stage) blocks silent bypass.

### Threat Model

- **Attacker**: a network MITM (cannot block everything, but can selectively block/throttle just the `checksums.txt` URL)
- **Attack surface**: the silent fallback that used to install the binary even without checksums.txt
- **Consequence**: unwarned installation of an unsigned backdoored binary
- **CWE mapping**: CWE-345 (Insufficient Verification of Data Authenticity)

### Implementation Locations

- `internal/update/checker.go` — `downloadChecksumWithRetry(checksumsURL, archiveName, maxAttempts, baseDelay)` (`defaultChecksumMaxAttempts=3`, `defaultChecksumBaseDelay=2*time.Second`), the `ErrChecksumUnavailable` sentinel
- `internal/update/updater.go` — the `downloadAndVerify` empty-checksum guard
- The domain whitelist (`https://github.com/modu-ai/moai-adk/...`) is preserved as-is (no change to the SSRF surface)

### Self-Audit

```bash
# Verify release info + checksums.txt existence
moai update --check-only

# Normal flow (on success)
moai update
# Example output: Downloaded checksums.txt (verified)

# When checksums.txt download fails (intentional block, e.g. running after a VPN disconnect)
moai update
# Example output: error: checksum unavailable: persistent retry failure after 3 attempts
```

If the `ErrChecksumUnavailable` message appears, check the following.

1. Verify network connectivity (`curl -I https://github.com/modu-ai/moai-adk/releases/latest`)
2. Verify your proxy / firewall allows the GitHub release asset domain
3. Possible transient GitHub CDN outage — retry shortly
4. **No bypass option such as `--skip-checksum` is provided** — this is intentional policy

For persistent blockage, manual binary installation is recommended.

```bash
# Manual installation (verify integrity yourself)
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash
```

Details: [Updating](/en/cli-reference/update/)

## Self-Audit Checklist

All five items can be checked at once.

```bash
# 1. CWE-732 — settings.local.json permission
stat -c '%a' .claude/settings.local.json 2>/dev/null \
  || stat -f '%A' .claude/settings.local.json 2>/dev/null
# Expected: 600

# 2. CWE-214 — token argv exposure while CG mode is running (with cg mode active)
ps auxe 2>/dev/null | grep -i 'tmux set-environment.*ANTHROPIC_AUTH_TOKEN'
# Expected: 0 matches

# 3. CWE-214 — tmux sensitive temp directory integrity
ls -la ~/.moai/run/ 2>/dev/null
# Expected: empty directory or no stale files

# 4. CWE-345 — Update flow checksum behavior
moai update --check-only
# Expected: release + checksums.txt verified normally

# 5. GLM source file permission (user responsibility)
stat -c '%a' ~/.moai/.env.glm 2>/dev/null \
  || stat -f '%A' ~/.moai/.env.glm 2>/dev/null
# Expected: 600 (when the file exists)
```

If all 5 items meet the expected values, the v2.20.0-rc1 security hardening is functioning correctly.

## References

### CHANGELOG

[CHANGELOG `[Unreleased]` v2.20.0-rc1 Security section](https://github.com/modu-ai/moai-adk/blob/main/CHANGELOG.md)

### SPEC

- `SPEC-V3R5-SECURITY-CRIT-001` — upstream source of truth, status `implemented` v0.2.0
- PR #1032 merge commit `03a2552a2`

### Commits

- `b48bd86cb` — M1 settings.local.json 0o600 hardening (CWE-732/552)
- `10776c4b8` — M2 tmux sensitive env source-file injection (CWE-214)
- `ee1335282` — M3 mandatory checksum verification with retry (CWE-345)
- `b4e7115cb` — M4 cross-cutting verification + frontmatter

### CWE / OWASP

- [CWE-732](https://cwe.mitre.org/data/definitions/732.html) — Incorrect Permission Assignment for Critical Resource
- [CWE-552](https://cwe.mitre.org/data/definitions/552.html) — Files or Directories Accessible to External Parties
- [CWE-214](https://cwe.mitre.org/data/definitions/214.html) — Invocation of Process Using Visible Sensitive Information
- [CWE-345](https://cwe.mitre.org/data/definitions/345.html) — Insufficient Verification of Data Authenticity

### Related Pages

- [settings.json Guide](/en/advanced/settings-json/) — the `settings.local.json` permissions section
- [Updating](/en/cli-reference/update/) — the checksum verification section
- [CG Mode](/en/multi-llm/cg-mode/) — the tmux environment-variable injection security model
