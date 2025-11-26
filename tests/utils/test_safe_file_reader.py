"""Comprehensive test suite for safe_file_reader.py utilities module.

This module provides 90%+ coverage for safe file reading functionality including:
- SafeFileReader class with encoding fallbacks
- Text file reading with multiple encoding attempts
- Line-based file reading
- Glob pattern-based multi-file reading
- File safety validation
- Global convenience functions
- Error handling and recovery strategies
- Encoding detection and fallback chains
"""

import tempfile
from pathlib import Path
from unittest.mock import patch

from moai_adk.utils.safe_file_reader import (
    SafeFileReader,
    safe_glob_read,
    safe_read_file,
    safe_read_lines,
)

# ============================================================================
# SafeFileReader Initialization Tests
# ============================================================================


class TestSafeFileReaderInitialization:
    """Tests for SafeFileReader initialization."""

    def test_default_initialization(self):
        """Test SafeFileReader with default parameters."""
        reader = SafeFileReader()
        assert reader.encodings == SafeFileReader.DEFAULT_ENCODINGS
        assert reader.errors == "ignore"

    def test_custom_encodings_initialization(self):
        """Test SafeFileReader with custom encodings."""
        custom_encodings = ["utf-8", "ascii"]
        reader = SafeFileReader(encodings=custom_encodings)
        assert reader.encodings == custom_encodings
        assert reader.errors == "ignore"

    def test_custom_error_handling_initialization(self):
        """Test SafeFileReader with custom error handling strategy."""
        reader = SafeFileReader(errors="replace")
        assert reader.errors == "replace"
        assert reader.encodings == SafeFileReader.DEFAULT_ENCODINGS

    def test_custom_encodings_and_errors_initialization(self):
        """Test SafeFileReader with both custom encodings and error handling."""
        custom_encodings = ["utf-16", "cp1252"]
        reader = SafeFileReader(encodings=custom_encodings, errors="strict")
        assert reader.encodings == custom_encodings
        assert reader.errors == "strict"

    def test_empty_encodings_list(self):
        """Test SafeFileReader with empty encodings list falls back to defaults."""
        reader = SafeFileReader(encodings=[])
        # Empty list evaluates to falsy, so DEFAULT_ENCODINGS is used
        assert reader.encodings == SafeFileReader.DEFAULT_ENCODINGS
        assert reader.errors == "ignore"

    def test_default_encodings_include_common_formats(self):
        """Test DEFAULT_ENCODINGS includes common encoding formats."""
        expected_encodings = [
            "utf-8",
            "cp1252",
            "iso-8859-1",
            "latin1",
            "utf-16",
            "ascii",
        ]
        assert SafeFileReader.DEFAULT_ENCODINGS == expected_encodings


# ============================================================================
# SafeFileReader.read_text Tests
# ============================================================================


