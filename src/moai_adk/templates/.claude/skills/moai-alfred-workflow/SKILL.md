---
name: moai-alfred-workflow
version: 4.0.0
status: production
description: |
  Enterprise multi-agent workflow orchestration specialist. Master workflow 
  design, agent coordination, task delegation, and process automation with Context7 
  MCP integration and comprehensive monitoring. Build scalable, intelligent 
  workflow systems with fault tolerance and performance optimization.
allowed-tools: ["Read", "Write", "Edit", "Bash", "Glob", "WebFetch", "WebSearch"]
tags: ["workflow", "automation", "agents", "orchestration", "context7", "mcp", "multi-agent"]
---

# Alfred Workflow Orchestration

## Level 1: Quick Reference

### Core Capabilities
- **Multi-Agent Systems**: Coordinated agent workflows and delegation
- **Process Automation**: End-to-end workflow automation  
- **Task Orchestration**: Complex task scheduling and management
- **Context7 Integration**: 13,157+ code examples and documentation lookup
- **Monitoring**: Comprehensive workflow performance tracking

### Quick Setup

```python
# Basic workflow setup
from alfred_workflow import WorkflowEngine, Agent

engine = WorkflowEngine()
spec_agent = Agent("spec-builder", domain="requirements")
impl_agent = Agent("tdd-implementer", domain="development")
test_agent = Agent("quality-gate", domain="testing")

workflow = engine.create_workflow("feature_development")
workflow.add_stage("specification", spec_agent)
workflow.add_stage("implementation", impl_agent, depends_on=["specification"])
workflow.add_stage("testing", test_agent, depends_on=["implementation"])
result = engine.execute(workflow, input_data={"feature": "user auth"})
```

```python
# Context7 integration
from alfred_workflow import Context7Integration

context7 = Context7Integration()
examples = context7.search_code_examples(
    query="react authentication", language="javascript", framework="react"
)
best_practices = context7.get_best_practices(
    topic="database optimization", database="postgresql"
)
```

### Essential Patterns

| Pattern | Use Case | Benefit |
|---------|----------|---------|
| Sequential | Linear tasks | Predictable flow |
| Parallel | Independent tasks | Faster completion |
| Conditional | Decision-based | Adaptive workflows |
| Error Recovery | Fault tolerance | Reliable execution |

## Level 2: Core Implementation

### Architecture Overview

**WorkflowEngine**: Central orchestrator managing agents and workflows
**Agent System**: Specialized agents for different domains (spec-builder, tdd-implementer, quality-gate)
**Task Model**: Structured tasks with dependencies, priorities, and retry logic
**Template System**: Reusable workflow patterns (feature development, bug fix)

### Core Classes

```python
from dataclasses import dataclass, field
from enum import Enum
from datetime import datetime
from typing import Dict, List, Any, Optional

class TaskStatus(Enum):
    PENDING = "pending"
    RUNNING = "running" 
    COMPLETED = "completed"
    FAILED = "failed"

class AgentStatus(Enum):
    IDLE = "idle"
    BUSY = "busy"
    ERROR = "error"

@dataclass
class Task:
    id: str
    name: str
    description: str
    agent_type: str
    input_data: Dict[str, Any] = field(default_factory=dict)
    dependencies: List[str] = field(default_factory=list)
    status: TaskStatus = TaskStatus.PENDING
    retry_count: int = 0
    max_retries: int = 3
    result: Optional[Dict[str, Any]] = None

@dataclass  
class Agent:
    id: str
    name: str
    agent_type: str
    capabilities: List[str]
    status: AgentStatus = AgentStatus.IDLE
    current_task: Optional[Task] = None

class WorkflowEngine:
    def __init__(self, max_concurrent_tasks: int = 5):
        self.agents: Dict[str, Agent] = {}
        self.workflows: Dict[str, 'Workflow'] = {}
        self.max_concurrent_tasks = max_concurrent_tasks
        
    def register_agent(self, agent: Agent) -> None:
        self.agents[agent.id] = agent
        
    def create_workflow(self, name: str, description: str = "") -> 'Workflow':
        workflow = Workflow(name=name, description=description, engine=self)
        self.workflows[name] = workflow
        return workflow
        
    async def execute_workflow(self, workflow: 'Workflow') -> Dict[str, Any]:
        results = {}
        for task in workflow.get_execution_order():
            await self._wait_for_dependencies(task, results)
            result = await self.execute_task(task)
            results[task.id] = result
        return results

@dataclass
class Workflow:
    name: str
    description: str
    engine: WorkflowEngine
    tasks: List[Task] = field(default_factory=list)
    status: str = "created"
    created_at: datetime = field(default_factory=datetime.now)
    
    def add_stage(self, stage_name: str, agent_type: str,
                  input_data: Dict[str, Any] = None,
                  depends_on: List[str] = None) -> Task:
        task = Task(
            id=f"{stage_name}_{len(self.tasks)}",
            name=stage_name,
            description=f"Workflow stage: {stage_name}",
            agent_type=agent_type,
            input_data=input_data or {},
            dependencies=depends_on or []
        )
        self.tasks.append(task)
        return task
```

