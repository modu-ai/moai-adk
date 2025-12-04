"""Unit tests for moai_adk.core.template.merger module."""

from pathlib import Path
from tempfile import TemporaryDirectory

import pytest

from moai_adk.core.template.merger import TemplateMerger


class TestTemplateMerger:
    """Test TemplateMerger class."""

    def test_init(self):
        """Test initialization."""
        with TemporaryDirectory() as tmpdir:
            merger = TemplateMerger(Path(tmpdir))
            assert merger.target_path == Path(tmpdir).resolve()

    def test_merge_claude_md_new_file(self):
        """Test merging CLAUDE.md for new project."""
        with TemporaryDirectory() as tmpdir:
            tmpdir_path = Path(tmpdir)
            merger = TemplateMerger(tmpdir_path)

            template_path = tmpdir_path / "template.md"
            existing_path = tmpdir_path / "existing.md"

            template_path.write_text("# Template\nContent")
            existing_path.write_text("# Existing\nContent")

            merger.merge_claude_md(template_path, existing_path)
            assert existing_path.exists()

    def test_find_project_info_section(self):
        """Test finding project info section."""
        with TemporaryDirectory() as tmpdir:
            merger = TemplateMerger(Path(tmpdir))
            content = "Text before\n## Project Information\nProject info\nText after"

            index, header = merger._find_project_info_section(content)
            assert index != -1
            assert "Project Information" in header

    def test_find_project_info_section_not_found(self):
        """Test when project info section not found."""
        with TemporaryDirectory() as tmpdir:
            merger = TemplateMerger(Path(tmpdir))
            content = "Text without section header"

            index, header = merger._find_project_info_section(content)
            assert index == -1
            assert header is None

    def test_merge_gitignore(self):
        """Test merging .gitignore."""
        with TemporaryDirectory() as tmpdir:
            tmpdir_path = Path(tmpdir)
            merger = TemplateMerger(tmpdir_path)

            template_path = tmpdir_path / "template.gitignore"
            existing_path = tmpdir_path / "existing.gitignore"

            template_path.write_text("*.pyc\n__pycache__/")
            existing_path.write_text("venv/\n.env")

            merger.merge_gitignore(template_path, existing_path)
            content = existing_path.read_text()

            assert "*.pyc" in content
            assert "venv/" in content

    def test_merge_config(self):
        """Test merging config.json."""
        with TemporaryDirectory() as tmpdir:
            tmpdir_path = Path(tmpdir)
            merger = TemplateMerger(tmpdir_path)

            result = merger.merge_config(detected_language="python")

            assert "projectName" in result
            assert result["language"] == "python"

    def test_merge_settings_json(self):
        """Test merging .claude/settings.json."""
        import json

        with TemporaryDirectory() as tmpdir:
            tmpdir_path = Path(tmpdir)
            merger = TemplateMerger(tmpdir_path)

            template_path = tmpdir_path / "template.json"
            existing_path = tmpdir_path / "existing.json"

            template_data = {
                "env": {"VAR1": "value1"},
                "permissions": {"allow": ["Read"]},
            }
            user_data = {
                "env": {"VAR2": "value2"},
                "permissions": {"allow": ["Write"]},
            }

            template_path.write_text(json.dumps(template_data))
            existing_path.write_text(json.dumps(user_data))

            merger.merge_settings_json(template_path, existing_path)
            result = json.loads(existing_path.read_text())

            assert "permissions" in result
            assert result["env"]["VAR1"] == "value1"

    def test_merge_github_workflows(self):
        """Test merging GitHub workflows."""
        with TemporaryDirectory() as tmpdir:
            tmpdir_path = Path(tmpdir)
            merger = TemplateMerger(tmpdir_path)

            template_dir = tmpdir_path / "template" / ".github"
            existing_dir = tmpdir_path / "existing" / ".github"

            template_dir.mkdir(parents=True)
            existing_dir.mkdir(parents=True)

            workflows_dir = template_dir / "workflows"
            workflows_dir.mkdir()
            (workflows_dir / "moai-test.yml").write_text("test")

            merger.merge_github_workflows(template_dir, existing_dir)
            assert (existing_dir / "workflows" / "moai-test.yml").exists()
