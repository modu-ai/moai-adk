"""Workflow Data Models

Pydantic models for workflow orchestration including
configuration, SPEC results, and workflow reports.

Part of SPEC-CMD-001: /all-is-well Command Implementation
"""

from datetime import datetime
from enum import Enum
from typing import Any, Optional

from pydantic import BaseModel, ConfigDict, Field, field_validator


class WorkflowPhase(str, Enum):
    """Workflow execution phases.

    Represents the different stages of the /all-is-well workflow:
    - CONFIGURATION: Load settings, parse flags, validate dependencies
    - PLANNING: Analyze requirements, create SPEC documents
    - IMPLEMENTATION: Execute TDD implementation for each SPEC
    - SYNC: Synchronize worktrees, create PRs
    - COMPLETION: Generate summary, cleanup
    """

    CONFIGURATION = "configuration"
    PLANNING = "planning"
    IMPLEMENTATION = "implementation"
    SYNC = "sync"
    COMPLETION = "completion"


class WorkflowStatus(str, Enum):
    """Workflow and SPEC execution status.

    Represents the current state of workflow or individual SPEC:
    - PENDING: Not yet started
    - IN_PROGRESS: Currently executing
    - CHECKPOINT: Waiting for user approval
    - COMPLETED: Successfully finished
    - FAILED: Execution failed
    """

    PENDING = "pending"
    IN_PROGRESS = "in_progress"
    CHECKPOINT = "checkpoint"
    COMPLETED = "completed"
    FAILED = "failed"


class WorkflowConfig(BaseModel):
    """Configuration for workflow execution.

    Attributes:
        features: List of features to implement (required, non-empty)
        use_worktree: Whether to use git worktrees for parallel development
        parallel_workers: Number of parallel workers for implementation
        create_branch: Whether to create feature branches
        create_pr: Whether to create pull requests
        auto_merge: Whether to auto-merge PRs after approval
        model: Default model to use (glm, opus, etc.)
        mode: Execution mode (personal, team)
    """

    model_config = ConfigDict(
        from_attributes=True,
        str_strip_whitespace=True,
    )

    features: list[str] = Field(..., min_length=1, description="Features to implement")
    use_worktree: bool = Field(default=False, description="Use git worktrees")
    parallel_workers: int = Field(default=1, ge=1, description="Parallel worker count")
    create_branch: bool = Field(default=True, description="Create feature branches")
    create_pr: bool = Field(default=True, description="Create pull requests")
    auto_merge: bool = Field(default=False, description="Auto-merge after approval")
    model: str = Field(default="glm", description="Default model")
    mode: str = Field(default="personal", description="Execution mode")

    @field_validator("features")
    @classmethod
    def validate_features(cls, v: list[str]) -> list[str]:
        """Validate that features list is not empty and has valid entries."""
        if not v:
            raise ValueError("Features list cannot be empty")
        # Strip whitespace and filter empty strings
        cleaned = [f.strip() for f in v if f.strip()]
        if not cleaned:
            raise ValueError("Features list cannot contain only empty strings")
        return cleaned


class SPECResult(BaseModel):
    """Result of a single SPEC implementation.

    Tracks the status and metrics for one SPEC within the workflow.

    Attributes:
        spec_id: SPEC identifier (e.g., SPEC-AUTH-001)
        title: Human-readable title
        status: Current execution status
        worktree_path: Path to worktree if using worktrees
        pr_url: URL of created pull request
        tokens_used: Total tokens consumed
        duration_seconds: Execution time in seconds
    """

    model_config = ConfigDict(
        from_attributes=True,
        use_enum_values=True,
    )

    spec_id: str = Field(..., description="SPEC identifier")
    title: str = Field(..., description="SPEC title")
    status: WorkflowStatus = Field(..., description="Execution status")
    worktree_path: Optional[str] = Field(default=None, description="Worktree path")
    pr_url: Optional[str] = Field(default=None, description="Pull request URL")
    tokens_used: int = Field(default=0, ge=0, description="Tokens consumed")
    duration_seconds: float = Field(default=0.0, ge=0.0, description="Duration in seconds")


class WorkflowReport(BaseModel):
    """Complete workflow execution report.

    Contains all information about a workflow run including
    configuration, individual SPEC results, and aggregate metrics.

    Attributes:
        workflow_id: Unique workflow identifier
        started_at: Workflow start timestamp
        completed_at: Workflow completion timestamp (None if still running)
        config: Workflow configuration used
        specs: List of SPEC results
        total_tokens: Total tokens used across all SPECs
        total_cost_usd: Total cost in USD
        phase: Current workflow phase
        status: Current workflow status
    """

    model_config = ConfigDict(
        from_attributes=True,
        use_enum_values=True,
    )

    workflow_id: str = Field(..., description="Workflow identifier")
    started_at: datetime = Field(..., description="Start timestamp")
    completed_at: Optional[datetime] = Field(default=None, description="Completion timestamp")
    config: WorkflowConfig = Field(..., description="Workflow configuration")
    specs: list[SPECResult] = Field(default_factory=list, description="SPEC results")
    total_tokens: int = Field(default=0, ge=0, description="Total tokens used")
    total_cost_usd: float = Field(default=0.0, ge=0.0, description="Total cost in USD")
    phase: WorkflowPhase = Field(..., description="Current phase")
    status: WorkflowStatus = Field(..., description="Current status")


class CheckpointData(BaseModel):
    """Checkpoint data for user approval flow.

    Used when the workflow reaches a checkpoint that requires
    user review and approval before proceeding.

    Attributes:
        workflow_id: Associated workflow identifier
        phase: Phase at which checkpoint was triggered
        message: Human-readable message for user
        requires_approval: Whether user approval is needed
        data: Additional context data for the checkpoint
    """

    model_config = ConfigDict(from_attributes=True)

    workflow_id: str = Field(..., description="Workflow identifier")
    phase: WorkflowPhase = Field(..., description="Checkpoint phase")
    message: str = Field(..., description="Checkpoint message")
    requires_approval: bool = Field(default=True, description="Requires user approval")
    data: Optional[dict[str, Any]] = Field(default=None, description="Additional context")


class WorkflowError(BaseModel):
    """Workflow error information.

    Captures error details when a workflow or SPEC execution fails.

    Attributes:
        workflow_id: Associated workflow identifier
        phase: Phase where error occurred
        error_code: Machine-readable error code
        message: Human-readable error message
        details: Additional error details
        recoverable: Whether the error can be recovered from
    """

    model_config = ConfigDict(from_attributes=True)

    workflow_id: str = Field(..., description="Workflow identifier")
    phase: WorkflowPhase = Field(..., description="Error phase")
    error_code: str = Field(..., description="Error code")
    message: str = Field(..., description="Error message")
    details: Optional[dict[str, Any]] = Field(default=None, description="Error details")
    recoverable: bool = Field(default=False, description="Is error recoverable")
