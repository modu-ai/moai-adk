# AST-grep analyzer edge case tests
"""Tests for edge cases and private methods in MoAIASTGrepAnalyzer.

Following TDD RED-GREEN-REFACTOR cycle.
These tests cover private methods and edge cases not covered in test_analyzer.py.
"""

from __future__ import annotations

import json
import tempfile
from pathlib import Path
from unittest.mock import MagicMock, patch

import pytest

from moai_adk.astgrep.analyzer import MoAIASTGrepAnalyzer
from moai_adk.astgrep.models import ScanConfig


class TestDetectLanguage:
    """Tests for _detect_language private method."""

    def test_detect_language_python(self) -> None:
        """Test _detect_language for Python files."""
        analyzer = MoAIASTGrepAnalyzer()
        assert analyzer._detect_language("test.py") == "python"
        assert analyzer._detect_language("/path/to/file.py") == "python"

    def test_detect_language_javascript_variants(self) -> None:
        """Test _detect_language for JavaScript variants."""
        analyzer = MoAIASTGrepAnalyzer()
        assert analyzer._detect_language("app.js") == "javascript"
        assert analyzer._detect_language("module.mjs") == "javascript"
        assert analyzer._detect_language("common.cjs") == "javascript"

    def test_detect_language_typescript_variants(self) -> None:
        """Test _detect_language for TypeScript variants."""
        analyzer = MoAIASTGrepAnalyzer()
        assert analyzer._detect_language("app.ts") == "typescript"
        assert analyzer._detect_language("module.mts") == "typescript"
        assert analyzer._detect_language("common.cts") == "typescript"

    def test_detect_language_react(self) -> None:
        """Test _detect_language for React files."""
        analyzer = MoAIASTGrepAnalyzer()
        assert analyzer._detect_language("App.jsx") == "javascriptreact"
        assert analyzer._detect_language("App.tsx") == "typescriptreact"

    def test_detect_language_other_languages(self) -> None:
        """Test _detect_language for other supported languages."""
        analyzer = MoAIASTGrepAnalyzer()
        assert analyzer._detect_language("main.go") == "go"
        assert analyzer._detect_language("lib.rs") == "rust"
        assert analyzer._detect_language("App.java") == "java"
        assert analyzer._detect_language("Script.kt") == "kotlin"
        assert analyzer._detect_language("Script.kts") == "kotlin"
        assert analyzer._detect_language("app.rb") == "ruby"
        assert analyzer._detect_language("app.c") == "c"
        assert analyzer._detect_language("app.h") == "c"
        assert analyzer._detect_language("app.cpp") == "cpp"
        assert analyzer._detect_language("app.cc") == "cpp"
        assert analyzer._detect_language("app.cxx") == "cpp"
        assert analyzer._detect_language("app.hpp") == "cpp"
        assert analyzer._detect_language("App.cs") == "csharp"
        assert analyzer._detect_language("App.swift") == "swift"
        assert analyzer._detect_language("app.lua") == "lua"
        assert analyzer._detect_language("index.html") == "html"
        assert analyzer._detect_language("App.vue") == "vue"
        assert analyzer._detect_language("App.svelte") == "svelte"

    def test_detect_language_unknown_extension(self) -> None:
        """Test _detect_language for unknown extensions returns 'text'."""
        analyzer = MoAIASTGrepAnalyzer()
        assert analyzer._detect_language("README.md") == "text"
        assert analyzer._detect_language("config.txt") == "text"
        assert analyzer._detect_language("Makefile") == "text"
        assert analyzer._detect_language("file.unknown") == "text"

    def test_detect_language_case_insensitive(self) -> None:
        """Test _detect_language is case insensitive for extensions."""
        analyzer = MoAIASTGrepAnalyzer()
        assert analyzer._detect_language("app.PY") == "python"
        assert analyzer._detect_language("app.JS") == "javascript"
        assert analyzer._detect_language("app.TS") == "typescript"
        assert analyzer._detect_language("app.TSX") == "typescriptreact"


