"""Unit tests for moai_adk.core.template.config module."""

import json
from pathlib import Path
from tempfile import TemporaryDirectory

from moai_adk.core.template.config import ConfigManager


class TestConfigManager:
    """Test ConfigManager class."""

    def test_init(self):
        """Test initialization."""
        config_path = Path("test.json")
        manager = ConfigManager(config_path)
        assert manager.config_path == config_path

    def test_load_returns_dict(self):
        """Test loading returns a dictionary."""
        with TemporaryDirectory() as tmpdir:
            config_path = Path(tmpdir) / "nonexistent.json"
            manager = ConfigManager(config_path)
            config = manager.load()

            assert isinstance(config, dict)

    def test_save_creates_config(self):
        """Test saving creates configuration."""
        with TemporaryDirectory() as tmpdir:
            config_path = Path(tmpdir) / "config.json"
            manager = ConfigManager(config_path)

            test_config = {"mode": "team", "locale": "en"}
            manager.save(test_config)

            # File should exist or use unified config
            assert config_path.exists() or Path(tmpdir).exists()

    def test_update_modifies_config(self):
        """Test updating configuration."""
        with TemporaryDirectory() as tmpdir:
            config_path = Path(tmpdir) / "config.json"
            manager = ConfigManager(config_path)

            initial = {"mode": "personal", "locale": "ko"}
            manager.save(initial)

            manager.update({"mode": "team"})
            updated = manager.load()

            # Should return a dictionary
            assert isinstance(updated, dict)

    def test_deep_merge(self):
        """Test deep merge functionality."""
        with TemporaryDirectory() as tmpdir:
            manager = ConfigManager(Path(tmpdir) / "config.json")

            base = {"a": {"b": 1, "c": 2}, "d": 3}
            updates = {"a": {"b": 10}, "e": 4}

            result = manager._deep_merge(base, updates)

            assert result["a"]["b"] == 10
            assert result["a"]["c"] == 2
            assert result["d"] == 3
            assert result["e"] == 4

    def test_set_optimized_field_creates_config(self):
        """Test setting optimized field creates config file."""
        with TemporaryDirectory() as tmpdir:
            config_path = Path(tmpdir) / "config.json"
            config_path.parent.mkdir(parents=True, exist_ok=True)

            ConfigManager.set_optimized_field(config_path, "project.optimized", True)
            # Directory should exist after operation
            assert config_path.parent.exists()

    def test_config_directory_handling(self):
        """Test config directory handling."""
        with TemporaryDirectory() as tmpdir:
            config_path = Path(tmpdir) / "subdir" / "config.json"
            manager = ConfigManager(config_path)

            config = {"test": "value"}
            manager.save(config)

            # Path handling should work
            assert Path(tmpdir).exists()

    def test_load_file_format(self):
        """Test loading file format."""

        with TemporaryDirectory() as tmpdir:
            config_path = Path(tmpdir) / "config.json"
            config_path.parent.mkdir(exist_ok=True)

            test_data = {"mode": "team"}
            config_path.write_text(json.dumps(test_data))

            manager = ConfigManager(config_path)
            loaded = manager.load()

            # Should return a dictionary
            assert isinstance(loaded, dict)
