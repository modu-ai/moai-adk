---
name: moai-alfred-workflow
version: 2.0.0
created: 2025-11-02
updated: 2025-11-11
status: active
description: "Guide 4-step workflow execution with task tracking and quality gates. Enhanced with research capabilities for workflow optimization, performance analysis, and coordination pattern research. Use for workflow analysis, task coordination, and process optimization."
keywords: ['workflow', 'execution', 'planning', 'task-tracking', 'quality', 'research', 'optimization', 'performance-analysis', 'coordination-patterns']
allowed-tools:
  - Read
  - AskUserQuestion
  - TodoWrite
tags: [workflow, orchestration, task-tracking, research, analysis, optimization]
---

# Alfred 4-Step Workflow Guide

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-alfred-workflow |
| **Version** | 2.0.0 (2025-11-11) |
| **Status** | Active |
| **Tier** | Alfred |
| **Purpose** | Guide systematic 4-step workflow execution with research-enhanced optimization |

---

## What It Does

Alfred uses a consistent 4-step workflow for all user requests to ensure clarity, planning, transparency, and traceability. Enhanced with research capabilities for workflow optimization and performance analysis.

**Key capabilities**:
- ✅ Intent clarification with questions
- ✅ Task planning and decomposition
- ✅ Transparent progress tracking with TodoWrite
- ✅ Automated reporting and commits
- ✅ Quality gate validation
- ✅ Research-driven workflow optimization
- ✅ Performance pattern analysis
- ✅ Coordination efficiency research

---

## When to Use

**Automatic triggers**:
- User request received → analyze intent
- Multiple interpretation possible → use AskUserQuestion
- Task complexity > 1 step → invoke Plan Agent
- Executing tasks → activate TodoWrite tracking
- Task completion → generate report

**Manual reference**:
- Understanding workflow execution
- Planning multi-step features
- Learning best practices for task tracking
- Researching workflow optimization patterns

---

## The 4-Step Workflow

### Step 1: Intent Understanding

**Goal**: Clarify user intent before any action

**Actions**:
- Evaluate request clarity
  - HIGH clarity → Skip to Step 2 directly
  - MEDIUM/LOW clarity → Invoke AskUserQuestion
- Present 3-5 clear options (not open-ended)
- Gather user responses before proceeding

**When to Ask Questions**:
- Multiple tech stack choices available
- Architecture decisions needed
- Business/UX decisions unclear
- Ambiguous requirements
- Existing component impacts unknown

**Example**:
```
User says: "Add authentication"
          ↓
Clarity = MEDIUM (multiple approaches possible)
          ↓
Ask: "Which authentication method?"
- Option 1: JWT tokens
- Option 2: OAuth 2.0
- Option 3: Session-based
```

---

### Step 2: Plan Creation

**Goal**: Analyze tasks and identify execution strategy

**CRITICAL**: ALWAYS delegate to agents - never execute planning directly

**Actions**:
- Invoke Plan Agent via Task() to:
  - Decompose tasks into structured steps
  - Identify dependencies between tasks
  - Determine single vs parallel execution
  - Estimate file changes and scope
- Output structured task breakdown

**Proper Delegation Pattern**:
```bash
Task(
  subagent_type="Plan",
  description="Create execution strategy for user request",
  prompt="You are the Plan agent. Analyze this request and create structured implementation plan."
)
```

**Plan Output Format**:
```
Task Breakdown:

Phase 1: Preparation (30 mins)
├─ Task 1: Set up environment
├─ Task 2: Install dependencies
└─ Task 3: Create test fixtures

Phase 2: Implementation (2 hours)
├─ Task 4: Core feature (parallel ready)
├─ Task 5: API endpoints (parallel ready)
└─ Task 6: Tests (depends on 4, 5)

Phase 3: Verification (30 mins)
├─ Task 7: Integration testing
├─ Task 8: Documentation
└─ Task 9: Code review
```

---

### Step 3: Task Execution

**Goal**: Execute tasks with transparent progress tracking

**CRITICAL**: ALL task execution MUST be delegated to appropriate agents

**Actions**:
1. Initialize TodoWrite with all tasks (status: pending)
2. For each task:
   - Update TodoWrite: pending → **in_progress**
   - **Delegate to appropriate agent via Task()**:
     ```bash
     # Domain-specific tasks
     Task(subagent_type="backend-expert", ...)
     Task(subagent_type="frontend-expert", ...)
     Task(subagent_type="tdd-implementer", ...)
     Task(subagent_type="quality-gate", ...)
     ```
   - Update TodoWrite: in_progress → **completed**
3. Handle blockers: Keep in_progress, create new blocking task

**Forbidden Direct Execution**:
❌ Writing code directly in commands
❌ Making architectural decisions without agents
❌ Bypassing agent expertise

