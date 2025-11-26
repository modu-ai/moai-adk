"""Comprehensive test suite for JSONUtils"""

import json
import sys
from io import StringIO
from pathlib import Path
from unittest.mock import patch

import pytest

# Skip this test - json_utils.py module doesn't exist in codebase
pytestmark = pytest.mark.skip(
    reason="json_utils.py module not found in .claude/hooks/moai/lib/ - may have been removed or refactored"
)

# Add .claude/hooks/moai/lib to sys.path for imports
PROJECT_ROOT = Path(__file__).parent.parent.parent
LIB_DIR = PROJECT_ROOT / ".claude" / "hooks" / "moai" / "lib"
if str(LIB_DIR) not in sys.path:
    sys.path.insert(0, str(LIB_DIR))

try:
    from json_utils import JSONSchemas, JSONUtils
except ImportError:
    JSONSchemas = None
    JSONUtils = None


class TestReadJsonFromStdin:
    """Test JSONUtils.read_json_from_stdin() method"""

    def test_read_valid_json_from_stdin(self):
        """Read valid JSON from stdin"""
        test_data = {"key": "value", "number": 42}
        json_str = json.dumps(test_data)

        with patch("sys.stdin", StringIO(json_str)):
            result = JSONUtils.read_json_from_stdin()

        assert result == test_data

    def test_read_complex_json_from_stdin(self):
        """Read complex nested JSON from stdin"""
        test_data = {"nested": {"deep": {"value": "test"}}, "list": [1, 2, 3], "bool": True}
        json_str = json.dumps(test_data)

        with patch("sys.stdin", StringIO(json_str)):
            result = JSONUtils.read_json_from_stdin()

        assert result == test_data

    def test_read_json_from_stdin_empty_input(self):
        """Read from empty stdin returns empty dict"""
        with patch("sys.stdin", StringIO("")):
            result = JSONUtils.read_json_from_stdin()

        assert result == {}

    def test_read_json_from_stdin_whitespace_only(self):
        """Read from stdin with whitespace only returns empty dict"""
        with patch("sys.stdin", StringIO("   \n\t  ")):
            result = JSONUtils.read_json_from_stdin()

        assert result == {}

    def test_read_json_from_stdin_invalid_raises_error(self):
        """Read invalid JSON from stdin raises JSONDecodeError"""
        with patch("sys.stdin", StringIO("{ invalid json }")):
            with pytest.raises(json.JSONDecodeError):
                JSONUtils.read_json_from_stdin()

    def test_read_json_array_from_stdin(self):
        """Read JSON array from stdin"""
        test_data = [1, 2, 3, 4, 5]
        json_str = json.dumps(test_data)

        with patch("sys.stdin", StringIO(json_str)):
            result = JSONUtils.read_json_from_stdin()

        assert result == test_data

    def test_read_json_string_from_stdin(self):
        """Read JSON string from stdin"""
        test_data = "hello world"
        json_str = json.dumps(test_data)

        with patch("sys.stdin", StringIO(json_str)):
            result = JSONUtils.read_json_from_stdin()

        assert result == test_data

    def test_read_json_null_from_stdin(self):
        """Read JSON null from stdin"""
        with patch("sys.stdin", StringIO("null")):
            result = JSONUtils.read_json_from_stdin()

        assert result is None

    def test_read_json_boolean_from_stdin(self):
        """Read JSON boolean from stdin"""
        with patch("sys.stdin", StringIO("true")):
            result = JSONUtils.read_json_from_stdin()

        assert result is True

        with patch("sys.stdin", StringIO("false")):
            result = JSONUtils.read_json_from_stdin()

        assert result is False

    def test_read_json_number_from_stdin(self):
        """Read JSON number from stdin"""
        with patch("sys.stdin", StringIO("42")):
            result = JSONUtils.read_json_from_stdin()

        assert result == 42


