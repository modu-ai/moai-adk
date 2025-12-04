"""
Comprehensive tests for RollbackManager module.

Tests all major methods with mocked file system and git operations.
"""

import json
import tempfile
from datetime import datetime, timezone
from pathlib import Path
from unittest.mock import MagicMock, Mock, patch

import pytest

from moai_adk.core.rollback_manager import (
    RollbackManager,
    RollbackPoint,
    RollbackResult,
)


class TestRollbackPointDataclass:
    """Test RollbackPoint dataclass initialization and properties."""

    def test_rollback_point_creation(self):
        """Test creating a RollbackPoint with all fields."""
        ts = datetime.now(timezone.utc)
        point = RollbackPoint(
            id="rollback_20250101_120000_abc12345",
            timestamp=ts,
            description="Test rollback",
            changes=["file1.py", "file2.md"],
            backup_path="/backups/rollback_001",
            checksum="abc123def456",
            metadata={"version": "1.0.0"},
        )
        assert point.id == "rollback_20250101_120000_abc12345"
        assert point.description == "Test rollback"
        assert point.changes == ["file1.py", "file2.md"]
        assert point.checksum == "abc123def456"

    def test_rollback_point_empty_changes(self):
        """Test RollbackPoint with empty changes list."""
        ts = datetime.now(timezone.utc)
        point = RollbackPoint(
            id="test_id",
            timestamp=ts,
            description="Test",
            changes=[],
            backup_path="/backup",
            checksum="check",
            metadata={},
        )
        assert point.changes == []


class TestRollbackResultDataclass:
    """Test RollbackResult dataclass."""

    def test_rollback_result_success(self):
        """Test creating successful RollbackResult."""
        result = RollbackResult(
            success=True,
            rollback_point_id="test_id",
            message="Success",
            restored_files=["file1.py"],
            failed_files=[],
        )
        assert result.success is True
        assert result.message == "Success"

    def test_rollback_result_failure(self):
        """Test creating failed RollbackResult."""
        result = RollbackResult(
            success=False,
            rollback_point_id="test_id",
            message="Failed",
            restored_files=[],
            failed_files=["file1.py"],
        )
        assert result.success is False


class TestRollbackManagerInit:
    """Test RollbackManager initialization."""

    @patch("moai_adk.core.rollback_manager.Path.mkdir")
    @patch("moai_adk.core.rollback_manager.RollbackManager._load_registry")
    def test_init_creates_directories(self, mock_load, mock_mkdir):
        """Test that __init__ creates required directories."""
        mock_load.return_value = {}
        manager = RollbackManager(project_root=Path("/test"))
        assert manager.project_root == Path("/test")
        assert manager.backup_root == Path("/test/.moai/rollbacks")
        assert manager.config_backup_dir == Path("/test/.moai/rollbacks/config")

    @patch("moai_adk.core.rollback_manager.Path.mkdir")
    @patch("moai_adk.core.rollback_manager.RollbackManager._load_registry")
    def test_init_loads_registry(self, mock_load, mock_mkdir):
        """Test that __init__ loads existing registry."""
        mock_load.return_value = {"rollback_001": {"id": "rollback_001"}}
        manager = RollbackManager(project_root=Path("/test"))
        assert mock_load.called


