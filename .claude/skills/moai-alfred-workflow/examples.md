# Alfred Workflow Examples

## Quick Start Examples

### Basic Multi-Agent Workflow

```python
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

### Feature Development Template

```python
from alfred_workflow.templates import FeatureDevelopmentTemplate

# Create template
template = FeatureDevelopmentTemplate()

# Configure workflow
config = {
    "feature_name": "user_authentication",
    "feature_description": "Implement secure user authentication",
    "requirements": [
        "User registration with email verification",
        "Secure password hashing",
        "JWT token generation",
        "Session management"
    ],
    "acceptance_criteria": [
        "Users can register with email and password",
        "Email verification required for activation",
        "Passwords are securely hashed using bcrypt",
        "JWT tokens expire after 24 hours",
        "Session persists across browser restarts"
    ],
    "technology_stack": ["Python", "FastAPI", "JWT", "bcrypt"],
    "test_types": ["unit", "integration", "security", "e2e"],
    "coverage_threshold": 85
}

# Submit workflow
workflow_id = await workflow_engine.submit_workflow("feature_development", config)
```

### Bug Fix Template

```python
from alfred_workflow.templates import BugFixTemplate

# Create bug fix workflow
template = BugFixTemplate()

config = {
    "bug_id": "BUG-2024-001",
    "bug_description": "User authentication fails on password reset",
    "error_logs": [
        "ERROR: Authentication failed: invalid token format",
        "WARNING: Password reset token expired"
    ],
    "reproduction_steps": [
        "1. Request password reset",
        "2. Click reset link in email",
        "3. Enter new password",
        "4. Submit form"
    ],
    "priority": "high",
    "environment": "production"
}

workflow_id = await workflow_engine.submit_workflow("bug_fix", config)
```

## Advanced Examples

### Custom Workflow Template

```python
from alfred_workflow import WorkflowTemplate, Workflow, WorkflowEngine
from dataclasses import dataclass
from typing import Dict, Any, Optional

class CodeReviewTemplate(WorkflowTemplate):
    """Custom template for code review workflow"""
    
    def __init__(self):
        super().__init__(
            name="code_review",
            description="Automated code review workflow"
        )
    
    def create_workflow(self, engine: WorkflowEngine, config: Dict[str, Any]) -> Workflow:
        workflow = engine.create_workflow(
            name=f"code_review_{config.get('pr_id', 'unknown')}",
            description=f"Review PR: {config.get('pr_title', 'unknown')}"
        )
        
        # Static analysis stage
        static_analysis_task = workflow.add_stage(
            stage_name="static_analysis",
            agent_type="quality-gate",
            input_data={
                "pr_url": config.get('pr_url'),
                "analysis_tools": ["pylint", "mypy", "security-scan"],
                "thresholds": {
                    "complexity": 10,
                    "coverage": 80,
                    "security": 0
                }
            }
        )
        
        # Performance review stage
        performance_task = workflow.add_stage(
            stage_name="performance_review",
            agent_type="performance-engineer",
            input_data={
                "pr_url": config.get('pr_url'),
                "benchmarks": ["load_test", "memory_usage", "response_time"],
                "regression_threshold": 5  # 5% performance degradation allowed
            },
            depends_on=[static_analysis_task.id]
        )
        
        # Security review stage
        security_task = workflow.add_stage(
            stage_name="security_review",
            agent_type="security-expert",
            input_data={
                "pr_url": config.get('pr_url'),
                "security_checks": ["owasp", "dependency_scan", "secret_scan"],
                "critical_issues_only": config.get('critical_only', False)
            },
            depends_on=[performance_task.id]
        )
        
        return workflow

# Register custom template
workflow_engine.register_template(CodeReviewTemplate())

# Use custom template
config = {
    "pr_id": "1234",
    "pr_title": "Add user authentication API",
    "pr_url": "https://github.com/repo/pull/1234",
    "critical_only": True
}

workflow_id = await workflow_engine.submit_workflow("code_review", config)
```

### Context7 Integration Examples

```python
from alfred_workflow.integrations import Context7Integration

