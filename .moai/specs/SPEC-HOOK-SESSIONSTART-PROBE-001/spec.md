---
id: SPEC-HOOK-SESSIONSTART-PROBE-001
title: "SessionStart Probe — Surface Silent moai-binary Degradation"
version: "0.1.0"
status: completed
created: 2026-07-07
updated: 2026-07-07
author: manager-spec
priority: P2
phase: "v14.4.0"
module: ".claude/hooks/moai"
lifecycle: spec-anchored
tags: "hook, sessionstart, probe, mode-a, genuine-risk"
tier: S
---

# SPEC-HOOK-SESSIONSTART-PROBE-001 — SessionStart Probe

## §A Problem Statement

### §A.1 Verified facts (cited evidence)

The 31 wrapper scripts at `.claude/hooks/moai/handle-*.sh` each carry an IDENTICAL
3-tier moai-binary resolution chain (only the `hook <event>` token differs):

1. `command -v moai` on PATH → `exec moai hook <event>`
2. elif `[ -f "$HOME/go/bin/moai" ]` → exec it
3. elif `[ -f "$HOME/.local/bin/moai" ]` → exec it
4. else `exit 0` — SILENT no-op ("Claude Code handles missing hooks gracefully")

**Evidence (command → observed)**, reproduced from
`.claude/rules/moai/development/hook-independence.md` §3 row A and re-verifiable
against the current tree:

- `ls .claude/hooks/moai/handle-*.sh | wc -l` → 31
- `grep -l 'command -v moai' .claude/hooks/moai/handle-*.sh | wc -l` → 31
- `for f in .claude/hooks/moai/handle-*.sh; do grep -q 'command -v moai' "$f" || echo "$f"; done` → (empty — 0 wrappers lack the chain)

### §A.2 Genuine-risk trigger (PRECISE — not over-stated)

The correlated-degradation trigger is the **conjunction of all three tiers being
absent**: `moai` unresolvable on PATH **AND** absent in `$HOME/go/bin` **AND**
absent in `$HOME/.local/bin`.

This is NOT "moai not in PATH" alone — tiers 2 and 3 are a real fallback, and
stating the first-tier check alone as the trigger would over-state the risk
(forbidden by `.claude/rules/moai/core/verification-claim-integrity.md` §1.1
surface 3 — unobserved defect-claim prohibition).