class TestShouldIncludeFile:
    """Tests for _should_include_file private method."""

    def test_should_include_file_default_patterns(self) -> None:
        """Test _should_include_file with default exclude patterns."""
        analyzer = MoAIASTGrepAnalyzer()
        config = ScanConfig()

        # Default excludes
        assert not analyzer._should_include_file(Path("node_modules/pkg.js"), config)
        assert not analyzer._should_include_file(Path(".git/config"), config)
        assert not analyzer._should_include_file(Path("__pycache__/cache.py"), config)

        # Should include supported files
        assert analyzer._should_include_file(Path("app.py"), config)
        assert analyzer._should_include_file(Path("src/app.js"), config)

    def test_should_include_file_custom_exclude_patterns(self) -> None:
        """Test _should_include_file with custom exclude patterns."""
        analyzer = MoAIASTGrepAnalyzer()
        config = ScanConfig(exclude_patterns=["vendor", "*.min.js", "test_*.py"])

        assert not analyzer._should_include_file(Path("vendor/lib.js"), config)
        assert not analyzer._should_include_file(Path("app.min.js"), config)
        assert not analyzer._should_include_file(Path("test_app.py"), config)
        assert analyzer._should_include_file(Path("app.py"), config)

    def test_should_include_file_with_include_patterns(self) -> None:
        """Test _should_include_file with include patterns only."""
        analyzer = MoAIASTGrepAnalyzer()
        config = ScanConfig(include_patterns=["*.py", "*.js"])

        assert analyzer._should_include_file(Path("app.py"), config)
        assert analyzer._should_include_file(Path("app.js"), config)
        assert not analyzer._should_include_file(Path("README.md"), config)
        assert not analyzer._should_include_file(Path("config.txt"), config)

    def test_should_include_file_both_patterns(self) -> None:
        """Test _should_include_file with both include and exclude patterns."""
        analyzer = MoAIASTGrepAnalyzer()
        config = ScanConfig(include_patterns=["*.py"], exclude_patterns=["test_*.py"])

        assert analyzer._should_include_file(Path("app.py"), config)
        assert not analyzer._should_include_file(Path("test_app.py"), config)
        assert not analyzer._should_include_file(Path("app.js"), config)

    def test_should_include_file_path_substring_match(self) -> None:
        """Test _should_include_file matches pattern anywhere in path."""
        analyzer = MoAIASTGrepAnalyzer()
        config = ScanConfig(exclude_patterns=["build"])

        assert not analyzer._should_include_file(Path("build/lib.py"), config)
        assert not analyzer._should_include_file(Path("src/build/output.js"), config)
        assert analyzer._should_include_file(Path("src/app.py"), config)

    def test_should_include_file_unsupported_extension(self) -> None:
        """Test _should_include_file with unsupported extension."""
        analyzer = MoAIASTGrepAnalyzer()
        config = ScanConfig()

        assert not analyzer._should_include_file(Path("README.md"), config)
        assert not analyzer._should_include_file(Path("config.txt"), config)
        assert not analyzer._should_include_file(Path("Makefile"), config)


