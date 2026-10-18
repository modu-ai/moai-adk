auditor-model: claude-sonnet-5-5

verdict: PASS
audited_sha: 22c68226033ee3eb722e9bb4fb63a5871fc29abc
required_backend: codex
overall_score: 0.95
blocking_count: 0

Final delta audit (rule C, no further rounds), card t1680, scope `git diff 31b25dfc2b 22c6822603`. R3-1 is fixed and reproduced fixed against a control that still fails. The required codex gate passed. No blocking findings; remaining items are optional debt.

Judging moai build: MCP server and installed CLI are build ee981a932 (ancestor of HEAD 22c6822603, per the lag note the audit_multi call returned). The reproduction and mutant runs used binaries and overlays built from this tree, not the installed build.

Receipt minted by my own call: rcpt-3ad9e80222ed6e67539b9db0

## Claim

1. R3-1 is fixed: with CLAUDE_PROJECT_DIR naming a linked worktree and CLAUDE_CODE_SESSION_ID unset, `moai integration release` without `--force` is refused and the window stays held; a control built from 31b25dfc2b releases it.
2. The new tests are non-vacuous: reverting the fix and each helper mutant is killed by at least one test; the control cases (explicit `--session`, set env var, variable unset) hold.
3. The correction does not change behaviour for the normal lane case (variable naming the primary checkout) or the unset case; the only behavioural change is the intended one.
4. Light checks at HEAD pass with nonzero ran-counts.
5. The one golangci-lint issue (paths.go:172:79 ST1008) is baseline, not this card's defect.
6. The required codex gate passes and the convergence check is ok.

## Evidence

(1) Reproduction, scratch repo under /private/tmp/t1680-iter4 (script /private/tmp/t1680-iter4/repro.sh). Primary `repo`, linked tree `wt`; sidecar at repo = HOLDER-SESSION, sidecar at wt = OTHER-SESSION; holder record seeded with session HOLDER-SESSION and pid 0 so the session id is the only ownership key.

Why the record is seeded rather than acquired through the verb: my first attempts used `moai integration acquire --session HOLDER-SESSION`, and the release succeeded on BOTH builds. The lock file shows why: `"pid": 18282, "pid_source": "session-owner"`, the owning-session pid of this audit's own process ancestry, which the later release resolves to the same value, so `releasableBy` admitted it by the second (pid) key (integration_lock.go:521-528). `MOAI_SESSION_PID=1` on acquire did not change the recorded pid. That run therefore did not discriminate and I discarded it. With the pid key neutralised (pid 0, no pid source), the id key alone decides:

Build from HEAD 22c6822603 (`go build -o /private/tmp/t1680-iter4/bin/moai-head ./cmd/moai`):

```
--- release from linked worktree, CLAUDE_PROJECT_DIR=wt, CLAUDE_CODE_SESSION_ID unset, no --force
   ERROR
  Release integration window is held by a different session: holder (pid 0) holds it.
exit=1
--- status after
release-integration window: held
  holder:   holder (HOLDER-SESSION, pid 0)
```

Control built from 31b25dfc2b via `go build -overlay` replacing internal/cli/integration.go with `git show 31b25dfc2b:internal/cli/integration.go`:

```
--- release from linked worktree, CLAUDE_PROJECT_DIR=wt, CLAUDE_CODE_SESSION_ID unset, no --force
release-integration window released (was HOLDER-SESSION on main)
exit=0
--- status after
release-integration window: free
```

Same script, same tree layout, only integration.go differs: the repro discriminates. The refusal names the holder by its session name label ("holder"), the record's own display label (`holderLabel` prefers the name); the id is in the record.

(2) Mutation, via `go build`/`go test -overlay` on internal/cli/integration.go only, selection `-run 'TestIntegrationSessionID_|TestIntegrationRelease_LinkedWorktreeCaller'`:

```
HEAD (unmutated):     5 tests, all "--- PASS"          ok  github.com/modu-ai/moai-adk/internal/cli  1.676s
M1 revert (31b25 file, i.e. read at integrationLockRoot):
   --- FAIL: TestIntegrationSessionID_ReadsTheSlotWhereTheHookWritesIt
   --- FAIL: TestIntegrationRelease_LinkedWorktreeCallerCannotReleaseAnotherSessionsWindow
     "release by CALLER-SESSION succeeded on a window held by HOLDER-SESSION without --force: release-integration window released (was HOLDER-SESSION on main)"
M2 helper always returns the common root (integrationLockRoot()):
   --- FAIL: ..._ReadsTheSlotWhereTheHookWritesIt   --- FAIL: ..._CannotReleaseAnotherSessionsWindow
M3 helper always returns cwd (os.Getwd()):
   --- FAIL: ..._ReadsTheSlotWhereTheHookWritesIt   --- FAIL: ..._CannotReleaseAnotherSessionsWindow
   --- FAIL: TestIntegrationSessionID_PlainProjectReadsWhereTheHookWrites
M4 unset-variable fallback replaced by ".":
   --- FAIL: TestIntegrationSessionID_ControlCasesKeepTheirPrecedence/variable_unset_keeps_the_git-resolved_fallback
```

