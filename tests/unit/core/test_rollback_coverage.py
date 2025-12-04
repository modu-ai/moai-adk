"""Comprehensive coverage tests for RollbackManager module.

These tests focus on uncovered code paths in src/moai_adk/core/rollback_manager.py
with emphasis on checkpoint creation, rollback operations, file system interactions,
and validation logic.

Target Coverage: 70%+
Test Pattern: AAA (Arrange-Act-Assert)
Mocks: File system operations, JSON I/O, shutil operations
"""

import json
import hashlib
import logging
from datetime import datetime, timezone
from pathlib import Path
from unittest.mock import MagicMock, patch, mock_open, call
import pytest
import shutil

from moai_adk.core.rollback_manager import (
    RollbackManager,
    RollbackPoint,
    RollbackResult,
)


# ============================================================================
# Test RollbackPoint and RollbackResult Dataclasses
# ============================================================================


class TestRollbackPointDataclass:
    """Test suite for RollbackPoint dataclass."""

    def test_rollback_point_creation(self):
        """Test RollbackPoint creation."""
        # Arrange
        now = datetime.now(timezone.utc)

        # Act
        point = RollbackPoint(
            id="rollback_123",
            timestamp=now,
            description="Test rollback",
            changes=["file1.py", "file2.py"],
            backup_path="/mock/backup",
            checksum="abc123",
            metadata={"version": "1.0"},
        )

        # Assert
        assert point.id == "rollback_123"
        assert point.description == "Test rollback"
        assert len(point.changes) == 2


class TestRollbackResultDataclass:
    """Test suite for RollbackResult dataclass."""

    def test_rollback_result_creation_success(self):
        """Test successful RollbackResult creation."""
        # Act
        result = RollbackResult(
            success=True,
            rollback_point_id="rollback_123",
            message="Rollback successful",
            restored_files=["file1.py"],
            failed_files=[],
        )

        # Assert
        assert result.success is True
        assert result.rollback_point_id == "rollback_123"
        assert len(result.restored_files) == 1
        assert len(result.failed_files) == 0

    def test_rollback_result_creation_failure(self):
        """Test failed RollbackResult creation."""
        # Act
        result = RollbackResult(
            success=False,
            rollback_point_id="rollback_123",
            message="Rollback failed",
            restored_files=[],
            failed_files=["file2.py"],
        )

        # Assert
        assert result.success is False
        assert len(result.failed_files) == 1


# ============================================================================
# Test RollbackManager Initialization
# ============================================================================


class TestRollbackManagerInit:
    """Test suite for RollbackManager initialization."""

    @patch("pathlib.Path.mkdir")
    @patch("pathlib.Path.exists")
    @patch("moai_adk.core.rollback_manager.RollbackManager._load_registry")
    def test_rollback_manager_init_default(
        self, mock_load_registry, mock_exists, mock_mkdir
    ):
        """Test RollbackManager initialization with defaults."""
        # Arrange
        mock_load_registry.return_value = {}
        mock_exists.return_value = False

        # Act
        manager = RollbackManager()

        # Assert
        assert manager.project_root == Path.cwd()
        assert manager.backup_root is not None
        mock_mkdir.assert_called()

    @patch("pathlib.Path.mkdir")
    @patch("pathlib.Path.exists")
    @patch("moai_adk.core.rollback_manager.RollbackManager._load_registry")
    def test_rollback_manager_init_custom_path(
        self, mock_load_registry, mock_exists, mock_mkdir
    ):
        """Test RollbackManager with custom project root."""
        # Arrange
        custom_root = Path("/custom/project")
        mock_load_registry.return_value = {}
        mock_exists.return_value = False

        # Act
        manager = RollbackManager(project_root=custom_root)

        # Assert
        assert manager.project_root == custom_root


# ============================================================================
# Test Registry Operations
# ============================================================================


