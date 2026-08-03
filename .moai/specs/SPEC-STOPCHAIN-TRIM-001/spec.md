---
id: SPEC-STOPCHAIN-TRIM-001
title: "Stop-chain shell trim, per-edit hook consolidation, and MOAI_AUTONOMY_TIER mode-aware hooks"
version: 0.1.0
status: in-progress
created: 2026-08-03
updated: 2026-08-03
author: manager-spec
priority: P0
phase: "v3.x target"
module: workflow-hooks
lifecycle: spec-anchored
tags: "stop-hook, per-edit-hook, mode-token, autonomy-tier, shell-trim, pre-tool, run-phase, autonomy-epic"
tier: M
related_specs: [SPEC-AUDIT-SNAPSHOT-001]
---

# SPEC-STOPCHAIN-TRIM-001 — Stop-chain trim + per-edit consolidation + mode token

## HISTORY

- 2026-08-03 — Initial draft. Codifies §3.5 items A8 + A10 + A11 of the autonomy-workflow redesign report (`moai-autonomy-workflow-redesign-20260803.html`), plus the §3.1 `MOAI_AUTONOMY_TIER` mode token those items depend on. Second P0 SPEC of the autonomy-workflow-epic; proceeds after `SPEC-AUDIT-SNAPSHOT-001` completes. No prior art in the SPEC catalog.

## §A. User Story

**As a** MoAI maintainer running long TDD cycles and unattended autonomous goal loops,
**I want** per-turn Stop-chain hooks to short-circuit at the shell layer when their precondition is absent, per-edit Pre/PostToolUse hooks consolidated to a per-cycle Stop hook, and a single `MOAI_AUTONOMY_TIER` mode token that the hooks read to relax advisory gates in higher autonomy tiers,
**so that** the per-turn `moai` binary cold-start tax (6 fork+exec, 4 cold) collapses to conditional 2 cold starts, the per-edit 2-spawn tax on TDD atomic edits disappears, and unattended autonomous loops stop paying per-turn / per-commit / per-teammate interrupt friction that the user already approved at Implementation Kickoff.

**Outcome hypotheses (from §3.5 + §1.4):**

