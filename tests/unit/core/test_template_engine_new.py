"""
Comprehensive tests for TemplateEngine and TemplateVariableValidator.

Tests template rendering, variable substitution, and validation.
"""

import tempfile
from pathlib import Path
from unittest.mock import MagicMock, Mock, patch

import pytest

from moai_adk.core.template_engine import (
    TemplateEngine,
    TemplateVariableValidator,
)


class TestTemplateEngineInit:
    """Test TemplateEngine initialization."""

    def test_init_strict_undefined_true(self):
        """Test initialization with strict_undefined=True."""
        engine = TemplateEngine(strict_undefined=True)

        assert engine.strict_undefined is True

    def test_init_strict_undefined_false(self):
        """Test initialization with strict_undefined=False."""
        engine = TemplateEngine(strict_undefined=False)

        assert engine.strict_undefined is False

    def test_init_default_strict_undefined(self):
        """Test default strict_undefined value."""
        engine = TemplateEngine()

        assert engine.strict_undefined is True


class TestRenderString:
    """Test render_string method."""

    def test_render_string_simple_variable(self):
        """Test rendering simple variable substitution."""
        engine = TemplateEngine()
        template = "Hello {{name}}!"
        variables = {"name": "World"}

        result = engine.render_string(template, variables)
        assert result == "Hello World!"

    def test_render_string_multiple_variables(self):
        """Test rendering multiple variables."""
        engine = TemplateEngine()
        template = "{{greeting}} {{name}}, you are {{age}} years old"
        variables = {
            "greeting": "Hello",
            "name": "Alice",
            "age": 30,
        }

        result = engine.render_string(template, variables)
        assert "Hello Alice" in result
        assert "30" in result

    def test_render_string_conditional(self):
        """Test rendering conditional sections."""
        engine = TemplateEngine()
        template = (
            "{% if enabled %}Feature is enabled{% else %}Feature is disabled{% endif %}"
        )
        variables = {"enabled": True}

        result = engine.render_string(template, variables)
        assert "Feature is enabled" in result

    def test_render_string_loop(self):
        """Test rendering loops."""
        engine = TemplateEngine()
        template = "{% for item in items %}- {{item}}\n{% endfor %}"
        variables = {"items": ["item1", "item2", "item3"]}

        result = engine.render_string(template, variables)
        assert "- item1" in result
        assert "- item2" in result
        assert "- item3" in result

    def test_render_string_undefined_strict_mode(self):
        """Test undefined variables raise error in strict mode."""
        engine = TemplateEngine(strict_undefined=True)
        template = "Hello {{name}}!"
        variables = {}

        with pytest.raises(RuntimeError):
            engine.render_string(template, variables)

    def test_render_string_undefined_lenient_mode(self):
        """Test undefined variables return empty in lenient mode."""
        engine = TemplateEngine(strict_undefined=False)
        template = "Hello {{name}}!"
        variables = {}

        result = engine.render_string(template, variables)
        assert "Hello !" in result

    def test_render_string_syntax_error(self):
        """Test syntax error handling."""
        engine = TemplateEngine()
        template = "{% if not_closed %}"
        variables = {}

        with pytest.raises(RuntimeError):
            engine.render_string(template, variables)

    def test_render_string_complex_expression(self):
        """Test rendering complex Jinja2 expressions."""
        engine = TemplateEngine()
        template = (
            "{% if count > 0 %}There are {{count}} items{% else %}No items{% endif %}"
        )
        variables = {"count": 5}

        result = engine.render_string(template, variables)
        assert "There are 5 items" in result

    def test_render_string_filter(self):
        """Test rendering with Jinja2 filters."""
        engine = TemplateEngine()
        template = "Project: {{project_name|upper}}"
        variables = {"project_name": "MyProject"}

        result = engine.render_string(template, variables)
        assert "MYPROJECT" in result


