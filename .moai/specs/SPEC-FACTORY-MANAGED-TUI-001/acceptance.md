---
id: SPEC-FACTORY-MANAGED-TUI-001
title: "acceptance.md — acceptance criteria"
version: "0.3.4"
created: 2026-10-03
updated: 2026-10-03
author: GOOS (manager-spec)
tier: M
---

# acceptance.md — acceptance criteria

> This file is the verification layer; scenarios are Given-When-Then. The requirements (GEARS) live in `spec.md` §C, REQ-MT-001..014. Every AC is binary.
>
> **Command rules.** Commands are single invocations; `-run` patterns are anchored. Commands that need an alternation `|` are in the fenced block of §1.1 and the table points there. Lane sessions run `go test` as one compound with the environment scrub: `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_FACTORY_ROLE MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS MOAI_FACTORY_AUTO_DISPATCH MOAI_FACTORY_CLEAR_POLICY MOAI_KANBAN_BACKEND && go test …` (the table omits the prefix). The `internal/cli` package suite is not run locally. `-race` needs `CGO_ENABLED=1`.
>
> **Empty-sweep rule.** Before reading any `-run` result, `go test ./internal/cli -list '<same pattern>'` must name every test the row lists. `[no tests to run]` or an `ok` that lists no names is a failure (unmeasured), not a pass.
>
> **Observation conventions.** Log lines are read through the existing log seam with a mutex-guarded sink, polled up to 5 s at 10 ms; blocking calls run in goroutines under a 5 s watchdog (the HARDEN-001 conventions). "Exactly once" means a 500 ms quiet window. Terminal-ness is a package-private predicate the tests set; stdin for attached cases is an `*os.File` pipe.
>
> **Three classes.** *Structural-RED* cells (§1.2 E1-E7) are measured now and flip when the code lands; they support, and do not replace, the behavior tests. E1 excludes test files, so the RED commit R cannot flip it; E4 flips at S. *Adoption-deferred* ACs (every test-based AC: 001-009, 011, 016) have no starting observation at plan time, because their tests cannot compile against the base (M1 adds behavior-neutral compile stubs first, so each test then fails on a named assertion and not on one package build error). Their RED-now cell (command, observed stdout, exit code, tree SHA) is recorded per AC in `red-baseline.md` at M1, and **M2 may not start until the delta check of §1.1 (AC-MT-014 M1 part) has run on that file and confirmed a cell for every one of them**; until then those ACs are not adopted. *Regression-guard* ACs (010, 012) are green on the card base by design and carry a positive control.

## §1. AC matrix

