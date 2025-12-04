"""Comprehensive coverage tests for moai_adk.utils.toon_utils module.

This module contains tests targeting uncovered code paths in toon_utils.py
to achieve 85%+ coverage.
"""

import json
from pathlib import Path
from unittest.mock import MagicMock, patch

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


class TestIsTabular:
    """Test _is_tabular helper function."""

    def test_is_tabular_empty_list(self):
        """Test _is_tabular with empty list."""
        assert _is_tabular([]) is False

    def test_is_tabular_non_list(self):
        """Test _is_tabular with non-list input."""
        assert _is_tabular("not a list") is False
        assert _is_tabular(None) is False
        assert _is_tabular(42) is False

    def test_is_tabular_list_of_dicts_same_keys(self):
        """Test _is_tabular with uniform dict list."""
        data = [
            {"id": 1, "name": "Alice"},
            {"id": 2, "name": "Bob"},
            {"id": 3, "name": "Charlie"},
        ]
        assert _is_tabular(data) is True

    def test_is_tabular_list_of_dicts_different_keys(self):
        """Test _is_tabular with non-uniform dict list."""
        data = [
            {"id": 1, "name": "Alice"},
            {"id": 2, "email": "bob@example.com"},
        ]
        assert _is_tabular(data) is False

    def test_is_tabular_list_of_non_dicts(self):
        """Test _is_tabular with list of non-dicts."""
        assert _is_tabular([1, 2, 3]) is False
        assert _is_tabular(["a", "b", "c"]) is False

    def test_is_tabular_mixed_types(self):
        """Test _is_tabular with mixed types."""
        data = [
            {"id": 1},
            "string",
        ]
        assert _is_tabular(data) is False

    def test_is_tabular_single_dict(self):
        """Test _is_tabular with single dict."""
        data = [{"id": 1, "name": "Alice"}]
        assert _is_tabular(data) is True

    def test_is_tabular_different_number_of_keys(self):
        """Test _is_tabular with different number of keys."""
        data = [
            {"id": 1, "name": "Alice", "age": 30},
            {"id": 2, "name": "Bob"},
        ]
        assert _is_tabular(data) is False


class TestEncodeValue:
    """Test _encode_value helper function."""

    def test_encode_value_none(self):
        """Test encoding None."""
        assert _encode_value(None) == "null"

    def test_encode_value_bool_true(self):
        """Test encoding True."""
        assert _encode_value(True) == "true"

    def test_encode_value_bool_false(self):
        """Test encoding False."""
        assert _encode_value(False) == "false"

    def test_encode_value_integer(self):
        """Test encoding integer."""
        assert _encode_value(42) == "42"
        assert _encode_value(0) == "0"
        assert _encode_value(-10) == "-10"

    def test_encode_value_float(self):
        """Test encoding float."""
        assert _encode_value(3.14) == "3.14"
        assert _encode_value(0.0) == "0.0"

    def test_encode_value_simple_string(self):
        """Test encoding simple string."""
        assert _encode_value("hello") == "hello"
        assert _encode_value("world") == "world"

    def test_encode_value_string_with_comma(self):
        """Test encoding string with special characters."""
        result = _encode_value("hello,world")
        assert result.startswith('"')  # Should be quoted

    def test_encode_value_string_with_colon(self):
        """Test encoding string with colon."""
        result = _encode_value("key:value")
        assert result.startswith('"')

    def test_encode_value_string_with_newline(self):
        """Test encoding string with newline."""
        result = _encode_value("line1\nline2")
        assert result.startswith('"')

    def test_encode_value_string_with_quote(self):
        """Test encoding string with quote."""
        result = _encode_value('say "hello"')
        assert result.startswith('"')

    def test_encode_value_string_with_bracket(self):
        """Test encoding string with bracket."""
        result = _encode_value("array[0]")
        assert result.startswith('"')

    def test_encode_value_complex_object(self):
        """Test encoding complex object."""
        obj = {"nested": "object"}
        result = _encode_value(obj)
        assert '"nested"' in result

    def test_encode_value_list(self):
        """Test encoding list."""
        lst = [1, 2, 3]
        result = _encode_value(lst)
        assert "[" in result


