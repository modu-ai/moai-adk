#!/usr/bin/env python3
"""Atomic file operations for MoAI hooks.

Provides safe file write operations with atomic semantics
to prevent race conditions and data corruption.
"""

from __future__ import annotations

import json
import os
import tempfile
from pathlib import Path
from typing import Any

# Maximum file size for state/cache files (1MB)
MAX_STATE_FILE_SIZE = 1024 * 1024  # 1MB


def _sanitize_surrogates(text: str) -> str:
    """Remove or replace surrogate characters that cannot be encoded in UTF-8.

    Surrogate characters (U+D800-U+DFFF) are used in UTF-16 but cannot be
    encoded in UTF-8. This function filters them out to prevent encoding errors.

    Args:
        text: Input text that may contain surrogate characters

    Returns:
        Text with surrogates replaced with the Unicode replacement character (U+FFFD)
    """
    # Replace surrogates with replacement character
    return text.encode("utf-8", errors="replace").decode("utf-8")


def atomic_write_text(
    file_path: Path | str,
    content: str,
    encoding: str = "utf-8",
    make_dirs: bool = True,
    sanitize_surrogates: bool = True,
) -> bool:
    """Atomically write text content to a file.

    Uses write-to-temp-then-rename pattern to prevent race conditions
    and partial writes from corrupting data.

    Args:
        file_path: Path to the file
        content: Content to write
        encoding: File encoding (default: utf-8)
        make_dirs: Create parent directories if needed (default: True)
        sanitize_surrogates: Remove surrogate characters before writing (default: True)

    Returns:
        True if write succeeded, False otherwise
    """
    path = Path(file_path)

    if make_dirs:
        path.parent.mkdir(parents=True, exist_ok=True)

    # Sanitize content to remove surrogates if requested
    write_content = _sanitize_surrogates(content) if sanitize_surrogates else content

    try:
        # Write to temporary file first
        fd, temp_path = tempfile.mkstemp(dir=path.parent, prefix=".tmp_", suffix=path.suffix or ".tmp")
        try:
            # Write content to temp file with error handling for surrogates
            with os.fdopen(fd, "w", encoding=encoding, errors="replace") as f:
                f.write(write_content)

            # Atomic rename (overwrites target if exists)
            os.replace(temp_path, path)
            return True
        except Exception:
            # Clean up temp file on error
            try:
                os.unlink(temp_path)
            except OSError:
                pass
            return False
    except (OSError, ValueError):
        return False


def atomic_write_json(
    file_path: Path | str,
    data: Any,
    indent: int = 2,
    encoding: str = "utf-8",
    ensure_ascii: bool = True,
    make_dirs: bool = True,
    sanitize_surrogates: bool = True,
) -> bool:
    """Atomically write JSON data to a file.

    Uses write-to-temp-then-rename pattern to prevent race conditions
    and partial writes from corrupting data.

    Args:
        file_path: Path to the file
        data: Data to serialize as JSON
        indent: JSON indentation (default: 2)
        encoding: File encoding (default: utf-8)
        ensure_ascii: Escape non-ASCII characters (default: True)
        make_dirs: Create parent directories if needed (default: True)
        sanitize_surrogates: Remove surrogate characters before writing (default: True)

    Returns:
        True if write succeeded, False otherwise
    """
    path = Path(file_path)

    if make_dirs:
        path.parent.mkdir(parents=True, exist_ok=True)

    try:
        # Serialize JSON to string first (allows us to sanitize)
        json_str = json.dumps(data, indent=indent, ensure_ascii=ensure_ascii)

        # Sanitize to remove surrogates if requested
        if sanitize_surrogates:
            json_str = _sanitize_surrogates(json_str)

        # Write to temporary file first
        fd, temp_path = tempfile.mkstemp(dir=path.parent, prefix=".tmp_", suffix=".json")
        try:
            # Write JSON to temp file with error handling for surrogates
            with os.fdopen(fd, "w", encoding=encoding, errors="replace") as f:
                f.write(json_str)

            # Atomic rename (overwrites target if exists)
            os.replace(temp_path, path)
            return True
        except Exception:
            # Clean up temp file on error
            try:
                os.unlink(temp_path)
            except OSError:
                pass
            return False
    except (OSError, ValueError, TypeError):
        return False


__all__ = [
    "MAX_STATE_FILE_SIZE",
    "atomic_write_text",
    "atomic_write_json",
]
