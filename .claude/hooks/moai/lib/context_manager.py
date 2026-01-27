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


def save_spec_state(project_root: Path, spec_data: dict[str, Any]) -> bool:
    """Save active SPEC state for session continuity.

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
            "version": CONTEXT_SNAPSHOT_VERSION,
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


def save_tasks_backup(project_root: Path, tasks: list[dict[str, Any]]) -> bool:
    """Save TodoWrite tasks backup for session continuity.

    Args:
        project_root: Project root directory
        tasks: List of task dictionaries with id, subject, status, description

    Returns:
        True if save was successful
    """
    try:
        memory_dir = project_root / ".moai" / "memory"
        memory_dir.mkdir(parents=True, exist_ok=True)

        backup_path = memory_dir / "tasks-backup.json"
        backup = {
            "version": CONTEXT_SNAPSHOT_VERSION,
            "saved_at": datetime.now().isoformat(),
            "tasks": tasks,
        }
        backup_path.write_text(json.dumps(backup, ensure_ascii=False, indent=2), encoding="utf-8")
        return True
    except Exception as e:
        logger.error(f"Failed to save tasks backup: {e}")
        return False


def load_tasks_backup(project_root: Path) -> list[dict[str, Any]]:
    """Load TodoWrite tasks from backup.

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


def generate_memory_mcp_payload(
    project_root: Path,
    context: dict[str, Any],
    summary: str = "",
) -> dict[str, Any]:
    """Generate payload for Memory MCP entity creation.

    This function creates structured data that Alfred can use to save
    session state to Memory MCP. Since hooks cannot call MCP tools directly,
    this generates the payload for Alfred to consume.

    Args:
        project_root: Project root directory
        context: Current context dictionary
        summary: Conversation summary

    Returns:
        Dictionary with entities and relations for Memory MCP
    """
    timestamp = datetime.now().isoformat()
    entities = []
    relations = []

    # SessionState entity
    spec_info = context.get("current_spec", {})
    session_entity = {
        "name": "session_current",
        "entityType": "SessionState",
        "observations": [
            f"timestamp: {timestamp}",
            f"summary: {summary}",
        ],
    }

    if spec_info.get("id"):
        session_entity["observations"].extend(
            [
                f"active_spec: {spec_info.get('id', '')}",
                f"phase: {spec_info.get('phase', '')}",
                f"progress: {spec_info.get('progress_percent', 0)}%",
                f"spec_description: {spec_info.get('description', '')}",
            ]
        )

    # Git state
    branch = context.get("current_branch", "")
    if branch:
        session_entity["observations"].append(f"branch: {branch}")

    if context.get("uncommitted_changes"):
        session_entity["observations"].append("has_uncommitted_changes: true")

    entities.append(session_entity)

    # ActiveTask entities
    tasks = context.get("active_tasks", [])
    for task in tasks[:5]:
        task_entity = {
            "name": f"task_{task.get('id', 'unknown')}",
            "entityType": "ActiveTask",
            "observations": [
                f"subject: {task.get('subject', '')}",
                f"status: {task.get('status', '')}",
                f"timestamp: {timestamp}",
            ],
        }
        if task.get("description"):
            task_entity["observations"].append(f"description: {task.get('description', '')[:200]}")
        entities.append(task_entity)

        relations.append(
            {
                "from": "session_current",
                "to": f"task_{task.get('id', 'unknown')}",
                "relationType": "has_active_task",
            }
        )

    # Recent decisions
    decisions = context.get("key_decisions", [])
    for i, decision in enumerate(decisions[:3]):
        decision_entity = {
            "name": f"decision_recent_{i}",
            "entityType": "UserDecision",
            "observations": [
                f"summary: {decision}",
                f"timestamp: {timestamp}",
            ],
        }
        entities.append(decision_entity)

        relations.append(
            {
                "from": "session_current",
                "to": f"decision_recent_{i}",
                "relationType": "made_decision",
            }
        )

    return {
        "entities": entities,
        "relations": relations,
        "generated_at": timestamp,
    }


