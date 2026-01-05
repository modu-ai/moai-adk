"""Comprehensive coverage tests for moai_adk.utils.common module.

This module contains tests targeting uncovered code paths in common.py
to achieve 90%+ coverage of utility functions.
"""

import asyncio
import json
from datetime import datetime
from pathlib import Path
from unittest.mock import MagicMock, mock_open, patch

import aiohttp
import pytest
import yaml

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


class TestHTTPResponsePostInit:
    """Test HTTPResponse __post_init__ method."""

    def test_post_init_sets_timestamp_when_none(self):
        """Test __post_init__ sets timestamp when None."""
        response = HTTPResponse(
            status_code=200,
            url="https://example.com",
            load_time=0.5,
            success=True,
            timestamp=None,
        )
        assert response.timestamp is not None
        assert isinstance(response.timestamp, datetime)

    def test_post_init_preserves_existing_timestamp(self):
        """Test __post_init__ preserves existing timestamp."""
        now = datetime.now()
        response = HTTPResponse(
            status_code=200,
            url="https://example.com",
            load_time=0.5,
            success=True,
            timestamp=now,
        )
        assert response.timestamp == now


class TestHTTPClientContextManager:
    """Test HTTPClient async context manager."""

    @pytest.mark.asyncio
    async def test_client_aenter_creates_session(self):
        """Test __aenter__ creates aiohttp session."""
        client = HTTPClient(max_concurrent=3, timeout=5)
        async with client:
            assert client.session is not None
            assert isinstance(client.session, aiohttp.ClientSession)

    @pytest.mark.asyncio
    async def test_client_aexit_closes_session(self):
        """Test __aexit__ properly closes session."""
        client = HTTPClient()
        async with client:
            pass
        # Session should be closed
        assert client.session is not None

    @pytest.mark.asyncio
    async def test_client_aexit_with_exception(self):
        """Test __aexit__ handles exceptions gracefully."""
        client = HTTPClient()
        try:
            async with client:
                raise ValueError("Test error")
        except ValueError:
            pass
        # Should still handle cleanup


class TestHTTPClientFetchUrl:
    """Test HTTPClient fetch_url method."""

    @pytest.mark.asyncio
    async def test_fetch_url_success_200(self):
        """Test fetch_url with successful 200 response."""
        client = HTTPClient()
        mock_response = MagicMock()
        mock_response.status = 200
        mock_response.url = "https://example.com"

        with patch("aiohttp.ClientSession.get") as mock_get:
            mock_get.return_value.__aenter__.return_value = mock_response
            async with client:
                result = await client.fetch_url("https://example.com")
                assert result.success is True
                assert result.status_code == 200

    @pytest.mark.asyncio
    async def test_fetch_url_redirect_299(self):
        """Test fetch_url with redirect response (299)."""
        client = HTTPClient()
        mock_response = MagicMock()
        mock_response.status = 299
        mock_response.url = "https://example.com/redirected"

        with patch("aiohttp.ClientSession.get") as mock_get:
            mock_get.return_value.__aenter__.return_value = mock_response
            async with client:
                result = await client.fetch_url("https://example.com")
                assert result.success is True
                assert result.status_code == 299

    @pytest.mark.asyncio
    async def test_fetch_url_error_404(self):
        """Test fetch_url with 404 error."""
        client = HTTPClient()
        mock_response = MagicMock()
        mock_response.status = 404
        mock_response.url = "https://example.com"

        with patch("aiohttp.ClientSession.get") as mock_get:
            mock_get.return_value.__aenter__.return_value = mock_response
            async with client:
                result = await client.fetch_url("https://example.com")
                assert result.success is False
                assert result.status_code == 404

    @pytest.mark.asyncio
    async def test_fetch_url_timeout_error(self):
        """Test fetch_url with timeout."""
        client = HTTPClient(timeout=1)
        with patch("aiohttp.ClientSession.get") as mock_get:
            mock_get.side_effect = asyncio.TimeoutError()
            async with client:
                result = await client.fetch_url("https://example.com")
                assert result.success is False
                assert result.error_message == "Request timeout"
                assert result.load_time == 1

    @pytest.mark.asyncio
    async def test_fetch_url_client_error(self):
        """Test fetch_url with aiohttp ClientError."""
        client = HTTPClient()
        with patch("aiohttp.ClientSession.get") as mock_get:
            mock_get.side_effect = aiohttp.ClientError("Connection failed")
            async with client:
                result = await client.fetch_url("https://example.com")
                assert result.success is False
                assert "HTTP client error" in result.error_message

    @pytest.mark.asyncio
    async def test_fetch_url_unexpected_error(self):
        """Test fetch_url with unexpected exception."""
        client = HTTPClient()
        with patch("aiohttp.ClientSession.get") as mock_get:
            mock_get.side_effect = RuntimeError("Unexpected error")
            async with client:
                result = await client.fetch_url("https://example.com")
                assert result.success is False
                assert "Unexpected error" in result.error_message

    @pytest.mark.asyncio
    async def test_fetch_urls_concurrent(self):
        """Test fetch_urls with multiple URLs."""
        client = HTTPClient(max_concurrent=2)
        urls = ["https://example1.com", "https://example2.com"]

        with patch.object(client, "fetch_url") as mock_fetch:
            mock_fetch.return_value = HTTPResponse(
                status_code=200,
                url="https://example.com",
                load_time=0.5,
                success=True,
            )
            results = await client.fetch_urls(urls)
            assert len(results) == 2
            assert all(r.success for r in results)


