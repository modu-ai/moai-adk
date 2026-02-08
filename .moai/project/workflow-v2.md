# MoAI-ADK Workflow v2.0 - Dual-Mode Architecture

Comprehensive workflow specification for MoAI-ADK's dual-mode execution engine supporting both sub-agent and Agent Teams modes.

Version: 2.0.0
Last Updated: 2026-02-06

---

## 1. Architecture Overview

MoAI-ADK v2.0 introduces a dual-mode execution engine that supports both traditional sub-agent workflows and the new Agent Teams collaborative workflows.

### Execution Modes

| Mode | Description | Best For |
|------|-------------|----------|
| Sub-agent | Sequential delegation to specialized agents via Task() | Simple tasks, single-domain work, quick fixes |
| Agent Teams | Parallel team-based development with shared task list | Complex multi-domain features, large-scale refactoring |
| Auto | Intelligent mode selection based on complexity analysis | Default behavior, adapts to task requirements |

### Architecture Diagram

```
User Request
    │
    ▼
┌──────────────────────────┐
│     MoAI Orchestrator     │
│  (Strategic Coordinator)  │
└──────────┬───────────────┘
           │
    ┌──────┴──────┐
    │  Mode       │
    │  Selector   │
    └──┬──────┬───┘
       │      │
  ┌────▼──┐ ┌─▼────────┐
  │Sub-   │ │Agent     │
  │Agent  │ │Teams     │
  │Mode   │ │Mode      │
  └───┬───┘ └────┬─────┘
      │          │
  ┌───▼───┐ ┌───▼──────────────┐
  │Task() │ │TeamCreate()      │
  │single │ │SendMessage()     │
  │agent  │ │Shared TaskList   │
  │return │ │TeamDelete()      │
  └───────┘ └──────────────────┘
```

---

## 2. Mode Selection Algorithm

### Input Sources (Priority Order)

1. CLI Flags: --team, --solo, --auto (highest priority)
2. Configuration: workflow.yaml execution_mode setting
3. Auto-Detection: Complexity scoring algorithm

### Complexity Scoring

| Factor | Condition | Score |
|--------|-----------|-------|
| Domain Count | >= 3 domains (backend, frontend, data, etc.) | +3 |
| File Count | >= 10 affected files | +3 |
| Cross-Layer | Changes span multiple architectural layers | +2 |
| Test Coverage | Requires new test infrastructure | +1 |
| Integration | External API/service integration | +1 |

Threshold: score >= 7 activates team mode (configurable in workflow.yaml)

### Decision Flow

```
1. Check --team flag → force team mode
2. Check --solo flag → force sub-agent mode
3. Read workflow.yaml execution_mode:
   - "subagent" → sub-agent mode
   - "team" → team mode (if available)
   - "auto" → calculate complexity score
4. If team mode selected:
   - Verify AGENT_TEAMS env is set
   - Verify Claude Code >= v2.1.32
   - If unavailable → warn user, fall back to sub-agent
```

---

## 3. Development Methodology

### Methodology Selection (Updated v2.0)

| Project Type | Test Coverage | Methodology | Rationale |
|-------------|---------------|-------------|-----------|
| Greenfield (new) | N/A | **Hybrid** (default) | TDD for new features, DDD structure for imported code |
| Brownfield | >= 50% | **Hybrid** | Strong test base, TDD for new + DDD for refactoring |
| Brownfield | 10-49% | **Hybrid** | Mixed coverage, expand with both approaches |
| Brownfield | < 10% | **DDD** | No tests, gradual characterization test creation |

### DDD Mode (ANALYZE-PRESERVE-IMPROVE)

Selected for: Existing/brownfield projects with legacy code (chosen at moai init)

- **ANALYZE**: Read existing code, map dependencies, identify side effects
- **PRESERVE**: Write characterization tests, create behavior snapshots
- **IMPROVE**: Make incremental changes, verify tests after each change

Agent: manager-ddd subagent

### TDD Mode (RED-GREEN-REFACTOR)

Selected for: Isolated new modules requiring strict test-first discipline (explicit selection only)

- **RED**: Write a failing test describing desired behavior
- **GREEN**: Write minimal code to pass the test
- **REFACTOR**: Clean up while keeping tests green

Agent: manager-tdd subagent

### Hybrid Mode (Default for New Projects)

Combines TDD and DDD based on change classification:

- New files → TDD rules (90%+ coverage target)
- Modified existing files → DDD rules (85%+ coverage target)
- New functions in existing files → TDD rules for those functions
- Deleted code → Verify characterization tests still pass

