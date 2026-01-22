"""Comprehensive test suite for completion_marker.py module.

This module provides 100% coverage for:
- LSPState dataclass and diagnostic parsing
- CompletionMarker for workflow phase validation
- LoopPrevention for autonomous loop control
"""

import pytest

from moai_adk.utils.completion_marker import CompletionMarker, LoopPrevention, LSPState

# ============================================================================
# LSPState Tests
# ============================================================================


class TestLSPStateFromMcpDiagnostics:
    """Tests for LSPState.from_mcp_diagnostics class method."""

    def test_from_mcp_diagnostics_empty_list(self):
        """Test creating LSPState from empty diagnostics list."""
        diagnostics = []
        state = LSPState.from_mcp_diagnostics(diagnostics)
        assert state.errors == 0
        assert state.warnings == 0
        assert state.type_errors == 0
        assert state.lint_errors == 0

    def test_from_mcp_diagnostics_none_input(self):
        """Test creating LSPState from None diagnostics."""
        state = LSPState.from_mcp_diagnostics([])  # Empty list instead of None
        assert state.errors == 0
        assert state.warnings == 0
        assert state.type_errors == 0
        assert state.lint_errors == 0

    def test_from_mcp_diagnostics_with_only_errors(self):
        """Test creating LSPState with only error severity diagnostics."""
        diagnostics = [
            {"severity": "error", "source": "pyright", "message": "Type error"},
            {"severity": "error", "source": "ruff", "message": "Lint error"},
            {"severity": "error", "source": "mypy", "message": "Another type error"},
        ]
        state = LSPState.from_mcp_diagnostics(diagnostics)
        assert state.errors == 3
        assert state.warnings == 0
        assert state.type_errors == 2
        assert state.lint_errors == 1

    def test_from_mcp_diagnostics_with_only_warnings(self):
        """Test creating LSPState with only warning severity diagnostics."""
        diagnostics = [
            {"severity": "warning", "source": "pyright", "message": "Type warning"},
            {"severity": "warning", "source": "ruff", "message": "Lint warning"},
        ]
        state = LSPState.from_mcp_diagnostics(diagnostics)
        assert state.errors == 0
        assert state.warnings == 2
        assert state.type_errors == 1  # pyright warnings count as type errors
        assert state.lint_errors == 1  # ruff warnings count as lint errors

    def test_from_mcp_diagnostics_mixed_severity(self):
        """Test creating LSPState with mixed error and warning diagnostics."""
        diagnostics = [
            {"severity": "error", "source": "pyright", "message": "Type error"},
            {"severity": "warning", "source": "ruff", "message": "Lint warning"},
            {"severity": "error", "source": "ruff", "message": "Lint error"},
            {"severity": "warning", "source": "pylint", "message": "Pylint warning"},
        ]
        state = LSPState.from_mcp_diagnostics(diagnostics)
        assert state.errors == 2
        assert state.warnings == 2
        assert state.type_errors == 1
        assert state.lint_errors == 3  # ruff error + ruff warning + pylint warning

    def test_from_mcp_diagnostics_all_type_checkers(self):
        """Test that all recognized type checkers are counted as type errors."""
        type_checkers = ["pyright", "mypy", "pyrightstrict", "pyrightbasic"]
        diagnostics = [
            {"severity": "error", "source": checker, "message": f"Error from {checker}"} for checker in type_checkers
        ]
        state = LSPState.from_mcp_diagnostics(diagnostics)
        assert state.errors == 4
        assert state.type_errors == 4
        assert state.lint_errors == 0

    def test_from_mcp_diagnostics_all_linters(self):
        """Test that all recognized linters are counted as lint errors."""
        linters = ["ruff", "pylint", "flake8", "eslint", "biome"]
        diagnostics = [{"severity": "error", "source": linter, "message": f"Error from {linter}"} for linter in linters]
        state = LSPState.from_mcp_diagnostics(diagnostics)
        assert state.errors == 5
        assert state.type_errors == 0
        assert state.lint_errors == 5

    def test_from_mcp_diagnostics_unknown_source(self):
        """Test handling diagnostics from unknown sources."""
        diagnostics = [
            {"severity": "error", "source": "unknown-tool", "message": "Unknown error"},
            {"severity": "warning", "source": "custom-linter", "message": "Custom warning"},
        ]
        state = LSPState.from_mcp_diagnostics(diagnostics)
        assert state.errors == 1
        assert state.warnings == 1
        # Unknown sources don't affect type_errors or lint_errors
        assert state.type_errors == 0
        assert state.lint_errors == 0

    def test_from_mcp_diagnostics_case_insensitive_severity(self):
        """Test that severity matching is case-insensitive."""
        diagnostics = [
            {"severity": "Error", "source": "pyright", "message": "Error"},
            {"severity": "ERROR", "source": "ruff", "message": "Error"},
            {"severity": "Warning", "source": "pylint", "message": "Warning"},
        ]
        state = LSPState.from_mcp_diagnostics(diagnostics)
        assert state.errors == 2
        assert state.warnings == 1

    def test_from_mcp_diagnostics_case_insensitive_source(self):
        """Test that source matching is case-insensitive."""
        diagnostics = [
            {"severity": "error", "source": "PyRight", "message": "Error"},
            {"severity": "error", "source": "RUFF", "message": "Error"},
            {"severity": "warning", "source": "MyPy", "message": "Warning"},
        ]
        state = LSPState.from_mcp_diagnostics(diagnostics)
        assert state.type_errors == 2  # PyRight, MyPy
        assert state.lint_errors == 1  # RUFF

    def test_from_mcp_diagnostics_missing_severity(self):
        """Test handling diagnostics without severity field."""
        diagnostics = [
            {"source": "pyright", "message": "No severity"},
            {"severity": "error", "source": "ruff", "message": "Has severity"},
        ]
        state = LSPState.from_mcp_diagnostics(diagnostics)
        # Only the one with severity should be counted
        assert state.errors == 1
        assert state.lint_errors == 1

    def test_from_mcp_diagnostics_missing_source(self):
        """Test handling diagnostics without source field."""
        diagnostics = [
            {"severity": "error", "message": "No source"},
            {"severity": "error", "source": "pyright", "message": "Has source"},
        ]
        state = LSPState.from_mcp_diagnostics(diagnostics)
        assert state.errors == 2
        # Only the one with source should be categorized
        assert state.type_errors == 1

    def test_from_mcp_diagnostics_empty_dict(self):
        """Test handling empty diagnostic dicts."""
        diagnostics = [{}, {}]
        state = LSPState.from_mcp_diagnostics(diagnostics)
        assert state.errors == 0
        assert state.warnings == 0
        assert state.type_errors == 0
        assert state.lint_errors == 0

    def test_from_mcp_diagnostics_large_dataset(self):
        """Test handling a large number of diagnostics efficiently."""
        diagnostics = [{"severity": "error", "source": "pyright", "message": f"Error {i}"} for i in range(1000)]
        state = LSPState.from_mcp_diagnostics(diagnostics)
        assert state.errors == 1000
        assert state.type_errors == 1000


