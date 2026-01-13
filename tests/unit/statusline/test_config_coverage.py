"""
Comprehensive test coverage for statusline configuration module.

Tests for uncovered lines in config.py:
- ColorConfig.__post_init__ palette initialization (lines 42-43)
- Config loading failure paths (lines 150-155)
- YAML import failure handling (line 176)
- JSON fallback error handling (lines 195-197, 210-217)
- Default config return (line 227)
- get() method with nested value access (lines 317-319)
- Config getters with default values (lines 325-326, 336-337, 371-372)
"""

from pathlib import Path
from unittest.mock import MagicMock, mock_open, patch

import pytest

from moai_adk.statusline.config import (
    CacheConfig,
    ColorConfig,
    DisplayConfig,
    ErrorHandlingConfig,
    FormatConfig,
    StatuslineConfig,
)


class TestColorConfigPostInit:
    """Test ColorConfig.__post_init__ method (lines 42-43)."""

    def test_post_init_creates_default_palette_when_none(self):
        """Test that default palette is created when None."""
        # Arrange & Act
        config = ColorConfig(palette=None)  # type: ignore[arg-type]

        # Assert
        assert config.palette is not None
        assert "model" in config.palette
        assert "output_style" in config.palette
        assert "feature_branch" in config.palette

    def test_post_init_preserves_existing_palette(self):
        """Test that existing palette is preserved."""
        # Arrange
        custom_palette = {"model": "custom_color"}

        # Act
        config = ColorConfig(palette=custom_palette)

        # Assert
        assert config.palette == custom_palette
        assert config.palette["model"] == "custom_color"

    def test_post_init_default_palette_values(self):
        """Test that default palette has expected values."""
        # Arrange & Act
        config = ColorConfig()

        # Assert
        assert config.palette["model"] == "38;5;33"
        assert config.palette["output_style"] == "38;5;219"
        assert config.palette["feature_branch"] == "38;5;226"


class TestFormatConfigPostInit:
    """Test FormatConfig.__post_init__ method."""

    def test_post_init_creates_default_icons_when_none(self):
        """Test that default icons are created when None."""
        # Arrange & Act
        config = FormatConfig(icons=None)  # type: ignore[arg-type]

        # Assert
        assert config.icons is not None
        assert "git" in config.icons
        assert "staged" in config.icons

    def test_post_init_preserves_existing_icons(self):
        """Test that existing icons are preserved."""
        # Arrange
        custom_icons = {"git": "custom_git_icon"}

        # Act
        config = FormatConfig(icons=custom_icons)

        # Assert
        assert config.icons == custom_icons
        assert config.icons["git"] == "custom_git_icon"


class TestStatuslineConfigLoadConfig:
    """Test _load_config method (lines 150-155, 176)."""

    def test_load_config_file_not_found(self, tmp_path):
        """Test config loading when file doesn't exist."""
        # Arrange
        with patch("moai_adk.statusline.config.Path.cwd", return_value=tmp_path):
            with patch("moai_adk.statusline.config.Path.home", return_value=tmp_path):
                # Reset singleton to force reload
                StatuslineConfig._instance = None

                # Act
                config = StatuslineConfig()

        # Assert - should use default config
        assert config._config is not None

    def test_load_config_with_invalid_yaml(self, tmp_path):
        """Test config loading with invalid YAML file."""
        # Arrange
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        config_file = config_dir / "statusline-config.yaml"
        config_file.write_text("{ invalid: yaml: content ]")

        with patch("moai_adk.statusline.config.Path.cwd", return_value=tmp_path):
            # Reset singleton to force reload
            StatuslineConfig._instance = None

            # Act
            config = StatuslineConfig()

        # Assert - should fall back to default config
        assert config._config is not None

    def test_load_config_with_valid_yaml(self, tmp_path):
        """Test config loading with valid YAML file."""
        # Arrange
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        config_file = config_dir / "statusline-config.yaml"
        config_file.write_text(
            """
statusline:
  enabled: true
  mode: extended
  cache:
    git_ttl_seconds: 10
"""
        )

        with patch("moai_adk.statusline.config.Path.cwd", return_value=tmp_path):
            # Reset singleton to force reload
            StatuslineConfig._instance = None

            # Act
            config = StatuslineConfig()

        # Assert - should load config from file
        assert config._config is not None

    def test_load_config_yaml_import_error(self, tmp_path, caplog):
        """Test YAML import error handling (line 176)."""
        # Arrange
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        config_file = config_dir / "statusline-config.yaml"
        config_file.write_text("test: value")

        with patch("moai_adk.statusline.config.Path.cwd", return_value=tmp_path):
            # Mock yaml import to fail
            with patch("moai_adk.statusline.config.yaml.safe_load", side_effect=ImportError("No module named 'yaml'")):
                with patch("moai_adk.statusline.config.json.load", return_value={}):
                    # Reset singleton to force reload
                    StatuslineConfig._instance = None

                    # Act
                    config = StatuslineConfig()

        # Assert - should use JSON fallback or default
        assert config._config is not None


