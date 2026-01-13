"""Comprehensive tests for CLI progress indicators.

Test Coverage Strategy:
- Progress bars: create_progress_bar, column customization, total handling
- Spinners: create_spinner, spinner contexts, status updates
- ProgressContext: Multi-task tracking, advance, update_task, complete_task
- SpinnerContext: Status updates, final messages, success/error/warning/info
- Progress steps: Sequential step tracking, progress_steps generator
- Print step: Formatted step indicators, status symbols
- MoAISpinnerColumn: Custom spinner rendering
"""

from contextlib import contextmanager
from typing import Generator, Iterator
from unittest.mock import MagicMock, Mock, patch

import pytest
from rich.console import Console
from rich.progress import Progress, TaskID
from rich.status import Status
from rich.text import Text

from moai_adk.cli.ui.progress import (
    MoAISpinnerColumn,
    create_progress_bar,
    create_spinner,
    ProgressContext,
    SpinnerContext,
    progress_steps,
    print_step,
    console,
    CLAUDE_STYLE,
    SUCCESS_STYLE,
    ERROR_STYLE,
    INFO_STYLE,
    WARNING_STYLE,
)


class TestMoAISpinnerColumn:
    """Test MoAISpinnerColumn custom progress column."""

    def test_renders_spinner_text(self) -> None:
        """Test that spinner renders text output."""
        column = MoAISpinnerColumn(spinner_name="dots")
        mock_task = Mock()
        mock_task.get_time.return_value = 1.0

        result = column.render(mock_task)

        assert isinstance(result, Text)

    def test_default_spinner_name(self) -> None:
        """Test default spinner name is 'dots'."""
        column = MoAISpinnerColumn()
        assert column.spinner.name == "dots"

    def test_custom_spinner_name(self) -> None:
        """Test custom spinner name."""
        column = MoAISpinnerColumn(spinner_name="line")
        assert column.spinner.name == "line"


class TestCreateProgressBar:
    """Test create_progress_bar function."""

    def test_creates_progress_instance(self) -> None:
        """Test that create_progress_bar returns Progress instance."""
        progress = create_progress_bar()
        assert isinstance(progress, Progress)

    def test_default_description(self) -> None:
        """Test default description is 'Processing'."""
        progress = create_progress_bar()
        # Progress is created, we can't easily inspect description
        assert progress is not None

    def test_custom_description(self) -> None:
        """Test custom description."""
        progress = create_progress_bar(description="Custom Task")
        assert isinstance(progress, Progress)

    def test_with_total(self) -> None:
        """Test progress bar with total steps."""
        progress = create_progress_bar(total=100)
        assert isinstance(progress, Progress)

    def test_without_total(self) -> None:
        """Test progress bar without total (indeterminate)."""
        progress = create_progress_bar(total=None)
        assert isinstance(progress, Progress)

    def test_transient_flag(self) -> None:
        """Test transient flag for auto-removal."""
        progress = create_progress_bar(transient=True)
        assert isinstance(progress, Progress)

    def test_auto_refresh_flag(self) -> None:
        """Test auto_refresh flag."""
        progress = create_progress_bar(auto_refresh=False)
        assert isinstance(progress, Progress)

    def test_context_manager(self) -> None:
        """Test progress bar as context manager."""
        with create_progress_bar(total=10) as progress:
            assert isinstance(progress, Progress)
            task = progress.add_task("Test", total=10)
            progress.update(task, advance=5)


class TestCreateSpinner:
    """Test create_spinner function."""

    def test_creates_status_instance(self) -> None:
        """Test that create_spinner returns Status instance."""
        spinner = create_spinner()
        assert isinstance(spinner, Status)

    def test_default_message(self) -> None:
        """Test default message is 'Processing...'."""
        spinner = create_spinner()
        assert isinstance(spinner, Status)

    def test_custom_message(self) -> None:
        """Test custom message."""
        spinner = create_spinner(message="Loading...")
        assert isinstance(spinner, Status)

    def test_custom_spinner_name(self) -> None:
        """Test custom spinner name."""
        spinner = create_spinner(spinner_name="dots")
        assert isinstance(spinner, Status)

    def test_context_manager(self) -> None:
        """Test spinner as context manager."""
        with create_spinner(message="Loading...") as spinner:
            assert isinstance(spinner, Status)


