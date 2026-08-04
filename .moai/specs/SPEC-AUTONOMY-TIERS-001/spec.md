---
id: SPEC-AUTONOMY-TIERS-001
title: "Autonomy 3-mode selection + behavior wiring on the MOAI_AUTONOMY_TIER token"
version: 0.1.0
status: completed
created: 2026-08-04
updated: 2026-08-05
sync_commit_sha: pending-backfill-sync
author: manager-spec
priority: P1
phase: "v3.x target"
module: config-autonomy
lifecycle: spec-anchored
tags: "autonomy-tiers, autonomy, semi-auto, automatic, fully-autonomous, defaultmode, deny-ask, kill-switch, autonomy-epic"
tier: M
amendment_of: SPEC-AUTONOMY-TIERS-001
related_specs: [SPEC-STOPCHAIN-TRIM-001, SPEC-INFINITE-GOAL-001]
---

# SPEC-AUTONOMY-TIERS-001 — Autonomy 3-mode selection + behavior wiring

## HISTORY

- 2026-08-04 — Initial draft. Codifies §3.1 of the autonomy-workflow redesign report (`moai-autonomy-workflow-redesign-20260803.html`): the user-facing 3-mode selection system (`semi-auto` / `automatic` / `fully-autonomous`) and the behavior bundle each tier drives. First P1 of the autonomy-workflow epic. **Builds ON** the `MOAI_AUTONOMY_TIER` mode token that `SPEC-STOPCHAIN-TRIM-001` already landed (token + Go reader `AutonomyTier()` + env-key constant + 3-value enum) — this SPEC delivers the user-facing mode selection (moai init / moai web), the tier→permission-bundle mapping, and the manager kill-switch on top of that token. The token itself is NOT re-scoped.

### Amendments

- 2026-08-04 — In-place amendment (S1).
  - **Prior status**: completed.
  - **Prior completed SHA**: `4ab512c89` (Automerge: feat(SPEC-AUTONOMY-TIERS-001)) — resolved from `progress.md` §E.4 `sync_commit_sha` field (the spec.md frontmatter `sync_commit_sha: pending-backfill` was not backfilled, so progress.md was the authoritative source per the amendment task guidance).
  - **Rationale**: S1 — correct §C deny/ask binding claim. The §C table row + "Key invariant" paragraph claimed the deny/ask rule set is "ALL tiers binding" and the protected ops are "always ask". This is factually wrong under `bypassPermissions` (the fully-autonomous `defaultMode`): `ask` rules are treated as auto-approved under `bypassPermissions` (the prompt is skipped because all prompts are skipped) and bind only at `semi-auto`/`automatic`; only `deny` rules mechanically bind at every tier. The fully-autonomous safety boundary is the **verified sandbox** (S3, REQ-002 / OQ-1 / OQ-4) + the `deny` list, NOT `ask`. Identified by /moai review. Corroborated internally by `.claude/rules/moai/development/coding-standards.md` §Bash Risk-Amplifier Doctrine (3) — the doctrine has to impose an agent-side "explicit confirmation EVEN in `bypassPermissions`" obligation precisely because the runtime does NOT enforce `ask` under bypass.
  - **Scope**: §C table row + Key invariant wording correction (split into §C row + §C.1 binding-semantics note); matching §A User Story line correction; AC-004 clarification (deny arrays byte-identical / ask arrays byte-identical across the 3 rendered tiers). NO REQ change — REQ-004 protected-set enumeration unchanged (the rule set stays tier-invariant; only the per-tier binding semantics are clarified). Code/YAML implementation fixes (S2 tool-policy.yaml rules + S3 `autonomy_tiers.go` `SandboxProofKind` hardening) are NOT in this amendment's scope — they are delivered alongside in run-phase by manager-develop.
