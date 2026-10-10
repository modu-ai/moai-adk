# SPEC-LOCAL-MAIN-FLOW-001 — Progress

Plan-phase skeleton. The four §E headings are the canonical markers, in this order. Each phase writes only its own section.

## §E.1 Plan-phase Audit-Ready Signal

```yaml
audit_ready: true
```

## §E.2 Run-phase Evidence

Range: `0106da184..af8199f8f` (`git log --format='%h %s'`, read at HEAD `af8199f8f` before this commit). Excluded: the plan-phase tip `0106da184` and the absorb merge `e32f69c46`. Subjects are verbatim; each group is in chronological order.

**M1** (REQ-LMF-001, 002, 003, 004, 006, and the primary re-sync of REQ-LMF-014; `eb3c477c0` is a one-line R2_D correction and is part of the M1 list below)
- `fe2362f5d` test(SPEC-LOCAL-MAIN-FLOW-001): RED tests for primary-merge ignored-content shapes
- `8526e72de` test(SPEC-LOCAL-MAIN-FLOW-001): RED tests for integration surface, configuration, and verb
- `c6dab38f9` test(SPEC-LOCAL-MAIN-FLOW-001): repair merge RED tests to observe the step
- `dfee70154` test(SPEC-LOCAL-MAIN-FLOW-001): RED tests for the primary re-sync with compile-only stubs
- `e0f132a56` feat(SPEC-LOCAL-MAIN-FLOW-001): dirty-primary guidance and post-merge symbolic HEAD hold
- `ccfa83598` feat(SPEC-LOCAL-MAIN-FLOW-001): local-main integration config key, shipped off
- `237432a1a` feat(SPEC-LOCAL-MAIN-FLOW-001): primary integration surface behind the local_main_integration gate
- `d0d7901a5` feat(SPEC-LOCAL-MAIN-FLOW-001): factory complete lands the card through the primary when the gate is on
- `a65a4f179` feat(SPEC-LOCAL-MAIN-FLOW-001): primary re-sync and the integration resync verb
- `eb3c477c0` test(SPEC-LOCAL-MAIN-FLOW-001): case-insensitive provisioning match in TestR2_D (card t1616)
- `7fe19c387` test(SPEC-LOCAL-MAIN-FLOW-001): holder-calls-resync keeps its window (card t1616)
- `cbbae93de` fix(SPEC-LOCAL-MAIN-FLOW-001): resync keeps a window the caller already held (card t1616)

**M2** (REQ-LMF-005, separate worktree surface)
- `5db413a58` test(SPEC-LOCAL-MAIN-FLOW-001): RED tests for the separate-worktree overlap decision (card t1616)
- `c08605b44` feat(SPEC-LOCAL-MAIN-FLOW-001): separate-worktree surface uses the B5 overlap decision (card t1616)

**M3** (REQ-LMF-008, 009; M3a `5e202adf1` lands before M3b `1236348ac`)
- `5e202adf1` chore(SPEC-LOCAL-MAIN-FLOW-001): auto_merge off before develop_branch moves to main (card t1616)
- `1236348ac` chore(SPEC-LOCAL-MAIN-FLOW-001): local-main integration on, develop_branch main (card t1616)

**M4** (REQ-LMF-010 to 014)
- `9c7a56fe0` docs(SPEC-LOCAL-MAIN-FLOW-001): AGENTS.md branch wording from the template (card t1616)
- `80ecc6145` docs(SPEC-LOCAL-MAIN-FLOW-001): supersede markers in AGENTS.local.md (card t1616)
- `c556cf063` docs(SPEC-LOCAL-MAIN-FLOW-001): AGENTS.local.md §4.0 default flow, local-main integration (card t1616)
- `34739c6ef` docs(SPEC-LOCAL-MAIN-FLOW-001): gitflow-lane-protocol.md local-main flow (card t1616)
- `363e6ff83` docs(SPEC-LOCAL-MAIN-FLOW-001): repo-local-pr-policy.md local-main flow (card t1616)
- `dea179c6b` docs(SPEC-LOCAL-MAIN-FLOW-001): gitflow-integration-chain.md local-main flow (card t1616)
- `af8199f8f` docs(SPEC-LOCAL-MAIN-FLOW-001): hns-release-specialist.md local-main flow (card t1616)

**No milestone** (artifact commits: measurement pins and operator verdicts)
- `7642093a0` docs(SPEC-LOCAL-MAIN-FLOW-001): measurement pins against 0106da184 and the audit-ready signal
- `9ca36a428` docs(SPEC-LOCAL-MAIN-FLOW-001): operator verdicts 3fbd in decision-index (card t1616)

**M5**: satisfied by the AGENTS.local.md §4.0 section committed in `c556cf063` (plan §C); no separate commit.

