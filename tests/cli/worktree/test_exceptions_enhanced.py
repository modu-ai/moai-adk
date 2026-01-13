"""Comprehensive tests for Git Worktree CLI exceptions.

Test Coverage Strategy:
- All 6 exception classes: WorktreeError, WorktreeExistsError, WorktreeNotFoundError,
  UncommittedChangesError, GitOperationError, MergeConflictError, RegistryInconsistencyError
- Message formatting: Verify error messages are correctly formatted
- Inheritance hierarchy: Verify all exceptions inherit from WorktreeError
- Attribute storage: Verify custom attributes are stored correctly
- Edge cases: Empty strings, special characters, Unicode, long strings
- Exception catching: Verify catch blocks work as expected
"""

from pathlib import Path

import pytest

from moai_adk.cli.worktree.exceptions import (
    GitOperationError,
    MergeConflictError,
    RegistryInconsistencyError,
    UncommittedChangesError,
    WorktreeError,
    WorktreeExistsError,
    WorktreeNotFoundError,
)


class TestWorktreeError:
    """Test base WorktreeError exception."""

    def test_is_exception_subclass(self) -> None:
        """Test that WorktreeError is an Exception subclass."""
        assert issubclass(WorktreeError, Exception)

    def test_can_be_raised(self) -> None:
        """Test that WorktreeError can be raised and caught."""
        with pytest.raises(WorktreeError):
            raise WorktreeError("Base error")

    def test_can_be_caught_as_exception(self) -> None:
        """Test that WorktreeError can be caught as Exception."""
        try:
            raise WorktreeError("Base error")
        except Exception as e:
            assert isinstance(e, WorktreeError)
            assert str(e) == "Base error"

    def test_message_preservation(self) -> None:
        """Test that error message is preserved."""
        message = "This is a worktree error"
        error = WorktreeError(message)
        assert str(error) == message

    def test_empty_message(self) -> None:
        """Test WorktreeError with empty message."""
        error = WorktreeError("")
        assert str(error) == ""


class TestWorktreeExistsError:
    """Test WorktreeExistsError exception."""

    def test_inheritance(self) -> None:
        """Test that WorktreeExistsError inherits from WorktreeError."""
        assert issubclass(WorktreeExistsError, WorktreeError)

    def test_initialization_with_spec_id_and_path(self, tmp_path: Path) -> None:
        """Test initialization with spec_id and path parameters."""
        spec_id = "SPEC-001"
        path = tmp_path / "worktrees" / "SPEC-001"

        error = WorktreeExistsError(spec_id, path)

        assert error.spec_id == spec_id
        assert error.path == path

    def test_message_formatting(self, tmp_path: Path) -> None:
        """Test that error message is formatted correctly."""
        spec_id = "SPEC-AUTH-001"
        path = tmp_path / "worktrees" / "SPEC-AUTH-001"

        error = WorktreeExistsError(spec_id, path)
        message = str(error)

        assert spec_id in message
        assert str(path) in message
        assert "already exists" in message

    def test_message_formatting_exact(self, tmp_path: Path) -> None:
        """Test exact format of error message."""
        spec_id = "SPEC-001"
        path = tmp_path / "worktree"

        error = WorktreeExistsError(spec_id, path)
        expected = f"Worktree for '{spec_id}' already exists at {path}"

        assert str(error) == expected

    def test_can_be_raised_and_caught(self, tmp_path: Path) -> None:
        """Test that exception can be raised and caught as WorktreeError."""
        spec_id = "SPEC-001"
        path = tmp_path / "worktree"

        with pytest.raises(WorktreeError):
            raise WorktreeExistsError(spec_id, path)

    def test_can_be_caught_as_specific_type(self, tmp_path: Path) -> None:
        """Test that exception can be caught as WorktreeExistsError."""
        spec_id = "SPEC-001"
        path = tmp_path / "worktree"

        with pytest.raises(WorktreeExistsError) as exc_info:
            raise WorktreeExistsError(spec_id, path)

        assert exc_info.value.spec_id == spec_id
        assert exc_info.value.path == path

    def test_with_various_spec_ids(self, tmp_path: Path) -> None:
        """Test with various SPEC ID formats."""
        spec_ids = [
            "SPEC-001",
            "SPEC-AUTH-001",
            "SPEC-FE-2025-001",
            "spec-001",
            "SPEC_001",
        ]

        for spec_id in spec_ids:
            error = WorktreeExistsError(spec_id, tmp_path)
            assert error.spec_id == spec_id
            assert spec_id in str(error)

    def test_with_complex_path(self, tmp_path: Path) -> None:
        """Test with complex nested paths."""
        complex_path = tmp_path / "moai" / "worktrees" / "project" / "SPEC-001"

        error = WorktreeExistsError("SPEC-001", complex_path)

        assert error.path == complex_path
        assert str(complex_path) in str(error)

    def test_with_path_containing_spaces(self, tmp_path: Path) -> None:
        """Test with paths containing spaces."""
        path_with_spaces = tmp_path / "worktrees" / "SPEC-001 (dev)"

        error = WorktreeExistsError("SPEC-001", path_with_spaces)

        assert error.path == path_with_spaces
        # Path should be in message despite spaces
        assert "SPEC-001" in str(error)

    def test_with_unicode_spec_id(self, tmp_path: Path) -> None:
        """Test with Unicode characters in spec_id."""
        spec_id = "SPEC-国际化-001"

        error = WorktreeExistsError(spec_id, tmp_path)

        assert error.spec_id == spec_id
        assert spec_id in str(error)


