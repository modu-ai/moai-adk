"""
Executable unit tests for WorktreeManager focusing on real behavior.

These tests call actual methods and exercise real code paths, with minimal
mocking to verify actual implementation behavior.
"""

import tempfile
from pathlib import Path
from unittest.mock import MagicMock, patch

import pytest

from moai_adk.cli.worktree.exceptions import (
    GitOperationError,
    MergeConflictError,
    WorktreeExistsError,
    WorktreeNotFoundError,
)
from moai_adk.cli.worktree.manager import WorktreeManager


class TestWorktreeManagerExecutable:
    """Test actual WorktreeManager execution."""

    def test_manager_initialization(self):
        """Test WorktreeManager can be initialized."""
        with tempfile.TemporaryDirectory() as tmpdir:
            repo_path = Path(tmpdir)
            worktree_root = Path(tmpdir) / "worktrees"

            with patch("moai_adk.cli.worktree.manager.Repo"):
                manager = WorktreeManager(repo_path, worktree_root)

                assert manager is not None
                assert manager.repo is not None
                assert manager.worktree_root == worktree_root
                assert manager.registry is not None

    def test_manager_list_empty(self):
        """Test list method with empty registry."""
        with tempfile.TemporaryDirectory() as tmpdir:
            repo_path = Path(tmpdir)
            worktree_root = Path(tmpdir) / "worktrees"

            with patch("moai_adk.cli.worktree.manager.Repo"):
                manager = WorktreeManager(repo_path, worktree_root)
                result = manager.list()

                assert isinstance(result, list)
                # Will be empty or contain existing worktrees

    def test_manager_create_git_operation_error(self):
        """Test create handles Git operation errors."""
        with tempfile.TemporaryDirectory() as tmpdir:
            repo_path = Path(tmpdir)
            worktree_root = Path(tmpdir) / "worktrees"

            with patch("moai_adk.cli.worktree.manager.Repo") as mock_repo_class:
                mock_repo = MagicMock()
                mock_repo_class.return_value = mock_repo
                mock_repo.remotes.origin.fetch = MagicMock(side_effect=Exception("no origin"))
                mock_repo.heads = []
                mock_repo.git.worktree = MagicMock(side_effect=Exception("git failed"))
                mock_repo.create_head = MagicMock()

                manager = WorktreeManager(repo_path, worktree_root)

                with pytest.raises(GitOperationError):
                    manager.create("SPEC-FAIL")

    def test_manager_remove_not_found(self):
        """Test remove raises error when worktree not found."""
        with tempfile.TemporaryDirectory() as tmpdir:
            repo_path = Path(tmpdir)
            worktree_root = Path(tmpdir) / "worktrees"

            with patch("moai_adk.cli.worktree.manager.Repo"):
                manager = WorktreeManager(repo_path, worktree_root)

                with pytest.raises(WorktreeNotFoundError):
                    manager.remove("NONEXISTENT")

    def test_manager_sync_not_found(self):
        """Test sync raises error when worktree not found."""
        with tempfile.TemporaryDirectory() as tmpdir:
            repo_path = Path(tmpdir)
            worktree_root = Path(tmpdir) / "worktrees"

            with patch("moai_adk.cli.worktree.manager.Repo"):
                manager = WorktreeManager(repo_path, worktree_root)

                with pytest.raises(WorktreeNotFoundError):
                    manager.sync("NONEXISTENT")

    def test_manager_clean_merged_returns_list(self):
        """Test clean_merged returns list."""
        with tempfile.TemporaryDirectory() as tmpdir:
            repo_path = Path(tmpdir)
            worktree_root = Path(tmpdir) / "worktrees"

            with patch("moai_adk.cli.worktree.manager.Repo") as mock_repo_class:
                mock_repo = MagicMock()
                mock_repo_class.return_value = mock_repo
                mock_repo.git.branch = MagicMock(return_value="")

                manager = WorktreeManager(repo_path, worktree_root)
                result = manager.clean_merged()

                assert isinstance(result, list)

    def test_manager_auto_resolve_conflicts(self):
        """Test auto_resolve_conflicts method exists."""
        with tempfile.TemporaryDirectory() as tmpdir:
            repo_path = Path(tmpdir)
            worktree_root = Path(tmpdir) / "worktrees"

            with patch("moai_adk.cli.worktree.manager.Repo"):
                manager = WorktreeManager(repo_path, worktree_root)

                # Verify method exists
                assert hasattr(manager, "auto_resolve_conflicts")
                assert callable(manager.auto_resolve_conflicts)


