"""Tests for Workflow Service

TDD RED phase: Tests for workflow orchestration service.
Tests should FAIL initially until service is implemented.
"""

import pytest


class TestWorkflowOrchestrator:
    """Test WorkflowOrchestrator class."""

    def test_orchestrator_initialization(self):
        """Test orchestrator can be initialized."""
        from moai_adk.web.services.workflow_service import WorkflowOrchestrator

        orchestrator = WorkflowOrchestrator()
        assert orchestrator is not None

    def test_orchestrator_with_cost_tracker(self):
        """Test orchestrator can be initialized with cost tracker."""
        from moai_adk.web.services.cost_tracker import CostTracker
        from moai_adk.web.services.workflow_service import WorkflowOrchestrator

        cost_tracker = CostTracker()
        orchestrator = WorkflowOrchestrator(cost_tracker=cost_tracker)
        assert orchestrator.cost_tracker is cost_tracker

    @pytest.mark.asyncio
    async def test_start_workflow_creates_report(self):
        """Test starting workflow creates initial report."""
        from moai_adk.web.models.workflow import (
            WorkflowConfig,
            WorkflowPhase,
            WorkflowStatus,
        )
        from moai_adk.web.services.workflow_service import WorkflowOrchestrator

        orchestrator = WorkflowOrchestrator()
        config = WorkflowConfig(features=["user authentication"])

        report = await orchestrator.start_workflow(config)

        assert report.workflow_id is not None
        assert report.status == WorkflowStatus.IN_PROGRESS
        assert report.phase == WorkflowPhase.CONFIGURATION
        assert report.config == config

    @pytest.mark.asyncio
    async def test_start_workflow_generates_unique_ids(self):
        """Test each workflow gets unique ID."""
        from moai_adk.web.models.workflow import WorkflowConfig
        from moai_adk.web.services.workflow_service import WorkflowOrchestrator

        orchestrator = WorkflowOrchestrator()
        config = WorkflowConfig(features=["test"])

        report1 = await orchestrator.start_workflow(config)
        report2 = await orchestrator.start_workflow(config)

        assert report1.workflow_id != report2.workflow_id

    @pytest.mark.asyncio
    async def test_get_workflow_status(self):
        """Test retrieving workflow status by ID."""
        from moai_adk.web.models.workflow import WorkflowConfig
        from moai_adk.web.services.workflow_service import WorkflowOrchestrator

        orchestrator = WorkflowOrchestrator()
        config = WorkflowConfig(features=["test"])
        report = await orchestrator.start_workflow(config)

        status = await orchestrator.get_workflow_status(report.workflow_id)

        assert status is not None
        assert status.workflow_id == report.workflow_id

    @pytest.mark.asyncio
    async def test_get_workflow_status_not_found(self):
        """Test getting status for non-existent workflow."""
        from moai_adk.web.services.workflow_service import WorkflowOrchestrator

        orchestrator = WorkflowOrchestrator()
        status = await orchestrator.get_workflow_status("non-existent-id")

        assert status is None

    @pytest.mark.asyncio
    async def test_cancel_workflow(self):
        """Test cancelling a running workflow."""
        from moai_adk.web.models.workflow import WorkflowConfig, WorkflowStatus
        from moai_adk.web.services.workflow_service import WorkflowOrchestrator

        orchestrator = WorkflowOrchestrator()
        config = WorkflowConfig(features=["test"])
        report = await orchestrator.start_workflow(config)

        result = await orchestrator.cancel_workflow(report.workflow_id)

        assert result is True
        status = await orchestrator.get_workflow_status(report.workflow_id)
        assert status.status == WorkflowStatus.FAILED

    @pytest.mark.asyncio
    async def test_cancel_non_existent_workflow(self):
        """Test cancelling non-existent workflow returns False."""
        from moai_adk.web.services.workflow_service import WorkflowOrchestrator

        orchestrator = WorkflowOrchestrator()
        result = await orchestrator.cancel_workflow("non-existent")

        assert result is False

    @pytest.mark.asyncio
    async def test_list_active_workflows(self):
        """Test listing active workflows."""
        from moai_adk.web.models.workflow import WorkflowConfig
        from moai_adk.web.services.workflow_service import WorkflowOrchestrator

        orchestrator = WorkflowOrchestrator()
        config = WorkflowConfig(features=["test"])

        await orchestrator.start_workflow(config)
        await orchestrator.start_workflow(config)

        workflows = await orchestrator.list_active_workflows()

        assert len(workflows) >= 2