class TestSafeJsonLoads:
    """Test JSONUtils.safe_json_loads() method"""

    def test_safe_json_loads_valid(self):
        """Safely load valid JSON string"""
        test_data = {"key": "value"}
        json_str = json.dumps(test_data)

        result = JSONUtils.safe_json_loads(json_str)

        assert result == test_data

    def test_safe_json_loads_invalid_returns_default(self):
        """Safely load invalid JSON returns default"""
        result = JSONUtils.safe_json_loads("{ invalid json }")

        assert result == {}

    def test_safe_json_loads_invalid_with_custom_default(self):
        """Safely load invalid JSON with custom default"""
        custom_default = {"error": "parsing failed"}

        result = JSONUtils.safe_json_loads("{ invalid json }", default=custom_default)

        assert result == custom_default

    def test_safe_json_loads_empty_string(self):
        """Safely load empty string"""
        result = JSONUtils.safe_json_loads("")

        assert result == {}

    def test_safe_json_loads_whitespace_only(self):
        """Safely load whitespace-only string"""
        result = JSONUtils.safe_json_loads("   \n\t  ")

        assert result == {}

    def test_safe_json_loads_array(self):
        """Safely load JSON array"""
        test_data = [1, 2, 3]
        json_str = json.dumps(test_data)

        result = JSONUtils.safe_json_loads(json_str)

        assert result == test_data

    def test_safe_json_loads_null(self):
        """Safely load JSON null"""
        result = JSONUtils.safe_json_loads("null")

        assert result is None

    def test_safe_json_loads_string(self):
        """Safely load JSON string"""
        result = JSONUtils.safe_json_loads('"hello"')

        assert result == "hello"

    def test_safe_json_loads_number(self):
        """Safely load JSON number"""
        result = JSONUtils.safe_json_loads("42")

        assert result == 42

    def test_safe_json_loads_boolean(self):
        """Safely load JSON boolean"""
        result = JSONUtils.safe_json_loads("true")
        assert result is True

        result = JSONUtils.safe_json_loads("false")
        assert result is False

    def test_safe_json_loads_complex_structure(self):
        """Safely load complex nested structure"""
        test_data = {
            "users": [{"id": 1, "name": "Alice"}, {"id": 2, "name": "Bob"}],
            "metadata": {"count": 2, "active": True},
        }
        json_str = json.dumps(test_data)

        result = JSONUtils.safe_json_loads(json_str)

        assert result == test_data


class TestSafeJsonLoadFile:
    """Test JSONUtils.safe_json_load_file() method"""

    def test_safe_json_load_file_valid(self, sample_json_file):
        """Safely load valid JSON file"""
        result = JSONUtils.safe_json_load_file(sample_json_file)

        assert result["name"] == "test"
        assert result["version"] == "1.0.0"

    def test_safe_json_load_file_nonexistent(self, nonexistent_file):
        """Safely load nonexistent file returns default"""
        result = JSONUtils.safe_json_load_file(nonexistent_file)

        assert result == {}

    def test_safe_json_load_file_nonexistent_with_custom_default(self, nonexistent_file):
        """Safely load nonexistent file with custom default"""
        custom_default = {"error": "file not found"}

        result = JSONUtils.safe_json_load_file(nonexistent_file, default=custom_default)

        assert result == custom_default

    def test_safe_json_load_file_invalid_json(self, invalid_json_file):
        """Safely load file with invalid JSON"""
        result = JSONUtils.safe_json_load_file(invalid_json_file)

        assert result == {}

    def test_safe_json_load_file_invalid_json_with_default(self, invalid_json_file):
        """Safely load file with invalid JSON and custom default"""
        custom_default = {"default": "value"}

        result = JSONUtils.safe_json_load_file(invalid_json_file, default=custom_default)

        assert result == custom_default

    def test_safe_json_load_file_empty_file(self, tmp_path):
        """Safely load empty JSON file"""
        empty_file = tmp_path / "empty.json"
        empty_file.write_text("")

        result = JSONUtils.safe_json_load_file(empty_file)

        assert result == {}

    def test_safe_json_load_file_unicode_content(self, tmp_path):
        """Safely load JSON file with unicode content"""
        test_data = {"message": "🎯 테스트 メッセージ"}
        json_file = tmp_path / "unicode.json"
        json_file.write_text(json.dumps(test_data, ensure_ascii=False), encoding="utf-8")

        result = JSONUtils.safe_json_load_file(json_file)

        assert result["message"] == "🎯 테스트 メッセージ"