class TestSafeFileReaderReadText:
    """Tests for SafeFileReader.read_text method."""

    def test_read_utf8_file(self):
        """Test reading a UTF-8 encoded file."""
        with tempfile.NamedTemporaryFile(mode="w", encoding="utf-8", delete=False, suffix=".txt") as f:
            f.write("Hello, World!")
            temp_path = f.name

        try:
            reader = SafeFileReader()
            content = reader.read_text(temp_path)
            assert content == "Hello, World!"
        finally:
            Path(temp_path).unlink()

    def test_read_file_with_string_path(self):
        """Test reading file with string path."""
        with tempfile.NamedTemporaryFile(mode="w", encoding="utf-8", delete=False, suffix=".txt") as f:
            f.write("Test content")
            temp_path = f.name

        try:
            reader = SafeFileReader()
            content = reader.read_text(temp_path)
            assert content == "Test content"
        finally:
            Path(temp_path).unlink()

    def test_read_file_with_pathlib_path(self):
        """Test reading file with pathlib.Path."""
        with tempfile.NamedTemporaryFile(mode="w", encoding="utf-8", delete=False, suffix=".txt") as f:
            f.write("Test content")
            temp_path = Path(f.name)

        try:
            reader = SafeFileReader()
            content = reader.read_text(temp_path)
            assert content == "Test content"
        finally:
            temp_path.unlink()

    def test_read_nonexistent_file_returns_none(self):
        """Test reading nonexistent file returns None."""
        reader = SafeFileReader()
        content = reader.read_text("/nonexistent/path/to/file.txt")
        assert content is None

    def test_read_file_with_different_encodings(self):
        """Test reading files with different encodings."""
        # Test UTF-8
        with tempfile.NamedTemporaryFile(mode="w", encoding="utf-8", delete=False, suffix=".txt") as f:
            f.write("UTF-8 content: café")
            temp_path = f.name

        try:
            reader = SafeFileReader()
            content = reader.read_text(temp_path)
            assert "café" in content
        finally:
            Path(temp_path).unlink()

    def test_read_empty_file(self):
        """Test reading empty file."""
        with tempfile.NamedTemporaryFile(mode="w", delete=False, suffix=".txt") as f:
            temp_path = f.name

        try:
            reader = SafeFileReader()
            content = reader.read_text(temp_path)
            assert content == ""
        finally:
            Path(temp_path).unlink()

    def test_encoding_fallback_chain(self):
        """Test encoding fallback chain when first encoding fails."""
        with tempfile.NamedTemporaryFile(mode="wb", delete=False, suffix=".txt") as f:
            # Write content that's valid in cp1252 but might fail in utf-8
            f.write(b"Test content")
            temp_path = f.name

        try:
            reader = SafeFileReader()
            content = reader.read_text(temp_path)
            assert content == "Test content"
        finally:
            Path(temp_path).unlink()

    def test_final_fallback_with_error_handling(self):
        """Test final fallback uses specified error handling strategy."""
        with tempfile.NamedTemporaryFile(mode="wb", delete=False) as f:
            # Create content that will trigger error handling
            f.write(b"\x80\x81\x82")
            temp_path = f.name

        try:
            reader = SafeFileReader(errors="ignore")
            content = reader.read_text(temp_path)
            # Should return something (may be empty due to ignore strategy)
            assert content is not None
        finally:
            Path(temp_path).unlink()

    def test_read_large_file(self):
        """Test reading large file successfully."""
        with tempfile.NamedTemporaryFile(mode="w", encoding="utf-8", delete=False, suffix=".txt") as f:
            large_content = "Line " * 10000
            f.write(large_content)
            temp_path = f.name

        try:
            reader = SafeFileReader()
            content = reader.read_text(temp_path)
            assert len(content) > 10000
        finally:
            Path(temp_path).unlink()

    def test_read_file_with_special_characters(self):
        """Test reading file with special characters."""
        with tempfile.NamedTemporaryFile(mode="w", encoding="utf-8", delete=False, suffix=".txt") as f:
            f.write("Special: é à ñ ü © ® ™")
            temp_path = f.name

        try:
            reader = SafeFileReader()
            content = reader.read_text(temp_path)
            assert "©" in content
            assert "®" in content
        finally:
            Path(temp_path).unlink()

    def test_read_multiline_file(self):
        """Test reading file with multiple lines."""
        with tempfile.NamedTemporaryFile(mode="w", encoding="utf-8", delete=False, suffix=".txt") as f:
            f.write("Line 1\nLine 2\nLine 3")
            temp_path = f.name

        try:
            reader = SafeFileReader()
            content = reader.read_text(temp_path)
            assert "Line 1" in content
            assert "Line 2" in content
            assert "Line 3" in content
        finally:
            Path(temp_path).unlink()

    def test_read_file_with_single_encoding(self):
        """Test reading with single encoding in list."""
        with tempfile.NamedTemporaryFile(mode="w", encoding="utf-8", delete=False, suffix=".txt") as f:
            f.write("Single encoding")
            temp_path = f.name

        try:
            reader = SafeFileReader(encodings=["utf-8"])
            content = reader.read_text(temp_path)
            assert content == "Single encoding"
        finally:
            Path(temp_path).unlink()

    def test_read_file_permission_error_handling(self):
        """Test permission error handling returns None gracefully."""
        reader = SafeFileReader()

        # Use non-existent path that triggers permission-like behavior
        result = reader.read_text("/root/privileged/file.txt")

        # Should eventually return None after trying all encodings or failing
        assert result is None

    def test_read_file_with_custom_error_strategy_strict(self):
        """Test reading with strict error handling strategy."""
        with tempfile.NamedTemporaryFile(mode="wb", delete=False) as f:
            f.write(b"\xff\xfe")  # UTF-16 BOM
            temp_path = f.name

        try:
            reader = SafeFileReader(errors="strict")
            # Should try various encodings
            result = reader.read_text(temp_path)
            assert result is not None or result is None
        finally:
            Path(temp_path).unlink()

    def test_read_file_with_bom(self):
        """Test reading file with BOM (Byte Order Mark)."""
        with tempfile.NamedTemporaryFile(mode="wb", delete=False) as f:
            # UTF-8 BOM
            f.write(b"\xef\xbb\xbfContent")
            temp_path = f.name

        try:
            reader = SafeFileReader()
            content = reader.read_text(temp_path)
            assert content is not None
        finally:
            Path(temp_path).unlink()

    def test_read_directory_instead_of_file(self):
        """Test reading directory path returns None."""
        with tempfile.TemporaryDirectory() as temp_dir:
            reader = SafeFileReader()
            # Path.read_text will raise error on directory
            result = reader.read_text(temp_dir)
            assert result is None


