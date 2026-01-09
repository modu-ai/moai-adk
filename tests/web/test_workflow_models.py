"""Tests for Workflow Data Models

TDD RED phase: These tests define expected behavior for workflow models.
Tests should FAIL initially until models are implemented.
"""

from datetime import datetime

import pytest


class TestWorkflowPhase:
    """Test WorkflowPhase enum."""

    def test_workflow_phase_values(self):
        """Test all expected phase values exist."""
        from moai_adk.web.models.workflow import WorkflowPhase

        assert WorkflowPhase.CONFIGURATION == "configuration"
        assert WorkflowPhase.PLANNING == "planning"
        assert WorkflowPhase.IMPLEMENTATION == "implementation"
        assert WorkflowPhase.SYNC == "sync"
        assert WorkflowPhase.COMPLETION == "completion"

    def test_workflow_phase_iteration(self):
        """Test that all phases are iterable."""
        from moai_adk.web.models.workflow import WorkflowPhase

        phases = list(WorkflowPhase)
        assert len(phases) == 5


class TestWorkflowStatus:
    """Test WorkflowStatus enum."""

    def test_workflow_status_values(self):
        """Test all expected status values exist."""
        from moai_adk.web.models.workflow import WorkflowStatus

        assert WorkflowStatus.PENDING == "pending"
        assert WorkflowStatus.IN_PROGRESS == "in_progress"
        assert WorkflowStatus.CHECKPOINT == "checkpoint"
        assert WorkflowStatus.COMPLETED == "completed"
        assert WorkflowStatus.FAILED == "failed"

    def test_workflow_status_iteration(self):
        """Test that all statuses are iterable."""
        from moai_adk.web.models.workflow import WorkflowStatus

        statuses = list(WorkflowStatus)
        assert len(statuses) == 5


class TestWorkflowConfig:
    """Test WorkflowConfig dataclass."""

    def test_workflow_config_creation_minimal(self):
        """Test creating config with minimal required fields."""
        from moai_adk.web.models.workflow import WorkflowConfig

        config = WorkflowConfig(features=["user auth", "dashboard"])

        assert config.features == ["user auth", "dashboard"]
        assert config.use_worktree is False
        assert config.parallel_workers == 1
        assert config.create_branch is True
        assert config.create_pr is True
        assert config.auto_merge is False
        assert config.model == "glm"
        assert config.mode == "personal"

    def test_workflow_config_creation_full(self):
        """Test creating config with all fields."""
        from moai_adk.web.models.workflow import WorkflowConfig

        config = WorkflowConfig(
            features=["auth", "api"],
            use_worktree=True,
            parallel_workers=3,
            create_branch=False,
            create_pr=False,
            auto_merge=True,
            model="opus",
            mode="team",
        )

        assert config.features == ["auth", "api"]
        assert config.use_worktree is True
        assert config.parallel_workers == 3
        assert config.create_branch is False
        assert config.create_pr is False
        assert config.auto_merge is True
        assert config.model == "opus"
        assert config.mode == "team"

    def test_workflow_config_validation_empty_features(self):
        """Test that empty features list raises validation error."""
        from moai_adk.web.models.workflow import WorkflowConfig
        from pydantic import ValidationError

        with pytest.raises(ValidationError):
            WorkflowConfig(features=[])

    def test_workflow_config_validation_negative_workers(self):
        """Test that negative parallel_workers raises validation error."""
        from moai_adk.web.models.workflow import WorkflowConfig
        from pydantic import ValidationError

        with pytest.raises(ValidationError):
            WorkflowConfig(features=["test"], parallel_workers=-1)

    def test_workflow_config_model_dump(self):
        """Test config serialization to dict."""
        from moai_adk.web.models.workflow import WorkflowConfig

        config = WorkflowConfig(features=["auth"])
        data = config.model_dump()

        assert isinstance(data, dict)
        assert data["features"] == ["auth"]
        assert "use_worktree" in data


