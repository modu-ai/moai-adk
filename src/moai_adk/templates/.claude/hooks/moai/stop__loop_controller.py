#!/usr/bin/env python3
"""Stop Hook: Loop Controller for Ralph Engine feedback loop.

Claude Code Event: Stop
Purpose: Check completion conditions after Claude response and control feedback loop

This hook integrates with the Ralph Engine to provide automated feedback loops.
It checks whether all completion conditions are met (zero errors, tests pass, etc.)
and either signals completion or requests Claude to continue working.

Exit Codes:
- 0: Loop complete or disabled (no action needed)
- 1: Continue loop (more work needed)

Output:
- JSON with hookSpecificOutput containing loop status and next actions
- When exit code is 1, Claude will continue processing
"""

from __future__ import annotations

import json
import os
import shutil
import subprocess
import sys
from dataclasses import dataclass
from pathlib import Path
from typing import Any

# Environment variables
DISABLE_ENV_VAR = "MOAI_DISABLE_LOOP_CONTROLLER"
LOOP_ACTIVE_ENV_VAR = "MOAI_LOOP_ACTIVE"
LOOP_ITERATION_ENV_VAR = "MOAI_LOOP_ITERATION"

# State file for tracking loop status across invocations
STATE_FILE_NAME = ".moai_loop_state.json"

# Maximum state file size (64KB should be more than enough)
MAX_STATE_FILE_SIZE = 64 * 1024

# Completion markers (MoAI branding)
COMPLETION_MARKERS = [
    "<moai>DONE</moai>",
    "<moai>COMPLETE</moai>",
    "<moai:done />",
    "<moai:complete />",
]


@dataclass
class LoopState:
    """State of the feedback loop."""

    active: bool = False
    iteration: int = 0
    max_iterations: int = 10
    last_error_count: int = 0
    last_warning_count: int = 0
    files_modified: list[str] | None = None
    start_time: float = 0.0
    completion_reason: str | None = None

    def to_dict(self) -> dict[str, Any]:
        """Convert to dictionary for JSON serialization."""
        return {
            "active": self.active,
            "iteration": self.iteration,
            "max_iterations": self.max_iterations,
            "last_error_count": self.last_error_count,
            "last_warning_count": self.last_warning_count,
            "files_modified": self.files_modified or [],
            "start_time": self.start_time,
            "completion_reason": self.completion_reason,
        }

    @classmethod
    def from_dict(cls, data: dict[str, Any]) -> LoopState:
        """Create from dictionary."""
        return cls(
            active=data.get("active", False),
            iteration=data.get("iteration", 0),
            max_iterations=data.get("max_iterations", 10),
            last_error_count=data.get("last_error_count", 0),
            last_warning_count=data.get("last_warning_count", 0),
            files_modified=data.get("files_modified"),
            start_time=data.get("start_time", 0.0),
            completion_reason=data.get("completion_reason"),
        )


@dataclass
class CompletionStatus:
    """Status of completion conditions."""

    zero_errors: bool = False
    zero_warnings: bool = True  # Warnings often allowed
    tests_pass: bool = False
    coverage_met: bool = False
    all_conditions_met: bool = False
    details: dict[str, Any] | None = None


def get_project_dir() -> Path:
    """Get the project directory from environment or current working directory."""
    project_dir = os.environ.get("CLAUDE_PROJECT_DIR", os.getcwd())
    return Path(project_dir)


def get_state_file_path() -> Path:
    """Get path to the loop state file."""
    return get_project_dir() / ".moai" / "cache" / STATE_FILE_NAME


def load_loop_state() -> LoopState:
    """Load loop state from file or environment.

    Uses direct file access without exists() check to prevent TOCTOU race conditions.
    Includes file size limit to prevent memory exhaustion.

    Returns:
        LoopState instance.
    """
    # First check environment variables (set by commands)
    if os.environ.get(LOOP_ACTIVE_ENV_VAR, "").lower() in ("1", "true", "yes"):
        iteration = int(os.environ.get(LOOP_ITERATION_ENV_VAR, "0"))
        return LoopState(active=True, iteration=iteration)

    # Then check state file - use try/except instead of exists() to prevent race condition
    state_path = get_state_file_path()
    try:
        # Check file size before reading to prevent memory exhaustion
        file_size = state_path.stat().st_size
        if file_size > MAX_STATE_FILE_SIZE:
            return LoopState()  # File too large, return default

        with open(state_path, encoding="utf-8", errors="replace") as f:
            data = json.load(f)
            return LoopState.from_dict(data)
    except FileNotFoundError:
        pass  # File doesn't exist, return default
    except (OSError, json.JSONDecodeError, ValueError, KeyError):
        pass  # File corrupted or invalid, return default

    return LoopState()


