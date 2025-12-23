"""
Integration tests for batch_generate.py module with API mocking.

Tests cover:
- generate_single_image(): Async image generation
- run_batch_generation(): Batch processing with concurrency
- save_batch_report(): Report generation
- CLI interface: Argument parsing

TDD Approach: All tests use mocked API calls (no real API calls).
"""

import asyncio
import json
import sys
from pathlib import Path
from unittest.mock import MagicMock, patch, AsyncMock

import pytest

from batch_generate import (
    generate_single_image,
    run_batch_generation,
    save_batch_report,
    main,
    ImageTask,
    BatchResult,
    MODEL_NAME,
)


class TestGenerateSingleImageAsync:
    """Tests for async single image generation."""

    @pytest.fixture
    def mock_async_genai_client(self, mock_successful_response):
        """Create a mock genai client for async testing."""
        mock_client = MagicMock()
        mock_client.models.generate_content.return_value = mock_successful_response
        return mock_client

    @pytest.mark.asyncio
    async def test_successful_single_image_generation(
        self, temp_output_dir, mock_async_genai_client, mock_successful_response
    ):
        """Should successfully generate a single image."""
        task = ImageTask(
            prompt="Test prompt",
            output_path=str(temp_output_dir / "test.png"),
            task_id=1
        )

        # Mock the executor to return sync result
        with patch("asyncio.get_event_loop") as mock_loop:
            mock_executor = AsyncMock(return_value=mock_successful_response)
            mock_loop.return_value.run_in_executor = mock_executor

            result = await generate_single_image(
                mock_async_genai_client,
                task,
                max_retries=1,
                verbose=False
            )

        assert result["success"] is True
        assert result["task_id"] == 1
        assert result["prompt"] == "Test prompt"

    @pytest.mark.asyncio
    async def test_single_image_retries_on_error(
        self, temp_output_dir, mock_successful_response, mock_rate_limit_error
    ):
        """Should retry on rate limit errors."""
        task = ImageTask(
            prompt="Retry test",
            output_path=str(temp_output_dir / "retry.png"),
            task_id=2
        )

        mock_client = MagicMock()

        call_count = 0

        def side_effect_func(*args, **kwargs):
            nonlocal call_count
            call_count += 1
            if call_count == 1:
                raise mock_rate_limit_error
            return mock_successful_response

        with patch("asyncio.get_event_loop") as mock_loop:
            mock_loop.return_value.run_in_executor = AsyncMock(
                side_effect=side_effect_func
            )

            with patch("asyncio.sleep", new_callable=AsyncMock):
                result = await generate_single_image(
                    mock_client,
                    task,
                    max_retries=3,
                    verbose=False
                )

        assert result["success"] is True
        assert result["attempts"] == 2

    @pytest.mark.asyncio
    async def test_single_image_fails_after_max_retries(
        self, temp_output_dir, mock_rate_limit_error
    ):
        """Should fail after exhausting retries."""
        task = ImageTask(
            prompt="Fail test",
            output_path=str(temp_output_dir / "fail.png"),
            task_id=3
        )

        mock_client = MagicMock()

        with patch("asyncio.get_event_loop") as mock_loop:
            mock_loop.return_value.run_in_executor = AsyncMock(
                side_effect=mock_rate_limit_error
            )

            with patch("asyncio.sleep", new_callable=AsyncMock):
                result = await generate_single_image(
                    mock_client,
                    task,
                    max_retries=2,
                    verbose=False
                )

        assert result["success"] is False
        assert result["attempts"] == 2
        assert "error" in result


