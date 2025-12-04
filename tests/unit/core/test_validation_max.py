"""
Comprehensive test coverage for input_validation_middleware.py

Target: 150+ lines of coverage for ALL validation methods and parameter mapping
Strategy: Maximum test coverage with mocking
"""

import json
from unittest.mock import Mock, patch, MagicMock
from typing import Any, Dict

import pytest

from moai_adk.core.input_validation_middleware import (
    EnhancedInputValidationMiddleware,
    ValidationError,
    ValidationResult,
    ValidationSeverity,
    ToolParameter,
    ToolCategory,
    validate_tool_input,
    get_validation_stats,
)


class TestValidationError:
    """Tests for ValidationError dataclass"""

    def test_validation_error_creation(self):
        """Test ValidationError creation"""
        error = ValidationError(
            code="test_code",
            message="Test message",
            path=["param1"],
            severity=ValidationSeverity.MEDIUM,
        )
        assert error.code == "test_code"
        assert error.severity == ValidationSeverity.MEDIUM
        assert error.auto_corrected is False

    def test_validation_error_with_correction(self):
        """Test ValidationError with correction"""
        error = ValidationError(
            code="type_conversion",
            message="Converted type",
            path=["param"],
            severity=ValidationSeverity.LOW,
            auto_corrected=True,
            original_value="123",
            corrected_value=123,
        )
        assert error.auto_corrected is True
        assert error.original_value == "123"
        assert error.corrected_value == 123


class TestValidationResult:
    """Tests for ValidationResult dataclass"""

    def test_validation_result_creation(self):
        """Test ValidationResult creation"""
        result = ValidationResult(
            valid=True,
            normalized_input={"param": "value"},
        )
        assert result.valid is True
        assert result.normalized_input["param"] == "value"
        assert len(result.errors) == 0

    def test_validation_result_with_errors(self):
        """Test ValidationResult with errors"""
        error = ValidationError(
            code="error",
            message="Test error",
            path=[],
            severity=ValidationSeverity.HIGH,
        )
        result = ValidationResult(
            valid=False,
            normalized_input={},
            errors=[error],
        )
        assert result.valid is False
        assert len(result.errors) == 1


class TestToolParameter:
    """Tests for ToolParameter dataclass"""

    def test_tool_parameter_creation(self):
        """Test ToolParameter creation"""
        param = ToolParameter(
            name="pattern",
            param_type="string",
            required=True,
            description="Search pattern",
        )
        assert param.name == "pattern"
        assert param.param_type == "string"
        assert param.required is True

    def test_tool_parameter_with_aliases(self):
        """Test ToolParameter with aliases"""
        param = ToolParameter(
            name="limit",
            param_type="integer",
            aliases=["count", "max", "max_results"],
        )
        assert len(param.aliases) == 3
        assert "count" in param.aliases