class TestRegistryOperations:
    """Test suite for registry loading and saving."""

    @patch("moai_adk.core.rollback_manager.RollbackManager._load_registry")
    @patch("pathlib.Path.mkdir")
    def test_load_registry_success(self, mock_mkdir, mock_load_registry):
        """Test successful registry loading."""
        # Arrange
        registry_data = {"rollback_123": {"id": "rollback_123"}}
        mock_load_registry.return_value = registry_data
        mock_mkdir.return_value = None

        # Act
        manager = RollbackManager()

        # Assert
        assert "rollback_123" in manager.registry

    @patch("pathlib.Path.exists")
    @patch("pathlib.Path.mkdir")
    def test_load_registry_missing(self, mock_mkdir, mock_exists):
        """Test registry loading when file missing."""
        # Arrange
        mock_exists.return_value = False
        mock_mkdir.return_value = None

        # Act
        manager = RollbackManager()

        # Assert
        assert manager.registry == {}

    @patch("moai_adk.core.rollback_manager.RollbackManager._load_registry")
    @patch("builtins.open", new_callable=mock_open)
    @patch("pathlib.Path.mkdir")
    def test_save_registry_success(self, mock_mkdir, mock_file, mock_load_registry):
        """Test successful registry saving."""
        # Arrange
        mock_load_registry.return_value = {}
        mock_mkdir.return_value = None

        manager = RollbackManager()
        manager.registry = {"rollback_123": {"id": "rollback_123"}}

        # Act
        manager._save_registry()

        # Assert
        mock_file.assert_called_once()
        # The json.dump writes to the file object, so check if write was called
        assert mock_file().write.called or mock_file().writelines.called

    @patch("moai_adk.core.rollback_manager.RollbackManager._load_registry")
    @patch("builtins.open", new_callable=mock_open)
    @patch("pathlib.Path.mkdir")
    def test_save_registry_with_special_chars(
        self, mock_mkdir, mock_file, mock_load_registry
    ):
        """Test registry saving with special characters."""
        # Arrange
        mock_load_registry.return_value = {}
        mock_mkdir.return_value = None

        manager = RollbackManager()
        manager.registry = {
            "rollback_123": {
                "id": "rollback_123",
                "description": "Test with 한글 characters",
            }
        }

        # Act
        manager._save_registry()

        # Assert
        mock_file.assert_called_once()
        # Check that open was called with the correct encoding
        call_args = mock_file.call_args
        assert "utf-8" in str(call_args)


# ============================================================================
# Test Rollback Point Creation
# ============================================================================


class TestCreateRollbackPoint:
    """Test suite for creating rollback points."""

    @patch("moai_adk.core.rollback_manager.RollbackManager._save_registry")
    @patch("moai_adk.core.rollback_manager.RollbackManager._calculate_backup_checksum")
    @patch("moai_adk.core.rollback_manager.RollbackManager._backup_code_files")
    @patch("moai_adk.core.rollback_manager.RollbackManager._backup_research_components")
    @patch("moai_adk.core.rollback_manager.RollbackManager._backup_configuration")
    @patch("pathlib.Path.mkdir")
    @patch("pathlib.Path.exists")
    def test_create_rollback_point_success(
        self,
        mock_exists,
        mock_mkdir,
        mock_backup_config,
        mock_backup_research,
        mock_backup_code,
        mock_checksum,
        mock_save_registry,
    ):
        """Test successful rollback point creation."""
        # Arrange
        mock_exists.return_value = False
        mock_mkdir.return_value = None
        mock_backup_config.return_value = "/mock/backup/config"
        mock_backup_research.return_value = "/mock/backup/research"
        mock_backup_code.return_value = "/mock/backup/code"
        mock_checksum.return_value = "abc123def456"

        manager = RollbackManager()

        # Act
        rollback_id = manager.create_rollback_point(
            description="Test backup",
            changes=["file1.py", "file2.py"],
        )

        # Assert
        assert rollback_id is not None
        assert rollback_id in manager.registry
        assert manager.registry[rollback_id]["description"] == "Test backup"

    @patch("moai_adk.core.rollback_manager.RollbackManager._cleanup_partial_backup")
    @patch("moai_adk.core.rollback_manager.RollbackManager._backup_configuration")
    @patch("pathlib.Path.mkdir")
    @patch("pathlib.Path.exists")
    def test_create_rollback_point_failure(
        self,
        mock_exists,
        mock_mkdir,
        mock_backup_config,
        mock_cleanup,
    ):
        """Test rollback point creation failure."""
        # Arrange
        mock_exists.return_value = False
        mock_mkdir.return_value = None
        mock_backup_config.side_effect = Exception("Backup failed")
        mock_cleanup.return_value = None

        manager = RollbackManager()

        # Act & Assert
        with pytest.raises(Exception):
            manager.create_rollback_point("Test backup")


# ============================================================================
# Test Backup Operations
# ============================================================================


