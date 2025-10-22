#!/usr/bin/env python3
"""Unit tests for PostToolUse hook handler

@TAG:POSTTOOL-AUTOTEST-001 | SPEC: posttool-autotest-design.md

TDD History:
- RED: Write comprehensive test suite (13 test cases)
- GREEN: Implement handler and 8 helper functions
- REFACTOR: Optimize performance, enhance error handling
"""

import subprocess
from pathlib import Path
from unittest.mock import MagicMock, Mock, patch

import pytest

# Import from hooks/alfred/handlers/tool.py
import sys
sys.path.insert(0, str(Path(__file__).parent.parent.parent / ".claude" / "hooks" / "alfred"))

from handlers.tool import (
    _extract_file_paths,
    _format_result,
    _get_test_command,
    _is_test_file,
    _parse_output,
    _run_tests,
    _should_run_tests,
    handle_post_tool_use,
)


class TestExtractFilePaths:
    """Test _extract_file_paths() helper function"""

    def test_write_tool_single_file(self):
        """Write tool: Extract single file path"""
        payload = {
            "tool": "Write",
            "arguments": {"file_path": "src/auth.py"},
            "cwd": "."
        }
        result = _extract_file_paths(payload)
        assert result == ["src/auth.py"]

    def test_edit_tool_single_file(self):
        """Edit tool: Extract single file path"""
        payload = {
            "tool": "Edit",
            "arguments": {"file_path": "src/user.py"},
            "cwd": "."
        }
        result = _extract_file_paths(payload)
        assert result == ["src/user.py"]

    def test_multiedit_tool_multiple_files(self):
        """MultiEdit tool: Extract multiple file paths"""
        payload = {
            "tool": "MultiEdit",
            "arguments": {
                "files": [
                    {"path": "src/auth.py"},
                    {"path": "src/user.py"},
                ]
            },
            "cwd": "."
        }
        result = _extract_file_paths(payload)
        assert result == ["src/auth.py", "src/user.py"]

    def test_missing_file_path(self):
        """Missing file_path: Return empty list"""
        payload = {
            "tool": "Write",
            "arguments": {},
            "cwd": "."
        }
        result = _extract_file_paths(payload)
        assert result == []

    def test_bash_tool_no_file_path(self):
        """Bash tool: Return empty list (no file_path)"""
        payload = {
            "tool": "Bash",
            "arguments": {"command": "ls -la"},
            "cwd": "."
        }
        result = _extract_file_paths(payload)
        assert result == []


class TestIsTestFile:
    """Test _is_test_file() filter function"""

    def test_python_test_file_prefix(self):
        """Python test file: test_*.py → True"""
        assert _is_test_file("tests/test_auth.py") is True
        assert _is_test_file("test_user.py") is True

    def test_python_test_file_suffix(self):
        """Python test file: *_test.py → True"""
        assert _is_test_file("auth_test.py") is True

    def test_typescript_test_file(self):
        """TypeScript test file: *.test.ts → True"""
        assert _is_test_file("auth.test.ts") is True
        assert _is_test_file("src/auth.test.ts") is True

    def test_javascript_test_file(self):
        """JavaScript test file: *.test.js → True"""
        assert _is_test_file("auth.test.js") is True

    def test_go_test_file(self):
        """Go test file: *_test.go → True"""
        assert _is_test_file("auth_test.go") is True

    def test_ruby_spec_file(self):
        """Ruby spec file: *_spec.rb → True"""
        assert _is_test_file("auth_spec.rb") is True
        assert _is_test_file("spec_auth.rb") is True

    def test_tests_directory(self):
        """tests/ directory: → True"""
        assert _is_test_file("tests/auth.py") is True
        assert _is_test_file("tests/unit/auth.py") is True

    def test_implementation_file(self):
        """Implementation file: → False"""
        assert _is_test_file("src/auth.py") is False
        assert _is_test_file("auth.py") is False
        assert _is_test_file("src/user.ts") is False


