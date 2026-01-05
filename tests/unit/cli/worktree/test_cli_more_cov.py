"""Additional comprehensive tests for CLI commands with 60%+ coverage target.

Tests all uncovered CLI command paths with proper mocking of WorktreeManager
and Git operations. Uses AAA pattern and @patch decorators for dependencies.
"""

from pathlib import Path
from unittest.mock import MagicMock, patch

from click.testing import CliRunner

from moai_adk.cli.worktree.cli import (
    worktree,
)
from moai_adk.cli.worktree.exceptions import (
    GitOperationError,
    MergeConflictError,
    UncommittedChangesError,
    WorktreeExistsError,
    WorktreeNotFoundError,
)
from moai_adk.cli.worktree.models import WorktreeInfo


class TestNewWorktreeCommand:
    """Test new worktree command."""

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_new_worktree_success(self, mock_get_manager):
        """Test successful worktree creation."""
        # Arrange
        runner = CliRunner()
        mock_manager = MagicMock()
        mock_get_manager.return_value = mock_manager

        now = "2025-01-01T12:00:00Z"
        worktree_info = WorktreeInfo(
            spec_id="SPEC-001",
            path=Path("/tmp/SPEC-001"),
            branch="feature/SPEC-001",
            created_at=now,
            last_accessed=now,
            status="active",
        )
        mock_manager.create.return_value = worktree_info

        # Act
        result = runner.invoke(worktree, ["new", "SPEC-001"])

        # Assert
        assert result.exit_code == 0
        assert "Worktree created successfully" in result.output
        assert "SPEC-001" in result.output
        mock_manager.create.assert_called_once()

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_new_worktree_with_custom_branch(self, mock_get_manager):
        """Test worktree creation with custom branch name."""
        # Arrange
        runner = CliRunner()
        mock_manager = MagicMock()
        mock_get_manager.return_value = mock_manager

        now = "2025-01-01T12:00:00Z"
        worktree_info = WorktreeInfo(
            spec_id="SPEC-002",
            path=Path("/tmp/SPEC-002"),
            branch="custom-branch",
            created_at=now,
            last_accessed=now,
            status="active",
        )
        mock_manager.create.return_value = worktree_info

        # Act
        result = runner.invoke(worktree, ["new", "SPEC-002", "--branch", "custom-branch"])

        # Assert
        assert result.exit_code == 0
        mock_manager.create.assert_called_once()
        call_kwargs = mock_manager.create.call_args[1]
        assert call_kwargs["branch_name"] == "custom-branch"

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_new_worktree_exists_error(self, mock_get_manager):
        """Test worktree creation when worktree already exists."""
        # Arrange
        runner = CliRunner()
        mock_manager = MagicMock()
        mock_get_manager.return_value = mock_manager
        mock_manager.create.side_effect = WorktreeExistsError("SPEC-001", Path("/tmp/SPEC-001"))

        # Act
        result = runner.invoke(worktree, ["new", "SPEC-001"])

        # Assert
        assert result.exit_code != 0
        assert "SPEC-001" in result.output

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_new_worktree_git_error(self, mock_get_manager):
        """Test worktree creation with Git operation error."""
        # Arrange
        runner = CliRunner()
        mock_manager = MagicMock()
        mock_get_manager.return_value = mock_manager
        mock_manager.create.side_effect = GitOperationError("Git operation failed")

        # Act
        result = runner.invoke(worktree, ["new", "SPEC-001"])

        # Assert
        assert result.exit_code != 0

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_new_worktree_with_force_flag(self, mock_get_manager):
        """Test worktree creation with force flag."""
        # Arrange
        runner = CliRunner()
        mock_manager = MagicMock()
        mock_get_manager.return_value = mock_manager

        now = "2025-01-01T12:00:00Z"
        worktree_info = WorktreeInfo(
            spec_id="SPEC-003",
            path=Path("/tmp/SPEC-003"),
            branch="feature/SPEC-003",
            created_at=now,
            last_accessed=now,
            status="active",
        )
        mock_manager.create.return_value = worktree_info

        # Act
        result = runner.invoke(worktree, ["new", "SPEC-003", "--force"])

        # Assert
        assert result.exit_code == 0
        call_kwargs = mock_manager.create.call_args[1]
        assert call_kwargs["force"] is True


