"""
Context Management Module for Commands Layer

Provides utilities for:
1. Path validation and absolute path conversion
2. Atomic JSON file operations
3. Phase result persistence and loading
4. Template variable substitution
"""

import json
import os
import re
import tempfile
from datetime import datetime, timezone
from typing import Any, Dict, Optional

# Constants
PROJECT_ROOT_SAFETY_MSG = "Path outside project root: {}"
PARENT_DIR_MISSING_MSG = "Parent directory not found: {}"


def _is_path_within_root(abs_path: str, project_root: str) -> bool:
    """
    Check if absolute path is within project root.

    Resolves symlinks to prevent escape attacks.

    Args:
        abs_path: Absolute path to check
        project_root: Project root directory

    Returns:
        True if path is within root, False otherwise
    """
    try:
        real_abs_path = os.path.realpath(abs_path)
        real_project_root = os.path.realpath(project_root)

        return real_abs_path == real_project_root or real_abs_path.startswith(real_project_root + os.sep)
    except OSError:
        return False


def validate_and_convert_path(relative_path: str, project_root: str) -> str:
    """
    Convert relative path to absolute path and validate it.

    Ensures path stays within project root and parent directories exist
    for file paths.

    Args:
        relative_path: Path to validate and convert (relative or absolute)
        project_root: Project root directory for relative path resolution

    Returns:
        Validated absolute path

    Raises:
        ValueError: If path is outside project root
        FileNotFoundError: If parent directory doesn't exist for file paths
    """
    # Convert to absolute path
    abs_path = os.path.abspath(os.path.join(project_root, relative_path))
    project_root_abs = os.path.abspath(project_root)

    # Security check: ensure path stays within project root
    if not _is_path_within_root(abs_path, project_root_abs):
        raise ValueError(PROJECT_ROOT_SAFETY_MSG.format(abs_path))

    # If it's a directory and exists, return it
    if os.path.isdir(abs_path):
        return abs_path

    # For files, check if parent directory exists
    parent_dir = os.path.dirname(abs_path)
    if not os.path.exists(parent_dir):
        raise FileNotFoundError(PARENT_DIR_MISSING_MSG.format(parent_dir))

    return abs_path


def _cleanup_temp_file(temp_fd: Optional[int], temp_path: Optional[str]) -> None:
    """
    Clean up temporary file handles and paths.

    Silently ignores errors during cleanup.

    Args:
        temp_fd: File descriptor to close, or None
        temp_path: Path to file to remove, or None
    """
    if temp_fd is not None:
        try:
            os.close(temp_fd)
        except OSError:
            pass

    if temp_path and os.path.exists(temp_path):
        try:
            os.unlink(temp_path)
        except OSError:
            pass


def save_phase_result(data: Dict[str, Any], target_path: str) -> None:
    """
    Atomically save phase result to JSON file.

    Uses temporary file and atomic rename to ensure data integrity
    even if write fails midway.

    Args:
        data: Dictionary to save
        target_path: Full path where JSON should be saved

    Raises:
        IOError: If write or rename fails
        OSError: If directory is not writable
    """
    target_dir = os.path.dirname(target_path)
    os.makedirs(target_dir, exist_ok=True)

    # Atomic write using temp file
    temp_fd = None
    temp_path = None

    try:
        # Create temp file in target directory for atomic rename
        temp_fd, temp_path = tempfile.mkstemp(dir=target_dir, prefix=".tmp_phase_", suffix=".json")

        # Write JSON to temp file
        with os.fdopen(temp_fd, "w", encoding="utf-8") as f:
            json.dump(data, f, indent=2, ensure_ascii=False)

        temp_fd = None  # File handle is now closed

        # Atomic rename
        os.replace(temp_path, target_path)

    except Exception as e:
        _cleanup_temp_file(temp_fd, temp_path)
        raise IOError(f"Failed to write {target_path}: {e}")


def load_phase_result(source_path: str) -> Dict[str, Any]:
    """
    Load phase result from JSON file.

    Args:
        source_path: Full path to JSON file to load

    Returns:
        Dictionary containing phase result

    Raises:
        FileNotFoundError: If file doesn't exist
        json.JSONDecodeError: If file is not valid JSON
    """
    if not os.path.exists(source_path):
        raise FileNotFoundError(f"Phase result file not found: {source_path}")

    with open(source_path, "r", encoding="utf-8") as f:
        data = json.load(f)

    return data


def substitute_template_variables(text: str, context: Dict[str, str]) -> str:
    """
    Replace template variables in text with values from context.

    Performs safe string substitution of {{VARIABLE}} placeholders.

    Args:
        text: Text containing {{VARIABLE}} placeholders
        context: Dictionary mapping variable names to values

    Returns:
        Text with variables substituted
    """
    result = text

    for key, value in context.items():
        placeholder = f"{{{{{key}}}}}"
        result = result.replace(placeholder, str(value))

    return result


# Regex pattern for detecting unsubstituted template variables
# Matches {{VARIABLE}}, {{VAR_NAME}}, {{VAR1}}, etc.
TEMPLATE_VAR_PATTERN = r"\{\{[A-Z_][A-Z0-9_]*\}\}"


def validate_no_template_vars(text: str) -> None:
    """
    Validate that text contains no unsubstituted template variables.

    Raises error if any {{VARIABLE}} patterns are found.

    Args:
        text: Text to validate

    Raises:
        ValueError: If unsubstituted variables are found
    """
    matches = re.findall(TEMPLATE_VAR_PATTERN, text)

    if matches:
        raise ValueError(f"Unsubstituted template variables found: {matches}")


class ContextManager:
    """
    Manages context passing and state persistence for Commands layer.

    Handles saving and loading phase results, managing state directory,
    and providing convenient access to command state.
    """

    def __init__(self, project_root: str):
        """
        Initialize ContextManager.

        Args:
            project_root: Root directory of the project
        """
        self.project_root = project_root
        self.state_dir = os.path.join(project_root, ".moai", "memory", "command-state")
        os.makedirs(self.state_dir, exist_ok=True)

    def save_phase_result(self, data: Dict[str, Any]) -> str:
        """
        Save phase result with timestamp.

        Args:
            data: Phase result data

        Returns:
            Path to saved file
        """
        # Generate filename with timestamp
        timestamp = datetime.now(timezone.utc).strftime("%Y%m%d-%H%M%S")
        phase_name = data.get("phase", "unknown")
        filename = f"{phase_name}-{timestamp}.json"
        target_path = os.path.join(self.state_dir, filename)

        save_phase_result(data, target_path)
        return target_path

    def load_latest_phase(self) -> Optional[Dict[str, Any]]:
        """
        Load the most recent phase result.

        Returns:
            Phase result dictionary or None if no phase files exist
        """
        # List all phase files
        phase_files = sorted([f for f in os.listdir(self.state_dir) if f.endswith(".json")])

        if not phase_files:
            return None

        # Load the latest (last in sorted order)
        latest_file = phase_files[-1]
        latest_path = os.path.join(self.state_dir, latest_file)

        return load_phase_result(latest_path)

    def get_state_dir(self) -> str:
        """Get the command state directory path."""
        return self.state_dir