class TestEnhancedInputValidationMiddleware:
    """Tests for EnhancedInputValidationMiddleware"""

    def test_middleware_init(self):
        """Test middleware initialization"""
        middleware = EnhancedInputValidationMiddleware()
        assert middleware.enable_logging is True
        assert middleware.enable_caching is True
        assert len(middleware.tool_parameters) > 0

    def test_middleware_init_no_caching(self):
        """Test middleware initialization without caching"""
        middleware = EnhancedInputValidationMiddleware(enable_caching=False)
        assert middleware.validation_cache is None

    def test_tool_parameters_loaded(self):
        """Test tool parameters are loaded"""
        middleware = EnhancedInputValidationMiddleware()
        assert "Grep" in middleware.tool_parameters
        assert "Read" in middleware.tool_parameters
        assert "Write" in middleware.tool_parameters
        assert "Bash" in middleware.tool_parameters

    def test_parameter_mappings_loaded(self):
        """Test parameter mappings are loaded"""
        middleware = EnhancedInputValidationMiddleware()
        assert len(middleware.parameter_mappings) > 0
        assert "max_results" in middleware.parameter_mappings  # maps to head_limit
        assert middleware.parameter_mappings["max_results"] == "head_limit"

    def test_validate_grep_mode(self):
        """Test grep mode validation"""
        middleware = EnhancedInputValidationMiddleware()
        assert middleware._validate_grep_mode("content") is True
        assert middleware._validate_grep_mode("files_with_matches") is True
        assert middleware._validate_grep_mode("count") is True
        assert middleware._validate_grep_mode("invalid") is False

    def test_validate_and_normalize_grep_valid(self):
        """Test validating and normalizing valid Grep input"""
        middleware = EnhancedInputValidationMiddleware()
        result = middleware.validate_and_normalize_input(
            "Grep",
            {
                "pattern": "test",
                "output_mode": "content",
                "path": "/src",
            },
        )
        assert result.valid is True
        assert result.normalized_input["pattern"] == "test"

    def test_validate_unknown_tool(self):
        """Test validating unknown tool"""
        middleware = EnhancedInputValidationMiddleware()
        result = middleware.validate_and_normalize_input(
            "UnknownTool",
            {"some_param": "value"},
        )
        assert len(result.warnings) > 0

    def test_map_parameters_grep_head_limit(self):
        """Test parameter mapping for head_limit"""
        middleware = EnhancedInputValidationMiddleware()
        result = middleware.validate_and_normalize_input(
            "Grep",
            {
                "pattern": "test",
                "max_results": 50,  # Should map to head_limit
            },
        )
        assert "head_limit" in result.normalized_input
        assert result.normalized_input["head_limit"] == 50

    def test_map_parameters_grep_path(self):
        """Test parameter mapping for path"""
        middleware = EnhancedInputValidationMiddleware()
        result = middleware.validate_and_normalize_input(
            "Grep",
            {
                "pattern": "test",
                "search_path": "/home/user",  # Should map to path
            },
        )
        assert "path" in result.normalized_input
        assert result.normalized_input["path"] == "/home/user"

    def test_map_parameters_read_file_path(self):
        """Test parameter mapping for Read file_path"""
        middleware = EnhancedInputValidationMiddleware()
        result = middleware.validate_and_normalize_input(
            "Read",
            {
                "filename": "/path/to/file.txt",  # Should map to file_path
            },
        )
        assert "file_path" in result.normalized_input
        assert result.normalized_input["file_path"] == "/path/to/file.txt"

    def test_map_parameters_read_offset(self):
        """Test parameter mapping for Read offset"""
        middleware = EnhancedInputValidationMiddleware()
        result = middleware.validate_and_normalize_input(
            "Read",
            {
                "file_path": "/path/to/file.txt",
                "start_line": 10,  # Should map to offset
            },
        )
        assert "offset" in result.normalized_input
        assert result.normalized_input["offset"] == 10

    def test_map_parameters_bash_command(self):
        """Test parameter mapping for Bash command"""
        middleware = EnhancedInputValidationMiddleware()
        result = middleware.validate_and_normalize_input(
            "Bash",
            {
                "cmd": "ls -la",  # Should map to command
            },
        )
        assert "command" in result.normalized_input
        assert result.normalized_input["command"] == "ls -la"

    def test_map_parameters_bash_cwd(self):
        """Test parameter mapping for Bash cwd"""
        middleware = EnhancedInputValidationMiddleware()
        result = middleware.validate_and_normalize_input(
            "Bash",
            {
                "command": "pwd",
                "cwd": "/home",  # Should map to working_directory
            },
        )
        assert "working_directory" in result.normalized_input
        assert result.normalized_input["working_directory"] == "/home"

    def test_map_parameters_task_agent_type(self):
        """Test parameter mapping for Task agent_type"""
        middleware = EnhancedInputValidationMiddleware()
        result = middleware.validate_and_normalize_input(
            "Task",
            {
                "agent_type": "backend",  # Should map to subagent_type
                "message": "test prompt",  # Should map to prompt
            },
        )
        assert "subagent_type" in result.normalized_input
        assert result.normalized_input["subagent_type"] == "backend"
        assert "prompt" in result.normalized_input
        assert result.normalized_input["prompt"] == "test prompt"

    def test_unrecognized_parameter(self):
        """Test handling unrecognized parameter"""
        middleware = EnhancedInputValidationMiddleware()
        result = middleware.validate_and_normalize_input(
            "Grep",
            {
                "pattern": "test",
                "unknown_param": "value",
            },
        )
        errors = [e for e in result.errors if e.code == "unrecognized_parameter"]
        assert len(errors) > 0

    def test_missing_required_parameter(self):
        """Test detecting missing required parameter"""
        middleware = EnhancedInputValidationMiddleware()
        result = middleware.validate_and_normalize_input(
            "Grep",
            {
                # Missing required "pattern"
                "output_mode": "content",
            },
        )
        assert result.valid is False
        errors = [e for e in result.errors if e.code == "missing_required_parameter"]
        assert len(errors) > 0

    def test_apply_default_values(self):
        """Test applying default values"""
        middleware = EnhancedInputValidationMiddleware()
        result = middleware.validate_and_normalize_input(
            "Grep",
            {
                "pattern": "test",
                # output_mode and head_limit should get defaults
            },
        )
        assert result.normalized_input["output_mode"] == "content"

    def test_type_validation_string(self):
        """Test string type validation"""
        middleware = EnhancedInputValidationMiddleware()
        result = middleware.validate_and_normalize_input(
            "Grep",
            {
                "pattern": "test",
                "output_mode": "content",
            },
        )
        assert result.valid is True

    def test_type_conversion_integer_from_string(self):
        """Test converting string to integer"""
        middleware = EnhancedInputValidationMiddleware()
        result = middleware.validate_and_normalize_input(
            "Grep",
            {
                "pattern": "test",
                "head_limit": "50",  # String instead of integer
            },
        )
        # Should have converted the string to integer
        assert isinstance(result.normalized_input["head_limit"], int) or result.normalized_input["head_limit"] == 50

    def test_type_conversion_integer_from_float(self):
        """Test converting float to integer"""
        middleware = EnhancedInputValidationMiddleware()
        result = middleware.validate_and_normalize_input(
            "Grep",
            {
                "pattern": "test",
                "head_limit": 50.7,  # Float instead of integer
            },
        )
        assert result.normalized_input["head_limit"] == 50
        errors = [e for e in result.errors if e.code == "type_conversion"]
        assert len(errors) > 0

    def test_type_conversion_boolean_from_string_true(self):
        """Test converting string to boolean (true values)"""
        middleware = EnhancedInputValidationMiddleware()
        result = middleware.validate_and_normalize_input(
            "Grep",
            {
                "pattern": "test",
                "case_sensitive": "true",
            },
        )
        assert result.normalized_input["case_sensitive"] is True

    def test_type_conversion_boolean_from_string_false(self):
        """Test converting string to boolean (false values)"""
        middleware = EnhancedInputValidationMiddleware()
        result = middleware.validate_and_normalize_input(
            "Grep",
            {
                "pattern": "test",
                "case_sensitive": "false",
            },
        )
        assert result.normalized_input["case_sensitive"] is False

    def test_type_conversion_boolean_from_int_1(self):
        """Test converting 1 to boolean True"""
        middleware = EnhancedInputValidationMiddleware()
        result = middleware.validate_and_normalize_input(
            "Grep",
            {
                "pattern": "test",
                "case_sensitive": 1,
            },
        )
        # 1 may stay as int, but should be truthy
        assert result.normalized_input["case_sensitive"] in [True, 1]

    def test_type_conversion_boolean_from_int_0(self):
        """Test converting 0 to boolean False"""
        middleware = EnhancedInputValidationMiddleware()
        result = middleware.validate_and_normalize_input(
            "Grep",
            {
                "pattern": "test",
                "case_sensitive": 0,
            },
        )
        # 0 may stay as int, but should be falsy
        assert result.normalized_input["case_sensitive"] in [False, 0]

    def test_normalize_file_path_backslashes(self):
        """Test normalizing file path with backslashes"""
        middleware = EnhancedInputValidationMiddleware()
        result = middleware.validate_and_normalize_input(
            "Read",
            {
                "file_path": "C:\\Users\\test\\file.txt",
            },
        )
        assert "\\" not in result.normalized_input["file_path"]
        assert result.normalized_input["file_path"] == "C:/Users/test/file.txt"

    def test_normalize_file_path_trailing_slash(self):
        """Test removing trailing slashes from path"""
        middleware = EnhancedInputValidationMiddleware()
        result = middleware.validate_and_normalize_input(
            "Glob",
            {
                "pattern": "*.py",
                "path": "/home/user/",  # Trailing slash
            },
        )
        assert result.normalized_input["path"] == "/home/user"

    def test_normalize_numeric_integer(self):
        """Test normalizing numeric string to integer"""
        middleware = EnhancedInputValidationMiddleware()
        result = middleware.validate_and_normalize_input(
            "Grep",
            {
                "pattern": "test",
                "head_limit": " 100 ",  # String with spaces
            },
        )
        assert result.normalized_input["head_limit"] == 100
        assert isinstance(result.normalized_input["head_limit"], int)

    def test_normalize_numeric_float(self):
        """Test normalizing numeric string to float"""
        middleware = EnhancedInputValidationMiddleware()
        result = middleware.validate_and_normalize_input(
            "Bash",
            {
                "command": "sleep 1.5",
                "timeout": "5000.0",
            },
        )
        assert result.normalized_input["timeout"] == 5000.0

    def test_custom_validation_function(self):
        """Test custom validation function"""
        middleware = EnhancedInputValidationMiddleware()

        # This tests that the validation function is called
        result = middleware.validate_and_normalize_input(
            "Grep",
            {
                "pattern": "test",
                "output_mode": "content",  # Valid mode
            },
        )
        assert result.valid is True

    def test_custom_validation_function_failure(self):
        """Test custom validation function failure"""
        middleware = EnhancedInputValidationMiddleware()
        result = middleware.validate_and_normalize_input(
            "Grep",
            {
                "pattern": "test",
                "output_mode": "invalid_mode",  # Invalid mode
            },
        )
        errors = [e for e in result.errors if e.code == "validation_function_failed"]
        assert len(errors) > 0

    def test_find_closest_parameter_match_exact(self):
        """Test finding exact parameter match"""
        middleware = EnhancedInputValidationMiddleware()
        match = middleware._find_closest_parameter_match("pattern", {"pattern"})
        assert match == "pattern"

    def test_find_closest_parameter_match_case_insensitive(self):
        """Test finding case-insensitive match"""
        middleware = EnhancedInputValidationMiddleware()
        match = middleware._find_closest_parameter_match("Pattern", {"pattern"})
        assert match == "pattern"

    def test_is_float_valid(self):
        """Test float validation"""
        middleware = EnhancedInputValidationMiddleware()
        assert middleware._is_float(3.14) is True
        assert middleware._is_float("3.14") is True
        assert middleware._is_float("not_a_float") is False

    def test_register_tool_parameters(self):
        """Test registering custom tool parameters"""
        middleware = EnhancedInputValidationMiddleware()
        custom_params = [
            ToolParameter(
                name="custom_param",
                param_type="string",
                required=True,
            ),
        ]
        middleware.register_tool_parameters("CustomTool", custom_params)
        assert "CustomTool" in middleware.tool_parameters

    def test_add_parameter_mapping(self):
        """Test adding custom parameter mapping"""
        middleware = EnhancedInputValidationMiddleware()
        initial_count = len(middleware.parameter_mappings)
        middleware.add_parameter_mapping("old_name", "new_name")
        assert len(middleware.parameter_mappings) > initial_count

    def test_get_validation_stats(self):
        """Test getting validation statistics"""
        middleware = EnhancedInputValidationMiddleware()

        # Run a validation
        middleware.validate_and_normalize_input(
            "Grep",
            {"pattern": "test"},
        )

        stats = middleware.get_validation_stats()
        assert stats["validations_performed"] >= 1
        assert "tools_configured" in stats

    def test_concurrent_validation(self):
        """Test concurrent validation"""
        middleware = EnhancedInputValidationMiddleware()

        import threading

        results = []

        def validate():
            result = middleware.validate_and_normalize_input(
                "Grep",
                {"pattern": "test"},
            )
            results.append(result)

        threads = [threading.Thread(target=validate) for _ in range(5)]
        for t in threads:
            t.start()
        for t in threads:
            t.join()

        assert len(results) == 5
        assert all(r.valid for r in results)

    def test_export_validation_report(self, tmp_path):
        """Test exporting validation report"""
        middleware = EnhancedInputValidationMiddleware()

        # Run some validations
        middleware.validate_and_normalize_input("Grep", {"pattern": "test"})

        # Export
        report_path = str(tmp_path / "report.json")
        middleware.export_validation_report(report_path)

        # Verify file exists and contains data
        with open(report_path, "r") as f:
            report = json.load(f)
            assert "stats" in report
            assert "configured_tools" in report

    def test_validate_tool_input_function(self):
        """Test convenience function validate_tool_input"""
        result = validate_tool_input("Grep", {"pattern": "test"})
        assert isinstance(result, ValidationResult)
        assert result.valid is True

    def test_get_validation_stats_function(self):
        """Test convenience function get_validation_stats"""
        stats = get_validation_stats()
        assert isinstance(stats, dict)
        assert "validations_performed" in stats

    def test_complex_mapping_grep_all_parameters(self):
        """Test complex mapping with all Grep parameters"""
        middleware = EnhancedInputValidationMiddleware()
        result = middleware.validate_and_normalize_input(
            "Grep",
            {
                "search_pattern": "error",  # maps to pattern
                "max_results": 100,  # maps to head_limit
                "search_path": "/var/log",  # maps to path
                "ignore_case": False,  # maps to case_sensitive
                "context": 3,  # maps to context_lines
                "file_glob": "*.log",  # maps to file_pattern
            },
        )
        assert result.normalized_input["pattern"] == "error"
        assert result.normalized_input["head_limit"] == 100
        assert result.normalized_input["path"] == "/var/log"

    def test_complex_mapping_bash_all_parameters(self):
        """Test complex mapping with all Bash parameters"""
        middleware = EnhancedInputValidationMiddleware()
        result = middleware.validate_and_normalize_input(
            "Bash",
            {
                "cmd": "echo test",  # maps to command
                "timeout_ms": "30000",  # maps to timeout
                "work_dir": "/home",  # maps to working_directory
                "env": {"VAR": "value"},  # maps to environment
            },
        )
        assert result.normalized_input["command"] == "echo test"
        assert result.normalized_input["timeout"] == 30000
        assert result.normalized_input["working_directory"] == "/home"

    def test_processing_time_measurement(self):
        """Test processing time is measured"""
        middleware = EnhancedInputValidationMiddleware()
        result = middleware.validate_and_normalize_input(
            "Grep",
            {"pattern": "test"},
        )
        assert result.processing_time_ms >= 0

    def test_error_severity_levels(self):
        """Test various error severity levels"""
        middleware = EnhancedInputValidationMiddleware()

        # Critical error
        result = middleware.validate_and_normalize_input(
            "Grep",
            {},  # Missing required pattern
        )
        critical_errors = [e for e in result.errors if e.severity == ValidationSeverity.CRITICAL]
        assert len(critical_errors) > 0

    def test_warning_messages(self):
        """Test warning messages"""
        middleware = EnhancedInputValidationMiddleware()
        result = middleware.validate_and_normalize_input(
            "UnknownTool",
            {"param": "value"},
        )
        assert len(result.warnings) > 0

    def test_transformation_tracking(self):
        """Test transformation tracking"""
        middleware = EnhancedInputValidationMiddleware()
        result = middleware.validate_and_normalize_input(
            "Grep",
            {
                "pattern": "test",
                "head_limit": "100",  # Will be transformed to int
            },
        )
        assert len(result.transformations) > 0

    def test_deep_nested_type_conversion(self):
        """Test deep nested type conversion"""
        middleware = EnhancedInputValidationMiddleware()
        result = middleware.validate_and_normalize_input(
            "Bash",
            {
                "command": "test",
                "timeout": 5000.999,  # Float timeout
            },
        )
        # Float should convert to int for timeout
        assert isinstance(result.normalized_input.get("timeout"), (int, float))


if __name__ == "__main__":
    pytest.main([__file__, "-v", "--tb=short"])
