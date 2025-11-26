"""
Unit Tests for Robust JSON Parser

Production-ready test suite covering all error recovery strategies,
edge cases, and performance requirements.

Author: MoAI-ADK Core Team
Version: 1.0.0
"""

import json
import os
import sys
import time
import unittest

# Add the src directory to the path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), "..", "src"))

from moai_adk.core.robust_json_parser import (
    ErrorSeverity,
    ParseResult,
    RobustJSONParser,
    get_parser_stats,
    parse_json,
    reset_parser_stats,
)


class TestRobustJSONParser(unittest.TestCase):
    """Comprehensive test suite for Robust JSON Parser"""

    def setUp(self):
        """Set up test fixtures"""
        self.parser = RobustJSONParser(max_recovery_attempts=3, enable_logging=False)

    def test_valid_json_direct_parse(self):
        """Test parsing valid JSON without any recovery needed"""
        test_json = '{"name": "test", "value": 123, "active": true}'
        result = self.parser.parse(test_json)

        self.assertTrue(result.success)
        self.assertEqual(result.recovery_attempts, 0)
        self.assertEqual(result.severity, ErrorSeverity.LOW)
        self.assertEqual(result.data["name"], "test")
        self.assertEqual(result.data["value"], 123)
        self.assertTrue(result.data["active"])

    def test_missing_quotes_recovery(self):
        """Test recovery of missing quotes around property names"""
        test_cases = [
            '{name: "test", value: 123}',
            '{ "name": "test", age: 30 }',
            '{"person": {name: "John", age: 25}}',
            "{items: [1, 2, 3], count: 3}",
        ]

        for test_json in test_cases:
            with self.subTest(test_json=test_json):
                result = self.parser.parse(test_json)
                self.assertTrue(result.success, f"Failed to parse: {test_json}")
                self.assertGreater(result.recovery_attempts, 0)
                self.assertTrue(any("quotes" in warning.lower() for warning in result.warnings))

    def test_trailing_comma_recovery(self):
        """Test recovery of trailing commas in objects and arrays"""
        test_cases = [
            '{"name": "test", "value": 123,}',
            "[1, 2, 3,]",
            '{"items": [1, 2,], "count": 2}',
            '{"nested": {"a": 1, "b": 2,}, "valid": true}',
        ]

        for test_json in test_cases:
            with self.subTest(test_json=test_json):
                result = self.parser.parse(test_json)
                self.assertTrue(result.success, f"Failed to parse: {test_json}")
                self.assertGreater(result.recovery_attempts, 0)
                self.assertTrue(any("comma" in warning.lower() for warning in result.warnings))

    def test_escape_sequence_recovery(self):
        """Test recovery of invalid escape sequences"""
        test_cases = [
            '{"name": "test\\invalid", "value": 123}',
            '{"path": "C:\\\\Users\\\\test\\x", "valid": true}',
            '{"unicode": "test\\uXYZ", "normal": "text"}',
        ]

        for test_json in test_cases:
            with self.subTest(test_json=test_json):
                result = self.parser.parse(test_json)
                self.assertTrue(result.success, f"Failed to parse: {test_json}")
                self.assertGreater(result.recovery_attempts, 0)

    def test_partial_object_recovery(self):
        """Test recovery of incomplete JSON objects"""
        test_cases = [
            '{"name": "test"',
            '{"items": [1, 2, 3]',
            "[1, 2, 3",
            '{"nested": {"a": 1}',
        ]

        for test_json in test_cases:
            with self.subTest(test_json=test_json):
                result = self.parser.parse(test_json)
                self.assertTrue(result.success, f"Failed to parse: {test_json}")
                self.assertGreater(result.recovery_attempts, 0)
                self.assertTrue(
                    any("brace" in warning.lower() or "bracket" in warning.lower() for warning in result.warnings)
                )

    def test_single_quote_recovery(self):
        """Test recovery of single quotes instead of double quotes"""
        test_cases = [
            "{'name': 'test', 'value': 123}",
            "{'items': ['a', 'b', 'c']}",
            "{'mixed': 'double', 'single': 'quotes'}",
        ]

        for test_json in test_cases:
            with self.subTest(test_json=test_json):
                result = self.parser.parse(test_json)
                self.assertTrue(result.success, f"Failed to parse: {test_json}")
                self.assertGreater(result.recovery_attempts, 0)

    def test_control_character_removal(self):
        """Test removal of control characters"""
        test_json = '{"name": "test\x00", "value": 123\x01\x02}'
        result = self.parser.parse(test_json)

        self.assertTrue(result.success)
        self.assertGreater(result.recovery_attempts, 0)
        self.assertTrue(any("control" in warning.lower() for warning in result.warnings))

    def test_complex_multiple_errors(self):
        """Test recovery of JSON with multiple types of errors"""
        test_json = "{name: 'test', value: 123, active: true,}\x00"
        result = self.parser.parse(test_json)

        self.assertTrue(result.success)
        self.assertGreater(result.recovery_attempts, 0)
        self.assertTrue(len(result.warnings) >= 2)  # Should have multiple warnings

    def test_partial_extraction(self):
        """Test extraction of valid JSON from larger text"""
        test_cases = [
            'Some text {"name": "test", "value": 123} more text',
            'Response: {"status": "success", "data": [1, 2, 3]} End',
            'Multiple {"first": 1} and {"second": 2} objects',
        ]

        for test_json in test_cases:
            with self.subTest(test_json=test_json):
                result = self.parser.parse(test_json)
                # Should extract the first valid JSON object
                self.assertTrue(result.success, f"Failed to extract from: {test_json}")

    def test_markdown_code_block_extraction(self):
        """Test extraction of JSON from markdown code blocks"""
        test_cases = [
            '```json\n{"name": "test", "value": 123}\n```',
            'Here is some JSON:\n```json\n{"status": "ok"}\n```\nEnd',
        ]

        for test_json in test_cases:
            with self.subTest(test_json=test_json):
                result = self.parser.parse(test_json)
                self.assertTrue(result.success, f"Failed to extract from markdown: {test_json}")

    def test_invalid_input_types(self):
        """Test handling of invalid input types"""
        invalid_inputs = [
            None,
            123,
            [],
            {"key": "value"},  # Dict instead of string
        ]

        for invalid_input in invalid_inputs:
            with self.subTest(invalid_input=invalid_input):
                result = self.parser.parse(invalid_input)
                self.assertFalse(result.success)
                self.assertEqual(result.severity, ErrorSeverity.CRITICAL)

    def test_empty_and_whitespace_inputs(self):
        """Test handling of empty and whitespace-only inputs"""
        test_cases = [
            "",
            "   ",
            "\n\t  ",
        ]

        for test_json in test_cases:
            with self.subTest(test_json=test_json):
                result = self.parser.parse(test_json)
                self.assertFalse(result.success)

    def test_large_json_performance(self):
        """Test performance with large JSON objects"""
        # Create a large JSON object
        large_data = {
            "users": [{"id": i, "name": f"user_{i}", "data": list(range(100))} for i in range(1000)],
            "metadata": {"total": 1000, "page": 1},
        }

        large_json = json.dumps(large_data)

        start_time = time.time()
        result = self.parser.parse(large_json)
        parse_time = time.time() - start_time

        self.assertTrue(result.success)
        self.assertLess(parse_time, 1.0)  # Should parse in under 1 second
        self.assertEqual(result.parse_time_ms, parse_time * 1000, msg="Parse time mismatch")

    def test_nested_structures(self):
        """Test parsing of deeply nested JSON structures"""
        nested_json = {"level1": {"level2": {"level3": {"level4": {"level5": "deep_value"}}}}}

        # Test with valid nested JSON
        valid_json = json.dumps(nested_json)
        result = self.parser.parse(valid_json)
        self.assertTrue(result.success)

        # Test with nested JSON that has errors
        broken_nested = '{"level1": {"level2": {"level3": {level4: {"level5": "deep_value"}}}}}'
        result = self.parser.parse(broken_nested)
        self.assertTrue(result.success)
        self.assertGreater(result.recovery_attempts, 0)

    def test_unicode_and_special_characters(self):
        """Test handling of Unicode and special characters"""
        test_cases = [
            '{"emoji": "😀", "chinese": "中文", "arabic": "العربية"}',
            '{"quotes": "She said \\"Hello\\" to him"}',
            '{"newlines": "Line 1\\nLine 2\\nLine 3"}',
            '{"tabs": "Col1\\tCol2\\tCol3"}',
            '{"backslashes": "C:\\\\Users\\\\test"}',
        ]

        for test_json in test_cases:
            with self.subTest(test_json=test_json):
                result = self.parser.parse(test_json)
                self.assertTrue(result.success, f"Failed to parse: {test_json}")

    def test_numeric_values(self):
        """Test parsing of various numeric formats"""
        test_cases = [
            '{"integer": 42}',
            '{"negative": -123}',
            '{"float": 3.14159}',
            '{"scientific": 1.23e-4}',
            '{"zero": 0}',
        ]

        for test_json in test_cases:
            with self.subTest(test_json=test_json):
                result = self.parser.parse(test_json)
                self.assertTrue(result.success, f"Failed to parse: {test_json}")

    def test_boolean_and_null_values(self):
        """Test parsing of boolean and null values"""
        test_cases = [
            '{"bool_true": true}',
            '{"bool_false": false}',
            '{"null_value": null}',
            '{"mixed": {"a": true, "b": false, "c": null}}',
        ]

        for test_json in test_cases:
            with self.subTest(test_json=test_json):
                result = self.parser.parse(test_json)
                self.assertTrue(result.success, f"Failed to parse: {test_json}")

    def test_statistics_tracking(self):
        """Test that statistics are properly tracked"""
        # Reset stats
        self.parser.reset_stats()

        # Perform some parses
        self.parser.parse('{"valid": true}')  # Should succeed
        self.parser.parse("{invalid: json}")  # Should recover
        self.parser.parse("not json at all")  # Should fail

        stats = self.parser.get_stats()

        self.assertEqual(stats["total_parses"], 3)
        self.assertEqual(stats["successful_parses"], 1)
        self.assertEqual(stats["recovered_parses"], 1)
        self.assertEqual(stats["failed_parses"], 1)
        self.assertEqual(stats["success_rate"], 2 / 3)
        self.assertEqual(stats["recovery_rate"], 1 / 3)
        self.assertEqual(stats["failure_rate"], 1 / 3)

    def test_convenience_functions(self):
        """Test the convenience functions"""
        # Test parse_json function
        result = parse_json('{"test": true}')
        self.assertTrue(result.success)

        # Test stats functions
        stats = get_parser_stats()
        self.assertIsInstance(stats, dict)

        # Test reset function
        reset_parser_stats()
        stats = get_parser_stats()
        self.assertEqual(stats["total_parses"], 0)

    def test_context_parameter(self):
        """Test that context parameter is properly handled"""
        context = {"source": "unit_test", "attempt": 1}
        result = self.parser.parse('{"test": true}', context=context)
        self.assertTrue(result.success)

    def test_error_severity_classification(self):
        """Test proper error severity classification"""
        # Valid JSON should have LOW severity
        result = self.parser.parse('{"valid": true}')
        self.assertEqual(result.severity, ErrorSeverity.LOW)

        # Recoverable errors should have MEDIUM severity
        result = self.parser.parse('{missing: "quotes"}')
        self.assertTrue(result.success)
        self.assertEqual(result.severity, ErrorSeverity.MEDIUM)

        # Invalid input type should have CRITICAL severity
        result = self.parser.parse(123)
        self.assertFalse(result.success)
        self.assertEqual(result.severity, ErrorSeverity.CRITICAL)

    def test_max_recovery_attempts(self):
        """Test that max recovery attempts are respected"""
        # Create a parser with limited attempts
        limited_parser = RobustJSONParser(max_recovery_attempts=1)

        # Parse something that requires multiple recovery attempts
        result = limited_parser.parse("{missing: quotes, trailing: comma,}")

        # Should succeed but with limited attempts
        self.assertTrue(result.success)
        self.assertLessEqual(result.recovery_attempts, 1)

    def test_warnings_accumulation(self):
        """Test that warnings are properly accumulated"""
        # JSON with multiple issues
        test_json = '{name: "test", value: 123,}\x00'
        result = self.parser.parse(test_json)

        self.assertTrue(result.success)
        self.assertGreater(len(result.warnings), 1)
        self.assertIsInstance(result.warnings, list)


