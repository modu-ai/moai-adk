#!/usr/bin/env python3
"""JSON Utilities for Alfred Hooks

Provides consistent JSON handling, validation, and serialization across all hooks.
"""

import json
import sys
from pathlib import Path
from typing import Any, Dict, List, Optional, Union


class JSONUtils:
    """Utility class for consistent JSON handling in hooks."""

    @staticmethod
    def read_json_from_stdin() -> Dict[str, Any]:
        """Read and parse JSON from stdin.

        Returns:
            Parsed JSON data as dictionary

        Raises:
            JSONDecodeError: If JSON parsing fails
        """
        # Handle Docker/non-interactive environments by checking TTY
        input_data = sys.stdin.read() if not sys.stdin.isatty() else "{}"
        if input_data.strip():
            return json.loads(input_data)
        return {}

    @staticmethod
    def safe_json_loads(json_str: str, default: Optional[Any] = None) -> Union[Dict[str, Any], Any]:
        """Safely parse JSON string with fallback.

        Args:
            json_str: JSON string to parse
            default: Default value if parsing fails

        Returns:
            Parsed JSON data or default value
        """
        try:
            if json_str.strip():
                return json.loads(json_str)
        except json.JSONDecodeError:
            pass

        return default if default is not None else {}

    @staticmethod
    def safe_json_load_file(file_path: Path, default: Optional[Any] = None) -> Union[Dict[str, Any], Any]:
        """Safely load JSON from file with fallback.

        Args:
            file_path: Path to JSON file
            default: Default value if file doesn't exist or parsing fails

        Returns:
            Parsed JSON data or default value
        """
        try:
            if file_path.exists():
                with open(file_path, 'r', encoding='utf-8') as f:
                    return json.load(f)
        except (json.JSONDecodeError, IOError, OSError):
            pass

        return default if default is not None else {}

    @staticmethod
    def write_json_to_file(data: Dict[str, Any], file_path: Path, indent: int = 2) -> bool:
        """Write JSON data to file with error handling.

        Args:
            data: JSON data to write
            file_path: Path to write to
            indent: JSON indentation level

        Returns:
            True if successful, False otherwise
        """
        try:
            # Ensure parent directory exists
            file_path.parent.mkdir(parents=True, exist_ok=True)

            with open(file_path, 'w', encoding='utf-8') as f:
                json.dump(data, f, indent=indent, ensure_ascii=False)
            return True
        except (IOError, OSError, TypeError):
            return False

    @staticmethod
    def validate_json_schema(data: Dict[str, Any], required_fields: List[str]) -> bool:
        """Validate JSON data has required fields.

        Args:
            data: JSON data to validate
            required_fields: List of required field names

        Returns:
            True if all required fields are present, False otherwise
        """
        if not isinstance(data, dict):
            return False

        return all(field in data for field in required_fields)

    @staticmethod
    def get_nested_value(data: Dict[str, Any], keys: List[str], default: Optional[Any] = None) -> Any:
        """Get nested value from dictionary using dot notation.

        Args:
            data: Dictionary to search in
            keys: List of keys (path)
            default: Default value if key not found

        Returns:
            Nested value or default
        """
        current = data
        for key in keys:
            if isinstance(current, dict) and key in current:
                current = current[key]
            else:
                return default
        return current

    @staticmethod
    def merge_json(base: Dict[str, Any], updates: Dict[str, Any]) -> Dict[str, Any]:
        """Merge two JSON dictionaries recursively.

        Args:
            base: Base dictionary
            updates: Dictionary with updates

        Returns:
            Merged dictionary
        """
        result = base.copy()

        for key, value in updates.items():
            if key in result and isinstance(result[key], dict) and isinstance(value, dict):
                result[key] = JSONUtils.merge_json(result[key], value)
            else:
                result[key] = value

        return result

    @staticmethod
    def create_standard_response(
        success: bool = True,
        message: Optional[str] = None,
        error: Optional[str] = None,
        data: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """Create standardized JSON response.

        Args:
            success: Whether operation was successful
            message: Descriptive message
            error: Error message if failed
            data: Additional data payload

        Returns:
            Standardized response dictionary
        """
        response: dict[str, Any] = {"success": success}

        if message:
            response["message"] = message
        if error:
            response["error"] = error
        if data:
            response["data"] = data

        return response

    @staticmethod
    def compact_json(data: Dict[str, Any]) -> str:
        """Generate compact JSON string without whitespace.

        Args:
            data: Dictionary to serialize

        Returns:
            Compact JSON string
        """
        return json.dumps(data, separators=(',', ':'), ensure_ascii=False)

    @staticmethod
    def pretty_json(data: Dict[str, Any], indent: int = 2) -> str:
        """Generate pretty JSON string with indentation.

        Args:
            data: Dictionary to serialize
            indent: Indentation level

        Returns:
            Pretty JSON string
        """
        return json.dumps(data, indent=indent, ensure_ascii=False)


# Common JSON validation schemas
class JSONSchemas:
    """Common JSON schemas for hook validation."""

    # Hook input schema
    HOOK_INPUT_SCHEMA = {
        "type": "object",
        "properties": {
            "tool_name": {"type": "string"},
            "tool_args": {"type": "object"},
            "tool_result": {"type": "object"}
        },
        "required": ["tool_name"]
    }

    # Hook configuration schema
    CONFIG_SCHEMA = {
        "type": "object",
        "properties": {
            "hooks": {
                "type": "object",
                "properties": {
                    "timeout": {"type": "number", "minimum": 1},
                    "enabled": {"type": "boolean"},
                    "graceful_degradation": {"type": "boolean"}
                }
            },
            "tags": {
                "type": "object",
                "properties": {
                    "policy": {
                        "type": "object",
                        "properties": {
                            "enforcement_mode": {"type": "string"},
                            "require_spec_before_code": {"type": "boolean"},
                            "require_test_for_code": {"type": "boolean"}
                        }
                    }
                }
            }
        }
    }

    @staticmethod
    def validate_input_schema(data: Dict[str, Any]) -> bool:
        """Validate hook input data against schema."""
        return JSONUtils.validate_json_schema(data, ["tool_name"])

    @staticmethod
    def validate_config_schema(data: Dict[str, Any]) -> bool:
        """Validate configuration data against schema."""
        # Basic validation - can be extended with more detailed validation
        return isinstance(data, dict) and ("hooks" in data or "tags" in data)
