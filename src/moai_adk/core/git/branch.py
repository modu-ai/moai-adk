# @CODE:CORE-GIT-001 | SPEC: SPEC-CORE-GIT-001.md | TEST: tests/unit/test_git_utils.py
"""브랜치 유틸리티 함수"""
from typing import Literal

def generate_branch_name(
    spec_id: str,
    branch_type: Literal['feature', 'bugfix', 'hotfix', 'release'] = 'feature'
) -> str:
    """
    SPEC ID로부터 브랜치명 생성

    Args:
        spec_id: SPEC ID (예: AUTH-001, CORE-GIT-001)
        branch_type: 브랜치 유형 (기본값: 'feature')

    Returns:
        브랜치명 (예: feature/SPEC-CORE-GIT-001)

    Examples:
        >>> generate_branch_name("AUTH-001")
        'feature/SPEC-AUTH-001'
        >>> generate_branch_name("CORE-GIT-001", branch_type="bugfix")
        'bugfix/SPEC-CORE-GIT-001'
    """
    return f"{branch_type}/SPEC-{spec_id}"