# SPEC-FACTORY-LEADER-ADOPT-CORE-001 — progress

status: draft (plan-phase artifacts authored; status lives in spec.md frontmatter)

Artifacts: spec.md (10 REQ), plan.md (M1-M4), acceptance.md (10 AC), decision-index.md (3 rows).
Tree: `2178966bb90b28576e8a4c3bce340947b83c7cda`. Lane-owned card record: `.moai/reports/t1863/progress.md` (not written by this agent).

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-10-18
requirements: 10
acceptance_criteria: 10
tier: M

- Artifact set complete for Tier M: spec.md, plan.md, acceptance.md (+ this progress.md skeleton and decision-index.md; `interview.decision_gate: on`).
- SPEC ID pre-write self-check: `ID="SPEC-FACTORY-LEADER-ADOPT-CORE-001"` against `^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$` → `PASS` (executed Bash, verbatim).
- Evidence ledger E1-E6 (acceptance.md) measured on this tree in this run: E1 0/3, E2 0/9 (RED by the swept-count rule), E3 10/10 GREEN, E4 9/9 GREEN, E5 anchor absent in all ten guidance files, four banned README substrings present at line 70 of each locale.
- Counts verified: 10 `^- **REQ-FLAC` definitions in spec.md; 10 `### AC-FLAC` headings in acceptance.md — matching the counts above (the predecessor's D17 class).
- `moai spec lint --strict SPEC-FACTORY-LEADER-ADOPT-CORE-001` → `✓ No findings — all SPEC documents are valid` (judged by the installed build `ee981a932`, a verified ancestor of HEAD; a tree-built lint binary was not produced in the plan phase).
- Decision gate: decision-index.md carries 3 FOUNDER rows (Q1-Q3), all product-level, operator verdicts empty — they block the autonomous Kickoff until settled.
- Plan-audit iteration 2 (verdict `.moai/reports/t1863/plan-audit-iter2.md`, FAIL 0.94, D1-D6 verified RESOLVED): D7-D8 repaired in spec v0.3.0 — D7 AC-FLAC-006 gains rows 13 (`--resume` over the two-orphan fixture: the resume adoption adopts neither of two candidates, warns and mints, both orphans survive for the gate to see) and 14 (`--factory-run` adopting A while A2 survives); the resume-condition and `--factory-run`-condition mutant catches re-attributed to the two-orphan rows, since a single orphan is consumed by the adopting path before the gate and masks the mutant. D8 AC-FLAC-003's live-owner seed crossed to `claude` so AC-FLAC-008's each leg adopts a run seeded with the other launcher's backend (backend equality non-vacuous on both legs); per-launcher backend-omission mutant named. Counts unchanged: 10 REQ, 10 AC.
- Plan-audit iteration 1 (verdict `.moai/reports/t1863/plan-audit-iter1.md`, FAIL 0.94, codex gate fail on content): D1-D6 repaired in spec v0.2.0 — D1 the adoption write now plans the `runs.lead_backend` write in `AdoptRun`'s own transaction (plan M2 item 3, AC-003/008 seeded cross-backend so the clause cannot pass vacuously); D2 second qualifying orphan in AC-001/AC-005 with the first-only mutant named; D3 boot-proof-basis run H in AC-001 with the not-BasisPeer mutant; D4 counting-reader instrument for AC-002's no-stdin guarantee; D5 the `-cr` row 12 and corrected mutant attribution; D6 `classifyRuns` citation :172-277. Counts unchanged: 10 REQ, 10 AC.
- Gaps: the board rulings behind DP-1/DP-3 cannot be digest-pinned (`moai decision read` exposes no digest; same limitation the t1849 iter-3 audit measured). No behavioral RED observed (tests absent — E1/E2). The docs verify recipe and Hugo build are run-phase gates for M4.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