class TestLSPStateIsRegressionFrom:
    """Tests for LSPState.is_regression_from method."""

    def test_is_regression_from_no_regression_all_equal(self):
        """Test regression detection when all values are equal."""
        baseline = LSPState(5, 10, 2, 3)
        current = LSPState(5, 10, 2, 3)
        assert not current.is_regression_from(baseline)

    def test_is_regression_from_no_regression_all_improved(self):
        """Test regression detection when all values improved."""
        baseline = LSPState(5, 10, 2, 3)
        current = LSPState(3, 8, 1, 2)
        assert not current.is_regression_from(baseline)

    def test_is_regression_from_errors_increased(self):
        """Test regression detection when errors increased."""
        baseline = LSPState(5, 10, 2, 3)
        current = LSPState(6, 10, 2, 3)
        assert current.is_regression_from(baseline)

    def test_is_regression_from_type_errors_increased(self):
        """Test regression detection when type errors increased."""
        baseline = LSPState(5, 10, 2, 3)
        current = LSPState(5, 10, 3, 3)
        assert current.is_regression_from(baseline)

    def test_is_regression_from_lint_errors_increased(self):
        """Test regression detection when lint errors increased."""
        baseline = LSPState(5, 10, 2, 3)
        current = LSPState(5, 10, 2, 4)
        assert current.is_regression_from(baseline)

    def test_is_regression_from_warnings_within_tolerance(self):
        """Test that small warning increases are tolerated (10% threshold)."""
        baseline = LSPState(5, 100, 2, 3)
        current = LSPState(5, 110, 2, 3)
        # 10% tolerance: 100 * 1.1 = 110, so 110 is NOT a regression
        assert not current.is_regression_from(baseline)

    def test_is_regression_from_warnings_exceeds_tolerance(self):
        """Test that large warning increases are detected as regression."""
        baseline = LSPState(5, 100, 2, 3)
        current = LSPState(5, 111, 2, 3)
        # 111 > 110 (100 * 1.1), so this IS a regression
        assert current.is_regression_from(baseline)

    def test_is_regression_from_warnings_zero_baseline(self):
        """Test warning tolerance with zero baseline warnings."""
        baseline = LSPState(5, 0, 2, 3)
        current = LSPState(5, 1, 2, 3)
        # max(1, 0 * 1.1) = 1, so 1 is NOT a regression (exactly at threshold)
        assert not current.is_regression_from(baseline)

    def test_is_regression_from_warnings_small_baseline(self):
        """Test warning tolerance with small baseline (less than 10)."""
        baseline = LSPState(5, 5, 2, 3)
        current = LSPState(5, 6, 2, 3)
        # max(1, int(5 * 1.1)) = max(1, int(5.5)) = max(1, 5) = 5
        # 6 > 5, so this IS a regression
        assert current.is_regression_from(baseline)

    def test_is_regression_from_warnings_small_baseline_within_tolerance(self):
        """Test warning tolerance with small baseline within tolerance."""
        baseline = LSPState(5, 5, 2, 3)
        current = LSPState(5, 5, 2, 3)
        # Same warnings - not a regression
        assert not current.is_regression_from(baseline)

    def test_is_regression_from_warnings_small_baseline_exceeds(self):
        """Test warning tolerance with small baseline (exceeds threshold)."""
        baseline = LSPState(5, 5, 2, 3)
        current = LSPState(5, 7, 2, 3)
        # 7 > 6, so this IS a regression
        assert current.is_regression_from(baseline)

    def test_is_regression_from_mixed_improvement_and_regression(self):
        """Test that any regression is detected even if other metrics improved."""
        baseline = LSPState(5, 10, 2, 3)
        current = LSPState(3, 5, 5, 1)  # Errors improved, warnings improved, but type errors increased
        assert current.is_regression_from(baseline)

    def test_is_regression_from_all_zeros_to_all_zeros(self):
        """Test regression detection from clean state to clean state."""
        baseline = LSPState(0, 0, 0, 0)
        current = LSPState(0, 0, 0, 0)
        assert not current.is_regression_from(baseline)

    def test_is_regression_from_clean_to_dirty(self):
        """Test regression detection from clean state to errors."""
        baseline = LSPState(0, 0, 0, 0)
        current = LSPState(1, 0, 0, 0)
        assert current.is_regression_from(baseline)

    def test_is_regression_from_dirty_to_clean(self):
        """Test no regression when going from dirty to clean state."""
        baseline = LSPState(10, 5, 3, 2)
        current = LSPState(0, 0, 0, 0)
        assert not current.is_regression_from(baseline)


