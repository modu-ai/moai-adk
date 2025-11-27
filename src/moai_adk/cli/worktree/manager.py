"""Core manager for Git worktree operations."""

from datetime import datetime
from pathlib import Path

from git import Repo

from moai_adk.cli.worktree.exceptions import (
    GitOperationError,
    MergeConflictError,
    UncommittedChangesError,
    WorktreeExistsError,
    WorktreeNotFoundError,
)
from moai_adk.cli.worktree.models import WorktreeInfo
from moai_adk.cli.worktree.registry import WorktreeRegistry


class WorktreeManager:
    """Manages Git worktrees for parallel SPEC development.

    This class provides high-level operations for creating, removing,
    switching, and maintaining Git worktrees. It integrates with
    GitPython for Git operations and WorktreeRegistry for metadata
    persistence.
    """

    def __init__(self, repo_path: Path, worktree_root: Path) -> None:
        """Initialize the worktree manager.

        Args:
            repo_path: Path to the main Git repository.
            worktree_root: Root directory for all worktrees.
        """
        self.repo = Repo(repo_path)
        self.worktree_root = Path(worktree_root)
        self.registry = WorktreeRegistry(self.worktree_root)

    def create(
        self,
        spec_id: str,
        branch_name: str | None = None,
        base_branch: str = "main",
        force: bool = False,
    ) -> WorktreeInfo:
        """Create a new worktree for a SPEC.

        Creates a new Git worktree and registers it in the registry.

        Args:
            spec_id: SPEC ID (e.g., 'SPEC-AUTH-001').
            branch_name: Custom branch name (defaults to spec_id).
            base_branch: Base branch to create from (defaults to 'main').
            force: Force creation even if worktree exists.

        Returns:
            WorktreeInfo for the created worktree.

        Raises:
            WorktreeExistsError: If worktree already exists (unless force=True).
            GitOperationError: If Git operation fails.
        """
        # Check if worktree already exists
        existing = self.registry.get(spec_id)
        if existing and not force:
            raise WorktreeExistsError(spec_id, existing.path)

        # If force and exists, remove first
        if existing and force:
            try:
                self.remove(spec_id, force=True)
            except WorktreeNotFoundError:
                pass

        # Determine branch name
        if branch_name is None:
            branch_name = f"feature/{spec_id}"

        # Create worktree path
        worktree_path = self.worktree_root / spec_id

        try:
            # Create parent directory if needed
            self.worktree_root.mkdir(parents=True, exist_ok=True)

            # Fetch latest
            try:
                self.repo.remotes.origin.fetch()
            except Exception:
                # No origin or fetch fails, continue with local
                pass

            # Check if branch exists, if not create it
            try:
                self.repo.heads[base_branch]
            except IndexError:
                # Branch doesn't exist locally, try to fetch
                try:
                    self.repo.remotes.origin.fetch(base_branch)
                except Exception:
                    pass

            # Create branch if it doesn't exist
            if branch_name not in [h.name for h in self.repo.heads]:
                try:
                    self.repo.create_head(branch_name, base_branch)
                except Exception as e:
                    # If create fails, check if it already exists
                    if branch_name not in [h.name for h in self.repo.heads]:
                        raise GitOperationError(f"Failed to create branch: {e}")

            # Create worktree using git command
            try:
                self.repo.git.worktree("add", str(worktree_path), branch_name)
            except Exception as e:
                raise GitOperationError(f"Failed to create worktree: {e}")

            # Create WorktreeInfo
            now = datetime.now().isoformat() + "Z"
            info = WorktreeInfo(
                spec_id=spec_id,
                path=worktree_path,
                branch=branch_name,
                created_at=now,
                last_accessed=now,
                status="active",
            )

            # Register in registry
            self.registry.register(info)

            return info

        except WorktreeExistsError:
            raise
        except GitOperationError:
            raise
        except Exception as e:
            raise GitOperationError(f"Failed to create worktree: {e}")

    def remove(self, spec_id: str, force: bool = False) -> None:
        """Remove a worktree.

        Args:
            spec_id: SPEC ID of worktree to remove.
            force: Force removal even with uncommitted changes.

        Raises:
            WorktreeNotFoundError: If worktree doesn't exist.
            UncommittedChangesError: If worktree has uncommitted changes (unless force=True).
            GitOperationError: If Git operation fails.
        """
        info = self.registry.get(spec_id)
        if not info:
            raise WorktreeNotFoundError(spec_id)

        try:
            # Check for uncommitted changes
            if not force:
                try:
                    status = self.repo.git.status("--porcelain", spec_id)
                    if status.strip():
                        raise UncommittedChangesError(spec_id)
                except Exception:
                    # Ignore status check errors
                    pass

            # Remove worktree
            try:
                self.repo.git.worktree("remove", str(info.path), "--force")
            except Exception:
                # Try to remove directory manually if git command fails
                import shutil
                if info.path.exists():
                    shutil.rmtree(info.path)

            # Unregister from registry
            self.registry.unregister(spec_id)

        except (WorktreeNotFoundError, UncommittedChangesError):
            raise
        except GitOperationError:
            raise
        except Exception as e:
            raise GitOperationError(f"Failed to remove worktree: {e}")

    def list(self) -> list["WorktreeInfo"]:
        """List all worktrees.

        Returns:
            List of WorktreeInfo instances.
        """
        return self.registry.list_all()

    def sync(self, spec_id: str, base_branch: str = "main") -> None:
        """Sync worktree with base branch.

        Fetches latest changes from base branch and merges them.

        Args:
            spec_id: SPEC ID of worktree to sync.
            base_branch: Branch to sync from (defaults to 'main').

        Raises:
            WorktreeNotFoundError: If worktree doesn't exist.
            MergeConflictError: If merge conflict occurs.
            GitOperationError: If Git operation fails.
        """
        info = self.registry.get(spec_id)
        if not info:
            raise WorktreeNotFoundError(spec_id)

        try:
            # Change to worktree directory
            worktree_repo = Repo(info.path)

            # Fetch latest
            try:
                worktree_repo.remotes.origin.fetch()
            except Exception:
                pass

            # Merge base branch
            try:
                worktree_repo.git.merge(f"origin/{base_branch}")
            except Exception as e:
                # Check for merge conflicts
                try:
                    status = worktree_repo.git.status("--porcelain")
                    conflicted = [
                        line.split()[-1]
                        for line in status.split("\n")
                        if line.startswith("UU") or line.startswith("DD")
                    ]
                    if conflicted:
                        raise MergeConflictError(spec_id, conflicted)
                except MergeConflictError:
                    raise
                except Exception:
                    pass
                raise GitOperationError(f"Failed to sync worktree: {e}")

            # Update last accessed time
            info.last_accessed = datetime.now().isoformat() + "Z"
            self.registry.register(info)

        except (WorktreeNotFoundError, MergeConflictError):
            raise
        except GitOperationError:
            raise
        except Exception as e:
            raise GitOperationError(f"Failed to sync worktree: {e}")

    def clean_merged(self) -> list[str]:
        """Clean up worktrees for merged branches.

        Removes worktrees whose branches have been merged to main.

        Returns:
            List of spec_ids that were cleaned up.
        """
        cleaned = []

        try:
            # Get list of merged branches
            try:
                merged_branches = self.repo.git.branch(
                    "--merged", "main"
                ).split("\n")
                merged_branches = [b.strip().lstrip("*").strip() for b in merged_branches]
            except Exception:
                merged_branches = []

            # Check each worktree
            for info in self.list():
                if info.branch in merged_branches:
                    try:
                        self.remove(info.spec_id, force=True)
                        cleaned.append(info.spec_id)
                    except Exception:
                        pass

        except Exception:
            pass

        return cleaned
