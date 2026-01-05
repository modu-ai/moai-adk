"""Comprehensive tests for UpdateChecker with 80% coverage target."""

import json
import urllib.error
import urllib.request
from datetime import datetime, timedelta
from unittest.mock import MagicMock, patch

from moai_adk.statusline.update_checker import UpdateChecker, UpdateInfo


class TestUpdateCheckerInit:
    """Test UpdateChecker initialization."""

    def test_init_default(self):
        """Test default initialization."""
        checker = UpdateChecker()
        assert checker._cached_info is None
        assert checker._cache_time is None
        assert checker._cached_version is None
        assert isinstance(checker._cache_ttl, timedelta)

    def test_init_cache_ttl(self):
        """Test cache TTL value."""
        checker = UpdateChecker()
        assert checker._CACHE_TTL_SECONDS == 300
        assert checker._cache_ttl.total_seconds() == 300

    def test_init_pypi_api_url(self):
        """Test PyPI API URL."""
        checker = UpdateChecker()
        assert checker._PYPI_API_URL == "https://pypi.org/pypi/moai-adk/json"

    def test_init_timeout(self):
        """Test timeout setting."""
        checker = UpdateChecker()
        assert checker._TIMEOUT_SECONDS == 5


class TestUpdateCheckerCheckForUpdate:
    """Test main check_for_update method."""

    def test_check_for_update_update_available(self):
        """Test detection of available update."""
        checker = UpdateChecker()
        json_response = {"info": {"version": "0.21.0"}}

        with patch("urllib.request.urlopen") as mock_urlopen:
            mock_response = MagicMock()
            mock_response.read.return_value = json.dumps(json_response).encode("utf-8")
            mock_response.__enter__.return_value = mock_response
            mock_urlopen.return_value = mock_response

            result = checker.check_for_update("0.20.0")
            assert result.available is True
            assert result.latest_version == "0.21.0"

    def test_check_for_update_no_update(self):
        """Test when no update is available."""
        checker = UpdateChecker()
        json_response = {"info": {"version": "0.20.0"}}

        with patch("urllib.request.urlopen") as mock_urlopen:
            mock_response = MagicMock()
            mock_response.read.return_value = json.dumps(json_response).encode("utf-8")
            mock_response.__enter__.return_value = mock_response
            mock_urlopen.return_value = mock_response

            result = checker.check_for_update("0.20.0")
            assert result.available is False

    def test_check_for_update_cache_hit(self):
        """Test cache hit for same version."""
        checker = UpdateChecker()
        cached_info = UpdateInfo(available=True, latest_version="0.21.0")
        checker._cached_info = cached_info
        checker._cache_time = datetime.now()
        checker._cached_version = "0.20.0"

        result = checker.check_for_update("0.20.0")
        assert result == cached_info

    def test_check_for_update_cache_different_version(self):
        """Test cache miss when version changes."""
        checker = UpdateChecker()
        old_info = UpdateInfo(available=True, latest_version="0.21.0")
        checker._cached_info = old_info
        checker._cache_time = datetime.now()
        checker._cached_version = "0.20.0"

        json_response = {"info": {"version": "0.22.0"}}

        with patch("urllib.request.urlopen") as mock_urlopen:
            mock_response = MagicMock()
            mock_response.read.return_value = json.dumps(json_response).encode("utf-8")
            mock_response.__enter__.return_value = mock_response
            mock_urlopen.return_value = mock_response

            result = checker.check_for_update("0.21.0")
            assert result.available is True
            assert result.latest_version == "0.22.0"


