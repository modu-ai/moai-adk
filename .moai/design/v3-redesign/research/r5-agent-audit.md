# R5 — moai Agent 1:1 Audit

> Research team: R5
> Scope: 22 agents in `.claude/agents/moai/` (template and local are byte-identical)
> Date: 2026-04-23
> Target reference: `.claude/rules/moai/development/agent-authoring.md`, `.claude/rules/moai/core/agent-common-protocol.md`, `.claude/rules/moai/core/moai-constitution.md`

---

## Executive summary

- **Total agents audited**: 22 (template and local paths contain the same 22 files — `diff -r` returned zero changes)
- **Directory layout**: `.claude/agents/moai/` only — there is no root-level `.claude/agents/*.md` content; all live under the `moai/` namespace
- **Categories**: manager (8), expert (8), builder (3), evaluator (2), meta (1 researcher). Meta-helpers claimed by the task prompt (Explore, Plan, general-purpose, statusline-setup, claude-code-guide) are NOT present as agent definitions in this repo — they are Claude Code built-ins surfaced at runtime and should not be audited here.
- **Verdict distribution**:
  - KEEP: 13 — manager-spec, manager-strategy, manager-ddd, manager-tdd, manager-git, manager-quality, manager-docs, expert-backend, expert-frontend, expert-security, expert-devops, evaluator-active, plan-auditor
  - REFACTOR: 6 — manager-project, expert-debug, expert-testing, expert-performance, expert-refactoring, researcher
  - MERGE: 3 — builder-agent + builder-skill + builder-plugin → single `builder-platform`; expert-debug into manager-tdd/ddd protocols; expert-performance into expert-backend
  - RETIRE: 0 outright (researcher could be retired as a fallback if not actively invoked, but it is a legitimate niche)
  - SPLIT: 2 candidates — expert-frontend (UI/design vs frontend code), manager-project (initialization vs document maintenance)
- **Overlap clusters**:
  - `manager-quality` vs `evaluator-active` vs `plan-auditor` — three overlapping skeptical-review agents with similar language and similar HARD rules, but different artifacts (code vs SPEC) and different trigger phases (run vs plan); taxonomy is defensible, but user-facing description needs tighter disambiguation.
  - `manager-ddd` vs `manager-tdd` — nearly duplicate bodies (cycle names, LSP baseline, checkpoint, @MX, decision guide). Consolidation candidate.
  - `expert-debug` vs `manager-ddd`/`expert-refactoring` — debug agent explicitly says "delegate implementation to manager-ddd", so it is a 30%-code router with the rest of its body overlapping diagnostic work already inside manager-quality and expert-refactoring.
  - `expert-refactoring` vs `manager-ddd` — both cover AST-based cross-file transformation. manager-ddd owns behavior-preserving refactors with tests; expert-refactoring owns codemod with ast-grep. Boundary needs explicit rule.
  - `expert-testing` vs `manager-tdd` vs `expert-backend` test sections — three agents own test strategy at different granularities with no hard border.
  - `builder-agent` vs `builder-skill` vs `builder-plugin` — three near-identical frontmatters (sonnet, bypass, memory:user), same tools list, same workflow phases. Can be a single multi-tool builder.
- **Effort calibration drift rate**: 7 of 22 (≈32%) do not match the Opus 4.7 guidance in `.claude/rules/moai/core/moai-constitution.md` §"Opus 4.7 Prompt Philosophy"
- **Common Protocol violations**: 3 hard violations — (1) manager-project and manager-spec explicitly instruct the subagent to use AskUserQuestion, which the protocol forbids for subagents; (2) manager-strategy §Phase 0/0.75 HARD-mandates AskUserQuestion inside the subagent; (3) several agents include the literal string "AskUserQuestion" in their Delegation or Workflow sections. Every occurrence should be rewritten as "return a blocker report requesting the orchestrator to ask".

---

## Per-agent audit table

Dimension key (in order): FM (frontmatter completeness), TR (description triggers EN/KO/JA/ZH), RC (role clarity), SC (scope boundaries), TS (tool scope), MD (model calibration), EF (effort calibration), BQ (body quality), BL (body length / token economy), OV (overlap discipline), WF (worktree/background/permissionMode calibration). Scores 0-3 per dimension, total out of 33.

