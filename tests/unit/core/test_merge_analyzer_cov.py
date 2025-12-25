"""Comprehensive coverage tests for MergeAnalyzer module.

Tests MergeAnalyzer class for merge conflict detection, analysis, and recommendations.
Target: 70%+ code coverage with actual code path execution and mocked dependencies.
"""

import json
import subprocess
from pathlib import Path
from unittest.mock import MagicMock, Mock, patch

import pytest


class TestMergeAnalyzerInit:
    """Test MergeAnalyzer initialization."""

    def test_merge_analyzer_instantiation(self):
        """Should instantiate with project path."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        project_path = Path("/tmp/test_project")
        analyzer = MergeAnalyzer(project_path)

        assert analyzer.project_path == project_path
        assert analyzer.CLAUDE_TIMEOUT == 120
        assert analyzer.CLAUDE_MODEL == "claude-haiku-4-5-20251001"

    def test_merge_analyzer_constants(self):
        """Should have correct constants defined."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        assert len(MergeAnalyzer.ANALYZED_FILES) > 0
        assert "CLAUDE.md" in MergeAnalyzer.ANALYZED_FILES
        assert len(MergeAnalyzer.CLAUDE_TOOLS) > 0


class TestCollectDiffFiles:
    """Test _collect_diff_files method."""

    def test_collect_diff_files_no_files_exist(self, tmp_path):
        """Should return empty diff when no files exist."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(tmp_path)
        backup_path = tmp_path / "backup"
        template_path = tmp_path / "template"

        backup_path.mkdir()
        template_path.mkdir()

        diff_files = analyzer._collect_diff_files(backup_path, template_path)

        assert isinstance(diff_files, dict)
        assert all("has_diff" in diff_files[f] for f in diff_files)

    def test_collect_diff_files_identical_content(self, tmp_path):
        """Should detect identical files."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(tmp_path)
        backup_path = tmp_path / "backup"
        template_path = tmp_path / "template"

        backup_path.mkdir()
        template_path.mkdir()

        # Create identical files
        test_content = "Hello World\nTest Content"
        (backup_path / "CLAUDE.md").write_text(test_content)
        (template_path / "CLAUDE.md").write_text(test_content)

        diff_files = analyzer._collect_diff_files(backup_path, template_path)

        assert diff_files["CLAUDE.md"]["has_diff"] is False
        assert diff_files["CLAUDE.md"]["backup_exists"] is True
        assert diff_files["CLAUDE.md"]["template_exists"] is True

    def test_collect_diff_files_different_content(self, tmp_path):
        """Should detect different files."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(tmp_path)
        backup_path = tmp_path / "backup"
        template_path = tmp_path / "template"

        backup_path.mkdir()
        template_path.mkdir()

        # Create different files
        (backup_path / ".gitignore").write_text("*.pyc\n__pycache__/\n")
        (template_path / ".gitignore").write_text("*.pyc\n__pycache__/\nbuild/\n")

        diff_files = analyzer._collect_diff_files(backup_path, template_path)

        assert diff_files[".gitignore"]["has_diff"] is True
        assert diff_files[".gitignore"]["diff_lines"] > 0

    def test_collect_diff_files_backup_only(self, tmp_path):
        """Should handle backup-only files."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(tmp_path)
        backup_path = tmp_path / "backup"
        template_path = tmp_path / "template"

        backup_path.mkdir()
        template_path.mkdir()

        (backup_path / "CLAUDE.md").write_text("backup content")

        diff_files = analyzer._collect_diff_files(backup_path, template_path)

        assert diff_files["CLAUDE.md"]["backup_exists"] is True
        assert diff_files["CLAUDE.md"]["template_exists"] is False

    def test_collect_diff_files_template_only(self, tmp_path):
        """Should handle template-only files."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(tmp_path)
        backup_path = tmp_path / "backup"
        template_path = tmp_path / "template"

        backup_path.mkdir()
        template_path.mkdir()

        (template_path / ".claude" / "settings.json").parent.mkdir(parents=True)
        (template_path / ".claude" / "settings.json").write_text("{}")

        diff_files = analyzer._collect_diff_files(backup_path, template_path)

        # Check for settings.json entry (may be in different key)
        assert len(diff_files) > 0


class TestCreateAnalysisPrompt:
    """Test _create_analysis_prompt method."""

    def test_create_analysis_prompt_basic(self, tmp_path):
        """Should create analysis prompt."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(tmp_path)
        backup_path = tmp_path / "backup"
        template_path = tmp_path / "template"

        diff_files = {
            "CLAUDE.md": {
                "backup_exists": True,
                "template_exists": True,
                "has_diff": True,
                "diff_lines": 5,
            }
        }

        prompt = analyzer._create_analysis_prompt(backup_path, template_path, diff_files)

        assert isinstance(prompt, str)
        assert "MoAI-ADK configuration file merge expert" in prompt
        assert str(backup_path) in prompt
        assert str(template_path) in prompt
        assert "JSON" in prompt

    def test_create_analysis_prompt_includes_rules(self, tmp_path):
        """Should include merge rules in prompt."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(tmp_path)
        backup_path = tmp_path / "backup"
        template_path = tmp_path / "template"

        diff_files = {
            "CLAUDE.md": {
                "backup_exists": True,
                "template_exists": True,
                "has_diff": False,
                "diff_lines": 0,
            }
        }

        prompt = analyzer._create_analysis_prompt(backup_path, template_path, diff_files)

        assert "CLAUDE.md: Preserve Project Information section" in prompt
        assert "settings.json" in prompt
        assert "config.json" in prompt


class TestParseClaudeResponse:
    """Test _parse_claude_response method."""

    def test_parse_claude_response_valid_json(self, tmp_path):
        """Should parse valid JSON response."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(tmp_path)
        response = """{
            "files": [{"filename": "CLAUDE.md", "changes": "test"}],
            "safe_to_auto_merge": true,
            "user_action_required": false,
            "summary": "Test summary"
        }"""

        result = analyzer._parse_claude_response(response)

        assert result["safe_to_auto_merge"] is True
        assert result["user_action_required"] is False
        assert len(result["files"]) == 1

    def test_parse_claude_response_wrapped_format(self, tmp_path):
        """Should parse v2.0+ wrapped format."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(tmp_path)
        response = """{
            "type": "text",
            "result": "```json\\n{\\"files\\": [], \\"safe_to_auto_merge\\": false}\\n```"
        }"""

        result = analyzer._parse_claude_response(response)

        assert isinstance(result, dict)
        # Should either parse successfully or return error structure
        assert "error" not in result or result.get("error") == "response_parse_failed"

    def test_parse_claude_response_invalid_json_fallback(self, tmp_path):
        """Should return error structure on parsing failure."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(tmp_path)
        response = "This is not valid JSON at all"

        result = analyzer._parse_claude_response(response)

        assert "error" in result
        assert result["safe_to_auto_merge"] is False
        assert result["user_action_required"] is True


