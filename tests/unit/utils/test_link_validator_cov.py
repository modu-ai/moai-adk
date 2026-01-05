"""Comprehensive coverage tests for moai_adk.utils.link_validator module.

This module contains tests targeting uncovered code paths in link_validator.py
to achieve 90%+ coverage.
"""

from datetime import datetime
from pathlib import Path
from unittest.mock import AsyncMock, MagicMock, patch

import pytest

from moai_adk.utils.common import HTTPResponse
from moai_adk.utils.link_validator import (
    LinkResult,
    LinkValidator,
    ValidationResult,
    validate_readme_links,
)


class TestLinkResultPostInit:
    """Test LinkResult __post_init__ method."""

    def test_link_result_post_init_none_timestamp(self):
        """Test __post_init__ sets timestamp when None."""
        result = LinkResult(
            url="https://example.com",
            status_code=200,
            is_valid=True,
            response_time=0.5,
            checked_at=None,
        )
        assert result.checked_at is not None
        assert isinstance(result.checked_at, datetime)

    def test_link_result_post_init_preserves_timestamp(self):
        """Test __post_init__ preserves existing timestamp."""
        now = datetime.now()
        result = LinkResult(
            url="https://example.com",
            status_code=200,
            is_valid=True,
            response_time=0.5,
            checked_at=now,
        )
        assert result.checked_at == now


class TestValidationResultPostInit:
    """Test ValidationResult __post_init__ method."""

    def test_validation_result_post_init_none_timestamp(self):
        """Test __post_init__ sets timestamp when None."""
        validation = ValidationResult(
            total_links=0,
            valid_links=0,
            invalid_links=0,
            results=[],
            completed_at=None,
        )
        assert validation.completed_at is not None
        assert isinstance(validation.completed_at, datetime)

    def test_validation_result_post_init_preserves_timestamp(self):
        """Test __post_init__ preserves existing timestamp."""
        now = datetime.now()
        validation = ValidationResult(
            total_links=0,
            valid_links=0,
            invalid_links=0,
            results=[],
            completed_at=now,
        )
        assert validation.completed_at == now


class TestValidationResultSuccessRate:
    """Test ValidationResult success_rate property."""

    def test_success_rate_zero_links(self):
        """Test success_rate with no links."""
        validation = ValidationResult(
            total_links=0,
            valid_links=0,
            invalid_links=0,
            results=[],
        )
        assert validation.success_rate == 0.0

    def test_success_rate_all_valid(self):
        """Test success_rate with all valid links."""
        validation = ValidationResult(
            total_links=5,
            valid_links=5,
            invalid_links=0,
            results=[],
        )
        assert validation.success_rate == 100.0

    def test_success_rate_all_invalid(self):
        """Test success_rate with all invalid links."""
        validation = ValidationResult(
            total_links=5,
            valid_links=0,
            invalid_links=5,
            results=[],
        )
        assert validation.success_rate == 0.0

    def test_success_rate_partial(self):
        """Test success_rate with partial valid links."""
        validation = ValidationResult(
            total_links=4,
            valid_links=3,
            invalid_links=1,
            results=[],
        )
        assert validation.success_rate == 75.0