class TestCheckpointManager:
    """Test CheckpointManager class."""

    def test_checkpoint_manager_initialization(self):
        """Test checkpoint manager can be initialized."""
        from moai_adk.web.services.workflow_service import CheckpointManager

        manager = CheckpointManager()
        assert manager is not None

    @pytest.mark.asyncio
    async def test_create_checkpoint(self):
        """Test creating a checkpoint."""
        from moai_adk.web.models.workflow import WorkflowPhase
        from moai_adk.web.services.workflow_service import CheckpointManager

        manager = CheckpointManager()

        checkpoint = await manager.create_checkpoint(
            workflow_id="wf-123",
            phase=WorkflowPhase.PLANNING,
            message="Review SPECs before proceeding",
            data={"specs": ["SPEC-001"]},
        )

        assert checkpoint.workflow_id == "wf-123"
        assert checkpoint.phase == WorkflowPhase.PLANNING
        assert checkpoint.requires_approval is True

    @pytest.mark.asyncio
    async def test_get_pending_checkpoint(self):
        """Test getting pending checkpoint for workflow."""
        from moai_adk.web.models.workflow import WorkflowPhase
        from moai_adk.web.services.workflow_service import CheckpointManager

        manager = CheckpointManager()
        await manager.create_checkpoint(
            workflow_id="wf-123",
            phase=WorkflowPhase.PLANNING,
            message="Review",
        )

        checkpoint = await manager.get_pending_checkpoint("wf-123")

        assert checkpoint is not None
        assert checkpoint.workflow_id == "wf-123"

    @pytest.mark.asyncio
    async def test_approve_checkpoint(self):
        """Test approving a checkpoint."""
        from moai_adk.web.models.workflow import WorkflowPhase
        from moai_adk.web.services.workflow_service import CheckpointManager

        manager = CheckpointManager()
        await manager.create_checkpoint(
            workflow_id="wf-123",
            phase=WorkflowPhase.PLANNING,
            message="Review",
        )

        result = await manager.approve_checkpoint("wf-123")

        assert result is True

        # After approval, no pending checkpoint
        checkpoint = await manager.get_pending_checkpoint("wf-123")
        assert checkpoint is None

    @pytest.mark.asyncio
    async def test_approve_non_existent_checkpoint(self):
        """Test approving non-existent checkpoint."""
        from moai_adk.web.services.workflow_service import CheckpointManager

        manager = CheckpointManager()
        result = await manager.approve_checkpoint("non-existent")

        assert result is False

    @pytest.mark.asyncio
    async def test_reject_checkpoint(self):
        """Test rejecting a checkpoint."""
        from moai_adk.web.models.workflow import WorkflowPhase
        from moai_adk.web.services.workflow_service import CheckpointManager

        manager = CheckpointManager()
        await manager.create_checkpoint(
            workflow_id="wf-123",
            phase=WorkflowPhase.PLANNING,
            message="Review",
        )

        result = await manager.reject_checkpoint("wf-123", reason="Not ready")

        assert result is True


class TestParallelExecutor:
    """Test ParallelExecutor class."""

    def test_parallel_executor_initialization(self):
        """Test parallel executor initialization."""
        from moai_adk.web.services.workflow_service import ParallelExecutor

        executor = ParallelExecutor(max_workers=3)
        assert executor.max_workers == 3

    def test_parallel_executor_default_workers(self):
        """Test default worker count."""
        from moai_adk.web.services.workflow_service import ParallelExecutor

        executor = ParallelExecutor()
        assert executor.max_workers == 1

    @pytest.mark.asyncio
    async def test_execute_single_task(self):
        """Test executing a single task."""
        from moai_adk.web.services.workflow_service import ParallelExecutor

        executor = ParallelExecutor(max_workers=1)

        async def sample_task(spec_id: str) -> dict:
            return {"spec_id": spec_id, "status": "completed"}

        results = await executor.execute_tasks(
            tasks=[("SPEC-001", sample_task)],
        )

        assert len(results) == 1
        assert results[0]["spec_id"] == "SPEC-001"

    @pytest.mark.asyncio
    async def test_execute_multiple_tasks_parallel(self):
        """Test executing multiple tasks in parallel."""
        import asyncio

        from moai_adk.web.services.workflow_service import ParallelExecutor

        executor = ParallelExecutor(max_workers=3)
        execution_order = []

        async def sample_task(spec_id: str) -> dict:
            execution_order.append(f"start-{spec_id}")
            await asyncio.sleep(0.01)  # Small delay
            execution_order.append(f"end-{spec_id}")
            return {"spec_id": spec_id, "status": "completed"}

        results = await executor.execute_tasks(
            tasks=[
                ("SPEC-001", sample_task),
                ("SPEC-002", sample_task),
                ("SPEC-003", sample_task),
            ],
        )

        assert len(results) == 3
        # All tasks should start before any finishes (parallel execution)
        # Due to concurrent execution, starts should happen before all ends

    @pytest.mark.asyncio
    async def test_execute_task_with_error(self):
        """Test handling task execution errors."""
        from moai_adk.web.services.workflow_service import ParallelExecutor

        executor = ParallelExecutor(max_workers=1)

        async def failing_task(spec_id: str) -> dict:
            raise ValueError(f"Task {spec_id} failed")

        results = await executor.execute_tasks(
            tasks=[("SPEC-001", failing_task)],
        )

        assert len(results) == 1
        assert "error" in results[0]


