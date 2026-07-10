# SPEC-CLIFIX-CONTRACT-001 — Acceptance Criteria

## §A Scenarios (Given-When-Then)

1. Given a CI pipeline invoking `moai spec lint --json --sarif` (invalid combination), When the command exits, Then the exit code is 3 and stderr explains the invalid arguments.
2. Given a pre-push hook rejecting a push, When git surfaces the hook output, Then the developer sees the violation reason (stderr) instead of a bare exit-2 block.
3. Given `moai github link-spec --dry-run`, When it completes, Then the registry file hash is unchanged and stdout lists the mutation that would have occurred.

## §B AC ↔ REQ Traceability Summary

| AC | REQ | Scope |
|---|---|---|
| AC-CONT-001-001 | REQ-CONT-001-001 | os.Exit removal (11 sites / 8 files) + approved-exception classification |
| AC-CONT-001-002 | REQ-CONT-001-002 | github --dry-run wiring |
| AC-CONT-001-003 | REQ-CONT-001-003 | spec_status --yes gate + --confirm removal + non-TTY abort |
| AC-CONT-001-004 | REQ-CONT-001-004 | astgrep format-independent HasErrors exit 1 |
| AC-CONT-001-005 | REQ-CONT-001-005 | exit-code contracts (constitution=2, spec lint invalid=3, spec audit MUST-FIX=2) |
| AC-CONT-001-006 | REQ-CONT-001-006 | pre-push stderr routing |
| AC-CONT-001-007 | REQ-CONT-001-007 | flock test rename + testing-package unlink |
| AC-CONT-001-008 | REQ-CONT-001-008 | per-command exit-code contract tests |

All 8 REQs are covered by exactly one AC; no orphan ACs or uncovered REQs.

## §C Edge Cases

- os.Exit removal must preserve cobra `SilenceUsage`/`SilenceErrors` behavior — no double-printed errors after the boundary mapping.
- astgrep with zero findings and `--json`: exit 0 with valid empty JSON document (contract unchanged).
- spec_status `--sync-git --yes` in a repo where projectRoot != cwd: git commands run against projectRoot.
- pre-push with stdin read error (fail-open finding in the audit is a SEPARATE minor; do not silently change fail-open→fail-closed here — out of scope, note only).
- The renamed flock test on non-unix builds: file carries the unix build tag; Windows build stays green.
- hook_pre_push.go:196 removal: the inline comment at :194 ("This is the ONLY os.Exit site") becomes stale after the ExitCoder conversion and must be updated in the same run-phase edit (the `decideExit` purity claim at :102-103 stays valid — it is the boundary that changes, not the pure function).
- Dead-flag removal of `--confirm`: check whether any script, test, or doc references `--confirm` before deleting; if a caller exists, alias `--confirm` to `--yes` for one release instead of hard-removing.

## §D AC Matrix (machine-verifiable)

| AC | REQ | Verification command | Expected outcome |
|---|---|---|---|
| AC-CONT-001-001 | REQ-CONT-001-001 | `grep -rn 'os\.Exit' internal/cli internal/cli/harness internal/cli/agentlint --include='*.go' \| grep -v '_test.go' \| grep -vE ':[0-9]+:[[:space:]]*//'` | Only the approved-exception sites remain WITHIN the grep scope: `launch_exec_windows.go:37,41` + `update.go:487` (Windows process-replacement — `syscall.Exec` unavailable on Windows, `os.Exit` after child re-exec is the standard boundary). Zero matches inside RunE/PostRunE bodies of the 8 listed files. (`cmd/moai/main.go` ExitCoder boundary is OUT of this grep's scope — `internal/cli` only — and is verified separately by AC-CONT-001-008.) The comment-exclusion filter (`grep -vE ':[0-9]+:[[:space:]]*//'`) is required because `hook_pre_push.go` contains literal "os.Exit" substrings inside comments at lines 51, 102-103, and 194. |
| AC-CONT-001-002 | REQ-CONT-001-002 | `go test ./internal/cli/ -run 'GithubDryRun' -count=1 -v` | PASS — registry file byte-identical after --dry-run; planned mutation printed |
| AC-CONT-001-003 | REQ-CONT-001-003 | `go test ./internal/cli/ -run 'SpecStatusConfirm' -count=1 -v` | PASS — non-TTY without `--yes` aborts (no hang, no git mutation); with `--yes` proceeds using `git -C <projectRoot>`; `--confirm` is removed (or aliased to `--yes` for one release if a caller exists) |
| AC-CONT-001-004 | REQ-CONT-001-004 | `go test ./internal/cli/ -run 'AstgrepExitCode' -count=1 -v` | PASS — error findings produce exit 1 under text, json, and sarif formats alike |
| AC-CONT-001-005 | REQ-CONT-001-005 | `go test ./internal/cli/ -run 'ExitCodeContract' -count=1 -v` | PASS — constitution failure=2, spec lint invalid args=3, spec audit MUST-FIX=2 |
| AC-CONT-001-006 | REQ-CONT-001-006 | `go test ./internal/cli/ -run 'PrePushStderr' -count=1 -v` | PASS — on exit 2 the violation details are asserted present on stderr |
| AC-CONT-001-007 | REQ-CONT-001-007 | `ls internal/cli/team_spawn_lock_unix_test.go && go test ./internal/cli/ -run 'Flock\|ClaimLock' -count=1 -v` | File exists with `_test.go` suffix; test is discovered and PASSES; `go list -deps ./cmd/moai \| grep -c '^testing$'` returns 0 |
| AC-CONT-001-008 | REQ-CONT-001-008 | `go test ./internal/cli/... -run 'HelpExitContract' -count=1 -v` | PASS — for each changed command, declared exit codes in help text match produced codes |

## §D.5 Quality Gate / Definition of Done

- All 8 AC rows PASS with verbatim command output cited in progress.md §E.2.
- `go build ./...` green on darwin/linux/windows (CI matrix); `go test ./internal/cli/... -count=1` green.
- `go vet ./...` and `golangci-lint run` introduce no new findings.
- Exit-code contract table (command → declared → produced) recorded in progress.md §E.2 as evidence.
