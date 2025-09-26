"""
@FEATURE:TEMPLATE-001 Template Management for MoAI-ADK

@TASK:TEMPLATE-001 Manages template discovery, loading, and rendering
@DESIGN:MODULE-SPLIT-001 Extracted from resource_manager.py for TRUST principle compliance
"""

import logging
from importlib import resources
from pathlib import Path
from string import Template as StrTemplate

logger = logging.getLogger(__name__)


MEMORY_STACK_TEMPLATE_MAP: dict[str, list[str]] = {
    "python": ["backend-python"],
    "fastapi": ["backend-python", "backend-fastapi"],
    "django": ["backend-python"],
    "flask": ["backend-python"],
    "java": ["backend-spring"],
    "spring": ["backend-spring"],
    "spring boot": ["backend-spring"],
    "springboot": ["backend-spring"],
    "spring-boot": ["backend-spring"],
    "react": ["frontend-react"],
    "nextjs": ["frontend-react", "frontend-next"],
    "vue": ["frontend-vue"],
    "nuxt": ["frontend-vue"],
    "angular": ["frontend-angular"],
    "typescript": ["frontend-react"],
    "javascript": ["frontend-react"],
}


class TemplateManager:
    """Template discovery, loading, and rendering manager."""

    def __init__(self):
        """Initialize template manager."""
        self.templates_root = None
        try:
            # Try to get templates path from package
            self.templates_root = resources.files("moai_adk.resources.templates")
            logger.debug("Template manager initialized successfully")
        except Exception as e:
            logger.error(f"Failed to initialize template manager: {e}")
            self.templates_root = None

    def get_version(self) -> str:
        """Get MoAI-ADK version."""
        try:
            from .._version import __version__
            return __version__
        except ImportError:
            return "unknown"

    def get_template_path(self, template_name: str) -> Path:
        """Get path to a template."""
        return self.templates_root / template_name

    def get_template_content(self, template_name: str) -> str | None:
        """Get template content as string."""
        try:
            template_path = self.get_template_path(template_name)
            with resources.as_file(template_path) as template_file:
                if template_file.is_file():
                    return template_file.read_text(encoding="utf-8")
                else:
                    logger.warning(f"Template {template_name} is not a file")
                    return None
        except Exception as e:
            logger.error(f"Failed to read template {template_name}: {e}")
            return None

    def list_templates(self) -> list[str]:
        """List available templates."""
        try:
            if not self.templates_root:
                return []

            templates = []
            with resources.as_file(self.templates_root) as templates_dir:
                for item in templates_dir.iterdir():
                    if item.is_dir() or item.name.endswith((".md", ".json", ".yaml", ".yml")):
                        templates.append(item.name)
            return sorted(templates)
        except Exception as e:
            logger.error(f"Failed to list templates: {e}")
            return []

    def render_template_with_context(
        self, template_content: str, context: dict[str, str]
    ) -> str:
        """
        Render template content with context variables.

        Args:
            template_content: Template content with ${variable} placeholders
            context: Context dictionary for variable substitution

        Returns:
            Rendered template content
        """
        try:
            template = StrTemplate(template_content)
            return template.safe_substitute(**context)
        except Exception as e:
            logger.error(f"Failed to render template: {e}")
            return template_content

    def get_memory_templates_for_stack(self, tech_stack: list[str]) -> list[str]:
        """
        Get memory template names for given tech stack.

        Args:
            tech_stack: List of technologies

        Returns:
            List of memory template names to copy
        """
        templates_to_copy: list[str] = ["development-guide"]

        for tech in tech_stack:
            tech_key = tech.lower()
            if tech_key in MEMORY_STACK_TEMPLATE_MAP:
                templates_to_copy.extend(MEMORY_STACK_TEMPLATE_MAP[tech_key])

        # Remove duplicates while preserving order
        seen = set()
        unique_templates = []
        for template in templates_to_copy:
            if template not in seen:
                seen.add(template)
                unique_templates.append(template)

        return unique_templates

    def substitute_template_variables(
        self, content: str, project_context: dict[str, str]
    ) -> str:
        """
        Substitute template variables in content.

        Args:
            content: Template content
            project_context: Project context variables

        Returns:
            Content with substituted variables
        """
        try:
            # Create template and substitute variables
            template = StrTemplate(content)
            return template.safe_substitute(**project_context)
        except Exception as e:
            logger.error(f"Failed to substitute template variables: {e}")
            return content

    def apply_project_context(
        self, template_path: Path, context: dict[str, str]
    ) -> bool:
        """
        Apply project context to template file in place.

        Args:
            template_path: Path to template file
            context: Context variables

        Returns:
            True if successful, False otherwise
        """
        try:
            if not template_path.exists():
                logger.warning(f"Template file not found: {template_path}")
                return False

            # Read original content
            original_content = template_path.read_text(encoding="utf-8")

            # Apply context substitution
            processed_content = self.substitute_template_variables(
                original_content, context
            )

            # Write back if changed
            if processed_content != original_content:
                template_path.write_text(processed_content, encoding="utf-8")
                logger.info(f"Applied project context to: {template_path}")

            return True
        except Exception as e:
            logger.error(f"Failed to apply project context to {template_path}: {e}")
            return False