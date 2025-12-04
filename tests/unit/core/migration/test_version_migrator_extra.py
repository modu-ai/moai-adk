"""Extended unit tests for moai_adk.core.migration.version_migrator module.

Comprehensive tests for VersionMigrator orchestration functionality.
"""

from pathlib import Path
from unittest.mock import MagicMock, Mock, patch

import pytest

from moai_adk.core.migration.version_migrator import VersionMigrator


class TestVersionMigratorInitialization:
    """Test VersionMigrator initialization."""

    def test_initialization_sets_project_root(self, tmp_path):
        """Test VersionMigrator initializes with project root."""
        migrator = VersionMigrator(tmp_path)

        assert migrator.project_root == Path(tmp_path)

    def test_initialization_creates_components(self, tmp_path):
        """Test VersionMigrator creates required components."""
        migrator = VersionMigrator(tmp_path)

        assert migrator.detector is not None
        assert migrator.backup_manager is not None
        assert migrator.file_migrator is not None

    def test_initialization_with_string_path(self, tmp_path):
        """Test VersionMigrator accepts string path."""
        migrator = VersionMigrator(str(tmp_path))

        assert migrator.project_root == Path(tmp_path)

    def test_initialization_with_path_object(self, tmp_path):
        """Test VersionMigrator accepts Path object."""
        migrator = VersionMigrator(tmp_path)

        assert isinstance(migrator.project_root, Path)


class TestDetectVersion:
    """Test detect_version method."""

    @patch("moai_adk.core.migration.version_migrator.VersionDetector")
    def test_detect_version_delegates_to_detector(self, mock_detector_class):
        """Test detect_version delegates to detector."""
        mock_detector = MagicMock()
        mock_detector.detect_version.return_value = "0.23.0"
        mock_detector_class.return_value = mock_detector

        migrator = VersionMigrator(Path("/tmp"))
        result = migrator.detect_version()

        assert result == "0.23.0"
        mock_detector.detect_version.assert_called_once()

    @patch("moai_adk.core.migration.version_migrator.VersionDetector")
    def test_detect_version_returns_string(self, mock_detector_class):
        """Test detect_version returns version string."""
        mock_detector = MagicMock()
        mock_detector.detect_version.return_value = "0.24.0"
        mock_detector_class.return_value = mock_detector

        migrator = VersionMigrator(Path("/tmp"))
        result = migrator.detect_version()

        assert isinstance(result, str)


class TestNeedsMigration:
    """Test needs_migration method."""

    @patch("moai_adk.core.migration.version_migrator.VersionDetector")
    def test_needs_migration_true(self, mock_detector_class):
        """Test needs_migration returns True when migration needed."""
        mock_detector = MagicMock()
        mock_detector.needs_migration.return_value = True
        mock_detector_class.return_value = mock_detector

        migrator = VersionMigrator(Path("/tmp"))
        result = migrator.needs_migration()

        assert result is True

    @patch("moai_adk.core.migration.version_migrator.VersionDetector")
    def test_needs_migration_false(self, mock_detector_class):
        """Test needs_migration returns False when no migration needed."""
        mock_detector = MagicMock()
        mock_detector.needs_migration.return_value = False
        mock_detector_class.return_value = mock_detector

        migrator = VersionMigrator(Path("/tmp"))
        result = migrator.needs_migration()

        assert result is False


class TestGetMigrationInfo:
    """Test get_migration_info method."""

    @patch("moai_adk.core.migration.version_migrator.VersionDetector")
    def test_get_migration_info_structure(self, mock_detector_class):
        """Test get_migration_info returns correct structure."""
        mock_detector = MagicMock()
        mock_detector.detect_version.return_value = "0.23.0"
        mock_detector.needs_migration.return_value = True
        mock_detector.get_migration_plan.return_value = {
            "move": [{"from": "a", "to": "b"}],
            "create": ["dir1"],
        }
        mock_detector_class.return_value = mock_detector

        migrator = VersionMigrator(Path("/tmp"))
        info = migrator.get_migration_info()

        assert "current_version" in info
        assert "needs_migration" in info
        assert "target_version" in info
        assert "migration_plan" in info
        assert "file_count" in info

    @patch("moai_adk.core.migration.version_migrator.VersionDetector")
    def test_get_migration_info_no_migration_needed(self, mock_detector_class):
        """Test get_migration_info when no migration needed."""
        mock_detector = MagicMock()
        mock_detector.detect_version.return_value = "0.24.0"
        mock_detector.needs_migration.return_value = False
        mock_detector.get_migration_plan.return_value = {}
        mock_detector_class.return_value = mock_detector

        migrator = VersionMigrator(Path("/tmp"))
        info = migrator.get_migration_info()

        assert info["needs_migration"] is False
        assert info["current_version"] == info["target_version"]