- 2026-08-05 — In-place amendment (S4).
  - **Prior status**: completed (post-S1).
  - **Rationale**: S4 — correct §C.1 ask-under-bypass framing + retire the S3 Linux fingerprint cross-check. Verified against the official Claude Code docs (`code.claude.com/docs/en/permissions` + `/sandbox-environments`, 2026-08-05): content-scoped `ask` rules (e.g. `Bash(git push *)`) **still prompt under `bypassPermissions`** — only the default per-tool prompt and bare-tool ask are skipped. The S1 claim that "`ask` is auto-approved under bypass" was categorical and inaccurate. Separately, the S3 `SandboxProofKind` Linux-only container-fingerprint cross-check was a false gate: it could only validate Docker/Podman, false-rejected the other 6 kinds (gvisor/firecracker/e2b/devcontainer/kata) on non-container hosts (CI runners), and gave weak real security ("any container passes"). The allowlist already achieves S3's spoof-rejection goal (reject `=1`/`=foo`; accept only known isolation kinds), so the fingerprint gate is retired and the marker is treated as the attestation it is, with a stderr advisory on every platform.
  - **Scope**: §C.1 binding-semantics paragraph refined (deny always binds; content-scoped ask survives bypass; effective safety = sandbox + deny + content-scoped ask). `internal/config/autonomy_tiers.go` `SandboxProofKind` rewritten to allowlist-as-proof (Linux fingerprint branch + `hasContainerFingerprint` removed; advisory emitted on all platforms). `internal/config/defaults.go` `SandboxProofKinds` extended with `sandbox-runtime` (an official Claude Code full-process isolation option). §A line 44 "ask advisory under bypass" softened to match. NO REQ change — REQ-002 (proof required) and REQ-004 (deny/ask enumeration) are unchanged; only the proof *mechanism* and the ask-binding *detail* are corrected.

## §A. User Story

**As a** MoAI user choosing how much autonomy to grant a session,
**I want** to select an autonomy tier (`semi-auto` default / `automatic` general-purpose / `fully-autonomous` sandbox-only) via the `moai init` interactive wizard or the `moai web` console toggle, and have that single selection drive a coherent bundle of native Claude Code primitives (`defaultMode` + `deny`/`ask` permission lists) written to the correct config scope (USER vs PROJECT),
**so that** "one selection = one consistent behavior bundle" — I no longer have to hand-wire `defaultMode`, `deny`, and `ask` arrays in the right files and hope they compose; the tier selection does it for me, the dangerous tier stays opt-in and sandbox-gated, and a maintainer kill-switch can disable `bypassPermissions` enterprise-wide.

**Outcome hypotheses (from §3.1):**

- `semi-auto` (default): zero behavior delta vs today. `defaultMode: default` in USER settings, full sync-gate blocking, commit gate ON, allowlist-only Bash = per-tool prompt. Backward-compat invariant: a session that does not opt in pays zero behavior change.
- `automatic` (general-purpose recommended): `defaultMode: auto`, commit gate OFF (destructive-pattern deny still binds), sync-gate build-only-block, classifier safety net with fallback. The daily Tier S/M work tier.
- `fully-autonomous` (sandbox-only, opt-in): `defaultMode: bypassPermissions` + sandbox/container proof required, sync-gate advisory-only, commit gate OFF, subagent lifecycle hooks dormant — with `deny` rules mechanically binding at every tier (non-bypassable) and content-scoped `ask` rules still prompting under `bypassPermissions` while the default per-tool prompt is skipped (see §C.1 binding note, S4-refined); the fully-autonomous safety boundary is the verified sandbox + the deny list + content-scoped ask. The unattended-loop tier.

The mode-token → knob mapping (§3.1 table) makes each tier a coherent combination of EXISTING knobs — no new mechanism is created; the user-facing selector + the tier→bundle renderer + the kill-switch are the net-new surfaces.

## §B. Scope

**In scope — 3-mode SELECTION + BEHAVIOR WIRING on the existing token, from §3.1:**

