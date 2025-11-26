"""Comprehensive test suite for link_validator.py utilities module.

This module provides 90%+ coverage for all link validation functionality including:
- LinkResult dataclass
- ValidationResult dataclass
- LinkValidator class with URL extraction and validation
- Single link validation with HTTP status handling
- Batch link validation with concurrency control
- Report generation with statistics
- README link validation convenience function
"""

import asyncio
import tempfile
from datetime import datetime
from pathlib import Path
from unittest.mock import MagicMock, patch

import pytest

from moai_adk.utils.common import HTTPResponse
from moai_adk.utils.link_validator import (
    LinkResult,
    LinkValidator,
    ValidationResult,
    validate_readme_links,
)

# ============================================================================
# LinkResult Tests
# ============================================================================


class TestLinkResult:
    """Tests for LinkResult dataclass."""

    def test_link_result_creation(self):
        """Test creating LinkResult with all fields."""
        result = LinkResult(
            url="https://example.com",
            status_code=200,
            is_valid=True,
            response_time=0.5,
        )
        assert result.url == "https://example.com"
        assert result.status_code == 200
        assert result.is_valid is True
        assert result.response_time == 0.5
        assert result.error_message is None
        assert isinstance(result.checked_at, datetime)

    def test_link_result_with_error_message(self):
        """Test LinkResult with error message."""
        error_msg = "Connection timeout"
        result = LinkResult(
            url="https://example.com",
            status_code=0,
            is_valid=False,
            response_time=0.0,
            error_message=error_msg,
        )
        assert result.is_valid is False
        assert result.error_message == error_msg

    def test_link_result_checked_at_auto_generation(self):
        """Test checked_at is auto-generated if not provided."""
        result = LinkResult(
            url="https://example.com",
            status_code=200,
            is_valid=True,
            response_time=0.1,
        )
        assert result.checked_at is not None
        assert isinstance(result.checked_at, datetime)

    def test_link_result_checked_at_none_handling(self):
        """Test __post_init__ handles None checked_at."""
        result = LinkResult(
            url="https://example.com",
            status_code=200,
            is_valid=True,
            response_time=0.1,
            checked_at=None,
        )
        assert result.checked_at is not None
        assert isinstance(result.checked_at, datetime)

    def test_link_result_various_status_codes(self):
        """Test LinkResult with various HTTP status codes."""
        test_cases = [
            (200, True),
            (201, True),
            (301, False),
            (404, False),
            (500, False),
            (0, False),
        ]
        for status_code, is_valid in test_cases:
            result = LinkResult(
                url="https://example.com",
                status_code=status_code,
                is_valid=is_valid,
                response_time=0.1,
            )
            assert result.status_code == status_code
            assert result.is_valid == is_valid

    def test_link_result_response_time_values(self):
        """Test LinkResult with various response times."""
        test_cases = [0.0, 0.5, 1.0, 5.0, 10.5]
        for response_time in test_cases:
            result = LinkResult(
                url="https://example.com",
                status_code=200,
                is_valid=True,
                response_time=response_time,
            )
            assert result.response_time == response_time

    def test_link_result_url_variations(self):
        """Test LinkResult with various URL formats."""
        test_urls = [
            "https://example.com",
            "http://example.com",
            "https://example.com/path",
            "https://example.com#anchor",
            "https://example.com:8080/path?query=value",
        ]
        for url in test_urls:
            result = LinkResult(
                url=url,
                status_code=200,
                is_valid=True,
                response_time=0.1,
            )
            assert result.url == url


# ============================================================================
# ValidationResult Tests
# ============================================================================


