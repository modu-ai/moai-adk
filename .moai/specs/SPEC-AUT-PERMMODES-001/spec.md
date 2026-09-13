---
id: SPEC-AUT-PERMMODES-001
title: "init wizard autonomy question redefined as Claude Code permission modes — acceptEdits default, auto/bypass opt-in, REQ-007 zero-delta re-scoped"
version: "0.1.1"
status: in-progress
created: 2026-09-13
updated: 2026-09-13
author: manager-spec
priority: P1
phase: "v3.2.1 target"
module: "internal/cli, internal/cli/wizard, internal/config, internal/core/project"
lifecycle: spec-anchored
tags: "autonomy-tier, permission-modes, acceptEdits, auto, bypassPermissions, defaultMode, req-007, req-006, wizard, downgrade-gate"
tier: M
related_specs: [SPEC-AUTONOMY-TIERS-001, SPEC-INIT-WIZARD-REPAIR-001, SPEC-CLI-WIZARD-RESTRUCTURE-001, SPEC-INIT-TUX-I18N-001]
---

# SPEC-AUT-PERMMODES-001 — init wizard autonomy question redefined as Claude Code permission modes

> Card: **t584** (operator-approved, from `.moai/reports/init-tui-audit-20260909.md` card-split table C2; the report lives in the primary checkout — `.moai/reports/` content is local-only). Operator decision 2026-09-09: the wizard's autonomy question presents Claude Code's REAL permission modes — three choices `bypassPermissions` / auto mode / `acceptEdits` ("accept edits on") — with **acceptEdits as the new default**.

## HISTORY

| Version | Date | Change | Provenance |
|---|---|---|---|
| "0.1.0" | 2026-09-13 | Initial plan-phase draft (spec.md + plan.md + acceptance.md + progress.md) | card t584, commit 0ddd1a282 |
| "0.1.1" | 2026-09-13 | Plan revision iter2 — AC-011/AC-012 added covering REQ-009/REQ-010 (plan-audit D1), §D.2 traceability rewritten; D2-D5 optional fixes (HISTORY section, quoted version, owning-SPEC line relabel, requirement subjects de-named) | plan-audit iter1 FAIL 0.75 (`.moai/reports/t584/plan-audit.md`) |

## §A. Background and Motivation

SPEC-AUTONOMY-TIERS-001 shipped a 3-tier autonomy selector (semi-auto / automatic / fully-autonomous) whose tier→knob mapping (`TierDefaultMode`, internal/config/autonomy_tiers.go) writes `defaultMode` values `"default"` / `"auto"` / `"bypassPermissions"`. Two problems surfaced in the 2026-09-09 init/update audit:

1. **The labels are MoAI-invented vocabulary, not Claude Code's.** Users see "semi-auto / automatic / fully-autonomous" in the wizard but the values landed in `.claude/settings.json` are permission-mode tokens. The operator directed the question to speak Claude Code's language directly.
2. **The zero-delta invariant is now wrong-shaped.** REQ-007 of SPEC-AUTONOMY-TIERS-001 pins "unset / semi-auto selection → renderer output byte-identical, zero behavior delta (no file written)". With acceptEdits as the new default, the default selection MUST write a permission record (USER-scope `defaultMode: "acceptEdits"`), so the invariant as worded is unimplementable — it must be re-scoped, not silently broken.

### defaultMode token verdict (plan-phase research, Evidence)

**`auto` IS a valid `permissions.defaultMode` value in current Claude Code.** The current official permission-modes table (fetched 2026-09-13, `https://code.claude.com/docs/en/permissions` § Permission modes) lists SIX values:

| Token | Meaning (per official docs) |
|---|---|
| `default` | Prompts per tool first use; CLI label "Manual"; alias `manual` (CC v2.1.200+) |
| `acceptEdits` | Auto-accepts file edits + common filesystem commands in the working directory |
| `plan` | Read-only exploration |
| `auto` | Auto-approves tool calls with background safety checks (classifier) |
| `dontAsk` | Auto-denies what would otherwise prompt |
| `bypassPermissions` | Skips prompts (except actions no mode auto-approves) |

