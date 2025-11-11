#!/usr/bin/env python3
# @CODE:CLEANUP-MANAGER-001 | TAG: TAG-CLEANUP-MANAGER-001

"""
Cleanup Manager Module

Provides comprehensive cleanup functionality for session management,
including file cleanup, statistics tracking, and progress reporting.

Features:
- Configurable file cleanup patterns and retention policies
- Comprehensive statistics and reporting
- Robust exception handling and logging
- Background cleanup support
- Performance-optimized file operations
"""

import json
import logging
import shutil
import threading
import time
from datetime import datetime, timedelta
from pathlib import Path
from typing import Dict, List, Optional, Callable, Any
import uuid

logger = logging.getLogger(__name__)


class CleanupStats:
    """Statistics container for cleanup operations"""

    def __init__(self):
        self.files_cleaned: int = 0
        self.directories_cleaned: int = 0
        self.total_size_cleaned: int = 0
        self.errors_encountered: int = 0
        self.operations_completed: int = 0
        self.start_time: Optional[float] = None
        self.end_time: Optional[float] = None

    def to_dict(self) -> Dict[str, Any]:
        """Convert statistics to dictionary"""
        return {
            "files_cleaned": self.files_cleaned,
            "directories_cleaned": self.directories_cleaned,
            "total_size_cleaned": self.total_size_cleaned,
            "errors_encountered": self.errors_encountered,
            "operations_completed": self.operations_completed,
            "duration_seconds": self.get_duration(),
            "start_time": self.start_time,
            "end_time": self.end_time
        }

    def get_duration(self) -> float:
        """Get cleanup duration in seconds"""
        if not self.start_time:
            return 0.0
        end = self.end_time or time.time()
        return end - self.start_time

    def add_file(self, size: int = 0):
        """Add cleaned file to statistics"""
        self.files_cleaned += 1
        self.total_size_cleaned += size
        self.operations_completed += 1

    def add_directory(self):
        """Add cleaned directory to statistics"""
        self.directories_cleaned += 1
        self.operations_completed += 1

    def add_error(self):
        """Add encountered error to statistics"""
        self.errors_encountered += 1


class CleanupConfig:
    """Configuration for cleanup operations"""

    def __init__(self, config_dict: Optional[Dict[str, Any]] = None):
        self.config = config_dict or {}

    def get_cleanup_days(self) -> int:
        """Get cleanup days retention"""
        value = self.config.get("cleanup_days", 7)
        if isinstance(value, str):
            try:
                return int(value)
            except (ValueError, TypeError):
                return 7
        return int(value) if value is not None else 7

    def get_max_reports(self) -> int:
        """Get maximum reports to keep"""
        value = self.config.get("max_reports", 10)
        if isinstance(value, str):
            try:
                return int(value)
            except (ValueError, TypeError):
                return 10
        return int(value) if value is not None else 10

    def get_enabled_patterns(self) -> List[str]:
        """Get enabled file patterns"""
        patterns = self.config.get("cleanup_patterns", ["*.tmp", "*.log", "*.cache"])
        if isinstance(patterns, str):
            return [patterns]
        return list(patterns) if patterns else ["*.tmp", "*.log", "*.cache"]

    def is_enabled(self) -> bool:
        """Check if cleanup is enabled"""
        value = self.config.get("enabled", True)
        if isinstance(value, str):
            return value.lower() in ("true", "yes", "1", "on")
        return bool(value)

    def get_background_cleanup(self) -> bool:
        """Check if background cleanup is enabled"""
        value = self.config.get("background_cleanup", False)
        if isinstance(value, str):
            return value.lower() in ("true", "yes", "1", "on")
        return bool(value)

    def get_background_interval(self) -> int:
        """Get background cleanup interval in seconds"""
        value = self.config.get("background_interval", 3600)
        if isinstance(value, str):
            try:
                return int(value)
            except (ValueError, TypeError):
                return 3600
        return int(value) if value is not None else 3600


class ProgressCallback:
    """Progress callback interface"""

    def __init__(self, callback: Optional[Callable[[Dict[str, Any]], None]] = None):
        self.callback = callback
        self.last_progress_time: float = 0
        self.progress_interval: float = 1.0  # 1 second

    def report(self, stats: CleanupStats, current_file: str = "", total_files: int = 0):
        """Report progress"""
        current_time = time.time()
        if current_time - self.last_progress_time >= self.progress_interval:
            progress_data = {
                "stats": stats.to_dict(),
                "current_file": current_file,
                "total_files": total_files,
                "timestamp": datetime.now().isoformat()
            }
            if self.callback:
                self.callback(progress_data)
            self.last_progress_time = current_time