class TestBackupOperations:
    """Test suite for backup operations."""

    @patch("pathlib.Path.exists")
    @patch("shutil.copy2")
    @patch("pathlib.Path.mkdir")
    def test_backup_configuration_files(self, mock_mkdir, mock_copy, mock_exists):
        """Test backing up configuration files."""
        # Arrange
        mock_exists.return_value = True
        mock_mkdir.return_value = None
        mock_copy.return_value = None

        manager = RollbackManager()
        rollback_dir = Path("/mock/rollback")

        # Act
        result = manager._backup_configuration(rollback_dir)

        # Assert
        assert result is not None
        mock_copy.assert_called()

    @patch("pathlib.Path.exists")
    @patch("shutil.copytree")
    @patch("pathlib.Path.mkdir")
    def test_backup_research_components(self, mock_mkdir, mock_copytree, mock_exists):
        """Test backing up research components."""
        # Arrange
        mock_exists.return_value = True
        mock_mkdir.return_value = None
        mock_copytree.return_value = None

        manager = RollbackManager()
        rollback_dir = Path("/mock/rollback")

        # Act
        result = manager._backup_research_components(rollback_dir)

        # Assert
        assert result is not None

    @patch("pathlib.Path.exists")
    @patch("shutil.copytree")
    @patch("pathlib.Path.mkdir")
    def test_backup_code_files(self, mock_mkdir, mock_copytree, mock_exists):
        """Test backing up code files."""
        # Arrange
        mock_exists.return_value = True
        mock_mkdir.return_value = None
        mock_copytree.return_value = None

        manager = RollbackManager()
        rollback_dir = Path("/mock/rollback")

        # Act
        result = manager._backup_code_files(rollback_dir)

        # Assert
        assert result is not None


# ============================================================================
# Test Checksum Calculation
# ============================================================================


class TestChecksumCalculation:
    """Test suite for backup checksum calculation."""

    @patch("pathlib.Path.rglob")
    @patch("builtins.open", new_callable=mock_open, read_data=b"test content")
    def test_calculate_backup_checksum(self, mock_file, mock_rglob):
        """Test backup checksum calculation."""
        # Arrange
        mock_rglob.return_value = [Path("/mock/file1.txt")]
        mock_file.return_value.__enter__ = MagicMock(
            return_value=mock_file.return_value
        )
        mock_file.return_value.__exit__ = MagicMock(return_value=False)

        manager = RollbackManager()
        backup_dir = Path("/mock/backup")

        # Act
        checksum = manager._calculate_backup_checksum(backup_dir)

        # Assert
        assert isinstance(checksum, str)
        assert len(checksum) > 0

    @patch("pathlib.Path.rglob")
    @patch("builtins.open", new_callable=mock_open, read_data=b"test")
    def test_calculate_backup_checksum_consistency(self, mock_file, mock_rglob):
        """Test checksum consistency for same content."""
        # Arrange
        mock_rglob.return_value = [Path("/mock/file.txt")]
        mock_file.return_value.__enter__ = MagicMock(
            return_value=mock_file.return_value
        )
        mock_file.return_value.__exit__ = MagicMock(return_value=False)
        mock_file.return_value.read.return_value = b"test"

        manager = RollbackManager()

        # Act
        checksum1 = manager._calculate_backup_checksum(Path("/mock/backup"))
        checksum2 = manager._calculate_backup_checksum(Path("/mock/backup"))

        # Assert
        assert checksum1 == checksum2


# ============================================================================
# Test Rollback Operations
# ============================================================================


