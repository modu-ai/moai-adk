"""
Hook utilities library - Optimized and consolidated

This module provides centralized access to all hook-related utilities:
- Configuration management (ConfigManager)
- Core utilities (timeout, common utilities)
- Data models (HookPayload, HookResult)
- Project utilities (language detection, git info, SPEC counting)
- Checkpoint utilities (risky operation detection, checkpoint creation)
"""

try:
    # Import core components
    # Import checkpoint utilities
    from lib.checkpoint import (
        create_checkpoint,
        detect_risky_operation,
        list_checkpoints,
    )

    # Import utilities
    from lib.common import (
        format_duration,
        get_file_pattern_category,
        get_summary_stats,
        is_root_whitelisted,
        suggest_moai_location,
    )
    from lib.config_manager import (
        ConfigManager,
        get_config,
        get_config_manager,
        get_exit_code,
        get_graceful_degradation,
        get_timeout_seconds,
    )

    # Import models
    from lib.models import HookPayload, HookResult

    # Import project utilities
    from lib.project import (
        count_specs,
        find_project_root,
        get_git_info,
    )
    from lib.timeout import CrossPlatformTimeout, TimeoutError, timeout_context

    __all__ = [
        # Core - Timeout
        "CrossPlatformTimeout",
        "TimeoutError",
        "timeout_context",
        # Core - Configuration
        "ConfigManager",
        "get_config_manager",
        "get_config",
        "get_timeout_seconds",
        "get_graceful_degradation",
        "get_exit_code",
        # Common utilities
        "format_duration",
        "get_summary_stats",
        "is_root_whitelisted",
        "get_file_pattern_category",
        "suggest_moai_location",
        # Models
        "HookPayload",
        "HookResult",
        # Checkpoint
        "create_checkpoint",
        "detect_risky_operation",
        "list_checkpoints",
        # Project
        "find_project_root",
        "get_git_info",
        "count_specs",
    ]

except ImportError:
    # Fallback if not all imports are available
    __all__ = []

__version__ = "1.0.0"
__author__ = "MoAI-ADK Team"
