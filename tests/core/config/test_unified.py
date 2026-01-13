"""Tests for core.config.unified module."""

import json
import tempfile
import threading
import time
from pathlib import Path

from moai_adk.core.config.unified import (
    UnifiedConfigManager,
    get_unified_config,
)


class TestDeepMerge:
    """Test _deep_merge static method."""

    def test_deep_merge_simple_dicts(self):
        """Test merging two simple dictionaries."""
        base = {"a": 1, "b": 2}
        updates = {"c": 3, "d": 4}
        result = UnifiedConfigManager._deep_merge(base, updates)
        assert result == {"a": 1, "b": 2, "c": 3, "d": 4}

    def test_deep_merge_nested_dicts(self):
        """Test merging nested dictionaries."""
        base = {"config": {"timeout": 1000, "retries": 3}}
        updates = {"config": {"timeout": 2000}}
        result = UnifiedConfigManager._deep_merge(base, updates)
        assert result == {"config": {"timeout": 2000, "retries": 3}}

    def test_deep_merge_override_values(self):
        """Test that updates override base values."""
        base = {"a": 1, "b": 2, "c": 3}
        updates = {"a": 10, "b": 20}
        result = UnifiedConfigManager._deep_merge(base, updates)
        assert result == {"a": 10, "b": 20, "c": 3}

    def test_deep_merge_preserve_base(self):
        """Test that base dict is not modified."""
        base = {"a": 1}
        updates = {"b": 2}
        original_base = base.copy()
        UnifiedConfigManager._deep_merge(base, updates)
        assert base == original_base

    def test_deep_merge_empty_updates(self):
        """Test merging with empty updates."""
        base = {"a": 1, "b": 2}
        updates = {}
        result = UnifiedConfigManager._deep_merge(base, updates)
        assert result == {"a": 1, "b": 2}

    def test_deep_merge_empty_base(self):
        """Test merging with empty base."""
        base = {}
        updates = {"a": 1, "b": 2}
        result = UnifiedConfigManager._deep_merge(base, updates)
        assert result == {"a": 1, "b": 2}


class TestGetDefaultConfig:
    """Test _get_default_config static method."""

    def test_get_default_config_structure(self):
        """Test default config has expected structure."""
        config = UnifiedConfigManager._get_default_config()
        assert isinstance(config, dict)
        assert "moai" in config
        assert "project" in config
        assert "hooks" in config
        assert "session" in config
        assert "language" in config

    def test_get_default_config_values(self):
        """Test default config values."""
        config = UnifiedConfigManager._get_default_config()
        assert config["moai"]["version"] == "0.28.0"
        assert config["hooks"]["timeout_ms"] == 2000
        assert config["language"]["conversation_language"] == "en"


class TestGetConfig:
    """Test get method."""

    def setup_method(self):
        """Reset singleton before each test."""
        UnifiedConfigManager._instance = None
        UnifiedConfigManager._config = {}

    def test_get_simple_key(self):
        """Test getting a simple key."""
        with tempfile.TemporaryDirectory() as tmpdir:
            config_path = Path(tmpdir) / "config.json"
            config_path.write_text(json.dumps({"test_key": "test_value"}))

            config = UnifiedConfigManager(config_path)
            assert config.get("test_key") == "test_value"

    def test_get_nested_key_dot_notation(self):
        """Test getting nested key with dot notation."""
        with tempfile.TemporaryDirectory() as tmpdir:
            config_path = Path(tmpdir) / "config.json"
            config_data = {"parent": {"child": "nested_value"}}
            config_path.write_text(json.dumps(config_data))

            config = UnifiedConfigManager(config_path)
            assert config.get("parent.child") == "nested_value"

    def test_get_missing_key_with_default(self):
        """Test getting missing key returns default."""
        with tempfile.TemporaryDirectory() as tmpdir:
            config_path = Path(tmpdir) / "config.json"
            config_path.write_text(json.dumps({}))

            config = UnifiedConfigManager(config_path)
            assert config.get("missing_key", "default_value") == "default_value"

    def test_get_missing_key_no_default(self):
        """Test getting missing key without default returns None."""
        with tempfile.TemporaryDirectory() as tmpdir:
            config_path = Path(tmpdir) / "config.json"
            config_path.write_text(json.dumps({}))

            config = UnifiedConfigManager(config_path)
            assert config.get("missing_key") is None

    def test_get_partially_missing_nested_key(self):
        """Test getting partially missing nested key."""
        with tempfile.TemporaryDirectory() as tmpdir:
            config_path = Path(tmpdir) / "config.json"
            config_path.write_text(json.dumps({"parent": {}}))

            config = UnifiedConfigManager(config_path)
            assert config.get("parent.child.missing") is None

    def test_get_non_dict_in_path(self):
        """Test getting key when intermediate value is not a dict."""
        with tempfile.TemporaryDirectory() as tmpdir:
            config_path = Path(tmpdir) / "config.json"
            config_path.write_text(json.dumps({"parent": "not_a_dict"}))

            config = UnifiedConfigManager(config_path)
            assert config.get("parent.child", "default") == "default"


