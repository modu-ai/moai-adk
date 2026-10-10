---
id: SPEC-LOCAL-MAIN-FLOW-001
title: "Local-main integration flow: tool-supported card landing into the primary checkout's main, one release pull request per batch"
version: "0.8.0"
status: in-progress
created: 2026-10-10
updated: 2026-10-10
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/cli, internal/factory, internal/homestate, internal/config, internal/template"
lifecycle: spec-anchored
tags: "local-main, integration, factory, release-batch, config, docs"
tier: M
---

# SPEC-LOCAL-MAIN-FLOW-001

## §A — History

- **2026-10-10 (v0.1.0)** — plan-phase authored from factory card t1616 with three artifacts. Initial operator decision recorded: develop is retired as the integration branch; card work merges into the local main of the primary checkout; origin/main advances only through one release pull request per batch (a release branch, a merge commit, no squash).
- **2026-10-10 (v0.2.0)** — leader decisions applied. (1) The landing surface is the local main of the primary checkout (decided). (2) The develop-into-main commit `09a42899c` is absorbed into WT-10-10-class as commit `e32f69c46` (parents `2aab5f797` and `09a42899c`), so M3 and M4 are ordered after the absorb instead of gated on the precondition. (3) The closure path from merged-local to a closed record is designed here, and its implementation boundary is stated against sibling card t1621. (4) AGENTS.md sections 2 and 3 take the template's configuration-driven wording. (5) The integration-surface selection, the tool-owned merge on the primary checkout, its treatment by the branch guard, and the dirty-primary refusal are specified. (6) Two operator-held decisions are recorded in `decision-index.md`; they gate the criteria that depend on them. Tier M: 15 requirements and 16 acceptance criteria (ceilings 16 and 16).
- **2026-10-10 (v0.3.0)** — revision request #4 applied. (1) The evidence ledger (acceptance.md §E) is re-measured at HEAD `e32f69c46`. The new cells cover the tree identity of `e32f69c46` and `09a42899c`, the `core.hooksPath` setting, the seven supersede lines, and the closure anchors. (2) The absorb gate is `git merge-base --is-ancestor 09a42899c HEAD`, replacing the earlier main-ancestry wording. (3) Leader guidance items (a), (b), and (c) are explicit in REQ-LMF-002, REQ-LMF-003, and REQ-LMF-004, with the guard treatment stated per clause. (4) The closure (REQ-LMF-007) is tool-owned on both merge paths. (5) MI-5 is resolved: seven supersede markers (lines 179, 195, 196, 199, 216, 218, and 219); lines 212 and 214 are not superseded. (6) OQ-8 is redesigned as a tool-owned fast-forward re-sync (REQ-LMF-014, plan §B3a); its diverged case stays open in decision-index Q2. OQ-3 and OQ-6 are settled as recorded in plan §H. (7) Residual risks are recorded in plan §K. (8) decision-index.md is rebuilt: Q1 and Q2 stay open with empty verdicts. Tier M: 15 requirements and 16 acceptance criteria (ceilings 16 and 16).
- **2026-10-10 (v0.4.0)** — lane decisions on the open items applied. (1) MI-5 is closed: the supersede set is seven markers (AGENTS.local.md lines 179, 195, 196, 199, 216, 218, and 219); lines 212 and 214 stay unmarked, with the rationale in plan.md §H. (2) OQ-3 is closed as residual risk R-2: this card makes no code change to the session-exit auto-merge path, and REQ-LMF-008 sets `workflow.worktree.auto_merge` to false (v0.4.0 read it as already false; it is `true` at `workflow.yaml:164`). (3) OQ-6 is closed: `workflow` stays `git-flow`. (4) OQ-8 is redirected: the local-main re-sync is a tool-owned `--ff-only` fast-forward run by the verb inside the integration window under the B3 guard treatment, not an operator-terminal step; its diverged case stays in the Q2 option scope. (5) OQ-7, OQ-9, and OQ-10 remain open items under residual risk, owned by the leader, with no acceptance criterion. Tier M: 15 requirements and 16 acceptance criteria (unchanged).
- **2026-10-10 (v0.5.0)** — lead response to v0.4 applied. (1) AC-LMF-015 states where E-22 and E-27 now probe (the working tree on HEAD `e32f69c46`, untracked, no commit yet), that the RED cell is a regression guard and not a release gate, and the severity split (14 Must, 2 Should). (2) The Q1 default in decision-index.md is "not ranked", with the REQ-LMF-009 fallback; its alternate and note are removed. (3) Claims that depend on a ref are labeled `[local]` or `[origin]`; BASELINE_SHA is used only for `[origin]` claims, and the first-merge checks stay local ancestry checks. (4) MI-4 is confirmed by the lead. (5) progress.md in the SPEC directory is accepted as the phase skeleton. (6) The first-merge checks in AC-LMF-016 and plan §J test local main only; no first-merge check names origin/main. (7) The lint and audit results are recorded in acceptance.md §E with their judging builds (cells E-61 to E-64), and the ancestry check is E-65. (8) AC-LMF-013 is Should for the reason stated in its scenario. No requirement or criterion is added: 15 and 16.
- **2026-10-10 (v0.6.0)** — revision request #5 applied to the plan-audit FAIL at `0cca5364f` (score 0.69; twelve blocking findings and the non-blocking list). (1) REQ-LMF-002 states the gate-false case with the primary checkout holding the configured branch as an explicit refusal, in one state-driven form. REQ-LMF-006 is Event-driven. REQ-LMF-014 is one Event-driven form, and the batch release procedure moves to REQ-LMF-011. (2) REQ-LMF-003 conditions the commit, revert, and cherry-pick denial on `deny_commits_on` containing `main`, and it re-reads the symbolic HEAD after the merge. (3) `workflow.worktree.auto_merge` is `true` at `workflow.yaml:164`, and the v0.4.0 statement is corrected. (4) Milestones M3 and M4 replace the earlier R3 and R4 wording. The requirement and criterion counts are unchanged: 15 and 16.
- **2026-10-10 (v0.7.0)** — revision request #6 applied to the plan-audit iteration 2 FAIL at `fa41b427a` (score 0.75; eight blocking items I2-1 to I2-8). (1) The three clauses of REQ-LMF-002 are exclusive, because git checks a branch out in one worktree at a time: a separate holder; the primary holding the configured branch with the gate true; otherwise a refusal. (2) REQ-LMF-004 also refuses a merge, with the exit class `MergeExitWorktreeDirty`, when an ignored file sits at a primary target path; plan §B3 step 4a implements it. (3) REQ-LMF-012 names the four documents and states the OQ-7 limit for `hns-release-specialist.md`. (4) On the no-fast-forward path the flag does not protect ignored files (cell E-78); the protection is the pre-merge overlap check. (5) Decision options (ii) and (iii) name the follow-up SPEC that they require. The requirement and criterion counts are unchanged: 15 and 16.
- **2026-10-10 (v0.8.0)** — revision 0.8 applied to the plan-audit iteration 3 final FAIL at `77100366d` (finding I3-1), under the operator's decision d-20261010T000832Z-efde. (1) REQ-LMF-003 states that the verb does not depend on `isExemptAgent`, identity, or sentinel for its allow decision. (2) REQ-LMF-004 refuses a merge with exit class `MergeExitCollision` when the collision check `FindAddedPathCollisions` returns a non-empty set for the pinned card tip; the dirty-primary refusal keeps `MergeExitWorktreeDirty`. (3) Plan §B3 step 4a runs that collision check, and the B5 intersection is kept as the separate-worktree input only. (4) AC-LMF-004 has three shape scenarios (an ignored ancestor, a file over an ignored directory, and a case alias), and E-79 is their release-blocking RED cell. (5) REQ-LMF-005 states that its ignored-file clause is the same collision check. The requirement and criterion counts are unchanged: 15 and 16.

