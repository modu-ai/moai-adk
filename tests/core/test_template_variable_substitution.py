# @TEST:TEMPLATE-SUBSTITUTION-001 | SPEC: SPEC-TEMPLATE-VAR-SUBSTITUTION-001 | CODE: src/moai_adk/core/template_engine.py
"""
Tests for template variable substitution ({{MOAI_VERSION}} → actual version)

@TEST:TEMPLATE-SUBSTITUTION-001 - {{MOAI_VERSION}} substitution in config.json template
@TEST:TEMPLATE-SUBSTITUTION-002 - {{MOAI_VERSION}} substitution in CLAUDE.md template
@TEST:TEMPLATE-SUBSTITUTION-003 - Multiple template variables simultaneous substitution
@TEST:TEMPLATE-SUBSTITUTION-004 - Version substitution in file-based templates
@TEST:TEMPLATE-SUBSTITUTION-005 - Version substitution in directory-based rendering
"""

import json
import tempfile
from pathlib import Path

import pytest

from moai_adk.core.template_engine import TemplateEngine


class TestTemplateVariableSubstitution:
    """Test template variable substitution for version field"""

    def test_moai_version_substitution_in_config_template(self) -> None:
        """
        GIVEN: A config.json template with {{MOAI_VERSION}} placeholder
        WHEN: Template is rendered with version variable
        THEN: {{MOAI_VERSION}} should be replaced with actual version string
        """
        # @TEST:TEMPLATE-SUBSTITUTION-001
        template_engine = TemplateEngine()

        # Template content with placeholder
        template_content = """
{
  "moai": {
    "version": "{{MOAI_VERSION}}",
    "update_check_frequency": "daily"
  },
  "project": {
    "name": "{{PROJECT_NAME}}",
    "mode": "team"
  }
}
"""

        # Variables to substitute
        variables = {
            "MOAI_VERSION": "0.22.4",
            "PROJECT_NAME": "TestProject"
        }

        # Render template
        rendered = template_engine.render_string(template_content, variables)

        # Parse JSON to verify
        config = json.loads(rendered)

        # RED: These assertions will fail because substitution is not working correctly
        assert "moai" in config, "Config should have 'moai' section"
        assert "version" in config["moai"], "moai section should have 'version' field"

        version = config["moai"]["version"]
        assert version == "0.22.4", f"Version should be '0.22.4', got '{version}'"
        assert "{{MOAI_VERSION}}" not in version, "Version should not contain placeholder"

    def test_moai_version_substitution_in_claude_template(self) -> None:
        """
        GIVEN: A CLAUDE.md template with {{MOAI_VERSION}} placeholder
        WHEN: Template is rendered with version variable
        THEN: {{MOAI_VERSION}} should be replaced with actual version string
        """
        # @TEST:TEMPLATE-SUBSTITUTION-002
        template_engine = TemplateEngine()

        # Template content with placeholder
        template_content = """
# {{PROJECT_NAME}}

**SPEC-First TDD Development with Alfred SuperAgent**

> **Document Language**: {{CONVERSATION_LANGUAGE_NAME}}
> **Project Owner**: {{PROJECT_OWNER}}
> **Config**: `.moai/config.json`
> **Version**: {{MOAI_VERSION}}
> **Current Conversation Language**: {{CONVERSATION_LANGUAGE_NAME}}

## Project Information

- **Name**: {{PROJECT_NAME}}
- **Description**: {{PROJECT_DESCRIPTION}}
- **Version**: {{MOAI_VERSION}}
- **Mode**: {{PROJECT_MODE}}
"""

        # Variables to substitute
        variables = {
            "MOAI_VERSION": "0.22.4",
            "PROJECT_NAME": "TestProject",
            "PROJECT_DESCRIPTION": "Test Description",
            "PROJECT_MODE": "team",
            "CONVERSATION_LANGUAGE_NAME": "English",
            "PROJECT_OWNER": "@user"
        }

        # Render template
        rendered = template_engine.render_string(template_content, variables)

        # RED: These assertions will fail because substitution is not working correctly
        assert "{{MOAI_VERSION}}" not in rendered, "Template should not contain placeholder"
        assert "0.22.4" in rendered, "Template should contain actual version"
        assert "> Version: 0.22.4" in rendered, "Version should be properly substituted"

    def test_multiple_template_variables_substitution(self) -> None:
        """
        GIVEN: A template with multiple placeholders including {{MOAI_VERSION}}
        WHEN: Template is rendered with all variables
        THEN: All placeholders should be substituted correctly
        """
        # @TEST:TEMPLATE-SUBSTITUTION-003
        template_engine = TemplateEngine()

        # Template content with multiple placeholders
        template_content = """
{
  "moai": {
    "version": "{{MOAI_VERSION}}",
    "update_check_frequency": "daily"
  },
  "project": {
    "name": "{{PROJECT_NAME}}",
    "description": "{{PROJECT_DESCRIPTION}}",
    "mode": "{{PROJECT_MODE}}",
    "version": "{{PROJECT_VERSION}}"
  },
  "language": {
    "conversation_language": "{{CONVERSATION_LANGUAGE}}"
  }
}
"""

        # Variables to substitute
        variables = {
            "MOAI_VERSION": "0.22.4",
            "PROJECT_NAME": "TestProject",
            "PROJECT_DESCRIPTION": "Test Description",
            "PROJECT_MODE": "team",
            "PROJECT_VERSION": "1.0.0",
            "CONVERSATION_LANGUAGE": "en"
        }

        # Render template
        rendered = template_engine.render_string(template_content, variables)

        # Parse JSON to verify
        config = json.loads(rendered)

        # RED: These assertions will fail because substitution is not working correctly
        assert config["moai"]["version"] == "0.22.4", f"moai.version should be '0.22.4', got '{config['moai']['version']}'"
        assert config["project"]["name"] == "TestProject", f"project.name should be 'TestProject', got '{config['project']['name']}'"
        assert config["project"]["mode"] == "team", f"project.mode should be 'team', got '{config['project']['mode']}'"
        assert config["project"]["version"] == "1.0.0", f"project.version should be '1.0.0', got '{config['project']['version']}'"
        assert config["language"]["conversation_language"] == "en", f"language.conversation_language should be 'en', got '{config['language']['conversation_language']}'"

        # Verify no placeholders remain
        rendered_without_placeholders = template_engine.render_string(template_content, variables)
        assert "{{MOAI_VERSION}}" not in rendered_without_placeholders, "Template should not contain MOAI_VERSION placeholder"
        assert "{{PROJECT_NAME}}" not in rendered_without_placeholders, "Template should not contain PROJECT_NAME placeholder"

    def test_version_substitution_in_file_based_templates(self, tmp_path: Path) -> None:
        """
        GIVEN: Template files with {{MOAI_VERSION}} placeholder
        WHEN: Files are rendered using TemplateEngine
        THEN: {{MOAI_VERSION}} should be replaced with actual version
        """
        # @TEST:TEMPLATE-SUBSTITUTION-004
        template_engine = TemplateEngine()

        # Create template files
        templates_dir = tmp_path / "templates"
        templates_dir.mkdir()

        # Config template
        config_template = templates_dir / "config.json.template"
        config_template.write_text("""
{
  "moai": {
    "version": "{{MOAI_VERSION}}",
    "update_check_frequency": "daily"
  },
  "project": {
    "name": "{{PROJECT_NAME}}"
  }
}
""")

        # Render template to output file
        output_dir = tmp_path / "output"
        output_dir.mkdir()

        variables = {
            "MOAI_VERSION": "0.22.4",
            "PROJECT_NAME": "TestProject"
        }

        # Render file
        rendered = template_engine.render_file(
            config_template, variables, output_dir / "config.json"
        )

        # Verify rendered content
        config = json.loads(rendered)

        # RED: These assertions will fail because substitution is not working correctly
        assert config["moai"]["version"] == "0.22.4", f"moai.version should be '0.22.4', got '{config['moai']['version']}'"
        assert config["project"]["name"] == "TestProject", f"project.name should be 'TestProject', got '{config['project']['name']}'"

        # Verify file was written
        output_file = output_dir / "config.json"
        assert output_file.exists(), "Output file should exist"
        assert "{{MOAI_VERSION}}" not in output_file.read_text(), "Output file should not contain placeholder"

    def test_version_substitution_in_directory_based_rendering(self, tmp_path: Path) -> None:
        """
        GIVEN: A directory of template files with {{MOAI_VERSION}} placeholders
        WHEN: Directory is rendered using TemplateEngine
        THEN: {{MOAI_VERSION}} should be replaced with actual version in all files
        """
        # @TEST:TEMPLATE-SUBSTITUTION-005
        template_engine = TemplateEngine()

        # Create template directory
        templates_dir = tmp_path / "templates"
        templates_dir.mkdir()

        # Create multiple template files
        templates_dir / "config.json.template".write_text("""
{
  "moai": {
    "version": "{{MOAI_VERSION}}",
    "update_check_frequency": "daily"
  },
  "project": {
    "name": "{{PROJECT_NAME}}"
  }
}
""")

        templates_dir / "CLAUDE.md.template".write_text("""
# {{PROJECT_NAME}}

> **Version**: {{MOAI_VERSION}}

This is a test project with version {{MOAI_VERSION}}.
""")

        # Render directory
        output_dir = tmp_path / "output"
        output_dir.mkdir()

        variables = {
            "MOAI_VERSION": "0.22.4",
            "PROJECT_NAME": "TestProject"
        }

        # Render all templates
        results = template_engine.render_directory(
            templates_dir, output_dir, variables
        )

        # Verify results
        assert len(results) == 2, f"Should render 2 files, got {len(results)}"

        # Check config.json
        config_content = json.loads(results["config.json"])
        assert config_content["moai"]["version"] == "0.22.4", f"moai.version should be '0.22.4', got '{config_content['moai']['version']}'"
        assert config_content["project"]["name"] == "TestProject", f"project.name should be 'TestProject', got '{config_content['project']['name']}'"

        # Check CLAUDE.md
        claude_content = results["CLAUDE.md.template"]
        assert "{{MOAI_VERSION}}" not in claude_content, "CLAUDE.md should not contain placeholder"
        assert "0.22.4" in claude_content, "CLAUDE.md should contain actual version"
        assert "> Version: 0.22.4" in claude_content, "CLAUDE.md should have substituted version"

        # Verify files were written
        assert (output_dir / "config.json").exists(), "config.json should exist"
        assert (output_dir / "CLAUDE.md").exists(), "CLAUDE.md should exist"

    def test_version_substitution_with_default_variables(self, tmp_path: Path) -> None:
        """
        GIVEN: Template config with moai.version
        WHEN: get_default_variables is called
        THEN: MOAI_VERSION should be extracted from config
        """
        template_engine = TemplateEngine()

        # Test config with version field
        config = {
            "moai": {
                "version": "1.5.0",
                "update_check_frequency": "weekly"
            },
            "project": {
                "name": "TestProject"
            }
        }

        # Get default variables
        variables = template_engine.get_default_variables(config)

        # RED: This assertion will fail because the current implementation has a bug
        # It's using a hardcoded default version instead of extracting from config
        assert "MOAI_VERSION" in variables, "MOAI_VERSION should be in variables"
        assert variables["MOAI_VERSION"] == "1.5.0", \
            f"MOAI_VERSION should be '1.5.0', got '{variables['MOAI_VERSION']}'"

        # Test config without version field (should return default)
        config_without_version = {
            "moai": {
                "update_check_frequency": "daily"
            },
            "project": {
                "name": "TestProject"
            }
        }

        variables = template_engine.get_default_variables(config_without_version)

        # RED: This assertion will fail because the current implementation incorrectly extracts from config
        # When no version is present, it should return the default "0.7.0"
        assert variables["MOAI_VERSION"] == "0.7.0", \
            f"MOAI_VERSION should be default '0.7.0', got '{variables['MOAI_VERSION']}'"

    def test_version_substitution_preserves_existing_values(self) -> None:
        """
        GIVEN: Template with existing values and {{MOAI_VERSION}} placeholder
        WHEN: Template is rendered
        THEN: Existing values should be preserved, only placeholder substituted
        """
        template_engine = TemplateEngine()

        # Template content
        template_content = """
{
  "moai": {
    "version": "{{MOAI_VERSION}}",
    "update_check_frequency": "daily"
  },
  "constitution": {
    "test_coverage_target": 85,
    "enforce_tdd": true
  }
}
"""

        # Variables to substitute
        variables = {
            "MOAI_VERSION": "0.22.4"
        }

        # Render template
        rendered = template_engine.render_string(template_content, variables)

        # Parse JSON to verify
        config = json.loads(rendered)

        # RED: These assertions will fail because the test demonstrates the current behavior
        # The existing values should be preserved
        assert config["moai"]["version"] == "0.22.4", f"moai.version should be substituted"
        assert config["moai"]["update_check_frequency"] == "daily", \
            f"moai.update_check_frequency should be preserved"
        assert config["constitution"]["test_coverage_target"] == 85, \
            f"constitution.test_coverage_target should be preserved"
        assert config["constitution"]["enforce_tdd"] is True, \
            f"constitution.enforce_tdd should be preserved"

    def test_version_substitution_error_handling(self) -> None:
        """
        GIVEN: Template with undefined MOAI_VERSION variable
        WHEN: Template is rendered in strict mode
        THEN: Should raise TemplateRuntimeError or similar
        """
        template_engine = TemplateEngine(strict_undefined=True)

        # Template content with undefined variable
        template_content = """
{
  "moai": {
    "version": "{{MOAI_VERSION}}"
  }
}
"""

        # Variables without MOAI_VERSION
        variables = {
            "PROJECT_NAME": "TestProject"
        }

        # RED: This assertion will fail because the current implementation doesn't handle missing variables correctly
        # Should raise an error when undefined variable is used in strict mode
        with pytest.raises(Exception):  # Should be TemplateRuntimeError or RuntimeError
            template_engine.render_string(template_content, variables)

    def test_version_substitution_non_strict_mode(self) -> None:
        """
        GIVEN: Template with undefined MOAI_VERSION variable in non-strict mode
        WHEN: Template is rendered
        THEN: Undefined variable should be rendered as empty string
        """
        template_engine = TemplateEngine(strict_undefined=False)

        # Template content with undefined variable
        template_content = """
{
  "moai": {
    "version": "{{MOAI_VERSION}}"
  }
}
"""

        # Variables without MOAI_VERSION
        variables = {
            "PROJECT_NAME": "TestProject"
        }

        # Render template
        rendered = template_engine.render_string(template_content, variables)

        # Parse JSON to verify
        config = json.loads(rendered)

        # RED: This assertion will fail because the current implementation doesn't handle strict mode correctly
        # In non-strict mode, undefined variables should be empty strings
        assert config["moai"]["version"] == "", \
            f"moai.version should be empty string, got '{config['moai']['version']}'"