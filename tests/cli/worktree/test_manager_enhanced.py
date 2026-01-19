"""Enhanced tests for worktree manager with extended coverage.

This module provides additional test coverage for:
- Git failures (fetch, branch creation, worktree add)
- Merge conflicts and auto-resolve
- LLM config copying with environment variable substitution
- Force operations (force create, force remove)
- Sync operations with rebase and ff_only
- Clean merged functionality
- Done workflow (merge to base and remove)
"""

from pathlib import Path
from typing import Generator
from unittest.mock import MagicMock, patch

import pytest
from git import Repo

from moai_adk.cli.worktree.exceptions import (
    GitOperationError,
    MergeConflictError,
    UncommittedChangesError,
    WorktreeExistsError,
    WorktreeNotFoundError,
)
from moai_adk.cli.worktree.manager import WorktreeManager


@pytest.fixture
def temp_repo_dir(tmp_path: Path) -> Generator[Path, None, None]:
    """Create a temporary Git repository for testing.

    Yields:
        Path to the temporary repository.
    """
    repo_dir = tmp_path / "test_repo"
    repo_dir.mkdir(parents=True, exist_ok=True)

    # Initialize Git repository with explicit initial branch name
    repo = Repo.init(repo_dir, initial_branch="main")
    repo.config_writer().set_value("user", "name", "Test User").release()
    repo.config_writer().set_value("user", "email", "test@example.com").release()

    # Create initial commit
    test_file = repo_dir / "README.md"
    test_file.write_text("# Test Repo")
    repo.index.add([str(test_file)])
    repo.index.commit("Initial commit")

    yield repo_dir


@pytest.fixture
def temp_worktree_root(tmp_path: Path) -> Generator[Path, None, None]:
    """Create a temporary worktree root directory.

    Yields:
        Path to the temporary worktree root.
    """
    worktree_root = tmp_path / "worktrees"
    worktree_root.mkdir(parents=True, exist_ok=True)
    yield worktree_root


@pytest.fixture
def manager(temp_repo_dir: Path, temp_worktree_root: Path) -> WorktreeManager:
    """Create a WorktreeManager instance for testing.

    Args:
        temp_repo_dir: Temporary repository directory.
        temp_worktree_root: Temporary worktree root directory.

    Returns:
        WorktreeManager instance.
    """
    return WorktreeManager(repo_path=temp_repo_dir, worktree_root=temp_worktree_root)


class TestWorktreeManagerForceOperations:
    """Test force operations in worktree manager."""

    def test_create_with_force_removes_existing_worktree(self, manager: WorktreeManager) -> None:
        """Test that force=True removes existing worktree before creating new one."""
        spec_id = "SPEC-FORCE-001"

        # Create initial worktree
        info1 = manager.create(spec_id=spec_id, base_branch="main")
        assert info1.path.exists()

        # Create with force=True should replace existing
        info2 = manager.create(spec_id=spec_id, base_branch="main", force=True)

        # Verify worktree still exists (was recreated)
        assert info2.path.exists()
        assert info2.spec_id == spec_id

    def test_create_without_force_raises_on_existing(self, manager: WorktreeManager) -> None:
        """Test that force=False raises WorktreeExistsError when worktree exists."""
        spec_id = "SPEC-EXISTS-001"

        # Create initial worktree
        manager.create(spec_id=spec_id, base_branch="main")

        # Attempt to create without force should raise
        with pytest.raises(WorktreeExistsError):
            manager.create(spec_id=spec_id, base_branch="main", force=False)

    def test_remove_with_force_ignores_uncommitted_changes(self, manager: WorktreeManager) -> None:
        """Test that force=True removes worktree even with uncommitted changes."""
        spec_id = "SPEC-FORCE-REMOVE"

        # Create worktree
        info = manager.create(spec_id=spec_id, base_branch="main")

        # Add uncommitted changes
        test_file = info.path / "uncommitted.txt"
        test_file.write_text("Uncommitted content")

        # Should not raise with force=True
        manager.remove(spec_id=spec_id, force=True)

        # Verify worktree was removed
        assert not info.path.exists()

    def test_remove_without_force_raises_on_uncommitted(self, manager: WorktreeManager) -> None:
        """Test that force=False raises UncommittedChangesError when changes exist."""
        spec_id = "SPEC-CHANGES-001"

        # Create worktree
        info = manager.create(spec_id=spec_id, base_branch="main")

        # Add uncommitted changes
        test_file = info.path / "uncommitted.txt"
        test_file.write_text("Uncommitted content")

        # Should raise with force=False
        with pytest.raises(UncommittedChangesError):
            manager.remove(spec_id=spec_id, force=False)


