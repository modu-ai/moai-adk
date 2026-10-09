# SPEC-LOCAL-MAIN-FLOW-001 — Implementation Plan (v0.7)

Measurements in revision 0.6 were taken at HEAD `366b45155` (the revision 0.5 commit), with the revision 0.6 edits to the SPEC directory uncommitted. Revision 0.7 changes the V4 block (§F) and re-runs E-69 with that block; its measurements are recorded in acceptance.md §E. Cells pinned to `e32f69c46` measure code that is unchanged at `366b45155` (cell E-66). Each anchor in §A carries one provenance label: `[cell E-nn]` means the anchor is backed by that cell in acceptance.md §E; `[read]` means it was read at `e32f69c46` without a ledger cell; `[carried]` means it comes from the card base `2aab5f797` and was not re-measured in this revision. The §A gaps are stated after the table. The lane decisions of 2026-10-10 (MI-5, OQ-3, OQ-6, OQ-8) are applied in this revision and recorded in §H. Claims that depend on a ref are labeled `[local]` (the local repository: HEAD, local main, the primary checkout) or `[origin]` (a remote-tracking ref, or what a remote holds). BASELINE_SHA, the origin/main value recorded at pre-flight, is used only for `[origin]` claims. Priority order follows decision reversibility: §B lists the choices most likely to change, §C the milestones that depend on them.

## A. Context and anchors

