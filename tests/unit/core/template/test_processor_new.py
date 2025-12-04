"""
Comprehensive tests for TemplateProcessor module.

Tests cover:
- TemplateProcessor class initialization
- set_context method with variable validation
- Template variable substitution
- File copying with substitution
- Directory copying with substitution
- Configuration merging
- Version handling and formatting
"""

import json
import tempfile
from pathlib import Path
from unittest import mock
from unittest.mock import MagicMock, patch

import pytest

from moai_adk.core.template.processor import (
    TemplateProcessor,
    TemplateProcessorConfig,
)


class TestTemplateProcessorConfig:
    """Test suite for TemplateProcessorConfig class."""

    def test_default_initialization(self):
        """Test TemplateProcessorConfig default initialization."""
        # Arrange & Act
        config = TemplateProcessorConfig()

        # Assert
        assert config.version_cache_ttl_seconds == 120
        assert config.version_fallback == "unknown"
        assert config.enable_version_validation is True
        assert config.preserve_user_version is True
        assert config.validate_template_variables is True
        assert config.max_variable_length == 50

    def test_from_dict_creates_config(self):
        """Test from_dict creates config from dictionary."""
        # Arrange
        config_dict = {
            "version_cache_ttl_seconds": 60,
            "enable_version_validation": False,
        }

        # Act
        config = TemplateProcessorConfig.from_dict(config_dict)

        # Assert
        assert config.version_cache_ttl_seconds == 60
        assert config.enable_version_validation is False
        assert config.version_fallback == "unknown"  # Default value

    def test_from_dict_with_none(self):
        """Test from_dict handles None input."""
        # Arrange & Act
        config = TemplateProcessorConfig.from_dict(None)

        # Assert
        assert config.version_cache_ttl_seconds == 120

    def test_from_dict_preserves_defaults(self):
        """Test from_dict preserves default values for missing keys."""
        # Arrange
        config_dict = {"version_cache_ttl_seconds": 30}

        # Act
        config = TemplateProcessorConfig.from_dict(config_dict)

        # Assert
        assert config.version_cache_ttl_seconds == 30
        assert config.enable_caching is True  # Default


