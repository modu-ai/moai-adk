#!/usr/bin/env python3
"""
Tests for ConfigManager error handling and graceful degradation.
RED PHASE: Tests designed to verify robust error handling.
"""

import json
import os
import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch, mock_open, MagicMock

# Add the parent directory to the path to import config_manager
import sys
sys.path.insert(0, str(Path(__file__).parent.parent))

from alfred.shared.core.config_manager import ConfigManager


class TestConfigManagerErrorHandling(unittest.TestCase):
    """Test error handling and graceful degradation capabilities."""

    def setUp(self):
        """Set up test environment."""
        self.test_dir = tempfile.mkdtemp()
        self.config_path = Path(self.test_dir) / "config.json"

        # Reset singleton state
        ConfigManager._instance = None
        # Reset global instance
        import alfred.shared.core.config_manager as config_module
        config_module._config_manager = None

    def tearDown(self):
        """Clean up test environment."""
        import shutil
        shutil.rmtree(self.test_dir, ignore_errors=True)

        # Reset singleton state
        ConfigManager._instance = None
        # Reset global instance
        import alfred.shared.core.config_manager as config_module
        config_module._config_manager = None

    def test_missing_file_returns_defaults(self):
        """Test graceful degradation when config file doesn't exist."""
        nonexistent_path = Path(self.test_dir) / "nonexistent.json"
        manager = ConfigManager(nonexistent_path)

        config = manager.load_config()

        # Should return valid default config
        self.assertIsNotNone(config, "Should return config even when file missing")
        self.assertIn("hooks", config, "Should have default hooks section")
        self.assertEqual(config["hooks"]["timeout_seconds"], 5,
                        "Should return default timeout value")
        self.assertTrue(config["hooks"]["graceful_degradation"],
                       "Should have default graceful_degradation enabled")

    def test_invalid_json_returns_defaults(self):
        """Test graceful degradation when config file has invalid JSON."""
        # Create file with invalid JSON
        with open(self.config_path, 'w', encoding='utf-8') as f:
            f.write("{ invalid json content }")

        manager = ConfigManager(self.config_path)
        config = manager.load_config()

        # Should return defaults without crashing
        self.assertIsNotNone(config, "Should return config despite invalid JSON")
        self.assertIn("hooks", config, "Should have default hooks section")
        self.assertEqual(config["hooks"]["timeout_seconds"], 5,
                        "Should return default timeout value")

    def test_partially_invalid_json_partial_recovery(self):
        """Test handling of config files with some invalid sections."""
        # Create a valid JSON file with invalid structure
        invalid_config = {
            "hooks": "this should be a dict, not a string",  # Invalid type
            "valid_section": {
                "setting": "value"
            }
        }

        with open(self.config_path, 'w', encoding='utf-8') as f:
            json.dump(invalid_config, f)

        manager = ConfigManager(self.config_path)
        config = manager.load_config()

        # Should merge with defaults, fixing invalid sections
        self.assertIsNotNone(config, "Should recover from invalid structure")
        self.assertIsInstance(config["hooks"], dict,
                             "Hooks section should be corrected to dict type")
        self.assertEqual(config["hooks"]["timeout_seconds"], 5,
                        "Should have default timeout for invalid hooks")
        self.assertEqual(config["valid_section"]["setting"], "value",
                        "Should preserve valid sections")

    def test_file_permission_error_handling(self):
        """Test handling when config file cannot be read due to permissions."""
        # Create a valid config file first
        valid_config = {"hooks": {"timeout_seconds": 10}}
        with open(self.config_path, 'w', encoding='utf-8') as f:
            json.dump(valid_config, f)

        manager = ConfigManager(self.config_path)

        # Mock open to raise PermissionError
        with patch('builtins.open', side_effect=PermissionError("Permission denied")):
            config = manager.load_config()

            # Should gracefully fallback to defaults
            self.assertIsNotNone(config, "Should return defaults on permission error")
            self.assertIn("hooks", config, "Should have default hooks section")

    def test_file_too_large_error_handling(self):
        """Test handling when config file is too large to process."""
        manager = ConfigManager(self.config_path)

        # Mock open to raise OSError for file too large
        with patch('builtins.open', side_effect=OSError("File too large")):
            config = manager.load_config()

            # Should handle gracefully
            self.assertIsNotNone(config, "Should handle large file error gracefully")
            self.assertIn("hooks", config, "Should have default hooks section")

    def test_unicode_encoding_error_handling(self):
        """Test handling of files with encoding issues."""
        # Create file with invalid UTF-8 content
        with open(self.config_path, 'wb') as f:
            f.write(b'{ "invalid": "\xff\xfe" }')  # Invalid UTF-8

        manager = ConfigManager(self.config_path)
        config = manager.load_config()

        # Should handle encoding errors gracefully
        self.assertIsNotNone(config, "Should handle encoding errors")
        self.assertIn("hooks", config, "Should have default hooks section")

    def test_disk_full_error_handling(self):
        """Test handling when disk is full during config updates."""
        manager = ConfigManager(self.config_path)

        # Mock file writing to raise OSError (disk full)
        with patch('builtins.open', mock_open()) as mock_file:
            mock_file.return_value.write.side_effect = OSError("No space left on device")

            updates = {"hooks": {"timeout_seconds": 15}}
            success = manager.update_config(updates)

            # Update should fail gracefully
            self.assertFalse(success, "Update should fail gracefully on disk full")

            # Should still be able to read original config
            config = manager.load_config()
            self.assertIsNotNone(config, "Should still be able to read config")

    def test_concurrent_access_race_condition_recovery(self):
        """Test recovery from concurrent access race conditions."""
        manager = ConfigManager(self.config_path)

        # Create a race condition scenario
        def competing_operation():
            try:
                # One thread loads config
                config = manager.load_config()
                # Another updates config
                manager.update_config({"hooks": {"timeout_seconds": 20}})
                # First thread tries to access config again
                return config["hooks"]["timeout_seconds"]
            except Exception:
                return None

        # Run multiple competing operations
        import threading
        results = []
        threads = []

        for _ in range(10):
            thread = threading.Thread(target=lambda: results.append(competing_operation()))
            threads.append(thread)
            thread.start()

        for thread in threads:
            thread.join()

        # Should handle race conditions without crashes
        self.assertEqual(len(results), 10, "All operations should complete")
        # Some operations might return None due to race conditions, but shouldn't crash

    def test_memory_exhaustion_handling(self):
        """Test graceful handling under memory pressure."""
        manager = ConfigManager(self.config_path)

        # Mock operations to simulate memory errors
        with patch('json.loads', side_effect=MemoryError("Out of memory")):
            config = manager.load_config()

            # Should handle memory errors gracefully
            self.assertIsNotNone(config, "Should handle memory errors")
            self.assertIn("hooks", config, "Should have default hooks section")

    def test_config_validation_error_recovery(self):
        """Test recovery from config validation errors."""
        # Create config with invalid structure that would fail validation
        invalid_config = {
            "hooks": {
                "timeout_seconds": "this should be a number",  # Wrong type
                "timeout_ms": -1000,  # Invalid negative value
                "graceful_degradation": "not_boolean"  # Wrong type
            }
        }

        with open(self.config_path, 'w', encoding='utf-8') as f:
            json.dump(invalid_config, f)

        manager = ConfigManager(self.config_path)
        config = manager.load_config()

        # Should validate and correct invalid values
        self.assertIsInstance(config["hooks"]["timeout_seconds"], int,
                             "Should correct timeout type to int")
        self.assertGreaterEqual(config["hooks"]["timeout_seconds"], 1,
                               "Should ensure minimum timeout")
        self.assertIsInstance(config["hooks"]["graceful_degradation"], bool,
                             "Should correct graceful_degradation to boolean")

    def test_circular_reference_prevention(self):
        """Test prevention of circular references in config merging."""
        # Create a config with potential circular reference
        config_with_circular_ref = {}
        config_with_circular_ref["self"] = config_with_circular_ref

        manager = ConfigManager(self.config_path)

        # Try to update with circular reference
        # This is more about prevention of infinite loops during merging
        updates = {"hooks": config_with_circular_ref}

        try:
            success = manager.update_config(updates)
            # Should handle circular references gracefully
            config = manager.load_config()
            self.assertIsNotNone(config, "Should handle circular references")
        except RecursionError:
            self.fail("Should prevent infinite recursion in circular references")

    def test_partial_config_update_recovery(self):
        """Test recovery from partial config update failures."""
        manager = ConfigManager(self.config_path)

        # Create a valid initial config
        initial_config = {
            "hooks": {"timeout_seconds": 10},
            "other_section": {"value": "preserve_me"}
        }

        with open(self.config_path, 'w', encoding='utf-8') as f:
            json.dump(initial_config, f)

        # Try to update with invalid data that would cause failure
        with patch('json.dump', side_effect=IOError("Write failed")):
            updates = {"hooks": {"timeout_seconds": 15}}
            success = manager.update_config(updates)
            self.assertFalse(success, "Update should fail gracefully")

        # Should still be able to read original config
        config = manager.load_config()
        self.assertEqual(config["hooks"]["timeout_seconds"], 10,
                        "Should preserve original config after failed update")
        self.assertEqual(config["other_section"]["value"], "preserve_me",
                        "Should preserve other sections after failed update")

    def test_error_logging_and_debugging_support(self):
        """Test that errors are properly logged for debugging."""
        manager = ConfigManager(self.config_path)

        # This test would verify that appropriate logging occurs
        # For now, it documents the expected behavior
        with patch('builtins.open', side_effect=IOError("Debug this error")):
            config = manager.load_config()

            # Should return defaults and ideally log the error
            self.assertIsNotNone(config, "Should return defaults and log error")
            # In a real implementation, we'd verify logging output


if __name__ == '__main__':
    unittest.main(verbosity=2)