Kill switches: `permissions.disableBypassPermissionsMode` AND `permissions.disableAutoMode`, both settable to `"disable"` in any settings file.

**The repo's own bundled reference is stale**: `internal/template/templates/.claude/skills/moai-foundation-cc/reference/claude-code-iam-official.md:30-40` states "exactly four values" (default/plan/acceptEdits/bypassPermissions) and explicitly denies `auto`. That claim is outdated relative to the current official docs. This SPEC refreshes that reference as part of scope (REQ-007 below).

### The three wizard choices → settings values mapping

| Wizard choice (user-facing) | `permissions.defaultMode` written | Persisted tier token (unchanged) | Gating |
|---|---|---|---|
| Accept edits on (**default**, Recommended) | `"acceptEdits"` | `semi-auto` | none |
| Auto mode | `"auto"` | `automatic` | none |
| Bypass permissions | `"bypassPermissions"` | `fully-autonomous` | sandbox proof + kill-switch downgrade (unchanged) |

Internal tier tokens (`semi-auto` / `automatic` / `fully-autonomous`, internal/config/defaults.go:161-163) stay as the persisted `workflow.autonomy_tier` values — zero config migration; only labels, the knob mapping, and the zero-delta wording change.

## §B. Requirements (EARS)

### REQ-001 — Wizard question presents Claude Code permission modes

The init wizard autonomy question SHALL present exactly three options labeled with Claude Code permission-mode vocabulary, with values unchanged from the existing tier tokens:

- "Accept edits on" (Recommended) → value `semi-auto`
- "Auto mode" → value `automatic`
- "Bypass permissions" → value `fully-autonomous`

### REQ-002 — acceptEdits is the default selection

The wizard question SHALL pre-select "Accept edits on" (`Default: "semi-auto"`, label carrying the Recommended signal). REQ-006 of SPEC-AUTONOMY-TIERS-001 is AMENDED, not violated: its invariant "the wizard MUST NOT default to the highest tier" is preserved — the default moves from semi-auto→`default` to semi-auto→`acceptEdits`, and `bypassPermissions` remains never-default and never-preselected.

### REQ-003 — Tier→knob mapping remap

The tier-to-mode mapping SHALL map: `semi-auto` → `"acceptEdits"` (was `"default"`); `automatic` → `"auto"` (unchanged); `fully-autonomous` → `"bypassPermissions"` (unchanged). Any unknown value SHALL still map to `"default"` (the fail-safe never-silently-enable guarantee is retained at the MOST restrictive mode, not the new default).

### REQ-004 — REQ-007 re-scoped (bounded delta, not zero delta)

REQ-007 of SPEC-AUTONOMY-TIERS-001 ("unset / semi-auto → zero behavior delta, no file written") is RE-SCOPED: the unset / `semi-auto` selection SHALL write ONLY the USER-scope `defaultMode: "acceptEdits"` record and NOTHING else — the PROJECT-scope `allow`/`ask`/`deny` arrays, the deployed template files, and every other file MUST remain byte-identical. The sanctioned delta is exactly one JSON key in exactly one file. Deny/ask tier-invariance (REQ-004 of the owning SPEC) is unchanged.

### REQ-005 — bypassPermissions gating preserved

The fully-autonomous gating SHALL remain anchored to `bypassPermissions`: when a user selects "Bypass permissions" WITHOUT a sandbox proof, or with the kill-switch active, the selection SHALL downgrade to `automatic` (`defaultMode: "auto"`) and SHALL append an advisory record to `.moai/logs/autonomy-downgrade.log`. The gating logic's behavior is otherwise unchanged.

### REQ-006 — Downgrade regression tests preserved

The bypass-selection downgrade regression tests SHALL be kept under the new option set (adapted, not deleted): `TestApplyAutonomyTierBundle_FullyAutonomousDowngradedWithoutProof`, `TestApplyAutonomyTierBundle_FullyAutonomousWithProofDeploysBypass`, `TestRunInit_FlagFullyAutonomousWithoutProofDowngrades`, `TestAppendDowngradeAdvisory`, `TestAutonomyTierQuestion_FullyAutonomousNotRecommended`.

