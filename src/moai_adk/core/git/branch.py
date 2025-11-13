"""
Branch naming utilities.

SPEC: .moai/specs/SPEC-CORE-GIT-001/spec.md
"""


def generate_branch_name(spec_id: str) -> str:
    """
    Generate a branch name from a SPEC ID.

    Args:
        spec_id: SPEC identifier (e.g., "AUTH-001").

    Returns:
        Branch name in the feature/SPEC-XXX format.

    Examples:
        >>> generate_branch_name("AUTH-001")
        'feature/SPEC-AUTH-001'

        >>> generate_branch_name("CORE-GIT-001")
        'feature/SPEC-CORE-GIT-001'
    """
    return f"feature/SPEC-{spec_id}"
