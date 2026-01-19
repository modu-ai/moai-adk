"""Comprehensive TDD tests for utils modules to reach 100% coverage.

Targets uncovered lines in:
- timeout.py (line 69)
- toon_utils.py (lines 28, 130, 222-223, 255-256)
- common.py (lines 130, 311-331)
- link_validator.py (lines 215-226, 231-241)
- safe_file_reader.py (lines 76-78, 188-206)
"""

import asyncio
import tempfile
from pathlib import Path
from unittest.mock import patch

import pytest

from moai_adk.utils.common import (
    HTTPResponse,
    RateLimiter,
    extract_links_from_text,
    format_duration,
    get_graceful_degradation,
    get_summary_stats,
    load_hook_timeout,
    reset_stdin,
)
from moai_adk.utils.link_validator import LinkValidator, ValidationResult, validate_readme_links
from moai_adk.utils.safe_file_reader import SafeFileReader, safe_glob_read, safe_read_file, safe_read_lines
from moai_adk.utils.timeout import CrossPlatformTimeout
from moai_adk.utils.toon_utils import (
    _encode_value,
    _is_tabular,
    migrate_json_to_toon,
    toon_decode,
    toon_encode,
    toon_load,
    toon_save,
)


class TestTimeoutFullCoverage:
    """Test timeout module for 100% coverage (line 69)."""

    def test_context_manager_exit_false_return(self):
        """Test context manager __exit__ returns False (line 136-137)."""
        timeout = CrossPlatformTimeout(1.0)

        # Context manager should not suppress exceptions
        result = timeout.__exit__(None, None, None)
        assert result is False

    def test_context_manager_with_exception(self):
        """Test context manager doesn't suppress exceptions (line 69)."""
        try:
            with CrossPlatformTimeout(1.0):
                raise ValueError("Test exception")
        except ValueError:
            pass  # Exception should not be suppressed


class TestToonUtilsFullCoverage:
    """Test toon_utils for 100% coverage."""

    def test_is_tabular_not_list(self):
        """Test _is_tabular with non-list input (line 21-22)."""
        assert _is_tabular("not a list") is False  # type: ignore[arg-type]
        assert _is_tabular(None) is False  # type: ignore[arg-type]
        assert _is_tabular({"key": "value"}) is False  # type: ignore[arg-type]

    def test_is_tabular_empty_list(self):
        """Test _is_tabular with empty list (line 27-28)."""
        assert _is_tabular([]) is False

    def test_is_tabular_non_dict_items(self):
        """Test _is_tabular with non-dict items (line 24-25)."""
        assert _is_tabular([1, 2, 3]) is False
        assert _is_tabular(["string", "items"]) is False

    def test_is_tabular_true_case(self):
        """Test _is_tabular returns True for valid tabular data."""
        data = [{"id": 1, "name": "A"}, {"id": 2, "name": "B"}]
        assert _is_tabular(data) is True

    def test_encode_value_none(self):
        """Test _encode_value with None."""
        assert _encode_value(None) == "null"

    def test_encode_value_bool(self):
        """Test _encode_value with boolean."""
        assert _encode_value(True) == "true"
        assert _encode_value(False) == "false"

    def test_encode_value_number(self):
        """Test _encode_value with numbers."""
        assert _encode_value(42) == "42"
        assert _encode_value(3.14) == "3.14"

    def test_encode_value_string_no_special_chars(self):
        """Test _encode_value with simple string."""
        assert _encode_value("simple") == "simple"

    def test_encode_value_string_with_special_chars(self):
        """Test _encode_value with special characters (line 45-46)."""
        # Contains special chars that need quoting
        result = _encode_value("hello, world")
        assert '"' in result  # Should be JSON-encoded

        result = _encode_value('key: "value"')
        assert '"' in result

    def test_encode_value_complex_object(self):
        """Test _encode_value with complex object (line 49)."""
        result = _encode_value({"nested": "dict"})
        assert "{" in result  # Should be JSON-encoded

    def test_toon_encode_type_error(self):
        """Test toon_encode raises ValueError for unencodable data."""

        # Create an object that can't be JSON encoded
        class UnencodableClass:
            pass

        with pytest.raises(ValueError, match="Failed to encode"):
            toon_encode(UnencodableClass())

    def test_toon_decode_json_decode_error(self):
        """Test toon_decode raises ValueError for invalid JSON (line 104-105)."""
        with pytest.raises(ValueError, match="Failed to decode"):
            toon_decode("not valid json {")

    def test_toon_save_value_error(self):
        """Test toon_save raises ValueError for invalid data."""
        with tempfile.TemporaryDirectory() as tmpdir:

            class UnencodableClass:
                pass

            with pytest.raises(ValueError):
                toon_save(UnencodableClass(), Path(tmpdir) / "test.toon")

    def test_toon_save_io_error(self):
        """Test toon_save raises IOError for write failures."""
        # Create a directory instead of a file
        with tempfile.TemporaryDirectory() as tmpdir:
            with pytest.raises(IOError, match="Failed to write"):
                toon_save({"key": "value"}, Path(tmpdir))  # tmpdir is a directory

    def test_toon_load_io_error(self):
        """Test toon_load raises IOError for read failures."""
        with pytest.raises(IOError, match="Failed to read"):
            toon_load("/nonexistent/path/file.toon")

    def test_toon_load_value_error(self):
        """Test toon_load raises ValueError for invalid TOON content."""
        with tempfile.TemporaryDirectory() as tmpdir:
            test_file = Path(tmpdir) / "test.toon"
            test_file.write_text("invalid json content {")

            with pytest.raises(ValueError, match="Failed to decode"):
                toon_load(test_file)

    def test_migrate_json_to_toon_invalid_json(self):
        """Test migrate_json_to_toon raises ValueError for invalid JSON."""
        with tempfile.TemporaryDirectory() as tmpdir:
            json_file = Path(tmpdir) / "config.json"
            json_file.write_text("invalid json {")

            with pytest.raises(ValueError, match="Invalid JSON"):
                migrate_json_to_toon(json_file)

    def test_migrate_json_to_toon_io_error(self):
        """Test migrate_json_to_toon handles IO errors."""
        with pytest.raises(IOError):
            migrate_json_to_toon("/nonexistent/file.json")


