"""Tests for worktree exceptions module."""

import pytest

from moai_adk.cli.worktree.exceptions import RegistryInconsistencyError


class TestRegistryInconsistencyError:
    """Test RegistryInconsistencyError exception."""

    def test_exception_message_format(self):
        """Test exception message is formatted correctly."""
        error = RegistryInconsistencyError("test error message")

        assert str(error) == "Registry inconsistency: test error message"

    def test_exception_inheritance(self):
        """Test exception inherits from WorktreeError."""
        from moai_adk.cli.worktree.exceptions import WorktreeError

        error = RegistryInconsistencyError("test")

        assert isinstance(error, WorktreeError)

    def test_exception_can_be_raised(self):
        """Test exception can be raised and caught."""
        with pytest.raises(RegistryInconsistencyError) as exc_info:
            raise RegistryInconsistencyError("inconsistency detected")

        assert "inconsistency detected" in str(exc_info.value)
