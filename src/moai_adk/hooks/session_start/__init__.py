"""Session start hook module

Provides hooks for automatically cleaning up old files and generating daily analysis reports.

This module contains 5 specialized submodules:
- state_cleanup: Configuration and state management
- core_cleanup: File system cleanup operations
- analysis_report: Session analysis and reporting
- orchestrator: Overall execution coordination
"""

from moai_adk.hooks.session_start.analysis_report import (
    AnalysisError,
    analyze_session_logs,
    format_analysis_report,
    generate_daily_analysis,
)
from moai_adk.hooks.session_start.core_cleanup import (
    CleanupError,
    cleanup_directory,
    cleanup_old_files,
    update_cleanup_stats,
)
from moai_adk.hooks.session_start.orchestrator import (
    OrchestratorError,
    execute_analysis_sequence,
    execute_cleanup_sequence,
    format_hook_result,
    handle_timeout,
    main,
)
from moai_adk.hooks.session_start.state_cleanup import (
    StateError,
    get_graceful_degradation,
    load_config,
    load_hook_timeout,
    should_cleanup_today,
    validate_cleanup_config,
)

__version__ = "0.26.0"

__all__ = [
    # Version
    "__version__",
    # State cleanup exports
    "StateError",
    "load_hook_timeout",
    "get_graceful_degradation",
    "load_config",
    "should_cleanup_today",
    "validate_cleanup_config",
    # Core cleanup exports
    "CleanupError",
    "cleanup_old_files",
    "cleanup_directory",
    "update_cleanup_stats",
    # Analysis report exports
    "AnalysisError",
    "generate_daily_analysis",
    "analyze_session_logs",
    "format_analysis_report",
    # Orchestrator exports
    "OrchestratorError",
    "execute_cleanup_sequence",
    "execute_analysis_sequence",
    "format_hook_result",
    "handle_timeout",
    "main",
]
