# @TEST:TEST-COVERAGE-001 | SPEC: SPEC-TEST-COVERAGE-001.md
"""Integration tests for TemplateProcessor

Tests template file operations:
- Template copying with backup
- Protected paths handling
- File merging (CLAUDE.md, .gitignore, config.json)
- Backup creation
"""

import json
import sys
from pathlib import Path
from unittest.mock import Mock, patch

import pytest

from moai_adk.core.template.processor import TemplateProcessor


class TestTemplateProcessorInit:
    """Test TemplateProcessor initialization"""

    def test_init_resolves_target_path(self, tmp_path: Path) -> None:
        """Should resolve target path to absolute path"""
        processor = TemplateProcessor(tmp_path)
        assert processor.target_path.is_absolute()
        assert processor.target_path == tmp_path.resolve()

    def test_init_gets_template_root(self, tmp_path: Path) -> None:
        """Should find template root directory"""
        processor = TemplateProcessor(tmp_path)
        assert processor.template_root.exists()
        assert processor.template_root.name == "templates"


class TestCopyTemplates:
    """Test template copying workflow"""

    @patch("moai_adk.core.template.processor.Console")
    def test_copy_templates_without_backup(
        self, mock_console: Mock, tmp_path: Path
    ) -> None:
        """Should copy templates without backup when no existing files"""
        processor = TemplateProcessor(tmp_path)
        processor.copy_templates(backup=False, silent=True)

        # Should create .moai directory
        assert (tmp_path / ".moai").exists()
        # Should copy .github workflow templates
        assert (tmp_path / ".github").exists()
        assert (tmp_path / ".github" / "workflows").exists()

    @patch("moai_adk.core.template.processor.Console")
    def test_copy_templates_with_backup(
        self, mock_console: Mock, tmp_path: Path
    ) -> None:
        """Should create backup when existing files present"""
        # Create existing files
        (tmp_path / ".moai").mkdir()
        (tmp_path / ".moai" / "config.json").write_text("{}")
        (tmp_path / ".github").mkdir()

        processor = TemplateProcessor(tmp_path)
        processor.copy_templates(backup=True, silent=True)

        # Should create backup directory (.moai-backups/{timestamp}/)
        backup_dir = tmp_path / ".moai-backups"
        assert backup_dir.exists()

    @patch("moai_adk.core.template.processor.Console")
    def test_copy_templates_silent_mode(
        self, mock_console: Mock, tmp_path: Path
    ) -> None:
        """Should not print messages in silent mode"""
        processor = TemplateProcessor(tmp_path)
        processor.copy_templates(silent=True)

        # Console.print should not be called in silent mode
        mock_console.return_value.print.assert_not_called()

    def test_copy_github_overwrites_existing_directory(self, tmp_path: Path) -> None:
        """Should replace existing .github directory with template version"""
        github_dir = tmp_path / ".github"
        workflows_dir = github_dir / "workflows"
        workflows_dir.mkdir(parents=True)
        # create stale file
        stale = workflows_dir / "old.yml"
        stale.write_text("# old workflow")

        processor = TemplateProcessor(tmp_path)
        processor._copy_github(silent=True)

        assert not stale.exists()
        assert (github_dir / "workflows" / "moai-gitflow.yml").exists()


class TestClaudeTemplate:
    """Test CLAUDE.md template copying"""

    def test_copy_claude_md_uses_english_template(self, tmp_path: Path) -> None:
        """Should copy English CLAUDE.md template by default"""
        processor = TemplateProcessor(tmp_path)
        processor._copy_claude_md(silent=True)

        content = (tmp_path / "CLAUDE.md").read_text(encoding="utf-8")
        assert "Meet Alfred: Your MoAI SuperAgent" in content
        assert "Project Information" in content
        assert "페르소나" not in content


