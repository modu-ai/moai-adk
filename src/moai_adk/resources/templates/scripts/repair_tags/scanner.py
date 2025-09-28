#!/usr/bin/env python3
# @TASK:TAG-REPAIR-SCANNER-001
"""
TAG Scanner Module

Handles scanning and extracting TAG references from project files.
Focuses on efficient file scanning and TAG pattern recognition.
"""

import re
from pathlib import Path
from collections import defaultdict


class TagScanner:
    """Scans project files for TAG references."""

    def __init__(self, project_root: Path):
        self.project_root = project_root
        self.moai_path = project_root / ".moai"

    def scan_project_tags(self) -> dict[str, list[str]]:
        """Scan entire project for TAG references."""
        all_tags = defaultdict(list)

        # Scan .moai directory
        for md_file in self.moai_path.rglob("*.md"):
            if md_file.is_file():
                try:
                    content = md_file.read_text(encoding="utf-8", errors="ignore")
                    tags = self.extract_tags(content)
                    for tag in tags:
                        all_tags[tag].append(str(md_file.relative_to(self.project_root)))
                except (UnicodeDecodeError, PermissionError):
                    continue

        # Scan source code
        for py_file in self.project_root.rglob("*.py"):
            if ".moai" not in str(py_file) and py_file.is_file():
                try:
                    content = py_file.read_text(encoding="utf-8", errors="ignore")
                    tags = self.extract_tags(content)
                    for tag in tags:
                        all_tags[tag].append(str(py_file.relative_to(self.project_root)))
                except (UnicodeDecodeError, PermissionError):
                    continue

        return dict(all_tags)

    def extract_tags(self, content: str) -> list[str]:
        """Extract TAG patterns from content."""
        pattern = r"@([A-Z]+:[A-Z0-9_-]+)"
        matches = re.findall(pattern, content)
        return list(set(matches))  # Remove duplicates