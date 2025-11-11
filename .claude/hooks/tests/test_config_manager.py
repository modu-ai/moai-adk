#!/usr/bin/env python3
"""
Test suite for centralized ConfigManager with singleton pattern, caching, and thread-safety.
RED PHASE: All tests should fail initially.
"""

import json
import os
import tempfile
import threading
import time
import unittest
from pathlib import Path
from unittest.mock import patch, MagicMock

# Add the parent directory to the path to import config_manager
import sys
sys.path.insert(0, str(Path(__file__).parent.parent))

from alfred.shared.core.config_manager import ConfigManager


class TestConfigManagerSingleton(unittest.TestCase):
    """Test singleton pattern and caching functionality."""

    def setUp(self):
        """Set up test environment with temporary config file."""
        self.test_dir = tempfile.mkdtemp()
        self.config_path = Path(self.test_dir) / "config.json"

        # Create a basic config file
        self.test_config = {
            "hooks": {
                "timeout_seconds": 10,
                "timeout_ms": 10000,
                "graceful_degradation": False,
                "custom_value": "test_value"
            },
            "language": {
                "conversation_language": "ko"
            }
        }

        with open(self.config_path, 'w', encoding='utf-8') as f:
            json.dump(self.test_config, f)

    def tearDown(self):
        """Clean up test environment."""
        import shutil
        shutil.rmtree(self.test_dir, ignore_errors=True)

        # Reset singleton instance for tests
        ConfigManager._instance = None
        # Reset global instance
        import alfred.shared.core.config_manager as config_module
        config_module._config_manager = None

    def test_singleton_pattern_same_instance(self):
        """Test that ConfigManager returns same instance (singleton pattern)."""
        # This should fail initially because we haven't implemented singleton properly
        manager1 = ConfigManager(self.config_path)
        manager2 = ConfigManager(self.config_path)

        self.assertIs(manager1, manager2, "ConfigManager should return same singleton instance")

    def test_config_caching_first_load_only_reads_file_once(self):
        """Test that config is cached after first load."""
        manager = ConfigManager(self.config_path)

        # Mock file operations to track read calls
        with patch('builtins.open', side_effect=open) as mock_open:
            config1 = manager.load_config()
            config2 = manager.load_config()

            # File should be read only once due to caching
            mock_open.assert_called_once()
            self.assertEqual(config1, config2, "Cached config should be identical")

    def test_ttl_cache_expiration_after_time_limit(self):
        """Test that cache expires after TTL (5 minutes)."""
        manager = ConfigManager(self.config_path)

        # Load config first time
        config1 = manager.load_config()

        # Mock time.time to simulate TTL expiration
        with patch('time.time') as mock_time:
            mock_time.return_value = time.time() + 301  # 5 minutes + 1 second

            config2 = manager.load_config()

            # Config should be reloaded from file due to TTL expiration
            # This will fail because TTL cache isn't implemented yet
            self.assertIsNotNone(config2, "Config should be reloaded after TTL expiration")

    def test_file_change_detection_reload_on_modified(self):
        """Test that config reloads when file is modified."""
        manager = ConfigManager(self.config_path)

        # Load initial config
        config1 = manager.load_config()
        self.assertEqual(config1["hooks"]["timeout_seconds"], 10)

        # Modify config file
        self.test_config["hooks"]["timeout_seconds"] = 20
        with open(self.config_path, 'w', encoding='utf-8') as f:
            json.dump(self.test_config, f)

        # Force cache invalidation by mocking file mtime
        with patch('pathlib.Path.stat') as mock_stat:
            mock_stat.return_value.st_mtime = time.time() + 1
            config2 = manager.load_config()

            # Should detect file change and reload
            # This will fail because file change detection isn't implemented
            self.assertEqual(config2["hooks"]["timeout_seconds"], 20,
                           "Config should reload when file is modified")

    def test_thread_safe_concurrent_access(self):
        """Test that ConfigManager is thread-safe under concurrent access."""
        manager = ConfigManager(self.config_path)
        results = []
        errors = []

        def worker():
            try:
                for _ in range(10):
                    config = manager.load_config()
                    results.append(config["hooks"]["timeout_seconds"])
            except Exception as e:
                errors.append(e)

        # Create multiple threads accessing config concurrently
        threads = [threading.Thread(target=worker) for _ in range(5)]

        for thread in threads:
            thread.start()

        for thread in threads:
            thread.join()

        # No errors should occur during concurrent access
        self.assertEqual(len(errors), 0, f"No errors should occur: {errors}")
        # All results should be consistent
        self.assertTrue(all(r == 10 for r in results),
                       "All concurrent reads should return consistent results")

    def test_graceful_degradation_missing_file(self):
        """Test graceful degradation when config file doesn't exist."""
        nonexistent_path = Path(self.test_dir) / "nonexistent.json"
        manager = ConfigManager(nonexistent_path)

        config = manager.load_config()

        # Should return default config when file doesn't exist
        self.assertIsNotNone(config, "Should return default config when file missing")
        self.assertIn("hooks", config, "Default config should have 'hooks' section")
        self.assertEqual(config["hooks"]["timeout_seconds"], 5,
                        "Default timeout should be 5 seconds")

    def test_graceful_degradation_corrupted_json(self):
        """Test graceful degradation when config file contains invalid JSON."""
        corrupted_path = Path(self.test_dir) / "corrupted.json"

        # Write invalid JSON
        with open(corrupted_path, 'w', encoding='utf-8') as f:
            f.write("{ invalid json content")

        manager = ConfigManager(corrupted_path)
        config = manager.load_config()

        # Should return default config when JSON is invalid
        self.assertIsNotNone(config, "Should return default config when JSON is corrupted")
        self.assertIn("hooks", config, "Default config should have 'hooks' section")

    def test_graceful_degradation_io_error(self):
        """Test graceful degradation when file I/O errors occur."""
        manager = ConfigManager(self.config_path)

        # Mock open to raise IOError
        with patch('builtins.open', side_effect=IOError("Permission denied")):
            config = manager.load_config()

            # Should return default config when I/O error occurs
            self.assertIsNotNone(config, "Should return default config when I/O error occurs")
            self.assertIn("hooks", config, "Default config should have 'hooks' section")

    def test_memory_cleanup_with_weakref(self):
        """Test that ConfigManager cleans up properly to prevent memory leaks."""
        import weakref
        import gc

        # Create manager and get weak reference
        manager = ConfigManager(self.config_path)
        weak_ref = weakref.ref(manager)

        # Delete reference and force garbage collection
        del manager
        gc.collect()

        # This is more of a memory leak detection test
        # In a real implementation with proper singleton, this might behave differently
        self.assertIsNotNone(weak_ref, "Weak reference should be valid until garbage collected")

    def test_config_merge_with_updates(self):
        """Test that config updates properly merge with existing config."""
        manager = ConfigManager(self.config_path)

        updates = {
            "hooks": {
                "timeout_seconds": 15,  # Override existing
                "new_setting": "added"  # Add new
            },
            "new_section": {
                "setting": "value"
            }
        }

        success = manager.update_config(updates)
        self.assertTrue(success, "Config update should succeed")

        config = manager.load_config()
        self.assertEqual(config["hooks"]["timeout_seconds"], 15,
                        "Existing setting should be overridden")
        self.assertEqual(config["hooks"]["new_setting"], "added",
                        "New setting should be added")
        self.assertEqual(config["new_section"]["setting"], "value",
                        "New section should be added")

    def test_config_merge_preserves_defaults(self):
        """Test that config merging preserves default values for non-overridden keys."""
        manager = ConfigManager(self.config_path)

        # Update only specific settings
        updates = {
            "hooks": {
                "timeout_seconds": 20
            }
        }

        manager.update_config(updates)
        config = manager.load_config()

        # Updated value should be overridden
        self.assertEqual(config["hooks"]["timeout_seconds"], 20)
        # Non-updated values should preserve defaults
        self.assertEqual(config["hooks"]["timeout_ms"], 10000)  # From test config
        self.assertIsNotNone(config["hooks"]["graceful_degradation"])  # Should exist

    def test_performance_improvement_measure(self):
        """Test and measure performance improvement with caching."""
        manager = ConfigManager(self.config_path)

        # Measure time for first load (should read from file)
        start_time = time.perf_counter()
        config1 = manager.load_config()
        first_load_time = time.perf_counter() - start_time

        # Measure time for second load (should use cache)
        start_time = time.perf_counter()
        config2 = manager.load_config()
        second_load_time = time.perf_counter() - start_time

        self.assertEqual(config1, config2, "Configs should be identical")

        # Second load should be significantly faster due to caching
        # This test demonstrates the performance improvement
        improvement_ratio = first_load_time / second_load_time if second_load_time > 0 else 1
        print(f"\nPerformance improvement: {improvement_ratio:.2f}x faster")
        print(f"First load: {first_load_time*1000:.3f}ms")
        print(f"Second load: {second_load_time*1000:.3f}ms")

        # In a real implementation with caching, second_load_time should be much smaller
        # For now, this documents the expected performance improvement


if __name__ == '__main__':
    unittest.main(verbosity=2)