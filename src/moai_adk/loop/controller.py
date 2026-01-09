# Loop Controller
"""Ralph-style autonomous feedback loop controller.

Provides the main controller for orchestrating feedback loops that
integrate LSP diagnostics and AST-grep analysis to guide Claude
through iterative code improvement.
"""

from __future__ import annotations

import uuid
from datetime import datetime, timezone

from moai_adk.loop.feedback import FeedbackGenerator
from moai_adk.loop.state import (
    ASTIssueSnapshot,
    CompletionResult,
    DiagnosticSnapshot,
    FeedbackResult,
    LoopState,
    LoopStatus,
)
from moai_adk.loop.storage import LoopStorage


class MoAILoopController:
    """Ralph-style feedback loop controller.

    Orchestrates autonomous feedback loops that continuously gather
    LSP diagnostics and AST-grep results, providing feedback to Claude
    until the specified completion condition (promise) is met.

    Attributes:
        storage: The storage backend for loop states.
        feedback_generator: Generator for feedback text.
    """

    def __init__(self, storage: LoopStorage | None = None):
        """Initialize the loop controller.

        Args:
            storage: Optional storage backend. If not provided, creates
                a default storage in .moai/loop directory.
        """
        self.storage = storage or LoopStorage()
        self.feedback_generator = FeedbackGenerator()

    def start_loop(
        self,
        promise: str,
        max_iterations: int = 10,
    ) -> LoopState:
        """Start a new feedback loop.

        Creates and persists a new loop state with the specified
        completion condition and iteration limit.

        Args:
            promise: The completion condition for the loop.
            max_iterations: Maximum number of iterations allowed.

        Returns:
            The initial LoopState for the new loop.
        """
        now = datetime.now(timezone.utc)
        loop_id = f"loop-{uuid.uuid4().hex[:8]}"

        state = LoopState(
            loop_id=loop_id,
            promise=promise,
            current_iteration=0,
            max_iterations=max_iterations,
            status=LoopStatus.RUNNING,
            created_at=now,
            updated_at=now,
        )

        self.storage.save_state(state)
        return state

    def check_completion(self, state: LoopState) -> CompletionResult:
        """Check if the loop's completion condition is met.

        Analyzes the current state to determine if the promise
        has been fulfilled (no remaining issues).

        Args:
            state: The current loop state to check.

        Returns:
            CompletionResult indicating completion status and progress.
        """
        # Get the most recent snapshots
        current_lsp_issues = 0
        current_ast_issues = 0

        if state.diagnostics_history:
            latest_diag = state.diagnostics_history[-1]
            current_lsp_issues = latest_diag.total_issues()

        if state.ast_issues_history:
            latest_ast = state.ast_issues_history[-1]
            current_ast_issues = latest_ast.total_issues

        total_issues = current_lsp_issues + current_ast_issues

        # Calculate progress
        initial_issues = self._get_initial_issue_count(state)
        if initial_issues > 0:
            progress = ((initial_issues - total_issues) / initial_issues) * 100
            progress = max(0, min(100, progress))
        else:
            progress = 100.0 if total_issues == 0 else 0.0

        # Check completion
        if total_issues == 0:
            return CompletionResult(
                is_complete=True,
                reason="All issues resolved. Promise fulfilled.",
                remaining_issues=0,
                progress_percentage=100.0,
            )

        # Not complete
        reason = f"{total_issues} issues remaining ({current_lsp_issues} LSP, {current_ast_issues} AST)"
        return CompletionResult(
            is_complete=False,
            reason=reason,
            remaining_issues=total_issues,
            progress_percentage=progress,
        )

    async def run_feedback_loop(self, state: LoopState) -> FeedbackResult:
        """Execute a single feedback loop iteration.

        Gathers current LSP diagnostics and AST-grep issues,
        generates feedback, and updates the loop state.

        Args:
            state: The current loop state.

        Returns:
            FeedbackResult containing the iteration results.
        """
        now = datetime.now(timezone.utc)

        # Get current diagnostics
        lsp_snapshot = self._get_lsp_diagnostics()
        ast_snapshot = self._get_ast_issues()

        # Get previous snapshot for comparison
        previous = None
        if state.diagnostics_history and state.ast_issues_history:
            previous = (state.diagnostics_history[-1], state.ast_issues_history[-1])

        # Generate feedback
        feedback_text = self.feedback_generator.generate_feedback(
            lsp_snapshot,
            ast_snapshot,
            previous_snapshot=previous,
        )

        # Get priority issues
        priority_issues = self.feedback_generator.prioritize_issues([], [])

        # Create feedback result
        result = FeedbackResult(
            lsp_diagnostics=lsp_snapshot,
            ast_issues=ast_snapshot,
            feedback_text=feedback_text,
            priority_issues=priority_issues,
        )

        # Update state with new history
        new_diagnostics_history = list(state.diagnostics_history) + [lsp_snapshot]
        new_ast_history = list(state.ast_issues_history) + [ast_snapshot]

        updated_state = LoopState(
            loop_id=state.loop_id,
            promise=state.promise,
            current_iteration=state.current_iteration + 1,
            max_iterations=state.max_iterations,
            status=state.status,
            created_at=state.created_at,
            updated_at=now,
            diagnostics_history=new_diagnostics_history,
            ast_issues_history=new_ast_history,
        )

        self.storage.save_state(updated_state)

        return result

    def cancel_loop(self, loop_id: str) -> bool:
        """Cancel an active loop.

        Marks the loop as cancelled if it exists and is still active.

        Args:
            loop_id: The ID of the loop to cancel.

        Returns:
            True if the loop was cancelled, False otherwise.
        """
        state = self.storage.load_state(loop_id)
        if state is None:
            return False

        if not state.is_active():
            return False

        now = datetime.now(timezone.utc)
        cancelled_state = LoopState(
            loop_id=state.loop_id,
            promise=state.promise,
            current_iteration=state.current_iteration,
            max_iterations=state.max_iterations,
            status=LoopStatus.CANCELLED,
            created_at=state.created_at,
            updated_at=now,
            diagnostics_history=state.diagnostics_history,
            ast_issues_history=state.ast_issues_history,
        )

        self.storage.save_state(cancelled_state)
        return True

    def get_active_loop(self) -> LoopState | None:
        """Get the currently active loop.

        Returns:
            The active LoopState, or None if no loop is active.
        """
        active_id = self.storage.get_active_loop_id()
        if active_id is None:
            return None
        return self.storage.load_state(active_id)

    def get_loop_status(self, loop_id: str) -> LoopState | None:
        """Get the status of a specific loop.

        Args:
            loop_id: The ID of the loop to query.

        Returns:
            The LoopState, or None if not found.
        """
        return self.storage.load_state(loop_id)

    def _get_initial_issue_count(self, state: LoopState) -> int:
        """Get the initial issue count from the first snapshots.

        Args:
            state: The loop state to analyze.

        Returns:
            The initial total issue count.
        """
        initial_count = 0

        if state.diagnostics_history:
            initial_count += state.diagnostics_history[0].total_issues()

        if state.ast_issues_history:
            initial_count += state.ast_issues_history[0].total_issues

        return initial_count

    def _get_lsp_diagnostics(self) -> DiagnosticSnapshot:
        """Get current LSP diagnostics.

        This method should be overridden or mocked in tests.
        In production, it would integrate with the LSP client.

        Returns:
            DiagnosticSnapshot of current issues.
        """
        # Default implementation returns empty snapshot
        # In production, this would call the LSP client
        return DiagnosticSnapshot(
            timestamp=datetime.now(timezone.utc),
            error_count=0,
            warning_count=0,
            info_count=0,
            files_affected=[],
        )

    def _get_ast_issues(self) -> ASTIssueSnapshot:
        """Get current AST-grep issues.

        This method should be overridden or mocked in tests.
        In production, it would integrate with the AST-grep analyzer.

        Returns:
            ASTIssueSnapshot of current issues.
        """
        # Default implementation returns empty snapshot
        # In production, this would call the AST-grep analyzer
        return ASTIssueSnapshot(
            timestamp=datetime.now(timezone.utc),
            total_issues=0,
            by_severity={},
            by_rule={},
            files_affected=[],
        )
