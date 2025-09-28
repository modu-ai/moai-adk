#!/usr/bin/env python3
"""
@FEATURE:VERSION-SYNC-001 MoAI-ADK Automated Version Synchronization System
Legacy compatibility wrapper for modularized version sync system
"""

# Import from modularized version_sync package for backward compatibility
from .version_sync import VersionSyncManager, main

# Maintain backward compatibility
__all__ = ['VersionSyncManager', 'main']


if __name__ == "__main__":
    main()