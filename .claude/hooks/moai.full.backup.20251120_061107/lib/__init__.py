"""
Hook utilities library - Consolidated from shared/, utils/, and handlers/

This module provides centralized access to all hook-related utilities:
- Configuration management (ConfigManager)
- Core utilities (timeout, error handling, JSON utilities)
- Event handlers (session, tool, user, notification)
- Project and state tracking
"""

from typing import Any, Dict, List, Optional
from dataclasses import dataclass, asdict


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

    def update(self, other: Dict[str, Any]) -> None:
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


try:
    # Import core components
    from lib.config_manager import (
        ConfigManager,
        get_config_manager,
        get_config,
        get_timeout_seconds,
        get_graceful_degradation,
        get_exit_code,
    )
    from lib.timeout import CrossPlatformTimeout
    from lib.error_handler import HookError, ToolError, handle_hook_error
    from lib.json_utils import safe_json_loads, safe_json_dumps
    from lib.agent_context import AgentContext, get_agent_context

    # Import handlers
    from lib.session import handle_session_start
    from lib.tool import handle_pre_tool_use, handle_post_tool_use
    from lib.user import handle_user_prompt_submit
    from lib.notification import handle_subagent_start, handle_subagent_stop

    # Import utilities
    from lib.common import format_duration, get_module_version
    from lib.hook_config import (
        load_hook_timeout,
        get_hook_execution_config,
    )

    __all__ = [
        # Hook payload/result classes
        "HookPayload",
        "HookResult",

        # Configuration
        "ConfigManager",
        "get_config_manager",
        "get_config",
        "get_timeout_seconds",
        "get_graceful_degradation",
        "get_exit_code",

        # Core
        "CrossPlatformTimeout",
        "HookError",
        "ToolError",
        "handle_hook_error",
        "safe_json_loads",
        "safe_json_dumps",
        "AgentContext",
        "get_agent_context",

        # Handlers
        "handle_session_start",
        "handle_pre_tool_use",
        "handle_post_tool_use",
        "handle_user_prompt_submit",
        "handle_subagent_start",
        "handle_subagent_stop",

        # Utilities
        "format_duration",
        "get_module_version",
        "load_hook_timeout",
        "get_hook_execution_config",
    ]

except ImportError:
    # Fallback if not all imports are available
    __all__ = [
        # Always include hook payload/result classes
        "HookPayload",
        "HookResult",
    ]

__version__ = "1.0.0"
__author__ = "MoAI-ADK Team"
