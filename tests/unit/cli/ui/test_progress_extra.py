"""Enhanced unit tests for CLI progress indicators module.

This module tests:
- MoAISpinnerColumn for custom spinner rendering
- create_progress_bar with various configurations
- create_spinner for indeterminate operations
- ProgressContext for multi-task tracking
- SpinnerContext for status updates
- progress_steps helper and print_step utility
"""

from pathlib import Path
from unittest import mock

import pytest
from rich.progress import Progress
from rich.status import Status

from moai_adk.cli.ui.progress import (
    MoAISpinnerColumn,
    ProgressContext,
    SpinnerContext,
    create_progress_bar,
    create_spinner,
    print_step,
    progress_steps,
)


class TestMoAISpinnerColumn:
    """Test MoAISpinnerColumn custom column."""

    def test_spinner_column_initialization(self):
        """Test spinner column initialization."""
        column = MoAISpinnerColumn(spinner_name="dots")
        assert column.spinner is not None

    def test_spinner_column_initialization_custom(self):
        """Test spinner column initialization with custom spinner."""
        column = MoAISpinnerColumn(spinner_name="line")
        assert column.spinner is not None

    def test_spinner_column_render(self):
        """Test spinner column rendering."""
        column = MoAISpinnerColumn()
        mock_task = mock.MagicMock()
        mock_task.get_time.return_value = 0.5
        result = column.render(mock_task)
        assert result is not None


class TestCreateProgressBar:
    """Test create_progress_bar function."""

    def test_create_progress_bar_default(self):
        """Test creating progress bar with defaults."""
        progress = create_progress_bar()
        assert isinstance(progress, Progress)

    def test_create_progress_bar_with_description(self):
        """Test creating progress bar with description."""
        progress = create_progress_bar(description="Processing files")
        assert isinstance(progress, Progress)

    def test_create_progress_bar_with_total(self):
        """Test creating progress bar with total steps."""
        progress = create_progress_bar(description="Test", total=100)
        assert isinstance(progress, Progress)

    def test_create_progress_bar_transient(self):
        """Test creating transient progress bar."""
        progress = create_progress_bar(transient=True)
        assert isinstance(progress, Progress)

    def test_create_progress_bar_no_refresh(self):
        """Test creating progress bar without auto-refresh."""
        progress = create_progress_bar(auto_refresh=False)
        assert isinstance(progress, Progress)


class TestCreateSpinner:
    """Test create_spinner function."""

    def test_create_spinner_default(self):
        """Test creating spinner with default message."""
        spinner = create_spinner()
        assert isinstance(spinner, Status)

    def test_create_spinner_custom_message(self):
        """Test creating spinner with custom message."""
        spinner = create_spinner(message="Loading data...")
        assert isinstance(spinner, Status)

    def test_create_spinner_custom_animation(self):
        """Test creating spinner with custom animation."""
        spinner = create_spinner(spinner_name="line")
        assert isinstance(spinner, Status)


class TestProgressContext:
    """Test ProgressContext context manager."""

    def test_progress_context_initialization(self):
        """Test ProgressContext initialization."""
        ctx = ProgressContext("Test Operation")
        assert ctx.title == "Test Operation"
        assert len(ctx.tasks) == 0
        assert ctx.progress is None

    def test_progress_context_as_context_manager(self):
        """Test using ProgressContext as context manager."""
        with ProgressContext("Test") as ctx:
            assert ctx.progress is not None
            assert ctx._started is True

    def test_progress_context_add_task(self):
        """Test adding task to ProgressContext."""
        with ProgressContext("Test") as ctx:
            task_id = ctx.add_task("Process files", total=100)
            assert task_id is not None
            assert "Process files" in ctx.tasks

    def test_progress_context_advance(self):
        """Test advancing task progress."""
        with ProgressContext("Test") as ctx:
            ctx.add_task("Process", total=100)
            ctx.advance("Process", advance=10)
            # Should complete without error

    def test_progress_context_update_task(self):
        """Test updating task."""
        with ProgressContext("Test") as ctx:
            ctx.add_task("Process", total=100)
            ctx.update_task("Process", completed=50)
            # Should complete without error

    def test_progress_context_update_task_description(self):
        """Test updating task description."""
        with ProgressContext("Test") as ctx:
            ctx.add_task("Process", total=100)
            ctx.update_task("Process", new_description="New description")
            assert "New description" in ctx.tasks

    def test_progress_context_complete_task(self):
        """Test completing task."""
        with ProgressContext("Test") as ctx:
            ctx.add_task("Process", total=100)
            ctx.complete_task("Process")
            # Should complete without error

    def test_progress_context_advance_nonexistent_task(self):
        """Test advancing nonexistent task."""
        with ProgressContext("Test") as ctx:
            ctx.advance("Nonexistent")
            # Should handle gracefully

    def test_progress_context_transient(self):
        """Test transient ProgressContext."""
        with ProgressContext("Test", transient=True) as ctx:
            ctx.add_task("Process")
            assert ctx.transient is True


