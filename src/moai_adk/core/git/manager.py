# @CODE:CORE-GIT-001 | SPEC: SPEC-CORE-GIT-001.md | TEST: tests/unit/test_git.py
"""
GitPython 기반 Git 저장소 관리.

SPEC: .moai/specs/SPEC-CORE-GIT-001/spec.md
"""

from git import InvalidGitRepositoryError, Repo


class GitManager:
    """Git 저장소 관리 클래스."""

    def __init__(self, repo_path: str = "."):
        """
        GitManager 초기화.

        Args:
            repo_path: Git 저장소 경로 (기본값: 현재 디렉토리)

        Raises:
            InvalidGitRepositoryError: Git 저장소가 아닐 경우
        """
        self.repo = Repo(repo_path)
        self.git = self.repo.git

    def is_repo(self) -> bool:
        """
        Git 저장소 여부 확인.

        Returns:
            Git 저장소이면 True, 아니면 False

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
        현재 브랜치명 반환.

        Returns:
            현재 활성 브랜치명

        Examples:
            >>> manager = GitManager()
            >>> manager.current_branch()
            'main'
        """
        return self.repo.active_branch.name

    def is_dirty(self) -> bool:
        """
        작업 디렉토리 변경사항 확인.

        Returns:
            변경사항이 있으면 True (dirty), 없으면 False (clean)

        Examples:
            >>> manager = GitManager()
            >>> manager.is_dirty()
            False
        """
        return self.repo.is_dirty()

    def create_branch(self, branch_name: str, from_branch: str | None = None) -> None:
        """
        새 브랜치 생성 및 전환.

        Args:
            branch_name: 생성할 브랜치명
            from_branch: 기준 브랜치 (기본값: None = 현재 브랜치)

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
        파일 스테이징 및 커밋.

        Args:
            message: 커밋 메시지
            files: 커밋할 파일 목록 (기본값: None = 모든 변경사항)

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
        원격 저장소에 푸시.

        Args:
            branch: 푸시할 브랜치 (기본값: None = 현재 브랜치)
            set_upstream: upstream 설정 여부

        Examples:
            >>> manager = GitManager()
            >>> manager.push(set_upstream=True)
        """
        if set_upstream:
            target_branch = branch or self.current_branch()
            self.git.push("--set-upstream", "origin", target_branch)
        else:
            self.git.push()