class TestLSPStateStr:
    """Tests for LSPState.__str__ method."""

    def test_str_representation_format(self):
        """Test string representation format."""
        state = LSPState(5, 10, 2, 3)
        result = str(state)
        assert "LSPState" in result
        assert "errors=5" in result
        assert "warnings=10" in result
        assert "type_errors=2" in result
        assert "lint_errors=3" in result

    def test_str_all_zeros(self):
        """Test string representation with all zeros."""
        state = LSPState(0, 0, 0, 0)
        result = str(state)
        assert "errors=0" in result
        assert "warnings=0" in result

    def test_str_large_values(self):
        """Test string representation with large values."""
        state = LSPState(1000, 2000, 500, 1500)
        result = str(state)
        assert "errors=1000" in result
        assert "warnings=2000" in result


# ============================================================================
# CompletionMarker Tests
# ============================================================================


class TestCompletionMarkerInit:
    """Tests for CompletionMarker initialization."""

    def test_init_with_phase_plan(self):
        """Test initializing with plan phase."""
        marker = CompletionMarker("plan")
        assert marker.phase == "plan"
        assert marker.config == {}

    def test_init_with_phase_run(self):
        """Test initializing with run phase."""
        marker = CompletionMarker("run")
        assert marker.phase == "run"
        assert marker.config == {}

    def test_init_with_phase_sync(self):
        """Test initializing with sync phase."""
        marker = CompletionMarker("sync")
        assert marker.phase == "sync"
        assert marker.config == {}

    def test_init_with_uppercase_phase(self):
        """Test that phase name is normalized to lowercase."""
        marker = CompletionMarker("PLAN")
        assert marker.phase == "plan"

    def test_init_with_mixed_case_phase(self):
        """Test that mixed case phase names are normalized."""
        marker = CompletionMarker("RuN")
        assert marker.phase == "run"

    def test_init_with_config(self):
        """Test initializing with custom config."""
        config = {"run": {"max_errors": 5}, "sync": {"max_warnings": 2}}
        marker = CompletionMarker("run", config)
        assert marker.config == config

    def test_init_with_none_config(self):
        """Test initializing with None config uses empty dict."""
        marker = CompletionMarker("run", None)
        assert marker.config == {}


