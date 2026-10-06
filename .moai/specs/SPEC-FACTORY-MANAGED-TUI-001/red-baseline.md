# red-baseline.md — SPEC-FACTORY-MANAGED-TUI-001 (card t1408)

RED-now baseline of the adoption-deferred acceptance criteria (AC-MT-001..009, 011, 016): every command was run in this worktree on the stub commit S (`d222d310e`, behavior-neutral compile stubs) with the reproduction tests of the R commit applied, before any fix. Each test fails on a **named assertion** (a `--- FAIL:` line of a test or subtest), not on a package build error, a crash or a timeout. Tree for every row: S = `d222d310e` + the R tests (R's own tree depends on this file and cannot be cited).

Command-form deviation, stated: the acceptance.md commands end the `-run` pattern with a dollar sign. The lane's worktree guard refuses a command carrying that character, so the patterns below are prefix-anchored (`^Name`). Uniqueness was read from `go test ./internal/cli -list TestManagedCodex` and `-list TestManagedTUI` (all 15 names listed once; no other test shares one of these prefixes), so the swept set equals the intended set. Alternation blocks (AC-MT-006, AC-MT-009) are run as one invocation per test name for the same reason.

Output below is the filtered log of each run: the `--- FAIL`/`--- PASS` lines, the first line of every assertion message (`managed_codex_tui_test.go:<line>:`, multi-line dumps cut to that line), and the summary line. Full raw logs are local scratch and are not cited.

## Summary table

| id | command | exit | named failing tests |
|---|---|---|---|
| AC-MT-001 | `go test -race ./internal/cli -run '^TestManagedCodexTUIAttachCommand' -count=1 -v` | 1 | TestManagedCodexTUIAttachCommand |
| AC-MT-002 | `go test -race ./internal/cli -run '^TestManagedCodexTUIStartsAfterPrimingAndBind' -count=1 -v` | 1 | TestManagedCodexTUIStartsAfterPrimingAndBind |
| AC-MT-003 | `go test -race ./internal/cli -run '^TestManagedCodexTUIPreconditionsAndFallback' -count=1 -v` | 1 | TestManagedCodexTUIPreconditionsAndFallback |
| AC-MT-003 | `go test ./internal/cli -run '^TestManagedCodexRemoteSupportProbe' -count=1 -v` | 1 | TestManagedCodexRemoteSupportProbe |
| AC-MT-004 | `go test -race ./internal/cli -run '^TestManagedCodexTUIOwnsOperatorStdin' -count=1 -v` | 1 | TestManagedCodexTUIOwnsOperatorStdin |
| AC-MT-005 | `go test -race ./internal/cli -run '^TestManagedCodexTUIKeepsTerminalClean' -count=1 -v` | 1 | TestManagedCodexTUIKeepsTerminalClean |
| AC-MT-006 | `go test -race ./internal/cli -run '^TestManagedCodexTUIDefersDeliveryWhileBusy' -count=1 -v` | 1 | TestManagedCodexTUIDefersDeliveryWhileBusy |
| AC-MT-006 | `go test -race ./internal/cli -run '^TestManagedCodexReaderSurvivesOperatorTurns' -count=1 -v` | 1 | TestManagedCodexReaderSurvivesOperatorTurns |
| AC-MT-006 | `go test -race ./internal/cli -run '^TestManagedCodexBusyLongTurnKeepsDeferring' -count=1 -v` | 1 | TestManagedCodexBusyLongTurnKeepsDeferring |
| AC-MT-007 | `go test -race ./internal/cli -run '^TestManagedCodexServerRequestScopingWithTUI' -count=1 -v` | 1 | TestManagedCodexServerRequestScopingWithTUI |
| AC-MT-008 | `go test -race ./internal/cli -run '^TestManagedCodexTUIExitEndsSession' -count=1 -v` | 1 | TestManagedCodexTUIExitEndsSession |
| AC-MT-009 | `go test -race ./internal/cli -run '^TestManagedCodexServerDeathStopsTUI' -count=1 -v` | 1 | TestManagedCodexServerDeathStopsTUI |
| AC-MT-009 | `go test -race ./internal/cli -run '^TestManagedCodexSessionEndStopsTUI' -count=1 -v` | 1 | TestManagedCodexSessionEndStopsTUI |
| AC-MT-011 | `go test -race ./internal/cli -run '^TestManagedCodexTUILoopbackRoundTrip' -count=1 -v` | 1 | TestManagedCodexTUILoopbackRoundTrip |
| AC-MT-016 | `go test ./internal/cli -run '^TestManagedTUINeverReachedWithoutOptIn' -count=1 -v` | 1 | TestManagedTUINeverReachedWithoutOptIn |

## AC-MT-001 — `'^TestManagedCodexTUIAttachCommand'`

Command: `go test -race ./internal/cli -run '^TestManagedCodexTUIAttachCommand' -count=1 -v`  
Exit code: 1  
Tree: S `d222d310e` + R tests

```text
    managed_codex_tui_test.go:888: the TUI child never started within 10s (no tui-argv line in the fake log)
    managed_codex_tui_test.go:899: tui-argv lines = 0, want 1
    managed_codex_tui_test.go:922: no TUI command line was logged, so its token handling is unobserved
    managed_codex_tui_test.go:929: TUI env token [], App Server token ["bb7dde8007905d1b2a473e28e0733ad8c787ac2dac421d5b33e6a0a541f6b6f3"]: want the same non-empty value
    managed_codex_tui_test.go:941: no TUI command line was logged
    managed_codex_tui_test.go:957: TUI argv misses the operator argument "model_reasoning_effort=\\\"high\\\"":
    managed_codex_tui_test.go:957: TUI argv misses the operator argument "\"-m\",\"fake-model\"":
    managed_codex_tui_test.go:964: tui-cwd [], thread-cwd ["/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestManagedCodexTUIAttachCommand3176398292/005"]: want one of each
--- FAIL: TestManagedCodexTUIAttachCommand (10.50s)
    --- FAIL: TestManagedCodexTUIAttachCommand/argv_exact (0.00s)
    --- FAIL: TestManagedCodexTUIAttachCommand/token_not_in_argv (0.00s)
    --- FAIL: TestManagedCodexTUIAttachCommand/token_in_env (0.00s)
    --- PASS: TestManagedCodexTUIAttachCommand/app_server_keeps_token_file (0.00s)
    --- FAIL: TestManagedCodexTUIAttachCommand/no_generated_approval_args (0.00s)
    --- FAIL: TestManagedCodexTUIAttachCommand/operator_args_forwarded (0.00s)
    --- FAIL: TestManagedCodexTUIAttachCommand/cwd_matches_thread (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	11.895s
FAIL
```

## AC-MT-002 — `'^TestManagedCodexTUIStartsAfterPrimingAndBind'`

Command: `go test -race ./internal/cli -run '^TestManagedCodexTUIStartsAfterPrimingAndBind' -count=1 -v`  
Exit code: 1  
Tree: S `d222d310e` + R tests

```text
    managed_codex_tui_test.go:982: the TUI never started within 10s
    managed_codex_tui_test.go:989: tui-start (offset -1) must follow the priming turn/completed (offset 906)
    managed_codex_tui_test.go:993: at TUI start the broker endpoint was bound to [], want "fake-thread-1"
--- FAIL: TestManagedCodexTUIStartsAfterPrimingAndBind (11.36s)
    --- FAIL: TestManagedCodexTUIStartsAfterPrimingAndBind/order (10.42s)
    --- PASS: TestManagedCodexTUIStartsAfterPrimingAndBind/priming_failure_never_attaches (0.94s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	12.527s
FAIL
```

## AC-MT-003 — `'^TestManagedCodexTUIPreconditionsAndFallback'`

Command: `go test -race ./internal/cli -run '^TestManagedCodexTUIPreconditionsAndFallback' -count=1 -v`  
Exit code: 1  
Tree: S `d222d310e` + R tests

```text
    managed_codex_tui_test.go:1114: stderr holds 0 fallback notices, want exactly 1:
    managed_codex_tui_test.go:1114: stderr holds 0 fallback notices, want exactly 1:
    managed_codex_tui_test.go:1114: stderr holds 0 fallback notices, want exactly 1:
    managed_codex_tui_test.go:1114: stderr holds 0 fallback notices, want exactly 1:
    managed_codex_tui_test.go:1114: stderr holds 0 fallback notices, want exactly 1:
    managed_codex_tui_test.go:1114: stderr holds 0 fallback notices, want exactly 1:
    managed_codex_tui_test.go:1114: stderr holds 0 fallback notices, want exactly 1:
    managed_codex_tui_test.go:1114: stderr holds 0 fallback notices, want exactly 1:
    managed_codex_tui_test.go:1114: stderr holds 0 fallback notices, want exactly 1:
    managed_codex_tui_test.go:1114: stderr holds 0 fallback notices, want exactly 1:
    managed_codex_tui_test.go:1114: stderr holds 0 fallback notices, want exactly 1:
    managed_codex_tui_test.go:1119: after a failed TUI start the App Server stderr must be in the session log file "":
    managed_codex_tui_test.go:1125: the log file path was printed 0 times, want once:
--- FAIL: TestManagedCodexTUIPreconditionsAndFallback (10.45s)
    --- FAIL: TestManagedCodexTUIPreconditionsAndFallback/stdin_not_terminal (0.95s)
    --- FAIL: TestManagedCodexTUIPreconditionsAndFallback/stdin_not_file (0.72s)
    --- FAIL: TestManagedCodexTUIPreconditionsAndFallback/stdout_not_terminal (0.75s)
    --- FAIL: TestManagedCodexTUIPreconditionsAndFallback/probe_lacks_options (0.97s)
    --- FAIL: TestManagedCodexTUIPreconditionsAndFallback/probe_times_out (0.79s)
    --- FAIL: TestManagedCodexTUIPreconditionsAndFallback/probe_fails_to_run (0.74s)
    --- FAIL: TestManagedCodexTUIPreconditionsAndFallback/opt_out_0 (0.99s)
    --- FAIL: TestManagedCodexTUIPreconditionsAndFallback/opt_out_false (1.02s)
    --- FAIL: TestManagedCodexTUIPreconditionsAndFallback/opt_out_off (1.08s)
    --- FAIL: TestManagedCodexTUIPreconditionsAndFallback/opt_out_uppercase_padded (1.05s)
    --- FAIL: TestManagedCodexTUIPreconditionsAndFallback/tui_start_fails (1.40s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	11.457s
FAIL
```

## AC-MT-003 — `'^TestManagedCodexRemoteSupportProbe'`

Command: `go test ./internal/cli -run '^TestManagedCodexRemoteSupportProbe' -count=1 -v`  
Exit code: 1  
Tree: S `d222d310e` + R tests

```text
    managed_codex_tui_test.go:1025: the vendored codex-cli 0.160.0 resume help must be read as supporting --remote and --remote-auth-token-env
--- FAIL: TestManagedCodexRemoteSupportProbe (0.00s)
    --- FAIL: TestManagedCodexRemoteSupportProbe/real_help_supported (0.00s)
    --- PASS: TestManagedCodexRemoteSupportProbe/no_options_unsupported (0.00s)
    --- PASS: TestManagedCodexRemoteSupportProbe/empty_unsupported (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.974s
FAIL
```

## AC-MT-004 — `'^TestManagedCodexTUIOwnsOperatorStdin'`

Command: `go test -race ./internal/cli -run '^TestManagedCodexTUIOwnsOperatorStdin' -count=1 -v`  
Exit code: 1  
Tree: S `d222d310e` + R tests

```text
    managed_codex_tui_test.go:1138: the TUI never started within 10s
    managed_codex_tui_test.go:1144: the session ended on /exit while the TUI owns the terminal
    managed_codex_tui_test.go:1159: the fake TUI did not read "tui-stdin hello" from the shared stdin:
    managed_codex_tui_test.go:1159: the fake TUI did not read "tui-stdin /exit" from the shared stdin:
    managed_codex_tui_test.go:1163: the launcher turned an operator stdin line into a turn while the TUI is attached
--- FAIL: TestManagedCodexTUIOwnsOperatorStdin (12.37s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	13.611s
FAIL
```

## AC-MT-005 — `'^TestManagedCodexTUIKeepsTerminalClean'`

Command: `go test -race ./internal/cli -run '^TestManagedCodexTUIKeepsTerminalClean' -count=1 -v`  
Exit code: 1  
Tree: S `d222d310e` + R tests

```text
    managed_codex_tui_test.go:1175: the TUI never started within 10s
    managed_codex_tui_test.go:1192: no session log file exists under .moai/logs
    managed_codex_tui_test.go:1203: terminal stderr must hold exactly the one line naming the log file, got:
    managed_codex_tui_test.go:1219: the App Server child's stderr must stay in the file
--- FAIL: TestManagedCodexTUIKeepsTerminalClean (15.91s)
    --- FAIL: TestManagedCodexTUIKeepsTerminalClean/attached_lines_in_file (0.00s)
    --- FAIL: TestManagedCodexTUIKeepsTerminalClean/terminal_only_has_path_line (0.00s)
    --- FAIL: TestManagedCodexTUIKeepsTerminalClean/post_reap_line_on_terminal (0.01s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	17.128s
FAIL
```

## AC-MT-006 — `'^TestManagedCodexTUIDefersDeliveryWhileBusy'`

Command: `go test -race ./internal/cli -run '^TestManagedCodexTUIDefersDeliveryWhileBusy' -count=1 -v`  
Exit code: 1  
Tree: S `d222d310e` + R tests

```text
    managed_codex_tui_test.go:1236: the TUI never started within 10s
    managed_codex_tui_test.go:1241: the owner never saw the operator-started turn as busy
    managed_codex_tui_test.go:1251: a launcher turn started while the operator turn ran (prompts: 2)
--- FAIL: TestManagedCodexTUIDefersDeliveryWhileBusy (15.35s)
    --- FAIL: TestManagedCodexTUIDefersDeliveryWhileBusy/deferred_then_delivered (14.79s)
    --- PASS: TestManagedCodexTUIDefersDeliveryWhileBusy/frames_not_delivered (0.55s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	16.452s
FAIL
```

## AC-MT-006 — `'^TestManagedCodexReaderSurvivesOperatorTurns'`

Command: `go test -race ./internal/cli -run '^TestManagedCodexReaderSurvivesOperatorTurns' -count=1 -v`  
Exit code: 1  
Tree: S `d222d310e` + R tests

```text
    managed_codex_tui_test.go:1288: the TUI never started within 10s
    managed_codex_tui_test.go:1299: a server request after 40 operator turns was not answered within 5s: the reader stalled behind the event channel
--- FAIL: TestManagedCodexReaderSurvivesOperatorTurns (16.21s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	17.320s
FAIL
```

## AC-MT-006 — `'^TestManagedCodexBusyLongTurnKeepsDeferring'`

Command: `go test -race ./internal/cli -run '^TestManagedCodexBusyLongTurnKeepsDeferring' -count=1 -v`  
Exit code: 1  
Tree: S `d222d310e` + R tests

```text
    managed_codex_tui_test.go:1313: the TUI never started within 10s
    managed_codex_tui_test.go:1325: a busy turn that outlasts the warn interval must log one line per elapsed interval (want at least 2):
--- FAIL: TestManagedCodexBusyLongTurnKeepsDeferring (19.53s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	20.882s
FAIL
```

## AC-MT-007 — `'^TestManagedCodexServerRequestScopingWithTUI'`

Command: `go test -race ./internal/cli -run '^TestManagedCodexServerRequestScopingWithTUI' -count=1 -v`  
Exit code: 1  
Tree: S `d222d310e` + R tests

```text
    managed_codex_tui_test.go:1355: the session did not attach the TUI
    managed_codex_tui_test.go:1378: item/commandExecution/requestApproval naming an operator-started turn was answered by the launcher: {"id":"op-req-0","result":{"decision":"decline"}}
    managed_codex_tui_test.go:1378: item/fileChange/requestApproval naming an operator-started turn was answered by the launcher: {"id":"op-req-1","result":{"decision":"decline"}}
    managed_codex_tui_test.go:1378: item/tool/requestUserInput naming an operator-started turn was answered by the launcher: {"error":{"code":-32000,"message":"managed Factory session cannot answer \"item/tool/requestUserInput\": no operator is attached"},"id":"op-req-2"}
    managed_codex_tui_test.go:1378: mcpServer/elicitation/request naming an operator-started turn was answered by the launcher: {"id":"op-req-3","result":{"action":"decline"}}
    managed_codex_tui_test.go:1378: item/permissions/requestApproval naming an operator-started turn was answered by the launcher: {"id":"op-req-4","result":{"permissions":{}}}
    managed_codex_tui_test.go:1378: item/tool/call naming an operator-started turn was answered by the launcher: {"id":"op-req-5","result":{"contentItems":[{"text":"managed Factory session registers no dynamic tools","type":"inputText"}],"success":false}}
--- FAIL: TestManagedCodexServerRequestScopingWithTUI (2.20s)
    --- PASS: TestManagedCodexServerRequestScopingWithTUI/owned_turn_answered (0.12s)
    --- FAIL: TestManagedCodexServerRequestScopingWithTUI/operator_turn_unanswered (0.51s)
    --- PASS: TestManagedCodexServerRequestScopingWithTUI/armed_window_answered (0.15s)
    --- PASS: TestManagedCodexServerRequestScopingWithTUI/no_turn_id_kinds_answered (0.08s)
    --- PASS: TestManagedCodexServerRequestScopingWithTUI/never_accepts (0.00s)
    --- PASS: TestManagedCodexServerRequestScopingWithTUI/detached_unchanged (0.54s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	3.197s
FAIL
```

## AC-MT-008 — `'^TestManagedCodexTUIExitEndsSession'`

Command: `go test -race ./internal/cli -run '^TestManagedCodexTUIExitEndsSession' -count=1 -v`  
Exit code: 1  
Tree: S `d222d310e` + R tests

```text
    managed_codex_tui_test.go:1453: the TUI never started within 10s
    managed_codex_tui_test.go:1453: the TUI never started within 10s
    managed_codex_tui_test.go:1487: TUI exit 7 must surface as ExitCode 7, got <nil> (code 0, coder=false)
    managed_codex_tui_test.go:1453: the TUI never started within 10s
    managed_codex_tui_test.go:1494: a TUI ended by a signal must surface as ExitCode 1, got <nil> (code 0, coder=false)
    managed_codex_tui_test.go:1515: the TUI never started, so the precedence rule is unobserved
    managed_codex_tui_test.go:1545: the TUI never started within 10s
    managed_codex_tui_test.go:1555: barrier: the launcher's turn/start write never entered write()
    managed_codex_tui_test.go:1563: barrier: the second sender never entered write()
    managed_codex_tui_test.go:1571: the TUI exit did not release the blocked turn/start write within 5s
    managed_codex_tui_test.go:1595: the TUI exit did not release the waiting turn within 5s
--- FAIL: TestManagedCodexTUIExitEndsSession (66.69s)
    --- FAIL: TestManagedCodexTUIExitEndsSession/exit_zero (10.45s)
    --- FAIL: TestManagedCodexTUIExitEndsSession/exit_seven (10.33s)
    --- FAIL: TestManagedCodexTUIExitEndsSession/signaled (10.36s)
    --- FAIL: TestManagedCodexTUIExitEndsSession/session_error_wins (1.49s)
    --- FAIL: TestManagedCodexTUIExitEndsSession/blocked_turn_start_released (28.21s)
    --- FAIL: TestManagedCodexTUIExitEndsSession/blocked_wait_turn_released (5.85s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	67.766s
FAIL
```

## AC-MT-009 — `'^TestManagedCodexServerDeathStopsTUI'`

Command: `go test -race ./internal/cli -run '^TestManagedCodexServerDeathStopsTUI' -count=1 -v`  
Exit code: 1  
Tree: S `d222d310e` + R tests

```text
    managed_codex_tui_test.go:1615: the TUI never started within 10s
    managed_codex_tui_test.go:1622: the owner did not return after the App Server connection closed
    managed_codex_tui_test.go:1630: the TUI child (pid 0) outlived the owner's return
    managed_codex_tui_test.go:1634: a TUI that handles the interrupt was not interrupted
    managed_codex_tui_test.go:1615: the TUI never started within 10s
    managed_codex_tui_test.go:1622: the owner did not return after the App Server connection closed
    managed_codex_tui_test.go:1630: the TUI child (pid 0) outlived the owner's return
--- FAIL: TestManagedCodexServerDeathStopsTUI (40.83s)
    --- FAIL: TestManagedCodexServerDeathStopsTUI/handles_interrupt (20.41s)
    --- FAIL: TestManagedCodexServerDeathStopsTUI/ignores_interrupt (20.42s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	41.790s
FAIL
```

## AC-MT-009 — `'^TestManagedCodexSessionEndStopsTUI'`

Command: `go test -race ./internal/cli -run '^TestManagedCodexSessionEndStopsTUI' -count=1 -v`  
Exit code: 1  
Tree: S `d222d310e` + R tests

```text
    managed_codex_tui_test.go:1669: the TUI child (pid 0) is still alive after Close returned
    managed_codex_tui_test.go:1673: a TUI that handles the interrupt was not interrupted at session end
    managed_codex_tui_test.go:1669: the TUI child (pid 0) is still alive after Close returned
    managed_codex_tui_test.go:1676: a TUI that ignores the interrupt must be killed only after the grace (elapsed 3.149916ms, interrupted=false)
--- FAIL: TestManagedCodexSessionEndStopsTUI (1.81s)
    --- FAIL: TestManagedCodexSessionEndStopsTUI/handles_interrupt (0.82s)
    --- FAIL: TestManagedCodexSessionEndStopsTUI/ignores_interrupt (0.99s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	3.215s
FAIL
```

## AC-MT-011 — `'^TestManagedCodexTUILoopbackRoundTrip'`

Command: `go test -race ./internal/cli -run '^TestManagedCodexTUILoopbackRoundTrip' -count=1 -v`  
Exit code: 1  
Tree: S `d222d310e` + R tests

```text
    managed_codex_tui_test.go:1690: the fake TUI never attached with the env token and its own WebSocket connection
    managed_codex_tui_test.go:1721: the injected message was never acknowledged through the attached session
    managed_codex_tui_test.go:1726: the session did not end after the fake TUI exited
    managed_codex_tui_test.go:1733: acknowledged=0 pending=0, want 1/0
    managed_codex_tui_test.go:1736: the model did not read the body by claim token:
    managed_codex_tui_test.go:1744: teardown incomplete: App Server pid 32761 alive
--- FAIL: TestManagedCodexTUILoopbackRoundTrip (50.46s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	52.780s
FAIL
```

## AC-MT-016 — `'^TestManagedTUINeverReachedWithoutOptIn'`

Command: `go test ./internal/cli -run '^TestManagedTUINeverReachedWithoutOptIn' -count=1 -v`  
Exit code: 1  
Tree: S `d222d310e` + R tests

```text
    managed_codex_tui_test.go:1824: probe=0 tui spawns=0 direct door=0, want 1/0/0
--- FAIL: TestManagedTUINeverReachedWithoutOptIn (0.05s)
    --- PASS: TestManagedTUINeverReachedWithoutOptIn/no_switch (0.00s)
    --- PASS: TestManagedTUINeverReachedWithoutOptIn/switch_only (0.00s)
    --- PASS: TestManagedTUINeverReachedWithoutOptIn/stamps_only (0.00s)
    --- FAIL: TestManagedTUINeverReachedWithoutOptIn/opted_in_probes_once (0.05s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.088s
FAIL
```

