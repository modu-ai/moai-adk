---
id: SPEC-V3R6-HARNESS-V4-001
title: "Harness v4: orchestrator-direct Builder + dynamic-workflow Runner rebuild (absorb revfactory strengths, remove 7-Phase)"
version: "0.3.0"
status: completed
created: 2026-06-19
updated: 2026-06-20
author: manager-spec
priority: P0
phase: "v3.0.0"
module: "internal/harness, internal/cli, .claude/commands/harness, .claude/workflows, .claude/agents/harness"
lifecycle: spec-anchored
era: V3R6
tier: L
tags: "harness, dynamic-workflow, orchestrator-direct, worktree, revfactory, meta-harness, namespace, sprint-contract, gan"
depends_on:
  - SPEC-V3R6-HARNESS-NAMESPACE-V2-001
related_specs:
  - SPEC-V3R6-LIFECYCLE-REDESIGN-001
  - SPEC-V3R6-WORKFLOW-EFFORT-MAP-001
---

# SPEC-V3R6-HARNESS-V4-001 — Harness v4 Dynamic-Workflow Rebuild

> **Era classification**: V3R6 modern era (3-phase plan→run→sync; MX Tag is a cross-cutting sync concern, NOT a separate phase — per `SPEC-V3R6-LIFECYCLE-REDESIGN-001`). This SPEC's 4-phase Builder (ANALYZE / PLAN / GENERATE / ACTIVATE) is an internal harness-construction pipeline that runs ENTIRELY within the plan/run lifecycle of this SPEC; it is NOT a SPEC-lifecycle phase and does NOT conflict with the 3-phase V3R6 lifecycle.

> **Relationship to revfactory/harness**: This SPEC **absorbs** the strengths of `github.com/revfactory/harness` (Apache-2.0) — domain-sentence → team-architecture auto-generation, the 6-patterns catalog, Progressive-Disclosure skill generation, Generator-Evaluator separation, and with/without A/B validation — and **removes** the rest (static fixed 7-Phase pipeline, hard Agent-Teams dependency, Skeleton↔Customization 2-phase split, LEARNING as a separate phase, revfactory-specific plugin assumptions, and Socratic-only discovery without codebase/doc ground-truth analysis). The rebuild is grounded on Claude Code dynamic-workflow + conditional-worktree-isolation primitives. Full upstream analysis and verbatim source citations live in `research.md`.

## §A. Problem Statement

moai-adk's harness subsystem today (the `moai-meta-harness` skill, adapted from revfactory/harness's 7-Phase workflow) carries six over-engineering liabilities for 2026:

1. **Static fixed 7-Phase pipeline.** The pipeline runs Phase 1→7 unconditionally regardless of task shape. Anthropic's harness-design guidance ("find the simplest solution possible, only increase complexity when needed") calls for phase synthesis from task signals, not a fixed sequence.
2. **Hard Agent-Teams dependency.** The 7-Phase workflow assumes Agent-Teams (3-5 teammate cap per Anthropic guidance) as its execution vehicle. For large parallel work this is too narrow (Claude Code dynamic-workflow supports 16 concurrent / 1000 total agents).
3. **Skeleton ↔ Customization 2-phase split.** Phases 4+5 (Skeleton generation, then Customization) split one logical fan-out into two sequential phases. A single dynamic-workflow fan-out collapses both.
4. **LEARNING as a separate phase.** Phase 7 (Iteration/LEARNING) is a separate phase, but learning is a cross-cutting concern that belongs in the Sprint Contract, not its own phase.
5. **No codebase/doc ground-truth analysis.** Phase 1 Discovery is Socratic-only — it interviews the user but never reads the actual codebase or docs. This produces harnesses that contradict existing conventions.
6. **revfactory-specific assumptions.** The 7-Phase workflow carries revfactory plugin/marketplace assumptions that do not map to moai-adk-native execution primitives (sub-agent / dynamic-workflow / worktree / `/goal`).

The user's directive (verbatim, natural language): "moai-adk 개발을 위한 하네스를 구축해줘" → the system analyzes the request, plans, generates all harness files, and runs via sub-agent / dynamic-workflow / agent-worktree / `/goal` — using Claude Code and moai-adk features to the fullest.

## §B. Scope

