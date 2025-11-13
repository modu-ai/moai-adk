---
name: moai-alfred-workflow
version: 4.0.0
status: production
description: |
  Enterprise multi-agent workflow orchestration specialist. Master workflow 
  design, agent coordination, task delegation, and process automation. Build 
  scalable, intelligent workflow systems with Context7 MCP integration and 
  comprehensive monitoring. Design complex agent interactions with fault tolerance 
  and performance optimization.
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
- **Workflow Engines**: Airflow, Prefect, Dagster integration
- **Monitoring**: Comprehensive workflow performance tracking

### Quick Setup Examples

```python
# Basic Alfred workflow setup
from alfred_workflow import WorkflowEngine, Agent, Task

# Create workflow engine
engine = WorkflowEngine()

# Define agents
spec_agent = Agent("spec-builder", domain="requirements")
impl_agent = Agent("tdd-implementer", domain="development")
test_agent = Agent("quality-gate", domain="testing")

# Create workflow
workflow = engine.create_workflow("feature_development")

# Add stages
workflow.add_stage("specification", spec_agent)
workflow.add_stage("implementation", impl_agent, depends_on=["specification"])
workflow.add_stage("testing", test_agent, depends_on=["implementation"])

# Execute workflow
result = engine.execute(workflow, input_data={"feature": "user authentication"})
```

```python
# Context7 MCP integration
from alfred_workflow import Context7Integration

# Initialize Context7
context7 = Context7Integration()

# Search for code examples
examples = context7.search_code_examples(
    query="react authentication",
    language="javascript",
    framework="react"
)

# Get best practices
best_practices = context7.get_best_practices(
    topic="database optimization",
    database="postgresql"
)

# Access documentation
docs = context7.get_documentation(
    library="tensorflow",
    version="2.20.0",
    topic="neural networks"
)
```

## Level 2: Practical Implementation

### Multi-Agent Workflow Architecture

#### 1. Agent-Based Workflow Engine

