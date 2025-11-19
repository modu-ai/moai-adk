#!/usr/bin/env python3
"""Standard Error Handler for Alfred Hooks

Provides consistent error handling, logging, and response formatting across all hooks.
"""

import json
import logging
import os
import sys
import time
from datetime import datetime
from typing import Any, Dict, Optional

from lib.config_manager import get_config_manager, get_timeout_seconds


class HookErrorHandler:
    """Standardized error handling for Alfred hooks.

    Provides consistent error responses, logging, and graceful degradation.
    """

    def __init__(self, hook_name: str, timeout_seconds: Optional[int] = None):
        """Initialize error handler.

        Args:
            hook_name: Name of the hook for logging
            timeout_seconds: Default timeout for operations (uses config if None)
        """
        self.hook_name = hook_name
        self.timeout_seconds = timeout_seconds or get_timeout_seconds()
        self.start_time = time.time()

        # Setup logging
        self._setup_logging()

    def _setup_logging(self):
        """Setup logging configuration."""
        # Create logger with hook name
        self.logger = logging.getLogger(f"alfred.hooks.{self.hook_name}")

        # Create console handler if not exists
        if not self.logger.handlers:
            handler = logging.StreamHandler(sys.stderr)
            formatter = logging.Formatter(
                f"[{self.hook_name}] [%(levelname)s] %(message)s"
            )
            handler.setFormatter(formatter)
            self.logger.addHandler(handler)

            # Set level based on environment
            log_level = getattr(logging, os.environ.get("HOOK_LOG_LEVEL", "WARNING").upper())
            self.logger.setLevel(log_level)

    def create_response(
        self,
        success: bool = True,
        error: Optional[str] = None,
        message: Optional[str] = None,
        data: Optional[Dict[str, Any]] = None,
        graceful_degradation: bool = False,
        execution_time: Optional[float] = None
    ) -> Dict[str, Any]:
        """Create standardized response structure.

        Args:
            success: Whether operation was successful
            error: Error message if failed
            message: Descriptive message
            data: Additional data payload
            graceful_degradation: Whether graceful degradation was applied
            execution_time: How long the operation took

        Returns:
            Standardized response dictionary
        """
        response = {
            "continue": True,  # Always allow operation to continue by default
            "success": success,
            "hook_name": self.hook_name,
            "timestamp": datetime.now().isoformat(),
            "execution_time": execution_time or (time.time() - self.start_time),
            "graceful_degradation": graceful_degradation,
        }

        if error:
            response["error"] = error
        if message:
            response["message"] = message
        if data:
            response["data"] = data

        return response

    def handle_timeout(self, operation_name: str = "operation") -> Dict[str, Any]:
        """Handle timeout errors consistently.

        Args:
            operation_name: Name of the timed-out operation

        Returns:
            Timeout error response
        """
        error_msg = f"{operation_name} timed out after {self.timeout_seconds} seconds"
        self.logger.warning(error_msg)

        # Get localized message from config
        message = get_config_manager().get_message("timeout", self.hook_name, "timeout")
        if message == f"Message not found: timeout.{self.hook_name}.timeout":
            message = f"⏰ {operation_name} timed out after {self.timeout_seconds} seconds"

        return self.create_response(
            success=False,
            error=error_msg,
            message=message,
            graceful_degradation=True
        )

    def handle_json_error(self, error: Exception, context: str = "JSON parsing") -> Dict[str, Any]:
        """Handle JSON parsing errors consistently.

        Args:
            error: The JSONDecodeError exception
            context: Context where the error occurred

        Returns:
            JSON error response
        """
        error_msg = f"{context} error: {str(error)}"
        self.logger.error(error_msg)

        return self.create_response(
            success=False,
            error=error_msg,
            message=f"⚠️ {error_msg}",
            graceful_degradation=True
        )

    def handle_import_error(self, module_name: str) -> Dict[str, Any]:
        """Handle import errors with fallback information.

        Args:
            module_name: Name of the module that failed to import

        Returns:
            Import error response
        """
        error_msg = f"Module '{module_name}' not available"
        self.logger.error(error_msg)

        return self.create_response(
            success=False,
            error=error_msg,
            message=f"⚠️ {error_msg}. Using fallback implementation.",
            graceful_degradation=True
        )

    def handle_generic_error(self, error: Exception, context: str = "operation") -> Dict[str, Any]:
        """Handle generic exceptions consistently.

        Args:
            error: The exception that occurred
            context: Context where the error occurred

        Returns:
            Generic error response
        """
        error_msg = f"{context} error: {str(error)}"
        self.logger.error(error_msg)

        return self.create_response(
            success=False,
            error=error_msg,
            message=f"❌ {error_msg}",
            graceful_degradation=True
        )

    def handle_config_error(self, error: Exception, config_path: str) -> Dict[str, Any]:
        """Handle configuration loading errors.

        Args:
            error: The exception that occurred
            config_path: Path to the config file

        Returns:
            Config error response
        """
        error_msg = f"Failed to load config from {config_path}: {str(error)}"
        self.logger.error(error_msg)

        return self.create_response(
            success=False,
            error=error_msg,
            message=f"⚙️ {error_msg}. Using defaults.",
            graceful_degradation=True
        )

    def handle_success(self, message: str = "Operation completed successfully", data: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        """Handle successful operations.

        Args:
            message: Success message
            data: Additional data to include

        Returns:
            Success response
        """
        self.logger.info(message)

        return self.create_response(
            success=True,
            message=message,
            data=data
        )

    def create_success(self, message: str = "Operation completed successfully", data: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        """Alias for handle_success for backward compatibility.

        Args:
            message: Success message
            data: Additional data to include

        Returns:
            Success response (same as handle_success)
        """
        return self.handle_success(message, data)

    def print_and_exit(self, response: Dict[str, Any], exit_code: int = 0):
        """Print response and exit with appropriate code.

        Args:
            response: Response to print
            exit_code: Exit code (0=success, 1=error, 2=critical, 3=config)
        """
        print(json.dumps(response, ensure_ascii=False, indent=2))
        sys.exit(exit_code)
