# Alfred Workflow Reference

## Architecture Overview

### Core Components

#### WorkflowEngine
The central orchestrator that manages workflow execution, agent coordination, and resource management.

**Key Responsibilities:**
- Agent lifecycle management
- Task scheduling and execution
- Error handling and recovery
- Performance monitoring and optimization

#### Agent System
Multi-agent architecture with specialized agents for different domains.

**Agent Types:**
- **spec-builder**: Requirements analysis and SPEC creation
- **plan-agent**: Implementation planning and task decomposition
- **tdd-implementer**: Test-driven development and code implementation
- **quality-gate**: Testing validation and quality assurance
- **doc-syncer**: Documentation generation and synchronization
- **debug-helper**: Error analysis and troubleshooting

#### Workflow Templates
Pre-built workflow patterns for common scenarios.

**Built-in Templates:**
- Feature Development
- Bug Fix
- Code Review
- Security Scan
- Deployment Pipeline

## API Reference

### WorkflowEngine Class

```python
class WorkflowEngine:
    """Central workflow orchestration engine"""
    
    def __init__(self, config: Dict[str, Any] = None):
        """Initialize workflow engine with configuration"""
        
    async def submit_workflow(self, template_name: str, config: Dict[str, Any],
                            priority: WorkflowPriority = WorkflowPriority.MEDIUM,
                            scheduled_time: Optional[datetime] = None) -> str:
        """Submit a workflow for execution"""
        
    async def execute_workflow(self, workflow: Workflow) -> Dict[str, Any]:
        """Execute a single workflow"""
        
    def register_agent(self, agent: Agent) -> None:
        """Register a new agent with the engine"""
        
    def register_template(self, template: WorkflowTemplate) -> None:
        """Register a new workflow template"""
        
    def get_workflow_status(self, workflow_id: str) -> Dict[str, Any]:
        """Get current status of a workflow"""
        
    async def wait_for_completion(self, workflow_id: str) -> Dict[str, Any]:
        """Wait for workflow completion"""
```

### Agent Classes

```python
@dataclass
class Agent:
    """Represents a workflow agent"""
    id: str
    name: str
    agent_type: str
    domain: str
    capabilities: List[str]
    status: AgentStatus = AgentStatus.IDLE
    max_concurrent_tasks: int = 1
    metadata: Dict[str, Any] = field(default_factory=dict)

class WorkflowAgent:
    """Base class for workflow agents"""
    
    async def execute_task(self, task: Task) -> TaskResult:
        """Execute a single task"""
        
    async def handle_error(self, task: Task, error: Exception) -> bool:
        """Handle task execution errors"""
```

### Task Classes

```python
@dataclass
class Task:
    """Represents a workflow task"""
    id: str
    name: str
    description: str
    agent_type: str
    input_data: Dict[str, Any]
    dependencies: List[str] = field(default_factory=list)
    status: TaskStatus = TaskStatus.PENDING
    retry_count: int = 0
    max_retries: int = 3
    timeout: int = 300  # 5 minutes default

@dataclass
class TaskResult:
    """Result of task execution"""
    task_id: str
    status: TaskStatus
    output_data: Dict[str, Any]
    execution_time: float
    error_message: Optional[str] = None
```

### Workflow Classes

```python
@dataclass
class Workflow:
    """Represents a complete workflow"""
    name: str
    description: str
    engine: WorkflowEngine
    tasks: List[Task] = field(default_factory=list)
    status: str = "created"
    created_at: datetime = field(default_factory=datetime.now)
    
    def add_task(self, task: Task) -> None:
        """Add a task to the workflow"""
        
    def add_stage(self, stage_name: str, agent_type: str,
                  input_data: Dict[str, Any] = None,
                  depends_on: List[str] = None) -> Task:
        """Add a stage (task) to the workflow"""
        
    async def execute(self) -> Dict[str, Any]:
        """Execute the workflow"""
```

## Configuration

### WorkflowEngine Configuration

```python
config = {
    # Agent configuration
    "max_agents": 10,
    "agent_timeout": 300,  # 5 minutes
    
    # Task execution
    "max_concurrent_tasks": 20,
    "task_retry_attempts": 3,
    "task_timeout": 600,  # 10 minutes
    
    # Workflow management
    "workflow_timeout": 3600,  # 1 hour
    "max_workflows": 100,
    
    # Performance monitoring
    "enable_metrics": True,
    "metrics_retention_days": 30,
    
    # Error handling
    "error_recovery_strategy": "retry_with_backoff",
    "alert_on_failure": True,
    
    # Integration settings
    "context7_enabled": True,
    "github_integration": True,
    "notification_channels": ["slack", "email"]
}

workflow_engine = WorkflowEngine(config)
```

### Agent Configuration