def _extract_spec_description(spec_path: Path) -> str:
    """Extract description from SPEC markdown file.

    Looks for the first heading after the frontmatter or the '## 개요' section.

    Args:
        spec_path: Path to spec.md file

    Returns:
        Description string or empty string if not found
    """
    try:
        content = spec_path.read_text(encoding="utf-8")
        lines = content.split("\n")

        # Skip YAML frontmatter
        in_frontmatter = False
        start_idx = 0
        for i, line in enumerate(lines):
            if i == 0 and line.strip() == "---":
                in_frontmatter = True
                continue
            if in_frontmatter and line.strip() == "---":
                start_idx = i + 1
                break

        # Find the first h1 heading (# Title) after frontmatter
        for line in lines[start_idx:]:
            if line.startswith("# "):
                # Extract title without "# " prefix and SPEC ID
                title = line[2:].strip()
                # Remove SPEC ID prefix if present (e.g., "SPEC-XXX: Title" -> "Title")
                if ":" in title:
                    title = title.split(":", 1)[1].strip()
                return title[:100]  # Limit length
        return ""
    except Exception:
        return ""


def _detect_spec_phase(project_root: Path, spec_id: str) -> tuple[str, int]:
    """Detect current phase and progress for a SPEC.

    Checks for existence of plan.md, acceptance.md to determine phase.

    Args:
        project_root: Project root directory
        spec_id: SPEC ID (e.g., "SPEC-XXX")

    Returns:
        Tuple of (phase, progress_percent)
    """
    spec_dir = project_root / ".moai" / "specs" / spec_id

    if not spec_dir.exists():
        return "unknown", 0

    has_spec = (spec_dir / "spec.md").exists()
    has_plan = (spec_dir / "plan.md").exists()
    has_acceptance = (spec_dir / "acceptance.md").exists()

    # Check acceptance.md for completion markers
    if has_acceptance:
        try:
            acceptance_content = (spec_dir / "acceptance.md").read_text(encoding="utf-8").lower()
            # Count checked items [x] vs unchecked [ ]
            checked = acceptance_content.count("[x]")
            unchecked = acceptance_content.count("[ ]")
            total = checked + unchecked

            if total > 0:
                progress = int((checked / total) * 100)
                if progress >= 100:
                    return "sync", 100
                elif progress > 0:
                    return "run", progress
        except Exception:
            pass

    # Determine phase based on file existence
    if has_plan and has_spec:
        return "run", 30  # Default progress for run phase
    elif has_spec:
        return "plan", 50  # Plan phase, spec created
    else:
        return "plan", 10  # Early plan phase

    return "plan", 0


