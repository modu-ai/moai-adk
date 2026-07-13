---
id: SPEC-HNS-PREFIX-RENAME-001
updated: 2026-07-13
document: progress
plan_status: audit-ready
---

# Progress — SPEC-HNS-PREFIX-RENAME-001

## §E.1 Plan-phase Audit-Ready Signal

- Plan-phase artifacts authored 2026-07-13 by manager-spec: spec.md (26 REQs, REQ-HPR-001..026), plan.md (M1–M4 milestones, zero NEEDS CLARIFICATION markers as of iteration 2), acceptance.md (16 ACs + 3 GWT scenarios + 6 edge cases).
- SPEC ID self-check: `decomposition: SPEC ✓ | HNS ✓ | PREFIX ✓ | RENAME ✓ | 001 ✓ → PASS` (executed regex evidence: `PASS`, re-run at iteration 2).
- Scope baseline measured 2026-07-13 (see plan.md §A.1); run-phase must re-verify anchors.
- **Iteration 2 revision (2026-07-13; plan-auditor iter-1 FAIL 0.86 → defects D1–D5 addressed; artifacts at v0.1.1)**:
  - **D1 (BLOCKING)**: artifact prefix fixed to **lowercase `hns-`** per final user decision (lowercase-kebab matches the Claude Code skill/agent naming convention; uppercase runtime-acceptance risk eliminated). NEEDS CLARIFICATION marker, pre-M3 probe gate, and fallback branch all removed — zero markers remain. SPEC ID and REQ-HPR IDs unchanged.
  - **D2**: plan.md §B.2 collision analysis re-grounded with live evidence — legacy `REQ-HNS-*`/`AC-HNS-*` comment tokens found in 5 of 8 production Go files (16 occurrences: update.go 10, doctor_harness.go 2, prefix_conflict.go 2, doctor_skills.go 1, frozen_guard.go 1); lowercase `hns-` = **0 case-sensitive matches** in the production/artifact tree (live grep 2026-07-13); [HARD] all run-phase sweeps case-sensitive (no `grep -i`).
  - **D3**: REQ-HPR-012 corrected — doctor Runner resolution is manifest `runner_workflow`-driven (prefix-agnostic path join, ground-truth read of `internal/cli/harness/doctor.go`); only `runnerSpecialistRE` needs the dual pattern `(harness|hns)-[a-z0-9-]+-specialist`.
  - **D4**: AC-HPR-012 non-target baseline-delta file set pinned to 10 named files (4 handle-harness-observe hook templates, settings.json.tmpl, 4 config section files, catalog_loader.go).
  - **D5**: spec.md §E group-B row gains AC-HPR-007 as REQ-HPR-009's binding AC.
  - Consistency note: lowercase `hns-` also satisfies the doctor NAME-prefix resolution (doctor_harness.go `skills:` reference resolution, plan-time ~L276) without a third pattern (plan.md §B.1). M2-before-M3 ordering invariant unchanged. plan_status remains audit-ready.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
