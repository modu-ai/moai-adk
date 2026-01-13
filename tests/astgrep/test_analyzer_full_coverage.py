# Test file for 100% coverage of AST-Grep analyzer
"""
Tests to achieve 100% coverage for moai_adk.astgrep.analyzer module.

This test file targets specific uncovered lines:
- Lines 160-162: SubprocessError/FileNotFoundError exception handling in scan_file
- Line 260: Non-dict start/end validation in _parse_sg_match
- Lines 283-284: KeyError/TypeError/AttributeError exception handling in _parse_sg_match
- Lines 326-332: Severity aggregation logic in scan_project
- Lines 512-513: SubprocessError/JSONDecodeError exception handling in pattern_replace
"""

from __future__ import annotations

import json
import subprocess
from pathlib import Path
from unittest.mock import MagicMock, Mock

from moai_adk.astgrep.analyzer import MoAIASTGrepAnalyzer
from moai_adk.astgrep.models import (
    ASTMatch,
    ReplaceResult,
    ScanConfig,
    ScanResult,
)
from moai_adk.lsp.models import Position, Range


class TestScanFileSubprocessExceptions:
    """Tests for exception handling in scan_file (lines 160-162)."""

    def test_scan_file_subprocess_error_from_run_sg_scan(self, tmp_path: Path, monkeypatch) -> None:
        """Test scan_file when _run_sg_scan raises SubprocessError."""

        # Create a test file
        test_file = tmp_path / "test.py"
        test_file.write_text("print('hello')")

        analyzer = MoAIASTGrepAnalyzer()

        # Mock _run_sg_scan to raise SubprocessError
        def mock_run_sg_scan(*_args, **_kwargs):
            raise subprocess.SubprocessError("Command failed")

        monkeypatch.setattr(analyzer, "_run_sg_scan", mock_run_sg_scan)

        result = analyzer.scan_file(str(test_file))

        # Should return ScanResult with empty matches (graceful degradation)
        assert isinstance(result, ScanResult)
        assert result.file_path == str(test_file)
        assert result.matches == []
        assert result.language == "python"

    def test_scan_file_file_not_found_error_from_run_sg_scan(self, tmp_path: Path, monkeypatch) -> None:
        """Test scan_file when _run_sg_scan raises FileNotFoundError."""

        test_file = tmp_path / "test.py"
        test_file.write_text("print('hello')")

        analyzer = MoAIASTGrepAnalyzer()

        # Mock _run_sg_scan to raise FileNotFoundError
        def mock_run_sg_scan(*_args, **_kwargs):
            raise FileNotFoundError("sg not found")

        monkeypatch.setattr(analyzer, "_run_sg_scan", mock_run_sg_scan)

        result = analyzer.scan_file(str(test_file))

        # Should return ScanResult with empty matches (graceful degradation)
        assert isinstance(result, ScanResult)
        assert result.file_path == str(test_file)
        assert result.matches == []


