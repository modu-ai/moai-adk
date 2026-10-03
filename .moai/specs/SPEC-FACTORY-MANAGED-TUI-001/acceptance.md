---
id: SPEC-FACTORY-MANAGED-TUI-001
title: "acceptance.md — acceptance criteria"
version: "0.1.0"
created: 2026-10-03
updated: 2026-10-03
author: GOOS (manager-spec)
tier: M
---

# acceptance.md — AC-MT-001..015

> This file is the verification layer; scenarios are Given-When-Then. The requirements (GEARS) live in `spec.md` §C, REQ-MT-001..014. Every AC is binary.
>
> **Command rules.** Commands are single invocations; `-run` patterns are anchored. Commands that need an alternation `|` are in the fenced block of §1.1 and the table points there. Lane sessions run `go test` as one compound with the environment scrub: `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_FACTORY_ROLE MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS MOAI_FACTORY_AUTO_DISPATCH MOAI_FACTORY_CLEAR_POLICY MOAI_KANBAN_BACKEND && go test …` (the table omits the prefix). The `internal/cli` package suite is not run locally. `-race` needs `CGO_ENABLED=1`.
>
> **Empty-sweep rule.** Before reading any `-run` result, `go test ./internal/cli -list '<same pattern>'` must name every test the row lists. `[no tests to run]` or an `ok` that lists no names is a failure (unmeasured), not a pass.
>
> **Observation conventions.** Log lines are read through the existing log seam with a mutex-guarded sink, polled up to 5 s at 10 ms; blocking calls run in goroutines under a 5 s watchdog (the HARDEN-001 conventions). "Exactly once" means a 500 ms quiet window. Terminal-ness is a package-private predicate the tests set; stdin for attached cases is an `*os.File` pipe.
>
> **Two classes.** *Release-blocking* ACs are adopted by M1 with a RED-now cell (§1.2). *Regression-guard* ACs are green on the card base by design and carry a positive control; they are not RED-adopted.

## §1. AC matrix

