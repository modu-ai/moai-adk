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
- iteration 3 (0.3.0, audited sha `b2c254d444bad455dbcca599ca726a97bd98a946`): FAIL 0.79 (threshold 0.80), no must-pass failed, blocking D34, D35, D36, D38, D39, D40. Report: `.moai/reports/t1502/plan-audit-iter3.md` (gitignored card evidence). Single-model verdict, no cross-model receipt. Disposition of every iteration-2 finding: `plan.md` §M; iteration-1 optional findings: `plan.md` §L.
- ceiling reached; hold 2026-10-04, resumed 2026-10-05 on the operator's approval relayed by the leader: run enters as PASS-with-debt. Delta 4 (0.4.0, commit `30fd5d1b9`) wrote the iteration-3 findings down as run-entry repairs (`plan.md` §N); no confirming audit runs. The audit verdict of record stays FAIL 0.79 — the debt is accepted, not cleared.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

Mode Selection (orchestrator log, written before the first run-phase `Agent()` spawn).

- Input parameters: tier M; scope 15 certain files (plan.md §A); domains 3 (Go CLI, taxonomy package, SPEC artifacts); file language 100% Go plus markdown; concurrency benefit LOW (coding-heavy, shared package `internal/cli`); Agent Teams not requested.
- Mode evaluation: direct — not selected (15 files, new code); serial — **selected**; fanout — not selected (coding-heavy); sweep — not selected (not mechanical-uniform); agent-team — not selected (explicit-request only).
- Decision: serial — one `manager-develop` per milestone, M0 first, then M1 to M4; the lane verifies each milestone's landing on disk before the next spawn.
- Justification: coding-heavy work in one package; per `orchestration-mode-selection.md` §B.2 coding-heavy + multi-file defaults to serial. A milestone's output is the next milestone's input (the shared core feeds the doctor and the fold).

### Kickoff record (plan to run gate, operator form)

- Gate form: operator form, not the autonomous transition. The autonomous predicate cannot hold: plan-audit verdict of record is FAIL 0.79 (`auto-semantics.md` §9.1 keeps FAIL a hard block) and `blocking_count` is not 0.
- Approval: the operator's approval of 2026-10-05, **relayed by the leader** in a cross-session message (`leader-disposition.md` Disposition 4); the lane did not observe it first-hand. Recorded as a Gap, not as a first-hand confirmation.
- Scope of approval: run, then sync, then a codex card review (at most 2 rounds, the leader may run it), then report merge-ready. Not granted: push, merge window, real memory store access.
- Preferences collected: Q1 default OFF behind an env gate and Q7 split to the follow-up card (operator-confirmed 2026-10-04); Q2, Q4, Q5, Q8, Q9, Q10 DEFAULT-APPLIED; Q3 and Q6 carry empty verdicts (D41) — run-time defaults, recorded here as accepted debt.
- Progression mode: semi-autonomous; no `/moai goal` armed (the lane is its own judge, completion is read from evidence).

### Run-entry debt register (iteration-3 findings accepted as debt)

| Finding | Class | Where it is closed | State at run entry |
|---|---|---|---|
| D34 | DoD clause red at arrival | `acceptance.md` §7, `plan.md` §D and M5 line (delta 4, `30fd5d1b9`) | closed in SPEC text; the verifying command runs at the end of run |
| D35 | existing `internal/cli` close-path tests not HOME-isolated | `plan.md` M0 — run-mandatory FIRST code step, seam `userHomeDirFn` in `internal/cli/memory.go:123`, containment cell AC-MFB-008 (xi) | open — code, M0 |
| D36 | OD-11 row missing | `decision-index.md` row Q10 (delta 4) | closed in SPEC text |
| D38 | `linkage_test.go` assertions :99/:377/:392 | `plan.md` §A recount, B8 (delta 4) | closed in SPEC text |
| D39 | reordered multi-line fold passes every cell | `acceptance.md` AC-MFB-004 variant, AC-MFB-003 fixed expectations (delta 4) | closed in SPEC text; the variant cell must be RED-then-GREEN in the run |
| D40 | "checker and doctor cannot disagree" false | `spec.md` §1.5, AC-MFB-003 (delta 4) | closed in SPEC text |
| D37 | stamps predating rewritten text | `decision-index.md` re-stamped (delta 4) | closed |
| D41 | Q3/Q6 EVIDENCE-NEEDED rows block an autonomous Kickoff | n/a — operator form used here | open debt, named |
| D42 | optional | not addressed | open debt, named |

Absorption note: local `develop` moved 22 commits past the plan base `2f492df19` (tip `5e26ee139` at 2026-10-05); the run stays on `30fd5d1b9` as dispatched and absorbs `develop` before the merge window, with the pinned plan counts (file census, assertion counts) re-derived on the absorbed tree.