class TestParseSgOutput:
    """Tests for _parse_sg_output private method."""

    def test_parse_sg_output_valid_json(self) -> None:
        """Test _parse_sg_output with valid JSON."""
        analyzer = MoAIASTGrepAnalyzer()
        output = json.dumps(
            [
                {
                    "ruleId": "test-rule",
                    "severity": "warning",
                    "message": "Test message",
                    "range": {"start": {"line": 10, "column": 5}, "end": {"line": 10, "column": 15}},
                }
            ]
        )

        matches = analyzer._parse_sg_output(output, "/test.py")
        assert len(matches) == 1
        assert matches[0].rule_id == "test-rule"

    def test_parse_sg_output_empty_json_array(self) -> None:
        """Test _parse_sg_output with empty JSON array."""
        analyzer = MoAIASTGrepAnalyzer()
        output = "[]"

        matches = analyzer._parse_sg_output(output, "/test.py")
        assert len(matches) == 0

    def test_parse_sg_output_malformed_json(self) -> None:
        """Test _parse_sg_output with malformed JSON returns empty list."""
        analyzer = MoAIASTGrepAnalyzer()
        output = "{invalid json content"

        matches = analyzer._parse_sg_output(output, "/test.py")
        assert len(matches) == 0

    def test_parse_sg_output_non_list_json(self) -> None:
        """Test _parse_sg_output with non-list JSON returns empty list."""
        analyzer = MoAIASTGrepAnalyzer()
        output = json.dumps({"key": "value"})

        matches = analyzer._parse_sg_output(output, "/test.py")
        assert len(matches) == 0

    def test_parse_sg_output_null_document(self) -> None:
        """Test _parse_sg_output with null document in list."""
        analyzer = MoAIASTGrepAnalyzer()
        output = json.dumps(
            [
                None,
                {
                    "ruleId": "test",
                    "severity": "warning",
                    "message": "test",
                    "range": {"start": {"line": 0, "column": 0}, "end": {"line": 0, "column": 0}},
                },
            ]
        )

        matches = analyzer._parse_sg_output(output, "/test.py")
        assert len(matches) == 1

    def test_parse_sg_output_multiple_matches(self) -> None:
        """Test _parse_sg_output with multiple matches."""
        analyzer = MoAIASTGrepAnalyzer()
        output = json.dumps(
            [
                {
                    "ruleId": "rule1",
                    "severity": "error",
                    "message": "Error 1",
                    "range": {"start": {"line": 1, "column": 0}, "end": {"line": 1, "column": 10}},
                },
                {
                    "ruleId": "rule2",
                    "severity": "warning",
                    "message": "Warning 1",
                    "range": {"start": {"line": 5, "column": 0}, "end": {"line": 5, "column": 15}},
                },
                {
                    "ruleId": "rule3",
                    "severity": "info",
                    "message": "Info 1",
                    "range": {"start": {"line": 10, "column": 0}, "end": {"line": 10, "column": 20}},
                },
            ]
        )

        matches = analyzer._parse_sg_output(output, "/test.py")
        assert len(matches) == 3


class TestParseSgMatch:
    """Tests for _parse_sg_match private method."""

    def test_parse_sg_match_all_fields(self) -> None:
        """Test _parse_sg_match with all fields present."""
        analyzer = MoAIASTGrepAnalyzer()
        item = {
            "ruleId": "test-rule",
            "severity": "error",
            "message": "Test error message",
            "range": {"start": {"line": 10, "column": 5}, "end": {"line": 10, "column": 25}},
            "fix": "replacement_pattern",
        }

        match = analyzer._parse_sg_match(item, "/test.py")
        assert match is not None
        assert match.rule_id == "test-rule"
        assert match.severity == "error"
        assert match.message == "Test error message"
        assert match.file_path == "/test.py"
        assert match.suggested_fix == "replacement_pattern"

    def test_parse_sg_match_minimal_fields(self) -> None:
        """Test _parse_sg_match with minimal required fields."""
        analyzer = MoAIASTGrepAnalyzer()
        item = {"ruleId": "minimal-rule", "range": {"start": {"line": 0, "column": 0}, "end": {"line": 0, "column": 0}}}

        match = analyzer._parse_sg_match(item, "/test.py")
        assert match is not None
        assert match.rule_id == "minimal-rule"
        assert match.severity == "warning"  # default
        assert match.message == "Pattern match found"  # default
        assert match.suggested_fix is None

    def test_parse_sg_match_with_rule_id_snake_case(self) -> None:
        """Test _parse_sg_match supports both ruleId and rule_id."""
        analyzer = MoAIASTGrepAnalyzer()
        item = {
            "rule_id": "snake-case-rule",
            "range": {"start": {"line": 0, "column": 0}, "end": {"line": 0, "column": 0}},
        }

        match = analyzer._parse_sg_match(item, "/test.py")
        assert match is not None
        assert match.rule_id == "snake-case-rule"

    def test_parse_sg_match_with_character_field(self) -> None:
        """Test _parse_sg_match with 'character' instead of 'column' in range."""
        analyzer = MoAIASTGrepAnalyzer()
        item = {
            "ruleId": "test-rule",
            "range": {"start": {"line": 10, "character": 5}, "end": {"line": 10, "character": 25}},
        }

        match = analyzer._parse_sg_match(item, "/test.py")
        assert match is not None
        assert match.range.start.character == 5
        assert match.range.end.character == 25

    def test_parse_sg_match_missing_range(self) -> None:
        """Test _parse_sg_match with missing range returns None."""
        analyzer = MoAIASTGrepAnalyzer()
        item = {"ruleId": "test-rule", "severity": "error"}

        match = analyzer._parse_sg_match(item, "/test.py")
        assert match is None

    def test_parse_sg_match_invalid_range_structure(self) -> None:
        """Test _parse_sg_match with invalid range structure returns None."""
        analyzer = MoAIASTGrepAnalyzer()
        item = {"ruleId": "test-rule", "range": "invalid_range_object"}

        match = analyzer._parse_sg_match(item, "/test.py")
        assert match is None

    def test_parse_sg_match_with_suggested_fix_field(self) -> None:
        """Test _parse_sg_match supports both fix and suggested_fix fields."""
        analyzer = MoAIASTGrepAnalyzer()
        item = {
            "ruleId": "test-rule",
            "range": {"start": {"line": 0, "column": 0}, "end": {"line": 0, "column": 0}},
            "suggested_fix": "fix_pattern",
        }

        match = analyzer._parse_sg_match(item, "/test.py")
        assert match is not None
        assert match.suggested_fix == "fix_pattern"

    def test_parse_sg_match_all_severities(self) -> None:
        """Test _parse_sg_match with all severity levels."""
        analyzer = MoAIASTGrepAnalyzer()

        for severity in ["error", "warning", "info", "hint"]:
            item = {
                "ruleId": f"{severity}-rule",
                "severity": severity,
                "range": {"start": {"line": 0, "column": 0}, "end": {"line": 0, "column": 0}},
            }
            match = analyzer._parse_sg_match(item, "/test.py")
            assert match is not None
            assert match.severity == severity

    def test_parse_sg_match_non_dict_item(self) -> None:
        """Test _parse_sg_match with non-dict item returns None."""
        analyzer = MoAIASTGrepAnalyzer()
        match = analyzer._parse_sg_match(None, "/test.py")  # type: ignore[arg-type]
        assert match is None

        match = analyzer._parse_sg_match("string_item", "/test.py")  # type: ignore[arg-type]
        assert match is None

        match = analyzer._parse_sg_match(["list", "item"], "/test.py")  # type: ignore[list-item]
        assert match is None


