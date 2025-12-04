"""Simple tests for rollback_manager module.

Tests basic rollback point creation, listing, and cleanup operations
with mocked file operations.
"""

import json
import unittest
from datetime import datetime, timezone
from pathlib import Path
from unittest.mock import MagicMock, Mock, patch

from src.moai_adk.core.rollback_manager import (
    RollbackManager,
    RollbackPoint,
    RollbackResult,
)


class TestRollbackPoint(unittest.TestCase):
    """Test RollbackPoint dataclass."""

    def test_rollback_point_creation(self):
        """Test creating a RollbackPoint instance."""
        # Arrange
        now = datetime.now(timezone.utc)
        rollback_point = RollbackPoint(
            id="rollback_001",
            timestamp=now,
            description="Test rollback",
            changes=["file1.py", "file2.py"],
            backup_path="/backup/rollback_001",
            checksum="abc123def456",
            metadata={"version": "1.0.0"},
        )

        # Act
        # Assert
        self.assertEqual(rollback_point.id, "rollback_001")
        self.assertEqual(rollback_point.description, "Test rollback")
        self.assertEqual(len(rollback_point.changes), 2)
        self.assertEqual(rollback_point.checksum, "abc123def456")


class TestRollbackResult(unittest.TestCase):
    """Test RollbackResult dataclass."""

    def test_rollback_result_success(self):
        """Test creating a successful RollbackResult."""
        # Arrange
        result = RollbackResult(
            success=True,
            rollback_point_id="rollback_001",
            message="Rollback completed",
            restored_files=["file1.py", "file2.py"],
            failed_files=[],
        )

        # Act
        # Assert
        self.assertTrue(result.success)
        self.assertEqual(result.message, "Rollback completed")
        self.assertEqual(len(result.restored_files), 2)
        self.assertEqual(len(result.failed_files), 0)

    def test_rollback_result_failure(self):
        """Test creating a failed RollbackResult."""
        # Arrange
        result = RollbackResult(
            success=False,
            rollback_point_id="rollback_001",
            message="Rollback failed",
            restored_files=[],
            failed_files=["file1.py"],
        )

        # Act
        # Assert
        self.assertFalse(result.success)
        self.assertEqual(len(result.failed_files), 1)


class TestRollbackManagerInitialization(unittest.TestCase):
    """Test RollbackManager initialization."""

    @patch("src.moai_adk.core.rollback_manager.Path.mkdir")
    @patch("src.moai_adk.core.rollback_manager.Path.exists")
    def test_rollback_manager_init(self, mock_exists, mock_mkdir):
        """Test RollbackManager initialization."""
        # Arrange
        mock_exists.return_value = False

        # Act
        with patch("src.moai_adk.core.rollback_manager.Path.cwd") as mock_cwd:
            mock_cwd.return_value = Path("/test/project")
            manager = RollbackManager(Path("/test/project"))

        # Assert
        self.assertIsNotNone(manager)
        self.assertEqual(str(manager.project_root), "/test/project")

    @patch("src.moai_adk.core.rollback_manager.Path.mkdir")
    @patch("src.moai_adk.core.rollback_manager.Path.exists")
    def test_rollback_manager_directories(self, mock_exists, mock_mkdir):
        """Test that rollback manager creates required directories."""
        # Arrange
        mock_exists.return_value = False

        # Act
        with patch("src.moai_adk.core.rollback_manager.Path.cwd") as mock_cwd:
            mock_cwd.return_value = Path("/test/project")
            manager = RollbackManager(Path("/test/project"))

        # Assert
        self.assertIsNotNone(manager.backup_root)
        self.assertIsNotNone(manager.config_backup_dir)
        self.assertIsNotNone(manager.code_backup_dir)


