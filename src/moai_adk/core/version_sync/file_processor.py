#!/usr/bin/env python3
"""
@FEATURE:FILE-PROCESSOR-001 Secure File Processing
Dedicated module for secure file processing with validation
"""

import re
from pathlib import Path
from typing import List, Dict

from moai_adk.utils.logger import get_logger

logger = get_logger(__name__)


class SecureFileProcessor:
    """@TASK:FILE-PROCESSOR-001 Secure file processor with validation"""

    def __init__(self, project_root: Path):
        """
        Initialize secure file processor

        Args:
            project_root: Project root directory
        """
        self.project_root = project_root
        self.skip_patterns = [
            ".git/",
            "__pycache__/",
            ".pytest_cache/",
            "node_modules/",
            ".venv/",
            "venv/",
            "dist/",
            "build/",
            ".mypy_cache/",
        ]

    def should_skip_file(self, file_path: Path) -> bool:
        """
        Check if file should be skipped based on security patterns

        Args:
            file_path: Path to file to check

        Returns:
            True if file should be skipped
        """
        file_str = str(file_path)
        return any(pattern in file_str for pattern in self.skip_patterns)

    def apply_replacements(self, content: str, rules: List[Dict]) -> str:
        """
        Apply version replacements to content

        Args:
            content: File content to process
            rules: List of replacement rules

        Returns:
            Modified content
        """
        modified_content = content

        for rule in rules:
            pattern = rule["pattern"]
            replacement = rule["replacement"]

            if re.search(pattern, modified_content):
                modified_content = re.sub(pattern, replacement, modified_content)

        return modified_content

    def process_file(self, file_path: Path, rules: List[Dict], dry_run: bool = False) -> bool:
        """
        Process single file with version replacements

        Args:
            file_path: Path to file to process
            rules: List of replacement rules
            dry_run: If True, don't write changes

        Returns:
            True if changes were made
        """
        try:
            with open(file_path, encoding="utf-8") as f:
                content = f.read()
        except UnicodeDecodeError:
            # Skip binary files
            logger.debug(f"Skipping binary file: {file_path}")
            return False

        original_content = content
        modified_content = self.apply_replacements(content, rules)

        if modified_content != original_content:
            if not dry_run:
                with open(file_path, "w", encoding="utf-8") as f:
                    f.write(modified_content)
            return True

        return False