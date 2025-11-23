#!/usr/bin/env python3
"""Session event handlers

SessionStart, SessionEnd event handling
"""


from lib import HookPayload, HookResult
from lib.checkpoint import list_checkpoints
from lib.project import count_specs, get_git_info, get_package_version_info


def handle_session_start(payload: HookPayload) -> HookResult:
    """SessionStart event handler with GRACEFUL DEGRADATION

    When Claude Code Session starts, it displays a summary of project status.
    You can check the language, Git status, SPEC progress, and checkpoint list at a glance.
    All optional operations are wrapped in try-except to ensure hook completes quickly even if
    Git commands, file I/O, or other operations timeout or fail.

    Args:
        payload: Claude Code event payload (cwd key required)

    Returns:
        HookResult(system_message=project status summary message)

    Message Format:
        🚀 MoAI-ADK Session Started
           [Version: {version}] - optional if version check fails
           [Branch: {branch} ({commit hash})] - optional if git fails
           [Changes: {Number of Changed Files}] - optional if git fails
           [SPEC Progress: {Complete}/{Total} ({percent}%)] - optional if specs fail
           [Checkpoints: {number} available] - optional if checkpoint list fails

    Graceful Degradation Strategy:
        - OPTIONAL: Version info (skip if timeout/failure)
        - OPTIONAL: Git info (skip if timeout/failure)
        - OPTIONAL: SPEC progress (skip if timeout/failure)
        - OPTIONAL: Checkpoint list (skip if timeout/failure)
        - Always display SOMETHING to user, never return empty message

    Note:
        - Claude Code processes SessionStart in several stages (clear → compact)
        - Display message only at "compact" stage to prevent duplicate output
        - "clear" step returns minimal result (empty hookSpecificOutput)
        - CRITICAL: All optional operations must complete within 2-3 seconds total

    TDD History:
        - RED: Session startup message format test
        - GREEN: Generate status message by combining helper functions
        - REFACTOR: Improved message format, improved readability, added checkpoint list
        - FIX: Prevent duplicate output of clear step (only compact step is displayed)
        - UPDATE: Migrated to Claude Code standard Hook schema
        - HOTFIX: Add graceful degradation for timeout scenarios (Issue #66)

    """
    # Claude Code SessionStart runs in several stages (clear, compact, etc.)
    # Ignore the "clear" stage and output messages only at the "compact" stage
    event_phase = payload.get("phase", "")
    if event_phase == "clear":
        # Return minimal valid Hook result for clear phase
        return HookResult(continue_execution=True)

    cwd = payload.get("cwd", ".")

    # OPTIONAL: Git info - skip if timeout/failure
    git_info = {}
    try:
        git_info = get_git_info(cwd)
    except Exception:
        # Graceful degradation - continue without git info
        pass

    # OPTIONAL: SPEC progress - skip if timeout/failure
    specs = {"completed": 0, "total": 0, "percentage": 0}
    try:
        specs = count_specs(cwd)
    except Exception:
        # Graceful degradation - continue without spec info
        pass

    # OPTIONAL: Checkpoint list - skip if timeout/failure
    checkpoints = []
    try:
        checkpoints = list_checkpoints(cwd, max_count=10)
    except Exception:
        # Graceful degradation - continue without checkpoints
        pass

    # OPTIONAL: Package version info - skip if timeout/failure
    version_info = {}
    try:
        version_info = get_package_version_info()
    except Exception:
        # Graceful degradation - continue without version info
        pass

    # Build message with available information
    branch = git_info.get("branch", "N/A") if git_info else "N/A"
    commit = git_info.get("commit", "N/A")[:7] if git_info else "N/A"
    changes = git_info.get("changes", 0) if git_info else 0
    spec_progress = f"{specs['completed']}/{specs['total']}" if specs["total"] > 0 else "0/0"

    # system_message: displayed directly to the user
    lines = [
        "🚀 MoAI-ADK Session Started",
        "",  # Blank line after title
    ]

    # Add version info first (at the top, right after title)
    if version_info and version_info.get("current") != "unknown":
        if version_info.get("update_available"):
            # Check if this is a major version update
            if version_info.get("is_major_update"):
                # Major version warning
                lines.append(
                    f"   ⚠️  Major version update available: {version_info['current']} → {version_info['latest']}"
                )
                lines.append("   Breaking changes detected. Review release notes:")
                if version_info.get("release_notes_url"):
                    lines.append(f"   📝 {version_info['release_notes_url']}")
            else:
                # Regular update
                lines.append(
                    f"   🗿 MoAI-ADK Ver: {version_info['current']} → {version_info['latest']} available ✨"
                )
                if version_info.get("release_notes_url"):
                    lines.append(f"   📝 Release Notes: {version_info['release_notes_url']}")

            # Add upgrade recommendation
            if version_info.get("upgrade_command"):
                lines.append(f"   ⬆️  Upgrade: {version_info['upgrade_command']}")
        else:
            # No update available - show current version only
            lines.append(f"   🗿 MoAI-ADK Ver: {version_info['current']}")


    # Add Git info only if available (not degraded)
    if git_info:
        lines.append(f"   🌿 Branch: {branch} ({commit})")
        lines.append(f"   📝 Changes: {changes}")

        # Add last commit message if available
        last_commit = git_info.get("last_commit", "")
        if last_commit:
            lines.append(f"   🔨 Last: {last_commit}")

    # Add Checkpoint list (show only the latest 3 items)
    if checkpoints:
        lines.append(f"   🗂️  Checkpoints: {len(checkpoints)} available")
        for cp in reversed(checkpoints[-3:]):  # Latest 3 items
            branch_short = cp["branch"].replace("before-", "")
            lines.append(f"      📌 {branch_short}")
        lines.append("")  # Blank line before restore command
        lines.append("   ↩️  Restore: /alfred:0-project restore")

    # Add SPEC progress only if available (not degraded) - at the bottom
    if specs["total"] > 0:
        lines.append(f"   📋 SPEC Progress: {spec_progress} ({specs['percentage']}%)")

    system_message = "\n".join(lines)

    return HookResult(system_message=system_message)


def handle_session_end(payload: HookPayload) -> HookResult:
    """SessionEnd event handler (default implementation)"""
    return HookResult()


__all__ = ["handle_session_start", "handle_session_end"]
