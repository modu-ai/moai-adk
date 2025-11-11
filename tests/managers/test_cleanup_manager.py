#!/usr/bin/env python3
# @TEST:CLEANUP-MANAGER-001 | TAG: TAG-CLEANUP-MANAGER-TEST-001

"""
Test suite for Cleanup Manager module

This test suite covers all cleanup manager functionality including:
- File cleanup operations
- Statistics tracking
- Configuration management
- Progress reporting
- Background operations
- Exception handling
"""

import json
import os
import shutil
import tempfile
import threading
import time
from datetime import datetime, timedelta
from pathlib import Path
from unittest.mock import Mock, patch, MagicMock

import pytest

import sys
from pathlib import Path

# Add the hooks directory to Python path for imports
sys.path.insert(0, str(Path(__file__).parent.parent.parent / ".claude/hooks/alfred"))

from managers.cleanup_manager import (
    CleanupManager,
    CleanupConfig,
    CleanupStats,
    ProgressCallback,
    get_cleanup_manager
)


class TestCleanupStats:
    """Test CleanupStats functionality"""

    def test_initialization(self):
        """Test CleanupStats initialization"""
        stats = CleanupStats()
        assert stats.files_cleaned == 0
        assert stats.directories_cleaned == 0
        assert stats.total_size_cleaned == 0
        assert stats.errors_encountered == 0
        assert stats.operations_completed == 0
        assert stats.start_time is None
        assert stats.end_time is None

    def test_add_file(self):
        """Test adding cleaned file to statistics"""
        stats = CleanupStats()
        stats.add_file(1000)
        stats.add_file(2000)

        assert stats.files_cleaned == 2
        assert stats.total_size_cleaned == 3000
        assert stats.operations_completed == 2

    def test_add_directory(self):
        """Test adding cleaned directory to statistics"""
        stats = CleanupStats()
        stats.add_directory()
        stats.add_directory()

        assert stats.directories_cleaned == 2
        assert stats.operations_completed == 2

    def test_add_error(self):
        """Test adding error to statistics"""
        stats = CleanupStats()
        stats.add_error()
        stats.add_error()

        assert stats.errors_encountered == 2

    def test_duration_calculation(self):
        """Test duration calculation"""
        stats = CleanupStats()
        stats.start_time = time.time() - 5.0  # 5 seconds ago

        duration = stats.get_duration()
        assert 4.5 <= duration <= 5.5  # Allow small timing variation

    def test_to_dict(self):
        """Test statistics to dictionary conversion"""
        stats = CleanupStats()
        stats.add_file(1000)
        stats.add_directory()
        stats.add_error()

        stats_dict = stats.to_dict()
        assert stats_dict["files_cleaned"] == 1
        assert stats_dict["directories_cleaned"] == 1
        assert stats_dict["total_size_cleaned"] == 1000
        assert stats_dict["errors_encountered"] == 1
        assert stats_dict["operations_completed"] == 2


class TestCleanupConfig:
    """Test CleanupConfig functionality"""

    def test_default_config(self):
        """Test default configuration values"""
        config = CleanupConfig()

        assert config.get_cleanup_days() == 7
        assert config.get_max_reports() == 10
        assert config.get_enabled_patterns() == ["*.tmp", "*.log", "*.cache"]
        assert config.is_enabled() == True
        assert config.get_background_cleanup() == False
        assert config.get_background_interval() == 3600

    def test_custom_config(self):
        """Test custom configuration values"""
        config_dict = {
            "cleanup_days": 14,
            "max_reports": 5,
            "cleanup_patterns": ["*.tmp", "*.log"],
            "enabled": False,
            "background_cleanup": True,
            "background_interval": 1800
        }

        config = CleanupConfig(config_dict)

        assert config.get_cleanup_days() == 14
        assert config.get_max_reports() == 5
        assert config.get_enabled_patterns() == ["*.tmp", "*.log"]
        assert config.is_enabled() == False
        assert config.get_background_cleanup() == True
        assert config.get_background_interval() == 1800

    def test_empty_config(self):
        """Test empty configuration"""
        config = CleanupConfig({})

        assert config.get_cleanup_days() == 7
        assert config.get_max_reports() == 10
        assert config.is_enabled() == True


