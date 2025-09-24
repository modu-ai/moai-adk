#!/usr/bin/env python3
"""
Performance benchmark tests for .claude/hooks optimization (TDD RED phase)

Tests that define performance targets and verify current hooks don't meet them.
These tests should FAIL initially to drive the optimization implementation.

@TEST:UNIT-HOOKS-PERF
@REQ:PERF-HOOKS-001
"""

import time
import psutil
import os
import sys
from pathlib import Path
from unittest.mock import patch, MagicMock
import pytest

# Add project root to path for imports
PROJECT_ROOT = Path(__file__).resolve().parents[3]
HOOKS_DIR = PROJECT_ROOT / ".claude" / "hooks" / "moai"
sys.path.insert(0, str(HOOKS_DIR))


class TestHooksPerformance:
    """Performance benchmark tests for hook optimization

    @TEST:UNIT-PERF-BENCHMARKS
    Target metrics:
    - Session start time ≤ 2.5s (current: ~5s)
    - Memory usage 30% reduction
    - File I/O 50% reduction
    - Code size ~1,000 lines total (current: 3,853)
    """

    def setup_method(self):
        """Setup test environment"""
        self.project_root = PROJECT_ROOT
        self.hooks_dir = HOOKS_DIR

    def test_session_start_time_performance(self):
        """Test session start hook execution time ≤ 2.5s

        @TEST:UNIT-SESSION-START-PERF
        This test should FAIL initially (RED phase)
        """
        # Import session_start_notice
        try:
            import session_start_notice
        except ImportError:
            pytest.skip("session_start_notice.py not found")

        # Mock environment to avoid actual file operations
        with patch('session_start_notice.Path.exists', return_value=True), \
             patch('session_start_notice.Path.is_file', return_value=True), \
             patch('builtins.open', MagicMock()):

            start_time = time.time()

            # Simulate session start processing
            try:
                notifier = session_start_notice.SessionNotifier(self.project_root)
                # This should be fast but currently isn't
                status = notifier.get_project_status()
            except Exception as e:
                # Even with errors, measure the time taken
                pass

            execution_time = time.time() - start_time

            # Performance target: ≤ 2.5 seconds
            # This should FAIL with current implementation
            assert execution_time <= 2.5, f"Session start too slow: {execution_time:.2f}s > 2.5s target"

    def test_memory_usage_optimization(self):
        """Test memory usage reduction target (30% improvement)

        @TEST:UNIT-MEMORY-OPT
        This test should FAIL initially (RED phase)
        """
        process = psutil.Process()
        initial_memory = process.memory_info().rss / 1024 / 1024  # MB

        # Import and use the heavy session_start_notice
        try:
            import session_start_notice
            with patch('session_start_notice.Path.exists', return_value=True):
                notifier = session_start_notice.SessionNotifier(self.project_root)
                # This loads a lot of unnecessary code
        except ImportError:
            pytest.skip("session_start_notice.py not found")

        current_memory = process.memory_info().rss / 1024 / 1024  # MB
        memory_increase = current_memory - initial_memory

        # Target: ≤ 10MB memory increase for hook operations
        # Current implementation likely uses more
        assert memory_increase <= 10.0, f"Memory usage too high: {memory_increase:.2f}MB > 10MB target"

    def test_file_io_operations_count(self):
        """Test file I/O operations reduction (50% target)

        @TEST:UNIT-FILE-IO-OPT
        This test should FAIL initially (RED phase)
        """
        file_operations_count = 0

        # Mock file operations to count them
        def counting_open(*args, **kwargs):
            nonlocal file_operations_count
            file_operations_count += 1
            return MagicMock()

        def counting_exists(*args, **kwargs):
            nonlocal file_operations_count
            file_operations_count += 1
            return True

        def counting_is_file(*args, **kwargs):
            nonlocal file_operations_count
            file_operations_count += 1
            return True

        def counting_iterdir(*args, **kwargs):
            nonlocal file_operations_count
            file_operations_count += 1
            return []

        with patch('builtins.open', side_effect=counting_open), \
             patch.object(Path, 'exists', side_effect=counting_exists), \
             patch.object(Path, 'is_file', side_effect=counting_is_file), \
             patch.object(Path, 'iterdir', side_effect=counting_iterdir):

            try:
                import session_start_notice
                notifier = session_start_notice.SessionNotifier(self.project_root)
                notifier.get_project_status()
            except Exception:
                # Count operations even if they fail
                pass

        # Target: ≤ 10 file I/O operations for session start
        # Current implementation likely does much more
        assert file_operations_count <= 10, f"Too many file I/O operations: {file_operations_count} > 10 target"

    def test_total_code_size_target(self):
        """Test total hook files code size ≤ 1,000 lines

        @TEST:UNIT-CODE-SIZE-TARGET
        This test should FAIL initially (RED phase)
        """
        total_lines = 0

        # Count lines in all hook files
        for py_file in self.hooks_dir.glob("*.py"):
            with open(py_file, 'r', encoding='utf-8') as f:
                lines = len([line for line in f if line.strip() and not line.strip().startswith('#')])
                total_lines += lines

        # Target: ~1,000 lines total (74% reduction from 3,853)
        # This should FAIL with current implementation
        assert total_lines <= 1000, f"Too much code: {total_lines} lines > 1000 lines target"

    def test_session_start_notice_size_target(self):
        """Test session_start_notice.py file size ≤ 200 lines

        @TEST:UNIT-SESSION-SIZE
        This test should FAIL initially (RED phase)
        """
        session_file = self.hooks_dir / "session_start_notice.py"
        if not session_file.exists():
            pytest.skip("session_start_notice.py not found")

        with open(session_file, 'r', encoding='utf-8') as f:
            lines = len([line for line in f if line.strip() and not line.strip().startswith('#')])

        # Target: ≤ 200 lines (reduced from 2,133 lines)
        # This should FAIL with current implementation
        assert lines <= 200, f"session_start_notice.py too large: {lines} lines > 200 lines target"

    def test_file_monitor_integration_target(self):
        """Test unified file_monitor.py size ≤ 150 lines

        @TEST:UNIT-FILE-MONITOR-SIZE
        This test should FAIL initially (RED phase) as files are separate
        """
        # After integration, should have single file_monitor.py
        file_monitor_path = self.hooks_dir / "file_monitor.py"

        if file_monitor_path.exists():
            with open(file_monitor_path, 'r', encoding='utf-8') as f:
                lines = len([line for line in f if line.strip() and not line.strip().startswith('#')])

            # Target: ≤ 150 lines (merged from 545 lines total)
            assert lines <= 150, f"file_monitor.py too large: {lines} lines > 150 lines target"
        else:
            # Should fail because integration not done yet
            assert False, "file_monitor.py integration not completed - separate files still exist"

    def test_removed_unnecessary_hooks(self):
        """Test that unnecessary hook files are removed

        @TEST:UNIT-HOOK-REMOVAL
        This test should FAIL initially (RED phase)
        """
        unnecessary_files = [
            "tag_validator.py",     # 430 lines - move to CI/CD
            "check_style.py"        # 241 lines - replace with native linters
        ]

        for filename in unnecessary_files:
            file_path = self.hooks_dir / filename
            assert not file_path.exists(), f"Unnecessary file still exists: {filename} should be removed"

    def test_optimized_pre_write_guard_size(self):
        """Test pre_write_guard.py optimization ≤ 50 lines

        @TEST:UNIT-PRE-WRITE-GUARD-SIZE
        This test should FAIL initially (RED phase)
        """
        guard_file = self.hooks_dir / "pre_write_guard.py"
        if not guard_file.exists():
            pytest.skip("pre_write_guard.py not found")

        with open(guard_file, 'r', encoding='utf-8') as f:
            lines = len([line for line in f if line.strip() and not line.strip().startswith('#')])

        # Target: ≤ 50 lines (reduced from 131 lines)
        assert lines <= 50, f"pre_write_guard.py too large: {lines} lines > 50 lines target"


if __name__ == "__main__":
    pytest.main([__file__, "-v"])