- A user-facing tier selector in `moai init` (interactive wizard page) and `moai web` (console toggle), with non-interactive `--autonomy-tier` flag parity (mirrors the existing `--profile` / `--harness-profile` pattern).
- A tier→bundle mapping renderer that writes the native primitive bundle per tier: `defaultMode` to USER settings (`auto` / `bypassPermissions` are USER-scope per Claude Code v2.1.142+, PROJECT/local cannot override them downward), and `deny`/`ask` rule sets to PROJECT settings (merged across all sessions). The bundle is the §3.1 mode-token → knob table codified as a renderer, NOT a new mechanism.
- A `disableBypassPermissionsMode` manager kill-switch gate: when the managed/enterprise config disables bypass, the `fully-autonomous` tier is unselectable in every surface (selector greyed-out, `--autonomy-tier fully-autonomous` rejected, web toggle disabled), and an existing session that somehow carries `bypassPermissions` is downgraded to `automatic`-equivalent behavior. This is the Claude Code documented enterprise kill-switch wired into the tier system.
- **OQ-1 (G7) — fully-autonomous sandbox proof**: the recommended option (CLI flag `--fully-autonomous` requiring sandbox/container attestation; web selector DISABLES fully-autonomous unless a sandbox proof is present; deny/ask pins irreversible actions; manager kill-switch gates the whole tier) is captured here with the decision DEFERRED to Implementation Kickoff.

**Out of scope — sibling epic items:** the `MOAI_AUTONOMY_TIER` mode token, the Go reader, the env-key constant, and the 3-value enum (owned by `SPEC-STOPCHAIN-TRIM-001` — already landed); the goal-evaluator HTML dashboard and `moai goal render` (sibling P1 `SPEC-GOAL-HTML-FLOW`); the stateful MCP tool layer (sibling P2 `SPEC-MOAI-MCP-SERVER`); the hierarchical team leader (sibling P2 `SPEC-HIERARCHICAL-TEAM`); the 3-way audit model (sibling P1 `SPEC-AUDIT-MULTI-MODEL`).

### Out of Scope — The mode token itself

- The `MOAI_AUTONOMY_TIER` env-key constant (`internal/config/envkeys.go:108`), the 3-value enum (`internal/config/defaults.go:107-109`), the `AutonomyTier()` reader (`internal/config/autonomy.go`), and the `IsAutonomyTierCommitGateOff` / `IsAutonomyTierLifecycleDormant` predicates — ALL owned by `SPEC-STOPCHAIN-TRIM-001` (status: completed). This SPEC CONSUMES those surfaces to drive the selector + renderer; it does NOT redefine them.
- The mode-aware HOOK behavior (sync-gate advisory-ness, commit gate on/off, subagent lifecycle dormancy) — also owned by `SPEC-STOPCHAIN-TRIM-001`. This SPEC's renderer emits the tier selection that the hooks then read; the hook branching is NOT re-scoped.

### Out of Scope — Sibling P1/P2 epic items

- The goal-evaluator HTML dashboard, `moai goal render`, plan-phase HTML report, resume auto-rearm UI — sibling P1 `SPEC-GOAL-HTML-FLOW`.
- The hierarchical team leader role, worktree-isolated writer re-keying, milestone Context-Folding — sibling P2 `SPEC-HIERARCHICAL-TEAM`.
- The 3-way (claude/codex/glm) audit model, gate 3-stage (off/advisory/required), authentication branching — sibling P1 `SPEC-AUDIT-MULTI-MODEL`.
- The stateful MCP tool layer (`moai_goal_arm`, `moai_verify_snapshot`) and trend MCP tools — sibling P2 `SPEC-MOAI-MCP-SERVER` / `SPEC-TREND-MCP`.

### Out of Scope — Template-default fully-autonomous

- The distributed template (`internal/template/templates/**`) SHALL NOT ship `fully-autonomous` as a default or a recommended selection. `fully-autonomous` is `settings.local.json` / explicit-`--fully-autonomous`-flag opt-in ONLY. Template neutrality (CLAUDE.local.md §15 / §25) is preserved: the selector offers the tier, the template does not pre-pick the dangerous one.

## §C. Context — mode-token → knob mapping (from §3.1)

The §3.1 table, restated here as the renderer's contract. Each tier is a coherent combination of EXISTING knobs (no new mechanism):

