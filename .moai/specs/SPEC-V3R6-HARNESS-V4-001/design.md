# design.md — SPEC-V3R6-HARNESS-V4-001 Architecture

> The v4 harness rebuild architecture. Formalizes the entry-point model, 4-phase Builder, Runner Workflow, manifest schema (full), worktree policy, and the revfactory 7-Phase → v4 mapping. Grounded in the research findings of `research.md`.

## §A. Design Principles

1. **Absorb strengths, remove the rest.** From revfactory/harness: keep domain-sentence → team-architecture, 6-pattern catalog, Progressive-Disclosure skill generation, Generator-Evaluator separation, A/B validation. Remove: static 7-Phase, Agent-Teams hard dependency, Skeleton↔Customization split, LEARNING separate phase, revfactory plugin assumptions, Socratic-only discovery.
2. **Anthropic "simplest solution first"** (research.md §B). Every load-bearing component must justify its cost. Phase synthesis from task signals, NOT fixed pipeline. Evaluator conditional. Components removable as models improve.
3. **Two execution axes — orchestrator-direct Builder + dynamic-workflow Runner** (research.md §C). The v4 design splits execution across two distinct primitives:
   - **Builder = orchestrator-direct.** The orchestrator runs ANALYZE / PLAN / GENERATE / ACTIVATE as orchestrator-side logic using ordinary `Agent()` spawn. Intermediate results are held in the orchestrator's session context (the plan lives in Claude's context, NOT in a script). AskUserQuestion is available at stage boundaries (the PLAN→GENERATE approval gate is first-class because the orchestrator holds the boundary).
   - **Runner = dynamic-workflow script.** The generated `harness-<name>-run.js` consumes `manifest.json` and dispatches specialists per their `primitive`. The script holds the plan; intermediate results live in script variables; up to 16 concurrent / 1000 total agents. The `codemaps-extract.js` precedent is the Runner's pattern. The Runner CANNOT call AskUserQuestion mid-run (all preferences must be drained before launch).

   This split resolves the design self-contradiction where C-HV4-004 forbade mid-run AskUserQuestion but §D.2 required a PLAN→GENERATE approval gate: the gate now lives in the orchestrator-direct Builder (where AskUserQuestion is available), and the AskUserQuestion prohibition applies only inside the Runner script.
4. **Generator-Evaluator separation** (research.md §B, Anthropic GAN). The self-evaluation bias is the core problem; separating generator from evaluator is the strongest lever. The evaluator earns its cost only when the task exceeds the model's solo range.
5. **Conditional worktree isolation** (spec.md REQ-HV4-007, moai-adk §14 advisory). No mandatory top-level worktree. Sub-agent-granular `Agent(isolation:"worktree")` only for conflict-prone parallel generation or risky changes.

## §B. Entry-Point Model — Creation vs Execution Split

The v4 design splits harness creation/management from harness execution:

```
┌─────────────────────────────────────────────────────────────┐
│  CREATION & MANAGEMENT (orchestrator-direct)                 │
│  /moai:harness <NL request>  → Builder (orchestrator-direct) │
│  /moai:harness list|edit|remove <name>  → lifecycle          │
└──────────────────────────────┬──────────────────────────────┘
                               │ Builder GENERATE emits
                               ▼
┌─────────────────────────────────────────────────────────────┐
│  EXECUTION (dynamic-workflow)                                │
│  /harness:<name>  → harness-<name>-run.js (Runner Workflow)  │
│   ↑ thin wrapper command at .claude/commands/harness/<name>.md │
└─────────────────────────────────────────────────────────────┘
```

### §B.1 Creation: `/moai:harness <NL request>`

User issues natural language: `/moai:harness build a harness for moai-adk CLI template development`. The orchestrator:

