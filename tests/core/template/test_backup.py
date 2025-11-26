"""
Tests for core/template/backup.py module

Achieves 90%+ coverage by testing all paths including:
- Initialization and backup creation
- Directory exclusion logic
- Backup restoration
- Edge cases and error handling
"""

import shutil
from pathlib import Path

import pytest

from moai_adk.core.template.backup import TemplateBackup


@pytest.fixture
def tmp_project(tmp_path: Path) -> Path:
    """Create temporary project structure"""
    # Create basic project structure
    (tmp_path / ".moai").mkdir()
    (tmp_path / ".claude").mkdir()
    (tmp_path / ".github").mkdir()
    (tmp_path / "CLAUDE.md").write_text("# Project")

    # Create config with protected content
    (tmp_path / ".moai" / "config").mkdir()
    (tmp_path / ".moai" / "config" / "config.json").write_text('{"test": "value"}')

    # Create protected directories
    (tmp_path / ".moai" / "specs").mkdir()
    (tmp_path / ".moai" / "specs" / "spec-001.md").write_text("# SPEC-001")
    (tmp_path / ".moai" / "reports").mkdir()
    (tmp_path / ".moai" / "reports" / "report.md").write_text("# Report")

    return tmp_path


class TestTemplateBackupInit:
    """Test TemplateBackup initialization"""

    def test_init_with_absolute_path(self, tmp_path: Path) -> None:
        """Should initialize with absolute path"""
        backup = TemplateBackup(tmp_path)
        assert backup.target_path == tmp_path.resolve()

    def test_init_resolves_relative_path(self, tmp_path: Path) -> None:
        """Should resolve relative paths to absolute"""
        relative = Path(".")
        backup = TemplateBackup(relative)
        assert backup.target_path.is_absolute()

    def test_backup_dir_property(self, tmp_path: Path) -> None:
        """Should return correct backup directory path"""
        backup = TemplateBackup(tmp_path)
        expected = tmp_path / ".moai-backups"
        assert backup.backup_dir == expected


class TestHasExistingFiles:
    """Test has_existing_files detection"""

    def test_has_existing_files_when_moai_exists(self, tmp_path: Path) -> None:
        """Should return True when .moai exists"""
        (tmp_path / ".moai").mkdir()
        backup = TemplateBackup(tmp_path)
        assert backup.has_existing_files() is True

    def test_has_existing_files_when_claude_exists(self, tmp_path: Path) -> None:
        """Should return True when .claude exists"""
        (tmp_path / ".claude").mkdir()
        backup = TemplateBackup(tmp_path)
        assert backup.has_existing_files() is True

    def test_has_existing_files_when_github_exists(self, tmp_path: Path) -> None:
        """Should return True when .github exists"""
        (tmp_path / ".github").mkdir()
        backup = TemplateBackup(tmp_path)
        assert backup.has_existing_files() is True

    def test_has_existing_files_when_claude_md_exists(self, tmp_path: Path) -> None:
        """Should return True when CLAUDE.md exists"""
        (tmp_path / "CLAUDE.md").write_text("# Test")
        backup = TemplateBackup(tmp_path)
        assert backup.has_existing_files() is True

    def test_has_existing_files_when_none_exist(self, tmp_path: Path) -> None:
        """Should return False when no tracked files exist"""
        backup = TemplateBackup(tmp_path)
        assert backup.has_existing_files() is False

    def test_has_existing_files_with_multiple_items(self, tmp_project: Path) -> None:
        """Should return True when multiple items exist"""
        backup = TemplateBackup(tmp_project)
        assert backup.has_existing_files() is True