class TestHasExistingFiles:
    """Test existing files detection"""

    def test_has_existing_files_with_moai(self, tmp_path: Path) -> None:
        """Should return True when .moai exists"""
        (tmp_path / ".moai").mkdir()
        processor = TemplateProcessor(tmp_path)
        assert processor._has_existing_files() is True

    def test_has_existing_files_with_claude(self, tmp_path: Path) -> None:
        """Should return True when .claude exists"""
        (tmp_path / ".claude").mkdir()
        processor = TemplateProcessor(tmp_path)
        assert processor._has_existing_files() is True

    def test_has_existing_files_with_claude_md(self, tmp_path: Path) -> None:
        """Should return True when CLAUDE.md exists"""
        (tmp_path / "CLAUDE.md").write_text("# Project")
        processor = TemplateProcessor(tmp_path)
        assert processor._has_existing_files() is True

    def test_has_existing_files_with_github(self, tmp_path: Path) -> None:
        """Should return True when .github exists"""
        (tmp_path / ".github").mkdir()
        processor = TemplateProcessor(tmp_path)
        assert processor._has_existing_files() is True

    def test_has_existing_files_with_none(self, tmp_path: Path) -> None:
        """Should return False when no existing files"""
        processor = TemplateProcessor(tmp_path)
        assert processor._has_existing_files() is False


class TestCreateBackup:
    """Test backup creation"""

    def test_create_backup_creates_directory(self, tmp_path: Path) -> None:
        """Should create timestamped backup directory (.moai-backups/{timestamp}/)"""
        (tmp_path / ".moai").mkdir()
        (tmp_path / ".moai" / "config.json").write_text("{}")

        processor = TemplateProcessor(tmp_path)
        backup_path = processor.create_backup()

        assert backup_path.exists()
        assert backup_path.parent.name == ".moai-backups"
        assert len(backup_path.name) == 15  # YYYYMMDD-HHMMSS

    def test_create_backup_copies_moai_directory(self, tmp_path: Path) -> None:
        """Should backup .moai directory"""
        moai_dir = tmp_path / ".moai"
        moai_dir.mkdir()
        (moai_dir / "config.json").write_text('{"test": "value"}')

        processor = TemplateProcessor(tmp_path)
        backup_path = processor.create_backup()

        # Should copy .moai
        backed_up_config = backup_path / ".moai" / "config.json"
        assert backed_up_config.exists()
        assert "test" in backed_up_config.read_text()

    def test_create_backup_excludes_protected_paths(self, tmp_path: Path) -> None:
        """Should exclude protected paths from backup"""
        moai_dir = tmp_path / ".moai"
        moai_dir.mkdir()
        (moai_dir / "config.json").write_text("{}")

        # Create protected paths
        specs_dir = moai_dir / "specs"
        specs_dir.mkdir()
        (specs_dir / "test.md").write_text("# SPEC")

        processor = TemplateProcessor(tmp_path)
        backup_path = processor.create_backup()

        # Protected paths should not be in backup
        assert not (backup_path / ".moai" / "specs").exists()

    def test_create_backup_copies_claude_md(self, tmp_path: Path) -> None:
        """Should backup CLAUDE.md file"""
        (tmp_path / "CLAUDE.md").write_text("# Project Doc")

        processor = TemplateProcessor(tmp_path)
        backup_path = processor.create_backup()

        backed_up_md = backup_path / "CLAUDE.md"
        assert backed_up_md.exists()
        assert "Project Doc" in backed_up_md.read_text()

    def test_create_backup_copies_github_directory(self, tmp_path: Path) -> None:
        """Should backup .github directory"""
        workflows_dir = tmp_path / ".github" / "workflows"
        workflows_dir.mkdir(parents=True, exist_ok=True)
        (workflows_dir / "custom.yml").write_text("# workflow")

        processor = TemplateProcessor(tmp_path)
        backup_path = processor.create_backup()

        backed_up_workflow = backup_path / ".github" / "workflows" / "custom.yml"
        assert backed_up_workflow.exists()


