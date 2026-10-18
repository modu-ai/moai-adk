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
| `EmptyTargetCause` / `EmptyTargetGuidance` | git-flow causes, fixes name "workflow: github-flow (target main)" | agentic-trunk never reaches an empty cause (target is `main`); the two git-flow fix strings add "or agentic-trunk (target main)" | two strings |
| `invalid workflow` diagnostics (`integration.go:212`, doctor, guidance) | print `AllowedWorkflows()` | print 5 values | free (dynamic) |
| `workflowStandingBranches` (doctor) | 4 phrases | add: "single trunk main; cards land serially in the integration worktree; one batch pull request per lead_push_threshold landings; tagged release" | one case |
| `TestGitHubFlowSweepGuard` arming (`== WorkflowGitHubFlow`) | github-flow only | stays so (OQ-2) | one added assertion in the arming subtest |
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

A binary built before M1 reads `agentic-trunk` as invalid: `IntegrationTarget` is empty, `worktree done`/`sweep`, `goal approve`, `factory merge ready`/`complete` and the card diff base refuse an empty target (t1453 M1), the doctor warns, `integration acquire` warns. The installed `moai` is one shared binary and every running `moai mcp-server` keeps the build it started with. Therefore:

1. M1 to M5 land with the live value still `git-flow`.
2. The leader runs the repo-local rc build (protocol section 9: tree-built binary, `rm -f` then `cp`, `scripts/verify-local-install.sh`) from the integration worktree and reconnects each lane's `moai mcp-server`.
3. Only then does the one-line live flip land (M6), as a commit of its own touching `.moai/config/sections/git-strategy.yaml` alone, so a revert is a single `git revert` of that commit.
4. The doctor of the tree-built binary (invoked by path, judging-build commit named next to the tree HEAD, verification-claim-integrity section 2.2) must report `agentic-trunk` OK before the flip lands (REQ-AGT-017).

Rollback: revert the M6 commit; `git-flow` plus `develop_branch: main` is a state every shipped binary understands.

## 5. Template-First, and what must not change for users

Edits go to `internal/template/templates/**`, then `make build` regenerates the embedded catalog and the deployed `.claude/**` copies are brought in line by the repo's documented sync. Rules about `agentic-trunk` are additive sentences ("a second workflow uses this route"); the defaults, `github-flow` as "the distributed default", and every `[HARD]` rule stay. `git-strategy.yaml.tmpl` gets no value change.

`delivery.md` gets a third strategy block. Design of that block: value domain line lists three values; the table gains a row (`agentic-trunk | main plus WT-* branches | integration-worktree merge, no PR | multi-lane projects landing locally with batch pull requests`); a `Strategy: agentic-trunk` section reuses the git-flow `WT-*` route steps 1 to 5, 7 and 8 (no pull request, merge-window coordination, enter the designated integration worktree, `git merge --no-ff`, resolve, exit and report) and OMITS the push step (the leader's batch pull request is the push); a non-`WT-*` branch under agentic-trunk reports the branch state and stops (no per-card pull request exists in this model). The unmatched-value stop keeps its behavior with the domain extended.

## 6. Repo-local guidance plan

`gitflow-lane-protocol.md`: H1 becomes "agentic-trunk Lane Protocol (moai-adk-go, local-only)"; the opening blockquote keeps the 2026-08-27 and 2026-10-05 history but states the current name; section titles that carry "git-flow" are renamed; file path unchanged (OQ-6). `AGENTS.local.md` section 4: heading names; section 4.0 intro states the four-step model once; 4.1 keeps its `[SUPERSEDED]` markers. `repo-local-pr-policy.md`: "git-flow transition" wording becomes the model name. The three `.moai/docs` doctrine notices: one line each naming the current name. Substance of every rule stays; t1830's sentences are kept where it already rewrote them.

## 7. File inventory (for the Tier L count)

Go source (8): `loader_workflow_disposition.go`, `loader_integration_branch.go`, `integration_target_guidance.go`, `session_worktree_automerge.go`, `integration.go` (diagnostic text only if needed), `doctor_git_strategy_workflow.go`, `statusline/renderer.go`, one new `statusline` provider file. Go tests (10): the matching `_test.go` files plus `reconcile_test.go` and `github_flow_sweep_guard_test.go`. Template sources (10): `delivery.md`, `factory-dispatch-mechanics.md`, `factory-dispatch.md`, `factory-dispatch-core.md`, `worktree-integration.md`, `delivery-policy.md`, `spec-workflow.md`, `moai-ref-git-workflow/SKILL.md`, `workflow.yaml` (comment), and `manager-git.md` / `manager-git.toml` only if they enumerate the value domain (M5 measures). Deployed projections (about 10). Repo-local (7): `git-strategy.yaml`, `gitflow-lane-protocol.md`, `AGENTS.local.md`, `repo-local-pr-policy.md`, three `.moai/docs` notices. Docs-site (4): `moai-sync.md` per locale.
