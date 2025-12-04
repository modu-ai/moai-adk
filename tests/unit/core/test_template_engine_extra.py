"""Extended unit tests for moai_adk.core.template_engine module.

Comprehensive tests for TemplateEngine and TemplateVariableValidator.
"""

from pathlib import Path
from unittest.mock import MagicMock, Mock, patch

import pytest

from moai_adk.core.template_engine import (
    TemplateEngine,
    TemplateVariableValidator,
    _get_file_environment,
    _get_string_environment,
)


class TestTemplateEngineInitialization:
    """Test TemplateEngine initialization."""

    def test_initialization_default_strict(self):
        """Test TemplateEngine initializes with strict undefined by default."""
        engine = TemplateEngine()
        assert engine.strict_undefined is True

    def test_initialization_non_strict(self):
        """Test TemplateEngine can initialize with non-strict mode."""
        engine = TemplateEngine(strict_undefined=False)
        assert engine.strict_undefined is False

    def test_initialization_sets_undefined_behavior(self):
        """Test initialization sets correct undefined behavior."""
        engine_strict = TemplateEngine(strict_undefined=True)
        assert engine_strict.undefined_behavior is not None

        engine_loose = TemplateEngine(strict_undefined=False)
        assert engine_loose.undefined_behavior is not None


class TestTemplateEngineRenderString:
    """Test render_string method."""

    def test_render_string_basic_substitution(self):
        """Test render_string performs basic variable substitution."""
        engine = TemplateEngine()
        template = "Hello {{ name }}!"
        variables = {"name": "World"}

        result = engine.render_string(template, variables)

        assert result == "Hello World!"

    def test_render_string_multiple_variables(self):
        """Test render_string with multiple variables."""
        engine = TemplateEngine()
        template = "{{ greeting }} {{ name }}, you are {{ age }} years old"
        variables = {"greeting": "Hello", "name": "Alice", "age": 25}

        result = engine.render_string(template, variables)

        assert result == "Hello Alice, you are 25 years old"

    def test_render_string_with_conditionals(self):
        """Test render_string with Jinja2 conditionals."""
        engine = TemplateEngine()
        template = "{% if enabled %}Enabled{% else %}Disabled{% endif %}"
        variables = {"enabled": True}

        result = engine.render_string(template, variables)

        assert result == "Enabled"

    def test_render_string_with_loops(self):
        """Test render_string with Jinja2 loops."""
        engine = TemplateEngine()
        template = "{% for item in items %}{{ item }},{% endfor %}"
        variables = {"items": ["a", "b", "c"]}

        result = engine.render_string(template, variables)

        assert "a," in result
        assert "b," in result
        assert "c," in result

    def test_render_string_undefined_variable_strict(self):
        """Test render_string raises on undefined variable in strict mode."""
        engine = TemplateEngine(strict_undefined=True)
        template = "{{ undefined_var }}"
        variables = {}

        with pytest.raises(Exception):  # TemplateRuntimeError wrapped in RuntimeError
            engine.render_string(template, variables)

    def test_render_string_undefined_variable_loose(self):
        """Test render_string handles undefined variable in loose mode."""
        engine = TemplateEngine(strict_undefined=False)
        template = "Hello {{ name }}!"
        variables = {}

        result = engine.render_string(template, variables)
        # In loose mode, undefined renders as empty string
        assert "Hello !" in result

    def test_render_string_invalid_syntax(self):
        """Test render_string raises on invalid template syntax."""
        engine = TemplateEngine()
        template = "{{ unclosed variable"
        variables = {}

        with pytest.raises(RuntimeError):
            engine.render_string(template, variables)

    def test_render_string_empty_template(self):
        """Test render_string with empty template."""
        engine = TemplateEngine()
        result = engine.render_string("", {})
        assert result == ""

    def test_render_string_no_variables_needed(self):
        """Test render_string with template needing no variables."""
        engine = TemplateEngine()
        template = "Static content"
        result = engine.render_string(template, {})
        assert result == "Static content"

    def test_render_string_special_characters(self):
        """Test render_string handles special characters."""
        engine = TemplateEngine()
        template = "{{ message }}"
        variables = {"message": "Special @#$% chars & symbols"}

        result = engine.render_string(template, variables)

        assert result == "Special @#$% chars & symbols"


