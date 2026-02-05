"""MoAI-ADK CLI Progress Indicators

Modern progress bars and spinners using Rich for visual feedback during long operations.
"""

from contextlib import contextmanager
from typing import Any, Generator, Iterator, Optional, Sequence

from rich.console import Console
from rich.progress import (
    BarColumn,
    Progress,
    ProgressColumn,
    SpinnerColumn,
    TaskID,
    TaskProgressColumn,
    TextColumn,
    TimeElapsedColumn,
    TimeRemainingColumn,
)
from rich.spinner import Spinner
from rich.status import Status
from rich.style import Style
from rich.text import Text

from .theme import MOAI_COLORS

# Create console instance
console = Console()

# Claude Code themed styles
CLAUDE_STYLE = Style(color=MOAI_COLORS["primary"])
SUCCESS_STYLE = Style(color=MOAI_COLORS["secondary"])
ERROR_STYLE = Style(color=MOAI_COLORS["error"])
INFO_STYLE = Style(color=MOAI_COLORS["info"])
WARNING_STYLE = Style(color=MOAI_COLORS["warning"])


class MoAISpinnerColumn(ProgressColumn):
    """Custom spinner column with Claude Code terra cotta color."""

    def __init__(self, spinner_name: str = "dots") -> None:
        super().__init__()
        self.spinner = Spinner(spinner_name, style=CLAUDE_STYLE)

    def render(self, task: Any) -> Text:
        rendered = self.spinner.render(task.get_time())
        if isinstance(rendered, Text):
            return rendered
        return Text(str(rendered))


def create_progress_bar(
    description: str = "Processing",
    total: Optional[int] = None,
    transient: bool = False,
    auto_refresh: bool = True,
) -> Progress:
    """Create a styled progress bar for tracking operations.

    Args:
        description: Text shown before the progress bar
        total: Total number of steps (None for indeterminate)
        transient: Whether to remove progress bar when complete
        auto_refresh: Whether to auto-refresh the display

    Returns:
        Rich Progress instance

    Example:
        >>> with create_progress_bar("Updating files", total=10) as progress:
        ...     task = progress.add_task("Processing", total=10)
        ...     for i in range(10):
        ...         progress.update(task, advance=1)
        ...         time.sleep(0.1)
    """
    columns = [
        SpinnerColumn(spinner_name="dots", style=CLAUDE_STYLE),
        TextColumn("[bold]{task.description}[/bold]", style=CLAUDE_STYLE),
        BarColumn(
            bar_width=40,
            style=MOAI_COLORS["muted"],
            complete_style=CLAUDE_STYLE,
            finished_style=SUCCESS_STYLE,
        ),
        TaskProgressColumn(),
        TextColumn("•"),
        TimeElapsedColumn(),
    ]

    if total is not None:
        columns.insert(-1, TextColumn("•"))
        columns.insert(-1, TimeRemainingColumn())

    return Progress(
        *columns,
        console=console,
        transient=transient,
        auto_refresh=auto_refresh,
    )


def create_spinner(
    message: str = "Processing...",
    spinner_name: str = "dots",
) -> Status:
    """Create a styled spinner for indeterminate operations.

    Args:
        message: Text shown next to the spinner
        spinner_name: Name of spinner animation

    Returns:
        Rich Status instance

    Example:
        >>> with create_spinner("Checking for updates..."):
        ...     time.sleep(2)  # Long operation
    """
    return console.status(
        message,
        spinner=spinner_name,
        spinner_style=CLAUDE_STYLE,
    )


class ProgressContext:
    """Context manager for progress tracking with multiple tasks.

    Provides a convenient way to track progress across multiple operations
    with automatic cleanup and error handling.

    Example:
        >>> with ProgressContext("Updating MoAI-ADK") as ctx:
        ...     ctx.add_task("Downloading updates", total=100)
        ...     for i in range(100):
        ...         ctx.advance("Downloading updates")
        ...         time.sleep(0.01)
        ...     ctx.complete_task("Downloading updates")
    """

    def __init__(
        self,
        title: str,
        transient: bool = False,
    ) -> None:
        """Initialize progress context.

        Args:
            title: Overall operation title
            transient: Whether to remove progress display when done
        """
        self.title = title
        self.transient = transient
        self.progress: Optional[Progress] = None
        self.tasks: dict[str, TaskID] = {}
        self._started = False

    def __enter__(self) -> "ProgressContext":
        """Start the progress display."""
        self.progress = create_progress_bar(
            description=self.title,
            transient=self.transient,
        )
        self.progress.start()
        self._started = True
        return self

    def __exit__(self, exc_type: Any, exc_val: Any, exc_tb: Any) -> None:
        """Stop the progress display."""
        if self.progress:
            self.progress.stop()
        self._started = False

    def add_task(
        self,
        description: str,
        total: Optional[float] = None,
        visible: bool = True,
    ) -> TaskID:
        """Add a new task to track.

        Args:
            description: Task description
            total: Total steps (None for indeterminate)
            visible: Whether task is visible

        Returns:
            Task ID for future reference
        """
        if not self.progress:
            raise RuntimeError("ProgressContext not started")

        task_id = self.progress.add_task(
            description,
            total=total or 100,
            visible=visible,
        )
        self.tasks[description] = task_id
        return task_id

    def advance(
        self,
        description: str,
        advance: float = 1,
    ) -> None:
        """Advance a task's progress.

        Args:
            description: Task description (must match add_task)
            advance: Amount to advance
        """
        if not self.progress:
            return

        task_id = self.tasks.get(description)
        if task_id is not None:
            self.progress.advance(task_id, advance)

    def update_task(
        self,
        description: str,
        completed: Optional[float] = None,
        total: Optional[float] = None,
        new_description: Optional[str] = None,
    ) -> None:
        """Update a task's state.

        Args:
            description: Task description
            completed: New completed value
            total: New total value
            new_description: New description text
        """
        if not self.progress:
            return

        task_id = self.tasks.get(description)
        if task_id is not None:
            kwargs: dict[str, Any] = {}
            if completed is not None:
                kwargs["completed"] = completed
            if total is not None:
                kwargs["total"] = total
            if new_description is not None:
                kwargs["description"] = new_description
                # Update task mapping
                del self.tasks[description]
                self.tasks[new_description] = task_id

            self.progress.update(task_id, **kwargs)

    def complete_task(self, description: str) -> None:
        """Mark a task as complete.

        Args:
            description: Task description
        """
        if not self.progress:
            return

        task_id = self.tasks.get(description)
        if task_id is not None:
            task = self.progress.tasks[task_id]
            self.progress.update(task_id, completed=task.total)


