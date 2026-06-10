---
description: "Six architectural agent orchestration patterns for MoAI-ADK agent design"
paths: ".claude/agents/**/*.md,.claude/rules/moai/development/agent-authoring.md"
---

<!-- Source: revfactory/harness — Apache License 2.0 — see .claude/rules/moai/NOTICE.md -->

# Agent Design Patterns

Six foundational orchestration patterns for designing agent systems in MoAI-ADK. These patterns describe how agents interact, communicate, and coordinate work.

## Pattern 1: Pipeline

Sequential execution where each agent's output becomes the next agent's input.

**Structure**: Agent A → Agent B → Agent C → Agent D

**When to Use**: 
- Requirements are strongly dependent on previous stage outputs
- Each stage has clear input/output contracts
- Linear workflow with minimal branching

**Example**: Content creation pipeline — research → outline → draft → edit

**Anti-Pattern**: Using pipeline when stages are independent (wastes sequential overhead)

---

## Pattern 2: Fan-out/Fan-in

Parallel processing with independent analysis followed by result aggregation.

**Structure**: 
```
         ┌→ Agent A ┐
Dispatch ┼→ Agent B ┼→ Synthesize
         └→ Agent C ┘
```

**When to Use**:
- Same input needs multiple independent perspectives
- Parallel analysis improves coverage
- Results can be meaningfully combined

**Example**: Multi-angle research — technical + market + user perspectives analyzed in parallel, then synthesized

**Anti-Pattern**: Over-specifying how to combine results; let agents negotiate

---

## Pattern 3: Expert Pool

Dynamic selection of appropriate specialist for each task.

**Structure**: Router → { Expert A | Expert B | Expert C }

**When to Use**:
- Input type determines which expert to invoke
- Not all experts needed for each task
- Expert selection is well-defined

**Example**: Code review routing — security expert for auth code, performance expert for queries, architecture expert for design

**Anti-Pattern**: Always invoking all experts even when only one is relevant

---

## Pattern 4: Producer-Reviewer

Iterative quality cycle with specialized producer and reviewer agents.

**Structure**: Producer → Reviewer → (Feedback loop) → Producer redo → Reviewer

**When to Use**:
- Quality validation has clear objective criteria
- Feedback can guide refinement
- Bounded iteration count prevents infinite loops

**Example**: Content generation with review — writer creates → editor reviews → writer revises (max 3 iterations)

**Anti-Pattern**: Unbounded review cycles; always set maximum iterations

---

## Pattern 5: Supervisor

Central coordinator agent manages work distribution to subordinate agents.

**Structure**:
```
         ┌→ Worker A
Supervisor ┼→ Worker B  (Dynamic assignment)
         └→ Worker C
```

**When to Use**:
- Work volume or complexity is variable
- Tasks need dynamic distribution
- Coordination overhead is acceptable

**Example**: Large codebase migration — supervisor analyzes files, delegates modules to workers, tracks progress

**Anti-Pattern**: Supervisor becomes a bottleneck; use for high-level coordination only

---

## Pattern 6: Hierarchical Delegation

Multi-level delegation where agents recursively decompose problems.

**Structure**:
```
[Executive] → [Manager A] → [Specialist A1]
                           → [Specialist A2]
           → [Manager B] → [Specialist B1]
```

**When to Use**:
- Problem naturally decomposes hierarchically
- Clear team boundaries exist
- Depth is limited (max 2-3 levels)

**Example**: Full-stack development — lead → backend team lead → (database + API + tests) + frontend team lead → (UI + state management)

**Anti-Pattern**: Nesting more than 2-3 levels; causes coordination overhead and context loss

---

## Pattern Selection Matrix

| Scenario | Pattern | Why |
|----------|---------|-----|
| Linear requirements | Pipeline | Clear ordering |
| Parallel analysis | Fan-out/Fan-in | Independent perspectives |
| Variable expertise | Expert Pool | Right tool per task |
| Quality iteration | Producer-Reviewer | Focused refinement loop |
| Variable workload | Supervisor | Dynamic allocation |
| Complex decomposition | Hierarchical | Natural boundaries |

---

## Anti-Patterns to Avoid

**Anti-Pattern 1**: "Pipeline everything"
- Using sequential patterns when tasks are independent
- Fix: Use Fan-out/Fan-in for parallel work

**Anti-Pattern 2**: "All experts always"
- Invoking every expert for every task
- Fix: Route to relevant experts only via Expert Pool

**Anti-Pattern 3**: "Unbounded review"
- Producer-Reviewer without iteration limits
- Fix: Set `max_iterations: 3` to prevent infinite loops

**Anti-Pattern 4**: "Supervisor bottleneck"
- Central coordinator doing all coordination work
- Fix: Use shared task lists for self-coordination

**Anti-Pattern 5**: "Deep hierarchies"
- Nesting more than 2-3 delegation levels
- Fix: Flatten structure or switch to Supervisor pattern

---

## Combination Strategies

Real-world systems often combine patterns:

| Combination | Structure | Use Case |
|-------------|-----------|----------|
| Pipeline + Fan-out | Sequential stages with parallel substeps | Analysis (gather data) → Design → Implement (parallel experts) |
| Fan-out/Fan-in + Producer-Reviewer | Parallel analysis then iterative refinement | Research (parallel) → Synthesize → Review/improve |
| Supervisor + Expert Pool | Coordinator routes to specialists | Large project with dynamic workload distribution |

The key is matching pattern structure to actual workflow dependencies, not forcing dependencies into convenient patterns.

---

## MoAI-ADK Pattern Vocabulary

MoAI-ADK applies the six generic harness patterns using these specialized naming conventions:

### Team
Maps to **Fan-out/Fan-in**. Multiple agents analyze independently, results synthesized. Use when one request needs multiple perspectives: security review + performance audit + architecture check all run in parallel, then consolidated.