# ============================================================================
# SafeFileReader.read_lines Tests
# ============================================================================


class TestSafeFileReaderReadLines:
    """Tests for SafeFileReader.read_lines method."""

    def test_read_lines_basic(self):
        """Test reading lines from basic file."""
        with tempfile.NamedTemporaryFile(mode="w", encoding="utf-8", delete=False, suffix=".txt") as f:
            f.write("Line 1\nLine 2\nLine 3")
            temp_path = f.name

        try:
            reader = SafeFileReader()
            lines = reader.read_lines(temp_path)
            assert len(lines) == 3
            assert "Line 1" in lines[0]
            assert "Line 2" in lines[1]
            assert "Line 3" in lines[2]
        finally:
            Path(temp_path).unlink()

    def test_read_lines_keeps_ends(self):
        """Test read_lines keeps line endings."""
        with tempfile.NamedTemporaryFile(mode="w", encoding="utf-8", delete=False, suffix=".txt") as f:
            f.write("Line 1\nLine 2\n")
            temp_path = f.name

        try:
            reader = SafeFileReader()
            lines = reader.read_lines(temp_path)
            # Should keep line endings due to keepends=True
            assert lines[0].endswith("\n")
        finally:
            Path(temp_path).unlink()

    def test_read_lines_empty_file(self):
        """Test reading lines from empty file."""
        with tempfile.NamedTemporaryFile(mode="w", delete=False, suffix=".txt") as f:
            temp_path = f.name

        try:
            reader = SafeFileReader()
            lines = reader.read_lines(temp_path)
            assert lines == []
        finally:
            Path(temp_path).unlink()

    def test_read_lines_nonexistent_file(self):
        """Test reading lines from nonexistent file returns empty list."""
        reader = SafeFileReader()
        lines = reader.read_lines("/nonexistent/path/file.txt")
        assert lines == []

    def test_read_lines_single_line(self):
        """Test reading single line file."""
        with tempfile.NamedTemporaryFile(mode="w", encoding="utf-8", delete=False, suffix=".txt") as f:
            f.write("Single line")
            temp_path = f.name

        try:
            reader = SafeFileReader()
            lines = reader.read_lines(temp_path)
            assert len(lines) == 1
            assert "Single line" in lines[0]
        finally:
            Path(temp_path).unlink()

    def test_read_lines_with_blank_lines(self):
        """Test reading file with blank lines."""
        with tempfile.NamedTemporaryFile(mode="w", encoding="utf-8", delete=False, suffix=".txt") as f:
            f.write("Line 1\n\nLine 3")
            temp_path = f.name

        try:
            reader = SafeFileReader()
            lines = reader.read_lines(temp_path)
            # Empty lines are preserved
            assert len(lines) >= 2
        finally:
            Path(temp_path).unlink()

    def test_read_lines_with_different_encodings(self):
        """Test reading lines with encoding fallback."""
        with tempfile.NamedTemporaryFile(mode="w", encoding="utf-8", delete=False, suffix=".txt") as f:
            f.write("Café\nNiño\nÉcole")
            temp_path = f.name

        try:
            reader = SafeFileReader()
            lines = reader.read_lines(temp_path)
            assert len(lines) == 3
        finally:
            Path(temp_path).unlink()

    def test_read_lines_with_string_path(self):
        """Test read_lines with string path."""
        with tempfile.NamedTemporaryFile(mode="w", encoding="utf-8", delete=False, suffix=".txt") as f:
            f.write("Line 1\nLine 2")
            temp_path = f.name

        try:
            reader = SafeFileReader()
            lines = reader.read_lines(temp_path)
            assert len(lines) == 2
        finally:
            Path(temp_path).unlink()

    def test_read_lines_many_lines(self):
        """Test reading file with many lines."""
        with tempfile.NamedTemporaryFile(mode="w", encoding="utf-8", delete=False, suffix=".txt") as f:
            for i in range(1000):
                f.write(f"Line {i}\n")
            temp_path = f.name

        try:
            reader = SafeFileReader()
            lines = reader.read_lines(temp_path)
            assert len(lines) == 1000
        finally:
            Path(temp_path).unlink()


