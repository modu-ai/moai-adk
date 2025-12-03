"""
Tests for template variable substitution ({{MOAI_VERSION}} → actual version)

"""

import json
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
        variables = {"MOAI_VERSION": "0.22.4", "PROJECT_NAME": "TestProject"}

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
            "PROJECT_OWNER": "@user",
        }

        # Render template
        rendered = template_engine.render_string(template_content, variables)

        # Verify substitution worked correctly
        assert "{{MOAI_VERSION}}" not in rendered, "Template should not contain placeholder"
        assert "0.22.4" in rendered, "Template should contain actual version"
        assert "> **Version**: 0.22.4" in rendered, "Version should be properly substituted"

    def test_multiple_template_variables_substitution(self) -> None:
        """
        GIVEN: A template with multiple placeholders including {{MOAI_VERSION}}
        WHEN: Template is rendered with all variables
        THEN: All placeholders should be substituted correctly
        """
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
            "CONVERSATION_LANGUAGE": "en",
        }

        # Render template
        rendered = template_engine.render_string(template_content, variables)

        # Parse JSON to verify
        config = json.loads(rendered)

        # RED: These assertions will fail because substitution is not working correctly
        assert (
            config["moai"]["version"] == "0.22.4"
        ), f"moai.version should be '0.22.4', got '{config['moai']['version']}'"
        assert (
            config["project"]["name"] == "TestProject"
        ), f"project.name should be 'TestProject', got '{config['project']['name']}'"
        assert config["project"]["mode"] == "team", f"project.mode should be 'team', got '{config['project']['mode']}'"
        assert (
            config["project"]["version"] == "1.0.0"
        ), f"project.version should be '1.0.0', got '{config['project']['version']}'"
        assert (
            config["language"]["conversation_language"] == "en"
        ), f"language.conversation_language should be 'en', got '{config['language']['conversation_language']}'"

        # Verify no placeholders remain
        rendered_without_placeholders = template_engine.render_string(template_content, variables)
        assert (
            "{{MOAI_VERSION}}" not in rendered_without_placeholders
        ), "Template should not contain MOAI_VERSION placeholder"
        assert (
            "{{PROJECT_NAME}}" not in rendered_without_placeholders
        ), "Template should not contain PROJECT_NAME placeholder"

    def test_version_substitution_in_file_based_templates(self, tmp_path: Path) -> None:
        """
        GIVEN: Template files with {{MOAI_VERSION}} placeholder
        WHEN: Files are rendered using TemplateEngine
        THEN: {{MOAI_VERSION}} should be replaced with actual version
        """
        template_engine = TemplateEngine()

        # Create template files
        templates_dir = tmp_path / "templates"
        templates_dir.mkdir()

        # Config template
        config_template = templates_dir / "config.json.template"
        config_template.write_text(
            """
{
  "moai": {
    "version": "{{MOAI_VERSION}}",
    "update_check_frequency": "daily"
  },
  "project": {
    "name": "{{PROJECT_NAME}}"
  }
}
"""
        )

        # Render template to output file
        output_dir = tmp_path / "output"
        output_dir.mkdir()

        variables = {"MOAI_VERSION": "0.22.4", "PROJECT_NAME": "TestProject"}

        # Render file
        rendered = template_engine.render_file(config_template, variables, output_dir / "config.json")

        # Verify rendered content
        config = json.loads(rendered)

        # RED: These assertions will fail because substitution is not working correctly
        assert (
            config["moai"]["version"] == "0.22.4"
        ), f"moai.version should be '0.22.4', got '{config['moai']['version']}'"
        assert (
            config["project"]["name"] == "TestProject"
        ), f"project.name should be 'TestProject', got '{config['project']['name']}'"

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
        template_engine = TemplateEngine()

        # Create template directory
        templates_dir = tmp_path / "templates"
        templates_dir.mkdir()

        # Create multiple template files
        (templates_dir / "config.json").write_text(
            """
{
  "moai": {
    "version": "{{MOAI_VERSION}}",
    "update_check_frequency": "daily"
  },
  "project": {
    "name": "{{PROJECT_NAME}}"
  }
}
"""
        )

        (templates_dir / "CLAUDE.md").write_text(
            """
# {{PROJECT_NAME}}

> **Version**: {{MOAI_VERSION}}

This is a test project with version {{MOAI_VERSION}}.
"""
        )

        # Render directory
        output_dir = tmp_path / "output"
        output_dir.mkdir()

        variables = {"MOAI_VERSION": "0.22.4", "PROJECT_NAME": "TestProject"}

        # Render all templates
        results = template_engine.render_directory(templates_dir, output_dir, variables)

        # Verify results
        assert len(results) == 2, f"Should render 2 files, got {len(results)}"

        # Check config.json
        config_content = json.loads(results["config.json"])
        assert (
            config_content["moai"]["version"] == "0.22.4"
        ), f"moai.version should be '0.22.4', got '{config_content['moai']['version']}'"
        assert (
            config_content["project"]["name"] == "TestProject"
        ), f"project.name should be 'TestProject', got '{config_content['project']['name']}'"

        # Check CLAUDE.md
        claude_content = results["CLAUDE.md"]
        assert "{{MOAI_VERSION}}" not in claude_content, "CLAUDE.md should not contain placeholder"
        assert "0.22.4" in claude_content, "CLAUDE.md should contain actual version"
        assert "> **Version**: 0.22.4" in claude_content, "CLAUDE.md should have substituted version"

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
        config = {"moai": {"version": "1.5.0", "update_check_frequency": "weekly"}, "project": {"name": "TestProject"}}

        # Get default variables
        variables = template_engine.get_default_variables(config)

        # RED: This assertion will fail because the current implementation has a bug
        # It's using a hardcoded default version instead of extracting from config
        assert "MOAI_VERSION" in variables, "MOAI_VERSION should be in variables"
        assert (
            variables["MOAI_VERSION"] == "1.5.0"
        ), f"MOAI_VERSION should be '1.5.0', got '{variables['MOAI_VERSION']}'"

        # Test config without version field (should return default)
        config_without_version = {"moai": {"update_check_frequency": "daily"}, "project": {"name": "TestProject"}}

        variables = template_engine.get_default_variables(config_without_version)

        # RED: This assertion will fail because the current implementation incorrectly extracts from config
        # When no version is present, it should return the default "0.7.0"
        assert (
            variables["MOAI_VERSION"] == "0.7.0"
        ), f"MOAI_VERSION should be default '0.7.0', got '{variables['MOAI_VERSION']}'"

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
        variables = {"MOAI_VERSION": "0.22.4"}

        # Render template
        rendered = template_engine.render_string(template_content, variables)

        # Parse JSON to verify
        config = json.loads(rendered)

        # RED: These assertions will fail because the test demonstrates the current behavior
        # The existing values should be preserved
        assert config["moai"]["version"] == "0.22.4", "moai.version should be substituted"
        assert config["moai"]["update_check_frequency"] == "daily", "moai.update_check_frequency should be preserved"
        assert (
            config["constitution"]["test_coverage_target"] == 85
        ), "constitution.test_coverage_target should be preserved"
        assert config["constitution"]["enforce_tdd"] is True, "constitution.enforce_tdd should be preserved"

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
        variables = {"PROJECT_NAME": "TestProject"}

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
        variables = {"PROJECT_NAME": "TestProject"}

        # Render template
        rendered = template_engine.render_string(template_content, variables)

        # Parse JSON to verify
        config = json.loads(rendered)

        # RED: This assertion will fail because the current implementation doesn't handle strict mode correctly
        # In non-strict mode, undefined variables should be empty strings
        assert (
            config["moai"]["version"] == ""
        ), f"moai.version should be empty string, got '{config['moai']['version']}'"

    def test_user_name_variable_with_user_config(self) -> None:
        """
        GIVEN: Config with user.name field
        WHEN: get_default_variables is called
        THEN: USER_NAME should be extracted from config.user.name
        """
        template_engine = TemplateEngine()

        # Test config with user.name
        config = {"project": {"name": "TestProject", "owner": "GoosLab"}, "user": {"name": "철수"}}

        # Get default variables
        variables = template_engine.get_default_variables(config)

        # Verify USER_NAME is extracted from config.user.name
        assert "USER_NAME" in variables, "USER_NAME should be in variables"
        assert variables["USER_NAME"] == "철수", f"USER_NAME should be '철수', got '{variables['USER_NAME']}'"

        # Verify PROJECT_OWNER is separate from USER_NAME
        assert (
            variables["PROJECT_OWNER"] == "GoosLab"
        ), f"PROJECT_OWNER should be 'GoosLab', got '{variables['PROJECT_OWNER']}'"

    def test_user_name_variable_empty_fallback(self) -> None:
        """
        GIVEN: Config without user.name field
        WHEN: get_default_variables is called
        THEN: USER_NAME should return empty string (fallback)
        """
        template_engine = TemplateEngine()

        # Test config without user section
        config = {
            "project": {"name": "TestProject", "owner": "GoosLab"}
            # user section missing
        }

        # Get default variables
        variables = template_engine.get_default_variables(config)

        # Verify USER_NAME returns empty string
        assert "USER_NAME" in variables, "USER_NAME should be in variables"
        assert variables["USER_NAME"] == "", f"USER_NAME should be empty string, got '{variables['USER_NAME']}'"

    def test_user_name_variable_with_empty_string(self) -> None:
        """
        GIVEN: Config with user.name as empty string
        WHEN: get_default_variables is called
        THEN: USER_NAME should return empty string
        """
        template_engine = TemplateEngine()

        # Test config with empty user.name
        config = {"project": {"name": "TestProject", "owner": "GoosLab"}, "user": {"name": ""}}

        # Get default variables
        variables = template_engine.get_default_variables(config)

        # Verify USER_NAME is empty string
        assert variables["USER_NAME"] == "", f"USER_NAME should be empty string, got '{variables['USER_NAME']}'"

    def test_user_name_variable_with_unicode_names(self) -> None:
        """
        GIVEN: Config with various unicode names (Korean, English, Japanese)
        WHEN: get_default_variables is called
        THEN: USER_NAME should support all unicode characters
        """
        template_engine = TemplateEngine()

        # Test cases with different unicode names
        test_cases = [
            ("철수", "철수"),  # Korean
            ("John", "John"),  # English
            ("田中", "田中"),  # Japanese
            ("Москва", "Москва"),  # Russian
            ("李明", "李明"),  # Chinese
        ]

        for input_name, expected_name in test_cases:
            config = {"project": {"name": "TestProject", "owner": "TestOwner"}, "user": {"name": input_name}}

            variables = template_engine.get_default_variables(config)

            assert (
                variables["USER_NAME"] == expected_name
            ), f"USER_NAME should be '{expected_name}', got '{variables['USER_NAME']}'"

    def test_user_name_substitution_in_config_template(self) -> None:
        """
        GIVEN: A config template with {{USER_NAME}} placeholder
        WHEN: Template is rendered with user name variable
        THEN: {{USER_NAME}} should be replaced with actual name
        """
        template_engine = TemplateEngine()

        template_content = """
{
  "user": {
    "name": "{{USER_NAME}}"
  },
  "project": {
    "owner": "{{PROJECT_OWNER}}"
  }
}
"""

        variables = {"USER_NAME": "철수", "PROJECT_OWNER": "GoosLab"}

        # Render template
        rendered = template_engine.render_string(template_content, variables)

        # Parse JSON to verify
        config = json.loads(rendered)

        assert config["user"]["name"] == "철수", f"user.name should be '철수', got '{config['user']['name']}'"
        assert (
            config["project"]["owner"] == "GoosLab"
        ), f"project.owner should be 'GoosLab', got '{config['project']['owner']}'"

    def test_user_name_variable_not_confused_with_project_owner(self) -> None:
        """
        GIVEN: Config with both user.name and project.owner
        WHEN: Variables are extracted
        THEN: USER_NAME and PROJECT_OWNER should be distinct variables
        """
        template_engine = TemplateEngine()

        config = {
            "project": {"name": "TestProject", "owner": "GoosLab"},  # GitHub username
            "user": {"name": "김철수"},  # Personal name
        }

        variables = template_engine.get_default_variables(config)

        # Both variables should exist but have different values
        assert variables["PROJECT_OWNER"] == "GoosLab", "PROJECT_OWNER should be GitHub username"
        assert variables["USER_NAME"] == "김철수", "USER_NAME should be personal name"

        # They should NOT be the same
        assert variables["PROJECT_OWNER"] != variables["USER_NAME"], "PROJECT_OWNER and USER_NAME should be different"

    def test_render_file_with_nonexistent_template(self) -> None:
        """
        GIVEN: A path to a non-existent template file
        WHEN: render_file is called
        THEN: FileNotFoundError should be raised
        """
        template_engine = TemplateEngine()
        nonexistent_path = Path("/nonexistent/path/template.md")
        output_path = Path("/tmp/output.md")

        with pytest.raises(FileNotFoundError) as excinfo:
            template_engine.render_file(nonexistent_path, {}, output_path)

        assert "Template file not found" in str(excinfo.value)

    def test_render_directory_with_nonexistent_directory(self) -> None:
        """
        GIVEN: A path to a non-existent template directory
        WHEN: render_directory is called
        THEN: FileNotFoundError should be raised
        """
        template_engine = TemplateEngine()
        nonexistent_dir = Path("/nonexistent/template/directory")
        output_dir = Path("/tmp/output")

        with pytest.raises(FileNotFoundError) as excinfo:
            template_engine.render_directory(nonexistent_dir, output_dir, {})

        assert "Template directory not found" in str(excinfo.value)

    def test_render_file_with_write_to_output_path(self, tmp_path: Path) -> None:
        """
        GIVEN: A template file and an output path
        WHEN: render_file is called with output_path
        THEN: Rendered content should be written to output file
        """
        template_engine = TemplateEngine()

        # Create template file
        templates_dir = tmp_path / "templates"
        templates_dir.mkdir()
        template_file = templates_dir / "test.txt"
        template_file.write_text("Hello {{NAME}}")

        # Render with output path
        output_file = tmp_path / "output.txt"
        rendered = template_engine.render_file(template_file, {"NAME": "World"}, output_file)

        # Verify rendered content
        assert rendered == "Hello World"
        assert output_file.exists()
        assert output_file.read_text() == "Hello World"

    def test_render_template_with_syntax_error(self) -> None:
        """
        GIVEN: A template with syntax error
        WHEN: render_string is called
        THEN: RuntimeError should be raised
        """
        template_engine = TemplateEngine()

        # Invalid template syntax
        template_content = "{{#if test}} unclosed"

        with pytest.raises(RuntimeError) as excinfo:
            template_engine.render_string(template_content, {})

        assert "Template rendering error" in str(excinfo.value)

    def test_template_variable_validator_validate_required(self) -> None:
        """
        GIVEN: Variables with missing required field
        WHEN: validate is called
        THEN: Should return False and error list with missing variable
        """
        from moai_adk.core.template_engine import TemplateVariableValidator

        # Missing PROJECT_NAME
        variables = {
            "PROJECT_OWNER": "owner",
            "CODEBASE_LANGUAGE": "python",
            "SPEC_DIR": ".moai/specs",
            "DOCS_DIR": ".moai/docs",
            "TEST_DIR": "tests",
            "CONVERSATION_LANGUAGE": "en",
        }

        is_valid, errors = TemplateVariableValidator.validate(variables)

        assert not is_valid
        assert any("PROJECT_NAME" in error for error in errors)

    def test_template_variable_validator_validate_type_error(self) -> None:
        """
        GIVEN: Variables with incorrect type
        WHEN: validate is called
        THEN: Should return False and error list with type error
        """
        from moai_adk.core.template_engine import TemplateVariableValidator

        # PROJECT_NAME should be str, not int
        variables = {
            "PROJECT_NAME": 123,  # Wrong type
            "PROJECT_OWNER": "owner",
            "CODEBASE_LANGUAGE": "python",
            "SPEC_DIR": ".moai/specs",
            "DOCS_DIR": ".moai/docs",
            "TEST_DIR": "tests",
            "CONVERSATION_LANGUAGE": "en",
        }

        is_valid, errors = TemplateVariableValidator.validate(variables)

        assert not is_valid
        assert any("Invalid type for PROJECT_NAME" in error for error in errors)

    def test_template_variable_validator_validate_optional_correct(self) -> None:
        """
        GIVEN: Variables with correct optional fields
        WHEN: validate is called
        THEN: Should return True if all required are present
        """
        from moai_adk.core.template_engine import TemplateVariableValidator

        variables = {
            "PROJECT_NAME": "TestProject",
            "PROJECT_OWNER": "owner",
            "CODEBASE_LANGUAGE": "python",
            "SPEC_DIR": ".moai/specs",
            "DOCS_DIR": ".moai/docs",
            "TEST_DIR": "tests",
            "CONVERSATION_LANGUAGE": "en",
            "USER_NAME": "TestUser",
            "ENABLE_TRUST_5": True,
        }

        is_valid, errors = TemplateVariableValidator.validate(variables)

        assert is_valid
        assert len(errors) == 0

    def test_template_variable_validator_validate_optional_wrong_type(self) -> None:
        """
        GIVEN: Optional variable with wrong type
        WHEN: validate is called
        THEN: Should return False with type error
        """
        from moai_adk.core.template_engine import TemplateVariableValidator

        variables = {
            "PROJECT_NAME": "TestProject",
            "PROJECT_OWNER": "owner",
            "CODEBASE_LANGUAGE": "python",
            "SPEC_DIR": ".moai/specs",
            "DOCS_DIR": ".moai/docs",
            "TEST_DIR": "tests",
            "CONVERSATION_LANGUAGE": "en",
            "ENABLE_TRUST_5": "yes",  # Should be bool, not str
        }

        is_valid, errors = TemplateVariableValidator.validate(variables)

        assert not is_valid
        assert any("ENABLE_TRUST_5" in error for error in errors)

    def test_render_file_with_template_syntax_error_in_file(self, tmp_path: Path) -> None:
        """
        GIVEN: A template file with invalid Jinja2 syntax
        WHEN: render_file is called
        THEN: RuntimeError should be raised with error message
        """
        template_engine = TemplateEngine()

        # Create template file with syntax error
        templates_dir = tmp_path / "templates"
        templates_dir.mkdir()
        template_file = templates_dir / "bad.txt"
        template_file.write_text("{{#if invalid}} unclosed")

        with pytest.raises(RuntimeError) as excinfo:
            template_engine.render_file(template_file, {})

        assert "Template rendering error" in str(excinfo.value)

    def test_render_directory_with_rendering_error(self, tmp_path: Path) -> None:
        """
        GIVEN: A directory with template that causes rendering error
        WHEN: render_directory is called
        THEN: RuntimeError should be raised with error details
        """
        template_engine = TemplateEngine()

        # Create template directory with invalid template
        templates_dir = tmp_path / "templates"
        templates_dir.mkdir()
        (templates_dir / "bad.txt").write_text("{{#if bad}} unclosed")

        output_dir = tmp_path / "output"

        with pytest.raises(RuntimeError) as excinfo:
            template_engine.render_directory(templates_dir, output_dir, {})

        assert "Error rendering" in str(excinfo.value)

    def test_template_variable_validator_optional_tuple_type(self) -> None:
        """
        GIVEN: Optional variable with tuple type specification and wrong value
        WHEN: validate is called
        THEN: Should return False with type error mentioning both valid types
        """
        from moai_adk.core.template_engine import TemplateVariableValidator

        # USER_NAME accepts (str, type(None)) - test with wrong type
        variables = {
            "PROJECT_NAME": "TestProject",
            "PROJECT_OWNER": "owner",
            "CODEBASE_LANGUAGE": "python",
            "SPEC_DIR": ".moai/specs",
            "DOCS_DIR": ".moai/docs",
            "TEST_DIR": "tests",
            "CONVERSATION_LANGUAGE": "en",
            "USER_NAME": 123,  # Should be str or None, not int
        }

        is_valid, errors = TemplateVariableValidator.validate(variables)

        assert not is_valid
        assert any("USER_NAME" in error for error in errors)

    def test_template_variable_validator_all_optional_none_values(self) -> None:
        """
        GIVEN: All optional variables set to None (valid for some)
        WHEN: validate is called
        THEN: Should validate correctly for optional fields that accept None
        """
        from moai_adk.core.template_engine import TemplateVariableValidator

        variables = {
            "PROJECT_NAME": "TestProject",
            "PROJECT_OWNER": "owner",
            "CODEBASE_LANGUAGE": "python",
            "SPEC_DIR": ".moai/specs",
            "DOCS_DIR": ".moai/docs",
            "TEST_DIR": "tests",
            "CONVERSATION_LANGUAGE": "en",
            "USER_NAME": None,  # Valid - USER_NAME accepts (str, None)
        }

        is_valid, errors = TemplateVariableValidator.validate(variables)

        assert is_valid
        assert len(errors) == 0