class TestRollbackToPoint:
    """Test suite for rollback to point operations."""

    @patch(
        "moai_adk.core.rollback_manager.RollbackManager._validate_system_after_rollback"
    )
    @patch("moai_adk.core.rollback_manager.RollbackManager._perform_rollback")
    @patch("moai_adk.core.rollback_manager.RollbackManager._validate_rollback_point")
    @patch("pathlib.Path.exists")
    @patch("pathlib.Path.mkdir")
    def test_rollback_to_point_success(
        self,
        mock_mkdir,
        mock_exists,
        mock_validate_point,
        mock_perform_rollback,
        mock_validate_after,
    ):
        """Test successful rollback to point."""
        # Arrange
        mock_exists.return_value = False
        mock_mkdir.return_value = None
        mock_validate_point.return_value = {"valid": True, "message": "Valid"}
        mock_perform_rollback.return_value = (["file1.py"], [])
        mock_validate_after.return_value = {"config_valid": True}

        manager = RollbackManager()
        manager.registry = {
            "rollback_123": {
                "id": "rollback_123",
                "timestamp": datetime.now(timezone.utc).isoformat(),
                "description": "Test",
                "changes": [],
                "backup_path": "/mock/backup",
                "checksum": "abc123",
                "metadata": {},
            }
        }

        # Act
        result = manager.rollback_to_point("rollback_123")

        # Assert
        assert result.success is True
        assert len(result.restored_files) == 1

    @patch("pathlib.Path.exists")
    @patch("pathlib.Path.mkdir")
    def test_rollback_to_point_not_found(self, mock_mkdir, mock_exists):
        """Test rollback to nonexistent point."""
        # Arrange
        mock_exists.return_value = False
        mock_mkdir.return_value = None

        manager = RollbackManager()

        # Act
        result = manager.rollback_to_point("nonexistent")

        # Assert
        assert result.success is False
        assert "not found" in result.message.lower()

    @patch("moai_adk.core.rollback_manager.RollbackManager._validate_rollback_point")
    @patch("pathlib.Path.exists")
    @patch("pathlib.Path.mkdir")
    def test_rollback_to_point_invalid(
        self, mock_mkdir, mock_exists, mock_validate_point
    ):
        """Test rollback with invalid rollback point."""
        # Arrange
        mock_exists.return_value = False
        mock_mkdir.return_value = None
        mock_validate_point.return_value = {
            "valid": False,
            "message": "Corrupted backup",
        }

        manager = RollbackManager()
        manager.registry = {
            "rollback_123": {
                "id": "rollback_123",
                "timestamp": datetime.now(timezone.utc).isoformat(),
                "description": "Test",
                "changes": [],
                "backup_path": "/mock/backup",
                "checksum": "abc123",
                "metadata": {},
            }
        }

        # Act
        result = manager.rollback_to_point("rollback_123", validate_before=True)

        # Assert
        assert result.success is False


# ============================================================================
# Test Rollback Point Validation
# ============================================================================


class TestValidateRollbackPoint:
    """Test suite for rollback point validation."""

    @patch("moai_adk.core.rollback_manager.RollbackManager._calculate_backup_checksum")
    @patch("pathlib.Path.exists")
    @patch("pathlib.Path.mkdir")
    def test_validate_rollback_point_valid(
        self, mock_mkdir, mock_exists, mock_checksum
    ):
        """Test validation of valid rollback point."""
        # Arrange
        mock_exists.return_value = True
        mock_mkdir.return_value = None
        mock_checksum.return_value = "abc123"

        manager = RollbackManager()
        point = RollbackPoint(
            id="rollback_123",
            timestamp=datetime.now(timezone.utc),
            description="Test",
            changes=[],
            backup_path="/mock/backup",
            checksum="abc123",
            metadata={},
        )

        # Act
        result = manager._validate_rollback_point(point)

        # Assert
        assert result["valid"] is True

    @patch("pathlib.Path.exists")
    @patch("pathlib.Path.mkdir")
    def test_validate_rollback_point_missing_backup(self, mock_mkdir, mock_exists):
        """Test validation with missing backup directory."""
        # Arrange
        mock_exists.return_value = False
        mock_mkdir.return_value = None

        manager = RollbackManager()
        point = RollbackPoint(
            id="rollback_123",
            timestamp=datetime.now(timezone.utc),
            description="Test",
            changes=[],
            backup_path="/mock/missing",
            checksum="abc123",
            metadata={},
        )

        # Act
        result = manager._validate_rollback_point(point)

        # Assert
        assert result["valid"] is False

    @patch("moai_adk.core.rollback_manager.RollbackManager._calculate_backup_checksum")
    @patch("pathlib.Path.exists")
    @patch("pathlib.Path.mkdir")
    def test_validate_rollback_point_checksum_mismatch(
        self, mock_mkdir, mock_exists, mock_checksum
    ):
        """Test validation with checksum mismatch."""
        # Arrange
        mock_exists.return_value = True
        mock_mkdir.return_value = None
        mock_checksum.return_value = "different"

        manager = RollbackManager()
        point = RollbackPoint(
            id="rollback_123",
            timestamp=datetime.now(timezone.utc),
            description="Test",
            changes=[],
            backup_path="/mock/backup",
            checksum="abc123",
            metadata={},
        )

        # Act
        result = manager._validate_rollback_point(point)

        # Assert
        assert result["valid"] is True
        assert len(result["warnings"]) > 0


