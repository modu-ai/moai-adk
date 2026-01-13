"""Comprehensive TDD tests for atomic file operations module.

Targets 100% coverage for atomic_ops.py including:
- atomic_write_text function (lines 16-64)
- atomic_write_json function (lines 67-119)
- Error handling and edge cases
"""

import json
import os
from pathlib import Path
from unittest.mock import patch

from moai_adk.tag_system import atomic_ops
from moai_adk.tag_system.atomic_ops import atomic_write_json


class TestAtomicWriteTextSuccessCases:
    """Test atomic_write_text success scenarios."""

    def test_atomic_write_text_new_file(self):
        """Test atomic_write_text creating new file."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            file_path = Path(temp_dir) / "test.txt"
            content = "Hello, world!"

            result = atomic_ops.atomic_write_text(file_path, content)

            assert result is True
            assert file_path.exists()
            assert file_path.read_text(encoding="utf-8") == content

    def test_atomic_write_text_overwrite_existing(self):
        """Test atomic_write_text overwriting existing file."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            file_path = Path(temp_dir) / "test.txt"
            file_path.write_text("old content")

            result = atomic_ops.atomic_write_text(file_path, "new content")

            assert result is True
            assert file_path.read_text(encoding="utf-8") == "new content"

    def test_atomic_write_text_with_path_object(self):
        """Test atomic_write_text with Path object."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            file_path = Path(temp_dir) / "test.txt"
            result = atomic_ops.atomic_write_text(file_path, "content")

            assert result is True

    def test_atomic_write_text_with_string_path(self):
        """Test atomic_write_text with string path."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            file_path_str = str(Path(temp_dir) / "test.txt")
            result = atomic_ops.atomic_write_text(file_path_str, "content")

            assert result is True
            assert Path(file_path_str).exists()

    def test_atomic_write_text_creates_parent_directories(self):
        """Test atomic_write_text creates parent directories."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            file_path = Path(temp_dir) / "subdir" / "nested" / "test.txt"

            result = atomic_ops.atomic_write_text(file_path, "content")

            assert result is True
            assert file_path.exists()
            assert file_path.parent.exists()

    def test_atomic_write_text_with_custom_encoding(self):
        """Test atomic_write_text with custom encoding."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            file_path = Path(temp_dir) / "test.txt"
            content = "Hello 世界"  # Chinese characters

            result = atomic_ops.atomic_write_text(file_path, content, encoding="utf-8")

            assert result is True
            assert file_path.read_text(encoding="utf-8") == content

    def test_atomic_write_text_with_unicode_content(self):
        """Test atomic_write_text with various Unicode characters."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            file_path = Path(temp_dir) / "test.txt"
            content = "English, 한국어, 日本語, العربية, עברית"

            result = atomic_ops.atomic_write_text(file_path, content)

            assert result is True
            assert file_path.read_text(encoding="utf-8") == content

    def test_atomic_write_text_empty_content(self):
        """Test atomic_write_text with empty content."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            file_path = Path(temp_dir) / "test.txt"

            result = atomic_ops.atomic_write_text(file_path, "")

            assert result is True
            assert file_path.read_text(encoding="utf-8") == ""

    def test_atomic_write_text_multiline_content(self):
        """Test atomic_write_text with multiline content."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            file_path = Path(temp_dir) / "test.txt"
            content = "Line 1\nLine 2\r\nLine 3\rLine 4"

            result = atomic_ops.atomic_write_text(file_path, content)

            assert result is True
            # Python normalizes line endings when reading
            read_content = file_path.read_text(encoding="utf-8")
            assert "Line 1" in read_content
            assert "Line 2" in read_content
            assert "Line 3" in read_content
            assert "Line 4" in read_content

    def test_atomic_write_text_large_content(self):
        """Test atomic_write_text with large content."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            file_path = Path(temp_dir) / "test.txt"
            content = "x" * 1000000  # 1MB of data

            result = atomic_ops.atomic_write_text(file_path, content)

            assert result is True
            assert len(file_path.read_text(encoding="utf-8")) == 1000000

    def test_atomic_write_text_special_characters(self):
        """Test atomic_write_text with special characters."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            file_path = Path(temp_dir) / "test.txt"
            content = "Tab:\t, Newline:\n, Backslash:\\, Quote: \""

            result = atomic_ops.atomic_write_text(file_path, content)

            assert result is True
            assert file_path.read_text(encoding="utf-8") == content


class TestAtomicWriteTextErrorHandling:
    """Test atomic_write_text error handling."""

    def test_atomic_write_text_with_make_dirs_false(self):
        """Test atomic_write_text with make_dirs=False and missing parent."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            file_path = Path(temp_dir) / "nonexistent" / "test.txt"

            result = atomic_ops.atomic_write_text(file_path, "content", make_dirs=False)

            # Should fail because parent doesn't exist
            assert result is False
            assert not file_path.exists()

    def test_atomic_write_text_cleanup_temp_file_on_error(self):
        """Test that temp file is cleaned up on write error."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            file_path = Path(temp_dir) / "test.txt"

            # Mock fdopen to raise error
            with patch("os.fdopen") as mock_fdopen:
                mock_fdopen.side_effect = IOError("Write error")

                result = atomic_ops.atomic_write_text(file_path, "content")

                assert result is False

                # Check for leftover temp files
                temp_files = list(Path(temp_dir).glob(".tmp_*"))
                assert len(temp_files) == 0

    def test_atomic_write_text_handles_os_error(self):
        """Test atomic_write_text handles OSError."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            # Create a directory with the target name (will cause os.replace to fail)
            file_path = Path(temp_dir) / "test.txt"
            file_path.mkdir()

            result = atomic_ops.atomic_write_text(file_path, "content")

            # Should return False on error
            assert result is False

    def test_atomic_write_text_handles_value_error(self):
        """Test atomic_write_text handles ValueError from mkstemp."""
        # This is difficult to test directly, but we verify the function returns False
        # when mkstemp raises an exception
        with patch("tempfile.mkstemp") as mock_mkstemp:
            mock_mkstemp.side_effect = ValueError("Invalid value")

            result = atomic_ops.atomic_write_text(Path("test.txt"), "content")

            assert result is False