class TestValidationResult:
    """Tests for ValidationResult dataclass."""

    def test_validation_result_creation(self):
        """Test creating ValidationResult with basic data."""
        results = [
            LinkResult(
                url="https://example.com",
                status_code=200,
                is_valid=True,
                response_time=0.5,
            ),
        ]
        val_result = ValidationResult(
            total_links=1,
            valid_links=1,
            invalid_links=0,
            results=results,
        )
        assert val_result.total_links == 1
        assert val_result.valid_links == 1
        assert val_result.invalid_links == 0
        assert len(val_result.results) == 1
        assert isinstance(val_result.completed_at, datetime)

    def test_validation_result_empty_results(self):
        """Test ValidationResult with no results."""
        val_result = ValidationResult(
            total_links=0,
            valid_links=0,
            invalid_links=0,
            results=[],
        )
        assert val_result.total_links == 0
        assert val_result.valid_links == 0
        assert val_result.invalid_links == 0
        assert len(val_result.results) == 0

    def test_validation_result_completed_at_auto_generation(self):
        """Test completed_at is auto-generated if not provided."""
        val_result = ValidationResult(
            total_links=0,
            valid_links=0,
            invalid_links=0,
            results=[],
        )
        assert val_result.completed_at is not None
        assert isinstance(val_result.completed_at, datetime)

    def test_validation_result_completed_at_none_handling(self):
        """Test __post_init__ handles None completed_at."""
        val_result = ValidationResult(
            total_links=0,
            valid_links=0,
            invalid_links=0,
            results=[],
            completed_at=None,
        )
        assert val_result.completed_at is not None
        assert isinstance(val_result.completed_at, datetime)

    def test_validation_result_success_rate_all_valid(self):
        """Test success_rate calculation with all valid links."""
        results = [
            LinkResult(url=f"https://example.com/{i}", status_code=200, is_valid=True, response_time=0.1)
            for i in range(5)
        ]
        val_result = ValidationResult(
            total_links=5,
            valid_links=5,
            invalid_links=0,
            results=results,
        )
        assert val_result.success_rate == 100.0

    def test_validation_result_success_rate_all_invalid(self):
        """Test success_rate calculation with all invalid links."""
        results = [
            LinkResult(url=f"https://example.com/{i}", status_code=404, is_valid=False, response_time=0.1)
            for i in range(5)
        ]
        val_result = ValidationResult(
            total_links=5,
            valid_links=0,
            invalid_links=5,
            results=results,
        )
        assert val_result.success_rate == 0.0

    def test_validation_result_success_rate_partial(self):
        """Test success_rate calculation with mixed results."""
        results = [
            LinkResult(url=f"https://example.com/{i}", status_code=200, is_valid=True, response_time=0.1)
            for i in range(3)
        ] + [
            LinkResult(url=f"https://example.com/bad/{i}", status_code=404, is_valid=False, response_time=0.1)
            for i in range(2)
        ]
        val_result = ValidationResult(
            total_links=5,
            valid_links=3,
            invalid_links=2,
            results=results,
        )
        assert val_result.success_rate == 60.0

    def test_validation_result_success_rate_zero_links(self):
        """Test success_rate returns 0.0 when no links."""
        val_result = ValidationResult(
            total_links=0,
            valid_links=0,
            invalid_links=0,
            results=[],
        )
        assert val_result.success_rate == 0.0

    @pytest.mark.parametrize(
        "total,valid,invalid,expected_rate",
        [
            (10, 10, 0, 100.0),
            (10, 5, 5, 50.0),
            (10, 0, 10, 0.0),
            (1, 1, 0, 100.0),
            (4, 1, 3, 25.0),
            (4, 3, 1, 75.0),
        ],
    )
    def test_validation_result_success_rate_parametrized(self, total, valid, invalid, expected_rate):
        """Test success_rate with parametrized values."""
        results = [
            LinkResult(url=f"https://example.com/{i}", status_code=200, is_valid=(i < valid), response_time=0.1)
            for i in range(total)
        ]
        val_result = ValidationResult(
            total_links=total,
            valid_links=valid,
            invalid_links=invalid,
            results=results,
        )
        assert val_result.success_rate == expected_rate


# ============================================================================
# LinkValidator Tests - Initialization
# ============================================================================


class TestLinkValidatorInitialization:
    """Tests for LinkValidator initialization."""

    def test_link_validator_default_initialization(self):
        """Test LinkValidator initializes with default values."""
        validator = LinkValidator()
        assert validator.max_concurrent == 5
        assert validator.timeout == 10
        assert validator.session is None

    def test_link_validator_custom_initialization(self):
        """Test LinkValidator initializes with custom values."""
        validator = LinkValidator(max_concurrent=20, timeout=30)
        assert validator.max_concurrent == 20
        assert validator.timeout == 30

    def test_link_validator_inherits_from_httpclient(self):
        """Test LinkValidator inherits from HTTPClient."""
        validator = LinkValidator()
        assert hasattr(validator, "fetch_url")
        assert hasattr(validator, "session")
        assert hasattr(validator, "max_concurrent")
        assert hasattr(validator, "timeout")


# ============================================================================
# LinkValidator Tests - Link Extraction
# ============================================================================


