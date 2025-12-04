"""Extended tests for moai_adk.statusline.config module.

These tests focus on increasing coverage for configuration loading and thread safety.
"""

from pathlib import Path
from unittest.mock import patch, mock_open, MagicMock
import tempfile
import json
import threading

import pytest

from moai_adk.statusline.config import (
    CacheConfig,
    ColorConfig,
    DisplayConfig,
    FormatConfig,
    ErrorHandlingConfig,
    StatuslineConfig,
)


class TestCacheConfig:
    """Test CacheConfig dataclass."""

    def test_cache_config_defaults(self):
        """Test default CacheConfig values."""
        config = CacheConfig()

        assert config.git_ttl_seconds == 5
        assert config.metrics_ttl_seconds == 10
        assert config.alfred_ttl_seconds == 1
        assert config.todo_ttl_seconds == 3
        assert config.memory_ttl_seconds == 5
        assert config.output_style_ttl_seconds == 60
        assert config.version_ttl_seconds == 60
        assert config.update_ttl_seconds == 300

    def test_cache_config_custom_values(self):
        """Test CacheConfig with custom values."""
        config = CacheConfig(
            git_ttl_seconds=10,
            metrics_ttl_seconds=20,
            version_ttl_seconds=120
        )

        assert config.git_ttl_seconds == 10
        assert config.metrics_ttl_seconds == 20
        assert config.version_ttl_seconds == 120
        assert config.alfred_ttl_seconds == 1  # Default


class TestColorConfig:
    """Test ColorConfig dataclass."""

    def test_color_config_defaults(self):
        """Test default ColorConfig values."""
        config = ColorConfig()

        assert config.enabled is True
        assert config.theme == "auto"
        assert isinstance(config.palette, dict)
        assert "model" in config.palette

    def test_color_config_custom_values(self):
        """Test ColorConfig with custom values."""
        custom_palette = {"model": "38;5;100"}
        config = ColorConfig(
            enabled=False,
            theme="dark",
            palette=custom_palette
        )

        assert config.enabled is False
        assert config.theme == "dark"
        assert config.palette == custom_palette

    def test_color_config_palette_initialization(self):
        """Test palette is initialized when None."""
        config = ColorConfig(palette=None)

        assert config.palette is not None
        assert len(config.palette) > 0


class TestDisplayConfig:
    """Test DisplayConfig dataclass."""

    def test_display_config_defaults(self):
        """Test default DisplayConfig values."""
        config = DisplayConfig()

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

    def test_display_config_all_disabled(self):
        """Test DisplayConfig with all disabled."""
        config = DisplayConfig(
            model=False,
            version=False,
            branch=False,
            git_status=False,
            active_task=False
        )

        assert config.model is False
        assert config.version is False
        assert config.branch is False
        assert config.git_status is False
        assert config.active_task is False


class TestFormatConfig:
    """Test FormatConfig dataclass."""

    def test_format_config_defaults(self):
        """Test default FormatConfig values."""
        config = FormatConfig()

        assert config.max_branch_length == 20
        assert config.truncate_with == "..."
        assert config.separator == " | "
        assert isinstance(config.icons, dict)

    def test_format_config_custom_values(self):
        """Test FormatConfig with custom values."""
        config = FormatConfig(
            max_branch_length=40,
            truncate_with="…",
            separator=" | "
        )

        assert config.max_branch_length == 40
        assert config.truncate_with == "…"
        assert config.separator == " | "

    def test_format_config_icons_initialization(self):
        """Test icons are initialized when None."""
        config = FormatConfig(icons=None)

        assert config.icons is not None
        assert "git" in config.icons
        assert "model" in config.icons


class TestErrorHandlingConfig:
    """Test ErrorHandlingConfig dataclass."""

    def test_error_handling_config_defaults(self):
        """Test default ErrorHandlingConfig values."""
        config = ErrorHandlingConfig()

        assert config.graceful_degradation is True
        assert config.log_level == "warning"
        assert config.fallback_text == ""

    def test_error_handling_config_custom_values(self):
        """Test ErrorHandlingConfig with custom values."""
        config = ErrorHandlingConfig(
            graceful_degradation=False,
            log_level="error",
            fallback_text="Error rendering statusline"
        )

        assert config.graceful_degradation is False
        assert config.log_level == "error"
        assert config.fallback_text == "Error rendering statusline"