## §B — Problem

The operator's flow is local main, then a worktree, then a local-main merge, then a push, then a release pull request. The tree does not yet support it.

1. **The integration verbs refuse the primary checkout as a merge target.** `internal/cli/integration_merge.go:137-139` and `internal/cli/factory_card.go:1893-1895` refuse when the only tree holding the integration branch is the primary checkout. The landing cannot reach the local main through the tooling.
2. **The merge step requires a fully clean integration worktree.** `internal/factory/integration_merge_step.go:243-249` refuses any dirty or untracked byte, and the same rule repeats after an abort (line 472) and after the merge (line 533). A dirty file the merge never touches blocks the merge. For a dedicated integration worktree the rule should narrow to overlap. The primary checkout is shared, so it needs a stricter rule.
3. **A card cannot close while a remote is configured.** The no-remote done edge refuses when any remote exists (`internal/homestate/card_transition.go:668-670`). The reserved pushed-to-ci-green edge is refused (`card_transition.go:207-209`, consulted at line 333). The merged-local-to-pushed edge requires a remote (`card_transition.go:686-688`). This repository has origin, so no card can close locally.
4. **Configuration and documents still describe develop as the integration branch.** `.moai/config/sections/git-strategy.yaml:16` sets `develop_branch: develop`. `.moai/config/sections/workflow.yaml:164` sets `auto_merge: true`, a session-exit merge that targets that branch. `AGENTS.md` hardcodes main-flow sentences at lines 91, 94, and 166. The local rules describe the develop chain. No document states the batch release procedure.

