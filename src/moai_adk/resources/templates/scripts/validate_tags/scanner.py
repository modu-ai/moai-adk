#!/usr/bin/env python3
"""
File scanning logic for validate_tags module
"""

import re
from pathlib import Path
from typing import List
from .parser import TagReference


def scan_project_files(project_root: Path) -> List[TagReference]:
    """Scan project files for all tags"""
    tag_pattern = r'@([A-Z]+)[-:]([A-Z0-9-_]+)(?:\s+"([^"]*)")?'
    found_tags = []

    # File extensions to scan (excluding JSON)
    scan_extensions = [".md", ".py", ".js", ".ts", ".tsx", ".jsx", ".yml", ".yaml"]

    # Directories to exclude
    exclude_dirs = {
        "node_modules", "__pycache__", ".git", "dist",
        "build", "venv", ".env"
    }

    for file_path in project_root.rglob("*"):
        if file_path.is_file() and file_path.suffix in scan_extensions:
            # Check excluded directories
            if any(excluded in file_path.parts for excluded in exclude_dirs):
                continue

            try:
                content = file_path.read_text(encoding="utf-8", errors="ignore")
                lines = content.split("\n")

                for line_num, line in enumerate(lines, 1):
                    matches = re.finditer(tag_pattern, line)

                    for match in matches:
                        rel_path = file_path.relative_to(project_root)
                        found_tags.append(TagReference(
                            tag_type=match.group(1),
                            tag_id=match.group(2),
                            file_path=str(rel_path),
                            line_number=line_num,
                            context=line.strip()
                        ))

            except Exception:
                pass  # Skip files that can't be read

    return found_tags