class TestCreateBackup:
    """Test backup creation"""

    def test_create_backup_creates_directory(self, tmp_project: Path) -> None:
        """Should create backup directory"""
        backup = TemplateBackup(tmp_project)
        backup_path = backup.create_backup()

        assert backup_path.exists()
        assert backup_path.is_dir()
        assert backup_path == tmp_project / ".moai-backups" / "backup"

    def test_create_backup_copies_moai_directory(self, tmp_project: Path) -> None:
        """Should copy .moai directory"""
        backup = TemplateBackup(tmp_project)
        backup_path = backup.create_backup()

        # Check .moai was copied
        assert (backup_path / ".moai").exists()
        assert (backup_path / ".moai" / "config" / "config.json").exists()

    def test_create_backup_copies_claude_directory(self, tmp_project: Path) -> None:
        """Should copy .claude directory"""
        backup = TemplateBackup(tmp_project)
        backup_path = backup.create_backup()

        assert (backup_path / ".claude").exists()

    def test_create_backup_copies_github_directory(self, tmp_project: Path) -> None:
        """Should copy .github directory"""
        backup = TemplateBackup(tmp_project)
        backup_path = backup.create_backup()

        assert (backup_path / ".github").exists()

    def test_create_backup_copies_claude_md(self, tmp_project: Path) -> None:
        """Should copy CLAUDE.md file"""
        backup = TemplateBackup(tmp_project)
        backup_path = backup.create_backup()

        assert (backup_path / "CLAUDE.md").exists()
        assert "# Project" in (backup_path / "CLAUDE.md").read_text()

    def test_create_backup_excludes_specs_directory(self, tmp_project: Path) -> None:
        """Should exclude specs directory from backup"""
        backup = TemplateBackup(tmp_project)
        backup_path = backup.create_backup()

        # specs should not be backed up
        assert not (backup_path / ".moai" / "specs").exists()

    def test_create_backup_excludes_reports_directory(self, tmp_project: Path) -> None:
        """Should exclude reports directory from backup"""
        backup = TemplateBackup(tmp_project)
        backup_path = backup.create_backup()

        # reports should not be backed up
        assert not (backup_path / ".moai" / "reports").exists()

    def test_create_backup_overwrites_existing_backup(self, tmp_project: Path) -> None:
        """Should overwrite existing backup"""
        backup = TemplateBackup(tmp_project)

        # Create first backup
        backup_path = backup.create_backup()
        (backup_path / "marker.txt").write_text("first backup")

        # Create second backup
        backup_path2 = backup.create_backup()

        # Should be same path
        assert backup_path == backup_path2

        # Old marker should not exist
        assert not (backup_path2 / "marker.txt").exists()

    def test_create_backup_skips_nonexistent_items(self, tmp_path: Path) -> None:
        """Should skip items that don't exist"""
        # Only create .moai
        (tmp_path / ".moai").mkdir()
        (tmp_path / ".moai" / "test.txt").write_text("test")

        backup = TemplateBackup(tmp_path)
        backup_path = backup.create_backup()

        # Only .moai should be in backup
        assert (backup_path / ".moai").exists()
        assert not (backup_path / ".claude").exists()
        assert not (backup_path / ".github").exists()
        assert not (backup_path / "CLAUDE.md").exists()


class TestCopyExcludeProtected:
    """Test _copy_exclude_protected private method"""

    def test_copy_exclude_protected_copies_allowed_files(self, tmp_path: Path) -> None:
        """Should copy allowed files"""
        src = tmp_path / "src"
        dst = tmp_path / "dst"
        src.mkdir()

        # Create allowed files
        (src / "config.json").write_text('{"key": "value"}')
        (src / "subdir").mkdir()
        (src / "subdir" / "file.txt").write_text("content")

        backup = TemplateBackup(tmp_path)
        backup._copy_exclude_protected(src, dst)

        assert (dst / "config.json").exists()
        assert (dst / "subdir" / "file.txt").exists()

    def test_copy_exclude_protected_excludes_specs(self, tmp_path: Path) -> None:
        """Should exclude specs directory"""
        src = tmp_path / "src"
        dst = tmp_path / "dst"
        src.mkdir()

        # Create specs directory (should be excluded)
        (src / "specs").mkdir()
        (src / "specs" / "spec-001.md").write_text("# SPEC")

        backup = TemplateBackup(tmp_path)
        backup._copy_exclude_protected(src, dst)

        # specs should not be copied
        assert not (dst / "specs").exists()

    def test_copy_exclude_protected_excludes_reports(self, tmp_path: Path) -> None:
        """Should exclude reports directory"""
        src = tmp_path / "src"
        dst = tmp_path / "dst"
        src.mkdir()

        # Create reports directory (should be excluded)
        (src / "reports").mkdir()
        (src / "reports" / "report.md").write_text("# Report")

        backup = TemplateBackup(tmp_path)
        backup._copy_exclude_protected(src, dst)

        # reports should not be copied
        assert not (dst / "reports").exists()

    def test_copy_exclude_protected_handles_nested_protected(self, tmp_path: Path) -> None:
        """Should exclude top-level protected directories"""
        src = tmp_path / "src"
        dst = tmp_path / "dst"
        src.mkdir()

        # Create top-level protected directory (reports)
        (src / "reports").mkdir()
        (src / "reports" / "nested.md").write_text("# Nested Report")

        # Create allowed directory
        (src / "config").mkdir()
        (src / "config" / "settings.json").write_text("{}")

        backup = TemplateBackup(tmp_path)
        backup._copy_exclude_protected(src, dst)

        # reports (top-level protected) should not be copied
        assert not (dst / "reports").exists()
        # config (allowed) should be copied
        assert (dst / "config").exists()
        assert (dst / "config" / "settings.json").exists()


