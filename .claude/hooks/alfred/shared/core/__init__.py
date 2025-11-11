#!/usr/bin/env python3
"""Core module for Alfred Hooks

Common type definitions and utility functions
"""

from dataclasses import dataclass, field
from typing import Any, Literal, NotRequired, TypedDict


class HookPayload(TypedDict):
    """Claude Code Hook event payload type definition

    Data structure that Claude Code passes to the Hook script.
    Use NotRequired because fields may vary depending on the event.
    """

    cwd: str
    userPrompt: NotRequired[str]  # Includes only UserPromptSubmit events
    tool: NotRequired[str]  # PreToolUse/PostToolUse events
    arguments: NotRequired[dict[str, Any]]  # Tool arguments


@dataclass
class HookResult:
    """Hook execution result following Claude Code standard schema.

    Attributes conform to Claude Code Hook output specification:
    https://docs.claude.com/en/docs/claude-code/hooks

    Standard Fields (Claude Code schema - included in JSON output):
        continue_execution: Allow execution to continue (default True)
        suppress_output: Suppress hook output display (default False)
        decision: "approve" or "block" operation (optional)
        reason: Explanation for decision (optional)
        permission_decision: "allow", "deny", or "ask" (optional)
        system_message: Message displayed to user (top-level field)

    Internal Fields (MoAI-ADK only - NOT in JSON output):
        context_files: List of context files to load (internal use only)
        suggestions: Suggestions for user (internal use only)
        exit_code: Exit code for diagnostics (internal use only)

    Note:
        - systemMessage appears at TOP LEVEL in JSON output
        - hookSpecificOutput is ONLY used for UserPromptSubmit events
        - Internal fields are used for Python logic but not serialized to JSON
    """

    # Claude Code standard fields
    continue_execution: bool = True
    suppress_output: bool = False
    decision: Literal["approve", "block"] | None = None
    reason: str | None = None
    permission_decision: Literal["allow", "deny", "ask"] | None = None

    # MoAI-ADK custom fields (wrapped in hookSpecificOutput)
    system_message: str | None = None
    context_files: list[str] = field(default_factory=list)
    suggestions: list[str] = field(default_factory=list)
    exit_code: int = 0

    def to_dict(self) -> dict[str, Any]:
        """Convert to Claude Code standard Hook output schema.

        Returns:
            Dictionary conforming to Claude Code Hook specification with:
            - Top-level fields: continue, suppressOutput, decision, reason,
              permissionDecision, systemMessage
            - MoAI-ADK internal fields (context_files, suggestions, exit_code)
              are NOT included in JSON output (used for internal logic only)

        Examples:
            >>> result = HookResult(continue_execution=True)
            >>> result.to_dict()
            {'continue': True}

            >>> result = HookResult(decision="block", reason="Dangerous")
            >>> result.to_dict()
            {'decision': 'block', 'reason': 'Dangerous'}

            >>> result = HookResult(system_message="Test")
            >>> result.to_dict()
            {'continue': True, 'systemMessage': 'Test'}

        Note:
            - systemMessage is a TOP-LEVEL field (not nested in hookSpecificOutput)
            - hookSpecificOutput is ONLY used for UserPromptSubmit events
            - context_files, suggestions, exit_code are internal-only fields
        """
        output: dict[str, Any] = {}

        # Add decision or continue flag
        if self.decision:
            output["decision"] = self.decision
        else:
            output["continue"] = self.continue_execution

        # Add reason if provided (works with both decision and permissionDecision)
        if self.reason:
            output["reason"] = self.reason

        # Add suppressOutput if True
        if self.suppress_output:
            output["suppressOutput"] = True

        # Add permissionDecision if set
        if self.permission_decision:
            output["permissionDecision"] = self.permission_decision

        # Add systemMessage at TOP LEVEL (required by Claude Code schema)
        if self.system_message:
            output["systemMessage"] = self.system_message

        # Note: context_files, suggestions, exit_code are internal-only fields
        # and are NOT included in the JSON output per Claude Code schema

        return output

    def to_user_prompt_submit_dict(self) -> dict[str, Any]:
        """UserPromptSubmit Hook-specific output format.

        Claude Code requires a special schema for UserPromptSubmit events.
        The result is wrapped in the standard Hook schema with hookSpecificOutput.

        Returns:
            Claude Code UserPromptSubmit Hook Dictionary matching schema:
            {
                "continue": true,
                "hookSpecificOutput": {
                    "hookEventName": "UserPromptSubmit",
                    "additionalContext": "string"
                }
            }

        Examples:
            >>> result = HookResult(context_files=["tests/"])
            >>> result.to_user_prompt_submit_dict()
            {'continue': True, 'hookSpecificOutput': \
{'hookEventName': 'UserPromptSubmit', 'additionalContext': '📎 Context: tests/'}}
        """
        # Convert context_files to additionalContext string
        if self.context_files:
            context_str = "\n".join([f"📎 Context: {f}" for f in self.context_files])
        else:
            context_str = ""

        # Add system_message if there is one
        if self.system_message:
            if context_str:
                context_str = f"{self.system_message}\n\n{context_str}"
            else:
                context_str = self.system_message

        return {
            "continue": self.continue_execution,
            "hookSpecificOutput": {
                "hookEventName": "UserPromptSubmit",
                "additionalContext": context_str,
            },
        }


