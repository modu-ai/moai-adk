# research.md — SPEC-GITSTRAT-WORKFLOW-READER-001

All evidence below re-verified in-tree on worktree WT-git-flow-reader (base = local develop b1bd81b23) on 2026-09-14, by manager-spec during plan phase, unless attributed otherwise.

## §1 Discarded premise (correction record)

**Original card t656 premise: "git_strategy.<mode>.workflow has 0 production readers" — DISCARDED as stale.**

Cards t449 and t637 already landed the reader (provenance attributed to the orchestrator's git grep on develop b1bd81b23, re-verified in-tree by this agent):

| Card | Landed | Evidence in-tree (measured 2026-09-14) |
|---|---|---|
| t449 | `LoadGitFlowDevelopBranch` — single-key reader for the git-flow integration branch | `internal/config/loader_integration_branch.go` header comment "card t449"; function at line 34 |
| t637 | `GitFlowIntegrationConfig` struct + `IsGitFlow()` — separates "is this project git-flow" from "what is its develop branch"; connected `develop_branch` as the acquire integration target | same file, lines 43-83; t637 sync-audit recorded loader functions 100% covered; `git log -S LoadGitFlowIntegrationConfig -- internal/cli/integration.go` → `ba0725be8` (t637) landed the `integration.go:283` consumer |
| t655 | Connected `develop_branch` as the session-exit auto-merge integration target | `git log -S` → `1d0f08f10` (SPEC-WORKTREE-KEY-WIRING-001) created `internal/cli/session_worktree_automerge.go` (consumer lines 75,169). t655 did NOT touch integration.go:283 |

## §2 In-tree verified evidence (this run, this tree)

- `internal/config/types.go:107` — `ModeProfile.Workflow string \`yaml:"workflow"\`` (per-mode profile struct manual/personal/team). Confirmed by Read.
- `internal/config/loader_integration_branch.go` — `const gitFlowWorkflow = "git-flow"` (line 25); `LoadGitFlowIntegrationConfig` reads `.moai/config/sections/git-strategy.yaml` via `ActiveModeProfile()`; `GitFlowWorkflow = profile.Workflow == "git-flow"`; `Manual = mode == "manual"`; `IsGitFlow() = Manual && GitFlowWorkflow`; every failure path yields zero value. Confirmed by Read (full file).
- Production consumers (grep, non-test files only): `internal/cli/integration.go:283,318` (acquire integration-target resolution + t637 warn-only fallback) and `internal/cli/session_worktree_automerge.go:75,169` (auto-merge gate: `!gitFlow.IsGitFlow() || gitFlow.DevelopBranch == ""`). TWO consumer sites — corrected per plan-audit iter1 D5: `internal/config/loader_slot_lease.go` is NOT a consumer; its line-6 comment says it is *modelled on* `LoadGitFlowDevelopBranch` (design precedent) and it reads `workflow.slot_lease.default_max_duration`; non-test `LoadGitFlowDevelopBranch(` callers from slot-lease code: 0.
- Existing tests: `internal/config/loader_integration_branch_test.go` (t637 suite — M1 extends this file).
- Fail-open single-key reader precedent: `internal/config/loader_worktree_base.go` (`LoadWorktreeBaseBranch`). Doctor-diagnostic precedent: `internal/cli/doctor_worktree_base.go` (four-state check, distinct repair next-steps, seams for test injection).
- Templates: `internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl` — `workflow: github-flow` at lines 18, 50, 86 (all 3 modes). Confirmed by grep. `internal/config/defaults.go:763,775,788` also set `Workflow: "github-flow"` for all three profiles. **Template default = github-flow confirmed** (card claim verified).
- shipped_key_inventory: `internal/config/testdata/shipped_key_inventory.yaml` lines 500/557/623 carry `git_strategy.{manual,personal,team}.workflow` with `class: W / evidence: reader` (generic label — M4 sharpens it).
- Wizard: no workflow write path — `internal/cli/wizard_config_test.go` shows the wizard writing only `mode` and `provider` into git-strategy.yaml (basis for D3 deferral in spec.md §C.3).
- Dogfood local config (this repo's `.moai/config/sections/git-strategy.yaml`): manual mode carries `workflow: git-flow`; personal/team carry `github-flow` — the two shipped value classes are already live in the wild, which is why the 4-value set must include both.

## §3 Key behavioral fact motivating D1

Today the reader is binary: `Workflow == "git-flow"` or not. `github-flow` (the shipped DEFAULT), `gitlab-flow`, `release-flow`, a typo, and `trunk-based` are all indistinguishable at the reader. Consumer consequences are identical (caller fallback), so this SPEC's change is purely additive diagnosability + interpretation — no existing path changes behavior. That is what makes characterization-first tractable: the M1 suite pins a stable surface.

## §4 Open items

Open questions: none — the card was operator-resolved at dispatch; design decisions D1-D4 are owned in spec.md §C. The wizard question (D3) is deferred with rationale, not open. (Repair note: this section previously carried the literal clarification-marker token in a negative count statement; rephrased per plan-audit iter1 D8 — a resolution statement containing the token still trips the mechanical grep gate.)