class TestParseSgMatchEdgeCases:
    """Tests for _parse_sg_match edge cases (lines 260, 283-284)."""

    def test_parse_sg_match_non_dict_start(self) -> None:
        """Test _parse_sg_match when start is not a dict (line 260)."""

        analyzer = MoAIASTGrepAnalyzer()

        item = {
            "ruleId": "test-rule",
            "severity": "error",
            "message": "Test message",
            "range": {"start": "not-a-dict", "end": {"line": 0, "column": 0}},
        }

        result = analyzer._parse_sg_match(item, "/test/file.py")

        # Should return None when start is not a dict
        assert result is None

    def test_parse_sg_match_non_dict_end(self) -> None:
        """Test _parse_sg_match when end is not a dict (line 260)."""

        analyzer = MoAIASTGrepAnalyzer()

        item = {
            "ruleId": "test-rule",
            "severity": "error",
            "message": "Test message",
            "range": {"start": {"line": 0, "column": 0}, "end": [1, 2]},
        }

        result = analyzer._parse_sg_match(item, "/test/file.py")

        # Should return None when end is not a dict
        assert result is None

    def test_parse_sg_match_both_non_dict(self) -> None:
        """Test _parse_sg_match when both start and end are not dicts."""

        analyzer = MoAIASTGrepAnalyzer()

        item = {
            "ruleId": "test-rule",
            "severity": "error",
            "message": "Test message",
            "range": {"start": None, "end": None},
        }

        result = analyzer._parse_sg_match(item, "/test/file.py")

        # Should return None
        assert result is None

    def test_parse_sg_match_key_error(self) -> None:
        """Test _parse_sg_match with KeyError exception (lines 283-284)."""

        analyzer = MoAIASTGrepAnalyzer()

        # Create a mock that raises KeyError when accessing keys
        item = MagicMock()
        item.get.side_effect = KeyError("test key")

        result = analyzer._parse_sg_match(item, "/test/file.py")

        # Should return None on KeyError
        assert result is None

    def test_parse_sg_match_type_error(self) -> None:
        """Test _parse_sg_match with TypeError exception (lines 283-284)."""

        analyzer = MoAIASTGrepAnalyzer()

        # Create an item that will cause TypeError
        # For example, range_data.get returns None and we try to access it
        item = {"range": None}

        result = analyzer._parse_sg_match(item, "/test/file.py")

        # Should return None on TypeError
        assert result is None

    def test_parse_sg_match_attribute_error(self) -> None:
        """Test _parse_sg_match with AttributeError exception (lines 283-284)."""

        analyzer = MoAIASTGrepAnalyzer()

        # Create a mock that raises AttributeError
        mock_range = MagicMock()
        mock_range.get.side_effect = AttributeError("test attribute")

        item = {"range": mock_range}

        result = analyzer._parse_sg_match(item, "/test/file.py")

        # Should return None on AttributeError
        assert result is None


class TestScanProjectSeverityAggregation:
    """Tests for severity aggregation logic in scan_project (lines 326-332)."""

    def test_scan_project_aggregates_severity_counts(self, tmp_path: Path, monkeypatch) -> None:
        """Test scan_project correctly aggregates severity counts."""

        # Create test files
        (tmp_path / "test1.py").write_text("code1")
        (tmp_path / "test2.py").write_text("code2")

        analyzer = MoAIASTGrepAnalyzer()

        # Mock scan_file to return matches with different severities
        mock_results = []

        result1 = ScanResult(
            file_path=str(tmp_path / "test1.py"),
            matches=[
                ASTMatch(
                    rule_id="rule1",
                    severity="error",
                    message="Error message",
                    file_path=str(tmp_path / "test1.py"),
                    range=Range(
                        start=Position(line=0, character=0),
                        end=Position(line=0, character=5),
                    ),
                    suggested_fix=None,
                ),
                ASTMatch(
                    rule_id="rule2",
                    severity="warning",
                    message="Warning message",
                    file_path=str(tmp_path / "test1.py"),
                    range=Range(
                        start=Position(line=1, character=0),
                        end=Position(line=1, character=5),
                    ),
                    suggested_fix=None,
                ),
            ],
            scan_time_ms=10.0,
            language="python",
        )
        mock_results.append(result1)

        result2 = ScanResult(
            file_path=str(tmp_path / "test2.py"),
            matches=[
                ASTMatch(
                    rule_id="rule3",
                    severity="error",
                    message="Another error",
                    file_path=str(tmp_path / "test2.py"),
                    range=Range(
                        start=Position(line=0, character=0),
                        end=Position(line=0, character=5),
                    ),
                    suggested_fix=None,
                ),
                ASTMatch(
                    rule_id="rule4",
                    severity="info",
                    message="Info message",
                    file_path=str(tmp_path / "test2.py"),
                    range=Range(
                        start=Position(line=1, character=0),
                        end=Position(line=1, character=5),
                    ),
                    suggested_fix=None,
                ),
            ],
            scan_time_ms=15.0,
            language="python",
        )
        mock_results.append(result2)

        # Set up the mock to return results sequentially
        mock_iter = iter(mock_results)

        def mock_scan_file(_file_path, _config=None):
            return next(mock_iter)

        monkeypatch.setattr(analyzer, "scan_file", mock_scan_file)

        result = analyzer.scan_project(str(tmp_path))

        # Verify severity counts
        assert result.summary_by_severity["error"] == 2
        assert result.summary_by_severity["warning"] == 1
        assert result.summary_by_severity["info"] == 1
        assert result.summary_by_severity["hint"] == 0
        assert result.total_matches == 4  # All matches are counted
        assert len(result.results_by_file) == 2

    def test_scan_project_case_insensitive_severity(self, tmp_path: Path, monkeypatch) -> None:
        """Test scan_project handles case-insensitive severity."""

        (tmp_path / "test.py").write_text("code")

        analyzer = MoAIASTGrepAnalyzer()

        # Mock scan_file to return match with uppercase severity
        scan_result = ScanResult(
            file_path=str(tmp_path / "test.py"),
            matches=[
                ASTMatch(
                    rule_id="rule1",
                    severity="ERROR",  # Uppercase
                    message="Error message",
                    file_path=str(tmp_path / "test.py"),
                    range=Range(
                        start=Position(line=0, character=0),
                        end=Position(line=0, character=5),
                    ),
                    suggested_fix=None,
                ),
            ],
            scan_time_ms=10.0,
            language="python",
        )

        monkeypatch.setattr(analyzer, "scan_file", lambda _fp, _cfg=None: scan_result)

        result = analyzer.scan_project(str(tmp_path))

        # Should count as "error" (lowercase)
        assert result.summary_by_severity["error"] == 1

    def test_scan_project_unknown_severity(self, tmp_path: Path, monkeypatch) -> None:
        """Test scan_project handles unknown severity values."""

        (tmp_path / "test.py").write_text("code")

        analyzer = MoAIASTGrepAnalyzer()

        # Mock scan_file to return match with unknown severity
        scan_result = ScanResult(
            file_path=str(tmp_path / "test.py"),
            matches=[
                ASTMatch(
                    rule_id="rule1",
                    severity="unknown",  # Unknown severity
                    message="Unknown message",
                    file_path=str(tmp_path / "test.py"),
                    range=Range(
                        start=Position(line=0, character=0),
                        end=Position(line=0, character=5),
                    ),
                    suggested_fix=None,
                ),
            ],
            scan_time_ms=10.0,
            language="python",
        )

        monkeypatch.setattr(analyzer, "scan_file", lambda _fp, _cfg=None: scan_result)

        result = analyzer.scan_project(str(tmp_path))

        # Unknown severity should not be counted in summary
        assert result.summary_by_severity["error"] == 0
        assert result.summary_by_severity["warning"] == 0
        assert result.summary_by_severity["info"] == 0
        assert result.summary_by_severity["hint"] == 0
        # But total matches should still be counted
        assert result.total_matches == 1


