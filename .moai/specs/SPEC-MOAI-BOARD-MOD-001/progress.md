# Progress — SPEC-MOAI-BOARD-MOD-001

## §E.1 Plan-phase Audit-Ready Signal

- Tier M audit ceiling exceeded with leader approval (2026-10-02): plan-audit ran twice (iter1 FAIL 0.79, iter2 FAIL 0.79) and the leader approved one extra delta iteration limited to the iter2 defects (revision 0.3.0)
- Tier: M (`plan.md` §A) — files 14 (13 under `mods/moai-board/` + `.gitignore`), REQ 15/16, AC 15/16 (13 blocking + 2 manual Gap-class; 15 headings, 15 matrix rows, 15 distinct ids by the repo counter's regex — `grep -o -h -E 'AC-([A-Z0-9]+-)*[0-9]+[a-z]?' acceptance.md | sort -u | wc -l` printed `15`; the earlier 18 came from the a/b sub-ids, now folded; `.moai/reports/t338/ac-count-baseline.txt` has no row for this SPEC and was not edited)
- Artifacts: `spec.md`, `plan.md`, `acceptance.md`, `decision-index.md` (decision gate on), `progress.md`
- Card: t1436 (Class C), branch `WT-moai-board-mod`, tree `802a72235536958ada5b7cd5876a168e4b8c325f`; evidence path for the lane's verdict: `.moai/reports/t1436/verdict.md`
- Open decisions: Q1-Q7 in `decision-index.md`, none with a preferred answer (Q7 = must the mod load under the operator profile, whose rollout switch refuses hooks modules); Escalations: none of the leader-first list is required; the operator-profile refusal (known external issue, leader reports it, not a SPEC blocker) and two manual items are listed in `spec.md` §5
- Revision 0.2.0 (plan-audit iter1 FAIL 0.79 repair, must-fix only): D1 pure/engine split — pure criteria closed under `bun test` (developer-local, lane evidence only), the engine parts of AC-MBM-004, -005, -009 and -012 UNOBSERVED while `claude plugin test` is refused (superseded in 0.3.0 by the temp-config runner); D2 SPEC-id rule stated once; D3 "action" defined; D4 `dispatch:` test + `buildPickArgv(` call-site check; D5 label "Picked"; D7 M-10 (274 bytes) and M-11 re-measured
- Revision 0.3.0 (plan-audit iter2 FAIL 0.79 delta repair): N1 every `$.…` call in `hooks/register.tsx`, helpers `$`-free (validate refuses `$` across an import, M-17); N2 AC sub-IDs folded to 15; N3 AC-MBM-003 (iii) pins occurrences of the bare `buildPickArgv`, false aliasing cover removed; N4 pure recipe adds `<skipped` = 0 (scratch control); N6 frontmatter version 0.3.0; N5/N8 engine output format observed, `plugin.json` `types` pointer; engine runner works under `CLAUDE_CONFIG_DIR=<empty temp dir>` (M-16) so the engine parts have real RED-now cells and no UNOBSERVED label
- Plan-phase self-verification (this run, tree `802a72235`): SPEC-ID pattern check printed `PASS`; `ls .moai/specs | grep -c MOAI-BOARD` printed `0` before authoring; `moai spec lint SPEC-MOAI-BOARD-MOD-001` and `moai spec lint --strict SPEC-MOAI-BOARD-MOD-001` both printed `0 error(s), 0 warning(s)`, exit 0, at 0.1.0 as "No findings" and at 0.2.0 with one INFO `OwnershipTransitionUnmeasured` (the plan commit 83ea1f975 has no `Authored-By-Agent` trailer; not a SPEC defect). An earlier lint run reported 15 `CoverageIncomplete` warnings, fixed by naming `REQ-MBM-nnn` in the AC matrix. REQ-to-AC traceability re-checked at 0.2.0 by `grep`/`diff`: 15 REQ ids = 15 ids on AC matrix rows, none missing either way
- Not verified at plan-phase: no mod file exists yet — RED-now cells in `acceptance.md` §D (validate exit 1; engine runner under the temp config dir exit 1 "no such plugin folder"; bun exit 1); the operator profile still refuses `claude plugin test` (`spec.md` M-13, G-11); interactive behaviour, the live pick, tsc and the in-engine parse CPU time remain Gaps (`spec.md` §7)
- status: draft — plan-audit PASS-WITH-DEBT 0.86 (delta, `.moai/reports/t1436/plan-audit-iter3-delta.md`, audited sha 347bb1c94); spec artifacts unchanged since that sha (`git diff --stat 347bb1c94 HEAD -- .moai/specs/SPEC-MOAI-BOARD-MOD-001` printed nothing); optional debt P1-P4 carried into run-phase

### Kickoff decision record (2026-10-02)

decision record: decided_by=lane-12+orchestrator evidence_refs=.moai/reports/t1436/plan-audit-iter3-delta.md(PASS-WITH-DEBT 0.86, nine must-pass criteria PASS, no blocking defect),.moai/reports/t1436/plan-audit-iter2.md,.moai/reports/t1436/plan-audit-iter1.md,leader-dispatch(delta iteration approved 10-02; resume instruction 10-02) ladder_path=gate-row:plan→run Kickoff AUTONOMOUS §9.1

- Verdict class: PASS-WITH-DEBT is read as a passing verdict (score 0.86 over the Tier M threshold 0.80, no must-pass failure); a strict "PASS only" reading would hold the gate — the leader has been told the class (message of the delta report) and resumed the lane without objection
- Debt carried: P1 (drop the pipe from AC-003 (iii-a)), P2 (exclude `*.md` from the AC-003 (i) grep), P3 (promote the namespace-import mutant into AC-003), P4 (date-stamp the operator-profile refusal wording; re-measure before any report to the operator — the auditor saw the profile accept `claude plugin test` twice); the run phase may fix these as AC-text hygiene only
- Progression mode: autonomous; no goal armed (arm-only, nothing to judge until the mod exists)

## §E.2 Run-phase Evidence

Run phase by manager-develop (cycle tdd), card t1436, branch `WT-moai-board-mod`. Code state = commit `e323eba4f`; later commits touch only `acceptance.md` text. Builds: `claude 2.1.287`, `moai v3.2.0-rc.25` (802a72235, an ancestor of HEAD; no Go code changed since, so the installed build judges `moai spec lint` correctly), `bun 1.4.2`. Every figure below was observed in this run against this tree. Pre-flight (plan §C) reproduced: C1 `2.1.287 (Claude Code)`; C7 exit 1; C8 exit 1 `no such plugin folder`; C9 `ls mods` exit 1; C10 exit 1 "did not match any test files".

### Milestone to commit

| M | Commit | Content |
|---|---|---|
| M1 | `0fa6b34bc` | state contract, argv table, pick helpers, manifest, skeleton; `draft -> in-progress` on this first run-phase commit |
| M2 | `efa1633c5` | queue and lanes data path, fail-soft classification, polling gate and loop |
| M3 | `114c8cef3` | Pane UI (Queue, Lanes), confirmed pick, fail-soft hooks, board tests |
| M4 | `a925f7236` | SPEC tab: list, guards, chunker, file wiring |
| M5 | `e323eba4f` | `.gitignore` rule, README, three MX notes |
| hygiene | `c0cc0915c` | acceptance.md text P1, P2, P4 (no criterion changes meaning) |

### Red then green (TDD)

- M1 pure: `bun test mods/moai-board/tests/pure/ ...` before `data.ts` existed: exit 1, `error: Cannot find module '../../hooks/data'`; after: 3 pass.
- M2 pure: the 24-test spec run against the M1-committed `data.ts` (scratch copy): exit 1, `SyntaxError: Export named 'clampInterval' not found`; after: 24 pass. M2 engine against the M1 skeleton: exit 1, 1 pass (`poll: no process while the pane is closed`, a negative guard that is green on the skeleton) and 3 fail (`Expected: >= 10 / Received: 0` and kin); after: 4 pass.
- M3 engine: exit 1, 4 pass and 9 fail (`press: no Button of moai-board keyed "refresh" is drawn`); after: 13 pass.
- M4 pure: exit 1, `Cannot find module '../../hooks/specs'`; after: 35 pass in total. The 2 `spec-tab:` engine tests were written together with the wiring (no separate red).
- Mutants (scratch copies, each killed by the named test): confirm check removed -> `pick: cancel runs no process` and `pick: other text runs no process` fail; timer not stopped on close -> `poll: timer cancelled on ui.close` fails; pick argv run from `session.start` -> `dispatch:` fails; real-path check removed -> `specs-guard: real path outside base denied` and `spec-tab: a file resolving outside the folder is not read` fail; id rule `^SPEC-` -> `specs: list rows...`, `specs: default filter`, `specs-guard: id pattern` fail; closing fence not appended -> `specs-chunk: fences closed and reopened` fails.

### AC to evidence (final state, this tree)

| AC | Result | Command and observed output |
|---|---|---|
| AC-MBM-001 | PASS | `claude plugin validate mods/moai-board` exit 0, `hooks: session.start, command.run{command=moai-board}, ui.close, session.end, ui.render{component=Pane, requestId=moai-board}`; `grep -rn "name: '[^']*:" mods/moai-board/hooks` exit 1 |
| AC-MBM-002 | PASS | `calls: $.clock.every (via startPolling), $.command.register, $.fs.read (via openSpecFile), $.fs.stat (via openSpecFile), $.process.run (via runMoai), $.session.root (via rootOf), $.state.get (...), $.state.set (...), $.ui.ask (via onPickPress), $.ui.close (via closePane), $.ui.open, $.ui.resolve, $.ui.toast (via onPickPress)`; forbidden-count pipeline `0`, control `grep -c -F '$.process.run'` `1` |
| AC-MBM-003 | PASS | (i) one line `mods/moai-board/hooks/register.tsx:46`; (ii) one line `hooks/data.ts:57`; (iii-a) `grep -rno 'buildPickArgv' mods/moai-board/hooks` 3 lines (data.ts:56, register.tsx:11, register.tsx:165); (iii-b) `grep -c 'onPickPress' hooks/register.tsx` `2`; (iv) over data.ts, specs.ts, view.tsx exit 1. Residual as stated in the AC: occurrence pins, not control flow; the behavioural check is `dispatch:` |
| AC-MBM-004 | PASS pure + engine | pure `pick-pure:` 5, engine `^(pass) pick:` 4 |
| AC-MBM-005 | PASS pure + engine | pure `poll-pure:` 2; engine `poll:` 3, `dispatch:` 1 |
| AC-MBM-006 | PASS pure | `queue:` 4 (incl. the 2 MB payload, >= 1,900,000 chars) |
| AC-MBM-007 | PASS pure | `queue-fallback:` 4, `queue-text:` 1 |
| AC-MBM-008 | PASS pure | `lanes:` 4 |
| AC-MBM-009 | PASS pure + engine | `failsoft-pure:` 4; engine `failsoft:` 2 |
| AC-MBM-010 | PASS pure | `specs:` 3 |
| AC-MBM-011 | PASS pure | `specs-guard:` 4, `specs-chunk:` 4 |
| AC-MBM-012 | PASS engine | `board:` 3 (terminal and desktop inside each test) |
| AC-MBM-013 | PASS | (a) `git check-ignore -v mods/moai-board/.claude-plugin/types/claude-code/index.d.ts` exit 0 `.gitignore:447:mods/*/.claude-plugin/types/`; (b) register.tsx and `types/index.d.ts` exit 1; the 14 paths of plan B.1 plus README exit 1; control `node_modules/x` exit 0 (`.gitignore:365`); (c) `git ls-files -- 'internal/template/templates/*moai-board*'` empty, control zone-registry path printed; (d) `grep -n mods .goreleaser.yml` exit 1 |
| AC-MBM-014 | UNOBSERVED | manual, needs an interactive terminal; Gap below |
| AC-MBM-015 | UNOBSERVED | manual live pick, needs a non-lane session and an operator-admitted throwaway card; not attempted |

Runner evidence: pure `bun test mods/moai-board/tests/pure/ --reporter=junit --reporter-outfile=/tmp/mbm-junit.xml` exit 0, 35 pass, `<failure` 0, `<skipped` 0 (pure, developer-local evidence only, spec.md G-13). Engine `CLAUDE_CONFIG_DIR=/tmp/mbm-claude-cfg claude plugin test mods/moai-board > /tmp/mbm-engine.out 2>&1` exit 0, 15 pass, `^(fail)` 0, run under the empty temp config dir (it stayed empty: `ls -la` total 0). Engine parts of AC-004, -005, -009, -012 were observed on an executing run (acceptance.md §A.4, §F item 2).

Other checks: `moai spec lint SPEC-MOAI-BOARD-MOD-001 --strict` exit 0, `No findings`; `go test ./internal/template/ -run 'Gitignore|NamespaceProtection|TemplateNeutrality|InternalContentLeak'` exit 0; `go test ./internal/spec/ -run 'ACCounter|Lint'` exit 0; `grep` over `.github/workflows`, `.golangci.yml`, `lefthook.yml`, `Makefile` for `mods` or `**/*.ts(x)` printed nothing. Plan files untouched apart from the owned edits: spec.md `status` and acceptance.md text P1, P2, P4.

### Deviations and decisions to read

- A `close` button (key `close`, hotkey `x`) is drawn in the pane header. The kit cannot raise `ui.close`, and a plugin's own `$.ui.close` does not reach its own `ui.close` hook (measured), so `poll: timer cancelled on ui.close` presses this button; the handler stops the timer then calls the engine's close. The `ui.close` hook (person's close mark, Esc) also stops the timer but cannot be raised by the kit (Gap). Closing the pane is the mod's own view state, not a fourth action; the sync auditor may judge otherwise.
- State keys: the plan listed view, queue, lanes, specs, notice; a sixth key `doc` holds the SPEC file chunks (<= 12 x 9,000 chars), cleared on back (G-8 unmeasured).
- The pane is opened only by `/moai-board` (nothing opens it at `session.start`); the poll timer starts in that command after the pane opens, so closing and reopening restarts it.
- `at` of a feed is the last time the data changed: an unchanged poll writes nothing (REQ-MBM-007), so the age shown beside a failure reads "last updated Xm ago" from that moment; no age-only staleness dimming exists (dimming happens only while the latest read failed).
- A user action (tab switch, refresh) that arrives during a poll is queued and served right after it (still one in flight); timer ticks that find a poll running are dropped.
- Text elements take no `key` in the engine; every line a test must find sits in a keyed `Box`.
- P3 (promote the namespace-import check into AC-MBM-003) was not applied: it adds a check, which is not a text-only edit. Left for the sync audit with N13, N14, D6, D9, D10.
- `mod` pure tests judged under bun are not CI evidence; no CI job exists for the mod (out of scope).

### Gaps (unobserved, not passes)

- G-1 interactive behaviour: pane paint, docking, scrolling (63 queue rows may not scroll), hotkey delivery, `Markdown` painting; AC-MBM-014. To close: an operator runs `claude --plugin-dir <absolute path to mods/moai-board>` in an interactive terminal and types `/moai-board`, then records date and build here.
- G-2 / AC-MBM-015: live pick; needs a non-lane session and one throwaway card the operator admitted; the confirmation dialog's real look and the CLI's real success output are unobserved. A lane cannot run it (queue mutation is refused, M-10).
- The person-origin `ui.close` hook is registered (validate lists it) but never dispatched by any test.
- G-3 in-engine CPU time of parsing about 1.9 MB against the 10 s hook budget; the 2 MB test measures completion under bun only. G-5 the host's `PATH`. G-6 engine-written typings landing in a `--plugin-dir` folder. G-7 `tsc` is not on PATH, so no type check ran (the typings are advisory). G-9 `vscode` and `mobile` surfaces. G-10 early-access API drift. G-11 whether the mod loads in the operator's own session (rollout switch; Q7 open); the switch was not re-read at the end of this run.
- TDD ordering is session-asserted: the red runs above are recorded here, but git cannot witness them because tests and code share each milestone commit (verification-claim-integrity.md §2.3).
- Refused-tool notes (§3.1): the worktree guard refused a heredoc-bearing append, a `git show` inside a compound, and shell-variable paths; each was redone as a plain single command or through the file tools. No verification result was inferred from reading source in place of a refused command.

### Residual risk

- The platform contract is early access; the kit's behaviour (answers shaped `{ value }`, `{ cwd }`) was learned by running, not documented.
- Polling cost under the real 1.9 MB payload inside the isolate is unmeasured; the 15,000 ms floor is the plan's, not a measured optimum (Q3).
- Pick confirmation uses the exact label `Pick`; a dialog variant that returns the label with extra text would read as a no-op (safe direction).

## §E.3 Run-phase Audit-Ready Signal

run_status: audit-ready (engine and pure parts green; AC-MBM-014 and -015 remain manual Gap-class and are not recorded as passes)
run_complete_at: 2026-10-02
run_head: c0cc0915c (code at e323eba4f)

## §E.4 Sync-phase Audit-Ready Signal

_pending sync-phase_
