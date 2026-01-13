"""Additional coverage tests for template merger.

Tests for lines not covered by existing tests.
"""

import json
from pathlib import Path
from unittest.mock import MagicMock, patch

from moai_adk.core.template.merger import TemplateMerger


class TestTemplateMergerMergeClaudeWithProjectInfo:
    """Test merge_claude_md with project info in template."""

    def test_merge_claude_md_removes_template_project_section(self, tmp_path):
        """Should remove project info section from template when merging."""
        merger = TemplateMerger(tmp_path)

        # Create template with project info
        template_content = """# Some content

## Project Information

Name: Test
Description: Test project
"""
        template_path = tmp_path / "CLAUDE.md.template"
        template_path.write_text(template_content)

        # Create existing file with project info
        existing_path = tmp_path / "CLAUDE.md"
        existing_content = """# Existing content

## Project Information

Name: Existing Project
Description: My project
"""
        existing_path.write_text(existing_content)

        # Merge - should preserve existing project info and use template content
        merger.merge_claude_md(template_path, existing_path)

        # Verify template section was removed and existing project info was preserved
        result = existing_path.read_text()
        assert "## Project Information" in result
        assert "Existing Project" in result
        assert "Test project" not in result  # Template project info should be removed


class TestTemplateMergerMergeConfigYamlImport:
    """Test merge_config yaml import handling."""

    def test_merge_config_yaml_import_error_fallback(self, tmp_path):
        """Should fallback to JSON when yaml import fails."""
        merger = TemplateMerger(tmp_path)

        # Create legacy config.json
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        config_path = config_dir / "config.json"
        config_data = {"projectName": "test", "mode": "personal"}
        config_path.write_text(json.dumps(config_data))

        # Mock yaml import to fail
        with patch("builtins.__import__", side_effect=ImportError("No yaml module")):
            result = merger.merge_config()

        # Should still return config from JSON fallback
        assert result["projectName"] == "test"

    def test_merge_config_loads_section_yaml_files(self, tmp_path):
        """Should load configuration from section YAML files."""
        merger = TemplateMerger(tmp_path)

        # Create section YAML files
        sections_dir = tmp_path / ".moai" / "config" / "sections"
        sections_dir.mkdir(parents=True)

        (sections_dir / "project.yaml").write_text("name: test_project")
        (sections_dir / "language.yaml").write_text("conversation_language: en")

        result = merger.merge_config()

        # Should load from section files (target_path.name is used as fallback)
        assert result.get("projectName") == tmp_path.name

    def test_merge_config_skips_invalid_yaml_sections(self, tmp_path):
        """Should skip section files that fail to parse."""
        merger = TemplateMerger(tmp_path)

        # Create section YAML files
        sections_dir = tmp_path / ".moai" / "config" / "sections"
        sections_dir.mkdir(parents=True)

        (sections_dir / "project.yaml").write_text("name: valid_project")
        (sections_dir / "invalid.yaml").write_text("{{{ invalid yaml content")

        result = merger.merge_config()

        # Should load valid sections and skip invalid ones (target_path.name is fallback)
        assert result.get("projectName") == tmp_path.name


class TestTemplateMergerMergeSettingsJsonBackupPath:
    """Test merge_settings_json with backup path."""

    def test_merge_settings_json_uses_backup_when_exists(self, tmp_path):
        """Should use backup data when backup_path exists."""
        merger = TemplateMerger(tmp_path)

        # Create template
        template_path = tmp_path / "settings.json.template"
        template_data = {
            "env": {"VAR1": "template_value"},
            "permissions": {"allow": ["tool1"]},
        }
        template_path.write_text(json.dumps(template_data))

        # Create backup
        backup_path = tmp_path / "settings.json.backup"
        backup_data = {
            "env": {"VAR2": "backup_value"},
            "outputStyle": "compact",
        }
        backup_path.write_text(json.dumps(backup_data))

        # Create existing
        existing_path = tmp_path / "settings.json"
        existing_path.write_text("{}")

        merger.merge_settings_json(template_path, existing_path, backup_path)

        # Verify backup data was used
        result = json.loads(existing_path.read_text())
        assert result["env"]["VAR2"] == "backup_value"
        assert result["outputStyle"] == "compact"


