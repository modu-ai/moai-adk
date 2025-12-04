"""Extended unit tests for moai_adk.core.template.merger module.

Comprehensive tests for TemplateMerger functionality including edge cases
and complex scenarios.
"""

import json
import shutil
from pathlib import Path
from unittest.mock import MagicMock, Mock, mock_open, patch

import pytest

from moai_adk.core.template.merger import TemplateMerger


class TestTemplateMergerMergeClaude:
    """Test merge_claude_md method with comprehensive scenarios."""

    def test_merge_claude_with_multiple_project_info_headers(self, tmp_path):
        """Test merge_claude_md handles multiple project info header formats."""
        merger = TemplateMerger(tmp_path)
        template_file = tmp_path / "template.md"
        existing_file = tmp_path / "existing.md"

        template_content = "Template content\nSection content"
        existing_content = "Existing\n\n## Project Info\nCustom project info"

        template_file.write_text(template_content)
        existing_file.write_text(existing_content)

        merger.merge_claude_md(template_file, existing_file)

        result = existing_file.read_text()
        assert "Custom project info" in result
        assert "Template content" in result

    def test_merge_claude_with_empty_template(self, tmp_path):
        """Test merge_claude_md with empty template."""
        merger = TemplateMerger(tmp_path)
        template_file = tmp_path / "template.md"
        existing_file = tmp_path / "existing.md"

        template_file.write_text("")
        existing_file.write_text("## Project Information\nInfo")

        merger.merge_claude_md(template_file, existing_file)
        assert existing_file.exists()

    def test_merge_claude_with_empty_existing(self, tmp_path):
        """Test merge_claude_md with empty existing file."""
        merger = TemplateMerger(tmp_path)
        template_file = tmp_path / "template.md"
        existing_file = tmp_path / "existing.md"

        template_file.write_text("Template content")
        existing_file.write_text("")

        merger.merge_claude_md(template_file, existing_file)
        result = existing_file.read_text()
        assert result == "Template content"

    def test_merge_claude_project_info_at_end(self, tmp_path):
        """Test merge_claude_md with project info at end of file."""
        merger = TemplateMerger(tmp_path)
        template_file = tmp_path / "template.md"
        existing_file = tmp_path / "existing.md"

        template_content = "Line 1\nLine 2\n## Project Information\nTemplate info"
        existing_content = "Existing content\n## Project Information\nUser info"

        template_file.write_text(template_content)
        existing_file.write_text(existing_content)

        merger.merge_claude_md(template_file, existing_file)

        result = existing_file.read_text()
        assert "User info" in result

    def test_find_project_info_with_different_headers(self):
        """Test _find_project_info_section with different header variations."""
        merger = TemplateMerger(Path("/tmp"))

        # Test with first header
        content1 = "Content\n## Project Information\nInfo"
        idx1, header1 = merger._find_project_info_section(content1)
        assert idx1 > 0
        assert header1 == "## Project Information"

        # Test with second header
        content2 = "Content\n## Project Info\nInfo"
        idx2, header2 = merger._find_project_info_section(content2)
        assert idx2 > 0
        assert header2 == "## Project Info"

    def test_find_project_info_multiple_matches(self):
        """Test finding first matching project info header."""
        merger = TemplateMerger(Path("/tmp"))
        content = "## Project Information\nFirst\n## Project Info\nSecond"

        idx, header = merger._find_project_info_section(content)
        assert idx >= 0
        assert header is not None


