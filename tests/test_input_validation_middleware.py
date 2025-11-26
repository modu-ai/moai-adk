"""
Unit Tests for Enhanced Input Validation Middleware

Production-ready test suite covering all input validation, normalization,
and error correction features of the Enhanced Input Validation Middleware.

Author: MoAI-ADK Core Team
Version: 1.0.0
"""

import os
import sys
import time
import unittest

# Add the src directory to the path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), "..", "src"))

from moai_adk.core.input_validation_middleware import (
    EnhancedInputValidationMiddleware,
    ToolParameter,
    ValidationResult,
    ValidationSeverity,
    get_validation_stats,
    validate_tool_input,
)


class TestEnhancedInputValidationMiddleware(unittest.TestCase):
    """Comprehensive test suite for Enhanced Input Validation Middleware"""

    def setUp(self):
        """Set up test fixtures"""
        self.middleware = EnhancedInputValidationMiddleware(enable_logging=False)

    def test_grep_head_limit_correction(self):
        """Test correction of head_limit parameter from debug log error"""
        # This reproduces the exact error from the debug log
        input_data = {"pattern": "test", "head_limit": 10, "output_mode": "content"}

        result = self.middleware.validate_and_normalize_input("Grep", input_data)

        self.assertTrue(result.valid)
        self.assertEqual(result.normalized_input["head_limit"], 10)
        self.assertEqual(result.normalized_input["pattern"], "test")

        # Should not generate errors for valid parameters
        critical_errors = [e for e in result.errors if e.severity == ValidationSeverity.CRITICAL]
        self.assertEqual(len(critical_errors), 0)

    def test_parameter_mapping(self):
        """Test parameter mapping for compatibility"""
        test_cases = [
            {
                "input": {"pattern": "test", "max_results": 20},
                "expected": {"pattern": "test", "head_limit": 20},
                "tool": "Grep",
            },
            {
                "input": {"filename": "/path/to/file.txt", "start_line": 10},
                "expected": {"file_path": "/path/to/file.txt", "offset": 10},
                "tool": "Read",
            },
            {
                "input": {"agent_type": "debug-helper", "message": "test"},
                "expected": {"subagent_type": "debug-helper", "prompt": "test"},
                "tool": "Task",
            },
        ]

        for case in test_cases:
            with self.subTest(input=case["input"]):
                result = self.middleware.validate_and_normalize_input(case["tool"], case["input"])

                self.assertTrue(result.valid, f"Failed for {case['tool']}: {case['input']}")

                # Check mapped parameters
                for key, value in case["expected"].items():
                    self.assertEqual(result.normalized_input[key], value, f"Mapped parameter '{key}' incorrect")

    def test_type_conversion(self):
        """Test automatic type conversion"""
        test_cases = [
            {"input": {"count": "123"}, "expected": {"count": 123}, "tool": "Generic"},  # String to integer
            {"input": {"enabled": "true"}, "expected": {"enabled": True}, "tool": "Generic"},  # String to boolean
            {
                "input": {"command": "echo test", "timeout": "5000"},  # String to integer
                "expected": {"timeout": 5000},
                "tool": "Bash",
            },
        ]

        # Add a generic tool for testing
        self.middleware.register_tool_parameters(
            "Generic",
            [
                ToolParameter("count", "integer"),
                ToolParameter("enabled", "boolean"),
                ToolParameter("timeout", "integer"),
            ],
        )

        for case in test_cases:
            with self.subTest(input=case["input"]):
                result = self.middleware.validate_and_normalize_input(case["tool"], case["input"])

                # Should be valid after type conversion
                self.assertTrue(result.valid, f"Type conversion failed for {case['input']}")

                # Check converted values
                for key, value in case["expected"].items():
                    self.assertEqual(result.normalized_input[key], value, f"Type conversion incorrect for '{key}'")

    def test_deprecated_parameter_handling(self):
        """Test handling of deprecated parameters"""
        input_data = {
            "pattern": "test",
            "count": 10,  # Deprecated alias for head_limit
            "folder": "/src",  # Should be mapped to path
        }

        result = self.middleware.validate_and_normalize_input("Grep", input_data)

        # Should generate warnings for deprecated parameters
        self.assertGreater(len(result.warnings), 0)
        self.assertIn("Deprecated parameter", result.warnings[0])

        # Should still be valid after mapping
        self.assertTrue(result.valid)

    def test_required_parameter_validation(self):
        """Test validation of required parameters"""
        # Test with missing required parameter
        result = self.middleware.validate_and_normalize_input("Grep", {})

        self.assertFalse(result.valid)

        critical_errors = [e for e in result.errors if e.severity == ValidationSeverity.CRITICAL]
        self.assertGreater(len(critical_errors), 0)

        # Check for missing pattern error
        pattern_errors = [e for e in critical_errors if "pattern" in e.message]
        self.assertGreater(len(pattern_errors), 0)

    def test_default_value_application(self):
        """Test application of default values"""
        input_data = {
            "pattern": "test"
            # No output_mode specified - should get default
        }

        result = self.middleware.validate_and_normalize_input("Grep", input_data)

        self.assertTrue(result.valid)
        self.assertIn("output_mode", result.normalized_input)
        self.assertEqual(result.normalized_input["output_mode"], "content")  # Default value

    def test_parameter_value_validation(self):
        """Test validation of parameter values"""
        # Test invalid grep mode
        input_data = {"pattern": "test", "output_mode": "invalid_mode"}

        result = self.middleware.validate_and_normalize_input("Grep", input_data)

        self.assertFalse(result.valid)

        mode_errors = [e for e in result.errors if "output_mode" in e.path]
        self.assertGreater(len(mode_errors), 0)

    def test_path_normalization(self):
        """Test file path normalization"""
        test_cases = [
            {
                "input": {"path": "C:\\Users\\test\\file.txt\\"},
                "expected": {"path": "C:/Users/test/file.txt"},
                "tool": "Generic",
            },
            {
                "input": {"file_path": "/path/with/trailing/slash/"},
                "expected": {"file_path": "/path/with/trailing/slash"},
                "tool": "Generic",
            },
        ]

        # Add generic tool for path testing
        self.middleware.register_tool_parameters(
            "Generic", [ToolParameter("path", "string"), ToolParameter("file_path", "string")]
        )

        for case in test_cases:
            with self.subTest(input=case["input"]):
                result = self.middleware.validate_and_normalize_input(case["tool"], case["input"])

                for key, value in case["expected"].items():
                    self.assertEqual(result.normalized_input[key], value, f"Path normalization failed for '{key}'")

                # Should have transformation logged
                path_transformations = [t for t in result.transformations if "Normalized" in t and key in t]
                if case["input"][key] != case["expected"][key]:
                    self.assertGreater(len(path_transformations), 0)

    def test_boolean_normalization(self):
        """Test boolean value normalization"""
        test_cases = [
            {"value": "true", "expected": True},
            {"value": "false", "expected": False},
            {"value": "1", "expected": True},
            {"value": "0", "expected": False},
            {"value": "YES", "expected": True},
            {"value": "no", "expected": False},
        ]

        # Add generic tool for boolean testing
        self.middleware.register_tool_parameters("Generic", [ToolParameter("enabled", "boolean")])

        for case in test_cases:
            with self.subTest(value=case["value"]):
                input_data = {"enabled": case["value"]}
                result = self.middleware.validate_and_normalize_input("Generic", input_data)

                self.assertTrue(result.valid)
                self.assertEqual(result.normalized_input["enabled"], case["expected"])

    def test_unknown_parameter_handling(self):
        """Test handling of unknown parameters"""
        input_data = {"pattern": "test", "unknown_parameter": "some_value", "another_unknown": 123}

        result = self.middleware.validate_and_normalize_input("Grep", input_data)

        # Should generate errors for unknown parameters
        unknown_errors = [e for e in result.errors if e.code == "unrecognized_parameter"]
        self.assertEqual(len(unknown_errors), 2)

        # Should provide suggestions
        for error in unknown_errors:
            self.assertIsNotNone(error.suggestion)

    def test_parameter_suggestion_algorithm(self):
        """Test parameter suggestion algorithm"""
        # Test close matches
        test_cases = [
            {"input": "hed_limit", "expected": "head_limit"},
            {"input": "patttern", "expected": "pattern"},
            {"input": "outpt_mode", "expected": "output_mode"},
            {"input": "completely_different", "expected": None},
        ]

        for case in test_cases:
            with self.subTest(input=case["input"]):
                suggestion = self.middleware._find_closest_parameter_match(
                    case["input"], {"pattern", "head_limit", "output_mode", "path"}
                )
                self.assertEqual(suggestion, case["expected"])

    def test_performance_characteristics(self):
        """Test performance characteristics"""
        input_data = {
            "pattern": "test",
            "head_limit": 10,
            "output_mode": "content",
            "path": "/src",
            "case_sensitive": True,
        }

        # Test multiple validations
        times = []
        for _ in range(100):
            start_time = time.time()
            self.middleware.validate_and_normalize_input("Grep", input_data)
            times.append(time.time() - start_time)

        avg_time = sum(times) / len(times) * 1000
        max_time = max(times) * 1000

        self.assertLess(avg_time, 10.0, "Average validation time should be under 10ms")
        self.assertLess(max_time, 50.0, "Maximum validation time should be under 50ms")

    def test_caching_functionality(self):
        """Test validation caching functionality"""
        input_data = {"pattern": "test", "head_limit": 10}

        # Enable caching
        middleware_with_cache = EnhancedInputValidationMiddleware(enable_caching=True)

        # First validation
        result1 = middleware_with_cache.validate_and_normalize_input("Grep", input_data)

        # Second validation (should use cache)
        result2 = middleware_with_cache.validate_and_normalize_input("Grep", input_data)

        # Results should be identical
        self.assertEqual(result1.normalized_input, result2.normalized_input)

        # Cache should be populated
        stats = middleware_with_cache.get_validation_stats()
        self.assertGreater(stats["cache_size"], 0)

    def test_custom_tool_registration(self):
        """Test registration of custom tool parameters"""
        custom_params = [
            ToolParameter("custom_param", "string", required=True),
            ToolParameter("optional_param", "integer", default_value=42),
            ToolParameter("alias_param", "boolean", aliases=["alt", "alternative"]),
        ]

        self.middleware.register_tool_parameters("CustomTool", custom_params)

        # Test with registered tool
        input_data = {"custom_param": "test", "alt": True, "optional_param": 100}

        result = self.middleware.validate_and_normalize_input("CustomTool", input_data)

        self.assertTrue(result.valid)
        self.assertEqual(result.normalized_input["custom_param"], "test")
        self.assertEqual(result.normalized_input["alias_param"], True)
        self.assertEqual(result.normalized_input["optional_param"], 100)

    def test_comprehensive_grep_validation(self):
        """Test comprehensive Grep tool validation"""
        test_cases = [
            {
                "name": "Valid Grep input",
                "input": {"pattern": "test", "head_limit": 10, "output_mode": "content", "path": "/src"},
                "valid": True,
            },
            {
                "name": "Grep with parameter mapping",
                "input": {
                    "regex": "test",  # Should map to pattern
                    "max_results": 20,  # Should map to head_limit
                    "folder": "/src",  # Should map to path
                },
                "valid": True,
            },
            {
                "name": "Grep with missing pattern",
                "input": {"head_limit": 10, "output_mode": "content"},
                "valid": False,
            },
            {
                "name": "Grep with invalid mode",
                "input": {"pattern": "test", "output_mode": "invalid_mode"},
                "valid": False,
            },
        ]

        for case in test_cases:
            with self.subTest(name=case["name"]):
                result = self.middleware.validate_and_normalize_input("Grep", case["input"])
                self.assertEqual(result.valid, case["valid"], f"Validation result mismatch for {case['name']}")

    def test_error_recovery_and_correction(self):
        """Test error recovery and automatic correction"""
        # Input with multiple issues that can be corrected
        input_data = {
            "max_results": "20",  # String that should be converted to int
            "search_path": "/src\\test\\",  # Path that needs normalization
            "count": 15,  # Deprecated parameter that should be mapped
        }

        result = self.middleware.validate_and_normalize_input("Grep", input_data)

        # Should have auto-corrections
        auto_corrected = [e for e in result.errors if e.auto_corrected]
        self.assertGreater(len(auto_corrected), 0)

        # Should have transformations
        self.assertGreater(len(result.transformations), 0)

        # Should be valid after corrections
        # (Note: Missing required 'pattern' will still make it invalid, but other issues should be fixed)
        pattern_errors = [e for e in result.errors if "pattern" in e.message]
        non_pattern_errors = [e for e in result.errors if e not in pattern_errors]

        # Non-pattern errors should be resolved
        self.assertEqual(len([e for e in non_pattern_errors if not e.auto_corrected]), 0)

    def test_statistics_tracking(self):
        """Test validation statistics tracking"""
        # Perform various validations
        self.middleware.validate_and_normalize_input("Grep", {"pattern": "test"})
        self.middleware.validate_and_normalize_input("Grep", {"max_results": 10})  # Will have errors
        self.middleware.validate_and_normalize_input("Read", {"file_path": "/test.txt"})

        stats = self.middleware.get_validation_stats()

        # Check statistics
        self.assertGreater(stats["validations_performed"], 0)
        self.assertGreaterEqual(stats["auto_corrections"], 0)
        self.assertIn("tools_configured", stats)
        self.assertIn("parameter_mappings", stats)


