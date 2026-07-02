---
description: >
  Orchestrator-direct Builder for harness creation. Runs the 4 signal-driven
  phases ANALYZE, PLAN, GENERATE, ACTIVATE as orchestrator-side logic using
  ordinary Agent() spawn (NOT a dynamic-workflow script, NOT a separate
  subagent). Intermediate results are held in the orchestrator's session
  context. Documents the 6-pattern catalog, the per-specialist primitive
  mapping, the conditional worktree-isolation policy, and the GENERATE output
  contract (the 5 artifact types M3/M4 consume).
user-invocable: false
metadata:
  version: "1.0.0"
  category: "workflow"
  status: "active"
  updated: "2026-06-19"
  tags: "harness, builder, orchestrator-direct, 4-phase, pattern-catalog, generate-contract"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["harness builder", "harness analyze", "harness generate", "harness activate"]
  agents: ["moai-meta-harness"]
  phases: ["harness"]
---

# Module: harness-builder — Orchestrator-Direct Builder (4 Phases)

Purpose: This module documents the **Builder**, the orchestrator-direct processing that turns a confirmed harness-creation profile (domain / goal / constraints / scope + derived `<name>`, produced by the entry workflow's Context-First Discovery and approval gate) into a concrete, manifest-driven harness.

The Builder is **orchestrator-side logic**. It is NOT a dynamic-workflow script and it is NOT a separate subagent. The orchestrator runs the 4 phases using ordinary `Agent()` spawn; intermediate results are held in the orchestrator's session context (the plan lives in Claude's context, NOT in a script). This is what makes the PLAN→GENERATE approval gate first-class: because the orchestrator holds the boundary directly, `AskUserQuestion` is available at that stage boundary.

> **Architecture note (orchestrator-direct pivot).** An earlier design specified the Builder as a Claude Code dynamic-workflow script. That design was superseded because a dynamic-workflow script cannot call `AskUserQuestion` mid-run, which made the PLAN→GENERATE approval gate unreachable. Orchestrator-direct processing resolves the contradiction. The **Runner** (`harness-<name>-run.js`) STAYS a dynamic-workflow script — execution runs INSIDE the generated `/harness:<name>` command. Only the Builder (creation) is orchestrator-direct.

## Authoritative Sources

- Manifest schema: the companion design document's manifest-schema section (8 top-level fields, specialist shape, Sprint Contract).
- Runner Workflow contract: the companion design document's Runner section (reads `manifest.json`, dispatches per `specialist.primitive`).
- Dynamic-workflow primitive + determinism + purpose-driven effort taxonomy: `.claude/rules/moai/workflow/dynamic-workflows.md`.
- Orchestration mode selection (parallel multi-spawn / sequential single-spawn): `.claude/rules/moai/workflow/orchestration-mode-selection.md`.
- AskUserQuestion channel monopoly + preload: `.claude/rules/moai/core/askuser-protocol.md`.
- Orchestrator-subagent boundary: `.claude/rules/moai/core/agent-common-protocol.md` § User Interaction Boundary.
- Conditional worktree isolation advisory: `.claude/rules/moai/workflow/worktree-integration.md`.
- `/goal` autonomous convergence: `.claude/rules/moai/workflow/goal-directive.md`.
- 5-Section Evidence-Bearing Report Format: `.claude/rules/moai/core/verification-claim-integrity.md`.

## Handoff from the Entry Workflow

The entry workflow (`harness-build-entry.md`) runs Phases 0-3 (reserved-verb guard, Context-First Discovery, harness `<name>` derivation, orchestrator-issued approval gate). On **Build** approval, the entry hands the confirmed profile + derived name to this Builder module. The orchestrator does NOT delegate to a script or a separate agent — it continues executing in the same session, transitioning from Discovery into the Builder's 4 phases.

**Carry-over invariant.** The confirmed profile + derived name handed off from the entry is the single source of truth for the manifest's `source_request` field. The Builder MUST carry the original natural-language request verbatim into `manifest.json.source_request`.

## The 4 Phases (signal-driven, load-bearing minimum)