class TestPhaseExecutor:
    """Test phase execution logic."""

    @pytest.mark.asyncio
    async def test_execute_configuration_phase(self):
        """Test configuration phase execution."""
        from moai_adk.web.models.workflow import (
            WorkflowConfig,
            WorkflowPhase,
        )
        from moai_adk.web.services.workflow_service import WorkflowOrchestrator

        orchestrator = WorkflowOrchestrator()
        config = WorkflowConfig(features=["test"])
        report = await orchestrator.start_workflow(config)

        # Configuration phase should complete
        updated_report = await orchestrator.execute_phase(
            report.workflow_id,
            WorkflowPhase.CONFIGURATION,
        )

        assert updated_report is not None

    @pytest.mark.asyncio
    async def test_phase_transition(self):
        """Test transitioning between phases."""
        from moai_adk.web.models.workflow import (
            WorkflowConfig,
            WorkflowPhase,
        )
        from moai_adk.web.services.workflow_service import WorkflowOrchestrator

        orchestrator = WorkflowOrchestrator()
        config = WorkflowConfig(features=["test"])
        report = await orchestrator.start_workflow(config)

        # Move to planning phase
        updated = await orchestrator.advance_phase(report.workflow_id)

        assert updated.phase == WorkflowPhase.PLANNING


class TestSPECGeneration:
    """Test SPEC generation within workflow."""

    @pytest.mark.asyncio
    async def test_generate_specs_from_features(self):
        """Test generating SPECs from feature list."""
        from moai_adk.web.models.workflow import WorkflowConfig
        from moai_adk.web.services.workflow_service import WorkflowOrchestrator

        orchestrator = WorkflowOrchestrator()
        config = WorkflowConfig(features=["user auth", "dashboard"])

        specs = await orchestrator.generate_specs(config.features)

        assert len(specs) == 2
        assert specs[0].spec_id.startswith("SPEC-")
        assert specs[1].spec_id.startswith("SPEC-")

    @pytest.mark.asyncio
    async def test_spec_ids_are_sequential(self):
        """Test SPEC IDs are generated sequentially."""
        from moai_adk.web.models.workflow import WorkflowConfig
        from moai_adk.web.services.workflow_service import WorkflowOrchestrator

        orchestrator = WorkflowOrchestrator()
        config = WorkflowConfig(features=["feat1", "feat2", "feat3"])

        specs = await orchestrator.generate_specs(config.features)

        # IDs should be sequential
        ids = [s.spec_id for s in specs]
        assert len(set(ids)) == 3  # All unique


class TestRejectCheckpointNonExistent:
    """Test checkpoint rejection edge cases."""

    @pytest.mark.asyncio
    async def test_reject_non_existent_checkpoint(self):
        """Test rejecting non-existent checkpoint returns False."""
        from moai_adk.web.services.workflow_service import CheckpointManager

        manager = CheckpointManager()
        result = await manager.reject_checkpoint("non-existent-wf")

        assert result is False