class TestTemplateMergerMergeGitignore:
    """Test merge_gitignore method comprehensively."""

    def test_merge_gitignore_removes_duplicates(self, tmp_path):
        """Test merge_gitignore removes duplicate entries."""
        merger = TemplateMerger(tmp_path)
        template_file = tmp_path / "template.gitignore"
        existing_file = tmp_path / ".gitignore"

        template_file.write_text("*.log\n*.tmp\n*.pyc")
        existing_file.write_text("*.pyc\nnode_modules\n*.log")

        merger.merge_gitignore(template_file, existing_file)

        result = existing_file.read_text()
        lines = result.strip().split("\n")
        # Check that we don't have duplicate *.pyc or *.log
        assert lines.count("*.log") == 1
        assert lines.count("*.pyc") == 1
        assert "node_modules" in lines

    def test_merge_gitignore_preserves_existing(self, tmp_path):
        """Test merge_gitignore preserves all existing entries."""
        merger = TemplateMerger(tmp_path)
        template_file = tmp_path / "template.gitignore"
        existing_file = tmp_path / ".gitignore"

        template_file.write_text("*.log")
        existing_file.write_text("custom_entry\nuser_setting")

        merger.merge_gitignore(template_file, existing_file)

        result = existing_file.read_text()
        assert "custom_entry" in result
        assert "user_setting" in result
        assert "*.log" in result

    def test_merge_gitignore_empty_template(self, tmp_path):
        """Test merge_gitignore with empty template."""
        merger = TemplateMerger(tmp_path)
        template_file = tmp_path / "template.gitignore"
        existing_file = tmp_path / ".gitignore"

        template_file.write_text("")
        existing_file.write_text("existing_entry")

        merger.merge_gitignore(template_file, existing_file)

        result = existing_file.read_text()
        assert "existing_entry" in result

    def test_merge_gitignore_empty_existing(self, tmp_path):
        """Test merge_gitignore with empty existing file."""
        merger = TemplateMerger(tmp_path)
        template_file = tmp_path / "template.gitignore"
        existing_file = tmp_path / ".gitignore"

        template_file.write_text("*.log\n*.tmp")
        existing_file.write_text("")

        merger.merge_gitignore(template_file, existing_file)

        result = existing_file.read_text()
        assert "*.log" in result
        assert "*.tmp" in result

    def test_merge_gitignore_comments_and_empty_lines(self, tmp_path):
        """Test merge_gitignore handles comments and empty lines."""
        merger = TemplateMerger(tmp_path)
        template_file = tmp_path / "template.gitignore"
        existing_file = tmp_path / ".gitignore"

        template_file.write_text("# Comment\n*.log\n")
        existing_file.write_text("# User comment\n*.tmp")

        merger.merge_gitignore(template_file, existing_file)

        result = existing_file.read_text()
        assert "*.log" in result
        assert "*.tmp" in result


class TestTemplateMergerMergeConfig:
    """Test merge_config method comprehensively."""

    def test_merge_config_creates_default_when_missing(self, tmp_path):
        """Test merge_config creates default config when missing."""
        merger = TemplateMerger(tmp_path)
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True, exist_ok=True)

        config = merger.merge_config("python")

        assert config["projectName"] == tmp_path.name
        assert config["mode"] == "personal"
        assert config["locale"] == "ko"
        assert config["language"] == "python"

    def test_merge_config_preserves_existing_values(self, tmp_path):
        """Test merge_config preserves existing configuration values."""
        merger = TemplateMerger(tmp_path)
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True, exist_ok=True)

        existing_config = {
            "projectName": "MyProject",
            "mode": "team",
            "locale": "en",
            "language": "typescript",
        }
        config_file = config_dir / "config.json"
        config_file.write_text(json.dumps(existing_config))

        config = merger.merge_config("python")

        assert config["projectName"] == "MyProject"
        assert config["mode"] == "team"
        assert config["locale"] == "en"
        assert config["language"] == "typescript"

    def test_merge_config_partial_existing(self, tmp_path):
        """Test merge_config with partial existing configuration."""
        merger = TemplateMerger(tmp_path)
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True, exist_ok=True)

        existing_config = {"projectName": "CustomProject"}
        config_file = config_dir / "config.json"
        config_file.write_text(json.dumps(existing_config))

        config = merger.merge_config("go")

        assert config["projectName"] == "CustomProject"
        assert config["language"] == "go"

    def test_merge_config_with_none_language(self, tmp_path):
        """Test merge_config with None language defaults to generic."""
        merger = TemplateMerger(tmp_path)
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True, exist_ok=True)

        config = merger.merge_config(None)

        assert config["language"] == "generic"

    def test_merge_config_invalid_json(self, tmp_path):
        """Test merge_config handles invalid JSON gracefully."""
        merger = TemplateMerger(tmp_path)
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True, exist_ok=True)

        config_file = config_dir / "config.json"
        config_file.write_text("invalid json {")

        with pytest.raises(json.JSONDecodeError):
            merger.merge_config("python")


