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


def atomic_write_text(
    file_path: Path | str,
    content: str,
    encoding: str = "utf-8",
    make_dirs: bool = True,
) -> bool:
    """Atomically write text content to a file.

    Uses write-to-temp-then-rename pattern to prevent race conditions
    and partial writes from corrupting data.

    Args:
        file_path: Path to the file
        content: Content to write
        encoding: File encoding (default: utf-8)
        make_dirs: Create parent directories if needed (default: True)

    Returns:
        True if write succeeded, False otherwise
    """
    path = Path(file_path)

    if make_dirs:
        path.parent.mkdir(parents=True, exist_ok=True)

    try:
        # Write to temporary file first
        fd, temp_path = tempfile.mkstemp(dir=path.parent, prefix=".tmp_", suffix=path.suffix or ".tmp")
        try:
            # Write content to temp file
            with os.fdopen(fd, "w", encoding=encoding) as f:
                f.write(content)

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

    Returns:
        True if write succeeded, False otherwise
    """
    path = Path(file_path)

    if make_dirs:
        path.parent.mkdir(parents=True, exist_ok=True)

    try:
        # Write to temporary file first
        fd, temp_path = tempfile.mkstemp(dir=path.parent, prefix=".tmp_", suffix=".json")
        try:
            # Write JSON to temp file
            with os.fdopen(fd, "w", encoding=encoding) as f:
                json.dump(data, f, indent=indent, ensure_ascii=ensure_ascii)

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
