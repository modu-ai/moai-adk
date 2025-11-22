# Advanced Patterns - Claude Code Commands

## Pattern 1: Command Chaining with Dependency Resolution

Chain multiple commands with automatic dependency detection and execution ordering.

```python
from dataclasses import dataclass
from typing import Dict, List, Set, Callable
from enum import Enum

class CommandStatus(Enum):
    PENDING = "pending"
    RUNNING = "running"
    COMPLETED = "completed"
    FAILED = "failed"

@dataclass
class CommandNode:
    name: str
    command: Callable
    dependencies: List[str]
    status: CommandStatus = CommandStatus.PENDING
    output: Optional[str] = None
    error: Optional[str] = None

class CommandChainExecutor:
    """Execute commands with dependency resolution."""

    def __init__(self):
        self.commands: Dict[str, CommandNode] = {}
        self.execution_order: List[str] = []

    def register_command(
        self,
        name: str,
        command: Callable,
        dependencies: List[str] = None
    ):
        """Register command with dependencies."""
        self.commands[name] = CommandNode(
            name=name,
            command=command,
            dependencies=dependencies or []
        )

    def resolve_execution_order(self) -> List[str]:
        """Resolve command execution order using topological sort."""
        visited: Set[str] = set()
        order: List[str] = []

        def visit(node_name: str):
            if node_name in visited:
                return
            visited.add(node_name)

            node = self.commands[node_name]
            for dep in node.dependencies:
                visit(dep)

            order.append(node_name)

        for name in self.commands.keys():
            visit(name)

        return order

    async def execute_chain(self) -> Dict[str, str]:
        """Execute commands in dependency order."""
        execution_order = self.resolve_execution_order()
        results = {}

        for cmd_name in execution_order:
            node = self.commands[cmd_name]
            node.status = CommandStatus.RUNNING

            try:
                # Get dependency outputs
                deps_output = {
                    dep: results.get(dep)
                    for dep in node.dependencies
                }

                # Execute command
                output = await node.command(deps_output)
                node.output = output
                node.status = CommandStatus.COMPLETED
                results[cmd_name] = output

                print(f"✓ {cmd_name} completed")

            except Exception as e:
                node.status = CommandStatus.FAILED
                node.error = str(e)
                results[cmd_name] = None
                print(f"✗ {cmd_name} failed: {e}")

        return results

# Usage
executor = CommandChainExecutor()

executor.register_command(
    "checkout",
    async_checkout_feature,
    dependencies=[]
)

executor.register_command(
    "install",
    async_install_deps,
    dependencies=["checkout"]
)

executor.register_command(
    "test",
    async_run_tests,
    dependencies=["install"]
)

executor.register_command(
    "build",
    async_build_project,
    dependencies=["test"]
)

results = await executor.execute_chain()
```

## Pattern 2: Dynamic Command Parameter Processing

Process command parameters with type validation, interpolation, and expansion.

```python
from typing import Union, Any
import re

class ParameterProcessor:
    """Process and validate command parameters."""

    def __init__(self, context: Dict[str, Any]):
        self.context = context
        self.validators: Dict[str, Callable] = {}

    def register_validator(self, param_type: str, validator: Callable):
        """Register custom parameter validator."""
        self.validators[param_type] = validator

    def process_parameter(
        self,
        value: str,
        param_type: str = 'string',
        interpolate: bool = True
    ) -> Any:
        """Process single parameter with validation and interpolation."""

        # Interpolate context variables
        if interpolate:
            value = self._interpolate_variables(value)

        # Type conversion and validation
        if param_type == 'path':
            return self._validate_path(value)
        elif param_type == 'integer':
            return int(value)
        elif param_type == 'boolean':
            return value.lower() in ('true', '1', 'yes')
        elif param_type in self.validators:
            return self.validators[param_type](value)
        else:
            return value

    def _interpolate_variables(self, value: str) -> str:
        """Interpolate {{variable}} patterns."""

        def replace_var(match):
            var_name = match.group(1)
            return str(self.context.get(var_name, match.group(0)))

        return re.sub(r'\{\{(\w+)\}\}', replace_var, value)

    def _validate_path(self, path: str) -> str:
        """Validate file path."""
        from pathlib import Path
        p = Path(path)
        if not p.exists():
            raise ValueError(f"Path does not exist: {path}")
        return str(p.resolve())

# Usage
context = {
    'project_root': '/home/user/project',
    'feature_name': 'auth-feature',
    'version': '1.0.0'
}

processor = ParameterProcessor(context)

# Interpolate variables
path = processor.process_parameter(
    '{{project_root}}/{{feature_name}}',
    param_type='path'
)

count = processor.process_parameter('42', param_type='integer')
enabled = processor.process_parameter('yes', param_type='boolean')
```

## Pattern 3: Conditional Command Execution with Branching

