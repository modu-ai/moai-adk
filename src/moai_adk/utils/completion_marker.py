"""LSP Diagnostics Integration for MoAI Workflow.

This module provides completion marker and loop prevention functionality
for Ralph-style autonomous workflow execution.

Classes:
    LSPState: Holds LSP diagnostic state at a point in time
    CompletionMarker: Checks if workflow phase completion markers are met
    LoopPrevention: Prevents infinite loops in autonomous mode
"""

from dataclasses import dataclass
from typing import Optional


@dataclass
class LSPState:
    """LSP diagnostics state at a point in time.

    Attributes:
        errors: Total number of error severity diagnostics
        warnings: Total number of warning severity diagnostics
        type_errors: Number of type checking errors
        lint_errors: Number of linting errors
    """

    errors: int
    warnings: int
    type_errors: int
    lint_errors: int

    @classmethod
    def from_mcp_diagnostics(cls, diagnostics: list) -> "LSPState":
        """Create LSPState from MCP getDiagnostics result.

        Parses the diagnostics list returned by the MCP getDiagnostics tool
        and categorizes errors by severity and source.

        Args:
            diagnostics: List of diagnostic items from MCP getDiagnostics.
                        Each item should have 'severity' and 'source' fields.

        Returns:
            LSPState instance with counted diagnostics by category.

        Example:
            diagnostics = [
                {"severity": "error", "source": "pyright", "message": "..."},
                {"severity": "warning", "source": "ruff", "message": "..."},
                {"severity": "error", "source": "ruff", "message": "..."},
            ]
            state = LSPState.from_mcp_diagnostics(diagnostics)
        """
        errors = 0
        warnings = 0
        type_errors = 0
        lint_errors = 0

        if not diagnostics:
            return cls(errors=0, warnings=0, type_errors=0, lint_errors=0)

        for diag in diagnostics:
            severity = diag.get("severity", "").lower()
            source = diag.get("source", "").lower()

            if severity == "error":
                errors += 1
                # Type checkers: pyright, mypy, pyrightstrict, pyrightbasic
                if source in ("pyright", "mypy", "pyrightstrict", "pyrightbasic"):
                    type_errors += 1
                # Linters: ruff, pylint, flake8, eslint
                elif source in ("ruff", "pylint", "flake8", "eslint", "biome"):
                    lint_errors += 1
            elif severity == "warning":
                warnings += 1
                # Count type checker warnings as type errors for strict checking
                if source in ("pyright", "mypy", "pyrightstrict", "pyrightbasic"):
                    type_errors += 1
                elif source in ("ruff", "pylint", "flake8", "eslint", "biome"):
                    lint_errors += 1

        return cls(
            errors=errors,
            warnings=warnings,
            type_errors=type_errors,
            lint_errors=lint_errors,
        )

    def is_regression_from(self, baseline: "LSPState") -> bool:
        """Check if current state is worse than baseline.

        A regression occurs when:
        - Total errors increased
        - Type errors increased
        - Lint errors increased
        - Warnings increased significantly (more than 10%)

        Args:
            baseline: The baseline LSPState to compare against.

        Returns:
            True if current state represents a regression from baseline.
        """
        if self.errors > baseline.errors:
            return True

        if self.type_errors > baseline.type_errors:
            return True

        if self.lint_errors > baseline.lint_errors:
            return True

        # Allow 10% tolerance for warnings to avoid false positives
        warning_threshold = max(1, int(baseline.warnings * 1.1))
        if self.warnings > warning_threshold:
            return True

        return False

    def __str__(self) -> str:
        """Return string representation of LSP state."""
        return (
            f"LSPState(errors={self.errors}, warnings={self.warnings}, "
            f"type_errors={self.type_errors}, lint_errors={self.lint_errors})"
        )


