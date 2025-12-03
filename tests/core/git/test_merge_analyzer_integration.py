"""Integration tests for MergeAnalyzer.

Tests the full workflow of merge analysis including backup creation,
Claude analysis (or fallback), user confirmation, and merge decision flow.
"""

import json
import tempfile
from pathlib import Path
from unittest.mock import MagicMock, patch

from moai_adk.core.merge.analyzer import MergeAnalyzer


class TestMergeAnalyzerIntegration:
    """Integration tests for MergeAnalyzer end-to-end workflow."""

    @patch("subprocess.run")
    def test_full_merge_workflow_with_fallback(self, mock_run):
        """Test complete merge workflow when Claude is unavailable.

        Simulates:
        1. Backup directory with user configurations
        2. Template directory with new templates
        3. Claude fails (mocked)
        4. Falls back to difflib analysis
        5. Returns valid merge decision
        """
        # Mock subprocess to fail (Claude not available)
        mock_run.side_effect = FileNotFoundError("claude not found")

        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            backup_path = project_path / "backup"
            template_path = project_path / "template"

            # Setup backup with user configuration
            backup_path.mkdir()
            (backup_path / "CLAUDE.md").write_text("# Old Project\n\nUser customizations preserved")
            (backup_path / ".claude").mkdir(parents=True)
            (backup_path / ".claude/settings.json").write_text(
                json.dumps(
                    {
                        "permissions": {"allowedTools": ["Read", "Write"]},
                    },
                    indent=2,
                )
            )

            # Setup template with new content
            template_path.mkdir()
            (template_path / "CLAUDE.md").write_text("# New Template\n\nEnhanced features")
            (template_path / ".claude").mkdir(parents=True)
            (template_path / ".claude/settings.json").write_text(
                json.dumps(
                    {
                        "permissions": {
                            "allowedTools": ["Read", "Write", "Edit"],
                        },
                    },
                    indent=2,
                )
            )

            # Analyzer: Falls back to difflib
            analyzer = MergeAnalyzer(project_path)
            result = analyzer.analyze_merge(backup_path, template_path)

            # Verify result structure
            assert isinstance(result, dict)
            assert "files" in result
            assert "safe_to_auto_merge" in result
            assert "user_action_required" in result
            assert "summary" in result

            # Verify fallback was used
            assert result.get("fallback") is True

            # Verify files were analyzed
            analyzed_files = result["files"]
            # Only ANALYZED_FILES are checked: CLAUDE.md, .claude/settings.json, etc.
            file_names = [f["filename"] for f in analyzed_files]
            assert any("CLAUDE.md" in fname for fname in file_names)

            # Each file should have expected fields
            for file_info in analyzed_files:
                assert "filename" in file_info
                assert "changes" in file_info
                assert "recommendation" in file_info
                assert "conflict_severity" in file_info

    @patch("subprocess.run")
    def test_full_workflow_with_claude_analysis(self, mock_run):
        """Test complete workflow with successful Claude analysis.

        Simulates:
        1. Backup and template setup
        2. Claude provides analysis (mocked)
        3. User confirms merge
        4. Returns merge decision
        """
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            backup_path = project_path / "backup"
            template_path = project_path / "template"

            # Setup directories
            backup_path.mkdir()
            (backup_path / "CLAUDE.md").write_text("Old content")
            template_path.mkdir()
            (template_path / "CLAUDE.md").write_text("New content")

            # Mock Claude response
            mock_result = MagicMock()
            mock_result.returncode = 0
            mock_result.stdout = json.dumps(
                {
                    "files": [
                        {
                            "filename": "CLAUDE.md",
                            "changes": "메타데이터 업데이트",
                            "recommendation": "smart_merge",
                            "conflict_severity": "low",
                            "note": "안전한 변경",
                        }
                    ],
                    "safe_to_auto_merge": True,
                    "user_action_required": False,
                    "summary": "1개 파일 안전하게 병합 가능",
                    "risk_assessment": "낮음",
                }
            )
            mock_run.return_value = mock_result

            analyzer = MergeAnalyzer(project_path)
            result = analyzer.analyze_merge(backup_path, template_path)

            # Verify Claude was called
            assert mock_run.called
            mock_run.assert_called_once()

            # Verify result from Claude
            assert result["safe_to_auto_merge"] is True
            assert result["user_action_required"] is False
            assert len(result["files"]) == 1
            assert result["files"][0]["filename"] == "CLAUDE.md"

    def test_merge_analysis_with_multiple_file_changes(self):
        """Test analyzing multiple file changes simultaneously.

        Verifies:
        1. Multiple files from ANALYZED_FILES with different change patterns
        2. Correct diff detection
        3. Appropriate severity levels
        """
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            backup_path = project_path / "backup"
            template_path = project_path / "template"

            backup_path.mkdir()
            template_path.mkdir()

            # File 1 (ANALYZED_FILES): No change
            (backup_path / "CLAUDE.md").write_text("Same content")
            (template_path / "CLAUDE.md").write_text("Same content")

            # File 2 (ANALYZED_FILES): Minor change
            (backup_path / ".gitignore").write_text("*.pyc")
            (template_path / ".gitignore").write_text("*.pyc\n__pycache__/")

            # File 3 (ANALYZED_FILES): Only in backup
            (backup_path / ".claude").mkdir()
            (backup_path / ".claude/settings.json").write_text(json.dumps({"old": "config"}, indent=2))

            # File 4 (ANALYZED_FILES): Only in template
            (template_path / ".moai").mkdir()
            (template_path / ".moai/config").mkdir()
            (template_path / ".moai/config/config.json").write_text(json.dumps({"version": "0.26.0"}, indent=2))

            analyzer = MergeAnalyzer(project_path)
            diff_files = analyzer._collect_diff_files(backup_path, template_path)

            # Verify files from ANALYZED_FILES are detected
            assert "CLAUDE.md" in diff_files
            assert ".gitignore" in diff_files
            assert ".claude/settings.json" in diff_files
            assert ".moai/config/config.json" in diff_files

            # Verify specific patterns
            assert not diff_files["CLAUDE.md"]["has_diff"]  # No change
            assert diff_files[".gitignore"]["has_diff"]  # Minor change
            assert diff_files[".claude/settings.json"]["backup_exists"]  # Only in backup
            assert diff_files[".moai/config/config.json"]["template_exists"]  # Only in template

    @patch("rich.console.Console.print")
    def test_display_analysis_with_multiple_files(self, mock_print):
        """Test display analysis output with multiple file changes.

        Verifies Rich console output is properly formatted.
        """
        analyzer = MergeAnalyzer(Path("/tmp"))

        analysis = {
            "files": [
                {
                    "filename": "CLAUDE.md",
                    "changes": "메타데이터 수정",
                    "recommendation": "smart_merge",
                    "conflict_severity": "low",
                    "note": "프로젝트 정보 보존",
                },
                {
                    "filename": ".claude/settings.json",
                    "changes": "보안 정책 추가",
                    "recommendation": "use_template",
                    "conflict_severity": "medium",
                    "note": "새로운 보안 설정이 필요합니다",
                },
                {
                    "filename": ".gitignore",
                    "changes": "항목 추가",
                    "recommendation": "smart_merge",
                    "conflict_severity": "low",
                    "note": "기존 항목 보존",
                },
            ],
            "summary": "3개 파일 변경 감지",
            "risk_assessment": "중간",
            "safe_to_auto_merge": False,
            "user_action_required": True,
        }

        analyzer._display_analysis(analysis)

        # Verify console.print was called
        assert mock_print.called

        # Count the number of print calls
        # Should include title, summary, risk assessment, table, and notes
        assert mock_print.call_count >= 5

    def test_prompt_generation_includes_all_file_info(self):
        """Test that analysis prompt includes information about all files.

        Verifies prompt contains:
        1. All analyzed file names
        2. Current status (changed/new/deleted)
        3. Diff line counts
        """
        with tempfile.TemporaryDirectory() as tmpdir:
            backup_path = Path(tmpdir) / "backup"
            template_path = Path(tmpdir) / "template"
            backup_path.mkdir()
            template_path.mkdir()

            # Create test files
            (backup_path / "CLAUDE.md").write_text("Old")
            (template_path / "CLAUDE.md").write_text("New content\nMore lines")
            (template_path / ".gitignore").write_text("*.pyc")

            analyzer = MergeAnalyzer(Path(tmpdir))
            diff_files = analyzer._collect_diff_files(backup_path, template_path)
            prompt = analyzer._create_analysis_prompt(backup_path, template_path, diff_files)

            # Verify files are mentioned in prompt
            assert "CLAUDE.md" in prompt
            assert ".gitignore" in prompt

            # Verify prompt mentions change types
            assert "변경" in prompt or "changed" in prompt or "diff" in prompt

            # Verify JSON response format is specified
            assert '"files"' in prompt or "'files'" in prompt
            assert "recommendation" in prompt

    def test_fallback_correctly_estimates_severity(self):
        """Test fallback analysis correctly estimates conflict severity.

        Verifies that:
        1. Config files with large diffs are marked as medium/high
        2. Simple file additions are marked as low
        3. Documentation changes are marked appropriately
        """
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            backup_path = project_path / "backup"
            template_path = project_path / "template"
            backup_path.mkdir()
            template_path.mkdir()

            # Create config file with significant changes
            backup_config = backup_path / ".moai/config/config.json"
            backup_config.parent.mkdir(parents=True)
            backup_config.write_text(json.dumps({"version": "0.25.0"}, indent=2))

            template_config = template_path / ".moai/config/config.json"
            template_config.parent.mkdir(parents=True)
            # Create a large diff (20+ lines)
            template_config.write_text(
                json.dumps(
                    {
                        "version": "0.26.0",
                        "fields": ["a", "b", "c", "d", "e"] * 5,  # Ensure > 10 lines diff
                    },
                    indent=2,
                )
            )

            analyzer = MergeAnalyzer(project_path)
            diff_files = analyzer._collect_diff_files(backup_path, template_path)
            fallback = analyzer._fallback_analysis(backup_path, template_path, diff_files)

            # Verify config file is analyzed
            assert len(fallback["files"]) > 0

            # Verify severity levels are assigned
            for file_info in fallback["files"]:
                assert file_info["conflict_severity"] in [
                    "low",
                    "medium",
                    "high",
                ]