class CleanupManager:
    """Main cleanup manager for session operations"""

    def __init__(self, config: Optional[Dict[str, Any]] = None):
        """Initialize cleanup manager

        Args:
            config: Configuration dictionary for cleanup operations
        """
        self.config = CleanupConfig(config)
        self.stats = CleanupStats()
        self.progress_callback = ProgressCallback()
        self._cleanup_threads: List[threading.Thread] = []
        self._stop_event = threading.Event()

    def set_progress_callback(self, callback: Optional[Callable[[Dict[str, Any]], None]]):
        """Set progress callback function"""
        self.progress_callback = ProgressCallback(callback)

    def should_cleanup_today(self, last_cleanup: Optional[str], cleanup_days: int = None) -> bool:
        """Check if cleanup should be performed today

        Args:
            last_cleanup: Last cleanup date (YYYY-MM-DD format)
            cleanup_days: Cleanup interval in days

        Returns:
            True if cleanup is needed, False otherwise
        """
        if not last_cleanup:
            return True

        if cleanup_days is None:
            cleanup_days = self.config.get_cleanup_days()

        try:
            last_date = datetime.strptime(last_cleanup, "%Y-%m-%d")
            next_cleanup = last_date + timedelta(days=cleanup_days)
            return datetime.now() >= next_cleanup
        except ValueError:
            return True

    def cleanup_old_files(self, target_dirs: List[str]) -> Dict[str, int]:
        """Clean up old files from specified directories

        Args:
            target_dirs: List of directory paths to clean

        Returns:
            Dictionary with cleanup statistics
        """
        if not self.config.is_enabled():
            return {"files_cleaned": 0, "directories_cleaned": 0, "total_size": 0}

        self.stats = CleanupStats()
        self.stats.start_time = time.time()

        cleanup_days = self.config.get_cleanup_days()
        cutoff_date = datetime.now() - timedelta(days=cleanup_days)

        total_stats = {
            "files_cleaned": 0,
            "directories_cleaned": 0,
            "total_size": 0
        }

        for dir_path in target_dirs:
            try:
                directory = Path(dir_path)
                if not directory.exists():
                    continue

                dir_stats = self._cleanup_directory(directory, cutoff_date)
                total_stats["files_cleaned"] += dir_stats["files_cleaned"]
                total_stats["directories_cleaned"] += dir_stats["directories_cleaned"]
                total_stats["total_size"] += dir_stats["total_size"]

            except Exception as e:
                logger.error(f"Failed to clean directory {dir_path}: {e}")
                self.stats.add_error()

        self.stats.end_time = time.time()

        # Update configuration with last cleanup time
        return total_stats

    def _cleanup_directory(self, directory: Path, cutoff_date: datetime,
                          max_files: Optional[int] = None) -> Dict[str, int]:
        """Clean up a specific directory

        Args:
            directory: Directory path to clean
            cutoff_date: Cutoff date for file deletion
            max_files: Maximum number of files to keep

        Returns:
            Directory cleanup statistics
        """
        stats = {"files_cleaned": 0, "directories_cleaned": 0, "total_size": 0}
        files_found = 0

        try:
            # Collect files matching patterns
            files_to_process = []
            for pattern in self.config.get_enabled_patterns():
                files_to_process.extend(directory.rglob(pattern))

            files_found = len(files_to_process)
            files_to_process.sort(key=lambda f: f.stat().st_mtime)

            # Process files
            for file_path in files_to_process:
                try:
                    file_mtime = datetime.fromtimestamp(file_path.stat().st_mtime)

                    # Check if file should be deleted based on date or file count
                    should_delete = False
                    if file_mtime < cutoff_date:
                        should_delete = True
                    elif max_files is not None and stats["files_cleaned"] >= max_files:
                        should_delete = False

                    if should_delete and file_path.exists():
                        file_size = file_path.stat().st_size if file_path.is_file() else 0

                        if file_path.is_file():
                            file_path.unlink()
                            stats["files_cleaned"] += 1
                            stats["total_size"] += file_size
                        elif file_path.is_dir():
                            shutil.rmtree(file_path)
                            stats["directories_cleaned"] += 1

                except Exception as e:
                    logger.warning(f"Failed to process {file_path}: {e}")
                    self.stats.add_error()
                    continue

        except Exception as e:
            logger.error(f"Directory cleanup failed for {directory}: {e}")
            self.stats.add_error()

        return stats

    def cleanup_system_temp_files(self, temp_dirs: List[str] = None) -> Dict[str, int]:
        """Clean up system temporary files

        Args:
            temp_dirs: List of temporary directories to clean

        Returns:
            Cleanup statistics
        """
        if temp_dirs is None:
            temp_dirs = [
                "/tmp",
                "/var/tmp",
                Path.home() / ".cache",
                Path.home() / ".tmp"
            ]

        return self.cleanup_old_files(temp_dirs)

    def cleanup_report_files(self, reports_dir: str = ".moai/reports") -> Dict[str, int]:
        """Clean up report files

        Args:
            reports_dir: Reports directory path

        Returns:
            Cleanup statistics
        """
        reports_stats = {"files_cleaned": 0, "directories_cleaned": 0, "total_size": 0}

        try:
            directory = Path(reports_dir)
            if not directory.exists():
                return reports_stats

            # Clean old reports
            cleanup_days = self.config.get_cleanup_days()
            max_reports = self.config.get_max_reports()
            cutoff_date = datetime.now() - timedelta(days=cleanup_days)

            # Get all report files
            report_files = list(directory.glob("*.md")) + list(directory.glob("*.json"))
            report_files.sort(key=lambda f: f.stat().st_mtime)

            # Clean old files and exceed count
            stats = self._cleanup_directory(directory, cutoff_date, max_reports)
            reports_stats.update(stats)

        except Exception as e:
            logger.error(f"Report cleanup failed: {e}")
            self.stats.add_error()

        return reports_stats

    def start_background_cleanup(self, interval_seconds: int = None) -> None:
        """Start background cleanup thread

        Args:
            interval_seconds: Cleanup interval in seconds
        """
        if not self.config.get_background_cleanup():
            return

        if interval_seconds is None:
            interval_seconds = self.config.get_background_interval()

        def background_cleanup_task():
            while not self._stop_event.is_set():
                if self._stop_event.wait(timeout=interval_seconds):
                    break

                try:
                    self.perform_cleanup()
                except Exception as e:
                    logger.error(f"Background cleanup failed: {e}")

        cleanup_thread = threading.Thread(target=background_cleanup_task, daemon=True)
        cleanup_thread.start()
        self._cleanup_threads.append(cleanup_thread)

    def stop_background_cleanup(self, timeout: float = 5.0) -> None:
        """Stop background cleanup threads

        Args:
            timeout: Maximum time to wait for threads to stop
        """
        self._stop_event.set()

        for thread in self._cleanup_threads:
            if thread.is_alive():
                thread.join(timeout=timeout)
                if thread.is_alive():
                    logger.warning(f"Cleanup thread did not finish within timeout")

        self._cleanup_threads.clear()

    def perform_cleanup(self) -> Dict[str, Any]:
        """Perform full cleanup operation

        Returns:
            Complete cleanup results including statistics
        """
        cleanup_dirs = [
            ".moai/temp",
            ".moai/cache",
            ".moai/reports"
        ]

        total_stats = self.cleanup_old_files(cleanup_dirs)
        report_stats = self.cleanup_report_files()
        temp_stats = self.cleanup_system_temp_files()

        # Combine all statistics
        combined_stats = {
            "total_files_cleaned": total_stats["files_cleaned"],
            "total_directories_cleaned": total_stats["directories_cleaned"],
            "total_size_cleaned": total_stats["total_size"],
            "report_files_cleaned": report_stats["files_cleaned"],
            "temp_files_cleaned": temp_stats["files_cleaned"],
            "execution_time": self.stats.get_duration(),
            "timestamp": datetime.now().isoformat()
        }

        return combined_stats

    def get_statistics(self) -> Dict[str, Any]:
        """Get cleanup statistics

        Returns:
            Complete statistics dictionary
        """
        return {
            "session_stats": self.stats.to_dict(),
            "config": self.config.config,
            "total_files_cleaned": self.stats.files_cleaned,
            "total_directories_cleaned": self.stats.directories_cleaned,
            "total_size_cleaned": self.stats.total_size_cleaned,
            "errors_encountered": self.stats.errors_encountered,
            "operations_completed": self.stats.operations_completed
        }

    def save_statistics(self, stats_file: str = ".moai/cache/cleanup_stats.json") -> None:
        """Save statistics to file

        Args:
            stats_file: Path to save statistics
        """
        try:
            stats_path = Path(stats_file)
            stats_path.parent.mkdir(parents=True, exist_ok=True)

            # Load existing statistics
            existing_stats = {}
            if stats_path.exists():
                with open(stats_path, 'r', encoding='utf-8') as f:
                    existing_stats = json.load(f)

            # Add current session statistics
            today = datetime.now().strftime("%Y-%m-%d")
            existing_stats[today] = self.get_statistics()

            # Keep only last 30 days
            cutoff_date = datetime.now() - timedelta(days=30)
            filtered_stats = {}
            for date, stats in existing_stats.items():
                try:
                    stat_date = datetime.strptime(date, "%Y-%m-%d")
                    if stat_date >= cutoff_date:
                        filtered_stats[date] = stats
                except ValueError:
                    continue

            # Save updated statistics
            with open(stats_path, 'w', encoding='utf-8') as f:
                json.dump(filtered_stats, f, indent=2, ensure_ascii=False)

        except Exception as e:
            logger.error(f"Failed to save statistics: {e}")
            self.stats.add_error()

    def cleanup(self) -> None:
        """Cleanup resources"""
        self.stop_background_cleanup()

    def __enter__(self):
        """Context manager entry"""
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        """Context manager exit"""
        self.cleanup()
        return False


def get_cleanup_manager(config: Optional[Dict[str, Any]] = None) -> CleanupManager:
    """Get cleanup manager instance

    Args:
        config: Optional configuration dictionary

    Returns:
        CleanupManager instance
    """
    return CleanupManager(config)