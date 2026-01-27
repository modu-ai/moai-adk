#!/usr/bin/env python3
"""PreCompact Hook: Save Context Before Clear or Auto Compact

Claude Code Event: PreCompact
Purpose: Save conversation context before /clear command or auto compact
Execution: Triggered automatically before context compaction

This hook enables seamless session continuation by:
1. Collecting current working context (SPEC, tasks, files, decisions)
2. Saving context snapshot to .moai/memory/context-snapshot.json
3. Optionally backing up to Memory MCP

The saved context can be loaded by SessionStart hook to resume work.
"""

from __future__ import annotations

import json
import logging
import sys
from pathlib import Path
from typing import Any

# =============================================================================
# Windows UTF-8 Encoding Fix (Issue #249)
# =============================================================================
if sys.platform == "win32":
    try:
        sys.stdout.reconfigure(encoding="utf-8", errors="replace")
        sys.stderr.reconfigure(encoding="utf-8", errors="replace")
    except (AttributeError, OSError):
        pass

# =============================================================================
# Setup import path for shared modules
# =============================================================================
HOOKS_DIR = Path(__file__).parent
LIB_DIR = HOOKS_DIR / "lib"
if str(LIB_DIR) not in sys.path:
    sys.path.insert(0, str(LIB_DIR))

from lib.context_manager import (  # noqa: E402
    collect_current_context,
    generate_memory_mcp_payload,
    parse_transcript_context,
    save_context_snapshot,
    save_spec_state,
    save_tasks_backup,
)
from lib.path_utils import find_project_root  # noqa: E402

# Import timeout manager
try:
    from lib.unified_timeout_manager import (
        HookTimeoutConfig,
        HookTimeoutError,
        TimeoutPolicy,
        get_timeout_manager,
    )
except ImportError:

    def get_timeout_manager():
        return None

    class HookTimeoutConfig:
        def __init__(self, **kwargs):
            pass

    class TimeoutPolicy:
        FAST = "fast"
        NORMAL = "normal"
        SLOW = "slow"

    class HookTimeoutError(Exception):
        pass


logger = logging.getLogger(__name__)


def generate_conversation_summary(context: dict[str, Any]) -> str:
    """Generate a brief summary of the current conversation context.

    Args:
        context: Current working context dictionary

    Returns:
        Brief summary string
    """
    parts = []

    # SPEC information
    spec = context.get("current_spec", {})
    if spec.get("id"):
        spec_id = spec.get("id", "")
        spec_desc = spec.get("description", "")
        phase = spec.get("phase", "")
        progress = spec.get("progress_percent", 0)

        if spec_desc:
            parts.append(f"{spec_id}: {spec_desc}")
        else:
            parts.append(spec_id)

        if phase and progress:
            parts.append(f"{phase} phase ({progress}% complete)")

    # Active tasks
    tasks = context.get("active_tasks", [])
    in_progress = [t for t in tasks if t.get("status") == "in_progress"]
    if in_progress:
        current_task = in_progress[0].get("subject", "")
        if current_task:
            parts.append(f"Working on: {current_task}")

    # Key decisions
    decisions = context.get("key_decisions", [])
    if decisions:
        parts.append(f"Decisions made: {len(decisions)}")

    # Uncommitted changes
    if context.get("uncommitted_changes"):
        parts.append("Has uncommitted changes")

    return ". ".join(parts) if parts else "Session context saved"