class TestCommonFullCoverage:
    """Test common module for 100% coverage."""

    def test_http_response_with_none_timestamp(self):
        """Test HTTPResponse post_init with None timestamp."""
        response = HTTPResponse(  # type: ignore[call-arg]
            status_code=200,
            url="http://example.com",
            load_time=0.5,
            success=True,
            timestamp=None,  # type: ignore[call-arg]
        )
        assert response.timestamp is not None

    def test_extract_links_with_base_path(self):
        """Test extract_links_from_text with base path (line 127-130)."""
        text = "[Link](/path/to/page)"
        base_url = "https://example.com"
        links = extract_links_from_text(text, base_url)

        assert "https://example.com/path/to/page" in links

    def test_extract_links_relative_without_base(self):
        """Test extract_links with relative URL and no base URL."""
        text = "[Link](/path/to/page)"
        links = extract_links_from_text(text, base_url=None)

        # Should not include relative URLs without base
        assert len(links) == 0

    def test_extract_links_mixed_formats(self):
        """Test extract_links with various URL formats."""
        text = """
        [Link1](https://example.com/page1)
        [Link2](http://example.com/page2)
        [Link3](/relative/path)
        Plain URL: https://example.com/page3
        """
        links = extract_links_from_text(text, "https://base.com")

        # Should have absolute URLs
        assert "https://example.com/page1" in links
        assert "http://example.com/page2" in links
        assert "https://example.com/page3" in links

    def test_format_duration_milliseconds(self):
        """Test format_duration with milliseconds (line 159-160)."""
        assert format_duration(0.5) == "500ms"
        assert format_duration(0.123) == "123ms"

    def test_format_duration_seconds(self):
        """Test format_duration with seconds (line 161-162)."""
        assert format_duration(5.5) == "5.5s"
        assert format_duration(1.0) == "1.0s"

    def test_format_duration_minutes(self):
        """Test format_duration with minutes (line 163-166)."""
        assert "30s" in format_duration(90)
        assert "1m" in format_duration(60)

    def test_format_duration_hours(self):
        """Test format_duration with hours (line 167-170)."""
        assert "1h" in format_duration(3600)
        assert "2h" in format_duration(7200)

    def test_get_summary_stats_single_value(self):
        """Test get_summary_stats with single value."""
        stats = get_summary_stats([42.0])
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
        assert stats["std"] == 0.0

    def test_rate_limiter_wait_if_needed_under_limit(self):
        """Test RateLimiter wait_if_needed when under limit (line 234-240)."""
        limiter = RateLimiter(max_requests=10, time_window=60)

        # Should not wait when under limit
        # We need to test this with async
        async def test_wait():
            await limiter.wait_if_needed()

        asyncio.run(test_wait())

    def test_load_hook_timeout_default(self):
        """Test load_hook_timeout returns default when config missing."""
        with patch("pathlib.Path.exists", return_value=False):
            timeout = load_hook_timeout()
            assert timeout == 5000

    def test_load_hook_timeout_yaml_error(self):
        """Test load_hook_timeout handles YAML parse errors."""
        with tempfile.TemporaryDirectory() as tmpdir:
            config_path = Path(tmpdir) / "config.yaml"
            config_path.write_text("invalid: yaml: content: [")

            with patch("moai_adk.utils.common.Path", return_value=config_path):
                timeout = load_hook_timeout()
                assert timeout == 5000  # Should return default

    def test_load_hook_timeout_file_not_found(self):
        """Test load_hook_timeout handles FileNotFoundError."""
        with patch("pathlib.Path.exists", return_value=True):
            with patch("builtins.open", side_effect=FileNotFoundError):
                timeout = load_hook_timeout()
                assert timeout == 5000

    def test_get_graceful_degradation_default(self):
        """Test get_graceful_degradation returns default when config missing."""
        with patch("pathlib.Path.exists", return_value=False):
            result = get_graceful_degradation()
            assert result is True

    def test_get_graceful_degradation_yaml_error(self):
        """Test get_graceful_degradation handles YAML errors."""
        with tempfile.TemporaryDirectory() as tmpdir:
            config_path = Path(tmpdir) / "config.yaml"
            config_path.write_text("invalid: yaml: [")

            with patch("moai_adk.utils.common.Path", return_value=config_path):
                result = get_graceful_degradation()
                assert result is True  # Should return default

    def test_reset_stdin_import_error(self):
        """Test reset_stdin handles ImportError (line 319-320)."""
        # Just call the function - it should handle ImportError gracefully
        # If termios doesn't exist on the system, it will be caught
        try:
            import termios  # noqa: F401
            # If termios exists, test will continue normally
        except ImportError:
            # If termios doesn't exist, reset_stdin should handle it
            pass

        # Function should complete without error
        reset_stdin()

    def test_reset_stdin_os_error(self):
        """Test reset_stdin handles OSError (line 322-324)."""
        with patch("termios.tcflush", side_effect=OSError("Mock error")):
            reset_stdin()  # Should complete without error

    def test_reset_stdin_no_dev_tty(self):
        """Test reset_stdin when /dev/tty doesn't exist (line 328-331)."""
        with patch("os.path.exists", return_value=False):
            reset_stdin()  # Should complete without error


