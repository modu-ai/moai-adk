#!/usr/bin/env python3
# @CODE:HOOKS-CLARITY-CLEAN | SPEC: Auto-cleanup handler implementation
"""Auto-cleanup Handler: Clean up old reports and maintain .moai directory health

This handler implements automatic cleanup functionality for old reports,
maintaining the .moai directory structure and health.
"""

import json
from datetime import datetime, timedelta
from pathlib import Path
from typing import Dict, List, Any, NamedTuple

from dataclasses import dataclass


@dataclass
class CleanupResult:
    """Result of auto-cleanup operation"""
    continue_execution: bool = True
    files_removed: int = 0
    space_freed_mb: float = 0.0
    message: str = ""


def handle_auto_cleanup(session_data: Dict[str, Any]) -> CleanupResult:
    """Handle automatic cleanup of old reports and .moai maintenance

    Args:
        session_data: Session start data from Claude Code

    Returns:
        CleanupResult: Cleanup operation result
    """
    project_root = Path(session_data.get("projectPath", "."))
    moai_dir = project_root / ".moai"

    if not moai_dir.exists():
        return CleanupResult(
            message="No .moai directory found - skipping cleanup"
        )

    # Check if auto-cleanup is enabled
    config_file = moai_dir / "config.json"
    if not config_file.exists():
        return CleanupResult(
            message="No config.json found - skipping cleanup"
        )

    try:
        config = json.loads(config_file.read_text(encoding="utf-8"))
        auto_cleanup_enabled = config.get("auto_cleanup", {}).get("enabled", False)

        if not auto_cleanup_enabled:
            return CleanupResult(
                message="Auto-cleanup disabled in config"
            )
    except (json.JSONDecodeError, KeyError):
        return CleanupResult(
            message="Invalid config.json format - skipping cleanup"
        )

    # Perform cleanup
    return _perform_cleanup(moai_dir)


def _perform_cleanup(moai_dir: Path) -> CleanupResult:
    """Perform the actual cleanup operation

    Args:
        moai_dir: Path to .moai directory

    Returns:
        CleanupResult: Cleanup operation result
    """
    reports_dir = moai_dir / "reports"
    if not reports_dir.exists():
        return CleanupResult(
            message="No reports directory found - nothing to clean"
        )

    # Get cleanup settings
    cleanup_age_days = 30  # Default: 30 days
    max_reports = 50      # Default: keep 50 most recent reports

    files_removed = 0
    space_freed = 0

    # Find old report files
    cutoff_date = datetime.now() - timedelta(days=cleanup_age_days)
    report_files = list(reports_dir.glob("*.md"))

    # Sort by modification time (newest first)
    report_files.sort(key=lambda f: f.stat().st_mtime, reverse=True)

    # Remove files that are too old or exceed the limit
    files_to_remove = []

    # First, remove files older than cutoff date
    for report_file in report_files:
        if report_file.stat().st_mtime < cutoff_date.timestamp():
            files_to_remove.append(report_file)

    # Then, if still too many files, remove oldest ones
    if len(report_files) > max_reports:
        excess_count = len(report_files) - max_reports
        # Take the oldest files beyond the limit
        old_files = report_files[excess_count:]
        files_to_remove.extend(old_files)

    # Remove duplicates and actually delete files
    files_to_remove = list(set(files_to_remove))

    for file_path in files_to_remove:
        try:
            file_size = file_path.stat().st_size
            file_path.unlink()
            files_removed += 1
            space_freed += file_size
        except OSError:
            continue  # Skip files that can't be deleted

    # Convert to MB
    space_freed_mb = space_freed / (1024 * 1024)

    if files_removed > 0:
        message = f"Cleaned up {files_removed} old report files (freed {space_freed_mb:.1f} MB)"
    else:
        message = "No files needed cleanup"

    return CleanupResult(
        continue_execution=True,
        files_removed=files_removed,
        space_freed_mb=space_freed_mb,
        message=message
    )


def to_dict(self) -> Dict[str, Any]:
    """Convert CleanupResult to dictionary for JSON serialization"""
    return {
        "continue": self.continue_execution,
        "hookSpecificOutput": {
            "files_removed": self.files_removed,
            "space_freed_mb": self.space_freed_mb,
            "message": self.message
        }
    }


# Add to_dict method to CleanupResult
CleanupResult.to_dict = to_dict