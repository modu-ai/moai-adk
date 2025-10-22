#!/usr/bin/env python3
"""Tool usage handlers

PreToolUse, PostToolUse event handling
"""

import re
import subprocess
from pathlib import Path

from core import HookPayload, HookResult
from core.checkpoint import create_checkpoint, detect_risky_operation
from core.project import get_project_language


def handle_pre_tool_use(payload: HookPayload) -> HookResult:
    """PreToolUse event handler (Event-Driven Checkpoint integration)

    Automatically creates checkpoints before dangerous operations.
    Called before using the Claude Code tool, it notifies the user when danger is detected.

    Args:
        payload: Claude Code event payload
                 (includes tool, arguments, cwd keys)

    Returns:
        HookResult(
            message=checkpoint creation notification (when danger is detected);
            blocked=False (always continue operation)
        )

    Checkpoint Triggers:
        - Bash:
            * critical-delete: rm -rf /, rm -rf /home, rm -rf /Users (system-level)
            * delete: rm -rf, git rm (project-level)
            * merge: git merge, git reset --hard
            * script: Python, Node, Bash execution
        - Edit/Write: CLAUDE.md, config.json (critical-file)
        - MultiEdit: ≥10 files (refactor)

    Examples:
        Bash tool (rm -rf /) critical detection:
        → "🚨 CRITICAL ALERT: System-level deletion detected!\n
           Checkpoint created: before-critical-delete-20251015-143000\n
           ⚠️  This operation could destroy your system.\n
           Please verify the command before proceeding."

        Bash tool (rm -rf src/) detection:
        → "🛡️ Checkpoint created: before-delete-20251015-143000\n
           Operation: delete"

    Notes:
        - Return blocked=False even after detection of danger (continue operation)
        - Work continues even when checkpoint fails (ignores)
        - Transparent background operation

    @TAG:CHECKPOINT-EVENT-001
    """
    tool_name = payload.get("tool", "")
    tool_args = payload.get("arguments", {})
    cwd = payload.get("cwd", ".")

    # Dangerous operation detection
    is_risky, operation_type = detect_risky_operation(tool_name, tool_args, cwd)

    # Create checkpoint when danger is detected
    if is_risky:
        checkpoint_branch = create_checkpoint(cwd, operation_type)

        if checkpoint_branch != "checkpoint-failed":
            # Critical warning for system-level deletion
            if operation_type == "critical-delete":
                message = (
                    f"🚨 CRITICAL ALERT: System-level deletion detected!\n"
                    f"   Checkpoint created: {checkpoint_branch}\n"
                    f"   ⚠️  This operation could destroy your system.\n"
                    f"   Please verify the command before proceeding."
                )
            else:
                message = (
                    f"🛡️ Checkpoint created: {checkpoint_branch}\n"
                    f"   Operation: {operation_type}"
                )

            return HookResult(message=message, blocked=False)

    return HookResult(blocked=False)


def handle_post_tool_use(payload: HookPayload) -> HookResult:
    """PostToolUse event handler (Auto Test Execution)

    Automatically runs tests after code file edits.

    Args:
        payload: Claude Code event payload
                 (includes tool, arguments, cwd keys)

    Returns:
        HookResult(
            message=test execution result summary;
            blocked=False (never blocks)
        )

    Trigger Conditions:
        - Tool: Write, Edit, MultiEdit
        - Target: Code files (not test files)
        - Detection: Language-aware test command selection

    Examples:
        Python file edit:
        → "✅ Tests passed: pytest tests/test_auth.py -v (2 passed)"

        Test failure:
        → "❌ Tests failed: 1 failed, 2 passed (see details above)"

    @TAG:POSTTOOL-AUTOTEST-001
    """
    tool_name = payload.get("tool", "")
    tool_args = payload.get("arguments", {})
    cwd = payload.get("cwd", ".")

    # Check if we should run tests
    if not _should_run_tests(tool_name, tool_args):
        return HookResult(blocked=False)

    # Extract file paths
    file_paths = _extract_file_paths(payload)
    if not file_paths:
        return HookResult(blocked=False)

    # Filter out test files
    code_files = [f for f in file_paths if not _is_test_file(f)]
    if not code_files:
        return HookResult(blocked=False)

    # Detect project language
    language = get_project_language(cwd)
    if not language or language == "Unknown Language":
        return HookResult(blocked=False)

    # Get test command
    test_command = _get_test_command(language.lower(), Path(cwd))
    if not test_command:
        return HookResult(blocked=False)

    # Run tests (non-blocking, timeout 10s)
    passed, output = _run_tests(test_command, cwd, timeout=10)

    # Format result message
    message = _format_result(language, passed, output)

    return HookResult(message=message, blocked=False)


