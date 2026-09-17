# Progress — SPEC-GTD-CANON-BODY-001 (card t867)

## §E.1 Plan-phase Audit-Ready Signal

### Iteration 1

- Artifacts: spec.md, plan.md, acceptance.md, progress.md (Tier M)
- SPEC ID regex self-check: `PASS`; uniqueness `ls .moai/specs | grep -ci gtd-canon` → `0`
- Base: develop `f67d2193f`, branch `WT-gtd-canon`, commit `114737ea1`
- Plan-audit iteration 1: FAIL 0.71 (`.moai/reports/plan-audit/SPEC-GTD-CANON-BODY-001-review-1.md`)

### Iteration 2 (revision for D1-D14)

- D1: AC-GCB-007/010 rebased on a named baseline (spec.md §C.3, the plan-auditor's observation at `114737ea1`, not re-measured by manager-spec; the run phase re-measures at M0). The two gtd-adjacent failures are stated as not caused by this SPEC. `-timeout` and slot lease added.
- D2/D5/D10/D13: gates moved into `check-residual.sh` and `check-gtd-body.sh`, with observed RED-now and mutant controls (acceptance.md ledger L1-L7).
- D3: mirror SKILL.md uses `.claude/skills/moai/workflows/gtd.md`.
- D4: `catalog.yaml` regeneration added to the change set and to AC-GCB-009.
- D6: CHANGELOG.md and top-level `reports/` added to the historical set (13 citation lines measured).
- D7/D14: counts measured as 19 surviving scoped files and 31 paths in the change set.
- D8: comment citations required, and the verdict-survivor escape removed.
- D9: exact test names used.
- D11: runtime and codemaps wording declared Out of Scope.
- D12: REQ-GCB-011 split into 011/012 (baseline became 013).
- Plan-audit iteration 2: PASS 0.86 (`.moai/reports/plan-audit/SPEC-GTD-CANON-BODY-001-review-2.md`)

### Iteration 2 pre-run amendments (no re-audit required)

- N1: both scripts refuse non-bash shells with exit 2 (ledger L10); AC-GCB-002 PASS-line floor = 113 (108 + 5 new set-equality checks, observed on control L6).
- N2: `002-flag-set-equal:<verb>` asserts that the frozen flag list equals the flags parsed from help, in both directions (L12 mutant goes red).
- N4: one slot lease per package, `--max-duration 45m` (30m timeout plus 15m margin). Exit 3 or 4 means wait and retry, at most 6 attempts, then record a Gap. Never run unleased, never `--force`.
- N5: an M0 timeout is a Gap; re-run once; a truncated baseline is never authoritative.
- N6: the sixth-stage regex covers six/sixth/6/6th and "answer … is a … stage" (L11 mutant goes red).
- N10: single evidence naming `.moai/reports/t867/<m0|m5>-<pkg>.{log,exit,fail-names}` in plan.md and acceptance.md.
- Extras taken: N3 (`[build failed]` / new package FAIL line counts as a new failure), N7 (inventory excludes this SPEC's own review files), N8 (`TestManifestHashFormat` added to AC-GCB-009).
- Not taken: N9 (line-based survivor rule accepts an unrelated `compat alias` on the same line; accepted as inherent).

### Scope change 0.3.0 (lead, card t854 handover)

- `TestGTDCanonicalSurfaceGolden` and `TestGTDAllTodoVerbsParity` are now owned by t867 (REQ-GCB-014, AC-GCB-012). They were removed from the baseline set, which keeps the five t854-owned tests.
- Measured at `e4cc628e9`, each test run alone:
  - L13: `gtd_canonical_surface_test.go:24 … is not a thin gtd compatibility path`. The literal the test requires predates the t860/t861 wording.
  - L14: `gtd_compat_test.go:61 gtd verbs = [.. answer ..], want [..no answer..]`. The `answer` verb landed in `1b644372d`; `moai todo` has no `answer` subcommand.
- The parity repair direction (A: `answer` stays gtd-only, test expectation only; B: expose `answer` on the todo alias) is an open operator decision, recorded in plan.md §B.1 and blocking M3.8.
- plan.md §D now requires absorbing local develop (plan-time `27220fb94`, t783 merge pending) before M0, and measuring the M0 baseline on the absorbed tree.
- This amendment changes the plan-artifact hash, so the cached iteration-2 PASS no longer satisfies the skip-eligibility hash condition; a Phase 1 re-audit follows.

### 0.3.1 — operator decision

- Kickoff approved by operator via lead 2026-09-18, autonomous progression.
- Parity repair: option A (`answer` gtd-only; `"answer"` added to `gtdWant` only; CLI unchanged). M3.8 unblocked.
- Added REQ-GCB-015 / AC-GCB-013: isolated `moai todo answer t1 x` observation (`MOAI_HOME` and `CLAUDE_PROJECT_DIR` set to temp dirs, real-queue sentinel check); a card-adding result is a finding and is not fixed.
- The M3.7 commit must cite `1dcaad954` (t860) and `61582178d` (t861), measured with `git log --oneline -- internal/template/templates/.claude/commands/moai/todo.md`.

### 0.3.2 — plan-audit iteration 3 (FAIL 0.75) revisions

- B1: `$BASE` anchor (`.moai/reports/t867/base.txt`, written at M0 after absorbing develop). Measured pre-absorption `git merge-base develop HEAD` = `f67d2193f`; develop tip = `a851b205c`.
- B2: registration-line diff guard. Controls: Long-help-only diff → grep exit 1; `AddCommand` diff → exit 0.
- B3: 1/1 numstat pins, the gtdWant line-equivalence check (good PASS; drop-engage and no-answer mutants FAIL), a literal check (verbatim PASS, loosened FAIL), and assertion counts on `4cc8ee74e`: canonical 3, compat 23. RED-now: numstat vs `f67d2193f` prints 0 lines.
- O1: `last_seq` sentinel. Plan-time read: exit 0, `last_seq` 870, `items` 143.
- O2: blocker contingency for failures of previously unreached assertions.
- O3: exact safe `answer` sentence mandated (proven by L6).

plan_audit: iteration 2 PASS 0.86 (pre-0.3.0 artifacts); iteration 3 FAIL 0.75 (0.3.1 artifacts) → revised in 0.3.2
plan_complete_at: 2026-09-18T01:45:23+09:00
plan_status: audit-ready

## §E.2 Run-phase Evidence

Run phase executed in worktree `.claude/worktrees/t867`, branch `WT-gtd-canon`, `cycle_type=ddd`.
`$BASE` = `0236646653179d83c17c1919c7fb389d510f067b` (`.moai/reports/t867/base.txt`, written at M0 from `git merge-base develop HEAD`; local develop was already absorbed by `c20b4c035` before run start, so no re-absorption occurred and `base-history.txt` was never created).
M0 pre-edit HEAD `c20b4c035`; verification HEAD `33f2d53e4`. `$BIN` = `./bin/moai`, built from the migrated tree by `make build`.
Full evidence report with verbatim output: `.moai/reports/t867/verdict.md`.

### AC matrix

| AC | Status | Verification command | Actual output |
|---|---|---|---|
| AC-GCB-001 | PASS | `bash $SPEC/check-gtd-body.sh ./bin/moai` | `fails=0`, exit 0; every `CHECK 001-*` PASS; `001-floor lines=363 floor=267` |
| AC-GCB-002 | PASS | same run | every `CHECK 002-*` PASS; 114 ` PASS ` lines (floor 113), 0 ` FAIL `; `002-required-count required=23 want=23`; `002-answer-not-stage six/sixth-stage mentions=0 want=0` |
| AC-GCB-003 | PASS | `test ! -e` on both `workflows/todo.md` paths | `local_absent_exit=0`, `mirror_absent_exit=0` |
| AC-GCB-004 | PASS | SKILL.md greps + `go test -run '^TestSkillTreeHasNoClaudeSkillDirToken$' ./internal/template/` | old-heading grep exit 1; local `CLAUDE_SKILL_DIR}/workflows/gtd.md`=1; mirror `.claude/skills/moai/workflows/gtd.md`=1; mirror `CLAUDE_SKILL_DIR`=0; test one `--- PASS:`, zero `--- FAIL:`; SKILL.md:106 routes Backlog language to `**gtd**`; SKILL.md:174 `compat alias` line names `/moai todo` and `moai todo` in both trees |
| AC-GCB-005 | PASS | `grep -rn 'workflows/todo\.md' .claude internal cmd pkg .moai/docs Makefile` | no output, exit 1 (21 lines at plan time) |
| AC-GCB-006 | PASS | `bash $SPEC/check-residual.sh` | `files=19` / `grep_rc=0 offending=0 survivors=6` / `PASS`, exit 0 (M0: `offending=69 survivors=0`, exit 1) |
| AC-GCB-007 | PASS | `./bin/moai todo list --help`; `grep -cE 'gtd.*\$ARGUMENTS'` on the 3 alias files; `TestTodoBareInvocationLists` | exit 0; counts 1 / 2 / 2 (each >= 1); test one `--- PASS:`, zero `--- FAIL:` |
| AC-GCB-008 | PASS | path grep + `go test -run '^(TestTodoSkillDocumentsHistoryVerb\|TestTodoDoctrine_MirrorParityAndStatedColumnCount\|TestTodoListJSONShapeMatchesDoc)$' ./internal/cli/` + `-run '^TestBacklogJSONDisclosure_' ./internal/template/` + `./bin/moai todo --help` | grep exit 1; cli run 3/3 `--- PASS:`; template run 2/2 `--- PASS:`; 0 `--- FAIL:` in both; `todo --help` contains `workflows/todo.md` 0 times |
| AC-GCB-009 | PASS | `make build`; porcelain; `$BASE..HEAD` catalog diff + log; `gen-catalog-hashes.go --all` + `git diff --exit-code`; 4 catalog tests; `009-cmp:*` | `make build` exit 0, `catalog.yaml updated successfully (13145 bytes)`; porcelain prints nothing; diff prints exactly `internal/template/catalog.yaml`; log lists `33f2d53e4`; regen diff exit 0; 4 `--- PASS:` / 0 `--- FAIL:`; all six `009-cmp:*` PASS |
| AC-GCB-010 | PASS | per-package leased runs at M0 and M5 (see below) | `comm -13` empty for both packages; 0 t867-owned names in either m5 fail set; both m5 logs carry package `ok `/`FAIL\t` lines; no `[build failed]` / `[setup failed]` / `panic: test timed out`; `go vet ./internal/cli/... ./internal/template/...` exit 0 |
| AC-GCB-011 | PASS | `$BASE..HEAD` diffs + inventory grep | historical diff lists only `.moai/specs/SPEC-GTD-CANON-BODY-001/*`; config diff prints nothing; verdict.md holds 141 inventory entries, equal to the grep's 141 lines, and contains `workflow.todo.enabled` |
| AC-GCB-012 | PASS | `go test -run '^TestGTDCanonicalSurfaceGolden$' ./internal/template/` and `-run '^TestGTDAllTodoVerbsParity$' ./internal/cli/`; 3 mechanical pins; registration guard | each exactly one `--- PASS:`, zero `--- FAIL:`, no `[no tests to run]`, exit 0; numstat `1	1` on both files; assertion counts HEAD 3/23 vs `$BASE` 3/23; registration grep exit 1 with `internal/cli/todo.go` in the swept diff (non-vacuous); M3.7 commit `4e511f512` body contains `1dcaad954` and `61582178d` |
| AC-GCB-013 | PASS — classified **refused** | isolated `MOAI_HOME` + `CLAUDE_PROJECT_DIR` run (steps 1-6 verbatim in verdict.md §2.3) | control `todo add` → `t1 1`, exactly one `backlog.db` under `$T/moai-home`; `todo answer t1 x` → exit 1, `"answer" is not a todo verb …`; read-back `items=1` (control only); real-queue `last_seq` 871 before and after |

Invariants:

| Invariant | Status | Evidence |
|---|---|---|
| Compat aliases keep working (REQ-GCB-007) | PASS | `./bin/moai todo list --help` exit 0; `TestTodoBareInvocationLists` PASS; `TestGTDAllTodoVerbsParity` PASS with `todoWant` untouched |
| No CLI behavior change (REQ-GCB-014) | PASS | registration-line diff guard over `internal/cli/*.go` excluding tests: grep exit 1 on a non-vacuous diff |
| `workflow.todo.enabled` not renamed (REQ-GCB-010) | PASS | `git diff "$BASE" HEAD -- internal/config .moai/config internal/template/templates/.moai/config` prints nothing |
| Historical records untouched (REQ-GCB-011) | PASS | `git diff --name-only "$BASE" HEAD -- CHANGELOG.md reports .moai/specs .moai/reports` lists only `.moai/specs/SPEC-GTD-CANON-BODY-001/*` |
| Template-First mirror parity (REQ-GCB-009) | PASS | six `009-cmp:*` byte-identical pairs; `manager-lead.toml` via `make agents-emit`, `catalog.yaml` via `make build`, neither hand-edited, both committed in `33f2d53e4` |
| No `go test ./...` (REQ-GCB-013) | PASS | not invoked in this run; only `./internal/cli/...` and `./internal/template/...` were executed |

### AC-GCB-010 baseline delta

Each package ran alone under its own `moai slot acquire --resource go-test --max-duration 45m` lease, released before the next. Every acquire returned exit 0 on the first attempt; no exit 3 / exit 4 retry was needed and `--force` was never used.

| Package | M0 fail names | M5 fail names | New (`comm -13`) |
|---|---|---|---|
| `./internal/template/...` | `TestBoundaryFlagsRecorded`, `TestGTDCanonicalSurfaceGolden`, `TestRealSetCodexShape` | `TestBoundaryFlagsRecorded`, `TestRealSetCodexShape` | *(none)* |
| `./internal/cli/...` | 10 names incl. `TestGTDAllTodoVerbsParity` | the same 10 minus `TestGTDAllTodoVerbsParity` | *(none)* |

Both t867-owned names turned green; no baseline name regressed. Evidence files: `.moai/reports/t867/{m0,m5}-{template,cli}.{log,exit,fail-names}`, `m{0,5}-check-{residual,body}.txt`, `m5-named-{cli,template}.log`.

### Findings for the lead

1. **`workflow.todo.enabled` survives under its todo name** — by decision (REQ-GCB-010). Renaming is a follow-up candidate.
2. **Pre-existing failures grew from five to ten.** spec.md §C.3 named five t854-owned failures at `114737ea1`; the M0 re-measurement on the absorbed tree found ten. The five new names (`TestRejectFactoryOnCG`, `TestResolveAgentWiringWithWizard_PrecedenceTable` + 3 subtests, `TestUpdateCodexOnlyNoClaudeResurrection`, `TestUpdatePreservesHarnessKey`) arrived with the absorbed develop commits and need an owning card. None is caused by this SPEC and none was fixed by it. The §C.3 "10-minute package timeout under load" did not reproduce (955.852s at M0, 933.214s at M5, both inside the 30m timeout).
3. **141 historical citations of the deleted path** across 55 files (`.moai/specs` 132, `CHANGELOG.md` 6, `reports/` 3), none edited per REQ-GCB-011. Full `path:line` inventory in verdict.md.
4. **Local `.claude/commands/moai/todo.md` still carries the old one-line dispatch wording** while the template mirror carries the harness-neutral two-branch form (spec.md §C.5). Finding only; a future `moai update` closes it.
5. **`check-gtd-body.sh` emits 114 PASS lines, one above the 113 floor** — the migrated section names 24 distinct flag tokens (23 required + `--json`) and `002-no-invented-flag:<flag>` emits one PASS per distinct token. `fails=0` with 0 FAIL lines; the floor guards against a count below it.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-18T04:05:00+09:00
run_commit_sha: pending-backfill-run
run_status: audit-ready
ac_pass_count: 13
ac_fail_count: 0
ac_pass_with_debt_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: not-applicable  # no push performed; lane does not push per the git-flow lane protocol
l44_post_push_fetch: not-applicable   # no push performed
new_warnings_or_lints_introduced: unmeasured  # go vet exit 0 observed; golangci-lint not run (no AC requires it) — recorded as a Gap in verdict.md §4
cross_platform_build:
  darwin: pass       # go build via `make build`, exit 0
  windows: unmeasured  # GOOS=windows build not run — no new Go code paths; recorded as a Gap in verdict.md §4
  linux: unmeasured
total_run_phase_files: 32   # git-visible, excluding this SPEC's own artifacts; plus the untracked .moai/reports/t867/ evidence set. plan.md §A predicted 31 before the 0.3.0 scope change added the two t867-owned test files.
m1_to_mN_commit_strategy: five commits on WT-gtd-canon, one per milestone group
commits:
  - 8f9b15919  # M1 + M2, carries draft -> in-progress
  - c331cf0e3  # M3.1-M3.6
  - 4e511f512  # M3.7 (cites 1dcaad954, 61582178d)
  - 87fdca7f4  # M3.8
  - 33f2d53e4  # M4 + M5.1
push_state: not-pushed  # lane reports the local state; integration and push are the lead's acts
evidence_root: .moai/reports/t867/
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