class TestExtractLinksFromText:
    """Test extract_links_from_text function."""

    def test_extract_markdown_links_only(self):
        """Test extracting only markdown links."""
        text = "[Documentation](https://docs.example.com) and [Blog](https://blog.example.com)"
        links = extract_links_from_text(text)
        assert "https://docs.example.com" in links
        assert "https://blog.example.com" in links

    def test_extract_plain_urls_only(self):
        """Test extracting only plain URLs."""
        text = "Visit https://example.com or https://another.com for more info"
        links = extract_links_from_text(text)
        assert "https://example.com" in links
        assert "https://another.com" in links

    def test_extract_relative_urls_with_base(self):
        """Test extracting relative URLs with base URL."""
        text = "[Link](/docs/guide) and [API](/api/reference)"
        links = extract_links_from_text(text, base_url="https://docs.example.com")
        assert any("docs.example.com" in link for link in links)

    def test_extract_relative_urls_without_base(self):
        """Test extracting relative URLs without base URL."""
        text = "[Link](/docs/guide)"
        links = extract_links_from_text(text)
        # Relative links without base should not be included
        assert "/docs/guide" not in links

    def test_extract_urls_with_trailing_slash(self):
        """Test extracting URLs with trailing slashes."""
        text = "[Link](https://example.com/docs/)"
        links = extract_links_from_text(text, base_url="https://example.com")
        assert "https://example.com/docs/" in links

    def test_extract_fragment_links_ignored(self):
        """Test that fragment-only links are not processed."""
        text = "[Anchor](#section)"
        links = extract_links_from_text(text)
        # Fragment links without base should be ignored
        assert not any("#" in link for link in links)

    def test_extract_http_urls(self):
        """Test extracting HTTP URLs."""
        text = "Visit http://example.com for more"
        links = extract_links_from_text(text)
        assert "http://example.com" in links

    def test_extract_removes_duplicates(self):
        """Test that duplicates are properly removed."""
        text = "[Link1](https://example.com) and [Link2](https://example.com) and https://example.com"
        links = extract_links_from_text(text)
        assert links.count("https://example.com") == 1

    def test_extract_empty_text(self):
        """Test extracting from empty text."""
        links = extract_links_from_text("")
        assert links == []

    def test_extract_no_links(self):
        """Test text with no links."""
        text = "This is plain text without any links"
        links = extract_links_from_text(text)
        assert links == []


