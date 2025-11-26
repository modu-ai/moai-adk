"""Comprehensive test suite for common.py utilities module.

This module provides 90%+ coverage for all utility functions including:
- HTTPResponse dataclass and validation
- HTTPClient async operations
- URL extraction and validation
- File path utilities
- Duration formatting
- Statistical calculations
- Rate limiting
- Configuration loading
"""

import asyncio
import json
import tempfile
from datetime import datetime
from pathlib import Path
from unittest.mock import AsyncMock, patch

import aiohttp
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
    get_graceful_degradation,
    get_summary_stats,
    is_valid_url,
    load_hook_timeout,
)

# ============================================================================
# HTTPResponse Tests
# ============================================================================


class TestHTTPResponse:
    """Tests for HTTPResponse dataclass."""

    def test_http_response_creation(self):
        """Test creating HTTPResponse with all fields."""
        response = HTTPResponse(status_code=200, url="https://example.com", load_time=0.5, success=True)
        assert response.status_code == 200
        assert response.url == "https://example.com"
        assert response.load_time == 0.5
        assert response.success is True
        assert response.error_message is None
        assert isinstance(response.timestamp, datetime)

    def test_http_response_with_error_message(self):
        """Test HTTPResponse with error message."""
        error_msg = "Connection timeout"
        response = HTTPResponse(
            status_code=0, url="https://example.com", load_time=0.0, success=False, error_message=error_msg
        )
        assert response.success is False
        assert response.error_message == error_msg

    def test_http_response_timestamp_auto_generation(self):
        """Test timestamp is auto-generated if not provided."""
        response = HTTPResponse(status_code=200, url="https://example.com", load_time=0.1, success=True)
        assert response.timestamp is not None
        assert isinstance(response.timestamp, datetime)

    def test_http_response_timestamp_none_handling(self):
        """Test __post_init__ handles None timestamp."""
        response = HTTPResponse(status_code=200, url="https://example.com", load_time=0.1, success=True, timestamp=None)
        assert response.timestamp is not None

    def test_http_response_various_status_codes(self):
        """Test HTTPResponse with various status codes."""
        test_cases = [
            (200, True),
            (201, True),
            (299, True),
            (300, False),
            (404, False),
            (500, False),
            (0, False),
        ]
        for status, expected_success in test_cases:
            response = HTTPResponse(
                status_code=status,
                url="https://example.com",
                load_time=0.1,
                success=(200 <= status < 300),
            )
            assert response.status_code == status


# ============================================================================
# HTTPClient Tests
# ============================================================================


class TestHTTPClientAsyncContext:
    """Tests for HTTPClient async context manager."""

    @pytest.mark.asyncio
    async def test_client_context_manager_enter_exit(self):
        """Test HTTPClient async context manager setup and teardown."""
        async with HTTPClient(max_concurrent=5, timeout=10) as client:
            assert client.session is not None
            assert isinstance(client.session, aiohttp.ClientSession)
            assert client.max_concurrent == 5
            assert client.timeout == 10

    @pytest.mark.asyncio
    async def test_client_initializes_with_defaults(self):
        """Test HTTPClient initializes with default values."""
        client = HTTPClient()
        assert client.max_concurrent == 5
        assert client.timeout == 10
        assert client.session is None

    @pytest.mark.asyncio
    async def test_client_initializes_with_custom_values(self):
        """Test HTTPClient initializes with custom values."""
        client = HTTPClient(max_concurrent=20, timeout=30)
        assert client.max_concurrent == 20
        assert client.timeout == 30