| # | Agent | Category | FM | TR | RC | SC | TS | MD | EF | BQ | BL | OV | WF | /33 | Verdict | Overlap? |
|---|-------|----------|----|----|----|----|----|----|----|----|----|----|----|-----|---------|----------|
| 1 | manager-spec | manager | 3 | 3 | 3 | 3 | 2 | 3 | 3 | 3 | 2 | 2 | 2 | 29 | KEEP | minor w/ plan-auditor (clear boundary) |
| 2 | manager-strategy | manager | 3 | 3 | 3 | 3 | 3 | 3 | 3 | 3 | 2 | 2 | 2 | 30 | KEEP | minor w/ manager-spec |
| 3 | manager-ddd | manager | 3 | 3 | 3 | 3 | 2 | 3 | 2 | 3 | 2 | 1 | 2 | 27 | KEEP | heavy w/ manager-tdd + expert-refactoring |
| 4 | manager-tdd | manager | 3 | 3 | 3 | 3 | 2 | 3 | 2 | 3 | 2 | 1 | 2 | 27 | KEEP | heavy w/ manager-ddd + expert-testing |
| 5 | manager-quality | manager | 3 | 3 | 3 | 3 | 3 | 3 | 3 | 3 | 3 | 2 | 3 | 32 | KEEP | partial w/ evaluator-active |
| 6 | manager-docs | manager | 3 | 3 | 3 | 3 | 2 | 2 | 2 | 2 | 3 | 3 | 3 | 29 | KEEP | none |
| 7 | manager-git | manager | 3 | 3 | 3 | 3 | 3 | 3 | 2 | 3 | 3 | 3 | 3 | 32 | KEEP | none |
| 8 | manager-project | manager | 2 | 3 | 2 | 2 | 2 | 3 | 1 | 2 | 3 | 2 | 2 | 24 | REFACTOR | overlap w/ manager-spec + builder-* |
| 9 | expert-backend | expert | 3 | 3 | 3 | 3 | 2 | 3 | 2 | 3 | 2 | 2 | 2 | 28 | KEEP | partial w/ expert-performance |
| 10 | expert-frontend | expert | 3 | 3 | 3 | 3 | 1 | 3 | 2 | 3 | 2 | 2 | 2 | 27 | KEEP | partial w/ expert-debug (Pencil scope) |
| 11 | expert-security | expert | 3 | 3 | 3 | 3 | 3 | 3 | 3 | 3 | 3 | 3 | 3 | 33 | KEEP | none |
| 12 | expert-devops | expert | 3 | 3 | 3 | 3 | 2 | 3 | 2 | 3 | 2 | 3 | 3 | 30 | KEEP | none |
| 13 | expert-performance | expert | 3 | 3 | 3 | 3 | 3 | 3 | 2 | 3 | 3 | 1 | 3 | 30 | REFACTOR | heavy w/ expert-backend + expert-frontend |
| 14 | expert-debug | expert | 3 | 3 | 2 | 2 | 3 | 3 | 2 | 2 | 3 | 1 | 3 | 27 | REFACTOR | heavy w/ manager-ddd + manager-quality |
| 15 | expert-testing | expert | 3 | 3 | 2 | 2 | 2 | 3 | 2 | 3 | 3 | 1 | 3 | 27 | REFACTOR | heavy w/ manager-tdd + manager-ddd |
| 16 | expert-refactoring | expert | 3 | 3 | 3 | 3 | 3 | 3 | 2 | 3 | 3 | 1 | 3 | 30 | REFACTOR | heavy w/ manager-ddd (IMPROVE phase) |
| 17 | builder-agent | builder | 3 | 3 | 3 | 3 | 3 | 3 | 2 | 3 | 2 | 2 | 3 | 30 | MERGE | w/ builder-skill + builder-plugin |
| 18 | builder-skill | builder | 3 | 3 | 3 | 3 | 3 | 3 | 2 | 3 | 3 | 2 | 3 | 31 | MERGE | w/ builder-agent + builder-plugin |
| 19 | builder-plugin | builder | 3 | 3 | 3 | 3 | 3 | 3 | 2 | 3 | 2 | 2 | 3 | 30 | MERGE | w/ builder-agent + builder-skill |
| 20 | evaluator-active | evaluator | 3 | 3 | 3 | 3 | 3 | 3 | 3 | 3 | 3 | 2 | 3 | 32 | KEEP | partial w/ manager-quality + plan-auditor |
| 21 | plan-auditor | evaluator | 3 | 3 | 3 | 3 | 3 | 2 | 3 | 3 | 1 | 3 | 3 | 30 | KEEP | partial w/ evaluator-active |
| 22 | researcher | meta | 2 | 3 | 2 | 2 | 2 | 3 | 1 | 2 | 3 | 3 | 2 | 25 | REFACTOR | low; niche role |

Mean score: 28.8/33 (87.2%). Median: 30/33.

---

## Detailed findings (non-KEEP verdicts)

### manager-project (REFACTOR)

File: `.claude/agents/moai/manager-project.md`

- [Frontmatter] No `effort` field declared despite being selected as Sonnet for structured interviews. Body has complexity tiers (Simple/Medium/Complex) but no reasoning budget annotation. Recommend `effort: high` for Medium/Complex initialization, `medium` default.
- [Role clarity] Runs 6 routing modes: `language_first_initialization`, `fresh_install`, `settings_modification`, `language_change`, `template_update_optimization`, `glm_configuration`. This is doing work that belongs to three agents: project init (this one), settings mutation (should live in the CLI `moai cc`/`moai glm`/`moai update` layer, not an agent), and template optimization (should route to builder).
- [Common Protocol violation] Lines 31-35 explicitly say "CANNOT use AskUserQuestion — all user choices must be pre-collected by the command" — this IS correct and shows awareness. But lines 131-134, 92 and 99 include the string "AskUserQuestion" as a normal workflow step. The agent is inconsistent about its own constraint.
- [Scope] Line 108 says "File creation restricted to `.moai/project/` directory only" but the body also covers `.moai/config/config.yaml` reading and GLM configuration — out of stated scope.
- **SPLIT recommendation**: Extract init-interview flows into an orchestrator-driven skill (not an agent); keep a thin agent that only writes `.moai/project/{product,structure,tech}.md`. The "settings_modification", "glm_configuration", and "template_update_optimization" modes are not agent work — they are CLI side-effects and belong in the `moai` binary.

### expert-debug (REFACTOR)

File: `.claude/agents/moai/expert-debug.md`