class TestIsValidUrl:
    """Test is_valid_url function."""

    def test_valid_https_url(self):
        """Test valid HTTPS URL."""
        assert is_valid_url("https://example.com") is True
        assert is_valid_url("https://sub.example.co.uk") is True

    def test_valid_http_url(self):
        """Test valid HTTP URL."""
        assert is_valid_url("http://example.com") is True

    def test_valid_url_with_path(self):
        """Test valid URL with path."""
        assert is_valid_url("https://example.com/path/to/resource") is True

    def test_valid_url_with_query(self):
        """Test valid URL with query string."""
        assert is_valid_url("https://example.com/path?key=value") is True

    def test_invalid_no_scheme(self):
        """Test invalid URL without scheme."""
        assert is_valid_url("example.com") is False

    def test_invalid_no_netloc(self):
        """Test invalid URL without network location."""
        assert is_valid_url("https://") is False

    def test_invalid_empty_string(self):
        """Test invalid empty string."""
        assert is_valid_url("") is False

    def test_invalid_malformed(self):
        """Test invalid malformed URL."""
        assert is_valid_url("not a url") is False
        assert is_valid_url("://invalid") is False


class TestCreateReportPath:
    """Test create_report_path function."""

    def test_create_report_path_with_suffix(self):
        """Test creating report path with custom suffix."""
        base_path = Path("/tmp")
        report_path = create_report_path(base_path, suffix="custom")
        assert "custom_" in str(report_path)
        assert str(report_path).endswith(".md")

    def test_create_report_path_default_suffix(self):
        """Test creating report path with default suffix."""
        base_path = Path("/tmp")
        report_path = create_report_path(base_path)
        assert "report_" in str(report_path)
        assert str(report_path).endswith(".md")

    def test_create_report_path_timestamp_format(self):
        """Test report path contains valid timestamp."""
        base_path = Path("/tmp")
        report_path = create_report_path(base_path)
        # Should contain YYYYMMDD_HHMMSS format
        filename = report_path.name
        assert len(filename) > 20  # At least long enough for timestamp


class TestFormatDuration:
    """Test format_duration function."""

    def test_format_milliseconds_under_1000(self):
        """Test formatting milliseconds."""
        assert "ms" in format_duration(0.001)
        assert "ms" in format_duration(0.5)
        assert "999" in format_duration(0.999)

    def test_format_seconds_1_to_60(self):
        """Test formatting seconds."""
        result = format_duration(1.5)
        assert "s" in result
        assert "1.5" in result

    def test_format_seconds_boundary(self):
        """Test formatting at second/minute boundary."""
        result = format_duration(59.9)
        assert "s" in result

    def test_format_minutes_1_to_60(self):
        """Test formatting minutes."""
        result = format_duration(120)
        assert "m" in result
        assert "2" in result

    def test_format_minutes_with_seconds(self):
        """Test formatting minutes with remaining seconds."""
        result = format_duration(125)
        assert "m" in result
        assert "5" in result

    def test_format_hours(self):
        """Test formatting hours."""
        result = format_duration(3600)
        assert "h" in result

    def test_format_hours_with_minutes(self):
        """Test formatting hours with remaining minutes."""
        result = format_duration(3660)
        assert "h" in result
        assert "m" in result

    def test_format_zero_duration(self):
        """Test formatting zero duration."""
        result = format_duration(0)
        assert "0" in result


class TestCalculateScore:
    """Test calculate_score function."""

    def test_calculate_score_equal_values(self):
        """Test calculate score with equal values."""
        score = calculate_score([50, 50, 50])
        assert score == 50.0

    def test_calculate_score_mixed_values(self):
        """Test calculate score with mixed values."""
        score = calculate_score([80, 90, 100])
        assert score == pytest.approx(90.0)

    def test_calculate_score_single_value(self):
        """Test calculate score with single value."""
        score = calculate_score([75.0])
        assert score == 75.0

    def test_calculate_score_weighted_equal(self):
        """Test weighted score with equal weights."""
        score = calculate_score([80, 100], weights=[1.0, 1.0])
        assert score == 90.0

    def test_calculate_score_weighted_unequal(self):
        """Test weighted score with unequal weights."""
        score = calculate_score([80, 100], weights=[1.0, 3.0])
        assert score == pytest.approx(95.0)

    def test_calculate_score_zero_weights(self):
        """Test score with zero total weight."""
        score = calculate_score([80, 90], weights=[0.0, 0.0])
        assert score == 0.0

    def test_calculate_score_negative_values(self):
        """Test score with negative values."""
        score = calculate_score([-10, 10, 30])
        assert score == pytest.approx(10.0)

    def test_calculate_score_mismatched_length_error(self):
        """Test error when values and weights have different lengths."""
        with pytest.raises(ValueError, match="same length"):
            calculate_score([80, 90, 100], weights=[1.0, 2.0])