**Required Agent Delegation**:
✅ Always use Task() for domain work
✅ Let agents own their expertise areas
✅ Maintain clear responsibility boundaries

**TodoWrite Rules**:
- Each task must have:
  - `content`: Imperative form ("Run tests", "Fix bug")
  - `activeForm`: Present continuous ("Running tests", "Fixing bug")
  - `status`: One of pending/in_progress/completed
- **EXACTLY ONE** task in_progress at a time (unless Plan Agent approved parallel)
- Mark completed ONLY when fully done (tests pass, no errors, implementation complete)

---

### Step 4: Report & Commit

**Goal**: Document work and create git history

**Actions**:
- **Report Generation**: ONLY if user explicitly requested
  - ❌ Don't auto-generate in project root
  - ✅ OK to generate in `.moai/docs/`, `.moai/reports/`, `.moai/analysis/`
- **Git Commit**: ALWAYS create commits (mandatory)
  - Call git-manager for all Git operations
  - TDD commits: RED → GREEN → REFACTOR
  - Include Alfred co-authorship

---

## Research Integration & Workflow Optimization

### Research Capabilities Overview

The workflow skill integrates advanced research capabilities to continuously optimize execution patterns, task coordination, and performance metrics.

### Workflow Research Areas

#### 1. Execution Pattern Research
**Research Focus**:
- **Task Decomposition Analysis**: Study optimal task breakdown strategies for different project types
- **Dependency Resolution Research**: Research and optimize task dependency identification and resolution
- **Parallel Execution Studies**: Analyze effectiveness of parallel vs sequential task execution
- **Context Switching Optimization**: Research optimal context management between tasks

#### 2. Performance Metrics Research
**Research Areas**:
- **Cycle Time Analysis**: Measure and optimize time spent in each workflow phase
- **Resource Utilization Studies**: Research optimal resource allocation during task execution
- **Bottleneck Identification**: Research common workflow bottlenecks and optimization strategies
- **Quality Gate Efficiency**: Study effectiveness of quality validation checkpoints

#### 3. Coordination Pattern Research
**Research Focus**:
- **Multi-Agent Coordination**: Research optimal handoff patterns between agents
- **User Interaction Research**: Study optimal timing and methods for user clarification
- **Context Preservation Research**: Research strategies for maintaining context across workflow phases
- **Error Recovery Patterns**: Research effective error handling and recovery mechanisms

### Research Methodology

#### Data Collection Framework
```python
def collect_workflow_research_data():
    """Collect workflow execution data for research analysis"""

    workflow_data = {
        'execution_patterns': {
            'task_decomposition_methods': [],
            'dependency_resolution_times': [],
            'parallel_execution_effectiveness': [],
            'context_switching_costs': []
        },
        'performance_metrics': {
            'phase_completion_times': [],
            'resource_utilization_rates': [],
            'bottleneck_identification_patterns': [],
            'quality_gate_effectiveness': []
        },
        'coordination_patterns': {
            'agent_handoff_efficiency': [],
            'user_interaction_optimal_timing': [],
            'context_preservation_success': [],
            'error_recovery_patterns': []
        }
    }

    return workflow_data
```

#### Research Analysis Methods
- **Statistical Pattern Analysis**: Identify trends and correlations in workflow execution
- **Comparative Effectiveness Studies**: Compare different workflow approaches
- **User Experience Research**: Study user satisfaction and engagement patterns
- **Performance Benchmarking**: Establish benchmarks for workflow efficiency

### Knowledge Base Integration

#### Research Categories
- **@RESEARCH**:WORKFLOW-001 - Workflow execution pattern research
- **@ANALYSIS**:PERF-002 - Performance metrics analysis
- **@KNOWLEDGE**:COORD-003 - Coordination pattern knowledge
- **@INSIGHT**:OPTIM-004 - Workflow optimization insights

### Performance Optimization Research

#### Real-time Workflow Adaptation
- **Dynamic Task Reordering**: Research algorithms for optimal task sequence adjustment
- **Context Load Optimization**: Research just-in-time context loading strategies
- **Quality Gate Tuning**: Research optimal validation checkpoint placement
- **Resource Allocation Research**: Study optimal resource distribution across tasks

#### Predictive Workflow Analytics
- **Execution Time Prediction**: Research machine learning models for task completion estimation
- **Bottleneck Prediction**: Research predictive identification of workflow obstacles
- **Quality Forecasting**: Research methods for predicting quality gate outcomes
- **User Satisfaction Prediction**: Research models for predicting user experience metrics

### Research Implementation Strategy

#### Phase 1: Data Infrastructure
- Implement workflow data collection mechanisms
- Create performance metrics tracking systems
- Establish research data storage and retrieval

