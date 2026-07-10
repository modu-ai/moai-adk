# SPEC-CLIFIX-CONTRACT-001 — Acceptance Criteria

## §A Scenarios (Given-When-Then)

1. Given a CI pipeline invoking `moai spec lint --json --sarif` (invalid combination), When the command exits, Then the exit code is 3 and stderr explains the invalid arguments.
2. Given a pre-push hook rejecting a push, When git surfaces the hook output, Then the developer sees the violation reason (stderr) instead of a bare exit-2 block.
3. Given `moai github link-spec --dry-run`, When it completes, Then the registry file hash is unchanged and stdout lists the mutation that would have occurred.

## §D AC Matrix (machine-verifiable)

| AC | REQ | Verification command | Expected outcome |
|---|---|---|---|
| AC-CONT-001-001 | REQ-CONT-001-001 | `grep -rn 'os\.Exit' internal/cli internal/cli/harness internal/cli/agentlint --include='*.go' \| grep -v '_test.go'` | Only the approved-exception sites remain: cmd/moai/main.go ExitCoder boundary mapping + launch_exec_windows.go exec-mirror (documented process-replacement semantics); zero matches inside RunE/PostRunE bodies of the 8 listed files |
| AC-CONT-001-002 | REQ-CONT-001-002 | `go test ./internal/cli/ -run 'GithubDryRun' -count=1 -v` | PASS — registry file byte-identical after --dry-run; planned mutation printed |
| AC-CONT-001-003 | REQ-CONT-001-003 | `go test ./internal/cli/ -run 'SpecStatusConfirm' -count=1 -v` | PASS — non-TTY without --confirm aborts (no hang, no git mutation); with --confirm proceeds using `git -C <projectRoot>` |
| AC-CONT-001-004 | REQ-CONT-001-004 | `go test ./internal/cli/ -run 'AstgrepExitCode' -count=1 -v` | PASS — error findings produce exit 1 under text, json, and sarif formats alike |
| AC-CONT-001-005 | REQ-CONT-001-005 | `go test ./internal/cli/ -run 'ExitCodeContract' -count=1 -v` | PASS — constitution failure=2, spec lint invalid args=3, spec audit MUST-FIX=2 |
| AC-CONT-001-006 | REQ-CONT-001-006 | `go test ./internal/cli/ -run 'PrePushStderr' -count=1 -v` | PASS — on exit 2 the violation details are asserted present on stderr |
| AC-CONT-001-007 | REQ-CONT-001-007 | `ls internal/cli/team_spawn_lock_unix_test.go && go test ./internal/cli/ -run 'Flock\|ClaimLock' -count=1 -v` | File exists with `_test.go` suffix; test is discovered and PASSES; `go list -deps ./cmd/moai \| grep -c '^testing$'` returns 0 |
| AC-CONT-001-008 | REQ-CONT-001-008 | `go test ./internal/cli/... -run 'HelpExitContract' -count=1 -v` | PASS — for each changed command, declared exit codes in help text match produced codes |

## §C Edge Cases

- os.Exit removal must preserve cobra `SilenceUsage`/`SilenceErrors` behavior — no double-printed errors after the boundary mapping.
- astgrep with zero findings and `--json`: exit 0 with valid empty JSON document (contract unchanged).
- spec_status `--sync-git --confirm` in a repo where projectRoot != cwd: git commands run against projectRoot.
- pre-push with stdin read error (fail-open finding in the audit is a SEPARATE minor; do not silently change fail-open→fail-closed here — out of scope, note only).
- The renamed flock test on non-unix builds: file carries the unix build tag; Windows build stays green.

## §D.5 Quality Gate / Definition of Done

- All 8 AC rows PASS with verbatim command output cited in progress.md §E.2.
- `go build ./...` green on darwin/linux/windows (CI matrix); `go test ./internal/cli/... -count=1` green.
- `go vet ./...` and `golangci-lint run` introduce no new findings.
- Exit-code contract table (command → declared → produced) recorded in progress.md §E.2 as evidence.