| AC | Requirement | Scenario (Given-When-Then) | Verification → expected |
|---|---|---|---|
| AC-MT-001 | REQ-MT-001, REQ-MT-002, REQ-MT-009 | Given a managed Codex session over the fake codex shim, terminal predicate true, stdin an `*os.File`, operator args `-c model_reasoning_effort="high" -m fake-model`, When the owner attaches, Then the TUI child's argv is exactly `resume --remote ws://127.0.0.1:<port> --remote-auth-token-env MOAI_FACTORY_APP_SERVER_TOKEN -c model_reasoning_effort="high" -m fake-model <thread-id>` with `<thread-id>` equal to the id the fake server issued and the broker endpoint's session UUID; the env var named in argv is in the child's environment with the token-file content; the token string is in no argv of either child; the App Server argv still carries `--ws-token-file` and the three generated approval overrides; the TUI argv has none of `mcp_servers.moai` / `default_tools_approval_mode` that the operator did not type; the TUI's cwd equals the thread's cwd | `go test -race ./internal/cli -run '^TestManagedCodexTUIAttachCommand$' -count=1 -v` → subtests `argv_exact`, `token_not_in_argv`, `token_in_env`, `app_server_keeps_token_file`, `no_generated_approval_args`, `operator_args_forwarded`, `cwd_matches_thread` PASS, exit 0 |
| AC-MT-002 | REQ-MT-001 | Given the fake server and a shared append log, When the session starts, Then `tui-start` follows the priming `turn/completed` and the broker bind, never precedes either; and Given a failing priming turn, Then `tui-start` never appears and the owner returns the error | `go test -race ./internal/cli -run '^TestManagedCodexTUIStartsAfterPrimingAndBind$' -count=1 -v` → subtests `order`, `priming_failure_never_attaches` PASS, exit 0 |
| AC-MT-003 | REQ-MT-003, REQ-MT-004 | Given each of the cases named below, When the session runs, Then exactly one stderr line names the reason, no TUI runs, an operator stdin line still becomes a turn (headless), and the exit status is unchanged; in case `tui_start_fails` the App Server's stderr line is in the session log file (not on the terminal) and the log path was printed once. Cases (subtests): `stdin_not_terminal`, `stdin_not_file`, `stdout_not_terminal`, `probe_lacks_options`, `probe_times_out`, `probe_fails_to_run`, `opt_out_0`, `opt_out_false`, `opt_out_off`, `opt_out_uppercase_padded`, `tui_start_fails`. And the probe parser returns supported for the vendored real 0.160.0 help text (fixture `internal/cli/testdata/codex-0.160.0/resume-help.txt`) and unsupported for a help text without either option and for empty output | `go test -race ./internal/cli -run '^TestManagedCodexTUIPreconditionsAndFallback$' -count=1 -v` → the 11 named subtests PASS · `go test ./internal/cli -run '^TestManagedCodexRemoteSupportProbe$' -count=1 -v` → subtests `real_help_supported`, `no_options_unsupported`, `empty_unsupported` PASS, exit 0 · evidence for the refused-log case: `go test ./internal/cli -run '^TestManagedTUILogRefusalFallsBackHeadless$' -count=1 -v` → PASS (a symlink or non-directory component makes the session run headless with the one stderr line, no session error) |
| AC-MT-004 | REQ-MT-005 | Given an attached fake TUI reading its stdin, When `hello` and `/exit` are written to the launcher's stdin pipe, Then the fake TUI logs both lines, the fake server logs no `turn-prompt hello`, and the session does not end on `/exit` | `go test -race ./internal/cli -run '^TestManagedCodexTUIOwnsOperatorStdin$' -count=1 -v` → PASS, exit 0 |
| AC-MT-005 | REQ-MT-006 | Given an attached session whose fake App Server writes a line to its stderr, whose launcher answers a server request, logs a turn failure and hits a driver inbox error, When the TUI is attached, Then the captured terminal stderr holds only the one pre-attach line naming the log file, and the session log file under `.moai/logs/` holds all four. And When the fake TUI has exited and been reaped, Then the next owner log line (a server request the fake server sends during teardown) is on the terminal stderr and not in the file, while the App Server child's own stderr stays in the file | `go test -race ./internal/cli -run '^TestManagedCodexTUIKeepsTerminalClean$' -count=1 -v` → subtests `attached_lines_in_file`, `terminal_only_has_path_line`, `post_reap_line_on_terminal` PASS · E2 and E3 counts become `0` with exit 1 (§1.2) · evidence for the log-file open: `go test ./internal/cli -run '^TestManagedTUILogFileIsSafe$' -count=1 -v` (subtests `symlink_not_followed`, `existing_file_mode_tightened`, `new_file_mode`) and `go test ./internal/cli -run '^TestManagedTUILogOpenRefusesAliases$' -count=1 -v` (subtests `hard_link_at_log_path`, `symlinked_logs_dir`, `symlinked_moai_dir`) → PASS; `go test ./internal/cli -list '^TestManagedTUILog'` names 3 tests |
| AC-MT-006 | REQ-MT-007 | Given the fake server emits a turn that the TUI (not the launcher) started and a broker message is waiting, When the thread is busy, Then no claim and no launcher `turn/start` happens; after that turn completes the launcher claims and starts exactly one turn. Given 40 operator-turn frame pairs between launcher turns, Then a later server request is still answered within 5 s. Given a busy turn that outlasts the (lowered) warn interval, Then one log line per elapsed interval appears, the launcher still starts no turn, and delivery resumes when the turn completes. Given a fake server that never sends lifecycle frames for operator turns (P6 false), Then the launcher neither deadlocks nor hangs (the degraded path of known debt 3) | §1.1 AC-MT-006 block → `TestManagedCodexTUIDefersDeliveryWhileBusy` (subtests `deferred_then_delivered`, `frames_not_delivered`), `TestManagedCodexReaderSurvivesOperatorTurns`, `TestManagedCodexBusyLongTurnKeepsDeferring` PASS, exit 0 |
| AC-MT-007 | REQ-MT-008 | Given the TUI attached, When the fake server sends each of the six turn-carrying request kinds naming (i) a launcher-started turn, (ii) an operator-started turn, (iii) no known turn id because the launcher's own `turn/start` is outstanding (REQ-MT-008's exception), Then (i) and (iii) get today's HARDEN-001 answer, (ii) gets no reply from the launcher, and no answer is ever an accept; the four kinds without `turnId` keep today's answer; and with no TUI attached (ii) is declined exactly as before | `go test -race ./internal/cli -run '^TestManagedCodexServerRequestScopingWithTUI$' -count=1 -v` → subtests `owned_turn_answered`, `operator_turn_unanswered`, `armed_window_answered`, `no_turn_id_kinds_answered`, `detached_unchanged`, `never_accepts` PASS, exit 0 |
| AC-MT-008 | REQ-MT-010 | Given an attached fake TUI, When it exits 0, 7, or is killed by a signal, Then the driver returns nil, an error with `ExitCode()` 7, and an error with `ExitCode()` 1 respectively, the App Server child is reaped and the token directory removed in each, and when the driver had already failed, the driver's error stands. And (`blocked_turn_start_released`) Given the fake server has stopped reading and a test barrier shows the launcher's `turn/start` write has **entered `write()` and not returned**, the per-write deadline is set **longer than the 5 s watchdog** (so the deadline cannot be what releases it), and a **second sender is queued on the write lock** behind the blocked write, When the fake TUI exits with code 7 (the TUI status is recorded before the connection is closed), Then the wait goroutine closes the WebSocket connection, the blocked write and the queued sender both return within the 5 s watchdog, the driver returns the TUI's result (exit code 7), and the TUI child is reaped; and (`blocked_wait_turn_released`) the same exit code 7 for a launcher turn waiting for a completion. A release implemented only through a `select` channel leaves the write blocked and fails the first subtest | `go test -race ./internal/cli -run '^TestManagedCodexTUIExitEndsSession$' -count=1 -v` → subtests `exit_zero`, `exit_seven`, `signaled`, `session_error_wins`, `blocked_turn_start_released`, `blocked_wait_turn_released` PASS, exit 0 |
| AC-MT-009 | REQ-MT-011 | Given an attached fake TUI that handles interrupt, and another that ignores it, When the App Server connection closes, and When the driver ends the session after the consecutive-failure ceiling, Then the interrupting TUI is interrupted and exits, the ignoring one is killed after the grace, the owner returns only after the TUI is reaped, and no TUI process remains | §1.1 AC-MT-009 block → `TestManagedCodexServerDeathStopsTUI`, `TestManagedCodexSessionEndStopsTUI` PASS, exit 0 |
| AC-MT-010 | REQ-MT-001 (gate), REQ-MT-012 | Regression guard. Given no `MOAI_FACTORY_MANAGED`, or stamps without it, or it with no stamps, When a Codex launch runs, Then the four opt-in tests that exist now stay green: `TestFactoryManagedRequested`, `TestManagedLaunchRequiresOptIn`, `TestManagedCodexLaunchRequiresOptIn` (in `managed_optin_test.go` at lines 34, 64 and 111, measured on `42a952661`) and `TestManagedCardChildSwitchOffKeepsDirectDoor` (in `managed_card_child_test.go` at line 216). The original fourth name, `TestManagedSwitchDoesNotReachCodexLaneLoop`, was removed upstream by card t1440 (commit `a184aa89c`); the intent (no managed branch without the opt-in) is carried by these tests plus AC-MT-016. Falsifiable: `-list` must name exactly these 4 (count 4, not 0) | §1.1 AC-MT-010 block → 4 tests PASS, exit 0 |
| AC-MT-011 | REQ-MT-001, REQ-MT-007, REQ-MT-010 | Loopback, no real codex. Given a temp broker with one inbox message and the fake codex shim in both roles, When the owner starts, the fake TUI attaches with the env token and its own WS connection, and the fake model acts on the injected prompt, Then metadata-only injection, body read by claim token, and receipt complete (acknowledged 1, pending 0); then the fake TUI exits 0 on a line written to the launcher's stdin pipe and the session ends nil with teardown complete | `go test -race ./internal/cli -run '^TestManagedCodexTUILoopbackRoundTrip$' -count=1 -v` → PASS, exit 0 |
| AC-MT-012 | REQ-MT-012 | Regression guard. Given the changed tree, Then the cross build passes, `managed_*` files (tests included) hold no `syscall.`, and `store.go` and both parent SPEC directories are unchanged | `GOOS=windows GOARCH=amd64 go build ./...` → exit 0 · `grep -rn 'syscall\.' internal/cli/managed_*.go` → no lines, exit 1 (positive control `grep -ln 'syscall\.' internal/cli/launch_exec_posix.go` → that path, exit 0) · `git diff --stat "$(git merge-base develop HEAD)"..HEAD -- internal/factorymsg/store.go .moai/specs/SPEC-FACTORY-MANAGED-SESSION-001 .moai/specs/SPEC-FACTORY-MANAGED-HARDEN-001` → no lines (read the merge-base at evaluation, before merge; positive control the same range on `internal/cli` is non-empty) |
| AC-MT-013 | REQ-MT-013 | Given the synced tree, Then the two sentences saying the screen shows nothing are gone from the operator document, it cites this SPEC, and every disclosure anchor of §1.3 is present | the grep list of §1.3 |
| AC-MT-014 | REQ-MT-014 | Two parts. **M1 part (evaluated at the end of M1):** one behavior-neutral stub commit S precedes the RED commit R; R holds the tests and the tracked `red-baseline.md`; `red-baseline.md` has rows for **11 distinct** adoption-deferred AC ids (001-009, 011, 016), failure lines for **15 distinct named top-level tests** (the 15 tests the rows 001-009, 011 and 016 name, one failure line each), and no `undefined:`, `build failed`, `panic:` or `test timed out` line; its "tree SHA" cell is S's commit SHA with the R tests applied (R's own tree depends on the file's content and cannot be cited). **Closing part (re-run at M5 and in the final evidence):** R is an ancestor of every commit F that changes non-test files under `internal/cli` or `internal/config` after S | M1 part, the stated delta-check command run by the orchestrator or the lane (§1.1 AC-MT-014 block) → `11`, `11` or more, `0`, plus `git merge-base --is-ancestor <S> <R>` exit 0 and the reverse exit 1, `git ls-files .moai/specs/SPEC-FACTORY-MANAGED-TUI-001/red-baseline.md` → that path, `git check-ignore -v` on it → no output, exit 1 · closing part: `git merge-base --is-ancestor <R> <F>` → exit 0 and the reverse exit 1 for each F (SHAs cited in progress.md §E.2) |
| AC-MT-015 | REQ-MT-001, REQ-MT-005, REQ-MT-007, REQ-MT-008, REQ-MT-010, REQ-MT-011 | **MANUAL, operator-held (keep-set), NOT RUN IN CI, not release-blocking.** The only observation of a real TUI; procedure in §4. No terminal is designated; the sync-phase deliverable records "operator confirmation pending, no terminal designated" until the operator runs it | operator records `.moai/reports/t1408/manual-attach-check.md`; no command is claimed as run |
| AC-MT-016 | REQ-MT-001 (gate) | Adoption-deferred. Given the opt-in absent, or stamps only, or the switch only, When the owner entry is reached through the launcher seams, Then no capability probe process is started and no TUI child is started (probe and TUI spawn seams record zero calls); and with the opt-in present the probe runs once | `go test ./internal/cli -run '^TestManagedTUINeverReachedWithoutOptIn$' -count=1 -v` → subtests `no_switch`, `switch_only`, `stamps_only`, `opted_in_probes_once` PASS, exit 0 (the test does not exist on the base; its RED cell is recorded at M1) |