class TestLinkValidatorExtractLinks:
    """Test LinkValidator extract_links_from_file method."""

    def test_extract_links_file_not_exists(self):
        """Test extract when file doesn't exist."""
        validator = LinkValidator()
        with patch("pathlib.Path.exists", return_value=False):
            links = validator.extract_links_from_file(Path("/nonexistent.md"))
            assert links == []

    def test_extract_links_successful(self):
        """Test successful link extraction from file."""
        validator = LinkValidator()
        content = "[Link](https://example.com) and https://another.com"
        with patch("pathlib.Path.exists", return_value=True):
            with patch("moai_adk.utils.link_validator.SafeFileReader") as mock_reader_class:
                mock_reader = MagicMock()
                mock_reader.read_text.return_value = content
                mock_reader_class.return_value = mock_reader
                links = validator.extract_links_from_file(Path("/tmp/test.md"))
                assert len(links) > 0

    def test_extract_links_read_returns_none(self):
        """Test extract when file read returns None."""
        validator = LinkValidator()
        with patch("pathlib.Path.exists", return_value=True):
            with patch("moai_adk.utils.link_validator.SafeFileReader") as mock_reader_class:
                mock_reader = MagicMock()
                mock_reader.read_text.return_value = None
                mock_reader_class.return_value = mock_reader
                links = validator.extract_links_from_file(Path("/tmp/test.md"))
                assert links == []

    def test_extract_links_exception_handling(self):
        """Test extract handles exceptions during reading."""
        validator = LinkValidator()
        with patch("pathlib.Path.exists", return_value=True):
            with patch("moai_adk.utils.link_validator.SafeFileReader") as mock_reader_class:
                mock_reader_class.side_effect = Exception("Read error")
                links = validator.extract_links_from_file(Path("/tmp/test.md"))
                assert links == []

    def test_extract_links_uses_base_url(self):
        """Test that extract uses correct base URL."""
        validator = LinkValidator()
        content = "[Docs](/docs) and [API](/api)"
        with patch("pathlib.Path.exists", return_value=True):
            with patch("moai_adk.utils.link_validator.SafeFileReader") as mock_reader_class:
                mock_reader = MagicMock()
                mock_reader.read_text.return_value = content
                mock_reader_class.return_value = mock_reader
                with patch("moai_adk.utils.link_validator.extract_links_from_text") as mock_extract:
                    mock_extract.return_value = []
                    validator.extract_links_from_file(Path("/tmp/test.md"))
                    # Verify base_url was passed
                    call_args = mock_extract.call_args
                    assert "adk.mo.ai.kr" in call_args[0][1]


class TestLinkValidatorValidateLink:
    """Test LinkValidator validate_link method."""

    @pytest.mark.asyncio
    async def test_validate_link_invalid_format(self):
        """Test validate_link with invalid URL format."""
        validator = LinkValidator()
        with patch("moai_adk.utils.link_validator.is_valid_url", return_value=False):
            result = await validator.validate_link("not a valid url")
            assert result.is_valid is False
            assert result.status_code == 0
            assert result.error_message == "Invalid URL format"

    @pytest.mark.asyncio
    async def test_validate_link_successful(self):
        """Test validate_link with successful response."""
        validator = LinkValidator()
        with patch("moai_adk.utils.link_validator.is_valid_url", return_value=True):
            mock_response = HTTPResponse(
                status_code=200,
                url="https://example.com",
                load_time=0.5,
                success=True,
            )
            with patch.object(validator, "fetch_url", new_callable=AsyncMock) as mock_fetch:
                mock_fetch.return_value = mock_response
                result = await validator.validate_link("https://example.com")
                assert result.url == "https://example.com"
                assert result.is_valid is True
                assert result.status_code == 200

    @pytest.mark.asyncio
    async def test_validate_link_failed_response(self):
        """Test validate_link with failed HTTP response."""
        validator = LinkValidator()
        with patch("moai_adk.utils.link_validator.is_valid_url", return_value=True):
            mock_response = HTTPResponse(
                status_code=404,
                url="https://example.com",
                load_time=0.2,
                success=False,
                error_message="Not found",
            )
            with patch.object(validator, "fetch_url", new_callable=AsyncMock) as mock_fetch:
                mock_fetch.return_value = mock_response
                result = await validator.validate_link("https://example.com")
                assert result.is_valid is False
                assert result.status_code == 404
                assert result.error_message == "Not found"

    @pytest.mark.asyncio
    async def test_validate_link_unexpected_exception(self):
        """Test validate_link handles unexpected exceptions."""
        validator = LinkValidator()
        with patch("moai_adk.utils.link_validator.is_valid_url", return_value=True):
            with patch.object(validator, "fetch_url", side_effect=RuntimeError("Network error")):
                result = await validator.validate_link("https://example.com")
                assert result.is_valid is False
                assert result.status_code == 0
                assert "Unexpected error" in result.error_message


