"""Session end hook for automatic token usage submission.

This module provides a hook that is called when a Claude Code session ends,
automatically submitting the session's token usage to the MoAI Rank service.

The hook is installed globally at ~/.claude/hooks/moai/ by default to collect
session data from all projects. Users can opt-out specific projects via
~/.moai/rank/config.yaml configuration.

Token Usage Collection:
    Claude Code's SessionEnd hook only provides metadata (session_id, transcript_path,
    cwd, etc.) - NOT token counts. Token usage must be extracted from the transcript
    JSONL file by parsing message.usage fields.

Cost Calculation:
    Based on Anthropic's pricing (as of 2024):
    - Claude 3.5 Sonnet: Input $3/MTok, Output $15/MTok
    - Claude 3 Opus: Input $15/MTok, Output $75/MTok
    - Cache Creation: $3.75/MTok (Sonnet), $18.75/MTok (Opus)
    - Cache Read: $0.30/MTok (Sonnet), $1.50/MTok (Opus)
"""

import hashlib
import os
from dataclasses import dataclass
from pathlib import Path
from typing import Any, Optional

from moai_adk.rank.client import RankClient, RankClientError, SessionSubmission
from moai_adk.rank.config import RankConfig

# Token pricing per million tokens (USD)
# Source: https://platform.claude.com/docs/en/about-claude/pricing
# Last updated: 2025-01
MODEL_PRICING: dict[str, dict[str, float]] = {
    # Claude Opus 4.5 - Latest flagship model
    "claude-opus-4-5-20251101": {
        "input": 5.00,
        "output": 25.00,
        "cache_creation": 6.25,  # 5m cache writes
        "cache_read": 0.50,
    },
    # Claude Opus 4.1
    "claude-opus-4-1-20250414": {
        "input": 15.00,
        "output": 75.00,
        "cache_creation": 18.75,
        "cache_read": 1.50,
    },
    # Claude Opus 4
    "claude-opus-4-20250514": {
        "input": 15.00,
        "output": 75.00,
        "cache_creation": 18.75,
        "cache_read": 1.50,
    },
    # Claude Sonnet 4.5
    "claude-sonnet-4-5-20251022": {
        "input": 3.00,
        "output": 15.00,
        "cache_creation": 3.75,
        "cache_read": 0.30,
    },
    # Claude Sonnet 4
    "claude-sonnet-4-20250514": {
        "input": 3.00,
        "output": 15.00,
        "cache_creation": 3.75,
        "cache_read": 0.30,
    },
    # Claude Sonnet 3.7 (deprecated but still supported)
    "claude-3-7-sonnet-20250219": {
        "input": 3.00,
        "output": 15.00,
        "cache_creation": 3.75,
        "cache_read": 0.30,
    },
    # Claude Haiku 4.5
    "claude-haiku-4-5-20251022": {
        "input": 1.00,
        "output": 5.00,
        "cache_creation": 1.25,
        "cache_read": 0.10,
    },
    # Claude Haiku 3.5
    "claude-3-5-haiku-20241022": {
        "input": 0.80,
        "output": 4.00,
        "cache_creation": 1.00,
        "cache_read": 0.08,
    },
    # Claude Opus 3 (deprecated)
    "claude-3-opus-20240229": {
        "input": 15.00,
        "output": 75.00,
        "cache_creation": 18.75,
        "cache_read": 1.50,
    },
    # Claude Haiku 3
    "claude-3-haiku-20240307": {
        "input": 0.25,
        "output": 1.25,
        "cache_creation": 0.30,
        "cache_read": 0.03,
    },
    # Legacy Claude 3.5 Sonnet versions
    "claude-3-5-sonnet-20241022": {
        "input": 3.00,
        "output": 15.00,
        "cache_creation": 3.75,
        "cache_read": 0.30,
    },
    "claude-3-5-sonnet-20240620": {
        "input": 3.00,
        "output": 15.00,
        "cache_creation": 3.75,
        "cache_read": 0.30,
    },
    # Default fallback (Sonnet 4 pricing)
    "default": {
        "input": 3.00,
        "output": 15.00,
        "cache_creation": 3.75,
        "cache_read": 0.30,
    },
}