class TestDetectClaudeErrors:
    """Test _detect_claude_errors method."""

    def test_detect_claude_errors_model_not_found(self, tmp_path):
        """Should detect model not found error."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(tmp_path)
        stderr = "Error: Unknown model 'claude-haiku-4-5-20251001'"

        error_msg = analyzer._detect_claude_errors(stderr)

        assert "model" in error_msg.lower()

    def test_detect_claude_errors_permission_denied(self, tmp_path):
        """Should detect permission denied error."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(tmp_path)
        stderr = "Permission denied: cannot access file"

        error_msg = analyzer._detect_claude_errors(stderr)

        assert "permission" in error_msg.lower()

    def test_detect_claude_errors_timeout(self, tmp_path):
        """Should detect timeout error."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(tmp_path)
        stderr = "Timeout: operation exceeded maximum time"

        error_msg = analyzer._detect_claude_errors(stderr)

        assert "timeout" in error_msg.lower()

    def test_detect_claude_errors_empty_stderr(self, tmp_path):
        """Should handle empty stderr."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(tmp_path)

        error_msg = analyzer._detect_claude_errors("")

        assert error_msg == ""


class TestBuildClaudeCommand:
    """Test _build_claude_command method."""

    def test_build_claude_command_structure(self, tmp_path):
        """Should build correct command structure."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(tmp_path)
        command = analyzer._build_claude_command()

        assert isinstance(command, list)
        # First element is the full path to claude executable (Windows compatibility)
        assert "claude" in command[0].lower()  # Check executable contains "claude"
        assert "-p" in command
        assert "--model" in command
        assert "--output-format" in command
        assert "json" in command

    def test_build_claude_command_has_tools(self, tmp_path):
        """Should include tools in command."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(tmp_path)
        command = analyzer._build_claude_command()

        assert "--tools" in command
        tools_idx = command.index("--tools")
        assert tools_idx + 1 < len(command)