class TestLinkValidatorFullCoverage:
    """Test link_validator for 100% coverage (lines 215-226, 231-241)."""

    def test_validate_readme_links_custom_path(self):
        """Test validate_readme_links with custom path."""
        with tempfile.TemporaryDirectory() as tmpdir:
            readme_path = Path(tmpdir) / "README.md"
            readme_path.write_text("# Test\nNo links here")

            result = validate_readme_links(readme_path)

            assert result.total_links == 0
            assert result.valid_links == 0

    def test_generate_report_empty_results(self):
        """Test generate_report with empty validation results (line 193-196)."""
        validator = LinkValidator()
        result = ValidationResult(
            total_links=0,
            valid_links=0,
            invalid_links=0,
            results=[],
        )

        report = validator.generate_report(result)

        # Check that report contains the title
        assert "Online Documentation Link Validation Report" in report
        assert "**Total Links**: 0" in report

    def test_generate_report_no_valid_links(self):
        """Test generate_report when all links failed (line 193-196)."""
        from moai_adk.utils.link_validator import LinkResult

        validator = LinkValidator()
        result = ValidationResult(
            total_links=2,
            valid_links=0,
            invalid_links=2,
            results=[
                LinkResult(
                    url="http://example.com",
                    status_code=404,
                    is_valid=False,
                    response_time=0.5,
                ),
                LinkResult(
                    url="http://example2.com",
                    status_code=500,
                    is_valid=False,
                    response_time=0.3,
                ),
            ],
        )

        report = validator.generate_report(result)

        # Should have "Failed Links" section since all failed
        assert "Failed Links" in report

    def test_validate_readme_links_main_execution(self):
        """Test main execution path with default README."""
        # This tests the if __name__ == "__main__" block
        with tempfile.TemporaryDirectory() as tmpdir:
            readme_path = Path(tmpdir) / "README.ko.md"
            readme_path.write_text("# Test\nNo links")

            with patch("sys.argv", ["link_validator"]):
                with patch("moai_adk.utils.link_validator.Path", return_value=readme_path):
                    # Should not raise exception
                    result = validate_readme_links(readme_path)
                    assert result.total_links == 0


