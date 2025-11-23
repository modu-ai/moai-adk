"""
Nano Banana Pro Error Handler Module

Handle Gemini API errors and manage retry strategies.

API Error Reference:
- RESOURCE_EXHAUSTED (429): Too many requests, backoff required
- SAFETY: Safety filter violation, prompt modification required
- RECITATION: Training data recitation risk, prompt modification required
- INVALID_ARGUMENT: Invalid input, no retry possible
- INTERNAL: Server error, retry possible

Usage:
    from modules.error_handler import ErrorHandler

    error_handler = ErrorHandler(error_response)
    if error_handler.is_retryable():
        wait_time = error_handler.get_retry_delay()
        time.sleep(wait_time)
        # retry
    else:
        print(error_handler.get_message())
"""

from typing import Optional, Dict, Any
import time


class ErrorHandler:
    """Gemini API Error Handler Class"""

    # Error type classification
    RETRYABLE_ERRORS = {
        "RESOURCE_EXHAUSTED",
        "INTERNAL",
        "UNAVAILABLE",
        "DEADLINE_EXCEEDED",
    }

    NON_RETRYABLE_ERRORS = {
        "INVALID_ARGUMENT",
        "NOT_FOUND",
        "PERMISSION_DENIED",
        "UNAUTHENTICATED",
        "FAILED_PRECONDITION",
        "ABORTED",
        "OUT_OF_RANGE",
        "UNIMPLEMENTED",
    }

    # Safety-related errors
    SAFETY_ERRORS = {
        "SAFETY",
        "BLOCKED_REASON",
    }

    # Data-related errors
    DATA_ERRORS = {
        "RECITATION",
        "COPYRIGHT",
    }

    # Retry policy
    RETRY_CONFIG = {
        "RESOURCE_EXHAUSTED": {
            "retryable": True,
            "max_retries": 5,
            "initial_delay": 60,  # 초
            "backoff_multiplier": 2.0,
        },
        "INTERNAL": {
            "retryable": True,
            "max_retries": 3,
            "initial_delay": 1,
            "backoff_multiplier": 2.0,
        },
        "UNAVAILABLE": {
            "retryable": True,
            "max_retries": 3,
            "initial_delay": 1,
            "backoff_multiplier": 2.0,
        },
        "DEADLINE_EXCEEDED": {
            "retryable": True,
            "max_retries": 2,
            "initial_delay": 5,
            "backoff_multiplier": 2.0,
        },
    }

    def __init__(self, error_response: Dict[str, Any]):
        """
        ErrorHandler 초기화

        Args:
            error_response: API 에러 응답
        """
        self.error_response = error_response
        self.error_code = self._extract_error_code()
        self.error_message = self._extract_error_message()
        self.error_details = self._extract_error_details()
        self.retry_count = 0

    def _extract_error_code(self) -> str:
        """Extract error code"""
        error = self.error_response.get("error", {})

        # 구조 1: {'error': {'code': ..., 'message': ...}}
        if isinstance(error, dict):
            return error.get("code", "UNKNOWN")

        # 구조 2: {'error': "CODE"}
        if isinstance(error, str):
            return error

        return "UNKNOWN"

    def _extract_error_message(self) -> str:
        """Extract error message"""
        error = self.error_response.get("error", {})

        if isinstance(error, dict):
            return error.get("message", "Unknown error occurred")

        return str(error)

    def _extract_error_details(self) -> Dict[str, Any]:
        """Extract error details"""
        error = self.error_response.get("error", {})

        if isinstance(error, dict):
            return error.get("details", {})

        return {}

    def is_retryable(self) -> bool:
        """
        Determine if the error is retryable.

        Returns:
            bool: Whether the error is retryable
        """
        if self.error_code in self.RETRYABLE_ERRORS:
            return True

        # Safety errors are not retryable (prompt modification required)
        if self.error_code in self.SAFETY_ERRORS:
            return False

        # Data-related errors are not retryable
        if self.error_code in self.DATA_ERRORS:
            return False

        # Explicit non-retryable errors
        if self.error_code in self.NON_RETRYABLE_ERRORS:
            return False

        # Other errors are not retryable by default
        return False

    def get_retry_delay(self) -> float:
        """
        Calculate retry wait time in seconds.

        Returns:
            float: Wait time in seconds
        """
        config = self.RETRY_CONFIG.get(
            self.error_code,
            {"initial_delay": 1, "backoff_multiplier": 2.0}
        )

        initial_delay = config["initial_delay"]
        multiplier = config["backoff_multiplier"]

        # Exponential backoff calculation
        delay = initial_delay * (multiplier ** self.retry_count)

        # Maximum wait time limit (30 minutes)
        max_delay = 1800
        return min(delay, max_delay)

    def get_max_retries(self) -> int:
        """
        Get maximum retry count.

        Returns:
            int: Maximum retry count
        """
        config = self.RETRY_CONFIG.get(self.error_code, {})
        return config.get("max_retries", 0)

    def get_message(self) -> str:
        """
        Get user-friendly error message.

        Returns:
            str: Formatted error message
        """
        error_type = self._classify_error()

        messages = {
            "retryable": f"A temporary error occurred. Retrying...\n"
                        f"[{self.error_code}] {self.error_message}",
            "safety": f"Image violates safety policy.\n"
                     f"Please modify the prompt and try again.\n"
                     f"Details: {self.error_message}",
            "recitation": f"Content similar to training data detected.\n"
                         f"Try with a different style or composition.\n"
                         f"Details: {self.error_message}",
            "invalid": f"Invalid request. Please check your input.\n"
                      f"[{self.error_code}] {self.error_message}",
            "unknown": f"An unknown error occurred.\n"
                      f"[{self.error_code}] {self.error_message}",
        }

        return messages.get(error_type, messages["unknown"])

    def get_resolution_action(self) -> str:
        """
        Get error resolution action.

        Returns:
            str: Recommended resolution action
        """
        error_type = self._classify_error()

        actions = {
            "retryable": "System will automatically retry.",
            "safety": "Remove unsafe elements from the prompt.",
            "recitation": "Try a different prompt or style.",
            "invalid": "Check input format and try again.",
            "unknown": "Try again later or contact support.",
        }

        return actions.get(error_type, actions["unknown"])

    def _classify_error(self) -> str:
        """Classify error"""
        if self.error_code in self.SAFETY_ERRORS:
            return "safety"
        elif self.error_code in self.DATA_ERRORS:
            return "recitation"
        elif self.error_code in self.RETRYABLE_ERRORS:
            return "retryable"
        elif self.error_code in self.NON_RETRYABLE_ERRORS:
            return "invalid"
        else:
            return "unknown"

    def get_error_code(self) -> str:
        """Return error code"""
        return self.error_code

    def get_error_details(self) -> Dict[str, Any]:
        """Return error details"""
        return {
            "code": self.error_code,
            "message": self.error_message,
            "details": self.error_details,
            "retryable": self.is_retryable(),
            "classification": self._classify_error(),
        }

    def increment_retry_count(self) -> None:
        """Increment retry count"""
        self.retry_count += 1


