"""Workflow Router

REST API endpoints for workflow orchestration including
starting, monitoring, and managing /all-is-well workflows.

Part of SPEC-CMD-001: /all-is-well Command Implementation
"""

from typing import Optional

from fastapi import APIRouter, Depends, HTTPException, status
from pydantic import BaseModel, Field

from moai_adk.web.models.workflow import (
    WorkflowConfig,
    WorkflowReport,
)
from moai_adk.web.services.workflow_service import WorkflowOrchestrator

router = APIRouter(tags=["workflow"])

# Global orchestrator instance (dependency injection)
_orchestrator: Optional[WorkflowOrchestrator] = None


def get_orchestrator() -> WorkflowOrchestrator:
    """Dependency to get the workflow orchestrator.

    Returns:
        WorkflowOrchestrator instance
    """
    global _orchestrator
    if _orchestrator is None:
        _orchestrator = WorkflowOrchestrator()
    return _orchestrator


# Request/Response Models
class StartWorkflowRequest(BaseModel):
    """Request model for starting a workflow."""

    features: list[str] = Field(..., min_length=1, description="Features to implement")
    use_worktree: bool = Field(default=False, description="Use git worktrees")
    parallel_workers: int = Field(default=1, ge=1, description="Parallel worker count")
    create_branch: bool = Field(default=True, description="Create feature branches")
    create_pr: bool = Field(default=True, description="Create pull requests")
    auto_merge: bool = Field(default=False, description="Auto-merge after approval")
    model: str = Field(default="glm", description="Default model to use")


class WorkflowResponse(BaseModel):
    """Response model for workflow operations."""

    workflow_id: str
    status: str
    phase: str
    config: dict
    specs: list[dict] = Field(default_factory=list)
    total_tokens: int = 0
    total_cost_usd: float = 0.0


class WorkflowListResponse(BaseModel):
    """Response model for workflow list."""

    workflows: list[WorkflowResponse]


class CancelResponse(BaseModel):
    """Response model for cancel operation."""

    cancelled: bool
    workflow_id: str


class CheckpointApproveResponse(BaseModel):
    """Response model for checkpoint approval."""

    approved: bool
    workflow_id: str


class RejectCheckpointRequest(BaseModel):
    """Request model for rejecting a checkpoint."""

    reason: Optional[str] = None


def _report_to_response(report: WorkflowReport) -> WorkflowResponse:
    """Convert WorkflowReport to API response.

    Args:
        report: WorkflowReport instance

    Returns:
        WorkflowResponse for API
    """
    return WorkflowResponse(
        workflow_id=report.workflow_id,
        status=report.status if isinstance(report.status, str) else report.status.value,
        phase=report.phase if isinstance(report.phase, str) else report.phase.value,
        config=report.config.model_dump(),
        specs=[s.model_dump() for s in report.specs],
        total_tokens=report.total_tokens,
        total_cost_usd=report.total_cost_usd,
    )


@router.post(
    "/start",
    response_model=WorkflowResponse,
    status_code=status.HTTP_201_CREATED,
    summary="Start a new workflow",
    description="Start a new /all-is-well workflow with the specified features.",
)
async def start_workflow(
    request: StartWorkflowRequest,
    orchestrator: WorkflowOrchestrator = Depends(get_orchestrator),
) -> WorkflowResponse:
    """Start a new workflow.

    Args:
        request: Workflow configuration request
        orchestrator: Workflow orchestrator service

    Returns:
        Initial workflow status
    """
    config = WorkflowConfig(
        features=request.features,
        use_worktree=request.use_worktree,
        parallel_workers=request.parallel_workers,
        create_branch=request.create_branch,
        create_pr=request.create_pr,
        auto_merge=request.auto_merge,
        model=request.model,
    )

    report = await orchestrator.start_workflow(config)
    return _report_to_response(report)


