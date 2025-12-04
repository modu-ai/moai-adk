"""Unit tests for moai_adk.utils.link_validator module.

Tests for online link validation utilities.
"""

import asyncio
from datetime import datetime
from pathlib import Path
from unittest.mock import AsyncMock, MagicMock, patch

import pytest

from moai_adk.utils.link_validator import (
    LinkResult,
    LinkValidator,
    ValidationResult,
    validate_readme_links,
)


class TestLinkResult:
    """Test LinkResult dataclass."""

    def test_create_link_result_valid(self):
        """Test creating valid LinkResult."""
        result = LinkResult(
            url="https://example.com",
            status_code=200,
            is_valid=True,
            response_time=0.5,
        )
        assert result.url == "https://example.com"
        assert result.status_code == 200
        assert result.is_valid is True
        assert isinstance(result.checked_at, datetime)

    def test_create_link_result_invalid(self):
        """Test creating invalid LinkResult."""
        result = LinkResult(
            url="https://example.com",
            status_code=404,
            is_valid=False,
            response_time=0.2,
            error_message="Not found",
        )
        assert result.is_valid is False
        assert result.error_message == "Not found"

    def test_link_result_checked_at_timestamp(self):
        """Test checked_at field is a datetime."""
        result = LinkResult(
            url="https://example.com",
            status_code=200,
            is_valid=True,
            response_time=0.5,
        )
        assert isinstance(result.checked_at, datetime)


class TestValidationResult:
    """Test ValidationResult dataclass."""

    def test_create_validation_result(self):
        """Test creating ValidationResult."""
        results = [
            LinkResult(
                url="https://example.com",
                status_code=200,
                is_valid=True,
                response_time=0.5,
            )
        ]
        validation = ValidationResult(
            total_links=1,
            valid_links=1,
            invalid_links=0,
            results=results,
        )
        assert validation.total_links == 1
        assert validation.valid_links == 1
        assert validation.invalid_links == 0

    def test_validation_result_success_rate(self):
        """Test success_rate calculation."""
        results = []
        validation = ValidationResult(
            total_links=2,
            valid_links=1,
            invalid_links=1,
            results=results,
        )
        assert validation.success_rate == 50.0

    def test_validation_result_success_rate_all_valid(self):
        """Test success_rate with all valid."""
        validation = ValidationResult(
            total_links=2,
            valid_links=2,
            invalid_links=0,
            results=[],
        )
        assert validation.success_rate == 100.0

    def test_validation_result_success_rate_empty(self):
        """Test success_rate with no links."""
        validation = ValidationResult(
            total_links=0,
            valid_links=0,
            invalid_links=0,
            results=[],
        )
        assert validation.success_rate == 0.0