| Knob | semi-auto (default) | automatic (recommended) | fully-autonomous (sandbox-only) |
|------|---------------------|-------------------------|---------------------------------|
| `defaultMode` (USER setting) | `default` | `auto` | `bypassPermissions` + sandbox proof |
| sync-gate mode | full-blocking (vet+build+lint) | build-only-block | advisory (`systemMessage` only) |
| commit gate (`pre_tool.go` IsGitCommit) | ON | OFF | OFF |
| subagent lifecycle (SubagentStop/TeammateIdle) | blocking | blocking | dormant |
| deny/ask rules (PROJECT) — see §C.1 binding note | identical rule set across tiers — main push / force-push / secrets / `rm -rf` / prod deploy / `terraform destroy` / IAM grant | identical | identical |

Key invariant (§3.1): the deny/ask **rule set** is tier-INVARIANT — no tier weakens a `deny` rule or drops a protected op from the enumerated set (REQ-004 binds this enumeration). The **binding semantics** differ by decision level: `deny` rules bind at every tier (mechanically enforced, including under `bypassPermissions`); `ask` rules bind at `semi-auto`/`automatic` and are **advisory** under `fully-autonomous` (`bypassPermissions` skips ask prompts — the prompt is auto-approved, not blocked). The fully-autonomous safety boundary is the **verified sandbox** (OQ-1 / OQ-4 sandbox proof, REQ-002) + the `deny` list, NOT `ask`. The tier's effect is limited to `defaultMode`, sync-gate mode, commit gate on/off, and subagent lifecycle dormancy — exactly the four surfaces above. The PermissionRequest handler stays a no-op (it is NOT a lever — `defaultMode` + allowlist + PreToolUse are the real levers, per §3.1).

### §C.1 Binding-semantics note (deny vs ask under `bypassPermissions`)

Claude Code permission semantics (verified against the official docs — `code.claude.com/docs/en/permissions` + `/sandbox-environments`, 2026-08-05): rules evaluate in order **deny → ask → allow** in every mode, and the first match decides. `deny` rules are mechanically enforced at every tier (block, non-bypassable — including under `bypassPermissions`; this is the one list that hard-blocks under `fully-autonomous`). `allow` rules are always auto-approved. `ask` splits by scope: a **content-scoped** ask rule like `Bash(git push *)` **still forces a prompt even under `bypassPermissions`** (`--dangerously-skip-permissions` skips only the default per-tool prompt and bare-tool ask, NOT explicit content-scoped ask rules; connector tools set to `ask` and MCP `requiresUserInteraction` tools also still prompt). This refines the prior S1 framing: it is NOT accurate that "`ask` is auto-approved under bypass" categorically — only the default/bare-tool prompt is skipped. This is corroborated internally by `.claude/rules/moai/development/coding-standards.md` §Bash Risk-Amplifier Doctrine (3), which imposes an agent-side "explicit confirmation EVEN in `bypassPermissions`" obligation for destructive primitives — complementary defense for the surface bare-tool/default prompts do NOT cover. Consequence: effective `fully-autonomous` safety = **the verified sandbox + the `deny` list + content-scoped `ask` rules**; the sandbox remains defense-in-depth for what no permission rule gates (prompt-injection, runaway, ungated tools). The rule-set enumeration (REQ-004) is unchanged.

The hook-side branching of sync-gate mode / commit gate / subagent dormancy is OWNED by `SPEC-STOPCHAIN-TRIM-001` (already implemented). This SPEC's contribution is the user-facing SELECTION and the tier→bundle RENDERER that writes the `defaultMode` + deny/ask bundle; the hook behavior is downstream consumption.

## §D. Requirements (GEARS)

### REQ-AUTONOMY-TIERS-001 — Tier selector in `moai init` (interactive + non-interactive)