# ============================================================================
# Helper Functions (8 functions)
# ============================================================================


def _extract_file_paths(payload: HookPayload) -> list[str]:
    """Extract file path(s) from tool arguments.

    Args:
        payload: Hook payload containing tool and arguments

    Returns:
        List of file paths (empty if none found)

    Examples:
        Write/Edit: ["src/auth.py"]
        MultiEdit: ["src/auth.py", "src/user.py"]
    """
    tool_name = payload.get("tool", "")
    tool_args = payload.get("arguments", {})

    if tool_name == "MultiEdit":
        # MultiEdit format: {"files": [{"path": "..."}]}
        files = tool_args.get("files", [])
        return [f.get("path", "") for f in files if f.get("path")]
    elif tool_name in ("Write", "Edit"):
        # Write/Edit format: {"file_path": "..."}
        file_path = tool_args.get("file_path", "")
        return [file_path] if file_path else []
    else:
        return []


def _is_test_file(file_path: str) -> bool:
    """Check if file is a test file.

    Args:
        file_path: File path to check

    Returns:
        True if test file, False otherwise

    Test file patterns:
        - tests/ directory
        - test_*.py, *_test.py
        - *.test.ts, *.test.js
        - *_spec.rb, *_test.go
    """
    path_lower = file_path.lower()

    # Check directory
    if "/tests/" in path_lower or path_lower.startswith("tests/"):
        return True

    # Check filename patterns
    test_patterns = [
        "test_",  # Python: test_auth.py
        "_test.",  # Go: auth_test.go
        ".test.",  # TS/JS: auth.test.ts
        "_spec.",  # Ruby: auth_spec.rb
        "spec_",  # Ruby: spec_auth.rb
    ]

    return any(pattern in path_lower for pattern in test_patterns)


def _should_run_tests(tool_name: str, tool_args: dict) -> bool:
    """Check if tests should run for this tool invocation.

    Args:
        tool_name: Claude Code tool name
        tool_args: Tool arguments

    Returns:
        True if tests should run, False otherwise
    """
    # Only trigger for Write/Edit/MultiEdit tools
    return tool_name in ("Write", "Edit", "MultiEdit")


def _get_test_command(language: str, cwd: Path) -> list[str] | None:
    """Get language-specific test command.

    Args:
        language: Detected language (lowercase)
        cwd: Project root directory

    Returns:
        Test command list, or None if no test framework

    Examples:
        Python → ["pytest", "-v", "--tb=short"]
        TypeScript → ["pnpm", "test"]
        Go → ["go", "test", "-v", "./..."]
    """
    # Language → test command mapping
    test_commands = {
        "python": _get_python_test_cmd,
        "typescript": _get_typescript_test_cmd,
        "javascript": _get_javascript_test_cmd,
        "go": _get_go_test_cmd,
        "rust": _get_rust_test_cmd,
        "java": _get_java_test_cmd,
        "kotlin": _get_kotlin_test_cmd,
        "swift": _get_swift_test_cmd,
        "dart": _get_dart_test_cmd,
    }

    handler = test_commands.get(language)
    if not handler:
        return None

    return handler(cwd)


def _get_python_test_cmd(cwd: Path) -> list[str] | None:
    """Python test command (pytest)."""
    pytest_ini = cwd / "pytest.ini"
    pyproject = cwd / "pyproject.toml"

    if not (pytest_ini.exists() or pyproject.exists()):
        return None

    return ["pytest", "-v", "--tb=short"]


def _get_typescript_test_cmd(cwd: Path) -> list[str] | None:
    """TypeScript test command (Vitest/Jest)."""
    package_json = cwd / "package.json"
    if not package_json.exists():
        return None

    # Check for pnpm first, then npm
    if (cwd / "pnpm-lock.yaml").exists():
        return ["pnpm", "test"]
    else:
        return ["npm", "test"]


def _get_javascript_test_cmd(cwd: Path) -> list[str] | None:
    """JavaScript test command (Jest/Mocha)."""
    package_json = cwd / "package.json"
    if not package_json.exists():
        return None

    # Check for pnpm first, then npm
    if (cwd / "pnpm-lock.yaml").exists():
        return ["pnpm", "test"]
    else:
        return ["npm", "test"]


def _get_go_test_cmd(cwd: Path) -> list[str] | None:
    """Go test command."""
    go_mod = cwd / "go.mod"
    if not go_mod.exists():
        return None

    return ["go", "test", "-v", "./..."]


def _get_rust_test_cmd(cwd: Path) -> list[str] | None:
    """Rust test command (cargo test)."""
    cargo_toml = cwd / "Cargo.toml"
    if not cargo_toml.exists():
        return None

    return ["cargo", "test", "--", "--nocapture"]