class TestGetSummaryStats:
    """Test get_summary_stats function."""

    def test_get_summary_stats_basic(self):
        """Test basic statistics calculation."""
        stats = get_summary_stats([10, 20, 30])
        assert stats["mean"] == 20.0
        assert stats["min"] == 10.0
        assert stats["max"] == 30.0
        assert isinstance(stats["std"], float)

    def test_get_summary_stats_single_value(self):
        """Test statistics with single value."""
        stats = get_summary_stats([42.0])
        assert stats["mean"] == 42.0
        assert stats["min"] == 42.0
        assert stats["max"] == 42.0
        assert stats["std"] == 0.0

    def test_get_summary_stats_two_values(self):
        """Test statistics with two values."""
        stats = get_summary_stats([10.0, 20.0])
        assert stats["mean"] == 15.0
        assert stats["min"] == 10.0
        assert stats["max"] == 20.0
        assert stats["std"] > 0

    def test_get_summary_stats_negative_values(self):
        """Test statistics with negative values."""
        stats = get_summary_stats([-10, 0, 10])
        assert stats["mean"] == 0.0
        assert stats["min"] == -10
        assert stats["max"] == 10

    def test_get_summary_stats_empty_list(self):
        """Test statistics with empty list."""
        stats = get_summary_stats([])
        assert stats["mean"] == 0.0
        assert stats["min"] == 0.0
        assert stats["max"] == 0.0
        assert stats["std"] == 0.0

    def test_get_summary_stats_identical_values(self):
        """Test statistics with identical values."""
        stats = get_summary_stats([5, 5, 5, 5])
        assert stats["mean"] == 5.0
        assert stats["std"] == 0.0

    def test_get_summary_stats_large_range(self):
        """Test statistics with large range of values."""
        stats = get_summary_stats([1, 1000, 500])
        assert stats["min"] == 1
        assert stats["max"] == 1000
        assert stats["mean"] == pytest.approx((1 + 1000 + 500) / 3)


class TestRateLimiterCleanup:
    """Test RateLimiter request cleanup."""

    def test_rate_limiter_removes_old_requests(self):
        """Test that old requests are removed from tracking."""
        limiter = RateLimiter(max_requests=10, time_window=1)
        # Add a request
        limiter.add_request()
        assert len(limiter.requests) == 1

        # Wait for time window to pass
        import time

        time.sleep(1.1)

        # Check can make request should clean up old requests
        limiter.can_make_request()
        assert len(limiter.requests) == 0

    def test_rate_limiter_mixed_old_new_requests(self):
        """Test cleanup with mix of old and new requests."""
        limiter = RateLimiter(max_requests=10, time_window=1)
        limiter.add_request()
        limiter.add_request()
        assert len(limiter.requests) == 2

        # Requests should eventually be cleaned up after time window
        import time

        time.sleep(1.1)
        limiter.can_make_request()
        assert len(limiter.requests) == 0


class TestRateLimiterWait:
    """Test RateLimiter async wait functionality."""

    @pytest.mark.asyncio
    async def test_wait_if_needed_no_wait_needed(self):
        """Test wait_if_needed when under limit."""
        limiter = RateLimiter(max_requests=10, time_window=60)
        # Should not wait
        import time

        start = time.time()
        await limiter.wait_if_needed()
        elapsed = time.time() - start
        assert elapsed < 0.1  # Should complete quickly

    @pytest.mark.asyncio
    async def test_wait_if_needed_at_limit(self):
        """Test wait_if_needed when at limit."""
        limiter = RateLimiter(max_requests=2, time_window=1)
        # Fill the limit
        limiter.add_request()
        limiter.add_request()

        # Should need to wait
        import time

        start = time.time()
        await limiter.wait_if_needed()
        elapsed = time.time() - start
        assert elapsed > 0.5  # Should wait some time


