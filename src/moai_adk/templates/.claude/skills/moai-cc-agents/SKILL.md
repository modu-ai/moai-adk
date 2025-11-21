---
name: moai-cc-agents
description: Claude Code Agents architecture, lifecycle management, and multi-agent orchestration patterns with Context7 integration for latest agent design patterns.
---

## Quick Reference (30 seconds)

Claude Code **Agents** are specialized AI workers that execute specific tasks through delegation patterns. Each agent has a defined role, capabilities, and communication protocols.

**Core Capabilities**:
- Task delegation and specialization
- Multi-agent coordination workflows
- Agent lifecycle management (create, activate, suspend, retire)
- Inter-agent communication protocols
- Resource allocation and load balancing

**When to Use**:
- Complex tasks requiring specialized expertise
- Parallel workflow orchestration
- Distributed problem-solving
- Large-scale automation requiring coordination

**Key Concept**: Agents don't execute tools directly - they delegate to Skills that encapsulate domain knowledge.

---

## Implementation Guide

### Agent Architecture Fundamentals

**Agent Types & Roles**:

```
Specialist Agents:
  └─ Domain-specific expertise (backend-expert, frontend-expert, security-expert)
  └─ Deep knowledge in single area
  └─ High accuracy, narrow scope

Coordinator Agents:
  └─ Workflow orchestration (alfred, project-manager, implementation-planner)
  └─ Task routing and delegation
  └─ Multi-agent synchronization

Validation Agents:
  └─ Quality assurance (quality-gate, trust-checker, code-reviewer)
  └─ TRUST 5 compliance verification
  └─ Output validation and approval

Automation Agents:
  └─ Repetitive task handling (doc-syncer, git-manager, format-expert)
  └─ Batch operations
  └─ Scheduled maintenance
```

### Agent Lifecycle Management

**Lifecycle States**:

```
1. CREATED:
   └─ Agent definition registered
   └─ Initial configuration loaded
   └─ Not yet active

2. ACTIVE:
   └─ Ready to accept tasks
   └─ Processing workload
   └─ Can delegate to Skills

3. SUSPENDED:
   └─ Temporarily paused
   └─ Resources released
   └─ Can be reactivated

4. RETIRED:
   └─ No longer accepting tasks
   └─ Historical data preserved
   └─ Replacement agent may exist
```

**Lifecycle Operations**:

```python
# Agent activation pattern
def activate_agent(agent_name: str, config: dict) -> AgentInstance:
    """
    Activate agent with configuration.
    
    Args:
        agent_name: Unique agent identifier
        config: Agent-specific configuration
    
    Returns:
        Activated agent instance
    """
    agent = load_agent_definition(agent_name)
    agent.configure(config)
    agent.load_skills()  # Load required Skills
    agent.state = AgentState.ACTIVE
    register_agent(agent)
    return agent

# Agent delegation pattern
def delegate_task(agent: AgentInstance, task: Task) -> TaskResult:
    """
    Delegate task to specialized agent.
    
    Args:
        agent: Target agent instance
        task: Task specification
    
    Returns:
        Task execution result
    """
    if not agent.can_handle(task):
        raise AgentCapabilityError(f"{agent.name} cannot handle {task.type}")
    
    # Agent invokes Skills for execution
    skills = agent.select_skills(task)
    result = agent.execute_with_skills(task, skills)
    
    return result
```

### Multi-Agent Coordination Patterns

**Pattern 1: Sequential Delegation**

```
Task Flow:
  User Request
    ↓
  Coordinator Agent (alfred)
    ↓
  Specialist Agent 1 (spec-builder)
    ↓
  Specialist Agent 2 (tdd-implementer)
    ↓
  Validation Agent (quality-gate)
    ↓
  Completed Result
```

**Implementation**:

```python
async def sequential_workflow(request: UserRequest) -> WorkflowResult:
    """Execute tasks in sequence with agent delegation."""
    
    # Step 1: Specification design
    spec_result = await delegate_to(
        agent="spec-builder",
        task={"type": "spec_creation", "input": request}
    )
    
    # Step 2: Implementation
    impl_result = await delegate_to(
        agent="tdd-implementer",
        task={"type": "implementation", "spec": spec_result}
    )
    
    # Step 3: Quality validation
    qa_result = await delegate_to(
        agent="quality-gate",
        task={"type": "validation", "implementation": impl_result}
    )
    
    return WorkflowResult(spec=spec_result, impl=impl_result, qa=qa_result)
```

**Pattern 2: Parallel Execution**

```
Task Flow:
  User Request
    ↓
  Coordinator Agent
    ├─> Agent A (backend-expert)
    ├─> Agent B (frontend-expert)
    └─> Agent C (database-expert)
    ↓
  Merge Results
    ↓
  Completed Result
```

**Implementation**:

```python
async def parallel_workflow(request: UserRequest) -> WorkflowResult:
    """Execute tasks in parallel with agent delegation."""
    
    tasks = [
        delegate_to(agent="backend-expert", task={"type": "api", "spec": request}),
        delegate_to(agent="frontend-expert", task={"type": "ui", "spec": request}),
        delegate_to(agent="database-expert", task={"type": "schema", "spec": request})
    ]
    
    # Wait for all agents to complete
    results = await asyncio.gather(*tasks)
    
    # Merge results with conflict resolution
    merged = merge_agent_outputs(results)
    
    return WorkflowResult(merged=merged, individual=results)
```

**Pattern 3: Conditional Routing**

```
Task Flow:
  User Request
    ↓
  Coordinator Agent
    ├─> If security-critical → security-expert
    ├─> If performance-critical → performance-engineer
    └─> Else → backend-expert
```