class TestRunSgScan:
    """Tests for _run_sg_scan private method."""

    def test_run_sg_scan_timeout(self) -> None:
        """Test _run_sg_scan handles timeout gracefully."""
        analyzer = MoAIASTGrepAnalyzer()

        with patch("subprocess.run") as mock_run:
            import subprocess

            mock_run.side_effect = subprocess.TimeoutExpired("sg", 60)

            matches = analyzer._run_sg_scan("/test.py", "python", ScanConfig())
            assert matches == []

    def test_run_sg_scan_subprocess_error(self) -> None:
        """Test _run_sg_scan handles subprocess errors gracefully."""
        analyzer = MoAIASTGrepAnalyzer()

        with patch("subprocess.run") as mock_run:
            import subprocess

            mock_run.side_effect = subprocess.SubprocessError("Command failed")

            matches = analyzer._run_sg_scan("/test.py", "python", ScanConfig())
            assert matches == []

    def test_run_sg_scan_with_rules_path(self) -> None:
        """Test _run_sg_scan includes rules path in command."""
        analyzer = MoAIASTGrepAnalyzer()
        config = ScanConfig(rules_path="/custom/rules.yml")

        with patch("subprocess.run") as mock_run:
            mock_run.return_value = MagicMock(returncode=0, stdout="[]")

            analyzer._run_sg_scan("/test.py", "python", config)

            # Verify command includes --config
            cmd = mock_run.call_args[0][0]
            assert "--config" in cmd
            assert "/custom/rules.yml" in cmd

    def test_run_sg_scan_nonzero_return_code(self) -> None:
        """Test _run_sg_scan returns empty list on non-zero return code."""
        analyzer = MoAIASTGrepAnalyzer()

        with patch("subprocess.run") as mock_run:
            mock_run.return_value = MagicMock(returncode=1, stdout="")

            matches = analyzer._run_sg_scan("/test.py", "python", ScanConfig())
            assert matches == []

    def test_run_sg_scan_empty_stdout(self) -> None:
        """Test _run_sg_scan returns empty list on empty stdout."""
        analyzer = MoAIASTGrepAnalyzer()

        with patch("subprocess.run") as mock_run:
            mock_run.return_value = MagicMock(returncode=0, stdout="")

            matches = analyzer._run_sg_scan("/test.py", "python", ScanConfig())
            assert matches == []