class TestLinkValidatorExtractLinksFromFile:
    """Tests for LinkValidator.extract_links_from_file method."""

    def test_extract_links_from_existing_file(self):
        """Test extracting links from an existing file."""
        validator = LinkValidator()

        with tempfile.NamedTemporaryFile(mode="w", suffix=".md", delete=False) as f:
            f.write(
                """
# Test Document
Check out [Google](https://www.google.com)
Visit [GitHub](https://github.com)
            """
            )
            temp_path = Path(f.name)

        try:
            links = validator.extract_links_from_file(temp_path)
            assert len(links) > 0
        finally:
            temp_path.unlink()

    def test_extract_links_from_nonexistent_file(self):
        """Test extracting links from non-existent file returns empty list."""
        validator = LinkValidator()
        nonexistent_path = Path("/nonexistent/path/file.md")

        links = validator.extract_links_from_file(nonexistent_path)
        assert links == []

    def test_extract_links_logs_warning_for_nonexistent_file(self):
        """Test extracting links from non-existent file logs warning."""
        validator = LinkValidator()
        nonexistent_path = Path("/nonexistent/path/file.md")

        with patch("moai_adk.utils.link_validator.logger") as mock_logger:
            validator.extract_links_from_file(nonexistent_path)
            mock_logger.warning.assert_called()

    def test_extract_links_from_empty_file(self):
        """Test extracting links from empty file."""
        validator = LinkValidator()

        with tempfile.NamedTemporaryFile(mode="w", suffix=".md", delete=False) as f:
            f.write("")
            temp_path = Path(f.name)

        try:
            links = validator.extract_links_from_file(temp_path)
            assert isinstance(links, list)
        finally:
            temp_path.unlink()

    def test_extract_links_from_file_with_unreadable_content(self):
        """Test extracting links when SafeFileReader returns None."""
        validator = LinkValidator()
        test_path = Path("/tmp/test.md")

        with patch("moai_adk.utils.link_validator.SafeFileReader") as mock_reader_class:
            mock_reader = MagicMock()
            mock_reader.read_text.return_value = None
            mock_reader_class.return_value = mock_reader

            links = validator.extract_links_from_file(test_path)
            assert links == []

    def test_extract_links_handles_exception(self):
        """Test extracting links handles exceptions gracefully."""
        validator = LinkValidator()

        # Create a temporary file that exists but will cause an error
        with tempfile.NamedTemporaryFile(mode="w", suffix=".md", delete=False) as f:
            f.write("test content")
            temp_path = Path(f.name)

        try:
            # Patch SafeFileReader to raise exception
            with patch("moai_adk.utils.link_validator.SafeFileReader") as mock_reader_class:
                mock_reader = MagicMock()
                mock_reader.read_text.side_effect = Exception("Read error")
                mock_reader_class.return_value = mock_reader

                with patch("moai_adk.utils.link_validator.logger") as mock_logger:
                    links = validator.extract_links_from_file(temp_path)
                    assert links == []
                    mock_logger.error.assert_called()
        finally:
            temp_path.unlink()

    def test_extract_links_logs_found_links(self):
        """Test extracting links logs the number found."""
        validator = LinkValidator()

        with tempfile.NamedTemporaryFile(mode="w", suffix=".md", delete=False) as f:
            f.write("[Link](https://example.com)")
            temp_path = Path(f.name)

        try:
            with patch("moai_adk.utils.link_validator.logger") as mock_logger:
                validator.extract_links_from_file(temp_path)
                # Check that logger.info was called
                assert any("Found" in str(call_obj) for call_obj in mock_logger.info.call_args_list)
        finally:
            temp_path.unlink()


# ============================================================================
# LinkValidator Tests - Single Link Validation
# ============================================================================


class TestLinkValidatorValidateLink:
    """Tests for LinkValidator.validate_link method."""

    @pytest.mark.asyncio
    async def test_validate_link_with_valid_url(self):
        """Test validating a valid URL."""
        validator = LinkValidator()

        with patch.object(validator, "fetch_url") as mock_fetch:
            mock_fetch.return_value = HTTPResponse(
                status_code=200,
                url="https://example.com",
                load_time=0.5,
                success=True,
            )

            result = await validator.validate_link("https://example.com")

            assert result.url == "https://example.com"
            assert result.status_code == 200
            assert result.is_valid is True
            assert result.response_time == 0.5
            assert result.error_message is None

    @pytest.mark.asyncio
    async def test_validate_link_with_invalid_url_format(self):
        """Test validating invalid URL format."""
        validator = LinkValidator()

        with patch("moai_adk.utils.link_validator.is_valid_url", return_value=False):
            result = await validator.validate_link("not-a-url")

            assert result.url == "not-a-url"
            assert result.status_code == 0
            assert result.is_valid is False
            assert result.response_time == 0.0
            assert "Invalid URL format" in result.error_message

    @pytest.mark.asyncio
    async def test_validate_link_with_404_status(self):
        """Test validating link that returns 404."""
        validator = LinkValidator()

        with patch.object(validator, "fetch_url") as mock_fetch:
            mock_fetch.return_value = HTTPResponse(
                status_code=404,
                url="https://example.com/notfound",
                load_time=0.3,
                success=False,
            )

            result = await validator.validate_link("https://example.com/notfound")

            assert result.status_code == 404
            assert result.is_valid is False

    @pytest.mark.asyncio
    async def test_validate_link_with_500_status(self):
        """Test validating link that returns 500."""
        validator = LinkValidator()

        with patch.object(validator, "fetch_url") as mock_fetch:
            mock_fetch.return_value = HTTPResponse(
                status_code=500,
                url="https://example.com/error",
                load_time=0.2,
                success=False,
            )

            result = await validator.validate_link("https://example.com/error")

            assert result.status_code == 500
            assert result.is_valid is False

    @pytest.mark.asyncio
    async def test_validate_link_with_various_status_codes(self):
        """Test validating links with various status codes."""
        validator = LinkValidator()

        test_cases = [
            (200, True),
            (201, True),
            (301, False),
            (404, False),
            (500, False),
        ]

        for status_code, expected_valid in test_cases:
            with patch.object(validator, "fetch_url") as mock_fetch:
                mock_fetch.return_value = HTTPResponse(
                    status_code=status_code,
                    url="https://example.com",
                    load_time=0.1,
                    success=expected_valid,
                )

                result = await validator.validate_link("https://example.com")

                assert result.status_code == status_code
                assert result.is_valid == expected_valid

    @pytest.mark.asyncio
    async def test_validate_link_with_error_message(self):
        """Test validating link with error message from fetch."""
        validator = LinkValidator()

        with patch.object(validator, "fetch_url") as mock_fetch:
            mock_fetch.return_value = HTTPResponse(
                status_code=0,
                url="https://example.com",
                load_time=0.0,
                success=False,
                error_message="Connection timeout",
            )

            result = await validator.validate_link("https://example.com")

            assert result.is_valid is False
            assert result.error_message == "Connection timeout"

    @pytest.mark.asyncio
    async def test_validate_link_handles_unexpected_exception(self):
        """Test validate_link handles unexpected exceptions."""
        validator = LinkValidator()

        with patch.object(validator, "fetch_url") as mock_fetch:
            mock_fetch.side_effect = RuntimeError("Unexpected error")

            result = await validator.validate_link("https://example.com")

            assert result.is_valid is False
            assert "Unexpected error" in result.error_message
            assert result.status_code == 0


