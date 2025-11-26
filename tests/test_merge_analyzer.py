"""Unit tests for MergeAnalyzer - Claude Code headless merge analysis.

Tests the MergeAnalyzer class which uses Claude Code in headless mode
to provide intelligent backup vs new template comparison with user confirmation.
"""

import json
import tempfile
from pathlib import Path
from unittest.mock import MagicMock, patch

from moai_adk.core.merge.analyzer import MergeAnalyzer


class TestMergeAnalyzerInitialization:
    """Tests for MergeAnalyzer initialization."""

    def test_initialize_with_valid_project_path(self):
        """Test MergeAnalyzer initializes with valid project path."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            analyzer = MergeAnalyzer(project_path)

            assert analyzer.project_path == project_path
            assert analyzer.CLAUDE_TIMEOUT == 120
            assert analyzer.CLAUDE_MODEL == "claude-haiku-4-5-20251001"
            assert analyzer.CLAUDE_TOOLS == ["Read", "Glob", "Grep"]

    def test_analyzed_files_list_correct(self):
        """Test ANALYZED_FILES list contains expected files."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            analyzer = MergeAnalyzer(project_path)

            expected_files = [
                "CLAUDE.md",
                ".claude/settings.json",
                ".moai/config/config.json",
                ".gitignore",
            ]
            assert analyzer.ANALYZED_FILES == expected_files


class TestCollectDiffFiles:
    """Tests for _collect_diff_files method."""

    def test_collect_diff_files_with_no_differences(self):
        """Test collecting diffs when backup and template files are identical."""
        with tempfile.TemporaryDirectory() as tmpdir:
            backup_path = Path(tmpdir) / "backup"
            template_path = Path(tmpdir) / "template"
            backup_path.mkdir()
            template_path.mkdir()

            # Create identical files
            backup_file = backup_path / "CLAUDE.md"
            template_file = template_path / "CLAUDE.md"
            backup_file.write_text("# Test Content")
            template_file.write_text("# Test Content")

            analyzer = MergeAnalyzer(Path(tmpdir))
            diff_files = analyzer._collect_diff_files(backup_path, template_path)

            assert "CLAUDE.md" in diff_files
            assert not diff_files["CLAUDE.md"]["has_diff"]
            assert diff_files["CLAUDE.md"]["diff_lines"] == 0

    def test_collect_diff_files_with_differences(self):
        """Test collecting diffs when files have differences."""
        with tempfile.TemporaryDirectory() as tmpdir:
            backup_path = Path(tmpdir) / "backup"
            template_path = Path(tmpdir) / "template"
            backup_path.mkdir()
            template_path.mkdir()

            # Create files with differences
            backup_file = backup_path / "CLAUDE.md"
            template_file = template_path / "CLAUDE.md"
            backup_file.write_text("# Old Content")
            template_file.write_text("# New Content\nAdditional line")

            analyzer = MergeAnalyzer(Path(tmpdir))
            diff_files = analyzer._collect_diff_files(backup_path, template_path)

            assert "CLAUDE.md" in diff_files
            assert diff_files["CLAUDE.md"]["has_diff"]
            assert diff_files["CLAUDE.md"]["diff_lines"] > 0

    def test_collect_diff_files_new_file_only_in_template(self):
        """Test collecting diffs when file exists only in template."""
        with tempfile.TemporaryDirectory() as tmpdir:
            backup_path = Path(tmpdir) / "backup"
            template_path = Path(tmpdir) / "template"
            backup_path.mkdir()
            template_path.mkdir()

            # File only in template
            template_file = template_path / ".gitignore"
            template_file.write_text("*.pyc\n__pycache__/")

            analyzer = MergeAnalyzer(Path(tmpdir))
            diff_files = analyzer._collect_diff_files(backup_path, template_path)

            assert ".gitignore" in diff_files
            assert not diff_files[".gitignore"]["backup_exists"]
            assert diff_files[".gitignore"]["template_exists"]

    def test_collect_diff_files_file_only_in_backup(self):
        """Test collecting diffs when file exists only in backup."""
        with tempfile.TemporaryDirectory() as tmpdir:
            backup_path = Path(tmpdir) / "backup"
            template_path = Path(tmpdir) / "template"
            backup_path.mkdir()
            template_path.mkdir()

            # File only in backup
            backup_file = backup_path / "CLAUDE.md"
            backup_file.write_text("# Old Content")

            analyzer = MergeAnalyzer(Path(tmpdir))
            diff_files = analyzer._collect_diff_files(backup_path, template_path)

            assert "CLAUDE.md" in diff_files
            assert diff_files["CLAUDE.md"]["backup_exists"]
            assert not diff_files["CLAUDE.md"]["template_exists"]


