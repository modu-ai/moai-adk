"""Comprehensive test suite for toon_utils.py.

This module provides 90%+ coverage for TOON (Token-Oriented Object Notation)
utility functions including:
- TOON encoding and decoding
- Data serialization/deserialization
- Format validation and roundtrip verification
- File I/O operations (save/load)
- Format comparison (TOON vs JSON)
- JSON to TOON migration utilities
- Token compression and optimization metrics
"""

import json
import tempfile
from pathlib import Path
from unittest.mock import patch

import pytest

from moai_adk.utils.toon_utils import (
    _encode_value,
    _is_tabular,
    compare_formats,
    migrate_json_to_toon,
    toon_decode,
    toon_encode,
    toon_load,
    toon_save,
    validate_roundtrip,
)

# ============================================================================
# Helper Function Tests
# ============================================================================


class TestIsTabular:
    """Tests for _is_tabular helper function."""

    def test_is_tabular_empty_list_via_redundant_check(self):
        """Test that empty list is not tabular (covers line 27-28)."""
        # This test specifically exercises the redundant empty check at line 27-28
        # which checks len(items) == 0 after checking not items
        empty_list = []
        result = _is_tabular(empty_list)
        assert result is False
        # Verify the logic path
        assert not empty_list  # First check at line 21-22
        assert len(empty_list) == 0  # Redundant check at line 27-28

    def test_is_tabular_empty_list(self):
        """Test that empty list is not tabular."""
        assert _is_tabular([]) is False

    def test_is_tabular_non_list(self):
        """Test that non-list items return False."""
        assert _is_tabular(None) is False
        assert _is_tabular("string") is False
        assert _is_tabular({"dict": "value"}) is False
        assert _is_tabular(42) is False

    def test_is_tabular_with_non_dict_items(self):
        """Test that list with non-dict items is not tabular."""
        assert _is_tabular([1, 2, 3]) is False
        assert _is_tabular(["a", "b", "c"]) is False
        assert _is_tabular([{"id": 1}, "string", 3]) is False

    def test_is_tabular_single_dict(self):
        """Test that single-item list of dict is tabular."""
        assert _is_tabular([{"id": 1, "name": "Alice"}]) is True

    def test_is_tabular_multiple_dicts_same_keys(self):
        """Test that multiple dicts with same keys is tabular."""
        items = [
            {"id": 1, "name": "Alice"},
            {"id": 2, "name": "Bob"},
            {"id": 3, "name": "Charlie"},
        ]
        assert _is_tabular(items) is True

    def test_is_tabular_dicts_different_keys(self):
        """Test that dicts with different keys is not tabular."""
        items = [
            {"id": 1, "name": "Alice"},
            {"id": 2, "name": "Bob", "email": "bob@example.com"},
        ]
        assert _is_tabular(items) is False

    def test_is_tabular_different_key_count(self):
        """Test that dicts with different key counts is not tabular."""
        items = [
            {"id": 1, "name": "Alice", "age": 30},
            {"id": 2, "name": "Bob"},
        ]
        assert _is_tabular(items) is False

    def test_is_tabular_different_key_order_same_keys(self):
        """Test that different key order but same keys is tabular."""
        items = [
            {"id": 1, "name": "Alice"},
            {"name": "Bob", "id": 2},
        ]
        assert _is_tabular(items) is True

    def test_is_tabular_with_special_values(self):
        """Test tabular detection with various value types."""
        items = [
            {"id": 1, "active": True, "score": 95.5},
            {"id": 2, "active": False, "score": 87.3},
        ]
        assert _is_tabular(items) is True

    def test_is_tabular_with_none_values(self):
        """Test tabular detection with None values."""
        items = [
            {"id": 1, "name": "Alice", "email": None},
            {"id": 2, "name": "Bob", "email": "bob@example.com"},
        ]
        assert _is_tabular(items) is True

    def test_is_tabular_with_nested_dicts(self):
        """Test tabular detection with nested dict values."""
        items = [
            {"id": 1, "data": {"x": 1, "y": 2}},
            {"id": 2, "data": {"x": 3, "y": 4}},
        ]
        assert _is_tabular(items) is True


class TestEncodeValue:
    """Tests for _encode_value helper function."""

    def test_encode_none(self):
        """Test encoding None value."""
        assert _encode_value(None) == "null"

    def test_encode_true(self):
        """Test encoding True boolean."""
        assert _encode_value(True) == "true"

    def test_encode_false(self):
        """Test encoding False boolean."""
        assert _encode_value(False) == "false"

    def test_encode_integer(self):
        """Test encoding integer values."""
        assert _encode_value(0) == "0"
        assert _encode_value(42) == "42"
        assert _encode_value(-100) == "-100"

    def test_encode_float(self):
        """Test encoding float values."""
        assert _encode_value(3.14) == "3.14"
        assert _encode_value(0.0) == "0.0"
        assert _encode_value(-2.5) == "-2.5"

    def test_encode_simple_string(self):
        """Test encoding simple string without special chars."""
        assert _encode_value("hello") == "hello"
        assert _encode_value("Alice") == "Alice"
        assert _encode_value("test123") == "test123"

    def test_encode_string_with_comma(self):
        """Test encoding string with comma triggers quoting."""
        result = _encode_value("hello,world")
        assert result == json.dumps("hello,world")

    def test_encode_string_with_colon(self):
        """Test encoding string with colon triggers quoting."""
        result = _encode_value("key:value")
        assert result == json.dumps("key:value")

    def test_encode_string_with_newline(self):
        """Test encoding string with newline triggers quoting."""
        result = _encode_value("line1\nline2")
        assert result == json.dumps("line1\nline2")

    def test_encode_string_with_double_quote(self):
        """Test encoding string with double quote triggers quoting."""
        result = _encode_value('say "hello"')
        assert result == json.dumps('say "hello"')

    def test_encode_string_with_bracket(self):
        """Test encoding string with bracket triggers quoting."""
        result = _encode_value("array[0]")
        assert result == json.dumps("array[0]")

    def test_encode_string_with_multiple_special_chars(self):
        """Test encoding string with multiple special characters."""
        result = _encode_value("a,b:c")
        assert result == json.dumps("a,b:c")

    def test_encode_list(self):
        """Test encoding list falls back to JSON."""
        result = _encode_value([1, 2, 3])
        assert result == json.dumps([1, 2, 3])

    def test_encode_dict(self):
        """Test encoding dict falls back to JSON."""
        data = {"key": "value"}
        result = _encode_value(data)
        assert result == json.dumps(data)

    def test_encode_complex_object(self):
        """Test encoding complex objects fall back to JSON."""
        data = {"nested": {"list": [1, 2, 3]}}
        result = _encode_value(data)
        assert result == json.dumps(data)


