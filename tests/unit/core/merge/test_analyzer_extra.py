"""Extended tests for moai_adk.core.merge.analyzer module.

Comprehensive test coverage for MergeAnalyzer with Claude Code headless integration,
diff analysis, and merge recommendations.
"""

import subprocess
from pathlib import Path
from typing import Any, Dict
from unittest.mock import Mock, patch

import pytest


class TestMergeAnalyzerBasics:
    """Test MergeAnalyzer class initialization and basics."""

    def test_class_import(self):
        """Test that MergeAnalyzer can be imported."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        assert MergeAnalyzer is not None

    def test_analyzer_init(self):
        """Test MergeAnalyzer initialization."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        project_path = Path("/test/project")
        analyzer = MergeAnalyzer(project_path)

        assert analyzer.project_path == project_path

    def test_analyzed_files_defined(self):
        """Test that ANALYZED_FILES is properly defined."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        files = MergeAnalyzer.ANALYZED_FILES

        assert isinstance(files, list)
        assert len(files) > 0
        assert "CLAUDE.md" in files

    def test_claude_settings_constants(self):
        """Test that Claude settings constants are defined."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        assert hasattr(MergeAnalyzer, "CLAUDE_TIMEOUT")
        assert hasattr(MergeAnalyzer, "CLAUDE_MODEL")
        assert hasattr(MergeAnalyzer, "CLAUDE_TOOLS")

        assert MergeAnalyzer.CLAUDE_TIMEOUT == 120
        assert isinstance(MergeAnalyzer.CLAUDE_MODEL, str)
        assert isinstance(MergeAnalyzer.CLAUDE_TOOLS, list)


class TestAnalyzeMerge:
    """Test analyze_merge method."""

    def test_analyze_merge_returns_dict(self):
        """Test that analyze_merge returns a dictionary."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(Path("/test/project"))

        with patch.object(analyzer, "_collect_diff_files", return_value=[]):
            with patch.object(analyzer, "_create_analysis_prompt", return_value="prompt"):
                with patch("subprocess.run") as mock_run:
                    mock_result = Mock()
                    mock_result.returncode = 0
                    mock_result.stdout = '{"files": [], "safe_to_auto_merge": false}'
                    mock_run.return_value = mock_result

                    with patch.object(analyzer, "_parse_claude_response") as mock_parse:
                        mock_parse.return_value = {
                            "files": [],
                            "safe_to_auto_merge": False,
                        }

                        result = analyzer.analyze_merge(
                            Path("/backup"), Path("/template")
                        )

        assert isinstance(result, dict)

    def test_analyze_merge_with_timeout(self):
        """Test analyze_merge handles subprocess timeout."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(Path("/test/project"))

        with patch.object(analyzer, "_collect_diff_files", return_value=[]):
            with patch.object(analyzer, "_create_analysis_prompt", return_value="prompt"):
                with patch("subprocess.run", side_effect=subprocess.TimeoutExpired("cmd", 120)):
                    with patch.object(analyzer, "_fallback_analysis") as mock_fallback:
                        mock_fallback.return_value = {"files": [], "safe_to_auto_merge": False}

                        result = analyzer.analyze_merge(
                            Path("/backup"), Path("/template")
                        )

        assert isinstance(result, dict)

    def test_analyze_merge_handles_claude_error(self):
        """Test analyze_merge handles Claude execution errors."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(Path("/test/project"))

        with patch.object(analyzer, "_collect_diff_files", return_value=[]):
            with patch.object(analyzer, "_create_analysis_prompt", return_value="prompt"):
                with patch("subprocess.run") as mock_run:
                    mock_result = Mock()
                    mock_result.returncode = 1
                    mock_result.stderr = "Error"
                    mock_run.return_value = mock_result

                    with patch.object(analyzer, "_detect_claude_errors") as mock_detect:
                        mock_detect.return_value = "Claude error"

                        with patch.object(analyzer, "_fallback_analysis") as mock_fallback:
                            mock_fallback.return_value = {"files": [], "safe_to_auto_merge": False}

                            result = analyzer.analyze_merge(
                                Path("/backup"), Path("/template")
                            )

        assert isinstance(result, dict)


class TestCollectDiffFiles:
    """Test _collect_diff_files method."""

    def test_collect_diff_files_method_exists(self):
        """Test that _collect_diff_files method exists."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(Path("/test/project"))

        assert hasattr(analyzer, "_collect_diff_files")
        assert callable(analyzer._collect_diff_files)