| AC | Requirement | Scenario (Given-When-Then) | Verification → expected |
|---|---|---|---|
| AC-MT-001 | REQ-MT-001, REQ-MT-002, REQ-MT-009 | Given a managed Codex session over the fake codex shim, terminal predicate true, stdin an `*os.File`, operator args `-c model_reasoning_effort="high" -m fake-model`, When the owner attaches, Then the TUI child's argv is exactly `resume --remote ws://127.0.0.1:<port> --remote-auth-token-env MOAI_FACTORY_APP_SERVER_TOKEN -c model_reasoning_effort="high" -m fake-model <thread-id>` with `<thread-id>` equal to the id the fake server issued and the broker endpoint's session UUID; the env var named in argv is in the child's environment with the token-file content; the token string is in no argv of either child; the App Server argv still carries `--ws-token-file` and the three generated approval overrides; the TUI argv has none of `mcp_servers.moai` / `default_tools_approval_mode` that the operator did not type; the TUI's cwd equals the thread's cwd | `go test -race ./internal/cli -run '^TestManagedCodexTUIAttachCommand$' -count=1 -v` → subtests `argv_exact`, `token_not_in_argv`, `token_in_env`, `app_server_keeps_token_file`, `no_generated_approval_args`, `operator_args_forwarded`, `cwd_matches_thread` PASS, exit 0 |
| AC-MT-002 | REQ-MT-001 | Given the fake server and a shared append log, When the session starts, Then `tui-start` follows the priming `turn/completed` and the broker bind, never precedes either; and Given a failing priming turn, Then `tui-start` never appears and the owner returns the error | `go test -race ./internal/cli -run '^TestManagedCodexTUIStartsAfterPrimingAndBind$' -count=1 -v` → subtests `order`, `priming_failure_never_attaches` PASS, exit 0 |
| AC-MT-003 | REQ-MT-003, REQ-MT-004 | Given each of: stdin not a terminal; stdin not an `*os.File`; stdout not a terminal; a `resume --help` without the options; a probe that exceeds the timeout; a probe that fails to run; opt-out `0`, `false`, `off` (and `OFF`, padded); a TUI that fails to start, When the session runs, Then exactly one stderr line names the reason, no TUI runs, an operator stdin line still becomes a turn (headless), and the exit status is unchanged. And the probe parser returns supported for the vendored real 0.160.0 help text and unsupported for a help text without either option and for empty output | `go test -race ./internal/cli -run '^TestManagedCodexTUIPreconditionsAndFallback$' -count=1 -v` → 11 subtests PASS · `go test ./internal/cli -run '^TestManagedCodexRemoteSupportProbe$' -count=1 -v` → 3 subtests PASS, exit 0 |
| AC-MT-004 | REQ-MT-005 | Given an attached fake TUI reading its stdin, When `hello` and `/exit` are written to the launcher's stdin pipe, Then the fake TUI logs both lines, the fake server logs no `turn-prompt hello`, and the session does not end on `/exit` | `go test -race ./internal/cli -run '^TestManagedCodexTUIOwnsOperatorStdin$' -count=1 -v` → PASS, exit 0 |
| AC-MT-005 | REQ-MT-006 | Given an attached session whose fake App Server writes a line to its stderr, whose launcher answers a server request, logs a turn failure and hits a driver inbox error, When the TUI is attached, Then the captured terminal stderr holds only the one pre-attach line naming the log file, the session log file under `.moai/logs/` holds all four, and after the TUI exits the next log line goes to the terminal again | `go test -race ./internal/cli -run '^TestManagedCodexTUIKeepsTerminalClean$' -count=1 -v` → PASS · E2 and E3 counts become `0` with exit 1 (§1.2) |
| AC-MT-006 | REQ-MT-007 | Given the fake server emits a turn that the TUI (not the launcher) started and a broker message is waiting, When the thread is busy, Then no claim and no launcher `turn/start` happens; after that turn completes the launcher claims and starts exactly one turn. Given 40 operator-turn frame pairs between launcher turns, Then a later server request is still answered within 5 s. Given a busy turn that never completes, Then after the ceiling busy reads false with one log line | §1.1 AC-MT-006 block → 3 tests PASS (`TestManagedCodexTUIDefersDeliveryWhileBusy`, `TestManagedCodexReaderSurvivesOperatorTurns`, `TestManagedCodexBusyCeilingReleasesStuckBusy`), exit 0 |
| AC-MT-007 | REQ-MT-008 | Given the TUI attached, When the fake server sends each of the six turn-carrying request kinds naming (i) a launcher-started turn, (ii) an operator-started turn, (iii) no turn id yet while the launcher's window is open, Then (i) and (iii) get today's HARDEN-001 answer, (ii) gets no reply from the launcher, and no answer is ever an accept; the four kinds without `turnId` keep today's answer; and with no TUI attached (ii) is declined exactly as before | `go test -race ./internal/cli -run '^TestManagedCodexServerRequestScopingWithTUI$' -count=1 -v` → subtests `owned_turn_answered`, `operator_turn_unanswered`, `armed_window_answered`, `no_turn_id_kinds_answered`, `detached_unchanged`, `never_accepts` PASS, exit 0 |
| AC-MT-008 | REQ-MT-010 | Given an attached fake TUI, When it exits 0, 7, or is killed by a signal, Then the driver returns nil, an error with `ExitCode()` 7, and an error with `ExitCode()` 1 respectively, the App Server child is reaped and the token directory removed in each, and when the driver had already failed, the driver's error stands | `go test -race ./internal/cli -run '^TestManagedCodexTUIExitEndsSession$' -count=1 -v` → subtests `exit_zero`, `exit_seven`, `signaled`, `session_error_wins` PASS, exit 0 |
| AC-MT-009 | REQ-MT-011 | Given an attached fake TUI that handles interrupt, and another that ignores it, When the App Server connection closes, and When the driver ends the session after the consecutive-failure ceiling, Then the interrupting TUI is interrupted and exits, the ignoring one is killed after the grace, the owner returns only after the TUI is reaped, and no TUI process remains | §1.1 AC-MT-009 block → `TestManagedCodexServerDeathStopsTUI`, `TestManagedCodexSessionEndStopsTUI` PASS, exit 0 |
| AC-MT-010 | REQ-MT-001 (gate), REQ-MT-012 | Regression guard. Given no `MOAI_FACTORY_MANAGED`, or stamps without it, or it with no stamps, When a Codex launch runs, Then no probe and no TUI run; the four opt-in tests of the parent stay green | §1.1 AC-MT-010 block → 5 tests PASS, exit 0 |
| AC-MT-011 | REQ-MT-001, REQ-MT-007, REQ-MT-010 | Loopback, no real codex. Given a temp broker with one inbox message and the fake codex shim in both roles, When the owner starts, the fake TUI attaches with the env token and its own WS connection, and the fake model acts on the injected prompt, Then metadata-only injection, body read by claim token, and receipt complete (acknowledged 1, pending 0); then the fake TUI exits 0 on a line written to the launcher's stdin pipe and the session ends nil with teardown complete | `go test -race ./internal/cli -run '^TestManagedCodexTUILoopbackRoundTrip$' -count=1 -v` → PASS, exit 0 |
| AC-MT-012 | REQ-MT-012 | Regression guard. Given the changed tree, Then the cross build passes, `managed_*` files (tests included) hold no `syscall.`, and `store.go` and both parent SPEC directories are unchanged | `GOOS=windows GOARCH=amd64 go build ./...` → exit 0 · `grep -rn 'syscall\.' internal/cli/managed_*.go` → no lines, exit 1 (positive control `grep -ln 'syscall\.' internal/cli/launch_exec_posix.go` → that path, exit 0) · `git diff --stat "$(git merge-base develop HEAD)"..HEAD -- internal/factorymsg/store.go .moai/specs/SPEC-FACTORY-MANAGED-SESSION-001 .moai/specs/SPEC-FACTORY-MANAGED-HARDEN-001` → no lines (read the merge-base at evaluation, before merge; positive control the same range on `internal/cli` is non-empty) |
| AC-MT-013 | REQ-MT-013 | Given the synced tree, Then the two sentences saying the screen shows nothing are gone from the operator document, it cites this SPEC, and every disclosure anchor of §1.3 is present | the grep list of §1.3 |
| AC-MT-014 | REQ-MT-014 | Given the run history, Then the tests and the tracked `red-baseline.md` are in one commit that is an ancestor of every fix commit, and the file is tracked and not ignored | `git merge-base --is-ancestor <RED> <FIX>` → exit 0 and the reverse → exit 1 · `git ls-files .moai/specs/SPEC-FACTORY-MANAGED-TUI-001/red-baseline.md` → that path · `git check-ignore -v .moai/specs/SPEC-FACTORY-MANAGED-TUI-001/red-baseline.md` → no output, exit 1 (SHAs cited in progress.md §E.2) |
| AC-MT-015 | REQ-MT-001, REQ-MT-005, REQ-MT-007, REQ-MT-008, REQ-MT-010, REQ-MT-011 | **MANUAL, operator-run, NOT RUN IN CI.** The only observation of a real TUI; procedure in §4 | operator records `.moai/reports/t1408/manual-attach-check.md`; no command is claimed as run |

