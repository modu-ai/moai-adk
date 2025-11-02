"""
Template engine for parameterizing GitHub templates and other configuration files.

Supports Jinja2-style templating with variable substitution and conditional sections.
Enables users to customize MoAI-ADK templates for their own projects.

@TAG:TEMPLATE-ENGINE-001 - Template variable substitution system
@TAG:GITHUB-CUSTOMIZATION-001 - GitHub template parameterization
"""

from pathlib import Path
from typing import Any, Dict, Optional

from jinja2 import (
    Environment,
    FileSystemLoader,
    StrictUndefined,
    TemplateNotFound,
    TemplateRuntimeError,
    TemplateSyntaxError,
)


class TemplateEngine:
    """
    Jinja2-based template engine for MoAI-ADK configuration and GitHub templates.

    Supports:
    - Variable substitution: {{PROJECT_NAME}}, {{SPEC_DIR}}, etc.
    - Conditional sections: {{#ENABLE_TRUST_5}}...{{/ENABLE_TRUST_5}}
    - File-based and string-based template rendering
    """

    def __init__(self, strict_undefined: bool = True):
        """
        Initialize the template engine.

        Args:
            strict_undefined: If True, raise error on undefined variables (default: True).
                             If False, render undefined variables as empty strings.

        Note:
            Changed to strict_undefined=True (v0.10.2+) for safer template rendering.
            Variables must be explicitly provided to avoid silent template failures.
        """
        self.strict_undefined = strict_undefined
        self.undefined_behavior = StrictUndefined if strict_undefined else None

    def render_string(
        self,
        template_string: str,
        variables: Dict[str, Any]
    ) -> str:
        """
        Render a Jinja2 template string with provided variables.

        Args:
            template_string: The template content as a string
            variables: Dictionary of variables to substitute

        Returns:
            Rendered template string

        Raises:
            TemplateSyntaxError: If template syntax is invalid
            TemplateRuntimeError: If variable substitution fails in strict mode
        """
        try:
            env = Environment(
                undefined=self.undefined_behavior,
                trim_blocks=False,
                lstrip_blocks=False
            )
            template = env.from_string(template_string)
            return template.render(**variables)
        except (TemplateSyntaxError, TemplateRuntimeError) as e:
            raise RuntimeError(f"Template rendering error: {e}")

    def render_file(
        self,
        template_path: Path,
        variables: Dict[str, Any],
        output_path: Optional[Path] = None
    ) -> str:
        """
        Render a Jinja2 template file with provided variables.

        Args:
            template_path: Path to the template file
            variables: Dictionary of variables to substitute
            output_path: If provided, write rendered content to this path

        Returns:
            Rendered template content

        Raises:
            FileNotFoundError: If template file doesn't exist
            TemplateSyntaxError: If template syntax is invalid
            TemplateRuntimeError: If variable substitution fails in strict mode
        """
        if not template_path.exists():
            raise FileNotFoundError(f"Template file not found: {template_path}")

        template_dir = template_path.parent
        template_name = template_path.name

        try:
            env = Environment(
                loader=FileSystemLoader(str(template_dir)),
                undefined=self.undefined_behavior,
                trim_blocks=False,
                lstrip_blocks=False
            )
            template = env.get_template(template_name)
            rendered = template.render(**variables)

            if output_path:
                output_path.parent.mkdir(parents=True, exist_ok=True)
                output_path.write_text(rendered, encoding='utf-8')

            return rendered
        except TemplateNotFound:
            raise FileNotFoundError(f"Template not found in {template_dir}: {template_name}")
        except (TemplateSyntaxError, TemplateRuntimeError) as e:
            raise RuntimeError(f"Template rendering error in {template_path}: {e}")

    def render_directory(
        self,
        template_dir: Path,
        output_dir: Path,
        variables: Dict[str, Any],
        pattern: str = "**/*.{md,yml,yaml,json}"
    ) -> Dict[str, str]:
        """
        Render all template files in a directory.

        Args:
            template_dir: Source directory containing templates
            output_dir: Destination directory for rendered files
            variables: Dictionary of variables to substitute
            pattern: Glob pattern for files to process (default: template files)

        Returns:
            Dictionary mapping input paths to rendered content

        Raises:
            FileNotFoundError: If template directory doesn't exist
        """
        if not template_dir.exists():
            raise FileNotFoundError(f"Template directory not found: {template_dir}")

        results = {}
        output_dir.mkdir(parents=True, exist_ok=True)

        for template_file in template_dir.glob(pattern):
            if template_file.is_file():
                relative_path = template_file.relative_to(template_dir)
                output_file = output_dir / relative_path

                try:
                    rendered = self.render_file(template_file, variables, output_file)
                    results[str(relative_path)] = rendered
                except Exception as e:
                    raise RuntimeError(f"Error rendering {relative_path}: {e}")

        return results

    @staticmethod
    def get_default_variables(config: Dict[str, Any]) -> Dict[str, Any]:
        """
        Extract template variables from project configuration.

        Args:
            config: Project configuration dictionary (from .moai/config.json)

        Returns:
            Dictionary of template variables
        """
        github_config = config.get("github", {}).get("templates", {})
        project_config = config.get("project", {})

        return {
            # Project information
            "PROJECT_NAME": project_config.get("name", "MyProject"),
            "PROJECT_DESCRIPTION": project_config.get("description", ""),
            "PROJECT_MODE": project_config.get("mode", "team"),  # team or personal

            # Directory structure
            "SPEC_DIR": github_config.get("spec_directory", ".moai/specs"),
            "DOCS_DIR": github_config.get("docs_directory", ".moai/docs"),
            "TEST_DIR": github_config.get("test_directory", "tests"),

            # Feature flags
            "ENABLE_TRUST_5": github_config.get("enable_trust_5", True),
            "ENABLE_TAG_SYSTEM": github_config.get("enable_tag_system", True),
            "ENABLE_ALFRED_COMMANDS": github_config.get("enable_alfred_commands", True),

            # Language configuration
            "CONVERSATION_LANGUAGE": project_config.get("conversation_language", "en"),
            "CONVERSATION_LANGUAGE_NAME": project_config.get("conversation_language_name", "English"),

            # Additional metadata
            "MOAI_VERSION": config.get("moai", {}).get("version", "0.7.0"),
        }