- [Role clarity] The agent's own §Core Responsibilities explicitly say "Delegate Implementation: All code modifications delegated to specialized agents". The body is ≈100 LOC and 70% is a delegation table. This is not an expert — it is an intake router that returns diagnostic summaries. Either:
  1. Convert to a diagnostic-only pass (no `Write`/`Edit` in tools) and rename to `triage-debug` or fold into `manager-quality` as a sub-mode, or
  2. Give it real fix authority (reproduction test + minimal fix) so it justifies `expert-*` naming.
- [Tools mismatch] `tools:` line 12 does NOT include `Write` or `Edit` — already read-only — but `hooks:` includes `PostToolUse` with matcher `Write|Edit`. Since the agent has no Write/Edit tools, this hook never fires. Dead configuration.
- [Overlap] Delegation table (L89-93) duplicates manager-quality, manager-ddd, manager-git routing.
- **MERGE recommendation**: Fold expert-debug into a new "diagnostic" mode of `manager-quality` (which also has the skeptical evaluator stance). The `PostToolUse` hook for debug-verification action can be preserved as a `quality-debug` hook action.

### expert-testing (REFACTOR)

File: `.claude/agents/moai/expert-testing.md`

- [Role clarity] Body covers "test strategy design" and "E2E/integration/load testing". But:
  - Load testing is already IN SCOPE of `expert-performance` (line 42: "Load testing with k6, Locust, Apache JMeter").
  - Unit testing is OUT OF SCOPE (L55: "Unit test implementation (manager-ddd)").
  - E2E and integration is what's left — a thin slice.
- [Overlap] `expert-testing` §Delegation Protocol explicitly hands off to manager-ddd, expert-performance, expert-security, expert-backend — every actual testing activity is delegated elsewhere. Agent mostly delivers a strategy doc, which is orchestrator-level work.
- [Tools] Includes `mcp__claude-in-chrome__*` but the Playwright-like tooling is part of `moai:e2e` skill. No reason this agent should own browser MCP when `mcp__chrome-devtools__*` is also available but not listed.
- **MERGE recommendation**: Merge into `expert-performance` and manager-tdd. E2E/integration strategy sits with manager-tdd (test-first thinking). Load-test execution sits with expert-performance. expert-testing as a pure strategy doc writer is redundant with manager-strategy's test-plan role.

### expert-performance (REFACTOR)

File: `.claude/agents/moai/expert-performance.md`

- [Frontmatter] No `effort` field. Sonnet agent doing profiling + capacity planning reasoning — should be `effort: high`.
- [Scope] §OUT OF SCOPE says "Actual implementation of optimizations (delegate to expert-backend/expert-frontend)" — this makes the agent a pure recommender, same failure pattern as expert-debug/expert-testing.
- [Tool] Does NOT include `Write` or `Edit` (correct given read-only stance), but §Step 5 says "Create `.moai/docs/performance-analysis-{SPEC-ID}.md`" — inconsistent. Either grant `Write` scoped to `.moai/docs/` or remove the write instruction from the workflow.
- **REFACTOR recommendation**: Either grant `Write` scoped to `.moai/docs/` and keep as advisory-writer, or eliminate and fold into expert-backend + expert-frontend as a `--deepthink performance` mode.

### expert-refactoring (REFACTOR)

File: `.claude/agents/moai/expert-refactoring.md`

- [Role clarity] Good — AST-based transformations, codemods, bulk renames is a distinct value proposition over manager-ddd.
- [Overlap] manager-ddd §STEP 4 IMPROVE phase also uses AST-grep for multi-file transforms. The HARD rule at line 101 "Run tests after every refactoring" is identical to manager-ddd's PRESERVE contract. Boundary needs explicit rule: expert-refactoring = one-shot codemods without behavior contract; manager-ddd = multi-cycle refactors WITH characterization tests.
- [Effort] Has `effort: high` — correct for Opus 4.7 on `sonnet`? `effort` on sonnet is valid but non-idiomatic; Anthropic guidance says high/xhigh/max require Opus 4.7. This field is silently ignored on Sonnet — neither harmful nor useful.
- [Body length] 103 LOC — lean. Good.
- **REFACTOR recommendation**: Add explicit boundary rule: "This agent runs single-pass codemods without a test harness. For multi-cycle behavior-preserving refactors, use manager-ddd."

### researcher (REFACTOR)

File: `.claude/agents/moai/researcher.md`

- [Frontmatter] No `effort` field despite `model: opus`. An experimentation loop on Opus 4.7 should use `effort: xhigh` per constitution §Opus 4.7 guidance.
- [Role clarity] "Active self-research agent that optimizes moai-adk components... iterative experimentation with binary eval criteria" — niche and experimental. Is it actively invoked? Grep shows no `/moai` command invokes it; only the `moai-workflow-research` skill references it.
- [Tools] `Read, Write, Edit, Grep, Glob, Bash` — missing `TodoWrite` for experiment logging (L50 says "Log every experiment") and missing `Agent` to spawn sub-researchers for parallel experiments.
- [permissionMode] `acceptEdits` — correct for write-mode agent, but should pair with `isolation: worktree` per `.claude/rules/moai/workflow/worktree-integration.md` HARD rule for cross-file write agents. Missing.
- [Body length] 58 LOC — too terse. The "Workflow" section has 5 steps but lacks evaluator profile binding, stagnation thresholds, rollback protocol. Compare to manager-ddd (180 LOC).
- **REFACTOR recommendation**: Add `effort: xhigh`, `isolation: worktree`, expand §Rules with FrozenGuard binding, staleness_window, rollback protocol; cross-link to `.moai/research/evolution-log.md` (constitution §6). Or, if not actively used, RETIRE and fold into `moai-workflow-research` skill.

