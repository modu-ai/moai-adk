"""Unit tests for moai_adk.utils.toon_utils module.

Tests for TOON encoding/decoding utilities.
"""

import json
from pathlib import Path
from unittest.mock import MagicMock, patch

import pytest

from moai_adk.utils.toon_utils import (
    compare_formats,
    toon_decode,
    toon_encode,
    toon_load,
    toon_save,
    validate_roundtrip,
)


class TestToonEncode:
    """Test toon_encode function."""

    def test_encode_simple_dict(self):
        """Test encoding simple dictionary."""
        data = {"name": "Alice", "age": 30}
        result = toon_encode(data)
        assert "Alice" in result
        assert "30" in result

    def test_encode_list(self):
        """Test encoding list."""
        data = [1, 2, 3, 4, 5]
        result = toon_encode(data)
        assert "1" in result
        assert "5" in result

    def test_encode_nested_dict(self):
        """Test encoding nested dictionary."""
        data = {"users": [{"name": "Alice", "age": 30}]}
        result = toon_encode(data)
        assert "Alice" in result
        assert "users" in result

    def test_encode_returns_string(self):
        """Test encode returns string."""
        data = {"test": "value"}
        result = toon_encode(data)
        assert isinstance(result, str)

    def test_encode_invalid_data(self):
        """Test encode with unencodable data."""
        class CustomClass:
            pass

        # Create an object that can't be JSON encoded
        data = {"obj": CustomClass()}
        with pytest.raises(ValueError):
            toon_encode(data)

    def test_encode_with_none(self):
        """Test encoding None values."""
        data = {"value": None}
        result = toon_encode(data)
        assert "null" in result

    def test_encode_with_bool(self):
        """Test encoding boolean values."""
        data = {"enabled": True, "disabled": False}
        result = toon_encode(data)
        assert "true" in result
        assert "false" in result


class TestToonDecode:
    """Test toon_decode function."""

    def test_decode_simple_json(self):
        """Test decoding simple JSON."""
        toon_str = '{"name": "Alice", "age": 30}'
        result = toon_decode(toon_str)
        assert result["name"] == "Alice"
        assert result["age"] == 30

    def test_decode_list(self):
        """Test decoding list."""
        toon_str = "[1, 2, 3, 4, 5]"
        result = toon_decode(toon_str)
        assert result == [1, 2, 3, 4, 5]

    def test_decode_invalid_json(self):
        """Test decode with invalid JSON."""
        with pytest.raises(ValueError):
            toon_decode("not valid json {")

    def test_decode_empty_dict(self):
        """Test decode empty dict."""
        result = toon_decode("{}")
        assert result == {}

    def test_decode_empty_list(self):
        """Test decode empty list."""
        result = toon_decode("[]")
        assert result == []


class TestValidateRoundtrip:
    """Test validate_roundtrip function."""

    def test_roundtrip_simple_dict(self):
        """Test roundtrip with simple dictionary."""
        data = {"name": "Alice", "age": 30}
        assert validate_roundtrip(data) is True

    def test_roundtrip_nested_structure(self):
        """Test roundtrip with nested structure."""
        data = {
            "users": [
                {"name": "Alice", "age": 30},
                {"name": "Bob", "age": 25},
            ]
        }
        assert validate_roundtrip(data) is True

    def test_roundtrip_with_none(self):
        """Test roundtrip with None values."""
        data = {"value": None, "name": "test"}
        assert validate_roundtrip(data) is True

    def test_roundtrip_with_booleans(self):
        """Test roundtrip with boolean values."""
        data = {"enabled": True, "disabled": False}
        assert validate_roundtrip(data) is True

    def test_roundtrip_empty_structures(self):
        """Test roundtrip with empty structures."""
        assert validate_roundtrip({}) is True
        assert validate_roundtrip([]) is True


class TestCompareFormats:
    """Test compare_formats function."""

    def test_compare_simple_data(self):
        """Test comparing formats for simple data."""
        data = {"name": "Alice", "age": 30}
        result = compare_formats(data)
        assert "json" in result
        assert "toon" in result
        assert "reduction" in result

    def test_compare_has_required_fields(self):
        """Test compare result has required fields."""
        data = {"test": "data"}
        result = compare_formats(data)
        assert "json" in result
        assert "toon" in result
        assert result["json"]["size_bytes"] > 0
        assert result["toon"]["size_bytes"] > 0

    def test_compare_reduction_percentage(self):
        """Test reduction percentage is calculated."""
        data = {"items": [{"id": i, "name": f"Item{i}"} for i in range(10)]}
        result = compare_formats(data)
        assert "reduction" in result
        assert isinstance(result["reduction"], float)

    def test_compare_size_reduction_percent(self):
        """Test size reduction percent is calculated."""
        data = {"test": "data"}
        result = compare_formats(data)
        assert "size_reduction_percent" in result


class TestToonSave:
    """Test toon_save function."""

    def test_save_creates_file(self):
        """Test save creates file."""
        with patch("pathlib.Path.parent") as mock_parent:
            with patch("pathlib.Path.write_text") as mock_write:
                mock_parent.mkdir = MagicMock()
                data = {"test": "data"}
                toon_save(data, Path("/tmp/test.toon"))
                assert mock_write.called

    def test_save_with_path_string(self):
        """Test save accepts string path."""
        with patch("pathlib.Path.write_text"):
            with patch("pathlib.Path.mkdir"):
                data = {"test": "data"}
                toon_save(data, "/tmp/test.toon")

    def test_save_creates_parent_directory(self):
        """Test save creates parent directories."""
        with patch("pathlib.Path.write_text"):
            with patch("pathlib.Path.mkdir") as mock_mkdir:
                data = {"test": "data"}
                toon_save(data, Path("/tmp/nested/dir/test.toon"))
                assert mock_mkdir.called


class TestToonLoad:
    """Test toon_load function."""

    def test_load_from_file(self):
        """Test load from file."""
        test_data = {"test": "data"}
        json_str = json.dumps(test_data)
        with patch("pathlib.Path.read_text", return_value=json_str):
            result = toon_load(Path("/tmp/test.toon"))
            assert result == test_data

    def test_load_with_string_path(self):
        """Test load accepts string path."""
        test_data = {"test": "data"}
        json_str = json.dumps(test_data)
        with patch("pathlib.Path.read_text", return_value=json_str):
            result = toon_load("/tmp/test.toon")
            assert result == test_data

    def test_load_invalid_file(self):
        """Test load with invalid JSON."""
        with patch("pathlib.Path.read_text", return_value="invalid json {"):
            with pytest.raises(ValueError):
                toon_load(Path("/tmp/test.toon"))


class TestToonIntegration:
    """Integration tests for TOON utilities."""

    def test_encode_decode_roundtrip(self):
        """Test complete encode/decode roundtrip."""
        data = {
            "users": [
                {"id": 1, "name": "Alice", "active": True},
                {"id": 2, "name": "Bob", "active": False},
            ],
            "count": 2,
        }
        encoded = toon_encode(data)
        decoded = toon_decode(encoded)
        assert decoded == data

    def test_all_data_types_roundtrip(self):
        """Test roundtrip with various data types."""
        data = {
            "string": "value",
            "integer": 42,
            "float": 3.14,
            "boolean": True,
            "null": None,
            "list": [1, 2, 3],
            "dict": {"nested": "value"},
        }
        assert validate_roundtrip(data) is True