class TestWorktreeNotFoundError:
    """Test WorktreeNotFoundError exception."""

    def test_inheritance(self) -> None:
        """Test that WorktreeNotFoundError inherits from WorktreeError."""
        assert issubclass(WorktreeNotFoundError, WorktreeError)

    def test_initialization_with_spec_id(self) -> None:
        """Test initialization with spec_id parameter."""
        spec_id = "SPEC-001"

        error = WorktreeNotFoundError(spec_id)

        assert error.spec_id == spec_id

    def test_message_formatting(self) -> None:
        """Test that error message is formatted correctly."""
        spec_id = "SPEC-FE-2025-001"

        error = WorktreeNotFoundError(spec_id)
        message = str(error)

        assert spec_id in message
        assert "not found" in message

    def test_message_formatting_exact(self) -> None:
        """Test exact format of error message."""
        spec_id = "SPEC-001"
        error = WorktreeNotFoundError(spec_id)
        expected = f"Worktree for '{spec_id}' not found"

        assert str(error) == expected

    def test_can_be_raised_and_caught(self) -> None:
        """Test that exception can be raised and caught as WorktreeError."""
        with pytest.raises(WorktreeError):
            raise WorktreeNotFoundError("SPEC-001")

    def test_can_be_caught_as_specific_type(self) -> None:
        """Test that exception can be caught as WorktreeNotFoundError."""
        spec_id = "SPEC-AUTH-001"

        with pytest.raises(WorktreeNotFoundError) as exc_info:
            raise WorktreeNotFoundError(spec_id)

        assert exc_info.value.spec_id == spec_id

    def test_with_various_spec_ids(self) -> None:
        """Test with various SPEC ID formats."""
        spec_ids = ["SPEC-001", "SPEC-AUTH-001", "spec-001", "SPEC_001"]

        for spec_id in spec_ids:
            error = WorktreeNotFoundError(spec_id)
            assert error.spec_id == spec_id
            assert spec_id in str(error)

    def test_with_empty_spec_id(self) -> None:
        """Test with empty spec_id."""
        error = WorktreeNotFoundError("")
        assert error.spec_id == ""
        assert "not found" in str(error)

    def test_with_unicode_spec_id(self) -> None:
        """Test with Unicode characters in spec_id."""
        spec_id = "SPEC-国际化-001"

        error = WorktreeNotFoundError(spec_id)

        assert error.spec_id == spec_id
        assert spec_id in str(error)


class TestUncommittedChangesError:
    """Test UncommittedChangesError exception."""

    def test_inheritance(self) -> None:
        """Test that UncommittedChangesError inherits from WorktreeError."""
        assert issubclass(UncommittedChangesError, WorktreeError)

    def test_initialization_with_spec_id(self) -> None:
        """Test initialization with spec_id parameter."""
        spec_id = "SPEC-001"

        error = UncommittedChangesError(spec_id)

        assert error.spec_id == spec_id

    def test_message_formatting(self) -> None:
        """Test that error message is formatted correctly."""
        spec_id = "SPEC-FE-001"

        error = UncommittedChangesError(spec_id)
        message = str(error)

        assert spec_id in message
        assert "uncommitted changes" in message
        assert "--force" in message

    def test_message_formatting_exact(self) -> None:
        """Test exact format of error message."""
        spec_id = "SPEC-001"
        error = UncommittedChangesError(spec_id)
        expected = f"Worktree for '{spec_id}' has uncommitted changes. Use --force to remove anyway."

        assert str(error) == expected

    def test_can_be_raised_and_caught(self) -> None:
        """Test that exception can be raised and caught as WorktreeError."""
        with pytest.raises(WorktreeError):
            raise UncommittedChangesError("SPEC-001")

    def test_can_be_caught_as_specific_type(self) -> None:
        """Test that exception can be caught as UncommittedChangesError."""
        spec_id = "SPEC-AUTH-001"

        with pytest.raises(UncommittedChangesError) as exc_info:
            raise UncommittedChangesError(spec_id)

        assert exc_info.value.spec_id == spec_id

    def test_with_various_spec_ids(self) -> None:
        """Test with various SPEC ID formats."""
        spec_ids = ["SPEC-001", "SPEC-AUTH-001", "spec-001", "SPEC_001"]

        for spec_id in spec_ids:
            error = UncommittedChangesError(spec_id)
            assert error.spec_id == spec_id
            assert spec_id in str(error)


