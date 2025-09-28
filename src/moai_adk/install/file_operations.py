"""
@REFACTOR:FILE-OPS-001 File Operations Orchestrator for MoAI-ADK

@TASK:FILE-OPS-001 Orchestrates file copying, directory operations, and permission management
@DESIGN:MODULE-SPLIT-001 Refactored for TRUST principle compliance - delegates to specialized modules
"""

import logging
import shutil
from collections.abc import Callable
from importlib import resources
from pathlib import Path

from .permission_manager import PermissionManager
from .resource_copier import ResourceCopier
from .template_manager import TemplateManager
from .template_processor import TemplateProcessor

logger = logging.getLogger(__name__)


class FileOperations:
    """File operations manager for MoAI-ADK resources."""

    def __init__(self):
        """Initialize file operations manager with specialized components."""
        self.template_manager = TemplateManager()
        self.template_processor = TemplateProcessor()
        self.permission_manager = PermissionManager()
        self.resource_copier = ResourceCopier(file_operations_ref=self)

    def copy_template(
        self,
        template_name: str,
        target_path: Path,
        overwrite: bool = False,
        exclude_subdirs: list[str] | None = None,
    ) -> bool:
        """
        Copy template to target path.

        Args:
            template_name: Template name (.claude, .moai, etc.)
            target_path: Target copy path
            overwrite: Whether to overwrite existing files

        Returns:
            bool: Copy success status
        """
        try:
            # Resolve to absolute path
            target_path = target_path.resolve()
            template_path = self.template_manager.get_template_path(template_name)

            # Handle existing target path
            if target_path.exists():
                if target_path.is_file():
                    if not overwrite:
                        logger.info(
                            f"Target file already exists, skipping: {target_path}"
                        )
                        return True
                    else:
                        logger.info(f"Overwriting existing file: {target_path}")
                        target_path.unlink()
                else:
                    # Directory case - allow merging
                    logger.info(
                        f"Target directory exists, will merge contents: {target_path}"
                    )

            # Create parent directories
            target_path.parent.mkdir(parents=True, exist_ok=True)

            # Copy from package
            with resources.as_file(template_path) as source_path:
                if source_path.is_dir():
                    # Directory copy - allow merging with existing directory
                    def copy_function(src, dst, **kwargs):
                        # Apply per-file overwrite policy
                        if Path(dst).exists() and not overwrite:
                            logger.debug(f"File exists, skipping: {dst}")
                            return dst
                        return shutil.copy2(src, dst, **kwargs)

                    ignore: Callable | None = None
                    if exclude_subdirs:
                        ignore = shutil.ignore_patterns(*exclude_subdirs)

                    shutil.copytree(
                        source_path,
                        target_path,
                        dirs_exist_ok=True,
                        copy_function=copy_function if not overwrite else shutil.copy2,
                        ignore=ignore,
                    )
                else:
                    shutil.copy2(source_path, target_path)

            logger.info(f"Successfully copied {template_name} to {target_path}")

            # Post-processing: ensure .claude/hooks/moai/*.py execution permissions
            try:
                if target_path.is_dir():
                    # Only post-process if template_name is '.claude' or target is .claude root
                    if (
                        template_name == ".claude"
                        or target_path.name == ".claude"
                    ):
                        self.permission_manager.ensure_hook_permissions(target_path)
            except Exception as e:
                logger.warning(f"Failed to set hook permissions: {e}")

            return True

        except Exception as e:
            logger.error(f"Failed to copy template {template_name}: {e}")
            return False

    def copy_template_with_substitution(
        self,
        template_name: str,
        target_path: Path,
        context: dict[str, str] | None = None,
        overwrite: bool = False,
        exclude_subdirs: list[str] | None = None,
    ) -> bool:
        """
        Copy template to target path with variable substitution for text files.

        Args:
            template_name: Template name (.claude, .moai, etc.)
            target_path: Target copy path
            context: Context variables for template substitution
            overwrite: Whether to overwrite existing files
            exclude_subdirs: Subdirectories to exclude from copy

        Returns:
            bool: Copy success status
        """
        try:
            # Resolve to absolute path
            target_path = target_path.resolve()
            template_path = self.template_manager.get_template_path(template_name)

            if context is None:
                context = {}

            # Handle existing target path
            if target_path.exists():
                if target_path.is_file():
                    if not overwrite:
                        logger.info(
                            f"Target file already exists, skipping: {target_path}"
                        )
                        return False
                    else:
                        target_path.unlink()

            # Ensure parent directory exists
            if not target_path.parent.exists():
                target_path.parent.mkdir(parents=True, exist_ok=True)

            with resources.as_file(template_path) as source_path:
                if source_path.is_dir():
                    # Directory copy with per-file template substitution
                    self.template_processor.copy_directory_with_substitution(
                        source_path, target_path, context, overwrite, exclude_subdirs
                    )
                else:
                    # Single file copy with template substitution
                    self.template_processor.copy_file_with_substitution(source_path, target_path, context)

            logger.info(f"Successfully copied {template_name} to {target_path} with substitution")

            # Post-processing: ensure .claude/hooks/moai/*.py execution permissions
            try:
                if target_path.is_dir():
                    if template_name == ".claude" or target_path.name == ".claude":
                        self.permission_manager.ensure_hook_permissions(target_path)
            except Exception as e:
                logger.warning(f"Failed to set hook permissions: {e}")

            return True

        except Exception as e:
            logger.error(f"Failed to copy template {template_name} with substitution: {e}")
            return False

    def copy_claude_resources(
        self,
        project_path: Path,
        overwrite: bool = False,
        python_command: str = "python",
    ) -> bool:
        """
        Copy .claude directory resources.

        Args:
            project_path: Project root path
            overwrite: Overwrite existing files
            python_command: Python command to use in hooks

        Returns:
            True if successful, False otherwise
        """
        return self.resource_copier.copy_claude_resources(project_path, overwrite, python_command)

    def copy_moai_resources(
        self,
        project_path: Path,
        overwrite: bool = False,
        exclude_templates: bool = False,
    ) -> bool:
        """
        Copy .moai directory resources.

        Args:
            project_path: Project root path
            overwrite: Overwrite existing files
            exclude_templates: Exclude _templates directory

        Returns:
            True if successful, False otherwise
        """
        return self.resource_copier.copy_moai_resources(project_path, overwrite, exclude_templates)

    def copy_github_resources(
        self, project_path: Path, overwrite: bool = False
    ) -> bool:
        """
        Copy GitHub workflow resources.

        Args:
            project_path: Project root path
            overwrite: Overwrite existing files

        Returns:
            True if successful, False otherwise
        """
        return self.resource_copier.copy_github_resources(project_path, overwrite)

    def copy_project_memory(self, project_path: Path, overwrite: bool = False) -> bool:
        """
        Copy project memory file (CLAUDE.md).

        Args:
            project_path: Project root path
            overwrite: Overwrite existing files

        Returns:
            bool: Copy success status
        """
        return self.resource_copier.copy_project_memory(project_path, overwrite)

    def copy_memory_templates(
        self,
        project_path: Path,
        tech_stack: list[str],
        context: dict[str, str],
        overwrite: bool = False,
    ) -> list[Path]:
        """Copy stack-specific memory templates into the project."""
        return self.resource_copier.copy_memory_templates(project_path, tech_stack, context, overwrite)

    # Deprecated methods - kept for backward compatibility
    # These delegate to the appropriate specialized modules

    def _copy_directory_with_substitution(
        self,
        source_dir: Path,
        target_dir: Path,
        context: dict[str, str],
        overwrite: bool,
        exclude_subdirs: list[str] | None = None,
    ) -> None:
        """Legacy method - delegates to TemplateProcessor."""
        self.template_processor.copy_directory_with_substitution(
            source_dir, target_dir, context, overwrite, exclude_subdirs
        )

    def _copy_file_with_substitution(
        self, source_file: Path, target_file: Path, context: dict[str, str]
    ) -> None:
        """Legacy method - delegates to TemplateProcessor."""
        self.template_processor.copy_file_with_substitution(source_file, target_file, context)

    def _should_process_as_template(self, file_path: Path) -> bool:
        """Legacy method - delegates to TemplateProcessor."""
        return self.template_processor.should_process_as_template(file_path)

    def _customize_settings_json(self, claude_root: Path, python_command: str) -> None:
        """Legacy method - delegates to PermissionManager."""
        self.permission_manager.customize_settings_json(claude_root, python_command)

    def _ensure_hook_permissions(self, claude_root: Path) -> None:
        """Legacy method - delegates to PermissionManager."""
        self.permission_manager.ensure_hook_permissions(claude_root)