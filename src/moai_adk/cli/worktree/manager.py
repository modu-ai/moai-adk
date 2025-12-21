"""Core manager for Git worktree operations."""

from datetime import datetime
from pathlib import Path
from typing import List

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
    """Manages Git worktrees for parallel SPEC development with project namespace support.

    This class provides high-level operations for creating, removing,
    switching, and maintaining Git worktrees. It integrates with
    GitPython for Git operations and WorktreeRegistry for metadata
    persistence. Worktrees are organized by project name namespace to
    support multiple projects: /worktrees/{project-name}/{SPEC-ID}
    """

    def __init__(self, repo_path: Path, worktree_root: Path, project_name: str | None = None) -> None:
        """Initialize the worktree manager.

        Args:
            repo_path: Path to the main Git repository.
            worktree_root: Root directory for all worktrees.
            project_name: Project name for namespace organization. Defaults to repo directory name.
        """
        self.repo = Repo(repo_path)
        self.worktree_root = Path(worktree_root)
        self.project_name = project_name or repo_path.name
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
        existing = self.registry.get(spec_id, project_name=self.project_name)
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

        # Create worktree path with project_name namespace
        worktree_path = self.worktree_root / self.project_name / spec_id

        try:
            # Create parent directory if needed (including project_name namespace)
            worktree_path.parent.mkdir(parents=True, exist_ok=True)

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
            self.registry.register(info, project_name=self.project_name)

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
        info = self.registry.get(spec_id, project_name=self.project_name)
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
            self.registry.unregister(spec_id, project_name=self.project_name)

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
        return self.registry.list_all(project_name=self.project_name)

    def sync(
        self,
        spec_id: str,
        base_branch: str = "main",
        rebase: bool = False,
        ff_only: bool = False,
        auto_resolve: bool = False,
    ) -> None:
        """Sync worktree with base branch.

        Fetches latest changes from base branch and merges them.

        Args:
            spec_id: SPEC ID of worktree to sync.
            base_branch: Branch to sync from (defaults to 'main').
            rebase: Use rebase instead of merge.
            ff_only: Only sync if fast-forward is possible.
            auto_resolve: Automatically attempt to resolve conflicts.

        Raises:
            WorktreeNotFoundError: If worktree doesn't exist.
            MergeConflictError: If merge conflict occurs and auto_resolve is False.
            GitOperationError: If Git operation fails.
        """
        info = self.registry.get(spec_id, project_name=self.project_name)
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

            # Sync with preferred strategy
            try:
                # Determine target branch (try remote first, then local)
                target_branch = None
                for candidate in [f"origin/{base_branch}", base_branch]:
                    try:
                        worktree_repo.git.rev_parse(candidate)
                        target_branch = candidate
                        break
                    except Exception:
                        continue

                if not target_branch:
                    raise GitOperationError(
                        f"Base branch '{base_branch}' not found (tried origin/{base_branch}, {base_branch})"
                    )

                if ff_only:
                    # Fast-forward only
                    worktree_repo.git.merge(target_branch, "--ff-only")
                elif rebase:
                    # Rebase strategy
                    worktree_repo.git.rebase(target_branch)
                else:
                    # Default merge strategy with conflict handling
                    worktree_repo.git.merge(target_branch)
            except Exception as e:
                # Handle merge conflicts and provide auto-abort/auto-resolve
                try:
                    status = worktree_repo.git.status("--porcelain")
                    conflicted = [
                        line.split()[-1]
                        for line in status.split("\n")
                        if line.startswith("UU")
                        or line.startswith("DD")
                        or line.startswith("AA")
                        or line.startswith("DU")
                    ]

                    if conflicted:
                        if auto_resolve:
                            # Attempt auto-resolution of conflicts
                            try:
                                # Try common conflict resolution strategies
                                for file_path in conflicted:
                                    try:
                                        # Strategy 1: Accept current changes (ours)
                                        worktree_repo.git.checkout("--ours", file_path)
                                        worktree_repo.git.add(file_path)
                                    except Exception:
                                        # Strategy 2: Accept incoming changes (theirs) if ours fails
                                        try:
                                            worktree_repo.git.checkout("--theirs", file_path)
                                            worktree_repo.git.add(file_path)
                                        except Exception:
                                            # Strategy 3: Remove conflict markers and keep both
                                            try:
                                                # Simple conflict marker removal - keep both versions
                                                file_full_path = info.path / file_path
                                                if file_full_path.exists():
                                                    with open(file_full_path, "r") as f:
                                                        content = f.read()

                                                    # Remove conflict markers and keep both versions
                                                    lines = content.split("\n")
                                                    cleaned_lines = []
                                                    skip_next = False
                                                    in_conflict = False

                                                    for line in lines:
                                                        if "<<<<<<<" in line or ">>>>>>>" in line:
                                                            in_conflict = True
                                                            continue
                                                        elif "======" in line:
                                                            skip_next = True
                                                            in_conflict = False
                                                            continue
                                                        elif skip_next:
                                                            skip_next = False
                                                            continue
                                                        elif not in_conflict:
                                                            cleaned_lines.append(line)

                                                    with open(file_full_path, "w") as f:
                                                        f.write("\n".join(cleaned_lines))

                                                    worktree_repo.git.add(file_path)
                                            except Exception:
                                                pass

                                # Stage resolved files
                                worktree_repo.git.add(".")

                                # Commit the merge resolution
                                worktree_repo.git.commit(
                                    "-m",
                                    f"Auto-resolved conflicts during sync of {spec_id}",
                                )

                            except Exception as resolve_error:
                                # Auto-resolution failed, fall back to manual conflict
                                try:
                                    worktree_repo.git.merge("--abort")
                                except Exception:
                                    pass
                                try:
                                    worktree_repo.git.rebase("--abort")
                                except Exception:
                                    pass
                                conflict_list = conflicted if isinstance(conflicted, list) else [str(conflicted)]
                                error_msg = f"auto-resolve failed: {resolve_error}"
                                raise MergeConflictError(spec_id, conflict_list + [error_msg])
                        else:
                            # Auto-abort merge/rebase on conflicts
                            try:
                                worktree_repo.git.merge("--abort")
                            except Exception:
                                pass
                            try:
                                worktree_repo.git.rebase("--abort")
                            except Exception:
                                pass

                            raise MergeConflictError(spec_id, conflicted)

                    # If no conflicts but sync failed, raise general error
                    raise GitOperationError(f"Failed to sync worktree: {e}")

                except MergeConflictError:
                    raise
                except Exception:
                    # If conflict detection fails, still try to clean up
                    try:
                        worktree_repo.git.merge("--abort")
                    except Exception:
                        pass
                    try:
                        worktree_repo.git.rebase("--abort")
                    except Exception:
                        pass
                    raise GitOperationError(f"Failed to sync worktree: {e}")

            # Update last accessed time
            info.last_accessed = datetime.now().isoformat() + "Z"
            self.registry.register(info, project_name=self.project_name)

        except (WorktreeNotFoundError, MergeConflictError):
            raise
        except GitOperationError:
            raise
        except Exception as e:
            raise GitOperationError(f"Failed to sync worktree: {e}")

    def clean_merged(self) -> List[str]:
        """Clean up worktrees for merged branches.

        Removes worktrees whose branches have been merged to main.

        Returns:
            List of spec_ids that were cleaned up.
        """
        cleaned = []

        try:
            # Get list of merged branches
            try:
                merged_branches = self.repo.git.branch("--merged", "main").split("\n")
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

    def auto_resolve_conflicts(self, worktree_repo: Repo, spec_id: str, conflicted_files: List[str]) -> None:
        """Automatically resolve conflicts in worktree.

        Args:
            worktree_repo: Git repository object for the worktree.
            spec_id: SPEC ID for error reporting.
            conflicted_files: List of conflicted files.

        Raises:
            GitOperationError: If auto-resolution fails.
        """
        try:
            # Try common conflict resolution strategies
            for file_path in conflicted_files:
                try:
                    # Strategy 1: Accept current changes (ours)
                    worktree_repo.git.checkout("--ours", file_path)
                    worktree_repo.git.add(file_path)
                except Exception:
                    # Strategy 2: Accept incoming changes (theirs) if ours fails
                    try:
                        worktree_repo.git.checkout("--theirs", file_path)
                        worktree_repo.git.add(file_path)
                    except Exception:
                        # Strategy 3: Remove conflict markers and keep both
                        try:
                            # Simple conflict marker removal - keep both versions

                            worktree_path = Path(worktree_repo.working_dir)
                            file_full_path = worktree_path / file_path

                            if file_full_path.exists():
                                with open(file_full_path, "r") as f:
                                    content = f.read()

                                # Remove conflict markers and keep both versions
                                lines = content.split("\n")
                                cleaned_lines = []
                                skip_next = False
                                in_conflict = False

                                for line in lines:
                                    if "<<<<<<<" in line or ">>>>>>>" in line:
                                        in_conflict = True
                                        continue
                                    elif "======" in line:
                                        skip_next = True
                                        in_conflict = False
                                        continue
                                    elif skip_next:
                                        skip_next = False
                                        continue
                                    elif not in_conflict:
                                        cleaned_lines.append(line)

                                with open(file_full_path, "w") as f:
                                    f.write("\n".join(cleaned_lines))

                                worktree_repo.git.add(file_path)
                        except Exception:
                            pass

            # Stage resolved files
            worktree_repo.git.add(".")

            # Commit the merge resolution
            worktree_repo.git.commit("-m", f"Auto-resolved conflicts during sync of {spec_id}")

        except Exception as e:
            # Auto-resolution failed, raise error for manual intervention
            raise GitOperationError(f"Auto-resolution failed for {spec_id}: {e}")