| Location | Verified content | Provenance |
|----------|------------------|------------|
| `internal/cli/integration_merge.go:24-25` | the landing verb: `Use: "merge --card <id> [--session <id>]"`; `Short` says "holder only" | [read] |
| `internal/cli/integration_merge.go:100` | the landing verb calls `factory.RunMergeStep` and prints the merge SHA; it does not call `completeTransitions`, so it does not close the card | [read] |
| `internal/cli/integration_merge.go:128-141` | `integrationMergeWorktree`: no-tree refusal at 130–132; primary refusal at 137–139 | [read] |
| `internal/cli/factory_card.go:1754` | `Use: "complete <card> [remeasure]"` (lane session only) | [carried] |
| `internal/cli/factory_card.go:1883-1895` | integration-tree provisioning: `factorySameTree(integTree, primary)` refusal at 1893–1895 | [carried] |
| `internal/cli/factory_card.go:1924-1960` | adoption path: a merge commit whose second parent is the card tip (1934); a valid re-measure record for the tip tree (1936); adopted transitions and window release (1941–1959) | [cell E-54] |
| `internal/cli/factory_card.go:1962-2010` | non-adoption path: candidate-tree record check (1983, refusal at 1985); merge step call (1993–2004); merge-step refusal mapping (2006–2009) | [cell E-55] |
| `internal/cli/factory_card.go:2012-2060` | `completeTransitions` after either path: T14 then T16; the T16 verifier reads the record keyed to the merge tree (2055–2060) | [cell E-56] |
| `internal/cli/factory_card.go:2138` | `factoryCommitParents` | [carried] |
| `internal/cli/factory_card.go:2834` | empty integration-branch refusal text (`git_strategy develop_branch`) | [carried] |
| `internal/cli/factory_merge.go:101-126, 218` | `moai factory merge` readiness: three recorded checks before the window; it neither merges nor closes | [read] |
| `internal/factory/integration_merge_step.go:243-249` | clean precondition: call at 243 (`gitIntegrationWorktreeClean`), `if !clean {` at 247 | [cell E-05] [cell E-06] |
| `internal/factory/integration_merge_step.go:275-279` | pinned-tree record check: `ReadRemeasureRecord(in.Root, pinnedTree)` at 275, `ValidateRemeasureRecord` at 279 | [read] |
| `internal/factory/integration_merge_step.go:466` | merge invocation `git merge --no-ff --no-overwrite-ignore -q -m <msg> <pinned>` | [carried] |
| `internal/factory/integration_merge_step.go:471-473` | abort, then clean check (call at 472) | [cell E-06] |
| `internal/factory/integration_merge_step.go:533-535` | post-merge clean check | [cell E-06] |
| `internal/factory/integration_remeasure.go:68` | `WriteRemeasureRecord`: the store keyed by tree | [cell E-53] |
| `internal/factory/integration_remeasure.go:550` | `RunRemeasure`: runs the command in the worktree, requires a clean tree and unchanged HEAD | [cell E-53] |
| `internal/factory/integration_remeasure.go:563` | the lane's command runs as `sh -c <command>` | [read] |
| `internal/factory/integration_remeasure.go:610` | `RunRemeasure` writes the record keyed by the HEAD tree (its only call to `WriteRemeasureRecord`) | [cell E-53] |
| `internal/homestate/card_transition.go:157` | T18 edge `{"T18", CardMergedLocal, CardDone, guardNoRemote}` | [cell E-58] |
| `internal/homestate/card_transition.go:203-209` | `isReservedEdge`: pushed to ci-green is reserved | [cell E-26] |
| `internal/homestate/card_transition.go:333` | the reserved-edge refusal consults `isReservedEdge` | [cell E-26] |
| `internal/homestate/card_transition.go:660-713` | `guardPush` and `guardNoRemote`: the no-remote guard at 668 and its refusal at 670 | [cell E-25] [cell E-60] |
| `internal/homestate/card_transition.go:686-688` | T17 remote requirement: `pushed` requires a configured remote | [cell E-59] |
| `internal/homestate/card_evidence.go:51-73` | `hasRemote`; `PushGateTarget` returns `pushed` with a remote, `done` without | [carried] |
| `internal/homestate/card_evidence_readers.go:290-338` | `verifyMerge`: two-parent check; tree identity against the second parent's tree at 313–315; reachability; the record verifier at 323; a lane-supplied remeasure file read at 335 | [cell E-52] [cell E-57] |
| `internal/hook/branch_guard.go:155` | branch-state pattern `\bgit\s+merge\s` | [read] |
| `internal/hook/branch_guard.go:369` | `matchBranchStateCommand` | [read] |
| `internal/hook/branch_guard.go:434` | `protectedCommitPattern` (regex in the fenced block below this table) | [read] |
| `internal/hook/branch_guard.go:451, 480` | `matchProtectedCommitCommand` (451); `checkProtectedCommit` (480) | [read] (480); [carried] (451) |
| `internal/hook/branch_guard.go:978, 1001, 1036, 1111-1118` | `isExemptAgent` (978), `isPrimaryCheckout` (1001), `checkBranchState` (1036), `extractBranchStateCommand` (1111–1118; reads the Bash tool's `command` field) | [read] (1111–1118); [carried] (others) |
| `internal/config/defaults.go:1446` | `AutoMerge` distributed default `false` | [carried] |
| `internal/config/defaults.go:1491` | `DenyCommitsOn: []string{}` (template default) | [read] |
| `internal/config/types.go:497, 544, 758, 809` | `BranchGuard`, `SlotLease` in the workflow config; `BranchGuardConfig`, `SlotLeaseConfig` | [carried] |
| `internal/config/loader_slot_lease.go:19` | loader analog for the new key | [carried] |
| `internal/config/loader_integration_branch.go:116-123` | integration target taken from `DevelopBranch` for manual git-flow | [carried] |
| `internal/cli/integration.go:280-282` | `configuredIntegrationBranch` returns the flow-scoped integration target | [read] |
| `.moai/config/sections/git-strategy.yaml:8, 16` | `manual:` parent (8); `develop_branch: develop` (16) | [cell E-08] |
| `.moai/config/sections/workflow.yaml:151-164` | `worktree:` block; the auto_merge comment at 153; `auto_merge: true` at 164 | [cell E-09] |
| `.moai/config/sections/workflow.yaml:173-178` | `branch_guard`: `enabled: true` at 174 [read]; comment 175–177 (SPEC-MAIN-COMMIT-BAN-001); `deny_commits_on: [main]` at 178 | [cell E-10] |
| `internal/template/templates/.moai/config/sections/workflow.yaml` | template copy; no `local_main_integration` key | [cell E-01] |
| `internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl` | the distributed template has no `develop_branch` line | [carried] |
| `internal/template/templates/AGENTS.md.tmpl` | generic wording: section 2 clause at lines 92–95 [read]; section 3 "Start a new card" paragraph at lines 160–163 [read], including the sentence "Create the fresh tree from the remote default branch." [read]; `remote default branch` once [cell E-13] | [read] [cell E-13] |
| `AGENTS.md:90-96` and `AGENTS.md:161-167` | hardcoded main-flow sentences at lines 91, 94, and 166 | [cell E-11] |
| `AGENTS.md` size | 24,228 bytes and 24,122 characters | [cell E-20] [cell E-20c] |
| `AGENTS.local.md` | tracked by git (`git ls-files` lists it; `git check-ignore` exits 1), so its edits enter commits | [carried] |
| `AGENTS.local.md:175` | the §4.1 heading | [read] |
| `AGENTS.local.md:179, 195, 196, 199` | chain heading (179); GitHub Flow rules 1 and 2 (195, 196); the local-main commit-dead rule (199) | [cell E-41] [cell E-42] [cell E-43] [cell E-44] |
| `AGENTS.local.md:212, 214` | the Factory leader request (212) and the self-dispatch exception (214); both stay unmarked (MI-5) | [cell E-36] [cell E-37] |
| `AGENTS.local.md:216, 218, 219` | integration-window duties (216, 218); the card-PR-only public path (219) | [cell E-38] [cell E-39] [cell E-40] |
| `AGENTS.local.md:217` | the lane merges into main only through the verb (the designed path) | [cell E-50] |
| `.claude/rules/local/gitflow-lane-protocol.md` | §1 commit-dead paragraph; §2 develop integration worktree; §3 window; §4 leader push; §8 slot leases; §9 rc runbook; §10–§11 develop release and refresh; cites `CLAUDE.local.md` 6 times | [cell E-18] (count); headings [read] |
| `.claude/rules/local/repo-local-pr-policy.md:2, 11-13` | develop statements; line 11 is the card-branch point | [cell E-19] (count) |
| `.moai/docs/gitflow-integration-chain.md` (88 lines) | retired-file pointers at lines 3–4; develop commands 11–17; develop pushes 33–43; threshold 36–37; residue section 60–87; `CLAUDE.local.md` 8 times | [cell E-17] (count); ranges [carried] |
| `.claude/agents/harness/hns-release-specialist.md` | release from develop (lines 4, 19–25, 49–56, 68, 81–83, 93–96, 114–120); Phase 9 back-merge (330–356); summary lines 364–365, 390–391, 397, 411–412 | [carried] |
| `internal/cli/integration_merge_worktree_test.go` | `TestIntegrationMergeWorktreeRefusesPrimaryHoldingBranch` (the gate-off regression guard) | [cell E-02] |
| `internal/cli/factory_complete_mwq19_test.go:201` | `TestMWQ19_Scenario5_AdoptionRefusedAfterNewCommit` | [cell E-03] |
| `internal/factory/integration_merge_step_test.go:410` | table case `12 dirty before merge` asserting `MergeExitWorktreeDirty`, inside `TestMergeStepPreMergeCausesReleaseWithDistinctCodes` (line 320) | [read] |
| `internal/cli/session_worktree_automerge_test.go:252-1085` | fourteen `TestAutoMerge*` tests | [read] |
| `.moai/specs/SPEC-FACTORY-MANAGED-CARD-CHILD-001/decision-index.md` | format example for the operator decision record | [carried] |
| `[local]` local main and the absorb | HEAD `e32f69c46`; local main `2aab5f797`; `09a42899c` is an ancestor of HEAD and not of local main | [cell E-33] [cell E-45] [cell E-30] [cell E-23] |
| `[local]` tree identity | `e32f69c46^{tree}` and `09a42899c^{tree}` are both `ddc07c9a3800ce62be9a87d54c06ed145283eec7` | [cell E-46] [cell E-47] |
| `[local]` hook configuration | `git config --get core.hooksPath` is `/dev/null` | [cell E-48] |

```text
protectedCommitPattern, internal/hook/branch_guard.go:434 (read at e32f69c46; a Go raw-string regexp):
(?i)\bgit\s+(commit|revert|cherry-pick)\b
```

**Gaps in §A (not re-measured in revisions 0.3 through 0.6).** The rows marked `[carried]` were not re-measured. Their paths are byte-identical between `2aab5f797` and `e32f69c46` by the earlier `git diff --name-only` check, and the V-commands in §F re-check the paths at run start. The session-exit auto-merge code (`internal/cli/session_worktree_automerge.go:156-304`) and the pre-push hook body (`.git_hooks/pre-push`) were read at grep depth in revision 0.2 and are not re-read here (residual risks R-2 and R-3). The line numbers at `internal/cli/integration_merge_worktree_test.go` beyond the test name were not measured.

## B. Decisions most likely to change

### B1. Landing surface

Decided by the operator and relayed by the leader: the landing surface is the local main of the primary checkout. Develop is retired as the integration surface. Consequences: `develop_branch` becomes `main` (REQ-LMF-008); card worktrees are cut from local main, not from the generic "remote default branch" sentence (B9); the batch release starts from local main (REQ-LMF-014).

### B2. Surface selection by configuration (REQ-LMF-002)

Leader guidance item (a). The surface is resolved from two configured facts and the git state: the configured integration branch name `B` (`git_strategy.manual.develop_branch`) and the gate `workflow.local_main_integration.enabled`. The rule has one implementation, used by the merge verb (`integration_merge.go:128-141`) and by factory complete (`factory_card.go:1883-1895`).

| Case | Holder of `B` | Gate | Surface and result |
|------|---------------|------|--------------------|
| 1 | `B` is empty | any | refuse (today's text, `factory_card.go:2834`) |
| 2 | a separate integration worktree | any | that worktree, REQ-LMF-002 case (a); today's flow; REQ-LMF-005 and REQ-LMF-006 apply |
| 3 | no separate worktree; the primary checkout, HEAD names `B` | false | refuse with guidance and no merge, as today (`integration_merge.go:137-139`); REQ-LMF-002 case (c) |
| 4 | no separate worktree; the primary checkout, HEAD names `B` | true | the primary checkout, REQ-LMF-002 case (b); the designed path of B3, B3a, and B4 |
| 5 | no separate worktree; the primary checkout does not hold `B` (another branch, or none) | any | refuse with guidance (below); REQ-LMF-002 case (c) |

"Local main directly" is case 4 with `B = main`. "A separate integration branch with its own integration worktree" is case 2 with `B ≠ main`; the gate has no effect there.

Guidance for case 5: "No tree holds the integration branch {B}. Provision its worktree (the leader does this at batch start), or check {B} out in the primary checkout from your own terminal. The tool does not switch branches." The no-tree refusal stays unconditional on both gate values, as the existing clause at `integration_merge.go:130-132` does.

### B3. The tool-owned local-main merge and its guard treatment (REQ-LMF-003)

Leader guidance item (b). Performer: the landing verb `moai integration merge --card <id>` (`integration_merge.go:24`), reached through `moai factory complete <card>` (`factory_card.go:1754`) from a lane session. In this design no other path writes the primary checkout's local main, apart from the re-sync in B3a.

Sequence for case 4:
1. The gate is true, and the caller holds the integration window (the existing acquisition, `.claude/rules/local/gitflow-lane-protocol.md` §3).
2. `[local]` Read the primary's HEAD symbolically. Require that it names `B`; refuse otherwise with the case-5 guidance. The tool never switches branches.
3. `[local]` Require a fully clean primary (B4).
4. Pin the card tip (the existing step). The status set before the merge is empty, by step 3.
4a. `[local]` Compute T and I of B5 with the primary as W and the pinned card SHA as P. If I is non-empty, refuse with exit class `MergeExitWorktreeDirty`, release the window, and run no merge. Guidance text (operator-facing, English): "The primary checkout holds ignored file(s) at paths this merge would write: {paths}. The merge would overwrite them. Move or rename them from your own terminal, then re-run. The tool does not remove or stash files." This check is the protection against overwriting ignored files. The flag in step 5 is not: cell E-78 shows that `--no-overwrite-ignore` replaced an ignored file on the no-fast-forward path.
5. Run `git -c merge.autoStash=false merge --no-ff --no-overwrite-ignore -q -m <msg> <pinned>` with the primary as the working directory. The flag is a second guard; step 4a decides whether an ignored file is overwritten.
6. `[local]` Read HEAD again. Require that its first parent equals the pre-merge HEAD. Re-read the symbolic HEAD and require that it still names `B`: SHA equality alone does not prove that the branch did not change, because HEAD could move to another branch at the same SHA between steps 2 and 5. Require that the status set equals the pre-merge set (REQ-LMF-006).
7. `[local]` Record the merge SHA and tree. The closure follows B7.

Guard treatment (an explicit design, not a bypass):
- The PreToolUse guard matches the command text the agent issues (`extractBranchStateCommand`, `branch_guard.go:1111`; `matchBranchStateCommand`, `:369`). The verb's `git merge` runs as a child process of the `moai` binary, not as command text the agent wrote. This follows from the matcher and has not been observed yet. Check V10 observes it in a scratch primary during the run phase.
- Agent-issued `git merge` in the primary stays denied by the branch-state pattern (`branch_guard.go:155`). Agent-issued `git commit`, `git revert`, and `git cherry-pick` on a listed branch stay denied (`:434`, `:480`). Q1 changes only the list those patterns consult.
- `isExemptAgent` (`:978`) and the `MOAI_BRANCH_GUARD_EXEMPT` sentinel are neither used nor changed by the verb.
- The verb never runs checkout, switch, reset, stash, rebase, or any merge other than the two forms named in this plan: the no-fast-forward merge above and the `--ff-only` re-sync of B3a. With autostash disabled it never stashes.
- The verb does not read `deny_commits_on`, and neither does the re-sync of B3a. Their paths are the same under each Q1 option.
- The designed exception is written in AGENTS.local.md §4.0 and in the verb's help text, not in the universal AGENTS.md text.
- The branch guard source is not modified by this SPEC.

### B3a. Re-sync of local main (OQ-8, redirected to a tool-owned fast-forward)

The lane decision of 2026-10-10 redirects OQ-8. The re-sync is tool-owned. It is not an operator-terminal step, and no part of it requires the operator's terminal.

- **Trigger.** `[origin]` A release pull request has landed on `origin/main`, and local main must follow it.
- **Scope.** The primary-checkout surface only (B2 case 4). On case 2 this SPEC changes nothing.
- **Verb.** A fast-forward-only subcommand of the integration verb family (`moai integration`), run inside the integration window. The subcommand's name and flag shape are fixed in the run phase and recorded in the run progress record; this SPEC fixes its behavior.
- **Sequence, inside the window.**
  1. Acquire the window (the existing acquisition).
  2. `[local]` Read the primary's HEAD symbolically and require that it names `B`; refuse otherwise with the case-5 guidance.
  3. `[local]` Require a fully clean primary (B4).
  4. `[origin]` Run `git fetch origin <B>` and observe its exit status before step 5. Record the fetched origin value of `<B>` as BASELINE_SHA. Every later origin-facing comparison in this section uses BASELINE_SHA; no step re-reads the remote.
  5. `[local]` Classify HEAD against BASELINE_SHA with read-only ancestry tests, then act. (a) If HEAD is an ancestor of BASELINE_SHA, first check for ignored files that the fast-forward would overwrite: intersect the paths of `git diff --name-only -z HEAD BASELINE_SHA` with `git ls-files --others --ignored --exclude-standard -z`. A non-empty intersection refuses with guidance that lists the paths; HEAD does not move, and the window is released. Otherwise run `git merge --ff-only --no-overwrite-ignore` with BASELINE_SHA as the target and the primary as the working directory. (b) If BASELINE_SHA is an ancestor of HEAD (local main already contains origin, or equals it), report that no fast-forward is needed and leave HEAD unchanged. (c) If neither is an ancestor of the other (diverged), refuse as described under Refusal below.
  6. `[local]` After case (a), read HEAD again and require it to equal BASELINE_SHA. Re-read the symbolic HEAD and require that it still names `B`, as in B3 step 6.
  7. Report the old and new SHAs, or the no-op of case (b).
  8. Release the window.
- **Refusal (diverged).** Case (c) refuses with guidance: `[local]` local main holds commits that BASELINE_SHA `[origin]` lacks, and the tool does not reconcile them. HEAD does not move, and the window is released. The method for this case is decision Q2 (the diverged-case options in decision-index.md), not an operator-terminal step.
- **Why `--ff-only`.** A fast-forward creates no commit, so the verb cannot write a commit that Q1 governs. It is the only merge form the verb runs apart from the no-fast-forward merge of B3.
- **Guard treatment (the B3 treatment applies unchanged).** The merge is a child process of the `moai` binary, not command text the agent issues. The verb runs no checkout, switch, reset, rebase, or stash. HEAD must name the surface branch (step 2). The `MOAI_BRANCH_GUARD_EXEMPT` sentinel and the `manager-git` identity are neither used nor changed.
- **Verification.** `TestLocalMainResyncFastForwards`, `TestLocalMainResyncRefusesDiverged`, `TestLocalMainResyncAheadIsNoop`, and `TestLocalMainResyncPreservesIgnoredFile` (M1 step 1) run under V4. The last one places an ignored file at a path the fast-forward would write, and asserts that the verb refuses, HEAD does not move, and the file keeps its content. `[origin]` They set `refs/remotes/origin/main` with `git update-ref` in a scratch repository, and no test contacts a remote.
- **Gap.** The production path includes a remote fetch. This SPEC's run phase executes no remote command (spec §E), so the fetch is specified here and not run here.

### B4. Fully clean primary: refusal and guidance (REQ-LMF-004)

Leader guidance item (c). Choice: on the primary surface, the tool requires `git status --porcelain=v1 -z --untracked-files=all` to be empty. This is stricter than the overlap rule of B5, and that is deliberate. The primary checkout is shared by concurrent sessions (`main-checkout-branch-guard.md`, "Why the race is quiet"). A process-registry lookup is not a reliable ownership signal (`main-checkout-branch-guard.md`, "Detecting Concurrent Sessions"). The lane therefore cannot tell whose uncommitted path it would leave beside the merge, and an overlap test cannot judge a path whose owner it cannot identify. A dedicated integration worktree has no foreign owners, so the overlap rule applies there (REQ-LMF-005). The clean check does not list ignored files, so the primary surface checks them separately: the overlap check of B3 step 4a refuses when a path the merge would write exists as an ignored file. The flag `--no-overwrite-ignore` in step 5 of B3 is a second guard only, because cell E-78 shows that it does not protect ignored files on the no-fast-forward path. The re-sync carries the same flag and checks the ignored files itself (B3a, step 5); cell E-72 shows that the flag protects them on the fast-forward path.

Placement: the check runs inside the window, before the merge. On refusal the window is released, following the release-on-refusal pattern at `integration_merge_step.go:245` and `:248`.

Refusal: exit class `MergeExitWorktreeDirty`, for an uncommitted or untracked change, and for an ignored overlap under B3 step 4a.

Guidance text (operator-facing, English): "The primary checkout has {N} uncommitted or untracked change(s). The local-main merge needs a fully clean primary, because changes owned by other sessions cannot be told apart from yours. Move your changes into a card worktree and commit them there, or ask the session that owns them to commit or discard them. Do not stash: the stash is repository-wide. Then re-run."

### B5. Overlap and status-set definitions (REQ-LMF-005 and REQ-LMF-006)

Let W be the integration worktree and P the pinned card SHA. All reads are read-only commands inside W.

- **D (dirty set).** Records of `git status --porcelain=v1 -z --untracked-files=all`. Each record is `XY`, a space, and a path, NUL-terminated. When X or Y is `R` or `C`, the following NUL-terminated field is a second path (the source). Both belong to D.
- **T (target set).** Paths of `git diff --name-only -z --no-renames B0 P`, where `B0` is `git merge-base HEAD P`. Additions, modifications, and deletions are all included.
- **I (ignored overlap).** Paths in T that exist in W as ignored files: the intersection of T with `git ls-files --others --ignored --exclude-standard -z`.
- **Overlap(D, T).** True when some d in D and t in T satisfy d equal to t, or d plus `/` is a prefix of t, or t plus `/` is a prefix of d.
- **Decision (separate worktree only).** Refuse with `MergeExitWorktreeDirty` if and only if Overlap(D, T) or I is non-empty. Otherwise proceed. On the primary surface the decision is the full-clean rule of B4 together with the ignored check of B3 step 4a (a non-empty I refuses).
- **Status set S (both surfaces).** The records of `git status --porcelain=v1 -z --untracked-files=all` in W, sorted bytewise by path, compared as exact record bytes. S-before is captured in the same serialized section, immediately before the merge subprocess. S-after is taken after the merge commit. S-abort is taken after `git merge --abort`. A difference from S-before at either point is classified with the existing hold classes: `MergeExitPostMerge` after the merge, `MergeExitMergeDirty` after the abort.
- **Merge invocation (both surfaces).** `git -c merge.autoStash=false merge --no-ff --no-overwrite-ignore -q -m <msg> P`. Disabling autostash makes the status-set comparison independent of the user's git configuration. Without it, git could stash and reapply disjoint changes and change the status set. On the no-fast-forward path the flag does not protect ignored files (cell E-78), so the overlap check of B3 step 4a does.

### B6. Key shape and default (REQ-LMF-001)

`workflow.local_main_integration.enabled` (boolean). The nested block follows the convention of `branch_guard` and `slot_lease` (`types.go:497, 544`). The template default is `false`, following the opt-in precedent of `AutoMerge` (`defaults.go:1446`). A flat key has no precedent in the workflow block. Reusing `branch_guard` would change that guard's meaning, so it is rejected.

### B7. Closure rule and implementation boundary (REQ-LMF-007) — tool-owned

Current behavior, read at `e32f69c46`:
- The no-remote edge T18 (`card_transition.go:157`, cell E-58) refuses when any remote exists (`:668-670`, cells E-25 and E-60). Its admission requires the human decider (`:661-663`) and verification of the leader approval receipt (`:676`).
- The merged-local to pushed edge requires a remote (`:686-688`, cell E-59). `PushGateTarget` (`card_evidence.go:64-73`) returns `pushed` when a remote exists and `done` otherwise.
- The reserved pushed to ci-green edge (`:207-209`, consulted at `:333`, cells E-26) is refused by this code. So with a remote configured, the route to done is closed at two points.
- The closure happens only in `factory complete`, on either merge path: the adoption path (`factory_card.go:1924-1960`) and the non-adoption step (`:1962-2010`), both through `completeTransitions` (`:2039`). The landing verb alone does not close a card (`integration_merge.go:100`). A merge made through the verb is therefore picked up by the adoption path (see R-7 for the state a card is left in between).

`[local]` This repository has origin configured, so while the refusal at 668-670 stands, merged-local → done is not reachable on this card alone; opening done is the sibling card t1621's scope.

Decided rule (tool-owned on both merge paths; the lane supplies no evidence). For a merged-local card whose merge is on the local-main surface (REQ-LMF-002, case 4):

- (a) `[local]` The merge commit has two parents, its tree equals its second parent's tree (`card_evidence_readers.go:313-315`, cell E-52), and the merge is reachable from the integration branch (a read-only ancestry test).
- (b) The store record keyed by that tree passes `ReadRemeasureRecord` (`factory/integration_remeasure.go:81`) and `ValidateRemeasureRecord` (`:128`). The record is written only by `RunRemeasure` (`factory/integration_remeasure.go:550`, writing at `:610`; cell E-53).
- (c) The decider is human and the leader approval receipt verifies (the existing T18 rule).
- (d) Both merge paths apply (a) through (c) the same way. The adoption path reads the record for the integration tip's tree (`factory_card.go:1936`, cell E-54). The non-adoption path checks the candidate-tree record before its merge (`:1983`, cell E-55), and under (a) that key is the merge tree's key too; the T16 verifier reads it after the merge (`:2056`, cell E-56).
- (e) The lane supplies no evidence file on this surface. `card_evidence_readers.go:335` reads a lane-given remeasure path when one is given, after the record check has run (`:323`, cell E-57). The record check is mandatory whether or not a path is given, and a lane file never substitutes for it. The gate on this surface reads the store.
- (f) No push is performed and none is required.

When (a) through (d) hold, the T18 remote refusal does not apply, and `PushGateTarget` returns `done` for this surface. Remote-delivered cards keep the T17 edge and the reserved edge unchanged.

Limits of the rule, recorded as residual risks in §K: the record's command is declared by the lane (`RunRemeasure` runs `sh -c <command>`, `integration_remeasure.go:563`), and the record is authenticated by its tree key and build identity rather than by a signature (the trust model at `integration_remeasure.go:17-20`). That is R-1.

Interaction with `deny_commits_on`: none. The closure writes no commit and runs no commit verb. It reads refs and records the factory row. The rule holds under each Q1 option. The re-sync (B3a) does not change the closure.

Closure is not disposal. The card worktree's branch has no remote copy until the release batch that contains the merge lands on origin. AGENTS.md section 3 ("dispose of no worktree until the branch is integrated and the remote merge has landed") stays binding. For this surface, "the remote merge has landed" means that the batch containing the card's merge has landed on origin, as AGENTS.local.md §4.0 states.

Boundary, stated explicitly:
- This card decides the rule and records it in this plan (§B7, §G.1), in AGENTS.local.md §4.0 (M4), and in spec.md (REQ-LMF-007).
- This card does not modify `internal/homestate/card_transition.go` or `internal/homestate/card_evidence.go`.
- Sibling card t1621 owns the implementation: the T18 admission for this surface, the `PushGateTarget` change, the repair of existing merged-local records, opening the done transition, and the removal of the lane-file read on this surface (cell E-57).
- The RED observations that t1621 must flip are E-25 (the refusal at `card_transition.go:670` is live at this tree) and E-26 (the reserved-edge refusal at line 333).

### B8. auto_merge reconciliation (closed by the lane decision; residual risk R-2)

The comment at `workflow.yaml:153` reads "auto_merge 유지 (사용자 수동 worktree 생성 시에만 자동 머지)", which means "keep auto_merge; merge automatically only when the user creates the worktree manually."

- `sessionExitAutoMerge` (`session_worktree_automerge.go:156`) reads `workflow.worktree.auto_merge` (line 162). It takes its target from `develop_branch` (lines 176–183), takes the integration window with the target tree as worktree (230–246), and runs `git merge` in that tree (284). [carried]
- With `develop_branch: main`, the only tree holding main is the primary checkout. The auto path has no primary-checkout refusal. The refusal in REQ-LMF-002 lives in the merge verb, not in this path. An enabled auto-merge would therefore run `git merge` in the primary checkout, which AGENTS.md section 2 forbids.
- The comment's premise (automatic merge on manual worktree creation into the integration branch) does not hold under the local-main flow.

Closed by the lane decision of 2026-10-10 (former OQ-3): the value is `true` in this repository today (`workflow.yaml:164`). REQ-LMF-008 sets it to `false` in M3, the template default is `false` (`defaults.go:1446`), and this card makes no code change to `internal/cli/session_worktree_automerge.go`. Until M3 lands, the path stays enabled. The path's missing primary-checkout refusal is residual risk R-2 (§K). This card adds no acceptance criterion for it. Re-enabling the switch is outside this card and needs a later SPEC; that SPEC would decide whether the feature should exist for a local-main target at all.

The replacement comment for line 153 reads, in English:

```yaml
        # auto_merge: false. Session-exit auto-merge merges into
        # git_strategy.manual.develop_branch (main in this repository). The only
        # tree holding main is the primary checkout, and the auto path has no
        # primary-checkout refusal (internal/cli/session_worktree_automerge.go).
        # Re-enable only after a SPEC gives that path a primary-aware refusal.
```

### B9. AGENTS.md wording and the repository's branch-point statement (REQ-LMF-010, REQ-LMF-011)

Section 2: AGENTS.md lines 90–96 (564 bytes; cell E-35 measures the clause) are replaced by template lines 92–95 (370 bytes, carried), so the paragraph reads as the template does. Section 3: AGENTS.md lines 161–167 (602 bytes, carried) are replaced by template lines 160–163 (348 bytes, carried). The mapping of the template lines to sections is MI-4, confirmed by the lead on 2026-10-10 (§H): lines 92–95 go in §2, and lines 160–163 go in §3, including the sentence "Create the fresh tree from the remote default branch."

The template's section 3 sentence says "Create the fresh tree from the remote default branch." This repository cuts card worktrees from local main, the integration tip, because origin/main lags the local-main-only commits. The repository's statement goes into AGENTS.local.md §4.0, which names local main as the card branch point and overrides that sentence explicitly for this repository. The distributed template is not changed.

### B10. Supersede set (REQ-LMF-011) — seven markers (MI-5 resolved)

Markers go in place on seven lines of AGENTS.local.md, the set the lane decision of 2026-10-10 fixes:

- 179: the §4.1 chain heading (cell E-41).
- 195: rule 1, a GitHub Flow rule: card branches start from main (cell E-42).
- 196: rule 2, a GitHub Flow rule: main is the remote default branch (cell E-43).
- 199: rule 5, the local-main commit-dead rule, which becomes a configuration choice under Q1 (cell E-44).
- 216: an integration-window duty (cell E-38).
- 218: an integration-window duty of the lane (cell E-39).
- 219: the card-PR-only public path, which contradicts REQ-LMF-014 (cell E-40).

That is seven markers, and the ledger cells E-38 to E-44 are exactly these seven lines. Lines 212 and 214 are not superseded; the rationale is in §H. Markers are added and no line is deleted.

### B11. Operator decisions (not decided here)

Q1 (the value of `deny_commits_on`) and Q2 (the execution method of the first merge into local main, and the method for a diverged local main during the re-sync) are recorded in `decision-index.md` with their options, consequences, and citations. Q2's diverged-case options are the only part of the re-sync that remains open; its fast-forward design is B3a. This plan does not choose either option and applies no default. AC-LMF-009 is gated on Q1. AC-LMF-016 is gated on Q1 and Q2. The last-write analysis for each Q2 option is in §J.

## C. Milestones

### ABS-0 (done) — absorb of 09a42899c

Commit `e32f69c46` merges `09a42899c` into WT-10-10-class with `--no-ff`; its parents are `2aab5f797` and `09a42899c`. Verified at this revision: `[local]` `git merge-base --is-ancestor 09a42899c HEAD` exits 0 (cell E-30), and `[local]` `git merge-base --is-ancestor 09a42899c 2aab5f797` exits 1 (cell E-31, the control). Local main is `2aab5f797` (cell E-45), and the trees of `e32f69c46` and `09a42899c` are identical (cells E-46, E-47).

### M1 (High, independent) — REQ-LMF-001, 002, 003, 004, 006, and the primary-surface re-sync of REQ-LMF-014

1. RED first. Add these tests and observe the behavior tests fail on the unchanged code.
   - Configuration: `TestLocalMainIntegrationDefaultsFalse` (including the absent-key case), `TestLocalMainIntegrationReadsEnabled`.
   - Surface: `TestIntegrationSurfaceSelectsPrimaryWhenEnabled`, `TestIntegrationSurfaceRefusesPrimaryWhenDisabled`, `TestIntegrationSurfaceRefusesPrimaryOffBranch`, `TestIntegrationSurfaceRefusesNoHolder`, `TestIntegrationSurfaceSeparateWorktreeUnchanged`.
   - Primary merge: `TestLocalMainMergeMergesIntoPrimary` (with the symbolic-HEAD assertion of B3 step 6), `TestLocalMainMergeRefusesDirtyPrimary`, `TestLocalMainMergeRefusesMovedHead`, `TestLocalMainMergeStatusSetUnchanged`, `TestLocalMainMergePreservesIgnoredFile` (B3 step 4a).
   - Re-sync (B3a): `TestLocalMainResyncFastForwards`, `TestLocalMainResyncRefusesDiverged`, `TestLocalMainResyncAheadIsNoop`, `TestLocalMainResyncPreservesIgnoredFile`.
   - Verb and complete: `TestIntegrationMergeWorktreeAcceptsPrimaryWhenEnabled`, `TestFactoryCompletePrimaryTreeGateEnabled`, `TestFactoryCompletePrimaryTreeGateDisabled`, `TestFactoryCompleteNoIntegrationTreeRefused`.
2. Template key, placed beside the `branch_guard` block in `internal/template/templates/.moai/config/sections/workflow.yaml`:

```yaml
    # local_main_integration: lets the integration verbs use the primary checkout
    # as the tree that holds the integration branch (local-main flow). Shipped
    # default false, which keeps the primary-refusal behavior unchanged.
    local_main_integration:
        enabled: false
```

3. Go. `internal/config/types.go`: add `LocalMainIntegrationConfig` with `Enabled bool` tagged `yaml:"enabled"`, and a field `LocalMainIntegration` tagged `yaml:"local_main_integration"` in the workflow struct beside `SlotLease`. `internal/config/defaults.go`: default `false`. New `internal/config/loader_local_main_integration.go` exposing `LoadLocalMainIntegrationEnabled(projectRoot string) bool`, modeled on `loader_slot_lease.go`.
4. Surface selection. One function resolves the surface (B2) and replaces both refusals: the clause at `integration_merge.go:137-139` and the clause at `factory_card.go:1893-1895`. The no-tree clauses (`integration_merge.go:130-132`, `factory_card.go:1893`, empty-tree half) stay unconditional.
5. Primary path in the merge step (B3 sequence) and the re-sync (B3a, with its three-case classification). The status-set helpers from B5 are used for both surfaces (REQ-LMF-006).
6. Guidance strings (B2 case 5, B3a refusal, B4).
7. `make build` (V6), then the repository copy changes only in M3.

### M2 (High, independent) — REQ-LMF-005 on the separate worktree surface

1. RED first. Add `TestMergeStepDirtyDisjointPathsProceeds`, `TestMergeStepDirtyPathOverlapRefuses` (the table includes the rename-source case), `TestMergeStepDirtyDirectoryPrefixOverlapRefuses`, and `TestMergeStepDirtyIgnoredTargetRefuses`. Observe the disjoint case fail on the current code, which refuses at line 247.
2. Implement D, T, I, and Overlap (B5). Replace the clean precondition at `integration_merge_step.go:243-249` for the separate-worktree surface with the overlap decision. Keep the 12-dirty case (`integration_merge_step_test.go:410`) refusing: its fixture must dirty a path the card changes.
3. Status-set tests: `TestMergeStepStatusSetChangedAfterMergeHolds` and `TestMergeStepStatusSetChangedAfterAbortHolds`. The helper is the one from M1.

### M3 (Medium, ordered after ABS-0) — REQ-LMF-008 and REQ-LMF-009 (repository configuration)

Two commits, in this order (B4 of revision 0.6):

1. Commit M3a, `auto_merge` first. `.moai/config/sections/workflow.yaml:164`: `auto_merge: true` becomes `auto_merge: false`, and the single comment line at 153 is replaced by the B8 comment block. Reason for this order: `sessionExitAutoMerge` reads the switch at `internal/cli/session_worktree_automerge.go:162` and takes its target from `develop_branch` at lines 176–183. Once `develop_branch` is `main`, a path whose switch is still `true` would run `git merge` in the primary checkout. Setting the switch false first closes that window.
2. Commit M3b. `.moai/config/sections/git-strategy.yaml:16`: `develop_branch: develop` becomes `develop_branch: main` (the parent key `manual:` at line 8 is unchanged). The M1 block with `enabled: true` is added beside `branch_guard` in `workflow.yaml`.

`workflow.yaml:178` (`deny_commits_on: [main]`) is not edited (Q1). Each commit satisfies `git merge-base --is-ancestor 09a42899c HEAD` (V8) before it is made.

### M4 (Medium, ordered after ABS-0) — REQ-LMF-010 to REQ-LMF-014 (documents and release procedure)

1. `AGENTS.md`: replace lines 90–96 with template lines 92–95, and lines 161–167 with template lines 160–163 (B9). Do not touch the template.
2. `AGENTS.local.md`: insert `### §4.0 Default development flow (local-main integration)` before the heading at line 175. The section is written in Korean to match its neighbors. It states the landing flow (B2, B3), the re-sync (B3a), the fully-clean rule (B4), the branch point and the override (B9), the closure rule and boundary (B7), the batch release procedure (REQ-LMF-011), and the batch-close rule restated against main (spec.md §G).
3. `AGENTS.local.md`: place the seven markers of B10 in place. Markers are added; no line is deleted. Lines 212 and 214 carry no marker.
4. `.claude/rules/local/gitflow-lane-protocol.md`: correct §1, §2, §4, §6, §7, §9, §10, and §11 to the local-main flow, and replace each `CLAUDE.local.md` pointer with `AGENTS.local.md` §4.1.
5. `.claude/rules/local/repo-local-pr-policy.md`: restate lines 2, 11, and 12 for the local-main flow. Line 13 (no card PRs; lanes do not open PRs against main) stays.
6. `.moai/docs/gitflow-integration-chain.md`: redirect lines 3–4 to `AGENTS.local.md` §4.1; restate the develop commands at 11–17, the lead push at 33–43, and the threshold at 36–37; mark the residue section 60–87 as a historical record.
7. `.claude/agents/harness/hns-release-specialist.md`: correct the release-branch source and the Phase 9 back-merge only as far as OQ-7 allows. Sentences that depend on OQ-7 carry a superseded marker and stay open.
8. Every single edit to `AGENTS.local.md` or to a rule file that grows it by more than 1,000 bytes carries the byte statement in its commit body (§E.2).

### M5 (Medium, independent) — REQ-LMF-007 (closure record)

The rule and the boundary are recorded in §B7 and §G.1 of this plan, in the spec (REQ-LMF-007), and in AGENTS.local.md §4.0 (M4). The RED anchors E-25, E-26, E-52 to E-57, and the citations E-58 to E-60, are recorded in acceptance.md §E. This card contains no implementation of the closure. The implementation is sibling card t1621.

### M6 (Low, independent) — REQ-LMF-015 (gaps b and c)

Specified in §G.2 and §G.3. No code and no hook.

### M7 (operator-gated; not required) — the first merge of this card into local main

Executed only after Q1 and Q2 are decided, by the decided method (§J). No milestone depends on it, and no acceptance criterion except AC-LMF-016 mentions it.

## D. Independence and ordering

- M1 and M2 change only the surface resolution, the merge step, and the re-sync. Their tests build fixture roots under `t.TempDir`, so no value of `develop_branch` or `deny_commits_on` enters them. The RED cells E-01 to E-07 measure code, not configuration values. The re-sync tests set `refs/remotes/origin/main` with `git update-ref` and contact no remote.
- M3 and M4 edit files whose develop-era text came in with the absorb. They are ordered after ABS-0, which is satisfied. Each commit's ancestry is checked (V8).
- M5 and M6 are documentation. M5's implementation is in t1621, so the card has no code dependence on it.
- The decision gates touch only AC-LMF-009 and AC-LMF-016.

## E. Byte deltas (rule-authoring duty)

Baselines are measured at tree `e32f69c46` where a cell exists, and marked carried otherwise. Characters are the figure the 40,000-character ceiling names. Bytes are the figure the rule-authoring statement names.

| File | Always-loaded slot | Bytes now | Characters now | Planned delta | Statement required |
|------|--------------------|-----------|----------------|---------------|--------------------|
| `AGENTS.md` | slot 2 | 24,228 (cell E-20) | 24,122 (cell E-20c) | §2: lines 90–96 are 564 bytes (cell E-35), template lines 92–95 are 370 bytes (carried), −194; §3: lines 161–167 are 602 bytes (carried), template lines 160–163 are 348 bytes (carried), −254; net −448 bytes | No (the file shrinks) |
| `AGENTS.local.md` | read at session start (AGENTS.md section 8); tracked by git | 40,850 (cell E-21) | 28,364 (cell E-21c) | §4.0 insertion about +3,500 to +5,000 bytes (Korean text is three bytes per character; an estimate from the planned text, measured in the run phase); seven markers about +25 bytes each (about +175); pointer edits about 0 | Yes, for the §4.0 insertion |
| `.claude/rules/local/gitflow-lane-protocol.md` | slot 1 (no top-level `paths:`) | measured at run start | measured at run start | each edit kept under 1,000 bytes; planned net at most +600 bytes | only for an edit above 1,000 bytes |
| `.claude/rules/local/repo-local-pr-policy.md` | slot 1 (no top-level `paths:`) | measured at run start | measured at run start | planned net at most +200 bytes | only for an edit above 1,000 bytes |
| template `workflow.yaml` | not an instruction slot | 18,367 (carried from `2aab5f797`) | not measured | about +200 bytes (the key block with its comment) | No |
| template `AGENTS.md.tmpl` | not changed | — | — | 0 | No |

The planned figures are estimates. The run phase measures each file with `wc -c` (bytes) and `wc -m` (characters) before and after each edit, and the commit body records the measured values.

### E.2 Byte-size and cost statement (draft for the §4.0 insertion)

The commit body states the measured before and after byte counts from `wc -c`, and the character counts from `wc -m`. Cost statement: sessions that never create or merge a card pay the §4.0 bytes on every turn and again after each `/clear`. The sessions that need §4.0 cannot be identified by path, because any session can create a card worktree, so path scoping does not apply. The file stays under the 40,000-character ceiling by the planned margin.

## F. Verification plan (lane-local and scoped)

No command in this plan runs the full test suite. Each proving command names touched packages and an anchored `-run` pattern that lists exact test names, so that a name matching nothing or a longer name cannot pass silently. A pass is counted only when the verbose output contains a line `--- PASS: <name> ` (the name followed by a space) for every named test. A run whose output contains `[no tests to run]` fails (verification-completeness §1.1).

| ID | Command | Lease | Proves |
|----|---------|-------|--------|
| V1 | `gofmt -l internal/cli internal/factory internal/config internal/template` | none | formatting; empty output |
| V2 | block V2 below | none | REQ-LMF-001 defaults and reads, struct and YAML symmetry, integration-target guidance |
| V3 | block V3 below | none | the template key honesty guard (the template comment must not claim the key is unread), disclosure, auto-merge checks |
| V4 | block V4 below | none | REQ-LMF-002, 003, 004, 006 on the primary surface; REQ-LMF-014 re-sync (B3a, three cases); REQ-LMF-008 auto-merge toggle and its fourteen tests (B8, no code change) |
| V5 | block V5 below | `moai slot acquire --resource internal-factory-mergestep` before; `moai slot release --resource internal-factory-mergestep` after | REQ-LMF-005 and REQ-LMF-006 on the separate surface (the step family creates git fixtures in bulk, and lane protocol §8 names factory suites as heavy) |
| V6 | `make build` | none (a build, not a suite) | REQ-LMF-001 template regeneration |
| V7 | the file checks of `acceptance.md` §F, and the SPEC lint and audit (cells E-61 to E-65 in `acceptance.md` §E, each with its judging build) | none | REQ-LMF-008 to REQ-LMF-015 |
| V8 | `git merge-base --is-ancestor 09a42899c HEAD` | none | the absorb gate (REQ-LMF-008); required before each M3 and M4 commit |
| V9 | `git status --short` immediately before each staging step | none | explicit pathspec staging only (no sweep staging in the primary checkout) |
| V10 | run phase, scratch primary: feed the PreToolUse event of `moai hook` a Bash payload whose command is `moai integration merge --card t1616`, with the working directory set to a scratch primary whose workflow enables the gate | none | the guard treatment in B3: expected decision allow, with no `BRANCH_GUARD_VIOLATION:` line. The entry point is confirmed with the CLI help before use and recorded in the run progress record |
| V11 | `git diff --quiet e32f69c46 HEAD -- internal/template/templates/AGENTS.md.tmpl` | none | the distributed AGENTS template is unchanged by this card (exit 0) |

Proving blocks. Each block is a fenced shell sequence whose selector is plain `|` alternation inside single quotes, so the bytes in the block are the bytes the shell receives (verification-completeness §2.1). The first command counts the named tests that exist, with `-list`; `grep -c` exits non-zero on an empty list, so the `&&` stops the block before any run. The second command runs the names verbosely, and each named test must show a line `--- PASS: <name> ` (the name followed by a space). The required count of each block is the number of names in it: V2 13, V3 3, V4 40, V5 9. Before M1 and M2, the count is a count observation, recorded in acceptance.md §E (cells E-67 to E-70); the names that M1 and M2 add are not yet in the tree, so the observed count stays below the required count until they land.

V2 (required count 13):

```bash
go test ./internal/config/ -list '^(TestLocalMainIntegrationDefaultsFalse|TestLocalMainIntegrationReadsEnabled|TestStructYAMLSymmetry|TestStructYAMLSymmetry_Constitution|TestStructYAMLSymmetry_Context|TestStructYAMLSymmetry_Interview|TestStructYAMLSymmetry_Design|TestStructYAMLSymmetry_Statusline|TestStructYAMLSymmetry_GitConvention|TestStructYAMLSymmetry_Gate|TestEmptyTargetGuidance|TestEmptyTargetGuidanceResolvedTargetIsSilent|TestTargetProvenance)$' | grep -c '^Test' && go test ./internal/config/ -run '^(TestLocalMainIntegrationDefaultsFalse|TestLocalMainIntegrationReadsEnabled|TestStructYAMLSymmetry|TestStructYAMLSymmetry_Constitution|TestStructYAMLSymmetry_Context|TestStructYAMLSymmetry_Interview|TestStructYAMLSymmetry_Design|TestStructYAMLSymmetry_Statusline|TestStructYAMLSymmetry_GitConvention|TestStructYAMLSymmetry_Gate|TestEmptyTargetGuidance|TestEmptyTargetGuidanceResolvedTargetIsSilent|TestTargetProvenance)$' -v -count=1
```

V3 (required count 3):

```bash
go test ./internal/template/ -list '^(TestWorkflowWorktreeKeyHonesty|TestAgentsDisclosureCompleteness|TestAutoMergeRequiredChecks)$' | grep -c '^Test' && go test ./internal/template/ -run '^(TestWorkflowWorktreeKeyHonesty|TestAgentsDisclosureCompleteness|TestAutoMergeRequiredChecks)$' -v -count=1
```

V4 (required count 40):

```bash
go test ./internal/cli/ -list '^(TestIntegrationSurfaceSelectsPrimaryWhenEnabled|TestIntegrationSurfaceRefusesPrimaryWhenDisabled|TestIntegrationSurfaceRefusesPrimaryOffBranch|TestIntegrationSurfaceRefusesNoHolder|TestIntegrationSurfaceSeparateWorktreeUnchanged|TestLocalMainMergeMergesIntoPrimary|TestLocalMainMergeRefusesDirtyPrimary|TestLocalMainMergeRefusesMovedHead|TestLocalMainMergeStatusSetUnchanged|TestLocalMainMergePreservesIgnoredFile|TestLocalMainResyncFastForwards|TestLocalMainResyncRefusesDiverged|TestLocalMainResyncAheadIsNoop|TestLocalMainResyncPreservesIgnoredFile|TestIntegrationMergeWorktreeRefusesPrimaryHoldingBranch|TestIntegrationMergeWorktreeAcceptsPrimaryWhenEnabled|TestIntegrationMergeWorktreeRefusesUnheldBranch|TestFactoryCompletePrimaryTreeGateEnabled|TestFactoryCompletePrimaryTreeGateDisabled|TestFactoryCompleteNoIntegrationTreeRefused|TestMWQ19_Scenario5_AdoptionRefusedAfterNewCommit|TestSD_AC013_ClaudeCompleteViaIntegrationWorktree|TestSD_AC024_CodexMergeRefusedComplete|TestSD_AC024_CodexMergeRefusedStage|TestSD_AC024_CodexMergeRefusedMCP|TestSD_AC025_IntegrationWindowSerializes|TestAutoMergeOffBaseline|TestAutoMergeHappyPath|TestAutoMergeNonCleanExit|TestAutoMergeUnconfiguredTarget|TestAutoMergeBusyWindow|TestAutoMergeZeroPush|TestAutoMergeConflict|TestAutoMergeSourceDirty|TestAutoMergeTargetGuards|TestAutoMergeNoOpSilent|TestAutoMergeNoticePrefixDistinct|TestAutoMergeToggleIndependence|TestAutoMergeFailurePaths|TestAutoMergeRealImplErrorPaths)$' | grep -c '^Test' && go test ./internal/cli/ -run '^(TestIntegrationSurfaceSelectsPrimaryWhenEnabled|TestIntegrationSurfaceRefusesPrimaryWhenDisabled|TestIntegrationSurfaceRefusesPrimaryOffBranch|TestIntegrationSurfaceRefusesNoHolder|TestIntegrationSurfaceSeparateWorktreeUnchanged|TestLocalMainMergeMergesIntoPrimary|TestLocalMainMergeRefusesDirtyPrimary|TestLocalMainMergeRefusesMovedHead|TestLocalMainMergeStatusSetUnchanged|TestLocalMainMergePreservesIgnoredFile|TestLocalMainResyncFastForwards|TestLocalMainResyncRefusesDiverged|TestLocalMainResyncAheadIsNoop|TestLocalMainResyncPreservesIgnoredFile|TestIntegrationMergeWorktreeRefusesPrimaryHoldingBranch|TestIntegrationMergeWorktreeAcceptsPrimaryWhenEnabled|TestIntegrationMergeWorktreeRefusesUnheldBranch|TestFactoryCompletePrimaryTreeGateEnabled|TestFactoryCompletePrimaryTreeGateDisabled|TestFactoryCompleteNoIntegrationTreeRefused|TestMWQ19_Scenario5_AdoptionRefusedAfterNewCommit|TestSD_AC013_ClaudeCompleteViaIntegrationWorktree|TestSD_AC024_CodexMergeRefusedComplete|TestSD_AC024_CodexMergeRefusedStage|TestSD_AC024_CodexMergeRefusedMCP|TestSD_AC025_IntegrationWindowSerializes|TestAutoMergeOffBaseline|TestAutoMergeHappyPath|TestAutoMergeNonCleanExit|TestAutoMergeUnconfiguredTarget|TestAutoMergeBusyWindow|TestAutoMergeZeroPush|TestAutoMergeConflict|TestAutoMergeSourceDirty|TestAutoMergeTargetGuards|TestAutoMergeNoOpSilent|TestAutoMergeNoticePrefixDistinct|TestAutoMergeToggleIndependence|TestAutoMergeFailurePaths|TestAutoMergeRealImplErrorPaths)$' -v -count=1
```

V5 (required count 9):

```bash
go test ./internal/factory/ -list '^(TestMergeStepPreMergeCausesReleaseWithDistinctCodes|TestMergeStepHappyPathCreatesNoFFMergeAndReleases|TestMergeStepMergeFailureCleanAbortsCause6|TestMergeStepDirtyDisjointPathsProceeds|TestMergeStepDirtyPathOverlapRefuses|TestMergeStepDirtyDirectoryPrefixOverlapRefuses|TestMergeStepDirtyIgnoredTargetRefuses|TestMergeStepStatusSetChangedAfterMergeHolds|TestMergeStepStatusSetChangedAfterAbortHolds)$' | grep -c '^Test' && go test ./internal/factory/ -run '^(TestMergeStepPreMergeCausesReleaseWithDistinctCodes|TestMergeStepHappyPathCreatesNoFFMergeAndReleases|TestMergeStepMergeFailureCleanAbortsCause6|TestMergeStepDirtyDisjointPathsProceeds|TestMergeStepDirtyPathOverlapRefuses|TestMergeStepDirtyDirectoryPrefixOverlapRefuses|TestMergeStepDirtyIgnoredTargetRefuses|TestMergeStepStatusSetChangedAfterMergeHolds|TestMergeStepStatusSetChangedAfterAbortHolds)$' -v -count=1
```

Full package runs (`go test ./internal/cli/`, `./internal/factory/`, `./internal/hook/` without `-run`) are not proof here. A run phase that needs one takes the matching slot lease (for example `moai slot acquire --resource internal-cli-suite`) as lane protocol §8 requires.

## G. Design and gap specifications

### G.1 Closure design record (REQ-LMF-007)

The rule is B7. The RED evidence is E-25 (the refusal at `card_transition.go:670` is live) and E-26 (the reserved-edge refusal at line 333). The tool-owned anchors are E-52 to E-57, and the citations of B7 are E-58 to E-60. The boundary sentence to carry into AGENTS.local.md §4.0 reads: "Closing a merged-local card on the local-main surface does not push and does not wait for a release batch. Disposing of the card worktree waits until the batch that carries its merge has landed on origin." Implementation: sibling card t1621. No code in this card.

### G.2 Gap (b): the pre-push hook and a `main:release/*` push

Current behavior, read from source and not yet observed. The hook body `prePushHookContent` (`internal/cli/hook_install.go:42-121`) has no branch-name rule. For every push it runs `make -s ci-local` and exits non-zero on failure. Then, for each pushed ref, it collects commit subjects: for a new remote ref, `git log --format=%s <local_oid> --not --remotes`; otherwise `remote_oid..local_oid`. If `moai` is on PATH and the subject set is non-empty, it pipes the subjects to `moai hook pre-push` (`internal/cli/hook_pre_push.go:24`). That verb runs only when the `enforce_on_push` gate is on, and its severity follows `git_strategy.<mode>.hooks.pre_push`, which is `warn` in the manual profile.

Consequence, derived from source: `[origin]` `git push origin main:refs/heads/release/main-batch-YYYYMMDD` creates a new remote ref. Its subject set is every commit reachable from local main and absent from all remote-tracking refs, which is the whole unpushed batch. The full `ci-local` runs first.

This checkout sets `core.hooksPath` to `/dev/null` (cell E-48), so the installed hook does not run here. The run-phase procedure therefore runs the hook body directly, outside git, and records that it did so. A push from a checkout where the hook is installed runs it; this card does not observe such a checkout (residual risk R-3).

Run-phase procedure, with no remote contact: in a scratch repository with a stub `Makefile` whose `ci-local` target succeeds, run the hook body with the crafted stdin line `refs/heads/main <local sha> refs/heads/release/main-batch-20261010 0000000000000000000000000000000000000000`. Record whether `ci-local` ran, the subject set printed, and the exit code. Repeat with the gate on in a scratch configuration to observe the severity.

Open: OQ-10 (leader-owned, residual risk R-6). Whether release pushes run the full `ci-local` under a slot lease, and whether a subject violation in a batch should warn or block.

### G.3 Gap (c): the leader guard exemption (MI-2, open)

No guard in `internal/hook` names a leader exemption. Two exemptions exist.
- `internal/hook/contract_sign_guard.go`, the role gate. `moai contract decide`, and `sign --signer llm` or `llm+jev`, are allowed when the session is not lane-refused: no `MOAI_FACTORY_ROLE` marker equal to the role constant, no lane-label variable, and `MOAI_KANBAN_BACKEND` not naming Codex. The file describes this allow direction as the leader's own decide path. The deny sentinel is `CONTRACT_SIGN_AGENT_VIOLATION:`.
- `internal/hook/branch_guard.go`, `isExemptAgent` (line 978). Exempt when `MOAI_BRANCH_GUARD_EXEMPT=1` (main thread only) or when the invoking agent identity is `manager-git` (reachable from spawned agents, as `main-checkout-branch-guard.md` records).

Run-phase procedure for the first candidate: feed the guard a crafted PreToolUse payload with the Bash command `moai contract decide t1616 --spec SPEC-LOCAL-MAIN-FLOW-001 --judgement -`, under each environment state (no marker; marker set to worker; a lane label set), and record allow or deny with the sentinel. For the second candidate, the procedure is the one already recorded in `main-checkout-branch-guard.md`.

Open: MI-2. The operator must name which exemption the "leader guard" refers to. This plan does not guess.

## H. Missing inputs, lane decisions, open items, and resolved items

### Lane decisions and confirmations (2026-10-10)

- **MI-5 — resolved: seven supersede markers.** Line 219 joins the set. Its text says the card PR to base main is the only public path, which contradicts the release-PR-only model (REQ-LMF-014). The set is 179, 195, 196, 199, 216, 218, and 219 (cells E-41, E-42, E-43, E-44, E-38, E-39, and E-40; AC-LMF-011 expects seven). Lines 212 and 214 stay out of the set, for these reasons. Line 214 is the self-dispatch exception, and it already lands on main through the tool, so it is consistent with the new model (cell E-37). Line 212 is the lane duty for lanes that stop at merge-ready (Codex, REQ-SD-025), which the new model does not change (cell E-36).
- **OQ-3 — closed; moved to residual risk R-2.** `workflow.worktree.auto_merge` is `true` today (`workflow.yaml:164`). REQ-LMF-008 sets it to `false` in M3, and until M3 lands the session-exit path stays enabled. This card makes no code change to `internal/cli/session_worktree_automerge.go`, and it adds no acceptance criterion.
- **OQ-6 — closed.** `workflow` stays `git-flow`. No change.
- **OQ-8 — redirected.** The local-main re-sync is not an operator-terminal requirement. It is a tool-owned `--ff-only` fast-forward from `origin/main`, run by the verb inside the integration window under the B3 guard treatment (B3a). The diverged case, which a fast-forward cannot cover, is inside the Q2 option scope.
- **MI-4 — confirmed by the lead.** Template lines 92–95 go in §2. Template lines 160–163 go in §3, including "remote default branch". The local-main override in AGENTS.local.md §4.0 names local main as the card branch point and overrides that sentence explicitly.
- **progress.md — accepted as the plan-phase skeleton.** The §E skeleton in the SPEC directory is the phase record. The lane's card record at `.moai/reports/t1616/progress.md` stays outside the SPEC directory.

### Missing inputs (not operator decisions)

| ID | Input needed | Why it cannot be settled here | Blocks |
|----|--------------|-------------------------------|--------|
| MI-2 | Which exemption "leader guard" names: the contract-sign role gate, or the branch-guard identity exemption | no guard in the tree is named for a leader exemption; the leader has not answered | G.3 measurement |

### Open items carried as residual risk (leader-owned; no acceptance criterion)

- **OQ-7 — open** (residual risk R-4). The product release harness: `hns-release-specialist` cuts release branches from develop and back-merges main into develop (Phase 9). The cut point under the local-main flow, and whether Phase 9 still applies, are open.
- **OQ-9 — open; cross-SPEC** (residual risk R-5). SPEC-WORKTREE-SWEEP-001 REQ-WS-004 defaults its base to `origin/develop`. This SPEC does not edit that SPEC.
- **OQ-10 — open** (residual risk R-6; G.2). Whether release pushes run the full `ci-local` under a slot lease, and whether a subject violation in a batch warns or blocks.

### Resolved in this revision

- **Landing surface.** Local main of the primary checkout, decided by the operator (B1).
- **Absorb gate.** `[local]` `git merge-base --is-ancestor 09a42899c HEAD` (ABS-0, spec §E).
- **First-merge checks.** The checks in AC-LMF-016 and §J test local main only: local main contains the card tip after the merge. No first-merge check names origin/main; BASELINE_SHA applies only to the re-sync (plan §B3a).
- **Closure.** Tool-owned on both merge paths (B7, REQ-LMF-007).
- **Re-sync.** Tool-owned fast-forward with three cases, comparing against BASELINE_SHA (B3a, REQ-LMF-014).
- **Q1 and Q2.** Recorded open in decision-index.md with empty verdicts.
- **Branch point for card worktrees.** Local main for this repository, stated as an override in AGENTS.local.md §4.0 (B9).

## I. Observed drift outside this SPEC's scope

- `workflow.yaml` `worktree:` block (lines 151–163): the comment at 151 says `auto_create` is false, but the value at 163 is `auto_create: true`. The comment at 155 says "auto_cleanup: true → false", but the value at 162 is `auto_cleanup: true`. Not corrected here, except the auto_merge sentence that REQ-LMF-008 requires. The template copy's honesty guard reads the template, not this file.
- `AGENTS.local.md:199` says the config comment still names `origin/develop..develop`. The config comment at `git-strategy.yaml:23-24` names `origin/<integration-branch>`, so the local note is itself stale. Corrected in M4.
- `git-strategy.yaml:15` sets `main_branch: ""` in the manual profile, while the personal and team profiles set `main_branch: main`. Not used by this SPEC.
- `.git/config` sets `core.hooksPath` to `/dev/null` (cell E-48) in this checkout. Not corrected; residual risk R-3.

## J. Who performs the card's last write to main

The last write is the merge of this card's tip into local main. The design names the performer for each Q2 option (decision-index.md). No option uses a branch-guard exemption, the `MOAI_BRANCH_GUARD_EXEMPT` sentinel, or the `manager-git` identity. No option in the table is a fixed step of any card. Each claim in the verification column is labeled `[local]` or `[origin]`.

| Q2 option | Performer | Verification |
|-----------|-----------|--------------|
| (a) operator's one-off merge | the operator, in the primary checkout's own terminal | `[local]` the card tip is an ancestor of local main at evaluation time (`git merge-base --is-ancestor <card tip> main` exits 0); `[local]` a remeasure record exists for the merge tree |
| (b) designed landing verb | the lane session, through `moai factory complete <card>` (B3) | `[local]` the merge SHA is recorded; `[local]` the first parent equals the pre-merge HEAD; `[local]` the status set is equal before and after (REQ-LMF-006); `[local]` the remeasure record is valid; the closure follows B7 |
| (c) separate integration branch | no write to main; the card lands on the integration branch | `[local]` `git rev-parse main` is unchanged; any later promotion is a separate operator act |

On the lane's final merge: the deny list does not match the verb's merge (B3), because the verb's merge is not an agent commit verb. A merge that an agent types into Bash (`git merge`) is denied by the branch-state pattern at `branch_guard.go:155` whatever the list says. That is why the verb is the designed path and a typed Bash merge is not.

Status: OPEN (Q2). Q1 changes only the agent Bash commit verbs (B3). This plan does not choose.

## K. Residual risks

Each risk states what the observations do not rule out, so a reader can judge it without re-deriving the measurement. The owner is the party who can reduce the risk. R-2, R-4, R-5, R-6, and R-9 are the open items carried as residual risk, owned by the leader; none of them has an acceptance criterion.

- **R-1 — The re-measure command is declared by the lane.** `RunRemeasure` runs the lane's command through `sh -c` (`internal/factory/integration_remeasure.go:563`), and the record is trusted by its tree key and build identity, not by a signature (the trust model stated at `integration_remeasure.go:17-20`). `ValidateRemeasureRecord` refuses a non-zero exit and an empty sweep, but a lane can still declare a command that does not test what the card changed. Owner: the lane protocol and the card review. Reduction: the card-review evidence names the command, and the leader reads it.
- **R-2 — Session-exit auto-merge has no primary-checkout refusal (former OQ-3, closed by the lane decision).** The switch `workflow.worktree.auto_merge` is `true` today (`workflow.yaml:164`). REQ-LMF-008 sets it to `false` in M3, and until then the path is enabled. This card makes no code change to `internal/cli/session_worktree_automerge.go`. The path (`sessionExitAutoMerge`, from line 156) runs `git merge` in the target tree without a primary-checkout refusal. That code was read at grep depth in revision 0.2 and is not re-read here. The residual is that re-enabling the switch without a later SPEC would let the path merge in the primary checkout. Owner: the leader.
- **R-3 — `core.hooksPath` is `/dev/null` in this checkout.** The installed pre-push hook does not run here (cell E-48), so gap (b) can be observed only by running the hook body directly (G.2). A push from another checkout still runs the hook, and this card does not observe that checkout. Owner: the leader, who can confirm where the setting came from.
- **R-4 — The release harness still describes develop cuts (OQ-7, open).** `hns-release-specialist` cuts release branches from develop and back-merges main into develop. A release run that follows it would cut from develop until M4 item 7 is complete, and that item is bounded by OQ-7. Owner: the leader.
- **R-5 — The sweep base is stale (OQ-9, open, cross-SPEC).** SPEC-WORKTREE-SWEEP-001 REQ-WS-004 defaults its remote-landing base to `origin/develop`. Under this flow a sweep can judge a card landed against the wrong base. This SPEC does not edit that SPEC. Owner: the leader.
- **R-6 — The release push gate is unmeasured (OQ-10, open).** `[origin]` A push of `main:release/*` creates a new remote ref, runs the full `ci-local`, and runs the subject check (G.2). Whether a batch warns or blocks on a subject violation is not measured. Owner: the leader.
- **R-7 — The closure is not implemented here.** Until t1621 lands, the no-remote refusal at `card_transition.go:670` (cells E-25 and E-60) and the reserved-edge refusal at line 333 (cell E-26) keep a merged-local card from reaching done while origin is configured. A merge made through the landing verb alone also leaves the card unclosed until `factory complete` runs the adoption path. Owner: t1621, and the lane that runs the complete step.
- **R-8 — Q1 applies to this repository only.** The decision changes this repository's commit denial, and the template default stays `[]` (`defaults.go:1491`). A decision recorded here does not change what other projects receive. Owner: the operator (Q1).
- **R-9 — The leader guard exemption is unidentified (MI-2, open).** G.3 measures two candidates inside `internal/hook`. An exemption outside that package would not be found by this plan. Owner: the leader.
- **R-10 — The absorb changed this worktree's hooks and rule files.** The absorb replaced them with the develop-branch versions, so the guard behavior observed in this session can differ from the card base. V10 observes the verb's guard treatment in a scratch primary during the run phase, not in this session. Owner: the lane, in the run phase.
