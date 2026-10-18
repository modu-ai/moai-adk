auditor-model: claude-sonnet-5-5
verdict: FAIL
audited_sha: 92d7270a32ab9f6429d3a906ad090c5499050a43

# t1866 sync audit, round 8 (delta 92d7270a32 over fe67c01f2)

Score: Functionality 80, Security 95, Craft 88, Consistency 90. Overall FAIL: the required codex gate returned fail (receipt rcpt-9833ef82b66be72a725eb4d3) and I reproduced its finding F17 with a scratch-overlay probe.

## Claim

C1. F15 is repaired for the Claude backend: an unset profile on `--continue`, `-c` or `DO_CLAUDE_CONTINUE` whose child exits 1 now starts the fallback new session with `--permission-mode auto` and the `auto` settings.local.json sync.
C2. Stored profile mode, typed `--permission-mode`, `DO_CLAUDE_BYPASS` legacy path and settings-chosen mode are never replaced (the captured fact `profileModeUnset` is false for each); the exit-2+ branch and the successful-resume path are unchanged; `buildArgs` reads `permMode` by closure, so the reassignment reaches the fallback argv; env/lease handling of the fallback exec is unchanged.
C3. Tests prove the first call has no permission flag and the second has `auto` with the file synced; the mutation claim is plausible.
C4. NOT TRUE: the fallback never damages anything on the GLM backend. For an unset GLM profile the fallback calls `syncPermissionModeToSettingsLocal(path, "")`, which deletes ANY `permissions.defaultMode` (a hand-written `plan`, `acceptEdits`, `default`), while a successful GLM resume and the pre-repair fallback left it alone (F17).
C5. CHANGELOG and progress.md claims are true of code and tests, with one omission tied to F17.

## Evidence

Code (`internal/cli/launcher.go`, HEAD 92d7270a32):

```
724:	profileModeUnset := permMode == "" && !consumed.PermissionModeSet
725:	if profileModeUnset && !cont {
726:		permMode = launchDefaultPermissionMode(permGLM)
900:		if errors.As(err, &ee) && ee.ExitCode() == 1 {
904:			if profileModeUnset {
905:				permMode = launchDefaultPermissionMode(permGLM)
906:				if settingsLocalPath != "" {
907:					if err := syncPermissionModeToSettingsLocal(settingsLocalPath, syncPermMode); err != nil {
912:		} else {  // exit 2+: returns "resume session failed", untouched
924:	return execOrSpawnClaudeFunc(claudeBin, buildArgs(false), launchEnv)
778:	buildArgs := func(withContinue bool) []string { ... if permMode != "" { a = append(a, "--permission-mode", permMode) } ...
```

Reading results:
- GLM never receives auto: `launchDefaultPermissionMode(true)` returns "" (line 1107-1112); a stored/settings `auto` is already zeroed at 681-687 before 724, and `syncPermMode` is zeroed for GLM at 685. Held. The GLM resume drop (744-748) runs before the exec and is unchanged. A stale `auto` on a GLM fallback is already removed by that drop and again by the "" sync.
- Stored mode / typed flag / settings-chosen mode / legacy bypass: `permMode != ""` or `PermissionModeSet`, so `profileModeUnset` is false; no default, no sync. Held.
- A fallback for a profile with a stored or settings-chosen mode gets that mode's `--permission-mode` flag on argv (flag beats the file) but no settings sync. The difference to a fresh launch is only the file content; argv decides the session mode. I judge this an accepted, documented non-goal (progress.md Gaps states it). Not a defect.
- Other paths starting a new session after a failed resume: `grep -rln "No previous session" internal cmd --include='*.go'` returns only `internal/cli/launcher.go`; the managed factory path refuses `cont` (line 880-882). No other path.
- DEFECT: GLM unset profile reaches 904 (`profileModeUnset` true because `permMode` is "" for GLM). `syncPermMode` is "" there, and `syncPermissionModeToSettingsLocal` with "" executes the else branch (lines 1172-1180 of launcher.go) that deletes `permissions.defaultMode` whatever its value. On a successful GLM resume only a value equal to `auto` is removed (`dropAutoDefaultModeFromSettingsLocal`, 1118-1137).

Scratch probe (overlay outside the worktree, nothing written into it), GLM backend, unset profile, hand-written `defaultMode: plan`, stub exits 1 on `--continue`:

```
go test -overlay .../ov/overlay-probe.json ./internal/cli -run '^TestProbe_GLMFallback' -count=1 -v
=== RUN   TestProbe_GLMFallbackKeepsHandWrittenDefaultMode
No previous session found, starting new session...
    probe_glm_fallback_test.go:16: calls=[[--continue] []]
    probe_glm_fallback_test.go:18: settings.local.json defaultMode after GLM fallback = "", want plan kept
--- FAIL: TestProbe_GLMFallbackKeepsHandWrittenDefaultMode (0.24s)
```

Scoped tests, each as one invocation starting with the `unset MOAI_KANBAN_*` trio:

```
go test ./internal/cli -run '^TestLaunch_' -count=1 -v   -> exit=0; 37 "--- PASS", 0 "--- FAIL", 0 "--- SKIP"; last lines: PASS / ok  github.com/modu-ai/moai-adk/internal/cli	4.986s
--- PASS: TestLaunch_ClaudeResumeFallbackNewSessionGetsAutoDefault (0.75s)  [3 subtests PASS: --continue, -c, DO_CLAUDE_CONTINUE=true]
--- PASS: TestLaunch_ClaudeResumeFallbackKeepsChosenModeAndGLMStaysWithoutAuto (0.74s)  [3 subtests PASS: stored_acceptEdits_profile, typed_plan, GLM_backend_unset_profile]
go test ./internal/settings ./internal/cli/wizard ./internal/core/project -count=1
ok  	github.com/modu-ai/moai-adk/internal/settings	0.617s
ok  	github.com/modu-ai/moai-adk/internal/cli/wizard	6.437s
ok  	github.com/modu-ai/moai-adk/internal/core/project	3.572s
```

Mutation check (scratch overlay replacing launcher.go with a copy where line 904 reads `if false && profileModeUnset {`; worktree untouched, `git status --short` empty afterwards):

```
go test -overlay .../ov/overlay-mut.json ./internal/cli -run '^TestLaunch_ClaudeResumeFallback' -count=1
--- FAIL: TestLaunch_ClaudeResumeFallbackNewSessionGetsAutoDefault (0.71s)
    --- FAIL: .../--continue (0.25s)
    --- FAIL: .../-c (0.24s)
    --- FAIL: .../DO_CLAUDE_CONTINUE=true (0.22s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	2.622s
```

The progress.md mutation claim is therefore true (the guard test `...KeepsChosenModeAndGLMStaysWithoutAuto` is not expected to fail under that mutation, and I did not claim it).

Test review (from bodies in `internal/cli/launcher_permission_resume_test.go`):
- `pinFallbackClaude` appends every invocation's argv to a file and exits 1 when argv contains `--continue`; it also routes the exec seam through the stub so the fallback call is recorded. `recordedChildCalls` fails when the file is missing and keeps an empty line as a no-arg call (helper fix confirmed: `strings.Fields("")` yields an empty slice, and my probe saw `calls=[[--continue] []]`, so the no-arg call is counted). Assertions: exactly 2 calls, call 1 has `--continue` and no `--permission-mode`, call 2 has no `--continue` and `--permission-mode auto`, and the file default is `auto`. Meaningful; no skip, no weakened assertion.
- Guard cases: stored acceptEdits (both calls carry it), typed plan (both calls carry it), unset GLM (neither call has a mode, file default is not auto). Meaningful for the stated claims, with one hole: the GLM case asserts only `!= "auto"` on the file and never seeds a hand-written non-auto value, so F17 slips through.

Required gate, `mcp__moai__audit_multi` (one call, project_root = worktree toplevel):

```
overall_verdict: fail
plan_source: config
per_backend_verdicts: backend=codex gate=required verdict=fail
audit_receipt: rcpt-9833ef82b66be72a725eb4d3
gate_unmet: (absent in result)
residual_risk_note: required-backend FAIL: codex
build_commit: ee981a932 (ancestor of HEAD 92d7270a3; binary lag noted by the tool)
codex finding [P2] internal/cli/launcher.go:907: GLM unset profile enters the fallback branch and syncs with an empty syncPermMode, deleting permissions.defaultMode in settings.local.json, so a hand-written plan restriction is lost; plan, acceptEdits and default were preserved at fe67c01f2 and deleted at 92d7270a32. Confidence high. Recommendation: limit the fallback default and sync to profileModeUnset && !permGLM and add a GLM fallback regression test that keeps a hand-written value.
```

The orchestrator's `moai verify audit-plan --result-file` check was not run by me; this verdict is not final until that check passes.

