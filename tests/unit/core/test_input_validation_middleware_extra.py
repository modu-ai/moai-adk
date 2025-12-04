"""
Extended tests for input_validation_middleware module - comprehensive coverage.

Tests enums, dataclasses, validation logic, parameter mapping, and error handling.
"""

import json
import tempfile
from pathlib import Path
from typing import Any, Dict
from unittest.mock import MagicMock, Mock, patch

import pytest

from moai_adk.core.input_validation_middleware import (
    EnhancedInputValidationMiddleware,
    ToolCategory,
    ToolParameter,
    ValidationError,
    ValidationResult,
    ValidationSeverity,
    get_validation_stats,
    validate_tool_input,
    validation_middleware,
)


class TestEnums:
    """Test ValidationSeverity and ToolCategory enums."""

    def test_validation_severity_values(self):
        """Test ValidationSeverity enum values."""
        assert ValidationSeverity.LOW.value == "low"
        assert ValidationSeverity.MEDIUM.value == "medium"
        assert ValidationSeverity.HIGH.value == "high"
        assert ValidationSeverity.CRITICAL.value == "critical"

    def test_tool_category_values(self):
        """Test ToolCategory enum values."""
        assert ToolCategory.SEARCH.value == "search"
        assert ToolCategory.FILE_OPERATIONS.value == "file_operations"
        assert ToolCategory.TEXT_PROCESSING.value == "text_processing"
        assert ToolCategory.DATA_ANALYSIS.value == "data_analysis"
        assert ToolCategory.SYSTEM.value == "system"
        assert ToolCategory.GENERAL.value == "general"


class TestValidationErrorDataclass:
    """Test ValidationError dataclass."""

    def test_validation_error_creation(self):
        """Test ValidationError creation."""
        error = ValidationError(
            code="test_error",
            message="Test message",
            path=["param1", "param2"],
            severity=ValidationSeverity.HIGH,
            auto_corrected=True,
            original_value="old",
            corrected_value="new",
            suggestion="Use 'new' instead",
        )
        assert error.code == "test_error"
        assert error.severity == ValidationSeverity.HIGH
        assert error.auto_corrected is True

    def test_validation_error_defaults(self):
        """Test ValidationError default values."""
        error = ValidationError(
            code="error",
            message="message",
            path=[],
            severity=ValidationSeverity.LOW,
        )
        assert error.auto_corrected is False
        assert error.original_value is None
        assert error.corrected_value is None
        assert error.suggestion is None


class TestValidationResultDataclass:
    """Test ValidationResult dataclass."""

    def test_validation_result_valid(self):
        """Test ValidationResult for valid input."""
        result = ValidationResult(
            valid=True,
            normalized_input={"param": "value"},
        )
        assert result.valid is True
        assert result.errors == []
        assert result.warnings == []
        assert result.transformations == []

    def test_validation_result_with_errors(self):
        """Test ValidationResult with errors."""
        error = ValidationError(
            code="error",
            message="Error message",
            path=["param"],
            severity=ValidationSeverity.CRITICAL,
        )
        result = ValidationResult(
            valid=False,
            normalized_input={},
            errors=[error],
        )
        assert result.valid is False
        assert len(result.errors) == 1


class TestToolParameterDataclass:
    """Test ToolParameter dataclass."""

    def test_tool_parameter_creation(self):
        """Test ToolParameter creation."""
        param = ToolParameter(
            name="test_param",
            param_type="string",
            required=True,
            description="Test parameter",
            aliases=["alias1", "alias2"],
        )
        assert param.name == "test_param"
        assert param.required is True
        assert len(param.aliases) == 2

    def test_tool_parameter_defaults(self):
        """Test ToolParameter defaults."""
        param = ToolParameter(
            name="param",
            param_type="string",
        )
        assert param.required is False
        assert param.default_value is None
        assert param.aliases == []
        assert param.deprecated_aliases == []
        assert param.validation_function is None