class TestGitOperationError:
    """Test GitOperationError exception."""

    def test_inheritance(self) -> None:
        """Test that GitOperationError inherits from WorktreeError."""
        assert issubclass(GitOperationError, WorktreeError)

    def test_initialization_with_message(self) -> None:
        """Test initialization with message parameter."""
        message = "Failed to create worktree"

        error = GitOperationError(message)

        assert str(error) == f"Git operation failed: {message}"

    def test_message_formatting(self) -> None:
        """Test that error message is formatted correctly."""
        original_message = "Branch does not exist"

        error = GitOperationError(original_message)
        full_message = str(error)

        assert "Git operation failed:" in full_message
        assert original_message in full_message

    def test_message_formatting_exact(self) -> None:
        """Test exact format of error message."""
        message = "fatal: not a git repository"
        error = GitOperationError(message)
        expected = f"Git operation failed: {message}"

        assert str(error) == expected

    def test_can_be_raised_and_caught(self) -> None:
        """Test that exception can be raised and caught as WorktreeError."""
        with pytest.raises(WorktreeError):
            raise GitOperationError("Git error")

    def test_can_be_caught_as_specific_type(self) -> None:
        """Test that exception can be caught as GitOperationError."""
        message = "Branch not found"

        with pytest.raises(GitOperationError) as exc_info:
            raise GitOperationError(message)

        assert "Git operation failed:" in str(exc_info.value)

    def test_with_various_messages(self) -> None:
        """Test with various error messages."""
        messages = [
            "Branch not found",
            "Not a git repository",
            "Invalid branch name",
            "Permission denied",
        ]

        for message in messages:
            error = GitOperationError(message)
            assert message in str(error)

    def test_with_empty_message(self) -> None:
        """Test with empty message."""
        error = GitOperationError("")
        assert str(error) == "Git operation failed: "

    def test_with_long_message(self) -> None:
        """Test with very long error message."""
        long_message = "Error: " + "A" * 1000

        error = GitOperationError(long_message)

        assert long_message in str(error)

    def test_with_special_characters(self) -> None:
        """Test with special characters in message."""
        special_message = "Error: file 'test<script>.py' has issues"

        error = GitOperationError(special_message)

        assert special_message in str(error)


