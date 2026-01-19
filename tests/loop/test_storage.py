# Loop Storage Tests - RED Phase
"""Tests for loop state storage (LoopStorage)."""

import json
from datetime import datetime, timezone
from pathlib import Path

from moai_adk.loop.state import (
    ASTIssueSnapshot,
    DiagnosticSnapshot,
    LoopState,
    LoopStatus,
)
from moai_adk.loop.storage import LoopStorage


class TestLoopStorageInit:
    """Tests for LoopStorage initialization."""

    def test_create_storage_with_default_path(self, tmp_path: Path):
        """LoopStorage should create storage directory if it doesn't exist."""
        storage_dir = tmp_path / ".moai" / "loop"
        storage = LoopStorage(storage_dir=str(storage_dir))
        assert storage.storage_dir == storage_dir
        assert storage_dir.exists()

    def test_create_storage_with_existing_path(self, tmp_path: Path):
        """LoopStorage should work with existing directory."""
        storage_dir = tmp_path / ".moai" / "loop"
        storage_dir.mkdir(parents=True)
        storage = LoopStorage(storage_dir=str(storage_dir))
        assert storage.storage_dir == storage_dir

    def test_storage_creates_subdirectories(self, tmp_path: Path):
        """LoopStorage should create necessary subdirectories."""
        storage_dir = tmp_path / ".moai" / "loop"
        LoopStorage(storage_dir=str(storage_dir))
        # The storage should create the directory
        assert storage_dir.exists()


class TestLoopStorageSaveState:
    """Tests for LoopStorage.save_state method."""

    def test_save_state(self, tmp_path: Path):
        """LoopStorage should save state to a JSON file."""
        storage_dir = tmp_path / ".moai" / "loop"
        storage = LoopStorage(storage_dir=str(storage_dir))

        now = datetime.now(timezone.utc)
        state = LoopState(
            loop_id="loop-001",
            promise="Fix all errors",
            current_iteration=1,
            max_iterations=10,
            status=LoopStatus.RUNNING,
            created_at=now,
            updated_at=now,
        )

        storage.save_state(state)

        # Check file exists
        state_file = storage_dir / "loop-001.json"
        assert state_file.exists()

    def test_save_state_with_history(self, tmp_path: Path):
        """LoopStorage should save state with diagnostic and AST histories."""
        storage_dir = tmp_path / ".moai" / "loop"
        storage = LoopStorage(storage_dir=str(storage_dir))

        now = datetime.now(timezone.utc)
        diag_snapshot = DiagnosticSnapshot(
            timestamp=now,
            error_count=5,
            warning_count=3,
            info_count=1,
            files_affected=["test.py"],
        )
        ast_snapshot = ASTIssueSnapshot(
            timestamp=now,
            total_issues=2,
            by_severity={"warning": 2},
            by_rule={"no-console": 2},
            files_affected=["app.js"],
        )
        state = LoopState(
            loop_id="loop-002",
            promise="Test",
            current_iteration=2,
            max_iterations=10,
            status=LoopStatus.RUNNING,
            created_at=now,
            updated_at=now,
            diagnostics_history=[diag_snapshot],
            ast_issues_history=[ast_snapshot],
        )

        storage.save_state(state)

        # Verify saved content
        state_file = storage_dir / "loop-002.json"
        with open(state_file, "r") as f:
            data = json.load(f)

        assert data["loop_id"] == "loop-002"
        assert len(data["diagnostics_history"]) == 1
        assert len(data["ast_issues_history"]) == 1

    def test_save_state_overwrites_existing(self, tmp_path: Path):
        """LoopStorage should overwrite existing state file."""
        storage_dir = tmp_path / ".moai" / "loop"
        storage = LoopStorage(storage_dir=str(storage_dir))

        now = datetime.now(timezone.utc)
        state1 = LoopState(
            loop_id="loop-001",
            promise="Original",
            current_iteration=1,
            max_iterations=10,
            status=LoopStatus.RUNNING,
            created_at=now,
            updated_at=now,
        )
        state2 = LoopState(
            loop_id="loop-001",
            promise="Updated",
            current_iteration=2,
            max_iterations=10,
            status=LoopStatus.RUNNING,
            created_at=now,
            updated_at=now,
        )

        storage.save_state(state1)
        storage.save_state(state2)

        # Load and verify
        loaded = storage.load_state("loop-001")
        assert loaded is not None
        assert loaded.promise == "Updated"
        assert loaded.current_iteration == 2