Agent: manager-tdd (new code) + manager-ddd (existing code)

---

## 4. SPEC Pipeline - Sub-Agent Mode

### Phase Overview

| Phase | Command | Agent | Token Budget | Purpose |
|-------|---------|-------|--------------|---------|
| Plan | /moai plan | manager-spec | 30K | Create SPEC document |
| Run | /moai run | manager-ddd/tdd | 180K | Implementation |
| Sync | /moai sync | manager-docs | 40K | Documentation sync |

### Plan Phase (Sub-Agent)

```
MoAI → Task(manager-spec) → SPEC document
                                │
                                ▼
                    .moai/specs/SPEC-XXX/spec.md
```

Steps:
1. Launch 3 parallel exploration agents (Explore, Research, Quality)
2. Collect findings from all agents
3. Delegate SPEC creation to manager-spec with findings
4. User approval of SPEC document
5. Execute /clear for next phase

### Run Phase (Sub-Agent)

```
MoAI → Task(manager-ddd or manager-tdd) → Implementation
                                              │
                                              ▼
                                    Source code + tests
```

Steps:
1. Read SPEC document
2. Read quality.yaml for development_mode
3. Route to appropriate methodology agent
4. DDD: ANALYZE-PRESERVE-IMPROVE cycle
5. TDD: RED-GREEN-REFACTOR cycle
6. Quality validation via manager-quality
7. Git operations via manager-git

### Sync Phase (Sub-Agent)

```
MoAI → Task(manager-docs) → Documentation + PR
```

Steps:
1. Quality verification
2. Documentation generation
3. README/CHANGELOG update
4. Pull request creation

---

## 5. SPEC Pipeline - Agent Teams Mode

### Phase Overview

| Phase | Team Composition | Mode | Coordination |
|-------|-----------------|------|-------------|
| Plan | researcher + analyst + architect | TeamCreate | Shared task list |
| Run | backend-dev + frontend-dev + tester + quality | TeamCreate | File ownership |
| Sync | manager-docs (sub-agent) | Task() | Always sub-agent |

### Plan Phase (Agent Teams)

```
MoAI (Team Lead)
    │
    ├── TeamCreate("moai-plan-{feature}")
    │
    ├── Spawn: researcher (haiku)
    │   └── Explore codebase, find patterns
    │
    ├── Spawn: analyst (sonnet)
    │   └── Analyze requirements, edge cases
    │
    ├── Spawn: architect (sonnet)
    │   └── Design approach, evaluate alternatives
    │
    ├── Collect findings via automatic messages
    │
    ├── Synthesize into SPEC (manager-spec sub-agent)
    │
    ├── User approval
    │
    ├── Shutdown teammates
    │
    └── TeamDelete + /clear
```

Key Differences from Sub-Agent Mode:
- 3 parallel research streams instead of sequential exploration
- Real-time coordination via SendMessage
- MoAI acts as Team Lead, synthesizes findings
- Faster plan creation for complex features

### Run Phase (Agent Teams)

```
MoAI (Team Lead)
    │
    ├── TeamCreate("moai-run-SPEC-XXX")
    │
    ├── Task Decomposition (file ownership assignment)
    │   ├── Backend tasks → backend-dev owns src/api/**, src/models/**
    │   ├── Frontend tasks → frontend-dev owns src/ui/**, src/components/**
    │   └── Test tasks → tester owns tests/**, *_test.go
    │
    ├── Spawn: backend-dev (sonnet)
    │   └── Implement API endpoints, models, services
    │
    ├── Spawn: frontend-dev (sonnet)
    │   └── Implement UI components, pages (waits for API contracts)
    │
    ├── Spawn: tester (sonnet)
    │   └── Write integration tests (waits for implementations)
    │
    ├── Coordinate via SendMessage
    │   └── Forward API contracts from backend to frontend
    │
    ├── Quality validation (team-quality or manager-quality sub-agent)
    │
    ├── Git operations (manager-git sub-agent)
    │
    ├── Shutdown teammates
    │
    └── TeamDelete
```

Key Differences from Sub-Agent Mode:
- Parallel implementation across domains
- Real-time cross-team coordination
- File ownership prevents write conflicts
- Shared task list for self-claiming work
- Quality validation after all work completes

### Sync Phase (Always Sub-Agent)

The sync phase always uses sub-agent mode regardless of team/auto setting:
- manager-docs handles documentation generation
- manager-git handles PR creation
- Single-agent work, no parallel benefit from teams

