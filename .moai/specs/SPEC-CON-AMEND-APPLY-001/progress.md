# progress.md — SPEC-CON-AMEND-APPLY-001

Card t659 · branch `WT-amend-apply` · worktree `.claude/worktrees/t659`

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: draft — revision 0.1.5 fixed plan-audit iteration 1 (FAIL 0.78) except D4; revision 0.1.6 applied the operator decision on D4; plan-audit iteration 2 FAIL 0.84; revision 0.1.7 fixed its defects N1–N6; plan-audit iteration 3 PASS 0.89 on 0.1.7; revision 0.1.8 applies that audit's optional findings X1–X4 and records the operator's approval of the backup-write / temporary-write Gap, both by operator decision and **without re-audit** — the 0.1.8 changes were not audited, per that decision; awaiting `moai spec lint` on 0.1.8, then run-phase entry
- tier: L, revision 0.1.8 — raised from M in revision 0.1.4 by operator decision (verdict §13.3): 21 requirements and 25 acceptance criteria each exceed the Tier M ceiling of 16; 25 acceptance criteria equals the Tier L ceiling
- plan artifacts: six files — `spec.md`, `plan.md`, `acceptance.md`, `design.md`, `research.md` (revision 0.1.8; `design.md` and `research.md` added in 0.1.4; `research.md` unchanged since 0.1.4), and this file
- plan commits: `7b4d1ac89` (0.1.0, rulings §7); `01af243ae` (0.1.1, rulings §8, adds this file); `e693f0583` (0.1.2, ruling §9); `4e9273d0b` (0.1.3 partial: spec.md, plan.md, and this file; acceptance.md not yet edited); `fa966740d` (0.1.3 completion: acceptance.md, reconciliation of all four files); `51454e5e5` (0.1.4 tier raise, adds `design.md` and `research.md`); `193136a6a` (0.1.5 plan-audit iteration 1 fixes D1–D3, D5–D16); `564c370b5` (0.1.6 operator decision D4); `183bba916` (0.1.7 plan-audit iteration 2 fixes N1–N6); the 0.1.8 optional findings X1–X4 and Gap approval are the commit that updates this section, kept separate from every other commit as the operator decision requires
- rulings encoded: `.moai/reports/t659/verdict.md` §7 (Q1–Q5), §8 (G1–G5, scope addition (a)), §9 (G6, option (ii)), §11 (G7, option A), §12.5 (two plan-phase extensions accepted), §13.3 (Tier L); plan-audit iteration 1 defects D1–D3 and D5–D16 (`.moai/reports/t659/plan-audit-iter1.md`, verdict §15); operator decision D4 (amend path only); plan-audit iteration 2 defects N1–N6 (`.moai/reports/t659/plan-audit-iter2.md`); plan-audit iteration 3 optional findings X1–X4 (`.moai/reports/t659/plan-audit-iter3.md`) and the operator decision quoted under Implementation Kickoff Approval below
- counts: 21 requirements, 25 acceptance criteria, 29 mutants (unchanged in 0.1.8, 0.1.7, 0.1.6, 0.1.5, and 0.1.4)
- changed in 0.1.8 (plan-audit iteration 3 optional findings X1–X4 and the Gap approval, operator decision, not re-audited): acceptance.md — revision line; the common rules gain the path-formatting rule (`%s`/`%v`, not `%q`; X4); AC-CAA-023's parenthetical states the check runs first on the amend path and the loader's retained refusal would refuse the same path (X1, no wording assertion added); AC-CAA-025's preservation paragraph names the two blind spots of the `--first-parent --no-merges` range and makes the test run the primary witness (X3 option (a)). spec.md: frontmatter `version: "0.1.8"`, HISTORY, authoring-tree paragraph, REQ-CAA-020 rationale (a read means opening a file's contents; X2), §E.4 backup-write / temporary-write Gap recorded as approved by the operator as a blank (no seam added; the run report carries it under Residual-risk; REQ-CAA-010 not narrowed). plan.md: header, §C.1 (X3 blind-spot line). design.md: header, §H.6 (Gap approval). No ID added or renumbered
- changed in 0.1.7 (plan-audit iteration 2, N1–N6): acceptance.md — the common rules gain the numeric-assertion rule; AC-CAA-002, AC-CAA-003, AC-CAA-014, AC-CAA-018 (fixture and assertions), and AC-CAA-019 state their numbers under it; AC-CAA-024 and AC-CAA-025 use a resolved base `B`; the AC-CAA-024 CLI case admits both path forms; the AC-CAA-025 preservation assertion ranges from `BASELINE_SHA` with `git log --first-parent --no-merges`; the §D matrix rows of AC-CAA-024 and AC-CAA-025; §D.2 rows M-2, M-10, M-15, M-20 (variant (iv) added), M-21, and M-22; the closing mutant paragraph. plan.md: §C.1 (`BASELINE_SHA`; `_test.go` filter, 8 lines expected), M2 approach (check order) and exit (M-20 four variants), R-7. design.md: header, §B, §C.2, §C.6. spec.md: HISTORY, authoring-tree paragraph, §E.4 (check-order Gap). No ID added or renumbered
- changed in 0.1.6 (D4, operator decision): the containment check of REQ-CAA-021 runs on the amend path only — `Execute` and `runConstitutionAmend`; `LoadRegistry` keeps its present behaviour and its five non-amend call sites are unchanged; extending the check to read-only callers is a follow-up card candidate recorded by the lead. spec.md: REQ-CAA-020, REQ-CAA-021, §D terms, §A item 6, §C, §E.2, §F (new exclusion), §G (D4 resolved). plan.md: §A, §B, §C.1, §D verification scope (adds the `internal/spec` selector), M2 approach and exit, R-8. design.md: §A, §B, §C.2–§C.5. acceptance.md: AC-CAA-022 and AC-CAA-023 wording; AC-CAA-024 preservation case `loader_unchanged`; AC-CAA-025 preservation run of `TestLinter_AC08_DanglingRuleReference`; M-19 names AC-CAA-023's divergent rows; M-20 gains variant (iii) (check moved into `LoadRegistry`) and variant (i) drops AC-CAA-023; M-24 kills AC-CAA-025 `dry_run` instead of `load`. No ID added or renumbered
- changed in 0.1.5: REQ-CAA-004 (error names the registry path), REQ-CAA-008 and REQ-CAA-009 (`When` form), REQ-CAA-013 (`--dry-run` flag only), REQ-CAA-018 (temporary files removed on a failed restore), REQ-CAA-021 (lock file excepted), seven GEARS labels; AC-CAA-003, AC-CAA-005 (b), AC-CAA-007, AC-CAA-012 (rename-then-fail injector; subtests `first_rename_applied`, `third_rename_applied`), AC-CAA-014, AC-CAA-024 (rows `symlinked_registry`, `symlinked_file`; discriminating CLI case), AC-CAA-025 (count derived from the copy, target entry by rule, `load` entry point); M-5b and M-23 split into variants, the registry-path mutant gains a CLI-only variant; RED statements labelled as predictions; spec.md §E.4 gains an unapproved Gap (D7) and §G records D4; no ID added or renumbered
- added in 0.1.4: `design.md`, `research.md`; frontmatter `tier: L`, `version: "0.1.4"`; no requirement, acceptance criterion, mutant, ruling, or scope changed, no ID renumbered
- added in 0.1.3: REQ-CAA-021, AC-CAA-024, AC-CAA-025, M-21, M-22, M-23, M-24
- changed in 0.1.3: REQ-CAA-020 (its closing sentence points at REQ-CAA-021), AC-CAA-023 (asserts the refused path, not the loader's wording), M-20 (names the registry-path site of the one check; also kills AC-CAA-024 `relative_env_escape` and its CLI case)
- reconciled in the 0.1.3 completion: plan.md declared `M-21a` / `M-21b` while spec.md and this file declared one M-21; folded into M-21 with two variants, and M-22 given two variants (no-boundary prefix; no-Clean prefix), so M-15, M-18, M-21, and M-22 carry variants and the distinct mutant count stays 29; AC-CAA-024 gains a CLI dry-run case for the shared-check clause of REQ-CAA-021
- resolved question: D4 (plan-audit iteration 1) — operator decision in 0.1.6: amend path only (`spec.md` §G). Backup-write / temporary-write Gap (plan-audit iteration 1 D7) — operator decision in 0.1.8: approved as a blank, carried to the run report's Residual-risk (`spec.md` §E.4). No open question remains.
- spec lint: 0.1.8 is recorded by the lane in `.moai/reports/t659/lint-0.1.8.txt` (tree `54b298476`: `0 error(s), 0 warning(s)`, `LINT_EXIT=0`; judging build in `lint-binary-0.1.8.txt`: `v3.2.0-rc.7`, commit `ed71054d3`, a strict ancestor of the tree; `git diff --stat ed71054d3 54b298476 -- internal/spec internal/cli/spec*.go` prints nothing, so no lint rule landed between the judging build and the tree). 0.1.7 is recorded by the lane in `.moai/reports/t659/lint-0.1.7.txt` (tree `183bba916`: `0 error(s), 0 warning(s)`, `LINT_EXIT=0`; judging build in `lint-binary-0.1.7.txt`) and is not evidence for 0.1.8. 0.1.6 is recorded by the lane in `.moai/reports/t659/lint-0.1.6.txt` (tree `564c370b5`: `0 error(s), 0 warning(s)`, `LINT_EXIT=0`; judging build in `lint-binary-0.1.6.txt`) and is not evidence for 0.1.7. No 0.1.5 result is recorded under `.moai/reports/t659/` (a working-tree listing of that directory taken with HEAD at `193136a6a` shows no `lint-0.1.5.txt`). 0.1.4 is recorded by the lane in `.moai/reports/t659/lint-0.1.4.txt` and verdict §14.2 (tree `51454e5e5`). Earlier results are recorded by the lane: 0.1.2 in `.moai/reports/t659/lint-0.1.2.txt` and verdict §10.2 (tree `e693f0583`), 0.1.3 in `.moai/reports/t659/lint-0.1.3.txt` and verdict §12.3 (tree `fa966740d`); none is evidence for 0.1.5, 0.1.6, 0.1.7, or 0.1.8.
- plan-audit verdict: iteration 1 FAIL 0.78 (`.moai/reports/t659/plan-audit-iter1.md`, audit HEAD `76144d40a`, verdict §15); iteration 2 FAIL 0.84 (`.moai/reports/t659/plan-audit-iter2.md`, audit HEAD `0085766cd`; D1–D16 all resolved; new defects N1–N6, of which N1–N3 blocking); iteration 3 PASS 0.89 (`.moai/reports/t659/plan-audit-iter3.md`, audit HEAD `e39168140d499307fd3daf4ed2a19a0f6ae1b039` — the parent of `0c4d2e4e7` — judging SPEC revision 0.1.7, whose files last changed in `183bba916`; N1–N6 resolved, no D1–D16 regression, all must-pass items passed; four new defects X1–X4, all optional, none blocking) — the last under the Tier L cap. Revision 0.1.8 applies X1–X4 **without re-audit**, by operator decision: no audit has judged the 0.1.8 changes, and the iteration 3 verdict covers 0.1.7 only. The audit interrupted on `f2ba39c94` produced no artifact and is not a verdict basis (verdict §13.2).
- Implementation Kickoff Approval: approved — operator decision relayed by the lead session, 2026-09-11, quoted verbatim:

  > t659 운영자 결정: (1) Implementation Kickoff 승인. (2) 백업·임시 쓰기 실패 복원 Gap(spec.md:L171) = 공백 승인 — run 보고서 Residual-risk로 남기세요(시접 추가 없음). (3) optional X1–X4 = run 전 SPEC에 반영(괄호문 정정·read 정의·first-parent 사각지대·Windows %q), 재감사 없이 진입 — 반영 커밋을 따로 두고 progress에 운영자 결정 인용.

  `spec.md:L171` in the quotation is the §E.4 Gap bullet as numbered at `0c4d2e4e7`; the 0.1.8 HISTORY row moves it to L172. Item (3)'s X3 "first-parent 사각지대" is applied as option (a) — the blind spot documented, the range unchanged.

## §E.2 Run-phase Evidence

- run-phase entry: 2026-09-11, cycle_type=tdd, go1.26.8 darwin/arm64
- BASELINE_SHA = `fe8cc9875aea7bad57fe05a88a1b33f8d196fe53` — the HEAD before the first run-phase commit (plan.md §C.1); it is the merge that absorbed local develop `85868148c`, and `git diff --stat 7a6af9ea0 fe8cc9875 -- internal/constitution internal/cli/constitution*.go internal/spec/lint*.go` printed nothing (lead measurement relayed in the run dispatch)
- pre-flight re-measure: every plan.md §C.1 expectation matched — `.moai/reports/t659/run/preflight.txt` (2 stub lines; 0 / 7 yaml tags; 1 resolver; 0 / 0 EvalSymlinks; 1 absolute-only check; 8 `LoadRegistry(` lines; real registry sha256 `f7707b1d…be21`, real log sha256 `f5735051…69f0`)

- scope of verification: `go test ./internal/constitution/ -count=1`, anchored `internal/constitution` selectors, and `go test ./internal/spec/ -run '^TestLinter_AC08_DanglingRuleReference$' -count=1 -v`. **No `internal/cli` test was run** — the compile slot was not granted to this lane; every cell that needs one is marked `pending compile slot`, never PASS. `go vet ./internal/cli/` (allowed) type-checks the CLI code. All evidence files live under `.moai/reports/t659/run/` (tracked).

### E.2.1 Run-phase commits (BASELINE_SHA..HEAD, first-parent)

| Commit | Kind | Milestone |
|---|---|---|
| `32d946797` | chore — run entry, `status: draft → in-progress`, pre-flight | — |
| `81c87c11b` | test — M1 baseline RED (7 top-level tests, all FAIL) | M1 |
| `53ea919c9` | feat — M1 evolution-log writer/reader | M1 |
| `3f9921a74` | test — M1 mutants (9 runs) | M1 |
| `5b16cdd11` | test — M2–M6 Execute-level baseline RED (12 top-level tests) | M2–M6 |
| `0b28804c7` | feat — M2 resolver + amend-path containment check (constitution) | M2 |
| `299bae37d` | test — CLI tests before the CLI change (not run: compile slot) | M2/M6 |
| `38928086f` | feat — M2 CLI uses the shared resolver and `LoadAmendRegistry` | M2 |
| `e09700b23` | feat — M3 source/registry transforms | M3 |
| `78a3dd926` | test — M4 seam tests before the seams (package test build fails: the non-discriminating RED) | M4 |
| `1995d3427` | feat — M4 seams, backups, commit sequence | M4 |
| `ff25d6745` | feat — M5 wiring into `Execute`, stub tests retired | M5 |
| `e79b17147` | feat — M6 dry-run validation, fixture update | M6 |
| `db6b70166` | test — M2–M6 mutants (29 runs) | M2–M6 |
| (this commit) | chore — M7 isolation witness, M-14, vet/lint/coverage, this section | M7 |
| (guards commit, after `ae58f0383`) | test — the two out-of-SPEC guards pinned by one test and one mutant each (§E.2.6) | — |
| `0893611ad` | test — G-B test asserts 0 gate calls, RED on the unmoved code (§E.2.6 G-B move) | — |
| `5801ebda0` | fix — G-B moved into `Execute` before Layer 1, removed from `prepareApply` (§E.2.6 G-B move) | — |
| (record commit, after `5801ebda0`) | docs — G-B move evidence, mutants re-run, this record | — |

### E.2.2 RED-before-GREEN evidence (baseline-first ACs)

| Evidence file | Tree it ran on | Content |
|---|---|---|
| `m1-baseline-red.txt` | `32d946797` (production = BASELINE) | 7 top-level RUN, all FAIL: writer emits `ruleid:`/`zonebefore: 1`; both-forms `RuleID = "B"`; `entries = 0` on markdown noise, the verbatim EVO-HRN-002 section, and the real log; no fail-closed error; `TestLoadEvolutionLogs` `RuleID mismatch: got ""` |
| `m2-m6-baseline-red.txt` | `3f9921a74` (`pipeline.go`, `loader.go`, `internal/cli` = BASELINE) | 12 top-level RUN; RED cells fail on the stub error `updateSourceFile: not yet implemented`, on `dryRun=true: want an error, got nil`, on `gate doubles called 4 times, want 0`, on `want a registry load error, got nil`; the cells predicted GREEN are GREEN (AC-CAA-014 `valid`, AC-CAA-024 `in_root_control/*/dry_run` and `loader_unchanged` ×4, AC-CAA-025 `load`, `dry_run`) |
| `m4-seams-red-compile.txt` | `e09700b23` | `p.rename undefined … p.restore undefined … undefined: restoreFile … [build failed]` — the non-discriminating RED acceptance.md records for AC-CAA-012/013/021 |
| (commit `e79b17147` message) | the M6 production change with the old fixture | `TestPipeline_Execute_DryRun_Success` / `…CanaryUnavailable_Continues`: `amendment validation error: rule file …/dummy.md: stat …: no such file or directory` — observed, then the fixture was updated (plan.md §C.2) |

### E.2.3 AC matrix (E1)

| AC | Status | Command | Observed |
|---|---|---|---|
| AC-CAA-001 | PASS | `go test ./internal/constitution/ -run '^TestApply_ExactlyOnce_Success$' -count=1 -v` | `m5-green.txt`: `--- PASS: TestApply_ExactlyOnce_Success` |
| AC-CAA-002 | PASS | `… -run '^TestApply_OccurrenceCount_Rejected$'` | `m5-green.txt`: PASS, subtests `zero`, `two` |
| AC-CAA-003 | PASS | `… -run '^TestApply_NoWhitespaceNormalization$'` | `m5-green.txt`: PASS |
| AC-CAA-004 | PASS | `… -run '^TestUpdateRegistryClause_SingleLineRoundTrip$'` | `m5-green.txt`: PASS |
| AC-CAA-005 | PASS | `… -run '^TestApply_RegistryReparse_Rejects$'` | `m5-green.txt`: PASS, `continuation`, `no_clause_line`, `newline_clause` |
| AC-CAA-006 | PASS | `… -run '^TestAppendEvolutionLog_SnakeCaseAndZoneNames$'` | `m1-green.txt`: PASS |
| AC-CAA-007 | PASS | `… -run '^TestLoadEvolutionLogs_LegacyKeys$'` | `m1-green.txt`: PASS |
| AC-CAA-008 | PASS | `… -run '^TestLoadEvolutionLogs$'` | `m1-green.txt`: PASS (3 subtests) |
| AC-CAA-009 | PASS | `… -run '^TestLoadEvolutionLogs_MarkdownNoise$'` | `m1-green.txt`: PASS |
| AC-CAA-010 | PASS | `… -run '^TestLoadEvolutionLogs_HumanFormat$'` | `m1-green.txt`: PASS, `verbatim_block`, `real_file_readonly`, `malformed_timestamp_fails_closed` |
| AC-CAA-011 | PASS | `… -run '^TestRateLimiter_SeesHumanFormatEntry$'` | `m1-green.txt`: PASS |
| AC-CAA-012 | PASS | `… -run '^TestApply_RenameFault_RestoresAll$'` | `m5-green.txt`: PASS, all six subtests |
| AC-CAA-013 | PASS | `… -run '^TestApply_RenameOrder$'` | `m5-green.txt`: PASS |
| AC-CAA-014 | PASS | `… -run '^TestPipeline_Execute_DryRun_Validates$'` | `m6-green.txt`: PASS, all six subtests |
| AC-CAA-015 | pending compile slot | `unset MOAI_CONSTITUTION_REGISTRY CLAUDE_PROJECT_DIR MOAI_CONSTITUTION_DRY_RUN && go test ./internal/cli/ -run '^TestConstitutionAmend_DryRun_SurfacesValidation$' -count=1 -v -timeout 600s` | not run (test committed in `299bae37d`; vet-clean) |
| AC-CAA-016 | PASS | `/usr/bin/grep -rn 'not yet implemented' internal/constitution/pipeline.go internal/constitution/pipeline_test.go` | no output, exit 1; control `grep -c 'func (p \*Pipeline) Execute'` → `1`; stub-name grep → no output; each replacement test name → `1` |
| AC-CAA-017 | PASS (constitution half) / pending compile slot (the AC-CAA-015 run) | `unset MOAI_CONSTITUTION_REGISTRY CLAUDE_PROJECT_DIR && go test ./internal/constitution/ -count=1`; `CLAUDE_PROJECT_DIR=<worktree root> go test ./internal/constitution/ -count=1` | both `ok`; real registry `f7707b1d…be21` and real log `f5735051…69f0` unchanged before/after; `git status --porcelain -- .claude/rules .moai/research` empty; `internal/constitution/.moai` and `internal/cli/.moai` absent |
| AC-CAA-018 | PASS | `… -run '^TestLoadEvolutionLogs_FailClosedErrorLocation$'` | `m1-green.txt`: PASS, `unparseable_timestamp`, `missing_timestamp` |
| AC-CAA-019 | PASS | `… -run '^TestApply_NewClausePresent_Rejected$'` | `m5-green.txt`: PASS |
| AC-CAA-020 | PASS | `… -run '^TestPipeline_Execute_StaleBefore_Rejected$'` | `m5-green.txt`: PASS, `dry_run`, `real` |
| AC-CAA-021 | PASS | `… -run '^TestApply_RestoreFault_KeepsBackups$'` | `m5-green.txt`: PASS |
| AC-CAA-022 | PASS (constitution) / pending compile slot (CLI) | `… -run '^TestExecute_UsesSharedRegistryResolver$'`; CLI `go test ./internal/cli/ -run '^TestResolveRegistryPath_MatchesExecute$' -count=1 -v -timeout 600s` | `m5-green.txt`: PASS; CLI not run |
| AC-CAA-023 | PASS | `… -run '^TestExecute_RegistryOutsideProjectDir_Refused$'` | divergent rows PASS from M2 (`m2-constitution.txt`); all three rows PASS (`m5-green.txt`) |
| AC-CAA-024 | PASS (constitution) / pending compile slot (CLI case) | `… -run '^TestExecute_ContainmentCheck_RefusesEscapes$'`; CLI `go test ./internal/cli/ -run '^TestConstitutionAmend_ContainmentCheck_RelativeEnvEscape$' -count=1 -v -timeout 600s` | `m5-green.txt`: PASS, 0 skipped — every escape row (real + dry_run), `in_root_control` (plain, symlinked_root), `loader_unchanged` (4 cases); CLI not run |
| AC-CAA-025 | PASS | `… -run '^TestExecute_RealRegistryShape_Admitted$'`; `go test ./internal/spec/ -run '^TestLinter_AC08_DanglingRuleReference$' -count=1 -v`; `git log --first-parent --no-merges --format=%H fe8cc9875aea7bad57fe05a88a1b33f8d196fe53..HEAD -- internal/spec/lint.go internal/spec/lint_test.go` | `load`, `dry_run`, `real` PASS; `--- PASS: TestLinter_AC08_DanglingRuleReference` (`m7-spec-preservation.txt`); log range prints nothing (control on `internal/constitution/pipeline.go` over the same range prints 4 commits) |

### E.2.4 Mutant table (E2)

Runner `.moai/reports/t659/run/mutate.py` (apply alone → run the killing selector → restore byte for byte → KILLED only when the selector reached its tests and failed); specs `mutants-m1.json`, `mutants-m2m6.json`, `mutants-m7.json`; per-mutant output in `.moai/reports/t659/run/mutants/`.

| Mutant | Result | Named cells observed RED |
|---|---|---|
| M-1 | killed | AC-CAA-002 `two` |
| M-2 | killed | AC-CAA-003 (`2` present, `0` absent in the path-stripped error) |
| M-3a / M-3b | killed / killed | AC-CAA-004 |
| M-4 | killed | AC-CAA-005 `continuation` |
| M-5a | killed | AC-CAA-012 all five fault subtests |
| M-5b (i) / (ii) | killed / killed | (i) `first_rename_applied`, `third_rename_applied`, `third_rename_log_absent`; (ii) `third_rename_log_absent` |
| M-5c | killed | AC-CAA-013 |
| M-6, M-7 | killed | AC-CAA-007; AC-CAA-008 |
| M-8a / M-8b | killed / killed | AC-CAA-010 (a) + AC-CAA-011; AC-CAA-010 (c) + AC-CAA-018 `unparseable_timestamp` |
| M-9 | killed | AC-CAA-006 |
| M-10 | killed | AC-CAA-014 (b)–(e) dry-run |
| M-11a / M-11b | killed / killed | AC-CAA-012 `no_fault_clean`; AC-CAA-014 `valid` snapshot |
| M-12 | killed | AC-CAA-009 (extra-rule variant) |
| M-13 | pending compile slot | AC-CAA-015 `two_occurrences` |
| M-14 | killed | AC-CAA-017: package run FAILs with `CLAUDE_PROJECT_DIR` = repo root and PASSes scrubbed (`M-14-session-env.txt` vs `M-14-scrubbed-env.txt`) — the two runs disagree |
| M-15 (i) / (ii) / (iii) | killed ×3 | AC-CAA-018 both subtests each |
| M-16, M-17 | killed | AC-CAA-019; AC-CAA-020 |
| M-18 (i) / (ii) | killed / killed | AC-CAA-021 |
| M-19 | killed | AC-CAA-022; AC-CAA-023 `divergent_root_real`, `divergent_root_dry_run` |
| M-20 (i) | killed (constitution cells) / pending compile slot (CLI cell) | `relative_env_escape` ×4, `symlinked_registry` ×2; CLI case not run |
| M-20 (ii) | pending compile slot | AC-CAA-024 CLI case |
| M-20 (iii) | killed | `loader_unchanged` `relative_registry`, `absolute_file`, `symlinked_registry` (also `absolute_escape`, which the moved check refuses first); internal/spec `TestLinter_AC08_DanglingRuleReference` |
| M-20 (iv) | killed | `loader_unchanged/absolute_escape` |
| M-21 file / log | killed / killed | `absolute_file` ×4, `dotdot_file`, `sibling_prefix_file`, `symlinked_file`; `symlinked_log` |
| M-22 (i) / (ii) | killed / killed | `sibling_prefix_file`; `dotdot_file` |
| M-23 (i) / (ii) / (iii) | killed ×3 | `symlinked_registry`; `symlinked_file`; `symlinked_log` (each also fails `in_root_control/symlinked_root`) |
| M-24 | killed | `in_root_control/symlinked_root`; AC-CAA-025 `dry_run` |

Counted: 39 mutant runs killed (38 distinct mutants or variants — M-20 (iii) ran against two selectors), 0 survived; pending compile slot: M-13, M-20 (ii), and the CLI cell of M-20 (i).

### E.2.5 Quality (E3, E5) and isolation (E4)

- `go vet ./internal/constitution/ ./internal/cli/` → exit 0, no output (`m7-vet.txt`, 0 bytes)
- `golangci-lint run ./internal/constitution/...` → `0 issues.` (`m7-lint-constitution.txt`); the first run found 2 `unused` (the orphaned `fakeOversight`), removed in the M7 commit
- `go test -cover ./internal/constitution/ -count=1` → `coverage: 88.1% of statements` (`m7-cover.txt`)
- `GOOS=windows GOARCH=amd64 go vet ./internal/constitution/` and `GOOS=windows GOARCH=amd64 go build ./internal/cli/` → exit 0 (compile only; no Windows test run)
- REQ-CAA-015 witness: see AC-CAA-017 row; the mutant runs (including M-14 under the session environment) left both real sha256 values unchanged

### E.2.6 Out-of-SPEC guards G-A (empty After) and G-B (rule file is the registry or the log)

SPEC 밖 추가, 리드 인정 (2026-09-11, relayed by the lead session)

The run added two checks to `internal/constitution/pipeline.go` that no REQ or AC asks for. The lead accepted keeping both as trust-boundary input validation on condition that each is pinned by exactly one test and one mutant, recorded here, and named in the sync report for sync-auditor review. They carry no AC id and are not counted in the §E.2.3 matrix or the §E.3 AC counts.

| Guard | Where | What it does | Modes it covers |
|---|---|---|---|
| G-A | `Execute`, right after the stale-`Before` check (REQ-CAA-017), before Layer 1 | rejects a proposal whose `After` is empty: `proposal for rule <id>: After is empty` | dry-run and real (no gate is called in either) |
| G-B | `Execute`, right after G-A, before Layer 1 (moved from `prepareApply` on 2026-09-12, see "G-B move" below) | rejects a target entry whose `file:` names the registry or the evolution log: `rule file <path>: is also the registry or the evolution log`; `sameFile` compares `filepath.Abs` results of `ruleFilePath(projectDir, file)`, the resolved registry path, and the evolution-log path — no file I/O | dry-run and real (no gate is called in either) |

What each guard adds over the code without it (observed on the mutants, below): without G-A an empty `After` is still refused, but only after all four gate doubles ran, by the log-entry validation (`evolution log <path>: clause is empty`), with no rule id in the message. Without G-B both halves are admitted: a `file:` pointing at the registry, or at a log that carries the current clause once, passes every later check — `Execute` returns no error in dry-run and real mode (observed). That the real run then commits two changes to the same path is read from `prepareApply` / `commitChanges`, not observed: the test stops at the nil error.

Tests (in `internal/constitution/apply_test.go`):

- `TestExecute_EmptyAfter_Rejected` — subtests `dry_run`, `real`. Asserts the substring `After is empty`, the rule id, 0 gate calls, the fixture tree's path set + per-file sha256 unchanged (`snapshotTree` / `assertSameTree`), and the lock dir empty.
- `TestExecute_RuleFileIsRegistryOrLog_Rejected` — subtests `registry/{dry_run,real}` (entry `file:` = `.claude/rules/moai/core/zone-registry.md`), `evolution_log/{dry_run,real}` (entry `file:` = `.moai/research/evolution-log.md`, log prose carries the current clause once), `distinct_control/{dry_run,real}` (the standard `rules/target.md`; must NOT be rejected — `Execute` returns no error). The two rejecting cases assert the substring `is also the registry or the evolution log`, the rule-file path (`containsPathForm`), 0 gate calls (4 before the G-B move), the tree snapshot unchanged, and the lock dir empty.

The GREEN run, mutant table, and package state below record the guards commit, when G-B still sat in `prepareApply`; the G-B move that follows supersedes the G-B rows for the current tree.

GREEN (unmutated tree), `.moai/reports/t659/run/guards/green.txt`:

```
$ go test ./internal/constitution/ -run '^(TestExecute_EmptyAfter_Rejected|TestExecute_RuleFileIsRegistryOrLog_Rejected)$' -count=1 -v
exit=0   top-level "=== RUN" lines: 2
--- PASS: TestExecute_EmptyAfter_Rejected (0.01s)
    --- PASS: TestExecute_EmptyAfter_Rejected/dry_run (0.01s)
    --- PASS: TestExecute_EmptyAfter_Rejected/real (0.01s)
--- PASS: TestExecute_RuleFileIsRegistryOrLog_Rejected (0.03s)
    --- PASS: TestExecute_RuleFileIsRegistryOrLog_Rejected/registry/dry_run (0.00s)
    --- PASS: TestExecute_RuleFileIsRegistryOrLog_Rejected/registry/real (0.00s)
    --- PASS: TestExecute_RuleFileIsRegistryOrLog_Rejected/evolution_log/dry_run (0.00s)
    --- PASS: TestExecute_RuleFileIsRegistryOrLog_Rejected/evolution_log/real (0.00s)
    --- PASS: TestExecute_RuleFileIsRegistryOrLog_Rejected/distinct_control/dry_run (0.00s)
    --- PASS: TestExecute_RuleFileIsRegistryOrLog_Rejected/distinct_control/real (0.01s)
ok  	github.com/modu-ai/moai-adk/internal/constitution	0.489s
```

Mutants — runner `.moai/reports/t659/run/mutate.py`, specs `.moai/reports/t659/run/guards/mutants-guards.json`, per-mutant output `.moai/reports/t659/run/guards/M-*.txt`; `pipeline.go` sha256 `eb258e65…a4a2` before and after both mutant passes (restored byte-exact). The guards predate the tests, so each mutant run is the RED observation for its test.

| Mutant | Edit | Selector | Top-level RUN | Result | Failing subtests |
|---|---|---|---|---|---|
| M-GA | G-A deleted | `^TestExecute_EmptyAfter_Rejected$` | 1 | killed (exit 1) | `dry_run`, `real` — error `…evolution log <path>: clause is empty` lacks `After is empty` and the rule id; `gate doubles called 4 times, want 0` |
| M-GB | condition → `false` | `^TestExecute_RuleFileIsRegistryOrLog_Rejected$` | 1 | killed (exit 1) | `registry/*`, `evolution_log/*` — `want a rule-file-is-registry-or-log error, got nil` |
| M-GB-reg | registry half dropped (log half kept) | `…_Rejected$/^registry$` | 1 | killed (exit 1) | `registry/dry_run`, `registry/real` — `got nil` |
| M-GB-log | log half dropped (registry half kept) | `…_Rejected$/^evolution_log$` | 1 | killed (exit 1) | `evolution_log/dry_run`, `evolution_log/real` — `got nil` |
| M-GB-always | condition → `true` | `…_Rejected$/^distinct_control$` | 1 | killed (exit 1) | `distinct_control/dry_run`, `distinct_control/real` — the distinct rule file is rejected |

Counted: 5 mutant runs, 5 killed, 0 survived. M-GA and M-GB are the "one mutant each"; M-GB-reg and M-GB-log pin each half of the `||` with its own subtest; M-GB-always pins the control.

Package state after the change (this run, the guards-commit tree):

- `go test ./internal/constitution/ -count=1 -cover` → `ok  	github.com/modu-ai/moai-adk/internal/constitution	1.107s	coverage: 88.3% of statements` (`guards/cover.txt`; the M7 figure in §E.2.5 was 88.1% on the `ae58f0383` tree)
- `go vet ./internal/constitution/` → exit 0, no output (`guards/vet.txt`)
- `golangci-lint run ./internal/constitution/...` → `0 issues.` (`guards/lint.txt`)

#### G-B move — before the gates (lead decision, 2026-09-12)

> 사람이 Y로 승인한 뒤 거부되는 순서는 승인을 헛되게 만드는 UX 결함 (lead, 2026-09-12)

In `prepareApply` G-B ran after all five layers, Layer 5 HumanOversight included, so a user's approval was spent before the rejection. The lead ordered the check moved into the pre-gate validation of `Execute` and its test to assert zero gate calls.

Equivalence of the paths (read from the code at `90ea2b26f`, before the move): `Execute` computes `registryPath := ResolveRegistryPath(projectDir)`, `evolutionLogPath := filepath.Join(projectDir, ".moai", "research", "evolution-log.md")`, and `currentRule` before Layer 1, and passes exactly `currentRule`, `registryPath`, `evolutionLogPath` to `prepareApply`, which built the rule path with the pure `ruleFilePath(projectDir, rule.File)`; `readForChange` stores the path it is given unchanged (`apply_commit.go`, `c := &fileChange{role: role, path: path, …}`). The moved check compares the same three strings.

One ordering difference follows from the move (read from the code, not observed): the old check ran after the three `readForChange` calls, so a `file:` naming an evolution log that does not exist was refused by the rule-file read (`rule file <path>: … no such file …`) before G-B could fire; the moved check refuses it with the G-B message. The test's `evolution_log` case writes the log, so it does not cover the missing-log form.

Commits (the commit graph witnesses RED before the fix, verification-claim-integrity §2.3):

- `0893611ad` — `test(t659): G-B test asserts no gate call before the guard moves`: `TestExecute_RuleFileIsRegistryOrLog_Rejected` requires 0 gate calls in `registry/*` and `evolution_log/*`; commits the RED output with the test, production code untouched
- `5801ebda0` — `fix(t659): check the rule-file alias before the gates`: the check moved into `Execute` right after G-A, same message text with `%s`; the `prepareApply` copy deleted (no duplicate); `sameFile` unchanged; the `prepareApply` doc comment notes the check now runs earlier

RED on the unmoved code, `.moai/reports/t659/run/guards/gb-move-red.txt` (tree `90ea2b26f` + the test edit):

```
$ go test ./internal/constitution/ -run '^TestExecute_RuleFileIsRegistryOrLog_Rejected$' -count=1 -v
exit=1   "=== RUN" lines: 7 (1 top-level + 6 subtests)
    apply_test.go:1264: gate doubles called 4 times, want 0 (the check runs before Layer 1)   (×4)
    --- FAIL: TestExecute_RuleFileIsRegistryOrLog_Rejected/registry/dry_run (0.01s)
    --- FAIL: TestExecute_RuleFileIsRegistryOrLog_Rejected/registry/real (0.01s)
    --- FAIL: TestExecute_RuleFileIsRegistryOrLog_Rejected/evolution_log/dry_run (0.01s)
    --- FAIL: TestExecute_RuleFileIsRegistryOrLog_Rejected/evolution_log/real (0.01s)
    --- PASS: TestExecute_RuleFileIsRegistryOrLog_Rejected/distinct_control/dry_run (0.01s)
    --- PASS: TestExecute_RuleFileIsRegistryOrLog_Rejected/distinct_control/real (0.01s)
FAIL	github.com/modu-ai/moai-adk/internal/constitution	0.527s
```

The only failing assertion is the gate count: the error text, path, snapshot, and lock assertions already held before the move.

Mutants on the new location — runner `.moai/reports/t659/run/mutate.py`, specs `.moai/reports/t659/run/guards/mutants-guards-moved.json`, per-mutant output `.moai/reports/t659/run/guards/M-*-moved.txt`, summary `mutants-moved-summary.txt`. `pipeline.go` sha256 `05934830…1dd2` and `apply_test.go` sha256 `bfdddf3e…3424` before and after the pass (restored byte-exact; the runner also checks restoration).

| Mutant | Edit (at the moved check) | Selector | Top-level RUN / all RUN | Result | Failing subtests |
|---|---|---|---|---|---|
| M-GA | G-A deleted (re-run) | `^TestExecute_EmptyAfter_Rejected$` | 1 / 3 | killed (exit 1) | `dry_run`, `real` — error lacks `After is empty`; `gate doubles called 4 times, want 0` |
| M-GB | condition → `false` | `^TestExecute_RuleFileIsRegistryOrLog_Rejected$` | 1 / 7 | killed (exit 1) | `registry/*`, `evolution_log/*` — `want a rule-file-is-registry-or-log error, got nil` |
| M-GB-reg | registry half dropped (log half kept) | `…_Rejected$/^registry$` | 1 / 3 | killed (exit 1) | `registry/dry_run`, `registry/real` — `got nil` |
| M-GB-log | log half dropped (registry half kept) | `…_Rejected$/^evolution_log$` | 1 / 3 | killed (exit 1) | `evolution_log/dry_run`, `evolution_log/real` — `got nil` |
| M-GB-always | condition → `true` | `…_Rejected$/^distinct_control$` | 1 / 3 | killed (exit 1) | `distinct_control/dry_run`, `distinct_control/real` — the distinct rule file is rejected |

Counted: 5 mutant runs, 5 killed, 0 survived. The earlier `M-*.txt` files are kept as the pre-move record; the `-moved` files are the current one. The 0-gate-call assertion's own discriminating power is the RED above (4 calls on the unmoved code); the four G-B mutants die earlier, on the nil error.

Package state after the move (tree `5801ebda0`):

- `go test ./internal/constitution/ -count=1 -cover` → `ok  	github.com/modu-ai/moai-adk/internal/constitution	1.954s	coverage: 88.3% of statements` (`guards/cover-moved.txt`)
- `go vet ./internal/constitution/` → exit 0, no output (`guards/vet-moved.txt`)
- `golangci-lint run ./internal/constitution/...` → `0 issues.` (`guards/lint-moved.txt`)

The symlink limit of `sameFile` is carried in §E.3 `residual_risk`.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-11
run_commit_sha: ae58f038352d7cebef944a58c352cd196550b895   # the M7 commit carrying this block (backfilled in the sync commit, see §E.4 run_commit_sha_backfill)
baseline_sha: fe8cc9875aea7bad57fe05a88a1b33f8d196fe53
run_status: complete                   # compile slot measured 2026-09-12 (lane re-measure, .moai/reports/t659/run/slot/summary.md)
ac_pass_count: 25                      # every cell observed PASS; the four slot cells (AC-CAA-015, 017 CLI half, 022 CLI, 024 CLI) ran in the slot
ac_partial_pending_slot: []
ac_pending_slot: []
ac_fail_count: 0
mutants_killed_runs: 42                # 39 re-measured by the lane in the slot (remeasure/) + 3 CLI mutants (M-13, M-20-i CLI cell, M-20-ii); guard mutants counted separately
mutants_survived: 0
mutants_pending_slot: []
cli_baseline_observation: "299bae37d (tests before CLI change 38928086f), exported with git archive: AC-CAA-015 two_occurrences FAIL (got nil) and AC-CAA-024 CLI FAIL (clause mismatch) as predicted; AC-CAA-022 CLI PASS (preservation) — slot/s5-cli-baseline-red-299bae37d.txt"
cli_lint: "golangci-lint run ./internal/cli/... — 0 issues (slot/s3-lint-cli.txt)"
preserve_list_post_run_count: "LoadRegistry unchanged (loader.go not in BASELINE..HEAD); filepath.IsAbs(cleanPath) count 1"
l44_pre_commit_fetch: not-run           # lane does not fetch/push (lead batch push, CLAUDE.local.md §4.1)
l44_post_push_fetch: not-applicable      # nothing pushed
new_warnings_or_lints_introduced: 0     # golangci-lint 0 issues on ./internal/constitution/... and ./internal/cli/... (the latter in the slot)
cross_platform_build:
  darwin_arm64_tests: pass
  windows_amd64_vet_constitution: pass
  windows_amd64_build_cli: pass
total_run_phase_files: 15              # Go source and test files; evidence and SPEC records excluded
m1_to_mN_commit_strategy: per-milestone commits, baseline RED commits ahead of each production change
compile_slot_commands_run_2026_09_12:  # all run in the slot; results in .moai/reports/t659/run/slot/summary.md
  - "unset MOAI_CONSTITUTION_REGISTRY CLAUDE_PROJECT_DIR MOAI_CONSTITUTION_DRY_RUN && go test ./internal/cli/ -run '^TestConstitutionAmend_DryRun_SurfacesValidation$' -count=1 -v -timeout 600s"
  - "go test ./internal/cli/ -run '^TestResolveRegistryPath_MatchesExecute$' -count=1 -v -timeout 600s"
  - "go test ./internal/cli/ -run '^TestConstitutionAmend_ContainmentCheck_RelativeEnvEscape$' -count=1 -v -timeout 600s"
  - "go test ./internal/cli/ -run '^TestConstitution' -count=1 -timeout 600s   # existing constitution CLI tests re-run (plan.md M2 exit)"
  - "baseline observation: the three CLI tests at 299bae37d (before 38928086f) — AC-CAA-024 CLI case predicted `clause mismatch`, AC-CAA-015 two_occurrences predicted success line"
  - "mutants M-13, M-20 (ii), M-20 (i) CLI cell via mutate.py against internal/cli"
  - "golangci-lint run ./internal/cli/..."
out_of_spec_guards:                     # §E.2.6 — SPEC 밖 추가, 리드 인정; not counted in ac_* or mutants_* above
  - "G-A: Execute rejects an empty After before Layer 1 — TestExecute_EmptyAfter_Rejected, mutant M-GA killed"
  - "G-B: Execute rejects a rule file that is the registry or the evolution log before Layer 1 (moved from prepareApply 2026-09-12, lead decision; RED 0893611ad, fix 5801ebda0) — TestExecute_RuleFileIsRegistryOrLog_Rejected asserts 0 gate calls, mutants M-GB, M-GB-reg, M-GB-log, M-GB-always killed at the new location, M-GA re-run killed"
residual_risk:
  - "G-B symlink limit: sameFile compares filepath.Abs results and resolves no symbolic link, so a rule-file entry that reaches the registry or the evolution log through a symlinked alias is not caught by G-B; no test covers that case"
sync_report_obligation: "the sync report MUST name both guards (G-A, G-B) for sync-auditor review, as additions outside the SPEC accepted by the lead; both run before Layer 1"
sync_report_behavior_change: "the sync report's behavior-change item MUST carry one line: a rule file whose file: points at an evolution log that does not exist yet was rejected with a 'no such file' error before the G-B move and is now rejected with the G-B message, before any gate (lead, 2026-09-12: intended behavior, kept; inferred from code reading, no test covers it — a test is optional)"
```

## §E.4 Sync-phase Audit-Ready Signal

Sync phase run by manager-docs on 2026-09-12 inside the card worktree (branch `WT-amend-apply`, base HEAD `128a5ea52`). git-flow repository: no PR; the lane merges the branch into local develop through the lead's integration window. Nothing was pushed or merged here.

```yaml
sync_complete_at: 2026-09-12
sync_commit_sha: pending-backfill        # the sync commit carrying this block; backfilled by the follow-up commit
sync_status: complete
sync_agent: manager-docs
sync_base_head: 128a5ea52c24b92f3e8e1e0bd7f1bb392e2e0086   # last run-phase commit (compile-slot record)
changelog_entry_position: "CHANGELOG.md [Unreleased] → ### Fixed, first bullet"
b12_self_test_a: "grep -c 'SPEC-CON-AMEND-APPLY-001' CHANGELOG.md → 0 before emission, 1 line after"
b12_self_test_b: "grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' acceptance.md | sort -u | wc -l → 25 (AC-CAA-001 … AC-CAA-025); 0 occurrences marked [RETIRED] or [REF]; 25 live, 0 ambiguous; the entry states 25/25"
b12_self_test_c: "ls on the 14 paths the entry cites → exit 0, all present; both guard test functions found in internal/constitution/apply_test.go (grep -c → 2)"
frontmatter_status_transitions:
  spec_md: "status in-progress → implemented → completed, updated 2026-09-11 → 2026-09-12, in the single sync commit; no other frontmatter field and no body line changed"
  sibling_artifacts: "plan.md, acceptance.md, design.md, research.md carry no frontmatter (stateless on the status axis, spec-frontmatter-schema.md § Artifact Statelessness); progress.md records phases in its body"
run_commit_sha_backfill: "§E.3 run_commit_sha ← ae58f038352d7cebef944a58c352cd196550b895. The field's own comment names 'the M7 commit carrying this block', and git log -S'run_commit_sha' over progress.md lists only ae58f0383, the commit that introduced the block. The last run-phase commit, 128a5ea52 (compile-slot record), is recorded above as sync_base_head, not in §E.3. Written by manager-docs at the lead's instruction as a mechanical placeholder completion; no other §E.2 / §E.3 line changed"
docs_site: "no change in any locale — docs-site/content/{ko,en,ja,zh}/cli-reference/constitution.md say the amendment is applied only after the five-layer gate and that --dry-run simulates without modifying files; both statements are true of the implementation, and none of the four pages claims the apply step is unimplemented or describes dry-run validation or error behavior"
readme: "no change — no README makes a claim about moai constitution amend"
project_docs_finding: ".moai/project/structure.md:78 says internal/constitution has '13 non-test files'; measured 18 at this HEAD and 14 at BASELINE_SHA (already stale before this card). Not edited — outside the sync dispatch scope; reported for the lead"
mx_validation:
  added: "internal/constitution/registry_path.go ruleFilePath — @MX:ANCHOR + @MX:REASON (fan_in 3 distinct production callers: LoadAmendRegistry, Pipeline.Execute, prepareApply); comment-only change, gofmt clean, go vet clean, go test ./internal/constitution/ -count=1 ok (292 RUN lines, env scrubbed)"
  present_and_valid: "ResolveRegistryPath @MX:ANCHOR (2 callers, REQ-CAA-019 boundary); LoadAmendRegistry @MX:ANCHOR (2 callers, REQ-CAA-021 boundary); commitChanges @MX:WARN with @MX:REASON (multi-file write with rollback)"
  stale_anchor_reported: "LoadEvolutionLogs @MX:ANCHOR (added in this run) gives 'fan_in >= 3 (rateLimiter.Admit, applyAmendment, MarkRolledBack)'; applyAmendment no longer exists — the apply step calls parseEvolutionLog directly — so production fan_in is 2 (rate_limiter.go:42, evolution_log.go:81). The protocol demotes an ANCHOR below 3 callers to NOTE only via report, never automatically: not edited, reported"
  warn_candidates_reported: "gocyclo ≥ 15 on new or changed production functions with no @MX:WARN: Pipeline.Execute 22, rewriteRegistryClause 18, evolutionLogBlocks 17, (*logEntryDecoder).entry 17 (golangci-lint gocyclo, min-complexity 15, this tree). SHOULD-level; not added; the baseline complexity of Execute was not measured"
  below_threshold_note: "checkContained has 3 call sites in 2 functions (LoadAmendRegistry ×2, Execute ×1), below the 3-caller ANCHOR threshold; it is reached through the LoadAmendRegistry ANCHOR"
out_of_spec_guards:                     # SPEC 밖 추가, 리드 인정 — for sync-auditor review
  - "G-A: Execute rejects an empty After before Layer 1 — TestExecute_EmptyAfter_Rejected (dry_run, real; asserts 0 gate calls), mutant M-GA killed"
  - "G-B: Execute rejects a rule file that is the registry or the evolution log before Layer 1 — moved from prepareApply by lead decision 2026-09-12 (RED 0893611ad, fix 5801ebda0); TestExecute_RuleFileIsRegistryOrLog_Rejected asserts 0 gate calls; mutants M-GB, M-GB-reg, M-GB-log, M-GB-always killed at the new location, M-GA re-run killed"
behavior_change: "a rule file whose file: points at an evolution log that does not exist yet was rejected with a 'no such file' error before the G-B move and is now rejected with the G-B message before any gate (lead: intended, kept) — inferred from code reading, no test covers it"
residual_risk:
  - "G-B symlink limit: sameFile compares filepath.Abs results without symlink resolution, so a rule-file entry reaching the registry or the evolution log through a symlinked alias is not caught; no test covers that case"
  - "backup-write / temporary-write failure restore (spec.md §E.4): implemented in commitChanges but not fault-injected; approved by the operator as a blank, no seam added"
  - "Windows not run locally: GOOS=windows go vet ./internal/constitution/ and go build ./internal/cli/ pass (compile only); the test verdict on Windows is CI's"
  - "G5 (spec.md §E.4): the real-apply path through runConstitutionAmend is not exercised at CLI level; the apply is covered at Execute level"
  - "default lock path .moai/research/.amendment.lock is relative to the process working directory, not projectDir (spec.md §F, observed, not fixed)"
sync_lint: "moai spec lint SPEC-CON-AMEND-APPLY-001 is run after the sync commit; its result is reported by manager-docs to the lane and is not recorded in this block"
```
