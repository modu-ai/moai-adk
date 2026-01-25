r"""Path utility functions for MoAI-ADK hooks

Provides safe project root detection and .moai directory management.
Prevents .moai directory creation outside of project root.

Environment Variables:
- MOAI_PROJECT_ROOT: Override project root detection (absolute path)
- CLAUDE_PROJECT_DIR: Claude Code project directory (fallback)

WSL Support:
- Automatically normalizes Windows paths to WSL format when running in WSL
- Handles CLAUDE_PROJECT_DIR with Windows path format (C:\...) in WSL environment
"""

from __future__ import annotations

import os
from pathlib import Path
from typing import Optional

try:
    from moai_adk.utils.path_converter import is_wsl, normalize_path_for_wsl
except ImportError:
    # Fallback for when moai_adk is not installed
    def is_wsl() -> bool:
        """Check if running in WSL (Windows Subsystem for Linux)."""
        return "WSL_DISTRO_NAME" in os.environ or "WSLENV" in os.environ or "WSL_INTEROP" in os.environ

    def normalize_path_for_wsl(path: str) -> str:
        """Basic WSL path normalization fallback."""
        import re

        if not path or not is_wsl():
            return path

        # Check for Windows drive letter
        drive_pattern = re.compile(r"^([a-zA-Z]):[/\\](.*)$")
        match = drive_pattern.match(path)

        if match:
            drive = match.group(1).lower()
            rest = match.group(2).replace("\\", "/")
            return f"/mnt/{drive}/{rest}"

        return path


# Project root markers (files/dirs that indicate project root)
PROJECT_ROOT_MARKERS = [
    ".moai/config/config.yaml",  # Primary: MoAI config
    ".moai/config/sections",  # Section-based config
    "CLAUDE.md",  # Claude Code project marker
    ".claude",  # Claude directory
    ".git",  # Git repository root
    "pyproject.toml",  # Python project
    "package.json",  # Node.js project
    "Cargo.toml",  # Rust project
    "go.mod",  # Go project
]

# Cache for project root to avoid repeated filesystem traversals
_project_root_cache: Path | None = None


def get_project_root_from_env() -> Optional[Path]:
    """Get project root from environment variables.

    Checks in order:
    1. MOAI_PROJECT_ROOT - explicit override
    2. CLAUDE_PROJECT_DIR - Claude Code's project directory (WSL-normalized)

    WSL Support:
        In WSL environment, CLAUDE_PROJECT_DIR may contain Windows path format
        (e.g., C:/Users/...). This function automatically converts to WSL format
        (/mnt/c/Users/...) for compatibility with bash hooks.

    Returns:
        Path if valid directory found, None otherwise
    """
    # Check MOAI_PROJECT_ROOT first (explicit override)
    moai_root = os.environ.get("MOAI_PROJECT_ROOT")
    if moai_root:
        root_path = Path(moai_root).resolve()
        if root_path.is_dir():
            return root_path

    # Check CLAUDE_PROJECT_DIR (Claude Code sets this)
    claude_dir = os.environ.get("CLAUDE_PROJECT_DIR")
    if claude_dir:
        # WSL: Normalize Windows paths to WSL format
        # This handles cases where Claude Code sets CLAUDE_PROJECT_DIR
        # to Windows path (C:\Users\...) in WSL environment
        if is_wsl():
            claude_dir = normalize_path_for_wsl(claude_dir)

        root_path = Path(claude_dir).resolve()
        if root_path.is_dir():
            return root_path

    return None


def find_project_root(start_path: Path | None = None) -> Path:
    """Find project root by locating project markers.

    Search order:
    1. Environment variables (MOAI_PROJECT_ROOT, CLAUDE_PROJECT_DIR)
    2. Traverse upward from start_path looking for project markers
    3. Fallback to current working directory

    Args:
        start_path: Starting directory for search (defaults to current file location)

    Returns:
        Path: Project root directory

    Note:
        Results are cached for performance. Use clear_project_root_cache()
        to force re-detection.
    """
    global _project_root_cache

    # Return cached result if available
    if _project_root_cache is not None:
        return _project_root_cache

    # Check environment variables first
    env_root = get_project_root_from_env()
    if env_root:
        _project_root_cache = env_root
        return env_root

    # Determine starting point
    if start_path is None:
        start_path = Path(__file__).resolve().parent
    else:
        start_path = Path(start_path).resolve()

    # Traverse upward to find project markers
    current = start_path
    while current != current.parent:
        for marker in PROJECT_ROOT_MARKERS:
            marker_path = current / marker
            if marker_path.exists():
                _project_root_cache = current
                return current
        current = current.parent

    # Fallback to current working directory
    cwd = Path.cwd()
    _project_root_cache = cwd
    return cwd


def clear_project_root_cache() -> None:
    """Clear the cached project root.

    Use when project root may have changed (e.g., after directory change).
    """
    global _project_root_cache
    _project_root_cache = None


def get_moai_dir() -> Path:
    """Get the .moai directory path in project root.

    Returns:
        Path: Absolute path to .moai directory

    Note:
        Does NOT create the directory. Use ensure_moai_dir() for that.
    """
    return find_project_root() / ".moai"


def ensure_moai_dir(subpath: str = "") -> Path:
    """Ensure .moai subdirectory exists in project root.

    Creates the directory structure if it doesn't exist.
    ONLY creates within project root to prevent pollution.

    Args:
        subpath: Subdirectory path within .moai (e.g., "cache", "logs/sessions")

    Returns:
        Path: Absolute path to the directory

    Raises:
        ValueError: If attempting to create outside project root
    """
    project_root = find_project_root()
    moai_dir = project_root / ".moai"

    if subpath:
        target_dir = moai_dir / subpath
    else:
        target_dir = moai_dir

    # Safety check: ensure target is within project root
    try:
        target_dir.resolve().relative_to(project_root.resolve())
    except ValueError:
        raise ValueError(f"Cannot create .moai directory outside project root: {target_dir}")

    # Create directory if it doesn't exist
    target_dir.mkdir(parents=True, exist_ok=True)
    return target_dir


def is_within_project_root(path: Path) -> bool:
    """Check if a path is within the project root.

    Args:
        path: Path to check

    Returns:
        True if path is within project root, False otherwise
    """
    project_root = find_project_root()
    try:
        Path(path).resolve().relative_to(project_root.resolve())
        return True
    except ValueError:
        return False


def get_safe_moai_path(relative_path: str) -> Path:
    """Get absolute path within .moai directory safely.

    This is the recommended way to get paths within .moai.
    Always returns absolute path based on project root.

    Args:
        relative_path: Path relative to .moai (e.g., "cache/version.json")

    Returns:
        Path: Absolute path within project root's .moai directory

    Example:
        >>> get_safe_moai_path("logs/sessions")
        PosixPath('/path/to/project/.moai/logs/sessions')

        >>> get_safe_moai_path("memory/last-session-state.json")
        PosixPath('/path/to/project/.moai/memory/last-session-state.json')
    """
    return find_project_root() / ".moai" / relative_path


def validate_cwd_is_project_root() -> bool:
    """Validate that current working directory is the project root.

    Returns:
        True if CWD is project root, False otherwise

    Use Case:
        Hooks should check this before creating .moai directories
        to prevent creating them in wrong locations.
    """
    cwd = Path.cwd().resolve()
    project_root = find_project_root().resolve()
    return cwd == project_root