class TestToonEncodeStrict:
    """Test toon_encode with strict parameter."""

    def test_encode_with_strict_false(self):
        """Test encode with strict=False."""
        data = {"test": "data"}
        result = toon_encode(data, strict=False)
        assert isinstance(result, str)
        assert "test" in result

    def test_encode_with_strict_true(self):
        """Test encode with strict=True."""
        data = {"test": "data"}
        result = toon_encode(data, strict=True)
        assert isinstance(result, str)


class TestToonEncodeDetectTabular:
    """Test toon_encode with detect_tabular parameter."""

    def test_encode_with_detect_tabular_false(self):
        """Test encode with detect_tabular=False."""
        data = {"users": [{"id": 1, "name": "Alice"}]}
        result = toon_encode(data, detect_tabular=False)
        assert isinstance(result, str)

    def test_encode_with_detect_tabular_true(self):
        """Test encode with detect_tabular=True."""
        data = {"users": [{"id": 1, "name": "Alice"}]}
        result = toon_encode(data, detect_tabular=True)
        assert isinstance(result, str)


class TestToonDecodeStrict:
    """Test toon_decode with strict parameter."""

    def test_decode_with_strict_false(self):
        """Test decode with strict=False."""
        toon_str = '{"test": "data"}'
        result = toon_decode(toon_str, strict=False)
        assert result["test"] == "data"

    def test_decode_with_strict_true(self):
        """Test decode with strict=True."""
        toon_str = '{"test": "data"}'
        result = toon_decode(toon_str, strict=True)
        assert result["test"] == "data"


class TestToonDecodeErrors:
    """Test toon_decode error handling."""

    def test_decode_invalid_json_syntax(self):
        """Test decode with various invalid JSON."""
        with pytest.raises(ValueError):
            toon_decode("{invalid json")

    def test_decode_unclosed_brace(self):
        """Test decode with unclosed brace."""
        with pytest.raises(ValueError):
            toon_decode('{"key": "value"')

    def test_decode_unclosed_bracket(self):
        """Test decode with unclosed bracket."""
        with pytest.raises(ValueError):
            toon_decode("[1, 2, 3")

    def test_decode_trailing_comma(self):
        """Test decode with trailing comma."""
        with pytest.raises(ValueError):
            toon_decode('{"key": "value",}')


class TestToonSaveCreateDirectories:
    """Test toon_save creates parent directories."""

    def test_save_creates_nested_directories(self):
        """Test save creates nested directory structure."""
        with patch("pathlib.Path.mkdir") as mock_mkdir:
            with patch("pathlib.Path.write_text"):
                data = {"test": "data"}
                toon_save(data, Path("/tmp/a/b/c/test.toon"))
                assert mock_mkdir.called

    def test_save_with_string_path(self):
        """Test save accepts string path."""
        with patch("pathlib.Path.mkdir"):
            with patch("pathlib.Path.write_text"):
                data = {"test": "data"}
                toon_save(data, "/tmp/test.toon")
                # Should not raise

    def test_save_handles_write_error(self):
        """Test save propagates IOError."""
        with patch("pathlib.Path.mkdir"):
            with patch("pathlib.Path.write_text", side_effect=IOError("Write failed")):
                with pytest.raises(IOError):
                    toon_save({"test": "data"}, Path("/tmp/test.toon"))

    def test_save_encodes_before_writing(self):
        """Test save encodes data before writing."""
        with patch("pathlib.Path.mkdir"):
            with patch("pathlib.Path.write_text") as mock_write:
                data = {"key": "value", "number": 42}
                toon_save(data, Path("/tmp/test.toon"))
                assert mock_write.called
                written_content = mock_write.call_args[0][0]
                assert "key" in written_content


class TestToonLoadErrors:
    """Test toon_load error handling."""

    def test_load_invalid_json_file(self):
        """Test load with invalid JSON in file."""
        with patch("pathlib.Path.read_text", return_value="{invalid"):
            with pytest.raises(ValueError):
                toon_load(Path("/tmp/test.toon"))

    def test_load_file_not_found(self):
        """Test load when file doesn't exist."""
        with patch("pathlib.Path.read_text", side_effect=IOError("File not found")):
            with pytest.raises(IOError):
                toon_load(Path("/tmp/nonexistent.toon"))

    def test_load_with_string_path(self):
        """Test load accepts string path."""
        data = {"test": "data"}
        json_str = json.dumps(data)
        with patch("pathlib.Path.read_text", return_value=json_str):
            result = toon_load("/tmp/test.toon")
            assert result == data