### REQ-007 — Bundled reference doc refreshed to the current enum

The bundled Claude Code IAM reference (`internal/template/templates/.claude/skills/moai-foundation-cc/reference/claude-code-iam-official.md` § Permission Modes) SHALL be updated to the current six-value enum, stating that `auto` and `dontAsk` are valid values and noting the kill-switch pair (`disableBypassPermissionsMode` + `disableAutoMode`). The template mirror rule applies (edit the template source; `.claude/` copy follows `moai update`).

### REQ-008 — Option labels localized

The three option labels and descriptions SHALL be localized across the four wizard locales (ko / en / ja / zh) in `internal/cli/wizard/translations.go`, replacing the retired "Semi-auto / Automatic / Fully-autonomous" strings. The translation completeness test must stay green.

### REQ-009 — Flag surface aligned

The `--autonomy-tier` flag help text (internal/cli/init.go:126) SHALL be updated to describe the three permission-mode choices and the new default. The closed-set validation (`ValidateAutonomyTierSelection`) SHALL be unchanged — the persisted token set does not move.

### REQ-010 — Doc comments state the new mapping

The godoc on `TierDefaultMode` and `ApplyAutonomyTierBundle` SHALL state the new mapping and the re-scoped REQ-004 invariant (English comments, per `code_comments` config).

## §C. Design Notes (WHAT/WHY boundaries)

- The question shape (select, one of three, grouped in the "Agents & Autonomy" page from t586's restructure) does not change. Only labels, descriptions, the knob mapping, and the delta policy change.
- The downgrade target stays `automatic` (→ `auto`), NOT `semi-auto` (→ `acceptEdits`): a user who reached for bypass was asking for prompt-free execution; `auto` preserves that intent under classifier checks.
- The stale-reference root cause is recorded: the bundled IAM reference was authored when CC shipped 4 modes; the "auto" token pre-dated CC's own `auto` mode. Future CC enum drift is caught by re-checking the official permissions page at run time of any future mode-mapping change.

## §D. Out of Scope

### Out of Scope — plan / dontAsk modes

- The wizard does not offer `plan` or `dontAsk`; they remain valid CC modes unused by MoAI's question.

### Out of Scope — disableAutoMode kill switch wiring

- `permissions.disableAutoMode` (new CC kill switch) is documented in REQ-007's refresh but is NOT wired into `EffectiveTierWithGates`; gating stays bypass-anchored per the operator decision.

### Out of Scope — migration of already-deployed projects

- Projects initialized before this change keep whatever `defaultMode` they have; `moai update` re-applies the bundle only per its existing snapshot rules, and no one-time migration writes a new default.

### Out of Scope — web console tier toggle redesign

- `moai web`'s tier toggle rows (TierToggleOptions) keep their existing labels this card; only the init wizard question is redefined.

## §H. Cross-References

- Owning SPEC of the amended invariants: `.moai/specs/SPEC-AUTONOMY-TIERS-001/spec.md` (AC-AUTONOMY-TIERS-006 at :167, AC-AUTONOMY-TIERS-007 at :168 — the AC summary rows carrying the REQ-006/REQ-007 invariants; the REQ rows live in that SPEC's §B EARS body)
- Post-t586 wizard structure: `internal/cli/wizard/questions.go` (autonomy_tier question), `internal/cli/wizard/translations.go`
- Bundle apply: `internal/core/project/autonomy_bundle.go` (`ApplyAutonomyTierBundle`), `internal/cli/update_settings_snapshot.go` (update-path consumer)
- Tier mapping + gates: `internal/config/autonomy_tiers.go` (`TierDefaultMode`, `EffectiveTierWithGates`, `AppendDowngradeAdvisory`), `internal/config/defaults.go:161-163` (tier tokens)
- Audit source: `.moai/reports/init-tui-audit-20260909.md` (primary checkout; card C2, operator decision §운영자 결정 반영)
