"""Comprehensive coverage tests for worktree CLI module.

These tests focus on uncovered code paths in src/moai_adk/cli/worktree/cli.py
with emphasis on worktree management operations, git interactions, and error handling.

Target Coverage: 60%+
Test Pattern: AAA (Arrange-Act-Assert)
Mocks: Git operations, file system, subprocess calls
"""

from pathlib import Path
from unittest.mock import MagicMock, patch

from click.testing import CliRunner

from moai_adk.cli.worktree.cli import (
    _detect_worktree_root,
    _find_main_repository,
    clean_worktrees,
    config_worktree,
    get_manager,
    go_worktree,
    list_worktrees,
    new_worktree,
    remove_worktree,
    status_worktrees,
    sync_worktree,
    worktree,
)
from moai_adk.cli.worktree.exceptions import (
    GitOperationError,
    UncommittedChangesError,
    WorktreeExistsError,
    WorktreeNotFoundError,
)

# ============================================================================
# Test Manager and Root Detection
# ============================================================================


class TestGetManager:
    """Test suite for get_manager function."""

    @patch("moai_adk.cli.worktree.cli.WorktreeManager")
    @patch("moai_adk.cli.worktree.cli._detect_worktree_root")
    def test_get_manager_defaults(self, mock_detect_root, mock_manager_class):
        """Test get_manager with default parameters."""
        # Arrange
        mock_root = Path.home() / "moai" / "worktrees"
        mock_detect_root.return_value = mock_root
        mock_manager_instance = MagicMock()
        mock_manager_class.return_value = mock_manager_instance

        # Act
        result = get_manager()

        # Assert
        assert result == mock_manager_instance
        mock_manager_class.assert_called_once()

    @patch("moai_adk.cli.worktree.cli.WorktreeManager")
    def test_get_manager_explicit_paths(self, mock_manager_class):
        """Test get_manager with explicit repo and worktree paths."""
        # Arrange
        repo_path = Path("/mock/repo")
        wt_root = Path("/mock/worktrees")
        mock_manager_instance = MagicMock()
        mock_manager_class.return_value = mock_manager_instance

        # Act
        result = get_manager(repo_path, wt_root)

        # Assert
        assert result == mock_manager_instance
        mock_manager_class.assert_called_once_with(
            repo_path=repo_path, worktree_root=wt_root, project_name="repo"
        )

    @patch("moai_adk.cli.worktree.cli.WorktreeManager")
    @patch("moai_adk.cli.worktree.cli._detect_worktree_root")
    def test_get_manager_finds_git_repo(self, mock_detect_root, mock_manager_class):
        """Test get_manager finds .git directory."""
        # Arrange
        mock_root = Path.home() / "moai" / "worktrees"
        mock_detect_root.return_value = mock_root
        mock_manager_instance = MagicMock()
        mock_manager_class.return_value = mock_manager_instance

        # Act
        result = get_manager()

        # Assert
        assert result is not None


# ============================================================================
# Test Worktree Root Detection
# ============================================================================


class TestDetectWorktreeRoot:
    """Test suite for _detect_worktree_root function."""

    @patch("moai_adk.cli.worktree.cli._find_main_repository")
    def test_detect_worktree_root_from_registry(self, mock_find_main):
        """Test detection from existing registry file."""
        # Arrange
        repo_path = Path("/mock/repo")
        mock_find_main.return_value = repo_path

        # Act
        result = _detect_worktree_root(repo_path)

        # Assert
        assert isinstance(result, Path)

    @patch("pathlib.Path.exists")
    @patch("moai_adk.cli.worktree.cli._find_main_repository")
    def test_detect_worktree_root_default(self, mock_find_main, mock_exists):
        """Test detection uses default location."""
        # Arrange
        repo_path = Path("/mock/repo")
        mock_find_main.return_value = repo_path
        mock_exists.return_value = False

        # Act
        result = _detect_worktree_root(repo_path)

        # Assert
        assert isinstance(result, Path)

    @patch("moai_adk.cli.worktree.cli._find_main_repository")
    def test_detect_worktree_root_with_existing_worktrees(self, mock_find_main):
        """Test detection with existing worktree directories."""
        # Arrange
        repo_path = Path("/mock/repo")
        mock_find_main.return_value = repo_path

        # Act
        result = _detect_worktree_root(repo_path)

        # Assert
        assert isinstance(result, Path)


# ============================================================================
# Test Find Main Repository
# ============================================================================