class TestTemplateProcessor:
    """Test suite for TemplateProcessor class."""

    def test_initialization(self):
        """Test TemplateProcessor initialization."""
        # Arrange
        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir)

            # Act
            processor = TemplateProcessor(path)

            # Assert
            assert processor.target_path == path.resolve()
            assert processor.template_root is not None
            assert processor.context == {}
            assert processor.config is not None

    def test_initialization_with_custom_config(self):
        """Test TemplateProcessor initialization with custom config."""
        # Arrange
        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir)
            config = TemplateProcessorConfig(version_cache_ttl_seconds=30)

            # Act
            processor = TemplateProcessor(path, config)

            # Assert
            assert processor.config.version_cache_ttl_seconds == 30

    def test_set_context_updates_context(self):
        """Test set_context updates processor context."""
        # Arrange
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))
            context = {"PROJECT_NAME": "test-project", "AUTHOR": "Test"}

            # Act
            processor.set_context(context)

            # Assert
            assert processor.context == context

    def test_set_context_clears_cache(self):
        """Test set_context clears substitution cache."""
        # Arrange
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))
            processor._substitution_cache[1] = ("test", [])
            context = {"PROJECT_NAME": "test"}

            # Act
            processor.set_context(context)

            # Assert
            assert len(processor._substitution_cache) == 0

    def test_set_context_adds_hook_project_dir_alias(self):
        """Test set_context adds HOOK_PROJECT_DIR alias for PROJECT_DIR."""
        # Arrange
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))
            context = {"PROJECT_DIR": "$CLAUDE_PROJECT_DIR"}

            # Act
            processor.set_context(context)

            # Assert
            assert processor.context["HOOK_PROJECT_DIR"] == "$CLAUDE_PROJECT_DIR"

    def test_substitute_variables_replaces_placeholders(self):
        """Test _substitute_variables replaces template placeholders."""
        # Arrange
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))
            processor.set_context({"PROJECT_NAME": "MyProject"})
            content = "Project: {{PROJECT_NAME}}"

            # Act
            result, warnings = processor._substitute_variables(content)

            # Assert
            assert "MyProject" in result
            assert "{{PROJECT_NAME}}" not in result

    def test_substitute_variables_warns_on_missing_variables(self):
        """Test _substitute_variables warns about unsubstituted variables."""
        # Arrange
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))
            processor.set_context({})
            content = "Project: {{PROJECT_NAME}}"

            # Act
            result, warnings = processor._substitute_variables(content)

            # Assert
            assert len(warnings) > 0
            assert any("PROJECT_NAME" in w for w in warnings)

    def test_substitute_variables_validates_before_substitution(self):
        """Test _substitute_variables validates variables before substitution."""
        # Arrange
        with tempfile.TemporaryDirectory() as temp_dir:
            config = TemplateProcessorConfig(validate_template_variables=True)
            processor = TemplateProcessor(Path(temp_dir), config)
            processor.set_context({"VALID_VAR": "value"})
            content = "{{VALID_VAR}}"

            # Act
            result, warnings = processor._substitute_variables(content)

            # Assert
            assert "value" in result

    def test_substitute_variables_caches_result(self):
        """Test _substitute_variables caches substitution results."""
        # Arrange
        with tempfile.TemporaryDirectory() as temp_dir:
            config = TemplateProcessorConfig(enable_caching=True)
            processor = TemplateProcessor(Path(temp_dir), config)
            processor.set_context({"VAR": "value"})
            content = "{{VAR}}"

            # Act
            result1, _ = processor._substitute_variables(content)
            result2, _ = processor._substitute_variables(content)

            # Assert
            assert result1 == result2
            assert len(processor._substitution_cache) > 0

    def test_format_short_version_removes_prefix(self):
        """Test _format_short_version removes 'v' prefix."""
        # Arrange
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))

            # Act & Assert
            assert processor._format_short_version("v1.2.3") == "1.2.3"
            assert processor._format_short_version("1.2.3") == "1.2.3"

    def test_format_display_version_adds_moai_prefix(self):
        """Test _format_display_version adds MoAI-ADK prefix."""
        # Arrange
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))

            # Act & Assert
            assert processor._format_display_version("1.2.3") == "MoAI-ADK v1.2.3"
            assert processor._format_display_version("v1.2.3") == "MoAI-ADK v1.2.3"
            assert processor._format_display_version("unknown") == "MoAI-ADK unknown version"

    def test_format_trimmed_version_respects_max_length(self):
        """Test _format_trimmed_version respects maximum length."""
        # Arrange
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))

            # Act
            result = processor._format_trimmed_version("1.2.3-beta-long", max_length=5)

            # Assert
            assert len(result) <= 5

    def test_format_semver_version_extracts_major_minor_patch(self):
        """Test _format_semver_version extracts major.minor.patch."""
        # Arrange
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))

            # Act & Assert
            assert processor._format_semver_version("1.2.3") == "1.2.3"
            assert processor._format_semver_version("v1.2.3-beta") == "1.2.3"
            assert processor._format_semver_version("unknown") == "0.0.0"

    def test_is_text_file_returns_true_for_text_extensions(self):
        """Test _is_text_file returns True for text file extensions."""
        # Arrange
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))

            # Act & Assert
            assert processor._is_text_file(Path("file.md")) is True
            assert processor._is_text_file(Path("file.json")) is True
            assert processor._is_text_file(Path("file.py")) is True
            assert processor._is_text_file(Path("file.ts")) is True

    def test_is_text_file_returns_false_for_binary_extensions(self):
        """Test _is_text_file returns False for binary extensions."""
        # Arrange
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))

            # Act & Assert
            assert processor._is_text_file(Path("file.png")) is False
            assert processor._is_text_file(Path("file.jpg")) is False
            assert processor._is_text_file(Path("file.zip")) is False

    def test_sanitize_value_removes_control_characters(self):
        """Test _sanitize_value removes control characters."""
        # Arrange
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))
            value = "test\x00value\x01test"

            # Act
            result = processor._sanitize_value(value)

            # Assert
            assert "\x00" not in result
            assert "\x01" not in result

    def test_sanitize_value_removes_placeholder_patterns(self):
        """Test _sanitize_value removes placeholder patterns."""
        # Arrange
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))
            value = "test {{VAR}} value"

            # Act
            result = processor._sanitize_value(value)

            # Assert
            assert "{{" not in result
            assert "}}" not in result

    def test_is_valid_template_variable_validates_name_format(self):
        """Test _is_valid_template_variable validates variable name format."""
        # Arrange
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))

            # Act & Assert
            assert processor._is_valid_template_variable("VALID_NAME", "value") is True
            assert processor._is_valid_template_variable("invalid-name", "value") is False

    def test_is_valid_template_variable_validates_length(self):
        """Test _is_valid_template_variable validates variable length."""
        # Arrange
        with tempfile.TemporaryDirectory() as temp_dir:
            config = TemplateProcessorConfig(max_variable_length=10)
            processor = TemplateProcessor(Path(temp_dir), config)

            # Act & Assert
            assert processor._is_valid_template_variable("SHORT", "value") is True
            assert processor._is_valid_template_variable("A" * 20, "value") is False

    def test_is_valid_template_variable_rejects_empty_values(self):
        """Test _is_valid_template_variable rejects empty values."""
        # Arrange
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))

            # Act & Assert
            assert processor._is_valid_template_variable("VAR", "") is False
            assert processor._is_valid_template_variable("VAR", "   ") is False

    def test_clear_substitution_cache_clears_cache(self):
        """Test clear_substitution_cache clears the cache."""
        # Arrange
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))
            processor._substitution_cache[1] = ("test", [])

            # Act
            processor.clear_substitution_cache()

            # Assert
            assert len(processor._substitution_cache) == 0

    def test_get_cache_stats_returns_dict(self):
        """Test get_cache_stats returns cache statistics."""
        # Arrange
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))

            # Act
            stats = processor.get_cache_stats()

            # Assert
            assert isinstance(stats, dict)
            assert "cache_size" in stats
            assert "max_cache_size" in stats
            assert "cache_enabled" in stats

    def test_get_enhanced_version_context_returns_dict(self):
        """Test get_enhanced_version_context returns version context dict."""
        # Arrange
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))

            # Act
            with patch.object(processor, "_get_version_reader") as mock_reader:
                mock_reader_instance = MagicMock()
                mock_reader.return_value = mock_reader_instance
                mock_reader_instance.get_version.return_value = "1.0.0"
                mock_reader_instance.get_cache_age_seconds.return_value = 10.0

                context = processor.get_enhanced_version_context()

            # Assert
            assert isinstance(context, dict)
            assert "MOAI_VERSION" in context
            assert "MOAI_VERSION_SHORT" in context
            assert "MOAI_VERSION_DISPLAY" in context

    def test_get_enhanced_version_context_handles_errors(self):
        """Test get_enhanced_version_context handles version read errors."""
        # Arrange
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))

            # Act
            with patch.object(processor, "_get_version_reader") as mock_reader:
                mock_reader_instance = MagicMock()
                mock_reader.return_value = mock_reader_instance
                mock_reader_instance.get_version.side_effect = Exception("Read error")

                context = processor.get_enhanced_version_context()

            # Assert
            assert "MOAI_VERSION" in context
            assert context["MOAI_VERSION"] == "unknown"

    def test_copy_file_with_substitution_creates_file(self):
        """Test _copy_file_with_substitution creates file with substitution."""
        # Arrange
        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir)
            src_file = path / "src.md"
            dst_file = path / "dst.md"
            src_file.write_text("# {{PROJECT_NAME}}")

            processor = TemplateProcessor(path)
            processor.set_context({"PROJECT_NAME": "TestProject"})

            # Act
            processor._copy_file_with_substitution(src_file, dst_file)

            # Assert
            assert dst_file.exists()
            content = dst_file.read_text()
            assert "TestProject" in content

    def test_copy_file_with_substitution_handles_binary_files(self):
        """Test _copy_file_with_substitution handles binary files."""
        # Arrange
        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir)
            src_file = path / "src.png"
            dst_file = path / "dst.png"
            src_file.write_bytes(b"\x89PNG\r\n\x1a\n")

            processor = TemplateProcessor(path)

            # Act
            processor._copy_file_with_substitution(src_file, dst_file)

            # Assert
            assert dst_file.exists()
            assert dst_file.read_bytes() == b"\x89PNG\r\n\x1a\n"

    def test_copy_file_with_substitution_sets_executable_for_shell_scripts(self):
        """Test _copy_file_with_substitution sets executable bit for shell scripts."""
        # Arrange
        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir)
            src_file = path / "script.sh"
            dst_file = path / "dest.sh"
            src_file.write_text("#!/bin/bash\necho 'test'")

            processor = TemplateProcessor(path)

            # Act
            processor._copy_file_with_substitution(src_file, dst_file)

            # Assert
            assert dst_file.exists()
            import stat

            mode = dst_file.stat().st_mode
            assert mode & stat.S_IXUSR

    def test_copy_dir_with_substitution_creates_directory(self):
        """Test _copy_dir_with_substitution creates directory structure."""
        # Arrange
        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir)
            src_dir = path / "src"
            dst_dir = path / "dst"
            src_dir.mkdir()
            (src_dir / "file.md").write_text("# {{PROJECT_NAME}}")

            processor = TemplateProcessor(path)
            processor.set_context({"PROJECT_NAME": "Test"})

            # Act
            processor._copy_dir_with_substitution(src_dir, dst_dir)

            # Assert
            assert (dst_dir / "file.md").exists()
            content = (dst_dir / "file.md").read_text()
            assert "Test" in content

    def test_validate_template_variables_checks_names(self):
        """Test _validate_template_variables checks variable name format."""
        # Arrange
        with tempfile.TemporaryDirectory() as temp_dir:
            config = TemplateProcessorConfig(validate_template_variables=True)
            processor = TemplateProcessor(Path(temp_dir), config)
            context = {"VALID_NAME": "value"}

            # Act & Assert - should not raise
            processor._validate_template_variables(context)

    def test_validate_template_variables_warns_on_invalid_names(self):
        """Test _validate_template_variables warns on invalid names."""
        # Arrange
        with tempfile.TemporaryDirectory() as temp_dir:
            config = TemplateProcessorConfig(validate_template_variables=True)
            processor = TemplateProcessor(Path(temp_dir), config)
            context = {"invalid-name": "value"}

            # Act & Assert - should not raise with graceful degradation
            processor._validate_template_variables(context)

    def test_protected_paths_exclude_user_content(self):
        """Test PROTECTED_PATHS includes user content directories."""
        # Arrange & Act
        protected = TemplateProcessor.PROTECTED_PATHS

        # Assert
        assert ".moai/specs/" in protected
        assert ".moai/reports/" in protected
        assert ".moai/project/" in protected

    def test_common_template_variables_defined(self):
        """Test COMMON_TEMPLATE_VARIABLES includes expected variables."""
        # Arrange & Act
        vars_dict = TemplateProcessor.COMMON_TEMPLATE_VARIABLES

        # Assert
        assert "PROJECT_DIR" in vars_dict
        assert "PROJECT_NAME" in vars_dict
        assert "MOAI_VERSION" in vars_dict
        assert "AUTHOR" in vars_dict
        assert "CONVERSATION_LANGUAGE" in vars_dict

    def test_get_template_root_returns_path(self):
        """Test _get_template_root returns template path."""
        # Arrange
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))

            # Act
            result = processor._get_template_root()

            # Assert
            assert isinstance(result, Path)
            assert "templates" in str(result)

    def test_is_valid_version_format_validates_format(self):
        """Test _is_valid_version_format validates version format."""
        # Arrange
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))

            # Act & Assert
            assert processor._is_valid_version_format("1.2.3") is True
            assert processor._is_valid_version_format("v1.2.3") is True
            assert processor._is_valid_version_format("1.2.3-beta") is True
            assert processor._is_valid_version_format("invalid") is False

    def test_has_existing_files_delegates_to_backup(self):
        """Test _has_existing_files delegates to backup."""
        # Arrange
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))

            # Act
            with patch.object(processor.backup, "has_existing_files", return_value=True):
                result = processor._has_existing_files()

            # Assert
            assert result is True

    def test_create_backup_delegates_to_backup(self):
        """Test create_backup delegates to backup."""
        # Arrange
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))

            # Act
            with patch.object(processor.backup, "create_backup", return_value=Path("backup")):
                result = processor.create_backup()

            # Assert
            assert result == Path("backup")