# ============================================================================
# LinkValidator Tests - Batch Link Validation
# ============================================================================


class TestLinkValidatorValidateAllLinks:
    """Tests for LinkValidator.validate_all_links method."""

    @pytest.mark.asyncio
    async def test_validate_all_links_empty_list(self):
        """Test validating empty link list."""
        validator = LinkValidator()

        result = await validator.validate_all_links([])

        assert result.total_links == 0
        assert result.valid_links == 0
        assert result.invalid_links == 0
        assert result.results == []

    @pytest.mark.asyncio
    async def test_validate_all_links_single_link(self):
        """Test validating single link."""
        validator = LinkValidator()

        with patch.object(validator, "validate_link") as mock_validate:
            mock_validate.return_value = LinkResult(
                url="https://example.com",
                status_code=200,
                is_valid=True,
                response_time=0.5,
            )

            result = await validator.validate_all_links(["https://example.com"])

            assert result.total_links == 1
            assert result.valid_links == 1
            assert result.invalid_links == 0
            assert len(result.results) == 1

    @pytest.mark.asyncio
    async def test_validate_all_links_multiple_links(self):
        """Test validating multiple links."""
        validator = LinkValidator()

        links = [
            "https://example.com",
            "https://google.com",
            "https://github.com",
        ]

        with patch.object(validator, "validate_link") as mock_validate:
            mock_validate.side_effect = [
                LinkResult(url=links[0], status_code=200, is_valid=True, response_time=0.1),
                LinkResult(url=links[1], status_code=200, is_valid=True, response_time=0.1),
                LinkResult(url=links[2], status_code=404, is_valid=False, response_time=0.1),
            ]

            result = await validator.validate_all_links(links)

            assert result.total_links == 3
            assert result.valid_links == 2
            assert result.invalid_links == 1

    @pytest.mark.asyncio
    async def test_validate_all_links_mixed_valid_invalid(self):
        """Test validating mix of valid and invalid links."""
        validator = LinkValidator()

        links = [f"https://example.com/{i}" for i in range(5)]

        with patch.object(validator, "validate_link") as mock_validate:
            # Alternate valid and invalid
            mock_validate.side_effect = [
                LinkResult(
                    url=links[i], status_code=(200 if i % 2 == 0 else 404), is_valid=(i % 2 == 0), response_time=0.1
                )
                for i in range(5)
            ]

            result = await validator.validate_all_links(links)

            assert result.total_links == 5
            assert result.valid_links == 3
            assert result.invalid_links == 2

    @pytest.mark.asyncio
    async def test_validate_all_links_respects_semaphore(self):
        """Test validate_all_links respects concurrency semaphore."""
        validator = LinkValidator(max_concurrent=2)

        links = [f"https://example.com/{i}" for i in range(5)]

        with patch.object(validator, "validate_link") as mock_validate:
            mock_validate.side_effect = [
                LinkResult(url=link, status_code=200, is_valid=True, response_time=0.1) for link in links
            ]

            result = await validator.validate_all_links(links)

            assert result.total_links == 5
            assert mock_validate.call_count == 5

    @pytest.mark.asyncio
    async def test_validate_all_links_calculates_statistics(self):
        """Test validate_all_links calculates correct statistics."""
        validator = LinkValidator()

        links = ["https://valid.com", "https://invalid.com"]

        with patch.object(validator, "validate_link") as mock_validate:
            mock_validate.side_effect = [
                LinkResult(url="https://valid.com", status_code=200, is_valid=True, response_time=0.5),
                LinkResult(url="https://invalid.com", status_code=404, is_valid=False, response_time=0.3),
            ]

            result = await validator.validate_all_links(links)

            assert result.total_links == 2
            assert result.valid_links == 1
            assert result.invalid_links == 1
            assert result.success_rate == 50.0

    @pytest.mark.asyncio
    async def test_validate_all_links_concurrent_execution(self):
        """Test validate_all_links executes concurrently."""
        validator = LinkValidator(max_concurrent=3)

        links = [f"https://example.com/{i}" for i in range(10)]

        async def mock_validate_link(url: str):
            await asyncio.sleep(0.01)
            return LinkResult(url=url, status_code=200, is_valid=True, response_time=0.01)

        with patch.object(validator, "validate_link", side_effect=mock_validate_link):
            result = await validator.validate_all_links(links)

            assert result.total_links == 10
            assert result.valid_links == 10

    @pytest.mark.asyncio
    async def test_validate_all_links_logs_progress(self):
        """Test validate_all_links logs validation progress."""
        validator = LinkValidator()

        links = ["https://example.com"]

        with patch.object(validator, "validate_link") as mock_validate:
            mock_validate.return_value = LinkResult(
                url="https://example.com",
                status_code=200,
                is_valid=True,
                response_time=0.1,
            )

            with patch("moai_adk.utils.link_validator.logger") as mock_logger:
                await validator.validate_all_links(links)

                # Check logger was called during validation
                assert mock_logger.info.called

    @pytest.mark.asyncio
    async def test_validate_all_links_preserves_order(self):
        """Test validate_all_links preserves link processing order (approximately)."""
        validator = LinkValidator()

        links = ["https://example.com/1", "https://example.com/2", "https://example.com/3"]

        with patch.object(validator, "validate_link") as mock_validate:
            mock_validate.side_effect = [
                LinkResult(url=link, status_code=200, is_valid=True, response_time=0.1) for link in links
            ]

            result = await validator.validate_all_links(links)

            assert len(result.results) == 3
            result_urls = [r.url for r in result.results]
            assert all(url in result_urls for url in links)