class TestPerformanceRequirements(unittest.TestCase):
    """Performance-specific tests for Robust JSON Parser"""

    def setUp(self):
        """Set up performance test fixtures"""
        self.parser = RobustJSONParser(enable_logging=False)

    def test_parsing_speed_requirements(self):
        """Test that parsing meets speed requirements"""
        # Small JSON should be very fast
        small_json = '{"test": true, "number": 123}'

        start_time = time.time()
        result = self.parser.parse(small_json)
        parse_time = time.time() - start_time

        self.assertTrue(result.success)
        self.assertLess(parse_time, 0.01)  # Should be under 10ms

    def test_memory_efficiency(self):
        """Test memory efficiency with large inputs"""
        # Create a large JSON string
        large_list = list(range(10000))
        large_json = json.dumps(large_list)

        result = self.parser.parse(large_json)
        self.assertTrue(result.success)
        self.assertEqual(len(result.data), 10000)

    def test_error_recovery_performance(self):
        """Test that error recovery doesn't significantly impact performance"""
        # JSON that requires recovery
        broken_json = '{name: "test", value: 123,}'

        start_time = time.time()
        result = self.parser.parse(broken_json)
        recovery_time = time.time() - start_time

        self.assertTrue(result.success)
        # Recovery should still be reasonably fast
        self.assertLess(recovery_time, 0.1)  # Should be under 100ms