class TestSetConfig:
    """Test set method."""

    def setup_method(self):
        """Reset singleton before each test."""
        UnifiedConfigManager._instance = None
        UnifiedConfigManager._config = {}

    def test_set_simple_key(self):
        """Test setting a simple key."""
        with tempfile.TemporaryDirectory() as tmpdir:
            config_path = Path(tmpdir) / "config.json"
            config_path.write_text(json.dumps({}))

            config = UnifiedConfigManager(config_path)
            config.set("new_key", "new_value")
            assert config.get("new_key") == "new_value"

    def test_set_nested_key_dot_notation(self):
        """Test setting nested key with dot notation."""
        with tempfile.TemporaryDirectory() as tmpdir:
            config_path = Path(tmpdir) / "config.json"
            config_path.write_text(json.dumps({}))

            config = UnifiedConfigManager(config_path)
            config.set("parent.child", "nested_value")
            assert config.get("parent.child") == "nested_value"

    def test_set_update_existing_key(self):
        """Test updating existing key."""
        with tempfile.TemporaryDirectory() as tmpdir:
            config_path = Path(tmpdir) / "config.json"
            config_path.write_text(json.dumps({"existing": "old"}))

            config = UnifiedConfigManager(config_path)
            config.set("existing", "new")
            assert config.get("existing") == "new"

    def test_set_creates_intermediate_dicts(self):
        """Test that set creates intermediate dicts."""
        with tempfile.TemporaryDirectory() as tmpdir:
            config_path = Path(tmpdir) / "config.json"
            config_path.write_text(json.dumps({}))

            config = UnifiedConfigManager(config_path)
            config.set("a.b.c.d", "value")
            assert config.get("a.b.c.d") == "value"


class TestUpdateConfig:
    """Test update method."""

    def setup_method(self):
        """Reset singleton before each test."""
        UnifiedConfigManager._instance = None
        UnifiedConfigManager._config = {}

    def test_update_deep_merge_true(self):
        """Test update with deep merge enabled."""
        with tempfile.TemporaryDirectory() as tmpdir:
            config_path = Path(tmpdir) / "config.json"
            config_data = {"section": {"key1": "value1", "key2": "value2"}}
            config_path.write_text(json.dumps(config_data))

            config = UnifiedConfigManager(config_path)
            config.update({"section": {"key1": "new_value1"}})
            assert config.get("section.key1") == "new_value1"
            assert config.get("section.key2") == "value2"

    def test_update_deep_merge_false(self):
        """Test update with deep merge disabled."""
        with tempfile.TemporaryDirectory() as tmpdir:
            config_path = Path(tmpdir) / "config.json"
            config_data = {"section": {"key1": "value1", "key2": "value2"}}
            config_path.write_text(json.dumps(config_data))

            config = UnifiedConfigManager(config_path)
            config.update({"section": {"key3": "value3"}}, deep_merge=False)
            # Deep merge false replaces entire section
            assert config.get("section.key1") is None
            assert config.get("section.key3") == "value3"


class TestGetAllConfig:
    """Test get_all method."""

    def setup_method(self):
        """Reset singleton before each test."""
        UnifiedConfigManager._instance = None
        UnifiedConfigManager._config = {}

    def test_get_all_returns_copy(self):
        """Test that get_all returns a copy, not reference."""
        with tempfile.TemporaryDirectory() as tmpdir:
            config_path = Path(tmpdir) / "config.json"
            config_path.write_text(json.dumps({"key": "value"}))

            config = UnifiedConfigManager(config_path)
            all_config = config.get_all()
            all_config["key"] = "modified"
            # Original should be unchanged
            assert config.get("key") == "value"


class TestResetToDefaults:
    """Test reset_to_defaults method."""

    def setup_method(self):
        """Reset singleton before each test."""
        UnifiedConfigManager._instance = None
        UnifiedConfigManager._config = {}

    def test_reset_to_defaults(self):
        """Test resetting to default configuration."""
        with tempfile.TemporaryDirectory() as tmpdir:
            config_path = Path(tmpdir) / "config.json"
            config_path.write_text(json.dumps({"custom": "value"}))

            config = UnifiedConfigManager(config_path)
            config.reset_to_defaults()
            assert config.get("moai") is not None
            assert config.get("custom") is None