---

## 6. Team Composition Patterns

### Pattern 1: Plan Research Team

| Role | Agent | Model | Purpose |
|------|-------|-------|---------|
| researcher | team-researcher | haiku | Fast codebase exploration |
| analyst | general-purpose | sonnet | Requirements analysis |
| architect | general-purpose | sonnet | Technical design |

Use: Complex SPEC creation requiring multi-angle exploration
Duration: Short-lived (plan phase only)

### Pattern 2: Implementation Team

| Role | Agent | Model | Purpose |
|------|-------|-------|---------|
| backend-dev | team-backend-dev | sonnet | Server-side implementation |
| frontend-dev | team-frontend-dev | sonnet | Client-side implementation |
| tester | team-tester | sonnet | Test creation and coverage |

Use: Cross-layer feature implementation
Duration: Medium (full run phase)

### Pattern 3: Full-Stack Team

| Role | Agent | Model | Purpose |
|------|-------|-------|---------|
| api-layer | team-backend-dev | sonnet | API and business logic |
| ui-layer | team-frontend-dev | sonnet | UI and components |
| data-layer | team-backend-dev | sonnet | Database and schema |
| quality | team-quality | sonnet | Quality validation |

Use: Large-scale full-stack features
Duration: Medium-long

### Pattern 4: Investigation Team

| Role | Agent | Model | Purpose |
|------|-------|-------|---------|
| hypothesis-1 | general-purpose | haiku | First theory investigation |
| hypothesis-2 | general-purpose | haiku | Second theory investigation |
| hypothesis-3 | general-purpose | haiku | Third theory investigation |

Use: Complex debugging with competing theories
Duration: Short

### Pattern 5: Review Team

| Role | Agent | Model | Purpose |
|------|-------|-------|---------|
| security-reviewer | expert-security | sonnet | Security audit |
| perf-reviewer | expert-performance | sonnet | Performance review |
| quality-reviewer | team-quality | sonnet | Quality standards |

Use: Multi-perspective code review
Duration: Short

---

## 7. File Ownership Strategy

Prevent write conflicts by assigning exclusive file ownership per teammate.

### Default Ownership Map