class TestRollbackManagerCreateCheckpoint:
    """Test create_rollback_point method."""

    @patch("moai_adk.core.rollback_manager.Path.mkdir")
    @patch("moai_adk.core.rollback_manager.RollbackManager._load_registry")
    @patch("moai_adk.core.rollback_manager.RollbackManager._backup_configuration")
    @patch("moai_adk.core.rollback_manager.RollbackManager._backup_research_components")
    @patch("moai_adk.core.rollback_manager.RollbackManager._backup_code_files")
    @patch("moai_adk.core.rollback_manager.RollbackManager._calculate_backup_checksum")
    @patch("moai_adk.core.rollback_manager.RollbackManager._save_registry")
    def test_create_rollback_point_success(
        self,
        mock_save,
        mock_checksum,
        mock_code_backup,
        mock_research_backup,
        mock_config_backup,
        mock_load,
        mock_mkdir,
    ):
        """Test successful rollback point creation."""
        mock_load.return_value = {}
        mock_config_backup.return_value = "/backup/config"
        mock_research_backup.return_value = "/backup/research"
        mock_code_backup.return_value = "/backup/code"
        mock_checksum.return_value = "abc123"

        manager = RollbackManager(project_root=Path("/test"))
        rollback_id = manager.create_rollback_point("Test change", ["file1.py"])

        assert rollback_id.startswith("rollback_")
        assert rollback_id in manager.registry
        assert manager.registry[rollback_id]["description"] == "Test change"
        assert mock_save.called

    @patch("moai_adk.core.rollback_manager.Path.mkdir")
    @patch("moai_adk.core.rollback_manager.RollbackManager._load_registry")
    @patch("moai_adk.core.rollback_manager.RollbackManager._backup_configuration")
    @patch("moai_adk.core.rollback_manager.RollbackManager._cleanup_partial_backup")
    def test_create_rollback_point_failure_cleanup(self, mock_cleanup, mock_config_backup, mock_load, mock_mkdir):
        """Test rollback point creation failure triggers cleanup."""
        mock_load.return_value = {}
        mock_config_backup.side_effect = Exception("Backup failed")

        manager = RollbackManager(project_root=Path("/test"))

        with pytest.raises(Exception):
            manager.create_rollback_point("Test change")

        assert mock_cleanup.called


class TestRollbackManagerRollback:
    """Test rollback_to_point method."""

    @patch("moai_adk.core.rollback_manager.Path.mkdir")
    @patch("moai_adk.core.rollback_manager.RollbackManager._load_registry")
    @patch("moai_adk.core.rollback_manager.RollbackManager._validate_rollback_point")
    @patch("moai_adk.core.rollback_manager.RollbackManager._perform_rollback")
    @patch("moai_adk.core.rollback_manager.RollbackManager._validate_system_after_rollback")
    @patch("moai_adk.core.rollback_manager.RollbackManager._mark_rollback_as_used")
    def test_rollback_to_point_success(
        self,
        mock_mark,
        mock_validate_after,
        mock_perform,
        mock_validate_before,
        mock_load,
        mock_mkdir,
    ):
        """Test successful rollback to point."""
        ts = datetime.now(timezone.utc)
        registry = {
            "test_rollback": {
                "id": "test_rollback",
                "timestamp": ts.isoformat(),
                "description": "Test",
                "changes": [],
                "backup_path": "/backup",
                "checksum": "abc",
                "metadata": {},
            }
        }
        mock_load.return_value = registry
        mock_validate_before.return_value = {
            "valid": True,
            "message": "OK",
            "warnings": [],
        }
        mock_perform.return_value = (["file1.py"], [])
        mock_validate_after.return_value = {
            "config_valid": True,
            "research_valid": True,
            "issues": [],
        }

        manager = RollbackManager(project_root=Path("/test"))
        result = manager.rollback_to_point("test_rollback")

        assert result.success is True
        assert result.restored_files == ["file1.py"]
        assert mock_mark.called

    @patch("moai_adk.core.rollback_manager.Path.mkdir")
    @patch("moai_adk.core.rollback_manager.RollbackManager._load_registry")
    def test_rollback_to_point_not_found(self, mock_load, mock_mkdir):
        """Test rollback with non-existent rollback ID."""
        mock_load.return_value = {}
        manager = RollbackManager(project_root=Path("/test"))
        result = manager.rollback_to_point("nonexistent")

        assert result.success is False
        assert "not found" in result.message


