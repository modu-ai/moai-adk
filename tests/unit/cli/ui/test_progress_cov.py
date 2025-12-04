"""Comprehensive tests for progress.py module.

Focus on uncovered code paths with actual execution using mocked dependencies.
"""

import pytest
from pathlib import Path
from unittest.mock import MagicMock, Mock, patch, call
from typing import Any

from moai_adk.cli.ui.progress import (
    MoAISpinnerColumn,
    create_progress_bar,
    create_spinner,
    ProgressContext,
    SpinnerContext,
    progress_steps,
    print_step,
    console,
)
from rich.text import Text
from rich.style import Style
from rich.progress import Progress, TaskID


class TestMoAISpinnerColumn:
    """Test MoAISpinnerColumn class."""

    def test_init_with_default_spinner(self):
        """Test initialization with default spinner name."""
        column = MoAISpinnerColumn()
        assert column.spinner is not None
        assert column.spinner.name == "dots"

    def test_init_with_custom_spinner(self):
        """Test initialization with custom spinner name."""
        column = MoAISpinnerColumn(spinner_name="arc")
        assert column.spinner is not None
        assert column.spinner.name == "arc"

    def test_render_returns_text(self):
        """Test render method returns Text object."""
        column = MoAISpinnerColumn()
        task = MagicMock()
        task.get_time.return_value = 1.5

        result = column.render(task)
        assert isinstance(result, Text)

    def test_render_handles_non_text_result(self):
        """Test render handles spinner returning non-Text object."""
        column = MoAISpinnerColumn()
        task = MagicMock()
        task.get_time.return_value = 0.0

        # Mock spinner.render to return a string
        original_render = column.spinner.render

        with patch.object(column.spinner, "render", return_value="spinner_frame"):
            result = column.render(task)
            assert isinstance(result, Text)
            assert "spinner_frame" in str(result)

    def test_render_calls_task_get_time(self):
        """Test that render calls task.get_time()."""
        column = MoAISpinnerColumn()
        task = MagicMock()
        task.get_time.return_value = 2.0

        column.render(task)
        task.get_time.assert_called_once()


class TestCreateProgressBar:
    """Test create_progress_bar function."""

    def test_default_parameters(self):
        """Test progress bar with default parameters."""
        progress = create_progress_bar()
        assert isinstance(progress, Progress)
        # Progress object is created successfully
        assert progress is not None

    def test_with_custom_description(self):
        """Test progress bar with custom description."""
        progress = create_progress_bar(description="Downloading files")
        assert isinstance(progress, Progress)

    def test_with_total_steps(self):
        """Test progress bar with total steps."""
        progress = create_progress_bar(description="Processing", total=100)
        assert isinstance(progress, Progress)

    def test_without_total_steps(self):
        """Test progress bar without total (indeterminate)."""
        progress = create_progress_bar(description="Processing", total=None)
        assert isinstance(progress, Progress)

    def test_transient_mode(self):
        """Test progress bar with transient mode enabled."""
        progress = create_progress_bar(transient=True)
        assert isinstance(progress, Progress)

    def test_auto_refresh_disabled(self):
        """Test progress bar with auto-refresh disabled."""
        progress = create_progress_bar(auto_refresh=False)
        assert isinstance(progress, Progress)

    def test_columns_configuration_with_total(self):
        """Test that columns are properly configured with total."""
        progress = create_progress_bar(total=100)
        # Progress bar should have time remaining columns when total is specified
        assert isinstance(progress, Progress)

    def test_columns_configuration_without_total(self):
        """Test columns configuration without total."""
        progress = create_progress_bar(total=None)
        # Progress bar should not have time remaining when total is None
        assert isinstance(progress, Progress)