### §1.1 Commands containing a pipe (verbatim, outside the table)

**AC-MT-006**

```
go test ./internal/cli -list '^(TestManagedCodexTUIDefersDeliveryWhileBusy|TestManagedCodexReaderSurvivesOperatorTurns|TestManagedCodexBusyLongTurnKeepsDeferring)$'
go test -race ./internal/cli -run '^(TestManagedCodexTUIDefersDeliveryWhileBusy|TestManagedCodexReaderSurvivesOperatorTurns|TestManagedCodexBusyLongTurnKeepsDeferring)$' -count=1 -v
```

**AC-MT-009**

```
go test ./internal/cli -list '^(TestManagedCodexServerDeathStopsTUI|TestManagedCodexSessionEndStopsTUI)$'
go test -race ./internal/cli -run '^(TestManagedCodexServerDeathStopsTUI|TestManagedCodexSessionEndStopsTUI)$' -count=1 -v
```

**AC-MT-010**

```
go test ./internal/cli -list '^(TestFactoryManagedRequested|TestManagedLaunchRequiresOptIn|TestManagedCodexLaunchRequiresOptIn|TestManagedCardChildSwitchOffKeepsDirectDoor)$'
go test ./internal/cli -run '^(TestFactoryManagedRequested|TestManagedLaunchRequiresOptIn|TestManagedCodexLaunchRequiresOptIn|TestManagedCardChildSwitchOffKeepsDirectDoor)$' -count=1 -v
```

