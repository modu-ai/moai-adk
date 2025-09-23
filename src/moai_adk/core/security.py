"""
Security utilities for MoAI-ADK.

Provides basic path validation and safe operations for a local development tool.
"""

import os
from pathlib import Path
from typing import Set

from ..utils.logger import get_logger

logger = get_logger(__name__)


class SecurityError(Exception):
    """Security-related exception."""
    pass


class SecurityManager:
    """Manages basic security operations for local development environment."""

    def __init__(self):
        self.critical_paths = self._get_critical_paths()

    def _get_critical_paths(self) -> Set[Path]:
        """Get system critical paths that should never be deleted."""
        critical_paths = {Path.home(), Path("/")}

        # Add Windows system paths if on Windows
        if os.name == 'nt':
            critical_paths.add(Path("C:\\"))
            critical_paths.add(Path("C:\\Windows"))
            critical_paths.add(Path("C:\\Program Files"))
            critical_paths.add(Path("C:\\Program Files (x86)"))

        return critical_paths

    def validate_path_safety(self, path: Path, base_path: Path) -> bool:
        """
        Basic path validation to prevent path traversal.

        Args:
            path: Path to validate
            base_path: Base path that should contain the target

        Returns:
            bool: True if path is safe to use
        """
        try:
            # Resolve paths to handle symlinks and relative components
            resolved_path = path.resolve()
            resolved_base = base_path.resolve()

            # Check if path is within base directory
            try:
                resolved_path.relative_to(resolved_base)
                return True
            except ValueError:
                logger.warning("Path outside base directory: %s not in %s", resolved_path, resolved_base)
                return False

        except Exception as e:
            logger.error("Path validation failed: %s", e)
            return False

    def safe_rmtree(self, path: Path) -> bool:
        """
        Safely remove a directory tree after validation.

        Args:
            path: Directory to remove

        Returns:
            bool: True if removal was successful
        """
        try:
            resolved_path = path.resolve()

            # Never delete critical system paths
            if resolved_path in self.critical_paths:
                logger.error("Attempted to delete critical system path: %s", resolved_path)
                raise SecurityError(f"Cannot delete critical system path: {resolved_path}")

            # Additional safety checks
            if resolved_path == Path.home():
                raise SecurityError("Cannot delete user home directory")

            if resolved_path.is_mount():
                raise SecurityError(f"Cannot delete mount point: {resolved_path}")

            # Check if path exists
            if not resolved_path.exists():
                logger.debug("Path does not exist, nothing to remove: %s", resolved_path)
                return True

            if not resolved_path.is_dir():
                logger.error("Path is not a directory: %s", resolved_path)
                return False

            # Safe to remove
            import shutil
            shutil.rmtree(resolved_path)
            logger.info("Successfully removed directory: %s", resolved_path)
            return True

        except SecurityError:
            raise  # Re-raise security errors
        except Exception as e:
            logger.error("Failed to remove directory %s: %s", path, e)
            return False

    def validate_file_creation(self, file_path: Path, base_path: Path) -> bool:
        """
        Validate that a file can be safely created.

        Args:
            file_path: Path where file will be created
            base_path: Base project directory

        Returns:
            bool: True if file creation is safe
        """
        try:
            # Basic path safety check
            if not self.validate_path_safety(file_path, base_path):
                return False

            # Check parent directory
            parent_dir = file_path.parent
            if not parent_dir.exists():
                # Parent directory will be created, validate it too
                if not self.validate_path_safety(parent_dir, base_path):
                    return False

            return True

        except Exception as e:
            logger.error("File creation validation failed: %s", e)
            return False

    def validate_subprocess_path(self, path: Path, base_path: Path) -> bool:
        """
        Validate that a subprocess can be safely executed within the given paths.

        Args:
            path: Path where subprocess will operate
            base_path: Base project directory

        Returns:
            bool: True if subprocess execution is safe
        """
        try:
            # Use the existing path safety validation
            return self.validate_path_safety(path, base_path)
        except Exception as e:
            logger.error("Subprocess path validation failed: %s", e)
            return False

    def sanitize_filename(self, filename: str) -> str:
        """
        Sanitize filename to prevent filesystem issues.

        Args:
            filename: Original filename

        Returns:
            str: Sanitized filename
        """
        if not filename:
            return "unnamed_file"

        # Remove or replace dangerous characters
        # Keep alphanumeric, dots, hyphens, underscores
        import re
        sanitized = re.sub(r'[^a-zA-Z0-9._-]', '_', filename)

        # Remove leading/trailing dots and spaces
        sanitized = sanitized.strip('. ')

        # Ensure not empty after sanitization
        if not sanitized:
            sanitized = 'unnamed_file'

        return sanitized

    def safe_subprocess_run(self, command, *args, **kwargs):
        """
        Safe subprocess execution with basic validation.

        Args:
            command: Command to execute
            *args: Additional arguments
            **kwargs: Additional keyword arguments

        Returns:
            subprocess result
        """
        import subprocess

        # Basic safety checks
        if isinstance(command, str):
            # Check for dangerous patterns
            dangerous_patterns = ['rm -rf /', 'dd if=/dev/zero', ':(){:|:&};:']
            for pattern in dangerous_patterns:
                if pattern in command.lower():
                    raise SecurityError(f"Dangerous command pattern detected: {pattern}")

        return subprocess.run(command, *args, **kwargs)

    def sanitize_command_args(self, args):
        """
        Sanitize command arguments for safe execution.

        Args:
            args: Command arguments to sanitize

        Returns:
            Sanitized arguments
        """
        if isinstance(args, str):
            return args.strip()
        elif isinstance(args, list):
            return [str(arg).strip() for arg in args]
        else:
            return args

    def validate_path_safety_enhanced(self, path: Path, base_path: Path, allow_creation: bool = False) -> bool:
        """
        Enhanced path validation with additional checks.

        Args:
            path: Path to validate
            base_path: Base path that should contain the target
            allow_creation: Whether to allow path creation

        Returns:
            bool: True if path is safe to use
        """
        # Use existing validation logic
        if not self.validate_path_safety(path, base_path):
            return False

        # Additional checks for creation
        if allow_creation and not path.exists():
            return self.validate_file_creation(path, base_path)

        return True

    def validate_file_size(self, file_path: Path, max_size_mb: int = 100) -> bool:
        """
        Validate file size to prevent resource exhaustion.

        Args:
            file_path: Path to file
            max_size_mb: Maximum allowed size in megabytes

        Returns:
            bool: True if file size is acceptable
        """
        try:
            if not file_path.exists():
                return True  # Non-existent files are safe

            file_size = file_path.stat().st_size
            max_size_bytes = max_size_mb * 1024 * 1024

            if file_size > max_size_bytes:
                logger.warning("File size exceeds limit: %s (%d bytes > %d bytes)",
                             file_path, file_size, max_size_bytes)
                return False

            return True

        except Exception as e:
            logger.error("File size validation failed: %s", e)
            return False