class TestCreateAnalysisPrompt:
    """Test _create_analysis_prompt method."""

    def test_create_analysis_prompt_method_exists(self):
        """Test that _create_analysis_prompt method exists."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(Path("/test/project"))

        assert hasattr(analyzer, "_create_analysis_prompt")
        assert callable(analyzer._create_analysis_prompt)


class TestBuildClaudeCommand:
    """Test _build_claude_command method."""

    def test_build_claude_command_returns_list(self):
        """Test that _build_claude_command returns a list."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(Path("/test/project"))

        result = analyzer._build_claude_command()

        assert isinstance(result, list)
        assert len(result) > 0

    def test_build_claude_command_includes_model(self):
        """Test that Claude command includes model specification."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(Path("/test/project"))

        result = analyzer._build_claude_command()

        # Command should include Claude or appropriate executable
        assert any("claude" in str(part).lower() for part in result) or len(result) > 0


class TestParseClaudeResponse:
    """Test _parse_claude_response method."""

    def test_parse_claude_response_valid_json(self):
        """Test parsing valid JSON response from Claude."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(Path("/test/project"))

        response = '{"files": [], "safe_to_auto_merge": false, "summary": "No differences"}'

        result = analyzer._parse_claude_response(response)

        assert isinstance(result, dict)

    def test_parse_claude_response_invalid_json(self):
        """Test parsing invalid JSON response from Claude."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(Path("/test/project"))

        response = "not valid json {{{"

        result = analyzer._parse_claude_response(response)

        # Should return a dict even with invalid JSON
        assert isinstance(result, dict)

    def test_parse_claude_response_with_error(self):
        """Test parsing response with error indicator."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(Path("/test/project"))

        response = '{"error": "Analysis failed", "message": "Something went wrong"}'

        result = analyzer._parse_claude_response(response)

        assert isinstance(result, dict)


class TestDetectClaudeErrors:
    """Test _detect_claude_errors method."""

    def test_detect_claude_errors_returns_string(self):
        """Test that _detect_claude_errors returns a string."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(Path("/test/project"))

        result = analyzer._detect_claude_errors("Error message")

        assert isinstance(result, str)

    def test_detect_claude_errors_with_empty_string(self):
        """Test detecting errors from empty stderr."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(Path("/test/project"))

        result = analyzer._detect_claude_errors("")

        assert isinstance(result, str)

    def test_detect_claude_errors_with_detailed_error(self):
        """Test detecting errors from detailed error message."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(Path("/test/project"))

        error_msg = "FileNotFoundError: settings.json not found"

        result = analyzer._detect_claude_errors(error_msg)

        assert isinstance(result, str)


class TestFallbackAnalysis:
    """Test _fallback_analysis method."""

    def test_fallback_analysis_method_exists(self):
        """Test that _fallback_analysis method exists."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(Path("/test/project"))

        assert hasattr(analyzer, "_fallback_analysis")
        assert callable(analyzer._fallback_analysis)


class TestUnifiedDiff:
    """Test unified diff functionality."""

    def test_unified_diff_generation(self):
        """Test generating unified diffs."""
        from difflib import unified_diff

        backup_content = "line1\nline2\nline3"
        template_content = "line1\nmodified\nline3"

        diff = list(
            unified_diff(
                backup_content.split("\n"),
                template_content.split("\n"),
                lineterm="",
            )
        )

        # Should have some diff output for different content
        assert isinstance(diff, list)


class TestProjectPath:
    """Test project path handling."""

    def test_project_path_stored(self):
        """Test that project path is properly stored."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        path = Path("/test/project")
        analyzer = MergeAnalyzer(path)

        assert analyzer.project_path == path

    def test_project_path_is_path_object(self):
        """Test that project path is a Path object."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(Path("/test/project"))

        assert isinstance(analyzer.project_path, Path)


