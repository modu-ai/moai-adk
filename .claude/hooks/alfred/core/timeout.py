#!/usr/bin/env python3
"""Cross-Platform Timeout Handler for Windows & Unix Compatibility

Provides a unified timeout mechanism that works on both Windows (threading-based)
and Unix/POSIX systems (signal-based).

Architecture:
  - Windows: threading.Timer with exception injection
  - Unix/POSIX: signal.SIGALRM (traditional approach)
  - Both: Context manager for clean cancellation
"""

import platform
import signal
import threading
from contextlib import contextmanager
from typing import Optional


class TimeoutError(Exception):
    """Timeout exception raised when deadline exceeded"""
    pass


class CrossPlatformTimeout:
    """Cross-platform timeout handler supporting Windows and Unix.

    Windows: Uses threading.Timer to schedule timeout exception
    Unix: Uses signal.SIGALRM for timeout handling

    Usage:
        # Context manager (recommended)
        with CrossPlatformTimeout(5):
            long_running_operation()

        # Manual control
        timeout = CrossPlatformTimeout(5)
        timeout.start()
        try:
            long_running_operation()
        finally:
            timeout.cancel()
    """

    def __init__(self, timeout_seconds: int):
        """Initialize timeout with duration in seconds.

        Args:
            timeout_seconds: Timeout duration in seconds
        """
        self.timeout_seconds = timeout_seconds
        self.timer: Optional[threading.Timer] = None
        self._is_windows = platform.system() == "Windows"
        self._old_handler = None

    def start(self) -> None:
        """Start timeout countdown."""
        if self._is_windows:
            self._start_windows_timeout()
        else:
            self._start_unix_timeout()

    def cancel(self) -> None:
        """Cancel timeout (must call before timeout expires)."""
        if self._is_windows:
            self._cancel_windows_timeout()
        else:
            self._cancel_unix_timeout()

    def _start_windows_timeout(self) -> None:
        """Windows: Use threading.Timer to raise exception."""
        def timeout_handler():
            raise TimeoutError(
                f"Operation exceeded {self.timeout_seconds}s timeout (Windows threading)"
            )

        self.timer = threading.Timer(self.timeout_seconds, timeout_handler)
        self.timer.daemon = True
        self.timer.start()

    def _cancel_windows_timeout(self) -> None:
        """Windows: Cancel timer thread."""
        if self.timer:
            self.timer.cancel()
            self.timer = None

    def _start_unix_timeout(self) -> None:
        """Unix/POSIX: Use signal.SIGALRM for timeout."""
        def signal_handler(signum, frame):
            raise TimeoutError(
                f"Operation exceeded {self.timeout_seconds}s timeout (Unix signal)"
            )

        # Save old handler to restore later
        self._old_handler = signal.signal(signal.SIGALRM, signal_handler)
        signal.alarm(self.timeout_seconds)

    def _cancel_unix_timeout(self) -> None:
        """Unix/POSIX: Cancel alarm and restore old handler."""
        signal.alarm(0)  # Cancel pending alarm
        if self._old_handler is not None:
            signal.signal(signal.SIGALRM, self._old_handler)
            self._old_handler = None

    def __enter__(self):
        """Context manager entry."""
        self.start()
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        """Context manager exit - always cancel."""
        self.cancel()
        return False  # Don't suppress exceptions


@contextmanager
def timeout_context(timeout_seconds: int):
    """Decorator/context manager for timeout.

    Usage:
        with timeout_context(5):
            slow_function()

    Args:
        timeout_seconds: Timeout duration in seconds

    Yields:
        CrossPlatformTimeout instance
    """
    timeout = CrossPlatformTimeout(timeout_seconds)
    timeout.start()
    try:
        yield timeout
    finally:
        timeout.cancel()
