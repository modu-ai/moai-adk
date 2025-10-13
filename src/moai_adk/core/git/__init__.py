# @CODE:CORE-GIT-001 | SPEC: SPEC-CORE-GIT-001.md | TEST: tests/unit/test_git_manager.py, tests/unit/test_git_utils.py
"""
Git 모듈

GitPython 기반 Git 워크플로우 관리 시스템

주요 기능:
- GitManager: Git 저장소 조작 (브랜치, 커밋, 푸시)
- Branch utils: 브랜치명 생성
- Commit utils: TDD 커밋 메시지 포맷팅
- PR utils: GitHub Draft PR 생성 및 저장소 상태 조회
"""

from moai_adk.core.git.branch import generate_branch_name
from moai_adk.core.git.commit import format_commit_message
from moai_adk.core.git.manager import GitManager
from moai_adk.core.git.pr import create_draft_pr, get_repo_status

__all__ = [
    "GitManager",
    "generate_branch_name",
    "format_commit_message",
    "create_draft_pr",
    "get_repo_status",
]
