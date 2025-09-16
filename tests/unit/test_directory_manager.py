"""
Unit tests for the directory manager module.

Tests the DirectoryManager class and its directory operation methods
to ensure proper directory management and security validation.
"""

import pytest
import tempfile
import shutil
from pathlib import Path
from unittest.mock import patch, MagicMock, call

from moai_adk.core.directory_manager import DirectoryManager
from moai_adk.core.security import SecurityManager, SecurityError
from moai_adk.config import Config, RuntimeConfig

class TestDirectoryManager:
    """Test cases for DirectoryManager class."""

    @pytest.fixture
    def temp_dir(self):
        """Create a temporary directory for testing."""
        with tempfile.TemporaryDirectory() as temp_dir:
            yield Path(temp_dir)

    @pytest.fixture
    def security_manager(self):
        """Create a mock security manager for testing."""
        return MagicMock(spec=SecurityManager)

    @pytest.fixture
    def dir_manager(self, security_manager):
        """Create a DirectoryManager instance for testing."""
        return DirectoryManager(security_manager)

    @pytest.fixture
    def sample_config(self, temp_dir):
        """Create a sample Config instance for testing."""
        return Config(
            name="test-project",
            
            template="standard",
            runtime=RuntimeConfig("python"),
            path=temp_dir / "test_project"
        )

    def test_init_with_security_manager(self, security_manager):
        """Test DirectoryManager initialization with security manager."""
        manager = DirectoryManager(security_manager)
        assert manager.security_manager == security_manager

    def test_init_without_security_manager(self):
        """Test DirectoryManager initialization without security manager."""
        manager = DirectoryManager()
        assert isinstance(manager.security_manager, SecurityManager)

    def test_create_project_directory_new_directory(self, dir_manager, temp_dir):
        """Test creating a new project directory."""
        config = Config(
            name="new-project",
            
            template="standard",
            runtime=RuntimeConfig("python"),
            path=temp_dir / "new_project"
        )

        dir_manager.create_project_directory(config)

        assert config.project_path.exists()
        assert config.project_path.is_dir()

    def test_create_project_directory_current_directory(self, dir_manager, temp_dir):
        """Test creating project directory when using current directory."""
        config = Config(
            name="current-project",
            
            template="standard",
            runtime=RuntimeConfig("python"),
            path=Path(".")
        )

        # Should not fail and should preserve existing structure
        dir_manager.create_project_directory(config)
        # No assertion needed - should not crash

    def test_create_project_directory_existing_no_overwrite(self, dir_manager, temp_dir):
        """Test creating project directory when directory exists without overwrite."""
        project_path = temp_dir / "existing_project"
        project_path.mkdir()
        (project_path / "existing_file.txt").write_text("existing content")

        config = Config(
            name="existing-project",
            
            template="standard",
            runtime=RuntimeConfig("python"),
            path=project_path,
            is_existing_project=True,
            force_overwrite=False
        )

        dir_manager.create_project_directory(config)

        # Should preserve existing files
        assert (project_path / "existing_file.txt").exists()
        assert (project_path / "existing_file.txt").read_text() == "existing content"

    def test_create_project_directory_force_overwrite(self, dir_manager, temp_dir):
        """Test creating project directory with force overwrite."""
        project_path = temp_dir / "overwrite_project"
        project_path.mkdir()
        (project_path / "old_file.txt").write_text("old content")

        config = Config(
            name="overwrite-project",
            
            template="standard",
            runtime=RuntimeConfig("python"),
            path=project_path,
            force_overwrite=True
        )

        # Mock security manager to allow safe removal
        dir_manager.security_manager.safe_rmtree.return_value = True

        dir_manager.create_project_directory(config)

        # Old file should be gone, directory should exist
        assert project_path.exists()
        assert not (project_path / "old_file.txt").exists()

    def test_handle_force_overwrite_with_git(self, dir_manager, temp_dir):
        """Test force overwrite while preserving .git directory."""
        project_path = temp_dir / "git_project"
        project_path.mkdir()

        # Create .git directory with content
        git_dir = project_path / ".git"
        git_dir.mkdir()
        (git_dir / "config").write_text("git config content")

        # Create other files to be removed
        (project_path / "old_file.txt").write_text("old content")

        # Mock security manager to allow safe removal
        dir_manager.security_manager.safe_rmtree.return_value = True

        dir_manager._handle_force_overwrite(project_path)

        # .git directory should be preserved
        assert git_dir.exists()
        assert (git_dir / "config").exists()
        assert (git_dir / "config").read_text() == "git config content"

        # Old files should be gone
        assert not (project_path / "old_file.txt").exists()

    def test_handle_force_overwrite_security_failure(self, dir_manager, temp_dir):
        """Test force overwrite when security validation fails."""
        project_path = temp_dir / "security_fail_project"
        project_path.mkdir()

        # Mock security manager to fail safe removal
        dir_manager.security_manager.safe_rmtree.return_value = False

        with pytest.raises(SecurityError):
            dir_manager._handle_force_overwrite(project_path)

    def test_create_directory_structure_success(self, dir_manager, temp_dir):
        """Test successful creation of MoAI-ADK directory structure."""
        # Mock security validation to return True
        dir_manager.security_manager.validate_path_safety.return_value = True

        created_dirs = dir_manager.create_directory_structure(temp_dir)

        # Should create standard directories
        expected_dirs = [
            ".claude",
            ".claude/commands/moai",
            ".claude/agents/moai",
            ".claude/hooks/moai",
            ".claude/memory",
            ".claude/logs",
            ".claude/output-styles",
            ".moai",
            ".moai/templates",
            ".moai/steering",
            ".moai/memory",
            ".moai/memory/decisions",
            ".moai/specs",
            ".moai/indexes",
            ".moai/reports",
            ".moai/scripts",
            ".github/workflows"
        ]

        assert len(created_dirs) == len(expected_dirs)

        for dir_name in expected_dirs:
            expected_path = temp_dir / dir_name
            assert expected_path.exists()
            assert expected_path.is_dir()

    def test_create_directory_structure_security_validation_fails(self, dir_manager, temp_dir):
        """Test directory structure creation when security validation fails."""
        # Mock security validation to return False
        dir_manager.security_manager.validate_path_safety.return_value = False

        created_dirs = dir_manager.create_directory_structure(temp_dir)

        # Should create no directories due to security failure
        assert created_dirs == []

    def test_create_directory_structure_partial_failure(self, dir_manager, temp_dir):
        """Test directory structure creation with partial failures."""
        # Mock security validation to fail for some directories
        def mock_validate_path_safety(path, base):
            # Fail validation for .claude directories
            return ".claude" not in str(path)

        dir_manager.security_manager.validate_path_safety.side_effect = mock_validate_path_safety

        created_dirs = dir_manager.create_directory_structure(temp_dir)

        # Should only create .moai and .github directories
        assert len(created_dirs) > 0
        assert all(".claude" not in str(d) for d in created_dirs)
        assert any(".moai" in str(d) for d in created_dirs)

    def test_ensure_directory_exists_new_directory(self, dir_manager, temp_dir):
        """Test ensuring a new directory exists."""
        new_dir = temp_dir / "new_directory"

        # Mock security validation to return True
        dir_manager.security_manager.validate_path_safety.return_value = True

        result = dir_manager.ensure_directory_exists(new_dir, temp_dir)

        assert result is True
        assert new_dir.exists()
        assert new_dir.is_dir()

    def test_ensure_directory_exists_existing_directory(self, dir_manager, temp_dir):
        """Test ensuring an existing directory exists."""
        existing_dir = temp_dir / "existing_directory"
        existing_dir.mkdir()

        result = dir_manager.ensure_directory_exists(existing_dir, temp_dir)

        assert result is True
        assert existing_dir.exists()

    def test_ensure_directory_exists_security_validation_fails(self, dir_manager, temp_dir):
        """Test ensuring directory exists when security validation fails."""
        new_dir = temp_dir / "blocked_directory"

        # Mock security validation to return False
        dir_manager.security_manager.validate_path_safety.return_value = False

        result = dir_manager.ensure_directory_exists(new_dir, temp_dir)

        assert result is False
        assert not new_dir.exists()

    def test_ensure_directory_exists_without_base_path(self, dir_manager, temp_dir):
        """Test ensuring directory exists without base path validation."""
        new_dir = temp_dir / "no_validation_dir"

        result = dir_manager.ensure_directory_exists(new_dir)

        assert result is True
        assert new_dir.exists()

    def test_ensure_directory_exists_permission_error(self, dir_manager, temp_dir):
        """Test ensuring directory exists with permission error."""
        new_dir = temp_dir / "permission_error_dir"

        # Mock mkdir to raise PermissionError
        with patch.object(Path, 'mkdir', side_effect=PermissionError("Permission denied")):
            result = dir_manager.ensure_directory_exists(new_dir)

        assert result is False
        assert not new_dir.exists()

    def test_get_directory_info_nonexistent(self, dir_manager, temp_dir):
        """Test getting info for nonexistent directory."""
        nonexistent_dir = temp_dir / "nonexistent"

        info = dir_manager.get_directory_info(nonexistent_dir)

        expected = {
            'exists': False,
            'is_directory': False,
            'file_count': 0,
            'subdirectory_count': 0,
            'total_size': 0
        }
        assert info == expected

    def test_get_directory_info_file(self, dir_manager, temp_dir):
        """Test getting info for a file (not directory)."""
        test_file = temp_dir / "test_file.txt"
        test_file.write_text("test content")

        info = dir_manager.get_directory_info(test_file)

        assert info['exists'] is True
        assert info['is_directory'] is False
        assert info['total_size'] > 0

    def test_get_directory_info_empty_directory(self, dir_manager, temp_dir):
        """Test getting info for empty directory."""
        empty_dir = temp_dir / "empty_dir"
        empty_dir.mkdir()

        info = dir_manager.get_directory_info(empty_dir)

        expected = {
            'exists': True,
            'is_directory': True,
            'file_count': 0,
            'subdirectory_count': 0,
            'total_size': 0,
            'size_mb': 0.0
        }
        assert info == expected

    def test_get_directory_info_with_content(self, dir_manager, temp_dir):
        """Test getting info for directory with content."""
        content_dir = temp_dir / "content_dir"
        content_dir.mkdir()

        # Create files
        (content_dir / "file1.txt").write_text("content1")
        (content_dir / "file2.txt").write_text("content2")

        # Create subdirectory with file
        subdir = content_dir / "subdir"
        subdir.mkdir()
        (subdir / "file3.txt").write_text("content3")

        info = dir_manager.get_directory_info(content_dir)

        assert info['exists'] is True
        assert info['is_directory'] is True
        assert info['file_count'] == 3
        assert info['subdirectory_count'] == 1
        assert info['total_size'] > 0
        assert info['size_mb'] >= 0.0

    def test_get_directory_info_error_handling(self, dir_manager, temp_dir):
        """Test error handling in get_directory_info."""
        test_dir = temp_dir / "test_dir"
        test_dir.mkdir()

        # Mock rglob to raise an exception
        with patch.object(Path, 'rglob', side_effect=PermissionError("Permission denied")):
            info = dir_manager.get_directory_info(test_dir)

        assert info['exists'] is True
        assert info['is_directory'] is True
        assert info['file_count'] == 0
        assert 'error' in info

    def test_clean_directory_empty(self, dir_manager, temp_dir):
        """Test cleaning an empty directory."""
        empty_dir = temp_dir / "empty_dir"
        empty_dir.mkdir()

        result = dir_manager.clean_directory(empty_dir)

        assert result is True
        assert empty_dir.exists()  # Directory itself should remain

    def test_clean_directory_with_files(self, dir_manager, temp_dir):
        """Test cleaning directory with files."""
        test_dir = temp_dir / "test_dir"
        test_dir.mkdir()

        # Create test files
        (test_dir / "file1.txt").write_text("content1")
        (test_dir / "file2.txt").write_text("content2")

        # Create subdirectory
        subdir = test_dir / "subdir"
        subdir.mkdir()
        (subdir / "file3.txt").write_text("content3")

        # Mock security manager for safe removal
        dir_manager.security_manager.safe_rmtree.return_value = True

        result = dir_manager.clean_directory(test_dir)

        assert result is True
        assert test_dir.exists()  # Main directory should remain
        assert not (test_dir / "file1.txt").exists()
        assert not (test_dir / "file2.txt").exists()
        assert not subdir.exists()

    def test_clean_directory_preserve_patterns(self, dir_manager, temp_dir):
        """Test cleaning directory while preserving patterns."""
        test_dir = temp_dir / "test_dir"
        test_dir.mkdir()

        # Create test files
        (test_dir / "keep.txt").write_text("keep this")
        (test_dir / "remove.txt").write_text("remove this")
        (test_dir / "keep.log").write_text("keep this log")

        result = dir_manager.clean_directory(test_dir, preserve_patterns=["keep.*"])

        assert result is True
        assert (test_dir / "keep.txt").exists()
        assert (test_dir / "keep.log").exists()
        assert not (test_dir / "remove.txt").exists()

    def test_clean_directory_nonexistent(self, dir_manager, temp_dir):
        """Test cleaning nonexistent directory."""
        nonexistent_dir = temp_dir / "nonexistent"

        result = dir_manager.clean_directory(nonexistent_dir)

        assert result is False

    def test_clean_directory_file_instead_of_dir(self, dir_manager, temp_dir):
        """Test cleaning when path points to a file."""
        test_file = temp_dir / "test_file.txt"
        test_file.write_text("content")

        result = dir_manager.clean_directory(test_file)

        assert result is False

    def test_clean_directory_security_failure(self, dir_manager, temp_dir):
        """Test cleaning directory when security manager fails."""
        test_dir = temp_dir / "test_dir"
        test_dir.mkdir()
        subdir = test_dir / "subdir"
        subdir.mkdir()

        # Mock security manager to fail removal
        dir_manager.security_manager.safe_rmtree.return_value = False

        result = dir_manager.clean_directory(test_dir)

        assert result is False

    def test_create_backup_directory_file(self, dir_manager, temp_dir):
        """Test creating backup of a file."""
        source_file = temp_dir / "source.txt"
        source_file.write_text("source content")

        # Mock security validation to return True
        dir_manager.security_manager.validate_path_safety.return_value = True

        backup_path = dir_manager.create_backup_directory(source_file, temp_dir)

        assert backup_path.exists()
        assert backup_path.is_file()
        assert backup_path.read_text() == "source content"
        assert "backup_" in backup_path.name
        assert source_file.name in backup_path.name

    def test_create_backup_directory_directory(self, dir_manager, temp_dir):
        """Test creating backup of a directory."""
        source_dir = temp_dir / "source_dir"
        source_dir.mkdir()
        (source_dir / "file.txt").write_text("file content")

        # Mock security validation to return True
        dir_manager.security_manager.validate_path_safety.return_value = True

        backup_path = dir_manager.create_backup_directory(source_dir, temp_dir)

        assert backup_path.exists()
        assert backup_path.is_dir()
        assert (backup_path / "file.txt").exists()
        assert (backup_path / "file.txt").read_text() == "file content"

    def test_create_backup_directory_default_location(self, dir_manager, temp_dir):
        """Test creating backup with default backup location."""
        source_file = temp_dir / "source.txt"
        source_file.write_text("content")

        # Mock security validation to return True
        dir_manager.security_manager.validate_path_safety.return_value = True

        backup_path = dir_manager.create_backup_directory(source_file)

        # Should create backup in same directory as source
        assert backup_path.parent == source_file.parent
        assert backup_path.exists()

    def test_create_backup_directory_security_failure(self, dir_manager, temp_dir):
        """Test creating backup when security validation fails."""
        source_file = temp_dir / "source.txt"
        source_file.write_text("content")

        # Mock security validation to return False
        dir_manager.security_manager.validate_path_safety.return_value = False

        with pytest.raises(SecurityError):
            dir_manager.create_backup_directory(source_file, temp_dir)

    def test_create_backup_directory_nonexistent_source(self, dir_manager, temp_dir):
        """Test creating backup of nonexistent source."""
        nonexistent_source = temp_dir / "nonexistent.txt"

        # Mock security validation to return True
        dir_manager.security_manager.validate_path_safety.return_value = True

        with pytest.raises(ValueError):
            dir_manager.create_backup_directory(nonexistent_source, temp_dir)

    @patch('shutil.copytree', side_effect=IOError("Disk full"))
    def test_create_backup_directory_io_error(self, mock_copytree, dir_manager, temp_dir):
        """Test creating backup with IO error."""
        source_dir = temp_dir / "source_dir"
        source_dir.mkdir()

        # Mock security validation to return True
        dir_manager.security_manager.validate_path_safety.return_value = True

        with pytest.raises(IOError):
            dir_manager.create_backup_directory(source_dir, temp_dir)

    def test_integration_workflow(self, dir_manager, temp_dir):
        """Test complete directory management workflow."""
        # Mock security validation to return True
        dir_manager.security_manager.validate_path_safety.return_value = True
        dir_manager.security_manager.safe_rmtree.return_value = True

        project_path = temp_dir / "integration_project"

        # Create project config
        config = Config(
            name="integration-project",
            
            template="standard",
            runtime=RuntimeConfig("python"),
            path=project_path
        )

        # 1. Create project directory
        dir_manager.create_project_directory(config)
        assert project_path.exists()

        # 2. Create directory structure
        created_dirs = dir_manager.create_directory_structure(project_path)
        assert len(created_dirs) > 0

        # 3. Get directory info
        info = dir_manager.get_directory_info(project_path)
        assert info['exists'] is True
        assert info['subdirectory_count'] > 0

        # 4. Ensure additional directory exists
        additional_dir = project_path / "additional"
        result = dir_manager.ensure_directory_exists(additional_dir, project_path)
        assert result is True
        assert additional_dir.exists()

        # 5. Create backup
        backup_path = dir_manager.create_backup_directory(project_path, temp_dir)
        assert backup_path.exists()

    def test_directory_structure_completeness(self, dir_manager, temp_dir):
        """Test that all expected MoAI-ADK directories are created."""
        # Mock security validation to return True
        dir_manager.security_manager.validate_path_safety.return_value = True

        created_dirs = dir_manager.create_directory_structure(temp_dir)

        # Convert to relative paths for easier comparison
        created_relative = [d.relative_to(temp_dir) for d in created_dirs]
        created_strings = [str(p) for p in created_relative]

        # Check for key directories
        assert ".claude" in created_strings
        assert str(Path(".claude") / "commands" / "moai") in created_strings
        assert str(Path(".claude") / "agents" / "moai") in created_strings
        assert str(Path(".claude") / "hooks" / "moai") in created_strings
        assert str(Path(".moai") / "memory") in created_strings
        assert str(Path(".moai") / "scripts") in created_strings
        assert str(Path(".github") / "workflows") in created_strings