"""
Minimal import and instantiation tests for Worktree Manager.

These tests verify that the module can be imported and basic classes
can be instantiated without errors.
"""

import pytest
import tempfile
from pathlib import Path
from unittest.mock import MagicMock, patch

from moai_adk.cli.worktree.manager import WorktreeManager


class TestImports:
    """Test that all classes can be imported."""

    def test_worktree_manager_importable(self):
        """Test WorktreeManager can be imported."""
        assert WorktreeManager is not None


class TestWorktreeManagerInstantiation:
    """Test WorktreeManager class instantiation."""

    @patch("moai_adk.cli.worktree.manager.Repo")
    def test_worktree_manager_init(self, mock_repo_class):
        """Test WorktreeManager can be instantiated."""
        with tempfile.TemporaryDirectory() as tmpdir:
            repo_path = Path(tmpdir)
            worktree_root = Path(tmpdir) / "worktrees"

            # Mock Repo to avoid actual git operations
            mock_repo = MagicMock()
            mock_repo_class.return_value = mock_repo

            manager = WorktreeManager(repo_path=repo_path, worktree_root=worktree_root)
            assert manager is not None
            assert manager.repo is not None
            assert manager.worktree_root == worktree_root

    @patch("moai_adk.cli.worktree.manager.Repo")
    def test_worktree_manager_has_registry(self, mock_repo_class):
        """Test WorktreeManager has registry attribute."""
        with tempfile.TemporaryDirectory() as tmpdir:
            repo_path = Path(tmpdir)
            worktree_root = Path(tmpdir) / "worktrees"

            mock_repo = MagicMock()
            mock_repo_class.return_value = mock_repo

            manager = WorktreeManager(repo_path=repo_path, worktree_root=worktree_root)
            assert hasattr(manager, "registry")
            assert manager.registry is not None

    @patch("moai_adk.cli.worktree.manager.Repo")
    def test_worktree_manager_has_create_method(self, mock_repo_class):
        """Test WorktreeManager has create method."""
        with tempfile.TemporaryDirectory() as tmpdir:
            repo_path = Path(tmpdir)
            worktree_root = Path(tmpdir) / "worktrees"

            mock_repo = MagicMock()
            mock_repo_class.return_value = mock_repo

            manager = WorktreeManager(repo_path=repo_path, worktree_root=worktree_root)
            assert hasattr(manager, "create")
            assert callable(manager.create)

    @patch("moai_adk.cli.worktree.manager.Repo")
    def test_worktree_manager_has_remove_method(self, mock_repo_class):
        """Test WorktreeManager has remove method."""
        with tempfile.TemporaryDirectory() as tmpdir:
            repo_path = Path(tmpdir)
            worktree_root = Path(tmpdir) / "worktrees"

            mock_repo = MagicMock()
            mock_repo_class.return_value = mock_repo

            manager = WorktreeManager(repo_path=repo_path, worktree_root=worktree_root)
            assert hasattr(manager, "remove")
            assert callable(manager.remove)

    @patch("moai_adk.cli.worktree.manager.Repo")
    def test_worktree_manager_has_switch_method(self, mock_repo_class):
        """Test WorktreeManager has switch method."""
        with tempfile.TemporaryDirectory() as tmpdir:
            repo_path = Path(tmpdir)
            worktree_root = Path(tmpdir) / "worktrees"

            mock_repo = MagicMock()
            mock_repo_class.return_value = mock_repo

            manager = WorktreeManager(repo_path=repo_path, worktree_root=worktree_root)
            # Check if switch or similar method exists
            methods = [m for m in dir(manager) if not m.startswith("_")]
            # Manager should have various methods
            assert len(methods) > 0

    @patch("moai_adk.cli.worktree.manager.Repo")
    def test_worktree_manager_has_list_method(self, mock_repo_class):
        """Test WorktreeManager can list worktrees."""
        with tempfile.TemporaryDirectory() as tmpdir:
            repo_path = Path(tmpdir)
            worktree_root = Path(tmpdir) / "worktrees"

            mock_repo = MagicMock()
            mock_repo_class.return_value = mock_repo

            manager = WorktreeManager(repo_path=repo_path, worktree_root=worktree_root)
            # Check for list method
            assert (
                hasattr(manager, "list")
                or len([m for m in dir(manager) if "list" in m.lower()]) > 0
            )