class TestSpinnerContext:
    """Test SpinnerContext context manager."""

    def test_spinner_context_initialization(self):
        """Test SpinnerContext initialization."""
        ctx = SpinnerContext("Loading...")
        assert ctx.initial_message == "Loading..."
        assert ctx.spinner_name == "dots"

    def test_spinner_context_as_context_manager(self):
        """Test using SpinnerContext as context manager."""
        with SpinnerContext("Loading...") as ctx:
            assert ctx.status is not None

    def test_spinner_context_update_message(self):
        """Test updating spinner message."""
        with SpinnerContext("Loading...") as ctx:
            ctx.update("Processing...")
            # Should complete without error

    def test_spinner_context_success(self):
        """Test setting success message."""
        with SpinnerContext("Loading...") as ctx:
            ctx.success("Completed successfully")
            assert ctx._final_message is not None

    def test_spinner_context_error(self):
        """Test setting error message."""
        with SpinnerContext("Loading...") as ctx:
            ctx.error("An error occurred")
            assert ctx._final_message is not None

    def test_spinner_context_warning(self):
        """Test setting warning message."""
        with SpinnerContext("Loading...") as ctx:
            ctx.warning("Warning message")
            assert ctx._final_message is not None

    def test_spinner_context_info(self):
        """Test setting info message."""
        with SpinnerContext("Loading...") as ctx:
            ctx.info("Info message")
            assert ctx._final_message is not None

    def test_spinner_context_custom_animation(self):
        """Test spinner with custom animation."""
        with SpinnerContext("Loading...", spinner_name="line") as ctx:
            assert ctx.spinner_name == "line"


class TestProgressSteps:
    """Test progress_steps generator function."""

    def test_progress_steps_basic(self):
        """Test basic progress_steps functionality."""
        steps = ["Step 1", "Step 2", "Step 3"]
        with progress_steps(steps) as step_iter:
            step_list = list(step_iter)
            assert len(step_list) == 3
            assert "Step 1" in step_list

    def test_progress_steps_custom_title(self):
        """Test progress_steps with custom title."""
        steps = ["Download", "Extract", "Install"]
        with progress_steps(steps, title="Installation") as step_iter:
            step_list = list(step_iter)
            assert len(step_list) == 3

    def test_progress_steps_single_step(self):
        """Test progress_steps with single step."""
        with progress_steps(["Only"]) as step_iter:
            step_list = list(step_iter)
            assert len(step_list) == 1

    def test_progress_steps_empty(self):
        """Test progress_steps with empty list."""
        with progress_steps([]) as step_iter:
            step_list = list(step_iter)
            assert len(step_list) == 0


class TestPrintStep:
    """Test print_step utility function."""

    @mock.patch("moai_adk.cli.ui.progress.console")
    def test_print_step_running(self, mock_console):
        """Test printing running step."""
        print_step(1, 5, "Processing files", status="running")
        mock_console.print.assert_called()

    @mock.patch("moai_adk.cli.ui.progress.console")
    def test_print_step_complete(self, mock_console):
        """Test printing completed step."""
        print_step(2, 5, "Files processed", status="complete")
        mock_console.print.assert_called()

    @mock.patch("moai_adk.cli.ui.progress.console")
    def test_print_step_error(self, mock_console):
        """Test printing error step."""
        print_step(3, 5, "Error occurred", status="error")
        mock_console.print.assert_called()

    @mock.patch("moai_adk.cli.ui.progress.console")
    def test_print_step_skipped(self, mock_console):
        """Test printing skipped step."""
        print_step(4, 5, "Step skipped", status="skipped")
        mock_console.print.assert_called()

    @mock.patch("moai_adk.cli.ui.progress.console")
    def test_print_step_invalid_status(self, mock_console):
        """Test printing step with invalid status."""
        print_step(5, 5, "Unknown status", status="unknown")
        mock_console.print.assert_called()


class TestProgressContextAdvanced:
    """Test advanced ProgressContext features."""

    def test_progress_context_multiple_tasks(self):
        """Test managing multiple tasks."""
        with ProgressContext("Operations") as ctx:
            task1 = ctx.add_task("Task 1", total=50)
            task2 = ctx.add_task("Task 2", total=100)

            assert len(ctx.tasks) == 2

            ctx.advance("Task 1", advance=5)
            ctx.advance("Task 2", advance=10)

    def test_progress_context_update_total(self):
        """Test updating task total."""
        with ProgressContext("Test") as ctx:
            ctx.add_task("Process", total=50)
            ctx.update_task("Process", total=100)
            # Should handle gracefully

    def test_progress_context_visibility(self):
        """Test task visibility control."""
        with ProgressContext("Test") as ctx:
            ctx.add_task("Visible", visible=True)
            ctx.add_task("Hidden", visible=False)
            assert len(ctx.tasks) == 2


class TestSpinnerContextEdgeCases:
    """Test SpinnerContext edge cases."""

    def test_spinner_context_multiple_messages(self):
        """Test setting multiple status messages."""
        with SpinnerContext("Start") as ctx:
            ctx.update("Step 1")
            ctx.update("Step 2")
            ctx.success("Completed")
            # Should handle multiple updates

    def test_spinner_context_final_override(self):
        """Test overriding final message."""
        with SpinnerContext("Load") as ctx:
            ctx.success("Success!")
            ctx.error("Actually failed")
            # Last message should be error

    def test_spinner_context_no_final_message(self):
        """Test spinner without final message."""
        with SpinnerContext("Processing") as ctx:
            ctx.update("Still processing")
            # Should exit without final message


class TestProgressBarIntegration:
    """Test ProgressContext and progress_steps integration."""

    def test_progress_steps_with_operations(self):
        """Test progress_steps with actual operations."""
        operations = ["Download", "Extract", "Verify", "Install"]
        with progress_steps(operations, title="Install Process") as step_iter:
            completed = 0
            for step in step_iter:
                completed += 1
            assert completed == len(operations)

    def test_nested_progress_operations(self):
        """Test nested progress operations."""
        with ProgressContext("Outer") as outer:
            outer.add_task("Outer task", total=10)

            with ProgressContext("Inner") as inner:
                inner.add_task("Inner task", total=5)
                inner.advance("Inner task", advance=5)

            outer.advance("Outer task", advance=5)


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