**AC-MT-014 (M1 part — the delta check)**

Run by the orchestrator, or by the lane that wrote M1, at the end of M1 and again by whoever advances the card to M2; the result line (three numbers and the ancestry exit codes) is pasted into progress.md §E.2. No new audit role is involved. The numbers are a floor, not proof: a fabricated cell passes them. What stops a fabricated cell is that whoever advances the card to M2 **re-runs at least one cited RED command** from `red-baseline.md` on S with the R tests applied and sees the same named failure.

```
grep -o -E '^[|] AC-MT-[0-9]+' .moai/specs/SPEC-FACTORY-MANAGED-TUI-001/red-baseline.md | sort -u | wc -l
grep -o -E -e '--- FAIL: [A-Za-z0-9_]+' .moai/specs/SPEC-FACTORY-MANAGED-TUI-001/red-baseline.md | sort -u | wc -l
grep -c -e 'undefined:' -e 'build failed' -e 'panic:' -e 'test timed out' .moai/specs/SPEC-FACTORY-MANAGED-TUI-001/red-baseline.md
```

Expected `11` (distinct ids), then `15` or more (distinct named tests), then `0` (exit 1 for the last, which is the passing result).

### §1.2 RED-now evidence ledger (tree `2b9e4a4d0`, measured in this plan run)

