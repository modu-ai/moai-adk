# Design — SPEC-GIT-STRATEGY-AGENTIC-TRUNK-001

Read with research.md (evidence) and spec.md (behavior). This file holds the how: names, predicates, ordering. Tree `6041c38a4`.

## 1. The model being named

```
lane worktrees (branch from local main)
   -> serial landing in ONE dedicated integration worktree (moai integration merge --card, scoped tests, re-measure)
   -> about every lead_push_threshold (20) landings: ONE batch PR (release/main-batch-YYYYMMDD -> main), ONE CI run
   -> tagged release from main
```

Standing branches: `main` only. No `develop`. No per-card PR. No per-version release branch cut by lanes.

## 2. Behavior mapping, code level (decision D1)

`agentic-trunk` takes the TARGET rule of github-flow and the DELIVERY rule of git-flow. Each row is the decision for one code site (research R2).

| Site | Today | Under agentic-trunk | Change |
|---|---|---|---|
| `ClassifyWorkflowDisposition` | git-flow -> `DispositionGitFlow`; other valid -> `DispositionValidNonGitFlow` | `DispositionValidNonGitFlow` | add the value to the valid case; no new disposition (OQ not needed: the class is consumed only for "invalid" and "is it git-flow") |
| `AllowedWorkflows` | 4 values | 5 values, `agentic-trunk` last | append |
| `WorkflowIntegrationTarget` | table of 4 | fixed `main` via a new constant beside `githubFlowFixedTarget` | one case |
| `LoadGitFlowIntegrationConfig` | `GitFlowWorkflow` from the literal `gitFlowWorkflow`; `DevelopBranch` gated by `Manual && GitFlowWorkflow` | add `AgenticTrunkWorkflow` (active profile workflow equals the constant); `DevelopBranch` = `main` when `Manual && AgenticTrunkWorkflow`; `IntegrationTarget` from the table with no mode gate | one field, one branch, constants instead of the second `git-flow` literal |
| `IsGitFlow()` | `Manual && GitFlowWorkflow` | unchanged (false for agentic-trunk) | none — its callers keep their meaning |
| new `UsesLocalIntegration()` | n/a | `Manual && (GitFlowWorkflow \|\| AgenticTrunkWorkflow)` | new method |
| `factoryGitHubFlow` | `Workflow == github-flow` | false | none; a pinning test (the verb is already correct) |
| `sessionExitAutoMerge` gate (`session_worktree_automerge.go:176`) | `!IsGitFlow() \|\| DevelopBranch == ""` | `!UsesLocalIntegration() \|\| DevelopBranch == ""`; notice text names both flows | one predicate, one string |
| `integration acquire` caller-fallback warning (`integration.go:528`) | `IsGitFlow()` | unchanged: with `DevelopBranch == main` the caller fallback never happens in manual mode | none |
| `integration acquire` target (`integration.go:466`) | `DevelopBranch` | `main` in manual mode; in a non-manual mode the t886 "unwired target" warning fires exactly as it does for github-flow | none |
| contract push check (`contract_pushcheck.go:37`) | `LoadGitFlowDevelopBranch(root)`; empty prints `inactive (no git-flow integration branch configured)` and the check does nothing | `main` through the projection, so the check stays active after `develop_branch` is removed from the live yaml (REQ-AGT-018) | none in code; test `TestAgenticTrunkPushCheckIntegration` (AC-AGT-025) |
| closure push guard (`hook/closure_push.go:153-161`, `integrationBranchFor`) | the same loader call, then the primary checkout's value through the git common directory; empty leaves the guard inert | `main` through the projection, in the tree and in a config-orphaned worktree | none in code; test `TestAgenticTrunkIntegrationBranchFor` (AC-AGT-025) |
| card base (`escalation/newapi.go:126-143`, `cardBase`) | a non-empty `LoadGitFlowDevelopBranch` is the sole merge-base candidate; empty falls to origin/HEAD, `main`, `master` | `main` is the sole candidate | none in code; test `TestAgenticTrunkCardBase` with an origin/HEAD that points elsewhere (AC-AGT-025) |
| `loader_slot_lease.go` | a comment cites `LoadGitFlowDevelopBranch` as a modelling reference; the file calls nothing | no behavior | none (comment only, so no test) |
| loader projection (the single source of the three consumers above) | `DevelopBranch` is `develop_branch` under manual git-flow | `main` under manual agentic-trunk whether or not the key is present | test `TestAgenticTrunkDevelopBranchProjection` (AC-AGT-025) |
| `EmptyTargetCause` / `EmptyTargetGuidance` | git-flow causes, fixes name "workflow: github-flow (target main)" | agentic-trunk never reaches an empty cause (target is `main`); the two git-flow fix strings add "or agentic-trunk (target main)" | two strings |
| `invalid workflow` diagnostics (`integration.go:212`, doctor, guidance) | print `AllowedWorkflows()` | print 5 values | free (dynamic) |
| `workflowStandingBranches` (doctor) | 4 phrases | add: "single trunk main; cards land serially in the integration worktree; one batch pull request per lead_push_threshold landings; tagged release" | one case |
| `sweepGuardArmed` (`== WorkflowGitHubFlow`) | github-flow only | stays so (OQ-2) | a dedicated test `TestAgenticTrunkSweepGuardNotArmed` over three synthetic roots; the existing `arming-switch` subtest of `TestGitHubFlowSweepGuard` already passes today, so it cannot be the flip judge (AC-AGT-015) |
| `worktree done` / `sweep` landing checks | ancestry first, then patch-id and PR state | unchanged; no workflow read | none; a grep guard |
| `merge_method`, `lead_push_threshold`, `worktree_base_branch` | read as keys | read as keys | none |
| `develop_branch`, `release_branch_prefix`, `rc_version_format` | git-flow keys | ignored, no diagnostic | none (keys stay in the struct for git-flow) |

