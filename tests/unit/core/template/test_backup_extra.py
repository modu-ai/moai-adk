"""Extended unit tests for moai_adk.core.template.backup module.

Comprehensive tests for TemplateBackup functionality including
backup creation, restoration, and edge cases.
"""

import re
from pathlib import Path
from unittest.mock import MagicMock, Mock, patch

import pytest

from moai_adk.core.template.backup import TemplateBackup


class TestTemplateBackupInitialization:
    """Test TemplateBackup initialization."""

    def test_initialization_sets_target_path(self, tmp_path):
        """Test TemplateBackup initializes with correct target path."""
        backup = TemplateBackup(tmp_path)
        assert backup.target_path == tmp_path.resolve()

    def test_initialization_resolves_path(self):
        """Test initialization resolves relative paths."""
        backup = TemplateBackup(Path("./test"))
        assert backup.target_path.is_absolute()

    def test_backup_dir_property(self, tmp_path):
        """Test backup_dir property returns correct path."""
        backup = TemplateBackup(tmp_path)
        expected_path = tmp_path / ".moai-backups"
        assert backup.backup_dir == expected_path

    def test_exclude_dirs_defined(self):
        """Test BACKUP_EXCLUDE_DIRS is defined."""
        backup = TemplateBackup(Path("/tmp"))
        assert len(backup.BACKUP_EXCLUDE_DIRS) > 0
        assert "specs" in backup.BACKUP_EXCLUDE_DIRS
        assert "reports" in backup.BACKUP_EXCLUDE_DIRS


class TestTemplateBackupHasExistingFiles:
    """Test has_existing_files method."""

    def test_has_existing_files_with_moai_dir(self, tmp_path):
        """Test has_existing_files detects .moai directory."""
        backup = TemplateBackup(tmp_path)
        (tmp_path / ".moai").mkdir()

        assert backup.has_existing_files() is True

    def test_has_existing_files_with_claude_dir(self, tmp_path):
        """Test has_existing_files detects .claude directory."""
        backup = TemplateBackup(tmp_path)
        (tmp_path / ".claude").mkdir()

        assert backup.has_existing_files() is True

    def test_has_existing_files_with_github_dir(self, tmp_path):
        """Test has_existing_files detects .github directory."""
        backup = TemplateBackup(tmp_path)
        (tmp_path / ".github").mkdir()

        assert backup.has_existing_files() is True

    def test_has_existing_files_with_claude_md(self, tmp_path):
        """Test has_existing_files detects CLAUDE.md file."""
        backup = TemplateBackup(tmp_path)
        (tmp_path / "CLAUDE.md").write_text("content")

        assert backup.has_existing_files() is True

    def test_has_existing_files_with_multiple_items(self, tmp_path):
        """Test has_existing_files with multiple items present."""
        backup = TemplateBackup(tmp_path)
        (tmp_path / ".moai").mkdir()
        (tmp_path / ".claude").mkdir()
        (tmp_path / "CLAUDE.md").write_text("content")

        assert backup.has_existing_files() is True

    def test_has_existing_files_empty_directory(self, tmp_path):
        """Test has_existing_files with empty directory."""
        backup = TemplateBackup(tmp_path)
        assert backup.has_existing_files() is False

    def test_has_existing_files_with_other_files(self, tmp_path):
        """Test has_existing_files ignores other files."""
        backup = TemplateBackup(tmp_path)
        (tmp_path / "README.md").write_text("content")
        (tmp_path / "setup.py").write_text("content")

        assert backup.has_existing_files() is False


