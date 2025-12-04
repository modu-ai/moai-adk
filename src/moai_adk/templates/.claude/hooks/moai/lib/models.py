"""
Hook model classes - Data structures for hook payloads and results.

Provides HookPayload and HookResult classes used by hook handlers
to process events and return execution results to Claude Code.
"""

from dataclasses import asdict, dataclass
from typing import Any, Dict, List, Optional


class HookPayload(dict):
    """
    A dictionary subclass for hook event payloads.

    Provides dict-like access to hook event data with .get() method support.
    Used to pass event data from Claude Code to hook handlers.

    Example:
        payload = HookPayload({
            "tool": "Read",
            "cwd": "/path/to/project",
            "userPrompt": "Read the file"
        })

        tool_name = payload.get("tool", "")
        working_dir = payload.get("cwd", ".")
    """

    def __init__(self, data: Optional[Dict[str, Any]] = None):
        """Initialize HookPayload with optional data dictionary."""
        super().__init__(data or {})

    def get(self, key: str, default: Any = None) -> Any:
        """Get value from payload with optional default."""
        return super().get(key, default)

    def __setitem__(self, key: str, value: Any) -> None:
        """Set value in payload."""
        super().__setitem__(key, value)

    def update(self, other: Dict[str, Any]) -> None:  # type: ignore[override]
        """Update payload with another dictionary."""
        super().update(other)


@dataclass
class HookResult:
    """
    A class representing the result of a hook execution.

    Used by hook handlers to return execution results, messages, and metadata
    back to Claude Code. Supports JSON serialization via to_dict() method.

    Attributes:
        system_message (Optional[str]): Message to display to user
        continue_execution (bool): Whether to continue execution (default: True)
        context_files (List[str]): List of context file paths to load
        hook_specific_output (Optional[Dict[str, Any]]): Hook-specific data
        block_execution (bool): Whether to block execution (default: False)

    Example:
        # Simple result with just a message
        return HookResult(system_message="Operation completed")

        # Result with context files
        return HookResult(
            system_message="Loaded 3 context files",
            context_files=["README.md", "config.json"]
        )

        # Result that stops execution
        return HookResult(
            system_message="Dangerous operation blocked",
            continue_execution=False,
            block_execution=True
        )
    """

    system_message: Optional[str] = None
    continue_execution: bool = True
    context_files: List[str] = None
    hook_specific_output: Optional[Dict[str, Any]] = None
    block_execution: bool = False

    def __post_init__(self):
        """Post-initialization to set default values."""
        if self.context_files is None:
            self.context_files = []
        if self.hook_specific_output is None:
            self.hook_specific_output = {}

    def to_dict(self) -> Dict[str, Any]:
        """
        Convert HookResult to a dictionary for JSON serialization.

        Returns:
            Dict[str, Any]: Dictionary representation of the HookResult
        """
        result = asdict(self)
        # Remove empty/None values to keep output clean
        return {k: v for k, v in result.items() if v or v is False or v == 0}
