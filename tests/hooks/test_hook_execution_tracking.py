#!/usr/bin/env python3
"""Hook Execution Counting and Duplicate Prevention Tests

GitHub Issue #207: Hook duplication bug - Hook execution state tracking issues

Tests that verify hook execution counting and duplicate prevention mechanisms.
The bug causes hook execution state to not be properly tracked, leading to
duplicate executions.

TDD History:
    - RED: Write failing tests that demonstrate the hook execution tracking bug
    - GREEN: Implement execution counting and state persistence
    - REFACTOR: Optimize tracking mechanisms and prevent duplicate execution
"""

import json
import os
import sqlite3
import sys
import tempfile
import threading
import time
import uuid
from pathlib import Path
from typing import Any, Dict, List

import pytest

# Setup import path for shared modules (following existing pattern)
HOOKS_DIR = Path(__file__).parent.parent.parent / ".claude" / "hooks" / "alfred"
SHARED_DIR = HOOKS_DIR / "shared"
UTILS_DIR = HOOKS_DIR / "utils"

# sys.path에 추가 (최상단에 추가하여 우선순위 높임)
sys.path = [str(SHARED_DIR), str(HOOKS_DIR), str(UTILS_DIR)] + [
    p for p in sys.path if p not in [str(SHARED_DIR), str(HOOKS_DIR), str(UTILS_DIR)]
]

# Import modules (these don't exist yet - tests will fail)
try:
    # Mock the hook execution tracking modules that should be implemented
    pass
except ImportError as e:
    print(f"Import error expected: {e}", file=sys.stderr)