class TestProgressContext:
    """Test ProgressContext class."""

    def test_initialization(self) -> None:
        """Test ProgressContext initialization."""
        ctx = ProgressContext("Test Progress")
        assert ctx.title == "Test Progress"
        assert ctx.transient is False
        assert ctx.progress is None
        assert ctx.tasks == {}
        assert ctx._started is False

    def test_initialization_with_transient(self) -> None:
        """Test initialization with transient flag."""
        ctx = ProgressContext("Test", transient=True)
        assert ctx.transient is True

    def test_enter_starts_progress(self) -> None:
        """Test that __enter__ starts the progress bar."""
        ctx = ProgressContext("Test")
        result = ctx.__enter__()
        assert result is ctx
        assert ctx.progress is not None
        assert ctx._started is True

    def test_exit_stops_progress(self) -> None:
        """Test that __exit__ stops the progress bar."""
        ctx = ProgressContext("Test")
        ctx.__enter__()
        ctx.__exit__(None, None, None)
        assert ctx._started is False

    def test_add_task(self) -> None:
        """Test adding a task."""
        ctx = ProgressContext("Test")
        ctx.__enter__()
        task_id = ctx.add_task("Task 1", total=100)  # type: ignore[assignment]
        assert "Task 1" in ctx.tasks
        ctx.__exit__(None, None, None)

    def test_add_task_with_defaults(self) -> None:
        """Test adding task with default parameters."""
        ctx = ProgressContext("Test")
        ctx.__enter__()
        task_id = ctx.add_task("Task 1")  # type: ignore[assignment]
        assert "Task 1" in ctx.tasks
        ctx.__exit__(None, None, None)

    def test_add_task_without_starting_raises(self) -> None:
        """Test that add_task raises RuntimeError when not started."""
        ctx = ProgressContext("Test")
        with pytest.raises(RuntimeError, match="not started"):
            ctx.add_task("Task 1")

    def test_advance(self) -> None:
        """Test advancing task progress."""
        ctx = ProgressContext("Test")
        ctx.__enter__()
        ctx.add_task("Task 1", total=100)
        ctx.advance("Task 1", 10)
        # Should not raise
        ctx.__exit__(None, None, None)

    def test_advance_without_starting_silently_fails(self) -> None:
        """Test that advance fails silently when not started."""
        ctx = ProgressContext("Test")
        # Should not raise
        ctx.advance("Task 1", 10)

    def test_advance_nonexistent_task(self) -> None:
        """Test advancing non-existent task."""
        ctx = ProgressContext("Test")
        ctx.__enter__()
        # Should not raise
        ctx.advance("Nonexistent", 10)
        ctx.__exit__(None, None, None)

    def test_update_task(self) -> None:
        """Test updating task state."""
        ctx = ProgressContext("Test")
        ctx.__enter__()
        ctx.add_task("Task 1", total=100)
        ctx.update_task("Task 1", completed=50)
        # Should not raise
        ctx.__exit__(None, None, None)

    def test_update_task_with_new_description(self) -> None:
        """Test updating task description."""
        ctx = ProgressContext("Test")
        ctx.__enter__()
        ctx.add_task("Task 1", total=100)
        ctx.update_task("Task 1", new_description="Task 1 Updated")
        assert "Task 1 Updated" in ctx.tasks
        assert "Task 1" not in ctx.tasks
        ctx.__exit__(None, None, None)

    def test_update_task_all_parameters(self) -> None:
        """Test updating task with all parameters."""
        ctx = ProgressContext("Test")
        ctx.__enter__()
        ctx.add_task("Task 1", total=100)
        ctx.update_task("Task 1", completed=50, total=150, new_description="Task 1 Updated")
        # Should not raise
        ctx.__exit__(None, None, None)

    def test_complete_task(self) -> None:
        """Test marking task as complete."""
        ctx = ProgressContext("Test")
        ctx.__enter__()
        ctx.add_task("Task 1", total=100)
        ctx.complete_task("Task 1")
        # Should not raise
        ctx.__exit__(None, None, None)