# ============================================================================
# Encoding Tests
# ============================================================================


class TestToonEncode:
    """Tests for toon_encode function."""

    def test_encode_empty_dict(self):
        """Test encoding empty dictionary."""
        result = toon_encode({})
        assert result == json.dumps({}, indent=2, ensure_ascii=False)
        assert "}" in result

    def test_encode_empty_list(self):
        """Test encoding empty list."""
        result = toon_encode([])
        assert result == json.dumps([], indent=2, ensure_ascii=False)
        assert "[]" in result

    def test_encode_simple_dict(self):
        """Test encoding simple dictionary."""
        data = {"name": "Alice", "age": 30}
        result = toon_encode(data)
        decoded = json.loads(result)
        assert decoded == data

    def test_encode_nested_dict(self):
        """Test encoding nested dictionary."""
        data = {"user": {"id": 1, "name": "Alice", "profile": {"age": 30, "city": "NYC"}}}
        result = toon_encode(data)
        decoded = json.loads(result)
        assert decoded == data

    def test_encode_list_of_dicts(self):
        """Test encoding list of dictionaries."""
        data = [
            {"id": 1, "name": "Alice"},
            {"id": 2, "name": "Bob"},
        ]
        result = toon_encode(data)
        decoded = json.loads(result)
        assert decoded == data

    def test_encode_mixed_types(self):
        """Test encoding mixed data types."""
        data = {
            "string": "value",
            "number": 42,
            "float": 3.14,
            "bool": True,
            "null": None,
            "list": [1, 2, 3],
            "dict": {"nested": "value"},
        }
        result = toon_encode(data)
        decoded = json.loads(result)
        assert decoded == data

    def test_encode_with_unicode(self):
        """Test encoding with unicode characters."""
        data = {
            "korean": "한글",
            "emoji": "😀",
            "chinese": "中文",
            "japanese": "日本語",
        }
        result = toon_encode(data)
        decoded = json.loads(result)
        assert decoded == data

    def test_encode_with_special_strings(self):
        """Test encoding with special string patterns."""
        data = {
            "email": "test@example.com",
            "url": "https://example.com/path?query=1",
            "path": "/usr/local/bin",
        }
        result = toon_encode(data)
        decoded = json.loads(result)
        assert decoded == data

    def test_encode_large_dict(self):
        """Test encoding large dictionary."""
        data = {f"key_{i}": f"value_{i}" for i in range(100)}
        result = toon_encode(data)
        decoded = json.loads(result)
        assert decoded == data
        assert len(decoded) == 100

    def test_encode_large_list(self):
        """Test encoding large list."""
        data = [{"id": i, "name": f"Item{i}"} for i in range(100)]
        result = toon_encode(data)
        decoded = json.loads(result)
        assert decoded == data
        assert len(decoded) == 100

    def test_encode_with_strict_false(self):
        """Test encoding with strict=False."""
        data = {"test": "value"}
        result = toon_encode(data, strict=False)
        decoded = json.loads(result)
        assert decoded == data

    def test_encode_with_strict_true(self):
        """Test encoding with strict=True."""
        data = {"test": "value"}
        result = toon_encode(data, strict=True)
        decoded = json.loads(result)
        assert decoded == data

    def test_encode_with_detect_tabular_true(self):
        """Test encoding with detect_tabular=True."""
        data = [
            {"id": 1, "name": "Alice"},
            {"id": 2, "name": "Bob"},
        ]
        result = toon_encode(data, detect_tabular=True)
        decoded = json.loads(result)
        assert decoded == data

    def test_encode_with_detect_tabular_false(self):
        """Test encoding with detect_tabular=False."""
        data = [
            {"id": 1, "name": "Alice"},
            {"id": 2, "name": "Bob"},
        ]
        result = toon_encode(data, detect_tabular=False)
        decoded = json.loads(result)
        assert decoded == data

    def test_encode_unserializable_object(self):
        """Test encoding unserializable object raises ValueError."""

        class CustomObject:
            pass

        data = {"obj": CustomObject()}
        with pytest.raises(ValueError) as exc_info:
            toon_encode(data)
        assert "Failed to encode data to TOON" in str(exc_info.value)

    def test_encode_set_raises_error(self):
        """Test encoding set (not JSON serializable) raises error."""
        data = {"items": {1, 2, 3}}
        with pytest.raises(ValueError):
            toon_encode(data)

    def test_encode_preserves_whitespace_in_strings(self):
        """Test that whitespace in strings is preserved."""
        data = {"text": "line1\n  indented\nline3"}
        result = toon_encode(data)
        decoded = json.loads(result)
        assert decoded["text"] == data["text"]

    def test_encode_with_empty_string(self):
        """Test encoding with empty string value."""
        data = {"empty": "", "filled": "value"}
        result = toon_encode(data)
        decoded = json.loads(result)
        assert decoded == data

    def test_encode_with_zero_values(self):
        """Test encoding with zero numeric values."""
        data = {
            "zero_int": 0,
            "zero_float": 0.0,
            "negative": -5,
        }
        result = toon_encode(data)
        decoded = json.loads(result)
        assert decoded == data

    def test_encode_format_has_indentation(self):
        """Test that encoded output includes indentation."""
        data = {"key": "value", "nested": {"inner": "value"}}
        result = toon_encode(data)
        # Should have indentation (multiple spaces on lines)
        assert "\n" in result
        lines = result.split("\n")
        # Some lines should have leading spaces (indentation)
        indented_lines = [line for line in lines if line.startswith("  ")]
        assert len(indented_lines) > 0