import json
import logging
import os
import time
import threading
from dataclasses import dataclass, field
from pathlib import Path
from typing import Any, Dict, List, Literal, NotRequired, TypedDict


@dataclass
class HookConfiguration:
    """Centralized configuration for hook deduplication and execution behavior

    Provides configurable parameters for:
    - Command deduplication time windows
    - Hook execution deduplication settings
    - State management options
    - Performance tuning parameters
    - Debugging and logging options
    """

    # Time windows for deduplication (in seconds)
    command_dedupe_window: float = 3.0
    hook_dedupe_window: float = 1.0
    state_cache_ttl: float = 5.0

    # State management settings
    max_state_file_age_hours: int = 24
    max_execution_history_size: int = 1000
    enable_state_persistence: bool = True
    fallback_to_temp: bool = True

    # Performance optimization settings
    enable_caching: bool = True
    cache_cleanup_interval: int = 300  # 5 minutes
    enable_performance_monitoring: bool = True

    # Error handling settings
    graceful_degradation: bool = True
    max_retries_on_io_error: int = 3
    retry_delay_seconds: float = 0.1

    # Debugging and logging settings
    debug_mode: bool = False
    enable_verbose_logging: bool = False
    log_state_changes: bool = False

    # Concurrency settings
    enable_thread_safety: bool = True
    lock_timeout_seconds: float = 5.0
    enable_concurrent_access: bool = True

    # File I/O settings
    state_file_encoding: str = "utf-8"
    state_file_indent: int = 2
    backup_on_write: bool = True

    @classmethod
    def from_env(cls) -> "HookConfiguration":
        """Load configuration from environment variables

        Returns:
            HookConfiguration instance with values from environment variables
            or defaults if variables are not set
        """
        return cls(
            command_dedupe_window=float(os.getenv("ALFRED_COMMAND_DEDUPE_WINDOW", "3.0")),
            hook_dedupe_window=float(os.getenv("ALFRED_HOOK_DEDUPE_WINDOW", "1.0")),
            state_cache_ttl=float(os.getenv("ALFRED_STATE_CACHE_TTL", "5.0")),
            max_state_file_age_hours=int(os.getenv("ALFRED_MAX_STATE_FILE_AGE_HOURS", "24")),
            max_execution_history_size=int(os.getenv("ALFRED_MAX_EXECUTION_HISTORY_SIZE", "1000")),
            enable_state_persistence=os.getenv("ALFRED_ENABLE_STATE_PERSISTENCE", "true").lower() == "true",
            fallback_to_temp=os.getenv("ALFRED_FALLBACK_TO_TEMP", "true").lower() == "true",
            enable_caching=os.getenv("ALFRED_ENABLE_CACHING", "true").lower() == "true",
            cache_cleanup_interval=int(os.getenv("ALFRED_CACHE_CLEANUP_INTERVAL", "300")),
            enable_performance_monitoring=os.getenv("ALFRED_ENABLE_PERFORMANCE_MONITORING", "true").lower() == "true",
            graceful_degradation=os.getenv("ALFRED_GRACEFUL_DEGRADATION", "true").lower() == "true",
            max_retries_on_io_error=int(os.getenv("ALFRED_MAX_RETRIES_ON_IO_ERROR", "3")),
            retry_delay_seconds=float(os.getenv("ALFRED_RETRY_DELAY_SECONDS", "0.1")),
            debug_mode=os.getenv("ALFRED_DEBUG_MODE", "false").lower() == "true",
            enable_verbose_logging=os.getenv("ALFRED_ENABLE_VERBOSE_LOGGING", "false").lower() == "true",
            log_state_changes=os.getenv("ALFRED_LOG_STATE_CHANGES", "false").lower() == "true",
            enable_thread_safety=os.getenv("ALFRED_ENABLE_THREAD_SAFETY", "true").lower() == "true",
            lock_timeout_seconds=float(os.getenv("ALFRED_LOCK_TIMEOUT_SECONDS", "5.0")),
            enable_concurrent_access=os.getenv("ALFRED_ENABLE_CONCURRENT_ACCESS", "true").lower() == "true",
            state_file_encoding=os.getenv("ALFRED_STATE_FILE_ENCODING", "utf-8"),
            state_file_indent=int(os.getenv("ALFRED_STATE_FILE_INDENT", "2")),
            backup_on_write=os.getenv("ALFRED_BACKUP_ON_WRITE", "true").lower() == "true",
        )

    def get_state_dir(self, cwd: str) -> Path:
        """Get state directory path with fallback logic

        Args:
            cwd: Current working directory

        Returns:
            Path to state directory (either .moai/memory or temp fallback)
        """
        try:
            # Primary: Use .moai/memory directory
            state_dir = Path(cwd) / ".moai" / "memory"
            state_dir.mkdir(parents=True, exist_ok=True)

            # Test if directory is writable
            test_file = state_dir / ".test_write"
            test_file.touch()
            test_file.unlink()

            return state_dir
        except (OSError, PermissionError) as e:
            if self.fallback_to_temp:
                # Fallback to temporary directory
                import tempfile
                temp_dir = Path(tempfile.gettempdir()) / "alfred_hooks" / cwd.replace("/", "_").replace("\\", "_")
                temp_dir.mkdir(parents=True, exist_ok=True)
                return temp_dir
            else:
                raise e