class TestCreateSpinner:
    """Test create_spinner function."""

    def test_default_parameters(self):
        """Test spinner with default parameters."""
        with patch.object(console, "status") as mock_status:
            mock_status.return_value = MagicMock()
            spinner = create_spinner()
            mock_status.assert_called_once()
            args, kwargs = mock_status.call_args
            assert args[0] == "Processing..."

    def test_custom_message(self):
        """Test spinner with custom message."""
        with patch.object(console, "status") as mock_status:
            mock_status.return_value = MagicMock()
            create_spinner(message="Checking updates...")
            args, kwargs = mock_status.call_args
            assert args[0] == "Checking updates..."

    def test_custom_spinner_name(self):
        """Test spinner with custom spinner name."""
        with patch.object(console, "status") as mock_status:
            mock_status.return_value = MagicMock()
            create_spinner(spinner_name="arc")
            _, kwargs = mock_status.call_args
            assert kwargs["spinner"] == "arc"

    def test_returns_status_object(self):
        """Test that create_spinner returns Status object."""
        with patch.object(console, "status") as mock_status:
            status_mock = MagicMock()
            mock_status.return_value = status_mock
            result = create_spinner()
            assert result == status_mock


class TestProgressContext:
    """Test ProgressContext class."""

    def test_initialization(self):
        """Test ProgressContext initialization."""
        ctx = ProgressContext("Processing files")
        assert ctx.title == "Processing files"
        assert ctx.transient is False
        assert ctx.progress is None
        assert ctx._started is False
        assert ctx.tasks == {}

    def test_context_manager_enter(self):
        """Test entering context manager."""
        with patch("moai_adk.cli.ui.progress.create_progress_bar") as mock_create:
            mock_progress = MagicMock()
            mock_create.return_value = mock_progress

            ctx = ProgressContext("Test task")
            result = ctx.__enter__()

            assert result is ctx
            assert ctx._started is True
            assert ctx.progress is mock_progress
            mock_progress.start.assert_called_once()

    def test_context_manager_exit(self):
        """Test exiting context manager."""
        with patch("moai_adk.cli.ui.progress.create_progress_bar") as mock_create:
            mock_progress = MagicMock()
            mock_create.return_value = mock_progress

            with ProgressContext("Test task") as ctx:
                pass

            mock_progress.stop.assert_called_once()
            assert ctx._started is False

    def test_context_manager_with_transient(self):
        """Test context manager with transient mode."""
        with patch("moai_adk.cli.ui.progress.create_progress_bar") as mock_create:
            mock_progress = MagicMock()
            mock_create.return_value = mock_progress

            with ProgressContext("Test task", transient=True) as ctx:
                args, kwargs = mock_create.call_args
                assert kwargs["transient"] is True

    def test_add_task_success(self):
        """Test adding a task."""
        with patch("moai_adk.cli.ui.progress.create_progress_bar") as mock_create:
            mock_progress = MagicMock()
            mock_progress.add_task.return_value = TaskID(1)
            mock_create.return_value = mock_progress

            with ProgressContext("Test") as ctx:
                task_id = ctx.add_task("Download", total=100)

                assert "Download" in ctx.tasks
                assert ctx.tasks["Download"] == TaskID(1)
                mock_progress.add_task.assert_called_once_with("Download", total=100, visible=True)

    def test_add_task_without_progress_raises(self):
        """Test add_task raises when context not started."""
        ctx = ProgressContext("Test")
        with pytest.raises(RuntimeError, match="ProgressContext not started"):
            ctx.add_task("Download")

    def test_add_task_defaults_total_to_100(self):
        """Test add_task defaults total to 100."""
        with patch("moai_adk.cli.ui.progress.create_progress_bar") as mock_create:
            mock_progress = MagicMock()
            mock_progress.add_task.return_value = TaskID(1)
            mock_create.return_value = mock_progress

            with ProgressContext("Test") as ctx:
                ctx.add_task("Task without total")
                _, kwargs = mock_progress.add_task.call_args
                assert kwargs["total"] == 100

    def test_advance_task(self):
        """Test advancing a task."""
        with patch("moai_adk.cli.ui.progress.create_progress_bar") as mock_create:
            mock_progress = MagicMock()
            mock_progress.add_task.return_value = TaskID(1)
            mock_create.return_value = mock_progress

            with ProgressContext("Test") as ctx:
                ctx.add_task("Download", total=100)
                ctx.advance("Download", advance=10)

                mock_progress.advance.assert_called_once_with(TaskID(1), 10)

    def test_advance_task_default_advance(self):
        """Test advancing task with default advance value."""
        with patch("moai_adk.cli.ui.progress.create_progress_bar") as mock_create:
            mock_progress = MagicMock()
            mock_progress.add_task.return_value = TaskID(1)
            mock_create.return_value = mock_progress

            with ProgressContext("Test") as ctx:
                ctx.add_task("Download")
                ctx.advance("Download")

                mock_progress.advance.assert_called_once_with(TaskID(1), 1)

    def test_advance_unknown_task(self):
        """Test advancing a task that doesn't exist."""
        with patch("moai_adk.cli.ui.progress.create_progress_bar") as mock_create:
            mock_progress = MagicMock()
            mock_create.return_value = mock_progress

            with ProgressContext("Test") as ctx:
                ctx.advance("Nonexistent", advance=5)
                mock_progress.advance.assert_not_called()

    def test_advance_without_progress(self):
        """Test advance when no progress object."""
        ctx = ProgressContext("Test")
        ctx.advance("Task")
        # Should not raise

    def test_update_task_completed(self):
        """Test updating task completed value."""
        with patch("moai_adk.cli.ui.progress.create_progress_bar") as mock_create:
            mock_progress = MagicMock()
            mock_progress.add_task.return_value = TaskID(1)
            mock_create.return_value = mock_progress

            with ProgressContext("Test") as ctx:
                ctx.add_task("Download", total=100)
                ctx.update_task("Download", completed=50)

                mock_progress.update.assert_called_once_with(TaskID(1), completed=50)

    def test_update_task_total(self):
        """Test updating task total value."""
        with patch("moai_adk.cli.ui.progress.create_progress_bar") as mock_create:
            mock_progress = MagicMock()
            mock_progress.add_task.return_value = TaskID(1)
            mock_create.return_value = mock_progress

            with ProgressContext("Test") as ctx:
                ctx.add_task("Download", total=100)
                ctx.update_task("Download", total=200)

                mock_progress.update.assert_called_once_with(TaskID(1), total=200)

    def test_update_task_description(self):
        """Test updating task description."""
        with patch("moai_adk.cli.ui.progress.create_progress_bar") as mock_create:
            mock_progress = MagicMock()
            mock_progress.add_task.return_value = TaskID(1)
            mock_create.return_value = mock_progress

            with ProgressContext("Test") as ctx:
                ctx.add_task("Download", total=100)
                ctx.update_task("Download", new_description="Installing")

                # Should have renamed task mapping
                assert "Installing" in ctx.tasks
                assert "Download" not in ctx.tasks

    def test_update_task_multiple_values(self):
        """Test updating multiple task values at once."""
        with patch("moai_adk.cli.ui.progress.create_progress_bar") as mock_create:
            mock_progress = MagicMock()
            mock_progress.add_task.return_value = TaskID(1)
            mock_create.return_value = mock_progress

            with ProgressContext("Test") as ctx:
                ctx.add_task("Download", total=100)
                ctx.update_task("Download", completed=75, total=100)

                _, kwargs = mock_progress.update.call_args
                assert kwargs["completed"] == 75
                assert kwargs["total"] == 100

    def test_complete_task(self):
        """Test completing a task."""
        with patch("moai_adk.cli.ui.progress.create_progress_bar") as mock_create:
            mock_progress = MagicMock()
            mock_task = MagicMock()
            mock_task.total = 100
            mock_progress.add_task.return_value = TaskID(1)
            mock_progress.tasks = {TaskID(1): mock_task}
            mock_create.return_value = mock_progress

            with ProgressContext("Test") as ctx:
                ctx.add_task("Download", total=100)
                ctx.complete_task("Download")

                mock_progress.update.assert_called_with(TaskID(1), completed=100)

    def test_complete_unknown_task(self):
        """Test completing a task that doesn't exist."""
        with patch("moai_adk.cli.ui.progress.create_progress_bar") as mock_create:
            mock_progress = MagicMock()
            mock_create.return_value = mock_progress

            with ProgressContext("Test") as ctx:
                ctx.complete_task("Nonexistent")
                mock_progress.update.assert_not_called()

    def test_complete_task_without_progress(self):
        """Test complete_task when no progress object."""
        ctx = ProgressContext("Test")
        ctx.complete_task("Task")
        # Should not raise

    def test_update_task_without_progress(self):
        """Test update_task when no progress object."""
        ctx = ProgressContext("Test")
        ctx.update_task("Task", completed=50)
        # Should not raise