class TestRollbackPointCreation(unittest.TestCase):
    """Test rollback point creation."""

    @patch("src.moai_adk.core.rollback_manager.RollbackManager._load_registry")
    @patch("src.moai_adk.core.rollback_manager.RollbackManager._backup_code_files")
    @patch("src.moai_adk.core.rollback_manager.RollbackManager._backup_research_components")
    @patch("src.moai_adk.core.rollback_manager.RollbackManager._backup_configuration")
    @patch("src.moai_adk.core.rollback_manager.RollbackManager._calculate_backup_checksum")
    @patch("src.moai_adk.core.rollback_manager.RollbackManager._save_registry")
    @patch("src.moai_adk.core.rollback_manager.Path.mkdir")
    @patch("src.moai_adk.core.rollback_manager.Path.exists")
    def test_create_rollback_point(
        self,
        mock_exists,
        mock_mkdir,
        mock_save_registry,
        mock_calculate_checksum,
        mock_backup_config,
        mock_backup_research,
        mock_backup_code,
        mock_load_registry,
    ):
        """Test creating a rollback point."""
        # Arrange
        mock_exists.return_value = False
        mock_load_registry.return_value = {}
        mock_calculate_checksum.return_value = "test_checksum"
        mock_backup_config.return_value = "/backup/config"
        mock_backup_research.return_value = "/backup/research"
        mock_backup_code.return_value = "/backup/code"

        with patch("src.moai_adk.core.rollback_manager.Path.cwd") as mock_cwd:
            mock_cwd.return_value = Path("/test/project")
            manager = RollbackManager(Path("/test/project"))

            # Act
            rollback_id = manager.create_rollback_point("Test backup", ["file1.py"])

        # Assert
        self.assertIsNotNone(rollback_id)
        self.assertIn("rollback_", rollback_id)
        self.assertIn(rollback_id, manager.registry)

    @patch("src.moai_adk.core.rollback_manager.RollbackManager._load_registry")
    @patch("src.moai_adk.core.rollback_manager.RollbackManager._backup_code_files")
    @patch("src.moai_adk.core.rollback_manager.RollbackManager._backup_research_components")
    @patch("src.moai_adk.core.rollback_manager.RollbackManager._backup_configuration")
    @patch("src.moai_adk.core.rollback_manager.RollbackManager._calculate_backup_checksum")
    @patch("src.moai_adk.core.rollback_manager.RollbackManager._save_registry")
    @patch("src.moai_adk.core.rollback_manager.Path.mkdir")
    @patch("src.moai_adk.core.rollback_manager.Path.exists")
    def test_create_rollback_point_with_changes(
        self,
        mock_exists,
        mock_mkdir,
        mock_save_registry,
        mock_calculate_checksum,
        mock_backup_config,
        mock_backup_research,
        mock_backup_code,
        mock_load_registry,
    ):
        """Test creating a rollback point with specific changes."""
        # Arrange
        mock_exists.return_value = False
        mock_load_registry.return_value = {}
        mock_calculate_checksum.return_value = "checksum123"
        mock_backup_config.return_value = "/backup/config"
        mock_backup_research.return_value = "/backup/research"
        mock_backup_code.return_value = "/backup/code"
        changes = ["file1.py", "file2.py", "config.json"]

        with patch("src.moai_adk.core.rollback_manager.Path.cwd") as mock_cwd:
            mock_cwd.return_value = Path("/test/project")
            manager = RollbackManager(Path("/test/project"))

            # Act
            rollback_id = manager.create_rollback_point("Multi-file backup", changes)

        # Assert
        self.assertIsNotNone(rollback_id)
        registry_entry = manager.registry[rollback_id]
        self.assertEqual(len(registry_entry["changes"]), 3)


class TestListRollbackPoints(unittest.TestCase):
    """Test listing rollback points."""

    @patch("src.moai_adk.core.rollback_manager.RollbackManager._load_registry")
    @patch("src.moai_adk.core.rollback_manager.Path.mkdir")
    @patch("src.moai_adk.core.rollback_manager.Path.exists")
    def test_list_rollback_points_empty(
        self, mock_exists, mock_mkdir, mock_load_registry
    ):
        """Test listing rollback points when none exist."""
        # Arrange
        mock_exists.return_value = False
        mock_load_registry.return_value = {}

        with patch("src.moai_adk.core.rollback_manager.Path.cwd") as mock_cwd:
            mock_cwd.return_value = Path("/test/project")
            manager = RollbackManager(Path("/test/project"))

            # Act
            rollback_points = manager.list_rollback_points()

        # Assert
        self.assertEqual(len(rollback_points), 0)

    @patch("src.moai_adk.core.rollback_manager.RollbackManager._load_registry")
    @patch("src.moai_adk.core.rollback_manager.Path.mkdir")
    @patch("src.moai_adk.core.rollback_manager.Path.exists")
    def test_list_rollback_points_with_data(
        self, mock_exists, mock_mkdir, mock_load_registry
    ):
        """Test listing rollback points with existing data."""
        # Arrange
        now = datetime.now(timezone.utc).isoformat()
        registry_data = {
            "rollback_001": {
                "id": "rollback_001",
                "timestamp": now,
                "description": "First backup",
                "changes": ["file1.py"],
                "backup_path": "/backup/rollback_001",
                "checksum": "abc123",
                "metadata": {},
            },
            "rollback_002": {
                "id": "rollback_002",
                "timestamp": now,
                "description": "Second backup",
                "changes": ["file2.py"],
                "backup_path": "/backup/rollback_002",
                "checksum": "def456",
                "metadata": {},
            },
        }
        mock_exists.return_value = False
        mock_load_registry.return_value = registry_data

        with patch("src.moai_adk.core.rollback_manager.Path.cwd") as mock_cwd:
            mock_cwd.return_value = Path("/test/project")
            manager = RollbackManager(Path("/test/project"))

            # Act
            rollback_points = manager.list_rollback_points(limit=10)

        # Assert
        self.assertEqual(len(rollback_points), 2)
        self.assertEqual(rollback_points[0]["id"], "rollback_001")

    @patch("src.moai_adk.core.rollback_manager.RollbackManager._load_registry")
    @patch("src.moai_adk.core.rollback_manager.Path.mkdir")
    @patch("src.moai_adk.core.rollback_manager.Path.exists")
    def test_list_rollback_points_limit(
        self, mock_exists, mock_mkdir, mock_load_registry
    ):
        """Test listing rollback points respects limit."""
        # Arrange
        now = datetime.now(timezone.utc).isoformat()
        registry_data = {
            f"rollback_{i:03d}": {
                "id": f"rollback_{i:03d}",
                "timestamp": now,
                "description": f"Backup {i}",
                "changes": [],
                "backup_path": f"/backup/rollback_{i:03d}",
                "checksum": f"checksum{i}",
                "metadata": {},
            }
            for i in range(20)
        }
        mock_exists.return_value = False
        mock_load_registry.return_value = registry_data

        with patch("src.moai_adk.core.rollback_manager.Path.cwd") as mock_cwd:
            mock_cwd.return_value = Path("/test/project")
            manager = RollbackManager(Path("/test/project"))

            # Act
            rollback_points = manager.list_rollback_points(limit=5)

        # Assert
        self.assertEqual(len(rollback_points), 5)


