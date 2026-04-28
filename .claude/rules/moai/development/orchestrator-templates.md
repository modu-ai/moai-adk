---
description: "Three orchestrator templates for task coordination in MoAI-ADK workflows"
paths: ".claude/rules/moai/core/moai-constitution.md,CLAUDE.md"
---

<!-- Source: revfactory/harness — Apache License 2.0 — see .claude/rules/moai/NOTICE.md -->

# Orchestrator Templates

Three orchestration patterns for the MoAI orchestrator when coordinating sub-agents and teams.

---

## Team-orchestrator

**When to Use**:
- Multiple agents need to communicate (not just return results)
- Parallel exploration with cross-agent feedback is valuable
- High-coordination complexity (shared task list, dynamic assignment)

**Structure**:
```
User Request
    ↓
[MoAI] → TeamCreate → [Member A, Member B, Member C]
         ↓
      Coordination Loop:
      - SendMessage (member-to-member)
      - TaskCreate/Update (shared list)
      - TeammateIdle hook (members idle when waiting)
      ↓
      Final Result Assembly
      ↓
      User Response
```

**How to Spawn**:
```
Agent(
  subagent_type: "general-purpose",
  mode: "acceptEdits",
  prompt: "You are leading a 3-person team to solve X. "
          "Create team via TeamCreate with members: [researcher, implementer, reviewer]. "
          "Use SendMessage to coordinate. Use TaskCreate for shared work list. "
          "Present final result to user."
)
```

**Error Recovery**:
- **Member fails task**: Other members continue, lead gets partial result
- **Member goes idle**: Lead can SendMessage to resume or TaskCreate new work
- **Communication loop endless**: Set max iteration count, escalate if stuck

**Escalation Rules**:
- If 2+ members report blocker → User AskUserQuestion for guidance
- If member rejected shutdown → Lead waits 10s, then collects available output
- If team creation fails → Fall back to sub-agent mode

**Example: Research + Analysis + Synthesis**
```
Team Created: Researcher, Analyst, Synthesizer

1. Researcher: "Investigating X..." → creates research-deep-dive.md
2. Analyst: "Cross-checking findings..." → creates analysis-validation.md
3. Researcher (via SendMessage): "Found Y, contradicts expectation"
4. Analyst (reply): "Confirmed, Z explains it"
5. Synthesizer: "Creating unified report..."
6. All members idle
7. Lead: "Here's the final report"
8. Leader sends shutdown_request
9. All approve
10. TeamDelete
```

---

## Sub-orchestrator

**When to Use**:
- Sequential task handoff (output A → input B)
- Agent coordination not needed (results only matter)
- Simpler, lower-coordination overhead

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
// Sequential
const resultA = Agent(prompt: "Analyze X", subagent_type: "analyzer")
const resultB = Agent(prompt: `Design based on: ${resultA}`, subagent_type: "designer")
const resultC = Agent(prompt: `Implement: ${resultB}`, subagent_type: "implementer")

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
- If result quality degrading → Switch to team mode for renegotiation

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

## Hybrid-orchestrator

**When to Use**:
- Mix of sequential and parallel stages
- Some stages need agent coordination, others don't
- Maximum flexibility for complex workflows

**Structure**:
```
Stage 1: Sequential (Sub-orchestrator)
  Research Agent → Analysis Agent

Stage 2: Parallel (Team-Orchestrator)
  [TeamCreate] → [Implementation Team with Backend, Frontend, Tests]

Stage 3: Sequential (Sub-orchestrator)
  Reviewer Agent (review all outputs)

Stage 4: Consolidate
  Final result to user
```

**How to Spawn**:
```
// Stage 1: Sequential research
const research = Agent(prompt: "Research X and Y")

// Stage 2: Parallel implementation
const implementation = Agent(
  prompt: "Lead a team to implement based on research...",
  teamCreate: true
)

// Stage 3: Sequential review
const reviewed = Agent(
  prompt: `Review this implementation: ${implementation}`,
  subagent_type: "reviewer"
)

// Consolidate
MoAI: "Implementation complete and reviewed"
```