### builder-agent / builder-skill / builder-plugin (MERGE)

Files: `builder-agent.md`, `builder-skill.md`, `builder-plugin.md`

All three have near-identical frontmatter:
- `model: sonnet`, `permissionMode: bypassPermissions`, `memory: user`
- Same tool list (all 3): `Read, Write, Edit, Grep, Glob, WebFetch, WebSearch, Bash, TodoWrite, Agent, Skill, mcp__sequential-thinking__*, mcp__context7__*`
- Same skill imports: `moai-foundation-cc`, `moai-foundation-core`, `moai-workflow-project|moai-workflow-templates`

Body structure is identical too: Phase 1 Requirements Analysis → Phase 2 Research (Context7) → Phase 3 Architecture → Phase 4 Implementation → Phase 5 Validation.

Each agent differs only in the artifact: agent.md, SKILL.md, plugin manifest. Three agents for three file templates is over-decomposition.

- **MERGE recommendation**: Collapse to a single `builder-platform` agent that takes `artifact_type: agent|skill|plugin|command|hook|mcp-server|lsp-server` as its first parameter. Body becomes a single phased workflow with artifact-specific templates in a supporting skill (`moai-foundation-cc` already has the blueprints). Benefit: single agent to maintain for Claude Code schema drift, ≈60% body size reduction.

---

## Role taxonomy audit

### Manager agents (8) — taxonomy

```
manager-spec     | SPEC authoring          | KEEP
manager-strategy | Implementation planning | KEEP
manager-ddd      | Legacy refactor cycle   | KEEP (tight w/ TDD)
manager-tdd      | Greenfield cycle        | KEEP (tight w/ DDD)
manager-quality  | TRUST 5 verification    | KEEP (tight w/ evaluator)
manager-docs     | Documentation           | KEEP
manager-git      | Git workflow            | KEEP
manager-project  | Project init + config   | REFACTOR (split)
```

- Phantom boundaries: manager-ddd vs manager-tdd. Body overlap is ≈60% (LSP baseline, checkpoint, @MX, decision guide, steps 1-5). Real difference is 3 phase names (ANALYZE-PRESERVE-IMPROVE vs RED-GREEN-REFACTOR) and one decision heuristic (>10% coverage → TDD). Candidate for a single `manager-cycle` with `cycle_type: ddd|tdd` parameter.
- Phantom boundary 2: manager-quality vs evaluator-active. Both have "Skeptical Evaluation Mandate" as identical HARD rule block (same 6 bullets). Difference: quality runs during manager-ddd/tdd handoff; evaluator runs at harness=thorough Phase 2.8a. Defensible split, but HARD rule blocks should be de-duplicated into `agent-common-protocol.md`.
- Missing role: no `manager-design` agent even though `.claude/rules/moai/design/constitution.md` v3.3.0 describes a design pipeline. Currently `/moai design` routes directly to `moai-domain-copywriting` + `moai-domain-brand-design` skills with no coordinating agent. In v3, if design is a first-class flow, it likely needs one.

### Expert agents (8) — taxonomy

```
expert-backend     | Server + DB              | KEEP
expert-frontend    | UI + design + Pencil     | KEEP (consider SPLIT)
expert-security    | OWASP + threat modeling  | KEEP (exemplar)
expert-devops      | CI/CD + IaC              | KEEP
expert-performance | Profiling + load         | REFACTOR (advisor vs doer)
expert-debug       | Diagnosis router         | REFACTOR (router, not expert)
expert-testing     | Test strategy            | REFACTOR (strategy only)
expert-refactoring | AST codemods             | REFACTOR (boundary w/ DDD)
```

- Pattern: 4 of 8 experts are pure advisors that delegate all implementation back to the manager tier. This is over-decomposition. Advisors should be merged into the manager-strategy "expert consultation" flow, with the agent definitions retained only when they own DO authority over files (backend, frontend, security scan execution, devops CI config writing).
- expert-frontend SPLIT candidate: body mixes React/Vue code implementation (Phase 2/3) with Pencil MCP UI-design workflow (entire Pencil MCP section, 6 MCP tool families). Pencil is design-import, not frontend code. If v3 retains `moai-workflow-design-import` skill as the Path A handler, expert-frontend's Pencil section moves into a dedicated `expert-designer` or back into the `moai-domain-brand-design` skill.

### Builder agents (3) — taxonomy

Already discussed → MERGE to one `builder-platform`.

### Evaluator agents (2) — taxonomy

```
evaluator-active | Code-phase adversarial review | KEEP
plan-auditor     | SPEC-phase adversarial review | KEEP
```

Clean split by phase (code vs plan). Both have identical "Skeptical Evaluator Mandate" — the block should be extracted into `agent-common-protocol.md` §"Skeptical Evaluation Stance" and referenced. No action needed beyond DRY.

### Meta / researcher (1) — taxonomy

```
researcher | Self-research binary-eval experimenter | REFACTOR (or RETIRE if unused)
```

---

## Effort-level calibration matrix

Constitution (`.claude/rules/moai/core/moai-constitution.md` §Opus 4.7 Prompt Philosophy) prescribes:
- Reasoning-intensive (manager-spec, manager-strategy, plan-auditor, evaluator-active, expert-security, expert-refactoring) → `effort: xhigh` or `high`
- Implementation (expert-backend, expert-frontend, builder-*) → `effort: high` (default for Opus 4.7)
- Speed-critical (manager-git, Explore) → `effort: medium`

