"""
@TEST:CORE-CONFIG-MANAGER-COVERAGE-001 Config Manager Test Coverage

Tests for core configuration management functionality to achieve 85% coverage target.
Focuses on configuration loading, validation, and management operations.
"""

import json
import tempfile
from pathlib import Path
from unittest.mock import Mock, patch

import pytest

from moai_adk.core.config_manager import ConfigManager
from moai_adk.config import Config


class TestConfigManager:
    """Test ConfigManager functionality."""

    def setup_method(self):
        """Set up test environment."""
        self.temp_dir = Path(tempfile.mkdtemp())
        self.config_path = self.temp_dir / "config.json"
        self.manager = ConfigManager(str(self.config_path))

    def teardown_method(self):
        """Clean up test environment."""
        if self.temp_dir.exists():
            import shutil
            shutil.rmtree(self.temp_dir)

    def test_should_initialize_with_config_path(self):
        """Test ConfigManager initialization with config path."""
        manager = ConfigManager(str(self.config_path))

        assert manager is not None
        assert manager.config_path == self.config_path

    def test_should_create_default_config(self):
        """Test default configuration creation."""
        config = self.manager.create_default_config()

        assert isinstance(config, dict)
        assert "version" in config
        assert "created_at" in config

    def test_should_save_config_to_file(self):
        """Test saving configuration to file."""
        config_data = {
            "version": "1.0.0",
            "name": "test-project"
        }

        result = self.manager.save_config(config_data)

        assert result is True
        assert self.config_path.exists()

        # Verify content
        loaded_data = json.loads(self.config_path.read_text())
        assert loaded_data["version"] == "1.0.0"
        assert loaded_data["name"] == "test-project"

    def test_should_load_config_from_file(self):
        """Test loading configuration from file."""
        # Create config file
        config_data = {"version": "1.0.0", "name": "test-project"}
        self.config_path.write_text(json.dumps(config_data))

        loaded_config = self.manager.load_config()

        assert loaded_config is not None
        assert loaded_config["version"] == "1.0.0"
        assert loaded_config["name"] == "test-project"

    def test_should_handle_missing_config_file(self):
        """Test handling of missing configuration file."""
        # Ensure config file doesn't exist
        if self.config_path.exists():
            self.config_path.unlink()

        loaded_config = self.manager.load_config()

        # Should return default config or None
        assert loaded_config is None or isinstance(loaded_config, dict)

    def test_should_validate_config_schema(self):
        """Test configuration schema validation."""
        valid_config = {
            "version": "1.0.0",
            "name": "test-project",
            "created_at": "2023-01-01T00:00:00Z"
        }

        result = self.manager.validate_config(valid_config)

        assert result is True

    def test_should_reject_invalid_config_schema(self):
        """Test rejection of invalid configuration schema."""
        invalid_config = {
            "invalid_key": "invalid_value"
        }

        result = self.manager.validate_config(invalid_config)

        assert result is False

    def test_should_update_config_values(self):
        """Test updating configuration values."""
        # Create initial config
        initial_config = {"version": "1.0.0", "name": "old-name"}
        self.manager.save_config(initial_config)

        # Update config
        updates = {"name": "new-name", "description": "updated description"}
        result = self.manager.update_config(updates)

        assert result is True

        # Verify updates
        updated_config = self.manager.load_config()
        assert updated_config["name"] == "new-name"
        assert updated_config["description"] == "updated description"
        assert updated_config["version"] == "1.0.0"  # Should be preserved

    def test_should_backup_config_before_changes(self):
        """Test configuration backup before changes."""
        # Create initial config
        initial_config = {"version": "1.0.0", "name": "test-project"}
        self.manager.save_config(initial_config)

        # Update with backup
        updates = {"name": "updated-name"}
        backup_path = self.manager.update_config_with_backup(updates)

        assert backup_path is not None
        assert backup_path.exists()

        # Verify backup contains original data
        backup_data = json.loads(backup_path.read_text())
        assert backup_data["name"] == "test-project"

    def test_should_restore_config_from_backup(self):
        """Test configuration restoration from backup."""
        # Create and modify config
        original_config = {"version": "1.0.0", "name": "original-name"}
        self.manager.save_config(original_config)

        backup_path = self.manager.create_backup()

        # Modify config
        modified_config = {"version": "1.0.0", "name": "modified-name"}
        self.manager.save_config(modified_config)

        # Restore from backup
        result = self.manager.restore_from_backup(backup_path)

        assert result is True

        # Verify restoration
        restored_config = self.manager.load_config()
        assert restored_config["name"] == "original-name"