class TestSPECResult:
    """Test SPECResult dataclass."""

    def test_spec_result_creation_minimal(self):
        """Test creating result with minimal fields."""
        from moai_adk.web.models.workflow import SPECResult, WorkflowStatus

        result = SPECResult(
            spec_id="SPEC-AUTH-001",
            title="User Authentication",
            status=WorkflowStatus.PENDING,
        )

        assert result.spec_id == "SPEC-AUTH-001"
        assert result.title == "User Authentication"
        assert result.status == WorkflowStatus.PENDING
        assert result.worktree_path is None
        assert result.pr_url is None
        assert result.tokens_used == 0
        assert result.duration_seconds == 0.0

    def test_spec_result_creation_full(self):
        """Test creating result with all fields."""
        from moai_adk.web.models.workflow import SPECResult, WorkflowStatus

        result = SPECResult(
            spec_id="SPEC-AUTH-001",
            title="User Authentication",
            status=WorkflowStatus.COMPLETED,
            worktree_path="/tmp/worktrees/spec-auth-001",
            pr_url="https://github.com/org/repo/pull/123",
            tokens_used=15000,
            duration_seconds=120.5,
        )

        assert result.worktree_path == "/tmp/worktrees/spec-auth-001"
        assert result.pr_url == "https://github.com/org/repo/pull/123"
        assert result.tokens_used == 15000
        assert result.duration_seconds == 120.5

    def test_spec_result_status_transitions(self):
        """Test that spec result status can be updated."""
        from moai_adk.web.models.workflow import SPECResult, WorkflowStatus

        result = SPECResult(
            spec_id="SPEC-001",
            title="Test",
            status=WorkflowStatus.PENDING,
        )

        # Create new instance with updated status (immutable pattern)
        updated = result.model_copy(update={"status": WorkflowStatus.IN_PROGRESS})
        assert updated.status == WorkflowStatus.IN_PROGRESS

    def test_spec_result_model_dump(self):
        """Test result serialization."""
        from moai_adk.web.models.workflow import SPECResult, WorkflowStatus

        result = SPECResult(
            spec_id="SPEC-001",
            title="Test",
            status=WorkflowStatus.COMPLETED,
        )
        data = result.model_dump()

        assert data["spec_id"] == "SPEC-001"
        assert data["status"] == "completed"


class TestWorkflowReport:
    """Test WorkflowReport dataclass."""

    def test_workflow_report_creation(self):
        """Test creating workflow report."""
        from moai_adk.web.models.workflow import (
            SPECResult,
            WorkflowConfig,
            WorkflowPhase,
            WorkflowReport,
            WorkflowStatus,
        )

        config = WorkflowConfig(features=["auth", "api"])
        spec1 = SPECResult(
            spec_id="SPEC-001",
            title="Auth",
            status=WorkflowStatus.COMPLETED,
            tokens_used=10000,
        )
        spec2 = SPECResult(
            spec_id="SPEC-002",
            title="API",
            status=WorkflowStatus.IN_PROGRESS,
            tokens_used=5000,
        )
        started_at = datetime(2024, 1, 10, 10, 0, 0)

        report = WorkflowReport(
            workflow_id="wf-123",
            started_at=started_at,
            completed_at=None,
            config=config,
            specs=[spec1, spec2],
            total_tokens=15000,
            total_cost_usd=0.45,
            phase=WorkflowPhase.IMPLEMENTATION,
            status=WorkflowStatus.IN_PROGRESS,
        )

        assert report.workflow_id == "wf-123"
        assert report.started_at == started_at
        assert report.completed_at is None
        assert len(report.specs) == 2
        assert report.total_tokens == 15000
        assert report.total_cost_usd == 0.45
        assert report.phase == WorkflowPhase.IMPLEMENTATION
        assert report.status == WorkflowStatus.IN_PROGRESS

    def test_workflow_report_with_completion(self):
        """Test workflow report with completion timestamp."""
        from moai_adk.web.models.workflow import (
            WorkflowConfig,
            WorkflowPhase,
            WorkflowReport,
            WorkflowStatus,
        )

        config = WorkflowConfig(features=["test"])
        started_at = datetime(2024, 1, 10, 10, 0, 0)
        completed_at = datetime(2024, 1, 10, 12, 30, 0)

        report = WorkflowReport(
            workflow_id="wf-456",
            started_at=started_at,
            completed_at=completed_at,
            config=config,
            specs=[],
            total_tokens=50000,
            total_cost_usd=1.25,
            phase=WorkflowPhase.COMPLETION,
            status=WorkflowStatus.COMPLETED,
        )

        assert report.completed_at == completed_at
        assert report.phase == WorkflowPhase.COMPLETION
        assert report.status == WorkflowStatus.COMPLETED

    def test_workflow_report_model_dump(self):
        """Test report serialization includes all nested models."""
        from moai_adk.web.models.workflow import (
            SPECResult,
            WorkflowConfig,
            WorkflowPhase,
            WorkflowReport,
            WorkflowStatus,
        )

        config = WorkflowConfig(features=["test"])
        spec = SPECResult(
            spec_id="SPEC-001",
            title="Test",
            status=WorkflowStatus.PENDING,
        )
        report = WorkflowReport(
            workflow_id="wf-789",
            started_at=datetime.now(),
            completed_at=None,
            config=config,
            specs=[spec],
            total_tokens=0,
            total_cost_usd=0.0,
            phase=WorkflowPhase.CONFIGURATION,
            status=WorkflowStatus.PENDING,
        )

        data = report.model_dump()

        assert "workflow_id" in data
        assert "config" in data
        assert isinstance(data["config"], dict)
        assert "specs" in data
        assert len(data["specs"]) == 1
        assert data["specs"][0]["spec_id"] == "SPEC-001"

    def test_workflow_report_json_serialization(self):
        """Test report can be serialized to JSON."""
        from moai_adk.web.models.workflow import (
            WorkflowConfig,
            WorkflowPhase,
            WorkflowReport,
            WorkflowStatus,
        )

        config = WorkflowConfig(features=["test"])
        report = WorkflowReport(
            workflow_id="wf-json",
            started_at=datetime(2024, 1, 10, 10, 0, 0),
            completed_at=None,
            config=config,
            specs=[],
            total_tokens=0,
            total_cost_usd=0.0,
            phase=WorkflowPhase.CONFIGURATION,
            status=WorkflowStatus.PENDING,
        )

        json_str = report.model_dump_json()
        assert isinstance(json_str, str)
        assert "wf-json" in json_str


