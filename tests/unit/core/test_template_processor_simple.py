"""
Simple, working tests for moai_adk.core.template.processor module.

Focus: TemplateProcessor class initialization, variable substitution, and template loading.
Target: 50%+ code coverage with AAA pattern.
"""

from pathlib import Path
from unittest.mock import patch

import pytest

from moai_adk.core.template.processor import (
    TemplateProcessor,
    TemplateProcessorConfig,
)


class TestTemplateProcessorConfig:
    """Test TemplateProcessorConfig initialization and factory method."""

    def test_config_init_default_values(self):
        """Test config initialization with default values."""
        # Arrange & Act
        config = TemplateProcessorConfig()

        # Assert
        assert config.version_cache_ttl_seconds == 120
        assert config.version_fallback == "unknown"
        assert config.validate_template_variables is True
        assert config.enable_caching is True

    def test_config_from_dict_with_empty_dict(self):
        """Test config creation from empty dictionary."""
        # Arrange
        empty_config = {}

        # Act
        config = TemplateProcessorConfig.from_dict(empty_config)

        # Assert
        assert config.version_cache_ttl_seconds == 120
        assert config.version_fallback == "unknown"

    def test_config_from_dict_with_values(self):
        """Test config creation from dictionary with custom values."""
        # Arrange
        config_dict = {
            "version_cache_ttl_seconds": 240,
            "version_fallback": "v0.0.0",
            "validate_template_variables": False,
            "enable_caching": False,
        }

        # Act
        config = TemplateProcessorConfig.from_dict(config_dict)

        # Assert
        assert config.version_cache_ttl_seconds == 240
        assert config.version_fallback == "v0.0.0"
        assert config.validate_template_variables is False
        assert config.enable_caching is False

    def test_config_from_dict_with_none(self):
        """Test config creation from None (edge case)."""
        # Arrange & Act
        config = TemplateProcessorConfig.from_dict(None)

        # Assert
        assert config.version_cache_ttl_seconds == 120


class TestTemplateProcessorInit:
    """Test TemplateProcessor initialization."""

    def test_init_with_default_config(self):
        """Test initialization with default configuration."""
        # Arrange
        target_path = Path("/test/project")

        # Act
        with patch.object(TemplateProcessor, "_get_template_root") as mock_template:
            mock_template.return_value = Path("/test/templates")
            processor = TemplateProcessor(target_path)

        # Assert
        assert processor.target_path == target_path.resolve()
        assert processor.config is not None
        assert isinstance(processor.config, TemplateProcessorConfig)

    def test_init_with_custom_config(self):
        """Test initialization with custom configuration."""
        # Arrange
        target_path = Path("/test/project")
        custom_config = TemplateProcessorConfig(version_cache_ttl_seconds=300)

        # Act
        with patch.object(TemplateProcessor, "_get_template_root") as mock_template:
            mock_template.return_value = Path("/test/templates")
            processor = TemplateProcessor(target_path, config=custom_config)

        # Assert
        assert processor.config.version_cache_ttl_seconds == 300

    def test_init_initializes_caches(self):
        """Test that initialization sets up caches."""
        # Arrange
        target_path = Path("/test/project")

        # Act
        with patch.object(TemplateProcessor, "_get_template_root") as mock_template:
            mock_template.return_value = Path("/test/templates")
            processor = TemplateProcessor(target_path)

        # Assert
        assert isinstance(processor._substitution_cache, dict)
        assert len(processor._substitution_cache) == 0
        assert isinstance(processor._variable_validation_cache, dict)

    def test_init_creates_context(self):
        """Test that initialization creates empty context."""
        # Arrange
        target_path = Path("/test/project")

        # Act
        with patch.object(TemplateProcessor, "_get_template_root") as mock_template:
            mock_template.return_value = Path("/test/templates")
            processor = TemplateProcessor(target_path)

        # Assert
        assert isinstance(processor.context, dict)
        assert len(processor.context) == 0