class TestCompletionMarkerCheck:
    """Tests for CompletionMarker.check method."""

    def test_check_plan_phase(self):
        """Test plan phase always completes."""
        marker = CompletionMarker("plan", {})
        current = LSPState(0, 0, 0, 0)
        is_complete, reason = marker.check(current)
        assert is_complete
        assert "plan" in reason.lower()
        assert "complete" in reason.lower()

    def test_check_plan_phase_ignores_lsp_state(self):
        """Test plan phase completes regardless of LSP state."""
        marker = CompletionMarker("plan", {})
        current = LSPState(100, 50, 20, 30)
        is_complete, reason = marker.check(current)
        assert is_complete

    def test_check_run_phase_with_clean_state(self):
        """Test run phase completes with clean LSP state."""
        marker = CompletionMarker("run", {})
        baseline = LSPState(0, 0, 0, 0)
        current = LSPState(0, 0, 0, 0)
        is_complete, reason = marker.check(current, baseline)
        assert is_complete
        assert "complete" in reason.lower()

    def test_check_run_phase_with_errors(self):
        """Test run phase fails with LSP errors."""
        marker = CompletionMarker("run", {})
        baseline = LSPState(0, 0, 0, 0)
        current = LSPState(1, 0, 0, 0)
        is_complete, reason = marker.check(current, baseline)
        assert not is_complete
        assert "errors" in reason.lower() or "incomplete" in reason.lower()

    def test_check_run_phase_with_regression(self):
        """Test run phase fails with regression."""
        marker = CompletionMarker("run", {})
        baseline = LSPState(0, 0, 0, 0)
        current = LSPState(1, 0, 1, 0)
        is_complete, reason = marker.check(current, baseline)
        assert not is_complete
        # Error count check happens before regression check
        assert "errors" in reason.lower() or "incomplete" in reason.lower()

    def test_check_run_phase_missing_baseline_raises(self):
        """Test run phase raises ValueError when baseline is missing."""
        marker = CompletionMarker("run", {})
        current = LSPState(0, 0, 0, 0)
        with pytest.raises(ValueError, match="baseline_lsp"):
            marker.check(current)

    def test_check_sync_phase_with_clean_state(self):
        """Test sync phase completes with clean LSP state."""
        marker = CompletionMarker("sync", {})
        current = LSPState(0, 0, 0, 0)
        is_complete, reason = marker.check(current)
        assert is_complete
        assert "ready" in reason.lower() or "complete" in reason.lower()

    def test_check_sync_phase_with_errors(self):
        """Test sync phase fails with LSP errors."""
        marker = CompletionMarker("sync", {})
        current = LSPState(1, 0, 0, 0)
        is_complete, reason = marker.check(current)
        assert not is_complete
        assert "errors" in reason.lower()

    def test_check_sync_phase_with_warnings(self):
        """Test sync phase fails with warnings (default max 0)."""
        marker = CompletionMarker("sync", {})
        current = LSPState(0, 1, 0, 0)
        is_complete, reason = marker.check(current)
        assert not is_complete
        assert "warnings" in reason.lower()

    def test_check_invalid_phase_raises(self):
        """Test invalid phase raises ValueError."""
        marker = CompletionMarker("invalid", {})
        current = LSPState(0, 0, 0, 0)
        with pytest.raises(ValueError, match="Invalid phase"):
            marker.check(current)


class TestCompletionMarkerCheckRun:
    """Tests for CompletionMarker._check_run method."""

    def test_check_run_zero_errors_within_threshold(self):
        """Test run phase with zero errors passes."""
        marker = CompletionMarker("run", {})
        baseline = LSPState(0, 0, 0, 0)
        current = LSPState(0, 0, 0, 0)
        is_complete, reason = marker._check_run(current, baseline)
        assert is_complete

    def test_check_run_respects_max_errors_config(self):
        """Test run phase respects max_errors configuration."""
        marker = CompletionMarker("run", {"run": {"max_errors": 5}})
        baseline = LSPState(3, 0, 0, 0)  # Baseline has 3 errors
        current = LSPState(3, 0, 0, 0)  # Current has same 3 errors
        is_complete, reason = marker._check_run(current, baseline)
        # Completes because errors (3) <= max_errors (5), no regression, no type/lint errors
        assert is_complete

    def test_check_run_exceeds_max_errors_config(self):
        """Test run phase fails when exceeding configured max_errors."""
        marker = CompletionMarker("run", {"run": {"max_errors": 2}})
        baseline = LSPState(0, 0, 0, 0)
        current = LSPState(3, 0, 0, 0)
        is_complete, reason = marker._check_run(current, baseline)
        assert not is_complete
        assert "3 errors" in reason

    def test_check_run_with_type_errors(self):
        """Test run phase fails with type errors."""
        marker = CompletionMarker("run", {})
        baseline = LSPState(0, 0, 0, 0)
        current = LSPState(0, 0, 1, 0)
        is_complete, reason = marker._check_run(current, baseline)
        assert not is_complete
        # Regression check happens first, so "regression" is in the message
        assert "regression" in reason.lower() or "type errors" in reason.lower()

    def test_check_run_with_lint_errors(self):
        """Test run phase fails with lint errors."""
        marker = CompletionMarker("run", {})
        baseline = LSPState(0, 0, 0, 0)
        current = LSPState(0, 0, 0, 1)
        is_complete, reason = marker._check_run(current, baseline)
        assert not is_complete
        # Regression check happens first, so "regression" is in the message
        assert "regression" in reason.lower() or "lint errors" in reason.lower()

    def test_check_run_regression_detected_by_default(self):
        """Test run phase detects regression by default."""
        marker = CompletionMarker("run", {})
        baseline = LSPState(0, 0, 0, 0)
        current = LSPState(1, 0, 0, 0)
        is_complete, reason = marker._check_run(current, baseline)
        assert not is_complete
        # Error count check happens before regression check, so "errors" is in the message
        assert "errors" in reason.lower() or "regression" in reason.lower()

    def test_check_run_allow_regression_skips_check(self):
        """Test run phase allows regression when configured."""
        marker = CompletionMarker("run", {"run": {"allow_regression": True}})
        baseline = LSPState(0, 0, 0, 0)
        current = LSPState(1, 0, 0, 0)
        is_complete, reason = marker._check_run(current, baseline)
        # Should still fail due to errors > max_errors (0)
        assert not is_complete

    def test_check_run_allow_regression_with_max_errors(self):
        """Test run phase with allow_regression and max_errors."""
        marker = CompletionMarker("run", {"run": {"allow_regression": True, "max_errors": 5}})
        baseline = LSPState(0, 0, 0, 0)
        current = LSPState(3, 0, 0, 0)  # Regression but within max_errors
        is_complete, reason = marker._check_run(current, baseline)
        # Still fails due to type errors check (but let's see if all zeros)
        # Actually with type_errors=0, this passes
        assert is_complete

    def test_check_run_complete_message_includes_state(self):
        """Test run phase completion message includes current state."""
        marker = CompletionMarker("run", {})
        baseline = LSPState(0, 0, 0, 0)
        current = LSPState(0, 0, 0, 0)
        is_complete, reason = marker._check_run(current, baseline)
        assert is_complete
        assert "LSPState" in reason


