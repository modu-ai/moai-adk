# Loop Controller Tests - RED Phase
"""Tests for MoAILoopController."""

from datetime import datetime, timezone
from pathlib import Path
from unittest.mock import MagicMock, patch

import pytest

from moai_adk.loop.controller import MoAILoopController
from moai_adk.loop.state import (
    ASTIssueSnapshot,
    CompletionResult,
    DiagnosticSnapshot,
    FeedbackResult,
    LoopState,
    LoopStatus,
)
from moai_adk.loop.storage import LoopStorage


class TestMoAILoopControllerInit:
    """Tests for MoAILoopController initialization."""

    def test_create_controller_default(self, tmp_path: Path):
        """MoAILoopController should be creatable with default settings."""
        storage = LoopStorage(storage_dir=str(tmp_path / ".moai" / "loop"))
        controller = MoAILoopController(storage=storage)
        assert controller is not None

    def test_create_controller_without_storage(self, tmp_path: Path):
        """MoAILoopController should create default storage if not provided."""
        with patch("moai_adk.loop.controller.LoopStorage") as mock_storage_class:
            mock_storage_class.return_value = MagicMock()
            controller = MoAILoopController()
            assert controller is not None


class TestMoAILoopControllerStartLoop:
    """Tests for MoAILoopController.start_loop method."""

    def test_start_loop_basic(self, tmp_path: Path):
        """MoAILoopController should start a new loop."""
        storage = LoopStorage(storage_dir=str(tmp_path / ".moai" / "loop"))
        controller = MoAILoopController(storage=storage)

        state = controller.start_loop(promise="Fix all LSP errors")

        assert state is not None
        assert state.promise == "Fix all LSP errors"
        assert state.status == LoopStatus.RUNNING
        assert state.current_iteration == 0
        assert state.max_iterations == 10  # default

    def test_start_loop_with_max_iterations(self, tmp_path: Path):
        """MoAILoopController should respect max_iterations parameter."""
        storage = LoopStorage(storage_dir=str(tmp_path / ".moai" / "loop"))
        controller = MoAILoopController(storage=storage)

        state = controller.start_loop(promise="Test", max_iterations=5)

        assert state.max_iterations == 5

    def test_start_loop_saves_state(self, tmp_path: Path):
        """MoAILoopController should persist the loop state."""
        storage = LoopStorage(storage_dir=str(tmp_path / ".moai" / "loop"))
        controller = MoAILoopController(storage=storage)

        state = controller.start_loop(promise="Persistent loop")

        # Verify it was saved
        loaded = storage.load_state(state.loop_id)
        assert loaded is not None
        assert loaded.promise == "Persistent loop"

    def test_start_loop_generates_unique_id(self, tmp_path: Path):
        """MoAILoopController should generate unique loop IDs."""
        storage = LoopStorage(storage_dir=str(tmp_path / ".moai" / "loop"))
        controller = MoAILoopController(storage=storage)

        state1 = controller.start_loop(promise="Loop 1")
        state2 = controller.start_loop(promise="Loop 2")

        assert state1.loop_id != state2.loop_id


