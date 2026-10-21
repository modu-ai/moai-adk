---
id: SPEC-GIT-STRATEGY-AGENTIC-TRUNK-001
title: "The agentic-trunk workflow value: name the lane -> serial landing into main (integration-window local merge, or the card's pull request where only the primary checkout holds main) -> leader release batch per lead_push_threshold -> tagged release model in git_strategy.manual.workflow and reflect it in the reader, consumers, doctor, statusline and guidance"
version: "0.4.0"
status: in-progress
created: 2026-10-18
updated: 2026-10-21
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/config, internal/cli, internal/statusline, internal/template, .claude/rules, .claude/skills, .moai/config, docs-site"
lifecycle: spec-anchored
tags: "agentic-trunk, git-strategy, workflow-value, disposition, integration-target, statusline, doctor, guidance, template-first"
tier: L
related_specs: [SPEC-GITSTRAT-WORKFLOW-READER-001, SPEC-GITHUB-FLOW-DEFAULT-001, SPEC-LOCAL-MAIN-FLOW-001, SPEC-PRIMARY-LANDING-ONLY-001, SPEC-LANDING-SURFACE-RULES-001]
---

# SPEC-GIT-STRATEGY-AGENTIC-TRUNK-001

Card: **t1831** (Class C, P1, plan -> run -> sync; operator directive 2026-10-17, naming delegated to the lane). Transferred scope: item F of t1830 (SPEC-PRIMARY-LANDING-ONLY-001, plan held at the audit ceiling).

## HISTORY

| Date | Author | Change |
|------|--------|--------|
| 2026-10-18 | manager-spec | Initial plan-phase artifacts (0.1.0). Evidence read in tree `6041c38a4` (branch `WT-agentic-trunk`, local main after t1846); quantified in research.md. |
| 2026-10-18 | manager-spec | 0.2.0: plan-audit iteration 1 repairs: D1-D7 (RED-now cells authored at plan time on tree `f4175587119b5cfeb8d8348fc74cada8764e89ed`; the `factory-dispatch.md` template/deployed divergence recorded and decided; one cutover precondition set P1-P5 across REQ-AGT-017, plan M6 and design section 4; judges rebuilt for AC-AGT-015, -020 and -023; the DevelopBranch consumers added as REQ-AGT-025 and AC-AGT-025). Optional D8-D11 folded: GEARS rewording of REQ-AGT-018 and REQ-AGT-024, template-match exclusions named, REQ-AGT-019 sub-item judges added, the raw-merge demotion recorded in design section 5. 25 requirements, 25 acceptance criteria. |
| 2026-10-18 | manager-spec | 0.3.0: plan-audit iteration 2 repairs: D1, D2 and the optional/codex findings. D1: REQ-AGT-021 now carries the one recorded exception (the live-only tail sentence of `factory-dispatch.md` line 222, commit `7ff326d99`, card t1815), so requirement and AC-AGT-021 agree. D2: REQ-AGT-020 states that `agentic-trunk` follows neither the per-version release branch nor a lane push, and the line-222 edit no longer attaches `agentic-trunk` to the `release/vX.Y.Z` push (the `names git-flow or agentic-trunk` sentinel is retired and its absence is judged). Optional: AC-AGT-020 judges the whole agentic-trunk block for any push instruction; AC-AGT-023 is non-vacuous at N0 = 0 and rejects a bare label; the P4 precondition of REQ-AGT-017 is probed over all running servers, not one project directory. 25 requirements, 25 acceptance criteria. |
| 2026-10-18 | manager-spec | 0.3.1: plan-audit iteration 3 repair: D1 only. AC-AGT-020 Judge 6 now ends the agentic-trunk block at the next sibling (a heading of level 1 to 4, or the next `##### Strategy:` heading, outside a fenced snippet) instead of at any heading, so a push placed after a sub-heading or after a `#` comment line inside a fenced snippet is counted; two failing controls added (ledger E26). D2 to D6 are recorded as known debt in progress.md, not repaired. 25 requirements, 25 acceptance criteria. |
| 2026-10-21 | manager-spec | 0.4.0: mid-run rescope per the operator-corrected directive relayed by the leader (operator verbatim: "지금 사용하는 흐름은 agentic-trunk — 최근 추가된 워크플로우"). The dedicated `main` integration worktree was abolished on 2026-10-21; the primary checkout now holds `main` as a shared surface (verified: `git worktree list` row 1 `/Users/goos/MoAI/moai-adk-go 28b727065b [main]`, no dedicated main worktree). Delivery redesigned as the target-holder routing (design B, the PR-edge hybrid; candidate A, serial landing on the shared primary, assessed and rejected in design.md section 1): agentic-trunk keeps the integration-window local merge where a dedicated worktree holds `main` (REQ-AGT-006/007/008 re-aimed with the where-clause) and takes the card's own pull request, merged by the leader with no auto-merge request (the `factory_card_pr.go:444` `gh pr merge --auto` conflict resolved toward the operator's leader-only-merge principle), where only the primary checkout holds it. `lead_push_threshold` redefined as the leader's release-batch cadence metric (REQ-AGT-010). The M6 cutover is RESTORED in P1-P5 shape with the 2026-10-21 bridge value `github-flow` as its starting state (REQ-AGT-017/018); M0-M5 implementations stand as generic product capability, unchanged. M5b (delivery-routing delta) added to plan.md; M7 re-aimed at the hybrid flow. Fresh RED-now cells observed on tree `a6c69e154b` (ledger E27 to E32). 25 requirements, 25 acceptance criteria unchanged in count. |