@dataclass
class ExecutionResult:
    """Standardized result format for hook and command execution

    Provides consistent return format for deduplication operations
    with detailed information about execution decisions and outcomes.
    """

    # Basic execution info
    executed: bool
    duplicate: bool
    execution_id: str
    timestamp: float

    # Context information
    hook_name: str = None
    command: str = None
    phase: str = None
    reason: str = None

    # Performance metrics
    execution_time_ms: float = 0.0
    cache_hit: bool = False
    state_operations_count: int = 0

    # Counters
    execution_count: int = 0
    duplicate_count: int = 0

    # Error information
    error: str = None
    warning: str = None

    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for JSON serialization

        Returns:
            Dictionary representation of execution result
        """
        return {
            "executed": self.executed,
            "duplicate": self.duplicate,
            "execution_id": self.execution_id,
            "timestamp": self.timestamp,
            "hook_name": self.hook_name,
            "command": self.command,
            "phase": self.phase,
            "reason": self.reason,
            "execution_time_ms": self.execution_time_ms,
            "cache_hit": self.cache_hit,
            "state_operations_count": self.state_operations_count,
            "execution_count": self.execution_count,
            "duplicate_count": self.duplicate_count,
            "error": self.error,
            "warning": self.warning,
        }


@dataclass
class PerformanceMetrics:
    """Performance monitoring and benchmarking metrics

    Tracks execution times, cache hits, error rates, and other
    performance indicators for hook deduplication system.
    """

    # Basic counters
    total_executions: int = 0
    successful_executions: int = 0
    failed_executions: int = 0
    duplicate_executions: int = 0

    # Time tracking
    total_execution_time: float = 0.0
    average_execution_time: float = 0.0
    max_execution_time: float = 0.0
    min_execution_time: float = 0.0

    # Cache metrics
    cache_hits: int = 0
    cache_misses: int = 0
    cache_hit_rate: float = 0.0

    # Error metrics
    io_errors: int = 0
    concurrent_access_errors: int = 0
    other_errors: int = 0

    # State management metrics
    state_file_writes: int = 0
    state_file_reads: int = 0
    state_file_creates: int = 0

    def record_execution(self, execution_time: float, success: bool, is_duplicate: bool = False):
        """Record execution metrics

        Args:
            execution_time: Time taken for execution in milliseconds
            success: Whether execution was successful
            is_duplicate: Whether this was a duplicate execution
        """
        self.total_executions += 1
        self.total_execution_time += execution_time

        if success:
            self.successful_executions += 1
        else:
            self.failed_executions += 1

        if is_duplicate:
            self.duplicate_executions += 1

        # Update min/max execution times
        if self.max_execution_time == 0 or execution_time > self.max_execution_time:
            self.max_execution_time = execution_time
        if self.min_execution_time == 0 or execution_time < self.min_execution_time:
            self.min_execution_time = execution_time

        # Update average
        self.average_execution_time = self.total_execution_time / self.total_executions

    def record_cache_hit(self):
        """Record a cache hit"""
        self.cache_hits += 1

    def record_cache_miss(self):
        """Record a cache miss"""
        self.cache_misses += 1
        self.update_cache_hit_rate()

    def update_cache_hit_rate(self):
        """Update cache hit rate percentage"""
        total_cache_ops = self.cache_hits + self.cache_misses
        if total_cache_ops > 0:
            self.cache_hit_rate = (self.cache_hits / total_cache_ops) * 100

    def record_io_error(self):
        """Record an I/O error"""
        self.io_errors += 1

    def record_concurrent_access_error(self):
        """Record a concurrent access error"""
        self.concurrent_access_errors += 1

    def record_other_error(self):
        """Record other errors"""
        self.other_errors += 1

    def record_state_write(self):
        """Record a state file write operation"""
        self.state_file_writes += 1

    def record_state_read(self):
        """Record a state file read operation"""
        self.state_file_reads += 1

    def record_state_create(self):
        """Record a state file creation operation"""
        self.state_file_creates += 1

    def get_summary(self) -> Dict[str, Any]:
        """Get performance summary

        Returns:
            Dictionary with performance summary statistics
        """
        return {
            "total_executions": self.total_executions,
            "successful_executions": self.successful_executions,
            "failed_executions": self.failed_executions,
            "duplicate_executions": self.duplicate_executions,
            "success_rate": (self.successful_executions / self.total_executions * 100) if self.total_executions > 0 else 0,
            "duplicate_rate": (self.duplicate_executions / self.total_executions * 100) if self.total_executions > 0 else 0,
            "average_execution_time_ms": self.average_execution_time,
            "max_execution_time_ms": self.max_execution_time,
            "min_execution_time_ms": self.min_execution_time,
            "cache_hit_rate": self.cache_hit_rate,
            "io_errors": self.io_errors,
            "concurrent_access_errors": self.concurrent_access_errors,
            "other_errors": self.other_errors,
            "state_operations_count": self.state_file_writes + self.state_file_reads + self.state_file_creates,
        }


# Global performance metrics instance
_global_performance_metrics = PerformanceMetrics()
_global_metrics_lock = threading.RLock()


def get_performance_metrics() -> PerformanceMetrics:
    """Get global performance metrics instance

    Returns:
        Global performance metrics instance
    """
    return _global_performance_metrics


def record_execution_metrics(execution_time: float, success: bool, is_duplicate: bool = False):
    """Record execution metrics globally

    Args:
        execution_time: Time taken for execution in milliseconds
        success: Whether execution was successful
        is_duplicate: Whether this was a duplicate execution
    """
    with _global_metrics_lock:
        _global_performance_metrics.record_execution(execution_time, success, is_duplicate)


def record_cache_hit():
    """Record a cache hit globally"""
    with _global_metrics_lock:
        _global_performance_metrics.record_cache_hit()


def record_cache_miss():
    """Record a cache miss globally"""
    with _global_metrics_lock:
        _global_performance_metrics.record_cache_miss()


# Configure logging
def configure_logging(debug_mode: bool = False, verbose: bool = False):
    """Configure logging for hook system

    Args:
        debug_mode: Enable debug mode logging
        verbose: Enable verbose logging
    """
    # Create memory directory if it doesn't exist
    memory_dir = Path.cwd() / ".moai" / "memory"
    memory_dir.mkdir(parents=True, exist_ok=True)

    level = logging.DEBUG if debug_mode or verbose else logging.INFO

    handlers = [logging.StreamHandler()]

    # Try to add file handler, but don't fail if it doesn't work
    try:
        file_handler = logging.FileHandler(memory_dir / "hook_debug.log", mode='a')
        handlers.append(file_handler)
    except (OSError, IOError) as e:
        # Fallback to console only if file logging fails
        print(f"Warning: Could not create file handler: {e}")

    logging.basicConfig(
        level=level,
        format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
        handlers=handlers,
        force=True  # Override existing configuration
    )


# Initialize logging with default configuration
configure_logging()

# Global logger instance
_logger = logging.getLogger("alfred.hooks")


def get_logger() -> logging.Logger:
    """Get the global logger instance

    Returns:
        Global logger instance for hook system
    """
    return _logger


__all__ = [
    "HookPayload",
    "HookResult",
    "HookConfiguration",
    "ExecutionResult",
    "PerformanceMetrics",
    "configure_logging",
    "get_logger",
    "get_performance_metrics",
    "record_execution_metrics",
    "record_cache_hit",
    "record_cache_miss",
]