**Error Recovery**:
- **Sequential stage fails**: Retry or skip to next stage
- **Team stage fails**: Fall back to sequential sub-agents for that stage
- **Cascade failure**: Collect partial results, present to user with blockers

**Escalation Rules**:
- If any stage fails after 2 retries → AskUserQuestion for decision
- If token budget exceeded → Prioritize remaining stages
- If team created but failing → Offer to switch to sequential approach

**Example: Full-Stack Feature Development**
```
Stage 1 (Research):
- Research Agent: "Analyze requirements and existing architecture"

Stage 2 (Implementation Team):
- TeamCreate with Backend, Frontend, Tester
- Parallel development for 2 days
- Integration testing

Stage 3 (Review):
- Review Agent: "Security + performance review"

Stage 4 (User Presentation):
- "Feature ready, here's the PR"
```

---

## Decision Matrix: Which Template?

| Scenario | Template | Why |
|----------|----------|-----|
| Single feature, clear handoff | Sub | No coordination needed, simpler |
| Feature with coordination required | Team | Agents need to negotiate |
| Large feature with research + dev | Hybrid | Research sequential, dev parallel |
| Bug investigation with hypotheses | Team | Debuggers need to share findings |
| Simple change (style update) | Sub | Overkill for complexity |
| Redesign involving multiple teams | Hybrid | Multiple phases with different needs |
| API + Client in parallel | Team | Tight coordination needed |

---

## Common Patterns Within Templates

### Sub-Orchestrator: Fan-out (Parallel Independent Results)
```
[MoAI] → Agent A, Agent B, Agent C (all in parallel)
        Combine A, B, C results
```

**When**: Independent analyses that don't depend on each other
**Example**: Concurrent code reviews from security + performance + architecture

### Sub-Orchestrator: Pipeline (Sequential Handoff)
```
[MoAI] → Agent A → (result) → Agent B → (result) → Agent C
```

**When**: Each stage requires previous stage's output
**Example**: Research → Design → Implementation → Testing

### Team-Orchestrator: Shared Work List
```
[Lead] → TaskCreate([Task1, Task2, Task3])
         Members self-select tasks
         → TaskUpdate(Task1 complete)
         → TaskUpdate(Task2 complete)
         → Remaining tasks handled...
```

**When**: Work is decomposable, members can self-coordinate
**Example**: Code migration (split files, each member migrates independently)

---

## Transition Rules

**When to Switch Templates During Execution**:

1. **Sub → Team**: If agents start reporting dependency on unplanned communication
   - Action: Collect current results, spawn team for next stage

2. **Team → Sub**: If team reaches deadlock or endless coordination loop
   - Action: Break down remaining tasks for sequential agents

3. **Hybrid → Simpler**: If complex workflow can be simplified
   - Action: Skip stages, consolidate results

**Signals to Transition**:
- **Sub feeling like Team**: Agents sending feedback to each other → Create team instead
- **Team feeling like Sub**: Agents work independently, little communication → Use sub-agents
- **Over-engineering**: Simple task using complex template → Simplify

---

## Monitoring & Adjustment

**During Execution**:
- Monitor agent progress via status messages
- Track token consumption against budget
- Check for communication loops or idle periods
- Assess quality of intermediate results

**Decision Points**:
- After each major stage: Is approach working? Continue or adjust?
- At 75% token budget: Prepare to conclude or escalate
- After first failure: Retry with same template or switch?

**Adjustment Options**:
- Continue with current template
- Switch to different template mid-stream
- Add more agents / members to current team
- Remove members / agents if not contributing
- Escalate to user for guidance

These three templates provide proven structures for most MoAI orchestration scenarios. Mix and match as needed for your specific workflow.
