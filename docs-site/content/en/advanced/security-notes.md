---
title: Security Notes
description: "MoAI-ADK v2.20.0-rc1 security hardening — CWE-732/214/345 mappings and self-audit procedures"
weight: 72
draft: false
tags: ["security", "cwe", "audit"]
---

# Security Notes

This page documents the **user-visible security changes** introduced in MoAI-ADK v2.20.0-rc1. Each entry includes the CWE mapping, the behavioral change, and self-audit commands.

## Why this page exists

`SPEC-V3R5-SECURITY-CRIT-001` (PR #1032, merge commit `03a2552a2`) corrected **three P0 release-blocker security defects** identified in the v2.14.0 → v2.20.0-rc1 code review. This page makes the corrections and user-side verification procedures part of the official 4-locale documentation.

All three defects relate to the GLM integration + auto-update paths:

- **CWE-732 / CWE-552** — `.claude/settings.local.json` is now forced to mode `0o600` (owner-only read/write).
- **CWE-214** — `moai cg` now injects tmux environment variables via source-file instead of argv, so GLM tokens are no longer visible in process listings.
- **CWE-345** — `moai update` now performs mandatory checksum verification; download failure aborts the update.

Each fix is locked by regression tests so future regressions are blocked.

## CWE-732 — settings.local.json Permission Hardening {#cwe-732}

### What changed

`.claude/settings.local.json` is now created and updated with file mode **`0o600`** (owner-only read/write). Previously it was created with `0o644` (owner read/write + group/world read), which on multi-user workstations allowed other local users to read sensitive credentials including `ANTHROPIC_AUTH_TOKEN`.

### Threat model

- **Attacker**: Low-privilege local user on the same host
- **Attack surface**: Group/world read permission on `.claude/settings.local.json`
- **Leaked data**: GLM API token (`ANTHROPIC_AUTH_TOKEN`), OAuth refresh token, other `settings.Env` values
- **CWE mapping**: CWE-732 (Incorrect Permission Assignment for Critical Resource), CWE-552 (Files or Directories Accessible to External Parties)

### Implementation location

- `internal/hook/settings_io.go` — constant `secureSettingsMode os.FileMode = 0o600` + `writeSettingsSecure` helper
- `internal/hook/session_start.go` — all `settings.local.json` writers (`ensureGLMCredentials`, `ensureClaudeEnvFile`)
- `internal/hook/session_end.go` — GLM keys write-back path

### Self-audit

Check the permission on your existing `settings.local.json`:

```bash
# Linux
stat -c '%a' .claude/settings.local.json
# expect: 600

# macOS
stat -f '%A' .claude/settings.local.json
# expect: 600
```

If the permission shows `644` or anything looser, MoAI-ADK will auto-correct it to `0o600` on the next session start. To fix immediately:

```bash
chmod 0600 .claude/settings.local.json
```

### Trade-off

Workflows that expected the file to be `group-readable` (e.g., another OS user reading the same project directory — an extremely rare scenario) will break. This trade-off is intentional; the security recovery is the clear priority.

## CWE-214 — tmux IPC Token argv Exposure Mitigation {#cwe-214}

### What changed

When `moai cg` (CG mode) injects the GLM token (`ANTHROPIC_AUTH_TOKEN`) into a tmux session environment, it now uses the **source-file channel** (`tmux source-file <tmp>`) instead of the **argv channel** (`tmux set-environment <KEY> <VALUE>`). The token is no longer visible in `ps auxe`, `/proc/<pid>/cmdline`, auditd logs, sysmon traces, or crash dumps.

### Implementation flow

1. Create a temp file under `~/.moai/run/` via `mkstemp` (mode `0o600` by default + explicit `chmod 0o600`).
2. Write a single line `set-environment -t <session> <KEY> <VALUE>` to the temp file.
3. Invoke `tmux source-file <tmp>` so tmux reads the file and injects the variable.
4. Immediately `os.Remove` the temp file.

Only the temp file path is visible in argv. The token itself never appears in argv.

### Threat model

- **Attacker**: Local user on the same host + system log collectors (`ps`, `/proc`, auditd, sysmon)
- **Attack surface**: argv channel of tmux env injection
- **Leaked data**: Momentary exposure of the GLM API token
- **CWE mapping**: CWE-214 (Invocation of Process Using Visible Sensitive Information)

### Implementation location

- `internal/tmux/session.go` — `InjectSensitiveEnv` method, `sensitiveTempDir = ".moai/run"`, `mkstemp` + `chmod 0o600` + `tmux source-file` + `os.Remove`
- `internal/tmux/errors.go` — `ErrTmuxSensitiveInjectFailed` sentinel
- `internal/hook/glm_tmux.go` — `ensureTmuxGLMEnv` branches `ANTHROPIC_AUTH_TOKEN` to the sensitive path; non-sensitive values (URL, model names, etc.) keep the argv path.

### Non-sensitive values keep argv

`CLAUDE_CONFIG_DIR` (directory path), `ANTHROPIC_BASE_URL` (URL), `ANTHROPIC_DEFAULT_*_MODEL` (model names), and similar non-token values keep using the argv path. This is intentional and unrelated to the token leak risk.

### Failure behavior

If the source-file injection fails (disk full, `tmux source-file` failure, etc.), the implementation **does not fall back to argv** — it returns the `ErrTmuxSensitiveInjectFailed` sentinel and aborts the injection.

### Self-audit

Verify the token is not exposed in argv while CG mode is running:

```bash
# Inside a new tmux session after running moai cg
ps auxe | grep -i 'tmux set-environment.*ANTHROPIC_AUTH_TOKEN'
# expect: 0 matches (token absent from argv)
```

Verify the temp files are unlinked correctly:

```bash
ls -la ~/.moai/run/ 2>/dev/null
# expect: empty directory or no stale files
```

If stale files remain in `~/.moai/run/` after a session ends, you may remove them manually — they are no longer a security risk (they were already targeted for unlink).

### User responsibility

The `~/.moai/.env.glm` source file must retain `0o600` permission. `moai glm` sets this automatically:

```bash
stat -c '%a' ~/.moai/.env.glm    # Linux: 600
stat -f '%A' ~/.moai/.env.glm    # macOS: 600
```

See: [CG Mode](/en/multi-llm/cg-mode/)

## CWE-345 — Update Flow Mandatory Checksum Verification {#cwe-345}

### What changed

The `moai update` automatic update flow **cannot bypass checksum verification**. If the release's `checksums.txt` download fails or parsing fails, the system returns the sentinel error `ErrChecksumUnavailable` and **aborts** the update — no binary download is attempted.

### Retry policy

`checksums.txt` downloads are retried **3 times** with exponential backoff:

| Attempt | Wait time |
|---------|-----------|
| 1st (immediate) | 0s |
| 2nd retry | 2s |
| 3rd retry | 4s |
| no further retry | total wait ~6s before failure |

(Internal implementation: base delay 2s × 2^(attempt-1) exponential backoff)

If all retries fail, the update aborts with the `ErrChecksumUnavailable` sentinel. **No `--skip-checksum` opt-out exists.**

### Defense-in-depth

If `version.Checksum` arrives at `downloadAndVerify` as an empty string, the binary download is rejected and `ErrChecksumUnavailable` is returned. Two layers of protection (checker stage + updater stage) block silent bypass.

### Threat model

- **Attacker**: Network MITM (cannot block everything but can selectively block/throttle the `checksums.txt` URL)
- **Attack surface**: Silent fallback that installed binaries without checksums.txt
- **Outcome**: Silent installation of an unsigned backdoor binary
- **CWE mapping**: CWE-345 (Insufficient Verification of Data Authenticity)

### Implementation location

- `internal/update/checker.go` — `downloadChecksumWithRetry(checksumsURL, archiveName, maxAttempts, baseDelay)` (`defaultChecksumMaxAttempts=3`, `defaultChecksumBaseDelay=2*time.Second`), `ErrChecksumUnavailable` sentinel
- `internal/update/updater.go` — `downloadAndVerify` empty-checksum guard
- Domain whitelist (`https://github.com/modu-ai/moai-adk/...`) is preserved unchanged (no SSRF surface change)

### Self-audit

```bash
# Verify release info + checksums.txt is reachable
moai update --check-only

# Happy path (success case)
moai update
# example output: Downloaded checksums.txt (verified)

# Intentional failure case (e.g., after VPN disconnection)
moai update
# example output: error: checksum unavailable: persistent retry failure after 3 attempts
```

If you see an `ErrChecksumUnavailable` message:

1. Check network connectivity (`curl -I https://github.com/modu-ai/moai-adk/releases/latest`)
2. Verify your proxy/firewall allows the GitHub release-asset domain
3. Consider transient GitHub CDN issues — retry after a short wait
4. **No `--skip-checksum` opt-out is provided** — this is intentional policy

For persistent blocks, manual binary installation is recommended:

```bash
# Manual installation (you verify integrity yourself)
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash
```

See: [Update](/en/getting-started/update/)

## Self-Audit Checklist

```bash
# 1. CWE-732 — settings.local.json permission
stat -c '%a' .claude/settings.local.json 2>/dev/null \
  || stat -f '%A' .claude/settings.local.json 2>/dev/null
# expect: 600

# 2. CWE-214 — token argv exposure while CG mode is active
ps auxe 2>/dev/null | grep -i 'tmux set-environment.*ANTHROPIC_AUTH_TOKEN'
# expect: 0 matches

# 3. CWE-214 — tmux sensitive temp directory integrity
ls -la ~/.moai/run/ 2>/dev/null
# expect: empty directory or no stale files

# 4. CWE-345 — Update flow checksum behavior
moai update --check-only
# expect: release + checksums.txt confirmed

# 5. GLM source file permission (user responsibility)
stat -c '%a' ~/.moai/.env.glm 2>/dev/null \
  || stat -f '%A' ~/.moai/.env.glm 2>/dev/null
# expect: 600 (if the file exists)
```

If all five checks meet expectations, the v2.20.0-rc1 security hardening is operating correctly.

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

### Related pages

- [settings.json Guide](/en/advanced/settings-json/) — `settings.local.json` permission section
- [Update](/en/getting-started/update/) — checksum verification section
- [CG Mode](/en/multi-llm/cg-mode/) — tmux environment variable injection security model