class TestUpdateCheckerFetchLatestVersion:
    """Test _fetch_latest_version method."""

    def test_fetch_latest_version_success(self):
        """Test successful version fetch from PyPI."""
        checker = UpdateChecker()
        json_response = {"info": {"version": "0.21.0"}}

        with patch("urllib.request.urlopen") as mock_urlopen:
            mock_response = MagicMock()
            mock_response.read.return_value = json.dumps(json_response).encode("utf-8")
            mock_response.__enter__.return_value = mock_response
            mock_urlopen.return_value = mock_response

            result = checker._fetch_latest_version("0.20.0")
            assert result.available is True
            assert result.latest_version == "0.21.0"

    def test_fetch_latest_version_no_info(self):
        """Test handling when no info section in response."""
        checker = UpdateChecker()
        json_response = {}

        with patch("urllib.request.urlopen") as mock_urlopen:
            mock_response = MagicMock()
            mock_response.read.return_value = json.dumps(json_response).encode("utf-8")
            mock_response.__enter__.return_value = mock_response
            mock_urlopen.return_value = mock_response

            result = checker._fetch_latest_version("0.20.0")
            assert result.available is False

    def test_fetch_latest_version_no_version_field(self):
        """Test handling when version field is missing."""
        checker = UpdateChecker()
        json_response = {"info": {}}

        with patch("urllib.request.urlopen") as mock_urlopen:
            mock_response = MagicMock()
            mock_response.read.return_value = json.dumps(json_response).encode("utf-8")
            mock_response.__enter__.return_value = mock_response
            mock_urlopen.return_value = mock_response

            result = checker._fetch_latest_version("0.20.0")
            assert result.available is False

    def test_fetch_latest_version_network_error(self):
        """Test handling of network errors."""
        checker = UpdateChecker()

        with patch("urllib.request.urlopen", side_effect=urllib.error.URLError("Network error")):
            result = checker._fetch_latest_version("0.20.0")
            assert result.available is False

    def test_fetch_latest_version_http_error(self):
        """Test handling of HTTP errors."""
        checker = UpdateChecker()

        with patch(
            "urllib.request.urlopen",
            side_effect=urllib.error.HTTPError(None, 404, "Not found", {}, None),
        ):
            result = checker._fetch_latest_version("0.20.0")
            assert result.available is False

    def test_fetch_latest_version_json_error(self):
        """Test handling of JSON parsing errors."""
        checker = UpdateChecker()

        with patch("urllib.request.urlopen") as mock_urlopen:
            mock_response = MagicMock()
            mock_response.read.return_value = b"invalid json {{"
            mock_response.__enter__.return_value = mock_response
            mock_urlopen.return_value = mock_response

            result = checker._fetch_latest_version("0.20.0")
            assert result.available is False

    def test_fetch_latest_version_timeout(self):
        """Test handling of timeout."""
        checker = UpdateChecker()

        with patch("urllib.request.urlopen", side_effect=Exception("Timeout")):
            result = checker._fetch_latest_version("0.20.0")
            assert result.available is False

    def test_fetch_latest_version_correct_url(self):
        """Test that correct PyPI URL is used."""
        checker = UpdateChecker()
        json_response = {"info": {"version": "0.21.0"}}

        with patch("urllib.request.urlopen") as mock_urlopen:
            mock_response = MagicMock()
            mock_response.read.return_value = json.dumps(json_response).encode("utf-8")
            mock_response.__enter__.return_value = mock_response
            mock_urlopen.return_value = mock_response

            checker._fetch_latest_version("0.20.0")
            mock_urlopen.assert_called_once()
            call_args = mock_urlopen.call_args
            assert call_args[0][0] == "https://pypi.org/pypi/moai-adk/json"

    def test_fetch_latest_version_timeout_param(self):
        """Test that correct timeout is passed."""
        checker = UpdateChecker()
        json_response = {"info": {"version": "0.21.0"}}

        with patch("urllib.request.urlopen") as mock_urlopen:
            mock_response = MagicMock()
            mock_response.read.return_value = json.dumps(json_response).encode("utf-8")
            mock_response.__enter__.return_value = mock_response
            mock_urlopen.return_value = mock_response

            checker._fetch_latest_version("0.20.0")
            call_args = mock_urlopen.call_args
            assert call_args.kwargs["timeout"] == 5