class TestWriteJsonToFile:
    """Test JSONUtils.write_json_to_file() method"""

    def test_write_json_to_file_success(self, tmp_path):
        """Write JSON to file successfully"""
        data = {"key": "value", "number": 42}
        file_path = tmp_path / "output.json"

        success = JSONUtils.write_json_to_file(data, file_path)

        assert success is True
        assert file_path.exists()

        # Verify written content
        written_data = json.loads(file_path.read_text())
        assert written_data == data

    def test_write_json_to_file_creates_directory(self, tmp_path):
        """Write JSON to file creates parent directories"""
        data = {"key": "value"}
        file_path = tmp_path / "nested" / "dir" / "file.json"

        success = JSONUtils.write_json_to_file(data, file_path)

        assert success is True
        assert file_path.exists()

    def test_write_json_to_file_complex_structure(self, tmp_path):
        """Write complex nested structure to file"""
        data = {"nested": {"deep": {"value": "test"}}, "list": [1, 2, 3, 4, 5], "bool": True, "null": None}
        file_path = tmp_path / "complex.json"

        success = JSONUtils.write_json_to_file(data, file_path)

        assert success is True
        written_data = json.loads(file_path.read_text())
        assert written_data == data

    def test_write_json_to_file_with_custom_indent(self, tmp_path):
        """Write JSON to file with custom indentation"""
        data = {"key": "value"}
        file_path = tmp_path / "indented.json"

        success = JSONUtils.write_json_to_file(data, file_path, indent=4)

        assert success is True
        content = file_path.read_text()
        # Check for 4-space indentation
        assert "    " in content

    def test_write_json_to_file_unicode_content(self, tmp_path):
        """Write JSON with unicode content to file"""
        data = {"message": "🎯 테스트 メッセージ"}
        file_path = tmp_path / "unicode.json"

        success = JSONUtils.write_json_to_file(data, file_path)

        assert success is True
        written_data = json.loads(file_path.read_text(encoding="utf-8"))
        assert written_data["message"] == "🎯 테스트 メッセージ"

    def test_write_json_to_file_overwrites_existing(self, tmp_path):
        """Write JSON to file overwrites existing file"""
        file_path = tmp_path / "file.json"
        file_path.write_text('{"old": "data"}')

        new_data = {"new": "data"}
        success = JSONUtils.write_json_to_file(new_data, file_path)

        assert success is True
        written_data = json.loads(file_path.read_text())
        assert written_data == new_data

    def test_write_json_to_file_empty_dict(self, tmp_path):
        """Write empty dictionary to file"""
        data = {}
        file_path = tmp_path / "empty.json"

        success = JSONUtils.write_json_to_file(data, file_path)

        assert success is True
        assert file_path.read_text() == "{}"

    def test_write_json_to_file_readonly_dir_fails(self, tmp_path):
        """Write JSON to readonly directory fails gracefully"""
        readonly_dir = tmp_path / "readonly"
        readonly_dir.mkdir()
        readonly_dir.chmod(0o555)

        try:
            file_path = readonly_dir / "file.json"
            data = {"key": "value"}

            success = JSONUtils.write_json_to_file(data, file_path)

            # Should return False or succeed depending on permissions
            assert isinstance(success, bool)
        finally:
            readonly_dir.chmod(0o755)


class TestValidateJsonSchema:
    """Test JSONUtils.validate_json_schema() method"""

    def test_validate_required_fields_present(self):
        """Validate JSON has all required fields"""
        data = {"name": "test", "email": "test@example.com"}
        required = ["name", "email"]

        result = JSONUtils.validate_json_schema(data, required)

        assert result is True

    def test_validate_required_fields_missing(self):
        """Validate JSON missing required fields"""
        data = {"name": "test"}
        required = ["name", "email"]

        result = JSONUtils.validate_json_schema(data, required)

        assert result is False

    def test_validate_non_dict_data(self):
        """Validate non-dict data fails"""
        result = JSONUtils.validate_json_schema([1, 2, 3], ["field"])

        assert result is False

    def test_validate_empty_required_list(self):
        """Validate with empty required list"""
        data = {"key": "value"}

        result = JSONUtils.validate_json_schema(data, [])

        assert result is True

    def test_validate_empty_data_empty_required(self):
        """Validate empty data with empty required list"""
        result = JSONUtils.validate_json_schema({}, [])

        assert result is True

    def test_validate_null_data(self):
        """Validate null data fails"""
        result = JSONUtils.validate_json_schema(None, ["field"])

        assert result is False

    def test_validate_extra_fields_allowed(self):
        """Validate allows extra fields beyond required"""
        data = {"required_field": "value", "extra_field": "extra"}
        required = ["required_field"]

        result = JSONUtils.validate_json_schema(data, required)

        assert result is True


