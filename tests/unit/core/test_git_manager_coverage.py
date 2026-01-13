"""Additional coverage tests for Git manager.

Tests for lines not covered by existing tests.
"""

from pathlib import Path
from unittest.mock import patch

from moai_adk.core.git.manager import GitManager


class TestGitManagerIsRepoException:
    """Test is_repo exception handling."""

    def test_is_repo_returns_true_for_valid_repo(self, tmp_git_repo: Path):
        """Should return True for valid Git repository."""
        manager = GitManager(str(tmp_git_repo))
        result = manager.is_repo()
        assert result is True

    def test_is_repo_returns_false_when_git_dir_raises_exception(self, tmp_git_repo: Path):
        """Should return False when accessing repo.git_dir raises an exception."""
        from unittest.mock import MagicMock, PropertyMock

        manager = GitManager(str(tmp_git_repo))

        # Create a mock repo that raises exception when accessing git_dir
        mock_repo = MagicMock()
        type(mock_repo).git_dir = PropertyMock(side_effect=Exception("Simulated error"))

        # Replace manager's repo with the mock
        with patch.object(manager, "repo", mock_repo):
            result = manager.is_repo()
            assert result is False


class TestGitManagerMergeConflictMethods:
    """Test merge conflict related methods."""

    def test_check_merge_conflicts_delegates_to_detector(self, tmp_git_repo: Path):
        """Should delegate to conflict detector."""
        manager = GitManager(str(tmp_git_repo))

        # Mock the conflict detector
        with patch.object(
            manager.conflict_detector,
            "can_merge",
            return_value={"can_merge": True, "conflicts": []},
        ):
            result = manager.check_merge_conflicts("feature", "main")

            assert result["can_merge"] is True

    def test_has_merge_conflicts_when_can_merge_not_in_result(self, tmp_git_repo: Path):
        """Should return True when can_merge is not in result."""
        manager = GitManager(str(tmp_git_repo))

        # Mock the conflict detector to return result without can_merge key
        with patch.object(manager.conflict_detector, "can_merge", return_value={"conflicts": []}):
            result = manager.has_merge_conflicts("feature", "main")

            assert result is True  # not result.get("can_merge", False) -> not False -> True

    def test_has_merge_conflicts_when_can_merge_is_true(self, tmp_git_repo: Path):
        """Should return False when can_merge is True."""
        manager = GitManager(str(tmp_git_repo))

        with patch.object(
            manager.conflict_detector,
            "can_merge",
            return_value={"can_merge": True, "conflicts": []},
        ):
            result = manager.has_merge_conflicts("feature", "main")

            assert result is False

    def test_has_merge_conflicts_when_can_merge_is_false(self, tmp_git_repo: Path):
        """Should return True when can_merge is False."""
        manager = GitManager(str(tmp_git_repo))

        with patch.object(
            manager.conflict_detector,
            "can_merge",
            return_value={"can_merge": False, "conflicts": []},
        ):
            result = manager.has_merge_conflicts("feature", "main")

            assert result is True


class TestGitManagerGetConflictSummary:
    """Test get_conflict_summary method."""

    def test_get_conflict_summary_delegates_to_detector(self, tmp_git_repo: Path):
        """Should delegate to conflict detector summarize method."""
        manager = GitManager(str(tmp_git_repo))

        with patch.object(
            manager.conflict_detector,
            "can_merge",
            return_value={
                "can_merge": False,
                "conflicts": [{"file": "test.py", "sections": ["code"]}],
            },
        ):
            with patch.object(
                manager.conflict_detector,
                "summarize_conflicts",
                return_value="Summary: test.py has conflicts",
            ):
                result = manager.get_conflict_summary("feature", "main")

                assert result == "Summary: test.py has conflicts"


class TestGitManagerAutoResolveSafeConflicts:
    """Test auto_resolve_safe_conflicts method."""

    def test_auto_resolve_safe_conflicts_delegates_to_detector(self, tmp_git_repo: Path):
        """Should delegate to conflict detector."""
        manager = GitManager(str(tmp_git_repo))

        with patch.object(manager.conflict_detector, "auto_resolve_safe", return_value=True):
            result = manager.auto_resolve_safe_conflicts()

            assert result is True


class TestGitManagerAbortMerge:
    """Test abort_merge method."""

    def test_abort_merge_delegates_to_detector(self, tmp_git_repo: Path):
        """Should delegate to conflict detector cleanup."""
        manager = GitManager(str(tmp_git_repo))

        with patch.object(manager.conflict_detector, "cleanup_merge_state") as mock_cleanup:
            manager.abort_merge()

            mock_cleanup.assert_called_once()


class TestGitManagerPushSetBranch:
    """Test push method with branch parameter."""

    def test_push_with_set_upstream_uses_current_branch(self, tmp_git_repo: Path):
        """Should use current branch when branch is None with set_upstream."""
        manager = GitManager(str(tmp_git_repo))

        # Mock current_branch to verify it's being called
        with patch.object(manager, "current_branch", return_value="main") as mock_branch:
            # The push method will fail without a remote, but we can check it tries
            try:
                manager.push(branch=None, set_upstream=True)
            except Exception:
                # Expected to fail without remote, that's OK
                pass

            # current_branch should have been called to determine the target
            mock_branch.assert_called_once()
