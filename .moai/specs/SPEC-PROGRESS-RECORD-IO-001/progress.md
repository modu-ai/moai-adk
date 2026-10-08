# SPEC-PROGRESS-RECORD-IO-001 — progress record

status: draft (plan-phase artifacts authored 2026-10-08, lane-5, card t1598, base a2a184ad3)

## §E.1 Plan-phase Audit-Ready Signal

- Artifacts authored: `spec.md`, `plan.md`, `acceptance.md`, `progress.md`, `decision-index.md`
  (Tier M set + decision index; `interview.decision_gate: on`).
- Tier: M (seeder rewrite + probe + family promotion; affected files < 8; REQ 9/16, AC 8/16
  within ceilings).
- Frontmatter: 12 canonical fields present; SPEC ID regex check PASS (verbatim `PASS` output
  cited in the authoring session); status `draft`.
- RED material: `.moai/state/verify/t1598-prework/held-audit_ceiling_axes_test.go` — the
  acceptance target, red-by-construction under `t.Setenv("PATH", "")` against the current
  exec-based darwin seeder; the in-package RED observation is scheduled at M2 entry (no test
  files written at plan phase).
- Scope correction recorded: F15/F16 measured ALREADY LANDED on a2a184ad3 (verification-only);
  F14 is the open P1 work item.
- Open decisions: `decision-index.md` Q1-Q3 unresolved (evidence/operator), Q4 default-applied.
- **Plan-audit round 1: FAIL 0.81** — verdict `.moai/reports/t1598/plan-audit-1.md` (audited
  SHA c404a0af4, artifact hash d262d9ad…dce6; blocking D1-D3, P2 D4-D7, P3 D8-D10). Revision
  round applied the full defect list without restructuring (MP-1~MP-9 clean results preserved):
  D1 AC-PRI-009 real-seeder fd-anchoring guard via the `seedFileMetadataFn` seam (mid-seed name
  swap, victim untouched; mutant-killer for path-based re-open regressions); D2 route-(ii)
  branch-conditional disposition in acceptance §D + conditional DoD in §F; D3 decisive surface
  re-pinned to `release-pr-multi-os.yml` (release/*→main PR or workflow_dispatch) with
  `-json` SKIP-vs-PASS recording rule, card-PR premise dropped (measured: ci.yml ubuntu-only,
  gate excludes internal/runtime); D4 close-hygiene probe `TestAppendProgressRecordSeedCloseHygiene`
  wired into M3 + AC-PRI-005; D5/D8 structured probe headings pinned in M1 step 6 and gated on
  the §E.2 content carrier; D6 AC-PRI-008 stale `victim-overwrite` text absence; D7 no-SKIP
  decisive runs for AC-PRI-003/005; D9 build-tag clause on the probe bullet; D10 t1560 citation
  location qualifier. REQ coverage now complete: REQ-PRI-004 → AC-PRI-009.
- Ready for plan-audit: re-audit pending (scoped to the defect delta per the Retry Loop
  Contract).

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