class TestTemplateEngineRenderFile:
    """Test render_file method."""

    def test_render_file_basic(self, tmp_path):
        """Test render_file renders template file."""
        engine = TemplateEngine()
        template_file = tmp_path / "template.txt"
        template_file.write_text("Hello {{ name }}!")

        result = engine.render_file(template_file, {"name": "World"})

        assert result == "Hello World!"

    def test_render_file_with_output_path(self, tmp_path):
        """Test render_file writes to output path."""
        engine = TemplateEngine()
        template_file = tmp_path / "template.txt"
        output_file = tmp_path / "output.txt"

        template_file.write_text("Content: {{ value }}")
        engine.render_file(template_file, {"value": "test"}, output_file)

        assert output_file.exists()
        assert output_file.read_text() == "Content: test"

    def test_render_file_creates_output_directories(self, tmp_path):
        """Test render_file creates output directories."""
        engine = TemplateEngine()
        template_file = tmp_path / "template.txt"
        output_file = tmp_path / "deep" / "nested" / "output.txt"

        template_file.write_text("Content")
        engine.render_file(template_file, {}, output_file)

        assert output_file.exists()

    def test_render_file_nonexistent_template(self, tmp_path):
        """Test render_file raises for nonexistent template."""
        engine = TemplateEngine()
        template_file = tmp_path / "nonexistent.txt"

        with pytest.raises(FileNotFoundError):
            engine.render_file(template_file, {})

    def test_render_file_with_complex_content(self, tmp_path):
        """Test render_file with complex template content."""
        engine = TemplateEngine()
        template_file = tmp_path / "template.txt"

        content = """# {{ title }}

Name: {{ name }}
Age: {{ age }}

{% if active %}
Status: Active
{% else %}
Status: Inactive
{% endif %}
"""
        template_file.write_text(content)

        result = engine.render_file(
            template_file,
            {"title": "User", "name": "Alice", "age": 30, "active": True},
        )

        assert "# User" in result
        assert "Name: Alice" in result
        assert "Status: Active" in result


class TestTemplateEngineRenderDirectory:
    """Test render_directory method."""

    def test_render_directory_renders_all_files(self, tmp_path):
        """Test render_directory renders all template files."""
        engine = TemplateEngine()
        template_dir = tmp_path / "templates"
        output_dir = tmp_path / "output"

        template_dir.mkdir()
        (template_dir / "file1.txt").write_text("File1: {{ var }}")
        (template_dir / "file2.txt").write_text("File2: {{ var }}")

        results = engine.render_directory(
            template_dir, output_dir, {"var": "value"}
        )

        assert "file1.txt" in results
        assert "file2.txt" in results
        assert (output_dir / "file1.txt").exists()
        assert (output_dir / "file2.txt").exists()

    def test_render_directory_with_subdirs(self, tmp_path):
        """Test render_directory preserves subdirectory structure."""
        engine = TemplateEngine()
        template_dir = tmp_path / "templates"
        output_dir = tmp_path / "output"

        template_dir.mkdir()
        (template_dir / "subdir").mkdir()
        (template_dir / "subdir" / "file.txt").write_text("Content: {{ value }}")

        results = engine.render_directory(
            template_dir, output_dir, {"value": "test"}
        )

        assert "subdir/file.txt" in results or "subdir\\file.txt" in results
        assert (output_dir / "subdir" / "file.txt").exists()

    def test_render_directory_with_pattern(self, tmp_path):
        """Test render_directory with custom glob pattern."""
        engine = TemplateEngine()
        template_dir = tmp_path / "templates"
        output_dir = tmp_path / "output"

        template_dir.mkdir()
        (template_dir / "file.txt").write_text("{{ var }}")
        (template_dir / "file.md").write_text("{{ var }}")
        (template_dir / "file.json").write_text("{{ var }}")

        results = engine.render_directory(
            template_dir, output_dir, {"var": "test"}, pattern="*.txt"
        )

        assert "file.txt" in results
        assert "file.md" not in results

    def test_render_directory_nonexistent_source(self, tmp_path):
        """Test render_directory raises for nonexistent source."""
        engine = TemplateEngine()
        template_dir = tmp_path / "nonexistent"
        output_dir = tmp_path / "output"

        with pytest.raises(FileNotFoundError):
            engine.render_directory(template_dir, output_dir, {})

    def test_render_directory_returns_mapping(self, tmp_path):
        """Test render_directory returns dict of rendered content."""
        engine = TemplateEngine()
        template_dir = tmp_path / "templates"
        output_dir = tmp_path / "output"

        template_dir.mkdir()
        (template_dir / "file.txt").write_text("{{ msg }}")

        results = engine.render_directory(
            template_dir, output_dir, {"msg": "hello"}
        )

        assert isinstance(results, dict)
        assert results["file.txt"] == "hello"