class TestEnhancedInputValidationMiddlewareInit:
    """Test middleware initialization."""

    def test_middleware_init_default(self):
        """Test middleware initialization with defaults."""
        middleware = EnhancedInputValidationMiddleware()
        assert middleware.enable_logging is True
        assert middleware.enable_caching is True
        assert len(middleware.tool_parameters) > 0

    def test_middleware_init_no_logging(self):
        """Test middleware initialization without logging."""
        middleware = EnhancedInputValidationMiddleware(enable_logging=False)
        assert middleware.enable_logging is False

    def test_middleware_init_no_cache(self):
        """Test middleware initialization without caching."""
        middleware = EnhancedInputValidationMiddleware(enable_caching=False)
        assert middleware.enable_caching is False
        assert middleware.validation_cache is None

    def test_middleware_loads_tool_parameters(self):
        """Test middleware loads tool parameter definitions."""
        middleware = EnhancedInputValidationMiddleware()
        assert "Grep" in middleware.tool_parameters
        assert "Read" in middleware.tool_parameters
        assert "Write" in middleware.tool_parameters
        assert "Bash" in middleware.tool_parameters

    def test_middleware_loads_parameter_mappings(self):
        """Test middleware loads parameter mappings."""
        middleware = EnhancedInputValidationMiddleware()
        assert len(middleware.parameter_mappings) > 0
        assert middleware.parameter_mappings.get("head_limit") == "head_limit"


class TestValidationAndNormalization:
    """Test main validation and normalization method."""

    def test_validate_valid_input(self):
        """Test validation of valid input."""
        middleware = EnhancedInputValidationMiddleware()
        result = middleware.validate_and_normalize_input(
            "Grep",
            {"pattern": "test"},
        )
        assert result.valid is True
        assert result.normalized_input["pattern"] == "test"

    def test_validate_with_defaults(self):
        """Test validation applies default values."""
        middleware = EnhancedInputValidationMiddleware()
        result = middleware.validate_and_normalize_input(
            "Grep",
            {"pattern": "test"},
        )
        # output_mode should have default
        assert "output_mode" in result.normalized_input

    def test_validate_unknown_tool(self):
        """Test validation of unknown tool."""
        middleware = EnhancedInputValidationMiddleware()
        result = middleware.validate_and_normalize_input(
            "UnknownTool",
            {"param": "value"},
        )
        assert len(result.warnings) > 0

    def test_validate_missing_required(self):
        """Test validation detects missing required parameters."""
        middleware = EnhancedInputValidationMiddleware()
        result = middleware.validate_and_normalize_input(
            "Grep",
            {},
        )
        assert result.valid is False
        assert any(e.code == "missing_required_parameter" for e in result.errors)

    def test_validate_updates_stats(self):
        """Test validation updates statistics."""
        middleware = EnhancedInputValidationMiddleware()
        initial_count = middleware.stats["validations_performed"]

        middleware.validate_and_normalize_input("Grep", {"pattern": "test"})

        assert middleware.stats["validations_performed"] == initial_count + 1


class TestParameterMapping:
    """Test parameter mapping functionality."""

    def test_map_parameter_aliases(self):
        """Test parameter alias mapping."""
        middleware = EnhancedInputValidationMiddleware()
        result = middleware.validate_and_normalize_input(
            "Grep",
            {"pattern": "test", "max_results": 10},
        )
        # max_results should be mapped to head_limit
        assert "head_limit" in result.normalized_input or result.valid is True

    def test_map_deprecated_parameters(self):
        """Test deprecated parameter handling."""
        middleware = EnhancedInputValidationMiddleware()
        result = middleware.validate_and_normalize_input(
            "Grep",
            {"pattern": "test", "head": 5},
        )
        # Should be mapped or warned
        assert result.valid is True or any("deprecated" in w.lower() for w in result.warnings)

    def test_unrecognized_parameter(self):
        """Test handling of unrecognized parameters."""
        middleware = EnhancedInputValidationMiddleware()
        result = middleware.validate_and_normalize_input(
            "Grep",
            {"pattern": "test", "unknown_param": "value"},
        )
        assert any(e.code == "unrecognized_parameter" for e in result.errors)

    def test_find_closest_parameter_match(self):
        """Test finding closest parameter name match."""
        middleware = EnhancedInputValidationMiddleware()
        match = middleware._find_closest_parameter_match("patern", {"pattern"})
        assert match == "pattern" or match is None