### Sub-agent
Maps to **Pipeline**. Sequential delegation where Agent A's output becomes Agent B's input. Use when work has clear dependencies: design spec → backend implementation → frontend integration → testing.

### Hybrid
Maps to **Supervisor** + **Expert Pool**. Central orchestrator routes tasks dynamically to specialists based on task type. Use when workload is heterogeneous: same orchestrator routes code review to security expert OR architecture expert depending on input.

### Orchestrator
Maps to **Hierarchical Delegation**. Multi-level decomposition where the orchestrator (MoAI) delegates to team leads, who delegate to specialists. Use for large projects with natural team boundaries: MoAI → backend lead → (database team + API team) + frontend lead → (UI team).

### Specialist
Maps to **Expert Pool** singleton. A single expert agent handles all tasks of a specific type without dynamic routing. Use for focused work: one code formatter, one security validator, one documentation generator.

### Pipeline
Generic **Pipeline** pattern. Sequential stages with each agent's output feeding the next. Already native to MoAI workflows. Use when linear dependency is explicit: requirements → design → implement → test → ship.

---

## Per-Spawn Domain Specialization

Per Anthropic best-practices alignment, domain expertise is canonically delivered through **per-spawn parameter injection**, not through static agent definition files. This replaces the 6 archived `expert-*` agents (backend, frontend, security, devops, performance, refactoring) that were retired during the catalog consolidation because they failed Anthropic's "Define a custom subagent when you keep spawning the same kind of worker with the same instructions" criterion.

### Canonical Spawn Pattern

```
Agent(
  subagent_type: "general-purpose",
  model: "opus",
  tools: "<domain-specific tool whitelist as CSV>",
  prompt: "You are a <domain> specialist. <domain-specific instructions tailored to the current task>"
)
```

The orchestrator composes the domain instructions inline at delegation time rather than referencing a pre-authored agent body. Domain knowledge lives in the active conversation context, not trapped inside individual agent files loaded only when explicitly invoked.

### Per-Domain Tool Whitelist Recommendations

| Domain | Recommended tools whitelist |
|--------|------------------------------|
| backend | `Read, Write, Edit, Bash, Grep, Glob` |
| frontend | `Read, Write, Edit, Bash, Grep, Glob` |
| security (review) | `Read, Grep, Bash` (read-only review surface; Write only if patching) |
| devops | `Read, Write, Edit, Bash, Grep, Glob` |
| performance | `Read, Bash, Grep, Glob` (profiling tools via Bash; no Write unless adding benchmarks) |
| refactoring | `Read, Write, Edit, Grep, Glob, Bash` |

### When to Author a Static Agent File Instead

Reserve `.claude/agents/*.md` static files for agents meeting Anthropic's "keep spawning the same kind of worker with the same instructions" criterion. The 7 MoAI-custom retained agents (`manager-spec`, `manager-develop`, `manager-docs`, `manager-git`, `plan-auditor`, `sync-auditor`, `builder-harness`) all satisfy this criterion via recurring SPEC-phase invocations. Domain-specific work that does NOT recur with identical instructions across SPEC sessions belongs in per-spawn pattern, not in a static file.

See `.claude/rules/moai/development/agent-authoring.md` § Static Agent File vs Per-Spawn Specialization Decision Tree for the authoring decision tree.

---

## Read-only Investigation — Explore Canonical Agent

Per Anthropic's published guidance (Claude Code ships exactly 3 built-in subagents: Explore, Plan, general-purpose), the Anthropic built-in `Explore` agent is the canonical read-only investigation surface for the MoAI orchestrator. This replaces the archived `claude-code-guide` and `researcher` custom agents (both retired during catalog consolidation for 0-invocation phantom-agent failure).

### Canonical Spawn Pattern

```
Agent(subagent_type: "Explore", prompt: "<investigation task description>")
```

`Explore` is Anthropic-shipped and not a MoAI file; it lives in the Claude Code runtime. No customization is performed.

### Use Cases

- Codebase exploration (find files matching patterns, summarize package structure)
- Upstream documentation investigation (fetch and analyze Claude Code release notes, library API surface)
- Auto-research (gather context on unfamiliar libraries, APIs, or domain concepts)
- Pre-flight reconnaissance before delegating implementation work (build a context summary the implementation agent consumes)

### Not for

- Implementation work — `Explore` is read-only; route Write/Edit work to `manager-develop` or a per-spawn `Agent(general-purpose, ...)` specialist
- SPEC body authoring — that is `manager-spec`'s scope
- Sync-phase documentation — that is `manager-docs`'s scope

---

## Deprecated: Hierarchical Manager Chain Pattern

The `manager-strategy → manager-develop` hierarchical chain pattern (previously documented as a viable Pattern 6 Hierarchical Delegation variant for MoAI workflows) is **deprecated** under the current agent catalog policy. Anthropic Sub-agents documentation states verbatim: *"Subagents cannot spawn other subagents. If your workflow requires nested delegation, use Skills or chain subagents from the main conversation."* The MoAI `manager-strategy → manager-develop` chain is therefore architecturally impossible to execute at runtime — the chain pattern was a documentation fiction rather than a runtime reality.

All multi-agent coordination MUST happen at the **orchestrator (L1) level** in the main conversation. The retained agent catalog respects this constraint: all retained sub-agents are leaves of the L1 orchestrator, never branches that recurse into further sub-agent spawns.

For migration from `manager-strategy` references in existing code, paste-ready resumes, or documentation, see `.claude/rules/moai/workflow/archived-agent-rejection.md` § Migration Table — the canonical replacement for the planning role absorbed by `manager-strategy` is `manager-spec` (planning IS strategy per Anthropic 4-phase Step 2).