# Initialize Context7 integration
context7 = Context7Integration(
    mcp_servers=["context7", "github", "filesystem"]
)

# Search for code examples
code_examples = await context7.search_code_examples(
    query="react authentication with JWT",
    language="typescript",
    framework="react",
    limit=10
)

# Get best practices
best_practices = await context7.get_best_practices(
    topic="python async programming",
    language="python"
)

# Documentation lookup
docs = await context7.lookup_documentation(
    library="fastapi",
    topic="authentication"
)
```

### Monitoring and Metrics

```python
# Get workflow engine metrics
metrics = workflow_engine.get_metrics()
print(f"Total workflows: {metrics['total_workflows']}")
print(f"Active workflows: {metrics['active_workflows']}")
print(f"Queue size: {metrics['queue_size']}")

# Get workflow performance data
performance_data = workflow_engine.get_performance_analysis()
print(f"Average execution time: {performance_data['avg_execution_time']}")
print(f"Success rate: {performance_data['success_rate']}")

# Monitor specific workflow
workflow_status = workflow_engine.get_workflow_status(workflow_id)
print(f"Workflow status: {workflow_status['status']}")
print(f"Current stage: {workflow_status['current_stage']}")
print(f"Progress: {workflow_status['progress']}%")
```

### Error Handling and Recovery

```python
from alfred_workflow.exceptions import WorkflowError, AgentError
import asyncio

async def robust_workflow_execution():
    try:
        # Submit workflow with retry logic
        workflow_id = await workflow_engine.submit_workflow(
            "feature_development",
            config,
            priority=WorkflowPriority.HIGH
        )
        
        # Wait for completion with timeout
        result = await asyncio.wait_for(
            workflow_engine.wait_for_completion(workflow_id),
            timeout=3600  # 1 hour timeout
        )
        
        return result
        
    except WorkflowError as e:
        logger.error(f"Workflow failed: {e}")
        
        # Get error details
        error_info = workflow_engine.get_workflow_error(workflow_id)
        
        # Decide on recovery strategy
        if error_info['retryable']:
            logger.info("Retrying workflow...")
            return await robust_workflow_execution()
        else:
            # Manual intervention required
            await alert_developer(error_info)
            raise
            
    except asyncio.TimeoutError:
        logger.error("Workflow timed out")
        await workflow_engine.cancel_workflow(workflow_id)
        raise

async def alert_developer(error_info):
    """Send alert to development team"""
    from alfred_workflow.notifications import NotificationManager
    
    notifications = NotificationManager()
    
    await notifications.send_slack_alert(
        channel="#dev-alerts",
        message=f"🚨 Workflow failed: {error_info['workflow_id']}",
        details=error_info
    )
    
    await notifications.send_email(
        to=["dev-team@company.com"],
        subject="Workflow Failure Alert",
        body=f"Workflow {error_info['workflow_id']} requires attention"
    )
```

### Scheduled Workflows

```python
from datetime import datetime, timedelta
import asyncio

async def setup_scheduled_workflows():
    """Setup recurring scheduled workflows"""
    
    # Daily security scan
    tomorrow = datetime.now() + timedelta(days=1)
    await workflow_engine.submit_workflow(
        "security_scan",
        {"scan_type": "full", "environment": "production"},
        scheduled_time=tomorrow.replace(hour=2, minute=0)  # 2 AM
    )
    
    # Weekly performance report
    next_week = datetime.now() + timedelta(weeks=1)
    await workflow_engine.submit_workflow(
        "performance_report",
        {"report_type": "weekly", "include_trends": True},
        scheduled_time=next_week.replace(hour=8, minute=0)  # 8 AM Monday
    )
    
    # Monthly deployment checklist
    next_month = datetime.now() + timedelta(days=30)
    await workflow_engine.submit_workflow(
        "deployment_checklist",
        {"environment": "production", "compliance": True},
        scheduled_time=next_month.replace(hour=10, minute=0)
    )

# Run scheduled workflows
asyncio.create_task(setup_scheduled_workflows())
asyncio.create_task(workflow_engine.start_scheduler())
```