class TestCopyExcludeProtected:
    """Test selective copying excluding protected paths"""

    def test_copy_exclude_protected_preserves_existing_files(
        self, tmp_path: Path
    ) -> None:
        """Should preserve existing files (v0.3.0 behavior)"""
        src = tmp_path / "src"
        dst = tmp_path / "dst"
        src.mkdir()

        # Create source file
        (src / "template.json").write_text('{"template": "value"}')

        # Create existing destination file
        dst.mkdir()
        (dst / "template.json").write_text('{"user": "data"}')

        processor = TemplateProcessor(tmp_path)
        processor._copy_exclude_protected(src, dst)

        # Should NOT overwrite existing file
        content = (dst / "template.json").read_text()
        assert "user" in content
        assert "template" not in content

    def test_copy_exclude_protected_skips_specs(self, tmp_path: Path) -> None:
        """Should skip specs/ directory"""
        src = tmp_path / "src"
        dst = tmp_path / "dst"
        src.mkdir()

        # Create specs directory
        specs_dir = src / "specs"
        specs_dir.mkdir()
        (specs_dir / "test.md").write_text("# SPEC")

        processor = TemplateProcessor(tmp_path)
        processor._copy_exclude_protected(src, dst)

        # specs should not be copied
        assert not (dst / "specs").exists()

    def test_copy_exclude_protected_skips_reports(self, tmp_path: Path) -> None:
        """Should skip reports/ directory"""
        src = tmp_path / "src"
        dst = tmp_path / "dst"
        src.mkdir()

        # Create reports directory
        reports_dir = src / "reports"
        reports_dir.mkdir()
        (reports_dir / "report.md").write_text("# Report")

        processor = TemplateProcessor(tmp_path)
        processor._copy_exclude_protected(src, dst)

        # reports should not be copied
        assert not (dst / "reports").exists()


class TestCopyClaude:
    """Test .claude directory copying"""

    @patch("moai_adk.core.template.processor.Console")
    def test_copy_claude_copies_directory(
        self, mock_console: Mock, tmp_path: Path
    ) -> None:
        """Should copy .claude directory from template"""
        processor = TemplateProcessor(tmp_path)
        processor._copy_claude(silent=True)

        # .claude should be copied
        claude_dir = tmp_path / ".claude"
        assert claude_dir.exists()

    @pytest.mark.skip(reason="Test requires _backup_alfred_folder method from develop branch")
    @patch("moai_adk.core.template.processor.Console")
    def test_copy_claude_overwrites_alfred_folders(
        self, mock_console: Mock, tmp_path: Path
    ) -> None:
        """Should overwrite Alfred folders but preserve other files"""
        # Create existing .claude with old file and Alfred folder
        old_claude = tmp_path / ".claude"
        old_claude.mkdir()
        (old_claude / "old.txt").write_text("old content")

        # Create old Alfred folder content
        alfred_hooks = old_claude / "hooks" / "alfred"
        alfred_hooks.mkdir(parents=True)
        (alfred_hooks / "old_hook.py").write_text("# old hook")

        processor = TemplateProcessor(tmp_path)
        processor._copy_claude(silent=True)

        # Alfred folders should be overwritten (old_hook.py removed)
        assert not (tmp_path / ".claude" / "hooks" / "alfred" / "old_hook.py").exists()

        # Other files should be preserved (old.txt still exists)
        assert (tmp_path / ".claude" / "old.txt").exists()
        assert (tmp_path / ".claude" / "old.txt").read_text() == "old content"

    @pytest.mark.skip(reason="Test requires _backup_alfred_folder method from develop branch")
    @patch("moai_adk.core.template.processor.Console")
    def test_copy_claude_all_alfred_folders_overwritten(
        self, mock_console: Mock, tmp_path: Path
    ) -> None:
        """Should overwrite all 4 Alfred folders"""
        # Create existing .claude with all Alfred folders containing old files
        old_claude = tmp_path / ".claude"
        old_claude.mkdir()

        alfred_folders = [
            "hooks/alfred",
            "commands/alfred",
            "output-styles/alfred",
            "agents/alfred",
        ]

        for folder in alfred_folders:
            folder_path = old_claude / folder
            folder_path.mkdir(parents=True)
            (folder_path / "old_file.txt").write_text(f"old content in {folder}")

        processor = TemplateProcessor(tmp_path)
        processor._copy_claude(silent=True)

        # All Alfred folder old files should be removed
        for folder in alfred_folders:
            old_file = tmp_path / ".claude" / folder / "old_file.txt"
            assert not old_file.exists(), f"Old file in {folder} should be removed"

    @pytest.mark.skip(reason="Test requires _backup_alfred_folder method from develop branch")
    @patch("moai_adk.core.template.processor.Console")
    def test_copy_claude_backups_alfred_folders_before_overwrite(
        self, mock_console: Mock, tmp_path: Path
    ) -> None:
        """Should backup Alfred folders before overwriting them"""
        # Create existing .claude with Alfred folder
        old_claude = tmp_path / ".claude"
        old_claude.mkdir()

        alfred_hooks = old_claude / "hooks" / "alfred"
        alfred_hooks.mkdir(parents=True)
        (alfred_hooks / "important_hook.py").write_text("# important")

        # Create backup first (simulating Phase 2 of copy_templates)
        processor = TemplateProcessor(tmp_path)
        backup_path = processor.create_backup()

        # Now copy .claude (which should find the backup)
        processor._copy_claude(silent=True)

        # Backup should exist at .moai-backups/{timestamp}/.claude/
        backup_claude = backup_path / ".claude"
        assert backup_claude.exists(), "Backup .claude directory should be created"

        # Check that hooks/alfred was backed up
        backup_hooks = backup_claude / "hooks" / "alfred"
        assert backup_hooks.exists(), "hooks/alfred should be backed up"
        assert (backup_hooks / "important_hook.py").exists(), "Hook files should be in backup"


