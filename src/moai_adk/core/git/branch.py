# @CODE:CORE-GIT-001 | SPEC: SPEC-CORE-GIT-001.md | TEST: tests/unit/test_git_utils.py
"""브랜치 유틸리티 함수"""


def generate_branch_name(spec_id: str) -> str:
    """SPEC ID로부터 브랜치명 생성

    Args:
        spec_id: SPEC ID (예: AUTH-001, CORE-GIT-001)

    Returns:
        feature/SPEC-{ID} 형식의 브랜치명

    Examples:
        >>> generate_branch_name("AUTH-001")
        'feature/SPEC-AUTH-001'
        >>> generate_branch_name("CORE-GIT-001")
        'feature/SPEC-CORE-GIT-001'
    """
    return f"feature/SPEC-{spec_id}"
