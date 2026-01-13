# AST-grep models edge case tests
"""Tests for edge cases and boundary conditions in data models.

Following TDD RED-GREEN-REFACTOR cycle.
These tests cover dataclass edge cases, boundary values, and type safety.
"""

from __future__ import annotations

from moai_adk.astgrep.models import (
    ASTMatch,
    ProjectScanResult,
    ReplaceResult,
    ScanConfig,
    ScanResult,
)
from moai_adk.lsp.models import Position, Range


class TestASTMatchEdgeCases:
    """Tests for ASTMatch edge cases."""

    def test_astmatch_with_zero_length_range(self) -> None:
        """Test ASTMatch with zero-length range (cursor position)."""
        match_range = Range(
            start=Position(line=5, character=10),
            end=Position(line=5, character=10),
        )

        match = ASTMatch(
            rule_id="cursor-rule",
            severity="info",
            message="Cursor position",
            file_path="/test.py",
            range=match_range,
            suggested_fix=None,
        )

        assert match.range.start.line == 5
        assert match.range.start.character == 10
        assert match.range.end.line == 5
        assert match.range.end.character == 10

    def test_astmatch_with_large_line_numbers(self) -> None:
        """Test ASTMatch with very large line numbers."""
        match_range = Range(
            start=Position(line=999999, character=0),
            end=Position(line=999999, character=100),
        )

        match = ASTMatch(
            rule_id="large-file-rule",
            severity="warning",
            message="Large file match",
            file_path="/large/file.py",
            range=match_range,
            suggested_fix=None,
        )

        assert match.range.start.line == 999999

    def test_astmatch_with_multiline_range(self) -> None:
        """Test ASTMatch with multiline range."""
        match_range = Range(
            start=Position(line=10, character=5),
            end=Position(line=15, character=10),
        )

        match = ASTMatch(
            rule_id="multiline-rule",
            severity="error",
            message="Multiline match",
            file_path="/test.py",
            range=match_range,
            suggested_fix=None,
        )

        assert match.range.start.line == 10
        assert match.range.end.line == 15

    def test_astmatch_with_empty_strings(self) -> None:
        """Test ASTMatch with empty string fields."""
        match_range = Range(
            start=Position(line=0, character=0),
            end=Position(line=0, character=0),
        )

        match = ASTMatch(
            rule_id="",
            severity="",
            message="",
            file_path="",
            range=match_range,
            suggested_fix="",
        )

        assert match.rule_id == ""
        assert match.severity == ""
        assert match.message == ""
        assert match.file_path == ""
        assert match.suggested_fix == ""

    def test_astmatch_with_unicode_file_path(self) -> None:
        """Test ASTMatch with Unicode characters in file path."""
        match_range = Range(
            start=Position(line=0, character=0),
            end=Position(line=0, character=10),
        )

        match = ASTMatch(
            rule_id="unicode-path",
            severity="warning",
            message="File with Unicode path",
            file_path="/path/to/文件_🎨.py",
            range=match_range,
            suggested_fix=None,
        )

        assert "文件" in match.file_path
        assert "🎨" in match.file_path

    def test_astmatch_with_all_severities(self) -> None:
        """Test ASTMatch with all severity values including edge cases."""
        match_range = Range(
            start=Position(line=0, character=0),
            end=Position(line=0, character=10),
        )

        # Standard severities
        for severity in ["error", "warning", "info", "hint"]:
            match = ASTMatch(
                rule_id=f"{severity}-rule",
                severity=severity,
                message="Test",
                file_path="/test.py",
                range=match_range,
                suggested_fix=None,
            )
            assert match.severity == severity

        # Case variations
        for severity in ["ERROR", "Warning", "Info", "HINT"]:
            match = ASTMatch(
                rule_id=f"{severity}-rule",
                severity=severity,
                message="Test",
                file_path="/test.py",
                range=match_range,
                suggested_fix=None,
            )
            assert match.severity == severity

    def test_astmatch_suggested_fix_with_unicode(self) -> None:
        """Test ASTMatch with Unicode in suggested fix."""
        match_range = Range(
            start=Position(line=0, character=0),
            end=Position(line=0, character=10),
        )

        match = ASTMatch(
            rule_id="unicode-fix",
            severity="warning",
            message="Fix with Unicode",
            file_path="/test.py",
            range=match_range,
            suggested_fix="logger.info('你好世界 🌍')",
        )

        assert "你好世界" in match.suggested_fix  # type: ignore[operator]
        assert "🌍" in match.suggested_fix  # type: ignore[operator]


