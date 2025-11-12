"""Unit tests for template/config.py module

Tests for ConfigManager class.
"""

import json
from pathlib import Path

from moai_adk.core.template.config import ConfigManager


class TestConfigManagerInit:
    """Test ConfigManager initialization"""

    def test_init_with_path(self, tmp_project_dir: Path):
        """Should initialize with given path"""
        config_path = tmp_project_dir / ".moai" / "config.json"
        manager = ConfigManager(config_path)
        assert manager.config_path == config_path

    def test_default_config_exists(self):
        """DEFAULT_CONFIG should be defined"""
        assert hasattr(ConfigManager, "DEFAULT_CONFIG")
        assert isinstance(ConfigManager.DEFAULT_CONFIG, dict)

    def test_default_config_contains_required_fields(self):
        """DEFAULT_CONFIG should contain mode, locale, and moai"""
        config = ConfigManager.DEFAULT_CONFIG
        assert "mode" in config
        assert "locale" in config
        assert "moai" in config
        assert "version" in config["moai"]


class TestConfigManagerLoad:
    """Test load method"""

    def test_load_returns_default_when_file_not_exists(self, tmp_project_dir: Path):
        """Should return default config when file doesn't exist"""
        config_path = tmp_project_dir / ".moai" / "config.json"
        manager = ConfigManager(config_path)

        result = manager.load()

        assert result == ConfigManager.DEFAULT_CONFIG

    def test_load_reads_existing_config(self, tmp_project_dir: Path):
        """Should read existing config file"""
        config_path = tmp_project_dir / ".moai" / "config.json"
        config_path.parent.mkdir(parents=True, exist_ok=True)

        test_config = {"mode": "team", "locale": "en", "custom": "value"}
        config_path.write_text(json.dumps(test_config, ensure_ascii=False), encoding="utf-8")

        manager = ConfigManager(config_path)
        result = manager.load()

        assert result == test_config
        assert result["mode"] == "team"
        assert result["locale"] == "en"
        assert result["custom"] == "value"

    def test_load_preserves_korean_characters(self, tmp_project_dir: Path):
        """Should preserve Korean characters (UTF-8)"""
        config_path = tmp_project_dir / ".moai" / "config.json"
        config_path.parent.mkdir(parents=True, exist_ok=True)

        test_config = {"project": {"name": "테스트 프로젝트", "description": "한글 설명"}}
        config_path.write_text(json.dumps(test_config, ensure_ascii=False), encoding="utf-8")

        manager = ConfigManager(config_path)
        result = manager.load()

        assert result["project"]["name"] == "테스트 프로젝트"
        assert result["project"]["description"] == "한글 설명"


class TestConfigManagerSave:
    """Test save method"""

    def test_save_creates_directory_if_not_exists(self, tmp_project_dir: Path):
        """Should create parent directory if it doesn't exist"""
        config_path = tmp_project_dir / ".moai" / "nested" / "config.json"
        manager = ConfigManager(config_path)

        test_config = {"mode": "personal"}
        manager.save(test_config)

        assert config_path.exists()
        assert config_path.parent.exists()

    def test_save_writes_config(self, tmp_project_dir: Path):
        """Should write config to file"""
        config_path = tmp_project_dir / ".moai" / "config.json"
        manager = ConfigManager(config_path)

        test_config = {"mode": "team", "locale": "ja", "nested": {"key": "value"}}
        manager.save(test_config)

        # Read back and verify
        with open(config_path, encoding="utf-8") as f:
            saved_config = json.load(f)

        assert saved_config == test_config

    def test_save_preserves_korean_characters(self, tmp_project_dir: Path):
        """Should save Korean characters without escaping"""
        config_path = tmp_project_dir / ".moai" / "config.json"
        manager = ConfigManager(config_path)

        test_config = {"project": {"name": "모아이 프로젝트"}}
        manager.save(test_config)

        # Read raw file content to check encoding
        content = config_path.read_text(encoding="utf-8")
        assert "모아이 프로젝트" in content
        assert "\\u" not in content  # Should not have unicode escapes

    def test_save_formats_with_indent(self, tmp_project_dir: Path):
        """Should format JSON with 2-space indent"""
        config_path = tmp_project_dir / ".moai" / "config.json"
        manager = ConfigManager(config_path)

        test_config = {"a": {"b": {"c": "value"}}}
        manager.save(test_config)

        content = config_path.read_text()
        # Check for proper indentation
        assert "  " in content  # Should have 2-space indent


class TestConfigManagerUpdate:
    """Test update method"""

    def test_update_creates_file_if_not_exists(self, tmp_project_dir: Path):
        """Should create config file if it doesn't exist"""
        config_path = tmp_project_dir / ".moai" / "config.json"
        manager = ConfigManager(config_path)

        updates = {"mode": "team"}
        manager.update(updates)

        assert config_path.exists()

    def test_update_merges_with_existing_config(self, tmp_project_dir: Path):
        """Should merge updates with existing config"""
        config_path = tmp_project_dir / ".moai" / "config.json"
        manager = ConfigManager(config_path)

        # Save initial config
        initial = {"mode": "personal", "locale": "ko"}
        manager.save(initial)

        # Update
        updates = {"mode": "team", "new_field": "value"}
        manager.update(updates)

        # Load and verify
        result = manager.load()
        assert result["mode"] == "team"  # Updated
        assert result["locale"] == "ko"  # Preserved
        assert result["new_field"] == "value"  # Added

    def test_update_deep_merge_nested_dicts(self, tmp_project_dir: Path):
        """Should deep merge nested dictionaries"""
        config_path = tmp_project_dir / ".moai" / "config.json"
        manager = ConfigManager(config_path)

        # Initial config
        initial = {"project": {"name": "test", "version": "1.0"}, "mode": "personal"}
        manager.save(initial)

        # Update nested field
        updates = {"project": {"version": "2.0", "author": "Alice"}}
        manager.update(updates)

        # Load and verify
        result = manager.load()
        assert result["project"]["name"] == "test"  # Preserved
        assert result["project"]["version"] == "2.0"  # Updated
        assert result["project"]["author"] == "Alice"  # Added
        assert result["mode"] == "personal"  # Preserved