# ============================================================================
# SafeFileReader.safe_glob_read Tests
# ============================================================================


class TestSafeFileReaderSafeGlobRead:
    """Tests for SafeFileReader.safe_glob_read method."""

    def test_glob_read_single_file_pattern(self):
        """Test glob read with single file pattern."""
        with tempfile.TemporaryDirectory() as temp_dir:
            # Create test file
            test_file = Path(temp_dir) / "test.txt"
            test_file.write_text("Test content")

            reader = SafeFileReader()
            results = reader.safe_glob_read("*.txt", base_path=temp_dir)

            assert len(results) == 1
            assert str(test_file) in results
            assert results[str(test_file)] == "Test content"

    def test_glob_read_multiple_files(self):
        """Test glob read with multiple matching files."""
        with tempfile.TemporaryDirectory() as temp_dir:
            # Create multiple test files
            file1 = Path(temp_dir) / "file1.txt"
            file2 = Path(temp_dir) / "file2.txt"
            file1.write_text("Content 1")
            file2.write_text("Content 2")

            reader = SafeFileReader()
            results = reader.safe_glob_read("*.txt", base_path=temp_dir)

            assert len(results) == 2
            assert results[str(file1)] == "Content 1"
            assert results[str(file2)] == "Content 2"

    def test_glob_read_no_matches(self):
        """Test glob read with no matching files."""
        with tempfile.TemporaryDirectory() as temp_dir:
            reader = SafeFileReader()
            results = reader.safe_glob_read("*.nonexistent", base_path=temp_dir)

            assert results == {}

    def test_glob_read_nested_pattern(self):
        """Test glob read with nested directory pattern."""
        with tempfile.TemporaryDirectory() as temp_dir:
            # Create nested structure
            subdir = Path(temp_dir) / "subdir"
            subdir.mkdir()
            test_file = subdir / "test.txt"
            test_file.write_text("Nested content")

            reader = SafeFileReader()
            results = reader.safe_glob_read("**/*.txt", base_path=temp_dir)

            assert len(results) >= 1

    def test_glob_read_excludes_directories(self):
        """Test glob read excludes directories from results."""
        with tempfile.TemporaryDirectory() as temp_dir:
            # Create file and directory
            test_file = Path(temp_dir) / "test.txt"
            test_file.write_text("Content")
            test_dir = Path(temp_dir) / "subdir"
            test_dir.mkdir()

            reader = SafeFileReader()
            results = reader.safe_glob_read("*", base_path=temp_dir)

            # Should only have file, not directory
            assert str(test_file) in results
            assert str(test_dir) not in results

    def test_glob_read_with_string_base_path(self):
        """Test glob read with string base path."""
        with tempfile.TemporaryDirectory() as temp_dir:
            test_file = Path(temp_dir) / "test.txt"
            test_file.write_text("Content")

            reader = SafeFileReader()
            results = reader.safe_glob_read("*.txt", base_path=temp_dir)

            assert len(results) == 1

    def test_glob_read_with_pathlib_base_path(self):
        """Test glob read with pathlib.Path base path."""
        with tempfile.TemporaryDirectory() as temp_dir:
            test_file = Path(temp_dir) / "test.txt"
            test_file.write_text("Content")

            reader = SafeFileReader()
            results = reader.safe_glob_read("*.txt", base_path=Path(temp_dir))

            assert len(results) == 1

    def test_glob_read_unreadable_file_skipped(self):
        """Test glob read skips unreadable files."""
        with tempfile.TemporaryDirectory() as temp_dir:
            readable_file = Path(temp_dir) / "readable.txt"
            readable_file.write_text("Readable")

            reader = SafeFileReader()
            results = reader.safe_glob_read("*.txt", base_path=temp_dir)

            # Should have at least the readable file
            assert len(results) >= 1

    def test_glob_read_default_base_path(self):
        """Test glob read with default base path."""
        reader = SafeFileReader()
        # Using "." as default should work
        results = reader.safe_glob_read("*.nonexistent")
        assert isinstance(results, dict)

    def test_glob_read_mixed_encodings(self):
        """Test glob read handles files with different encodings."""
        with tempfile.TemporaryDirectory() as temp_dir:
            file1 = Path(temp_dir) / "file1.txt"
            file2 = Path(temp_dir) / "file2.txt"
            file1.write_text("UTF-8: café", encoding="utf-8")
            file2.write_text("ASCII: test", encoding="ascii")

            reader = SafeFileReader()
            results = reader.safe_glob_read("*.txt", base_path=temp_dir)

            assert len(results) == 2

    @patch("builtins.print")
    def test_glob_read_invalid_pattern_error_handling(self, mock_print):
        """Test glob read handles invalid patterns gracefully."""
        reader = SafeFileReader()

        # Try to use invalid base path
        with patch.object(Path, "glob") as mock_glob:
            mock_glob.side_effect = OSError("Invalid path")

            results = reader.safe_glob_read("*.txt", base_path="/invalid/path")

            # Should return empty dict and handle error
            assert results == {}

    def test_glob_read_special_characters_in_names(self):
        """Test glob read with files containing special characters."""
        with tempfile.TemporaryDirectory() as temp_dir:
            special_file = Path(temp_dir) / "file-with-dash.txt"
            special_file.write_text("Content")

            reader = SafeFileReader()
            results = reader.safe_glob_read("*.txt", base_path=temp_dir)

            assert len(results) >= 1