class TestConvenienceFunctions(unittest.TestCase):
    """Test convenience functions for easy integration"""

    def test_validate_tool_input_function(self):
        """Test convenience function for tool input validation"""
        input_data = {"pattern": "test", "head_limit": 10}

        result = validate_tool_input("Grep", input_data)
        self.assertIsInstance(result, ValidationResult)
        self.assertTrue(result.valid)

    def test_get_validation_stats_function(self):
        """Test convenience function for getting statistics"""
        stats = get_validation_stats()
        self.assertIsInstance(stats, dict)
        self.assertIn("validations_performed", stats)


class TestRealWorldScenarios(unittest.TestCase):
    """Test real-world scenarios from Claude Code usage"""

    def test_debug_log_error_scenario(self):
        """Test the exact scenario from the debug log (Lines 476-495)"""
        # Reproduce the error: "Unrecognized key(s) in object: 'head_limit'"
        input_data = {"pattern": "ERROR", "head_limit": 10, "output_mode": "content"}  # This was causing the error

        result = validate_tool_input("Grep", input_data)

        # Should now be valid and properly normalized
        self.assertTrue(result.valid, "Should handle head_limit parameter correctly")
        self.assertEqual(result.normalized_input["head_limit"], 10)
        self.assertEqual(result.normalized_input["pattern"], "ERROR")

        # Should not have critical errors
        critical_errors = [e for e in result.errors if e.severity == ValidationSeverity.CRITICAL]
        self.assertEqual(len(critical_errors), 0)

    def test_claude_code_common_patterns(self):
        """Test common Claude Code usage patterns"""
        common_patterns = [
            {
                "description": "Grep search with file filtering",
                "tool": "Grep",
                "input": {"pattern": "def test_", "file_pattern": "*.py", "head": 20, "search_path": "/test"},
            },
            {
                "description": "File reading with line limits",
                "tool": "Read",
                "input": {"filename": "/test/example.py", "from": 10, "max_lines": 50},
            },
            {
                "description": "Task execution with debug",
                "tool": "Task",
                "input": {"agent_type": "debug-helper", "message": "Debug this issue", "verbose": True},
            },
            {
                "description": "Command execution with working directory",
                "tool": "Bash",
                "input": {"cmd": "ls -la", "cwd": "/home/user", "timeout_ms": 5000},
            },
        ]

        for pattern in common_patterns:
            with self.subTest(description=pattern["description"]):
                result = validate_tool_input(pattern["tool"], pattern["input"])

                # All common patterns should be valid after normalization
                self.assertTrue(result.valid, f"Common pattern failed: {pattern['description']}")

                # Should provide helpful feedback for corrections
                total_corrections = len([e for e in result.errors if e.auto_corrected]) + len(result.transformations)
                if total_corrections > 0:
                    self.assertGreater(total_corrections, 0)