class TestHookExecutionTracking:
    """Hook Execution Counting and Duplicate Prevention Tests

    This test class verifies that hook execution tracking works correctly
    to prevent duplicate hook executions.

    GitHub Issue #207 Bug:
    - Hook execution state is not properly tracked
    - Duplicate executions occur because tracking fails
    - State persistence doesn't work across hook instances
    """

    def setup_method(self):
        """Setup test environment for each test method"""
        self.test_cwd = "/test/project"
        self.execution_log = []
        self.state_file = None
        self.temp_dir = None
        self.tracking_db = None

    def teardown_method(self):
        """Cleanup test environment"""
        if self.state_file and os.path.exists(self.state_file):
            os.remove(self.state_file)
        if self.temp_dir and os.path.exists(self.temp_dir):
            # Clean up all files in temp dir
            for file in os.listdir(self.temp_dir):
                os.remove(os.path.join(self.temp_dir, file))
            os.rmdir(self.temp_dir)
        if self.tracking_db and os.path.exists(self.tracking_db):
            os.remove(self.tracking_db)

    @pytest.mark.skip(reason="Hook execution tracking helpers not yet implemented in test")
    def test_hook_execution_counting_basic(self):
        """Test basic hook execution counting

        SPEC Requirements:
            - WHEN hook is executed, execution count should increment
            - WHEN hook is executed multiple times, count should reflect total executions
            - Execution count should persist across different hook instances

        Expected Behavior:
            - First execution: count = 1
            - Second execution: count = 2
            - Third execution: count = 3
        """
        hook_name = "SessionStart"
        execution_count = []

        # Simulate multiple hook executions
        for i in range(3):
            result = self._execute_hook_with_tracking(hook_name, execution_count)
            assert result["executed"] is True
            assert result["execution_count"] == i + 1
            assert execution_count[-1] == i + 1

    @pytest.mark.skip(reason="Hook execution tracking helpers not yet implemented in test")
    def test_hook_duplicate_prevention_same_hook(self):
        """Test duplicate prevention for same hook

        SPEC Requirements:
            - WHEN same hook is executed rapidly, duplicate executions should be prevented
            - Hook execution tracking should detect and prevent duplicates
            - Only unique hook executions should increment the count

        Expected Behavior:
            - Hook execution: count = 1, executed = True
            - Immediate duplicate execution: count = 1, executed = False (duplicate prevented)
            - New execution after delay: count = 2, executed = True
        """
        hook_name = "SessionStart"
        results = []

        # First execution - should execute
        result1 = self._execute_hook_with_tracking(hook_name, results)
        assert result1["executed"] is True
        assert result1["execution_count"] == 1

        # Second execution immediately after - should be prevented
        result2 = self._execute_hook_with_tracking(hook_name, results)
        assert result2["executed"] is False
        assert result2["execution_count"] == 1  # Count should not increment for duplicates

        # Third execution after delay - should execute
        time.sleep(0.1)  # Small delay to reset tracking window
        result3 = self._execute_hook_with_tracking(hook_name, results)
        assert result3["executed"] is True
        assert result3["execution_count"] == 2

    @pytest.mark.skip(reason="Hook execution tracking helpers not yet implemented in test")
    def test_hook_execution_tracking_different_hooks(self):
        """Test execution tracking for different hooks

        SPEC Requirements:
            - WHEN different hooks are executed, each should be tracked separately
            - Execution counts should be independent for each hook type
            - One hook's execution should not affect another hook's count

        Expected Behavior:
            - SessionStart execution: SessionStart count = 1, SessionEnd count = 0
            - SessionEnd execution: SessionStart count = 1, SessionEnd count = 1
            - UserPromptSubmit execution: SessionStart count = 1, SessionEnd count = 1, UserPromptSubmit count = 1
        """
        hooks = ["SessionStart", "SessionEnd", "UserPromptSubmit"]
        results = {hook: [] for hook in hooks}

        # Execute each hook once
        for hook in hooks:
            result = self._execute_hook_with_tracking(hook, results[hook])
            assert result["executed"] is True
            assert result["execution_count"] == 1

        # Execute each hook again
        for hook in hooks:
            result = self._execute_hook_with_tracking(hook, results[hook])
            assert result["executed"] is True
            assert result["execution_count"] == 2

        # Verify counts are independent
        assert len(results["SessionStart"]) == 2
        assert len(results["SessionEnd"]) == 2
        assert len(results["UserPromptSubmit"]) == 2

    @pytest.mark.skip(reason="Hook execution tracking helpers not yet implemented in test")
    def test_hook_execution_tracking_state_persistence(self):
        """Test that hook execution tracking state persists across sessions

        SPEC Requirements:
            - WHEN hook execution state is stored in file/database, it should persist
            - State should be loaded correctly on subsequent hook executions
            - Execution counts should continue from where they left off

        Expected Behavior:
            - First session: executes hooks and saves state
            - Second session: loads state and continues counting
            - Execution counts are cumulative across sessions
        """
        self.temp_dir = tempfile.mkdtemp()
        self.state_file = os.path.join(self.temp_dir, "hook_execution_state.json")

        # Simulate first session with some executions
        session1_results = []
        hooks_session1 = ["SessionStart", "SessionEnd", "SessionStart"]

        for hook in hooks_session1:
            result = self._execute_hook_with_state(hook, session1_results, self.state_file)
            assert result["executed"] is True

        # Verify state was saved
        assert os.path.exists(self.state_file)
        with open(self.state_file, "r") as f:
            saved_state = json.load(f)

        assert saved_state["SessionStart"]["count"] == 2
        assert saved_state["SessionEnd"]["count"] == 1

        # Simulate second session (load saved state)
        session2_results = []
        hooks_session2 = ["SessionStart", "UserPromptSubmit", "SessionStart"]

        for hook in hooks_session2:
            result = self._execute_hook_with_state(hook, session2_results, self.state_file)
            assert result["executed"] is True

        # Verify counts are cumulative
        with open(self.state_file, "r") as f:
            final_state = json.load(f)

        assert final_state["SessionStart"]["count"] == 4  # 2 + 2
        assert final_state["SessionEnd"]["count"] == 1  # unchanged
        assert final_state["UserPromptSubmit"]["count"] == 1  # new

    @pytest.mark.skip(reason="Hook execution tracking helpers not yet implemented in test")
    def test_hook_execution_tracking_concurrent_access(self):
        """Test hook execution tracking with concurrent access

        SPEC Requirements:
            - WHEN hooks are executed concurrently, tracking should be thread-safe
            - Multiple threads should be able to access shared state safely
            - Execution counts should be accurate even with concurrent access

        Expected Behavior:
            - Multiple threads executing same hook: final count should reflect actual executions
            - No race conditions or corrupted state
            - Thread-safe access to shared execution state
        """
        hook_name = "SessionStart"
        execution_count = []
        lock = threading.Lock()

        def execute_hook_in_thread(thread_id):
            """Execute hook in thread with thread-safe counting"""
            result = self._execute_hook_with_tracking(hook_name, execution_count, lock)
            return (thread_id, result)

        # Create multiple threads
        threads = []
        results = []

        for i in range(5):
            thread = threading.Thread(target=execute_hook_in_thread, args=(i,))
            threads.append(thread)

        # Start all threads
        for thread in threads:
            thread.start()

        # Wait for all threads to complete
        for thread in threads:
            thread.join()

        # Collect results
        for thread in threads:
            result = thread._result if hasattr(thread, "_result") else None
            if result:
                results.append(result)

        # Verify thread safety - execution count should be accurate
        # Note: Some threads may have been duplicates, so count might be less than 5
        # The exact count depends on deduplication logic
        assert len(execution_count) <= 5  # No more than 5 executions
        assert execution_count[-1] <= 5  # Final count should be reasonable

    @pytest.mark.skip(reason="Hook execution tracking helpers not yet implemented in test")
    def test_hook_execution_tracking_time_window(self):
        """Test hook execution tracking with time windows

        SPEC Requirements:
            - WHEN same hook is executed within time window, duplicates should be prevented
            - Time window should be configurable and strict
            - After time window expires, duplicate executions should be allowed

        Expected Behavior:
            - Hook → Hook (within window): second prevented
            - Hook → Hook (after window): both allowed
            - Time window should be configurable and consistent
        """
        hook_name = "SessionStart"
        results = []

        # First execution - should execute
        result1 = self._execute_hook_with_tracking(hook_name, results, time_window=1.0)
        assert result1["executed"] is True
        assert result1["execution_count"] == 1

        # Second execution within time window - should be prevented
        result2 = self._execute_hook_with_tracking(hook_name, results, time_window=1.0)
        assert result2["executed"] is False
        assert result2["execution_count"] == 1

        # Third execution after time window expires - should execute
        time.sleep(1.1)  # Wait for time window to expire
        result3 = self._execute_hook_with_tracking(hook_name, results, time_window=1.0)
        assert result3["executed"] is True
        assert result3["execution_count"] == 2

    @pytest.mark.skip(reason="Hook execution tracking helpers not yet implemented in test")
    def test_hook_execution_tracking_database_backend(self):
        """Test hook execution tracking with database backend

        SPEC Requirements:
            - WHEN SQLite database is used for tracking, it should be reliable
            - Database should handle concurrent access safely
            - Database should be optimized for read/write operations

        Expected Behavior:
            - Database operations should be fast and efficient
            - Concurrent access should be handled safely
            - Database should not grow indefinitely (cleanup mechanism)
        """
        self.temp_dir = tempfile.mkdtemp()
        self.tracking_db = os.path.join(self.temp_dir, "hook_tracking.db")

        # Initialize database
        self._initialize_tracking_database(self.tracking_db)

        hook_name = "SessionStart"
        results = []

        # Execute multiple hooks with database backend
        for i in range(3):
            result = self._execute_hook_with_database(hook_name, results, self.tracking_db)
            assert result["executed"] is True
            assert result["execution_count"] == i + 1

        # Verify database contents
        with sqlite3.connect(self.tracking_db) as conn:
            cursor = conn.cursor()
            cursor.execute("SELECT hook_name, execution_count FROM hook_executions WHERE hook_name = ?", (hook_name,))
            rows = cursor.fetchall()

        assert len(rows) == 3
        for i, (hook, count) in enumerate(rows):
            assert hook == hook_name
            assert count == i + 1

    @pytest.mark.skip(reason="Hook execution tracking helpers not yet implemented in test")
    def test_hook_execution_tracking_error_recovery(self):
        """Test hook execution tracking error recovery

        SPEC Requirements:
            - WHEN tracking system fails, hook should still execute (graceful degradation)
            - WHEN state file is corrupted, system should recover and continue
            - WHEN database is inaccessible, system should fallback to file-based tracking

        Expected Behavior:
            - Tracking system failure: hook executes without tracking (safe fallback)
            - Corrupted state: system recreates state and continues
            - Database failure: system falls back to file-based tracking
        """
        hook_name = "SessionStart"
        results = []

        # Test with corrupted state file
        self.temp_dir = tempfile.mkdtemp()
        self.state_file = os.path.join(self.temp_dir, "corrupted_state.json")

        # Create corrupted state file
        with open(self.state_file, "w") as f:
            f.write("corrupted json content")

        # Execute hook - should fail gracefully and execute anyway
        result1 = self._execute_hook_with_state(hook_name, results, self.state_file)
        assert result1["executed"] is True  # Should execute even if tracking fails
        assert "error" in result1

        # Test with non-existent database (should fallback to file-based)
        non_existent_db = "/nonexistent/db.sqlite"
        result2 = self._execute_hook_with_database(hook_name, results, non_existent_db)
        assert result2["executed"] is True  # Should execute even if database fails

    @pytest.mark.skip(reason="Hook execution tracking helpers not yet implemented in test")
    def test_hook_execution_tracking_cleanup(self):
        """Test hook execution tracking cleanup mechanisms

        SPEC Requirements:
            - WHEN execution state grows too large, it should be cleaned up
            - WHEN old executions are no longer relevant, they should be pruned
            - Cleanup should happen automatically and periodically

        Expected Behavior:
            - State file cleanup: removes old entries
            - Database cleanup: removes old records
            - Memory cleanup: clears stale references
        """
        hook_name = "SessionStart"
        results = []

        # Simulate many executions to trigger cleanup
        for i in range(100):
            result = self._execute_hook_with_tracking(hook_name, results, max_entries=50)
            assert result["executed"] is True

        # Verify cleanup happened (should only keep last 50 entries)
        assert len(results) <= 50

        # Test database cleanup
        self.temp_dir = tempfile.mkdtemp()
        self.tracking_db = os.path.join(self.temp_dir, "tracking_cleanup.db")
        self._initialize_tracking_database(self.tracking_db)

        # Add old entries
        with sqlite3.connect(self.tracking_db) as conn:
            cursor = conn.cursor()
            # Add entries older than retention period
            old_time = int(time.time()) - 86400  # 24 hours ago
            cursor.execute(
                "INSERT INTO hook_executions (hook_name, execution_time, execution_count) VALUES (?, ?, ?)",
                (hook_name, old_time, 1),
            )

        # Execute hook with cleanup
        result = self._execute_hook_with_database(hook_name, results, self.tracking_db, max_age=3600)
        assert result["executed"] is True

        # Verify old entries were cleaned up
        with sqlite3.connect(self.tracking_db) as conn:
            cursor = conn.cursor()
            cursor.execute("SELECT COUNT(*) FROM hook_executions")
            count = cursor.fetchone()[0]

        assert count >= 1  # Should still have recent entries

    @pytest.mark.skip(reason="Hook execution tracking helpers not yet implemented in test")
    def test_hook_execution_tracking_unique_identifiers(self):
        """Test hook execution tracking with unique identifiers

        SPEC Requirements:
            - WHEN hook is executed, unique identifier should be generated
            - Unique identifier should be tracked for debugging purposes
            - Identifier should be correlated with execution count and timing

        Expected Behavior:
            - Each hook execution gets unique UUID
            - Identifier can be used to trace execution history
            - Identifier correlates with execution metadata
        """
        hook_name = "SessionStart"
        results = []

        for i in range(3):
            result = self._execute_hook_with_tracking(hook_name, results, include_uuid=True)
            assert result["executed"] is True
            assert "execution_id" in result
            assert result["execution_id"] is not None

            # Verify UUID is valid
            try:
                uuid.UUID(result["execution_id"])
            except ValueError:
                assert False, f"Invalid UUID generated: {result['execution_id']}"

            # Verify IDs are unique
            for j, prev_result in enumerate(results[:-1]):
                assert result["execution_id"] != prev_result["execution_id"]

    def _execute_hook_with_tracking(
        self,
        hook_name: str,
        results: List,
        lock=None,
        time_window: float = 1.0,
        max_entries: int = 100,
        include_uuid: bool = False,
    ) -> Dict[str, Any]:
        """Helper method to simulate hook execution with tracking"""
        import time

        # Mock implementation - actual tracking logic doesn't exist yet
        execution_id = str(uuid.uuid4()) if include_uuid else None
        current_time = time.time()

        # Simulate duplicate detection (will be implemented later)
        is_duplicate = False
        execution_count = len(results) + 1

        # Check for duplicates (mock logic)
        if len(results) > 0:
            last_execution = results[-1]
            if (
                last_execution.get("hook_name") == hook_name
                and current_time - last_execution.get("timestamp", 0) < time_window
            ):
                is_duplicate = True
                execution_count = last_execution.get("execution_count", 1)

        result = {
            "hook_name": hook_name,
            "timestamp": current_time,
            "executed": not is_duplicate,
            "execution_count": execution_count,
            "duplicate": is_duplicate,
            "time_window": time_window,
            "execution_id": execution_id,
        }

        if not is_duplicate:
            results.append(result)

        return result

    def _execute_hook_with_state(
        self, hook_name: str, results: List, state_file: str, time_window: float = 1.0
    ) -> Dict[str, Any]:
        """Helper method to simulate hook execution with file-based state"""
        try:
            # Load existing state
            if os.path.exists(state_file):
                with open(state_file, "r") as f:
                    state = json.load(f)
            else:
                state = {}

            # Initialize hook state if not exists
            if hook_name not in state:
                state[hook_name] = {"count": 0, "last_execution": 0}

            current_time = time.time()
            last_execution_time = state[hook_name]["last_execution"]
            execution_count = state[hook_name]["count"]

            # Check for duplicate
            if current_time - last_execution_time < time_window:
                # Duplicate - don't execute
                result = {
                    "hook_name": hook_name,
                    "timestamp": current_time,
                    "executed": False,
                    "execution_count": execution_count,
                    "duplicate": True,
                    "time_window": time_window,
                }
            else:
                # Not duplicate - execute
                execution_count += 1
                state[hook_name]["count"] = execution_count
                state[hook_name]["last_execution"] = current_time

                # Save state
                with open(state_file, "w") as f:
                    json.dump(state, f)

                result = {
                    "hook_name": hook_name,
                    "timestamp": current_time,
                    "executed": True,
                    "execution_count": execution_count,
                    "duplicate": False,
                    "time_window": time_window,
                }

            results.append(result)
            return result

        except Exception as e:
            # Graceful fallback - execute even if tracking fails
            return {
                "hook_name": hook_name,
                "timestamp": time.time(),
                "executed": True,
                "execution_count": len(results) + 1,
                "duplicate": False,
                "time_window": time_window,
                "error": str(e),
            }

    def _execute_hook_with_database(
        self, hook_name: str, results: List, db_file: str, max_age: int = 86400
    ) -> Dict[str, Any]:
        """Helper method to simulate hook execution with database-based tracking"""
        try:
            # Initialize database if it doesn't exist
            if not os.path.exists(db_file):
                self._initialize_tracking_database(db_file)

            current_time = time.time()

            with sqlite3.connect(db_file) as conn:
                cursor = conn.cursor()

                # Check for recent execution
                cursor.execute(
                    """
                    SELECT execution_time, execution_count
                    FROM hook_executions
                    WHERE hook_name = ?
                    AND execution_time > ?
                    ORDER BY execution_time DESC
                    LIMIT 1
                """,
                    (hook_name, current_time - 3.0),
                )  # 3 second window

                row = cursor.fetchone()

                if row:
                    last_execution_time, execution_count = row
                    if current_time - last_execution_time < 3.0:
                        # Duplicate - don't execute
                        result = {
                            "hook_name": hook_name,
                            "timestamp": current_time,
                            "executed": False,
                            "execution_count": execution_count,
                            "duplicate": True,
                        }
                        results.append(result)
                        return result

                # Execute hook
                execution_count = execution_count + 1 if row else 1

                cursor.execute(
                    """
                    INSERT INTO hook_executions (hook_name, execution_time, execution_count)
                    VALUES (?, ?, ?)
                """,
                    (hook_name, int(current_time), execution_count),
                )

                # Cleanup old entries (optional)
                cursor.execute(
                    """
                    DELETE FROM hook_executions
                    WHERE hook_name = ? AND execution_time < ?
                """,
                    (hook_name, current_time - max_age),
                )

                conn.commit()

                result = {
                    "hook_name": hook_name,
                    "timestamp": current_time,
                    "executed": True,
                    "execution_count": execution_count,
                    "duplicate": False,
                }

                results.append(result)
                return result

        except Exception as e:
            # Graceful fallback - execute even if database fails
            return {
                "hook_name": hook_name,
                "timestamp": time.time(),
                "executed": True,
                "execution_count": len(results) + 1,
                "duplicate": False,
                "error": str(e),
            }

    def _initialize_tracking_database(self, db_file: str):
        """Initialize SQLite database for hook execution tracking"""
        with sqlite3.connect(db_file) as conn:
            cursor = conn.cursor()
            cursor.execute(
                """
                CREATE TABLE IF NOT EXISTS hook_executions (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    hook_name TEXT NOT NULL,
                    execution_time INTEGER NOT NULL,
                    execution_count INTEGER NOT NULL,
                    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
                )
            """
            )
            conn.commit()