class TestAtomicWriteJsonSuccessCases:
    """Test atomic_write_json success scenarios."""

    def test_atomic_write_json_new_file(self):
        """Test atomic_write_json creating new file."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            file_path = Path(temp_dir) / "test.json"
            data = {"key": "value", "number": 42}

            result = atomic_ops.atomic_write_json(file_path, data)

            assert result is True
            assert file_path.exists()

            # Verify JSON content
            loaded_data = json.loads(file_path.read_text(encoding="utf-8"))
            assert loaded_data == data

    def test_atomic_write_json_overwrite_existing(self):
        """Test atomic_write_json overwriting existing file."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            file_path = Path(temp_dir) / "test.json"
            file_path.write_text('{"old": "data"}')

            new_data = {"new": "data"}
            result = atomic_ops.atomic_write_json(file_path, new_data)

            assert result is True

            loaded_data = json.loads(file_path.read_text(encoding="utf-8"))
            assert loaded_data == new_data

    def test_atomic_write_json_with_path_object(self):
        """Test atomic_write_json with Path object."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            file_path = Path(temp_dir) / "test.json"
            result = atomic_ops.atomic_write_json(file_path, {"key": "value"})

            assert result is True

    def test_atomic_write_json_with_string_path(self):
        """Test atomic_write_json with string path."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            file_path_str = str(Path(temp_dir) / "test.json")
            result = atomic_ops.atomic_write_json(file_path_str, {"key": "value"})

            assert result is True
            assert Path(file_path_str).exists()

    def test_atomic_write_json_creates_parent_directories(self):
        """Test atomic_write_json creates parent directories."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            file_path = Path(temp_dir) / "subdir" / "nested" / "test.json"

            result = atomic_ops.atomic_write_json(file_path, {"key": "value"})

            assert result is True
            assert file_path.exists()

    def test_atomic_write_json_default_indent(self):
        """Test atomic_write_json uses default indent of 2."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            file_path = Path(temp_dir) / "test.json"
            data = {"key": "value"}

            atomic_ops.atomic_write_json(file_path, data)

            content = file_path.read_text(encoding="utf-8")

            # Should be pretty-printed with 2-space indent
            assert '  "key"' in content

    def test_atomic_write_json_custom_indent(self):
        """Test atomic_write_json with custom indent."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            file_path = Path(temp_dir) / "test.json"
            data = {"key": "value"}

            atomic_ops.atomic_write_json(file_path, data, indent=4)

            content = file_path.read_text(encoding="utf-8")

            # Should be pretty-printed with 4-space indent
            assert '    "key"' in content

    def test_atomic_write_json_no_indent(self):
        """Test atomic_write_json with indent=None (compact)."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            file_path = Path(temp_dir) / "test.json"
            data = {"key": "value"}

            # Bypass type checking to test indent=None behavior
            result = atomic_write_json(file_path, data, indent=None)  # type: ignore[arg-type]

            assert result is True

            content = file_path.read_text(encoding="utf-8")

            # Should be compact without whitespace
            assert '{"key": "value"}' in content

    def test_atomic_write_json_nested_data(self):
        """Test atomic_write_json with nested data structures."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            file_path = Path(temp_dir) / "test.json"
            data = {
                "level1": {
                    "level2": {
                        "level3": ["a", "b", "c"]
                    }
                },
                "list": [1, 2, 3],
                "string": "test"
            }

            result = atomic_ops.atomic_write_json(file_path, data)

            assert result is True

            loaded_data = json.loads(file_path.read_text(encoding="utf-8"))
            assert loaded_data == data

    def test_atomic_write_json_unicode_content(self):
        """Test atomic_write_json with Unicode content."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            file_path = Path(temp_dir) / "test.json"
            data = {"text": "English, 한국어, 日本語"}

            result = atomic_ops.atomic_write_json(file_path, data)

            assert result is True

            loaded_data = json.loads(file_path.read_text(encoding="utf-8"))
            assert loaded_data["text"] == "English, 한국어, 日本語"

    def test_atomic_write_json_ensure_ascii_true(self):
        """Test atomic_write_json with ensure_ascii=True."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            file_path = Path(temp_dir) / "test.json"
            data = {"text": "한글"}

            atomic_ops.atomic_write_json(file_path, data, ensure_ascii=True)

            content = file_path.read_text(encoding="utf-8")

            # Should escape non-ASCII characters
            assert "\\ud55c\\uae00" in content or "\\u" in content

    def test_atomic_write_json_ensure_ascii_false(self):
        """Test atomic_write_json with ensure_ascii=False."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            file_path = Path(temp_dir) / "test.json"
            data = {"text": "한글"}

            atomic_ops.atomic_write_json(file_path, data, ensure_ascii=False)

            content = file_path.read_text(encoding="utf-8")

            # Should preserve Unicode characters
            assert "한글" in content

    def test_atomic_write_json_special_characters(self):
        """Test atomic_write_json with special characters."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            file_path = Path(temp_dir) / "test.json"
            data = {"special": "Tab:\t, Newline:\n, Quote:\""}

            result = atomic_ops.atomic_write_json(file_path, data)

            assert result is True

            loaded_data = json.loads(file_path.read_text(encoding="utf-8"))
            assert loaded_data == data

    def test_atomic_write_json_empty_dict(self):
        """Test atomic_write_json with empty dictionary."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            file_path = Path(temp_dir) / "test.json"

            result = atomic_ops.atomic_write_json(file_path, {})

            assert result is True

            loaded_data = json.loads(file_path.read_text(encoding="utf-8"))
            assert loaded_data == {}

    def test_atomic_write_json_empty_list(self):
        """Test atomic_write_json with empty list."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            file_path = Path(temp_dir) / "test.json"

            result = atomic_ops.atomic_write_json(file_path, [])

            assert result is True

            loaded_data = json.loads(file_path.read_text(encoding="utf-8"))
            assert loaded_data == []

    def test_atomic_write_json_large_data(self):
        """Test atomic_write_json with large data."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            file_path = Path(temp_dir) / "test.json"
            data = {"items": [{"id": i, "data": "x" * 100} for i in range(1000)]}

            result = atomic_ops.atomic_write_json(file_path, data)

            assert result is True

            loaded_data = json.loads(file_path.read_text(encoding="utf-8"))
            assert len(loaded_data["items"]) == 1000

    def test_atomic_write_json_boolean_values(self):
        """Test atomic_write_json with boolean values."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            file_path = Path(temp_dir) / "test.json"
            data = {"true": True, "false": False}

            result = atomic_ops.atomic_write_json(file_path, data)

            assert result is True

            loaded_data = json.loads(file_path.read_text(encoding="utf-8"))
            assert loaded_data["true"] is True
            assert loaded_data["false"] is False

    def test_atomic_write_json_null_values(self):
        """Test atomic_write_json with null values."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            file_path = Path(temp_dir) / "test.json"
            data = {"null": None, "value": "test"}

            result = atomic_ops.atomic_write_json(file_path, data)

            assert result is True

            loaded_data = json.loads(file_path.read_text(encoding="utf-8"))
            assert loaded_data["null"] is None

    def test_atomic_write_json_numeric_values(self):
        """Test atomic_write_json with various numeric types."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            file_path = Path(temp_dir) / "test.json"
            data = {
                "integer": 42,
                "float": 3.14,
                "negative": -10,
                "zero": 0,
                "scientific": 1.23e-4
            }

            result = atomic_ops.atomic_write_json(file_path, data)

            assert result is True

            loaded_data = json.loads(file_path.read_text(encoding="utf-8"))
            assert loaded_data["integer"] == 42
            assert loaded_data["float"] == 3.14


class TestAtomicWriteJsonErrorHandling:
    """Test atomic_write_json error handling."""

    def test_atomic_write_json_with_make_dirs_false(self):
        """Test atomic_write_json with make_dirs=False and missing parent."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            file_path = Path(temp_dir) / "nonexistent" / "test.json"

            result = atomic_ops.atomic_write_json(file_path, {"key": "value"}, make_dirs=False)

            # Should fail because parent doesn't exist
            assert result is False
            assert not file_path.exists()

    def test_atomic_write_json_cleanup_temp_file_on_error(self):
        """Test that temp file is cleaned up on JSON serialization error."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            file_path = Path(temp_dir) / "test.json"

            # Mock json.dump to raise error
            with patch("json.dump") as mock_dump:
                mock_dump.side_effect = TypeError("Not serializable")

                result = atomic_ops.atomic_write_json(file_path, {"key": object()})

                assert result is False

                # Check for leftover temp files
                temp_files = list(Path(temp_dir).glob(".tmp_*.json"))
                assert len(temp_files) == 0

    def test_atomic_write_json_handles_non_serializable_data(self):
        """Test atomic_write_json handles non-serializable objects."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            file_path = Path(temp_dir) / "test.json"

            # Try to serialize non-serializable object
            class CustomClass:
                pass

            result = atomic_ops.atomic_write_json(file_path, {"obj": CustomClass()})

            # Should return False on error
            assert result is False

    def test_atomic_write_json_handles_os_error(self):
        """Test atomic_write_json handles OSError."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            # Create a directory with the target name
            file_path = Path(temp_dir) / "test.json"
            file_path.mkdir()

            result = atomic_ops.atomic_write_json(file_path, {"key": "value"})

            # Should return False on error
            assert result is False

    def test_atomic_write_json_handles_value_error(self):
        """Test atomic_write_json handles ValueError from mkstemp."""
        with patch("tempfile.mkstemp") as mock_mkstemp:
            mock_mkstemp.side_effect = ValueError("Invalid value")

            result = atomic_ops.atomic_write_json(Path("test.json"), {"key": "value"})

            assert result is False

    def test_atomic_write_json_handles_type_error(self):
        """Test atomic_write_json handles TypeError."""
        with patch("tempfile.mkstemp") as mock_mkstemp:
            mock_mkstemp.side_effect = TypeError("Invalid type")

            result = atomic_ops.atomic_write_json(Path("test.json"), {"key": "value"})

            assert result is False


    def test_atomic_write_text_handles_unlink_os_error(self):
        """Test atomic_write_text handles OSError when cleaning up temp file."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            file_path = Path(temp_dir) / "test.txt"

            # Mock os.unlink to raise OSError
            with patch("os.unlink", side_effect=OSError("Unlink error")):
                # Mock fdopen to also raise error so we enter cleanup code
                with patch("os.fdopen") as mock_fdopen:
                    mock_fdopen.side_effect = IOError("Write error")

                    result = atomic_ops.atomic_write_text(file_path, "content")

                    # Should return False even if cleanup fails
                    assert result is False

    def test_atomic_write_json_handles_unlink_os_error(self):
        """Test atomic_write_json handles OSError when cleaning up temp file."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            file_path = Path(temp_dir) / "test.json"

            # Mock os.unlink to raise OSError
            with patch("os.unlink", side_effect=OSError("Unlink error")):
                # Mock fdopen to also raise error so we enter cleanup code
                with patch("os.fdopen") as mock_fdopen:
                    mock_fdopen.side_effect = IOError("Write error")

                    result = atomic_ops.atomic_write_json(file_path, {"key": "value"})

                    # Should return False even if cleanup fails
                    assert result is False


class TestAtomicWriteOperations:
    """Test atomic write operations behavior."""

    def test_atomic_replace_is_atomic(self):
        """Test that os.replace is atomic (doesn't leave partial files)."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            file_path = Path(temp_dir) / "test.txt"

            # Write initial content
            file_path.write_text("original")

            # Atomic write should replace entire file
            atomic_ops.atomic_write_text(file_path, "replaced")

            # File should have complete new content
            content = file_path.read_text(encoding="utf-8")
            assert content == "replaced"
            assert "original" not in content

    def test_temp_file_has_correct_suffix(self):
        """Test that temp file uses correct suffix."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            file_path = Path(temp_dir) / "test.txt"

            # Mock to capture temp file path
            original_mkstemp = tempfile.mkstemp
            temp_paths = []

            def mock_mkstemp(*args, **kwargs):
                fd, path = original_mkstemp(*args, **kwargs)
                temp_paths.append(path)
                return fd, path

            with patch("tempfile.mkstemp", side_effect=mock_mkstemp):
                atomic_ops.atomic_write_text(file_path, "content")

            # Temp file should have .tmp_ prefix and end with .txt or .tmp
            assert len(temp_paths) == 1
            assert ".tmp_" in temp_paths[0] or temp_paths[0].endswith(".tmp")

    def test_temp_file_for_json_has_json_suffix(self):
        """Test that temp file for JSON uses .json suffix."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            file_path = Path(temp_dir) / "test.json"

            # Mock to capture temp file path
            original_mkstemp = tempfile.mkstemp
            temp_paths = []

            def mock_mkstemp(*args, **kwargs):
                fd, path = original_mkstemp(*args, **kwargs)
                temp_paths.append(path)
                return fd, path

            with patch("tempfile.mkstemp", side_effect=mock_mkstemp):
                atomic_ops.atomic_write_json(file_path, {"key": "value"})

            # Temp file should have .json suffix
            assert len(temp_paths) == 1
            assert temp_paths[0].endswith(".json")

    def test_temp_file_cleanup_on_success(self):
        """Test that temp file is cleaned up after successful write."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            file_path = Path(temp_dir) / "test.txt"

            # Track temp files
            temp_files_before = list(Path(temp_dir).glob(".tmp_*"))

            atomic_ops.atomic_write_text(file_path, "content")

            temp_files_after = list(Path(temp_dir).glob(".tmp_*"))

            # Temp files should be cleaned up
            assert len(temp_files_after) == len(temp_files_before)

    def test_file_permissions_preserved(self):
        """Test that file permissions are handled correctly."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            file_path = Path(temp_dir) / "test.txt"

            atomic_ops.atomic_write_text(file_path, "content")

            # File should be readable
            assert file_path.exists()
            assert os.access(file_path, os.R_OK)

    def test_concurrent_writes_safe(self):
        """Test that concurrent writes are safe (atomicity)."""
        import tempfile
        import threading

        with tempfile.TemporaryDirectory() as temp_dir:
            file_path = Path(temp_dir) / "test.txt"

            # Write initial content
            file_path.write_text("initial")

            # Multiple threads writing concurrently
            def write_content(value):
                atomic_ops.atomic_write_text(file_path, f"content-{value}")

            threads = [threading.Thread(target=write_content, args=(i,)) for i in range(10)]

            for thread in threads:
                thread.start()

            for thread in threads:
                thread.join()

            # File should have one of the written values (not corrupted)
            final_content = file_path.read_text(encoding="utf-8")
            assert final_content.startswith("content-")