Two behaviors stay unmeasured and are specified only: how the pre-push hook treats a `main:release/*` push, and which guard the operator's "leader guard exemption" refers to.

## §C — Goal

- Make local-main landing tool-supported behind an opt-in gate whose distributed default keeps today's behavior.
- Specify how the integration surface is selected by configuration: the local main directly, or a separate integration branch with its own integration worktree.
- Specify the tool-owned merge on the primary checkout, including its treatment by the branch guard. The treatment is an explicit design, not a bypass.
- Specify the refusal and guidance when the primary checkout has uncommitted changes.
- Narrow the clean-tree rule for dedicated integration worktrees to overlap.
- Specify a push-free path from merged-local to a closed record, with the boundary to sibling card t1621 stated.
- Make AGENTS.md universal, and move the repository-specific flow to AGENTS.local.md §4.0.
- Document the batch release procedure, and specify (not implement) the two remaining unmeasured behaviors.

## §D — Requirements (GEARS)

15 requirements (Tier M ceiling 16). `<subject>` is the named component unless stated otherwise. Ordering and decision gates are stated per requirement and summarized in §E and §I.

The leader guidance of 2026-10-10 has three items. Item (a), surface selection by configuration, is REQ-LMF-002. Item (b), the tool's merge on the primary checkout and its guard treatment, is REQ-LMF-003. Item (c), the refusal and guidance for uncommitted changes in the primary checkout, is REQ-LMF-004.

### D.1 — Integration surface and landing

