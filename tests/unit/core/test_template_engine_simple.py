"""
Simple, working tests for template_engine.py module.

Tests TemplateEngine and TemplateVariableValidator with proper mocking.
Target: 70%+ coverage
"""

import tempfile
from pathlib import Path
from unittest.mock import MagicMock, Mock, patch

import pytest

from moai_adk.core.template_engine import (
    TemplateEngine,
    TemplateVariableValidator,
    _get_file_environment,
    _get_string_environment,
)


class TestGetStringEnvironment:
    """Test _get_string_environment module function."""

    def test_get_string_environment_strict_true(self):
        """Test getting environment with strict undefined."""
        # Act
        env = _get_string_environment(strict=True)

        # Assert
        assert env is not None

    def test_get_string_environment_strict_false(self):
        """Test getting environment with non-strict undefined."""
        # Act
        env = _get_string_environment(strict=False)

        # Assert
        assert env is not None

    def test_get_string_environment_caching(self):
        """Test that environments are cached."""
        # Act
        env1 = _get_string_environment(strict=True)
        env2 = _get_string_environment(strict=True)

        # Assert
        assert env1 is env2


class TestGetFileEnvironment:
    """Test _get_file_environment module function."""

    def test_get_file_environment_strict_true(self):
        """Test getting file environment with strict undefined."""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            # Act
            env = _get_file_environment(tmpdir, strict=True)

            # Assert
            assert env is not None

    def test_get_file_environment_strict_false(self):
        """Test getting file environment with non-strict undefined."""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            # Act
            env = _get_file_environment(tmpdir, strict=False)

            # Assert
            assert env is not None


class TestTemplateEngineInit:
    """Test TemplateEngine initialization."""

    def test_init_default_strict_undefined(self):
        """Test initialization with default strict_undefined."""
        # Act
        engine = TemplateEngine()

        # Assert
        assert engine.strict_undefined is True

    def test_init_strict_undefined_true(self):
        """Test initialization with strict_undefined=True."""
        # Act
        engine = TemplateEngine(strict_undefined=True)

        # Assert
        assert engine.strict_undefined is True

    def test_init_strict_undefined_false(self):
        """Test initialization with strict_undefined=False."""
        # Act
        engine = TemplateEngine(strict_undefined=False)

        # Assert
        assert engine.strict_undefined is False


class TestRenderString:
    """Test render_string method."""

    def test_render_simple_variable(self):
        """Test rendering simple variable substitution."""
        # Arrange
        engine = TemplateEngine()
        template = "Hello {{name}}!"
        variables = {"name": "World"}

        # Act
        result = engine.render_string(template, variables)

        # Assert
        assert result == "Hello World!"

    def test_render_multiple_variables(self):
        """Test rendering with multiple variables."""
        # Arrange
        engine = TemplateEngine()
        template = "{{greeting}} {{name}}, welcome to {{place}}!"
        variables = {"greeting": "Hello", "name": "Alice", "place": "Python"}

        # Act
        result = engine.render_string(template, variables)

        # Assert
        assert result == "Hello Alice, welcome to Python!"

    def test_render_with_conditional(self):
        """Test rendering with conditional section."""
        # Arrange
        engine = TemplateEngine()
        template = "{% if enabled %}Feature is enabled{% endif %}"
        variables = {"enabled": True}

        # Act
        result = engine.render_string(template, variables)

        # Assert
        assert "Feature is enabled" in result

    def test_render_undefined_variable_strict_mode(self):
        """Test undefined variable in strict mode raises error."""
        # Arrange
        engine = TemplateEngine(strict_undefined=True)
        template = "Hello {{name}}!"
        variables = {}

        # Act & Assert
        with pytest.raises(RuntimeError):
            engine.render_string(template, variables)

    def test_render_undefined_variable_non_strict_mode(self):
        """Test undefined variable in non-strict mode."""
        # Arrange
        engine = TemplateEngine(strict_undefined=False)
        template = "Hello {{name}}!"
        variables = {}

        # Act
        result = engine.render_string(template, variables)

        # Assert
        assert "Hello" in result

    def test_render_with_loop(self):
        """Test rendering with for loop."""
        # Arrange
        engine = TemplateEngine()
        template = "{% for item in items %}{{item}},{% endfor %}"
        variables = {"items": ["a", "b", "c"]}

        # Act
        result = engine.render_string(template, variables)

        # Assert
        assert "a," in result
        assert "b," in result
        assert "c," in result

    def test_render_empty_template(self):
        """Test rendering empty template."""
        # Arrange
        engine = TemplateEngine()
        template = ""
        variables = {}

        # Act
        result = engine.render_string(template, variables)

        # Assert
        assert result == ""

    def test_render_template_syntax_error(self):
        """Test error on invalid template syntax."""
        # Arrange
        engine = TemplateEngine()
        template = "{% if invalid %}"
        variables = {}

        # Act & Assert
        with pytest.raises(RuntimeError):
            engine.render_string(template, variables)


