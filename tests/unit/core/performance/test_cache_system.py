"""
Cache System Tests

Test cases for caching system functionality.
"""

import tempfile

import pytest


class TestCacheSystem:
    """Test suite for cache system functionality."""

    def test_cache_system_creation(self):
        """Test that cache system can be created successfully."""
        # This test should fail initially as the CacheSystem class doesn't exist
        from moai_adk.core.performance.cache_system import CacheSystem

        cache = CacheSystem()
        assert cache is not None

    def test_cache_system_with_temp_directory(self):
        """Test cache system with temporary directory."""
        # This test should fail initially
        from moai_adk.core.performance.cache_system import CacheSystem

        with tempfile.TemporaryDirectory() as temp_dir:
            cache = CacheSystem(cache_dir=temp_dir)
            assert cache.cache_dir == temp_dir

    def test_cache_system_get_set(self):
        """Test basic get and set operations."""
        # This test should fail initially
        from moai_adk.core.performance.cache_system import CacheSystem

        with tempfile.TemporaryDirectory() as temp_dir:
            cache = CacheSystem(cache_dir=temp_dir)

            # Set a value
            cache.set("test_key", {"data": "test_value"})

            # Get the value
            result = cache.get("test_key")
            assert result == {"data": "test_value"}

    def test_cache_system_get_nonexistent(self):
        """Test getting non-existent key."""
        # This test should fail initially
        from moai_adk.core.performance.cache_system import CacheSystem

        with tempfile.TemporaryDirectory() as temp_dir:
            cache = CacheSystem(cache_dir=temp_dir)

            # Get non-existent key
            result = cache.get("nonexistent_key")
            assert result is None

    def test_cache_system_delete(self):
        """Test deleting cached values."""
        # This test should fail initially
        from moai_adk.core.performance.cache_system import CacheSystem

        with tempfile.TemporaryDirectory() as temp_dir:
            cache = CacheSystem(cache_dir=temp_dir)

            # Set a value
            cache.set("test_key", {"data": "test_value"})
            assert cache.get("test_key") is not None

            # Delete the value
            cache.delete("test_key")
            assert cache.get("test_key") is None

    def test_cache_system_clear(self):
        """Test clearing all cached values."""
        # This test should fail initially
        from moai_adk.core.performance.cache_system import CacheSystem

        with tempfile.TemporaryDirectory() as temp_dir:
            cache = CacheSystem(cache_dir=temp_dir)

            # Set multiple values
            cache.set("key1", "value1")
            cache.set("key2", "value2")
            cache.set("key3", "value3")

            # Clear all values
            cache.clear()

            # Verify all values are cleared
            assert cache.get("key1") is None
            assert cache.get("key2") is None
            assert cache.get("key3") is None

    def test_cache_system_exists(self):
        """Test checking if key exists."""
        # This test should fail initially
        from moai_adk.core.performance.cache_system import CacheSystem

        with tempfile.TemporaryDirectory() as temp_dir:
            cache = CacheSystem(cache_dir=temp_dir)

            # Initially key shouldn't exist
            assert not cache.exists("test_key")

            # Set the key
            cache.set("test_key", {"data": "test_value"})

            # Now key should exist
            assert cache.exists("test_key")

    def test_cache_system_size(self):
        """Test getting cache size."""
        # This test should fail initially
        from moai_adk.core.performance.cache_system import CacheSystem

        with tempfile.TemporaryDirectory() as temp_dir:
            cache = CacheSystem(cache_dir=temp_dir)

            # Initially cache should be empty
            assert cache.size() == 0

            # Add some values
            cache.set("key1", "value1")
            cache.set("key2", "value2")

            # Cache should have 2 items
            assert cache.size() == 2

    def test_cache_system_ttl(self):
        """Test time-to-live functionality."""
        # This test should fail initially
        from moai_adk.core.performance.cache_system import CacheSystem

        with tempfile.TemporaryDirectory() as temp_dir:
            cache = CacheSystem(cache_dir=temp_dir)

            # Set value with TTL of 1 second
            cache.set("test_key", {"data": "test_value"}, ttl=1)

            # Value should exist immediately
            assert cache.get("test_key") is not None

            # Wait for TTL to expire
            import time

            time.sleep(1.1)

            # Value should now be expired
            assert cache.get("test_key") is None

    def test_cache_system_set_if_not_exists(self):
        """Test setting value only if key doesn't exist."""
        # This test should fail initially
        from moai_adk.core.performance.cache_system import CacheSystem

        with tempfile.TemporaryDirectory() as temp_dir:
            cache = CacheSystem(cache_dir=temp_dir)

            # Set initial value
            cache.set("test_key", {"data": "original_value"})

            # Try to set again with set_if_not_exists - should fail
            result = cache.set_if_not_exists("test_key", {"data": "new_value"})
            assert result is False
            assert cache.get("test_key") == {"data": "original_value"}

            # Set new key - should succeed
            result = cache.set_if_not_exists("new_key", {"data": "new_value"})
            assert result is True
            assert cache.get("new_key") == {"data": "new_value"}

    def test_cache_system_get_multiple(self):
        """Test getting multiple keys at once."""
        # This test should fail initially
        from moai_adk.core.performance.cache_system import CacheSystem

        with tempfile.TemporaryDirectory() as temp_dir:
            cache = CacheSystem(cache_dir=temp_dir)

            # Set multiple values
            cache.set("key1", "value1")
            cache.set("key2", "value2")
            cache.set("key3", "value3")

            # Get multiple keys
            result = cache.get_multiple(["key1", "key3", "nonexistent"])
            assert result == {"key1": "value1", "key3": "value3", "nonexistent": None}

    def test_cache_system_persistence(self):
        """Test that cache persists across CacheSystem instances."""
        # This test should fail initially
        from moai_adk.core.performance.cache_system import CacheSystem

        with tempfile.TemporaryDirectory() as temp_dir:
            # Set value with first instance
            cache1 = CacheSystem(cache_dir=temp_dir)
            cache1.set("persistent_key", {"data": "persistent_value"})

            # Get value with second instance
            cache2 = CacheSystem(cache_dir=temp_dir)
            result = cache2.get("persistent_key")
            assert result == {"data": "persistent_value"}

    def test_cache_system_validation(self):
        """Test input validation."""
        from moai_adk.core.performance.cache_system import CacheSystem

        with tempfile.TemporaryDirectory() as temp_dir:
            cache = CacheSystem(cache_dir=temp_dir)

            # Test invalid key type
            with pytest.raises(TypeError, match="Cache key must be a string"):
                cache.set(123, "value")

            # Test empty key
            with pytest.raises(ValueError, match="Cache key cannot be empty"):
                cache.set("", "value")

            # Test whitespace key
            with pytest.raises(ValueError, match="Cache key cannot be empty"):
                cache.set("   ", "value")

            # Test invalid value (non-serializable)
            with pytest.raises(TypeError, match="Cache value must be JSON serializable"):
                cache.set("bad_key", lambda x: x)

            # Test invalid TTL
            with pytest.raises(ValueError, match="TTL must be a positive number"):
                cache.set("key", "value", ttl=-1)

    def test_cache_system_auto_cleanup(self):
        """Test auto-cleanup functionality."""
        from moai_adk.core.performance.cache_system import CacheSystem

        with tempfile.TemporaryDirectory() as temp_dir:
            # Test with auto-cleanup enabled (default)
            cache1 = CacheSystem(cache_dir=temp_dir, auto_cleanup=True)
            cache1.set("key1", "value1", ttl=1)

            # Test with auto-cleanup disabled
            cache2 = CacheSystem(cache_dir=temp_dir, auto_cleanup=False)
            cache2.set("key2", "value2", ttl=1)

            # Wait for expiration
            import time

            time.sleep(1.1)

            # With auto-cleanup, key1 should be gone
            assert cache1.get("key1") is None

            # Without auto-cleanup, key2 should still exist but be expired
            assert cache2.get("key2") is None

    def test_cache_system_stats(self):
        """Test cache statistics."""
        from moai_adk.core.performance.cache_system import CacheSystem

        with tempfile.TemporaryDirectory() as temp_dir:
            cache = CacheSystem(cache_dir=temp_dir)

            # Initially empty
            stats = cache.get_stats()
            assert stats["total_files"] == 0
            assert stats["expired_files"] == 0
            assert stats["valid_files"] == 0

            # Add some values
            cache.set("valid_key", "valid_value")
            cache.set("expired_key", "expired_value", ttl=1)

            # Stats should show 2 files
            stats = cache.get_stats()
            assert stats["total_files"] == 2
            assert stats["expired_files"] == 0
            assert stats["valid_files"] == 2

            # Wait for expiration
            import time

            time.sleep(1.1)

            # Stats should show 1 expired file
            stats = cache.get_stats()
            assert stats["total_files"] == 2
            assert stats["expired_files"] == 1
            assert stats["valid_files"] == 1

    def test_cache_system_clear_count(self):
        """Test clear method returns count of removed files."""
        from moai_adk.core.performance.cache_system import CacheSystem

        with tempfile.TemporaryDirectory() as temp_dir:
            cache = CacheSystem(cache_dir=temp_dir)

            # Add multiple values
            cache.set("key1", "value1")
            cache.set("key2", "value2")
            cache.set("key3", "value3")

            # Clear should return 3
            count = cache.clear()
            assert count == 3

            # Cache should be empty
            assert cache.size() == 0

    def test_cache_system_delete_return_value(self):
        """Test delete method returns boolean indicating success."""
        from moai_adk.core.performance.cache_system import CacheSystem

        with tempfile.TemporaryDirectory() as temp_dir:
            cache = CacheSystem(cache_dir=temp_dir)

            # Delete non-existent key should return False
            assert not cache.delete("nonexistent_key")

            # Set a value and delete it should return True
            cache.set("existing_key", "value")
            assert cache.delete("existing_key")

            # Verify it's gone
            assert not cache.exists("existing_key")