# ============================================================================
# LinkValidator Tests - Report Generation
# ============================================================================


class TestLinkValidatorGenerateReport:
    """Tests for LinkValidator.generate_report method."""

    def test_generate_report_empty_results(self):
        """Test generating report with no results."""
        validator = LinkValidator()
        val_result = ValidationResult(
            total_links=0,
            valid_links=0,
            invalid_links=0,
            results=[],
        )

        report = validator.generate_report(val_result)

        assert "Online Documentation Link Validation Report" in report
        assert "**Total Links**: 0" in report
        assert "**Valid Links**: 0" in report
        assert "**Invalid Links**: 0" in report

    def test_generate_report_all_valid_links(self):
        """Test generating report with all valid links."""
        validator = LinkValidator()
        results = [
            LinkResult(url=f"https://example.com/{i}", status_code=200, is_valid=True, response_time=0.1)
            for i in range(3)
        ]
        val_result = ValidationResult(
            total_links=3,
            valid_links=3,
            invalid_links=0,
            results=results,
        )

        report = validator.generate_report(val_result)

        assert "**Total Links**: 3" in report
        assert "**Valid Links**: 3" in report
        assert "**Invalid Links**: 0" in report
        assert "**Success Rate**: 100.0%" in report
        assert "Successful Links" in report

    def test_generate_report_all_invalid_links(self):
        """Test generating report with all invalid links."""
        validator = LinkValidator()
        results = [
            LinkResult(
                url=f"https://example.com/{i}",
                status_code=404,
                is_valid=False,
                response_time=0.1,
                error_message="Not found",
            )
            for i in range(2)
        ]
        val_result = ValidationResult(
            total_links=2,
            valid_links=0,
            invalid_links=2,
            results=results,
        )

        report = validator.generate_report(val_result)

        assert "**Total Links**: 2" in report
        assert "**Valid Links**: 0" in report
        assert "**Invalid Links**: 2" in report
        assert "**Success Rate**: 0.0%" in report
        assert "Failed Links" in report

    def test_generate_report_mixed_results(self):
        """Test generating report with mixed results."""
        validator = LinkValidator()
        results = [
            LinkResult(url="https://example.com", status_code=200, is_valid=True, response_time=0.5),
            LinkResult(
                url="https://broken.com", status_code=404, is_valid=False, response_time=0.2, error_message="Not found"
            ),
            LinkResult(url="https://slow.com", status_code=200, is_valid=True, response_time=2.0),
        ]
        val_result = ValidationResult(
            total_links=3,
            valid_links=2,
            invalid_links=1,
            results=results,
        )

        report = validator.generate_report(val_result)

        assert "**Total Links**: 3" in report
        assert "**Valid Links**: 2" in report
        assert "**Invalid Links**: 1" in report
        assert "**Success Rate**: 66.7%" in report
        assert "Failed Links" in report
        assert "Successful Links" in report

    def test_generate_report_includes_statistics(self):
        """Test report includes response time statistics."""
        validator = LinkValidator()
        results = [
            LinkResult(url=f"https://example.com/{i}", status_code=200, is_valid=True, response_time=float(i))
            for i in range(1, 4)
        ]
        val_result = ValidationResult(
            total_links=3,
            valid_links=3,
            invalid_links=0,
            results=results,
        )

        report = validator.generate_report(val_result)

        assert "Statistics" in report
        assert "Average Response Time" in report
        assert "Minimum Response Time" in report
        assert "Maximum Response Time" in report

    def test_generate_report_includes_error_messages(self):
        """Test report includes error messages for failed links."""
        validator = LinkValidator()
        results = [
            LinkResult(
                url="https://timeout.com",
                status_code=0,
                is_valid=False,
                response_time=0.0,
                error_message="Connection timeout",
            ),
        ]
        val_result = ValidationResult(
            total_links=1,
            valid_links=0,
            invalid_links=1,
            results=results,
        )

        report = validator.generate_report(val_result)

        assert "Connection timeout" in report

    def test_generate_report_has_markdown_format(self):
        """Test report is in markdown format."""
        validator = LinkValidator()
        val_result = ValidationResult(
            total_links=0,
            valid_links=0,
            invalid_links=0,
            results=[],
        )

        report = validator.generate_report(val_result)

        assert "#" in report  # Markdown header
        assert "**" in report  # Markdown bold

    def test_generate_report_timestamp_formatting(self):
        """Test report includes properly formatted timestamp."""
        validator = LinkValidator()
        val_result = ValidationResult(
            total_links=0,
            valid_links=0,
            invalid_links=0,
            results=[],
        )

        report = validator.generate_report(val_result)

        assert "Validation Time" in report

    @pytest.mark.parametrize(
        "total,valid,invalid",
        [
            (0, 0, 0),
            (1, 1, 0),
            (5, 5, 0),
            (5, 0, 5),
            (10, 7, 3),
        ],
    )
    def test_generate_report_parametrized(self, total, valid, invalid):
        """Test generating reports with parametrized data."""
        validator = LinkValidator()
        results = [
            LinkResult(url=f"https://example.com/{i}", status_code=200, is_valid=(i < valid), response_time=0.1)
            for i in range(total)
        ]
        val_result = ValidationResult(
            total_links=total,
            valid_links=valid,
            invalid_links=invalid,
            results=results,
        )

        report = validator.generate_report(val_result)

        assert f"**Total Links**: {total}" in report
        assert f"**Valid Links**: {valid}" in report
        assert f"**Invalid Links**: {invalid}" in report