class TestSpinnerContext:
    """Test SpinnerContext class."""

    def test_initialization(self):
        """Test SpinnerContext initialization."""
        ctx = SpinnerContext("Processing...")
        assert ctx.initial_message == "Processing..."
        assert ctx.spinner_name == "dots"
        assert ctx.status is None
        assert ctx._final_message is None
        assert ctx._final_style is None

    def test_initialization_custom_spinner(self):
        """Test initialization with custom spinner name."""
        ctx = SpinnerContext("Processing", spinner_name="arc")
        assert ctx.spinner_name == "arc"

    def test_context_manager_enter(self):
        """Test entering context manager."""
        with patch.object(console, "status") as mock_status:
            mock_status_obj = MagicMock()
            mock_status.return_value = mock_status_obj

            ctx = SpinnerContext("Starting")
            result = ctx.__enter__()

            assert result is ctx
            assert ctx.status is mock_status_obj
            mock_status_obj.start.assert_called_once()

    def test_context_manager_exit_without_message(self):
        """Test exiting context manager without final message."""
        with patch.object(console, "status") as mock_status:
            mock_status_obj = MagicMock()
            mock_status.return_value = mock_status_obj

            with SpinnerContext("Processing"):
                pass

            mock_status_obj.stop.assert_called_once()

    def test_context_manager_exit_with_success_message(self):
        """Test exiting with success message."""
        with patch.object(console, "status") as mock_status:
            with patch.object(console, "print") as mock_print:
                mock_status_obj = MagicMock()
                mock_status.return_value = mock_status_obj

                with SpinnerContext("Processing") as ctx:
                    ctx.success("Operation completed")

                mock_status_obj.stop.assert_called_once()
                mock_print.assert_called_once()
                args, kwargs = mock_print.call_args
                assert "✓" in args[0]
                assert "Operation completed" in args[0]

    def test_context_manager_exit_with_error_message(self):
        """Test exiting with error message."""
        with patch.object(console, "status") as mock_status:
            with patch.object(console, "print") as mock_print:
                mock_status_obj = MagicMock()
                mock_status.return_value = mock_status_obj

                with SpinnerContext("Processing") as ctx:
                    ctx.error("Operation failed")

                mock_print.assert_called_once()
                args, _ = mock_print.call_args
                assert "✗" in args[0]
                assert "Operation failed" in args[0]

    def test_update_message(self):
        """Test updating spinner message."""
        with patch.object(console, "status") as mock_status:
            mock_status_obj = MagicMock()
            mock_status.return_value = mock_status_obj

            with SpinnerContext("Processing") as ctx:
                ctx.update("Step 2: Installing...")
                mock_status_obj.update.assert_called_with("Step 2: Installing...")

    def test_update_without_status(self):
        """Test update when status not initialized."""
        ctx = SpinnerContext("Processing")
        ctx.update("Message")  # Should not raise

    def test_success_sets_message_and_style(self):
        """Test success sets final message and style."""
        ctx = SpinnerContext("Processing")
        ctx.success("Task complete")

        assert ctx._final_message == "✓ Task complete"
        assert ctx._final_style is not None

    def test_error_sets_message_and_style(self):
        """Test error sets final message and style."""
        ctx = SpinnerContext("Processing")
        ctx.error("Task failed")

        assert ctx._final_message == "✗ Task failed"
        assert ctx._final_style is not None

    def test_warning_sets_message_and_style(self):
        """Test warning sets final message and style."""
        ctx = SpinnerContext("Processing")
        ctx.warning("Check this")

        assert ctx._final_message == "⚠ Check this"
        assert ctx._final_style is not None

    def test_info_sets_message_and_style(self):
        """Test info sets final message and style."""
        ctx = SpinnerContext("Processing")
        ctx.info("For your information")

        assert ctx._final_message == "ℹ For your information"
        assert ctx._final_style is not None

    def test_multiple_status_calls(self):
        """Test setting multiple status messages overwrites."""
        with patch.object(console, "status") as mock_status:
            with patch.object(console, "print") as mock_print:
                mock_status_obj = MagicMock()
                mock_status.return_value = mock_status_obj

                with SpinnerContext("Processing") as ctx:
                    ctx.success("First message")
                    ctx.error("Second message")

                # Last message should be the one printed
                mock_print.assert_called_once()
                args, _ = mock_print.call_args
                assert "✗" in args[0]
                assert "Second message" in args[0]


