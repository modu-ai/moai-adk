auditor-model: claude-sonnet-5-5

# t1680 sync-audit, round 2 (delta 11e07d2c5f..e872022eb3)

verdict: PASS-WITH-DEBT
no blocking defect; F1 from round 1 is fixed and pinned, the required codex gate passes; five optional debts listed below
audited_sha: e872022eb3b2269fe59d5474c8131a24d675dabe
required_backend: codex (gate required, plan_source config) -> pass, receipt rcpt-cf99d57ac958b0cb3612a404
overall_score: 0.90
blocking_count: 0

| Dimension | Score | Verdict |
|---|---|---|
| Functionality (40%) | 0.93 | PASS (F1 reproduced pre-fix, gone at HEAD, both surfaces) |
| Security (25%) | 0.92 | PASS (the window-bypass edge is closed; no input/secret surface) |
| Craft (20%) | 0.88 | PASS (RED-first shown; 3 of 5 mutants killed, 2 survivors classified below) |
| Consistency (15%) | 0.88 | PASS |

Judging builds (section 2.2): tree HEAD e872022eb3b2269fe59d5474c8131a24d675dabe. The installed `moai` (CLI `v3.2.0-rc.36-t1822c`) and the MCP server are commit ee981a932, an ancestor of HEAD (`build_lag` says so), so the audit_multi/codex verdict came from an older tool build than the tree. Every `go` result below was built from this tree (overlays under /private/tmp/t1680-audit2/), not from the installed binary.

## Defects (all optional; none blocks)