- **REQ-LMF-001** (Ubiquitous) — The workflow configuration shall define the boolean key `workflow.local_main_integration.enabled`, whose shipped template default is `false`. The key shall be authored in the template configuration first, the embedded template regenerated with `make build`, and the repository's copy synchronized only after that regeneration (Template-First).
- **REQ-LMF-002** (State-driven) — Leader guidance (a). While a separate integration worktree holds the configured integration branch `git_strategy.manual.develop_branch`, the integration surface shall be that worktree. While no separate integration worktree holds that branch, the primary checkout holds it, and `workflow.local_main_integration.enabled` is `true`, the integration surface shall be the primary checkout's local main. While no separate integration worktree holds that branch, and the primary checkout does not hold it with `workflow.local_main_integration.enabled` set to `true`, the integration merge verb and factory complete shall refuse with guidance, shall switch no branch, and shall perform no merge. The three conditions are exclusive, because git checks a branch out in one worktree at a time. Choosing between the local main directly and a separate integration branch is therefore a configuration choice together with which tree holds that branch. Both the integration merge verb and factory complete apply this single rule.
- **REQ-LMF-003** (State-driven) — Leader guidance (b). While the integration surface is the primary checkout's local main, the integration merge verb shall perform the merge itself, inside the integration window, in the no-fast-forward form `git -c merge.autoStash=false merge --no-ff --no-overwrite-ignore` on the pinned card tip. It shall first verify that the primary checkout's HEAD names the configured branch, and after the merge that the first parent of the new HEAD equals the pre-merge HEAD and that the symbolic HEAD still names the configured branch. The guard treatment is a designed exception, not a bypass. The merge is a child process of the moai binary, not command text the agent issues. The verb shall not read `workflow.branch_guard.deny_commits_on`, shall not use the `MOAI_BRANCH_GUARD_EXEMPT` sentinel or the `manager-git` identity, and shall run no checkout, switch, reset, stash, rebase, or merge other than the no-fast-forward form above and the fast-forward-only re-sync of REQ-LMF-014. The branch guard's rules, exemptions, and patterns shall remain unchanged. Agent-issued merge commands in the primary checkout shall remain denied by the branch-state pattern, and agent-issued commit, revert, and cherry-pick commands in the primary checkout shall remain denied while `workflow.branch_guard.deny_commits_on` contains `main`. The verb shall not depend on `isExemptAgent` (`internal/hook/branch_guard.go`), identity or sentinel, for its allow decision. Its command text is `moai …`, not a branch-state command.
- **REQ-LMF-004** (State-driven) — Leader guidance (c). While the integration surface is the primary checkout, the tool shall not merge into a primary checkout that has any uncommitted or untracked change. It shall refuse with the dirty-worktree exit class `MergeExitWorktreeDirty`. It shall refuse a merge with exit class `MergeExitCollision` when the collision check `FindAddedPathCollisions` (`internal/factory/integration_merge_collision.go:225`) returns a non-empty set for the pinned card tip. That set holds a path the merge would write, or an ancestor of one, that is ignored or untracked; or a directory at such a path that holds ignored or untracked bytes beneath it; and, on a case-insensitive volume, a case alias of either. The guidance for an uncommitted or untracked change shall name the remedies (move the changes into a card worktree and commit them there, or have the session that owns them commit or discard them) and shall forbid stashing, because the stash is repository-wide. The guidance for an ignored file shall name the paths and leave their move or rename to the operator's own terminal, because the tool removes and stashes nothing.
- **REQ-LMF-005** (Event-driven) — When the integration surface is a separate integration worktree and the merge step prepares a merge, the step shall refuse if and only if a dirty or untracked path of that worktree overlaps a path the merge would write, where overlap is equality or a directory-prefix relation, or if an ignored file sits at such a path. Its ignored-file clause is the collision check of REQ-LMF-004, the same call the separate-worktree step makes at `integration_merge_step.go:313`; the overlap rule above is its dirty-path part.
- **REQ-LMF-006** (Event-driven) — When the merge commit has been written or an abort has completed, on either surface, the merge step shall compare the integration worktree's status set with the set captured before the merge. The status set is read NUL-delimited, with untracked files listed individually. Any difference shall be classified with the existing post-merge hold class or the existing dirty-after-abort class.

### D.2 — Closure

