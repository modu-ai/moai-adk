# Research — SPEC-GIT-STRATEGY-AGENTIC-TRUNK-001

Measured in tree `6041c38a4` (branch `WT-agentic-trunk`, local main after t1846), 2026-10-18, by reading code and `git ls-files` / `git grep`. Every count below names its pattern. Pattern P = `git-flow|gitflow|github-flow|githubflow|GitHub Flow|git flow`, case-insensitive.

## R1. Where the flow name appears (tracked files matching P: 704)

| Area | Files | Disposition under this SPEC |
|---|---:|---|
| `.moai/specs/**` | 418 | untouched (history and SPEC bodies) |
| `.moai/reports/**` and `reports/**` | 90 | untouched |
| Go test files | 75 (490 matching lines) | 8 to 10 edited or extended; the rest untouched |
| Go source files | 38 (212 matching lines) | 6 to 8 edited (R2) |
| `internal/template/templates/**` | 15 | 8 to 10 edited (R4) |
| `.claude/**` deployed (rules, skills, agents) | 16 | projections of the template sources, plus 2 local-only rules |
| `.claude/rules/local/**` | 2 (`gitflow-lane-protocol.md` 5 lines, `repo-local-pr-policy.md` 3 lines) | edited |
| `.moai/docs/**` | 9 (`git-workflow-doctrine.md` 13, `git-local-workflow-doctrine.md` 5, `gitflow-integration-chain.md` 4, others 1 to 6) | 3 notices edited |
| `AGENTS.local.md` | 14 lines | edited (section 4) |
| `AGENTS.md`, `README*` | 0 | untouched |
| `docs-site/content/**` | 8 files (4 locale `moai-sync.md`, 3 `hooks-guide.md`, 1 other) | 4 `moai-sync.md` edited |
| `.github/**`, `scripts/**` | 13 | untouched (CI wiring and cutover scripts, not operating descriptions) |
| `.moai/config/sections/**` | 2 (`git-strategy.yaml` 6 lines, `shipped_key_inventory` testdata) | `git-strategy.yaml` edited |
| other testdata / json | 7 | untouched |

Files affected by the SPEC (recounted at version 0.2.0): about 8 Go source + 13 Go test + 9 template source + 8 deployed projections + 6 repo-local guidance + 1 live config + 4 docs-site = 49. That is above the 15-file line of Tier L.

## R2. The reader and its consumers (code read, not inferred)

- `internal/config/loader_workflow_disposition.go`: constants `WorkflowGitHubFlow`, `WorkflowGitFlow`, `WorkflowGitLabFlow`, `WorkflowReleaseFlow`; `AllowedWorkflows()` returns exactly four; `ClassifyWorkflowDisposition` maps `git-flow` to `DispositionGitFlow`, the other three to `DispositionValidNonGitFlow`, everything else (including an empty string) to `DispositionInvalid`; `WorkflowIntegrationTarget` is the D2 table: github-flow returns the fixed `main`, git-flow the trimmed `develop_branch`, gitlab-flow the `environment` key, release-flow the `release_branch_prefix` key.
- `internal/config/loader_integration_branch.go`: `gitFlowWorkflow = "git-flow"` (a second literal beside `WorkflowGitFlow`); `LoadGitFlowIntegrationConfig` sets `Manual` (mode is manual), `GitFlowWorkflow` (active profile workflow equals git-flow), `DevelopBranch` only when both hold, and `IntegrationTarget` from the D2 table (git-flow through the gated `DevelopBranch`). `IsGitFlow()` is `Manual && GitFlowWorkflow`. Fan-in: 14 non-test call sites of `LoadGitFlowIntegrationConfig` (`grep -rn` over `internal cmd`), 5 of `LoadGitFlowDevelopBranch`.
- The only places that compare the workflow VALUE (not the disposition) in non-test Go: `factory_card_pr.go:84` (`== WorkflowGitHubFlow`, the PR-delivery gate), `integration_target_guidance.go:61` (`case WorkflowGitFlow`), `doctor_git_strategy_workflow.go:28-36` (standing-branch phrase per flow), `loader_integration_branch.go:104,122` (the literal), and `internal/template/github_flow_sweep_guard_test.go:154` (arming). Everything else is reached through the predicates, so a new value needs a decision in exactly those places.
- `factoryGitHubFlow(root)` is true only for `github-flow`. `factoryCompleteCard` (`factory_card.go:2072`) and `factory merge ready` (`factory_merge.go:173`) take the pull-request edge on it and the integration-window path otherwise. A value that is not `github-flow` therefore already lands locally through the window with `moai integration merge --card`; agentic-trunk needs no change here, only a test that pins it.
- `IsGitFlow()` consumers: `session_worktree_automerge.go:176` (session-exit auto-merge needs manual git-flow plus a non-empty develop branch) and `integration.go:528` (the t637 caller-fallback warning). `DevelopBranch` consumers: `integration.go:466` (acquire target), `loader_slot_lease.go`, `contract_pushcheck.go:37`, `escalation/newapi.go:128`, `hook/closure_push.go:154,161`.
- Landing checks: `worktree/landing_predicate.go` keys on ancestry, patch-id and PR state; its header comment names git-flow and github-flow, but no code reads the workflow value (`grep` for the value literals in `internal/cli/worktree` returns comments only).
- Settings surfaces: the web settings schema exposes `mode`, `worktree_base_branch` and `merge_method` for git_strategy, not `workflow` (`internal/settings/schema_sections.go:127-150`); no enum of workflow values exists outside `AllowedWorkflows()`.

