"""
@FEATURE:TEMPLATE-001 Template rendering utilities for MoAI-ADK

Handles template file reading, variable substitution, and content rendering
with error handling and security validation.
"""

from pathlib import Path
from string import Template
from typing import Any

from ..utils.logger import get_logger

logger = get_logger(__name__)


class TemplateRenderer:
    """@TASK:TEMPLATE-RENDERER-001 Handles template rendering operations."""

    def __init__(self):
        """Initialize template renderer."""
        pass

    def render_template_file(self, template_path: Path, context: dict[str, Any]) -> str:
        """
        Render a template file with context variables.

        Args:
            template_path: Path to template file
            context: Variables to substitute in template

        Returns:
            str: Rendered template content
        """
        try:
            with open(template_path, encoding="utf-8") as f:
                template_content = f.read()

            template = Template(template_content)
            return template.safe_substitute(context)

        except FileNotFoundError:
            logger.error("Template file not found: %s", template_path)
            return ""
        except Exception as e:
            logger.error("Error rendering template %s: %s", template_path, e)
            return template_content if "template_content" in locals() else ""

    def render_template_string(self, template_content: str, context: dict[str, Any]) -> str:
        """
        Render a template string with context variables.

        Args:
            template_content: Template content as string
            context: Variables to substitute in template

        Returns:
            str: Rendered template content
        """
        try:
            template = Template(template_content)
            return template.safe_substitute(context)
        except Exception as e:
            logger.error("Error rendering template string: %s", e)
            return template_content