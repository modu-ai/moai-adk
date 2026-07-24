---
id: SPEC-CLI-WIZARD-RESTRUCTURE-001
title: "Progress — moai init wizard restructure (방안 A)"
version: "0.1.2"
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
- **Artifacts:** spec.md (19 GEARS REQs across 5 groups §B.1-§B.5 + 6 Out-of-Scope
  H3 sub-sections) · plan.md (§A context, §A.1 reversibility-ordered D1-D3, §A.5
  28-row CHANGE (C1-C28) + PRESERVE map, §B risks, §C-§H) · acceptance.md (16
  Given-When-Then ACs incl. AC-WIZ-012a + severity + full REQ→AC traceability).
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
- **Plan-audit fold (review-1): PASS 0.80; D1/D2 must-fix + D3-D7 cleanup applied; docs-site scoped to sync-phase (Tier M preserved).** D1 rewrote AC-WIZ-015's vacuous flag grep to the real cobra `.Bool("standard"/"advanced")` idiom; D2 scoped the docs-site 4-locale `--standard`/`--advanced` removal to sync-phase (REQ-WIZ-019 + new §C Out-of-Scope H3) keeping run-phase code at Tier M; D3 added the `init_update_notice.go:68` `runWizardFn` seam (§A.5 C28); D4 corrected C26 `Phase2Questions` mis-location (it lives in `advanced_gate.go`, deleted by C23); D5 fixed the B-brief-correction phantom `WithModelPolicy`/`initializer.go` citation → `context.go:106`; D6 added the reconfigure-membership leak note + AC-WIZ-012a; D7 corrected this §E.1 CHANGE-row count (22 → 28). All 4 artifacts bumped 0.1.1 → 0.1.2.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
