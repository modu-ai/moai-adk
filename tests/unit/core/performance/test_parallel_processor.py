"""
Parallel Processor Tests

Test cases for parallel processing core functionality.
"""

import asyncio
from typing import Any, Dict

import pytest


class TestParallelProcessor:
    """Test suite for parallel processing core functionality."""

    def test_parallel_processor_creation(self):
        """Test that parallel processor can be created successfully."""
        # This test should fail initially as the ParallelProcessor class doesn't exist
        from moai_adk.core.performance.parallel_processor import ParallelProcessor

        processor = ParallelProcessor()
        assert processor is not None

    def test_parallel_processor_with_empty_tasks(self):
        """Test processing empty task list."""
        # This test should fail initially
        from moai_adk.core.performance.parallel_processor import ParallelProcessor

        processor = ParallelProcessor()
        result = asyncio.run(processor.process_tasks([]))
        assert result == []

    def test_parallel_processor_with_single_task(self):
        """Test processing single task."""
        # This test should fail initially
        from moai_adk.core.performance.parallel_processor import ParallelProcessor

        processor = ParallelProcessor()

        async def sample_task(task_id: str) -> Dict[str, Any]:
            return {"id": task_id, "status": "completed"}

        result = asyncio.run(processor.process_tasks([lambda: sample_task("sample_task")]))
        assert len(result) == 1
        assert result[0]["id"] == "sample_task"
        assert result[0]["status"] == "completed"

    def test_parallel_processor_with_concurrent_tasks(self):
        """Test processing multiple tasks concurrently."""
        # This test should fail initially
        from moai_adk.core.performance.parallel_processor import ParallelProcessor

        processor = ParallelProcessor()

        async def sample_task(task_id: int) -> Dict[str, Any]:
            # Simulate some work
            await asyncio.sleep(0.1)
            return {"id": task_id, "status": "completed"}

        tasks = [sample_task(i) for i in range(5)]
        result = asyncio.run(processor.process_tasks(tasks))
        assert len(result) == 5

        # Check all tasks completed
        for i, task_result in enumerate(result):
            assert task_result["id"] == i
            assert task_result["status"] == "completed"

    def test_parallel_processor_with_error_handling(self):
        """Test error handling in parallel processing."""
        # This test should fail initially
        from moai_adk.core.performance.parallel_processor import ParallelProcessor

        processor = ParallelProcessor()

        async def failing_task(task_id: str) -> Dict[str, Any]:
            if task_id == "fail":
                raise ValueError("Task failed")
            return {"id": task_id, "status": "completed"}

        tasks = [lambda: failing_task("normal"), lambda: failing_task("fail")]

        # This should raise an exception
        with pytest.raises(ValueError, match="Task failed"):
            asyncio.run(processor.process_tasks(tasks))

    def test_parallel_processor_with_max_workers(self):
        """Test parallel processor with maximum worker limit."""
        # This test should fail initially
        from moai_adk.core.performance.parallel_processor import ParallelProcessor

        processor = ParallelProcessor(max_workers=2)

        async def sample_task(task_id: int) -> Dict[str, Any]:
            await asyncio.sleep(0.1)
            return {"id": task_id, "status": "completed"}

        tasks = [lambda: sample_task(i) for i in range(5)]
        result = asyncio.run(processor.process_tasks(tasks))
        assert len(result) == 5

    def test_parallel_processor_progress_callback(self):
        """Test progress callback functionality."""
        # This test should fail initially
        from moai_adk.core.performance.parallel_processor import ParallelProcessor

        processor = ParallelProcessor()
        progress_calls = []

        def progress_callback(completed: int, total: int):
            progress_calls.append((completed, total))

        async def sample_task(task_id: int) -> Dict[str, Any]:
            await asyncio.sleep(0.1)
            return {"id": task_id, "status": "completed"}

        tasks = [lambda: sample_task(i) for i in range(3)]
        asyncio.run(processor.process_tasks(tasks, progress_callback))

        # Check that progress was called
        assert len(progress_calls) > 0
        # Should have been called with (0, 3), (1, 3), (2, 3), (3, 3)
        assert progress_calls[0] == (0, 3)
        assert progress_calls[-1] == (3, 3)

    def test_parallel_processor_validation(self):
        """Test input validation."""
        from moai_adk.core.performance.parallel_processor import ParallelProcessor

        processor = ParallelProcessor()

        # Test invalid tasks type
        with pytest.raises(TypeError, match="tasks must be a list"):
            asyncio.run(processor.process_tasks("not a list"))

        # Test non-callable task
        with pytest.raises(TypeError, match="is not callable or a coroutine"):
            asyncio.run(processor.process_tasks([42]))

        # Test function that doesn't return coroutine
        def bad_task():
            return "not a coroutine"

        with pytest.raises(TypeError, match="must return a coroutine"):
            asyncio.run(processor.process_tasks([bad_task]))
