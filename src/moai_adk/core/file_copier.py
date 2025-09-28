"""
@FEATURE:FILE-COPY-001 File copying utilities for MoAI-ADK

Handles file copying operations with security validation, permission management,
and error handling for template and regular files.
"""

import shutil
from pathlib import Path
from typing import Any

from ..utils.logger import get_logger
from .security import SecurityManager
from .template_renderer import TemplateRenderer

logger = get_logger(__name__)


class FileCopier:
    """@TASK:FILE-COPIER-001 Handles secure file copying operations."""

    def __init__(self, security_manager: SecurityManager = None):
        """
        Initialize file copier.

        Args:
            security_manager: Security manager instance for validation
        """
        self.security_manager = security_manager or SecurityManager()
        self.template_renderer = TemplateRenderer()

    def copy_template_files(
        self,
        source_dir: Path,
        target_dir: Path,
        pattern: str,
        preserve_permissions: bool = False,
    ) -> list[Path]:
        """
        Copy template files matching pattern with security validation.

        Args:
            source_dir: Source directory containing templates
            target_dir: Target directory for copied files
            pattern: Glob pattern to match files
            preserve_permissions: Whether to preserve file permissions

        Returns:
            List[Path]: List of successfully copied files
        """
        if not source_dir.exists():
            logger.warning("Source directory not found: %s", source_dir)
            return []

        created_files = []
        target_dir.mkdir(parents=True, exist_ok=True)

        try:
            for source_file in source_dir.glob(pattern):
                if not source_file.is_file():
                    continue

                # Security validation
                target_file = target_dir / source_file.name
                if not self.security_manager.validate_file_creation(
                    target_file, target_dir
                ):
                    logger.error("Security validation failed for file: %s", target_file)
                    continue

                # Copy file
                shutil.copy2(source_file, target_file)

                # Set permissions if specified
                if preserve_permissions and source_file.stat().st_mode:
                    target_file.chmod(source_file.stat().st_mode)

                created_files.append(target_file)
                logger.debug("Copied template file: %s -> %s", source_file, target_file)

        except Exception as e:
            logger.error("Error copying template files: %s", e)

        return created_files

    def copy_and_render_template(
        self,
        source_path: Path,
        target_path: Path,
        context: dict[str, Any],
        create_dirs: bool = True,
    ) -> bool:
        """
        Copy and render a single template file.

        Args:
            source_path: Source template file
            target_path: Target file path
            context: Template variables
            create_dirs: Whether to create parent directories

        Returns:
            bool: True if successful
        """
        try:
            # Validate source file exists
            if not source_path.exists():
                logger.error("Template source not found: %s", source_path)
                return False

            # Security validation
            if not self.security_manager.validate_file_creation(
                target_path, target_path.parent.parent
            ):
                logger.error("Security validation failed for: %s", target_path)
                return False

            # Create target directory if needed
            if create_dirs:
                target_path.parent.mkdir(parents=True, exist_ok=True)

            # Render template
            rendered_content = self.template_renderer.render_template_file(
                source_path, context
            )

            # Write rendered content
            with open(target_path, "w", encoding="utf-8") as f:
                f.write(rendered_content)

            logger.debug("Rendered template: %s -> %s", source_path, target_path)
            return True

        except Exception as e:
            logger.error("Error copying and rendering template: %s", e)
            return False

    def copy_file_with_permissions(
        self, source_path: Path, target_path: Path, mode: int = None
    ) -> bool:
        """
        Copy a single file with specific permissions.

        Args:
            source_path: Source file path
            target_path: Target file path
            mode: File permissions (e.g., 0o755)

        Returns:
            bool: True if successful
        """
        try:
            if not source_path.exists():
                logger.error("Source file not found: %s", source_path)
                return False

            # Security validation
            if not self.security_manager.validate_file_creation(
                target_path, target_path.parent
            ):
                logger.error("Security validation failed for: %s", target_path)
                return False

            # Create target directory if needed
            target_path.parent.mkdir(parents=True, exist_ok=True)

            # Copy file
            shutil.copy2(source_path, target_path)

            # Set permissions if specified
            if mode is not None:
                target_path.chmod(mode)

            logger.debug("Copied file: %s -> %s", source_path, target_path)
            return True

        except Exception as e:
            logger.error("Error copying file: %s", e)
            return False