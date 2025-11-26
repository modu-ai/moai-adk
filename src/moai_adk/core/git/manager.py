"""
Git repository management built on GitPython.

SPEC: .moai/specs/SPEC-CORE-GIT-001/spec.md
"""

from pathlib import Path

from git import InvalidGitRepositoryError, Repo

from moai_adk.core.git.conflict_detector import GitConflictDetector


class GitManager:
    """Manage interactions with a Git repository."""

    def __init__(self, repo_path: str | Path = "."):
        """
        Initialize the GitManager.

        Args:
            repo_path: Path to the Git repository (default: current directory)

        Raises:
            InvalidGitRepositoryError: Raised when the path is not a Git repository.
        """
        self.repo = Repo(repo_path)
        self.git = self.repo.git
        self.repo_path = Path(repo_path).resolve()
        self.conflict_detector = GitConflictDetector(self.repo_path)

    def is_repo(self) -> bool:
        """
        Check whether the path points to a Git repository.

        Returns:
            True when the location is a Git repository, otherwise False.

        Examples:
            >>> manager = GitManager("/path/to/repo")
            >>> manager.is_repo()
            True
        """
        try:
            _ = self.repo.git_dir
            return True
        except (InvalidGitRepositoryError, Exception):
            return False

    def current_branch(self) -> str:
        """
        Return the active branch name.

        Returns:
            Name of the currently checked-out branch.

        Examples:
            >>> manager = GitManager()
            >>> manager.current_branch()
            'main'
        """
        return self.repo.active_branch.name

    def is_dirty(self) -> bool:
        """
        Check whether the working tree has uncommitted changes.

        Returns:
            True when the worktree is dirty, otherwise False.

        Examples:
            >>> manager = GitManager()
            >>> manager.is_dirty()
            False
        """
        return self.repo.is_dirty()

    def create_branch(self, branch_name: str, from_branch: str | None = None) -> None:
        """
        Create and switch to a new branch.

        Args:
            branch_name: Name of the branch to create.
            from_branch: Base branch (default: current branch).

        Examples:
            >>> manager = GitManager()
            >>> manager.create_branch("feature/SPEC-AUTH-001")
            >>> manager.current_branch()
            'feature/SPEC-AUTH-001'
        """
        if from_branch:
            self.git.checkout("-b", branch_name, from_branch)
        else:
            self.git.checkout("-b", branch_name)

    def commit(self, message: str, files: list[str] | None = None) -> None:
        """
        Stage files and create a commit.

        Args:
            message: Commit message.
            files: Optional list of files to commit (default: all changes).

        Examples:
            >>> manager = GitManager()
            >>> manager.commit("feat: add authentication", files=["auth.py"])
        """
        if files:
            self.repo.index.add(files)
        else:
            self.git.add(A=True)

        self.repo.index.commit(message)

    def push(self, branch: str | None = None, set_upstream: bool = False) -> None:
        """
        Push commits to the remote repository.

        Args:
            branch: Branch to push (default: current branch).
            set_upstream: Whether to set the upstream tracking branch.

        Examples:
            >>> manager = GitManager()
            >>> manager.push(set_upstream=True)
        """
        if set_upstream:
            target_branch = branch or self.current_branch()
            self.git.push("--set-upstream", "origin", target_branch)
        else:
            self.git.push()

    def check_merge_conflicts(self, feature_branch: str, base_branch: str) -> dict:
        """
        Check if merge is possible without conflicts.

        Args:
            feature_branch: Feature branch to merge from
            base_branch: Base branch to merge into

        Returns:
            Dictionary with merge status and conflict information

        Examples:
            >>> manager = GitManager()
            >>> result = manager.check_merge_conflicts("feature/auth", "develop")
            >>> if result["can_merge"]:
            ...     print("Ready to merge")
            ... else:
            ...     print(f"Conflicts: {result['conflicts']}")
        """
        return self.conflict_detector.can_merge(feature_branch, base_branch)

    def has_merge_conflicts(self, feature_branch: str, base_branch: str) -> bool:
        """
        Quick check if merge would have conflicts.

        Args:
            feature_branch: Feature branch to merge from
            base_branch: Base branch to merge into

        Returns:
            True if conflicts exist, False otherwise

        Examples:
            >>> manager = GitManager()
            >>> if manager.has_merge_conflicts("feature/auth", "develop"):
            ...     print("Conflicts detected")
        """
        result = self.conflict_detector.can_merge(feature_branch, base_branch)
        return not result.get("can_merge", False)

    def get_conflict_summary(self, feature_branch: str, base_branch: str) -> str:
        """
        Get human-readable summary of merge conflicts.

        Args:
            feature_branch: Feature branch to merge from
            base_branch: Base branch to merge into

        Returns:
            String summary of conflicts for user presentation

        Examples:
            >>> manager = GitManager()
            >>> summary = manager.get_conflict_summary("feature/auth", "develop")
            >>> print(summary)
        """
        result = self.conflict_detector.can_merge(feature_branch, base_branch)
        conflicts = result.get("conflicts", [])
        return self.conflict_detector.summarize_conflicts(conflicts)  # type: ignore[arg-type]

    def auto_resolve_safe_conflicts(self) -> bool:
        """
        Auto-resolve safe config file conflicts.

        Returns:
            True if auto-resolution succeeded, False otherwise

        Examples:
            >>> manager = GitManager()
            >>> if manager.auto_resolve_safe_conflicts():
            ...     print("Safe conflicts resolved automatically")
        """
        return self.conflict_detector.auto_resolve_safe()

    def abort_merge(self) -> None:
        """
        Abort an in-progress merge and clean up state.

        Examples:
            >>> manager = GitManager()
            >>> manager.abort_merge()
        """
        self.conflict_detector.cleanup_merge_state()
