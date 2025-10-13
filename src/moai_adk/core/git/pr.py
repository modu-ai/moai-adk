# @CODE:CORE-GIT-001 | SPEC: SPEC-CORE-GIT-001.md | TEST: tests/unit/test_git_utils.py
"""GitHub PR 관련 함수"""

import subprocess
from typing import TYPE_CHECKING, Any

if TYPE_CHECKING:
    from moai_adk.core.git.manager import GitManager


def create_draft_pr(
    title: str,
    body: str,
    base: str = "develop",
    head: str | None = None,
) -> str:
    """GitHub Draft PR 생성 (gh CLI 사용)

    Args:
        title: PR 제목
        body: PR 본문
        base: 기준 브랜치 (기본값: develop)
        head: 소스 브랜치 (None이면 현재 브랜치)

    Returns:
        생성된 PR의 URL

    Raises:
        subprocess.CalledProcessError: gh CLI 명령 실행 실패
        FileNotFoundError: gh CLI가 설치되지 않음

    Examples:
        >>> pr_url = create_draft_pr(
        ...     title="feat: Add new feature",
        ...     body="This PR adds...",
        ...     base="develop"
        ... )
        >>> print(pr_url)
        'https://github.com/user/repo/pull/123'

    Note:
        - gh CLI가 사전 설치되어 있어야 함
        - GitHub 인증이 완료되어 있어야 함
    """
    cmd = [
        "gh",
        "pr",
        "create",
        "--title",
        title,
        "--body",
        body,
        "--base",
        base,
        "--draft",
    ]

    if head:
        cmd.extend(["--head", head])

    result = subprocess.run(cmd, capture_output=True, text=True, check=True)
    return result.stdout.strip()


def get_repo_status(manager: "GitManager") -> dict[str, Any]:
    """저장소 상태 정보 반환

    Args:
        manager: GitManager 인스턴스

    Returns:
        저장소 상태를 담은 딕셔너리:
            - is_repo: Git 저장소 여부
            - current_branch: 현재 브랜치명
            - is_dirty: 변경사항 여부
            - untracked_files: 추적되지 않은 파일 목록
            - modified_files: 수정된 파일 목록

    Examples:
        >>> manager = GitManager()
        >>> status = get_repo_status(manager)
        >>> print(status["current_branch"])
        'develop'
        >>> print(status["untracked_files"])
        ['new_file.py']

    Note:
        - manager.is_repo()가 False면 일부 정보는 빈 값으로 반환됨
    """
    return {
        "is_repo": manager.is_repo(),
        "current_branch": manager.current_branch(),
        "is_dirty": manager.is_dirty(),
        "untracked_files": manager.repo.untracked_files if manager.repo else [],
        "modified_files": (
            [item.a_path for item in manager.repo.index.diff(None)] if manager.repo else []
        ),
    }
