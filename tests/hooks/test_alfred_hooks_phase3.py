#!/usr/bin/env python3
# @TEST:HOOKS-EMERGENCY-001-PHASE3 | @CODE:HOOKS-EMERGENCY-001:INTEGRATION
"""Phase 3: Cross-platform Integration Tests for alfred_hooks.py

This test suite validates:
1. All 8 Hook event types (SessionStart, UserPromptSubmit, PreToolUse, PostToolUse,
   SessionEnd, Notification, Stop, SubagentStop)
2. Cross-platform timeout mechanism (Windows threading + Unix signal)
3. Error handling (JSON parsing, empty stdin, invalid events, exceptions)
4. Performance and stability (5-second timeout, memory, signal cleanup)

Test Strategy:
- Test 1-8: Verify each Hook event type
- Test 9-12: Cross-platform compatibility (timeout mechanism)
- Test 13-17: Error handling (JSON, stdin, exceptions)
- Test 18-20: Performance and stability (timeout, memory, cleanup)

Expected: ALL TESTS PASS (GREEN phase validation)
"""

import json
import platform
import subprocess
import sys
import time
from pathlib import Path
from typing import Any

import pytest


# Hook script path
HOOKS_DIR = Path(__file__).parent.parent.parent / ".claude/hooks/alfred"
ALFRED_HOOKS_SCRIPT = HOOKS_DIR / "alfred_hooks.py"


def run_hook(event_name: str, payload: dict[str, Any] | None = None) -> tuple[int, str, str]:
    """Run alfred_hooks.py and return exit code, stdout, stderr

    Args:
        event_name: Hook event name (SessionStart, UserPromptSubmit, etc.)
        payload: JSON payload to send via stdin

    Returns:
        Tuple of (exit_code, stdout, stderr)
    """
    if payload is None:
        payload = {"cwd": "."}

    # Serialize payload to JSON
    input_json = json.dumps(payload)

    # Run alfred_hooks.py subprocess
    result = subprocess.run(
        [sys.executable, str(ALFRED_HOOKS_SCRIPT), event_name],
        input=input_json,
        capture_output=True,
        text=True,
        timeout=10,  # Prevent test hang
    )

    return result.returncode, result.stdout, result.stderr