When the trigger fires, all 31 wrappers degrade to silent `exit 0`
SIMULTANEOUSLY with no surfaced signal. A real incident of this family is
documented at `CLAUDE.md` §17 Troubleshooting ("moai hook subagent-stop fails —
binary not in PATH").

### §A.3 Classification

This is classified **genuine-risk (mode A)** in
`.claude/rules/moai/development/hook-independence.md` §3 and §5 — the SSOT for
the risk classification and the precise trigger. That doctrine §6 carries the
authoritative mitigation Recommendation that THIS SPEC implements:

> "Mode A (genuine-risk) — Recommendation: consider a one-time SessionStart probe
> that warns (once, non-blocking) when moai is unresolvable in all 3 tiers,
> converting the silent-simultaneous-degradation into a surfaced signal. The
> wrappers themselves are NOT to be edited."

The final clause ("wrappers NOT to be edited") is the constraint of THAT
audit-and-doctrine deliverable. This SPEC is the explicitly-authorized follow-up
that the Recommendation itself defers to ("This is a *recommendation*, deferred
to a separate follow-up — adding the probe is a hook/SessionStart change, not
part of this audit-and-doctrine deliverable"). See plan.md §Design for the
reconciliation.

## §B Goal

Convert the silent simultaneous degradation into a surfaced, non-blocking,
once-per-session signal — without altering the success-path behavior, without
touching the 30 non-SessionStart wrappers, and without introducing any new
shared failure mode into the hook layer.

> **"Surfaced" definition (used throughout this SPEC).** A "surfaced" signal
> means a signal injected into the **model's context** via SessionStart stdout
> `hookSpecificOutput.additionalContext`. The model typically relays such
> signals to the user in its next response, but relay is not deterministic.
> An after-the-fact audit log at `$HOME/.moai/logs/hook-stderr.log` provides
> a secondary channel for post-hoc discovery (NOT real-time user surfacing).
> See plan.md §Design.7 for the channel-role separation.

## §C Scope

**In scope:**

- The fallback branch of `handle-session-start.sh` (the SessionStart wrapper)
  ONLY — emit a surfaced warning when all 3 tiers are absent.
- Once-per-session dedup (see plan.md §Dedup Mechanism).
- Template-First mirror parity: `handle-session-start.sh.tmpl` edited in the
  same commit.

**Out of scope:** see §Exclusions.

## §D Requirements (GEARS)

### REQ-HOOK-001 (Ubiquitous — success-path preservation)

The SessionStart wrapper shall preserve byte-identical behavior on the success
path (when any of the 3 moai-binary resolution tiers resolves).

### REQ-HOOK-002 (Compound — surfaced signal on startup-only)

**When** all 3 moai-binary resolution tiers are simultaneously absent (`moai`
unresolvable on PATH AND absent in `$HOME/go/bin` AND absent in
`$HOME/.local/bin`), **while** the SessionStart event source is `startup`,
the SessionStart wrapper fallback shall emit a non-blocking surfaced warning
before `exit 0`.

### REQ-HOOK-003 (Ubiquitous — non-blocking)

The SessionStart wrapper shall preserve `exit 0` as the fallback's terminal
exit code (advisory only — no `exit 2`, no session halt, no `set -e`
propagation).

### REQ-HOOK-004 (When / event-detected — once-per-session dedup)

**When** SessionStart fires with `source` ∈ {`resume`, `clear`, `compact`} (the
non-startup matchers), the SessionStart wrapper fallback shall suppress the
warning emission (once-per-session dedup via `source=startup` gating).

### REQ-HOOK-005 (Ubiquitous — wrapper-layer scope discipline)

The 30 non-SessionStart wrapper scripts (`handle-*.sh` excluding
`handle-session-start.sh`) shall remain byte-identical in this SPEC's run-phase
commit (zero scope creep to the rest of the wrapper layer; the genuine-risk
mitigation is SessionStart-localized).

### REQ-HOOK-006 (Where / capability gate — Template-First parity)

**Where** the Template-First Rule (`CLAUDE.local.md` §2 [HARD]) applies, the
local `.claude/hooks/moai/handle-session-start.sh` and the template source
`internal/template/templates/.claude/hooks/moai/handle-session-start.sh.tmpl`
shall be edited in the same commit with byte-identical fallback-branch content
(no template-variable substitution in the edited branch — `$HOME` literal
preserved per CLAUDE.local.md §14).

### REQ-HOOK-007 (Ubiquitous — warning actionability)

The warning content shall name (a) the 3 absent resolution tiers (PATH,
`$HOME/go/bin`, `$HOME/.local/bin`), (b) the consequence (all 31 wrappers
silently no-op), (c) the non-blocking framing (advisory; session continues), and
(d) the remediation hint (reinstall moai or restore PATH) — so the surfaced
signal is actionable, not vacuous.

## §E Constraints (Non-Functional)

- **Tier S** (minimal, single-pass, additive-only) — see plan.md §F for the
  single M1 milestone.
- **No behavior change on the success path** — REQ-HOOK-001 byte-identical.
- **Non-blocking** — REQ-HOOK-003; the warning MUST NOT block session start.
- **Additive-only** — no regression to the existing 3-tier chain, no edit to
  the 30 non-SessionStart wrappers (REQ-HOOK-005).
- **hook-independence.md §7 authoring checklist compliance** — Q1 (self-contained
  bash; does NOT inherit mode A), Q4 (graceful degrade — warning emission itself
  fails open), Q5 (surfaced rather than silent).
- **verification-claim-integrity** — this SPEC's genuine-risk claim cites
  hook-independence.md §3/§5 evidence rows (command → observed), not prose
  assertion.

## §F Out of Scope — Exclusions

### Out of Scope — The 30 non-SessionStart wrapper scripts

- No edit to any `handle-*.sh` other than `handle-session-start.sh`. The
  genuine-risk mitigation is localized to the SessionStart event (which fires
  once per session startup) — fanning the probe out to the other 30 wrappers
  would multiply surface area without proportional benefit (the warning is
  session-scoped, not per-event-scoped).
- The other 30 wrappers retain their silent `exit 0` fallback. Their
  silent-degradation in the all-3-tiers-absent state is acknowledged but
  accepted — the SessionStart probe surfaces the condition once per session,
  which is sufficient signal.

### Out of Scope — The 3 governance-gate scripts

- `status-transition-ownership.sh`, `sync-phase-quality-gate.sh`,
  `team-ac-verify.sh` are NOT modified. Per hook-independence.md §4.1, these
  gates do NOT share mode A (they are self-contained bash with no moai-binary
  dependency) — they are the positive-signal layer that already provides
  defense depth. Touching them would be unrelated scope creep.

### Out of Scope — Resolution chain restructuring

- The 3-tier chain itself (PATH → `$HOME/go/bin` → `$HOME/.local/bin`) is NOT
  restructured, reordered, or extended. Adding a 4th tier (e.g.,
  `/opt/homebrew/bin/moai`) is out of scope — the probe only surfaces the
  all-3-tiers-absent state, it does not change how the tiers are queried.

### Out of Scope — Automatic remediation

- The probe does NOT attempt to reinstall moai, modify PATH, or write to any
  binary location. Remediation is a human action; the probe's job is to surface
  the condition, not fix it.

### Out of Scope — Non-SessionStart event probes

- No probe for PreToolUse, PostToolUse, Stop, SubagentStart, etc. The
  SessionStart event is uniquely suited (once-per-session startup semantics);
  probing other events is a separate follow-up if needed.

### Out of Scope — Settings.json changes

- The SessionStart registration in `.claude/settings.json` (timeout 30s,
  matcher `startup|resume|clear|compact`) is NOT modified. The probe runs
  within the existing 30s timeout budget; no `once: true` flag added (dedup is
  source-gated inside the wrapper, not runtime-gated by Claude Code).

### Out of Scope — Go handler changes

- `internal/hook/session_start.go` is NOT modified. The Go handler only
  executes after the wrapper successfully resolved the binary (it `exec`s
  `moai hook session-start`), so it cannot observe the all-3-tiers-absent
  state. The probe is a wrapper-layer concern, not a Go-handler concern.

## §G Traceability

| REQ | Acceptance Criterion | plan.md Milestone |
|-----|---------------------|-------------------|
| REQ-HOOK-001 | AC-HOOK-001 | M1 |
| REQ-HOOK-002 | AC-HOOK-002 | M1 |
| REQ-HOOK-003 | AC-HOOK-003 | M1 |
| REQ-HOOK-004 | AC-HOOK-004 | M1 |
| REQ-HOOK-005 | AC-HOOK-005 | M1 |
| REQ-HOOK-006 | AC-HOOK-006 | M1 |
| REQ-HOOK-007 | AC-HOOK-007 | M1 |

## §H Cross-References

- `.claude/rules/moai/development/hook-independence.md` §3 row A, §5, §6, §7 —
  risk SSOT, Recommendation authorization, authoring checklist.
- `.claude/rules/moai/core/hooks-system.md` § Hook Events (SessionStart stdin
  `source` field; stdout `hookSpecificOutput.additionalContext`) and § Timeout
  Configuration (SessionStart 30s).
- `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 3 —
  unobserved defect-claim prohibition (binds the genuine-risk trigger wording).
- `CLAUDE.local.md` §2 [HARD] Template-First Rule; §14 [HARD] `.HomeDir`/`.GoBinPath`
  prohibition in fallback paths.
- `CLAUDE.md` §17 Troubleshooting — documented real incident in this family
  ("moai hook subagent-stop fails — binary not in PATH").

---

Version: 0.1.0
Status: draft (plan-phase authoring complete; awaiting plan-auditor verdict)
Classification: SPEC (feature to implement) — implements hook-independence.md §6 Recommendation
