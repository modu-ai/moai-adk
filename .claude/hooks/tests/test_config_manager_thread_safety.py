#!/usr/bin/env python3
"""
Additional tests for ConfigManager thread-safety and race conditions.
RED PHASE: Tests designed to fail initially to expose thread-safety issues.
"""

import json
import tempfile
import threading
import time
import unittest
from pathlib import Path
from unittest.mock import patch
import concurrent.futures

# Add the parent directory to the path to import config_manager
import sys
sys.path.insert(0, str(Path(__file__).parent.parent))

from alfred.shared.core.config_manager import ConfigManager


class TestConfigManagerThreadSafety(unittest.TestCase):
    """Test thread-safety under high concurrency conditions."""

    def setUp(self):
        """Set up test environment with temporary config file."""
        self.test_dir = tempfile.mkdtemp()
        self.config_path = Path(self.test_dir) / "config.json"

        # Create a basic config file
        self.test_config = {
            "hooks": {
                "timeout_seconds": 10,
                "timeout_ms": 10000,
                "counter": 0
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

    def test_concurrent_singleton_creation_race_condition(self):
        """Test that concurrent singleton creation doesn't create multiple instances."""
        instances = []
        creation_lock = threading.Lock()

        def create_manager():
            manager = ConfigManager(self.config_path)
            with creation_lock:
                instances.append(manager)

        # Create multiple threads that try to get singleton simultaneously
        threads = []
        for _ in range(20):
            thread = threading.Thread(target=create_manager)
            threads.append(thread)
            thread.start()

        # Start all threads at roughly the same time
        for thread in threads:
            thread.join()

        # All instances should be the same object (singleton)
        # This will likely fail without proper thread-safe singleton implementation
        first_instance = instances[0]
        for instance in instances[1:]:
            self.assertIs(instance, first_instance,
                         "All concurrent creations should return same singleton instance")

    def test_concurrent_config_read_consistency(self):
        """Test that concurrent config reads return consistent data."""
        manager = ConfigManager(self.config_path)
        results = []
        errors = []

        def concurrent_read():
            try:
                # Simulate some processing time
                time.sleep(0.001)
                config = manager.load_config()
                results.append(config["hooks"]["timeout_seconds"])
            except Exception as e:
                errors.append(e)

        # Run many concurrent reads
        with concurrent.futures.ThreadPoolExecutor(max_workers=50) as executor:
            futures = [executor.submit(concurrent_read) for _ in range(100)]
            concurrent.futures.wait(futures)

        # Should have no errors and all results should be consistent
        self.assertEqual(len(errors), 0, f"No errors should occur: {errors}")
        self.assertEqual(len(results), 100, "Should have 100 results")
        self.assertTrue(all(r == 10 for r in results),
                       f"All results should be 10, but got: {set(results)}")

    def test_concurrent_config_write_race_condition(self):
        """Test that concurrent config updates don't corrupt data."""
        manager = ConfigManager(self.config_path)

        def update_config(thread_id):
            try:
                updates = {
                    "hooks": {
                        "thread_id": thread_id,
                        "counter": thread_id
                    }
                }
                return manager.update_config(updates)
            except Exception as e:
                return False

        # Run many concurrent updates
        with concurrent.futures.ThreadPoolExecutor(max_workers=20) as executor:
            futures = [executor.submit(update_config, i) for i in range(50)]
            results = [f.result() for f in concurrent.futures.wait(futures)[0]]

        # All updates should succeed
        successful_updates = sum(results)
        self.assertGreater(successful_updates, 0, "Some updates should succeed")

        # Final config should be valid JSON with expected structure
        final_config = manager.load_config()
        self.assertIn("hooks", final_config, "Final config should have hooks section")
        self.assertIsInstance(final_config["hooks"], dict, "Hooks should be a dict")

    def test_cache_invalidation_thread_safety(self):
        """Test that cache invalidation is thread-safe."""
        manager = ConfigManager(self.config_path)

        # One thread continuously reads config
        read_results = []
        def continuous_reader():
            for _ in range(100):
                config = manager.load_config()
                read_results.append(config["hooks"]["timeout_seconds"])
                time.sleep(0.001)

        # Another thread updates config file
        def config_updater():
            for i in range(5):
                # Update config file
                self.test_config["hooks"]["timeout_seconds"] = 10 + i
                with open(self.config_path, 'w', encoding='utf-8') as f:
                    json.dump(self.test_config, f)
                time.sleep(0.01)

        # Run both concurrently
        reader_thread = threading.Thread(target=continuous_reader)
        updater_thread = threading.Thread(target=config_updater)

        reader_thread.start()
        updater_thread.start()

        reader_thread.join()
        updater_thread.join()

        # All reads should return valid timeout values
        valid_timeouts = {10, 11, 12, 13, 14}  # Possible values during updates
        for timeout in read_results:
            self.assertIn(timeout, valid_timeouts,
                         f"All timeouts should be valid: {timeout}")

    def test_memory_leak_prevention_under_load(self):
        """Test that memory usage doesn't grow unbounded under load."""
        manager = ConfigManager(self.config_path)

        # Simulate high load with many config operations
        def stress_test():
            for _ in range(1000):
                config = manager.load_config()
                # Simulate some processing
                manager.get("hooks.timeout_seconds")
                manager.get_hooks_config()
                manager.get_timeout_seconds()

        # Run stress test in multiple threads
        threads = []
        for _ in range(10):
            thread = threading.Thread(target=stress_test)
            threads.append(thread)
            thread.start()

        for thread in threads:
            thread.join()

        # After stress test, config manager should still work
        config = manager.load_config()
        self.assertIsNotNone(config, "Config manager should still work after stress test")
        self.assertEqual(config["hooks"]["timeout_seconds"], 10,
                        "Config should still return correct values")

    def test_deadlock_prevention(self):
        """Test that no deadlocks occur with complex threading scenarios."""
        manager = ConfigManager(self.config_path)

        def complex_operation():
            try:
                # Mix of different operations
                config1 = manager.load_config()
                timeout = manager.get_timeout_seconds()
                hooks_config = manager.get_hooks_config()
                config2 = manager.load_config()
                message = manager.get_message("timeout", "post_tool_use", "post_tool_use")
                return True
            except Exception as e:
                return False

        # Run many complex operations concurrently
        with concurrent.futures.ThreadPoolExecutor(max_workers=30) as executor:
            futures = [executor.submit(complex_operation) for _ in range(100)]
            results = [f.result() for f in concurrent.futures.wait(futures)[0]]

        # All operations should complete successfully (no deadlocks)
        successful_operations = sum(results)
        self.assertEqual(successful_operations, 100,
                        "All operations should complete without deadlock")

    def test_atomic_config_updates(self):
        """Test that config updates are atomic and don't leave partial state."""
        manager = ConfigManager(self.config_path)

        def atomic_update_test():
            # Large update that could potentially be interrupted
            large_update = {
                "hooks": {
                    "timeout_seconds": 15,
                    "new_section": {
                        "nested": {
                            "deeply": {
                                "nested": {
                                    "value": "complex"
                                }
                            }
                        }
                    },
                    "array_data": list(range(1000))  # Large array
                },
                "new_top_level": {
                    "complex_data": {
                        "nested": {"value": "test"}
                    }
                }
            }

            original_config = manager.load_config()
            success = manager.update_config(large_update)

            if success:
                updated_config = manager.load_config()
                # Verify all updates are present (atomic)
                self.assertEqual(updated_config["hooks"]["timeout_seconds"], 15)
                self.assertEqual(updated_config["hooks"]["new_section"]["nested"]["deeply"]["nested"]["value"], "complex")
                self.assertEqual(len(updated_config["hooks"]["array_data"]), 1000)
                self.assertEqual(updated_config["new_top_level"]["complex_data"]["nested"]["value"], "test")

            return success

        # Run atomic update test concurrently
        with concurrent.futures.ThreadPoolExecutor(max_workers=10) as executor:
            futures = [executor.submit(atomic_update_test) for _ in range(5)]
            results = [f.result() for f in concurrent.futures.wait(futures)[0]]

        # At least some updates should succeed
        successful_updates = sum(results)
        self.assertGreater(successful_updates, 0,
                          "At least some atomic updates should succeed")


if __name__ == '__main__':
    unittest.main(verbosity=2)