class TestTemplateProcessorSetContext:
    """Test variable substitution context setting."""

    def test_set_context_stores_variables(self):
        """Test that set_context stores variables."""
        # Arrange
        target_path = Path("/test/project")
        test_context = {
            "PROJECT_DIR": "/home/user/project",
            "PROJECT_NAME": "MyProject",
        }

        with patch.object(TemplateProcessor, "_get_template_root") as mock_template:
            mock_template.return_value = Path("/test/templates")
            processor = TemplateProcessor(target_path)

            # Act
            processor.set_context(test_context)

        # Assert
        assert processor.context["PROJECT_DIR"] == "/home/user/project"
        assert processor.context["PROJECT_NAME"] == "MyProject"

    def test_set_context_clears_cache(self):
        """Test that set_context clears substitution cache."""
        # Arrange
        target_path = Path("/test/project")
        processor = TemplateProcessor(target_path)
        processor._substitution_cache[123] = ("test", [])

        # Act
        processor.set_context({"TEST": "value"})

        # Assert
        assert len(processor._substitution_cache) == 0

    def test_set_context_adds_deprecated_mapping(self):
        """Test that set_context adds HOOK_PROJECT_DIR for backward compatibility."""
        # Arrange
        target_path = Path("/test/project")
        test_context = {"PROJECT_DIR": "/test/dir"}

        with patch.object(TemplateProcessor, "_get_template_root") as mock_template:
            mock_template.return_value = Path("/test/templates")
            processor = TemplateProcessor(target_path)

            # Act
            processor.set_context(test_context)

        # Assert
        assert processor.context["HOOK_PROJECT_DIR"] == "/test/dir"

    def test_set_context_with_validation_disabled(self):
        """Test set_context with validation disabled."""
        # Arrange
        config = TemplateProcessorConfig(validate_template_variables=False)
        processor = TemplateProcessor(Path("/test"), config=config)

        # Act
        processor.set_context({"invalid-name": "value"})

        # Assert (should not raise)
        assert processor.context["invalid-name"] == "value"


class TestTemplateProcessorSubstitution:
    """Test variable substitution functionality."""

    def test_substitute_variables_simple(self):
        """Test simple variable substitution."""
        # Arrange
        processor = TemplateProcessor(Path("/test"))
        processor.set_context({"NAME": "John"})
        content = "Hello {{NAME}}"

        # Act
        result, warnings = processor._substitute_variables(content)

        # Assert
        assert result == "Hello John"
        assert len(warnings) == 0

    def test_substitute_variables_multiple(self):
        """Test substitution of multiple variables."""
        # Arrange
        processor = TemplateProcessor(Path("/test"))
        processor.set_context(
            {
                "FIRST": "John",
                "LAST": "Doe",
            }
        )
        content = "{{FIRST}} {{LAST}}"

        # Act
        result, warnings = processor._substitute_variables(content)

        # Assert
        assert result == "John Doe"

    def test_substitute_variables_with_caching(self):
        """Test that substitution uses caching when enabled."""
        # Arrange
        config = TemplateProcessorConfig(enable_caching=True)
        processor = TemplateProcessor(Path("/test"), config=config)
        processor.set_context({"VAR": "value"})
        content = "Test {{VAR}}"

        # Act
        result1, warnings1 = processor._substitute_variables(content)
        result2, warnings2 = processor._substitute_variables(content)

        # Assert (both should work and produce same result)
        assert result1 == result2 == "Test value"
        assert len(processor._substitution_cache) > 0

    def test_substitute_variables_unknown_variable(self):
        """Test substitution with unknown variable generates warning."""
        # Arrange
        config = TemplateProcessorConfig(enable_substitution_warnings=True)
        processor = TemplateProcessor(Path("/test"), config=config)
        processor.set_context({"KNOWN": "value"})
        content = "{{UNKNOWN}}"

        # Act
        result, warnings = processor._substitute_variables(content)

        # Assert
        assert "{{UNKNOWN}}" in result  # Not substituted
        assert len(warnings) > 0

    def test_substitute_variables_sanitizes_values(self):
        """Test that values are sanitized to prevent recursive substitution."""
        # Arrange
        processor = TemplateProcessor(Path("/test"))
        processor.set_context({"VAR": "{{RECURSIVE}}"})
        content = "Test {{VAR}}"

        # Act
        result, warnings = processor._substitute_variables(content)

        # Assert (braces should be removed)
        assert "{{RECURSIVE}}" not in result

    def test_sanitize_value_removes_control_chars(self):
        """Test that value sanitization removes control characters."""
        # Arrange
        processor = TemplateProcessor(Path("/test"))

        # Act
        sanitized = processor._sanitize_value("hello\x00world")

        # Assert
        assert "\x00" not in sanitized
        assert "helloworld" == sanitized

    def test_is_text_file_recognition(self):
        """Test text file detection."""
        # Arrange
        processor = TemplateProcessor(Path("/test"))

        # Act & Assert
        assert processor._is_text_file(Path("file.md")) is True
        assert processor._is_text_file(Path("file.json")) is True
        assert processor._is_text_file(Path("file.py")) is True
        assert processor._is_text_file(Path("file.bin")) is False
        assert processor._is_text_file(Path("file.exe")) is False


