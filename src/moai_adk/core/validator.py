#!/usr/bin/env python3
# @FEATURE:VALIDATOR-001
"""
MoAI-ADK Validation Utilities

Backward compatibility wrapper for the modularized validation system.
The actual implementation is now in the validator/ module.
"""

# Backward compatibility imports
from .validator.environment import (
    validate_python_version,
    validate_claude_code,
    validate_git_repository,
    validate_environment
)
from .validator.project import (
    validate_project_structure,
    validate_project_readiness
)
from .validator.structure import validate_moai_structure
from .validator.compliance import validate_trust_principles_compliance

# Export all validation functions for backward compatibility
__all__ = [
    "validate_python_version",
    "validate_claude_code",
    "validate_git_repository",
    "validate_environment",
    "validate_project_structure",
    "validate_project_readiness",
    "validate_moai_structure",
    "validate_trust_principles_compliance",
]