class TestGetTestCommand:
    """Test _get_test_command() language dispatcher"""

    def test_python_pytest(self, tmp_path):
        """Python: pytest command"""
        # Create pyproject.toml
        (tmp_path / "pyproject.toml").write_text("[tool.pytest]\n")

        cmd = _get_test_command("python", tmp_path)
        assert cmd is not None
        assert "pytest" in cmd

    def test_typescript_vitest(self, tmp_path):
        """TypeScript: vitest/jest command"""
        # Create package.json + pnpm-lock.yaml
        (tmp_path / "package.json").write_text('{"scripts": {"test": "vitest"}}\n')
        (tmp_path / "pnpm-lock.yaml").write_text("")

        cmd = _get_test_command("typescript", tmp_path)
        assert cmd is not None
        assert cmd == ["pnpm", "test"]

    def test_javascript_npm_test(self, tmp_path):
        """JavaScript: npm test command"""
        # Create package.json (no pnpm-lock.yaml)
        (tmp_path / "package.json").write_text('{"scripts": {"test": "jest"}}\n')

        cmd = _get_test_command("javascript", tmp_path)
        assert cmd is not None
        assert cmd == ["npm", "test"]

    def test_go_test(self, tmp_path):
        """Go: go test command"""
        # Create go.mod
        (tmp_path / "go.mod").write_text("module example\n")

        cmd = _get_test_command("go", tmp_path)
        assert cmd is not None
        assert cmd == ["go", "test", "-v", "./..."]

    def test_rust_cargo_test(self, tmp_path):
        """Rust: cargo test command"""
        # Create Cargo.toml
        (tmp_path / "Cargo.toml").write_text("[package]\n")

        cmd = _get_test_command("rust", tmp_path)
        assert cmd is not None
        assert cmd == ["cargo", "test", "--", "--nocapture"]

    def test_java_gradle_test(self, tmp_path):
        """Java: gradle test command"""
        # Create build.gradle.kts
        (tmp_path / "build.gradle.kts").write_text("")

        cmd = _get_test_command("java", tmp_path)
        assert cmd is not None
        assert cmd == ["gradle", "test"]

    def test_kotlin_gradle_test(self, tmp_path):
        """Kotlin: gradle test command"""
        # Create build.gradle.kts
        (tmp_path / "build.gradle.kts").write_text("")

        cmd = _get_test_command("kotlin", tmp_path)
        assert cmd is not None
        assert cmd == ["gradle", "test"]

    def test_swift_xcodebuild_test(self, tmp_path):
        """Swift: xcodebuild test command"""
        # Create Package.swift
        (tmp_path / "Package.swift").write_text("")

        cmd = _get_test_command("swift", tmp_path)
        assert cmd is not None
        assert cmd == ["swift", "test"]

    def test_dart_flutter_test(self, tmp_path):
        """Dart: flutter test command"""
        # Create pubspec.yaml
        (tmp_path / "pubspec.yaml").write_text("")

        cmd = _get_test_command("dart", tmp_path)
        assert cmd is not None
        assert cmd == ["flutter", "test"]

    def test_unsupported_language(self, tmp_path):
        """Unsupported language: Return None"""
        cmd = _get_test_command("cobol", tmp_path)
        assert cmd is None

    def test_no_test_framework(self, tmp_path):
        """No test framework detected: Return None"""
        # Empty directory (no config files)
        cmd = _get_test_command("python", tmp_path)
        assert cmd is None


class TestShouldRunTests:
    """Test _should_run_tests() trigger condition checker"""

    def test_edit_tool_triggers(self):
        """Edit tool: Should run tests"""
        assert _should_run_tests("Edit", {"file_path": "src/auth.py"}) is True

    def test_write_tool_triggers(self):
        """Write tool: Should run tests"""
        assert _should_run_tests("Write", {"file_path": "src/user.py"}) is True

    def test_multiedit_tool_triggers(self):
        """MultiEdit tool: Should run tests"""
        assert _should_run_tests("MultiEdit", {"files": [{"path": "src/auth.py"}]}) is True

    def test_bash_tool_skips(self):
        """Bash tool: Skip tests"""
        assert _should_run_tests("Bash", {"command": "ls"}) is False

    def test_read_tool_skips(self):
        """Read tool: Skip tests"""
        assert _should_run_tests("Read", {"file_path": "src/auth.py"}) is False

    def test_glob_tool_skips(self):
        """Glob tool: Skip tests"""
        assert _should_run_tests("Glob", {"pattern": "*.py"}) is False