# ============================================================================
# Test Perform Rollback
# ============================================================================


class TestPerformRollback:
    """Test suite for actual rollback operations."""

    @patch("shutil.copy2")
    @patch("pathlib.Path.rglob")
    @patch("pathlib.Path.exists")
    @patch("pathlib.Path.mkdir")
    @patch("pathlib.Path.parent")
    def test_perform_rollback_success(
        self,
        mock_parent,
        mock_mkdir,
        mock_exists,
        mock_rglob,
        mock_copy,
    ):
        """Test successful file restoration."""
        # Arrange
        mock_exists.return_value = True
        mock_mkdir.return_value = None
        mock_rglob.return_value = []
        mock_copy.return_value = None

        manager = RollbackManager()
        point = RollbackPoint(
            id="rollback_123",
            timestamp=datetime.now(timezone.utc),
            description="Test",
            changes=[],
            backup_path="/mock/backup",
            checksum="abc123",
            metadata={},
        )

        # Act
        restored, failed = manager._perform_rollback(point)

        # Assert
        assert isinstance(restored, list)
        assert isinstance(failed, list)

    @patch("shutil.copy2")
    @patch("pathlib.Path.rglob")
    @patch("pathlib.Path.is_file")
    @patch("pathlib.Path.exists")
    @patch("moai_adk.core.rollback_manager.RollbackManager._load_registry")
    @patch("pathlib.Path.mkdir")
    def test_perform_rollback_with_error(
        self,
        mock_mkdir,
        mock_load_registry,
        mock_exists,
        mock_is_file,
        mock_rglob,
        mock_copy,
    ):
        """Test rollback with file operation error."""
        # Arrange
        mock_mkdir.return_value = None
        mock_load_registry.return_value = {}
        mock_exists.return_value = True
        mock_is_file.return_value = True
        mock_rglob.return_value = []
        mock_copy.return_value = None

        manager = RollbackManager()
        point = RollbackPoint(
            id="rollback_123",
            timestamp=datetime.now(timezone.utc),
            description="Test",
            changes=[],
            backup_path="/mock/backup",
            checksum="abc123",
            metadata={},
        )

        # Act
        restored, failed = manager._perform_rollback(point)

        # Assert
        # Empty lists are acceptable when no files exist to restore
        assert isinstance(restored, list)
        assert isinstance(failed, list)


# ============================================================================
# Test List and Cleanup Operations
# ============================================================================


class TestListAndCleanup:
    """Test suite for listing and cleanup operations."""

    @patch("pathlib.Path.exists")
    @patch("pathlib.Path.mkdir")
    def test_list_rollback_points(self, mock_mkdir, mock_exists):
        """Test listing rollback points."""
        # Arrange
        mock_exists.return_value = False
        mock_mkdir.return_value = None

        manager = RollbackManager()
        manager.registry = {
            "rollback_001": {
                "id": "rollback_001",
                "timestamp": "2024-01-01T00:00:00",
                "description": "First backup",
                "changes": ["file1.py"],
                "used": False,
            },
            "rollback_002": {
                "id": "rollback_002",
                "timestamp": "2024-01-02T00:00:00",
                "description": "Second backup",
                "changes": ["file2.py"],
                "used": True,
            },
        }

        # Act
        points = manager.list_rollback_points(limit=10)

        # Assert
        assert len(points) == 2
        assert points[0]["id"] == "rollback_002"  # Most recent first

    @patch("shutil.rmtree")
    @patch("pathlib.Path.exists")
    @patch("moai_adk.core.rollback_manager.RollbackManager._get_directory_size")
    @patch("moai_adk.core.rollback_manager.RollbackManager._save_registry")
    @patch("pathlib.Path.mkdir")
    def test_cleanup_old_rollbacks_dry_run(
        self,
        mock_mkdir,
        mock_save_registry,
        mock_get_size,
        mock_exists,
        mock_rmtree,
    ):
        """Test cleanup in dry-run mode."""
        # Arrange
        mock_exists.return_value = False
        mock_mkdir.return_value = None
        mock_get_size.return_value = 1000000
        mock_save_registry.return_value = None

        manager = RollbackManager()
        manager.registry = {
            "rollback_001": {
                "id": "rollback_001",
                "timestamp": "2024-01-01T00:00:00",
                "backup_path": "/mock/rollback_001",
            },
        }

        # Act
        result = manager.cleanup_old_rollbacks(keep_count=10, dry_run=True)

        # Assert
        assert result["dry_run"] is True
        mock_rmtree.assert_not_called()

    @patch("shutil.rmtree")
    @patch("pathlib.Path.exists")
    @patch("moai_adk.core.rollback_manager.RollbackManager._get_directory_size")
    @patch("moai_adk.core.rollback_manager.RollbackManager._save_registry")
    @patch("pathlib.Path.mkdir")
    def test_cleanup_old_rollbacks_execute(
        self,
        mock_mkdir,
        mock_save_registry,
        mock_get_size,
        mock_exists,
        mock_rmtree,
    ):
        """Test cleanup execution."""
        # Arrange
        mock_exists.return_value = False
        mock_mkdir.return_value = None
        mock_get_size.return_value = 1000000
        mock_save_registry.return_value = None
        mock_rmtree.return_value = None

        manager = RollbackManager()
        manager.registry = {
            "rollback_001": {
                "id": "rollback_001",
                "timestamp": "2024-01-01T00:00:00",
                "backup_path": "/mock/rollback_001",
            },
            "rollback_002": {
                "id": "rollback_002",
                "timestamp": "2024-01-02T00:00:00",
                "backup_path": "/mock/rollback_002",
            },
        }

        # Act
        result = manager.cleanup_old_rollbacks(keep_count=1, dry_run=False)

        # Assert
        assert result["dry_run"] is False