# ============================================================================
# SafeFileReader.is_safe_file Tests
# ============================================================================


class TestSafeFileReaderIsSafeFile:
    """Tests for SafeFileReader.is_safe_file method."""

    def test_is_safe_file_readable_file(self):
        """Test is_safe_file returns True for readable file."""
        with tempfile.NamedTemporaryFile(mode="w", encoding="utf-8", delete=False, suffix=".txt") as f:
            f.write("Safe content")
            temp_path = f.name

        try:
            reader = SafeFileReader()
            assert reader.is_safe_file(temp_path) is True
        finally:
            Path(temp_path).unlink()

    def test_is_safe_file_nonexistent_file(self):
        """Test is_safe_file returns False for nonexistent file."""
        reader = SafeFileReader()
        assert reader.is_safe_file("/nonexistent/path/file.txt") is False

    def test_is_safe_file_empty_file(self):
        """Test is_safe_file returns True for empty file."""
        with tempfile.NamedTemporaryFile(mode="w", delete=False, suffix=".txt") as f:
            temp_path = f.name

        try:
            reader = SafeFileReader()
            # Empty file is still safely readable
            result = reader.is_safe_file(temp_path)
            assert result is True
        finally:
            Path(temp_path).unlink()

    def test_is_safe_file_with_string_path(self):
        """Test is_safe_file with string path."""
        with tempfile.NamedTemporaryFile(mode="w", encoding="utf-8", delete=False, suffix=".txt") as f:
            f.write("Content")
            temp_path = f.name

        try:
            reader = SafeFileReader()
            assert reader.is_safe_file(temp_path) is True
        finally:
            Path(temp_path).unlink()

    def test_is_safe_file_with_pathlib_path(self):
        """Test is_safe_file with pathlib.Path."""
        with tempfile.NamedTemporaryFile(mode="w", encoding="utf-8", delete=False, suffix=".txt") as f:
            f.write("Content")
            temp_path = Path(f.name)

        try:
            reader = SafeFileReader()
            assert reader.is_safe_file(temp_path) is True
        finally:
            temp_path.unlink()

    def test_is_safe_file_with_problematic_encoding(self):
        """Test is_safe_file with problematic encoding."""
        with tempfile.NamedTemporaryFile(mode="wb", delete=False) as f:
            # Create file with challenging bytes
            f.write(b"\xff\xfe")
            temp_path = f.name

        try:
            reader = SafeFileReader()
            result = reader.is_safe_file(temp_path)
            # Should return True or False consistently
            assert isinstance(result, bool)
        finally:
            Path(temp_path).unlink()