class TestConfigValidation:
    """Test configuration validation functionality."""

    def setup_method(self):
        """Set up test environment."""
        self.temp_dir = Path(tempfile.mkdtemp())
        self.config_path = self.temp_dir / "config.json"
        self.manager = ConfigManager(str(self.config_path))

    def teardown_method(self):
        """Clean up test environment."""
        if self.temp_dir.exists():
            import shutil
            shutil.rmtree(self.temp_dir)

    def test_should_validate_required_fields(self):
        """Test validation of required configuration fields."""
        configs_to_test = [
            ({"version": "1.0.0"}, True),  # Minimal valid
            ({}, False),  # Missing required fields
            ({"version": "1.0.0", "name": "test"}, True),  # Valid with name
        ]

        for config_data, expected_valid in configs_to_test:
            result = self.manager.validate_required_fields(config_data)
            assert result == expected_valid

    def test_should_validate_field_types(self):
        """Test validation of configuration field types."""
        configs_to_test = [
            ({"version": "1.0.0", "name": "test"}, True),  # Correct types
            ({"version": 1.0, "name": "test"}, False),  # Wrong version type
            ({"version": "1.0.0", "name": 123}, False),  # Wrong name type
        ]

        for config_data, expected_valid in configs_to_test:
            result = self.manager.validate_field_types(config_data)
            assert result == expected_valid

    def test_should_validate_version_format(self):
        """Test validation of version format."""
        version_tests = [
            ("1.0.0", True),
            ("1.0.0-alpha", True),
            ("invalid-version", False),
            ("1.0", False),  # Should require three parts
        ]

        for version, expected_valid in version_tests:
            result = self.manager.validate_version_format(version)
            assert result == expected_valid

    def test_should_validate_project_name_format(self):
        """Test validation of project name format."""
        name_tests = [
            ("valid-project-name", True),
            ("ValidProjectName", True),
            ("project_name_123", True),
            ("", False),  # Empty name
            ("project with spaces", False),  # Spaces not allowed
            ("project/with/slashes", False),  # Slashes not allowed
        ]

        for name, expected_valid in name_tests:
            result = self.manager.validate_project_name_format(name)
            assert result == expected_valid

    def test_should_validate_config_completeness(self):
        """Test validation of configuration completeness."""
        complete_config = {
            "version": "1.0.0",
            "name": "test-project",
            "created_at": "2023-01-01T00:00:00Z",
            "runtime": {
                "name": "python",
                "version": "3.11"
            }
        }

        incomplete_config = {
            "version": "1.0.0",
            "name": "test-project"
            # Missing created_at and runtime
        }

        assert self.manager.validate_completeness(complete_config) is True
        assert self.manager.validate_completeness(incomplete_config) is False


class TestConfigMigration:
    """Test configuration migration functionality."""

    def setup_method(self):
        """Set up test environment."""
        self.temp_dir = Path(tempfile.mkdtemp())
        self.config_path = self.temp_dir / "config.json"
        self.manager = ConfigManager(str(self.config_path))

    def teardown_method(self):
        """Clean up test environment."""
        if self.temp_dir.exists():
            import shutil
            shutil.rmtree(self.temp_dir)

    def test_should_migrate_from_old_version(self):
        """Test migration from older configuration version."""
        old_config = {
            "version": "0.9.0",
            "project_name": "old-format",  # Old field name
            "settings": {"old_setting": "value"}
        }

        migrated_config = self.manager.migrate_config(old_config)

        assert migrated_config["version"] == "1.0.0"  # Should be updated
        assert "name" in migrated_config  # Should have new field name
        assert migrated_config["name"] == "old-format"

    def test_should_preserve_user_data_during_migration(self):
        """Test that user data is preserved during migration."""
        user_config = {
            "version": "0.9.0",
            "project_name": "user-project",
            "custom_settings": {
                "user_preference": "important_data"
            }
        }

        migrated_config = self.manager.migrate_config(user_config)

        assert "custom_settings" in migrated_config
        assert migrated_config["custom_settings"]["user_preference"] == "important_data"

    def test_should_handle_missing_migration_path(self):
        """Test handling of unsupported migration paths."""
        unsupported_config = {
            "version": "2.0.0",  # Future version
            "name": "future-project"
        }

        result = self.manager.migrate_config(unsupported_config)

        # Should handle gracefully (return as-is or raise appropriate error)
        assert result is not None

    def test_should_create_migration_backup(self):
        """Test creation of backup during migration."""
        old_config = {
            "version": "0.9.0",
            "project_name": "backup-test"
        }

        self.manager.save_config(old_config)
        backup_path = self.manager.migrate_with_backup()

        assert backup_path is not None
        assert backup_path.exists()

        # Verify current config is updated
        current_config = self.manager.load_config()
        assert current_config["version"] != "0.9.0"


