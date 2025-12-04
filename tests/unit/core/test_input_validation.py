"""
Minimal import and instantiation tests for Input Validation Middleware.

These tests verify that the module can be imported and basic classes
can be instantiated without errors.
"""

import pytest
from unittest.mock import MagicMock

from moai_adk.core.input_validation_middleware import (
    ValidationSeverity,
    ToolCategory,
    ValidationError,
    ValidationResult,
    ToolParameter,
    EnhancedInputValidationMiddleware,
)


class TestImports:
    """Test that all enums and classes can be imported."""

    def test_validation_severity_enum_exists(self):
        """Test ValidationSeverity enum is importable."""
        assert ValidationSeverity is not None
        assert hasattr(ValidationSeverity, "CRITICAL")

    def test_tool_category_enum_exists(self):
        """Test ToolCategory enum is importable."""
        assert ToolCategory is not None
        assert hasattr(ToolCategory, "SEARCH")

    def test_validation_error_class_exists(self):
        """Test ValidationError class is importable."""
        assert ValidationError is not None

    def test_validation_result_class_exists(self):
        """Test ValidationResult class is importable."""
        assert ValidationResult is not None

    def test_tool_parameter_class_exists(self):
        """Test ToolParameter class is importable."""
        assert ToolParameter is not None

    def test_enhanced_input_validation_middleware_exists(self):
        """Test EnhancedInputValidationMiddleware class is importable."""
        assert EnhancedInputValidationMiddleware is not None


class TestValidationSeverityEnum:
    """Test ValidationSeverity enum values."""

    def test_validation_severity_low(self):
        """Test ValidationSeverity has LOW."""
        assert hasattr(ValidationSeverity, "LOW")
        assert ValidationSeverity.LOW.value == "low"

    def test_validation_severity_medium(self):
        """Test ValidationSeverity has MEDIUM."""
        assert hasattr(ValidationSeverity, "MEDIUM")
        assert ValidationSeverity.MEDIUM.value == "medium"

    def test_validation_severity_high(self):
        """Test ValidationSeverity has HIGH."""
        assert hasattr(ValidationSeverity, "HIGH")
        assert ValidationSeverity.HIGH.value == "high"

    def test_validation_severity_critical(self):
        """Test ValidationSeverity has CRITICAL."""
        assert hasattr(ValidationSeverity, "CRITICAL")
        assert ValidationSeverity.CRITICAL.value == "critical"


class TestToolCategoryEnum:
    """Test ToolCategory enum values."""

    def test_tool_category_search(self):
        """Test ToolCategory has SEARCH."""
        assert hasattr(ToolCategory, "SEARCH")
        assert ToolCategory.SEARCH.value == "search"

    def test_tool_category_file_operations(self):
        """Test ToolCategory has FILE_OPERATIONS."""
        assert hasattr(ToolCategory, "FILE_OPERATIONS")
        assert ToolCategory.FILE_OPERATIONS.value == "file_operations"

    def test_tool_category_text_processing(self):
        """Test ToolCategory has TEXT_PROCESSING."""
        assert hasattr(ToolCategory, "TEXT_PROCESSING")

    def test_tool_category_data_analysis(self):
        """Test ToolCategory has DATA_ANALYSIS."""
        assert hasattr(ToolCategory, "DATA_ANALYSIS")

    def test_tool_category_system(self):
        """Test ToolCategory has SYSTEM."""
        assert hasattr(ToolCategory, "SYSTEM")
        assert ToolCategory.SYSTEM.value == "system"

    def test_tool_category_general(self):
        """Test ToolCategory has GENERAL."""
        assert hasattr(ToolCategory, "GENERAL")
        assert ToolCategory.GENERAL.value == "general"


class TestValidationErrorInstantiation:
    """Test ValidationError dataclass instantiation."""

    def test_validation_error_basic_init(self):
        """Test ValidationError can be instantiated."""
        error = ValidationError(
            code="VAL001",
            message="Invalid parameter",
            path=["param1"],
            severity=ValidationSeverity.HIGH,
        )
        assert error.code == "VAL001"
        assert error.message == "Invalid parameter"
        assert error.severity == ValidationSeverity.HIGH

    def test_validation_error_defaults(self):
        """Test ValidationError respects default values."""
        error = ValidationError(
            code="VAL001",
            message="Test",
            path=["test"],
            severity=ValidationSeverity.MEDIUM,
        )
        assert error.auto_corrected is False
        assert error.original_value is None
        assert error.corrected_value is None
        assert error.suggestion is None

    def test_validation_error_with_suggestions(self):
        """Test ValidationError with correction suggestion."""
        error = ValidationError(
            code="VAL002",
            message="Type mismatch",
            path=["field"],
            severity=ValidationSeverity.HIGH,
            auto_corrected=True,
            original_value="123",
            corrected_value=123,
            suggestion="Convert string to integer",
        )
        assert error.auto_corrected is True
        assert error.original_value == "123"
        assert error.corrected_value == 123
        assert error.suggestion == "Convert string to integer"