| # | Agent | Current effort | Task type | Recommended | Rationale |
|---|-------|----------------|-----------|-------------|-----------|
| 1 | manager-spec | **xhigh** | EARS + reasoning-heavy | xhigh | Correct |
| 2 | manager-strategy | **xhigh** | Architecture + Philosopher | xhigh | Correct |
| 3 | manager-ddd | (unset → inherit) | Multi-cycle refactor | **high** | Implementation-heavy; Opus 4.7 default guidance |
| 4 | manager-tdd | (unset → inherit) | RED-GREEN-REFACTOR cycles | **high** | Same as DDD |
| 5 | manager-quality | (unset → inherit) | TRUST 5 skeptical review | **high** | Reasoning-moderate |
| 6 | manager-docs | (unset → inherit) | Markdown + Nextra generation | **medium** | Template-driven, low reasoning |
| 7 | manager-git | (unset → inherit) | Git commands | **medium** | Speed-critical per constitution |
| 8 | manager-project | (unset → inherit) | Interview + document gen | **medium** | Per-mode; medium average |
| 9 | expert-backend | (unset → inherit) | API + DB code | **high** | Opus 4.7 impl default |
| 10 | expert-frontend | (unset → inherit) | UI + Pencil | **high** | Opus 4.7 impl default |
| 11 | expert-security | **high** | OWASP threat modeling | **xhigh** | Reasoning-intensive per constitution |
| 12 | expert-devops | (unset → inherit) | CI/CD YAML | **medium** | Config-heavy, low reasoning |
| 13 | expert-performance | (unset → inherit) | Profiling analysis | **high** | Analysis-moderate |
| 14 | expert-debug | (unset → inherit) | Diagnosis | **medium** | Router-level |
| 15 | expert-testing | (unset → inherit) | Strategy doc | **medium** | Doc writer |
| 16 | expert-refactoring | **high** | AST codemods | **xhigh** | Reasoning-intensive per constitution |
| 17 | builder-agent | (unset → inherit) | Template rendering | **medium** | Mechanical |
| 18 | builder-skill | (unset → inherit) | Template rendering | **medium** | Mechanical |
| 19 | builder-plugin | (unset → inherit) | Template rendering | **medium** | Mechanical |
| 20 | evaluator-active | **high** | Skeptical review + test | **xhigh** | Constitution explicitly lists evaluator-active for xhigh |
| 21 | plan-auditor | **high** | Adversarial SPEC review | **xhigh** | Constitution explicitly lists plan-auditor for xhigh |
| 22 | researcher | (unset → inherit) | Binary-eval experimentation | **xhigh** | Opus + iteration loop |

**Drift summary**:
- Explicit drift (declared value != recommended): 3 agents — expert-security (high→xhigh), evaluator-active (high→xhigh), plan-auditor (high→xhigh). The constitution names all three as xhigh agents; current declarations are one notch low.
- Implicit drift (missing `effort` field where guidance prescribes one): 19 of 22 agents do not declare `effort`. Silent inheritance means session default applies, which is `medium` per settings.json in v2.1.111+. For Opus 4.7 implementation agents this under-invokes reasoning.
- Overall drift rate: 7/22 (≈32%) deviate from constitution. Bigger concern is that 19/22 have no declared value at all — this is a systemic omission, not seven isolated mistakes.

**Recommended remediation**: Populate the `effort` field on every agent in v3 per the matrix above. Document the matrix in `.claude/rules/moai/development/agent-authoring.md` as a Required Fields table.

---

## Agent Common Protocol compliance

`agent-common-protocol.md` lists 7 HARD rules: (1) subagents MUST NOT prompt user, (2) language handling, (3) output format Markdown, (4) MCP fallback, (5) CLAUDE.md reference, (6) background write restriction, (7) time estimation prohibition.

| Agent | Rule 1 (no AskUserQuestion) | Rule 2 (lang) | Rule 3 (MD output) | Rule 4 (MCP fallback) | Rule 6 (bg write) | Rule 7 (no time) |
|-------|---|---|---|---|---|---|
| manager-spec | **VIOLATION** L134 | ok | ok | n/a | n/a | ok |
| manager-strategy | **VIOLATION** L58-72 | ok | ok | n/a | n/a | ok |
| manager-ddd | ok | ok | ok | n/a | n/a | ok |
| manager-tdd | ok | ok | ok | n/a | n/a | ok |
| manager-quality | ok | ok | ok | n/a | n/a | ok |
| manager-docs | ok | ok | ok | n/a | n/a | ok |
| manager-git | ok | ok | ok | n/a | n/a | ok |
| manager-project | **VIOLATION** L92, L99, L131-134 (inconsistent self-awareness) | ok | ok | n/a | n/a | ok (priority labels) |
| expert-backend | **VIOLATION** L66 "Use AskUserQuestion" | ok | ok | n/a | n/a | ok |
| expert-frontend | **VIOLATION** L66, L94 | ok | ok | n/a | n/a | ok |
| expert-security | ok | ok | ok | n/a | n/a | ok |
| expert-devops | **VIOLATION** L63 "Use AskUserQuestion" | ok | ok | n/a | n/a | ok |
| expert-performance | ok | ok | ok | n/a | n/a | ok (priority labels L86) |
| expert-debug | ok | ok | ok | n/a | n/a | ok |
| expert-testing | ok | ok | ok | n/a | n/a | ok (phases L95) |
| expert-refactoring | ok | ok | ok | n/a | n/a | ok |
| builder-agent | **VIOLATION** L61 "Use AskUserQuestion" | ok | ok | n/a | n/a | ok |
| builder-skill | **VIOLATION** L61 "Use AskUserQuestion" | ok | ok | n/a | n/a | ok |
| builder-plugin | **VIOLATION** L82 "Use AskUserQuestion" | ok | ok | n/a | n/a | ok |
| evaluator-active | ok | ok | ok | n/a | n/a | ok |
| plan-auditor | ok | ok | ok | n/a | n/a | ok |
| researcher | ok | ok | ok | n/a | n/a | ok |