class TestRenderFile:
    """Test render_file method."""

    def test_render_file_basic(self):
        """Test rendering a template file."""
        with tempfile.TemporaryDirectory() as tmpdir:
            template_path = Path(tmpdir) / "template.txt"
            template_path.write_text("Hello {{name}}!")

            engine = TemplateEngine()
            result = engine.render_file(template_path, {"name": "World"})

            assert result == "Hello World!"

    def test_render_file_to_output(self):
        """Test rendering file and writing to output."""
        with tempfile.TemporaryDirectory() as tmpdir:
            template_path = Path(tmpdir) / "template.txt"
            output_path = Path(tmpdir) / "output.txt"
            template_path.write_text("Hello {{name}}!")

            engine = TemplateEngine()
            result = engine.render_file(
                template_path,
                {"name": "World"},
                output_path=output_path,
            )

            assert output_path.exists()
            assert output_path.read_text() == "Hello World!"

    def test_render_file_not_found(self):
        """Test rendering non-existent file."""
        with tempfile.TemporaryDirectory() as tmpdir:
            template_path = Path(tmpdir) / "nonexistent.txt"

            engine = TemplateEngine()

            with pytest.raises(FileNotFoundError):
                engine.render_file(template_path, {})

    def test_render_file_with_variables(self):
        """Test rendering file with multiple variables."""
        with tempfile.TemporaryDirectory() as tmpdir:
            template_path = Path(tmpdir) / "template.txt"
            template_path.write_text(
                """Project: {{PROJECT_NAME}}
Owner: {{OWNER}}
Language: {{LANGUAGE}}"""
            )

            engine = TemplateEngine()
            result = engine.render_file(
                template_path,
                {
                    "PROJECT_NAME": "MyApp",
                    "OWNER": "Alice",
                    "LANGUAGE": "Python",
                },
            )

            assert "MyApp" in result
            assert "Alice" in result
            assert "Python" in result

    def test_render_file_creates_output_directories(self):
        """Test that output directories are created."""
        with tempfile.TemporaryDirectory() as tmpdir:
            template_path = Path(tmpdir) / "template.txt"
            output_path = Path(tmpdir) / "subdir" / "deep" / "output.txt"
            template_path.write_text("Content")

            engine = TemplateEngine()
            engine.render_file(
                template_path,
                {},
                output_path=output_path,
            )

            assert output_path.exists()
            assert output_path.parent.exists()


class TestRenderDirectory:
    """Test render_directory method."""

    def test_render_directory_basic(self):
        """Test rendering directory of templates."""
        with tempfile.TemporaryDirectory() as tmpdir:
            template_dir = Path(tmpdir) / "templates"
            output_dir = Path(tmpdir) / "output"
            template_dir.mkdir()

            (template_dir / "file1.txt").write_text("File 1: {{value}}")
            (template_dir / "file2.txt").write_text("File 2: {{value}}")

            engine = TemplateEngine()
            results = engine.render_directory(
                template_dir,
                output_dir,
                {"value": "test"},
            )

            assert len(results) >= 2
            assert output_dir.exists()

    def test_render_directory_with_pattern(self):
        """Test rendering with glob pattern."""
        with tempfile.TemporaryDirectory() as tmpdir:
            template_dir = Path(tmpdir) / "templates"
            output_dir = Path(tmpdir) / "output"
            template_dir.mkdir()

            (template_dir / "template1.txt").write_text("T1: {{x}}")
            (template_dir / "template2.md").write_text("T2: {{x}}")
            (template_dir / "other.doc").write_text("Other")

            engine = TemplateEngine()
            results = engine.render_directory(
                template_dir,
                output_dir,
                {"x": "value"},
                pattern="**/*.txt",
            )

            assert "template1.txt" in results

    def test_render_directory_not_found(self):
        """Test rendering non-existent directory."""
        with tempfile.TemporaryDirectory() as tmpdir:
            template_dir = Path(tmpdir) / "nonexistent"
            output_dir = Path(tmpdir) / "output"

            engine = TemplateEngine()

            with pytest.raises(FileNotFoundError):
                engine.render_directory(template_dir, output_dir, {})

    def test_render_directory_preserves_structure(self):
        """Test that directory structure is preserved."""
        with tempfile.TemporaryDirectory() as tmpdir:
            template_dir = Path(tmpdir) / "templates"
            output_dir = Path(tmpdir) / "output"
            template_dir.mkdir()

            subdir = template_dir / "subdir"
            subdir.mkdir()
            (subdir / "nested.txt").write_text("Nested: {{val}}")

            engine = TemplateEngine()
            results = engine.render_directory(
                template_dir,
                output_dir,
                {"val": "data"},
            )

            nested_output = output_dir / "subdir" / "nested.txt"
            assert nested_output.exists()


