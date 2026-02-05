"""Path conversion utilities for WSL (Windows Subsystem for Linux) support.

This module provides functions to convert between Windows and WSL path formats,
enabling seamless cross-platform compatibility when running MoAI-ADK in WSL.

Critical Problem:
    In WSL, Claude Code may set CLAUDE_PROJECT_DIR to Windows path format
    (C:\\Users\\...) instead of WSL format (/mnt/c/Users/...).
    This causes hooks to fail because they execute in WSL bash environment.

Solution:
    This module detects WSL environment and converts Windows paths to WSL format
    automatically, ensuring hooks work correctly regardless of path format.

Environment Variables:
    WSL_DISTRO_NAME: WSL distribution name (set by WSL runtime)
    WSLENV: WSL environment variable list (set by WSL runtime)
    WSL_INTEROP: WSL interop socket path (set by WSL runtime)
"""

from __future__ import annotations

import os
import re
from pathlib import Path


def is_wsl() -> bool:
    """Check if running in WSL (Windows Subsystem for Linux).

    Detects WSL via environment variables set by WSL runtime.
    Works for both WSL 1 and WSL 2.

    Returns:
        True if running in WSL, False otherwise

    Examples:
        >>> # In WSL environment
        >>> is_wsl()
        True

        >>> # In native Linux or macOS
        >>> is_wsl()
        False
    """
    return "WSL_DISTRO_NAME" in os.environ or "WSLENV" in os.environ or "WSL_INTEROP" in os.environ


def is_windows_path(path: str) -> bool:
    """Check if a path is in Windows format.

    Detects Windows paths by checking for drive letters (C:, D:, etc.)
    and backslash separators.

    Args:
        path: Path string to check

    Returns:
        True if path appears to be Windows format, False otherwise

    Examples:
        >>> is_windows_path("C:\\Users\\goos\\project")
        True

        >>> is_windows_path("D:/code/app")
        True

        >>> is_windows_path("/mnt/c/Users/goos/project")
        False

        >>> is_windows_path("/home/user/project")
        False
    """
    if not path:
        return False

    # Check for drive letter pattern (C:, D:, etc.)
    # Matches: C:, C:\, C:/, c:, c:\, c:/
    drive_pattern = re.compile(r"^[a-zA-Z]:[/\\]")
    return bool(drive_pattern.match(path))


def convert_windows_to_wsl(path: str) -> str:
    """Convert Windows path to WSL path format.

    Converts Windows drive paths to WSL /mnt/ format:
    - C:\\Users\\goos\\project → /mnt/c/Users/goos/project
    - D:\\code\\app → /mnt/d/code/app

    Handles both backslash and forward slash separators.
    Preserves case sensitivity for paths.

    Args:
        path: Windows-style path

    Returns:
        WSL-style path

    Raises:
        ValueError: If path is not a valid Windows path

    Examples:
        >>> convert_windows_to_wsl("C:\\Users\\goos\\project")
        '/mnt/c/Users/goos/project'

        >>> convert_windows_to_wsl("D:/code/app")
        '/mnt/d/code/app'

        >>> convert_windows_to_wsl("E:\\Documents\\file.txt")
        '/mnt/e/Documents/file.txt'

    Edge Cases:
        >>> # Relative paths are not converted
        >>> convert_windows_to_wsl("relative/path")
        Traceback (most recent call last):
        ...
        ValueError: Not a valid Windows path: relative/path

        >>> # UNC paths are not supported
        >>> convert_windows_to_wsl("\\\\server\\share\\file")
        Traceback (most recent call last):
        ...
        ValueError: UNC paths are not supported: \\\\server\\share\\file
    """
    if not path:
        raise ValueError("Empty path provided")

    # Check for UNC paths (\\server\share)
    if path.startswith("\\\\") or path.startswith("//"):
        raise ValueError(f"UNC paths are not supported: {path}")

    # Validate Windows path format
    if not is_windows_path(path):
        raise ValueError(f"Not a valid Windows path: {path}")

    # Extract drive letter (e.g., "C:" from "C:\Users\...")
    drive_pattern = re.compile(r"^([a-zA-Z]):[/\\](.*)$")
    match = drive_pattern.match(path)

    if not match:
        raise ValueError(f"Invalid Windows path format: {path}")

    drive_letter = match.group(1).lower()
    rest_of_path = match.group(2)

    # Replace backslashes with forward slashes
    rest_of_path = rest_of_path.replace("\\", "/")

    # Construct WSL path: /mnt/{drive}/{rest}
    wsl_path = f"/mnt/{drive_letter}/{rest_of_path}"

    # Remove any double slashes
    wsl_path = re.sub(r"/{2,}", "/", wsl_path)

    return wsl_path