- **REQ-LMF-007** (State-driven) — While a merged-local card's merge is on the local-main integration surface, the closure to merged-local shall be performed by the tool on both merge paths: the adoption path (`internal/cli/factory_card.go:1924-1960`, a merge commit the lane made through the merge verb before complete ran) and the non-adoption merge step (`internal/cli/factory_card.go:1962-2010`, a merge the tool makes now). On each path the tool shall admit the merged-local transition only when (i) the merge commit has two parents, its tree equals its second parent's tree, and it is reachable from the integration branch; (ii) the store record keyed to that tree passes `ReadRemeasureRecord` and `ValidateRemeasureRecord`; and (iii) the leader approval receipt verifies. The non-adoption path also checks the candidate-tree record before its merge; under (i) that record is keyed to the same tree as the merge. On this surface the store record is the only re-measure evidence, and a lane-supplied evidence file shall not substitute for it. The transition shall perform no push and shall not require one. A card worktree shall not be disposed of until its merge has reached the remote through a release batch. The admission itself is implemented by sibling card t1621. This SPEC records the rule and the boundary.

### D.3 — This repository's configuration and documents

- **REQ-LMF-008** (Ubiquitous) — In this repository, `git_strategy.manual.develop_branch` shall be `main`, and `workflow.worktree.auto_merge` shall be `false`, with a comment that records the reason. Every commit that edits these two files, or the documents named in REQ-LMF-010 to REQ-LMF-012, shall satisfy `git merge-base --is-ancestor 09a42899c HEAD` with exit code 0. That is the absorb gate.
- **REQ-LMF-009** (Ubiquitous) — In this repository, `workflow.branch_guard.deny_commits_on` shall keep its current value `[main]` until operator decision Q1 is recorded, and this SPEC shall not change it. The repository's `workflow.local_main_integration.enabled` shall be `true`.
- **REQ-LMF-010** (Ubiquitous) — `AGENTS.md` sections 2 and 3 shall carry the template's configuration-driven wording: template `AGENTS.md.tmpl` lines 92–95 for section 2, and lines 160–163 for section 3. They shall carry no repository-specific text. The distributed template shall not be changed by this SPEC.
- **REQ-LMF-011** (Ubiquitous) — `AGENTS.local.md` shall contain section 4.0, "Default development flow (local-main integration)". It states the landing flow, the card branch point, the designed exception for the primary surface, the re-sync, the closure boundary, and the batch release procedure: a push of local main to `release/main-batch-YYYYMMDD` on origin, a pull request into `main`, and a merge commit without squash. Origin main shall advance only through such a release pull request. Each superseded clause shall carry a `[SUPERSEDED by §4.0]` marker in place, without deletion. The markers are on the seven lines named in plan §B10: 179, 195, 196, 199, 216, 218, and 219. Lines 212 and 214 are not superseded.
- **REQ-LMF-012** (Ubiquitous) — The four documents named in plan §M4 items 4 to 7 shall not state develop as the live integration branch: `.claude/rules/local/gitflow-lane-protocol.md`, `.claude/rules/local/repo-local-pr-policy.md`, `.moai/docs/gitflow-integration-chain.md`, and `.claude/agents/harness/hns-release-specialist.md`. Each such statement is corrected or carries a superseded marker, and every live pointer to the retired `CLAUDE.local.md` §4.1 names `AGENTS.local.md` §4.1. For `.claude/agents/harness/hns-release-specialist.md` the limit is OQ-7: the release-branch source and the Phase 9 back-merge are corrected only as far as OQ-7 allows, and the sentences that depend on OQ-7 carry a superseded marker and stay open (plan §K R-4).
- **REQ-LMF-013** (Unwanted) — A single edit to an always-loaded instruction file shall not grow it by more than 1,000 bytes in a commit whose body lacks the byte-size and cost statement required by the rule-authoring rule. `AGENTS.md` shall remain under 40,000 characters.
- **REQ-LMF-014** (Event-driven) — When a release batch has landed on `origin/main`, the integration surface shall bring local main up to `origin/main` by a tool-owned fast-forward-only merge, run by a subcommand of the integration verb family inside the integration window (plan §B3a), under the REQ-LMF-003 guard treatment. The subcommand shall report that no fast-forward is needed and leave HEAD unchanged for a local main that already contains `origin/main`, and shall refuse with guidance for a local main that has diverged from `origin/main`; the method for that case is decision Q2.