class TestProgressCallback:
    """Test ProgressCallback functionality"""

    def test_callback_execution(self):
        """Test callback execution"""
        callback_mock = Mock()
        progress_callback = ProgressCallback(callback_mock)

        stats = CleanupStats()
        stats.add_file(1000)

        progress_callback.report(stats, "test_file.txt", 1)

        # Callback should be executed
        callback_mock.assert_called_once()
        call_args = callback_mock.call_args[0][0]  # First argument of call

        assert "stats" in call_args
        assert call_args["current_file"] == "test_file.txt"
        assert call_args["total_files"] == 1
        assert "timestamp" in call_args

    def test_callback_throttling(self):
        """Test callback throttling"""
        callback_mock = Mock()
        progress_callback = ProgressCallback(callback_mock)
        progress_callback.progress_interval = 0.1  # 100ms interval

        stats = CleanupStats()

        # Report multiple times rapidly
        progress_callback.report(stats, "file1.txt", 1)
        progress_callback.report(stats, "file2.txt", 2)
        progress_callback.report(stats, "file3.txt", 3)

        # Callback should only be called once due to throttling
        callback_mock.assert_called_once()


class TestCleanupManager:
    """Test CleanupManager functionality"""

    def setup_method(self):
        """Set up test environment"""
        self.temp_dir = tempfile.mkdtemp()
        self.manager = CleanupManager()

    def teardown_method(self):
        """Clean up test environment"""
        shutil.rmtree(self.temp_dir, ignore_errors=True)
        self.manager.cleanup()

    def test_initialization(self):
        """Test CleanupManager initialization"""
        assert self.manager.config is not None
        assert self.manager.stats is not None
        assert self.manager.progress_callback is not None
        assert len(self.manager._cleanup_threads) == 0

    def test_should_cleanup_today_no_last_cleanup(self):
        """Test should_cleanup_today with no last cleanup date"""
        result = self.manager.should_cleanup_today(None)
        assert result == True

    def test_should_cleanup_today_recent_cleanup(self):
        """Test should_cleanup_today with recent cleanup date"""
        last_cleanup = (datetime.now() - timedelta(days=3)).strftime("%Y-%m-%d")
        result = self.manager.should_cleanup_today(last_cleanup)
        assert result == False

    def test_should_cleanup_old_cleanup(self):
        """Test should_cleanup_today with old cleanup date"""
        last_cleanup = (datetime.now() - timedelta(days=10)).strftime("%Y-%m-%d")
        result = self.manager.should_cleanup_today(last_cleanup)
        assert result == True

    def test_should_cleanup_today_invalid_date(self):
        """Test should_cleanup_today with invalid date"""
        result = self.manager.should_cleanup_today("invalid-date")
        assert result == True

    def test_should_cleanup_today_custom_days(self):
        """Test should_cleanup_today with custom cleanup days"""
        last_cleanup = (datetime.now() - timedelta(days=3)).strftime("%Y-%m-%d")
        result = self.manager.should_cleanup_today(last_cleanup, cleanup_days=2)
        assert result == True

    def test_cleanup_old_files_disabled(self):
        """Test cleanup_old_files with disabled cleanup"""
        self.manager.config.config["enabled"] = False

        temp_file = Path(self.temp_dir) / "test.txt"
        temp_file.write_text("test content")

        result = self.manager.cleanup_old_files([self.temp_dir])

        assert result["files_cleaned"] == 0
        assert temp_file.exists()  # File should not be deleted

    def test_cleanup_old_files_enabled(self):
        """Test cleanup_old_files with enabled cleanup"""
        # Create old files
        old_file = Path(self.temp_dir) / "old.tmp"
        old_file.write_text("old content")

        # Create recent files
        recent_file = Path(self.temp_dir) / "recent.tmp"
        recent_file.write_text("recent content")

        # Mock file modification times to simulate age
        old_time = time.time() - (86400 * 10)  # 10 days ago
        recent_time = time.time() - 3600  # 1 hour ago

        os.utime(old_file, (old_time, old_time))
        os.utime(recent_file, (recent_time, recent_time))

        result = self.manager.cleanup_old_files([self.temp_dir])

        # Old file should be cleaned, recent file should remain
        assert old_file.exists() == False
        assert recent_file.exists() == True
        assert result["files_cleaned"] == 1

    def test_cleanup_directory_with_max_files(self):
        """Test cleanup_directory with max_files limit"""
        # Create multiple files
        files = []
        for i in range(5):
            file_path = Path(self.temp_dir) / f"file{i}.tmp"
            file_path.write_text(f"content {i}")
            files.append(file_path)

        # Mock file modification times
        for i, file_path in enumerate(files):
            file_time = time.time() - (i * 3600)  # Spread over hours
            os.utime(file_path, (file_time, file_time))

        result = self.manager._cleanup_directory(Path(self.temp_dir),
                                               datetime.now() - timedelta(days=1),
                                               max_files=3)

        assert result["files_cleaned"] == 2  # Should keep 3, delete 2 oldest

    def test_cleanup_nonexistent_directory(self):
        """Test cleanup of non-existent directory"""
        nonexistent_dir = "/nonexistent/directory"
        result = self.manager.cleanup_old_files([nonexistent_dir])

        assert result["files_cleaned"] == 0
        assert result["directories_cleaned"] == 0

    def test_cleanup_with_patterns(self):
        """Test cleanup with file patterns"""
        # Create different file types
        Path(self.temp_dir).mkdir(parents=True, exist_ok=True)

        (Path(self.temp_dir) / "file.tmp").write_text("temp file")
        (Path(self.temp_dir) / "file.log").write_text("log file")
        (Path(self.temp_dir) / "file.txt").write_text("text file")

        # Configure cleanup to only clean .tmp and .log files
        self.manager.config.config["cleanup_patterns"] = ["*.tmp", "*.log"]

        result = self.manager.cleanup_old_files([self.temp_dir])

        # Only .tmp and .log files should be cleaned
        assert (Path(self.temp_dir) / "file.tmp").exists() == False
        assert (Path(self.temp_dir) / "file.log").exists() == False
        assert (Path(self.temp_dir) / "file.txt").exists() == True

    def test_cleanup_report_files(self):
        """Test cleanup_report_files functionality"""
        reports_dir = Path(self.temp_dir) / "reports"
        reports_dir.mkdir(parents=True, exist_ok=True)

        # Create old and recent report files
        old_report = reports_dir / "old-report.md"
        old_report.write_text("old report")

        recent_report = reports_dir / "recent-report.md"
        recent_report.write_text("recent report")

        # Mock file modification times
        old_time = time.time() - (86400 * 10)  # 10 days ago
        recent_time = time.time() - 3600  # 1 hour ago

        os.utime(old_report, (old_time, old_time))
        os.utime(recent_report, (recent_time, recent_time))

        result = self.manager.cleanup_report_files(str(reports_dir))

        assert old_report.exists() == False
        assert recent_report.exists() == True
        assert result["files_cleaned"] == 1

    def test_cleanup_system_temp_files(self):
        """Test cleanup_system_temp_files functionality"""
        # Create temporary directory
        temp_subdir = Path(self.temp_dir) / "temp"
        temp_subdir.mkdir(parents=True, exist_ok=True)

        temp_file = temp_subdir / "test.tmp"
        temp_file.write_text("test content")

        result = self.manager.cleanup_system_temp_files([str(temp_subdir)])

        assert temp_file.exists() == False
        assert result["files_cleaned"] == 1

    def test_progress_callback(self):
        """Test progress callback functionality"""
        callback_mock = Mock()
        self.manager.set_progress_callback(callback_mock)

        # Perform some operations
        stats = CleanupStats()
        stats.add_file(1000)

        self.manager.progress_callback.report(stats, "test.txt", 1)

        callback_mock.assert_called_once()

    def test_perform_cleanup(self):
        """Test perform_cleanup functionality"""
        # Create test files
        temp_dir = Path(self.temp_dir) / "temp"
        temp_dir.mkdir(parents=True, exist_ok=True)

        temp_file = temp_dir / "test.tmp"
        temp_file.write_text("test content")

        reports_dir = Path(self.temp_dir) / "reports"
        reports_dir.mkdir(parents=True, exist_ok=True)

        report_file = reports_dir / "report.md"
        report_file.write_text("report content")

        result = self.manager.perform_cleanup()

        # Check that files were cleaned
        assert temp_file.exists() == False
        assert report_file.exists() == False

        # Check result statistics
        assert result["total_files_cleaned"] > 0
        assert "execution_time" in result
        assert "timestamp" in result

    def test_get_statistics(self):
        """Test get_statistics functionality"""
        stats = CleanupStats()
        stats.add_file(1000)
        stats.add_directory()
        stats.add_error()

        self.manager.stats = stats

        result = self.manager.get_statistics()

        assert result["total_files_cleaned"] == 1
        assert result["total_directories_cleaned"] == 1
        assert result["total_size_cleaned"] == 1000
        assert result["errors_encountered"] == 1
        assert "session_stats" in result
        assert "config" in result

    def test_save_statistics(self):
        """Test save_statistics functionality"""
        stats_file = Path(self.temp_dir) / "stats.json"

        # Perform some operations
        self.manager.stats.add_file(1000)
        self.manager.save_statistics(str(stats_file))

        assert stats_file.exists()

        # Verify saved content
        with open(stats_file, 'r') as f:
            saved_stats = json.load(f)

        today = datetime.now().strftime("%Y-%m-%d")
        assert len(saved_stats) > 0
        assert today in saved_stats  # Should have today's statistics

    def test_context_manager(self):
        """Test CleanupManager as context manager"""
        with CleanupManager() as manager:
            manager.stats.add_file(1000)
            assert manager.stats.files_cleaned == 1

        # After context exit, manager should be cleaned up
        # (Test that cleanup was called)

    def test_exception_handling(self):
        """Test exception handling in cleanup operations"""
        # Test with invalid directory
        result = self.manager.cleanup_old_files(["/invalid/directory"])
        assert result["files_cleaned"] == 0
        assert result["directories_cleaned"] == 0

    def test_background_cleanup(self):
        """Test background cleanup functionality"""
        # Start background cleanup
        self.manager.config.config["background_cleanup"] = True
        self.manager.config.config["background_interval"] = 0.1  # 100ms for testing

        self.manager.start_background_cleanup()

        # Wait for background thread to start
        time.sleep(0.2)

        # Stop background cleanup
        self.manager.stop_background_cleanup()

        # Check that threads were cleaned up
        assert len(self.manager._cleanup_threads) == 0

    def test_invalid_config_handling(self):
        """Test handling of invalid configuration"""
        config = {
            "cleanup_days": "invalid",  # Should be integer
            "enabled": "yes",  # Should be boolean
        }

        manager = CleanupManager(config)

        # Should handle invalid configuration gracefully
        assert isinstance(manager.config.get_cleanup_days(), int)
        assert isinstance(manager.config.is_enabled(), bool)

    def test_concurrent_cleanup(self):
        """Test concurrent cleanup operations"""
        # Create test files
        temp_dir = Path(self.temp_dir) / "concurrent"
        temp_dir.mkdir(parents=True, exist_ok=True)

        files = []
        for i in range(10):
            file_path = temp_dir / f"file{i}.tmp"
            file_path.write_text(f"content {i}")
            files.append(file_path)

        # Perform cleanup multiple times concurrently
        def cleanup_worker():
            self.manager.cleanup_old_files([str(temp_dir)])

        threads = []
        for _ in range(3):
            thread = threading.Thread(target=cleanup_worker)
            threads.append(thread)
            thread.start()

        for thread in threads:
            thread.join()

        # All files should be cleaned
        for file_path in files:
            assert not file_path.exists()