**M6** (REQ-LMF-015): no commit. Specification only (plan §G.2 and §G.3); no code, no hook, and no measurement procedure run in this card. Counts observed in this commit against plan.md: `^### G\.[23] Gap` 2, `Run-phase procedure` 2, `^Open:` 2. M7 is operator-gated and not required.

**Verification at af8199f8f**: taken from the lane progress record, section 64 (`.moai/reports/t1616/progress.md`, gitignored); the command outputs are in the evidence files `.moai/reports/t1616/verify-*-af8199f8f.txt`. This commit did not re-run them.
- V2 13/13 PASS; V3 3/3 PASS; V4 43/43 PASS; V5 9/9 PASS; V6 `make build` exit 0 with the tree still clean; V8 exit 0; V11 empty diff.
- V1 is red as raw: one gofmt violation in `internal/factory/integration_remeasure_test.go`, a file the card did not change (Gap 9).
- V7 is not run at af8199f8f; it is measured at the final state after sync. V10 is not observed (Gap 13).

**Gaps** (numbers and titles; detail in lane record §63 items 1-12 and §64 item 13)
1. §4.0 adds +3,332 characters to an always-loaded file
2. draft → in-progress missing at M1; applied at this run-close commit
3. REQ-LMF-014 cites REQ-LMF-011 inconsistently
4. Bodies of `34739c6ef`, `363e6ff83`, `dea179c6b` carry no card id
5. Trailer layout: `card:` and `Authored-By-Agent:` are not parsed as git trailers (leader kept the layout)
6. decision-index.md header and plan §H still list Q1 and Q2 as open
7. §11 re-sync decision formula not checked against the code
8. lane-protocol §6 target rule for card done; done is blocked until t1621
9. V1 pre-existing gofmt violation in `internal/factory/integration_remeasure_test.go`: closed by `b542c1e57` (format-only, one line in that file; `gofmt -l internal/factory` prints nothing)
10. Process deviation: three git commands chained with `;` in one read-only Bash call
11. OQ-7 dependent markers still open (lane-protocol line 153; hns-release-specialist.md, 18 lines)
12. plan §E slot label is wrong for the two path-scoped rule files
13. V10 (PreToolUse hook verdict on a temporary primary): route A observed in `e831304a8` (`TestBranchGuardRouteAIntegrationVerb`, `internal/hook/branch_guard_integration_verb_test.go`). The test calls the in-process `preToolHandler.Handle` with the branch guard ON and `workflow.local_main_integration.enabled` ON, on a primary-checkout fixture (`primary=true`). `go test -count=1 -v`: 1 test and 2 subtests, all PASS. Verdict lines, verbatim from the run (full output in the gitignored scratch `.moai/reports/t1616/f8-route-a.txt`):
    - `verdict case=ALLOWED command="moai integration merge --card t1616" decision="allow" reason=""`
    - `verdict case=DENIED command="git merge --no-ff WT-10-10-class" decision="deny" reason="BRANCH_GUARD_VIOLATION: git merge in primary checkout (use a worktree; do not route around this by naming a spawned agent manager-git - the identity exemption does reach spawned agents, and using it that way defeats the guard; the MOAI_BRANCH_GUARD_EXEMPT sentinel is main-thread-only)"`
    - Negative control (branch guard OFF in the fixture; scratch `.moai/reports/t1616/f8-red-guard-off.txt`): the DENIED subtest FAILs with `decision = "allow", want "deny" (reason="")`, so the deny depends on the guard. The guard setting was then restored; the committed file is the restored version and re-runs to PASS.
    - Limits: the test calls the in-process hook handler. It does not exercise the installed moai binary's argv path or the Claude Code runtime hook wiring. The hook handler does not read `workflow.local_main_integration.enabled` (the CLI merge-target resolver reads it), so the flag is set in the fixture only and changes no hook verdict here.
    - Route B (an operator-terminal end-to-end probe) is still open and is on the operator-return list (leader board decision id d-20261010T094858Z-2793).

