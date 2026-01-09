# Loop State Tests - RED Phase
"""Tests for loop state data models (LoopStatus, LoopState, snapshots, etc.)."""

from dataclasses import asdict
from datetime import datetime, timezone

from moai_adk.loop.state import (
    ASTIssueSnapshot,
    CompletionResult,
    DiagnosticSnapshot,
    FeedbackResult,
    LoopState,
    LoopStatus,
)


class TestLoopStatus:
    """Tests for LoopStatus enum."""

    def test_status_values(self):
        """LoopStatus should have correct string values."""
        assert LoopStatus.RUNNING.value == "running"
        assert LoopStatus.COMPLETED.value == "completed"
        assert LoopStatus.CANCELLED.value == "cancelled"
        assert LoopStatus.FAILED.value == "failed"
        assert LoopStatus.PAUSED.value == "paused"

    def test_status_is_string_enum(self):
        """LoopStatus should be a string enum for JSON serialization."""
        assert isinstance(LoopStatus.RUNNING.value, str)

    def test_status_from_string(self):
        """LoopStatus should be creatable from string value."""
        status = LoopStatus("running")
        assert status == LoopStatus.RUNNING

    def test_all_status_values_exist(self):
        """All required status values should exist."""
        statuses = [s.value for s in LoopStatus]
        assert "running" in statuses
        assert "completed" in statuses
        assert "cancelled" in statuses
        assert "failed" in statuses
        assert "paused" in statuses


class TestDiagnosticSnapshot:
    """Tests for DiagnosticSnapshot dataclass."""

    def test_create_diagnostic_snapshot(self):
        """DiagnosticSnapshot should store all required fields."""
        now = datetime.now(timezone.utc)
        snapshot = DiagnosticSnapshot(
            timestamp=now,
            error_count=5,
            warning_count=10,
            info_count=3,
            files_affected=["file1.py", "file2.py"],
        )
        assert snapshot.timestamp == now
        assert snapshot.error_count == 5
        assert snapshot.warning_count == 10
        assert snapshot.info_count == 3
        assert snapshot.files_affected == ["file1.py", "file2.py"]

    def test_diagnostic_snapshot_zero_counts(self):
        """DiagnosticSnapshot should accept zero counts."""
        snapshot = DiagnosticSnapshot(
            timestamp=datetime.now(timezone.utc),
            error_count=0,
            warning_count=0,
            info_count=0,
            files_affected=[],
        )
        assert snapshot.error_count == 0
        assert snapshot.warning_count == 0
        assert snapshot.info_count == 0
        assert snapshot.files_affected == []

    def test_diagnostic_snapshot_to_dict(self):
        """DiagnosticSnapshot should be convertible to dict."""
        now = datetime.now(timezone.utc)
        snapshot = DiagnosticSnapshot(
            timestamp=now,
            error_count=1,
            warning_count=2,
            info_count=3,
            files_affected=["test.py"],
        )
        d = asdict(snapshot)
        assert d["error_count"] == 1
        assert d["warning_count"] == 2
        assert d["info_count"] == 3
        assert d["files_affected"] == ["test.py"]

    def test_diagnostic_snapshot_total_issues(self):
        """DiagnosticSnapshot should calculate total issues."""
        snapshot = DiagnosticSnapshot(
            timestamp=datetime.now(timezone.utc),
            error_count=5,
            warning_count=10,
            info_count=3,
            files_affected=["file1.py"],
        )
        assert snapshot.total_issues() == 18

    def test_diagnostic_snapshot_has_errors(self):
        """DiagnosticSnapshot should check if there are errors."""
        with_errors = DiagnosticSnapshot(
            timestamp=datetime.now(timezone.utc),
            error_count=1,
            warning_count=0,
            info_count=0,
            files_affected=["test.py"],
        )
        without_errors = DiagnosticSnapshot(
            timestamp=datetime.now(timezone.utc),
            error_count=0,
            warning_count=5,
            info_count=2,
            files_affected=["test.py"],
        )
        assert with_errors.has_errors() is True
        assert without_errors.has_errors() is False