**Where** the `moai init` interactive wizard runs, the wizard SHALL present an autonomy-tier selection page offering the 3 canonical values `{semi-auto, automatic, fully-autonomous}` with `semi-auto` pre-selected as the default. **Where** the wizard runs non-interactively (`--non-interactive` flag), the wizard SHALL accept a `--autonomy-tier <value>` flag with the same 3-value closed-set validation used by `config.AutonomyTier()` (case-insensitive, whitespace-trimmed, invalid → error, NOT silent fallback — the selector is the user-facing surface, fail-loud differs from the reader's fail-safe).

**When** the user selects a tier (or supplies `--autonomy-tier`), the wizard SHALL persist the selection via the M1 tier-persistence surface. The persisted selection SHALL be the single source the renderer (REQ-003) and the hooks (downstream, STOPCHAIN-TRIM) both consume. (The persistence-surface choice — env-key-only vs a new `workflow.yaml` key — is OQ-3, tracked in plan.md §B / §F; the REQ binds behavior, not the surface choice.)

### REQ-AUTONOMY-TIERS-002 — Tier selector in `moai web` (console toggle)

**Where** the `moai web` console is active, the console SHALL expose a tier toggle offering `{semi-auto, automatic, fully-autonomous}`. **When** the user toggles to `fully-autonomous`, the console SHALL require a sandbox/container proof (OQ-1 / OQ-4: the attestation mechanism — CLI flag, env marker, container fingerprint) BEFORE the toggle commits; **when** no proof is present, the console SHALL disable the `fully-autonomous` option (greyed-out, not silently selectable). **When** the `disableBypassPermissionsMode` kill-switch (REQ-005) is engaged, the console SHALL disable the `fully-autonomous` option regardless of sandbox proof.

### REQ-AUTONOMY-TIERS-003 — Tier → permission-bundle renderer (USER vs PROJECT scope)

**When** the renderer is invoked with a tier value, the renderer SHALL write the native primitive bundle per tier to the correct config scope: `permissions.defaultMode` to the USER-scope settings file (so `auto` / `bypassPermissions` take effect and cannot be weakened by PROJECT/local — Claude Code v2.1.142+), and the `deny` / `ask` rule arrays to the PROJECT-scope settings file (merged across all sessions). **Where** the tier is `semi-auto`, the renderer SHALL write `defaultMode: default` and the canonical deny/ask rule set. **Where** the tier is `automatic`, the renderer SHALL write `defaultMode: auto` and the SAME deny/ask rule set. **Where** the tier is `fully-autonomous`, the renderer SHALL write `defaultMode: bypassPermissions` and the SAME deny/ask rule set — the deny/ask invariance (§3.1) is preserved at every tier.

**The renderer SHALL reuse the existing toolpolicy codegen surface, not invent a parallel writer** — the tier→bundle mapping is a NEW CALLER of the existing codegen, NOT a new codegen. (Codegen-pointer detail — `BuildInto` / `BuildIntoAuto` signatures + the existing `--default-mode` flag site — is recorded in plan.md §A / §C Pre-flight §3; it is implementation detail, not a normative REQ concern.)

### REQ-AUTONOMY-TIERS-004 — deny/ask invariance (cross-cutting)

**The deny/ask rule set (main push, force-push, secrets, `rm -rf` on project paths, prod deploy, `terraform destroy`, IAM grant) SHALL remain identical across all 3 tiers.** No tier selection SHALL weaken a deny rule or downgrade an ask rule. The tier's effect is LIMITED to: (a) `defaultMode` value, written per REQ-003. The sync-gate mode / commit gate / subagent dormancy effects are OWNED by `SPEC-STOPCHAIN-TRIM-001` and are downstream of the tier selection this SPEC produces — they are NOT re-scoped here. REQ-AUTONOMY-TIERS-004 binds ONLY the renderer's deny/ask emission.

### REQ-AUTONOMY-TIERS-005 — Manager kill-switch (`disableBypassPermissionsMode`)

**Where** the managed/enterprise config sets `disableBypassPermissionsMode: true` (the Claude Code documented enterprise kill-switch), the tier system SHALL gate the `fully-autonomous` tier in every surface: the `moai init` selector SHALL reject `--autonomy-tier fully-autonomous` with an error citing the kill-switch, the `moai web` toggle SHALL disable the `fully-autonomous` option, and the renderer (REQ-003) SHALL refuse to write `defaultMode: bypassPermissions`. **When** an existing session carries `bypassPermissions` (e.g. a pre-existing `settings.local.json`) and the kill-switch is engaged, the system SHALL downgrade the effective behavior to `automatic`-equivalent (`defaultMode: auto`) and emit an advisory recording the downgrade. The kill-switch does NOT affect `semi-auto` or `automatic`.

### REQ-AUTONOMY-TIERS-006 — `fully-autonomous` opt-in / local-only (template neutrality)

**The distributed template (`internal/template/templates/**`) SHALL NOT ship `fully-autonomous` as a default, recommended, or pre-selected tier.** The template's `moai init` default page SHALL pre-select `semi-auto`. **Where** a user wants `fully-autonomous`, the user SHALL opt in explicitly via the `moai init` selector, the `--autonomy-tier fully-autonomous` flag, the `moai web` toggle (with sandbox proof), or a manual `settings.local.json` entry. This constraint preserves template neutrality (CLAUDE.local.md §15 / §25): the selector OFFERS the tier; the template does not PRE-PICK the dangerous one.

### REQ-AUTONOMY-TIERS-007 — Backward compatibility (zero behavior delta at unset / `semi-auto`)

**When** no tier selection is persisted (no `--autonomy-tier`, no `moai web` toggle, no `workflow.yaml` key per OQ-3, no env override), the system SHALL behave identically to `semi-auto` — `defaultMode: default`, full sync-gate, commit gate ON, allowlist-only Bash. **Where** the persisted tier is `semi-auto`, the renderer SHALL produce a settings block byte-identical (modulo whitespace) to today's template output. A session that does not opt in pays zero behavior delta (mirrors `SPEC-STOPCHAIN-TRIM-001` REQ-003's backward-compat invariant, extended to the selector surface).

