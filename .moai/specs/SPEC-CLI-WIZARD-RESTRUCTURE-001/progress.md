---
id: SPEC-CLI-WIZARD-RESTRUCTURE-001
title: "Progress — moai init wizard restructure (방안 A)"
version: "0.1.1"
status: draft
created: 2026-07-25
updated: 2026-07-25
author: manager-spec
tier: M
---

# Progress — SPEC-CLI-WIZARD-RESTRUCTURE-001

## §E.1 Plan-phase Audit-Ready Signal

- **Tier:** M (standard) — wizard refactor across ~8-10 Go files
  (`questions.go`, `wizard.go`, `translations.go`, `init.go`,
  `model_policy.go`, `context.go`, `profile.go` + test files); presentation/UX
  + config-default changes; no data-model rewrite, no new subsystem.
- **Artifacts:** spec.md (17 GEARS REQs across 4 groups + 5 Out-of-Scope H3
  sub-sections) · plan.md (§A context, §A.1 reversibility-ordered D1-D3, §A.5
  22-row CHANGE + PRESERVE map, §B risks incl. 1 NEEDS-CLARIFICATION, §C-§H) ·
  acceptance.md (14 Given-When-Then ACs + severity + full REQ→AC traceability).
- **SPEC-ID self-check:** `SPEC-CLI-WIZARD-RESTRUCTURE-001` → regex
  `^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$` → PASS
  (decomposition: SPEC ✓ | CLI ✓ | WIZARD ✓ | RESTRUCTURE ✓ | 001 ✓ → PASS).
- **Resolved clarification (advanced_gate, option A full retirement) — v0.1.1
  bump (2026-07-25):** the kickoff clarification is settled — the user chose
  방안 A **option A, FULL retirement** of the advanced-settings plumbing.
  plan.md §B rewritten to the RESOLVED decision entry; retirement targets
  enumerated as concrete plan.md M5 rows (C23 `advanced_gate.go` delete,
  C24/C25 `--standard`/`--advanced` flags + `RunWithDefaultsModes` mode params +
  orphaned `WizardResult.StandardMode`/`AdvancedMode` fields, C26 `Phase2Questions`
  + gated-`Phase1Questions` wrapper — Page-3 questions preserved), plus C27
  repo-wide caller reconciliation (`.github/` + docs + `_test.go`). spec.md gains
  REQ-WIZ-018/019 (retirement + caller reconciliation) and a reframed §C exclusion;
  acceptance.md gains AC-WIZ-015 (retirement grep, MUST). Rejected option B
  (hidden `--advanced` power-user path). All 4 artifacts bumped 0.1.0 → 0.1.1.
  No unresolved clarification markers remain in plan.md. M5 now concrete; M1-M4
  still land independently.
- **Verification-integrity corrections recorded:** the brief's model_policy
  4-site claim was verified against source — 2 sites real (questions.go default,
  model_policy.go const), 2 phantom (wizard.go L58 seed, init.go resolveModelPolicy
  — neither is a default seed), 2 additional real consumers found (context.go,
  profile.go). MUST-FIX reachability site identified: `init.go` `if
  result.StandardMode` application gate (C20) must be removed or Page-3 answers
  are discarded.
- **Plan-auditor target:** PASS ≥ 0.80 (Tier M threshold).

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