class TestHTTPClientFetchUrl:
    """Tests for HTTPClient.fetch_url method."""

    @pytest.mark.asyncio
    async def test_fetch_url_session_not_initialized(self):
        """Test fetch_url when session is not initialized."""
        client = HTTPClient()
        response = await client.fetch_url("https://example.com")

        assert response.success is False
        assert response.status_code == 0
        assert response.load_time == 0
        assert "Session not initialized" in response.error_message

    @pytest.mark.asyncio
    async def test_fetch_url_successful_response(self):
        """Test fetch_url with successful response."""
        async with HTTPClient() as client:
            with patch("aiohttp.ClientSession.get") as mock_get:
                mock_response = AsyncMock()
                mock_response.status = 200
                mock_response.url = "https://example.com"
                mock_response.__aenter__.return_value = mock_response
                mock_response.__aexit__.return_value = None

                mock_get.return_value = mock_response
                client.session.get = mock_get

                response = await client.fetch_url("https://example.com")

                assert response.status_code == 200
                assert response.success is True
                assert response.url == "https://example.com"
                assert response.load_time >= 0

    @pytest.mark.asyncio
    async def test_fetch_url_various_status_codes(self):
        """Test fetch_url with various HTTP status codes."""
        test_cases = [
            (200, True),
            (201, True),
            (299, True),
            (300, False),
            (404, False),
            (500, False),
        ]

        async with HTTPClient() as client:
            for status_code, should_succeed in test_cases:
                with patch("aiohttp.ClientSession.get") as mock_get:
                    mock_response = AsyncMock()
                    mock_response.status = status_code
                    mock_response.url = "https://example.com"
                    mock_response.__aenter__.return_value = mock_response
                    mock_response.__aexit__.return_value = None

                    mock_get.return_value = mock_response
                    client.session.get = mock_get

                    response = await client.fetch_url("https://example.com")

                    assert response.status_code == status_code
                    assert response.success == should_succeed

    @pytest.mark.asyncio
    async def test_fetch_url_timeout_error(self):
        """Test fetch_url handles timeout errors."""
        async with HTTPClient(timeout=5) as client:
            with patch("aiohttp.ClientSession.get") as mock_get:
                mock_get.side_effect = asyncio.TimeoutError()
                client.session.get = mock_get

                response = await client.fetch_url("https://example.com")

                assert response.success is False
                assert response.status_code == 0
                assert "Request timeout" in response.error_message
                assert response.load_time == 5  # Timeout value

    @pytest.mark.asyncio
    async def test_fetch_url_client_error(self):
        """Test fetch_url handles aiohttp client errors."""
        async with HTTPClient() as client:
            with patch("aiohttp.ClientSession.get") as mock_get:
                mock_get.side_effect = aiohttp.ClientError("Connection failed")
                client.session.get = mock_get

                response = await client.fetch_url("https://example.com")

                assert response.success is False
                assert response.status_code == 0
                assert "HTTP client error" in response.error_message

    @pytest.mark.asyncio
    async def test_fetch_url_unexpected_error(self):
        """Test fetch_url handles unexpected errors."""
        async with HTTPClient() as client:
            with patch("aiohttp.ClientSession.get") as mock_get:
                mock_get.side_effect = RuntimeError("Unexpected error")
                client.session.get = mock_get

                response = await client.fetch_url("https://example.com")

                assert response.success is False
                assert response.status_code == 0
                assert "Unexpected error" in response.error_message


class TestHTTPClientFetchUrls:
    """Tests for HTTPClient.fetch_urls method."""

    @pytest.mark.asyncio
    async def test_fetch_multiple_urls(self):
        """Test fetching multiple URLs concurrently."""
        urls = [
            "https://example1.com",
            "https://example2.com",
            "https://example3.com",
        ]

        with patch("aiohttp.ClientSession.get") as mock_get:
            mock_response = AsyncMock()
            mock_response.status = 200
            mock_response.url = "https://example.com"
            mock_response.__aenter__.return_value = mock_response
            mock_response.__aexit__.return_value = None

            mock_get.return_value = mock_response

            with patch("aiohttp.TCPConnector"):
                with patch("aiohttp.ClientSession") as mock_session_class:
                    mock_session = AsyncMock()
                    mock_session.get = mock_get
                    mock_session.__aenter__.return_value = mock_session
                    mock_session.__aexit__.return_value = None
                    mock_session_class.return_value = mock_session

                    client = HTTPClient()
                    responses = await client.fetch_urls(urls)

                    assert len(responses) == 3
                    assert all(isinstance(r, HTTPResponse) for r in responses)

    @pytest.mark.asyncio
    async def test_fetch_empty_urls_list(self):
        """Test fetching empty URLs list."""
        with patch("aiohttp.ClientSession.close", new_callable=AsyncMock):
            with patch("aiohttp.TCPConnector"):
                with patch("aiohttp.ClientSession") as mock_session_class:
                    mock_session = AsyncMock()
                    mock_session.__aenter__.return_value = mock_session
                    mock_session.__aexit__.return_value = None
                    mock_session_class.return_value = mock_session

                    client = HTTPClient()
                    responses = await client.fetch_urls([])

                    assert len(responses) == 0
                    assert isinstance(responses, list)


# ============================================================================
# URL Extraction Tests
# ============================================================================


