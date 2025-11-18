"""
Hook utilities library - Consolidated from shared/, utils/, and handlers/

This module provides centralized access to all hook-related utilities:
- Configuration management (ConfigManager)
- Core utilities (timeout, error handling, JSON utilities)
- Event handlers (session, tool, user, notification)
- Project and state tracking
"""

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
    __all__ = []

__version__ = "1.0.0"
__author__ = "MoAI-ADK Team"