Which phases run is **synthesized from task signals** — the Builder does NOT run a fixed pipeline. The "load-bearing minimum" principle (simplest-solution-first) means: for a task within the model's solo reliable range, phases that add no value are skipped, and the skip is recorded with rationale. The 4 phases below are the full set; a given harness build may run all 4 or skip one or more.

```
ANALYZE  →  PLAN (approval gate)  →  GENERATE  →  ACTIVATE (A/B optional)
   |             |                       |              |
   |             |                       |              └─ load-bearing minimum:
   |             |                       |                 A/B skipped for simple tasks
   |             |                       └─ conditional Agent(isolation:"worktree")
   |             |                          per conflict-prone specialist
   |             └─ orchestrator holds the boundary;
   |                AskUserQuestion fires here
   └─ orchestrator parallel Agent(Explore) fan-out, read-only, main tree
```

### Phase 1 — ANALYZE [orchestrator parallel Agent(Explore) fan-out, read-only, main tree]

**Primitive**: orchestrator-direct parallel `Agent()` fan-out — one `Agent(agentType: "Explore", effort: "low")` per source surface (orchestration-mode parallel multi-spawn).
**Isolation**: none (read-only — no write conflicts are possible).
**Purpose** (per the effort taxonomy): read-only-extract.

Fan out N Explore sub-agents across the codebase + docs surfaces relevant to the confirmed domain:

- Source packages — dependency-graph + public-surface extraction per source package (per the `codemaps-extract.js` precedent: the deterministic baseline is a tool like `go list -deps -json` + `go doc`; the per-package Explore agent layers architecture-REVIEW insight on top).
- Project docs — the project's docs directory, rules directory, root instruction file, and third-party notices file.
- Existing harness surfaces — existing harness agents, harness skills, and harness commands already in the project.
- SPEC history — prior harness-related SPECs in the project's specs directory.

Each Explore agent returns a structured markdown summary. The orchestrator receives each sub-agent's result and synthesizes in its own session context. The phase produces:

- A **domain profile** — the subject area, its boundaries, its canonical workflows, and its quality gates.
- A **task-pattern inventory** — the recurring shapes of work in this domain (sequential transformation stages, parallel independent work + aggregation, diverse expertise consulted in parallel, generator-evaluator separation, centralized coordination, multi-level delegation).

**Effort discipline.** Each Explore agent carries `effort: "low"` (read-only-extract purpose). Raising effort on the extraction step multiplies token cost without improving the mechanical baseline; the architecture-insight value comes from the prompt, not from raising effort. Omitting `effort` is a cost leak — it inherits the session default.

**Load-bearing minimum.** If the domain is already well-understood from the Discovery profile (e.g., a single-skill harness with a narrow, well-bounded scope), the orchestrator MAY collapse ANALYZE into a single Explore agent or skip the per-package fan-out. The collapse/skip is recorded with rationale.

### Phase 2 — PLAN [orchestrator spawns single Agent(opus, xhigh); AskUserQuestion gate]

**Primitive**: orchestrator-direct — the orchestrator spawns a single `Agent(model: "opus", effort: "xhigh")` sub-agent (orchestration-mode sequential single-spawn for the deep-reasoning step).
**Isolation**: none (single agent — no parallel writes).
**Purpose**: design-architecture (per the effort taxonomy).

One opus-xhigh sub-agent reasons over the ANALYZE aggregate (domain profile + task-pattern inventory) and:

1. **Selects/combines patterns** from the 6-pattern catalog (§ Pattern Catalog below) based on task signals: parallelism available, adversarial-verification need, supervision depth, expertise diversity.
2. **Defines specialist roles** — one specialist role per distinct responsibility the harness needs.
3. **Maps each specialist to an execution primitive** — `sub-agent` / `dynamic-workflow` / `worktree` / `/goal` / `adversarial-fan-out` (§ Primitive Mapping below).
4. **Decides per-specialist `isolation`** — `worktree` only for conflict-prone parallel generation targeting overlapping paths; `none` otherwise (§ Worktree Policy below).
5. **Assigns per-specialist `effort` + `model`** per the purpose-driven effort taxonomy (read-only-extract = haiku/low; design-architecture = opus/xhigh; etc.).
6. **Drafts the Sprint Contract** — the graded dimensions + thresholds the harness's evaluator (if invoked) will score against.
7. **Emits a draft manifest** — the 8 top-level fields populated, ready for validation against the canonical schema.

