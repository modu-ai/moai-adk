#!/usr/bin/env python3
"""Platform detection utilities for Alfred Hooks

Provides platform-specific logic to handle cross-platform compatibility issues.

This module enables graceful degradation when platform-specific bugs prevent
normal operation. Currently addresses Windows subprocess hang issue (Issue #107).

Usage:
    from core.platform import is_windows, get_platform

    if is_windows():
        # Use Windows-compatible workaround
        return minimal_response()
    else:
        # Full functionality on Unix systems
        return full_response()

@CODE:WINDOWS-HOOK-COMPAT-001
"""

import sys
from typing import Literal

PlatformType = Literal["windows", "darwin", "linux", "unknown"]


def get_platform() -> PlatformType:
    """Detect current operating system platform.

    Uses sys.platform for reliable OS detection across Python versions.
    This is more reliable than os.name or platform.system() for our use case.

    Returns:
        Platform identifier: "windows", "darwin", "linux", or "unknown"

    Examples:
        >>> get_platform()  # On Windows
        'windows'
        >>> get_platform()  # On macOS
        'darwin'
        >>> get_platform()  # On Linux
        'linux'

    Note:
        - Windows: sys.platform in ["win32", "cygwin", "msys"]
        - macOS: sys.platform == "darwin"
        - Linux: sys.platform starts with "linux"
        - Other platforms return "unknown"

    References:
        https://docs.python.org/3/library/sys.html#sys.platform
    """
    platform = sys.platform.lower()

    if platform.startswith("win") or platform in ["cygwin", "msys"]:
        return "windows"
    elif platform == "darwin":
        return "darwin"
    elif platform.startswith("linux"):
        return "linux"
    else:
        return "unknown"


def is_windows() -> bool:
    """Check if running on Windows platform.

    Convenience function for Windows-specific conditional logic.
    Useful for implementing platform-specific workarounds.

    Returns:
        True if running on Windows, False otherwise

    Examples:
        >>> is_windows()  # On Windows
        True
        >>> is_windows()  # On macOS/Linux
        False

    Use Cases:
        - Implementing Windows-specific workarounds (Issue #107)
        - Adjusting subprocess timeout values
        - Selecting platform-appropriate file paths
        - Enabling/disabling platform-specific features

    Note:
        This is a simple wrapper around get_platform() for readability.
        For multi-platform checks, use get_platform() directly:

        # Good for binary check
        if is_windows():
            ...

        # Better for multi-platform logic
        platform = get_platform()
        if platform == "windows":
            ...
        elif platform == "darwin":
            ...
    """
    return get_platform() == "windows"


__all__ = ["get_platform", "is_windows", "PlatformType"]