class TestCheckpointData:
    """Test CheckpointData model for checkpoint approval flow."""

    def test_checkpoint_data_creation(self):
        """Test creating checkpoint data."""
        from moai_adk.web.models.workflow import CheckpointData, WorkflowPhase

        checkpoint = CheckpointData(
            workflow_id="wf-123",
            phase=WorkflowPhase.PLANNING,
            message="SPECs created. Please review and approve.",
            requires_approval=True,
            data={"specs": ["SPEC-001", "SPEC-002"]},
        )

        assert checkpoint.workflow_id == "wf-123"
        assert checkpoint.phase == WorkflowPhase.PLANNING
        assert checkpoint.requires_approval is True
        assert checkpoint.data["specs"] == ["SPEC-001", "SPEC-002"]

    def test_checkpoint_data_optional_approval(self):
        """Test checkpoint without approval requirement."""
        from moai_adk.web.models.workflow import CheckpointData, WorkflowPhase

        checkpoint = CheckpointData(
            workflow_id="wf-456",
            phase=WorkflowPhase.SYNC,
            message="Sync completed.",
            requires_approval=False,
        )

        assert checkpoint.requires_approval is False
        assert checkpoint.data is None


class TestWorkflowError:
    """Test WorkflowError model."""

    def test_workflow_error_creation(self):
        """Test creating workflow error."""
        from moai_adk.web.models.workflow import WorkflowError, WorkflowPhase

        error = WorkflowError(
            workflow_id="wf-error-001",
            phase=WorkflowPhase.IMPLEMENTATION,
            error_code="IMPL_FAILED",
            message="Test execution failed",
            details={"test_file": "test_auth.py", "exit_code": 1},
            recoverable=True,
        )

        assert error.workflow_id == "wf-error-001"
        assert error.phase == WorkflowPhase.IMPLEMENTATION
        assert error.error_code == "IMPL_FAILED"
        assert error.recoverable is True

    def test_workflow_error_non_recoverable(self):
        """Test non-recoverable error."""
        from moai_adk.web.models.workflow import WorkflowError, WorkflowPhase

        error = WorkflowError(
            workflow_id="wf-fatal",
            phase=WorkflowPhase.PLANNING,
            error_code="FATAL_ERROR",
            message="Critical failure",
            recoverable=False,
        )

        assert error.recoverable is False
