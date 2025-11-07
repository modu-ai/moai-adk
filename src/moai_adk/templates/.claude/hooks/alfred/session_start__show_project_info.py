#!/usr/bin/env python3
# SessionStart Hook: Enhanced Project Information
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
    """Get comprehensive git information"""
    try:
        # Get current branch
        branch = subprocess.run(
            ["git", "branch", "--show-current"],
            capture_output=True,
            text=True,
            timeout=3
        ).stdout.strip()

        # Get last commit hash and message
        last_commit = subprocess.run(
            ["git", "log", "--pretty=format:%h %s", "-1"],
            capture_output=True,
            text=True,
            timeout=3
        ).stdout.strip()

        # Get commit time (relative)
        commit_time = subprocess.run(
            ["git", "log", "--pretty=format:%ar", "-1"],
            capture_output=True,
            text=True,
            timeout=3
        ).stdout.strip()

        # Get number of changed files
        changes = subprocess.run(
            ["git", "status", "--porcelain"],
            capture_output=True,
            text=True,
            timeout=3
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

    NOTE: SessionStart hook must complete quickly (<0.5s).
    Running pytest is too slow (5+ seconds), so we skip it and return unknown status.
    Users can run tests manually with: pytest --cov

    To check test status, use: /alfred:test-status (future feature)
    """
    # Skip pytest execution - it's too slow for SessionStart
    return {
        "coverage": "unknown",
        "status": "❓"
    }


def get_spec_progress() -> dict[str, Any]:
    """Get SPEC progress information"""
    try:
        specs_dir = Path.cwd() / ".moai" / "specs"
        if not specs_dir.exists():
            return {"completed": 0, "total": 0, "percentage": 0}

        spec_folders = [d for d in specs_dir.iterdir() if d.is_dir() and d.name.startswith("SPEC-")]
        total = len(spec_folders)

        # Simple completion check - look for spec.md files
        completed = sum(1 for folder in spec_folders if (folder / "spec.md").exists())

        percentage = (completed / total * 100) if total > 0 else 0

        return {
            "completed": completed,
            "total": total,
            "percentage": round(percentage, 0)
        }

    except Exception:
        return {"completed": 0, "total": 0, "percentage": 0}


def calculate_risk(git_info: dict, spec_progress: dict, test_info: dict) -> str:
    """Calculate overall project risk level"""
    risk_score = 0

    # Git changes contribute to risk
    if git_info["changes"] > 20:
        risk_score += 10
    elif git_info["changes"] > 10:
        risk_score += 5

    # SPEC progress contributes to risk
    if spec_progress["percentage"] < 50:
        risk_score += 15
    elif spec_progress["percentage"] < 80:
        risk_score += 8

    # Test status contributes to risk
    if test_info["status"] != "✅":
        risk_score += 12
    elif test_info["coverage"] == "unknown":
        risk_score += 5

    # Determine risk level
    if risk_score >= 20:
        return "HIGH"
    elif risk_score >= 10:
        return "MEDIUM"
    else:
        return "LOW"


def format_session_output() -> str:
    """Format the complete session start output"""
    # Gather information
    git_info = get_git_info()
    spec_progress = get_spec_progress()

    # Get MoAI version from config if available
    moai_version = "unknown"
    try:
        config_path = Path.cwd() / ".moai" / "config.json"
        if config_path.exists():
            config = json.loads(config_path.read_text())
            moai_version = config.get("moai", {}).get("version", "unknown")
    except Exception:
        pass

    # Format output
    output = [
        "🚀 MoAI-ADK Session Started",
        "",
        f"🗿 Version: {moai_version} | 🌿 {git_info['branch']}",
        f"📝 Changes: {git_info['changes']}",
        f"📋 SPEC Progress: {spec_progress['completed']}/{spec_progress['total']} ({spec_progress['percentage']}%)",
        f"🔨 Last: {git_info['last_commit']} ({git_info['commit_time']})"
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
    # Set 5-second timeout
    timeout = CrossPlatformTimeout(5)
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
        print("SessionStart hook timeout after 5 seconds", file=sys.stderr)
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