def save_loop_state(state: LoopState) -> None:
    """Save loop state to file using atomic write.

    Uses write-to-temp-then-rename pattern to prevent race conditions
    and partial writes from corrupting state.

    Args:
        state: LoopState to save.
    """
    import tempfile

    state_path = get_state_file_path()
    state_path.parent.mkdir(parents=True, exist_ok=True)

    try:
        # Write to temporary file first, then atomic rename
        # This prevents partial writes from corrupting state
        fd, temp_path = tempfile.mkstemp(dir=state_path.parent, prefix=".tmp_", suffix=".json")
        try:
            with os.fdopen(fd, "w", encoding="utf-8", errors="replace") as f:
                json.dump(state.to_dict(), f, indent=2, ensure_ascii=False)
            # Atomic rename (on POSIX systems)
            os.replace(temp_path, state_path)
        except Exception:
            # Clean up temp file on error
            try:
                os.unlink(temp_path)
            except OSError:
                pass
            raise
    except Exception:
        pass  # Graceful degradation


def clear_loop_state() -> None:
    """Clear the loop state file."""
    state_path = get_state_file_path()
    try:
        if state_path.exists():
            state_path.unlink()
    except Exception:
        pass


def load_ralph_config() -> dict[str, Any]:
    """Load Ralph configuration from ralph.yaml.

    Returns:
        Configuration dictionary with defaults.
    """
    config: dict[str, Any] = {
        "enabled": True,
        "loop": {
            "max_iterations": 10,
            "auto_fix": False,
            "require_confirmation": True,
            "cooldown_seconds": 2,
            "completion": {
                "zero_errors": True,
                "zero_warnings": False,
                "tests_pass": True,
                "coverage_threshold": 85,
            },
        },
        "hooks": {
            "stop_loop_controller": {
                "enabled": True,
                "check_completion": True,
            }
        },
    }

    # Try to load from config file
    config_path = get_project_dir() / ".moai" / "config" / "sections" / "ralph.yaml"
    if config_path.exists():
        try:
            import yaml

            with open(config_path) as f:
                loaded = yaml.safe_load(f)
                if loaded and "ralph" in loaded:
                    ralph = loaded["ralph"]
                    if "enabled" in ralph:
                        config["enabled"] = ralph["enabled"]
                    if "loop" in ralph:
                        config["loop"].update(ralph["loop"])
                    if "hooks" in ralph and "stop_loop_controller" in ralph["hooks"]:
                        config["hooks"]["stop_loop_controller"].update(ralph["hooks"]["stop_loop_controller"])
        except Exception:
            pass

    return config


def check_lsp_errors() -> tuple[int, int]:
    """Check for LSP errors in recently modified files.

    Returns:
        Tuple of (error_count, warning_count).
    """
    # This is a simplified check - in practice, would query LSP client
    # For now, check if there are any .pyc compilation errors
    error_count = 0
    warning_count = 0

    # Try ruff for Python files
    if shutil.which("ruff"):
        try:
            proc = subprocess.run(
                ["ruff", "check", "--output-format=json", "."],
                capture_output=True,
                text=True,
                timeout=60,
                cwd=str(get_project_dir()),
            )
            if proc.stdout.strip():
                issues = json.loads(proc.stdout)
                for issue in issues:
                    code = issue.get("code", "")
                    if code.startswith("E") or code.startswith("F"):
                        error_count += 1
                    else:
                        warning_count += 1
        except Exception:
            pass

    return error_count, warning_count


def check_tests_pass() -> tuple[bool, str]:
    """Check if tests pass.

    Returns:
        Tuple of (tests_pass, details).
    """
    project_dir = get_project_dir()

    # Detect test framework and run
    if (project_dir / "pyproject.toml").exists() or (project_dir / "pytest.ini").exists():
        # Python project - try pytest
        if shutil.which("pytest"):
            try:
                proc = subprocess.run(
                    ["pytest", "--tb=no", "-q", "--no-header"],
                    capture_output=True,
                    text=True,
                    timeout=120,
                    cwd=str(project_dir),
                )
                passed = proc.returncode == 0
                output = proc.stdout.strip() or proc.stderr.strip()
                return passed, output[:200]  # Truncate
            except subprocess.TimeoutExpired:
                return False, "Tests timed out"
            except Exception as e:
                return False, f"Test error: {str(e)}"

    if (project_dir / "package.json").exists():
        # JavaScript/TypeScript project - try npm test
        if shutil.which("npm"):
            try:
                proc = subprocess.run(
                    ["npm", "test", "--", "--passWithNoTests"],
                    capture_output=True,
                    text=True,
                    timeout=120,
                    cwd=str(project_dir),
                )
                passed = proc.returncode == 0
                output = proc.stdout.strip() or proc.stderr.strip()
                return passed, output[:200]
            except subprocess.TimeoutExpired:
                return False, "Tests timed out"
            except Exception as e:
                return False, f"Test error: {str(e)}"

    return True, "No test framework detected"