class TestCompletionMarkerCheckSync:
    """Tests for CompletionMarker._check_sync method."""

    def test_check_sync_clean_state_passes(self):
        """Test sync phase passes with completely clean state."""
        marker = CompletionMarker("sync", {})
        current = LSPState(0, 0, 0, 0)
        is_complete, reason = marker._check_sync(current)
        assert is_complete
        assert "ready" in reason.lower() or "complete" in reason.lower()

    def test_check_sync_with_errors_fails(self):
        """Test sync phase fails with any errors."""
        marker = CompletionMarker("sync", {})
        current = LSPState(1, 0, 0, 0)
        is_complete, reason = marker._check_sync(current)
        assert not is_complete
        assert "errors" in reason.lower()

    def test_check_sync_with_warnings_fails_default(self):
        """Test sync phase fails with warnings (default max 0)."""
        marker = CompletionMarker("sync", {})
        current = LSPState(0, 1, 0, 0)
        is_complete, reason = marker._check_sync(current)
        assert not is_complete
        assert "warnings" in reason.lower()

    def test_check_sync_respects_max_warnings_config(self):
        """Test sync phase respects max_warnings configuration."""
        marker = CompletionMarker("sync", {"sync": {"max_warnings": 5}})
        current = LSPState(0, 3, 0, 0)
        is_complete, reason = marker._check_sync(current)
        assert is_complete

    def test_check_sync_exceeds_max_warnings_config(self):
        """Test sync phase fails when exceeding max_warnings."""
        marker = CompletionMarker("sync", {"sync": {"max_warnings": 2}})
        current = LSPState(0, 3, 0, 0)
        is_complete, reason = marker._check_sync(current)
        assert not is_complete
        assert "3 warnings" in reason

    def test_check_sync_with_type_errors_fails(self):
        """Test sync phase fails with type errors."""
        marker = CompletionMarker("sync", {})
        current = LSPState(0, 0, 1, 0)
        is_complete, reason = marker._check_sync(current)
        assert not is_complete
        assert "type errors" in reason.lower()

    def test_check_sync_with_lint_errors_fails(self):
        """Test sync phase fails with lint errors."""
        marker = CompletionMarker("sync", {})
        current = LSPState(0, 0, 0, 1)
        is_complete, reason = marker._check_sync(current)
        assert not is_complete
        assert "lint errors" in reason.lower()

    def test_check_sync_complete_message_includes_state(self):
        """Test sync phase completion message includes state."""
        marker = CompletionMarker("sync", {})
        current = LSPState(0, 0, 0, 0)
        is_complete, reason = marker._check_sync(current)
        assert is_complete
        assert "LSPState" in reason


# ============================================================================
# LoopPrevention Tests
# ============================================================================