## §E. Assumptions

1. The existing `moai init` wizard supports adding a new interactive page (the wizard is paginated; the `--profile` and `--harness-profile` pages are existing precedents at `internal/cli/init.go:81-112`), and the `--autonomy-tier` non-interactive flag slots in alongside `--profile` without a wizard refactor.
2. The existing `internal/config/toolpolicy` codegen (`BuildInto` / `BuildIntoAuto` at `internal/config/toolpolicy/codegen.go:204-270`) accepts a `defaultModeOverride` argument already; the tier→bundle renderer is a new caller passing the tier's `defaultMode` as the override, NOT a codegen rewrite.
3. USER-scope vs PROJECT-scope settings files are distinguishable by path (the existing `internal/config/resolver.go` / `source.go` carry the scope distinction); the renderer writes `defaultMode` to the USER path and deny/ask to the PROJECT path without inventing a new scope concept.
4. `disableBypassPermissionsMode` is a Claude Code runtime-recognized settings key (per Claude Code Settings docs + §3.1 risk callout); writing it to the managed/enterprise config gates `bypassPermissions` enterprise-wide without a MoAI-side enforcement shim.
5. The `moai web` console has (or can gain) a toggle-control rendering surface; the tier toggle slots in alongside existing console controls without a console architecture change.
6. A "sandbox/container proof" signal (OQ-1 / OQ-4) is reachable from Go (env marker, cgroup inspection, or a `--sandbox-proof` attestation flag) without a container-runtime dependency that would break the non-containerized path.

## §F. Open Questions (for plan-auditor + Implementation Kickoff)

- **OQ-1 (G7 from the report)** — fully-autonomous sandbox-proof method: CLI flag `--fully-autonomous` requiring sandbox/container attestation vs `moai web` exposure. **Recommended option**: CLI flag `--fully-autonomous` + sandbox proof; web selector DISABLES fully-autonomous unless a sandbox proof is present; deny/ask pins irreversible actions; manager kill-switch (REQ-005) gates the whole tier. **Decision DEFERRED to Implementation Kickoff** (this is the load-bearing G7 blocker the report flags red).
- **OQ-2** — the concrete permission-bundle DELTA between tiers: beyond `defaultMode`, what (if anything) differs in the deny/ask arrays between `semi-auto` / `automatic` / `fully-autonomous`? REQ-004 asserts the deny/ask is IDENTICAL across tiers (the §3.1 invariant); OQ-2 asks whether the report's "classifier safety net" and "stop-goal type:agent promotion" at `automatic` imply a deny/ask delta or only a hook-side delta (already owned by STOPCHAIN-TRIM). **Recommended**: deny/ask identical; the delta is hook-side only (downstream of STOPCHAIN-TRIM).
- **OQ-3** — tier persistence surface: STOPCHAIN-TRIM resolved the hooks' read path to the env-key (shell hooks need `$MOAI_AUTONOMY_TIER` without the binary). But the SELECTOR needs a WRITABLE persistence surface. Is it (a) env-only (the selector writes to `settings.local.json` `env` block), (b) a new `workflow.yaml` key that the launcher then exports to env, or (c) both? **Recommended**: `workflow.yaml` key as the persisted user selection, launcher exports it to the env-key at session start (shell hooks keep reading the env-key unchanged). Decision affects M1.
- **OQ-4** — sandbox/container attestation mechanism (the "proof" of OQ-1): env marker (`MOAI_SANDBOX=1` set by a known launcher), cgroup inspection, a `--sandbox-proof <kind>` CLI attestation flag, or a container fingerprint. **Recommended**: a documented env marker (`MOAI_SANDBOX_PROOF=<kind>` set by container/devcontainer launchers MoAI recognizes) + a `--sandbox-proof` CLI flag for manual attestation, with the kind recorded in the audit log. Decision affects REQ-002 + REQ-005 gating.

