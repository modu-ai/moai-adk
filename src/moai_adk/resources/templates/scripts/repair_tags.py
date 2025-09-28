#!/usr/bin/env python3
# @TASK:TAG-REPAIR-011
"""
MoAI-ADK TAG Auto-Repair System

Backward compatibility wrapper for the modularized TAG repair system.
The actual implementation is now in the repair_tags/ module.
"""

# Backward compatibility imports
from .repair_tags.core import TagRepairer
from .repair_tags.main import main

# Legacy compatibility
def auto_repair_tags(*args, **kwargs):
    """Legacy function for backward compatibility."""
    from pathlib import Path
    project_root = kwargs.get('project_root', Path.cwd())
    repairer = TagRepairer(project_root)
    return repairer.auto_repair_tags(*args, **kwargs)

if __name__ == "__main__":
    exit(main())