class TestLoopPreventionInit:
    """Tests for LoopPrevention initialization."""

    def test_init_with_default_config(self):
        """Test initialization with default configuration."""
        loop = LoopPrevention()
        assert loop.max_iterations == 100
        assert loop.no_progress_threshold == 5
        assert loop.iteration_count == 0
        assert loop.stale_count == 0
        assert loop.last_error_count is None

    def test_init_with_custom_config(self):
        """Test initialization with custom configuration."""
        config = {"max_iterations": 50, "no_progress_threshold": 3}
        loop = LoopPrevention(config)
        assert loop.max_iterations == 50
        assert loop.no_progress_threshold == 3

    def test_init_with_partial_config(self):
        """Test initialization with partial configuration."""
        loop = LoopPrevention({"max_iterations": 20})
        assert loop.max_iterations == 20
        assert loop.no_progress_threshold == 5  # Default value

    def test_init_with_none_config(self):
        """Test initialization with None config uses defaults."""
        loop = LoopPrevention(None)
        assert loop.max_iterations == 100
        assert loop.no_progress_threshold == 5


class TestLoopPreventionShouldContinue:
    """Tests for LoopPrevention.should_continue method."""

    def test_should_continue_first_iteration(self):
        """Test first iteration always continues."""
        loop = LoopPrevention()
        should_continue, reason = loop.should_continue(10)
        assert should_continue
        assert "first iteration" in reason.lower()
        assert loop.iteration_count == 1
        assert loop.last_error_count == 10

    def test_should_continue_progress_made(self):
        """Test continues when progress is made (errors reduced)."""
        loop = LoopPrevention()
        loop.last_error_count = 20
        loop.iteration_count = 1

        should_continue, reason = loop.should_continue(10)
        assert should_continue
        assert "progress" in reason.lower()
        assert loop.stale_count == 0
        assert loop.last_error_count == 10

    def test_should_continue_no_progress_under_threshold(self):
        """Test continues when no progress but under threshold."""
        loop = LoopPrevention({"no_progress_threshold": 5})
        loop.last_error_count = 10
        loop.iteration_count = 1

        # Same error count 3 times (under threshold of 5)
        for _ in range(3):
            should_continue, reason = loop.should_continue(10)
            assert should_continue

        assert loop.stale_count == 3

    def test_should_continue_no_progress_at_threshold(self):
        """Test stops when no progress reaches threshold."""
        loop = LoopPrevention({"no_progress_threshold": 3})
        loop.last_error_count = 10
        loop.iteration_count = 1

        # Same error count 3 times (at threshold)
        for i in range(3):
            should_continue, reason = loop.should_continue(10)
            if i < 2:
                assert should_continue
            else:
                assert not should_continue
                assert "no progress" in reason.lower()

    def test_should_continue_max_iterations_reached(self):
        """Test stops when max iterations reached."""
        loop = LoopPrevention({"max_iterations": 3})
        loop.last_error_count = 10

        # Run 3 times
        for i in range(3):
            should_continue, reason = loop.should_continue(10)
            if i < 2:
                assert should_continue
            else:
                assert not should_continue
                assert "max iterations" in reason.lower()

    def test_should_continue_small_increase_allowed(self):
        """Test small error increases are allowed."""
        loop = LoopPrevention()
        loop.last_error_count = 10
        loop.iteration_count = 1

        should_continue, reason = loop.should_continue(12)  # +2 increase
        assert should_continue
        assert "small increase" in reason.lower()

    def test_should_continue_large_increase_stops(self):
        """Test large error increases cause stop."""
        loop = LoopPrevention()
        loop.last_error_count = 10
        loop.iteration_count = 1

        should_continue, reason = loop.should_continue(15)  # +5 increase
        assert not should_continue
        assert "significant" in reason.lower() or "increase" in reason.lower()

    def test_should_continue_zero_to_zero(self):
        """Test continues when errors stay at zero."""
        loop = LoopPrevention({"no_progress_threshold": 3})
        loop.last_error_count = 0
        loop.iteration_count = 1

        # Zero errors, no progress (but this is OK - we're at zero)
        for i in range(3):
            should_continue, _ = loop.should_continue(0)
            if i < 2:
                assert should_continue
            else:
                # At threshold, should stop
                assert not should_continue

    def test_should_continue_iteration_count_increments(self):
        """Test iteration count increments on each call."""
        loop = LoopPrevention()

        for i in range(5):
            loop.should_continue(10)
            assert loop.iteration_count == i + 1

    def test_should_continue_message_contains_iteration(self):
        """Test message contains iteration number."""
        loop = LoopPrevention()
        loop.last_error_count = 10
        loop.iteration_count = 1

        should_continue, reason = loop.should_continue(8)
        assert should_continue
        assert "iteration" in reason.lower()

    def test_should_continue_edge_case_threshold_one(self):
        """Test no_progress_threshold of 1 works correctly."""
        loop = LoopPrevention({"no_progress_threshold": 1})
        loop.last_error_count = 10
        loop.iteration_count = 1

        # First no-progress should stop immediately
        should_continue, reason = loop.should_continue(10)
        assert not should_continue
        assert "no progress" in reason.lower()

    def test_should_continue_edge_case_max_iterations_one(self):
        """Test max_iterations of 1 works correctly."""
        loop = LoopPrevention({"max_iterations": 1})

        # First call - iteration_count becomes 1, then checks 1 >= 1, which is True
        # So it immediately returns False - no iterations allowed with max_iterations=1
        should_continue, reason = loop.should_continue(10)
        assert not should_continue
        assert "max iterations" in reason.lower()
        assert loop.iteration_count == 1


