# acceptance.md — SPEC-AUTONOMY-TIERS-001

> Acceptance criteria for SPEC-AUTONOMY-TIERS-001. Each AC traces 1:1 to a REQ in spec.md §D. Format: AC-XXX labeled `Given … When … Then …`, binary-testable.

## §D. AC Matrix

| AC ID | Requirement | Verification | Pass condition |
|-------|-------------|--------------|----------------|
| AC-AUTONOMY-TIERS-001 | REQ-001 (init selector) | `moai init` wizard fixture + `--autonomy-tier` flag unit test | 3-tier wizard page present; flag accepts the closed set, rejects invalid |
| AC-AUTONOMY-TIERS-002 | REQ-002 (web toggle) | `moai web` console fixture | 3-tier toggle present; fully-autonomous disabled without sandbox proof AND under kill-switch |
| AC-AUTONOMY-TIERS-003 | REQ-003 (renderer scope) | renderer unit test (path resolution) | `defaultMode` written to USER-scope path; deny/ask written to PROJECT-scope path; per-tier `defaultMode` correct |
| AC-AUTONOMY-TIERS-004 | REQ-004 (deny/ask invariance) | renderer unit test (byte diff) | deny/ask arrays byte-identical across the 3 rendered tiers |
| AC-AUTONOMY-TIERS-005 | REQ-005 (kill-switch) | kill-switch unit test | `disableBypassPermissionsMode` rejects fully-autonomous in selector + web + renderer; existing bypass session downgrades to auto + advisory to `.moai/logs/autonomy-downgrade.log` (verified by `grep autonomy-downgrade .moai/logs/autonomy-downgrade.log`) |
| AC-AUTONOMY-TIERS-006 | REQ-006 (template opt-in) | template fixture + CI neutrality guard | template ships `semi-auto` default; no `fully-autonomous` default; selector copy uses generic prose (no SPEC ID / REQ token leak) |
| AC-AUTONOMY-TIERS-007 | REQ-007 (backward compat) | regression test (byte diff vs today's template) | unset / `semi-auto` → renderer output byte-identical to today's template output |

## §D.1 Severity

All 7 ACs are MUST-PASS (severity: critical). The deny/ask invariance (AC-004), the kill-switch (AC-005), the template opt-in (AC-006), and the backward-compat regression guard (AC-007) are load-bearing — a failure on any of these blocks merge. AC-001 / AC-002 / AC-003 are functional-critical (the selector and renderer are the SPEC's whole point).

## §D.2 Given-When-Then Scenarios

### AC-AUTONOMY-TIERS-001 — init selector

**Given** the `moai init` wizard is running interactively,
**When** the user reaches the autonomy-tier page,
**Then** the page offers `{semi-auto, automatic, fully-autonomous}` with `semi-auto` pre-selected.

**Given** the `moai init` command with `--non-interactive`,
**When** the user passes `--autonomy-tier automatic`,
**Then** the wizard persists `automatic` via M1's surface.

**Given** `moai init --non-interactive --autonomy-tier bogus`,
**When** the flag is parsed,
**Then** the command exits non-zero with an error naming the valid values (fail-loud, NOT silent fallback).

### AC-AUTONOMY-TIERS-002 — web toggle

**Given** the `moai web` console is active,
**When** the user opens the tier toggle,
**Then** the toggle offers 3 tiers; `fully-autonomous` is disabled when no sandbox proof is present.

**Given** `disableBypassPermissionsMode: true` in the managed config,
**When** the user opens the tier toggle,
**Then** `fully-autonomous` is disabled regardless of sandbox proof.

### AC-AUTONOMY-TIERS-003 — renderer scope

**Given** the renderer invoked with tier = `automatic`,
**When** the renderer writes,
**Then** `permissions.defaultMode: "auto"` lands in the USER-scope settings file AND the deny/ask arrays land in the PROJECT-scope settings file.

**Given** the renderer invoked with tier = `fully-autonomous` (sandbox proof present, kill-switch off),
**When** the renderer writes,
**Then** `permissions.defaultMode: "bypassPermissions"` lands in USER scope (NOT PROJECT scope).

### AC-AUTONOMY-TIERS-004 — deny/ask invariance

**Given** the renderer invoked with each of the 3 tiers in turn,
**When** the deny/ask arrays are diffed across the 3 outputs,
**Then** the diff is empty (byte-identical) — the deny/ask rule set is tier-invariant.

### AC-AUTONOMY-TIERS-005 — kill-switch

**Given** `disableBypassPermissionsMode: true`,
**When** the user runs `moai init --autonomy-tier fully-autonomous`,
**Then** the command exits non-zero with an error citing the kill-switch.

**Given** `disableBypassPermissionsMode: true` AND an existing `settings.local.json` carrying `defaultMode: bypassPermissions`,
**When** the session starts,
**Then** the effective behavior is `automatic`-equivalent (`defaultMode: auto`) AND an advisory recording the downgrade is appended to `.moai/logs/autonomy-downgrade.log` (verified by `grep autonomy-downgrade .moai/logs/autonomy-downgrade.log` after session start).

### AC-AUTONOMY-TIERS-006 — template opt-in

**Given** the distributed template (`internal/template/templates/**`),
**When** the template's `moai init` default page is rendered,
**Then** the pre-selected tier is `semi-auto` AND no `fully-autonomous` default ships AND the selector copy contains no internal SPEC ID / REQ token (CI neutrality guard passes).

### AC-AUTONOMY-TIERS-007 — backward compat

**Given** no tier selection persisted (unset) OR tier = `semi-auto`,
**When** the renderer writes,
**Then** the output is byte-identical (modulo whitespace) to today's template `settings.json` output — zero behavior delta for sessions that do not opt in.

## §D.3 Edge Cases

- **Tier persistence read after a manual `settings.local.json` edit**: a user who hand-edits `MOAI_AUTONOMY_TIER` in the env block MUST produce the same effective tier as the selector. The renderer reads the EFFECTIVE tier (env-key wins per STOPCHAIN-TRIM's canonical-source rule); the selector reads the PERSISTED tier (M1 surface). A mismatch is resolved env-wins (the hooks already read env).
- **Kill-switch engaged mid-session**: the downgrade to `auto` + advisory fires at the next session start, NOT retroactively mid-session (Claude Code reads `defaultMode` at session start). Document this in user-facing release notes.
- **Sandbox proof present but kill-switch engaged**: kill-switch wins (AC-005 trumps AC-002's sandbox-proof enablement).
- **`--autonomy-tier` flag conflicts with a persisted `workflow.yaml` key**: the flag wins for the invocation (mirrors `--profile` precedence at `init.go:138`); the persisted key is NOT overwritten unless the user confirms.

## §D.4 Quality Gate Criteria

- All 7 ACs PASS (severity: critical — no partial credit).
- `go test ./internal/cli/... ./internal/config/...` green.
- `golangci-lint run` green.
- CI neutrality guard (`template-neutrality-check.yaml`) green — no internal SPEC ID / REQ token leaked into the template.
- No regression in the existing `moai init` wizard tests (`internal/cli/init_test.go`).

## §D.5 Definition of Done

- 4 surfaces landed: selector (init), selector (web), renderer, kill-switch/sandbox-proof gate.
- Token foundation (`autonomy.go` / `envkeys.go` / `defaults.go`) UNMODIFIED (consumed, not re-scoped — AP-1 guard).
- deny/ask arrays identical across tiers (AC-004).
- Template ships `semi-auto` default (AC-006).
- Backward-compat regression guard green (AC-007).
- CHANGELOG entry + user-facing release note covering the kill-switch mid-session downgrade timing (§D.3 edge case).