## R3. The shipped defaults and the update path

- `internal/config/defaults.go:1332,1352,1365`: manual, personal and team profiles all default to `github-flow`; the template `git-strategy.yaml.tmpl` lines 18, 58, 94 say the same. Those must not move.
- `internal/cli/update/reconcile_test.go` `TestUpdate_GitStrategyValuesSurvive` (AC-UPM-021, release-blocking) pins that an operator-set `git-flow` survives `moai update` against the template's `github-flow`. The same hazard applies to `agentic-trunk`; it needs its own pin.
- `internal/cli/update_legacy_snapshot_test.go` (t1216) hand-edits `workflow: github-flow` to `workflow: git-flow` plus `develop_branch` in its fixtures.

## R4. Distributed guidance that describes the local route

- `factory-dispatch-mechanics.md` lines 142-149 and 243: "under github-flow (the distributed default) the verb opens a pull request ... the git-flow variant (integration window, `--no-ff` merge, release-branch push) ... applies only where the project's git strategy names git-flow".
- `factory-dispatch.md:219,222` and `factory-dispatch-core.md:64`: "[HARD] git-flow only: a verified lane merges its branch into `release/vX.Y.Z` itself ... under github-flow ... `moai factory complete`'s PR edge". That sentence describes a per-version release branch, which is NOT the agentic-trunk model (the batch branch is `release/main-batch-YYYYMMDD`, opened by the leader).
- `worktree-integration.md:48`, `spec-workflow.md:47`, `delivery-policy.md:15,18`, `sync/delivery.md:282`: "Where the integration branch is held by a dedicated worktree (the git-flow `WT-*` integration route) ..." — the sentence t1846 landed (commit `a67623ebd`). agentic-trunk uses this route; the parenthetical names only git-flow.
- `sync/delivery.md:22,29`: "The canonical value domain is exactly `github-flow` and `git-flow`" and "Unmatched value stops delivery ... report the offending value together with the canonical domain {github-flow, git-flow}". With the live value flipped and no delivery.md change, `/moai sync` Step 3.2 would stop delivery in this repository. Lines 268-326 hold the git-flow strategy (WT-* route, 8 steps; step 6 pushes the integration branch).
- `moai-ref-git-workflow/SKILL.md:38,51`: generic "GitHub Flow" and "GitFlow" sections.
- `tab_schema.json` (3 option lists, lines 352-360, 471-479, 660-668): `github-flow` and `git-flow` only.
- `docs-site/content/{ko,en,ja,zh}/workflow-commands/moai-sync.md` lines 113-132: the value domain is "two" and the routes are described per value.

## R5. Repo-local state