Four mutants, four killed, no survivor. Control cases at HEAD: explicit `--session` wins over the file, the set env var wins over the file, and the variable-unset fallback reads the git-resolved common root; all three sub-tests pass (and fail only under M4 as intended). The explicit and env precedence branches sit before the helper, so no helper mutant can reach them; they are guarded by these sub-tests and the older TestIntegrationSessionID_EnvBeatsSidecar (passes).

(3) Sweep. Readers of the side channel (`grep -rn CurrentSideChannelFile internal --include='*.go'`, non-test hits): internal/hook/session_start.go:483 (writer, at input.ProjectDir), internal/cli/session.go:282 (`moai session current`, reads at `resolveProjectDir()` which returns CLAUDE_PROJECT_DIR as given, falling back to cwd), internal/cli/integration.go:176 (this card, now `integrationSidecarRoot()`), and registry.go:83 (the constant). The new helper agrees with session.go:282 and the hook write root whenever the variable is set; the residual-risk sibling item from the iter3 report is therefore closed by reading, and the only difference is the unset case (helper: git-resolved common root as before; session.go: cwd), which is unchanged pre-existing behaviour.

Callers of `integrationSessionID` (`grep -rn "integrationSessionID(" internal --include='*.go'`, non-test): factory_card.go:2078, slot.go:117 and :304, integration_merge.go:43, session_worktree_automerge.go:70 (seam), integration.go:469 (acquire) and :654 (release), factory_merge.go:180, local_main_resync.go:116. All call it with the same semantics (explicit flag, then env var, then the file). The change alters the returned id only when all three hold: no explicit flag, CLAUDE_CODE_SESSION_ID empty, and CLAUDE_PROJECT_DIR set to a directory other than the repository common root (a linked worktree or a subdirectory). When the variable names the primary checkout (the normal lane case) `integrationCommonRootOf(primary)` is the primary itself, so the read location is byte-identical to before; when it is unset the path is the old integrationLockRoot fallback, unchanged. In the changed configuration the read now follows the slot the hook wrote for that session. A degraded side effect: if that slot is missing the id is empty, which acquire/merge/factory complete refuse with their existing "cannot resolve this session's id" errors and release refuses unless `--force`; this fails closed, where the old behaviour adopted another session's id. The MCP handlers pass project_root and do not call this helper for the root; the lock record, config and re-measure roots still come from integrationLockRoot, untouched by the delta.

(4) Light checks at HEAD 22c6822603:

```
go build ./...                -> BUILD_OK
GOOS=windows go build ./...   -> WIN_OK
go vet ./internal/cli/        -> VET_OK
gofmt -l internal/cli         -> (empty)
go test ./internal/cli/ -count=1 -v -run 'TestIntegrationSessionID_|TestIntegrationRelease_LinkedWorktreeCaller|TestMCPVerbParity_|TestIntegrationCommonRootOf_|TestIntegrationLockRoot_Linked|TestIntegrationWindowRootAgrees|^TestSD_AC|DecidePush'
  49 "--- PASS" lines, 0 FAIL; final line: ok  github.com/modu-ai/moai-adk/internal/cli  215.550s
  (includes TestIntegrationSessionID_x5, TestIntegrationRelease_LinkedWorktreeCaller..., TestMCPVerbParity_x5,
   TestIntegrationCommonRootOf_x5, TestIntegrationLockRoot_Linked..., TestIntegrationWindowRootAgrees..., TestSD_AC x27, DecidePush x3)
golangci-lint run --timeout 300s ./internal/cli/
  internal/cli/paths.go:172:79: ST1008: error should be returned as the last argument (staticcheck)
  1 issues: * staticcheck: 1
```