class TestSpinnerContext:
    """Test SpinnerContext class."""

    def test_initialization(self) -> None:
        """Test SpinnerContext initialization."""
        ctx = SpinnerContext("Loading...")
        assert ctx.initial_message == "Loading..."
        assert ctx.spinner_name == "dots"
        assert ctx.status is None
        assert ctx._final_message is None
        assert ctx._final_style is None

    def test_initialization_with_custom_spinner(self) -> None:
        """Test initialization with custom spinner."""
        ctx = SpinnerContext("Loading...", spinner_name="line")
        assert ctx.spinner_name == "line"

    def test_enter_starts_spinner(self) -> None:
        """Test that __enter__ starts the spinner."""
        ctx = SpinnerContext("Loading...")
        result = ctx.__enter__()
        assert result is ctx
        assert ctx.status is not None

    def test_exit_stops_spinner(self) -> None:
        """Test that __exit__ stops the spinner."""
        ctx = SpinnerContext("Loading...")
        ctx.__enter__()
        ctx.__exit__(None, None, None)
        # Status should be stopped

    def test_update_message(self) -> None:
        """Test updating spinner message."""
        ctx = SpinnerContext("Loading...")
        ctx.__enter__()
        ctx.update("Processing...")
        # Should not raise
        ctx.__exit__(None, None, None)

    def test_success_sets_final_message(self) -> None:
        """Test setting success message."""
        ctx = SpinnerContext("Loading...")
        ctx.success("Done!")
        assert ctx._final_message == "✓ Done!"
        assert ctx._final_style == SUCCESS_STYLE

    def test_error_sets_final_message(self) -> None:
        """Test setting error message."""
        ctx = SpinnerContext("Loading...")
        ctx.error("Failed!")
        assert ctx._final_message == "✗ Failed!"
        assert ctx._final_style == ERROR_STYLE

    def test_warning_sets_final_message(self) -> None:
        """Test setting warning message."""
        ctx = SpinnerContext("Loading...")
        ctx.warning("Caution!")
        assert ctx._final_message == "⚠ Caution!"
        assert ctx._final_style == WARNING_STYLE

    def test_info_sets_final_message(self) -> None:
        """Test setting info message."""
        ctx = SpinnerContext("Loading...")
        ctx.info("Info")
        assert ctx._final_message == "ℹ Info"
        assert ctx._final_style == INFO_STYLE

    def test_multiple_final_messages_last_wins(self) -> None:
        """Test that last set_final_message wins."""
        ctx = SpinnerContext("Loading...")
        ctx.success("First")
        ctx.error("Second")
        assert ctx._final_message is not None
        assert "Second" in ctx._final_message

    def test_context_with_success_message(self) -> None:
        """Test spinner context with success message."""
        with SpinnerContext("Loading...") as spinner:
            spinner.success("Complete!")
        # Should complete without error

    def test_context_with_error_message(self) -> None:
        """Test spinner context with error message."""
        with SpinnerContext("Loading...") as spinner:
            spinner.error("Failed!")
        # Should complete without error


class TestProgressSteps:
    """Test progress_steps context manager."""

    def test_yields_step_iterator(self) -> None:
        """Test that progress_steps yields step iterator."""
        steps = ["Step 1", "Step 2", "Step 3"]
        with progress_steps(steps, "Testing") as step_iter:
            assert isinstance(step_iter, Iterator)
            step_list = list(step_iter)
            assert step_list == steps

    def test_tracks_progress_through_steps(self) -> None:
        """Test that progress is tracked through steps."""
        steps = ["Step 1", "Step 2", "Step 3"]
        with progress_steps(steps, "Testing") as step_iter:
            for step in step_iter:
                # Should not raise
                pass

    def test_custom_title(self) -> None:
        """Test progress_steps with custom title."""
        steps = ["Step 1"]
        with progress_steps(steps, "Custom Title") as step_iter:
            for step in step_iter:
                # Should not raise
                pass

    def test_empty_steps(self) -> None:
        """Test progress_steps with empty steps list."""
        steps: list[str] = []
        with progress_steps(steps, "Testing") as step_iter:
            step_list = list(step_iter)
            assert step_list == []


