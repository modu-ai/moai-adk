"""
Unit tests for the progress tracker module.

Tests the ProgressTracker and MultiStageProgressTracker classes
to ensure proper progress tracking and display functionality.
"""

import pytest
from unittest.mock import patch, MagicMock, call
from io import StringIO

from moai_adk.utils.progress_tracker import ProgressTracker, MultiStageProgressTracker
class TestProgressTracker:
    """Test cases for ProgressTracker class."""

    @pytest.fixture
    def tracker(self):
        """Create a ProgressTracker instance for testing."""
        return ProgressTracker(total_steps=10)

    def test_init_default_steps(self):
        """Test ProgressTracker initialization with default steps."""
        tracker = ProgressTracker()
        assert tracker.total_steps == 18
        assert tracker.current_step == 0

    def test_init_custom_steps(self):
        """Test ProgressTracker initialization with custom steps."""
        tracker = ProgressTracker(total_steps=25)
        assert tracker.total_steps == 25
        assert tracker.current_step == 0

    def test_update_progress_with_callback(self, tracker):
        """Test update_progress with custom callback."""
        callback_calls = []

        def test_callback(step: str, current: int, total: int):
            callback_calls.append((step, current, total))

        tracker.update_progress("Test step", callback=test_callback)

        assert tracker.current_step == 1
        assert len(callback_calls) == 1
        assert callback_calls[0] == ("Test step", 1, 10)

    @patch('builtins.print')
    def test_update_progress_without_callback(self, mock_print, tracker):
        """Test update_progress with default display."""
        tracker.update_progress("Test step")

        assert tracker.current_step == 1
        mock_print.assert_called_once()

    @patch('builtins.print')
    def test_update_progress_multiple_steps(self, mock_print, tracker):
        """Test multiple progress updates."""
        steps = ["Step 1", "Step 2", "Step 3"]

        for step in steps:
            tracker.update_progress(step)

        assert tracker.current_step == 3
        assert mock_print.call_count == 3

    def test_get_progress_color_yellow(self, tracker):
        """Test progress color selection for low percentage."""
        color = tracker._get_progress_color(25)
        assert color == "\033[33m"  # Fore.YELLOW

    def test_get_progress_color_cyan(self, tracker):
        """Test progress color selection for medium percentage."""
        color = tracker._get_progress_color(50)
        assert color == "\033[36m"  # Fore.CYAN

    def test_get_progress_color_green(self, tracker):
        """Test progress color selection for high percentage."""
        color = tracker._get_progress_color(90)
        assert color == "\033[32m"  # Fore.GREEN

    def test_create_progress_bar_empty(self, tracker):
        """Test progress bar creation at 0% completion."""
        bar = tracker._create_progress_bar(width=10)
        assert bar == "░░░░░░░░░░"

    def test_create_progress_bar_half(self, tracker):
        """Test progress bar creation at 50% completion."""
        tracker.current_step = 5  # 5/10 = 50%
        bar = tracker._create_progress_bar(width=10)
        assert bar == "█████░░░░░"

    def test_create_progress_bar_full(self, tracker):
        """Test progress bar creation at 100% completion."""
        tracker.current_step = 10  # 10/10 = 100%
        bar = tracker._create_progress_bar(width=10)
        assert bar == "██████████"

    def test_create_progress_bar_over_completion(self, tracker):
        """Test progress bar caps at 100% even when over-completed."""
        tracker.current_step = 15  # 15/10 = 150%, should cap at 100%
        bar = tracker._create_progress_bar(width=10)
        assert bar == "██████████"

    def test_reset_progress(self, tracker):
        """Test resetting progress tracker."""
        tracker.current_step = 5
        tracker.reset()
        assert tracker.current_step == 0

    def test_set_total_steps(self, tracker):
        """Test updating total steps count."""
        tracker.set_total_steps(20)
        assert tracker.total_steps == 20

    def test_get_progress_info_initial(self, tracker):
        """Test progress info at initial state."""
        info = tracker.get_progress_info()
        expected = {
            'current_step': 0,
            'total_steps': 10,
            'percentage': 0,
            'is_complete': False
        }
        assert info == expected

    def test_get_progress_info_partial(self, tracker):
        """Test progress info at partial completion."""
        tracker.current_step = 3
        info = tracker.get_progress_info()
        expected = {
            'current_step': 3,
            'total_steps': 10,
            'percentage': 30,
            'is_complete': False
        }
        assert info == expected

    def test_get_progress_info_complete(self, tracker):
        """Test progress info at completion."""
        tracker.current_step = 10
        info = tracker.get_progress_info()
        expected = {
            'current_step': 10,
            'total_steps': 10,
            'percentage': 100,
            'is_complete': True
        }
        assert info == expected

    def test_is_complete_false(self, tracker):
        """Test is_complete returns False when not complete."""
        tracker.current_step = 5
        assert tracker.is_complete() is False

    def test_is_complete_true(self, tracker):
        """Test is_complete returns True when complete."""
        tracker.current_step = 10
        assert tracker.is_complete() is True

    def test_is_complete_over_completion(self, tracker):
        """Test is_complete returns True when over-completed."""
        tracker.current_step = 15
        assert tracker.is_complete() is True

    def test_skip_to_step_valid(self, tracker):
        """Test skipping to valid step number."""
        tracker.skip_to_step(5)
        assert tracker.current_step == 5

    def test_skip_to_step_zero(self, tracker):
        """Test skipping to step zero."""
        tracker.current_step = 5
        tracker.skip_to_step(0)
        assert tracker.current_step == 0

    def test_skip_to_step_max(self, tracker):
        """Test skipping to maximum step."""
        tracker.skip_to_step(10)
        assert tracker.current_step == 10

    def test_skip_to_step_invalid_negative(self, tracker):
        """Test skipping to invalid negative step."""
        original_step = tracker.current_step
        tracker.skip_to_step(-1)
        assert tracker.current_step == original_step  # Should not change

    def test_skip_to_step_invalid_over_max(self, tracker):
        """Test skipping to step over maximum."""
        original_step = tracker.current_step
        tracker.skip_to_step(15)
        assert tracker.current_step == original_step  # Should not change

    def test_add_steps(self, tracker):
        """Test adding additional steps."""
        original_total = tracker.total_steps
        tracker.add_steps(5)
        assert tracker.total_steps == original_total + 5

    def test_create_sub_tracker(self, tracker):
        """Test creating sub-tracker."""
        sub_tracker = tracker.create_sub_tracker(5)

        assert isinstance(sub_tracker, ProgressTracker)
        assert sub_tracker.total_steps == 5
        assert sub_tracker.current_step == 0
        # Should be independent of parent tracker
        assert sub_tracker is not tracker

    @patch('builtins.print')
    def test_display_completion_message_default(self, mock_print, tracker):
        """Test display completion message with default text."""
        tracker.display_completion_message()

        mock_print.assert_called_once()
        args = mock_print.call_args[0][0]
        assert "✅" in args
        assert "Installation completed successfully!" in args

    @patch('builtins.print')
    def test_display_completion_message_custom(self, mock_print, tracker):
        """Test display completion message with custom text."""
        custom_message = "Custom completion message"
        tracker.display_completion_message(custom_message)

        args = mock_print.call_args[0][0]
        assert custom_message in args

    @patch('builtins.print')
    def test_display_error_message(self, mock_print, tracker):
        """Test display error message."""
        error_message = "Something went wrong"
        tracker.display_error_message(error_message)

        args = mock_print.call_args[0][0]
        assert "❌" in args
        assert error_message in args

    @patch('builtins.print')
    def test_display_warning_message(self, mock_print, tracker):
        """Test display warning message."""
        warning_message = "Warning message"
        tracker.display_warning_message(warning_message)

        args = mock_print.call_args[0][0]
        assert "⚠️" in args
        assert warning_message in args

    @patch('builtins.print')
    def test_display_default_progress_formatting(self, mock_print, tracker):
        """Test default progress display formatting."""
        tracker.update_progress("Test step")

        mock_print.assert_called_once()
        # Check that print was called with proper formatting arguments
        call_args = mock_print.call_args
        assert call_args[1]['end'] == ""  # Should not add newline
        assert call_args[1]['flush'] is True  # Should flush output

    @patch('builtins.print')
    def test_display_adds_newline_when_complete(self, mock_print, tracker):
        """Test that display adds newline when progress is complete."""
        # Complete all steps except the last one
        for i in range(9):
            tracker.update_progress(f"Step {i+1}")

        mock_print.reset_mock()

        # Complete the final step
        tracker.update_progress("Final step")

        # Should have two print calls: progress display and final newline
        assert mock_print.call_count == 2
        # First call is the progress display, second is the newline
        assert mock_print.call_args_list[1] == call()