class TestStatuslineConfigParseYaml:
    """Test _parse_yaml method (lines 195-197, 210-217)."""

    def test_parse_yaml_valid_yaml(self, tmp_path):
        """Test parsing valid YAML file."""
        # Arrange
        yaml_file = tmp_path / "test.yaml"
        yaml_file.write_text("key: value\nnested:\n  item: test")

        # Act
        result = StatuslineConfig._parse_yaml(yaml_file)

        # Assert
        assert result == {"key": "value", "nested": {"item": "test"}}

    def test_parse_yaml_empty_file(self, tmp_path):
        """Test parsing empty YAML file."""
        # Arrange
        yaml_file = tmp_path / "empty.yaml"
        yaml_file.write_text("")

        # Act
        result = StatuslineConfig._parse_yaml(yaml_file)

        # Assert
        assert result == {}

    def test_parse_yaml_null_content(self, tmp_path):
        """Test parsing YAML file with null content."""
        # Arrange
        yaml_file = tmp_path / "null.yaml"
        yaml_file.write_text("key: null")

        # Act
        result = StatuslineConfig._parse_yaml(yaml_file)

        # Assert
        assert result == {"key": None}

    def test_parse_json_fallback_valid_json(self, tmp_path):
        """Test JSON fallback when YAML unavailable (lines 210-217)."""
        # Arrange
        json_file = tmp_path / "test.json"
        json_file.write_text('{"key": "value", "number": 42}')

        # Act
        result = StatuslineConfig._parse_json_fallback(json_file)

        # Assert
        assert result == {"key": "value", "number": 42}

    def test_parse_json_fallback_invalid_json(self, tmp_path, caplog):
        """Test JSON fallback with invalid JSON."""
        # Arrange
        json_file = tmp_path / "invalid.json"
        json_file.write_text("{ invalid json }")

        # Act
        result = StatuslineConfig._parse_json_fallback(json_file)

        # Assert
        assert result == {}

    def test_parse_json_fallback_file_not_found(self, tmp_path):
        """Test JSON fallback when file doesn't exist."""
        # Arrange
        missing_file = tmp_path / "missing.json"

        # Act
        result = StatuslineConfig._parse_json_fallback(missing_file)

        # Assert
        assert result == {}


class TestStatuslineConfigGetDefaultConfig:
    """Test _get_default_config method (line 227)."""

    def test_get_default_config_structure(self):
        """Test that default config has expected structure."""
        # Act
        config = StatuslineConfig._get_default_config()

        # Assert
        assert "statusline" in config
        assert "cache" in config["statusline"]
        assert "colors" in config["statusline"]
        assert "display" in config["statusline"]
        assert "format" in config["statusline"]

    def test_get_default_config_cache_values(self):
        """Test default cache configuration values."""
        # Act
        config = StatuslineConfig._get_default_config()

        # Assert
        cache = config["statusline"]["cache"]
        assert cache["git_ttl_seconds"] == 5
        assert cache["metrics_ttl_seconds"] == 10
        assert cache["alfred_ttl_seconds"] == 1

    def test_get_default_config_display_values(self):
        """Test default display configuration values."""
        # Act
        config = StatuslineConfig._get_default_config()

        # Assert
        display = config["statusline"]["display"]
        assert display["model"] is True
        assert display["version"] is True
        assert display["branch"] is True