```python
# Configure specific agents
agent_configs = {
    "spec-builder": {
        "max_concurrent_tasks": 2,
        "timeout": 600,
        "capabilities": ["requirements_analysis", "spec_creation", "acceptance_criteria"]
    },
    "tdd-implementer": {
        "max_concurrent_tasks": 4,
        "timeout": 1200,
        "capabilities": ["test_driven_development", "code_implementation", "refactoring"]
    },
    "quality-gate": {
        "max_concurrent_tasks": 3,
        "timeout": 900,
        "capabilities": ["testing", "quality_assurance", "code_review"]
    }
}
```

### Template Configuration

```python
# Custom template configuration
template_config = {
    "feature_development": {
        "stages": ["specification", "planning", "implementation", "testing", "documentation"],
        "required_approvals": 1,
        "timeout": 3600
    },
    "bug_fix": {
        "stages": ["bug_analysis", "fix_implementation", "testing", "verification"],
        "priority": "high",
        "timeout": 1800
    }
}
```

## Context7 MCP Integration

### Setup Configuration

```python
# .mcp.json configuration
{
    "mcpServers": {
        "context7": {
            "command": "npx",
            "args": ["-y", "@upstash/context7-mcp@latest"],
            "env": {
                "CONTEXT7_API_KEY": "your-api-key"
            }
        },
        "github": {
            "command": "npx",
            "args": ["-y", "@anthropic-ai/mcp-server-github"],
            "env": {
                "GITHUB_TOKEN": "your-github-token"
            }
        }
    }
}
```

### Integration Classes

```python
class Context7Integration:
    """Integration with Context7 MCP server"""
    
    def __init__(self, mcp_servers: List[str] = None):
        self.mcp_servers = mcp_servers or []
        self.cache = {}
        self.cache_ttl = 3600  # 1 hour
        
    async def search_code_examples(self, query: str, language: str = None,
                                 framework: str = None, limit: int = 10) -> List[Dict[str, Any]]:
        """Search for code examples using Context7"""
        
    async def get_best_practices(self, topic: str, language: str = None) -> Dict[str, Any]:
        """Get best practices for a specific topic"""
        
    async def lookup_documentation(self, library: str, topic: str = None) -> Dict[str, Any]:
        """Lookup official documentation"""
```

## Workflow Templates Reference

### FeatureDevelopmentTemplate

```python
class FeatureDevelopmentTemplate(WorkflowTemplate):
    """Template for feature development workflow"""
    
    def create_workflow(self, engine: WorkflowEngine, config: Dict[str, Any]) -> Workflow:
        """
        Configuration Parameters:
        - feature_name: str - Name of the feature
        - feature_description: str - Detailed description
        - requirements: List[str] - Feature requirements
        - acceptance_criteria: List[str] - Acceptance criteria
        - technology_stack: List[str] - Technologies to use
        - test_types: List[str] - Types of tests to run
        - coverage_threshold: int - Minimum code coverage percentage
        - auto_deploy: bool - Whether to auto-deploy after completion
        """
```

### BugFixTemplate

```python
class BugFixTemplate(WorkflowTemplate):
    """Template for bug fix workflow"""
    
    def create_workflow(self, engine: WorkflowEngine, config: Dict[str, Any]) -> Workflow:
        """
        Configuration Parameters:
        - bug_id: str - Unique bug identifier
        - bug_description: str - Detailed bug description
        - error_logs: List[str] - Related error logs
        - reproduction_steps: List[str] - Steps to reproduce bug
        - priority: str - Bug priority (low, medium, high, critical)
        - environment: str - Environment where bug occurs
        """
```

### SecurityScanTemplate

```python
class SecurityScanTemplate(WorkflowTemplate):
    """Template for security scanning workflow"""
    
    def create_workflow(self, engine: WorkflowEngine, config: Dict[str, Any]) -> Workflow:
        """
        Configuration Parameters:
        - scan_type: str - Type of security scan (quick, standard, comprehensive)
        - target: str - Target to scan (repository, application, infrastructure)
        - tools: List[str] - Security tools to use
        - severity_threshold: str - Minimum severity to report
        - generate_report: bool - Whether to generate detailed report
        """
```

## Error Handling

### Exception Classes

```python
class WorkflowError(Exception):
    """Base exception for workflow errors"""
    pass

class AgentError(WorkflowError):
    """Exception for agent-related errors"""
    pass

class TaskError(WorkflowError):
    """Exception for task-related errors"""
    pass

class TemplateError(WorkflowError):
    """Exception for template-related errors"""
    pass
```

### Error Recovery Strategies

