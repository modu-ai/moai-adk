"""
Tests for StatuslineConfig - Configuration loading and management

"""

import pytest

from moai_adk.statusline.config import (
    CacheConfig,
    ColorConfig,
    DisplayConfig,
    ErrorHandlingConfig,
    FormatConfig,
    StatuslineConfig,
)


class TestStatuslineConfigDefaults:
    """Test default configuration loading"""

    def test_config_singleton_pattern(self):
        """
        GIVEN: StatuslineConfig class
        WHEN: Creating multiple instances
        THEN: Both instances should be the same singleton
        """
        config1 = StatuslineConfig()
        config2 = StatuslineConfig()
        assert config1 is config2, "StatuslineConfig should be a singleton"

    def test_default_config_loads(self):
        """
        GIVEN: StatuslineConfig with no YAML file
        WHEN: Creating instance
        THEN: Default configuration is loaded
        """
        config = StatuslineConfig()
        assert config.get("statusline.enabled") is True
        assert config.get("statusline.mode") == "extended"
        assert config.get("statusline.refresh_interval_ms") == 300

    def test_default_mode_is_extended(self):
        """
        GIVEN: Default configuration
        WHEN: Getting statusline mode
        THEN: Mode should be 'extended' (not 'compact')
        """
        config = StatuslineConfig()
        mode = config.get("statusline.mode")
        assert mode == "extended", f"Expected 'extended' but got '{mode}'"

    def test_default_separator_is_pipe(self):
        """
        GIVEN: Default configuration
        WHEN: Getting format separator
        THEN: Separator should be ' | '
        """
        config = StatuslineConfig()
        separator = config.get("statusline.format.separator")
        assert separator == " | ", f"Expected ' | ' but got '{separator}'"


class TestStatuslineConfigValueRetrieval:
    """Test configuration value retrieval"""

    def test_get_with_dot_notation(self):
        """
        GIVEN: StatuslineConfig instance
        WHEN: Getting value with dot notation
        THEN: Correct value is returned
        """
        config = StatuslineConfig()

        # Test various keys
        assert config.get("statusline.enabled") is True
        assert config.get("statusline.mode") == "extended"
        assert config.get("statusline.cache.git_ttl_seconds") == 5
        assert config.get("statusline.cache.update_ttl_seconds") == 300

    def test_get_with_default_value(self):
        """
        GIVEN: StatuslineConfig instance
        WHEN: Getting non-existent key with default
        THEN: Default value is returned
        """
        config = StatuslineConfig()
        result = config.get("non.existent.key", "default_value")
        assert result == "default_value"

    def test_get_nonexistent_key_returns_none(self):
        """
        GIVEN: StatuslineConfig instance
        WHEN: Getting non-existent key without default
        THEN: None is returned
        """
        config = StatuslineConfig()
        result = config.get("totally.fake.key")
        assert result is None

    def test_get_nested_palette(self):
        """
        GIVEN: StatuslineConfig instance
        WHEN: Getting color palette values
        THEN: Correct ANSI codes are returned
        """
        config = StatuslineConfig()

        palette = config.get("statusline.colors.palette")
        assert isinstance(palette, dict)
        assert palette.get("model") == "38;5;33"  # Blue
        assert palette.get("staged") == "38;5;46"  # Green
        assert palette.get("modified") == "38;5;208"  # Orange