class TestTemplateEngineGetDefaultVariables:
    """Test get_default_variables static method."""

    def test_get_default_variables_basic(self):
        """Test get_default_variables returns basic variables."""
        config = {
            "project": {"name": "MyProject", "description": "A project"},
            "user": {"name": "John"},
            "github": {"templates": {}},
        }

        variables = TemplateEngine.get_default_variables(config)

        assert variables["PROJECT_NAME"] == "MyProject"
        assert variables["PROJECT_DESCRIPTION"] == "A project"
        assert variables["USER_NAME"] == "John"

    def test_get_default_variables_with_github_config(self):
        """Test get_default_variables extracts GitHub configuration."""
        config = {
            "project": {"name": "MyProject", "description": ""},
            "user": {},
            "github": {
                "templates": {
                    "spec_directory": "specs/",
                    "docs_directory": "docs/",
                }
            },
        }

        variables = TemplateEngine.get_default_variables(config)

        assert variables["SPEC_DIR"] == "specs/"
        assert variables["DOCS_DIR"] == "docs/"

    def test_get_default_variables_with_language(self):
        """Test get_default_variables extracts language configuration."""
        config = {
            "project": {"name": "MyProject", "codebase_language": "python"},
            "user": {},
            "github": {"templates": {}},
            "language": {
                "conversation_language": "en",
                "conversation_language_name": "English",
            },
        }

        variables = TemplateEngine.get_default_variables(config)

        assert variables["CODEBASE_LANGUAGE"] == "python"
        assert variables["CONVERSATION_LANGUAGE"] == "en"

    def test_get_default_variables_defaults(self):
        """Test get_default_variables provides sensible defaults."""
        config = {}

        variables = TemplateEngine.get_default_variables(config)

        assert variables["PROJECT_NAME"] == "MyProject"
        assert variables["PROJECT_MODE"] == "team"
        assert variables["CONVERSATION_LANGUAGE"] == "en"

    def test_get_default_variables_with_moai_version(self):
        """Test get_default_variables includes MoAI version."""
        config = {"moai": {"version": "0.25.0"}}

        variables = TemplateEngine.get_default_variables(config)

        assert variables["MOAI_VERSION"] == "0.25.0"


class TestTemplateVariableValidator:
    """Test TemplateVariableValidator class."""

    def test_validate_required_variables_present(self):
        """Test validate with all required variables."""
        variables = {
            "PROJECT_NAME": "MyProject",
            "PROJECT_OWNER": "John",
            "CODEBASE_LANGUAGE": "python",
            "SPEC_DIR": "specs",
            "DOCS_DIR": "docs",
            "TEST_DIR": "tests",
            "CONVERSATION_LANGUAGE": "en",
        }

        is_valid, errors = TemplateVariableValidator.validate(variables)

        assert is_valid is True
        assert len(errors) == 0

    def test_validate_missing_required_variable(self):
        """Test validate with missing required variable."""
        variables = {
            "PROJECT_NAME": "MyProject",
            # Missing PROJECT_OWNER
            "CODEBASE_LANGUAGE": "python",
            "SPEC_DIR": "specs",
            "DOCS_DIR": "docs",
            "TEST_DIR": "tests",
            "CONVERSATION_LANGUAGE": "en",
        }

        is_valid, errors = TemplateVariableValidator.validate(variables)

        assert is_valid is False
        assert any("PROJECT_OWNER" in error for error in errors)

    def test_validate_wrong_type_required(self):
        """Test validate detects wrong type for required variable."""
        variables = {
            "PROJECT_NAME": 123,  # Should be string
            "PROJECT_OWNER": "John",
            "CODEBASE_LANGUAGE": "python",
            "SPEC_DIR": "specs",
            "DOCS_DIR": "docs",
            "TEST_DIR": "tests",
            "CONVERSATION_LANGUAGE": "en",
        }

        is_valid, errors = TemplateVariableValidator.validate(variables)

        assert is_valid is False
        assert any("PROJECT_NAME" in error for error in errors)

    def test_validate_optional_variables(self):
        """Test validate handles optional variables."""
        variables = {
            "PROJECT_NAME": "MyProject",
            "PROJECT_OWNER": "John",
            "CODEBASE_LANGUAGE": "python",
            "SPEC_DIR": "specs",
            "DOCS_DIR": "docs",
            "TEST_DIR": "tests",
            "CONVERSATION_LANGUAGE": "en",
            "PROJECT_DESCRIPTION": "Optional description",
            "ENABLE_TRUST_5": True,
        }

        is_valid, errors = TemplateVariableValidator.validate(variables)

        assert is_valid is True

    def test_validate_wrong_type_optional(self):
        """Test validate detects wrong type for optional variable."""
        variables = {
            "PROJECT_NAME": "MyProject",
            "PROJECT_OWNER": "John",
            "CODEBASE_LANGUAGE": "python",
            "SPEC_DIR": "specs",
            "DOCS_DIR": "docs",
            "TEST_DIR": "tests",
            "CONVERSATION_LANGUAGE": "en",
            "ENABLE_TRUST_5": "yes",  # Should be bool
        }

        is_valid, errors = TemplateVariableValidator.validate(variables)

        assert is_valid is False
        assert any("ENABLE_TRUST_5" in error for error in errors)

    def test_validate_empty_dict(self):
        """Test validate with empty dictionary."""
        is_valid, errors = TemplateVariableValidator.validate({})

        assert is_valid is False
        assert len(errors) > 0

    def test_validate_extra_variables_allowed(self):
        """Test validate ignores extra variables."""
        variables = {
            "PROJECT_NAME": "MyProject",
            "PROJECT_OWNER": "John",
            "CODEBASE_LANGUAGE": "python",
            "SPEC_DIR": "specs",
            "DOCS_DIR": "docs",
            "TEST_DIR": "tests",
            "CONVERSATION_LANGUAGE": "en",
            "EXTRA_VAR": "extra value",
            "ANOTHER_EXTRA": 123,
        }

        is_valid, errors = TemplateVariableValidator.validate(variables)

        assert is_valid is True


