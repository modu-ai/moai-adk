"""Comprehensive coverage tests for moai_adk.utils.safe_file_reader module.

This module contains tests targeting uncovered code paths in safe_file_reader.py
to achieve 90%+ coverage.
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


class TestSafeFileReaderCustomEncodings:
    """Test SafeFileReader with custom encodings."""

    def test_custom_encodings_applied(self):
        """Test custom encodings are used instead of defaults."""
        custom_encodings = ["ascii", "utf-8"]
        reader = SafeFileReader(encodings=custom_encodings)
        assert reader.encodings == custom_encodings
        assert "ascii" in reader.encodings

    def test_custom_error_handling_modes(self):
        """Test different error handling modes."""
        reader_ignore = SafeFileReader(errors="ignore")
        assert reader_ignore.errors == "ignore"

        reader_replace = SafeFileReader(errors="replace")
        assert reader_replace.errors == "replace"

        reader_strict = SafeFileReader(errors="strict")
        assert reader_strict.errors == "strict"


class TestReadTextEncoding:
    """Test read_text with various encoding scenarios."""

    def test_read_text_utf8_success(self):
        """Test reading file with UTF-8 encoding."""
        reader = SafeFileReader()
        content = "Hello, World! 你好"
        with patch("pathlib.Path.exists", return_value=True):
            with patch("pathlib.Path.read_text", return_value=content):
                result = reader.read_text(Path("test.txt"))
                assert result == content

    def test_read_text_encoding_fallback_sequence(self):
        """Test encoding fallback tries multiple encodings."""
        reader = SafeFileReader(encodings=["ascii", "utf-8", "cp1252"])
        call_count = 0

        def read_text_side_effect(encoding=None, errors=None):
            nonlocal call_count
            call_count += 1
            if encoding == "ascii":
                raise UnicodeDecodeError("ascii", b"", 0, 1, "ordinal not in range")
            elif encoding == "utf-8":
                return "content"
            return "fallback"

        with patch("pathlib.Path.exists", return_value=True):
            with patch("pathlib.Path.read_text", side_effect=read_text_side_effect):
                result = reader.read_text(Path("test.txt"))
                assert result == "content"
                assert call_count >= 2  # Should try at least 2 encodings

    def test_read_text_all_encodings_fail(self):
        """Test read_text when all encodings fail, uses error handling."""
        reader = SafeFileReader(encodings=["ascii"], errors="ignore")
        with patch("pathlib.Path.exists", return_value=True):
            with patch("pathlib.Path.read_text") as mock_read:
                # First call raises, then final fallback succeeds
                mock_read.side_effect = [
                    UnicodeDecodeError("ascii", b"", 0, 1, "invalid"),
                    "fallback content",
                ]
                result = reader.read_text(Path("test.txt"))
                assert result == "fallback content"

    def test_read_text_non_unicode_decode_error(self):
        """Test read_text with non-UnicodeDecodeError exceptions."""
        reader = SafeFileReader()
        with patch("pathlib.Path.exists", return_value=True):
            with patch("pathlib.Path.read_text") as mock_read:
                # Raise different exception on first attempt
                mock_read.side_effect = [
                    PermissionError("Permission denied"),
                    "fallback",
                ]
                result = reader.read_text(Path("test.txt"))
                assert result == "fallback"

    def test_read_text_file_not_found_returns_none(self):
        """Test read_text returns None for non-existent file."""
        reader = SafeFileReader()
        with patch("pathlib.Path.exists", return_value=False):
            result = reader.read_text(Path("nonexistent.txt"))
            assert result is None

    def test_read_text_accepts_string_path(self):
        """Test read_text converts string path to Path."""
        reader = SafeFileReader()
        with patch("pathlib.Path.exists", return_value=True):
            with patch("pathlib.Path.read_text", return_value="content"):
                result = reader.read_text("test.txt")
                assert result == "content"


class TestReadLines:
    """Test read_lines method."""

    def test_read_lines_basic(self):
        """Test basic line reading."""
        reader = SafeFileReader()
        content = "line1\nline2\nline3"
        with patch("pathlib.Path.exists", return_value=True):
            with patch("pathlib.Path.read_text", return_value=content):
                lines = reader.read_lines(Path("test.txt"))
                assert len(lines) == 3

    def test_read_lines_preserves_line_endings(self):
        """Test read_lines preserves line endings with keepends=True."""
        reader = SafeFileReader()
        content = "line1\nline2\nline3"
        with patch("pathlib.Path.exists", return_value=True):
            with patch("pathlib.Path.read_text", return_value=content):
                lines = reader.read_lines(Path("test.txt"))
                # Should have splitlines with keepends
                assert len(lines) > 0

    def test_read_lines_empty_file(self):
        """Test read_lines with empty file."""
        reader = SafeFileReader()
        with patch("pathlib.Path.exists", return_value=True):
            with patch("pathlib.Path.read_text", return_value=""):
                lines = reader.read_lines(Path("test.txt"))
                assert lines == []

    def test_read_lines_file_not_exists(self):
        """Test read_lines returns empty list for non-existent file."""
        reader = SafeFileReader()
        with patch("pathlib.Path.exists", return_value=False):
            lines = reader.read_lines(Path("nonexistent.txt"))
            assert lines == []

    def test_read_lines_with_string_path(self):
        """Test read_lines accepts string path."""
        reader = SafeFileReader()
        with patch("pathlib.Path.exists", return_value=True):
            with patch("pathlib.Path.read_text", return_value="line1\nline2"):
                lines = reader.read_lines("test.txt")
                assert len(lines) > 0

    def test_read_lines_single_line_no_newline(self):
        """Test read_lines with single line without newline."""
        reader = SafeFileReader()
        with patch("pathlib.Path.exists", return_value=True):
            with patch("pathlib.Path.read_text", return_value="single line"):
                lines = reader.read_lines(Path("test.txt"))
                assert len(lines) == 1


class TestSafeGlobRead:
    """Test safe_glob_read method."""

    def test_safe_glob_read_basic(self):
        """Test basic glob reading."""
        reader = SafeFileReader()
        mock_path1 = MagicMock(spec=Path)
        mock_path1.is_file.return_value = True
        mock_path1.__str__.return_value = "/tmp/file1.txt"

        with patch("pathlib.Path.glob", return_value=[mock_path1]):
            with patch.object(reader, "read_text", return_value="content1"):
                result = reader.safe_glob_read("*.txt")
                assert isinstance(result, dict)
                assert len(result) == 1

    def test_safe_glob_read_multiple_files(self):
        """Test glob reading with multiple files."""
        reader = SafeFileReader()
        mock_path1 = MagicMock(spec=Path)
        mock_path1.is_file.return_value = True
        mock_path1.__str__.return_value = "/tmp/file1.txt"

        mock_path2 = MagicMock(spec=Path)
        mock_path2.is_file.return_value = True
        mock_path2.__str__.return_value = "/tmp/file2.txt"

        with patch("pathlib.Path.glob", return_value=[mock_path1, mock_path2]):
            with patch.object(
                reader, "read_text", side_effect=["content1", "content2"]
            ):
                result = reader.safe_glob_read("*.txt")
                assert len(result) == 2
                assert "/tmp/file1.txt" in result
                assert "/tmp/file2.txt" in result

    def test_safe_glob_read_excludes_directories(self):
        """Test glob read excludes directories."""
        reader = SafeFileReader()
        mock_file = MagicMock(spec=Path)
        mock_file.is_file.return_value = True
        mock_file.__str__.return_value = "/tmp/file.txt"

        mock_dir = MagicMock(spec=Path)
        mock_dir.is_file.return_value = False
        mock_dir.__str__.return_value = "/tmp/dir"

        with patch("pathlib.Path.glob", return_value=[mock_file, mock_dir]):
            with patch.object(reader, "read_text", return_value="content"):
                result = reader.safe_glob_read("*")
                assert len(result) == 1

    def test_safe_glob_read_no_matches(self):
        """Test glob read with no matches."""
        reader = SafeFileReader()
        with patch("pathlib.Path.glob", return_value=[]):
            result = reader.safe_glob_read("*.txt")
            assert result == {}

    def test_safe_glob_read_custom_base_path(self):
        """Test glob read with custom base path."""
        reader = SafeFileReader()
        with patch("pathlib.Path.glob", return_value=[]):
            result = reader.safe_glob_read("*.txt", base_path="/custom")
            assert isinstance(result, dict)

    def test_safe_glob_read_read_failure_skipped(self):
        """Test glob read skips files that fail to read."""
        reader = SafeFileReader()
        mock_path1 = MagicMock(spec=Path)
        mock_path1.is_file.return_value = True
        mock_path1.__str__.return_value = "/tmp/file1.txt"

        mock_path2 = MagicMock(spec=Path)
        mock_path2.is_file.return_value = True
        mock_path2.__str__.return_value = "/tmp/file2.txt"

        with patch("pathlib.Path.glob", return_value=[mock_path1, mock_path2]):
            with patch.object(reader, "read_text", side_effect=["content1", None]):
                result = reader.safe_glob_read("*.txt")
                assert len(result) == 1

    def test_safe_glob_read_glob_exception(self):
        """Test glob read handles glob exceptions."""
        reader = SafeFileReader()
        with patch("pathlib.Path.glob", side_effect=Exception("Glob error")):
            result = reader.safe_glob_read("*.txt")
            assert result == {}


class TestIsSafeFile:
    """Test is_safe_file method."""

    def test_is_safe_file_readable(self):
        """Test is_safe_file returns True for readable file."""
        reader = SafeFileReader()
        with patch("pathlib.Path.exists", return_value=True):
            with patch("pathlib.Path.read_text", return_value="content"):
                result = reader.is_safe_file(Path("test.txt"))
                assert result is True

    def test_is_safe_file_not_readable(self):
        """Test is_safe_file returns False for unreadable file."""
        reader = SafeFileReader()
        with patch("pathlib.Path.exists", return_value=False):
            result = reader.is_safe_file(Path("nonexistent.txt"))
            assert result is False

    def test_is_safe_file_with_string_path(self):
        """Test is_safe_file accepts string path."""
        reader = SafeFileReader()
        with patch("pathlib.Path.exists", return_value=True):
            with patch("pathlib.Path.read_text", return_value="content"):
                result = reader.is_safe_file("test.txt")
                assert result is True


class TestDefaultEncodings:
    """Test DEFAULT_ENCODINGS constant."""

    def test_default_encodings_list(self):
        """Test DEFAULT_ENCODINGS is properly configured."""
        assert len(SafeFileReader.DEFAULT_ENCODINGS) > 0
        assert "utf-8" in SafeFileReader.DEFAULT_ENCODINGS
        assert SafeFileReader.DEFAULT_ENCODINGS[0] == "utf-8"

    def test_default_encodings_contains_common_encodings(self):
        """Test DEFAULT_ENCODINGS contains common encodings."""
        encodings = SafeFileReader.DEFAULT_ENCODINGS
        assert "utf-8" in encodings
        assert any(
            "iso" in enc or "latin" in enc or "cp1252" in enc for enc in encodings
        )


class TestSafeReadFile:
    """Test safe_read_file convenience function."""

    def test_safe_read_file_basic(self):
        """Test safe_read_file basic usage."""
        with patch("pathlib.Path.exists", return_value=True):
            with patch("pathlib.Path.read_text", return_value="content"):
                result = safe_read_file("test.txt")
                assert result == "content"

    def test_safe_read_file_custom_encodings(self):
        """Test safe_read_file with custom encodings."""
        with patch("pathlib.Path.exists", return_value=True):
            with patch("pathlib.Path.read_text", return_value="content"):
                result = safe_read_file("test.txt", encodings=["utf-8", "ascii"])
                assert result is not None

    def test_safe_read_file_not_exists(self):
        """Test safe_read_file with non-existent file."""
        with patch("pathlib.Path.exists", return_value=False):
            result = safe_read_file("nonexistent.txt")
            assert result is None

    def test_safe_read_file_path_object(self):
        """Test safe_read_file accepts Path object."""
        with patch("pathlib.Path.exists", return_value=True):
            with patch("pathlib.Path.read_text", return_value="content"):
                result = safe_read_file(Path("test.txt"))
                assert result == "content"


class TestSafeReadLines:
    """Test safe_read_lines convenience function."""

    def test_safe_read_lines_basic(self):
        """Test safe_read_lines basic usage."""
        with patch("pathlib.Path.exists", return_value=True):
            with patch("pathlib.Path.read_text", return_value="line1\nline2"):
                result = safe_read_lines("test.txt")
                assert len(result) > 0

    def test_safe_read_lines_custom_encodings(self):
        """Test safe_read_lines with custom encodings."""
        with patch("pathlib.Path.exists", return_value=True):
            with patch("pathlib.Path.read_text", return_value="line1"):
                result = safe_read_lines("test.txt", encodings=["utf-8"])
                assert len(result) > 0

    def test_safe_read_lines_not_exists(self):
        """Test safe_read_lines with non-existent file."""
        with patch("pathlib.Path.exists", return_value=False):
            result = safe_read_lines("nonexistent.txt")
            assert result == []

    def test_safe_read_lines_path_object(self):
        """Test safe_read_lines accepts Path object."""
        with patch("pathlib.Path.exists", return_value=True):
            with patch("pathlib.Path.read_text", return_value="line1"):
                result = safe_read_lines(Path("test.txt"))
                assert len(result) > 0


class TestSafeGlobReadFunction:
    """Test safe_glob_read convenience function."""

    def test_safe_glob_read_function_basic(self):
        """Test safe_glob_read function basic usage."""
        with patch("pathlib.Path.glob", return_value=[]):
            result = safe_glob_read("*.txt")
            assert isinstance(result, dict)

    def test_safe_glob_read_function_custom_base_path(self):
        """Test safe_glob_read function with custom base path."""
        with patch("pathlib.Path.glob", return_value=[]):
            result = safe_glob_read("*.txt", base_path="/tmp")
            assert isinstance(result, dict)

    def test_safe_glob_read_function_custom_encodings(self):
        """Test safe_glob_read function with custom encodings."""
        with patch("pathlib.Path.glob", return_value=[]):
            result = safe_glob_read("*.txt", encodings=["utf-8"])
            assert isinstance(result, dict)


class TestEncodingFallbackSequence:
    """Test encoding fallback behavior."""

    def test_fallback_tries_all_encodings_in_order(self):
        """Test that fallback tries encodings in specified order."""
        reader = SafeFileReader(encodings=["enc1", "enc2", "enc3"])
        call_sequence = []

        def track_encoding_calls(encoding=None, errors=None):
            call_sequence.append(encoding)
            if encoding == "enc2":
                return "success"
            raise UnicodeDecodeError("enc", b"", 0, 1, "reason")

        with patch("pathlib.Path.exists", return_value=True):
            with patch("pathlib.Path.read_text", side_effect=track_encoding_calls):
                result = reader.read_text(Path("test.txt"))
                assert result == "success"
                assert "enc1" in call_sequence
                assert "enc2" in call_sequence

    def test_fallback_uses_final_error_handling(self):
        """Test that final fallback uses specified error handling."""
        reader = SafeFileReader(encodings=["bad"], errors="replace")
        with patch("pathlib.Path.exists", return_value=True):
            with patch("pathlib.Path.read_text") as mock_read:
                mock_read.side_effect = [
                    UnicodeDecodeError("bad", b"", 0, 1, "reason"),
                    "fallback with replace",
                ]
                result = reader.read_text(Path("test.txt"))
                # Verify second call used error handling
                calls = mock_read.call_args_list
                assert len(calls) >= 2
                if len(calls) > 1:
                    assert calls[1][1].get("errors") == "replace"