- D1 [low] [optional] internal/cli/mcp_factory_card.go:256 - mutant M2 (handler passes `root` instead of `integrationCommonRootOf(launch)`) survives every new test. On all layouts the tests build, `factoryRootOf` and `integrationCommonRootOf` agree, so the explicit call is unpinned. It only matters in layouts where they differ (D2). Debt: a row on such a layout would pin it.
- D2 [low] [optional] residual layouts - for a bare repository, a separate-git-dir repository, and a submodule, the helper's `.git`-only rule falls back to the directory itself. Measured (probe below): a linked worktree of a bare or separate-git-dir repo answers itself as window root while the factory record root answers the metadata dir. This is not new for the verb (`verbOld` equals `verbNow`), but the handler's window root for those layouts changed from the record root (shared metadata dir) to the tree itself, so it now matches the verb and no longer shares one window across trees. The helper comment discloses the fallback; the handler comment and the CHANGELOG say "repository common root" without that qualifier. Reach: layouts MoAI worktrees do not create (they are plain `git worktree add` from a `.git` primary).
- D3 [low] [optional] internal/cli/integration.go:176 - `integrationSessionID` reads the SessionStart side-channel file at `integrationLockRoot()`. That is the one caller of the 16 that is a per-tree-ish read: the hook writes the file at `input.ProjectDir`. If `CLAUDE_PROJECT_DIR` names a linked worktree and `CLAUDE_CODE_SESSION_ID` is empty (degraded path), the read now lands at the primary checkout. Env-unset sessions already read the primary, and Claude Code sets the variable to the project root, so this is consistent with the existing fallback and hypothetical in the supported flow.
- D4 [info] [optional] the deliberate consequence: `CLAUDE_PROJECT_DIR` naming a plain subdirectory of a repository now resolves to the repository root (probe: `verbNow=.../plain`, `verbOld=.../plain/a/b`). Acceptable: the window and re-measure records are per-repository, a subdirectory could never have been a meaningful window root, and the CHANGELOG states it.
- D5 [info] [optional] the env-unset branch of `integrationLockRoot` keeps its own copy of the `.git` common-dir rule (unspelled `filepath.Dir(p)`, `resolveProjectDir()` fallback) instead of calling the helper. Behaviourally compatible; a later cleanup could route it through the helper. Mutant M4 (helper's last fallback returns `dir` unspelled) survives and is an equivalent mutant: `canonicalTreeSpelling` already returns `dir` verbatim wherever git cannot answer. The `@MX:ANCHOR` count (3 vs 4) is observation only, per leader.

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

### Claim 1 - F1 reproduced against the pre-fix tree and gone at HEAD

Evidence. Pre-fix overlay (integration.go and mcp_factory_card.go from 11e07d2c5f; `integration_common_root_test.go` replaced by an empty `package cli` file because it needs the new helper), new parity table kept:

```
go test -count=1 -overlay /private/tmp/t1680-audit2/ov-pre.json -v -run 'TestMCPVerbParity_CompleteWindowRoot' ./internal/cli/
--- PASS: TestMCPVerbParity_CompleteWindowRoot (8.87s)
--- FAIL: TestMCPVerbParity_CompleteWindowRootNonParentEnv (3.72s)
    --- FAIL: .../a_window_another_lane_holds_refuses_the_second_lane_naming_the_holder (1.98s)
    --- FAIL: .../a_re-measure_record_written_at_the_parent_is_found_and_the_card_lands (1.74s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	13.832s
```

Both of codex's failures (window held at the parent not refused; re-measure record at the parent not found) reproduce with `CLAUDE_PROJECT_DIR` set to the linked worktree. At HEAD (built from the tree):

```
go test -count=1 -v -run 'TestIntegrationCommonRootOf_|TestIntegrationLockRoot_Linked|TestIntegrationWindowRootAgrees|TestMCPVerbParity_' ./internal/cli/
--- PASS: TestMCPVerbParity_LaneLabel (1.03s)   (9 leaf PASS)
--- PASS: TestMCPVerbParity_RecordRoot (5.43s)   (3 leaf PASS)
--- PASS: TestMCPVerbParity_CompleteWindowRoot (12.41s)   (2 leaf PASS)
--- PASS: TestMCPVerbParity_CompleteWindowRootNonParentEnv (13.88s)   (2 leaf PASS)
--- PASS: TestMCPVerbParity_DecideConfigRoot (3.58s)
--- PASS: TestIntegrationCommonRootOf_LinkedWorktreeResolvesToThePrimaryCheckout (0.11s)
--- PASS: TestIntegrationCommonRootOf_PrimaryCheckoutIsItself (0.11s)
--- PASS: TestIntegrationCommonRootOf_SubdirectoryOfALinkedWorktreeResolvesToThePrimaryCheckout (0.14s)
--- PASS: TestIntegrationCommonRootOf_NonRepoAndMissingPathsAreReturnedVerbatim (0.03s)
--- PASS: TestIntegrationCommonRootOf_ReturnsGitsOwnSpellingNotTheCallers (0.14s)
--- PASS: TestIntegrationLockRoot_LinkedWorktreeVariableResolvesToThePrimaryCheckout (0.12s)
--- PASS: TestIntegrationWindowRootAgreesWithTheFactoryRecordRoot (0.46s)
ok  	github.com/modu-ai/moai-adk/internal/cli	38.366s
```

The new `NonParentEnv` table runs the same two rows through the verb and the MCP handler (`completeWindowRootRows(t, true)`, `CLAUDE_PROJECT_DIR` = the acting lane's linked worktree), so "verb and MCP both refuse the second lane naming the holder" and "record at the parent found by both" are both in the passing set.

### Claim 2 - mutants

Overlay mutants of the HEAD files, run with `-run 'TestIntegrationCommonRootOf_|TestIntegrationLockRoot_Linked|TestIntegrationWindowRootAgrees|TestMCPVerbParity_CompleteWindowRoot'`:

| Mutant | Result |
|---|---|
| M1 helper stubbed to old verbatim `canonicalTreeSpelling(dir)` | killed: CompleteWindowRoot x2, NonParentEnv x2, CommonRootOf Linked / Subdirectory / Spelling, LockRoot_Linked (9 FAIL lines) |
| M2 handler passes `root` (drops the helper call) | SURVIVES (`ok ... 36.359s`), see D1; equivalent on every layout the tests build |
| M3 t766 break on the success path (`return filepath.Dir(p)` unspelled) | killed: `TestIntegrationCommonRootOf_ReturnsGitsOwnSpellingNotTheCallers` |
| M4 t766 break on the fallback (`return dir`) | survives, equivalent mutant (D5): fallback was already verbatim for any path git cannot answer |
| M5 verb ignores the helper (`integrationLockRoot` returns `canonicalTreeSpelling(dir)`) | killed: NonParentEnv x2, `LockRoot_Linked`, `WindowRootAgrees` |

The one that matters, "break the t766 spelling", is killed on the path git answers. M2 is classified debt (D1): the handler's behaviour with the call removed equals the HEAD behaviour on all tested layouts, so no test can tell, and where it would tell (D2) the layouts are untested.

### Claim 3 - caller regression and residual layouts

Code reading of the call sites listed by `grep -n "integrationLockRoot()" internal/cli/*.go` (non-test): factory_card.go:2050 (window and record roots for `complete`), factory_merge.go:93/113/173/184/296 (target root, remeasure record, github-flow check, lock path, lock read), integration_candidate.go:111, integration_policy.go:76, integration_remeasure.go:31, integration.go:332/458/651 (acquire / status / release roots), integration_merge.go:47, integration_settings_drift.go:218, local_main_resync.go:94. All use the root as the key of repository-wide shared state (window lock, remeasure records, config, integration worktree), which is the shared common root. `local_main_resync.go` and `acquireSettingsDriftPrecondition` are documented as primary-checkout operations, so resolving a linked-worktree variable to the primary is the right answer for them. Only `integration.go:176` (`integrationSessionID`, sidecar file at the root) is a per-tree-ish read: D3. `session_worktree_automerge.go:65` aliases the function and shares the same contract. The author's "no caller wants a per-tree read" is right for 15 of 16 and misses the degraded-path sidecar read (D3), which is optional.

Probe (scratch test via `-overlay`, `/private/tmp/t1680-audit2/zz_probe_test.go`, not in the tree), `CLAUDE_PROJECT_DIR` = dir:

```
plain subdir                 common=.../plain       record=.../plain        verbNow=.../plain       verbOld=.../plain/a/b
bare common / worktree       common=.../bare-wt     record=.../bare.git     verbNow=.../bare-wt     verbOld=.../bare-wt
separate-git-dir primary     common=.../sepgit      record=.../sepgit       verbNow=.../sepgit      verbOld=.../sepgit
separate-git-dir linked      common=.../sepgit-wt   record=.../sepgit.gitdir verbNow=.../sepgit-wt  verbOld=.../sepgit-wt
submodule                    common=.../plain/sm    record=.../plain/sm     verbNow=.../plain/sm    verbOld=.../plain/sm
```

Reading: the subdirectory case is the only verb change versus the old rule other than the linked-worktree case (D4, acceptable). The bare / separate-git-dir linked worktree divergence between window root and record root is pre-existing for the verb (`verbOld` = `verbNow`) and is therefore not a defect of this delta (D2); it is a delta effect only on the handler, which now matches the verb.

### Claim 4 - build, vet, format, regression selections

```
go build ./...                           -> exit 0
GOOS=windows GOARCH=amd64 go build ./... -> exit 0
go vet ./internal/cli/                   -> exit 0
gofmt -l internal/cli                    -> no output, exit 0
go test -count=1 -v -run '^TestIntegration|^TestCandidate|^TestObserveCandidate|^TestComplete|Remeasure|SettingsDrift|LocalMain|AutoMerge|TestQAS_AC008b|TestQWR_AC008|DecidePush' ./internal/cli/
   -> 125 top-level and leaf `--- PASS`, 0 FAIL, 0 SKIP, ok 173.563s
go test -count=1 -v -run '^TestSD_AC' ./internal/cli/
   -> 30 `--- PASS`, 0 FAIL, ok 129.056s
```

### Claim 5 - CHANGELOG, handler comment, scope

- CHANGELOG: a single t1680 entry (`git diff` shows one changed line, `2 +-`); it now says `factory_complete` uses the factory root for the card record and the repository common root for the window and re-measure record, names `integrationCommonRootOf`, states the CLI consequence (linked worktree or subdirectory now resolves to the primary checkout), and still says N3 / F8 / (c) are tracked in t1867. The unqualified "repository common root" wording does not mention the bare / separate-git-dir fallback (D2).
- Handler comment (mcp_factory_card.go:248-254) matches the code: `root` = `factoryRootOf(launch)` for record/config, `integrationCommonRootOf(launch)` for window/re-measure.
- Scope: `git diff 11e07d2c5f..e872022eb3 --name-status`: M progress.md, M CHANGELOG.md, M factory_mcp_verb_parity_test.go, M integration.go, A integration_common_root_test.go, M mcp_factory_card.go. `factory_card.go` diff: 0 lines. No `t.Skip`/timeout line added (grep over added lines hits only the progress record's `golangci-lint --timeout=5m` text); the parity test file has 0 removed lines.

### Convergence check

`audit_multi` (codex required; claude and glm gates off; `plan_source: config`; card_id t1680, target baseBranch, project_root = this tree) returned `overall_verdict: pass`, `per_backend_verdicts: [{backend: codex, gate: required, verdict: pass}]`, no `gate_unmet`, `audit_receipt: rcpt-cf99d57ac958b0cb3612a404`. Digest written to `.moai/state/audit-plan-result.json` (untracked), then `moai verify audit-plan --project-root <tree> --result-file .moai/state/audit-plan-result.json` ->

```
"convergence_check": {
    "ok": true,
    "unmet": []
}
```

### Baseline-attribution

All results above were produced in this session against tree HEAD e872022eb3b2269fe59d5474c8131a24d675dabe with Go building from the tree; overlays and scratch live under /private/tmp/t1680-audit2/; nothing in the tree was edited by me except this file and the untracked `.moai/state/audit-plan-result.json` (`git status --short` empty before the writes). The `moai` CLI and MCP server that produced the audit-plan and audit_multi output are build ee981a932, older than HEAD.

### Gaps

- I did not re-run codex's own scratch repro; the two failures it reported are reproduced by the repo's own `NonParentEnv` table against the pre-fix overlay, which exercises the same flows.
- The pre-fix overlay used 11e07d2c5f's files (the tree state round 1 judged, plus docs); I did not rebuild at 5ebcf93a62 separately.
- Not run: the full `internal/cli` suite (barred by ruling 35e9), `internal/config`, other packages; `golangci-lint` not run by me (the progress record cites `0 issues.`, which I did not re-observe); `-race` not run.
- The go-test-heavy slot was not acquired; the two regression runs took 174 s and 129 s and were each one process.
- D3 rests on code reading of the hook write site and `integrationSessionID`; I did not run a degraded-path repro.
- `/verify` wrapper not invoked. No refused command this round except the first heredoc-based mutant generator (refused by the worktree guard as too complex); I wrote the script to a file under /private/tmp and ran it plainly.

### Residual-risk

- If a lane's layout is bare or separate-git-dir, the window is per linked tree on both surfaces, so two lanes in different trees would not exclude each other (pre-existing for the verb).
- D1: a future edit dropping the handler's helper call is invisible to the table.
- D3: a lane with `CLAUDE_PROJECT_DIR` naming a linked tree and no `CLAUDE_CODE_SESSION_ID` could attribute the window to the session recorded in the primary's single-slot sidecar.
- The judging tool build is older than the tree; the codex review read the current tree's diff text, so the effect on this verdict is limited to the tool server, not the diff.

## Iteration history

Round 1 (11e07d2c5f, audited 950a3e7cba): FAIL 0.78, codex required fail, blocking F1. Round 2 (this file, delta only): PASS-WITH-DEBT 0.90, codex pass, F1 resolved, 0 blocking, 5 optional debts (D1-D5).