class TestWorktreeManagerGitFailures:
    """Test handling of Git operation failures."""

    def test_create_handles_git_fetch_failure(self, manager: WorktreeManager) -> None:
        """Test that worktree creation continues even if fetch fails."""
        spec_id = "SPEC-FETCH-FAIL"

        with patch.object(manager.repo.remotes, "origin", side_effect=Exception("No remote")):
            # Should still succeed despite fetch failure
            info = manager.create(spec_id=spec_id, base_branch="main")
            assert info.path.exists()

    def test_create_handles_branch_already_exists(self, manager: WorktreeManager) -> None:
        """Test that worktree creation handles when branch already exists."""
        spec_id = "SPEC-BRANCH-EXISTS"

        # Create branch manually first
        branch_name = f"feature/{spec_id}"
        try:
            manager.repo.create_head(branch_name, "main")
        except Exception:
            pass  # Branch might already exist

        # Should still succeed
        info = manager.create(spec_id=spec_id, base_branch="main", branch_name=branch_name)
        assert info.path.exists()

    def test_create_raises_on_git_worktree_failure(self, manager: WorktreeManager) -> None:
        """Test that GitOperationError is raised when git worktree add fails."""
        spec_id = "SPEC-GIT-FAIL-001"

        with patch.object(manager.repo.git, "worktree", side_effect=Exception("Git command failed")):
            with pytest.raises(GitOperationError):
                manager.create(spec_id=spec_id, base_branch="main")

    def test_remove_handles_git_worktree_remove_failure(self, manager: WorktreeManager) -> None:
        """Test that remove falls back to manual directory removal when git command fails."""
        spec_id = "SPEC-FALLBACK-001"

        # Create worktree
        info = manager.create(spec_id=spec_id, base_branch="main")

        # Mock git worktree remove to fail
        original_git_worktree = manager.repo.git.worktree

        def failing_worktree(*args, **kwargs):
            if "remove" in args:
                raise Exception("Git worktree remove failed")
            return original_git_worktree(*args, **kwargs)

        with patch.object(manager.repo.git, "worktree", side_effect=failing_worktree):
            # Should still succeed by falling back to manual removal
            manager.remove(spec_id=spec_id, force=True)

            # Verify worktree was removed
            assert not info.path.exists()

    def test_sync_handles_fetch_failure(self, manager: WorktreeManager) -> None:
        """Test that sync handles fetch failures gracefully."""
        spec_id = "SPEC-SYNC-FETCH"

        # Create worktree
        manager.create(spec_id=spec_id, base_branch="main")

        # Mock fetch to fail
        with patch("git.Repo") as MockRepo:
            mock_worktree_repo = MagicMock()
            MockRepo.return_value = mock_worktree_repo
            mock_worktree_repo.remotes.origin.fetch.side_effect =Exception("Fetch failed")
            mock_worktree_repo.git.merge.return_value = None
            mock_worktree_repo.git.status.return_value = ""

            # Should not raise despite fetch failure
            manager.sync(spec_id=spec_id, base_branch="main")

    def test_sync_raises_on_base_branch_not_found(self, manager: WorktreeManager) -> None:
        """Test that sync raises GitOperationError when base branch not found."""
        spec_id = "SPEC-BRANCH-NOT-FOUND"

        # Create worktree
        manager.create(spec_id=spec_id, base_branch="main")

        # Try to sync from non-existent branch
        with patch("git.Repo") as MockRepo:
            mock_worktree_repo = MagicMock()
            MockRepo.return_value = mock_worktree_repo
            mock_worktree_repo.remotes.origin.fetch.return_value = None
            mock_worktree_repo.git.rev_parse.side_effect = Exception("Branch not found")
            mock_worktree_repo.git.merge.side_effect = Exception("Merge failed")
            mock_worktree_repo.git.status.return_value = ""
            mock_worktree_repo.git.rebase.side_effect = Exception("Rebase failed")

            with pytest.raises(GitOperationError, match="Base branch"):
                manager.sync(spec_id=spec_id, base_branch="nonexistent-branch")