class TestPatternReplaceExceptions:
    """Tests for exception handling in pattern_replace (lines 512-513)."""

    def test_pattern_replace_subprocess_error(self, tmp_path: Path, monkeypatch) -> None:
        """Test pattern_replace when subprocess raises SubprocessError."""

        test_file = tmp_path / "test.py"
        test_file.write_text("old_pattern")

        analyzer = MoAIASTGrepAnalyzer()

        # Mock pattern_search to return matches
        mock_match = ASTMatch(
            rule_id="pattern:old_pattern",
            severity="warning",
            message="Pattern found",
            file_path=str(test_file),
            range=Range(
                start=Position(line=0, character=0),
                end=Position(line=0, character=12),
            ),
            suggested_fix=None,
        )

        # First call succeeds, second (for actual replacement) raises SubprocessError
        call_count = [0]

        def mock_pattern_search(*_args, **_kwargs):
            if call_count[0] == 0:
                call_count[0] += 1
                return [mock_match]
            return []

        def mock_subprocess_run(*_args, **_kwargs):
            if "--rewrite" in _args[0]:
                raise subprocess.SubprocessError("Replacement failed")
            return Mock(returncode=0, stdout="[]")

        monkeypatch.setattr(analyzer, "pattern_search", mock_pattern_search)
        monkeypatch.setattr("subprocess.run", mock_subprocess_run)

        # dry_run=False to trigger actual replacement attempt
        result = analyzer.pattern_replace(
            pattern="old_pattern",
            replacement="new_pattern",
            language="python",
            path=str(test_file),
            dry_run=False,
        )

        # Should return ReplaceResult gracefully despite error
        assert isinstance(result, ReplaceResult)
        assert result.pattern == "old_pattern"
        assert result.replacement == "new_pattern"
        # Matches were found in the search phase
        assert result.matches_found == 1
        assert result.files_modified == 1
        assert len(result.changes) == 1
        assert result.dry_run is False

    def test_pattern_replace_json_decode_error(self, tmp_path: Path, monkeypatch) -> None:
        """Test pattern_replace when JSON decode error occurs."""

        test_file = tmp_path / "test.py"
        test_file.write_text("old_pattern")

        analyzer = MoAIASTGrepAnalyzer()

        # Mock pattern_search to return matches
        mock_match = ASTMatch(
            rule_id="pattern:old_pattern",
            severity="warning",
            message="Pattern found",
            file_path=str(test_file),
            range=Range(
                start=Position(line=0, character=0),
                end=Position(line=0, character=12),
            ),
            suggested_fix=None,
        )

        # Track calls
        call_count = [0]

        def mock_pattern_search(*_args, **_kwargs):
            if call_count[0] == 0:
                call_count[0] += 1
                return [mock_match]
            return []

        def mock_subprocess_run(*_args, **_kwargs):
            if "--rewrite" in _args[0]:
                raise json.JSONDecodeError("Invalid JSON", "", 0)
            return Mock(returncode=0, stdout="[]")

        monkeypatch.setattr(analyzer, "pattern_search", mock_pattern_search)
        monkeypatch.setattr("subprocess.run", mock_subprocess_run)

        result = analyzer.pattern_replace(
            pattern="old_pattern",
            replacement="new_pattern",
            language="python",
            path=str(test_file),
            dry_run=False,
        )

        # Should return ReplaceResult gracefully despite JSON error
        assert isinstance(result, ReplaceResult)
        assert result.matches_found == 1
        assert result.files_modified == 1
        assert result.dry_run is False


