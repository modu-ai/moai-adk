---
description: "Orchestrator templates for task coordination in MoAI-ADK workflows"
paths: ".claude/rules/moai/core/moai-constitution.md,CLAUDE.md"
---

<!-- Source: revfactory/harness — Apache License 2.0 — see .claude/rules/moai/NOTICE.md -->

# Orchestrator Templates

Orchestration patterns for the MoAI orchestrator when coordinating sub-agents. Each template maps onto a mode in the Phase 4 execution-mode catalog (`.claude/rules/moai/workflow/orchestration-mode-selection.md` §A) — the catalog decides WHICH mode; this file describes HOW to run it.

| Template | Catalog mode | Shape |
|----------|--------------|-------|
| Sub-orchestrator | serial (`sub-agent`) | Sequential handoff, one agent per milestone — **the default** |
| Fan-out orchestrator | fanout (`parallel`) | 3-5 concurrent `Agent()` calls in one turn (advisory — hard bound is the runtime subagent cap, default 20), results return to the orchestrator |
| Hybrid orchestrator | fanout + serial | Sequential stages with a parallel stage in the middle |
| Workflow orchestrator | sweep (`workflow`) | Script-held plan, dozens of agents — see `.claude/rules/moai/workflow/dynamic-workflows.md` |
| Team-orchestrator | agent-team — **experimental (re-allowed)** | See § Team-orchestrator — experimental below |

---

## Sub-orchestrator (serial — default)

**When to Use**:
- Sequential task handoff (output A → input B)
- Coding-heavy work (most coding tasks decompose into few truly parallel subtasks)
- Any case where a simpler mode suffices — this is the default fallback

**Structure**:
```
User Request
    ↓
[MoAI] → Delegate to Agent A
         ↓ Result
         Delegate to Agent B (with A's output)
         ↓ Result
         Delegate to Agent C
         ↓ Result
         Consolidate
         ↓
      User Response
```

**How to Spawn**:
```javascript
// Sequential — domain expertise is injected per-spawn via the prompt,
// not read from a static config profile.
const resultA = Agent(prompt: "Analyze X", subagent_type: "general-purpose")
const resultB = Agent(prompt: `Design based on: ${resultA}`, subagent_type: "general-purpose")
const resultC = Agent(prompt: `Implement: ${resultB}`, subagent_type: "general-purpose")

// Consolidate
MoAI: "Here are the results from the pipeline..."
```

**Error Recovery**:
- **Agent A fails**: AskUserQuestion to proceed, skip, or retry
- **Agent B fails**: Provide feedback and retry with adjustment
- **Agent C fails**: Can use partial output from A+B or re-delegate

**Escalation Rules**:
- If 3+ agents fail in sequence → Escalate to user
- If total tokens > 80% budget → Collect results and conclude
- If a stage stalls on a decision → AskUserQuestion, then re-delegate with the answer injected

**Example: Feature Development Pipeline**
```
1. Designer: "Create API spec for feature X"
   → OpenAPI spec

2. Backend: "Implement the API based on spec"
   → API code + tests

3. Frontend: "Create components based on API"
   → UI code + hooks

4. Tester: "Verify API + UI together"
   → Integration test report

5. MoAI: "Feature complete, ready for PR"
```

---

## Fan-out orchestrator (fanout — parallel sub-agents)

**When to Use**:
- Multi-domain, research-heavy work (≥3 domains) where each angle is independent
- Multi-perspective review where each reviewer should form its own judgment
- Any place where parallel wall-clock speed matters and the agents do NOT need to talk to each other

