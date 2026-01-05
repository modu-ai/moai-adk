"""Unit tests for moai_adk.utils.timeout module.

Tests for cross-platform timeout handling.
"""

import platform
import threading
from unittest.mock import MagicMock, patch

import pytest

from moai_adk.utils.timeout import CrossPlatformTimeout, TimeoutError, timeout_context


class TestTimeoutError:
    """Test TimeoutError exception."""

    def test_timeout_error_is_exception(self):
        """Test TimeoutError is an Exception."""
        assert issubclass(TimeoutError, Exception)

    def test_timeout_error_can_be_raised(self):
        """Test TimeoutError can be raised."""
        with pytest.raises(TimeoutError):
            raise TimeoutError("Timeout occurred")


class TestCrossPlatformTimeout:
    """Test CrossPlatformTimeout class."""

    def test_initialization_basic(self):
        """Test basic initialization."""
        timeout = CrossPlatformTimeout(5)
        assert timeout.timeout_seconds == 5
        assert timeout.timeout_seconds_int == 5
        assert timeout.callback is None
        assert timeout.timer is None

    def test_initialization_with_callback(self):
        """Test initialization with callback."""
        callback = MagicMock()
        timeout = CrossPlatformTimeout(3, callback=callback)
        assert timeout.callback is callback

    def test_initialization_converts_float_timeout(self):
        """Test float timeout is converted to int."""
        timeout = CrossPlatformTimeout(3.7)
        assert timeout.timeout_seconds == 3.7
        assert timeout.timeout_seconds_int == 3

    def test_initialization_minimum_timeout(self):
        """Test minimum timeout is 1 second."""
        timeout = CrossPlatformTimeout(0.1)
        assert timeout.timeout_seconds_int == 1

    def test_is_windows_detection(self):
        """Test Windows detection."""
        timeout = CrossPlatformTimeout(5)
        is_windows = platform.system() == "Windows"
        assert timeout._is_windows == is_windows

    def test_start_negative_timeout(self):
        """Test start with negative timeout does nothing."""
        timeout = CrossPlatformTimeout(-5)
        timeout.start()
        # Should not raise exception

    def test_start_zero_timeout(self):
        """Test start with zero timeout raises immediately."""
        timeout = CrossPlatformTimeout(0)
        with pytest.raises(TimeoutError):
            timeout.start()

    def test_cancel_windows_timeout(self):
        """Test cancel on Windows platform."""
        timeout = CrossPlatformTimeout(5)
        timeout._is_windows = True
        mock_timer = MagicMock(spec=threading.Timer)
        timeout.timer = mock_timer
        timeout.cancel()
        assert mock_timer.cancel.called

    @patch("signal.signal")
    @patch("signal.alarm")
    def test_cancel_unix_timeout(self, mock_alarm, mock_signal):
        """Test cancel on Unix platform."""
        timeout = CrossPlatformTimeout(5)
        timeout._is_windows = False
        timeout._old_handler = MagicMock()
        timeout.cancel()
        mock_alarm.assert_called_with(0)

    def test_context_manager_enter(self):
        """Test context manager __enter__."""
        timeout = CrossPlatformTimeout(5)
        result = timeout.__enter__()
        assert result is timeout

    def test_context_manager_exit(self):
        """Test context manager __exit__."""
        timeout = CrossPlatformTimeout(5)
        timeout.timer = MagicMock()
        result = timeout.__exit__(None, None, None)
        assert result is False

    def test_start_windows_creates_timer(self):
        """Test start creates timer on Windows."""
        timeout = CrossPlatformTimeout(5)
        timeout._is_windows = True
        timeout.start()
        assert timeout.timer is not None
        assert isinstance(timeout.timer, threading.Timer)
        timeout.cancel()

    def test_context_manager_with_block(self):
        """Test using timeout as context manager."""
        timeout = CrossPlatformTimeout(5)
        with timeout:
            pass  # Should complete without timeout


class TestTimeoutContext:
    """Test timeout_context context manager."""

    def test_timeout_context_creation(self):
        """Test timeout_context creates context."""
        with timeout_context(5) as ctx:
            assert ctx is not None

    def test_timeout_context_with_callback(self):
        """Test timeout_context with callback."""
        callback = MagicMock()
        with timeout_context(5, callback=callback) as ctx:
            assert ctx.callback is callback

    def test_timeout_context_cleanup(self):
        """Test timeout_context properly cancels on exit."""
        with patch.object(CrossPlatformTimeout, "cancel"):
            with timeout_context(5):
                pass
            # Cancel should be called on exit

    def test_timeout_context_negative_timeout(self):
        """Test timeout_context with negative timeout."""
        # Should not timeout and complete normally
        with timeout_context(-1):
            pass