class TestValidateRoundtripEdgeCases:
    """Test validate_roundtrip with edge cases."""

    def test_roundtrip_special_characters(self):
        """Test roundtrip with special characters."""
        data = {
            "text": "Hello\nWorld\t!",
            "quote": 'Say "Hello"',
            "unicode": "你好世界",
        }
        assert validate_roundtrip(data) is True

    def test_roundtrip_large_numbers(self):
        """Test roundtrip with large numbers."""
        data = {
            "big_int": 999999999999999,
            "big_float": 1.123456789e10,
        }
        assert validate_roundtrip(data) is True

    def test_roundtrip_deeply_nested(self):
        """Test roundtrip with deeply nested structure."""
        data = {
            "level1": {
                "level2": {
                    "level3": {
                        "value": "deep"
                    }
                }
            }
        }
        assert validate_roundtrip(data) is True

    def test_roundtrip_large_list(self):
        """Test roundtrip with large list."""
        data = {"items": list(range(1000))}
        assert validate_roundtrip(data) is True

    def test_roundtrip_non_encodable_returns_false(self):
        """Test roundtrip with non-encodable object."""
        class CustomClass:
            pass

        data = {"obj": CustomClass()}
        assert validate_roundtrip(data) is False

    def test_roundtrip_empty_nested_structures(self):
        """Test roundtrip with empty nested structures."""
        data = {
            "empty_dict": {},
            "empty_list": [],
            "nested": {
                "also_empty": {}
            }
        }
        assert validate_roundtrip(data) is True


class TestCompareFormatsMetrics:
    """Test compare_formats return values."""

    def test_compare_formats_all_fields(self):
        """Test compare_formats returns all required fields."""
        data = {"test": "data"}
        result = compare_formats(data)
        assert "json" in result
        assert "toon" in result
        assert "reduction" in result
        assert "size_reduction_percent" in result

    def test_compare_formats_json_metrics(self):
        """Test JSON metrics in compare result."""
        data = {"test": "data"}
        result = compare_formats(data)
        json_metrics = result["json"]
        assert "size_bytes" in json_metrics
        assert "tokens" in json_metrics
        assert json_metrics["size_bytes"] > 0
        assert json_metrics["tokens"] >= 0

    def test_compare_formats_toon_metrics(self):
        """Test TOON metrics in compare result."""
        data = {"test": "data"}
        result = compare_formats(data)
        toon_metrics = result["toon"]
        assert "size_bytes" in toon_metrics
        assert "tokens" in toon_metrics
        assert toon_metrics["size_bytes"] > 0
        assert toon_metrics["tokens"] >= 0

    def test_compare_formats_reduction_percentage(self):
        """Test reduction percentage is calculated correctly."""
        data = {"test": "data"}
        result = compare_formats(data)
        reduction = result["reduction"]
        assert isinstance(reduction, float)
        # Reduction can be negative if TOON is larger than JSON
        assert -2.0 <= reduction <= 1.1

    def test_compare_formats_size_reduction_percent(self):
        """Test size reduction percent is calculated."""
        data = {"test": "data"}
        result = compare_formats(data)
        size_reduction = result["size_reduction_percent"]
        assert isinstance(size_reduction, float)
        # Can be negative if TOON is larger than JSON
        assert -100 <= size_reduction <= 100

    def test_compare_formats_empty_data(self):
        """Test compare with empty structures."""
        result = compare_formats({})
        assert isinstance(result, dict)
        assert "reduction" in result

    def test_compare_formats_large_data(self):
        """Test compare with large data."""
        data = {
            "items": [
                {"id": i, "name": f"Item{i}", "value": i * 10}
                for i in range(100)
            ]
        }
        result = compare_formats(data)
        assert result["json"]["size_bytes"] > 0
        assert result["toon"]["size_bytes"] > 0


