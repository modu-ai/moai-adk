"""Comprehensive coverage tests for Pure Python MergeAnalyzer module.

Tests MergeAnalyzer class for merge conflict detection, analysis, and recommendations.
Target: 70%+ code coverage with actual code path execution and mocked dependencies.
"""

import json
from pathlib import Path
from unittest.mock import patch

import pytest
import yaml


class TestMergeAnalyzerInit:
    """Test MergeAnalyzer initialization."""

    def test_merge_analyzer_instantiation(self):
        """Should instantiate with project path."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        project_path = Path("/tmp/test_project")
        analyzer = MergeAnalyzer(project_path)

        assert analyzer.project_path == project_path

    def test_merge_analyzer_constants(self):
        """Should have correct constants defined."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        assert len(MergeAnalyzer.ANALYZED_FILES) > 0
        assert "CLAUDE.md" in MergeAnalyzer.ANALYZED_FILES
        assert ".claude/settings.json" in MergeAnalyzer.ANALYZED_FILES
        assert ".moai/config/config.yaml" in MergeAnalyzer.ANALYZED_FILES
        assert ".gitignore" in MergeAnalyzer.ANALYZED_FILES

    def test_risk_factors_defined(self):
        """Should have risk factors defined."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        assert "user_section_modified" in MergeAnalyzer.RISK_FACTORS
        assert "permission_change" in MergeAnalyzer.RISK_FACTORS
        assert "env_variable_removed" in MergeAnalyzer.RISK_FACTORS


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

        (template_path / ".claude").mkdir(parents=True)
        (template_path / ".claude" / "settings.json").write_text("{}")

        diff_files = analyzer._collect_diff_files(backup_path, template_path)

        assert len(diff_files) > 0


class TestAnalyzeClaudeMd:
    """Test _analyze_claude_md method."""

    def test_analyze_claude_md_new_file(self, tmp_path):
        """Should handle new CLAUDE.md file."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(tmp_path)
        info = {
            "backup_exists": False,
            "template_exists": True,
            "has_diff": True,
            "diff_lines": 100,
            "backup_content": None,
            "template_content": "# CLAUDE.md content",
        }

        result = analyzer._analyze_claude_md("CLAUDE.md", info)

        assert result["recommendation"] == "use_template"
        assert result["conflict_severity"] == "low"

    def test_analyze_claude_md_with_user_sections(self, tmp_path):
        """Should detect user-customized sections."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(tmp_path)
        info = {
            "backup_exists": True,
            "template_exists": True,
            "has_diff": True,
            "diff_lines": 50,
            "backup_content": "# Project Information\nCustom content here",
            "template_content": "# New Content\nTemplate content",
        }

        result = analyzer._analyze_claude_md("CLAUDE.md", info)

        assert result["recommendation"] == "smart_merge"
        assert "user_customizations" in result


class TestAnalyzeSettingsJson:
    """Test _analyze_settings_json method."""

    def test_analyze_settings_json_new_file(self, tmp_path):
        """Should handle new settings.json file."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(tmp_path)
        info = {
            "backup_exists": False,
            "template_exists": True,
            "has_diff": True,
            "diff_lines": 20,
            "backup_content": None,
            "template_content": "{}",
        }

        result = analyzer._analyze_settings_json(".claude/settings.json", info)

        assert result["recommendation"] == "use_template"
        assert result["conflict_severity"] == "low"

    def test_analyze_settings_json_env_removed(self, tmp_path):
        """Should detect removed env variables."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(tmp_path)
        backup_json = {"env": {"MY_VAR": "value", "REMOVED_VAR": "old"}}
        template_json = {"env": {"MY_VAR": "value"}}

        info = {
            "backup_exists": True,
            "template_exists": True,
            "has_diff": True,
            "diff_lines": 5,
            "backup_content": json.dumps(backup_json),
            "template_content": json.dumps(template_json),
        }

        result = analyzer._analyze_settings_json(".claude/settings.json", info)

        assert result["recommendation"] == "smart_merge"
        assert any("Env vars removed" in c for c in result.get("critical_changes", []))

    def test_analyze_settings_json_permission_change(self, tmp_path):
        """Should detect permission changes."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(tmp_path)
        backup_json = {"permissions": {"deny": ["Bash"]}}
        template_json = {"permissions": {"deny": ["Bash", "Write"]}}

        info = {
            "backup_exists": True,
            "template_exists": True,
            "has_diff": True,
            "diff_lines": 5,
            "backup_content": json.dumps(backup_json),
            "template_content": json.dumps(template_json),
        }

        result = analyzer._analyze_settings_json(".claude/settings.json", info)

        assert result["recommendation"] == "smart_merge"

    def test_analyze_settings_json_parse_error(self, tmp_path):
        """Should handle JSON parse errors."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(tmp_path)
        info = {
            "backup_exists": True,
            "template_exists": True,
            "has_diff": True,
            "diff_lines": 5,
            "backup_content": "not valid json",
            "template_content": "{}",
        }

        result = analyzer._analyze_settings_json(".claude/settings.json", info)

        assert result["conflict_severity"] == "high"


class TestAnalyzeConfigYaml:
    """Test _analyze_config_yaml method."""

    def test_analyze_config_yaml_new_file(self, tmp_path):
        """Should handle new config.yaml file."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(tmp_path)
        info = {
            "backup_exists": False,
            "template_exists": True,
            "has_diff": True,
            "diff_lines": 20,
            "backup_content": None,
            "template_content": "version: 1.0",
        }

        result = analyzer._analyze_config_yaml(".moai/config/config.yaml", info)

        assert result["recommendation"] == "use_template"
        assert result["conflict_severity"] == "low"

    def test_analyze_config_yaml_user_settings(self, tmp_path):
        """Should detect user settings changes."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(tmp_path)
        backup_yaml = {"user": {"name": "CustomUser"}, "version": "1.0"}
        template_yaml = {"user": {"name": "DefaultUser"}, "version": "1.1"}

        info = {
            "backup_exists": True,
            "template_exists": True,
            "has_diff": True,
            "diff_lines": 10,
            "backup_content": yaml.dump(backup_yaml),
            "template_content": yaml.dump(template_yaml),
        }

        result = analyzer._analyze_config_yaml(".moai/config/config.yaml", info)

        assert result["recommendation"] == "smart_merge"
        assert "user" in result.get("user_customizations", [])

    def test_analyze_config_yaml_parse_error(self, tmp_path):
        """Should handle YAML parse errors."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(tmp_path)
        info = {
            "backup_exists": True,
            "template_exists": True,
            "has_diff": True,
            "diff_lines": 5,
            "backup_content": "invalid: yaml: content:",
            "template_content": "version: 1.0",
        }

        result = analyzer._analyze_config_yaml(".moai/config/config.yaml", info)

        # Should still return a result even with parse error
        assert "conflict_severity" in result