class TestProgressSteps:
    """Test progress_steps context manager."""

    def test_basic_iteration(self):
        """Test iterating through progress steps."""
        steps = ["Download", "Extract", "Install"]

        with patch("moai_adk.cli.ui.progress.ProgressContext") as mock_ctx_class:
            mock_ctx = MagicMock()
            mock_ctx_class.return_value = mock_ctx
            mock_ctx.__enter__.return_value = mock_ctx

            with progress_steps(steps, "Setup") as step_iter:
                result = list(step_iter)

            assert result == steps

    def test_step_descriptions_updated(self):
        """Test that step descriptions are updated during iteration."""
        steps = ["Step1", "Step2"]

        with patch("moai_adk.cli.ui.progress.ProgressContext") as mock_ctx_class:
            mock_ctx = MagicMock()
            mock_ctx_class.return_value = mock_ctx
            mock_ctx.__enter__.return_value = mock_ctx

            with progress_steps(steps, "Process") as step_iter:
                list(step_iter)

            # Verify update_task was called for each step
            update_calls = [c for c in mock_ctx.method_calls if "update_task" in str(c)]
            assert len(update_calls) >= len(steps)

    def test_progress_advanced(self):
        """Test that progress advances after each step."""
        steps = ["A", "B", "C"]

        with patch("moai_adk.cli.ui.progress.ProgressContext") as mock_ctx_class:
            mock_ctx = MagicMock()
            mock_ctx_class.return_value = mock_ctx
            mock_ctx.__enter__.return_value = mock_ctx

            with progress_steps(steps, "Test") as step_iter:
                for _ in step_iter:
                    pass

            # Verify advance was called
            advance_calls = [c for c in mock_ctx.method_calls if "advance" in str(c)]
            assert len(advance_calls) >= len(steps)

    def test_task_created(self):
        """Test that a task is created for steps."""
        steps = ["Step1", "Step2", "Step3"]

        with patch("moai_adk.cli.ui.progress.ProgressContext") as mock_ctx_class:
            mock_ctx = MagicMock()
            mock_ctx_class.return_value = mock_ctx
            mock_ctx.__enter__.return_value = mock_ctx

            with progress_steps(steps, "Installation") as step_iter:
                list(step_iter)

            # Verify add_task was called with correct total
            mock_ctx.add_task.assert_called_once()
            args, kwargs = mock_ctx.add_task.call_args
            assert kwargs["total"] == len(steps)

    def test_custom_title(self):
        """Test with custom title."""
        with patch("moai_adk.cli.ui.progress.ProgressContext") as mock_ctx_class:
            mock_ctx = MagicMock()
            mock_ctx_class.return_value = mock_ctx
            mock_ctx.__enter__.return_value = mock_ctx

            with progress_steps(["S1"], "Custom Title") as step_iter:
                list(step_iter)

            # Verify ProgressContext was created with custom title
            mock_ctx_class.assert_called_once_with("Custom Title")


