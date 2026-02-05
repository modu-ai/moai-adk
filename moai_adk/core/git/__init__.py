"""
Git management module.

Manage Git workflows using GitPython.

SPEC: .moai/specs/SPEC-CORE-GIT-001/spec.md
"""

from moai_adk.core.git.branch import generate_branch_name
from moai_adk.core.git.branch_manager import BranchManager
from moai_adk.core.git.checkpoint import CheckpointManager
from moai_adk.core.git.commit import format_commit_message
from moai_adk.core.git.conflict_detector import (
    ConflictFile,
    ConflictSeverity,
    GitConflictDetector,
)
from moai_adk.core.git.event_detector import EventDetector
from moai_adk.core.git.manager import GitManager

__all__ = [
    "GitManager",
    "generate_branch_name",
    "format_commit_message",
    "BranchManager",
    "CheckpointManager",
    "EventDetector",
    "GitConflictDetector",
    "ConflictFile",
    "ConflictSeverity",
]