```python
import asyncio
from typing import Dict, List, Any, Optional, Callable
from dataclasses import dataclass, field
from enum import Enum
import json
import time
import logging
from datetime import datetime

class AgentStatus(Enum):
    IDLE = "idle"
    BUSY = "busy"
    ERROR = "error"
    COMPLETED = "completed"

class TaskStatus(Enum):
    PENDING = "pending"
    RUNNING = "running"
    COMPLETED = "completed"
    FAILED = "failed"
    CANCELLED = "cancelled"

@dataclass
class Task:
    id: str
    name: str
    description: str
    agent_type: str
    input_data: Dict[str, Any] = field(default_factory=dict)
    dependencies: List[str] = field(default_factory=list)
    priority: int = 0
    timeout: Optional[int] = None
    retry_count: int = 0
    max_retries: int = 3
    status: TaskStatus = TaskStatus.PENDING
    created_at: datetime = field(default_factory=datetime.now)
    started_at: Optional[datetime] = None
    completed_at: Optional[datetime] = None
    result: Optional[Dict[str, Any]] = None
    error_message: Optional[str] = None

@dataclass
class Agent:
    id: str
    name: str
    agent_type: str
    capabilities: List[str]
    status: AgentStatus = AgentStatus.IDLE
    current_task: Optional[Task] = None
    completed_tasks: List[Task] = field(default_factory=list)
    performance_metrics: Dict[str, float] = field(default_factory=dict)

class WorkflowEngine:
    def __init__(self, max_concurrent_tasks: int = 5):
        self.agents: Dict[str, Agent] = {}
        self.workflows: Dict[str, 'Workflow'] = {}
        self.task_queue: asyncio.Queue = asyncio.Queue()
        self.max_concurrent_tasks = max_concurrent_tasks
        self.running_tasks: Dict[str, asyncio.Task] = {}
        self.context7_integration = None
        self.logger = logging.getLogger(__name__)
        
    def register_agent(self, agent: Agent) -> None:
        """Register an agent with the workflow engine"""
        self.agents[agent.id] = agent
        self.logger.info(f"Registered agent: {agent.name} ({agent.agent_type})")
    
    def create_workflow(self, name: str, description: str = "") -> 'Workflow':
        """Create a new workflow"""
        workflow = Workflow(name=name, description=description, engine=self)
        self.workflows[name] = workflow
        return workflow
    
    async def execute_task(self, task: Task) -> Dict[str, Any]:
        """Execute a single task"""
        
        # Find suitable agent
        suitable_agent = self._find_suitable_agent(task)
        if not suitable_agent:
            raise ValueError(f"No suitable agent found for task: {task.name}")
        
        # Update task status
        task.status = TaskStatus.RUNNING
        task.started_at = datetime.now()
        suitable_agent.status = AgentStatus.BUSY
        suitable_agent.current_task = task
        
        try:
            # Execute task with timeout
            if task.timeout:
                result = await asyncio.wait_for(
                    self._execute_agent_task(suitable_agent, task),
                    timeout=task.timeout
                )
            else:
                result = await self._execute_agent_task(suitable_agent, task)
            
            # Update task and agent status
            task.status = TaskStatus.COMPLETED
            task.completed_at = datetime.now()
            task.result = result
            suitable_agent.status = AgentStatus.IDLE
            suitable_agent.current_task = None
            suitable_agent.completed_tasks.append(task)
            
            # Update performance metrics
            execution_time = (task.completed_at - task.started_at).total_seconds()
            suitable_agent.performance_metrics['avg_execution_time'] = (
                suitable_agent.performance_metrics.get('avg_execution_time', 0) * 0.8 + execution_time * 0.2
            )
            suitable_agent.performance_metrics['success_rate'] = (
                len(suitable_agent.completed_tasks) / 
                max(1, len(suitable_agent.completed_tasks) + 
                    suitable_agent.performance_metrics.get('failed_tasks', 0))
            )
            
            self.logger.info(f"Task {task.name} completed by agent {suitable_agent.name}")
            return result
            
        except Exception as e:
            # Handle task failure
            task.status = TaskStatus.FAILED
            task.error_message = str(e)
            suitable_agent.status = AgentStatus.ERROR
            suitable_agent.current_task = None
            suitable_agent.performance_metrics['failed_tasks'] = (
                suitable_agent.performance_metrics.get('failed_tasks', 0) + 1
            )
            
            self.logger.error(f"Task {task.name} failed: {str(e)}")
            
            # Retry logic
            if task.retry_count < task.max_retries:
                task.retry_count += 1
                task.status = TaskStatus.PENDING
                await asyncio.sleep(2 ** task.retry_count)  # Exponential backoff
                return await self.execute_task(task)
            
            raise e
    
    def _find_suitable_agent(self, task: Task) -> Optional[Agent]:
        """Find a suitable agent for the task"""
        
        for agent in self.agents.values():
            if (agent.agent_type == task.agent_type and 
                agent.status == AgentStatus.IDLE and
                any(capability in task.description.lower() for capability in agent.capabilities)):
                return agent
        
        # Fallback: find any idle agent of the right type
        for agent in self.agents.values():
            if agent.agent_type == task.agent_type and agent.status == AgentStatus.IDLE:
                return agent
        
        return None
    
    async def _execute_agent_task(self, agent: Agent, task: Task) -> Dict[str, Any]:
        """Execute task using agent capabilities"""
        
        # Simulate agent execution (in real implementation, this would call the actual agent)
        await asyncio.sleep(1)  # Simulate work
        
        # Mock result based on task type
        result = {
            'task_id': task.id,
            'agent_id': agent.id,
            'execution_time': time.time(),
            'status': 'success',
            'output': f"Task '{task.name}' completed successfully by {agent.name}"
        }
        
        # Add Context7 integration if available
        if self.context7_integration and 'documentation' in task.name.lower():
            result['context7_data'] = await self.context7_integration.get_relevant_info(
                query=task.description,
                agent_type=agent.agent_type
            )
        
        return result
    
    async def execute_workflow(self, workflow: 'Workflow') -> Dict[str, Any]:
        """Execute an entire workflow"""
        
        self.logger.info(f"Executing workflow: {workflow.name}")
        workflow.status = "running"
        workflow.started_at = datetime.now()
        
        try:
            # Build dependency graph
            dependency_graph = self._build_dependency_graph(workflow.tasks)
            
            # Execute tasks in dependency order
            executed_tasks = []
            results = {}
            
            while len(executed_tasks) < len(workflow.tasks):
                # Find tasks that can be executed (no unmet dependencies)
                ready_tasks = [
                    task for task in workflow.tasks 
                    if (task.status == TaskStatus.PENDING and 
                        all(dep in [t.id for t in executed_tasks] for dep in task.dependencies))
                ]
                
                if not ready_tasks:
                    raise ValueError("Circular dependency detected or no tasks ready")
                
                # Execute ready tasks concurrently (up to max_concurrent_tasks)
                semaphore = asyncio.Semaphore(self.max_concurrent_tasks)
                
                async def execute_with_semaphore(task):
                    async with semaphore:
                        return await self.execute_task(task)
                
                # Execute batch of tasks
                batch_results = await asyncio.gather(
                    *[execute_with_semaphore(task) for task in ready_tasks],
                    return_exceptions=True
                )
                
                # Process results
                for task, result in zip(ready_tasks, batch_results):
                    if isinstance(result, Exception):
                        task.status = TaskStatus.FAILED
                        task.error_message = str(result)
                        raise result
                    else:
                        executed_tasks.append(task)
                        results[task.id] = result
            
            workflow.status = "completed"
            workflow.completed_at = datetime.now()
            
            final_result = {
                'workflow_id': workflow.name,
                'status': workflow.status,
                'execution_time': (workflow.completed_at - workflow.started_at).total_seconds(),
                'tasks_completed': len(executed_tasks),
                'results': results
            }
            
            self.logger.info(f"Workflow {workflow.name} completed successfully")
            return final_result
            
        except Exception as e:
            workflow.status = "failed"
            workflow.error_message = str(e)
            self.logger.error(f"Workflow {workflow.name} failed: {str(e)}")
            raise e
    
    def _build_dependency_graph(self, tasks: List[Task]) -> Dict[str, List[str]]:
        """Build dependency graph from tasks"""
        graph = {}
        for task in tasks:
            graph[task.id] = task.dependencies
        return graph

@dataclass
class Workflow:
    name: str
    description: str
    engine: WorkflowEngine
    tasks: List[Task] = field(default_factory=list)
    status: str = "created"
    created_at: datetime = field(default_factory=datetime.now)
    started_at: Optional[datetime] = None
    completed_at: Optional[datetime] = None
    error_message: Optional[str] = None
    
    def add_task(self, task: Task) -> None:
        """Add a task to the workflow"""
        self.tasks.append(task)
    
    def add_stage(self, stage_name: str, agent_type: str, 
                  input_data: Dict[str, Any] = None,
                  depends_on: List[str] = None) -> Task:
        """Add a stage (task) to the workflow"""
        
        task = Task(
            id=f"{stage_name}_{len(self.tasks)}",
            name=stage_name,
            description=f"Workflow stage: {stage_name}",
            agent_type=agent_type,
            input_data=input_data or {},
            dependencies=depends_on or []
        )
        
        self.add_task(task)
        return task
    
    async def execute(self) -> Dict[str, Any]:
        """Execute the workflow"""
        return await self.engine.execute_workflow(self)
```

#### 2. Context7 MCP Integration