- `.moai/config/sections/git-strategy.yaml` manual profile: `workflow: git-flow`, `develop_branch: main`, `release_branch_prefix: release/`, `rc_version_format: vX.Y.Z-rc.N`, `lead_push_threshold: 20`, `merge_method: squash`; the comment at line 28 says "workflow stays git-flow until the GitHub Flow residue sweep lands". `worktree_base_branch: main`.
- `gitflow-lane-protocol.md` (200 lines): H1 "git-flow Lane Protocol", opening blockquote "switched from GitHub Flow to git-flow", sections 1 to 11. Path cited by 222 SPEC files and 16 other tracked files (`git grep -c 'gitflow-lane-protocol' -- ':!.moai/specs' ':!.moai/reports'` lists 16 files), including runtime code `internal/cli/codex_review_scope.go`.
- `AGENTS.local.md` section 4.0 states the model (landing in the integration worktree, resync, batch release procedure, batch-closing rule); section 4.1 carries the superseded develop and GitHub Flow chains.
- SPEC-PRIMARY-LANDING-ONLY-001 (t1830, plan held at the audit ceiling; read from `.moai/worktrees/t1830`): its section G names the overlap with this card: it rewrites the landing-surface sentences (AGENTS.local.md section 4.0 intro and items 1 to 4; protocol sections 1, 2, 9, 11), this card renames the flow wording in the same files; "whichever card lands second reconciles". Its REQ-PLO-009 requires the statusline to render byte-identically when the checkout has a branch (its own change). Its out-of-scope list names scope F as moved here.

## R6. RED-now observations (tree 6041c38a4)

| Command | stdout | exit |
|---|---|---:|
| `grep -c agentic-trunk internal/config/loader_workflow_disposition.go` | `0` | 1 |
| `grep -c 'workflow: agentic-trunk' .moai/config/sections/git-strategy.yaml` | `0` | 1 |
| `grep -n 'stays git-flow' .moai/config/sections/git-strategy.yaml` | `28:        # workflow stays git-flow until the GitHub Flow residue sweep lands` | 0 |
| `grep -c agentic-trunk internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md` | `0` | 1 |
| `grep -n 'canonical value domain is exactly' internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md` | `22:The canonical value domain is exactly `github-flow` and `git-flow`.` | 0 |
| `grep -c 'workflow: github-flow' internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl` | `3` | 0 |
| `grep -c agentic-trunk .claude/rules/local/gitflow-lane-protocol.md` | `0` | 1 |
| `grep -c agentic-trunk internal/statusline/renderer.go` | `0` | 1 |
| `grep -lr agentic-trunk docs-site/content` | (empty) | 1 |
| `go test ./internal/config -run '^TestAgenticTrunk' -count=1 -v` | `testing: warning: no tests to run` / `PASS` / `ok  github.com/modu-ai/moai-adk/internal/config  0.323s [no tests to run]` | 0 |

The last row is the vacuous-green trap in its pure form: a selector that matches nothing exits 0 and prints `ok`. Every Go-test criterion in acceptance.md therefore names the `--- PASS:` lines it expects and treats `no tests to run` as red.

## R7. Not observed

- The armed finding count of `TestGitHubFlowSweepGuard` (set `MOAI_SWEEP_GUARD_ARMED=1`) was not run; it decides OQ-2's follow-up, and M4 measures it.
- (Superseded by R8: the RED-now observations for `internal/cli`, `internal/statusline`, `internal/cli/update`, `internal/template`, `internal/hook` and `internal/escalation` were taken in the plan-audit iteration 1 repair and live in acceptance.md.)
- The statusline render cost of one extra small YAML read per render was not measured.
- Whether another process caches the workflow value (long-lived `moai mcp-server` processes keep the build they started with — the server instructions of this session say so) was not enumerated; the cutover step treats every running server as stale.

## R8. Measurements added by the plan-audit iteration 1 repair (tree f4175587119b5cfeb8d8348fc74cada8764e89ed)