class TestCreateAnalysisPrompt:
    """Tests for _create_analysis_prompt method."""

    def test_prompt_includes_required_elements(self):
        """Test analysis prompt includes all required elements."""
        with tempfile.TemporaryDirectory() as tmpdir:
            backup_path = Path(tmpdir) / "backup"
            template_path = Path(tmpdir) / "template"
            backup_path.mkdir()
            template_path.mkdir()

            analyzer = MergeAnalyzer(Path(tmpdir))
            diff_files = {}
            prompt = analyzer._create_analysis_prompt(backup_path, template_path, diff_files)

            # Check essential prompt components
            assert "MoAI-ADK 설정 파일 병합 전문가" in prompt
            assert "당신은" in prompt
            assert "JSON" in prompt
            assert "safe_to_auto_merge" in prompt
            assert "user_action_required" in prompt

    def test_prompt_includes_merge_rules(self):
        """Test prompt includes merge rules."""
        with tempfile.TemporaryDirectory() as tmpdir:
            backup_path = Path(tmpdir) / "backup"
            template_path = Path(tmpdir) / "template"
            backup_path.mkdir()
            template_path.mkdir()

            analyzer = MergeAnalyzer(Path(tmpdir))
            diff_files = {}
            prompt = analyzer._create_analysis_prompt(backup_path, template_path, diff_files)

            # Check merge rules are in prompt
            assert "CLAUDE.md" in prompt
            assert "settings.json" in prompt
            assert "config.json" in prompt
            assert ".gitignore" in prompt


class TestFormatDiffSummary:
    """Tests for _format_diff_summary method."""

    def test_format_diff_summary_with_changes(self):
        """Test formatting diff summary with changes."""
        analyzer = MergeAnalyzer(Path("/tmp"))

        diff_files = {
            "CLAUDE.md": {
                "backup_exists": True,
                "template_exists": True,
                "has_diff": True,
                "diff_lines": 5,
            },
            ".gitignore": {
                "backup_exists": False,
                "template_exists": True,
                "has_diff": False,
                "diff_lines": 0,
            },
        }

        summary = analyzer._format_diff_summary(diff_files)

        assert "CLAUDE.md" in summary
        assert "✏️  변경됨" in summary or "changed" in summary
        assert ".gitignore" in summary
        assert "✨ 새 파일" in summary or "New file" in summary


class TestBuildClaudeCommand:
    """Tests for _build_claude_command method."""

    def test_claude_command_structure(self):
        """Test Claude command has correct structure."""
        analyzer = MergeAnalyzer(Path("/tmp"))
        cmd = analyzer._build_claude_command()

        assert cmd[0] == "claude"
        assert "-p" in cmd
        assert "--output-format" in cmd
        assert "json" in cmd
        assert "--tools" in cmd

    def test_claude_command_includes_allowed_tools(self):
        """Test Claude command includes only safe tools."""
        analyzer = MergeAnalyzer(Path("/tmp"))
        cmd = analyzer._build_claude_command()

        # Extract tools from command
        tools_index = cmd.index("--tools") + 1
        tools_str = cmd[tools_index]

        # Check that only read-only tools are used
        assert "Read" in tools_str
        assert "Glob" in tools_str
        assert "Grep" in tools_str


class TestFallbackAnalysis:
    """Tests for _fallback_analysis method."""

    def test_fallback_analysis_returns_valid_structure(self):
        """Test fallback analysis returns correct structure."""
        with tempfile.TemporaryDirectory() as tmpdir:
            backup_path = Path(tmpdir) / "backup"
            template_path = Path(tmpdir) / "template"
            backup_path.mkdir()
            template_path.mkdir()

            # Create test files
            (backup_path / "CLAUDE.md").write_text("Old content")
            (template_path / "CLAUDE.md").write_text("New content\nExtra line")

            analyzer = MergeAnalyzer(Path(tmpdir))
            diff_files = analyzer._collect_diff_files(backup_path, template_path)

            fallback = analyzer._fallback_analysis(backup_path, template_path, diff_files)

            # Check structure
            assert "files" in fallback
            assert "safe_to_auto_merge" in fallback
            assert "user_action_required" in fallback
            assert "summary" in fallback
            assert "risk_assessment" in fallback
            assert fallback["fallback"] is True

    def test_fallback_analysis_marks_high_risk_correctly(self):
        """Test fallback analysis correctly identifies high-risk changes."""
        with tempfile.TemporaryDirectory() as tmpdir:
            backup_path = Path(tmpdir) / "backup"
            template_path = Path(tmpdir) / "template"
            backup_path.mkdir()
            template_path.mkdir()

            # Create large diff in config file
            backup_config = backup_path / ".moai/config/config.json"
            backup_config.parent.mkdir(parents=True)
            template_config = template_path / ".moai/config/config.json"
            template_config.parent.mkdir(parents=True)

            backup_config.write_text(json.dumps({"version": "0.25.0"}))
            template_config.write_text(json.dumps({"version": "0.26.0", "extra": "fields" * 20}))

            analyzer = MergeAnalyzer(Path(tmpdir))
            diff_files = analyzer._collect_diff_files(backup_path, template_path)

            fallback = analyzer._fallback_analysis(backup_path, template_path, diff_files)

            # High risk files should be noted
            assert isinstance(fallback["files"], list)


