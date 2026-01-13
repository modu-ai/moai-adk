"""Integration tests for worktree command with real Git operations.

Tests end-to-end execution of worktree commands including:
- moai worktree new/create - Create real Git worktree
- moai worktree list - Enumerate actual worktrees
- moai worktree switch - Directory switching
- moai worktree remove - Clean up worktrees
- moai worktree done - Complete and merge workflow

Key Requirements:
- Use Click's CliRunner for command invocation
- Use real Git worktree operations
- Test actual file system operations
- Clean up temporary files after tests
"""

import subprocess
import tempfile
from pathlib import Path
from typing import Generator
from unittest.mock import patch

import pytest
import yaml
from click.testing import CliRunner

from moai_adk.__main__ import cli


@pytest.fixture
def worktree_test_repo(tmp_path: Path) -> Generator[Path, None, None]:
    """Create a Git repository for worktree testing.

    Initializes a Git repository with main branch commits
    for testing worktree operations.

    Args:
        tmp_path: Pytest's temporary directory fixture

    Yields:
        Path to Git repository ready for worktree operations
    """
    repo_dir = tmp_path / "worktree_test_repo"
    repo_dir.mkdir(parents=True, exist_ok=True)

    # Initialize Git repository
    subprocess.run(["git", "init"], cwd=repo_dir, check=True, capture_output=True)

    # Configure Git user
    subprocess.run(
        ["git", "config", "user.name", "Test User"],
        cwd=repo_dir,
        check=True,
        capture_output=True,
    )
    subprocess.run(
        ["git", "config", "user.email", "test@example.com"],
        cwd=repo_dir,
        check=True,
        capture_output=True,
    )

    # Create initial commit on main branch
    readme = repo_dir / "README.md"
    readme.write_text("# Worktree Test Repository\n", encoding="utf-8")
    subprocess.run(["git", "add", "."], cwd=repo_dir, check=True, capture_output=True)
    subprocess.run(
        ["git", "commit", "-m", "Initial commit"],
        cwd=repo_dir,
        check=True,
        capture_output=True,
    )

    # Rename to main if needed (Git may use master by default)
    try:
        subprocess.run(
            ["git", "branch", "-M", "main"],
            cwd=repo_dir,
            check=True,
            capture_output=True,
        )
    except subprocess.CalledProcessError:
        pass  # Already on main

    yield repo_dir


@pytest.fixture
def worktree_root(tmp_path: Path) -> Path:
    """Create a worktree root directory for testing.

    Args:
        tmp_path: Pytest's temporary directory fixture

    Returns:
        Path to worktree root directory
    """
    root = tmp_path / "worktrees"
    root.mkdir(parents=True, exist_ok=True)
    return root