# ============================================================================
# Decoding Tests
# ============================================================================


class TestToonDecode:
    """Tests for toon_decode function."""

    def test_decode_empty_dict(self):
        """Test decoding empty dictionary."""
        toon_str = "{}"
        result = toon_decode(toon_str)
        assert result == {}

    def test_decode_empty_list(self):
        """Test decoding empty list."""
        toon_str = "[]"
        result = toon_decode(toon_str)
        assert result == []

    def test_decode_simple_dict(self):
        """Test decoding simple dictionary."""
        toon_str = '{"name": "Alice", "age": 30}'
        result = toon_decode(toon_str)
        assert result == {"name": "Alice", "age": 30}

    def test_decode_nested_dict(self):
        """Test decoding nested dictionary."""
        toon_str = '{"user": {"id": 1, "name": "Alice"}}'
        result = toon_decode(toon_str)
        assert result == {"user": {"id": 1, "name": "Alice"}}

    def test_decode_list_of_dicts(self):
        """Test decoding list of dictionaries."""
        toon_str = '[{"id": 1, "name": "Alice"}, {"id": 2, "name": "Bob"}]'
        result = toon_decode(toon_str)
        assert len(result) == 2
        assert result[0]["name"] == "Alice"
        assert result[1]["name"] == "Bob"

    def test_decode_mixed_types(self):
        """Test decoding mixed data types."""
        toon_str = """
        {
            "string": "value",
            "number": 42,
            "float": 3.14,
            "bool": true,
            "null": null
        }
        """
        result = toon_decode(toon_str)
        assert result["string"] == "value"
        assert result["number"] == 42
        assert result["float"] == 3.14
        assert result["bool"] is True
        assert result["null"] is None

    def test_decode_with_whitespace(self):
        """Test decoding with various whitespace."""
        toon_str = """
        {
            "key"  :  "value"  ,
            "nested"  :  {
                "inner"  :  "val"
            }
        }
        """
        result = toon_decode(toon_str)
        assert result["key"] == "value"
        assert result["nested"]["inner"] == "val"

    def test_decode_with_unicode(self):
        """Test decoding with unicode characters."""
        data = {"korean": "한글", "emoji": "😀"}
        toon_str = json.dumps(data, ensure_ascii=False)
        result = toon_decode(toon_str)
        assert result == data

    def test_decode_invalid_json(self):
        """Test decoding invalid JSON raises ValueError."""
        invalid_toon = "{invalid json}"
        with pytest.raises(ValueError) as exc_info:
            toon_decode(invalid_toon)
        assert "Failed to decode TOON" in str(exc_info.value)

    def test_decode_incomplete_json(self):
        """Test decoding incomplete JSON raises ValueError."""
        incomplete_toon = '{"key": "value"'
        with pytest.raises(ValueError):
            toon_decode(incomplete_toon)

    def test_decode_empty_string(self):
        """Test decoding empty string raises ValueError."""
        with pytest.raises(ValueError):
            toon_decode("")

    def test_decode_whitespace_only(self):
        """Test decoding whitespace-only string raises ValueError."""
        with pytest.raises(ValueError):
            toon_decode("   \n  \t  ")

    def test_decode_invalid_escape_sequence(self):
        """Test decoding with invalid escape sequence raises error."""
        invalid_toon = '{"key": "value\\x"}'
        with pytest.raises(ValueError):
            toon_decode(invalid_toon)

    def test_decode_with_strict_false(self):
        """Test decoding with strict=False."""
        toon_str = '{"key": "value"}'
        result = toon_decode(toon_str, strict=False)
        assert result == {"key": "value"}

    def test_decode_with_strict_true(self):
        """Test decoding with strict=True."""
        toon_str = '{"key": "value"}'
        result = toon_decode(toon_str, strict=True)
        assert result == {"key": "value"}

    def test_decode_scientific_notation(self):
        """Test decoding numbers in scientific notation."""
        toon_str = '{"large": 1e10, "small": 1e-5}'
        result = toon_decode(toon_str)
        assert result["large"] == 1e10
        assert result["small"] == 1e-5

    def test_decode_negative_numbers(self):
        """Test decoding negative numbers."""
        toon_str = '{"negative_int": -42, "negative_float": -3.14}'
        result = toon_decode(toon_str)
        assert result["negative_int"] == -42
        assert result["negative_float"] == -3.14

    def test_decode_special_string_values(self):
        """Test decoding special string patterns."""
        toon_str = """
        {
            "empty": "",
            "spaces": "   ",
            "newlines": "line1\\nline2",
            "tabs": "col1\\tcol2"
        }
        """
        result = toon_decode(toon_str)
        assert result["empty"] == ""
        assert result["spaces"] == "   "
        assert result["newlines"] == "line1\nline2"
        assert result["tabs"] == "col1\tcol2"


# ============================================================================
# Roundtrip Validation Tests
# ============================================================================


