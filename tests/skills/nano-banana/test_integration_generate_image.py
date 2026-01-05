"""
Integration tests for generate_image.py module with API mocking.

Tests cover:
- generate_image(): Success case with mock response
- generate_image(): API error retry behavior
- CLI interface: Argument parsing and execution

TDD Approach: All tests use mocked API calls (no real API calls).
"""

import sys
from unittest.mock import MagicMock, patch

import pytest

# Import functions to test
from generate_image import (
    MODEL_NAME,
    generate_image,
    main,
)


class TestGenerateImageSuccess:
    """Integration tests for successful image generation."""

    def test_generate_image_success_returns_metadata(
        self, env_with_api_key, temp_output_dir, mock_successful_response
    ):
        """Successful generation should return metadata dict."""
        output_path = temp_output_dir / "test_image.png"

        with patch("generate_image.genai") as mock_genai:
            mock_client = MagicMock()
            mock_genai.Client.return_value = mock_client
            mock_client.models.generate_content.return_value = mock_successful_response

            result = generate_image(
                prompt="A test image",
                output_path=str(output_path),
                max_retries=1
            )

        assert result["success"] is True
        assert result["model"] == MODEL_NAME
        assert result["prompt"] == "A test image"
        assert "output_path" in result
        assert "generation_time_seconds" in result
        assert "timestamp" in result

    def test_generate_image_creates_output_file(
        self, env_with_api_key, temp_output_dir, mock_successful_response
    ):
        """Successful generation should create output file."""
        output_path = temp_output_dir / "test_output.png"

        with patch("generate_image.genai") as mock_genai:
            mock_client = MagicMock()
            mock_genai.Client.return_value = mock_client
            mock_client.models.generate_content.return_value = mock_successful_response

            generate_image(
                prompt="Test prompt",
                output_path=str(output_path),
                max_retries=1
            )

        assert output_path.exists()
        # Verify file has content (fake PNG header)
        assert output_path.stat().st_size > 0

    def test_generate_image_with_style_includes_style_in_final_prompt(
        self, env_with_api_key, temp_output_dir, mock_successful_response
    ):
        """Style should be included in the final prompt."""
        output_path = temp_output_dir / "styled_image.png"
        style = "photorealistic, 8K"

        with patch("generate_image.genai") as mock_genai:
            mock_client = MagicMock()
            mock_genai.Client.return_value = mock_client
            mock_client.models.generate_content.return_value = mock_successful_response

            result = generate_image(
                prompt="Mountain landscape",
                output_path=str(output_path),
                style=style,
                max_retries=1
            )

        assert result["success"] is True
        assert result["style"] == style
        assert result["final_prompt"] == f"{style}: Mountain landscape"

    def test_generate_image_creates_nested_output_directory(
        self, env_with_api_key, temp_output_dir, mock_successful_response
    ):
        """Should create nested directories if they don't exist."""
        output_path = temp_output_dir / "nested" / "dir" / "image.png"

        with patch("generate_image.genai") as mock_genai:
            mock_client = MagicMock()
            mock_genai.Client.return_value = mock_client
            mock_client.models.generate_content.return_value = mock_successful_response

            result = generate_image(
                prompt="Test",
                output_path=str(output_path),
                max_retries=1
            )

        assert result["success"] is True
        assert output_path.parent.exists()

    def test_generate_image_records_attempt_count(
        self, env_with_api_key, temp_output_dir, mock_successful_response
    ):
        """Should record the number of attempts in result."""
        output_path = temp_output_dir / "test.png"

        with patch("generate_image.genai") as mock_genai:
            mock_client = MagicMock()
            mock_genai.Client.return_value = mock_client
            mock_client.models.generate_content.return_value = mock_successful_response

            result = generate_image(
                prompt="Test",
                output_path=str(output_path),
                max_retries=3
            )

        assert result["attempts"] == 1  # Success on first attempt


