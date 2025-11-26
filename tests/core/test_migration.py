"""Test migration system functionality.

Tests the automatic migration system that handles breaking changes
between MoAI-ADK versions, specifically config file relocations.
"""

import json
from pathlib import Path
from typing import Any

import pytest

from moai_adk.core.migration import (
    BackupManager,
    FileMigrator,
    VersionDetector,
    VersionMigrator,
)


@pytest.fixture
def mock_v023_project(tmp_path: Path) -> Path:
    """Create a mock v0.23.0 project structure."""
    project_path = tmp_path / "test_project"
    project_path.mkdir()

    # Create v0.23.0 structure
    moai_dir = project_path / ".moai"
    moai_dir.mkdir()

    claude_dir = project_path / ".claude"
    claude_dir.mkdir()

    # Old config.json location
    config_data: dict[str, Any] = {
        "moai": {"version": "0.23.0"},
        "project": {
            "name": "test_project",
            "mode": "personal",
            "language": "python",
        },
    }
    (moai_dir / "config.json").write_text(json.dumps(config_data, indent=2))

    # Old statusline config location
    statusline_data = """moai:
  version: "0.23.0"
statusline:
  enabled: true
"""
    (claude_dir / "statusline-config.yaml").write_text(statusline_data)

    return project_path


@pytest.fixture
def mock_v024_project(tmp_path: Path) -> Path:
    """Create a mock v0.24.0+ project structure."""
    project_path = tmp_path / "test_project"
    project_path.mkdir(exist_ok=True)

    # Create v0.24.0+ structure
    moai_dir = project_path / ".moai"
    moai_dir.mkdir(exist_ok=True)
    config_dir = moai_dir / "config"
    config_dir.mkdir(exist_ok=True)

    # New config.json location
    config_data: dict[str, Any] = {
        "moai": {"version": "0.24.0"},
        "project": {
            "name": "test_project",
            "mode": "personal",
            "language": "python",
        },
    }
    (config_dir / "config.json").write_text(json.dumps(config_data, indent=2))

    # New statusline config location
    statusline_data = """moai:
  version: "0.24.0"
statusline:
  enabled: true
"""
    (config_dir / "statusline-config.yaml").write_text(statusline_data)

    return project_path


class TestVersionDetector:
    """Test VersionDetector functionality."""

    def test_detect_v023_version(self, mock_v023_project: Path) -> None:
        """Test detection of v0.23.0 project."""
        detector = VersionDetector(mock_v023_project)
        version = detector.detect_version()
        assert version == "0.23.0"
        assert detector.needs_migration() is True

    def test_detect_v024_version(self, mock_v024_project: Path) -> None:
        """Test detection of v0.24.0+ project."""
        detector = VersionDetector(mock_v024_project)
        version = detector.detect_version()
        assert version == "0.24.0+"
        assert detector.needs_migration() is False

    def test_get_migration_plan_v023(self, mock_v023_project: Path) -> None:
        """Test migration plan generation for v0.23.0."""
        detector = VersionDetector(mock_v023_project)
        plan = detector.get_migration_plan()

        assert "move" in plan
        assert "cleanup" in plan
        assert len(plan["move"]) == 2  # 2 config files


class TestBackupManager:
    """Test BackupManager functionality."""

    def test_create_backup(self, mock_v023_project: Path) -> None:
        """Test backup creation."""
        manager = BackupManager(mock_v023_project)
        backup_path = manager.create_backup("test migration")

        assert backup_path.exists()
        assert backup_path.is_dir()
        assert (backup_path / ".moai" / "config.json").exists()

    def test_list_backups(self, mock_v023_project: Path) -> None:
        """Test backup listing."""
        manager = BackupManager(mock_v023_project)

        # Create multiple backups
        manager.create_backup("backup 1")
        manager.create_backup("backup 2")

        backups = manager.list_backups()
        assert len(backups) == 2

    def test_cleanup_old_backups(self, mock_v023_project: Path) -> None:
        """Test old backup cleanup."""
        manager = BackupManager(mock_v023_project)

        # Create 7 backups
        for i in range(7):
            manager.create_backup(f"backup {i}")

        # Keep only 5 most recent
        removed = manager.cleanup_old_backups(keep_count=5)
        assert removed == 2

        remaining = manager.list_backups()
        assert len(remaining) == 5