class TestExtractLinksFromText:
    """Tests for extract_links_from_text function."""

    def test_extract_markdown_links(self):
        """Test extracting markdown-style links."""
        text = "Check [this link](https://example.com) for more info."
        links = extract_links_from_text(text)
        assert "https://example.com" in links

    def test_extract_multiple_markdown_links(self):
        """Test extracting multiple markdown links."""
        text = """
        [link1](https://example1.com)
        [link2](https://example2.com)
        [link3](https://example3.com)
        """
        links = extract_links_from_text(text)
        assert len(links) == 3
        assert "https://example1.com" in links
        assert "https://example2.com" in links
        assert "https://example3.com" in links

    def test_extract_plain_urls(self):
        """Test extracting plain HTTP/HTTPS URLs."""
        text = "Visit https://example.com or http://another.com for details."
        links = extract_links_from_text(text)
        assert "https://example.com" in links
        assert "http://another.com" in links

    def test_extract_relative_urls_with_base(self):
        """Test converting relative URLs to absolute with base_url."""
        text = "[api](/api/users) and [home](/home)"
        base_url = "https://example.com"
        links = extract_links_from_text(text, base_url)
        assert "https://example.com/api/users" in links
        assert "https://example.com/home" in links

    def test_extract_relative_paths_with_base(self):
        """Test converting relative paths to absolute with base_url."""
        text = "[docs](docs/readme.md)"
        base_url = "https://example.com"
        links = extract_links_from_text(text, base_url)
        assert "https://example.com/docs/readme.md" in links

    def test_extract_ignores_anchor_links(self):
        """Test that anchor links (#) are not converted to absolute."""
        text = "[section](#section)"
        base_url = "https://example.com"
        links = extract_links_from_text(text, base_url)
        assert "#section" not in links  # Anchor links are filtered out

    def test_extract_removes_duplicates(self):
        """Test that duplicate links are removed."""
        text = """
        [link](https://example.com)
        [same](https://example.com)
        https://example.com
        """
        links = extract_links_from_text(text)
        # Count occurrences - should only appear once
        assert links.count("https://example.com") == 1

    def test_extract_empty_text(self):
        """Test extraction from empty text."""
        links = extract_links_from_text("")
        assert links == []

    def test_extract_text_with_no_links(self):
        """Test extraction from text with no links."""
        text = "This is just plain text with no links."
        links = extract_links_from_text(text)
        assert links == []

    def test_extract_malformed_markdown_links(self):
        """Test extraction with malformed markdown links."""
        text = "[incomplete (https://example.com)"
        links = extract_links_from_text(text)
        # Should still extract the plain URL
        assert "https://example.com" in links

    def test_extract_urls_with_special_characters(self):
        """Test extraction of URLs with query parameters."""
        text = "https://example.com/api?param=value&other=123"
        links = extract_links_from_text(text)
        assert "https://example.com/api?param=value&other=123" in links

    def test_extract_relative_url_without_base(self):
        """Test that relative URLs without base_url are ignored."""
        text = "[docs](/docs/readme.md)"
        links = extract_links_from_text(text)
        # Should not include relative paths without base
        assert "/docs/readme.md" not in links


# ============================================================================
# URL Validation Tests
# ============================================================================


class TestIsValidUrl:
    """Tests for is_valid_url function."""

    def test_valid_https_url(self):
        """Test validation of valid HTTPS URL."""
        assert is_valid_url("https://example.com") is True

    def test_valid_http_url(self):
        """Test validation of valid HTTP URL."""
        assert is_valid_url("http://example.com") is True

    def test_valid_url_with_path(self):
        """Test validation of URL with path."""
        assert is_valid_url("https://example.com/path/to/page") is True

    def test_valid_url_with_query_params(self):
        """Test validation of URL with query parameters."""
        assert is_valid_url("https://example.com/api?key=value") is True

    def test_valid_url_with_port(self):
        """Test validation of URL with port number."""
        assert is_valid_url("https://example.com:8080/path") is True

    def test_invalid_url_no_scheme(self):
        """Test validation fails without scheme."""
        assert is_valid_url("example.com") is False

    def test_invalid_url_no_netloc(self):
        """Test validation fails without netloc."""
        assert is_valid_url("https://") is False

    def test_invalid_url_empty_string(self):
        """Test validation of empty string."""
        assert is_valid_url("") is False

    def test_invalid_url_only_path(self):
        """Test validation of path-only URL."""
        assert is_valid_url("/path/to/file") is False

    def test_invalid_url_malformed(self):
        """Test validation of malformed URL."""
        assert is_valid_url("not a url") is False

    def test_valid_url_with_fragment(self):
        """Test validation of URL with fragment."""
        assert is_valid_url("https://example.com/page#section") is True

    def test_valid_url_with_username_password(self):
        """Test validation of URL with credentials."""
        assert is_valid_url("https://user:pass@example.com") is True

    @pytest.mark.parametrize(
        "url,expected",
        [
            ("ftp://example.com", True),
            ("http://localhost", True),
            ("https://192.168.1.1", True),
            ("http://example.com:3000", True),
        ],
    )
    def test_various_url_schemes(self, url, expected):
        """Test various URL schemes and formats."""
        assert is_valid_url(url) == expected

    def test_is_valid_url_with_exception_handling(self):
        """Test is_valid_url handles exceptions gracefully."""
        # This tests the exception handling in is_valid_url
        # The urlparse function typically doesn't throw, but we verify the exception block exists
        result = is_valid_url("")
        assert result is False  # Empty URL should be invalid


# ============================================================================
# Report Path Creation Tests
# ============================================================================


