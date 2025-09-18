#!/usr/bin/env python3
"""
Backup Before Edit Hook for Claude Code
Creates automatic backup of files before any Edit operation for safety
"""

import os
import sys
import shutil
import time
from pathlib import Path

def create_backup(file_path):
    """Create timestamped backup of the file"""
    try:
        path = Path(file_path)

        # Only backup if file exists
        if not path.exists():
            return 0

        # Create backup filename with timestamp
        timestamp = int(time.time())
        backup_path = f"{file_path}.backup.{timestamp}"

        # Copy file to backup location
        shutil.copy2(file_path, backup_path)
        print(f"💾 Backup created: {Path(backup_path).name}")

        # Optional: Clean old backups (keep last 5)
        cleanup_old_backups(file_path)

        return 0

    except Exception as e:
        # Silent fail - don't block the workflow
        print(f"⚠️ Backup failed: {e}", file=sys.stderr)
        return 0

def cleanup_old_backups(file_path, keep_count=5):
    """Clean up old backup files, keeping only the most recent ones"""
    try:
        path = Path(file_path)
        parent_dir = path.parent
        base_name = path.name

        # Find all backup files for this file
        backup_pattern = f"{base_name}.backup.*"
        backup_files = list(parent_dir.glob(backup_pattern))

        # Sort by modification time (newest first)
        backup_files.sort(key=lambda x: x.stat().st_mtime, reverse=True)

        # Remove old backups
        for old_backup in backup_files[keep_count:]:
            try:
                old_backup.unlink()
                print(f"🗑️ Removed old backup: {old_backup.name}")
            except:
                pass  # Silent fail

    except:
        pass  # Silent fail

def main():
    file_path = os.environ.get('CLAUDE_TOOL_FILE_PATH')

    if not file_path:
        return 0  # No file path provided

    return create_backup(file_path)

if __name__ == "__main__":
    sys.exit(main())