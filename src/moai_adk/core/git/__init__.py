# @CODE:CORE-GIT-001 | SPEC: SPEC-CORE-GIT-001.md | TEST: tests/unit/test_git.py
"""
Git management module.

GitPython 기반 Git 워크플로우 관리 시스템.

SPEC: .moai/specs/SPEC-CORE-GIT-001/spec.md
"""

from moai_adk.core.git.branch import generate_branch_name
from moai_adk.core.git.commit import format_commit_message
from moai_adk.core.git.manager import GitManager

__all__ = [
    "GitManager",
    "generate_branch_name",
    "format_commit_message",
]