class TestCopyMoai:
    """Test .moai directory copying"""

    @patch("moai_adk.core.template.processor.Console")
    def test_copy_moai_preserves_existing_files(
        self, mock_console: Mock, tmp_path: Path
    ) -> None:
        """Should preserve existing user files (v0.3.0)"""
        # Create existing .moai with user file
        moai_dir = tmp_path / ".moai"
        moai_dir.mkdir()
        memory_dir = moai_dir / "memory"
        memory_dir.mkdir()
        user_file = memory_dir / "user-notes.md"
        user_file.write_text("# User Notes")

        processor = TemplateProcessor(tmp_path)
        processor._copy_moai(silent=True)

        # User file should be preserved
        assert user_file.exists()
        assert "User Notes" in user_file.read_text()

    @patch("moai_adk.core.template.processor.Console")
    def test_copy_moai_skips_specs_directory(
        self, mock_console: Mock, tmp_path: Path
    ) -> None:
        """Should not copy specs/ from template"""
        processor = TemplateProcessor(tmp_path)
        processor._copy_moai(silent=True)

        # Template specs should not be copied
        # (Only user specs should exist)
        specs_dir = tmp_path / ".moai" / "specs"
        if specs_dir.exists():
            # Should be empty or contain only user content
            assert len(list(specs_dir.iterdir())) == 0 or not (
                specs_dir / "SPEC-TEMPLATE"
            ).exists()