class TestGetDefaultVariables:
    """Test get_default_variables method."""

    def test_get_default_variables_minimal_config(self):
        """Test getting default variables with minimal config."""
        config = {
            "project": {"name": "TestProject"},
            "user": {"name": "TestUser"},
        }

        variables = TemplateEngine.get_default_variables(config)

        assert variables["PROJECT_NAME"] == "TestProject"
        assert variables["USER_NAME"] == "TestUser"

    def test_get_default_variables_full_config(self):
        """Test getting default variables with full config."""
        config = {
            "project": {
                "name": "MyProject",
                "description": "My Description",
                "owner": "Alice",
                "mode": "team",
                "codebase_language": "python",
            },
            "user": {"name": "Alice"},
            "github": {
                "templates": {
                    "spec_directory": ".moai/specs",
                    "docs_directory": ".moai/docs",
                    "test_directory": "tests",
                    "enable_trust_5": True,
                    "enable_alfred_commands": True,
                }
            },
            "language": {
                "conversation_language": "en",
                "conversation_language_name": "English",
                "agent_prompt_language": "english",
            },
            "moai": {"version": "1.0.0"},
        }

        variables = TemplateEngine.get_default_variables(config)

        assert variables["PROJECT_NAME"] == "MyProject"
        assert variables["PROJECT_DESCRIPTION"] == "My Description"
        assert variables["PROJECT_OWNER"] == "Alice"
        assert variables["CODEBASE_LANGUAGE"] == "python"
        assert variables["MOAI_VERSION"] == "1.0.0"

    def test_get_default_variables_empty_config(self):
        """Test getting default variables with empty config."""
        config = {}

        variables = TemplateEngine.get_default_variables(config)

        assert variables["PROJECT_NAME"] == "MyProject"  # Default
        assert "SPEC_DIR" in variables


class TestTemplateVariableValidator:
    """Test TemplateVariableValidator class."""

    def test_validate_required_variables_present(self):
        """Test validating when all required variables present."""
        variables = {
            "PROJECT_NAME": "MyProject",
            "PROJECT_OWNER": "Alice",
            "CODEBASE_LANGUAGE": "python",
            "SPEC_DIR": ".moai/specs",
            "DOCS_DIR": ".moai/docs",
            "TEST_DIR": "tests",
            "CONVERSATION_LANGUAGE": "en",
        }

        is_valid, errors = TemplateVariableValidator.validate(variables)

        assert is_valid is True
        assert errors == []

    def test_validate_required_variables_missing(self):
        """Test validation fails when required variables missing."""
        variables = {
            "PROJECT_NAME": "MyProject",
            # Missing PROJECT_OWNER and others
        }

        is_valid, errors = TemplateVariableValidator.validate(variables)

        assert is_valid is False
        assert len(errors) > 0

    def test_validate_variable_type_validation(self):
        """Test type validation for variables."""
        variables = {
            "PROJECT_NAME": "MyProject",
            "PROJECT_OWNER": "Alice",
            "CODEBASE_LANGUAGE": "python",
            "SPEC_DIR": ".moai/specs",
            "DOCS_DIR": ".moai/docs",
            "TEST_DIR": "tests",
            "CONVERSATION_LANGUAGE": "en",
            "ENABLE_TRUST_5": True,  # Wrong type for boolean check
        }

        is_valid, errors = TemplateVariableValidator.validate(variables)

        # Should pass because ENABLE_TRUST_5 is optional and correct type
        assert is_valid is True

    def test_validate_optional_variables(self):
        """Test validation of optional variables."""
        variables = {
            "PROJECT_NAME": "MyProject",
            "PROJECT_OWNER": "Alice",
            "CODEBASE_LANGUAGE": "python",
            "SPEC_DIR": ".moai/specs",
            "DOCS_DIR": ".moai/docs",
            "TEST_DIR": "tests",
            "CONVERSATION_LANGUAGE": "en",
            "PROJECT_DESCRIPTION": "Optional description",
            "USER_NAME": "Alice",
        }

        is_valid, errors = TemplateVariableValidator.validate(variables)

        assert is_valid is True
        assert errors == []

    def test_validate_optional_variable_wrong_type(self):
        """Test validation fails for wrong optional variable type."""
        variables = {
            "PROJECT_NAME": "MyProject",
            "PROJECT_OWNER": "Alice",
            "CODEBASE_LANGUAGE": "python",
            "SPEC_DIR": ".moai/specs",
            "DOCS_DIR": ".moai/docs",
            "TEST_DIR": "tests",
            "CONVERSATION_LANGUAGE": "en",
            "ENABLE_TRUST_5": "not_a_boolean",  # Wrong type
        }

        is_valid, errors = TemplateVariableValidator.validate(variables)

        assert is_valid is False
        assert len(errors) > 0

    def test_validate_none_for_optional(self):
        """Test validation allows None for optional variables."""
        variables = {
            "PROJECT_NAME": "MyProject",
            "PROJECT_OWNER": "Alice",
            "CODEBASE_LANGUAGE": "python",
            "SPEC_DIR": ".moai/specs",
            "DOCS_DIR": ".moai/docs",
            "TEST_DIR": "tests",
            "CONVERSATION_LANGUAGE": "en",
            "PROJECT_DESCRIPTION": None,  # Optional, can be None
            "USER_NAME": None,  # Optional, can be None
        }

        is_valid, errors = TemplateVariableValidator.validate(variables)

        assert is_valid is True


