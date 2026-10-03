# progress.md — SPEC-TEST-ENV-HERMETIC-001

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready (iteration 2 revision; re-audit pending)
- plan_complete_at: 2026-10-03
- tier: M
- artifacts: spec.md, plan.md, acceptance.md (Tier M 3) + decision-index.md (decision gate on) + progress.md
- measurement_tree: 2de0a2cb6 (branch `WT-test-env-hermetic-sweep`, clean at measurement time); iteration 2 re-measured on HEAD a5a63a0bc, whose Go sources are identical (`git rev-list --count 2de0a2cb6..a5a63a0bc` = 1; that commit's file list is five SPEC-directory paths)
- evidence: acceptance.md §D.0 ledger E-1..E-7 (commands, verbatim outputs, exit codes, tree SHAs). `.moai/reports/t1356/baseline.md` and `.moai/reports/t1356/plan-audit.md` are local-only (gitignored by operator directive 2026-09-14) and are cited by path, never committed.
- plan-phase measurement notes: the hook one-axis arms (E-3) isolate the `MOAI_KANBAN_ID` ∧ `MOAI_FACTORY_WORKERS` conjunction; the cli red is attributed to `MOAI_FACTORY_ROLE` by the explicit single-axis env of E-1 and the scrubbed arm E-1b. Whole-package runs were NOT performed at plan time (card constraint; minute-scale suites; machine load average above 60): the whole-package pairs and the `go test -list` counts are the M1 c1 obligation.
- gaps: (1) the ~357 nominated functions are unmeasured; (2) the effect of `MOAI_FACTORY_MANAGED`, `MOAI_FACTORY_CLEAR_POLICY`, `MOAI_FACTORY_AUTO_DISPATCH`, `MOAI_AUTONOMY_TIER`, `MOAI_FACTORY_SLOW_LAUNCH_MS` on any test is unmeasured (static read only); (3) `internal/cli/ptycaptest` unmeasured (own drift guard); (4) E-6 refusals and the discarded partial arms; (5) `go test -list` counts for cli and hook not taken at plan time; (6) the hook eight referenced axes outside the ID ∧ WORKERS conjunction are not isolated by one-axis arms.
- audit-ready: pending plan-audit iteration 2.

### Iteration 2 revision (plan-audit iteration 1: FAIL, 0.79 vs Tier M 0.80)

Every number the revision cites was re-measured in this run (acceptance.md ledger carries the commands and outputs). Dispositions:

| Finding | Disposition | What changed |
|---------|-------------|--------------|
| F1 AC-THE-005 not checkable from git log | fixed | One procedure (acceptance.md §D.8): c1 subject prescribed (`test(SPEC-TEST-ENV-HERMETIC-001): baseline record (card t1356)`) and its file set pinned to progress.md only; c2 identified as the single commit adding the two guard files and its file set pinned to exactly those two (`git diff-tree`); a new progress-only commit c2r (`guard red record`) separates the guard from the first fix; every `internal/` commit in `<BASE>..HEAD` other than c2 must satisfy `git merge-base --is-ancestor <c2r> x` **and** `git rev-list --count <c2r>..x` ≥ 1 (strictness is in the command: E-5 controls show `--is-ancestor A A` exits 0 while the count is 0). "Or any M4 pin" prose removed; the rule is stated once. Added beyond the finding: `<BASE>` re-derivation after a develop absorption (plan.md B10, from `gitflow-lane-protocol.md` §8) and `--no-merges`. REQ-THE-006 widened to match the AC (cross-layer sweep, `verification-completeness.md` §3). |
| F2 guard admits declared-but-unapplied scrub list; REQ-THE-005 not gated | fixed | Guard spec (spec.md §E) now has a coverage test and an applied-behaviour test: a self re-execution of the package test binary with every family axis set in `cmd.Env` (parent env minus family axes and minus the cli pin marker `factoryEnvPinnedEnv`, plus a witness) whose child asserts every referenced non-exempt axis is absent after its `TestMain`; chosen over a `TestMain` source scan because it reads behaviour and is test-file-only (REQ-THE-008 holds; precedents named). Exemption rows need a non-empty reason and a citation to an existing `*_test.go` file that references the axis. REQ-THE-009 and AC-THE-008 added; AC-THE-004 and AC-THE-008 are release-blocking on adoption (RED cell recorded at c2r on the c2 tree), required by the closure gate (§D.5); REQ-THE-005 now traces to a gated AC. |
| F3 axis enumerations do not match the tree | fixed | Family = 17 constants by the guard's own rule (name rule gives the same 17); per-package table in spec.md §A.6: cli references 17 / scrubbed 11 / uncovered 6 (adds `MOAI_FACTORY_SLOW_LAUNCH_MS`, `factory_launch_timing.go:172`); hook references 13 / scrubbed 0; discovery 1; ptycaptest 9. SlowLaunchMS carried through §A.1, §H O2, decision-index Q4 and plan M2. M3 now says the eight hook axes outside ID ∧ WORKERS are decided by guard + AC-THE-003, not by one-axis arms. The "ten constants: 22/8" figures replaced by one stated command and its counts (cli 24 production / 55 test, hook 9 / 33, discovery 3 / 2, ptycaptest 1 / 1). Not disputed but stated: the five-lane-axis test reach measures 46 cli / 28 hook with the command in spec.md §A.6 (the audit cited 47 cli — the difference is unexplained and the cited figure carries its command); `clearFactoryTestEnv(` is 44 call sites (45 occurrences with its definition) in 11 files. |
| F4 AC-THE-003 admits two arms dying early | fixed | Independent floor: L = lines of `go test -list '.*' P` starting `Test`/`Example`/`Fuzz`; each arm's terminal top-level count T (pass, fail, skip rows) must equal L; `panic: test timed out`, `goleak:` and build/setup-failure lines, and a non-zero exit with no test-level failure row, make an arm invalid. Equal skipped sets added. Lane arm stated to model the measuring session (nine axes) with the other eight unset in the same compound; the scrubbed arm unsets all 17; slot-lease and `unset … && go test … -json > <file>` forms named (the redirect form and the compound lane-arm form were verified to pass the worktree guard on a small package, E-7c and E-7d). |
| F5 E-1 lacks `-v` and the trailing FAIL | fixed | E-1 and E-2 re-executed now with `-v`; full stdout recorded including the trailing `FAIL` lines; "verbatim" qualified (as returned; durations and temp suffixes vary). |
| F6 cli scrubbed-arm green only in gitignored baseline | fixed | E-1b added with the four elements; labelled diagnostic (compound form); measured again this run (exit 0, three `--- PASS`). AC-THE-001 and AC-THE-003 cite it. |
| F7 stale-guard signal | fixed | Each package's coverage test asserts the sibling package's guard file exists and declares its guard tests; closure DoD adds `go test -list` for both guard pairs; residual (both guards deleted together, narrow selector) named honestly in spec.md §E and §D.6. |
| F8 discovery pair placement | fixed | One place: the plan-time measurement is E-7 (equal in both arms), re-recorded once in M1 c1, a contingent fix at M4; spec.md §B and plan.md aligned; the closure DoD requires the pair recorded. |
| F9 REQ-THE-008 diff vs CHANGELOG allowance | fixed | Judged at the M4 tip and again at the sync tip; CHANGELOG allowance dropped from M5 (sync adds SPEC-directory paths only). |
| F10 E-4a does not read the added axis | fixed | Disclosed under E-4; AC-THE-007 now cites the applied-behaviour test (AC-THE-008) as the separation control that reads every added axis. |
| F11 literal rule in plan §D.5 | fixed | Stated: only the exact quoted `"MOAI_FACTORY_ROLE"` is forbidden in `internal/hook` (`envkeys_factory_role_test.go:60`, verified); unquoted prefixes are legal; `config.EnvAutonomyTier` for the one full name needed. |
| F12 file count and re-exec claim | fixed | "about 5 files" corrected to 4 (two per package) in spec.md §D; plan B9 now names the four verified `exec.Command(os.Args[0], …)` sites and drops `session_start_drift_fill_test.go` (its `os.Executable()` hits are comments and an assertion message). |
| F13 milestone exits not bindable | fixed | `Exit: AC-THE-00N` lines in M1-M4; the CN-4 verb extracted from `plan-auditor.md` and run: `COLLECTED: 5 milestones in plan order (M1 M2 M3 M4 M5), 8 exit bindings, 12 ordering candidates`, no `CONFLICT:` line. |

## §E.2 Run-phase Evidence

_<pending run-phase — owned by manager-develop>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — owned by manager-develop>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — owned by manager-docs>_
