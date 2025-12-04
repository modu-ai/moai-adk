"""Unit tests for moai_adk.utils.common module.

Tests for basic functionality of utility classes and functions.
"""

import asyncio
from datetime import datetime
from pathlib import Path
from unittest.mock import AsyncMock, MagicMock, patch

import pytest

from moai_adk.utils.common import (
    HTTPClient,
    HTTPResponse,
    RateLimiter,
    RateLimitError,
    calculate_score,
    create_report_path,
    extract_links_from_text,
    format_duration,
    get_summary_stats,
    is_valid_url,
)


class TestHTTPResponse:
    """Test HTTPResponse dataclass."""

    def test_create_response_with_defaults(self):
        """Test creating HTTPResponse with required fields."""
        response = HTTPResponse(
            status_code=200,
            url="https://example.com",
            load_time=0.5,
            success=True,
        )
        assert response.status_code == 200
        assert response.url == "https://example.com"
        assert response.load_time == 0.5
        assert response.success is True
        assert response.error_message is None
        assert isinstance(response.timestamp, datetime)

    def test_create_response_with_error(self):
        """Test creating HTTPResponse with error message."""
        response = HTTPResponse(
            status_code=404,
            url="https://example.com",
            load_time=0.2,
            success=False,
            error_message="Not found",
        )
        assert response.status_code == 404
        assert response.success is False
        assert response.error_message == "Not found"

    def test_response_post_init_none_timestamp(self):
        """Test post_init handles None timestamp."""
        response = HTTPResponse(
            status_code=200,
            url="https://example.com",
            load_time=0.5,
            success=True,
            timestamp=None,
        )
        assert isinstance(response.timestamp, datetime)


class TestHTTPClient:
    """Test HTTPClient class."""

    def test_client_initialization(self):
        """Test HTTPClient initialization."""
        client = HTTPClient(max_concurrent=3, timeout=5)
        assert client.max_concurrent == 3
        assert client.timeout == 5
        assert client.session is None

    def test_client_default_values(self):
        """Test HTTPClient uses default values."""
        client = HTTPClient()
        assert client.max_concurrent == 5
        assert client.timeout == 10

    @pytest.mark.asyncio
    async def test_client_fetch_url_no_session(self):
        """Test fetch_url returns error when session not initialized."""
        client = HTTPClient()
        result = await client.fetch_url("https://example.com")
        assert result.success is False
        assert result.status_code == 0
        assert "Session not initialized" in result.error_message


class TestRateLimiter:
    """Test RateLimiter class."""

    def test_rate_limiter_initialization(self):
        """Test RateLimiter initialization."""
        limiter = RateLimiter(max_requests=5, time_window=30)
        assert limiter.max_requests == 5
        assert limiter.time_window == 30
        assert limiter.requests == []

    def test_rate_limiter_defaults(self):
        """Test RateLimiter default values."""
        limiter = RateLimiter()
        assert limiter.max_requests == 10
        assert limiter.time_window == 60

    def test_can_make_request_empty(self):
        """Test can_make_request when no requests made."""
        limiter = RateLimiter(max_requests=5)
        assert limiter.can_make_request() is True

    def test_can_make_request_below_limit(self):
        """Test can_make_request below limit."""
        limiter = RateLimiter(max_requests=5)
        for _ in range(3):
            limiter.add_request()
        assert limiter.can_make_request() is True

    def test_can_make_request_at_limit(self):
        """Test can_make_request at limit."""
        limiter = RateLimiter(max_requests=2)
        for _ in range(2):
            limiter.add_request()
        assert limiter.can_make_request() is False

    def test_add_request_exceeds_limit(self):
        """Test add_request raises error when limit exceeded."""
        limiter = RateLimiter(max_requests=1)
        limiter.add_request()
        with pytest.raises(RateLimitError):
            limiter.add_request()

    @pytest.mark.asyncio
    async def test_wait_if_needed_under_limit(self):
        """Test wait_if_needed when under limit."""
        limiter = RateLimiter(max_requests=5, time_window=60)
        await limiter.wait_if_needed()  # Should not wait


class TestUtilityFunctions:
    """Test utility functions."""

    def test_is_valid_url_valid(self):
        """Test is_valid_url with valid URL."""
        assert is_valid_url("https://example.com") is True
        assert is_valid_url("http://example.com") is True

    def test_is_valid_url_invalid(self):
        """Test is_valid_url with invalid URL."""
        assert is_valid_url("not a url") is False
        assert is_valid_url("") is False
        assert is_valid_url("example.com") is False

    def test_extract_links_from_text_markdown(self):
        """Test extracting markdown links."""
        text = "[Click here](https://example.com) and [Visit](https://another.com)"
        links = extract_links_from_text(text)
        assert "https://example.com" in links
        assert "https://another.com" in links

    def test_extract_links_from_text_plain_urls(self):
        """Test extracting plain URLs."""
        text = "Visit https://example.com or https://another.com"
        links = extract_links_from_text(text)
        assert "https://example.com" in links
        assert "https://another.com" in links

    def test_extract_links_removes_duplicates(self):
        """Test that duplicates are removed."""
        text = "[Link](https://example.com) and https://example.com"
        links = extract_links_from_text(text)
        assert links.count("https://example.com") == 1

    def test_format_duration_milliseconds(self):
        """Test format_duration with milliseconds."""
        result = format_duration(0.5)
        assert "ms" in result

    def test_format_duration_seconds(self):
        """Test format_duration with seconds."""
        result = format_duration(15.5)
        assert "s" in result
        assert "m" not in result

    def test_format_duration_minutes(self):
        """Test format_duration with minutes."""
        result = format_duration(120)
        assert "m" in result

    def test_format_duration_hours(self):
        """Test format_duration with hours."""
        result = format_duration(3600)
        assert "h" in result

    def test_calculate_score_basic(self):
        """Test calculate_score with default weights."""
        result = calculate_score([80, 90, 100])
        assert result == pytest.approx(90.0)

    def test_calculate_score_empty(self):
        """Test calculate_score with empty list."""
        result = calculate_score([])
        assert result == 0.0

    def test_calculate_score_weighted(self):
        """Test calculate_score with custom weights."""
        result = calculate_score([80, 90], weights=[1.0, 3.0])
        assert result == pytest.approx(87.5)

    def test_calculate_score_mismatched_lengths(self):
        """Test calculate_score raises error for mismatched lengths."""
        with pytest.raises(ValueError):
            calculate_score([80, 90], weights=[1.0])

    def test_get_summary_stats_basic(self):
        """Test get_summary_stats basic functionality."""
        stats = get_summary_stats([10, 20, 30])
        assert stats["mean"] == 20.0
        assert stats["min"] == 10.0
        assert stats["max"] == 30.0

    def test_get_summary_stats_single_value(self):
        """Test get_summary_stats with single value."""
        stats = get_summary_stats([42])
        assert stats["mean"] == 42.0
        assert stats["min"] == 42.0
        assert stats["max"] == 42.0
        assert stats["std"] == 0.0

    def test_get_summary_stats_empty(self):
        """Test get_summary_stats with empty list."""
        stats = get_summary_stats([])
        assert stats["mean"] == 0.0
        assert stats["min"] == 0.0
        assert stats["max"] == 0.0

    @patch("pathlib.Path.exists")
    def test_create_report_path(self, mock_exists):
        """Test create_report_path."""
        base_path = Path("/tmp")
        report_path = create_report_path(base_path, suffix="test")
        assert "test_" in str(report_path)
        assert str(report_path).endswith(".md")