**User approval gate (AskUserQuestion) — first-class.** BEFORE GENERATE, the orchestrator surfaces the draft manifest + planned specialists + Sprint Contract to the user via `AskUserQuestion`. The orchestrator (NOT the PLAN sub-agent) holds this boundary. This is the self-contradiction resolution: the gate could NOT fire if the Builder were a dynamic-workflow script (scripts cannot call `AskUserQuestion` mid-run), but the orchestrator-direct Builder makes it reachable.

The gate presents, at minimum:

- The derived harness `<name>` + domain.
- The selected patterns (from the 6-pattern catalog) with a one-line rationale each.
- The specialist roster — each specialist's role, primitive, isolation, effort, model.
- The Sprint Contract dimensions + thresholds.
- Options: **Proceed to GENERATE (Recommended)** / **Revise manifest** (return to PLAN with refinement) / **Abort**.

`ToolSearch(query: "select:AskUserQuestion")` preload is mandatory immediately before the gate call (deferred-tool prerequisite per `.claude/rules/moai/core/askuser-protocol.md`).

### Phase 3 — GENERATE [orchestrator fan-out emits 5 artifact types; conditional Agent(isolation:"worktree")]

**Primitive**: orchestrator-direct fan-out — the orchestrator spawns specialist `Agent()` calls (parallel where independent, sequential where dependent) that emit the 5 artifact types in § GENERATE Output Contract below.
**Isolation**: conditional per-specialist — `worktree` for conflict-prone parallel generation, `none` otherwise.

On Proceed, the orchestrator fans out specialist agents that emit the 5 artifacts. The spawn form is decided per-specialist by consulting each specialist's `isolation` field in the manifest (the field the PLAN phase populated via the isolation-decision helper):

- **`isolation: "worktree"`** → the orchestrator spawns `Agent(role, isolation:"worktree", ...)` for that specialist. This is the conflict-prone case: ≥2 specialists targeting overlapping paths, OR a specialist making a risky change (shared-infrastructure touch). The worktree isolates the blast radius so parallel writes do not collide.
- **`isolation: "none"`** → the orchestrator spawns a plain `Agent(role, ...)` (main-tree). This is the read-only or sequential case: no write conflict is possible, so no worktree is created.

The isolation-decision helper (`DecideIsolation` in the manifest Go package) codifies the rule the PLAN phase consults: read-only → `none`; risky change → `worktree`; parallel + overlapping paths → `worktree`; otherwise → `none`. The decision is **advisory**: per the L1 worktree autonomy policy, the orchestrator does NOT mandate worktree creation — `Agent(isolation:"worktree")` is runtime-autonomous, and the Claude Code runtime decides whether to materialize the worktree. The harness logic recommends; the runtime materializes.

There is NO mandatory top-level worktree wrapping the entire Builder. Worktree isolation is sub-agent-granular and conditional per-specialist only. ANALYZE (read-only fan-out) always uses `isolation: "none"` — zero worktrees are created for read-only analysis.

The Runner mirrors the same conditional on the execution side: at end-of-run it emits a worktree-cleanup directive ONLY when ≥1 specialist declared `isolation: "worktree"`. L1 worktree cleanup itself is runtime-autonomous; the Runner emits the directive, the runtime performs the cleanup.

The GENERATE output contract (§ GENERATE Output Contract below) is the handoff spec the manifest schema + Runner engine + command generation + lifecycle consume.

### Phase 4 — ACTIVATE [orchestrator-direct dry-run + /goal; A/B optional]

**Primitive**: orchestrator-direct — dry-run sub-agent + `/goal` (autonomous convergence) + optional with/without A/B. The orchestrator drives this phase directly.
**Isolation**: none (default) or `worktree` (if the sample task is risky).