Why a new predicate and not widening `IsGitFlow()`: `IsGitFlow()` carries the name and the tests of the git-flow contract (t637 warning text says "git-flow project"); widening it changes what a git-flow reader sees. A sibling method leaves every git-flow assertion untouched (REQ-AGT-001, REQ-AGT-005).

Why the target is not gated by mode: github-flow's target is mode-independent, and `main` is a fact of the model, not a manual-mode key. The legacy develop seam (`DevelopBranch`) stays manual-only because its consumers (session-exit auto-merge, acquire window recording) are opt-in manual-mode machinery (OQ-7).

Why disposition is not a new class: `DispositionGitFlow` is read only as "is it git-flow" inside the loader and diagnostics; `DispositionInvalid` is the one consumers branch on. A fifth class would force every `switch` to learn it for no behavioral difference.

## 3. Display

- Doctor: `workflowStandingBranches` case plus the existing OK message shape `git_strategy workflow = <v>; <phrase>; integration target: main`. The worktree_base_branch mismatch warning is unchanged (here both are `main`).
- Statusline: a small provider returns the workflow label when the active profile reads a valid value other than `github-flow`; the branch segment appends it after the dirty count as ` (agentic-trunk)`. The provider reads `git-strategy.yaml` through the existing reader at the statusline's config root (`config_root.go`); it spawns no git process. Label absent for: github-flow, absent or unreadable file, invalid value (the doctor reports that case). t1830's REQ-PLO-009 byte-identity claim concerns its own change and stays true for github-flow and for an unset value; see the coordination note in plan.md section 5.

## 4. Sequencing hazard (decision D2)

A binary built before M1 reads `agentic-trunk` as invalid: `IntegrationTarget` is empty, `worktree done`/`sweep`, `goal approve`, `factory merge ready`/`complete` and the card diff base refuse an empty target (t1453 M1), the doctor warns, `integration acquire` warns. Observed on the installed build at plan time (`v3.2.0-rc.36-t1822c`, commit `ee981a932`): the doctor prints a `warn` row for the value (acceptance.md ledger E18). The installed `moai` is one shared binary and every running `moai mcp-server` keeps the build it started with (28 servers match the installed binary on tree c6e590507, ledger E19; the count is volatile and no judge depends on it). Therefore the live value changes only when the single precondition list P1 to P5 of REQ-AGT-017 holds, in this order:

1. M1 to M5 land with the live value still `git-flow` (P1: their commits are ancestors of the flip commit).
2. The leader, with the operator's confirmation and while the integration window is free (P5), runs the repo-local rc build (protocol section 9: tree-built binary, `rm -f` then `cp`, `scripts/verify-local-install.sh`) from the integration worktree, so that the installed binary is byte-identical to the build of the flip commit's parent (P2).
3. The leader reconnects the running `moai mcp-server` processes and confirms two things (P4). First `moai doctor --check "MCP Server Version"` reads `ok` against the new commit: that row sees only the servers that stamp a record under one project directory (`<project>/.moai/state/mcp-server`, `internal/cli/mcp_server_runtime.go`). Second, the all-process probe: `pgrep -o -f '^moai mcp-server$'` names the oldest running server of ANY project, `ps -o pid=,lstart=,command= -p <pid>` gives its start time and `stat -f '%Sm' -t '%Y-%m-%dT%H:%M:%S' <installed moai>` the install time; the oldest server must have started after the install. Measured on tree c6e590507 (ledger E22): 32 server processes, 28 with `CLAUDE_PROJECT_DIR` equal to the primary checkout and a record there, 4 of another project (`/Users/goos/MoAI/mo.ai.kr`) with none; no live pid recorded under any worktree. The doctor therefore sees this repository's servers today, and the second probe closes the gap for a server started with another project directory. A server of another project that is left running is named in the progress record; the residual risk is that such a server, asked to act on this repository through an explicit `project_root`, still runs the old reader (AC-AGT-016).
4. The installed binary's doctor, run on a fixture project whose manual workflow is `agentic-trunk`, reports the `Git Strategy Workflow` row `ok` (P3). The installed binary is the subject here, so the judge names `moai` and not a path; the judging build is the installed one and its commit is named next to the tree HEAD (verification-claim-integrity section 2.2).
5. Only then does the one-line live flip land (M6), as a commit of its own touching `.moai/config/sections/git-strategy.yaml` alone, so a revert is a single `git revert` of that commit.

The cutover is leader-executed (leader ruling, OQ-3): replacing the installed binary affects every session on the machine, so no lane performs it. AC-AGT-016 is the one criterion that is judged from the leader's recorded evidence, and the sync audit reads that record.

Rollback: revert the M6 commit; `git-flow` plus `develop_branch: main` is a state every shipped binary understands.

## 5. Template-First, and what must not change for users

Edits go to `internal/template/templates/**`, then `make build` regenerates the embedded catalog and the deployed `.claude/**` copies are brought in line by the repo's documented sync. Rules about `agentic-trunk` are additive sentences ("a second workflow uses this route"); the defaults, `github-flow` as "the distributed default", and every `[HARD]` rule stay. `git-strategy.yaml.tmpl` gets no value change.

`delivery.md` gets a third strategy block, placed after the git-flow block. Design of that block: value domain line lists three values; the table gains a row (`agentic-trunk | main plus WT-* branches | integration-worktree merge, no PR | multi-lane projects landing locally with batch pull requests`); a `Strategy: agentic-trunk` section reuses the git-flow `WT-*` route steps 1 to 5, 7 and 8 (no pull request, merge-window coordination, enter the designated integration worktree, `git merge --no-ff`, resolve, exit and report) and OMITS the push step (the leader's batch pull request delivers the landings; the block's own prose avoids the word "push" and any `gh pr` call, so that the whole block can be scanned for push instructions, AC-AGT-020 Judge 6); a non-`WT-*` branch under agentic-trunk reports the branch state and stops (no per-card pull request exists in this model). The unmatched-value stop (line 29) keeps its behavior with the domain extended to the three values in its condition and in its report, so `/moai sync` does not stop on `agentic-trunk` after the flip (AC-AGT-020 Judges 4 and 5).

The block copies the git-flow steps as written, including the manual `git merge --no-ff`. The repository protocol (section 2) sends lane merges into local `main` through `moai integration merge --card`, and the code path that lands cards (`factory complete`) calls that verb; the existing git-flow block carries the same text, so this is inherited, not introduced. Plan-audit iteration 1 raised it (D9, optional) and the audit demoted it for exactly that reason; it is recorded here and left unchanged, because the distributed skill must read the same for both workflows that share the route and the repository carve-out lives in the repo-local protocol.

`factory-dispatch.md` is the one rule pair that is not byte-identical today: the deployed copy line 222 ends with a sentence about `moai worktree sweep` that the template does not carry. `git log -S` shows commit `7ff326d99` (card t1815) deliberately removed it from the template as a live-only sentence leaked into the mirror, so the reconciliation direction is NOT template-first for that sentence: the template stays without it, the deployed copy keeps it, and the `agentic-trunk` edit to that line goes into both. REQ-AGT-021 records this exception, so requirement and criterion agree: every edited pair is byte-identical except that one tail sentence (322 bytes today). The other seven pairs are byte-identical today and stay so (AC-AGT-021).

Which sentences take `agentic-trunk`, and how (iteration 2 repair of D2). Two families of sentence describe the git-flow variant, and they are edited differently.