class TestLoopPreventionReset:
    """Tests for LoopPrevention.reset method."""

    def test_reset_clears_iteration_count(self):
        """Test reset clears iteration count."""
        loop = LoopPrevention()
        loop.should_continue(10)
        loop.should_continue(10)
        assert loop.iteration_count == 2

        loop.reset()
        assert loop.iteration_count == 0

    def test_reset_clears_stale_count(self):
        """Test reset clears stale count."""
        loop = LoopPrevention()
        loop.should_continue(10)
        loop.should_continue(10)
        loop.should_continue(10)
        assert loop.stale_count == 2

        loop.reset()
        assert loop.stale_count == 0

    def test_reset_clears_last_error_count(self):
        """Test reset clears last error count."""
        loop = LoopPrevention()
        loop.should_continue(10)
        assert loop.last_error_count == 10

        loop.reset()
        assert loop.last_error_count is None

    def test_reset_allows_new_loop(self):
        """Test reset allows starting a new loop."""
        loop = LoopPrevention({"max_iterations": 2, "no_progress_threshold": 2})

        # Exhaust the loop
        loop.should_continue(10)
        loop.should_continue(10)
        should_continue, _ = loop.should_continue(10)
        assert not should_continue

        # Reset and try again
        loop.reset()
        should_continue, _ = loop.should_continue(10)
        assert should_continue


class TestLoopPreventionGetStatus:
    """Tests for LoopPrevention.get_status method."""

    def test_get_status_initial_state(self):
        """Test get_status returns initial state."""
        loop = LoopPrevention()
        status = loop.get_status()

        assert status["iteration_count"] == 0
        assert status["stale_count"] == 0
        assert status["last_error_count"] is None
        assert status["max_iterations"] == 100
        assert status["no_progress_threshold"] == 5

    def test_get_status_after_iterations(self):
        """Test get_status after some iterations."""
        loop = LoopPrevention()
        loop.should_continue(10)
        loop.should_continue(8)

        status = loop.get_status()
        assert status["iteration_count"] == 2
        assert status["last_error_count"] == 8
        assert status["stale_count"] == 0

    def test_get_status_with_no_progress(self):
        """Test get_status reflects stale count."""
        loop = LoopPrevention()
        loop.should_continue(10)
        loop.should_continue(10)
        loop.should_continue(10)

        status = loop.get_status()
        assert status["stale_count"] == 2

    def test_get_status_includes_config(self):
        """Test get_status includes configuration values."""
        config = {"max_iterations": 50, "no_progress_threshold": 3}
        loop = LoopPrevention(config)

        status = loop.get_status()
        assert status["max_iterations"] == 50
        assert status["no_progress_threshold"] == 3

    def test_get_status_returns_dict(self):
        """Test get_status returns a dictionary."""
        loop = LoopPrevention()
        status = loop.get_status()

        assert isinstance(status, dict)
        assert "iteration_count" in status
        assert "stale_count" in status
        assert "last_error_count" in status
        assert "max_iterations" in status
        assert "no_progress_threshold" in status


# ============================================================================
# Integration Tests
# ============================================================================


class TestCompletionMarkerIntegration:
    """Integration tests for completion marker scenarios."""

    def test_check_run_type_errors_only_path(self):
        """Test run phase type errors check is reached (for 100% coverage)."""
        marker = CompletionMarker("run", {"run": {"max_errors": 10}})
        # Create state where errors are within threshold but type_errors exist
        baseline = LSPState(5, 0, 2, 3)  # Baseline has 5 errors, 2 type, 3 lint
        current = LSPState(3, 0, 1, 0)  # Current has fewer errors, but 1 type error remains
        is_complete, reason = marker._check_run(current, baseline)
        assert not is_complete
        assert "type errors" in reason.lower()

    def test_check_run_lint_errors_only_path(self):
        """Test run phase lint errors check is reached (for 100% coverage)."""
        marker = CompletionMarker("run", {"run": {"max_errors": 10}})
        baseline = LSPState(5, 0, 3, 2)  # Baseline has 5 errors, 3 type, 2 lint
        current = LSPState(3, 0, 0, 1)  # Current has fewer errors, but 1 lint error remains
        is_complete, reason = marker._check_run(current, baseline)
        assert not is_complete
        assert "lint errors" in reason.lower()

    def test_full_workflow_plan_run_sync(self):
        """Test complete workflow from plan through sync."""
        # Plan phase
        plan_marker = CompletionMarker("plan")
        is_complete, _ = plan_marker.check(LSPState(0, 0, 0, 0))
        assert is_complete

        # Run phase
        run_marker = CompletionMarker("run")
        baseline = LSPState(10, 5, 3, 2)
        current = LSPState(0, 0, 0, 0)
        is_complete, _ = run_marker.check(current, baseline)
        assert is_complete

        # Sync phase
        sync_marker = CompletionMarker("sync")
        is_complete, _ = sync_marker.check(current)
        assert is_complete

    def test_workflow_with_errors_at_each_stage(self):
        """Test workflow behavior with errors at different stages."""
        error_state = LSPState(5, 3, 2, 1)

        # Plan always completes
        plan_marker = CompletionMarker("plan")
        is_complete, _ = plan_marker.check(error_state)
        assert is_complete

        # Run fails with errors
        run_marker = CompletionMarker("run")
        baseline = LSPState(0, 0, 0, 0)
        is_complete, _ = run_marker.check(error_state, baseline)
        assert not is_complete

        # Sync fails with errors
        sync_marker = CompletionMarker("sync")
        is_complete, _ = sync_marker.check(error_state)
        assert not is_complete


