auditor-model: claude-sonnet-5-5

verdict: FAIL
audited_sha: 4e9632a326172bd463430a0ff4b859f2b8dd6004
required_backend: codex
overall_score: 0.78
blocking_count: 1
required_backend_fail: codex

Reason for the FAIL: the required codex backend returned fail on a P1 that I then reproduced on a binary built from this HEAD. It is the degraded-path case recorded as optional debt D3 in round 2, now shown to be demonstrable, so it is promoted to blocking. Everything else in the scoped HEAD confirmation is clean.

Judging moai build: the MCP server and installed CLI are build ee981a932 (an ancestor of HEAD, confirmed by `git merge-base --is-ancestor ee981a932 HEAD`). The reproduction below used a binary built from the tree at HEAD 4e9632a326 (`go build -o /private/tmp/t1680-iter3/moai ./cmd/moai`).

Receipt minted by my own call: rcpt-a7d2ef2b7062ad5357bd36cb

## Claim

1. The card code delta is unchanged by the main absorb (four files, empty diff).
2. The CHANGELOG merge is clean: one t1680 entry, one t1859 entry, no markers, nothing else altered.
3. Light checks at HEAD pass: build, Windows build, vet, gofmt, regression selections.
4. Main's absorbed commits do not interact with the changed handlers or `integrationLockRoot` callers.
5. The required codex gate fails, on a P1 that reproduces: `integrationSessionID` now reads the SessionStart sidecar at the repository common root, so a session whose `CLAUDE_PROJECT_DIR` names a linked worktree and whose `CLAUDE_CODE_SESSION_ID` is unset releases another session's window without `--force`.

## Evidence

(1) Code delta:

```
$ git diff e872022eb3 4e9632a326 --stat -- internal/cli/integration.go internal/cli/mcp_factory_card.go internal/cli/factory_mcp_verb_parity_test.go internal/cli/integration_common_root_test.go
(no output)
```

(2) CHANGELOG, main side of the merge (`git diff 4e9632a326^1 4e9632a326 -- CHANGELOG.md --stat`): `CHANGELOG.md | 3 +++`, 3 insertions, 0 deletions. The hunks are: two added lines under Added (SPEC-MOAI-DB-PATHS-001, SPEC-LAUNCHER-ARGV-001, both main-side entries) and one added line under Fixed (the t1859 entry); the t1680 Fixed entry is already present as context in the diff, i.e. it came from the card side.

```
$ grep -c 't1680' CHANGELOG.md          -> 1
$ grep -c 't1859' CHANGELOG.md          -> 1
$ grep -c '^<<<<<<<\|^>>>>>>>\|^=======$' CHANGELOG.md   -> 0
```

(3) Light checks:

```
go build ./...                -> BUILD_OK
GOOS=windows go build ./...   -> WIN_OK
go vet ./internal/cli/        -> VET_OK
gofmt -l internal/cli         -> (empty; GOFMT_DONE)
go test ./internal/cli/ -count=1 -v -run 'TestMCPVerbParity_|TestIntegrationCommonRootOf_|TestIntegrationLockRoot_Linked|TestIntegrationWindowRootAgrees'
  12 tests ran, all "--- PASS" (TestMCPVerbParity_RecordRoot, _LaneLabel, _DecideConfigRoot,
  _CompleteWindowRootNonParentEnv, _CompleteWindowRoot; TestIntegrationWindowRootAgreesWithTheFactoryRecordRoot;
  TestIntegrationLockRoot_LinkedWorktreeVariableResolvesToThePrimaryCheckout;
  TestIntegrationCommonRootOf_ x5)
  ok  github.com/modu-ai/moai-adk/internal/cli  47.770s
go test ./internal/cli/ -count=1 -v -run '^TestSD_AC|DecidePush'
  33 "--- PASS", ok  github.com/modu-ai/moai-adk/internal/cli
```

(4) Absorbed commits since ae34f26831 touching the watched paths (`git log --oneline ae34f26831..7705b365e2 -- internal/cli/factory_card.go internal/cli/mcp_factory_card.go internal/cli/integration.go internal/homestate internal/factory`): 524f349d4a, 261aae8d19, 67095b7c32 (merges), 88a5d5e01f (t1688: factory complete with no card id is a usage error), 1fd3fffdaf (t1859: reassign for decision-waiting cards, in internal/homestate/card_reassign.go plus two help strings in factory_card.go), baecbf74c6, 93af4a485b, f10adec63e (t1827: moai paths, QueueLocationForRoot in internal/factory/queue_location.go). The card-owned files (integration.go, mcp_factory_card.go) have an empty diff against e872022eb3, so main did not touch them. The t1688 and t1859 changes sit in `factory_card.go` argument handling and `homestate` reassign, which the t1680 handlers do not call; t1827 adds a new file in `internal/factory`. I found no concrete line where they interact with `factoryLaneLabelFromEnv`, `factoryRootOf`, `integrationCommonRootOf` or the `integrationLockRoot` callers. Evidence for that is the empty four-file diff plus the passing regression selections above; I did not read every one of the eight commits line by line (see Gaps).

(5) Required gate. `moai verify audit-plan --project-root <tree>` reported `"cross_model_required": true`, `"enforced_required": ["codex"]`, codex gate required, claude and glm off. `mcp__moai__audit_multi` (project_root = tree, card_id t1680, target baseBranch) returned: overall fail, plan_source config, codex gate required, backend result fail, residual note "required-backend FAIL: codex", audit_receipt rcpt-a7d2ef2b7062ad5357bd36cb, build_commit ee981a932 (lag note: installed binary is an ancestor of HEAD).

