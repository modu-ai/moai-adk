#!/usr/bin/env python3
"""Context Manager for Session Continuity

Provides utilities for saving and loading conversation context
to enable seamless session continuation across /clear commands
or new sessions.

Key Features:
- Save context snapshot before /clear or auto compact
- Load previous context on session start
- Archive old snapshots for history
- Integrate with Memory MCP for backup
"""

from __future__ import annotations

import json
import logging
import shutil
from datetime import datetime
from pathlib import Path
from typing import Any

logger = logging.getLogger(__name__)

# Context snapshot version for format compatibility
CONTEXT_SNAPSHOT_VERSION = "1.0.0"

# Maximum number of archived snapshots to keep
MAX_ARCHIVED_SNAPSHOTS = 10


def get_context_snapshot_path(project_root: Path) -> Path:
    """Get the path to the current context snapshot file.

    Args:
        project_root: Project root directory

    Returns:
        Path to context-snapshot.json
    """
    return project_root / ".moai" / "memory" / "context-snapshot.json"


def get_context_archive_dir(project_root: Path) -> Path:
    """Get the path to the context archive directory.

    Args:
        project_root: Project root directory

    Returns:
        Path to context archive directory
    """
    return project_root / ".moai" / "memory" / "context-archive"


def save_context_snapshot(
    project_root: Path,
    trigger: str,
    context: dict[str, Any],
    conversation_summary: str = "",
    session_id: str | None = None,
) -> bool:
    """Save a context snapshot for session continuity.

    Args:
        project_root: Project root directory
        trigger: What triggered the save (pre_compact, session_end, manual)
        context: Dictionary containing:
            - current_spec: SPEC information dict
            - active_tasks: List of TodoWrite tasks
            - recent_files: List of recently modified files
            - key_decisions: List of important decisions made
            - current_branch: Git branch name
            - uncommitted_changes: Boolean
        conversation_summary: Brief summary of the conversation
        session_id: Optional session ID

    Returns:
        True if save was successful
    """
    try:
        # Ensure memory directory exists
        memory_dir = project_root / ".moai" / "memory"
        memory_dir.mkdir(parents=True, exist_ok=True)

        # Build snapshot data
        snapshot = {
            "version": CONTEXT_SNAPSHOT_VERSION,
            "saved_at": datetime.now().isoformat(),
            "trigger": trigger,
            "session_id": session_id or "",
            "context": context,
            "conversation_summary": conversation_summary,
        }

        # Save snapshot
        snapshot_path = get_context_snapshot_path(project_root)
        snapshot_path.write_text(json.dumps(snapshot, ensure_ascii=False, indent=2), encoding="utf-8")

        logger.info(f"Context snapshot saved: {snapshot_path}")
        return True

    except Exception as e:
        logger.error(f"Failed to save context snapshot: {e}")
        return False


def load_context_snapshot(project_root: Path) -> dict[str, Any] | None:
    """Load the most recent context snapshot.

    Args:
        project_root: Project root directory

    Returns:
        Snapshot data dictionary or None if not found/invalid
    """
    try:
        snapshot_path = get_context_snapshot_path(project_root)

        if not snapshot_path.exists():
            return None

        snapshot = json.loads(snapshot_path.read_text(encoding="utf-8"))

        # Validate version
        if snapshot.get("version") != CONTEXT_SNAPSHOT_VERSION:
            logger.warning(f"Context snapshot version mismatch: {snapshot.get('version')}")
            # Still return it, but log the warning

        return snapshot

    except json.JSONDecodeError as e:
        logger.error(f"Invalid JSON in context snapshot: {e}")
        return None
    except Exception as e:
        logger.error(f"Failed to load context snapshot: {e}")
        return None


def archive_context_snapshot(project_root: Path) -> bool:
    """Archive the current context snapshot.

    Moves the current snapshot to the archive directory with timestamp.
    Maintains MAX_ARCHIVED_SNAPSHOTS limit.

    Args:
        project_root: Project root directory

    Returns:
        True if archive was successful or no snapshot to archive
    """
    try:
        snapshot_path = get_context_snapshot_path(project_root)

        if not snapshot_path.exists():
            return True  # Nothing to archive

        # Ensure archive directory exists
        archive_dir = get_context_archive_dir(project_root)
        archive_dir.mkdir(parents=True, exist_ok=True)

        # Generate archive filename with timestamp
        timestamp = datetime.now().strftime("%Y%m%d-%H%M%S")
        archive_path = archive_dir / f"context-{timestamp}.json"

        # Move snapshot to archive
        shutil.move(str(snapshot_path), str(archive_path))
        logger.info(f"Context snapshot archived: {archive_path}")

        # Clean up old archives
        _cleanup_old_archives(archive_dir)

        return True

    except Exception as e:
        logger.error(f"Failed to archive context snapshot: {e}")
        return False