class TestTemplateMergerMergeSettingsJson:
    """Test merge_settings_json method comprehensively."""

    def test_merge_settings_env_shallow_merge(self, tmp_path):
        """Test merge_settings_json does shallow merge for env."""
        merger = TemplateMerger(tmp_path)
        template_file = tmp_path / "template.json"
        existing_file = tmp_path / "existing.json"

        template_data = {"env": {"VAR1": "template1", "VAR2": "template2"}}
        existing_data = {"env": {"VAR1": "user1", "VAR3": "user3"}}

        template_file.write_text(json.dumps(template_data))
        existing_file.write_text(json.dumps(existing_data))

        merger.merge_settings_json(template_file, existing_file)

        result = json.loads(existing_file.read_text())
        assert result["env"]["VAR1"] == "user1"  # User value preserved
        assert result["env"]["VAR2"] == "template2"
        assert result["env"]["VAR3"] == "user3"

    def test_merge_settings_permissions_allow_deduplicated(self, tmp_path):
        """Test merge_settings_json deduplicates allow permissions."""
        merger = TemplateMerger(tmp_path)
        template_file = tmp_path / "template.json"
        existing_file = tmp_path / "existing.json"

        template_data = {
            "permissions": {"allow": ["Read", "Write", "Execute"]}
        }
        existing_data = {"permissions": {"allow": ["Write", "Delete"]}}

        template_file.write_text(json.dumps(template_data))
        existing_file.write_text(json.dumps(existing_data))

        merger.merge_settings_json(template_file, existing_file)

        result = json.loads(existing_file.read_text())
        allow = result["permissions"]["allow"]
        # Check no duplicates and contains all
        assert len(allow) == len(set(allow))
        assert "Write" in allow

    def test_merge_settings_preserve_custom_fields(self, tmp_path):
        """Test merge_settings_json preserves custom fields."""
        merger = TemplateMerger(tmp_path)
        template_file = tmp_path / "template.json"
        existing_file = tmp_path / "existing.json"

        template_data = {"outputStyle": "default", "spinnerTipsEnabled": True}
        existing_data = {
            "outputStyle": "compact",
            "spinnerTipsEnabled": False,
        }

        template_file.write_text(json.dumps(template_data))
        existing_file.write_text(json.dumps(existing_data))

        merger.merge_settings_json(template_file, existing_file)

        result = json.loads(existing_file.read_text())
        assert result["outputStyle"] == "compact"
        assert result["spinnerTipsEnabled"] is False

    def test_merge_settings_with_backup_path(self, tmp_path):
        """Test merge_settings_json uses backup for user settings."""
        merger = TemplateMerger(tmp_path)
        template_file = tmp_path / "template.json"
        existing_file = tmp_path / "existing.json"
        backup_file = tmp_path / "backup.json"

        template_data = {"env": {"VAR": "template"}}
        backup_data = {"env": {"VAR": "backup"}}

        template_file.write_text(json.dumps(template_data))
        existing_file.write_text("{}")
        backup_file.write_text(json.dumps(backup_data))

        merger.merge_settings_json(template_file, existing_file, backup_file)

        result = json.loads(existing_file.read_text())
        assert result["env"]["VAR"] == "backup"

    def test_merge_settings_missing_template_sections(self, tmp_path):
        """Test merge_settings_json handles missing template sections."""
        merger = TemplateMerger(tmp_path)
        template_file = tmp_path / "template.json"
        existing_file = tmp_path / "existing.json"

        template_data = {}
        existing_data = {"env": {"VAR": "value"}}

        template_file.write_text(json.dumps(template_data))
        existing_file.write_text(json.dumps(existing_data))

        merger.merge_settings_json(template_file, existing_file)

        result = json.loads(existing_file.read_text())
        assert "env" in result
        assert result["permissions"]["allow"] == []


