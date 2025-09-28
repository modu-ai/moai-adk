#!/usr/bin/env python3
"""
@FEATURE:SYNC-EXECUTOR-001 Sync Execution Orchestrator
Dedicated module for coordinating version synchronization
"""

from pathlib import Path
from typing import Dict, List
import click

from .version_patterns import VersionPatternsProvider
from .file_processor import SecureFileProcessor
from moai_adk.utils.logger import get_logger

logger = get_logger(__name__)


class SyncExecutor:
    """@TASK:SYNC-EXECUTOR-001 Orchestrates version synchronization process"""

    def __init__(self, project_root: Path, current_version: str):
        """
        Initialize sync executor

        Args:
            project_root: Project root directory
            current_version: Current version for synchronization
        """
        self.project_root = project_root
        self.current_version = current_version
        self.patterns_provider = VersionPatternsProvider(current_version)
        self.file_processor = SecureFileProcessor(project_root)

    def sync_all_versions(self, dry_run: bool = False) -> Dict[str, List[str]]:
        """
        Execute complete version synchronization process

        Args:
            dry_run: If True, simulate changes without writing

        Returns:
            Dict mapping file patterns to changed files
        """
        results = {}

        logger.info(f"Starting version sync: v{self.current_version}, root: {self.project_root}")
        click.echo(f"🗿 MoAI-ADK version sync starting: v{self.current_version}")
        click.echo(f"Project root: {self.project_root}")

        patterns = self.patterns_provider.get_patterns()

        for pattern, rules in patterns.items():
            files_changed = self._sync_pattern(pattern, rules, dry_run)
            if files_changed:
                results[pattern] = files_changed

        if dry_run:
            logger.info("Dry run completed - no files were modified")
            click.echo("\\n✅ Dry run completed - no files were modified")
        else:
            logger.info("Version synchronization completed")
            click.echo("\\n✅ Version synchronization completed")

        return results

    def _sync_pattern(self, file_pattern: str, rules: List[Dict], dry_run: bool) -> List[str]:
        """
        Synchronize versions for specific file pattern

        Args:
            file_pattern: Pattern to match files
            rules: Replacement rules to apply
            dry_run: If True, don't write changes

        Returns:
            List of changed file paths
        """
        changed_files = []

        # Find files matching pattern
        if file_pattern.startswith("**"):
            files = list(self.project_root.glob(file_pattern))
        else:
            files = [self.project_root / file_pattern]
            files = [f for f in files if f.exists()]

        # Process each file
        for file_path in files:
            if self.file_processor.should_skip_file(file_path):
                continue

            try:
                changed = self.file_processor.process_file(file_path, rules, dry_run)
                if changed:
                    rel_path = str(file_path.relative_to(self.project_root))
                    changed_files.append(rel_path)
                    logger.info(f"Updated: {rel_path}")
                    click.echo(f"  ✓ {rel_path}")

            except Exception as e:
                rel_path = str(file_path.relative_to(self.project_root))
                logger.error(f"Update failed: {rel_path}: {e}")
                click.echo(f"  ❌ {rel_path}: {e}")

        return changed_files