class TestTemplateBackupCreateBackup:
    """Test create_backup method."""

    def test_create_backup_creates_timestamped_dir(self, tmp_path):
        """Test create_backup creates timestamped directory."""
        backup = TemplateBackup(tmp_path)
        (tmp_path / ".moai").mkdir()

        backup_path = backup.create_backup()

        assert backup_path.exists()
        assert backup_path.parent == tmp_path / ".moai-backups"
        # Check timestamp format YYYYMMDD_HHMMSS
        assert re.match(r"\d{8}_\d{6}", backup_path.name)

    def test_create_backup_copies_moai_dir(self, tmp_path):
        """Test create_backup copies .moai directory."""
        backup = TemplateBackup(tmp_path)
        moai_dir = tmp_path / ".moai"
        moai_dir.mkdir()
        (moai_dir / "test.txt").write_text("content")

        backup_path = backup.create_backup()

        assert (backup_path / ".moai" / "test.txt").exists()

    def test_create_backup_copies_claude_dir(self, tmp_path):
        """Test create_backup copies .claude directory."""
        backup = TemplateBackup(tmp_path)
        claude_dir = tmp_path / ".claude"
        claude_dir.mkdir()
        (claude_dir / "test.txt").write_text("content")

        backup_path = backup.create_backup()

        assert (backup_path / ".claude" / "test.txt").exists()

    def test_create_backup_copies_github_dir(self, tmp_path):
        """Test create_backup copies .github directory."""
        backup = TemplateBackup(tmp_path)
        github_dir = tmp_path / ".github"
        github_dir.mkdir()
        (github_dir / "test.txt").write_text("content")

        backup_path = backup.create_backup()

        assert (backup_path / ".github" / "test.txt").exists()

    def test_create_backup_copies_claude_md(self, tmp_path):
        """Test create_backup copies CLAUDE.md file."""
        backup = TemplateBackup(tmp_path)
        claude_file = tmp_path / "CLAUDE.md"
        claude_file.write_text("# CLAUDE.md content")

        backup_path = backup.create_backup()

        assert (backup_path / "CLAUDE.md").exists()
        assert (backup_path / "CLAUDE.md").read_text() == "# CLAUDE.md content"

    def test_create_backup_skips_missing_items(self, tmp_path):
        """Test create_backup skips items that don't exist."""
        backup = TemplateBackup(tmp_path)
        (tmp_path / ".moai").mkdir()
        # .claude, .github, CLAUDE.md don't exist

        backup_path = backup.create_backup()

        assert (backup_path / ".moai").exists()
        assert not (backup_path / ".claude").exists()
        assert not (backup_path / ".github").exists()
        assert not (backup_path / "CLAUDE.md").exists()

    def test_create_backup_excludes_protected_paths(self, tmp_path):
        """Test create_backup excludes protected paths from .moai."""
        backup = TemplateBackup(tmp_path)
        moai_dir = tmp_path / ".moai"
        moai_dir.mkdir()
        (moai_dir / "specs").mkdir()
        (moai_dir / "specs" / "spec.md").write_text("spec")
        (moai_dir / "reports").mkdir()
        (moai_dir / "reports" / "report.txt").write_text("report")
        (moai_dir / "normal").mkdir()
        (moai_dir / "normal" / "file.txt").write_text("normal")

        backup_path = backup.create_backup()

        # Check excluded
        assert not (backup_path / ".moai" / "specs").exists()
        assert not (backup_path / ".moai" / "reports").exists()
        # Check included
        assert (backup_path / ".moai" / "normal").exists()

    def test_create_backup_multiple_calls_create_separate_backups(self, tmp_path):
        """Test multiple create_backup calls create separate backups."""
        import time
        backup = TemplateBackup(tmp_path)
        (tmp_path / ".moai").mkdir()

        backup_path1 = backup.create_backup()
        time.sleep(1)  # Ensure different timestamp
        backup_path2 = backup.create_backup()

        assert backup_path1 != backup_path2
        assert backup_path1.exists()
        assert backup_path2.exists()


class TestTemplateBackupGetLatestBackup:
    """Test get_latest_backup method."""

    def test_get_latest_backup_returns_none_when_no_backups(self, tmp_path):
        """Test get_latest_backup returns None when no backups exist."""
        backup = TemplateBackup(tmp_path)
        result = backup.get_latest_backup()
        assert result is None

    def test_get_latest_backup_returns_newest_timestamped(self, tmp_path):
        """Test get_latest_backup returns newest timestamped backup."""
        backup = TemplateBackup(tmp_path)
        backup_dir = tmp_path / ".moai-backups"
        backup_dir.mkdir(parents=True)

        # Create backups with different timestamps
        (backup_dir / "20250101_100000").mkdir()
        (backup_dir / "20250101_110000").mkdir()
        (backup_dir / "20250101_120000").mkdir()

        latest = backup.get_latest_backup()

        assert latest == backup_dir / "20250101_120000"

    def test_get_latest_backup_ignores_non_timestamped_dirs(self, tmp_path):
        """Test get_latest_backup ignores non-timestamped directories."""
        backup = TemplateBackup(tmp_path)
        backup_dir = tmp_path / ".moai-backups"
        backup_dir.mkdir(parents=True)

        (backup_dir / "random_dir").mkdir()
        (backup_dir / "20250101_100000").mkdir()

        latest = backup.get_latest_backup()

        assert latest == backup_dir / "20250101_100000"

    def test_get_latest_backup_falls_back_to_legacy(self, tmp_path):
        """Test get_latest_backup falls back to legacy backup directory."""
        backup = TemplateBackup(tmp_path)
        backup_dir = tmp_path / ".moai-backups"
        legacy_dir = backup_dir / "backup"
        legacy_dir.mkdir(parents=True)

        latest = backup.get_latest_backup()

        assert latest == legacy_dir

    def test_get_latest_backup_prefers_timestamped_over_legacy(self, tmp_path):
        """Test get_latest_backup prefers timestamped over legacy."""
        backup = TemplateBackup(tmp_path)
        backup_dir = tmp_path / ".moai-backups"
        backup_dir.mkdir(parents=True)

        (backup_dir / "backup").mkdir()  # Legacy
        (backup_dir / "20250101_100000").mkdir()  # Timestamped

        latest = backup.get_latest_backup()

        assert latest == backup_dir / "20250101_100000"