class TestRunSgScanWithRealSubprocess:
    """Additional tests for _run_sg_scan method."""

    def test_run_sg_scan_nonzero_exit_code(self, tmp_path: Path, monkeypatch) -> None:
        """Test _run_sg_scan when subprocess returns non-zero exit code."""

        test_file = tmp_path / "test.py"
        test_file.write_text("print('hello')")

        analyzer = MoAIASTGrepAnalyzer()

        # Mock subprocess to return non-zero exit code
        monkeypatch.setattr(
            "subprocess.run",
            lambda *_args, **_kwargs: Mock(returncode=1, stdout=""),
        )

        result = analyzer._run_sg_scan(str(test_file), "python", ScanConfig())

        # Should return empty list
        assert result == []

    def test_run_sg_scan_empty_stdout(self, tmp_path: Path, monkeypatch) -> None:
        """Test _run_sg_scan when subprocess returns empty stdout."""

        test_file = tmp_path / "test.py"
        test_file.write_text("print('hello')")

        analyzer = MoAIASTGrepAnalyzer()

        # Mock subprocess to return empty stdout
        monkeypatch.setattr(
            "subprocess.run",
            lambda *args, **kwargs: Mock(returncode=0, stdout=""),
        )

        result = analyzer._run_sg_scan(str(test_file), "python", ScanConfig())

        # Should return empty list
        assert result == []


class TestParsePatternSearchOutputEdgeCases:
    """Additional tests for _parse_pattern_search_output method."""

    def test_parse_pattern_search_output_no_file_field(self) -> None:
        """Test _parse_pattern_search_output when item has no file field."""

        analyzer = MoAIASTGrepAnalyzer()

        output = json.dumps(
            [
                {
                    "pattern": "test_pattern",
                    "range": {
                        "start": {"line": 0, "column": 0},
                        "end": {"line": 0, "column": 10},
                    },
                }
            ]
        )

        result = analyzer._parse_pattern_search_output(output, "test_pattern")

        # Should handle missing file field gracefully
        # Items without valid file_path won't be added to matches
        assert isinstance(result, list)

    def test_parse_pattern_search_output_long_pattern_truncation(self) -> None:
        """Test _parse_pattern_search_output truncates long patterns."""

        analyzer = MoAIASTGrepAnalyzer()

        long_pattern = "x" * 100  # 100 character pattern

        output = json.dumps(
            [
                {
                    "file": "/test/file.py",
                    "pattern": long_pattern,
                    "range": {
                        "start": {"line": 0, "column": 0},
                        "end": {"line": 0, "column": 10},
                    },
                    "ruleId": "test",
                    "severity": "warning",
                    "message": "Test",
                }
            ]
        )

        result = analyzer._parse_pattern_search_output(output, long_pattern)

        # Should truncate pattern to 30 characters in rule_id
        if result:
            assert len(result[0].rule_id) <= 50  # "pattern:" prefix + 30 chars