class TestPhaseExecution:
    """Test all workflow phase executions."""

    @pytest.mark.asyncio
    async def test_execute_planning_phase(self):
        """Test planning phase generates SPECs."""
        from moai_adk.web.models.workflow import WorkflowConfig, WorkflowPhase
        from moai_adk.web.services.workflow_service import WorkflowOrchestrator

        orchestrator = WorkflowOrchestrator()
        config = WorkflowConfig(features=["auth", "dashboard"])
        report = await orchestrator.start_workflow(config)

        # Execute planning phase
        updated = await orchestrator.execute_phase(
            report.workflow_id,
            WorkflowPhase.PLANNING,
        )

        assert updated is not None
        assert len(updated.specs) == 2

    @pytest.mark.asyncio
    async def test_execute_implementation_phase(self):
        """Test implementation phase execution."""
        from moai_adk.web.models.workflow import WorkflowConfig, WorkflowPhase
        from moai_adk.web.services.workflow_service import WorkflowOrchestrator

        orchestrator = WorkflowOrchestrator()
        config = WorkflowConfig(features=["test"])
        report = await orchestrator.start_workflow(config)

        # Execute implementation phase
        updated = await orchestrator.execute_phase(
            report.workflow_id,
            WorkflowPhase.IMPLEMENTATION,
        )

        assert updated is not None

    @pytest.mark.asyncio
    async def test_execute_sync_phase(self):
        """Test sync phase execution."""
        from moai_adk.web.models.workflow import WorkflowConfig, WorkflowPhase
        from moai_adk.web.services.workflow_service import WorkflowOrchestrator

        orchestrator = WorkflowOrchestrator()
        config = WorkflowConfig(features=["test"])
        report = await orchestrator.start_workflow(config)

        # Execute sync phase
        updated = await orchestrator.execute_phase(
            report.workflow_id,
            WorkflowPhase.SYNC,
        )

        assert updated is not None

    @pytest.mark.asyncio
    async def test_execute_completion_phase(self):
        """Test completion phase marks workflow complete."""
        from moai_adk.web.models.workflow import (
            WorkflowConfig,
            WorkflowPhase,
            WorkflowStatus,
        )
        from moai_adk.web.services.workflow_service import WorkflowOrchestrator

        orchestrator = WorkflowOrchestrator()
        config = WorkflowConfig(features=["test"])
        report = await orchestrator.start_workflow(config)

        # Execute completion phase
        updated = await orchestrator.execute_phase(
            report.workflow_id,
            WorkflowPhase.COMPLETION,
        )

        assert updated is not None
        assert updated.status == WorkflowStatus.COMPLETED
        assert updated.completed_at is not None

    @pytest.mark.asyncio
    async def test_execute_phase_nonexistent_workflow(self):
        """Test executing phase on non-existent workflow."""
        from moai_adk.web.models.workflow import WorkflowPhase
        from moai_adk.web.services.workflow_service import WorkflowOrchestrator

        orchestrator = WorkflowOrchestrator()
        result = await orchestrator.execute_phase(
            "non-existent",
            WorkflowPhase.CONFIGURATION,
        )

        assert result is None


class TestUpdateSpecResult:
    """Test SPEC result update functionality."""

    @pytest.mark.asyncio
    async def test_update_spec_result_success(self):
        """Test updating a SPEC result within workflow."""
        from moai_adk.web.models.workflow import (
            WorkflowConfig,
            WorkflowPhase,
            WorkflowStatus,
        )
        from moai_adk.web.services.workflow_service import WorkflowOrchestrator

        orchestrator = WorkflowOrchestrator()
        config = WorkflowConfig(features=["test feature"])
        report = await orchestrator.start_workflow(config)

        # Execute planning to generate SPECs
        report = await orchestrator.execute_phase(
            report.workflow_id,
            WorkflowPhase.PLANNING,
        )

        spec_id = report.specs[0].spec_id

        # Update the SPEC result
        updated = await orchestrator.update_spec_result(
            report.workflow_id,
            spec_id,
            {"status": WorkflowStatus.COMPLETED, "tokens_used": 1000},
        )

        assert updated is not None
        updated_spec = next(s for s in updated.specs if s.spec_id == spec_id)
        assert updated_spec.status == WorkflowStatus.COMPLETED
        assert updated_spec.tokens_used == 1000

    @pytest.mark.asyncio
    async def test_update_spec_result_nonexistent_workflow(self):
        """Test updating SPEC on non-existent workflow."""
        from moai_adk.web.services.workflow_service import WorkflowOrchestrator

        orchestrator = WorkflowOrchestrator()
        result = await orchestrator.update_spec_result(
            "non-existent",
            "SPEC-001",
            {"status": "completed"},
        )

        assert result is None

    @pytest.mark.asyncio
    async def test_update_spec_result_nonexistent_spec(self):
        """Test updating non-existent SPEC ID preserves others."""
        from moai_adk.web.models.workflow import WorkflowConfig, WorkflowPhase
        from moai_adk.web.services.workflow_service import WorkflowOrchestrator

        orchestrator = WorkflowOrchestrator()
        config = WorkflowConfig(features=["test"])
        report = await orchestrator.start_workflow(config)
        report = await orchestrator.execute_phase(
            report.workflow_id,
            WorkflowPhase.PLANNING,
        )

        # Update non-existent SPEC
        updated = await orchestrator.update_spec_result(
            report.workflow_id,
            "SPEC-NONEXISTENT",
            {"tokens_used": 500},
        )

        # Original specs should be unchanged
        assert updated is not None
        assert len(updated.specs) == len(report.specs)