@dataclass
class TokenUsage:
    """Token usage extracted from a session transcript.

    Attributes:
        input_tokens: Total input tokens consumed
        output_tokens: Total output tokens generated
        cache_creation_tokens: Tokens used for cache creation
        cache_read_tokens: Tokens read from cache
        model_name: Primary model used in the session
        cost_usd: Calculated cost in USD

        # Dashboard fields (for activity visualization)
        started_at: Session start timestamp (UTC ISO format)
        duration_seconds: Total session duration in seconds
        turn_count: Number of user turns (messages)
        tool_usage: Tool usage counts (e.g., {"Read": 5, "Write": 3})
        model_usage: Per-model token usage (e.g., {"claude-opus-4-5": {"input": 5000, "output": 2000}})
        code_metrics: Code change metrics (linesAdded, linesDeleted, filesModified, filesCreated)
    """

    input_tokens: int = 0
    output_tokens: int = 0
    cache_creation_tokens: int = 0
    cache_read_tokens: int = 0
    model_name: Optional[str] = None
    cost_usd: float = 0.0

    # Dashboard fields for activity visualization
    started_at: Optional[str] = None
    duration_seconds: int = 0
    turn_count: int = 0
    tool_usage: Optional[dict[str, int]] = None
    model_usage: Optional[dict[str, dict[str, int]]] = None
    code_metrics: Optional[dict[str, int]] = None


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


def get_model_pricing(model_name: Optional[str]) -> dict[str, float]:
    """Get pricing for a specific model.

    Uses exact match first, then falls back to pattern matching
    for model families (opus, sonnet, haiku).

    Args:
        model_name: Model identifier string (e.g., 'claude-sonnet-4-20250514')

    Returns:
        Dictionary with pricing per million tokens:
        - input: Base input token price
        - output: Output token price
        - cache_creation: 5-minute cache write price
        - cache_read: Cache hit/refresh price
    """
    if not model_name:
        return MODEL_PRICING["default"]

    # Try exact match first
    if model_name in MODEL_PRICING:
        return MODEL_PRICING[model_name]

    # Try pattern matching for model families
    model_lower = model_name.lower()

    # Opus family - check version specifics
    if "opus" in model_lower:
        if "4-5" in model_lower or "4.5" in model_lower:
            return MODEL_PRICING["claude-opus-4-5-20251101"]
        elif "4-1" in model_lower or "4.1" in model_lower:
            return MODEL_PRICING["claude-opus-4-1-20250414"]
        elif "opus-4" in model_lower or "opus4" in model_lower:
            return MODEL_PRICING["claude-opus-4-20250514"]
        elif "opus-3" in model_lower or "opus3" in model_lower:
            return MODEL_PRICING["claude-3-opus-20240229"]
        # Default to Opus 4 pricing for unknown opus versions
        return MODEL_PRICING["claude-opus-4-20250514"]

    # Haiku family
    elif "haiku" in model_lower:
        if "4-5" in model_lower or "4.5" in model_lower:
            return MODEL_PRICING["claude-haiku-4-5-20251022"]
        elif "3-5" in model_lower or "3.5" in model_lower:
            return MODEL_PRICING["claude-3-5-haiku-20241022"]
        elif "haiku-3" in model_lower or "haiku3" in model_lower:
            return MODEL_PRICING["claude-3-haiku-20240307"]
        # Default to Haiku 3.5 pricing
        return MODEL_PRICING["claude-3-5-haiku-20241022"]

    # Sonnet family
    elif "sonnet" in model_lower:
        if "4-5" in model_lower or "4.5" in model_lower:
            return MODEL_PRICING["claude-sonnet-4-5-20251022"]
        elif "sonnet-4" in model_lower or "sonnet4" in model_lower:
            return MODEL_PRICING["claude-sonnet-4-20250514"]
        elif "3-7" in model_lower or "3.7" in model_lower:
            return MODEL_PRICING["claude-3-7-sonnet-20250219"]
        elif "3-5" in model_lower or "3.5" in model_lower:
            return MODEL_PRICING["claude-3-5-sonnet-20241022"]
        # Default to Sonnet 4 pricing
        return MODEL_PRICING["claude-sonnet-4-20250514"]

    return MODEL_PRICING["default"]


def calculate_cost(usage: TokenUsage) -> float:
    """Calculate the cost in USD for token usage.

    Args:
        usage: TokenUsage dataclass with token counts and model

    Returns:
        Cost in USD
    """
    pricing = get_model_pricing(usage.model_name)

    # Cost per token (pricing is per million tokens)
    input_cost = (usage.input_tokens / 1_000_000) * pricing["input"]
    output_cost = (usage.output_tokens / 1_000_000) * pricing["output"]
    cache_creation_cost = (usage.cache_creation_tokens / 1_000_000) * pricing["cache_creation"]
    cache_read_cost = (usage.cache_read_tokens / 1_000_000) * pricing["cache_read"]

    return input_cost + output_cost + cache_creation_cost + cache_read_cost