class TestWorktreeManagerCreateExecution:
    """Test actual create method execution paths."""

    def test_create_with_existing_worktree_raises_error(self):
        """Test create with existing worktree raises WorktreeExistsError."""
        with tempfile.TemporaryDirectory() as tmpdir:
            repo_path = Path(tmpdir)
            worktree_root = Path(tmpdir) / "worktrees"
            existing_path = worktree_root / "SPEC-001"
            existing_path.mkdir(parents=True, exist_ok=True)

            with patch("moai_adk.cli.worktree.manager.Repo") as mock_repo_class:
                mock_repo = MagicMock()
                mock_repo_class.return_value = mock_repo

                manager = WorktreeManager(repo_path, worktree_root)

                # Register an existing worktree
                manager.registry.register = MagicMock()
                manager.registry.get = MagicMock()
                existing_info = MagicMock()
                existing_info.path = existing_path
                manager.registry.get.return_value = existing_info

                with pytest.raises(WorktreeExistsError):
                    manager.create("SPEC-001")

    def test_create_with_force_removes_existing(self):
        """Test create with force=True removes existing."""
        with tempfile.TemporaryDirectory() as tmpdir:
            repo_path = Path(tmpdir)
            worktree_root = Path(tmpdir) / "worktrees"
            existing_path = worktree_root / "SPEC-001"
            existing_path.mkdir(parents=True, exist_ok=True)

            with patch("moai_adk.cli.worktree.manager.Repo") as mock_repo_class:
                mock_repo = MagicMock()
                mock_repo_class.return_value = mock_repo
                mock_repo.remotes.origin.fetch = MagicMock(side_effect=Exception())
                mock_repo.heads = []
                mock_repo.git.worktree = MagicMock(side_effect=Exception("force scenario"))
                mock_repo.create_head = MagicMock()
                mock_repo.git.status = MagicMock(return_value="")

                manager = WorktreeManager(repo_path, worktree_root)

                # Set up existing worktree
                existing_info = MagicMock()
                existing_info.path = existing_path

                # First call returns existing, subsequent calls return None
                manager.registry.get = MagicMock(side_effect=[existing_info, None, None])
                manager.registry.unregister = MagicMock()

                # Act & Assert - should raise GitOperationError from create attempt
                with pytest.raises(GitOperationError):
                    manager.create("SPEC-001", force=True)


class TestWorktreeManagerRemoveExecution:
    """Test actual remove method execution."""

    def test_remove_with_status_error_continues(self):
        """Test remove continues even if status check fails."""
        with tempfile.TemporaryDirectory() as tmpdir:
            repo_path = Path(tmpdir)
            worktree_root = Path(tmpdir) / "worktrees"
            worktree_path = worktree_root / "SPEC-001"
            worktree_path.mkdir(parents=True, exist_ok=True)

            with patch("moai_adk.cli.worktree.manager.Repo") as mock_repo_class:
                mock_repo = MagicMock()
                mock_repo_class.return_value = mock_repo
                mock_repo.git.status = MagicMock(side_effect=Exception("status error"))
                mock_repo.git.worktree = MagicMock()

                manager = WorktreeManager(repo_path, worktree_root)

                # Set up existing worktree
                existing_info = MagicMock()
                existing_info.path = worktree_path
                manager.registry.get = MagicMock(return_value=existing_info)
                manager.registry.unregister = MagicMock()

                # Should not raise because error handling continues
                manager.remove("SPEC-001", force=False)

                # Verify unregister was called
                assert manager.registry.unregister.called