### §B.1 In Scope

1. **Harness subsystem redesign** — entry commands, Builder orchestrator-direct, Runner dynamic-workflow, manifest schema, namespace policy, lifecycle, migration of the 4 existing specialists.
2. **`/moai:harness` NL-analysis entry** — Context-First Discovery on natural-language request; harness `<name>` derived, not statically supplied.
3. **`/harness:<name>` execution command namespace** — one auto-generated command per harness (Claude Code subdirectory-command pattern `.claude/commands/harness/<name>.md`).
4. **Builder (orchestrator-direct)** — 4 signal-driven phases (ANALYZE fan-out / PLAN sub-agent / GENERATE fan-out / ACTIVATE `/goal`+A/B); load-bearing minimum (phases skipped for simple tasks). The Builder runs as orchestrator-direct processing (NOT a separate JS script); intermediate results held in the orchestrator session context. The Runner stays a dynamic-workflow script (REQ-HV4-005/006, design §F).
5. **6-pattern catalog** — Pipeline, Fan-out/Fan-in, Expert Pool, Producer-Reviewer, Supervisor, Hierarchical; selected/combined dynamically per task signals.
6. **Execution-primitive mapping** — each specialist mapped to sub-agent / dynamic-workflow / worktree / `/goal` / adversarial-fan-out in `manifest.json`; consumed verbatim by Runner.
7. **Conditional worktree isolation** — sub-agent-granular `Agent(isolation:"worktree")` only for conflict-prone parallel generation or risky changes; Builder/Runner otherwise main-tree.
8. **Generator-Evaluator separation + Sprint Contract** (Anthropic GAN-inspired) — in ACTIVATE and Runner; skeptical-evaluator tuning; evaluator conditional (skipped when task within model's solo range).
9. **`manifest.json` schema** — name / domain / source_request / patterns / specialists[role,primitive,isolation,effort,model] / sprint_contract / entry_command / runner_workflow.
10. **Namespace policy extension** — `harness-*` skills + `.claude/agents/harness/` + `.claude/commands/harness/` + `.claude/workflows/harness-*.js` are USER-OWNED, protected by `moai update`; extends §24 to cover `commands/` and `workflows/`.
11. **Harness lifecycle** — list (`/moai:harness list`), edit (`/moai:harness edit <name>`), remove (`/moai:harness remove <name>` with command/workflow cleanup); orphan-command prevention.
12. **Dogfooding validation** — build a "moai-adk-dev" harness with v4 and use it for real moai-adk development; with/without A/B vs the revfactory +60% claim (disclosed as author-measured, 3rd-party pending).
13. **Migration** — migrate the existing 4 specialists (cli-template / quality / workflow / hook-ci) to v4 manifest format; deprecate the `/moai:harness` legacy 7-Phase path (becomes a redirect to v4 Builder); remove revfactory residuals.

### §B.2 Out of Scope (Exclusions)

### Out of Scope — Retained 8-agent catalog

- This SPEC does NOT modify the 8 retained agents (manager-spec / manager-develop / manager-docs / manager-git / plan-auditor / sync-auditor / builder-harness / Explore). The harness specialists it generates live under the separate user-owned `.claude/agents/harness/` namespace and are NOT part of the retained catalog.

### Out of Scope — SPEC workflow itself

- This SPEC does NOT change the SPEC plan/run/sync lifecycle, GEARS notation, lint engine, frontmatter schema, or plan-auditor gate. Those are orthogonal. Harness v4 CONSUMES the SPEC workflow when it builds harnesses for SPEC-driven projects; it does not redefine it.

### Out of Scope — Learning-subsystem internals

- The learning subsystem (Meta-Harness outer-loop optimization of harness code itself, forward-linked to arxiv 2603.28052) is declared as a cross-cutting concern folded into the Sprint Contract, but its internal optimization algorithm is NOT implemented in this SPEC. A follow-up SPEC (candidate `SPEC-V3R6-HARNESS-LEARNING-LOOP-001`) will own the outer-loop implementation.

### Out of Scope — 88 pre-v3 SPEC re-authoring

- This SPEC does NOT re-author the 88 pre-v3 SPECs to match harness v4 conventions. Those remain readable under the 6-month EARS backward-compatibility window.

### Out of Scope — Downstream user-project migration helper

- User projects that already have harnesses built under the legacy 7-Phase path are handled by a separate migration SPEC if required. This SPEC declares the v4 contract; it does not auto-migrate external installs.

## §C. Requirements (GEARS notation)

> **Subject convention**: requirements use a generalized `<subject>` (the v4 harness, the Builder, the Runner, the orchestrator, the manifest, the command generator, the namespace protector). GEARS compound clauses (`Where` / `While` / `When`) are used where preconditions, states, or events apply.

### REQ-HV4-001 — Natural-language harness creation entry

**Where** the user issues `/moai:harness <natural-language request>`, the orchestrator shall execute Context-First Discovery on the request to extract domain, goal, constraints, and scope; **when** intent clarity is below 100%, the orchestrator shall conduct AskUserQuestion Socratic rounds (max 4 questions per round) until clarity reaches 100%; the orchestrator shall derive the harness `<name>` from the request (not require the user to supply it statically) and obtain explicit final approval before delegating to the Builder.

### REQ-HV4-002 — Self-describing harness execution command

**When** the Builder successfully completes the GENERATE phase, the command generator shall create a thin-wrapper command file at `.claude/commands/harness/<name>.md` that dispatches to that harness's Runner Workflow; the orchestrator shall expose the harness for execution via the `/harness:<name>` command; there shall be exactly one such command per harness, and the command namespace `.claude/commands/harness/` shall be the canonical location for all generated harness entry points.

### REQ-HV4-003 — Builder as orchestrator-direct processing with signal-driven phases

**Where** the `/moai:harness <request>` creation flow is invoked, the Builder shall run as orchestrator-direct processing (orchestrator-side logic, NOT a separate dynamic-workflow script and NOT a separate agent) with four phases — ANALYZE (orchestrator parallel `Agent(agentType:"Explore", effort:"low")` fan-out, read-only), PLAN (orchestrator spawns a single `Agent(model:"opus", effort:"xhigh")` sub-agent), GENERATE (orchestrator fan-out emits the 5 artifact types with conditional worktree isolation), ACTIVATE (orchestrator-direct dry-run + `/goal` autonomous convergence + with/without A/B); the orchestrator holds intermediate results in its own session context (the plan lives in Claude's context, NOT in a script); the Builder shall synthesize which phases to run from task signals (load-bearing minimum), and **when** a task is simple enough to be within the model's solo range, the Builder shall skip ACTIVATE's A/B evaluation rather than always run it. The Runner (design §F) is a separate dynamic-workflow script and is covered by REQ-HV4-005/006, not by this requirement.

### REQ-HV4-004 — 6-pattern catalog dynamically selected

**Where** the PLAN phase selects execution patterns, the Builder shall choose from the 6-pattern catalog (Pipeline, Fan-out/Fan-in, Expert Pool, Producer-Reviewer, Supervisor, Hierarchical Delegation); the Builder shall select and/or combine patterns dynamically based on task signals (parallelism, adversarial-verification need, supervision depth, expertise diversity), NOT apply a fixed pattern; the selected patterns shall be recorded in `manifest.json`.

### REQ-HV4-005 — Execution-primitive mapping per specialist

**When** the PLAN phase defines specialist roles, the Builder shall map each specialist to exactly one execution primitive from the set {sub-agent, dynamic-workflow, worktree, `/goal`, adversarial-fan-out}; the mapping shall be declared in `manifest.json` under each specialist's `primitive` field; the Runner Workflow shall consume each specialist's `primitive` verbatim and dispatch accordingly — the Runner shall NOT re-derive the primitive from heuristics at run time.

### REQ-HV4-006 — manifest.json canonical schema

The manifest generator shall emit `manifest.json` with the canonical schema: `name` (string), `domain` (string), `source_request` (the original natural-language request verbatim), `patterns` (array of pattern names from the 6-pattern catalog), `specialists` (array of objects each with `role` / `primitive` / `isolation` / `effort` / `model`), `sprint_contract` (object with `dimensions` array and `thresholds` map), `entry_command` (the `/harness:<name>` string), and `runner_workflow` (the `harness-<name>-run.js` filename); the manifest shall be machine-readable JSON and the single source of truth consumed by the Runner.

### REQ-HV4-007 — Conditional worktree isolation

**Where** a Builder or Runner phase runs parallel file generation that is conflict-prone OR makes risky changes, the orchestrator shall spawn `Agent(isolation:"worktree")` sub-agents for the conflicting units; **while** a phase is read-only analysis (ANALYZE fan-out) or main-tree orchestrator-direct (PLAN sub-agent), the orchestrator shall NOT create a worktree; there shall be NO mandatory top-level worktree for the Builder or Runner as a whole — worktree isolation is sub-agent-granular and conditional per specialist's `isolation` field.

### REQ-HV4-008 — Generator-Evaluator separation + Sprint Contract

**Where** the ACTIVATE phase or the Runner evaluates generated harness output, the system shall apply Generator-Evaluator separation (Anthropic GAN-inspired): the agent that generates output shall be distinct from the agent that evaluates it; the system shall carry a Sprint Contract (pre-coding agreement on graded dimensions and thresholds) declared in `manifest.json.sprint_contract`; **when** the task is within the model's solo reliable range, the Runner shall skip the evaluator (the evaluator earns its cost only when the task exceeds what the model does reliably solo); the evaluator, when invoked, shall be tuned as a skeptical adversarial reviewer (not a yes/no gate).

### REQ-HV4-009 — revfactory 7-Phase residuals removed

The v4 harness artifacts (Builder orchestrator-direct logic, Runner Workflow, manifest schema, generated specialists, generated skills) shall NOT contain any static fixed 7-Phase pipeline, shall NOT hard-depend on Agent-Teams (3-5 cap) as the execution vehicle, shall NOT split generation into Skeleton and Customization phases, and shall NOT carry a LEARNING separate phase; learning shall be a cross-cutting concern folded into the Sprint Contract. (Acceptance: grep for "7-Phase", "Phase 7 LEARNING", "Skeleton", "Customization" in newly generated v4 artifacts returns zero matches.)

### REQ-HV4-010 — Namespace policy extension (commands/ + workflows/)

The namespace policy shall classify `.claude/commands/harness/` and `.claude/workflows/harness-*.js` as USER-OWNED, protected from `moai update` deletion/modification, alongside the existing `.claude/skills/harness-*/` and `.claude/agents/harness/` protected surfaces; the `moai-harness-*` and `moai-meta-harness` prefixes shall remain template-managed builder artifacts (NOT user-owned); `moai update` shall back up and preserve the user-owned harness artifacts before any template sync, and shall NOT leak `harness-*` / `.claude/commands/harness/` / `.claude/workflows/harness-*.js` into `internal/template/templates/`.

### REQ-HV4-011 — Harness lifecycle + orphan-command prevention

**When** the user issues `/moai:harness list`, the orchestrator shall enumerate all harnesses (by scanning `.claude/commands/harness/*.md` and joining with their `manifest.json`); **when** the user issues `/moai:harness remove <name>`, the orchestrator shall remove the harness's command file, Runner Workflow, specialists, skills, AND manifest atomically, such that NO orphan command (a `/harness:<name>` whose Runner Workflow or manifest has been removed) can remain; the remove operation shall fail closed (refuse to remove) if any referenced artifact cannot be located, rather than leave a partial state.

### REQ-HV4-012 — Dogfooding validation

The project shall validate v4 by building a "moai-adk-dev" harness with the v4 Builder and using it for real moai-adk development tasks; the validation shall run a with/without A/B comparison (harness vs no-harness baseline) and disclose the result as author-measured (n-sample disclosed, 3rd-party replication pending), NOT as a verified third-party claim; the validation report shall be recorded as a research artifact.

### REQ-HV4-013 — Migration of existing 4 specialists + legacy redirect

The migration shall port the 4 existing specialists (`cli-template-specialist`, `quality-specialist`, `workflow-specialist`, `hook-ci-specialist`) to the v4 manifest format (each gaining a `primitive` / `isolation` / `effort` / `model` mapping); the legacy `/moai:harness` 7-Phase path (`moai-meta-harness` skill) shall become a redirect to the v4 Builder — **when** a user invokes the legacy path, the system shall surface a deprecation notice and route to `/moai:harness` v4; the revfactory 7-Phase residuals in the `moai-meta-harness` skill body shall be marked superseded.

## §D. Constraints

- **C-HV4-001 (Anthropic "simplest solution first")**: Every load-bearing component in the Builder/Runner must justify its cost. As models improve, the v4 design must be amenable to removing components (context resets, sprint constructs, even the evaluator) one at a time and measuring — NOT treating any component as permanent.
- **C-HV4-002 (Subagent spawning ceiling)**: The Runner is a dynamic-workflow script; the Builder is orchestrator-direct (orchestrator-side logic, not a separate agent or script). Sub-agents spawned by either the orchestrator (Builder phases) or the Runner (`agent()` calls) are leaf agents and MUST NOT themselves spawn sub-agents (Anthropic sub-agents cannot spawn other sub-agents). Hierarchical delegation is expressed at the manifest level (supervisor specialist dispatches worker specialists), NOT by recursive subagent spawning.
- **C-HV4-003 (Dynamic-workflow determinism)**: The Runner Workflow script body MUST be deterministic — no `Date.now()` or `Math.random()` calls in the script body (resume caching keys on deterministic outputs). Timestamps and random values must be injected via script input arguments or stamped onto results after the run returns. The Builder is orchestrator-direct and follows normal orchestrator determinism discipline (no reliance on wall-clock/random for control flow); the determinism constraint on script bodies applies ONLY to the Runner.
- **C-HV4-004 (AskUserQuestion boundary — Runner-only)**: Runner sub-agents are invoked via `agent()` inside a dynamic-workflow script and CANNOT prompt the user mid-run; the orchestrator-direct Builder phases CAN call AskUserQuestion at stage boundaries — in particular the PLAN→GENERATE approval gate (REQ-HV4-001 approval pattern extended to the build). The orchestrator MUST drain all preferences for a Runner invocation (harness name confirmation, plan approval, A/B thresholds) before that dynamic-workflow launches, OR collect them at Builder stage boundaries (which is first-class because the orchestrator holds the boundary).
- **C-HV4-005 (Template neutrality §25)**: The v4 harness doctrine text embedded in `internal/template/templates/` MUST be generic (mechanism descriptions, public-source citations, permanent-rule citations) and MUST NOT leak internal-state markers (SPEC IDs, REQ tokens, audit citations, commit SHAs, archive paths).
- **C-HV4-006 (Worktree L1 autonomy)**: Per moai-adk §14 advisory policy + the 2026-05-17 user opt-in policy, L1 `Agent(isolation:"worktree")` is runtime-autonomous — the orchestrator does NOT mandate worktree creation; worktree isolation is advisory and sub-agent-granular.

## §E. Non-Functional Requirements

- **NFR-HV4-001 (Performance)**: The Builder's orchestrator-direct phases (ANALYZE → GENERATE) shall use parallel `Agent()` fan-out to minimize wall-clock time for multi-specialist generation. The Builder is NOT bound by the dynamic-workflow runtime budget (it uses ordinary `Agent()` spawn); the Runner IS bound by the dynamic-workflow runtime budget (1000 agents total / 16 concurrent) and that budget applies to Runner invocations only.
- **NFR-HV4-002 (Observability)**: Every Builder/Runner phase shall emit a structured progress record (phase name, primitive used, specialist count, isolation decisions, Sprint Contract dimensions) into the manifest's run log section, so that a third party can audit which phases ran, which were skipped, and why.
- **NFR-HV4-003 (Reversibility)**: `/moai:harness remove <name>` shall be a clean reverse of `/moai:harness <request>` — the harness can be removed without residue, and the legacy 7-Phase path remains accessible via git history for a deprecation window.
- **NFR-HV4-004 (Honesty / verification-claim integrity)**: The dogfooding validation report shall follow the 5-Section Evidence-Bearing Report Format (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk) per `.claude/rules/moai/core/verification-claim-integrity.md`. The +60% A/B claim shall be attributed to the exact command run and observed output; any unmeasured dimension shall be reported as a Gap, not a Claim.

## §F. Success Criteria

- 13 REQs (REQ-HV4-001..013) implemented and each has ≥1 matching AC in `acceptance.md` that passes.
- A "moai-adk-dev" harness built with v4 successfully runs a real moai-adk development task end-to-end (sample task disclosed in `acceptance.md`).
- `grep` for revfactory 7-Phase residuals ("7-Phase", "Phase 7 LEARNING", "Skeleton", "Customization") in newly generated v4 artifacts returns zero matches.
- `moai update` preserves user-owned `harness-*` / `.claude/commands/harness/` / `.claude/workflows/harness-*.js` artifacts (namespace AC passes).
- spec-lint passes on this SPEC's frontmatter (12 canonical fields, era: V3R6) and body (OutOfScope section present, GEARS notation compliant).

## §G. Risks

- **R-HV4-001 (Dynamic-workflow runtime availability — Runner-only)**: Dynamic workflows require Claude Code v2.1.154+ and a paid plan. The Builder is orchestrator-direct and does NOT depend on the dynamic-workflow runtime (it uses ordinary `Agent()` spawn, available in all Claude Code versions). The Runner DOES depend on the dynamic-workflow runtime; if that runtime is unavailable for a Runner invocation, the Runner falls back to sequential sub-agent dispatch per manifest.primitive (the primitive mapping is preserved; only the `agent()`-inside-a-script dispatch path degrades to sequential `Agent()` calls). Mitigation: design.md §F specifies the Runner's fallback path.
- **R-HV4-002 (Worktree cleanup)**: Conditional `Agent(isolation:"worktree")` spawns accumulate worktrees if cleanup fails. Mitigation: L1 worktree is runtime-autonomous-cleanup; the Runner emits a cleanup directive at end-of-run.
- **R-HV4-003 (Sprint Contract over-engineering)**: The Sprint Contract can become a ceremonial artifact that adds cost without value for simple tasks. Mitigation: REQ-HV4-008 makes the evaluator conditional; the Sprint Contract is load-bearing only when the task exceeds the model's solo range.
- **R-HV4-004 (Migration regression)**: Porting the 4 existing specialists to v4 manifest format could break the moai-adk Layer B specialist team. Mitigation: M6 milestone runs the existing Layer B regression suite as an AC.

## §H. Cross-References

- `research.md` — revfactory/harness 7-Phase analysis (verbatim strengths/removals), Anthropic harness-design GAN guidance, Claude Code dynamic-workflow spec, Sprint Contract pattern, Meta-Harness outer-loop forward-link.
- `design.md` — v4 architecture: entry-point model, 4-phase Builder, Runner Workflow, manifest schema (full), worktree policy, revfactory 7-Phase → v4 mapping table.
- `plan.md` — 6 milestones (M1 entry + namespace / M2 Builder orchestrator-direct logic + 6-pattern catalog + GENERATE output spec / M3 manifest + Runner / M4 `/harness:<name>` + lifecycle / M5 conditional worktree / M6 migrate 4 specialists + legacy redirect).
- `acceptance.md` — AC-HV4-NNN acceptance criteria, one or more per REQ.
- `progress.md` — §E skeleton (run-phase evidence placeholder).
- `.moai/specs/SPEC-V3R6-HARNESS-NAMESPACE-V2-001/` — completed baseline; §24 namespace protection this SPEC extends.
- `.moai/specs/SPEC-V3R6-LIFECYCLE-REDESIGN-001/` — 3-phase plan/run/sync lifecycle (MX Tag cross-cutting, not separate phase).
- `.moai/specs/SPEC-V3R6-WORKFLOW-EFFORT-MAP-001/` — dynamic-workflow `agent()` purpose-driven model/effort taxonomy (read-only-extract / mechanical-transform / synthesize / verify-judge / implement / design-architecture).
- `.claude/rules/moai/workflow/dynamic-workflows.md` — canonical dynamic-workflow primitive selection guide + determinism constraint + `codemaps-extract.js` precedent.
- `.claude/rules/moai/core/verification-claim-integrity.md` — 5-Section Evidence-Bearing Report Format (binds REQ-HV4-012 dogfooding validation report).
- `.claude/skills/moai-meta-harness/SKILL.md` — legacy 7-Phase generator; becomes a redirect target per REQ-HV4-013.

---

Version: 0.2.0 (run-phase, architecture pivot applied: Builder → orchestrator-direct; Runner stays dynamic-workflow)
Classification: SPEC — feature to implement (forward-looking v4 harness redesign)
