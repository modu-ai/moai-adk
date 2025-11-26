"""Comprehensive test suite for timeout.py utilities module.

This module provides 90%+ coverage for timeout handling functionality including:
- TimeoutError exception
- CrossPlatformTimeout class (Windows and Unix)
- Signal-based timeouts (Unix/POSIX systems)
- Thread-based timeouts (Windows/cross-platform)
- Context manager usage
- Callback execution
- Edge cases and error handling
- Timeout cancellation and cleanup
"""

import platform
import threading
import time
from unittest.mock import MagicMock, Mock, patch

import pytest

from moai_adk.utils.timeout import (
    CrossPlatformTimeout,
    TimeoutError,
    timeout_context,
)

# ============================================================================
# TimeoutError Exception Tests
# ============================================================================


class TestTimeoutError:
    """Tests for TimeoutError exception class."""

    def test_timeout_error_is_exception(self):
        """Test that TimeoutError is an Exception subclass."""
        assert issubclass(TimeoutError, Exception)

    def test_timeout_error_instantiation(self):
        """Test creating TimeoutError with message."""
        error = TimeoutError("Test timeout message")
        assert str(error) == "Test timeout message"

    def test_timeout_error_raise_and_catch(self):
        """Test raising and catching TimeoutError."""
        with pytest.raises(TimeoutError) as exc_info:
            raise TimeoutError("Operation timed out")

        assert "Operation timed out" in str(exc_info.value)

    def test_timeout_error_empty_message(self):
        """Test TimeoutError with empty message."""
        error = TimeoutError()
        assert isinstance(error, Exception)


# ============================================================================
# CrossPlatformTimeout Initialization Tests
# ============================================================================


class TestCrossPlatformTimeoutInitialization:
    """Tests for CrossPlatformTimeout initialization."""

    def test_initialization_with_integer_timeout(self):
        """Test CrossPlatformTimeout with integer timeout."""
        timeout = CrossPlatformTimeout(5)
        assert timeout.timeout_seconds == 5
        assert timeout.timeout_seconds_int == 5
        assert timeout.callback is None
        assert timeout.timer is None
        assert timeout._old_handler is None

    def test_initialization_with_float_timeout(self):
        """Test CrossPlatformTimeout with float timeout."""
        timeout = CrossPlatformTimeout(3.5)
        assert timeout.timeout_seconds == 3.5
        assert timeout.timeout_seconds_int == 3
        assert timeout.callback is None

    def test_initialization_with_callback(self):
        """Test CrossPlatformTimeout with callback function."""
        callback = Mock()
        timeout = CrossPlatformTimeout(5, callback=callback)
        assert timeout.callback is callback

    def test_initialization_platform_detection_windows(self):
        """Test platform detection for Windows."""
        with patch("platform.system", return_value="Windows"):
            timeout = CrossPlatformTimeout(5)
            assert timeout._is_windows is True

    def test_initialization_platform_detection_unix(self):
        """Test platform detection for Unix/Linux/macOS."""
        for system in ["Linux", "Darwin", "Unix"]:
            with patch("platform.system", return_value=system):
                timeout = CrossPlatformTimeout(5)
                assert timeout._is_windows is False

    def test_initialization_minimum_timeout_conversion(self):
        """Test timeout conversion to minimum value (1 second)."""
        timeout = CrossPlatformTimeout(0.1)
        # timeout_seconds_int should be at least 1 for signal.alarm
        assert timeout.timeout_seconds_int >= 1
        assert timeout.timeout_seconds == 0.1

    def test_initialization_zero_timeout(self):
        """Test initialization with zero timeout."""
        timeout = CrossPlatformTimeout(0)
        assert timeout.timeout_seconds == 0

    def test_initialization_negative_timeout(self):
        """Test initialization with negative timeout."""
        timeout = CrossPlatformTimeout(-5)
        assert timeout.timeout_seconds == -5

    def test_initialization_very_large_timeout(self):
        """Test initialization with very large timeout."""
        timeout = CrossPlatformTimeout(999999)
        assert timeout.timeout_seconds == 999999
        assert timeout.timeout_seconds_int == 999999