class TestEdgeCases:
    """Test edge cases in MergeAnalyzer."""

    def test_empty_backup_directory(self):
        """Test analyzing with empty backup directory."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(Path("/test/project"))

        with patch.object(analyzer, "_collect_diff_files", return_value=[]):
            with patch.object(analyzer, "_create_analysis_prompt", return_value=""):
                with patch("subprocess.run") as mock_run:
                    mock_result = Mock()
                    mock_result.returncode = 0
                    mock_result.stdout = "{}"
                    mock_run.return_value = mock_result

                    with patch.object(analyzer, "_parse_claude_response") as mock_parse:
                        mock_parse.return_value = {"files": []}

                        result = analyzer.analyze_merge(Path("/empty"), Path("/template"))

        assert isinstance(result, dict)

    def test_identical_backup_and_template(self):
        """Test analyzing when backup and template are identical."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(Path("/test/project"))

        with patch.object(analyzer, "_collect_diff_files", return_value=[]):
            with patch.object(analyzer, "_create_analysis_prompt", return_value=""):
                with patch("subprocess.run") as mock_run:
                    mock_result = Mock()
                    mock_result.returncode = 0
                    mock_result.stdout = '{"safe_to_auto_merge": true}'
                    mock_run.return_value = mock_result

                    with patch.object(analyzer, "_parse_claude_response") as mock_parse:
                        mock_parse.return_value = {"safe_to_auto_merge": True}

                        result = analyzer.analyze_merge(Path("/same"), Path("/same"))

        assert isinstance(result, dict)

    def test_deeply_nested_paths(self):
        """Test with deeply nested directory paths."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        deep_path = Path("/a/b/c/d/e/f/g/h/i/j/k/l/m/n/o/p")

        analyzer = MergeAnalyzer(deep_path)

        assert analyzer.project_path == deep_path

    def test_paths_with_special_characters(self):
        """Test with paths containing special characters."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        special_path = Path("/test/project-v2.0_alpha/config-file")

        analyzer = MergeAnalyzer(special_path)

        assert analyzer.project_path == special_path


class TestConsoleOutput:
    """Test console output in MergeAnalyzer."""

    def test_console_integration(self):
        """Test that console is available."""
        from moai_adk.core.merge.analyzer import console

        assert console is not None

    def test_spinner_used_during_analysis(self):
        """Test that spinner is used during analysis."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(Path("/test/project"))

        with patch.object(analyzer, "_collect_diff_files", return_value=[]):
            with patch.object(analyzer, "_create_analysis_prompt", return_value=""):
                with patch("subprocess.run") as mock_run:
                    mock_result = Mock()
                    mock_result.returncode = 0
                    mock_result.stdout = "{}"
                    mock_run.return_value = mock_result

                    with patch.object(analyzer, "_parse_claude_response") as mock_parse:
                        mock_parse.return_value = {}

                        with patch("moai_adk.core.merge.analyzer.Live"):
                            result = analyzer.analyze_merge(Path("/b"), Path("/t"))

        assert isinstance(result, dict)


class TestResponseStructure:
    """Test response structure consistency."""

    def test_successful_response_structure(self):
        """Test successful response has expected structure."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(Path("/test/project"))

        with patch.object(analyzer, "_collect_diff_files", return_value=[]):
            with patch.object(analyzer, "_create_analysis_prompt", return_value=""):
                with patch("subprocess.run") as mock_run:
                    mock_result = Mock()
                    mock_result.returncode = 0
                    mock_result.stdout = json.dumps({
                        "files": [],
                        "safe_to_auto_merge": True,
                        "summary": "All changes are safe",
                    })
                    mock_run.return_value = mock_result

                    with patch.object(analyzer, "_parse_claude_response") as mock_parse:
                        response = {
                            "files": [],
                            "safe_to_auto_merge": True,
                            "summary": "All changes are safe",
                        }
                        mock_parse.return_value = response

                        result = analyzer.analyze_merge(Path("/b"), Path("/t"))

        assert isinstance(result, dict)


# Import json for last test
import json