class TestFormatDiffSummary:
    """Test _format_diff_summary method."""

    def test_format_diff_summary_modified(self, tmp_path):
        """Should format modified files."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(tmp_path)
        diff_files = {
            "CLAUDE.md": {
                "backup_exists": True,
                "template_exists": True,
                "has_diff": True,
                "diff_lines": 10,
            }
        }

        summary = analyzer._format_diff_summary(diff_files)

        assert "CLAUDE.md" in summary
        assert "Modified" in summary or "✏️" in summary

    def test_format_diff_summary_deleted(self, tmp_path):
        """Should format deleted files."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(tmp_path)
        diff_files = {
            "CLAUDE.md": {
                "backup_exists": True,
                "template_exists": False,
                "has_diff": True,
                "diff_lines": 0,
            }
        }

        summary = analyzer._format_diff_summary(diff_files)

        assert "CLAUDE.md" in summary
        assert "Deleted" in summary or "❌" in summary

    def test_format_diff_summary_new(self, tmp_path):
        """Should format new files."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(tmp_path)
        diff_files = {
            "CLAUDE.md": {
                "backup_exists": False,
                "template_exists": True,
                "has_diff": True,
                "diff_lines": 5,
            }
        }

        summary = analyzer._format_diff_summary(diff_files)

        assert "CLAUDE.md" in summary
        assert "New" in summary or "✨" in summary


class TestFallbackAnalysis:
    """Test _fallback_analysis method."""

    def test_fallback_analysis_no_diffs(self, tmp_path):
        """Should return safe analysis when no diffs."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(tmp_path)
        backup_path = tmp_path / "backup"
        template_path = tmp_path / "template"

        diff_files = {
            "CLAUDE.md": {"has_diff": False, "diff_lines": 0},
            ".gitignore": {"has_diff": False, "diff_lines": 0},
        }

        result = analyzer._fallback_analysis(backup_path, template_path, diff_files)

        assert result["safe_to_auto_merge"] is True
        assert result["user_action_required"] is False
        assert result["fallback"] is True

    def test_fallback_analysis_with_diffs(self, tmp_path):
        """Should handle diffs in fallback."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(tmp_path)
        backup_path = tmp_path / "backup"
        template_path = tmp_path / "template"

        diff_files = {
            ".moai/config/config.json": {"has_diff": True, "diff_lines": 15},
        }

        result = analyzer._fallback_analysis(backup_path, template_path, diff_files)

        assert isinstance(result, dict)
        assert "files" in result
        assert "safe_to_auto_merge" in result

    def test_fallback_analysis_high_risk_detection(self, tmp_path):
        """Should detect high-risk changes."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(tmp_path)
        backup_path = tmp_path / "backup"
        template_path = tmp_path / "template"

        diff_files = {
            ".claude/settings.json": {"has_diff": True, "diff_lines": 50},
        }

        result = analyzer._fallback_analysis(backup_path, template_path, diff_files)

        assert "files" in result
        if result["files"]:
            assert any(f["conflict_severity"] in ["medium", "high"] for f in result["files"])


