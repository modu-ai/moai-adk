---
name: moai-alfred-todowrite-pattern
version: 2.0.0
created: 2025-11-02
updated: 2025-11-11
status: active
description: "TodoWrite auto-initialization patterns from Plan agent, task tracking best practices, status management. Enhanced with research capabilities for productivity optimization and workflow efficiency analysis. Use when setting up task tracking, managing TodoWrite states, or optimizing workflow productivity."
keywords: ['todowrite', 'task-tracking', 'productivity', 'workflow', 'automation', 'research', 'productivity-optimization', 'workflow-efficiency']
allowed-tools:
  - Read
  - AskUserQuestion
  - TodoWrite
---

# TodoWrite Pattern Skill

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-alfred-todowrite-pattern |
| **Version** | 2.0.0 (2025-11-11) |
| **Status** | Active |
| **Tier** | Alfred |
| **Purpose** | Optimize TodoWrite usage and task tracking workflows |

---

## What It Does

Comprehensive guide for TodoWrite auto-initialization, task tracking best practices, status management, and workflow optimization. Provides patterns for effective task management across Alfred's 4-step workflow.

**Key capabilities**:
- Auto-initialization patterns from Plan agent
- Task state management (not_started → in_progress → completed)
- Productivity optimization research
- Workflow efficiency analysis
- Cross-skill task coordination

---

## When to Use

**Automatic triggers**:
- Plan agent task decomposition
- Multi-step task execution
- Workflow state transitions
- Task completion tracking

**Manual invocation**:
- Optimizing TodoWrite usage
- Debugging task tracking issues
- Setting up task management workflows
- Analyzing productivity patterns

---

## TodoWrite Integration Patterns

### Auto-Initialization from Plan Agent

```python
# Plan agent automatically initializes TodoWrite
def plan_agent_workflow():
    """Standard Plan agent TodoWrite initialization"""

    # 1. Task decomposition
    tasks = decompose_user_request()

    # 2. TodoWrite initialization
    TodoWrite.create([
        {"task": "Analyze requirements", "status": "not_started"},
        {"task": "Create SPEC document", "status": "not_started"},
        {"task": "Write tests (RED)", "status": "not_started"},
        {"task": "Implement feature (GREEN)", "status": "not_started"},
        {"task": "Refactor code (REFACTOR)", "status": "not_started"},
        {"task": "Update documentation", "status": "not_started"}
    ])

    return tasks
```

### State Management Patterns

#### TDD Workflow State Transitions

```python
# RED Phase
TodoWrite.update("Write tests", "in_progress")
# ... write failing tests
TodoWrite.update("Write tests", "completed")

# GREEN Phase
TodoWrite.update("Implement feature", "in_progress")
# ... implement minimal code
TodoWrite.update("Implement feature", "completed")

# REFACTOR Phase
TodoWrite.update("Refactor code", "in_progress")
# ... refactor with safety
TodoWrite.update("Refactor code", "completed")
```

#### Multi-Agent Coordination

```python
# Agent handoff with TodoWrite continuity
def agent_coordination():
    """Maintain task state across agent transitions"""

    # Alfred creates initial tasks
    TodoWrite.create([
        {"task": "Security analysis", "status": "not_started", "assignee": "security-agent"},
        {"task": "Performance review", "status": "not_started", "assignee": "perf-agent"}
    ])

    # Specialist agents update their tasks
    security_agent.analyze()  # Updates TodoWrite internally
    perf_agent.review()       # Updates TodoWrite internally

    # Alfred reviews completion
    TodoWrite.get_all()  # Shows all task states
```

---

## Task Organization Patterns

### Hierarchical Task Structure

```markdown
## Project: Authentication System (AUTH-001)

### Main Tasks
- [ ] Design authentication flow
- [ ] Implement JWT tokens
- [ ] Add password reset
- [ ] Write documentation

### Subtasks (expandable)
#### Design authentication flow
- [ ] Research best practices
- [ ] Create sequence diagrams
- [ ] Define API endpoints
- [ ] Security requirements

#### Implement JWT tokens
- [ ] Setup JWT library
- [ ] Create token generation
- [ ] Implement validation middleware
- [ ] Add refresh token logic
```

### Priority-Based Task Management

```python
# Priority-based TodoWrite patterns
task_priorities = {
    "critical": ["Security fixes", "Database migrations", "API breaking changes"],
    "high": ["Feature implementation", "Test coverage", "Documentation updates"],
    "medium": ["Code refactoring", "Performance optimization", "Bug fixes"],
    "low": ["Code cleanup", "Dependency updates", "Minor improvements"]
}

def prioritize_tasks(tasks):
    """Organize tasks by priority"""
    organized = {}
    for task in tasks:
        for priority, keywords in task_priorities.items():
            if any(keyword in task.lower() for keyword in keywords):
                organized.setdefault(priority, []).append(task)
                break
    return organized
```

---

## Workflow Integration