### §1.1 Commands containing a pipe (verbatim, outside the table)

**AC-MT-006**

```
go test ./internal/cli -list '^(TestManagedCodexTUIDefersDeliveryWhileBusy|TestManagedCodexReaderSurvivesOperatorTurns|TestManagedCodexBusyCeilingReleasesStuckBusy)$'
go test -race ./internal/cli -run '^(TestManagedCodexTUIDefersDeliveryWhileBusy|TestManagedCodexReaderSurvivesOperatorTurns|TestManagedCodexBusyCeilingReleasesStuckBusy)$' -count=1 -v
```

**AC-MT-009**

```
go test ./internal/cli -list '^(TestManagedCodexServerDeathStopsTUI|TestManagedCodexSessionEndStopsTUI)$'
go test -race ./internal/cli -run '^(TestManagedCodexServerDeathStopsTUI|TestManagedCodexSessionEndStopsTUI)$' -count=1 -v
```

**AC-MT-010**

```
go test ./internal/cli -list '^(TestFactoryManagedRequested|TestManagedLaunchRequiresOptIn|TestManagedCodexLaunchRequiresOptIn|TestManagedSwitchDoesNotReachCodexLaneLoop|TestManagedTUINeverReachedWithoutOptIn)$'
go test ./internal/cli -run '^(TestFactoryManagedRequested|TestManagedLaunchRequiresOptIn|TestManagedCodexLaunchRequiresOptIn|TestManagedSwitchDoesNotReachCodexLaneLoop|TestManagedTUINeverReachedWithoutOptIn)$' -count=1 -v
```

### §1.2 RED-now evidence ledger (tree `2b9e4a4d0`, measured in this plan run)

| id | AC | Command | Stdout | Exit | Why red / green path |
|---|---|---|---|---|---|
| E1 | AC-MT-001 | `grep -rln --include='*.go' 'remote-auth-token-env' internal/cli` | (empty) | 1 | no code builds the remote TUI command line; green after M2: the new file and its test |
| E2 | AC-MT-005 | `grep -c 'fmt.Fprintln(os.Stderr, "Factory inbox:"' internal/cli/managed_factory_session.go` | `1` | 0 | driver writes straight to the terminal; green: `0`, exit 1 |
| E3 | AC-MT-005 | `grep -c 's.cmd.Stderr = os.Stderr' internal/cli/managed_codex_factory.go` | `1` | 0 | App Server stderr is the terminal; green: `0`, exit 1 |
| E4 | AC-MT-001 | `grep -n 'MOAI_FACTORY_APP_SERVER_TOKEN' internal/config/envkeys.go` | (empty) | 1 | constant absent; green: one definition line, exit 0 |
| E5 | AC-MT-013 | `grep -c '화면에는 아무것도 나타나지 않는다' .moai/docs/factory-managed-session.md` | `1` | 0 | green: `0`, exit 1 |
| E6 | AC-MT-013 | `grep -c 'Codex 관리 세션은 화면에 아무것도 보여 주지 않는다' .moai/docs/factory-managed-session.md` | `1` | 0 | green: `0`, exit 1 |
| E7 | AC-MT-013 | `grep -c 'SPEC-FACTORY-MANAGED-TUI-001' .moai/docs/factory-managed-session.md` | `0` | 1 | green: 1 or more |
| E8 | AC-MT-012 (guard) | `grep -rn 'syscall\.' internal/cli/managed_*.go` | (empty) | 1 | green on the base by design; positive control `grep -ln 'syscall\.' internal/cli/launch_exec_posix.go` → `internal/cli/launch_exec_posix.go`, exit 0 |

