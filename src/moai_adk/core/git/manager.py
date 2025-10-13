# @CODE:CORE-GIT-001 | SPEC: SPEC-CORE-GIT-001.md | TEST: tests/unit/test_git_manager.py
"""GitManager 클래스 - Git 저장소 조작"""

from pathlib import Path

from git import InvalidGitRepositoryError, Repo


class GitManager:
    """GitPython 기반 Git 저장소 관리 클래스

    Attributes:
        repo: GitPython Repo 객체
        git: GitPython Git 명령 인터페이스

    Note:
        - 유효하지 않은 Git 저장소일 경우 repo, git이 None으로 설정됨
        - is_repo() 메서드로 유효성 확인 필수
    """

    def __init__(self, repo_path: str | Path = ".") -> None:
        """GitManager 초기화

        Args:
            repo_path: Git 저장소 경로 (기본값: 현재 디렉토리)
        """
        try:
            self.repo: Repo | None = Repo(str(repo_path))
            self.git = self.repo.git
        except (InvalidGitRepositoryError, Exception):
            # 유효하지 않은 Git 저장소일 경우 None으로 설정
            self.repo = None
            self.git = None

    def is_repo(self) -> bool:
        """Git 저장소 여부 확인

        Returns:
            유효한 Git 저장소면 True, 아니면 False

        Examples:
            >>> manager = GitManager("/path/to/repo")
            >>> if manager.is_repo():
            ...     print("Valid Git repository")
        """
        try:
            if self.repo is None:
                return False
            _ = self.repo.git_dir
            return True
        except Exception:
            return False

    def current_branch(self) -> str:
        """현재 브랜치명 반환

        Returns:
            현재 활성 브랜치명

        Raises:
            AttributeError: repo가 None인 경우
            TypeError: detached HEAD 상태인 경우

        Examples:
            >>> manager = GitManager()
            >>> manager.current_branch()
            'develop'
        """
        if self.repo is None:
            raise AttributeError("Not a valid Git repository")
        return str(self.repo.active_branch.name)

    def is_dirty(self) -> bool:
        """작업 디렉토리 변경사항 확인

        Returns:
            수정/추가/삭제된 파일이 있으면 True, 없으면 False

        Raises:
            AttributeError: repo가 None인 경우

        Examples:
            >>> manager = GitManager()
            >>> if manager.is_dirty():
            ...     print("Working directory has changes")
        """
        if self.repo is None:
            raise AttributeError("Not a valid Git repository")
        return bool(self.repo.is_dirty())

    def create_branch(self, branch_name: str, from_branch: str = "develop") -> None:
        """새 브랜치 생성 및 전환

        Args:
            branch_name: 생성할 브랜치명 (예: feature/SPEC-AUTH-001)
            from_branch: 기준 브랜치 (기본값: develop)

        Raises:
            GitCommandError: Git 명령 실행 실패
            AttributeError: git이 None인 경우

        Examples:
            >>> manager = GitManager()
            >>> manager.create_branch("feature/new-feature", from_branch="main")
        """
        if self.git is None:
            raise AttributeError("Not a valid Git repository")
        self.git.checkout("-b", branch_name, from_branch)

    def commit(self, message: str, files: list[str] | None = None) -> None:
        """파일 스테이징 및 커밋

        Args:
            message: 커밋 메시지
            files: 커밋할 파일 목록 (None이면 모든 변경사항)

        Raises:
            GitCommandError: Git 명령 실행 실패
            AttributeError: repo가 None인 경우

        Examples:
            >>> manager = GitManager()
            >>> manager.commit("feat: Add new feature", files=["src/feature.py"])
            >>> manager.commit("docs: Update README")  # 모든 변경사항
        """
        if self.repo is None:
            raise AttributeError("Not a valid Git repository")

        if files:
            self.repo.index.add(files)
        else:
            self.repo.git.add(A=True)

        self.repo.index.commit(message)

    def push(self, branch: str | None = None, set_upstream: bool = False) -> None:
        """원격 저장소에 푸시

        Args:
            branch: 푸시할 브랜치명 (None이면 현재 브랜치)
            set_upstream: upstream 설정 여부 (기본값: False)

        Raises:
            GitCommandError: Git 명령 실행 실패
            AttributeError: git이 None인 경우

        Examples:
            >>> manager = GitManager()
            >>> manager.push(set_upstream=True)  # 현재 브랜치를 upstream 설정
            >>> manager.push(branch="feature/new-feature")
        """
        if self.git is None:
            raise AttributeError("Not a valid Git repository")

        if set_upstream:
            target_branch = branch or self.current_branch()
            self.git.push("--set-upstream", "origin", target_branch)
        else:
            self.git.push()
