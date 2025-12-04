"""
Comprehensive executable tests for Worktree CLI commands.

These tests exercise actual code paths including:
- Exception handling with correct signatures
- Git operation errors
"""

import pytest
from pathlib import Path
from unittest.mock import MagicMock, patch, call

from moai_adk.cli.worktree.exceptions import (
    GitOperationError,
    MergeConflictError,
    UncommittedChangesError,
    WorktreeExistsError,
    WorktreeNotFoundError,
)


class TestGitOperationError:
    """Test GitOperationError exception."""

    def test_create_git_operation_error(self):
        """Test creating GitOperationError."""
        error = GitOperationError("Git command failed")
        assert "Git command failed" in str(error)
        assert isinstance(error, Exception)

    def test_git_operation_error_wraps_message(self):
        """Test Git operation error wraps original message."""
        error = GitOperationError("permission denied")
        assert "Git operation failed" in str(error)
        assert "permission denied" in str(error)

    def test_raise_git_operation_error(self):
        """Test raising GitOperationError."""
        with pytest.raises(GitOperationError):
            raise GitOperationError("git rebase failed")


class TestWorktreeExistsError:
    """Test WorktreeExistsError exception."""

    def test_create_worktree_exists_error(self):
        """Test creating WorktreeExistsError."""
        error = WorktreeExistsError("spec-001", Path("/tmp/worktrees/spec-001"))
        assert "spec-001" in str(error)
        assert isinstance(error, Exception)

    def test_worktree_exists_error_attributes(self):
        """Test WorktreeExistsError stores attributes."""
        path = Path("/home/user/worktrees/feature")
        error = WorktreeExistsError("feature-spec", path)
        assert error.spec_id == "feature-spec"
        assert error.path == path

    def test_raise_worktree_exists_error(self):
        """Test raising WorktreeExistsError."""
        with pytest.raises(WorktreeExistsError):
            raise WorktreeExistsError("spec-001", Path("/tmp"))


class TestMergeConflictError:
    """Test MergeConflictError exception."""

    def test_create_merge_conflict_error(self):
        """Test creating MergeConflictError."""
        files = ["file1.py", "file2.py"]
        error = MergeConflictError("spec-001", files)
        assert "spec-001" in str(error)
        assert "file1.py" in str(error)
        assert isinstance(error, Exception)

    def test_merge_conflict_error_attributes(self):
        """Test MergeConflictError stores attributes."""
        files = ["a.py", "b.py", "c.py"]
        error = MergeConflictError("spec-merge", files)
        assert error.spec_id == "spec-merge"
        assert error.conflicted_files == files

    def test_merge_conflict_error_multiple_files(self):
        """Test merge conflict with multiple files."""
        files = ["src/main.py", "src/utils.py", "tests/test_main.py"]
        error = MergeConflictError("spec-002", files)
        for file in files:
            assert file in str(error)

    def test_raise_merge_conflict_error(self):
        """Test raising MergeConflictError."""
        with pytest.raises(MergeConflictError):
            raise MergeConflictError("spec-001", ["file.py"])


class TestUncommittedChangesError:
    """Test UncommittedChangesError exception."""

    def test_create_uncommitted_changes_error(self):
        """Test creating UncommittedChangesError."""
        error = UncommittedChangesError("spec-001")
        assert "spec-001" in str(error)
        assert "uncommitted" in str(error).lower()
        assert isinstance(error, Exception)

    def test_uncommitted_changes_error_attributes(self):
        """Test UncommittedChangesError stores spec_id."""
        error = UncommittedChangesError("spec-uncommitted")
        assert error.spec_id == "spec-uncommitted"

    def test_uncommitted_changes_error_message_format(self):
        """Test uncommitted changes error message format."""
        error = UncommittedChangesError("development")
        message = str(error)
        assert "development" in message
        assert "uncommitted" in message.lower()
        assert "--force" in message

    def test_raise_uncommitted_changes_error(self):
        """Test raising UncommittedChangesError."""
        with pytest.raises(UncommittedChangesError):
            raise UncommittedChangesError("spec-001")


class TestWorktreeNotFoundError:
    """Test WorktreeNotFoundError exception."""

    def test_create_worktree_not_found_error(self):
        """Test creating WorktreeNotFoundError."""
        error = WorktreeNotFoundError("spec-001")
        assert "spec-001" in str(error)
        assert isinstance(error, Exception)

    def test_worktree_not_found_error_attributes(self):
        """Test WorktreeNotFoundError stores spec_id."""
        error = WorktreeNotFoundError("spec-missing")
        assert error.spec_id == "spec-missing"

    def test_raise_worktree_not_found_error(self):
        """Test raising WorktreeNotFoundError."""
        with pytest.raises(WorktreeNotFoundError):
            raise WorktreeNotFoundError("spec-not-found")