Run a sample task dry-run through the newly built harness. Use `/goal "<harness> completes task X at quality Y"` for autonomous convergence (per `.claude/rules/moai/workflow/goal-directive.md`). The `/goal` evaluator (a small fast model) checks after each turn whether the harness has demonstrably completed the sample task.

**Reference-integrity smoke gate [HARD].** Before declaring the harness active, the ACTIVATE phase MUST run the reference-integrity smoke gate (`moai harness doctor`). The gate verifies the entry command, manifest, Runner (including its MANIFEST_PATH constant), and every specialist agent the Runner references all resolve to existing artifacts. If the gate reports any ERROR-severity finding, the workflow MUST NOT declare the harness active (do NOT expose `/harness:<name>`); regress to GENERATE and repair the flagged references first. A harness is exposed only after the smoke gate reports zero ERROR findings.

**With/without A/B (optional).** Run the same sample task once WITH the harness and once WITHOUT (baseline), then compare. This is the Anthropic Generator-Evaluator separation pattern. Pass → expose the harness via `/harness:<name>` (after the smoke gate above is clean). Fail → regress to GENERATE for refinement.

**Load-bearing minimum — A/B is SKIPPED for tasks within the model's solo reliable range.** For a simple, well-bounded harness (e.g., single-skill generation with no adversarial-verification need), the A/B evaluation adds cost without value. The orchestrator skips the A/B and records the skip with rationale ("task within solo range, evaluator skipped per simplest-solution-first"). The evaluator is invoked ONLY when the task exceeds the model's solo range.

## Pattern Catalog (6 patterns, dynamically selected by PLAN)

The PLAN phase selects ≥1 pattern per harness from this catalog. The selection is driven by task signals (NOT fixed) — different domain requests yield different pattern selections. The selected patterns are recorded in `manifest.json.patterns`.

| Pattern | When selected | Execution-primitive affinity |
|---------|---------------|------------------------------|
| **Pipeline** | Sequential transformation stages (output of stage N = input of stage N+1) | sub-agent chain |
| **Fan-out/Fan-in** | Parallel independent work + aggregation | dynamic-workflow |
| **Expert Pool** | Diverse expertise consulted in parallel | dynamic-workflow |
| **Producer-Reviewer** | Generator-Evaluator separation (adversarial verification) | adversarial-fan-out |
| **Supervisor** | Centralized coordination + dispatch | sub-agent (supervisor) + workers |
| **Hierarchical Delegation** | Multi-level delegation (manifest-level, NOT recursive subagent) | sub-agent per level |

**Selection guidance (PLAN uses these heuristics):**

- **Sequential transformation** (stage N feeds stage N+1) → Pipeline. Example: a doc-generation harness (extract → transform → render).
- **Parallel independent work + aggregation** → Fan-out/Fan-in. Example: a research harness (N independent searches aggregated into one report).
- **Diverse expertise consulted in parallel** → Expert Pool. Example: a code-review harness (security + perf + arch reviewers in parallel).
- **Generator-Evaluator separation needed** → Producer-Reviewer. Example: a design-quality harness (builder + adversarial evaluator).
- **Centralized coordination + dispatch** → Supervisor. Example: a migration harness (one coordinator dispatches per-file workers).
- **Multi-level delegation** → Hierarchical Delegation. Example: a large-scope harness (top-level breaks work into sub-scopes, each sub-scope has its own lead).

**Selection is NOT fixed.** Two different domain requests MUST be able to yield different pattern selections. A research harness typically selects Fan-out/Fan-in + Expert Pool; a code-review harness typically selects Producer-Reviewer + Pipeline. The PLAN sub-agent's job is to match the pattern to the task signals, not to default to a fixed combination.

**Combination is allowed.** A single harness MAY combine patterns — e.g., Pipeline (overall flow) + Producer-Reviewer (one stage is adversarial). The manifest's `patterns` array carries ≥1 entry; combinations are recorded as multiple entries.

## Primitive Mapping (specialist → execution primitive)

Each specialist in the manifest has a `primitive` field set to exactly one of 5 values. The Runner (M3) consumes this verbatim — it does NOT re-derive the primitive from heuristics.