Codex finding (P1, internal/cli/integration.go:52, translated from the Korean text): when `CLAUDE_PROJECT_DIR` points at a linked worktree and `CLAUDE_CODE_SESSION_ID` is unset, `integrationSessionID("")` reads the parent checkout's session file; if that file holds the lock holder's id, another session's `integration release` removes the lock without `--force`.

My reproduction on a binary built from HEAD, in a scratch repo under /private/tmp/t1680-iter3 (primary `repo`, linked tree `wt`):

```
env -u CLAUDE_CODE_SESSION_ID CLAUDE_PROJECT_DIR=<repo> moai integration acquire --name holder --session HOLDER-SESSION
  -> release-integration window acquired by HOLDER-SESSION on main
sidecar <repo>/.moai/state/current-session-id.txt = HOLDER-SESSION
sidecar <wt>/.moai/state/current-session-id.txt   = OTHER-SESSION   (the caller's own slot)
cd <wt>; env -u CLAUDE_CODE_SESSION_ID CLAUDE_PROJECT_DIR=<wt> moai integration release
  -> release-integration window released (was HOLDER-SESSION on main)
```

The caller's own slot said OTHER-SESSION, yet the release succeeded as HOLDER-SESSION, with no `--force`. The cause is `internal/cli/integration.go:176`: `os.ReadFile(filepath.Join(integrationLockRoot(), session.CurrentSideChannelFile))`, and `integrationLockRoot()` now returns `integrationCommonRootOf(dir)` for a set `CLAUDE_PROJECT_DIR` (line 52). The hook writes the sidecar at `input.ProjectDir` (`internal/hook/session_start.go:483`), so the read root and the write root now disagree whenever the variable names a linked worktree.

Convergence check, after writing the digest to `.moai/state/audit-plan-result.json` (untracked) and running `moai verify audit-plan --project-root <tree> --result-file .moai/state/audit-plan-result.json`:

```
"convergence_check": {
    "ok": true,
    "unmet": []
}
```

This only says the digest is consistent with the plan (a fail from the required backend was recorded as such); it does not turn the codex fail into a pass. The overall result stays fail and the verdict is not final until the orchestrator re-runs this check.

## Findings

- F1 [P1] [blocking] internal/cli/integration.go:176 (read root) with :52 (resolver) - `integrationSessionID` reads the SessionStart sidecar at `integrationLockRoot()`, which t1680 changed to the common root. When `CLAUDE_PROJECT_DIR` names a linked worktree and `CLAUDE_CODE_SESSION_ID` is empty, the caller adopts the session id of whichever session last started at the primary, and `release` (and any ownership check using the same id) passes for a window it does not hold. Reproduced above. Required fix: resolve the sidecar read root separately from the shared lock root (read it at the directory the hook writes to, i.e. `CLAUDE_PROJECT_DIR` as given, falling back to the lock root only when that read fails), keeping `integrationCommonRootOf` for the lock record, config and re-measure paths; add a regression test with a linked worktree, an unset session env var, and two different sidecar slots, asserting that `release` is refused. Alternative: when the caller's id cannot be established without the cross-root read, require an explicit `--session`.
- This is the same item as D3 in `.moai/reports/t1680/sync-audit-iter2.md` ("optional, hypothetical in the supported flow"). Round 2 rested on code reading and ran no repro; codex supplied one and I reproduced it, so it moves from optional to blocking. Probability in the supported flow is low (Claude Code sets the variable to the project root, which already read the primary), but the change makes the degraded path unsafe in the one configuration where it previously read the caller's own slot.

Debt unchanged from round 2 (optional, not driving this FAIL): D1, D2, D4, D5, and N3/F8/(c) tracked in t1867.

## Baseline-attribution

- Tree: /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1680, HEAD 4e9632a326172bd463430a0ff4b859f2b8dd6004, clean at audit start.
- Comparison base for the code delta: e872022eb3 (round 2 audited sha).
- Commands above were run in this session against this tree. The reproduction used a binary built from this tree, not the installed ee981a932 build. The codex audit was produced by the MCP server running build ee981a932 reading this tree via project_root.
- No go-test-heavy slot taken; only `-run` selections with nonzero ran-counts.

## Gaps

- No control run of the pre-t1680 build for the F1 reproduction (I did not build the previous commit); the claim that the old build refused rests on code reading (old resolver returned `CLAUDE_PROJECT_DIR` as written, so the read hit the caller's own slot).
- Item (4) rests on the empty four-file diff, the file list of the eight absorbed commits, and the passing selections; I did not read all eight commits' diffs in full.
- The package-wide `internal/cli` suite was not run (board ruling 35e9); coverage was not measured in this round.
- No refused commands this round. `go test -run` output was filtered with grep/awk for brevity, so only the pass lines are quoted.
- The codex finding text came back in Korean; the English summary above is mine.

## Residual-risk

- After F1 is fixed, a second pass should include the CLI session-resolution sibling at `internal/cli/session.go:282`, which also reads the sidecar from a project dir, to confirm it already reads the same root as the hook writes.
- The convergence check passing does not mean the gate passed; the orchestrator must treat the required-backend fail as final until a re-audit returns a pass.
- The scratch repo and binary live under /private/tmp/t1680-iter3 and are not part of the tree.