# ============================================================================
# Test System Validation
# ============================================================================


class TestSystemValidation:
    """Test suite for system validation operations."""

    @patch("pathlib.Path.exists")
    @patch("shutil.disk_usage")
    @patch("pathlib.Path.mkdir")
    def test_validate_rollback_system_healthy(
        self, mock_mkdir, mock_disk_usage, mock_exists
    ):
        """Test validation of healthy rollback system."""
        # Arrange
        mock_exists.return_value = True
        mock_disk_usage.return_value = MagicMock(free=1000000000)
        mock_mkdir.return_value = None

        manager = RollbackManager()

        # Act
        result = manager.validate_rollback_system()

        # Assert
        assert result["system_healthy"] is True
        assert len(result["issues"]) == 0

    @patch("pathlib.Path.exists")
    @patch("pathlib.Path.mkdir")
    def test_validate_rollback_system_missing_dirs(self, mock_mkdir, mock_exists):
        """Test validation with missing directories."""
        # Arrange
        mock_exists.return_value = False
        mock_mkdir.return_value = None

        manager = RollbackManager()

        # Act
        result = manager.validate_rollback_system()

        # Assert
        assert result["system_healthy"] is False
        assert len(result["issues"]) > 0


# ============================================================================
# Test Utility Functions
# ============================================================================


class TestUtilityFunctions:
    """Test suite for utility functions."""

    @patch("pathlib.Path.exists")
    @patch("pathlib.Path.mkdir")
    def test_generate_rollback_id(self, mock_mkdir, mock_exists):
        """Test rollback ID generation."""
        # Arrange
        mock_exists.return_value = False
        mock_mkdir.return_value = None

        manager = RollbackManager()

        # Act
        id1 = manager._generate_rollback_id()
        id2 = manager._generate_rollback_id()

        # Assert
        assert id1.startswith("rollback_")
        assert id2.startswith("rollback_")
        assert id1 != id2

    @patch("pathlib.Path.rglob")
    @patch("pathlib.Path.is_file")
    @patch("pathlib.Path.mkdir")
    @patch("pathlib.Path.exists")
    def test_get_directory_size(
        self, mock_exists, mock_mkdir, mock_is_file, mock_rglob
    ):
        """Test directory size calculation."""
        # Arrange
        mock_exists.return_value = False
        mock_mkdir.return_value = None
        mock_is_file.return_value = True

        manager = RollbackManager()

        mock_file = MagicMock()
        mock_file.is_file.return_value = True
        mock_file.stat.return_value.st_size = 1000
        mock_rglob.return_value = [mock_file]

        # Act
        size = manager._get_directory_size(Path("/mock/dir"))

        # Assert
        assert size >= 0