### Context7 Integration

```python
class Context7Integration:
    def __init__(self, mcp_servers: List[str] = None):
        self.mcp_servers = mcp_servers or []
        self.cache = {}
        self.cache_ttl = 3600
        
    async def search_code_examples(self, query: str, language: str = None,
                                 framework: str = None, limit: int = 10) -> List[Dict]:
        cache_key = f"code_examples_{query}_{language}_{framework}_{limit}"
        
        # Check cache first
        if cache_key in self.cache:
            cached = self.cache[cache_key]
            if time.time() - cached['timestamp'] < self.cache_ttl:
                return cached['data']
        
        # Execute Context7 search
        results = await self._context7_search(query, language, framework, limit)
        
        # Cache results
        self.cache[cache_key] = {'data': results, 'timestamp': time.time()}
        return results
    
    async def get_best_practices(self, topic: str, domain: str = None) -> Dict:
        search_query = f"best practices {topic}"
        if domain:
            search_query += f" {domain}"
        
        results = await self._context7_search(search_query, limit=5)
        return results[0] if results else {}
```

### Workflow Templates

```python
from abc import ABC, abstractmethod

class WorkflowTemplate(ABC):
    def __init__(self, name: str, description: str):
        self.name = name
        self.description = description
    
    @abstractmethod
    def create_workflow(self, engine: WorkflowEngine, config: Dict) -> Workflow:
        pass

class FeatureDevelopmentTemplate(WorkflowTemplate):
    def __init__(self):
        super().__init__("feature_development", "End-to-end feature development")
    
    def create_workflow(self, engine: WorkflowEngine, config: Dict) -> Workflow:
        workflow = engine.create_workflow(
            name=f"feature_{config.get('feature_name', 'unknown')}",
            description=config.get('feature_description', '')
        )
        
        # Specification stage
        spec_task = workflow.add_stage("specification", "spec-builder", {
            "feature_description": config.get('feature_description', ''),
            "requirements": config.get('requirements', [])
        })
        
        # Implementation stage  
        impl_task = workflow.add_stage("implementation", "tdd-implementer", {
            "spec_id": spec_task.id,
            "technology_stack": config.get('technology_stack', [])
        }, depends_on=[spec_task.id])
        
        # Testing stage
        workflow.add_stage("testing", "quality-gate", {
            "implementation_id": impl_task.id,
            "test_types": config.get('test_types', ['unit', 'integration'])
        }, depends_on=[impl_task.id])
        
        return workflow
```

## Level 3: Advanced Features

### Workflow Scheduling

```python
from enum import Enum
import uuid
import asyncio

class WorkflowPriority(Enum):
    LOW = 1
    MEDIUM = 2  
    HIGH = 3
    CRITICAL = 4

class WorkflowScheduler:
    def __init__(self, workflow_engine: WorkflowEngine):
        self.workflow_engine = workflow_engine
        self.workflow_queue = asyncio.Queue()
        self.template_manager = WorkflowTemplateManager()
        
    async def submit_workflow(self, template_name: str, config: Dict,
                            priority: WorkflowPriority = WorkflowPriority.MEDIUM,
                            scheduled_time: Optional[datetime] = None) -> str:
        workflow_id = str(uuid.uuid4())
        
        workflow = self.template_manager.create_workflow_from_template(
            template_name, self.workflow_engine, config
        )
        
        if not workflow:
            raise ValueError(f"Unknown template: {template_name}")
        
        workflow_item = {
            'workflow_id': workflow_id,
            'workflow': workflow,
            'priority': priority,
            'scheduled_time': scheduled_time or datetime.now()
        }
        
        await self.workflow_queue.put((-priority.value, workflow_item))
        return workflow_id
```

### Error Handling & Recovery