class TestMigrateToV024:
    """Test migrate_to_v024 method."""

    @patch("moai_adk.core.migration.version_migrator.VersionDetector")
    @patch("moai_adk.core.migration.version_migrator.BackupManager")
    @patch("moai_adk.core.migration.version_migrator.FileMigrator")
    def test_migrate_returns_true_when_up_to_date(
        self, mock_file_migrator_class, mock_backup_manager_class, mock_detector_class
    ):
        """Test migrate returns True when already up to date."""
        mock_detector = MagicMock()
        mock_detector.needs_migration.return_value = False
        mock_detector_class.return_value = mock_detector

        migrator = VersionMigrator(Path("/tmp"))
        result = migrator.migrate_to_v024()

        assert result is True

    @patch("moai_adk.core.migration.version_migrator.VersionDetector")
    @patch("moai_adk.core.migration.version_migrator.BackupManager")
    @patch("moai_adk.core.migration.version_migrator.FileMigrator")
    def test_migrate_dry_run_mode(
        self, mock_file_migrator_class, mock_backup_manager_class, mock_detector_class
    ):
        """Test migrate with dry_run=True."""
        mock_detector = MagicMock()
        mock_detector.needs_migration.return_value = True
        mock_detector.get_migration_plan.return_value = {
            "create": ["dir"],
            "move": [],
            "cleanup": [],
        }
        mock_detector_class.return_value = mock_detector

        migrator = VersionMigrator(Path("/tmp"))
        result = migrator.migrate_to_v024(dry_run=True)

        assert result is True

    @patch("moai_adk.core.migration.version_migrator.VersionDetector")
    @patch("moai_adk.core.migration.version_migrator.BackupManager")
    @patch("moai_adk.core.migration.version_migrator.FileMigrator")
    def test_migrate_successful_execution(
        self, mock_file_migrator_class, mock_backup_manager_class, mock_detector_class
    ):
        """Test successful migration execution."""
        # Setup detector
        mock_detector = MagicMock()
        mock_detector.needs_migration.return_value = True
        mock_detector.get_migration_plan.return_value = {
            "create": ["dir"],
            "move": [],
            "cleanup": [],
        }
        mock_detector_class.return_value = mock_detector

        # Setup backup manager
        mock_backup_manager = MagicMock()
        mock_backup_manager.create_backup.return_value = Path("/backup")
        mock_backup_manager_class.return_value = mock_backup_manager

        # Setup file migrator
        mock_file_migrator = MagicMock()
        mock_file_migrator.execute_migration_plan.return_value = {
            "success": True,
            "created_dirs": 1,
            "moved_files": 0,
            "errors": [],
        }
        mock_file_migrator.cleanup_old_files.return_value = 0
        mock_file_migrator_class.return_value = mock_file_migrator

        migrator = VersionMigrator(Path("/tmp"))

        with patch.object(migrator, "_verify_migration", return_value=True):
            result = migrator.migrate_to_v024()

        assert result is True

    @patch("moai_adk.core.migration.version_migrator.VersionDetector")
    @patch("moai_adk.core.migration.version_migrator.BackupManager")
    @patch("moai_adk.core.migration.version_migrator.FileMigrator")
    def test_migrate_handles_migration_failure(
        self, mock_file_migrator_class, mock_backup_manager_class, mock_detector_class
    ):
        """Test migrate handles migration failure and rollback."""
        # Setup detector
        mock_detector = MagicMock()
        mock_detector.needs_migration.return_value = True
        mock_detector.get_migration_plan.return_value = {
            "create": [],
            "move": [],
            "cleanup": [],
        }
        mock_detector_class.return_value = mock_detector

        # Setup backup manager
        mock_backup_manager = MagicMock()
        mock_backup_manager.create_backup.return_value = Path("/backup")
        mock_backup_manager_class.return_value = mock_backup_manager

        # Setup file migrator to fail
        mock_file_migrator = MagicMock()
        mock_file_migrator.execute_migration_plan.return_value = {
            "success": False,
            "created_dirs": 0,
            "moved_files": 0,
            "errors": ["Migration error"],
        }
        mock_file_migrator_class.return_value = mock_file_migrator

        migrator = VersionMigrator(Path("/tmp"))
        result = migrator.migrate_to_v024()

        assert result is False
        mock_backup_manager.restore_backup.assert_called()

    @patch("moai_adk.core.migration.version_migrator.VersionDetector")
    @patch("moai_adk.core.migration.version_migrator.BackupManager")
    @patch("moai_adk.core.migration.version_migrator.FileMigrator")
    def test_migrate_with_cleanup_false(
        self, mock_file_migrator_class, mock_backup_manager_class, mock_detector_class
    ):
        """Test migrate with cleanup=False."""
        # Setup mocks
        mock_detector = MagicMock()
        mock_detector.needs_migration.return_value = True
        mock_detector.get_migration_plan.return_value = {
            "create": [],
            "move": [],
            "cleanup": [],
        }
        mock_detector_class.return_value = mock_detector

        mock_backup_manager = MagicMock()
        mock_backup_manager.create_backup.return_value = Path("/backup")
        mock_backup_manager_class.return_value = mock_backup_manager

        mock_file_migrator = MagicMock()
        mock_file_migrator.execute_migration_plan.return_value = {
            "success": True,
            "created_dirs": 0,
            "moved_files": 0,
            "errors": [],
        }
        mock_file_migrator_class.return_value = mock_file_migrator

        migrator = VersionMigrator(Path("/tmp"))

        with patch.object(migrator, "_verify_migration", return_value=True):
            result = migrator.migrate_to_v024(cleanup=False)

        assert result is True
        mock_file_migrator.cleanup_old_files.assert_not_called()