| id | AC | Command | Stdout | Exit | Why red / green path |
|---|---|---|---|---|---|
| E1 | AC-MT-001 | `grep -rln --include='*.go' --exclude='*_test.go' 'remote-auth-token-env' internal/cli` | (empty) | 1 | no code builds the remote TUI command line; green after M2: the new file and its test |
| E2 | AC-MT-005 | `grep -c 'fmt.Fprintln(os.Stderr, "Factory inbox:"' internal/cli/managed_factory_session.go` | `1` | 0 | driver writes straight to the terminal; green: `0`, exit 1 |
| E3 | AC-MT-005 | `grep -c 's.cmd.Stderr = os.Stderr' internal/cli/managed_codex_factory.go` | `1` | 0 | App Server stderr is the terminal; green: `0`, exit 1 |
| E4 | AC-MT-001 | `grep -n 'MOAI_FACTORY_APP_SERVER_TOKEN' internal/config/envkeys.go` | (empty) | 1 | constant absent; flips at the stub commit S (a constant, not behavior — it does not witness the attach); green: one definition line, exit 0 |
| E5 | AC-MT-013 | `grep -c '화면에는 아무것도 나타나지 않는다' .moai/docs/factory-managed-session.md` | `1` | 0 | green: `0`, exit 1 |
| E6 | AC-MT-013 | `grep -c 'Codex 관리 세션은 화면에 아무것도 보여 주지 않는다' .moai/docs/factory-managed-session.md` | `1` | 0 | green: `0`, exit 1 |
| E7 | AC-MT-013 | `grep -c 'SPEC-FACTORY-MANAGED-TUI-001' .moai/docs/factory-managed-session.md` | `0` | 1 | green: 1 or more |
| E8 | AC-MT-012 (guard) | `grep -rn 'syscall\.' internal/cli/managed_*.go` | (empty) | 1 | green on the base by design; positive control `grep -ln 'syscall\.' internal/cli/launch_exec_posix.go` → `internal/cli/launch_exec_posix.go`, exit 0 |

Test-based ACs (001-009, 011, 016) are adoption-deferred (header, "Three classes"): no plan-phase RED-now cell exists for them, and none is faked with a "test name absent" grep. The M1 gate and the delta check on `red-baseline.md` are what adopt them. A grep for the chosen names over `internal/cli/*_test.go` found none (exit 1), so the names are free; the compiled `-list` was not run in this plan. This is a Gap, stated, not a pass.

### §1.3 AC-MT-013 grep list (synced tree)

