"""Session end hook for automatic token usage submission.

This module provides a hook that is called when a Claude Code session ends,
automatically submitting the session's token usage to the MoAI Rank service.
"""

import hashlib
import os
from pathlib import Path
from typing import Any, Optional

from moai_adk.rank.client import RankClient, RankClientError, SessionSubmission
from moai_adk.rank.config import RankConfig


def compute_anonymous_project_id(project_path: str) -> str:
    """Compute an anonymized project identifier.

    Uses a hash of the project path to track sessions by project
    without revealing the actual project name or path.

    Args:
        project_path: Full path to the project directory

    Returns:
        First 16 characters of the SHA-256 hash
    """
    # Normalize the path
    normalized = os.path.normpath(os.path.expanduser(project_path))
    # Hash it
    hash_value = hashlib.sha256(normalized.encode()).hexdigest()
    return hash_value[:16]


def parse_session_data(session_data: dict[str, Any]) -> Optional[dict[str, Any]]:
    """Parse and validate session data from Claude Code.

    Args:
        session_data: Raw session data from Claude Code

    Returns:
        Parsed session data dictionary, or None if invalid
    """
    try:
        # Extract required fields
        input_tokens = session_data.get("inputTokens", 0)
        output_tokens = session_data.get("outputTokens", 0)

        # Skip sessions with no token usage
        if input_tokens == 0 and output_tokens == 0:
            return None

        # Get optional fields
        cache_creation = session_data.get("cacheCreationTokens", 0)
        cache_read = session_data.get("cacheReadTokens", 0)
        model_name = session_data.get("modelName") or session_data.get("model")
        project_path = session_data.get("projectPath") or session_data.get("cwd")

        # Get or generate ended_at timestamp
        ended_at = session_data.get("endedAt")
        if not ended_at:
            from datetime import datetime as dt
            from datetime import timezone

            ended_at = dt.now(timezone.utc).isoformat().replace("+00:00", "Z")

        return {
            "input_tokens": int(input_tokens),
            "output_tokens": int(output_tokens),
            "cache_creation_tokens": int(cache_creation),
            "cache_read_tokens": int(cache_read),
            "model_name": model_name,
            "project_path": project_path,
            "ended_at": ended_at,
        }
    except (KeyError, ValueError, TypeError):
        return None


def submit_session_hook(session_data: dict[str, Any]) -> dict[str, Any]:
    """Hook function to submit session data to MoAI Rank.

    This function is called by the Claude Code hook system when a session ends.
    It extracts token usage data and submits it to the rank service.

    Args:
        session_data: Session data dictionary from Claude Code containing:
            - inputTokens: Number of input tokens used
            - outputTokens: Number of output tokens used
            - cacheCreationTokens: Optional cache creation tokens
            - cacheReadTokens: Optional cache read tokens
            - modelName: Optional model name
            - projectPath: Optional project path for anonymous tracking
            - endedAt: Optional ISO timestamp when session ended

    Returns:
        Result dictionary with status and message
    """
    result = {
        "success": False,
        "message": "",
        "session_id": None,
    }

    # Check if user has registered
    if not RankConfig.has_credentials():
        result["message"] = "Not registered with MoAI Rank"
        return result

    # Parse session data
    parsed = parse_session_data(session_data)
    if not parsed:
        result["message"] = "No token usage to submit"
        return result

    try:
        client = RankClient()

        # Compute session hash
        session_hash = client.compute_session_hash(
            input_tokens=parsed["input_tokens"],
            output_tokens=parsed["output_tokens"],
            cache_creation_tokens=parsed["cache_creation_tokens"],
            cache_read_tokens=parsed["cache_read_tokens"],
            model_name=parsed["model_name"],
            ended_at=parsed["ended_at"],
        )

        # Compute anonymous project ID if project path available
        anonymous_project_id = None
        if parsed["project_path"]:
            anonymous_project_id = compute_anonymous_project_id(parsed["project_path"])

        # Create submission
        submission = SessionSubmission(
            session_hash=session_hash,
            ended_at=parsed["ended_at"],
            input_tokens=parsed["input_tokens"],
            output_tokens=parsed["output_tokens"],
            cache_creation_tokens=parsed["cache_creation_tokens"],
            cache_read_tokens=parsed["cache_read_tokens"],
            model_name=parsed["model_name"],
            anonymous_project_id=anonymous_project_id,
        )

        # Submit to API
        response = client.submit_session(submission)

        result["success"] = True
        result["message"] = response.get("message", "Session submitted")
        result["session_id"] = response.get("sessionId")

    except RankClientError as e:
        result["message"] = f"Submission failed: {e}"

    return result


def create_hook_script() -> str:
    """Generate a hook script for Claude Code session end.

    Returns:
        Python script content for the hook
    """
    return '''#!/usr/bin/env python3
"""MoAI Rank Session Hook

This hook submits Claude Code session token usage to the MoAI Rank service.
It is triggered automatically when a session ends.
"""

import json
import sys

def main():
    # Read session data from stdin
    try:
        input_data = sys.stdin.read()
        if not input_data:
            return

        session_data = json.loads(input_data)

        # Lazy import to avoid startup delay
        from moai_adk.rank.hook import submit_session_hook

        result = submit_session_hook(session_data)

        if result["success"]:
            print(f"Session submitted to MoAI Rank", file=sys.stderr)
        elif result["message"] != "Not registered with MoAI Rank":
            print(f"MoAI Rank: {result['message']}", file=sys.stderr)

    except json.JSONDecodeError:
        pass
    except ImportError:
        # moai-adk not installed, silently skip
        pass
    except Exception as e:
        print(f"MoAI Rank hook error: {e}", file=sys.stderr)

if __name__ == "__main__":
    main()
'''


def install_hook(project_path: Optional[Path] = None) -> bool:
    """Install the session end hook for a project.

    Args:
        project_path: Project directory (defaults to current directory)

    Returns:
        True if hook was installed successfully
    """
    if project_path is None:
        project_path = Path.cwd()

    hooks_dir = project_path / ".claude" / "hooks" / "moai"
    hooks_dir.mkdir(parents=True, exist_ok=True)

    hook_file = hooks_dir / "session_end__rank_submit.py"

    try:
        hook_file.write_text(create_hook_script())
        hook_file.chmod(0o755)  # Make executable
        return True
    except OSError:
        return False
