auditor-model: claude-sonnet-5-5
verdict: FAIL
audited_sha: dd20c7ae0c2eab562920bb31c33bff93abebb10c

# t1866 sync audit, round 7 (delta dd20c7ae0 over 18e387a9b7)

Score: Functionality 78, Security 95, Craft 88, Consistency 92. Overall FAIL: the required codex gate returned fail, and my own reading independently found the same defect (F15).

## Claim

C1. The F14 repair works for the stated paths: an unset profile on `--continue`, `-c` or `DO_CLAUDE_CONTINUE=true` (Claude backend) passes no `--permission-mode`; a hand-written `defaultMode` is neither rewritten nor overridden.
C2. Earlier behaviour holds: GLM never receives `auto` from a profile/settings default; stored profile modes and typed flags are forwarded on a resume; a new unset Claude session still gets `auto` on argv and in the `settings.local.json` sync.
C3. Tests exercise the real child start and are meaningful.
C4. CHANGELOG and progress.md claims are true of code and tests.
C5. NOT TRUE: every path that starts a NEW Claude session receives the default. A resume whose child exits 1 ("No previous session found, starting new session...") starts a new session without `--permission-mode auto` (F15).

## Evidence

Launcher after repair (`internal/cli/launcher.go`):

```
688:	cont := settings["DO_CLAUDE_CONTINUE"] == "true"
699:	consumed := consumeLaunchArgs(extraArgs)
705:	if consumed.Continue {
718:	if permMode == "" && !cont && !consumed.PermissionModeSet {
719:		permMode = launchDefaultPermissionMode(permGLM)
728:	if !cont && settingsLocalPath != "" {   // sync, uses syncPermMode computed at 670-687
880:	if cont {
881:		tryCmd := exec.Command(claudeBin, buildArgs(true)[1:]...)
893:		if errors.As(err, &ee) && ee.ExitCode() == 1 {
894:			fmt.Fprintln(os.Stderr, "No previous session found, starting new session...")
907:	return execOrSpawnClaudeFunc(claudeBin, buildArgs(false), launchEnv)
```

Ordering checks (read from code):
- GLM drop of a stored/settings `auto` (681-687) precedes typed-argv consumption (699-704), so a typed `--permission-mode auto` keeps glm.go's refusal. Held.
- `launchDefaultPermissionMode(true)` returns "" (line 1090-1095), so the late default never yields `auto` on GLM. Held.
- `DO_CLAUDE_BYPASS` legacy path sets `permMode` before the default, so `permMode != ""` and the default is skipped; `syncPermMode` stays `prefs.PermissionMode` (673-675). Held (test `...LegacyBypass...` PASS).
- `syncPermMode` is computed independently of `permMode` (670), so a NEW unset session still syncs `auto` (728-732); a resume skips the sync (`!cont`). Held.
- GLM resume drop and its failing error (737-741) are unchanged. Held.
- Exec seam vs `tryCmd.Run()`: only the `cont` branch uses `tryCmd`; fresh launches use the seam at 907. The new tests start the `--continue` child through `tryCmd.Run()`.

Scoped tests (each prefixed with the `unset MOAI_KANBAN_*` trio):

```
go test ./internal/cli -run '^TestLaunch_' -count=1 -v  -> exit=0; 29 "--- PASS", 0 "--- FAIL", 0 "--- SKIP"; last lines: PASS / ok  github.com/modu-ai/moai-adk/internal/cli	3.423s
go test ./internal/settings ./internal/cli/wizard ./internal/core/project -count=1
ok  	github.com/modu-ai/moai-adk/internal/settings	0.437s
ok  	github.com/modu-ai/moai-adk/internal/cli/wizard	4.787s
ok  	github.com/modu-ai/moai-adk/internal/core/project	1.828s
--- PASS: TestLaunch_UnsetPermissionModeDefaultsToAutoOnClaudeBackend (0.00s)
--- PASS: TestLaunch_UnsetPermissionModeStaysEmptyOnGLMBackend (0.00s)
--- PASS: TestLaunch_ClaudeResumeUnsetProfileAddsNoDefaultPermissionFlag (0.66s)
--- PASS: TestLaunch_ClaudeResumeUnsetProfileFollowsHandWrittenDefaultMode (0.21s)
```

Test review (judged from the bodies in `internal/cli/launcher_permission_resume_test.go`):
- `pinArgsRecordingClaude` writes the child argv to a file in its own temp dir; `recordedChildArgs` fails the test when the stub never ran, so an absent flag is never read from a child that did not start. The resume tests use `launchClaudeDefault` with `--continue`, i.e. `tryCmd.Run()`, not the exec seam. Meaningful.
- RED reasoning: before the repair the default was applied at the old line ~670 before argv consumption, so a resume with an unset profile carried `--permission-mode auto`; `hasArg(args, "--permission-mode")` therefore fails on the pre-repair code. The RED output quoted in progress.md matches that reasoning (`[--permission-mode auto --continue --model opus[1m]]`). I did not run the mutation myself.
- Positive controls: `TestLaunch_ClaudeResumeStillForwardsChosenPermissionMode` (stored auto, typed plan) and `TestLaunch_UnsetPermissionModeDefaultsToAutoOnClaudeBackend` prove the flag is observable when it should be there. No skip, no weakened assertion found.
- Coverage hole: no test for the resume-exits-1 fallback path (F15).

