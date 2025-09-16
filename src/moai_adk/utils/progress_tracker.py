"""
Progress tracking and display utilities for MoAI-ADK installation.

Provides clean, professional progress display with color coding and
customizable progress callbacks for different UI contexts.
"""

from typing import Any
from collections.abc import Callable
from colorama import Fore, Style


class ProgressTracker:
    """Handles progress tracking and display for installation processes."""

    def __init__(self, total_steps: int = 18):
        """
        Initialize progress tracker.

        Args:
            total_steps: Total number of steps in the process
        """
        self.total_steps = total_steps
        self.current_step = 0

    def update_progress(
        self,
        step: str,
        callback: Callable[[str, int, int], None] | None = None
    ) -> None:
        """
        Update installation progress with clean, professional display.

        Args:
            step: Description of current step
            callback: Optional custom progress callback
        """
        self.current_step += 1

        if callback:
            callback(step, self.current_step, self.total_steps)
        else:
            self._display_default_progress(step)

    def _display_default_progress(self, step: str) -> None:
        """Display default progress with color coding and progress bar."""
        # Clean progress display with proper formatting
        progress_bar = self._create_progress_bar()
        percentage = min(int((self.current_step / self.total_steps) * 100), 100)

        # Professional color scheme
        color = self._get_progress_color(percentage)

        # Clean, aligned output format with step counter
        step_indicator = f"({self.current_step:2d}/{self.total_steps})"
        print(
            f"\r{color}[{progress_bar}] {percentage:3d}% "
            f"{Style.RESET_ALL}{step_indicator} {step}",
            end="",
            flush=True
        )

        # Add newline when complete to prevent overlap
        if self.current_step == self.total_steps:
            print()  # Final newline for clean completion

    def _get_progress_color(self, percentage: int) -> str:
        """Get color based on progress percentage."""
        if percentage < 30:
            return Fore.YELLOW
        elif percentage < 80:
            return Fore.CYAN
        else:
            return Fore.GREEN

    def _create_progress_bar(self, width: int = 25) -> str:
        """Create clean visual progress bar with accurate calculation."""
        progress_ratio = min(self.current_step / self.total_steps, 1.0)
        filled = int(progress_ratio * width)
        bar = "█" * filled + "░" * (width - filled)
        return bar

    def reset(self) -> None:
        """Reset progress tracker to initial state."""
        self.current_step = 0

    def set_total_steps(self, total_steps: int) -> None:
        """Update total steps count."""
        self.total_steps = total_steps

    def get_progress_info(self) -> dict:
        """Get current progress information as dictionary."""
        percentage = min(int((self.current_step / self.total_steps) * 100), 100)
        return {
            'current_step': self.current_step,
            'total_steps': self.total_steps,
            'percentage': percentage,
            'is_complete': self.current_step >= self.total_steps
        }

    def is_complete(self) -> bool:
        """Check if progress is complete."""
        return self.current_step >= self.total_steps

    def skip_to_step(self, step_number: int) -> None:
        """Skip directly to a specific step number."""
        if 0 <= step_number <= self.total_steps:
            self.current_step = step_number

    def add_steps(self, additional_steps: int) -> None:
        """Add additional steps to the total."""
        self.total_steps += additional_steps

    def create_sub_tracker(self, sub_steps: int) -> 'ProgressTracker':
        """
        Create a sub-tracker for nested operations.

        Args:
            sub_steps: Number of steps in the sub-operation

        Returns:
            ProgressTracker: New tracker for sub-operation
        """
        return ProgressTracker(sub_steps)

    def display_completion_message(self, message: str = "Installation completed successfully!") -> None:
        """Display completion message with formatting."""
        print(f"\n{Fore.GREEN}✅ {message}{Style.RESET_ALL}")

    def display_error_message(self, message: str) -> None:
        """Display error message with formatting."""
        print(f"\n{Fore.RED}❌ {message}{Style.RESET_ALL}")

    def display_warning_message(self, message: str) -> None:
        """Display warning message with formatting."""
        print(f"\n{Fore.YELLOW}⚠️ {message}{Style.RESET_ALL}")


class MultiStageProgressTracker:
    """Progress tracker that handles multiple stages with different step counts."""

    def __init__(self):
        self.stages = {}
        self.current_stage = None
        self.stage_order = []

    def add_stage(self, stage_name: str, steps: int) -> None:
        """Add a new stage with specified step count."""
        self.stages[stage_name] = {
            'steps': steps,
            'tracker': ProgressTracker(steps),
            'completed': False
        }
        if stage_name not in self.stage_order:
            self.stage_order.append(stage_name)

    def start_stage(self, stage_name: str) -> ProgressTracker:
        """Start a specific stage and return its tracker."""
        if stage_name not in self.stages:
            raise ValueError(f"Stage '{stage_name}' not found")

        self.current_stage = stage_name
        tracker = self.stages[stage_name]['tracker']
        tracker.reset()

        print(f"\n{Fore.CYAN}🚀 Starting stage: {stage_name}{Style.RESET_ALL}")
        return tracker

    def complete_stage(self, stage_name: str) -> None:
        """Mark a stage as completed."""
        if stage_name in self.stages:
            self.stages[stage_name]['completed'] = True
            print(f"\n{Fore.GREEN}✅ Stage completed: {stage_name}{Style.RESET_ALL}")

    def get_overall_progress(self) -> dict:
        """Get overall progress across all stages."""
        total_steps = sum(stage['steps'] for stage in self.stages.values())
        completed_steps = sum(
            stage['tracker'].current_step for stage in self.stages.values()
        )

        percentage = int((completed_steps / total_steps) * 100) if total_steps > 0 else 0

        return {
            'completed_steps': completed_steps,
            'total_steps': total_steps,
            'percentage': percentage,
            'current_stage': self.current_stage,
            'completed_stages': [
                name for name, stage in self.stages.items() if stage['completed']
            ]
        }