class TemplateVariableValidator:
    """
    Validates template variables for completeness and correctness.
    Ensures all required variables are present before rendering.
    """

    REQUIRED_VARIABLES = {
        "PROJECT_NAME": str,
        "SPEC_DIR": str,
        "DOCS_DIR": str,
        "TEST_DIR": str,
    }

    OPTIONAL_VARIABLES = {
        "PROJECT_DESCRIPTION": (str, type(None)),
        "PROJECT_MODE": str,
        "ENABLE_TRUST_5": bool,
        "ENABLE_TAG_SYSTEM": bool,
        "ENABLE_ALFRED_COMMANDS": bool,
        "CONVERSATION_LANGUAGE": str,
        "CONVERSATION_LANGUAGE_NAME": str,
    }

    @classmethod
    def validate(cls, variables: Dict[str, Any]) -> tuple[bool, list[str]]:
        """
        Validate template variables.

        Args:
            variables: Dictionary of variables to validate

        Returns:
            Tuple of (is_valid, list_of_errors)
        """
        errors = []

        # Check required variables
        for var_name, var_type in cls.REQUIRED_VARIABLES.items():
            if var_name not in variables:
                errors.append(f"Missing required variable: {var_name}")
            elif not isinstance(variables[var_name], var_type):
                actual_type = type(variables[var_name]).__name__
                errors.append(
                    f"Invalid type for {var_name}: "
                    f"expected {var_type.__name__}, got {actual_type}"
                )

        # Check optional variables (if present)
        for var_name, var_type in cls.OPTIONAL_VARIABLES.items():
            if var_name in variables:
                if not isinstance(variables[var_name], var_type):
                    if isinstance(var_type, tuple):
                        type_names = " or ".join(t.__name__ for t in var_type)
                    else:
                        type_names = var_type.__name__
                    actual_type = type(variables[var_name]).__name__
                    errors.append(
                        f"Invalid type for {var_name}: "
                        f"expected {type_names}, got {actual_type}"
                    )

        return len(errors) == 0, errors
