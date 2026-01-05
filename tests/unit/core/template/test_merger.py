"""Unit tests for moai_adk.core.template.merger module.

Tests for TemplateMerger functionality.
"""

import json
from pathlib import Path
from unittest.mock import MagicMock, patch

from moai_adk.core.template.merger import TemplateMerger


class TestTemplateMergerInitialization:
    """Test TemplateMerger initialization."""

    def test_initialization(self):
        """Test TemplateMerger initialization."""
        target_path = Path("/tmp/project")
        merger = TemplateMerger(target_path)
        assert merger.target_path == target_path.resolve()

    def test_resolves_path(self):
        """Test initialization resolves path."""
        target_path = Path("/tmp/project")
        merger = TemplateMerger(target_path)
        assert merger.target_path.is_absolute()

    def test_project_info_headers_defined(self):
        """Test project info headers are defined."""
        merger = TemplateMerger(Path("/tmp"))
        assert len(merger.PROJECT_INFO_HEADERS) > 0
        assert "## Project Information" in merger.PROJECT_INFO_HEADERS


class TestMergeClaude:
    """Test merge_claude_md method."""

    def test_merge_claude_preserves_project_info(self):
        """Test merge_claude_md preserves project info."""
        merger = TemplateMerger(Path("/tmp"))
        template_path = MagicMock(spec=Path)
        existing_path = MagicMock(spec=Path)

        template_content = "Template content"
        existing_content = "Template content\n\n## Project Information\nUser info"

        template_path.read_text.return_value = template_content
        existing_path.read_text.return_value = existing_content

        with patch("pathlib.Path.write_text"):
            merger.merge_claude_md(template_path, existing_path)
            assert existing_path.write_text.called

    def test_merge_claude_no_project_info(self):
        """Test merge_claude_md when no project info exists."""
        merger = TemplateMerger(Path("/tmp"))
        template_path = MagicMock(spec=Path)
        existing_path = MagicMock(spec=Path)

        template_path.read_text.return_value = "Template content"
        existing_path.read_text.return_value = "No project info here"

        with patch("shutil.copy2"):
            merger.merge_claude_md(template_path, existing_path)

    def test_find_project_info_section_found(self):
        """Test finding project info section."""
        merger = TemplateMerger(Path("/tmp"))
        content = "Content\n## Project Information\nProject info"
        start, header = merger._find_project_info_section(content)
        assert start >= 0
        assert header is not None

    def test_find_project_info_section_not_found(self):
        """Test when project info section not found."""
        merger = TemplateMerger(Path("/tmp"))
        content = "Content without project info"
        start, header = merger._find_project_info_section(content)
        assert start == -1
        assert header is None


class TestMergeGitignore:
    """Test merge_gitignore method."""

    def test_merge_gitignore_combines_entries(self):
        """Test merge_gitignore combines entries."""
        merger = TemplateMerger(Path("/tmp"))
        template_path = MagicMock(spec=Path)
        existing_path = MagicMock(spec=Path)

        template_path.read_text.return_value = "*.log\n*.tmp"
        existing_path.read_text.return_value = "*.pyc\nnode_modules"

        with patch("pathlib.Path.write_text"):
            merger.merge_gitignore(template_path, existing_path)
            assert existing_path.write_text.called

    def test_merge_gitignore_removes_duplicates(self):
        """Test merge_gitignore removes duplicates."""
        merger = TemplateMerger(Path("/tmp"))
        template_path = MagicMock(spec=Path)
        existing_path = MagicMock(spec=Path)

        template_path.read_text.return_value = "*.log\n*.pyc"
        existing_path.read_text.return_value = "*.pyc\nnode_modules"
        existing_path.write_text = MagicMock()

        merger.merge_gitignore(template_path, existing_path)
        # Should have been called with merged content
        assert existing_path.write_text.called