class TestWorktreeManagerMergeConflicts:
    """Test merge conflict handling."""

    def test_sync_raises_on_merge_conflict(self, manager: WorktreeManager) -> None:
        """Test that sync raises MergeConflictError when conflicts occur."""
        spec_id = "SPEC-CONFLICT-001"

        # Create worktree
        manager.create(spec_id=spec_id, base_branch="main")

        # Mock merge with conflict
        with patch("git.Repo") as MockRepo:
            mock_worktree_repo = MagicMock()
            MockRepo.return_value = mock_worktree_repo
            mock_worktree_repo.remotes.origin.fetch.return_value = None
            mock_worktree_repo.git.rev_parse.return_value = "abc123"
            mock_worktree_repo.git.merge.side_effect = Exception("Merge conflict")
            mock_worktree_repo.git.status.return_value = "UU conflicting-file.txt"
            mock_worktree_repo.git.merge.__name__ = "merge"

            with pytest.raises(MergeConflictError):
                manager.sync(spec_id=spec_id, base_branch="main", auto_resolve=False)

    def test_sync_auto_resolves_conflicts(self, manager: WorktreeManager) -> None:
        """Test that sync with auto_resolve=True attempts to resolve conflicts."""
        spec_id = "SPEC-AUTO-RESOLVE"

        # Create worktree
        manager.create(spec_id=spec_id, base_branch="main")

        # Mock merge with conflict then successful resolution
        with patch("git.Repo") as MockRepo:
            mock_worktree_repo = MagicMock()
            MockRepo.return_value = mock_worktree_repo
            mock_worktree_repo.remotes.origin.fetch.return_value = None
            mock_worktree_repo.git.rev_parse.return_value = "abc123"

            # First merge fails, then resolution succeeds
            call_count = [0]

            def mock_merge(*args, **kwargs):
                call_count[0] += 1
                if call_count[0] == 1:
                    raise Exception("Merge conflict")
                return None

            mock_worktree_repo.git.merge = mock_merge
            mock_worktree_repo.git.status.return_value = "UU conflicting-file.txt"
            mock_worktree_repo.git.checkout.return_value = None
            mock_worktree_repo.git.add.return_value = None
            mock_worktree_repo.git.commit.return_value = None

            # Should auto-resolve and not raise
            manager.sync(spec_id=spec_id, base_branch="main", auto_resolve=True)

    def test_sync_auto_resolve_falls_back_to_manual(self, manager: WorktreeManager) -> None:
        """Test that auto_resolve falls back to manual when resolution fails."""
        spec_id = "SPEC-RESOLVE-FAIL"

        # Create worktree
        manager.create(spec_id=spec_id, base_branch="main")

        # Mock merge with conflict and failed resolution
        with patch("git.Repo") as MockRepo:
            mock_worktree_repo = MagicMock()
            MockRepo.return_value = mock_worktree_repo
            mock_worktree_repo.remotes.origin.fetch.return_value = None
            mock_worktree_repo.git.rev_parse.return_value = "abc123"
            mock_worktree_repo.git.merge.side_effect = Exception("Merge conflict")
            mock_worktree_repo.git.status.return_value = "UU conflicting-file.txt"
            mock_worktree_repo.git.checkout.side_effect = Exception("Checkout failed")
            mock_worktree_repo.git.merge.__name__ = "merge"

            with pytest.raises(MergeConflictError):
                manager.sync(spec_id=spec_id, base_branch="main", auto_resolve=True)

    def test_done_raises_on_merge_conflict(self, manager: WorktreeManager) -> None:
        """Test that done() raises MergeConflictError when merge conflicts occur."""
        spec_id = "SPEC-DONE-CONFLICT"

        # Create worktree
        manager.create(spec_id=spec_id, base_branch="main")

        # Mock merge with conflict
        with patch.object(manager.repo.git, "merge", side_effect=Exception("Merge conflict")):
            with patch.object(manager.repo.git, "status", return_value="UU conflicting-file.txt"):
                with patch.object(manager.repo.git, "merge__abort", return_value=None):
                    with pytest.raises(MergeConflictError):
                        manager.done(spec_id=spec_id, base_branch="main")