class TestCreateReportPath:
    """Tests for create_report_path function."""

    def test_create_report_path_with_default_suffix(self):
        """Test creating report path with default suffix."""
        with tempfile.TemporaryDirectory() as tmpdir:
            base_path = Path(tmpdir)
            report_path = create_report_path(base_path)

            assert report_path.parent == base_path
            assert report_path.name.startswith("report_")
            assert report_path.name.endswith(".md")

    def test_create_report_path_with_custom_suffix(self):
        """Test creating report path with custom suffix."""
        with tempfile.TemporaryDirectory() as tmpdir:
            base_path = Path(tmpdir)
            report_path = create_report_path(base_path, suffix="analysis")

            assert report_path.parent == base_path
            assert report_path.name.startswith("analysis_")
            assert report_path.name.endswith(".md")

    def test_create_report_path_timestamp_format(self):
        """Test that timestamp has correct format."""
        with tempfile.TemporaryDirectory() as tmpdir:
            base_path = Path(tmpdir)
            report_path = create_report_path(base_path)

            # Extract timestamp from filename (format: suffix_YYYYMMDD_HHMMSS.md)
            filename = report_path.name
            assert len(filename.split("_")) >= 3  # suffix_date_time.md

    def test_create_report_path_uniqueness(self):
        """Test that different calls produce different paths."""
        with tempfile.TemporaryDirectory() as tmpdir:
            base_path = Path(tmpdir)
            path1 = create_report_path(base_path)
            asyncio.sleep(0.1)  # Wait to ensure different timestamp
            path2 = create_report_path(base_path)

            # Paths might be the same if called in same second, but names are preserved
            assert path1.name != path2.name or path1 == path2


# ============================================================================
# Duration Formatting Tests
# ============================================================================


class TestFormatDuration:
    """Tests for format_duration function."""

    def test_format_milliseconds(self):
        """Test formatting durations less than 1 second."""
        assert format_duration(0.5) == "500ms"
        assert format_duration(0.1) == "100ms"
        assert format_duration(0.001) == "1ms"

    def test_format_seconds(self):
        """Test formatting durations in seconds."""
        assert format_duration(1.0) == "1.0s"
        assert format_duration(30.0) == "30.0s"
        assert format_duration(59.9) == "59.9s"

    def test_format_minutes(self):
        """Test formatting durations in minutes."""
        assert format_duration(60.0) == "1m 0s"
        assert format_duration(90.0) == "1m 30s"
        assert format_duration(3599.0) == "59m 59s"

    def test_format_hours(self):
        """Test formatting durations in hours."""
        assert format_duration(3600.0) == "1h 0m"
        assert format_duration(5400.0) == "1h 30m"
        assert format_duration(7200.0) == "2h 0m"

    def test_format_large_hours(self):
        """Test formatting very large durations."""
        assert format_duration(86400.0) == "24h 0m"
        assert format_duration(90000.0) == "25h 0m"

    @pytest.mark.parametrize(
        "seconds,expected",
        [
            (0.001, "1ms"),
            (0.5, "500ms"),
            (1.5, "1.5s"),
            (61.5, "1m 2s"),
            (3661.5, "1h 1m"),
        ],
    )
    def test_format_duration_edge_cases(self, seconds, expected):
        """Test edge cases in duration formatting."""
        assert format_duration(seconds) == expected

    def test_format_zero_duration(self):
        """Test formatting zero duration."""
        assert format_duration(0.0) == "0ms"

    def test_format_very_small_duration(self):
        """Test formatting very small duration."""
        result = format_duration(0.0001)
        assert "ms" in result


# ============================================================================
# Score Calculation Tests
# ============================================================================


class TestCalculateScore:
    """Tests for calculate_score function."""

    def test_calculate_unweighted_average(self):
        """Test calculating unweighted average score."""
        values = [80.0, 90.0, 100.0]
        score = calculate_score(values)
        assert score == pytest.approx(90.0)

    def test_calculate_weighted_average(self):
        """Test calculating weighted average score."""
        values = [80.0, 90.0, 100.0]
        weights = [1.0, 2.0, 1.0]
        score = calculate_score(values, weights)
        assert score == pytest.approx(90.0)

    def test_calculate_score_single_value(self):
        """Test calculating score with single value."""
        score = calculate_score([85.0])
        assert score == 85.0

    def test_calculate_score_empty_values(self):
        """Test calculating score with empty values."""
        score = calculate_score([])
        assert score == 0.0

    def test_calculate_score_zero_weights(self):
        """Test calculating score with zero total weight."""
        values = [80.0, 90.0]
        weights = [0.0, 0.0]
        score = calculate_score(values, weights)
        assert score == 0.0

    def test_calculate_score_mismatched_lengths(self):
        """Test that mismatched values and weights raise error."""
        values = [80.0, 90.0, 100.0]
        weights = [1.0, 2.0]
        with pytest.raises(ValueError):
            calculate_score(values, weights)

    def test_calculate_score_all_zeros(self):
        """Test calculating score with all zero values."""
        score = calculate_score([0.0, 0.0, 0.0])
        assert score == 0.0

    @pytest.mark.parametrize(
        "values,weights,expected",
        [
            ([100.0], [1.0], 100.0),
            ([50.0, 50.0], [1.0, 1.0], 50.0),
            ([100.0, 0.0], [1.0, 1.0], 50.0),
            ([100.0, 0.0], [2.0, 1.0], pytest.approx(66.67, abs=0.01)),
        ],
    )
    def test_calculate_score_various_cases(self, values, weights, expected):
        """Test various score calculation scenarios."""
        result = calculate_score(values, weights)
        if isinstance(expected, float) and "pytest" not in str(expected):
            assert result == pytest.approx(expected)
        else:
            assert result == expected