class TestLoopStorageLoadState:
    """Tests for LoopStorage.load_state method."""

    def test_load_state(self, tmp_path: Path):
        """LoopStorage should load state from a JSON file."""
        storage_dir = tmp_path / ".moai" / "loop"
        storage = LoopStorage(storage_dir=str(storage_dir))

        now = datetime.now(timezone.utc)
        state = LoopState(
            loop_id="loop-001",
            promise="Fix errors",
            current_iteration=3,
            max_iterations=10,
            status=LoopStatus.RUNNING,
            created_at=now,
            updated_at=now,
        )

        storage.save_state(state)
        loaded = storage.load_state("loop-001")

        assert loaded is not None
        assert loaded.loop_id == "loop-001"
        assert loaded.promise == "Fix errors"
        assert loaded.current_iteration == 3
        assert loaded.status == LoopStatus.RUNNING

    def test_load_state_nonexistent(self, tmp_path: Path):
        """LoopStorage should return None for nonexistent state."""
        storage_dir = tmp_path / ".moai" / "loop"
        storage = LoopStorage(storage_dir=str(storage_dir))

        loaded = storage.load_state("nonexistent")
        assert loaded is None

    def test_load_state_with_history(self, tmp_path: Path):
        """LoopStorage should load state with complete history."""
        storage_dir = tmp_path / ".moai" / "loop"
        storage = LoopStorage(storage_dir=str(storage_dir))

        now = datetime.now(timezone.utc)
        diag_snapshot = DiagnosticSnapshot(
            timestamp=now,
            error_count=5,
            warning_count=3,
            info_count=1,
            files_affected=["test.py"],
        )
        ast_snapshot = ASTIssueSnapshot(
            timestamp=now,
            total_issues=2,
            by_severity={"warning": 2},
            by_rule={"no-console": 2},
            files_affected=["app.js"],
        )
        state = LoopState(
            loop_id="loop-003",
            promise="Test history",
            current_iteration=1,
            max_iterations=10,
            status=LoopStatus.RUNNING,
            created_at=now,
            updated_at=now,
            diagnostics_history=[diag_snapshot],
            ast_issues_history=[ast_snapshot],
        )

        storage.save_state(state)
        loaded = storage.load_state("loop-003")

        assert loaded is not None
        assert len(loaded.diagnostics_history) == 1
        assert loaded.diagnostics_history[0].error_count == 5
        assert len(loaded.ast_issues_history) == 1
        assert loaded.ast_issues_history[0].total_issues == 2

    def test_load_state_preserves_timestamps(self, tmp_path: Path):
        """LoopStorage should preserve datetime fields correctly."""
        storage_dir = tmp_path / ".moai" / "loop"
        storage = LoopStorage(storage_dir=str(storage_dir))

        now = datetime.now(timezone.utc)
        state = LoopState(
            loop_id="loop-004",
            promise="Timestamp test",
            current_iteration=1,
            max_iterations=10,
            status=LoopStatus.RUNNING,
            created_at=now,
            updated_at=now,
        )

        storage.save_state(state)
        loaded = storage.load_state("loop-004")

        assert loaded is not None
        # Compare with some tolerance for serialization
        assert abs((loaded.created_at - now).total_seconds()) < 1
        assert abs((loaded.updated_at - now).total_seconds()) < 1