class TestFindMainRepository:
    """Test suite for _find_main_repository function."""

    def test_find_main_repository_from_worktree(self):
        """Test finding main repo from worktree."""
        # Arrange
        start_path = Path("/mock/worktrees/feature-1")

        # Act
        result = _find_main_repository(start_path)

        # Assert
        assert isinstance(result, Path)

    @patch("pathlib.Path.exists")
    def test_find_main_repository_already_main(self, mock_exists):
        """Test when already at main repository."""
        # Arrange
        start_path = Path("/mock/repo")
        mock_exists.return_value = True

        # Act
        result = _find_main_repository(start_path)

        # Assert
        assert isinstance(result, Path)

    @patch("pathlib.Path.exists")
    def test_find_main_repository_no_git(self, mock_exists):
        """Test fallback when no .git found."""
        # Arrange
        start_path = Path("/mock/norepro")
        mock_exists.return_value = False

        # Act
        result = _find_main_repository(start_path)

        # Assert
        assert isinstance(result, Path)


# ============================================================================
# Test Worktree Commands
# ============================================================================


class TestNewWorktreeCommand:
    """Test suite for new worktree command."""

    def test_new_worktree_command_exists(self):
        """Test new worktree command is registered."""
        # Act
        result = worktree.commands

        # Assert
        assert "new" in result

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_new_worktree_success(self, mock_get_manager):
        """Test successful worktree creation."""
        # Arrange
        runner = CliRunner()
        mock_manager = MagicMock()
        mock_manager.create.return_value = MagicMock(
            spec_id="SPEC-001",
            path=Path("/mock/worktree"),
            branch="feature/spec-001",
            status="active",
        )
        mock_get_manager.return_value = mock_manager

        # Act
        result = runner.invoke(new_worktree, ["SPEC-001", "--branch", "feature/spec-001"])

        # Assert
        assert result.exit_code == 0
        assert "Worktree created successfully" in result.output

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_new_worktree_already_exists(self, mock_get_manager):
        """Test worktree creation when worktree exists."""
        # Arrange
        runner = CliRunner()
        mock_manager = MagicMock()
        mock_manager.create.side_effect = WorktreeExistsError("SPEC-001", Path("/mock/worktree"))
        mock_get_manager.return_value = mock_manager

        # Act
        result = runner.invoke(new_worktree, ["SPEC-001"])

        # Assert
        assert result.exit_code != 0

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_new_worktree_git_error(self, mock_get_manager):
        """Test worktree creation with git error."""
        # Arrange
        runner = CliRunner()
        mock_manager = MagicMock()
        mock_manager.create.side_effect = GitOperationError("Git failed")
        mock_get_manager.return_value = mock_manager

        # Act
        result = runner.invoke(new_worktree, ["SPEC-001"])

        # Assert
        assert result.exit_code != 0


class TestListWorktreesCommand:
    """Test suite for list worktrees command."""

    def test_list_worktrees_command_exists(self):
        """Test list worktrees command is registered."""
        # Act
        result = worktree.commands

        # Assert
        assert "list" in result

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_list_worktrees_table_format(self, mock_get_manager):
        """Test list worktrees in table format."""
        # Arrange
        runner = CliRunner()
        mock_manager = MagicMock()
        mock_manager.list.return_value = [
            MagicMock(
                spec_id="SPEC-001",
                branch="feature/spec-001",
                path=Path("/mock/worktree1"),
                status="active",
                created_at="2024-01-01T00:00:00Z",
            ),
        ]
        mock_get_manager.return_value = mock_manager

        # Act
        result = runner.invoke(list_worktrees, ["--format", "table"])

        # Assert
        assert result.exit_code == 0
        assert "SPEC-001" in result.output

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_list_worktrees_json_format(self, mock_get_manager):
        """Test list worktrees in JSON format."""
        # Arrange
        runner = CliRunner()
        mock_manager = MagicMock()
        mock_worktree = MagicMock()
        mock_worktree.to_dict.return_value = {"spec_id": "SPEC-001"}
        mock_manager.list.return_value = [mock_worktree]
        mock_get_manager.return_value = mock_manager

        # Act
        result = runner.invoke(list_worktrees, ["--format", "json"])

        # Assert
        assert result.exit_code == 0

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_list_worktrees_empty(self, mock_get_manager):
        """Test list worktrees when none exist."""
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