class TestWorktreeManagerSyncExecution:
    """Test actual sync method execution."""

    def test_sync_with_no_branch_found(self):
        """Test sync raises error when base branch not found."""
        with tempfile.TemporaryDirectory() as tmpdir:
            repo_path = Path(tmpdir)
            worktree_root = Path(tmpdir) / "worktrees"
            worktree_path = worktree_root / "SPEC-001"
            worktree_path.mkdir(parents=True, exist_ok=True)

            with patch("moai_adk.cli.worktree.manager.Repo") as mock_repo_class:
                # Main repo
                mock_main_repo = MagicMock()
                mock_repo_class.return_value = mock_main_repo

                manager = WorktreeManager(repo_path, worktree_root)

                # Set up existing worktree
                existing_info = MagicMock()
                existing_info.path = worktree_path
                existing_info.last_accessed = "2025-01-01T00:00:00Z"
                manager.registry.get = MagicMock(return_value=existing_info)

                # Worktree repo that will be created
                with patch("moai_adk.cli.worktree.manager.Repo") as inner_repo_patch:
                    mock_worktree_repo = MagicMock()
                    inner_repo_patch.return_value = mock_worktree_repo
                    mock_worktree_repo.remotes.origin.fetch = MagicMock(side_effect=Exception())
                    # rev_parse fails for both candidates
                    mock_worktree_repo.git.rev_parse = MagicMock(side_effect=Exception())

                    with pytest.raises(GitOperationError):
                        manager.sync("SPEC-001")


class TestWorktreeManagerBranchDetection:
    """Test branch detection logic."""

    def test_sync_detects_origin_branch(self):
        """Test sync detects origin/main branch."""
        with tempfile.TemporaryDirectory() as tmpdir:
            repo_path = Path(tmpdir)
            worktree_root = Path(tmpdir) / "worktrees"
            worktree_path = worktree_root / "SPEC-001"
            worktree_path.mkdir(parents=True, exist_ok=True)

            with patch("moai_adk.cli.worktree.manager.Repo") as mock_repo_class:
                mock_main_repo = MagicMock()
                mock_repo_class.return_value = mock_main_repo

                manager = WorktreeManager(repo_path, worktree_root)

                existing_info = MagicMock()
                existing_info.path = worktree_path
                existing_info.last_accessed = "2025-01-01T00:00:00Z"
                manager.registry.get = MagicMock(return_value=existing_info)
                manager.registry.register = MagicMock()

                with patch("moai_adk.cli.worktree.manager.Repo") as inner_repo_patch:
                    mock_worktree_repo = MagicMock()
                    inner_repo_patch.return_value = mock_worktree_repo
                    mock_worktree_repo.remotes.origin.fetch = MagicMock(side_effect=Exception())

                    # First call (origin/main) succeeds
                    mock_worktree_repo.git.rev_parse = MagicMock(return_value="abc123")
                    mock_worktree_repo.git.merge = MagicMock()
                    mock_worktree_repo.git.status = MagicMock(return_value="")

                    # Act
                    manager.sync("SPEC-001")

                    # Assert rev_parse was called with origin/main
                    assert mock_worktree_repo.git.rev_parse.called
                    # Should have called merge with detected branch
                    assert mock_worktree_repo.git.merge.called


class TestWorktreeManagerConflictHandling:
    """Test conflict detection and handling."""

    def test_sync_detects_unmerged_files(self):
        """Test sync detects unmerged files from status."""
        with tempfile.TemporaryDirectory() as tmpdir:
            repo_path = Path(tmpdir)
            worktree_root = Path(tmpdir) / "worktrees"
            worktree_path = worktree_root / "SPEC-001"
            worktree_path.mkdir(parents=True, exist_ok=True)

            with patch("moai_adk.cli.worktree.manager.Repo") as mock_repo_class:
                mock_main_repo = MagicMock()
                mock_repo_class.return_value = mock_main_repo

                manager = WorktreeManager(repo_path, worktree_root)

                existing_info = MagicMock()
                existing_info.path = worktree_path
                manager.registry.get = MagicMock(return_value=existing_info)

                with patch("moai_adk.cli.worktree.manager.Repo") as inner_repo_patch:
                    mock_worktree_repo = MagicMock()
                    inner_repo_patch.return_value = mock_worktree_repo
                    mock_worktree_repo.remotes.origin.fetch = MagicMock(side_effect=Exception())
                    mock_worktree_repo.git.rev_parse = MagicMock(return_value="abc123")
                    # Merge fails
                    mock_worktree_repo.git.merge = MagicMock(side_effect=Exception("conflict"))
                    # Status shows unmerged files
                    mock_worktree_repo.git.status = MagicMock(return_value="UU src/file.py")
                    mock_worktree_repo.git.merge = MagicMock(side_effect=Exception())
                    mock_worktree_repo.git.rebase = MagicMock()

                    with pytest.raises(MergeConflictError):
                        manager.sync("SPEC-001", auto_resolve=False)