class TestMultiStageProgressTracker:
    """Test cases for MultiStageProgressTracker class."""

    @pytest.fixture
    def multi_tracker(self):
        """Create a MultiStageProgressTracker instance for testing."""
        return MultiStageProgressTracker()

    def test_init_empty(self, multi_tracker):
        """Test MultiStageProgressTracker initialization."""
        assert multi_tracker.stages == {}
        assert multi_tracker.current_stage is None
        assert multi_tracker.stage_order == []

    def test_add_stage(self, multi_tracker):
        """Test adding a stage."""
        multi_tracker.add_stage("setup", 5)

        assert "setup" in multi_tracker.stages
        assert multi_tracker.stages["setup"]["steps"] == 5
        assert isinstance(multi_tracker.stages["setup"]["tracker"], ProgressTracker)
        assert multi_tracker.stages["setup"]["completed"] is False
        assert "setup" in multi_tracker.stage_order

    def test_add_multiple_stages(self, multi_tracker):
        """Test adding multiple stages."""
        stages = [("setup", 5), ("install", 10), ("cleanup", 3)]

        for name, steps in stages:
            multi_tracker.add_stage(name, steps)

        assert len(multi_tracker.stages) == 3
        assert multi_tracker.stage_order == ["setup", "install", "cleanup"]

    def test_add_stage_duplicate_name(self, multi_tracker):
        """Test adding stage with duplicate name."""
        multi_tracker.add_stage("setup", 5)
        multi_tracker.add_stage("setup", 10)  # Should overwrite

        assert multi_tracker.stages["setup"]["steps"] == 10
        # Should not duplicate in stage_order
        assert multi_tracker.stage_order.count("setup") == 1

    @patch('builtins.print')
    def test_start_stage_valid(self, mock_print, multi_tracker):
        """Test starting a valid stage."""
        multi_tracker.add_stage("setup", 5)

        tracker = multi_tracker.start_stage("setup")

        assert multi_tracker.current_stage == "setup"
        assert isinstance(tracker, ProgressTracker)
        assert tracker.current_step == 0  # Should be reset
        mock_print.assert_called_once()

    def test_start_stage_invalid(self, multi_tracker):
        """Test starting a non-existent stage."""
        with pytest.raises(ValueError, match="Stage 'nonexistent' not found"):
            multi_tracker.start_stage("nonexistent")

    @patch('builtins.print')
    def test_start_stage_resets_tracker(self, mock_print, multi_tracker):
        """Test that starting a stage resets its tracker."""
        multi_tracker.add_stage("setup", 5)

        # Advance the tracker manually
        multi_tracker.stages["setup"]["tracker"].current_step = 3

        # Start the stage (should reset)
        tracker = multi_tracker.start_stage("setup")

        assert tracker.current_step == 0

    @patch('builtins.print')
    def test_complete_stage_valid(self, mock_print, multi_tracker):
        """Test completing a valid stage."""
        multi_tracker.add_stage("setup", 5)

        multi_tracker.complete_stage("setup")

        assert multi_tracker.stages["setup"]["completed"] is True
        mock_print.assert_called_once()

    @patch('builtins.print')
    def test_complete_stage_invalid(self, mock_print, multi_tracker):
        """Test completing a non-existent stage."""
        # Should not raise error, just not do anything
        multi_tracker.complete_stage("nonexistent")

        # No print should occur for non-existent stage
        mock_print.assert_not_called()

    def test_get_overall_progress_empty(self, multi_tracker):
        """Test overall progress with no stages."""
        progress = multi_tracker.get_overall_progress()

        expected = {
            'completed_steps': 0,
            'total_steps': 0,
            'percentage': 0,
            'current_stage': None,
            'completed_stages': []
        }
        assert progress == expected

    def test_get_overall_progress_single_stage(self, multi_tracker):
        """Test overall progress with single stage."""
        multi_tracker.add_stage("setup", 10)
        multi_tracker.stages["setup"]["tracker"].current_step = 3
        multi_tracker.current_stage = "setup"

        progress = multi_tracker.get_overall_progress()

        expected = {
            'completed_steps': 3,
            'total_steps': 10,
            'percentage': 30,
            'current_stage': "setup",
            'completed_stages': []
        }
        assert progress == expected

    def test_get_overall_progress_multiple_stages(self, multi_tracker):
        """Test overall progress with multiple stages."""
        multi_tracker.add_stage("setup", 10)
        multi_tracker.add_stage("install", 20)
        multi_tracker.add_stage("cleanup", 5)

        # Advance different stages
        multi_tracker.stages["setup"]["tracker"].current_step = 10
        multi_tracker.stages["install"]["tracker"].current_step = 5
        multi_tracker.stages["cleanup"]["tracker"].current_step = 0

        # Mark setup as completed
        multi_tracker.stages["setup"]["completed"] = True
        multi_tracker.current_stage = "install"

        progress = multi_tracker.get_overall_progress()

        expected = {
            'completed_steps': 15,  # 10 + 5 + 0
            'total_steps': 35,      # 10 + 20 + 5
            'percentage': 42,       # 15/35 * 100 = 42.857... -> 42
            'current_stage': "install",
            'completed_stages': ["setup"]
        }
        assert progress == expected

    def test_get_overall_progress_all_completed(self, multi_tracker):
        """Test overall progress with all stages completed."""
        stages = [("setup", 5), ("install", 10), ("cleanup", 3)]

        for name, steps in stages:
            multi_tracker.add_stage(name, steps)
            multi_tracker.stages[name]["tracker"].current_step = steps
            multi_tracker.stages[name]["completed"] = True

        progress = multi_tracker.get_overall_progress()

        assert progress['completed_steps'] == 18
        assert progress['total_steps'] == 18
        assert progress['percentage'] == 100
        assert len(progress['completed_stages']) == 3

    def test_stage_order_preserved(self, multi_tracker):
        """Test that stage order is preserved when adding stages."""
        stage_names = ["first", "second", "third", "fourth"]

        for name in stage_names:
            multi_tracker.add_stage(name, 5)

        assert multi_tracker.stage_order == stage_names

    def test_integration_workflow(self, multi_tracker):
        """Test complete workflow with multiple stages."""
        # Setup stages
        multi_tracker.add_stage("setup", 3)
        multi_tracker.add_stage("install", 5)
        multi_tracker.add_stage("cleanup", 2)

        with patch('builtins.print'):
            # Start and complete setup stage
            setup_tracker = multi_tracker.start_stage("setup")
            for i in range(3):
                setup_tracker.update_progress(f"Setup step {i+1}")
            multi_tracker.complete_stage("setup")

            # Start install stage
            install_tracker = multi_tracker.start_stage("install")
            for i in range(2):  # Partial completion
                install_tracker.update_progress(f"Install step {i+1}")

        # Check overall progress
        progress = multi_tracker.get_overall_progress()

        assert progress['completed_steps'] == 5  # 3 + 2
        assert progress['total_steps'] == 10     # 3 + 5 + 2
        assert progress['percentage'] == 50      # 5/10 * 100
        assert progress['current_stage'] == "install"
        assert progress['completed_stages'] == ["setup"]
        assert len(multi_tracker.stages) == 3