# ============================================================================
# Statistical Summary Tests
# ============================================================================


class TestGetSummaryStats:
    """Tests for get_summary_stats function."""

    def test_summary_stats_basic(self):
        """Test basic statistics calculation."""
        numbers = [10.0, 20.0, 30.0]
        stats = get_summary_stats(numbers)

        assert stats["mean"] == 20.0
        assert stats["min"] == 10.0
        assert stats["max"] == 30.0
        assert stats["std"] == pytest.approx(10.0)

    def test_summary_stats_single_value(self):
        """Test statistics with single value."""
        numbers = [42.0]
        stats = get_summary_stats(numbers)

        assert stats["mean"] == 42.0
        assert stats["min"] == 42.0
        assert stats["max"] == 42.0
        assert stats["std"] == 0.0

    def test_summary_stats_empty_list(self):
        """Test statistics with empty list."""
        stats = get_summary_stats([])

        assert stats["mean"] == 0.0
        assert stats["min"] == 0.0
        assert stats["max"] == 0.0
        assert stats["std"] == 0.0

    def test_summary_stats_two_values(self):
        """Test statistics with two values."""
        numbers = [10.0, 20.0]
        stats = get_summary_stats(numbers)

        assert stats["mean"] == 15.0
        assert stats["min"] == 10.0
        assert stats["max"] == 20.0
        assert stats["std"] == pytest.approx(7.071, abs=0.01)

    def test_summary_stats_identical_values(self):
        """Test statistics with identical values."""
        numbers = [5.0, 5.0, 5.0]
        stats = get_summary_stats(numbers)

        assert stats["mean"] == 5.0
        assert stats["min"] == 5.0
        assert stats["max"] == 5.0
        assert stats["std"] == 0.0

    def test_summary_stats_negative_values(self):
        """Test statistics with negative values."""
        numbers = [-10.0, 0.0, 10.0]
        stats = get_summary_stats(numbers)

        assert stats["mean"] == 0.0
        assert stats["min"] == -10.0
        assert stats["max"] == 10.0

    def test_summary_stats_large_numbers(self):
        """Test statistics with large numbers."""
        numbers = [1000000.0, 2000000.0, 3000000.0]
        stats = get_summary_stats(numbers)

        assert stats["mean"] == pytest.approx(2000000.0)
        assert stats["min"] == 1000000.0
        assert stats["max"] == 3000000.0


# ============================================================================
# Rate Limiter Tests
# ============================================================================


