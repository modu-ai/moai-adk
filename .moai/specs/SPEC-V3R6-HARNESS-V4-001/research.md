# research.md — SPEC-V3R6-HARNESS-V4-001 Research Findings

> This file documents the upstream research grounding the v4 harness redesign. All sources are cited verbatim where load-bearing; the v4 design decisions in `design.md` trace back to findings here.

## §A. revfactory/harness 7-Phase Analysis

**Source**: `github.com/revfactory/harness` (Apache License 2.0).
**moai-adk adaptation**: `.claude/skills/moai-meta-harness/SKILL.md` (the current harness generator, adapted from revfactory's 7-Phase workflow).
**Attribution**: Per `.claude/rules/moai/NOTICE.md`, moai-adk incorporates revfactory/harness reference documents under Apache 2.0; the v4 redesign is an independent derivation that absorbs strengths and removes the rest.

### §A.1 What revfactory/harness Is

revfactory/harness is an "L3 Meta-Factory / Team-Architecture Factory": given a domain sentence, it produces an agent team + companion skills. It defines 6 architectural patterns (Pipeline, Fan-out/Fan-in, Expert Pool, Producer-Reviewer, Supervisor, Hierarchical Delegation) and a 6-Phase workflow (Domain Analysis → Team Architecture Design → Agent Definition Generation → Skill Generation → Integration & Orchestration → Validation & Testing), extended to 7 with an Iteration/LEARNING phase in the moai-adk adaptation.

**Validation claim (upstream)**: A/B-validated at +60% avg quality (49.5 → 79.3), 15/15 win, −32% variance; n=15, author-measured, 3rd-party replication pending. This claim is cited as upstream provenance; the v4 dogfooding validation (REQ-HV4-012) re-runs its own A/B and discloses its result independently — it does NOT co-opt the upstream claim as a verified v4 claim.

### §A.2 Strengths to ABSORB (v4 keeps these)

| Strength | v4 incorporation |
|----------|------------------|
| (a) Domain-sentence → team-architecture auto-generation | REQ-HV4-001 (NL-analysis entry), PLAN phase (§D.2 of design.md) |
| (b) 6-patterns catalog (selected dynamically per task) | REQ-HV4-004, design.md §E (patterns selected by PLAN, NOT statically fixed) |
| (c) Progressive-Disclosure skill generation | GENERATE phase emits `harness-<name>-*/SKILL.md` per PD spec |
| (d) Generator-Evaluator separation + Sprint Contract | REQ-HV4-008, design.md §D.4, manifest `sprint_contract` field |
| (e) With/without A/B validation | ACTIVATE phase (REQ-HV4-003), REQ-HV4-012 (dogfooding) |

### §A.3 To REMOVE (over-engineering for 2026)

| Liability | Why removed | v4 alternative |
|-----------|-------------|----------------|
| (1) Static fixed 7-Phase pipeline | Runs Phase 1→7 unconditionally regardless of task shape; violates Anthropic "simplest solution first" | REQ-HV4-003: phases synthesized from task signals (load-bearing minimum); REQ-HV4-009: no static 7-Phase |
| (2) Hard Agent-Teams dependency (3-5 cap) | Anthropic guidance caps Agent-Teams at 3-5; too narrow for large parallel work | REQ-HV4-005: dynamic-workflow (16 concurrent / 1000 total) as the large-parallel primitive |
| (3) Skeleton ↔ Customization 2-phase split | Splits one logical fan-out into two sequential phases | REQ-HV4-009: single GENERATE fan-out collapses both |
| (4) LEARNING as a separate phase | Learning is cross-cutting, not a phase | REQ-HV4-009: folded into Sprint Contract dimensions |
| (5) revfactory-specific plugin/marketplace assumptions | Do not map to moai-adk-native primitives | REQ-HV4-005: execution-primitive set {sub-agent, dynamic-workflow, worktree, `/goal`, adversarial-fan-out} |
| (6) Socratic-only discovery (no codebase/doc analysis) | Interviews user but never reads actual codebase; produces harnesses that contradict conventions | REQ-HV4-001 + ANALYZE phase (design.md §D.1): mandatory fan-out over codebase + docs + SPEC history |

## §B. Anthropic Harness Design for Long-Running Application Development

**Source**: `anthropic.com/engineering/harness-design-long-running-apps` (2026).
**Loading note**: the verbatim phrases below are reproduced under fair-use academic-attribution conventions; no source code is incorporated.

### §B.1 Key Findings (verbatim where load-bearing)

1. **GAN-inspired Planner-Generator-Evaluator architecture**. The harness design draws on Generative Adversarial Network structure: a Planner decomposes, a Generator produces, an Evaluator scores.

2. **Self-evaluation bias is the core problem** (verbatim intent): separating the generator from the evaluator is the strongest single lever against self-evaluation bias. A generator that also evaluates itself produces sycophantic self-assessment.

3. **Sprint Contract** (verbatim): a pre-coding agreement that converts vague goals into graded dimensions agreed before code is written. The Sprint Contract is the mechanism that makes the evaluator's judgment non-arbitrary.

4. **"Find the simplest solution possible, only increase complexity when needed"** (verbatim). This is the load-bearing design principle for v4's load-bearing-minimum phase synthesis (REQ-HV4-003) and the conditional evaluator (REQ-HV4-008).

5. **As models improve, remove load-bearing components one at a time and measure.** (verbatim intent): context resets, sprint constructs, even the evaluator should be removable one at a time, measuring each removal. The harness is NOT a fixed artifact; it evolves with model capability.

6. **The evaluator earns its cost only when the task exceeds what the model does reliably solo** (verbatim intent). For tasks within the model's solo range, the evaluator is pure overhead. This grounds v4's conditional evaluator (REQ-HV4-008, AC-HV4-008b).

7. **"The space of interesting harness combinations doesn't shrink as models improve, it moves."** (verbatim). The harness design space is not a fixed target that models asymptotically make irrelevant; it shifts. v4's design ameniability to component removal (C-HV4-001) operationalizes this.

### §B.2 v4 Incorporation

- Generator-Evaluator separation → REQ-HV4-008, manifest `sprint_contract`, ACTIVATE adversarial A/B.
- Sprint Contract → manifest `sprint_contract` field (design.md §C).
- Simplest-solution-first → REQ-HV4-003 phase synthesis, REQ-HV4-008 conditional evaluator, C-HV4-001.
- Component-removability → C-HV4-001 (forward-link `SPEC-V3R6-HARNESS-COMPONENT-PRUNE-001`).

## §C. Claude Code Dynamic Workflows

**Source**: `code.claude.com/docs/en/workflows` (canonical Claude Code documentation).
**Loading note**: referenced per `.claude/rules/moai/workflow/dynamic-workflows.md` (the moai-adk canonical dynamic-workflow rule).

### §C.1 Key Findings

1. **Three orchestration primitives** (per moai-adk's dynamic-workflows.md consolidation):
   - **Subagents** (`Agent()`): Claude decides next step turn-by-turn; intermediate results in Claude's context.
   - **Agent Teams**: Claude + teammates via shared TaskList; each teammate has its own context.
   - **Dynamic Workflows**: a JavaScript script the runtime executes; the script holds the plan; intermediate results live in script variables; only the final answer returns to the session.

2. **Scale**: up to **16 concurrent agents** / **1000 total agents per run** (runaway-loop backstop).

3. **No mid-run user input** — only agent permission prompts can pause a run. For sign-off between stages, run each stage as its own workflow. (This grounds C-HV4-004 AskUserQuestion boundary.)

4. **Script determinism**: the script body MUST be deterministic — no `Date.now()` / `Math.random()` calls; resume caching keys on deterministic outputs. (This grounds C-HV4-003.)

5. **`/deep-research` is the canonical fan-out model** (bundled workflow that fans out web searches, cross-checks sources, votes on claims, returns a cited report).

6. **Purpose-driven model/effort selection** (per `SPEC-V3R6-WORKFLOW-EFFORT-MAP-001` taxonomy): `agent()` accepts `{model, effort, agentType, isolation, phase, schema, label}`. Read-only-extract → haiku/low; mechanical-transform → sonnet/medium; synthesize → sonnet/high; verify-judge → sonnet or opus/xhigh; implement → sonnet or opus/xhigh; design-architecture → opus/xhigh.

### §C.2 v4 Incorporation

- Builder as dynamic-workflow (script holds plan) → REQ-HV4-003, design.md §D.
- ANALYZE as Explore fan-out → design.md §D.1 (mirrors `codemaps-extract.js` precedent).
- PLAN as opus xhigh sub-agent → design.md §D.2 (design-architecture purpose).
- GENERATE as dynamic-workflow fan-out → design.md §D.3.
- Conditional `Agent(isolation:"worktree")` → REQ-HV4-007, design.md §G.
- Effort/model per specialist → manifest `specialists[].effort` and `.model` fields.

## §D. Sprint Contracts

**Source**: `agentpatterns.ai/agent-design/sprint-contracts` (Sprint Contract pattern reference).

### §D.1 Key Finding

A Sprint Contract is a pre-coding agreement that converts vague goals into graded dimensions agreed before code is written. The evaluator's judgment is then non-arbitrary: it scores against the pre-agreed dimensions.

### §D.2 v4 Incorporation

- manifest `sprint_contract` field (design.md §C): `{dimensions: [...], thresholds: {...}}`.
- The Runner applies the Sprint Contract per REQ-HV4-008; the evaluator scores against `dimensions` using `thresholds`.

## §E. Meta-Harness (Outer-Loop Optimization)

**Source**: `arxiv 2603.28052` (Meta-Harness: outer-loop optimization of harness code itself).

### §E.1 Key Finding

Meta-Harness proposes an outer loop that optimizes the harness code itself — the harness learns across runs. This is distinct from the inner loop (Generator-Evaluator within one run).

### §E.2 v4 Incorporation (forward-link only)

The Meta-Harness outer loop is declared as a cross-cutting concern (folded into Sprint Contract dimensions) but its internal optimization algorithm is OUT OF SCOPE for this SPEC (spec.md §B.2). Forward-link: candidate `SPEC-V3R6-HARNESS-LEARNING-LOOP-001`. This research.md cites Meta-Harness to establish that the learning-subsystem design space exists and to bound this SPEC's scope claim.

## §F. Prior moai-adk Baselines (cross-reference, not re-research)

| SPEC | Status | Relevance to v4 |
|------|--------|-----------------|
| `SPEC-V3R6-HARNESS-NAMESPACE-V2-001` | completed | §24 namespace protection baseline; v4 extends to `commands/harness/` + `workflows/harness-*.js` |
| `SPEC-V3R6-LIFECYCLE-REDESIGN-001` | completed | 3-phase plan/run/sync; MX Tag cross-cutting (NOT separate phase); v4's 4-phase Builder is internal, not lifecycle |
| `SPEC-V3R6-WORKFLOW-EFFORT-MAP-001` | completed | dynamic-workflow `agent()` purpose-driven model/effort taxonomy; v4 manifest `effort`/`model` fields consume this directly |
| `SPEC-V3R6-HARNESS-MOAI-NAMESPACE-001` | superseded | policy-reversal rejected (moai-harness-* remains template-managed); v4 honors the V2-001 resolution |
| `SPEC-V3R6-AGENT-TEAM-REBUILD-001` | completed | 8-agent retained catalog; harness specialists are NOT part of the retained catalog (separate user-owned namespace) |
| `SPEC-V3R3-HARNESS-001` | (legacy) | original moai-meta-harness adaptation from revfactory 7-Phase; v4 supersedes the 7-Phase path |

## §G. Research Gaps (acknowledged)

- **GAP-001**: Claude Code subdirectory-command resolution (`.claude/commands/harness/<name>.md` → `/harness:<name>`) is to be verified in M1 pre-flight (BI-001). If subdirectory commands do not resolve as expected, the fallback is flat-file `.claude/commands/harness-<name>.md` (which resolves to `/harness-<name>`).
- **GAP-002**: The dynamic-workflow runtime version requirement (Claude Code v2.1.154+) and paid-plan gating need environment verification at run-phase entry. If unavailable, the Builder falls back to sequential sub-agent mode (R-HV4-001).
- **GAP-003**: The +60% A/B claim is upstream (revfactory, author-measured, n=15). The v4 dogfooding re-runs its own A/B and discloses independently — it does NOT inherit the upstream claim's statistical weight.

## §H. Cross-References

- `spec.md` §A (problem statement) — traces to §A.3 above.
- `spec.md` §C (REQs) — each REQ traces to a finding above.
- `design.md` — the architecture formalizing the v4 decisions grounded here.
- `.claude/rules/moai/NOTICE.md` — Apache 2.0 attribution for revfactory/harness imported components.
- `.claude/rules/moai/workflow/dynamic-workflows.md` — canonical dynamic-workflow rule (purpose-driven effort taxonomy, determinism, primitive selection).
- `.claude/skills/moai-meta-harness/SKILL.md` — legacy 7-Phase generator (becomes redirect target per REQ-HV4-013).