class TestMoAILoopControllerCheckCompletion:
    """Tests for MoAILoopController.check_completion method."""

    def test_check_completion_no_issues(self, tmp_path: Path):
        """MoAILoopController should detect completion when no issues remain."""
        storage = LoopStorage(storage_dir=str(tmp_path / ".moai" / "loop"))
        controller = MoAILoopController(storage=storage)

        now = datetime.now(timezone.utc)
        state = LoopState(
            loop_id="test-loop",
            promise="Fix all errors",
            current_iteration=3,
            max_iterations=10,
            status=LoopStatus.RUNNING,
            created_at=now,
            updated_at=now,
            diagnostics_history=[
                DiagnosticSnapshot(
                    timestamp=now,
                    error_count=0,
                    warning_count=0,
                    info_count=0,
                    files_affected=[],
                )
            ],
            ast_issues_history=[
                ASTIssueSnapshot(
                    timestamp=now,
                    total_issues=0,
                    by_severity={},
                    by_rule={},
                    files_affected=[],
                )
            ],
        )

        result = controller.check_completion(state)

        assert isinstance(result, CompletionResult)
        assert result.is_complete is True
        assert result.remaining_issues == 0
        assert result.progress_percentage == 100.0

    def test_check_completion_with_errors(self, tmp_path: Path):
        """MoAILoopController should detect incomplete when errors exist."""
        storage = LoopStorage(storage_dir=str(tmp_path / ".moai" / "loop"))
        controller = MoAILoopController(storage=storage)

        now = datetime.now(timezone.utc)
        state = LoopState(
            loop_id="test-loop",
            promise="Fix all errors",
            current_iteration=3,
            max_iterations=10,
            status=LoopStatus.RUNNING,
            created_at=now,
            updated_at=now,
            diagnostics_history=[
                DiagnosticSnapshot(
                    timestamp=now,
                    error_count=5,
                    warning_count=3,
                    info_count=1,
                    files_affected=["main.py"],
                )
            ],
            ast_issues_history=[
                ASTIssueSnapshot(
                    timestamp=now,
                    total_issues=2,
                    by_severity={"error": 2},
                    by_rule={"sql-injection": 2},
                    files_affected=["db.py"],
                )
            ],
        )

        result = controller.check_completion(state)

        assert result.is_complete is False
        assert result.remaining_issues > 0

    def test_check_completion_calculates_progress(self, tmp_path: Path):
        """MoAILoopController should calculate progress percentage."""
        storage = LoopStorage(storage_dir=str(tmp_path / ".moai" / "loop"))
        controller = MoAILoopController(storage=storage)

        now = datetime.now(timezone.utc)
        # First snapshot with 10 errors
        first_snapshot = DiagnosticSnapshot(
            timestamp=now,
            error_count=10,
            warning_count=0,
            info_count=0,
            files_affected=["main.py"],
        )
        # Current snapshot with 5 errors (50% progress)
        current_snapshot = DiagnosticSnapshot(
            timestamp=now,
            error_count=5,
            warning_count=0,
            info_count=0,
            files_affected=["main.py"],
        )
        state = LoopState(
            loop_id="test-loop",
            promise="Fix all errors",
            current_iteration=3,
            max_iterations=10,
            status=LoopStatus.RUNNING,
            created_at=now,
            updated_at=now,
            diagnostics_history=[first_snapshot, current_snapshot],
            ast_issues_history=[],
        )

        result = controller.check_completion(state)

        assert result.progress_percentage >= 0
        assert result.progress_percentage <= 100


class TestMoAILoopControllerRunFeedbackLoop:
    """Tests for MoAILoopController.run_feedback_loop method."""

    @pytest.mark.asyncio
    async def test_run_feedback_loop_basic(self, tmp_path: Path):
        """MoAILoopController should execute feedback loop iteration."""
        storage = LoopStorage(storage_dir=str(tmp_path / ".moai" / "loop"))
        controller = MoAILoopController(storage=storage)

        now = datetime.now(timezone.utc)
        state = LoopState(
            loop_id="test-loop",
            promise="Fix errors",
            current_iteration=1,
            max_iterations=10,
            status=LoopStatus.RUNNING,
            created_at=now,
            updated_at=now,
        )

        # Mock the LSP and AST-grep integrations
        with patch.object(controller, "_get_lsp_diagnostics") as mock_lsp:
            with patch.object(controller, "_get_ast_issues") as mock_ast:
                mock_lsp.return_value = DiagnosticSnapshot(
                    timestamp=now,
                    error_count=2,
                    warning_count=1,
                    info_count=0,
                    files_affected=["main.py"],
                )
                mock_ast.return_value = ASTIssueSnapshot(
                    timestamp=now,
                    total_issues=1,
                    by_severity={"warning": 1},
                    by_rule={"no-console": 1},
                    files_affected=["app.js"],
                )

                result = await controller.run_feedback_loop(state)

        assert isinstance(result, FeedbackResult)
        assert result.lsp_diagnostics.error_count == 2
        assert result.ast_issues.total_issues == 1
        assert len(result.feedback_text) > 0

    @pytest.mark.asyncio
    async def test_run_feedback_loop_updates_history(self, tmp_path: Path):
        """MoAILoopController should update state history after feedback loop."""
        storage = LoopStorage(storage_dir=str(tmp_path / ".moai" / "loop"))
        controller = MoAILoopController(storage=storage)

        now = datetime.now(timezone.utc)
        state = LoopState(
            loop_id="test-loop",
            promise="Fix errors",
            current_iteration=1,
            max_iterations=10,
            status=LoopStatus.RUNNING,
            created_at=now,
            updated_at=now,
        )
        storage.save_state(state)

        with patch.object(controller, "_get_lsp_diagnostics") as mock_lsp:
            with patch.object(controller, "_get_ast_issues") as mock_ast:
                mock_lsp.return_value = DiagnosticSnapshot(
                    timestamp=now,
                    error_count=3,
                    warning_count=0,
                    info_count=0,
                    files_affected=["test.py"],
                )
                mock_ast.return_value = ASTIssueSnapshot(
                    timestamp=now,
                    total_issues=0,
                    by_severity={},
                    by_rule={},
                    files_affected=[],
                )

                await controller.run_feedback_loop(state)

        # Verify state was updated
        loaded = storage.load_state("test-loop")
        assert loaded is not None
        assert len(loaded.diagnostics_history) > 0