| Primitive | Runner dispatches as | When PLAN assigns it |
|-----------|---------------------|----------------------|
| `sub-agent` | `Agent(role, effort, model)` — ordinary sub-agent | Default for most single-task specialists |
| `dynamic-workflow` | dynamic-workflow `agent()` call | High-volume parallel independent work (Fan-out/Fan-in, Expert Pool) |
| `worktree` | `Agent(role, isolation:"worktree", ...)` | Conflict-prone parallel generation targeting overlapping paths |
| `/goal` | `/goal` autonomous-convergence directive | Long-running convergence on a verifiable end-state |
| `adversarial-fan-out` | Producer + Reviewer fan-out with confidence scoring | Generator-Evaluator separation (Producer-Reviewer pattern) |

**PLAN's mapping discipline:**

- Every specialist gets exactly one primitive — no free-text, no "auto".
- The primitive must match the pattern it serves (e.g., a Producer-Reviewer specialist is `adversarial-fan-out`, not `sub-agent`).
- `worktree` is assigned ONLY when the specialist's write-targets overlap with another specialist's — read-only or sequential specialists get `sub-agent` or `dynamic-workflow` with `isolation: none`.

## Worktree Policy (conditional, sub-agent-granular)

NO mandatory top-level worktree wraps the Builder or Runner. `Agent(isolation:"worktree")` is spawned ONLY for specialists whose manifest declares `isolation: worktree`. This corrects the earlier mistake of always creating a top-level worktree, which polluted the worktree registry and added cleanup overhead for read-only phases where no write conflict is possible.

| Phase | Default isolation | Rationale |
|-------|-------------------|-----------|
| ANALYZE | none (read-only fan-out) | Read-only; no write conflicts possible |
| PLAN | none (single sub-agent) | Single agent; no parallel writes |
| GENERATE | conditional per-specialist | Conflict-prone parallel generation → worktree; else main-tree |
| ACTIVATE | none (or worktree if sample task risky) | Dry-run; risky → isolate |

**L1 worktree cleanup is runtime-autonomous.** The Runner emits a cleanup directive at end-of-run; the actual `git worktree prune` is handled by the Claude Code runtime, not by the harness logic.

## GENERATE Output Contract (the 5 artifact types)

The GENERATE phase emits exactly 5 artifact types. This contract is the handoff spec M3 (manifest + Runner engine) and M4 (command generation + lifecycle) consume. Each artifact has a fixed location and a content contract.

### Artifact 1 — Thin-wrapper entry command

**Path**: `.claude/commands/harness/<name>.md`
**Purpose**: makes `/harness:<name>` resolve to this harness's Runner. Claude Code subdirectory-command resolution maps `commands/harness/<name>.md` → `/harness:<name>`.
**Content contract**:

- YAML frontmatter: `description` (one-sentence), `argument-hint`, `allowed-tools: Skill`.
- Body: a thin routing wrapper (under 20 LOC) that dispatches to the harness's Runner Workflow. No workflow logic inline.
- The command MUST reference the Runner Workflow filename (`harness-<name>-run.js`).

### Artifact 2 — Runner Workflow

**Path**: `.claude/workflows/harness-<name>-run.js`
**Purpose**: the dynamic-workflow script that consumes `manifest.json` and dispatches specialists per their declared `primitive`.
**Content contract**:

- Reads `manifest.json` as its single source of truth (exactly one config-read path).
- For each `specialist` in `manifest.specialists`: dispatches per `specialist.primitive` (sub-agent / dynamic-workflow / worktree / `/goal` / adversarial-fan-out) — verbatim, no re-derivation.
- Applies the Sprint Contract (if the evaluator is not skipped).
- Emits a worktree-cleanup directive at end-of-run if any specialist declared `isolation: worktree`.
- **Determinism**: the script body MUST NOT call `Date.now()` or `Math.random()`. Timestamps are injected via script input arguments or stamped onto results after the run returns (per `.claude/rules/moai/workflow/dynamic-workflows.md` determinism constraint).

### Artifact 3 — Specialist sub-agent definitions

**Path**: `.claude/agents/harness/harness-<name>-*-specialist.md`
**Purpose**: one sub-agent definition per specialist role.
**Content contract**:

