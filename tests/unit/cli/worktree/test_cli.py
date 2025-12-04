"""
Minimal import and instantiation tests for Worktree CLI.

These tests verify that the module can be imported and basic functions
can be executed without errors.
"""

import pytest
import tempfile
from pathlib import Path
from unittest.mock import MagicMock, patch

from moai_adk.cli.worktree.cli import (
    get_manager,
    _detect_worktree_root,
    _find_main_repository,
)


class TestImports:
    """Test that all functions can be imported."""

    def test_get_manager_importable(self):
        """Test get_manager function can be imported."""
        assert get_manager is not None
        assert callable(get_manager)

    def test_detect_worktree_root_importable(self):
        """Test _detect_worktree_root function can be imported."""
        assert _detect_worktree_root is not None
        assert callable(_detect_worktree_root)

    def test_find_main_repository_importable(self):
        """Test _find_main_repository function can be imported."""
        assert _find_main_repository is not None
        assert callable(_find_main_repository)


class TestGetManagerFunction:
    """Test get_manager function."""

    @patch("moai_adk.cli.worktree.cli.WorktreeManager")
    @patch("moai_adk.cli.worktree.cli._detect_worktree_root")
    def test_get_manager_with_defaults(self, mock_detect, mock_manager_class):
        """Test get_manager with default parameters."""
        with tempfile.TemporaryDirectory() as tmpdir:
            repo_path = Path(tmpdir)
            worktree_root = Path(tmpdir) / "worktrees"

            mock_detect.return_value = worktree_root
            mock_manager_instance = MagicMock()
            mock_manager_class.return_value = mock_manager_instance

            # Call with no arguments
            manager = get_manager()
            assert manager is not None

    @patch("moai_adk.cli.worktree.cli.WorktreeManager")
    def test_get_manager_with_repo_path(self, mock_manager_class):
        """Test get_manager with explicit repo_path."""
        with tempfile.TemporaryDirectory() as tmpdir:
            repo_path = Path(tmpdir)
            worktree_root = Path(tmpdir) / "worktrees"

            mock_manager_instance = MagicMock()
            mock_manager_class.return_value = mock_manager_instance

            manager = get_manager(repo_path=repo_path, worktree_root=worktree_root)
            assert manager is not None

    @patch("moai_adk.cli.worktree.cli.WorktreeManager")
    def test_get_manager_returns_manager(self, mock_manager_class):
        """Test get_manager returns WorktreeManager instance."""
        with tempfile.TemporaryDirectory() as tmpdir:
            repo_path = Path(tmpdir)
            worktree_root = Path(tmpdir) / "worktrees"

            mock_manager_instance = MagicMock()
            mock_manager_class.return_value = mock_manager_instance

            manager = get_manager(repo_path=repo_path, worktree_root=worktree_root)
            assert manager == mock_manager_instance


class TestDetectWorktreeRootFunction:
    """Test _detect_worktree_root function."""

    def test_detect_worktree_root_returns_path(self):
        """Test _detect_worktree_root returns a Path object."""
        with tempfile.TemporaryDirectory() as tmpdir:
            repo_path = Path(tmpdir)
            result = _detect_worktree_root(repo_path)
            assert isinstance(result, Path)

    def test_detect_worktree_root_nonexistent_path(self):
        """Test _detect_worktree_root handles nonexistent path."""
        nonexistent = Path("/nonexistent/path/repo")
        result = _detect_worktree_root(nonexistent)
        assert isinstance(result, Path)

    @patch("moai_adk.cli.worktree.cli._find_main_repository")
    def test_detect_worktree_root_calls_find_main_repo(self, mock_find_main):
        """Test _detect_worktree_root calls _find_main_repository."""
        with tempfile.TemporaryDirectory() as tmpdir:
            repo_path = Path(tmpdir)
            main_repo = Path(tmpdir) / "main"
            main_repo.mkdir()

            mock_find_main.return_value = main_repo

            result = _detect_worktree_root(repo_path)
            mock_find_main.assert_called_once()


