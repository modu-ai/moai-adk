---
id: SPEC-CLI-WIZARD-RESTRUCTURE-001
title: "Progress — moai init wizard restructure (방안 A)"
version: "0.1.0"
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
- **Open items for kickoff:** `[NEEDS CLARIFICATION: advanced_gate retirement
  scope]` (plan.md §B) — full retirement (option A) vs keep `--advanced`
  power-user path (option B). Isolated to milestone M5; M1-M4 land independently.
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