- A10: a session with no armed goal pays ZERO `moai hook stop-goal` cold-starts per turn-end (the wrapper `exit 0`s at the shell layer before exec'ing the binary); sync-gate pays nothing on non-sync commits.
- A8: a TDD cycle of N atomic edits pays ZERO `develop-pre-implementation` + `develop-post-implementation` spawns per edit (the bulk of the work moves to a per-cycle `develop-completion` Stop hook at milestone/commit boundaries).
- A11 + mode token: at `MOAI_AUTONOMY_TIER=fully-autonomous`, sync-gate is advisory (no `decision:block`), the commit gate is OFF, and SubagentStop/TeammateIdle/TaskCompleted hooks are dormant; at `automatic`, the commit gate is OFF but sync-gate still blocks on build failures; at `semi-auto` (default, also the unset/empty behavior), everything is unchanged from today.

## §B. Scope

**In scope — A8 + A10 + A11 + the mode token, all from §3.5/§3.1 (existing knobs re-wired):**

- **A10** — Stop-chain shell shortening: `handle-stop-goal.sh` exec's `moai hook stop-goal` ONLY when a goal state file exists; `sync-phase-quality-gate.sh` exec's its logic ONLY when the HEAD commit subject matches a sync-commit sentinel; the 4 observer/security Stop hooks transition to async advisory.
- **A8** — per-edit hook consolidation: the `develop-pre-implementation` (PreToolUse, 5s) and `develop-post-implementation` (PostToolUse, 10s) spawns on every Write|Edit move to a per-cycle `develop-completion` Stop hook at the milestone/commit boundary; PreToolUse retains ONLY the destructive-pattern + scope-discipline checks.
- **A11 + mode token** — a single mode token `MOAI_AUTONOMY_TIER` ∈ {`semi-auto`, `automatic`, `fully-autonomous`} (sourced from `workflow.yaml` or env) is read by the mode-aware hooks; sync-gate, the commit gate (`internal/hook/pre_tool.go` IsGitCommit ~L429-441), and subagent lifecycle hooks (SubagentStop / TeammateIdle / TaskCompleted) all branch on it.

**Out of scope — broader epic items:** the goal-evaluator HTML dashboard, the `moai_goal_render` surface, the stateful MCP tool layer, A1-A4 (owned by `SPEC-AUDIT-SNAPSHOT-001`), A5 (docs ∥ audit parallelization). These are sibling P0/P1/P2 SPECs in the epic.

### Out of Scope — Redesign items beyond A8/A10/A11 + mode token

- The goal-evaluator HTML dashboard and `moai_goal_render` surface — epic-level P1 work.
- The stateful MCP tool layer (`moai_goal_arm`, `moai_verify_snapshot`, etc.) — epic-level P2 work.
- A1 (sticky hash), A2 (skip alignment), A3 (4-dim binding), A4 (shared snapshot) — owned by `SPEC-AUDIT-SNAPSHOT-001`.
- A5 (docs ∥ audit parallelization) — sibling P1 SPEC.

### Out of Scope — Hook semantics outside the autonomy axis

- Changing WHAT the destructive-pattern denylist catches (the denylist content is immutable in this SPEC; A11 only changes WHEN the IsGitCommit gate fires, not what it denies).
- Changing WHAT sync-gate measures (lint/test/coverage content); A11 only changes the blocking-vs-advisory MODE per tier, mirroring the existing `MOAI_SYNC_GATE_BLOCKING` env knob.
- Adding new subagent lifecycle events (SubagentStop / TeammateIdle / TaskCompleted are the existing events; A11 makes them dormant at `fully-autonomous`, does not invent new events).

### Out of Scope — defaultMode / permission surface changes

- The report's §3.1 recommendation that `defaultMode` live in USER settings and `deny/ask` rules in PROJECT settings is a settings-layer re-architecture tracked separately; this SPEC only adds `MOAI_AUTONOMY_TIER` as a hook input, it does not move `defaultMode`.

## §C. Requirements (GEARS)

### REQ-STOPCHAIN-TRIM-001 — Stop-chain shell precondition short-circuit (A10)

**When** the Stop event fires at turn-end, the `handle-stop-goal.sh` wrapper SHALL first check at the shell layer (before any `moai` binary exec) whether a goal state file exists for the current session (`.moai/state/goal/<session-id>.json`); **when** the goal state file is absent, the wrapper SHALL `exit 0` immediately without invoking the `moai` binary.

**When** the Stop event fires and `sync-phase-quality-gate.sh` is the active Stop hook, the wrapper SHALL first check at the shell layer whether the current HEAD commit subject matches a sync-commit sentinel (the existing once-per-commit marker, reinforced); **when** the HEAD subject is NOT a sync commit, the wrapper SHALL `exit 0` without running `go vet` / `go build` / lint.

**Where** an observer or security Stop hook (the 4 such hooks identified in §1.4) currently runs synchronously, the hook SHALL transition to async advisory mode — its result is emitted as a `systemMessage` / advisory record but does NOT block the turn-end.

### REQ-STOPCHAIN-TRIM-002 — Per-edit hook consolidation (A8)

**Where** `manager-develop.md` frontmatter currently registers `develop-pre-implementation` (PreToolUse, 5s timeout) and `develop-post-implementation` (PostToolUse, 10s timeout) hooks that fire on every Write|Edit, the per-edit firing SHALL be removed; the bulk of the verification work those hooks performed SHALL move to a per-cycle `develop-completion` Stop hook that fires once per develop cycle (or at the milestone commit boundary), not once per edit.

**Where** the PreToolUse hook surface remains on Write|Edit, the residual PreToolUse hook SHALL perform ONLY the destructive-pattern detection and scope-discipline check (the "touch only what you were asked to touch" guard); PreToolUse SHALL NOT perform full lint / type-check / build verification on each edit.

### REQ-STOPCHAIN-TRIM-003 — MOAI_AUTONOMY_TIER mode token (§3.1)

The system SHALL define a single mode token `MOAI_AUTONOMY_TIER` whose value is one of {`semi-auto`, `automatic`, `fully-autonomous`}, sourced from a single canonical location: the env-key constant `MOAI_AUTONOMY_TIER` in `internal/config/envkeys.go` (read via `os.Getenv` on the Go side, and as `$MOAI_AUTONOMY_TIER` from pure shell). The token is NOT a `workflow.yaml` key — shell hooks (`handle-stop-goal.sh`, `sync-phase-quality-gate.sh`) MUST be able to read the token WITHOUT invoking the `moai` binary or parsing YAML, and only an env-key satisfies that requirement. The constant and the 3-value enum are recorded in `internal/config/defaults.go` for the Go-side reader helper.

**When** `MOAI_AUTONOMY_TIER` is unset or empty, the system SHALL behave identically to `semi-auto` — full backward compatibility, no regression for any session that does not opt in.

### REQ-STOPCHAIN-TRIM-004 — Mode-aware sync-phase gate (A11)

**Where** `sync-phase-quality-gate.sh` evaluates a quality decision, the hook SHALL read `MOAI_AUTONOMY_TIER` and branch: **While** `MOAI_AUTONOMY_TIER = fully-autonomous`, the hook SHALL emit advisory output only (a `systemMessage` recording the gate result) and SHALL NOT emit `decision:block`; **While** `MOAI_AUTONOMY_TIER = automatic`, the hook SHALL block ONLY on build failures (`go build` exit ≠ 0) and emit advisory for lint/test/coverage; **While** `MOAI_AUTONOMY_TIER = semi-auto` (or unset), the hook SHALL retain the current full-blocking behavior.

### REQ-STOPCHAIN-TRIM-005 — Mode-aware commit gate (A11)

**Where** `internal/hook/pre_tool.go` evaluates the IsGitCommit gate (~L429-441), the gate SHALL read `MOAI_AUTONOMY_TIER` and branch: **While** `MOAI_AUTONOMY_TIER ∈ {automatic, fully-autonomous}`, the synchronous vet+lint+test commit gate SHALL be OFF (the gate returns allow without invoking those tools); **While** `MOAI_AUTONOMY_TIER = semi-auto` (or unset), the gate SHALL retain its current full behavior.

**The deny/ask rule set (destructive-pattern denylist — `git push` to main, secrets, `rm -rf`, deploy, etc.) SHALL remain binding at every tier.** The mode token relaxes ONLY the synchronous verification gate, never the destructive-pattern deny path. This is the §3.1 invariant.

### REQ-STOPCHAIN-TRIM-006 — Mode-aware subagent lifecycle hooks (A11)

**Where** the SubagentStop, TeammateIdle, and TaskCompleted hooks fire, the hooks SHALL read `MOAI_AUTONOMY_TIER` and branch: **While** `MOAI_AUTONOMY_TIER = fully-autonomous`, the lifecycle hooks SHALL be dormant (observe-only — they may record audit-log entries but SHALL NOT block, reject, or translate to `AskUserQuestion`); **While** `MOAI_AUTONOMY_TIER ∈ {semi-auto, automatic}`, the lifecycle hooks SHALL retain their current behavior.

### REQ-STOPCHAIN-TRIM-007 — deny/ask invariance (cross-cutting)

**The deny/ask rule set (main push, secrets, `rm -rf`, deploy, and the rest of the destructive-pattern denylist) SHALL remain binding at every value of `MOAI_AUTONOMY_TIER` — `semi-auto`, `automatic`, `fully-autonomous`, and unset.** No tier SHALL weaken a destructive-pattern deny or an ask-rule. The mode token's effect is limited to: (a) advisory vs blocking mode of sync-gate, (b) on/off of the synchronous commit verification gate, (c) dormant vs active subagent lifecycle hooks.

## §D. Constraints

1. **deny/ask rules are tier-invariant** (from §3.1 invariant + report risk callout): the mode token NEVER weakens a destructive-pattern deny or an ask rule. The relaxations apply ONLY to advisory gates, the commit verification gate, and subagent lifecycle dormant mode. REQ-STOPCHAIN-TRIM-007 codifies this as a cross-cutting constraint; REQ-005 §2 restates it for the commit gate specifically.
2. **Backward compatibility** (from §3.1): `MOAI_AUTONOMY_TIER` unset/empty = `semi-auto` = today's behavior. Every change in this SPEC MUST be a no-op for sessions that do not set the token.
3. **Single source of truth for the mode token**: `MOAI_AUTONOMY_TIER` is defined in ONE place (env-keys constant OR `workflow.yaml` key, recorded in `defaults.go`). Parallel restatements in skill YAML / agent body / hook script MUST cross-reference, not re-define.
4. **No new user-visible CLI surface**: the mode token is read from config/env, not exposed as a new CLI flag. UI surfaces (web console toggle, `moai init` interactive) are tracked separately.
5. **Tier classification (Tier M)**: 4 change points across shell hooks + agent frontmatter + Go hook code + config; multi-surface but each change is small.
6. **Hook wrapper timeout unchanged at 5s** (per CLAUDE.local.md §7): the shell precondition check MUST fit within the existing 5s wrapper timeout — the check itself is a single file-existence test or a single `git log -1 --format=%s` call, both well under 5s. A11 does not raise the timeout.

## §E. Assumptions

1. The existing `handle-stop-goal.sh` wrapper is a shell script that currently `exec`s `moai hook stop-goal` unconditionally; a shell-layer `[ -f "$state_file" ] || exit 0` precondition is reachable without changing the wrapper's call contract.
2. The existing once-per-commit sentinel that gates `sync-phase-quality-gate.sh` (referenced in §3.5 A10 as "이미 있음 보강") is greppable from the HEAD commit subject; the reinforcement is a stricter match, not a new mechanism.
3. `internal/hook/pre_tool.go` IsGitCommit branch (~L429-441) is the sole synchronous commit gate; reading `MOAI_AUTONOMY_TIER` at that call site is reachable without a config-loader refactor (env-key read or already-loaded workflow config).
4. The 4 observer/security Stop hooks identified in §1.4 are enumerated in `settings.json` and individually wrappable in `&`-async or a non-blocking dispatcher; async transition does not require a runtime hook-type change.
5. The `develop-pre-implementation` / `develop-post-implementation` hooks in `manager-develop.md` frontmatter are the ONLY per-edit hooks whose removal is in scope; no other agent frontmatter carries per-edit hooks that this SPEC must touch.

## §F. Open Questions (for plan-auditor)

- **OQ-1** (RESOLVED by plan-auditor iter 1 — `internal/config/envkeys.go` inspection): the canonical source is the env-key constant `MOAI_AUTONOMY_TIER`, NOT a `workflow.yaml` key. Rationale: shell hooks (`handle-stop-goal.sh`, `sync-phase-quality-gate.sh`) MUST be able to read the token WITHOUT invoking the `moai` binary or parsing YAML — only an env-key satisfies that (`$MOAI_AUTONOMY_TIER` is reachable from pure shell). The Go side reads the same env-key via `os.Getenv`. Resolution folded into REQ-STOPCHAIN-TRIM-003 (env-key as the single canonical source) and plan M1 (file list now names `internal/config/envkeys.go` explicitly).
- **OQ-2**: Does the `develop-completion` per-cycle Stop hook already exist in any form (e.g. a `manager-develop` Stop hook that fires at cycle end), or is it net-new surface? If it exists, A8 is a content migration; if not, A8 adds a new Stop hook registration.
- **OQ-3**: For the 4 observer/security Stop hooks, is there an existing async/advisory hook channel (a `systemMessage`-only mode the runtime already supports), or does async-transition require a new hook-type enum value? The report names the transition but not the mechanism.

## §G. References

- Design authority: `.moai/reports/moai-autonomy-workflow-redesign-20260803.html` §3.5 (A8 / A10 / A11), §1.4 (hook·permission breakpoints), §3.1 (mode token → knob mapping).
- Stop wrapper: `.claude/hooks/moai/handle-stop-goal.sh` + `sync-phase-quality-gate.sh`.
- Per-edit hooks: `.claude/agents/moai/manager-develop.md` frontmatter (PreToolUse `develop-pre-implementation`, PostToolUse `develop-post-implementation`).
- Commit gate source: `internal/hook/pre_tool.go` IsGitCommit branch (~L429-441).
- Mode token source: `internal/config/envkeys.go` OR `internal/config/defaults.go` (OQ-1).
- Settings surface: `.claude/settings.json` (Stop hook registration).
- §3.1 invariant: deny/ask rules stay binding at every tier.

## §H. Acceptance Criteria (summary — full GWT in acceptance.md)

- AC-STOPCHAIN-TRIM-001 (A10): goal-absent session → no `moai hook stop-goal` exec.
- AC-STOPCHAIN-TRIM-002 (A10): non-sync HEAD → no `go vet`/`go build` from `sync-phase-quality-gate.sh`.
- AC-STOPCHAIN-TRIM-003 (A8): N-edit TDD cycle → 0 pre/post per-edit spawns (moved to per-cycle Stop).
- AC-STOPCHAIN-TRIM-004 (A11): `fully-autonomous` → sync-gate advisory only.
- AC-STOPCHAIN-TRIM-005 (A11): `automatic` → IsGitCommit gate OFF.
- AC-STOPCHAIN-TRIM-006 (A11): deny/ask rules still bind at every tier (regression guard).
- AC-STOPCHAIN-TRIM-007 (token): unset/empty `MOAI_AUTONOMY_TIER` = `semi-auto` (backward compat).