## 1. Context and problem

1. **The operating model has no name in the product.** This repository runs: card worktrees branch from local `main`; each card lands serially in one dedicated integration worktree through `moai integration merge --card` with scoped tests; roughly every `lead_push_threshold` (20) landings the leader opens one batch pull request from `release/main-batch-YYYYMMDD` with one CI run; `main` is then tagged for release. Its live config value is `git_strategy.manual.workflow: git-flow` with `develop_branch: main` (`.moai/config/sections/git-strategy.yaml`), a label from the retired develop era. The yaml carries a comment that still says "workflow stays git-flow until the GitHub Flow residue sweep lands".
2. **`git-flow` misdescribes it and `github-flow` is wrong in the other direction.** The model has no `develop` branch and no per-card pull request. The operator rejected `github-flow` because it misleads (Jev 0.92) and chose `agentic-trunk` (Jev 0.72 first choice; Codex runner-up).
3. **The reader rejects any new value.** The allowed set is exactly four values (`loader_workflow_disposition.go`); an unknown value classifies invalid, resolves an empty integration target, and the base-branch consumers (`worktree done`/`sweep`, `goal approve`, `factory merge ready`/`complete`, card diff base) refuse an empty target. Flipping the live yaml to a value the running binary does not know therefore stops the factory, not only a diagnostic.
4. **Guidance is spread over the whole tree.** 704 tracked files match the flow-name pattern; most are SPEC and report history that this SPEC leaves alone. The live operating descriptions are in the repo-local rule and local guide, in six distributed rule files and the git-workflow skill plus their deployed copies, in the sync delivery skill (whose value domain is "exactly `github-flow` and `git-flow`" and which stops delivery on any other value), and in the docs-site `moai-sync` page (4 locales).

## 2. Goal

Give the model one canonical value, `agentic-trunk`, with an explicit and tested behavior mapping whose delivery route follows the target holder — the integration-window local merge where a dedicated worktree holds `main`, the card's own pull request merged by the leader where only the primary checkout holds it — make the product read it (reader, consumers, doctor, statusline), reflect the name in repo-local and distributed guidance, and switch this repository's live value last, after a binary that understands it is installed and after the 2026-10-21 bridge value `github-flow` has served its purpose. `git-flow` and `github-flow` stay valid and behave as today; the distributed default stays `github-flow` for every mode.

## 3. Requirements (GEARS)

### 3.1 Value set and reader