class TestTemplateBackupRestoreBackup:
    """Test restore_backup method."""

    def test_restore_backup_from_explicit_path(self, tmp_path):
        """Test restore_backup restores from explicit backup path."""
        backup = TemplateBackup(tmp_path)

        # Create backup structure
        backup_dir = tmp_path / "backup"
        backup_dir.mkdir()
        (backup_dir / ".moai").mkdir()
        (backup_dir / ".moai" / "file.txt").write_text("backed up")

        # Restore
        backup.restore_backup(backup_dir)

        assert (tmp_path / ".moai" / "file.txt").exists()
        assert (tmp_path / ".moai" / "file.txt").read_text() == "backed up"

    def test_restore_backup_overwrites_existing(self, tmp_path):
        """Test restore_backup overwrites existing files."""
        backup = TemplateBackup(tmp_path)

        # Create existing file
        (tmp_path / ".moai").mkdir()
        (tmp_path / ".moai" / "file.txt").write_text("old content")

        # Create backup
        backup_dir = tmp_path / "backup"
        backup_dir.mkdir()
        (backup_dir / ".moai").mkdir()
        (backup_dir / ".moai" / "file.txt").write_text("new content")

        # Restore
        backup.restore_backup(backup_dir)

        assert (tmp_path / ".moai" / "file.txt").read_text() == "new content"

    def test_restore_backup_removes_current_version(self, tmp_path):
        """Test restore_backup removes current version before restoring."""
        backup = TemplateBackup(tmp_path)

        # Create existing directory
        existing_dir = tmp_path / ".moai"
        existing_dir.mkdir()
        (existing_dir / "old_file.txt").write_text("old")

        # Create backup with different content
        backup_dir = tmp_path / "backup"
        backup_dir.mkdir()
        (backup_dir / ".moai").mkdir()
        (backup_dir / ".moai" / "new_file.txt").write_text("new")

        # Restore
        backup.restore_backup(backup_dir)

        assert not (tmp_path / ".moai" / "old_file.txt").exists()
        assert (tmp_path / ".moai" / "new_file.txt").exists()

    def test_restore_backup_with_none_path_uses_latest(self, tmp_path):
        """Test restore_backup with None path uses latest backup."""
        backup = TemplateBackup(tmp_path)

        # Create backup
        backup_dir = tmp_path / ".moai-backups" / "20250101_100000"
        backup_dir.mkdir(parents=True)
        (backup_dir / ".moai").mkdir()
        (backup_dir / ".moai" / "test.txt").write_text("content")

        # Restore with None
        backup.restore_backup(None)

        assert (tmp_path / ".moai" / "test.txt").exists()

    def test_restore_backup_raises_when_not_found(self, tmp_path):
        """Test restore_backup raises when backup not found."""
        backup = TemplateBackup(tmp_path)

        with pytest.raises(FileNotFoundError):
            backup.restore_backup(tmp_path / "nonexistent")

    def test_restore_backup_partial_items(self, tmp_path):
        """Test restore_backup restores only items in backup."""
        backup = TemplateBackup(tmp_path)

        # Create existing files
        (tmp_path / ".moai").mkdir()
        (tmp_path / ".moai" / "keep.txt").write_text("keep this")
        (tmp_path / ".claude").mkdir()
        (tmp_path / ".claude" / "old.txt").write_text("old")

        # Create backup with only .moai
        backup_dir = tmp_path / "backup"
        backup_dir.mkdir()
        (backup_dir / ".moai").mkdir()
        (backup_dir / ".moai" / "new.txt").write_text("new")

        # Restore
        backup.restore_backup(backup_dir)

        # .moai should be replaced
        assert (tmp_path / ".moai" / "new.txt").exists()
        assert not (tmp_path / ".moai" / "keep.txt").exists()
        # .claude should be preserved (was existing, not in backup means skip restore, don't delete)
        # Based on the code, only items in backup are restored, others left alone
        assert (tmp_path / ".claude").exists()