class TestRollbackManagerListCheckpoints:
    """Test list_rollback_points method."""

    @patch("moai_adk.core.rollback_manager.Path.mkdir")
    @patch("moai_adk.core.rollback_manager.RollbackManager._load_registry")
    def test_list_rollback_points_empty(self, mock_load, mock_mkdir):
        """Test listing rollback points with empty registry."""
        mock_load.return_value = {}
        manager = RollbackManager(project_root=Path("/test"))
        points = manager.list_rollback_points()

        assert points == []

    @patch("moai_adk.core.rollback_manager.Path.mkdir")
    @patch("moai_adk.core.rollback_manager.RollbackManager._load_registry")
    def test_list_rollback_points_multiple(self, mock_load, mock_mkdir):
        """Test listing multiple rollback points."""
        ts1 = datetime(2025, 1, 1, 12, 0, 0, tzinfo=timezone.utc).isoformat()
        ts2 = datetime(2025, 1, 2, 12, 0, 0, tzinfo=timezone.utc).isoformat()

        registry = {
            "rollback_001": {
                "id": "rollback_001",
                "timestamp": ts1,
                "description": "First",
                "changes": ["file1.py"],
            },
            "rollback_002": {
                "id": "rollback_002",
                "timestamp": ts2,
                "description": "Second",
                "changes": ["file2.py"],
            },
        }
        mock_load.return_value = registry
        manager = RollbackManager(project_root=Path("/test"))
        points = manager.list_rollback_points(limit=10)

        assert len(points) == 2
        assert points[0]["id"] == "rollback_002"  # Newest first

    @patch("moai_adk.core.rollback_manager.Path.mkdir")
    @patch("moai_adk.core.rollback_manager.RollbackManager._load_registry")
    def test_list_rollback_points_limit(self, mock_load, mock_mkdir):
        """Test rollback points listing respects limit."""
        registry = {
            f"rollback_{i:03d}": {
                "id": f"rollback_{i:03d}",
                "timestamp": datetime.now(timezone.utc).isoformat(),
                "description": f"Rollback {i}",
                "changes": [],
            }
            for i in range(15)
        }
        mock_load.return_value = registry
        manager = RollbackManager(project_root=Path("/test"))
        points = manager.list_rollback_points(limit=5)

        assert len(points) == 5


class TestRollbackManagerValidateSystem:
    """Test validate_rollback_system method."""

    @patch("moai_adk.core.rollback_manager.Path.mkdir")
    @patch("moai_adk.core.rollback_manager.Path.exists")
    @patch("moai_adk.core.rollback_manager.RollbackManager._load_registry")
    @patch("moai_adk.core.rollback_manager.RollbackManager._calculate_backup_size")
    @patch("moai_adk.core.rollback_manager.shutil.disk_usage")
    def test_validate_system_healthy(self, mock_disk_usage, mock_backup_size, mock_load, mock_exists, mock_mkdir):
        """Test system validation when healthy."""
        mock_load.return_value = {}
        mock_exists.return_value = True
        mock_backup_size.return_value = 1000000
        mock_disk_usage.return_value = MagicMock(free=100000000)

        manager = RollbackManager(project_root=Path("/test"))
        result = manager.validate_rollback_system()

        assert result["system_healthy"] is True
        assert result["issues"] == []

    @patch("moai_adk.core.rollback_manager.Path.mkdir")
    @patch("moai_adk.core.rollback_manager.Path.exists")
    @patch("moai_adk.core.rollback_manager.RollbackManager._load_registry")
    @patch("moai_adk.core.rollback_manager.RollbackManager._calculate_backup_size")
    @patch("moai_adk.core.rollback_manager.shutil.disk_usage")
    def test_validate_system_missing_directories(
        self, mock_disk_usage, mock_backup_size, mock_load, mock_exists, mock_mkdir
    ):
        """Test validation detects missing directories."""
        mock_load.return_value = {}
        mock_exists.return_value = False  # Directories don't exist
        mock_backup_size.return_value = 0
        mock_disk_usage.return_value = MagicMock(free=100000000)

        manager = RollbackManager(project_root=Path("/test"))
        result = manager.validate_rollback_system()

        assert result["system_healthy"] is False
        assert len(result["issues"]) > 0