class TestLoadFromMonolithic:
    """Test _load_from_monolithic method."""

    def setup_method(self):
        """Reset singleton before each test."""
        UnifiedConfigManager._instance = None
        UnifiedConfigManager._config = {}
        UnifiedConfigManager._last_modified = None

    def test_load_json_file(self):
        """Test loading from JSON file."""
        with tempfile.TemporaryDirectory() as tmpdir:
            config_path = Path(tmpdir) / "config.json"
            config_data = {"test": "value", "number": 42}
            config_path.write_text(json.dumps(config_data))

            config = UnifiedConfigManager(config_path)
            assert config.get("test") == "value"
            assert config.get("number") == 42

    def test_load_invalid_json(self):
        """Test loading invalid JSON falls back to defaults."""
        with tempfile.TemporaryDirectory() as tmpdir:
            config_path = Path(tmpdir) / "config.json"
            config_path.write_text("invalid json {")

            config = UnifiedConfigManager(config_path)
            # Should have defaults
            assert config.get("moai") is not None

    def test_load_nonexistent_file(self):
        """Test loading non-existent file uses defaults."""
        with tempfile.TemporaryDirectory() as tmpdir:
            config_path = Path(tmpdir) / "nonexistent.json"

            config = UnifiedConfigManager(config_path)
            # Should have defaults
            assert config.get("moai") is not None


class TestSaveMonolithic:
    """Test _save_to_monolithic method."""

    def setup_method(self):
        """Reset singleton before each test."""
        UnifiedConfigManager._instance = None
        UnifiedConfigManager._config = {}
        UnifiedConfigManager._last_modified = None

    def test_save_json_file(self):
        """Test saving to JSON file."""
        with tempfile.TemporaryDirectory() as tmpdir:
            config_path = Path(tmpdir) / "config.json"
            config_path.write_text(json.dumps({}))

            config = UnifiedConfigManager(config_path)
            config.set("new_key", "new_value")
            result = config.save(backup=False)

            assert result is True
            # Verify file was saved
            saved_data = json.loads(config_path.read_text())
            assert saved_data["new_key"] == "new_value"

    def test_save_creates_backup(self):
        """Test that save creates backup when requested."""
        with tempfile.TemporaryDirectory() as tmpdir:
            config_path = Path(tmpdir) / "config.json"
            config_path.write_text(json.dumps({"original": "value"}))

            config = UnifiedConfigManager(config_path)
            config.set("updated", "value")
            config.save(backup=True)

            # Check backup exists
            backup_dir = Path(tmpdir) / "backups"
            assert backup_dir.exists()
            backups = list(backup_dir.glob("config_backup_*.json"))
            assert len(backups) >= 1


class TestModuleLevelSingleton:
    """Test module-level get_unified_config function."""

    def setup_method(self):
        """Reset singleton before each test."""
        UnifiedConfigManager._instance = None
        UnifiedConfigManager._config = {}

    def test_get_unified_config_singleton(self):
        """Test that get_unified_config returns singleton."""
        with tempfile.TemporaryDirectory() as tmpdir:
            config_path = Path(tmpdir) / "config.json"
            config_path.write_text(json.dumps({}))

            config1 = get_unified_config(config_path)
            config2 = get_unified_config(config_path)
            assert config1 is config2


class TestThreadSafety:
    """Test thread safety of singleton pattern."""

    def setup_method(self):
        """Reset singleton before each test."""
        UnifiedConfigManager._instance = None
        UnifiedConfigManager._config = {}

    def test_concurrent_singleton_creation(self):
        """Test that concurrent access creates only one instance."""
        with tempfile.TemporaryDirectory() as tmpdir:
            config_path = Path(tmpdir) / "config.json"
            config_path.write_text(json.dumps({}))

            instances = []
            errors = []

            def create_instance():
                try:
                    instance = UnifiedConfigManager(config_path)
                    instances.append(instance)
                except Exception as e:
                    errors.append(e)

            threads = [threading.Thread(target=create_instance) for _ in range(10)]
            for t in threads:
                t.start()
            for t in threads:
                t.join()

            assert len(errors) == 0
            # All instances should be the same
            first = instances[0]
            for instance in instances[1:]:
                assert instance is first


class TestCacheInvalidation:
    """Test cache invalidation on file modification."""

    def setup_method(self):
        """Reset singleton before each test."""
        UnifiedConfigManager._instance = None
        UnifiedConfigManager._config = {}
        UnifiedConfigManager._last_modified = None

    def test_reload_if_modified(self):
        """Test that config reloads when file is modified."""
        with tempfile.TemporaryDirectory() as tmpdir:
            config_path = Path(tmpdir) / "config.json"
            config_path.write_text(json.dumps({"key": "value1"}))

            config = UnifiedConfigManager(config_path)
            assert config.get("key") == "value1"

            # Wait a bit to ensure different mtime
            time.sleep(0.01)

            # Modify file
            config_path.write_text(json.dumps({"key": "value2"}))

            # Get should return updated value
            assert config.get("key") == "value2"