class TestRunBatchGeneration:
    """Tests for batch generation execution."""

    @pytest.mark.asyncio
    async def test_batch_generation_processes_all_tasks(
        self, env_with_api_key, temp_output_dir, mock_successful_response
    ):
        """Should process all tasks in batch."""
        tasks = [
            ImageTask(prompt=f"Prompt {i}", output_path=str(temp_output_dir / f"img_{i}.png"), task_id=i)
            for i in range(1, 4)
        ]

        with patch("batch_generate.genai") as mock_genai:
            mock_client = MagicMock()
            mock_genai.Client.return_value = mock_client

            with patch("asyncio.get_event_loop") as mock_loop:
                mock_loop.return_value.run_in_executor = AsyncMock(
                    return_value=mock_successful_response
                )

                with patch("asyncio.sleep", new_callable=AsyncMock):
                    result = await run_batch_generation(
                        tasks,
                        concurrency=2,
                        max_retries=1,
                        verbose=False,
                        delay_between_tasks=0
                    )

        assert result.total == 3
        assert result.successful == 3
        assert result.failed == 0

    @pytest.mark.asyncio
    async def test_batch_generation_handles_partial_failure(
        self, env_with_api_key, temp_output_dir, mock_successful_response, mock_rate_limit_error
    ):
        """Should handle partial failures in batch."""
        tasks = [
            ImageTask(prompt=f"Prompt {i}", output_path=str(temp_output_dir / f"img_{i}.png"), task_id=i)
            for i in range(1, 4)
        ]

        call_count = 0

        def side_effect_func(*args, **kwargs):
            nonlocal call_count
            call_count += 1
            if call_count == 2:
                raise mock_rate_limit_error
            return mock_successful_response

        with patch("batch_generate.genai") as mock_genai:
            mock_client = MagicMock()
            mock_genai.Client.return_value = mock_client

            with patch("asyncio.get_event_loop") as mock_loop:
                mock_loop.return_value.run_in_executor = AsyncMock(
                    side_effect=side_effect_func
                )

                with patch("asyncio.sleep", new_callable=AsyncMock):
                    result = await run_batch_generation(
                        tasks,
                        concurrency=1,
                        max_retries=1,
                        verbose=False,
                        delay_between_tasks=0
                    )

        assert result.total == 3
        assert result.successful == 2
        assert result.failed == 1

    @pytest.mark.asyncio
    async def test_batch_generation_records_timing(
        self, env_with_api_key, temp_output_dir, mock_successful_response
    ):
        """Should record start time, end time, and duration."""
        tasks = [
            ImageTask(prompt="Test", output_path=str(temp_output_dir / "test.png"), task_id=1)
        ]

        with patch("batch_generate.genai") as mock_genai:
            mock_client = MagicMock()
            mock_genai.Client.return_value = mock_client

            with patch("asyncio.get_event_loop") as mock_loop:
                mock_loop.return_value.run_in_executor = AsyncMock(
                    return_value=mock_successful_response
                )

                with patch("asyncio.sleep", new_callable=AsyncMock):
                    result = await run_batch_generation(
                        tasks,
                        concurrency=1,
                        max_retries=1,
                        delay_between_tasks=0
                    )

        assert result.start_time is not None
        assert result.end_time is not None
        assert result.total_duration_seconds >= 0


class TestSaveBatchReport:
    """Tests for batch report saving."""

    def test_save_report_creates_json_file(self, tmp_path):
        """Should create a valid JSON report file."""
        result = BatchResult(
            total=5,
            successful=4,
            failed=1,
            results=[{"task_id": 1, "success": True}],
            errors=[{"task_id": 2, "error": "Test error"}]
        )
        result.start_time = result.end_time = None
        result.total_duration_seconds = 10.5

        report_path = tmp_path / "report.json"
        save_batch_report(result, str(report_path))

        assert report_path.exists()

        with open(report_path) as f:
            report = json.load(f)

        assert report["summary"]["total"] == 5
        assert report["summary"]["successful"] == 4
        assert report["summary"]["failed"] == 1
        assert len(report["results"]) == 1
        assert len(report["errors"]) == 1

    def test_report_includes_success_rate(self, tmp_path):
        """Report should include calculated success rate."""
        result = BatchResult(total=10, successful=8, failed=2)
        result.total_duration_seconds = 5.0

        report_path = tmp_path / "rate_report.json"
        save_batch_report(result, str(report_path))

        with open(report_path) as f:
            report = json.load(f)

        assert report["summary"]["success_rate"] == "80.0%"


