#!/usr/bin/env python3
"""File utility functions for MoAI hooks.

Provides common file operations with safety checks.
"""

from __future__ import annotations

from pathlib import Path

# Default maximum file size for reading operations (10MB)
MAX_FILE_SIZE = 10 * 1024 * 1024  # 10MB in bytes


def check_file_size(file_path: Path | str, max_size: int = MAX_FILE_SIZE) -> tuple[bool, str]:
    """Check if file size is within acceptable limits.

    Args:
        file_path: Path to the file
        max_size: Maximum allowed file size in bytes (default: 10MB)

    Returns:
        Tuple of (is_safe, error_message)
        - is_safe: True if file is safe to read, False otherwise
        - error_message: Error message if unsafe, empty string if safe
    """
    path = Path(file_path)

    # Check if file exists
    if not path.exists():
        return False, f"File does not exist: {file_path}"

    # Check if it's a file (not directory)
    if not path.is_file():
        return False, f"Path is not a file: {file_path}"

    # Check file size
    try:
        file_size = path.stat().st_size
        if file_size > max_size:
            size_mb = file_size / (1024 * 1024)
            max_mb = max_size / (1024 * 1024)
            return (False, f"File too large: {size_mb:.2f}MB (max: {max_mb:.0f}MB): {file_path}")
    except OSError as e:
        return False, f"Cannot check file size: {e}"

    return True, ""


def safe_read_text(file_path: Path | str, max_size: int = MAX_FILE_SIZE) -> str | None:
    """Safely read text file with size validation.

    Args:
        file_path: Path to the file
        max_size: Maximum allowed file size in bytes (default: 10MB)

    Returns:
        File contents as string, or None if file is too large or cannot be read
    """
    is_safe, error_msg = check_file_size(file_path, max_size)
    if not is_safe:
        return None

    try:
        return Path(file_path).read_text(encoding="utf-8", errors="replace")
    except (OSError, UnicodeDecodeError):
        return None


def safe_read_binary(file_path: Path | str, max_size: int = MAX_FILE_SIZE) -> bytes | None:
    """Safely read binary file with size validation.

    Args:
        file_path: Path to the file
        max_size: Maximum allowed file size in bytes (default: 10MB)

    Returns:
        File contents as bytes, or None if file is too large or cannot be read
    """
    is_safe, error_msg = check_file_size(file_path, max_size)
    if not is_safe:
        return None

    try:
        return Path(file_path).read_bytes()
    except OSError:
        return None


__all__ = [
    "MAX_FILE_SIZE",
    "check_file_size",
    "safe_read_text",
    "safe_read_binary",
]