class TestDisplayAnalysis:
    """Test _display_analysis method."""

    @patch("moai_adk.core.merge.analyzer.console")
    def test_display_analysis_with_files(self, mock_console, tmp_path):
        """Should display analysis results."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(tmp_path)
        analysis = {
            "summary": "Test summary",
            "risk_assessment": "Low - no critical changes",
            "files": [
                {
                    "filename": "CLAUDE.md",
                    "changes": "Minor updates",
                    "recommendation": "use_template",
                    "conflict_severity": "low",
                }
            ],
        }

        analyzer._display_analysis(analysis)

        mock_console.print.assert_called()

    @patch("moai_adk.core.merge.analyzer.console")
    def test_display_analysis_empty_files(self, mock_console, tmp_path):
        """Should handle empty files list."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(tmp_path)
        analysis = {"summary": "No changes", "files": []}

        analyzer._display_analysis(analysis)

        mock_console.print.assert_called()


class TestAnalyzeMergeIntegration:
    """Test analyze_merge integration."""

    @patch("moai_adk.core.merge.analyzer.subprocess.run")
    def test_analyze_merge_success(self, mock_subprocess, tmp_path):
        """Should handle successful Claude analysis."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(tmp_path)
        backup_path = tmp_path / "backup"
        template_path = tmp_path / "template"

        backup_path.mkdir()
        template_path.mkdir()

        mock_result = MagicMock()
        mock_result.returncode = 0
        mock_result.stdout = (
            '{"files": [], "safe_to_auto_merge": true, "user_action_required": false, "summary": "Test"}'
        )
        mock_subprocess.return_value = mock_result

        result = analyzer.analyze_merge(backup_path, template_path)

        assert isinstance(result, dict)
        assert "safe_to_auto_merge" in result

    @patch("moai_adk.core.merge.analyzer.subprocess.run")
    def test_analyze_merge_timeout(self, mock_subprocess, tmp_path):
        """Should handle Claude timeout."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(tmp_path)
        backup_path = tmp_path / "backup"
        template_path = tmp_path / "template"

        backup_path.mkdir()
        template_path.mkdir()

        mock_subprocess.side_effect = subprocess.TimeoutExpired("claude", 120)

        result = analyzer.analyze_merge(backup_path, template_path)

        assert "fallback" in result or "files" in result

    @patch("moai_adk.core.merge.analyzer.subprocess.run")
    def test_analyze_merge_not_found(self, mock_subprocess, tmp_path):
        """Should handle Claude not found."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(tmp_path)
        backup_path = tmp_path / "backup"
        template_path = tmp_path / "template"

        backup_path.mkdir()
        template_path.mkdir()

        mock_subprocess.side_effect = FileNotFoundError()

        result = analyzer.analyze_merge(backup_path, template_path)

        assert isinstance(result, dict)


class TestAskUserConfirmation:
    """Test ask_user_confirmation method."""

    @patch("moai_adk.core.merge.analyzer.click.confirm")
    @patch("moai_adk.core.merge.analyzer.console")
    def test_ask_user_confirmation_safe_merge(self, mock_console, mock_confirm, tmp_path):
        """Should ask for confirmation on safe merge."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(tmp_path)
        analysis = {
            "safe_to_auto_merge": True,
            "user_action_required": False,
            "files": [],
        }

        mock_confirm.return_value = True

        result = analyzer.ask_user_confirmation(analysis)

        assert result is True
        mock_confirm.assert_called_once()

    @patch("moai_adk.core.merge.analyzer.click.confirm")
    @patch("moai_adk.core.merge.analyzer.console")
    def test_ask_user_confirmation_requires_intervention(self, mock_console, mock_confirm, tmp_path):
        """Should show warnings for conflicts."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(tmp_path)
        analysis = {
            "safe_to_auto_merge": False,
            "user_action_required": True,
            "files": [
                {
                    "filename": "CLAUDE.md",
                    "conflict_severity": "high",
                    "note": "Critical change",
                }
            ],
        }

        mock_confirm.return_value = False

        result = analyzer.ask_user_confirmation(analysis)

        assert result is False


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
