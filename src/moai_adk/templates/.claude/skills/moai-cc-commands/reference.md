# Claude Code Commands - Reference

## API Reference

### Command Class

```python
class Command:
    """Base command class"""

    name: str  # Command name
    description: str  # Command description

    def setup_parameters(self) -> None:
        """Configure parameters"""

    async def execute(self, params: CommandParams) -> dict:
        """Execute command"""

    def validate(self, params: CommandParams) -> bool:
        """Validate parameters"""
```

### CommandRegistry

```python
registry = CommandRegistry()
registry.register(command)  # Register command
command = registry.get(name)  # Retrieve command
```

### Workflow Class

```python
class Workflow:
    """Base workflow class"""

    name: str
    description: str

    def add_step(self, step: WorkflowStep) -> None:
        """Add step"""

    async def execute(self) -> WorkflowResult:
        """Execute workflow"""
```

## Command List

| Command | Description | Parameters |
|---------|-------------|-----------|
| `/moai:0-project` | Project initialization | None |
| `/moai:1-plan` | Generate SPEC | Description string |
| `/moai:2-run` | TDD implementation | SPEC ID |
| `/moai:3-sync` | Document synchronization | SPEC ID |

## Error Codes

| Code | Description |
|------|-------------|
| `CMD_001` | Command not found |
| `CMD_002` | Parameter validation failed |
| `CMD_003` | Command execution failed |
| `WF_001` | Workflow execution failed |

---

**Last Updated**: 2025-11-22