def collect_current_context(project_root: Path) -> dict[str, Any]:
    """Collect current working context from various sources.

    Gathers information from:
    - SPEC documents (including description and phase detection)
    - TodoWrite state
    - Git status
    - Recent file modifications
    - Decisions log

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
                spec_id = specs[0]
                spec_path = project_root / ".moai" / "specs" / spec_id / "spec.md"
                description = _extract_spec_description(spec_path) if spec_path.exists() else ""
                phase, progress = _detect_spec_phase(project_root, spec_id)

                context["current_spec"] = {
                    "id": spec_id,
                    "description": description,
                    "phase": phase,
                    "progress_percent": progress,
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

    try:
        # Get key decisions from decisions log file
        # This file is written by Alfred when important decisions are made
        decisions_file = project_root / ".moai" / "memory" / "decisions.jsonl"
        if decisions_file.exists():
            decisions = []
            for line in decisions_file.read_text(encoding="utf-8").strip().split("\n"):
                if line:
                    try:
                        decision = json.loads(line)
                        # Only include recent decisions (last 5)
                        decisions.append(decision.get("summary", decision.get("choice", "")))
                    except json.JSONDecodeError:
                        continue
            # Keep last 5 decisions, most recent first
            context["key_decisions"] = decisions[-5:][::-1]
    except Exception as e:
        logger.warning(f"Failed to load decisions: {e}")

    return context


def parse_transcript_context(
    transcript_path: Path | str,
    max_lines: int = 1000,
) -> dict[str, Any]:
    """Parse Claude Code transcript JSONL to extract rich session context.

    Reads the last N lines of the transcript to extract:
    - TaskCreate/TaskUpdate calls -> active_tasks
    - Write/Edit file paths -> recent_files
    - AskUserQuestion -> key_decisions
    - SPEC-XXX references -> current_spec

    Args:
        transcript_path: Path to the session transcript JSONL file
        max_lines: Maximum number of lines to read from end of file

    Returns:
        Context dictionary with extracted data
    """
    import re
    from collections import deque

    result: dict[str, Any] = {
        "active_tasks": [],
        "recent_files": [],
        "key_decisions": [],
        "current_spec": {},
    }

    tp = Path(transcript_path).expanduser()
    if not tp.exists():
        return result

    try:
        with open(tp, encoding="utf-8", errors="replace") as f:
            lines = deque(f, maxlen=max_lines)
    except OSError:
        return result

    # Track tasks by ID for final state resolution
    tasks_by_id: dict[str, dict[str, str]] = {}
    edited_files: list[str] = []
    decisions: list[str] = []
    spec_ids: set[str] = set()

    # Pre-filter keywords for fast skip
    tool_keywords = ("TaskCreate", "TaskUpdate", "Write", "Edit", "AskUserQuestion", "SPEC-")

    for raw_line in lines:
        raw_line = raw_line.strip()
        if not raw_line:
            continue

        # Fast pre-filter: skip lines without any relevant keyword
        if not any(kw in raw_line for kw in tool_keywords):
            continue

        try:
            obj = json.loads(raw_line)
        except json.JSONDecodeError:
            continue

        msg = obj.get("message", {})
        content = msg.get("content", "")

        if isinstance(content, list):
            for block in content:
                block_type = block.get("type", "")

                if block_type == "tool_use":
                    name = block.get("name", "")
                    inp = block.get("input", {})

                    if name == "TaskCreate":
                        subject = inp.get("subject", "")
                        if subject:
                            tid = str(len(tasks_by_id) + 1)
                            tasks_by_id[tid] = {
                                "id": tid,
                                "subject": subject,
                                "status": "pending",
                            }

                    elif name == "TaskUpdate":
                        tid = inp.get("taskId", "")
                        status = inp.get("status", "")
                        if tid and tid in tasks_by_id and status:
                            tasks_by_id[tid]["status"] = status

                    elif name in ("Write", "Edit"):
                        fp = inp.get("file_path", "")
                        if fp:
                            # Normalize: strip project root prefix
                            for marker in ("MoAI-ADK/", "MoAI/MoAI-ADK/"):
                                if marker in fp:
                                    fp = fp.split(marker)[-1]
                                    break
                            if fp not in edited_files:
                                edited_files.append(fp)

                    elif name == "AskUserQuestion":
                        questions = inp.get("questions", [])
                        for q in questions:
                            qt = q.get("question", "")
                            if qt:
                                decisions.append(qt[:100])

                if block_type == "text":
                    text = block.get("text", "")
                    if "SPEC-" in text:
                        found = re.findall(r"SPEC-[A-Z]+-\d+", text)
                        spec_ids.update(found)

        elif isinstance(content, str) and "SPEC-" in content:
            found = re.findall(r"SPEC-[A-Z]+-\d+", content)
            spec_ids.update(found)

    # Build result - only active/pending tasks
    active_tasks = [t for t in tasks_by_id.values() if t.get("status") in ("in_progress", "pending")]
    result["active_tasks"] = active_tasks[:10]
    result["recent_files"] = edited_files[:20]
    result["key_decisions"] = decisions[-5:]

    if spec_ids:
        latest_spec = sorted(spec_ids)[-1]
        result["current_spec"] = {"id": latest_spec}

    return result
