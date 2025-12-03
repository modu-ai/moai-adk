"""
Tests for UpdateChecker - 업데이트 확인

"""

from unittest.mock import MagicMock, patch

import pytest


class TestUpdateChecker:
    """업데이트 확인 테스트"""

    def test_check_update_available(self):
        """
        GIVEN: 현재 버전 0.20.1, PyPI 최신 0.21.0
        WHEN: check_for_update("0.20.1") 호출
        THEN: available=True, latest_version="0.21.0"
        """
        from moai_adk.statusline.update_checker import UpdateChecker

        checker = UpdateChecker()

        with patch("urllib.request.urlopen") as mock_urlopen:
            # Mock PyPI API response
            mock_response = MagicMock()
            mock_response.read.return_value = b'{"info": {"version": "0.21.0"}}'
            mock_response.__enter__.return_value = mock_response
            mock_urlopen.return_value = mock_response

            update_info = checker.check_for_update("0.20.1")

            assert update_info.available is True, "Expected available=True"
            assert "0.21.0" in str(update_info.latest_version), "Expected latest_version=0.21.0"

    def test_check_update_not_available(self):
        """
        GIVEN: 현재 버전 0.21.0, PyPI 최신 0.21.0
        WHEN: check_for_update("0.21.0") 호출
        THEN: available=False
        """
        from moai_adk.statusline.update_checker import UpdateChecker

        checker = UpdateChecker()

        with patch("urllib.request.urlopen") as mock_urlopen:
            mock_response = MagicMock()
            mock_response.read.return_value = b'{"info": {"version": "0.21.0"}}'
            mock_response.__enter__.return_value = mock_response
            mock_urlopen.return_value = mock_response

            update_info = checker.check_for_update("0.21.0")

            assert update_info.available is False, "Expected available=False when versions match"

    def test_api_failure_graceful(self):
        """
        GIVEN: PyPI API 호출 실패
        WHEN: check_for_update() 호출
        THEN: available=False (에러 무시)
        """
        from moai_adk.statusline.update_checker import UpdateChecker

        checker = UpdateChecker()

        with patch("urllib.request.urlopen") as mock_urlopen:
            mock_urlopen.side_effect = Exception("API call failed")

            update_info = checker.check_for_update("0.20.1")

            # Should gracefully handle error
            assert update_info.available is False, "Should gracefully handle API failure"

    def test_update_checker_caching(self):
        """
        GIVEN: UpdateChecker 인스턴스
        WHEN: 300초 이내에 두 번 check_for_update() 호출
        THEN: 캐시에서 반환
        """
        from moai_adk.statusline.update_checker import UpdateChecker

        checker = UpdateChecker()

        with patch("urllib.request.urlopen") as mock_urlopen:
            mock_response = MagicMock()
            mock_response.read.return_value = b'{"info": {"version": "0.21.0"}}'
            mock_response.__enter__.return_value = mock_response
            mock_urlopen.return_value = mock_response

            # First call
            checker.check_for_update("0.20.1")
            call_count_after_first = mock_urlopen.call_count

            # Second call immediately (should use cache)
            checker.check_for_update("0.20.1")
            call_count_after_second = mock_urlopen.call_count

            # API should only be called once (cache works)
            assert (
                call_count_after_first == call_count_after_second
            ), f"API called {call_count_after_second - call_count_after_first} more times (cache not working)"


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