Observed on the current tree (`42a952661`, 2026-10-04), same commands as the rows below: old sentence `화면에는 아무것도 나타나지 않는다` → `0`; `SPEC-FACTORY-MANAGED-TUI-001` in the operator document → `3`; anchors with same-line content: opt-out `1`, log-file `1`, quit `1`, probe `1`, signals(t1459) `1`, unobserved `3`, manual-check `1`, debt-status `1` (that line contains `addressed` once and `unverified` once, `resolved` zero times); card-child headless bullet `1`; CHANGELOG `SPEC-FACTORY-MANAGED-TUI-001` `1`, CHANGELOG `anchor:tui-` `6`. The bare reads `Codex 관리 세션은 화면에 아무것도 보여 주지 않는다` → `0` and `t1459` → `2` are recorded for the CARD-CHILD-001 operator-document criterion (twelfth)'s benefit and are not this AC's rows.

Base counts were measured on tree `2b9e4a4d0`; every row must change from its base. Disclosure lines carry a machine anchor on the same line as the required content, so a document that only lists search tokens does not satisfy a row.

| Command | Base (count, exit) | Expect after sync |
|---|---|---|
| `grep -c '화면에는 아무것도 나타나지 않는다' .moai/docs/factory-managed-session.md` | `1`, 0 | `0`, exit 1 |
| `grep -c -e 'moai codex -l. 의 관리 카드 자식은 TUI 를 붙이지 않고 헤드리스로 남는다' .moai/docs/factory-managed-session.md` | `0`, 1 on the pre-card-child base (not re-measurable now) | 1 or more — the headless fact lives in this card-child bullet; the old sentence `Codex 관리 세션은 화면에 아무것도 보여 주지 않는다` is deliberately **not** constrained here, because the completed SPEC-FACTORY-MANAGED-CARD-CHILD-001 the CARD-CHILD-001 operator-document criterion (twelfth) owns it (it expects 1 or more) |
| `grep -c 'SPEC-FACTORY-MANAGED-TUI-001' .moai/docs/factory-managed-session.md` | `0`, 1 | 1 or more |
| `grep -c 'anchor:tui-opt-out.*MOAI_FACTORY_MANAGED_TUI' .moai/docs/factory-managed-session.md` | `0`, 1 | 1 or more |
| `grep -c 'anchor:tui-log-file.*factory-managed-' .moai/docs/factory-managed-session.md` | `0`, 1 | 1 or more |
| `grep -c 'anchor:tui-quit.*/quit' .moai/docs/factory-managed-session.md` | `0`, 1 | 1 or more |
| `grep -c 'anchor:tui-probe.*--remote' .moai/docs/factory-managed-session.md` | `0`, 1 | 1 or more |
| `grep -c 'anchor:tui-signals.*t1459' .moai/docs/factory-managed-session.md` | `0`, 1 | 1 or more (the bare token `t1459` already counts `1` on the base and proves nothing) |
| `grep -c 'anchor:tui-unobserved.*미관측' .moai/docs/factory-managed-session.md` | `0`, 1 | 3 or more |
| `grep -c 'anchor:tui-manual-check.*AC-MT-015' .moai/docs/factory-managed-session.md` | `0`, 1 | 1 or more |
| `grep -c 'anchor:tui-debt-status' .moai/docs/factory-managed-session.md` | `0`, 1 | 1; the line says "addressed" and "unverified" (both words on that line); it may say "resolved" only after `.moai/reports/t1408/manual-attach-check.md` records steps 1, 2, 3, 6 as observed |
| `grep -c 'SPEC-FACTORY-MANAGED-TUI-001' CHANGELOG.md` | `0`, 1 | 1 or more |
| `grep -c 'anchor:tui-' CHANGELOG.md` | `0`, 1 | 5 or more (opt-out, log file, `/quit`, signals-open, debt status) |

## §2. Mutant probe

Each mutant is a one-edit change the named test must turn red; M1/M2 confirm each is writable.