class TestWorktreeManagerSyncModes:
    """Test different sync modes (merge, rebase, ff_only)."""

    def test_sync_with_rebase(self, manager: WorktreeManager) -> None:
        """Test sync with rebase=True uses rebase strategy."""
        spec_id = "SPEC-REBASE-001"

        # Create worktree
        manager.create(spec_id=spec_id, base_branch="main")

        with patch("git.Repo") as MockRepo:
            mock_worktree_repo = MagicMock()
            MockRepo.return_value = mock_worktree_repo
            mock_worktree_repo.remotes.origin.fetch.return_value = None
            mock_worktree_repo.git.rev_parse.return_value = "abc123"
            mock_worktree_repo.git.rebase.return_value = None

            # Sync with rebase
            manager.sync(spec_id=spec_id, base_branch="main", rebase=True)

            # Verify rebase was called
            mock_worktree_repo.git.rebase.assert_called_once()

    def test_sync_with_ff_only(self, manager: WorktreeManager) -> None:
        """Test sync with ff_only=True uses fast-forward only."""
        spec_id = "SPEC-FFONLY-001"

        # Create worktree
        manager.create(spec_id=spec_id, base_branch="main")

        with patch("git.Repo") as MockRepo:
            mock_worktree_repo = MagicMock()
            MockRepo.return_value = mock_worktree_repo
            mock_worktree_repo.remotes.origin.fetch.return_value = None
            mock_worktree_repo.git.rev_parse.return_value = "abc123"
            mock_worktree_repo.git.merge.return_value = None

            # Sync with ff_only
            manager.sync(spec_id=spec_id, base_branch="main", ff_only=True)

            # Verify merge with --ff-only was called
            mock_worktree_repo.git.merge.assert_called_once()
            call_args = mock_worktree_repo.git.merge.call_args
            assert "--ff-only" in call_args[0]

    def test_sync_default_uses_merge(self, manager: WorktreeManager) -> None:
        """Test that default sync uses merge strategy."""
        spec_id = "SPEC-MERGE-001"

        # Create worktree
        manager.create(spec_id=spec_id, base_branch="main")

        with patch("git.Repo") as MockRepo:
            mock_worktree_repo = MagicMock()
            MockRepo.return_value = mock_worktree_repo
            mock_worktree_repo.remotes.origin.fetch.return_value = None
            mock_worktree_repo.git.rev_parse.return_value = "abc123"
            mock_worktree_repo.git.merge.return_value = None

            # Sync with default (merge)
            manager.sync(spec_id=spec_id, base_branch="main")

            # Verify merge was called
            mock_worktree_repo.git.merge.assert_called_once()


class TestWorktreeManagerLLMConfig:
    """Test LLM config copying with environment variable substitution."""

    def test_create_copies_llm_config(self, manager: WorktreeManager, tmp_path: Path) -> None:
        """Test that llm_config_path copies config to worktree."""
        spec_id = "SPEC-LLM-CONFIG"

        # Create LLM config template
        llm_config_path = tmp_path / "llm_config.json"
        llm_config_content = '{"api_key": "${GLM_API_TOKEN}", "model": "glm-4"}'
        llm_config_path.write_text(llm_config_content)

        # Set environment variable
        with patch.dict("os.environ", {"GLM_API_TOKEN": "test-token-123"}):
            info = manager.create(
                spec_id=spec_id,
                base_branch="main",
                llm_config_path=llm_config_path
            )

        # Verify config was copied and substituted
        target_config = info.path / ".claude" / "settings.local.json"
        assert target_config.exists()

        config_content = target_config.read_text()
        assert "test-token-123" in config_content
        assert "${GLM_API_TOKEN}" not in config_content

    def test_create_skips_llm_config_if_path_missing(self, manager: WorktreeManager) -> None:
        """Test that missing llm_config_path is skipped without error."""
        spec_id = "SPEC-NO-CONFIG"

        non_existent_path = Path("/non/existent/path/config.json")

        # Should not raise despite missing config
        info = manager.create(
            spec_id=spec_id,
            base_branch="main",
            llm_config_path=non_existent_path
        )

        # Worktree should still be created
        assert info.path.exists()

    def test_llm_config_preserves_unknown_vars(self, manager: WorktreeManager, tmp_path: Path) -> None:
        """Test that unknown environment variables are preserved as-is."""
        spec_id = "SPEC-UNKNOWN-VAR"

        # Create LLM config with unknown variable
        llm_config_path = tmp_path / "unknown_var_config.json"
        llm_config_content = '{"key": "${UNKNOWN_VAR}"}'
        llm_config_path.write_text(llm_config_content)

        with patch.dict("os.environ", {}, clear=True):
            info = manager.create(
                spec_id=spec_id,
                base_branch="main",
                llm_config_path=llm_config_path
            )

        # Unknown variable should be preserved
        target_config = info.path / ".claude" / "settings.local.json"
        config_content = target_config.read_text()
        assert "${UNKNOWN_VAR}" in config_content