def parse_transcript_for_usage(transcript_path: str) -> Optional[TokenUsage]:
    """Parse a JSONL transcript file to extract total token usage and session metadata.

    Claude Code stores conversation transcripts as JSONL files where each line
    contains a message object. Token usage is in the message.usage field.

    Extracts:
        - Token counts: input, output, cache creation, cache read
        - Model name: Primary model used
        - Session timing: started_at, duration_seconds
        - User engagement: turn_count (number of user messages)
        - Tool usage: Count of each tool used (e.g., {"Read": 5, "Write": 3})

    Args:
        transcript_path: Path to the transcript JSONL file

    Returns:
        TokenUsage dataclass with aggregated data, or None if parsing fails
    """
    import json
    from datetime import datetime as dt

    usage = TokenUsage()
    transcript_file = Path(transcript_path)

    if not transcript_file.exists():
        return None

    first_timestamp: Optional[str] = None
    last_timestamp: Optional[str] = None
    turn_count = 0
    tool_counts: dict[str, int] = {}
    model_counts: dict[str, dict[str, int]] = {}
    current_model: Optional[str] = None

    # Code metrics tracking
    lines_added = 0
    lines_deleted = 0
    files_modified: set[str] = set()
    files_created: set[str] = set()

    def count_lines(text: str) -> int:
        """Count non-empty lines in text."""
        if not text:
            return 0
        return len([line for line in text.split("\n") if line.strip()])

    try:
        with open(transcript_file, encoding="utf-8") as f:
            for line in f:
                line = line.strip()
                if not line:
                    continue

                try:
                    data = json.loads(line)
                except json.JSONDecodeError:
                    continue

                # Extract timestamp for session timing
                timestamp = data.get("timestamp")
                if timestamp:
                    if first_timestamp is None:
                        first_timestamp = timestamp
                    last_timestamp = timestamp

                # Count user turns
                msg_type = data.get("type")
                if msg_type == "user":
                    turn_count += 1

                # Extract usage from message.usage field
                message = data.get("message", {})
                msg_usage = message.get("usage", {})

                # Extract model name first (needed for per-model tracking)
                model = message.get("model") or data.get("model")
                if model:
                    current_model = model
                    usage.model_name = model

                if msg_usage:
                    input_toks = msg_usage.get("input_tokens", 0)
                    output_toks = msg_usage.get("output_tokens", 0)

                    # Aggregate total tokens
                    usage.input_tokens += input_toks
                    usage.output_tokens += output_toks
                    usage.cache_creation_tokens += msg_usage.get("cache_creation_input_tokens", 0)
                    usage.cache_read_tokens += msg_usage.get("cache_read_input_tokens", 0)

                    # Track per-model token usage
                    if current_model:
                        if current_model not in model_counts:
                            model_counts[current_model] = {"input": 0, "output": 0}
                        model_counts[current_model]["input"] += input_toks
                        model_counts[current_model]["output"] += output_toks

                # Count tool usage and extract code metrics from message.content
                content = message.get("content", [])
                if isinstance(content, list):
                    for block in content:
                        if isinstance(block, dict) and block.get("type") == "tool_use":
                            tool_name = block.get("name", "unknown")
                            tool_counts[tool_name] = tool_counts.get(tool_name, 0) + 1

                            # Extract code metrics from Edit/Write tools
                            tool_input = block.get("input", {})
                            if isinstance(tool_input, dict):
                                file_path = tool_input.get("file_path", "")

                                if tool_name == "Edit":
                                    # Edit: count old_string (deleted) and new_string (added)
                                    old_str = tool_input.get("old_string", "")
                                    new_str = tool_input.get("new_string", "")
                                    lines_deleted += count_lines(old_str)
                                    lines_added += count_lines(new_str)
                                    if file_path:
                                        files_modified.add(file_path)

                                elif tool_name == "Write":
                                    # Write: count content lines as added
                                    content_str = tool_input.get("content", "")
                                    lines_added += count_lines(content_str)
                                    if file_path:
                                        files_created.add(file_path)

                                elif tool_name == "MultiEdit":
                                    # MultiEdit: process multiple edits
                                    edits = tool_input.get("edits", [])
                                    if isinstance(edits, list):
                                        for edit in edits:
                                            if isinstance(edit, dict):
                                                old_str = edit.get("old_string", "")
                                                new_str = edit.get("new_string", "")
                                                lines_deleted += count_lines(old_str)
                                                lines_added += count_lines(new_str)
                                    if file_path:
                                        files_modified.add(file_path)

        # Set dashboard fields
        usage.started_at = first_timestamp
        usage.turn_count = turn_count
        usage.tool_usage = tool_counts if tool_counts else None
        usage.model_usage = model_counts if model_counts else None

        # Set code metrics (only if there was any code activity)
        if lines_added > 0 or lines_deleted > 0 or files_modified or files_created:
            usage.code_metrics = {
                "linesAdded": lines_added,
                "linesDeleted": lines_deleted,
                "filesModified": len(files_modified),
                "filesCreated": len(files_created),
            }

        # Calculate duration in seconds
        if first_timestamp and last_timestamp:
            try:
                # Parse ISO format timestamps
                start_dt = dt.fromisoformat(first_timestamp.replace("Z", "+00:00"))
                end_dt = dt.fromisoformat(last_timestamp.replace("Z", "+00:00"))
                usage.duration_seconds = int((end_dt - start_dt).total_seconds())
            except (ValueError, TypeError):
                usage.duration_seconds = 0

        # Calculate cost
        if usage.input_tokens > 0 or usage.output_tokens > 0:
            usage.cost_usd = calculate_cost(usage)
            return usage

        return None

    except (OSError, IOError):
        return None