class FinishReasonHandler:
    """Gemini API Finish Reason Handler"""

    # Finish reasons
    FINISH_REASONS = {
        "STOP": "Normal completion",
        "MAX_TOKENS": "Maximum tokens reached",
        "SAFETY": "Safety filter triggered",
        "RECITATION": "Training data recitation",
        "OTHER": "Other reason",
    }

    @staticmethod
    def is_successful(finish_reason: str) -> bool:
        """
        Determine if the finish reason indicates success.

        Args:
            finish_reason: API response finish_reason

        Returns:
            bool: Whether the completion was successful
        """
        return finish_reason == "STOP"

    @staticmethod
    def get_description(finish_reason: str) -> str:
        """
        Get description of finish reason.

        Args:
            finish_reason: API response finish_reason

        Returns:
            str: Description message
        """
        return FinishReasonHandler.FINISH_REASONS.get(
            finish_reason,
            "Unknown completion reason"
        )

    @staticmethod
    def is_retryable(finish_reason: str) -> bool:
        """
        Determine if the completion can be retried based on finish reason.

        Args:
            finish_reason: API response finish_reason

        Returns:
            bool: Whether retry is possible
        """
        # SAFETY and RECITATION require prompt changes, not retry
        return finish_reason not in ("SAFETY", "RECITATION")


if __name__ == "__main__":
    # Test execution
    error_response = {
        "error": {
            "code": "RESOURCE_EXHAUSTED",
            "message": "429 Too Many Requests",
            "details": {}
        }
    }

    handler = ErrorHandler(error_response)
    print(f"Error Code: {handler.get_error_code()}")
    print(f"Is Retryable: {handler.is_retryable()}")
    print(f"Retry Delay: {handler.get_retry_delay()}s")
    print(f"Message: {handler.get_message()}")
    print(f"Resolution: {handler.get_resolution_action()}")