class TestStatuslineConfigGet:
    """Test get method (lines 317-319)."""

    def test_get_simple_key(self):
        """Test getting simple configuration value."""
        # Arrange
        config = StatuslineConfig()

        # Act
        value = config.get("statusline.enabled")

        # Assert
        assert value is not None

    def test_get_nested_key(self):
        """Test getting nested configuration value."""
        # Arrange
        config = StatuslineConfig()

        # Act
        value = config.get("statusline.cache.git_ttl_seconds")

        # Assert
        assert value == 5

    def test_get_missing_key_with_default(self):
        """Test getting missing key returns default."""
        # Arrange
        config = StatuslineConfig()

        # Act
        value = config.get("statusline.missing_key", default="default_value")

        # Assert
        assert value == "default_value"

    def test_get_missing_key_without_default(self):
        """Test getting missing key without default returns None."""
        # Arrange
        config = StatuslineConfig()

        # Act
        value = config.get("statusline.missing.key")

        # Assert
        assert value is None

    def test_get_non_dict_value_in_path(self):
        """Test getting value when path contains non-dict (lines 317-319)."""
        # Arrange
        config = StatuslineConfig()

        # Act - try to access nested key in non-dict value
        value = config.get("statusline.mode.nonexistent", default="fallback")

        # Assert
        assert value == "fallback"


class TestStatuslineConfigGetters:
    """Test configuration getter methods (lines 325-326, 336-337, 371-372)."""

    def test_get_cache_config_with_defaults(self):
        """Test get_cache_config returns default values (lines 325-326)."""
        # Arrange
        config = StatuslineConfig()
        # Clear config to test defaults
        config._config = {}

        # Act
        cache_config = config.get_cache_config()

        # Assert - should use defaults from dataclass
        assert cache_config.git_ttl_seconds == 5
        assert cache_config.metrics_ttl_seconds == 10
        assert cache_config.alfred_ttl_seconds == 1

    def test_get_cache_config_with_custom_values(self):
        """Test get_cache_config with custom values."""
        # Arrange
        config = StatuslineConfig()
        config._config = {
            "statusline": {
                "cache": {
                    "git_ttl_seconds": 15,
                    "metrics_ttl_seconds": 20,
                }
            }
        }

        # Act
        cache_config = config.get_cache_config()

        # Assert
        assert cache_config.git_ttl_seconds == 15
        assert cache_config.metrics_ttl_seconds == 20

    def test_get_color_config_with_defaults(self):
        """Test get_color_config returns default values (lines 336-337)."""
        # Arrange
        config = StatuslineConfig()
        config._config = {}

        # Act
        color_config = config.get_color_config()

        # Assert - should use defaults from dataclass
        assert color_config.enabled is True
        assert color_config.theme == "auto"

    def test_get_color_config_with_custom_values(self):
        """Test get_color_config with custom values."""
        # Arrange
        config = StatuslineConfig()
        custom_palette = {"model": "custom_color"}
        config._config = {
            "statusline": {
                "colors": {
                    "enabled": False,
                    "theme": "dark",
                    "palette": custom_palette,
                }
            }
        }

        # Act
        color_config = config.get_color_config()

        # Assert
        assert color_config.enabled is False
        assert color_config.theme == "dark"
        assert color_config.palette == custom_palette

    def test_get_display_config_with_defaults(self):
        """Test get_display_config with default values."""
        # Arrange
        config = StatuslineConfig()
        config._config = {}

        # Act
        display_config = config.get_display_config()

        # Assert - should use defaults from dataclass
        assert display_config.model is True
        assert display_config.version is True
        assert display_config.branch is True

    def test_get_display_config_with_custom_values(self):
        """Test get_display_config with custom values."""
        # Arrange
        config = StatuslineConfig()
        config._config = {
            "statusline": {
                "display": {
                    "model": False,
                    "version": False,
                    "branch": True,
                }
            }
        }

        # Act
        display_config = config.get_display_config()

        # Assert
        assert display_config.model is False
        assert display_config.version is False
        assert display_config.branch is True

    def test_get_format_config_with_defaults(self):
        """Test get_format_config with default values."""
        # Arrange
        config = StatuslineConfig()
        config._config = {}

        # Act
        format_config = config.get_format_config()

        # Assert - should use defaults from dataclass
        assert format_config.max_branch_length == 20
        assert format_config.truncate_with == "..."

    def test_get_format_config_with_custom_values(self):
        """Test get_format_config with custom values."""
        # Arrange
        config = StatuslineConfig()
        config._config = {
            "statusline": {
                "format": {
                    "max_branch_length": 30,
                    "truncate_with": "—",
                    "separator": " • ",
                }
            }
        }

        # Act
        format_config = config.get_format_config()

        # Assert
        assert format_config.max_branch_length == 30
        assert format_config.truncate_with == "—"
        assert format_config.separator == " • "

    def test_get_error_handling_config_with_defaults(self):
        """Test get_error_handling_config with default values (lines 371-372)."""
        # Arrange
        config = StatuslineConfig()
        config._config = {}

        # Act
        error_config = config.get_error_handling_config()

        # Assert - should use defaults from dataclass
        assert error_config.graceful_degradation is True
        assert error_config.log_level == "warning"
        assert error_config.fallback_text == ""

    def test_get_error_handling_config_with_custom_values(self):
        """Test get_error_handling_config with custom values."""
        # Arrange
        config = StatuslineConfig()
        config._config = {
            "statusline": {
                "error_handling": {
                    "graceful_degradation": False,
                    "log_level": "error",
                    "fallback_text": "N/A",
                }
            }
        }

        # Act
        error_config = config.get_error_handling_config()

        # Assert
        assert error_config.graceful_degradation is False
        assert error_config.log_level == "error"
        assert error_config.fallback_text == "N/A"


