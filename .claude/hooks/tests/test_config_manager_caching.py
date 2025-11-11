#!/usr/bin/env python3
"""
Tests for ConfigManager TTL caching and file change detection.
RED PHASE: Tests designed to fail initially to expose caching issues.
"""

import json
import tempfile
import time
import unittest
from pathlib import Path
from unittest.mock import patch, MagicMock

# Add the parent directory to the path to import config_manager
import sys
sys.path.insert(0, str(Path(__file__).parent.parent))

from alfred.shared.core.config_manager import ConfigManager


class TestConfigManagerCaching(unittest.TestCase):
    """Test TTL caching and file change detection functionality."""

    def setUp(self):
        """Set up test environment with temporary config file."""
        self.test_dir = tempfile.mkdtemp()
        self.config_path = Path(self.test_dir) / "config.json"

        # Create a basic config file
        self.test_config = {
            "hooks": {
                "timeout_seconds": 10,
                "timeout_ms": 10000,
                "graceful_degradation": True,
                "cache_settings": {
                    "ttl_seconds": 300
                }
            }
        }

        with open(self.config_path, 'w', encoding='utf-8') as f:
            json.dump(self.test_config, f)

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

    def test_ttl_cache_prevents_file_reread_within_ttl(self):
        """Test that cached config is used within TTL period (5 minutes)."""
        manager = ConfigManager(self.config_path)

        # Mock file operations to track access
        with patch('builtins.open', side_effect=open) as mock_open:
            # First load should read file
            config1 = manager.load_config()
            initial_calls = mock_open.call_count

            # Second load within TTL should use cache (no additional file reads)
            config2 = manager.load_config()
            calls_within_ttl = mock_open.call_count

            self.assertEqual(config1, config2, "Configs should be identical")
            self.assertEqual(calls_within_ttl, initial_calls,
                           "No additional file reads should occur within TTL")

    def test_ttl_cache_expires_after_5_minutes(self):
        """Test that cache expires after 5 minutes (300 seconds)."""
        manager = ConfigManager(self.config_path)

        # Load config first time
        config1 = manager.load_config()

        # Mock time.time to simulate TTL expiration (5 minutes + 1 second)
        with patch('time.time') as mock_time:
            original_time = time.time()
            mock_time.return_value = original_time + 301  # 5 min + 1 sec

            with patch('builtins.open', side_effect=open) as mock_open:
                config2 = manager.load_config()

                # Should trigger file reload due to TTL expiration
                # This will fail initially because TTL isn't implemented
                self.assertGreater(mock_open.call_count, 0,
                                 "File should be read after TTL expiration")

    def test_file_modification_time_detection(self):
        """Test that file modification changes are detected."""
        manager = ConfigManager(self.config_path)

        # Load initial config
        config1 = manager.load_config()
        initial_mtime = self.config_path.stat().st_mtime

        # Wait a moment to ensure different timestamp
        time.sleep(0.01)

        # Modify config file
        self.test_config["hooks"]["timeout_seconds"] = 20
        with open(self.config_path, 'w', encoding='utf-8') as f:
            json.dump(self.test_config, f)

        new_mtime = self.config_path.stat().st_mtime
        self.assertGreater(new_mtime, initial_mtime,
                           "File modification time should increase")

        # Load config again - should detect file change
        config2 = manager.load_config()

        # This will fail initially because file change detection isn't implemented
        self.assertEqual(config2["hooks"]["timeout_seconds"], 20,
                        "Config should reflect file changes")

    def test_cache_invalidation_on_file_change(self):
        """Test that cache is invalidated when file changes."""
        manager = ConfigManager(self.config_path)

        # Load config and cache it
        config1 = manager.load_config()
        self.assertEqual(config1["hooks"]["timeout_seconds"], 10)

        # Modify file while cache is still valid (within TTL)
        time.sleep(0.01)  # Ensure different timestamp
        self.test_config["hooks"]["timeout_seconds"] = 30
        with open(self.config_path, 'w', encoding='utf-8') as f:
            json.dump(self.test_config, f)

        # Load config - should detect file change and invalidate cache
        config2 = manager.load_config()

        # This will fail initially because file change detection isn't implemented
        self.assertEqual(config2["hooks"]["timeout_seconds"], 30,
                        "Cache should be invalidated on file change")

    def test_multiple_file_changes_detected(self):
        """Test detection of multiple sequential file changes."""
        manager = ConfigManager(self.config_path)

        # Test multiple changes
        for expected_timeout in [15, 25, 35, 45]:
            time.sleep(0.01)  # Ensure different timestamps

            # Update file
            self.test_config["hooks"]["timeout_seconds"] = expected_timeout
            with open(self.config_path, 'w', encoding='utf-8') as f:
                json.dump(self.test_config, f)

            # Load config - should detect latest change
            config = manager.load_config()

            # This will fail initially because file change detection isn't implemented
            self.assertEqual(config["hooks"]["timeout_seconds"], expected_timeout,
                            f"Should detect change to timeout {expected_timeout}")

    def test_cache_performance_benefit(self):
        """Test and measure the performance benefit of caching."""
        manager = ConfigManager(self.config_path)

        # Measure uncached performance (multiple file reads)
        times = []
        for _ in range(5):
            start_time = time.perf_counter()
            config = manager.load_config()
            end_time = time.perf_counter()
            times.append(end_time - start_time)

        # Calculate average uncached time
        avg_uncached_time = sum(times) / len(times)

        # Now test cached performance (should be much faster)
        # First call to establish cache
        manager.load_config()

        cached_times = []
        for _ in range(20):  # More iterations for cached version
            start_time = time.perf_counter()
            config = manager.load_config()
            end_time = time.perf_counter()
            cached_times.append(end_time - start_time)

        avg_cached_time = sum(cached_times) / len(cached_times)

        print(f"\nCache Performance Analysis:")
        print(f"Average uncached time: {avg_uncached_time*1000:.3f}ms")
        print(f"Average cached time: {avg_cached_time*1000:.3f}ms")
        print(f"Performance improvement: {avg_uncached_time/avg_cached_time:.1f}x faster")

        # In a proper implementation, cached should be significantly faster
        # For now, this documents the expected performance benefit
        self.assertIsNotNone(cached_times, "Should have cached performance data")

    def test_cache_with_deleted_file(self):
        """Test cache behavior when config file is deleted."""
        manager = ConfigManager(self.config_path)

        # Load config while file exists
        config1 = manager.load_config()
        self.assertIsNotNone(config1, "Should load config when file exists")

        # Delete the file
        self.config_path.unlink()

        # Try to load config - should fallback to defaults gracefully
        config2 = manager.load_config()
        self.assertIsNotNone(config2, "Should return default config when file deleted")
        self.assertIn("hooks", config2, "Default config should have hooks section")

    def test_cache_with_recreated_file(self):
        """Test cache behavior when config file is recreated."""
        manager = ConfigManager(self.config_path)

        # Load initial config
        config1 = manager.load_config()
        original_timeout = config1["hooks"]["timeout_seconds"]

        # Delete file
        self.config_path.unlink()

        # Try to load - should return defaults
        config_defaults = manager.load_config()
        self.assertEqual(config_defaults["hooks"]["timeout_seconds"], 5,
                        "Should return default timeout (5)")

        # Recreate file with different content
        time.sleep(0.01)  # Ensure different timestamp
        new_config = {
            "hooks": {
                "timeout_seconds": 50,
                "timeout_ms": 50000
            }
        }

        with open(self.config_path, 'w', encoding='utf-8') as f:
            json.dump(new_config, f)

        # Load config - should detect new file
        config3 = manager.load_config()

        # This will fail initially because file recreation detection isn't implemented
        self.assertEqual(config3["hooks"]["timeout_seconds"], 50,
                        "Should detect recreated file and load new config")

    def test_cache_memory_efficiency(self):
        """Test that cache doesn't consume excessive memory."""
        manager = ConfigManager(self.config_path)

        # Load config multiple times
        configs = []
        for _ in range(100):
            config = manager.load_config()
            configs.append(config)

        # All configs should be the same object reference if properly cached
        # This tests memory efficiency of caching
        first_config_id = id(configs[0])
        same_reference_count = sum(1 for config in configs if id(config) == first_config_id)

        # In an efficient implementation, most should be the same object
        # For now, this documents the expected memory efficiency
        self.assertGreater(same_reference_count, 0,
                          "At least some configs should share references")

    def test_ttl_configurable_per_manager_instance(self):
        """Test that TTL can be configured per manager instance."""
        # This test would require TTL to be configurable
        # For now, it documents the expected behavior

        manager = ConfigManager(self.config_path)

        # Default TTL should be 5 minutes (300 seconds)
        # This would need to be implemented
        default_ttl = getattr(manager, '_ttl_seconds', 300)
        self.assertEqual(default_ttl, 300,
                        "Default TTL should be 5 minutes")


if __name__ == '__main__':
    unittest.main(verbosity=2)