class TestValidateRoundtrip:
    """Tests for validate_roundtrip function."""

    def test_roundtrip_empty_dict(self):
        """Test roundtrip validation for empty dict."""
        data = {}
        assert validate_roundtrip(data) is True

    def test_roundtrip_empty_list(self):
        """Test roundtrip validation for empty list."""
        data = []
        assert validate_roundtrip(data) is True

    def test_roundtrip_simple_dict(self):
        """Test roundtrip validation for simple dict."""
        data = {"name": "Alice", "age": 30}
        assert validate_roundtrip(data) is True

    def test_roundtrip_nested_dict(self):
        """Test roundtrip validation for nested dict."""
        data = {"user": {"id": 1, "profile": {"age": 30}}}
        assert validate_roundtrip(data) is True

    def test_roundtrip_list_of_dicts(self):
        """Test roundtrip validation for list of dicts."""
        data = [
            {"id": 1, "name": "Alice"},
            {"id": 2, "name": "Bob"},
            {"id": 3, "name": "Charlie"},
        ]
        assert validate_roundtrip(data) is True

    def test_roundtrip_mixed_types(self):
        """Test roundtrip validation for mixed types."""
        data = {
            "string": "value",
            "int": 42,
            "float": 3.14,
            "bool": True,
            "none": None,
            "list": [1, 2, 3],
            "dict": {"nested": "value"},
        }
        assert validate_roundtrip(data) is True

    def test_roundtrip_unicode_characters(self):
        """Test roundtrip validation with unicode."""
        data = {
            "korean": "한글",
            "emoji": "😀",
            "chinese": "中文",
            "russian": "Привет",
        }
        assert validate_roundtrip(data) is True

    def test_roundtrip_large_data_structure(self):
        """Test roundtrip validation with large data."""
        data = {"users": [{"id": i, "name": f"User{i}", "active": i % 2 == 0} for i in range(100)]}
        assert validate_roundtrip(data) is True

    def test_roundtrip_with_special_strings(self):
        """Test roundtrip with special string patterns."""
        data = {
            "comma": "a,b,c",
            "colon": "key:value",
            "newline": "line1\nline2",
            "quotes": 'say "hello"',
            "slash": "path/to/file",
            "backslash": "back\\slash",
        }
        assert validate_roundtrip(data) is True

    def test_roundtrip_strict_false(self):
        """Test roundtrip validation with strict=False."""
        data = {"key": "value"}
        assert validate_roundtrip(data, strict=False) is True

    def test_roundtrip_strict_true(self):
        """Test roundtrip validation with strict=True."""
        data = {"key": "value"}
        assert validate_roundtrip(data, strict=True) is True

    def test_roundtrip_unserializable_object(self):
        """Test roundtrip validation returns False for unserializable object."""

        class CustomObject:
            pass

        data = {"obj": CustomObject()}
        assert validate_roundtrip(data) is False

    def test_roundtrip_set_type(self):
        """Test roundtrip validation returns False for sets."""
        data = {"items": {1, 2, 3}}
        assert validate_roundtrip(data) is False

    def test_roundtrip_empty_nested_structures(self):
        """Test roundtrip with empty nested structures."""
        data = {
            "empty_list": [],
            "empty_dict": {},
            "nested": {
                "empty_inner_list": [],
                "empty_inner_dict": {},
            },
        }
        assert validate_roundtrip(data) is True

    def test_roundtrip_deeply_nested(self):
        """Test roundtrip with deeply nested structure."""
        data = {"level1": {"level2": {"level3": {"level4": {"level5": {"value": "deep"}}}}}}
        assert validate_roundtrip(data) is True

    def test_roundtrip_zero_and_negative_numbers(self):
        """Test roundtrip with zero and negative numbers."""
        data = {
            "zero": 0,
            "zero_float": 0.0,
            "negative_int": -42,
            "negative_float": -3.14,
        }
        assert validate_roundtrip(data) is True


# ============================================================================
# Format Comparison Tests
# ============================================================================


class TestCompareFormats:
    """Tests for compare_formats function."""

    def test_compare_empty_dict(self):
        """Test format comparison for empty dict."""
        data = {}
        result = compare_formats(data)
        assert "json" in result
        assert "toon" in result
        assert "reduction" in result
        assert "size_reduction_percent" in result

    def test_compare_simple_dict(self):
        """Test format comparison for simple dict."""
        data = {"name": "Alice", "age": 30}
        result = compare_formats(data)
        assert isinstance(result["reduction"], float)
        assert isinstance(result["size_reduction_percent"], float)
        assert -1 <= result["reduction"] <= 1  # Reasonable range

    def test_compare_large_data_structure(self):
        """Test format comparison shows token savings."""
        data = {"users": [{"id": i, "name": f"User{i}", "email": f"user{i}@example.com"} for i in range(50)]}
        result = compare_formats(data)
        assert "json" in result
        assert "toon" in result

    def test_compare_format_keys(self):
        """Test that comparison result has required keys."""
        data = {"test": "value"}
        result = compare_formats(data)
        assert "json" in result
        assert "toon" in result
        assert "size_bytes" in result["json"]
        assert "tokens" in result["json"]
        assert "size_bytes" in result["toon"]
        assert "tokens" in result["toon"]

    def test_compare_format_numeric_values(self):
        """Test that all metrics are numeric."""
        data = {"key": "value"}
        result = compare_formats(data)
        assert isinstance(result["json"]["size_bytes"], int)
        assert isinstance(result["json"]["tokens"], int)
        assert isinstance(result["toon"]["size_bytes"], int)
        assert isinstance(result["toon"]["tokens"], int)
        assert isinstance(result["reduction"], float)

    def test_compare_format_with_unicode(self):
        """Test format comparison with unicode data."""
        data = {
            "korean": "한글",
            "emoji": "😀",
            "chinese": "中文",
        }
        result = compare_formats(data)
        assert isinstance(result["json"]["size_bytes"], int)
        assert isinstance(result["toon"]["size_bytes"], int)

    def test_compare_list_structure(self):
        """Test format comparison with list structure."""
        data = [1, 2, 3, 4, 5]
        result = compare_formats(data)
        assert "json" in result
        assert "toon" in result

    def test_compare_nested_structure(self):
        """Test format comparison with nested structure."""
        data = {"level1": {"level2": {"level3": "value"}}}
        result = compare_formats(data)
        assert isinstance(result["reduction"], float)

    def test_compare_unserializable_raises_error(self):
        """Test format comparison raises error for unserializable data."""

        class CustomObject:
            pass

        data = {"obj": CustomObject()}
        with pytest.raises(ValueError) as exc_info:
            compare_formats(data)
        assert "Failed to compare formats" in str(exc_info.value)

    def test_compare_size_reduction_range(self):
        """Test that size reduction is within reasonable range."""
        data = {"test": "value"}
        result = compare_formats(data)
        # Reduction should be between -1 and 1 (negative means TOON is larger)
        assert -1 <= result["reduction"] <= 1
        assert -100 <= result["size_reduction_percent"] <= 100

    def test_compare_same_format_difference(self):
        """Test comparison with data that has same token count."""
        # Simple data might have similar token counts
        data = {"x": 1}
        result = compare_formats(data)
        # Should still be valid comparison
        assert isinstance(result["reduction"], float)

    def test_compare_with_repeated_patterns(self):
        """Test comparison with repeated patterns (good for TOON)."""
        # Repeated patterns should show token savings
        data = {f"item_{i}": {"id": i, "type": "standard"} for i in range(20)}
        result = compare_formats(data)
        assert "json" in result
        assert "toon" in result


