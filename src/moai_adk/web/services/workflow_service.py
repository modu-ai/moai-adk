"""Workflow Service

Service for orchestrating the /all-is-well workflow including
configuration, planning, implementation, sync, and completion phases.

Part of SPEC-CMD-001: /all-is-well Command Implementation
"""

import asyncio
from datetime import datetime
from typing import Any, Callable, Coroutine, Optional
from uuid import uuid4

from moai_adk.web.models.workflow import (
    CheckpointData,
    SPECResult,
    WorkflowConfig,
    WorkflowError,
    WorkflowPhase,
    WorkflowReport,
    WorkflowStatus,
)
from moai_adk.web.services.cost_tracker import CostTracker


class CheckpointManager:
    """Manages workflow checkpoints for user approval.

    Handles checkpoint creation, approval, and rejection during
    workflow execution.

    Attributes:
        _checkpoints: Dictionary of pending checkpoints by workflow_id
    """

    def __init__(self) -> None:
        """Initialize the checkpoint manager."""
        self._checkpoints: dict[str, CheckpointData] = {}

    async def create_checkpoint(
        self,
        workflow_id: str,
        phase: WorkflowPhase,
        message: str,
        data: Optional[dict[str, Any]] = None,
        requires_approval: bool = True,
    ) -> CheckpointData:
        """Create a new checkpoint for workflow.

        Args:
            workflow_id: Workflow identifier
            phase: Current workflow phase
            message: Checkpoint message for user
            data: Additional context data
            requires_approval: Whether user approval is needed

        Returns:
            Created CheckpointData instance
        """
        checkpoint = CheckpointData(
            workflow_id=workflow_id,
            phase=phase,
            message=message,
            requires_approval=requires_approval,
            data=data,
        )
        self._checkpoints[workflow_id] = checkpoint
        return checkpoint

    async def get_pending_checkpoint(self, workflow_id: str) -> Optional[CheckpointData]:
        """Get pending checkpoint for workflow.

        Args:
            workflow_id: Workflow identifier

        Returns:
            CheckpointData if pending checkpoint exists, None otherwise
        """
        return self._checkpoints.get(workflow_id)

    async def approve_checkpoint(self, workflow_id: str) -> bool:
        """Approve a pending checkpoint.

        Args:
            workflow_id: Workflow identifier

        Returns:
            True if checkpoint was approved, False if not found
        """
        if workflow_id not in self._checkpoints:
            return False

        del self._checkpoints[workflow_id]
        return True

    async def reject_checkpoint(self, workflow_id: str, reason: Optional[str] = None) -> bool:
        """Reject a pending checkpoint.

        Args:
            workflow_id: Workflow identifier
            reason: Optional rejection reason

        Returns:
            True if checkpoint was rejected, False if not found
        """
        if workflow_id not in self._checkpoints:
            return False

        del self._checkpoints[workflow_id]
        return True


class ParallelExecutor:
    """Executes tasks in parallel with configurable worker count.

    Manages parallel execution of SPEC implementations with
    semaphore-based concurrency control.

    Attributes:
        max_workers: Maximum number of parallel workers
    """

    def __init__(self, max_workers: int = 1) -> None:
        """Initialize the parallel executor.

        Args:
            max_workers: Maximum concurrent tasks (default 1)
        """
        self.max_workers = max_workers

    async def execute_tasks(
        self,
        tasks: list[tuple[str, Callable[[str], Coroutine[Any, Any, dict]]]],
    ) -> list[dict]:
        """Execute multiple tasks in parallel.

        Args:
            tasks: List of (spec_id, async_function) tuples

        Returns:
            List of results from each task
        """
        semaphore = asyncio.Semaphore(self.max_workers)

        async def run_with_semaphore(
            spec_id: str,
            task_func: Callable[[str], Coroutine[Any, Any, dict]],
        ) -> dict:
            async with semaphore:
                try:
                    return await task_func(spec_id)
                except Exception as e:
                    return {"spec_id": spec_id, "error": str(e)}

        coroutines = [run_with_semaphore(spec_id, func) for spec_id, func in tasks]
        results = await asyncio.gather(*coroutines, return_exceptions=False)

        return list(results)