class TestPrintStep:
    """Test print_step function."""

    def test_running_status(self):
        """Test printing step with running status."""
        with patch.object(console, "print") as mock_print:
            print_step(1, 5, "Processing files", status="running")
            mock_print.assert_called_once()
            args, _ = mock_print.call_args
            assert "[1/5]" in args[0]
            assert "Processing files" in args[0]
            assert "→" in args[0]

    def test_complete_status(self):
        """Test printing step with complete status."""
        with patch.object(console, "print") as mock_print:
            print_step(3, 5, "Download", status="complete")
            mock_print.assert_called_once()
            args, _ = mock_print.call_args
            assert "✓" in args[0]

    def test_error_status(self):
        """Test printing step with error status."""
        with patch.object(console, "print") as mock_print:
            print_step(2, 5, "Installation failed", status="error")
            mock_print.assert_called_once()
            args, _ = mock_print.call_args
            assert "✗" in args[0]

    def test_skipped_status(self):
        """Test printing step with skipped status."""
        with patch.object(console, "print") as mock_print:
            print_step(4, 5, "Optional feature", status="skipped")
            mock_print.assert_called_once()
            args, _ = mock_print.call_args
            assert "○" in args[0]

    def test_unknown_status_defaults(self):
        """Test unknown status uses default symbol."""
        with patch.object(console, "print") as mock_print:
            print_step(1, 5, "Step", status="unknown")
            mock_print.assert_called_once()
            args, _ = mock_print.call_args
            assert "•" in args[0]

    def test_step_number_format(self):
        """Test step number formatting."""
        with patch.object(console, "print") as mock_print:
            print_step(7, 12, "Final step")
            args, _ = mock_print.call_args
            assert "[7/12]" in args[0]

    def test_message_included(self):
        """Test that message is included in output."""
        with patch.object(console, "print") as mock_print:
            print_step(1, 1, "Only step")
            args, _ = mock_print.call_args
            assert "Only step" in args[0]

    def test_style_applied(self):
        """Test that style is applied to output."""
        with patch.object(console, "print") as mock_print:
            print_step(1, 3, "Processing")
            _, kwargs = mock_print.call_args
            assert "style" in kwargs
            assert kwargs["style"] is not None


