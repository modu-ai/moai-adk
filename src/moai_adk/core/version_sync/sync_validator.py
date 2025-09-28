#!/usr/bin/env python3
"""
@FEATURE:SYNC-VALIDATOR-001 Sync Validation System
Dedicated module for version consistency validation
"""

import re
from pathlib import Path
from typing import Dict, List
import click

from .file_processor import SecureFileProcessor
from moai_adk.utils.logger import get_logger

logger = get_logger(__name__)


class SyncValidator:
    """@TASK:SYNC-VALIDATOR-001 Validates version synchronization consistency"""

    def __init__(self, project_root: Path, current_version: str):
        """
        Initialize sync validator

        Args:
            project_root: Project root directory
            current_version: Expected version for validation
        """
        self.project_root = project_root
        self.current_version = current_version
        self.file_processor = SecureFileProcessor(project_root)

    def verify_sync(self) -> Dict[str, List[str]]:
        """
        Verify version synchronization - detect remaining inconsistencies

        Returns:
            Dict mapping patterns to files with mismatches
        """
        logger.info("Starting version sync verification")
        click.echo("\\n🔍 Verifying version synchronization...")

        inconsistent_files = {}

        # Define validation patterns
        validation_patterns = [
            (r"v[0-9]+\.[0-9]+\.[0-9]+", f"v{self.current_version}"),
            (r"version.*[0-9]+\.[0-9]+\.[0-9]+", f"version {self.current_version}"),
            (r"MoAI-ADK v[0-9]+\.[0-9]+\.[0-9]+", f"MoAI-ADK v{self.current_version}"),
        ]

        for pattern, expected in validation_patterns:
            mismatches = self.find_pattern_mismatches(pattern, expected)
            if mismatches:
                inconsistent_files[pattern] = mismatches

        if inconsistent_files:
            logger.warning(f"Version mismatches found: {len(inconsistent_files)} patterns")
            click.echo("⚠️  Version mismatches found in the following files:")
            for pattern, files in inconsistent_files.items():
                click.echo(f"  Pattern: {pattern}")
                for file_info in files:
                    click.echo(f"    {file_info}")
        else:
            logger.info("All version information is consistent")
            click.echo("✅ All version information is consistent")

        return inconsistent_files

    def find_pattern_mismatches(self, pattern: str, expected: str) -> List[str]:
        """
        Find files with version pattern mismatches

        Args:
            pattern: Regex pattern to search for
            expected: Expected version string

        Returns:
            List of files with mismatches
        """
        mismatches = []

        for file_path in self.project_root.glob("**/*"):
            if not file_path.is_file() or self.file_processor.should_skip_file(file_path):
                continue

            try:
                with open(file_path, encoding="utf-8") as f:
                    content = f.read()

                matches = re.findall(pattern, content, re.IGNORECASE)
                unexpected_matches = [m for m in matches if m != expected.split()[-1]]

                if unexpected_matches:
                    rel_path = file_path.relative_to(self.project_root)
                    mismatches.append(f"{rel_path}: {unexpected_matches}")

            except (UnicodeDecodeError, OSError):
                continue

        return mismatches