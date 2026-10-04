# progress.md — SPEC-MEMORY-FOLD-BUDGET-001

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-10-04
- revision: 0.3.0 (plan delta for plan-audit iteration 2; the previous audited revision was 0.2.0 at `0dfae6d4a`, and 0.1.0 at `73c4ab646` before it)
- tier: M
- artifacts: spec.md + plan.md + acceptance.md (Tier M set) + progress.md + decision-index.md (conditional, `interview.decision_gate: on`) + fixtures/store-A/ (synthetic, read-only acceptance input, unchanged in this revision; SHA-256 of its two index files re-measured equal to the recorded values)
- req_count: 12
- ac_count: 13
- baseline_tree: 2f492df19
- certain_file_count: 15 (Tier M band 5-15: on the upper edge, inside it; three named contingencies take it to 16-18, plan.md §A)

### Plan-phase lint (tool provenance: judging build vs measured tree)

- judging build: `./bin/moai`, built from this tree and invoked by path, commit stamped `2f492df19` through `-ldflags` (`./bin/moai version` prints `v3.1.3   2f492df19   built 2026-10-04`); measured tree HEAD for the 0.3.0 run: `0dfae6d4a` (parent chain `73c4ab646` ← `2f492df19`; no Go source differs). The Go VCS stamp of a plain `go build` in this worktree named another checkout (`c8f245c2c9a5`, not an ancestor of HEAD), so it was not trusted.
- 0.3.0 run: `./bin/moai spec lint SPEC-MEMORY-FOLD-BUDGET-001` → stdout `✓ No findings — all SPEC documents are valid`, exit 0.
- 0.3.0 run: `./bin/moai spec lint --strict SPEC-MEMORY-FOLD-BUDGET-001` → same stdout, exit 0.
- 0.1.0 run (kept for the record): same two commands, same output, exit 0; positive control (a deliberately bad scratch spec, `phase: plan`, uppercase conditional keywords, no AC, no exclusions) → exit 1, `1 error(s), 4 warning(s)` — the lint can fail.
- `go test ./internal/spec/` was not run: nothing under `fixtures/` changed in this revision.

### Revision 0.3.0 measurements

- File recount by extraction from plan.md §F (`sed -n '/^## §F/,/^## §G/p' plan.md`, backticked `*.go` names, de-duplicated): 16 strings, of which `linkage.go` and `internal/hook/memo/taxonomy/linkage.go` are one file → 15 distinct files (10 non-test source, 5 named test files). No existing test file is edited by plan; three contingent edits are named in plan.md §A.
- Cross-layer sweep of the moved SessionStart items over all five files (pattern: SessionStart, session_start, memory-budget, AdvisoryJoin, join bound, home_isolation, main_test, G-HOOK, OD-9, HomeJoin, TestMain, additionalContext, M5, E5a/E5b, E7a/E7b): the remaining hits are only statements that the item moved (spec HISTORY and §1.2 note and §4 Out of Scope; plan §A.1 note, B-list note, M5 "moved", §G, §I, §L, §M; acceptance §2 "Removed cells" and §4.1; decision-index Q7 and its header). Requirement/criterion identifiers above 012 and 013 do not occur outside HISTORY and split-history text.
- REQ definitions in spec.md: REQ-MFB-001 … REQ-MFB-012 (12, gap-free); AC headings in acceptance.md: AC-MFB-001 … AC-MFB-013 (13, gap-free); the requirement-to-criterion table is acceptance.md §3 (12 requirements covered, no orphan criterion).
- Exact-plan mutants (acceptance.md §5): `oracle3.py` and the doctor were run on scratch copies of `fixtures/store-A` outside the tree; the correct fold passes (a)-(d), the three further mutants pass (a)-(c) and fail (d); the doctor reports the same four findings for the correct fold, the link-free-deletion mutant and the stray-edit mutant.
- RED-now ledger: E1, E2, E3, E6a-todo/-autodone/-auto, E6b-todo/-autodone/-auto re-executed against tree `0dfae6d4a` with the `2f492df19` binary, all unchanged; E8a/E8b added and executed; E5a/E5b/E7a/E7b removed with the SessionStart half.

### Plan-audit history

- iteration 1 (0.1.0, audited sha `73c4ab646880c68743966b9d6e8c1b98e3112632`): FAIL 0.69, 12 blocking (D1-D12), 8 optional (D13-D20). Report: `.moai/reports/t1502/plan-audit.md` (gitignored card evidence). Leader disposition: `.moai/reports/t1502/leader-disposition.md`.
- iteration 2 (0.2.0, audited sha `0dfae6d4a172cd8609c3c310988d74be12774b5b`): FAIL 0.71, D8 carried plus D21-D26 blocking, D27-D33 optional. Report: `.moai/reports/t1502/plan-audit-iter2.md` (gitignored card evidence). Second leader disposition: second split, one more delta authorized, Q1 and Q7 confirmed by the operator.
- iteration 3 input (0.3.0): this revision. Disposition of every iteration-2 finding: `plan.md` §M; iteration-1 optional findings: `plan.md` §L. A confirming audit has not run.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
