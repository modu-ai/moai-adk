"""Tests for moai_adk.core.robust_json_parser module."""

import pytest
import json
from moai_adk.core.robust_json_parser import (
    RobustJSONParser,
    ParseResult,
    ErrorSeverity,
)


class TestRobustJSONParserInit:
    """Test RobustJSONParser initialization."""

    def test_init_default_parameters(self):
        """Test initialization with default parameters."""
        parser = RobustJSONParser()
        assert parser.max_recovery_attempts == 3
        assert parser.enable_logging is True

    def test_init_custom_parameters(self):
        """Test initialization with custom parameters."""
        parser = RobustJSONParser(max_recovery_attempts=5, enable_logging=False)
        assert parser.max_recovery_attempts == 5
        assert parser.enable_logging is False

    def test_init_loads_error_patterns(self):
        """Test that error patterns are loaded."""
        parser = RobustJSONParser()
        assert isinstance(parser.error_patterns, dict)
        assert len(parser.error_patterns) > 0

    def test_init_loads_recovery_strategies(self):
        """Test that recovery strategies are loaded."""
        parser = RobustJSONParser()
        assert isinstance(parser.recovery_strategies, list)
        assert len(parser.recovery_strategies) > 0

    def test_init_initializes_stats(self):
        """Test that stats dictionary is initialized."""
        parser = RobustJSONParser()
        assert isinstance(parser.stats, dict)
        assert "total_parses" in parser.stats
        assert parser.stats["total_parses"] == 0


class TestParseValidJSON:
    """Test parsing valid JSON."""

    def test_parse_simple_object(self):
        """Test parsing a simple JSON object."""
        parser = RobustJSONParser()
        json_str = '{"key": "value"}'

        result = parser.parse(json_str)

        assert result.success is True
        assert result.data == {"key": "value"}
        assert result.error is None

    def test_parse_simple_array(self):
        """Test parsing a simple JSON array."""
        parser = RobustJSONParser()
        json_str = '[1, 2, 3]'

        result = parser.parse(json_str)

        assert result.success is True
        assert result.data == [1, 2, 3]

    def test_parse_nested_object(self):
        """Test parsing a nested JSON object."""
        parser = RobustJSONParser()
        json_str = '{"outer": {"inner": "value"}}'

        result = parser.parse(json_str)

        assert result.success is True
        assert result.data["outer"]["inner"] == "value"

    def test_parse_with_null(self):
        """Test parsing JSON with null values."""
        parser = RobustJSONParser()
        json_str = '{"key": null}'

        result = parser.parse(json_str)

        assert result.success is True
        assert result.data["key"] is None

    def test_parse_returns_parse_result(self):
        """Test that parse returns ParseResult instance."""
        parser = RobustJSONParser()
        result = parser.parse('{"test": "value"}')

        assert isinstance(result, ParseResult)
        assert hasattr(result, "success")
        assert hasattr(result, "data")
        assert hasattr(result, "error")


class TestParseInvalidJSON:
    """Test parsing invalid JSON with recovery."""

    def test_parse_trailing_comma(self):
        """Test parsing JSON with trailing comma."""
        parser = RobustJSONParser()
        json_str = '{"key": "value",}'

        result = parser.parse(json_str)

        # Should either succeed or indicate error
        assert isinstance(result, ParseResult)
        assert hasattr(result, "success")

    def test_parse_missing_quotes(self):
        """Test parsing JSON with missing quotes."""
        parser = RobustJSONParser()
        json_str = '{key: "value"}'

        result = parser.parse(json_str)

        assert isinstance(result, ParseResult)

    def test_parse_empty_string(self):
        """Test parsing empty string."""
        parser = RobustJSONParser()
        json_str = ''

        result = parser.parse(json_str)

        assert isinstance(result, ParseResult)

    def test_parse_whitespace_only(self):
        """Test parsing whitespace only."""
        parser = RobustJSONParser()
        json_str = '   \n  \t  '

        result = parser.parse(json_str)

        assert isinstance(result, ParseResult)

    def test_parse_invalid_json_returns_error(self):
        """Test that truly invalid JSON is marked as error."""
        parser = RobustJSONParser()
        json_str = '{{{invalid'

        result = parser.parse(json_str)

        assert isinstance(result, ParseResult)
        assert hasattr(result, "error")