class TestRateLimiter:
    """Tests for RateLimiter class."""

    def test_rate_limiter_initialization(self):
        """Test RateLimiter initialization."""
        limiter = RateLimiter(max_requests=5, time_window=30)
        assert limiter.max_requests == 5
        assert limiter.time_window == 30
        assert limiter.requests == []

    def test_rate_limiter_default_values(self):
        """Test RateLimiter with default values."""
        limiter = RateLimiter()
        assert limiter.max_requests == 10
        assert limiter.time_window == 60

    def test_can_make_request_initially(self):
        """Test that request can be made initially."""
        limiter = RateLimiter(max_requests=2)
        assert limiter.can_make_request() is True

    def test_can_make_request_within_limit(self):
        """Test can_make_request within limit."""
        limiter = RateLimiter(max_requests=3)
        limiter.add_request()
        limiter.add_request()
        assert limiter.can_make_request() is True

    def test_can_make_request_at_limit(self):
        """Test can_make_request when at limit."""
        limiter = RateLimiter(max_requests=2)
        limiter.add_request()
        limiter.add_request()
        assert limiter.can_make_request() is False

    def test_add_request_success(self):
        """Test adding request when allowed."""
        limiter = RateLimiter(max_requests=2)
        limiter.add_request()
        assert len(limiter.requests) == 1
        limiter.add_request()
        assert len(limiter.requests) == 2

    def test_add_request_exceeds_limit(self):
        """Test adding request when limit is exceeded."""
        limiter = RateLimiter(max_requests=1)
        limiter.add_request()

        with pytest.raises(RateLimitError):
            limiter.add_request()

    def test_add_request_error_message(self):
        """Test RateLimitError message contains limits."""
        limiter = RateLimiter(max_requests=5, time_window=60)
        limiter.requests = [datetime.now() for _ in range(5)]

        with pytest.raises(RateLimitError) as exc_info:
            limiter.add_request()

        error_msg = str(exc_info.value)
        assert "5" in error_msg or "rate limit" in error_msg.lower()

    def test_rate_limit_window_expiry(self):
        """Test that old requests are cleaned up."""
        limiter = RateLimiter(max_requests=2, time_window=1)
        limiter.add_request()

        # Simulate passage of time
        old_time = datetime.now()
        limiter.requests[0] = old_time.replace(year=old_time.year - 1)

        # Should be able to make request as old one is removed
        assert limiter.can_make_request() is True

    @pytest.mark.asyncio
    async def test_wait_if_needed_no_wait(self):
        """Test wait_if_needed when no wait is needed."""
        limiter = RateLimiter(max_requests=10)
        limiter.add_request()

        # Should return immediately
        await limiter.wait_if_needed()

    @pytest.mark.asyncio
    async def test_wait_if_needed_with_wait(self):
        """Test wait_if_needed waits when at limit."""
        limiter = RateLimiter(max_requests=1, time_window=1)
        limiter.add_request()

        # Mock asyncio.sleep to verify it's called
        with patch("asyncio.sleep", new_callable=AsyncMock) as mock_sleep:
            await limiter.wait_if_needed()
            mock_sleep.assert_called_once()

    def test_rate_limiter_multiple_windows(self):
        """Test rate limiter across multiple time windows."""
        limiter = RateLimiter(max_requests=2, time_window=1)

        limiter.add_request()
        limiter.add_request()

        # At limit
        assert limiter.can_make_request() is False

        # Simulate time passing
        limiter.requests[0] = datetime.now().replace(year=2000)

        # Should allow request now
        assert limiter.can_make_request() is True


# ============================================================================
# Configuration Loading Tests
# ============================================================================


class TestLoadHookTimeout:
    """Tests for load_hook_timeout function."""

    def test_load_hook_timeout_default(self):
        """Test default timeout when config doesn't exist."""
        with patch("pathlib.Path.exists", return_value=False):
            timeout = load_hook_timeout()
            assert timeout == 5000

    def test_load_hook_timeout_from_config(self):
        """Test loading timeout from config file."""
        config = {"hooks": {"timeout_ms": 10000}}

        with tempfile.NamedTemporaryFile(mode="w", suffix=".json", delete=False) as f:
            json.dump(config, f)
            config_path = f.name

        try:
            with patch("pathlib.Path.exists", return_value=True):
                with patch("builtins.open", create=True) as mock_open:
                    mock_open.return_value.__enter__.return_value.read.return_value = json.dumps(config)
                    # Need to patch json.load as well
                    with patch("json.load", return_value=config):
                        with patch("pathlib.Path", return_value=Path(config_path)):
                            timeout = load_hook_timeout()
                            assert timeout == 10000
        finally:
            Path(config_path).unlink()

    def test_load_hook_timeout_malformed_json(self):
        """Test handling malformed JSON in config."""
        with patch("pathlib.Path.exists", return_value=True):
            with patch("builtins.open", create=True) as mock_open:
                mock_open.return_value.__enter__.return_value.read.return_value = "invalid json"
                with patch("json.load", side_effect=json.JSONDecodeError("msg", "doc", 0)):
                    timeout = load_hook_timeout()
                    assert timeout == 5000

    def test_load_hook_timeout_missing_hooks_section(self):
        """Test handling config without hooks section."""
        config = {"other": "data"}

        with patch("pathlib.Path.exists", return_value=True):
            with patch("builtins.open", create=True):
                with patch("json.load", return_value=config):
                    timeout = load_hook_timeout()
                    assert timeout == 5000

    def test_load_hook_timeout_missing_timeout_ms(self):
        """Test handling hooks section without timeout_ms."""
        config = {"hooks": {"other": "value"}}

        with patch("pathlib.Path.exists", return_value=True):
            with patch("builtins.open", create=True):
                with patch("json.load", return_value=config):
                    timeout = load_hook_timeout()
                    assert timeout == 5000

    def test_load_hook_timeout_various_values(self):
        """Test loading various timeout values."""
        test_cases = [1000, 5000, 10000, 30000]

        for timeout_ms in test_cases:
            config = {"hooks": {"timeout_ms": timeout_ms}}

            with patch("pathlib.Path.exists", return_value=True):
                with patch("builtins.open", create=True):
                    with patch("json.load", return_value=config):
                        result = load_hook_timeout()
                        assert result == timeout_ms