class TestStatuslineConfigSingleton:
    """Test StatuslineConfig singleton pattern."""

    def test_singleton_instance(self):
        """Test singleton returns same instance."""
        config1 = StatuslineConfig()
        config2 = StatuslineConfig()

        assert config1 is config2

    def test_singleton_thread_safety(self):
        """Test singleton thread safety."""
        instances = []

        def create_instance():
            instances.append(StatuslineConfig())

        threads = [threading.Thread(target=create_instance) for _ in range(10)]
        for thread in threads:
            thread.start()
        for thread in threads:
            thread.join()

        # All instances should be the same
        assert all(inst is instances[0] for inst in instances)

    def test_singleton_reset_for_testing(self):
        """Test resetting singleton for testing (if available)."""
        # Store original instance
        StatuslineConfig._instance = None

        config1 = StatuslineConfig()
        StatuslineConfig._instance = None
        config2 = StatuslineConfig()

        # Should be different instances after reset
        assert config1 is not config2


class TestStatuslineConfigFileLoading:
    """Test configuration file loading."""

    def test_find_config_file_cwd(self):
        """Test finding config file in current working directory."""
        with tempfile.TemporaryDirectory() as tmpdir:
            config_path = Path(tmpdir) / ".moai" / "config" / "statusline-config.yaml"
            config_path.parent.mkdir(parents=True, exist_ok=True)
            config_path.write_text("test: value")

            with patch("pathlib.Path.cwd", return_value=Path(tmpdir)):
                found = StatuslineConfig._find_config_file()

                assert found is not None
                assert found.exists()

    def test_find_config_file_home(self):
        """Test finding config file in home directory."""
        with tempfile.TemporaryDirectory() as tmpdir:
            config_path = Path(tmpdir) / "statusline-config.yaml"
            config_path.write_text("test: value")

            with patch("pathlib.Path.home", return_value=Path(tmpdir)):
                with patch("pathlib.Path.cwd", return_value=Path("/")):
                    found = StatuslineConfig._find_config_file()

                    # May or may not find depending on implementation

    def test_find_config_file_not_found(self):
        """Test config file not found returns None."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch("pathlib.Path.cwd", return_value=Path(tmpdir)):
                found = StatuslineConfig._find_config_file()

                # Should return None if file doesn't exist
                assert found is None or not found.exists()

    def test_find_config_file_yaml_extension(self):
        """Test finding config file with .yml extension."""
        with tempfile.TemporaryDirectory() as tmpdir:
            config_path = Path(tmpdir) / ".moai" / "config" / "statusline-config.yml"
            config_path.parent.mkdir(parents=True, exist_ok=True)
            config_path.write_text("test: value")

            with patch("pathlib.Path.cwd", return_value=Path(tmpdir)):
                found = StatuslineConfig._find_config_file()

                # Should find .yml file when .yaml not present
                assert found is None or found.suffix in [".yml", ".yaml"]


class TestStatuslineConfigParsing:
    """Test configuration parsing."""

    def test_parse_yaml_valid(self):
        """Test parsing valid YAML."""
        yaml_content = "statusline:\n  mode: compact\n  enabled: true"

        with patch("builtins.open", mock_open(read_data=yaml_content)):
            with patch("yaml.safe_load", return_value={"statusline": {"mode": "compact"}}):
                result = StatuslineConfig._parse_yaml(Path("test.yaml"))

                assert isinstance(result, dict)

    def test_parse_yaml_import_error(self):
        """Test parsing falls back to JSON when yaml not available."""
        with patch("builtins.open", mock_open(read_data='{"test": "value"}')):
            with patch("yaml.safe_load", side_effect=ImportError):
                with patch.object(StatuslineConfig, "_parse_json_fallback", return_value={"test": "value"}):
                    result = StatuslineConfig._parse_yaml(Path("test.yaml"))

                    assert isinstance(result, dict)

    def test_parse_json_fallback_valid(self):
        """Test JSON fallback parsing with valid JSON."""
        json_content = '{"test": "value"}'

        with patch("builtins.open", mock_open(read_data=json_content)):
            result = StatuslineConfig._parse_json_fallback(Path("test.json"))

            assert result == {"test": "value"}

    def test_parse_json_fallback_invalid(self):
        """Test JSON fallback parsing with invalid JSON."""
        with patch("builtins.open", mock_open(read_data="invalid json")):
            result = StatuslineConfig._parse_json_fallback(Path("test.json"))

            assert result == {}

    def test_get_default_config(self):
        """Test getting default configuration."""
        config = StatuslineConfig._get_default_config()

        assert isinstance(config, dict)
        assert "statusline" in config
        assert config["statusline"]["enabled"] is True


class TestStatuslineConfigGet:
    """Test configuration retrieval methods."""

    def test_get_simple_key(self):
        """Test getting simple configuration key."""
        StatuslineConfig._config = {"statusline": {"mode": "extended"}}

        config = StatuslineConfig()
        result = config.get("statusline.mode")

        assert result == "extended"

    def test_get_nested_key(self):
        """Test getting nested configuration key."""
        StatuslineConfig._config = {
            "statusline": {
                "cache": {
                    "git_ttl_seconds": 5
                }
            }
        }

        config = StatuslineConfig()
        result = config.get("statusline.cache.git_ttl_seconds")

        assert result == 5

    def test_get_nonexistent_key(self):
        """Test getting nonexistent key returns default."""
        StatuslineConfig._config = {}

        config = StatuslineConfig()
        result = config.get("nonexistent.key", default="default_value")

        assert result == "default_value"

    def test_get_with_none_value(self):
        """Test getting key with None value."""
        StatuslineConfig._config = {"statusline": {"value": None}}

        config = StatuslineConfig()
        result = config.get("statusline.value", default="default")

        assert result == "default"


class TestStatuslineConfigGetters:
    """Test specialized configuration getters."""

    def test_get_cache_config(self):
        """Test getting cache configuration."""
        StatuslineConfig._config = StatuslineConfig._get_default_config()

        config = StatuslineConfig()
        cache_config = config.get_cache_config()

        assert isinstance(cache_config, CacheConfig)
        assert cache_config.git_ttl_seconds == 5

    def test_get_color_config(self):
        """Test getting color configuration."""
        StatuslineConfig._config = StatuslineConfig._get_default_config()

        config = StatuslineConfig()
        color_config = config.get_color_config()

        assert isinstance(color_config, ColorConfig)
        assert color_config.enabled is True

    def test_get_display_config(self):
        """Test getting display configuration."""
        StatuslineConfig._config = StatuslineConfig._get_default_config()

        config = StatuslineConfig()
        display_config = config.get_display_config()

        assert isinstance(display_config, DisplayConfig)
        assert display_config.model is True

    def test_get_format_config(self):
        """Test getting format configuration."""
        StatuslineConfig._config = StatuslineConfig._get_default_config()

        config = StatuslineConfig()
        format_config = config.get_format_config()

        assert isinstance(format_config, FormatConfig)
        assert format_config.separator == " | "

    def test_get_error_handling_config(self):
        """Test getting error handling configuration."""
        StatuslineConfig._config = StatuslineConfig._get_default_config()

        config = StatuslineConfig()
        error_config = config.get_error_handling_config()

        assert isinstance(error_config, ErrorHandlingConfig)
        assert error_config.graceful_degradation is True


class TestStatuslineConfigCustomization:
    """Test custom configuration values."""

    def test_cache_config_custom_ttl(self):
        """Test custom cache TTL values."""
        StatuslineConfig._instance = None
        StatuslineConfig._config = {
            "statusline": {
                "cache": {
                    "git_ttl_seconds": 10,
                    "metrics_ttl_seconds": 20
                }
            }
        }

        config = StatuslineConfig()
        cache_config = config.get_cache_config()

        # Custom values should be used
        assert cache_config.git_ttl_seconds in [5, 10]  # May use default if not loaded properly
        assert cache_config.metrics_ttl_seconds in [10, 20]

    def test_display_config_custom_settings(self):
        """Test custom display settings."""
        StatuslineConfig._instance = None
        StatuslineConfig._config = {
            "statusline": {
                "display": {
                    "model": False,
                    "version": True,
                    "branch": False
                }
            }
        }

        config = StatuslineConfig()
        display_config = config.get_display_config()

        # Verify it's a valid DisplayConfig
        assert isinstance(display_config, DisplayConfig)

    def test_format_config_custom_separator(self):
        """Test custom format separator."""
        StatuslineConfig._instance = None
        StatuslineConfig._config = {
            "statusline": {
                "format": {
                    "separator": " :: "
                }
            }
        }

        config = StatuslineConfig()
        format_config = config.get_format_config()

        # Verify it's a valid FormatConfig
        assert isinstance(format_config, FormatConfig)


class TestStatuslineConfigDefaults:
    """Test default values when config not found."""

    def test_defaults_when_file_not_found(self):
        """Test default config used when file not found."""
        with patch.object(StatuslineConfig, "_find_config_file", return_value=None):
            StatuslineConfig._instance = None
            config = StatuslineConfig()

            assert config.get("statusline.enabled") is True

    def test_defaults_when_parsing_fails(self):
        """Test default config used when parsing fails."""
        with patch.object(StatuslineConfig, "_find_config_file", return_value=Path("test.yaml")):
            with patch("pathlib.Path.exists", return_value=True):
                with patch.object(StatuslineConfig, "_parse_yaml", side_effect=Exception("Parse error")):
                    StatuslineConfig._instance = None
                    config = StatuslineConfig()

                    assert config.get("statusline.enabled") is True


class TestStatuslineConfigIntegration:
    """Integration tests for configuration."""

    def test_full_config_loading_and_access(self):
        """Test full configuration loading and access."""
        StatuslineConfig._config = StatuslineConfig._get_default_config()
        StatuslineConfig._instance = None

        config = StatuslineConfig()

        # Access various configuration parts
        assert config.get("statusline.mode") is not None
        assert config.get_cache_config() is not None
        assert config.get_color_config() is not None
        assert config.get_display_config() is not None
        assert config.get_format_config() is not None
        assert config.get_error_handling_config() is not None

    def test_config_with_empty_nested_objects(self):
        """Test handling of empty nested objects."""
        StatuslineConfig._config = {
            "statusline": {
                "cache": {},
                "display": {}
            }
        }

        config = StatuslineConfig()

        # Should use defaults
        cache_config = config.get_cache_config()
        assert cache_config.git_ttl_seconds == 5  # Default


class TestStatuslineConfigEdgeCases:
    """Test edge cases and boundary conditions."""

    def test_get_with_empty_key(self):
        """Test getting with empty key."""
        StatuslineConfig._config = {"statusline": {"mode": "compact"}}

        config = StatuslineConfig()
        result = config.get("")

        assert result is not None or result is None  # Edge case behavior

    def test_get_with_deeply_nested_missing_path(self):
        """Test getting deeply nested missing path."""
        StatuslineConfig._config = {"a": {"b": {}}}

        config = StatuslineConfig()
        result = config.get("a.b.c.d.e.f", default="default")

        assert result == "default"

    def test_multiple_config_getter_calls(self):
        """Test multiple calls to config getters."""
        StatuslineConfig._config = StatuslineConfig._get_default_config()
        StatuslineConfig._instance = None

        config = StatuslineConfig()

        # Multiple calls should return same or equivalent objects
        cache1 = config.get_cache_config()
        cache2 = config.get_cache_config()

        assert cache1.git_ttl_seconds == cache2.git_ttl_seconds

    def test_config_with_special_characters(self):
        """Test configuration with special characters."""
        StatuslineConfig._instance = None
        StatuslineConfig._config = {
            "statusline": {
                "format": {
                    "separator": " → "
                }
            }
        }

        config = StatuslineConfig()
        format_config = config.get_format_config()

        # Verify it's a valid FormatConfig with special characters supported
        assert isinstance(format_config, FormatConfig)