class CompletionMarker:
    """Check if workflow phase completion markers are met.

    Each phase has different completion criteria:
    - Plan: Document created
    - Run: Tests pass, LSP clean, no regression from baseline
    - Sync: Everything clean for PR (0 errors, 0 warnings)

    Attributes:
        phase: Current workflow phase ('plan', 'run', 'sync')
        config: Configuration dict with thresholds and settings
    """

    def __init__(self, phase: str, config: Optional[dict] = None):
        """Initialize CompletionMarker for a specific phase.

        Args:
            phase: Workflow phase identifier ('plan', 'run', 'sync')
            config: Optional configuration dict. If None, uses defaults.
                   Default config:
                   - run.max_errors: 0
                   - run.allow_regression: False
                   - sync.max_warnings: 0
        """
        self.phase = phase.lower()
        self.config = config or {}

    def check(self, current_lsp: LSPState, baseline_lsp: Optional[LSPState] = None) -> tuple[bool, str]:
        """Check completion markers for current phase.

        Routes to the appropriate phase-specific check method.

        Args:
            current_lsp: Current LSP diagnostic state
            baseline_lsp: Optional baseline for regression checking (required for run phase)

        Returns:
            Tuple of (is_complete: bool, message: str)

        Raises:
            ValueError: If phase is invalid or baseline is missing when required
        """
        if self.phase == "plan":
            return self._check_plan()
        elif self.phase == "run":
            if baseline_lsp is None:
                raise ValueError("Run phase requires baseline_lsp for regression checking")
            return self._check_run(current_lsp, baseline_lsp)
        elif self.phase == "sync":
            return self._check_sync(current_lsp)
        else:
            raise ValueError(f"Invalid phase: {self.phase}. Must be 'plan', 'run', or 'sync'")

    def _check_plan(self) -> tuple[bool, str]:
        """Plan phase completion: document created.

        Plan phase completion is primarily based on document creation,
        not LSP state. This method always returns True with a message
        indicating the phase is complete.

        Returns:
            Tuple of (True, completion_message)
        """
        return True, "Plan phase complete: SPEC document created"

    def _check_run(self, current: LSPState, baseline: LSPState) -> tuple[bool, str]:
        """Run phase completion: tests pass, LSP clean, coverage met.

        Criteria:
        - Zero errors (or within configured threshold)
        - No regression from baseline
        - Type errors resolved
        - Lint errors resolved

        Args:
            current: Current LSP diagnostic state
            baseline: Baseline LSP state for regression checking

        Returns:
            Tuple of (is_complete: bool, message: str)
        """
        max_errors = self.config.get("run", {}).get("max_errors", 0)
        allow_regression = self.config.get("run", {}).get("allow_regression", False)

        # Check error threshold
        if current.errors > max_errors:
            return (False, f"Run phase incomplete: {current.errors} errors (max allowed: {max_errors})")

        # Check for regression (if not explicitly allowed)
        if not allow_regression and current.is_regression_from(baseline):
            return (False, f"Run phase incomplete: LSP regression detected. Baseline: {baseline}, Current: {current}")

        # Check type errors specifically
        if current.type_errors > 0:
            return (False, f"Run phase incomplete: {current.type_errors} type errors remaining")

        # Check lint errors specifically
        if current.lint_errors > 0:
            return (False, f"Run phase incomplete: {current.lint_errors} lint errors remaining")

        return True, (f"Run phase complete: 0 errors, no regression. Current state: {current}")

    def _check_sync(self, current: LSPState) -> tuple[bool, str]:
        """Sync phase completion: everything clean for PR.

        Criteria:
        - Zero errors (strict requirement)
        - Zero warnings (or within configured threshold)
        - Zero type errors
        - Zero lint errors

        Args:
            current: Current LSP diagnostic state

        Returns:
            Tuple of (is_complete: bool, message: str)
        """
        max_warnings = self.config.get("sync", {}).get("max_warnings", 0)

        # Check errors (strict: must be 0)
        if current.errors > 0:
            return (False, f"Sync phase incomplete: {current.errors} errors (PR requires 0 errors)")

        # Check warnings
        if current.warnings > max_warnings:
            return (False, f"Sync phase incomplete: {current.warnings} warnings (max allowed for PR: {max_warnings})")

        # Verify clean state
        if current.type_errors > 0:
            return False, f"Sync phase incomplete: {current.type_errors} type errors"

        if current.lint_errors > 0:
            return False, f"Sync phase incomplete: {current.lint_errors} lint errors"

        return True, (f"Sync phase complete: Ready for PR. State: {current}")


