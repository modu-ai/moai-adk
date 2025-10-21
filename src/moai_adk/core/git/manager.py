# @CODE:CORE-GIT-001 | SPEC: SPEC-CORE-GIT-001/spec.md | TEST: tests/unit/test_git.py
"""
Git repository management built on GitPython.

SPEC: .moai/specs/SPEC-CORE-GIT-001/spec.md
"""

from git import InvalidGitRepositoryError, Repo


class GitManager:
    """Manage interactions with a Git repository."""

    def __init__(self, repo_path: str = "."):
        """
        Initialize the GitManager.

        Args:
            repo_path: Path to the Git repository (default: current directory)

        Raises:
            InvalidGitRepositoryError: Raised when the path is not a Git repository.
        """
        self.repo = Repo(repo_path)
        self.git = self.repo.git

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
