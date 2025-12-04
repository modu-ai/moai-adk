"""Enhanced unit tests for TemplateProcessor module with extended coverage.

This module tests:
- TemplateProcessorConfig dataclass and factory methods
- TemplateProcessor initialization and core methods
- Version formatting and validation
- Template variable substitution and caching
- File operations and variable substitution
- Error handling and edge cases
"""

import json
import tempfile
from datetime import datetime
from pathlib import Path
from unittest import mock

import pytest
import yaml

from moai_adk.core.template.processor import (
    TemplateProcessor,
    TemplateProcessorConfig,
)


class TestTemplateProcessorConfig:
    """Test TemplateProcessorConfig dataclass."""

    def test_config_default_initialization(self):
        """Test config creation with default values."""
        config = TemplateProcessorConfig()
        assert config.version_cache_ttl_seconds == 120
        assert config.version_fallback == "unknown"
        assert config.enable_version_validation is True
        assert config.preserve_user_version is True
        assert config.validate_template_variables is True
        assert config.enable_caching is True
        assert config.cache_size == 100

    def test_config_custom_initialization(self):
        """Test config creation with custom values."""
        config = TemplateProcessorConfig(
            version_cache_ttl_seconds=300,
            version_fallback="v0.0.0",
            enable_version_validation=False,
            cache_size=50,
        )
        assert config.version_cache_ttl_seconds == 300
        assert config.version_fallback == "v0.0.0"
        assert config.enable_version_validation is False
        assert config.cache_size == 50

    def test_config_from_dict_full(self):
        """Test creating config from complete dictionary."""
        config_dict = {
            "version_cache_ttl_seconds": 180,
            "version_fallback": "v1.0.0",
            "enable_version_validation": False,
            "cache_size": 200,
            "async_operations": True,
        }
        config = TemplateProcessorConfig.from_dict(config_dict)
        assert config.version_cache_ttl_seconds == 180
        assert config.version_fallback == "v1.0.0"
        assert config.enable_version_validation is False
        assert config.cache_size == 200
        assert config.async_operations is True

    def test_config_from_dict_partial(self):
        """Test creating config from partial dictionary."""
        config_dict = {"version_fallback": "v2.0.0"}
        config = TemplateProcessorConfig.from_dict(config_dict)
        assert config.version_fallback == "v2.0.0"
        assert config.version_cache_ttl_seconds == 120  # default

    def test_config_from_dict_empty(self):
        """Test creating config from empty dictionary."""
        config = TemplateProcessorConfig.from_dict({})
        assert config.version_cache_ttl_seconds == 120
        assert config.version_fallback == "unknown"

    def test_config_from_dict_none(self):
        """Test creating config from None."""
        config = TemplateProcessorConfig.from_dict(None)
        assert config.version_cache_ttl_seconds == 120
        assert config.version_fallback == "unknown"


class TestTemplateProcessorInitialization:
    """Test TemplateProcessor initialization."""

    def test_processor_initialization_default(self):
        """Test processor initialization with default config."""
        with tempfile.TemporaryDirectory() as temp_dir:
            target_path = Path(temp_dir)
            processor = TemplateProcessor(target_path)

            assert processor.target_path == target_path.resolve()
            assert processor.config is not None
            assert len(processor.context) == 0
            assert len(processor._substitution_cache) == 0

    def test_processor_initialization_custom_config(self):
        """Test processor initialization with custom config."""
        with tempfile.TemporaryDirectory() as temp_dir:
            target_path = Path(temp_dir)
            config = TemplateProcessorConfig(cache_size=50)
            processor = TemplateProcessor(target_path, config)

            assert processor.config.cache_size == 50
            assert processor.target_path == target_path.resolve()

    def test_processor_protected_paths_constant(self):
        """Test that PROTECTED_PATHS contains expected values."""
        assert ".moai/specs/" in TemplateProcessor.PROTECTED_PATHS
        assert ".moai/reports/" in TemplateProcessor.PROTECTED_PATHS
        assert ".moai/project/" in TemplateProcessor.PROTECTED_PATHS

    def test_processor_common_template_variables(self):
        """Test COMMON_TEMPLATE_VARIABLES constant."""
        assert "PROJECT_DIR" in TemplateProcessor.COMMON_TEMPLATE_VARIABLES
        assert "MOAI_VERSION" in TemplateProcessor.COMMON_TEMPLATE_VARIABLES
        assert "CONVERSATION_LANGUAGE" in TemplateProcessor.COMMON_TEMPLATE_VARIABLES