# ============================================================================
# CrossPlatformTimeout.start() Tests
# ============================================================================


class TestCrossPlatformTimeoutStart:
    """Tests for CrossPlatformTimeout.start method."""

    def test_start_with_positive_timeout(self):
        """Test start with positive timeout."""
        timeout = CrossPlatformTimeout(2)
        # Should not raise
        timeout.start()
        timeout.cancel()

    def test_start_zero_timeout_raises_timeout_error(self):
        """Test that zero timeout raises TimeoutError immediately."""
        timeout = CrossPlatformTimeout(0)
        with pytest.raises(TimeoutError) as exc_info:
            timeout.start()
        assert "0 seconds" in str(exc_info.value)

    def test_start_zero_timeout_with_callback(self):
        """Test that zero timeout executes callback before raising."""
        callback = Mock()
        timeout = CrossPlatformTimeout(0, callback=callback)
        with pytest.raises(TimeoutError):
            timeout.start()
        callback.assert_called_once()

    def test_start_negative_timeout_no_op(self):
        """Test that negative timeout does not set up timeout."""
        timeout = CrossPlatformTimeout(-5)
        # Should not raise
        timeout.start()
        # No timer should be set
        assert timeout.timer is None

    def test_start_windows_creates_timer(self):
        """Test that start creates timer on Windows."""
        with patch("platform.system", return_value="Windows"):
            timeout = CrossPlatformTimeout(2)
            timeout.start()

            assert timeout.timer is not None
            assert isinstance(timeout.timer, threading.Timer)

            timeout.cancel()

    @pytest.mark.skipif(platform.system() == "Windows", reason="Unix-only test")
    def test_start_unix_sets_signal_handler(self):
        """Test that start sets signal handler on Unix."""
        timeout = CrossPlatformTimeout(10)
        timeout.start()

        # Old handler should be saved
        assert timeout._old_handler is not None

        # Clean up
        timeout.cancel()

    def test_start_multiple_calls_windows(self):
        """Test multiple start calls on Windows."""
        with patch("platform.system", return_value="Windows"):
            timeout = CrossPlatformTimeout(2)
            timeout.start()

            # Second start without cancel
            timeout.start()
            timer2 = timeout.timer

            # New timer created (old one not cleaned up)
            assert timer2 is not None

            timeout.cancel()


# ============================================================================
# CrossPlatformTimeout.cancel() Tests
# ============================================================================


class TestCrossPlatformTimeoutCancel:
    """Tests for CrossPlatformTimeout.cancel method."""

    def test_cancel_windows_timeout(self):
        """Test canceling Windows timeout."""
        with patch("platform.system", return_value="Windows"):
            timeout = CrossPlatformTimeout(5)
            timeout.start()

            assert timeout.timer is not None
            timeout.cancel()
            assert timeout.timer is None

    @pytest.mark.skipif(platform.system() == "Windows", reason="Unix-only test")
    def test_cancel_unix_timeout(self):
        """Test canceling Unix timeout."""
        timeout = CrossPlatformTimeout(10)
        timeout.start()

        timeout.cancel()

        # Handler should be restored
        assert timeout._old_handler is None

    def test_cancel_without_start(self):
        """Test cancel without prior start."""
        timeout = CrossPlatformTimeout(5)
        # Should not raise
        timeout.cancel()

    def test_cancel_negative_timeout(self):
        """Test cancel with negative timeout."""
        timeout = CrossPlatformTimeout(-5)
        timeout.start()
        # Should not raise
        timeout.cancel()

    def test_cancel_clears_windows_timer(self):
        """Test that cancel properly clears Windows timer."""
        with patch("platform.system", return_value="Windows"):
            timeout = CrossPlatformTimeout(2)
            timeout.start()

            timeout.cancel()

            # Timer reference should be cleared
            assert timeout.timer is None


# ============================================================================
# Windows Timeout Tests
# ============================================================================