class TestRenderFile:
    """Test render_file method."""

    def test_render_file_success(self):
        """Test rendering file successfully."""
        # Arrange
        engine = TemplateEngine()
        with tempfile.TemporaryDirectory() as tmpdir:
            template_dir = Path(tmpdir)
            template_file = template_dir / "test.txt"
            template_file.write_text("Hello {{name}}!")

            variables = {"name": "World"}

            # Act
            result = engine.render_file(template_file, variables)

            # Assert
            assert result == "Hello World!"

    def test_render_file_with_output_path(self):
        """Test rendering file and writing to output."""
        # Arrange
        engine = TemplateEngine()
        with tempfile.TemporaryDirectory() as tmpdir:
            template_dir = Path(tmpdir)
            template_file = template_dir / "test.txt"
            template_file.write_text("Content: {{text}}")

            output_dir = Path(tmpdir) / "output"
            output_file = output_dir / "result.txt"

            variables = {"text": "Success"}

            # Act
            result = engine.render_file(template_file, variables, output_file)

            # Assert
            assert result == "Content: Success"
            assert output_file.exists()
            assert output_file.read_text() == "Content: Success"

    def test_render_file_not_found(self):
        """Test error when template file not found."""
        # Arrange
        engine = TemplateEngine()
        with tempfile.TemporaryDirectory() as tmpdir:
            template_file = Path(tmpdir) / "nonexistent.txt"
            variables = {}

            # Act & Assert
            with pytest.raises(FileNotFoundError):
                engine.render_file(template_file, variables)

    def test_render_file_creates_parent_directories(self):
        """Test that parent directories are created for output file."""
        # Arrange
        engine = TemplateEngine()
        with tempfile.TemporaryDirectory() as tmpdir:
            template_dir = Path(tmpdir)
            template_file = template_dir / "test.txt"
            template_file.write_text("Test")

            output_file = Path(tmpdir) / "deep" / "nested" / "output.txt"
            variables = {}

            # Act
            engine.render_file(template_file, variables, output_file)

            # Assert
            assert output_file.parent.exists()
            assert output_file.exists()