### 4-Step Alfred Workflow Integration

```markdown
## Step 1: Intent Understanding
TodoWrite.create([
    {"task": "Clarify user requirements", "status": "in_progress"}
])

## Step 2: Plan Creation
TodoWrite.create([
    {"task": "Create execution plan", "status": "not_started"},
    {"task": "Identify dependencies", "status": "not_started"},
    {"task": "Estimate work scope", "status": "not_started"}
])

## Step 3: Task Execution
TodoWrite.update("Create execution plan", "in_progress")
# ... planning activities
TodoWrite.update("Create execution plan", "completed")

## Step 4: Report & Commit
TodoWrite.create([
    {"task": "Generate completion report", "status": "not_started"},
    {"task": "Create git commit", "status": "not_started"}
])
```

### Cross-Session Task Persistence

```python
# Session handoff with TodoWrite continuity
def session_handoff():
    """Maintain task continuity across sessions"""

    # Save current state
    current_tasks = TodoWrite.get_all()
    session_context = {
        "active_tasks": current_tasks,
        "session_progress": calculate_progress(),
        "next_session_focus": identify_next_priorities()
    }

    # Save to session memory
    save_to_memory(session_context)

    # Next session restores
    restored_context = load_from_memory()
    TodoWrite.restore(restored_context["active_tasks"])
```

---

## Research Integration

### Productivity Optimization Research Capabilities

**Task Management Research**:
- **Task completion pattern analysis**: Research on effective task decomposition and sequencing strategies
- **Priority optimization studies**: Analyze the effectiveness of different prioritization methods
- **Workflow efficiency research**: Study optimal task management patterns for different project types
- **Agent coordination optimization**: Research on task distribution across specialist agents

**Productivity Measurement Research**:
- **Task completion rate analysis**: Research factors that influence task completion speed and quality
- **Workflow bottleneck identification**: Study common bottlenecks in task execution and optimization strategies
- **Session productivity patterns**: Research productivity patterns across different session lengths and types
- **Multi-agent efficiency studies**: Analyze the efficiency of task distribution across multiple agents

**Research Methodology**:
- **Task performance tracking**: Monitor task completion times, success rates, and quality metrics
- **Workflow pattern analysis**: Study successful task management patterns across different projects
- **Productivity correlation studies**: Analyze the relationship between task management approaches and project success
- **Agent coordination research**: Study optimal task delegation and coordination patterns

### Task Management Research Framework

#### 1. Task Decomposition Research
- **Optimal task sizing**: Research on ideal task granularity for different types of work
- **Dependency management studies**: Analyze effective task dependency identification and management
- **Task sequencing optimization**: Research optimal task ordering for maximum efficiency
- **Subtask coordination patterns**: Study effective management of hierarchical task structures

#### 2. Priority Management Research
- **Priority algorithm effectiveness**: Research different prioritization methods and their effectiveness
- **Dynamic prioritization studies**: Analyze adaptive priority adjustment based on changing conditions
- **Priority communication research**: Study effective communication of task priorities to team members
- **Priority conflict resolution**: Research methods for resolving conflicting task priorities

#### 3. Workflow Optimization Research
```
Task Management Research Framework:
├── Productivity Analysis
│   ├── Task completion rate studies
│   ├── Workflow bottleneck identification
│   ├── Session productivity patterns
│   └── Quality vs. speed analysis
├── Coordination Research
│   ├── Multi-agent task distribution
│   ├── Cross-session continuity patterns
│   ├── Agent handoff optimization
│   └── Collaborative task management
└── Optimization Development
        ├── Priority algorithm research
        ├── Task sequencing optimization
        ├── Workflow efficiency studies
        └── Productivity benchmarking
```

**Current Research Focus Areas**:
- Task decomposition optimization for different project types
- Priority management algorithm improvement
- Multi-agent task coordination efficiency
- Cross-session task continuity enhancement
- Productivity metrics development and validation

---

## Integration with Research System

The TodoWrite pattern system integrates with MoAI-ADK's research framework by:

1. **Collecting task management data**: Track task completion patterns, productivity metrics, and workflow efficiency
2. **Validating optimization strategies**: Provide real-world testing ground for new task management approaches and algorithms
3. **Documenting productivity patterns**: Capture successful task management patterns and share them across teams
4. **Benchmarking workflow approaches**: Measure the effectiveness of different task management strategies and identify improvements

**Research Collaboration**:
- **Productivity research team**: Share data on task completion rates and workflow efficiency
- **Agent coordination team**: Provide insights on multi-agent task distribution and handoff optimization
- **Workflow optimization team**: Collaborate on workflow efficiency improvements and bottleneck identification
- **User behavior research team**: Study task management pattern adoption and user productivity improvements

---

**Learn more in `reference.md` for detailed TodoWrite patterns, productivity optimization strategies, and workflow integration examples.

**Related Skills**: moai-alfred-workflow, moai-alfred-practices, moai-foundation-tags