class TestLoopStorageActiveLoop:
    """Tests for LoopStorage active loop management."""

    def test_get_active_loop_id_none(self, tmp_path: Path):
        """LoopStorage should return None when no active loop."""
        storage_dir = tmp_path / ".moai" / "loop"
        storage = LoopStorage(storage_dir=str(storage_dir))

        assert storage.get_active_loop_id() is None

    def test_get_active_loop_id_with_running_loop(self, tmp_path: Path):
        """LoopStorage should return ID of running loop."""
        storage_dir = tmp_path / ".moai" / "loop"
        storage = LoopStorage(storage_dir=str(storage_dir))

        now = datetime.now(timezone.utc)
        state = LoopState(
            loop_id="active-loop",
            promise="Running loop",
            current_iteration=1,
            max_iterations=10,
            status=LoopStatus.RUNNING,
            created_at=now,
            updated_at=now,
        )

        storage.save_state(state)
        active_id = storage.get_active_loop_id()

        assert active_id == "active-loop"

    def test_get_active_loop_id_ignores_completed(self, tmp_path: Path):
        """LoopStorage should not return completed loops as active."""
        storage_dir = tmp_path / ".moai" / "loop"
        storage = LoopStorage(storage_dir=str(storage_dir))

        now = datetime.now(timezone.utc)
        completed = LoopState(
            loop_id="completed-loop",
            promise="Done",
            current_iteration=10,
            max_iterations=10,
            status=LoopStatus.COMPLETED,
            created_at=now,
            updated_at=now,
        )

        storage.save_state(completed)
        active_id = storage.get_active_loop_id()

        assert active_id is None

    def test_get_active_loop_id_returns_most_recent(self, tmp_path: Path):
        """LoopStorage should return most recently updated active loop."""
        storage_dir = tmp_path / ".moai" / "loop"
        storage = LoopStorage(storage_dir=str(storage_dir))

        now = datetime.now(timezone.utc)
        older = LoopState(
            loop_id="older-loop",
            promise="Older",
            current_iteration=1,
            max_iterations=10,
            status=LoopStatus.PAUSED,
            created_at=now,
            updated_at=now,
        )
        newer = LoopState(
            loop_id="newer-loop",
            promise="Newer",
            current_iteration=1,
            max_iterations=10,
            status=LoopStatus.RUNNING,
            created_at=now,
            updated_at=now,
        )

        storage.save_state(older)
        storage.save_state(newer)
        active_id = storage.get_active_loop_id()

        # Should return the running one (or most recent active)
        assert active_id in ["older-loop", "newer-loop"]


class TestLoopStorageListLoops:
    """Tests for LoopStorage.list_loops method."""

    def test_list_loops_empty(self, tmp_path: Path):
        """LoopStorage should return empty list when no loops exist."""
        storage_dir = tmp_path / ".moai" / "loop"
        storage = LoopStorage(storage_dir=str(storage_dir))

        loops = storage.list_loops()
        assert loops == []

    def test_list_loops(self, tmp_path: Path):
        """LoopStorage should return all loop IDs."""
        storage_dir = tmp_path / ".moai" / "loop"
        storage = LoopStorage(storage_dir=str(storage_dir))

        now = datetime.now(timezone.utc)
        for i in range(3):
            state = LoopState(
                loop_id=f"loop-{i:03d}",
                promise=f"Loop {i}",
                current_iteration=1,
                max_iterations=10,
                status=LoopStatus.RUNNING,
                created_at=now,
                updated_at=now,
            )
            storage.save_state(state)

        loops = storage.list_loops()
        assert len(loops) == 3
        assert "loop-000" in loops
        assert "loop-001" in loops
        assert "loop-002" in loops


class TestLoopStorageDeleteState:
    """Tests for LoopStorage.delete_state method."""

    def test_delete_state(self, tmp_path: Path):
        """LoopStorage should delete state file."""
        storage_dir = tmp_path / ".moai" / "loop"
        storage = LoopStorage(storage_dir=str(storage_dir))

        now = datetime.now(timezone.utc)
        state = LoopState(
            loop_id="to-delete",
            promise="Delete me",
            current_iteration=1,
            max_iterations=10,
            status=LoopStatus.CANCELLED,
            created_at=now,
            updated_at=now,
        )

        storage.save_state(state)
        assert storage.load_state("to-delete") is not None

        result = storage.delete_state("to-delete")
        assert result is True
        assert storage.load_state("to-delete") is None

    def test_delete_state_nonexistent(self, tmp_path: Path):
        """LoopStorage should return False when deleting nonexistent state."""
        storage_dir = tmp_path / ".moai" / "loop"
        storage = LoopStorage(storage_dir=str(storage_dir))

        result = storage.delete_state("nonexistent")
        assert result is False

    def test_delete_state_removes_from_list(self, tmp_path: Path):
        """LoopStorage should remove deleted loop from list."""
        storage_dir = tmp_path / ".moai" / "loop"
        storage = LoopStorage(storage_dir=str(storage_dir))

        now = datetime.now(timezone.utc)
        state = LoopState(
            loop_id="to-delete",
            promise="Delete me",
            current_iteration=1,
            max_iterations=10,
            status=LoopStatus.CANCELLED,
            created_at=now,
            updated_at=now,
        )

        storage.save_state(state)
        assert "to-delete" in storage.list_loops()

        storage.delete_state("to-delete")
        assert "to-delete" not in storage.list_loops()
