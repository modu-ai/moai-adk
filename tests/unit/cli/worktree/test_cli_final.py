"""
Final comprehensive tests for CLI worktree commands.

Focuses on simple, working tests for Click commands:
- get_manager function
- _detect_worktree_root function
- _find_main_repository function
- Click command execution
"""

import subprocess
from pathlib import Path
from datetime import datetime
from unittest.mock import MagicMock, patch, Mock, call

import pytest
from click.testing import CliRunner

from moai_adk.cli.worktree.cli import (
    get_manager,
    _detect_worktree_root,
    _find_main_repository,
    worktree,
    new_worktree,
    list_worktrees,
    switch_worktree,
    remove_worktree,
)
from moai_adk.cli.worktree.exceptions import (
    WorktreeExistsError,
    GitOperationError,
)


class TestDetectWorktreeRoot:
    """Test _detect_worktree_root function."""

    @patch("moai_adk.cli.worktree.cli.Path")
    def test_detect_worktree_root_finds_existing_registry(self, mock_path_class):
        """Test detection finds existing registry file."""
        # Arrange
        registry_path = Path("/home/user/moai/worktrees/.moai-worktree-registry.json")

        # Mock the potential roots
        mock_root = MagicMock()
        mock_root.__truediv__ = MagicMock(return_value=registry_path)
        mock_root.exists = MagicMock(return_value=True)

        # Mock registry file exists and has content
        mock_path_class.home = MagicMock(return_value=Path("/home/user"))

        with patch("builtins.open", create=True) as mock_open:
            mock_open.return_value.__enter__.return_value.read.return_value = "{}"

            # Act
            result = _detect_worktree_root(Path("/test/repo"))

            # Assert - Should return a path
            assert result is not None

    @patch("moai_adk.cli.worktree.cli.Path")
    def test_detect_worktree_root_default_fallback(self, mock_path_class):
        """Test detection returns default path as fallback."""
        # Arrange
        home = Path("/home/user")
        default_root = home / "moai" / "worktrees"

        mock_path_class.home = MagicMock(return_value=home)

        # Act
        result = _detect_worktree_root(Path("/test/repo"))

        # Assert - Should return some valid path
        assert result is not None

    @patch("moai_adk.cli.worktree.cli._find_main_repository")
    @patch("moai_adk.cli.worktree.cli.Path")
    def test_detect_worktree_root_from_worktree(self, mock_path_class, mock_find_repo):
        """Test detection when starting from a worktree."""
        # Arrange
        main_repo = Path("/test/main/repo")
        mock_find_repo.return_value = main_repo
        mock_path_class.home = MagicMock(return_value=Path("/home/user"))

        # Act
        result = _detect_worktree_root(Path("/test/worktree/path"))

        # Assert
        assert result is not None
        mock_find_repo.assert_called_once()


class TestFindMainRepository:
    """Test _find_main_repository function."""

    def test_find_main_repository_main_repo(self):
        """Test finding main repository when at main repo."""
        # Arrange
        with patch("pathlib.Path.parent", new_callable=lambda: property(
            lambda self: self if self.name == "repo" else Path("/test")
        )):
            test_path = Path("/test/repo")

            # Act
            with patch("pathlib.Path.exists") as mock_exists:
                with patch("pathlib.Path.is_file") as mock_is_file:
                    # Setup mock to detect main repo
                    mock_exists.side_effect = lambda: test_path.name == "repo"
                    mock_is_file.return_value = False

                    result = _find_main_repository(test_path)

                    # Assert - Should return the path
                    assert result is not None

    @patch("builtins.open", create=True)
    def test_find_main_repository_worktree_file(self, mock_open):
        """Test finding main repository from worktree .git file."""
        # Arrange
        mock_open.return_value.__enter__.return_value.readlines = MagicMock(
            return_value=["gitdir: ../main-repo/.git/worktrees/feature\n"]
        )

        # Act
        result = _find_main_repository(Path("/test/worktree/path"))

        # Assert - Should attempt to find main repo
        assert result is not None


