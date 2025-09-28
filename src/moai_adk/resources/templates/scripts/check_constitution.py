#!/usr/bin/env python3
# @TASK:CONSTITUTION-CHECK-011
"""
MoAI-ADK TRUST 5 Principles Verification System

Backward compatibility wrapper for the modularized TRUST principles checker.
The actual implementation is now in the check_constitution/ module.
"""

# Backward compatibility imports
from .check_constitution.core import TrustPrinciplesChecker
from .check_constitution.main import main

# Legacy compatibility
def check_trust_principles(*args, **kwargs):
    """Legacy function for backward compatibility."""
    from pathlib import Path
    project_root = kwargs.get('project_root', Path.cwd())
    checker = TrustPrinciplesChecker(project_root, kwargs.get('verbose', False))
    return checker.run_full_check()

if __name__ == "__main__":
    exit(main())