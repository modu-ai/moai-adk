# plan.md — SPEC-V3R6-HARNESS-V4-001 Implementation Plan

> **Tier L** implementation plan. 6 milestones, priority-ordered, no time estimates (per `sprint-round-naming.md` and `agent-common-protocol.md` § Time Estimation). Each milestone is a `manager-develop` delegation unit (cycle_type=tdd unless noted).

## §A. Context

This SPEC rebuilds the moai-adk harness subsystem on Claude Code dynamic-workflow + conditional-worktree-isolation primitives, absorbing the strengths of `revfactory/harness`'s 7-Phase workflow and removing the rest. The rebuild is grounded in three prior completed baselines:

1. **SPEC-V3R6-HARNESS-NAMESPACE-V2-001** (completed) — resolved `my-harness-*` → `harness-*` namespace doctrine-code drift. §24 namespace protection currently covers `.claude/skills/harness-*/` and `.claude/agents/harness/` but NOT `.claude/commands/harness/` or `.claude/workflows/harness-*.js` — those are the NEW surfaces this SPEC introduces.
2. **SPEC-V3R6-LIFECYCLE-REDESIGN-001** (plan-phase ready) — retired the separate Mx phase; 3-phase plan→run→sync is the modern lifecycle, MX Tag is cross-cutting.
3. **SPEC-V3R6-WORKFLOW-EFFORT-MAP-001** (completed) — dynamic-workflow `agent()` purpose-driven model/effort taxonomy that the v4 Builder's PLAN phase consumes directly.

## §B. Known Issues

- **BI-001**: `.claude/commands/` is currently flat (97/98/99 + `moai/` subdirectory). The `/harness:<name>` namespace requires a new `.claude/commands/harness/` subdirectory. Claude Code subdirectory-command resolution must be verified (does `commands/harness/foo.md` resolve to `/harness:foo`?).
- **BI-002**: `.claude/workflows/` exists but contains only `codemaps-extract.js`. The `harness-<name>-run.js` Runner Workflow pattern needs a precedent for harness-namespaced workflows (does `moai update` need new preserve logic for `.claude/workflows/harness-*.js`?).
- **BI-003**: The existing 4 specialists (`.claude/agents/harness/*-specialist.md`) predate v4 manifest format. Porting them must preserve their current behavior (regression suite must pass post-port).
- **BI-004**: `moai-meta-harness` SKILL.md contains the 7-Phase workflow text with a `@MX:NOTE [AUTO] V3R4 contract` annotation marking it preserved unchanged per SPEC-V3R4-HARNESS-001 §10. Converting it to a v4 redirect must honor or explicitly supersede that annotation.

## §C. Pre-flight (before M1)

- [ ] Verify Claude Code subdirectory-command resolution: does `.claude/commands/harness/dev.md` produce `/harness:dev`?
- [ ] Verify `moai update` preserve logic surface: enumerate the Go functions that classify user-owned paths (`isUserOwnedNamespace`, `isUserAreaPath` in `internal/cli/update*.go`) and confirm `.claude/commands/harness/` + `.claude/workflows/harness-*.js` are NOT yet protected.
- [ ] Verify dynamic-workflow runtime version requirement (Claude Code v2.1.154+) and document the fallback path if unavailable.

## §D. Constraints (restated from spec.md §D)

- C-HV4-001 simplest-solution-first; C-HV4-002 no recursive subagent spawning; C-HV4-003 workflow script determinism; C-HV4-004 AskUserQuestion mid-run prohibition; C-HV4-005 template neutrality §25; C-HV4-006 worktree L1 autonomy advisory.

## §E. Self-Verification (manager-develop §E binding)

Each milestone commit MUST carry the manager-develop §E self-verification matrix (E1-E7 per `.claude/rules/moai/development/manager-develop-prompt-template.md`): AC PASS/FAIL matrix, cross-platform build, coverage, subagent-boundary grep, lint, push state, and the residual-risk note. The dogfooding validation (REQ-HV4-012) additionally carries the 5-Section Evidence-Bearing Report Format.

## §F. Milestones

### M1 — `/moai:harness` NL-analysis entry + §24 namespace extension
**Priority**: P0 (entry point + namespace protection unblock downstream milestones).