# Integration Tests


class TestProgressIntegration:
    """Integration tests combining multiple components."""

    def test_progress_context_full_cycle(self):
        """Test full progress context cycle."""
        with patch("moai_adk.cli.ui.progress.create_progress_bar") as mock_create:
            mock_progress = MagicMock()
            mock_progress.add_task.return_value = TaskID(1)
            mock_task = MagicMock()
            mock_task.total = 100
            mock_progress.tasks = {TaskID(1): mock_task}
            mock_create.return_value = mock_progress

            with ProgressContext("Multi-task") as ctx:
                task1 = ctx.add_task("Task 1", total=100)
                ctx.advance("Task 1", 50)
                ctx.update_task("Task 1", completed=50)
                ctx.complete_task("Task 1")

            assert mock_progress.add_task.called
            assert mock_progress.advance.called
            assert mock_progress.update.called

    def test_spinner_context_with_multiple_updates(self):
        """Test spinner with multiple status updates."""
        with patch.object(console, "status") as mock_status:
            with patch.object(console, "print") as mock_print:
                mock_status_obj = MagicMock()
                mock_status.return_value = mock_status_obj

                with SpinnerContext("Start") as ctx:
                    ctx.update("Progress 25%")
                    ctx.update("Progress 50%")
                    ctx.update("Progress 75%")
                    ctx.success("Complete")

                assert mock_status_obj.update.call_count == 3
                mock_print.assert_called_once()

    def test_progress_bar_with_different_configurations(self):
        """Test progress bar with various configurations."""
        configs = [
            {"description": "Test 1", "total": 100},
            {"description": "Test 2", "total": None},
            {"description": "Test 3", "transient": True},
            {"description": "Test 4", "auto_refresh": False},
        ]

        for config in configs:
            progress = create_progress_bar(**config)
            assert isinstance(progress, Progress)

    def test_progress_steps_with_multiple_iterations(self):
        """Test progress_steps with actual iteration."""
        steps = ["Setup", "Build", "Test", "Deploy"]

        with patch("moai_adk.cli.ui.progress.ProgressContext") as mock_ctx_class:
            mock_ctx = MagicMock()
            mock_ctx_class.return_value = mock_ctx
            mock_ctx.__enter__.return_value = mock_ctx

            collected_steps = []
            with progress_steps(steps, "Pipeline") as step_iter:
                for step in step_iter:
                    collected_steps.append(step)

            assert collected_steps == steps
            assert mock_ctx.add_task.called
            assert mock_ctx.advance.called