**Summary**: 9 of 22 agents embed literal "AskUserQuestion" instructions in their body, violating `agent-common-protocol.md` Rule 1 ([HARD] Subagents MUST NOT prompt the user). This is the single largest protocol violation cluster.

Remediation pattern — every "Use AskUserQuestion to clarify X" line should be rewritten as:

> "If X is ambiguous, return a structured 'missing inputs' report to the orchestrator specifying which option set the user must choose. The orchestrator will ask via AskUserQuestion and re-invoke with the chosen value."

No other protocol rules are violated. Rule 7 (no time estimates) is well-followed — every agent uses phase ordering or priority labels.

---

## Language-trigger coverage

Constitution requires EN/KO/JA/ZH trigger keyword sets for MUST INVOKE auto-delegation.

| Agent | EN | KO | JA | ZH | Completeness |
|-------|----|----|----|----|--------------|
| manager-spec | 7 | 7 | 6 | 6 | full |
| manager-strategy | 5 | 5 | 4 | 4 | full |
| manager-ddd | 6 | 6 | 6 | 6 | full |
| manager-tdd | 7 | 7 | 7 | 7 | full |
| manager-quality | 7 | 7 | 6 | 6 | full |
| manager-docs | 7 | 7 | 6 | 6 | full |
| manager-git | 14 | 12 | 9 | 9 | full |
| manager-project | 6 | 6 | 5 | 5 | full |
| expert-backend | 24 | 23 | 18 | 18 | full |
| expert-frontend | 16 | 15 | 14 | 13 | full |
| expert-security | 9 | 9 | 8 | 8 | full |
| expert-devops | 8 | 8 | 7 | 7 | full |
| expert-performance | 8 | 8 | 7 | 7 | full |
| expert-debug | 8 | 8 | 7 | 7 | full |
| expert-testing | 7 | 7 | 7 | 7 | full |
| expert-refactoring | 10 | 9 | 9 | 9 | full |
| builder-agent | 6 | 6 | 4 | 5 | full |
| builder-skill | 5 | 5 | 5 | 5 | full |
| builder-plugin | 9 | 8 | 7 | 7 | full |
| evaluator-active | 6 | 6 | 6 | 6 | full |
| plan-auditor | 8 | 8 | 8 | 8 | full |
| researcher | 6 | 6 | 6 | 6 | full |

**All 22 agents have complete 4-language trigger coverage.** No missing language blocks. This is the strongest dimension in the audit.

Minor quality issues:
- expert-backend list is 24 EN tokens (duplicates: "Oracle" appears twice in EN and in KO translation "오라클, Oracle"). Over-enumeration dilutes weighted routing signal.
- manager-git trigger set is 14 items and includes "checkout", "rebase", "stash" — arguably too granular; may match on any git-adjacent mention.

---

## Worktree / background / permissionMode correctness

Per `.claude/rules/moai/workflow/worktree-integration.md` HARD rules:
- Implementation agents making cross-file writes SHOULD use `isolation: worktree`
- Read-only agents MUST NOT use `isolation: worktree`
- Background agents MUST NOT perform Write/Edit

| Agent | permissionMode | isolation | background | Correct? |
|-------|----------------|-----------|------------|----------|
| manager-spec | bypassPermissions | none | n/a | ok |
| manager-strategy | plan | none | n/a | ok (read-only) |
| manager-ddd | bypassPermissions | **none** | n/a | SHOULD add `isolation: worktree` for cross-file refactors |
| manager-tdd | bypassPermissions | **none** | n/a | SHOULD add `isolation: worktree` |
| manager-quality | plan | none | n/a | ok (read-only) |
| manager-docs | bypassPermissions | none | n/a | ok (scoped to docs) |
| manager-git | bypassPermissions | none | n/a | ok (operates on main worktree by design) |
| manager-project | bypassPermissions | none | n/a | ok (scoped to `.moai/project/`) |
| expert-backend | bypassPermissions | **none** | n/a | SHOULD add `isolation: worktree` for cross-file writes |
| expert-frontend | bypassPermissions | **none** | n/a | SHOULD add `isolation: worktree` |
| expert-security | bypassPermissions | none | n/a | OK (analysis; if writing fixes, worktree) |
| expert-devops | bypassPermissions | none | n/a | ok (config files, scoped) |
| expert-performance | bypassPermissions | none | n/a | ok (no Write in tools; read-only) |
| expert-debug | bypassPermissions | none | n/a | ok (no Write in tools) |
| expert-testing | bypassPermissions | none | n/a | SHOULD add if writing test files |
| expert-refactoring | bypassPermissions | **none** | n/a | MUST add `isolation: worktree` — cross-file codemods |
| builder-agent | bypassPermissions | none | n/a | ok (single-file writes) |
| builder-skill | bypassPermissions | none | n/a | ok |
| builder-plugin | bypassPermissions | none | n/a | ok (but multi-file; consider worktree) |
| evaluator-active | plan | none | n/a | ok (read-only) |
| plan-auditor | default | none | n/a | borderline — writes report under `.moai/reports/`; default works but acceptEdits or bypassPermissions for that path would be cleaner |
| researcher | acceptEdits | **none** | n/a | MUST add `isolation: worktree` per its own L49 "All experiments in worktree isolation when possible" |