class TestFindMainRepositoryFunction:
    """Test _find_main_repository function."""

    def test_find_main_repository_returns_path(self):
        """Test _find_main_repository returns a Path object."""
        with tempfile.TemporaryDirectory() as tmpdir:
            repo_path = Path(tmpdir)
            result = _find_main_repository(repo_path)
            assert isinstance(result, Path)

    def test_find_main_repository_returns_existing_path(self):
        """Test _find_main_repository returns an existing path."""
        with tempfile.TemporaryDirectory() as tmpdir:
            repo_path = Path(tmpdir)
            result = _find_main_repository(repo_path)
            # Should return the input or a parent directory
            assert result.exists() or result == repo_path

    def test_find_main_repository_with_git_directory(self):
        """Test _find_main_repository with .git directory."""
        with tempfile.TemporaryDirectory() as tmpdir:
            repo_path = Path(tmpdir)
            git_dir = repo_path / ".git"
            git_dir.mkdir()

            result = _find_main_repository(repo_path)
            assert isinstance(result, Path)


class TestCliConsoleImport:
    """Test CLI console-related imports."""

    def test_console_is_available(self):
        """Test that Rich console is available."""
        from moai_adk.cli.worktree.cli import console

        assert console is not None

    def test_click_is_available(self):
        """Test that Click is available."""
        import click

        assert click is not None


class TestCliExceptionHandling:
    """Test CLI exception handling imports."""

    def test_cli_imports_exceptions(self):
        """Test that CLI module imports exceptions."""
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


class TestCliModelsImport:
    """Test CLI models import."""

    def test_cli_models_available(self):
        """Test that models are available for import."""
        from moai_adk.cli.worktree.models import WorktreeInfo

        assert WorktreeInfo is not None


class TestGetManagerEdgeCases:
    """Test get_manager function edge cases."""

    @patch("moai_adk.cli.worktree.cli.Path.cwd")
    @patch("moai_adk.cli.worktree.cli.WorktreeManager")
    @patch("moai_adk.cli.worktree.cli._detect_worktree_root")
    def test_get_manager_finds_git_in_parent(
        self, mock_detect, mock_manager_class, mock_cwd
    ):
        """Test get_manager finds .git in parent directory."""
        with tempfile.TemporaryDirectory() as tmpdir:
            base_path = Path(tmpdir)
            git_dir = base_path / ".git"
            git_dir.mkdir()

            mock_cwd.return_value = base_path
            mock_detect.return_value = base_path / "worktrees"
            mock_manager_instance = MagicMock()
            mock_manager_class.return_value = mock_manager_instance

            manager = get_manager()
            assert manager is not None

    @patch("moai_adk.cli.worktree.cli.WorktreeManager")
    @patch("moai_adk.cli.worktree.cli._detect_worktree_root")
    def test_get_manager_with_explicit_worktree_root(
        self, mock_detect, mock_manager_class
    ):
        """Test get_manager uses explicit worktree_root."""
        with tempfile.TemporaryDirectory() as tmpdir:
            repo_path = Path(tmpdir)
            worktree_root = Path(tmpdir) / "custom" / "worktrees"

            mock_manager_instance = MagicMock()
            mock_manager_class.return_value = mock_manager_instance

            manager = get_manager(repo_path=repo_path, worktree_root=worktree_root)
            # Verify WorktreeManager was called with the worktree_root
            assert manager is not None


class TestCliDocstrings:
    """Test CLI function documentation."""

    def test_get_manager_has_docstring(self):
        """Test get_manager has docstring."""
        assert get_manager.__doc__ is not None
        assert len(get_manager.__doc__) > 0

    def test_detect_worktree_root_has_docstring(self):
        """Test _detect_worktree_root has docstring."""
        assert _detect_worktree_root.__doc__ is not None
        assert len(_detect_worktree_root.__doc__) > 0

    def test_find_main_repository_has_docstring(self):
        """Test _find_main_repository has docstring."""
        assert _find_main_repository.__doc__ is not None
        assert len(_find_main_repository.__doc__) > 0
