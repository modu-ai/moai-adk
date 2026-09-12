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
| G-B | `Execute`, right after G-A, before Layer 1 (moved from `prepareApply` on 2026-09-12, see "G-B move" below) | rejects a target entry whose `file:` names the registry or the evolution log: `rule file <path>: is also the registry or the evolution log`; `sameFile` compares `ruleFilePath(projectDir, file)` against the resolved registry path and the evolution-log path — by file identity (`os.Stat` + `os.SameFile`) when both paths exist, otherwise by `filepath.Abs` results (identity branch added by F1 on 2026-09-12, see "F1 alias fix" below; before it the comparison was `filepath.Abs` only, with no file I/O) | dry-run and real (no gate is called in either) |

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

#### F1 alias fix — file identity when both paths exist (lead decision, 2026-09-12)

> G-B는 이 카드가 넣은 가드인데, 기본 macOS APFS(대소문자 무시)에서 ZONE-REGISTRY.md 같은 별칭으로 우회된다면 가드가 지킨다고 주장하는 것을 지키지 못합니다 (lead, 2026-09-12)

Sync-audit finding F1 (`.moai/reports/t659/sync-audit.md`): `sameFile` compared `filepath.Abs` strings only, so a `file:` entry that names the registry or the evolution log through an alias — a case-variant name on a case-insensitive filesystem, a hard link, or a symbolic link — passed G-B. The containment check (REQ-CAA-020/021) admits all three, because each alias lies inside `projectDir`.