```python
import asyncio
from typing import Dict, List, Any, Optional
import json

class Context7Integration:
    def __init__(self, mcp_servers: List[str] = None):
        self.mcp_servers = mcp_servers or []
        self.cache = {}
        self.cache_ttl = 3600  # 1 hour
        self.logger = logging.getLogger(__name__)
        
    async def search_code_examples(self, query: str, language: str = None,
                                 framework: str = None, limit: int = 10) -> List[Dict[str, Any]]:
        """Search for code examples using Context7"""
        
        cache_key = f"code_examples_{query}_{language}_{framework}_{limit}"
        
        # Check cache
        if cache_key in self.cache:
            cached_result = self.cache[cache_key]
            if time.time() - cached_result['timestamp'] < self.cache_ttl:
                return cached_result['data']
        
        try:
            # Build search query
            search_terms = [query]
            if language:
                search_terms.append(f"language:{language}")
            if framework:
                search_terms.append(f"framework:{framework}")
            
            search_query = " ".join(search_terms)
            
            # Simulate Context7 MCP call (in real implementation)
            # This would use the actual MCP tools
            code_examples = await self._mock_context7_search(
                query=search_query,
                content_type="code_examples",
                limit=limit
            )
            
            # Cache results
            self.cache[cache_key] = {
                'data': code_examples,
                'timestamp': time.time()
            }
            
            return code_examples
            
        except Exception as e:
            self.logger.error(f"Error searching code examples: {str(e)}")
            return []
    
    async def get_best_practices(self, topic: str, domain: str = None,
                               technology: str = None) -> Dict[str, Any]:
        """Get best practices for a specific topic"""
        
        cache_key = f"best_practices_{topic}_{domain}_{technology}"
        
        if cache_key in self.cache:
            cached_result = self.cache[cache_key]
            if time.time() - cached_result['timestamp'] < self.cache_ttl:
                return cached_result['data']
        
        try:
            # Build search terms
            search_terms = [f"{topic} best practices"]
            if domain:
                search_terms.append(domain)
            if technology:
                search_terms.append(technology)
            
            best_practices = await self._mock_context7_search(
                query=" ".join(search_terms),
                content_type="best_practices",
                limit=5
            )
            
            # Process and structure best practices
            structured_practices = {
                'topic': topic,
                'domain': domain,
                'technology': technology,
                'practices': best_practices,
                'guidelines': self._extract_guidelines(best_practices),
                'code_examples': self._extract_code_examples(best_practices)
            }
            
            # Cache results
            self.cache[cache_key] = {
                'data': structured_practices,
                'timestamp': time.time()
            }
            
            return structured_practices
            
        except Exception as e:
            self.logger.error(f"Error getting best practices: {str(e)}")
            return {}
    
    async def get_documentation(self, library: str, version: str = None,
                              topic: str = None) -> Dict[str, Any]:
        """Get documentation for a specific library"""
        
        cache_key = f"docs_{library}_{version}_{topic}"
        
        if cache_key in self.cache:
            cached_result = self.cache[cache_key]
            if time.time() - cached_result['timestamp'] < self.cache_ttl:
                return cached_result['data']
        
        try:
            search_terms = [library]
            if version:
                search_terms.append(f"version:{version}")
            if topic:
                search_terms.append(topic)
            
            documentation = await self._mock_context7_search(
                query=" ".join(search_terms),
                content_type="documentation",
                limit=10
            )
            
            # Structure documentation
            structured_docs = {
                'library': library,
                'version': version,
                'topic': topic,
                'content': documentation,
                'api_reference': self._extract_api_reference(documentation),
                'tutorials': self._extract_tutorials(documentation),
                'examples': self._extract_code_examples(documentation)
            }
            
            # Cache results
            self.cache[cache_key] = {
                'data': structured_docs,
                'timestamp': time.time()
            }
            
            return structured_docs
            
        except Exception as e:
            self.logger.error(f"Error getting documentation: {str(e)}")
            return {}
    
    async def get_relevant_info(self, query: str, agent_type: str = None,
                              context: Dict[str, Any] = None) -> Dict[str, Any]:
        """Get relevant information for agent tasks"""
        
        # Determine what type of information to search for based on agent type
        search_strategy = self._determine_search_strategy(agent_type, query)
        
        results = {}
        
        if search_strategy['search_code_examples']:
            results['code_examples'] = await self.search_code_examples(
                query=query,
                language=search_strategy.get('language'),
                framework=search_strategy.get('framework')
            )
        
        if search_strategy['search_best_practices']:
            results['best_practices'] = await self.get_best_practices(
                topic=search_strategy.get('topic', query),
                domain=search_strategy.get('domain'),
                technology=search_strategy.get('technology')
            )
        
        if search_strategy['search_documentation']:
            results['documentation'] = await self.get_documentation(
                library=search_strategy.get('library'),
                version=search_strategy.get('version'),
                topic=search_strategy.get('topic')
            )
        
        return results
    
    def _determine_search_strategy(self, agent_type: str, query: str) -> Dict[str, Any]:
        """Determine search strategy based on agent type and query"""
        
        strategy = {
            'search_code_examples': True,
            'search_best_practices': True,
            'search_documentation': True
        }
        
        # Agent-specific strategies
        if agent_type == "tdd-implementer":
            strategy.update({
                'language': self._extract_language_from_query(query),
                'framework': self._extract_framework_from_query(query),
                'domain': 'development'
            })
        elif agent_type == "spec-builder":
            strategy.update({
                'domain': 'requirements',
                'topic': 'specification'
            })
        elif agent_type == "quality-gate":
            strategy.update({
                'domain': 'testing',
                'technology': 'quality_assurance'
            })
        elif agent_type == "doc-syncer":
            strategy.update({
                'domain': 'documentation'
            })
        
        return strategy
    
    def _extract_language_from_query(self, query: str) -> Optional[str]:
        """Extract programming language from query"""
        languages = ['python', 'javascript', 'typescript', 'java', 'go', 'rust', 'cpp']
        for language in languages:
            if language.lower() in query.lower():
                return language
        return None
    
    def _extract_framework_from_query(self, query: str) -> Optional[str]:
        """Extract framework from query"""
        frameworks = ['react', 'vue', 'angular', 'django', 'flask', 'fastapi', 'express', 'tensorflow', 'pytorch']
        for framework in frameworks:
            if framework.lower() in query.lower():
                return framework
        return None
    
    async def _mock_context7_search(self, query: str, content_type: str,
                                   limit: int = 10) -> List[Dict[str, Any]]:
        """Mock Context7 search (in real implementation, use actual MCP tools)"""
        
        # Simulate API call delay
        await asyncio.sleep(0.1)
        
        # Mock results based on content type
        if content_type == "code_examples":
            return [
                {
                    'title': f'Example for {query}',
                    'code': f'// Example code for {query}\nfunction example() {{\n  // implementation\n}}',
                    'language': 'javascript',
                    'framework': 'react',
                    'description': 'Example implementation',
                    'url': f'https://example.com/code/{query}'
                }
            ]
        elif content_type == "best_practices":
            return [
                {
                    'title': f'Best Practices for {query}',
                    'practice': 'Always write tests before implementation',
                    'category': 'TDD',
                    'description': 'Follow test-driven development principles',
                    'source': 'industry_experts'
                }
            ]
        elif content_type == "documentation":
            return [
                {
                    'title': f'{query} Documentation',
                    'content': f'Detailed documentation for {query}',
                    'api_reference': {'endpoints': [], 'parameters': []},
                    'examples': []
                }
            ]
        
        return []
    
    def _extract_guidelines(self, practices: List[Dict[str, Any]]) -> List[str]:
        """Extract guidelines from best practices"""
        guidelines = []
        for practice in practices:
            if 'practice' in practice:
                guidelines.append(practice['practice'])
        return guidelines
    
    def _extract_code_examples(self, content: List[Dict[str, Any]]) -> List[Dict[str, Any]]:
        """Extract code examples from content"""
        examples = []
        for item in content:
            if 'code' in item:
                examples.append({
                    'code': item['code'],
                    'language': item.get('language', 'unknown'),
                    'description': item.get('description', '')
                })
        return examples
    
    def _extract_api_reference(self, documentation: List[Dict[str, Any]]) -> Dict[str, Any]:
        """Extract API reference from documentation"""
        api_ref = {}
        for doc in documentation:
            if 'api_reference' in doc:
                api_ref.update(doc['api_reference'])
        return api_ref
    
    def _extract_tutorials(self, documentation: List[Dict[str, Any]]) -> List[Dict[str, Any]]:
        """Extract tutorials from documentation"""
        tutorials = []
        for doc in documentation:
            if 'type' in doc and doc['type'] == 'tutorial':
                tutorials.append(doc)
        return tutorials
```