class TestFileMigrator:
    """Test FileMigrator functionality."""

    def test_move_file(self, mock_v023_project: Path) -> None:
        """Test file migration."""
        migrator = FileMigrator(mock_v023_project)

        source = mock_v023_project / ".moai" / "config.json"
        destination = mock_v023_project / ".moai" / "config" / "config.json"

        # Create destination directory
        destination.parent.mkdir(parents=True, exist_ok=True)

        success = migrator.move_file(source, destination)
        assert success is True
        assert destination.exists()
        assert source.exists()  # Original still exists (copy, not move)

    def test_execute_migration_plan(self, mock_v023_project: Path) -> None:
        """Test migration plan execution."""
        detector = VersionDetector(mock_v023_project)
        plan = detector.get_migration_plan()

        migrator = FileMigrator(mock_v023_project)
        result = migrator.execute_migration_plan(plan)

        assert result["success"] is True
        assert result["moved_files"] > 0

    def test_cleanup_old_files(self, mock_v023_project: Path) -> None:
        """Test old file cleanup."""
        migrator = FileMigrator(mock_v023_project)

        cleanup_list = [".moai/config.json", ".claude/statusline-config.yaml"]
        removed = migrator.cleanup_old_files(cleanup_list)

        assert removed == 2
        assert not (mock_v023_project / ".moai" / "config.json").exists()


class TestVersionMigrator:
    """Test VersionMigrator orchestration."""

    def test_needs_migration_v023(self, mock_v023_project: Path) -> None:
        """Test migration need detection for v0.23.0."""
        migrator = VersionMigrator(mock_v023_project)
        assert migrator.needs_migration() is True

    def test_needs_migration_v024(self, mock_v024_project: Path) -> None:
        """Test migration need detection for v0.24.0+."""
        migrator = VersionMigrator(mock_v024_project)
        assert migrator.needs_migration() is False

    def test_get_migration_info(self, mock_v023_project: Path) -> None:
        """Test migration info retrieval."""
        migrator = VersionMigrator(mock_v023_project)
        info = migrator.get_migration_info()

        assert info["current_version"] == "0.23.0"
        assert info["target_version"] == "0.24.0"
        assert info["needs_migration"] is True
        assert info["file_count"] > 0

    def test_migrate_to_v024_dry_run(self, mock_v023_project: Path) -> None:
        """Test dry run migration (no actual changes)."""
        migrator = VersionMigrator(mock_v023_project)
        success = migrator.migrate_to_v024(dry_run=True, cleanup=False)

        assert success is True
        # Original files should still exist in dry run
        assert (mock_v023_project / ".moai" / "config.json").exists()

    def test_migrate_to_v024_full(self, mock_v023_project: Path) -> None:
        """Test full migration with cleanup."""
        migrator = VersionMigrator(mock_v023_project)
        success = migrator.migrate_to_v024(dry_run=False, cleanup=True)

        assert success is True

        # New locations should exist
        new_config = mock_v023_project / ".moai" / "config" / "config.json"
        assert new_config.exists()

        # Old locations should be cleaned up
        old_config = mock_v023_project / ".moai" / "config.json"
        assert not old_config.exists()

        # Backup should exist
        backups_dir = mock_v023_project / ".moai" / "backups"
        assert backups_dir.exists()
        backups = list(backups_dir.iterdir())
        assert len(backups) > 0

    def test_rollback_to_latest_backup(self, mock_v023_project: Path) -> None:
        """Test rollback functionality."""
        migrator = VersionMigrator(mock_v023_project)

        # Create backup
        migrator.backup_manager.create_backup("before migration")

        # Modify project
        (mock_v023_project / ".moai" / "config.json").unlink()

        # Rollback
        success = migrator.rollback_to_latest_backup()
        assert success is True

        # File should be restored
        assert (mock_v023_project / ".moai" / "config.json").exists()