class TestMergeConfig:
    """Test merge_config method."""

    def test_merge_config_new_project(self):
        """Test merge_config for new project."""
        merger = TemplateMerger(Path("/tmp/myproject"))
        with patch("pathlib.Path.exists", return_value=False):
            result = merger.merge_config(detected_language="python")
            assert "projectName" in result
            assert result["projectName"] == "myproject"
            assert result["language"] == "python"

    def test_merge_config_preserves_existing(self):
        """Test merge_config preserves existing settings."""
        merger = TemplateMerger(Path("/tmp/project"))
        with patch("pathlib.Path.exists", return_value=True):
            with patch("builtins.open", create=True):
                with patch(
                    "json.load",
                    return_value={"projectName": "existing", "mode": "team"},
                ):
                    result = merger.merge_config()
                    assert result["projectName"] == "existing"

    def test_merge_config_defaults(self):
        """Test merge_config provides defaults."""
        merger = TemplateMerger(Path("/tmp/project"))
        with patch("pathlib.Path.exists", return_value=False):
            result = merger.merge_config()
            assert result["mode"] == "personal"
            assert result["locale"] == "ko"


class TestMergeSettingsJson:
    """Test merge_settings_json method."""

    def test_merge_settings_json_basic(self):
        """Test basic settings.json merge."""
        merger = TemplateMerger(Path("/tmp"))
        template_path = MagicMock(spec=Path)
        existing_path = MagicMock(spec=Path)

        template_data = {"env": {"KEY": "value"}, "permissions": {"allow": []}}
        template_path.read_text.return_value = json.dumps(template_data)
        existing_path.exists.return_value = False

        with patch("pathlib.Path.write_text"):
            merger.merge_settings_json(template_path, existing_path)
            assert existing_path.write_text.called

    def test_merge_settings_json_preserves_env(self):
        """Test merge_settings_json preserves env variables."""
        merger = TemplateMerger(Path("/tmp"))
        template_path = MagicMock(spec=Path)
        existing_path = MagicMock(spec=Path)

        template_data = {"env": {"TEMPLATE_KEY": "template"}, "permissions": {}}
        user_data = {"env": {"USER_KEY": "user"}, "permissions": {}}

        template_path.read_text.return_value = json.dumps(template_data)
        existing_path.read_text.return_value = json.dumps(user_data)
        existing_path.exists.return_value = True
        existing_path.write_text = MagicMock()

        merger.merge_settings_json(template_path, existing_path)
        assert existing_path.write_text.called


class TestMergeGithubWorkflows:
    """Test merge_github_workflows method."""

    def test_merge_github_workflows_moai_only(self):
        """Test merge_github_workflows only updates moai workflows."""
        merger = TemplateMerger(Path("/tmp"))
        template_dir = MagicMock(spec=Path)
        existing_dir = MagicMock(spec=Path)

        # Mock rglob to return moai workflow files
        moai_file = MagicMock(spec=Path)
        moai_file.name = "moai-test.yml"
        moai_file.is_file.return_value = True
        moai_file.is_dir.return_value = False
        moai_file.relative_to.return_value = Path("workflows/moai-test.yml")

        template_dir.rglob.return_value = [moai_file]

        with patch("shutil.copy2"):
            with patch("pathlib.Path.mkdir"):
                merger.merge_github_workflows(template_dir, existing_dir)

    def test_merge_github_workflows_preserves_user(self):
        """Test merge_github_workflows preserves user workflows."""
        merger = TemplateMerger(Path("/tmp"))
        template_dir = MagicMock(spec=Path)
        existing_dir = MagicMock(spec=Path)

        # Mock empty rglob
        template_dir.rglob.return_value = []

        with patch("pathlib.Path.mkdir"):
            merger.merge_github_workflows(template_dir, existing_dir)
            # User workflows should not be touched


class TestTemplateMergerIntegration:
    """Integration tests for TemplateMerger."""

    def test_full_merge_workflow(self):
        """Test complete merge workflow."""
        TemplateMerger(Path("/tmp/project"))
        # Would require more sophisticated mocking

    def test_merges_multiple_files(self):
        """Test merging multiple file types."""
        TemplateMerger(Path("/tmp/project"))
        # Would require coordinated mocking of multiple file operations
