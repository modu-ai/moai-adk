# SPEC-FACTORY-RETIRE-GUARD-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
audit_ready: true
- plan_complete_at: 2026-10-17T13:46:00Z
- Authoring: spec.md, plan.md, acceptance.md, decision-index.md written by manager-spec (Tier M, card t1852, Class C; part C only of the card line t1825 -> t1848). Tree `d2d9ac08fb1899440da911121ddf13eecc806a3f`, branch `WT-factory-retire-guard`. Not yet audited.
- Counts: 7 requirements (Tier M ceiling 16), 8 acceptance criteria (ceiling 16), 2 milestones (M1, M2).
- Tier choice: M. Tier S would fit the production LOC but not the file count (5 production files, 2 new test files, 8 sync-phase documents) and carries a plan-audit ceiling of 1 (`harness.plan_audit_tier_ceilings` S: 1, M: 2), which leaves no repair round.
- SPEC ID self-check, executed as Bash: `ID="SPEC-FACTORY-RETIRE-GUARD-001"; [[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL` printed `PASS`; `ls .moai/specs | grep -i -E "RETIRE-GUARD|RETIRE-TRANSFER|CARD-TRANSFER|RUN-SELECT|LEADER-ADOPT"` printed nothing (the ID is unique at this tree).
- Lint, two measurements: `moai spec lint --strict .moai/specs/SPEC-FACTORY-RETIRE-GUARD-001/spec.md` under the installed build (`moai version`: `v3.2.0-rc.36-t1822c`, commit `ee981a932`; `git merge-base --is-ancestor ee981a932 HEAD` exit 0, so a strict ancestor of the tree's HEAD) printed `✓ No findings — all SPEC documents are valid`. A binary built from this tree (`go build -o <scratchpad>/moai-tree ./cmd/moai` at HEAD `d2d9ac08f`, invoked by path; its version line reads `v3.1.3 none built unknown` because no ldflags were passed) printed the same line. The tree-built result is the one that is a measurement by a build of this tree (verification-claim-integrity §2.2); the lint ran on spec.md only, not on the other four files.
- Id consistency: `grep -ohE "REQ-RRG-[0-9]+"` over spec.md, plan.md and acceptance.md lists REQ-RRG-001..007 with no gap and no id outside that range; the same grep for `AC-RRG-[0-9]+` over plan.md and acceptance.md lists AC-RRG-001..008; `grep -n -E "RSC|RUNSEL"` over the five files printed nothing (no stale id prefix of the sibling SPECs); the sibling SPEC names appear only as prose citations in spec.md §F.
- RED-now evidence: `acceptance.md` ledger E1, E2 (empty sweeps, 5 + 4 named tests, exit 0), P1-P5 (behavioral probes observing the old behavior directly; probe sources deleted, never committed), E3 and E4 (existing guards GREEN). No behavioral RED of the named tests was observed (they do not exist yet).
- Decision rows: Q1-Q5 `DEFAULT-APPLIED` (all implementation-level); none open. Q3 carries a classification note: if the operator reads the added informational block as a change of a shipped command's default user-visible behavior, it becomes product-level.
- Inherited closed decisions (spec.md §F) and the code facts re-read at this tree are listed in the plan-phase return report; none was re-argued.
- Dispatch-fixed departure from the sibling SPEC: operator-facing prose of the `--retire` surface is Korean (plan §E); the sibling kept it English.

## Plan audit history

- Iteration 1: FAIL 0.88 (`.moai/reports/t1852/plan-audit.md`, audited commit `b26050d9b`; 3 blocking: D1 false "the one exit that still works" claim and an unscoped revival statement, D2 cards-before-owner ordering with a false "non-destructive exits" sentence, D3 abandon caveats missing from REQ-RRG-005, the stranded block and AC-RRG-006). Tier M ceiling 2, so this is the one repair round and the next audit is the last.
- Repair (version stays `0.1.0`; REQ 001-007 and AC 001-008 id sets unchanged), inside the audit's fix_scope: D1 spec REQ-RRG-005 and §C.7 (exclusivity removed; `--factory-run` statement scoped to that flag; revival routes `RecordRun` and `ResumeRun` recorded as a factual note and a residual risk, offered as no exit, no new requirement), plan §E stranded exit/note, AC-RRG-006; D2 plan §C.2 (order inside `retireRun` is status -> stamp -> cards -> flip -> event; the false ordering sentence replaced), AC-RRG-003 row (b) and mutant probes, Edge-cases third bullet; D3 REQ-RRG-005, plan §E (new stranded abandon-note row), AC-RRG-006 (abandon note asserted, mutant probe).
- Out-of-anchor changes, each the smallest edit that keeps a sentence true: `decision-index.md` Q3 Default (the phrase "the one exit that works for a non-active run" was the same false exclusivity claim). No other out-of-anchor edit. O1, O2 (AC-002, REQ-007) not taken; O3 (English error text duplicating the card list) not taken because the fix is the `Error()` sentence in plan §C.1, outside the anchors.
- Sentences re-checked for consistency with the new order and left unchanged: REQ-RRG-001 and REQ-RRG-002 (the owner gate still precedes the card read; "no row changed" stays true because the stamp check refuses before the flip), AC-RRG-002 (the hook still fires right after the in-transaction `SELECT`, after the stamp check), plan §C.1 (hook placement), plan §C.3 (the printed classification now always belongs to a still-unchanged owner).
- Repair-time lint: `moai spec lint --strict` on spec.md under the installed build and the tree-built binary, see the repair commit.
- Iteration 2: PASS 0.94 (.moai/reports/t1852/plan-audit-iter2.md, audited commit 9d171631a, plan_artifact_hash 0e68a9981361c7a407aa050f25c1614940ffd9cd07b668c04e2b17d1fd2e5110, codex receipt rcpt-dfdb489c540acf07a44604ff).

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §G Override and Refusal Record

- 2026-10-17T14:10:57Z SPEC-FACTORY-RETIRE-GUARD-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 2 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1852/.moai/reports/t1852/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1852/.moai/reports/t1852/plan-audit.md
- 2026-10-17T14:11:36Z SPEC-FACTORY-RETIRE-GUARD-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 2 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1852/.moai/reports/t1852/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1852/.moai/reports/t1852/plan-audit.md
- 2026-10-17T14:12:52Z SPEC-FACTORY-RETIRE-GUARD-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 2 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1852/.moai/reports/t1852/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1852/.moai/reports/t1852/plan-audit.md