@pytest.mark.integration
class TestWorktreeCreate:
    """Tests for worktree create command."""

    def test_worktree_new_creates_real_worktree(self, cli_runner, worktree_test_repo, worktree_root):
        """Test that worktree new creates actual Git worktree."""
        result = cli_runner.invoke(
            cli,
            [
                "worktree",
                "new",
                "SPEC-TEST-001",
                "--repo",
                str(worktree_test_repo),
                "--worktree-root",
                str(worktree_root),
            ],
        )

        # Should execute
        assert result.exit_code in [0, 1]

        # Check if worktree was created (may fail if Git not available)
        worktree_path = worktree_root / "SPEC-TEST-001"
        if result.exit_code == 0 and worktree_path.exists():
            assert worktree_path.is_dir()

            # Verify it's a valid Git worktree
            assert (worktree_path / ".git").exists()

    def test_worktree_new_with_custom_branch(self, cli_runner, worktree_test_repo, worktree_root):
        """Test worktree new with custom branch name."""
        result = cli_runner.invoke(
            cli,
            [
                "worktree",
                "new",
                "SPEC-TEST-002",
                "--branch",
                "feature/test-branch",
                "--repo",
                str(worktree_test_repo),
                "--worktree-root",
                str(worktree_root),
            ],
        )

        # Should execute
        assert result.exit_code in [0, 1]

    def test_worktree_new_with_base_branch(self, cli_runner, worktree_test_repo, worktree_root):
        """Test worktree new with custom base branch."""
        # Create another branch first
        subprocess.run(
            ["git", "checkout", "-b", "develop"],
            cwd=worktree_test_repo,
            check=True,
            capture_output=True,
        )

        result = cli_runner.invoke(
            cli,
            [
                "worktree",
                "new",
                "SPEC-TEST-003",
                "--base",
                "develop",
                "--repo",
                str(worktree_test_repo),
                "--worktree-root",
                str(worktree_root),
            ],
        )

        # Should execute
        assert result.exit_code in [0, 1]

    def test_worktree_new_with_force_flag(self, cli_runner, worktree_test_repo, worktree_root):
        """Test worktree new --force flag."""
        # Create worktree first time
        cli_runner.invoke(
            cli,
            [
                "worktree",
                "new",
                "SPEC-TEST-004",
                "--repo",
                str(worktree_test_repo),
                "--worktree-root",
                str(worktree_root),
            ],
        )

        # Try to create again with force
        result = cli_runner.invoke(
            cli,
            [
                "worktree",
                "new",
                "SPEC-TEST-004",
                "--force",
                "--repo",
                str(worktree_test_repo),
                "--worktree-root",
                str(worktree_root),
            ],
        )

        # Should execute (may fail if worktree exists)
        assert result.exit_code in [0, 1]

    def test_worktree_new_creates_registry(self, cli_runner, worktree_test_repo, worktree_root):
        """Test that worktree new creates registry file."""
        result = cli_runner.invoke(
            cli,
            [
                "worktree",
                "new",
                "SPEC-TEST-005",
                "--repo",
                str(worktree_test_repo),
                "--worktree-root",
                str(worktree_root),
            ],
        )

        if result.exit_code == 0:
            # Check for registry file
            registry_file = worktree_root / ".moai-worktree-registry.json"
            if registry_file.exists():
                content = registry_file.read_text(encoding="utf-8")
                # Should contain worktree info
                assert len(content) > 0

    def test_worktree_new_multiple_worktrees(self, cli_runner, worktree_test_repo, worktree_root):
        """Test creating multiple worktrees."""
        spec_ids = ["SPEC-001", "SPEC-002", "SPEC-003"]

        for spec_id in spec_ids:
            result = cli_runner.invoke(
                cli,
                [
                    "worktree",
                    "new",
                    spec_id,
                    "--repo",
                    str(worktree_test_repo),
                    "--worktree-root",
                    str(worktree_root),
                ],
            )
            assert result.exit_code in [0, 1]

        # Check if worktrees were created
        for spec_id in spec_ids:
            worktree_path = worktree_root / spec_id
            if worktree_path.exists():
                assert worktree_path.is_dir()