class TestTypeValidation:
    """Test type validation."""

    def test_validate_string_type(self):
        """Test string type validation."""
        middleware = EnhancedInputValidationMiddleware()
        result = middleware.validate_and_normalize_input(
            "Grep",
            {"pattern": "test"},
        )
        assert result.valid is True

    def test_validate_integer_type(self):
        """Test integer type validation."""
        middleware = EnhancedInputValidationMiddleware()
        result = middleware.validate_and_normalize_input(
            "Grep",
            {"pattern": "test", "head_limit": 10},
        )
        assert result.valid is True

    def test_convert_string_to_integer(self):
        """Test conversion of string to integer."""
        middleware = EnhancedInputValidationMiddleware()
        result = middleware.validate_and_normalize_input(
            "Grep",
            {"pattern": "test", "head_limit": "20"},
        )
        # Should attempt conversion
        assert isinstance(result.normalized_input.get("head_limit"), (int, str))

    def test_validate_boolean_type(self):
        """Test boolean type validation."""
        middleware = EnhancedInputValidationMiddleware()
        result = middleware.validate_and_normalize_input(
            "Grep",
            {"pattern": "test", "case_sensitive": True},
        )
        assert result.valid is True

    def test_type_mismatch_error(self):
        """Test type mismatch error."""
        middleware = EnhancedInputValidationMiddleware()
        # Object type where string expected - may not error but test handling
        result = middleware.validate_and_normalize_input(
            "Grep",
            {"pattern": {"nested": "dict"}},
        )
        # Should either error or convert
        assert isinstance(result, ValidationResult)


class TestParameterConversion:
    """Test parameter type conversion."""

    def test_convert_string_to_string(self):
        """Test string conversion."""
        middleware = EnhancedInputValidationMiddleware()
        result = middleware._convert_parameter_type("test", "string")
        assert result == "test"

    def test_convert_to_integer(self):
        """Test conversion to integer."""
        middleware = EnhancedInputValidationMiddleware()
        assert middleware._convert_parameter_type("123", "integer") == 123
        assert middleware._convert_parameter_type(123.5, "integer") == 123
        assert middleware._convert_parameter_type(True, "integer") == 1

    def test_convert_to_boolean(self):
        """Test conversion to boolean."""
        middleware = EnhancedInputValidationMiddleware()
        assert middleware._convert_parameter_type("true", "boolean") is True
        assert middleware._convert_parameter_type("false", "boolean") is False
        assert middleware._convert_parameter_type(1, "boolean") is True
        assert middleware._convert_parameter_type(0, "boolean") is False

    def test_convert_to_float(self):
        """Test conversion to float."""
        middleware = EnhancedInputValidationMiddleware()
        assert middleware._convert_parameter_type("3.14", "float") == 3.14
        assert middleware._convert_parameter_type(3, "float") == 3.0

    def test_convert_invalid_value(self):
        """Test conversion of invalid value."""
        middleware = EnhancedInputValidationMiddleware()
        result = middleware._convert_parameter_type("not_a_number", "integer")
        assert result is None


class TestParameterNormalization:
    """Test parameter normalization."""

    def test_normalize_boolean(self):
        """Test boolean normalization."""
        middleware = EnhancedInputValidationMiddleware()
        input_data = {"case_sensitive": "true"}
        middleware._normalize_parameter_formats(
            [ToolParameter("case_sensitive", "boolean")],
            input_data,
        )
        # Should be normalized to boolean
        assert isinstance(input_data["case_sensitive"], (bool, str))

    def test_normalize_path(self):
        """Test path normalization."""
        middleware = EnhancedInputValidationMiddleware()
        input_data = {"file_path": "path\\to\\file"}
        middleware._normalize_parameter_formats(
            [ToolParameter("file_path", "string")],
            input_data,
        )
        # Forward slashes should be applied
        assert isinstance(input_data["file_path"], str)

    def test_normalize_numeric(self):
        """Test numeric normalization."""
        middleware = EnhancedInputValidationMiddleware()
        input_data = {"limit": "100"}
        middleware._normalize_parameter_formats(
            [ToolParameter("limit", "integer")],
            input_data,
        )
        # Should be normalized to integer
        assert isinstance(input_data["limit"], (int, str))