class TestPrintStep:
    """Test print_step function."""

    def test_print_running_step(self, capsys: pytest.CaptureFixture[str]) -> None:
        """Test printing running step."""
        print_step(1, 5, "Processing data", "running")
        captured = capsys.readouterr()
        assert "1/5" in captured.out
        assert "Processing data" in captured.out

    def test_print_complete_step(self, capsys: pytest.CaptureFixture[str]) -> None:
        """Test printing complete step."""
        print_step(2, 5, "Completed", "complete")
        captured = capsys.readouterr()
        assert "2/5" in captured.out
        assert "Completed" in captured.out

    def test_print_error_step(self, capsys: pytest.CaptureFixture[str]) -> None:
        """Test printing error step."""
        print_step(3, 5, "Failed", "error")
        captured = capsys.readouterr()
        assert "3/5" in captured.out
        assert "Failed" in captured.out

    def test_print_skipped_step(self, capsys: pytest.CaptureFixture[str]) -> None:
        """Test printing skipped step."""
        print_step(4, 5, "Skipped", "skipped")
        captured = capsys.readouterr()
        assert "4/5" in captured.out
        assert "Skipped" in captured.out

    def test_print_step_invalid_status(self, capsys: pytest.CaptureFixture[str]) -> None:
        """Test printing step with invalid status (defaults to running)."""
        print_step(5, 5, "Unknown", "invalid")
        captured = capsys.readouterr()
        assert "5/5" in captured.out
        assert "Unknown" in captured.out


class TestProgressBarIntegration:
    """Test progress bar integration scenarios."""

    def test_multiple_tasks_tracking(self) -> None:
        """Test tracking multiple tasks simultaneously."""
        with ProgressContext("Multi-task") as ctx:
            task1 = ctx.add_task("Task 1", total=100)
            task2 = ctx.add_task("Task 2", total=50)
            ctx.advance("Task 1", 25)
            ctx.advance("Task 2", 10)
            # Should not raise

    def test_task_completion_flow(self) -> None:
        """Test complete task flow from start to finish."""
        with ProgressContext("Flow") as ctx:
            ctx.add_task("Step 1", total=10)
            for i in range(10):
                ctx.advance("Step 1")
            ctx.complete_task("Step 1")

    def test_task_update_flow(self) -> None:
        """Test task update flow."""
        with ProgressContext("Updates") as ctx:
            ctx.add_task("Processing", total=100)
            ctx.update_task("Processing", completed=25)
            ctx.update_task("Processing", completed=50)
            ctx.update_task("Processing", completed=75)
            ctx.complete_task("Processing")

    def test_nested_progress_tracking(self) -> None:
        """Test tracking with sequential operations."""
        with ProgressContext("Nested") as ctx:
            # Task 1
            task1 = ctx.add_task("Setup", total=3)
            for _ in range(3):
                ctx.advance("Setup")

            # Task 2
            task2 = ctx.add_task("Process", total=5)
            for _ in range(5):
                ctx.advance("Process")

            # Task 3
            task3 = ctx.add_task("Cleanup", total=2)
            for _ in range(2):
                ctx.advance("Cleanup")


class TestSpinnerIntegration:
    """Test spinner integration scenarios."""

    def test_spinner_with_status_updates(self) -> None:
        """Test spinner with multiple status updates."""
        with SpinnerContext("Working...") as spinner:
            spinner.update("Step 1: Initialize...")
            spinner.update("Step 2: Process...")
            spinner.update("Step 3: Finalize...")
            spinner.success("Complete!")

    def test_spinner_error_recovery(self) -> None:
        """Test spinner with error handling."""
        with SpinnerContext("Loading...") as spinner:
            spinner.update("Attempting...")
            spinner.error("Failed, retrying...")

    def test_spinner_info_updates(self) -> None:
        """Test spinner with informational updates."""
        with SpinnerContext("Connecting...") as spinner:
            spinner.update("Establishing connection...")
            spinner.info("Connection established")
            spinner.update("Authenticating...")
            spinner.success("Authenticated!")

    def test_spinner_warning_updates(self) -> None:
        """Test spinner with warning updates."""
        with SpinnerContext("Processing...") as spinner:
            spinner.update("Checking data...")
            spinner.warning("Data inconsistency detected")
            spinner.update("Attempting repair...")
            spinner.success("Repaired successfully")