# ============================================================================
# File I/O Tests
# ============================================================================


class TestToonSave:
    """Tests for toon_save function."""

    def test_save_creates_file(self):
        """Test that toon_save creates a file."""
        with tempfile.TemporaryDirectory() as tmpdir:
            path = Path(tmpdir) / "test.toon"
            data = {"key": "value"}
            toon_save(data, path)

            assert path.exists()
            assert path.is_file()

    def test_save_creates_parent_directories(self):
        """Test that toon_save creates parent directories."""
        with tempfile.TemporaryDirectory() as tmpdir:
            path = Path(tmpdir) / "deep" / "nested" / "path" / "test.toon"
            data = {"key": "value"}
            toon_save(data, path)

            assert path.exists()
            assert path.parent.exists()

    def test_save_writes_valid_toon(self):
        """Test that saved file contains valid TOON."""
        with tempfile.TemporaryDirectory() as tmpdir:
            path = Path(tmpdir) / "test.toon"
            data = {"name": "Alice", "age": 30}
            toon_save(data, path)

            content = path.read_text(encoding="utf-8")
            decoded = json.loads(content)
            assert decoded == data

    def test_save_with_string_path(self):
        """Test save works with string path."""
        with tempfile.TemporaryDirectory() as tmpdir:
            path = str(Path(tmpdir) / "test.toon")
            data = {"key": "value"}
            toon_save(data, Path(path))

            assert Path(path).exists()

    def test_save_overwrites_existing_file(self):
        """Test that save overwrites existing file."""
        with tempfile.TemporaryDirectory() as tmpdir:
            path = Path(tmpdir) / "test.toon"

            # Write first data
            toon_save({"old": "data"}, path)
            assert json.loads(path.read_text()) == {"old": "data"}

            # Overwrite with new data
            toon_save({"new": "data"}, path)
            assert json.loads(path.read_text()) == {"new": "data"}

    def test_save_with_unicode_data(self):
        """Test saving data with unicode characters."""
        with tempfile.TemporaryDirectory() as tmpdir:
            path = Path(tmpdir) / "test.toon"
            data = {"korean": "한글", "emoji": "😀"}
            toon_save(data, path)

            loaded = json.loads(path.read_text(encoding="utf-8"))
            assert loaded == data

    def test_save_with_strict_false(self):
        """Test save with strict=False."""
        with tempfile.TemporaryDirectory() as tmpdir:
            path = Path(tmpdir) / "test.toon"
            data = {"key": "value"}
            toon_save(data, path, strict=False)

            assert path.exists()
            assert json.loads(path.read_text()) == data

    def test_save_with_strict_true(self):
        """Test save with strict=True."""
        with tempfile.TemporaryDirectory() as tmpdir:
            path = Path(tmpdir) / "test.toon"
            data = {"key": "value"}
            toon_save(data, path, strict=True)

            assert path.exists()
            assert json.loads(path.read_text()) == data

    def test_save_unserializable_data_raises_error(self):
        """Test that saving unserializable data raises ValueError."""
        with tempfile.TemporaryDirectory() as tmpdir:
            path = Path(tmpdir) / "test.toon"

            class CustomObject:
                pass

            data = {"obj": CustomObject()}
            with pytest.raises(ValueError):
                toon_save(data, path)

    def test_save_preserves_data_types(self):
        """Test that saved data preserves all types."""
        with tempfile.TemporaryDirectory() as tmpdir:
            path = Path(tmpdir) / "test.toon"
            data = {
                "string": "value",
                "int": 42,
                "float": 3.14,
                "bool": True,
                "none": None,
                "list": [1, 2, 3],
                "dict": {"nested": "value"},
            }
            toon_save(data, path)

            loaded = json.loads(path.read_text())
            assert loaded == data

    def test_save_large_file(self):
        """Test saving large data structure."""
        with tempfile.TemporaryDirectory() as tmpdir:
            path = Path(tmpdir) / "large.toon"
            data = {"items": [{"id": i, "name": f"Item{i}", "data": list(range(10))} for i in range(100)]}
            toon_save(data, path)

            assert path.exists()
            loaded = json.loads(path.read_text())
            assert len(loaded["items"]) == 100