CHANGELOG / progress.md review:
- CHANGELOG sentence "if the resume child finds no previous session ... that new session gets the `auto` default and the `settings.local.json` sync like any new session (for a profile that never chose a mode, on the Claude backend only; a stored profile mode and a typed `--permission-mode` are forwarded unchanged, and a GLM backend still receives no `auto`)" is true for the Claude backend and for "no auto on GLM". It says "the sync ... on the Claude backend only", but the code also runs the (empty) sync on GLM, which is the behavior F17 flags; the sentence therefore promises less than the code does and is wrong by omission.
- progress.md "Audit repair round 5": RED/GREEN quotes are plausible; the 37 PASS / 0 FAIL / 0 SKIP count reproduced; the mutation claim reproduced; the helper-bug note is consistent with the code. The statement "a GLM backend gets "" ... so it stays unchanged" is false for the file: GLM's `defaultMode` is deleted by the sync.

## Baseline-attribution

- Tree: worktree `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1866`, HEAD `92d7270a32ab9f6429d3a906ad090c5499050a43` (`git rev-parse HEAD`, this run), `git status --short` empty before and after the probes.
- Go tests, the mutation check and the probe ran from the Go toolchain against this tree (overlay probes read files from a scratch dir outside the tree); no installed moai binary judged them.
- `moai verify audit-plan` and `audit_multi` used the running moai build `ee981a932` (MCP server pid 15518, built 2026-10-15), reported by the tool as an ancestor of HEAD `92d7270a3`. They judged policy and the codex backend review, not the Go test results; audit-plan output: `config_status: ok`, codex `required`, `cross_model_active: true`.

## Gaps

- No whole-package run of `internal/cli`, `internal/web`, `internal/template` or `internal/config` (per instruction); no live Claude process (stub only).
- The codex parent-vs-child comparison (`fe67c01f2` preserved plan) was reported by the codex leg; I confirmed the defect by code reading and a probe at HEAD, and confirmed the parent preserved the file by reading (no sync on the cont path, drop removes only `auto`), not by running the parent tree.
- `gate_unmet` was absent from the audit_multi output; I did not run the orchestrator's `moai verify audit-plan --result-file` check (read-only role).
- No guard refusal occurred this round; the worktree guard was not exercised. No guard record was touched.

## Residual-risk

- After F17 is fixed, the same unconditional "" sync still runs on every fresh GLM launch (pre-existing, pre-dates this card): a fresh GLM session also deletes a hand-written `defaultMode`. F17 is a regression only relative to the resume/fallback consistency, so the fixer may argue the fallback should mirror a fresh launch; the codex gate and I read the safer, narrower fix as preferred.
- Stub-based tests do not prove how a real Claude process interprets the flag next to `defaultMode`.
- A fallback for a profile with a stored or settings-chosen mode still skips the file sync (accepted, documented).

## Findings

- F17 [BLOCKING, required codex gate fail; independently reproduced] `internal/cli/launcher.go:904-911`. The resume-fallback re-application runs for an unset GLM profile too (`profileModeUnset` is true there), and `syncPermissionModeToSettingsLocal(path, "")` deletes any `permissions.defaultMode`, including a hand-written `plan`/`acceptEdits`/`default`. A successful GLM resume (only `auto` is dropped) and the pre-repair fallback kept it. Required fix: restrict the fallback block to the Claude backend, i.e. `if profileModeUnset && !permGLM {` (the GLM branch already needs neither the default, which is "", nor the sync, since `dropAutoDefaultModeFromSettingsLocal` ran on the resume path). Add a regression test next to `TestLaunch_ClaudeResumeFallbackKeepsChosenModeAndGLMStaysWithoutAuto`: GLM backend, unset profile, hand-written `defaultMode` seeded to `plan`, stub exits 1 on `--continue`, assert the value is still `plan` after the fallback and neither call has `--permission-mode`.
- F18 [SHOULD-FIX, optional until F17 lands] CHANGELOG and progress.md: state precisely what the fallback does on GLM (nothing: no flag, no file change) after F17 is fixed, and drop the claim "a GLM backend gets "" ... stays unchanged" if the sync is kept.
- F19 [MINOR, optional] The GLM guard case asserts `got == "auto"` is absent but never seeds a non-auto hand-written value, so the test could not see F17. Folded into the F17 test above.
- No other defects found in items 1, 2 (apart from the F17 test hole) and 4 of the brief: GLM never receives `auto` by flag, stored and typed modes are forwarded, the successful-resume path passes no default flag and does not rewrite the file, and the exit-2+ branch is unchanged.

AUDIT-VERDICT: FAIL spec=t1866 receipts=rcpt-9833ef82b66be72a725eb4d3