class TestListWorktreesCommand:
    """Test list worktrees command."""

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_list_worktrees_table_format(self, mock_get_manager):
        """Test listing worktrees in table format."""
        # Arrange
        runner = CliRunner()
        mock_manager = MagicMock()
        mock_get_manager.return_value = mock_manager

        now = "2025-01-01T12:00:00Z"
        worktree_info = WorktreeInfo(
            spec_id="SPEC-001",
            path=Path("/tmp/SPEC-001"),
            branch="feature/SPEC-001",
            created_at=now,
            last_accessed=now,
            status="active",
        )
        mock_manager.list.return_value = [worktree_info]

        # Act
        result = runner.invoke(worktree, ["list"])

        # Assert
        assert result.exit_code == 0
        assert "Git Worktrees" in result.output
        assert "SPEC-001" in result.output

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_list_worktrees_json_format(self, mock_get_manager):
        """Test listing worktrees in JSON format."""
        # Arrange
        runner = CliRunner()
        mock_manager = MagicMock()
        mock_get_manager.return_value = mock_manager

        now = "2025-01-01T12:00:00Z"
        worktree_info = WorktreeInfo(
            spec_id="SPEC-001",
            path=Path("/tmp/SPEC-001"),
            branch="feature/SPEC-001",
            created_at=now,
            last_accessed=now,
            status="active",
        )
        mock_manager.list.return_value = [worktree_info]

        # Act
        result = runner.invoke(worktree, ["list", "--format", "json"])

        # Assert
        assert result.exit_code == 0
        assert "SPEC-001" in result.output

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_list_worktrees_empty(self, mock_get_manager):
        """Test listing when no worktrees exist."""
        # Arrange
        runner = CliRunner()
        mock_manager = MagicMock()
        mock_get_manager.return_value = mock_manager
        mock_manager.list.return_value = []

        # Act
        result = runner.invoke(worktree, ["list"])

        # Assert
        assert result.exit_code == 0
        assert "No worktrees found" in result.output


class TestRemoveWorktreeCommand:
    """Test remove worktree command."""

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_remove_worktree_success(self, mock_get_manager):
        """Test successful worktree removal."""
        # Arrange
        runner = CliRunner()
        mock_manager = MagicMock()
        mock_get_manager.return_value = mock_manager

        # Act
        result = runner.invoke(worktree, ["remove", "SPEC-001"])

        # Assert
        assert result.exit_code == 0
        assert "removed" in result.output
        mock_manager.remove.assert_called_once()

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_remove_worktree_not_found(self, mock_get_manager):
        """Test removing non-existent worktree."""
        # Arrange
        runner = CliRunner()
        mock_manager = MagicMock()
        mock_get_manager.return_value = mock_manager
        mock_manager.remove.side_effect = WorktreeNotFoundError("SPEC-001")

        # Act
        result = runner.invoke(worktree, ["remove", "SPEC-001"])

        # Assert
        assert result.exit_code != 0

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_remove_worktree_uncommitted_changes(self, mock_get_manager):
        """Test removing worktree with uncommitted changes."""
        # Arrange
        runner = CliRunner()
        mock_manager = MagicMock()
        mock_get_manager.return_value = mock_manager
        mock_manager.remove.side_effect = UncommittedChangesError("SPEC-001")

        # Act
        result = runner.invoke(worktree, ["remove", "SPEC-001"])

        # Assert
        assert result.exit_code != 0

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_remove_worktree_with_force(self, mock_get_manager):
        """Test removing worktree with force flag."""
        # Arrange
        runner = CliRunner()
        mock_manager = MagicMock()
        mock_get_manager.return_value = mock_manager

        # Act
        result = runner.invoke(worktree, ["remove", "SPEC-001", "--force"])

        # Assert
        assert result.exit_code == 0
        call_kwargs = mock_manager.remove.call_args[1]
        assert call_kwargs["force"] is True


class TestStatusWorktreesCommand:
    """Test status worktrees command."""

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_status_worktrees_success(self, mock_get_manager):
        """Test status command with worktrees."""
        # Arrange
        runner = CliRunner()
        mock_manager = MagicMock()
        mock_get_manager.return_value = mock_manager

        now = "2025-01-01T12:00:00Z"
        worktree_info = WorktreeInfo(
            spec_id="SPEC-001",
            path=Path("/tmp/SPEC-001"),
            branch="feature/SPEC-001",
            created_at=now,
            last_accessed=now,
            status="active",
        )
        mock_manager.list.return_value = [worktree_info]

        # Act
        result = runner.invoke(worktree, ["status"])

        # Assert
        assert result.exit_code == 0
        assert "SPEC-001" in result.output
        mock_manager.registry.sync_with_git.assert_called_once()

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_status_worktrees_empty(self, mock_get_manager):
        """Test status command with no worktrees."""
        # Arrange
        runner = CliRunner()
        mock_manager = MagicMock()
        mock_get_manager.return_value = mock_manager
        mock_manager.list.return_value = []

        # Act
        result = runner.invoke(worktree, ["status"])

        # Assert
        assert result.exit_code == 0
        assert "No worktrees found" in result.output