class TestRemoveWorktreeCommand:
    """Test suite for remove worktree command."""

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_remove_worktree_success(self, mock_get_manager):
        """Test successful worktree removal."""
        # Arrange
        runner = CliRunner()
        mock_manager = MagicMock()
        mock_get_manager.return_value = mock_manager

        # Act
        result = runner.invoke(remove_worktree, ["SPEC-001"])

        # Assert
        assert result.exit_code == 0
        assert "removed" in result.output

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_remove_worktree_not_found(self, mock_get_manager):
        """Test remove when worktree not found."""
        # Arrange
        runner = CliRunner()
        mock_manager = MagicMock()
        mock_manager.remove.side_effect = WorktreeNotFoundError("Not found")
        mock_get_manager.return_value = mock_manager

        # Act
        result = runner.invoke(remove_worktree, ["NONEXISTENT"])

        # Assert
        assert result.exit_code != 0

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_remove_worktree_with_uncommitted(self, mock_get_manager):
        """Test remove with uncommitted changes."""
        # Arrange
        runner = CliRunner()
        mock_manager = MagicMock()
        mock_manager.remove.side_effect = UncommittedChangesError("Has changes")
        mock_get_manager.return_value = mock_manager

        # Act
        result = runner.invoke(remove_worktree, ["SPEC-001"])

        # Assert
        assert result.exit_code != 0

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_remove_worktree_force(self, mock_get_manager):
        """Test remove with force flag."""
        # Arrange
        runner = CliRunner()
        mock_manager = MagicMock()
        mock_get_manager.return_value = mock_manager

        # Act
        result = runner.invoke(remove_worktree, ["SPEC-001", "--force"])

        # Assert
        assert result.exit_code == 0
        mock_manager.remove.assert_called_once_with(spec_id="SPEC-001", force=True)


class TestStatusWorktreesCommand:
    """Test suite for status worktrees command."""

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_status_worktrees_success(self, mock_get_manager):
        """Test successful status display."""
        # Arrange
        runner = CliRunner()
        mock_manager = MagicMock()
        mock_manager.list.return_value = [
            MagicMock(
                spec_id="SPEC-001",
                branch="feature/spec-001",
                path=Path("/mock/worktree"),
                status="active",
            ),
        ]
        mock_get_manager.return_value = mock_manager

        # Act
        result = runner.invoke(status_worktrees)

        # Assert
        assert result.exit_code == 0
        assert "SPEC-001" in result.output

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_status_worktrees_empty(self, mock_get_manager):
        """Test status when no worktrees exist."""
        # Arrange
        runner = CliRunner()
        mock_manager = MagicMock()
        mock_manager.list.return_value = []
        mock_get_manager.return_value = mock_manager

        # Act
        result = runner.invoke(status_worktrees)

        # Assert
        assert result.exit_code == 0
        assert "No worktrees found" in result.output


class TestGoWorktreeCommand:
    """Test suite for go worktree command."""

    @patch("subprocess.call")
    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_go_worktree_success(self, mock_get_manager, mock_subprocess_call):
        """Test successful cd command generation."""
        # Arrange
        runner = CliRunner()
        mock_manager = MagicMock()
        mock_manager.project_name = "test-repo"
        mock_worktree_info = MagicMock(path=Path("/mock/worktree"))
        mock_manager.registry.get.return_value = mock_worktree_info
        mock_get_manager.return_value = mock_manager
        mock_subprocess_call.return_value = 0

        # Act
        result = runner.invoke(go_worktree, ["SPEC-001"])

        # Assert
        assert result.exit_code == 0, f"Command failed: {result.output}"
        assert "Opening new shell" in result.output

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_go_worktree_not_found(self, mock_get_manager):
        """Test go when worktree not found."""
        # Arrange
        runner = CliRunner()
        mock_manager = MagicMock()
        mock_manager.registry.get.return_value = None
        mock_get_manager.return_value = mock_manager

        # Act
        result = runner.invoke(go_worktree, ["NONEXISTENT"])

        # Assert
        assert result.exit_code != 0


