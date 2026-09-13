# SPEC-AUT-PERMMODES-001 — Progress

> Card t584. Evidence path for the dispatch: `.moai/specs/SPEC-AUT-PERMMODES-001/progress.md`.

## §E.1 Plan-phase Audit-Ready Signal

```yaml
phase: plan
spec: SPEC-AUT-PERMMODES-001
card: t584
tier: M
artifacts: [spec.md, plan.md, acceptance.md]
req_count: 10
ac_count: 10
id_regex_check: PASS
owning_spec_of_amended_reqs: SPEC-AUTONOMY-TIERS-001 (REQ-006, REQ-007)
defaultmode_verdict: "auto is a VALID CC defaultMode value (6-value enum, official docs fetched 2026-09-13); bundled IAM reference stale -> refreshed in-scope (REQ-007)"
open_blockers: none
```

Plan-phase research notes (attributable):

- SPEC ID regex check `SPEC-AUT-PERMMODES-001` → PASS (verbatim output in plan-phase session).
- Official permission-modes table fetched 2026-09-13 from `https://code.claude.com/docs/en/permissions` (§ Permission modes): `default` (alias `manual`, CC v2.1.200+) / `acceptEdits` / `plan` / `auto` / `dontAsk` / `bypassPermissions`; kill switches `permissions.disableBypassPermissionsMode` + `permissions.disableAutoMode`.
- Stale-reference evidence: `internal/template/templates/.claude/skills/moai-foundation-cc/reference/claude-code-iam-official.md:30-40` ("exactly four values", denies `auto`).
- Downgrade path on this tree: `internal/config/autonomy_tiers.go` (`EffectiveTierWithGates`, `AppendDowngradeAdvisory`) + `internal/core/project/autonomy_bundle.go:75` (log sink `.moai/logs/autonomy-downgrade.log`).
- Regression tests located: `internal/core/project/autonomy_bundle_test.go:89,115,138,158`; `internal/config/autonomy_tiers_toggle_test.go:57`; `internal/cli/init_autonomy_wiring_test.go:237`; `internal/cli/wizard/autonomy_test.go:49`.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