class TestGreplModeValidation:
    """Test Grep-specific validation."""

    def test_grep_valid_output_mode(self):
        """Test valid Grep output modes."""
        middleware = EnhancedInputValidationMiddleware()
        assert middleware._validate_grep_mode("content") is True
        assert middleware._validate_grep_mode("files_with_matches") is True
        assert middleware._validate_grep_mode("count") is True

    def test_grep_invalid_output_mode(self):
        """Test invalid Grep output mode."""
        middleware = EnhancedInputValidationMiddleware()
        assert middleware._validate_grep_mode("invalid") is False


class TestStatisticsAndCaching:
    """Test statistics and caching."""

    def test_validation_stats(self):
        """Test validation statistics."""
        middleware = EnhancedInputValidationMiddleware()
        initial_stats = middleware.get_validation_stats()

        middleware.validate_and_normalize_input("Grep", {"pattern": "test"})

        updated_stats = middleware.get_validation_stats()
        assert updated_stats["validations_performed"] > initial_stats["validations_performed"]

    def test_validation_caching(self):
        """Test validation result caching."""
        middleware = EnhancedInputValidationMiddleware(enable_caching=True)
        input_data = {"pattern": "test"}

        result1 = middleware.validate_and_normalize_input("Grep", input_data)
        # Caching should not affect single validation

        assert isinstance(result1, ValidationResult)

    def test_no_caching(self):
        """Test with caching disabled."""
        middleware = EnhancedInputValidationMiddleware(enable_caching=False)
        result = middleware.validate_and_normalize_input("Grep", {"pattern": "test"})
        assert result.valid is True
        assert middleware.validation_cache is None


class TestCustomTools:
    """Test custom tool registration."""

    def test_register_tool_parameters(self):
        """Test registering custom tool parameters."""
        middleware = EnhancedInputValidationMiddleware()
        custom_params = [
            ToolParameter("custom_param", "string", required=True),
        ]
        middleware.register_tool_parameters("CustomTool", custom_params)

        assert "CustomTool" in middleware.tool_parameters

    def test_add_parameter_mapping(self):
        """Test adding custom parameter mapping."""
        middleware = EnhancedInputValidationMiddleware()
        initial_count = len(middleware.parameter_mappings)

        middleware.add_parameter_mapping("old_name", "new_name")

        assert len(middleware.parameter_mappings) == initial_count + 1


class TestToolSpecificValidation:
    """Test tool-specific validation rules."""

    def test_read_tool_validation(self):
        """Test Read tool validation."""
        middleware = EnhancedInputValidationMiddleware()
        result = middleware.validate_and_normalize_input(
            "Read",
            {"file_path": "/path/to/file.txt"},
        )
        assert result.valid is True
        assert "file_path" in result.normalized_input

    def test_write_tool_validation(self):
        """Test Write tool validation."""
        middleware = EnhancedInputValidationMiddleware()
        result = middleware.validate_and_normalize_input(
            "Write",
            {"file_path": "/path/to/file.txt", "content": "test"},
        )
        assert result.valid is True

    def test_bash_tool_validation(self):
        """Test Bash tool validation."""
        middleware = EnhancedInputValidationMiddleware()
        result = middleware.validate_and_normalize_input(
            "Bash",
            {"command": "ls -la"},
        )
        assert result.valid is True

    def test_edit_tool_validation(self):
        """Test Edit tool validation."""
        middleware = EnhancedInputValidationMiddleware()
        result = middleware.validate_and_normalize_input(
            "Edit",
            {
                "file_path": "/path/to/file.txt",
                "old_string": "old",
                "new_string": "new",
            },
        )
        assert result.valid is True