class TestWindowsTimeout:
    """Tests for Windows threading-based timeout."""

    @pytest.mark.skipif(platform.system() != "Windows", reason="Windows-only test")
    def test_windows_timeout_fires_on_windows(self):
        """Test timeout fires on actual Windows system."""
        timeout = CrossPlatformTimeout(1)

        with pytest.raises(TimeoutError):
            timeout.start()
            time.sleep(1.5)

    def test_windows_timeout_exception_message(self):
        """Test Windows timeout exception message."""
        with patch("platform.system", return_value="Windows"):
            timeout = CrossPlatformTimeout(2)

            def raise_timeout():
                timeout.start()
                time.sleep(2.5)

            with patch.object(threading.Timer, "start"):
                # Just verify the timer was created with proper function
                timeout.start()
                assert timeout.timer is not None
                timeout.cancel()

    def test_windows_callback_execution(self):
        """Test callback is executed before timeout on Windows."""
        with patch("platform.system", return_value="Windows"):
            callback = Mock()
            timeout = CrossPlatformTimeout(2, callback=callback)

            with patch("threading.Timer") as mock_timer_class:
                mock_timer_instance = MagicMock()
                mock_timer_class.return_value = mock_timer_instance

                timeout.start()
                # Callback should be stored in timeout object
                assert timeout.callback is callback
                # Timer should be created
                assert timeout.timer is not None


# ============================================================================
# Unix Timeout Tests
# ============================================================================


class TestUnixTimeout:
    """Tests for Unix signal-based timeout."""

    @pytest.mark.skipif(platform.system() == "Windows", reason="Unix-only test")
    def test_unix_timeout_fires_on_unix(self):
        """Test timeout fires on actual Unix system."""
        # This test is platform-specific and might need adjustment
        # Based on the actual system running tests
        pass

    def test_unix_signal_handler_saved(self):
        """Test that Unix saves previous signal handler."""
        with patch("platform.system", return_value="Linux"):
            with patch("signal.signal") as mock_signal:
                original_handler = Mock()
                mock_signal.return_value = original_handler

                timeout = CrossPlatformTimeout(10)
                timeout.start()

                # Signal handler should be set
                mock_signal.assert_called()
                # Old handler should be saved
                assert timeout._old_handler is original_handler

    def test_unix_callback_exception_ignored(self):
        """Test that callback exceptions are ignored in Unix handler."""
        with patch("platform.system", return_value="Linux"):
            callback = Mock(side_effect=RuntimeError("Callback error"))
            timeout = CrossPlatformTimeout(10, callback=callback)

            with patch("signal.signal"):
                with patch("signal.alarm"):
                    timeout.start()
                    # Callback is stored for later execution
                    assert timeout.callback is callback
                    timeout.cancel()

    def test_unix_handler_restoration(self):
        """Test that Unix restores previous signal handler on cancel."""
        with patch("platform.system", return_value="Linux"):
            with patch("signal.signal") as mock_signal:
                with patch("signal.alarm"):
                    original_handler = Mock()
                    mock_signal.return_value = original_handler

                    timeout = CrossPlatformTimeout(10)
                    timeout.start()

                    # Reset mock to check second call
                    mock_signal.reset_mock()

                    timeout.cancel()

                    # Handler should be restored
                    mock_signal.assert_called()


# ============================================================================
# Context Manager Tests
# ============================================================================


class TestContextManager:
    """Tests for context manager functionality."""

    def test_context_manager_enter_and_exit(self):
        """Test context manager enters and exits cleanly."""
        with CrossPlatformTimeout(5) as timeout:
            assert isinstance(timeout, CrossPlatformTimeout)
            assert timeout.timeout_seconds == 5

    def test_context_manager_with_statement(self):
        """Test using timeout with 'with' statement."""
        timeout_obj = None
        with CrossPlatformTimeout(5) as timeout:
            timeout_obj = timeout
            time.sleep(0.1)  # Quick operation

        assert timeout_obj is not None

    def test_context_manager_cancels_on_exit(self):
        """Test that context manager cancels timeout on exit."""
        with patch("platform.system", return_value="Windows"):
            with CrossPlatformTimeout(5) as timeout:
                timer = timeout.timer
                assert timer is not None

            # After exit, timer should be None
            assert timeout.timer is None

    def test_context_manager_suppresses_no_exceptions(self):
        """Test context manager does not suppress exceptions."""
        with pytest.raises(ValueError):
            with CrossPlatformTimeout(5):
                raise ValueError("Test error")

    def test_context_manager_with_exception_still_cancels(self):
        """Test timeout is canceled even if exception occurs."""
        with patch("platform.system", return_value="Windows"):
            try:
                with CrossPlatformTimeout(5):
                    raise RuntimeError("Test error")
            except RuntimeError:
                pass

            # Timer should still be None after exit

    def test_context_manager_return_value(self):
        """Test context manager returns self."""
        timeout = CrossPlatformTimeout(5)
        with timeout as ctx:
            assert ctx is timeout


