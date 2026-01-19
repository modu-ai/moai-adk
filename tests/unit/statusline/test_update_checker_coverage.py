"""
Comprehensive test coverage for update checker module.

Tests for uncovered lines in update_checker.py:
- _is_update_available error handling (lines 115-117)
"""

from unittest.mock import MagicMock, patch

from moai_adk.statusline.update_checker import UpdateChecker, UpdateInfo


class TestIsUpdateAvailableErrorHandling:
    """Test _is_update_available error handling (lines 115-117)."""

    def test_is_update_available_non_numeric_current(self):
        """Test _is_update_available with non-numeric current version."""
        # Arrange
        checker = UpdateChecker()

        # Act
        result = checker._is_update_available("abc", "1.0.0")

        # Assert - should handle gracefully and return False
        assert result is False

    def test_is_update_available_non_numeric_latest(self):
        """Test _is_update_available with non-numeric latest version."""
        # Arrange
        checker = UpdateChecker()

        # Act
        result = checker._is_update_available("1.0.0", "xyz")

        # Assert - should handle gracefully and return False
        assert result is False

    def test_is_update_available_empty_strings(self):
        """Test _is_update_available with empty strings."""
        # Arrange
        checker = UpdateChecker()

        # Act
        result = checker._is_update_available("", "")

        # Assert
        assert result is False

    def test_is_update_available_with_special_characters(self):
        """Test _is_update_available with special characters in version."""
        # Arrange
        checker = UpdateChecker()

        # Act
        result = checker._is_update_available("1.0.0-beta", "1.0.0")

        # Assert - should handle beta versions
        # The regex splits on non-digits, so "1.0.0-beta" becomes [1, 0, 0]
        # And "1.0.0" becomes [1, 0, 0], so they're equal
        assert result is False

    def test_is_update_available_with_v_prefix_both(self):
        """Test _is_update_available with v prefix on both."""
        # Arrange
        checker = UpdateChecker()

        # Act
        result = checker._is_update_available("v1.0.0", "v1.1.0")

        # Assert
        assert result is True

    def test_is_update_available_with_v_prefix_current_only(self):
        """Test _is_update_available with v prefix on current only."""
        # Arrange
        checker = UpdateChecker()

        # Act
        result = checker._is_update_available("v1.0.0", "1.1.0")

        # Assert
        assert result is True

    def test_is_update_available_with_v_prefix_latest_only(self):
        """Test _is_update_available with v prefix on latest only."""
        # Arrange
        checker = UpdateChecker()

        # Act
        result = checker._is_update_available("1.0.0", "v1.1.0")

        # Assert
        assert result is True

    def test_is_update_available_with_pre_release_identifiers(self):
        """Test _is_update_available with pre-release identifiers."""
        # Arrange
        checker = UpdateChecker()

        # Act - compare pre-release with stable
        result = checker._is_update_available("1.0.0-alpha", "1.0.0")

        # Assert - should handle pre-release versions
        # The regex extracts only digits, so both become [1, 0, 0]
        assert result is False

    def test_is_update_available_with_build_metadata(self):
        """Test _is_update_available with build metadata."""
        # Arrange
        checker = UpdateChecker()

        # Act
        result = checker._is_update_available("1.0.0+build123", "1.0.0+build456")

        # Assert - build metadata should be ignored
        # Both become [1, 0, 0]
        assert result is False

    def test_is_update_available_with_dots_only(self):
        """Test _is_update_available with version containing only dots."""
        # Arrange
        checker = UpdateChecker()

        # Act
        result = checker._is_update_available("...", "1.0.0")

        # Assert - should handle invalid version format
        # After regex split, we get empty list or list with empty strings
        # The list comprehension filters out non-digits, resulting in []
        assert result is False

    def test_is_update_available_very_long_version(self):
        """Test _is_update_available with very long version numbers."""
        # Arrange
        checker = UpdateChecker()

        # Act
        result = checker._is_update_available("1.2.3.4.5.6.7.8.9.10", "1.2.3.4.5.6.7.8.9.11")

        # Assert
        assert result is True

    def test_is_update_available_current_greater(self):
        """Test _is_update_available when current is greater than latest."""
        # Arrange
        checker = UpdateChecker()

        # Act
        result = checker._is_update_available("2.0.0", "1.0.0")

        # Assert
        assert result is False

    def test_is_update_available_same_version(self):
        """Test _is_update_available with same versions."""
        # Arrange
        checker = UpdateChecker()

        # Act
        result = checker._is_update_available("1.0.0", "1.0.0")

        # Assert
        assert result is False

    def test_is_update_available_minor_version_greater(self):
        """Test _is_update_available with minor version greater."""
        # Arrange
        checker = UpdateChecker()

        # Act
        result = checker._is_update_available("1.0.0", "1.1.0")

        # Assert
        assert result is True

    def test_is_update_available_patch_version_greater(self):
        """Test _is_update_available with patch version greater."""
        # Arrange
        checker = UpdateChecker()

        # Act
        result = checker._is_update_available("1.0.0", "1.0.1")

        # Assert
        assert result is True

    def test_is_update_available_major_version_greater(self):
        """Test _is_update_available with major version greater."""
        # Arrange
        checker = UpdateChecker()

        # Act
        result = checker._is_update_available("1.0.0", "2.0.0")

        # Assert
        assert result is True