**Worktree gaps**: 6 cross-file write agents (manager-ddd, manager-tdd, expert-backend, expert-frontend, expert-refactoring, researcher) lack `isolation: worktree`. `.claude/rules/moai/workflow/worktree-integration.md` §"HARD Rules" says this is SHOULD for one-shot sub-agents making cross-file changes. In v3 this should be upgraded to MUST for implementation agents that touch 3+ files per invocation.

---

## Tool scope audit (least-privilege review)

| Agent | Over-permissioned? | Under-permissioned? |
|-------|-------------------|---------------------|
| manager-spec | MultiEdit valid; mcp__context7__* justified | none |
| manager-strategy | none | none (plan mode sufficient) |
| manager-ddd | Skill included but all skills already preloaded via `skills:` — redundant | none |
| manager-tdd | same as DDD | none |
| manager-quality | none | none |
| manager-docs | WebFetch, WebSearch — only Context7 docs needed in practice | none |
| manager-git | lean tool set; good | none |
| manager-project | WebFetch/WebSearch not listed but Context7 present — inconsistent | WebFetch for competitor research (L89 says "Context7 competitor research") |
| expert-backend | 10+ tools — large surface | none |
| expert-frontend | **17 MCP tools** (pencil + chrome-in-chrome + context7) — largest MCP footprint in repo | none |
| expert-security | Agent tool — unusual for a subagent (can't spawn). Should remove | none |
| expert-devops | mcp__github__create-or-update-file — GitHub MCP write without a gh CLI equivalent; can be replaced by Bash+gh | none |
| expert-performance | Read-only tool list — good | none |
| expert-debug | Read-only — good (contradicts hooks on Write) | none |
| expert-testing | mcp__claude-in-chrome__* — single browser MCP; why not chrome-devtools? | none |
| expert-refactoring | Lean — good | none |
| builder-agent | Agent tool — subagent can't spawn. Remove. | none |
| builder-skill | Agent tool — same | none |
| builder-plugin | Agent tool — same | none |
| evaluator-active | Lean — good | none |
| plan-auditor | Write + Edit for writing audit report to `.moai/reports/plan-audit/` — justified | none |
| researcher | Lean | TodoWrite for experiment log, Agent for parallel experiments |

**Key finding**: 4 agents (expert-security, builder-agent, builder-skill, builder-plugin) declare the `Agent` tool. Per agent-authoring.md, "Sub-agents cannot spawn other sub-agents (Claude Code limitation)". These 4 declarations are dead configuration.

---

## Recommended v3 agent inventory

Target: 18-20 agents (down from 22). Core principle: each agent must own either (a) a unique phase, (b) a unique file-write authority, or (c) a unique reasoning stance. Pure advisors/routers are collapsed into the nearest manager.

### Managers (7)

1. `manager-spec` — SPEC authoring (unchanged)
2. `manager-strategy` — Implementation planning (unchanged, drop "AskUserQuestion" lines)
3. `manager-cycle` — **NEW**: unifies manager-ddd + manager-tdd with `cycle_type: ddd|tdd` parameter. Body: shared 60% content, parameterized phase names.
4. `manager-quality` — TRUST 5 verification (unchanged)
5. `manager-docs` — Documentation (unchanged)
6. `manager-git` — Git workflow (unchanged)
7. `manager-project` — **REFACTORED**: scope limited to `.moai/project/` document generation. Settings-modification/glm-configuration modes moved to CLI.

### Experts (6)

8. `expert-backend` — unchanged (add `isolation: worktree`)
9. `expert-frontend` — unchanged (add `isolation: worktree`; optionally split Pencil scope into new expert-designer if v3 design flow is first-class)
10. `expert-security` — unchanged (effort→xhigh; drop `Agent` tool)
11. `expert-devops` — unchanged
12. `expert-refactoring` — unchanged (add `isolation: worktree`; effort→xhigh; document boundary vs manager-cycle IMPROVE phase)
13. `expert-performance` — REFACTORED: grants `Write` for `.moai/docs/` only; owns actual profiling execution via `go tool pprof`/`k6`; replaces retired expert-testing load-test portion

### Builders (1)

14. `builder-platform` — **MERGED** from builder-agent + builder-skill + builder-plugin. Single agent accepting `artifact: agent|skill|plugin|command|hook|mcp|lsp` parameter.

### Evaluators (2)

15. `evaluator-active` — unchanged (effort→xhigh; drop duplicate Skeptical block, reference common protocol)
16. `plan-auditor` — unchanged (effort→xhigh)

### Meta (1, optional)

17. `researcher` — KEEP-or-RETIRE decision. If retained: add `effort: xhigh`, `isolation: worktree`, TodoWrite tool, expand body with FrozenGuard binding. If retired: fold into `moai-workflow-research` skill as runbook.

### Retired / merged (4)

- `expert-debug` → merged into `manager-quality` as diagnostic sub-mode
- `expert-testing` → strategy role merged into `manager-cycle`; load-test role into `expert-performance`; E2E execution into `/moai e2e` command
- `builder-agent`, `builder-skill`, `builder-plugin` → merged into `builder-platform`

**Net change**: 22 → 17 agents (−5, −23%). Preserves all unique authorities; eliminates advisor-only agents and triple-builder duplication.

---

## Additional systemic findings

1. **`memory:` field usage**: 18 of 22 agents have `memory: project`. 3 builders have `memory: user`. plan-auditor has no memory field. Inconsistency: plan-auditor is most memory-worthy (audit patterns across SPECs) but has none declared.
2. **Hook action matrix** (`.claude/rules/moai/core/agent-hooks.md` L22-34): 10 agents declare hooks. Action `debug-verification` fires on Write/Edit but expert-debug has no Write/Edit tools — dead configuration. evaluator-active has SubagentStop but no PostToolUse even though it writes .moai/reports — missing hook.
3. **Skills injection parity**: Most agents preload `moai-foundation-core`. manager-project, plan-auditor, researcher do not. Consistency suggests every agent should preload `moai-foundation-core` and `moai-foundation-cc` where applicable.
4. **`--deepthink` handling**: Every description line says "`--deepthink flag: Activate Sequential Thinking MCP`". This is boilerplate repeated 22x — candidate for removal from description and move to a global skill behavior, or to `agent-common-protocol.md`.
5. **`mcp__context7__*` usage**: 16 of 22 agents preload Context7 MCP. Context7 is read-only library docs. Over-inclusion in agents that never look up library docs (manager-git, manager-quality) is avoidable.
6. **Body length distribution**: Min 58 LOC (researcher), Max 271 LOC (plan-auditor). Mean 140 LOC. Per Opus 4.7 "one-turn fully-loaded" guidance, 100-200 LOC is the sweet spot. 5 agents below 100 (manager-docs 108, expert-debug 107, expert-performance 111, researcher 58, evaluator-active 113), 2 agents above 200 (manager-ddd 163 — ok, plan-auditor 271 — borderline justified by adversarial rubric complexity).

---

## Sources: file paths read

- `/Users/goos/MoAI/moai-adk-go/.claude/agents/moai/builder-agent.md`
- `/Users/goos/MoAI/moai-adk-go/.claude/agents/moai/builder-plugin.md`
- `/Users/goos/MoAI/moai-adk-go/.claude/agents/moai/builder-skill.md`
- `/Users/goos/MoAI/moai-adk-go/.claude/agents/moai/evaluator-active.md`
- `/Users/goos/MoAI/moai-adk-go/.claude/agents/moai/expert-backend.md`
- `/Users/goos/MoAI/moai-adk-go/.claude/agents/moai/expert-debug.md`
- `/Users/goos/MoAI/moai-adk-go/.claude/agents/moai/expert-devops.md`
- `/Users/goos/MoAI/moai-adk-go/.claude/agents/moai/expert-frontend.md`
- `/Users/goos/MoAI/moai-adk-go/.claude/agents/moai/expert-performance.md`
- `/Users/goos/MoAI/moai-adk-go/.claude/agents/moai/expert-refactoring.md`
- `/Users/goos/MoAI/moai-adk-go/.claude/agents/moai/expert-security.md`
- `/Users/goos/MoAI/moai-adk-go/.claude/agents/moai/expert-testing.md`
- `/Users/goos/MoAI/moai-adk-go/.claude/agents/moai/manager-ddd.md`
- `/Users/goos/MoAI/moai-adk-go/.claude/agents/moai/manager-docs.md`
- `/Users/goos/MoAI/moai-adk-go/.claude/agents/moai/manager-git.md`
- `/Users/goos/MoAI/moai-adk-go/.claude/agents/moai/manager-project.md`
- `/Users/goos/MoAI/moai-adk-go/.claude/agents/moai/manager-quality.md`
- `/Users/goos/MoAI/moai-adk-go/.claude/agents/moai/manager-spec.md`
- `/Users/goos/MoAI/moai-adk-go/.claude/agents/moai/manager-strategy.md`
- `/Users/goos/MoAI/moai-adk-go/.claude/agents/moai/manager-tdd.md`
- `/Users/goos/MoAI/moai-adk-go/.claude/agents/moai/plan-auditor.md`
- `/Users/goos/MoAI/moai-adk-go/.claude/agents/moai/researcher.md`
- `/Users/goos/MoAI/moai-adk-go/.claude/rules/moai/development/agent-authoring.md`
- `/Users/goos/MoAI/moai-adk-go/.claude/rules/moai/core/agent-common-protocol.md`
- `/Users/goos/MoAI/moai-adk-go/.claude/rules/moai/core/agent-hooks.md` (loaded via system reminder)
- `/Users/goos/MoAI/moai-adk-go/.claude/rules/moai/core/moai-constitution.md` (loaded via system reminder)
- `/Users/goos/MoAI/moai-adk-go/.claude/rules/moai/development/model-policy.md` (loaded via system reminder)
- `/Users/goos/MoAI/moai-adk-go/.claude/rules/moai/workflow/worktree-integration.md` (loaded via system reminder)

Verified byte-identity: `diff -r .claude/agents/moai/ internal/template/templates/.claude/agents/moai/` → no differences; audit applies to both local and template trees.