class TestGoWorktreeCommand:
    """Test go worktree command."""

    @patch("subprocess.call")
    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_go_worktree_success(self, mock_get_manager, mock_subprocess_call):
        """Test go command returns cd command."""
        # Arrange
        runner = CliRunner()
        mock_manager = MagicMock()
        mock_manager.project_name = "test-repo"
        mock_get_manager.return_value = mock_manager
        mock_subprocess_call.return_value = 0

        now = "2025-01-01T12:00:00Z"
        worktree_info = WorktreeInfo(
            spec_id="SPEC-001",
            path=Path("/tmp/SPEC-001"),
            branch="feature/SPEC-001",
            created_at=now,
            last_accessed=now,
            status="active",
        )
        mock_manager.registry.get.return_value = worktree_info

        # Act
        result = runner.invoke(worktree, ["go", "SPEC-001"])

        # Assert
        assert result.exit_code == 0, f"Command failed: {result.output}"
        assert "Opening new shell" in result.output

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_go_worktree_not_found(self, mock_get_manager):
        """Test go command with non-existent worktree."""
        # Arrange
        runner = CliRunner()
        mock_manager = MagicMock()
        mock_get_manager.return_value = mock_manager
        mock_manager.registry.get.return_value = None

        # Act
        result = runner.invoke(worktree, ["go", "NONEXISTENT"])

        # Assert
        assert result.exit_code != 0


class TestSyncWorktreeCommand:
    """Test sync worktree command."""

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_sync_single_worktree_success(self, mock_get_manager):
        """Test syncing a single worktree."""
        # Arrange
        runner = CliRunner()
        mock_manager = MagicMock()
        mock_get_manager.return_value = mock_manager

        # Act
        result = runner.invoke(worktree, ["sync", "SPEC-001"])

        # Assert
        assert result.exit_code == 0
        mock_manager.sync.assert_called_once()

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_sync_worktree_not_found(self, mock_get_manager):
        """Test syncing non-existent worktree."""
        # Arrange
        runner = CliRunner()
        mock_manager = MagicMock()
        mock_get_manager.return_value = mock_manager
        mock_manager.sync.side_effect = WorktreeNotFoundError("SPEC-001")

        # Act
        result = runner.invoke(worktree, ["sync", "SPEC-001"])

        # Assert
        assert result.exit_code != 0

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_sync_worktree_with_rebase(self, mock_get_manager):
        """Test syncing worktree with rebase strategy."""
        # Arrange
        runner = CliRunner()
        mock_manager = MagicMock()
        mock_get_manager.return_value = mock_manager

        # Act
        result = runner.invoke(worktree, ["sync", "SPEC-001", "--rebase"])

        # Assert
        assert result.exit_code == 0
        call_kwargs = mock_manager.sync.call_args[1]
        assert call_kwargs["rebase"] is True

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_sync_worktree_with_ff_only(self, mock_get_manager):
        """Test syncing worktree with fast-forward only."""
        # Arrange
        runner = CliRunner()
        mock_manager = MagicMock()
        mock_get_manager.return_value = mock_manager

        # Act
        result = runner.invoke(worktree, ["sync", "SPEC-001", "--ff-only"])

        # Assert
        assert result.exit_code == 0
        call_kwargs = mock_manager.sync.call_args[1]
        assert call_kwargs["ff_only"] is True

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_sync_all_worktrees(self, mock_get_manager):
        """Test syncing all worktrees."""
        # Arrange
        runner = CliRunner()
        mock_manager = MagicMock()
        mock_get_manager.return_value = mock_manager

        now = "2025-01-01T12:00:00Z"
        worktree_info = WorktreeInfo(
            spec_id="SPEC-001",
            path=Path("/tmp/SPEC-001"),
            branch="feature/SPEC-001",
            created_at=now,
            last_accessed=now,
            status="active",
        )
        mock_manager.list.return_value = [worktree_info]

        # Act
        result = runner.invoke(worktree, ["sync", "--all"])

        # Assert
        assert result.exit_code == 0
        assert "Syncing" in result.output

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_sync_all_no_worktrees(self, mock_get_manager):
        """Test syncing all when no worktrees exist."""
        # Arrange
        runner = CliRunner()
        mock_manager = MagicMock()
        mock_get_manager.return_value = mock_manager
        mock_manager.list.return_value = []

        # Act
        result = runner.invoke(worktree, ["sync", "--all"])

        # Assert
        assert result.exit_code == 0
        assert "No worktrees found" in result.output

    def test_sync_without_spec_or_all_flag(self):
        """Test sync command without SPEC_ID or --all flag."""
        # Arrange
        runner = CliRunner()

        # Act
        result = runner.invoke(worktree, ["sync"])

        # Assert
        assert result.exit_code != 0
        assert "Either SPEC_ID or --all option is required" in result.output

    def test_sync_with_spec_and_all_flag(self):
        """Test sync command with both SPEC_ID and --all flag."""
        # Arrange
        runner = CliRunner()

        # Act
        result = runner.invoke(worktree, ["sync", "SPEC-001", "--all"])

        # Assert
        assert result.exit_code != 0
        assert "Cannot use both SPEC_ID and --all option" in result.output