class TestEdgeCases(unittest.TestCase):
    """Test edge cases and boundary conditions"""

    def setUp(self):
        """Set up edge case test fixtures"""
        self.parser = RobustJSONParser(enable_logging=False)

    def test_extremely_long_strings(self):
        """Test parsing of JSON with extremely long string values"""
        long_string = "a" * 10000
        test_json = f'{{"long_string": "{long_string}"}}'

        result = self.parser.parse(test_json)
        self.assertTrue(result.success)
        self.assertEqual(len(result.data["long_string"]), 10000)

    def test_unicode_edge_cases(self):
        """Test Unicode edge cases"""
        test_cases = [
            '{"zero_width": "\u200b"}',
            '{"surrogate": "\ud83d\ude00"}',  # Emoji surrogate pair
            '{"invalid_unicode": "\\uXYZ"}',  # Invalid Unicode sequence
        ]

        for test_json in test_cases:
            with self.subTest(test_json=test_json):
                result = self.parser.parse(test_json)
                # Should either succeed or fail gracefully
                self.assertIsInstance(result, ParseResult)

    def test_very_deep_nesting(self):
        """Test parsing of very deeply nested structures"""
        # Create deeply nested structure
        nested = "value"
        for _ in range(100):  # 100 levels deep
            nested = f'{{"level": {nested}}}'

        result = self.parser.parse(nested)
        # This might fail due to Python recursion limits, but shouldn't crash
        self.assertIsInstance(result, ParseResult)


def run_performance_benchmark():
    """Run a quick performance benchmark"""
    parser = RobustJSONParser(enable_logging=False)

    test_cases = [
        ("Valid JSON", '{"test": true, "number": 123}'),
        ("Missing Quotes", '{name: "test", value: 123}'),
        ("Trailing Comma", '{"test": true, "number": 123,}'),
        ("Mixed Errors", '{name: "test", value: 123,}'),
    ]

    print("\nPerformance Benchmark:")
    print("-" * 50)

    for name, test_json in test_cases:
        times = []
        for _ in range(100):  # Run 100 times
            start_time = time.time()
            parser.parse(test_json)
            times.append(time.time() - start_time)

        avg_time = sum(times) / len(times) * 1000  # Convert to ms
        success_rate = sum(1 for _ in range(100) if parser.parse(test_json).success) / 100

        print(f"{name:20}: {avg_time:6.2f}ms avg, {success_rate:5.1%} success rate")


if __name__ == "__main__":
    # Run unit tests
    print("Running Robust JSON Parser Unit Tests...")
    print("=" * 60)

    unittest.main(verbosity=2, exit=False)

    # Run performance benchmark
    run_performance_benchmark()
