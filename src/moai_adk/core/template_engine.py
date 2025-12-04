"""
Template engine for parameterizing GitHub templates and other configuration files.

Supports Jinja2-style templating with variable substitution and conditional sections.
Enables users to customize MoAI-ADK templates for their own projects.

Performance: Jinja2 Environment instances are cached to avoid recreation overhead
(40-60ms per render → ~1ms after caching).
"""

from functools import lru_cache
from pathlib import Path
from typing import Any, Dict, Optional

from jinja2 import (
    Environment,
    FileSystemLoader,
    StrictUndefined,
    TemplateNotFound,
    TemplateRuntimeError,
    TemplateSyntaxError,
    Undefined,
)


# Module-level cached Environment factory functions
@lru_cache(maxsize=2)
def _get_string_environment(strict: bool) -> Environment:
    """
    Get cached Jinja2 Environment for string rendering.

    Args:
        strict: If True, raise error on undefined variables

    Returns:
        Cached Environment instance

    Performance: Cached to avoid Environment recreation overhead
    """
    return Environment(
        undefined=StrictUndefined if strict else Undefined,
        trim_blocks=False,
        lstrip_blocks=False,
    )


@lru_cache(maxsize=8)
def _get_file_environment(template_dir: str, strict: bool) -> Environment:
    """
    Get cached Jinja2 Environment for file rendering.

    Args:
        template_dir: Template directory path (string for hashability)
        strict: If True, raise error on undefined variables

    Returns:
        Cached Environment instance with FileSystemLoader

    Performance: Cached to avoid Environment recreation overhead
    """
    return Environment(
        loader=FileSystemLoader(template_dir),
        undefined=StrictUndefined if strict else Undefined,
        trim_blocks=False,
        lstrip_blocks=False,
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
        self.undefined_behavior = StrictUndefined if strict_undefined else Undefined

    def render_string(self, template_string: str, variables: Dict[str, Any]) -> str:
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

        Performance: Uses cached Environment instance (40-60ms → ~1ms)
        """
        try:
            # Use cached Environment instead of creating new one
            env = _get_string_environment(strict=self.strict_undefined)
            template = env.from_string(template_string)
            return template.render(**variables)
        except (TemplateSyntaxError, TemplateRuntimeError) as e:
            raise RuntimeError(f"Template rendering error: {e}")

    def render_file(
        self,
        template_path: Path,
        variables: Dict[str, Any],
        output_path: Optional[Path] = None,
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

        Performance: Uses cached Environment instance (40-60ms → ~1ms)
        """
        if not template_path.exists():
            raise FileNotFoundError(f"Template file not found: {template_path}")

        template_dir = template_path.parent
        template_name = template_path.name

        try:
            # Use cached Environment instead of creating new one
            env = _get_file_environment(template_dir=str(template_dir), strict=self.strict_undefined)
            template = env.get_template(template_name)
            rendered = template.render(**variables)

            if output_path:
                output_path.parent.mkdir(parents=True, exist_ok=True)
                output_path.write_text(rendered, encoding="utf-8")

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
        pattern: str = "**/*",
    ) -> Dict[str, str]:
        """
        Render all template files in a directory.

        Args:
            template_dir: Source directory containing templates
            output_dir: Destination directory for rendered files
            variables: Dictionary of variables to substitute
            pattern: Glob pattern for files to process (default: all files)

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
            config: Project configuration dictionary (from .moai/config/config.json)

        Returns:
            Dictionary of template variables
        """
        github_templates_config = config.get("github", {}).get("templates", {})
        github_config = config.get("github", {})
        project_config = config.get("project", {})
        user_config = config.get("user", {})

        return {
            # Project information
            "PROJECT_NAME": project_config.get("name", "MyProject"),
            "PROJECT_DESCRIPTION": project_config.get("description", ""),
            "PROJECT_OWNER": project_config.get("owner", ""),  # GitHub username (project owner)
            "GITHUB_PROFILE_NAME": github_config.get("profile_name", ""),
            "PROJECT_MODE": project_config.get("mode", "team"),  # team or personal
            "CODEBASE_LANGUAGE": project_config.get("codebase_language", "python"),
            # User information
            "USER_NAME": user_config.get("name", ""),
            # Directory structure
            "SPEC_DIR": github_templates_config.get("spec_directory", ".moai/specs"),
            "DOCS_DIR": github_templates_config.get("docs_directory", ".moai/docs"),
            "TEST_DIR": github_templates_config.get("test_directory", "tests"),
            # Feature flags
            "ENABLE_TRUST_5": github_config.get("enable_trust_5", True),
            "ENABLE_ALFRED_COMMANDS": github_templates_config.get("enable_alfred_commands", True),
            # Language configuration
            "CONVERSATION_LANGUAGE": config.get("language", {}).get("conversation_language", "en"),
            "CONVERSATION_LANGUAGE_NAME": config.get("language", {}).get("conversation_language_name", "English"),
            "AGENT_PROMPT_LANGUAGE": config.get("language", {}).get("agent_prompt_language", "english"),
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
        "PROJECT_OWNER": str,
        "CODEBASE_LANGUAGE": str,
        "SPEC_DIR": str,
        "DOCS_DIR": str,
        "TEST_DIR": str,
        "CONVERSATION_LANGUAGE": str,
    }

    OPTIONAL_VARIABLES: Dict[str, Any] = {
        "PROJECT_DESCRIPTION": (str, type(None)),
        "PROJECT_MODE": str,
        "ENABLE_TRUST_5": bool,
        "ENABLE_ALFRED_COMMANDS": bool,
        "CONVERSATION_LANGUAGE": str,
        "CONVERSATION_LANGUAGE_NAME": str,
        "USER_NAME": (str, type(None)),
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
                errors.append(f"Invalid type for {var_name}: expected {var_type.__name__}, got {actual_type}")

        # Check optional variables (if present)
        for var_name, var_type in cls.OPTIONAL_VARIABLES.items():
            if var_name in variables:
                if not isinstance(variables[var_name], var_type):
                    if isinstance(var_type, tuple):
                        type_names = (
                            " or ".join(getattr(t, "__name__", str(t)) for t in var_type)
                            if var_type is not None
                            else "unknown"
                        )  # type: ignore[union-attr]
                    else:
                        type_names = getattr(var_type, "__name__", str(var_type))
                    actual_type = type(variables[var_name]).__name__
                    errors.append(f"Invalid type for {var_name}: expected {type_names}, got {actual_type}")

        return len(errors) == 0, errors
