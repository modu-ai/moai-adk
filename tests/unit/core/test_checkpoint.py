"""Unit tests for checkpoint.py module

Tests for CheckpointManager class covering event-driven checkpoint creation.
"""

from pathlib import Path
from unittest.mock import MagicMock, patch

import pytest
from git import Repo


class TestCheckpointManagerInit:
    """Test CheckpointManager initialization."""

    @pytest.fixture
    def mock_repo(self):
        """Create a mock Git repo."""
        repo = MagicMock(spec=Repo)
        return repo

    @pytest.fixture
    def temp_project(self, tmp_path):
        """Create a temporary project directory."""
        project = tmp_path / "test_project"
        project.mkdir()
        (project / ".moai").mkdir()
        return project

    def test_init_creates_components(self, mock_repo, temp_project):
        """Should initialize with all required components."""
        from moai_adk.core.git.checkpoint import CheckpointManager

        manager = CheckpointManager(mock_repo, temp_project)

        assert manager.repo == mock_repo
        assert manager.project_root == temp_project
        assert manager.event_detector is not None
        assert manager.branch_manager is not None
        assert manager.log_file == temp_project / ".moai" / "checkpoints.log"

    def test_init_creates_log_directory(self, mock_repo, tmp_path):
        """Should create log directory if it doesn't exist."""
        from moai_adk.core.git.checkpoint import CheckpointManager

        project = tmp_path / "new_project"
        # Don't create .moai directory

        CheckpointManager(mock_repo, project)

        # Directory should be created
        assert (project / ".moai").exists()


class TestCreateCheckpointIfRisky:
    """Test create_checkpoint_if_risky method."""

    @pytest.fixture
    def manager(self, tmp_path):
        """Create a CheckpointManager for testing."""
        mock_repo = MagicMock(spec=Repo)
        project = tmp_path / "test_project"
        project.mkdir()
        (project / ".moai").mkdir()

        from moai_adk.core.git.checkpoint import CheckpointManager

        return CheckpointManager(mock_repo, project)

    def test_returns_none_for_safe_operation(self, manager):
        """Should return None when operation is not risky."""
        result = manager.create_checkpoint_if_risky("safe_operation")

        assert result is None

    def test_creates_checkpoint_for_risky_deletion(self, manager):
        """Should create checkpoint for risky file deletions."""
        with patch.object(manager.event_detector, "is_risky_deletion", return_value=True):
            with patch.object(manager.branch_manager, "create_checkpoint_branch", return_value="checkpoint-001"):
                result = manager.create_checkpoint_if_risky("delete", deleted_files=["file1.py", "file2.py"])

                assert result == "checkpoint-001"

    def test_creates_checkpoint_for_risky_refactoring(self, manager):
        """Should create checkpoint for large-scale refactoring."""
        with patch.object(manager.event_detector, "is_risky_refactoring", return_value=True):
            with patch.object(manager.branch_manager, "create_checkpoint_branch", return_value="checkpoint-002"):
                # This covers line 66
                result = manager.create_checkpoint_if_risky(
                    "rename",
                    renamed_files=[("old1.py", "new1.py"), ("old2.py", "new2.py")],
                )

                assert result == "checkpoint-002"

    def test_creates_checkpoint_for_critical_file_modification(self, manager, tmp_path):
        """Should create checkpoint when critical files are modified."""
        # Mock event_detector to identify a file as critical
        with patch.object(manager.event_detector, "is_critical_file", return_value=True):
            with patch.object(manager.branch_manager, "create_checkpoint_branch", return_value="checkpoint-003"):
                # This covers lines 70-73
                result = manager.create_checkpoint_if_risky("modify", modified_files=[Path("config.yaml")])

                assert result == "checkpoint-003"

    def test_stops_checking_after_first_critical_file(self, manager, tmp_path):
        """Should stop checking files after finding first critical one."""
        check_count = [0]

        def mock_is_critical(file_path):
            check_count[0] += 1
            # First file is critical, should stop there
            return check_count[0] == 1

        with patch.object(manager.event_detector, "is_critical_file", side_effect=mock_is_critical):
            with patch.object(manager.branch_manager, "create_checkpoint_branch", return_value="checkpoint-004"):
                # Check multiple files, but should stop after first critical one
                result = manager.create_checkpoint_if_risky(
                    "modify",
                    modified_files=[
                        Path("critical.py"),
                        Path("also_critical.py"),
                        Path("not_critical.py"),
                    ],
                )

                assert result == "checkpoint-004"
                # Should have stopped checking after first file (break on line 73)
                assert check_count[0] == 1

    def test_logs_checkpoint_metadata(self, manager):
        """Should log checkpoint when created."""
        with patch.object(manager.event_detector, "is_risky_deletion", return_value=True):
            with patch.object(manager.branch_manager, "create_checkpoint_branch", return_value="checkpoint-005"):
                manager.create_checkpoint_if_risky("delete", deleted_files=["file.py"])

                # Check log file contains metadata
                log_content = manager.log_file.read_text()
                assert "checkpoint-005" in log_content
                assert "delete" in log_content
                assert "is_safety: False" in log_content