class TestCleanWorktreesCommand:
    """Test clean worktrees command."""

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_clean_merged_only(self, mock_get_manager):
        """Test cleaning merged branches only."""
        # Arrange
        runner = CliRunner()
        mock_manager = MagicMock()
        mock_get_manager.return_value = mock_manager
        mock_manager.clean_merged.return_value = ["SPEC-001"]

        # Act
        result = runner.invoke(worktree, ["clean", "--merged-only"])

        # Assert
        assert result.exit_code == 0
        mock_manager.clean_merged.assert_called_once()

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_clean_default_all_worktrees(self, mock_get_manager):
        """Test clean command without flags removes all worktrees."""
        # Arrange
        runner = CliRunner()
        mock_manager = MagicMock()
        mock_get_manager.return_value = mock_manager

        now = "2025-01-01T12:00:00Z"
        worktree_info = WorktreeInfo(
            spec_id="SPEC-001",
            path=Path("/tmp/SPEC-001"),
            branch="feature/SPEC-001",
            created_at=now,
            last_accessed=now,
            status="active",
        )
        mock_manager.list.return_value = [worktree_info]

        # Act
        result = runner.invoke(worktree, ["clean"])

        # Assert
        assert result.exit_code == 0
        assert "Removing all worktrees" in result.output

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_clean_empty_worktrees(self, mock_get_manager):
        """Test clean command with no worktrees."""
        # Arrange
        runner = CliRunner()
        mock_manager = MagicMock()
        mock_get_manager.return_value = mock_manager
        mock_manager.list.return_value = []
        mock_manager.clean_merged.return_value = []

        # Act
        result = runner.invoke(worktree, ["clean", "--merged-only"])

        # Assert
        assert result.exit_code == 0


class TestConfigWorktreeCommand:
    """Test config worktree command."""

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_config_show_all(self, mock_get_manager):
        """Test showing all configuration."""
        # Arrange
        runner = CliRunner()
        mock_manager = MagicMock()
        mock_get_manager.return_value = mock_manager
        mock_manager.worktree_root = Path("/tmp/worktrees")
        mock_manager.registry.registry_path = Path("/tmp/worktrees/.moai-worktree-registry.json")

        # Act
        result = runner.invoke(worktree, ["config"])

        # Assert
        assert result.exit_code == 0
        assert "Configuration" in result.output

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_config_get_root(self, mock_get_manager):
        """Test getting worktree root configuration."""
        # Arrange
        runner = CliRunner()
        mock_manager = MagicMock()
        mock_get_manager.return_value = mock_manager
        mock_manager.worktree_root = Path("/tmp/worktrees")

        # Act
        result = runner.invoke(worktree, ["config", "root"])

        # Assert
        assert result.exit_code == 0
        assert "Worktree root" in result.output

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_config_get_registry(self, mock_get_manager):
        """Test getting registry path configuration."""
        # Arrange
        runner = CliRunner()
        mock_manager = MagicMock()
        mock_get_manager.return_value = mock_manager
        mock_manager.registry.registry_path = Path("/tmp/worktrees/.moai-worktree-registry.json")

        # Act
        result = runner.invoke(worktree, ["config", "registry"])

        # Assert
        assert result.exit_code == 0
        assert "Registry path" in result.output

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_config_unknown_key(self, mock_get_manager):
        """Test config with unknown key."""
        # Arrange
        runner = CliRunner()
        mock_manager = MagicMock()
        mock_get_manager.return_value = mock_manager

        # Act
        result = runner.invoke(worktree, ["config", "unknown"])

        # Assert
        assert result.exit_code == 0
        assert "Unknown config key" in result.output

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_config_set_root_not_allowed(self, mock_get_manager):
        """Test that setting root is not allowed via config command."""
        # Arrange
        runner = CliRunner()
        mock_manager = MagicMock()
        mock_get_manager.return_value = mock_manager

        # Act
        result = runner.invoke(worktree, ["config", "root", "/new/root"])

        # Assert
        assert result.exit_code == 0
        assert "Use --worktree-root option" in result.output


