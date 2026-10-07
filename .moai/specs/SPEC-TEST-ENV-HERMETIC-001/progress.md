# progress.md — SPEC-TEST-ENV-HERMETIC-001

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready (v0.8.1; the plan-audit surface CLOSED at iteration 13 under the leader-bounded disposition (ii), accepted tree HEAD `46b8eeb75` (branch `WT-test-env-hermetic-sweep`, clean), receipt `rcpt-9f93b72c6eb4dbba07f004bf`, verdict `.moai/reports/t1356/plan-audit-iter13.md` — this line previously described the v0.8.0/iteration-11 state; refreshed 2026-10-06 on the first §E.2 write per the leader's binding condition BC-3)
- plan_complete_at: 2026-10-03
- tier: M
- artifacts: spec.md, plan.md, acceptance.md (Tier M 3) + decision-index.md (decision gate on) + progress.md
- measurement_tree: 2de0a2cb6 (branch `WT-test-env-hermetic-sweep`, clean at measurement time); iteration 2 re-measured on HEAD a5a63a0bc, whose Go sources are identical (`git rev-list --count 2de0a2cb6..a5a63a0bc` = 1; that commit's file list is five SPEC-directory paths)
- evidence: acceptance.md §D.0 ledger E-1..E-9 (commands, verbatim outputs, exit codes, tree SHAs). `.moai/reports/t1356/baseline.md`, `.moai/reports/t1356/plan-audit.md` and `.moai/reports/t1356/plan-audit-iter2.md` are local-only (gitignored by operator directive 2026-09-14) and are cited by path, never committed.
- plan-phase measurement notes: the hook one-axis arms (E-3) isolate the `MOAI_KANBAN_ID` ∧ `MOAI_FACTORY_WORKERS` conjunction; the cli red is attributed to `MOAI_FACTORY_ROLE` by the explicit single-axis env of E-1 and the scrubbed arm E-1b. Whole-package test runs were NOT performed at plan time (card constraint; minute-scale suites; machine load average above 50): the whole-package pairs are the M1 c1 obligation; the `go test -list` counts were taken at iteration 3 (E-8: cli 4884, hook 1322) and are re-recorded at c1.
- gaps: (1) the ~357 nominated functions are unmeasured; (2) the effect of `MOAI_FACTORY_MANAGED`, `MOAI_FACTORY_CLEAR_POLICY`, `MOAI_FACTORY_AUTO_DISPATCH`, `MOAI_AUTONOMY_TIER`, `MOAI_FACTORY_SLOW_LAUNCH_MS`, `MOAI_FACTORY_MANAGED_TUI`, `MOAI_FACTORY_APP_SERVER_TOKEN` on any test is unmeasured (static read only; the two absorption-added axes joined this list at the v0.8.0 revision, 2026-10-06); (3) `internal/cli/ptycaptest` unmeasured (own drift guard); (4) E-6 refusals and the discarded partial arms; (5) T == L on the whole cli and hook packages is unobserved (whole-package runs are the c1 obligation); (6) the hook eight referenced axes outside the ID ∧ WORKERS conjunction are not isolated by one-axis arms; (7) the guard pair, the applied-behaviour probe and the c1, c2 and c2r commits do not exist yet — their behaviour is specified, not observed.
- disclosed design residual (iter12 N4, leader acceptance 2026-10-06 — recorded, not a defect to repair): the c1 lane arm carries the secret-valued axis at its real value while the final lane arm replays the `<redacted>` literal, so the clause (f) byte-identity over that axis proves presence and mask form only; a test whose verdict depends on the token's VALUE rather than its presence could differ between c1 and final for a reason clauses (b)/(e) do not isolate.
- audit-ready: closed — the run phase is authorized on the accepted tree by the leader's message of 2026-10-06 plus that verdict (the audit-cross decision record). The run carries the leader's binding conditions: BC-1 — at run phase ANY command 11 / skip-names difference is treated as a clause (b) failure unless the load/environment cause is independently proven, and a difference attributable to the audited family axes is NEVER excusable (a finding, not noise); BC-2 — the skip-equality excusal boundary stays recorded design debt, applied as written, no mid-run redesign; BC-3 — the §E.1 refresh on the first §E.2 write (this write). `.moai/reports/t1356/plan-audit-iter3.md` through `plan-audit-iter13.md` are the audit reports, local and gitignored, cited by path only.

### Iteration 2 revision (plan-audit iteration 1: FAIL, 0.79 vs Tier M 0.80)

Every number the revision cites was re-measured in this run (acceptance.md ledger carries the commands and outputs). Dispositions:

| Finding | Disposition | What changed |
|---------|-------------|--------------|
| F1 AC-THE-005 not checkable from git log | fixed | One procedure (acceptance.md §D.8): c1 subject prescribed (`test(SPEC-TEST-ENV-HERMETIC-001): baseline record (card t1356)`) and its file set pinned to progress.md only; c2 identified as the single commit adding the two guard files and its file set pinned to exactly those two (`git diff-tree`); a new progress-only commit c2r (`guard red record`) separates the guard from the first fix; every `internal/` commit in `<BASE>..HEAD` other than c2 must satisfy `git merge-base --is-ancestor <c2r> x` **and** `git rev-list --count <c2r>..x` ≥ 1 (strictness is in the command: E-5 controls show `--is-ancestor A A` exits 0 while the count is 0). "Or any M4 pin" prose removed; the rule is stated once. Added beyond the finding: `<BASE>` re-derivation after a develop absorption (plan.md B10, from `gitflow-lane-protocol.md` §8) and `--no-merges`. REQ-THE-006 widened to match the AC (cross-layer sweep, `verification-completeness.md` §3). |
| F2 guard admits declared-but-unapplied scrub list; REQ-THE-005 not gated | fixed | Guard spec (spec.md §E) now has a coverage test and an applied-behaviour test: a self re-execution of the package test binary with every family axis set in `cmd.Env` (parent env minus family axes and minus the cli pin marker `factoryEnvPinnedEnv`, plus a witness) whose child asserts every referenced non-exempt axis is absent after its `TestMain`; chosen over a `TestMain` source scan because it reads behaviour and is test-file-only (REQ-THE-008 holds; precedents named). Exemption rows need a non-empty reason and a citation to an existing `*_test.go` file that references the axis. REQ-THE-009 and AC-THE-008 added; AC-THE-004 and AC-THE-008 are guard criteria whose cell is completed by the c2r record on the c2 tree (relabelled at iteration 3, D4), required by the closure gate (§D.5); REQ-THE-005 now traces to a gated AC. |
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

### Iteration 3 revision (plan-audit iteration 2: FAIL, 0.86 vs Tier M 0.80, driven by one must-fix mutant hole)

The report is `.moai/reports/t1356/plan-audit-iter2.md` (local, gitignored, cited by path only). Every figure the revision relies on was re-measured in this run at HEAD `669cf18c9` (Go sources equal `2de0a2cb6`): family 17, the quoted-literal and identifier reference checks, L cli 4884 / hook 1322, the E-5 controls, the `comm -13` and `git grep` forms, `clearFactoryTestEnv(` 45 occurrences in 11 files. REQ count stays 9 and AC count 8 (Tier M ceiling 16 each).

| Finding | Disposition | What changed |
|---------|-------------|--------------|
| D1 AC-THE-003 admits a change-induced failure as env-unrelated (must-fix) | fixed | Clause (e) added to AC-THE-003 and to REQ-THE-003: every name in a final-tree arm's failing set must be in the c1 failing set of the same arm type; the difference `final − c1` is recorded by three plain commands (acceptance.md §D.3 commands 6-8: `grep -oE` to a names file, `sort -o`, `comm -13 <c1 names> <final names>`) and required empty. The form was run on three small real arms (narrow selector, `./internal/hook`): it prints the one name outside the baseline, prints both names against an empty baseline, and prints nothing for a subset or an identical set (all exit 0). R6 reworded ("identical in both arms AND present in the c1 baseline"). The §D.1 mutant probe now names the closed mutant: a fix that breaks a test setting a family axis itself before `clearFactoryTestEnv(t)`, failing in both final arms. DoD requires the sorted c1 name files in progress.md §E.2; plan M1, M4 and AP-12 carry it. |
| D2 AC-THE-004 mutant-probe sentence overstates closure | fixed | The sentence now names the harmful variant (an unscrubbed axis X plus an exemption row `{X, "n/a", <an existing *_test.go that references X>}`, which passes both the coverage and the applied test) and states it is left to review; DoD item added: the final exemption tables are listed in progress.md §E.2 and each surviving row cites the M4 whole-package scrubbed-arm test that went red when that axis was stripped (empty table: no row). Reconciled in spec.md §E and §H O2 and plan D8. |
| D3 hook coverage test has two reds at c2 | fixed (option 1) | Liveness assertions use `t.Errorf` and never stop the comparison, so the hook coverage test at c2 carries both the empty-scrub-set message and the thirteen axis names; the cli coverage test at c2 carries the six names only. Stated consistently in spec.md §E input (3), plan D4/D5/M1 and AC-THE-004; the c2r cell must record which message(s) the observed red carries. |
| D4 class label; c2r content witness | fixed | AC-THE-004 and AC-THE-008 carry "regression-guard at plan time; becomes release-blocking when its four-element cell is recorded at c2r" everywhere (acceptance matrix, DoD, spec.md HISTORY, this file); the old label is gone (`grep -rn` over the SPEC directory returns no match). AC-THE-005 step 7 added (acceptance.md §D.8): `git grep -c --all-match` with the four guard-test `--- FAIL: <name> ` patterns on the c2r commit prints one line (one printed line = every name present), plus `git grep -c -e '<c2>'` for the c2 SHA; controls recorded as ledger E-5d at tree `669cf18c9` (present dashed pattern, `--all-match` with two present names, with one absent name, a REQ id). The sync auditor re-executes one of the four cells at the c2 tree (matrix preface, step 7, DoD). The lint's `VacuousTestAssertion` warning on the first draft of the patterns (no space after the name) was fixed by ending each pattern in a space and re-running the controls. |
| D5 Exit binding of AC-THE-004/008 vs probes at M4 | fixed | plan M3 `Exit: AC-THE-002, AC-THE-007`; plan M4 `Exit: AC-THE-003, AC-THE-004, AC-THE-006, AC-THE-008`; M3 keeps the green path in the matrix; the matrix preface states the Exit/green-path split and that M1 evaluates c1/c2/c2r shape and chain (steps 0-5, 7) while M4 re-evaluates the step 6 enumeration. The CN-4 verb was refused by the worktree guard (as for the auditor); a line-for-line Python port run from a script file printed `COLLECTED: 5 milestones in plan order (M1 M2 M3 M4 M5), 8 exit bindings, 13 ordering candidates` and no `CONFLICT:` line (a port, not the canonical verb; the auditor decides CN-4). |
| D6 lease cap vs `-timeout` | fixed | Cap = `max(20m, 1.5 x the longest whole-package runtime recorded so far in progress.md §E.2)` (20m for the first c1 arm); `-timeout` = the cap minus 2m, strictly below it. spec.md §G, acceptance §D.3, plan §C step 5. |
| D7 applied-behaviour probe mechanics | fixed | plan D7: bounded `context.WithTimeout` on the child (precedent `codex_launcher_exec_posix_test.go:121`, 20 s, verified; cli child `TestMain` runs `warmUpCommandTree`, `main_test.go:380`, verified; the bound is sized from a measured child runtime under load, never below 20 s), child output logged on success as well as failure, case-insensitive key matching when stripping family axes on Windows. |
| D8 reference rule, liveness floor, Go-only disclosure | fixed | Reference pinned in spec.md §A.6 and §E: identifier `config.<Name>` OR quoted literal (`os.Getenv("MOAI_AUTONOMY_TIER")` at `codex_sync_gate.go:262` verified; the counts 17 and 13 hold under either form — measured: the one quoted family literal in `internal/cli` is also referenced by identifier at `kanban.go`, and `internal/hook` production has none); family-size floor 17 added to liveness input (3) beside `referenced >= 1`; the guard's Go-only reach disclosed in §G R4 (shell-gate tests passing `tierEnv(...)`, `sync_gate_failstate_test.go:303` verified, are outside it). |
| D9 stale E-5 control; stale E-6 excuse | fixed | E-5 control pinned `2de0a2cb6..a5a63a0bc` (prints the one commit; the `..HEAD` form prints two, re-measured); E-6 item (5) withdrawn; plan-time L recorded as ledger entry E-8 (cli 4884, hook 1322, re-measured with `go test -list` only); plan §C step 6 and §D.3 updated. |
| D10 REQ-THE-009 trigger wording | fixed | "When the guarded package's applied-behaviour test starts the package test binary with every lane/kanban gate axis present …, the applied-behaviour test shall fail and shall name each such axis." |
| D11 vocabulary for a cell completed by a later record | fixed | One phrase, "cell completed by the c1 record" / "by the c2r record", in the matrix preface, the AC-THE-003, 004 and 008 RED-now cells, §D.2 and the DoD. AC-THE-003 keeps the plain class (its witness-by-existence cell is re-executable now). |

Unobserved at this iteration (Gaps): the guard pair, the applied probe, and the c1/c2/c2r commits do not exist (specification only); no whole-package `go test` was run (instruction; machine load above 50), so T == L on the whole cli and hook packages and the absence of any test that sets a newly scrubbed axis before `clearFactoryTestEnv(t)` remain unobserved (the `comm -13` form was exercised on narrow arms only); the canonical CN-4 awk verb was refused by the worktree guard and replaced by a port.

### Iteration 4 revision (plan-audit iteration 3: FAIL, 0.87 vs Tier M 0.80, two must-fix holes in AC-THE-003)

Delta-scoped; REQ count stays 9 and AC count 8 (ceiling 16 each). Every cited form was re-run in this session (acceptance.md ledger E-9); no Go file touched, no whole-package `go test` run.

| Finding | Disposition | What changed |
|---------|-------------|--------------|
| MF-1 lane arm has no positive control | fixed | Clause (f) in AC-THE-003 and §D.3: the lane-arm env (set with values, unset) is recorded per arm and the four lines (c1 and final, cli and hook) are identical; the arm must carry every modelled axis (the pre-flight family read, minimum ROLE=lane plus non-empty WORKER, WORKERS, KANBAN_ID); the c1 lane names file holds the five reds (command 10: 3 cli / 2 hook) and the c1 scrubbed names file none; a session without lane axes makes the arm INVALID and the AC failed, with a blocker report. DoD §E.2 list gains the env lines and command 10 outputs; the §D.1 "echoed" sentence is reworded to cite clause (f). Pre-flight read made an explicit step with its command (plan §C step 2, M1 c1, M4 step 1), R8, REQ-THE-003, AP-13. |
| MF-2 clause (e) compares top-level names only | fixed | §D.3 commands 6-8 now match `"Test":"[^"]+"` (subtests included); commands 1-3 and L stay top-level where they must match `go test -list`; command 2/3 also subtest-inclusive; new command 9 (`comm -3`, clause (b) by full path). REQ-THE-003 and REQ-THE-007 state the unit is the Go test row, subtests included. Controls E-9: arms A/B/C, a subtest case (the new form prints `TestParent/sb`, the old form nothing), build and `TestMain` failures with no test-level row (invalid by commands 4 and 5). |
| SF-1 exemption row vs equality | fixed (smaller change) | A test red in the all-unset scrubbed arm is fixed by an arm-independent pin; a true exemption axis is left at its lane value in the re-run scrubbed arm, so it is identical-valued in both arms and (b) holds. Stated alike in acceptance.md (AC-THE-003(b), §D.3, DoD), plan.md M4 step 2 and the spec's exemption text is unchanged. |
| SF-2 c2r witness proves strings, not runs | fixed | The c2r cell carries the exact command, the exit code as its own field, the c2 SHA and the pre-run `git rev-parse --short HEAD` read; the sync auditor re-executes one cell from a `git archive` extraction of c2 (no worktree) and records its own stdout and exit code; acceptance.md §D.8 step 7 says plainly that string presence is not execution evidence and the re-execution closes it. |
| SF-3 no repeat rule | fixed | One repeat of the affected whole-package arm for a name printed by command 8 or 9; only a name failing in both runs counts; comparison stays by full path (acceptance.md §D.3, spec.md R7). Does not reopen MF-2 (a repeat is a whole-package arm, never an isolated re-run). |
| SF-4 bare `return` hollowing | partly fixed, residual stated | AC-THE-006 gains two greps on the saved modified-file diff (removed assertion lines; added bare `return` / `testing.Short()` guards), controls 45 / 5 and an empty card-range diff (E-9). Not closed mechanically: a hollowing built from added lines alone (a swallowing `defer`, panic and recover) is invisible to a line grep and is left to review (acceptance.md §D.1). The `--- PASS` line of each previously red test in the final tree is already required by AC-THE-001/002. |
| N1 wording slip (spec §F) | fixed | "every commit other than c2 that touches `internal/`". |
| N2 `LC_ALL=C` | fixed | On `sort -o` and `comm` (verified: both forms still behave, exit `0`). |
| N3 Bash 600 s ceiling vs lease cap | fixed | Stated once in spec §G, plan §C step 5 and acceptance.md header: the cap minus 2m is at least 18m, a foreground call ends at 600 s, so each whole-package arm is a background Bash call. |

Unobserved at this iteration (Gaps): no whole-package run; the c1 lane arms and the five reds in them are specified, not observed (the narrow-selector compound lane form was measured, E-9); a session without lane axes was not observed; the guard pair and the c1/c2/c2r commits do not exist.

2026-10-04 (SF-5 disposition): the v0.4.0 `CONFLICT:` line the iteration-4 CN-4 port printed — `CONFLICT: acceptance.md:291 orders AC-THE-004 before M2 but plan binds it to M4` — was a cross-cell false positive of the verb: the matched keyword `before` belonged to the `git rev-parse --short HEAD` clause of a different cell of the same table row, not to AC-THE-004's own ordering, and the AC's obligations (RED on c2, green path M2/M3, probes recorded at M4, Exit at M4) are jointly satisfiable. Manual read judged it PASS (MP-9, iteration 4). v0.5.0 additionally removes the trigger at the source: the acceptance.md c2r clause is reworded without an ordering keyword, so the canonical verb prints no `CONFLICT:` line for the row.

2026-10-04 (iteration 5 revision round): develop tip `30ce3a02d` (551 commits) absorbed at merge `960ea3012`; the five SPEC-directory files are byte-identical through the merge (`git diff 646a7860d..960ea3012 --stat -- .moai/specs/SPEC-TEST-ENV-HERMETIC-001/` is empty). The five observed reds were re-observed on `960ea3012` and are intact: `env MOAI_FACTORY_ROLE=lane go test ./internal/cli -count=1 -run '^TestTodoClaim_LaneGovernance$|^TestTodoClaimMCP_Mirror$|^TestTodoPickInFactoryProvenanceFailsOpenWithoutSpecOrGit$'` exits FAIL with `--- FAIL` x3 (todo_claim_test.go:291, todo_claim_test.go:416, todo_test.go:190, each `refused — lane boundary`); `env MOAI_FACTORY_ROLE=lane MOAI_FACTORY_WORKER=lane-27 MOAI_FACTORY_WORKERS=0 MOAI_KANBAN_BACKEND=claude MOAI_KANBAN_ID=tm9i7y go test ./internal/hook -count=1 -run '^TestStaleRunNoticeLegacyLeaderSpelling$|^TestStaleRunNoticeLegacySessionRecord$'` exits FAIL with `--- FAIL` x2 (stale_run_m1_test.go:42, stale_run_m1_test.go:85). REQ count stays 9 and AC count 8 (ceiling 16 each).

### Iteration 5 revision (plan-audit iteration 4: FAIL, 0.87 vs Tier M 0.80, one new must-fix MF-3 plus a required-backend FAIL from codex)

Delta-scoped; the report is `.moai/reports/t1356/plan-audit-iter4.md` (local, gitignored, cited by path only). The E-9 escape-aware control was measured in this session (acceptance.md ledger E-9, final bullet); no Go file touched, no whole-package `go test` run.

| Finding | Disposition | What changed |
|---------|-------------|--------------|
| MF-3 the names regex truncates quote-bearing subtest names | fixed | The name pattern is escape-aware (`"Test":"([^"\\]|\\.)+"`) in acceptance.md §D.3 commands 2 and 6, and command 3 (skip rows) states it uses the same pattern; the naive `[^"]+` stopped at the first `\"` so two quote-bearing names sharing the text up to the quote collapsed into one line and `comm -13` stayed empty. Measured control in E-9 (scratch module `t1356e9`, tree-independent): the old form truncates `q"uote` to `TestParent/q\` and its `comm -13` prints nothing across `q"uote` → `q"uoted` (exit 0, the miss); the new form prints both names whole plus `back\\slash` and its `comm -13` prints `"Test":"TestParent/q\"uoted"`, the new failing name (exit 0). Commands 1 and the L-count form stay top-level (`[^/"]+`, unaffected). |
| SF-1 REQ/AC divergence on what the scrubbed arm is | fixed | REQ-THE-001/002 each gain one clause: the scrubbed arm is a scrubbed env that leaves an axis carrying a reasoned, cited exemption (spec.md §E) at its lane value, so the exempt axis carries the identical value in both arms — the requirement layer now matches acceptance.md §D.3(b), §D.3 and plan.md M4; the exemption tables themselves untouched. |
| SF-2 c2r witness: one-of-four re-execution overstates closure; `git status --short` missing | fixed (reword option) | The c2r cell obligations (AC-THE-004 row, AC-THE-008 row, DoD) gain the `git status --short` read (expected empty) taken beside the pre-run `git rev-parse --short HEAD` read; the DoD's "this re-execution is what closes that gap" is reworded to close the gap for the one cell it re-executes, the other three cells staying witnessed by their recorded strings. The one-of-four sampling is unchanged. |
| SF-3 repeat rule under-specified | fixed | §D.3: every name printed in either run is recorded with both outputs; a name printed only in the repeat counts as a hit against (b) or (e), never load noise; load noise is the reverse case; for command 9 BOTH arms of the pair are repeated and the repeat comparison runs between the two repeat files (acceptance.md §D.3, spec.md R7 — R7 now cites commands 8 and 9). |
| SF-4 command 10 cited for any lane arm; final lane-arm env self-recorded | fixed | §D.3 command 10 is scoped to the c1 lane arms in the validity paragraph (it runs on the recorded c1 names files); a final-tree lane arm's lane-ness rests on its recorded env line plus its child-visible env line — the new `unset <family> && env <nine lane axes> env > <file>` witness, §D.3 lane-arm definition — identical to the c1 line; carried in AC-THE-003 clause (f) and the DoD §E.2 evidence list. |
| SF-5 CN-4 verb false positive at the acceptance.md c2r row | fixed (both halves) | Dated disposition note above records the iteration-4 judgment (cross-cell false positive, PASS by manual read, MP-9); v0.5.0 removes the trigger at the source — the row's clause is reworded to "taken as the immediately preceding step of the run", carrying none of the verb's ordering keywords. |
| SF-6 weakened-condition hollowing unnamed | fixed | §D.1 residual names the second hollowing: `- if got == "" {` / `+ if false {` removes a line carrying no assertion token, so the removed-assertion grep stays `0` — left to review like the added-lines-only residual. |
| N1 skipped-set equality has no command form | fixed | New §D.3 command 11: per arm, an escape-aware names file over the `"Action":"skip"` rows (`grep -oE ... > <skip names file>`, `sort -o`), then `LC_ALL=C comm -3 <c1 skip names> <final skip names>` must print nothing; the "the skipped sets are equal" clause cites command 11; command 3 states it reads the same rows; the DoD §E.2 evidence list gains the command 11 outputs. |

Unobserved at this iteration (Gaps): the E-9 escape-aware control ran on a scratch module (tree-independent, go1.26.8), not on real whole-package JSON streams; whether any c1 failing subtest of the real packages carries a quote is decided by the c1 names files at M1; the guard pair and the c1/c2/c2r commits still do not exist; no whole-package run (lane-local rule).

### Iteration 6 revision (plan-audit iteration 5: FAIL, 0.82 vs Tier M 0.80 and below iteration 4's 0.87 — STOP signal; three new blocking defects plus a should-fix)

Delta-scoped; the report is `.moai/reports/t1356/plan-audit-iter5.md` (local, gitignored, cited by path only). The leader's second cross-session dispatch of 2026-10-04 authorized this one delta round. The census was re-measured in this session (record below); REQ count stays 9 and AC count 8 (ceiling 16 each).

| Finding | Disposition | What changed |
|---------|-------------|--------------|
| I5-D1 R7's tail inverts the repeat rule | fixed | spec.md R7's tail rewritten to the polarity §D.3 defines — a name printed in both runs counts as a hit, a repeat-only name also counts as a hit and is never load noise, load noise is the initial-run-only case — keeping the command 8/9 citations and the iteration-4 SF-3 widening (not the v0.4.0 narrower text). Polarity: P1 in the table below. |
| I5-D2 command 11 measures the wrong pair for clause (b) | fixed | §D.3 command 11 re-wired to the cross-arm pairing — per tree stage (c1 and final), escape-aware skip-rows names files from each arm, then `LC_ALL=C comm -3 <lane skip names> <scrubbed skip names>` must print nothing on the final tree (the c1-stage run is the baseline record); the c1-vs-final form kept as a labelled no-new-skip extra; the clause citation (validity paragraph, AC-THE-003(b), §D.4) states the axis; the DoD carries the c1-side skip names files and the restated command 11 outputs. Polarity: P2-P3. Command 3 (arm-file skip-row reader) unchanged. |
| I5-D3 family census stale after the absorption (floor 17 unsatisfiable at 16) | fixed | Census re-derived at HEAD `8cb2444e7` in this session (command and counts below): **16** axes (17 at `646a7860d`). Floor restated to 16 (spec.md §E input (3) with the re-derivation rule, §A.6 with the census command, plan §C step 2 and D4, AC-THE-004); per-package sets restated (cli 16/16 referenced, 8 covered, 8 uncovered; hook 10/16; discovery 1; ptycaptest 6); expected guard reds restated (eight cli axes, ten hook axes) in AC-THE-004, AC-THE-008, plan D5/M1/M2/M3; R8's remainder 8→7; §H O2 and decision-index Q4 carry the seven-axis list; plan.md's census-carrying lines swept (list below). The five observed reds and command 10's 3/2 counts untouched (verified unchanged). Polarity: P4-P6. |
| I5-D4 credential-class axis recorded verbatim in the env recording | fixed | Secret-valued axes marked in the §A.6 census (at HEAD exactly one: `MOAI_FACTORY_APP_SERVER_TOKEN`) with one redaction rule stated once there and applied in §D.3's arm-env line, the child-visible env witness, AC-THE-003 clause (f) and the DoD evidence list: the recorded forms are `NAME=<redacted>` and the byte-identity comparisons run over the redacted forms. The redaction narrows nothing the witness needs — presence and the non-secret values, never the secret's value. Polarity: P7. |
| I5-D5 stale references (minor, audit-directed "next artifact touch") | fixed (incidental) | The four lines refreshed: `spec.md` §A preamble and `progress.md` §E.1 "E-1..E-8" → "E-1..E-9"; `spec.md` §I lists iter4/iter5 reports; `progress.md` plan_status/audit-ready lines carry the second dispatch. `plan.md:21`'s "ledger E-1..E-8" NOT touched — plan.md is census-limited this round; flagged to the leader. |

Census re-derivation (I5-D3), measured in this session at HEAD `8cb2444e7` (clean tree, worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1356`, branch `WT-test-env-hermetic-sweep`):

- family: `grep -oE '"MOAI_(FACTORY_|KANBAN)[A-Z_]*"|"MOAI_AUTONOMY_TIER"' internal/config/envkeys.go`, `sort -u`, `wc -l` → **16** (exit 0); the same on `git show 646a7860d:internal/config/envkeys.go` → **17**.
- diff 646a7860d → 8cb2444e7: − `MOAI_KANBAN`, − `MOAI_KANBAN_SPEC`, − `MOAI_KANBAN_LABEL`; + `MOAI_FACTORY_APP_SERVER_TOKEN` (`envkeys.go:302`, passed as `--remote-auth-token-env` at `managed_codex_tui.go:338,345`), + `MOAI_FACTORY_MANAGED_TUI` (`envkeys.go:296`, read at `managed_codex_tui.go:186`).
- per-package production references (identifier `config.<Name>` OR quoted literal, non-`_test.go`): cli **16** of 16; hook **10** of 16; discovery **1** (KanbanID, 3 production files); ptycaptest **6** (the Kanban axes).
- cli start-up scrub covers **8** of the 16 (Workers, Worker, KanbanID, KanbanSettingsInjected, KanbanLeadAddr, KanbanBackend, KanbanCard, KanbanLeadName; the three retired names stay in the slice as markers) → uncovered **8**; hook uncovered **10**; expected c2 guard reds: eight cli axes, ten hook axes.
- reach at `8cb2444e7` (either-form rule per file): cli 27 production / 67 test (26 identifier-only); hook 8 / 34; discovery 3 / 2; ptycaptest 1 / 1; five-lane-axis test files cli 53 (54 with ptycaptest), hook 28.
- floor restated **16**; R8 remainder **7** (16 − 9 session axes).

plan.md lines touched (census/floor-carrying sweep): §C step 2 (family count 17→16 with re-derivation rule; "other eight"→"other seven"), §E D4 (floor 17→16; absorption re-derivation sentence added to the relief valve), §E D5 (thirteen→ten, six→eight), M1 c2r (six/thirteen→eight/ten, ×2), M2 (five→seven uncovered axes, "all six"→"all eight", static-read scope note), M3 (thirteen→ten; the eight-axis isolation list re-derived to the current hook referenced set). decision-index.md: Q4 only. Places checked and unchanged: spec.md §A.5 (helper counts), §D (helper/call-site counts), §E liveness wording otherwise, §H O3 (discovery still 1 axis / 3 production files — re-measured), acceptance.md E-9's dated clause (f) control ("the eight family axes the session lacks", "the 17-axis scrubbed arms" — dated records at `47ae8ecb1`, pre-absorption, left as measured), E-8's L counts (not family-bound), §D.3's nine-axis session list (all nine axes still in the family).

#### Polarity-comparison table (leader requirement: every rewritten normative sentence, old vs new)

| # | Sentence (site) | Old (v0.5.0) | New (v0.6.0) | Polarity relation |
|---|-----------------|--------------|--------------|-------------------|
| P1 | spec.md R7 tail (I5-D1) | "a name printed in either run is recorded with both outputs, a name printed only in the repeat counts as a hit, and only a name failing in both runs is excused as load noise" | "every name printed in either run is recorded with both outputs: a name printed in **both** runs counts as a hit against the clause it was printed for, a name printed only in the repeat — absent from the initial run — **also** counts as a hit and is never load noise, and load noise is the reverse case, a name printed in the initial run and absent from the repeat" | First clause (repeat-only = hit): same direction, kept. Second clause: **deliberately reversed** — old excused the both-runs case; new makes both-runs a hit and excuses the initial-run-only case, matching the §D.3 rule R7 is cited to govern and keeping the SF-3 widening (the v0.4.0 "only a name failing in both runs counts" narrower rule was NOT restored). |
| P2 | acceptance.md §D.3 command 11 comparison (I5-D2) | "`LC_ALL=C comm -3 <c1 skip names file of the same arm type> <final skip names file>` must print nothing" | "`LC_ALL=C comm -3 <lane skip names file> <scrubbed skip names file>` of the same tree stage ... must print nothing on the final tree"; c1-vs-final kept as a labelled extra | **Deliberately changed axis** — the "must print nothing" verdict predicate is unchanged in direction; the pairing it judges moved from c1-vs-final (within arm type) to lane-vs-scrubbed (across arms), the pairing clause (b) states and command 9 uses for the failing half. The extra keeps the old pairing with the same direction. |
| P3 | acceptance.md validity paragraph + AC-THE-003(b) + §D.4 (I5-D2 citations) | "the skipped sets by command 11, the same form over the skip rows" / "and their skipped sets are equal;" / "the two skipped sets are equal" | "the same cross-arm pairing over the skip rows — the lane arm against the scrubbed arm of the same tree" / "and their skipped sets are equal — §D.3 command 11, the lane arm's skip set against the scrubbed arm's skip set of the same tree..." / "the two skipped sets are equal (command 11, the lane-versus-scrubbed pairing of the same tree)" | Same direction (skips must be equal, by command 11); only the pairing the citation names was corrected to match the clause — no requirement flipped. |
| P4 | AC-THE-004/AC-THE-008 expected reds + floor (I5-D3) | "expected: the six cli axes; for hook the thirteen axes"; "family size ≥ the floor 17"; "(expected: six cli axes, thirteen hook axes)" | "expected: the eight cli axes; for hook the ten axes"; "family size ≥ the floor 16"; "(expected: eight cli axes, ten hook axes)" | Same direction — under-scrubbed references still go red naming the axes, the floor is still a lower bound; only the counts were re-derived, none inverted. |
| P5 | spec.md R8 remainder (I5-D3) | "sets the nine family axes the measuring session exported and unsets the other eight" | "...unsets the other seven" | Same direction (arm = nine modelled axes set + the rest unset); remainder re-derived 17−9 → 16−9. |
| P6 | spec.md §E input (3) floor + red counts (I5-D3) | "fewer members than the floor recorded at c2 (**17**, §A.6)"; "the thirteen uncovered-axis names"; "the six uncovered-axis names only" | "(**16**, §A.6 — the census of the tree c2 is built on, re-derived at pre-flight and after every develop absorption)"; "the ten uncovered-axis names"; "the eight uncovered-axis names only" | Same direction (floor is a lower bound; the red enumerates uncovered axes); values re-derived, none inverted. The added absorption clause changes no direction — it names which event re-derives the floor. |
| P7 | acceptance.md §D.3 env recording + clause (f) + DoD (I5-D4) | "set `NAME=value …` (names sorted, a set-empty axis written `NAME=`)"; "the four lines are byte-identical" | same, plus "a secret-valued axis — §A.6's census marker — written `NAME=<redacted>`"; "the four lines are byte-identical **in their redacted form**" | Narrowing, disclosed: one axis's value (at HEAD `MOAI_FACTORY_APP_SERVER_TOKEN`) is no longer verbatim; the comparison still proves presence and the non-secret values, never the secret's value — the witness needs presence, not the secret's value. No requirement direction changed. |

Unobserved at this iteration (Gaps): the census was measured by the stated commands in this session, not by the guard at c2 (the guard re-reads `envkeys.go` at test time, so c2 re-measures on its own tree); the secret marker is a maintained census list, not mechanically derived (a future credential-class axis joins it in the change that adds the constant); the guard pair and the c1/c2/c2r commits still do not exist; no whole-package run (lane-local rule); the c1-stage command 11 baseline (a pre-existing env-gated skip, or its absence) is unobserved until M1.

Iteration-6 repair (v0.7.0, 2026-10-05): the four round-6 defects repaired per the leader's dispatch — I5-D4 residual/I6-D2 (the §A.6 redaction scope widened to every recording surface of this SPEC: plan.md §C step 2's recorded read output, the M1 c1 commit, the recorded lane-arm commands in §D.3; masking the value, keeping the variable name; set-empty written form stated), I6-D1 (R7's opening sentence narrowed to the initial-run-only case, matching its repaired tail and §D.3), I6-D3 (replay semantics for a redacted axis defined — the replay reproduces the recorded masked state, never the then-current session value — in §D.3's replay clause, AC-THE-003 clause (f) and plan.md M4 step 1), I6-D4 (the sorted c1 skip names files of all four arms and the c1-stage command 11 output added to the c1-commit enumerations in plan.md M1 c1 and §D.3's baseline-record paragraph; command 11 added to plan.md M4 step 1). `moai spec lint` clean, exit 0. No REQ or AC added.

Iteration-11 repair (v0.8.0, 2026-10-06, leader gate ruling — one pass): the six round-11 findings repaired. C1: the §A.6 masked form `NAME=<redacted>` declared the recording form (never a shell command to paste), the shell-executed quoted form `NAME='<redacted>'` stated beside it at every surface showing the form (§A.6, §D.3 arm-env line, env-line format, replay clause, clause (f), DoD list, plan.md §C step 2); the mask string and the byte-identity subject are unchanged. C2: the family census re-derived at HEAD `c5260970e` (2026-10-06) after the `26fbe130d` absorption of develop `985bd43da` — measured this session with §A.6's exact command (grep → file, `sort -u`, `wc -l`): **16**, the sorted list identical member-for-member with the §A.6 sixteen, and the `646a7860d` control still **17**; recorded as a new dated re-derivation in §A.6 and plan.md §C step 2; the standing `8cb2444e7` figures untouched. C3: `<BASE>` re-derived per plan.md B10 at HEAD `c5260970e` → `985bd43daa9494b9bda3c88fb4a5f3281cc1a81b`; E-5a/E-5a2/E-5b/E-5c re-run with that base print empty stdout exit 0 (measured this session); the literal `2de0a2cb6..HEAD` E-5c form prints 232 lines (develop commits, first `d21e4079ecf8…`) — dated re-derivation bullet added to the E-5 cell area; the dated rows untouched. C4: the §D.3 child-visible env witness now states the recorded line is extracted by the §C step 2 family filter then redacted, the raw full-environment dump is machine-local scratch under `.moai/state/verify/t1356/` never cited and deleted immediately after extraction. C5: the §D.3 repeat rule and R7 extended to command 11 — both arms repeated, repeat comparison between the two repeat files' skip names files, polarity unchanged (both-runs = hit, repeat-only = hit, initial-only = load noise). C6: progress §E.1 plan_status/audit-ready refreshed to the v0.8.0 iteration-11 state, spec.md §I report list through iter11, gaps item (2) widened to seven unmeasured cli axes (added `MOAI_FACTORY_MANAGED_TUI`, `MOAI_FACTORY_APP_SERVER_TOKEN`), plan.md §A measurement-tree note refreshed. Polarity rows below; `moai spec lint` clean exit 0 before commit; no REQ or AC added.

Iteration-12 repair (v0.8.1, 2026-10-06, leader-dispatched one-pass micro round): the iter12 findings dispositioned — codex P2 (the C5 tightening: a skip-row difference is not excused as load noise merely for vanishing on the repeat; §D.3 and R7, P13-P14), N1 (plan.md ledger reference E-1..E-9), N2 (the quoted execution form on the last two surfaces, P15), N3 (the §D.8 `<BASE>` dated pointer, P16), N5 (c1-stage repeat scope, P17), N6 (the P8..P12 rows attached to this table — this header), and N4 recorded as a disclosed design residual in §E.1 (the bullet after `gaps:`). `moai spec lint` exit 0 before commit. No REQ or AC added.

| # | Sentence (site) | Old (before the revision named in the row) | New (after it) | Polarity relation |
|---|-----------------|------|------|-------------------|
| P8 | §A.6 + §D.3 + plan §C step 2 masked form (C1) | "writes a secret-valued axis as `NAME=<redacted>` — the value masked, the variable name kept" (no execution form stated) | same, plus "the masked form is a recording form, never a shell command to paste; an execution writes it quoted `NAME='<redacted>'`; quotes are shell syntax, not part of the recorded value" | Unchanged direction — clarification. The mask string and every byte-identity subject stay `NAME=<redacted>`; the added sentence removes the paste/redirect hazard the audit named. |
| P9 | §A.6 + plan §C step 2 census pin (C2) | "the rule yields **16** constants at HEAD `8cb2444e7`" | same, plus the dated re-derivation record "the same command at HEAD `c5260970e` (2026-10-06, after the `26fbe130d` absorption) prints **16** again — identical member set" | Unchanged direction — added dated observation; the standing figures stay as measured history. |
| P10 | acceptance.md E-5 cell area (C3) | (no re-derived `<BASE>` recorded in the cell; header rule only) | dated re-derivation bullet: `<BASE>` = `985bd43da` at HEAD `c5260970e`, the four E-5 forms empty under it, literal base prints 232 lines for E-5c | Unchanged direction — the re-derivation rule existed; this records its execution without rewriting the dated rows. |
| P11 | acceptance.md §D.3 child-visible witness (C4) | "formatted the same way after the same redaction is applied to the file before recording" (dump scope and disposition unspecified) | "the recorded line is extracted by the §C step 2 family filter then passes the §A.6 redaction; the dump is the session's FULL environment — machine-local scratch under `.moai/state/verify/t1356/`, never a citation target, deleted immediately after extraction — only the filtered, redacted line reaches progress.md §E.2" | Narrowing, disclosed: what reaches the committed record is now explicitly family-filtered + redacted; the witness identity requirement (identical to the c1 line) is unchanged, and the redaction narrows nothing the witness needs. |
| P12 | §D.3 repeat rule + R7 (C5) | "A name printed by command 8 or 9 is not final on one run…Command 9 compares two arms, so both arms are repeated" | "command 8, 9 or 11…Command 11 compares two arms too, so it takes the same whole-arm repeat semantics as command 9 — both arms repeated, repeat comparison between the two repeat files' skip names files; same polarity" | Widened, disclosed: the established polarity (both-runs = hit, repeat-only = hit, initial-only = load noise) is applied unchanged to one more command; no polarity inverted. |
| P13 | acceptance.md §D.3 command 11 repeat tail (iter12 codex P2, C5 tightening) | "printed only in the initial run, load noise" — the excusal applied to a skip-row difference unconditionally | the both-runs/repeat-only polarity kept; the initial-run-only excusal for a skip-row difference now requires independent proof of a load or environment cause — without it the clause (b) failure stands | Narrowing, disclosed: the excusal now carries an evidence burden for skip rows only; the failing-name polarity of commands 8/9 (both-runs = hit, repeat-only = hit, initial-run-only may be load noise) is untouched. |
| P14 | spec.md R7 (iter12 codex P2, the mirror) | R7 applied the command 8/9/11 polarity with the initial-run-only excusal unconditionally, command 11 included | same polarity for failing names, plus one restriction binding command 11 only — a skip difference is not excused for vanishing on the repeat unless a load or environment cause is independently proven | Same direction as P13 — one rule mirrored at its second citation site; nothing inverted. |
| P15 | AC-THE-003 clause (f) + §D.5 DoD item (iter12 N2) | showed only the recording form `NAME=<redacted>` | same, plus the quoted execution form `NAME='<redacted>'` stated beside each (the pattern every other surface already carries) | Unchanged direction — clarification; the recording form and byte-identity subject stay `NAME=<redacted>`. |
| P16 | acceptance.md §D.8 `<BASE>` parenthetical (iter12 N3) | "(or the value re-derived and recorded per plan.md B10)" — no value findable (§E.2 still the run-phase placeholder) | same, plus the dated pointer: the 2026-10-06 re-derivation at HEAD `c5260970e` recorded `985bd43daa9494b9bda3c88fb4a5f3281cc1a81b` (E-5 re-derivation note, progress §E.1) | Unchanged direction — the B10 rule existed; its recorded execution is now findable from §D.8. |
| P17 | acceptance.md §D.3 repeat-trigger scope (iter12 N5) | unstated whether the both-arms repeat was required at the c1 stage (a literal reading forced whole-package repeats of both c1 arms) | explicit — the repeat requirement binds final-tree prints only; the c1-stage command 11 baseline record and command 9's five designed reds are baseline measurements, not repeat triggers | Unchanged direction — scope clarification; final-tree repeat obligations unchanged. |

## §E.2 Run-phase Evidence

### Run entry and authorization (2026-10-06)

- Accepted tree: HEAD `46b8eeb75` (branch `WT-test-env-hermetic-sweep`, clean at entry); run
  authorization = the leader's message of 2026-10-06 + the iteration-13 verdict
  `.moai/reports/t1356/plan-audit-iter13.md` (disposition (ii), receipt
  `rcpt-9f93b72c6eb4dbba07f004bf`) — the audit-cross decision record.
- First run-phase commit: `86aa3b1d6` `fix(SPEC-TEST-ENV-HERMETIC-001): draft -> in-progress on
  run entry (card t1356)` — spec.md frontmatter `status: draft` → `status: in-progress` only
  (Status Transition Ownership Matrix, manager-develop). It precedes the c1 commit so the c1
  file set stays exactly progress.md (§D.8 step 4); both are progress.md/spec.md-only commits
  and neither touches `internal/`, so the §D.8 step 6 enumeration is unaffected.
- `<BASE>` re-derived at run entry: `git merge-base develop HEAD` →
  `985bd43daa9494b9bda3c88fb4a5f3281cc1a81b` (the 2026-10-06 re-derivation, unchanged by the
  iter-12 commit); `git rev-list --count 985bd43da..HEAD` = **11** (≥ 1, §D.8 step 0 positive
  control).

### Pre-flight record (plan.md §C)

- **§C step 2 — the session family env read** (one read before any arm; this session models the
  lane arm). Command: `env | grep -E '^(MOAI_FACTORY_|MOAI_KANBAN|MOAI_AUTONOMY_TIER)'`, exit 0,
  output verbatim (nine axes; no secret-valued axis is present in this session, so nothing
  requires masking; `MOAI_FACTORY_APP_SERVER_TOKEN` is among the absent seven):

```
MOAI_KANBAN_ID=tm9i7y
MOAI_AUTONOMY_TIER=fully-autonomous
MOAI_FACTORY_WORKER=lane-23
MOAI_FACTORY_ROLE=lane
MOAI_FACTORY_WORKERS=0
MOAI_FACTORY_CLEAR_POLICY=
MOAI_FACTORY_AUTO_DISPATCH=auto
MOAI_KANBAN_BACKEND=glm
MOAI_KANBAN_SETTINGS_INJECTED=1
```

  Clause (f) minimum holds: `MOAI_FACTORY_ROLE=lane` present, `MOAI_FACTORY_WORKER=lane-23`,
  `MOAI_FACTORY_WORKERS=0` (non-empty string), `MOAI_KANBAN_ID=tm9i7y` non-empty. The lane arm
  models THIS session (`MOAI_KANBAN_BACKEND=glm`, `MOAI_FACTORY_WORKER=lane-23` — not the
  plan-time `claude`/`lane-6` values; spec.md R8: the arm models the measuring session).
- **Census sanity (§A.6)**: `grep -oE '"MOAI_(FACTORY_|KANBAN)[A-Z_]*"|"MOAI_AUTONOMY_TIER"'
  internal/config/envkeys.go` → sorted-unique file, `wc -l` = **16**, member set identical to the
  §A.6 sixteen (re-derivation at run entry; the family rule yields the same 16 as at
  `c5260970e`).
- **B2 conflict scan**:
  `grep -rn "Retired\|TestHarnessRetirement\|superseded" internal/cli internal/hook | head` —
  hits only in unrelated test fixtures and comments (a `plan_audit_d7_d8_test.go` harness
  retirement-conflict test, a CG-retirement error-symbol test, translation fixtures mentioning a
  superseded model id). No scrub-mechanism conflict. No conflicting retired policy found.
- **§C step 1 — the five observed reds re-established** on the accepted tree `46b8eeb75` (Go
  sources of the same lineage as the pinned measurement trees):
  - E-1 form: `env MOAI_FACTORY_ROLE=lane go test ./internal/cli -count=1 -v -run
    '^(TestTodoClaim_LaneGovernance|TestTodoClaimMCP_Mirror|TestTodoPickInFactoryProvenanceFailsOpenWithoutSpecOrGit)$'`
    — exit **1**, three `--- FAIL` with the ledger's messages (`todo_claim_test.go:291`,
    `todo_claim_test.go:416`, `todo_test.go:190`, each `refused — lane boundary`).
  - E-2 form (lane env with this session's values): exit **1**, two `--- FAIL`
    (`stale_run_m1_test.go:42` and `:85`, `staleRunNoticeFor = ""` both).
  - E-1b form: `unset MOAI_FACTORY_ROLE && go test ./internal/cli -count=1 -v -run …` — exit
    **0**, three `--- PASS`.
- **§C step 3 — hook child re-exec census** (B9's four `os.Args[0]` sites): `slot_lease_guard_test.go`
  (`exitedChildPID`) sets no `cmd.Env` (inherits); `session_start_drift_fill_burst_test.go`
  appends only its `helperDriftFillBurst*` vars to `os.Environ()`;
  `factory_handoff_race_test.go` and `factory_handoff_bind_test.go` append only
  `ownerHelperEnv=1`. **None carries a family axis as payload** — spec.md §D precondition (ii)
  holds; M3 needs no pin-marker pattern.
- **Builds**: `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- **Lint baseline**: `golangci-lint run --timeout=2m` → `0 issues.` (exit 0).
- **Windows vet baseline (pre-existing, not introduced by this run)**:
  `GOOS=windows GOARCH=amd64 go vet ./internal/cli ./internal/hook` exits **1**:
  `internal/cli/audit_gates_failclosed_test.go:159:3: undefined: installFakeCodex` and
  `:162:8: undefined: runAudit`. Recorded as the §C step 4 baseline; the M4 delta check is
  against this. Not this SPEC's defect and not repaired here (scope discipline); flagged to the
  leader in the run report.
- **§C step 6 — listed counts L0** (warm build cache, plain `go test -list '.*' P` redirected,
  then `grep -cE '^(Test|Example|Fuzz)'`): cli **5122** (exit 0; non-matching lines:
  `BenchmarkIsTrivialCommand`, `BenchmarkTodoAuditAnalyze`, the `ok` line), hook **1315** (exit
  0; non-matching lines: the five `Benchmark*` rows, the `ok` line). The plan-time figures
  (E-8: 4884 / 1322) moved with the develop absorptions, as E-8 anticipated; these are the
  pre-guard L0 of this tree.
- Guard files confirmed absent at entry (`ls` on both paths: No such file).

### Slot leases (per §C step 5, per-arm model)

- First acquisition: both `whole-package-test-suite` (leader-directed resource) and `heavy-test`
  (SPEC §D.3/§G resource) at `--max-duration 40m` — released unused; see the discarded attempt
  below.
- cli lane arm: both resources acquired at 80m (holder `0bb55ed0-b1f3-4665-816b-2c88a8678321`,
  until 14:29:30Z), released after the arm.
- cli scrubbed arm: both re-acquired at **84m** (until 15:31:41Z), `-timeout 82m` — the cap is
  1.5 × the longest runtime recorded so far (the cli lane arm's ~56m); released when the arm was
  killed (below).
- Resuming session (2026-10-07), per-arm model kept; cap **116m** / `-timeout 114m` throughout
  (= 1.5 × the survivor observation's 4649.376s ≈ 77.5m, the longest whole-package runtime
  recorded — see the cli scrubbed arm's post-seal observation above; moai build `2a4fd910c`,
  not an ancestor of tree HEAD `efc099b67`, §2.2 second-coordinate note):
  - cli scrubbed re-run: both resources acquired 2026-10-06T18:32:18/19Z (holder session
    `01e8bc01-2113-44d3-b6ae-48446f8bc46e`, until 20:28:18/19Z), released 19:37:07Z after the
    arm completed.
  - hook lane arm: both re-acquired 19:37:07Z (until 21:33:07Z), released 20:18:38Z after the
    arm completed.
  - hook scrubbed arm: both re-acquired 20:18:38Z (until 22:14:38Z), to be released after the
    arm completes.

### c1 arm records

**Discarded first attempt (cli lane arm)** — launched at 40m lease / `-timeout 38m` from a cold
build cache; measured rate ~110 tests/min (compilation included) projected a ~47m total, past
the 38m timeout. Killed by the session before any verdict (exit 144 after `pkill`), partial file
discarded (overwritten by the restart). Not a measurement; recorded for the lease-cap history.

**cli lane arm — COMPLETE, VALID.**

- Command (as executed, one compound invocation; the seven absent axes unset, the nine session
  axes set verbatim in the same compound per §D.3/B4):

```
unset MOAI_FACTORY_APP_SERVER_TOKEN MOAI_FACTORY_MANAGED MOAI_FACTORY_MANAGED_TUI MOAI_FACTORY_SLOW_LAUNCH_MS MOAI_KANBAN_CARD MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_LEAD_NAME && env MOAI_AUTONOMY_TIER=fully-autonomous MOAI_FACTORY_AUTO_DISPATCH=auto MOAI_FACTORY_CLEAR_POLICY= MOAI_FACTORY_ROLE=lane MOAI_FACTORY_WORKER=lane-23 MOAI_FACTORY_WORKERS=0 MOAI_KANBAN_BACKEND=glm MOAI_KANBAN_ID=tm9i7y MOAI_KANBAN_SETTINGS_INJECTED=1 go test ./internal/cli -count=1 -timeout 78m -json > .moai/state/verify/t1356/c1-cli-lane.json 2> .moai/state/verify/t1356/c1-cli-lane.err
```

- lane-arm env line (§D.3 recording form; names sorted; no secret-valued axis in this session's
  env, so no masked entry applies):

```
lane-arm env: set MOAI_AUTONOMY_TIER=fully-autonomous MOAI_FACTORY_AUTO_DISPATCH=auto MOAI_FACTORY_CLEAR_POLICY= MOAI_FACTORY_ROLE=lane MOAI_FACTORY_WORKER=lane-23 MOAI_FACTORY_WORKERS=0 MOAI_KANBAN_BACKEND=glm MOAI_KANBAN_ID=tm9i7y MOAI_KANBAN_SETTINGS_INJECTED=1 ; unset MOAI_FACTORY_APP_SERVER_TOKEN MOAI_FACTORY_MANAGED MOAI_FACTORY_MANAGED_TUI MOAI_FACTORY_SLOW_LAUNCH_MS MOAI_KANBAN_CARD MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_LEAD_NAME
```

- Exit code: **1** (`ARM-EXIT=1`). Window: 13:09:38Z → 14:05:48Z (**~56m runtime**; the recorded
  runtime the cap formula uses). Load average ~14-16 during the run.
- Validity (§D.3 commands 1/4/5): T = **5122** = L0 cli 5122 (command 1); invalid-arm markers
  (command 4, on both the JSON stream and the stderr capture) = **0 / 0**; exit non-zero with
  fail rows present (command 5) = **397** (≥ 1). **The arm is valid.**
- Failing rows: **397** (subtests included), **280** unique top-level tests. Dominated by the
  survey's todo-family nomination (177 of 280 unique names carry `todo`) plus factory-gate tests
  (`TestFactoryNext*`, `TestSD_AC0xx*`, `TestJevFinding*`, `TestTransitionStamps*`) — the
  pre-fix flip census this sweep exists to measure. The three cli observed reds are present
  (command 10 below).
- Skipped rows: **56** (the live-contract and credential-gated skips; compared cross-arm at the
  c1-stage command 11, pending the scrubbed arm).
- §D.3 command 10 (clause (f) positive control, on the sorted names file):

```
$ grep -cE '"Test":"(TestTodoClaim_LaneGovernance|TestTodoClaimMCP_Mirror|TestTodoPickInFactoryProvenanceFailsOpenWithoutSpecOrGit)"' c1-cli-lane.names.txt
3
```

  Expected 3, printed **3** — the cli observed reds are in the c1 lane arm.
- Sorted failing-name file (command 6 + 7, verbatim; machine-local source
  `.moai/state/verify/t1356/c1-cli-lane.names.txt`):

```
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestAdoptAppendsResumptionRecordWithoutTouchingPriorRecords"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestAdoptBriefsFromRecordedProgressAndEvidenceBeforeWork"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestAdoptDerivesSyncPhaseFromCloseMarker"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestAdoptRefusesCardNotPicked"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestAdoptRefusesWhenNothingRecorded"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestAutoHelpAndRefusalDoNotAssertPickOrder"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestCodexAudit_NonRequiredGateGoldenByteIdentical"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestCodexAudit_NonRequiredGateGoldenByteIdentical/corrupt-yaml"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestCodexAudit_NonRequiredGateGoldenByteIdentical/required-uppercase"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestCodexLaneChildEnvOmitsLabelMarker"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestCodexLaneEntryStartsRelaunchLane"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestCodexLaneLoopDefaultLaunchContinuesAfterFirstCard"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestDoneVerdict_CoexistsWithOperatorEvidence"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestDoneVerdict_NotPersistedWithoutFlag"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestDoneVerdict_PersistedWithRefAndTime"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestDoneVerdict_ReAdjudicationReDerivesSHA"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestDoneVerdict_RefusalArchivesNothing"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestDoneVerdict_RefusedWithoutRefThroughStoreAPI"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFDA_AuditDecideReadsTheQueueHold"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFDA_LaneAuditDecideAdmission"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFR_AC002_LegacyAbandonViaDecide"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFR_AC013_DecideAfterClockAdvance"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFR_AC014_AssignHints"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFR_AC015_DecideKickoffBatch"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFR_AC016_DecideQuestionChoices"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFR_AC017_DecideUnblock"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFR_AC018_DecidePushGate"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFR_AC018_PushGateNeverFetches"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFR_AC021_QueueSchemaUntouched"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFR_AC022_AssignRequiresQueuePicked"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFR_AC023_StatusIsReadOnly"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFR_AC024_CharacterizeGTDDispatch"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFR_AC024_GTDDispatchMirrorsFactoryRecord"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFR_AC025_GoalDispatchMirrorsFactoryRecord"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFactoryCardVerbsResolveLaneFromWorkerMarker"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFactoryEntryMatrix"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFactoryEntryMatrix/codex_lane"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFactoryNextAssignedSerialCardHoldsSlotInPickedArms"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFactoryNextAssignedSerialCardHoldsSlotInPickedArms/arm-b-recorded-picked-ownerless"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFactoryNextAssignedSerialCardHoldsSlotInPickedArms/arm-b2-queue-picked-unrecorded"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFactoryNextAssignedSerialCardStillHoldsSlotAgainstNewTakes"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFactoryNextDuplicateDispatchGuard"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFactoryNextExpiredLeaseReleasesSerialSlot"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFactoryNextExpiredLeaseReleasesSerialSlot/leased"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFactoryNextExpiredLeaseReleasesSerialSlot/plan"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFactoryNextExpiredLeaseReleasesSerialSlot/plan-audit"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFactoryNextExpiredLeaseReleasesSerialSlot/run"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFactoryNextExpiredLeaseReleasesSerialSlot/sync"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFactoryNextExpiredLeaseReleasesSerialSlot/sync-audit"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFactoryNextFailedSerialRowReleasesSlot"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFactoryNextFailedSerialRowReleasesSlot/blocked"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFactoryNextFailedSerialRowReleasesSlot/failed"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFactoryNextFailedSerialRowReleasesSlot/needs-decision"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFactoryNextLiveLeaseHoldsSerialSlotInEveryState"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFactoryNextLiveLeaseHoldsSerialSlotInEveryState/leased"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFactoryNextLiveLeaseHoldsSerialSlotInEveryState/plan"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFactoryNextLiveLeaseHoldsSerialSlotInEveryState/plan-audit"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFactoryNextLiveLeaseHoldsSerialSlotInEveryState/run"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFactoryNextLiveLeaseHoldsSerialSlotInEveryState/sync"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFactoryNextLiveLeaseHoldsSerialSlotInEveryState/sync-audit"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFactoryNextOwnAssignedSerialCardBlockedByLiveSerialLease"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFactoryNextOwnAssignedSerialCardBlockedByPickedSibling"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFactoryNextOwnAssignedSerialCardLeasesPastSiblingAssigned"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFactoryNextParallelizableConcurrentLeases"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFactoryNextParallelizableLeasesBesideLiveSerial"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFactoryNextParallelizableLeasesBesideLiveSerial/arm-a-assigned-parallelizable"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFactoryNextParallelizableLeasesBesideLiveSerial/arm-c-queued-parallelizable"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFactoryNextPickedOwnerlessRowHoldsSlot_OutOfExpiryScope"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFactoryNextRecordAndClaimRaceOnLeasedRow"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFactoryNextSerialMutualExclusivity"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFactoryNextSerialSlotLeaseExpiryBoundary"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFactoryNextSerialSlotLeaseExpiryBoundary/exactly-now"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFactoryNextSerialSlotLeaseExpiryBoundary/future"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFactoryNextSerialSlotLeaseExpiryBoundary/no-expiry-recorded"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFactoryNextSerialSlotLeaseExpiryBoundary/past"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFactoryNextSkipsClassificationBlocked"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFactoryStatusShowsHolderModePriority"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestGTDAllTodoVerbsParity"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestGTDFiveStageCLIUsesSameSQLite"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestGuardBypassMutant_ObserveHomePollution"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestHistoryArchivedRowAbsentStampsRenderDash"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestHistoryArchivedRowExposesTimeAxis"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestHistoryListingExposesTimeAxis"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestHistoryLiveRowExposesPickedAt"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestJevFinding_AdmissionOnly_ReSweepNeverCallsIt"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestJevFinding_PrecedenceHalfA_Suppressed"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestJevFinding_PrecedenceHalfA_Suppressed/agent"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestJevFinding_PrecedenceHalfA_Suppressed/mechanical"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestJevFinding_WritesNoCardField"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestJevFinding_WrittenAtAdmission"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestJevProbe_DisabledCapabilityProducesNothing"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestLandedGitCallsGuardEndOfOptions"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestLandedOnArchivedCardNamesTheArchive"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestLandedOnUnknownIDStillSaysNoBacklogItem"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestLandedOptionShapedRefIsRefused"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestLandingBackfillPathForAlreadyClosedCard"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestLandingEvidenceRoundTrip"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestLandingEvidenceSurvivesArchive_StoreLevel"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestLaneMarkerGoldenMatchesFLane"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestLaneMarkerGoldenMatchesFLane/codex"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestListJSONStampsOmitEmpty"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestLiveReadersUnchangedByHistoryVerb"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestMachineOnlyMarkAppearsAndClears"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestRelateAndUnrelateRefusals"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestSD_AC003_CodexRelaunchPerCard"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestSD_AC006_LaneCycleWithoutRemote"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestSD_AC008_NextSelectionOrderAndOutput"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestSD_AC008_NextSelectionOrderAndOutput/assigned-to-this-lane_wins_over_unowned_picked"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestSD_AC008_NextSelectionOrderAndOutput/card_assigned_to_another_lane_is_never_taken;_exit_3"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestSD_AC008_NextSelectionOrderAndOutput/oldest_queued_promoted_and_leased;_newer_stays_queued"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestSD_AC008_NextSelectionOrderAndOutput/two_lanes_concurrently:_exactly_one_leases_the_picked_card"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestSD_AC008_NextSelectionOrderAndOutput/unowned_picked_wins_over_older_queued"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestSD_AC010_MCPNextParentCheck"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestSD_AC010_NextRefusedOutsideParent"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestSD_AC011_CardWorktreeCreateReuseRefuse"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestSD_AC011_CardWorktreeCreateReuseRefuse/existing_directory_no_card_record_names_it:_refuse,_row_unchanged"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestSD_AC011_CardWorktreeCreateReuseRefuse/leased_card_with_no_recorded_worktree_gains_one_through_the_materializer"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestSD_AC011_CardWorktreeCreateReuseRefuse/recorded_worktree_is_reused_without_creating_another"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestSD_AC013_ClaudeCompleteViaIntegrationWorktree"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestSD_AC013_ClaudeCompleteViaIntegrationWorktree/caller-source_window:_refused_naming_--branch"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestSD_AC013_ClaudeCompleteViaIntegrationWorktree/complete_performs_the_merge_itself_and_records_the_re-measure_evidence"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestSD_AC013_ClaudeCompleteViaIntegrationWorktree/integration_branch_held_by_no_tree:_not_provisioned"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestSD_AC013_ClaudeCompleteViaIntegrationWorktree/integration_branch_held_only_by_the_parent_checkout:_not_provisioned"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestSD_AC013_ClaudeCompleteViaIntegrationWorktree/pre-merged_card_reaches_merged-local;_the_window_stays_held"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestSD_AC013_ClaudeCompleteViaIntegrationWorktree/window_naming_the_card's_own_branch:_refused"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestSD_AC014_MCPMatchesCLIWithProjectRoot"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestSD_AC014_MCPMatchesCLIWithProjectRoot/todo_list_renders_identically;_an_unusable_root_is_refused_identically"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestSD_AC015_LabelOnlyIsNotALane"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestSD_AC015_LaneQueueAllowlistWalk"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestSD_AC015_MCPTodoAddRefused"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestSD_AC016_LaneDecideRefused"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestSD_AC016_MCPDecideRefused"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestSD_AC018_ParentCheckoutUntouched"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestSD_AC020_ClearPolicies"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestSD_AC020_ClearPolicies/relaunch_continues_after_a_failed_child_session"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestSD_AC020_ClearPolicies/relaunch_supervising_loop_starts_one_session_per_card"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestSD_AC020_ClearPolicies/relaunch_without_the_claude_binary_is_refused"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestSD_AC023_CodexNextSkipsUnadvanceableCard"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestSD_AC024_CodexMergeRefusedComplete"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestSD_AC024_CodexMergeRefusedMCP"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestSD_AC024_CodexMergeRefusedStage"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestSD_AC025_IntegrationWindowSerializes"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestSemanticRelationsChangeNothing"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTempOriginGuidance_NamesRootsAndContinues"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTempOriginGuidance_SilentOnNonTemporaryBase"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTempOriginGuidance_SilentWithExplicitAbsoluteMOAIHome"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAddClassificationFile"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAddClassificationFileUnavailableFallsBack"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAddClassificationFileValidatesAndRecords"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAddClassificationFileValidatesAndRecords/decider_jev"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAddClassificationFileValidatesAndRecords/malformed_json"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAddClassificationFileValidatesAndRecords/mode_out_of_set"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAddClassificationFileValidatesAndRecords/priority_out_of_set"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAddClassificationFileValidatesAndRecords/valid"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAddDeciderFailureFallsBackWithNotice"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAddDeciderFailureFallsBackWithNotice/failure_promotes_the_fail-safe_default_with_one_notice"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAddDeciderFailureFallsBackWithNotice/positive_control:_healthy_decider_records_its_real_judgment"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAddDefaultDeciderPrintsNoNotice"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAddForce"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAddForceAdmitsAndRecords"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAddLeadingDashText"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAddNearDuplicateRecordsOnly"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAddPick"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAddPick_ConcurrentProcesses"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAddPick_OneLockedWrite"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAddRecordsClassificationInLockedWrite"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAddRefusesExactDuplicate"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAddReturnedIDAddressesALiveRow"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAddSortsQueueAndPrintsSortedPosition"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAdd_LLMFailureDegradesWithExactlyOneNotice"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAdd_LLMFailureDegradesWithExactlyOneNotice/append_path"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAdd_LLMFailureDegradesWithExactlyOneNotice/pick_path"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAdd_LLMSecretNeverPrinted"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAdd_LLMSecretNeverPrinted/failure_path"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAdd_LLMSecretNeverPrinted/happy_path"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAdd_LLMSelectionRecordsModelJudgment"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAdd_PrintsIDAndPosition"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAdd_SuppliedFileBeatsLLMSelection"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAdd_UnsetEnvBehavesPreSPEC"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAnalysisNeverReordersQueue"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAnalyzeRerunIsIdempotent"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAuditDropReasonSingleLine"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAuditMultilineRows"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAuditMultilineRows/next"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAuditPickDropped"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAutoDoneSkipsHeldCard"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAutoDone_CloseLineContract"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAutoDone_CloseSurfaceExclusivity"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAutoDone_CloseSurfaceExclusivity/todo_landed_transitions_nothing"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAutoDone_CollisionRecordedSHAOverrides"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAutoDone_CollisionSkipsAmbiguous"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAutoDone_DryRunByteIdentity"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAutoDone_ExecutionLog"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAutoDone_FalseNegativeShapes"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAutoDone_FetchBoundary"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAutoDone_FetchBoundary/with_--fetch_runs_exactly_one_fetch"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAutoDone_FetchBoundary/without_--fetch_runs_zero_fetches"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAutoDone_FormRecordedSHA"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAutoDone_FormSubjectAttribution"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAutoDone_Idempotence"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAutoDone_InconclusiveNeverCloses"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAutoDone_JSONOutput"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAutoDone_LiveFilter"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAutoDone_NegationAttributesNothing"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAutoDone_NoLandingColumnWrites"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAutoDone_RecordedSHAUnreachableSkips"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAutoDone_ReissuedIDOlderCommitSkips"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAutoDone_ReversalRow"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAutoDone_SpecNotCompletedSkips"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAutoDone_SpecUnreadableIsNotCompleted"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoAutoEntryPointFlag"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoBareFallthrough"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoBareInvocationLists"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoClaimMCP_Mirror"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoClaim_LaneGovernance"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoClaim_ListHistoryExposesLeaseColumns"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoClaim_NoCardExit"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoClaim_RefusesNonQueued"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoClaim_Renew"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoClaim_Success"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoComposedUpgrade_ForwardCompatibleFieldsSurvive"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoComposedUpgrade_FromLegacyV312Layout"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoConcurrentAdd_8Processes"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoDisclosure_LeavesBacklogJSONUntouched"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoDoneReclaimsFindings"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoDoneUndone_NeverPrompt"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoDoneUndone_RefusalsWriteNothing"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoDone_ArchivedRowsInvisibleToLiveReaders"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoDone_ArchivesRatherThanDiscards"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoDone_DisclosesLevel3"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoDone_ExpectRefusesMismatch"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoDone_MissReportedFileUntouched"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoDone_NoDisclosureAtLevel1"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoDone_NoLandingQueryWithoutTheFlag"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoDone_RemovesByID"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoDone_RequireLandedProceedsWhenInconclusive"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoDone_RequireLandedRefusesWhenNotLanded"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoDone_StdoutCarriesTheLandingVerdict"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoDone_StdoutCarriesTheLandingVerdict/landed"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoDone_StdoutCarriesTheLandingVerdict/no_query_at_all"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoDone_StdoutCarriesTheLandingVerdict/unanswerable"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoDone_UndoneRoundTripIsByteIdentical"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoDone_UnknownStillProceeds"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoDone_VerdictNamesRefAndDisclosesLevel2"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoDone_WithoutFlagNoRefNamed"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoDrop_AlreadyDropped_Refused"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoDrop_DroppedCardIsNotAPickCandidate"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoDrop_ExpectMatchAllowsTheDrop"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoDrop_MarksDroppedAndRecordsTheReason"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoDrop_RefusalsLeaveTheFileByteIdentical"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoDrop_UndropIsAnExactReversal"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoDrop_UndropRestoresAHandWrittenDroppedCard"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoEdit_EmptyText_RefusedNoWrite"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoEdit_ExpectMismatch_RefusedNoWrite"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoEdit_PrintsPriorTextSoTheEditIsReversible"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoEdit_RewritesTextPreservingIdentity"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoEdit_UnknownID_RefusedNoWrite"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoExactRefusalWorksWithoutAgent"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoExportJSON_CarriesArchiveAndDisclosesDowngrade"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoExportJSON_FailurePathsSurface"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoExportJSON_FailurePathsSurface/no_residue_is_left_when_the_write_fails"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoExportJSON_FailurePathsSurface/unwritable_queue_directory_is_reported,_not_swallowed"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoExportJSON_LeavesTheLiveStoreAuthoritative"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoExportJSON_NoDisclosureWithoutArchivedRows"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoExportJSON_NoTempResidue"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoExportJSON_RoundTripsThroughTheLegacyShape"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoExportJSON_SurvivesSubsequentVerbs"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoFindingLineSuggestsDroppingTheSubjectNotTheRow"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoFutureStateCardIsNeverSelectedByActionablePaths"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoHistoryDegradesWithoutArchiveTables"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoHistoryDegradesWithoutArchiveTables/dropped_archive_tables"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoHistoryDisclosesPreArchiveQueue"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoHistoryEmptyArchiveIsExplicit"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoHistoryLeavesStorageByteIdentical"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoHistoryLimitBound"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoHistoryListsNewestFirst"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoHistoryNormalizesBareOrdinal"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoHistoryRefusesNegativeLimit"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoHistoryReportsAbsentCard"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoHistoryReportsArchivedCard"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoHistoryReportsLiveCard"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoHistoryStatesWithheldCount"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoHold_ExpectMismatchRefuses"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoHold_MovesQueuedCardToHold"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoHold_RefusesNonQueuedStates"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoHold_RefusesUnknownState"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoLanded_ClearIsExclusive"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoLanded_NoSHADerivedFromTheGrep"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoLanded_OneRecordUnderTheLock"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoLanded_ReplaceAndClear"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoLanded_SHAValidation"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoLanded_SHAValidation/no_git_is_unrunnable_and_refuses"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoLanded_SHAValidation/reachable_accepts_and_stores_the_full_SHA"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoLanded_SHAValidation/refusing_branches_exit_1,_write_nothing,_and_are_distinguishable"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoLanded_SHAValidation/refusing_branches_exit_1,_write_nothing,_and_are_distinguishable/non-existent_object"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoLanded_SHAValidation/refusing_branches_exit_1,_write_nothing,_and_are_distinguishable/unreachable_commit"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoLanded_SHAValidation/refusing_branches_exit_1,_write_nothing,_and_are_distinguishable/unresolvable_ref"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoLanded_SHAValidation/the_card_id_reaches_neither_check"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoLanded_SpecStatusReadNeverInvented"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoLanded_StateCheckAndStatesUntouched"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoLanded_WholeQueueUnmoved"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoLegacyRecordRoundTrips"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoListRendersHeldCardTruthfully"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoList_AllDroppedDefaultView"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoList_DroppedFlagEmptySaysSo"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoList_DroppedFlagRendersDroppedOnly"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoList_DroppedHiddenByDefault"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoList_JSONKeepsDroppedCards"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoList_JSONStructured"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoList_LockFreeWhileForeignProcessHoldsLock"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoMachineLeaseSelectsOnlyQueued"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoMachineLeaseSelectsOnlyQueued/order-0"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoMachineLeaseSelectsOnlyQueued/order-1"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoMachineLeaseSelectsOnlyQueued/order-2"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoMachineLeaseSelectsOnlyQueued/order-3"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoMachineLeaseSelectsOnlyQueued/order-4"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoMove_BeforeAndAfter"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoMove_PositionFlagContract"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoMove_PreservesEveryItem"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoMove_TopAndBottom"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoMultiWordFallthroughAdds"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoNaturalLanguageCardsSurviveVerbGuard"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoNaturalLanguageCardsSurviveVerbGuard/fix_3_flaky_tests"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoNaturalLanguageCardsSurviveVerbGuard/fix_the_drift_found_in_t151"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoNaturalLanguageCardsSurviveVerbGuard/fix_the_flaky_gate"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoNaturalLanguageCardsSurviveVerbGuard/t151_regression_follow-up"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoNewVerbsAreHeadless"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoNewVerbsAreHeadless/add"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoNewVerbsAreHeadless/analyze"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoNewVerbsAreHeadless/relate"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoNewVerbsAreHeadless/unrelate"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoNextBare_ReadOnlyOldestFirst"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoNextPick_ConfirmationShowsText"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoNextPick_ExpectGuard"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoNextPick_OneLockedWrite"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoNext_OutOfRangeFileUntouched"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoNext_PickOnHeldCardRefused"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoPickInFactoryCapturesLinkedWorktreeSpecAndHEAD"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoPickInFactoryProvenanceFailsOpenWithoutSpecOrGit"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoQueue_WorktreeSeesPrimaryQueue"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoReadSurface_DisclosesNonAuthoritativeJSON"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoReadSurface_DisclosesNonAuthoritativeJSON/bare"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoReadSurface_DisclosesNonAuthoritativeJSON/history"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoReadSurface_DisclosesNonAuthoritativeJSON/list"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoReadSurface_DisclosesNonAuthoritativeJSON/pr"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoReadSurface_DisclosesNonAuthoritativeJSON/why"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoReadSurface_SilentWithoutJSON"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoReadSurface_SilentWithoutJSON/bare"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoReadSurface_SilentWithoutJSON/history"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoReadSurface_SilentWithoutJSON/list"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoReadSurface_SilentWithoutJSON/pr"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoReadSurface_SilentWithoutJSON/why"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoReadSurface_StdoutUnpolluted"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoReadSurface_StdoutUnpolluted/bare"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoReadSurface_StdoutUnpolluted/history"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoReadSurface_StdoutUnpolluted/list"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoReadSurface_StdoutUnpolluted/pr"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoReadSurface_StdoutUnpolluted/why"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoRelateAndUnrelateTouchNoCard"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoRelateCycleGuardShapes"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoRelateCycleGuardShapes/3-cycle_through_an_intermediate_card_is_refused"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoRelateCycleGuardShapes/blocks_spelling_closes_a_cycle_like_depends"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoRelateCycleGuardShapes/open_chain_into_a_new_card_is_allowed"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoRelateCycleGuardShapes/same-pair_opposite_spelling_encodes_no_cycle"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoRelateRefusesCycle"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoRelateSequencingKinds"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoShow"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoShow/absent_id_prints_absent_and_the_issued-mark_qualifier"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoShow/body_with_tabs_and_newlines_flattens_without_truncation"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoShow/live_line_carries_the_full_text"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoShowReadOnly"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoStaleStoreDisclosure_AddVerb"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoUndone_EmptiesTheArchiveEntry"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoUndone_ReissuedIDRefuses"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoUndone_RestoresTheCard"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoUndone_SurvivesMigrationFromLegacyJSON"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoUnhold_RefusesNonHeldStates"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoUnhold_ReturnsHeldCardToQueued"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoUnpick_RefusalsLeaveFileUntouched"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoUnpick_RevertsPickedToQueued"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoVerbExitCodesUnchanged"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoVerbGuardStillAddsNonAddresses"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoVerbGuardStillAddsNonAddresses/epic_7_planning"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoVerbGuardStillAddsNonAddresses/fix_3_flaky_tests"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoVerbGuardStillAddsNonAddresses/peek_T401"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoVerbGuardStillAddsNonAddresses/peek_card"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoVerbGuardStillAddsNonAddresses/peek_t401x"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoVerbGuardStillAddsNonAddresses/재현_t401"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoVerbsUnaffectedByFlag"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTodoWriteVerbs_CarryNoDisclosure"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTransitionStamps_ArchivePreservesStampsAndStampsArchivedAt"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTransitionStamps_DropThenDoneCarriesDroppedStampOnly"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTransitionStamps_DroppedAtStampsAndClears"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTransitionStamps_GtdEngagePickAndGoalMissionPick"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTransitionStamps_GtdEngagePickAndGoalMissionPick/gtd_engage_--pick"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTransitionStamps_NextPickStamps"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestTransitionStamps_PickedAtStampsAndClears"
```

- Sorted skip names file (command 11 extraction, verbatim; machine-local source
  `.moai/state/verify/t1356/c1-cli-lane.skipnames.txt`):

```
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestAuditPinLive_CodexPinConfirmation"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestAuditPinLive_GLMDifferential"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestAxisACanaryHomeSweep_TodoFamily"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestCleanupMoaiWorktrees"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestCodexAuditLaunchLiveContract"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestCodexAuditLaunchLiveReadOnlyRoles"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestCodexLive_ExplicitReadOnlyApprovalStall"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestCodexLive_OmittedSandboxPolicyBaseline"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestCodexLive_ReviewStartEmitsTurnStarted"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestCodexLive_SandboxPolicyStickiness"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestCodexLive_ThreadReuseAndTurnInterrupt"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestCodexRoleLiveLoadAndReadOnly"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestCrossCompileScriptMirrorsCanonical"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFactoryLeaseSerialCrossProcessHelper"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFactoryLiveCardFlowClaudeClaude"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFactoryLiveClaudeClaudeCompletionSeparation"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFactoryLiveHookBoundaryIdleTruth"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestGLMAudit_NoAskUserQuestion"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestGLMToolsEnable_AtomicWriteProtectsOriginal"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestGateLockHelperSleep"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestHeadroomInitSurfaceExport"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestHomeStateChangedSurfaceCoverageRunsBoundedFocusedSuite"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestLegacySkillIDsNotEmbedded/manifest_empty"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestLegacySkillIDsNotEmbedded/manifest_error"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestLiveClaudeAudit_GPTAndGLMOriginsUseSubscriptionReadOnly_AC_CLA_003_005_006_009_010_015"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestLiveCodexCompactFires"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestLiveCodexGoalContinueUntilMet"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestLiveCodexInterruptFires"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestLiveCodexNeedsInputOutcome"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestLiveCodexPermissionRequestFires"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestLiveCodexStopTimeoutCeiling"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestLiveHarnessIsolation"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestLiveHookFaultOutcome"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestLiveStopChainGoalContinuation"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestMCPServer_StdioRoundTripSubprocess"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestManagedCodexFactoryBrokerLive"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestManagedCodexFakeAppServer"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestManagedCodexTUIFakeCodex"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestManagedLoopbackChild"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestMigrationRollback_M001_Rejected"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestMigrationRollback_NoRollbackable"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestMigrationRollback_Succeeds"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestMigrationRun_AppliesPending"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestMigrationRun_NoPending"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestMigrationStatus_Human"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestMigrationStatus_JSON"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestPtyCaptureChild"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestPtyCapture_DowngradeConfirmButtonAlignment"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestPtyCapture_DowngradeConfirmLocalized"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestPtyCapture_FailWithoutTmux"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestPtyCapture_InitFirstScreen"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestPtyCapture_SkipWithoutGate"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestResolveRulesDir/windows_volume-letter_gate.yaml_value_passes_through_unjoined"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestSessionPIDStampExecHelper"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestVerifyRunHelperProcess"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestViewAcceptanceCriteria_ShapeTraceE2E"
```
**cli scrubbed arm — NOT COMPLETED (killed before any verdict; no measurement).**

- Command (as launched): the 16-axis scrubbed compound per §D.3 — `unset MOAI_AUTONOMY_TIER
  MOAI_FACTORY_APP_SERVER_TOKEN MOAI_FACTORY_AUTO_DISPATCH MOAI_FACTORY_CLEAR_POLICY
  MOAI_FACTORY_MANAGED MOAI_FACTORY_MANAGED_TUI MOAI_FACTORY_ROLE MOAI_FACTORY_SLOW_LAUNCH_MS
  MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS MOAI_KANBAN_BACKEND MOAI_KANBAN_CARD MOAI_KANBAN_ID
  MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_LEAD_NAME MOAI_KANBAN_SETTINGS_INJECTED && go test
  ./internal/cli -count=1 -timeout 82m -json > .moai/state/verify/t1356/c1-cli-scrubbed.json 2>
  .moai/state/verify/t1356/c1-cli-scrubbed.err` (one invocation; wrapped here for readability).
- Killed by the session (exit 144 after `pkill`) at 14:52:16Z with **1966/5122** top-level
  terminal rows and **4** fail rows partial. Two independent reasons, both recorded:
  1. **Projected timeout** — measured rate ~50 tests/min against a completion need of ~69 more
     minutes while the `-timeout 82m` bound (from lease start 14:07:41Z) expired at 15:29:40Z.
     The arm could not finish before its own bound.
  2. **Concurrent-load contamination** — at the kill decision the machine carried **three
     foreign `go test` processes against the same `./internal/cli` package** (two from card
     t1538: a `-run 'TestFR_|TestTodo|TestAutoDone|TestFactory'` selector run and a
     `-cover ./internal/cli/... ./internal/homestate/... ./internal/factory/...` run; one from
     card t1509: a template/cli review-scope run — `pgrep` evidence captured 14:41Z), with load
     average **31.15**. A measurement taken under that load is not a clean baseline (spec.md R7;
     `.claude/rules/local/gitflow-lane-protocol.md` §8's serialization intent).
- Disposition: the arm is **invalid-by-omission, never a pass** — the c1 record is explicitly
  partial (see Partial-record statement below), and the resuming session re-runs this arm as a
  whole-package measurement on a quiet machine before c2. The partial 4 fail rows are NOT
  extrapolated to any claim.

**cli scrubbed arm — COMPLETE, VALID (re-run by the resuming session 2026-10-07; supersedes the
discarded first attempt above, whose kill did not in fact land).**

- Post-seal observation (recorded before anything else): the first attempt's process survived the
  recorded `pkill` and wrote a COMPLETE stream into the old `c1-cli-scrubbed.json` — terminal
  rows `FAIL github.com/modu-ai/moai-adk/internal/cli 4649.376s` at 2026-10-07 00:25:22+09:00,
  top-level rows T=5122, 6 fail rows, 0 invalid-arm markers, stream start back-computed to
  2026-10-06 23:07:45+09:00 = the original launch. It is DISCARDED, never the arm record: it
  outlived its declared kill, ran lease-less after the seal released its lease, and straddled the
  recorded contention window (load up to 31.15). Preserved as
  `.moai/state/verify/t1356/c1-cli-scrubbed-survivor.json` (machine-local scratch, never cited
  as an arm record; full disposition note in scratch `arm-survivor-note.md`). Its ONE recorded
  use: the runtime figure 4649.376s ≈ 77.5m feeds the D6 cap formula as the longest
  whole-package runtime ⇒ cap = 1.5 × 77.5m ≈ **116m**, `-timeout 114m` (the lease below was
  sized from the scratch-recorded survivor observation before this extension was committed).
- Command (as executed, one compound invocation — the 16-axis scrubbed compound per §D.3):

```
unset MOAI_AUTONOMY_TIER MOAI_FACTORY_APP_SERVER_TOKEN MOAI_FACTORY_AUTO_DISPATCH MOAI_FACTORY_CLEAR_POLICY MOAI_FACTORY_MANAGED MOAI_FACTORY_MANAGED_TUI MOAI_FACTORY_ROLE MOAI_FACTORY_SLOW_LAUNCH_MS MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS MOAI_KANBAN_BACKEND MOAI_KANBAN_CARD MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_LEAD_NAME MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/cli -count=1 -timeout 114m -json > .moai/state/verify/t1356/c1-cli-scrubbed.json 2> .moai/state/verify/t1356/c1-cli-scrubbed.err
```

- Exit code: **1** (`ARM-EXIT=1`). Window: 2026-10-06T18:32:19Z → 2026-10-06T19:35:40Z
  (**3767.007s ≈ 62.8m runtime**). Quiet-machine precondition at launch: load average
  **8.99-10.40**, foreign `go test` `pgrep` empty at 18:27Z (5 minutes pre-launch); mid-run load
  was not sampled — recorded as a measurement gap, mitigated by the arm's own validity signals
  (T=L, 0 invalid markers) and the 62.8m runtime consistent with an uncontended run.
- Validity (§D.3 commands 1/4/5): T = **5122** = L0 cli 5122 (command 1); invalid-arm markers
  (command 4, on both the JSON stream and the stderr capture) = **0 / 0**; exit non-zero with
  fail rows present (command 5) = **3** (≥ 1). **The arm is valid.**
- Failing rows: **3** — one test, `TestCodexAudit_NonRequiredGateGoldenByteIdentical`, with
  subtests `/corrupt-yaml` and `/required-uppercase`. The same test is present in the c1 cli
  lane names file, so it is a both-arms failure candidate for clause (c), not a lane-only red.
- §D.3 command 10 (clause (f), on the sorted scrubbed names file): expected **0**, printed
  **0** — the five observed reds are absent from the scrubbed arm.
- Sorted failing-name file (commands 6+7, verbatim; machine-local source
  `.moai/state/verify/t1356/c1-cli-scrubbed.names.txt`):

```
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestCodexAudit_NonRequiredGateGoldenByteIdentical"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestCodexAudit_NonRequiredGateGoldenByteIdentical/corrupt-yaml"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestCodexAudit_NonRequiredGateGoldenByteIdentical/required-uppercase"
```

- Skipped rows: **56** (`c1-cli-scrubbed.skipnames.txt`); the cross-arm comparison is the
  c1-stage command 11, run after the hook arms land.

**hook lane arm — COMPLETE, VALID (run 2026-10-07 by the resuming session; the recorded c1
cli-lane env line replayed verbatim per clause (f), keeping the four lane-arm env lines
byte-identical).**

- Command (as executed, one compound invocation — the c1 cli-lane env replayed: the seven
  absent axes unset, the nine session axes set to the c1 recorded values
  `MOAI_FACTORY_WORKER=lane-23` and `MOAI_KANBAN_ID=tm9i7y`, NOT the resuming session's own
  `lane-5`/`tmhxo0` values — AC-THE-003(f)'s four-line byte-identity requirement decides this:
  the c1 line is the reference every later lane arm replays, per §D.3 "the final arm replays
  the c1 line"):

```
unset MOAI_FACTORY_APP_SERVER_TOKEN MOAI_FACTORY_MANAGED MOAI_FACTORY_MANAGED_TUI MOAI_FACTORY_SLOW_LAUNCH_MS MOAI_KANBAN_CARD MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_LEAD_NAME && env MOAI_AUTONOMY_TIER=fully-autonomous MOAI_FACTORY_AUTO_DISPATCH=auto MOAI_FACTORY_CLEAR_POLICY= MOAI_FACTORY_ROLE=lane MOAI_FACTORY_WORKER=lane-23 MOAI_FACTORY_WORKERS=0 MOAI_KANBAN_BACKEND=glm MOAI_KANBAN_ID=tm9i7y MOAI_KANBAN_SETTINGS_INJECTED=1 go test ./internal/hook -count=1 -timeout 114m -json > .moai/state/verify/t1356/c1-hook-lane.json 2> .moai/state/verify/t1356/c1-hook-lane.err
```

- lane-arm env line (§D.3 recording form; byte-identical to the c1 cli-lane line above):

```
lane-arm env: set MOAI_AUTONOMY_TIER=fully-autonomous MOAI_FACTORY_AUTO_DISPATCH=auto MOAI_FACTORY_CLEAR_POLICY= MOAI_FACTORY_ROLE=lane MOAI_FACTORY_WORKER=lane-23 MOAI_FACTORY_WORKERS=0 MOAI_KANBAN_BACKEND=glm MOAI_KANBAN_ID=tm9i7y MOAI_KANBAN_SETTINGS_INJECTED=1 ; unset MOAI_FACTORY_APP_SERVER_TOKEN MOAI_FACTORY_MANAGED MOAI_FACTORY_MANAGED_TUI MOAI_FACTORY_SLOW_LAUNCH_MS MOAI_KANBAN_CARD MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_LEAD_NAME
```

- Exit code: **1** (`ARM-EXIT=1`). Window: 2026-10-06T20:08:54Z → 2026-10-06T20:17:38Z
  (**487.874s ≈ 8.1m runtime**). Quiet-machine precondition at launch: no `*.test` binaries
  and no real `go test` invocations by `pgrep`; load average 19.05-20.66 — decaying residue of
  foreign suites that ended minutes before; the recorded cli-lane precedent accepted load
  14-16. Mid-run load not sampled (same gap note as the cli scrubbed arm).
- Validity (§D.3 commands 1/4/5): T = **1315** = L0 hook 1315 (command 1); invalid-arm markers
  (command 4, on both the JSON stream and the stderr capture) = **0 / 0**; exit non-zero with
  fail rows present (command 5) = **3** (≥ 1). **The arm is valid.**
- Failing rows: **3** — the two observed hook reds **plus one additional lane-only name**,
  `TestStaleRunNoticeFactoryLegacyLabel` (not one of the five observed reds; a measured flip
  this sweep exists to record — the hook scrubbed arm decides whether it is lane-only, an M3
  fix target, or a both-arms failure candidate for clause (c)).
- §D.3 command 10 (clause (f), on the sorted hook lane names file): expected **2**, printed
  **2** — the hook observed reds are present.
- Sorted failing-name file (commands 6+7, verbatim; machine-local source
  `.moai/state/verify/t1356/c1-hook-lane.names.txt`):

```
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/hook","Test":"TestStaleRunNoticeFactoryLegacyLabel"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/hook","Test":"TestStaleRunNoticeLegacyLeaderSpelling"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/hook","Test":"TestStaleRunNoticeLegacySessionRecord"
```

- Skipped rows: **6** (`c1-hook-lane.skipnames.txt`); the cross-arm comparison is the c1-stage
  command 11, run after the hook scrubbed arm lands.
- Lease lines: both resources acquired at 116m cap at 2026-10-06T19:37:07Z (holder session
  `01e8bc01-2113-44d3-b6ae-48446f8bc46e`, until 21:33:07Z), held across this arm, released at
  20:18:38Z and immediately re-acquired (until 22:14:38Z) for the hook scrubbed arm; moai
  build `2a4fd910c`, not an ancestor of tree HEAD `efc099b67` (§2.2 second-coordinate note).
  Load-collision note: two foreign `cli.test` binaries (other sessions' suites) were running
  at this arm's launch — a DIFFERENT package from this arm, load below the recorded incident
  band (31); the arm's own validity signals carried the verdict.

**hook scrubbed arm — COMPLETE, VALID (run 2026-10-07 by the resuming session; exit 0 — the
arm is fully green).**

- Command (as executed, one compound invocation — the 16-axis scrubbed compound per §D.3):

```
unset MOAI_AUTONOMY_TIER MOAI_FACTORY_APP_SERVER_TOKEN MOAI_FACTORY_AUTO_DISPATCH MOAI_FACTORY_CLEAR_POLICY MOAI_FACTORY_MANAGED MOAI_FACTORY_MANAGED_TUI MOAI_FACTORY_ROLE MOAI_FACTORY_SLOW_LAUNCH_MS MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS MOAI_KANBAN_BACKEND MOAI_KANBAN_CARD MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_LEAD_NAME MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/hook -count=1 -timeout 114m -json > .moai/state/verify/t1356/c1-hook-scrubbed.json 2> .moai/state/verify/t1356/c1-hook-scrubbed.err
```

- Exit code: **0** (`ARM-EXIT=0`; terminal rows `ok github.com/modu-ai/moai-adk/internal/hook
  436.166s`). Window: 2026-10-06T20:19:26Z → 2026-10-06T20:26:43Z (**436.166s ≈ 7.3m
  runtime**). Launch conditions as the hook lane arm (no test binaries by `pgrep`, load
  16.82 decaying, two foreign `cli.test` suites of a DIFFERENT package running; mid-run load
  not sampled).
- Validity (§D.3 commands 1/4/5): T = **1315** = L0 hook 1315 (command 1); invalid-arm markers
  (command 4, JSON + stderr) = **0 / 0**; exit zero — command 5 not applicable (it binds
  non-zero exits only). **The arm is valid.**
- Failing rows: **0** — the names file is empty (`grep` exits 1 with an empty file, a valid
  empty set per §D.3 once commands 1 and 4 hold; they do).
- §D.3 command 10 (on the sorted scrubbed names file): expected **0**, printed **0** — the
  hook observed reds are absent.
- Skipped rows: **6** — sorted skip-names file (commands 11 extraction + sort, verbatim;
  machine-local source `.moai/state/verify/t1356/c1-hook-scrubbed.skipnames.txt`):

```
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/hook","Test":"TestConsumerOnly_M0AndMxByteUnchanged"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/hook","Test":"TestDriftFillBurstHelper"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/hook","Test":"TestFactoryHookBenchmarkBudget"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/hook","Test":"TestScanWriteContent_SupportedExtension_ScannerNotAvailable"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/hook","Test":"TestSessionStart_DriftCacheProbe"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/hook","Test":"TestSessionStart_HandleStageProbe"
```

- Lease lines: both resources acquired 2026-10-06T20:18:38Z (holder session
  `01e8bc01-2113-44d3-b6ae-48446f8bc46e`, until 22:14:38Z), released 20:27:47Z after the arm
  completed (see the Slot leases section above for the resuming session's full lease history).

### c1-stage commands 9/10/11 — COMPLETE (the c1 baseline comparisons)

- **Command 9, hook** (clause (b) red state — expected to print at c1): `LC_ALL=C comm -3
  c1-hook-lane.names.txt c1-hook-scrubbed.names.txt` prints exactly **3** rows, all lane-only
  (verbatim):

```
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/hook","Test":"TestStaleRunNoticeFactoryLegacyLabel"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/hook","Test":"TestStaleRunNoticeLegacyLeaderSpelling"
"Action":"fail","Package":"github.com/modu-ai/moai-adk/internal/hook","Test":"TestStaleRunNoticeLegacySessionRecord"
```

  The two observed reds are the designed content. The third row,
  `TestStaleRunNoticeFactoryLegacyLabel`, is a **newly measured lane-only flip** — absent from
  the scrubbed arm — and is therefore an M3 fix target alongside the two observed reds (it
  reads the same stale-run-notice family axes; the c1 lane names file is the witness it fails
  only under the lane env).
- **Command 9, cli** (red state): `LC_ALL=C comm -3 c1-cli-lane.names.txt
  c1-cli-scrubbed.names.txt` prints **394** rows (grep/comm exit 1). Every printed line is a
  row of the c1 cli lane names file recorded above (the 397-row file minus the three rows the
  scrubbed arm shares — `TestCodexAudit_NonRequiredGateGoldenByteIdentical` and its two
  subtests, recorded verbatim in the cli scrubbed arm block); both input files are carried in
  this section, so the output is reconstructible byte-for-byte. The five observed cli reds are
  among the printed rows (command 10 witnesses them at **3** on the lane names file, recorded
  in the cli lane arm block).
- **Command 10** (clause (f) positive control, c1 arms only): cli lane names file printed
  **3** (expected 3 — recorded in the cli lane arm block); hook lane names file printed **2**
  (expected 2 — recorded in the hook lane arm block); cli scrubbed names file printed **0**
  (expected 0 — recorded in the cli scrubbed arm block); hook scrubbed names file printed
  **0** (expected 0 — recorded in the hook scrubbed arm block above).
- **Command 11** (skip-set equality, cross-arm pairing lane-vs-scrubbed of the same tree):
  - cli: `LC_ALL=C comm -3 c1-cli-lane.skipnames.txt c1-cli-scrubbed.skipnames.txt` prints
    **nothing** — the skip sets are equal (56 rows = 56 rows). The cli scrubbed skip-names
    file (the lane-side file is recorded in the cli lane arm block above), sorted, verbatim:

```
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestAuditPinLive_CodexPinConfirmation"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestAuditPinLive_GLMDifferential"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestAxisACanaryHomeSweep_TodoFamily"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestCleanupMoaiWorktrees"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestCodexAuditLaunchLiveContract"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestCodexAuditLaunchLiveReadOnlyRoles"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestCodexLive_ExplicitReadOnlyApprovalStall"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestCodexLive_OmittedSandboxPolicyBaseline"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestCodexLive_ReviewStartEmitsTurnStarted"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestCodexLive_SandboxPolicyStickiness"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestCodexLive_ThreadReuseAndTurnInterrupt"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestCodexRoleLiveLoadAndReadOnly"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestCrossCompileScriptMirrorsCanonical"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFactoryLeaseSerialCrossProcessHelper"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFactoryLiveCardFlowClaudeClaude"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFactoryLiveClaudeClaudeCompletionSeparation"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestFactoryLiveHookBoundaryIdleTruth"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestGLMAudit_NoAskUserQuestion"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestGLMToolsEnable_AtomicWriteProtectsOriginal"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestGateLockHelperSleep"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestHeadroomInitSurfaceExport"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestHomeStateChangedSurfaceCoverageRunsBoundedFocusedSuite"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestLegacySkillIDsNotEmbedded/manifest_empty"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestLegacySkillIDsNotEmbedded/manifest_error"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestLiveClaudeAudit_GPTAndGLMOriginsUseSubscriptionReadOnly_AC_CLA_003_005_006_009_010_015"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestLiveCodexCompactFires"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestLiveCodexGoalContinueUntilMet"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestLiveCodexInterruptFires"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestLiveCodexNeedsInputOutcome"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestLiveCodexPermissionRequestFires"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestLiveCodexStopTimeoutCeiling"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestLiveHarnessIsolation"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestLiveHookFaultOutcome"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestLiveStopChainGoalContinuation"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestMCPServer_StdioRoundTripSubprocess"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestManagedCodexFactoryBrokerLive"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestManagedCodexFakeAppServer"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestManagedCodexTUIFakeCodex"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestManagedLoopbackChild"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestMigrationRollback_M001_Rejected"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestMigrationRollback_NoRollbackable"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestMigrationRollback_Succeeds"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestMigrationRun_AppliesPending"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestMigrationRun_NoPending"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestMigrationStatus_Human"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestMigrationStatus_JSON"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestPtyCaptureChild"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestPtyCapture_DowngradeConfirmButtonAlignment"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestPtyCapture_DowngradeConfirmLocalized"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestPtyCapture_FailWithoutTmux"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestPtyCapture_InitFirstScreen"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestPtyCapture_SkipWithoutGate"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestResolveRulesDir/windows_volume-letter_gate.yaml_value_passes_through_unjoined"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestSessionPIDStampExecHelper"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestVerifyRunHelperProcess"
"Action":"skip","Package":"github.com/modu-ai/moai-adk/internal/cli","Test":"TestViewAcceptanceCriteria_ShapeTraceE2E"
```

  - hook: `LC_ALL=C comm -3 c1-hook-lane.skipnames.txt c1-hook-scrubbed.skipnames.txt` prints
    **nothing** — the skip sets are equal (6 rows = 6 rows). The fenced file above is the
    scrubbed file verbatim (source `c1-hook-scrubbed.skipnames.txt`); the lane file
    `c1-hook-lane.skipnames.txt` is set-identical — the empty `comm -3` itself is the witness.
  - **BC-1 dispositions owed: none.** No skip-row difference exists in either package to
    adjudicate — the c1 baseline is clean on the skip axis.

**c1 completion statement (2026-10-07, resuming session — supersedes the seal's partial-record
statement, which the seal commit `efc099b67` preserves verbatim in history).** All four §D.3
c1 arms are now measured and valid — cli lane (previous session), cli scrubbed re-run, hook
lane, hook scrubbed (this session; commands, validity signals and lease lines above) — and the
c1-stage commands 9/10/11 are complete (the section above; command 11 prints nothing in both
packages, so the skip sets are equal and **no BC-1 disposition is owed**). Baseline summary for
the fix milestones: cli lane 397 failing rows / 280 unique vs scrubbed 3 rows / 1 unique (the
394-row lane-only diff is the flip census this sweep exists to fix); hook lane 3 rows vs
scrubbed 0 — the two observed reds plus one newly measured lane-only flip,
`TestStaleRunNoticeFactoryLegacyLabel`, which is an M3 fix target alongside the observed reds.
This extension is a progress.md-only commit: the §D.8 step 1 subject grep keeps finding exactly
one line, step 4's c1 shape (progress.md only) is preserved, and the c1 → c2 ancestry holds
from the seal commit.

### c2 guard-red record (the designed red state, tree `ef18813fc`)

- c2 commit `ef18813fc` lands EXACTLY the two persisted guard drafts, byte-identical to
  `.moai/state/verify/t1356/factory_env_axes_test.go` and `lane_env_axes_test.go` (`cmp` clean
  both), gofmt-clean, with both package test binaries compiling (`go test -c -o /dev/null`,
  exit 0 both). The hook scrub list `laneEnvScrubAxes` is declared EMPTY at c2 (D5) with no
  `TestMain` call yet — the guards are designed red on this tree.
- The eight guard reds — four guard tests × two env forms, every run exit **1**, each a single
  `go test -run '^<name>$' -count=1 -v` invocation on the c2 tree (`ef18813fc`). The verbatim
  `-v` verdict lines:

  Form A — plain-scrubbed (the 16-axis unset compound of §D.3):

```
--- FAIL: TestFactoryEnvAxesCovered (0.12s)
--- FAIL: TestFactoryEnvAxesScrubApplied (0.28s)
--- FAIL: TestLaneEnvAxesCovered (0.03s)
--- FAIL: TestLaneEnvAxesScrubApplied (0.09s)
```

  Form B — explicit-lane (the recorded c1 lane env replayed verbatim):

```
--- FAIL: TestFactoryEnvAxesCovered (0.19s)
--- FAIL: TestFactoryEnvAxesScrubApplied (0.41s)
--- FAIL: TestLaneEnvAxesCovered (0.03s)
--- FAIL: TestLaneEnvAxesScrubApplied (0.12s)
```

  Both forms red: the guards' redness is form-independent by construction (the coverage guard
  is static analysis over referenced-vs-scrubbed axes; the applied guard composes its own
  sentinel child env) and the eight runs confirm it on the c2 tree.
- Why each guard is red (representative `t.Errorf` lines, scrubbed form, verbatim):
  - TestFactoryEnvAxesCovered — referenced axes uncovered: `family axis MOAI_AUTONOMY_TIER is
    referenced by production (codex_sync_gate.go, factory_launch_helpers.go) but is in neither
    the TestMain scrub set nor the exemption table`
  - TestFactoryEnvAxesScrubApplied — sentinel survives: `family axis MOAI_AUTONOMY_TIER is
    still present in the child after TestMain: the start-up scrub is declared but not applied`
  - TestLaneEnvAxesCovered — `declared scrub set is empty: TestMain would strip no family axis`
  - TestLaneEnvAxesScrubApplied — `family axis MOAI_FACTORY_AUTO_DISPATCH is still present in
    the child after TestMain: the start-up scrub is declared but not applied`
- Machine-local run logs (never a citation target): `/tmp/t1356-c2r-<pkg>-<test>-<form>.txt`,
  eight files; the verdict lines above are the committed carrier.

### M2/M3 fix record (tree `63b4e77e4`)

**M2 — commit `dd2303570`** (`fix(SPEC-TEST-ENV-HERMETIC-001): M2 extend the cli TestMain
scrub set with the eight uncovered family axes (card t1356)`): `factoryAmbientEnvKeys` gains
`config.EnvFactoryRole` plus the seven other uncovered axes. Verified (all exit 0):

- AC-THE-001 narrow lane command (`env MOAI_FACTORY_ROLE=lane go test ./internal/cli -count=1
  -v -run '^(TestTodoClaim_LaneGovernance|TestTodoClaimMCP_Mirror|TestTodoPickInFactoryProvenanceFailsOpenWithoutSpecOrGit)$'`):
  **exactly 3 `--- PASS`** (`TestTodoClaim_LaneGovernance (0.68s)`, `TestTodoClaimMCP_Mirror
  (0.31s)`, `TestTodoPickInFactoryProvenanceFailsOpenWithoutSpecOrGit (0.99s)`), **0 `--- SKIP`**.
- Both cli guard tests PASS: `--- PASS: TestFactoryEnvAxesCovered (0.04s)`,
  `--- PASS: TestFactoryEnvAxesScrubApplied (0.11s)`.
- Existing pin/clear tests PASS: `--- PASS: TestFactoryAmbientEnvClearedInTestMain (0.00s)`,
  `--- PASS: TestFactoryEnvPinnedSkipsTestMainClear (0.00s)`.
- gofmt clean (empty `gofmt -l`).

**M3 single-axis arms (measured BEFORE choosing the set, on tree `dd2303570`, per plan.md M3
step 1; each arm = the recorded c1 lane env minus the named axis, both StaleRunNotice tests,
exit codes as shown):**

| arm (lane env minus one axis) | exit | verdict |
|---|---|---|
| unset `MOAI_KANBAN_ID` | **0** | both `--- PASS` — responsible axis |
| unset `MOAI_FACTORY_WORKERS` | **0** | both `--- PASS` — responsible axis |
| unset `MOAI_FACTORY_ROLE` | 1 | both `--- FAIL` — not responsible |
| unset `MOAI_FACTORY_WORKER` | 1 | both `--- FAIL` — not responsible |
| unset `MOAI_KANBAN_BACKEND` | 1 | both `--- FAIL` — not responsible |

  Matches the plan-phase result: the stale-run gate reads the `MOAI_KANBAN_ID` ∧
  `MOAI_FACTORY_WORKERS` conjunction. Machine-local arm logs:
  `/tmp/t1356-m3-arm-unset-<axis>.txt`.

**M3 — commit `63b4e77e4`** (`fix(SPEC-TEST-ENV-HERMETIC-001): M3 fill the hook scrub set
with the ten referenced family axes and wire TestMain (card t1356)`): `laneEnvScrubAxes`
filled with the ten production-referenced axes; `hook TestMain` calls `scrubLaneEnvAxes()`
beside the existing `CLAUDE_PROJECT_DIR` scrub. Precondition (i) held from the c1 hook
scrubbed arm (all sixteen axes unset, exit 0, zero fail rows). Exemption table stays EMPTY
(no measurement shows a test needing an ambient family value). Verified (all exit 0):

- E-2 / AC-THE-002 lane command: **exactly 2 `--- PASS`** (`TestStaleRunNoticeLegacyLeaderSpelling
  (0.00s)`, `TestStaleRunNoticeLegacySessionRecord (0.00s)`), **0 `--- SKIP`**.
- Both hook guard tests PASS: `--- PASS: TestLaneEnvAxesCovered (0.04s)`,
  `--- PASS: TestLaneEnvAxesScrubApplied (0.08s)`.
- The third measured flip green under the c1 lane env replay:
  `--- PASS: TestStaleRunNoticeFactoryLegacyLabel (0.61s)`.
- AC-THE-002 assertion-diff clause: `git diff 985bd43d… -- internal/hook/stale_run_m1_test.go`
  prints nothing.
- gofmt clean (empty `gofmt -l` on both files).

**L1 listings on the post-M3 tree** (`go test -list '.*'`, `grep -cE '^(Test|Example|Fuzz)'`):
cli **5124** = L0 5122 + 2 guard tests; hook **1317** = L0 1315 + 2 guard tests (both ≥ L0 + 2).

**§D.8 step 6 enumeration at `63b4e77e4`** (re-evaluated at the M4 tip after the final
record lands): `git rev-list --no-merges --reverse 985bd43d… -- internal/` lists exactly
`ef18813fc…` (c2, the excluded commit), `dd2303570…` (M2), `63b4e77e4…` (M3); M2 and M3 each
descend from c2r `9048623ec` (`merge-base --is-ancestor` exit 0, `rev-list --count` ≥ 1).

**Mutation probes P1-P5 (run on tree `63b4e77e4`, each mutation reverted and the revert
verified by an empty `git diff` on the touched code paths).** Run BEFORE the final pairs —
the machine was under sustained foreign load (up to 107) that blocks the whole-package arms,
the probes are load-tolerant sub-second guard runs, and they will be re-affirmed if any
post-pair change (a contingent pin) lands; none is expected since M2/M3 fixed every measured
flip:

| probe | mutation | verdict (verbatim) | naming line |
|---|---|---|---|
| P1 | `config.EnvFactoryRole` removed from `factoryAmbientEnvKeys` | `--- FAIL: TestFactoryEnvAxesCovered (0.12s)` | `family axis MOAI_FACTORY_ROLE is referenced by production (codex_factory.go, codex_launcher.go, factory.go, factory_card.go)` |
| P2 cli | `clearFactoryAmbientEnv()` disabled in cli `TestMain` | `--- FAIL: TestFactoryEnvAxesScrubApplied (0.74s)` | 16 `still present in the child` lines (every family axis survives) |
| P2 hook | `scrubLaneEnvAxes()` disabled in hook `TestMain` | `--- FAIL: TestLaneEnvAxesScrubApplied (0.22s)` | 10 `still present in the child` lines — the ten referenced hook axes, AC-THE-008's prediction |
| P3 | padded exemption row (reason `n/a`; citation `update_disk_backup_test.go`, first verified to not mention the axis) | `--- FAIL: TestFactoryEnvAxesCovered (0.24s)` | `exemption row for MOAI_FACTORY_ROLE cites update_disk_backup_test.go, which does not reference the axis` |
| P4 | reference scan pointed at `t.TempDir()` (an empty directory) | `--- FAIL: TestFactoryEnvAxesCovered (0.00s)` | `reference scan found zero production references: the scan has silently stopped matching (moved files, renamed constants)` |
| P5 | hook guard file moved aside and restored | `--- FAIL: TestFactoryEnvAxesCovered (0.24s)` | `sibling guard file ../hook/lane_env_axes_test.go is missing: open ../hook/lane_env_axes_test.go: no such file or directory` |

  Probe logs `/tmp/t1356-p1.txt`, `p2-cli.txt`, `p2-hook.txt`, `p3.txt`, `p4.txt`, `p5.txt`
  (machine-local, never a citation target). After P5's restore the code tree was verified
  byte-identical to HEAD `63b4e77e4` (`git status --short` names only the in-progress
  progress.md record; `git diff` on code paths empty).

### M4 final pairs — hook pair, first measurement (tree `63b4e77e4`)

**final hook lane arm — COMPLETE, VALID, GREEN (first measurement).**

- Command (as executed — the c1 hook lane env replayed verbatim, `-timeout 114m`):

```
unset MOAI_FACTORY_APP_SERVER_TOKEN MOAI_FACTORY_MANAGED MOAI_FACTORY_MANAGED_TUI MOAI_FACTORY_SLOW_LAUNCH_MS MOAI_KANBAN_CARD MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_LEAD_NAME && env MOAI_AUTONOMY_TIER=fully-autonomous MOAI_FACTORY_AUTO_DISPATCH=auto MOAI_FACTORY_CLEAR_POLICY= MOAI_FACTORY_ROLE=lane MOAI_FACTORY_WORKER=lane-23 MOAI_FACTORY_WORKERS=0 MOAI_KANBAN_BACKEND=glm MOAI_KANBAN_ID=tm9i7y MOAI_KANBAN_SETTINGS_INJECTED=1 go test ./internal/hook -count=1 -timeout 114m -json > .moai/state/verify/t1356/final-hook-lane.json 2> .moai/state/verify/t1356/final-hook-lane.err
```

- Exit code: **0** (`ARM-EXIT=0`; terminal `ok github.com/modu-ai/moai-adk/internal/hook
  506.147s`). Window: 2026-10-06T23:56:17Z → 2026-10-07T00:06:44Z (**506.147s ≈ 8.4m**).
  Launch: stable quiet window (load 12.46, zero `*.test` binaries — the watcher's
  two-consecutive-poll condition), leases re-acquired at 23:56:17Z (holder session
  `01e8bc01-2113-44d3-b6ae-48446f8bc46e`, until 01:52:16/17Z; §2.2 note: moai build
  `2a4fd910c`, not an ancestor of tree HEAD).
- Validity (§D.3 commands 1/4/5): T = **1317** = L1 hook (command 1); invalid-arm markers
  (command 4) = **0 / 0**; exit zero — command 5 not applicable. **The arm is valid.**
- Failing rows: **0** — the fix holds at whole-package scale under the full lane env: the c1
  lane-only flips (the two observed reds plus `TestStaleRunNoticeFactoryLegacyLabel`) are
  gone.
- Clause (f): the lane-arm env line replays the c1 recorded line byte for byte; the
  child-visible env witness (`unset <the sixteen family axes> && env <the nine lane axes>
  env`, the arm command's prefix with `env` in place of the test run) printed exactly the
  nine replayed axes through the §C step 2 family filter — byte-identical to the c1
  recorded env line, no unset axis present — and the full-env dump was deleted immediately
  after extraction per §D.3.
- Skipped rows: **6** (`final-hook-lane.skipnames.txt`).

**final hook scrubbed arm — COMPLETE, VALID (first measurement), 1 failing row.**

- Command (as executed — the 16-axis scrubbed compound, `-timeout 114m`):

```
unset MOAI_AUTONOMY_TIER MOAI_FACTORY_APP_SERVER_TOKEN MOAI_FACTORY_AUTO_DISPATCH MOAI_FACTORY_CLEAR_POLICY MOAI_FACTORY_MANAGED MOAI_FACTORY_MANAGED_TUI MOAI_FACTORY_ROLE MOAI_FACTORY_SLOW_LAUNCH_MS MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS MOAI_KANBAN_BACKEND MOAI_KANBAN_CARD MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_LEAD_NAME MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/hook -count=1 -timeout 114m -json > .moai/state/verify/t1356/final-hook-scrubbed.json 2> .moai/state/verify/t1356/final-hook-scrubbed.err
```

- Exit code: **1** (`ARM-EXIT=1`; terminal `FAIL github.com/modu-ai/moai-adk/internal/hook
  786.863s`). Window: 2026-10-07T00:07:18Z → 2026-10-07T00:20:36Z (**786.863s ≈ 13.1m**);
  same lease window, launch conditions as the lane arm.
- Validity (§D.3 commands 1/4/5): T = **1317** = L1 hook; invalid-arm markers = **0 / 0**;
  exit non-zero with **1** fail row (command 5 ≥ 1). **The arm is valid.**
- Failing row: **`TestFactoryHookContextAndContinuationSafety`** — scrubbed-ONLY (the lane
  arm passes it). Static read of the failure block (`factory_messages_test.go:706`, "first
  stop continuation"): the test expects `DecisionBlock` from the Stop handler after seeding
  the ledger; the Stop gate reads the ambient lane axes, and with them unset the gate does
  not fire and `Decision` stays empty — an env-deterministic, non-load-shaped failure shape
  (the same stale-run gate family as the two observed reds, here asserted from the positive
  side). Machine-local run log `.moai/state/verify/t1356/final-hook-scrubbed.json`.
- §D.3 command 9 (final, first measurement): prints exactly this one name — **clause (b)
  candidate, pending the repeat protocol below**. Command 11 (final skips): prints nothing.
  The labelled no-new-skip extra vs c1: prints nothing (both packages of the hook pair).

**Repeat protocol (owed before any clause (b) verdict on this name).** Both arms repeat —
the identical commands on the same tree — on a stable quiet window: under load a repeat
could falsely convict a load-flaky test, so the repeats carry the same quiet-machine
discipline as the arms themselves (the two-consecutive-poll watcher condition). If the
scrubbed repeat re-prints the name (printed in both runs), clause (b) fails and the M4
step-2 contingent fix applies — a `t.Setenv` pin of the axes the gate reads makes the
verdict arm-independent, an exemption row only if a measurement shows the test truly needs
the ambient value — followed by re-measuring the pair on the post-pin tree (the final-tree
measurements must be on the tree the AC judges). If the name prints only in the initial
run, it is load noise, recorded with both outputs.

### Discovery narrow pair re-record (plan.md M1 c1; both arms COMPLETE)

- Discovery lane arm (the compound lane form of this session) — exit **0**: 14 `--- PASS`,
  1 `--- SKIP: TestHelperLeaderChild (0.00s)` (`factory_discovery_live_test.go:91: helper only`),
  `ok … internal/discovery 2.404s`.
- Discovery scrubbed arm (all 16 family axes unset, one compound) — exit **0**: the same 14
  `--- PASS` and the same one `--- SKIP`, `ok … internal/discovery 1.877s`.
- **Equal** — same pass set, same pre-existing helper skip, no flip. Matches plan-time E-7.

### Guard-draft persistence note (leader drain order item ③)

The two guard drafts prepared for c2 are persisted at
`.moai/state/verify/t1356/factory_env_axes_test.go` and
`.moai/state/verify/t1356/lane_env_axes_test.go` (machine-local scratch, NEVER committed, never
cited as evidence; `/tmp` copies may not survive the operator reboot). Design points frozen in
the drafts: family read from `../config/envkeys.go` at test time (regex verified against the
real file in a scratch module — exactly 16 axes, the §A.6 identifiers); cli scrub set =
`factoryAmbientEnvKeys` ∩ family; hook scrub set `laneEnvScrubAxes` declared EMPTY at c2 (D5) —
the draft omits the scrub function at c2 (no unused symbol) and adds it at M3 with the `TestMain`
call; applied-behaviour probes re-execute `os.Args[0]` with all 16 axes at a sentinel value
(`t1356-axes-sentinel`) plus a witness var (`MOAI_CLI_TEST_FACTORY_AXES_PROBE` /
`MOAI_HOOK_TEST_LANE_AXES_PROBE`), child bound 300s (never below the 20s precedent; the measured
child runtime is recorded at M4), case-insensitive axis stripping for Windows; sibling-guard
stale signal reads `../hook/lane_env_axes_test.go` / `../cli/factory_env_axes_test.go` for
`func <name>(`; no quoted family literal appears in the hook file (the
`envkeys_factory_role_test.go:60` rule); gofmt clean.

### RUN-PHASE SEAL (leader drain order 2026-10-06, operator reboot preparation)

- **Freeze point**: c1 commit `39364e133` (`test(SPEC-TEST-ENV-HERMETIC-001): baseline record
  (card t1356)`, progress.md only). Chain so far: `86aa3b1d6` (draft → in-progress) →
  `39364e133` (c1). Branch `WT-test-env-hermetic-sweep`; NO push (remote freeze holds; the
  leader batch-pushes). Slot leases `whole-package-test-suite` and `heavy-test`: **released** —
  the resuming session acquires fresh per arm.
- **Complete at freeze**: pre-flight in full (including the five observed reds re-established,
  L0 cli 5122 / hook 1315, census 16, hook child census clean, lint 0 issues, pre-existing
  Windows-vet failure recorded as the §C baseline); discovery narrow pair re-recorded equal;
  the cli lane c1 arm VALID (T=L=5122, 397 fail rows / 280 unique, command 10 = 3); guard
  drafts designed + persisted.
- **NOT complete at freeze** (the resuming session, in order, before c2): ① re-run the c1 cli
  scrubbed arm + run the hook lane and scrubbed arms (whole-package, per-arm leases, quiet
  machine — the 2026-10-06 contention of three foreign cli suites at load 31 is recorded above
  and must NOT be reproduced); ② extract names/skip files and run c1-stage commands 9 (five
  designed reds), 10 (hook = 2), 11 (skip-equality baseline); ③ extend §E.2 in a
  progress.md-only commit (the c1 subject grep of §D.8 step 1 must keep finding exactly one
  line); ④ then c2 (the two guard files, EXACTLY those two, from the persisted drafts) → run
  the guard reds on the c2 tree (plain-scrubbed and explicit-lane forms, 8 runs) → c2r
  (progress.md only) → M2 (cli fix) → M3 (hook fix) → M4 (final pairs, mutation probes P1-P5,
  AC re-evaluations). BC-1/BC-2/BC-3 stay binding (BC-1: any command-11/skip difference is a
  clause-(b) failure unless the load/environment cause is independently proven; family-axis
  differences are never excusable).
- **Scratch persistence**: `/tmp` may not survive the reboot — everything load-bearing was
  persisted under `.moai/state/verify/t1356/` (arm JSON + err, names/skip files, the two guard
  drafts, this §E.2's assembly parts). `/tmp` remnants (`/tmp/t1356-guards/`,
  `/tmp/t1356-regex-check/`, `/tmp/t1356-L0-*.txt`, `/tmp/t1356-census-*.txt`) are disposable.
  The scratch dir stays machine-local: never committed, never cited; delete after §E.4 lands.
- **Ceiling**: nothing after c1 was started this session — no c2 files, no fix commits, no
  mutation probes. The new session takes the run from the resume list above.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — owned by manager-develop>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — owned by manager-docs>_
