"""Core cleanup module for session_start hook

Handles file system cleanup operations including old file removal and stats tracking.

Responsibilities:
- Clean old files from reports, cache, and temp directories
- Delete files based on age and count limits
- Track and update cleanup statistics
"""

import json
import logging
import shutil
from datetime import datetime, timedelta
from pathlib import Path
from typing import Any, Dict, List, Optional

logger = logging.getLogger(__name__)


class CleanupError(Exception):
    """Exception raised for cleanup-related errors"""

    pass


def cleanup_old_files(config: Dict[str, Any]) -> Dict[str, int]:
    """Clean old files from multiple directories

    Cleans files from:
    - .moai/reports: Old analysis reports
    - .moai/cache: Old cache files
    - .moai/temp: Old temporary files

    Args:
        config: Configuration dictionary

    Returns:
        Dictionary with cleanup statistics:
        {
            'reports_cleaned': int,
            'cache_cleaned': int,
            'temp_cleaned': int,
            'total_cleaned': int
        }

    Raises:
        CleanupError: If cleanup operations fail
    """
    stats = {
        "reports_cleaned": 0,
        "cache_cleaned": 0,
        "temp_cleaned": 0,
        "total_cleaned": 0,
    }

    try:
        cleanup_config = config.get("auto_cleanup", {})
        if not cleanup_config.get("enabled", True):
            return stats

        cleanup_days = cleanup_config.get("cleanup_days", 7)
        max_reports = cleanup_config.get("max_reports", 10)

        cutoff_date = datetime.now() - timedelta(days=cleanup_days)

        # Clean reports directory
        reports_dir = Path(".moai/reports")
        if reports_dir.exists():
            stats["reports_cleaned"] = cleanup_directory(
                reports_dir,
                cutoff_date,
                max_reports,
                patterns=["*.json", "*.md"],
            )

        # Clean cache directory
        cache_dir = Path(".moai/cache")
        if cache_dir.exists():
            stats["cache_cleaned"] = cleanup_directory(
                cache_dir,
                cutoff_date,
                None,  # No count limit for cache
                patterns=["*"],
            )

        # Clean temp directory
        temp_dir = Path(".moai/temp")
        if temp_dir.exists():
            stats["temp_cleaned"] = cleanup_directory(
                temp_dir,
                cutoff_date,
                None,  # No count limit for temp
                patterns=["*"],
            )

        stats["total_cleaned"] = (
            stats["reports_cleaned"]
            + stats["cache_cleaned"]
            + stats["temp_cleaned"]
        )

        logger.info(f"Cleanup completed: {stats['total_cleaned']} files removed")

    except Exception as e:
        logger.error(f"File cleanup failed: {e}")
        raise CleanupError(f"Failed to cleanup old files: {e}") from e

    return stats


def cleanup_directory(
    directory: Path,
    cutoff_date: datetime,
    max_files: Optional[int],
    patterns: List[str],
) -> int:
    """Clean files in a directory based on age and count

    Args:
        directory: Target directory path
        cutoff_date: Files older than this date will be deleted
        max_files: Maximum number of files to keep (None for unlimited)
        patterns: File patterns to match (e.g., ["*.json", "*.md"])

    Returns:
        Number of files deleted

    Raises:
        CleanupError: If cleanup operations fail
    """
    if not directory.exists():
        return 0

    cleaned_count = 0

    try:
        # Collect files matching patterns
        files_to_check: List[Path] = []
        for pattern in patterns:
            files_to_check.extend(directory.glob(pattern))

        # Remove duplicates and sort by modification time (oldest first)
        files_to_check = list(set(files_to_check))
        files_to_check.sort(key=lambda f: f.stat().st_mtime)

        # Process files
        for file_path in files_to_check:
            try:
                if not file_path.exists():
                    continue

                # Get file modification time
                file_mtime = datetime.fromtimestamp(file_path.stat().st_mtime)

                # Delete if older than cutoff date
                should_delete = False
                if file_mtime < cutoff_date:
                    should_delete = True

                # Check if exceeds max file count
                elif max_files is not None:
                    # Count existing files that are newer than cutoff
                    remaining_files = [
                        f
                        for f in files_to_check
                        if f.exists()
                        and datetime.fromtimestamp(f.stat().st_mtime) >= cutoff_date
                    ]
                    if len(remaining_files) > max_files:
                        should_delete = True

                # Perform deletion
                if should_delete:
                    if file_path.is_file():
                        file_path.unlink()
                        cleaned_count += 1
                    elif file_path.is_dir():
                        shutil.rmtree(file_path)
                        cleaned_count += 1

            except OSError as e:
                logger.warning(f"Failed to delete {file_path}: {e}")
                continue
            except Exception as e:
                logger.warning(f"Unexpected error deleting {file_path}: {e}")
                continue

    except Exception as e:
        logger.error(f"Directory cleanup failed for {directory}: {e}")
        raise CleanupError(
            f"Failed to cleanup directory {directory}: {e}"
        ) from e

    logger.debug(f"Cleaned {cleaned_count} files from {directory}")
    return cleaned_count


def update_cleanup_stats(cleanup_stats: Dict[str, int]) -> None:
    """Update cleanup statistics in persistent storage

    Maintains a JSON file with daily cleanup statistics for the last 30 days.

    Args:
        cleanup_stats: Statistics from cleanup operation

    Raises:
        CleanupError: If unable to write statistics
    """
    try:
        stats_file = Path(".moai/cache/cleanup_stats.json")
        stats_file.parent.mkdir(exist_ok=True, parents=True)

        # Load existing statistics
        existing_stats: Dict[str, Any] = {}
        if stats_file.exists():
            with open(stats_file, "r", encoding="utf-8") as f:
                existing_stats = json.load(f)

        # Add new statistics for today
        today = datetime.now().strftime("%Y-%m-%d")
        existing_stats[today] = {
            "cleaned_files": cleanup_stats["total_cleaned"],
            "reports_cleaned": cleanup_stats["reports_cleaned"],
            "cache_cleaned": cleanup_stats["cache_cleaned"],
            "temp_cleaned": cleanup_stats["temp_cleaned"],
            "timestamp": datetime.now().isoformat(),
        }

        # Keep only last 30 days of statistics
        cutoff_date = datetime.now() - timedelta(days=30)
        filtered_stats: Dict[str, Any] = {}
        for date_str, stats in existing_stats.items():
            try:
                stat_date = datetime.strptime(date_str, "%Y-%m-%d")
                if stat_date >= cutoff_date:
                    filtered_stats[date_str] = stats
            except ValueError:
                continue

        # Write updated statistics
        with open(stats_file, "w", encoding="utf-8") as f:
            json.dump(filtered_stats, f, indent=2, ensure_ascii=False)

        logger.info(f"Cleanup stats updated: {today}")

    except Exception as e:
        logger.error(f"Failed to update cleanup stats: {e}")
        raise CleanupError(f"Failed to update cleanup stats: {e}") from e