class TestMergeClaudeMd:
    """Test CLAUDE.md merging"""

    @pytest.mark.skipif(sys.platform == "win32", reason="Windows charmap encoding issue")
    def test_merge_claude_md_preserves_project_info(self, tmp_path: Path) -> None:
        """Should preserve project info section when merging"""
        # Create template CLAUDE.md
        template = tmp_path / "template.md"
        template.write_text(
            "# Template\n\nContent\n\n## 프로젝트 정보\n\n- Template Project"
        )

        # Create existing CLAUDE.md with user project info
        existing = tmp_path / "existing.md"
        existing.write_text(
            "# Old Template\n\n## 프로젝트 정보\n\n- My Project\n- Version: 1.0.0"
        )

        processor = TemplateProcessor(tmp_path)
        processor._merge_claude_md(template, existing)

        # Should merge: template content + user project info
        merged = existing.read_text()
        assert "# Template" in merged
        assert "Content" in merged
        assert "My Project" in merged
        assert "Version: 1.0.0" in merged
        assert "Template Project" not in merged

    @pytest.mark.skipif(sys.platform == "win32", reason="Windows charmap encoding issue")
    def test_merge_claude_md_preserves_project_info_en(self, tmp_path: Path) -> None:
        """Should preserve project info section for English templates"""
        template = tmp_path / "template.md"
        template.write_text(
            "# Template\n\nContent\n\n## Project Information\n\n- Template Project"
        )

        existing = tmp_path / "existing.md"
        existing.write_text(
            "# Old Template\n\n## Project Information\n\n- My Project\n- Version: 1.0.0"
        )

        processor = TemplateProcessor(tmp_path)
        processor._merge_claude_md(template, existing)

        merged = existing.read_text()
        assert "Project Information" in merged
        assert "My Project" in merged
        assert "Template Project" not in merged

    def test_merge_claude_md_without_project_info(self, tmp_path: Path) -> None:
        """Should use template as-is when no project info exists"""
        template = tmp_path / "template.md"
        template.write_text("# Template\n\nContent")

        existing = tmp_path / "existing.md"
        existing.write_text("# Old Content")

        processor = TemplateProcessor(tmp_path)
        processor._merge_claude_md(template, existing)

        # Should replace with template
        merged = existing.read_text()
        assert "# Template" in merged
        assert "Content" in merged
        assert "Old Content" not in merged


class TestMergeGitignore:
    """Test .gitignore merging"""

    def test_merge_gitignore_combines_entries(self, tmp_path: Path) -> None:
        """Should combine template and existing .gitignore entries"""
        template = tmp_path / "template.gitignore"
        template.write_text("node_modules/\n.env\n")

        existing = tmp_path / "existing.gitignore"
        existing.write_text(".vscode/\n.DS_Store\n")

        processor = TemplateProcessor(tmp_path)
        processor._merge_gitignore(template, existing)

        merged = existing.read_text()
        assert "node_modules/" in merged
        assert ".env" in merged
        assert ".vscode/" in merged
        assert ".DS_Store" in merged

    def test_merge_gitignore_removes_duplicates(self, tmp_path: Path) -> None:
        """Should not duplicate existing entries"""
        template = tmp_path / "template.gitignore"
        template.write_text("node_modules/\n.env\n")

        existing = tmp_path / "existing.gitignore"
        existing.write_text("node_modules/\n.vscode/\n")

        processor = TemplateProcessor(tmp_path)
        processor._merge_gitignore(template, existing)

        merged = existing.read_text()
        # node_modules should appear only once
        assert merged.count("node_modules/") == 1


class TestMergeConfig:
    """Test config.json merging"""

    def test_merge_config_preserves_existing_values(self, tmp_path: Path) -> None:
        """Should preserve existing config values"""
        config_dir = tmp_path / ".moai"
        config_dir.mkdir()
        config_path = config_dir / "config.json"
        config_path.write_text(
            json.dumps(
                {
                    "projectName": "MyProject",
                    "mode": "team",
                    "locale": "en",
                    "language": "typescript",
                }
            )
        )

        processor = TemplateProcessor(tmp_path)
        merged = processor.merge_config(detected_language="python")

        # Should preserve existing values, not use detected language
        assert merged["projectName"] == "MyProject"
        assert merged["mode"] == "team"
        assert merged["locale"] == "en"
        assert merged["language"] == "typescript"

    def test_merge_config_uses_defaults_for_new_project(self, tmp_path: Path) -> None:
        """Should use defaults when no existing config"""
        config_dir = tmp_path / ".moai"
        config_dir.mkdir()

        processor = TemplateProcessor(tmp_path)
        merged = processor.merge_config(detected_language="python")

        # Should use defaults
        assert merged["projectName"] == tmp_path.name
        assert merged["mode"] == "personal"
        assert merged["locale"] == "ko"
        assert merged["language"] == "python"

    def test_merge_config_uses_detected_language_when_no_existing(
        self, tmp_path: Path
    ) -> None:
        """Should use detected language for new projects"""
        config_dir = tmp_path / ".moai"
        config_dir.mkdir()

        processor = TemplateProcessor(tmp_path)
        merged = processor.merge_config(detected_language="go")

        assert merged["language"] == "go"