class TestWorktreeManagerCleanMerged:
    """Test clean_merged functionality."""

    def test_clean_merged_removes_merged_branches(self, manager: WorktreeManager) -> None:
        """Test that clean_merged removes worktrees with merged branches."""
        spec_id = "SPEC-MERGED-001"

        # Create worktree
        manager.create(spec_id=spec_id, base_branch="main")

        # Mock git branch --merged to include our branch
        with patch.object(manager.repo.git, "branch", return_value=f"* main\n  feature/{spec_id}"):
            cleaned = manager.clean_merged()

            # Verify our worktree was cleaned
            assert spec_id in cleaned

    def test_clean_merged_skips_unmerged_branches(self, manager: WorktreeManager) -> None:
        """Test that clean_merged skips worktrees with unmerged branches."""
        spec_id = "SPEC-UNMERGED-001"

        # Create worktree
        manager.create(spec_id=spec_id, base_branch="main")

        # Mock git branch --merged to not include our branch
        with patch.object(manager.repo.git, "branch", return_value="* main"):
            cleaned = manager.clean_merged()

            # Verify our worktree was not cleaned
            assert spec_id not in cleaned

    def test_clean_merged_handles_git_command_failure(self, manager: WorktreeManager) -> None:
        """Test that clean_merged handles git command failures gracefully."""
        spec_id = "SPEC-CLEAN-FAIL"

        # Create worktree
        manager.create(spec_id=spec_id, base_branch="main")

        # Mock git branch to fail
        with patch.object(manager.repo.git, "branch", side_effect=Exception("Git failed")):
            # Should not raise, returns empty list
            cleaned = manager.clean_merged()
            assert cleaned == []


class TestWorktreeManagerDone:
    """Test done workflow (merge to base and remove)."""

    def test_done_merges_and_removes_worktree(self, manager: WorktreeManager) -> None:
        """Test that done() merges branch to base and removes worktree."""
        spec_id = "SPEC-DONE-001"

        # Create worktree
        manager.create(spec_id=spec_id, base_branch="main")

        # Mock successful merge
        with patch.object(manager.repo.git, "checkout", return_value=None):
            with patch.object(manager.repo.git, "merge", return_value=None):
                with patch.object(manager.repo.git, "branch", return_value=None):
                    result = manager.done(spec_id=spec_id, base_branch="main")

        # Verify result
        assert result["merged_branch"] == f"feature/{spec_id}"
        assert result["base_branch"] == "main"
        assert result["pushed"] is False

    def test_done_with_push(self, manager: WorktreeManager) -> None:
        """Test that done() with push=True pushes to remote."""
        spec_id = "SPEC-DONE-PUSH"

        # Create worktree
        manager.create(spec_id=spec_id, base_branch="main")

        # Mock successful merge and push
        with patch.object(manager.repo.git, "checkout", return_value=None):
            with patch.object(manager.repo.git, "merge", return_value=None):
                with patch.object(manager.repo.git, "push", return_value=None):
                    with patch.object(manager.repo.git, "branch", return_value=None):
                        result = manager.done(spec_id=spec_id, base_branch="main", push=True)

        # Verify push happened
        assert result["pushed"] is True

    def test_done_raises_on_nonexistent_worktree(self, manager: WorktreeManager) -> None:
        """Test that done() raises WorktreeNotFoundError for non-existent worktree."""
        with pytest.raises(WorktreeNotFoundError):
            manager.done(spec_id="SPEC-NONEXISTENT", base_branch="main")

    def test_done_handles_checkout_failure(self, manager: WorktreeManager) -> None:
        """Test that done() handles checkout failure gracefully."""
        spec_id = "SPEC-DONE-CHECKOUT-FAIL"

        # Create worktree
        manager.create(spec_id=spec_id, base_branch="main")

        # Mock checkout failure
        with patch.object(manager.repo.git, "checkout", side_effect=Exception("Checkout failed")):
            with pytest.raises(GitOperationError):
                manager.done(spec_id=spec_id, base_branch="main")

    def test_done_handles_branch_delete_failure(self, manager: WorktreeManager) -> None:
        """Test that done() continues even if branch deletion fails."""
        spec_id = "SPEC-DONE-DELETE-FAIL"

        # Create worktree
        manager.create(spec_id=spec_id, base_branch="main")

        # Mock successful merge but failed branch delete
        with patch.object(manager.repo.git, "checkout", return_value=None):
            with patch.object(manager.repo.git, "merge", return_value=None):
                with patch.object(manager.repo.git, "branch", side_effect=Exception("Branch protected")):
                    # Should not raise despite branch delete failure
                    result = manager.done(spec_id=spec_id, base_branch="main")

        # Verify worktree was still removed
        assert result["merged_branch"] == f"feature/{spec_id}"


