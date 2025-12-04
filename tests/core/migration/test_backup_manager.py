"""Tests for backup_manager module."""

import json
import tempfile
from pathlib import Path

import pytest

from moai_adk.core.migration.backup_manager import BackupManager


@pytest.fixture
def temp_project_root():
    """Create temporary project root."""
    with tempfile.TemporaryDirectory() as tmp_dir:
        project_root = Path(tmp_dir)
        # Create basic structure
        moai_dir = project_root / ".moai"
        moai_dir.mkdir()
        (moai_dir / "config").mkdir()

        claude_dir = project_root / ".claude"
        claude_dir.mkdir()

        yield project_root


@pytest.fixture
def backup_manager(temp_project_root):
    """Create BackupManager instance."""
    return BackupManager(temp_project_root)


class TestBackupManager:
    """Test BackupManager functionality."""

    def test_create_backup_with_no_files(self, backup_manager, temp_project_root):
        """Test backup creation when no files exist."""
        # No config files exist
        backup_dir = backup_manager.create_backup("empty_backup")

        assert backup_dir.exists()
        assert (backup_dir / "backup_metadata.json").exists()

        # Check metadata
        with open(backup_dir / "backup_metadata.json", "r") as f:
            metadata = json.load(f)

        assert metadata["description"] == "empty_backup"
        assert len(metadata["backed_up_files"]) == 0

    def test_create_backup_with_files(self, backup_manager, temp_project_root):
        """Test backup creation with actual files."""
        # Create test files
        config_file = temp_project_root / ".moai" / "config" / "config.json"
        config_file.write_text('{"test": "data"}')

        statusline_file = temp_project_root / ".moai" / "config" / "statusline-config.yaml"
        statusline_file.write_text("test: yaml")

        backup_dir = backup_manager.create_backup("test_backup")

        assert backup_dir.exists()
        assert len(list(backup_dir.rglob("*.json"))) >= 1  # At least metadata

        # Verify backed up files
        backed_up_config = backup_dir / ".moai" / "config" / "config.json"
        assert backed_up_config.exists()

    def test_list_backups_empty(self, backup_manager):
        """Test listing backups when none exist."""
        backups = backup_manager.list_backups()
        assert backups == []

    def test_list_backups_with_corrupt_metadata(self, backup_manager, temp_project_root):
        """Test listing backups with corrupted metadata."""
        # Create backup directory with corrupted metadata
        backup_dir = backup_manager.backup_base_dir / "corrupt_20250101_120000"
        backup_dir.mkdir(parents=True)

        metadata_path = backup_dir / "backup_metadata.json"
        metadata_path.write_text("{invalid json")

        backups = backup_manager.list_backups()
        # Should skip corrupted backup
        assert len(backups) == 0

    def test_restore_backup_nonexistent(self, backup_manager, temp_project_root):
        """Test restoring nonexistent backup."""
        fake_path = temp_project_root / "nonexistent"
        result = backup_manager.restore_backup(fake_path)
        assert result is False

    def test_restore_backup_no_metadata(self, backup_manager, temp_project_root):
        """Test restoring backup without metadata."""
        backup_dir = backup_manager.backup_base_dir / "no_metadata"
        backup_dir.mkdir(parents=True)

        result = backup_manager.restore_backup(backup_dir)
        assert result is False

    def test_restore_backup_with_missing_files(self, backup_manager, temp_project_root):
        """Test restoring backup when backed up files are missing."""
        # Create metadata referencing non-existent files
        backup_dir = backup_manager.backup_base_dir / "missing_files"
        backup_dir.mkdir(parents=True)

        metadata = {
            "timestamp": "20250101_120000",
            "description": "test",
            "backed_up_files": [".moai/config/missing.json"],
            "project_root": str(temp_project_root),
        }

        metadata_path = backup_dir / "backup_metadata.json"
        with open(metadata_path, "w") as f:
            json.dump(metadata, f)

        # Should succeed even with missing files
        result = backup_manager.restore_backup(backup_dir)
        assert result is True

    def test_cleanup_old_backups_below_threshold(self, backup_manager):
        """Test cleanup when backup count is below threshold."""
        # Create 3 backups (below default keep_count=5)
        for i in range(3):
            backup_manager.create_backup(f"backup_{i}")

        deleted = backup_manager.cleanup_old_backups(keep_count=5)
        assert deleted == 0

    def test_cleanup_old_backups_above_threshold(self, backup_manager):
        """Test cleanup when backup count exceeds threshold."""
        # Create 7 backups
        for i in range(7):
            backup_manager.create_backup(f"backup_{i}")

        deleted = backup_manager.cleanup_old_backups(keep_count=5)
        assert deleted == 2

        # Verify only 5 remain
        remaining = backup_manager.list_backups()
        assert len(remaining) == 5

    def test_get_latest_backup_none(self, backup_manager):
        """Test getting latest backup when none exist."""
        latest = backup_manager.get_latest_backup()
        assert latest is None

    def test_get_latest_backup_exists(self, backup_manager):
        """Test getting latest backup."""
        backup_manager.create_backup("first")
        backup_manager.create_backup("second")

        latest = backup_manager.get_latest_backup()
        assert latest is not None
        assert "second" in str(latest)

    def test_create_full_project_backup_no_files(self, backup_manager, temp_project_root):
        """Test full project backup with no existing files."""
        backup_dir = backup_manager.create_full_project_backup("full_backup")

        assert backup_dir.exists()
        assert ".moai-backups" in str(backup_dir)

        # Check metadata
        metadata_path = backup_dir / "backup_metadata.json"
        assert metadata_path.exists()

        with open(metadata_path, "r") as f:
            metadata = json.load(f)

        assert metadata["backup_type"] == "full_project"

    def test_create_full_project_backup_with_files(self, backup_manager, temp_project_root):
        """Test full project backup with existing files."""
        # Create test files
        (temp_project_root / ".moai" / "test.txt").write_text("test")
        (temp_project_root / ".claude" / "test.md").write_text("# Test")
        (temp_project_root / "CLAUDE.md").write_text("# Claude")

        backup_dir = backup_manager.create_full_project_backup("full_backup")

        # Verify backed up items
        assert (backup_dir / ".moai" / "test.txt").exists()
        assert (backup_dir / ".claude" / "test.md").exists()
        assert (backup_dir / "CLAUDE.md").exists()
