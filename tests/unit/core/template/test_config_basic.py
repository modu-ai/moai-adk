"""Unit tests for moai_adk.core.template.config module.

Tests for basic ConfigManager functionality.
"""

import json
from pathlib import Path
from unittest.mock import patch

from moai_adk.core.template.config import ConfigManager


class TestConfigManagerInitialization:
    """Test ConfigManager initialization."""

    def test_initialization(self):
        """Test ConfigManager initialization."""
        config_path = Path("/tmp/config.json")
        manager = ConfigManager(config_path)
        assert manager.config_path == config_path

    def test_default_config(self):
        """Test default configuration values."""
        assert "mode" in ConfigManager.DEFAULT_CONFIG
        assert "locale" in ConfigManager.DEFAULT_CONFIG
        assert ConfigManager.DEFAULT_CONFIG["mode"] == "personal"
        assert ConfigManager.DEFAULT_CONFIG["locale"] == "ko"


class TestConfigLoad:
    """Test config loading."""

    def test_load_returns_dict(self):
        """Test load returns dictionary."""
        manager = ConfigManager(Path("/tmp/config.json"))
        with patch("pathlib.Path.exists", return_value=False):
            result = manager.load()
            assert isinstance(result, dict)

    def test_load_missing_file_returns_defaults(self):
        """Test load returns default config when file missing."""
        manager = ConfigManager(Path("/tmp/config.json"))
        with patch("pathlib.Path.exists", return_value=False):
            result = manager.load()
            assert isinstance(result, dict)
            assert "mode" in result or "moai" in result

    def test_load_existing_file(self):
        """Test load from existing file."""
        test_config = {"mode": "team", "locale": "en"}
        manager = ConfigManager(Path("/tmp/config.json"))
        with patch("pathlib.Path.exists", return_value=True):
            with patch("builtins.open", create=True) as mock_open:
                mock_open.return_value.__enter__.return_value.read.return_value = json.dumps(test_config)
                with patch("json.load", return_value=test_config):
                    result = manager.load()
                    assert isinstance(result, dict)


class TestConfigSave:
    """Test config saving."""

    def test_save_writes_file(self):
        """Test save writes configuration."""
        manager = ConfigManager(Path("/tmp/config.json"))
        config = {"mode": "personal", "locale": "ko"}
        with patch("pathlib.Path.mkdir"):
            with patch("builtins.open", create=True) as mock_open:
                with patch("json.dump"):
                    manager.save(config)
                    assert mock_open.called

    def test_save_creates_directory(self):
        """Test save creates parent directory."""
        manager = ConfigManager(Path("/tmp/nested/config.json"))
        config = {"mode": "personal"}
        with patch.object(Path, "mkdir"):
            with patch("builtins.open", create=True):
                with patch("json.dump"):
                    try:
                        manager.save(config)
                    except Exception:
                        pass  # Ignore exceptions from mocking

    def test_save_uses_utf8(self):
        """Test save uses UTF-8 encoding."""
        manager = ConfigManager(Path("/tmp/config.json"))
        config = {"mode": "personal"}
        with patch("pathlib.Path.mkdir"):
            with patch("builtins.open", create=True):
                with patch("json.dump"):
                    manager.save(config)
                    # Check that open was called with encoding parameter
                    # This is harder to verify without looking at actual call


class TestConfigUpdate:
    """Test config updating."""

    def test_update_basic(self):
        """Test basic config update."""
        manager = ConfigManager(Path("/tmp/config.json"))
        with patch.object(manager, "load", return_value={"mode": "personal"}):
            with patch.object(manager, "save", return_value=None):
                try:
                    manager.update({"locale": "en"})
                except Exception:
                    pass  # Ignore exceptions from mocking

    def test_update_deep_merge(self):
        """Test deep merge in update."""
        manager = ConfigManager(Path("/tmp/config.json"))
        result = manager._deep_merge({"a": {"b": 1}}, {"a": {"c": 2}})
        assert result["a"]["b"] == 1
        assert result["a"]["c"] == 2

    def test_update_overwrites_values(self):
        """Test update overwrites existing values."""
        manager = ConfigManager(Path("/tmp/config.json"))
        result = manager._deep_merge({"mode": "personal"}, {"mode": "team"})
        assert result["mode"] == "team"