def _cleanup_old_archives(archive_dir: Path) -> None:
    """Clean up old archived snapshots, keeping only the most recent.

    Args:
        archive_dir: Archive directory path
    """
    try:
        archives = sorted(archive_dir.glob("context-*.json"), key=lambda p: p.stat().st_mtime, reverse=True)

        # Remove old archives beyond the limit
        for old_archive in archives[MAX_ARCHIVED_SNAPSHOTS:]:
            old_archive.unlink()
            logger.debug(f"Removed old archive: {old_archive}")

    except Exception as e:
        logger.warning(f"Failed to cleanup old archives: {e}")


def format_context_for_injection(snapshot: dict[str, Any], language: str = "en") -> str:
    """Format context snapshot for injection into systemMessage.

    Args:
        snapshot: Context snapshot dictionary
        language: User's conversation language (ko, en, ja, zh)

    Returns:
        Formatted string for systemMessage injection
    """
    context = snapshot.get("context", {})
    summary = snapshot.get("conversation_summary", "")
    saved_at = snapshot.get("saved_at", "")

    # Multilingual labels
    labels = {
        "ko": {
            "header": "이전 세션 컨텍스트",
            "spec": "SPEC",
            "phase": "단계",
            "progress": "진행률",
            "current_task": "현재 작업",
            "pending_tasks": "대기 작업",
            "recent_files": "최근 파일",
            "decisions": "주요 결정",
            "branch": "브랜치",
            "uncommitted": "미커밋 변경",
            "summary": "작업 요약",
            "saved_at": "저장 시간",
            "continue_prompt": "이전 세션을 이어서 진행하시겠습니까?",
            "yes": "예",
            "no": "아니오",
        },
        "ja": {
            "header": "前回セッションコンテキスト",
            "spec": "SPEC",
            "phase": "フェーズ",
            "progress": "進捗",
            "current_task": "現在のタスク",
            "pending_tasks": "保留中のタスク",
            "recent_files": "最近のファイル",
            "decisions": "主な決定",
            "branch": "ブランチ",
            "uncommitted": "未コミットの変更",
            "summary": "作業概要",
            "saved_at": "保存時刻",
            "continue_prompt": "前回のセッションを続けますか？",
            "yes": "はい",
            "no": "いいえ",
        },
        "zh": {
            "header": "上次会话上下文",
            "spec": "SPEC",
            "phase": "阶段",
            "progress": "进度",
            "current_task": "当前任务",
            "pending_tasks": "待处理任务",
            "recent_files": "最近文件",
            "decisions": "关键决策",
            "branch": "分支",
            "uncommitted": "未提交更改",
            "summary": "工作摘要",
            "saved_at": "保存时间",
            "continue_prompt": "是否继续上次会话？",
            "yes": "是",
            "no": "否",
        },
        "en": {
            "header": "Previous Session Context",
            "spec": "SPEC",
            "phase": "Phase",
            "progress": "Progress",
            "current_task": "Current Task",
            "pending_tasks": "Pending Tasks",
            "recent_files": "Recent Files",
            "decisions": "Key Decisions",
            "branch": "Branch",
            "uncommitted": "Uncommitted Changes",
            "summary": "Work Summary",
            "saved_at": "Saved At",
            "continue_prompt": "Would you like to continue from the previous session?",
            "yes": "Yes",
            "no": "No",
        },
    }

    # Get labels for the specified language, fallback to English
    lbl = labels.get(language, labels["en"])

    lines = [f"\n📋 [{lbl['header']}]"]

    # SPEC information
    spec_info = context.get("current_spec", {})
    if spec_info:
        spec_id = spec_info.get("id", "")
        spec_desc = spec_info.get("description", "")
        phase = spec_info.get("phase", "")
        progress = spec_info.get("progress_percent", 0)

        if spec_id:
            lines.append(f"   - {lbl['spec']}: {spec_id} ({spec_desc})")
        if phase:
            lines.append(f"   - {lbl['phase']}: {phase}")
        if progress:
            lines.append(f"   - {lbl['progress']}: {progress}%")

    # Active tasks
    tasks = context.get("active_tasks", [])
    if tasks:
        in_progress = [t for t in tasks if t.get("status") == "in_progress"]
        pending = [t for t in tasks if t.get("status") == "pending"]

        if in_progress:
            current = in_progress[0].get("subject", "")
            lines.append(f"   - {lbl['current_task']}: {current}")

        if pending:
            pending_subjects = [t.get("subject", "") for t in pending[:3]]
            lines.append(f"   - {lbl['pending_tasks']}: {', '.join(pending_subjects)}")

    # Recent files
    recent_files = context.get("recent_files", [])
    if recent_files:
        files_display = ", ".join(recent_files[:5])
        lines.append(f"   - {lbl['recent_files']}: {files_display}")

    # Key decisions
    decisions = context.get("key_decisions", [])
    if decisions:
        for decision in decisions[:3]:
            lines.append(f"   - {lbl['decisions']}: {decision}")

    # Git information
    branch = context.get("current_branch", "")
    if branch:
        lines.append(f"   - {lbl['branch']}: {branch}")

    uncommitted = context.get("uncommitted_changes", False)
    if uncommitted:
        lines.append(f"   - {lbl['uncommitted']}: ⚠️ {lbl['yes']}")

    # Summary
    if summary:
        lines.append(f"   - {lbl['summary']}: {summary}")

    # Saved time
    if saved_at:
        lines.append(f"   - {lbl['saved_at']}: {saved_at}")

    # Continue prompt
    lines.append(f"\n{lbl['continue_prompt']}")

    return "\n".join(lines)