```python
# Retry with exponential backoff
RETRY_WITH_BACKOFF = {
    "max_retries": 3,
    "backoff_factor": 2,
    "base_delay": 1.0
}

# Circuit breaker pattern
CIRCUIT_BREAKER = {
    "failure_threshold": 5,
    "recovery_timeout": 60,
    "expected_exception": WorkflowError
}

# Dead letter queue
DEAD_LETTER_QUEUE = {
    "enabled": True,
    "max_retries": 3,
    "retry_delay": 300
}
```

## Monitoring and Metrics

### Metrics Collection

```python
class WorkflowMetrics:
    """Metrics collection for workflow engine"""
    
    def get_workflow_metrics(self) -> Dict[str, Any]:
        """Get workflow performance metrics"""
        return {
            "total_workflows": int,
            "successful_workflows": int,
            "failed_workflows": int,
            "average_execution_time": float,
            "success_rate": float,
            "throughput": float  # workflows per hour
        }
        
    def get_agent_metrics(self) -> Dict[str, Any]:
        """Get agent performance metrics"""
        return {
            "agent_utilization": Dict[str, float],
            "task_completion_rate": Dict[str, float],
            "average_task_time": Dict[str, float],
            "error_rate": Dict[str, float]
        }
```

### Performance Monitoring

```python
# Performance thresholds
performance_thresholds = {
    "workflow_execution_time": 3600,  # 1 hour
    "task_execution_time": 600,       # 10 minutes
    "agent_response_time": 30,        # 30 seconds
    "memory_usage": 0.8,              # 80% of available memory
    "cpu_usage": 0.7                  # 70% CPU utilization
}
```

## Security and Compliance

### Security Configuration

```python
security_config = {
    "authentication": {
        "enabled": True,
        "method": "jwt",
        "token_expiry": 3600
    },
    "authorization": {
        "enabled": True,
        "rbac": True,
        "default_permissions": ["read", "execute"]
    },
    "encryption": {
        "data_at_rest": True,
        "data_in_transit": True,
        "encryption_algorithm": "AES-256"
    },
    "audit": {
        "enabled": True,
        "log_level": "INFO",
        "retention_days": 90
    }
}
```

### Compliance Standards

```python
compliance_standards = {
    "owasp": {
        "enabled": True,
        "checks": ["sql_injection", "xss", "csrf", "authentication"],
        "severity_threshold": "medium"
    },
    "gdpr": {
        "enabled": True,
        "data_retention": 365,
        "right_to_deletion": True
    },
    "sox": {
        "enabled": False,
        "audit_trail": True,
        "segregation_of_duties": True
    }
}
```

## External Integrations

### GitHub Integration

```python
class GitHubIntegration:
    """Integration with GitHub API"""
    
    def __init__(self, token: str):
        self.token = token
        
    async def create_pull_request(self, repo: str, title: str, head: str, 
                                base: str = "main", body: str = "") -> Dict[str, Any]:
        """Create a pull request"""
        
    async def add_reviewers(self, repo: str, pull_number: int, 
                          reviewers: List[str]) -> None:
        """Add reviewers to a pull request"""
        
    async def create_issue(self, repo: str, title: str, body: str, 
                          labels: List[str] = None) -> Dict[str, Any]:
        """Create an issue"""
```

### Slack Integration

```python
class SlackIntegration:
    """Integration with Slack API"""
    
    def __init__(self, token: str):
        self.token = token
        
    async def send_message(self, channel: str, message: str, 
                         blocks: List[Dict] = None) -> None:
        """Send message to Slack channel"""
        
    async def upload_file(self, channel: str, file_path: str, 
                         title: str = None) -> None:
        """Upload file to Slack channel"""
```

## Best Practices

### Workflow Design

1. **Keep workflows simple**: Limit to 5-7 stages maximum
2. **Clear dependencies**: Explicitly define task dependencies
3. **Error handling**: Implement proper error handling at each stage
4. **Timeouts**: Set appropriate timeouts for tasks and workflows
5. **Idempotency**: Design workflows to be idempotent when possible

### Agent Development

1. **Single responsibility**: Each agent should have a clear, focused purpose
2. **Async operations**: Use async/await for I/O operations
3. **Error handling**: Implement comprehensive error handling
4. **Logging**: Include detailed logging for debugging
5. **Resource management**: Properly manage resources and connections

### Performance Optimization

1. **Parallel execution**: Run independent tasks in parallel
2. **Resource pooling**: Use connection pools for external services
3. **Caching**: Cache frequently accessed data
4. **Batching**: Batch operations when possible
5. **Monitoring**: Monitor performance metrics continuously

### Security Considerations

1. **Input validation**: Validate all inputs thoroughly
2. **Secrets management**: Use proper secrets management
3. **Audit logging**: Log all workflow executions
4. **Access control**: Implement proper access controls
5. **Regular updates**: Keep dependencies and security tools updated