class TestCachedEnvironment:
    """Test cached Jinja2 environment."""

    def test_string_environment_caching(self):
        """Test that string environment is cached."""
        from moai_adk.core.template_engine import _get_string_environment

        env1 = _get_string_environment(strict=True)
        env2 = _get_string_environment(strict=True)

        assert env1 is env2  # Should be same instance

    def test_file_environment_caching(self):
        """Test that file environment is cached."""
        from moai_adk.core.template_engine import _get_file_environment

        with tempfile.TemporaryDirectory() as tmpdir:
            env1 = _get_file_environment(tmpdir, strict=True)
            env2 = _get_file_environment(tmpdir, strict=True)

            assert env1 is env2  # Should be same instance

    def test_different_strict_creates_different_environment(self):
        """Test that different strict values create different environments."""
        from moai_adk.core.template_engine import _get_string_environment

        env_strict = _get_string_environment(strict=True)
        env_lenient = _get_string_environment(strict=False)

        assert env_strict is not env_lenient


class TestTemplateIntegration:
    """Integration tests for template rendering."""

    def test_github_workflow_template_rendering(self):
        """Test rendering a realistic GitHub workflow template."""
        with tempfile.TemporaryDirectory() as tmpdir:
            template_path = Path(tmpdir) / "workflow.yml"
            template_path.write_text(
                """name: {{WORKFLOW_NAME}}
on:
  push:
    branches: [{{MAIN_BRANCH}}]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Run {{PROJECT_NAME}} tests
        run: pytest"""
            )

            engine = TemplateEngine()
            result = engine.render_file(
                template_path,
                {
                    "WORKFLOW_NAME": "CI",
                    "MAIN_BRANCH": "main",
                    "PROJECT_NAME": "MyProject",
                },
            )

            assert "name: CI" in result
            assert "branches: [main]" in result
            assert "MyProject" in result

    def test_config_file_template_rendering(self):
        """Test rendering configuration file template."""
        with tempfile.TemporaryDirectory() as tmpdir:
            template_path = Path(tmpdir) / "config.json"
            template_path.write_text(
                """{
  "project": "{{PROJECT_NAME}}",
  "language": "{{LANGUAGE}}",
  "features": {
    "trust5": {{ENABLE_TRUST_5|lower}}
  }
}"""
            )

            engine = TemplateEngine()
            result = engine.render_file(
                template_path,
                {
                    "PROJECT_NAME": "MyApp",
                    "LANGUAGE": "python",
                    "ENABLE_TRUST_5": True,
                },
            )

            assert '"project": "MyApp"' in result
            assert '"language": "python"' in result
            assert '"trust5": true' in result