class TestASTIssueSnapshot:
    """Tests for ASTIssueSnapshot dataclass."""

    def test_create_ast_issue_snapshot(self):
        """ASTIssueSnapshot should store all required fields."""
        now = datetime.now(timezone.utc)
        snapshot = ASTIssueSnapshot(
            timestamp=now,
            total_issues=15,
            by_severity={"error": 5, "warning": 10},
            by_rule={"no-console": 3, "no-eval": 2},
            files_affected=["app.js", "utils.js"],
        )
        assert snapshot.timestamp == now
        assert snapshot.total_issues == 15
        assert snapshot.by_severity == {"error": 5, "warning": 10}
        assert snapshot.by_rule == {"no-console": 3, "no-eval": 2}
        assert snapshot.files_affected == ["app.js", "utils.js"]

    def test_ast_issue_snapshot_empty(self):
        """ASTIssueSnapshot should accept empty dicts and lists."""
        snapshot = ASTIssueSnapshot(
            timestamp=datetime.now(timezone.utc),
            total_issues=0,
            by_severity={},
            by_rule={},
            files_affected=[],
        )
        assert snapshot.total_issues == 0
        assert snapshot.by_severity == {}
        assert snapshot.by_rule == {}
        assert snapshot.files_affected == []

    def test_ast_issue_snapshot_to_dict(self):
        """ASTIssueSnapshot should be convertible to dict."""
        snapshot = ASTIssueSnapshot(
            timestamp=datetime.now(timezone.utc),
            total_issues=5,
            by_severity={"error": 5},
            by_rule={"sql-injection": 5},
            files_affected=["db.py"],
        )
        d = asdict(snapshot)
        assert d["total_issues"] == 5
        assert d["by_severity"] == {"error": 5}
        assert d["by_rule"] == {"sql-injection": 5}

    def test_ast_issue_snapshot_has_critical_issues(self):
        """ASTIssueSnapshot should check for critical (error) issues."""
        with_critical = ASTIssueSnapshot(
            timestamp=datetime.now(timezone.utc),
            total_issues=5,
            by_severity={"error": 3, "warning": 2},
            by_rule={},
            files_affected=["test.py"],
        )
        without_critical = ASTIssueSnapshot(
            timestamp=datetime.now(timezone.utc),
            total_issues=5,
            by_severity={"warning": 3, "info": 2},
            by_rule={},
            files_affected=["test.py"],
        )
        assert with_critical.has_critical_issues() is True
        assert without_critical.has_critical_issues() is False