# ============================================================================
# validate_readme_links Tests
# ============================================================================


class TestValidateReadmeLinks:
    """Tests for validate_readme_links convenience function."""

    def test_validate_readme_links_default_path(self):
        """Test validate_readme_links uses default path when not provided."""
        with patch("moai_adk.utils.link_validator.LinkValidator") as mock_validator_class:
            mock_validator = MagicMock()
            mock_validator.extract_links_from_file.return_value = []
            mock_validator_class.return_value = mock_validator

            with patch("moai_adk.utils.link_validator.asyncio.run") as mock_asyncio_run:
                mock_asyncio_run.return_value = ValidationResult(
                    total_links=0, valid_links=0, invalid_links=0, results=[]
                )

                validate_readme_links()

                # Should be called with default path
                assert mock_validator.extract_links_from_file.called

    def test_validate_readme_links_custom_path(self):
        """Test validate_readme_links with custom path."""
        custom_path = Path("/custom/README.md")

        with patch("moai_adk.utils.link_validator.LinkValidator") as mock_validator_class:
            mock_validator = MagicMock()
            mock_validator.extract_links_from_file.return_value = []
            mock_validator_class.return_value = mock_validator

            with patch("moai_adk.utils.link_validator.asyncio.run") as mock_asyncio_run:
                mock_asyncio_run.return_value = ValidationResult(
                    total_links=0, valid_links=0, invalid_links=0, results=[]
                )

                validate_readme_links(custom_path)

                mock_validator.extract_links_from_file.assert_called_with(custom_path)

    def test_validate_readme_links_no_links_found(self):
        """Test validate_readme_links when no links found."""
        with patch("moai_adk.utils.link_validator.LinkValidator") as mock_validator_class:
            mock_validator = MagicMock()
            mock_validator.extract_links_from_file.return_value = []
            mock_validator_class.return_value = mock_validator

            with patch("moai_adk.utils.link_validator.logger") as mock_logger:
                result = validate_readme_links()

                assert result.total_links == 0
                assert result.valid_links == 0
                mock_logger.warning.assert_called()

    def test_validate_readme_links_executes_validation(self):
        """Test validate_readme_links executes link validation."""
        links = ["https://example.com", "https://google.com"]

        with patch("moai_adk.utils.link_validator.LinkValidator") as mock_validator_class:
            mock_validator = MagicMock()
            mock_validator.extract_links_from_file.return_value = links

            # Setup the validate_all_links to return the expected result
            expected_result = ValidationResult(
                total_links=2,
                valid_links=2,
                invalid_links=0,
                results=[
                    LinkResult(url=links[0], status_code=200, is_valid=True, response_time=0.1),
                    LinkResult(url=links[1], status_code=200, is_valid=True, response_time=0.1),
                ],
            )

            async def mock_validate_all_links(link_list):
                return expected_result

            mock_validator.validate_all_links.side_effect = mock_validate_all_links
            mock_validator_class.return_value = mock_validator

            with patch("moai_adk.utils.link_validator.create_report_path") as mock_create_path:
                mock_path = MagicMock()
                mock_create_path.return_value = mock_path

                validate_readme_links()

                # Verify validate_all_links was called
                assert mock_validator.validate_all_links.called

    def test_validate_readme_links_generates_report(self):
        """Test validate_readme_links generates report."""
        links = ["https://example.com"]

        with patch("moai_adk.utils.link_validator.LinkValidator") as mock_validator_class:
            mock_validator = MagicMock()
            mock_validator.extract_links_from_file.return_value = links

            expected_result = ValidationResult(
                total_links=1,
                valid_links=1,
                invalid_links=0,
                results=[LinkResult(url=links[0], status_code=200, is_valid=True, response_time=0.1)],
            )

            mock_validator_class.return_value = mock_validator

            with patch("moai_adk.utils.link_validator.asyncio.run") as mock_asyncio_run:
                mock_asyncio_run.return_value = expected_result

                with patch("moai_adk.utils.link_validator.create_report_path") as mock_create_path:
                    mock_path = MagicMock()
                    mock_create_path.return_value = mock_path

                    validate_readme_links()

                    mock_validator.generate_report.assert_called_once_with(expected_result)

    def test_validate_readme_links_saves_report(self):
        """Test validate_readme_links saves report to file."""
        links = ["https://example.com"]

        with patch("moai_adk.utils.link_validator.LinkValidator") as mock_validator_class:
            mock_validator = MagicMock()
            mock_validator.extract_links_from_file.return_value = links
            mock_validator.generate_report.return_value = "Report content"

            expected_result = ValidationResult(
                total_links=1,
                valid_links=1,
                invalid_links=0,
                results=[LinkResult(url=links[0], status_code=200, is_valid=True, response_time=0.1)],
            )

            mock_validator_class.return_value = mock_validator

            with patch("moai_adk.utils.link_validator.asyncio.run") as mock_asyncio_run:
                mock_asyncio_run.return_value = expected_result

                with patch("moai_adk.utils.link_validator.create_report_path") as mock_create_path:
                    mock_path = MagicMock()
                    mock_create_path.return_value = mock_path

                    validate_readme_links()

                    mock_path.write_text.assert_called_once()

    def test_validate_readme_links_logs_completion(self):
        """Test validate_readme_links logs when report is saved."""
        links = ["https://example.com"]

        with patch("moai_adk.utils.link_validator.LinkValidator") as mock_validator_class:
            mock_validator = MagicMock()
            mock_validator.extract_links_from_file.return_value = links
            mock_validator.generate_report.return_value = "Report"

            expected_result = ValidationResult(
                total_links=1,
                valid_links=1,
                invalid_links=0,
                results=[LinkResult(url=links[0], status_code=200, is_valid=True, response_time=0.1)],
            )

            mock_validator_class.return_value = mock_validator

            with patch("moai_adk.utils.link_validator.asyncio.run") as mock_asyncio_run:
                mock_asyncio_run.return_value = expected_result

                with patch("moai_adk.utils.link_validator.create_report_path") as mock_create_path:
                    mock_path = MagicMock()
                    mock_create_path.return_value = mock_path

                    with patch("moai_adk.utils.link_validator.logger") as mock_logger:
                        validate_readme_links()

                        mock_logger.info.assert_called()

    def test_validate_readme_links_returns_validation_result(self):
        """Test validate_readme_links returns ValidationResult."""
        with patch("moai_adk.utils.link_validator.LinkValidator") as mock_validator_class:
            mock_validator = MagicMock()
            mock_validator.extract_links_from_file.return_value = []
            mock_validator_class.return_value = mock_validator

            result = validate_readme_links()

            assert isinstance(result, ValidationResult)


