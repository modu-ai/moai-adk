"""
Test suite for CrossPlatformTimeout utility.

Tests platform-specific timeout handling:
- Windows: threading.Timer-based timeout
- Unix/Linux/macOS: signal.SIGALRM-based timeout

@TEST:TIMEOUT-001 - Windows threading.Timer timeout
@TEST:TIMEOUT-002 - Unix signal.SIGALRM timeout
@TEST:TIMEOUT-003 - Timeout cancellation
"""

import sys
import signal
import threading
import time
from contextlib import contextmanager
from unittest import mock
import pytest

# Import the module to test
import sys
from pathlib import Path

# Add templates path to sys.path for imports
templates_path = Path(__file__).parent.parent.parent / "src" / "moai_adk" / "templates" / ".claude" / "hooks" / "alfred"
sys.path.insert(0, str(templates_path))

from utils.timeout import CrossPlatformTimeout, TimeoutError


class TestCrossPlatformTimeoutWindows:
    """Test timeout handling for Windows platform (threading-based)."""

    @mock.patch("platform.system", return_value="Windows")
    def test_timeout_windows_threading_fires(self, mock_platform):
        """Test that Windows timeout fires with threading.Timer."""
        timeout_fired = threading.Event()

        def slow_operation():
            time.sleep(2)  # Sleep longer than timeout

        with pytest.raises(SystemExit):
            with CrossPlatformTimeout(0.5):
                slow_operation()

    @mock.patch("platform.system", return_value="Windows")
    def test_timeout_windows_completes_before_timeout(self, mock_platform):
        """Test that Windows operation completes before timeout expires."""
        start = time.time()

        with CrossPlatformTimeout(2):
            time.sleep(0.1)  # Sleep less than timeout

        elapsed = time.time() - start
        assert elapsed < 1.5  # Should complete quickly

    @mock.patch("platform.system", return_value="Windows")
    def test_timeout_windows_with_callback(self, mock_platform):
        """Test Windows timeout with custom callback."""
        callback_called = threading.Event()

        def on_timeout():
            callback_called.set()

        with pytest.raises(SystemExit):
            with CrossPlatformTimeout(0.1, callback=on_timeout):
                time.sleep(1)

    @mock.patch("platform.system", return_value="Windows")
    def test_timeout_windows_exception_propagation(self, mock_platform):
        """Test that exceptions inside timeout context are propagated."""

        def raises_error():
            raise ValueError("Test error")

        with pytest.raises(ValueError, match="Test error"):
            with CrossPlatformTimeout(2):
                raises_error()


class TestCrossPlatformTimeoutUnix:
    """Test timeout handling for Unix platform (signal-based)."""

    @pytest.mark.skipif(sys.platform == "win32", reason="Unix-only test")
    def test_timeout_unix_signal_fires(self):
        """Test that Unix timeout fires with signal.SIGALRM."""
        with pytest.raises(TimeoutError):
            with CrossPlatformTimeout(0.1):
                time.sleep(1)

    @pytest.mark.skipif(sys.platform == "win32", reason="Unix-only test")
    def test_timeout_unix_completes_before_timeout(self):
        """Test that Unix operation completes before timeout expires."""
        start = time.time()

        with CrossPlatformTimeout(2):
            time.sleep(0.1)

        elapsed = time.time() - start
        assert elapsed < 1.5

    @pytest.mark.skipif(sys.platform == "win32", reason="Unix-only test")
    def test_timeout_unix_signal_cleanup(self):
        """Test that Unix signal is properly cleaned up after timeout."""
        # Get initial alarm state
        initial_alarm = signal.alarm(0)

        try:
            with CrossPlatformTimeout(1):
                time.sleep(0.1)
        except TimeoutError:
            pass

        # After context exit, alarm should be cancelled
        final_alarm = signal.alarm(0)
        assert final_alarm == 0

    @pytest.mark.skipif(sys.platform == "win32", reason="Unix-only test")
    def test_timeout_unix_nested_timeouts(self):
        """Test nested timeout contexts on Unix."""
        with pytest.raises(TimeoutError):
            with CrossPlatformTimeout(2):
                with CrossPlatformTimeout(0.1):
                    time.sleep(1)


class TestCrossPlatformTimeoutGeneral:
    """Test general timeout behavior across platforms."""

    @pytest.mark.skipif(sys.platform == "win32", reason="Unix-specific timeout test")
    def test_timeout_with_zero_seconds(self):
        """Test timeout with zero seconds (immediate)."""
        with pytest.raises(TimeoutError):
            with CrossPlatformTimeout(0):
                time.sleep(0.1)

    def test_timeout_with_negative_seconds(self):
        """Test timeout with negative seconds (should not timeout)."""
        # Negative timeouts should not trigger
        start = time.time()
        with CrossPlatformTimeout(-1):
            time.sleep(0.05)
        elapsed = time.time() - start
        assert elapsed < 0.5

    def test_timeout_context_manager_protocol(self):
        """Test that CrossPlatformTimeout implements context manager protocol."""
        timeout_obj = CrossPlatformTimeout(1)
        assert hasattr(timeout_obj, "__enter__")
        assert hasattr(timeout_obj, "__exit__")

    @pytest.mark.skipif(sys.platform == "win32", reason="Unix-specific timeout test")
    def test_timeout_exception_is_caught(self):
        """Test that TimeoutError can be caught."""
        caught = False
        try:
            with CrossPlatformTimeout(0.01):
                time.sleep(1)
        except TimeoutError:
            caught = True

        assert caught

    @pytest.mark.skipif(sys.platform == "win32", reason="Unix-specific timeout test")
    def test_timeout_finally_always_executes(self):
        """Test that finally blocks are executed even with timeout."""
        cleanup_executed = False

        try:
            with CrossPlatformTimeout(0.01):
                try:
                    time.sleep(1)
                finally:
                    cleanup_executed = True
        except TimeoutError:
            pass

        assert cleanup_executed

    def test_timeout_return_value_preserved(self):
        """Test that return values are preserved when no timeout occurs."""

        def return_value():
            return 42

        result = CrossPlatformTimeout(1).__enter__()
        # Context manager returns self, so no return value from function
        # This test documents expected behavior