### D.4 — Gaps

- **REQ-LMF-015** (Ubiquitous) — Gap (b), the pre-push handling of `main:release/*` pushes, and gap (c), the measurement of the leader guard exemption, shall each be specified in `plan.md` §G with a measurement procedure and an open question. This SPEC shall change no hook or guard for them.

## §E — Constraints

- **Local landing only.** The run phase performs no push, no pull request, and no remote command. Every criterion is provable locally. The re-sync's `git fetch` belongs to the production tool path and is not run by this SPEC's run phase (plan §B3a).
- **Absorb first.** Commit `e32f69c46` (the merge of `09a42899c` into this branch, parents `2aab5f797` and `09a42899c`) is in the branch at this revision. Every edit commit of M3 and M4 shall satisfy `git merge-base --is-ancestor 09a42899c HEAD` with exit code 0 (AC-LMF-008, AC-LMF-010 to AC-LMF-014). This gate replaces the earlier "landed on main" wording. `[local]` Local main does not contain `09a42899c` (E-23, E-45), so a gate on main would fail for a reason this card does not control.
- **Decision gates.** Operator decisions Q1 and Q2 are recorded in `decision-index.md` with empty verdicts. The criteria that depend on them are marked GATED-ON-DECISION. No criterion requires the first merge of this card into main, and this SPEC ranks no default for Q1 or Q2. The status quo is not a default: while Q1 is undecided, REQ-LMF-009 keeps `[main]`, and while Q2 is undecided, no first-merge method is selected.
- **No card step is fixed to the operator's terminal.** The first-merge method is operator decision Q2 (leader guidance, 2026-10-10).
- **One designed writer for local main.** The landing verb and its fast-forward re-sync (plan §B3, §B3a) are the only tool paths that write the primary checkout's local main. The branch guard is not modified.
- **Template-First.** The only distributed artifact this SPEC changes is the template workflow configuration (REQ-LMF-001). The template `AGENTS.md.tmpl` is read, not changed.
- **Tier M budget.** 15 requirements (ceiling 16) and 16 acceptance criteria (ceiling 16). Any addition requires a tier change, not a relaxed budget.
- **No time estimates** appear in any SPEC artifact.

## §F — Out of Scope

### Out of Scope — remote operations

- No push, pull request, origin release branch, or other remote command is performed by this SPEC. REQ-LMF-014 documents the release procedure. It is not executed here.

### Out of Scope — the value of deny_commits_on (operator decision Q1)

- Changing `workflow.branch_guard.deny_commits_on` is an operator decision. This SPEC keeps the value unchanged and records the options in `decision-index.md` (Q1).

### Out of Scope — the first merge of this card into main and the diverged re-sync case (operator decision Q2)

- The method of the first merge into local main, and the handling of a diverged local main, are operator decision Q2. This SPEC does not require that merge, does not fix its method, and does not run the re-sync. AC-LMF-016 is gated on Q1 and Q2 and is not a required step.

### Out of Scope — the done-transition implementation

- The T18 admission for the local-main surface, the `PushGateTarget` change, and the repair of existing merged-local records belong to sibling card t1621. This SPEC specifies the rule (REQ-LMF-007) and does not modify `internal/homestate`.

### Out of Scope — implementing gaps (b) and (c)

- Gaps (b) and (c) are specified in plan.md §G.2 and §G.3. No hook, verb, or guard changes for them (REQ-LMF-015).

### Out of Scope — flipping the workflow mode

- `workflow` stays `git-flow` (OQ-6, closed by the lane decision of 2026-10-10; plan.md §H). Flipping it to github-flow arms `TestGitHubFlowSweepGuard` and would be a separate decision.

### Out of Scope — pre-existing comment drift in the worktree block

- Recorded in plan.md §I. Only the auto_merge sentence is changed, because REQ-LMF-008 requires it.