class TestLoopState:
    """Tests for LoopState dataclass."""

    def test_create_loop_state(self):
        """LoopState should store all required fields."""
        now = datetime.now(timezone.utc)
        state = LoopState(
            loop_id="loop-001",
            promise="Fix all LSP errors",
            current_iteration=1,
            max_iterations=10,
            status=LoopStatus.RUNNING,
            created_at=now,
            updated_at=now,
        )
        assert state.loop_id == "loop-001"
        assert state.promise == "Fix all LSP errors"
        assert state.current_iteration == 1
        assert state.max_iterations == 10
        assert state.status == LoopStatus.RUNNING
        assert state.created_at == now
        assert state.updated_at == now

    def test_loop_state_default_histories(self):
        """LoopState should have empty histories by default."""
        now = datetime.now(timezone.utc)
        state = LoopState(
            loop_id="loop-001",
            promise="Test",
            current_iteration=0,
            max_iterations=5,
            status=LoopStatus.RUNNING,
            created_at=now,
            updated_at=now,
        )
        assert state.diagnostics_history == []
        assert state.ast_issues_history == []

    def test_loop_state_with_history(self):
        """LoopState should store diagnostic and AST histories."""
        now = datetime.now(timezone.utc)
        diag_snapshot = DiagnosticSnapshot(
            timestamp=now,
            error_count=5,
            warning_count=0,
            info_count=0,
            files_affected=["test.py"],
        )
        ast_snapshot = ASTIssueSnapshot(
            timestamp=now,
            total_issues=3,
            by_severity={"warning": 3},
            by_rule={"no-console": 3},
            files_affected=["app.js"],
        )
        state = LoopState(
            loop_id="loop-001",
            promise="Test",
            current_iteration=1,
            max_iterations=10,
            status=LoopStatus.RUNNING,
            created_at=now,
            updated_at=now,
            diagnostics_history=[diag_snapshot],
            ast_issues_history=[ast_snapshot],
        )
        assert len(state.diagnostics_history) == 1
        assert len(state.ast_issues_history) == 1
        assert state.diagnostics_history[0].error_count == 5

    def test_loop_state_is_active(self):
        """LoopState should check if loop is active."""
        now = datetime.now(timezone.utc)
        running = LoopState(
            loop_id="loop-001",
            promise="Test",
            current_iteration=1,
            max_iterations=10,
            status=LoopStatus.RUNNING,
            created_at=now,
            updated_at=now,
        )
        paused = LoopState(
            loop_id="loop-002",
            promise="Test",
            current_iteration=1,
            max_iterations=10,
            status=LoopStatus.PAUSED,
            created_at=now,
            updated_at=now,
        )
        completed = LoopState(
            loop_id="loop-003",
            promise="Test",
            current_iteration=5,
            max_iterations=10,
            status=LoopStatus.COMPLETED,
            created_at=now,
            updated_at=now,
        )
        assert running.is_active() is True
        assert paused.is_active() is True
        assert completed.is_active() is False

    def test_loop_state_can_continue(self):
        """LoopState should check if loop can continue iteration."""
        now = datetime.now(timezone.utc)
        can_continue = LoopState(
            loop_id="loop-001",
            promise="Test",
            current_iteration=5,
            max_iterations=10,
            status=LoopStatus.RUNNING,
            created_at=now,
            updated_at=now,
        )
        at_limit = LoopState(
            loop_id="loop-002",
            promise="Test",
            current_iteration=10,
            max_iterations=10,
            status=LoopStatus.RUNNING,
            created_at=now,
            updated_at=now,
        )
        assert can_continue.can_continue() is True
        assert at_limit.can_continue() is False

    def test_loop_state_to_dict(self):
        """LoopState should be convertible to dict for serialization."""
        now = datetime.now(timezone.utc)
        state = LoopState(
            loop_id="loop-001",
            promise="Test promise",
            current_iteration=1,
            max_iterations=10,
            status=LoopStatus.RUNNING,
            created_at=now,
            updated_at=now,
        )
        d = asdict(state)
        assert d["loop_id"] == "loop-001"
        assert d["promise"] == "Test promise"
        assert d["current_iteration"] == 1
        assert d["max_iterations"] == 10

    def test_loop_state_increment_iteration(self):
        """LoopState should increment iteration count."""
        now = datetime.now(timezone.utc)
        state = LoopState(
            loop_id="loop-001",
            promise="Test",
            current_iteration=1,
            max_iterations=10,
            status=LoopStatus.RUNNING,
            created_at=now,
            updated_at=now,
        )
        new_state = state.increment_iteration()
        assert new_state.current_iteration == 2
        assert new_state.loop_id == state.loop_id
        # Original should be unchanged (immutability)
        assert state.current_iteration == 1


class TestCompletionResult:
    """Tests for CompletionResult dataclass."""

    def test_create_completion_result(self):
        """CompletionResult should store all required fields."""
        result = CompletionResult(
            is_complete=True,
            reason="All LSP errors resolved",
            remaining_issues=0,
            progress_percentage=100.0,
        )
        assert result.is_complete is True
        assert result.reason == "All LSP errors resolved"
        assert result.remaining_issues == 0
        assert result.progress_percentage == 100.0

    def test_completion_result_incomplete(self):
        """CompletionResult should represent incomplete state."""
        result = CompletionResult(
            is_complete=False,
            reason="5 errors remaining",
            remaining_issues=5,
            progress_percentage=50.0,
        )
        assert result.is_complete is False
        assert result.remaining_issues == 5
        assert result.progress_percentage == 50.0

    def test_completion_result_to_dict(self):
        """CompletionResult should be convertible to dict."""
        result = CompletionResult(
            is_complete=False,
            reason="In progress",
            remaining_issues=10,
            progress_percentage=25.0,
        )
        d = asdict(result)
        assert d["is_complete"] is False
        assert d["reason"] == "In progress"
        assert d["remaining_issues"] == 10
        assert d["progress_percentage"] == 25.0


