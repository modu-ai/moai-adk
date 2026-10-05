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

### M0 (prior session, card t1502) — commit `74fdbaaa9`

- The (xi) containment cell written first, verbatim RED captured, then the one-line seam edit (`memory.go:123` `userHomeDir()` → `userHomeDirFn()`). Attributed to the prior session's branch history (`WT-memory-fold-budget`), not re-observed in this session.

### M1 shared core — commit `9edc8361f` (+`8ed5536e7` gofmt)

- RED-first per TDD contract (stub implementation, 11+13 intended-assertion failures, swept counts recorded); GREEN on committed HEAD `9edc8361f`: taxonomy `ok 0.411s` (AC-MFB-013: correct fold PASS, 8 lossy mutants FAIL — five (a)-(c), three (d); threshold boundary at the accessor), config `ok`. windows build exit 0; lint 0 issues; taxonomy coverage 90.4%.
- Lane verification: `go test -count=1 ./internal/hook/memo/taxonomy/ ./internal/config/` → both `ok`; gofmt fix applied by the lane (`8ed5536e7`).

### M2 doctor — commit `bc00fb8b3`

- G-DOCTOR 5/5 PASS (Measures/TopicCapUnchanged/BudgetBoundaries/BytesProxyWarns/LinkClasses); gate command green both packages. windows build exit 0; lint 0 issues; taxonomy coverage 90.1%.
- 5 gate-review findings fixed in-round (secondary-index qualification by full path; overflow-replaces-budget-warning shared line count; audit without topic files; fixture newline arithmetic; text-render details for the new codes), each RED-first.

### M3 fold core — commits `fecf5a7eb` + `725e109d9`

- G-FOLD 8/8 PASS; D39 reordered-variant RED (`[d2]` at the multi-line cell) then GREEN. windows build exit 0; lint 0 issues; taxonomy 89.5%, memory_fold.go 89.6% avg/19 funcs.
- 3 rounds of gate-review findings folded RED-first: retry-path archive re-verification made unconditional (whole-byte expected-archive comparison on every path, not gated on Appended); original permission bits preserved on replace (0600 survives rename); the content re-check moved to the LAST step before each rename (temp-prep window race closed). (iii-b) overlay cell added as the regression seal.

### M4 card-close wiring — commit `1d0983a68`

- AC-MFB-008 cell matrix 12/12 PASS (lane re-run `ok 15.318s`): disabled differential + gate accepted-values table, enabled fold, fail-open absent archive, ordering after queue write, seeded panic, blocked-read FIFO (bound 200ms, fold wait <400ms, store zero-write), gate-off never-opens (3 paths <1s, empty recorder), production bound ∈ [2s,3s], bound constant ==2s ≤5s, M0 (xi) containment unchanged, gate-P1-4 abandoned-step regression.
- Gate-review findings folded: the three close-path wirings themselves (the gate's first-round core finding — no operational caller existed); FIFO fixtures split build-tagged (unix/windows) reusing mkfifoForTest (windows `go test -c` exit 0); reader/writer deadlock restructured; abandoned-step cancellation propagated to the write path (mutant probe: FIFO-fed data recovery was a weak assertion — rewritten as a 500ms pause-seam between plan and apply before it bit); timing assertion re-bound to the fold wait window (close-path wall clock excluded), `-count=3` stable.
- Coverage (lane-measured, selector `TestMemoryFold|TestMemoryDoctor|TestMemoryFoldOnDone`, package internal/cli): memory_fold.go per-function — foldOnDoneGateOpen 100%, foldClosedCardMemory 100%, foldOnDoneReport 100%, foldIndexGuard 83.3%, foldOnDoneStep 83.8%, classifier/renderer helpers 80–100%. Package-level 7.2% (diluted — internal/cli is a 200+-file package; the selector exercises the memory surface only).
- Wiring attribution: one call each in todo.go / todo_autodone.go / todo_auto.go after successful queue mutation, outside the lock (plan §F M4).

### Cross-milestone process note

- M3's final gate-fix commit (`725e109d9`) landed while the M4 delegate was reading in the same tree; the M4 delegate detected the concurrent writer (lsof, PID 93991 = this lane's own session) and stopped with a structured blocker, zero writes. The lane re-verified M3's landing (G-FOLD green on `725e109d9`) before resuming M4 with an all-clear. Lesson recorded in lane memory.

### M0 — test-containment repair (seam `userHomeDirFn`) — prior session (branch history), not re-measured here