class TestVersionFormatting:
    """Test version formatting methods."""

    def test_format_short_version_with_v_prefix(self):
        """Test short version formatting removes v prefix."""
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))
            result = processor._format_short_version("v1.2.3")
            assert result == "1.2.3"

    def test_format_short_version_without_prefix(self):
        """Test short version formatting for version without prefix."""
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))
            result = processor._format_short_version("1.2.3")
            assert result == "1.2.3"

    def test_format_display_version_unknown(self):
        """Test display version formatting for unknown version."""
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))
            result = processor._format_display_version("unknown")
            assert result == "MoAI-ADK unknown version"

    def test_format_display_version_with_v(self):
        """Test display version formatting with v prefix."""
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))
            result = processor._format_display_version("v1.2.3")
            assert result == "MoAI-ADK v1.2.3"

    def test_format_display_version_without_v(self):
        """Test display version formatting without v prefix."""
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))
            result = processor._format_display_version("1.2.3")
            assert result == "MoAI-ADK v1.2.3"

    def test_format_trimmed_version(self):
        """Test trimmed version formatting."""
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))
            result = processor._format_trimmed_version("v1.2.3-beta", max_length=5)
            assert len(result) <= 5

    def test_format_semver_version_valid(self):
        """Test semantic version formatting with valid version."""
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))
            result = processor._format_semver_version("v1.2.3-beta")
            assert result == "1.2.3"

    def test_format_semver_version_unknown(self):
        """Test semantic version formatting for unknown."""
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))
            result = processor._format_semver_version("unknown")
            assert result == "0.0.0"


class TestVersionValidation:
    """Test version validation methods."""

    def test_is_valid_version_format_valid(self):
        """Test version format validation with valid formats."""
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))
            assert processor._is_valid_version_format("v1.2.3") is True
            assert processor._is_valid_version_format("1.2.3") is True
            assert processor._is_valid_version_format("v1.2.3-beta") is True

    def test_is_valid_version_format_invalid(self):
        """Test version format validation with invalid formats."""
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))
            assert processor._is_valid_version_format("invalid") is False
            assert processor._is_valid_version_format("1.2") is False


class TestTemplateVariableValidation:
    """Test template variable validation."""

    def test_validate_template_variables_valid(self):
        """Test validation of valid template variables."""
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))
            context = {
                "PROJECT_DIR": "/path/to/project",
                "PROJECT_NAME": "MyProject",
            }
            # Should not raise
            processor._validate_template_variables(context)

    def test_validate_template_variables_with_graceful_degradation(self):
        """Test validation with graceful degradation enabled."""
        with tempfile.TemporaryDirectory() as temp_dir:
            config = TemplateProcessorConfig(graceful_degradation=True)
            processor = TemplateProcessor(Path(temp_dir), config)
            context = {
                "INVALID!VAR": "value",
            }
            # Should not raise due to graceful degradation
            processor._validate_template_variables(context)

    def test_is_valid_template_variable_valid(self):
        """Test individual variable validation."""
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))
            assert processor._is_valid_template_variable("PROJECT_DIR", "/path") is True
            assert processor._is_valid_template_variable("NAME", "test") is True

    def test_is_valid_template_variable_empty_value(self):
        """Test variable validation with empty value."""
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))
            assert processor._is_valid_template_variable("PROJECT_DIR", "") is False
            assert processor._is_valid_template_variable("PROJECT_DIR", "   ") is False


class TestTemplateSubstitution:
    """Test template variable substitution."""

    def test_substitute_variables_simple(self):
        """Test simple variable substitution."""
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))
            processor.set_context({"PROJECT_NAME": "MyProject"})

            content = "Project: {{PROJECT_NAME}}"
            result, warnings = processor._substitute_variables(content)

            assert "MyProject" in result
            assert "{{PROJECT_NAME}}" not in result

    def test_substitute_variables_multiple(self):
        """Test substitution of multiple variables."""
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))
            processor.set_context({
                "PROJECT_NAME": "MyProject",
                "AUTHOR": "John Doe",
            })

            content = "Project: {{PROJECT_NAME}}, Author: {{AUTHOR}}"
            result, warnings = processor._substitute_variables(content)

            assert "MyProject" in result
            assert "John Doe" in result

    def test_substitute_variables_missing(self):
        """Test substitution with missing variables."""
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))
            processor.set_context({"PROJECT_NAME": "MyProject"})

            content = "Project: {{PROJECT_NAME}}, Author: {{AUTHOR}}"
            result, warnings = processor._substitute_variables(content)

            assert "MyProject" in result
            assert len(warnings) > 0  # Should have warning about missing AUTHOR

    def test_substitute_variables_caching(self):
        """Test variable substitution caching."""
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))
            processor.set_context({"PROJECT_NAME": "MyProject"})

            content = "Project: {{PROJECT_NAME}}"
            result1, _ = processor._substitute_variables(content)
            result2, _ = processor._substitute_variables(content)

            assert result1 == result2
            assert len(processor._substitution_cache) == 1