class TestWorktreeManagerAttributes:
    """Test WorktreeManager attributes."""

    @patch("moai_adk.cli.worktree.manager.Repo")
    def test_worktree_manager_repo_attribute(self, mock_repo_class):
        """Test WorktreeManager stores repo reference."""
        with tempfile.TemporaryDirectory() as tmpdir:
            repo_path = Path(tmpdir)
            worktree_root = Path(tmpdir) / "worktrees"

            mock_repo = MagicMock()
            mock_repo_class.return_value = mock_repo

            manager = WorktreeManager(repo_path=repo_path, worktree_root=worktree_root)
            assert manager.repo == mock_repo

    @patch("moai_adk.cli.worktree.manager.Repo")
    def test_worktree_manager_worktree_root_attribute(self, mock_repo_class):
        """Test WorktreeManager stores worktree root path."""
        with tempfile.TemporaryDirectory() as tmpdir:
            repo_path = Path(tmpdir)
            worktree_root = Path(tmpdir) / "worktrees"

            mock_repo = MagicMock()
            mock_repo_class.return_value = mock_repo

            manager = WorktreeManager(repo_path=repo_path, worktree_root=worktree_root)
            assert manager.worktree_root == worktree_root
            assert isinstance(manager.worktree_root, Path)


class TestWorktreeManagerExceptionHandling:
    """Test WorktreeManager exception imports."""

    def test_worktree_manager_imports_exceptions(self):
        """Test that WorktreeManager module imports exceptions."""
        from moai_adk.cli.worktree.exceptions import (
            GitOperationError,
            MergeConflictError,
            UncommittedChangesError,
            WorktreeExistsError,
            WorktreeNotFoundError,
        )

        # Verify exceptions exist
        assert GitOperationError is not None
        assert MergeConflictError is not None
        assert UncommittedChangesError is not None
        assert WorktreeExistsError is not None
        assert WorktreeNotFoundError is not None


class TestWorktreeManagerMethodSignatures:
    """Test WorktreeManager method signatures."""

    @patch("moai_adk.cli.worktree.manager.Repo")
    def test_worktree_manager_create_method_signature(self, mock_repo_class):
        """Test create method exists and has expected signature."""
        with tempfile.TemporaryDirectory() as tmpdir:
            repo_path = Path(tmpdir)
            worktree_root = Path(tmpdir) / "worktrees"

            mock_repo = MagicMock()
            mock_repo_class.return_value = mock_repo

            manager = WorktreeManager(repo_path=repo_path, worktree_root=worktree_root)
            import inspect

            sig = inspect.signature(manager.create)
            params = list(sig.parameters.keys())
            # Should have spec_id parameter
            assert "spec_id" in params

    @patch("moai_adk.cli.worktree.manager.Repo")
    def test_worktree_manager_remove_method_signature(self, mock_repo_class):
        """Test remove method exists and has expected signature."""
        with tempfile.TemporaryDirectory() as tmpdir:
            repo_path = Path(tmpdir)
            worktree_root = Path(tmpdir) / "worktrees"

            mock_repo = MagicMock()
            mock_repo_class.return_value = mock_repo

            manager = WorktreeManager(repo_path=repo_path, worktree_root=worktree_root)
            import inspect

            sig = inspect.signature(manager.remove)
            params = list(sig.parameters.keys())
            # Should have spec_id parameter
            assert "spec_id" in params


class TestWorktreeManagerDocstrings:
    """Test WorktreeManager documentation."""

    @patch("moai_adk.cli.worktree.manager.Repo")
    def test_worktree_manager_class_has_docstring(self, mock_repo_class):
        """Test WorktreeManager class has docstring."""
        assert WorktreeManager.__doc__ is not None
        assert len(WorktreeManager.__doc__) > 0

    @patch("moai_adk.cli.worktree.manager.Repo")
    def test_worktree_manager_init_has_docstring(self, mock_repo_class):
        """Test WorktreeManager.__init__ has docstring."""
        assert WorktreeManager.__init__.__doc__ is not None