class TestStatuslineConfigObjects:
    """Test configuration object generation"""

    def test_get_cache_config(self):
        """
        GIVEN: StatuslineConfig instance
        WHEN: Calling get_cache_config()
        THEN: CacheConfig object is returned with correct values
        """
        config = StatuslineConfig()
        cache_config = config.get_cache_config()

        assert isinstance(cache_config, CacheConfig)
        assert cache_config.git_ttl_seconds == 5
        assert cache_config.metrics_ttl_seconds == 10
        assert cache_config.alfred_ttl_seconds == 1
        assert cache_config.version_ttl_seconds == 60
        assert cache_config.update_ttl_seconds == 300

    def test_get_color_config(self):
        """
        GIVEN: StatuslineConfig instance
        WHEN: Calling get_color_config()
        THEN: ColorConfig object is returned with correct values
        """
        config = StatuslineConfig()
        color_config = config.get_color_config()

        assert isinstance(color_config, ColorConfig)
        assert color_config.enabled is True
        assert color_config.theme == "auto"
        assert "model" in color_config.palette
        assert color_config.palette["model"] == "38;5;33"

    def test_get_display_config(self):
        """
        GIVEN: StatuslineConfig instance
        WHEN: Calling get_display_config()
        THEN: DisplayConfig object is returned with all fields True
        """
        config = StatuslineConfig()
        display_config = config.get_display_config()

        assert isinstance(display_config, DisplayConfig)
        assert display_config.model is True
        assert display_config.duration is True
        assert display_config.directory is True
        assert display_config.version is True
        assert display_config.branch is True
        assert display_config.git_status is True
        assert display_config.active_task is True
        assert display_config.update_indicator is True

    def test_get_format_config(self):
        """
        GIVEN: StatuslineConfig instance
        WHEN: Calling get_format_config()
        THEN: FormatConfig object is returned with correct values
        """
        config = StatuslineConfig()
        format_config = config.get_format_config()

        assert isinstance(format_config, FormatConfig)
        assert format_config.max_branch_length == 20
        assert format_config.truncate_with == "..."
        assert format_config.separator == " | "

    def test_get_error_handling_config(self):
        """
        GIVEN: StatuslineConfig instance
        WHEN: Calling get_error_handling_config()
        THEN: ErrorHandlingConfig object is returned with correct values
        """
        config = StatuslineConfig()
        error_config = config.get_error_handling_config()

        assert isinstance(error_config, ErrorHandlingConfig)
        assert error_config.graceful_degradation is True
        assert error_config.log_level == "warning"
        assert error_config.fallback_text == ""


class TestCacheConfigDataclass:
    """Test CacheConfig dataclass"""

    def test_cache_config_defaults(self):
        """
        GIVEN: CacheConfig with default values
        WHEN: Creating instance
        THEN: All defaults are set correctly
        """
        config = CacheConfig()
        assert config.git_ttl_seconds == 5
        assert config.metrics_ttl_seconds == 10
        assert config.alfred_ttl_seconds == 1
        assert config.version_ttl_seconds == 60
        assert config.update_ttl_seconds == 300

    def test_cache_config_custom_values(self):
        """
        GIVEN: CacheConfig with custom values
        WHEN: Creating instance
        THEN: Custom values are preserved
        """
        config = CacheConfig(
            git_ttl_seconds=10,
            metrics_ttl_seconds=20,
            alfred_ttl_seconds=2,
        )
        assert config.git_ttl_seconds == 10
        assert config.metrics_ttl_seconds == 20
        assert config.alfred_ttl_seconds == 2


class TestColorConfigDataclass:
    """Test ColorConfig dataclass"""

    def test_color_config_defaults(self):
        """
        GIVEN: ColorConfig with default values
        WHEN: Creating instance
        THEN: All defaults are set correctly
        """
        config = ColorConfig()
        assert config.enabled is True
        assert config.theme == "auto"
        assert "model" in config.palette
        assert config.palette["model"] == "38;5;33"

    def test_color_config_custom_palette(self):
        """
        GIVEN: ColorConfig with custom palette
        WHEN: Creating instance
        THEN: Custom palette is used
        """
        custom_palette = {"model": "31"}  # Red
        config = ColorConfig(enabled=True, theme="light", palette=custom_palette)
        assert config.palette == custom_palette


class TestFormatConfigDataclass:
    """Test FormatConfig dataclass"""

    def test_format_config_defaults(self):
        """
        GIVEN: FormatConfig with default values
        WHEN: Creating instance
        THEN: All defaults are set correctly
        """
        config = FormatConfig()
        assert config.max_branch_length == 20
        assert config.truncate_with == "..."
        assert config.separator == " | "

    def test_format_config_custom_separator(self):
        """
        GIVEN: FormatConfig with custom separator
        WHEN: Creating instance
        THEN: Custom separator is used
        """
        config = FormatConfig(separator=" • ")
        assert config.separator == " • "
        assert config.max_branch_length == 20  # Default


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