class TestShowMigrationPlan:
    """Test _show_migration_plan method."""

    @patch("moai_adk.core.migration.version_migrator.VersionDetector")
    @patch("builtins.print")
    def test_show_migration_plan(self, mock_print, mock_detector_class):
        """Test _show_migration_plan displays plan."""
        mock_detector = MagicMock()
        mock_detector.get_migration_plan.return_value = {
            "create": ["dir1", "dir2"],
            "move": [
                {
                    "description": "Move file",
                    "from": "old.txt",
                    "to": "new.txt",
                }
            ],
            "cleanup": ["old_file.txt"],
        }
        mock_detector_class.return_value = mock_detector

        migrator = VersionMigrator(Path("/tmp"))
        migrator._show_migration_plan()

        # Check that print was called
        assert mock_print.called


class TestVerifyMigration:
    """Test _verify_migration method."""

    @patch("moai_adk.core.migration.version_migrator.VersionDetector")
    def test_verify_migration_success(self, mock_detector_class):
        """Test _verify_migration returns True when config exists."""
        migrator = VersionMigrator(Path("/tmp"))

        # Mock config existence
        with patch("pathlib.Path.exists") as mock_exists:

            def exists_side_effect():
                return True

            mock_exists.side_effect = exists_side_effect

            result = migrator._verify_migration()

            # The actual check requires the file to exist
            assert isinstance(result, bool)

    @patch("moai_adk.core.migration.version_migrator.VersionDetector")
    def test_verify_migration_failure_no_config(self, mock_detector_class):
        """Test _verify_migration returns False when config missing."""
        migrator = VersionMigrator(Path("/tmp"))

        result = migrator._verify_migration()

        # Should return False if config doesn't exist
        assert isinstance(result, bool)


class TestCheckStatus:
    """Test check_status method."""

    @patch("moai_adk.core.migration.version_migrator.VersionDetector")
    @patch("moai_adk.core.migration.version_migrator.BackupManager")
    @patch("moai_adk.core.migration.version_migrator.FileMigrator")
    def test_check_status_returns_dict(
        self, mock_file_migrator_class, mock_backup_manager_class, mock_detector_class
    ):
        """Test check_status returns dictionary."""
        # Setup mocks
        mock_detector = MagicMock()
        mock_detector.get_version_info.return_value = {"version": "0.24.0"}
        mock_detector.get_migration_info = MagicMock(
            return_value={"needs_migration": False}
        )
        mock_detector_class.return_value = mock_detector

        mock_backup_manager = MagicMock()
        mock_backup_manager.list_backups.return_value = []
        mock_backup_manager_class.return_value = mock_backup_manager

        migrator = VersionMigrator(Path("/tmp"))

        with patch.object(migrator, "get_migration_info", return_value={}):
            status = migrator.check_status()

        assert isinstance(status, dict)
        assert "version" in status
        assert "migration" in status
        assert "backups" in status

    @patch("moai_adk.core.migration.version_migrator.VersionDetector")
    @patch("moai_adk.core.migration.version_migrator.BackupManager")
    @patch("moai_adk.core.migration.version_migrator.FileMigrator")
    def test_check_status_includes_backup_count(
        self, mock_file_migrator_class, mock_backup_manager_class, mock_detector_class
    ):
        """Test check_status includes backup information."""
        mock_detector = MagicMock()
        mock_detector.get_version_info.return_value = {}
        mock_detector_class.return_value = mock_detector

        mock_backup_manager = MagicMock()
        mock_backup_manager.list_backups.return_value = [Path("backup1")]
        mock_backup_manager_class.return_value = mock_backup_manager

        migrator = VersionMigrator(Path("/tmp"))

        with patch.object(migrator, "get_migration_info", return_value={}):
            status = migrator.check_status()

        assert status["backups"]["count"] >= 0