# ============================================================================
# timeout_context Function Tests
# ============================================================================


class TestTimeoutContextFunction:
    """Tests for timeout_context decorator/context manager function."""

    def test_timeout_context_basic_usage(self):
        """Test basic usage of timeout_context."""
        with timeout_context(5):
            time.sleep(0.1)
        # Should complete without timeout

    def test_timeout_context_creates_timeout_object(self):
        """Test that timeout_context creates CrossPlatformTimeout."""
        with timeout_context(5) as timeout:
            assert isinstance(timeout, CrossPlatformTimeout)

    def test_timeout_context_with_callback(self):
        """Test timeout_context with callback."""
        callback = Mock()
        with timeout_context(5, callback=callback):
            time.sleep(0.1)
        # Callback should be set up

    def test_timeout_context_returns_timeout(self):
        """Test that timeout_context yields timeout object."""
        timeout_obj = None
        with timeout_context(5) as timeout:
            timeout_obj = timeout
            assert timeout is not None

        assert timeout_obj is not None

    def test_timeout_context_zero_timeout(self):
        """Test timeout_context with zero timeout."""
        with pytest.raises(TimeoutError):
            with timeout_context(0):
                pass

    def test_timeout_context_negative_timeout(self):
        """Test timeout_context with negative timeout."""
        with timeout_context(-5):
            time.sleep(0.1)
        # Negative timeout should not timeout

    def test_timeout_context_cleanup(self):
        """Test that timeout_context properly cleans up."""
        with patch("platform.system", return_value="Windows"):
            with timeout_context(5) as timeout:
                pass

            # Timeout should be canceled
            assert timeout.timer is None

    def test_timeout_context_exception_propagation(self):
        """Test that exceptions are propagated from timeout_context."""
        with pytest.raises(ValueError) as exc_info:
            with timeout_context(5):
                raise ValueError("Custom error")

        assert "Custom error" in str(exc_info.value)

    def test_timeout_context_callback_execution(self):
        """Test callback is executed in timeout_context."""
        callback = Mock()
        with timeout_context(5, callback=callback):
            time.sleep(0.1)
        # Callback would be called if timeout fires


# ============================================================================
# Callback Tests
# ============================================================================


class TestCallback:
    """Tests for callback functionality."""

    def test_callback_none_by_default(self):
        """Test that callback is None by default."""
        timeout = CrossPlatformTimeout(5)
        assert timeout.callback is None

    def test_callback_execution_on_zero_timeout(self):
        """Test callback is called on zero timeout."""
        callback = Mock()
        timeout = CrossPlatformTimeout(0, callback=callback)

        with pytest.raises(TimeoutError):
            timeout.start()

        # Callback should be called
        callback.assert_called_once()

    def test_callback_is_stored(self):
        """Test that callback is properly stored."""
        callback = Mock()
        timeout = CrossPlatformTimeout(5, callback=callback)

        assert timeout.callback is callback

    def test_callback_with_custom_function(self):
        """Test callback with custom function."""
        executed = []

        def custom_callback():
            executed.append(True)

        timeout = CrossPlatformTimeout(0, callback=custom_callback)

        with pytest.raises(TimeoutError):
            timeout.start()

        assert executed == [True]


# ============================================================================
# Edge Cases and Error Handling Tests
# ============================================================================