1. **Context-First Discovery** on the request — extract domain, goal, constraints, scope (REQ-HV4-001). Source: codebase + docs (.moai/docs, .claude/rules, CLAUDE.md, NOTICE.md) + existing harness/* + SPEC history.
2. **AskUserQuestion Socratic rounds** if clarity <100% (≤4 questions/round, `(Recommended)` first option). Repeat until 100%.
3. **Derive harness `<name>`** from the confirmed intent (NOT user-supplied statically). E.g., "moai-adk CLI template development" → `moai-adk-dev` or `cli-template`.
4. **Explicit approval gate** — orchestrator surfaces the derived name + extracted profile + planned approach; user approves via AskUserQuestion.
5. **Enter Builder orchestrator-direct processing** — the orchestrator runs ANALYZE / PLAN / GENERATE / ACTIVATE directly (NOT a delegation to a JS script). Discovery hands off to the Builder phases within the orchestrator session; the PLAN→GENERATE approval gate fires at the stage boundary (AskUserQuestion, first-class because the orchestrator holds the boundary).

This UPGRADES the existing `/moai:harness` command (which currently routes to the legacy 7-Phase `moai-meta-harness` skill). Per REQ-HV4-013, the legacy 7-Phase path becomes a redirect to v4.

### §B.2 Execution: `/harness:<name>`

Each harness auto-generates its own thin-wrapper command at `.claude/commands/harness/<name>.md` at Builder GENERATE time. The command dispatches to that harness's Runner Workflow `harness-<name>-run.js`. Claude Code subdirectory-command resolution maps `.claude/commands/harness/dev.md` → `/harness:dev` (BI-001 pre-flight verifies this).

The harness is **self-describing**: invoking `/harness:<name>` is equivalent to "run this harness's Runner, which reads its own manifest and dispatches specialists". The user does not need to know the Runner Workflow filename.

### §B.3 Lifecycle

- `/moai:harness list` — enumerate harnesses by scanning `.claude/commands/harness/*.md` and joining with their `manifest.json`.
- `/moai:harness edit <name>` — open the manifest + specialists for editing (manifest is the SSOT; editing it propagates to Runner behavior on next invocation).
- `/moai:harness remove <name>` — atomic removal of command + workflow + specialists + skills + manifest. Fail-closed if any referenced artifact is missing (orphan prevention, REQ-HV4-011).

## §C. manifest.json Canonical Schema

```json
{
  "name": "<harness-name>",
  "domain": "<domain-description>",
  "source_request": "<original natural-language request verbatim>",
  "patterns": ["pipeline", "fan-out-fan-in", "..."],
  "specialists": [
    {
      "role": "<specialist-role-description>",
      "primitive": "sub-agent | dynamic-workflow | worktree | /goal | adversarial-fan-out",
      "isolation": "none | worktree",
      "effort": "low | medium | high | xhigh | max",
      "model": "inherit | haiku | sonnet | opus"
    }
  ],
  "sprint_contract": {
    "dimensions": ["<graded-dimension-1>", "<graded-dimension-2>"],
    "thresholds": { "<dimension-1>": "<threshold-value>" }
  },
  "entry_command": "/harness:<name>",
  "runner_workflow": "harness-<name>-run.js"
}
```

### §C.1 Field Semantics

- **`name`** — the harness name (derived from the NL request). Matches the `/harness:<name>` command and the `harness-<name>-run.js` Runner Workflow filename. Constraint: `[a-z][a-z0-9-]*` (kebab-case, DNS-safe).
- **`domain`** — short human-readable domain description (e.g., "moai-adk CLI template development").
- **`source_request`** — the original natural-language request verbatim. Preserved for audit/re-generation.
- **`patterns`** — array of pattern names drawn from the 6-pattern catalog (§E). Selected/combined by PLAN phase dynamically.
- **`specialists[]`** — array of specialist role definitions. Each has:
  - `role` — the specialist's responsibility (e.g., "template-neutrality-auditor").
  - `primitive` — the execution primitive (REQ-HV4-005). Runner dispatches verbatim.
  - `isolation` — `none` (main-tree) or `worktree` (Agent(isolation:"worktree") sub-agent). Per-specialist, conditional (REQ-HV4-007).
  - `effort` — reasoning effort level (low/medium/high/xhigh/max). Per `SPEC-V3R6-WORKFLOW-EFFORT-MAP-001` purpose-driven taxonomy.
  - `model` — model tier (inherit/haiku/sonnet/opus). Per the effort-map purpose taxonomy.
- **`sprint_contract`** — Anthropic GAN Sprint Contract (REQ-HV4-008). `dimensions` is the array of graded dimensions agreed pre-coding; `thresholds` maps each dimension to its pass threshold.
- **`entry_command`** — the `/harness:<name>` string (redundant with `name` but explicit for tooling).
- **`runner_workflow`** — the Runner Workflow filename `harness-<name>-run.js`.

### §C.2 Schema Validation Rules

- All 8 top-level fields required (REQ-HV4-006 / AC-HV4-006a).
- `specialists[]` MUST be non-empty (≥1 specialist).
- Each specialist's `primitive` MUST be exactly one of the 5 primitives (no free-text).
- Each specialist's `isolation` MUST be `none` or `worktree`.
- `patterns[]` entries MUST be from the 6-pattern catalog (no custom patterns).

## §D. Builder (orchestrator-direct) — 4 Signal-Driven Phases

The Builder is orchestrator-direct processing (NOT a dynamic-workflow script and NOT a separate agent). The orchestrator runs the 4 phases using ordinary `Agent()` spawn; intermediate results are held in the orchestrator's session context (the plan lives in Claude's context, NOT in a script). Which phases run is synthesized from task signals (load-bearing minimum). The AskUserQuestion boundary is first-class at Builder stage boundaries — the PLAN→GENERATE approval gate (§D.2) fires because the orchestrator holds the boundary directly.

### §D.1 ANALYZE [orchestrator parallel Agent(Explore) fan-out, read-only, main tree]

**Primitive**: orchestrator-direct parallel `Agent()` fan-out — `Agent(agentType: "Explore", effort: "low")` per source package/surface (orchestration-mode Mode 4 parallel multi-spawn; see `.claude/rules/moai/workflow/orchestration-mode-selection.md` §B).
**Isolation**: none (read-only).
**Purpose** (per effort-map): read-only-extract.

Fan out N Explore sub-agents across the codebase + docs surfaces:
- Source packages (dep-graph + public-surface extraction, per the `codemaps-extract.js` precedent).
- `.moai/docs/*`, `.claude/rules/moai/*`, `CLAUDE.md`, `NOTICE.md`.
- Existing `.claude/agents/harness/*` + `.claude/skills/harness-*/` + `.claude/commands/harness/*`.
- SPEC history (`.moai/specs/SPEC-*-HARNESS-*/`).

Aggregate in the orchestrator's session context (the orchestrator receives each sub-agent's result and synthesizes). Produce: domain profile + task-pattern inventory.

### §D.2 PLAN [orchestrator spawns single Agent(opus,xhigh), main tree; AskUserQuestion gate]

**Primitive**: orchestrator-direct — the orchestrator spawns a single `Agent(model: "opus", effort: "xhigh")` sub-agent (orchestration-mode Mode 5 sequential single-spawn for the deep-reasoning step).
**Isolation**: none.
**Purpose**: design-architecture (per effort-map).

One opus xhigh sub-agent reasons over the ANALYZE aggregate and:
- Selects/combines patterns from the 6-pattern catalog (§E) based on task signals (parallelism, adversarial-verification need, supervision depth, expertise diversity).
- Defines specialist roles.
- Maps each specialist to an execution primitive (sub-agent / dynamic-workflow / worktree / `/goal` / adversarial-fan-out).
- Decides per-specialist `isolation` (worktree only for conflict-prone parallel generation).
- Drafts the Sprint Contract (graded dimensions + thresholds).
- Emits a draft manifest.

**User approval gate** (AskUserQuestion) BEFORE GENERATE — the orchestrator surfaces the draft manifest + planned specialists + Sprint Contract; the user approves or revises. This gate is **first-class** because the orchestrator holds the PLAN→GENERATE boundary directly (the Builder is orchestrator-direct, so AskUserQuestion is available at this stage boundary — this is the self-contradiction resolution: the gate could NOT fire if the Builder were a dynamic-workflow script, but the orchestrator-direct Builder makes it reachable).

### §D.3 GENERATE [orchestrator fan-out emits 5 artifact types; conditional `Agent(isolation:"worktree")`]

**Primitive**: orchestrator-direct fan-out — the orchestrator spawns specialist `Agent()` calls (parallel where independent, sequential where dependent) that emit the 5 artifact types below.
**Isolation**: conditional per-specialist (worktree for conflict-prone, none otherwise).

Fan out specialist agents that emit:
- `.claude/commands/harness/<name>.md` (thin-wrapper entry command → dispatchs to Runner).
- `harness-<name>-run.js` (the Runner Workflow — a dynamic-workflow script, lives at `.claude/workflows/harness-<name>-run.js`).
- `harness-<name>-*-specialist.md` (sub-agent definitions, live at `.claude/agents/harness/`).
- `harness-<name>-*/SKILL.md` (companion Progressive-Disclosure skills, live at `.claude/skills/harness-<name>-*/`).
- `manifest.json` (final, validated against §C schema — the SSOT consumed by the Runner).

Conflict-prone parallel generation (≥2 specialists targeting overlapping paths) spawns `Agent(isolation:"worktree")` per the manifest's `isolation` declarations. Read-only or sequential generation runs main-tree. The GENERATE output contract (the 5 artifact types + their content contracts) is the handoff spec M3/M4 consume.

### §D.4 ACTIVATE [orchestrator-direct dry-run + `/goal`, main tree unless risky]

**Primitive**: orchestrator-direct — dry-run sub-agent + `/goal` (autonomous convergence) + with/without A/B. The orchestrator drives this phase directly (it is NOT a dynamic-workflow script).
**Isolation**: none (default) or worktree (if the sample task is risky).

Run a sample task dry-run through the newly built harness. Use `/goal "<harness> completes task X at quality Y"` for autonomous convergence (per `.claude/rules/moai/workflow/goal-directive.md`). Run a with/without A/B (Anthropic skeptical evaluator). Pass → expose the harness via `/harness:<name>`. Fail → regress to GENERATE for refinement.

**Load-bearing minimum**: for tasks within the model's solo reliable range, ACTIVATE's A/B is SKIPPED (REQ-HV4-008, C-HV4-001). The phase records the skip with rationale.

## §E. 6-Pattern Catalog

Inherited from revfactory/harness (research.md §A), selected/combined dynamically by PLAN:

| Pattern | When selected | Execution primitive affinity |
|---------|---------------|------------------------------|
| **Pipeline** | Sequential transformation stages (output of stage N = input of stage N+1) | sub-agent chain |
| **Fan-out/Fan-in** | Parallel independent work + aggregation | dynamic-workflow |
| **Expert Pool** | Diverse expertise consulted in parallel | dynamic-workflow |
| **Producer-Reviewer** | Generator-Evaluator separation (adversarial) | adversarial-fan-out |
| **Supervisor** | Centralized coordination + dispatch | sub-agent (supervisor) + workers |
| **Hierarchical Delegation** | Multi-level delegation (manifest-level, NOT recursive subagent) | sub-agent per level |

PLAN selects ≥1 pattern per harness; the selection is recorded in `manifest.json.patterns`. AC-HV4-004a requires ≥2 distinct patterns across 2 different domain requests (the selection is NOT fixed).

## §F. Runner Workflow (`harness-<name>-run.js`)

The Runner is generated per harness by the Builder GENERATE phase. It reads `manifest.json` and dispatches specialists per their declared primitive:

```
read manifest.json
for each specialist in manifest.specialists:
    switch specialist.primitive:
        case "sub-agent":      Agent(specialist.role, effort: specialist.effort, model: specialist.model)
        case "dynamic-workflow": dynamicWorkflow(specialist.role, ...)
        case "worktree":        Agent(specialist.role, isolation: "worktree", ...)
        case "/goal":           goalDirective(specialist.role, ...)
        case "adversarial-fan-out": adversarialFanOut(specialist.role, ...)
    apply Sprint Contract (if evaluator not skipped)