Scope:
- Extend `moai update` preserve logic to protect `.claude/commands/harness/` and `.claude/workflows/harness-*.js` as user-owned (extend `isUserOwnedNamespace` / `isUserAreaPath` in `internal/cli/update*.go`).
- Author the `/moai:harness` NL-analysis entry: Context-First Discovery on natural-language request, harness `<name>` derivation (not user-supplied), explicit approval gate before Builder delegation.
- Update §24 doctrine text (`harness-namespace-doctrine.md`) to cover the two new user-owned surfaces.

Deliverables:
- `internal/cli/update_*.go` namespace protection extension + unit tests.
- `/moai:harness` command file (`internal/template/templates/.claude/commands/moai/harness.md`) — NL-analysis path.
- `harness-namespace-doctrine.md` §24 extension (commands/ + workflows/ rows).

AC bindings: AC-HV4-001a/b (NL analysis), AC-HV4-010a/b (namespace protection).

### M2 — Builder Workflow `harness-build.js` (4 signal-driven phases)
**Priority**: P0 (the core dynamic-workflow that generates harnesses).

Scope:
- Author `.claude/workflows/harness-build.js` as a deterministic dynamic-workflow script (script holds the plan).
- ANALYZE phase: Explore fan-out (read-only, main tree) — codebase + docs + existing harness/* + SPEC history; produce domain profile + task-pattern inventory.
- PLAN phase: single opus xhigh sub-agent — select/combine from 6 patterns, define specialist roles, map each to an execution primitive, decide per-specialist worktree isolation; produce draft manifest. User approval gate (AskUserQuestion) BEFORE GENERATE.
- GENERATE phase: dynamic-workflow fan-out — specialist agents, companion skills, Runner Workflow script, manifest.json, AND the `/harness:<name>` command file.
- ACTIVATE phase: sample-task dry-run + `/goal` autonomous convergence + with/without A/B; load-bearing minimum (ACTIVATE A/B skipped for simple tasks).

Deliverables:
- `.claude/workflows/harness-build.js` Builder Workflow script.
- 6-pattern catalog module (`harness-patterns.js` or equivalent) referenced by PLAN.
- PLAN-phase primitive-mapping logic (specialist → primitive).

AC bindings: AC-HV4-003a/b (4-phase Builder, ≥1 phase skipped under load-bearing-minimum), AC-HV4-004a/b (6-pattern dynamic selection ≥2 patterns for ≥2 different domain requests).

### M3 — manifest.json schema + Runner primitive-mapping engine
**Priority**: P0 (the Runner consumes the manifest verbatim).

Scope:
- Define the canonical `manifest.json` schema (REQ-HV4-006): name / domain / source_request / patterns / specialists[role,primitive,isolation,effort,model] / sprint_contract[dimensions,thresholds] / entry_command / runner_workflow.
- Author the Runner Workflow template `harness-<name>-run.js` (generated per harness by the Builder GENERATE phase). The Runner reads `manifest.json` primitive-mapping and dispatches sub-agent / dynamic-workflow / worktree / `/goal` / adversarial-fan-out per task signals.
- Sprint Contract (Generator-Evaluator separation) carried in manifest; evaluator conditional (skipped when task within model's solo range).

Deliverables:
- `manifest.json` JSON schema (documented in design.md §C).
- Runner Workflow generator (lives inside Builder GENERATE phase; emits one `harness-<name>-run.js` per harness).
- Runner primitive-dispatch engine (consumes specialist.primitive verbatim).

AC bindings: AC-HV4-005a/b (primitive-mapping consumed verbatim by Runner), AC-HV4-006a/b (manifest schema validation), AC-HV4-008a/b (Sprint Contract present, evaluator conditional).

### M4 — `/harness:<name>` dynamic command generation + lifecycle + orphan prevention
**Priority**: P1 (execution entry + lifecycle management).

Scope:
- Command generator: at Builder GENERATE time, emit `.claude/commands/harness/<name>.md` (thin wrapper → Runner Workflow) so `/harness:<name>` resolves to that harness's Runner.
- Lifecycle commands: `/moai:harness list` (enumerate harnesses by scanning `commands/harness/*.md` joined with manifests), `/moai:harness edit <name>` (open manifest + specialists for editing), `/moai:harness remove <name>` (atomic removal of command + workflow + specialists + skills + manifest).
- Orphan-command prevention: remove fails closed if any referenced artifact is missing; remove is atomic (all-or-nothing).

Deliverables:
- Command-file generator (Builder GENERATE emits `commands/harness/<name>.md`).
- `/moai:harness list|edit|remove` command handlers in `internal/cli/`.
- Orphan-prevention guard (atomicity check + fail-closed).

AC bindings: AC-HV4-002a/b (`/harness:<name>` auto-generation + invocation), AC-HV4-011a/b/c (list/edit/remove + orphan prevention).

### M5 — Conditional worktree-isolation sub-agent logic
**Priority**: P1 (sub-agent-granular isolation).

Scope:
- Per-specialist `isolation` field in manifest (`none` / `worktree`).
- PLAN phase decides per-specialist isolation based on conflict-risk (parallel file generation targeting same paths → worktree; read-only or sequential → none).
- GENERATE phase spawns `Agent(isolation:"worktree")` for `isolation:worktree` specialists, plain sub-agent for `isolation:none`. NO mandatory top-level worktree for Builder/Runner.
- Runner end-of-run emits worktree cleanup directive (L1 autonomous cleanup).

Deliverables:
- PLAN-phase isolation-decision logic (per specialist, not per Builder).
- GENERATE-phase conditional `Agent(isolation:"worktree")` spawn.
- Runner cleanup directive.

AC bindings: AC-HV4-007a/b (0 worktrees for read-only ANALYZE, ≥1 for conflict-prone GENERATE).

### M6 — Migrate existing 4 specialists + deprecate legacy 7-Phase + remove revfactory residuals
**Priority**: P2 (migration + cleanup; runs after v4 is proven on a new harness).

Scope:
- Port the 4 existing specialists (`.claude/agents/harness/{cli-template,quality,workflow,hook-ci}-specialist.md`) to v4 manifest format (each gains primitive / isolation / effort / model). Preserve behavior; Layer B regression suite must pass.
- Convert `moai-meta-harness` legacy 7-Phase path to a redirect: on invocation, surface deprecation notice + route to `/moai:harness` v4. Honor or explicitly supersede the `@MX:NOTE [AUTO] V3R4 contract` annotation.
- Remove revfactory 7-Phase residuals from v4 artifacts (grep "7-Phase", "Phase 7 LEARNING", "Skeleton", "Customization" → 0 matches).
- Dogfooding validation (REQ-HV4-012): build "moai-adk-dev" harness with v4, run real moai-adk development task, with/without A/B, disclose as author-measured.

Deliverables:
- 4 migrated specialists (v4 manifest format, behavior preserved).
- `moai-meta-harness` redirect (deprecation notice + v4 routing).
- Revfactory-residual grep = 0 verification.
- Dogfooding validation report (5-Section Evidence-Bearing Format).

AC bindings: AC-HV4-009a (revfactory residual grep = 0), AC-HV4-012a/b (dogfooding validation, A/B disclosed as author-measured), AC-HV4-013a/b (specialists migrated, legacy redirect live).

## §G. Anti-Patterns (this plan avoids)

- **AP-HV4-P001 — Mandatory top-level worktree**: forcing Builder or Runner into a top-level `Agent(isolation:"worktree")` regardless of phase risk. v4 makes worktree conditional and sub-agent-granular (REQ-HV4-007).
- **AP-HV4-P002 — Always-on evaluator**: running the Generator-Evaluator A/B on every task regardless of complexity. v4 makes the evaluator conditional (REQ-HV4-008, C-HV4-001 simplest-solution-first).
- **AP-HV4-P003 — Static 7-Phase pipeline**: running all 4 Builder phases unconditionally. v4 synthesizes phases from task signals (load-bearing minimum).
- **AP-HV4-P004 — Recursive subagent spawning**: specialist agents spawning further sub-agents. Prohibited by Anthropic subagent ceiling (C-HV4-002); hierarchical delegation is manifest-level only.
- **AP-HV4-P005 — Workflow script non-determinism**: `Date.now()` / `Math.random()` inside the script body breaking resume caching (C-HV4-003).

## §H. Cross-References

- `spec.md` — 13 REQs, 6 constraints, NFRs, success criteria.
- `acceptance.md` — AC-HV4-NNN traceable to REQs.
- `design.md` — v4 architecture, manifest schema (full), revfactory 7-Phase → v4 mapping.
- `research.md` — revfactory 7-Phase analysis, Anthropic GAN guidance, verbatim source citations.
- `.claude/rules/moai/workflow/dynamic-workflows.md` — dynamic-workflow primitive selection + determinism constraint + purpose-driven model/effort taxonomy.
- `.claude/rules/moai/development/sprint-round-naming.md` — Milestone (within-SPEC step) vs Round (SSE-stall split) vs Sprint (multi-SPEC container).