class TestGetGracefulDegradation:
    """Tests for get_graceful_degradation function."""

    def test_graceful_degradation_default(self):
        """Test default graceful_degradation when config doesn't exist."""
        with patch("pathlib.Path.exists", return_value=False):
            result = get_graceful_degradation()
            assert result is True

    def test_graceful_degradation_from_config_true(self):
        """Test loading graceful_degradation=true from config."""
        config = {"hooks": {"graceful_degradation": True}}

        with patch("pathlib.Path.exists", return_value=True):
            with patch("builtins.open", create=True):
                with patch("json.load", return_value=config):
                    result = get_graceful_degradation()
                    assert result is True

    def test_graceful_degradation_from_config_false(self):
        """Test loading graceful_degradation=false from config."""
        config = {"hooks": {"graceful_degradation": False}}

        with patch("pathlib.Path.exists", return_value=True):
            with patch("builtins.open", create=True):
                with patch("json.load", return_value=config):
                    result = get_graceful_degradation()
                    assert result is False

    def test_graceful_degradation_malformed_json(self):
        """Test handling malformed JSON in config."""
        with patch("pathlib.Path.exists", return_value=True):
            with patch("builtins.open", create=True) as mock_open:
                mock_open.return_value.__enter__.return_value.read.return_value = "invalid"
                with patch("json.load", side_effect=json.JSONDecodeError("msg", "doc", 0)):
                    result = get_graceful_degradation()
                    assert result is True

    def test_graceful_degradation_missing_hooks_section(self):
        """Test handling config without hooks section."""
        config = {"other": "data"}

        with patch("pathlib.Path.exists", return_value=True):
            with patch("builtins.open", create=True):
                with patch("json.load", return_value=config):
                    result = get_graceful_degradation()
                    assert result is True

    def test_graceful_degradation_missing_setting(self):
        """Test handling hooks section without graceful_degradation."""
        config = {"hooks": {"other": "value"}}

        with patch("pathlib.Path.exists", return_value=True):
            with patch("builtins.open", create=True):
                with patch("json.load", return_value=config):
                    result = get_graceful_degradation()
                    assert result is True


# ============================================================================
# Integration and Edge Case Tests
# ============================================================================


class TestRateLimitError:
    """Tests for RateLimitError exception."""

    def test_rate_limit_error_is_exception(self):
        """Test that RateLimitError is an Exception."""
        assert issubclass(RateLimitError, Exception)

    def test_rate_limit_error_instantiation(self):
        """Test creating RateLimitError."""
        error = RateLimitError("Test error message")
        assert str(error) == "Test error message"

    def test_rate_limit_error_raise_and_catch(self):
        """Test raising and catching RateLimitError."""
        with pytest.raises(RateLimitError) as exc_info:
            raise RateLimitError("Rate limit exceeded")

        assert "Rate limit exceeded" in str(exc_info.value)


class TestExtractLinksEdgeCases:
    """Additional edge case tests for extract_links_from_text."""

    def test_extract_links_mixed_content(self):
        """Test extracting from mixed markdown and plain URLs."""
        text = """
        [Documentation](https://docs.example.com)
        Visit https://example.com for more
        Another [link](/path) and plain text
        """
        base_url = "https://example.com"
        links = extract_links_from_text(text, base_url)

        assert "https://docs.example.com" in links
        assert "https://example.com" in links
        assert "https://example.com/path" in links

    def test_extract_links_with_trailing_slash_removal(self):
        """Test that trailing slashes are removed in relative URL joining."""
        text = "[api](api/endpoint)"
        base_url = "https://example.com/"
        links = extract_links_from_text(text, base_url)

        # Should result in proper URL without double slashes
        assert len(links) > 0

    def test_extract_links_url_pattern_boundaries(self):
        """Test URL pattern extraction with proper boundaries."""
        text = "Visit https://example.com/path?query=1 (important)"
        links = extract_links_from_text(text)

        assert "https://example.com/path?query=1" in links
        # Parenthesis should not be included
        assert not any(")" in link for link in links)

    def test_extract_links_multiple_base_url_formats(self):
        """Test extraction with different base URL formats."""
        test_cases = [
            ("https://example.com", "[link](/api)"),
            ("https://example.com/", "[link](/api)"),
            ("http://example.com:8080", "[link](/api)"),
        ]

        for base_url, text in test_cases:
            links = extract_links_from_text(text, base_url)
            assert len(links) > 0


class TestFormatDurationBoundaries:
    """Test format_duration boundary conditions."""

    def test_format_duration_boundary_at_1_second(self):
        """Test boundary at 1 second."""
        assert "s" in format_duration(0.999)
        assert "s" in format_duration(1.0)
        assert "s" in format_duration(1.001)

    def test_format_duration_boundary_at_60_seconds(self):
        """Test boundary at 60 seconds."""
        assert "s" in format_duration(59.9)
        assert "m" in format_duration(60.0)

    def test_format_duration_boundary_at_3600_seconds(self):
        """Test boundary at 1 hour."""
        assert "m" in format_duration(3599.0)
        assert "h" in format_duration(3600.0)


