# Claude Code Commands - Practical Examples

## Example 1: Project Initialization Command

```python
# commands/project_init.py
from moai_cc_commands import Command, CommandRegistry

class ProjectInitCommand(Command):
    """Project initialization command"""

    name = "project-init"
    description = "Initialize and configure project"

    def setup_parameters(self):
        self.add_parameter("project_name", "Project name")
        self.add_parameter("--template", "Select project template")
        self.add_parameter("--skip-git", "Skip Git initialization")

    async def execute(self, params):
        # Create project directory
        await self.create_directory(params.project_name)

        # Apply template
        if params.template:
            await self.apply_template(params.template)

        # Initialize Git repository
        if not params.skip_git:
            await self.init_git_repo()

        return {
            "status": "success",
            "project": params.project_name,
            "message": "Project initialization complete"
        }

# Register
registry = CommandRegistry()
registry.register(ProjectInitCommand())
```

## Example 2: Workflow Orchestration

```python
# workflows/deployment.py
from moai_cc_commands import Workflow, WorkflowStep

class DeploymentWorkflow(Workflow):
    """Deployment workflow"""

    name = "deploy"
    description = "Application deployment"

    def setup_steps(self):
        # Step 1: Build
        self.add_step(WorkflowStep(
            name="build",
            command="build",
            description="Build application"
        ))

        # Step 2: Test
        self.add_step(WorkflowStep(
            name="test",
            command="test",
            description="Run tests",
            depends_on=["build"]
        ))

        # Step 3: Deploy
        self.add_step(WorkflowStep(
            name="deploy",
            command="deploy",
            description="Deploy to production",
            depends_on=["test"],
            on_failure="rollback"
        ))

    async def on_step_complete(self, step_name, result):
        """Handle step completion"""
        if step_name == "test" and result.status == "failed":
            await self.notify_team(f"Test failed: {result.error}")
```

## Example 3: Parameter Validation

```python
# validators/project_validator.py
from moai_cc_commands import Validator

class ProjectNameValidator(Validator):
    """Project name validator"""

    def validate(self, value):
        # Check for empty value
        if not value:
            raise ValueError("Project name is required")

        # Check length
        if len(value) < 3:
            raise ValueError("Project name must be at least 3 characters")

        # Check characters
        if not value.isalnum():
            raise ValueError("Project name can only contain alphanumeric characters")

        return value
```

## Example 4: Command Chaining

```python
# commands/chained_commands.py
from moai_cc_commands import CommandChain

async def run_deployment_pipeline():
    """Execute deployment pipeline"""

    chain = CommandChain()

    # Build chain
    chain.add_command("/moai:2-run", {"spec_id": "SPEC-001"})
    chain.add_command("/moai:3-sync", {"spec_id": "SPEC-001"})
    chain.add_command("/deploy", {"environment": "production"})

    # Execute sequentially
    result = await chain.execute()

    return result
```

## Example 5: CLI Help System

```python
# cli/help_system.py
from moai_cc_commands import Command

class HelpCommand(Command):
    """Help command"""

    name = "help"
    description = "Command help information"

    def format_help(self, command):
        """Format command help"""
        return f"""
        Command: {command.name}
        Description: {command.description}

        Parameters:
        {self._format_parameters(command)}

        Usage Example:
        {command.usage_example}
        """

    async def execute(self, params):
        command_name = params.get("command")
        command = self.registry.get(command_name)

        if not command:
            return {"error": f"Command not found: {command_name}"}

        return {"help": self.format_help(command)}
```

## Example 6: Error Handling and Retry

```python
# error_handling/retry_logic.py
from moai_cc_commands import CommandExecutor

class ResilientCommandExecutor(CommandExecutor):
    """Resilient command executor"""

    async def execute_with_retry(self, command, params, max_retries=3):
        """Execute with retry logic"""

        last_error = None

        for attempt in range(max_retries):
            try:
                result = await command.execute(params)
                return result

            except Exception as e:
                last_error = e

                # Wait before retry
                if attempt < max_retries - 1:
                    wait_time = 2 ** attempt  # Exponential backoff
                    await asyncio.sleep(wait_time)

        # All retries failed
        raise last_error
```

## Example 7: Progress Tracking

```python
# progress/progress_tracker.py
from moai_cc_commands import ProgressTracker

class WorkflowProgressTracker(ProgressTracker):
    """Workflow progress tracker"""

    async def track_workflow(self, workflow):
        """Track workflow progress"""

        total_steps = len(workflow.steps)

        for idx, step in enumerate(workflow.steps):
            # Update progress
            progress = (idx / total_steps) * 100

            self.update(
                current=idx + 1,
                total=total_steps,
                percentage=progress,
                message=f"Executing: {step.name}"
            )

            # Execute step
            result = await step.execute()

            self.log_result(step.name, result)
```

---

**Last Updated**: 2025-11-22
**Status**: Production Ready