- REQ-AGT-001: The workflow value set shall contain `agentic-trunk` as a fifth value appended after the four existing values, and `github-flow`, `git-flow`, `gitlab-flow` and `release-flow` shall classify and resolve their integration targets exactly as before.
- REQ-AGT-002: The disposition classifier shall classify `agentic-trunk` as valid and not git-flow (the class `github-flow` has), by exact case-sensitive match.
- REQ-AGT-003: The integration target of `agentic-trunk` shall be the fixed branch `main`, independent of the `develop_branch`, `environment`, `release_branch_prefix` and `main_branch` keys.
- REQ-AGT-004: When the active mode profile's workflow is `agentic-trunk`, the strategy reader shall report the integration target `main` in every `git_strategy.mode`, and shall project the legacy develop-branch seam to `main` only in manual mode.
- REQ-AGT-005: The strategy reader shall expose a local-integration predicate that is true for manual-mode `git-flow` and for manual-mode `agentic-trunk` and false otherwise, and the existing git-flow-only predicate shall keep its present answers.

### 3.2 Behavior mapping in consumers

- REQ-AGT-006: While the workflow is `agentic-trunk`, `factory complete` and `factory merge ready` shall route the card by the target holder: where a worktree other than the primary checkout holds the integration target `main`, they shall take the integration-window local-merge path and shall open no pull request; where only the primary checkout holds `main`, they shall open the card's pull request, record `pr-open`, and shall issue no pull-request merge request of any kind — the merge is the leader's act, judged by the operator's leader-only-merge principle (the required checks green, a real review artifact present, the branch up to date), and once the leader has merged the pull request a later completion run shall record `merged-pr`.
- REQ-AGT-007: Where session-exit auto-merge is enabled and the project is manual-mode `agentic-trunk`, the auto-merge shall merge the session branch into `main` only where a worktree other than the primary checkout holds `main`, under the guards manual `git-flow` has today; where only the primary checkout holds `main` it shall skip with a notice naming the missing dedicated landing surface, and its skip notice shall name both workflows.
- REQ-AGT-008: Where a worktree other than the primary checkout holds the integration target and `moai integration acquire` runs in a manual-mode `agentic-trunk` project without `--branch`, it shall record the window against that worktree and shall write no caller-fallback warning; while only the primary checkout holds the target, the pull-request route of REQ-AGT-006 applies and no integration window is taken in the primary's working tree.
- REQ-AGT-009: The allowed-set listings in the invalid-value diagnostics shall include `agentic-trunk`, and each fix suggestion that names `github-flow (target main)` shall also name `agentic-trunk`.
- REQ-AGT-010: While the workflow is `agentic-trunk`, the keys `develop_branch`, `release_branch_prefix` and `rc_version_format` shall be ignored without any diagnostic, the key `lead_push_threshold` shall count card landings (integration-window merges on the dedicated-worktree route, merged card pull requests on the primary route) toward the leader's release-batch cadence, and the keys `merge_method` and `worktree_base_branch` shall keep their present meaning.
- REQ-AGT-011: The card-landing checks of `moai worktree done` and `moai worktree sweep` shall treat a card landed under `agentic-trunk` as they treat a card merged with a merge commit, and shall read no workflow value.

### 3.3 Display and compatibility