class TestRestoreCheckpoint:
    """Test restore_checkpoint method."""

    @pytest.fixture
    def manager(self, tmp_path):
        """Create a CheckpointManager for testing."""
        mock_repo = MagicMock(spec=Repo)
        project = tmp_path / "test_project"
        project.mkdir()
        (project / ".moai").mkdir()

        from moai_adk.core.git.checkpoint import CheckpointManager

        return CheckpointManager(mock_repo, project)

    def test_restore_creates_safety_checkpoint_first(self, manager):
        """Should create safety checkpoint before restoring."""
        with patch.object(manager.branch_manager, "create_checkpoint_branch", return_value="safety-001"):
            manager.restore_checkpoint("checkpoint-100")

            # Should have created safety checkpoint
            log_content = manager.log_file.read_text()
            assert "safety-001" in log_content
            assert "restore" in log_content
            assert "is_safety: True" in log_content

    def test_restore_checkouts_target_checkpoint(self, manager):
        """Should checkout the target checkpoint branch."""
        with patch.object(manager.branch_manager, "create_checkpoint_branch", return_value="safety-002"):
            manager.restore_checkpoint("checkpoint-target")

            # Should have checked out the target branch
            manager.repo.git.checkout.assert_called_once_with("checkpoint-target")


class TestListCheckpoints:
    """Test list_checkpoints method."""

    @pytest.fixture
    def manager(self, tmp_path):
        """Create a CheckpointManager for testing."""
        mock_repo = MagicMock(spec=Repo)
        project = tmp_path / "test_project"
        project.mkdir()
        (project / ".moai").mkdir()

        from moai_adk.core.git.checkpoint import CheckpointManager

        return CheckpointManager(mock_repo, project)

    def test_delegates_to_branch_manager(self, manager):
        """Should delegate listing to branch manager."""
        expected = ["checkpoint-001", "checkpoint-002"]
        with patch.object(manager.branch_manager, "list_checkpoint_branches", return_value=expected):
            result = manager.list_checkpoints()

            assert result == expected


class TestLogCheckpoint:
    """Test _log_checkpoint private method."""

    @pytest.fixture
    def manager(self, tmp_path):
        """Create a CheckpointManager for testing."""
        mock_repo = MagicMock(spec=Repo)
        project = tmp_path / "test_project"
        project.mkdir()
        (project / ".moai").mkdir()

        from moai_adk.core.git.checkpoint import CheckpointManager

        return CheckpointManager(mock_repo, project)

    def test_writes_log_entry_to_file(self, manager):
        """Should write log entry with correct format."""
        manager._log_checkpoint("ckpt-001", "test_operation", is_safety=False)

        log_content = manager.log_file.read_text()

        assert "ckpt-001" in log_content
        assert "test_operation" in log_content
        assert "is_safety: False" in log_content
        assert "---" in log_content

    def test_appends_to_existing_log(self, manager):
        """Should append new entries without overwriting."""
        manager._log_checkpoint("ckpt-001", "op1")

        # Write another entry
        manager._log_checkpoint("ckpt-002", "op2")

        log_content = manager.log_file.read_text()

        # Both entries should be present
        assert "ckpt-001" in log_content
        assert "ckpt-002" in log_content
        assert "op1" in log_content
        assert "op2" in log_content