class TestRunTests:
    """Test _run_tests() executor"""

    @patch('subprocess.run')
    def test_successful_test_run(self, mock_run):
        """Successful test: Return success=True"""
        mock_run.return_value = Mock(
            returncode=0,
            stdout="2 passed in 0.5s",
            stderr=""
        )

        passed, output = _run_tests(["pytest", "-v"], ".", timeout=10)

        assert passed is True
        assert "2 passed" in output
        mock_run.assert_called_once()

    @patch('subprocess.run')
    def test_failed_test_run(self, mock_run):
        """Failed test: Return success=False"""
        mock_run.return_value = Mock(
            returncode=1,
            stdout="1 failed, 1 passed",
            stderr=""
        )

        passed, output = _run_tests(["pytest", "-v"], ".", timeout=10)

        assert passed is False
        assert "failed" in output
        mock_run.assert_called_once()

    @patch('subprocess.run')
    def test_timeout_handling(self, mock_run):
        """Timeout: Return success=False with timeout message"""
        mock_run.side_effect = subprocess.TimeoutExpired(cmd="pytest", timeout=10)

        passed, output = _run_tests(["pytest", "-v"], ".", timeout=10)

        assert passed is False
        assert "timeout" in output.lower()

    @patch('subprocess.run')
    def test_command_error_handling(self, mock_run):
        """Command error: Return success=False with error message"""
        mock_run.side_effect = FileNotFoundError("pytest not found")

        passed, output = _run_tests(["pytest", "-v"], ".", timeout=10)

        assert passed is False
        assert "error" in output.lower()

    @patch('subprocess.run')
    def test_output_length_limit(self, mock_run):
        """Output length limit: Truncate to 1000 chars"""
        long_output = "x" * 2000
        mock_run.return_value = Mock(
            returncode=0,
            stdout=long_output,
            stderr=""
        )

        passed, output = _run_tests(["pytest", "-v"], ".", timeout=10)

        assert len(output) <= 1000


class TestParseOutput:
    """Test _parse_output() result parser"""

    def test_pytest_success_output(self):
        """pytest success: Parse passed count"""
        output = "test_auth.py::test_login PASSED\n2 passed in 0.5s"
        passed, message = _parse_output(output, "pytest")

        assert passed is True
        assert "2 passed" in message

    def test_pytest_failure_output(self):
        """pytest failure: Parse failed/passed counts"""
        output = "test_auth.py::test_login FAILED\n1 failed, 1 passed in 0.7s"
        passed, message = _parse_output(output, "pytest")

        assert passed is False
        assert "1 failed" in message

    def test_jest_success_output(self):
        """jest success: Parse test results"""
        output = "PASS  test/auth.test.ts\nTests: 2 passed, 2 total"
        passed, message = _parse_output(output, "npm test")

        assert passed is True
        assert "Tests passed" == message

    def test_go_test_success_output(self):
        """go test success: Parse PASS status"""
        output = "PASS\nok  \texample/auth\t0.5s"
        passed, message = _parse_output(output, "go test")

        assert passed is True
        assert "Tests passed" == message

    def test_cargo_test_success_output(self):
        """cargo test success: Parse passed count"""
        output = "test result: ok. 2 passed; 0 failed"
        passed, message = _parse_output(output, "cargo test")

        assert passed is True
        # Check for either specific count or generic "Tests passed"
        assert "2 passed" in message or message == "Tests passed"