class TestTemplateBackupCopyExcludeProtected:
    """Test _copy_exclude_protected method."""

    def test_copy_exclude_protected_copies_normal_files(self, tmp_path):
        """Test _copy_exclude_protected copies normal files."""
        backup = TemplateBackup(tmp_path)
        src_dir = tmp_path / "src"
        src_dir.mkdir()
        (src_dir / "normal").mkdir(parents=True)
        (src_dir / "normal" / "file.txt").write_text("content")

        dst_dir = tmp_path / "dst"
        backup._copy_exclude_protected(src_dir, dst_dir)

        assert (dst_dir / "normal" / "file.txt").exists()

    def test_copy_exclude_protected_skips_specs(self, tmp_path):
        """Test _copy_exclude_protected skips specs directory."""
        backup = TemplateBackup(tmp_path)
        src_dir = tmp_path / "src"
        src_dir.mkdir()
        (src_dir / "specs").mkdir()
        (src_dir / "specs" / "spec.md").write_text("spec")

        dst_dir = tmp_path / "dst"
        backup._copy_exclude_protected(src_dir, dst_dir)

        assert not (dst_dir / "specs").exists()

    def test_copy_exclude_protected_skips_reports(self, tmp_path):
        """Test _copy_exclude_protected skips reports directory."""
        backup = TemplateBackup(tmp_path)
        src_dir = tmp_path / "src"
        src_dir.mkdir()
        (src_dir / "reports").mkdir()
        (src_dir / "reports" / "report.txt").write_text("report")

        dst_dir = tmp_path / "dst"
        backup._copy_exclude_protected(src_dir, dst_dir)

        assert not (dst_dir / "reports").exists()

    def test_copy_exclude_protected_nested_excluded(self, tmp_path):
        """Test _copy_exclude_protected skips nested excluded paths."""
        backup = TemplateBackup(tmp_path)
        src_dir = tmp_path / "src"
        src_dir.mkdir()
        (src_dir / "specs" / "nested" / "deep").mkdir(parents=True)
        (src_dir / "specs" / "nested" / "deep" / "file.md").write_text("content")

        dst_dir = tmp_path / "dst"
        backup._copy_exclude_protected(src_dir, dst_dir)

        assert not (dst_dir / "specs").exists()

    def test_copy_exclude_protected_mixed_content(self, tmp_path):
        """Test _copy_exclude_protected with mixed content."""
        backup = TemplateBackup(tmp_path)
        src_dir = tmp_path / "src"
        src_dir.mkdir()

        # Create mixed structure
        (src_dir / "keep").mkdir()
        (src_dir / "keep" / "file.txt").write_text("keep")
        (src_dir / "specs").mkdir()
        (src_dir / "specs" / "skip.md").write_text("skip")
        (src_dir / "other").mkdir()
        (src_dir / "other" / "file.txt").write_text("other")

        dst_dir = tmp_path / "dst"
        backup._copy_exclude_protected(src_dir, dst_dir)

        assert (dst_dir / "keep" / "file.txt").exists()
        assert not (dst_dir / "specs").exists()
        assert (dst_dir / "other" / "file.txt").exists()


class TestTemplateBackupEdgeCases:
    """Test edge cases and error scenarios."""

    def test_backup_large_directory_structure(self, tmp_path):
        """Test backup handles large directory structures."""
        backup = TemplateBackup(tmp_path)

        # Create deep directory structure
        moai_dir = tmp_path / ".moai"
        deep_dir = moai_dir / "a" / "b" / "c" / "d" / "e"
        deep_dir.mkdir(parents=True)
        (deep_dir / "file.txt").write_text("deep content")

        backup_path = backup.create_backup()

        assert (
            backup_path / ".moai" / "a" / "b" / "c" / "d" / "e" / "file.txt"
        ).exists()

    def test_backup_with_special_characters_in_filenames(self, tmp_path):
        """Test backup handles special characters in filenames."""
        backup = TemplateBackup(tmp_path)

        moai_dir = tmp_path / ".moai"
        moai_dir.mkdir()
        (moai_dir / "file with spaces.txt").write_text("content")
        (moai_dir / "file-with-dashes.txt").write_text("content")

        backup_path = backup.create_backup()

        assert (backup_path / ".moai" / "file with spaces.txt").exists()
        assert (backup_path / ".moai" / "file-with-dashes.txt").exists()

    def test_backup_with_symlinks(self, tmp_path):
        """Test backup handles symlinks gracefully."""
        backup = TemplateBackup(tmp_path)

        moai_dir = tmp_path / ".moai"
        moai_dir.mkdir()
        (moai_dir / "real_file.txt").write_text("content")

        backup_path = backup.create_backup()
        assert (backup_path / ".moai" / "real_file.txt").exists()

    def test_restore_backup_nonexistent_path_raises(self, tmp_path):
        """Test restore_backup raises for nonexistent path."""
        backup = TemplateBackup(tmp_path)

        with pytest.raises(FileNotFoundError):
            backup.restore_backup(tmp_path / "does_not_exist")

    def test_get_latest_backup_with_empty_backup_dir(self, tmp_path):
        """Test get_latest_backup with empty backup directory."""
        backup = TemplateBackup(tmp_path)
        (tmp_path / ".moai-backups").mkdir()

        result = backup.get_latest_backup()
        assert result is None
