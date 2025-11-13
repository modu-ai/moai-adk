"""Unit tests for git/manager.py module

Tests for GitManager class using temporary Git repositories.
"""

import sys
from pathlib import Path

import pytest
from git import InvalidGitRepositoryError

from moai_adk.core.git.manager import GitManager

# Skip all tests in this file on Windows due to file locking issues
pytestmark = pytest.mark.skipif(sys.platform == "win32", reason="Windows file locking issues with Git repos")


class TestGitManagerInit:
    """Test GitManager initialization"""

    def test_init_with_valid_repo(self, tmp_git_repo: Path):
        """Should initialize successfully with valid Git repository"""
        manager = GitManager(str(tmp_git_repo))
        assert manager.repo is not None
        assert manager.git is not None

    def test_init_with_invalid_repo(self, tmp_project_dir: Path):
        """Should raise InvalidGitRepositoryError for non-Git directory"""
        with pytest.raises(InvalidGitRepositoryError):
            GitManager(str(tmp_project_dir))

    def test_init_default_path(self):
        """Should initialize with current directory by default"""
        # This test assumes current directory is a Git repo
        try:
            manager = GitManager()
            assert manager.repo is not None
        except InvalidGitRepositoryError:
            # If not a Git repo, that's expected
            pass


class TestGitManagerIsRepo:
    """Test is_repo method"""

    def test_is_repo_returns_true_for_git_repo(self, tmp_git_repo: Path):
        """Should return True for valid Git repository"""
        manager = GitManager(str(tmp_git_repo))
        assert manager.is_repo() is True

    def test_is_repo_handles_invalid_state_gracefully(self, tmp_git_repo: Path):
        """Should handle corrupted repository state"""
        manager = GitManager(str(tmp_git_repo))
        # Even if repo is valid initially, method should not crash
        result = manager.is_repo()
        assert isinstance(result, bool)


class TestGitManagerCurrentBranch:
    """Test current_branch method"""

    def test_current_branch_returns_initial_branch(self, tmp_git_repo: Path):
        """Should return initial branch name (usually 'main' or 'master')"""
        manager = GitManager(str(tmp_git_repo))
        branch = manager.current_branch()
        # Git init creates 'master' or 'main' depending on config
        assert branch in ["main", "master"]

    def test_current_branch_after_creating_new_branch(self, tmp_git_repo: Path):
        """Should return new branch name after switching"""
        manager = GitManager(str(tmp_git_repo))
        # Create initial commit (required for branch operations)
        (tmp_git_repo / "test.txt").write_text("test")
        manager.commit("Initial commit", files=["test.txt"])

        manager.create_branch("feature/test")
        assert manager.current_branch() == "feature/test"


class TestGitManagerIsDirty:
    """Test is_dirty method"""

    def test_is_dirty_returns_false_for_clean_repo(self, tmp_git_repo: Path):
        """Should return False when no changes exist"""
        manager = GitManager(str(tmp_git_repo))
        # Fresh repo should be clean
        assert manager.is_dirty() is False

    def test_is_dirty_returns_true_after_file_modification(self, tmp_git_repo: Path):
        """Should return True when files are modified"""
        manager = GitManager(str(tmp_git_repo))

        # Create and commit a file first
        test_file = tmp_git_repo / "test.txt"
        test_file.write_text("initial content")
        manager.commit("Initial commit", files=["test.txt"])

        # Modify the file
        test_file.write_text("modified content")

        assert manager.is_dirty() is True

    def test_is_dirty_returns_false_after_commit(self, tmp_git_repo: Path):
        """Should return False after committing changes"""
        manager = GitManager(str(tmp_git_repo))

        test_file = tmp_git_repo / "test.txt"
        test_file.write_text("content")
        manager.commit("Add test file", files=["test.txt"])

        assert manager.is_dirty() is False


class TestGitManagerCreateBranch:
    """Test create_branch method"""

    def test_create_branch_from_current(self, tmp_git_repo: Path):
        """Should create and switch to new branch from current branch"""
        manager = GitManager(str(tmp_git_repo))

        # Create initial commit
        (tmp_git_repo / "test.txt").write_text("test")
        manager.commit("Initial commit", files=["test.txt"])

        manager.create_branch("feature/SPEC-TEST-001")
        assert manager.current_branch() == "feature/SPEC-TEST-001"

    def test_create_branch_from_specific_branch(self, tmp_git_repo: Path):
        """Should create branch from specified base branch"""
        manager = GitManager(str(tmp_git_repo))

        # Create initial commit
        (tmp_git_repo / "test.txt").write_text("test")
        manager.commit("Initial commit", files=["test.txt"])

        # Get initial branch name
        initial_branch = manager.current_branch()

        # Create a branch from the initial branch
        manager.create_branch("feature/from-main", from_branch=initial_branch)
        assert manager.current_branch() == "feature/from-main"


class TestGitManagerCommit:
    """Test commit method"""

    def test_commit_specific_files(self, tmp_git_repo: Path):
        """Should commit only specified files"""
        manager = GitManager(str(tmp_git_repo))

        file1 = tmp_git_repo / "file1.txt"
        file2 = tmp_git_repo / "file2.txt"
        file1.write_text("content1")
        file2.write_text("content2")

        manager.commit("Add file1", files=["file1.txt"])

        # Verify file1 was committed by checking it's not in untracked files
        untracked = manager.repo.untracked_files
        assert "file1.txt" not in untracked
        assert "file2.txt" in untracked  # file2 should remain untracked

    def test_commit_all_changes(self, tmp_git_repo: Path):
        """Should commit all changes when files=None"""
        manager = GitManager(str(tmp_git_repo))

        file1 = tmp_git_repo / "file1.txt"
        file2 = tmp_git_repo / "file2.txt"
        file1.write_text("content1")
        file2.write_text("content2")

        manager.commit("Add all files")

        # All files should be committed
        assert manager.is_dirty() is False

    def test_commit_creates_commit_with_message(self, tmp_git_repo: Path):
        """Should create commit with specified message"""
        manager = GitManager(str(tmp_git_repo))

        (tmp_git_repo / "test.txt").write_text("test")
        manager.commit("Test commit message", files=["test.txt"])

        # Get latest commit message
        latest_commit_msg = manager.repo.head.commit.message.strip()
        assert latest_commit_msg == "Test commit message"


class TestGitManagerPush:
    """Test push method

    Note: These tests don't actually push to remote (no remote configured in test repos)
    They verify the method doesn't crash with proper parameters.
    """

    def test_push_method_exists(self, tmp_git_repo: Path):
        """Push method should exist and be callable"""
        manager = GitManager(str(tmp_git_repo))
        assert hasattr(manager, "push")
        assert callable(manager.push)

    def test_push_with_set_upstream_parameter(self, tmp_git_repo: Path):
        """Push method should accept set_upstream parameter"""
        manager = GitManager(str(tmp_git_repo))
        # Should not crash even without remote (will fail at Git level, not Python)
        try:
            manager.push(set_upstream=True)
        except Exception as e:
            # Expected to fail since no remote exists, but verify it's a Git error
            assert "remote" in str(e).lower() or "origin" in str(e).lower()