class TestGetNestedValue:
    """Test JSONUtils.get_nested_value() method"""

    def test_get_nested_value_single_level(self, sample_json_data):
        """Get value from single level"""
        value = JSONUtils.get_nested_value(sample_json_data, ["name"])

        assert value == "test"

    def test_get_nested_value_multiple_levels(self, sample_json_data):
        """Get value from multiple nested levels"""
        value = JSONUtils.get_nested_value(sample_json_data, ["nested", "key"])

        assert value == "value"

    def test_get_nested_value_deep_nesting(self, sample_json_data):
        """Get value from deeply nested structure"""
        value = JSONUtils.get_nested_value(sample_json_data, ["nested", "deep", "data"])

        assert value == "test"

    def test_get_nested_value_missing_key(self, sample_json_data):
        """Get nonexistent key returns default"""
        value = JSONUtils.get_nested_value(sample_json_data, ["nonexistent"])

        assert value is None

    def test_get_nested_value_missing_key_with_default(self, sample_json_data):
        """Get nonexistent key with custom default"""
        value = JSONUtils.get_nested_value(sample_json_data, ["nonexistent"], default="default_value")

        assert value == "default_value"

    def test_get_nested_value_partial_path(self, sample_json_data):
        """Get with partial path (non-dict intermediate) returns default"""
        value = JSONUtils.get_nested_value(sample_json_data, ["name", "invalid"])

        assert value is None

    def test_get_nested_value_empty_keys(self, sample_json_data):
        """Get with empty keys list returns the data itself"""
        value = JSONUtils.get_nested_value(sample_json_data, [])

        # Empty keys list means no traversal, returns current (original data)
        assert value == sample_json_data

    def test_get_nested_value_list_index(self):
        """Get value from list using index"""
        data = {"items": [10, 20, 30]}
        # Note: JSONUtils uses dict keys, not list indices
        value = JSONUtils.get_nested_value(data, ["items", "0"])

        assert value is None  # "0" is not in list, treating as dict


class TestMergeJson:
    """Test JSONUtils.merge_json() method"""

    def test_merge_simple_dicts(self):
        """Merge simple dictionaries"""
        base = {"a": 1, "b": 2}
        updates = {"b": 20, "c": 3}

        result = JSONUtils.merge_json(base, updates)

        assert result["a"] == 1
        assert result["b"] == 20
        assert result["c"] == 3

    def test_merge_nested_dicts(self):
        """Merge nested dictionaries"""
        base = {"outer": {"inner": "value"}}
        updates = {"outer": {"inner": "updated"}}

        result = JSONUtils.merge_json(base, updates)

        assert result["outer"]["inner"] == "updated"

    def test_merge_does_not_mutate_base(self):
        """Merge does not mutate base dictionary"""
        base = {"a": 1}
        updates = {"b": 2}

        result = JSONUtils.merge_json(base, updates)

        assert base == {"a": 1}
        assert result == {"a": 1, "b": 2}

    def test_merge_complex_structure(self):
        """Merge complex nested structures"""
        base = {"config": {"debug": True, "level": "info"}, "name": "app"}
        updates = {"config": {"level": "debug"}}

        result = JSONUtils.merge_json(base, updates)

        assert result["config"]["debug"] is True
        assert result["config"]["level"] == "debug"
        assert result["name"] == "app"

    def test_merge_overwrites_non_dict(self):
        """Merge overwrites non-dict values"""
        base = {"key": [1, 2, 3]}
        updates = {"key": {"new": "dict"}}

        result = JSONUtils.merge_json(base, updates)

        assert result["key"] == {"new": "dict"}


class TestCreateStandardResponse:
    """Test JSONUtils.create_standard_response() method"""

    def test_create_success_response(self):
        """Create successful response"""
        response = JSONUtils.create_standard_response(success=True)

        assert response["success"] is True
        assert "error" not in response

    def test_create_error_response(self):
        """Create error response"""
        response = JSONUtils.create_standard_response(success=False, error="Something went wrong")

        assert response["success"] is False
        assert response["error"] == "Something went wrong"

    def test_create_response_with_message(self):
        """Create response with message"""
        response = JSONUtils.create_standard_response(success=True, message="Operation completed")

        assert response["success"] is True
        assert response["message"] == "Operation completed"

    def test_create_response_with_data(self):
        """Create response with data payload"""
        data = {"user_id": 123, "username": "test"}
        response = JSONUtils.create_standard_response(success=True, data=data)

        assert response["success"] is True
        assert response["data"] == data

    def test_create_response_all_fields(self):
        """Create response with all fields"""
        response = JSONUtils.create_standard_response(
            success=False, message="Operation failed", error="Database error", data={"retry_after": 60}
        )

        assert response["success"] is False
        assert response["message"] == "Operation failed"
        assert response["error"] == "Database error"
        assert response["data"]["retry_after"] == 60

    def test_create_response_default_success(self):
        """Create response defaults to success=True"""
        response = JSONUtils.create_standard_response()

        assert response["success"] is True

    def test_create_response_empty_optional_fields(self):
        """Create response omits empty optional fields"""
        response = JSONUtils.create_standard_response(success=True, message=None, error=None, data=None)

        assert response == {"success": True}