**Repair round (2026-10-10)**: the commits of the repair round in order; subjects verbatim from `git log --format=%h %s`:
- `f3175f00d` test(SPEC-LOCAL-MAIN-FLOW-001): regression tests for sync-audit defects F1-F6 (card t1616)
- `c3309ed2e` fix(SPEC-LOCAL-MAIN-FLOW-001): refuse a pre-existing merge and abort only our own (card t1616)
- `a3606dc7f` test(SPEC-LOCAL-MAIN-FLOW-001): RED for a merge another actor begins after the probe (card t1616)
- `afe74bc1b` fix(SPEC-LOCAL-MAIN-FLOW-001): do not abort a merge another actor began after the probe (card t1616)
- `73d7e3e0d` fix(SPEC-LOCAL-MAIN-FLOW-001): re-sync validates before the move, takes the fetched baseline, keeps holder metadata, holds after a post-merge anomaly (card t1616)
- `b542c1e57` style(SPEC-LOCAL-MAIN-FLOW-001): gofmt integration_remeasure_test.go (card t1616)
- `5a3eed857` test(SPEC-LOCAL-MAIN-FLOW-001): RED for a foreign fetch that moves FETCH_HEAD under the re-sync (card t1616)
- `b677c6741` fix(SPEC-LOCAL-MAIN-FLOW-001): re-sync takes BASELINE_SHA from an explicit origin/main fetch, not FETCH_HEAD (card t1616)
- `e831304a8` test(SPEC-LOCAL-MAIN-FLOW-001): route-A observation of the integration verb under the branch guard (card t1616)

## §E.3 Run-phase Audit-Ready Signal

```yaml
audit_ready: true
```

Status transition: draft → in-progress at this run-close commit (Gap; ruling d-20261010T062901Z-6d3d, item 2).

## §E.4 Sync-phase Audit-Ready Signal

```yaml
audit_ready: true
sync_commit_sha: 4d039f2dc
```

Status transition: the implemented → completed transition is made by this sync commit, and the sync_commit_sha backfill follows in the next commit.

**Sync record**: The b677c6741 body's "Pre-commit gate" PASS phrase reports direct executions by manager-develop of golangci-lint, gofmt -l, go vet, and go test, because no git hook runs in this repository (`core.hooksPath=/dev/null` in `.git/config`), and its evidence is in the gitignored `.moai/reports/t1616/gate-fetch-*.txt` and `green-fetch-*.txt`. The codex review gate key is `workflow.codex.review_gate.enabled` in `.moai/config/sections/workflow.yaml` (`grep -n -A1 'review_gate:'` finds it; line numbers vary by tree), and its value at HEAD is `enabled: true`. The e831304a8 branch-guard observation does not depend on `workflow.local_main_integration.enabled`, because the hook package's non-test code does not read that key; the key's non-test readers are in `internal/config`, `internal/cli/integration_merge.go`, and `internal/cli/local_main_resync.go`, and Route B is not performed and is on the operator-return list (leader board decision id d-20261010T094858Z-2793). F13 disposition, decision (A) (ruling 060a): no lint.skip is added to spec.md. The finding is a documented false positive: the lint ends the subtest name at the apostrophe in window_naming_the_card's_own_branch. Because the SPEC is completed, the warning is Advisory in the baseline count, and a skip would hide only future real findings. Follow-up card t1667 carries the root fix, and the other baseline findings (+5 VacuousTestAssertion and +1 StatusTransitionInvalid, from other SPECs) do not block this card's local landing and are a CI re-baseline target (operator decision).

**Repair round (2026-10-10)**: F5 is fixed, with RED 975a59294 holding test `TestLocalMainResyncPostFastForwardLateTxtRefuses`, which uses the untracked file `late.txt` (porcelain line `?? late.txt`). GREEN 9d438dc97 reads the status after the fast-forward; a non-empty result writes the policy hold with the cause and the fast-forward target SHA, then refuses with `MergeExitMergeDirty` (class 7). F1's same-SHA case is fixed by a local check: the merge call's exit status is 1 only when this call began the merge and stopped, and any other status with `MERGE_HEAD` present is refused, not aborted (RED 503a7b073; GREEN cee04a872). This behavior is observed on git 2.54.0 and is not derived from git's source. Accepted residuals, per the leader's rulings: F2, the window between the pre-move re-validation and the merge command, with follow-up card t1670 (atomic re-sync under the factory lock, depends on t1616); and the F1 window that git's lack of an atomic abort leaves. Gate wording: the commit body of 9d438dc97 says "pre-commit gate", but no git hook runs in this repository (`core.hooksPath=/dev/null`), so those results are direct executions. The body of 9d438dc97 says the gate was scoped to `internal/cli/local_main_resync.go`, but lint covered the whole `internal/cli` package. The gate's test row of cee04a872 exceeded its 60-second budget; the explicit race run of 36 top-level PASS in internal/factory (green-race-mergestep-final.txt) is the merge-step test verdict.

Gap: tool-build VCS stamp mismatch: `go version -m bin/moai` on the worktree build reports `vcs.revision=2aab5f797b75983e132af451da68f69e3426557b` and `vcs.modified=true` (its ldflags-injected `Commit` reads `af8199f8f`), while the tree HEAD of this measurement is `bd88ff760`; `2aab5f797` is a strict ancestor of `bd88ff760` (`git rev-list --left-right --count 2aab5f797...bd88ff760` reads `0 196`); the cause is not established.