class TestRollbackManagerCleanup:
    """Test cleanup_old_rollbacks method."""

    @patch("moai_adk.core.rollback_manager.Path.mkdir")
    @patch("moai_adk.core.rollback_manager.RollbackManager._load_registry")
    @patch("moai_adk.core.rollback_manager.RollbackManager._get_directory_size")
    def test_cleanup_dry_run(self, mock_get_size, mock_load, mock_mkdir):
        """Test cleanup in dry-run mode."""
        ts = datetime.now(timezone.utc).isoformat()
        registry = {
            f"rollback_{i:03d}": {
                "id": f"rollback_{i:03d}",
                "timestamp": ts,
                "description": f"Rollback {i}",
                "changes": [],
                "backup_path": f"/backup/{i}",
            }
            for i in range(15)
        }
        mock_load.return_value = registry
        mock_get_size.return_value = 1000

        manager = RollbackManager(project_root=Path("/test"))
        result = manager.cleanup_old_rollbacks(keep_count=5, dry_run=True)

        assert result["dry_run"] is True
        assert result["would_delete_count"] == 10
        assert result["would_keep_count"] == 5

    @patch("moai_adk.core.rollback_manager.Path.mkdir")
    @patch("moai_adk.core.rollback_manager.RollbackManager._load_registry")
    @patch("moai_adk.core.rollback_manager.RollbackManager._get_directory_size")
    @patch("moai_adk.core.rollback_manager.RollbackManager._save_registry")
    @patch("moai_adk.core.rollback_manager.Path.exists")
    @patch("moai_adk.core.rollback_manager.shutil.rmtree")
    def test_cleanup_execute(self, mock_rmtree, mock_exists, mock_save, mock_get_size, mock_load, mock_mkdir):
        """Test cleanup in execute mode."""
        ts = datetime.now(timezone.utc).isoformat()
        registry = {
            f"rollback_{i:03d}": {
                "id": f"rollback_{i:03d}",
                "timestamp": ts,
                "description": f"Rollback {i}",
                "changes": [],
                "backup_path": f"/backup/{i}",
            }
            for i in range(5)
        }
        mock_load.return_value = registry
        mock_get_size.return_value = 1000
        mock_exists.return_value = True

        manager = RollbackManager(project_root=Path("/test"))
        result = manager.cleanup_old_rollbacks(keep_count=2, dry_run=False)

        assert result["dry_run"] is False
        assert result["deleted_count"] <= 3
        assert mock_save.called


class TestRollbackManagerBackupMethods:
    """Test private backup methods."""

    @patch("moai_adk.core.rollback_manager.Path.mkdir")
    @patch("moai_adk.core.rollback_manager.RollbackManager._load_registry")
    @patch("moai_adk.core.rollback_manager.Path.exists")
    @patch("moai_adk.core.rollback_manager.shutil.copy2")
    def test_backup_configuration(self, mock_copy, mock_exists, mock_load, mock_mkdir):
        """Test _backup_configuration method."""
        mock_load.return_value = {}
        mock_exists.side_effect = [
            True,
            True,
            False,
        ]  # config.json, settings.json, settings.local.json

        manager = RollbackManager(project_root=Path("/test"))
        result = manager._backup_configuration(Path("/backup"))

        assert result == str(Path("/backup/config"))

    @patch("moai_adk.core.rollback_manager.Path.mkdir")
    @patch("moai_adk.core.rollback_manager.RollbackManager._load_registry")
    @patch("moai_adk.core.rollback_manager.Path.exists")
    @patch("moai_adk.core.rollback_manager.shutil.copytree")
    def test_backup_research_components(self, mock_copytree, mock_exists, mock_load, mock_mkdir):
        """Test _backup_research_components method."""
        mock_load.return_value = {}
        mock_exists.return_value = True

        manager = RollbackManager(project_root=Path("/test"))
        result = manager._backup_research_components(Path("/backup"))

        assert result == str(Path("/backup/research"))