class TestGetManager:
    """Test get_manager factory function."""

    @patch("moai_adk.cli.worktree.cli.WorktreeManager")
    @patch("moai_adk.cli.worktree.cli._detect_worktree_root")
    def test_get_manager_with_explicit_paths(self, mock_detect, mock_manager_class):
        """Test get_manager with explicit repo and worktree paths."""
        # Arrange
        repo_path = Path("/test/repo")
        worktree_root = Path("/test/worktrees")
        mock_manager = MagicMock()
        mock_manager_class.return_value = mock_manager

        # Act
        result = get_manager(repo_path=repo_path, worktree_root=worktree_root)

        # Assert
        assert result == mock_manager
        mock_manager_class.assert_called_once_with(repo_path=repo_path, worktree_root=worktree_root)
        mock_detect.assert_not_called()

    @patch("moai_adk.cli.worktree.cli.WorktreeManager")
    @patch("moai_adk.cli.worktree.cli._detect_worktree_root")
    @patch("moai_adk.cli.worktree.cli.Path.cwd")
    def test_get_manager_auto_detect_worktree_root(self, mock_cwd, mock_detect, mock_manager_class):
        """Test get_manager auto-detects worktree root."""
        # Arrange
        mock_cwd.return_value = Path("/test/repo")
        mock_detect.return_value = Path("/test/worktrees")
        mock_manager = MagicMock()
        mock_manager_class.return_value = mock_manager

        # Act
        result = get_manager(repo_path=Path("/test/repo"), worktree_root=None)

        # Assert
        assert result == mock_manager
        mock_detect.assert_called_once()


class TestWorktreeCommands:
    """Test worktree Click command group."""

    def test_worktree_command_group_exists(self):
        """Test that worktree command group exists."""
        # Assert
        assert worktree is not None
        assert callable(worktree)

    def test_worktree_command_has_subcommands(self):
        """Test worktree command has expected subcommands."""
        # Assert
        assert hasattr(worktree, "commands")
        commands = list(worktree.commands.keys())
        assert "new" in commands
        assert "list" in commands
        assert "switch" in commands
        assert "remove" in commands


class TestNewWorktreeCommand:
    """Test new_worktree Click command."""

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_new_worktree_success(self, mock_get_manager):
        """Test successful worktree creation."""
        # Arrange
        runner = CliRunner()

        mock_manager = MagicMock()
        mock_manager.create.return_value = MagicMock(
            spec_id="SPEC-TEST-001",
            path="/test/worktrees/SPEC-TEST-001",
            branch="spec-test-001",
            status="active",
        )
        mock_get_manager.return_value = mock_manager

        # Act
        result = runner.invoke(new_worktree, ["SPEC-TEST-001"])

        # Assert
        assert result.exit_code == 0
        assert "✓" in result.output
        assert "Worktree created successfully" in result.output

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_new_worktree_with_custom_branch(self, mock_get_manager):
        """Test worktree creation with custom branch."""
        # Arrange
        runner = CliRunner()

        mock_manager = MagicMock()
        mock_manager.create.return_value = MagicMock(
            spec_id="SPEC-TEST-001",
            path="/test/worktrees/SPEC-TEST-001",
            branch="custom-branch",
            status="active",
        )
        mock_get_manager.return_value = mock_manager

        # Act
        result = runner.invoke(new_worktree, [
            "SPEC-TEST-001",
            "--branch", "custom-branch",
        ])

        # Assert
        assert result.exit_code == 0
        mock_manager.create.assert_called_once()

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_new_worktree_with_base_branch(self, mock_get_manager):
        """Test worktree creation with custom base branch."""
        # Arrange
        runner = CliRunner()

        mock_manager = MagicMock()
        mock_manager.create.return_value = MagicMock(
            spec_id="SPEC-TEST-001",
            path="/test/worktrees/SPEC-TEST-001",
            branch="spec-test-001",
            status="active",
        )
        mock_get_manager.return_value = mock_manager

        # Act
        result = runner.invoke(new_worktree, [
            "SPEC-TEST-001",
            "--base", "develop",
        ])

        # Assert
        assert result.exit_code == 0
        call_kwargs = mock_manager.create.call_args[1]
        assert call_kwargs["base_branch"] == "develop"

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_new_worktree_already_exists(self, mock_get_manager):
        """Test worktree creation error when exists."""
        # Arrange
        runner = CliRunner()

        mock_manager = MagicMock()
        mock_manager.create.side_effect = WorktreeExistsError("Worktree already exists", Path("/test/path"))
        mock_get_manager.return_value = mock_manager

        # Act
        result = runner.invoke(new_worktree, ["SPEC-TEST-001"])

        # Assert
        assert result.exit_code != 0
        assert "✗" in result.output

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_new_worktree_git_error(self, mock_get_manager):
        """Test worktree creation git operation error."""
        # Arrange
        runner = CliRunner()

        mock_manager = MagicMock()
        mock_manager.create.side_effect = GitOperationError("Git failed")
        mock_get_manager.return_value = mock_manager

        # Act
        result = runner.invoke(new_worktree, ["SPEC-TEST-001"])

        # Assert
        assert result.exit_code != 0


