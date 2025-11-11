"""
Migration module for MoAI-ADK version upgrades

Handles automatic migration of configuration files and project structure
when upgrading between versions.
"""

from .version_migrator import VersionMigrator
from .version_detector import VersionDetector
from .backup_manager import BackupManager
from .file_migrator import FileMigrator

__all__ = [
    "VersionMigrator",
    "VersionDetector",
    "BackupManager",
    "FileMigrator",
]