#### 3. Workflow Templates and Patterns

```python
from abc import ABC, abstractmethod
from typing import Dict, Any, List
import json

class WorkflowTemplate(ABC):
    """Abstract base class for workflow templates"""
    
    def __init__(self, name: str, description: str):
        self.name = name
        self.description = description
    
    @abstractmethod
    def create_workflow(self, engine: WorkflowEngine, config: Dict[str, Any]) -> Workflow:
        """Create a workflow instance from this template"""
        pass

class FeatureDevelopmentTemplate(WorkflowTemplate):
    """Template for feature development workflow"""
    
    def __init__(self):
        super().__init__(
            name="feature_development",
            description="Complete feature development workflow from specification to deployment"
        )
    
    def create_workflow(self, engine: WorkflowEngine, config: Dict[str, Any]) -> Workflow:
        """Create feature development workflow"""
        
        workflow = engine.create_workflow(
            name=f"feature_development_{config.get('feature_name', 'unknown')}",
            description=f"Develop feature: {config.get('feature_name', 'unknown')}"
        )
        
        # Stage 1: Specification
        spec_task = workflow.add_stage(
            stage_name="specification",
            agent_type="spec-builder",
            input_data={
                "feature_description": config.get('feature_description', ''),
                "requirements": config.get('requirements', []),
                "acceptance_criteria": config.get('acceptance_criteria', [])
            }
        )
        
        # Stage 2: Planning
        planning_task = workflow.add_stage(
            stage_name="planning",
            agent_type="plan-agent",
            input_data={
                "spec_id": spec_task.id,
                "complexity": config.get('complexity', 'medium'),
                "timeline": config.get('timeline', '2_weeks')
            },
            depends_on=[spec_task.id]
        )
        
        # Stage 3: Implementation
        impl_task = workflow.add_stage(
            stage_name="implementation",
            agent_type="tdd-implementer",
            input_data={
                "plan_id": planning_task.id,
                "technology_stack": config.get('technology_stack', []),
                "coding_standards": config.get('coding_standards', {})
            },
            depends_on=[planning_task.id]
        )
        
        # Stage 4: Testing
        testing_task = workflow.add_stage(
            stage_name="testing",
            agent_type="quality-gate",
            input_data={
                "implementation_id": impl_task.id,
                "test_types": config.get('test_types', ['unit', 'integration', 'e2e']),
                "coverage_threshold": config.get('coverage_threshold', 80)
            },
            depends_on=[impl_task.id]
        )
        
        # Stage 5: Documentation
        docs_task = workflow.add_stage(
            stage_name="documentation",
            agent_type="doc-syncer",
            input_data={
                "feature_name": config.get('feature_name', ''),
                "technical_docs": config.get('technical_docs', True),
                "user_docs": config.get('user_docs', True)
            },
            depends_on=[testing_task.id]
        )
        
        # Stage 6: Deployment (optional)
        if config.get('auto_deploy', False):
            deploy_task = workflow.add_stage(
                stage_name="deployment",
                agent_type="deployment-agent",
                input_data={
                    "environment": config.get('deploy_environment', 'staging'),
                    "rollback_strategy": config.get('rollback_strategy', True)
                },
                depends_on=[docs_task.id]
            )
        
        return workflow

class BugFixTemplate(WorkflowTemplate):
    """Template for bug fix workflow"""
    
    def __init__(self):
        super().__init__(
            name="bug_fix",
            description="Bug fix workflow from identification to resolution"
        )
    
    def create_workflow(self, engine: WorkflowEngine, config: Dict[str, Any]) -> Workflow:
        """Create bug fix workflow"""
        
        workflow = engine.create_workflow(
            name=f"bug_fix_{config.get('bug_id', 'unknown')}",
            description=f"Fix bug: {config.get('bug_description', 'unknown')}"
        )
        
        # Stage 1: Bug Analysis
        analysis_task = workflow.add_stage(
            stage_name="bug_analysis",
            agent_type="debug-helper",
            input_data={
                "bug_description": config.get('bug_description', ''),
                "error_logs": config.get('error_logs', []),
                "reproduction_steps": config.get('reproduction_steps', [])
            }
        )
        
        # Stage 2: Root Cause Analysis
        rca_task = workflow.add_stage(
            stage_name="root_cause_analysis",
            agent_type="debug-helper",
            input_data={
                "analysis_id": analysis_task.id,
                "investigation_depth": config.get('investigation_depth', 'deep')
            },
            depends_on=[analysis_task.id]
        )
        
        # Stage 3: Fix Implementation
        fix_task = workflow.add_stage(
            stage_name="fix_implementation",
            agent_type="tdd-implementer",
            input_data={
                "root_cause_id": rca_task.id,
                "fix_approach": config.get('fix_approach', 'minimal'),
                "test_required": config.get('test_required', True)
            },
            depends_on=[rca_task.id]
        )
        
        # Stage 4: Verification
        verification_task = workflow.add_stage(
            stage_name="verification",
            agent_type="quality-gate",
            input_data={
                "fix_id": fix_task.id,
                "regression_tests": config.get('regression_tests', True),
                "verification_steps": config.get('verification_steps', [])
            },
            depends_on=[fix_task.id]
        )
        
        return workflow

class CodeReviewTemplate(WorkflowTemplate):
    """Template for code review workflow"""
    
    def __init__(self):
        super().__init__(
            name="code_review",
            description="Automated code review workflow"
        )
    
    def create_workflow(self, engine: WorkflowEngine, config: Dict[str, Any]) -> Workflow:
        """Create code review workflow"""
        
        workflow = engine.create_workflow(
            name=f"code_review_{config.get('pr_id', 'unknown')}",
            description=f"Review PR: {config.get('pr_id', 'unknown')}"
        )
        
        # Stage 1: Code Analysis
        analysis_task = workflow.add_stage(
            stage_name="code_analysis",
            agent_type="code-analyzer",
            input_data={
                "pr_id": config.get('pr_id', ''),
                "files_changed": config.get('files_changed', []),
                "analysis_depth": config.get('analysis_depth', 'standard')
            }
        )
        
        # Stage 2: Security Review
        security_task = workflow.add_stage(
            stage_name="security_review",
            agent_type="security-expert",
            input_data={
                "analysis_id": analysis_task.id,
                "security_level": config.get('security_level', 'standard')
            },
            depends_on=[analysis_task.id]
        )
        
        # Stage 3: Performance Review
        performance_task = workflow.add_stage(
            stage_name="performance_review",
            agent_type="performance-engineer",
            input_data={
                "analysis_id": analysis_task.id,
                "performance_tests": config.get('performance_tests', True)
            },
            depends_on=[analysis_task.id]
        )
        
        # Stage 4: Test Coverage Review
        coverage_task = workflow.add_stage(
            stage_name="coverage_review",
            agent_type="quality-gate",
            input_data={
                "analysis_id": analysis_task.id,
                "coverage_threshold": config.get('coverage_threshold', 80)
            },
            depends_on=[analysis_task.id]
        )
        
        # Stage 5: Approval Decision
        approval_task = workflow.add_stage(
            stage_name="approval_decision",
            agent_type="review-approver",
            input_data={
                "security_review": security_task.id,
                "performance_review": performance_task.id,
                "coverage_review": coverage_task.id,
                "approval_criteria": config.get('approval_criteria', {})
            },
            depends_on=[security_task.id, performance_task.id, coverage_task.id]
        )
        
        return workflow

class WorkflowTemplateManager:
    """Manager for workflow templates"""
    
    def __init__(self):
        self.templates = {
            'feature_development': FeatureDevelopmentTemplate(),
            'bug_fix': BugFixTemplate(),
            'code_review': CodeReviewTemplate()
        }
        self.custom_templates = {}
    
    def register_template(self, template: WorkflowTemplate) -> None:
        """Register a custom workflow template"""
        self.custom_templates[template.name] = template
    
    def get_template(self, template_name: str) -> Optional[WorkflowTemplate]:
        """Get a workflow template by name"""
        return self.templates.get(template_name) or self.custom_templates.get(template_name)
    
    def list_templates(self) -> Dict[str, str]:
        """List all available templates"""
        all_templates = {**self.templates, **self.custom_templates}
        return {name: template.description for name, template in all_templates.items()}
    
    def create_workflow_from_template(self, template_name: str, engine: WorkflowEngine,
                                    config: Dict[str, Any]) -> Optional[Workflow]:
        """Create a workflow from a template"""
        template = self.get_template(template_name)
        if template:
            return template.create_workflow(engine, config)
        return None
    
    def save_template_config(self, template_name: str, config: Dict[str, Any],
                           filename: str = None) -> str:
        """Save template configuration to file"""
        if filename is None:
            filename = f"workflow_config_{template_name}_{int(time.time())}.json"
        
        config_data = {
            'template_name': template_name,
            'config': config,
            'created_at': datetime.now().isoformat()
        }
        
        with open(filename, 'w') as f:
            json.dump(config_data, f, indent=2)
        
        return filename
    
    def load_template_config(self, filename: str) -> Dict[str, Any]:
        """Load template configuration from file"""
        with open(filename, 'r') as f:
            return json.load(f)
```