class TestMoAILoopControllerCancelLoop:
    """Tests for MoAILoopController.cancel_loop method."""

    def test_cancel_loop_success(self, tmp_path: Path):
        """MoAILoopController should cancel an active loop."""
        storage = LoopStorage(storage_dir=str(tmp_path / ".moai" / "loop"))
        controller = MoAILoopController(storage=storage)

        state = controller.start_loop(promise="Cancel me")
        result = controller.cancel_loop(state.loop_id)

        assert result is True

        # Verify state was updated
        loaded = storage.load_state(state.loop_id)
        assert loaded is not None
        assert loaded.status == LoopStatus.CANCELLED

    def test_cancel_loop_nonexistent(self, tmp_path: Path):
        """MoAILoopController should return False for nonexistent loop."""
        storage = LoopStorage(storage_dir=str(tmp_path / ".moai" / "loop"))
        controller = MoAILoopController(storage=storage)

        result = controller.cancel_loop("nonexistent")

        assert result is False

    def test_cancel_loop_already_completed(self, tmp_path: Path):
        """MoAILoopController should return False for already completed loop."""
        storage = LoopStorage(storage_dir=str(tmp_path / ".moai" / "loop"))
        controller = MoAILoopController(storage=storage)

        now = datetime.now(timezone.utc)
        state = LoopState(
            loop_id="completed-loop",
            promise="Already done",
            current_iteration=10,
            max_iterations=10,
            status=LoopStatus.COMPLETED,
            created_at=now,
            updated_at=now,
        )
        storage.save_state(state)

        result = controller.cancel_loop("completed-loop")

        assert result is False


class TestMoAILoopControllerGetActiveLoop:
    """Tests for MoAILoopController.get_active_loop method."""

    def test_get_active_loop_exists(self, tmp_path: Path):
        """MoAILoopController should return active loop state."""
        storage = LoopStorage(storage_dir=str(tmp_path / ".moai" / "loop"))
        controller = MoAILoopController(storage=storage)

        controller.start_loop(promise="Active loop")
        active = controller.get_active_loop()

        assert active is not None
        assert active.promise == "Active loop"
        assert active.status == LoopStatus.RUNNING

    def test_get_active_loop_none(self, tmp_path: Path):
        """MoAILoopController should return None when no active loop."""
        storage = LoopStorage(storage_dir=str(tmp_path / ".moai" / "loop"))
        controller = MoAILoopController(storage=storage)

        active = controller.get_active_loop()

        assert active is None


class TestMoAILoopControllerGetLoopStatus:
    """Tests for MoAILoopController.get_loop_status method."""

    def test_get_loop_status_exists(self, tmp_path: Path):
        """MoAILoopController should return loop state by ID."""
        storage = LoopStorage(storage_dir=str(tmp_path / ".moai" / "loop"))
        controller = MoAILoopController(storage=storage)

        state = controller.start_loop(promise="Status check")
        loaded = controller.get_loop_status(state.loop_id)

        assert loaded is not None
        assert loaded.loop_id == state.loop_id
        assert loaded.promise == "Status check"

    def test_get_loop_status_nonexistent(self, tmp_path: Path):
        """MoAILoopController should return None for nonexistent loop."""
        storage = LoopStorage(storage_dir=str(tmp_path / ".moai" / "loop"))
        controller = MoAILoopController(storage=storage)

        loaded = controller.get_loop_status("nonexistent")

        assert loaded is None