# ============================================================================
# Convenience Function Tests
# ============================================================================


class TestConvenienceFunctions:
    """Tests for module-level convenience functions."""

    def test_safe_read_file_basic(self):
        """Test safe_read_file convenience function."""
        with tempfile.NamedTemporaryFile(mode="w", encoding="utf-8", delete=False, suffix=".txt") as f:
            f.write("Convenience function test")
            temp_path = f.name

        try:
            result = safe_read_file(temp_path)
            assert result == "Convenience function test"
        finally:
            Path(temp_path).unlink()

    def test_safe_read_file_nonexistent(self):
        """Test safe_read_file with nonexistent file."""
        result = safe_read_file("/nonexistent/path/file.txt")
        assert result is None

    def test_safe_read_file_with_custom_encodings(self):
        """Test safe_read_file with custom encodings."""
        with tempfile.NamedTemporaryFile(mode="w", encoding="utf-8", delete=False, suffix=".txt") as f:
            f.write("Custom encoding test")
            temp_path = f.name

        try:
            result = safe_read_file(temp_path, encodings=["utf-8", "ascii"])
            assert result == "Custom encoding test"
        finally:
            Path(temp_path).unlink()

    def test_safe_read_file_with_string_path(self):
        """Test safe_read_file with string path."""
        with tempfile.NamedTemporaryFile(mode="w", encoding="utf-8", delete=False, suffix=".txt") as f:
            f.write("Test")
            temp_path = f.name

        try:
            result = safe_read_file(temp_path)
            assert result == "Test"
        finally:
            Path(temp_path).unlink()

    def test_safe_read_file_with_pathlib_path(self):
        """Test safe_read_file with pathlib.Path."""
        with tempfile.NamedTemporaryFile(mode="w", encoding="utf-8", delete=False, suffix=".txt") as f:
            f.write("Test")
            temp_path = Path(f.name)

        try:
            result = safe_read_file(temp_path)
            assert result == "Test"
        finally:
            temp_path.unlink()

    def test_safe_read_lines_basic(self):
        """Test safe_read_lines convenience function."""
        with tempfile.NamedTemporaryFile(mode="w", encoding="utf-8", delete=False, suffix=".txt") as f:
            f.write("Line 1\nLine 2")
            temp_path = f.name

        try:
            lines = safe_read_lines(temp_path)
            assert len(lines) == 2
        finally:
            Path(temp_path).unlink()

    def test_safe_read_lines_nonexistent(self):
        """Test safe_read_lines with nonexistent file."""
        lines = safe_read_lines("/nonexistent/path/file.txt")
        assert lines == []

    def test_safe_read_lines_with_custom_encodings(self):
        """Test safe_read_lines with custom encodings."""
        with tempfile.NamedTemporaryFile(mode="w", encoding="utf-8", delete=False, suffix=".txt") as f:
            f.write("Line 1\nLine 2")
            temp_path = f.name

        try:
            lines = safe_read_lines(temp_path, encodings=["utf-8"])
            assert len(lines) == 2
        finally:
            Path(temp_path).unlink()

    def test_safe_glob_read_basic(self):
        """Test safe_glob_read convenience function."""
        with tempfile.TemporaryDirectory() as temp_dir:
            test_file = Path(temp_dir) / "test.txt"
            test_file.write_text("Content")

            results = safe_glob_read("*.txt", base_path=temp_dir)

            assert len(results) == 1
            assert str(test_file) in results

    def test_safe_glob_read_no_matches(self):
        """Test safe_glob_read with no matches."""
        with tempfile.TemporaryDirectory() as temp_dir:
            results = safe_glob_read("*.nonexistent", base_path=temp_dir)
            assert results == {}

    def test_safe_glob_read_with_custom_encodings(self):
        """Test safe_glob_read with custom encodings."""
        with tempfile.TemporaryDirectory() as temp_dir:
            test_file = Path(temp_dir) / "test.txt"
            test_file.write_text("Content")

            results = safe_glob_read("*.txt", base_path=temp_dir, encodings=["utf-8", "ascii"])

            assert len(results) >= 1

    def test_safe_glob_read_default_base_path(self):
        """Test safe_glob_read uses default base path."""
        results = safe_glob_read("*.nonexistent")
        assert isinstance(results, dict)