## Level 3: Advanced Integration

### Enterprise Workflow Management

#### 4. Advanced Workflow Orchestration

```python
import asyncio
from typing import Dict, List, Any, Optional, Callable
from dataclasses import dataclass, field
from enum import Enum
import uuid
from datetime import datetime, timedelta

class WorkflowPriority(Enum):
    LOW = 1
    MEDIUM = 2
    HIGH = 3
    CRITICAL = 4

class WorkflowStatus(Enum):
    PENDING = "pending"
    RUNNING = "running"
    COMPLETED = "completed"
    FAILED = "failed"
    CANCELLED = "cancelled"
    PAUSED = "paused"

@dataclass
class WorkflowMetrics:
    total_workflows: int = 0
    completed_workflows: int = 0
    failed_workflows: int = 0
    average_execution_time: float = 0.0
    success_rate: float = 0.0
    agent_utilization: Dict[str, float] = field(default_factory=dict)

class AdvancedWorkflowEngine:
    """Advanced workflow engine with enterprise features"""
    
    def __init__(self, max_concurrent_workflows: int = 10):
        self.workflow_engine = WorkflowEngine(max_concurrent_tasks=max_concurrent_workflows)
        self.template_manager = WorkflowTemplateManager()
        self.context7 = Context7Integration()
        
        # Workflow management
        self.workflow_queue: asyncio.PriorityQueue = asyncio.PriorityQueue()
        self.active_workflows: Dict[str, Workflow] = {}
        self.completed_workflows: Dict[str, Workflow] = {}
        self.workflow_metrics = WorkflowMetrics()
        
        # Scheduling and monitoring
        self.scheduler_running = False
        self.monitoring_enabled = True
        self.performance_history: List[Dict[str, Any]] = []
        
        # Event handlers
        self.event_handlers: Dict[str, List[Callable]] = {}
        
        self.logger = logging.getLogger(__name__)
    
    async def submit_workflow(self, template_name: str, config: Dict[str, Any],
                            priority: WorkflowPriority = WorkflowPriority.MEDIUM,
                            scheduled_time: Optional[datetime] = None) -> str:
        """Submit a workflow for execution"""
        
        workflow_id = str(uuid.uuid4())
        
        # Create workflow from template
        workflow = self.template_manager.create_workflow_from_template(
            template_name, self.workflow_engine, config
        )
        
        if not workflow:
            raise ValueError(f"Unknown template: {template_name}")
        
        # Schedule workflow
        workflow_item = {
            'workflow_id': workflow_id,
            'workflow': workflow,
            'priority': priority,
            'scheduled_time': scheduled_time or datetime.now(),
            'submitted_at': datetime.now()
        }
        
        # Add to queue (priority queue uses negative priority for max-heap behavior)
        await self.workflow_queue.put((-priority.value, workflow_item))
        
        self.active_workflows[workflow_id] = workflow
        
        # Trigger event
        await self._trigger_event('workflow_submitted', {
            'workflow_id': workflow_id,
            'template_name': template_name,
            'priority': priority
        })
        
        self.logger.info(f"Workflow {workflow_id} submitted using template {template_name}")
        return workflow_id
    
    async def start_scheduler(self) -> None:
        """Start the workflow scheduler"""
        
        if self.scheduler_running:
            return
        
        self.scheduler_running = True
        self.logger.info("Workflow scheduler started")
        
        try:
            while self.scheduler_running:
                try:
                    # Get next workflow (with timeout to allow graceful shutdown)
                    priority_item = await asyncio.wait_for(
                        self.workflow_queue.get(),
                        timeout=1.0
                    )
                    
                    priority, workflow_item = priority_item
                    workflow = workflow_item['workflow']
                    workflow_id = list(self.active_workflows.keys())[
                        list(self.active_workflows.values()).index(workflow)
                    ]
                    
                    # Check if scheduled time has arrived
                    if workflow_item['scheduled_time'] <= datetime.now():
                        # Execute workflow
                        asyncio.create_task(self._execute_workflow(workflow_id, workflow))
                    else:
                        # Re-queue for later
                        await asyncio.sleep(
                            (workflow_item['scheduled_time'] - datetime.now()).total_seconds()
                        )
                        await self.workflow_queue.put((priority, workflow_item))
                
                except asyncio.TimeoutError:
                    # No workflows in queue, continue
                    continue
                    
        except Exception as e:
            self.logger.error(f"Scheduler error: {str(e)}")
        finally:
            self.scheduler_running = False
            self.logger.info("Workflow scheduler stopped")
    
    async def _execute_workflow(self, workflow_id: str, workflow: Workflow) -> None:
        """Execute a workflow with monitoring"""
        
        start_time = datetime.now()
        
        try:
            # Update workflow status
            workflow.status = "running"
            workflow.started_at = start_time
            
            # Trigger event
            await self._trigger_event('workflow_started', {
                'workflow_id': workflow_id,
                'workflow_name': workflow.name
            })
            
            # Execute workflow
            result = await self.workflow_engine.execute_workflow(workflow)
            
            # Update metrics
            execution_time = (datetime.now() - start_time).total_seconds()
            self._update_metrics(workflow, execution_time, success=True)
            
            # Move to completed
            self.completed_workflows[workflow_id] = self.active_workflows.pop(workflow_id)
            
            # Trigger event
            await self._trigger_event('workflow_completed', {
                'workflow_id': workflow_id,
                'execution_time': execution_time,
                'result': result
            })
            
            self.logger.info(f"Workflow {workflow_id} completed successfully")
            
        except Exception as e:
            # Update metrics
            execution_time = (datetime.now() - start_time).total_seconds()
            self._update_metrics(workflow, execution_time, success=False)
            
            # Update workflow status
            workflow.status = "failed"
            workflow.error_message = str(e)
            
            # Trigger event
            await self._trigger_event('workflow_failed', {
                'workflow_id': workflow_id,
                'error': str(e),
                'execution_time': execution_time
            })
            
            self.logger.error(f"Workflow {workflow_id} failed: {str(e)}")
    
    def _update_metrics(self, workflow: Workflow, execution_time: float, success: bool) -> None:
        """Update workflow metrics"""
        
        self.workflow_metrics.total_workflows += 1
        
        if success:
            self.workflow_metrics.completed_workflows += 1
        else:
            self.workflow_metrics.failed_workflows += 1
        
        # Update average execution time
        total_completed = self.workflow_metrics.completed_workflows + self.workflow_metrics.failed_workflows
        self.workflow_metrics.average_execution_time = (
            (self.workflow_metrics.average_execution_time * (total_completed - 1) + execution_time) / 
            total_completed
        )
        
        # Update success rate
        self.workflow_metrics.success_rate = (
            self.workflow_metrics.completed_workflows / total_completed
        )
        
        # Update agent utilization
        self._update_agent_utilization()
        
        # Store performance history
        self.performance_history.append({
            'timestamp': datetime.now().isoformat(),
            'workflow_name': workflow.name,
            'execution_time': execution_time,
            'success': success,
            'tasks_count': len(workflow.tasks)
        })
    
    def _update_agent_utilization(self) -> None:
        """Update agent utilization metrics"""
        
        for agent_id, agent in self.workflow_engine.agents.items():
            total_tasks = len(agent.completed_tasks)
            failed_tasks = agent.performance_metrics.get('failed_tasks', 0)
            
            if total_tasks > 0:
                utilization = (total_tasks - failed_tasks) / total_tasks
            else:
                utilization = 0.0
            
            self.workflow_metrics.agent_utilization[agent_id] = utilization
    
    async def cancel_workflow(self, workflow_id: str) -> bool:
        """Cancel a workflow"""
        
        if workflow_id in self.active_workflows:
            workflow = self.active_workflows[workflow_id]
            
            # Cancel running tasks
            for task in workflow.tasks:
                if task.status == TaskStatus.RUNNING:
                    task.status = TaskStatus.CANCELLED
            
            workflow.status = "cancelled"
            
            # Move to completed
            self.completed_workflows[workflow_id] = self.active_workflows.pop(workflow_id)
            
            # Trigger event
            await self._trigger_event('workflow_cancelled', {
                'workflow_id': workflow_id
            })
            
            self.logger.info(f"Workflow {workflow_id} cancelled")
            return True
        
        return False
    
    async def pause_workflow(self, workflow_id: str) -> bool:
        """Pause a workflow"""
        
        if workflow_id in self.active_workflows:
            workflow = self.active_workflows[workflow_id]
            workflow.status = "paused"
            
            # Trigger event
            await self._trigger_event('workflow_paused', {
                'workflow_id': workflow_id
            })
            
            self.logger.info(f"Workflow {workflow_id} paused")
            return True
        
        return False
    
    async def resume_workflow(self, workflow_id: str) -> bool:
        """Resume a paused workflow"""
        
        if workflow_id in self.active_workflows:
            workflow = self.active_workflows[workflow_id]
            
            if workflow.status == "paused":
                workflow.status = "running"
                
                # Resume execution
                asyncio.create_task(self._execute_workflow(workflow_id, workflow))
                
                # Trigger event
                await self._trigger_event('workflow_resumed', {
                    'workflow_id': workflow_id
                })
                
                self.logger.info(f"Workflow {workflow_id} resumed")
                return True
        
        return False
    
    def add_event_handler(self, event_type: str, handler: Callable) -> None:
        """Add an event handler"""
        
        if event_type not in self.event_handlers:
            self.event_handlers[event_type] = []
        
        self.event_handlers[event_type].append(handler)
    
    async def _trigger_event(self, event_type: str, data: Dict[str, Any]) -> None:
        """Trigger an event"""
        
        if event_type in self.event_handlers:
            for handler in self.event_handlers[event_type]:
                try:
                    if asyncio.iscoroutinefunction(handler):
                        await handler(data)
                    else:
                        handler(data)
                except Exception as e:
                    self.logger.error(f"Event handler error: {str(e)}")
    
    def get_workflow_status(self, workflow_id: str) -> Optional[Dict[str, Any]]:
        """Get workflow status"""
        
        workflow = self.active_workflows.get(workflow_id) or self.completed_workflows.get(workflow_id)
        
        if workflow:
            return {
                'workflow_id': workflow_id,
                'name': workflow.name,
                'status': workflow.status,
                'created_at': workflow.created_at.isoformat(),
                'started_at': workflow.started_at.isoformat() if workflow.started_at else None,
                'completed_at': workflow.completed_at.isoformat() if workflow.completed_at else None,
                'tasks_count': len(workflow.tasks),
                'completed_tasks': len([t for t in workflow.tasks if t.status == TaskStatus.COMPLETED]),
                'failed_tasks': len([t for t in workflow.tasks if t.status == TaskStatus.FAILED]),
                'error_message': workflow.error_message
            }
        
        return None
    
    def get_metrics(self) -> Dict[str, Any]:
        """Get workflow engine metrics"""
        
        return {
            'total_workflows': self.workflow_metrics.total_workflows,
            'active_workflows': len(self.active_workflows),
            'completed_workflows': self.workflow_metrics.completed_workflows,
            'failed_workflows': self.workflow_metrics.failed_workflows,
            'success_rate': self.workflow_metrics.success_rate,
            'average_execution_time': self.workflow_metrics.average_execution_time,
            'agent_utilization': self.workflow_metrics.agent_utilization,
            'queue_size': self.workflow_queue.qsize(),
            'scheduler_running': self.scheduler_running
        }
    
    async def generate_performance_report(self, timeframe_hours: int = 24) -> Dict[str, Any]:
        """Generate performance report for specified timeframe"""
        
        cutoff_time = datetime.now() - timedelta(hours=timeframe_hours)
        
        # Filter performance history
        recent_performance = [
            entry for entry in self.performance_history
            if datetime.fromisoformat(entry['timestamp']) >= cutoff_time
        ]
        
        if not recent_performance:
            return {'error': 'No performance data available for the specified timeframe'}
        
        # Calculate statistics
        execution_times = [entry['execution_time'] for entry in recent_performance]
        success_count = sum(1 for entry in recent_performance if entry['success'])
        
        report = {
            'timeframe_hours': timeframe_hours,
            'total_workflows': len(recent_performance),
            'successful_workflows': success_count,
            'failed_workflows': len(recent_performance) - success_count,
            'success_rate': success_count / len(recent_performance),
            'avg_execution_time': sum(execution_times) / len(execution_times),
            'min_execution_time': min(execution_times),
            'max_execution_time': max(execution_times),
            'performance_trend': self._calculate_performance_trend(recent_performance),
            'agent_performance': self._get_agent_performance_report(),
            'recommendations': self._generate_performance_recommendations(recent_performance)
        }
        
        return report
    
    def _calculate_performance_trend(self, performance_data: List[Dict[str, Any]]) -> str:
        """Calculate performance trend"""
        
        if len(performance_data) < 2:
            return "insufficient_data"
        
        # Compare first half with second half
        mid_point = len(performance_data) // 2
        first_half = performance_data[:mid_point]
        second_half = performance_data[mid_point:]
        
        first_half_success_rate = sum(1 for entry in first_half if entry['success']) / len(first_half)
        second_half_success_rate = sum(1 for entry in second_half if entry['success']) / len(second_half)
        
        if second_half_success_rate > first_half_success_rate + 0.1:
            return "improving"
        elif second_half_success_rate < first_half_success_rate - 0.1:
            return "declining"
        else:
            return "stable"
    
    def _get_agent_performance_report(self) -> Dict[str, Any]:
        """Get agent performance report"""
        
        agent_report = {}
        
        for agent_id, agent in self.workflow_engine.agents.items():
            agent_report[agent_id] = {
                'name': agent.name,
                'type': agent.agent_type,
                'completed_tasks': len(agent.completed_tasks),
                'success_rate': agent.performance_metrics.get('success_rate', 0.0),
                'avg_execution_time': agent.performance_metrics.get('avg_execution_time', 0.0),
                'status': agent.status.value
            }
        
        return agent_report
    
    def _generate_performance_recommendations(self, performance_data: List[Dict[str, Any]]) -> List[str]:
        """Generate performance recommendations"""
        
        recommendations = []
        
        # Analyze success rate
        success_rate = sum(1 for entry in performance_data if entry['success']) / len(performance_data)
        if success_rate < 0.9:
            recommendations.append("Consider reviewing failure patterns and improving error handling")
        
        # Analyze execution time
        execution_times = [entry['execution_time'] for entry in performance_data]
        avg_time = sum(execution_times) / len(execution_times)
        
        if avg_time > 300:  # 5 minutes
            recommendations.append("Consider optimizing workflows for better performance")
        
        # Analyze agent utilization
        for agent_id, utilization in self.workflow_metrics.agent_utilization.items():
            if utilization < 0.5:
                recommendations.append(f"Agent {agent_id} has low utilization, consider load balancing")
        
        return recommendations
    
    async def shutdown(self) -> None:
        """Gracefully shutdown the workflow engine"""
        
        self.logger.info("Shutting down workflow engine...")
        
        # Stop scheduler
        self.scheduler_running = False
        
        # Cancel active workflows
        for workflow_id in list(self.active_workflows.keys()):
            await self.cancel_workflow(workflow_id)
        
        self.logger.info("Workflow engine shutdown complete")

# Example event handlers
async def workflow_started_handler(data: Dict[str, Any]) -> None:
    """Handle workflow started event"""
    print(f"Workflow started: {data['workflow_id']}")

async def workflow_completed_handler(data: Dict[str, Any]) -> None:
    """Handle workflow completed event"""
    print(f"Workflow completed: {data['workflow_id']} in {data['execution_time']:.2f}s")

# Usage example
async def main():
    # Create advanced workflow engine
    engine = AdvancedWorkflowEngine(max_concurrent_workflows=5)
    
    # Register event handlers
    engine.add_event_handler('workflow_started', workflow_started_handler)
    engine.add_event_handler('workflow_completed', workflow_completed_handler)
    
    # Start scheduler
    scheduler_task = asyncio.create_task(engine.start_scheduler())
    
    # Submit workflows
    workflow_id1 = await engine.submit_workflow(
        'feature_development',
        {
            'feature_name': 'user_authentication',
            'feature_description': 'Add JWT-based authentication',
            'technology_stack': ['react', 'node.js', 'postgresql']
        },
        priority=WorkflowPriority.HIGH
    )
    
    workflow_id2 = await engine.submit_workflow(
        'bug_fix',
        {
            'bug_id': 'BUG-1234',
            'bug_description': 'Login page not redirecting properly',
            'reproduction_steps': ['Go to login page', 'Enter credentials', 'Click login']
        },
        priority=WorkflowPriority.CRITICAL
    )
    
    # Monitor workflows
    await asyncio.sleep(10)
    
    # Get metrics
    metrics = engine.get_metrics()
    print(f"Engine metrics: {metrics}")
    
    # Generate performance report
    report = await engine.generate_performance_report()
    print(f"Performance report: {report}")
    
    # Shutdown
    await engine.shutdown()

if __name__ == "__main__":
    asyncio.run(main())
```

