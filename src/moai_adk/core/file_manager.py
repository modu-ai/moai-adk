"""
@FEATURE:FILE-001 Unified file management facade for MoAI-ADK

Provides a unified interface for file operations by coordinating
specialized modules for templates, copying, and script installation.
"""

from pathlib import Path
from typing import Any

from ..utils.logger import get_logger
from .file_copier import FileCopier
from .script_installer import ScriptInstaller
from .security import SecurityManager
from .template_renderer import TemplateRenderer

logger = get_logger(__name__)


class FileManager:
    """@TASK:FILE-MANAGER-001 Unified facade for file operations in MoAI-ADK."""

    def __init__(self, template_dir: Path, security_manager: SecurityManager = None):
        """
        Initialize file manager with specialized modules.

        Args:
            template_dir: Directory containing template files
            security_manager: Security manager instance for validation
        """
        self.template_dir = template_dir
        self.security_manager = security_manager or SecurityManager()

        # Initialize specialized modules
        self.template_renderer = TemplateRenderer()
        self.file_copier = FileCopier(self.security_manager)
        self.script_installer = ScriptInstaller(template_dir, self.security_manager)

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
        return self.file_copier.copy_template_files(
            source_dir, target_dir, pattern, preserve_permissions
        )

    def render_template_file(self, template_path: Path, context: dict[str, Any]) -> str:
        """
        Render a template file with context variables.

        Args:
            template_path: Path to template file
            context: Variables to substitute in template

        Returns:
            str: Rendered template content
        """
        return self.template_renderer.render_template_file(template_path, context)

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
        return self.file_copier.copy_and_render_template(
            source_path, target_path, context, create_dirs
        )

    def copy_hook_scripts(self, target_dir: Path) -> list[Path]:
        """
        Copy MoAI Hook scripts to target directory.

        Args:
            target_dir: Directory to copy hooks to

        Returns:
            List[Path]: List of copied hook files
        """
        return self.script_installer.copy_hook_scripts(target_dir)

    def copy_verification_scripts(self, target_dir: Path) -> list[Path]:
        """
        Copy MoAI verification scripts to target directory.

        Args:
            target_dir: Directory to copy scripts to

        Returns:
            List[Path]: List of copied script files
        """
        return self.script_installer.copy_verification_scripts(target_dir)

    def install_output_styles(
        self, target_dir: Path, context: dict[str, Any]
    ) -> list[Path]:
        """
        Install MoAI-ADK output styles with template rendering.

        Args:
            target_dir: Directory to install styles to
            context: Template context for rendering

        Returns:
            List[Path]: List of installed style files
        """
        return self.script_installer.install_output_styles(target_dir, context)

    def create_gitignore(self, gitignore_path: Path) -> bool:
        """
        Create a comprehensive .gitignore file.

        Args:
            gitignore_path: Path where .gitignore should be created

        Returns:
            bool: True if successful
        """
        return self.script_installer.create_gitignore(gitignore_path)
