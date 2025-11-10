# Core Architecture: Alfred Hybrid System

Deep dive into the architecture of Alfred SuperAgent, the heart of MoAI-ADK.

## Overall Structure

### 4-Layer Stack

```
┌─────────────────────────────────────────────┐
│           Commands (/alfred:*)              │
│   (Workflow Orchestration & User Entry)     │
└────────────────┬────────────────────────────┘
                 │
┌────────────────▼────────────────────────────┐
│        Sub-agents (19 Team Members)          │
│   (Deep reasoning & Decision making)        │
└────────────────┬────────────────────────────┘
                 │
┌────────────────▼────────────────────────────┐
│       Claude Skills (93 Skills)              │
│   (Reusable knowledge capsules)             │
└────────────────┬────────────────────────────┘
                 │
┌────────────────▼────────────────────────────┐
│           Hooks (<100ms)                    │
│   (Guardrails & Context)                    │
└─────────────────────────────────────────────┘
```

______________________________________________________________________

## Alfred SuperAgent

Alfred is the central orchestrator that coordinates the **SPEC → TDD → Sync** workflow.

### Core Characteristics

| Characteristic | Description                                       |
| ------------- | ------------------------------------------ |
| **Autonomy**  | Understands user intent and makes independent decisions |
| **Reasoning** | Breaks down complex tasks step by step                |
| **Team Coordination** | Optimally deploys 19 specialists              |
| **Learning** | Continuously improves from session logs                |
| **Transparency** | Records all decisions in a traceable manner             |

______________________________________________________________________

## 19-Member Team Structure

### 10 Core Sub-agents

| Agent                      | Role                     | Activation Condition          |
| -------------------------- | ------------------------ | -------------------- |
| **project-manager**        | Project initialization and setup  | `/alfred:0-project`  |
| **spec-builder**           | SPEC writing (EARS syntax)    | `/alfred:1-plan`     |
| **implementation-planner** | Architecture and implementation planning    | `/alfred:2-run` start |
| **tdd-implementer**        | RED→GREEN→REFACTOR execution  | `/alfred:2-run` during |
| **doc-syncer**             | Automatic documentation generation and sync | `/alfred:3-sync`     |
| **tag-agent**              | TAG validation and traceability management  | `/alfred:3-sync`     |
| **git-manager**            | Git workflow automation    | All stages            |
| **trust-checker**          | TRUST 5 principle validation        | `/alfred:2-run` completion |
| **quality-gate**           | Release readiness check    | `/alfred:3-sync`     |
| **debug-helper**           | Error analysis and resolution        | Auto-activated when needed   |

### 6 Expert Agents

| Expert              | Domain                        | Activation Condition              |
| ------------------- | ----------------------------- | ------------------------ |
| **backend-expert**  | API, server, DB architecture        | SPEC contains server/API keywords   |
| **frontend-expert** | UI, state management, performance            | SPEC contains frontend keywords |
| **devops-expert**   | Deployment, CI/CD, infrastructure           | SPEC contains deployment keywords       |
| **ui-ux-expert**    | Design system, accessibility         | SPEC contains design keywords     |
| **security-expert** | Security analysis, vulnerability diagnosis        | SPEC contains security keywords       |
| **database-expert** | DB design, optimization, migration | SPEC contains DB keywords         |

### 2 Built-in Agents (Claude)

- **Claude Opus/Sonnet**: For complex reasoning needs
- **Claude Haiku**: For lightweight tasks

______________________________________________________________________

## Hybrid Pattern

### Lead-Specialist Pattern

Specialized domain experts support the lead agent.

```
User Request
    ↓
Alfred (Lead)
    ├─→ Frontend keyword detected
    │   └─→ frontend-expert activated (Specialist)
    ├─→ Database keyword detected
    │   └─→ database-expert activated (Specialist)
    └─→ Security keyword detected
        └─→ security-expert activated (Specialist)
```

**Use Cases**:

- UI component design needed → UI/UX Expert
- DB performance optimization → Database Expert
- Security review → Security Expert

### Master-Clone Pattern

Large-scale tasks are processed in parallel by Alfred clones.