class TestValidationResultInstantiation:
    """Test ValidationResult dataclass instantiation."""

    def test_validation_result_valid_init(self):
        """Test ValidationResult for valid input."""
        result = ValidationResult(
            valid=True,
            normalized_input={"param1": "value1"},
        )
        assert result.valid is True
        assert result.normalized_input == {"param1": "value1"}

    def test_validation_result_invalid_init(self):
        """Test ValidationResult for invalid input."""
        errors = [
            ValidationError(
                code="VAL001",
                message="Invalid value",
                path=["param"],
                severity=ValidationSeverity.CRITICAL,
            )
        ]
        result = ValidationResult(
            valid=False,
            normalized_input={},
            errors=errors,
        )
        assert result.valid is False
        assert len(result.errors) == 1
        assert result.errors[0].code == "VAL001"

    def test_validation_result_defaults(self):
        """Test ValidationResult respects default values."""
        result = ValidationResult(
            valid=True,
            normalized_input={},
        )
        assert isinstance(result.errors, list)
        assert len(result.errors) == 0
        assert isinstance(result.warnings, list)
        assert isinstance(result.transformations, list)
        assert result.processing_time_ms == 0.0

    def test_validation_result_with_warnings(self):
        """Test ValidationResult with warnings."""
        result = ValidationResult(
            valid=True,
            normalized_input={"param": "value"},
            warnings=["Deprecated parameter used"],
        )
        assert len(result.warnings) == 1
        assert "Deprecated" in result.warnings[0]

    def test_validation_result_with_transformations(self):
        """Test ValidationResult with transformations."""
        result = ValidationResult(
            valid=True,
            normalized_input={"param": 123},
            transformations=["Converted string to integer", "Removed whitespace"],
        )
        assert len(result.transformations) == 2


class TestToolParameterInstantiation:
    """Test ToolParameter dataclass instantiation."""

    def test_tool_parameter_required_field(self):
        """Test ToolParameter with required field."""
        param = ToolParameter(
            name="filename",
            param_type="string",
            required=True,
        )
        assert param.name == "filename"
        assert param.param_type == "string"
        assert param.required is True

    def test_tool_parameter_optional_field(self):
        """Test ToolParameter with optional field."""
        param = ToolParameter(
            name="timeout",
            param_type="integer",
            required=False,
            default_value=30,
        )
        assert param.required is False
        assert param.default_value == 30

    def test_tool_parameter_with_aliases(self):
        """Test ToolParameter with aliases."""
        param = ToolParameter(
            name="file_path",
            param_type="string",
            aliases=["path", "file", "fp"],
            deprecated_aliases=["old_path", "file_name"],
        )
        assert "path" in param.aliases
        assert "old_path" in param.deprecated_aliases

    def test_tool_parameter_with_validation(self):
        """Test ToolParameter with validation function."""
        validation_func = MagicMock()
        param = ToolParameter(
            name="port",
            param_type="integer",
            validation_function=validation_func,
            description="Port number",
        )
        assert param.validation_function is validation_func
        assert param.description == "Port number"

    def test_tool_parameter_defaults(self):
        """Test ToolParameter respects default values."""
        param = ToolParameter(
            name="test_param",
            param_type="string",
        )
        assert param.required is False
        assert param.default_value is None
        assert isinstance(param.aliases, list)
        assert len(param.aliases) == 0
        assert param.validation_function is None
        assert param.description == ""


class TestEnhancedInputValidationMiddlewareInstantiation:
    """Test EnhancedInputValidationMiddleware instantiation."""

    def test_middleware_init_default(self):
        """Test EnhancedInputValidationMiddleware with default parameters."""
        middleware = EnhancedInputValidationMiddleware()
        assert middleware is not None
        assert middleware.enable_logging is True
        assert middleware.enable_caching is True

    def test_middleware_init_custom(self):
        """Test EnhancedInputValidationMiddleware with custom parameters."""
        middleware = EnhancedInputValidationMiddleware(
            enable_logging=False,
            enable_caching=False,
        )
        assert middleware.enable_logging is False
        assert middleware.enable_caching is False

    def test_middleware_has_tool_parameters(self):
        """Test EnhancedInputValidationMiddleware has tool_parameters."""
        middleware = EnhancedInputValidationMiddleware()
        assert hasattr(middleware, "tool_parameters")

    def test_middleware_has_validation_methods(self):
        """Test EnhancedInputValidationMiddleware has validation methods."""
        middleware = EnhancedInputValidationMiddleware()
        # Check for methods by looking at callable attributes
        methods = [m for m in dir(middleware) if not m.startswith("_") and callable(getattr(middleware, m))]
        assert len(methods) > 0


class TestEnumValues:
    """Test enum value types and formats."""

    def test_validation_severity_values_are_strings(self):
        """Test all ValidationSeverity values are strings."""
        for severity in ValidationSeverity:
            assert isinstance(severity.value, str)

    def test_tool_category_values_are_strings(self):
        """Test all ToolCategory values are strings."""
        for category in ToolCategory:
            assert isinstance(category.value, str)