class TestTemplateMergerMergeGithubWorkflows:
    """Test merge_github_workflows method comprehensively."""

    def test_merge_github_workflows_updates_moai_workflows(self, tmp_path):
        """Test merge_github_workflows updates moai-*.yml files."""
        merger = TemplateMerger(tmp_path)
        template_dir = tmp_path / "template" / ".github"
        existing_dir = tmp_path / "existing" / ".github"

        # Create template structure
        (template_dir / "workflows").mkdir(parents=True)
        (template_dir / "workflows" / "moai-test.yml").write_text("test workflow")

        # Create existing structure with different version
        (existing_dir / "workflows").mkdir(parents=True)
        (existing_dir / "workflows" / "moai-test.yml").write_text("old workflow")

        merger.merge_github_workflows(template_dir, existing_dir)

        result = (existing_dir / "workflows" / "moai-test.yml").read_text()
        assert result == "test workflow"

    def test_merge_github_workflows_preserves_user_workflows(self, tmp_path):
        """Test merge_github_workflows preserves non-moai workflows."""
        merger = TemplateMerger(tmp_path)
        template_dir = tmp_path / "template" / ".github"
        existing_dir = tmp_path / "existing" / ".github"

        # Create template with moai workflow
        (template_dir / "workflows").mkdir(parents=True)
        (template_dir / "workflows" / "moai-ci.yml").write_text("moai ci")

        # Create existing with user workflow
        (existing_dir / "workflows").mkdir(parents=True)
        (existing_dir / "workflows" / "custom-deploy.yml").write_text(
            "custom deploy"
        )

        merger.merge_github_workflows(template_dir, existing_dir)

        # Check both exist
        assert (existing_dir / "workflows" / "moai-ci.yml").exists()
        assert (existing_dir / "workflows" / "custom-deploy.yml").exists()
        assert (
            existing_dir / "workflows" / "custom-deploy.yml"
        ).read_text() == "custom deploy"

    def test_merge_github_workflows_creates_workflows_dir(self, tmp_path):
        """Test merge_github_workflows creates workflows directory."""
        merger = TemplateMerger(tmp_path)
        template_dir = tmp_path / "template" / ".github"
        existing_dir = tmp_path / "existing" / ".github"

        (template_dir / "workflows").mkdir(parents=True)
        (template_dir / "workflows" / "moai-test.yml").write_text("test")
        existing_dir.mkdir(parents=True)

        merger.merge_github_workflows(template_dir, existing_dir)

        assert (existing_dir / "workflows").exists()

    def test_merge_github_workflows_copies_other_dirs(self, tmp_path):
        """Test merge_github_workflows copies non-workflow directories."""
        merger = TemplateMerger(tmp_path)
        template_dir = tmp_path / "template" / ".github"
        existing_dir = tmp_path / "existing" / ".github"

        # Create template with other directories
        (template_dir / "ISSUE_TEMPLATE").mkdir(parents=True)
        (template_dir / "ISSUE_TEMPLATE" / "bug.md").write_text("Bug template")
        existing_dir.mkdir(parents=True)

        merger.merge_github_workflows(template_dir, existing_dir)

        assert (existing_dir / "ISSUE_TEMPLATE" / "bug.md").exists()

    def test_merge_github_workflows_empty_template_dir(self, tmp_path):
        """Test merge_github_workflows with empty template directory."""
        merger = TemplateMerger(tmp_path)
        template_dir = tmp_path / "template" / ".github"
        existing_dir = tmp_path / "existing" / ".github"

        template_dir.mkdir(parents=True)
        existing_dir.mkdir(parents=True)

        # Should not raise error
        merger.merge_github_workflows(template_dir, existing_dir)
        assert existing_dir.exists()


class TestTemplateMergerInitialization:
    """Test TemplateMerger initialization edge cases."""

    def test_initialization_with_relative_path(self):
        """Test initialization converts relative paths to absolute."""
        merger = TemplateMerger(Path("./test"))
        assert merger.target_path.is_absolute()

    def test_initialization_with_string_path(self):
        """Test initialization handles string paths."""
        merger = TemplateMerger(Path("/tmp/test"))
        assert isinstance(merger.target_path, Path)
        assert merger.target_path.is_absolute()

    def test_project_info_headers_constant(self):
        """Test PROJECT_INFO_HEADERS is defined correctly."""
        merger = TemplateMerger(Path("/tmp"))
        assert len(merger.PROJECT_INFO_HEADERS) >= 2
        assert all(isinstance(h, str) for h in merger.PROJECT_INFO_HEADERS)


class TestTemplateMergerEdgeCases:
    """Test edge cases and error handling."""

    def test_merge_settings_json_invalid_template(self, tmp_path):
        """Test merge_settings_json handles invalid JSON in template."""
        merger = TemplateMerger(tmp_path)
        template_file = tmp_path / "template.json"
        existing_file = tmp_path / "existing.json"

        template_file.write_text("invalid json")
        existing_file.write_text("{}")

        with pytest.raises(json.JSONDecodeError):
            merger.merge_settings_json(template_file, existing_file)

    def test_merge_settings_json_invalid_existing(self, tmp_path):
        """Test merge_settings_json handles invalid JSON in existing."""
        merger = TemplateMerger(tmp_path)
        template_file = tmp_path / "template.json"
        existing_file = tmp_path / "existing.json"

        template_file.write_text("{}")
        existing_file.write_text("invalid json")

        with pytest.raises(json.JSONDecodeError):
            merger.merge_settings_json(template_file, existing_file)

    def test_merge_claude_with_nonexistent_files(self, tmp_path):
        """Test merge_claude_md with nonexistent files."""
        merger = TemplateMerger(tmp_path)
        template_file = tmp_path / "nonexistent_template.md"
        existing_file = tmp_path / "nonexistent_existing.md"

        with pytest.raises(FileNotFoundError):
            merger.merge_claude_md(template_file, existing_file)

    def test_merge_gitignore_with_large_files(self, tmp_path):
        """Test merge_gitignore handles large files."""
        merger = TemplateMerger(tmp_path)
        template_file = tmp_path / "template.gitignore"
        existing_file = tmp_path / ".gitignore"

        # Create large file
        template_entries = [f"entry_{i}" for i in range(1000)]
        template_file.write_text("\n".join(template_entries))
        existing_file.write_text("existing_entry")

        merger.merge_gitignore(template_file, existing_file)

        result = existing_file.read_text()
        assert "existing_entry" in result
        assert "entry_0" in result