| Role | Owned Paths | Access |
|------|------------|--------|
| backend-dev | src/api/**, src/models/**, src/services/**, internal/** | Read + Write |
| frontend-dev | src/ui/**, src/components/**, src/pages/**, public/** | Read + Write |
| tester | tests/**, __tests__/**, *_test.go, *.test.*, *.spec.* | Read + Write |
| quality | (none - read-only access to all files) | Read Only |
| researcher | (none - read-only access to all files) | Read Only |

### Ownership Rules

1. No two teammates own the same file
2. Shared types/interfaces: owned by the creating teammate, shared via message
3. Config files: owned by team lead or explicitly assigned
4. If ownership conflict: team lead resolves via SendMessage
5. Teammates can READ any file but only WRITE to owned files

---

## 8. Configuration

### workflow.yaml

```yaml
workflow:
  # Execution mode: auto, subagent, team
  execution_mode: "auto"

  # Auto-clear context between phases
  auto_clear:
    enabled: true
    after_plan: true
    token_threshold: 150000

  # Token budget per phase
  token_budget:
    plan: 30000
    run: 180000
    sync: 40000

  # Agent Teams configuration
  team:
    enabled: false  # Set to true to enable Agent Teams
    max_teammates: 10
    default_model: "sonnet"
    require_plan_approval: true
    delegate_mode: true

    # Auto-selection thresholds
    auto_selection:
      min_domains_for_team: 3
      min_files_for_team: 10
      min_complexity_score: 7

    # Pre-defined team patterns
    patterns:
      plan_research:
        roles: [researcher, analyst, architect]
        models: [haiku, sonnet, sonnet]
      implementation:
        roles: [backend-dev, frontend-dev, tester]
        models: [sonnet, sonnet, sonnet]
      full_stack:
        roles: [api-layer, ui-layer, data-layer, quality]
        models: [sonnet, sonnet, sonnet, sonnet]
      investigation:
        roles: [hypothesis-1, hypothesis-2, hypothesis-3]
        models: [haiku, haiku, haiku]
      review:
        roles: [security-reviewer, perf-reviewer, quality-reviewer]
        models: [sonnet, sonnet, sonnet]
```

### Enabling Agent Teams

1. Set environment variable in `.claude/settings.json`:
```json
{
  "env": {
    "CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS": "1"
  }
}
```

2. Enable in workflow.yaml:
```yaml
workflow:
  team:
    enabled: true
```

3. Use --team flag or set execution_mode to "auto"

---

## 9. /moai --auto End-to-End Flow

The complete autonomous workflow from SPEC to SYNC using Agent Teams:

### Step 1: Request Analysis

```
User: /moai "Build user authentication with JWT, login page, and API endpoints"
```

MoAI analyzes:
- Domains: backend (API), frontend (login page), data (user model) = 3 domains
- Estimated files: 12+ files
- Cross-layer: yes (API + UI + data)
- Complexity score: 3 + 3 + 2 = 8 (>= 7 threshold)
- Decision: **Team mode activated**

### Step 2: Plan Phase (Team)

```
1. TeamCreate("moai-plan-auth")
2. Spawn researcher → explores auth patterns in codebase
3. Spawn analyst → identifies JWT requirements, security needs
4. Spawn architect → designs token flow, middleware, API structure
5. Collect findings (automatic messages)
6. Forward researcher findings to architect
7. Delegate to manager-spec with all findings
8. Generate SPEC-AUTH-001/spec.md
9. User approval
10. Shutdown team + TeamDelete
11. /clear (free context for run phase)
```

### Step 3: Run Phase (Team)

```
1. Read SPEC-AUTH-001, analyze task scope
2. Check quality.yaml: development_mode = hybrid
   - New JWT code → TDD (RED-GREEN-REFACTOR)
   - Existing user model modifications → DDD (ANALYZE-PRESERVE-IMPROVE)

3. Task Decomposition:
   Task 1: "Create JWT token service" (TDD, no deps)
   Task 2: "Implement /auth/login endpoint" (TDD, blocked by Task 1)
   Task 3: "Implement /auth/refresh endpoint" (TDD, blocked by Task 1)
   Task 4: "Create login page component" (TDD, blocked by Task 2)
   Task 5: "Add auth middleware to existing routes" (DDD, blocked by Task 1)
   Task 6: "Write integration tests" (blocked by Tasks 2-5)
   Task 7: "TRUST 5 quality validation" (blocked by Task 6)

4. TeamCreate("moai-run-SPEC-AUTH-001")

5. Spawn teammates:
   - backend-dev: "Own src/api/auth/**, src/services/jwt/**. Follow TDD for new code."
   - frontend-dev: "Own src/pages/login/**, src/components/auth/**. Wait for API contracts."
   - tester: "Own tests/auth/**, tests/integration/**. Write integration tests after impl."

6. Parallel execution:
   - backend-dev claims Tasks 1, 2, 3, 5 (sequential within backend)
   - frontend-dev waits, then claims Task 4
   - MoAI forwards API contract info when backend Tasks 2-3 complete
   - tester claims Task 6 after implementations complete

7. Quality validation: Task 7 via team-quality or manager-quality

8. Git: manager-git sub-agent creates commit

9. Shutdown team + TeamDelete
```

### Step 4: Sync Phase (Sub-Agent)

```
1. Delegate to manager-docs:
   - Generate API documentation for auth endpoints
   - Update README with authentication section
   - Add CHANGELOG entry
2. Create pull request via manager-git
3. Present PR URL to user
```

### Step 5: Completion

```
MoAI - Complete
SPEC-AUTH-001 Implementation Complete
EXECUTION SUMMARY:
  - Mode: Agent Teams (auto-selected, score: 8)
  - Files Created: 12 files
  - Tests: 28/28 passing (100%)
  - Coverage: 92% (new code: 95%, modified: 87%)
  - Team Size: 3 teammates + lead
DELIVERABLES:
  - JWT token service
  - Login/refresh API endpoints
  - Login page component
  - Auth middleware
  - Integration test suite
  - API documentation
  - Pull request
<moai>COMPLETE</moai>
```

---

## 10. Error Recovery

### Team Mode Failures

| Scenario | Recovery Action |
|----------|----------------|
| Teammate crash | Spawn replacement with same role, resume from last task |
| Task stuck | Team lead reassigns to different teammate |
| File conflict | Team lead mediates via SendMessage, adjusts ownership |
| All teammates idle | Check TaskList for remaining work, assign or shutdown |
| Token limit | Shutdown team gracefully, fall back to sub-agent |
| AGENT_TEAMS unavailable | Warn user, fall back to sub-agent mode |
| Team creation fails | Fall back to sub-agent mode immediately |

### Fallback Strategy

If team mode fails at any point:
1. Shutdown remaining teammates gracefully
2. Fall back to sub-agent workflow
3. Continue from the last completed task (task list preserved)
4. Log warning about team mode failure
5. No data loss or state corruption

---

## 11. Token Economics

### Sub-Agent Mode

| Phase | Budget | Context Usage | Notes |
|-------|--------|---------------|-------|
| Plan | 30K | Single agent context | /clear after completion |
| Run | 180K | Sequential agent calls | Context accumulates |
| Sync | 40K | Single agent context | Final phase |
| **Total** | **250K** | | |

### Agent Teams Mode

| Phase | Budget | Context Usage | Notes |
|-------|--------|---------------|-------|
| Plan | 30K | Lead + 3 teammates | Each has independent context |
| Run | 180K | Lead + 3-4 teammates | Parallel contexts, shared task list |
| Sync | 40K | Single agent (sub-agent) | Same as sub-agent mode |
| **Total** | **250K** | | Higher parallelism |

### Cost Considerations

- Team mode uses more total API calls but completes faster
- haiku model teammates (researcher) are cost-effective for exploration
- Developer teammates inherit user's current model (configurable per pattern)
- Sync phase always uses sub-agent (no team overhead)
- Auto mode optimizes by only using teams when complexity warrants it

---

## 12. Comparison Summary

| Aspect | Sub-Agent Mode | Agent Teams Mode |
|--------|---------------|-----------------|
| Execution | Sequential delegation | Parallel collaboration |
| Communication | Return values only | Real-time SendMessage |
| State | Isolated per agent | Shared task list |
| File Safety | No conflict (sequential) | File ownership boundaries |
| Speed | Moderate | Fast (parallel work) |
| Complexity | Simple | Higher coordination |
| Best For | Single-domain tasks | Multi-domain features |
| Token Usage | Linear | Parallel (higher total) |
| Error Recovery | Retry single agent | Replace teammate, continue |
| Team Awareness | None | Full visibility |
| MoAI Role | Delegator | Team Lead |
| Max Agents | 1 at a time | Up to 5 concurrent |

---

## 13. Migration Guide

### From v1.x to v2.0

No breaking changes. Existing workflows continue to work unchanged.

New capabilities:
1. Add workflow.yaml to .moai/config/sections/
2. Team agent definitions added to .claude/agents/moai/
3. Team workflow skills added to .claude/skills/moai/
4. SKILL.md and CLAUDE.md updated with team mode sections
5. workflow-modes.md updated: Hybrid is default for new projects
6. spec-workflow.md updated: Agent Teams variant section added

### Enabling Team Mode

1. Ensure Claude Code >= v2.1.32
2. Set CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1
3. Set workflow.team.enabled: true
4. Use --team flag or --auto for intelligent selection

### Configuration Migration

No existing configuration changes required. New workflow.yaml is additive:
- Default execution_mode: "auto"
- Team disabled by default (enabled: false)
- All thresholds have sensible defaults

---

## Appendix: File Inventory

### New Files (v2.0)

| File | Purpose |
|------|---------|
| .moai/config/sections/workflow.yaml | Workflow and team configuration |
| .claude/agents/moai/team-researcher.md | Research teammate agent |
| .claude/agents/moai/team-backend-dev.md | Backend developer teammate |
| .claude/agents/moai/team-frontend-dev.md | Frontend developer teammate |
| .claude/agents/moai/team-tester.md | Testing specialist teammate |
| .claude/agents/moai/team-quality.md | Quality validation teammate |
| .claude/skills/moai/moai-workflow-team/SKILL.md | Team workflow management skill |
| .claude/skills/moai/workflows/team-plan.md | Team plan phase workflow |
| .claude/skills/moai/workflows/team-run.md | Team run phase workflow |
| .moai/docs/workflow-v2.md | This document |

### Modified Files (v2.0)

| File | Changes |
|------|---------|
| CLAUDE.md | Added Team Agents catalog, Agent Teams section (Section 15) |
| .claude/skills/moai/SKILL.md | Added team tools, flags, agents, execution steps |
| .claude/rules/moai/workflow/workflow-modes.md | DDD for existing, Hybrid as default |
| .claude/rules/moai/workflow/spec-workflow.md | Added Agent Teams variant section |
| .claude/skills/moai/workflows/moai.md | Added team mode routing, flags, v2.0 |

---

Version: 2.0.0
Last Updated: 2026-02-06
Author: MoAI Orchestrator
Status: Active