class TestRestoreBackup:
    """Test backup restoration"""

    def test_restore_backup_from_default_path(self, tmp_project: Path) -> None:
        """Should restore from default backup path"""
        backup = TemplateBackup(tmp_project)
        backup.create_backup()

        # Remove current files
        shutil.rmtree(tmp_project / ".moai")
        shutil.rmtree(tmp_project / ".claude")

        # Restore
        backup.restore_backup()

        # Files should be restored
        assert (tmp_project / ".moai").exists()
        assert (tmp_project / ".claude").exists()

    def test_restore_backup_from_custom_path(self, tmp_project: Path) -> None:
        """Should restore from custom backup path"""
        backup = TemplateBackup(tmp_project)

        # Create backup in custom location
        custom_backup = tmp_project / "custom-backup"
        custom_backup.mkdir()
        (custom_backup / ".moai").mkdir()
        (custom_backup / ".moai" / "test.txt").write_text("custom")

        # Remove current .moai
        shutil.rmtree(tmp_project / ".moai")

        # Restore from custom path
        backup.restore_backup(custom_backup)

        # Should restore custom backup
        assert (tmp_project / ".moai" / "test.txt").exists()
        assert "custom" in (tmp_project / ".moai" / "test.txt").read_text()

    def test_restore_backup_raises_if_not_found(self, tmp_path: Path) -> None:
        """Should raise FileNotFoundError if backup doesn't exist"""
        backup = TemplateBackup(tmp_path)

        with pytest.raises(FileNotFoundError, match="Backup not found"):
            backup.restore_backup()

    def test_restore_backup_replaces_existing_files(self, tmp_project: Path) -> None:
        """Should replace existing files during restore"""
        backup = TemplateBackup(tmp_project)
        backup.create_backup()

        # Modify current files
        (tmp_project / ".moai" / "config" / "config.json").write_text('{"modified": true}')

        # Restore
        backup.restore_backup()

        # Should have original content
        config = (tmp_project / ".moai" / "config" / "config.json").read_text()
        assert '{"test": "value"}' in config
        assert "modified" not in config

    def test_restore_backup_handles_directory_restore(self, tmp_project: Path) -> None:
        """Should restore directories correctly"""
        backup = TemplateBackup(tmp_project)
        backup.create_backup()

        # Remove .moai directory
        shutil.rmtree(tmp_project / ".moai")

        # Restore
        backup.restore_backup()

        # Directory structure should be restored
        assert (tmp_project / ".moai").exists()
        assert (tmp_project / ".moai" / "config").exists()
        assert (tmp_project / ".moai" / "config" / "config.json").exists()

    def test_restore_backup_handles_file_restore(self, tmp_project: Path) -> None:
        """Should restore files correctly"""
        backup = TemplateBackup(tmp_project)
        backup.create_backup()

        # Remove CLAUDE.md
        (tmp_project / "CLAUDE.md").unlink()

        # Restore
        backup.restore_backup()

        # File should be restored
        assert (tmp_project / "CLAUDE.md").exists()
        assert "# Project" in (tmp_project / "CLAUDE.md").read_text()

    def test_restore_backup_skips_missing_items_in_backup(self, tmp_project: Path) -> None:
        """Should skip items not present in backup"""
        backup = TemplateBackup(tmp_project)

        # Create minimal backup
        backup_path = tmp_project / ".moai-backups" / "backup"
        backup_path.mkdir(parents=True)
        (backup_path / ".moai").mkdir()
        (backup_path / ".moai" / "test.txt").write_text("test")

        # Restore
        backup.restore_backup()

        # Only .moai should be restored
        assert (tmp_project / ".moai" / "test.txt").exists()
        # Other items should remain unchanged
        assert (tmp_project / ".claude").exists()
        assert (tmp_project / ".github").exists()


class TestBackupIntegration:
    """Integration tests for full backup workflow"""

    def test_full_backup_restore_cycle(self, tmp_project: Path) -> None:
        """Should complete full backup and restore cycle"""
        backup = TemplateBackup(tmp_project)

        # Create backup
        backup_path = backup.create_backup()
        assert backup_path.exists()

        # Verify backup contents
        assert (backup_path / ".moai" / "config" / "config.json").exists()
        assert (backup_path / "CLAUDE.md").exists()

        # Modify original
        (tmp_project / "CLAUDE.md").write_text("# Modified")

        # Restore
        backup.restore_backup()

        # Should have original content
        assert "# Project" in (tmp_project / "CLAUDE.md").read_text()
        assert "Modified" not in (tmp_project / "CLAUDE.md").read_text()

    def test_multiple_backup_cycles(self, tmp_project: Path) -> None:
        """Should handle multiple backup cycles correctly"""
        backup = TemplateBackup(tmp_project)

        # First backup
        (tmp_project / "CLAUDE.md").write_text("# Version 1")
        backup.create_backup()

        # Second backup (should overwrite)
        (tmp_project / "CLAUDE.md").write_text("# Version 2")
        backup.create_backup()

        # Restore should have Version 2
        (tmp_project / "CLAUDE.md").write_text("# Current")
        backup.restore_backup()

        content = (tmp_project / "CLAUDE.md").read_text()
        assert "# Version 2" in content
        assert "# Version 1" not in content
        assert "# Current" not in content