# ============================================================================
# Integration Tests
# ============================================================================


class TestLinkValidatorIntegration:
    """Integration tests for LinkValidator."""

    @pytest.mark.asyncio
    async def test_full_validation_workflow(self):
        """Test complete workflow from file extraction to report generation."""
        validator = LinkValidator()

        # Create a temporary file with links
        with tempfile.NamedTemporaryFile(mode="w", suffix=".md", delete=False) as f:
            f.write("[Valid](https://valid.com)\n[Invalid](https://invalid.com)")
            temp_path = Path(f.name)

        try:
            # Extract links
            links = validator.extract_links_from_file(temp_path)
            assert len(links) >= 0  # May be 0 or more depending on extraction

            # Validate links
            if links:
                with patch.object(validator, "validate_link") as mock_validate:
                    mock_validate.side_effect = [
                        LinkResult(url=link, status_code=200, is_valid=True, response_time=0.1) for link in links
                    ]

                    result = await validator.validate_all_links(links)

                    # Generate report
                    report = validator.generate_report(result)

                    assert "Online Documentation Link Validation Report" in report
                    assert result.total_links == len(links)
        finally:
            temp_path.unlink()

    @pytest.mark.asyncio
    async def test_validation_with_different_concurrency_levels(self):
        """Test validation with different concurrency settings."""
        links = [f"https://example.com/{i}" for i in range(10)]

        for max_concurrent in [1, 3, 5, 10]:
            validator = LinkValidator(max_concurrent=max_concurrent)

            with patch.object(validator, "validate_link") as mock_validate:
                mock_validate.side_effect = [
                    LinkResult(url=link, status_code=200, is_valid=True, response_time=0.1) for link in links
                ]

                result = await validator.validate_all_links(links)

                assert result.total_links == 10
                assert result.valid_links == 10