The lint run finished on the package. Attribution of that single issue: `git log -1 --format=... -- internal/cli/paths.go` gives 2a8a66bf596c92ace6668cde2585105eb1dfcd98 (main's t1827, Oct 18 09:55); `git merge-base --is-ancestor 2a8a66bf59 7705b365e2` exits 0 (an ancestor of local main); `git log --oneline main..HEAD -- internal/cli/paths.go` is empty and `git diff 7705b365e2 22c6822603 --stat -- internal/cli/paths.go` is empty, so the file is identical to main's. Correction to the brief: the file does appear in `git diff ae34f26831 22c6822603 --stat` (+355 lines), because that range spans the absorb of main and so includes everything main added since ae34f26831; the file is outside the card's own delta against main, which is the test that matters. Classification: baseline from t1827, not a defect of this card.

(5) Required gate. `moai verify audit-plan --project-root <tree>` reported codex required, claude and glm off, cross_model_required true, enforced_required codex. `mcp__moai__audit_multi` with project_root = the tree, card_id t1680, target baseBranch, focused on the delta, returned: overall result pass, plan_source config, one backend entry codex, gate required, result pass, zero findings, fail_open_backends empty, audit_receipt rcpt-3ad9e80222ed6e67539b9db0, build_commit ee981a932 with the lag note (installed binary an ancestor of HEAD). Digest written to `.moai/state/audit-plan-result.json` (untracked; `git status --short` empty afterwards) and checked with `moai verify audit-plan --project-root <tree> --result-file .moai/state/audit-plan-result.json`:

```
  "convergence_check": {
    "ok": true,
    "unmet": []
  }
```

Because the codex gate did not fail, there is no codex finding to reproduce this round. This audit's result is not final until the orchestrator re-runs that check.

## Findings

No blocking findings.

- F1 [optional] [none] internal/cli/integration.go:186-196 - `integrationSidecarRoot` reads `CLAUDE_PROJECT_DIR` as given with no git resolution, so a relative or symlinked spelling is used verbatim. This is intentional (it must match the string the hook joined) and the hook side does the same; no defect, recorded only so the next reader knows the spelling is deliberate.
- D6 [optional debt] integration_session_sidecar_test.go - the release test seeds the holder record with `factory.AcquireIntegrationLock` without a session-owner pid, so it exercises the session-id key only; it does not cover the pid key. That is the right isolation for R3-1, and the pid key has its own tests; mentioned because my first reproduction was confounded by exactly that key.

Debt unchanged and not re-judged: D1, D2, D4, D5, and N3/F8/(c) tracked in t1867.

## Baseline-attribution

- Tree: /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1680, HEAD 22c68226033ee3eb722e9bb4fb63a5871fc29abc, clean at audit start and at end (`git status --short` empty).
- Delta compared: 31b25dfc2b to 22c6822603 (integration.go +17/-1, integration_session_sidecar_test.go new 160 lines, progress.md +40).
- All commands above were run in this session against this tree. Reproduction binaries: moai-head (HEAD) and moai-ctrl (31b25dfc2b overlay), both in /private/tmp/t1680-iter4/bin. Mutants and overlays in /private/tmp/t1680-iter4/ov; nothing in the tree was edited. The codex audit was produced by the MCP server running build ee981a932 reading this tree via project_root.
- No go-test-heavy slot taken; the largest run was the requested light selection (215 s, `-run` selections with nonzero ran-counts, never the full internal/cli suite).

## Gaps

- The pre-t1680 behaviour (before common-root normalisation) was not rebuilt as a third control; the control used is 31b25dfc2b, the commit that exhibited R3-1, which is the comparison the brief asked for.
- Refusals this round: one compound command (`git show ... > file && cat > ...`) and one redirect-plus-`echo exit=$?` variant were refused by the worktree guard; both were split into plain commands and re-run, with the same inputs. No result rests on the refused forms.
- `go test` output was filtered with grep for the pass/fail lines; the unfiltered lines were not retained. The 49-pass count is from those filtered lines.
- Whether Claude Code in a worktree session sets CLAUDE_PROJECT_DIR to the primary checkout or to the worktree was not re-measured this round; the analysis covers both and the result does not depend on it.
- Only the delta was audited; the full internal/cli package suite was not run (board ruling 35e9) and coverage was not measured.
- The first reproduction attempts (acquire through the verb) released on both builds because of the owner-pid key; I report them as non-discriminating rather than evidence either way.

## Residual-risk

- The degraded path (variable set, env session id empty) is only reachable in an unusual launch; the fix closes the cross-session adoption there and fails closed when the slot is absent. A lane that hits the fail-closed error needs `--session` or `--force`, a visible and recorded step.
- The sidecar holds one slot per project directory with last-SessionStart-wins semantics; two sessions starting in the same directory can still overwrite each other's slot. That limitation is documented in session.go and is outside this card.
- The convergence check passing records consistency of the digest with the plan; the required backend actually returned pass this round, so there is no waiver involved.
- The scratch repos and binaries under /private/tmp/t1680-iter4 are not part of the tree.