class TestAnalyzeGitignore:
    """Test _analyze_gitignore method."""

    def test_analyze_gitignore_new_file(self, tmp_path):
        """Should handle new .gitignore file."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(tmp_path)
        info = {
            "backup_exists": False,
            "template_exists": True,
            "has_diff": True,
            "diff_lines": 10,
            "backup_content": None,
            "template_content": "*.pyc\n__pycache__/",
        }

        result = analyzer._analyze_gitignore(".gitignore", info)

        assert result["recommendation"] == "use_template"
        assert result["conflict_severity"] == "low"

    def test_analyze_gitignore_user_entries(self, tmp_path):
        """Should detect user entries to preserve."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(tmp_path)
        info = {
            "backup_exists": True,
            "template_exists": True,
            "has_diff": True,
            "diff_lines": 5,
            "backup_content": "*.pyc\n__pycache__/\nmy_custom_dir/",
            "template_content": "*.pyc\n__pycache__/\nbuild/",
        }

        result = analyzer._analyze_gitignore(".gitignore", info)

        assert result["recommendation"] == "smart_merge"
        assert "my_custom_dir/" in result.get("user_entries", [])
        assert "build/" in result.get("new_entries", [])


class TestAnalyzeMergeIntegration:
    """Test analyze_merge integration."""

    def test_analyze_merge_success(self, tmp_path):
        """Should perform semantic analysis."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(tmp_path)
        backup_path = tmp_path / "backup"
        template_path = tmp_path / "template"

        backup_path.mkdir()
        template_path.mkdir()

        # Create test files
        (backup_path / "CLAUDE.md").write_text("# Old content")
        (template_path / "CLAUDE.md").write_text("# New content")

        result = analyzer.analyze_merge(backup_path, template_path)

        assert isinstance(result, dict)
        assert "safe_to_auto_merge" in result
        assert "files" in result
        assert result["analysis_method"] == "pure_python_semantic"

    def test_analyze_merge_no_changes(self, tmp_path):
        """Should handle no changes gracefully."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(tmp_path)
        backup_path = tmp_path / "backup"
        template_path = tmp_path / "template"

        backup_path.mkdir()
        template_path.mkdir()

        # Create identical files
        content = "# Same content"
        (backup_path / "CLAUDE.md").write_text(content)
        (template_path / "CLAUDE.md").write_text(content)

        result = analyzer.analyze_merge(backup_path, template_path)

        assert result["safe_to_auto_merge"] is True