class TestBatchGenerateCLI:
    """Tests for CLI interface."""

    def test_cli_requires_config_or_prompts(self, capsys):
        """CLI should require either --config or --prompts."""
        test_args = ["batch_generate.py", "-d", "output/"]

        with patch.object(sys, "argv", test_args):
            with pytest.raises(SystemExit) as exc_info:
                main()

        assert exc_info.value.code == 2

    def test_cli_requires_output_dir(self, capsys):
        """CLI should require --output-dir."""
        test_args = ["batch_generate.py", "--prompts", "Test"]

        with patch.object(sys, "argv", test_args):
            with pytest.raises(SystemExit) as exc_info:
                main()

        assert exc_info.value.code == 2

    def test_cli_accepts_prompts_list(
        self, env_with_api_key, temp_output_dir, mock_successful_response
    ):
        """CLI should accept multiple prompts."""
        test_args = [
            "batch_generate.py",
            "--prompts", "Cat", "Dog", "Bird",
            "-d", str(temp_output_dir)
        ]

        with patch.object(sys, "argv", test_args):
            with patch("batch_generate.genai") as mock_genai:
                mock_client = MagicMock()
                mock_genai.Client.return_value = mock_client

                with patch("batch_generate.asyncio.run") as mock_run:
                    mock_result = BatchResult(total=3, successful=3, failed=0)
                    mock_result.total_duration_seconds = 1.0
                    mock_run.return_value = mock_result

                    main()

        # Verify asyncio.run was called
        mock_run.assert_called_once()

    def test_cli_accepts_config_file(
        self, env_with_api_key, temp_output_dir, tmp_path, mock_successful_response
    ):
        """CLI should accept config file."""
        config_file = tmp_path / "config.json"
        config_file.write_text(json.dumps({
            "images": ["Prompt 1", "Prompt 2"]
        }))

        test_args = [
            "batch_generate.py",
            "-c", str(config_file),
            "-d", str(temp_output_dir)
        ]

        with patch.object(sys, "argv", test_args):
            with patch("batch_generate.genai") as mock_genai:
                mock_client = MagicMock()
                mock_genai.Client.return_value = mock_client

                with patch("batch_generate.asyncio.run") as mock_run:
                    mock_result = BatchResult(total=2, successful=2, failed=0)
                    mock_result.total_duration_seconds = 1.0
                    mock_run.return_value = mock_result

                    main()

        mock_run.assert_called_once()

    def test_cli_with_report_option(
        self, env_with_api_key, temp_output_dir, tmp_path
    ):
        """CLI should save report when --report is specified."""
        report_path = tmp_path / "report.json"

        test_args = [
            "batch_generate.py",
            "--prompts", "Test",
            "-d", str(temp_output_dir),
            "--report", str(report_path)
        ]

        with patch.object(sys, "argv", test_args):
            with patch("batch_generate.genai") as mock_genai:
                mock_client = MagicMock()
                mock_genai.Client.return_value = mock_client

                with patch("batch_generate.asyncio.run") as mock_run:
                    mock_result = BatchResult(total=1, successful=1, failed=0)
                    mock_result.total_duration_seconds = 1.0
                    mock_result.results = []
                    mock_result.errors = []
                    mock_run.return_value = mock_result

                    main()

        # Report should be created
        assert report_path.exists()

    def test_cli_exits_with_error_on_failures(
        self, env_with_api_key, temp_output_dir
    ):
        """CLI should exit with code 1 if any images fail."""
        test_args = [
            "batch_generate.py",
            "--prompts", "Test",
            "-d", str(temp_output_dir)
        ]

        with patch.object(sys, "argv", test_args):
            with patch("batch_generate.genai"):
                with patch("batch_generate.asyncio.run") as mock_run:
                    mock_result = BatchResult(total=1, successful=0, failed=1)
                    mock_result.total_duration_seconds = 1.0
                    mock_result.errors = [{"task_id": 1, "error": "Failed"}]
                    mock_run.return_value = mock_result

                    with pytest.raises(SystemExit) as exc_info:
                        main()

        assert exc_info.value.code == 1

    def test_cli_concurrency_option(
        self, env_with_api_key, temp_output_dir
    ):
        """CLI should accept --concurrency option."""
        test_args = [
            "batch_generate.py",
            "--prompts", "Test",
            "-d", str(temp_output_dir),
            "--concurrency", "5"
        ]

        with patch.object(sys, "argv", test_args):
            with patch("batch_generate.genai"):
                with patch("batch_generate.asyncio.run") as mock_run:
                    mock_result = BatchResult(total=1, successful=1, failed=0)
                    mock_result.total_duration_seconds = 1.0
                    mock_run.return_value = mock_result

                    main()

        # Verify asyncio.run was called (concurrency is passed internally)
        mock_run.assert_called_once()

    def test_cli_empty_config_exits_with_error(
        self, env_with_api_key, temp_output_dir, tmp_path
    ):
        """CLI should exit if config has no valid tasks."""
        config_file = tmp_path / "empty_config.json"
        config_file.write_text(json.dumps({"images": []}))

        test_args = [
            "batch_generate.py",
            "-c", str(config_file),
            "-d", str(temp_output_dir)
        ]

        with patch.object(sys, "argv", test_args):
            with pytest.raises(SystemExit) as exc_info:
                main()

        assert exc_info.value.code == 1