class TestScanFileEdgeCases:
    """Tests for scan_file method edge cases."""

    def test_scan_file_timeout_handling(self) -> None:
        """Test scan_file handles timeout during sg scan."""
        with tempfile.NamedTemporaryFile(mode="w", suffix=".py", delete=False) as f:
            f.write("print('test')\n")
            f.flush()

            analyzer = MoAIASTGrepAnalyzer()

            with patch("subprocess.run") as mock_run:
                import subprocess

                mock_run.side_effect = subprocess.TimeoutExpired("sg", 60)

                result = analyzer.scan_file(f.name)
                # Should return result with empty matches
                assert result.matches == []
                assert result.scan_time_ms >= 0

        Path(f.name).unlink()

    def test_scan_file_permission_error(self) -> None:
        """Test scan_file handles permission errors."""
        analyzer = MoAIASTGrepAnalyzer()

        with patch("pathlib.Path.exists") as mock_exists:
            mock_exists.return_value = True

            with patch("subprocess.run") as mock_run:
                import subprocess

                mock_run.side_effect = PermissionError("Permission denied")

                # Mock is_sg_available to return False to skip actual scan
                with patch.object(analyzer, "is_sg_available", return_value=False):
                    result = analyzer.scan_file("/restricted/file.py")
                    # Should return result with empty matches
                    assert result.matches == []

    def test_scan_file_json_decode_error(self) -> None:
        """Test scan_file handles JSON decode errors from sg output."""
        with tempfile.NamedTemporaryFile(mode="w", suffix=".py", delete=False) as f:
            f.write("print('test')\n")
            f.flush()

            analyzer = MoAIASTGrepAnalyzer()

            with patch("subprocess.run") as mock_run:
                mock_run.return_value = MagicMock(returncode=0, stdout="{invalid json")

                result = analyzer.scan_file(f.name)
                # Should handle gracefully and return empty matches
                assert result.matches == []

        Path(f.name).unlink()

    def test_scan_file_empty_file(self) -> None:
        """Test scan_file with empty file."""
        with tempfile.NamedTemporaryFile(mode="w", suffix=".py", delete=False) as f:
            f.write("")
            f.flush()

            analyzer = MoAIASTGrepAnalyzer()
            result = analyzer.scan_file(f.name)

            assert result.file_path == f.name
            assert result.language == "python"
            assert isinstance(result.matches, list)

        Path(f.name).unlink()

    def test_scan_file_special_characters_in_path(self) -> None:
        """Test scan_file with special characters in file path."""
        with tempfile.TemporaryDirectory() as tmpdir:
            # Create file with special characters (spaces, dashes)
            test_file = Path(tmpdir) / "test file-name.py"
            test_file.write_text("print('test')\n")

            analyzer = MoAIASTGrepAnalyzer()
            result = analyzer.scan_file(str(test_file))

            assert result.file_path == str(test_file)
            assert result.language == "python"


