# Loop State Data Models
"""Data models for the Ralph-style autonomous feedback loop controller.

These dataclasses represent the core types used for tracking loop execution,
diagnostic snapshots, AST issue snapshots, and feedback results.
"""

from __future__ import annotations

from dataclasses import dataclass, field, replace
from datetime import datetime
from enum import Enum


class LoopStatus(Enum):
    """Status of a feedback loop execution.

    Values represent the current state of a loop:
    - RUNNING: Loop is actively executing
    - COMPLETED: Loop finished successfully (promise fulfilled)
    - CANCELLED: Loop was cancelled by user
    - FAILED: Loop failed due to an error
    - PAUSED: Loop is temporarily paused
    """

    RUNNING = "running"
    COMPLETED = "completed"
    CANCELLED = "cancelled"
    FAILED = "failed"
    PAUSED = "paused"


@dataclass
class DiagnosticSnapshot:
    """A snapshot of LSP diagnostics at a point in time.

    Captures the state of LSP diagnostics during a loop iteration,
    including counts by severity and affected files.

    Attributes:
        timestamp: When the snapshot was taken.
        error_count: Number of error-level diagnostics.
        warning_count: Number of warning-level diagnostics.
        info_count: Number of info-level diagnostics.
        files_affected: List of file paths with diagnostics.
    """

    timestamp: datetime
    error_count: int
    warning_count: int
    info_count: int
    files_affected: list[str]

    def total_issues(self) -> int:
        """Calculate total number of issues across all severities.

        Returns:
            Sum of error, warning, and info counts.
        """
        return self.error_count + self.warning_count + self.info_count

    def has_errors(self) -> bool:
        """Check if there are any error-level diagnostics.

        Returns:
            True if error_count > 0, False otherwise.
        """
        return self.error_count > 0


@dataclass
class ASTIssueSnapshot:
    """A snapshot of AST-grep issues at a point in time.

    Captures the state of AST-grep analysis during a loop iteration,
    including issue counts by severity and rule.

    Attributes:
        timestamp: When the snapshot was taken.
        total_issues: Total number of AST issues found.
        by_severity: Count of issues by severity level.
        by_rule: Count of issues by rule ID.
        files_affected: List of file paths with issues.
    """

    timestamp: datetime
    total_issues: int
    by_severity: dict[str, int]
    by_rule: dict[str, int]
    files_affected: list[str]

    def has_critical_issues(self) -> bool:
        """Check if there are any critical (error) issues.

        Returns:
            True if there are error-severity issues, False otherwise.
        """
        return self.by_severity.get("error", 0) > 0


@dataclass
class LoopState:
    """State of a feedback loop execution.

    Tracks the complete state of a feedback loop including iteration
    count, status, and history of diagnostics and AST issues.

    Attributes:
        loop_id: Unique identifier for the loop.
        promise: The completion condition (e.g., "Fix all LSP errors").
        current_iteration: Current iteration number (1-based).
        max_iterations: Maximum allowed iterations.
        status: Current loop status.
        created_at: When the loop was created.
        updated_at: When the loop was last updated.
        diagnostics_history: History of diagnostic snapshots.
        ast_issues_history: History of AST issue snapshots.
    """

    loop_id: str
    promise: str
    current_iteration: int
    max_iterations: int
    status: LoopStatus
    created_at: datetime
    updated_at: datetime
    diagnostics_history: list[DiagnosticSnapshot] = field(default_factory=list)
    ast_issues_history: list[ASTIssueSnapshot] = field(default_factory=list)

    def is_active(self) -> bool:
        """Check if the loop is in an active state.

        Active states are RUNNING and PAUSED. Completed, cancelled,
        and failed loops are not active.

        Returns:
            True if the loop is active, False otherwise.
        """
        return self.status in (LoopStatus.RUNNING, LoopStatus.PAUSED)

    def can_continue(self) -> bool:
        """Check if the loop can continue to the next iteration.

        A loop can continue if it is active and has not reached
        the maximum iteration limit.

        Returns:
            True if the loop can continue, False otherwise.
        """
        return self.is_active() and self.current_iteration < self.max_iterations

    def increment_iteration(self) -> LoopState:
        """Create a new LoopState with incremented iteration count.

        Returns a new LoopState instance with current_iteration + 1,
        preserving immutability of the original state.

        Returns:
            New LoopState with incremented iteration.
        """
        return replace(self, current_iteration=self.current_iteration + 1)


@dataclass
class CompletionResult:
    """Result of checking loop completion conditions.

    Contains information about whether the loop's promise has been
    fulfilled and progress towards completion.

    Attributes:
        is_complete: Whether the completion condition is met.
        reason: Human-readable explanation of the result.
        remaining_issues: Number of issues still remaining.
        progress_percentage: Percentage of progress towards completion.
    """

    is_complete: bool
    reason: str
    remaining_issues: int
    progress_percentage: float


@dataclass
class FeedbackResult:
    """Result of executing a feedback loop iteration.

    Contains the diagnostic and AST snapshots from the iteration,
    along with formatted feedback text and priority issues.

    Attributes:
        lsp_diagnostics: Snapshot of LSP diagnostics.
        ast_issues: Snapshot of AST-grep issues.
        feedback_text: Formatted feedback text for Claude.
        priority_issues: List of highest-priority issues to address.
    """

    lsp_diagnostics: DiagnosticSnapshot
    ast_issues: ASTIssueSnapshot
    feedback_text: str
    priority_issues: list[str]

    def has_issues(self) -> bool:
        """Check if there are any issues (LSP or AST).

        Returns:
            True if there are any LSP diagnostics or AST issues.
        """
        return self.lsp_diagnostics.total_issues() > 0 or self.ast_issues.total_issues > 0
