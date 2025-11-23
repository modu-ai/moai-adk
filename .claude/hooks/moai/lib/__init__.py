"""
Hook utilities library - Consolidated from shared/, utils/, and handlers/

This module provides centralized access to all hook-related utilities:
- Configuration management (ConfigManager)
- Core utilities (timeout, error handling, JSON utilities)
- Event handlers (session, tool)
- Project and state tracking
"""

try:
    # Import model classes
    # Import utilities
    from lib.common import format_duration

    # Import core components
    from lib.config_manager import (
        ConfigManager,
        get_config,
        get_config_manager,
        get_exit_code,
        get_graceful_degradation,
        get_timeout_seconds,
    )
    from lib.models import HookPayload, HookResult

    # Import handlers
    from lib.session import handle_session_start
    from lib.timeout import CrossPlatformTimeout
    from lib.tool import handle_post_tool_use, handle_pre_tool_use

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

        # Handlers
        "handle_session_start",
        "handle_pre_tool_use",
        "handle_post_tool_use",

        # Utilities
        "format_duration",
    ]

except ImportError:
    # Fallback if not all imports are available
    __all__ = []

__version__ = "1.0.0"
__author__ = "MoAI-ADK Team"