class TestRecordError:
    """Test error recording functionality."""

    @pytest.mark.asyncio
    async def test_record_recoverable_error(self):
        """Test recording a recoverable error."""
        from moai_adk.web.models.workflow import WorkflowConfig, WorkflowStatus
        from moai_adk.web.services.workflow_service import WorkflowOrchestrator

        orchestrator = WorkflowOrchestrator()
        config = WorkflowConfig(features=["test"])
        report = await orchestrator.start_workflow(config)

        error = await orchestrator.record_error(
            report.workflow_id,
            error_code="VALIDATION_ERROR",
            message="Invalid input",
            details={"field": "features"},
            recoverable=True,
        )

        assert error.error_code == "VALIDATION_ERROR"
        assert error.recoverable is True

        # Workflow should still be in progress
        status = await orchestrator.get_workflow_status(report.workflow_id)
        assert status.status == WorkflowStatus.IN_PROGRESS

    @pytest.mark.asyncio
    async def test_record_non_recoverable_error(self):
        """Test recording non-recoverable error fails workflow."""
        from moai_adk.web.models.workflow import WorkflowConfig, WorkflowStatus
        from moai_adk.web.services.workflow_service import WorkflowOrchestrator

        orchestrator = WorkflowOrchestrator()
        config = WorkflowConfig(features=["test"])
        report = await orchestrator.start_workflow(config)

        error = await orchestrator.record_error(
            report.workflow_id,
            error_code="FATAL_ERROR",
            message="Critical failure",
            recoverable=False,
        )

        assert error.recoverable is False

        # Workflow should be failed
        status = await orchestrator.get_workflow_status(report.workflow_id)
        assert status.status == WorkflowStatus.FAILED
        assert status.completed_at is not None

    @pytest.mark.asyncio
    async def test_record_error_nonexistent_workflow(self):
        """Test recording error for non-existent workflow."""
        from moai_adk.web.models.workflow import WorkflowPhase
        from moai_adk.web.services.workflow_service import WorkflowOrchestrator

        orchestrator = WorkflowOrchestrator()

        error = await orchestrator.record_error(
            "non-existent",
            error_code="TEST_ERROR",
            message="Test",
        )

        # Should use default phase
        assert error.phase == WorkflowPhase.CONFIGURATION


class TestWorkflowIntegration:
    """Integration tests for complete workflow."""

    @pytest.mark.asyncio
    async def test_full_workflow_happy_path(self):
        """Test complete workflow execution (mocked)."""
        from moai_adk.web.models.workflow import (
            WorkflowConfig,
            WorkflowStatus,
        )
        from moai_adk.web.services.workflow_service import WorkflowOrchestrator

        orchestrator = WorkflowOrchestrator()
        config = WorkflowConfig(features=["simple feature"])

        # Start workflow
        report = await orchestrator.start_workflow(config)
        assert report.status == WorkflowStatus.IN_PROGRESS

        # Workflow ID should be tracked
        assert report.workflow_id in orchestrator._workflows

    @pytest.mark.asyncio
    async def test_workflow_with_worktree_enabled(self):
        """Test workflow with worktree support enabled."""
        from moai_adk.web.models.workflow import WorkflowConfig
        from moai_adk.web.services.workflow_service import WorkflowOrchestrator

        orchestrator = WorkflowOrchestrator()
        config = WorkflowConfig(
            features=["feature"],
            use_worktree=True,
            parallel_workers=2,
        )

        report = await orchestrator.start_workflow(config)

        assert report.config.use_worktree is True
        assert report.config.parallel_workers == 2

    @pytest.mark.asyncio
    async def test_workflow_cost_tracking(self):
        """Test that workflow tracks costs."""
        from moai_adk.web.models.workflow import WorkflowConfig
        from moai_adk.web.services.cost_tracker import CostTracker
        from moai_adk.web.services.workflow_service import WorkflowOrchestrator

        cost_tracker = CostTracker()
        orchestrator = WorkflowOrchestrator(cost_tracker=cost_tracker)
        config = WorkflowConfig(features=["test"])

        report = await orchestrator.start_workflow(config)

        # Cost should start at 0
        assert report.total_cost_usd >= 0.0
        assert report.total_tokens >= 0