class TestRenderDirectory:
    """Test render_directory method."""

    def test_render_directory_success(self):
        """Test rendering directory with multiple files."""
        # Arrange
        engine = TemplateEngine()
        with tempfile.TemporaryDirectory() as tmpdir:
            template_dir = Path(tmpdir) / "templates"
            template_dir.mkdir()

            (template_dir / "file1.txt").write_text("File 1: {{content}}")
            (template_dir / "file2.txt").write_text("File 2: {{content}}")

            output_dir = Path(tmpdir) / "output"
            variables = {"content": "data"}

            # Act
            result = engine.render_directory(template_dir, output_dir, variables)

            # Assert
            assert len(result) == 2
            assert "file1.txt" in result
            assert "file2.txt" in result

    def test_render_directory_nested_files(self):
        """Test rendering directory with nested file structure."""
        # Arrange
        engine = TemplateEngine()
        with tempfile.TemporaryDirectory() as tmpdir:
            template_dir = Path(tmpdir) / "templates"
            template_dir.mkdir()

            (template_dir / "subdir").mkdir()
            (template_dir / "subdir" / "file.txt").write_text("{{text}}")

            output_dir = Path(tmpdir) / "output"
            variables = {"text": "nested"}

            # Act
            result = engine.render_directory(template_dir, output_dir, variables)

            # Assert
            assert "subdir/file.txt" in result or "subdir\\file.txt" in result

    def test_render_directory_not_found(self):
        """Test error when template directory not found."""
        # Arrange
        engine = TemplateEngine()
        with tempfile.TemporaryDirectory() as tmpdir:
            template_dir = Path(tmpdir) / "nonexistent"
            output_dir = Path(tmpdir) / "output"

            # Act & Assert
            with pytest.raises(FileNotFoundError):
                engine.render_directory(template_dir, output_dir, {})

    def test_render_directory_with_pattern(self):
        """Test rendering with glob pattern filter."""
        # Arrange
        engine = TemplateEngine()
        with tempfile.TemporaryDirectory() as tmpdir:
            template_dir = Path(tmpdir) / "templates"
            template_dir.mkdir()

            (template_dir / "file.txt").write_text("{{text}}")
            (template_dir / "ignore.md").write_text("{{text}}")

            output_dir = Path(tmpdir) / "output"
            variables = {"text": "data"}

            # Act
            result = engine.render_directory(template_dir, output_dir, variables, pattern="*.txt")

            # Assert
            assert "file.txt" in result
            assert "ignore.md" not in result


class TestGetDefaultVariables:
    """Test get_default_variables static method."""

    def test_get_default_variables_full_config(self):
        """Test extracting variables from full config."""
        # Arrange
        config = {
            "project": {
                "name": "MyProject",
                "description": "Test project",
                "owner": "test_owner",
                "mode": "team",
                "codebase_language": "python",
            },
            "user": {"name": "Test User"},
            "language": {
                "conversation_language": "en",
                "conversation_language_name": "English",
                "agent_prompt_language": "english",
            },
            "github": {
                "templates": {
                    "spec_directory": ".moai/specs",
                    "docs_directory": ".moai/docs",
                    "test_directory": "tests",
                    "enable_trust_5": True,
                    "enable_alfred_commands": True,
                }
            },
            "moai": {"version": "1.0.0"},
        }

        # Act
        variables = TemplateEngine.get_default_variables(config)

        # Assert
        assert variables["PROJECT_NAME"] == "MyProject"
        assert variables["PROJECT_OWNER"] == "test_owner"
        assert variables["CODEBASE_LANGUAGE"] == "python"
        assert variables["CONVERSATION_LANGUAGE"] == "en"

    def test_get_default_variables_minimal_config(self):
        """Test extracting variables from minimal config."""
        # Arrange
        config = {"project": {}, "language": {}, "github": {"templates": {}}}

        # Act
        variables = TemplateEngine.get_default_variables(config)

        # Assert
        assert variables["PROJECT_NAME"] == "MyProject"
        assert variables["CODEBASE_LANGUAGE"] == "python"
        assert variables["CONVERSATION_LANGUAGE"] == "en"

    def test_get_default_variables_empty_config(self):
        """Test extracting variables from empty config."""
        # Arrange
        config = {}

        # Act
        variables = TemplateEngine.get_default_variables(config)

        # Assert
        assert "PROJECT_NAME" in variables
        assert "CODEBASE_LANGUAGE" in variables


