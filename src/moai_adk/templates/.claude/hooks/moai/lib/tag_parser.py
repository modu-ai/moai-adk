"""TAG parser module for TAG System v2.0.

Implements T2: TAG Parser for extracting TAG annotations from
Python source code comments using AST parsing.
"""

from __future__ import annotations

import ast
import logging
from pathlib import Path
from typing import List

from . import tag_validator

logger = logging.getLogger(__name__)


def extract_tags_from_source(source: str, file_path: Path) -> List[tag_validator.TAG]:
    """Extract all @SPEC TAGs from Python source code.

    Args:
        source: Python source code as string
        file_path: Path to the source file

    Returns:
        List of TAG objects found in the source

    Examples:
        >>> source = "# @SPEC SPEC-AUTH-001\\ndef func(): pass"
        >>> tags = extract_tags_from_source(source, Path("test.py"))
        >>> len(tags)
        1
    """
    tags: List[tag_validator.TAG] = []

    if not isinstance(source, str):
        return tags

    # Split source into lines and extract TAGs from comments
    lines = source.splitlines()

    for line_number, line in enumerate(lines, start=1):
        # Handle both standalone comments and inline comments
        # Find the comment portion if it exists
        comment_start = line.find("#")

        if comment_start == -1:
            # No comment in this line
            continue

        # Extract the comment portion (from # to end of line)
        comment_portion = line[comment_start:]

        # Parse TAG from comment
        tag = tag_validator.parse_tag_string(comment_portion, file_path, line_number)

        if tag is not None:
            tags.append(tag)

    return tags


def extract_tags_from_file(file_path: Path) -> List[tag_validator.TAG]:
    """Extract all @SPEC TAGs from a Python file.

    Args:
        file_path: Path to the Python file

    Returns:
        List of TAG objects found in the file, empty list if file doesn't exist or has errors

    Examples:
        >>> tags = extract_tags_from_file(Path("auth.py"))
        >>> len(tags)
        2
    """
    if not isinstance(file_path, Path):
        file_path = Path(file_path)

    # Check if file exists
    if not file_path.exists():
        logger.warning(f"File not found: {file_path}")
        return []

    # Check if file is readable
    if not file_path.is_file():
        logger.warning(f"Not a file: {file_path}")
        return []

    try:
        # Read file content
        source = file_path.read_text(encoding="utf-8")

        # Validate Python syntax
        try:
            ast.parse(source)
        except SyntaxError as e:
            logger.warning(f"Syntax error in {file_path}: {e}")
            return []

        # Extract TAGs
        return extract_tags_from_source(source, file_path)

    except Exception as e:
        logger.error(f"Error reading file {file_path}: {e}")
        return []


def extract_tags_from_files(file_paths: List[Path]) -> List[tag_validator.TAG]:
    """Extract TAGs from multiple files.

    Args:
        file_paths: List of file paths to process

    Returns:
        List of all TAGs found across all files

    Examples:
        >>> files = [Path("a.py"), Path("b.py")]
        >>> tags = extract_tags_from_files(files)
        >>> len(tags)
        5
    """
    all_tags: List[tag_validator.TAG] = []

    for file_path in file_paths:
        tags = extract_tags_from_file(file_path)
        all_tags.extend(tags)

    return all_tags


def extract_tags_from_directory(
    directory: Path,
    pattern: str = "*.py",
    recursive: bool = True,
) -> List[tag_validator.TAG]:
    """Extract TAGs from all Python files in a directory.

    Args:
        directory: Directory path to search
        pattern: File pattern to match (default: "*.py")
        recursive: Whether to search recursively (default: True)

    Returns:
        List of all TAGs found in matching files

    Examples:
        >>> tags = extract_tags_from_directory(Path("src/"))
        >>> len(tags)
        42
    """
    if not isinstance(directory, Path):
        directory = Path(directory)

    if not directory.exists() or not directory.is_dir():
        logger.warning(f"Not a directory: {directory}")
        return []

    # Find matching files
    if recursive:
        files = list(directory.rglob(pattern))
    else:
        files = list(directory.glob(pattern))

    # Extract TAGs from all files
    return extract_tags_from_files(files)


__all__ = [
    "extract_tags_from_source",
    "extract_tags_from_file",
    "extract_tags_from_files",
    "extract_tags_from_directory",
]