| Mutant | Caught by |
|---|---|
| mu1 token value in TUI argv | AC-MT-001 `token_not_in_argv` |
| mu2 generated approval overrides added to TUI argv | AC-MT-001 `no_generated_approval_args` |
| mu3 attach before the priming turn completes | AC-MT-002 `order` |
| mu4 attach with a non-terminal stdin | AC-MT-003 |
| mu5 driver keeps its stdin reader while attached | AC-MT-004 |
| mu6 launcher log line written to terminal stderr while attached | AC-MT-005 |
| mu7 App Server stderr left on the terminal | AC-MT-005 |
| mu8 claim ignores busy | AC-MT-006 `TestManagedCodexTUIDefersDeliveryWhileBusy` |
| mu9 reader feeds the event channel with operator-turn frames | AC-MT-006 `TestManagedCodexReaderSurvivesOperatorTurns` |
| mu10 operator-turn request answered by the launcher | AC-MT-007 `operator_turn_unanswered` |
| mu11 scoping left on when detached | AC-MT-007 `detached_unchanged` |
| mu12 TUI exit status dropped to 0 | AC-MT-008 `exit_seven` |
| mu13 TUI not stopped at session end | AC-MT-009 |
| mu14 managed TUI reachable without the opt-in | AC-MT-016 `no_switch`, `switch_only`, `stamps_only` |
| mu15 TUI exit does not release a blocked `turn/start` write or `waitTurn`; mu19 release only through a `select` channel (the write stays blocked, the barrier shows it, the subtest fails) | AC-MT-008 `blocked_turn_start_released`, `blocked_wait_turn_released` |
| mu16 busy flag self-clears after the warn interval | AC-MT-006 `TestManagedCodexBusyLongTurnKeepsDeferring` |
| mu17 documentation stuffed with search tokens and no content | AC-MT-013 (anchor and content on one line; base counts) |
| mu18 post-reap log line still written to the file | AC-MT-005 `post_reap_line_on_terminal` |

## §3. Edge cases

- Keystrokes typed during `Start` and priming wait in the tty buffer and reach the TUI.
- The TUI exits before the first claim: the session ends, no delivery was attempted.
- Bind failure or failing priming: the TUI is never started.
- A claimed batch and an operator turn starting in the race window: steered into the active turn (known debt 4); the lease (2 min) bounds the batch.
- Stdin at EOF with a terminal predicate true is not a terminal case; EOF only matters headless.
- A path with spaces in the codex binary: the exec form takes argv, no shell.
- An operator turn that starts and completes inside the owner's armed window is attributed to the owner (known debt 14); `TestManagedOperatorTurnInsideArmedWindowKnownDebt` pins the current behavior and is not an acceptance test of a fix.

## §4. AC-MT-015 — manual check (operator, NOT RUN IN CI)

Preconditions: a real terminal, authenticated codex-cli 0.160.0, a factory run with the stamps, `MOAI_FACTORY_MANAGED=1`, launch `moai codex`. Record each step's result and the codex version in `.moai/reports/t1408/manual-attach-check.md`. Nothing here is claimed as run by this plan.

1. The TUI appears after the priming turn; the launcher printed the log-file path first.
2. Type a message: the model answers in the TUI (P9: resume by id works).
3. From the leader, send a broker message while the thread is idle: a launcher-started turn runs, is visible in the TUI (P9), and the receipt is recorded.
4. Start a long operator turn, then send a broker message: the batch waits until the turn ends (P6); record whether it waited or steered.
5. Trigger a command-approval in an operator turn: it appears in the TUI and the launcher did not decline it (P7); record the log file lines.
6. `/quit` in the TUI: the shell prompt returns, exit 0, `pgrep -f 'codex app-server'` finds none, terminal usable.
7. Kill the App Server from another shell: the TUI stops, exit 1, terminal state recorded (known debt 5).
8. `MOAI_FACTORY_MANAGED_TUI=off`: headless behavior, one notice.

Failure disposition: any failed step becomes a named known debt or a follow-up card; step 2 or 3 failing means the opt-out should become opt-in. Until the operator runs this check the documentation says "addressed, unverified" (REQ-MT-013).

## §5. Quality gates and Definition of Done

TRUST 5: new file coverage 85% or more by `go test -cover` on the touched tests; `gofmt -l internal/cli internal/config` empty; `go vet ./internal/cli ./internal/config`; `golangci-lint run ./internal/cli/... ./internal/config/...` at the CI version; managed top-level test count not below 79 plus the new tests (`go test ./internal/cli -list '^TestManaged'` names 85 on `42a952661`, measured). Coverage of `internal/cli/managed_codex_tui.go` after the repairs: 219/241 statements = 90.9%, as reported by the repair lane (not re-measured by the SPEC author); the earlier 208/224 figure is superseded. DoD: AC-MT-001..014 and 016 PASS with command and output; AC-MT-015 is operator-held and not release-blocking: the sync-phase deliverable records "operator confirmation pending, no terminal designated" and the docs say "addressed, unverified", never "resolved", until the operator runs it; cross build and grep guards cited; CHANGELOG and operator document updated; the sync audit re-reads the premise ledger.