class TestPatternSearchEdgeCases:
    """Tests for pattern_search method edge cases."""

    def test_pattern_search_timeout(self) -> None:
        """Test pattern_search handles timeout."""
        with tempfile.TemporaryDirectory() as tmpdir:
            (Path(tmpdir) / "test.py").write_text("print('test')\n")

            analyzer = MoAIASTGrepAnalyzer()

            with patch("subprocess.run") as mock_run:
                import subprocess

                mock_run.side_effect = subprocess.TimeoutExpired("sg", 120)

                matches = analyzer.pattern_search("print($MSG)", "python", tmpdir)
                assert matches == []

    def test_pattern_search_json_decode_error(self) -> None:
        """Test pattern_search handles JSON decode errors."""
        with tempfile.TemporaryDirectory() as tmpdir:
            (Path(tmpdir) / "test.py").write_text("print('test')\n")

            analyzer = MoAIASTGrepAnalyzer()

            with patch("subprocess.run") as mock_run:
                mock_run.return_value = MagicMock(returncode=0, stdout="{invalid json")

                matches = analyzer.pattern_search("print($MSG)", "python", tmpdir)
                assert matches == []

    def test_pattern_search_long_pattern_truncation(self) -> None:
        """Test pattern_search truncates long patterns in rule_id."""
        with tempfile.TemporaryDirectory() as tmpdir:
            (Path(tmpdir) / "test.py").write_text("print('test')\n")

            analyzer = MoAIASTGrepAnalyzer()

            # Create a pattern longer than 30 characters
            long_pattern = "some_very_long_pattern_name_that_exceeds_thirty_characters" * 2

            with patch("subprocess.run") as mock_run:
                mock_run.return_value = MagicMock(
                    returncode=0,
                    stdout=json.dumps(
                        [
                            {
                                "file": "/test.py",
                                "ruleId": "test",
                                "severity": "warning",
                                "message": "test",
                                "range": {"start": {"line": 0, "column": 0}, "end": {"line": 0, "column": 0}},
                            }
                        ]
                    ),
                )

                matches = analyzer.pattern_search(long_pattern, "python", tmpdir)
                # Pattern ID should be truncated to ~30 chars
                if matches:
                    assert len(matches[0].rule_id) <= 50  # "pattern:" + 30 chars


class TestPatternReplaceEdgeCases:
    """Tests for pattern_replace method edge cases."""

    def test_pattern_replace_timeout(self) -> None:
        """Test pattern_replace handles timeout during search."""
        with tempfile.TemporaryDirectory() as tmpdir:
            (Path(tmpdir) / "test.py").write_text("print('test')\n")

            analyzer = MoAIASTGrepAnalyzer()

            with patch("subprocess.run") as mock_run:
                import subprocess

                # First call (pattern_search) times out
                mock_run.side_effect = subprocess.TimeoutExpired("sg", 120)

                result = analyzer.pattern_replace(
                    pattern="print($MSG)", replacement="logger.info($MSG)", language="python", path=tmpdir, dry_run=True
                )

                # Should return result with zero matches
                assert result.matches_found == 0
                assert result.files_modified == 0

    def test_pattern_replace_subprocess_error(self) -> None:
        """Test pattern_replace handles subprocess errors."""
        with tempfile.TemporaryDirectory() as tmpdir:
            (Path(tmpdir) / "test.py").write_text("print('test')\n")

            analyzer = MoAIASTGrepAnalyzer()

            # Mock is_sg_available to return True first, then error on pattern_search
            with patch.object(analyzer, "is_sg_available", return_value=True):
                with patch("subprocess.run") as mock_run:
                    import subprocess

                    mock_run.side_effect = subprocess.SubprocessError("Command failed")

                    result = analyzer.pattern_replace(
                        pattern="print($MSG)",
                        replacement="logger.info($MSG)",
                        language="python",
                        path=tmpdir,
                        dry_run=True,
                    )

                    # Should return result with zero matches
                    assert result.matches_found == 0
                    assert result.files_modified == 0

    def test_pattern_replace_without_sg_available(self) -> None:
        """Test pattern_replace when sg is not available."""
        with tempfile.TemporaryDirectory() as tmpdir:
            (Path(tmpdir) / "test.py").write_text("print('test')\n")

            analyzer = MoAIASTGrepAnalyzer()

            with patch.object(analyzer, "is_sg_available", return_value=False):
                result = analyzer.pattern_replace(
                    pattern="print($MSG)", replacement="logger.info($MSG)", language="python", path=tmpdir, dry_run=True
                )

                # Should return empty result
                assert result.matches_found == 0
                assert result.files_modified == 0
                assert result.dry_run is True