Change (lead's design, `internal/constitution/pipeline.go` `sameFile` and its comment only): `os.Stat` both paths; when both succeed, return `os.SameFile(infoA, infoB)` — `os.Stat` follows links, so symbolic-link aliases are caught with hard links and case variants. When either `os.Stat` fails — the evolution log may not exist yet, and any other Stat error takes the same branch so that a Stat failure never becomes a rejection of its own — compare `filepath.Abs` results as before. G-B keeps its position (after G-A, before Layer 1) and its error text (`rule file %s: is also the registry or the evolution log`). No @MX tag sits on `sameFile`; none was added.

Test — `TestExecute_RuleFileAliasOfRegistryOrLog_Rejected`, a sibling top-level test in `internal/constitution/apply_test.go` (a sibling rather than new rows in `TestExecute_RuleFileIsRegistryOrLog_Rejected`, because each alias needs its own setup step). Every subtest runs in `dry_run` and `real` and asserts the G-B substring, the rule-file path (`containsPathForm`), 0 gate calls, the fixture tree snapshot unchanged, and the lock dir empty. Alias subtests first check, independently of `sameFile`, that the alias reads back as the registry's bytes.

- `hardlink_registry` — `file: rules/alias.md`, created with `os.Link` to the registry (skips with a stated reason if the platform refuses a hard link).
- `case_variant_registry` — `file: .claude/rules/moai/core/ZONE-REGISTRY.md`. `skipUnlessCaseInsensitive` writes a probe under a separate `t.TempDir()` and stats its case-flipped name; on a case-sensitive filesystem the subtest skips, naming that reason. **On this machine it ran** (not skipped): the RED and GREEN outputs both list `case_variant_registry/{dry_run,real}` with no `SKIP`.
- `symlink_registry` — `file: rules/alias.md`, a symbolic link to the registry (`symlinkOrSkip`).
- `absent_log_fallback` — `file: .moai/research/evolution-log.md` with no log written; pins the Abs-fallback branch. It also covers the missing-log form that the G-B move record above noted as untested.

Commits (the commit graph witnesses RED before the fix, verification-claim-integrity §2.3):

- `2496053a8` — `test(t659): G-B must catch hard-link and case-variant aliases (F1 red)`: the new test plus `.moai/reports/t659/run/guards/f1-red.txt`; production code untouched
- `bca8cf96a` — `fix(t659): compare G-B aliases by file identity when both exist (F1)`: `sameFile` only

RED on the pre-fix code, `.moai/reports/t659/run/guards/f1-red.txt` (tree `e8d16eaee` + the test):

```
$ go test ./internal/constitution/ -run '^TestExecute_RuleFileAliasOfRegistryOrLog_Rejected$' -count=1 -v
exit=1   "=== RUN" lines: 9 (1 top-level + 8 subtests)
    apply_test.go:1324: Execute: want a rule-file-is-registry-or-log error, got nil   (×6)
    --- FAIL: …/hardlink_registry/dry_run, …/hardlink_registry/real
    --- FAIL: …/case_variant_registry/dry_run, …/case_variant_registry/real
    --- FAIL: …/symlink_registry/dry_run, …/symlink_registry/real
    --- PASS: …/absent_log_fallback/dry_run, …/absent_log_fallback/real
FAIL	github.com/modu-ai/moai-adk/internal/constitution	0.442s
```

Every alias failure is the missing G-B rejection (nil error), not setup: the alias read-back and skip-probe `Fatalf`/`Skipf` lines never fire. `absent_log_fallback` passes on both trees by design.

GREEN, `.moai/reports/t659/run/guards/f1-green.txt` (tree `bca8cf96a`): `TestExecute_EmptyAfter_Rejected`, `TestExecute_RuleFileIsRegistryOrLog_Rejected`, and `TestExecute_RuleFileAliasOfRegistryOrLog_Rejected` — 3 top-level, 19 `=== RUN` lines, all PASS, 0 skipped, `ok … 0.413s`.

Mutants — runner `.moai/reports/t659/run/mutate.py`, specs `.moai/reports/t659/run/guards/mutants-guards-f1.json` (new file; the M-GA / M-GB* anchors still match unchanged and were copied only so the outputs carry the `-f1` suffix without overwriting the `-moved` evidence), per-mutant output `.moai/reports/t659/run/guards/M-*-f1.txt`, summary `mutants-f1-summary.txt`. `pipeline.go` sha256 `1ef48dcc…2467` before and after the pass (restored byte-exact; the runner also checks restoration); `apply_test.go` is not edited by any mutant (sha256 `bc5896ec…b11f`).

| Mutant | Edit | Selector | Top-level RUN / all RUN | Result | Failing subtests |
|---|---|---|---|---|---|
| M-F1-samefile | identity branch removed (straight to the Abs comparison — the pre-F1 behavior) | `^TestExecute_RuleFileAliasOfRegistryOrLog_Rejected$` | 1 / 9 | killed (exit 1) | `hardlink_registry/*`, `case_variant_registry/*`, `symlink_registry/*` — `got nil`; `absent_log_fallback/*` passes |
| M-F1-fallback | Stat-failure branch returns `false` instead of comparing Abs | `…Alias…_Rejected$/^absent_log_fallback$` | 1 / 3 | killed (exit 1) | `absent_log_fallback/dry_run`, `/real` — the rule-file read fails with `no such file or directory` after the gates, 4 gate calls |
| M-GA | G-A deleted (re-run) | `^TestExecute_EmptyAfter_Rejected$` | 1 / 3 | killed (exit 1) | `dry_run`, `real` |
| M-GB | condition → `false` (re-run) | `^TestExecute_RuleFileIsRegistryOrLog_Rejected$` | 1 / 7 | killed (exit 1) | `registry/*`, `evolution_log/*` |
| M-GB-reg | registry half dropped (re-run) | `…IsRegistryOrLog_Rejected$/^registry$` | 1 / 3 | killed (exit 1) | `registry/dry_run`, `registry/real` |
| M-GB-log | log half dropped (re-run) | `…IsRegistryOrLog_Rejected$/^evolution_log$` | 1 / 3 | killed (exit 1) | `evolution_log/dry_run`, `evolution_log/real` |
| M-GB-always | condition → `true` (re-run) | `…IsRegistryOrLog_Rejected$/^distinct_control$` | 1 / 3 | killed (exit 1) | `distinct_control/dry_run`, `distinct_control/real` |

Counted: 7 mutant runs, 7 killed, 0 survived. RUN counts: `grep -c '^=== RUN   [A-Za-z0-9_]*$'` → 1 per file; `grep -c '^=== RUN'` → the second figure.

Package state after the fix (tree `bca8cf96a`):

- `go test ./internal/constitution/ -count=1 -cover` → `ok  	github.com/modu-ai/moai-adk/internal/constitution	0.793s	coverage: 88.4% of statements` (`guards/cover-f1.txt`)
- `go vet ./internal/constitution/` → exit 0, no output (`guards/vet-f1.txt`)
- `golangci-lint run ./internal/constitution/...` → `0 issues.` (`guards/lint-f1.txt`)
- `GOOS=windows GOARCH=amd64 go vet ./internal/constitution/` → exit 0, no output (`guards/vet-windows-f1.txt`) — compile only; on Windows the hard-link and symlink subtests skip with a stated reason if the platform refuses the alias, and the case-variant subtest runs or skips by the same probe

What the alias fix leaves open is carried in §E.3 `residual_risk`.

#### D1/D2 — delta-audit follow-ups (lead decision, 2026-09-12)

> 살아남은 뮤턴트가 있는데 CHANGELOG가 그걸 확인한다고 적고 있으면 주장 쪽을 줄이는 대신 테스트로 받치는 게 맞습니다 (lead, 2026-09-12)

The delta audit (`.moai/reports/t659/sync-audit.md` § Delta audit (e5a18feae..b36f50c4c)) raised two Low/optional findings; the lead chose to fix both in this card.

- **D1** — the G-B call-site comment in `Execute` said the check "needs only the three paths, no file I/O"; since F1, `sameFile` stats both paths. The comment now reads that it needs only the three paths, which it stats, and reads or writes no file. Comment only; no code line changed.
- **D2** — `TestExecute_RuleFileAliasOfRegistryOrLog_Rejected` built registry-side aliases only, and a mutant reverting just the log half of the G-B condition survived the package, while the CHANGELOG credits this test with log-alias rejection. New subtest `hardlink_log/{dry_run,real}`: `file: rules/alias.md`, created with `os.Link` to the evolution log, which this fixture writes (skips with a stated reason if the platform refuses a hard link). Same assertions as the other alias subtests (G-B substring, rule-file path, 0 gate calls, tree snapshot unchanged, lock dir empty). The setup sanity check now compares the alias with its own target (registry or log) through a new `aliasOfLog` case field, so the log alias is checked to read back as the log's bytes, independently of `sameFile`.

Commit (the production code already handled the log alias, so the RED is observed on a mutant, not on a pre-fix tree):

- `3949cdba4` — `test(t659): back the log-alias claim with a hardlink_log subtest (D2) and fix the G-B comment (D1)`: the subtest, the D1 comment, and the evidence below

RED on the log-half mutant — spec `.moai/reports/t659/run/guards/mutants-guards-d2.json`, runner `.moai/reports/t659/run/mutate.py`, output `.moai/reports/t659/run/guards/d2-red-mutant.txt`. `M-D2-logabs` replaces only `sameFile(rulePath, evolutionLogPath)` with a `filepath.Abs` string equality of the two paths (the pre-F1 comparison); the registry half is untouched. `pipeline.go` sha256 `aa2b1965…1737` and `apply_test.go` sha256 `9eba5497…f427` before and after (restored byte-exact; `shasum -a 256 -c` → OK for both):

```
$ go test ./internal/constitution/ -run ^TestExecute_RuleFileAliasOfRegistryOrLog_Rejected$/^hardlink_log$ -count=1 -v
exit=1   "=== RUN" lines: 3 (1 top-level + 2 subtests)
    apply_test.go:1339: error "amendment validation error: rule file …/rules/alias.md: the current clause occurs 0 time(s); want exactly one" lacks "is also the registry or the evolution log"   (×2)
    apply_test.go:1345: gate doubles called 4 times, want 0 (the check runs before Layer 1)   (×2)
    --- FAIL: …/hardlink_log/dry_run (0.00s)
    --- FAIL: …/hardlink_log/real (0.00s)
FAIL	github.com/modu-ai/moai-adk/internal/constitution	0.429s
```

The failure is the missing G-B rejection: G-B admitted the alias, all four gate doubles ran, and a later apply-step check refused it with a different message. The setup sanity check did not fire.

GREEN on the real code, `.moai/reports/t659/run/guards/d2-green.txt`: `^TestExecute_RuleFileAliasOfRegistryOrLog_Rejected$` — 1 top-level + 10 subtests, all PASS, 0 skipped (`hardlink_log/{dry_run,real}` included), `ok … 0.396s`.

Guard mutant re-run on the new tree — spec `.moai/reports/t659/run/guards/mutants-guards-f1-d2rerun.json` (a copy of `mutants-guards-f1.json` with only the `out` paths and descriptions changed; every anchor matched unchanged, because the D1 edit touched a comment line none of them uses), per-mutant output `.moai/reports/t659/run/guards/M-*-f1-d2rerun.txt`, summary `mutants-f1-d2rerun-summary.txt` (derived mechanically from the output files). `pipeline.go` sha256 `aa2b1965…1737` after the pass, equal to before.

| Mutant | Selector | Top-level RUN / all RUN | Result | Failing subtests |
|---|---|---|---|---|
| M-F1-samefile | `^TestExecute_RuleFileAliasOfRegistryOrLog_Rejected$` | 1 / 11 | killed (exit 1) | 8 — `hardlink_registry/*`, `case_variant_registry/*`, `symlink_registry/*`, and now `hardlink_log/*`; `absent_log_fallback/*` passes |
| M-F1-fallback | `…Alias…_Rejected$/^absent_log_fallback$` | 1 / 3 | killed (exit 1) | 2 — `absent_log_fallback/*` |
| M-GA | `^TestExecute_EmptyAfter_Rejected$` | 1 / 3 | killed (exit 1) | 2 — `dry_run`, `real` |
| M-GB | `^TestExecute_RuleFileIsRegistryOrLog_Rejected$` | 1 / 7 | killed (exit 1) | 4 — `registry/*`, `evolution_log/*` |
| M-GB-reg | `…IsRegistryOrLog_Rejected$/^registry$` | 1 / 3 | killed (exit 1) | 2 — `registry/*` |
| M-GB-log | `…IsRegistryOrLog_Rejected$/^evolution_log$` | 1 / 3 | killed (exit 1) | 2 — `evolution_log/*` |
| M-GB-always | `…IsRegistryOrLog_Rejected$/^distinct_control$` | 1 / 3 | killed (exit 1) | 2 — `distinct_control/*` |

Counted: 7 re-runs, 7 killed, 0 survived; with `M-D2-logabs`, 8 guard mutants killed on this tree.

Package state (tree `3949cdba4`):

- `go test ./internal/constitution/ -count=1 -cover` → `ok  	github.com/modu-ai/moai-adk/internal/constitution	0.616s	coverage: 88.4% of statements` (`guards/cover-d2.txt`)
- `go vet ./internal/constitution/` → exit 0, no output (`guards/vet-d2.txt`)
- `GOOS=windows GOARCH=amd64 go vet ./internal/constitution/` → exit 0, no output (`guards/vet-windows-d2.txt`)
- `golangci-lint run ./internal/constitution/...` → `0 issues.` (`guards/lint-d2.txt`)
- `gofmt -l internal/constitution/` → no output

[SUPERSEDED by the D2 completion addendum below] Still untested on the log side: a case-variant name and a symbolic link naming the evolution log. Both reach G-B through the same `sameFile` call that `hardlink_log` now pins, so a regression confined to them would have to live inside `sameFile`, where the registry-side case-variant and symlink subtests already guard it. — This reasoning did not hold: a regression confined to the log half of the condition (not inside `sameFile`) survived, so the gap was closed by tests instead.

##### D2 completion — log-side case-variant and symlink aliases (lead rule above, 2026-09-12)

- `2160f06f4` — `test(t659): cover log-side case-variant and symlink aliases (D2 completion)`: new subtests `case_variant_log/{dry_run,real}` (`file: .moai/research/EVOLUTION-LOG.md`, same case-insensitivity probe as the registry case, `t.Skip` with a stated reason on a case-sensitive filesystem) and `symlink_log/{dry_run,real}` (`file: rules/alias.md`, a symbolic link to the evolution log via `symlinkOrSkip`, skip only where the platform refuses a symlink); both set `aliasOfLog`, same assertions as `hardlink_log`. Evidence under `.moai/reports/t659/run/guards/` (`*d2b*`, `mutants-d2b-summary.txt`).
- Branches that ran here (darwin/arm64): both new subtests ran in both modes, no skip — the temp filesystem is case-insensitive and symlinks were created. GREEN `d2b-green.txt`: 1 top-level + 14 subtests PASS, 0 SKIP.

| Mutant | Log-half replacement | Selector | Top-level RUN | Result | Failing subtests |
|---|---|---|---|---|---|
| M-D2-loglstat | inline `os.Lstat` both + `os.SameFile`, `filepath.Abs` equality on error | `…Alias…_Rejected$/^symlink_log$` | 1 | killed (exit 1) | `symlink_log/{dry_run,real}` |
| M-D2-logabs-case | plain `filepath.Abs` equality | `…Alias…_Rejected$/^case_variant_log$` | 1 | killed (exit 1) | `case_variant_log/{dry_run,real}` |
| M-D2-logabs (re-run) | plain `filepath.Abs` equality | `…Alias…_Rejected$/^hardlink_log$` | 1 | killed (exit 1) | `hardlink_log/{dry_run,real}` |

Failure reason in each killed subtest: the error lacks `is also the registry or the evolution log` (the run passed G-B and failed later on `the current clause occurs 0 time(s)`) and the gate doubles ran 4 times instead of 0 — the G-B rejection did not fire. `pipeline.go` sha256 `aa2b1965…1737` before and after (byte-exact restore). With these, 10 guard mutant runs are killed on this tree (7 re-runs + `M-D2-logabs` + the two above), covering 9 distinct mutations: `M-D2-logabs-case` is the `M-D2-logabs` mutation (byte-identical `edits`) re-run against the `case_variant_log` selector.

Package state (tree `2160f06f4`): `go test ./internal/constitution/ -count=1 -cover` → `coverage: 88.4% of statements` (`guards/cover-d2b.txt`); `go vet` darwin and `GOOS=windows GOARCH=amd64` → exit 0, no output (`guards/vet-d2b.txt`, `guards/vet-windows-d2b.txt`); `golangci-lint run ./internal/constitution/...` → `0 issues.` (`guards/lint-d2b.txt`); `gofmt -l internal/constitution/` → no output.

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
  - "G-B F1 alias fix (sync-audit F1, lead decision 2026-09-12; RED 2496053a8, fix bca8cf96a): sameFile compares by file identity (os.Stat + os.SameFile) when both paths exist — TestExecute_RuleFileAliasOfRegistryOrLog_Rejected covers hard-link, case-variant (ran here, case-insensitive temp filesystem) and symbolic-link aliases of the registry plus the Abs fallback for an absent log, 0 gate calls; mutants M-F1-samefile, M-F1-fallback killed, M-GA and the four M-GB* re-run killed (7/7)"
residual_risk:
  - "G-B identity comparison applies only when both paths can be stat'ed. When either cannot (the rule file or the evolution log does not exist yet, a symbolic link dangles, or any other Stat error), G-B falls back to comparing filepath.Abs strings, so an alias of a file that does not exist yet (for example a case-variant name of an absent evolution log) is not caught by G-B; such an entry names no existing file, and is refused later by the rule-file read ('no such file') after the gates, with no file written — that later refusal is read from the code for the alias forms; the same refusal was observed for the plain absent-log path with the fallback removed (mutant M-F1-fallback: 'no such file or directory', 4 gate calls)"
  - "G-B decides file identity once, before the gates; an alias created or removed between that check and prepareApply (while Layer 5 waits for approval) is not re-checked — read from the code, not tested"
  - "Case-variant alias subtests (case_variant_registry, case_variant_log) depend on the temp filesystem: they ran here (case-insensitive APFS); on a case-sensitive filesystem they skip with a stated reason, and the Windows cells are compile-only locally (vet), their run is CI's"
sync_report_obligation: "the sync report MUST name both guards (G-A, G-B) for sync-auditor review, as additions outside the SPEC accepted by the lead; both run before Layer 1"
sync_report_behavior_change: "the sync report's behavior-change item MUST carry one line: a rule file whose file: points at an evolution log that does not exist yet was rejected with a 'no such file' error before the G-B move and is now rejected with the G-B message, before any gate (lead, 2026-09-12: intended behavior, kept; inferred from code reading, no test covers it — a test is optional) [annotation, F1 2026-09-12: the post-move half is now covered by TestExecute_RuleFileAliasOfRegistryOrLog_Rejected/absent_log_fallback; the pre-move 'no such file' half remains read from code]"
```

## §E.4 Sync-phase Audit-Ready Signal

Sync phase run by manager-docs on 2026-09-12 inside the card worktree (branch `WT-amend-apply`, base HEAD `128a5ea52`). git-flow repository: no PR; the lane merges the branch into local develop through the lead's integration window. Nothing was pushed or merged here.

```yaml
sync_complete_at: 2026-09-12
sync_commit_sha: 1f9946188b750bd2ae6dbca36a8c948c0bb741c9   # the sync commit carrying this block; backfilled by the follow-up commit
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
  - "G-B F1 alias fix — changed AFTER the sync commit, for the sync-auditor delta check (sync-audit F1, lead decision 2026-09-12; RED 2496053a8, fix bca8cf96a, record 393a4af32): sameFile now compares by file identity (os.Stat + os.SameFile) when both paths can be stat'ed, catching hard-link, case-variant and symbolic-link aliases, and falls back to filepath.Abs equality otherwise; TestExecute_RuleFileAliasOfRegistryOrLog_Rejected (hardlink_registry, case_variant_registry, symlink_registry, absent_log_fallback; dry_run and real; 0 gate calls); mutants M-F1-samefile, M-F1-fallback killed, M-GA and the four M-GB* re-run killed (7/7); coverage 88.4% (guards/cover-f1.txt). Evidence §E.2.6 'F1 alias fix'; production diff is sameFile in internal/constitution/pipeline.go only"
behavior_change: "a rule file whose file: points at an evolution log that does not exist yet was rejected with a 'no such file' error before the G-B move and is now rejected with the G-B message before any gate (lead: intended, kept) — [corrected after F1, 2026-09-12] the new rejection is covered by TestExecute_RuleFileAliasOfRegistryOrLog_Rejected/absent_log_fallback; the pre-move 'no such file' half remains read from code. The sync-commit wording 'inferred from code reading, no test covers it' is superseded"
residual_risk:
  - "G-B alias residuals: the three items in §E.3 residual_risk are authoritative (Abs fallback when either path cannot be stat'ed, so an alias of a not-yet-existing file is not caught by G-B; identity decided once before the gates, not re-checked while Layer 5 waits; case-variant subtests (registry and log) depend on a case-insensitive temp filesystem, Windows cells compile-only locally). They replace the sync-commit line 'G-B symlink limit: sameFile compares filepath.Abs results without symlink resolution …', which the F1 fix (bca8cf96a) made false"
  - "backup-write / temporary-write failure restore (spec.md §E.4): implemented in commitChanges but not fault-injected; approved by the operator as a blank, no seam added"
  - "Windows not run locally: GOOS=windows go vet ./internal/constitution/ and go build ./internal/cli/ pass (compile only); the test verdict on Windows is CI's"
  - "G5 (spec.md §E.4): the real-apply path through runConstitutionAmend is not exercised at CLI level; the apply is covered at Execute level"
  - "default lock path .moai/research/.amendment.lock is relative to the process working directory, not projectDir (spec.md §F, observed, not fixed)"
sync_lint: "moai spec lint SPEC-CON-AMEND-APPLY-001 was run after the sync commit on tree d26cf19bc: '✓ No findings — all SPEC documents are valid', LINT_EXIT=0 (.moai/reports/t659/sync/lint.txt, committed in f0ab0ec17)"
sync_lint_gap: "GAP under verification-claim-integrity §2.2 — the sync-phase lint was judged by the installed build ~/go/bin/moai v3.2.0-rc.7, commit ed71054d3, a strict ancestor of the tree (git merge-base --is-ancestor ed71054d3 HEAD → exit 0); git diff --stat ed71054d3 HEAD -- internal/spec lists internal/spec/parser.go (77 lines changed), internal/spec/lint_coverage_sibling.go (2) and internal/spec/parser_ac_heading_anchor_test.go (143, test) as changed since, so any rule change they carry did not run. The result is not evidence for this tree; a lint by a binary built from the tree runs once in the integration-window re-measure"
post_sync_reconciliation: "follow-up commit after sync-audit (lead decisions 2026-09-12): CHANGELOG F4 dry-run clause corrected (the baseline dry-run read the registry and Layer 4 read the log; it skipped the apply step, so it never validated the amendment against the files); CHANGELOG G-B, behavior-change, guard-mutant count (5 → 7), coverage (88.3% → 88.4%) and known-limit wording aligned with the F1 fix; this block's behavior_change, residual_risk and sync_lint lines corrected, F1 added to out_of_spec_guards. sync_commit_sha unchanged"
```