class TestUpdateCheckerIntegration:
    """Test UpdateChecker integration scenarios."""

    def test_check_for_update_caches_by_version(self):
        """Test that check_for_update caches by version."""
        # Arrange
        checker = UpdateChecker()

        mock_response = MagicMock()
        mock_response.read.return_value.decode.return_value = '{"info": {"version": "1.2.0"}}'

        with patch("urllib.request.urlopen", return_value=mock_response):
            # Act - First call
            result1 = checker.check_for_update("1.0.0")

            # Second call with same version should use cache
            result2 = checker.check_for_update("1.0.0")

        # Assert - both should return same result
        assert result1.available == result2.available
        assert result1.latest_version == result2.latest_version

    def test_check_for_update_different_version_bypasses_cache(self):
        """Test that different version bypasses cache."""
        # Arrange
        checker = UpdateChecker()

        mock_response = MagicMock()
        mock_response.read.return_value.decode.return_value = '{"info": {"version": "1.2.0"}}'

        call_count = 0

        def mock_urlopen(*args, **kwargs):
            nonlocal call_count
            call_count += 1
            return mock_response

        with patch("urllib.request.urlopen", side_effect=mock_urlopen):
            # Act - First call
            result1 = checker.check_for_update("1.0.0")

            # Second call with different version
            result2 = checker.check_for_update("2.0.0")

        # Assert - should make two network calls
        assert call_count == 2

    def test_check_for_update_timeout(self):
        """Test check_for_update with timeout."""
        # Arrange
        checker = UpdateChecker()

        # Assert timeout is set correctly
        assert checker._TIMEOUT_SECONDS == 5

    def test_check_for_update_cache_ttl(self):
        """Test cache TTL is set correctly."""
        # Arrange
        checker = UpdateChecker()

        # Assert cache TTL is 300 seconds (5 minutes)
        assert checker._CACHE_TTL_SECONDS == 300

    def test_check_for_update_pypi_url(self):
        """Test PyPI URL is set correctly."""
        # Arrange
        checker = UpdateChecker()

        # Assert PyPI URL is correct
        assert checker._PYPI_API_URL == "https://pypi.org/pypi/moai-adk/json"

    def test_check_for_update_handles_zero_version(self):
        """Test handling of zero version components."""
        # Arrange
        checker = UpdateChecker()

        # Act
        result = checker._is_update_available("0.0.0", "0.0.1")

        # Assert
        assert result is True

    def test_check_for_update_handles_many_zeros(self):
        """Test handling of version with many zeros."""
        # Arrange
        checker = UpdateChecker()

        # Act
        result = checker._is_update_available("0.0.0.0", "0.0.0.1")

        # Assert
        assert result is True