class TestValidateRollbackSystem(unittest.TestCase):
    """Test rollback system validation."""

    @patch("src.moai_adk.core.rollback_manager.RollbackManager._load_registry")
    @patch("src.moai_adk.core.rollback_manager.Path.mkdir")
    @patch("src.moai_adk.core.rollback_manager.Path.exists")
    def test_validate_rollback_system_empty(
        self, mock_exists, mock_mkdir, mock_load_registry
    ):
        """Test validating rollback system with no data."""
        # Arrange
        mock_exists.return_value = False
        mock_load_registry.return_value = {}

        with patch("src.moai_adk.core.rollback_manager.Path.cwd") as mock_cwd:
            mock_cwd.return_value = Path("/test/project")
            manager = RollbackManager(Path("/test/project"))

            # Act
            validation = manager.validate_rollback_system()

        # Assert
        self.assertIn("system_healthy", validation)
        self.assertIn("rollback_points_count", validation)
        self.assertEqual(validation["rollback_points_count"], 0)

    @patch("src.moai_adk.core.rollback_manager.RollbackManager._load_registry")
    @patch("src.moai_adk.core.rollback_manager.RollbackManager._calculate_backup_size")
    @patch("src.moai_adk.core.rollback_manager.shutil.disk_usage")
    @patch("src.moai_adk.core.rollback_manager.Path.mkdir")
    @patch("src.moai_adk.core.rollback_manager.Path.exists")
    def test_validate_rollback_system_with_backups(
        self,
        mock_exists,
        mock_mkdir,
        mock_disk_usage,
        mock_calc_size,
        mock_load_registry,
    ):
        """Test validating rollback system with existing backups."""
        # Arrange
        now = datetime.now(timezone.utc).isoformat()
        registry_data = {
            "rollback_001": {
                "id": "rollback_001",
                "timestamp": now,
                "description": "Test",
                "changes": [],
                "backup_path": "/backup/rollback_001",
                "checksum": "abc123",
                "metadata": {},
            }
        }
        mock_exists.return_value = False
        mock_load_registry.return_value = registry_data
        mock_calc_size.return_value = 1024 * 1024  # 1 MB
        mock_disk_usage.return_value = Mock(free=1024 * 1024 * 1024)  # 1 GB free

        with patch("src.moai_adk.core.rollback_manager.Path.cwd") as mock_cwd:
            mock_cwd.return_value = Path("/test/project")
            manager = RollbackManager(Path("/test/project"))

            # Act
            validation = manager.validate_rollback_system()

        # Assert
        self.assertEqual(validation["rollback_points_count"], 1)
        self.assertIn("backup_size", validation)