def collect_current_context(project_root: Path) -> dict[str, Any]:
    """Collect current working context from various sources.

    Gathers information from:
    - SPEC documents
    - TodoWrite state
    - Git status
    - Recent file modifications

    Args:
        project_root: Project root directory

    Returns:
        Context dictionary
    """
    context: dict[str, Any] = {
        "current_spec": {},
        "active_tasks": [],
        "recent_files": [],
        "key_decisions": [],
        "current_branch": "",
        "uncommitted_changes": False,
    }

    try:
        # Get current SPEC from last-session-state
        state_file = project_root / ".moai" / "memory" / "last-session-state.json"
        if state_file.exists():
            state = json.loads(state_file.read_text(encoding="utf-8"))
            specs = state.get("specs_in_progress", [])
            if specs:
                context["current_spec"] = {
                    "id": specs[0],
                    "description": "",
                    "phase": "run",
                    "progress_percent": 0,
                }
            context["current_branch"] = state.get("current_branch", "")
            context["uncommitted_changes"] = bool(state.get("uncommitted_files", 0))
    except Exception as e:
        logger.warning(f"Failed to load session state: {e}")

    try:
        # Get TodoWrite tasks
        todo_file = project_root / ".moai" / "memory" / "todo-state.json"
        if todo_file.exists():
            todo_data = json.loads(todo_file.read_text(encoding="utf-8"))
            tasks = todo_data.get("tasks", [])
            context["active_tasks"] = [
                {"id": t.get("id"), "subject": t.get("subject"), "status": t.get("status")}
                for t in tasks
                if t.get("status") in ("in_progress", "pending")
            ][:10]  # Limit to 10 tasks
    except Exception as e:
        logger.warning(f"Failed to load todo state: {e}")

    try:
        # Get recent files from git status
        import subprocess

        result = subprocess.run(
            ["git", "status", "--porcelain"], capture_output=True, text=True, timeout=3, cwd=str(project_root)
        )
        if result.returncode == 0:
            files = []
            for line in result.stdout.strip().split("\n"):
                if line:
                    # Extract filename from git status output
                    parts = line.split()
                    if len(parts) >= 2:
                        files.append(parts[-1])
            context["recent_files"] = files[:10]  # Limit to 10 files
    except Exception as e:
        logger.warning(f"Failed to get recent files: {e}")

    return context