class TestUpdateCheckerVersionComparison:
    """Test _is_update_available static method."""

    def test_is_update_available_patch_version(self):
        """Test patch version update detection."""
        assert UpdateChecker._is_update_available("0.20.0", "0.20.1") is True
        assert UpdateChecker._is_update_available("0.20.1", "0.20.0") is False

    def test_is_update_available_minor_version(self):
        """Test minor version update detection."""
        assert UpdateChecker._is_update_available("0.20.0", "0.21.0") is True
        assert UpdateChecker._is_update_available("0.21.0", "0.20.0") is False

    def test_is_update_available_major_version(self):
        """Test major version update detection."""
        assert UpdateChecker._is_update_available("0.20.0", "1.0.0") is True
        assert UpdateChecker._is_update_available("1.0.0", "0.20.0") is False

    def test_is_update_available_same_version(self):
        """Test same version detection."""
        assert UpdateChecker._is_update_available("0.20.0", "0.20.0") is False

    def test_is_update_available_with_v_prefix(self):
        """Test version comparison with v prefix."""
        assert UpdateChecker._is_update_available("v0.20.0", "v0.21.0") is True
        assert UpdateChecker._is_update_available("v0.20.0", "0.21.0") is True
        assert UpdateChecker._is_update_available("0.20.0", "v0.21.0") is True

    def test_is_update_available_prerelease(self):
        """Test prerelease version comparison."""
        assert UpdateChecker._is_update_available("0.20.0", "0.21.0-alpha") is True
        assert UpdateChecker._is_update_available("0.20.0-beta", "0.21.0") is True

    def test_is_update_available_complex_versions(self):
        """Test complex version comparison."""
        assert UpdateChecker._is_update_available("0.20.1", "0.21.0") is True
        assert UpdateChecker._is_update_available("1.0.0", "1.0.1") is True
        assert UpdateChecker._is_update_available("1.5.2", "1.5.3") is True

    def test_is_update_available_invalid_format(self):
        """Test handling of invalid version format."""
        result = UpdateChecker._is_update_available("invalid", "0.21.0")
        # Should handle gracefully
        assert isinstance(result, bool)

    def test_is_update_available_empty_versions(self):
        """Test handling of empty version strings."""
        result = UpdateChecker._is_update_available("", "0.21.0")
        assert isinstance(result, bool)

    def test_is_update_available_single_digit_versions(self):
        """Test version strings with different digit counts."""
        assert UpdateChecker._is_update_available("0.20", "0.21") is True
        assert UpdateChecker._is_update_available("20", "21") is True


class TestUpdateCheckerCache:
    """Test caching behavior."""

    def test_is_cache_valid_with_valid_cache(self):
        """Test cache validation with valid cache."""
        checker = UpdateChecker()
        checker._cached_info = UpdateInfo(available=True, latest_version="0.21.0")
        checker._cache_time = datetime.now()
        assert checker._is_cache_valid()

    def test_is_cache_valid_with_expired_cache(self):
        """Test cache validation with expired cache."""
        checker = UpdateChecker()
        checker._cached_info = UpdateInfo(available=True, latest_version="0.21.0")
        checker._cache_time = datetime.now() - timedelta(seconds=400)
        assert not checker._is_cache_valid()

    def test_is_cache_valid_with_no_cache(self):
        """Test cache validation with no cached info."""
        checker = UpdateChecker()
        checker._cached_info = None
        assert not checker._is_cache_valid()

    def test_update_cache_with(self):
        """Test cache update."""
        checker = UpdateChecker()
        update_info = UpdateInfo(available=True, latest_version="0.21.0")
        checker._update_cache_with(update_info, "0.20.0")

        assert checker._cached_info == update_info
        assert checker._cached_version == "0.20.0"
        assert checker._cache_time is not None


class TestUpdateCheckerUpdateInfo:
    """Test UpdateInfo dataclass."""

    def test_updateinfo_creation(self):
        """Test UpdateInfo creation."""
        info = UpdateInfo(available=True, latest_version="0.21.0")
        assert info.available is True
        assert info.latest_version == "0.21.0"

    def test_updateinfo_not_available(self):
        """Test UpdateInfo when update not available."""
        info = UpdateInfo(available=False, latest_version=None)
        assert info.available is False
        assert info.latest_version is None

    def test_updateinfo_equality(self):
        """Test UpdateInfo equality."""
        info1 = UpdateInfo(available=True, latest_version="0.21.0")
        info2 = UpdateInfo(available=True, latest_version="0.21.0")
        assert info1 == info2

    def test_updateinfo_inequality(self):
        """Test UpdateInfo inequality."""
        info1 = UpdateInfo(available=True, latest_version="0.21.0")
        info2 = UpdateInfo(available=False, latest_version=None)
        assert info1 != info2