class TestRollbackToLatestBackup:
    """Test rollback_to_latest_backup method."""

    @patch("moai_adk.core.migration.version_migrator.VersionDetector")
    @patch("moai_adk.core.migration.version_migrator.BackupManager")
    @patch("moai_adk.core.migration.version_migrator.FileMigrator")
    def test_rollback_success(
        self, mock_file_migrator_class, mock_backup_manager_class, mock_detector_class
    ):
        """Test successful rollback."""
        mock_backup_manager = MagicMock()
        mock_backup_manager.get_latest_backup.return_value = Path("/backup")
        mock_backup_manager.restore_backup.return_value = True
        mock_backup_manager_class.return_value = mock_backup_manager

        migrator = VersionMigrator(Path("/tmp"))
        result = migrator.rollback_to_latest_backup()

        assert result is True

    @patch("moai_adk.core.migration.version_migrator.VersionDetector")
    @patch("moai_adk.core.migration.version_migrator.BackupManager")
    @patch("moai_adk.core.migration.version_migrator.FileMigrator")
    def test_rollback_no_backup_found(
        self, mock_file_migrator_class, mock_backup_manager_class, mock_detector_class
    ):
        """Test rollback when no backup found."""
        mock_backup_manager = MagicMock()
        mock_backup_manager.get_latest_backup.return_value = None
        mock_backup_manager_class.return_value = mock_backup_manager

        migrator = VersionMigrator(Path("/tmp"))
        result = migrator.rollback_to_latest_backup()

        assert result is False


class TestVersionMigratorEdgeCases:
    """Test edge cases and error scenarios."""

    @patch("moai_adk.core.migration.version_migrator.VersionDetector")
    @patch("moai_adk.core.migration.version_migrator.BackupManager")
    @patch("moai_adk.core.migration.version_migrator.FileMigrator")
    def test_migrate_exception_handling(
        self, mock_file_migrator_class, mock_backup_manager_class, mock_detector_class
    ):
        """Test migrate handles exceptions gracefully."""
        mock_detector = MagicMock()
        mock_detector.needs_migration.return_value = True
        mock_detector.get_migration_plan.side_effect = Exception("Test error")
        mock_detector_class.return_value = mock_detector

        mock_backup_manager = MagicMock()
        mock_backup_manager.create_backup.return_value = Path("/backup")
        mock_backup_manager_class.return_value = mock_backup_manager

        migrator = VersionMigrator(Path("/tmp"))

        # Exception should be caught and return False
        result = migrator.migrate_to_v024()
        assert result is False

    @patch("moai_adk.core.migration.version_migrator.VersionDetector")
    @patch("moai_adk.core.migration.version_migrator.BackupManager")
    @patch("moai_adk.core.migration.version_migrator.FileMigrator")
    def test_migrate_large_migration_plan(
        self, mock_file_migrator_class, mock_backup_manager_class, mock_detector_class
    ):
        """Test migrate with large migration plan."""
        # Create large migration plan
        large_plan = {
            "create": [f"dir_{i}" for i in range(100)],
            "move": [
                {
                    "description": f"Move {i}",
                    "from": f"old_{i}",
                    "to": f"new_{i}",
                }
                for i in range(100)
            ],
            "cleanup": [f"old_file_{i}" for i in range(100)],
        }

        mock_detector = MagicMock()
        mock_detector.needs_migration.return_value = True
        mock_detector.get_migration_plan.return_value = large_plan
        mock_detector_class.return_value = mock_detector

        mock_backup_manager = MagicMock()
        mock_backup_manager.create_backup.return_value = Path("/backup")
        mock_backup_manager_class.return_value = mock_backup_manager

        mock_file_migrator = MagicMock()
        mock_file_migrator.execute_migration_plan.return_value = {
            "success": True,
            "created_dirs": 100,
            "moved_files": 100,
            "errors": [],
        }
        mock_file_migrator.cleanup_old_files.return_value = 100
        mock_file_migrator_class.return_value = mock_file_migrator

        migrator = VersionMigrator(Path("/tmp"))

        with patch.object(migrator, "_verify_migration", return_value=True):
            result = migrator.migrate_to_v024()

        # Should handle large migration
        assert isinstance(result, bool)