class TestFormatResult:
    """Test _format_result() message formatter"""

    def test_python_success_message(self):
        """Python success: Format with pytest"""
        message = _format_result("python", True, "2 passed in 0.5s")

        assert "✅" in message
        assert "pytest" in message
        assert "2 passed" in message

    def test_python_failure_message(self):
        """Python failure: Format with pytest"""
        message = _format_result("python", False, "1 failed, 1 passed")

        assert "❌" in message
        assert "pytest" in message
        assert "failed" in message

    def test_typescript_success_message(self):
        """TypeScript success: Format with vitest/jest"""
        message = _format_result("typescript", True, "2 passed")

        assert "✅" in message
        assert ("vitest" in message or "jest" in message or "Tests passed" in message)

    def test_timeout_message(self):
        """Timeout: Format timeout message"""
        message = _format_result("python", False, "Test execution timeout (10s exceeded)")

        assert "⏱️" in message or "timeout" in message.lower()


class TestHandlePostToolUse:
    """Integration tests for handle_post_tool_use()"""

    @patch('handlers.tool._run_tests')
    @patch('handlers.tool._get_test_command')
    @patch('handlers.tool.get_project_language')
    def test_python_file_edit_triggers_tests(self, mock_lang, mock_cmd, mock_run):
        """Python file edit: Trigger pytest"""
        mock_lang.return_value = "python"
        mock_cmd.return_value = ["pytest", "-v"]
        mock_run.return_value = (True, "2 passed in 0.5s")

        payload = {
            "tool": "Edit",
            "arguments": {"file_path": "src/auth.py"},
            "cwd": "."
        }

        result = handle_post_tool_use(payload)

        assert result.blocked is False
        assert result.message is not None
        assert "✅" in result.message
        mock_run.assert_called_once()

    @patch('handlers.tool._run_tests')
    @patch('handlers.tool._get_test_command')
    @patch('handlers.tool.get_project_language')
    def test_typescript_file_edit_triggers_tests(self, mock_lang, mock_cmd, mock_run):
        """TypeScript file edit: Trigger vitest/jest"""
        mock_lang.return_value = "typescript"
        mock_cmd.return_value = ["pnpm", "test"]
        mock_run.return_value = (True, "2 passed")

        payload = {
            "tool": "Edit",
            "arguments": {"file_path": "src/auth.ts"},
            "cwd": "."
        }

        result = handle_post_tool_use(payload)

        assert result.blocked is False
        assert result.message is not None

    def test_test_file_edit_skips(self):
        """Test file edit: Skip test execution"""
        payload = {
            "tool": "Edit",
            "arguments": {"file_path": "tests/test_auth.py"},
            "cwd": "."
        }

        result = handle_post_tool_use(payload)

        assert result.blocked is False
        assert result.message is None or result.message == ""

    def test_bash_tool_skips(self):
        """Bash tool: Skip test execution"""
        payload = {
            "tool": "Bash",
            "arguments": {"command": "ls -la"},
            "cwd": "."
        }

        result = handle_post_tool_use(payload)

        assert result.blocked is False
        assert result.message is None or result.message == ""

    @patch('handlers.tool._get_test_command')
    @patch('handlers.tool.get_project_language')
    def test_no_test_framework_skips(self, mock_lang, mock_cmd):
        """No test framework: Skip silently"""
        mock_lang.return_value = "python"
        mock_cmd.return_value = None  # No test framework detected

        payload = {
            "tool": "Edit",
            "arguments": {"file_path": "src/auth.py"},
            "cwd": "."
        }

        result = handle_post_tool_use(payload)

        assert result.blocked is False
        assert result.message is None or result.message == ""

    @patch('handlers.tool._run_tests')
    @patch('handlers.tool._get_test_command')
    @patch('handlers.tool.get_project_language')
    def test_test_failure_does_not_block(self, mock_lang, mock_cmd, mock_run):
        """Test failure: Return message but do not block"""
        mock_lang.return_value = "python"
        mock_cmd.return_value = ["pytest", "-v"]
        mock_run.return_value = (False, "1 failed, 1 passed")

        payload = {
            "tool": "Edit",
            "arguments": {"file_path": "src/auth.py"},
            "cwd": "."
        }

        result = handle_post_tool_use(payload)

        assert result.blocked is False  # Non-blocking!
        assert result.message is not None
        assert "❌" in result.message