class TestRiskCalculation:
    """Test risk calculation methods."""

    def test_risk_score_to_severity_low(self, tmp_path):
        """Should return low severity for low scores."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(tmp_path)

        assert analyzer._risk_score_to_severity(0) == "low"
        assert analyzer._risk_score_to_severity(1) == "low"
        assert analyzer._risk_score_to_severity(2) == "low"

    def test_risk_score_to_severity_medium(self, tmp_path):
        """Should return medium severity for medium scores."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(tmp_path)

        assert analyzer._risk_score_to_severity(3) == "medium"
        assert analyzer._risk_score_to_severity(4) == "medium"
        assert analyzer._risk_score_to_severity(5) == "medium"

    def test_risk_score_to_severity_high(self, tmp_path):
        """Should return high severity for high scores."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(tmp_path)

        assert analyzer._risk_score_to_severity(6) == "high"
        assert analyzer._risk_score_to_severity(10) == "high"

    def test_calculate_risk_level(self, tmp_path):
        """Should calculate overall risk level."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(tmp_path)

        assert analyzer._calculate_risk_level(0) == "low"
        assert analyzer._calculate_risk_level(3) == "low"
        assert analyzer._calculate_risk_level(5) == "medium"
        assert analyzer._calculate_risk_level(8) == "medium"
        assert analyzer._calculate_risk_level(9) == "high"


class TestFlattenKeys:
    """Test _flatten_keys helper method."""

    def test_flatten_keys_simple(self, tmp_path):
        """Should flatten simple dictionary."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(tmp_path)
        d = {"a": 1, "b": 2}

        keys = analyzer._flatten_keys(d)

        assert "a" in keys
        assert "b" in keys

    def test_flatten_keys_nested(self, tmp_path):
        """Should flatten nested dictionary."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(tmp_path)
        d = {"a": {"b": {"c": 1}}}

        keys = analyzer._flatten_keys(d)

        assert "a" in keys
        assert "a.b" in keys
        assert "a.b.c" in keys


class TestDisplayAnalysis:
    """Test _display_analysis method."""

    @patch("moai_adk.core.merge.analyzer.console")
    def test_display_analysis_with_files(self, mock_console, tmp_path):
        """Should display analysis results."""
        from moai_adk.core.merge.analyzer import MergeAnalyzer

        analyzer = MergeAnalyzer(tmp_path)
        analysis = {
            "summary": "Test summary",
            "risk_assessment": "Low",
            "analysis_method": "pure_python_semantic",
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
        analysis = {"summary": "No changes", "files": [], "analysis_method": "pure_python_semantic"}

        analyzer._display_analysis(analysis)

        mock_console.print.assert_called()


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