def check_coverage(threshold: int = 85) -> tuple[bool, float]:
    """Check if test coverage meets threshold.

    Uses try/except instead of exists() check to prevent TOCTOU race conditions.

    Args:
        threshold: Minimum coverage percentage.

    Returns:
        Tuple of (coverage_met, actual_coverage).
    """
    # Maximum coverage file size (1MB should be more than enough)
    max_coverage_file_size = 1024 * 1024

    # Try to read coverage from common locations
    project_dir = get_project_dir()

    # Check for coverage.json (pytest-cov) - use try/except to prevent race condition
    coverage_json = project_dir / "coverage.json"
    try:
        file_size = coverage_json.stat().st_size
        if file_size <= max_coverage_file_size:
            with open(coverage_json, encoding="utf-8", errors="replace") as f:
                data = json.load(f)
                total = data.get("totals", {}).get("percent_covered", 0)
                return total >= threshold, total
    except FileNotFoundError:
        pass  # File doesn't exist, try next option
    except (OSError, json.JSONDecodeError, ValueError, KeyError):
        pass  # File corrupted or invalid, try next option

    # Check for coverage.xml - use try/except to prevent race condition
    coverage_xml = project_dir / "coverage.xml"
    try:
        import xml.etree.ElementTree as ET

        file_size = coverage_xml.stat().st_size
        if file_size <= max_coverage_file_size:
            tree = ET.parse(coverage_xml)
            root = tree.getroot()
            line_rate = float(root.attrib.get("line-rate", 0)) * 100
            return line_rate >= threshold, line_rate
    except FileNotFoundError:
        pass  # File doesn't exist
    except (OSError, ValueError, KeyError, Exception):
        pass  # File corrupted, invalid XML, or other errors

    # No coverage data available - consider met to not block
    return True, -1.0


def check_completion_promise(conversation_text: str) -> bool:
    """Check if completion promise marker is present in conversation.

    Args:
        conversation_text: Full conversation text to search.

    Returns:
        True if any completion marker is found.
    """
    if not conversation_text:
        return False

    text_lower = conversation_text.lower()
    for marker in COMPLETION_MARKERS:
        if marker.lower() in text_lower:
            return True
    return False


def check_completion_conditions(config: dict[str, Any]) -> CompletionStatus:
    """Check all completion conditions.

    Args:
        config: Ralph configuration.

    Returns:
        CompletionStatus with all condition results.
    """
    completion_config = config.get("loop", {}).get("completion", {})
    status = CompletionStatus()
    details: dict[str, Any] = {}

    # Check zero errors
    if completion_config.get("zero_errors", True):
        error_count, warning_count = check_lsp_errors()
        status.zero_errors = error_count == 0
        status.zero_warnings = warning_count == 0 if completion_config.get("zero_warnings", False) else True
        details["errors"] = error_count
        details["warnings"] = warning_count
    else:
        status.zero_errors = True
        status.zero_warnings = True

    # Check tests pass
    if completion_config.get("tests_pass", True):
        tests_pass, test_details = check_tests_pass()
        status.tests_pass = tests_pass
        details["tests"] = test_details
    else:
        status.tests_pass = True

    # Check coverage
    threshold = completion_config.get("coverage_threshold", 85)
    if threshold > 0:
        coverage_met, actual_coverage = check_coverage(threshold)
        status.coverage_met = coverage_met
        details["coverage"] = actual_coverage
        details["coverage_threshold"] = threshold
    else:
        status.coverage_met = True

    # Determine if all conditions met
    required_conditions = []
    if completion_config.get("zero_errors", True):
        required_conditions.append(status.zero_errors)
    if completion_config.get("zero_warnings", False):
        required_conditions.append(status.zero_warnings)
    if completion_config.get("tests_pass", True):
        required_conditions.append(status.tests_pass)
    if threshold > 0:
        required_conditions.append(status.coverage_met)

    status.all_conditions_met = all(required_conditions) if required_conditions else True
    status.details = details

    return status