class WorkflowOrchestrator:
    """Orchestrates the complete /all-is-well workflow.

    Manages workflow lifecycle including phase execution,
    checkpoint handling, and cost tracking.

    Attributes:
        cost_tracker: Optional cost tracking service
        checkpoint_manager: Checkpoint management service
        _workflows: Dictionary of active workflows
        _spec_counter: Counter for generating SPEC IDs
    """

    # Phase order for transitions
    PHASE_ORDER = [
        WorkflowPhase.CONFIGURATION,
        WorkflowPhase.PLANNING,
        WorkflowPhase.IMPLEMENTATION,
        WorkflowPhase.SYNC,
        WorkflowPhase.COMPLETION,
    ]

    def __init__(
        self,
        cost_tracker: Optional[CostTracker] = None,
    ) -> None:
        """Initialize the workflow orchestrator.

        Args:
            cost_tracker: Optional cost tracking service
        """
        self.cost_tracker = cost_tracker
        self.checkpoint_manager = CheckpointManager()
        self._workflows: dict[str, WorkflowReport] = {}
        self._spec_counter = 0

    def _generate_workflow_id(self) -> str:
        """Generate unique workflow ID.

        Returns:
            Unique workflow identifier
        """
        return f"wf-{uuid4().hex[:12]}"

    def _generate_spec_id(self) -> str:
        """Generate unique SPEC ID.

        Returns:
            SPEC identifier in format SPEC-XXX
        """
        self._spec_counter += 1
        return f"SPEC-{self._spec_counter:03d}"

    async def start_workflow(self, config: WorkflowConfig) -> WorkflowReport:
        """Start a new workflow.

        Creates a new workflow with the given configuration and
        initializes it in CONFIGURATION phase.

        Args:
            config: Workflow configuration

        Returns:
            Initial WorkflowReport
        """
        workflow_id = self._generate_workflow_id()

        report = WorkflowReport(
            workflow_id=workflow_id,
            started_at=datetime.now(),
            completed_at=None,
            config=config,
            specs=[],
            total_tokens=0,
            total_cost_usd=0.0,
            phase=WorkflowPhase.CONFIGURATION,
            status=WorkflowStatus.IN_PROGRESS,
        )

        self._workflows[workflow_id] = report
        return report

    async def get_workflow_status(self, workflow_id: str) -> Optional[WorkflowReport]:
        """Get current workflow status.

        Args:
            workflow_id: Workflow identifier

        Returns:
            WorkflowReport if found, None otherwise
        """
        return self._workflows.get(workflow_id)

    async def cancel_workflow(self, workflow_id: str) -> bool:
        """Cancel a running workflow.

        Args:
            workflow_id: Workflow identifier

        Returns:
            True if cancelled, False if not found
        """
        if workflow_id not in self._workflows:
            return False

        report = self._workflows[workflow_id]
        updated = report.model_copy(
            update={
                "status": WorkflowStatus.FAILED,
                "completed_at": datetime.now(),
            }
        )
        self._workflows[workflow_id] = updated
        return True

    async def list_active_workflows(self) -> list[WorkflowReport]:
        """List all active workflows.

        Returns:
            List of active workflow reports
        """
        return [report for report in self._workflows.values() if report.status == WorkflowStatus.IN_PROGRESS]

    async def execute_phase(
        self,
        workflow_id: str,
        phase: WorkflowPhase,
    ) -> Optional[WorkflowReport]:
        """Execute a specific workflow phase.

        Args:
            workflow_id: Workflow identifier
            phase: Phase to execute

        Returns:
            Updated WorkflowReport or None if not found
        """
        report = self._workflows.get(workflow_id)
        if not report:
            return None

        # Execute phase-specific logic
        if phase == WorkflowPhase.CONFIGURATION:
            # Configuration phase - validate setup
            pass
        elif phase == WorkflowPhase.PLANNING:
            # Planning phase - generate SPECs
            specs = await self.generate_specs(report.config.features)
            report = report.model_copy(update={"specs": specs})
        elif phase == WorkflowPhase.IMPLEMENTATION:
            # Implementation phase - execute TDD
            pass
        elif phase == WorkflowPhase.SYNC:
            # Sync phase - merge and create PRs
            pass
        elif phase == WorkflowPhase.COMPLETION:
            # Completion phase - finalize
            report = report.model_copy(
                update={
                    "status": WorkflowStatus.COMPLETED,
                    "completed_at": datetime.now(),
                }
            )

        self._workflows[workflow_id] = report
        return report

    async def advance_phase(self, workflow_id: str) -> Optional[WorkflowReport]:
        """Advance workflow to next phase.

        Args:
            workflow_id: Workflow identifier

        Returns:
            Updated WorkflowReport or None if not found
        """
        report = self._workflows.get(workflow_id)
        if not report:
            return None

        current_idx = self.PHASE_ORDER.index(report.phase)
        if current_idx < len(self.PHASE_ORDER) - 1:
            next_phase = self.PHASE_ORDER[current_idx + 1]
            report = report.model_copy(update={"phase": next_phase})
            self._workflows[workflow_id] = report

        return report

    async def generate_specs(self, features: list[str]) -> list[SPECResult]:
        """Generate SPECs from feature list.

        Creates a SPECResult for each feature with unique IDs.

        Args:
            features: List of feature descriptions

        Returns:
            List of SPECResult instances
        """
        specs = []
        for feature in features:
            spec_id = self._generate_spec_id()
            spec = SPECResult(
                spec_id=spec_id,
                title=feature,
                status=WorkflowStatus.PENDING,
            )
            specs.append(spec)

        return specs

    async def update_spec_result(
        self,
        workflow_id: str,
        spec_id: str,
        updates: dict[str, Any],
    ) -> Optional[WorkflowReport]:
        """Update a specific SPEC result within workflow.

        Args:
            workflow_id: Workflow identifier
            spec_id: SPEC identifier
            updates: Dictionary of fields to update

        Returns:
            Updated WorkflowReport or None if not found
        """
        report = self._workflows.get(workflow_id)
        if not report:
            return None

        updated_specs = []
        for spec in report.specs:
            if spec.spec_id == spec_id:
                updated_specs.append(spec.model_copy(update=updates))
            else:
                updated_specs.append(spec)

        report = report.model_copy(update={"specs": updated_specs})
        self._workflows[workflow_id] = report
        return report

    async def record_error(
        self,
        workflow_id: str,
        error_code: str,
        message: str,
        details: Optional[dict[str, Any]] = None,
        recoverable: bool = False,
    ) -> WorkflowError:
        """Record an error for workflow.

        Args:
            workflow_id: Workflow identifier
            error_code: Machine-readable error code
            message: Human-readable message
            details: Additional error details
            recoverable: Whether error can be recovered from

        Returns:
            WorkflowError instance
        """
        report = self._workflows.get(workflow_id)
        phase = report.phase if report else WorkflowPhase.CONFIGURATION

        error = WorkflowError(
            workflow_id=workflow_id,
            phase=phase,
            error_code=error_code,
            message=message,
            details=details,
            recoverable=recoverable,
        )

        # If not recoverable, mark workflow as failed
        if not recoverable and report:
            updated = report.model_copy(
                update={
                    "status": WorkflowStatus.FAILED,
                    "completed_at": datetime.now(),
                }
            )
            self._workflows[workflow_id] = updated

        return error