class TestUpdateCheckerIntegration:
    """Integration tests for UpdateChecker."""

    def test_full_update_check_flow(self):
        """Test complete update check flow."""
        checker = UpdateChecker()
        json_response = {"info": {"version": "0.21.0"}}

        with patch("urllib.request.urlopen") as mock_urlopen:
            mock_response = MagicMock()
            mock_response.read.return_value = json.dumps(json_response).encode("utf-8")
            mock_response.__enter__.return_value = mock_response
            mock_urlopen.return_value = mock_response

            result = checker.check_for_update("0.20.0")
            assert result.available is True
            assert result.latest_version == "0.21.0"

    def test_multiple_update_checks_with_cache(self):
        """Test multiple update checks with caching."""
        checker = UpdateChecker()
        json_response = {"info": {"version": "0.21.0"}}

        with patch("urllib.request.urlopen") as mock_urlopen:
            mock_response = MagicMock()
            mock_response.read.return_value = json.dumps(json_response).encode("utf-8")
            mock_response.__enter__.return_value = mock_response
            mock_urlopen.return_value = mock_response

            # First check
            result1 = checker.check_for_update("0.20.0")
            call_count_1 = mock_urlopen.call_count

            # Second check with same version (should use cache)
            result2 = checker.check_for_update("0.20.0")
            call_count_2 = mock_urlopen.call_count

            assert result1 == result2
            assert call_count_1 == call_count_2


class TestUpdateCheckerEdgeCases:
    """Test edge cases and error conditions."""

    def test_version_with_special_characters(self):
        """Test version comparison with special characters."""
        result = UpdateChecker._is_update_available("0.20.0-rc1", "0.20.0")
        assert isinstance(result, bool)

    def test_version_with_build_metadata(self):
        """Test version with build metadata."""
        result = UpdateChecker._is_update_available("0.20.0+build.123", "0.21.0")
        assert isinstance(result, bool)

    def test_pypi_response_with_extra_fields(self):
        """Test PyPI response parsing with extra fields."""
        checker = UpdateChecker()
        json_response = {
            "info": {
                "version": "0.21.0",
                "author": "Test",
                "description": "Test package",
            }
        }

        with patch("urllib.request.urlopen") as mock_urlopen:
            mock_response = MagicMock()
            mock_response.read.return_value = json.dumps(json_response).encode("utf-8")
            mock_response.__enter__.return_value = mock_response
            mock_urlopen.return_value = mock_response

            result = checker._fetch_latest_version("0.20.0")
            assert result.available is True

    def test_unicode_in_version_response(self):
        """Test handling of unicode in version response."""
        checker = UpdateChecker()
        json_response = {
            "info": {"version": "0.21.0"},
            "description": "Package with unicode: 日本語",
        }

        with patch("urllib.request.urlopen") as mock_urlopen:
            mock_response = MagicMock()
            mock_response.read.return_value = json.dumps(json_response).encode("utf-8")
            mock_response.__enter__.return_value = mock_response
            mock_urlopen.return_value = mock_response

            result = checker._fetch_latest_version("0.20.0")
            assert result.available is True

    def test_concurrent_update_checks(self):
        """Test behavior with potential concurrent checks."""
        checker = UpdateChecker()
        json_response = {"info": {"version": "0.21.0"}}

        with patch("urllib.request.urlopen") as mock_urlopen:
            mock_response = MagicMock()
            mock_response.read.return_value = json.dumps(json_response).encode("utf-8")
            mock_response.__enter__.return_value = mock_response
            mock_urlopen.return_value = mock_response

            results = []
            for _ in range(3):
                result = checker.check_for_update("0.20.0")
                results.append(result)

            assert all(r.available for r in results)

    def test_very_old_version_comparison(self):
        """Test version comparison with very old versions."""
        assert UpdateChecker._is_update_available("0.1.0", "0.20.0") is True

    def test_future_version_comparison(self):
        """Test version comparison with future versions."""
        assert UpdateChecker._is_update_available("0.20.0", "1.0.0") is True
        assert UpdateChecker._is_update_available("1.0.0", "2.0.0") is True