## Related Skills

- **moai-cc-skill-factory**: Workflow template creation
- **moai-alfred-agent-guide**: Agent coordination patterns
- **moai-domain-testing**: Quality gate implementation
- **moai-document-processing**: Documentation automation

## Quick Start Checklist

- [ ] Set up workflow engine with appropriate agents
- [ ] Configure Context7 MCP integration
- [ ] Create workflow templates for common patterns
- [ ] Implement event handlers for monitoring
- [ ] Set up performance metrics and monitoring
- [ ] Configure workflow scheduling and priorities
- [ ] Test workflow execution and error handling
- [ ] Implement graceful shutdown procedures

## Workflow Best Practices

1. **Template Design**: Create reusable workflow templates for common patterns
2. **Agent Specialization**: Design agents with specific capabilities and clear responsibilities
3. **Error Handling**: Implement comprehensive error handling and retry mechanisms
4. **Monitoring**: Track workflow performance and agent utilization
5. **Event-Driven Architecture**: Use events for loose coupling and extensibility
6. **Context Integration**: Leverage Context7 for intelligent decision making
7. **Resource Management**: Balance concurrent workflows and agent utilization
8. **Documentation**: Maintain clear documentation for workflow templates and configurations

---

**Alfred Workflow Orchestration** - Build intelligent, scalable workflow systems with multi-agent coordination, Context7 integration, and comprehensive monitoring capabilities.