@router.get(
    "/{workflow_id}/status",
    response_model=WorkflowResponse,
    summary="Get workflow status",
    description="Get the current status of a workflow.",
)
async def get_workflow_status(
    workflow_id: str,
    orchestrator: WorkflowOrchestrator = Depends(get_orchestrator),
) -> WorkflowResponse:
    """Get workflow status.

    Args:
        workflow_id: Workflow identifier
        orchestrator: Workflow orchestrator service

    Returns:
        Current workflow status

    Raises:
        HTTPException: If workflow not found
    """
    report = await orchestrator.get_workflow_status(workflow_id)
    if report is None:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Workflow {workflow_id} not found",
        )
    return _report_to_response(report)


@router.delete(
    "/{workflow_id}",
    response_model=CancelResponse,
    summary="Cancel a workflow",
    description="Cancel a running workflow.",
)
async def cancel_workflow(
    workflow_id: str,
    orchestrator: WorkflowOrchestrator = Depends(get_orchestrator),
) -> CancelResponse:
    """Cancel a workflow.

    Args:
        workflow_id: Workflow identifier
        orchestrator: Workflow orchestrator service

    Returns:
        Cancellation result

    Raises:
        HTTPException: If workflow not found
    """
    result = await orchestrator.cancel_workflow(workflow_id)
    if not result:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Workflow {workflow_id} not found",
        )
    return CancelResponse(cancelled=True, workflow_id=workflow_id)


@router.get(
    "/list",
    response_model=WorkflowListResponse,
    summary="List active workflows",
    description="List all active workflows.",
)
async def list_workflows(
    orchestrator: WorkflowOrchestrator = Depends(get_orchestrator),
) -> WorkflowListResponse:
    """List active workflows.

    Args:
        orchestrator: Workflow orchestrator service

    Returns:
        List of active workflows
    """
    workflows = await orchestrator.list_active_workflows()
    return WorkflowListResponse(workflows=[_report_to_response(w) for w in workflows])


@router.post(
    "/{workflow_id}/checkpoint/approve",
    response_model=CheckpointApproveResponse,
    summary="Approve a checkpoint",
    description="Approve a pending checkpoint to continue workflow.",
)
async def approve_checkpoint(
    workflow_id: str,
    orchestrator: WorkflowOrchestrator = Depends(get_orchestrator),
) -> CheckpointApproveResponse:
    """Approve a checkpoint.

    Args:
        workflow_id: Workflow identifier
        orchestrator: Workflow orchestrator service

    Returns:
        Approval result

    Raises:
        HTTPException: If checkpoint not found
    """
    result = await orchestrator.checkpoint_manager.approve_checkpoint(workflow_id)
    if not result:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Checkpoint for workflow {workflow_id} not found",
        )
    return CheckpointApproveResponse(approved=True, workflow_id=workflow_id)


@router.post(
    "/{workflow_id}/checkpoint/reject",
    response_model=CheckpointApproveResponse,
    summary="Reject a checkpoint",
    description="Reject a pending checkpoint to stop or modify workflow.",
)
async def reject_checkpoint(
    workflow_id: str,
    request: RejectCheckpointRequest,
    orchestrator: WorkflowOrchestrator = Depends(get_orchestrator),
) -> CheckpointApproveResponse:
    """Reject a checkpoint.

    Args:
        workflow_id: Workflow identifier
        request: Rejection request with optional reason
        orchestrator: Workflow orchestrator service

    Returns:
        Rejection result

    Raises:
        HTTPException: If checkpoint not found
    """
    result = await orchestrator.checkpoint_manager.reject_checkpoint(workflow_id, request.reason)
    if not result:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Checkpoint for workflow {workflow_id} not found",
        )
    return CheckpointApproveResponse(approved=False, workflow_id=workflow_id)


@router.post(
    "/{workflow_id}/advance",
    response_model=WorkflowResponse,
    summary="Advance workflow phase",
    description="Advance workflow to the next phase.",
)
async def advance_workflow(
    workflow_id: str,
    orchestrator: WorkflowOrchestrator = Depends(get_orchestrator),
) -> WorkflowResponse:
    """Advance workflow to next phase.

    Args:
        workflow_id: Workflow identifier
        orchestrator: Workflow orchestrator service

    Returns:
        Updated workflow status

    Raises:
        HTTPException: If workflow not found
    """
    report = await orchestrator.advance_phase(workflow_id)
    if report is None:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Workflow {workflow_id} not found",
        )
    return _report_to_response(report)