class SpinnerContext:
    """Context manager for spinner with status updates.

    Provides a way to show an indeterminate spinner with updateable status text.

    Example:
        >>> with SpinnerContext("Processing") as spinner:
        ...     spinner.update("Step 1: Downloading...")
        ...     time.sleep(1)
        ...     spinner.update("Step 2: Installing...")
        ...     time.sleep(1)
        ...     spinner.success("Complete!")
    """

    def __init__(
        self,
        initial_message: str = "Processing...",
        spinner_name: str = "dots",
    ) -> None:
        """Initialize spinner context.

        Args:
            initial_message: Initial status message
            spinner_name: Name of spinner animation
        """
        self.initial_message = initial_message
        self.spinner_name = spinner_name
        self.status: Optional[Status] = None
        self._final_message: Optional[str] = None
        self._final_style: Optional[Style] = None

    def __enter__(self) -> "SpinnerContext":
        """Start the spinner."""
        self.status = console.status(
            self.initial_message,
            spinner=self.spinner_name,
            spinner_style=CLAUDE_STYLE,
        )
        self.status.start()
        return self

    def __exit__(self, exc_type: Any, exc_val: Any, exc_tb: Any) -> None:
        """Stop the spinner and show final message."""
        if self.status:
            self.status.stop()

        if self._final_message:
            console.print(self._final_message, style=self._final_style)

    def update(self, message: str) -> None:
        """Update the spinner message.

        Args:
            message: New status message
        """
        if self.status:
            self.status.update(message)

    def success(self, message: str) -> None:
        """Set success message to show after spinner stops.

        Args:
            message: Success message
        """
        self._final_message = f"✓ {message}"
        self._final_style = SUCCESS_STYLE

    def error(self, message: str) -> None:
        """Set error message to show after spinner stops.

        Args:
            message: Error message
        """
        self._final_message = f"✗ {message}"
        self._final_style = ERROR_STYLE

    def warning(self, message: str) -> None:
        """Set warning message to show after spinner stops.

        Args:
            message: Warning message
        """
        self._final_message = f"⚠ {message}"
        self._final_style = WARNING_STYLE

    def info(self, message: str) -> None:
        """Set info message to show after spinner stops.

        Args:
            message: Info message
        """
        self._final_message = f"ℹ {message}"
        self._final_style = INFO_STYLE


@contextmanager
def progress_steps(
    steps: Sequence[str],
    title: str = "Processing",
) -> Generator[Iterator[str], None, None]:
    """Context manager for tracking a sequence of steps.

    Yields each step name as it should be executed.

    Args:
        steps: List of step descriptions
        title: Overall operation title

    Yields:
        Iterator over step names

    Example:
        >>> steps = ["Download", "Extract", "Install", "Configure"]
        >>> with progress_steps(steps, "Installing") as step_iter:
        ...     for step in step_iter:
        ...         # Do work for this step
        ...         time.sleep(0.5)
    """
    with ProgressContext(title) as ctx:
        ctx.add_task(title, total=len(steps))

        def step_generator() -> Iterator[str]:
            for i, step in enumerate(steps):
                ctx.update_task(title, new_description=f"{title}: {step}")
                yield step
                ctx.advance(title)

        yield step_generator()


def print_step(
    step_number: int,
    total_steps: int,
    message: str,
    status: str = "running",
) -> None:
    """Print a formatted step indicator.

    Args:
        step_number: Current step number (1-indexed)
        total_steps: Total number of steps
        message: Step description
        status: One of 'running', 'complete', 'error', 'skipped'
    """
    status_symbols = {
        "running": ("→", CLAUDE_STYLE),
        "complete": ("✓", SUCCESS_STYLE),
        "error": ("✗", ERROR_STYLE),
        "skipped": ("○", Style(color=MOAI_COLORS["muted"])),
    }

    symbol, style = status_symbols.get(status, ("•", CLAUDE_STYLE))
    step_text = f"[{step_number}/{total_steps}]"

    console.print(f"{symbol} {step_text} {message}", style=style)