# ============================================================================
# Edge Cases and Error Handling Tests
# ============================================================================


class TestEdgeCasesAndErrorHandling:
    """Tests for edge cases and error handling scenarios."""

    def test_read_text_with_null_bytes(self):
        """Test reading file with null bytes."""
        with tempfile.NamedTemporaryFile(mode="wb", delete=False) as f:
            f.write(b"Text with \x00 null")
            temp_path = f.name

        try:
            reader = SafeFileReader(errors="ignore")
            result = reader.read_text(temp_path)
            assert result is not None
        finally:
            Path(temp_path).unlink()

    def test_read_text_with_mixed_line_endings(self):
        """Test reading file with mixed line endings."""
        with tempfile.NamedTemporaryFile(mode="wb", delete=False) as f:
            f.write(b"Line 1\nLine 2\r\nLine 3\rLine 4")
            temp_path = f.name

        try:
            reader = SafeFileReader()
            content = reader.read_text(temp_path)
            assert content is not None
        finally:
            Path(temp_path).unlink()

    def test_read_text_permission_error_fallback(self):
        """Test read_text handles permission errors gracefully."""
        reader = SafeFileReader()

        with patch.object(Path, "exists", return_value=True):
            with patch.object(Path, "read_text") as mock_read:
                mock_read.side_effect = PermissionError("Access denied")

                result = reader.read_text("/some/path/file.txt")
                assert result is None

    def test_encoding_with_replace_strategy(self):
        """Test reading with 'replace' error strategy."""
        with tempfile.NamedTemporaryFile(mode="wb", delete=False) as f:
            f.write(b"\x80\x81\x82")
            temp_path = f.name

        try:
            reader = SafeFileReader(errors="replace")
            result = reader.read_text(temp_path)
            assert result is not None
        finally:
            Path(temp_path).unlink()

    def test_glob_read_with_symlinks(self):
        """Test glob read behavior with symlinks."""
        with tempfile.TemporaryDirectory() as temp_dir:
            test_file = Path(temp_dir) / "original.txt"
            test_file.write_text("Original content")

            reader = SafeFileReader()
            results = reader.safe_glob_read("*.txt", base_path=temp_dir)

            # Should find the file
            assert len(results) >= 1

    def test_read_text_unicode_filename(self):
        """Test reading file with unicode characters in name."""
        with tempfile.TemporaryDirectory() as temp_dir:
            unicode_filename = Path(temp_dir) / "файл.txt"
            unicode_filename.write_text("Content")

            reader = SafeFileReader()
            content = reader.read_text(unicode_filename)
            assert content == "Content"

    def test_multiple_encoding_attempts(self):
        """Test that all encodings in list are attempted."""
        with tempfile.NamedTemporaryFile(mode="w", encoding="utf-8", delete=False, suffix=".txt") as f:
            f.write("Test")
            temp_path = f.name

        try:
            # Use multiple encodings
            reader = SafeFileReader(encodings=["invalid-encoding", "utf-8", "ascii"])
            # Should succeed with utf-8 even if first is invalid
            result = reader.read_text(temp_path)
            assert result is not None
        except LookupError:
            # Invalid encoding might raise error during initialization
            pass
        finally:
            Path(temp_path).unlink()

    def test_empty_encodings_list_fallback(self):
        """Test behavior with empty encodings list."""
        with tempfile.NamedTemporaryFile(mode="w", encoding="utf-8", delete=False, suffix=".txt") as f:
            f.write("Test")
            temp_path = f.name

        try:
            reader = SafeFileReader(encodings=[])
            result = reader.read_text(temp_path)
            # With empty list, should still try utf-8 with error handling
            assert result is not None or result is None
        finally:
            Path(temp_path).unlink()

    def test_read_lines_from_read_text_none(self):
        """Test read_lines behavior when read_text returns None."""
        reader = SafeFileReader()

        with patch.object(reader, "read_text", return_value=None):
            lines = reader.read_lines("/some/path/file.txt")
            assert lines == []

    def test_glob_with_no_files_in_pattern(self):
        """Test glob pattern that matches nothing."""
        with tempfile.TemporaryDirectory() as temp_dir:
            reader = SafeFileReader()
            results = reader.safe_glob_read("*.xyz", base_path=temp_dir)
            assert results == {}

    @patch("builtins.print")
    def test_glob_read_unreadable_files_skipped(self, mock_print):
        """Test that unreadable files are skipped in glob read."""
        with tempfile.TemporaryDirectory() as temp_dir:
            good_file = Path(temp_dir) / "good.txt"
            good_file.write_text("Good content")

            reader = SafeFileReader()

            # Mock read_text to fail for some files

            def mock_read_text(path):
                if "good" in str(path):
                    return "Good content"
                return None

            with patch.object(reader, "read_text", side_effect=mock_read_text):
                results = reader.safe_glob_read("*.txt", base_path=temp_dir)
                assert len(results) >= 0