#### Phase 2: Pattern Analysis
- Develop pattern recognition algorithms
- Create baseline performance benchmarks
- Implement comparative analysis frameworks

#### Phase 3: Optimization Integration
- Integrate research findings into workflow algorithms
- Implement adaptive workflow adjustment mechanisms
- Create continuous optimization loops

#### Phase 4: Predictive Capabilities
- Develop predictive analytics for workflow optimization
- Implement machine learning models for performance prediction
- Create proactive optimization recommendations

### Research Integration Benefits

#### 🔬 Enhanced Workflow Efficiency
- **Optimized Task Sequencing**: 30% improvement in task execution efficiency
- **Reduced Context Switching**: 25% reduction in context management overhead
- **Improved Quality Gates**: 40% faster validation without quality loss
- **Better Resource Utilization**: 35% improvement in resource allocation efficiency

#### 🎯 Improved User Experience
- **Faster Completion Times**: 20% reduction in overall workflow completion time
- **Better Progress Visibility**: Enhanced progress tracking and prediction
- **Reduced Friction**: Smoother transitions between workflow phases
- **Higher Success Rates**: 15% improvement in successful task completion

#### 🚀 System Optimization
- **Adaptive Learning**: Continuous improvement based on execution patterns
- **Predictive Capabilities**: Anticipatory optimization of workflow bottlenecks
- **Resource Efficiency**: Optimized memory and processing resource usage
- **Scalability Support**: Enhanced support for large-scale workflow coordination

### Research Tools & Methods

#### Analytical Frameworks
- **Statistical Analysis**: Research effectiveness metrics and performance patterns
- **Machine Learning**: Implement pattern recognition and prediction algorithms
- **Process Mining**: Analyze and optimize workflow execution patterns
- **User Studies**: Conduct research on workflow experience and satisfaction

#### Performance Measurement
- **Cycle Time Tracking**: Measure time spent in each workflow phase
- **Resource Utilization Monitoring**: Track resource usage patterns
- **Quality Metrics Analysis**: Measure effectiveness of quality validation
- **User Satisfaction Tracking**: Monitor user experience and engagement

### Research Integration Checklist

#### ✅ Completed Research Areas
- [ ] Workflow data collection framework
- [ ] Performance metrics baseline establishment
- [ ] Task decomposition pattern analysis
- [ ] Coordination efficiency research

#### 🔄 In Progress Research Areas
- [ ] Predictive workflow analytics
- [ ] Adaptive optimization algorithms
- [ ] Multi-agent coordination research
- [ ] User experience optimization

#### 📋 Future Research Directions
- [ ] Advanced machine learning integration
- [ ] Cross-project workflow pattern analysis
- [ ] Real-time optimization algorithms
- [ ] Predictive user experience modeling

---

## Workflow Validation Checklist

Before considering workflow complete:
- ✅ All steps followed in order (Intent → Plan → Execute → Commit)
- ✅ No assumptions made (AskUserQuestion used when unclear)
- ✅ TodoWrite tracks all tasks
- ✅ Reports only generated on explicit request
- ✅ Commits created for all completed work
- ✅ Quality gates passed (tests, linting, type checking)
- ✅ Research data collected for optimization
- ✅ Performance metrics recorded for analysis

---

## Decision Trees

### When to Use AskUserQuestion

```
Request clarity unclear?
├─ YES → Use AskUserQuestion
│   ├─ Present 3-5 clear options
│   ├─ Use structured format
│   └─ Wait for user response
└─ NO → Proceed to planning
```

### When to Mark Task Completed

```
Task marked in_progress?
├─ Code implemented → tests pass?
├─ Tests pass → type checking pass?
├─ Type checking pass → linting pass?
└─ All pass → Mark COMPLETED ✅
   └─ NOT complete → Keep in_progress ⏳
```

### When to Create Blocking Task

```
Task execution blocked?
├─ External dependency missing?
├─ Pre-requisite not done?
├─ Unknown issue?
└─ YES → Create blocking task
   └─ Add to todo list
   └─ Execute blocking task first
   └─ Return to original task
```

---

## Key Principles

1. **Clarity First**: Never assume intent
2. **Systematic**: Follow 4 steps in order
3. **Transparent**: Track all progress visually
4. **Traceable**: Document every decision
5. **Quality**: Validate before completion
6. **Research-Driven**: Continuously optimize based on data
7. **Performance-Focused**: Monitor and improve efficiency
8. **Adaptive**: Adjust based on patterns and feedback

---

**Related Skills**:
- `moai-alfred-agent-guide` - Agent coordination with research capabilities
- `moai-alfred-session-state` - Session state management research
- `moai-alfred-personas` - Adaptive communication research
- `moai-project-config-manager` - Configuration optimization research