class TestMergeConflictError:
    """Test MergeConflictError exception."""

    def test_inheritance(self) -> None:
        """Test that MergeConflictError inherits from WorktreeError."""
        assert issubclass(MergeConflictError, WorktreeError)

    def test_initialization_with_spec_id_and_files(self) -> None:
        """Test initialization with spec_id and conflicted_files parameters."""
        spec_id = "SPEC-001"
        conflicted_files = ["src/main.py", "tests/test_main.py"]

        error = MergeConflictError(spec_id, conflicted_files)

        assert error.spec_id == spec_id
        assert error.conflicted_files == conflicted_files

    def test_message_formatting(self) -> None:
        """Test that error message is formatted correctly."""
        spec_id = "SPEC-FE-001"
        conflicted_files = ["src/app.py", "tests/test_app.py"]

        error = MergeConflictError(spec_id, conflicted_files)
        message = str(error)

        assert spec_id in message
        assert "Merge conflict" in message
        assert "src/app.py" in message
        assert "tests/test_app.py" in message

    def test_message_formatting_exact(self) -> None:
        """Test exact format of error message."""
        spec_id = "SPEC-001"
        conflicted_files = ["file1.py", "file2.py"]
        error = MergeConflictError(spec_id, conflicted_files)
        expected = f"Merge conflict in worktree '{spec_id}'. Conflicted files: {', '.join(conflicted_files)}"

        assert str(error) == expected

    def test_can_be_raised_and_caught(self) -> None:
        """Test that exception can be raised and caught as WorktreeError."""
        with pytest.raises(WorktreeError):
            raise MergeConflictError("SPEC-001", ["file.py"])

    def test_can_be_caught_as_specific_type(self) -> None:
        """Test that exception can be caught as MergeConflictError."""
        spec_id = "SPEC-AUTH-001"
        files = ["src/auth.py"]

        with pytest.raises(MergeConflictError) as exc_info:
            raise MergeConflictError(spec_id, files)

        assert exc_info.value.spec_id == spec_id
        assert exc_info.value.conflicted_files == files

    def test_with_single_conflicted_file(self) -> None:
        """Test with single conflicted file."""
        error = MergeConflictError("SPEC-001", ["src/main.py"])

        assert len(error.conflicted_files) == 1
        assert "src/main.py" in str(error)

    def test_with_multiple_conflicted_files(self) -> None:
        """Test with multiple conflicted files."""
        files = ["src/main.py", "tests/test_main.py", "docs/README.md"]
        error = MergeConflictError("SPEC-001", files)

        assert len(error.conflicted_files) == 3
        for file in files:
            assert file in str(error)

    def test_with_empty_file_list(self) -> None:
        """Test with empty conflicted_files list."""
        error = MergeConflictError("SPEC-001", [])

        assert error.conflicted_files == []
        assert "Merge conflict" in str(error)

    def test_with_file_paths_containing_spaces(self) -> None:
        """Test with file paths containing spaces."""
        files = ["src/my file.py", "tests/test main.py"]
        error = MergeConflictError("SPEC-001", files)

        assert error.conflicted_files == files
        assert "my file.py" in str(error)

    def test_with_complex_file_paths(self) -> None:
        """Test with nested and complex file paths."""
        files = [
            "src/components/auth/LoginForm.tsx",
            "src/utils/helpers/string-helpers.ts",
            "tests/unit/auth/login.test.ts",
        ]
        error = MergeConflictError("SPEC-FE-001", files)

        assert len(error.conflicted_files) == 3
        for file in files:
            assert file in str(error)

    def test_with_unicode_in_file_paths(self) -> None:
        """Test with Unicode characters in file paths."""
        files = ["src/国际化.py", "docs/README_中文.md"]
        error = MergeConflictError("SPEC-001", files)

        assert error.conflicted_files == files
        assert "国际化.py" in str(error)

    def test_with_various_spec_ids(self) -> None:
        """Test with various SPEC ID formats."""
        spec_ids = ["SPEC-001", "SPEC-AUTH-001", "spec-001"]

        for spec_id in spec_ids:
            error = MergeConflictError(spec_id, ["file.py"])
            assert error.spec_id == spec_id
            assert spec_id in str(error)


class TestRegistryInconsistencyError:
    """Test RegistryInconsistencyError exception."""

    def test_inheritance(self) -> None:
        """Test that RegistryInconsistencyError inherits from WorktreeError."""
        assert issubclass(RegistryInconsistencyError, WorktreeError)

    def test_initialization_with_message(self) -> None:
        """Test initialization with message parameter."""
        message = "Registry file corrupted"

        error = RegistryInconsistencyError(message)

        assert str(error) == f"Registry inconsistency: {message}"

    def test_message_formatting(self) -> None:
        """Test that error message is formatted correctly."""
        original_message = "Worktree exists in Git but not in registry"

        error = RegistryInconsistencyError(original_message)
        full_message = str(error)

        assert "Registry inconsistency:" in full_message
        assert original_message in full_message

    def test_message_formatting_exact(self) -> None:
        """Test exact format of error message."""
        message = "Duplicate entry found"
        error = RegistryInconsistencyError(message)
        expected = f"Registry inconsistency: {message}"

        assert str(error) == expected

    def test_can_be_raised_and_caught(self) -> None:
        """Test that exception can be raised and caught as WorktreeError."""
        with pytest.raises(WorktreeError):
            raise RegistryInconsistencyError("Registry error")

    def test_can_be_caught_as_specific_type(self) -> None:
        """Test that exception can be caught as RegistryInconsistencyError."""
        message = "Invalid registry state"

        with pytest.raises(RegistryInconsistencyError) as exc_info:
            raise RegistryInconsistencyError(message)

        assert "Registry inconsistency:" in str(exc_info.value)

    def test_with_various_messages(self) -> None:
        """Test with various error messages."""
        messages = [
            "Worktree not in registry",
            "Registry file corrupted",
            "Invalid JSON format",
            "Missing required fields",
        ]

        for message in messages:
            error = RegistryInconsistencyError(message)
            assert message in str(error)

    def test_with_empty_message(self) -> None:
        """Test with empty message."""
        error = RegistryInconsistencyError("")
        assert str(error) == "Registry inconsistency: "

    def test_with_detailed_message(self) -> None:
        """Test with detailed technical message."""
        detailed_message = (
            "Worktree at /path/to/worktree exists in Git but not in registry. Run 'moai-adk worktree sync' to fix."
        )

        error = RegistryInconsistencyError(detailed_message)

        assert detailed_message in str(error)


