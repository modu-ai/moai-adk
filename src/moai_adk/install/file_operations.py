"""
@FEATURE:FILE-OPS-001 File Operations for MoAI-ADK

@TASK:FILE-OPS-001 Handles file copying, directory operations, and permission management
@DESIGN:MODULE-SPLIT-001 Extracted from resource_manager.py for TRUST principle compliance
"""

import logging
import shutil
from collections.abc import Callable
from importlib import resources
from pathlib import Path

from .template_manager import TemplateManager

logger = logging.getLogger(__name__)


class FileOperations:
    """File operations manager for MoAI-ADK resources."""

    def __init__(self):
        """Initialize file operations manager."""
        self.template_manager = TemplateManager()

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
                        self._ensure_hook_permissions(target_path)
            except Exception as e:
                logger.warning(f"Failed to set hook permissions: {e}")

            return True

        except Exception as e:
            logger.error(f"Failed to copy template {template_name}: {e}")
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
        try:
            claude_root = project_path / ".claude"
            success = self.copy_template(".claude", claude_root, overwrite)

            if success:
                # Customize settings.json with Python command
                self._customize_settings_json(claude_root, python_command)

                # Ensure hook permissions
                self._ensure_hook_permissions(claude_root)

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

            return self.copy_template(".moai", moai_root, overwrite, exclude_subdirs)

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
            return self.copy_template(".github", github_root, overwrite)
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
            return self.copy_template("CLAUDE.md", target_path, overwrite)
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

    def _customize_settings_json(self, claude_root: Path, python_command: str) -> None:
        """
        Customize .claude/settings.json with Python command.

        Args:
            claude_root: .claude directory path
            python_command: Python command to use
        """
        try:
            import json

            settings_path = claude_root / "settings.json"
            if not settings_path.exists():
                logger.warning(f"settings.json not found: {settings_path}")
                return

            # Read current settings
            with open(settings_path) as f:
                settings = json.load(f)

            # Update Python command in hooks if present
            if "hooks" in settings:
                for hook_name, hook_config in settings["hooks"].items():
                    if isinstance(hook_config, dict) and "command" in hook_config:
                        # Replace python/python3 with user's preference
                        command = hook_config["command"]
                        if command.startswith(("python ", "python3 ")):
                            hook_config["command"] = command.replace(
                                command.split()[0], python_command, 1
                            )

            # Write back
            with open(settings_path, "w") as f:
                json.dump(settings, f, indent=2)

            logger.info(f"Customized settings.json with Python command: {python_command}")

        except Exception as e:
            logger.warning(f"Failed to customize settings.json: {e}")

    def _ensure_hook_permissions(self, claude_root: Path) -> None:
        """
        Ensure .claude/hooks/moai/*.py files have execution permissions.

        Args:
            claude_root: .claude directory path
        """
        try:
            hooks_dir = claude_root / "hooks" / "moai"
            if not hooks_dir.exists():
                return

            python_files = list(hooks_dir.glob("*.py"))
            for py_file in python_files:
                # Add execute permission for owner
                current_permissions = py_file.stat().st_mode
                py_file.chmod(current_permissions | 0o100)

            if python_files:
                logger.info(f"Set execution permissions for {len(python_files)} hook files")

        except Exception as e:
            logger.warning(f"Failed to set hook permissions: {e}")