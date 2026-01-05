"""
Unit tests for generate_image.py module.

Tests cover:
- validate_aspect_ratio(): Supported and unsupported aspect ratios
- validate_resolution(): 1K, 2K, 4K and lowercase conversion
- build_prompt_with_style(): Style prefix application
- calculate_backoff_delay(): Exponential backoff calculation
- get_api_key(): Environment variable loading

TDD Approach: RED-GREEN-REFACTOR cycle
"""

import os
import random
from unittest.mock import patch

import pytest

# Import functions to test
from generate_image import (
    BASE_DELAY,
    MAX_DELAY,
    SUPPORTED_ASPECT_RATIOS,
    SUPPORTED_RESOLUTIONS,
    build_prompt_with_style,
    calculate_backoff_delay,
    get_api_key,
    validate_aspect_ratio,
    validate_resolution,
)


class TestValidateAspectRatio:
    """Tests for validate_aspect_ratio() function."""

    @pytest.mark.parametrize("ratio", SUPPORTED_ASPECT_RATIOS)
    def test_supported_aspect_ratios_return_unchanged(self, ratio):
        """Supported aspect ratios should be returned unchanged."""
        result = validate_aspect_ratio(ratio)
        assert result == ratio

    @pytest.mark.parametrize("invalid_ratio", [
        "4:4",      # Not in supported list
        "10:9",     # Not in supported list
        "invalid",  # Invalid format
        "",         # Empty string
        "16:10",    # Common but not supported
    ])
    def test_unsupported_aspect_ratios_return_default(self, invalid_ratio, capsys):
        """Unsupported aspect ratios should return default 16:9."""
        result = validate_aspect_ratio(invalid_ratio)
        assert result == "16:9"

        # Verify warning message was printed
        captured = capsys.readouterr()
        assert "Warning" in captured.out
        assert invalid_ratio in captured.out

    def test_default_aspect_ratio_when_none_like_values(self, capsys):
        """None-like string values should return default."""
        result = validate_aspect_ratio("none")
        assert result == "16:9"


class TestValidateResolution:
    """Tests for validate_resolution() function."""

    @pytest.mark.parametrize("resolution", SUPPORTED_RESOLUTIONS)
    def test_supported_resolutions_return_unchanged(self, resolution):
        """Supported resolutions (uppercase) should be returned unchanged."""
        result = validate_resolution(resolution)
        assert result == resolution

    @pytest.mark.parametrize("lowercase,expected", [
        ("1k", "1K"),
        ("2k", "2K"),
        ("4k", "4K"),
    ])
    def test_lowercase_resolutions_normalized_to_uppercase(self, lowercase, expected):
        """Lowercase resolutions should be normalized to uppercase."""
        result = validate_resolution(lowercase)
        assert result == expected

    @pytest.mark.parametrize("mixed_case,expected", [
        ("1K", "1K"),
        ("2k", "2K"),
        ("4K", "4K"),
    ])
    def test_mixed_case_resolutions_normalized(self, mixed_case, expected):
        """Mixed case resolutions should be normalized to uppercase."""
        result = validate_resolution(mixed_case)
        assert result == expected

    @pytest.mark.parametrize("invalid_resolution", [
        "3K",       # Not in supported list
        "8K",       # Not in supported list
        "HD",       # Different format
        "",         # Empty string
        "invalid",  # Invalid format
    ])
    def test_unsupported_resolutions_return_default(self, invalid_resolution, capsys):
        """Unsupported resolutions should return default 2K."""
        result = validate_resolution(invalid_resolution)
        assert result == "2K"

        # Verify warning message was printed
        captured = capsys.readouterr()
        assert "Warning" in captured.out


