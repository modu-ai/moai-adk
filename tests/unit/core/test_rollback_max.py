"""
Comprehensive test coverage for rollback_manager.py

Target: 100+ lines of coverage for ALL rollback operations
Strategy: Maximum test coverage with mocking file operations
"""

import json
import tempfile
from datetime import datetime, timezone
from pathlib import Path

import pytest

from moai_adk.core.rollback_manager import (
    RollbackManager,
    RollbackPoint,
    RollbackResult,
)


class TestRollbackPoint:
    """Tests for RollbackPoint dataclass"""

    def test_rollback_point_creation(self):
        """Test RollbackPoint creation"""
        now = datetime.now(timezone.utc)
        point = RollbackPoint(
            id="rb_123",
            timestamp=now,
            description="Test rollback",
            changes=["file1.py", "file2.py"],
            backup_path="/backup/rb_123",
            checksum="abc123def456",
            metadata={"version": "1.0.0"},
        )
        assert point.id == "rb_123"
        assert point.description == "Test rollback"
        assert len(point.changes) == 2


class TestRollbackResult:
    """Tests for RollbackResult dataclass"""

    def test_rollback_result_success(self):
        """Test successful RollbackResult"""
        result = RollbackResult(
            success=True,
            rollback_point_id="rb_123",
            message="Rollback completed successfully",
            restored_files=["file1.py", "file2.py"],
        )
        assert result.success is True
        assert len(result.restored_files) == 2

    def test_rollback_result_failure(self):
        """Test failed RollbackResult"""
        result = RollbackResult(
            success=False,
            rollback_point_id="rb_123",
            message="Rollback failed",
            restored_files=[],
            failed_files=["file3.py"],
        )
        assert result.success is False
        assert len(result.failed_files) == 1


