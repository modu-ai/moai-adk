# progress.md — SPEC-SYNC-GATE-SKIP-SUBSHELL-001

## §E.1 Plan-phase Audit-Ready Signal

- SPEC authored 2026-10-02 at plan phase for card t1395; tier M; status `draft`.
- All six card premises verified against tree `89e3164aff20345618c5db053107b066adc98316` (spec.md §B, evidence file:line cited per premise).
- Observational RED executed NOW and RE-EXECUTED from the committed harness (`red-now-t1395.sh` in this SPEC directory): 3/3 defect cells CONFIRMED, 3/3 positive controls PASS, harness exit 0, both runs; `git rev-parse --short HEAD` = `89e3164af` before and after; gate sha256 prefix `19180598f67db114` unchanged (spec.md §B P4, acceptance.md §D.1).
- Static sweep discriminator proven on pre-fix commit: 4 hits working copy / 4 hits `git show 89e3164af:<template>` (acceptance.md §D.2).
- ID uniqueness: no prior SPEC carries the SYNC-GATE-SKIP-SUBSHELL domain (catalog grep, 1009 SPEC dirs).
- Decision gate `on` (`.moai/config/sections/interview.yaml:6`) → `decision-index.md` authored (1 row, FOUNDER).
- Plan-phase artifacts: spec.md, plan.md, acceptance.md, progress.md, decision-index.md.

## §E.2 Run-phase Evidence

<pending run-phase>

## §E.3 Run-phase Audit-Ready Signal

<pending run-phase>

## §E.4 Sync-phase Audit-Ready Signal

<pending sync-phase>