def parse_session_data(session_data: dict[str, Any]) -> Optional[dict[str, Any]]:
    """Parse and validate session data from Claude Code SessionEnd hook.

    IMPORTANT: Claude Code's SessionEnd hook does NOT provide token counts directly.
    Token usage must be extracted from the transcript_path JSONL file.

    Args:
        session_data: Raw session data from Claude Code containing:
            - session_id: Unique session identifier
            - transcript_path: Path to the JSONL transcript file
            - cwd: Current working directory (project path)
            - permission_mode: Permission mode used
            - hook_event_name: Name of the hook event
            - reason: Reason for session end

    Returns:
        Parsed session data dictionary with token usage, or None if invalid
    """
    try:
        # Get transcript path - this is where token data lives
        transcript_path = session_data.get("transcript_path")
        if not transcript_path:
            return None

        # Parse the transcript file for token usage
        usage = parse_transcript_for_usage(transcript_path)
        if not usage:
            return None

        # Skip sessions with no token usage
        if usage.input_tokens == 0 and usage.output_tokens == 0:
            return None

        # Get project path from cwd
        project_path = session_data.get("cwd")

        # Generate ended_at timestamp
        from datetime import datetime as dt
        from datetime import timezone

        ended_at = dt.now(timezone.utc).isoformat().replace("+00:00", "Z")

        return {
            "input_tokens": usage.input_tokens,
            "output_tokens": usage.output_tokens,
            "cache_creation_tokens": usage.cache_creation_tokens,
            "cache_read_tokens": usage.cache_read_tokens,
            "model_name": usage.model_name,
            "project_path": project_path,
            "ended_at": ended_at,
            "cost_usd": usage.cost_usd,
            "session_id": session_data.get("session_id"),
            # Dashboard fields
            "started_at": usage.started_at,
            "duration_seconds": usage.duration_seconds,
            "turn_count": usage.turn_count,
            "tool_usage": usage.tool_usage,
            "model_usage": usage.model_usage,
            "code_metrics": usage.code_metrics,
        }
    except (KeyError, ValueError, TypeError):
        return None