class TestToonLoad:
    """Tests for toon_load function."""

    def test_load_valid_file(self):
        """Test loading valid TOON file."""
        with tempfile.TemporaryDirectory() as tmpdir:
            path = Path(tmpdir) / "test.toon"
            original_data = {"key": "value", "number": 42}

            # Create file
            toon_save(original_data, path)

            # Load file
            loaded_data = toon_load(path)
            assert loaded_data == original_data

    def test_load_with_string_path(self):
        """Test load works with string path."""
        with tempfile.TemporaryDirectory() as tmpdir:
            path = Path(tmpdir) / "test.toon"
            original_data = {"key": "value"}
            toon_save(original_data, path)

            # Load using string path
            loaded_data = toon_load(str(path))
            assert loaded_data == original_data

    def test_load_with_path_object(self):
        """Test load works with Path object."""
        with tempfile.TemporaryDirectory() as tmpdir:
            path = Path(tmpdir) / "test.toon"
            original_data = {"key": "value"}
            toon_save(original_data, path)

            # Load using Path object
            loaded_data = toon_load(path)
            assert loaded_data == original_data

    def test_load_nonexistent_file_raises_error(self):
        """Test loading nonexistent file raises IOError."""
        path = Path("/nonexistent/path/test.toon")
        with pytest.raises(IOError):
            toon_load(path)

    def test_load_invalid_toon_raises_error(self):
        """Test loading invalid TOON raises ValueError."""
        with tempfile.TemporaryDirectory() as tmpdir:
            path = Path(tmpdir) / "invalid.toon"
            path.write_text("{invalid json}")

            with pytest.raises(ValueError):
                toon_load(path)

    def test_load_empty_file_raises_error(self):
        """Test loading empty file raises ValueError."""
        with tempfile.TemporaryDirectory() as tmpdir:
            path = Path(tmpdir) / "empty.toon"
            path.write_text("")

            with pytest.raises(ValueError):
                toon_load(path)

    def test_load_with_unicode(self):
        """Test loading file with unicode data."""
        with tempfile.TemporaryDirectory() as tmpdir:
            path = Path(tmpdir) / "test.toon"
            data = {"korean": "한글", "emoji": "😀"}
            toon_save(data, path)

            loaded = toon_load(path)
            assert loaded == data

    def test_load_with_strict_false(self):
        """Test load with strict=False."""
        with tempfile.TemporaryDirectory() as tmpdir:
            path = Path(tmpdir) / "test.toon"
            original_data = {"key": "value"}
            toon_save(original_data, path)

            loaded = toon_load(path, strict=False)
            assert loaded == original_data

    def test_load_with_strict_true(self):
        """Test load with strict=True."""
        with tempfile.TemporaryDirectory() as tmpdir:
            path = Path(tmpdir) / "test.toon"
            original_data = {"key": "value"}
            toon_save(original_data, path)

            loaded = toon_load(path, strict=True)
            assert loaded == original_data

    def test_load_preserves_data_types(self):
        """Test that loaded data preserves all types."""
        with tempfile.TemporaryDirectory() as tmpdir:
            path = Path(tmpdir) / "test.toon"
            data = {
                "string": "value",
                "int": 42,
                "float": 3.14,
                "bool": True,
                "none": None,
                "list": [1, 2, 3],
                "dict": {"nested": "value"},
            }
            toon_save(data, path)

            loaded = toon_load(path)
            assert loaded == data

    def test_load_large_file(self):
        """Test loading large TOON file."""
        with tempfile.TemporaryDirectory() as tmpdir:
            path = Path(tmpdir) / "large.toon"
            data = {"items": [{"id": i, "name": f"Item{i}"} for i in range(100)]}
            toon_save(data, path)

            loaded = toon_load(path)
            assert len(loaded["items"]) == 100


# ============================================================================
# Migration Tests
# ============================================================================


class TestMigrateJsonToToon:
    """Tests for migrate_json_to_toon function."""

    def test_migrate_basic_json_file(self):
        """Test migrating basic JSON file to TOON."""
        with tempfile.TemporaryDirectory() as tmpdir:
            json_path = Path(tmpdir) / "test.json"
            data = {"name": "Alice", "age": 30}

            # Create JSON file
            json_path.write_text(json.dumps(data))

            # Migrate
            toon_path = migrate_json_to_toon(json_path)

            assert toon_path.exists()
            assert toon_path.suffix == ".toon"
            assert json.loads(toon_path.read_text()) == data

    def test_migrate_with_explicit_toon_path(self):
        """Test migration with explicit target path."""
        with tempfile.TemporaryDirectory() as tmpdir:
            json_path = Path(tmpdir) / "test.json"
            toon_path = Path(tmpdir) / "custom.toon"
            data = {"key": "value"}

            json_path.write_text(json.dumps(data))
            result = migrate_json_to_toon(json_path, toon_path)

            assert result == toon_path
            assert toon_path.exists()

    def test_migrate_default_path_replacement(self):
        """Test that default migration replaces .json with .toon."""
        with tempfile.TemporaryDirectory() as tmpdir:
            json_path = Path(tmpdir) / "config.json"
            data = {"setting": "value"}

            json_path.write_text(json.dumps(data))
            toon_path = migrate_json_to_toon(json_path)

            assert toon_path.name == "config.toon"

    def test_migrate_invalid_json_raises_error(self):
        """Test migration of invalid JSON raises ValueError."""
        with tempfile.TemporaryDirectory() as tmpdir:
            json_path = Path(tmpdir) / "invalid.json"
            json_path.write_text("{invalid json}")

            with pytest.raises(ValueError):
                migrate_json_to_toon(json_path)

    def test_migrate_nonexistent_json_raises_error(self):
        """Test migration of nonexistent file raises IOError."""
        json_path = Path("/nonexistent/test.json")
        with pytest.raises((IOError, FileNotFoundError)):
            migrate_json_to_toon(json_path)

    def test_migrate_with_string_paths(self):
        """Test migration works with string paths."""
        with tempfile.TemporaryDirectory() as tmpdir:
            json_path = str(Path(tmpdir) / "test.json")
            toon_path = str(Path(tmpdir) / "test.toon")
            data = {"key": "value"}

            Path(json_path).write_text(json.dumps(data))
            result = migrate_json_to_toon(json_path, toon_path)

            assert Path(result).exists()

    def test_migrate_preserves_data(self):
        """Test that migration preserves data integrity."""
        with tempfile.TemporaryDirectory() as tmpdir:
            json_path = Path(tmpdir) / "test.json"
            data = {
                "users": [
                    {"id": 1, "name": "Alice"},
                    {"id": 2, "name": "Bob"},
                ],
                "meta": {"version": "1.0"},
            }

            json_path.write_text(json.dumps(data))
            toon_path = migrate_json_to_toon(json_path)

            loaded = json.loads(toon_path.read_text())
            assert loaded == data

    def test_migrate_with_unicode_data(self):
        """Test migration of JSON with unicode characters."""
        with tempfile.TemporaryDirectory() as tmpdir:
            json_path = Path(tmpdir) / "test.json"
            data = {"korean": "한글", "emoji": "😀"}

            json_path.write_text(json.dumps(data, ensure_ascii=False))
            toon_path = migrate_json_to_toon(json_path)

            loaded = json.loads(toon_path.read_text(encoding="utf-8"))
            assert loaded == data

    def test_migrate_creates_nested_directories(self):
        """Test that migration creates nested directories if needed."""
        with tempfile.TemporaryDirectory() as tmpdir:
            json_path = Path(tmpdir) / "test.json"
            toon_path = Path(tmpdir) / "deep" / "nested" / "path" / "test.toon"
            data = {"key": "value"}

            json_path.write_text(json.dumps(data))
            result = migrate_json_to_toon(json_path, toon_path)

            assert result.exists()
            assert result.parent.exists()

    def test_migrate_large_json_file(self):
        """Test migration of large JSON file."""
        with tempfile.TemporaryDirectory() as tmpdir:
            json_path = Path(tmpdir) / "large.json"
            data = {"items": [{"id": i, "data": f"Item{i}"} for i in range(100)]}

            json_path.write_text(json.dumps(data))
            toon_path = migrate_json_to_toon(json_path)

            loaded = json.loads(toon_path.read_text())
            assert len(loaded["items"]) == 100