- Parity of the template/deployed pairs in scope: seven `cmp -s` pairs exit 0 (`factory-dispatch-mechanics`, `factory-dispatch-core`, `worktree-integration`, `delivery-policy`, `spec-workflow`, `sync/delivery.md`, `moai-ref-git-workflow/SKILL.md`); `factory-dispatch.md` exits 1 with exactly one differing line (222). The deployed line carries a trailing sentence about `moai worktree sweep` that the template lacks. `git log -S'disposing the card'` shows commit `7ff326d99` (card t1815, "drop live-only worktree-sweep sentence leaked into factory-dispatch mirror") removed it from the template deliberately, so the divergence is intended and the reconciliation keeps it.
- `DevelopBranch` consumers (non-test Go, all through `config.LoadGitFlowDevelopBranch`): `contract_pushcheck.go:37` (empty prints `inactive (no git-flow integration branch configured)`), `hook/closure_push.go:153-161` (`integrationBranchFor`, with a primary-checkout fallback through the git common directory), `escalation/newapi.go:126-143` (`cardBase`: non-empty is the sole merge-base candidate, empty falls to origin/HEAD, `main`, `master`). `loader_slot_lease.go` only mentions the function in a comment. Plus the two `LoadGitFlowIntegrationConfig` readers already in design section 2 (`integration.go:466`, `session_worktree_automerge.go:176-183`).
- The doctor already carries the two cutover witnesses: the `Git Strategy Workflow` row (a `warn` for an unknown value on the installed build, exit 0) and the `MCP Server Version` row (`ok ... match the installed binary (<commit>)`, or `warn ... running MCP server is stale`); `moai doctor --check "<row name>"` runs one row. `scripts/verify-local-install.sh <built> <installed> <short-sha>` checks byte identity and the installed binary's reported commit.
- The sweep guard disarms on this repository (`sweep guard disarmed: git_strategy workflow="git-flow" (armed only when "github-flow")`); the existing `arming-switch` subtest passes today, which is why a dedicated test is the flip judge.
- Template files matching the flow-name pattern: 15. Edited: six rule files, `sync/delivery.md`, `moai-ref-git-workflow/SKILL.md`, `workflow.yaml` (comment). Excluded by name: `auto-semantics.md`, `spec-workflow-detail.md`, `manager-git.md`, `manager-git.toml`, `tab_schema.json`, `git-strategy.yaml.tmpl`.

## R9. Measurements added by the plan-audit iteration 2 repair (tree c6e590507b39e27acc3305a3a1aa3dc488aef0c0)

- `git diff --stat f4175587119b5cfeb8d8348fc74cada8764e89ed..c6e590507b39e27acc3305a3a1aa3dc488aef0c0 -- internal .claude AGENTS.local.md docs-site .moai/config .moai/docs` prints nothing (exit 0): every code, template, config and guidance observation of R1 to R8 holds on both trees.
- `factory-dispatch.md` line 222 is the git-flow variant ("a lane merges its own branch into `release/vX.Y.Z` itself ... pushes `release/vX.Y.Z`, and leaves the batch pull request with the leader"); its scoping sentence "it applies only where the project's git strategy names git-flow" occurs once per copy. Widening that sentence to `agentic-trunk` was the iteration 2 defect D2: it made `agentic-trunk` push a per-version release branch, which spec.md section 4 excludes. `factory-dispatch-core.md` line 64 and `factory-dispatch-mechanics.md` lines 142 to 149 and 243 carry the same git-flow wording. The two copies of `factory-dispatch.md` measure 25924 bytes (template) and 26246 bytes (deployed): the 322-byte difference is the live-only tail sentence of line 222.
- `sync/delivery.md` (template): the git-flow strategy block measures 34 non-blank lines of which 8 match `[Pp]ush|gh pr`; no `agentic-trunk` block exists. The next heading after the git-flow block is `#### Step 3.3: PR Ready Transition (Team Mode)`, which is where the new block ends.
- MCP servers (observed 2026-10-18, installed build `v3.2.0-rc.36-t1822c`, commit `ee981a932`, install time `2026-10-16T06:56:39`): `pgrep -f '^moai mcp-server$'` listed 32 processes (31 a few minutes later); `moai doctor --check "MCP Server Version"` read 28 matching; `ps eww -p <pid>` showed `CLAUDE_PROJECT_DIR` equal to the primary checkout for 28 of them (all with a record under the primary `.moai/state/mcp-server`) and `/Users/goos/MoAI/mo.ai.kr` for 4 (no record here). Worktrees of this repository hold stale record files from earlier sessions (for example `.moai/worktrees/t1465` holds 3), but no live pid has one. The earlier statement that every session records under the primary holds for this repository's sessions today; it does not cover a server started with another project directory, which is why the P4 precondition adds an all-process probe. `pgrep` without anchors also matches the invoking shell, whose command line carries the same words (observed: 31 against 32 anchored on one read), so the probe anchors its pattern.
- Landing-surface sentences in the repo-local files: `performs landing and resync only` occurs once in each of `AGENTS.local.md` (line 184) and `gitflow-lane-protocol.md` (line 37); `Integration-Worktree Rule` occurs 0 times in each. SPEC-PRIMARY-LANDING-ONLY-001 REQ-PLO-002 lists `AGENTS.local.md:184` among the sentences it drops, so the t1846 phrase count may fall to 0 when t1830 lands; the AC-AGT-023 baselines are therefore recorded at M7 start rather than frozen here.