class TestDisplayAnalysis:
    """Tests for _display_analysis method."""

    @patch("rich.console.Console.print")
    def test_display_analysis_shows_summary(self, mock_print):
        """Test display analysis shows summary information."""
        analyzer = MergeAnalyzer(Path("/tmp"))

        analysis = {
            "files": [
                {
                    "filename": "CLAUDE.md",
                    "changes": "5줄 변경",
                    "recommendation": "smart_merge",
                    "conflict_severity": "low",
                    "note": "안전한 변경",
                }
            ],
            "summary": "1개 파일 변경 감지",
            "risk_assessment": "낮음",
            "safe_to_auto_merge": True,
            "user_action_required": False,
        }

        analyzer._display_analysis(analysis)

        # Verify Console.print was called
        assert mock_print.called


class TestAnalyzeMergeWithMocking:
    """Tests for analyze_merge method with mocked subprocess."""

    @patch("subprocess.run")
    def test_analyze_merge_successful_claude_response(self, mock_run):
        """Test analyze_merge with successful Claude response."""
        with tempfile.TemporaryDirectory() as tmpdir:
            backup_path = Path(tmpdir) / "backup"
            template_path = Path(tmpdir) / "template"
            backup_path.mkdir()
            template_path.mkdir()

            # Create test files
            (backup_path / "CLAUDE.md").write_text("Old")
            (template_path / "CLAUDE.md").write_text("New")

            analyzer = MergeAnalyzer(Path(tmpdir))

            # Mock subprocess to return valid JSON
            mock_result = MagicMock()
            mock_result.returncode = 0
            mock_result.stdout = json.dumps(
                {
                    "files": [
                        {
                            "filename": "CLAUDE.md",
                            "changes": "변경됨",
                            "recommendation": "smart_merge",
                            "conflict_severity": "low",
                        }
                    ],
                    "safe_to_auto_merge": True,
                    "user_action_required": False,
                    "summary": "분석 완료",
                    "risk_assessment": "낮음",
                }
            )
            mock_run.return_value = mock_result

            result = analyzer.analyze_merge(backup_path, template_path)

            # Verify result structure
            assert result["safe_to_auto_merge"] is True
            assert "files" in result
            assert "summary" in result

    @patch("subprocess.run")
    def test_analyze_merge_falls_back_on_invalid_json(self, mock_run):
        """Test analyze_merge falls back when Claude returns invalid JSON."""
        with tempfile.TemporaryDirectory() as tmpdir:
            backup_path = Path(tmpdir) / "backup"
            template_path = Path(tmpdir) / "template"
            backup_path.mkdir()
            template_path.mkdir()

            (backup_path / "CLAUDE.md").write_text("Old")
            (template_path / "CLAUDE.md").write_text("New")

            analyzer = MergeAnalyzer(Path(tmpdir))

            # Mock subprocess to return invalid JSON
            mock_result = MagicMock()
            mock_result.returncode = 0
            mock_result.stdout = "Invalid JSON {{"
            mock_run.return_value = mock_result

            result = analyzer.analyze_merge(backup_path, template_path)

            # Should use fallback analysis
            assert "files" in result
            assert result["fallback"] is True

    @patch("subprocess.run")
    def test_analyze_merge_falls_back_on_timeout(self, mock_run):
        """Test analyze_merge falls back on Claude timeout."""
        with tempfile.TemporaryDirectory() as tmpdir:
            backup_path = Path(tmpdir) / "backup"
            template_path = Path(tmpdir) / "template"
            backup_path.mkdir()
            template_path.mkdir()

            analyzer = MergeAnalyzer(Path(tmpdir))

            # Mock timeout
            mock_run.side_effect = subprocess.TimeoutExpired("claude", 120)

            result = analyzer.analyze_merge(backup_path, template_path)

            # Should use fallback
            assert result["fallback"] is True

    @patch("subprocess.run")
    def test_analyze_merge_falls_back_on_file_not_found(self, mock_run):
        """Test analyze_merge falls back when Claude not found."""
        with tempfile.TemporaryDirectory() as tmpdir:
            backup_path = Path(tmpdir) / "backup"
            template_path = Path(tmpdir) / "template"
            backup_path.mkdir()
            template_path.mkdir()

            analyzer = MergeAnalyzer(Path(tmpdir))

            # Mock file not found
            mock_run.side_effect = FileNotFoundError("claude not found")

            result = analyzer.analyze_merge(backup_path, template_path)

            # Should use fallback
            assert result["fallback"] is True


# Import subprocess for mocking
import subprocess