- `internal/cli/memory.go`: the home lookup in `memoryCandidateStores` goes through `userHomeDirFn` (the TestMain-sandboxed seam), not `userHomeDir` directly (commit `3577e290b` absorb). Containment cell `TestMemoryFoldOnDone_ExistingClosePathsContained` (AC-MFB-008 (xi)) lives in `memory_fold_wiring_test.go`; M4 extends the cell to the three close paths.

### M1 — shared core (budget + reachability) — prior session (branch history), not re-measured here

- `internal/config/defaults.go` (`DefaultMemoryIndexByteCap` 25000, `DefaultMemoryIndexWarnPercent` 80, `DefaultMemoryFoldOnDone` false, `DefaultMemoryFoldOnDoneBound` 2s), `internal/config/envkeys.go` (`EnvMemoryFoldOnDone`), `internal/hook/memo/taxonomy/linkage.go` (`SecondaryIndexLinkThreshold()` accessor + class-aware dangling + `MEMORY_REPO_RELATIVE_LINK`), `budget.go` (MeasureIndex / AuditIndexBudget), `reach.go` (classification, snapshots, I/R/T, A(S), `CheckFoldInvariants` (a)-(d); commits `9edc8361f`, `8ed5536e7`). M3 additions to `reach.go` (this session): `ResolvedLinkCount`, `LineTargetSetKey`, `IsArchiveIndexName` — re-exposures of the checker's own internals, one statement each.
- E8 RED-before-GREEN for the M1 checker self-test was the M1 session's record; its verbatim output is not re-quoted here (not measured in this session). Covered post-M3 in this session: `go test -count=1 ./internal/hook/memo/taxonomy` → `ok github.com/modu-ai/moai-adk/internal/hook/memo/taxonomy 0.301s`, exit 0.

### M2 — doctor (budget + link classes) — prior session (branch history), not re-measured here

- `internal/cli/memory.go` (commit `bc00fb8b3`): `index_bytes` / `index_chars` / `index_loaded_chars` / `index_link_targets`, `byte_cap` / `warn_percent` / `line_cap`, flags `--byte-cap` / `--line-cap` / `--warn-percent`. The plan-phase RED (E2/E3) is the branch history's; not re-observed here.
- This session's regression evidence: `go test -run '^TestMemoryDoctor_Measures$|^TestMemoryDoctor_TopicCapUnchanged$|^TestMemoryDoctor_BudgetBoundaries$|^TestMemoryDoctor_BytesProxyWarns$|^TestMemoryDoctor_LinkClasses$' -count=1 -v ./internal/cli` → exit 0, 5/5 `--- PASS`, `ok … 2.377s`.

### M3 — fold core (this session, branch WT-memory-doctor-budget)