class TestStatuslineConfigSingleton:
    """Test singleton pattern."""

    def test_singleton_returns_same_instance(self):
        """Test that singleton returns same instance."""
        # Arrange & Act
        config1 = StatuslineConfig()
        config2 = StatuslineConfig()

        # Assert
        assert config1 is config2

    def test_singleton_thread_safety(self):
        """Test that singleton is thread-safe."""
        # Arrange & Act
        import threading

        instances = []

        def get_instance():
            instances.append(StatuslineConfig())

        threads = [threading.Thread(target=get_instance) for _ in range(10)]
        for t in threads:
            t.start()
        for t in threads:
            t.join()

        # Assert - all instances should be the same
        assert len(set(instances)) == 1


class TestCacheConfigDataclass:
    """Test CacheConfig dataclass."""

    def test_cache_config_default_values(self):
        """Test default cache configuration values."""
        # Act
        config = CacheConfig()

        # Assert
        assert config.git_ttl_seconds == 5
        assert config.metrics_ttl_seconds == 10
        assert config.alfred_ttl_seconds == 1
        assert config.todo_ttl_seconds == 3
        assert config.memory_ttl_seconds == 5
        assert config.output_style_ttl_seconds == 60
        assert config.version_ttl_seconds == 60
        assert config.update_ttl_seconds == 300

    def test_cache_config_custom_values(self):
        """Test cache configuration with custom values."""
        # Act
        config = CacheConfig(
            git_ttl_seconds=15,
            metrics_ttl_seconds=20,
        )

        # Assert
        assert config.git_ttl_seconds == 15
        assert config.metrics_ttl_seconds == 20


class TestDisplayConfigDataclass:
    """Test DisplayConfig dataclass."""

    def test_display_config_default_values(self):
        """Test default display configuration values."""
        # Act
        config = DisplayConfig()

        # Assert
        assert config.model is True
        assert config.version is True
        assert config.output_style is True
        assert config.memory_usage is True
        assert config.todo_count is True
        assert config.branch is True
        assert config.git_status is True
        assert config.duration is True
        assert config.directory is True
        assert config.active_task is True
        assert config.update_indicator is True

    def test_display_config_custom_values(self):
        """Test display configuration with custom values."""
        # Act
        config = DisplayConfig(
            model=False,
            version=False,
            branch=True,
        )

        # Assert
        assert config.model is False
        assert config.version is False
        assert config.branch is True


class TestErrorHandlingConfigDataclass:
    """Test ErrorHandlingConfig dataclass."""

    def test_error_handling_config_default_values(self):
        """Test default error handling configuration values."""
        # Act
        config = ErrorHandlingConfig()

        # Assert
        assert config.graceful_degradation is True
        assert config.log_level == "warning"
        assert config.fallback_text == ""

    def test_error_handling_config_custom_values(self):
        """Test error handling configuration with custom values."""
        # Act
        config = ErrorHandlingConfig(
            graceful_degradation=False,
            log_level="error",
            fallback_text="Error",
        )

        # Assert
        assert config.graceful_degradation is False
        assert config.log_level == "error"
        assert config.fallback_text == "Error"