class TestCleanupManagerIntegration:
    """Integration tests for CleanupManager"""

    def test_full_cleanup_workflow(self):
        """Test complete cleanup workflow"""
        with tempfile.TemporaryDirectory() as temp_dir:
            # Setup test environment
            manager = CleanupManager({
                "cleanup_days": 1,
                "enabled": True
            })

            # Create test files
            temp_path = Path(temp_dir)

            # Create old files
            old_file = temp_path / "old.tmp"
            old_file.write_text("old content")
            old_time = time.time() - (86400 * 2)  # 2 days ago
            os.utime(old_file, (old_time, old_time))

            # Create recent files
            recent_file = temp_path / "recent.tmp"
            recent_file.write_text("recent content")

            # Test cleanup
            stats = manager.cleanup_old_files([temp_dir])

            # Verify results
            assert not old_file.exists()
            assert recent_file.exists()
            assert stats["files_cleaned"] == 1

            # Test statistics
            manager_stats = manager.get_statistics()
            assert manager_stats["total_files_cleaned"] == 1

            # Test saving statistics
            stats_file = temp_path / "cleanup_stats.json"
            manager.save_statistics(str(stats_file))
            assert stats_file.exists()

    def test_performance_with_large_directory(self):
        """Test cleanup performance with large directory"""
        with tempfile.TemporaryDirectory() as temp_dir:
            manager = CleanupManager({
                "cleanup_days": 1,
                "enabled": True
            })

            temp_path = Path(temp_dir)

            # Create many test files
            start_time = time.time()

            for i in range(100):
                file_path = temp_path / f"file{i:03d}.tmp"
                file_path.write_text(f"content {i}")

            # Clean old files (all files are recent)
            stats = manager.cleanup_old_files([temp_dir])

            end_time = time.time()
            duration = end_time - start_time

            # Performance should be reasonable
            assert duration < 5.0  # Should complete within 5 seconds
            assert stats["files_cleaned"] == 0  # No files should be cleaned

    def test_memory_usage_efficiency(self):
        """Test memory usage efficiency"""
        import psutil
        import os

        process = psutil.Process(os.getpid())
        initial_memory = process.memory_info().rss

        with tempfile.TemporaryDirectory() as temp_dir:
            manager = CleanupManager()

            # Create many files
            temp_path = Path(temp_dir)
            for i in range(50):
                file_path = temp_path / f"file{i:03d}.tmp"
                file_path.write_text(f"content {i}")

            # Perform cleanup
            stats = manager.cleanup_old_files([temp_dir])

        final_memory = process.memory_info().rss
        memory_increase = final_memory - initial_memory

        # Memory increase should be reasonable (< 10MB)
        assert memory_increase < 10 * 1024 * 1024


class TestCleanupManagerFactory:
    """Test CleanupManager factory functions"""

    def test_get_cleanup_manager_default_config(self):
        """Test get_cleanup_manager with default config"""
        manager = get_cleanup_manager()
        assert isinstance(manager, CleanupManager)
        assert manager.config is not None

    def test_get_cleanup_manager_custom_config(self):
        """Test get_cleanup_manager with custom config"""
        custom_config = {
            "cleanup_days": 30,
            "max_reports": 5,
            "enabled": False
        }

        manager = get_cleanup_manager(custom_config)
        assert isinstance(manager, CleanupManager)
        assert manager.config.get_cleanup_days() == 30
        assert manager.config.get_max_reports() == 5
        assert manager.config.is_enabled() == False

    def test_get_cleanup_manager_none_config(self):
        """Test get_cleanup_manager with None config"""
        manager = get_cleanup_manager(None)
        assert isinstance(manager, CleanupManager)
        assert manager.config is not None


if __name__ == "__main__":
    pytest.main([__file__, "-v"])