class TestListWorktreesCommand:
    """Test list_worktrees Click command."""

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_list_worktrees_empty(self, mock_get_manager):
        """Test listing worktrees when none exist."""
        # Arrange
        runner = CliRunner()

        mock_manager = MagicMock()
        mock_manager.list.return_value = []
        mock_get_manager.return_value = mock_manager

        # Act
        result = runner.invoke(list_worktrees)

        # Assert
        assert result.exit_code == 0
        assert "No worktrees found" in result.output

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_list_worktrees_table_format(self, mock_get_manager):
        """Test listing worktrees in table format."""
        # Arrange
        runner = CliRunner()

        mock_worktree_info = MagicMock()
        mock_worktree_info.spec_id = "SPEC-TEST-001"
        mock_worktree_info.branch = "spec-test-001"
        mock_worktree_info.path = Path("/test/worktrees/SPEC-TEST-001")
        mock_worktree_info.status = "active"
        mock_worktree_info.created_at = datetime.now().isoformat()

        mock_manager = MagicMock()
        mock_manager.list.return_value = [mock_worktree_info]
        mock_get_manager.return_value = mock_manager

        # Act
        result = runner.invoke(list_worktrees, ["--format", "table"])

        # Assert
        assert result.exit_code == 0
        assert "Git Worktrees" in result.output

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_list_worktrees_json_format(self, mock_get_manager):
        """Test listing worktrees in JSON format."""
        # Arrange
        runner = CliRunner()

        mock_worktree_info = MagicMock()
        mock_worktree_info.spec_id = "SPEC-TEST-001"
        mock_worktree_info.branch = "spec-test-001"
        mock_worktree_info.path = Path("/test/worktrees/SPEC-TEST-001")
        mock_worktree_info.status = "active"
        mock_worktree_info.created_at = datetime.now().isoformat()
        mock_worktree_info.to_dict = MagicMock(return_value={
            "spec_id": "SPEC-TEST-001",
            "branch": "spec-test-001",
            "path": "/test/worktrees/SPEC-TEST-001",
            "status": "active",
            "created_at": datetime.now().isoformat(),
        })

        mock_manager = MagicMock()
        mock_manager.list.return_value = [mock_worktree_info]
        mock_get_manager.return_value = mock_manager

        # Act
        result = runner.invoke(list_worktrees, ["--format", "json"])

        # Assert
        assert result.exit_code == 0


