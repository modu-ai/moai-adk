#!/usr/bin/env python3
"""Event-Driven Checkpoint system

Detect risky tasks and create automatic checkpoints
"""

import json
import subprocess
from datetime import datetime
from pathlib import Path
from typing import Any


def detect_risky_operation(tool_name: str, tool_args: dict[str, Any], cwd: str) -> tuple[bool, str]:
    """Risk task detection (for Event-Driven Checkpoint)

    Claude Code tool automatically detects dangerous tasks before use.
    When a risk is detected, a checkpoint is automatically created to enable rollback.

    Args:
        tool_name: Name of the Claude Code tool (Bash, Edit, Write, MultiEdit)
        tool_args: Tool argument dictionary
        cwd: Project root directory path

    Returns:
        (is_risky, operation_type) tuple
        - is_risky: Whether the operation is dangerous (bool)
        - operation_type: operation type (str: delete, merge, script, critical-file, refactor)

    Risky Operations:
        - Bash tool: rm -rf, git merge, git reset --hard, git rebase, script execution
        - Edit/Write tool: CLAUDE.md, config.json, .claude/skills/*.md
        - MultiEdit tool: Edit ≥10 items File simultaneously
        - Script execution: Python, Node, Java, Go, Rust, Dart, Swift, Kotlin, Shell scripts

    Examples:
        >>> detect_risky_operation("Bash", {"command": "rm -rf src/"}, ".")
        (True, 'delete')
        >>> detect_risky_operation("Edit", {"file_path": "CLAUDE.md"}, ".")
        (True, 'critical-file')
        >>> detect_risky_operation("Read", {"file_path": "test.py"}, ".")
        (False, '')

    Notes:
        - Minimize false positives: ignore safe operations
        - Performance: lightweight string matching (< 1ms)
        - Extensibility: Easily added to the patterns dictionary

    """
    # Bash tool: Detect dangerous commands
    if tool_name == "Bash":
        command = tool_args.get("command", "")

        # Mass Delete
        if any(pattern in command for pattern in ["rm -rf", "git rm"]):
            return (True, "delete")

        # Git merge/reset/rebase
        if any(pattern in command for pattern in ["git merge", "git reset --hard", "git rebase"]):
            return (True, "merge")

        # Execute external script (potentially destructive)
        if any(command.startswith(prefix) for prefix in ["python ", "node ", "bash ", "sh "]):
            return (True, "script")

    # Edit/Write tool: Detect important files
    if tool_name in ("Edit", "Write"):
        file_path = tool_args.get("file_path", "")

        critical_files = [
            "CLAUDE.md",
            "config.json",
            ".claude/skills/moai-core-dev-guide/reference.md",
            ".claude/skills/moai-core-spec-metadata-extended/reference.md",
            ".moai/config/config.json",
        ]

        if any(cf in file_path for cf in critical_files):
            return (True, "critical-file")

    # MultiEdit tool: Detect large edits
    if tool_name == "MultiEdit":
        edits = tool_args.get("edits", [])
        if len(edits) >= 10:
            return (True, "refactor")

    return (False, "")


def create_checkpoint(cwd: str, operation_type: str) -> str:
    """Create checkpoint (Git local branch)

    Automatically creates checkpoints before dangerous operations.
    Prevent remote repository contamination by creating a Git local branch.

    Args:
        cwd: Project root directory path
        operation_type: operation type (delete, merge, script, etc.)

    Returns:
        checkpoint_branch: Created branch name
        Returns "checkpoint-failed" on failure

    Branch Naming:
        before-{operation}-{YYYYMMDD-HHMMSS}
        Example: before-delete-20251015-143000

    Examples:
        >>> create_checkpoint(".", "delete")
        'before-delete-20251015-143000'

    Notes:
        - Create only local branch (no remote push)
        - Fallback in case of Git error (ignore and continue)
        - Do not check dirty working directory (allow uncommitted changes)
        - Automatically record checkpoint logs (.moai/checkpoints.log)

    """
    timestamp = datetime.now().strftime("%Y%m%d-%H%M%S")
    branch_name = f"before-{operation_type}-{timestamp}"

    try:
        # Create a new local branch from the current branch (without checking out)
        subprocess.run(
            ["git", "branch", branch_name],
            cwd=cwd,
            check=True,
            capture_output=True,
            text=True,
            timeout=2,
        )

        # Checkpoint log records
        log_checkpoint(cwd, branch_name, operation_type)

        return branch_name

    except (
        subprocess.CalledProcessError,
        subprocess.TimeoutExpired,
        FileNotFoundError,
    ):
        # Fallback (ignore) in case of Git error
        return "checkpoint-failed"


def log_checkpoint(cwd: str, branch_name: str, operation_type: str) -> None:
    """Checkpoint log records (.moai/checkpoints.log)

    Checkpoint creation history is recorded in JSON Lines format.
    SessionStart reads this log to display a list of checkpoints.

    Args:
        cwd: Project root directory path
        branch_name: Created checkpoint branch name
        operation_type: operation type

    Log Format (JSON Lines):
        {"timestamp": "2025-10-15T14:30:00", "branch": "before-delete-...", "operation": "delete"}

    Examples:
        >>> log_checkpoint(".", "before-delete-20251015-143000", "delete")
        # Add 1 line to .moai/checkpoints.log

    Notes:
        - If the file does not exist, it is automatically created.
        - Record in append mode (preserve existing logs)
        - Ignored in case of failure (not critical)

    """
    log_file = Path(cwd) / ".moai" / "checkpoints.log"

    try:
        log_file.parent.mkdir(parents=True, exist_ok=True)

        log_entry = {
            "timestamp": datetime.now().isoformat(),
            "branch": branch_name,
            "operation": operation_type,
        }

        with log_file.open("a") as f:
            f.write(json.dumps(log_entry) + "\n")

    except (OSError, PermissionError):
        # Ignore log failures (not critical)
        pass


def list_checkpoints(cwd: str, max_count: int = 10) -> list[dict[str, str]]:
    """Checkpoint list (parsing .moai/checkpoints.log)

    Returns a list of recently created checkpoints.
    Used in the SessionStart, /moai:0-project restore command.

    Args:
        cwd: Project root directory path
        max_count: Maximum number to return (default 10 items)

    Returns:
        Checkpoint list (most recent)
        [{"timestamp": "...", "branch": "...", "operation": "..."}, ...]

    Examples:
        >>> list_checkpoints(".")
        [
            {"timestamp": "2025-10-15T14:30:00", "branch": "before-delete-...", "operation": "delete"},
            {"timestamp": "2025-10-15T14:25:00", "branch": "before-merge-...", "operation": "merge"},
        ]

    Notes:
        - If there is no log file, an empty list is returned.
        - Ignore lines where JSON parsing fails
        - Return only the latest max_count

    """
    log_file = Path(cwd) / ".moai" / "checkpoints.log"

    if not log_file.exists():
        return []

    checkpoints = []

    try:
        with log_file.open("r") as f:
            for line in f:
                try:
                    checkpoints.append(json.loads(line.strip()))
                except json.JSONDecodeError:
                    # Ignore lines where parsing failed
                    pass
    except (OSError, PermissionError):
        return []

    # Return only the most recent max_count items (in order of latest)
    return checkpoints[-max_count:]


__all__ = [
    "detect_risky_operation",
    "create_checkpoint",
    "log_checkpoint",
    "list_checkpoints",
]