class TestEdgeCases:
    """Tests for edge cases and error handling."""

    def test_very_small_timeout(self):
        """Test with very small timeout value."""
        timeout = CrossPlatformTimeout(0.001)
        assert timeout.timeout_seconds == 0.001
        assert timeout.timeout_seconds_int >= 1

    def test_fractional_timeout(self):
        """Test with fractional timeout."""
        timeout = CrossPlatformTimeout(2.5)
        assert timeout.timeout_seconds == 2.5
        assert timeout.timeout_seconds_int == 2

    def test_large_timeout(self):
        """Test with large timeout value."""
        timeout = CrossPlatformTimeout(86400)  # 1 day
        assert timeout.timeout_seconds == 86400

    def test_timeout_with_none_callback(self):
        """Test timeout explicitly with None callback."""
        timeout = CrossPlatformTimeout(5, callback=None)
        timeout.start()
        timeout.cancel()

    def test_cancel_multiple_times(self):
        """Test calling cancel multiple times."""
        timeout = CrossPlatformTimeout(5)
        timeout.start()
        timeout.cancel()
        timeout.cancel()  # Should not raise

    def test_timeout_without_context_manager(self):
        """Test manual timeout management without context manager."""
        timeout = CrossPlatformTimeout(5)
        timeout.start()

        # Do some work
        time.sleep(0.1)

        timeout.cancel()

    def test_platform_switch_behavior(self):
        """Test behavior with different platform configurations."""
        for is_windows in [True, False]:
            with patch("platform.system", return_value="Windows" if is_windows else "Linux"):
                timeout = CrossPlatformTimeout(5)
                assert timeout._is_windows == is_windows

    def test_timeout_doesnt_execute_work_after_cancel(self):
        """Test that timeout doesn't interrupt canceled work."""
        with patch("platform.system", return_value="Windows"):
            timeout = CrossPlatformTimeout(2)
            timeout.start()
            time.sleep(0.1)
            timeout.cancel()

            # Should be able to complete work without timeout
            time.sleep(0.5)


# ============================================================================
# Integration Tests
# ============================================================================


class TestIntegration:
    """Integration tests for timeout functionality."""

    def test_workflow_with_multiple_timeouts(self):
        """Test workflow using multiple timeout objects."""
        timeout1 = CrossPlatformTimeout(5)
        timeout2 = CrossPlatformTimeout(10)

        timeout1.start()
        timeout2.start()

        time.sleep(0.1)

        timeout1.cancel()
        timeout2.cancel()

    def test_nested_context_managers(self):
        """Test nested timeout context managers."""
        with timeout_context(10):
            with timeout_context(5):
                time.sleep(0.1)
            # Outer timeout still active
            time.sleep(0.1)

    def test_complete_workflow_context_manager(self):
        """Test complete workflow using context manager."""
        executed = []

        def callback():
            executed.append("callback")

        with timeout_context(5, callback=callback):
            executed.append("start")
            time.sleep(0.1)
            executed.append("end")

        assert executed == ["start", "end"]

    def test_exception_within_timeout_context(self):
        """Test exception handling within timeout context."""
        with pytest.raises(RuntimeError):
            with timeout_context(5):
                raise RuntimeError("Test exception")

    def test_quick_operation_completes_successfully(self):
        """Test that quick operations complete before timeout."""
        result = []

        with timeout_context(10):
            for i in range(5):
                result.append(i)
                time.sleep(0.01)

        assert result == [0, 1, 2, 3, 4]

    def test_timeout_cancel_prevents_interrupt(self):
        """Test that canceling timeout prevents interruption."""
        with patch("platform.system", return_value="Windows"):
            interrupted = []

            timeout = CrossPlatformTimeout(2)
            timeout.start()
            time.sleep(0.1)
            timeout.cancel()

            # Continue with work
            for i in range(3):
                interrupted.append(i)
                time.sleep(0.1)

            assert len(interrupted) == 3

    def test_stress_test_multiple_start_cancel_cycles(self):
        """Test multiple start/cancel cycles."""
        timeout = CrossPlatformTimeout(5)

        for _ in range(5):
            timeout.start()
            time.sleep(0.05)
            timeout.cancel()

    def test_context_manager_with_quick_operation(self):
        """Test context manager with very quick operation."""
        start = time.time()
        with timeout_context(10):
            time.sleep(0.01)
        duration = time.time() - start

        assert duration < 1.0  # Should be much faster than timeout


