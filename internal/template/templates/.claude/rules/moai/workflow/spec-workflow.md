---
paths: "**/.moai/specs/**,**/.moai/config/sections/quality.yaml"
---

# SPEC Workflow

MoAI's three-phase development workflow with token budget management.

## Phase Overview

| Phase | Command | Agent | Token Budget | Purpose |
|-------|---------|-------|--------------|---------|
| Plan | /moai plan | manager-spec | 30K | Create SPEC document |
| Run | /moai run | manager-develop (per quality.yaml development_mode; cycle_type=ddd / tdd / autofix) | 180K | DDD / TDD / autofix implementation |
| Sync | /moai sync | manager-docs | 40K | Documentation sync |

Per the canonical agent catalog policy, the MoAI agent catalog consists of exactly 11 retained agents (`manager-spec`, `manager-develop`, `manager-docs`, `manager-git`, `manager-design`, `e2e-tester`, `plan-auditor`, `sync-auditor`, `builder-harness`, `super-advisor`, plus the Anthropic built-in `Explore` — per CLAUDE.md §4). 12 phantom and domain-expert agents (`manager-strategy`, `manager-quality`, `manager-brain`, `manager-project`, `claude-code-guide`, `researcher`, and the 6 `expert-*` agents) were archived offline during the catalog consolidation. For migration guidance and the per-archived-agent replacement pattern, see `.claude/rules/moai/workflow/archived-agent-rejection.md`.

## SPEC Phase Discipline

> L2/L3 worktree usage is opt-in. Default flow executes all phases on main checkout with a feature branch. See `.claude/rules/moai/workflow/worktree-integration.md` § Terminology Glossary for L1/L2/L3 layer definitions.

[ZONE:Frozen] [HARD] Every MoAI SPEC follows the three-phase lifecycle (plan → run → sync). How each phase transition is *triggered* depends on the **route** the SPEC takes. There are exactly TWO routes, and the route is determined by Tier (per § SPEC Complexity Tier) and the explicit `--pr` flag:

- **Route A — Hybrid Trunk main-direct (default; Tier S / Tier M):** manager-develop commits and pushes directly to `main`; there is NO per-phase PR and NO per-phase branch. Phase transitions are triggered by commit / push events (Conventional-Commit subjects pushed to `main` + green CI), NOT by PR merges. This is the 1-person-OSS Hybrid Trunk policy (CLAUDE.md §5 + `manager-develop-prompt-template.md` §B9).
- **Route B — PR route (Tier L OR explicit `--pr`):** `manager-git` creates a feature branch and opens a PR per phase (`gh pr create`); phase transitions are triggered by PR merges into `main`. This is the route the Late-Branch closure pattern (below) applies to.

The route governs the trigger vocabulary in § Phase Transitions below (commit/push event vs PR merge). Neither route changes the phase *ordering* (plan → run → sync) or the *artifact* set (per Tier).

**Route A — Hybrid Trunk main-direct (default, Tier S / M):**

| Step | Location | Command | Branch | Merge strategy | Lifecycle event (trigger) |
|------|----------|---------|--------|----------------|---------------------------|
| 1 (plan) | main checkout | `/moai plan SPEC-XXX` | `main` (direct) | n/a (no PR) | plan-phase artifacts committed + pushed to `main` |
| 2 (run)  | main checkout | `/moai run SPEC-XXX` | `main` (direct) | n/a (no PR) | run-phase commits pushed to `main` + tests green |
| 3 (sync) | main checkout | `/moai sync SPEC-XXX` | `main` (direct) | n/a (no PR) | single sync commit pushed to `main` (carries `implemented → completed`) |

**Route B — PR route (Tier L OR explicit `--pr`):**

| Step | Location | Command | Branch | PR strategy | Lifecycle event (trigger) |
|------|----------|---------|--------|-------------|---------------------------|
| 1 (plan) | main checkout | `/moai plan SPEC-XXX` | `plan/SPEC-XXX` | configured* | plan PR merged into main |
| 2 (run)  | main checkout (default) OR L2 SPEC worktree (opt-in) | (opt-in) `moai worktree new SPEC-XXX --base origin/main` then `/moai run SPEC-XXX`; OR `/moai run SPEC-XXX` on `feat/SPEC-XXX` branch in main checkout | `feat/SPEC-XXX` | configured* | run PR merged into main |
| 3 (sync) | same as Step 2 | `/moai sync SPEC-XXX` (same L2 worktree as Step 2 if L2 was used; otherwise same feature branch) | `sync/SPEC-XXX` (or `chore/SPEC-XXX-sync`) | configured* | sync PR merged into main |
| 4 (cleanup) | host checkout (only if L2 was created) | `moai worktree done SPEC-XXX` | n/a | n/a | L2 worktree disposed |