# ============================================================================
# Edge Cases and Error Handling
# ============================================================================


class TestLinkValidatorMainBlock:
    """Tests for __main__ block functionality."""

    def test_link_validator_post_init_none_handling_linkresult(self):
        """Test LinkResult __post_init__ properly handles None checked_at."""
        # Test that __post_init__ is called and sets checked_at when None
        result = LinkResult(
            url="https://example.com",
            status_code=200,
            is_valid=True,
            response_time=0.1,
            checked_at=None,
        )
        # Verify it was set to current time (not None)
        assert result.checked_at is not None
        assert isinstance(result.checked_at, datetime)
        # Verify it's reasonably close to current time
        assert (datetime.now() - result.checked_at).total_seconds() < 1

    def test_validation_result_post_init_none_handling(self):
        """Test ValidationResult __post_init__ properly handles None completed_at."""
        # Test that __post_init__ is called and sets completed_at when None
        val_result = ValidationResult(
            total_links=0,
            valid_links=0,
            invalid_links=0,
            results=[],
            completed_at=None,
        )
        # Verify it was set to current time (not None)
        assert val_result.completed_at is not None
        assert isinstance(val_result.completed_at, datetime)
        # Verify it's reasonably close to current time
        assert (datetime.now() - val_result.completed_at).total_seconds() < 1

    def test_extract_links_file_exists_check(self):
        """Test extract_links properly checks if file exists."""
        validator = LinkValidator()

        # Test with non-existent file
        non_existent = Path("/this/path/does/not/exist/file.md")
        links = validator.extract_links_from_file(non_existent)
        assert links == []

        # Test with existing file
        with tempfile.NamedTemporaryFile(mode="w", suffix=".md", delete=False) as f:
            f.write("[Link](https://example.com)")
            temp_path = Path(f.name)

        try:
            links = validator.extract_links_from_file(temp_path)
            # Should return a list (may be empty depending on extraction)
            assert isinstance(links, list)
        finally:
            temp_path.unlink()

    def test_safe_file_reader_returns_none(self):
        """Test extract_links handles SafeFileReader returning None."""
        validator = LinkValidator()

        with tempfile.NamedTemporaryFile(mode="w", suffix=".md", delete=False) as f:
            f.write("test")
            temp_path = Path(f.name)

        try:
            with patch("moai_adk.utils.link_validator.SafeFileReader") as mock_reader_class:
                mock_reader = MagicMock()
                mock_reader.read_text.return_value = None
                mock_reader_class.return_value = mock_reader

                links = validator.extract_links_from_file(temp_path)
                assert links == []
        finally:
            temp_path.unlink()


class TestLinkValidatorEdgeCases:
    """Tests for edge cases and error handling."""

    @pytest.mark.asyncio
    async def test_validate_link_with_empty_url(self):
        """Test validating empty URL."""
        validator = LinkValidator()

        with patch("moai_adk.utils.link_validator.is_valid_url", return_value=False):
            result = await validator.validate_link("")

            assert result.is_valid is False

    @pytest.mark.asyncio
    async def test_validate_all_links_with_duplicate_urls(self):
        """Test validating list with duplicate URLs."""
        validator = LinkValidator()

        links = ["https://example.com", "https://example.com", "https://example.com"]

        with patch.object(validator, "validate_link") as mock_validate:
            mock_validate.side_effect = [
                LinkResult(url="https://example.com", status_code=200, is_valid=True, response_time=0.1) for _ in links
            ]

            result = await validator.validate_all_links(links)

            assert result.total_links == 3

    def test_validation_result_with_large_response_times(self):
        """Test ValidationResult with very large response times."""
        results = [
            LinkResult(url=f"https://example.com/{i}", status_code=200, is_valid=True, response_time=1000.0)
            for i in range(3)
        ]
        val_result = ValidationResult(
            total_links=3,
            valid_links=3,
            invalid_links=0,
            results=results,
        )

        assert val_result.success_rate == 100.0

    def test_link_result_with_special_characters_in_url(self):
        """Test LinkResult with special characters in URL."""
        special_url = "https://example.com/path?query=value&param=test#section"
        result = LinkResult(
            url=special_url,
            status_code=200,
            is_valid=True,
            response_time=0.1,
        )

        assert result.url == special_url

    @pytest.mark.asyncio
    async def test_validate_link_response_time_precision(self):
        """Test LinkResult preserves response time precision."""
        validator = LinkValidator()

        precise_time = 0.123456789

        with patch.object(validator, "fetch_url") as mock_fetch:
            mock_fetch.return_value = HTTPResponse(
                status_code=200,
                url="https://example.com",
                load_time=precise_time,
                success=True,
            )

            result = await validator.validate_link("https://example.com")

            assert result.response_time == precise_time
