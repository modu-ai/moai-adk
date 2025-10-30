"""
Cross-platform timeout handler for hooks.

@SPEC:BUGFIX-001
@CODE:BUGFIX-001
@CODE:TIMEOUT-001

This module provides a platform-agnostic timeout mechanism:
- Windows: Uses threading.Timer (no signal support)
- Unix/Linux/macOS: Uses signal.SIGALRM (POSIX standard)

Usage:
    from utils.timeout import CrossPlatformTimeout, TimeoutError

    # Method 1: Context manager (recommended)
    with CrossPlatformTimeout(5):
        # code that might timeout
        pass

    # Method 2: Manual control
    timeout = CrossPlatformTimeout(5)
    timeout.start()
    try:
        # code that might timeout
        pass
    finally:
        timeout.cancel()
"""

import platform
import signal
import sys
import threading
from typing import Callable, Optional


class TimeoutError(Exception):
    """Timeout exception raised when hook execution exceeds limit."""

    pass


class CrossPlatformTimeout:
    """
    Cross-platform timeout handler supporting Windows, Unix, Linux, macOS.

    Attributes:
        timeout_seconds: Timeout duration in seconds
        callback: Callback function to execute on timeout
        timer: Threading.Timer instance (Windows only)
        is_windows: Boolean indicating if running on Windows
    """

    def __init__(self, timeout_seconds: int, callback: Optional[Callable] = None):
        """
        Initialize timeout handler.

        Args:
            timeout_seconds: Timeout duration in seconds
            callback: Optional callback to execute on timeout (default: sys.exit(1))
        """
        self.timeout_seconds = timeout_seconds
        self.callback = callback or self._default_timeout_callback
        self.timer: Optional[threading.Timer] = None
        self.is_windows = platform.system() == "Windows"

    def _default_timeout_callback(self):
        """Default timeout callback: print error and exit."""
        print(f"Hook timeout after {self.timeout_seconds} seconds", file=sys.stderr)
        sys.exit(1)

    def _unix_signal_handler(self, signum, frame):
        """Signal handler for Unix platforms."""
        raise TimeoutError(f"Hook execution exceeded {self.timeout_seconds}-second timeout")

    def start(self):
        """Start timeout monitoring."""
        if self.is_windows:
            # Windows: use threading.Timer
            self.timer = threading.Timer(self.timeout_seconds, self.callback)
            self.timer.daemon = True
            self.timer.start()
        else:
            # Unix/Linux/macOS: use signal.SIGALRM
            # Note: signal.alarm() requires integer seconds
            signal.signal(signal.SIGALRM, self._unix_signal_handler)
            signal.alarm(max(1, int(self.timeout_seconds)))

    def cancel(self):
        """Cancel timeout monitoring."""
        if self.is_windows:
            if self.timer:
                self.timer.cancel()
        else:
            signal.alarm(0)

    def __enter__(self):
        """Context manager entry."""
        self.start()
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        """Context manager exit."""
        self.cancel()
        return False