class TestScanConfigEdgeCases:
    """Tests for ScanConfig edge cases."""

    def test_scanconfig_with_empty_patterns(self) -> None:
        """Test ScanConfig with empty include/exclude pattern lists."""
        config = ScanConfig(
            include_patterns=[],
            exclude_patterns=[],
        )

        assert config.include_patterns == []
        assert config.exclude_patterns == []

    def test_scanconfig_with_overlapping_patterns(self) -> None:
        """Test ScanConfig with overlapping include/exclude patterns."""
        config = ScanConfig(
            include_patterns=["*.py", "*.js"],
            exclude_patterns=["*.py", "test_*"],
        )

        # Both can coexist (behavior depends on implementation)
        assert "*.py" in config.include_patterns
        assert "*.py" in config.exclude_patterns

    def test_scanconfig_with_complex_patterns(self) -> None:
        """Test ScanConfig with complex glob patterns."""
        config = ScanConfig(
            include_patterns=[
                "test_*.py",
                "*_test.py",
                "**/test_*.py",
                "src/**/*.py",
            ],
            exclude_patterns=[
                "node_modules/**",
                "**/__pycache__/**",
                ".git/**",
                "**/*.min.js",
            ],
        )

        assert len(config.include_patterns) == 4
        assert len(config.exclude_patterns) == 4

    def test_scanconfig_with_unicode_patterns(self) -> None:
        """Test ScanConfig with Unicode in patterns."""
        config = ScanConfig(
            include_patterns=["*中文*.py", "*测试*.py"],
            exclude_patterns=["*旧的*"],
        )

        assert "中文" in config.include_patterns[0]
        assert "测试" in config.include_patterns[1]
        assert "旧的" in config.exclude_patterns[0]

    def test_scanconfig_with_special_characters(self) -> None:
        """Test ScanConfig with special characters in patterns."""
        config = ScanConfig(
            include_patterns=[
                "test-file.py",
                "file[1].py",
                "file?.py",
            ],
        )

        assert len(config.include_patterns) == 3


class TestScanResultEdgeCases:
    """Tests for ScanResult edge cases."""

    def test_scanresult_with_zero_scan_time(self) -> None:
        """Test ScanResult with zero scan time."""
        result = ScanResult(
            file_path="/test.py",
            matches=[],
            scan_time_ms=0.0,
            language="python",
        )

        assert result.scan_time_ms == 0.0

    def test_scanresult_with_very_fast_scan(self) -> None:
        """Test ScanResult with very fast scan time (< 1ms)."""
        result = ScanResult(
            file_path="/test.py",
            matches=[],
            scan_time_ms=0.5,
            language="python",
        )

        assert result.scan_time_ms == 0.5

    def test_scanresult_with_very_slow_scan(self) -> None:
        """Test ScanResult with very slow scan time (> 1000ms)."""
        result = ScanResult(
            file_path="/large/file.py",
            matches=[],
            scan_time_ms=5000.0,
            language="python",
        )

        assert result.scan_time_ms == 5000.0

    def test_scanresult_with_negative_scan_time(self) -> None:
        """Test ScanResult accepts negative scan time (should not happen in practice)."""
        result = ScanResult(
            file_path="/test.py",
            matches=[],
            scan_time_ms=-1.0,
            language="python",
        )

        assert result.scan_time_ms == -1.0

    def test_scanresult_with_unicode_file_path(self) -> None:
        """Test ScanResult with Unicode file path."""
        result = ScanResult(
            file_path="/路径/文件.py",
            matches=[],
            scan_time_ms=10.0,
            language="python",
        )

        assert "路径" in result.file_path
        assert "文件" in result.file_path

    def test_scanresult_matches_immutability(self) -> None:
        """Test that ScanResult matches list can be modified."""
        from moai_adk.astgrep.models import ASTMatch

        match_range = Range(
            start=Position(line=0, character=0),
            end=Position(line=0, character=10),
        )

        result = ScanResult(
            file_path="/test.py",
            matches=[],
            scan_time_ms=10.0,
            language="python",
        )

        # Modify matches list
        match = ASTMatch(
            rule_id="test",
            severity="warning",
            message="Test",
            file_path="/test.py",
            range=match_range,
            suggested_fix=None,
        )
        result.matches.append(match)

        assert len(result.matches) == 1