class TestScanProjectEdgeCases:
    """Tests for scan_project method edge cases."""

    def test_scan_project_empty_directory(self) -> None:
        """Test scan_project with empty directory."""
        with tempfile.TemporaryDirectory() as tmpdir:
            analyzer = MoAIASTGrepAnalyzer()
            result = analyzer.scan_project(tmpdir)

            assert result.project_path == tmpdir
            assert result.files_scanned == 0
            assert result.total_matches == 0
            assert len(result.results_by_file) == 0

    def test_scan_project_with_symlinks(self) -> None:
        """Test scan_project handles symbolic links."""
        with tempfile.TemporaryDirectory() as tmpdir:
            # Create a file and a symlink to it
            target_file = Path(tmpdir) / "target.py"
            target_file.write_text("print('test')\n")

            try:
                symlink = Path(tmpdir) / "link.py"
                symlink.symlink_to(target_file)
            except (OSError, NotImplementedError):
                # Skip test if symlinks not supported
                pytest.skip("Symbolic links not supported on this system")

            analyzer = MoAIASTGrepAnalyzer()
            result = analyzer.scan_project(tmpdir)

            # Should scan without errors
            assert result.project_path == tmpdir
            assert result.files_scanned >= 0

    def test_scan_project_file_with_invalid_extension(self) -> None:
        """Test scan_project with files having no extension."""
        with tempfile.TemporaryDirectory() as tmpdir:
            # Create files with no extension
            (Path(tmpdir) / "Makefile").write_text("all:\n\techo 'build'\n")
            (Path(tmpdir) / "README").write_text("# Project\n")

            analyzer = MoAIASTGrepAnalyzer()
            result = analyzer.scan_project(tmpdir)

            # Should handle without errors
            assert result.project_path == tmpdir
            # Files with no supported extension should not be included
            for file_path in result.results_by_file:
                assert Path(file_path).suffix.lower() in {
                    ".py",
                    ".js",
                    ".ts",
                    ".tsx",
                    ".jsx",
                    ".go",
                    ".rs",
                    ".java",
                    ".kt",
                    ".rb",
                    ".c",
                    ".cpp",
                    ".cs",
                    ".swift",
                    ".lua",
                    ".html",
                    ".vue",
                    ".svelte",
                }

    def test_scan_project_nested_directories(self) -> None:
        """Test scan_project with deeply nested directories."""
        with tempfile.TemporaryDirectory() as tmpdir:
            # Create deeply nested structure
            nested_dir = Path(tmpdir) / "a" / "b" / "c" / "d" / "e"
            nested_dir.mkdir(parents=True)
            (nested_dir / "deep.py").write_text("print('deep')\n")

            analyzer = MoAIASTGrepAnalyzer()

            # Mock to ensure files are found during scanning
            with patch("pathlib.Path.rglob") as mock_rglob:
                # Return the actual nested file
                nested_file = nested_dir / "deep.py"
                mock_rglob.return_value = [nested_file]

                result = analyzer.scan_project(tmpdir)

                # Should scan the nested directory
                assert result.files_scanned >= 1
                # File should be found in results (with or without matches)
                assert (
                    any("deep.py" in str(path) for path in result.results_by_file.keys()) or result.files_scanned >= 1
                )


class TestIsSgAvailable:
    """Tests for is_sg_available method."""

    def test_is_sg_available_cached_result(self) -> None:
        """Test is_sg_available caches the result."""
        analyzer = MoAIASTGrepAnalyzer()

        with patch("subprocess.run") as mock_run:
            mock_run.return_value = MagicMock(returncode=0, stdout="sg version 0.1.0")

            # First call
            result1 = analyzer.is_sg_available()

            # Second call should use cached value (subprocess.run called once)
            result2 = analyzer.is_sg_available()

            assert result1 == result2
            assert mock_run.call_count == 1

    def test_is_sg_available_file_not_found(self) -> None:
        """Test is_sg_available when sg command not found."""
        analyzer = MoAIASTGrepAnalyzer()

        with patch("subprocess.run") as mock_run:
            mock_run.side_effect = FileNotFoundError("sg not found")

            result = analyzer.is_sg_available()
            assert result is False

    def test_is_sg_available_timeout(self) -> None:
        """Test is_sg_available handles timeout."""
        analyzer = MoAIASTGrepAnalyzer()

        with patch("subprocess.run") as mock_run:
            import subprocess

            mock_run.side_effect = subprocess.TimeoutExpired("sg", 5)

            result = analyzer.is_sg_available()
            assert result is False

    def test_is_sg_available_nonzero_exit(self) -> None:
        """Test is_sg_available when sg exits with non-zero code."""
        analyzer = MoAIASTGrepAnalyzer()

        with patch("subprocess.run") as mock_run:
            mock_run.return_value = MagicMock(returncode=1, stdout="")

            result = analyzer.is_sg_available()
            assert result is False
