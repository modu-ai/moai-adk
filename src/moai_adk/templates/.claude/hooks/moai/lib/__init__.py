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
    from lib.models import HookPayload, HookResult

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

    # Import handlers
    from lib.session import handle_session_start
    from lib.tool import handle_pre_tool_use, handle_post_tool_use

    # Import utilities
    from lib.common import format_duration

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