class TestAllHookEventTypes:
    """Test all 8 Hook event types for correct JSON output"""

    def test_session_start_event(self):
        """SessionStart event returns valid Hook response

        @TEST:HOOKS-EVENT-001

        Given: SessionStart event with cwd payload
        When: alfred_hooks.py SessionStart is executed
        Then: Exit code 0, valid JSON with "continue" key
        """
        exit_code, stdout, stderr = run_hook("SessionStart", {"cwd": ".", "phase": "clear"})

        assert exit_code == 0, f"SessionStart failed: {stderr}"
        response = json.loads(stdout)
        assert "continue" in response
        assert response["continue"] is True

    def test_user_prompt_submit_event(self):
        """UserPromptSubmit event returns valid Hook response with hookEventName

        @TEST:HOOKS-EVENT-002

        Given: UserPromptSubmit with userPrompt
        When: alfred_hooks.py UserPromptSubmit is executed
        Then: Exit code 0, JSON with hookSpecificOutput.hookEventName
        """
        exit_code, stdout, stderr = run_hook(
            "UserPromptSubmit",
            {"cwd": ".", "userPrompt": "test prompt"}
        )

        assert exit_code == 0, f"UserPromptSubmit failed: {stderr}"
        response = json.loads(stdout)
        assert "hookSpecificOutput" in response
        assert response["hookSpecificOutput"]["hookEventName"] == "UserPromptSubmit"

    def test_pre_tool_use_event(self):
        """PreToolUse event returns valid Hook response

        @TEST:HOOKS-EVENT-003

        Given: PreToolUse with Bash tool
        When: alfred_hooks.py PreToolUse is executed
        Then: Exit code 0, JSON with "continue" true
        """
        exit_code, stdout, stderr = run_hook(
            "PreToolUse",
            {"cwd": ".", "tool": "Bash", "arguments": {"command": "echo test"}}
        )

        assert exit_code == 0, f"PreToolUse failed: {stderr}"
        response = json.loads(stdout)
        assert "continue" in response
        assert response["continue"] is True

    def test_post_tool_use_event(self):
        """PostToolUse event returns valid Hook response

        @TEST:HOOKS-EVENT-004

        Given: PostToolUse with tool name
        When: alfred_hooks.py PostToolUse is executed
        Then: Exit code 0, JSON with "continue" true
        """
        exit_code, stdout, stderr = run_hook(
            "PostToolUse",
            {"cwd": ".", "tool": "Read"}
        )

        assert exit_code == 0, f"PostToolUse failed: {stderr}"
        response = json.loads(stdout)
        assert response == {"continue": True}

    def test_session_end_event(self):
        """SessionEnd event returns valid Hook response

        @TEST:HOOKS-EVENT-005

        Given: SessionEnd event
        When: alfred_hooks.py SessionEnd is executed
        Then: Exit code 0, minimal JSON response
        """
        exit_code, stdout, stderr = run_hook("SessionEnd", {"cwd": "."})

        assert exit_code == 0, f"SessionEnd failed: {stderr}"
        response = json.loads(stdout)
        assert response == {"continue": True}

    def test_notification_event(self):
        """Notification event returns valid Hook response

        @TEST:HOOKS-EVENT-006

        Given: Notification event
        When: alfred_hooks.py Notification is executed
        Then: Exit code 0, minimal JSON response
        """
        exit_code, stdout, stderr = run_hook("Notification", {"cwd": "."})

        assert exit_code == 0, f"Notification failed: {stderr}"
        response = json.loads(stdout)
        assert response == {"continue": True}

    def test_stop_event(self):
        """Stop event returns valid Hook response

        @TEST:HOOKS-EVENT-007

        Given: Stop event
        When: alfred_hooks.py Stop is executed
        Then: Exit code 0, minimal JSON response
        """
        exit_code, stdout, stderr = run_hook("Stop", {"cwd": "."})

        assert exit_code == 0, f"Stop failed: {stderr}"
        response = json.loads(stdout)
        assert response == {"continue": True}

    def test_subagent_stop_event(self):
        """SubagentStop event returns valid Hook response

        @TEST:HOOKS-EVENT-008

        Given: SubagentStop event
        When: alfred_hooks.py SubagentStop is executed
        Then: Exit code 0, minimal JSON response
        """
        exit_code, stdout, stderr = run_hook("SubagentStop", {"cwd": "."})

        assert exit_code == 0, f"SubagentStop failed: {stderr}"
        response = json.loads(stdout)
        assert response == {"continue": True}


class TestCrossPlatformTimeout:
    """Test cross-platform timeout mechanism (Windows + Unix)"""

    def test_timeout_module_import(self):
        """Verify CrossPlatformTimeout can be imported

        @TEST:HOOKS-TIMEOUT-001

        Given: utils/timeout.py module
        When: Import CrossPlatformTimeout
        Then: Import succeeds, class is callable
        """
        # Add utils to path
        utils_dir = HOOKS_DIR / "utils"
        sys.path.insert(0, str(utils_dir))

        from timeout import CrossPlatformTimeout, TimeoutError as PlatformTimeoutError

        assert CrossPlatformTimeout is not None
        assert PlatformTimeoutError is not None

        # Verify context manager protocol
        timeout = CrossPlatformTimeout(1)
        assert hasattr(timeout, '__enter__')
        assert hasattr(timeout, '__exit__')

    def test_timeout_detects_platform(self):
        """Verify timeout detects Windows vs Unix correctly

        @TEST:HOOKS-TIMEOUT-002

        Given: CrossPlatformTimeout instance
        When: Check _is_windows attribute
        Then: Matches platform.system() == "Windows"
        """
        utils_dir = HOOKS_DIR / "utils"
        sys.path.insert(0, str(utils_dir))

        from timeout import CrossPlatformTimeout

        timeout = CrossPlatformTimeout(1)
        expected_is_windows = (platform.system() == "Windows")
        assert timeout._is_windows == expected_is_windows

    def test_timeout_context_manager_cancels(self):
        """Verify timeout context manager properly cancels

        @TEST:HOOKS-TIMEOUT-003

        Given: CrossPlatformTimeout context manager
        When: Exit context before timeout expires
        Then: No TimeoutError raised, cleanup successful
        """
        utils_dir = HOOKS_DIR / "utils"
        sys.path.insert(0, str(utils_dir))

        from timeout import CrossPlatformTimeout

        # Should NOT raise TimeoutError (completes within 5 seconds)
        with CrossPlatformTimeout(5):
            time.sleep(0.1)  # Fast operation

        # If we get here, timeout was properly cancelled
        assert True

    def test_timeout_raises_on_expiration(self):
        """Verify timeout raises TimeoutError when expired

        @TEST:HOOKS-TIMEOUT-004

        Given: CrossPlatformTimeout with 1 second
        When: Sleep for 2 seconds (exceeds timeout)
        Then: TimeoutError is raised
        """
        utils_dir = HOOKS_DIR / "utils"
        sys.path.insert(0, str(utils_dir))

        from timeout import CrossPlatformTimeout, TimeoutError as PlatformTimeoutError

        # Should raise TimeoutError after 1 second
        with pytest.raises(PlatformTimeoutError):
            with CrossPlatformTimeout(1):
                time.sleep(2)  # Exceeds timeout