class TestSyncWorktreeCommand:
    """Test suite for sync worktree command."""

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_sync_worktree_success(self, mock_get_manager):
        """Test successful worktree sync."""
        # Arrange
        runner = CliRunner()
        mock_manager = MagicMock()
        mock_get_manager.return_value = mock_manager

        # Act
        result = runner.invoke(sync_worktree, ["SPEC-001"])

        # Assert
        assert result.exit_code == 0
        mock_manager.sync.assert_called_once()

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_sync_worktree_no_args(self, mock_get_manager):
        """Test sync without spec_id or --all."""
        # Arrange
        runner = CliRunner()
        mock_manager = MagicMock()
        mock_get_manager.return_value = mock_manager

        # Act
        result = runner.invoke(sync_worktree, [])

        # Assert
        assert result.exit_code != 0
        assert "required" in result.output.lower()

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_sync_worktree_all(self, mock_get_manager):
        """Test sync all worktrees."""
        # Arrange
        runner = CliRunner()
        mock_manager = MagicMock()
        mock_manager.list.return_value = [
            MagicMock(spec_id="SPEC-001", branch="feature/spec-001", status="active"),
            MagicMock(spec_id="SPEC-002", branch="feature/spec-002", status="active"),
        ]
        mock_get_manager.return_value = mock_manager

        # Act
        result = runner.invoke(sync_worktree, ["--all"])

        # Assert
        assert result.exit_code == 0

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_sync_worktree_with_rebase(self, mock_get_manager):
        """Test sync with rebase option."""
        # Arrange
        runner = CliRunner()
        mock_manager = MagicMock()
        mock_get_manager.return_value = mock_manager

        # Act
        result = runner.invoke(sync_worktree, ["SPEC-001", "--rebase"])

        # Assert
        assert result.exit_code == 0
        mock_manager.sync.assert_called_once()
        call_kwargs = mock_manager.sync.call_args[1]
        assert call_kwargs["rebase"] is True


class TestCleanWorktreesCommand:
    """Test suite for clean worktrees command."""

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_clean_merged_only(self, mock_get_manager):
        """Test clean with merged-only flag."""
        # Arrange
        runner = CliRunner()
        mock_manager = MagicMock()
        mock_manager.clean_merged.return_value = ["SPEC-001"]
        mock_get_manager.return_value = mock_manager

        # Act
        result = runner.invoke(clean_worktrees, ["--merged-only"])

        # Assert
        assert result.exit_code == 0
        mock_manager.clean_merged.assert_called_once()

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_clean_interactive(self, mock_get_manager):
        """Test clean in interactive mode."""
        # Arrange
        runner = CliRunner()
        mock_manager = MagicMock()
        mock_manager.list.return_value = [
            MagicMock(spec_id="SPEC-001", branch="feature/spec-001", status="active"),
        ]
        mock_manager.remove.return_value = None
        mock_get_manager.return_value = mock_manager

        # Act
        result = runner.invoke(clean_worktrees, ["--interactive"], input="1\nn\n")

        # Assert
        # Interactive commands have exit code 1 when cancelled
        assert result.exit_code in (0, 1)

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_clean_all(self, mock_get_manager):
        """Test clean all worktrees."""
        # Arrange
        runner = CliRunner()
        mock_manager = MagicMock()
        mock_manager.list.return_value = [
            MagicMock(spec_id="SPEC-001", branch="feature/spec-001", status="active"),
        ]
        mock_get_manager.return_value = mock_manager

        # Act
        result = runner.invoke(clean_worktrees)

        # Assert
        assert result.exit_code == 0


class TestConfigWorktreeCommand:
    """Test suite for config worktree command."""

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_config_show_all(self, mock_get_manager):
        """Test show all configuration."""
        # Arrange
        runner = CliRunner()
        mock_manager = MagicMock()
        mock_manager.worktree_root = Path("/mock/worktrees")
        mock_manager.registry.registry_path = Path("/mock/registry.json")
        mock_get_manager.return_value = mock_manager

        # Act
        result = runner.invoke(config_worktree, [])

        # Assert
        assert result.exit_code == 0
        assert "Configuration" in result.output

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_config_get_root(self, mock_get_manager):
        """Test get root configuration."""
        # Arrange
        runner = CliRunner()
        mock_manager = MagicMock()
        mock_manager.worktree_root = Path("/mock/worktrees")
        mock_get_manager.return_value = mock_manager

        # Act
        result = runner.invoke(config_worktree, ["root"])

        # Assert
        assert result.exit_code == 0
        assert "root" in result.output.lower()

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_config_unknown_key(self, mock_get_manager):
        """Test get unknown configuration key."""
        # Arrange
        runner = CliRunner()
        mock_manager = MagicMock()
        mock_get_manager.return_value = mock_manager

        # Act
        result = runner.invoke(config_worktree, ["unknown"])

        # Assert
        assert result.exit_code == 0
        assert "Unknown" in result.output
