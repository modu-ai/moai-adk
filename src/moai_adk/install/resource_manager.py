"""
@FEATURE:RESOURCE-001 MoAI-ADK Resource Manager

@TASK:RESOURCE-001 패키지 내장 리소스를 관리하는 모듈입니다.
@TASK:RESOURCE-002 심볼릭 링크 대신 패키지에서 직접 리소스를 복사하여 관리합니다.
@DESIGN:MODULE-SPLIT-001 Refactored to use separate managers for TRUST principle compliance
"""

import logging
from pathlib import Path

from .file_operations import FileOperations
from .resource_validator import ResourceValidator
from .template_manager import TemplateManager

logger = logging.getLogger(__name__)


class ResourceManager:
    """
    MoAI-ADK Resource Manager

    Orchestrates template management, file operations, and resource validation.
    Refactored for TRUST principle compliance with separate managers.
    """

    def __init__(self):
        """Initialize ResourceManager with component managers."""
        self.template_manager = TemplateManager()
        self.file_operations = FileOperations()
        self.validator = ResourceValidator()
        # Provide backwards compatibility for templates_root access
        self.templates_root = self.template_manager.templates_root
        logger.debug("ResourceManager initialized with component managers")

    def get_version(self) -> str:
        """Get MoAI-ADK version."""
        return self.template_manager.get_version()

    def get_template_path(self, template_name: str) -> Path:
        """Get path to a template."""
        return self.template_manager.get_template_path(template_name)

    def get_template_content(self, template_name: str) -> str | None:
        """Get template content as string."""
        return self.template_manager.get_template_content(template_name)

    def copy_template(
        self,
        template_name: str,
        target_path: Path,
        overwrite: bool = False,
        exclude_subdirs: list[str] | None = None,
    ) -> bool:
        """Copy template to target path."""
        # Validate path safety first
        if not self.validator.validate_safe_path(target_path):
            return False

        return self.file_operations.copy_template(
            template_name, target_path, overwrite, exclude_subdirs
        )

    def copy_claude_resources(
        self,
        project_path: Path,
        overwrite: bool = False,
        python_command: str = "python",
    ) -> bool:
        """Copy .claude directory resources."""
        return self.file_operations.copy_claude_resources(
            project_path, overwrite, python_command
        )

    def copy_moai_resources(
        self,
        project_path: Path,
        overwrite: bool = False,
        exclude_templates: bool = False,
    ) -> bool:
        """Copy .moai directory resources."""
        return self.file_operations.copy_moai_resources(
            project_path, overwrite, exclude_templates
        )

    def copy_github_resources(
        self, project_path: Path, overwrite: bool = False
    ) -> bool:
        """Copy GitHub workflow resources."""
        return self.file_operations.copy_github_resources(project_path, overwrite)

    def copy_project_memory(self, project_path: Path, overwrite: bool = False) -> bool:
        """Copy project memory file (CLAUDE.md)."""
        return self.file_operations.copy_project_memory(project_path, overwrite)

    def copy_memory_templates(
        self,
        project_path: Path,
        tech_stack: list[str],
        context: dict[str, str],
        overwrite: bool = False,
    ) -> list[Path]:
        """Copy stack-specific memory templates into the project."""
        return self.file_operations.copy_memory_templates(
            project_path, tech_stack, context, overwrite
        )

    def list_templates(self) -> list[str]:
        """List available templates."""
        return self.template_manager.list_templates()

    def validate_project_resources(self, project_path: Path) -> bool:
        """Validate project has required resources."""
        return self.validator.validate_project_resources(project_path)

    def validate_required_resources(self, project_path: Path) -> bool:
        """Validate all required MoAI-ADK resources exist."""
        return self.validator.validate_required_resources(project_path)

    def validate_safe_path(self, target_path: Path) -> bool:
        """Validate that target path is safe for operations."""
        return self.validator.validate_safe_path(target_path)

    def check_path_conflicts(self, project_path: Path) -> list[str]:
        """Check for potential path conflicts."""
        return self.validator.check_path_conflicts(project_path)

    def get_project_status(self, project_path: Path) -> dict[str, bool]:
        """Get comprehensive project status."""
        return self.validator.get_project_status(project_path)

    def validate_clean_installation(self, target_path: Path) -> bool:
        """Validate the target path is suitable for clean installation."""
        return self.validator.validate_clean_installation(target_path)

    def apply_project_context(
        self, template_path: Path, context: dict[str, str]
    ) -> bool:
        """Apply project context to template file in place."""
        return self.template_manager.apply_project_context(template_path, context)

    def substitute_template_variables(
        self, content: str, project_context: dict[str, str]
    ) -> str:
        """Substitute template variables in content."""
        return self.template_manager.substitute_template_variables(
            content, project_context
        )

    def render_template_with_context(
        self, template_content: str, context: dict[str, str]
    ) -> str:
        """Render template content with context variables."""
        return self.template_manager.render_template_with_context(
            template_content, context
        )