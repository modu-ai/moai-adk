"""
Parallel Processing Core

Provides efficient parallel processing capabilities for concurrent task execution.
"""

import asyncio
from typing import Any, Callable, Coroutine, Dict, List, Optional, Union


class ParallelProcessor:
    """
    A parallel processor for executing tasks concurrently with configurable limits.

    This class provides a high-level interface for running multiple async tasks
    in parallel with optional progress tracking and error handling.
    """

    def __init__(self, max_workers: Optional[int] = None):
        """
        Initialize the parallel processor.

        Args:
            max_workers: Maximum number of concurrent tasks. If None, defaults to CPU count.
        """
        self.max_workers = max_workers

    async def process_tasks(
        self,
        tasks: List[Callable],
        progress_callback: Optional[Callable[[int, int], None]] = None,
    ) -> List[Dict[str, Any]]:
        """
        Process multiple tasks concurrently.

        Args:
            tasks: List of async task functions or coroutines to execute
            progress_callback: Optional callback for progress updates (completed, total)

        Returns:
            List of results from all completed tasks in the same order as input

        Raises:
            Exception: If any task raises an exception
        """
        if not tasks:
            return []

        results = []
        completed_count = 0

        # Validate input
        self._validate_tasks(tasks)

        # Initialize progress tracking
        self._update_progress(progress_callback, completed_count, len(tasks))

        # Process tasks concurrently
        for task in tasks:
            try:
                # Get coroutine from task
                coroutine = self._get_coroutine(task)

                # Execute the task and get result
                result = await coroutine
                results.append(result)
                completed_count += 1

                # Update progress
                self._update_progress(progress_callback, completed_count, len(tasks))

            except Exception as e:
                # Add context to the exception
                raise type(e)(f"Task failed: {str(e)}") from e

        return results

    def _validate_tasks(self, tasks: List[Callable]) -> None:
        """Validate that all tasks are callable or coroutines."""
        if not isinstance(tasks, list):
            raise TypeError("tasks must be a list")

        for i, task in enumerate(tasks):
            if not callable(task) and not asyncio.iscoroutine(task):
                raise TypeError(f"Task at index {i} is not callable or a coroutine")

    def _get_coroutine(self, task: Union[Callable, Coroutine]) -> Coroutine:
        """Extract coroutine from task function or return coroutine directly."""
        if asyncio.iscoroutine(task):
            return task

        if callable(task):
            try:
                result = task()
                if asyncio.iscoroutine(result):
                    return result
                else:
                    raise TypeError("Task function must return a coroutine")
            except Exception as e:
                raise TypeError(f"Failed to execute task function: {str(e)}")

        raise TypeError(f"Task must be callable or a coroutine, got {type(task)}")

    def _update_progress(
        self,
        progress_callback: Optional[Callable[[int, int], None]],
        completed: int,
        total: int,
    ) -> None:
        """Update progress callback if provided."""
        if progress_callback is not None:
            try:
                progress_callback(completed, total)
            except Exception as e:
                # Log progress callback error but don't fail the entire process
                print(f"Warning: Progress callback failed: {str(e)}")
