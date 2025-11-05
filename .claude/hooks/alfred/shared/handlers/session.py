#!/usr/bin/env python3
"""Session event handlers

SessionStart, SessionEnd event handling
"""

from core import HookPayload, HookResult
from core.checkpoint import list_checkpoints
from core.project import count_specs, get_git_info, get_package_version_info


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
        - Phase 3: Add major version warning and release notes display (@TEST:MAJOR-UPDATE-001-07/08)

    @TAG:CHECKPOINT-EVENT-001
    @TAG:HOOKS-TIMEOUT-001
    @CODE:MAJOR-UPDATE-WARN-001
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
    """SessionEnd event handler with session cleanup functionality

    Performs comprehensive session cleanup:
    - Save session metrics to .moai/session_metrics.json
    - Clean up temporary files and caches
    - Log session summary for analysis
    - Prepare session handoff notes

    Args:
        payload: Claude Code event payload

    Returns:
        HookResult(system_message=session cleanup summary)

    @TAG:SESSION-CLEANUP-001
    """
    import json
    import os
    import shutil
    import glob
    from datetime import datetime
    from pathlib import Path

    cwd = payload.get("cwd", ".")
    project_root = Path(cwd)

    # Collect session metrics
    session_metrics = {
        "session_end": datetime.now().isoformat(),
        "project_root": str(project_root),
        "cleanup_actions": [],
        "files_processed": 0,
        "space_freed_mb": 0.0,
        "errors": []
    }

    try:
        # 1. Clean up temporary files
        temp_patterns = [
            ".moai/temp/*",
            ".moai/cache/*.tmp",
            "*.tmp",
            ".pytest_cache/__pycache__/*",
            ".coverage.*.tmp"
        ]

        for pattern in temp_patterns:
            try:
                temp_files = list(project_root.glob(pattern))
                for temp_file in temp_files:
                    if temp_file.is_file():
                        file_size_mb = temp_file.stat().st_size / (1024 * 1024)
                        temp_file.unlink()
                        session_metrics["space_freed_mb"] += file_size_mb
                        session_metrics["files_processed"] += 1

                session_metrics["cleanup_actions"].append(f"Cleaned {pattern}")
            except Exception as e:
                session_metrics["errors"].append(f"Failed to clean {pattern}: {str(e)}")

        # 2. Clean up old checkpoints (keep latest 5)
        checkpoint_dir = project_root / ".moai" / "checkpoints"
        if checkpoint_dir.exists():
            try:
                checkpoints = list(checkpoint_dir.glob("checkpoint-*"))
                checkpoints.sort(key=lambda x: x.stat().st_mtime, reverse=True)

                for old_checkpoint in checkpoints[5:]:  # Keep latest 5
                    if old_checkpoint.is_dir():
                        shutil.rmtree(old_checkpoint)
                        session_metrics["files_processed"] += 1
                        session_metrics["cleanup_actions"].append("Removed old checkpoint")
            except Exception as e:
                session_metrics["errors"].append(f"Failed to clean checkpoints: {str(e)}")

        # 3. Save session metrics
        try:
            metrics_dir = project_root / ".moai"
            metrics_dir.mkdir(exist_ok=True)
            metrics_file = metrics_dir / "session_metrics.json"

            # Load existing metrics or create new
            existing_metrics = []
            if metrics_file.exists():
                try:
                    with open(metrics_file, 'r') as f:
                        existing_metrics = json.load(f)
                        if not isinstance(existing_metrics, list):
                            existing_metrics = []
                except:
                    existing_metrics = []

            # Add current session metrics (keep last 10 sessions)
            existing_metrics.append(session_metrics)
            if len(existing_metrics) > 10:
                existing_metrics = existing_metrics[-10:]

            with open(metrics_file, 'w') as f:
                json.dump(existing_metrics, f, indent=2)

            session_metrics["cleanup_actions"].append("Saved session metrics")
        except Exception as e:
            session_metrics["errors"].append(f"Failed to save metrics: {str(e)}")

        # 4. Create session summary message
        summary_lines = [
            "🧹 Session Cleanup Complete",
            "",
            f"   📁 Files processed: {session_metrics['files_processed']}",
            f"   💾 Space freed: {session_metrics['space_freed_mb']:.1f}MB",
        ]

        if session_metrics["cleanup_actions"]:
            summary_lines.append("   ✅ Actions completed:")
            for action in session_metrics["cleanup_actions"]:
                summary_lines.append(f"      • {action}")

        if session_metrics["errors"]:
            summary_lines.append("   ⚠️  Errors encountered:")
            for error in session_metrics["errors"]:
                summary_lines.append(f"      • {error}")

        summary_lines.extend([
            "",
            "   📊 Session metrics saved to .moai/session_metrics.json",
            "   🔄 Ready for next session",
        ])

        system_message = "\n".join(summary_lines)

        return HookResult(system_message=system_message)

    except Exception as e:
        # Fallback error handling
        error_message = f"Session cleanup failed: {str(e)}"
        return HookResult(
            system_message=f"⚠️  {error_message}",
            hook_specific_output={"error": error_message}
        )


__all__ = ["handle_session_start", "handle_session_end"]
