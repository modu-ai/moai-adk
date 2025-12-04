"""Unit tests for moai_adk.utils.safe_file_reader module.

Tests for safe file reading with encoding fallbacks.
"""

from pathlib import Path
from unittest.mock import MagicMock, patch

import pytest

from moai_adk.utils.safe_file_reader import (
    SafeFileReader,
    safe_glob_read,
    safe_read_file,
    safe_read_lines,
)


class TestSafeFileReaderInitialization:
    """Test SafeFileReader initialization."""

    def test_initialization_default_encodings(self):
        """Test initialization with default encodings."""
        reader = SafeFileReader()
        assert len(reader.encodings) > 0
        assert "utf-8" in reader.encodings
        assert reader.errors == "ignore"

    def test_initialization_custom_encodings(self):
        """Test initialization with custom encodings."""
        custom_encodings = ["utf-8", "ascii"]
        reader = SafeFileReader(encodings=custom_encodings)
        assert reader.encodings == custom_encodings

    def test_initialization_custom_error_handling(self):
        """Test initialization with custom error handling."""
        reader = SafeFileReader(errors="replace")
        assert reader.errors == "replace"

    def test_default_encodings_order(self):
        """Test default encodings are in correct order."""
        reader = SafeFileReader()
        assert reader.encodings[0] == "utf-8"


class TestReadText:
    """Test read_text method."""

    def test_read_text_file_not_exists(self):
        """Test read_text when file doesn't exist."""
        reader = SafeFileReader()
        with patch("pathlib.Path.exists", return_value=False):
            result = reader.read_text(Path("/nonexistent/file.txt"))
            assert result is None

    def test_read_text_successful(self):
        """Test successful text reading."""
        reader = SafeFileReader()
        test_content = "Hello, World!"
        with patch("pathlib.Path.exists", return_value=True):
            with patch("pathlib.Path.read_text", return_value=test_content):
                result = reader.read_text(Path("/tmp/test.txt"))
                assert result == test_content

    def test_read_text_with_string_path(self):
        """Test read_text accepts string path."""
        reader = SafeFileReader()
        with patch("pathlib.Path.exists", return_value=True):
            with patch("pathlib.Path.read_text", return_value="content"):
                result = reader.read_text("/tmp/test.txt")
                assert result is not None

    def test_read_text_encoding_fallback(self):
        """Test encoding fallback mechanism."""
        reader = SafeFileReader(encodings=["ascii", "utf-8"])
        with patch("pathlib.Path.exists", return_value=True):
            # Mock first encoding to fail, second to succeed
            def read_text_side_effect(encoding=None, errors=None):
                if encoding == "ascii":
                    raise UnicodeDecodeError("ascii", b"", 0, 1, "invalid")
                return "content"

            with patch("pathlib.Path.read_text", side_effect=read_text_side_effect):
                result = reader.read_text(Path("/tmp/test.txt"))
                assert result == "content"


class TestReadLines:
    """Test read_lines method."""

    def test_read_lines_success(self):
        """Test successful line reading."""
        reader = SafeFileReader()
        test_content = "line1\nline2\nline3"
        with patch("pathlib.Path.exists", return_value=True):
            with patch("pathlib.Path.read_text", return_value=test_content):
                result = reader.read_lines(Path("/tmp/test.txt"))
                assert len(result) == 3

    def test_read_lines_file_not_exists(self):
        """Test read_lines when file doesn't exist."""
        reader = SafeFileReader()
        with patch("pathlib.Path.exists", return_value=False):
            result = reader.read_lines(Path("/nonexistent/file.txt"))
            assert result == []

    def test_read_lines_with_string_path(self):
        """Test read_lines accepts string path."""
        reader = SafeFileReader()
        with patch("pathlib.Path.exists", return_value=True):
            with patch("pathlib.Path.read_text", return_value="line1\nline2"):
                result = reader.read_lines("/tmp/test.txt")
                assert len(result) == 2