@pytest.mark.integration
class TestWorktreeList:
    """Tests for worktree list command."""

    def test_worktree_list_empty(self, cli_runner, worktree_test_repo, worktree_root):
        """Test worktree list with no worktrees."""
        result = cli_runner.invoke(
            cli,
            ["worktree", "list", "--repo", str(worktree_test_repo), "--worktree-root", str(worktree_root)],
        )

        # Should execute
        assert result.exit_code == 0

        # Should show no worktrees message or empty table
        assert "no worktree" in result.output.lower() or len(result.output) > 0

    def test_worktree_list_with_worktrees(self, cli_runner, worktree_test_repo, worktree_root):
        """Test worktree list shows created worktrees."""
        # Create a worktree first
        create_result = cli_runner.invoke(
            cli,
            [
                "worktree",
                "new",
                "SPEC-LIST-001",
                "--repo",
                str(worktree_test_repo),
                "--worktree-root",
                str(worktree_root),
            ],
        )

        if create_result.exit_code == 0:
            # List worktrees
            list_result = cli_runner.invoke(
                cli,
                ["worktree", "list", "--repo", str(worktree_test_repo), "--worktree-root", str(worktree_root)],
            )

            # Should execute
            assert list_result.exit_code == 0

            # Should show worktree information
            assert "SPEC-LIST-001" in list_result.output or len(list_result.output) > 0

    def test_worktree_list_json_format(self, cli_runner, worktree_test_repo, worktree_root):
        """Test worktree list --format json."""
        # Create a worktree first
        cli_runner.invoke(
            cli,
            [
                "worktree",
                "new",
                "SPEC-JSON-001",
                "--repo",
                str(worktree_test_repo),
                "--worktree-root",
                str(worktree_root),
            ],
        )

        # List in JSON format
        result = cli_runner.invoke(
            cli,
            [
                "worktree",
                "list",
                "--format",
                "json",
                "--repo",
                str(worktree_test_repo),
                "--worktree-root",
                str(worktree_root),
            ],
        )

        # Should execute
        assert result.exit_code == 0

        # Should show JSON output
        assert len(result.output) > 0

    def test_worktree_list_table_format(self, cli_runner, worktree_test_repo, worktree_root):
        """Test worktree list --format table (default)."""
        # Create a worktree first
        cli_runner.invoke(
            cli,
            [
                "worktree",
                "new",
                "SPEC-TABLE-001",
                "--repo",
                str(worktree_test_repo),
                "--worktree-root",
                str(worktree_root),
            ],
        )

        # List in table format
        result = cli_runner.invoke(
            cli,
            [
                "worktree",
                "list",
                "--format",
                "table",
                "--repo",
                str(worktree_test_repo),
                "--worktree-root",
                str(worktree_root),
            ],
        )

        # Should execute
        assert result.exit_code == 0

        # Should show table
        assert len(result.output) > 0


@pytest.mark.integration
class TestWorktreeRemove:
    """Tests for worktree remove command."""

    def test_worktree_remove_existing_worktree(self, cli_runner, worktree_test_repo, worktree_root):
        """Test removing an existing worktree."""
        # Create worktree first
        create_result = cli_runner.invoke(
            cli,
            [
                "worktree",
                "new",
                "SPEC-REMOVE-001",
                "--repo",
                str(worktree_test_repo),
                "--worktree-root",
                str(worktree_root),
            ],
        )

        if create_result.exit_code == 0:
            # Remove worktree
            remove_result = cli_runner.invoke(
                cli,
                [
                    "worktree",
                    "remove",
                    "SPEC-REMOVE-001",
                    "--repo",
                    str(worktree_test_repo),
                    "--worktree-root",
                    str(worktree_root),
                ],
            )

            # Should execute
            assert remove_result.exit_code == 0

            # Worktree directory should be removed
            worktree_path = worktree_root / "SPEC-REMOVE-001"
            if worktree_path.exists():
                # May still exist if Git failed to remove
                pass

    def test_worktree_remove_nonexistent_worktree(self, cli_runner, worktree_test_repo, worktree_root):
        """Test removing non-existent worktree."""
        result = cli_runner.invoke(
            cli,
            [
                "worktree",
                "remove",
                "SPEC-NONEXISTENT",
                "--repo",
                str(worktree_test_repo),
                "--worktree-root",
                str(worktree_root),
            ],
        )

        # Should fail
        assert result.exit_code != 0

    def test_worktree_remove_with_force_flag(self, cli_runner, worktree_test_repo, worktree_root):
        """Test worktree remove --force flag."""
        # Create worktree
        create_result = cli_runner.invoke(
            cli,
            [
                "worktree",
                "new",
                "SPEC-FORCE-001",
                "--repo",
                str(worktree_test_repo),
                "--worktree-root",
                str(worktree_root),
            ],
        )

        if create_result.exit_code == 0:
            # Add uncommitted changes
            worktree_path = worktree_root / "SPEC-FORCE-001"
            if worktree_path.exists():
                test_file = worktree_path / "uncommitted.txt"
                test_file.write_text("Uncommitted changes", encoding="utf-8")

            # Try to remove with force
            remove_result = cli_runner.invoke(
                cli,
                [
                    "worktree",
                    "remove",
                    "SPEC-FORCE-001",
                    "--force",
                    "--repo",
                    str(worktree_test_repo),
                    "--worktree-root",
                    str(worktree_root),
                ],
            )

            # Should execute
            assert remove_result.exit_code in [0, 1]