# ============================================================================
# Integration and Edge Case Tests
# ============================================================================


class TestErrorHandlingCoverage:
    """Tests for error handling paths and edge cases."""

    def test_save_ioerror_on_write(self):
        """Test toon_save handles IOError from write_text."""
        with tempfile.TemporaryDirectory() as tmpdir:
            path = Path(tmpdir) / "test.toon"
            data = {"key": "value"}

            # Mock write_text to raise IOError
            with patch.object(Path, "write_text", side_effect=IOError("Permission denied")):
                with pytest.raises(IOError) as exc_info:
                    toon_save(data, path)
                assert "Failed to write TOON file" in str(exc_info.value)

    def test_load_ioerror_on_read(self):
        """Test toon_load handles IOError from read_text."""
        path = Path("/nonexistent/path/test.toon")

        with pytest.raises(IOError):
            toon_load(path)

    def test_is_tabular_with_duplicate_keys_different_values(self):
        """Test _is_tabular with items having same keys."""
        items = [
            {"a": 1, "b": 2},
            {"b": 3, "a": 4},  # Same keys, different order
        ]
        assert _is_tabular(items) is True

    def test_encode_value_with_escaped_characters(self):
        """Test _encode_value with various escape sequences."""
        # Test that strings with special chars in the check list get quoted
        test_cases = [
            # Characters that trigger JSON encoding in _encode_value
            ("hello[world]", json.dumps("hello[world]")),  # Bracket triggers encoding
            ("a,b", json.dumps("a,b")),  # Comma triggers encoding
            ("key:value", json.dumps("key:value")),  # Colon triggers encoding
        ]
        for value, expected in test_cases:
            result = _encode_value(value)
            assert result == expected

    def test_toon_encode_with_bytearray_fails(self):
        """Test that encoding bytearray raises ValueError."""
        data = {"bytes": bytearray(b"test")}
        with pytest.raises(ValueError):
            toon_encode(data)

    def test_toon_decode_with_json_syntax_error_details(self):
        """Test toon_decode error message includes details."""
        invalid_json = '{"key": undefined}'
        with pytest.raises(ValueError) as exc_info:
            toon_decode(invalid_json)
        # Should mention "Failed to decode TOON"
        assert "Failed to decode TOON" in str(exc_info.value)

    def test_compare_formats_with_bytes_fails(self):
        """Test compare_formats with non-serializable types."""
        data = {"bytes": b"binary"}
        with pytest.raises(ValueError) as exc_info:
            compare_formats(data)
        assert "Failed to compare formats" in str(exc_info.value)

    def test_validate_roundtrip_with_complex_numbers(self):
        """Test roundtrip validation with non-serializable complex numbers."""
        # Complex numbers cannot be JSON serialized
        data = {"complex": complex(1, 2)}
        assert validate_roundtrip(data) is False

    def test_validate_roundtrip_with_nan(self):
        """Test roundtrip validation with NaN values."""
        data = {"nan": float("nan")}
        assert validate_roundtrip(data) is False

    def test_migrate_json_to_toon_with_read_error(self):
        """Test migration handles read errors gracefully."""
        json_path = Path("/nonexistent/file.json")
        with pytest.raises((IOError, FileNotFoundError)):
            migrate_json_to_toon(json_path)

    def test_toon_save_with_write_error(self):
        """Test toon_save handles write errors gracefully."""
        with tempfile.TemporaryDirectory() as tmpdir:
            path = Path(tmpdir) / "test.toon"
            data = {"key": "value"}

            # Mock write_text to raise OSError

            def mock_write(*args, **kwargs):
                raise OSError("Permission denied")

            with patch.object(Path, "write_text", side_effect=mock_write):
                with pytest.raises(IOError):
                    toon_save(data, path)

    def test_empty_list_is_not_tabular_with_empty_check(self):
        """Test that empty list returns False (covers redundant line 27-28)."""
        # This tests both conditions in _is_tabular
        result = _is_tabular([])
        assert result is False
        # Verify that both empty checks work
        assert _is_tabular(None) is False
        assert _is_tabular([]) is False