class TestProjectScanResultEdgeCases:
    """Tests for ProjectScanResult edge cases."""

    def test_projectscanresult_with_zero_files(self) -> None:
        """Test ProjectScanResult with zero files scanned."""
        result = ProjectScanResult(
            project_path="/empty",
            files_scanned=0,
            total_matches=0,
            results_by_file={},
            summary_by_severity={"error": 0, "warning": 0, "info": 0, "hint": 0},
            scan_time_ms=0.0,
        )

        assert result.files_scanned == 0
        assert result.total_matches == 0

    def test_projectscanresult_with_large_file_counts(self) -> None:
        """Test ProjectScanResult with very large file counts."""
        result = ProjectScanResult(
            project_path="/large",
            files_scanned=1000000,
            total_matches=5000000,
            results_by_file={},
            summary_by_severity={"error": 1000, "warning": 50000, "info": 100000, "hint": 100000},
            scan_time_ms=300000.0,
        )

        assert result.files_scanned == 1000000
        assert result.total_matches == 5000000

    def test_projectscanresult_with_custom_severity_counts(self) -> None:
        """Test ProjectScanResult with custom severity keys."""
        result = ProjectScanResult(
            project_path="/test",
            files_scanned=10,
            total_matches=5,
            results_by_file={},
            summary_by_severity={
                "critical": 1,
                "error": 2,
                "warning": 1,
                "info": 1,
                "hint": 0,
            },
            scan_time_ms=100.0,
        )

        assert result.summary_by_severity["critical"] == 1
        assert result.summary_by_severity["error"] == 2

    def test_projectscanresult_with_unicode_project_path(self) -> None:
        """Test ProjectScanResult with Unicode project path."""
        result = ProjectScanResult(
            project_path="/路径/项目",
            files_scanned=5,
            total_matches=3,
            results_by_file={},
            summary_by_severity={"error": 1, "warning": 2, "info": 0, "hint": 0},
            scan_time_ms=50.0,
        )

        assert "路径" in result.project_path
        assert "项目" in result.project_path

    def test_projectscanresult_results_by_file_mutability(self) -> None:
        """Test that results_by_file dict can be modified."""
        from moai_adk.astgrep.models import ASTMatch, ScanResult

        match_range = Range(
            start=Position(line=0, character=0),
            end=Position(line=0, character=10),
        )

        scan_result = ScanResult(
            file_path="/test.py",
            matches=[
                ASTMatch(
                    rule_id="test",
                    severity="warning",
                    message="Test",
                    file_path="/test.py",
                    range=match_range,
                    suggested_fix=None,
                )
            ],
            scan_time_ms=10.0,
            language="python",
        )

        result = ProjectScanResult(
            project_path="/test",
            files_scanned=1,
            total_matches=1,
            results_by_file={},
            summary_by_severity={"error": 0, "warning": 0, "info": 0, "hint": 0},
            scan_time_ms=10.0,
        )

        # Modify results_by_file
        result.results_by_file["/test.py"] = scan_result

        assert len(result.results_by_file) == 1