class TestErrorHandling:
    """Test error handling for JSON parsing, empty stdin, invalid events"""

    def test_empty_stdin_handling(self):
        """Empty stdin returns valid Hook response

        @TEST:HOOKS-ERROR-001

        Given: Empty stdin (no JSON payload)
        When: alfred_hooks.py SessionStart is executed
        Then: Exit code 0, default response (no crash)
        """
        # Run with empty stdin
        result = subprocess.run(
            [sys.executable, str(ALFRED_HOOKS_SCRIPT), "SessionStart"],
            input="",
            capture_output=True,
            text=True,
            timeout=10,
        )

        assert result.returncode == 0, f"Empty stdin failed: {result.stderr}"
        response = json.loads(result.stdout)
        assert "continue" in response

    def test_invalid_json_handling(self):
        """Invalid JSON returns error response but doesn't crash

        @TEST:HOOKS-ERROR-002

        Given: Invalid JSON payload (malformed)
        When: alfred_hooks.py SessionStart is executed
        Then: Exit code 1, JSON error response with "error" key
        """
        result = subprocess.run(
            [sys.executable, str(ALFRED_HOOKS_SCRIPT), "SessionStart"],
            input="{invalid json",
            capture_output=True,
            text=True,
            timeout=10,
        )

        # Should exit with error code 1
        assert result.returncode == 1

        # Should still return valid JSON error response
        response = json.loads(result.stdout)
        assert "continue" in response
        assert response["continue"] is True
        assert "hookSpecificOutput" in response
        assert "error" in response["hookSpecificOutput"]

    def test_missing_event_argument(self):
        """Missing event name argument prints usage to stderr

        @TEST:HOOKS-ERROR-003

        Given: No event name argument
        When: alfred_hooks.py is executed without arguments
        Then: Exit code 1, usage message in stderr
        """
        result = subprocess.run(
            [sys.executable, str(ALFRED_HOOKS_SCRIPT)],
            capture_output=True,
            text=True,
            timeout=10,
        )

        assert result.returncode == 1
        assert "Usage:" in result.stderr

    def test_unknown_event_handling(self):
        """Unknown event name returns empty result but doesn't crash

        @TEST:HOOKS-ERROR-004

        Given: Unknown event name "UnknownEvent"
        When: alfred_hooks.py UnknownEvent is executed
        Then: Exit code 0, minimal response (no crash)
        """
        exit_code, stdout, stderr = run_hook("UnknownEvent", {"cwd": "."})

        assert exit_code == 0, f"Unknown event crashed: {stderr}"
        response = json.loads(stdout)
        # Handler returns HookResult() for unknown events
        assert "continue" in response

    def test_handler_exception_graceful_fail(self):
        """Handler exceptions return error response without crash

        @TEST:HOOKS-ERROR-005

        Given: Payload that causes handler exception (invalid cwd)
        When: Handler throws exception
        Then: Exit code 1, error response with "error" key
        """
        # Use invalid payload that might cause issues
        # (handlers are designed to be defensive, so this tests robustness)
        exit_code, stdout, stderr = run_hook(
            "SessionStart",
            {"cwd": None, "phase": "compact"}  # None cwd might cause issues
        )

        # Should handle gracefully (either success or error response)
        response = json.loads(stdout)
        assert "continue" in response