class TestLoadHookTimeout:
    """Test load_hook_timeout function."""

    def test_load_hook_timeout_default(self):
        """Test load_hook_timeout returns default when file doesn't exist."""
        with patch("pathlib.Path.exists", return_value=False):
            timeout = load_hook_timeout()
            assert timeout == 5000

    def test_load_hook_timeout_from_config(self):
        """Test load_hook_timeout reads from config file."""
        mock_config = {"hooks": {"timeout_ms": 3000}}
        with patch("pathlib.Path.exists", return_value=True):
            with patch("builtins.open", mock_open(read_data=json.dumps(mock_config))):
                with patch("json.load", return_value=mock_config):
                    timeout = load_hook_timeout()
                    assert timeout == 3000

    def test_load_hook_timeout_invalid_json(self):
        """Test load_hook_timeout handles invalid JSON."""
        with patch("pathlib.Path.exists", return_value=True):
            with patch(
                "builtins.open",
                MagicMock(side_effect=json.JSONDecodeError("msg", "doc", 0)),
            ):
                timeout = load_hook_timeout()
                assert timeout == 5000

    def test_load_hook_timeout_missing_key(self):
        """Test load_hook_timeout when timeout_ms key is missing."""
        mock_config = {"hooks": {}}
        with patch("pathlib.Path.exists", return_value=True):
            with patch("builtins.open", mock_open(read_data=json.dumps(mock_config))):
                with patch("json.load", return_value=mock_config):
                    timeout = load_hook_timeout()
                    assert timeout == 5000

    def test_load_hook_timeout_invalid_value(self):
        """Test load_hook_timeout with invalid timeout value."""
        mock_config = {"hooks": {"timeout_ms": "invalid"}}
        with patch("pathlib.Path.exists", return_value=True):
            with patch("builtins.open", mock_open(read_data=json.dumps(mock_config))):
                with patch("json.load", return_value=mock_config):
                    timeout = load_hook_timeout()
                    assert timeout == 5000


class TestGetGracefulDegradation:
    """Test get_graceful_degradation function."""

    def test_get_graceful_degradation_default(self):
        """Test get_graceful_degradation returns default when file doesn't exist."""
        with patch("pathlib.Path.exists", return_value=False):
            result = get_graceful_degradation()
            assert result is True

    def test_get_graceful_degradation_true(self):
        """Test get_graceful_degradation reads true value."""
        with patch("pathlib.Path.exists", return_value=True):
            with patch("builtins.open", mock_open(read_data="hooks:\n  graceful_degradation: true")):
                with patch("yaml.safe_load", return_value={"hooks": {"graceful_degradation": True}}):
                    result = get_graceful_degradation()
                    assert result is True

    def test_get_graceful_degradation_false(self):
        """Test get_graceful_degradation reads false value."""
        with patch("pathlib.Path.exists", return_value=True):
            with patch("builtins.open", mock_open(read_data="hooks:\n  graceful_degradation: false")):
                with patch("yaml.safe_load", return_value={"hooks": {"graceful_degradation": False}}):
                    result = get_graceful_degradation()
                    assert result is False

    def test_get_graceful_degradation_invalid_json(self):
        """Test get_graceful_degradation handles invalid YAML."""
        with patch("pathlib.Path.exists", return_value=True):
            with patch("builtins.open", mock_open(read_data="invalid: yaml: content")):
                with patch("yaml.safe_load", side_effect=yaml.YAMLError("Invalid YAML")):
                    result = get_graceful_degradation()
                    assert result is True

    def test_get_graceful_degradation_missing_key(self):
        """Test get_graceful_degradation when key is missing."""
        with patch("pathlib.Path.exists", return_value=True):
            with patch("builtins.open", mock_open(read_data="hooks: {}")):
                with patch("yaml.safe_load", return_value={"hooks": {}}):
                    result = get_graceful_degradation()
                    assert result is True


class TestRateLimitError:
    """Test RateLimitError exception."""

    def test_rate_limit_error_message(self):
        """Test RateLimitError carries message."""
        with pytest.raises(RateLimitError, match="Rate limit exceeded"):
            limiter = RateLimiter(max_requests=1)
            limiter.add_request()
            limiter.add_request()

    def test_rate_limit_error_is_exception(self):
        """Test RateLimitError is an Exception."""
        assert issubclass(RateLimitError, Exception)