class TestReplaceResultEdgeCases:
    """Tests for ReplaceResult edge cases."""

    def test_replaceresult_with_empty_changes(self) -> None:
        """Test ReplaceResult with no changes."""
        result = ReplaceResult(
            pattern="test_pattern",
            replacement="test_replacement",
            matches_found=0,
            files_modified=0,
            changes=[],
            dry_run=True,
        )

        assert result.matches_found == 0
        assert result.files_modified == 0
        assert result.changes == []

    def test_replaceresult_with_large_match_count(self) -> None:
        """Test ReplaceResult with very large match count."""
        result = ReplaceResult(
            pattern="old_pattern",
            replacement="new_pattern",
            matches_found=1000000,
            files_modified=50000,
            changes=[],
            dry_run=False,
        )

        assert result.matches_found == 1000000
        assert result.files_modified == 50000

    def test_replaceresult_with_complex_changes(self) -> None:
        """Test ReplaceResult with complex change descriptions."""
        changes = [
            {
                "file_path": "/path/to/文件_🎨.py",
                "old_code": "print('旧代码')",
                "new_code": "logger.info('新代码')",
                "range": {
                    "start": {"line": 10, "character": 0},
                    "end": {"line": 10, "character": 20},
                },
            },
            {
                "file_path": "/another/file.py",
                "old_code": "old_func()",
                "new_code": "new_func()",
                "range": {
                    "start": {"line": 20, "character": 5},
                    "end": {"line": 25, "character": 15},
                },
            },
        ]

        result = ReplaceResult(
            pattern="old_func",
            replacement="new_func",
            matches_found=2,
            files_modified=2,
            changes=changes,
            dry_run=True,
        )

        assert len(result.changes) == 2
        assert "文件" in result.changes[0]["file_path"]
        assert "🎨" in result.changes[0]["file_path"]

    def test_replaceresult_with_unicode_patterns(self) -> None:
        """Test ReplaceResult with Unicode in pattern and replacement."""
        result = ReplaceResult(
            pattern="打印($MSG)",
            replacement="日志器.信息($MSG)",
            matches_found=5,
            files_modified=2,
            changes=[],
            dry_run=True,
        )

        assert "打印" in result.pattern
        assert "日志器" in result.replacement

    def test_replaceresult_both_dry_run_modes(self) -> None:
        """Test ReplaceResult with both dry_run values."""
        for dry_run in [True, False]:
            result = ReplaceResult(
                pattern="test",
                replacement="fixed",
                matches_found=1,
                files_modified=1,
                changes=[],
                dry_run=dry_run,
            )

            assert result.dry_run == dry_run

    def test_replaceresult_changes_mutability(self) -> None:
        """Test that changes list can be modified."""
        result = ReplaceResult(
            pattern="test",
            replacement="fixed",
            matches_found=0,
            files_modified=0,
            changes=[],
            dry_run=True,
        )

        # Modify changes list
        result.changes.append(
            {
                "file_path": "/test.py",
                "old_code": "old",
                "new_code": "new",
                "range": {"start": {"line": 0, "character": 0}, "end": {"line": 0, "character": 3}},
            }
        )

        assert len(result.changes) == 1


class TestDataclassImmutability:
    """Tests for dataclass field mutability."""

    def test_models_are_standard_dataclasses(self) -> None:
        """Test that all models are standard mutable dataclasses."""
        match_range = Range(
            start=Position(line=0, character=0),
            end=Position(line=0, character=10),
        )

        # ASTMatch
        match = ASTMatch(
            rule_id="test",
            severity="warning",
            message="Test",
            file_path="/test.py",
            range=match_range,
            suggested_fix=None,
        )
        match.severity = "error"
        assert match.severity == "error"

        # ScanConfig
        config = ScanConfig()
        config.security_scan = False
        assert config.security_scan is False

        # ScanResult
        result = ScanResult(
            file_path="/test.py",
            matches=[],
            scan_time_ms=10.0,
            language="python",
        )
        result.language = "javascript"
        assert result.language == "javascript"

    def test_models_support_equality(self) -> None:
        """Test that models support equality comparison."""
        match_range1 = Range(
            start=Position(line=0, character=0),
            end=Position(line=0, character=10),
        )

        match_range2 = Range(
            start=Position(line=0, character=0),
            end=Position(line=0, character=10),
        )

        match1 = ASTMatch(
            rule_id="test",
            severity="warning",
            message="Test",
            file_path="/test.py",
            range=match_range1,
            suggested_fix=None,
        )

        match2 = ASTMatch(
            rule_id="test",
            severity="warning",
            message="Test",
            file_path="/test.py",
            range=match_range2,
            suggested_fix=None,
        )

        assert match1 == match2

    def test_models_support_repr(self) -> None:
        """Test that models have string representations."""
        match_range = Range(
            start=Position(line=0, character=0),
            end=Position(line=0, character=10),
        )

        match = ASTMatch(
            rule_id="test",
            severity="warning",
            message="Test",
            file_path="/test.py",
            range=match_range,
            suggested_fix=None,
        )

        repr_str = repr(match)
        assert "ASTMatch" in repr_str
        assert "test" in repr_str