def _get_java_test_cmd(cwd: Path) -> list[str] | None:
    """Java test command (Gradle)."""
    build_gradle = cwd / "build.gradle.kts"
    if not build_gradle.exists():
        return None

    return ["gradle", "test"]


def _get_kotlin_test_cmd(cwd: Path) -> list[str] | None:
    """Kotlin test command (Gradle)."""
    build_gradle = cwd / "build.gradle.kts"
    if not build_gradle.exists():
        return None

    return ["gradle", "test"]


def _get_swift_test_cmd(cwd: Path) -> list[str] | None:
    """Swift test command."""
    package_swift = cwd / "Package.swift"
    if not package_swift.exists():
        return None

    return ["swift", "test"]


def _get_dart_test_cmd(cwd: Path) -> list[str] | None:
    """Dart/Flutter test command."""
    pubspec = cwd / "pubspec.yaml"
    if not pubspec.exists():
        return None

    return ["flutter", "test"]


def _run_tests(cmd: list[str], cwd: str, timeout: int = 10) -> tuple[bool, str]:
    """Run test command and return results.

    Args:
        cmd: Test command to execute
        cwd: Working directory
        timeout: Timeout in seconds (default 10)

    Returns:
        Tuple of (passed: bool, output: str)
        - passed: True if tests passed, False otherwise
        - output: Test output (truncated to 1000 chars)

    Examples:
        Success: (True, "2 passed in 0.5s")
        Failure: (False, "1 failed, 1 passed in 0.7s")
        Timeout: (False, "Test execution timeout (10s exceeded)")
    """
    try:
        result = subprocess.run(
            cmd,
            cwd=cwd,
            capture_output=True,
            text=True,
            timeout=timeout,
            check=False,  # Don't raise on non-zero exit
        )

        # Combine stdout and stderr
        output = result.stdout + result.stderr
        output = output[:1000]  # Truncate to 1000 chars

        # Parse test results
        passed, message = _parse_output(output, " ".join(cmd))

        return (passed, message)

    except subprocess.TimeoutExpired:
        return (False, f"Test execution timeout ({timeout}s exceeded)")
    except Exception as e:
        return (False, f"Test execution error: {str(e)}")


def _parse_output(output: str, command: str) -> tuple[bool, str]:
    """Parse test framework output.

    Args:
        output: Raw test output
        command: Test command string (for framework detection)

    Returns:
        Tuple of (passed: bool, parsed_message: str)
    """
    # Pytest output parsing
    if "pytest" in command:
        passed_match = re.search(r"(\d+) passed", output)
        failed_match = re.search(r"(\d+) failed", output)

        passed = int(passed_match.group(1)) if passed_match else 0
        failed = int(failed_match.group(1)) if failed_match else 0

        if failed == 0 and passed > 0:
            return (True, f"{passed} passed")
        else:
            return (False, f"{failed} failed, {passed} passed")

    # Jest/Vitest output parsing
    if "npm test" in command or "pnpm test" in command:
        if "PASS" in output or "passed" in output.lower():
            return (True, "Tests passed")
        elif "FAIL" in output or "failed" in output.lower():
            return (False, "Tests failed")

    # Go test output parsing
    if "go test" in command:
        if "PASS" in output or "ok" in output:
            return (True, "Tests passed")
        else:
            return (False, "Tests failed")

    # Cargo test output parsing
    if "cargo test" in command:
        # Look for "test result: ok. X passed"
        passed_match = re.search(r"(\d+) passed", output)
        if passed_match:
            passed = int(passed_match.group(1))
            return (True, f"{passed} passed")
        # Fall back to checking for "ok" keyword
        elif "ok" in output.lower():
            return (True, "Tests passed")
        else:
            return (False, "Tests failed")

    # Generic fallback
    return (True, "Tests completed") if "error" not in output.lower() else (False, "Tests failed")


def _format_result(language: str, passed: bool, output: str) -> str:
    """Format test result message.

    Args:
        language: Project language
        passed: Whether tests passed
        output: Test output message

    Returns:
        Formatted message string with emoji and framework info
    """
    # Get framework name
    framework_map = {
        "python": "pytest",
        "typescript": "vitest/jest",
        "javascript": "jest",
        "go": "go test",
        "rust": "cargo test",
        "java": "gradle test",
        "kotlin": "gradle test",
        "swift": "swift test",
        "dart": "flutter test",
    }
    framework = framework_map.get(language.lower(), "test runner")

    # Format message
    if passed:
        return f"✅ Tests passed ({framework})\n   {output}"
    else:
        return f"❌ Tests failed ({framework})\n   {output}"


__all__ = ["handle_pre_tool_use", "handle_post_tool_use"]