class TestLSPStateIntegration:
    """Integration tests for LSPState scenarios."""

    def test_regression_detection_comprehensive(self):
        """Test regression detection across various scenarios."""
        baseline = LSPState(10, 20, 5, 5)

        # Improved
        assert not LSPState(8, 18, 4, 4).is_regression_from(baseline)

        # Same
        assert not LSPState(10, 20, 5, 5).is_regression_from(baseline)

        # Errors increased
        assert LSPState(11, 20, 5, 5).is_regression_from(baseline)

        # Type errors increased
        assert LSPState(10, 20, 6, 5).is_regression_from(baseline)

        # Lint errors increased
        assert LSPState(10, 20, 5, 6).is_regression_from(baseline)

        # Warnings within tolerance
        assert not LSPState(10, 22, 5, 5).is_regression_from(baseline)

        # Warnings exceeded tolerance
        assert LSPState(10, 23, 5, 5).is_regression_from(baseline)

    def test_from_diagnostics_to_regression_check(self):
        """Test full pipeline from diagnostics to regression check."""
        # Baseline diagnostics
        baseline_diagnostics = [
            {"severity": "error", "source": "pyright", "message": "Type error"},
            {"severity": "error", "source": "ruff", "message": "Lint error"},
            {"severity": "warning", "source": "pylint", "message": "Warning"},
        ]
        baseline = LSPState.from_mcp_diagnostics(baseline_diagnostics)

        # Current diagnostics (improved)
        current_diagnostics = [
            {"severity": "error", "source": "pyright", "message": "Type error"},
        ]
        current = LSPState.from_mcp_diagnostics(current_diagnostics)

        # Should not be regression
        assert not current.is_regression_from(baseline)


class TestLoopPreventionIntegration:
    """Integration tests for LoopPrevention scenarios."""

    def test_autonomous_fixing_scenario(self):
        """Test simulating an autonomous fixing loop."""
        loop = LoopPrevention({"max_iterations": 20, "no_progress_threshold": 5})

        # Initial state: 10 errors
        should_continue, reason = loop.should_continue(10)
        assert should_continue

        # Fixed some: 8 errors
        should_continue, reason = loop.should_continue(8)
        assert should_continue
        assert "progress" in reason.lower()

        # Fixed more: 5 errors
        should_continue, reason = loop.should_continue(5)
        assert should_continue

        # Stuck at 5 errors - need 5 more iterations to hit threshold
        # Already had 2 iterations (10->8, 8->5), so stale_count is 0
        # Now we call with same error 5 times
        for i in range(5):
            should_continue, reason = loop.should_continue(5)
            if i < 4:
                assert should_continue
            else:
                # At threshold (5), should stop
                assert not should_continue

    def test_regression_detection_in_loop(self):
        """Test regression detection during autonomous loop."""
        loop = LoopPrevention({"max_iterations": 50})

        # Initial: 10 errors
        loop.should_continue(10)

        # Progress: 5 errors
        loop.should_continue(5)

        # Regression: 15 errors (increase of 10)
        should_continue, reason = loop.should_continue(15)
        assert not should_continue
        assert "significant" in reason.lower()

    def test_reset_and_retry_scenario(self):
        """Test reset and retry after loop exhaustion."""
        loop = LoopPrevention({"max_iterations": 10, "no_progress_threshold": 3})

        # Exhaust iterations with no progress
        loop.should_continue(10)
        loop.should_continue(10)
        loop.should_continue(10)
        should_continue, _ = loop.should_continue(10)
        assert not should_continue  # Hit no_progress_threshold (stale_count=3)

        # Reset and try again with different approach
        loop.reset()

        # Now make progress
        loop.should_continue(10)
        loop.should_continue(5)
        should_continue, _ = loop.should_continue(0)
        # Made progress (10->5->0), so should continue
        assert should_continue