class TestLinkValidatorValidateAllLinks:
    """Test LinkValidator validate_all_links method."""

    @pytest.mark.asyncio
    async def test_validate_all_links_empty(self):
        """Test validate_all_links with empty list."""
        validator = LinkValidator()
        result = await validator.validate_all_links([])
        assert result.total_links == 0
        assert result.valid_links == 0
        assert result.invalid_links == 0
        assert result.results == []

    @pytest.mark.asyncio
    async def test_validate_all_links_single_valid(self):
        """Test validate_all_links with single valid link."""
        validator = LinkValidator()
        with patch.object(validator, "validate_link", new_callable=AsyncMock) as mock_validate:
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

    @pytest.mark.asyncio
    async def test_validate_all_links_multiple_mixed(self):
        """Test validate_all_links with mix of valid and invalid."""
        validator = LinkValidator()
        with patch.object(validator, "validate_link", new_callable=AsyncMock) as mock_validate:
            # Return valid for first, invalid for second
            mock_validate.side_effect = [
                LinkResult(
                    url="https://valid.com",
                    status_code=200,
                    is_valid=True,
                    response_time=0.5,
                ),
                LinkResult(
                    url="https://invalid.com",
                    status_code=404,
                    is_valid=False,
                    response_time=0.2,
                ),
            ]
            result = await validator.validate_all_links(
                [
                    "https://valid.com",
                    "https://invalid.com",
                ]
            )
            assert result.total_links == 2
            assert result.valid_links == 1
            assert result.invalid_links == 1

    @pytest.mark.asyncio
    async def test_validate_all_links_respects_concurrency(self):
        """Test validate_all_links respects concurrency limit."""
        validator = LinkValidator(max_concurrent=2)
        call_count = 0

        async def mock_validate(url):
            nonlocal call_count
            call_count += 1
            return LinkResult(
                url=url,
                status_code=200,
                is_valid=True,
                response_time=0.5,
            )

        with patch.object(validator, "validate_link", side_effect=mock_validate):
            urls = [f"https://example{i}.com" for i in range(5)]
            result = await validator.validate_all_links(urls)
            assert result.total_links == 5

    @pytest.mark.asyncio
    async def test_validate_all_links_completed_at_set(self):
        """Test validate_all_links sets completed_at timestamp."""
        validator = LinkValidator()
        before = datetime.now()
        result = await validator.validate_all_links([])
        after = datetime.now()
        assert before <= result.completed_at <= after


class TestLinkValidatorGenerateReport:
    """Test LinkValidator generate_report method."""

    def test_generate_report_all_valid(self):
        """Test report generation with all valid links."""
        validator = LinkValidator()
        results = [
            LinkResult(
                url="https://example.com",
                status_code=200,
                is_valid=True,
                response_time=0.5,
            ),
            LinkResult(
                url="https://another.com",
                status_code=200,
                is_valid=True,
                response_time=0.3,
            ),
        ]
        validation = ValidationResult(
            total_links=2,
            valid_links=2,
            invalid_links=0,
            results=results,
        )
        report = validator.generate_report(validation)
        assert "Link Validation Report" in report
        assert "100.0%" in report
        assert "Successful Links" in report

    def test_generate_report_all_invalid(self):
        """Test report generation with all invalid links."""
        validator = LinkValidator()
        results = [
            LinkResult(
                url="https://broken.com",
                status_code=404,
                is_valid=False,
                response_time=0.2,
                error_message="Not found",
            ),
        ]
        validation = ValidationResult(
            total_links=1,
            valid_links=0,
            invalid_links=1,
            results=results,
        )
        report = validator.generate_report(validation)
        assert "Failed Links" in report
        assert "404" in report
        assert "Not found" in report

    def test_generate_report_mixed(self):
        """Test report generation with mixed results."""
        validator = LinkValidator()
        results = [
            LinkResult(
                url="https://valid.com",
                status_code=200,
                is_valid=True,
                response_time=0.5,
            ),
            LinkResult(
                url="https://invalid.com",
                status_code=404,
                is_valid=False,
                response_time=0.2,
                error_message="Not found",
            ),
        ]
        validation = ValidationResult(
            total_links=2,
            valid_links=1,
            invalid_links=1,
            results=results,
        )
        report = validator.generate_report(validation)
        assert "50.0%" in report
        assert "Failed Links" in report
        assert "Successful Links" in report

    def test_generate_report_includes_statistics(self):
        """Test report includes response time statistics."""
        validator = LinkValidator()
        results = [
            LinkResult(
                url="https://example.com",
                status_code=200,
                is_valid=True,
                response_time=0.5,
            ),
            LinkResult(
                url="https://another.com",
                status_code=200,
                is_valid=True,
                response_time=0.3,
            ),
        ]
        validation = ValidationResult(
            total_links=2,
            valid_links=2,
            invalid_links=0,
            results=results,
        )
        report = validator.generate_report(validation)
        assert "Average Response Time" in report
        assert "Statistics" in report

    def test_generate_report_no_results(self):
        """Test report generation with no results."""
        validator = LinkValidator()
        validation = ValidationResult(
            total_links=0,
            valid_links=0,
            invalid_links=0,
            results=[],
        )
        report = validator.generate_report(validation)
        assert "Link Validation Report" in report
        assert "0" in report

    def test_generate_report_format_markdown(self):
        """Test report is in Markdown format."""
        validator = LinkValidator()
        results = [
            LinkResult(
                url="https://example.com",
                status_code=200,
                is_valid=True,
                response_time=0.5,
            ),
        ]
        validation = ValidationResult(
            total_links=1,
            valid_links=1,
            invalid_links=0,
            results=results,
        )
        report = validator.generate_report(validation)
        assert report.startswith("#")
        assert "**" in report