class TestSafeGlobRead:
    """Test safe_glob_read method."""

    def test_safe_glob_read_success(self):
        """Test successful glob read."""
        reader = SafeFileReader()
        mock_path = MagicMock(spec=Path)
        mock_path.is_file.return_value = True

        with patch("pathlib.Path.glob", return_value=[mock_path]):
            with patch("pathlib.Path.exists", return_value=True):
                with patch.object(reader, "read_text", return_value="content"):
                    result = reader.safe_glob_read("*.txt")
                    assert isinstance(result, dict)

    def test_safe_glob_read_no_matches(self):
        """Test glob read with no matches."""
        reader = SafeFileReader()
        with patch("pathlib.Path.glob", return_value=[]):
            result = reader.safe_glob_read("*.txt")
            assert result == {}

    def test_safe_glob_read_with_base_path(self):
        """Test glob read with custom base path."""
        reader = SafeFileReader()
        with patch("pathlib.Path.glob", return_value=[]):
            result = reader.safe_glob_read("*.txt", base_path="/tmp")
            assert isinstance(result, dict)


class TestIsSafeFile:
    """Test is_safe_file method."""

    def test_is_safe_file_readable(self):
        """Test is_safe_file with readable file."""
        reader = SafeFileReader()
        with patch("pathlib.Path.exists", return_value=True):
            with patch("pathlib.Path.read_text", return_value="content"):
                result = reader.is_safe_file(Path("/tmp/test.txt"))
                assert result is True

    def test_is_safe_file_not_readable(self):
        """Test is_safe_file with unreadable file."""
        reader = SafeFileReader()
        with patch("pathlib.Path.exists", return_value=False):
            result = reader.is_safe_file(Path("/tmp/test.txt"))
            assert result is False


class TestGlobalFunctions:
    """Test global convenience functions."""

    def test_safe_read_file(self):
        """Test safe_read_file function."""
        with patch("pathlib.Path.exists", return_value=True):
            with patch("pathlib.Path.read_text", return_value="content"):
                result = safe_read_file("/tmp/test.txt")
                assert result == "content"

    def test_safe_read_file_with_custom_encodings(self):
        """Test safe_read_file with custom encodings."""
        with patch("pathlib.Path.exists", return_value=True):
            with patch("pathlib.Path.read_text", return_value="content"):
                result = safe_read_file(
                    "/tmp/test.txt", encodings=["utf-8", "ascii"]
                )
                assert result is not None

    def test_safe_read_lines(self):
        """Test safe_read_lines function."""
        with patch("pathlib.Path.exists", return_value=True):
            with patch("pathlib.Path.read_text", return_value="line1\nline2"):
                result = safe_read_lines("/tmp/test.txt")
                assert len(result) == 2

    def test_safe_read_lines_with_custom_encodings(self):
        """Test safe_read_lines with custom encodings."""
        with patch("pathlib.Path.exists", return_value=True):
            with patch("pathlib.Path.read_text", return_value="line1"):
                result = safe_read_lines(
                    "/tmp/test.txt", encodings=["utf-8"]
                )
                assert len(result) >= 1

    def test_safe_glob_read_function(self):
        """Test safe_glob_read global function."""
        with patch("pathlib.Path.glob", return_value=[]):
            result = safe_glob_read("*.txt")
            assert isinstance(result, dict)

    def test_safe_glob_read_function_with_base_path(self):
        """Test safe_glob_read with base path."""
        with patch("pathlib.Path.glob", return_value=[]):
            result = safe_glob_read("*.txt", base_path="/tmp")
            assert isinstance(result, dict)


class TestEncodingPriority:
    """Test encoding priority system."""

    def test_utf8_priority(self):
        """Test UTF-8 is first priority."""
        reader = SafeFileReader()
        assert reader.encodings[0] == "utf-8"

    def test_multiple_encodings_attempted(self):
        """Test multiple encodings are in the list."""
        reader = SafeFileReader()
        assert len(reader.encodings) > 2
        assert "cp1252" in reader.encodings or "iso-8859-1" in reader.encodings