class TestMigrateJsonToToon:
    """Test migrate_json_to_toon function."""

    def test_migrate_basic_json(self):
        """Test migrating basic JSON file."""
        json_data = {"test": "data", "number": 42}
        json_str = json.dumps(json_data)

        with patch("pathlib.Path.read_text", return_value=json_str):
            with patch("pathlib.Path.with_suffix") as mock_with_suffix:
                toon_path = MagicMock()
                toon_path.__truediv__ = MagicMock(return_value=toon_path)
                mock_with_suffix.return_value = toon_path

                with patch("moai_adk.utils.toon_utils.toon_save") as mock_save:
                    migrate_json_to_toon(Path("data.json"))
                    assert mock_save.called

    def test_migrate_with_custom_target(self):
        """Test migrate with custom target path."""
        json_data = {"test": "data"}
        json_str = json.dumps(json_data)

        with patch("pathlib.Path.read_text", return_value=json_str):
            with patch("moai_adk.utils.toon_utils.toon_save") as mock_save:
                result = migrate_json_to_toon(Path("data.json"), Path("custom.toon"))
                assert result == Path("custom.toon")

    def test_migrate_invalid_json(self):
        """Test migrate with invalid JSON."""
        with patch("pathlib.Path.read_text", return_value="{invalid"):
            with pytest.raises(ValueError):
                migrate_json_to_toon(Path("bad.json"))

    def test_migrate_returns_toon_path(self):
        """Test migrate returns Path object."""
        json_data = {"test": "data"}
        json_str = json.dumps(json_data)

        with patch("pathlib.Path.read_text", return_value=json_str):
            with patch("pathlib.Path.with_suffix", return_value=Path("data.toon")):
                with patch("moai_adk.utils.toon_utils.toon_save"):
                    result = migrate_json_to_toon(Path("data.json"))
                    assert isinstance(result, Path)

    def test_migrate_string_path(self):
        """Test migrate accepts string paths."""
        json_data = {"test": "data"}
        json_str = json.dumps(json_data)

        with patch("pathlib.Path.read_text", return_value=json_str):
            with patch("pathlib.Path.with_suffix") as mock_with_suffix:
                toon_path = MagicMock()
                mock_with_suffix.return_value = toon_path

                with patch("moai_adk.utils.toon_utils.toon_save"):
                    migrate_json_to_toon("data.json")
                    # Should convert string to Path


class TestToonIntegrationComplex:
    """Integration tests with complex data structures."""

    def test_complex_nested_roundtrip(self):
        """Test roundtrip with complex nested structure."""
        data = {
            "users": [
                {
                    "id": 1,
                    "name": "Alice",
                    "roles": ["admin", "user"],
                    "metadata": {
                        "created": "2023-01-01",
                        "active": True,
                    }
                },
                {
                    "id": 2,
                    "name": "Bob",
                    "roles": ["user"],
                    "metadata": {
                        "created": "2023-02-01",
                        "active": False,
                    }
                }
            ],
            "total": 2,
            "timestamp": "2023-12-01T12:00:00Z"
        }

        encoded = toon_encode(data)
        decoded = toon_decode(encoded)
        assert decoded == data

    def test_all_json_types_roundtrip(self):
        """Test roundtrip with all JSON types."""
        data = {
            "string": "value",
            "integer": 42,
            "float": 3.14159,
            "boolean_true": True,
            "boolean_false": False,
            "null_value": None,
            "array": [1, "two", 3.0, None, True],
            "object": {
                "nested": "value",
                "number": 100
            }
        }
        assert validate_roundtrip(data) is True

    def test_unicode_handling(self):
        """Test roundtrip with various unicode characters."""
        data = {
            "english": "Hello World",
            "chinese": "你好世界",
            "arabic": "مرحبا بالعالم",
            "emoji": "🎉🎊🎈",
            "mixed": "Hello 世界 🌍"
        }
        assert validate_roundtrip(data) is True

    def test_numeric_precision(self):
        """Test roundtrip preserves numeric precision."""
        data = {
            "pi": 3.141592653589793,
            "large_int": 1234567890123456789,
            "small_float": 0.000000001,
            "scientific": 1.23e-10
        }
        assert validate_roundtrip(data) is True


class TestToonSaveLoadRoundtrip:
    """Test save and load roundtrip."""

    def test_save_load_roundtrip(self):
        """Test data survives save and load cycle."""
        data = {"test": "data", "number": 42}

        with patch("pathlib.Path.mkdir"):
            with patch("pathlib.Path.write_text") as mock_write:
                toon_save(data, Path("/tmp/test.toon"))
                # Capture what was written
                written_content = mock_write.call_args[0][0]

        with patch("pathlib.Path.read_text", return_value=written_content):
            loaded = toon_load(Path("/tmp/test.toon"))
            assert loaded == data