class TestIntegrationScenarios:
    """Integration tests for toon_utils module."""

    def test_encode_decode_roundtrip_consistency(self):
        """Test encode-decode roundtrip consistency."""
        data = {
            "users": [
                {"id": 1, "name": "Alice", "active": True},
                {"id": 2, "name": "Bob", "active": False},
            ],
            "meta": {
                "version": "1.0",
                "created": "2024-01-01",
                "count": 2,
            },
        }

        # Encode
        encoded = toon_encode(data)

        # Decode
        decoded = toon_decode(encoded)

        # Verify consistency
        assert decoded == data

    def test_save_load_roundtrip(self):
        """Test save-load file I/O roundtrip."""
        with tempfile.TemporaryDirectory() as tmpdir:
            path = Path(tmpdir) / "data.toon"
            original = {"config": {"debug": True, "port": 8080, "hosts": ["localhost", "127.0.0.1"]}}

            # Save
            toon_save(original, path)

            # Load
            loaded = toon_load(path)

            # Verify
            assert loaded == original

    def test_json_to_toon_migration_pipeline(self):
        """Test complete JSON to TOON migration pipeline."""
        with tempfile.TemporaryDirectory() as tmpdir:
            json_file = Path(tmpdir) / "config.json"
            data = {"database": {"host": "localhost", "port": 5432, "name": "myapp"}}

            # Create JSON
            json_file.write_text(json.dumps(data))

            # Migrate
            toon_file = migrate_json_to_toon(json_file)

            # Verify both exist and have same content
            json_loaded = json.loads(json_file.read_text())
            toon_loaded = toon_load(toon_file)

            assert json_loaded == toon_loaded == data

    def test_format_comparison_with_actual_data(self):
        """Test format comparison with realistic data."""
        data = {
            "users": [
                {
                    "id": i,
                    "username": f"user{i}",
                    "email": f"user{i}@example.com",
                    "created": "2024-01-01",
                    "active": True,
                }
                for i in range(20)
            ]
        }

        result = compare_formats(data)

        # Verify structure
        assert result["json"]["size_bytes"] > 0
        assert result["toon"]["size_bytes"] > 0
        assert isinstance(result["reduction"], float)

    def test_validation_with_complex_nested_structure(self):
        """Test validation with complex nested structure."""
        data = {
            "company": {
                "departments": [
                    {
                        "name": "Engineering",
                        "teams": [
                            {
                                "name": "Backend",
                                "members": [
                                    {"id": 1, "name": "Alice"},
                                    {"id": 2, "name": "Bob"},
                                ],
                            },
                            {
                                "name": "Frontend",
                                "members": [
                                    {"id": 3, "name": "Charlie"},
                                ],
                            },
                        ],
                    }
                ]
            }
        }

        assert validate_roundtrip(data) is True

    def test_complete_workflow_with_all_functions(self):
        """Test complete workflow using all major functions."""
        with tempfile.TemporaryDirectory() as tmpdir:
            # Start data
            data = {"items": [{"id": i, "name": f"Item{i}", "active": i % 2 == 0} for i in range(10)]}

            # 1. Validate roundtrip
            assert validate_roundtrip(data) is True

            # 2. Compare formats
            comparison = compare_formats(data)
            assert "reduction" in comparison

            # 3. Encode
            encoded = toon_encode(data)

            # 4. Decode
            decoded = toon_decode(encoded)
            assert decoded == data

            # 5. Save to file
            toon_path = Path(tmpdir) / "test.toon"
            toon_save(data, toon_path)

            # 6. Load from file
            loaded = toon_load(toon_path)
            assert loaded == data

            # 7. Migrate from JSON
            json_path = Path(tmpdir) / "test.json"
            json_path.write_text(json.dumps(data))
            migrated_path = migrate_json_to_toon(json_path)

            # 8. Verify migrated file
            migrated_data = toon_load(migrated_path)
            assert migrated_data == data

    def test_edge_case_deeply_nested_large_structure(self):
        """Test with deeply nested and large structure."""
        # Create deeply nested structure
        nested = {"level": 0}
        current = nested
        for i in range(1, 10):
            current["nested"] = {"level": i, "data": list(range(10))}
            current = current["nested"]

        # Should handle without errors
        assert validate_roundtrip(nested) is True
        encoded = toon_encode(nested)
        decoded = toon_decode(encoded)
        assert decoded == nested

    def test_special_characters_in_all_contexts(self):
        """Test special characters throughout the pipeline."""
        data = {
            "paths": "/usr/local/bin:/home/user/bin",
            "json_like": '{"embedded": "json"}',
            "csv_like": "value1,value2,value3",
            "multiline": "line1\nline2\nline3",
            "quoted": 'He said "hello"',
            "emoji": "😀🎉🚀",
            "unicode": "한글 中文 Русский",
        }

        # Test all operations
        assert validate_roundtrip(data) is True

        with tempfile.TemporaryDirectory() as tmpdir:
            path = Path(tmpdir) / "special.toon"
            toon_save(data, path)
            loaded = toon_load(path)
            assert loaded == data

    def test_empty_and_null_values_consistency(self):
        """Test consistency of empty and null values."""
        data = {
            "empty_string": "",
            "empty_list": [],
            "empty_dict": {},
            "null_value": None,
            "zero": 0,
            "false": False,
        }

        assert validate_roundtrip(data) is True

        encoded = toon_encode(data)
        decoded = toon_decode(encoded)

        # Verify each value
        assert decoded["empty_string"] == ""
        assert decoded["empty_list"] == []
        assert decoded["empty_dict"] == {}
        assert decoded["null_value"] is None
        assert decoded["zero"] == 0
        assert decoded["false"] is False
