#!/usr/bin/env python3
"""Session event handlers

SessionStart, SessionEnd event handling
"""

from core import HookPayload, HookResult
from core.checkpoint import list_checkpoints
from core.project import count_specs, detect_language, get_git_info


def handle_session_start(payload: HookPayload) -> HookResult:
    """SessionStart event handler (with Checkpoint list)

    When Claude Code Session starts, it displays a summary of project status.
    You can check the language, Git status, SPEC progress, and checkpoint list at a glance.

    Args:
        payload: Claude Code event payload (cwd key required)

    Returns:
        HookResult(message=project status summary message, systemMessage=for user display)

    Message Format:
        🚀 MoAI-ADK Session Started
           Language: {language}
           Branch: {branch} ({commit hash})
           Changes: {Number of Changed Files}
           SPEC Progress: {Complete}/{Total} ({percent}%)
           Checkpoints: {number} available (showing the latest 3 items)

    Note:
        - Claude Code processes SessionStart in several stages (clear → compact)
        - Display message only at “compact” stage to prevent duplicate output
        - "clear" step returns empty result (invisible to user)

    TDD History:
        - RED: Session startup message format test
        - GREEN: Generate status message by combining helper functions
        - REFACTOR: Improved message format, improved readability, added checkpoint list
        - FIX: Prevent duplicate output of clear step (only compact step is displayed)

    @TAG:CHECKPOINT-EVENT-001
    """
    # Claude Code SessionStart runs in several stages (clear, compact, etc.)
    # Ignore the "clear" stage and output messages only at the "compact" stage
    event_phase = payload.get("phase", "")
    if event_phase == "clear":
        return HookResult() # returns an empty result (prevents duplicate output)

    cwd = payload.get("cwd", ".")
    language = detect_language(cwd)
    git_info = get_git_info(cwd)
    specs = count_specs(cwd)
    checkpoints = list_checkpoints(cwd, max_count=10)

    branch = git_info.get("branch", "N/A")
    commit = git_info.get("commit", "N/A")[:7]
    changes = git_info.get("changes", 0)
    spec_progress = f"{specs['completed']}/{specs['total']}"

    # systemMessage: displayed directly to the user
    lines = [
        "🚀 MoAI-ADK Session Started",
        f"   Language: {language}",
        f"   Branch: {branch} ({commit})",
        f"   Changes: {changes}",
        f"   SPEC Progress: {spec_progress} ({specs['percentage']}%)",
    ]

    # Add Checkpoint list (show only the latest 3 items)
    if checkpoints:
        lines.append(f"   Checkpoints: {len(checkpoints)} available")
        for cp in reversed(checkpoints[-3:]):  # Latest 3 items
            branch_short = cp["branch"].replace("before-", "")
            lines.append(f"      - {branch_short}")
        lines.append("   Restore: /alfred:0-project restore")

    system_message = "\n".join(lines)

    return HookResult(
        message=system_message, # for Claude context
        systemMessage=system_message, # For user display
    )


def handle_session_end(payload: HookPayload) -> HookResult:
    """SessionEnd event handler (default implementation)"""
    return HookResult()


__all__ = ["handle_session_start", "handle_session_end"]