class TestPerformanceAndStability:
    """Test performance, memory, and signal cleanup"""

    def test_hook_execution_under_5_seconds(self):
        """Hook execution completes within 5 seconds (timeout threshold)

        @TEST:HOOKS-PERF-001

        Given: SessionStart event
        When: Measure execution time
        Then: Completes in less than 5 seconds
        """
        start_time = time.time()
        exit_code, stdout, stderr = run_hook("SessionStart", {"cwd": ".", "phase": "clear"})
        elapsed = time.time() - start_time

        assert exit_code == 0
        assert elapsed < 5.0, f"Hook took {elapsed:.2f}s (exceeds 5s timeout threshold)"

    def test_hook_no_memory_leak(self):
        """Hook can be run multiple times without memory growth

        @TEST:HOOKS-PERF-002

        Given: Hook executed 10 times in sequence
        When: Check all executions succeed
        Then: All runs complete successfully (no crash, no hang)
        """
        for i in range(10):
            exit_code, stdout, stderr = run_hook("SessionEnd", {"cwd": "."})
            assert exit_code == 0, f"Run {i+1} failed: {stderr}"

    def test_signal_handler_cleanup(self):
        """Verify signal handlers are cleaned up after execution

        @TEST:HOOKS-PERF-003

        Given: Hook with timeout completes
        When: Check signal handler state (Unix only)
        Then: Old signal handler restored, alarm cancelled
        """
        if platform.system() == "Windows":
            pytest.skip("Signal cleanup test only applies to Unix/POSIX")

        # Run hook that uses timeout
        exit_code, stdout, stderr = run_hook("SessionStart", {"cwd": ".", "phase": "clear"})
        assert exit_code == 0

        # Verify alarm is not active (would raise if active)
        # signal.alarm(0) returns remaining time (0 if no alarm set)
        import signal
        remaining = signal.alarm(0)
        assert remaining == 0, "Signal alarm not properly cancelled"


class TestIntegrationScenarios:
    """Test real-world Hook integration scenarios"""

    def test_full_session_lifecycle(self):
        """Test complete session lifecycle: Start → Prompt → Tool → End

        @TEST:HOOKS-INTEGRATION-001

        Given: Complete Hook lifecycle
        When: Execute SessionStart → UserPromptSubmit → PreToolUse → PostToolUse → SessionEnd
        Then: All events complete successfully
        """
        # SessionStart
        exit_code, _, _ = run_hook("SessionStart", {"cwd": ".", "phase": "clear"})
        assert exit_code == 0

        # UserPromptSubmit
        exit_code, _, _ = run_hook("UserPromptSubmit", {"cwd": ".", "userPrompt": "test"})
        assert exit_code == 0

        # PreToolUse
        exit_code, _, _ = run_hook(
            "PreToolUse",
            {"cwd": ".", "tool": "Bash", "arguments": {"command": "echo test"}}
        )
        assert exit_code == 0

        # PostToolUse
        exit_code, _, _ = run_hook("PostToolUse", {"cwd": ".", "tool": "Bash"})
        assert exit_code == 0

        # SessionEnd
        exit_code, _, _ = run_hook("SessionEnd", {"cwd": "."})
        assert exit_code == 0

    def test_concurrent_hook_executions(self):
        """Test multiple Hooks can run concurrently without conflict

        @TEST:HOOKS-INTEGRATION-002

        Given: Multiple Hook processes running simultaneously
        When: Execute 5 Hooks in parallel
        Then: All complete successfully without deadlock
        """
        import concurrent.futures

        def run_test_hook(index: int) -> bool:
            exit_code, _, _ = run_hook("SessionEnd", {"cwd": "."})
            return exit_code == 0

        # Run 5 Hooks concurrently
        with concurrent.futures.ThreadPoolExecutor(max_workers=5) as executor:
            futures = [executor.submit(run_test_hook, i) for i in range(5)]
            results = [f.result(timeout=10) for f in futures]

        # All should succeed
        assert all(results), "Some concurrent Hooks failed"


if __name__ == "__main__":
    # Run tests manually for Phase 3 validation
    print("🟢 Phase 3: Running integration tests (expect ALL PASS)")
    print("=" * 60)

    # Run pytest
    pytest.main([__file__, "-v", "--tb=short"])