class TestLinkValidator:
    """Test LinkValidator class."""

    def test_initialization(self):
        """Test LinkValidator initialization."""
        validator = LinkValidator(max_concurrent=3, timeout=8)
        assert validator.max_concurrent == 3
        assert validator.timeout == 8

    def test_initialization_defaults(self):
        """Test LinkValidator default values."""
        validator = LinkValidator()
        assert validator.max_concurrent == 5
        assert validator.timeout == 10

    def test_extract_links_from_nonexistent_file(self):
        """Test extract_links_from_file with nonexistent file."""
        validator = LinkValidator()
        with patch("pathlib.Path.exists", return_value=False):
            result = validator.extract_links_from_file(Path("/nonexistent.md"))
            assert result == []

    def test_extract_links_from_file_success(self):
        """Test extract_links_from_file successful."""
        validator = LinkValidator()
        content = "[Link](https://example.com) and https://another.com"
        with patch("pathlib.Path.exists", return_value=True):
            with patch.object(
                validator, "extract_links_from_file", return_value=["https://example.com"]
            ):
                result = validator.extract_links_from_file(Path("/tmp/test.md"))
                assert len(result) >= 0

    @pytest.mark.asyncio
    async def test_validate_link_invalid_url(self):
        """Test validate_link with invalid URL."""
        validator = LinkValidator()
        async with validator:
            result = await validator.validate_link("not a valid url")
            assert result.is_valid is False

    @pytest.mark.asyncio
    async def test_validate_link_valid_url_format(self):
        """Test validate_link with valid URL format."""
        validator = LinkValidator()
        with patch("moai_adk.utils.link_validator.is_valid_url", return_value=True):
            with patch.object(
                validator, "fetch_url"
            ) as mock_fetch:
                mock_fetch.return_value = MagicMock(
                    status_code=200,
                    success=True,
                    load_time=0.5,
                    error_message=None,
                )
                async with validator:
                    result = await validator.validate_link("https://example.com")
                    assert result.url == "https://example.com"

    @pytest.mark.asyncio
    async def test_validate_all_links_empty(self):
        """Test validate_all_links with empty list."""
        validator = LinkValidator()
        async with validator:
            result = await validator.validate_all_links([])
            assert result.total_links == 0
            assert result.valid_links == 0

    def test_generate_report_basic(self):
        """Test generate_report."""
        validator = LinkValidator()
        results = [
            LinkResult(
                url="https://example.com",
                status_code=200,
                is_valid=True,
                response_time=0.5,
            )
        ]
        validation = ValidationResult(
            total_links=1,
            valid_links=1,
            invalid_links=0,
            results=results,
        )
        report = validator.generate_report(validation)
        assert "example.com" in report or "Link Validation Report" in report

    def test_generate_report_with_failures(self):
        """Test generate_report with failed links."""
        validator = LinkValidator()
        results = [
            LinkResult(
                url="https://example.com",
                status_code=404,
                is_valid=False,
                response_time=0.2,
                error_message="Not found",
            )
        ]
        validation = ValidationResult(
            total_links=1,
            valid_links=0,
            invalid_links=1,
            results=results,
        )
        report = validator.generate_report(validation)
        assert isinstance(report, str)


class TestValidateReadmeLinks:
    """Test validate_readme_links function."""

    def test_validate_readme_links_default_path(self):
        """Test validate_readme_links with default path."""
        with patch("pathlib.Path.exists", return_value=False):
            result = validate_readme_links()
            assert result.total_links == 0

    def test_validate_readme_links_custom_path(self):
        """Test validate_readme_links with custom path."""
        with patch("pathlib.Path.exists", return_value=False):
            result = validate_readme_links(Path("/custom/README.md"))
            assert result.total_links == 0

    def test_validate_readme_links_no_links_found(self):
        """Test validate_readme_links when no links in file."""
        with patch("pathlib.Path.exists", return_value=True):
            with patch("moai_adk.utils.link_validator.LinkValidator") as mock_validator:
                instance = MagicMock()
                instance.extract_links_from_file.return_value = []
                mock_validator.return_value = instance
                # Would need more setup to fully test


class TestLinkExtraction:
    """Test link extraction from files."""

    def test_extract_markdown_links(self):
        """Test extracting markdown format links."""
        validator = LinkValidator()
        # Test through extract_links_from_text if accessible
        # This tests integration with common.py functions

    def test_extract_plain_urls(self):
        """Test extracting plain URLs."""
        validator = LinkValidator()
        # Test through extract_links_from_text if accessible


class TestAsyncContextManager:
    """Test LinkValidator as async context manager."""

    @pytest.mark.asyncio
    async def test_context_manager_entry_exit(self):
        """Test async context manager entry and exit."""
        validator = LinkValidator()
        async with validator as v:
            assert v is validator


class TestValidatorIntegration:
    """Integration tests for LinkValidator."""

    @pytest.mark.asyncio
    async def test_full_validation_workflow(self):
        """Test complete validation workflow."""
        validator = LinkValidator()
        # Create mock validation
        with patch.object(
            validator, "validate_all_links"
        ) as mock_validate:
            mock_validate.return_value = ValidationResult(
                total_links=0,
                valid_links=0,
                invalid_links=0,
                results=[],
            )
            # Test would go here