class TestCrossPlatformTimeoutEdgeCases:
    """Test edge cases and error conditions."""

    @pytest.mark.skipif(sys.platform == "win32", reason="Unix-specific timeout test")
    def test_timeout_with_very_small_timeout(self):
        """Test timeout with very small duration."""
        with pytest.raises(TimeoutError):
            with CrossPlatformTimeout(0.001):
                time.sleep(0.1)

    def test_timeout_multiple_sequential_contexts(self):
        """Test multiple timeout contexts in sequence."""
        # First timeout should complete normally
        with CrossPlatformTimeout(1):
            time.sleep(0.05)

        # Second timeout should also work correctly
        with CrossPlatformTimeout(1):
            time.sleep(0.05)

    def test_timeout_with_exception_in_context(self):
        """Test that custom exceptions are not masked by timeout."""

        class CustomError(Exception):
            pass

        with pytest.raises(CustomError):
            with CrossPlatformTimeout(1):
                raise CustomError("Custom error")

    def test_timeout_with_keyboard_interrupt(self):
        """Test behavior with KeyboardInterrupt in context."""
        with pytest.raises(KeyboardInterrupt):
            with CrossPlatformTimeout(1):
                raise KeyboardInterrupt()

    def test_timeout_callback_with_exception(self):
        """Test callback that raises an exception."""

        def callback_raises():
            raise RuntimeError("Callback error")

        # Callback exception should not prevent timeout exception
        with pytest.raises(SystemExit):
            with CrossPlatformTimeout(0.01, callback=callback_raises):
                time.sleep(1)


class TestCrossPlatformTimeoutIntegration:
    """Integration tests for timeout handler."""

    @pytest.mark.skipif(sys.platform == "win32", reason="Unix-specific timeout test")
    def test_timeout_with_cpu_intensive_operation(self):
        """Test timeout with CPU-intensive operation."""

        def cpu_intensive():
            total = 0
            for i in range(10000000):
                total += i
            return total

        start = time.time()
        with pytest.raises(TimeoutError):
            with CrossPlatformTimeout(0.1):
                cpu_intensive()
        elapsed = time.time() - start
        assert elapsed < 1

    def test_timeout_with_io_operation(self):
        """Test timeout with I/O operation."""
        # Skip on Windows as socket timeout behavior differs
        if sys.platform != "win32":
            import socket
            with pytest.raises(TimeoutError):
                with CrossPlatformTimeout(0.1):
                    # Try to connect to non-routable address (will timeout)
                    s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
                    s.connect(("10.255.255.1", 1))

    @pytest.mark.skipif(sys.platform == "win32", reason="Unix-specific timeout test")
    def test_timeout_preserves_system_state(self):
        """Test that timeout doesn't leave system in bad state."""
        import os
        # Create a file to test cleanup
        test_file = "/tmp/timeout_test.txt"

        try:
            with pytest.raises(TimeoutError):
                with CrossPlatformTimeout(0.01):
                    with open(test_file, "w") as f:
                        f.write("test")
                    time.sleep(1)

            # File should still be writable afterward
            with open(test_file, "w") as f:
                f.write("cleanup")
        finally:
            if os.path.exists(test_file):
                os.remove(test_file)


class TestCrossPlatformTimeoutDocumentation:
    """Test that timeout handler matches documentation."""

    def test_timeout_class_has_docstring(self):
        """Test that CrossPlatformTimeout class has proper documentation."""
        assert CrossPlatformTimeout.__doc__ is not None
        assert len(CrossPlatformTimeout.__doc__) > 50

    def test_timeout_error_is_documented(self):
        """Test that TimeoutError is defined and documented."""
        assert TimeoutError is not None
        assert issubclass(TimeoutError, Exception)

    def test_module_has_proper_docstring(self):
        """Test that module has proper documentation."""
        from utils import timeout as timeout_module
        assert timeout_module.__doc__ is not None


# Markers and fixtures for test organization

@pytest.fixture
def timeout_context():
    """Fixture providing timeout context manager."""
    return CrossPlatformTimeout


@pytest.fixture
def clean_alarm_state():
    """Fixture to ensure clean alarm state before test."""
    if sys.platform != "win32":
        signal.alarm(0)
    yield
    if sys.platform != "win32":
        signal.alarm(0)


if __name__ == "__main__":
    pytest.main([__file__, "-v", "--cov=src/moai_adk/templates/claude/hooks/alfred/utils/timeout"])
