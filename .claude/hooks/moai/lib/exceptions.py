#!/usr/bin/env python3
"""Exception hierarchy for MoAI-ADK hooks system.

Provides a consistent exception hierarchy for all hooks-related errors.
This enables structured error handling and proper exception chaining.

Hierarchy:
    HooksBaseError
    ├── TimeoutError (base timeout from timeout.py)
    │   └── HookTimeoutError (enhanced with context)
    ├── GitOperationError
    ├── ConfigurationError
    └── ValidationError
"""

from __future__ import annotations


class HooksBaseError(Exception):
    """Base exception for all hooks system errors.

    All hook-related exceptions should inherit from this class
    to enable consistent exception handling.
    """

    pass


# Re-export TimeoutError from timeout module for convenience
try:
    from .timeout import TimeoutError
except ImportError:

    class TimeoutError(HooksBaseError):  # type: ignore[no-redef]
        """Timeout exception (fallback if timeout module unavailable)."""

        pass


# Re-export HookTimeoutError from unified_timeout_manager for convenience
try:
    from .unified_timeout_manager import HookTimeoutError
except ImportError:

    class HookTimeoutError(TimeoutError):  # type: ignore[no-redef]
        """Hook-specific timeout with context (fallback)."""

        def __init__(
            self,
            message: str,
            hook_id: str = "",
            timeout_seconds: float = 0,
            execution_time: float = 0,
            will_retry: bool = False,
        ):
            super().__init__(message)
            self.hook_id = hook_id
            self.timeout_seconds = timeout_seconds
            self.execution_time = execution_time
            self.will_retry = will_retry


class GitOperationError(HooksBaseError):
    """Exception for Git operation failures.

    Attributes:
        command: The Git command that failed
        exit_code: Git command exit code (if available)
        stderr: Error output from Git command
    """

    def __init__(
        self,
        message: str,
        command: str = "",
        exit_code: int | None = None,
        stderr: str = "",
    ):
        super().__init__(message)
        self.command = command
        self.exit_code = exit_code
        self.stderr = stderr

    def __str__(self) -> str:
        parts = [super().__str__()]
        if self.command:
            parts.append(f"Command: {self.command}")
        if self.exit_code is not None:
            parts.append(f"Exit code: {self.exit_code}")
        if self.stderr:
            parts.append(f"Stderr: {self.stderr[:200]}")
        return " | ".join(parts)


class ConfigurationError(HooksBaseError):
    """Exception for configuration-related errors.

    Raised when configuration files are missing, malformed,
    or contain invalid values.

    Attributes:
        config_path: Path to the configuration file (if applicable)
        key: Configuration key that caused the error (if applicable)
    """

    def __init__(
        self,
        message: str,
        config_path: str = "",
        key: str = "",
    ):
        super().__init__(message)
        self.config_path = config_path
        self.key = key


class ValidationError(HooksBaseError):
    """Exception for validation failures.

    Raised when input validation or schema validation fails.

    Attributes:
        field: The field that failed validation
        value: The invalid value (if safe to include)
        expected: Description of expected value format
    """

    def __init__(
        self,
        message: str,
        field: str = "",
        value: str | None = None,
        expected: str = "",
    ):
        super().__init__(message)
        self.field = field
        self.value = value
        self.expected = expected


class SecurityError(HooksBaseError):
    """Exception for security-related issues.

    Raised when security checks fail (e.g., path traversal attempts,
    unauthorized file access).

    Attributes:
        operation: The operation that was blocked
        resource: The resource that was accessed
    """

    def __init__(
        self,
        message: str,
        operation: str = "",
        resource: str = "",
    ):
        super().__init__(message)
        self.operation = operation
        self.resource = resource


__all__ = [
    "HooksBaseError",
    "TimeoutError",
    "HookTimeoutError",
    "GitOperationError",
    "ConfigurationError",
    "ValidationError",
    "SecurityError",
]
