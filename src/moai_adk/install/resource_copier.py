"""
@FEATURE:RESOURCE-COPIER-001 Resource Copying for MoAI-ADK

@TASK:RESOURCE-COPIER-001 Handles specialized copying logic for different resource types
@DESIGN:MODULE-SPLIT-001 Extracted from file_operations.py for TRUST principle compliance
"""

import logging
from pathlib import Path

from .template_manager import TemplateManager

logger = logging.getLogger(__name__)


class ResourceCopier:
    """Specialized resource copying manager for different MoAI-ADK resource types."""

    def __init__(self, file_operations_ref=None):
        """
        Initialize resource copier.

        Args:
            file_operations_ref: Reference to FileOperations for basic copy operations
        """
        self.template_manager = TemplateManager()
        self.file_operations = file_operations_ref

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
        try:
            claude_root = project_path / ".claude"
            success = self.file_operations.copy_template(".claude", claude_root, overwrite)

            if success:
                # Import here to avoid circular imports
                from .permission_manager import PermissionManager
                permission_manager = PermissionManager()

                # Customize settings.json with Python command
                permission_manager.customize_settings_json(claude_root, python_command)

                # Ensure hook permissions
                permission_manager.ensure_hook_permissions(claude_root)

            return success

        except Exception as e:
            logger.error(f"Failed to copy Claude resources: {e}")
            return False

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
        try:
            moai_root = project_path / ".moai"
            exclude_subdirs = ["_templates"] if exclude_templates else None

            return self.file_operations.copy_template(".moai", moai_root, overwrite, exclude_subdirs)

        except Exception as e:
            logger.error(f"Failed to copy MoAI resources: {e}")
            return False

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
        try:
            github_root = project_path / ".github"
            return self.file_operations.copy_template(".github", github_root, overwrite)
        except Exception as e:
            logger.error(f"Failed to copy GitHub resources: {e}")
            return False

    def copy_project_memory(self, project_path: Path, overwrite: bool = False) -> bool:
        """
        Copy project memory file (CLAUDE.md).

        Args:
            project_path: Project root path
            overwrite: Overwrite existing files

        Returns:
            bool: Copy success status
        """
        try:
            target_path = project_path / "CLAUDE.md"
            return self.file_operations.copy_template("CLAUDE.md", target_path, overwrite)
        except Exception as e:
            logger.error(f"Failed to copy project memory: {e}")
            return False

    def copy_memory_templates(
        self,
        project_path: Path,
        tech_stack: list[str],
        context: dict[str, str],
        overwrite: bool = False,
    ) -> list[Path]:
        """Copy stack-specific memory templates into the project."""
        copied_files: list[Path] = []
        memory_dir = project_path / ".moai" / "memory"
        memory_dir.mkdir(parents=True, exist_ok=True)

        # Get templates for tech stack
        templates_to_copy = self.template_manager.get_memory_templates_for_stack(tech_stack)

        for template in templates_to_copy:
            try:
                template_name = f"{template}.md"
                source_path = self.template_manager.get_template_path(
                    f".moai/memory/{template_name}"
                )
                target_path = memory_dir / template_name

                # Skip if file exists and overwrite is False
                if target_path.exists() and not overwrite:
                    logger.info(f"Memory template exists, skipping: {template_name}")
                    continue

                # Get template content and render with context
                template_content = self.template_manager.get_template_content(
                    f".moai/memory/{template_name}"
                )

                if template_content:
                    rendered_content = self.template_manager.render_template_with_context(
                        template_content, context
                    )
                    target_path.write_text(rendered_content, encoding="utf-8")
                    copied_files.append(target_path)
                    logger.info(f"Copied memory template: {template_name}")

            except Exception as e:
                logger.error(f"Failed to copy memory template {template}: {e}")
                continue

        return copied_files