- TDD sequence: test file written first; RED captured against a compiling behavior-absent skeleton — `go test -run '<G-FOLD pattern>' -count=1 ./internal/cli` → exit 1, 8 `--- FAIL` lines, every failure at its intended assertion (first cell: `memory fold --card t9001 --dir … exited with error: memory fold: not implemented`); then implemented → GREEN.
- GREEN (final, this tree): `go test -run '^TestMemoryFold_DryRunWritesNothing$|^TestMemoryFold_Classification$|^TestMemoryFold_ReachabilityPreserved$|^TestMemoryFold_VerbatimFiling$|^TestMemoryFold_ArchiveSelection$|^TestMemoryFold_Idempotent$|^TestMemoryFold_EdgeInputs$|^TestMemoryFold_ApplyOrderAndAbort$' -count=1 -v ./internal/cli` → exit 0, 8/8 `--- PASS`, `ok github.com/modu-ai/moai-adk/internal/cli 8.365s`. Swept count: `go test -list '<same pattern>' ./internal/cli` lists exactly the 8 pattern branches.
- E8 D39 RED-then-GREEN (multi-line reordered variant): with a temporary reversed-append mutant in `buildFoldPlan`, `go test -run '^TestMemoryFold_VerbatimFiling$' -count=1 ./internal/cli` → exit 1, `memory_fold_test.go:712: fold violates invariants [d2] of the reachability model` (the D39 multi-line cell; single-line cells are reversal no-ops); mutant removed → the same command exit 0, `--- PASS: TestMemoryFold_VerbatimFiling (0.68s)`.
- Gate findings folded in before the commit (coordinator messages, both RED-first):
  - P1 retry path skipped archive re-verification when `Appended` is empty — regression cell added to `TestMemoryFold_Idempotent` (retry with a concurrent author stripping the moved line: RED `… exited 0, want non-zero`); fix: the archive re-read + target-set verification runs on every apply path.
  - P1 residual (gate overlay): the same retry must abort on ANY archive drift, not only loss of the moved line — a concurrent author removing the archive's OTHER links drops index qualification while the moved line survives, and the deletion made the card unreachable. Overlay cell added (RED: `… exited 0, want non-zero`); fix: unconditional `checkFoldUnchanged` whole-byte comparison of the archive against the plan-time content before any deletion (REQ-MFB-004's letter, now on every path). Post-fix: `TestMemoryFold_Idempotent` PASS.
  - P2 the atomic replace forced 0644 — regression cell added to `TestMemoryFold_VerbatimFiling` (0600 MEMORY.md + archive; RED `MEMORY.md permission = 644 after the fold, want the original 0600` ×2); fix: `atomicWriteFoldFile` stats the original and applies its permission bits to the temp file (0644 fallback only for a genuinely new file). Post-fix: PASS.
  - P1 second residual (gate overlay, rename window): the content re-check ran BEFORE the temp file was created, so a concurrent update landing during the temp-file preparation was overwritten by the rename with success returned. Overlay cell (iii-b) added to `TestMemoryFold_Idempotent` (a `mutateDuringWrite` seam fires inside `atomicWriteFoldFile` after temp prep; RED probe observed `err=<nil>` with the concurrent line gone — the gate's `replacement error=<nil>; concurrent update preserved=false`); fix: the re-check is the LAST step before each rename, after the temp file is fully prepared. Post-fix: `TestMemoryFold_Idempotent`/`ApplyOrderAndAbort`/`VerbatimFiling` PASS, and the final battery re-ran 8/8 `--- PASS`, exit 0, `ok … 4.950s`.
- E1 G-FOLD evidence: AC-MFB-001 → `DryRunWritesNothing`, AC-MFB-002 → `Classification`, AC-MFB-003 → `ReachabilityPreserved`, AC-MFB-004 → `VerbatimFiling` + `ArchiveSelection`, AC-MFB-005 → `Idempotent`, AC-MFB-006 → `EdgeInputs`, AC-MFB-007 → `ApplyOrderAndAbort`; all 8 PASS with the swept count above. AC-MFB-013 (checker, M1) re-confirmed via the full taxonomy package run (`ok … 0.301s`, exit 0). G-DOCTOR regression: 5/5 `--- PASS`, exit 0.
- E2 `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0 (post-gate-fix tree).
- E3 coverage: `go test -cover ./internal/hook/memo/taxonomy/` → `coverage: 89.5% of statements`. `internal/cli` measured with the scoped fold+doctor selector and `-coverprofile`: `memory_fold.go` 19 functions, average 89.6% (`buildFoldPlan` 98.1%, `applyFold` 86.4%, `classifyFoldLine` 100%; `resolveFoldStore` 33.3% — its auto-resolution branch is exercised by the M4 wiring cells, `atomicWriteFoldFile` 65.2% — error branches).
- E5 lint: `golangci-lint run --timeout=5m ./internal/cli/... ./internal/hook/memo/taxonomy/...` → exit 0, `0 issues.` (baseline on this tree: no findings; one earlier 2m-budget attempt timed out on an unrelated package scan and printed `0 issues.` before the timeout — read as a measurement artifact, retried at 5m).
- E4 subagent boundary: `grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/memory*.go` → empty.
- Baseline attribution: all M3 measurements on branch `WT-memory-doctor-budget`, tree HEAD at the M3 commit; `internal/cli` runs serialized (B9), scoped selectors only, no full-package run — the full suite waits for CI.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

sync_commit_sha: b7c0d2474
sync_note: single sync commit carrying the `implemented → completed` transition (spec.md frontmatter), the CHANGELOG [Unreleased] entry, and this §E.4 record; `sync_commit_sha` backfilled in the following commit per the D3 placeholder exemption.
sync_scope: auto — changed surfaces are the M0-M4 implementation files (memory.go, memory_fold.go, memory_fold_wiring_test.go, memory_budget_test.go, memory_fold_test.go, defaults.go, envkeys.go, taxonomy budget/reach/linkage), CHANGELOG.md, spec.md frontmatter transition, progress.md evidence; docs-site untouched (no user-facing doc change this SPEC); codemaps untouched (no exported-surface change beyond the new verb, already covered by the M4-era tree).

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