```python
class WorkflowErrorHandler:
    def __init__(self, workflow_engine: WorkflowEngine):
        self.workflow_engine = workflow_engine
    
    async def handle_task_failure(self, task: Task, error: Exception) -> bool:
        if task.retry_count < task.max_retries:
            return await self._retry_with_backoff(task, error)
        
        if self._is_retriable_error(error):
            return await self._retry_with_backoff(task, error)
        else:
            return await self._require_manual_intervention(task, error)
    
    async def _retry_with_backoff(self, task: Task, error: Exception) -> bool:
        delay = 2 ** task.retry_count
        task.retry_count += 1
        await asyncio.sleep(delay)
        
        try:
            await self.workflow_engine.execute_task(task)
            return True
        except Exception as e:
            return await self.handle_task_failure(task, e)
    
    def _is_retriable_error(self, error: Exception) -> bool:
        retriable_errors = ['TimeoutError', 'ConnectionError', 'TemporaryFailure']
        return type(error).__name__ in retriable_errors
```

### Performance Monitoring

```python
class WorkflowMetrics:
    def __init__(self):
        self.metrics = {
            'workflows_completed': 0,
            'workflows_failed': 0, 
            'total_execution_time': 0.0,
            'task_completion_times': []
        }
    
    def record_workflow_completion(self, workflow_id: str, execution_time: float, status: str):
        if status == 'completed':
            self.metrics['workflows_completed'] += 1
        else:
            self.metrics['workflows_failed'] += 1
        
        self.metrics['total_execution_time'] += execution_time
    
    def get_performance_summary(self) -> Dict[str, Any]:
        completed = self.metrics['workflows_completed']
        failed = self.metrics['workflows_failed']
        total = completed + failed
        
        if total == 0:
            return {'error': 'No workflows executed'}
        
        avg_time = self.metrics['total_execution_time'] / total if total > 0 else 0
        success_rate = (completed / total) * 100 if total > 0 else 0
        
        return {
            'total_workflows': total,
            'success_rate': f"{success_rate:.2f}%",
            'average_execution_time': f"{avg_time:.2f}s",
            'workflows_completed': completed,
            'workflows_failed': failed
        }
```

## Level 4: Reference & Integration

### When to Use

**Use for:**
- Multi-agent workflows requiring coordination
- Automated CI/CD pipelines with intelligent decision making  
- Enterprise process automation with error handling
- Tasks requiring Context7 integration for best practices
- Workflows needing comprehensive monitoring

**Avoid for:**
- Simple single-agent tasks (use specific domain skills)
- Basic automation without coordination needs
- Quick prototyping without enterprise requirements

### Common Usage Patterns

```python
# Feature Development Workflow
workflow_id = await scheduler.submit_workflow(
    "feature_development",
    {
        "feature_name": "user_authentication",
        "feature_description": "Implement secure user login",
        "requirements": ["JWT auth", "Password hashing"],
        "acceptance_criteria": ["Users can login securely"]
    }
)

# Bug Fix Workflow  
bug_workflow_id = await scheduler.submit_workflow(
    "bug_fix",
    {
        "bug_id": "BUG-001",
        "bug_description": "Login fails with invalid credentials",
        "error_logs": ["AuthException at line 42"],
        "reproduction_steps": ["Enter invalid password"]
    },
    priority=WorkflowPriority.HIGH
)
```

### Security & Compliance

**TRUST Principles Applied**:
- **Test First**: All workflows include validation stages and quality gates
- **Readable**: Clear workflow structure with comprehensive logging
- **Unified**: Consistent patterns across all workflow templates  
- **Secured**: Built-in input validation and security checks
- **Traceable**: Complete audit trail with workflow execution history

**Enterprise Security Features**:
- **Input Validation**: Automated security scanning of workflow inputs
- **Access Control**: Role-based access to workflow operations
- **Audit Logging**: Complete execution history with security events
- **Data Encryption**: Sensitive data protection in transit and at rest

### Related Skills

- `moai-alfred-agent-guide` - Agent selection and delegation patterns
- `moai-alfred-spec-authoring` - SPEC creation workflows
- `moai-essentials-debug` - Error handling and troubleshooting
- `moai-foundation-trust` - Security and compliance principles
- `moai-domain-backend` - Backend-specific workflow patterns
- `moai-domain-testing` - Testing workflow integration

---

**Enterprise v4.0 Compliance**: Progressive disclosure with comprehensive error handling, security controls, and monitoring.
**Last Updated**: 2025-11-13  
**Dependencies**: Context7 MCP integration, Alfred agent system
**See Also**: [examples.md](./examples.md) for detailed usage examples