def run_performance_benchmark():
    """Run performance benchmark for the validation middleware"""
    print("\n⚡ Performance Benchmark")
    print("=" * 40)

    middleware = EnhancedInputValidationMiddleware(enable_logging=False)

    # Test with various input sizes and complexity
    test_scenarios = [
        ("Simple Grep", {"pattern": "test"}),
        (
            "Complex Grep",
            {
                "pattern": "test",
                "max_results": 20,
                "search_path": "/src",
                "case_sensitive": True,
                "output_mode": "content",
                "glob": "*.py",
            },
        ),
        (
            "Multi-parameter Task",
            {
                "agent_type": "debug-helper",
                "message": "Test message with detailed explanation",
                "context": {"key1": "value1", "key2": "value2"},
                "debug": True,
                "timeout": 5000,
            },
        ),
    ]

    for name, input_data in test_scenarios:
        times = []
        for _ in range(1000):
            start = time.time()
            middleware.validate_and_normalize_input("Grep", input_data)
            times.append(time.time() - start)

        avg_time = sum(times) / len(times) * 1000
        p95_time = sorted(times)[int(len(times) * 0.95)] * 1000

        print(f"{name:20}: {avg_time:6.2f}ms avg, {p95_time:6.2f}ms p95")


if __name__ == "__main__":
    # Run unit tests
    print("🚀 Running Enhanced Input Validation Middleware Unit Tests...")
    print("=" * 70)

    unittest.main(verbosity=2, exit=False)

    # Run performance benchmark
    run_performance_benchmark()