class TestExceptionHierarchy:
    """Test exception inheritance hierarchy."""

    def test_all_are_exceptions(self):
        """Test all custom exceptions inherit from Exception."""
        exceptions = [
            GitOperationError("test"),
            WorktreeExistsError("spec", Path(".")),
            MergeConflictError("spec", []),
            UncommittedChangesError("spec"),
            WorktreeNotFoundError("spec"),
        ]

        for error in exceptions:
            assert isinstance(error, Exception)


class TestExceptionCatching:
    """Test catching exceptions."""

    def test_catch_git_operation_error(self):
        """Test catching GitOperationError."""
        try:
            raise GitOperationError("test error")
        except GitOperationError as e:
            assert "test error" in str(e)

    def test_catch_worktree_exists_error(self):
        """Test catching WorktreeExistsError."""
        try:
            raise WorktreeExistsError("spec-x", Path("/tmp"))
        except WorktreeExistsError as e:
            assert "spec-x" in str(e)

    def test_catch_merge_conflict_error(self):
        """Test catching MergeConflictError."""
        try:
            raise MergeConflictError("spec-y", ["file.txt"])
        except MergeConflictError as e:
            assert "spec-y" in str(e)

    def test_catch_uncommitted_changes_error(self):
        """Test catching UncommittedChangesError."""
        try:
            raise UncommittedChangesError("spec-z")
        except UncommittedChangesError as e:
            assert "spec-z" in str(e)

    def test_catch_worktree_not_found_error(self):
        """Test catching WorktreeNotFoundError."""
        try:
            raise WorktreeNotFoundError("spec-404")
        except WorktreeNotFoundError as e:
            assert "spec-404" in str(e)

    def test_catch_as_generic_exception(self):
        """Test catching all as generic Exception."""
        errors = [
            GitOperationError("test"),
            WorktreeExistsError("spec", Path(".")),
            MergeConflictError("spec", ["f"]),
            UncommittedChangesError("spec"),
            WorktreeNotFoundError("spec"),
        ]

        caught = 0
        for error in errors:
            try:
                raise error
            except Exception:
                caught += 1

        assert caught == 5


class TestErrorScenarios:
    """Test realistic error scenarios."""

    def test_git_rebase_failure(self):
        """Test git rebase failure scenario."""
        try:
            raise GitOperationError("fatal: Could not apply HEAD")
        except GitOperationError as e:
            assert "fatal" in str(e).lower() or "could not apply" in str(e).lower()

    def test_worktree_already_exists_scenario(self):
        """Test worktree creation when it already exists."""
        spec_id = "feature-x"
        path = Path(f"/worktrees/{spec_id}")
        try:
            raise WorktreeExistsError(spec_id, path)
        except WorktreeExistsError as e:
            assert spec_id in str(e)

    def test_merge_with_multiple_conflicts(self):
        """Test merge with multiple conflict files."""
        files = ["a.py", "b.py", "c.py"]
        try:
            raise MergeConflictError("spec-merge", files)
        except MergeConflictError as e:
            for f in files:
                assert f in str(e)

    def test_remove_worktree_with_changes(self):
        """Test removing worktree with uncommitted changes."""
        try:
            raise UncommittedChangesError("spec-dirty")
        except UncommittedChangesError as e:
            assert "force" in str(e).lower()

    def test_sync_missing_worktree(self):
        """Test syncing missing worktree."""
        try:
            raise WorktreeNotFoundError("spec-missing")
        except WorktreeNotFoundError as e:
            assert "missing" in str(e).lower() or "not found" in str(e).lower()


class TestExceptionContextPreservation:
    """Test exception context preservation."""

    def test_nested_exception_context(self):
        """Test nested exception maintains context."""
        try:
            try:
                raise ValueError("Original error")
            except ValueError as e:
                raise GitOperationError(f"Operation failed: {e}") from e
        except GitOperationError as e:
            assert "Original error" in str(e)


class TestExceptionMessages:
    """Test exception message formatting."""

    def test_git_operation_error_message(self):
        """Test git operation error message."""
        error = GitOperationError("command failed")
        msg = str(error)
        assert isinstance(msg, str)
        assert len(msg) > 0

    def test_worktree_exists_error_message(self):
        """Test worktree exists error message."""
        error = WorktreeExistsError("spec-1", Path("/tmp"))
        msg = str(error)
        assert isinstance(msg, str)
        assert "spec-1" in msg

    def test_merge_conflict_error_message(self):
        """Test merge conflict error message."""
        error = MergeConflictError("spec-1", ["a.py"])
        msg = str(error)
        assert isinstance(msg, str)
        assert "a.py" in msg

    def test_uncommitted_changes_error_message(self):
        """Test uncommitted changes error message."""
        error = UncommittedChangesError("spec-1")
        msg = str(error)
        assert isinstance(msg, str)
        assert "spec-1" in msg

    def test_worktree_not_found_error_message(self):
        """Test worktree not found error message."""
        error = WorktreeNotFoundError("spec-1")
        msg = str(error)
        assert isinstance(msg, str)
        assert "spec-1" in msg
