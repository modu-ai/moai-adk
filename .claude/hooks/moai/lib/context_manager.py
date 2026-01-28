#!/usr/bin/env python3
"""Context Manager for MoAI

Provides utilities for SPEC state, task backup, and decision tracking.
"""

from __future__ import annotations

import json
import logging
from datetime import datetime
from pathlib import Path
from typing import Any

logger = logging.getLogger(__name__)


def save_spec_state(project_root: Path, spec_data: dict[str, Any]) -> bool:
    """Save active SPEC state.

    Args:
        project_root: Project root directory
        spec_data: Dictionary with spec_id, phase, progress, description

    Returns:
        True if save was successful
    """
    try:
        memory_dir = project_root / ".moai" / "memory"
        memory_dir.mkdir(parents=True, exist_ok=True)

        state_path = memory_dir / "spec-state.json"
        state = {
            "version": "1.1.0",
            "updated_at": datetime.now().isoformat(),
            "active_spec": spec_data,
        }
        state_path.write_text(json.dumps(state, ensure_ascii=False, indent=2), encoding="utf-8")
        return True
    except Exception as e:
        logger.error(f"Failed to save spec state: {e}")
        return False


def load_spec_state(project_root: Path) -> dict[str, Any] | None:
    """Load active SPEC state.

    Args:
        project_root: Project root directory

    Returns:
        SPEC state dictionary or None
    """
    try:
        state_path = project_root / ".moai" / "memory" / "spec-state.json"
        if not state_path.exists():
            return None
        return json.loads(state_path.read_text(encoding="utf-8"))
    except Exception as e:
        logger.error(f"Failed to load spec state: {e}")
        return None


def save_tasks_backup(
    project_root: Path,
    tasks: list[dict[str, Any]],
    completed_tasks: list[dict[str, Any]] | None = None,
) -> bool:
    """Save task backup for project tracking.

    Args:
        project_root: Project root directory
        tasks: List of active task dictionaries
        completed_tasks: List of completed task dictionaries

    Returns:
        True if save was successful
    """
    try:
        memory_dir = project_root / ".moai" / "memory"
        memory_dir.mkdir(parents=True, exist_ok=True)

        backup_path = memory_dir / "tasks-backup.json"
        backup = {
            "version": "1.1.0",
            "saved_at": datetime.now().isoformat(),
            "tasks": tasks,
            "completed_tasks": completed_tasks or [],
        }
        backup_path.write_text(json.dumps(backup, ensure_ascii=False, indent=2), encoding="utf-8")
        return True
    except Exception as e:
        logger.error(f"Failed to save tasks backup: {e}")
        return False


def load_tasks_backup(project_root: Path) -> list[dict[str, Any]]:
    """Load tasks from backup.

    Args:
        project_root: Project root directory

    Returns:
        List of task dictionaries or empty list
    """
    try:
        backup_path = project_root / ".moai" / "memory" / "tasks-backup.json"
        if not backup_path.exists():
            return []
        data = json.loads(backup_path.read_text(encoding="utf-8"))
        return data.get("tasks", [])
    except Exception as e:
        logger.error(f"Failed to load tasks backup: {e}")
        return []


def append_decision(project_root: Path, decision: dict[str, Any]) -> bool:
    """Append a user decision to the decisions log.

    Args:
        project_root: Project root directory
        decision: Dictionary with question, choice, context, timestamp

    Returns:
        True if append was successful
    """
    try:
        memory_dir = project_root / ".moai" / "memory"
        memory_dir.mkdir(parents=True, exist_ok=True)

        decisions_path = memory_dir / "decisions.jsonl"
        entry = {
            "timestamp": datetime.now().isoformat(),
            **decision,
        }

        with open(decisions_path, "a", encoding="utf-8") as f:
            f.write(json.dumps(entry, ensure_ascii=False) + "\n")

        return True
    except Exception as e:
        logger.error(f"Failed to append decision: {e}")
        return False


def load_recent_decisions(project_root: Path, limit: int = 10) -> list[dict[str, Any]]:
    """Load recent decisions from the decisions log.

    Args:
        project_root: Project root directory
        limit: Maximum number of recent decisions to return

    Returns:
        List of decision dictionaries (most recent first)
    """
    try:
        decisions_path = project_root / ".moai" / "memory" / "decisions.jsonl"
        if not decisions_path.exists():
            return []

        decisions = []
        for line in decisions_path.read_text(encoding="utf-8").strip().split("\n"):
            if line.strip():
                try:
                    decisions.append(json.loads(line))
                except json.JSONDecodeError:
                    continue

        return decisions[-limit:][::-1]
    except Exception as e:
        logger.error(f"Failed to load decisions: {e}")
        return []


def collect_current_context(project_root: Path) -> dict[str, Any]:
    """Collect current working context from various sources.

    Args:
        project_root: Project root directory

    Returns:
        Context dictionary
    """
    context: dict[str, Any] = {
        "current_spec": {},
        "active_tasks": [],
        "completed_tasks": [],
        "recent_files": [],
        "key_decisions": [],
        "current_branch": "",
        "uncommitted_changes": False,
    }

    try:
        # Get current SPEC from spec-state
        state_file = project_root / ".moai" / "memory" / "spec-state.json"
        if state_file.exists():
            state = json.loads(state_file.read_text(encoding="utf-8"))
            spec = state.get("active_spec", {})
            if spec:
                context["current_spec"] = spec

        # Get recent files from git status
        import subprocess

        result = subprocess.run(
            ["git", "status", "--porcelain"], capture_output=True, text=True, timeout=3, cwd=str(project_root)
        )
        if result.returncode == 0:
            files = []
            for line in result.stdout.strip().split("\n"):
                if line:
                    parts = line.split()
                    if len(parts) >= 2:
                        files.append(parts[-1])
            context["recent_files"] = files[:10]

    except Exception as e:
        logger.warning(f"Failed to collect current context: {e}")

    return context