Execute commands conditionally based on previous results or environment state.

```python
from typing import Callable, Optional

@dataclass
class Condition:
    evaluate: Callable[[Dict], bool]
    on_true: str  # Command to execute if true
    on_false: Optional[str] = None  # Command to execute if false

class ConditionalCommandExecutor:
    """Execute commands with conditional branching."""

    def __init__(self):
        self.commands: Dict[str, Callable] = {}
        self.conditions: Dict[str, Condition] = {}
        self.execution_history: List[Dict] = []

    def register_conditional(
        self,
        condition_name: str,
        evaluate: Callable[[Dict], bool],
        on_true: str,
        on_false: Optional[str] = None
    ):
        """Register conditional command branching."""
        self.conditions[condition_name] = Condition(
            evaluate=evaluate,
            on_true=on_true,
            on_false=on_false
        )

    async def execute_with_conditions(
        self,
        start_command: str,
        results: Dict[str, Any]
    ) -> Dict[str, Any]:
        """Execute command chain with conditional branching."""

        current_command = start_command

        while current_command:
            # Check if current command is a condition
            if current_command in self.conditions:
                condition = self.conditions[current_command]

                # Evaluate condition
                if condition.evaluate(results):
                    current_command = condition.on_true
                else:
                    current_command = condition.on_false

                continue

            # Execute regular command
            if current_command in self.commands:
                cmd_func = self.commands[current_command]
                output = await cmd_func(results)

                results[current_command] = output
                self.execution_history.append({
                    'command': current_command,
                    'output': output,
                    'timestamp': time.time()
                })

            current_command = None  # No next command by default

        return results

# Usage
executor = ConditionalCommandExecutor()

# Register commands
executor.commands['build'] = async_build
executor.commands['test'] = async_test
executor.commands['deploy_prod'] = async_deploy_prod
executor.commands['deploy_staging'] = async_deploy_staging

# Register conditions
executor.register_conditional(
    'check_version',
    evaluate=lambda r: r.get('version', '').startswith('v1.'),
    on_true='deploy_staging',
    on_false='deploy_prod'
)

results = await executor.execute_with_conditions('build', {})
```

## Pattern 4: Macro Command Expansion

Define and expand macro commands that contain multiple sub-commands.

```python
from typing import List

class MacroCommand:
    """Reusable command macro with parameter substitution."""

    def __init__(self, name: str, steps: List[Dict]):
        self.name = name
        self.steps = steps  # List of {'command': 'name', 'params': {...}}

    def expand(self, params: Dict[str, str]) -> List[Dict]:
        """Expand macro with parameter substitution."""
        expanded = []

        for step in self.steps:
            # Substitute parameters in command
            cmd = step['command']
            step_params = step.get('params', {})

            # Replace {{param}} patterns
            for param_name, param_value in params.items():
                cmd = cmd.replace(f"{{{{{param_name}}}}}", param_value)
                for key, val in step_params.items():
                    if isinstance(val, str):
                        step_params[key] = val.replace(
                            f"{{{{{param_name}}}}}",
                            param_value
                        )

            expanded.append({
                'command': cmd,
                'params': step_params
            })

        return expanded

class MacroCommandExecutor:
    """Execute macro commands with expansion."""

    def __init__(self):
        self.macros: Dict[str, MacroCommand] = {}
        self.commands: Dict[str, Callable] = {}

    def register_macro(self, macro: MacroCommand):
        """Register reusable macro."""
        self.macros[macro.name] = macro

    async def execute_macro(
        self,
        macro_name: str,
        params: Dict[str, str]
    ) -> List[str]:
        """Execute macro with parameter substitution."""

        if macro_name not in self.macros:
            raise ValueError(f"Macro not found: {macro_name}")

        macro = self.macros[macro_name]
        steps = macro.expand(params)
        results = []

        for step in steps:
            if step['command'] in self.commands:
                result = await self.commands[step['command']](step['params'])
                results.append(result)

        return results

# Usage
executor = MacroCommandExecutor()

# Define macro
deploy_macro = MacroCommand('deploy', [
    {'command': 'checkout', 'params': {'branch': '{{branch}}'}},
    {'command': 'install', 'params': {}},
    {'command': 'test', 'params': {}},
    {'command': 'build', 'params': {}},
    {'command': 'deploy', 'params': {'env': '{{environment}}'}}
])

executor.register_macro(deploy_macro)

# Execute macro with parameters
results = await executor.execute_macro(
    'deploy',
    {'branch': 'feature/new-auth', 'environment': 'staging'}
)
```

## Pattern 5: Command Progress Tracking and Reporting

Track command execution progress with detailed metrics and reporting.