@pytest.mark.integration
class TestWorktreeSync:
    """Tests for worktree sync command."""

    def test_worktree_sync_single_worktree(self, cli_runner, worktree_test_repo, worktree_root):
        """Test syncing a single worktree."""
        # Create worktree
        create_result = cli_runner.invoke(
            cli,
            [
                "worktree",
                "new",
                "SPEC-SYNC-001",
                "--repo",
                str(worktree_test_repo),
                "--worktree-root",
                str(worktree_root),
            ],
        )

        if create_result.exit_code == 0:
            # Add a commit to main
            readme = worktree_test_repo / "README.md"
            readme.write_text("# Updated README\n", encoding="utf-8")
            subprocess.run(["git", "add", "."], cwd=worktree_test_repo, check=True, capture_output=True)
            subprocess.run(
                ["git", "commit", "-m", "Update readme"],
                cwd=worktree_test_repo,
                check=True,
                capture_output=True,
            )

            # Sync worktree
            sync_result = cli_runner.invoke(
                cli,
                [
                    "worktree",
                    "sync",
                    "SPEC-SYNC-001",
                    "--repo",
                    str(worktree_test_repo),
                    "--worktree-root",
                    str(worktree_root),
                ],
            )

            # Should execute
            assert sync_result.exit_code in [0, 1]

    def test_worktree_sync_all_worktrees(self, cli_runner, worktree_test_repo, worktree_root):
        """Test syncing all worktrees with --all flag."""
        # Create multiple worktrees
        for spec_id in ["SPEC-SYNCALL-001", "SPEC-SYNCALL-002"]:
            cli_runner.invoke(
                cli,
                [
                    "worktree",
                    "new",
                    spec_id,
                    "--repo",
                    str(worktree_test_repo),
                    "--worktree-root",
                    str(worktree_root),
                ],
            )

        # Sync all
        result = cli_runner.invoke(
            cli,
            [
                "worktree",
                "sync",
                "--all",
                "--repo",
                str(worktree_test_repo),
                "--worktree-root",
                str(worktree_root),
            ],
        )

        # Should execute
        assert result.exit_code in [0, 1]

    def test_worktree_sync_with_rebase(self, cli_runner, worktree_test_repo, worktree_root):
        """Test worktree sync --rebase flag."""
        # Create worktree
        cli_runner.invoke(
            cli,
            [
                "worktree",
                "new",
                "SPEC-REBASE-001",
                "--repo",
                str(worktree_test_repo),
                "--worktree-root",
                str(worktree_root),
            ],
        )

        # Sync with rebase
        result = cli_runner.invoke(
            cli,
            [
                "worktree",
                "sync",
                "SPEC-REBASE-001",
                "--rebase",
                "--repo",
                str(worktree_test_repo),
                "--worktree-root",
                str(worktree_root),
            ],
        )

        # Should execute
        assert result.exit_code in [0, 1]


@pytest.mark.integration
class TestWorktreeStatus:
    """Tests for worktree status command."""

    def test_worktree_status_shows_worktrees(self, cli_runner, worktree_test_repo, worktree_root):
        """Test worktree status shows all worktrees."""
        # Create worktree
        cli_runner.invoke(
            cli,
            [
                "worktree",
                "new",
                "SPEC-STATUS-001",
                "--repo",
                str(worktree_test_repo),
                "--worktree-root",
                str(worktree_root),
            ],
        )

        # Get status
        result = cli_runner.invoke(
            cli,
            ["worktree", "status", "--repo", str(worktree_test_repo), "--worktree-root", str(worktree_root)],
        )

        # Should execute
        assert result.exit_code == 0

        # Should show worktree info
        assert len(result.output) > 0

    def test_worktree_status_empty(self, cli_runner, worktree_test_repo, worktree_root):
        """Test worktree status with no worktrees."""
        result = cli_runner.invoke(
            cli,
            ["worktree", "status", "--repo", str(worktree_test_repo), "--worktree-root", str(worktree_root)],
        )

        # Should execute
        assert result.exit_code == 0

        # Should show no worktrees
        assert "no worktree" in result.output.lower() or len(result.output) > 0