class TestTemplateProcessorVersionFormatting:
    """Test version formatting methods."""

    def test_format_short_version_removes_v_prefix(self):
        """Test that short version format removes 'v' prefix."""
        # Arrange
        processor = TemplateProcessor(Path("/test"))

        # Act & Assert
        assert processor._format_short_version("v1.0.0") == "1.0.0"
        assert processor._format_short_version("1.0.0") == "1.0.0"

    def test_format_display_version_with_v_prefix(self):
        """Test display version formatting."""
        # Arrange
        processor = TemplateProcessor(Path("/test"))

        # Act
        result = processor._format_display_version("v1.0.0")

        # Assert
        assert result == "MoAI-ADK v1.0.0"

    def test_format_display_version_without_v_prefix(self):
        """Test display version adds v prefix if missing."""
        # Arrange
        processor = TemplateProcessor(Path("/test"))

        # Act
        result = processor._format_display_version("1.0.0")

        # Assert
        assert result == "MoAI-ADK v1.0.0"

    def test_format_display_version_unknown(self):
        """Test display version for unknown version."""
        # Arrange
        processor = TemplateProcessor(Path("/test"))

        # Act
        result = processor._format_display_version("unknown")

        # Assert
        assert result == "MoAI-ADK unknown version"

    def test_format_semver_version(self):
        """Test semantic version formatting."""
        # Arrange
        processor = TemplateProcessor(Path("/test"))

        # Act & Assert
        assert processor._format_semver_version("v1.2.3") == "1.2.3"
        assert processor._format_semver_version("1.2.3-alpha") == "1.2.3"
        assert processor._format_semver_version("unknown") == "0.0.0"

    def test_format_trimmed_version(self):
        """Test version trimming to max length."""
        # Arrange
        processor = TemplateProcessor(Path("/test"))

        # Act & Assert
        assert processor._format_trimmed_version("v1.2.3", max_length=10) == "1.2.3"
        assert processor._format_trimmed_version("v1.2.3-verylongprerelease", max_length=10) == "1.2.3-very"

    def test_is_valid_version_format(self):
        """Test version format validation."""
        # Arrange
        processor = TemplateProcessor(Path("/test"))

        # Act & Assert
        assert processor._is_valid_version_format("v1.0.0") is True
        assert processor._is_valid_version_format("1.0.0") is True
        assert processor._is_valid_version_format("invalid") is False


class TestTemplateProcessorCache:
    """Test caching functionality."""

    def test_clear_substitution_cache(self):
        """Test clearing substitution cache."""
        # Arrange
        processor = TemplateProcessor(Path("/test"))
        processor._substitution_cache[1] = ("test", [])

        # Act
        processor.clear_substitution_cache()

        # Assert
        assert len(processor._substitution_cache) == 0

    def test_get_cache_stats(self):
        """Test getting cache statistics."""
        # Arrange
        config = TemplateProcessorConfig(cache_size=100, enable_caching=True)
        processor = TemplateProcessor(Path("/test"), config=config)

        # Act
        stats = processor.get_cache_stats()

        # Assert
        assert "cache_size" in stats
        assert "max_cache_size" in stats
        assert "cache_enabled" in stats
        assert stats["max_cache_size"] == 100


class TestTemplateProcessorVariableValidation:
    """Test variable validation."""

    def test_is_valid_template_variable_valid(self):
        """Test valid template variable."""
        # Arrange
        processor = TemplateProcessor(Path("/test"))

        # Act
        result = processor._is_valid_template_variable("PROJECT_DIR", "/home/user")

        # Assert
        assert result is True

    def test_is_valid_template_variable_invalid_name(self):
        """Test invalid variable name format."""
        # Arrange
        processor = TemplateProcessor(Path("/test"))

        # Act
        result = processor._is_valid_template_variable("invalid-name", "value")

        # Assert
        assert result is False

    def test_is_valid_template_variable_empty_value(self):
        """Test validation rejects empty values."""
        # Arrange
        processor = TemplateProcessor(Path("/test"))

        # Act
        result = processor._is_valid_template_variable("VAR", "   ")

        # Assert
        assert result is False

    def test_is_valid_template_variable_too_long_value(self):
        """Test validation rejects very long values."""
        # Arrange
        processor = TemplateProcessor(Path("/test"))
        long_value = "x" * 1000

        # Act
        result = processor._is_valid_template_variable("VAR", long_value)

        # Assert
        assert result is False


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