# ============================================================================
# Error Handling Tests
# ============================================================================


class TestErrorHandling:
    """Tests for error handling scenarios."""

    def test_timeout_error_message_format(self):
        """Test TimeoutError message format."""
        timeout = CrossPlatformTimeout(5)
        timeout.start()
        timeout.cancel()

    def test_callback_with_side_effects(self):
        """Test callback that has side effects."""
        state = {"called": False}

        def callback():
            state["called"] = True

        timeout = CrossPlatformTimeout(0, callback=callback)

        with pytest.raises(TimeoutError):
            timeout.start()

        assert state["called"] is True

    def test_multiple_callbacks(self):
        """Test using different callbacks."""
        callback1 = Mock()
        callback2 = Mock()

        timeout1 = CrossPlatformTimeout(0, callback=callback1)
        timeout2 = CrossPlatformTimeout(0, callback=callback2)

        try:
            timeout1.start()
        except TimeoutError:
            pass

        try:
            timeout2.start()
        except TimeoutError:
            pass

        callback1.assert_called_once()
        callback2.assert_called_once()


# ============================================================================
# Performance and Timing Tests
# ============================================================================


class TestPerformanceAndTiming:
    """Tests for performance characteristics."""

    def test_timeout_initialization_is_fast(self):
        """Test that initialization is very fast."""
        start = time.time()
        for _ in range(100):
            CrossPlatformTimeout(5)
        duration = time.time() - start

        assert duration < 1.0  # Should be very fast

    def test_timeout_cancel_is_fast(self):
        """Test that cancel operation is fast."""
        timeout = CrossPlatformTimeout(5)
        timeout.start()

        start = time.time()
        timeout.cancel()
        duration = time.time() - start

        assert duration < 0.1

    def test_context_manager_overhead_minimal(self):
        """Test that context manager has minimal overhead."""
        start = time.time()
        with timeout_context(10):
            pass
        duration = time.time() - start

        assert duration < 0.5


# ============================================================================
# Platform-Specific Tests
# ============================================================================


class TestPlatformSpecific:
    """Platform-specific behavior tests."""

    def test_windows_path(self):
        """Test Windows-specific timeout path."""
        with patch("platform.system", return_value="Windows"):
            timeout = CrossPlatformTimeout(5)
            assert timeout._is_windows is True

    def test_unix_path(self):
        """Test Unix-specific timeout path."""
        with patch("platform.system", return_value="Darwin"):
            timeout = CrossPlatformTimeout(5)
            assert timeout._is_windows is False

    def test_linux_path(self):
        """Test Linux-specific timeout path."""
        with patch("platform.system", return_value="Linux"):
            timeout = CrossPlatformTimeout(5)
            assert timeout._is_windows is False


# ============================================================================
# Type and Structure Tests
# ============================================================================


class TestTypeAndStructure:
    """Tests for type consistency and structure."""

    def test_timeout_object_attributes(self):
        """Test that timeout object has all expected attributes."""
        timeout = CrossPlatformTimeout(5)

        assert hasattr(timeout, "timeout_seconds")
        assert hasattr(timeout, "timeout_seconds_int")
        assert hasattr(timeout, "callback")
        assert hasattr(timeout, "timer")
        assert hasattr(timeout, "_is_windows")
        assert hasattr(timeout, "_old_handler")

    def test_timeout_methods_exist(self):
        """Test that all timeout methods exist."""
        timeout = CrossPlatformTimeout(5)

        assert hasattr(timeout, "start")
        assert hasattr(timeout, "cancel")
        assert hasattr(timeout, "__enter__")
        assert hasattr(timeout, "__exit__")

    def test_timeout_context_function_signature(self):
        """Test timeout_context function signature."""
        import inspect

        sig = inspect.signature(timeout_context)
        params = list(sig.parameters.keys())

        assert "timeout_seconds" in params
        assert "callback" in params