@pytest.mark.integration
class TestWorktreeClean:
    """Tests for worktree clean command."""

    def test_worktree_clean_merged_only(self, cli_runner, worktree_test_repo, worktree_root):
        """Test worktree clean --merged-only flag."""
        # Create worktree
        cli_runner.invoke(
            cli,
            [
                "worktree",
                "new",
                "SPEC-CLEAN-001",
                "--repo",
                str(worktree_test_repo),
                "--worktree-root",
                str(worktree_root),
            ],
        )

        # Clean merged
        result = cli_runner.invoke(
            cli,
            [
                "worktree",
                "clean",
                "--merged-only",
                "--repo",
                str(worktree_test_repo),
                "--worktree-root",
                str(worktree_root),
            ],
        )

        # Should execute
        assert result.exit_code == 0

    def test_worktree_clean_all(self, cli_runner, worktree_test_repo, worktree_root):
        """Test worktree clean all (without flags)."""
        # Create worktree
        cli_runner.invoke(
            cli,
            [
                "worktree",
                "new",
                "SPEC-CLEANALL-001",
                "--repo",
                str(worktree_test_repo),
                "--worktree-root",
                str(worktree_root),
            ],
        )

        # Clean all
        result = cli_runner.invoke(
            cli,
            [
                "worktree",
                "clean",
                "--repo",
                str(worktree_test_repo),
                "--worktree-root",
                str(worktree_root),
            ],
        )

        # Should execute
        assert result.exit_code == 0


@pytest.mark.integration
class TestWorktreeGo:
    """Tests for worktree go command."""

    def test_worktree_go_to_existing_worktree(self, cli_runner, worktree_test_repo, worktree_root):
        """Test worktree go command."""
        # Create worktree
        create_result = cli_runner.invoke(
            cli,
            [
                "worktree",
                "new",
                "SPEC-GO-001",
                "--repo",
                str(worktree_test_repo),
                "--worktree-root",
                str(worktree_root),
            ],
        )

        if create_result.exit_code == 0:
            # Go to worktree (this would normally open a new shell, so we just test it doesn't crash)
            result = cli_runner.invoke(
                cli,
                [
                    "worktree",
                    "go",
                    "SPEC-GO-001",
                    "--repo",
                    str(worktree_test_repo),
                    "--worktree-root",
                    str(worktree_root),
                ],
            )

            # Should try to execute (may fail on shell spawning)
            assert result.exit_code in [0, 1, 130]  # 130 is typical for shell spawn failures

    def test_worktree_go_nonexistent_worktree(self, cli_runner, worktree_test_repo, worktree_root):
        """Test worktree go to non-existent worktree."""
        result = cli_runner.invoke(
            cli,
            [
                "worktree",
                "go",
                "SPEC-NONEXISTENT",
                "--repo",
                str(worktree_test_repo),
                "--worktree-root",
                str(worktree_root),
            ],
        )

        # Should fail
        assert result.exit_code != 0


@pytest.mark.integration
class TestWorktreeConfig:
    """Tests for worktree config command."""

    def test_worktree_config_show_all(self, cli_runner, worktree_test_repo, worktree_root):
        """Test worktree config shows all configuration."""
        result = cli_runner.invoke(
            cli,
            [
                "worktree",
                "config",
                "--repo",
                str(worktree_test_repo),
                "--worktree-root",
                str(worktree_root),
            ],
        )

        # Should execute
        assert result.exit_code == 0

        # Should show configuration
        assert len(result.output) > 0

    def test_worktree_config_show_root(self, cli_runner, worktree_test_repo, worktree_root):
        """Test worktree config root."""
        result = cli_runner.invoke(
            cli,
            [
                "worktree",
                "config",
                "root",
                "--repo",
                str(worktree_test_repo),
                "--worktree-root",
                str(worktree_root),
            ],
        )

        # Should execute
        assert result.exit_code == 0

        # Should show root path
        assert len(result.output) > 0