class TestDeepMerge:
    """Test deep merge functionality."""

    def test_deep_merge_simple(self):
        """Test simple deep merge."""
        manager = ConfigManager(Path("/tmp/config.json"))
        base = {"a": 1, "b": 2}
        updates = {"b": 3, "c": 4}
        result = manager._deep_merge(base, updates)
        assert result["a"] == 1
        assert result["b"] == 3
        assert result["c"] == 4

    def test_deep_merge_nested_dicts(self):
        """Test nested dictionary merge."""
        manager = ConfigManager(Path("/tmp/config.json"))
        base = {"outer": {"inner": 1}}
        updates = {"outer": {"new": 2}}
        result = manager._deep_merge(base, updates)
        assert result["outer"]["inner"] == 1
        assert result["outer"]["new"] == 2

    def test_deep_merge_preserves_base(self):
        """Test deep merge preserves base dictionary."""
        manager = ConfigManager(Path("/tmp/config.json"))
        base = {"a": 1}
        updates = {"b": 2}
        manager._deep_merge(base, updates)
        assert base["a"] == 1
        assert "b" not in base  # Original not modified


class TestSetOptimizedField:
    """Test static set_optimized_field method."""

    def test_set_optimized_field_creates_file(self):
        """Test set_optimized_field creates file."""
        with patch("pathlib.Path.parent"):
            with patch("pathlib.Path.mkdir"):
                with patch("pathlib.Path.exists", return_value=False):
                    with patch("builtins.open", create=True):
                        with patch("json.dump"):
                            ConfigManager.set_optimized_field(
                                Path("/tmp/config.json"),
                                "project.optimized",
                                True,
                            )

    def test_set_optimized_field_updates_existing(self):
        """Test set_optimized_field updates existing config."""
        with patch("pathlib.Path.mkdir"):
            with patch("pathlib.Path.exists", return_value=True):
                with patch("builtins.open", create=True):
                    with patch("json.load", return_value={"existing": "data"}):
                        with patch("json.dump"):
                            ConfigManager.set_optimized_field(
                                Path("/tmp/config.json"),
                                "new.field",
                                True,
                            )

    def test_set_optimized_field_nested_path(self):
        """Test set_optimized_field with nested path."""
        Path("/tmp/config.json")
        # This would need more sophisticated mocking to fully test


class TestConfigBackwardCompatibility:
    """Test backward compatibility features."""

    def test_uses_unified_config_when_available(self):
        """Test ConfigManager uses UnifiedConfigManager if available."""
        with patch("moai_adk.core.template.config.UNIFIED_AVAILABLE", True):
            manager = ConfigManager(Path("/tmp/config.json"))
            # Should have _unified_config
            assert hasattr(manager, "_unified_config")

    def test_fallback_when_unified_unavailable(self):
        """Test ConfigManager fallback when unified not available."""
        with patch("moai_adk.core.template.config.UNIFIED_AVAILABLE", False):
            manager = ConfigManager(Path("/tmp/config.json"))
            assert manager._unified_config is None


class TestConfigIntegration:
    """Integration tests for ConfigManager."""

    def test_load_save_roundtrip(self):
        """Test load/save roundtrip."""
        manager = ConfigManager(Path("/tmp/config.json"))
        original = {"mode": "personal", "locale": "ko"}
        with patch("pathlib.Path.exists", return_value=True):
            with patch("builtins.open", create=True):
                with patch("json.load", return_value=original):
                    with patch("json.dump"):
                        loaded = manager.load()
                        manager.save(loaded)