class TestGenerateImageRetry:
    """Integration tests for API error retry behavior."""

    def test_generate_image_retries_on_rate_limit_error(
        self, env_with_api_key, temp_output_dir, mock_successful_response, mock_rate_limit_error
    ):
        """Should retry on rate limit errors with backoff."""
        output_path = temp_output_dir / "retry_test.png"

        with patch("generate_image.genai") as mock_genai:
            mock_client = MagicMock()
            mock_genai.Client.return_value = mock_client

            # First call fails, second succeeds
            mock_client.models.generate_content.side_effect = [
                mock_rate_limit_error,
                mock_successful_response
            ]

            with patch("generate_image.time.sleep") as mock_sleep:
                result = generate_image(
                    prompt="Test",
                    output_path=str(output_path),
                    max_retries=3
                )

        assert result["success"] is True
        assert result["attempts"] == 2
        # Verify sleep was called for backoff
        mock_sleep.assert_called()

    def test_generate_image_fails_after_max_retries(
        self, env_with_api_key, temp_output_dir, mock_rate_limit_error
    ):
        """Should fail after exhausting all retries."""
        output_path = temp_output_dir / "fail_test.png"
        max_retries = 3

        with patch("generate_image.genai") as mock_genai:
            mock_client = MagicMock()
            mock_genai.Client.return_value = mock_client

            # All calls fail
            mock_client.models.generate_content.side_effect = mock_rate_limit_error

            with patch("generate_image.time.sleep"):
                result = generate_image(
                    prompt="Test",
                    output_path=str(output_path),
                    max_retries=max_retries
                )

        assert result["success"] is False
        assert result["attempts"] == max_retries
        assert "error" in result

    def test_generate_image_does_not_retry_on_api_key_error(
        self, env_with_api_key, temp_output_dir, mock_api_key_error
    ):
        """Should not retry on API key errors (non-retryable)."""
        output_path = temp_output_dir / "no_retry_test.png"

        with patch("generate_image.genai") as mock_genai:
            mock_client = MagicMock()
            mock_genai.Client.return_value = mock_client
            mock_client.models.generate_content.side_effect = mock_api_key_error

            with patch("generate_image.time.sleep"):
                result = generate_image(
                    prompt="Test",
                    output_path=str(output_path),
                    max_retries=5
                )

        assert result["success"] is False
        # Should break early without retrying
        assert mock_client.models.generate_content.call_count == 1

    def test_generate_image_handles_text_only_response(
        self, env_with_api_key, temp_output_dir, mock_text_only_response
    ):
        """Should fail gracefully when no image data in response."""
        output_path = temp_output_dir / "text_only.png"

        with patch("generate_image.genai") as mock_genai:
            mock_client = MagicMock()
            mock_genai.Client.return_value = mock_client
            mock_client.models.generate_content.return_value = mock_text_only_response

            with patch("generate_image.time.sleep"):
                result = generate_image(
                    prompt="Test",
                    output_path=str(output_path),
                    max_retries=2
                )

        assert result["success"] is False
        assert "No image data" in result["error"]


class TestGenerateImageCLI:
    """Tests for CLI interface."""

    def test_cli_requires_prompt_argument(self, capsys):
        """CLI should require --prompt argument."""
        test_args = ["generate_image.py", "-o", "output.png"]

        with patch.object(sys, "argv", test_args):
            with pytest.raises(SystemExit) as exc_info:
                main()

        # argparse exits with code 2 for missing required args
        assert exc_info.value.code == 2

    def test_cli_requires_output_argument(self, capsys):
        """CLI should require --output argument."""
        test_args = ["generate_image.py", "-p", "Test prompt"]

        with patch.object(sys, "argv", test_args):
            with pytest.raises(SystemExit) as exc_info:
                main()

        assert exc_info.value.code == 2

    def test_cli_accepts_valid_arguments(
        self, env_with_api_key, temp_output_dir, mock_successful_response
    ):
        """CLI should accept valid arguments and generate image."""
        output_path = temp_output_dir / "cli_test.png"
        test_args = [
            "generate_image.py",
            "-p", "A test prompt",
            "-o", str(output_path),
            "-r", "2K",
            "-a", "16:9"
        ]

        with patch.object(sys, "argv", test_args):
            with patch("generate_image.genai") as mock_genai:
                mock_client = MagicMock()
                mock_genai.Client.return_value = mock_client
                mock_client.models.generate_content.return_value = mock_successful_response

                # Should not raise
                main()

        assert output_path.exists()

    def test_cli_with_style_option(
        self, env_with_api_key, temp_output_dir, mock_successful_response
    ):
        """CLI should accept --style option."""
        output_path = temp_output_dir / "styled_cli.png"
        test_args = [
            "generate_image.py",
            "-p", "Mountain",
            "-o", str(output_path),
            "--style", "photorealistic"
        ]

        with patch.object(sys, "argv", test_args):
            with patch("generate_image.genai") as mock_genai:
                mock_client = MagicMock()
                mock_genai.Client.return_value = mock_client
                mock_client.models.generate_content.return_value = mock_successful_response

                main()

        assert output_path.exists()

    def test_cli_with_verbose_flag(
        self, env_with_api_key, temp_output_dir, mock_successful_response, capsys
    ):
        """CLI should provide verbose output with -v flag."""
        output_path = temp_output_dir / "verbose_test.png"
        test_args = [
            "generate_image.py",
            "-p", "Test",
            "-o", str(output_path),
            "-v"
        ]

        with patch.object(sys, "argv", test_args):
            with patch("generate_image.genai") as mock_genai:
                mock_client = MagicMock()
                mock_genai.Client.return_value = mock_client
                mock_client.models.generate_content.return_value = mock_successful_response

                main()

        captured = capsys.readouterr()
        assert "Model:" in captured.out

    def test_cli_exits_on_generation_failure(
        self, env_with_api_key, temp_output_dir, mock_rate_limit_error
    ):
        """CLI should exit with code 1 on generation failure."""
        output_path = temp_output_dir / "fail_cli.png"
        test_args = [
            "generate_image.py",
            "-p", "Test",
            "-o", str(output_path),
            "--max-retries", "1"
        ]

        with patch.object(sys, "argv", test_args):
            with patch("generate_image.genai") as mock_genai:
                mock_client = MagicMock()
                mock_genai.Client.return_value = mock_client
                mock_client.models.generate_content.side_effect = mock_rate_limit_error

                with patch("generate_image.time.sleep"):
                    with pytest.raises(SystemExit) as exc_info:
                        main()

        assert exc_info.value.code == 1