class TestFeedbackResult:
    """Tests for FeedbackResult dataclass."""

    def test_create_feedback_result(self):
        """FeedbackResult should store all required fields."""
        now = datetime.now(timezone.utc)
        lsp_snapshot = DiagnosticSnapshot(
            timestamp=now,
            error_count=2,
            warning_count=3,
            info_count=1,
            files_affected=["main.py"],
        )
        ast_snapshot = ASTIssueSnapshot(
            timestamp=now,
            total_issues=5,
            by_severity={"warning": 5},
            by_rule={"no-console": 5},
            files_affected=["app.js"],
        )
        result = FeedbackResult(
            lsp_diagnostics=lsp_snapshot,
            ast_issues=ast_snapshot,
            feedback_text="Found 2 errors and 5 AST issues",
            priority_issues=["Error in main.py:10", "Warning in app.js:5"],
        )
        assert result.lsp_diagnostics.error_count == 2
        assert result.ast_issues.total_issues == 5
        assert "2 errors" in result.feedback_text
        assert len(result.priority_issues) == 2

    def test_feedback_result_empty_priority(self):
        """FeedbackResult should accept empty priority issues."""
        now = datetime.now(timezone.utc)
        result = FeedbackResult(
            lsp_diagnostics=DiagnosticSnapshot(
                timestamp=now,
                error_count=0,
                warning_count=0,
                info_count=0,
                files_affected=[],
            ),
            ast_issues=ASTIssueSnapshot(
                timestamp=now,
                total_issues=0,
                by_severity={},
                by_rule={},
                files_affected=[],
            ),
            feedback_text="No issues found",
            priority_issues=[],
        )
        assert result.priority_issues == []

    def test_feedback_result_to_dict(self):
        """FeedbackResult should be convertible to dict."""
        now = datetime.now(timezone.utc)
        result = FeedbackResult(
            lsp_diagnostics=DiagnosticSnapshot(
                timestamp=now,
                error_count=1,
                warning_count=0,
                info_count=0,
                files_affected=["test.py"],
            ),
            ast_issues=ASTIssueSnapshot(
                timestamp=now,
                total_issues=1,
                by_severity={"error": 1},
                by_rule={"sql-injection": 1},
                files_affected=["db.py"],
            ),
            feedback_text="Critical issues found",
            priority_issues=["SQL injection in db.py"],
        )
        d = asdict(result)
        assert d["feedback_text"] == "Critical issues found"
        assert d["priority_issues"] == ["SQL injection in db.py"]

    def test_feedback_result_has_issues(self):
        """FeedbackResult should check if there are any issues."""
        now = datetime.now(timezone.utc)
        with_issues = FeedbackResult(
            lsp_diagnostics=DiagnosticSnapshot(
                timestamp=now,
                error_count=1,
                warning_count=0,
                info_count=0,
                files_affected=["test.py"],
            ),
            ast_issues=ASTIssueSnapshot(
                timestamp=now,
                total_issues=0,
                by_severity={},
                by_rule={},
                files_affected=[],
            ),
            feedback_text="Has issues",
            priority_issues=[],
        )
        without_issues = FeedbackResult(
            lsp_diagnostics=DiagnosticSnapshot(
                timestamp=now,
                error_count=0,
                warning_count=0,
                info_count=0,
                files_affected=[],
            ),
            ast_issues=ASTIssueSnapshot(
                timestamp=now,
                total_issues=0,
                by_severity={},
                by_rule={},
                files_affected=[],
            ),
            feedback_text="No issues",
            priority_issues=[],
        )
        assert with_issues.has_issues() is True
        assert without_issues.has_issues() is False