class TestTemplateEngineEnvironmentCaching:
    """Test environment caching functions."""

    def test_get_string_environment_caches_result(self):
        """Test _get_string_environment caches results."""
        env1 = _get_string_environment(strict=True)
        env2 = _get_string_environment(strict=True)

        assert env1 is env2  # Same instance due to caching

    def test_get_string_environment_different_strict_modes(self):
        """Test _get_string_environment different strict modes."""
        env_strict = _get_string_environment(strict=True)
        env_loose = _get_string_environment(strict=False)

        assert env_strict is not env_loose

    def test_get_file_environment_caches_result(self, tmp_path):
        """Test _get_file_environment caches results."""
        template_dir = str(tmp_path)

        env1 = _get_file_environment(template_dir, strict=True)
        env2 = _get_file_environment(template_dir, strict=True)

        assert env1 is env2  # Same instance due to caching

    def test_get_file_environment_different_dirs(self, tmp_path):
        """Test _get_file_environment different directories."""
        dir1 = str(tmp_path / "dir1")
        dir2 = str(tmp_path / "dir2")

        env1 = _get_file_environment(dir1, strict=True)
        env2 = _get_file_environment(dir2, strict=True)

        assert env1 is not env2


class TestTemplateEngineEdgeCases:
    """Test edge cases and error scenarios."""

    def test_render_string_with_filters(self):
        """Test render_string with Jinja2 filters."""
        engine = TemplateEngine()
        template = "{{ name | upper }}"
        variables = {"name": "world"}

        result = engine.render_string(template, variables)

        assert result == "WORLD"

    def test_render_string_with_macros(self):
        """Test render_string with Jinja2 macros."""
        engine = TemplateEngine()
        template = """
{% macro greet(name) %}
Hello {{ name }}!
{% endmacro %}
{{ greet('Alice') }}
"""
        result = engine.render_string(template, {})

        assert "Hello Alice" in result

    def test_render_file_with_unicode(self, tmp_path):
        """Test render_file handles unicode content."""
        engine = TemplateEngine()
        template_file = tmp_path / "template.txt"
        content = "Hello {{ greeting }} 你好 مرحبا"

        template_file.write_text(content, encoding="utf-8")

        result = engine.render_file(
            template_file, {"greeting": "world"}
        )

        assert "Hello world" in result

    def test_render_directory_empty_dir(self, tmp_path):
        """Test render_directory with empty directory."""
        engine = TemplateEngine()
        template_dir = tmp_path / "templates"
        output_dir = tmp_path / "output"

        template_dir.mkdir()

        results = engine.render_directory(template_dir, output_dir, {})

        assert isinstance(results, dict)
        assert len(results) == 0