### Out of Scope — redesigning the product release harness

- `.claude/agents/harness/hns-release-specialist.md` is changed only where it states develop as the integration branch. Its redesign is plan.md OQ-7.

## §G — Known Relations and Non-Goals

- **Sibling card t1621** covers repairs, including opening the done transition for the local-main surface. The boundary is stated in plan.md §B7 and §G.1.
- **SPEC-WORKTREE-SWEEP-001 (REQ-WS-004)** defaults the sweep's remote-landing base to `origin/develop`. Under this flow that default is stale. It is not changed here (plan.md OQ-9).
- **`lead_push_threshold`** counts `origin/<integration-branch>..<integration-branch>`. With `develop_branch: main`, the counted branch becomes main. The batch-close rule is restated against main in AGENTS.local.md §4.0 (REQ-LMF-014).
- **SPEC-MAIN-COMMIT-BAN-001 (card t1337)** is the origin of the `deny_commits_on: [main]` value (`workflow.yaml:175-177`). Its status is decision Q1.
- **Session-exit auto-merge** (`internal/cli/session_worktree_automerge.go`) has no primary-checkout refusal. Its switch, `workflow.worktree.auto_merge`, is `true` in this repository today (`workflow.yaml:164`), and REQ-LMF-008 sets it to `false`; until that change lands the path stays enabled. The lane decision of 2026-10-10 closes OQ-3 as residual risk R-2 (plan.md §B8, §K): this card makes no code change to the path. Re-enabling it needs a later SPEC.

## §H — Cross-References

- Card t1616 (factory dispatch; operator decision on the local-main flow). Sibling card t1621 (repairs, including the done transition).
- Absorb: commit `e32f69c46` (merge of `09a42899c`).
- Decisions: `decision-index.md` (Q1, Q2).
- Residual risks: plan.md §K.
- Rules and docs: `.claude/rules/local/gitflow-lane-protocol.md`, `.claude/rules/local/repo-local-pr-policy.md`, `.claude/rules/moai/workflow/main-checkout-branch-guard.md`, `.moai/docs/gitflow-integration-chain.md`, `AGENTS.local.md` §4.1, `.claude/rules/moai/development/rule-authoring.md`, `.claude/rules/moai/development/verification-completeness.md`.
- Code anchors: `internal/cli/integration_merge.go:24, 128-141`; `internal/cli/factory_card.go:1754, 1786, 1883-1895, 1924-1960, 1962-2010, 2012-2060`; `internal/factory/integration_merge_step.go:240-249, 466-473, 533-535`; `internal/factory/integration_remeasure.go:68, 550, 610`; `internal/homestate/card_transition.go:157, 207-209, 333, 660-713`; `internal/homestate/card_evidence.go:64-73`; `internal/homestate/card_evidence_readers.go:290-338`; `internal/hook/branch_guard.go:155, 369, 434, 451, 480, 978, 1001, 1111`; `internal/config/defaults.go:1446, 1491`; `internal/config/loader_integration_branch.go:116-123`; `internal/cli/session_worktree_automerge.go:156-304`; `.moai/config/sections/workflow.yaml:151-178`.

## §I — Open decisions (operator-held)

| Decision | Question | Acceptance criteria gated on it | Record |
|----------|----------|----------------------------------|--------|
| Q1 | Value of `workflow.branch_guard.deny_commits_on`: keep `[main]`, change to `[]`, or a per-user choice | AC-LMF-009; AC-LMF-016 | `decision-index.md` Q1 |
| Q2 | Execution method of the first merge into local main (an operator's one-off terminal merge, the designed landing verb, or a separate integration branch), and the method for a diverged local main during the re-sync | AC-LMF-016 | `decision-index.md` Q2 |

Until Q1 and Q2 are decided, this card leaves `deny_commits_on` unchanged and performs no merge into main. The open questions that are not operator decisions are listed in plan.md §H.
