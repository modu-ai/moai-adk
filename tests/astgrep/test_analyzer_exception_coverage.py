# Exception coverage tests for AST-Grep analyzer
"""Tests to achieve 100% coverage for exception handling paths.

This file targets specific exception handling branches that are hard to reach
with standard test inputs.
"""

from __future__ import annotations

from unittest.mock import Mock

from moai_adk.astgrep.analyzer import MoAIASTGrepAnalyzer


class TestParseSgMatchExceptionPaths:
    """Tests for exception paths in _parse_sg_match method.

    These tests target lines 283-284 which handle KeyError, TypeError, and AttributeError.
    """

    def test_parse_sg_match_dict_like_object_without_get(self) -> None:
        """Test _parse_sg_match with a dict-like object without get method."""

        analyzer = MoAIASTGrepAnalyzer()

        # Create a dict-like object that raises AttributeError when calling .get()
        class BrokenDict(dict):
            """A dict subclass that breaks .get() method."""

            def get(self, *args, **kwargs):  # type: ignore[override]
                raise AttributeError("get method broken")

        item = BrokenDict(
            {
                "ruleId": "test-rule",
                "severity": "error",
                "message": "Test message",
                "range": {
                    "start": {"line": 0, "column": 0},
                    "end": {"line": 1, "column": 0},
                },
            }
        )

        result = analyzer._parse_sg_match(item, "/test/file.py")

        # Should return None when AttributeError occurs
        assert result is None

    def test_parse_sg_match_range_get_raises_attribute_error(self) -> None:
        """Test _parse_sg_match when range.get() raises AttributeError."""

        analyzer = MoAIASTGrepAnalyzer()

        # Create a range dict-like object with broken .get()
        class BrokenRange(dict):
            """A dict subclass that breaks .get() method on specific key."""

            def get(self, key, *args):
                if key == "start":
                    raise AttributeError("start access broken")
                return super().get(key, *args)

        item = {
            "ruleId": "test-rule",
            "severity": "error",
            "message": "Test message",
            "range": BrokenRange(
                {
                    "start": {"line": 0, "column": 0},
                    "end": {"line": 1, "column": 0},
                }
            ),
        }

        result = analyzer._parse_sg_match(item, "/test/file.py")

        # Should return None when AttributeError occurs
        assert result is None

    def test_parse_sg_match_nested_get_raises_type_error(self) -> None:
        """Test _parse_sg_match when nested .get() raises TypeError."""

        analyzer = MoAIASTGrepAnalyzer()

        # Create a start object that returns None, causing TypeError
        # in the nested get() call: start.get("column", start.get("character", 0))
        # When start.get("column") returns None, start.get("character", 0) is called
        # But if start is not a dict, .get() doesn't exist
        class FakePosition:
            """A fake position object without get method."""

            line = 0
            column = 5

        item = {
            "ruleId": "test-rule",
            "severity": "error",
            "message": "Test message",
            "range": {
                "start": FakePosition(),  # Not a dict, will cause TypeError
                "end": {"line": 1, "column": 0},
            },
        }

        result = analyzer._parse_sg_match(item, "/test/file.py")

        # Should return None when TypeError occurs
        assert result is None

    def test_parse_sg_match_with_mock_exception_in_get(self) -> None:
        """Test _parse_sg_match when item.get() raises unexpected exception."""

        analyzer = MoAIASTGrepAnalyzer()

        # Create a mock that raises exception on .get()
        mock_item = Mock()
        mock_item.get.side_effect = RuntimeError("Unexpected error")

        result = analyzer._parse_sg_match(mock_item, "/test/file.py")

        # Should return None for any exception in try block
        # Note: This actually won't reach the try block because of the
        # isinstance check at line 242, so this test verifies that path
        assert result is None

    def test_parse_sg_match_valid_dict_with_all_fields(self) -> None:
        """Test _parse_sg_match with valid dict to verify normal path works."""

        analyzer = MoAIASTGrepAnalyzer()

        item = {
            "ruleId": "test-rule",
            "severity": "error",
            "message": "Test message",
            "range": {
                "start": {"line": 0, "column": 5},
                "end": {"line": 0, "column": 15},
            },
            "fix": "suggested_fix",
        }

        result = analyzer._parse_sg_match(item, "/test/file.py")

        # Should return valid ASTMatch
        assert result is not None
        assert result.rule_id == "test-rule"
        assert result.severity == "error"
        assert result.message == "Test message"
        assert result.file_path == "/test/file.py"
        assert result.range.start.line == 0
        assert result.range.start.character == 5
        assert result.range.end.line == 0
        assert result.range.end.character == 15
        assert result.suggested_fix == "suggested_fix"