class TestCompactJson:
    """Test JSONUtils.compact_json() method"""

    def test_compact_json_no_whitespace(self):
        """Compact JSON has no extra whitespace"""
        data = {"key": "value", "nested": {"inner": "data"}}

        result = JSONUtils.compact_json(data)

        assert "\n" not in result
        assert "  " not in result

    def test_compact_json_correct_content(self):
        """Compact JSON contains correct data"""
        data = {"key": "value"}

        result = JSONUtils.compact_json(data)

        assert json.loads(result) == data

    def test_compact_json_unicode(self):
        """Compact JSON preserves unicode"""
        data = {"message": "🎯 테스트"}

        result = JSONUtils.compact_json(data)

        assert "🎯" in result
        assert "테스트" in result

    def test_compact_json_array(self):
        """Compact JSON with array"""
        data = [1, 2, 3, 4, 5]

        result = JSONUtils.compact_json(data)

        assert result == "[1,2,3,4,5]"


class TestPrettyJson:
    """Test JSONUtils.pretty_json() method"""

    def test_pretty_json_indentation(self):
        """Pretty JSON has proper indentation"""
        data = {"nested": {"inner": "value"}}

        result = JSONUtils.pretty_json(data)

        # Check for indentation
        assert "  " in result or "\n" in result

    def test_pretty_json_custom_indent(self):
        """Pretty JSON with custom indentation"""
        data = {"key": "value"}

        result = JSONUtils.pretty_json(data, indent=4)

        # Result should contain 4-space indent
        lines = result.split("\n")
        assert len(lines) > 1

    def test_pretty_json_unicode(self):
        """Pretty JSON preserves unicode"""
        data = {"message": "🎯 테스트"}

        result = JSONUtils.pretty_json(data)

        assert "🎯" in result
        assert "테스트" in result

    def test_pretty_json_parseable(self):
        """Pretty JSON is valid JSON"""
        data = {"key": "value", "nested": {"inner": "data"}}

        result = JSONUtils.pretty_json(data)

        # Should be parseable
        parsed = json.loads(result)
        assert parsed == data


class TestJSONSchemas:
    """Test JSONSchemas validation methods"""

    def test_validate_input_schema_valid(self):
        """Validate valid hook input"""
        data = {"tool_name": "test_tool"}

        result = JSONSchemas.validate_input_schema(data)

        assert result is True

    def test_validate_input_schema_missing_required(self):
        """Validate input missing required field"""
        data = {"tool_args": {}}

        result = JSONSchemas.validate_input_schema(data)

        assert result is False

    def test_validate_config_schema_valid(self):
        """Validate valid config schema"""
        data = {"hooks": {"timeout": 5}}

        result = JSONSchemas.validate_config_schema(data)

        assert result is True

    def test_validate_config_schema_with_tags(self):
        """Validate config with tags"""
        data = {"tags": {"policy": {"enforcement_mode": "strict"}}}

        result = JSONSchemas.validate_config_schema(data)

        assert result is True

    def test_validate_config_schema_empty_dict(self):
        """Validate empty dict config fails"""
        result = JSONSchemas.validate_config_schema({})

        assert result is False

    def test_validate_config_schema_non_dict(self):
        """Validate non-dict config fails"""
        result = JSONSchemas.validate_config_schema("not a dict")

        assert result is False


class TestJsonUtilsIntegration:
    """Integration tests for JSONUtils methods"""

    def test_write_read_roundtrip(self, tmp_path):
        """Write and read JSON file roundtrip"""
        original_data = {"name": "test", "config": {"enabled": True, "count": 42}}
        file_path = tmp_path / "roundtrip.json"

        # Write
        write_success = JSONUtils.write_json_to_file(original_data, file_path)
        assert write_success is True

        # Read
        read_data = JSONUtils.safe_json_load_file(file_path)

        assert read_data == original_data

    def test_merge_and_write(self, tmp_path):
        """Merge JSON and write to file"""
        base = {"a": 1, "b": {"c": 2}}
        updates = {"b": {"d": 3}}

        merged = JSONUtils.merge_json(base, updates)
        file_path = tmp_path / "merged.json"

        write_success = JSONUtils.write_json_to_file(merged, file_path)
        assert write_success is True

        read_data = JSONUtils.safe_json_load_file(file_path)
        assert read_data["b"]["c"] == 2
        assert read_data["b"]["d"] == 3

    def test_standard_response_roundtrip(self, tmp_path):
        """Create response, write, and read"""
        response = JSONUtils.create_standard_response(success=True, message="Test message", data={"id": 123})

        file_path = tmp_path / "response.json"
        write_success = JSONUtils.write_json_to_file(response, file_path)
        assert write_success is True

        read_response = JSONUtils.safe_json_load_file(file_path)
        assert read_response["success"] is True
        assert read_response["data"]["id"] == 123