class TestConfigErrorHandling:
    """Test configuration manager error handling."""

    def setup_method(self):
        """Set up test environment."""
        self.temp_dir = Path(tempfile.mkdtemp())
        self.config_path = self.temp_dir / "config.json"
        self.manager = ConfigManager(str(self.config_path))

    def teardown_method(self):
        """Clean up test environment."""
        if self.temp_dir.exists():
            import shutil
            shutil.rmtree(self.temp_dir)

    def test_should_handle_corrupted_config_file(self):
        """Test handling of corrupted configuration file."""
        # Create corrupted JSON file
        self.config_path.write_text("{ invalid json content }")

        loaded_config = self.manager.load_config()

        # Should handle gracefully
        assert loaded_config is None or isinstance(loaded_config, dict)

    def test_should_handle_permission_errors(self):
        """Test handling of permission errors."""
        with patch('pathlib.Path.write_text') as mock_write:
            mock_write.side_effect = PermissionError("Permission denied")

            config_data = {"version": "1.0.0"}
            result = self.manager.save_config(config_data)

            assert result is False

    def test_should_handle_disk_full_errors(self):
        """Test handling of disk full errors."""
        with patch('pathlib.Path.write_text') as mock_write:
            mock_write.side_effect = OSError("No space left on device")

            config_data = {"version": "1.0.0"}
            result = self.manager.save_config(config_data)

            assert result is False

    def test_should_handle_concurrent_access(self):
        """Test handling of concurrent configuration access."""
        import threading

        results = []

        def save_config_thread(thread_id):
            config_data = {"version": "1.0.0", "thread_id": thread_id}
            result = self.manager.save_config(config_data)
            results.append(result)

        threads = []
        for i in range(5):
            thread = threading.Thread(target=save_config_thread, args=(i,))
            threads.append(thread)
            thread.start()

        for thread in threads:
            thread.join()

        # At least one should succeed
        assert any(results)

    def test_should_handle_unicode_content(self):
        """Test handling of unicode content in configuration."""
        unicode_config = {
            "version": "1.0.0",
            "name": "테스트-프로젝트",
            "description": "한글 설명이 포함된 프로젝트",
            "author": "개발자"
        }

        result = self.manager.save_config(unicode_config)
        assert result is True

        loaded_config = self.manager.load_config()
        assert loaded_config["name"] == "테스트-프로젝트"
        assert loaded_config["description"] == "한글 설명이 포함된 프로젝트"

    def test_should_handle_large_config_files(self):
        """Test handling of large configuration files."""
        large_config = {
            "version": "1.0.0",
            "name": "large-config-test",
            "large_data": "x" * 100000  # Large string
        }

        result = self.manager.save_config(large_config)
        assert result is True

        loaded_config = self.manager.load_config()
        assert loaded_config["name"] == "large-config-test"
        assert len(loaded_config["large_data"]) == 100000


class TestConfigIntegration:
    """Integration tests for configuration manager."""

    def setup_method(self):
        """Set up test environment."""
        self.temp_dir = Path(tempfile.mkdtemp())
        self.config_path = self.temp_dir / "config.json"
        self.manager = ConfigManager(str(self.config_path))

    def teardown_method(self):
        """Clean up test environment."""
        if self.temp_dir.exists():
            import shutil
            shutil.rmtree(self.temp_dir)

    def test_should_handle_complete_config_lifecycle(self):
        """Test complete configuration lifecycle."""
        # Create
        initial_config = self.manager.create_default_config()
        save_result = self.manager.save_config(initial_config)
        assert save_result is True

        # Load
        loaded_config = self.manager.load_config()
        assert loaded_config is not None

        # Update
        updates = {"name": "updated-project"}
        update_result = self.manager.update_config(updates)
        assert update_result is True

        # Validate
        final_config = self.manager.load_config()
        validation_result = self.manager.validate_config(final_config)
        assert validation_result is True

    def test_should_maintain_config_integrity(self):
        """Test that configuration integrity is maintained."""
        original_config = {
            "version": "1.0.0",
            "name": "integrity-test",
            "checksum": "abc123"
        }

        # Save and load multiple times
        for i in range(5):
            self.manager.save_config(original_config)
            loaded_config = self.manager.load_config()

            assert loaded_config["version"] == "1.0.0"
            assert loaded_config["name"] == "integrity-test"
            assert loaded_config["checksum"] == "abc123"

    def test_should_handle_config_synchronization(self):
        """Test configuration synchronization between managers."""
        # Create second manager for same config file
        manager2 = ConfigManager(str(self.config_path))

        # Save with first manager
        config1 = {"version": "1.0.0", "name": "sync-test"}
        self.manager.save_config(config1)

        # Load with second manager
        config2 = manager2.load_config()

        assert config2["name"] == "sync-test"

        # Update with second manager
        manager2.update_config({"name": "sync-updated"})

        # Verify with first manager
        updated_config = self.manager.load_config()
        assert updated_config["name"] == "sync-updated"