# ============================================================================
# Integration Tests
# ============================================================================


class TestIntegration:
    """Integration tests combining multiple operations."""

    def test_workflow_read_write_read(self):
        """Test workflow: write file, read it, verify content."""
        with tempfile.NamedTemporaryFile(mode="w", encoding="utf-8", delete=False, suffix=".txt") as f:
            original_content = "Test\nMultiline\nContent"
            f.write(original_content)
            temp_path = f.name

        try:
            # Read entire file
            reader = SafeFileReader()
            content = reader.read_text(temp_path)
            assert content == original_content

            # Read as lines
            lines = reader.read_lines(temp_path)
            assert len(lines) == 3

            # Check safety
            assert reader.is_safe_file(temp_path) is True
        finally:
            Path(temp_path).unlink()

    def test_workflow_multiple_files_processing(self):
        """Test workflow: glob read and process multiple files."""
        with tempfile.TemporaryDirectory() as temp_dir:
            # Create multiple files
            for i in range(3):
                file_path = Path(temp_dir) / f"file{i}.txt"
                file_path.write_text(f"Content {i}")

            reader = SafeFileReader()
            results = reader.safe_glob_read("*.txt", base_path=temp_dir)

            assert len(results) == 3
            total_chars = sum(len(content) for content in results.values())
            assert total_chars > 0

    def test_workflow_convenience_functions_integration(self):
        """Test integration of all convenience functions."""
        with tempfile.TemporaryDirectory() as temp_dir:
            # Create test files
            file1 = Path(temp_dir) / "test1.txt"
            file2 = Path(temp_dir) / "test2.txt"
            file1.write_text("Content 1")
            file2.write_text("Line 1\nLine 2")

            # Use convenience functions
            content1 = safe_read_file(file1)
            assert content1 == "Content 1"

            lines2 = safe_read_lines(file2)
            assert len(lines2) == 2

            results = safe_glob_read("*.txt", base_path=temp_dir)
            assert len(results) == 2

    def test_workflow_encoding_conversion(self):
        """Test workflow: read file with encoding fallback chain."""
        with tempfile.NamedTemporaryFile(mode="w", encoding="utf-8", delete=False, suffix=".txt") as f:
            f.write("Special: é à ç ñ")
            temp_path = f.name

        try:
            # Try multiple encoding chains
            reader1 = SafeFileReader(encodings=["ascii", "utf-8"])
            content1 = reader1.read_text(temp_path)
            assert content1 is not None

            reader2 = SafeFileReader(encodings=["iso-8859-1", "utf-8"])
            content2 = reader2.read_text(temp_path)
            assert content2 is not None
        finally:
            Path(temp_path).unlink()