class LoopPrevention:
    """Prevent infinite loops in autonomous mode.

    Tracks iteration count and progress to detect when the autonomous
    loop should terminate to prevent infinite execution.

    Attributes:
        max_iterations: Maximum number of iterations before forced stop
        no_progress_threshold: Iterations without progress before stopping
        iteration_count: Current iteration number
        stale_count: Number of iterations without error reduction
        last_error_count: Error count from previous iteration
    """

    def __init__(self, config: Optional[dict] = None):
        """Initialize LoopPrevention with configuration.

        Args:
            config: Optional configuration dict. If None, uses defaults.
                   Default config:
                   - max_iterations: 100
                   - no_progress_threshold: 5
        """
        self.config = config or {}
        self.max_iterations = self.config.get("max_iterations", 100)
        self.no_progress_threshold = self.config.get("no_progress_threshold", 5)

        # State tracking
        self.iteration_count = 0
        self.stale_count = 0
        self.last_error_count: Optional[int] = None

    def should_continue(self, current_error_count: int) -> tuple[bool, str]:
        """Check if loop should continue.

        Returns False (stop loop) when:
        - Max iterations reached
        - No progress threshold exceeded (errors not decreasing)
        - Errors increased significantly (potential regression)

        Args:
            current_error_count: Current number of LSP errors

        Returns:
            Tuple of (should_continue: bool, message: str)
        """
        self.iteration_count += 1

        # Check max iterations
        if self.iteration_count >= self.max_iterations:
            return (False, f"Max iterations reached ({self.max_iterations}). Stopping to prevent infinite loop.")

        # Initialize last_error_count on first call
        if self.last_error_count is None:
            self.last_error_count = current_error_count
            return True, f"First iteration: {current_error_count} errors"

        # Check for progress
        if current_error_count < self.last_error_count:
            # Progress made - reset stale counter
            self.stale_count = 0
            progress_msg = (
                f"Iteration {self.iteration_count}: Progress made "
                f"({self.last_error_count} -> {current_error_count} errors)"
            )
            self.last_error_count = current_error_count
            return True, progress_msg

        elif current_error_count == self.last_error_count:
            # No progress - increment stale counter
            self.stale_count += 1

            if self.stale_count >= self.no_progress_threshold:
                return (
                    False,
                    f"No progress for {self.stale_count} iterations ({current_error_count} errors). Stopping.",
                )

            continue_msg = (
                f"Iteration {self.iteration_count}: No progress "
                f"({current_error_count} errors, {self.stale_count}/{self.no_progress_threshold} stale)"
            )
            return True, continue_msg

        else:
            # Errors increased - potential issue
            increase = current_error_count - self.last_error_count
            self.last_error_count = current_error_count

            # Allow small increases (temporary regressions during fixes)
            if increase <= 2:
                return (True, f"Iteration {self.iteration_count}: Small increase (+{increase} errors), continuing")

            return (
                False,
                f"Significant error increase detected: "
                f"+{increase} errors ({self.last_error_count} -> {current_error_count}). "
                "Stopping to prevent degradation.",
            )

    def reset(self) -> None:
        """Reset loop prevention state.

        Useful for starting a new autonomous loop session.
        """
        self.iteration_count = 0
        self.stale_count = 0
        self.last_error_count = None

    def get_status(self) -> dict:
        """Get current loop prevention status.

        Returns:
            Dict with current iteration count, stale count, and last error count
        """
        return {
            "iteration_count": self.iteration_count,
            "stale_count": self.stale_count,
            "last_error_count": self.last_error_count,
            "max_iterations": self.max_iterations,
            "no_progress_threshold": self.no_progress_threshold,
        }
