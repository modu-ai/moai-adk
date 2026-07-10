---
id: SPEC-AGENT-TEAM-RETIRE-001
title: "Agent Teams retirement + replacement dynamic-workflow pair"
version: "0.1.2"
status: completed
created: 2026-07-11
updated: 2026-07-11
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/cli, internal/config, internal/web, internal/settings, .claude/workflows"
lifecycle: spec-anchored
tags: "agent-teams, retirement, dynamic-workflow, cleanup, web-console, template, lockfile"
era: V3R6
tier: L
related_specs: [SPEC-V3R6-AGENT-TEAM-REBUILD-001, SPEC-CLIFIX-CRITICAL-001, SPEC-SUBCOMMAND-RETIRE-001]
---

# SPEC-AGENT-TEAM-RETIRE-001 — Agent Teams retirement + replacement dynamic-workflow pair

## HISTORY

| Date | Version | Change | Author |
|------|---------|--------|--------|
| 2026-07-11 | 0.1.0 | Initial plan-phase draft. Scope A: remove MoAI's STATIC Agent Teams layer (role_profiles config + team_spawn coordination primitives + team rules/skills) while preserving the Claude Code NATIVE teammate runtime. Scope B: two replacement dynamic workflows (`sync-audit-4dim.js`, `plan-research-fanout.js`). Tier L. All tree anchors verified against HEAD at authoring time. | manager-spec |
| 2026-07-11 | 0.1.1 | Both plan-phase clarifications resolved by user decision (orchestrator-relayed): (1) team/glm.md = migrate-essentials-then-delete → new REQ-ATR-022 + AC-ATR-027 (grep token `CG Mode (Claude + GLM`, verified 0-count in target at plan time); (2) auto-select thresholds = prose-only SSOT (design.md D8 adopted) → REQ-ATR-010 extended + AC-ATR-028. [NEEDS CLARIFICATION] markers removed from plan.md. 22 REQ / 28 AC. | manager-spec |
| 2026-07-11 | 0.1.2 | plan-auditor iter-1 FAIL 0.86 fixes (D1-D8, all auditor findings re-verified with live greps): D1 AC-ATR-015 dangling-ref sweep re-rooted (explicit roots; `.claude/worktrees/` + `agent-memory/` excluded) + 4 extra sweep files; D2 fieldsets.templ has NO team fieldset (schema-driven UI) → REQ-ATR-008/AC-ATR-013 re-targeted to schema_sections.go workflow.team FieldDef rows; D3 `WorkflowTeamAutoSelection()` accessor + types_test.go blocks added to removal enumeration (REQ-ATR-005); D4 REQ-ATR-011 expanded to workflows/** submodules (case-insensitive — `## Team Mode` capital-T survivor class); D5 vacuous AC tokens replaced (tombstone `Mode 3 — RETIRED` baseline-0, run.md `Agent Teams 모드` baseline-1); D6 +AC-ATR-029 key-token sweep; D7 15 funcs; D8 AC-ATR-020/021/025 grep hardening. 22 REQ / 29 AC. | manager-spec |

## §A. Context and Intent

MoAI's Agent Teams mode is **disabled by default** since the Sonnet 5 / Opus 4.8
re-design (`orchestration-mode-selection.md` §C.1: "team mode is nonetheless kept
disabled by default because subagent fanout (Mode 4) and sequential sub-agent
(Mode 5) cover the practical surface with lower coordination overhead and lower
token cost"). The static team layer MoAI built on top of the native runtime —
`workflow.team.*` config (`TeamConfig` / `RoleProfileEntry` /
`TeamAutoSelectionConfig`), the `internal/cli/team_spawn.go` coordination
primitives (mailbox, roster, task-claim, team state), the web-console team
fieldset, the `team-protocol.md` / `team-pattern-cookbook.md` rules, and the six
`.claude/skills/moai/team/*.md` skills — is now dead weight: every symbol in
`team_spawn.go` has **0 non-test callers** (verified by grep at authoring time;
the only substring hit is the unrelated `internal/session` `AppendTaskLedger`).

**Boundary principle**: keep the Claude Code NATIVE teammate runtime
(`Agent(name=...)`, `~/.claude/teams/` registry, `teammateMode` launcher field,
`internal/cli/worktree/team_launch*` P1-P4 patterns, `internal/session`
`MergeTeamCheckpoints`); remove ONLY MoAI's static team layer. Two shared
dependencies must migrate BEFORE deletion (Phase 0 de-risk): the cross-platform
advisory file-lock pair (→ `internal/lockfile`) and the concurrency-safe
`ClaimTask` primitives (→ `internal/cli/taskledger`), because the
SPEC-CLIFIX-CRITICAL-001 P0 regression test (`TestClaimTaskAppend_Repro`)
depends on them.

**Replacement (Scope B)**: the two team use-cases that carried real value —
parallel multi-dimension quality judgment and parallel plan-phase research — are
re-provided as dynamic workflows in the user-owned `.claude/workflows/`
directory, following the validated `codemaps-extract.js` house style and the
Anthropic multi-agent findings (research parallelizes well; coding does not;
~15x token multiplier demands narrow, high-value fan-out; 4-element subagent
prompts).

## §B. Scope Summary

**In scope — Scope A (removal, 10 phases)**: Phase 0 dependency migration
(lockfile + taskledger); Phase 1 `team_spawn.go` + tests deletion; Phase 2
config type removal (`TeamConfig`, `RoleProfileEntry`, `TeamAutoSelectionConfig`,
`WorkflowConfig.Team` field, `defaults.go` team block); Phase 3 web console
(`f.workflow.team.*` i18n keys, `RoleProfileNames()` + team form fields, team
fieldset in `.templ` + regenerated `_templ.go`); Phase 4 `moai workflow lint`
repurpose (role_profiles → model_routing_profiles closed-set validation);
Phase 5 rules deletion/shrink; Phase 6 team skills deletion + workflow routing
cleanup; Phase 7 template mirror + `make build` + CI guards; Phase 8 test
reconciliation + whole-repo green.

**In scope — Scope B**: `.claude/workflows/sync-audit-4dim.js` (Context → 4
parallel Judges → in-script harmonic-mean Verdict) and
`.claude/workflows/plan-research-fanout.js` (3-4 lens Explore fan-out →
Synthesize), both local-only (`.claude/workflows/` is user-owned, NOT
template-managed per `dynamic-workflows.md`).

**Preserve (Phase 9 — do NOT touch)**: see REQ-ATR-006/007 preservation
invariants.

**Out of scope** — see §E.

## §C. Requirements (GEARS notation)

### C.1 De-risk migration (Phase 0)

- **REQ-ATR-001** (Ubiquitous): The codebase shall provide the cross-platform
  advisory file-lock primitives currently embedded in
  `internal/cli/team_spawn_lock_unix.go` / `team_spawn_lock_windows.go` (Unix
  `syscall.Flock`; Windows in-process per-path mutex fallback) as a standalone
  `internal/lockfile` package with an exported API, preserving the
  `//go:build !windows` / `//go:build windows` build-tag pair, migrating
  `team_spawn_lock_unix_test.go` alongside, and preserving the documented
  Windows cross-process limitation comment verbatim.
- **REQ-ATR-002** (Ubiquitous): The codebase shall provide the concurrency-safe
  task-claim primitives (`ClaimTask`, `TaskClaimer`, and their transitively
  required helpers currently in `internal/cli/team_spawn.go`) in a new
  `internal/cli/taskledger` package; the SPEC-CLIFIX-CRITICAL-001 regression
  test `TestClaimTaskAppend_Repro` (`internal/cli/clifix_critical_repro_test.go`)
  and the `TestClaimTask` / `TestClaimTaskConcurrent` coverage tests shall be
  re-pointed to the migrated symbols and remain green — the P0
  map-RMW / O_APPEND regression guard is not lost.
- **REQ-ATR-003** (State-driven): While the Phase 0 migration has not landed
  with a green whole-repo test suite (`go test ./...` exit 0), the deletion
  phases (M1 onward) shall not begin.

### C.2 Static team layer removal — Go (Phases 1-2)

- **REQ-ATR-004** (Ubiquitous): `internal/cli` shall not contain
  `team_spawn.go`, `team_spawn_test.go`, `team_spawn_lock_unix.go`,
  `team_spawn_lock_windows.go`, or `team_spawn_lock_unix_test.go` after
  removal (coordination-layer tests retired; lock tests and ClaimTask tests
  live in their migrated packages per REQ-ATR-001/002).
- **REQ-ATR-005** (Ubiquitous): `internal/config` shall not contain the
  `TeamConfig`, `RoleProfileEntry`, or `TeamAutoSelectionConfig` type
  declarations, the `WorkflowConfig` `Team` field (`yaml:"team"`), the
  `Team:` defaults block in `defaults.go` (auto-selection thresholds,
  role-profile keys and entries), or the `WorkflowTeamAutoSelection()`
  accessor in `workflow_accessors.go` (the type family's non-test source
  consumer).

### C.3 Preservation invariants (Phase 9 — native runtime untouched)

- **REQ-ATR-006** (Unwanted behavior): The removal shall not modify or delete
  any of the following native-teammate-runtime or unrelated-namesake surfaces:
  (a) the `internal/config` `RoleProfile` struct carrying the `Sandbox` field
  (@MX:SPEC SPEC-V3R2-RT-003 — a SEPARATE type from the deleted
  `internal/cli` `RoleProfile`); (b) the `GitStrategyConfig` mode profiles
  including the `Team ModeProfile` field and its `defaults_test.go`
  assertions; (c) `internal/session` `MergeTeamCheckpoints` and the session
  store; (d) `internal/cli/worktree/team_launch*`, `swarm_registry.go`, and
  `handoff_guidance.go` (P1-P4 launch patterns); (e) `teammateMode` handling
  in the `moai cg` / `moai glm` / `moai cc` launcher paths and
  `settings.local.json`; (f) `~/.claude/teams/` registry semantics.
- **REQ-ATR-007** (Unwanted behavior): The web-console key removal shall not
  delete the `f.git_strategy.team.*` i18n key family or the
  `f.git_strategy.mode.opt.team` option label — these belong to the preserved
  GitStrategy Team mode profile; a bare-substring `team` sweep is prohibited.

### C.4 Web console (Phase 3)

- **REQ-ATR-008** (Ubiquitous): The web console shall not render Agent Teams
  configuration: the `f.workflow.team.*` i18n key family shall be removed from
  `internal/web/assets/i18n.js` (with team-mentioning copy in
  `sec.workflow.desc` re-worded); `RoleProfileNames()` and the team form-field
  definitions shall be removed from `internal/settings/schema_sections.go` and
  their consumers (`internal/settings/schema.go`,
  `internal/web/schema_label.go`); the five `workflow.team` FieldDef rows
  (enabled / delegate_mode / max_teammates / default_model /
  require_plan_approval — content-token anchor `"workflow", "team"`) shall be
  removed from `internal/settings/schema_sections.go`. NOTE (auditor D2,
  verified): `internal/web/fieldsets.templ` carries NO team fieldset at plan
  time (schema-driven UI since the web-console redesign) — templ regeneration
  applies ONLY if run-phase re-measure finds team markup in the `.templ`
  source; when it does, edit the `.templ` and regenerate (never hand-edit the
  generated file alone).

### C.5 agentlint repurpose (Phase 4)

- **REQ-ATR-009** (Event-driven): When `moai workflow lint` is invoked after
  removal, the command shall validate `workflow.model_routing_profiles`
  entries against the closed sets (reusing `internal/config`
  `ValidateModelRoutingProfiles`) and shall exit non-zero on violations; the
  subcommand shall not become a no-op after its `role_profiles` isolation
  checks (`validateRoleProfiles` in
  `internal/cli/agentlint/workflow_lint.go`) are removed.

### C.6 Rules and skills (Phases 5-6)

- **REQ-ATR-010** (Ubiquitous): The rules tree shall not contain
  `.claude/rules/moai/workflow/team-protocol.md` or
  `team-pattern-cookbook.md`; `orchestration-mode-selection.md` shall shrink
  Mode 3 to a retirement note (catalog-row tombstone; §C.1 capability-gate
  detail removed) and shall not retain the machine-readable-source pointer to
  `workflow.yaml auto_selection` in §B.1 — the §B.1 prose thresholds
  (domains≥3 / files≥10 / score≥7) become the sole SSOT (user-adopted
  design.md D8); `spec-workflow.md` shall not contain the Agent Teams
  variant methodology sections; no retained rule, skill, or agent file shall
  carry a dangling reference to a deleted team file, nor a residual
  Agent-Teams config-key token (`workflow.team.enabled`,
  `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS`) outside the Mode 3 retirement
  tombstone — sweep scope is the retained-surface enumeration in plan.md M3
  (dangling-ref checks are rooted at the authored trees only;
  `.claude/worktrees/` and `.claude/agent-memory/` are protected historical
  artifacts per §E and are excluded from every sweep).
- **REQ-ATR-011** (Ubiquitous): The skills tree shall not contain the
  `.claude/skills/moai/team/` directory (6 files: plan / run / sync / debug /
  review / glm); team-mode routing references shall be removed from the
  `.claude/skills/moai/workflows/` tree — top-level files AND submodules —
  with a CASE-INSENSITIVE run-phase re-measure (the plan-time lowercase grep
  missed the capital-T `Team Mode` survivor class). Plan-time enumeration
  (auditor D4, verified): `workflows/{plan,run,fix,review,mx,moai}.md`,
  `workflows/sync.md` (loading-table row), `workflows/sync/delivery.md`
  (frontmatter + `## Team Mode` section),
  `run/{mode-orchestration,task-decomposition,context-loading,phase-execution}.md`,
  `plan/spec-assembly.md`. The run.md `--mode` dispatch axis (`team` value,
  `MODE_TEAM_UNAVAILABLE` sentinel, CI sentinel audit) is handled per plan.md
  M3 with the sentinel-audit caution.
- **REQ-ATR-022** (State-driven, migrate-then-delete): While the essential
  CG Mode (GLM teammate) guidance from `.claude/skills/moai/team/glm.md`
  (LLM mode detection, prerequisites, tmux environment variables, error
  recovery) has not been migrated into
  `.claude/rules/moai/core/glm-web-tooling.md` (both trees — the rule has a
  template mirror), the team-skills deletion of REQ-ATR-011 shall not
  proceed; after the deletion, the migrated guidance shall remain present in
  the target rule, carrying the distinctive heading token
  `CG Mode (Claude + GLM` (verified 0-count in the target at plan time —
  non-vacuous anchor).

### C.7 Template mirror (Phase 7)

- **REQ-ATR-012** (Ubiquitous): The template tree
  (`internal/template/templates/`) shall mirror every rule / skill / config
  removal in this SPEC (team rules, team skills directory, the
  `.moai/config/sections/workflow.yaml` `team:` block), regenerated via
  `make build`, with the template-neutrality CI guard
  (`template-neutrality-check.yaml`) and the internal-content-leak /
  neutrality tests passing.
- **REQ-ATR-013** (Event-driven): When `moai init` deploys a fresh project
  from the post-removal binary, the deployed tree shall not contain team
  rules, team skills, or a `workflow.yaml` `team:` block
  (resurrection-negative — `moai update` shall likewise not re-deploy them).

### C.8 Test reconciliation (Phase 8)

- **REQ-ATR-014** (Ubiquitous): Test reconciliation shall remove exactly the
  Agent-Teams assertions — the `TeamConfig` default rows in
  `internal/config/defaults_test.go` (the GitStrategy `Team` ModeProfile
  assertions at the AC-GSS-005 block are preserved), the team blocks in
  `workflow_nested_test.go`, and `workflow_role_profiles_test.go` — and the
  whole repository shall build and test green: `go build ./...`,
  `GOOS=windows GOARCH=amd64 go build ./...`, and `go test ./...` all exit 0.

### C.9 Replacement dynamic workflows (Scope B)

- **REQ-ATR-015** (Ubiquitous): `.claude/workflows/sync-audit-4dim.js` shall
  exist implementing three phases — (1) **Context**: one read-only Explore
  agent (effort `medium`) returning a schema-forced
  `{spec_id, acceptance_criteria, changed_files, test_command}` object;
  (2) **Judge**: four parallel read-only judge agents (effort `xhigh`), one
  per dimension (Functionality / Security / Craft / Consistency), each
  returning a schema-forced
  `{dimension, score, findings[{severity,summary,file,evidence}], evidence_gaps[]}`
  object (score 0-1) under a skeptical-auditor stance requiring command +
  verbatim-output evidence; (3) **Verdict**: the harmonic mean `n/Σ(1/sᵢ)`
  computed IN SCRIPT JavaScript with a zero-score guard, judged against a
  threshold supplied via `args` (default 0.85).
- **REQ-ATR-016** (Event-driven + Unwanted behavior): When any judge agent
  returns null, the sync-audit workflow shall return verdict `INCOMPLETE`
  naming the missing dimension(s); the workflow shall not compute a harmonic
  mean over fewer than four dimensions, shall not spawn a meta-judge agent to
  aggregate verdicts, shall not delegate score arithmetic to an LLM, and
  shall not grant Write/Edit capability to judge agents.
- **REQ-ATR-017** (Capability gate): Where the audited SPEC's tier is M or L,
  the sync-audit 4-dimension workflow gate applies; Tier S SPECs shall not
  route through the workflow (documented in the script header and honored via
  `args.tier`).
- **REQ-ATR-018** (Ubiquitous): `.claude/workflows/plan-research-fanout.js`
  shall exist implementing two phases — (1) **Explore**: 3-4 parallel
  read-only agents (`agentType: 'Explore'`, effort `medium`), one per
  distinct lens from `args.lenses` (defaults: codebase-precedent /
  external-docs / constraints-risks / prior-SPEC-memory), each prompt
  carrying the 4 elements (objective + output format + tool guidance +
  boundaries, including "do NOT cover other lenses"), each returning
  fixed-heading markdown (NOT a forced schema) with a mandatory
  `### confidence_and_gaps` section where "NONE found" is a valid answer;
  (2) **Synthesize**: one agent (effort `high`) that marks cross-lens
  contradictions explicitly and never smooths them; the workflow shall return
  `{lenses, per_lens_reports, research_md}`.
- **REQ-ATR-019** (Event-driven): When two or more lenses return null, the
  plan-research workflow shall abort the Synthesize phase and return
  `insufficient_coverage` naming the failed lenses.
- **REQ-ATR-020** (Unwanted behavior): The plan-research workflow shall not
  run more than 4 lenses, shall not set effort `xhigh` on explorer agents,
  and shall not perform in-workflow file writes — `research.md` is written by
  manager-spec / the orchestrator OUTSIDE the workflow (SPEC-artifact
  ownership + read-only workflow discipline).
- **REQ-ATR-021** (Ubiquitous): Both workflow scripts shall follow the
  `codemaps-extract.js` house style: header doctrine comment (verdict
  scoping + determinism + read-only notes), `export const meta` with
  `name` / `description` / `phases`, deterministic script body (no wall-clock
  or random-number CALLS — `Date.now()` / `Math.random()` never invoked),
  `args` inputs with defaults, `label: '<stage>:<item>'` on agent calls, and
  null-filtering of agent results before aggregation.

## §D. Acceptance Criteria

The full machine-verifiable AC matrix (AC-ATR-001 … AC-ATR-029) lives in
`acceptance.md` (SSOT). Every REQ above maps to at least one AC; preservation
REQs map to STILL-EXISTS assertions, not absence checks.

## §E. Exclusions

The following are explicitly out of scope for this SPEC.

### Out of Scope — Native teammate runtime removal

- Removing or altering `Agent(name=...)` native spawning, the
  `~/.claude/teams/` registry, `teammateMode` in `settings.local.json`, tmux
  CG/GLM launchers (`moai cg` / `moai glm`), or
  `internal/cli/worktree/team_launch*` / swarm registry — all PRESERVED per
  REQ-ATR-006.

### Out of Scope — GitStrategy Team mode profile

- The `git_strategy.team` ModeProfile (mode="team" git workflow profile), its
  defaults, tests, and `f.git_strategy.team.*` web-console keys — an
  unrelated namesake, PRESERVED per REQ-ATR-006/007.

### Out of Scope — Template-shipping the Scope B workflows

- `.claude/workflows/*.js` is user-owned and NOT template-managed
  (`dynamic-workflows.md` § MoAI Integration Notes); the two new scripts are
  local-only and are NOT mirrored into `internal/template/templates/`.

### Out of Scope — sync-auditor / plan-auditor agent retirement

- The `sync-auditor` and `plan-auditor` agents remain the binding
  PASS/FAIL verdict owners. The Scope B workflows are execution vehicles a
  future SPEC may wire into those agents' flows; no agent file is modified
  here beyond dangling-reference cleanup (REQ-ATR-010).

### Out of Scope — CHANGELOG / README / docs-site

- CHANGELOG.md is owned by manager-docs (sync-phase); README and docs-site
  4-locale updates for team-mode removal are a follow-up sync/docs concern.

### Out of Scope — Historical artifacts

- Completed SPEC bodies, archived memory entries, and merged commit messages
  referencing team mode remain unchanged (immutability per
  sprint-round-naming.md § Legacy Aliases precedent).

## §F. Cross-References

- `.claude/rules/moai/workflow/orchestration-mode-selection.md` §C.1 — Mode 3
  default-disabled rationale (the retirement premise).
- `.claude/rules/moai/workflow/dynamic-workflows.md` — workflow primitive
  semantics (16-concurrent cap, resume caching, determinism, user-owned
  `.claude/workflows/`).
- `.moai/specs/SPEC-CLIFIX-CRITICAL-001/` — P0 concurrency fixes whose repro
  test gates Phase 0.
- `.moai/specs/SPEC-V3R6-AGENT-TEAM-REBUILD-001/` — the static team layer
  origin this SPEC retires.
- `research.md` §D — external rationale sources (Anthropic multi-agent
  research findings; building-effective-agents patterns; Claude Code
  workflows documentation).