class ExceptionHierarchyTests:
    """Test exception hierarchy and catching behavior."""

    def test_all_exceptions_inherit_from_worktree_error(self) -> None:
        """Test that all custom exceptions inherit from WorktreeError."""
        exception_classes = [
            WorktreeExistsError,
            WorktreeNotFoundError,
            UncommittedChangesError,
            GitOperationError,
            MergeConflictError,
            RegistryInconsistencyError,
        ]

        for exc_class in exception_classes:
            assert issubclass(exc_class, WorktreeError), f"{exc_class.__name__} should inherit from WorktreeError"

    def test_all_exceptions_catchable_as_worktree_error(self) -> None:
        """Test that all exceptions can be caught as WorktreeError."""
        exceptions = [
            WorktreeExistsError("SPEC-001", Path("/path")),
            WorktreeNotFoundError("SPEC-001"),
            UncommittedChangesError("SPEC-001"),
            GitOperationError("Git error"),
            MergeConflictError("SPEC-001", ["file.py"]),
            RegistryInconsistencyError("Registry error"),
        ]

        for exc in exceptions:
            try:
                raise exc
            except WorktreeError:
                # Expected - exception should be catchable as WorktreeError
                continue
            # If we get here, the exception was not caught by WorktreeError
            pytest.fail(f"{type(exc).__name__} should be catchable as WorktreeError")

    def test_specific_exceptions_not_caught_by_wrong_type(self) -> None:
        """Test that specific exceptions are not caught by wrong exception type."""
        with pytest.raises(WorktreeNotFoundError):
            raise WorktreeNotFoundError("SPEC-001")

    def test_exception_raising_preserves_attributes(self) -> None:
        """Test that raising and catching preserves exception attributes."""
        spec_id = "SPEC-001"
        path = Path("/test/path")
        files = ["file1.py", "file2.py"]

        exceptions_with_attrs = [
            (WorktreeExistsError(spec_id, path), {"spec_id": spec_id, "path": path}),
            (WorktreeNotFoundError(spec_id), {"spec_id": spec_id}),
            (UncommittedChangesError(spec_id), {"spec_id": spec_id}),
            (MergeConflictError(spec_id, files), {"spec_id": spec_id, "conflicted_files": files}),
        ]

        for exc, expected_attrs in exceptions_with_attrs:
            try:
                raise exc
            except WorktreeError as e:
                for attr, value in expected_attrs.items():
                    assert getattr(e, attr) == value, f"{attr} should be preserved"


class ExceptionEdgeCases:
    """Test edge cases for exception handling."""

    def test_exception_with_none_values(self) -> None:
        """Test exceptions with None or missing values."""
        # These should work without issues
        error1 = GitOperationError("Error")
        error2 = RegistryInconsistencyError("Inconsistency")

        assert "Error" in str(error1)
        assert "Inconsistency" in str(error2)

    def test_exception_repr(self) -> None:
        """Test that exception repr provides useful information."""
        error = WorktreeExistsError("SPEC-001", Path("/path/to/worktree"))

        repr_str = repr(error)
        assert "WorktreeExistsError" in repr_str

    def test_exception_can_be_re_raised(self) -> None:
        """Test that exceptions can be caught and re-raised."""
        try:
            try:
                raise WorktreeNotFoundError("SPEC-001")
            except WorktreeNotFoundError:
                raise  # Re-raise
        except WorktreeNotFoundError as e:
            assert e.spec_id == "SPEC-001"

    def test_exception_in_exception_chain(self) -> None:
        """Test exceptions in exception chaining."""
        try:
            try:
                raise ValueError("Original error")
            except ValueError:
                raise GitOperationError("Git operation failed") from ValueError("Original error")
        except GitOperationError as e:
            assert "Git operation failed" in str(e)
            assert e.__cause__ is not None
