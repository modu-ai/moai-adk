"""Path utility functions for MoAI-ADK hooks"""

from pathlib import Path


def find_project_root() -> Path:
    """
    Find project root by locating .moai directory.

    Starts from current file location and traverses upward
    until .moai directory is found.

    Returns:
        Path: Project root directory containing .moai/

    Fallback:
        Returns Path.cwd() if .moai not found
    """
    current = Path(__file__).resolve().parent

    # Traverse upward to find .moai directory
    while current != current.parent:
        if (current / ".moai").is_dir():
            return current
        current = current.parent

    # Fallback to current working directory
    return Path.cwd()
