#!/usr/bin/env python3
"""
Test suite for template system functionality

This module tests the complete template system including:
- Variable substitution and validation
- File and directory rendering
- Error handling and edge cases
- Integration with MoAI-ADK configuration system

"""

import json
import tempfile
from pathlib import Path
from typing import Dict, Any

import pytest

from moai_adk.core.template_engine import TemplateEngine, TemplateVariableValidator


class TestTemplateEngine:
    """Test class for template engine functionality"""

    def test_string_rendering_basic(self):
        """Test basic string template rendering"""
        engine = TemplateEngine()
        template = "Hello {{PROJECT_NAME}}!"
        variables = {"PROJECT_NAME": "TestProject"}

        result = engine.render_string(template, variables)
        assert result == "Hello TestProject!"

    def test_string_rendering_with_conditions(self):
        """Test string template rendering with conditional sections"""
        engine = TemplateEngine()
        template = """Hello {{PROJECT_NAME}}
{% if ENABLE_TRUST_5 %}
Using TRUST 5 principles
{% endif %}"""
        variables = {
            "PROJECT_NAME": "TestProject",
            "ENABLE_TRUST_5": True
        }

        result = engine.render_string(template, variables)
        assert "Hello TestProject" in result
        assert "Using TRUST 5 principles" in result

    def test_file_rendering_basic(self):
        """Test file template rendering"""
        # Create temporary template file
        with tempfile.NamedTemporaryFile(mode='w', suffix='.md', delete=False) as f:
            f.write("# {{PROJECT_NAME}}\n\nDescription: {{PROJECT_DESCRIPTION}}")
            template_path = Path(f.name)

        try:
            engine = TemplateEngine()
            variables = {
                "PROJECT_NAME": "TestProject",
                "PROJECT_DESCRIPTION": "A test project"
            }

            result = engine.render_file(template_path, variables)
            assert "# TestProject" in result
            assert "Description: A test project" in result
        finally:
            template_path.unlink()

    def test_directory_rendering(self):
        """Test directory template rendering"""
        # Create temporary directory with templates
        with tempfile.TemporaryDirectory() as temp_dir:
            template_dir = Path(temp_dir) / "templates"
            output_dir = Path(temp_dir) / "output"
            template_dir.mkdir()

            # Create template files
            (template_dir / "file1.md").write_text("Hello {{PROJECT_NAME}}")
            (template_dir / "file2.md").write_text("Project: {{PROJECT_DESCRIPTION}}")

            variables = {
                "PROJECT_NAME": "TestProject",
                "PROJECT_DESCRIPTION": "A test project"
            }

            engine = TemplateEngine()
            results = engine.render_directory(template_dir, output_dir, variables)

            assert len(results) == 2
            assert "file1.md" in results
            assert "file2.md" in results
            assert "Hello TestProject" in results["file1.md"]
            assert "Project: A test project" in results["file2.md"]

    def test_error_handling_invalid_template(self):
        """Test error handling for invalid template syntax"""
        engine = TemplateEngine()
        template = "Hello {{PROJECT_NAME}"  # Missing closing brace

        with pytest.raises(RuntimeError):
            engine.render_string(template, {"PROJECT_NAME": "Test"})

    def test_error_handling_missing_file(self):
        """Test error handling for missing template file"""
        engine = TemplateEngine()
        nonexistent_path = Path("/nonexistent/template.md")

        with pytest.raises(FileNotFoundError):
            engine.render_file(nonexistent_path, {})

    def test_strict_undefined_behavior(self):
        """Test strict undefined variable behavior"""
        engine = TemplateEngine(strict_undefined=True)
        template = "Hello {{PROJECT_NAME}}, {{MISSING_VAR}}"

        with pytest.raises(RuntimeError):
            engine.render_string(template, {"PROJECT_NAME": "Test"})

    def test_lenient_undefined_behavior(self):
        """Test lenient undefined variable behavior"""
        engine = TemplateEngine(strict_undefined=False)
        template = "Hello {{PROJECT_NAME}}, {{MISSING_VAR}}"

        result = engine.render_string(template, {"PROJECT_NAME": "Test"})
        assert result == "Hello Test, "