class TestConfigManagerDeepMerge:
    """Test _deep_merge method"""

    def test_deep_merge_simple_merge(self, tmp_project_dir: Path):
        """Should merge simple dictionaries"""
        config_path = tmp_project_dir / ".moai" / "config.json"
        manager = ConfigManager(config_path)

        base = {"a": 1, "b": 2}
        updates = {"b": 3, "c": 4}

        result = manager._deep_merge(base, updates)

        assert result == {"a": 1, "b": 3, "c": 4}

    def test_deep_merge_nested_dicts(self, tmp_project_dir: Path):
        """Should recursively merge nested dicts"""
        config_path = tmp_project_dir / ".moai" / "config.json"
        manager = ConfigManager(config_path)

        base = {"level1": {"level2": {"a": 1, "b": 2}}}
        updates = {"level1": {"level2": {"b": 3, "c": 4}}}

        result = manager._deep_merge(base, updates)

        assert result == {"level1": {"level2": {"a": 1, "b": 3, "c": 4}}}

    def test_deep_merge_overwrites_non_dict_values(self, tmp_project_dir: Path):
        """Should overwrite non-dict values"""
        config_path = tmp_project_dir / ".moai" / "config.json"
        manager = ConfigManager(config_path)

        base = {"key": [1, 2, 3]}
        updates = {"key": [4, 5]}

        result = manager._deep_merge(base, updates)

        assert result == {"key": [4, 5]}  # List is replaced, not merged

    def test_deep_merge_preserves_base(self, tmp_project_dir: Path):
        """Should not modify original base dictionary"""
        config_path = tmp_project_dir / ".moai" / "config.json"
        manager = ConfigManager(config_path)

        base = {"a": 1}
        updates = {"b": 2}

        result = manager._deep_merge(base, updates)

        assert base == {"a": 1}  # Original base unchanged
        assert result == {"a": 1, "b": 2}


class TestConfigManagerSetOptimized:
    """Test set_optimized static method"""

    def test_set_optimized_with_existing_config(self, tmp_project_dir: Path):
        """Should set optimized field in existing config"""
        config_path = tmp_project_dir / ".moai" / "config.json"
        config_path.parent.mkdir(parents=True, exist_ok=True)

        # Create initial config
        test_config = {"mode": "personal"}
        config_path.write_text(json.dumps(test_config), encoding="utf-8")

        # Set optimized to True
        ConfigManager.set_optimized(config_path, True)

        # Verify
        result = json.loads(config_path.read_text(encoding="utf-8"))
        assert result["project"]["optimized"] is True

    def test_set_optimized_creates_project_section(self, tmp_project_dir: Path):
        """Should create project section if not exists"""
        config_path = tmp_project_dir / ".moai" / "config.json"
        config_path.parent.mkdir(parents=True, exist_ok=True)

        test_config = {"mode": "personal"}
        config_path.write_text(json.dumps(test_config), encoding="utf-8")

        ConfigManager.set_optimized(config_path, True)

        result = json.loads(config_path.read_text(encoding="utf-8"))
        assert "project" in result
        assert "optimized" in result["project"]

    def test_set_optimized_preserves_other_fields(self, tmp_project_dir: Path):
        """Should preserve other fields when setting optimized"""
        config_path = tmp_project_dir / ".moai" / "config.json"
        config_path.parent.mkdir(parents=True, exist_ok=True)

        test_config = {"mode": "team", "locale": "en"}
        config_path.write_text(json.dumps(test_config), encoding="utf-8")

        ConfigManager.set_optimized(config_path, False)

        result = json.loads(config_path.read_text(encoding="utf-8"))
        assert result["mode"] == "team"
        assert result["locale"] == "en"
        assert result["project"]["optimized"] is False

    def test_set_optimized_with_nonexistent_path(self, tmp_project_dir: Path):
        """Should do nothing when config path doesn't exist"""
        config_path = tmp_project_dir / ".moai" / "nonexistent.json"

        # Should not raise exception
        ConfigManager.set_optimized(config_path, True)

        # File should not be created
        assert not config_path.exists()

    def test_set_optimized_with_invalid_json(self, tmp_project_dir: Path):
        """Should handle invalid JSON gracefully"""
        config_path = tmp_project_dir / ".moai" / "config.json"
        config_path.parent.mkdir(parents=True, exist_ok=True)

        # Write invalid JSON
        config_path.write_text("{invalid json}", encoding="utf-8")

        # Should not raise exception
        ConfigManager.set_optimized(config_path, True)

        # Config should remain unchanged
        assert config_path.read_text() == "{invalid json}"

    def test_set_optimized_adds_trailing_newline(self, tmp_project_dir: Path):
        """Should add trailing newline to config file"""
        config_path = tmp_project_dir / ".moai" / "config.json"
        config_path.parent.mkdir(parents=True, exist_ok=True)

        test_config = {"mode": "personal"}
        config_path.write_text(json.dumps(test_config), encoding="utf-8")

        ConfigManager.set_optimized(config_path, True)

        content = config_path.read_text(encoding="utf-8")
        assert content.endswith("\n")