class TestTemplateVariableValidator:
    """Test TemplateVariableValidator class."""

    def test_validate_complete_variables(self):
        """Test validation with all required variables."""
        # Arrange
        variables = {
            "PROJECT_NAME": "MyProject",
            "PROJECT_OWNER": "owner",
            "CODEBASE_LANGUAGE": "python",
            "SPEC_DIR": ".moai/specs",
            "DOCS_DIR": ".moai/docs",
            "TEST_DIR": "tests",
            "CONVERSATION_LANGUAGE": "en",
        }

        # Act
        is_valid, errors = TemplateVariableValidator.validate(variables)

        # Assert
        assert is_valid is True
        assert len(errors) == 0

    def test_validate_missing_required_variable(self):
        """Test validation fails with missing required variable."""
        # Arrange
        variables = {
            "PROJECT_NAME": "MyProject",
            "PROJECT_OWNER": "owner",
            "CODEBASE_LANGUAGE": "python",
            "SPEC_DIR": ".moai/specs",
            "DOCS_DIR": ".moai/docs",
            "TEST_DIR": "tests",
            # Missing CONVERSATION_LANGUAGE
        }

        # Act
        is_valid, errors = TemplateVariableValidator.validate(variables)

        # Assert
        assert is_valid is False
        assert len(errors) > 0

    def test_validate_wrong_type_required_variable(self):
        """Test validation fails with wrong variable type."""
        # Arrange
        variables = {
            "PROJECT_NAME": 123,  # Should be string
            "PROJECT_OWNER": "owner",
            "CODEBASE_LANGUAGE": "python",
            "SPEC_DIR": ".moai/specs",
            "DOCS_DIR": ".moai/docs",
            "TEST_DIR": "tests",
            "CONVERSATION_LANGUAGE": "en",
        }

        # Act
        is_valid, errors = TemplateVariableValidator.validate(variables)

        # Assert
        assert is_valid is False
        assert any("PROJECT_NAME" in error for error in errors)

    def test_validate_optional_variables_valid(self):
        """Test validation with valid optional variables."""
        # Arrange
        variables = {
            "PROJECT_NAME": "MyProject",
            "PROJECT_OWNER": "owner",
            "CODEBASE_LANGUAGE": "python",
            "SPEC_DIR": ".moai/specs",
            "DOCS_DIR": ".moai/docs",
            "TEST_DIR": "tests",
            "CONVERSATION_LANGUAGE": "en",
            "PROJECT_DESCRIPTION": "A test project",
            "PROJECT_MODE": "personal",
        }

        # Act
        is_valid, errors = TemplateVariableValidator.validate(variables)

        # Assert
        assert is_valid is True

    def test_validate_optional_variables_wrong_type(self):
        """Test validation fails with wrong optional variable type."""
        # Arrange
        variables = {
            "PROJECT_NAME": "MyProject",
            "PROJECT_OWNER": "owner",
            "CODEBASE_LANGUAGE": "python",
            "SPEC_DIR": ".moai/specs",
            "DOCS_DIR": ".moai/docs",
            "TEST_DIR": "tests",
            "CONVERSATION_LANGUAGE": "en",
            "ENABLE_TRUST_5": "yes",  # Should be boolean
        }

        # Act
        is_valid, errors = TemplateVariableValidator.validate(variables)

        # Assert
        assert is_valid is False

    def test_validate_none_optional_variable(self):
        """Test validation allows None for optional variables."""
        # Arrange
        variables = {
            "PROJECT_NAME": "MyProject",
            "PROJECT_OWNER": "owner",
            "CODEBASE_LANGUAGE": "python",
            "SPEC_DIR": ".moai/specs",
            "DOCS_DIR": ".moai/docs",
            "TEST_DIR": "tests",
            "CONVERSATION_LANGUAGE": "en",
            "PROJECT_DESCRIPTION": None,
            "USER_NAME": None,
        }

        # Act
        is_valid, errors = TemplateVariableValidator.validate(variables)

        # Assert
        assert is_valid is True


class TestTemplateEngineErrorHandling:
    """Test error handling in TemplateEngine."""

    def test_render_string_with_filter(self):
        """Test rendering with Jinja2 filters."""
        # Arrange
        engine = TemplateEngine()
        template = "{{name | upper}}"
        variables = {"name": "world"}

        # Act
        result = engine.render_string(template, variables)

        # Assert
        assert result == "WORLD"

    def test_render_with_arithmetic(self):
        """Test rendering with arithmetic expressions."""
        # Arrange
        engine = TemplateEngine()
        template = "Total: {{count * price}}"
        variables = {"count": 5, "price": 10}

        # Act
        result = engine.render_string(template, variables)

        # Assert
        assert result == "Total: 50"

    def test_render_with_dictionary_access(self):
        """Test rendering with dictionary variable access."""
        # Arrange
        engine = TemplateEngine()
        template = "Name: {{user.name}}"
        variables = {"user": {"name": "Alice"}}

        # Act
        result = engine.render_string(template, variables)

        # Assert
        assert result == "Name: Alice"