class TestTemplateVariableValidator:
    """Test class for template variable validator"""

    def test_valid_variables(self):
        """Test validation of valid variables"""
        variables = {
            "PROJECT_NAME": "TestProject",
            "SPEC_DIR": ".moai/specs",
            "DOCS_DIR": ".moai/docs",
            "TEST_DIR": "tests",
            "PROJECT_DESCRIPTION": "A test project",
            "PROJECT_MODE": "team"
        }

        is_valid, errors = TemplateVariableValidator.validate(variables)
        assert is_valid
        assert len(errors) == 0

    def test_missing_required_variables(self):
        """Test validation with missing required variables"""
        variables = {
            "PROJECT_NAME": "TestProject"
            # Missing SPEC_DIR, DOCS_DIR, TEST_DIR
        }

        is_valid, errors = TemplateVariableValidator.validate(variables)
        assert not is_valid
        assert len(errors) == 3
        assert any("SPEC_DIR" in error for error in errors)
        assert any("DOCS_DIR" in error for error in errors)
        assert any("TEST_DIR" in error for error in errors)

    def test_invalid_variable_types(self):
        """Test validation with invalid variable types"""
        variables = {
            "PROJECT_NAME": "TestProject",
            "SPEC_DIR": 123,  # Should be string
            "DOCS_DIR": ".moai/docs",
            "TEST_DIR": "tests",
            "PROJECT_MODE": "team"
        }

        is_valid, errors = TemplateVariableValidator.validate(variables)
        assert not is_valid
        assert any("expected str, got int" in error for error in errors)

    def test_optional_variables(self):
        """Test validation of optional variables"""
        variables = {
            "PROJECT_NAME": "TestProject",
            "SPEC_DIR": ".moai/specs",
            "DOCS_DIR": ".moai/docs",
            "TEST_DIR": "tests"
            # No optional variables
        }

        is_valid, errors = TemplateVariableValidator.validate(variables)
        assert is_valid
        assert len(errors) == 0

    def test_get_default_variables(self):
        """Test getting default variables from config"""
        config = {
            "github": {
                "templates": {
                    "spec_directory": ".moai/specs",
                    "docs_directory": ".moai/docs",
                    "test_directory": "tests",
                    "enable_trust_5": True,
                    "enable_tag_system": True,
                    "enable_alfred_commands": True
                }
            },
            "project": {
                "name": "TestProject",
                "description": "A test project",
                "mode": "team",
                "conversation_language": "ko",
                "conversation_language_name": "Korean"
            },
            "moai": {
                "version": "0.22.4"
            }
        }

        variables = TemplateEngine.get_default_variables(config)

        assert variables["PROJECT_NAME"] == "TestProject"
        assert variables["SPEC_DIR"] == ".moai/specs"
        assert variables["DOCS_DIR"] == ".moai/docs"
        assert variables["TEST_DIR"] == "tests"
        assert variables["ENABLE_TRUST_5"] is True
        assert variables["CONVERSATION_LANGUAGE"] == "ko"
        assert variables["MOAI_VERSION"] == "0.22.4"


class TestTemplateSystemIntegration:
    """Integration tests for complete template system"""

    def test_full_template_workflow(self):
        """Test complete template workflow"""
        # Create test configuration
        config = {
            "github": {
                "templates": {
                    "spec_directory": ".moai/specs",
                    "docs_directory": ".moai/docs",
                    "test_directory": "tests",
                    "enable_trust_5": True
                }
            },
            "project": {
                "name": "IntegrationTestProject",
                "description": "Integration test project",
                "mode": "team",
                "conversation_language": "ko"
            },
            "moai": {
                "version": "0.22.4"
            }
        }

        # Get default variables
        variables = TemplateEngine.get_default_variables(config)

        # Validate variables
        is_valid, errors = TemplateVariableValidator.validate(variables)
        assert is_valid, f"Validation errors: {errors}"

        # Create template engine
        engine = TemplateEngine()

        # Test string rendering
        template = """# {{PROJECT_NAME}}
Description: {{PROJECT_DESCRIPTION}}
Mode: {{PROJECT_MODE}}
Trust 5: {{ENABLE_TRUST_5}}
Language: {{CONVERSATION_LANGUAGE}}"""

        result = engine.render_string(template, variables)
        assert "IntegrationTestProject" in result
        assert "Integration test project" in result
        assert "team" in result
        assert "True" in result
        assert "ko" in result

    def test_template_error_recovery(self):
        """Test error recovery in template processing"""
        engine = TemplateEngine(strict_undefined=False)  # Lenient mode

        # Test with missing variables
        template = """Hello {{PROJECT_NAME}},
Project Mode: {{PROJECT_MODE}},
Missing Variable: {{MISSING_VAR}}"""

        # Should not raise error and render missing vars as empty
        result = engine.render_string(template, {"PROJECT_NAME": "Test"})
        assert "Hello Test," in result
        assert "Project Mode: " in result
        assert "Missing Variable: " in result

    def test_performance_with_large_template(self):
        """Test template engine performance with large template"""
        engine = TemplateEngine()

        # Create a large template
        large_template = ""
        for i in range(1000):
            large_template += f"Line {{i}}: Hello {{PROJECT_NAME}}\n"

        variables = {"PROJECT_NAME": "PerformanceTest"}

        # Should render without excessive delay
        result = engine.render_string(large_template, variables)
        assert "Line 0: Hello PerformanceTest" in result
        assert "Line 999: Hello PerformanceTest" in result