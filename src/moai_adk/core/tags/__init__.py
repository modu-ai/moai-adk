#!/usr/bin/env python3
# @CODE:DOC-TAG-004 | TAG validation core module (Components 1 & 2)
"""TAG validation and management for MoAI-ADK

This module provides TAG validation functionality for:
- Pre-commit hook validation (Component 1)
- CI/CD pipeline validation (Component 2)
- TAG format checking
- Duplicate detection
- Orphan detection
"""

from .pre_commit_validator import (
    PreCommitValidator,
    ValidationResult,
    ValidationError,
    ValidationWarning,
)

from .ci_validator import CIValidator

__all__ = [
    "PreCommitValidator",
    "CIValidator",
    "ValidationResult",
    "ValidationError",
    "ValidationWarning",
]
