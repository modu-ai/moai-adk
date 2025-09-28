#!/usr/bin/env python3
"""
@FEATURE:VERSION-SYNC-001 MoAI-ADK Version Synchronization Package
Modularized version synchronization system following TRUST principles

This package maintains backward compatibility with the original VersionSyncManager
while providing improved modularity and maintainability.
"""

import re
from pathlib import Path
import click

from moai_adk._version import __version__
from moai_adk.utils.logger import get_logger
from .version_patterns import VersionPatternsProvider
from .file_processor import SecureFileProcessor
from .sync_executor import SyncExecutor
from .sync_validator import SyncValidator
from .script_generator import ScriptGenerator

logger = get_logger(__name__)


class VersionSyncManager:
    """
    @TASK:VERSION-SYNC-MANAGER-001 Backward-compatible version sync manager

    Maintains the original API while leveraging modularized components
    """

    def __init__(self, project_root: str | None = None):
        """
        Initialize version sync manager

        Args:
            project_root: Project root directory. Auto-detected if None
        """
        self.project_root = (
            Path(project_root) if project_root else self._find_project_root()
        )
        self.current_version = __version__

        # Initialize modular components
        self.patterns_provider = VersionPatternsProvider(self.current_version)
        self.file_processor = SecureFileProcessor(self.project_root)
        self.sync_executor = SyncExecutor(self.project_root, self.current_version)
        self.sync_validator = SyncValidator(self.project_root, self.current_version)
        self.script_generator = ScriptGenerator(self.project_root)

        # Maintain original interface
        self.version_patterns = self.patterns_provider.get_patterns()
        self.sync_log = []

    def _find_project_root(self) -> Path:
        """Find project root directory containing pyproject.toml"""
        current = Path(__file__).parent

        while current != current.parent:
            if (current / "pyproject.toml").exists():
                return current
            current = current.parent

        raise FileNotFoundError("pyproject.toml을 찾을 수 없습니다")

    def sync_all_versions(self, dry_run: bool = False) -> dict[str, list[str]]:
        """
        Synchronize all version information across the project

        Args:
            dry_run: If True, simulate changes without writing

        Returns:
            Dict mapping file patterns to changed files
        """
        return self.sync_executor.sync_all_versions(dry_run)

    def verify_sync(self) -> dict[str, list[str]]:
        """
        Verify version synchronization - detect remaining inconsistencies

        Returns:
            Dict mapping patterns to files with mismatches
        """
        return self.sync_validator.verify_sync()

    def create_version_update_script(self) -> str:
        """
        Generate version update automation script

        Returns:
            Path to generated script
        """
        return self.script_generator.create_version_update_script()

    # Legacy methods for backward compatibility
    def _load_version_patterns(self) -> dict[str, list[dict]]:
        """Legacy method - use patterns_provider instead"""
        return self.patterns_provider.get_patterns()

    def _sync_pattern(self, file_pattern: str, replacements: list[dict], dry_run: bool) -> list[str]:
        """Legacy method - use sync_executor instead"""
        return self.sync_executor._sync_pattern(file_pattern, replacements, dry_run)

    def _should_skip_file(self, file_path: Path) -> bool:
        """Legacy method - use file_processor instead"""
        return self.file_processor.should_skip_file(file_path)

    def _sync_file(self, file_path: Path, replacements: list[dict], dry_run: bool) -> bool:
        """Legacy method - use file_processor instead"""
        return self.file_processor.process_file(file_path, replacements, dry_run)

    def _find_version_mismatches(self, pattern: str, expected: str) -> list[str]:
        """Legacy method - use sync_validator instead"""
        return self.sync_validator.find_pattern_mismatches(pattern, expected)


# Export main classes for external use
__all__ = [
    'VersionSyncManager',
    'VersionPatternsProvider',
    'SecureFileProcessor',
    'SyncExecutor',
    'SyncValidator',
    'ScriptGenerator'
]


def main() -> None:
    """CLI entry point"""
    import argparse

    parser = argparse.ArgumentParser(description="MoAI-ADK version synchronization tool")
    parser.add_argument(
        "--dry-run", action="store_true", help="Simulate changes without writing"
    )
    parser.add_argument("--verify", action="store_true", help="Verify sync only")
    parser.add_argument(
        "--create-script", action="store_true", help="Generate version update script"
    )

    args = parser.parse_args()

    manager = VersionSyncManager()

    if args.verify:
        manager.verify_sync()
    elif args.create_script:
        manager.create_version_update_script()
    else:
        manager.sync_all_versions(dry_run=args.dry_run)
        if not args.dry_run:
            manager.verify_sync()


if __name__ == "__main__":
    main()