class TestBuildPromptWithStyle:
    """Tests for build_prompt_with_style() function."""

    def test_prompt_without_style_returns_original(self):
        """Prompt without style should return the original prompt."""
        prompt = "A beautiful mountain landscape"
        result = build_prompt_with_style(prompt)
        assert result == prompt

    def test_prompt_with_none_style_returns_original(self):
        """Prompt with None style should return the original prompt."""
        prompt = "A beautiful mountain landscape"
        result = build_prompt_with_style(prompt, None)
        assert result == prompt

    def test_prompt_with_style_prepends_style_prefix(self):
        """Prompt with style should prepend style prefix with colon."""
        prompt = "A beautiful mountain landscape"
        style = "photorealistic"
        result = build_prompt_with_style(prompt, style)
        assert result == f"{style}: {prompt}"

    def test_prompt_with_complex_style_prefix(self):
        """Complex style with multiple descriptors should work."""
        prompt = "A cat"
        style = "photorealistic, 8K, detailed, cinematic lighting"
        result = build_prompt_with_style(prompt, style)
        assert result == f"{style}: {prompt}"

    def test_empty_prompt_with_style(self):
        """Empty prompt with style should still apply style prefix."""
        prompt = ""
        style = "minimalist"
        result = build_prompt_with_style(prompt, style)
        assert result == f"{style}: {prompt}"

    def test_prompt_with_empty_style_returns_original(self):
        """Prompt with empty string style should return original (falsy)."""
        prompt = "A beautiful scene"
        style = ""
        result = build_prompt_with_style(prompt, style)
        # Empty string is falsy, so should return original prompt
        assert result == prompt


class TestCalculateBackoffDelay:
    """Tests for calculate_backoff_delay() function."""

    def test_first_attempt_returns_base_delay_range(self):
        """First attempt (0) should return delay around BASE_DELAY."""
        # With jitter disabled
        delay = calculate_backoff_delay(0, jitter=False)
        assert delay == BASE_DELAY

    def test_exponential_backoff_without_jitter(self):
        """Delays should increase exponentially without jitter."""
        delays = [calculate_backoff_delay(i, jitter=False) for i in range(5)]

        # Verify exponential increase (each should be 2x previous)
        assert delays[1] == delays[0] * 2
        assert delays[2] == delays[1] * 2
        assert delays[3] == delays[2] * 2

    def test_delay_capped_at_max_delay(self):
        """Delay should not exceed MAX_DELAY."""
        # High attempt number should still be capped
        delay = calculate_backoff_delay(100, jitter=False)
        assert delay == MAX_DELAY

    def test_jitter_adds_randomness(self):
        """With jitter enabled, delays should have randomness."""
        # Set seed for reproducibility
        random.seed(42)
        delays_with_jitter = [calculate_backoff_delay(1, jitter=True) for _ in range(10)]

        # Check that delays are not all the same (jitter adds variation)
        unique_delays = set(delays_with_jitter)
        assert len(unique_delays) > 1

    def test_jitter_within_expected_range(self):
        """Jittered delay should be between 50% and 150% of base delay."""
        base_delay = BASE_DELAY * 2  # Attempt 1

        # Run multiple times to check bounds
        for _ in range(100):
            delay = calculate_backoff_delay(1, jitter=True)
            min_expected = base_delay * 0.5
            max_expected = base_delay * 1.5
            assert min_expected <= delay <= max_expected

    def test_negative_attempt_handled(self):
        """Negative attempt numbers should still work (edge case)."""
        delay = calculate_backoff_delay(-1, jitter=False)
        # 2^-1 = 0.5, so BASE_DELAY * 0.5
        expected = BASE_DELAY * 0.5
        assert delay == expected


class TestGetApiKey:
    """Tests for get_api_key() function."""

    def test_returns_api_key_from_environment(self, env_with_api_key):
        """Should return API key when set in environment."""
        result = get_api_key()
        assert result == env_with_api_key

    def test_exits_when_api_key_not_set(self, env_without_api_key, capsys):
        """Should exit with error when API key not set."""
        with pytest.raises(SystemExit) as exc_info:
            get_api_key()

        assert exc_info.value.code == 1

        # Verify error message was printed
        captured = capsys.readouterr()
        assert "Error" in captured.out
        assert "GOOGLE_API_KEY" in captured.out

    def test_returns_non_empty_string(self, env_with_api_key):
        """API key should be a non-empty string."""
        result = get_api_key()
        assert isinstance(result, str)
        assert len(result) > 0

    def test_api_key_with_whitespace_handling(self):
        """API key with leading/trailing whitespace should work."""
        test_key = "  test_key_with_spaces  "
        with patch.dict(os.environ, {"GOOGLE_API_KEY": test_key}):
            result = get_api_key()
            # The function returns the key as-is (with spaces)
            assert result == test_key