class TestWorktreeManagerAutoResolveConflicts:
    """Test auto_resolve_conflicts method."""

    def test_auto_resolve_accepts_ours(self, manager: WorktreeManager, tmp_path: Path) -> None:
        """Test auto_resolve accepts 'ours' changes first."""
        spec_id = "SPEC-RESOLVE-OURS"

        # Create worktree with conflict file
        info = manager.create(spec_id=spec_id, base_branch="main")
        conflict_file = info.path / "conflict.txt"
        conflict_file.write_text("Original content")

        # Mock worktree repo
        mock_repo = MagicMock()

        # First checkout (--ours) succeeds
        mock_repo.git.checkout.return_value = None
        mock_repo.git.add.return_value = None
        mock_repo.git.commit.return_value = None

        # Call auto_resolve
        manager.auto_resolve_conflicts(mock_repo, spec_id, ["conflict.txt"])

        # Verify checkout --ours was called
        mock_repo.git.checkout.assert_called()
        call_args = mock_repo.git.checkout.call_args
        assert "--ours" in call_args[0]

    def test_auto_resolve_falls_back_to_theirs(self, manager: WorktreeManager, tmp_path: Path) -> None:
        """Test auto_resolve falls back to 'theirs' if 'ours' fails."""
        spec_id = "SPEC-RESOLVE-THEIRS"

        # Create worktree
        info = manager.create(spec_id=spec_id, base_branch="main")

        # Mock worktree repo
        mock_repo = MagicMock()

        # First checkout fails, second (--theirs) succeeds
        call_count = [0]

        def mock_checkout(*args, **kwargs):
            call_count[0] += 1
            if call_count[0] == 1 and "--ours" in args:
                raise Exception("Checkout --ours failed")
            return None

        mock_repo.git.checkout = mock_checkout
        mock_repo.git.add.return_value = None
        mock_repo.git.commit.return_value = None

        # Call auto_resolve
        manager.auto_resolve_conflicts(mock_repo, spec_id, ["conflict.txt"])

        # Verify both --ours and --theirs were attempted
        assert call_count[0] >= 2

    def test_auto_resolve_removes_conflict_markers(self, manager: WorktreeManager, tmp_path: Path) -> None:
        """Test auto_resolve removes conflict markers as last resort."""
        spec_id = "SPEC-RESOLVE-MARKERS"

        # Create worktree with conflict file containing markers
        info = manager.create(spec_id=spec_id, base_branch="main")
        conflict_file = info.path / "conflict.txt"
        conflict_content = """<<<<<<< HEAD
Our content
=======
Their content
>>>>>>> branch
"""
        conflict_file.write_text(conflict_content)

        # Mock worktree repo - both checkout strategies fail
        mock_repo = MagicMock()
        mock_repo.git.checkout.side_effect = Exception("Checkout failed")
        mock_repo.git.add.return_value = None
        mock_repo.git.commit.return_value = None

        # Call auto_resolve
        manager.auto_resolve_conflicts(mock_repo, spec_id, ["conflict.txt"])

        # Verify conflict markers were removed from file
        cleaned_content = conflict_file.read_text()
        assert "<<<<<<< HEAD" not in cleaned_content
        assert "=======" not in cleaned_content
        assert ">>>>>>> branch" not in cleaned_content
        assert "Our content" in cleaned_content or "Their content" in cleaned_content