class TestCacheManagement:
    """Test cache functionality."""

    def test_clear_substitution_cache(self):
        """Test clearing substitution cache."""
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))
            processor.set_context({"PROJECT_NAME": "MyProject"})

            content = "Project: {{PROJECT_NAME}}"
            processor._substitute_variables(content)
            assert len(processor._substitution_cache) > 0

            processor.clear_substitution_cache()
            assert len(processor._substitution_cache) == 0

    def test_get_cache_stats(self):
        """Test getting cache statistics."""
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))
            processor.set_context({"PROJECT_NAME": "MyProject"})

            content = "Project: {{PROJECT_NAME}}"
            processor._substitute_variables(content)

            stats = processor.get_cache_stats()
            assert "cache_size" in stats
            assert "max_cache_size" in stats
            assert "cache_enabled" in stats
            assert stats["cache_enabled"] is True


class TestFileOperations:
    """Test file handling operations."""

    def test_is_text_file_markdown(self):
        """Test text file detection for markdown."""
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))
            assert processor._is_text_file(Path("test.md")) is True
            assert processor._is_text_file(Path("test.json")) is True
            assert processor._is_text_file(Path("test.py")) is True

    def test_is_text_file_binary(self):
        """Test text file detection for binary files."""
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))
            assert processor._is_text_file(Path("test.bin")) is False
            assert processor._is_text_file(Path("test.exe")) is False

    def test_sanitize_value(self):
        """Test value sanitization."""
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))
            result = processor._sanitize_value("test {{value}}")
            assert "{{" not in result
            assert "}}" not in result

    def test_sanitize_value_preserves_normal_text(self):
        """Test value sanitization preserves normal text."""
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))
            result = processor._sanitize_value("normal text value")
            assert result == "normal text value"


class TestVersionContext:
    """Test version context generation."""

    def test_get_enhanced_version_context_structure(self):
        """Test enhanced version context has expected keys."""
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))
            context = processor.get_enhanced_version_context()

            assert "MOAI_VERSION" in context
            assert "MOAI_VERSION_SHORT" in context
            assert "MOAI_VERSION_DISPLAY" in context
            assert "MOAI_VERSION_VALID" in context

    def test_get_enhanced_version_context_fallback(self):
        """Test version context with fallback version."""
        with tempfile.TemporaryDirectory() as temp_dir:
            config = TemplateProcessorConfig(version_fallback="v1.0.0")
            processor = TemplateProcessor(Path(temp_dir), config)
            context = processor.get_enhanced_version_context()

            # Context may have actual version or fallback
            assert "MOAI_VERSION" in context
            assert isinstance(context["MOAI_VERSION"], str)


class TestSetContext:
    """Test setting context and variable validation."""

    def test_set_context_basic(self):
        """Test setting context."""
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))
            context = {"PROJECT_NAME": "MyProject"}
            processor.set_context(context)

            assert processor.context["PROJECT_NAME"] == "MyProject"

    def test_set_context_clears_cache(self):
        """Test that setting context clears substitution cache."""
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))
            processor.set_context({"PROJECT_NAME": "MyProject"})

            content = "Project: {{PROJECT_NAME}}"
            processor._substitute_variables(content)
            assert len(processor._substitution_cache) > 0

            processor.set_context({"PROJECT_NAME": "NewProject"})
            assert len(processor._substitution_cache) == 0

    def test_set_context_adds_hook_project_dir_mapping(self):
        """Test that setting context creates HOOK_PROJECT_DIR mapping."""
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))
            processor.set_context({"PROJECT_DIR": "/path/to/project"})

            assert processor.context["HOOK_PROJECT_DIR"] == "/path/to/project"


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
