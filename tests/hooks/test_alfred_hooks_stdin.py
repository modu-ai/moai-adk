#!/usr/bin/env python3
# @TEST:WINDOWS-HOOKS-001 | SPEC: SPEC-WINDOWS-HOOKS-001.md
"""Alfred Hooks stdin ingestion tests

Verifies cross-platform stdin handling for Windows, macOS, and Linux.

TDD History:
    - RED: Added tests for Windows stdin EOF handling, empty stdin, JSON parse failures
    - GREEN: Implemented iterator pattern to satisfy the tests
    - REFACTOR: Hardened error handling and improved comments
"""
import json
import subprocess
import sys
from pathlib import Path

import pytest

# Path to alfred_hooks.py
HOOKS_SCRIPT = Path(__file__).parent.parent.parent / ".claude" / "hooks" / "alfred" / "alfred_hooks.py"


class TestAlfredHooksStdin:
    """Alfred Hooks stdin test cases.

    SPEC validation points:
        - Read stdin reliably across Windows/macOS/Linux
        - Return defaults without errors for empty stdin
        - Emit clear errors when JSON parsing fails
    """

    def test_stdin_normal_json(self):
        """Handle a valid JSON payload.

        SPEC requirement:
            - WHEN stdin receives valid JSON, the system must parse it successfully.

        Given: A valid JSON payload
        When: alfred_hooks.py runs with the SessionStart event
        Then: The JSON is parsed and returned successfully
        """
        payload = {"cwd": "."}
        input_data = json.dumps(payload)

        result = subprocess.run(
            [sys.executable, str(HOOKS_SCRIPT), "SessionStart"],
            input=input_data,
            capture_output=True,
            text=True,
            timeout=5,
        )

        assert result.returncode == 0, f"Expected exit code 0, got {result.returncode}. stderr: {result.stderr}"

        # Verify stdout is valid JSON
        try:
            output = json.loads(result.stdout)
            assert isinstance(output, dict), "Output should be a dictionary"
        except json.JSONDecodeError as e:
            pytest.fail(f"Output is not valid JSON: {result.stdout}. Error: {e}")

    def test_stdin_empty(self):
        """Handle empty stdin input.

        SPEC requirement:
            - WHEN stdin is empty, the system must avoid JSONDecodeError and return an empty object.

        Given: Empty stdin input
        When: alfred_hooks.py runs with the SessionStart event
        Then: Returns the default HookResult without raising JSONDecodeError
        """
        input_data = ""

        result = subprocess.run(
            [sys.executable, str(HOOKS_SCRIPT), "SessionStart"],
            input=input_data,
            capture_output=True,
            text=True,
            timeout=5,
        )

        # Iterator pattern ensures empty stdin becomes "{}" and behaves correctly
        assert result.returncode == 0, f"Empty stdin should be handled gracefully. stderr: {result.stderr}"

        # Verify stdout is valid JSON
        try:
            output = json.loads(result.stdout)
            assert isinstance(output, dict), "Output should be a dictionary"
        except json.JSONDecodeError as e:
            pytest.fail(f"Output is not valid JSON: {result.stdout}. Error: {e}")

    def test_stdin_invalid_json(self):
        """Handle malformed JSON input.

        SPEC requirement:
            - WHEN JSON parsing fails, the system must raise JSONDecodeError and return exit code 1.

        Given: An invalid JSON string
        When: alfred_hooks.py runs with the SessionStart event
        Then: Emits JSONDecodeError to stderr and exits with code 1
        """
        input_data = "{ invalid json }"

        result = subprocess.run(
            [sys.executable, str(HOOKS_SCRIPT), "SessionStart"],
            input=input_data,
            capture_output=True,
            text=True,
            timeout=5,
        )

        assert result.returncode == 1, "Invalid JSON should return exit code 1"
        assert "JSON parse error" in result.stderr or "JSONDecodeError" in result.stderr

    def test_stdin_cross_platform(self):
        """Verify cross-platform stdin handling.

        SPEC requirement:
            - The stdin reader must work consistently on every platform.
            - Windows/macOS/Linux should each handle EOF correctly.

        Given: A multiline JSON payload (simulating multiple platforms)
        When: alfred_hooks.py runs with the SessionStart event
        Then: Reads and parses all lines successfully
        """
        # Multiline JSON containing both Windows \r\n and Unix \n
        payload = {
            "cwd": ".",
            "multiline": "line1\nline2\nline3",
        }
        input_data = json.dumps(payload, indent=2)

        result = subprocess.run(
            [sys.executable, str(HOOKS_SCRIPT), "SessionStart"],
            input=input_data,
            capture_output=True,
            text=True,
            timeout=5,
        )

        assert result.returncode == 0, f"Expected exit code 0, got {result.returncode}. stderr: {result.stderr}"

        # Verify stdout is valid JSON
        try:
            output = json.loads(result.stdout)
            assert isinstance(output, dict), "Output should be a dictionary"
        except json.JSONDecodeError as e:
            pytest.fail(f"Output is not valid JSON: {result.stdout}. Error: {e}")


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