class TestMoAILoopControllerMaxIterations:
    """Tests for max iteration enforcement (REQ-LOOP-005)."""

    def test_cannot_continue_at_max_iterations(self, tmp_path: Path):
        """MoAILoopController should enforce max iteration limit."""
        storage = LoopStorage(storage_dir=str(tmp_path / ".moai" / "loop"))
        controller = MoAILoopController(storage=storage)

        now = datetime.now(timezone.utc)
        state = LoopState(
            loop_id="max-iterations-loop",
            promise="At limit",
            current_iteration=10,
            max_iterations=10,
            status=LoopStatus.RUNNING,
            created_at=now,
            updated_at=now,
        )

        # The state's can_continue should be False
        assert state.can_continue() is False

    def test_check_completion_at_max_iterations(self, tmp_path: Path):
        """MoAILoopController should mark incomplete at max iterations."""
        storage = LoopStorage(storage_dir=str(tmp_path / ".moai" / "loop"))
        controller = MoAILoopController(storage=storage)

        now = datetime.now(timezone.utc)
        state = LoopState(
            loop_id="max-iterations-loop",
            promise="At limit",
            current_iteration=10,
            max_iterations=10,
            status=LoopStatus.RUNNING,
            created_at=now,
            updated_at=now,
            diagnostics_history=[
                DiagnosticSnapshot(
                    timestamp=now,
                    error_count=5,  # Still has errors
                    warning_count=0,
                    info_count=0,
                    files_affected=["main.py"],
                )
            ],
            ast_issues_history=[],
        )

        result = controller.check_completion(state)

        # Should not be complete because there are still errors
        assert result.is_complete is False
        assert result.remaining_issues > 0


class TestMoAILoopControllerIntegration:
    """Integration tests for MoAILoopController."""

    @pytest.mark.asyncio
    async def test_full_loop_lifecycle(self, tmp_path: Path):
        """MoAILoopController should handle complete loop lifecycle."""
        storage = LoopStorage(storage_dir=str(tmp_path / ".moai" / "loop"))
        controller = MoAILoopController(storage=storage)

        # 1. Start loop
        state = controller.start_loop(promise="Full lifecycle test", max_iterations=5)
        assert state.status == LoopStatus.RUNNING

        # 2. Run feedback loop iteration
        now = datetime.now(timezone.utc)
        with patch.object(controller, "_get_lsp_diagnostics") as mock_lsp:
            with patch.object(controller, "_get_ast_issues") as mock_ast:
                mock_lsp.return_value = DiagnosticSnapshot(
                    timestamp=now,
                    error_count=0,
                    warning_count=0,
                    info_count=0,
                    files_affected=[],
                )
                mock_ast.return_value = ASTIssueSnapshot(
                    timestamp=now,
                    total_issues=0,
                    by_severity={},
                    by_rule={},
                    files_affected=[],
                )

                result = await controller.run_feedback_loop(state)

        assert isinstance(result, FeedbackResult)

        # 3. Check completion
        updated_state = controller.get_loop_status(state.loop_id)
        completion = controller.check_completion(updated_state)

        # Should be complete with no issues
        assert completion.is_complete is True

        # 4. Cancel should fail on completed loop
        # First, mark as completed (in real scenario this would be done by the controller)
        now = datetime.now(timezone.utc)
        completed_state = LoopState(
            loop_id=state.loop_id,
            promise=state.promise,
            current_iteration=state.current_iteration,
            max_iterations=state.max_iterations,
            status=LoopStatus.COMPLETED,
            created_at=state.created_at,
            updated_at=now,
            diagnostics_history=updated_state.diagnostics_history if updated_state else [],
            ast_issues_history=updated_state.ast_issues_history if updated_state else [],
        )
        storage.save_state(completed_state)

        cancel_result = controller.cancel_loop(state.loop_id)
        assert cancel_result is False