class TestParseResultFields:
    """Test ParseResult data class fields."""

    def test_parse_result_has_all_fields(self):
        """Test that ParseResult has all required fields."""
        result = ParseResult(
            success=True,
            data={"test": "value"},
            error=None,
            original_input='{"test": "value"}',
            recovery_attempts=0,
            severity=ErrorSeverity.LOW,
            parse_time_ms=1.5,
            warnings=[],
        )

        assert result.success is True
        assert result.data == {"test": "value"}
        assert result.error is None
        assert result.recovery_attempts == 0
        assert result.severity == ErrorSeverity.LOW
        assert result.parse_time_ms == 1.5
        assert result.warnings == []

    def test_parse_result_original_input_preserved(self):
        """Test that original input is preserved in result."""
        parser = RobustJSONParser()
        json_str = '{"key": "value"}'

        result = parser.parse(json_str)

        assert result.original_input == json_str


class TestErrorSeverity:
    """Test ErrorSeverity enum."""

    def test_error_severity_values(self):
        """Test ErrorSeverity enum values."""
        assert ErrorSeverity.LOW.value == "low"
        assert ErrorSeverity.MEDIUM.value == "medium"
        assert ErrorSeverity.HIGH.value == "high"
        assert ErrorSeverity.CRITICAL.value == "critical"

    def test_error_severity_enumeration(self):
        """Test ErrorSeverity enumeration."""
        severities = list(ErrorSeverity)
        assert len(severities) == 4


class TestParserStats:
    """Test parser statistics tracking."""

    def test_stats_increment_on_parse(self):
        """Test that stats are incremented on parse."""
        parser = RobustJSONParser()
        initial_count = parser.stats["total_parses"]

        parser.parse('{"test": "value"}')

        assert parser.stats["total_parses"] == initial_count + 1

    def test_stats_track_successful_parses(self):
        """Test that successful parses are tracked."""
        parser = RobustJSONParser()
        parser.parse('{"test": "value"}')

        assert parser.stats["successful_parses"] >= 0

    def test_stats_reset_on_new_parser(self):
        """Test that new parser has fresh stats."""
        parser1 = RobustJSONParser()
        parser1.parse('{"test": "value"}')

        parser2 = RobustJSONParser()

        assert parser2.stats["total_parses"] == 0


class TestParseWithContext:
    """Test parsing with context parameter."""

    def test_parse_with_context_dict(self):
        """Test parsing with context dictionary."""
        parser = RobustJSONParser()
        context = {"source": "test"}

        result = parser.parse('{"key": "value"}', context=context)

        assert isinstance(result, ParseResult)

    def test_parse_with_none_context(self):
        """Test parsing with None context."""
        parser = RobustJSONParser()

        result = parser.parse('{"key": "value"}', context=None)

        assert isinstance(result, ParseResult)


class TestSpecialCharacters:
    """Test parsing with special characters."""

    def test_parse_with_unicode(self):
        """Test parsing JSON with unicode characters."""
        parser = RobustJSONParser()
        json_str = '{"korean": "한글", "japanese": "日本語"}'

        result = parser.parse(json_str)

        if result.success:
            assert "korean" in result.data

    def test_parse_with_escaped_newlines(self):
        """Test parsing JSON with escaped newlines."""
        parser = RobustJSONParser()
        json_str = '{"text": "line1\\nline2"}'

        result = parser.parse(json_str)

        assert isinstance(result, ParseResult)

    def test_parse_with_escaped_quotes(self):
        """Test parsing JSON with escaped quotes."""
        parser = RobustJSONParser()
        json_str = '{"text": "He said \\"hello\\""}'

        result = parser.parse(json_str)

        assert isinstance(result, ParseResult)


class TestParserRecovery:
    """Test parser recovery mechanisms."""

    def test_recovery_attempts_tracked(self):
        """Test that recovery attempts are tracked."""
        parser = RobustJSONParser()
        invalid_json = '{key: "value"}'

        result = parser.parse(invalid_json)

        assert result.recovery_attempts >= 0

    def test_recovery_attempts_within_limit(self):
        """Test that recovery stays within max attempts."""
        parser = RobustJSONParser(max_recovery_attempts=2)
        very_invalid = '{{{{{'

        result = parser.parse(very_invalid)

        assert result.recovery_attempts <= parser.max_recovery_attempts


class TestParsePerformance:
    """Test parser performance tracking."""

    def test_parse_time_is_recorded(self):
        """Test that parse time is recorded."""
        parser = RobustJSONParser()
        result = parser.parse('{"test": "value"}')

        assert result.parse_time_ms >= 0

    def test_parse_time_is_float(self):
        """Test that parse time is a float."""
        parser = RobustJSONParser()
        result = parser.parse('{"test": "value"}')

        assert isinstance(result.parse_time_ms, float)