```python
from dataclasses import field
from datetime import datetime

@dataclass
class CommandMetrics:
    command_name: str
    start_time: datetime
    end_time: Optional[datetime] = None
    duration_ms: float = 0
    status: str = "running"
    memory_usage_mb: float = 0
    success: bool = False
    error_message: Optional[str] = None

class ProgressTracker:
    """Track and report command execution progress."""

    def __init__(self):
        self.metrics: List[CommandMetrics] = []
        self.current_command: Optional[CommandMetrics] = None

    def start_command(self, command_name: str):
        """Mark command start."""
        self.current_command = CommandMetrics(
            command_name=command_name,
            start_time=datetime.now()
        )

    def end_command(self, success: bool = True, error: Optional[str] = None):
        """Mark command completion."""
        if self.current_command:
            self.current_command.end_time = datetime.now()
            self.current_command.duration_ms = (
                self.current_command.end_time - self.current_command.start_time
            ).total_seconds() * 1000
            self.current_command.success = success
            self.current_command.status = "completed"
            if error:
                self.current_command.error_message = error
            self.metrics.append(self.current_command)

    def generate_report(self) -> str:
        """Generate execution report."""
        report = ["Command Execution Report\n"]
        report.append("=" * 50)

        total_duration = sum(m.duration_ms for m in self.metrics)
        successful = sum(1 for m in self.metrics if m.success)

        report.append(f"\nTotal Commands: {len(self.metrics)}")
        report.append(f"Successful: {successful}/{len(self.metrics)}")
        report.append(f"Total Duration: {total_duration:.0f}ms\n")

        for metric in self.metrics:
            status_icon = "✓" if metric.success else "✗"
            report.append(
                f"{status_icon} {metric.command_name}: "
                f"{metric.duration_ms:.0f}ms"
            )

        return '\n'.join(report)
```

## Pattern 6: Command Validation and Pre-flight Checks

Validate command preconditions before execution.

```python
from abc import ABC, abstractmethod

class Validator(ABC):
    """Base class for command validators."""

    @abstractmethod
    def validate(self, context: Dict[str, Any]) -> Tuple[bool, str]:
        """Validate preconditions. Returns (valid, error_message)."""
        pass

class FileExistsValidator(Validator):
    """Validate that required files exist."""

    def __init__(self, required_files: List[str]):
        self.required_files = required_files

    def validate(self, context: Dict[str, Any]) -> Tuple[bool, str]:
        from pathlib import Path

        for file in self.required_files:
            if not Path(file).exists():
                return False, f"Required file not found: {file}"
        return True, ""

class EnvironmentValidator(Validator):
    """Validate environment variables."""

    def __init__(self, required_env_vars: List[str]):
        self.required_env_vars = required_env_vars

    def validate(self, context: Dict[str, Any]) -> Tuple[bool, str]:
        import os

        for var in self.required_env_vars:
            if not os.getenv(var):
                return False, f"Required env var not set: {var}"
        return True, ""

class CommandValidator:
    """Validate commands before execution."""

    def __init__(self):
        self.validators: List[Validator] = []

    def add_validator(self, validator: Validator):
        """Add validation rule."""
        self.validators.append(validator)

    async def validate_command(
        self,
        command_name: str,
        context: Dict[str, Any]
    ) -> Tuple[bool, List[str]]:
        """Run all validators."""
        errors = []

        for validator in self.validators:
            valid, error_msg = validator.validate(context)
            if not valid:
                errors.append(error_msg)

        return len(errors) == 0, errors
```

## Pattern 7: Command History and Undo/Redo

Track command history with undo/redo capabilities.

```python
class CommandHistory:
    """Manage command execution history with undo/redo."""

    def __init__(self):
        self.history: List[str] = []
        self.redo_stack: List[str] = []
        self.current_index: int = -1

    def record_command(self, command: str):
        """Record executed command."""
        self.history.append(command)
        self.current_index += 1
        self.redo_stack.clear()  # Clear redo after new command

    def undo(self) -> Optional[str]:
        """Get previous command."""
        if self.current_index > 0:
            self.current_index -= 1
            cmd = self.history[self.current_index]
            self.redo_stack.append(cmd)
            return cmd
        return None

    def redo(self) -> Optional[str]:
        """Get next command."""
        if self.redo_stack:
            cmd = self.redo_stack.pop()
            self.current_index += 1
            return cmd
        return None

    def get_history(self) -> List[str]:
        """Get complete history up to current point."""
        return self.history[:self.current_index + 1]

    def save_history(self, filepath: str):
        """Save history to file."""
        with open(filepath, 'w') as f:
            for cmd in self.get_history():
                f.write(f"{cmd}\n")

    def load_history(self, filepath: str):
        """Load history from file."""
        with open(filepath, 'r') as f:
            self.history = [line.strip() for line in f]
            self.current_index = len(self.history) - 1
```

---

**Advanced Patterns Summary**: 7 enterprise patterns for command chaining, parameter processing, conditional execution, macros, progress tracking, validation, and history management.