Required gate, `mcp__moai__audit_multi` (one call, project_root = worktree toplevel):

```
overall_verdict: fail
gate_unmet: (absent in result)
plan_source: config
per_backend_verdicts: backend=codex gate=required verdict=fail
audit_receipt: rcpt-c68710bc814236dc0a67251f
residual_risk_note: required-backend FAIL: codex
build_commit: ee981a932 (installed binary lag noted by the tool; see Baseline-attribution)
codex finding [P2] internal/cli/launcher.go:718: when the profile mode is empty and the --continue child exits 1, the launcher starts a new session, but cont stays true so the default is not applied; reproduction showed no permission flag in the new-session argv and the existing defaultMode: plan kept. Same reproduction on the pre-change launcher passed --permission-mode auto. Confidence high.
```

My independent confirmation (code reading): line 893-895 falls through to line 907 `buildArgs(false)`, and `permMode` was fixed empty at line 718 because `cont` was true; the sync at 728 was also skipped. So the fallback new session gets neither the `auto` flag nor the `auto` settings sync. This verdict is not final until the orchestrator's `convergence_check` passes.

CHANGELOG / progress.md review:
- CHANGELOG sentence "a resume ... passes no default flag for a profile that never chose a mode (a stored profile mode, `auto` included, and a typed `--permission-mode` are still forwarded on a resume)" is true of the code and of the three test groups.
- CHANGELOG "launches new Claude sessions in `auto`" is overstated for the resume-fallback new session (F15); every other sentence checked is true.
- progress.md "Audit repair round 4": the RED output, the 29 PASS / 0 FAIL / 0 SKIP count (I reproduced 29/0/0), the unchanged-items list and the Gaps are accurate. It does not mention the fallback path, which is the gap.

## Baseline-attribution

- Tree: worktree `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1866`, HEAD `dd20c7ae0c2eab562920bb31c33bff93abebb10c` (`git rev-parse HEAD`, this run).
- Tests were run with the Go toolchain against this tree (source build, no installed moai binary involved).
- `moai verify audit-plan` and `audit_multi` used the installed/running moai build `ee981a932` (MCP server pid 15518, built 2026-10-15), reported by the tool as an ancestor of HEAD `dd20c7ae0`. These tools judged policy and the codex backend review, not the Go test results; the audit-plan output was `config_status: ok`, `codex` required, `cross_model_active: true`.

## Gaps

- The mutation check (reverting the launcher condition) was not performed; RED is judged by reasoning plus the quoted RED output in progress.md (named, not observed by me).
- No whole-package run of `internal/cli`, `internal/web`, `internal/template` or `internal/config` (per instruction); no live Claude process (stub only).
- My first Bash call combining `git rev-parse --show-toplevel` with other commands was refused by the worktree guard ("names git in a form too complex to verify"); fallback: plain separate commands. No guard record was touched.
- The codex reproduction of F15 was reported by the codex leg; I confirmed it by reading the code, not by running a new test.
- `gate_unmet` was not present in the audit_multi output; the orchestrator's `moai verify audit-plan --result-file` check was not run by me (read-only role).

## Residual-risk

- After F15 is fixed, `moai cc -c` in a project with no prior session is the common trigger; the sync of `auto` into `settings.local.json` is also skipped on that path (the fallback arrives with `cont` true), so a fix must decide whether the fallback also syncs.
- A resume whose child exits 1 for a reason other than "no previous session" is treated the same by existing code; not in scope.
- Stub-based tests do not prove how a real Claude process interprets an absent flag next to `defaultMode`.

## Findings

- F15 [SHOULD-FIX, treated as BLOCKING because the required codex gate returned fail] `internal/cli/launcher.go:718` and `:893-907`. A resume (`-c`, `--continue`, `DO_CLAUDE_CONTINUE`) with an unset Claude-backend profile whose child exits 1 prints "No previous session found, starting new session..." and starts a NEW session with `buildArgs(false)` carrying no `--permission-mode auto`, and the `settings.local.json` sync was skipped. This contradicts the operator decision (auto for new sessions) and the CHANGELOG sentence. Required fix: just before the fallback exec at line 907, when the profile mode is still unset and no `--permission-mode` was typed (`!consumed.PermissionModeSet`, original `permMode == ""`), re-apply `launchDefaultPermissionMode(permGLM)` to `permMode` (GLM stays empty) and run the same settings.local.json sync as a fresh launch; add a test using a stub claude that exits 1 on `--continue` and records the second invocation's argv, asserting `--permission-mode auto` on the fallback session and no `auto` on GLM and no change when a stored mode or typed flag exists.
- F16 [MINOR, optional] progress.md and CHANGELOG do not mention the resume-fallback path; update after F15 is fixed.
- No other defects found in items 1, 2 and 4 of the audit brief: GLM never receives the default `auto`, stored and typed modes are forwarded on a resume, the new-session sync and the GLM resume drop are intact.

AUDIT-VERDICT: FAIL spec=t1866 receipts=rcpt-c68710bc814236dc0a67251f