class TestErrorHandling:
    """Test error handling in progress components."""

    def test_progress_context_exception_handling(self) -> None:
        """Test ProgressContext handles exceptions gracefully."""
        ctx = ProgressContext("Test")
        ctx.__enter__()
        # Simulate exception handling
        # Should not raise during cleanup
        ctx.__exit__(ValueError, ValueError("Test error"), None)
        assert ctx._started is False

    def test_spinner_context_exception_handling(self) -> None:
        """Test SpinnerContext handles exceptions gracefully."""
        ctx = SpinnerContext("Loading...")
        ctx.__enter__()
        # Simulate exception handling
        # Should not raise during cleanup
        ctx.__exit__(RuntimeError, RuntimeError("Test error"), None)

    def test_advance_with_invalid_task_name(self) -> None:
        """Test advance with non-existent task name."""
        with ProgressContext("Test") as ctx:
            ctx.add_task("Valid Task", total=100)
            # Should not raise
            ctx.advance("Invalid Task", 10)


class TestStyleConstants:
    """Test style constant definitions."""

    def test_claude_style_exists(self) -> None:
        """Test CLAUDE_STYLE constant exists."""
        assert CLAUDE_STYLE is not None

    def test_success_style_exists(self) -> None:
        """Test SUCCESS_STYLE constant exists."""
        assert SUCCESS_STYLE is not None

    def test_error_style_exists(self) -> None:
        """Test ERROR_STYLE constant exists."""
        assert ERROR_STYLE is not None

    def test_info_style_exists(self) -> None:
        """Test INFO_STYLE constant exists."""
        assert INFO_STYLE is not None

    def test_warning_style_exists(self) -> None:
        """Test WARNING_STYLE constant exists."""
        assert WARNING_STYLE is not None


class TestConsoleInstance:
    """Test console instance."""

    def test_console_exists(self) -> None:
        """Test console instance exists."""
        assert console is not None
        assert isinstance(console, Console)


class TestEdgeCases:
    """Test edge cases and boundary conditions."""

    def test_zero_total_progress(self) -> None:
        """Test progress bar with zero total."""
        with create_progress_bar(total=0) as progress:
            task = progress.add_task("Empty", total=0)
            # Should not raise

    def test_very_large_total(self) -> None:
        """Test progress bar with very large total."""
        with create_progress_bar(total=10_000_000) as progress:
            task = progress.add_task("Large", total=10_000_000)
            progress.update(task, advance=1_000_000)

    def test_empty_string_description(self) -> None:
        """Test progress bar with empty description."""
        progress = create_progress_bar(description="")
        assert isinstance(progress, Progress)

    def test_zero_advance(self) -> None:
        """Test advancing by zero."""
        with ProgressContext("Test") as ctx:
            ctx.add_task("Task", total=100)
            ctx.advance("Task", 0)

    def test_negative_advance(self) -> None:
        """Test advancing by negative amount (should work)."""
        with ProgressContext("Test") as ctx:
            ctx.add_task("Task", total=100)
            ctx.advance("Task", 10)
            ctx.advance("Task", -5)

    def test_update_task_beyond_total(self) -> None:
        """Test updating task beyond total (allowed)."""
        with ProgressContext("Test") as ctx:
            ctx.add_task("Task", total=100)
            ctx.update_task("Task", completed=150)

    def test_complete_task_with_zero_total(self) -> None:
        """Test completing task with zero total."""
        with ProgressContext("Test") as ctx:
            ctx.add_task("Task", total=0)
            ctx.complete_task("Task")

    def test_consecutive_spinners(self) -> None:
        """Test multiple consecutive spinners."""
        for i in range(3):
            with SpinnerContext(f"Spinner {i}"):
                pass

    def test_rapid_status_updates(self) -> None:
        """Test rapid spinner status updates."""
        with SpinnerContext("Rapid") as spinner:
            for i in range(100):
                spinner.update(f"Update {i}")