**Advisory band**: 3-5 concurrent `Agent()` calls in a single turn — a cache/coordination advisory (stagger same-type spawns so N−1 read the first spawn's cache), not a hard cap and not the Agent-Teams team-size number; the hard bound is the runtime subagent cap (default 20). Beyond the band, coordination and token cost outrun the benefit.

**Structure**:
```
User Request
    ↓
[MoAI] → Agent A, Agent B, Agent C   (all spawned in ONE turn)
         ↓        ↓        ↓
         Results return to the orchestrator's context
         ↓
      Orchestrator reconciles (contradictions surfaced, not smoothed)
         ↓
      User Response
```

**How to Spawn** — all calls in a single assistant turn:
```javascript
// One turn, multiple Agent() calls → they run concurrently.
Agent(prompt: "Review for security defects: ...", subagent_type: "general-purpose")
Agent(prompt: "Review for performance defects: ...", subagent_type: "general-purpose")
Agent(prompt: "Review for architectural drift: ...", subagent_type: "general-purpose")
```

**Prompt discipline** (each spawned agent needs all four, or coverage duplicates and gaps appear):
1. Objective — what this agent alone is responsible for
2. Output format — the shape the orchestrator will consume
3. Tool guidance — which tools to use and why
4. Boundaries — what this agent must NOT cover (the other lenses' scope)

**Error Recovery**:
- **One agent fails**: filter it out, reconcile the survivors, and name the missing lens in the report — never present partial coverage as complete
- **Agents contradict each other**: surface the contradiction to the user; do not silently pick one
- **All agents fail**: fall back to serial with a narrowed scope

**Escalation Rules**:
- If ≥2 agents report the same blocker → AskUserQuestion for guidance
- If the work turns out to need cross-agent negotiation → it was not fan-out-shaped; re-run as serial, or move it to a sweep where the script holds the plan

**Note**: write fan-out stays foreground and sequential (serial). Parallel spawns are for read-only investigation, research, and review — concurrent writers race on the working tree.

---

## Hybrid orchestrator (fanout + serial)

**When to Use**:
- Mix of sequential and parallel stages
- Research or review benefits from fan-out, but implementation must stay sequential

**Structure**:
```
Stage 1: Sequential (serial)
  Research Agent → Analysis Agent

Stage 2: Parallel fan-out (fanout, read-only)
  Agent(security-lens), Agent(perf-lens), Agent(arch-lens)   ← one turn

Stage 3: Sequential (serial)
  Implementation Agent (writes; foreground, sequential)

Stage 4: Consolidate
  Final result to user
```

**How to Spawn**:
```javascript
// Stage 1: Sequential research
const research = Agent(prompt: "Research X and Y", subagent_type: "general-purpose")

// Stage 2: Parallel read-only review — all in ONE turn
Agent(prompt: `Security lens over: ${research}`, subagent_type: "general-purpose")
Agent(prompt: `Performance lens over: ${research}`, subagent_type: "general-purpose")
Agent(prompt: `Architecture lens over: ${research}`, subagent_type: "general-purpose")

// Stage 3: Sequential implementation (writes — never parallel)
const impl = Agent(prompt: `Implement per the reconciled findings: ...`, subagent_type: "general-purpose")
```

**Error Recovery**:
- **Sequential stage fails**: retry with adjustment, or escalate
- **Parallel stage partially fails**: reconcile survivors, name the missing lens
- **Cascade failure**: collect partial results, present to user with blockers

**Escalation Rules**:
- If any stage fails after 2 retries → AskUserQuestion for decision
- If token budget exceeded → prioritize remaining stages

---

## Workflow orchestrator (sweep) — pointer

When the work fans out over dozens-to-hundreds of independent, mostly-mechanical items, the plan belongs in a script rather than in the orchestrator's context. That is the dynamic-workflow primitive, documented at `.claude/rules/moai/workflow/dynamic-workflows.md` (16 concurrent agents, 1,000-total backstop, intermediate results in script variables, no mid-run user input). It is NOT a template in this file — the script IS the template.

Reserve it for genuinely-parallel high-volume work. Coding-heavy, multi-domain, or new-code work stays serial.

---

## Team-orchestrator — experimental (re-allowed)

**agent-team — experimental (re-allowed; operator decision).** The Agent Teams pattern (a lead agent spawning named teammates that coordinate through a shared task list and peer messages) is again a sanctioned MoAI orchestration pattern, entered ONLY by explicit `--team` request; the Phase 4 decision tree never auto-selects it (constraints and genealogy: `orchestration-mode-selection.md` §C.1).

**Default routing while agent-team stays explicit-request-only**:

| Team use case | Default (auto) |
|-------------------|-----|
| Multi-domain research / multi-perspective review | Fan-out orchestrator (fanout) |
| Coding work across layers | Sub-orchestrator (serial) — sequential |
| Competing-hypothesis debugging | Fan-out orchestrator (fanout), one hypothesis per agent; the orchestrator falsifies |
| Decomposable bulk migration | Workflow orchestrator (sweep) |

The rationale is coordination cost: fan-out plus sequential covers the practical surface at lower token and latency cost than peer-coordinating teammates, which is why the auto-routes stay with fanout/serial/sweep and agent-team requires the explicit request.

**Native runtime sanctioned**: the Claude Code teammate runtime itself (`moai cg` GLM panes, `moai cc -w <name> --spawn` teammate windows, the teammate registry) works and is a sanctioned surface under the enabled flag — the retired era withdrew only MoAI's static orchestration layer, never the runtime primitive.

---

## Decision Matrix: Which Template?

| Scenario | Template | Why |
|----------|----------|-----|
| Single feature, clear handoff | Sub (serial) | No coordination needed, simplest |
| Simple change (style update) | Sub (serial) | Anything more is overkill |
| Coding across several layers | Sub (serial) | Coding rarely decomposes into truly parallel subtasks |
| Multi-domain research (≥3 domains) | Fan-out (fanout) | Independent angles, parallel speed pays off |
| Multi-perspective review | Fan-out (fanout) | Each reviewer forms an independent judgment |
| Bug investigation with rival hypotheses | Fan-out (fanout) | One hypothesis per agent; orchestrator falsifies |
| Large feature: research then build | Hybrid | Fan-out the research, sequence the writes |
| Bulk mechanical migration (dozens of sites) | Workflow (sweep) | The script should hold the plan |

---

## Common Patterns Within Templates

### Fan-out (Parallel Independent Results)
```
[MoAI] → Agent A, Agent B, Agent C (all in ONE turn)
        Reconcile A, B, C — surface contradictions, do not smooth them
```

**When**: Independent analyses that don't depend on each other
**Example**: Concurrent code reviews from security + performance + architecture

### Pipeline (Sequential Handoff)
```
[MoAI] → Agent A → (result) → Agent B → (result) → Agent C
```

**When**: Each stage requires the previous stage's output
**Example**: Research → Design → Implementation → Testing

### Reconcile (the step fan-out cannot skip)
```
[MoAI] ← results from N agents
        → dedupe, rank, and NAME any lens that failed or returned nothing
        → present contradictions as contradictions
```

**When**: Always, after a fan-out. A fan-out whose results are concatenated without reconciliation reports duplicate findings and hides coverage gaps.

---

## Transition Rules

**When to Switch Templates During Execution**:

1. **Sub → Fan-out**: work turns out to be several independent read-only investigations
   - Action: spawn the lenses in one turn, then reconcile
2. **Fan-out → Sub**: agents keep needing each other's intermediate results
   - Action: it was not fan-out-shaped; sequence it
3. **Fan-out → Workflow**: the item count grows past what one turn can coordinate
   - Action: move the plan into a script (sweep)
4. **Hybrid → Simpler**: the parallel stage adds no wall-clock benefit
   - Action: collapse to sequential

**Signals to Transition**:
- **Agents needing each other's output** → sequence them; do not try to make them talk
- **Agents working independently with little overlap** → fan out
- **Over-engineering**: a simple task using a complex template → simplify

---

## Monitoring & Adjustment

**During Execution**:
- Track token consumption against budget
- Watch for a stalled agent (no return) — close the ledger with a synthetic note rather than proceeding as if it returned
- Assess quality of intermediate results before feeding them downstream

**Decision Points**:
- After each major stage: is the approach working? Continue or adjust?
- At 75% token budget: prepare to conclude or escalate
- After a failure: retry with the same template, or switch?

**Adjustment Options**:
- Continue with the current template
- Switch templates mid-stream
- Narrow the scope and re-delegate
- Escalate to user for guidance

These templates cover most MoAI orchestration scenarios. Mix and match as needed — but let the Phase 4 catalog pick the mode, and let the simplest sufficient template win.