class TestRollbackManagerIdGeneration:
    """Test ID generation and helper methods."""

    @patch("moai_adk.core.rollback_manager.Path.mkdir")
    @patch("moai_adk.core.rollback_manager.RollbackManager._load_registry")
    def test_generate_rollback_id_format(self, mock_load, mock_mkdir):
        """Test generated rollback IDs have correct format."""
        mock_load.return_value = {}
        manager = RollbackManager(project_root=Path("/test"))

        id1 = manager._generate_rollback_id()
        id2 = manager._generate_rollback_id()

        assert id1.startswith("rollback_")
        assert id2.startswith("rollback_")
        assert id1 != id2  # IDs should be unique

    @patch("moai_adk.core.rollback_manager.Path.mkdir")
    @patch("moai_adk.core.rollback_manager.RollbackManager._load_registry")
    def test_calculate_backup_checksum(self, mock_load, mock_mkdir):
        """Test checksum calculation."""
        mock_load.return_value = {}

        with tempfile.TemporaryDirectory() as tmpdir:
            backup_dir = Path(tmpdir)
            test_file = backup_dir / "test.txt"
            test_file.write_text("test content")

            manager = RollbackManager(project_root=Path("/test"))
            checksum = manager._calculate_backup_checksum(backup_dir)

            assert isinstance(checksum, str)
            assert len(checksum) == 64  # SHA256 hex length


class TestRollbackManagerResearchRollback:
    """Test research-specific rollback operations."""

    @patch("moai_adk.core.rollback_manager.Path.mkdir")
    @patch("moai_adk.core.rollback_manager.RollbackManager._load_registry")
    @patch("moai_adk.core.rollback_manager.RollbackManager._find_research_rollback_points")
    def test_rollback_research_integration_no_points(self, mock_find, mock_load, mock_mkdir):
        """Test research rollback when no suitable points found."""
        mock_load.return_value = {}
        mock_find.return_value = []

        manager = RollbackManager(project_root=Path("/test"))
        result = manager.rollback_research_integration("skills")

        assert result.success is False
        assert "No suitable rollback points" in result.message


class TestRegistryLoadSave:
    """Test registry persistence."""

    @patch("moai_adk.core.rollback_manager.Path.mkdir")
    def test_load_registry_creates_empty_on_missing_file(self, mock_mkdir):
        """Test _load_registry returns empty dict when file missing."""
        with tempfile.TemporaryDirectory() as tmpdir:
            registry_file = Path(tmpdir) / "registry.json"
            manager = RollbackManager(project_root=Path(tmpdir))
            manager.registry_file = registry_file

            result = manager._load_registry()
            assert result == {}

    @patch("moai_adk.core.rollback_manager.Path.mkdir")
    def test_save_and_load_registry(self, mock_mkdir):
        """Test saving and loading registry."""
        with tempfile.TemporaryDirectory() as tmpdir:
            registry_file = Path(tmpdir) / "registry.json"
            registry_file.parent.mkdir(parents=True, exist_ok=True)

            manager = RollbackManager(project_root=Path(tmpdir))
            manager.registry_file = registry_file
            manager.registry = {
                "test_id": {
                    "id": "test_id",
                    "timestamp": datetime.now(timezone.utc).isoformat(),
                    "description": "Test",
                    "changes": [],
                    "backup_path": "/backup",
                    "checksum": "abc",
                    "metadata": {},
                }
            }

            manager._save_registry()
            assert registry_file.exists()

            loaded = manager._load_registry()
            assert "test_id" in loaded
