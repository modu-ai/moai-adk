"""
@TEST:TEMPLATE-001 Template variable substitution tests

Tests for template variable substitution functionality in MoAI-ADK.
"""

import pytest
from pathlib import Path
from moai_adk.install.template_manager import TemplateManager


class TestTemplateSubstitution:
    """Test template variable substitution functionality."""

    def setup_method(self):
        """Set up test fixtures."""
        self.template_manager = TemplateManager()
        self.test_context = {
            "PROJECT_NAME": "TestProject",
            "PROJECT_DESCRIPTION": "A test project",
            "CREATION_TIMESTAMP": "2025-01-15T10:30:00Z",
            "TECH_STACK": "Python, FastAPI, PostgreSQL",
        }

    def test_unified_substitute_square_brackets(self):
        """Test substitution of [VAR] format variables."""
        template = "# [PROJECT_NAME] - MoAI Agentic Development Kit"
        expected = "# TestProject - MoAI Agentic Development Kit"

        result = self.template_manager.unified_substitute_template_variables(
            template, self.test_context
        )

        assert result == expected

    def test_unified_substitute_double_braces(self):
        """Test substitution of {{VAR}} format variables."""
        template = '{"name": "{{PROJECT_NAME}}", "description": "{{PROJECT_DESCRIPTION}}"}'
        expected = '{"name": "TestProject", "description": "A test project"}'

        result = self.template_manager.unified_substitute_template_variables(
            template, self.test_context
        )

        assert result == expected

    def test_unified_substitute_dollar_braces(self):
        """Test substitution of ${VAR} format variables."""
        template = "# ${PROJECT_NAME} Technology Stack\n\n## Stack: ${TECH_STACK}"
        expected = "# TestProject Technology Stack\n\n## Stack: Python, FastAPI, PostgreSQL"

        result = self.template_manager.unified_substitute_template_variables(
            template, self.test_context
        )

        assert result == expected

    def test_unified_substitute_simple_dollar(self):
        """Test substitution of $VAR format variables."""
        template = "Project: $PROJECT_NAME\nDescription: $PROJECT_DESCRIPTION"
        expected = "Project: TestProject\nDescription: A test project"

        result = self.template_manager.unified_substitute_template_variables(
            template, self.test_context
        )

        assert result == expected

    def test_unified_substitute_mixed_formats(self):
        """Test substitution of mixed variable formats in same template."""
        template = """# [PROJECT_NAME] - ${PROJECT_DESCRIPTION}

Project created at: {{CREATION_TIMESTAMP}}
Tech stack: $TECH_STACK
"""
        expected = """# TestProject - A test project

Project created at: 2025-01-15T10:30:00Z
Tech stack: Python, FastAPI, PostgreSQL
"""

        result = self.template_manager.unified_substitute_template_variables(
            template, self.test_context
        )

        assert result == expected

    def test_unified_substitute_missing_variables(self):
        """Test handling of missing variables (should remain unchanged)."""
        template = "# ${PROJECT_NAME} - ${MISSING_VAR} - [ANOTHER_MISSING]"
        expected = "# TestProject - ${MISSING_VAR} - [ANOTHER_MISSING]"

        result = self.template_manager.unified_substitute_template_variables(
            template, self.test_context
        )

        assert result == expected

    def test_unified_substitute_empty_context(self):
        """Test substitution with empty context."""
        template = "# [PROJECT_NAME] - ${TECH_STACK}"
        expected = "# [PROJECT_NAME] - ${TECH_STACK}"  # Should remain unchanged

        result = self.template_manager.unified_substitute_template_variables(
            template, {}
        )

        assert result == expected

    def test_unified_substitute_complex_json_template(self):
        """Test substitution in complex JSON template."""
        template = """{
  "project": {
    "name": "{{PROJECT_NAME}}",
    "description": "{{PROJECT_DESCRIPTION}}",
    "created_at": "{{CREATION_TIMESTAMP}}",
    "initialized": true,
    "mode": "personal",
    "version": "0.1.0"
  },
  "tech_stack": "${TECH_STACK}"
}"""
        expected = """{
  "project": {
    "name": "TestProject",
    "description": "A test project",
    "created_at": "2025-01-15T10:30:00Z",
    "initialized": true,
    "mode": "personal",
    "version": "0.1.0"
  },
  "tech_stack": "Python, FastAPI, PostgreSQL"
}"""

        result = self.template_manager.unified_substitute_template_variables(
            template, self.test_context
        )

        assert result == expected

    def test_unified_substitute_special_characters(self):
        """Test substitution with special characters in values."""
        context = {
            "PROJECT_NAME": "Test-Project_2025!",
            "PROJECT_DESCRIPTION": 'A "test" project with \'quotes\'',
            "SPECIAL_CHARS": "Line 1\nLine 2\tTabbed",
        }

        template = "Name: ${PROJECT_NAME}\nDesc: [PROJECT_DESCRIPTION]\nSpecial: {{SPECIAL_CHARS}}"
        expected = "Name: Test-Project_2025!\nDesc: A \"test\" project with 'quotes'\nSpecial: Line 1\nLine 2\tTabbed"

        result = self.template_manager.unified_substitute_template_variables(
            template, context
        )

        assert result == expected

    def test_unified_substitute_case_sensitivity(self):
        """Test that variable substitution is case-sensitive."""
        template = "Upper: ${PROJECT_NAME}, Lower: ${project_name}"
        expected = "Upper: TestProject, Lower: ${project_name}"  # Only uppercase should be substituted

        result = self.template_manager.unified_substitute_template_variables(
            template, self.test_context  # Contains PROJECT_NAME but not project_name
        )

        assert result == expected

    def test_unified_substitute_word_boundaries(self):
        """Test that substitution respects word boundaries."""
        context = {"VAR": "value", "PROJECT_NAME": "TestProject"}
        template = "$VAR_NAME should not substitute, but $VAR should. ${PROJECT_NAME}_suffix is ok."
        expected = "$VAR_NAME should not substitute, but value should. TestProject_suffix is ok."

        result = self.template_manager.unified_substitute_template_variables(
            template, context
        )

        assert result == expected

    def test_apply_project_context_file(self, tmp_path):
        """Test applying project context to actual file."""
        # Create test template file
        test_file = tmp_path / "test_template.md"
        test_file.write_text("# ${PROJECT_NAME} - [PROJECT_DESCRIPTION]\n\nTech: {{TECH_STACK}}")

        # Apply context
        result = self.template_manager.apply_project_context(test_file, self.test_context)

        assert result == True

        # Check file contents
        updated_content = test_file.read_text()
        expected_content = "# TestProject - A test project\n\nTech: Python, FastAPI, PostgreSQL"
        assert updated_content == expected_content

    def test_apply_project_context_nonexistent_file(self, tmp_path):
        """Test applying project context to nonexistent file."""
        nonexistent_file = tmp_path / "nonexistent.md"

        result = self.template_manager.apply_project_context(
            nonexistent_file, self.test_context
        )

        assert result == False

    def test_backward_compatibility_old_method(self):
        """Test that old substitute_template_variables method still works."""
        template = "${PROJECT_NAME} - ${PROJECT_DESCRIPTION}"
        expected = "TestProject - A test project"

        result = self.template_manager.substitute_template_variables(
            template, self.test_context
        )

        assert result == expected


if __name__ == "__main__":
    pytest.main([__file__])