```
Large-scale task (100+ files, 5+ steps)
    ↓
Master Alfred (Coordination)
    ├─→ Clone-1: Refactor module A
    ├─→ Clone-2: Refactor module B
    └─→ Clone-3: Refactor module C
    ↓
Result merging and integration
```

**Use Cases**:

- Large-scale migration (v1.0 → v2.0)
- Full architecture refactoring
- Multi-domain simultaneous work

______________________________________________________________________

## 4-Step Workflow

### Phase 1: Intent Understanding

```
User Request → Clarity Assessment
├─ Clear: Proceed to Phase 2
└─ Unclear: AskUserQuestion → User Response → Proceed to Phase 2
```

**Alfred's Role**:

- Request analysis and classification
- Collect additional information if needed
- Confirm task scope

### Phase 2: Plan Creation

```
Plan Agent Call
    ↓
├─ Task Decomposition
├─ Dependency Analysis
├─ Parallelization Opportunity Identification
├─ File List Specification
└─ Time Estimation
    ↓
User Approval (AskUserQuestion)
    ↓
TodoWrite Initialization
```

### Phase 3: Task Execution

```
RED Phase
├─ Write Tests
└─ Confirm All Fail

GREEN Phase
├─ Minimal Implementation
└─ Confirm All Pass

REFACTOR Phase
├─ Code Improvement
└─ Maintain Tests
```

**TDD Strictness**:

- RED: No implementation code allowed
- GREEN: Add only minimum necessary
- REFACTOR: Maintain tests

### Phase 4: Report & Commit

```
Task Completion
    ↓
├─ Documentation Generation (based on settings)
├─ Git Commit (automatic)
├─ PR Creation (team mode)
└─ Cleanup
```

______________________________________________________________________

## TAG System (Traceability)

### TAG Chain

```
SPEC-001 (Requirement)
    ↓
@TEST:APP-001:* (Test)
    ↓
@CODE:APP-001:* (Implementation)
    ↓
@DOC:APP-001:* (Documentation)
    ↓
Cross-references (Complete Traceability)
```

### Traceability Guarantee

| Artifact | TAG                | Purpose          |
| -------- | ------------------ | ------------- |
| SPEC     | `SPEC-001`         | Requirement definition |
| Test     | `@TEST:SPEC-001`   | Requirement validation |
| Code     | `@CODE:SPEC-001:*` | Implementation tracking     |
| Documentation     | `@DOC:SPEC-001`    | Documentation sync   |

______________________________________________________________________

## Core Skill System

### 93 Claude Skills

Skills are loaded only when needed using the **Progressive Disclosure** principle.

#### Foundation Skills

- TRUST 5 principles
- TAG system
- SPEC writing
- Git workflow

#### Essential Skills

- Debugging
- Performance optimization
- Refactoring
- Test writing

#### Alfred Skills

- Agent guide
- Workflow
- Decision principles

#### Domain Skills

- Database
- Backend API
- Frontend UI
- Security

#### Language Skills

- Python 3.13+
- TypeScript 5.7+
- Go 1.24+
- 20+ other languages

______________________________________________________________________

## Safety Mechanisms (Hooks)

### SessionStart Hook

- Project status check
- Session log analysis
- Configuration validation

### PreToolUse Hook

- Block dangerous commands
- Permission verification
- Context delivery

### PostToolUse Hook

- Result analysis
- Error detection
- Automatic fix suggestions

______________________________________________________________________

## Performance Metrics

### Expected Execution Time

| Stage          | Average Time | Range       |
| ------------- | --------- | ---------- |
| Intent Understanding     | 1 min       | 1-5 min      |
| Plan Creation     | 2 min       | 1-5 min      |
| RED Stage      | 3 min       | 1-10 min     |
| GREEN Stage    | 5 min       | 2-15 min     |
| REFACTOR Stage | 5 min       | 2-15 min     |
| Sync        | 2 min       | 1-5 min      |
| **Total**      | **18 min**  | **8-55 min** |

### Productivity Improvement

Traditional Development vs MoAI-ADK:

| Metric            | Traditional | MoAI-ADK | Improvement    |
| --------------- | ---- | -------- | --------- |
| Development Speed       | 100% | 250%     | **+150%** |
| Bug Reduction       | 100% | 20%      | **-80%**  |
| Documentation Time     | 100% | 10%      | **-90%**  |
| Test Coverage | 60%  | 95%      | **+35%**  |

______________________________________________________________________

## Scalability

### Horizontal Scaling

Unlimited parallelization with Master-Clone pattern:

- 10 modules developed simultaneously
- 100+ files changed simultaneously
- Rewrites and migrations

### Vertical Scaling

Integration of additional expert agents:

- Add new domain experts
- Extend custom Skills
- Unlimited team size

______________________________________________________________________

## Next Steps

- [Alfred Workflow](../guides/alfred/index.md) - Detailed 4-step guide
- [SPEC Writing](../guides/specs/basics.md) - Master EARS syntax
- [TDD Implementation](../guides/tdd/index.md) - RED-GREEN-REFACTOR
- [TAG System](../guides/specs/tags.md) - Perfect traceability

______________________________________________________________________

**Questions?** Visit the [Online Documentation Portal](https://adk.mo.ai.kr).





______________________________________________________________________

## Advanced Skill System

### Skill Tier Breakdown

#### Tier 1: Foundation Skills (Always Loaded)

```
moai-foundation-tags         # TAG system basics
moai-foundation-trust        # TRUST 5 principles
moai-foundation-best-practices  # Development standards
moai-foundation-workflows    # Core workflows
```

**Characteristics**:
- Always in context (<2KB total)
- Essential for all operations
- Loaded at session start
- Never unloaded

#### Tier 2: Alfred Skills (On-Demand)

```
moai-alfred-agent-guide      # Agent selection & coordination
moai-alfred-ask-user-questions  # Interactive TUI
moai-alfred-best-practices   # Alfred-specific patterns
moai-alfred-personas         # Adaptive communication
moai-alfred-workflow         # 4-step workflow
```

**Characteristics**:
- Loaded when Alfred needs decision support
- Progressive disclosure (load as needed)
- Context-aware selection
- Medium size (2-5KB each)

#### Tier 3: Domain Skills (Specialist)

```
moai-domain-backend          # API, server architecture
moai-domain-frontend         # UI, state management
moai-domain-database         # DB design, optimization
moai-domain-ml               # Machine learning
moai-domain-devops           # CI/CD, deployment
moai-domain-security         # Security patterns
```

**Characteristics**:
- Loaded when domain keywords detected
- Deep technical knowledge
- Specialist content (3-8KB each)

#### Tier 4: Language Skills (Syntax)

```
moai-lang-python            # Python 3.13+
moai-lang-typescript        # TypeScript 5.7+
moai-lang-go                # Go 1.24+
moai-lang-rust              # Rust 1.75+
[18+ additional languages]
```

**Characteristics**:
- Loaded based on codebase_language
- Language-specific idioms
- Framework integration
- Medium size (4-6KB each)

#### Tier 5: Tool Skills (Integration)

```
moai-tool-git               # Git workflow automation
moai-tool-docker            # Container management
moai-tool-kubernetes        # K8s orchestration
moai-tool-terraform         # Infrastructure as code
```

**Characteristics**:
- Loaded when tool usage detected
- Integration patterns
- Best practices per tool

#### Tier 6: Custom Skills (User-Defined)

```
moai-custom-mlops           # Custom ML pipeline
moai-custom-doc-validator   # Documentation validation
moai-custom-deployment      # Deployment automation
```

**Characteristics**:
- Project-specific knowledge
- User-created content
- Flexible structure

______________________________________________________________________

## Progressive Disclosure Strategy

### Context Budget Management

**Total Context: 200,000 tokens**

```
Allocation Strategy:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

System Prompts:           10,000 tokens (5%)
  - CLAUDE.md
  - Agent persona
  - Core instructions

Foundation Skills:         5,000 tokens (2.5%)
  - Always loaded
  - TAG, TRUST, workflows

Active Skills:            20,000 tokens (10%)
  - Loaded on-demand
  - Domain/language/tool specific

Project Context:          50,000 tokens (25%)
  - SPEC documents
  - Code files
  - Test files
  - Documentation

Conversation History:     50,000 tokens (25%)
  - Recent exchanges
  - TodoWrite state
  - Decision history

Response Buffer:          65,000 tokens (32.5%)
  - Generated content
  - Analysis
  - Recommendations
```

### Loading Strategy

**Strategy 1: Lazy Loading**
```python
# Load skill only when needed
if task_type == "backend_api":
    Skill("moai-domain-backend")
    Skill("moai-lang-python")
    # Don't load frontend/ML skills
```

**Strategy 2: Batch Loading**
```python
# Load related skills together
if task_type == "full_stack":
    Skill("moai-domain-backend")
    Skill("moai-domain-frontend")
    Skill("moai-domain-database")
```

**Strategy 3: Unloading**
```python
# Unload when no longer needed
after_refactor_phase:
    unload("moai-domain-ml")  # Free context
    load("moai-tool-git")     # Load for commit
```

______________________________________________________________________

## Agent Coordination Patterns

### Pattern 1: Sequential Handoff

```
spec-builder
    ↓ (SPEC complete)
tdd-implementer
    ↓ (Code complete)
doc-syncer
    ↓ (Docs complete)
git-manager
```

**Use Case**: Linear workflow (most common)

### Pattern 2: Parallel Execution

```
Alfred (Master)
    ├─→ backend-expert  (API implementation)
    ├─→ frontend-expert (UI implementation)
    └─→ database-expert (Schema design)
    ↓ (All complete)
Integration & Testing
```

**Use Case**: Independent tasks, no dependencies

### Pattern 3: Recursive Decomposition

```
Large Task
    ↓
Alfred splits into 5 subtasks
    ├─→ Clone-1 (subtask 1)
    ├─→ Clone-2 (subtask 2)
    ├─→ Clone-3 (subtask 3)
    ├─→ Clone-4 (subtask 4)
    └─→ Clone-5 (subtask 5)
    ↓
Master Alfred merges results
```

**Use Case**: Large-scale refactoring, migrations

### Pattern 4: Specialist Consultation

```
tdd-implementer (Lead)
    ↓ (Security concern)
Consults security-expert
    ↓ (Returns recommendation)
tdd-implementer applies fix
```

**Use Case**: Ad-hoc expert advice

______________________________________________________________________

## Decision Trees

### Alfred's Intent Classification

```
User Request
    ↓
Intent Analysis
    ├─ Project Setup? → /alfred:0-project
    ├─ SPEC Creation? → /alfred:1-plan
    ├─ Implementation? → /alfred:2-run
    ├─ Documentation? → /alfred:3-sync
    ├─ Debugging? → debug-helper
    ├─ Feedback? → /alfred:9-feedback
    └─ Unclear? → AskUserQuestion
```

### Agent Selection Tree

```
Task Type
    ├─ Backend API
    │   └─ backend-expert + moai-lang-python
    ├─ Frontend UI
    │   └─ frontend-expert + moai-lang-typescript
    ├─ Database Design
    │   └─ database-expert + moai-domain-database
    ├─ Security Review
    │   └─ security-expert + moai-domain-security
    ├─ ML Model
    │   └─ custom ml-validator + moai-domain-ml
    └─ DevOps
        └─ devops-expert + moai-tool-kubernetes
```

### Skill Loading Tree

```
SPEC Keywords Detected
    ├─ "API", "REST", "GraphQL"
    │   └─ Load: moai-domain-backend
    ├─ "React", "Vue", "Angular"
    │   └─ Load: moai-domain-frontend
    ├─ "PostgreSQL", "MongoDB", "Redis"
    │   └─ Load: moai-domain-database
    ├─ "Machine Learning", "TensorFlow"
    │   └─ Load: moai-domain-ml
    └─ "Docker", "Kubernetes"
        └─ Load: moai-tool-docker, moai-tool-kubernetes
```

______________________________________________________________________

## Communication Protocols

### Agent-to-Agent Communication

**Protocol 1: Task Delegation**
```python
Task(
    description="Design database schema for user authentication",
    subagent_type="database-expert",
    context={
        "requirements": spec_content,
        "tech_stack": "PostgreSQL 16",
        "expected_load": "10k users"
    }
)
```

**Protocol 2: Result Sharing**
```python
# backend-expert completes API
api_result = {
    "endpoints": ["/api/login", "/api/logout"],
    "tests": "tests/test_auth.py",
    "coverage": 92
}

# Pass to frontend-expert
Task(
    description="Create UI for authentication",
    subagent_type="frontend-expert",
    context={"api_result": api_result}
)
```

**Protocol 3: Conflict Resolution**
```python
# Two agents disagree
if backend_expert.recommendation != security_expert.recommendation:
    # Escalate to Alfred
    Alfred.resolve_conflict(
        agent_a="backend-expert",
        agent_b="security-expert",
        issue="Session management approach"
    )
```

### Hook Communication

**SessionStart Hook Output**:
```python
{
    "project_status": {
        "initialized": true,
        "specs": ["SPEC-001", "SPEC-002"],
        "active_branch": "feature/SPEC-001",
        "last_commit": "3 hours ago"
    },
    "recommendations": [
        "SPEC-002 needs implementation",
        "Consider running tests"
    ]
}
```

**PreToolUse Hook Validation**:
```python
# Input
{
    "tool": "Bash",
    "command": "rm -rf src/"
}

# Output
{
    "allowed": false,
    "reason": "Destructive command blocked",
    "alternative": "Please specify files to delete"
}
```

______________________________________________________________________

## Scalability Patterns

### Horizontal Scaling

**Scenario**: 100-file refactoring

```
Alfred (Master Coordinator)
    ├─→ Clone-1: Refactor modules 1-20
    ├─→ Clone-2: Refactor modules 21-40
    ├─→ Clone-3: Refactor modules 41-60
    ├─→ Clone-4: Refactor modules 61-80
    └─→ Clone-5: Refactor modules 81-100
    ↓
Merge changes (conflict resolution)
    ↓
Run full test suite
    ↓
Generate comprehensive report
```

**Benefits**:
- Parallel execution (5x faster)
- Isolated contexts (no interference)
- Fault tolerance (1 clone fails, others continue)

### Vertical Scaling

**Scenario**: Deep domain expertise needed

```
Backend Expert (Primary)
    ├─ Skill("moai-domain-backend")
    ├─ Skill("moai-lang-python")
    ├─ Skill("moai-domain-database")
    ├─ Skill("moai-domain-security")
    └─ Skill("moai-tool-docker")
    ↓
Deep context (50,000 tokens of domain knowledge)
    ↓
Highly specialized recommendations
```

**Benefits**:
- Deep expertise
- Context-rich decisions
- Comprehensive analysis

______________________________________________________________________

## Performance Optimization

### Caching Strategy

**Strategy 1: Skill Caching**
```python
# First load
Skill("moai-domain-backend")  # Load from disk (50ms)

# Subsequent loads (same session)
Skill("moai-domain-backend")  # Load from cache (1ms)
```

**Strategy 2: Context Caching**
```python
# Cache expensive analyses
analysis_cache = {
    "code_complexity": calculate_complexity(),  # Expensive
    "test_coverage": get_coverage()             # Expensive
}

# Reuse in multiple agents
backend_expert.use_cache(analysis_cache)
security_expert.use_cache(analysis_cache)
```

### Memory Management

**Strategy 1: Sliding Window (Conversation History)**
```python
# Keep only recent 50 exchanges
history = session.get_history(limit=50)

# Summarize older exchanges
if len(session.history) > 100:
    summary = summarize(session.history[:-50])
    session.history = [summary] + session.history[-50:]
```

**Strategy 2: Context Compression**
```python
# Large SPEC document (10,000 tokens)
spec_content = read_spec("SPEC-001.md")

# Compress for context (2,000 tokens)
spec_summary = extract_key_points(spec_content)

# Use summary unless full detail needed
agent.context = spec_summary
```

______________________________________________________________________

## Failure Recovery

### Recovery Patterns

**Pattern 1: Retry with Backoff**
```python
def call_agent_with_retry(agent, max_attempts=3):
    for attempt in range(max_attempts):
        try:
            return agent.execute()
        except TemporaryError:
            wait_time = 2 ** attempt  # 1s, 2s, 4s
            time.sleep(wait_time)
    
    # Final attempt failed
    escalate_to_debug_helper()
```

**Pattern 2: Graceful Degradation**
```python
# Primary approach fails
try:
    result = complex_analysis()
except Exception:
    # Fallback to simpler approach
    result = simple_analysis()
```

**Pattern 3: Human-in-the-Loop**
```python
# Agent stuck
if agent.no_progress_for(minutes=5):
    AskUserQuestion(
        question="I'm having trouble with X. Should I try Y or Z?",
        options=["Approach Y", "Approach Z", "Explain the issue"]
    )
```

______________________________________________________________________

## Monitoring and Observability

### Key Metrics

**Performance Metrics**:
```
Average Response Time:        2.5s
Agent Activation Time:        0.8s
Skill Load Time:              0.05s
Context Switch Time:          0.3s
Total Workflow Time:          18 min
```

**Quality Metrics**:
```
SPEC Accuracy:                95%
Test Coverage:                92%
Code Quality Score:           8.7/10
Documentation Completeness:   90%
TAG Traceability:             100%
```

**Resource Metrics**:
```
Context Usage:                120,000 / 200,000 tokens
Memory Usage:                 4.2 GB
Active Skills:                12 / 93
Active Agents:                3 / 19
```

### Logging

**Log Levels**:
```python
# CRITICAL: System failures
log.critical("Agent crashed: backend-expert")

# ERROR: Recoverable errors
log.error("SPEC-001.md not found, creating new")

# WARNING: Potential issues
log.warning("Test coverage below 85%")

# INFO: Normal operations
log.info("Loaded skill: moai-domain-backend")

# DEBUG: Detailed traces
log.debug("Context usage: 45,000 tokens")
```

______________________________________________________________________

## Security Architecture

### Security Layers

**Layer 1: Input Validation** (Hooks)
```python
# PreToolUse Hook
validate_command(command)
validate_file_path(path)
validate_permissions(user, action)
```

**Layer 2: Sandboxing** (Agents)
```python
# Agents have restricted tool access
backend_expert.allowed_tools = ["Read", "Write", "Bash"]
frontend_expert.allowed_tools = ["Read", "Write"]
# No agent can use Bash for destructive operations
```

**Layer 3: Audit Logging** (PostToolUse)
```python
# Log all actions
audit_log.record({
    "agent": "backend-expert",
    "tool": "Write",
    "file": "src/api.py",
    "timestamp": "2025-11-10T14:30:00Z"
})
```

**Layer 4: Access Control** (Configuration)
```json
{
  "security": {
    "prevent_secrets_commit": true,
    "scan_files_before_git": true,
    "blocked_patterns": [".env", "credentials"]
  }
}
```

______________________________________________________________________

## Future Architecture Enhancements

### Roadmap

**Q1 2026**:
- Multi-modal support (images, diagrams)
- Real-time collaboration (multiple users)
- Plugin marketplace

**Q2 2026**:
- Cloud-based agent execution
- Distributed skill library
- Advanced analytics dashboard

**Q3 2026**:
- Custom LLM integration (GPT-5, Gemini 2.0)
- Enterprise features (SSO, RBAC)
- Advanced security scanning

______________________________________________________________________

## Conclusion

MoAI-ADK's architecture is designed for:
- **Scalability**: Handle projects of any size
- **Flexibility**: Adapt to any domain/language
- **Reliability**: Fail gracefully, recover automatically
- **Performance**: Optimize context usage, minimize latency
- **Security**: Multiple defense layers, audit everything

The 4-layer stack (Commands → Agents → Skills → Hooks) provides clear separation of concerns while enabling powerful orchestration of 19 AI experts across 93 knowledge domains.

______________________________________________________________________

**Related Documentation**:
- [Alfred Workflow Guide](../guides/alfred/index.md)
- [Skill Development](extensions.md)
- [Performance Tuning](performance.md)
- [Security Best Practices](security.md)