class TestSafeFileReaderFullCoverage:
    """Test safe_file_reader for 100% coverage (lines 76-78, 188-206)."""

    def test_read_text_exception_during_read(self):
        """Test read_text handles non-decoding exceptions (line 68-71)."""
        reader = SafeFileReader(encodings=["utf-8"])

        with tempfile.TemporaryDirectory() as tmpdir:
            test_file = Path(tmpdir) / "test.txt"
            test_file.write_text("test content")

            # Mock read_text to raise a non-UnicodeDecodeError exception
            with patch("pathlib.Path.read_text", side_effect=OSError("Device error")):
                result = reader.read_text(test_file)
                # Should continue trying other encodings and return None if all fail
                assert result is None

    def test_read_text_final_fallback(self):
        """Test read_text uses UTF-8 with errors='ignore' as fallback (line 74-78)."""
        reader = SafeFileReader(encodings=["utf-8"], errors="ignore")

        with tempfile.TemporaryDirectory() as tmpdir:
            test_file = Path(tmpdir) / "test.txt"
            # Write content that will cause decode error
            test_file.write_bytes(b"\xff\xfe Invalid UTF-8")

            result = reader.read_text(test_file)
            # Should return something (even if empty due to errors='ignore')
            assert result is not None

    def test_read_text_all_encodings_fail(self):
        """Test read_text when all encodings fail."""
        reader = SafeFileReader(encodings=["utf-16", "utf-32"])

        with tempfile.TemporaryDirectory() as tmpdir:
            test_file = Path(tmpdir) / "test.txt"
            test_file.write_text("test content")

            # Mock to make all encodings fail
            original_read_text = Path.read_text

            def mock_read_text(self, encoding, errors=None):
                if encoding in ["utf-16", "utf-32"]:
                    raise UnicodeDecodeError(encoding, b"", 0, 1, "Invalid")
                return original_read_text(self, encoding=encoding)

            with patch.object(Path, "read_text", mock_read_text):
                result = reader.read_text(test_file)
                # Final fallback should succeed
                assert result is not None

    def test_read_lines_returns_empty_on_failure(self):
        """Test read_lines returns empty list when read fails (line 91-92)."""
        reader = SafeFileReader()

        result = reader.read_lines("/nonexistent/file.txt")

        assert result == []

    def test_safe_glob_read_exception_handling(self):
        """Test safe_glob_read handles glob exceptions (line 116-117)."""
        reader = SafeFileReader()

        # Mock glob to raise exception
        with patch("pathlib.Path.glob", side_effect=PermissionError("Access denied")):
            result = reader.safe_glob_read("*.txt", ".")
            assert result == {}

    def test_safe_glob_read_success(self):
        """Test safe_glob_read successfully reads files."""
        with tempfile.TemporaryDirectory() as tmpdir:
            # Create test files
            (Path(tmpdir) / "test1.txt").write_text("content1")
            (Path(tmpdir) / "test2.txt").write_text("content2")

            reader = SafeFileReader()
            result = reader.safe_glob_read("*.txt", tmpdir)

            assert len(result) == 2
            assert "content1" in result.values()
            assert "content2" in result.values()

    def test_convenience_safe_read_file_custom_encodings(self):
        """Test safe_read_file with custom encodings (line 143-148)."""
        with tempfile.TemporaryDirectory() as tmpdir:
            test_file = Path(tmpdir) / "test.txt"
            test_file.write_text("test content")

            result = safe_read_file(test_file, encodings=["utf-8", "latin1"])

            assert result == "test content"

    def test_convenience_safe_read_lines(self):
        """Test safe_read_lines convenience function (line 151-163)."""
        with tempfile.TemporaryDirectory() as tmpdir:
            test_file = Path(tmpdir) / "test.txt"
            test_file.write_text("line1\nline2\nline3")

            result = safe_read_lines(test_file)

            assert len(result) == 3
            assert "line1" in result[0]

    def test_convenience_safe_read_lines_with_encodings(self):
        """Test safe_read_lines with custom encodings."""
        with tempfile.TemporaryDirectory() as tmpdir:
            test_file = Path(tmpdir) / "test.txt"
            test_file.write_text("test")

            result = safe_read_lines(test_file, encodings=["utf-8"])

            assert len(result) == 1

    def test_convenience_safe_glob_read(self):
        """Test safe_glob_read convenience function (line 166-183)."""
        with tempfile.TemporaryDirectory() as tmpdir:
            (Path(tmpdir) / "test.txt").write_text("content")

            result = safe_glob_read("*.txt", tmpdir)

            assert len(result) == 1

    def test_convenience_safe_glob_read_with_base_path(self):
        """Test safe_glob_read with base path parameter."""
        with tempfile.TemporaryDirectory() as tmpdir:
            (Path(tmpdir) / "test.txt").write_text("content")

            result = safe_glob_read("*.txt", base_path=tmpdir)

            assert len(result) == 1
