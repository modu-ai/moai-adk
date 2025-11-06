#!/usr/bin/env python3
# @CODE:HOOKS-CLARITY-001 | SPEC: Individual hook files for better UX
"""SessionStart Hook: Enhanced Project Information

Claude Code Event: SessionStart
Purpose: Display enhanced project status with Git info, test status, and SPEC progress
Execution: Triggered automatically when Claude Code session begins

Enhanced Features:
- Last commit information with relative time
- Test coverage and status
- Risk assessment
- Formatted output with clear sections
"""

import json
import subprocess
import sys
from pathlib import Path
from typing import Any

# Setup import path for shared modules
HOOKS_DIR = Path(__file__).parent
SHARED_DIR = HOOKS_DIR / "shared"
if str(SHARED_DIR) not in sys.path:
    sys.path.insert(0, str(SHARED_DIR))

# Try to import existing modules, provide fallbacks if not available
try:
    from utils.timeout import CrossPlatformTimeout
    from utils.timeout import TimeoutError as PlatformTimeoutError
except ImportError:
    # Fallback timeout implementation
    import time
    import signal

    class CrossPlatformTimeout:
        def __init__(self, seconds):
            self.seconds = seconds

        def start(self):
            pass

        def cancel(self):
            pass

    class PlatformTimeoutError(Exception):
        pass


def get_git_info() -> dict[str, Any]:
    """Get comprehensive git information - optimized for speed"""
    try:
        # Get current branch
        branch = subprocess.run(
            ["git", "branch", "--show-current"],
            capture_output=True,
            text=True,
            timeout=0.5
        ).stdout.strip()

        # Get last commit hash and message
        last_commit = subprocess.run(
            ["git", "log", "--pretty=format:%h %s", "-1"],
            capture_output=True,
            text=True,
            timeout=0.5
        ).stdout.strip()

        # Get commit time (relative)
        commit_time = subprocess.run(
            ["git", "log", "--pretty=format:%ar", "-1"],
            capture_output=True,
            text=True,
            timeout=0.5
        ).stdout.strip()

        # Get number of changed files
        changes = subprocess.run(
            ["git", "status", "--porcelain"],
            capture_output=True,
            text=True,
            timeout=0.5
        ).stdout.strip()
        num_changes = len(changes.splitlines()) if changes else 0

        return {
            "branch": branch,
            "last_commit": last_commit,
            "commit_time": commit_time,
            "changes": num_changes
        }

    except (subprocess.TimeoutExpired, subprocess.CalledProcessError, FileNotFoundError):
        return {
            "branch": "unknown",
            "last_commit": "unknown",
            "commit_time": "unknown",
            "changes": 0
        }


def get_test_info() -> dict[str, Any]:
    """Get test coverage and status information

    OPTIMIZATION: Skipped in SessionStart hook to avoid timeout
    Reason: Running 1112+ tests on every session start (5+ seconds) is inefficient
    Tests should be run on-demand via /alfred:2-run or explicit pytest commands
    """
    return {
        "coverage": "run-on-demand",
        "status": "⏭️"
    }




def format_session_output() -> str:
    """Format minimal session start output (optimized for speed)

    Only includes essential Git information.
    Removed slow operations:
    - SPEC progress scan
    - Risk calculation
    - Test coverage check
    """
    # Gather minimal information (fast)
    git_info = get_git_info()

    # Format minimal output
    output = [
        "🚀 MoAI-ADK Session",
        f"🌿 {git_info['branch']} | 📝 {git_info['changes']} changes",
        f"📌 {git_info['last_commit']} ({git_info['commit_time']})"
    ]

    return "\n".join(output)


def main() -> None:
    """Main entry point for enhanced SessionStart hook

    Displays enhanced project information including:
    - Programming language and version
    - Git branch, changes, and last commit with time
    - SPEC progress (completed/total)
    - Test coverage and status
    - Risk assessment

    Exit Codes:
        0: Success
        1: Error (timeout, JSON parse failure, handler exception)
    """
    # Set 2-second timeout (optimized after removing pytest execution)
    timeout = CrossPlatformTimeout(2)
    timeout.start()

    try:
        # Read JSON payload from stdin (for compatibility)
        input_data = sys.stdin.read()
        data = json.loads(input_data) if input_data.strip() else {}

        # Generate enhanced session output
        session_output = format_session_output()

        # Return as system message
        result: dict[str, Any] = {
            "continue": True,
            "systemMessage": session_output
        }

        print(json.dumps(result))
        sys.exit(0)

    except PlatformTimeoutError:
        # Timeout - return minimal valid response
        timeout_response: dict[str, Any] = {
            "continue": True,
            "systemMessage": "⚠️ Session start timeout - continuing without project info",
        }
        print(json.dumps(timeout_response))
        print("SessionStart hook timeout after 2 seconds", file=sys.stderr)
        sys.exit(1)

    except json.JSONDecodeError as e:
        # JSON parse error
        error_response: dict[str, Any] = {
            "continue": True,
            "hookSpecificOutput": {"error": f"JSON parse error: {e}"},
        }
        print(json.dumps(error_response))
        print(f"SessionStart JSON parse error: {e}", file=sys.stderr)
        sys.exit(1)

    except Exception as e:
        # Unexpected error
        error_response: dict[str, Any] = {
            "continue": True,
            "hookSpecificOutput": {"error": f"SessionStart error: {e}"},
        }
        print(json.dumps(error_response))
        print(f"SessionStart unexpected error: {e}", file=sys.stderr)
        sys.exit(1)

    finally:
        # Always cancel timeout
        timeout.cancel()


if __name__ == "__main__":
    main()