class TestRollbackManager:
    """Tests for RollbackManager"""

    def test_rollback_manager_init(self):
        """Test RollbackManager initialization"""
        with tempfile.TemporaryDirectory() as tmp_dir:
            project_root = Path(tmp_dir)
            manager = RollbackManager(project_root=project_root)

            assert manager.project_root == project_root
            assert manager.backup_root == project_root / ".moai" / "rollbacks"
            assert manager.backup_root.exists()

    def test_rollback_manager_creates_directories(self):
        """Test that RollbackManager creates necessary directories"""
        with tempfile.TemporaryDirectory() as tmp_dir:
            project_root = Path(tmp_dir)
            manager = RollbackManager(project_root=project_root)

            assert manager.config_backup_dir.exists()
            assert manager.code_backup_dir.exists()
            assert manager.docs_backup_dir.exists()

    def test_generate_rollback_id(self):
        """Test generating rollback ID"""
        with tempfile.TemporaryDirectory() as tmp_dir:
            manager = RollbackManager(project_root=Path(tmp_dir))
            id1 = manager._generate_rollback_id()
            id2 = manager._generate_rollback_id()

            assert id1.startswith("rollback_")
            assert id2.startswith("rollback_")
            assert id1 != id2

    def test_load_registry_empty(self):
        """Test loading empty registry"""
        with tempfile.TemporaryDirectory() as tmp_dir:
            manager = RollbackManager(project_root=Path(tmp_dir))
            registry = manager._load_registry()

            assert isinstance(registry, dict)
            assert len(registry) == 0

    def test_save_registry(self):
        """Test saving registry"""
        with tempfile.TemporaryDirectory() as tmp_dir:
            manager = RollbackManager(project_root=Path(tmp_dir))
            manager.registry = {"test_id": {"description": "test"}}
            manager._save_registry()

            # Verify file was saved
            assert manager.registry_file.exists()

            # Load and verify content
            with open(manager.registry_file, "r") as f:
                saved_registry = json.load(f)
                assert "test_id" in saved_registry

    def test_create_rollback_point_basic(self):
        """Test creating basic rollback point"""
        with tempfile.TemporaryDirectory() as tmp_dir:
            project_root = Path(tmp_dir)
            manager = RollbackManager(project_root=project_root)

            rollback_id = manager.create_rollback_point(
                description="Test rollback",
                changes=["file1.py"],
            )

            assert rollback_id in manager.registry
            assert manager.registry[rollback_id]["description"] == "Test rollback"

    def test_create_rollback_point_creates_directories(self):
        """Test that creating rollback point creates backup directories"""
        with tempfile.TemporaryDirectory() as tmp_dir:
            project_root = Path(tmp_dir)

            # Create a .moai/config directory with a config file
            config_dir = project_root / ".moai" / "config"
            config_dir.mkdir(parents=True, exist_ok=True)
            config_file = config_dir / "config.json"
            config_file.write_text('{"test": "value"}')

            manager = RollbackManager(project_root=project_root)
            rollback_id = manager.create_rollback_point("Test rollback")

            backup_dir = manager.backup_root / rollback_id
            assert backup_dir.exists()

    def test_list_rollback_points_empty(self):
        """Test listing rollback points when empty"""
        with tempfile.TemporaryDirectory() as tmp_dir:
            manager = RollbackManager(project_root=Path(tmp_dir))
            points = manager.list_rollback_points()

            assert len(points) == 0

    def test_list_rollback_points_multiple(self):
        """Test listing multiple rollback points"""
        with tempfile.TemporaryDirectory() as tmp_dir:
            project_root = Path(tmp_dir)
            manager = RollbackManager(project_root=project_root)

            # Create multiple rollback points
            for i in range(3):
                manager.create_rollback_point(f"Rollback {i}")

            points = manager.list_rollback_points()
            assert len(points) == 3

    def test_list_rollback_points_with_limit(self):
        """Test listing rollback points with limit"""
        with tempfile.TemporaryDirectory() as tmp_dir:
            project_root = Path(tmp_dir)
            manager = RollbackManager(project_root=project_root)

            # Create multiple rollback points
            for i in range(5):
                manager.create_rollback_point(f"Rollback {i}")

            points = manager.list_rollback_points(limit=2)
            assert len(points) <= 2

    def test_rollback_to_point_not_found(self):
        """Test rollback to non-existent point"""
        with tempfile.TemporaryDirectory() as tmp_dir:
            manager = RollbackManager(project_root=Path(tmp_dir))
            result = manager.rollback_to_point("nonexistent_id")

            assert result.success is False
            assert "not found" in result.message.lower()

    def test_rollback_research_integration_no_points(self):
        """Test research rollback with no suitable points"""
        with tempfile.TemporaryDirectory() as tmp_dir:
            manager = RollbackManager(project_root=Path(tmp_dir))
            result = manager.rollback_research_integration()

            assert result.success is False

    def test_validate_rollback_system_healthy(self):
        """Test validating healthy rollback system"""
        with tempfile.TemporaryDirectory() as tmp_dir:
            manager = RollbackManager(project_root=Path(tmp_dir))
            validation = manager.validate_rollback_system()

            assert validation["system_healthy"] is True
            assert validation["rollback_points_count"] == 0

    def test_validate_rollback_system_missing_dirs(self):
        """Test validation detects missing directories"""
        with tempfile.TemporaryDirectory() as tmp_dir:
            manager = RollbackManager(project_root=Path(tmp_dir))

            # Remove a directory
            import shutil

            shutil.rmtree(manager.config_backup_dir)

            validation = manager.validate_rollback_system()
            assert validation["system_healthy"] is False
            assert len(validation["issues"]) > 0

    def test_cleanup_old_rollbacks_dry_run(self):
        """Test cleanup with dry run"""
        with tempfile.TemporaryDirectory() as tmp_dir:
            project_root = Path(tmp_dir)
            manager = RollbackManager(project_root=project_root)

            # Create some rollback points
            for i in range(3):
                manager.create_rollback_point(f"Rollback {i}")

            result = manager.cleanup_old_rollbacks(keep_count=1, dry_run=True)

            assert result["dry_run"] is True
            assert result["would_delete_count"] == 2

    def test_cleanup_old_rollbacks_execute(self):
        """Test cleanup execution"""
        with tempfile.TemporaryDirectory() as tmp_dir:
            project_root = Path(tmp_dir)
            manager = RollbackManager(project_root=project_root)

            # Create some rollback points
            for i in range(3):
                manager.create_rollback_point(f"Rollback {i}")

            len(manager.registry)

            result = manager.cleanup_old_rollbacks(keep_count=1, dry_run=False)

            assert result["dry_run"] is False
            assert result["deleted_count"] > 0

    def test_calculate_backup_size(self):
        """Test calculating backup size"""
        with tempfile.TemporaryDirectory() as tmp_dir:
            project_root = Path(tmp_dir)
            manager = RollbackManager(project_root=project_root)

            # Create rollback point
            manager.create_rollback_point("Test rollback")

            size = manager._calculate_backup_size()
            assert size >= 0

    def test_get_directory_size(self):
        """Test getting directory size"""
        with tempfile.TemporaryDirectory() as tmp_dir:
            project_root = Path(tmp_dir)
            manager = RollbackManager(project_root=project_root)

            # Create test files
            test_dir = project_root / "test_dir"
            test_dir.mkdir()
            (test_dir / "file1.txt").write_text("test content")
            (test_dir / "file2.txt").write_text("more test content")

            size = manager._get_directory_size(test_dir)
            assert size > 0

    def test_backup_configuration(self):
        """Test backing up configuration files"""
        with tempfile.TemporaryDirectory() as tmp_dir:
            project_root = Path(tmp_dir)

            # Create config files
            config_dir = project_root / ".moai" / "config"
            config_dir.mkdir(parents=True, exist_ok=True)
            (config_dir / "config.json").write_text('{"test": "value"}')

            claude_dir = project_root / ".claude"
            claude_dir.mkdir()
            (claude_dir / "settings.json").write_text('{"setting": "value"}')

            manager = RollbackManager(project_root=project_root)
            rollback_dir = project_root / ".moai" / "rollbacks" / "test"
            rollback_dir.mkdir(parents=True, exist_ok=True)

            backup_path = manager._backup_configuration(rollback_dir)
            assert Path(backup_path).exists()

    def test_mark_rollback_as_used(self):
        """Test marking rollback as used"""
        with tempfile.TemporaryDirectory() as tmp_dir:
            project_root = Path(tmp_dir)
            manager = RollbackManager(project_root=project_root)

            rollback_id = manager.create_rollback_point("Test rollback")
            manager._mark_rollback_as_used(rollback_id)

            assert manager.registry[rollback_id]["used"] is True
            assert "used_timestamp" in manager.registry[rollback_id]

    def test_cleanup_partial_backup(self):
        """Test cleaning up partial backup"""
        with tempfile.TemporaryDirectory() as tmp_dir:
            project_root = Path(tmp_dir)
            manager = RollbackManager(project_root=project_root)

            # Create partial backup directory
            partial_dir = manager.backup_root / "partial_backup"
            partial_dir.mkdir()
            (partial_dir / "test.txt").write_text("test")

            manager._cleanup_partial_backup("partial_backup")

            # Directory should be removed
            assert not partial_dir.exists()

    def test_calculate_backup_checksum(self):
        """Test calculating backup checksum"""
        with tempfile.TemporaryDirectory() as tmp_dir:
            project_root = Path(tmp_dir)
            manager = RollbackManager(project_root=project_root)

            # Create files to backup
            backup_dir = project_root / "backup_test"
            backup_dir.mkdir()
            (backup_dir / "file1.txt").write_text("content1")
            (backup_dir / "file2.txt").write_text("content2")

            checksum1 = manager._calculate_backup_checksum(backup_dir)
            assert checksum1

            # Calculate again - should be same
            checksum2 = manager._calculate_backup_checksum(backup_dir)
            assert checksum1 == checksum2

    def test_validate_rollback_point_valid(self):
        """Test validating valid rollback point"""
        with tempfile.TemporaryDirectory() as tmp_dir:
            project_root = Path(tmp_dir)
            manager = RollbackManager(project_root=project_root)

            # Create a valid rollback point
            rollback_id = manager.create_rollback_point("Test rollback")
            rollback_point = RollbackPoint(**manager.registry[rollback_id])

            validation = manager._validate_rollback_point(rollback_point)
            assert validation["valid"] is True

    def test_validate_rollback_point_missing_backup(self):
        """Test validating rollback point with missing backup"""
        with tempfile.TemporaryDirectory() as tmp_dir:
            project_root = Path(tmp_dir)
            manager = RollbackManager(project_root=project_root)

            # Create rollback point with non-existent path
            point = RollbackPoint(
                id="missing",
                timestamp=datetime.now(timezone.utc),
                description="Missing backup",
                changes=[],
                backup_path="/nonexistent/path",
                checksum="dummy",
                metadata={},
            )

            validation = manager._validate_rollback_point(point)
            assert validation["valid"] is False

    def test_find_research_rollback_points_empty(self):
        """Test finding research rollback points when none exist"""
        with tempfile.TemporaryDirectory() as tmp_dir:
            manager = RollbackManager(project_root=Path(tmp_dir))
            points = manager._find_research_rollback_points()

            assert len(points) == 0

    def test_validate_system_after_rollback(self):
        """Test validating system after rollback"""
        with tempfile.TemporaryDirectory() as tmp_dir:
            project_root = Path(tmp_dir)

            # Create config file
            config_dir = project_root / ".moai" / "config"
            config_dir.mkdir(parents=True, exist_ok=True)
            (config_dir / "config.json").write_text('{"valid": "json"}')

            manager = RollbackManager(project_root=project_root)
            validation = manager._validate_system_after_rollback()

            assert validation["config_valid"] is True

    def test_validate_system_after_rollback_invalid_json(self):
        """Test validating system with invalid JSON"""
        with tempfile.TemporaryDirectory() as tmp_dir:
            project_root = Path(tmp_dir)

            # Create invalid config file
            config_dir = project_root / ".moai" / "config"
            config_dir.mkdir(parents=True, exist_ok=True)
            (config_dir / "config.json").write_text("invalid json {")

            manager = RollbackManager(project_root=project_root)
            validation = manager._validate_system_after_rollback()

            assert validation["config_valid"] is False

    def test_validate_research_components(self):
        """Test validating research components"""
        with tempfile.TemporaryDirectory() as tmp_dir:
            project_root = Path(tmp_dir)

            # Create research directories with valid files
            skills_dir = project_root / ".claude" / "skills"
            skills_dir.mkdir(parents=True, exist_ok=True)
            (skills_dir / "skill1.md").write_text("# Skill 1\nContent")

            manager = RollbackManager(project_root=project_root)
            validation = manager._validate_research_components()

            assert validation["skills_valid"] is True

    def test_create_rollback_point_with_source_files(self):
        """Test creating rollback with actual source files"""
        with tempfile.TemporaryDirectory() as tmp_dir:
            project_root = Path(tmp_dir)

            # Create source structure
            src_dir = project_root / "src"
            src_dir.mkdir()
            (src_dir / "main.py").write_text("print('hello')")

            tests_dir = project_root / "tests"
            tests_dir.mkdir()
            (tests_dir / "test_main.py").write_text("def test(): pass")

            manager = RollbackManager(project_root=project_root)
            rollback_id = manager.create_rollback_point("Test with sources")

            # Verify files were backed up
            backup_dir = manager.backup_root / rollback_id
            code_backup = backup_dir / "code"
            assert code_backup.exists()

    def test_concurrent_rollback_operations(self):
        """Test concurrent rollback operations"""
        with tempfile.TemporaryDirectory() as tmp_dir:
            project_root = Path(tmp_dir)
            manager = RollbackManager(project_root=project_root)

            import threading

            rollback_ids = []

            def create_rollback():
                rb_id = manager.create_rollback_point("Concurrent rollback")
                rollback_ids.append(rb_id)

            threads = [threading.Thread(target=create_rollback) for _ in range(3)]
            for t in threads:
                t.start()
            for t in threads:
                t.join()

            assert len(rollback_ids) == 3
            assert len(manager.registry) == 3


if __name__ == "__main__":
    pytest.main([__file__, "-v", "--tb=short"])