**Implementation**:

```python
async def conditional_workflow(request: UserRequest) -> WorkflowResult:
    """Route task to appropriate agent based on requirements."""
    
    # Analyze request characteristics
    requirements = analyze_requirements(request)
    
    # Select appropriate agent
    if requirements.security_level >= SecurityLevel.HIGH:
        agent = "security-expert"
    elif requirements.performance_critical:
        agent = "performance-engineer"
    else:
        agent = "backend-expert"
    
    # Delegate to selected agent
    result = await delegate_to(agent=agent, task=request)
    
    return WorkflowResult(agent_used=agent, result=result)
```

### Context7 Integration for Agent Patterns

**Fetch Latest Agent Design Patterns**:

```python
async def get_agent_design_patterns() -> AgentPatterns:
    """Fetch latest agent design patterns from Context7."""
    
    # Get Claude Code agent patterns
    patterns = await context7.get_library_docs(
        context7_library_id="/anthropic/claude-code",
        topic="agent architecture orchestration patterns 2025",
        tokens=3000
    )
    
    return AgentPatterns(
        coordination_patterns=patterns["coordination"],
        lifecycle_patterns=patterns["lifecycle"],
        best_practices=patterns["best_practices"]
    )
```

---

## Advanced Patterns

### Agent Communication Protocols

**Message Passing**:

```python
class AgentMessage:
    """Standard message format for inter-agent communication."""
    
    sender: str          # Originating agent
    recipient: str       # Target agent
    message_type: str    # Command, query, response, notification
    payload: dict        # Message content
    timestamp: datetime
    correlation_id: str  # For request-response tracking
    
def send_message(msg: AgentMessage) -> None:
    """Send message to another agent via message bus."""
    message_bus.publish(msg)

def receive_messages(agent_name: str) -> Iterator[AgentMessage]:
    """Subscribe to messages for specific agent."""
    return message_bus.subscribe(agent_name)
```

**Event-Driven Coordination**:

```python
# Agent publishes event when task complete
@agent.on_task_complete
def publish_completion_event(task: Task, result: TaskResult):
    """Publish completion event for downstream agents."""
    
    event = AgentEvent(
        event_type="task_completed",
        source_agent=agent.name,
        task_id=task.id,
        result=result,
        timestamp=datetime.now()
    )
    
    event_bus.publish(event)

# Downstream agent subscribes to events
@agent.on_event("task_completed")
async def handle_upstream_completion(event: AgentEvent):
    """React to upstream agent completion."""
    
    if event.task_id in agent.waiting_tasks:
        # Continue workflow with upstream result
        await agent.continue_task(event.task_id, event.result)
```

### Resource Allocation & Load Balancing

**Agent Resource Management**:

```python
class AgentResourceManager:
    """Manage compute resources across agents."""
    
    def __init__(self, total_budget: int):
        self.total_budget = total_budget
        self.allocations = {}
    
    def allocate(self, agent_name: str, requested: int) -> int:
        """Allocate resources to agent."""
        
        available = self.total_budget - sum(self.allocations.values())
        
        if requested > available:
            # Throttle or queue request
            return self.throttle_request(agent_name, requested, available)
        
        self.allocations[agent_name] = requested
        return requested
    
    def release(self, agent_name: str) -> None:
        """Release agent resources."""
        
        if agent_name in self.allocations:
            del self.allocations[agent_name]
```

**Load Balancing Pattern**:

```python
async def distribute_tasks(tasks: list[Task], agent_pool: list[str]) -> dict:
    """Distribute tasks across available agents."""
    
    # Group tasks by agent capability
    task_groups = group_by_capability(tasks)
    
    # Select agents with lowest current load
    assignments = {}
    for capability, task_list in task_groups.items():
        available_agents = [a for a in agent_pool if has_capability(a, capability)]
        
        # Round-robin or least-loaded assignment
        for task in task_list:
            agent = select_least_loaded(available_agents)
            assignments.setdefault(agent, []).append(task)
    
    # Execute tasks in parallel across agents
    results = await execute_distributed(assignments)
    
    return results
```

### Error Handling & Recovery

**Agent Failure Recovery**:

```python
async def execute_with_retry(agent: str, task: Task, max_retries: int = 3) -> TaskResult:
    """Execute task with automatic retry on agent failure."""
    
    for attempt in range(max_retries):
        try:
            result = await delegate_to(agent=agent, task=task)
            return result
        
        except AgentUnavailable:
            # Agent crashed or suspended
            logger.warning(f"Agent {agent} unavailable (attempt {attempt+1})")
            
            if attempt < max_retries - 1:
                # Try backup agent or wait for recovery
                backup_agent = get_backup_agent(agent)
                if backup_agent:
                    agent = backup_agent
                else:
                    await asyncio.sleep(5)  # Wait for recovery
            else:
                raise AgentExecutionFailed(f"Agent {agent} failed after {max_retries} attempts")
        
        except Exception as e:
            logger.error(f"Agent {agent} error: {e}")
            raise
```

---

## Works Well With

- `moai-cc-skills` - Agents delegate to Skills for execution
- `moai-cc-hooks` - Event-driven agent coordination
- `moai-alfred-agent-guide` - Agent selection patterns
- `moai-context7-integration` - Latest agent design patterns

---

## Changelog

- **v3.0.0** (2025-11-21): Enterprise 4-level progressive disclosure, Context7 integration, multi-agent orchestration patterns
- **v2.0.0** (2025-11-11): Added complete metadata, agent architecture patterns
- **v1.0.0** (2025-10-22): Initial agents system

---

**End of Skill** | Updated 2025-11-21