class TestTemplateMergerMergeSettingsJsonPreserveFields:
    """Test merge_settings_json field preservation."""

    def test_merge_settings_json_preserves_output_style(self, tmp_path):
        """Should preserve outputStyle from user data."""
        merger = TemplateMerger(tmp_path)

        # Create template
        template_path = tmp_path / "settings.json.template"
        template_data = {"permissions": {"allow": ["tool1"]}}
        template_path.write_text(json.dumps(template_data))

        # Create existing with outputStyle
        existing_path = tmp_path / "settings.json"
        existing_data = {"outputStyle": "yoda"}
        existing_path.write_text(json.dumps(existing_data))

        merger.merge_settings_json(template_path, existing_path)

        # Verify outputStyle was preserved
        result = json.loads(existing_path.read_text())
        assert result["outputStyle"] == "yoda"

    def test_merge_settings_json_preserves_spinner_tips(self, tmp_path):
        """Should preserve spinnerTipsEnabled from user data."""
        merger = TemplateMerger(tmp_path)

        # Create template
        template_path = tmp_path / "settings.json.template"
        template_data = {"permissions": {"allow": ["tool1"]}}
        template_path.write_text(json.dumps(template_data))

        # Create existing with spinnerTipsEnabled
        existing_path = tmp_path / "settings.json"
        existing_data = {"spinnerTipsEnabled": False}
        existing_path.write_text(json.dumps(existing_data))

        merger.merge_settings_json(template_path, existing_path)

        # Verify spinnerTipsEnabled was preserved
        result = json.loads(existing_path.read_text())
        assert result["spinnerTipsEnabled"] is False


class TestTemplateMergerMergeGithubWorkflowsUserWorkflows:
    """Test merge_github_workflows user workflow preservation."""

    def test_merge_github_workflows_tracks_user_workflows(self, tmp_path):
        """Should track existing user workflows for preservation."""
        merger = TemplateMerger(tmp_path)

        # Create template workflows
        template_dir = tmp_path / "template_github"
        template_dir.mkdir(parents=True)
        workflows_dir = template_dir / "workflows"
        workflows_dir.mkdir(parents=True)
        (workflows_dir / "moai-update.yml").write_text("name: Update")

        # Create existing workflows with user workflow
        existing_dir = tmp_path / "existing_github"
        existing_dir.mkdir(parents=True)
        existing_workflows = existing_dir / "workflows"
        existing_workflows.mkdir(parents=True)
        (existing_workflows / "my-custom-workflow.yml").write_text("name: Custom")

        merger.merge_github_workflows(template_dir, existing_dir)

        # Verify user workflow is preserved
        assert (existing_workflows / "my-custom-workflow.yml").exists()


class TestTemplateMergerMergeGithubWorkflowsNonWorkflowFiles:
    """Test merge_github_workflows with non-workflow files."""

    def test_merge_github_workflows_copies_non_workflow_files(self, tmp_path):
        """Should copy non-workflow files and create parent directories."""
        merger = TemplateMerger(tmp_path)

        # Create template with non-workflow file
        template_dir = tmp_path / "template_github"
        template_dir.mkdir(parents=True)
        issue_template_dir = template_dir / "ISSUE_TEMPLATE"
        issue_template_dir.mkdir(parents=True)
        (issue_template_dir / "bug_report.md").write_text("Bug Report")

        # Create existing directory
        existing_dir = tmp_path / "existing_github"
        existing_dir.mkdir(parents=True)

        merger.merge_github_workflows(template_dir, existing_dir)

        # Verify file was copied with parent directory created
        dst_file = existing_dir / "ISSUE_TEMPLATE" / "bug_report.md"
        assert dst_file.exists()
        assert dst_file.read_text() == "Bug Report"
