# Progress: SPEC-TODO-UNIFY-001

## §E.1 Plan-phase Audit-Ready Signal

plan_authored_at: 2026-10-18T06:37:06Z
plan_revised_at: 2026-10-18T07:35:00Z
plan_status: audit-ready
plan_revision: 2
audit_ready: true
revised_from: plan-audit iteration 1 FAIL, overall_score 0.62, must_pass_failed 1 (MP-7), blocking_count 10, audited_sha 63829fe20a0f13fa93e974fe9331bb455ccb1009, codex required leg fail, receipt rcpt-90e345b1d5488c65cc03140e (report `.moai/reports/t1641/plan-audit.md`)
tier: L (spec.md, plan.md, acceptance.md, design.md, research.md)
tree_basis: 026f48658f (branch WT-todo-unify); revision measured at HEAD 23e40eb7c4, which adds only plan artifacts and the audit report

Requirement and criterion counts: 25 requirements (REQ-TU-001 to REQ-TU-025), 25 criteria (AC-TU-001 to AC-TU-025), contiguous, counted with `grep -c -E '^\*\*(REQ|AC)-TU-'`; Tier L ceilings are 25 each.

Decisions: Q1 to Q9 are decided in `decision-index.md` (lane decision ladder, one decision record each); Q5 is marked `lane-judgment, leader-confirmed 10-18 (conditional)`. No clarification marker remains in any plan artifact (`git grep -n -c -F` for the marker over the SPEC directory printed nothing).

Defects of the audit and where each went: D1 decision-index.md and plan.md §B; D2 plan.md §B and decision-index.md Q5 to Q7; D3 spec.md REQ-TU-017 to REQ-TU-020, REQ-TU-025 and acceptance.md AC-TU-014 to AC-TU-020; D4 spec.md REQ-TU-019, design.md §D.6, acceptance.md AC-TU-018; D5 spec.md REQ-TU-020 and REQ-TU-022, design.md §D.5 and §D.7, acceptance.md AC-TU-023 and AC-TU-025; D6 plan.md §A (graph) and M3b files; D7 spec.md REQ-TU-003, design.md §D.2, research.md §R.11, acceptance.md AC-TU-002; D8 acceptance.md AC-TU-013; D9 acceptance.md AC-TU-004, AC-TU-007, AC-TU-022; D10 acceptance.md AC-TU-005, AC-TU-009, AC-TU-010 and research.md §R.5b; D11 acceptance.md AC-TU-011; D12 plan.md §C item 7, §E M8, §G and research.md §R.10; D13 acceptance.md AC-TU-001, AC-TU-010 (proxy flip conditions, E-15); D14 spec.md REQ-TU-023 (event-detected form) and research.md §R.8; D15 plan.md §E (M2 and M3 split seams, M8 unit) and the dropped docs-site page merge.

Lint evidence (this run): `moai spec lint SPEC-TODO-UNIFY-001` and `moai spec lint --strict SPEC-TODO-UNIFY-001`, each with a binary built from this tree at HEAD `23e40eb7c4` by plain `go build -o <scratchpad>/moai-t1641 ./cmd/moai` (no Makefile LDFLAGS, so the version stamp is not the build's identity, the tree is), printed `0 error(s), 0 warning(s)` and one INFO finding `OwnershipTransitionUnmeasured` (the plan commit `63829fe20a` carries no `Authored-By-Agent` trailer); both exited 0.

Verified this round (command run, output seen): the cobra execution order by reading `command.go:900-1010` of v1.10.2; the receipt gate's checks by reading `leader_approval.go` and `leader_approval_gate.go`; the verdict reader and merge verifier by reading `card_evidence_readers.go`; the five rule files and the generated embed line that name `workflows/gtd.md`; the two copies of the workflow document (`diff -q`); `reflect --rebuild-projection` writes files (grep of `gtd_private.go`); the guidance-token and swept-file counts (§R.5b); the board record for Q9 (read from `moai decision read`); E-15 to E-18 empty or matching as recorded; the existing test names the criteria cite (`git grep -n '^func Test'`).

Not verified (Gaps): no `go test` was run (card constraint), so every class B criterion stays unadopted until its RED is recorded; the cobra ordering was read, not executed with a throwaway program; the exit status of the empty `git grep` runs (E-15, E-17, E-18) was not displayed by the tool; whether each of the 67 conversion paths has a template source was not individually checked (pre-flight item 7 does it); whether `sh` is available to the script test on every CI runner; the 516-commit lag of the installed `moai`, the AGENTS.md ceiling value 24,576, and the docs-site build were not re-measured; the T29 edge, the revocation table, and the factory-before-queue order are design decisions with no code behind them yet; the leader confirmation of Q5 is the coordinator's message, not a board record; the worktree guard refused two commands this round (a `python3 -c` calling git after `cd`, and a `git grep --include=*.go`), see `research.md` §R.1.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
