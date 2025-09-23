"""
Template engine for dynamic file generation in MoAI-ADK.

This module provides a simple template system for generating project files
from templates with variable substitution.
"""

import json
from datetime import datetime
from pathlib import Path
from string import Template
from typing import Any, Dict, Optional
from importlib import resources

from ..utils.logger import get_logger
from .._version import get_version, __version__, VERSIONS, VERSION_FORMATS

logger = get_logger(__name__)


class TemplateEngine:
    """
    Simple template engine for MoAI-ADK project file generation.

    Handles dynamic creation of project files from templates with
    variable substitution and context management.
    """

    def __init__(self, template_dir: Path):
        """
        Initialize template engine.

        Args:
            template_dir: Path to template directory containing _templates/
        """
        self.template_dir = template_dir
        self.templates_root = template_dir / "_templates"

        # Package-level templates fallback (read-only)
        # moai_adk.resources/templates/.moai/_templates
        try:
            self.pkg_templates_root = (
                resources.files('moai_adk.resources') / 'templates' / '.moai' / '_templates'
            )
        except Exception:
            self.pkg_templates_root = None

        logger.info(f"TemplateEngine initialized with template dir: {template_dir}")

    def create_from_template(
        self,
        template_name: str,
        target_path: Path,
        context: Dict[str, Any],
        create_dirs: bool = True
    ) -> bool:
        """
        Create a file from template with variable substitution.

        Args:
            template_name: Template file name (e.g., 'specs/spec.template.md')
            target_path: Target file path to create
            context: Variables for template substitution
            create_dirs: Whether to create parent directories

        Returns:
            bool: True if successful
        """
        try:
            # Locate template file (project → package fallback)
            template_path: Optional[Path] = None
            template_content: Optional[str] = None

            for extension in ['.template.md', '.template.json', '.template']:
                # 1) Project-local template
                test_path = self.templates_root / f"{template_name}{extension}"
                if test_path.exists():
                    template_path = test_path
                    template_content = test_path.read_text(encoding='utf-8')
                    break

                # 2) Package fallback template
                if self.pkg_templates_root is not None:
                    try:
                        traversable = self.pkg_templates_root / f"{template_name}{extension}"
                        with resources.as_file(traversable) as pkg_path:
                            if pkg_path.exists():
                                template_path = pkg_path
                                template_content = pkg_path.read_text(encoding='utf-8')
                                break
                    except Exception:
                        # Continue trying other extensions
                        pass

            if template_content is None:
                logger.error(
                    "Template not found for: %s (searched %s and package fallback)",
                    template_name,
                    self.templates_root,
                )
                return False
            logger.debug("Template content loaded: %s (%d chars)", template_name, len(template_content))

            # Process template with context
            rendered_content = self._render_template(template_content, context)

            # Create target directory if needed
            if create_dirs:
                target_path.parent.mkdir(parents=True, exist_ok=True)

            # Write rendered content
            target_path.write_text(rendered_content, encoding='utf-8')
            logger.info("Template rendered successfully: %s -> %s", template_name, target_path)
            return True

        except Exception as e:
            logger.error("Failed to create from template %s: %s", template_name, e)
            return False

    def create_spec_from_template(
        self,
        spec_id: str,
        spec_name: str,
        description: str,
        target_path: Path
    ) -> bool:
        """Create SPEC file from template."""
        context = self._create_spec_context(spec_id, spec_name, description)
        return self.create_from_template('specs/spec', target_path, context)

    def create_steering_from_template(
        self,
        steering_type: str,
        project_name: str,
        context: Dict[str, Any],
        target_path: Path
    ) -> bool:
        """Create Steering document from template."""
        # 기본 컨텍스트 보강: 날짜/설명 값이 없을 경우 안전한 기본값 채움
        base_context = {
            'PROJECT_NAME': project_name,
            'STEERING_TYPE': steering_type.upper(),
            **context,
        }
        if 'CREATED_AT' not in base_context:
            base_context['CREATED_AT'] = datetime.now().isoformat()
        if 'CREATED_DATE' not in base_context:
            base_context['CREATED_DATE'] = datetime.now().strftime('%Y-%m-%d')
        if 'PROJECT_DESCRIPTION' not in base_context:
            base_context['PROJECT_DESCRIPTION'] = ''

        enhanced_context = {
            **base_context,
            **self._enhance_context_with_version(base_context),
        }
        return self.create_from_template(f'steering/{steering_type}', target_path, enhanced_context)

    def create_constitution_from_template(
        self,
        project_name: str,
        project_type: str,
        target_path: Path
    ) -> bool:
        """Create 개발 가이드 document from template."""
        context = {
            'PROJECT_NAME': project_name,
            'PROJECT_TYPE': project_type.upper(),
            'CREATED_AT': datetime.now().isoformat(),
            'CREATED_DATE': datetime.now().strftime('%Y-%m-%d'),
        }
        enhanced_context = self._enhance_context_with_version(context)
        return self.create_from_template('memory/constitution', target_path, enhanced_context)

    def should_copy_as_template(self, file_path: Path) -> bool:
        """
        Determine if a file should be processed as a template.

        Args:
            file_path: Path to check

        Returns:
            bool: True if file should be templated
        """
        # Skip non-text files
        if file_path.suffix not in ['.md', '.json', '.yml', '.yaml', '.txt', '.py', '.js', '.ts']:
            return False

        # Template files are processed
        template_indicators = [
            "template" in file_path.name.lower(),
            "sample" in file_path.name.lower(),
            str(file_path).endswith(('.template.md', '.template.json')),
            # Specific problematic files
            "SPEC-001-sample" in str(file_path),
            "ADR-001-sample" in str(file_path),
        ]

        return any(template_indicators)

    def _render_template(self, template_content: str, context: Dict[str, Any]) -> str:
        """
        Render template content with context variables including version info.

        Args:
            template_content: Raw template content
            context: Variables for substitution

        Returns:
            str: Rendered content
        """
        try:
            # Inject version information into context
            enhanced_context = self._enhance_context_with_version(context)

            template = Template(template_content)
            return template.safe_substitute(enhanced_context)
        except Exception as e:
            logger.error("Template rendering failed: %s", e)
            return template_content

    def _enhance_context_with_version(self, context: Dict[str, Any]) -> Dict[str, Any]:
        """
        Enhance context with automatic version information.

        Args:
            context: Original context

        Returns:
            Dict[str, Any]: Enhanced context with version info
        """
        enhanced = context.copy()

        # Add version information
        enhanced.update({
            # Main version
            'MOAI_VERSION': get_version(),
            'VERSION': __version__,

            # Formatted versions
            'VERSION_FULL': VERSION_FORMATS.get('full', f'MoAI-ADK v{__version__}'),
            'VERSION_SHORT': VERSION_FORMATS.get('short', f'v{__version__}'),
            'VERSION_BANNER': VERSION_FORMATS.get('banner', f'🗿 MoAI-ADK v{__version__}'),

            # 개발 가이드 and pipeline versions
            'CONSTITUTION_VERSION': VERSIONS.get('constitution', '1.0'),
            'PIPELINE_VERSION': VERSIONS.get('pipeline', '1.0.0'),

            # Timestamps
            'LAST_UPDATED': datetime.now().strftime('%Y-%m-%d'),
            'CURRENT_YEAR': datetime.now().year,
            'CURRENT_DATE': datetime.now().isoformat(),

            # Version comparison helpers
            'IS_BETA': 'beta' in __version__.lower(),
            'IS_RELEASE': 'beta' not in __version__.lower() and 'alpha' not in __version__.lower(),
        })

        return enhanced

    def _create_spec_context(self, spec_id: str, spec_name: str, description: str) -> Dict[str, Any]:
        """Create context for SPEC template rendering."""
        return {
            'SPEC_ID': spec_id,
            'SPEC_NAME': spec_name,
            'SPEC_DESCRIPTION': description,
            'CREATED_AT': datetime.now().isoformat(),
            'CREATED_DATE': datetime.now().strftime('%Y-%m-%d'),
            'VERSION': '1.0',
            'STATUS': 'DRAFT',
            'AUTHOR': 'MoAI-ADK',
            'PRIORITY': 'HIGH',
            'COMPLEXITY': 'MEDIUM'
        }