class TestUpdateInfoDataclass:
    """Test UpdateInfo dataclass."""

    def test_update_info_creation(self):
        """Test UpdateInfo creation."""
        # Act
        info = UpdateInfo(available=True, latest_version="1.2.0")

        # Assert
        assert info.available is True
        assert info.latest_version == "1.2.0"

    def test_update_info_with_none_version(self):
        """Test UpdateInfo with None latest_version."""
        # Act
        info = UpdateInfo(available=False, latest_version=None)

        # Assert
        assert info.available is False
        assert info.latest_version is None

    def test_update_info_equality(self):
        """Test UpdateInfo equality."""
        # Arrange
        info1 = UpdateInfo(available=True, latest_version="1.2.0")
        info2 = UpdateInfo(available=True, latest_version="1.2.0")
        info3 = UpdateInfo(available=False, latest_version="1.2.0")

        # Assert
        assert info1 == info2
        assert info1 != info3


class TestEdgeCases:
    """Test edge cases and error conditions."""

    def test_fetch_latest_version_network_error(self):
        """Test fetch_latest_version with network error."""
        # Arrange
        checker = UpdateChecker()

        with patch("urllib.request.urlopen", side_effect=OSError("Network error")):
            # Act
            result = checker._fetch_latest_version("1.0.0")

        # Assert - should return default UpdateInfo
        assert result.available is False
        assert result.latest_version is None

    def test_fetch_latest_version_json_decode_error(self):
        """Test fetch_latest_version with JSON decode error."""
        # Arrange
        checker = UpdateChecker()

        mock_response = MagicMock()
        mock_response.read.return_value.decode.return_value = "invalid json"

        with patch("urllib.request.urlopen", return_value=mock_response):
            # Act
            result = checker._fetch_latest_version("1.0.0")

        # Assert - should return default UpdateInfo
        assert result.available is False
        assert result.latest_version is None

    def test_fetch_latest_version_missing_info_key(self):
        """Test fetch_latest_version when info key is missing."""
        # Arrange
        checker = UpdateChecker()

        mock_response = MagicMock()
        mock_response.read.return_value.decode.return_value = '{"data": "value"}'

        with patch("urllib.request.urlopen", return_value=mock_response):
            # Act
            result = checker._fetch_latest_version("1.0.0")

        # Assert - should return default UpdateInfo
        assert result.available is False
        assert result.latest_version is None

    def test_fetch_latest_version_missing_version_key(self):
        """Test fetch_latest_version when version key is missing."""
        # Arrange
        checker = UpdateChecker()

        mock_response = MagicMock()
        mock_response.read.return_value.decode.return_value = '{"info": {"name": "moai-adk"}}'

        with patch("urllib.request.urlopen", return_value=mock_response):
            # Act
            result = checker._fetch_latest_version("1.0.0")

        # Assert - should return UpdateInfo with available=False
        assert result.available is False
        assert result.latest_version is None

    def test_fetch_latest_version_empty_version(self):
        """Test fetch_latest_version when version is empty string."""
        # Arrange
        checker = UpdateChecker()

        mock_response = MagicMock()
        mock_response.read.return_value.decode.return_value = '{"info": {"version": ""}}'

        with patch("urllib.request.urlopen", return_value=mock_response):
            # Act
            result = checker._fetch_latest_version("1.0.0")

        # Assert
        assert result.available is False
        assert result.latest_version is None

    def test_is_update_available_unicode_versions(self):
        """Test _is_update_available with Unicode characters."""
        # Arrange
        checker = UpdateChecker()

        # Act
        result = checker._is_update_available("1.0.0-beta🚀", "1.0.0")

        # Assert - should handle Unicode gracefully
        # Regex will extract [1, 0, 0]
        assert result is False

    def test_is_cache_valid_with_none_cache(self):
        """Test _is_cache_valid when cache is None."""
        # Arrange
        checker = UpdateChecker()
        checker._cached_info = None
        checker._cache_time = None

        # Act
        result = checker._is_cache_valid()

        # Assert
        assert result is False

    def test_update_cache_with_stores_correctly(self):
        """Test _update_cache_with stores version info."""
        # Arrange
        checker = UpdateChecker()
        update_info = UpdateInfo(available=True, latest_version="1.2.0")

        # Act
        checker._update_cache_with(update_info, "1.0.0")

        # Assert
        assert checker._cached_info == update_info
        assert checker._cached_version == "1.0.0"
        assert checker._cache_time is not None