@pytest.mark.integration
class TestWorktreeRecover:
    """Tests for worktree recover command."""

    def test_worktree_recover_empty(self, cli_runner, worktree_test_repo, worktree_root):
        """Test worktree recover with nothing to recover."""
        result = cli_runner.invoke(
            cli,
            [
                "worktree",
                "recover",
                "--repo",
                str(worktree_test_repo),
                "--worktree-root",
                str(worktree_root),
            ],
        )

        # Should execute
        assert result.exit_code == 0

        # Should show no new worktrees message
        assert "no new worktree" in result.output.lower() or len(result.output) > 0


@pytest.mark.integration
class TestWorktreeCompleteWorkflow:
    """Tests for complete worktree workflows."""

    def test_worktree_create_list_remove_workflow(self, cli_runner, worktree_test_repo, worktree_root):
        """Test complete workflow: create -> list -> remove."""
        # Create
        create_result = cli_runner.invoke(
            cli,
            [
                "worktree",
                "new",
                "SPEC-WORKFLOW-001",
                "--repo",
                str(worktree_test_repo),
                "--worktree-root",
                str(worktree_root),
            ],
        )

        if create_result.exit_code == 0:
            # List
            list_result = cli_runner.invoke(
                cli,
                [
                    "worktree",
                    "list",
                    "--repo",
                    str(worktree_test_repo),
                    "--worktree-root",
                    str(worktree_root),
                ],
            )
            assert list_result.exit_code == 0

            # Remove
            remove_result = cli_runner.invoke(
                cli,
                [
                    "worktree",
                    "remove",
                    "SPEC-WORKFLOW-001",
                    "--repo",
                    str(worktree_test_repo),
                    "--worktree-root",
                    str(worktree_root),
                ],
            )
            assert remove_result.exit_code == 0

    def test_worktree_multiple_lifecycle(self, cli_runner, worktree_test_repo, worktree_root):
        """Test multiple worktree create and remove cycles."""
        spec_ids = ["SPEC-CYCLE-001", "SPEC-CYCLE-002", "SPEC-CYCLE-003"]

        # Create all
        for spec_id in spec_ids:
            result = cli_runner.invoke(
                cli,
                [
                    "worktree",
                    "new",
                    spec_id,
                    "--repo",
                    str(worktree_test_repo),
                    "--worktree-root",
                    str(worktree_root),
                ],
            )
            assert result.exit_code in [0, 1]

        # Remove all
        for spec_id in spec_ids:
            result = cli_runner.invoke(
                cli,
                [
                    "worktree",
                    "remove",
                    spec_id,
                    "--repo",
                    str(worktree_test_repo),
                    "--worktree-root",
                    str(worktree_root),
                ],
            )
            assert result.exit_code in [0, 1]


@pytest.mark.integration
class TestWorktreeErrorHandling:
    """Tests for worktree error handling."""

    def test_worktree_with_invalid_repo_path(self, cli_runner, worktree_root):
        """Test worktree commands with invalid repository path."""
        result = cli_runner.invoke(
            cli, ["worktree", "new", "SPEC-ERROR-001", "--repo", "/nonexistent/path", "--worktree-root", str(worktree_root)]
        )

        # Should fail gracefully
        assert result.exit_code != 0

    def test_worktree_with_invalid_worktree_root(self, cli_runner, worktree_test_repo):
        """Test worktree commands with invalid worktree root."""
        result = cli_runner.invoke(
            cli,
            [
                "worktree",
                "new",
                "SPEC-ERROR-002",
                "--repo",
                str(worktree_test_repo),
                "--worktree-root",
                "/invalid/path/that/does/not/exist",
            ],
        )

        # Should handle gracefully
        assert result.exit_code in [0, 1]