1. The shared `WT-*` integration route (the landing-surface paragraph that t1846 landed: "where the integration branch is held by a dedicated worktree (the git-flow `WT-*` integration route) ..." in `worktree-integration.md`, `spec-workflow.md`, `delivery-policy.md` and the git-flow block of `delivery.md`). The route is the integration window, the dedicated integration worktree and `git merge --no-ff`; it carries no push. `agentic-trunk` uses it, so these sentences name both workflows.
2. The per-version release-branch push (`factory-dispatch.md` lines 219 and 222, `factory-dispatch-core.md` line 64, `factory-dispatch-mechanics.md` lines 142 to 149 and 243): "a lane merges its own branch into `release/vX.Y.Z` itself ... pushes `release/vX.Y.Z`". spec.md section 4 says `agentic-trunk` does not follow this (the batch branch is `release/main-batch-YYYYMMDD`, opened by the leader). These sentences stay git-flow only: their "applies only where the project's git strategy names git-flow" wording is not widened. Each file gains one sentence saying that `agentic-trunk` takes no per-version release branch and runs no lane push, because `moai factory complete` lands the card through the integration-window local merge into `main` and the only push is the leader's batch pull request. In `factory-dispatch.md` the sentence is inserted into line 222 after "leaves the batch pull request with the leader." and before "The full window procedure:", and it opens with the sentinel `Under agentic-trunk the lane takes no`. The earlier sentinel `names git-flow or agentic-trunk` is retired: it would have made `agentic-trunk` a variant that pushes a per-version release branch.

Judges: AC-AGT-021 Judge B (iii) pins the inserted sentence, the untouched git-flow-only wording and the absence of the retired coupling; Judge C guards the coupling over the three release-branch files; Judge B (iv) pins the 322-byte difference between the two `factory-dispatch.md` copies.

Named exclusions among the 15 template files that match the flow-name pattern: `auto-semantics.md` (line 376, github-flow archive path only), `spec-workflow-detail.md` (line 31, `merge_method`), `manager-git.md` and `manager-git.toml` (agent definitions; REQ-AGT-020 scopes rules and skills), `tab_schema.json` (wizard, OQ-5) and `git-strategy.yaml.tmpl` (defaults, REQ-AGT-015). The template `workflow.yaml` is edited for its comment only (REQ-AGT-007's inert statement names manual git-flow); it has no deployed byte-parity pair because the live file holds this repository's values.

## 6. Repo-local guidance plan

`gitflow-lane-protocol.md`: H1 becomes "agentic-trunk Lane Protocol (moai-adk-go, local-only)"; the opening blockquote keeps the 2026-08-27 and 2026-10-05 history but states the current name; section titles that carry "git-flow" are renamed; file path unchanged (OQ-6). `AGENTS.local.md` section 4: heading names; section 4.0 intro states the four-step model once; 4.1 keeps its `[SUPERSEDED]` markers. `repo-local-pr-policy.md`: "git-flow transition" wording becomes the model name. The three `.moai/docs` doctrine notices: one line each naming the current name. Substance of every rule stays; t1830's sentences are kept where it already rewrote them.

## 7. File inventory (for the Tier L count)

Go source (8): `loader_workflow_disposition.go`, `loader_integration_branch.go`, `integration_target_guidance.go`, `session_worktree_automerge.go`, `integration.go` (diagnostic text only if needed), `doctor_git_strategy_workflow.go`, `statusline/renderer.go`, one new `statusline` provider file. Go tests (13): the matching `_test.go` files plus `reconcile_test.go`, `github_flow_sweep_guard_test.go`, and the three consumer tests of AC-AGT-025 (`hook/closure_push_test.go`, `escalation/newapi_test.go`, the contract push-check test file). Template sources (9): `delivery.md`, `factory-dispatch-mechanics.md`, `factory-dispatch.md`, `factory-dispatch-core.md`, `worktree-integration.md`, `delivery-policy.md`, `spec-workflow.md`, `moai-ref-git-workflow/SKILL.md`, `workflow.yaml` (comment); `manager-git.md` and `manager-git.toml` are excluded by name (section 5). Deployed projections (8): the six rules, `delivery.md` and the git-workflow skill. Repo-local (7): `git-strategy.yaml`, `gitflow-lane-protocol.md`, `AGENTS.local.md`, `repo-local-pr-policy.md`, three `.moai/docs` notices. Docs-site (4): `moai-sync.md` per locale.