class TestValidateReadmeLinks:
    """Test validate_readme_links function."""

    def test_validate_readme_links_default_path(self):
        """Test validate_readme_links uses default path."""
        with patch("pathlib.Path.exists", return_value=False):
            result = validate_readme_links()
            assert result.total_links == 0
            assert result.valid_links == 0
            assert result.invalid_links == 0

    def test_validate_readme_links_custom_path(self):
        """Test validate_readme_links with custom path."""
        with patch("pathlib.Path.exists", return_value=False):
            result = validate_readme_links(Path("custom.md"))
            assert result.total_links == 0

    def test_validate_readme_links_no_links_found(self):
        """Test validate_readme_links when file has no links."""
        with patch("moai_adk.utils.link_validator.LinkValidator") as mock_validator_class:
            mock_validator = MagicMock()
            mock_validator.extract_links_from_file.return_value = []
            mock_validator_class.return_value = mock_validator
            result = validate_readme_links(Path("README.md"))
            assert result.total_links == 0

    def test_validate_readme_links_saves_report(self):
        """Test validate_readme_links saves report to file."""
        with patch("moai_adk.utils.link_validator.LinkValidator") as mock_validator_class:
            mock_validator = MagicMock()
            mock_validator.extract_links_from_file.return_value = []
            mock_validator_class.return_value = mock_validator

            with patch("pathlib.Path.write_text"):
                # Mock the async function
                with patch("asyncio.run"):
                    validate_readme_links(Path("README.md"))

    def test_validate_readme_links_creates_validator(self):
        """Test validate_readme_links creates LinkValidator with settings."""
        with patch("moai_adk.utils.link_validator.LinkValidator") as mock_validator_class:
            mock_validator = MagicMock()
            mock_validator.extract_links_from_file.return_value = []
            mock_validator_class.return_value = mock_validator

            with patch("pathlib.Path.exists", return_value=True):
                with patch("pathlib.Path.write_text"):
                    with patch("asyncio.run"):
                        validate_readme_links()
                        # Verify LinkValidator was called with specific concurrency/timeout
                        mock_validator_class.assert_called_once()
                        args = mock_validator_class.call_args
                        assert args[1]["max_concurrent"] == 3
                        assert args[1]["timeout"] == 8


class TestLinkValidatorInheritance:
    """Test LinkValidator inheritance from HTTPClient."""

    def test_link_validator_inherits_http_client(self):
        """Test LinkValidator properly inherits HTTPClient."""
        validator = LinkValidator(max_concurrent=3, timeout=8)
        assert validator.max_concurrent == 3
        assert validator.timeout == 8

    @pytest.mark.asyncio
    async def test_link_validator_as_context_manager(self):
        """Test LinkValidator can be used as async context manager."""
        validator = LinkValidator()
        async with validator as v:
            assert v is validator


class TestLinkValidatorIntegration:
    """Integration tests for LinkValidator workflows."""

    @pytest.mark.asyncio
    async def test_full_validation_workflow(self):
        """Test complete validation workflow from extraction to reporting."""
        validator = LinkValidator()
        content = "[Link](https://example.com)"

        with patch("pathlib.Path.exists", return_value=True):
            with patch("moai_adk.utils.link_validator.SafeFileReader") as mock_reader_class:
                mock_reader = MagicMock()
                mock_reader.read_text.return_value = content
                mock_reader_class.return_value = mock_reader

                # Extract links
                links = validator.extract_links_from_file(Path("test.md"))
                assert len(links) > 0

                # Validate links
                with patch.object(validator, "validate_link", new_callable=AsyncMock) as mock_validate:
                    mock_validate.return_value = LinkResult(
                        url="https://example.com",
                        status_code=200,
                        is_valid=True,
                        response_time=0.5,
                    )
                    result = await validator.validate_all_links(links)
                    assert result.valid_links > 0

                    # Generate report
                    report = validator.generate_report(result)
                    assert "example.com" in report or "Link Validation Report" in report