def convert_wsl_to_windows(path: str) -> str:
    """Convert WSL path to Windows path format.

    Converts WSL /mnt/ paths to Windows drive format:
    - /mnt/c/Users/goos/project → C:\\Users\\goos\\project
    - /mnt/d/code/app → D:\\code\\app

    Args:
        path: WSL-style path

    Returns:
        Windows-style path with backslashes

    Raises:
        ValueError: If path is not a valid WSL /mnt/ path

    Examples:
        >>> convert_wsl_to_windows("/mnt/c/Users/goos/project")
        'C:\\\\Users\\\\goos\\\\project'

        >>> convert_wsl_to_windows("/mnt/d/code/app")
        'D:\\\\code\\\\app'

        >>> convert_wsl_to_windows("/mnt/e/Documents/file.txt")
        'E:\\\\Documents\\\\file.txt'

    Edge Cases:
        >>> # Non-/mnt/ paths are not converted
        >>> convert_wsl_to_windows("/home/user/project")
        Traceback (most recent call last):
        ...
        ValueError: Not a valid WSL /mnt/ path: /home/user/project

        >>> # /mnt/ without drive letter fails
        >>> convert_wsl_to_windows("/mnt/")
        Traceback (most recent call last):
        ...
        ValueError: Invalid WSL /mnt/ path format: /mnt/
    """
    if not path:
        raise ValueError("Empty path provided")

    # Check for /mnt/{drive}/ pattern
    mnt_pattern = re.compile(r"^/mnt/([a-zA-Z])(/.*)?$")
    match = mnt_pattern.match(path)

    if not match:
        raise ValueError(f"Not a valid WSL /mnt/ path: {path}")

    drive_letter = match.group(1).upper()
    rest_of_path = match.group(2) or ""

    # Remove leading slash and replace forward slashes with backslashes
    if rest_of_path.startswith("/"):
        rest_of_path = rest_of_path[1:]

    rest_of_path = rest_of_path.replace("/", "\\")

    # Construct Windows path: {Drive}:\{rest}
    windows_path = f"{drive_letter}:\\{rest_of_path}"

    return windows_path


def normalize_path_for_wsl(path: str) -> str:
    """Normalize path for WSL environment.

    Smart normalization that converts Windows paths to WSL format
    only when running in WSL environment. Otherwise, returns path as-is.

    This is the recommended function for path normalization in MoAI-ADK.

    Args:
        path: Any path string (Windows or Unix format)

    Returns:
        WSL-compatible path if in WSL and path is Windows format,
        otherwise returns original path

    Examples:
        >>> # In WSL environment with Windows path
        >>> os.environ["WSL_DISTRO_NAME"] = "Ubuntu"
        >>> normalize_path_for_wsl("C:\\Users\\goos\\project")
        '/mnt/c/Users/goos/project'

        >>> # In WSL environment with Unix path (no conversion needed)
        >>> normalize_path_for_wsl("/home/user/project")
        '/home/user/project'

        >>> # In native Linux (not WSL)
        >>> del os.environ["WSL_DISTRO_NAME"]
        >>> normalize_path_for_wsl("C:\\Users\\goos\\project")
        'C:\\\\Users\\\\goos\\\\project'

    Use Cases:
        This function is ideal for normalizing CLAUDE_PROJECT_DIR in WSL:

        >>> claude_dir = os.environ.get("CLAUDE_PROJECT_DIR")
        >>> if claude_dir:
        ...     claude_dir = normalize_path_for_wsl(claude_dir)
        ...     root_path = Path(claude_dir).resolve()
    """
    if not path:
        return path

    # Only convert if running in WSL and path is Windows format
    if is_wsl() and is_windows_path(path):
        try:
            return convert_windows_to_wsl(path)
        except ValueError:
            # If conversion fails, return original path
            # This handles edge cases like UNC paths
            return path

    # Return as-is for:
    # - Non-WSL environments
    # - Unix paths in WSL
    # - Invalid paths
    return path


def get_normalized_path(path: str | Path) -> Path:
    """Get normalized Path object for WSL compatibility.

    Convenience function that combines normalization with Path object creation.
    Handles both string and Path inputs.

    Args:
        path: Path string or Path object

    Returns:
        Normalized and resolved Path object

    Examples:
        >>> # In WSL with Windows path
        >>> p = get_normalized_path("C:\\Users\\goos\\project")
        >>> isinstance(p, Path)
        True

        >>> # With Path object
        >>> p = get_normalized_path(Path("/home/user/project"))
        >>> isinstance(p, Path)
        True

    Note:
        This function calls resolve() which makes the path absolute
        and resolves any symlinks.
    """
    if isinstance(path, Path):
        path_str = str(path)
    else:
        path_str = path

    normalized = normalize_path_for_wsl(path_str)
    return Path(normalized).resolve()