- REQ-AGT-012: When the active workflow is `agentic-trunk`, the doctor check "Git Strategy Workflow" shall report status OK with a message naming the flow, the integration target `main` and the standing-branch interpretation (a single trunk, cards landing one at a time — through the integration-window local merge, or the card's pull request where only the primary checkout holds `main` — a leader release batch per `lead_push_threshold` landings, tagged release).
- REQ-AGT-013: Where the active workflow is a valid value other than `github-flow`, the statusline shall show the workflow value beside the branch, and for `github-flow`, an absent, unreadable or invalid value it shall render byte-identically to today, adding no git invocation.
- REQ-AGT-014: When `moai update` runs against a project whose manual-mode workflow is `agentic-trunk`, it shall preserve that value verbatim.

### 3.4 Defaults, guards and the live configuration

- REQ-AGT-015: The shipped defaults (the template `git-strategy.yaml.tmpl` and the built-in default configuration) shall keep `github-flow` for the manual, personal and team profiles.
- REQ-AGT-016: The github-flow sweep guard shall stay armed only for `github-flow`, and its arming test shall assert that an `agentic-trunk` root does not arm it.
- REQ-AGT-017: The live value of this repository's `git_strategy.manual.workflow` shall change from the bridge value `github-flow` — set by the leader on 2026-10-21 while the installed binary (build `8249bff8c3`) predated the branch and lacked the value — to `agentic-trunk` only in a separate commit for which all five preconditions hold: (P1) the commits that implement REQ-AGT-001 to REQ-AGT-016, REQ-AGT-020, REQ-AGT-021 and REQ-AGT-025 (plan milestones M1 to M5 and the delivery-routing delta M5b) are its ancestors, (P2) the installed `moai` binary is byte-identical to the build of the commit's parent and reports that commit, (P3) that installed binary's doctor check reports `agentic-trunk` as OK, (P4) no running `moai mcp-server` process on the machine predates the installed binary, whichever project directory records it: the doctor's MCP server version check reports no stale server and the oldest running server started after the install (a server of another project that is left running is named in the progress record), and (P5) the leader runs the cutover with the operator's confirmation while no card pull request of this repository sits in `pr-open` awaiting the leader's merge and no `factory complete` run is mid-transition.
- REQ-AGT-018: When the live value of this repository's `git_strategy.manual.workflow` is `agentic-trunk`, the file `.moai/config/sections/git-strategy.yaml` shall carry `workflow: agentic-trunk` in the manual profile only, keep `lead_push_threshold: 20`, carry no `develop_branch`, `release_branch_prefix` or `rc_version_format` key in the manual profile, and carry no comment that presents `git-flow` or the bridge value `github-flow` as the standing workflow.

### 3.5 Guidance

- REQ-AGT-019: The repo-local guidance (`.claude/rules/local/gitflow-lane-protocol.md` title and section titles, `AGENTS.local.md` section 4, `.claude/rules/local/repo-local-pr-policy.md`, and the notices of the three `.moai/docs` doctrine files) shall name this repository's model `agentic-trunk`, state its four steps, keep the rule file's path, and leave every rule's substance unchanged.
- REQ-AGT-020: The distributed rules and skills that describe the local `WT-*` integration route as a `git-flow` variant shall name `agentic-trunk` as the second workflow using that route, where they describe the per-version release-branch push of the git-flow variant they shall state that `agentic-trunk` takes no per-version release branch and runs no lane push, the sync delivery skill shall list `agentic-trunk` in its value domain and route it without a push step, and every file shall keep `github-flow` as the distributed default.
- REQ-AGT-021: The edited template sources shall each be byte-identical to their deployed counterparts after the template build, except that the deployed `factory-dispatch.md` shall keep the one live-only sentence that ends its line 222, which the template omits on purpose (commit `7ff326d99`, card t1815).
- REQ-AGT-022: The docs-site `moai-sync` page shall list `agentic-trunk` in all four locales (ko canonical, then en, ja, zh) in the same landing.
- REQ-AGT-023: The edits to `AGENTS.local.md` and `gitflow-lane-protocol.md` shall carry both this SPEC's flow naming and the Integration-Worktree Rule of SPEC-PRIMARY-LANDING-ONLY-001 without removing either.

### 3.6 Verification discipline

- REQ-AGT-024: While this SPEC's acceptance criteria are judged, the judges shall use test runs anchored by `-run`, shall add no test skip, and shall not widen any timeout; where a package suite is red on main for a reason outside this SPEC, the acceptance criteria shall still name the anchored tests that decide.

### 3.7 Consumers of the legacy develop-branch seam

- REQ-AGT-025: While the workflow is `agentic-trunk` in manual mode and the manual profile carries no `develop_branch` key, the closure push guard, the contract push check and the card-base resolution shall use `main` as the integration branch, with the outcome a git-flow profile whose `develop_branch` is `main` has today.

## 4. Behavior mapping (summary; the code-level mapping is in design.md section 2)

| Behavior | agentic-trunk follows | Evidence |
|---|---|---|
| Delivery of a card | target-holder routing: git-flow's local integration window where a worktree other than the primary holds `main`; github-flow's per-card pull request — opened by `factory complete`, merged by the leader, no auto-merge request — where only the primary checkout holds it | REQ-AGT-006 |
| Integration target | github-flow: fixed `main` | REQ-AGT-003 |
| Standing branches | github-flow: `main` only, no `develop`, no per-release branch cut by lanes | REQ-AGT-010 |
| Session-exit auto-merge | git-flow: manual-mode gate, same guards — only where a dedicated worktree holds `main`; skip with notice on the primary route | REQ-AGT-007 |
| Window recording by `acquire` | git-flow: records the worktree holding the target (dedicated-worktree route only) | REQ-AGT-008 |
| Leader release batch | neither: leader-run cadence of one release batch per `lead_push_threshold` landings — the `release/main-batch-YYYYMMDD` batch PR on the dedicated-worktree route, a tagged release batch on the pull-request route | REQ-AGT-010 |
| Release branch per version (`release/vX.Y.Z` lane merge) | not followed | REQ-AGT-010, REQ-AGT-020 |

## 5. Open questions (defaults recorded; none blocks run unless marked)

| Id | Question | Recommended default |
|---|---|---|
| OQ-1 | Is the target fixed `main`, or does it honor `develop_branch` / `main_branch`? | Fixed `main` (REQ-AGT-003). Alternate: honor `main_branch` when set. |
| OQ-2 | Does the sweep guard arm for `agentic-trunk`? | No in this SPEC (REQ-AGT-016); pre-measure the armed finding count at M4 and open a follow-up card if it is zero. |
| OQ-3 | Who runs the cutover (install the tree-built rc, reconnect every `moai mcp-server`, flip the live value)? Blocks M6 only. | Decided (leader ruling): the leader runs the cutover as the operator-requested rc build of the repo-local protocol section 9, with the operator's confirmation, while no card pull request sits in `pr-open` awaiting the leader's merge (P5), because replacing the installed binary affects every session on the machine. Preconditions P1 to P5 are REQ-AGT-017; the starting state is the bridge value `github-flow` (2026-10-21). |
| OQ-4 | Statusline rule: label every valid non-`github-flow` value, or `agentic-trunk` only? | Every valid non-`github-flow` value (REQ-AGT-013). Needs a note to t1830, whose REQ-PLO-009 requires byte-identical render today. |
| OQ-5 | Offer `agentic-trunk` in the project wizard (`tab_schema.json`, 3 option lists)? | No: it is a factory operating value set by editing the yaml; the wizard keeps its two simple choices. |
| OQ-6 | Rename `gitflow-lane-protocol.md`? | Keep the path: 222 SPEC files and 16 other tracked files outside `.moai/specs` and `.moai/reports` (including `internal/cli/codex_review_scope.go`) cite it. |
| OQ-7 | `agentic-trunk` in a non-manual mode: keep the develop-seam gate? | Yes (REQ-AGT-004): the seam stays manual-only as for `git-flow`; the target is mode-independent as for `github-flow`. |

## 6. Exclusions

### Out of Scope — behavior changes beyond the value

- Changing how the leader judges a pull request before merging it: the operator's leader-only-merge principle (required checks green, a real review artifact present, the branch up to date) is applied as written; this SPEC removes only the auto-merge request, not the leader's judgment procedure.
- Automating the leader's merge decision, the release-batch close or the tagged release: they stay leader-run steps.
- Re-provisioning a dedicated integration worktree for this repository: the primary checkout holds `main` as a shared surface by operator decision (2026-10-21); the dedicated-worktree route stays a product capability for projects that run one.
- A new disposition class: `agentic-trunk` takes the existing valid-non-git-flow class.

### Out of Scope — default values and user projects

- Changing the shipped default of any mode, or making `agentic-trunk` the default for any user project.
- Offering `agentic-trunk` in the wizard (OQ-5 default).
- Migrating existing user projects' values.

### Out of Scope — the audit-side gates and history

- Arming the github-flow sweep guard for `agentic-trunk` (OQ-2 default); a follow-up card owns it.
- Editing SPEC bodies, `.moai/reports/`, `CHANGELOG.md` history, or the already-superseded develop-era text marked `[SUPERSEDED]` in `AGENTS.local.md`.
- Repairing the red main test `TestNoHardcodedContexts` (card t1858).

### Out of Scope — adjacent landing-surface work

- Everything in SPEC-PRIMARY-LANDING-ONLY-001 scopes A to E (integration-worktree designation key, primary staleness follow step, detached-HEAD statusline, card base ref). This SPEC only reconciles the shared files (REQ-AGT-023) and notes the statusline interaction (OQ-4).
