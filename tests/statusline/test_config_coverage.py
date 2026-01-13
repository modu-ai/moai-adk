"""Additional coverage tests for StatuslineConfig.

Tests for error handling and edge cases not covered by existing tests.
"""

from unittest.mock import patch

from moai_adk.statusline.config import StatuslineConfig


class TestConfigLoadErrorHandling:
    """Test error handling in config loading."""

    def test_load_config_falls_back_to_defaults_on_exception(self, tmp_path):
        """Should fall back to defaults when config loading fails."""
        # Create a malformed YAML file
        config_file = tmp_path / ".moai" / "config" / "statusline-config.yaml"
        config_file.parent.mkdir(parents=True)
        config_file.write_text("invalid: yaml: content: [")

        with patch("pathlib.Path.cwd", return_value=tmp_path):
            # Reset singleton to force reload
            StatuslineConfig._instance = None
            config = StatuslineConfig()

            # Should have default config
            assert config.get("statusline.enabled") is True
            assert config.get("statusline.mode") == "extended"

    def test_parse_yaml_returns_empty_on_import_error(self, tmp_path):
        """Should return empty dict when yaml import fails."""
        # Create a config file
        config_file = tmp_path / ".moai" / "config" / "statusline-config.yaml"
        config_file.parent.mkdir(parents=True)
        config_file.write_text("statusline:\n  enabled: true")

        with patch("pathlib.Path.cwd", return_value=tmp_path):
            # Mock yaml import to fail
            with patch("builtins.__import__", side_effect=ImportError("No yaml")):
                # This will trigger the JSON fallback which should also work
                # since our test YAML is also valid JSON
                StatuslineConfig._instance = None
                config = StatuslineConfig()
                # Should have loaded via JSON fallback
                assert config.get("statusline.enabled") is True

    def test_parse_json_fallback_returns_empty_on_error(self, tmp_path):
        """Should return empty dict when JSON fallback fails."""
        config_file = tmp_path / ".moai" / "config" / "statusline-config.yaml"
        config_file.parent.mkdir(parents=True)
        # Write invalid YAML that can't be parsed as JSON either
        config_file.write_text("{ invalid yaml }")

        with patch("pathlib.Path.cwd", return_value=tmp_path):
            # Mock yaml import to trigger JSON fallback
            with patch("builtins.__import__", side_effect=ImportError("No yaml")):
                StatuslineConfig._instance = None
                config = StatuslineConfig()
                # Should fall back to defaults when both fail
                assert config.get("statusline.enabled") is True


class TestConfigFileNotFound:
    """Test behavior when config file is not found."""

    def test_find_config_file_returns_none_when_no_config(self, tmp_path):
        """Should return None when no config file exists."""
        with patch("pathlib.Path.cwd", return_value=tmp_path):
            with patch("pathlib.Path.home", return_value=tmp_path):
                result = StatuslineConfig._find_config_file()
                assert result is None

    def test_load_config_uses_defaults_when_file_not_found(self, tmp_path):
        """Should use defaults when config file doesn't exist."""
        # Ensure no config file exists
        with patch("pathlib.Path.cwd", return_value=tmp_path):
            with patch("pathlib.Path.home", return_value=tmp_path):
                StatuslineConfig._instance = None
                config = StatuslineConfig()

                # Should have default config
                assert config.get("statusline.enabled") is True
                assert config.get("statusline.mode") == "extended"


class TestConfigGetWithNoneValue:
    """Test get() method with None values."""

    def test_get_returns_default_when_value_is_none(self):
        """Should return default when nested value resolves to None."""
        config = StatuslineConfig()

        # This key exists but value is None (not found in defaults)
        result = config.get("statusline.nonexistent_key", "default")
        assert result == "default"


class TestConfigGetMissingCacheFields:
    """Test get_cache_config with missing fields."""

    def test_get_cache_config_uses_defaults_for_missing_fields(self):
        """Should use default values when cache fields are missing."""
        config = StatuslineConfig()

        # Override config to have missing fields
        config._config = {"statusline": {"cache": {}}}

        cache_config = config.get_cache_config()

        # Should have defaults
        assert cache_config.git_ttl_seconds == 5
        assert cache_config.metrics_ttl_seconds == 10
        assert cache_config.alfred_ttl_seconds == 1