Test-based ACs (001-011) have no starting observation yet: their tests do not exist, and a `-run` selecting zero names prints `[no tests to run]`, which is an empty sweep and proves nothing. A grep for the chosen names over `internal/cli/*_test.go` found none (exit 1), so the names are free. The compiled `-list` was not run in this plan; M1 records the real RED output in `red-baseline.md`. This is a Gap, not a pass.

### §1.3 AC-MT-013 grep list (synced tree)

| Command | Expect |
|---|---|
| `grep -c '화면에는 아무것도 나타나지 않는다' .moai/docs/factory-managed-session.md` | `0`, exit 1 |
| `grep -c 'Codex 관리 세션은 화면에 아무것도 보여 주지 않는다' .moai/docs/factory-managed-session.md` | `0`, exit 1 |
| `grep -c 'SPEC-FACTORY-MANAGED-TUI-001' .moai/docs/factory-managed-session.md` | 1 or more |
| `grep -c 'MOAI_FACTORY_MANAGED_TUI' .moai/docs/factory-managed-session.md` | 1 or more |
| `grep -c 'factory-managed-' .moai/docs/factory-managed-session.md` | 1 or more (log file name) |
| `grep -c -e '--remote' .moai/docs/factory-managed-session.md` | 1 or more (probe) |
| `grep -c '/quit' .moai/docs/factory-managed-session.md` | 1 or more |
| `grep -c 't1459' .moai/docs/factory-managed-session.md` | 1 or more (signals still open) |
| `grep -c 'AC-MT-015' .moai/docs/factory-managed-session.md` | 1 or more |
| `grep -c '미관측' .moai/docs/factory-managed-session.md` | 3 or more |
| `grep -c 'SPEC-FACTORY-MANAGED-TUI-001' CHANGELOG.md` | 1 or more (baseline `0`, exit 1) |

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
| mu8 claim ignores busy | AC-MT-006 `…DefersDeliveryWhileBusy` |
| mu9 reader feeds the event channel with operator-turn frames | AC-MT-006 `…ReaderSurvivesOperatorTurns` |
| mu10 operator-turn request answered by the launcher | AC-MT-007 `operator_turn_unanswered` |
| mu11 scoping left on when detached | AC-MT-007 `detached_unchanged` |
| mu12 TUI exit status dropped to 0 | AC-MT-008 `exit_seven` |
| mu13 TUI not stopped at session end | AC-MT-009 |
| mu14 managed TUI reachable without the opt-in | AC-MT-010 `TestManagedTUINeverReachedWithoutOptIn` |

## §3. Edge cases

- Keystrokes typed during `Start` and priming wait in the tty buffer and reach the TUI.
- The TUI exits before the first claim: the session ends, no delivery was attempted.
- Bind failure or failing priming: the TUI is never started.
- A claimed batch and an operator turn starting in the race window: steered into the active turn (known debt 4); the lease (2 min) bounds the batch.
- Stdin at EOF with a terminal predicate true is not a terminal case; EOF only matters headless.
- A path with spaces in the codex binary: the exec form takes argv, no shell.

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

Failure disposition: any failed step becomes a named known debt or a follow-up card; step 2 or 3 failing means the opt-out should become opt-in.

## §5. Quality gates and Definition of Done

TRUST 5: new file coverage 85% or more by `go test -cover` on the touched tests; `gofmt -l internal/cli internal/config` empty; `go vet ./internal/cli ./internal/config`; `golangci-lint run ./internal/cli/... ./internal/config/...` at the CI version; managed top-level test count not below 79 plus the new tests. DoD: AC-MT-001..014 PASS with command and output; AC-MT-015 recorded as manual and either observed or listed as an open Gap; cross build and grep guards cited; CHANGELOG and operator document updated; the sync audit re-reads the premise ledger.