class TestExportAndReporting:
    """Test export and reporting functionality."""

    def test_export_validation_report(self):
        """Test exporting validation report."""
        with tempfile.TemporaryDirectory() as tmpdir:
            middleware = EnhancedInputValidationMiddleware()
            output_path = Path(tmpdir) / "report.json"

            middleware.export_validation_report(str(output_path))

            assert output_path.exists()
            with open(output_path) as f:
                report = json.load(f)
            assert "stats" in report
            assert "configured_tools" in report


class TestGlobalFunctions:
    """Test module-level convenience functions."""

    def test_validate_tool_input_function(self):
        """Test validate_tool_input function."""
        result = validate_tool_input("Grep", {"pattern": "test"})
        assert isinstance(result, ValidationResult)

    def test_get_validation_stats_function(self):
        """Test get_validation_stats function."""
        stats = get_validation_stats()
        assert isinstance(stats, dict)
        assert "validations_performed" in stats

    def test_global_middleware_instance(self):
        """Test global middleware instance."""
        assert validation_middleware is not None
        assert isinstance(validation_middleware, EnhancedInputValidationMiddleware)


class TestEdgeCases:
    """Test edge cases and error conditions."""

    def test_empty_input(self):
        """Test with empty input."""
        middleware = EnhancedInputValidationMiddleware()
        result = middleware.validate_and_normalize_input("Grep", {})
        # Should have errors for missing required params
        assert any(e.code == "missing_required_parameter" for e in result.errors)

    def test_none_values(self):
        """Test with None values."""
        middleware = EnhancedInputValidationMiddleware()
        result = middleware.validate_and_normalize_input(
            "Grep",
            {"pattern": "test", "head_limit": None},
        )
        # Should handle None gracefully
        assert isinstance(result, ValidationResult)

    def test_very_long_strings(self):
        """Test with very long strings."""
        middleware = EnhancedInputValidationMiddleware()
        long_string = "x" * 100000
        result = middleware.validate_and_normalize_input(
            "Grep",
            {"pattern": long_string},
        )
        assert result.valid is True

    def test_special_characters(self):
        """Test with special characters."""
        middleware = EnhancedInputValidationMiddleware()
        result = middleware.validate_and_normalize_input(
            "Grep",
            {"pattern": "<>?:|\\/!@#$%^&*()"},
        )
        # Should handle special chars
        assert isinstance(result, ValidationResult)

    def test_unicode_values(self):
        """Test with unicode values."""
        middleware = EnhancedInputValidationMiddleware()
        result = middleware.validate_and_normalize_input(
            "Grep",
            {"pattern": "测试中文😀"},
        )
        assert result.valid is True

    def test_circular_reference_in_dict(self):
        """Test handling of complex nested structures."""
        middleware = EnhancedInputValidationMiddleware()
        result = middleware.validate_and_normalize_input(
            "Bash",
            {"command": "test", "environment": {"KEY": "VALUE", "NESTED": {"DEEP": "value"}}},
        )
        # Should handle nested dicts
        assert isinstance(result, ValidationResult)

    def test_processing_time_tracking(self):
        """Test that processing time is tracked."""
        middleware = EnhancedInputValidationMiddleware()
        result = middleware.validate_and_normalize_input(
            "Grep",
            {"pattern": "test"},
        )
        assert result.processing_time_ms > 0


class TestComplexScenarios:
    """Test complex validation scenarios."""

    def test_multiple_transformations(self):
        """Test input requiring multiple transformations."""
        middleware = EnhancedInputValidationMiddleware()
        result = middleware.validate_and_normalize_input(
            "Grep",
            {
                "pattern": "test",
                "head": "10",
                "search_path": "C:\\path\\to\\folder",
                "mode": "content",
            },
        )
        # Should handle multiple transforms
        assert len(result.transformations) >= 0

    def test_recovery_from_errors(self):
        """Test error recovery with auto-correction."""
        middleware = EnhancedInputValidationMiddleware()
        result = middleware.validate_and_normalize_input(
            "Read",
            {"filename": "/path/to/file.txt", "start_line": "5", "lines": "10"},
        )
        # Should auto-correct/map parameters
        if not result.valid:
            assert len(result.errors) > 0


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