def execute_pre_compact():
    """Execute the pre-compact context save workflow."""
    # Read hook payload from stdin
    input_data = sys.stdin.read() if not sys.stdin.isatty() else "{}"
    payload = json.loads(input_data) if input_data.strip() else {}

    # Get session ID and transcript path from Claude Code payload
    session_id = payload.get("session_id", "")
    transcript_path = payload.get("transcript_path", "")

    # Find project root
    project_root = find_project_root()

    # Collect file-based context (git status, recent files)
    context = collect_current_context(project_root)

    # Enrich with transcript data (tasks, decisions, SPEC refs, edited files)
    if transcript_path:
        transcript_data = parse_transcript_context(transcript_path)
        # Merge: transcript data takes priority over empty file-based data
        if transcript_data.get("active_tasks"):
            context["active_tasks"] = transcript_data["active_tasks"]
        if transcript_data.get("recent_files"):
            context["recent_files"] = transcript_data["recent_files"]
        if transcript_data.get("key_decisions"):
            context["key_decisions"] = transcript_data["key_decisions"]
        if transcript_data.get("current_spec", {}).get("id"):
            # Only override if file-based context has no SPEC
            if not context.get("current_spec", {}).get("id"):
                context["current_spec"] = transcript_data["current_spec"]

    # Generate conversation summary
    summary = generate_conversation_summary(context)

    # Save context snapshot
    success = save_context_snapshot(
        project_root=project_root,
        trigger="pre_compact",
        context=context,
        conversation_summary=summary,
        session_id=session_id,
    )

    # Save SPEC state separately for quick access
    spec_info = context.get("current_spec", {})
    if spec_info.get("id"):
        save_spec_state(project_root, spec_info)

    # Save tasks backup separately for quick restoration
    tasks = context.get("active_tasks", [])
    if tasks:
        save_tasks_backup(project_root, tasks)

    # Generate Memory MCP payload for Alfred to consume on next session
    mcp_payload = generate_memory_mcp_payload(project_root, context, summary)
    try:
        mcp_payload_path = project_root / ".moai" / "memory" / "mcp-payload.json"
        mcp_payload_path.write_text(json.dumps(mcp_payload, ensure_ascii=False, indent=2), encoding="utf-8")
    except Exception:
        pass  # Non-critical

    # Build result - NOTE: hookSpecificOutput not allowed for PreCompact event
    # Only continue and systemMessage are valid for PreCompact hooks
    snapshot_path = str(project_root / ".moai" / "memory" / "context-snapshot.json")

    if success:
        system_message = f"💾 Context saved to {snapshot_path}. You can continue from this point in a new session."
    else:
        system_message = "⚠️ Failed to save context. Some work state may be lost."

    result: dict[str, Any] = {
        "continue": True,  # Allow compact to proceed
        "systemMessage": system_message,
    }

    return result


def main() -> None:
    """Main entry point for PreCompact hook.

    Saves conversation context before /clear or auto compact to enable
    seamless session continuation.

    Exit Codes:
        0: Success
        1: Error (timeout, exception)
    """
    # Configure timeout
    timeout_config = HookTimeoutConfig(
        policy=TimeoutPolicy.FAST,
        custom_timeout_ms=3000,  # 3 seconds - quick save
        retry_count=0,
        graceful_degradation=True,
        memory_limit_mb=50,
    )

    # Use timeout manager if available
    timeout_manager = get_timeout_manager()
    if timeout_manager:
        try:
            result = timeout_manager.execute_with_timeout(
                "pre_compact__save_context",
                execute_pre_compact,
                config=timeout_config,
            )
            print(json.dumps(result, ensure_ascii=False))
            sys.exit(0)

        except HookTimeoutError as e:
            timeout_response = {
                "continue": True,
                "systemMessage": "⚠️ Context save timeout - proceeding without save",
                "error_details": {
                    "hook_id": e.hook_id,
                    "timeout_seconds": e.timeout_seconds,
                },
            }
            print(json.dumps(timeout_response, ensure_ascii=False))
            sys.exit(1)

        except Exception as e:
            error_response = {
                "continue": True,
                "systemMessage": f"⚠️ Context save error: {e}",
                "error_details": {
                    "error_type": type(e).__name__,
                    "message": str(e),
                },
            }
            print(json.dumps(error_response, ensure_ascii=False))
            sys.exit(1)

    else:
        # Direct execution without timeout manager
        try:
            result = execute_pre_compact()
            print(json.dumps(result, ensure_ascii=False))
            sys.exit(0)

        except Exception as e:
            error_response = {
                "continue": True,
                "systemMessage": f"⚠️ Context save error: {e}",
            }
            print(json.dumps(error_response, ensure_ascii=False))
            sys.exit(1)


if __name__ == "__main__":
    main()