def format_loop_output(state: LoopState, status: CompletionStatus | None, action: str) -> str:
    """Format loop controller output for Claude.

    Args:
        state: Current loop state.
        status: Completion status (if checked).
        action: Action being taken.

    Returns:
        Formatted string for additionalContext.
    """
    parts = [f"Ralph Loop: {action}"]

    if state.active:
        parts.append(f"Iteration: {state.iteration}/{state.max_iterations}")

    if status and status.details:
        details = status.details
        if "errors" in details:
            parts.append(f"Errors: {details['errors']}")
        if "warnings" in details:
            parts.append(f"Warnings: {details['warnings']}")
        if "tests" in details and details["tests"] != "No test framework detected":
            parts.append(f"Tests: {'PASS' if status.tests_pass else 'FAIL'}")
        if "coverage" in details and details["coverage"] >= 0:
            parts.append(f"Coverage: {details['coverage']:.1f}%")

    if status and status.all_conditions_met:
        parts.append("Status: COMPLETE")
    elif state.active:
        parts.append("Status: CONTINUE")

    return " | ".join(parts)


def main() -> None:
    """Main hook entry point."""
    # Check if loop controller is disabled
    if os.environ.get(DISABLE_ENV_VAR, "").lower() in ("1", "true", "yes"):
        sys.exit(0)

    # Load configuration
    config = load_ralph_config()

    # Check if Ralph or loop controller hook is disabled
    if not config.get("enabled", True):
        sys.exit(0)

    hook_config = config.get("hooks", {}).get("stop_loop_controller", {})
    if not hook_config.get("enabled", True):
        sys.exit(0)

    # Load current loop state
    state = load_loop_state()

    # If loop is not active, just exit
    if not state.active:
        sys.exit(0)

    # Read input from stdin (contains conversation context)
    conversation_text = ""
    try:
        input_data = json.load(sys.stdin)
        # Extract conversation text for completion promise detection
        if "messages" in input_data:
            messages = input_data["messages"]
            if messages:
                # Get last few assistant messages for completion marker detection
                recent_messages = messages[-3:] if len(messages) >= 3 else messages
                conversation_text = " ".join(
                    [msg.get("content", "") for msg in recent_messages if msg.get("role") == "assistant"]
                )
    except (json.JSONDecodeError, OSError):
        pass  # No input available

    # PRIORITY CHECK: Completion promise marker (oh-my-opencode pattern)
    # This takes precedence over all other conditions
    if check_completion_promise(conversation_text):
        state.active = False
        state.completion_reason = "Completion promise detected"
        action = "COMPLETE - <moai>DONE</moai> detected"
        clear_loop_state()
        exit_code = 0

        output = {
            "hookSpecificOutput": {
                "hookEventName": "Stop",
                "additionalContext": f"Ralph Loop: {action}",
            }
        }
        print(json.dumps(output))
        sys.exit(0)

    # Check if we should check completion
    if not hook_config.get("check_completion", True):
        sys.exit(0)

    # Check completion conditions
    status = check_completion_conditions(config)

    # Update state
    state.last_error_count = status.details.get("errors", 0) if status.details else 0
    state.last_warning_count = status.details.get("warnings", 0) if status.details else 0

    # Determine action
    if status.all_conditions_met:
        # Loop complete
        state.active = False
        state.completion_reason = "All conditions met"
        action = "COMPLETE - All conditions satisfied"
        clear_loop_state()
        exit_code = 0
    elif state.iteration >= state.max_iterations:
        # Max iterations reached
        state.active = False
        state.completion_reason = "Max iterations reached"
        action = f"STOPPED - Max iterations ({state.max_iterations}) reached"
        clear_loop_state()
        exit_code = 0
    else:
        # Continue loop
        state.iteration += 1
        action = "CONTINUE - Issues remain"
        save_loop_state(state)
        exit_code = 1  # Signal to continue

    # Format output
    context = format_loop_output(state, status, action)

    # Build guidance for Claude
    guidance = ""
    if exit_code == 1:
        # Provide specific guidance on what to fix
        if status.details:
            issues = []
            if status.details.get("errors", 0) > 0:
                issues.append(f"Fix {status.details['errors']} error(s)")
            if not status.zero_warnings and status.details.get("warnings", 0) > 0:
                issues.append(f"Adddess {status.details['warnings']} warning(s)")
            if not status.tests_pass:
                issues.append("Fix failing tests")
            if not status.coverage_met and status.details.get("coverage", -1) >= 0:
                issues.append(
                    f"Increase coverage from {status.details['coverage']:.1f}% "
                    f"to {status.details['coverage_threshold']}%"
                )

            if issues:
                guidance = "\nNext actions: " + ", ".join(issues)

    # Prepare hook output
    output = {
        "hookSpecificOutput": {
            "hookEventName": "Stop",
            "additionalContext": context + guidance,
        }
    }

    # Output JSON
    print(json.dumps(output))

    sys.exit(exit_code)


if __name__ == "__main__":
    main()