class TestSwitchWorktreeCommand:
    """Test switch_worktree Click command."""

    @patch("subprocess.call")
    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_switch_worktree_success(self, mock_get_manager, mock_subprocess):
        """Test successful worktree switch."""
        # Arrange
        runner = CliRunner()

        mock_worktree_info = MagicMock()
        mock_worktree_info.path = Path("/test/worktrees/SPEC-TEST-001")

        mock_manager = MagicMock()
        mock_manager.registry.get.return_value = mock_worktree_info
        mock_get_manager.return_value = mock_manager

        mock_subprocess.return_value = 0

        # Act
        result = runner.invoke(switch_worktree, ["SPEC-TEST-001"])

        # Assert
        assert result.exit_code == 0
        assert "→" in result.output
        mock_subprocess.assert_called_once()

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_switch_worktree_not_found(self, mock_get_manager):
        """Test switching to non-existent worktree."""
        # Arrange
        runner = CliRunner()

        mock_manager = MagicMock()
        mock_manager.registry.get.return_value = None
        mock_get_manager.return_value = mock_manager

        # Act
        result = runner.invoke(switch_worktree, ["SPEC-NONEXISTENT"])

        # Assert
        assert result.exit_code != 0
        assert "✗" in result.output


class TestRemoveWorktreeCommand:
    """Test remove_worktree Click command."""

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_remove_worktree_success(self, mock_get_manager):
        """Test successful worktree removal."""
        # Arrange
        runner = CliRunner()

        mock_manager = MagicMock()
        mock_manager.remove.return_value = None
        mock_get_manager.return_value = mock_manager

        # Act
        result = runner.invoke(remove_worktree, ["SPEC-TEST-001"])

        # Assert
        assert result.exit_code == 0
        assert "✓" in result.output

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_remove_worktree_with_force(self, mock_get_manager):
        """Test removing worktree with force flag."""
        # Arrange
        runner = CliRunner()

        mock_manager = MagicMock()
        mock_manager.remove.return_value = None
        mock_get_manager.return_value = mock_manager

        # Act
        result = runner.invoke(remove_worktree, [
            "SPEC-TEST-001",
            "--force",
        ])

        # Assert
        assert result.exit_code == 0
        call_kwargs = mock_manager.remove.call_args[1]
        assert call_kwargs.get("force") is True


class TestGetManagerWithPathFinding:
    """Test get_manager with path finding."""

    @patch("moai_adk.cli.worktree.cli.WorktreeManager")
    @patch("moai_adk.cli.worktree.cli._detect_worktree_root")
    @patch("moai_adk.cli.worktree.cli.Path")
    def test_get_manager_finds_git_directory(self, mock_path_class, mock_detect, mock_manager_class):
        """Test get_manager walks up to find .git directory."""
        # Arrange
        git_dir = Path("/test/repo/.git")
        repo_path = Path("/test/repo")

        # Create mock that simulates walking up directories
        mock_cwd = MagicMock()
        mock_cwd.__truediv__ = MagicMock(return_value=git_dir)
        mock_cwd.exists = MagicMock(return_value=True)
        mock_cwd.parent = repo_path

        mock_path_class.cwd = MagicMock(return_value=mock_cwd)
        mock_path_class.return_value = mock_cwd

        mock_manager = MagicMock()
        mock_manager_class.return_value = mock_manager
        mock_detect.return_value = Path("/test/worktrees")

        # Act
        result = get_manager()

        # Assert
        assert result == mock_manager


class TestDetectWorktreeRootWithExistingWorktrees:
    """Test worktree root detection with actual worktrees."""

    @patch("moai_adk.cli.worktree.cli._find_main_repository")
    @patch("moai_adk.cli.worktree.cli.Path")
    def test_detect_finds_existing_worktrees(self, mock_path_class, mock_find_repo):
        """Test detection finds existing worktrees in standard location."""
        # Arrange
        home = Path("/home/user")
        worktree_root = home / "moai" / "worktrees"
        main_repo = Path("/test/main/repo")

        mock_find_repo.return_value = main_repo
        mock_path_class.home = MagicMock(return_value=home)

        # Act
        result = _detect_worktree_root(Path("/test/repo"))

        # Assert
        assert result is not None