emit cleanup directive (worktree cleanup if any isolation:worktree specialists)
```

The Runner consumes `specialist.primitive` verbatim — it does NOT re-derive the primitive from heuristics (REQ-HV4-005 / AC-HV4-005b). The Sprint Contract (Generator-Evaluator separation) is applied per REQ-HV4-008; the evaluator is conditional.

### §F.1 Script Determinism (C-HV4-003)

The Runner script body is deterministic — no `Date.now()` or `Math.random()` calls. Timestamps injected via script input arguments or stamped onto results after the run returns (per `.claude/rules/moai/workflow/dynamic-workflows.md` determinism constraint).

## §G. Worktree Policy — Conditional, Sub-Agent-Granular (v4 Core)

The v3 mistake was always creating a top-level worktree. v4 makes worktree isolation conditional and sub-agent-granular:

| Phase | Default isolation | Rationale |
|-------|-------------------|-----------|
| ANALYZE | none (read-only fan-out) | Read-only; no write conflicts possible |
| PLAN | none (single sub-agent) | Single agent; no parallel writes |
| GENERATE | conditional per-specialist | Conflict-prone parallel generation → worktree; else main-tree |
| ACTIVATE | none (or worktree if sample task risky) | Dry-run; risky → isolate |

NO mandatory top-level worktree wraps the Builder or Runner. `Agent(isolation:"worktree")` is spawned ONLY for specialists whose manifest declares `isolation:worktree`. L1 worktree cleanup is runtime-autonomous (C-HV4-006).

## §H. revfactory 7-Phase → v4 Mapping

| revfactory phase | v4 equivalent | Rationale |
|------------------|---------------|-----------|
| Phase 1 Discovery (Socratic) | ANALYZE (Explore fan-out) | interview-only → ground-truth full codebase+docs analysis |
| Phase 2 Analysis | merged into ANALYZE | separate phase unnecessary (model improved; simplest-solution-first) |
| Phase 3 Synthesis (SPEC EARS) | PLAN (1 opus xhigh agent) | kept, simplified; 6 patterns selected dynamically (not statically fixed) |
| Phase 4+5 Skeleton + Customization | GENERATE (1 fan-out) | 2 sequential phases → 1 parallel fan-out |
| Phase 6 Evaluation | ACTIVATE (dry-run + `/goal` + A/B) | single sync-auditor → autonomous `/goal` + adversarial A/B |
| Phase 7 Iteration (LEARNING) | cross-cutting (in Sprint Contract) | NOT a separate phase; learning folded into Sprint Contract dimensions |
| — (new in v4) | conditional worktree isolation | v4 new; main-tree pollution = 0 |

## §I. Namespace Policy Extension (§24 commands/ + workflows/)

Extends the completed `SPEC-V3R6-HARNESS-NAMESPACE-V2-001` namespace protection to cover the 2 new user-owned surfaces introduced by v4. (After the orchestrator-direct pivot, `.claude/workflows/harness-*.js` protects the **Runner** `harness-<name>-run.js` files — there is no longer a Builder JS file in this namespace. The `.claude/commands/harness/` row protects the thin-wrapper entry commands emitted by GENERATE.)

| Surface | Owner | `moai update` behavior |
|---------|-------|------------------------|
| `.claude/skills/harness-*/` | user-owned (existing, V2-001) | preserve (backup if needed) |
| `.claude/agents/harness/` | user-owned (existing, V2-001) | preserve |
| `.claude/commands/harness/` | **user-owned (NEW, this SPEC)** | **preserve** |
| `.claude/workflows/harness-*.js` (Runner scripts `harness-<name>-run.js`) | **user-owned (NEW, this SPEC)** | **preserve** |
| `.claude/skills/moai-harness-*/` | template-managed (builder) | sync (overwrite) |
| `.claude/skills/moai-meta-harness/` | template-managed (builder, legacy redirect) | sync (overwrite) |

The Go enforcement extends `isUserOwnedNamespace` / `isUserAreaPath` in `internal/cli/update*.go` to recognize the two new paths (M1 landed this at commit `823b20570` — `isUserAreaPath`/`isUserOwnedNamespace` now protect `.claude/commands/harness/` and `.claude/workflows/harness-*.js`). Template-neutrality §25 (C-HV4-005) forbids leaking any `harness-*` / `commands/harness/` / `workflows/harness-*.js` into `internal/template/templates/`.

## §J. Migration Path (M6)

1. Port the 4 existing specialists to v4 manifest format (each gains primitive/isolation/effort/model). Preserve behavior; Layer B regression suite must pass.
2. Convert `moai-meta-harness` legacy 7-Phase path to a redirect. Honor or explicitly supersede the `@MX:NOTE [AUTO] V3R4 contract` annotation (record rationale).
3. Grep revfactory residuals in new v4 artifacts → 0 matches (AC-HV4-009a).
4. Dogfooding: build "moai-adk-dev" harness with v4, run real task, with/without A/B, 5-Section Evidence-Bearing Format report.

## §K. Design Alternatives Considered

### Alternative A — Keep revfactory 7-Phase, patch incrementally
**Rejected**: the 6 over-engineering liabilities (spec.md §A) are structural; patching preserves them. The user's directive ("하네스를 구축해줘") calls for a rebuild, not a patch.

### Alternative B — Hard-depend on Agent-Teams (3-5 cap) as execution vehicle
**Rejected**: Anthropic guidance caps Agent-Teams at 3-5 for coordination-cost reasons. Large parallel harness generation (ANALYZE fan-out across 10+ packages) exceeds this cap. Dynamic-workflow (16 concurrent / 1000 total) is the right primitive.

### Alternative C — Mandatory top-level worktree for Builder/Runner
**Rejected**: v3 mistake. Always-worktree pollutes the worktree registry and adds L1 cleanup overhead for read-only phases where no write conflict is possible. Conditional sub-agent-granular isolation (v4) is simpler and aligns with §14 advisory policy.

### Alternative D — Implement learning-subsystem outer loop in this SPEC
**Rejected**: out of scope (spec.md §B.2). The Meta-Harness outer-loop optimization (arxiv 2603.28052) is a separate concern. This SPEC declares learning as cross-cutting (Sprint Contract dimension); a follow-up SPEC owns the outer loop.

### Alternative E — Builder as dynamic-workflow script `harness-build.js` (ORIGINAL v4 design, superseded mid-run-phase 2026-06-19)
**Superseded**: the original v4 design (drafts 0.1.0 → 0.1.1) specified the Builder as a Claude Code dynamic-workflow script `harness-build.js`, parallel to the Runner `harness-<name>-run.js`. This was **pivoted away** at M1-complete (commit `db4adec39`) for three verified reasons:

1. **design.md self-contradiction**: C-HV4-004 forbids mid-run AskUserQuestion inside a dynamic-workflow, but §D.2 requires a PLAN→GENERATE approval gate. These cannot coexist if the Builder is a dynamic-workflow (the gate could not fire inside the script). Orchestrator-direct Builder makes the approval gate first-class — the orchestrator calls AskUserQuestion directly at the stage boundary.
2. **Namespace conflict disappears**: no `harness-build.js` / `moai-harness-build.js` file is needed, so the §24.1 builder=`moai-harness-*` (template-managed) vs `harness-*` (user-owned) classification tension that blocked M2 is moot.
3. **Anthropic "simplest solution first" (C-HV4-001)**: the orchestrator already has Agent/Skill/AskUserQuestion; M1's `/moai:harness` NL entry already starts orchestrator-direct Discovery. Extending orchestrator-direct through ANALYZE/PLAN/GENERATE/ACTIVATE is the natural continuation — no need to jump into a workflow script and back.

The Runner stays a dynamic-workflow script (user-confirmed 2026-06-19: "execution runs INSIDE the generated `/harness:<name>` command, which drives workflow + sub-agent + dynamic-workflow per the manifest"). The `codemaps-extract.js` precedent remains the Runner's pattern (§F).

## §L. Cross-References

- `research.md` — revfactory 7-Phase analysis + Anthropic GAN guidance + verbatim source citations underpinning every design decision above.
- `spec.md` §C — 13 REQs this design implements.
- `acceptance.md` §D — ACs that verify each design element.
- `.claude/workflows/codemaps-extract.js` — the canonical dynamic-workflow precedent (read-only fan-out pattern) that the **Runner** follows (§F). The Builder's ANALYZE phase is now orchestrator-direct parallel `Agent(Explore)` fan-out, not a dynamic-workflow script.
- `.claude/rules/moai/workflow/dynamic-workflows.md` — primitive selection guide, determinism constraint, purpose-driven effort taxonomy.
- `.claude/rules/moai/workflow/goal-directive.md` — `/goal` autonomous convergence (used in ACTIVATE).
- `.claude/rules/moai/core/verification-claim-integrity.md` — 5-Section Evidence-Bearing Report Format (binds the dogfooding validation report).