class TestCliErrorHandling:
    """Test CLI error handling and edge cases."""

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_cli_handles_generic_exception(self, mock_get_manager):
        """Test CLI handles unexpected exceptions gracefully."""
        # Arrange
        runner = CliRunner()
        mock_manager = MagicMock()
        mock_get_manager.return_value = mock_manager
        mock_manager.create.side_effect = Exception("Unexpected error")

        # Act
        result = runner.invoke(worktree, ["new", "SPEC-001"])

        # Assert
        assert result.exit_code != 0

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_sync_with_merge_conflict_no_auto_resolve(self, mock_get_manager):
        """Test sync with merge conflict and no auto-resolve."""
        # Arrange
        runner = CliRunner()
        mock_manager = MagicMock()
        mock_get_manager.return_value = mock_manager
        mock_manager.sync.side_effect = MergeConflictError("SPEC-001", ["file1.py"])

        # Act
        result = runner.invoke(worktree, ["sync", "SPEC-001"])

        # Assert
        assert result.exit_code != 0

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_new_worktree_with_custom_repo_path(self, mock_get_manager):
        """Test creating worktree with custom repo path."""
        # Arrange
        runner = CliRunner()
        mock_manager = MagicMock()
        mock_get_manager.return_value = mock_manager

        now = "2025-01-01T12:00:00Z"
        worktree_info = WorktreeInfo(
            spec_id="SPEC-001",
            path=Path("/tmp/SPEC-001"),
            branch="feature/SPEC-001",
            created_at=now,
            last_accessed=now,
            status="active",
        )
        mock_manager.create.return_value = worktree_info

        # Act
        result = runner.invoke(worktree, ["new", "SPEC-001", "--repo", "/custom/repo"])

        # Assert
        assert result.exit_code == 0
        # Verify get_manager was called
        mock_get_manager.assert_called_once()

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_list_with_custom_repo_path(self, mock_get_manager):
        """Test listing worktrees with custom repo path."""
        # Arrange
        runner = CliRunner()
        mock_manager = MagicMock()
        mock_get_manager.return_value = mock_manager
        mock_manager.list.return_value = []

        # Act
        result = runner.invoke(worktree, ["list", "--repo", "/custom/repo"])

        # Assert
        assert result.exit_code == 0
        # Verify get_manager was called
        mock_get_manager.assert_called_once()


class TestWorktreeCommandIntegration:
    """Integration tests for worktree commands."""

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_worktree_group_has_subcommands(self, mock_get_manager):
        """Test that worktree group has all expected subcommands."""
        # Arrange
        runner = CliRunner()

        # Act
        result = runner.invoke(worktree, ["--help"])

        # Assert
        assert result.exit_code == 0
        expected_commands = [
            "new",
            "list",
            "remove",
            "status",
            "go",
            "sync",
            "clean",
            "config",
        ]
        for cmd in expected_commands:
            assert cmd in result.output

    @patch("moai_adk.cli.worktree.cli.get_manager")
    def test_sync_all_with_auto_resolve(self, mock_get_manager):
        """Test syncing all worktrees with auto-resolve."""
        # Arrange
        runner = CliRunner()
        mock_manager = MagicMock()
        mock_get_manager.return_value = mock_manager

        now = "2025-01-01T12:00:00Z"
        worktree_info = WorktreeInfo(
            spec_id="SPEC-001",
            path=Path("/tmp/SPEC-001"),
            branch="feature/SPEC-001",
            created_at=now,
            last_accessed=now,
            status="active",
        )
        mock_manager.list.return_value = [worktree_info]

        # Act
        result = runner.invoke(worktree, ["sync", "--all", "--auto-resolve"])

        # Assert
        assert result.exit_code == 0