def submit_session_hook(session_data: dict[str, Any]) -> dict[str, Any]:
    """Hook function to submit session data to MoAI Rank.

    This function is called by the Claude Code hook system when a session ends.
    It parses the transcript JSONL file to extract token usage and submits it
    to the rank service.

    Args:
        session_data: Session data dictionary from Claude Code SessionEnd hook:
            - session_id: Unique session identifier
            - transcript_path: Path to the JSONL transcript file (contains token data)
            - cwd: Current working directory (project path)
            - permission_mode: Permission mode used
            - hook_event_name: Name of the hook event
            - reason: Reason for session end

    Returns:
        Result dictionary with:
            - success: Whether submission succeeded
            - message: Status message
            - session_id: Session ID from server (if successful)
            - cost_usd: Calculated cost in USD (if successful)
    """
    result = {
        "success": False,
        "message": "",
        "session_id": None,
        "cost_usd": 0.0,
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

        # Create submission
        submission = SessionSubmission(
            session_hash=session_hash,
            ended_at=parsed["ended_at"],
            input_tokens=parsed["input_tokens"],
            output_tokens=parsed["output_tokens"],
            cache_creation_tokens=parsed["cache_creation_tokens"],
            cache_read_tokens=parsed["cache_read_tokens"],
            model_name=parsed["model_name"],
            # Dashboard fields
            started_at=parsed.get("started_at"),
            duration_seconds=parsed.get("duration_seconds", 0),
            turn_count=parsed.get("turn_count", 0),
            tool_usage=parsed.get("tool_usage"),
            model_usage=parsed.get("model_usage"),
            code_metrics=parsed.get("code_metrics"),
        )

        # Submit to API
        response = client.submit_session(submission)

        result["success"] = True
        result["message"] = response.get("message", "Session submitted")
        result["session_id"] = response.get("sessionId")
        result["cost_usd"] = parsed.get("cost_usd", 0.0)

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


def load_rank_config() -> dict[str, Any]:
    """Load rank configuration from ~/.moai/rank/config.yaml.

    Returns:
        Configuration dictionary with rank settings
    """
    import yaml

    config_file = Path.home() / ".moai" / "rank" / "config.yaml"
    if not config_file.exists():
        return {"enabled": True, "exclude_projects": []}

    try:
        with open(config_file) as f:
            data = yaml.safe_load(f) or {}
        rank_config = data.get("rank", {})
        return {
            "enabled": rank_config.get("enabled", True),
            "exclude_projects": rank_config.get("exclude_projects", []),
        }
    except (OSError, yaml.YAMLError):
        return {"enabled": True, "exclude_projects": []}


def is_project_excluded(project_path: Optional[str]) -> bool:
    """Check if a project is excluded from rank submission.

    Args:
        project_path: Path to the project directory

    Returns:
        True if project should be excluded from submission
    """
    import fnmatch

    if not project_path:
        return False

    config = load_rank_config()

    # Check if rank is globally disabled
    if not config.get("enabled", True):
        return True

    # Normalize the project path
    normalized_path = os.path.normpath(os.path.expanduser(project_path))

    # Check against exclusion patterns
    for pattern in config.get("exclude_projects", []):
        pattern = os.path.expanduser(pattern)
        # Support both exact paths and wildcard patterns
        if fnmatch.fnmatch(normalized_path, pattern):
            return True
        # Also check if pattern is a prefix (directory match)
        if normalized_path.startswith(os.path.normpath(pattern)):
            return True

    return False


def save_rank_config(config: dict[str, Any]) -> bool:
    """Save rank configuration to ~/.moai/rank/config.yaml.

    Args:
        config: Configuration dictionary with rank settings

    Returns:
        True if saved successfully
    """
    import yaml

    config_dir = Path.home() / ".moai" / "rank"
    config_file = config_dir / "config.yaml"

    try:
        config_dir.mkdir(parents=True, exist_ok=True)

        # Load existing config to preserve other settings
        existing = {}
        if config_file.exists():
            with open(config_file) as f:
                existing = yaml.safe_load(f) or {}

        existing["rank"] = config
        with open(config_file, "w") as f:
            yaml.safe_dump(existing, f, default_flow_style=False)
        return True
    except (OSError, yaml.YAMLError):
        return False


def add_project_exclusion(project_path: str) -> bool:
    """Add a project to the exclusion list.

    Args:
        project_path: Path or pattern to exclude

    Returns:
        True if added successfully
    """
    config = load_rank_config()
    exclude_list = config.get("exclude_projects", [])

    # Normalize the path
    normalized = os.path.normpath(os.path.expanduser(project_path))

    if normalized not in exclude_list:
        exclude_list.append(normalized)
        config["exclude_projects"] = exclude_list
        return save_rank_config(config)

    return True  # Already excluded


def remove_project_exclusion(project_path: str) -> bool:
    """Remove a project from the exclusion list.

    Args:
        project_path: Path or pattern to remove from exclusion

    Returns:
        True if removed successfully
    """
    config = load_rank_config()
    exclude_list = config.get("exclude_projects", [])

    normalized = os.path.normpath(os.path.expanduser(project_path))

    if normalized in exclude_list:
        exclude_list.remove(normalized)
        config["exclude_projects"] = exclude_list
        return save_rank_config(config)

    return True  # Not in list anyway


def create_global_hook_script() -> str:
    """Generate a global hook script with opt-out support.

    Returns:
        Python script content for the global hook
    """
    return '''#!/usr/bin/env python3
"""MoAI Rank Session Hook (Global)

This hook submits Claude Code session token usage to the MoAI Rank service.
It is installed globally at ~/.claude/hooks/moai/ and runs for all projects.

Opt-out: Configure ~/.moai/rank/config.yaml to exclude specific projects:
    rank:
      enabled: true
      exclude_projects:
        - "/path/to/private-project"
        - "*/confidential/*"
"""

import json
import sys

def main():
    try:
        input_data = sys.stdin.read()
        if not input_data:
            return

        session_data = json.loads(input_data)

        # Lazy import to avoid startup delay
        from moai_adk.rank.hook import is_project_excluded, submit_session_hook

        # Check if this project is excluded
        project_path = session_data.get("projectPath") or session_data.get("cwd")
        if is_project_excluded(project_path):
            return  # Silently skip excluded projects

        result = submit_session_hook(session_data)

        if result["success"]:
            print("Session submitted to MoAI Rank", file=sys.stderr)
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
    """Install the session end hook globally to ~/.claude/hooks/moai/.

    This installs the hook at the user level so it runs for all projects.
    Users can opt-out specific projects via ~/.moai/rank/config.yaml.

    Args:
        project_path: Deprecated, ignored. Hook is always installed globally.

    Returns:
        True if hook was installed successfully
    """
    _ = project_path  # Deprecated parameter, kept for backwards compatibility

    # Install to global user hooks directory
    hooks_dir = Path.home() / ".claude" / "hooks" / "moai"
    hooks_dir.mkdir(parents=True, exist_ok=True)

    hook_file = hooks_dir / "session_end__rank_submit.py"

    try:
        hook_file.write_text(create_global_hook_script())
        hook_file.chmod(0o755)  # Make executable
        return True
    except OSError:
        return False


def uninstall_hook() -> bool:
    """Uninstall the global session end hook.

    Returns:
        True if hook was uninstalled successfully
    """
    hook_file = Path.home() / ".claude" / "hooks" / "moai" / "session_end__rank_submit.py"

    try:
        if hook_file.exists():
            hook_file.unlink()
        return True
    except OSError:
        return False


def is_hook_installed() -> bool:
    """Check if the global hook is installed.

    Returns:
        True if the hook file exists
    """
    hook_file = Path.home() / ".claude" / "hooks" / "moai" / "session_end__rank_submit.py"
    return hook_file.exists()


def prompt_hook_installation(console: Any = None, confirm_func: Any = None) -> bool:
    """Prompt user to install MoAI Rank hook if eligible.

    This function checks if the user is registered with MoAI Rank
    and if the global hook is not yet installed. If both conditions
    are met, it prompts the user to install the hook.

    Args:
        console: Rich console instance for output (optional)
        confirm_func: Confirmation function like click.confirm (optional)

    Returns:
        True if hook was installed, False otherwise
    """
    # Check if user is registered with MoAI Rank
    if not RankConfig.has_credentials():
        return False

    # Check if hook is already installed
    if is_hook_installed():
        return False

    # Lazy import to avoid circular dependencies
    if console is None:
        from rich.console import Console

        console = Console()

    if confirm_func is None:
        import click

        confirm_func = click.confirm

    # Prompt user for hook installation
    console.print()
    console.print("[cyan]🏆 MoAI Rank Hook Installation[/cyan]")
    console.print("[dim]You are registered with MoAI Rank but the session tracking hook is not installed.[/dim]")
    console.print("[dim]The hook automatically submits your Claude Code session stats to the leaderboard.[/dim]")
    console.print()

    if confirm_func("Would you like to install the MoAI Rank session hook?", default=True):
        if install_hook():
            console.print()
            console.print("[green]✅ MoAI Rank hook installed successfully![/green]")
            console.print("[dim]Location: ~/.claude/hooks/moai/session_end__rank_submit.py[/dim]")
            console.print("[dim]To exclude specific projects: moai rank exclude /path/to/project[/dim]")
            console.print()
            return True
        else:
            console.print("[yellow]⚠ Failed to install hook. You can try later with: moai rank register[/yellow]")
            return False
    else:
        console.print("[dim]Skipped. You can install later with: moai rank register[/dim]")
        return False