\* Route B PR strategy is the configured `merge_method` (`git_strategy.<mode>.merge_method`; one of `squash` | `merge` | `rebase`), **default `squash`**. Squash remains the documented recommendation — one squash commit per phase yields clean, revertable SPEC history — and is the value applied when `merge_method` is absent or unset. The method is configurable (per the per-mode `merge_method` field) so that workflows such as gitflow `release/*` may opt into a merge commit; the FROZEN default and its rationale are unchanged. Route A has no PR and therefore no `merge_method` — it pushes directly to `main`. Step 4 (worktree cleanup) applies to Route B only when an L2 worktree was created.

[ZONE:Frozen] [HARD] Step ordering rules:
- Step 1 (plan) MUST execute in main checkout on BOTH routes. NO L2/L3 worktree at this step. Plan artifacts are markdown only — no code conflict — and main-authored plans enable cross-SPEC reference for plan-auditor and parallel SPEC scoping. On **Route A** the plan-phase artifacts are committed + pushed directly to `main` (no branch). On **Route B**, the **Late-branch precondition (the Late-Branch closure contract)** applies: when `team.branch_creation.auto_enabled == false` in `git-strategy.yaml`, Step 1 entry requires `git rev-parse --abbrev-ref HEAD == main` (or the user's chosen `main_branch` if it differs). No `plan/SPEC-XXX` branch is created at Step 1; plan-phase commits land directly on `main` and are pushed only after Phase C `git switch -c plan/SPEC-XXX` at PR creation time.
- Step 2 (run) — **Route A** commits + pushes directly to `main` (no branch, no worktree). **Route B** SHOULD create a fresh L2 SPEC worktree from the plan-merged main HEAD (`--base origin/main`) if the user opted into L2/L3; otherwise continue on the `feat/SPEC-XXX` branch in main checkout. When L2 is used, worktree base alignment is a precondition for `Agent(isolation: "worktree")` correctness.
- Step 3 (sync) — **Route A** emits the single sync commit directly on `main` (carrying the `implemented → completed` transition; see § Phase Transitions). **Route B** SHOULD reuse the SAME L2 worktree as Step 2 if L2 was used; otherwise continue on the same feature branch in main checkout. Sync rotates codemap / MX / docs in the run-modified tree; spawning a fresh L2 worktree at sync would lose run-state context.
- Step 4 (cleanup) applies to **Route B only**. It MUST happen ONLY after BOTH run AND sync PRs are merged, and ONLY when an L2 worktree was created. Premature `moai worktree done` between run-merge and sync-merge breaks Step 3. **Late-branch closure (the Late-Branch closure contract):** when `auto_enabled == false`, after squash merge of run-PR and sync-PR, the user (or `manager-git` automation) MUST execute the canonical Late-branch closure step:

  ```bash
  git checkout main
  git fetch origin
  git reset --hard origin/main
  git pull origin main   # verify
  ```

  Post-condition: `git status --porcelain` returns empty AND `git rev-parse main` == `git rev-parse origin/main`. Failure mode: skipping this step leaves local main with un-squashed history that conflicts with the next `git pull`. For the complete 4-phase Late-branch invocation pattern (A→D), see `.claude/agents/moai/manager-git.md` § Late-Branch Invocation Pattern.

[SHOULD] Anti-patterns (advisory):
- Creating an L2/L3 worktree for plan (Step 1). Plan-in-worktree forces a base rebase after plan PR merge and prevents parallel SPEC plan visibility.
- Stacking plan + run in the same L2 worktree. Once the plan PR merges, the worktree base becomes stale; subsequent run work either rebases (extra cost) or proceeds against a stale tree (correctness risk).
- Disposing the L2 worktree after run merge but before sync merge. Sync re-enters the tree with codemap / MX / docs writes; the host checkout cannot stand in for a disposed worktree.

Cross-reference: see `.claude/rules/moai/workflow/worktree-integration.md` § SPEC-to-Worktree Mapping for per-step L2 worktree applicability and decision tree.

## Conditional Design Route (UI-surfaced SPECs)

A SPEC that declares a UI surface routes `plan → design → run` instead of the standard `plan → run → sync`. The design phase — driven by the `manager-design` agent (`.claude/agents/moai/manager-design.md`) and the `.claude/skills/moai/workflows/design.md` D1-D5 pipeline — enters AFTER plan-audit PASS + Implementation Kickoff Approval and BEFORE run-phase M1 commit. The conditional route is **additive**: it does not change the `plan → run → sync` ordering for non-UI SPECs.

UI-surface heuristic (either satisfies): explicit frontend-component / view / page deliverable in `acceptance.md`, OR `tier: L` + a frontend module (`module:` references a frontend package). Where neither holds, the route remains `plan → run → sync` — skip design entirely. Design is NOT a substitute for Implementation Kickoff Approval; it executes inside the already-approved run envelope, before the first implementation commit. See `.claude/skills/moai/workflows/design.md` for the D1-D5 pipeline and the H1-H9 handoff contract.

## Subcommand Classification (Pipeline vs Multi-Agent)

*control-flow style* axis. The classification governs which agents are spawned,
how the `--mode` flag is interpreted, and which CI guards apply.

| Subcommand   | Class          | 3-phase contract (localize → repair → validate)                | `--mode` honored? | Default mode | Valid `--mode` values | Sentinel on invalid mode | Reference                                                    |
|--------------|----------------|-----------------------------------------------------------------|-------------------|--------------|-----------------------|--------------------------|--------------------------------------------------------------|
| `/moai fix`      | Pipeline (Agentless) | Parallel Scan + Classify + MX context → Auto-Fix → Verify        | No (info log)     | n/a (pipeline-fixed) | n/a (any ignored)        | `MODE_FLAG_IGNORED_FOR_UTILITY` (info log only) | `.claude/skills/moai/workflows/fix.md`                       |
| `/moai mx`       | Pipeline (Agentless) | Pass 1 + Pass 2 → Pass 3 → Post-edit scan                         | No (info log)     | n/a (pipeline-fixed) | n/a (any ignored)        | `MODE_FLAG_IGNORED_FOR_UTILITY` (info log only) | `.claude/skills/moai/workflows/mx.md`                        |
| `/moai codemaps` | Pipeline (Agentless) | Explore → Analyze + Generate → Verify                             | No (info log)     | n/a (pipeline-fixed) | n/a (any ignored)        | `MODE_FLAG_IGNORED_FOR_UTILITY` (info log only) | `.claude/skills/moai/workflows/codemaps.md`                  |
| `/moai clean`    | Pipeline (Agentless) | Static Analysis + Usage Graph → Safe Removal → Test Verification  | No (info log)     | n/a (pipeline-fixed) | n/a (any ignored)        | `MODE_FLAG_IGNORED_FOR_UTILITY` (info log only) | `.claude/skills/moai/workflows/clean.md`                     |
| `/moai plan`     | Multi-Agent    | n/a — open-ended (mode-NA — mode not applicable)                      | Yes (rejects `pipeline`) | `autopilot` | (none — `--mode` ignored) | `MODE_PIPELINE_ONLY_UTILITY` (only on `pipeline`) | `.claude/skills/moai/workflows/plan.md`               |
| `/moai run`      | Multi-Agent    | n/a — open-ended (`autopilot` / `loop` / `team` per the run-mode contract)        | Yes (rejects `pipeline`) | `autopilot` (harness `minimal`/`standard`); `team` (harness `thorough` + prereqs) | `autopilot`, `loop`, `team` | `MODE_UNKNOWN`, `MODE_TEAM_UNAVAILABLE`, `MODE_PIPELINE_ONLY_UTILITY` | `.claude/skills/moai/workflows/run.md`                |
| `/moai sync`     | Multi-Agent    | n/a — open-ended (mode-NA — mode not applicable)                      | Yes (rejects `pipeline`) | `autopilot` | (none — `--mode` ignored) | `MODE_PIPELINE_ONLY_UTILITY` (only on `pipeline`) | `.claude/skills/moai/workflows/sync.md`               |
| `/moai loop`     | Multi-Agent (alias for `/moai run --mode loop`) | n/a — delegates to `/moai run` mode dispatch | Yes (alias semantics) | (inherits from `run --mode loop`) | (alias only — `--mode` resolves via `run`) | (delegates to `run` sentinels) | `.claude/skills/moai/workflows/loop.md`               |

### Mode Dispatch Cross-Reference


`/moai loop` is an alias for `/moai run --mode loop` per the mode-dispatch contract. The alias disposition is settled as **KEEP** (backward-compat): both routes resolve to the goal-preset sweep — `/moai run --mode loop` is a historical entry point and retiring it would be a breaking change, so keeping the alias preserves that entry point. Both routes invoke the same goal-preset sweep (the Ralph engine re-expressed as a goal preset) identically.

Mode precedence (hard-coded):

1. CLI flag `--mode <value>` — highest priority.
2. Config field `workflow.default_mode` in `.moai/config/sections/workflow.yaml`.
3. Harness auto-selection — lowest priority (per `harness.yaml` level).

Auto-selection rules:

- Harness `minimal` or `standard` → default mode = `autopilot`
- Harness `thorough` → default mode = `autopilot` (the former `team` auto-select is retired with the Agent Teams static layer; a forced `--mode team` emits `MODE_TEAM_UNAVAILABLE` and falls back to `autopilot` with a `[mode-auto-downgrade]` info log).

See `.claude/skills/moai/workflows/run.md` § Mode Dispatch for the per-skill dispatch rules.

### Pipeline Class — Contract

Pipeline-classified subcommands MUST satisfy:
- Three deterministic phases (localize → repair → validate); no LLM dispatcher selects the next phase.
- `Agent()` invocations are permitted only as **executor delegation within a phase** (e.g., `manager-develop` runs the coverage tool); never to decide phase order.
- When localize finds zero targets, exit with status `no-op` and exit code 0.
- When repair encounters an unresolvable error, fail-fast (no multi-agent fallback).
- The CI guard `internal/template/agentless_audit_test.go` enforces the no-LLM-dispatch rule via static text scan.

### Out of scope of this matrix

`/moai feedback` and `/moai review` are *not* yet classified.
See `spec.md` §1.2 (Non-Goals) — they are deferred to a future SPEC.

### Cross-references

- `--mode` flag matrix (defines `autopilot|loop|team|pipeline`).
- Pipeline regression guard: `internal/template/agentless_audit_test.go`.
- Pattern source: the Agentless pipeline pattern (localize → repair → validate, no LLM dispatcher).
- Research source: Xia et al. 2024.

## SPEC Complexity Tier (S/M/L)

The SPEC complexity classification taxonomy is referred to interchangeably as "Tier S/M/L" or "SPEC tiers" throughout this rule set.

[ZONE:Evolvable] [HARD] Every SPEC plan-phase classifies the SPEC into one of three Tier S/M/L levels before artifact creation begins. The tier determines the artifact set, the delegation prompt template applicability, and the plan-auditor PASS threshold. Origin: the workflow-lean root-cause fix for over-formalization observed in plan-phase abandonment.

| Tier | Scope guidance (LOC) | Files affected | Artifact set | plan-auditor PASS threshold |
|------|----------------------|----------------|--------------|------------------------------|
| S (Simple) | < 300 LOC | < 5 files | **2 files**: spec.md + plan.md (AC inline in spec.md §3) | 0.75 |
| M (Medium) | 300 - 1000 LOC | 5 - 15 files | **3 files**: spec.md + plan.md + acceptance.md | 0.80 |
| L (Large) | > 1000 LOC or constitutional | > 15 files | **5 files**: spec.md + plan.md + acceptance.md + design.md + research.md | 0.85 |

Tier judgment: performed as a Socratic AskUserQuestion in `spec-assembly.md` (Tier judgment Socratic question). The LOC thresholds are guidance, not enforcement — the implementer's judgment supplements the question.

Tier field in frontmatter: optional. The `tier:` YAML field carries the classification (enum: S | M | L). Documented in `.claude/rules/moai/development/spec-frontmatter-schema.md` as an optional field. Backward compatibility rule: when `tier:` is absent, the SPEC is treated as **Tier L** to preserve existing 5-artifact default behavior for pre-LEAN SPECs.

Section A-E delegation template (`manager-develop-prompt-template.md`): REQUIRED for Tier M/L delegations, OPTIONAL for Tier S. Tier S delegations MAY use minimal prompts (~500-800 tokens) covering only goal + deliverables + constraints + self-verification. See `.claude/rules/moai/development/manager-develop-prompt-template.md` § Applicability.

plan-auditor escalation: iter(N+1) aggregate score lower than iter(N) triggers a STOP signal and scope-reduction proposal — no unconditional further iteration. Maximum 3 plan-auditor iterations per SPEC plan-phase; after iter3, escalate via PASS-with-debt OR scope-reduction OR explicit user override. See `.claude/agents/moai/plan-auditor.md` § Retry Loop Contract.

Anti-pattern: classifying a 1000+ LOC SPEC as Tier S to skip overhead. Mitigation: plan-auditor first-pass score regression triggers a tier-up suggestion to the user; the Tier field is recorded in the SPEC for retrospective audit.

## Plan Phase

[ZONE:Frozen] [HARD] Execute in main checkout. NO worktree at this step. See § SPEC Phase Discipline (Step 1).

Create comprehensive specification using EARS format.

Optional pre-plan technique — **Blind Spot Pass**: when the domain is unfamiliar and unknown-unknowns are suspected, the orchestrator MAY run a Blind Spot Pass before plan-phase entry to surface the user's likely unknown-unknowns via `Agent(Explore)` read-only reconnaissance + an `AskUserQuestion` round. SSOT: `.claude/rules/moai/core/askuser-protocol.md` § Blind Spot Pass.

Sub-phases:
1. Research: Deep codebase analysis producing research.md artifact
2. Planning: SPEC document creation with EARS format requirements
3. Annotation: Iterative plan review cycle (1-6 iterations) before implementation approval

Token Strategy:
- Allocation: 30,000 tokens
- Load requirements only
- Execute /clear after completion
- Saves 45-50K tokens for implementation

Output:
- Research document at `.moai/specs/SPEC-XXX/research.md` (deep codebase analysis)
- SPEC document at `.moai/specs/SPEC-XXX/spec.md`
- EARS format requirements
- Acceptance criteria
- Technical approach

## Run Phase

[SHOULD] When user has opted into L2/L3 worktree, execute in a fresh L2 SPEC worktree: `moai worktree new SPEC-XXX --base origin/main`; otherwise execute on the `feat/SPEC-XXX` branch in main checkout. See § SPEC Phase Discipline (Step 2). Per the opt-in policy, L2/L3 worktree is opt-in; default is main checkout + feature branch.

Implement specification using configured development methodology.

Token Strategy:
- Allocation: 180,000 tokens
- Selective file loading
- Enables 70% larger implementations

Development Methodology (configured in quality.yaml development_mode):

### DDD Mode — ANALYZE-PRESERVE-IMPROVE

Best for existing projects with < 10% test coverage. Uses manager-develop agent with cycle_type=ddd.

**ANALYZE**: Read existing code, map domain boundaries, identify side effects and implicit contracts.
**PRESERVE**: Write characterization tests capturing current behavior. Create behavior snapshots for regression detection.
**IMPROVE**: Make small incremental changes. Run characterization tests after each change. Refactor with test validation.

### TDD Mode — RED-GREEN-REFACTOR (default)

Best for all development work, new projects, and brownfield with 10%+ coverage. Uses manager-develop agent with cycle_type=tdd.

**RED**: Write a failing test describing desired behavior. Verify it fails. One test at a time.
**GREEN**: Write simplest implementation that passes. No premature optimization.
**REFACTOR**: Clean up while keeping tests green. Extract patterns, remove duplication.

Brownfield enhancement: Pre-RED step reads existing code to understand current behavior before writing the failing test.

### Methodology Auto-Detection

| Project State | Test Coverage | Recommendation |
|--------------|---------------|----------------|
| Greenfield (new) | N/A | TDD |
| Brownfield | >= 10% | TDD |
| Brownfield | < 10% | DDD |

Manual override: `constitution.development_mode` in quality.yaml (nested under the top-level `constitution:` block), `MOAI_DEVELOPMENT_MODE` env var, or `moai init --mode <ddd|tdd>`.

### Pre-submission Self-Review

Before marking implementation complete: review full diff against SPEC acceptance criteria. Ask "Is there a simpler approach?" and "Would removing any changes still satisfy the SPEC?" Skip for single-file changes under 50 lines, bug fixes with reproduction test, or user-approved annotation cycle changes.

### Drift Guard

After each methodology cycle, compare planned files against actual modifications. Warns at <= 30% drift. Triggers re-planning (Phase 14) above 30%.

### Methodology delegation (team mode retired)

The Agent Teams static layer is retired; the run-phase methodology (DDD/TDD) is applied by a single `manager-develop` sub-agent (Mode 5), with multi-domain research fanned out via Mode 4 (parallel read-only `Agent()`) where warranted. See § Agent Teams Variant — RETIRED. The native `moai cg` teammate runtime is unaffected.

### MX Tag Integration

| Phase | TDD Action | DDD Action |
|-------|-----------|-----------|
| Test/Analyze | RED: add `@MX:TODO` | ANALYZE: 3-Pass scan, identify targets |
| Implement/Preserve | GREEN: remove `@MX:TODO` | PRESERVE: validate tags, add `@MX:LEGACY` |
| Refactor/Improve | REFACTOR: add `@MX:NOTE` | IMPROVE: update tags, add `@MX:NOTE` |

Success Criteria:
- All SPEC requirements implemented
- Methodology-specific tests passing
- 85%+ code coverage
- TRUST 5 quality gates passed
- MX tags added for new code (NOTE, ANCHOR, WARN as appropriate)

### Re-planning Gate

Detect when implementation is stuck or diverging from SPEC and trigger re-assessment.

Triggers:
- 3+ iterations with no new SPEC acceptance criteria met
- Test coverage dropping instead of increasing across iterations
- New errors introduced exceed errors fixed in a cycle
- Agent explicitly reports inability to meet a SPEC requirement

Communication path:
- Implementation agent (manager-develop) detects trigger condition
- Agent returns structured stagnation report to MoAI (agents cannot call AskUserQuestion)
- MoAI presents gap analysis to user via AskUserQuestion with options:
  - Continue with current approach (minor adjustments needed)
  - Revise SPEC (requirements need refinement)
  - Try alternative approach (re-spawn manager-develop with revised cycle_type, or escalate to a per-spawn `Agent(general-purpose)` specialist with domain-specific instructions per `.claude/rules/moai/workflow/archived-agent-rejection.md` migration table)
  - Pause for manual intervention (user takes over)

Detection method:
- Append acceptance criteria completion count and error count delta to `.moai/specs/SPEC-{ID}/progress.md` at the end of each iteration
- Compare against previous entry to detect stagnation
- Flag stagnation when acceptance criteria completion rate is zero for 3+ consecutive entries

Integration: Referenced by run.md Phase 14 and loop.md iteration checks

## Sync Phase

[SHOULD] When an L2 SPEC worktree was used in run, continue in the SAME L2 worktree as run; do NOT create a new L2 worktree. Otherwise, continue on the same feature branch in main checkout. See § SPEC Phase Discipline (Step 3). Per the opt-in policy, L2 worktree usage is opt-in.

Generate documentation and prepare for deployment.

Token Strategy:
- Allocation: 40,000 tokens
- Result caching
- 60% fewer redundant file reads

Output:
- API documentation
- Updated README
- CHANGELOG entry
- Pull request

## Context Management

/clear Strategy:
- After /moai plan completion (mandatory)
- When context exceeds 150K tokens
- Before major phase transitions

Progressive Disclosure:
- Level 1: Metadata only (~100 tokens)
- Level 2: Skill body when triggered (~5000 tokens)
- Level 3: Bundled files on-demand

## Phase Transitions

Each transition below is stated per route (per § SPEC Phase Discipline). Route A (Hybrid Trunk main-direct, Tier S/M default) triggers on commit/push events; Route B (PR route, Tier L OR `--pr`) triggers on PR merges. The phase *ordering* (plan → run → sync) is identical on both routes; only the trigger vocabulary differs.

Plan to Run:
- Trigger (Route A): plan-phase artifacts committed + pushed to `main` AND SPEC document approved (annotation cycle completed, user confirmed "Proceed").
- Trigger (Route B): Plan PR merged into main (squash) AND SPEC document approved (annotation cycle completed, user confirmed "Proceed").
- Pre-condition: plan.md records `plan_complete_at` + `plan_status: audit-ready` in progress.md; on Route B the plan PR is additionally in MERGED state.
- Action: Execute /clear, then `/moai run SPEC-XXX`. Route A runs directly on `main` in main checkout. Route B runs on `feat/SPEC-XXX` branch in main checkout (default); OR if the user opted into L2: `moai worktree new SPEC-XXX --base origin/main`, then `/moai run SPEC-XXX` inside the L2 worktree.
- Gate: `/moai run` Phase 1 (Plan Audit Gate) executes automatically before any implementation.
  See "Phase 1: Plan Audit Gate" section below for details.
- [ZONE:Evolvable] Plan Audit Gate skip policy (single authoritative contract):
  the orchestrator MAY skip Phase 1 re-execution and proceed directly to
  Phase 1 **IF AND ONLY IF ALL FOUR** of the following hold for the most recent
  plan-auditor verdict on the SPEC:
    1. **Verdict is `PASS`** (NOT FAIL, NOT INCONCLUSIVE, NOT BYPASSED).
    2. **Overall score ≥ 0.90.**
    3. **Artifact-hash unchanged** since that verdict — no plan-phase artifact
       (spec.md / plan.md / acceptance.md / research.md / design.md) has been
       modified since the audit that produced the verdict (equivalently on
       Route B: no plan-PR commit has landed since that verdict). Note: the
       mechanical hash subject is the `ComputeHash` 4-file plan-artifact set
       (spec.md / plan.md / acceptance.md / tasks.md — see § Report
       Persistence); research.md / design.md changes are a conservative input
       to the manual skip judgment, not part of the mechanical plan-artifact
       hash.
    4. **Within 24h** — the verdict was produced no more than 24 hours ago.
  If ANY of the four fails, Phase 1 re-executes (the gate is never disabled by
  harness level; see Gate Entry Condition below). When the skip is taken, the
  skip decision AND the four satisfied conditions MUST be recorded in the
  run-phase delegation prompt (Section A: Context) so downstream actors
  (manager-develop, auditors) can verify the skip rationale. This is the ONE
  authoritative skip contract — any other surface (e.g. the skill-layer
  `run/phase-execution.md`) MUST cite this contract rather than restating a
  divergent condition set. Origin: the workflow-optimization layer (redundant
  audit re-execution removal), tightened to the 4-condition compound predicate.
  This skip is distinct from Implementation Kickoff Approval: skip-eligibility
  governs ONLY Phase 1 verdict re-execution — it NEVER auto-bypasses the
  plan-to-implement human gate (the mandatory blocking `AskUserQuestion` gate;
  see `.claude/rules/moai/workflow/orchestration-mode-selection.md` header for
  the Implementation Kickoff Approval mandatory-restoration policy).
- Concurrent plan-run pipeline (Route B only): the orchestrator MAY begin run-phase pre-flight
  (Section C of the manager-develop prompt) on a feature branch while the plan
  PR is still in CI/review, PROVIDED the SPEC plan-auditor verdict is already
  PASS and no manager-develop commit lands on the feature branch until the
  plan PR is in MERGED state. This overlap reduces serial CI wait.
  Route A has no plan PR to wait on, so this overlap does not apply — run-phase
  begins directly after the plan-phase push + plan-auditor PASS + Implementation
  Kickoff Approval.

## Phase 1: Plan Audit Gate

The Plan Audit Gate is a mandatory protocol executed at the start of every `/moai run` invocation,
before any implementation phase begins. The gate invokes the plan-auditor subagent to independently
review all SPEC plan artifacts. It prevents unreviewed or incomplete SPEC artifacts from entering
the implementation phase.

### Gate Entry Condition

- Triggered on every `/moai run <SPEC-ID>` invocation
- Applies to every `/moai run` invocation (workflows/run.md)
- Cannot be skipped by harness level — gate is never disabled, not even on `minimal`

### Depends_on Pre-flight Check

The Depends_on Pre-flight Check is the first sub-step of Phase 1, executed BEFORE the plan-auditor subagent invocation. It is NOT a separate Phase 2 — it extends Phase 1 as sub-step 0 (no phase inflation).

**Procedure:**
1. Load the SPEC's frontmatter `depends_on:` list (Optional field per `.claude/rules/moai/development/spec-frontmatter-schema.md` § Optional Fields).
2. Where `depends_on` is absent or empty, the pre-flight trivially PASSes and proceeds to the plan-auditor step.
3. Where `depends_on` lists one or more SPEC IDs, resolve each dependency's current `status:` frontmatter field by reading `.moai/specs/<dep-ID>/spec.md`.

**Fulfillment definition (strict):** dependency fulfillment is defined as the dependency SPEC's `status: completed` — all other 7 status values (draft, planned, in-progress, implemented, superseded, archived, rejected) are considered unfulfilled. The evaluation is deterministic per-status: no partial credit, no "near-completed" interpretation, no score-based bypass.

**Blocker on unfulfilled dependency (3-option):** When one or more `depends_on` entries are unfulfilled, the pre-flight SHALL NOT proceed to the plan-auditor step. The orchestrator SHALL surface a structured blocker via `AskUserQuestion` with three options:
- **wait** — abort run; re-invoke after deps complete
- **override** — proceed with `--ignore-deps` flag; logged to `.moai/logs/depends-on-override.log` (the override path MUST record the unfulfilled dependency IDs + override rationale in the log; a bare `--ignore-deps` without the logged rationale is prohibited)
- **abort** — cancel run

The `--ignore-deps` flag and `.moai/logs/depends-on-override.log` path are literal tokens. The pre-flight is orchestrator-side doctrine; Go implementation is deferred to a follow-up SPEC.

### Verdicts

| Verdict | Meaning | Action |
|---------|---------|--------|
| `PASS` | All must-pass criteria met | Persist to progress.md, proceed to Phase 1 |
| `FAIL` | One or more must-pass criteria failed | Block Phase 1, surface report, AskUserQuestion |
| `BYPASSED` | User passed `--skip-audit` or set `MOAI_SKIP_PLAN_AUDIT=1` | Record bypass in report, proceed |
| `INCONCLUSIVE` | Auditor timed out, errored, or returned malformed output | Block, AskUserQuestion (retry/proceed/abort) |

### Report Persistence

Two report streams coexist deliberately in `.moai/reports/plan-audit/`; they are distinct by design and mutually cross-referenced here and in `.claude/agents/moai/plan-auditor.md` § Output Format:

- **plan-phase review stream** — `{SPEC-ID}-review-{N}.md`, iteration-based. Written by the plan-auditor during plan-phase adversarial review; iteration `N` follows the plan-auditor Retry Loop Contract (max 3). Consumed by the plan workflow's assembly/annotation cycle.
- **run-gate stream** — `<SPEC-ID>-<YYYY-MM-DD>.md`, date-based. Written by the Phase 1 Plan Audit Gate (`internal/runtime/audit_report.go`). Every gate call persists a record here; multiple calls on the same day append to the same file. This date-file is the verdict **record surface** only — it is never the hash subject for skip-eligibility (see below).

Skip-eligibility inputs (normative, matching the Go implementation): (a) the "most recent plan-auditor verdict" the run-gate consults is the plan-phase review stream's **final-iteration verdict**; (b) the artifact-hash check recomputes and compares the **plan-artifact hash** — `internal/runtime/audit_cache.go` `ComputeHash` hashes the SPEC directory's plan artifacts (spec.md / plan.md / acceptance.md / tasks.md) as whitespace-normalized SHA-256, with cache key = (specID, planArtifactHash); (c) the run-gate stream's date-file records the verdict but is not hashed.

**Plan-artifact hash subject list (Go verbatim):** the 4-file hash subject set is `{spec.md, plan.md, acceptance.md, tasks.md}` — matching `internal/runtime/audit_cache.go` `planArtifactNames` verbatim. The `tasks.md` entry is a V3R4-era plan artifact name retained in the hash subject list for backward compatibility with grandfathered SPECs (V3R6 Tier L replaces it with design.md + research.md, which are NOT hash subjects). `design.md` and `research.md` are **manual-skip judgment inputs** — changes to them do NOT mechanically invalidate a cached skip verdict but MUST be considered by the orchestrator's manual skip decision alongside the 4-file hash.

**Amendment as cache-invalidating event:** when a SPEC is amended in-place per the `completed → in-progress (amendment)` transition (completed → in-progress, `## Amendments` HISTORY row added), the plan-artifact hash changes because `spec.md` is modified — this is a cache-invalidating event that invalidates any cached plan-auditor PASS verdict for the SPEC, forcing Phase 1 plan-audit re-execution on the next `/moai run`. During the amendment transition, the SPEC remains V3R6 modern era (subject to drift detection) because frontmatter status is `in-progress` (not `completed`), so the `internal/spec/audit.go` completed-no-drift predicate does not fire.

Reports in both streams are local artifacts (gitignored).

### Grace Window

7-day grace window after the previous merge: FAIL verdicts emit warnings only (FAIL_WARNED),
not blocking. After grace window expires, FAIL verdicts block Phase 1 unconditionally.
Grace window start: `.moai/state/audit-gate-merge-at.txt` (ISO-8601 timestamp).

Run to Sync:
- Trigger (Route A): run-phase commits pushed to `main` AND tests passing (green CI on `main`).
- Trigger (Route B): Run PR merged into main, tests passing.
- Action: Execute `/moai sync SPEC-XXX`. Route A runs directly on `main`. Route B runs on the same branch/location as run — in the SAME L2 worktree if L2 was used (do NOT create a new L2 worktree); otherwise on the same feature branch in main checkout.

Sync (close):
- Trigger (Route A): the single sync commit — carrying the `implemented → completed` transition (manager-docs) and populating `sync_commit_sha` in progress.md §E.4 — is pushed to `main`. This is the 3-phase close: there is NO separate Mx-phase commit (MX Tag validation is a sync sub-step). The SPEC is `completed` once this commit lands.
- Trigger (Route B): sync PR merged into main. The sync PR carries the same single sync commit (the `implemented → completed` transition + `sync_commit_sha` population).

Sync to Cleanup (Route B only):
- Trigger: Sync PR merged into main
- Pre-condition: BOTH run PR AND sync PR are in MERGED state (verify via `gh pr view <PR>`)
- Action (only if L2 worktree was created): `moai worktree done SPEC-XXX` (executed from host checkout, not from inside the worktree)
- See § SPEC Phase Discipline (Step 4). Route A has no PR and no worktree cleanup step.

## Agent Teams Variant — RETIRED

The MoAI Agent Teams static-orchestration layer is RETIRED. Mode 3 (`agent-team`)
of the Phase 4 catalog is a tombstone (`.claude/rules/moai/workflow/orchestration-mode-selection.md`
§C.1), and the `--team` / `--mode team` dispatch value emits `MODE_TEAM_UNAVAILABLE`
and falls back to sub-agent mode. The former team-mode plan/run/fix/review skill
files and the `workflow.yaml` team-config block were removed.

The practical multi-agent surface is covered without the static team layer:
- Multi-domain research/review → Mode 4 (parallel fan-out: 3-5 concurrent read-only `Agent()` in one turn).
- Coding-heavy implementation → Mode 5 (sequential sub-agent) per Anthropic's coding-task parallelism caveat.
- High-volume mechanical transformation → Mode 6 (workflow / dynamic-workflow fan-out).

The native Claude Code teammate runtime is UNAFFECTED: `moai cg` GLM teammate
panes, `worktree --team` P1-P4 launch, the `~/.claude/teams/` registry, and
`teammateMode` launcher handling remain supported (see
`.claude/rules/moai/core/glm-web-tooling.md` § CG Mode). Only MoAI's static
team-orchestration layer built on top of that runtime is retired.