## §G. References

- Design authority: `.moai/reports/moai-autonomy-workflow-redesign-20260803.html` §3.1 (autonomy 3-mode — selector + USER/PROJECT split + kill-switch), §6 risk row 2 (fully-autonomous + bypassPermissions security), §6 open question 1 (G7).
- Token foundation (consumed, NOT re-scoped): `internal/config/autonomy.go` (`AutonomyTier()` reader + `IsAutonomyTierCommitGateOff` / `IsAutonomyTierLifecycleDormant`), `internal/config/envkeys.go:99-108` (`EnvAutonomyTier`), `internal/config/defaults.go:94-109` (3-value enum + `semi-auto` fallback).
- Sibling SPEC format reference: `.moai/specs/SPEC-STOPCHAIN-TRIM-001/spec.md` (prerequisite; mirror REQ-STOPCHAIN-TRIM-007 deny/ask invariance framing).
- Init wizard surface: `internal/cli/init.go` (paginated wizard, `--non-interactive` + `--profile` / `--harness-profile` flag precedents at L81-112).
- Web console surface: `internal/cli/web.go`, `internal/cli/web_port.go`.
- Tool-policy codegen (renderer target): `internal/config/toolpolicy/codegen.go` (`BuildInto` / `BuildIntoAuto` + `defaultModeOverride` at L204-270), `internal/config/toolpolicy/settings_region.go` (`defaultMode` + `allow`/`ask`/`deny` modeling), `internal/cli/tool_policy.go:67` (existing `--default-mode` flag).
- Config scope (USER vs PROJECT): `internal/config/resolver.go`, `internal/config/source.go`.
- Template neutrality: CLAUDE.local.md §15 (16-language neutrality) + §25 (template internal-content isolation); `internal/template/templates/.claude/settings.json` is the rendered surface whose `defaultMode` the renderer writes.

## §H. Acceptance Criteria (summary — full GWT in acceptance.md)

- AC-AUTONOMY-TIERS-001 (REQ-001): `moai init` wizard offers 3-tier selection; `--autonomy-tier` flag validates the closed set.
- AC-AUTONOMY-TIERS-002 (REQ-002): `moai web` toggle offers 3 tiers; fully-autonomous disabled without sandbox proof; disabled under kill-switch.
- AC-AUTONOMY-TIERS-003 (REQ-003): renderer writes `defaultMode` to USER scope, deny/ask to PROJECT scope, per tier.
- AC-AUTONOMY-TIERS-004 (REQ-004): deny/ask arrays byte-identical across the 3 rendered tiers.
- AC-AUTONOMY-TIERS-005 (REQ-005): `disableBypassPermissionsMode` rejects fully-autonomous in all surfaces; existing bypass session downgrades to auto + advisory.
- AC-AUTONOMY-TIERS-006 (REQ-006): template ships `semi-auto` default; fully-autonomous is opt-in only.
- AC-AUTONOMY-TIERS-007 (REQ-007): unset / `semi-auto` selection → renderer output byte-identical to today's template (zero behavior delta).