- YAML frontmatter: `name`, `description`, `tools` (the tools the specialist needs).
- Body: the specialist's responsibility, its inputs/outputs, its quality bar, and any domain-specific guidance.
- The specialist's `effort` / `model` / `isolation` / `primitive` live in the manifest (NOT duplicated in the agent frontmatter) — the manifest is the SSOT for dispatch, the agent file is the SSOT for the specialist's reasoning.

### Artifact 4 — Companion Progressive-Disclosure skills

**Path**: `.claude/skills/harness-<name>-*/SKILL.md`
**Purpose**: companion skills that carry domain-specific guidance the specialists and the Runner reference.
**Content contract**:

- YAML frontmatter per the skill-authoring schema (description, paths, metadata, triggers).
- Body: progressive-disclosure structure (Quick Reference / Implementation Guide / Advanced) per skill-writing-craft.
- The skill namespace is `harness-<name>-*` (user-owned per the Skills Namespace Policy — `moai update` preserves it).

### Artifact 5 — manifest.json (the SSOT)

**Path**: `.claude/commands/harness/<name>/manifest.json` (co-located with the entry command, in a per-harness subdirectory) — the single canonical manifest location, recorded in the entry command for the Runner to find. This location matches the `moai harness` Go lifecycle norm; no alternate location is used.
**Purpose**: the single source of truth the Runner reads for its dispatch logic.
**Content contract** (8 top-level fields, per the canonical manifest schema):

- `name` — the harness name (kebab-case, DNS-safe, matches `/harness:<name>` and `harness-<name>-run.js`).
- `domain` — short human-readable domain description.
- `source_request` — the original natural-language request verbatim (carried from the entry workflow's confirmed profile).
- `patterns` — array of pattern names from the 6-pattern catalog (≥1 entry).
- `specialists` — non-empty array of specialist objects, each with `role`, `primitive`, `isolation`, `effort`, `model`.
- `sprint_contract` — object with `dimensions` (array) and `thresholds` (map).
- `entry_command` — the `/harness:<name>` string.
- `runner_workflow` — the `harness-<name>-run.js` filename.

The manifest is validated against the canonical schema before GENERATE completes. A manifest failing validation regresses to PLAN for correction.

## AskUserQuestion Boundary (orchestrator-side)

The Builder phases are orchestrator-direct. The skill body DESCRIBES that the orchestrator calls `AskUserQuestion` at the PLAN→GENERATE gate. The skill body itself MUST NOT contain an `AskUserQuestion(...)` invocation — skills are not the orchestrator. The approval gate is orchestrator behavior documented in prose here; the orchestrator implements it at runtime.

Subagents reachable from the Builder (Explore agents in ANALYZE, the opus-xhigh agent in PLAN, specialist agents in GENERATE) MUST NOT invoke `AskUserQuestion` — they return structured blocker reports and the orchestrator re-runs the round, per the asymmetric boundary in `.claude/rules/moai/core/agent-common-protocol.md` § User Interaction Boundary.

## Cross-references

- Entry workflow: `${CLAUDE_SKILL_DIR}/workflows/harness-build-entry.md` (Phases 0-3: Discovery, name derivation, approval gate).
- moai SKILL.md § harness Branch B (dispatcher routing).
- Manifest schema: companion design document § manifest schema section.
- Runner contract: companion design document § Runner section.
- Dynamic-workflow primitive: `.claude/rules/moai/workflow/dynamic-workflows.md`.
- Orchestration mode selection: `.claude/rules/moai/workflow/orchestration-mode-selection.md`.
- AskUserQuestion canonical: `.claude/rules/moai/core/askuser-protocol.md`.
- Orchestrator-subagent boundary: `.claude/rules/moai/core/agent-common-protocol.md` § User Interaction Boundary.
- `/goal` autonomous convergence: `.claude/rules/moai/workflow/goal-directive.md`.
- Conditional worktree isolation: `.claude/rules/moai/workflow/worktree-integration.md`.
- 5-Section Evidence-Bearing Report Format: `.claude/rules/moai/core/verification-claim-integrity.md`.