class TestCleanupOldRollbacks(unittest.TestCase):
    """Test cleanup of old rollback points."""

    @patch("src.moai_adk.core.rollback_manager.RollbackManager._load_registry")
    @patch("src.moai_adk.core.rollback_manager.Path.mkdir")
    @patch("src.moai_adk.core.rollback_manager.Path.exists")
    def test_cleanup_dry_run(
        self, mock_exists, mock_mkdir, mock_load_registry
    ):
        """Test cleanup in dry run mode."""
        # Arrange
        now = datetime.now(timezone.utc).isoformat()
        registry_data = {
            f"rollback_{i:03d}": {
                "id": f"rollback_{i:03d}",
                "timestamp": now,
                "description": f"Backup {i}",
                "changes": [],
                "backup_path": f"/backup/rollback_{i:03d}",
                "checksum": f"checksum{i}",
                "metadata": {},
            }
            for i in range(15)
        }
        mock_exists.return_value = False
        mock_load_registry.return_value = registry_data

        with patch("src.moai_adk.core.rollback_manager.Path.cwd") as mock_cwd:
            mock_cwd.return_value = Path("/test/project")
            manager = RollbackManager(Path("/test/project"))

            # Act
            result = manager.cleanup_old_rollbacks(keep_count=10, dry_run=True)

        # Assert
        self.assertTrue(result["dry_run"])
        self.assertEqual(result["would_delete_count"], 5)
        self.assertEqual(result["would_keep_count"], 10)

    @patch("src.moai_adk.core.rollback_manager.RollbackManager._load_registry")
    @patch("src.moai_adk.core.rollback_manager.RollbackManager._save_registry")
    @patch("src.moai_adk.core.rollback_manager.RollbackManager._get_directory_size")
    @patch("src.moai_adk.core.rollback_manager.shutil.rmtree")
    @patch("src.moai_adk.core.rollback_manager.Path.mkdir")
    @patch("src.moai_adk.core.rollback_manager.Path.exists")
    def test_cleanup_execute(
        self,
        mock_exists,
        mock_mkdir,
        mock_rmtree,
        mock_get_size,
        mock_save_registry,
        mock_load_registry,
    ):
        """Test cleanup execution."""
        # Arrange
        now = datetime.now(timezone.utc).isoformat()
        registry_data = {
            f"rollback_{i:03d}": {
                "id": f"rollback_{i:03d}",
                "timestamp": now,
                "description": f"Backup {i}",
                "changes": [],
                "backup_path": f"/backup/rollback_{i:03d}",
                "checksum": f"checksum{i}",
                "metadata": {},
            }
            for i in range(15)
        }
        mock_exists.return_value = False
        mock_load_registry.return_value = registry_data
        mock_get_size.return_value = 1024 * 1024  # 1 MB per directory

        with patch("src.moai_adk.core.rollback_manager.Path.cwd") as mock_cwd:
            mock_cwd.return_value = Path("/test/project")
            manager = RollbackManager(Path("/test/project"))

            # Act
            result = manager.cleanup_old_rollbacks(keep_count=10, dry_run=False)

        # Assert
        self.assertFalse(result["dry_run"])
        self.assertEqual(result["deleted_count"], 5)
        self.assertEqual(result["kept_count"], 10)


class TestRollbackHelper(unittest.TestCase):
    """Test rollback manager helper methods."""

    @patch("src.moai_adk.core.rollback_manager.RollbackManager._load_registry")
    @patch("src.moai_adk.core.rollback_manager.Path.mkdir")
    @patch("src.moai_adk.core.rollback_manager.Path.exists")
    def test_generate_rollback_id(
        self, mock_exists, mock_mkdir, mock_load_registry
    ):
        """Test rollback ID generation."""
        # Arrange
        mock_exists.return_value = False
        mock_load_registry.return_value = {}

        with patch("src.moai_adk.core.rollback_manager.Path.cwd") as mock_cwd:
            mock_cwd.return_value = Path("/test/project")
            manager = RollbackManager(Path("/test/project"))

            # Act
            rollback_id = manager._generate_rollback_id()

        # Assert
        self.assertIn("rollback_", rollback_id)
        self.assertRegex(rollback_id, r"rollback_\d{8}_\d{6}_[a-f0-9]{8}")

    @patch("src.moai_adk.core.rollback_manager.RollbackManager._load_registry")
    @patch("src.moai_adk.core.rollback_manager.Path.mkdir")
    @patch("src.moai_adk.core.rollback_manager.Path.exists")
    def test_calculate_backup_checksum(
        self, mock_exists, mock_mkdir, mock_load_registry
    ):
        """Test backup checksum calculation."""
        # Arrange
        mock_exists.return_value = False
        mock_load_registry.return_value = {}

        with patch("src.moai_adk.core.rollback_manager.Path.cwd") as mock_cwd:
            mock_cwd.return_value = Path("/test/project")
            manager = RollbackManager(Path("/test/project"))

            # Mock the backup directory
            with patch("src.moai_adk.core.rollback_manager.Path.rglob") as mock_rglob:
                mock_file = MagicMock()
                mock_file.is_file.return_value = True
                mock_file.read_bytes.return_value = b"test content"
                mock_file.relative_to.return_value = Path("test.txt")
                mock_rglob.return_value = [mock_file]

                # Act
                checksum = manager._calculate_backup_checksum(Path("/test/backup"))

        # Assert
        self.assertIsNotNone(checksum)
        self.assertEqual(len(checksum), 64)  # SHA256 hex digest


if __name__ == "__main__":
    unittest.main()