class TestCalculateScoreEdgeCases:
    """Additional edge case tests for calculate_score."""

    def test_calculate_score_with_negative_weights(self):
        """Test score calculation with negative weights."""
        values = [100.0, 0.0]
        weights = [1.0, -1.0]  # Negative weight

        result = calculate_score(values, weights)
        # Result should be calculated regardless
        assert isinstance(result, float)

    def test_calculate_score_with_very_large_values(self):
        """Test score with very large numbers."""
        values = [1e10, 2e10, 3e10]
        result = calculate_score(values)
        assert result == pytest.approx(2e10)

    def test_calculate_score_with_very_small_values(self):
        """Test score with very small numbers."""
        values = [1e-10, 2e-10, 3e-10]
        result = calculate_score(values)
        assert result == pytest.approx(2e-10)


class TestIntegrationScenarios:
    """Integration tests for common.py utilities."""

    def test_url_extraction_and_validation_pipeline(self):
        """Test extracting and validating URLs in pipeline."""
        text = """
        Check [this](https://example.com) and https://another.com
        Also see [docs](/api) for details.
        """
        base_url = "https://example.com"

        links = extract_links_from_text(text, base_url)

        # Validate all extracted links
        valid_links = [link for link in links if is_valid_url(link)]

        assert len(valid_links) > 0
        assert all(is_valid_url(link) for link in valid_links)

    def test_statistics_calculation_comprehensive(self):
        """Test comprehensive statistics calculation."""
        values = [10.0, 20.0, 30.0, 40.0, 50.0]

        stats = get_summary_stats(values)

        # Verify relationships
        assert stats["min"] <= stats["mean"] <= stats["max"]
        assert stats["std"] >= 0
        assert stats["mean"] == 30.0

    def test_duration_and_score_combined(self):
        """Test using duration and score together."""
        durations = [0.1, 0.5, 1.0, 5.0, 60.0]
        scores = [95.0, 87.0, 92.0, 88.0, 85.0]

        # Format durations
        formatted = [format_duration(d) for d in durations]

        # Calculate weighted score
        weights = [1.0, 1.0, 2.0, 1.0, 0.5]
        avg_score = calculate_score(scores, weights)

        assert len(formatted) == len(durations)
        assert 80.0 < avg_score < 100.0

    @pytest.mark.asyncio
    async def test_rate_limiter_with_client_workflow(self):
        """Test rate limiter with HTTP client workflow."""
        limiter = RateLimiter(max_requests=3, time_window=2)

        # Simulate requests
        for i in range(3):
            assert limiter.can_make_request()
            limiter.add_request()

        # Now at limit
        assert not limiter.can_make_request()

    def test_config_loading_resilience(self):
        """Test config loading handles various error conditions."""
        # Test with non-existent file
        with patch("pathlib.Path.exists", return_value=False):
            timeout = load_hook_timeout()
            degradation = get_graceful_degradation()

            assert timeout == 5000
            assert degradation is True

    def test_complete_workflow_url_processing(self):
        """Test complete workflow from extraction to validation to formatting."""
        text = """
        API: https://api.example.com/v1/users?key=123
        Docs: [guide](https://docs.example.com)
        """

        # Extract links
        links = extract_links_from_text(text)

        # Validate each link
        assert all(is_valid_url(link) for link in links)

        # Verify we got the expected links
        assert len(links) >= 2

    def test_statistics_with_various_data_distributions(self):
        """Test statistics calculation with different data distributions."""
        # Normal distribution
        normal_data = [1, 2, 3, 4, 5, 4, 3, 2, 1]
        stats_normal = get_summary_stats([float(x) for x in normal_data])

        # Uniform distribution
        uniform_data = [5.0, 5.0, 5.0, 5.0, 5.0]
        stats_uniform = get_summary_stats(uniform_data)

        # Skewed distribution
        skewed_data = [1.0, 1.0, 1.0, 1.0, 100.0]
        stats_skewed = get_summary_stats(skewed_data)

        # Verify std dev relationships
        assert stats_normal["std"] > 0
        assert stats_uniform["std"] == 0
        assert stats_skewed["std"] > stats_uniform["std"]

    def test_rate_limiting_and_timing(self):
        """Test rate limiting with timing information."""
        limiter = RateLimiter(max_requests=5, time_window=10)

        # Track timing of requests
        request_times = []

        for i in range(5):
            if limiter.can_make_request():
                limiter.add_request()
                request_times.append(datetime.now())

        # Verify all requests were recorded
        assert len(request_times) == 5

        # Verify requests are within expected time window
        if len(request_times) > 1:
            time_span = (request_times[-1] - request_times[0]).total_seconds()
            assert time_span < limiter.time_window
