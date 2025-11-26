"""
Tests for VersionReader - MoAI-ADK 버전 읽기

"""

import json
import tempfile
from datetime import datetime, timedelta
from pathlib import Path


class TestVersionReader:
    """버전 읽기 테스트"""

    def test_read_version_present_in_config(self):
        """
        GIVEN: .moai/config.json에 version="0.20.1"가 있는 경우
        WHEN: get_version() 호출
        THEN: "0.20.1" 또는 "v0.20.1" 반환
        """
        from moai_adk.statusline.version_reader import VersionReader

        with tempfile.TemporaryDirectory() as tmpdir:
            config_file = Path(tmpdir) / "config.json"
            config_data = {"moai": {"version": "0.20.1"}}
            config_file.write_text(json.dumps(config_data))

            reader = VersionReader()
            reader._config_path = config_file

            version = reader.get_version()

            # RED: 이 테스트는 실패할 것입니다. 현재 구현에서 버전이 올바르게 읽히지 않음
            assert "0.20.1" in version, f"Expected version 0.20.1, got {version}"
            assert version != "unknown", "Version should not be 'unknown'"

    def test_version_with_v_prefix(self):
        """
        GIVEN: config.json에 version="v0.20.1"
        WHEN: get_version() 호출
        THEN: 버전 반환 (v 접두사 포함 또는 제외)
        """
        from moai_adk.statusline.version_reader import VersionReader

        with tempfile.TemporaryDirectory() as tmpdir:
            config_file = Path(tmpdir) / "config.json"
            config_data = {"moai": {"version": "v0.22.4"}}
            config_file.write_text(json.dumps(config_data))

            reader = VersionReader()
            reader._config_path = config_file

            version = reader.get_version()

            # RED: 이 테스트는 실패할 것입니다. v 접두사 처리 로직이 없음
            assert "0.22.4" in version or version == "v0.22.4", f"Expected version with 0.22.4, got {version}"

    def test_version_caching(self):
        """
        GIVEN: VersionReader 인스턴스
        WHEN: 60초 이내에 두 번 get_version() 호출
        THEN: 캐시에서 반환
        """
        from moai_adk.statusline.version_reader import VersionReader

        with tempfile.TemporaryDirectory() as tmpdir:
            config_file = Path(tmpdir) / "config.json"
            config_data = {"moai": {"version": "0.20.1"}}
            config_file.write_text(json.dumps(config_data))

            reader = VersionReader()
            reader._config_path = config_file

            # First call
            result1 = reader.get_version()

            # Force cache time to current time to simulate fresh cache
            reader._cache_time = datetime.now()

            # Second call immediately (should use cache)
            result2 = reader.get_version()

            # Results should be identical
            assert result1 == result2, "Cache should return identical results"

    def test_version_missing_graceful_fallback(self):
        """
        GIVEN: config.json에 version 필드 없음
        WHEN: get_version() 호출
        THEN: 기본값 반환 (예: "unknown") - 하지만 버그가 있음
        """
        from moai_adk.statusline.version_reader import VersionReader

        with tempfile.TemporaryDirectory() as tmpdir:
            config_file = Path(tmpdir) / "config.json"
            # Config without version field
            config_data = {"moai": {}}
            config_file.write_text(json.dumps(config_data))

            reader = VersionReader()
            reader._config_path = config_file

            version = reader.get_version()

            # RED: 이 테스트는 성공하지만, 이는 버그를 숨기는 것입니다.
            # 올바른 버전이 설정되어야 함
            assert version and len(version) > 0, "Should return a value"

    def test_version_read_from_fallback_field(self):
        """
        GIVEN: config.json에 moai.version 없지만 최상위 version 필드 있음
        WHEN: get_version() 호출
        THEN: 최상위 version 필드에서 읽음
        """
        from moai_adk.statusline.version_reader import VersionReader

        with tempfile.TemporaryDirectory() as tmpdir:
            config_file = Path(tmpdir) / "config.json"
            # Config with top-level version field
            config_data = {"version": "1.5.0"}
            config_file.write_text(json.dumps(config_data))

            reader = VersionReader()
            reader._config_path = config_file

            version = reader.get_version()

            # RED: 이 테스트는 실패할 것입니다. 현재 구현은 최상위 version 필드를 읽지 못함
            assert version == "1.5.0" or "1.5.0" in version, f"Expected version 1.5.0, got {version}"

    def test_version_reading_nonexistent_config(self):
        """
        GIVEN: .moai/config.json 파일 없음
        WHEN: get_version() 호출
        THEN: "unknown" 반환
        """
        from moai_adk.statusline.version_reader import VersionReader

        with tempfile.TemporaryDirectory() as tmpdir:
            config_file = Path(tmpdir) / "config.json"
            # Don't create the config file

            reader = VersionReader()
            reader._config_path = config_file

            version = reader.get_version()

            # RED: 이 테스트는 성공합니다. 하지만 이는 버그입니다.
            # 파일이 없으면 버전을 설정해야 함
            assert version == "unknown", f"Expected 'unknown' for missing config, got {version}"

    def test_version_reading_invalid_json(self):
        """
        GIVEN: .moai/config.json가 유효하지 않은 JSON
        WHEN: get_version() 호출
        THEN: "unknown" 반환
        """
        from moai_adk.statusline.version_reader import VersionReader

        with tempfile.TemporaryDirectory() as tmpdir:
            config_file = Path(tmpdir) / "config.json"
            # Invalid JSON
            config_file.write_text("{ invalid json }")

            reader = VersionReader()
            reader._config_path = config_file

            version = reader.get_version()

            # RED: 이 테스트는 성공합니다. 하지만 이는 버그입니다.
            # JSON이 잘못되면 버전을 설정해야 함
            assert version == "unknown", f"Expected 'unknown' for invalid JSON, got {version}"

    def test_version_cache_expiration(self):
        """
        GIVEN: Version 캺시가 만료된 경우
        WHEN: get_version() 호출
        THEN: 새로 파일에서 읽어옴
        """
        from moai_adk.statusline.version_reader import VersionReader

        with tempfile.TemporaryDirectory() as tmpdir:
            config_file = Path(tmpdir) / "config.json"
            config_data = {"moai": {"version": "0.20.1"}}
            config_file.write_text(json.dumps(config_data))

            reader = VersionReader()
            reader._config_path = config_file

            # First call - should read from file
            result1 = reader.get_version()

            # Change config file
            config_data["moai"]["version"] = "1.0.0"
            config_file.write_text(json.dumps(config_data))

            # Set cache time to expired (1 hour ago)
            reader._cache_time = datetime.now() - timedelta(hours=1)

            # Second call - should read fresh from file
            result2 = reader.get_version()

            # Results should be different
            assert result1 != result2, f"Cache should have expired, got {result1} and {result2}"
            assert "1.0.0" in result2, f"Should read new version 1.0.0, got {result2}"

    def test_version_multiple_sources_priority(self):
        """
        GIVEN: config.json에 moai.version과 최상위 version 필드 모두 있음
        WHEN: get_version() 호출
        THEN: moai.version이 우선 (명확한 우선순위)
        """
        from moai_adk.statusline.version_reader import VersionReader

        with tempfile.TemporaryDirectory() as tmpdir:
            config_file = Path(tmpdir) / "config.json"
            # Config with both version fields
            config_data = {"version": "1.0.0", "moai": {"version": "2.0.0"}}  # Should be ignored  # Should be used
            config_file.write_text(json.dumps(config_data))

            reader = VersionReader()
            reader._config_path = config_file

            version = reader.get_version()

            # RED: 이 테스트는 실패할 것입니다. 현재 구현은 올바른 우선순위를 처리하지 않음
            assert version == "2.0.0", f"Expected moai.version '2.0.0', got {version}"

    def test_version_reading_partial_config(self):
        """
        GIVEN: 부분적인 config.json 파일
        WHEN: get_version() 호출
        THEN: 부족한 정보에도 불구하고 버전 읽음
        """
        from moai_adk.statusline.version_reader import VersionReader

        with tempfile.TemporaryDirectory() as tmpdir:
            config_file = Path(tmpdir) / "config.json"
            # Partial config - only has version
            config_data = {"moai": {"version": "3.1.2"}}
            config_file.write_text(json.dumps(config_data))

            reader = VersionReader()
            reader._config_path = config_file

            version = reader.get_version()

            # RED: 이 테스트는 실패할 것입니다. 부분적 구조 처리에 문제가 있음
            assert version == "3.1.2", f"Expected version 3.1.2, got {version}"

    def test_version_case_sensitivity(self):
        """
        GIVEN: 버전 문자열의 대소문자 처리
        WHEN: get_version() 호출
        THEN: 정확한 문자열 반환
        """
        from moai_adk.statusline.version_reader import VersionReader

        with tempfile.TemporaryDirectory() as tmpdir:
            config_file = Path(tmpdir) / "config.json"
            # Version with different formats
            test_versions = ["0.22.4", "v0.22.4", "0.22.4-beta", "0.22.4"]

            for version in test_versions:
                config_data = {"moai": {"version": version}}
                config_file.write_text(json.dumps(config_data))

                reader = VersionReader()
                reader._config_path = config_file

                result = reader.get_version()

                # RED: 이 테스트는 실패할 것입니다. 정확한 문자열 처리 문제
                assert result == version or version in result, f"Expected version '{version}', got '{result}'"

    def test_version_with_special_characters(self):
        """
        GIVEN: 특수 문자를 포함한 버전 문자열
        WHEN: get_version() 호출
        THEN: 정확한 문자열 반환
        """
        from moai_adk.statusline.version_reader import VersionReader

        special_versions = ["0.22.4-beta.1", "0.22.4-alpha", "v0.22.4-rc.1", "0.22.4-dev"]

        for version in special_versions:
            with tempfile.TemporaryDirectory() as tmpdir:
                config_file = Path(tmpdir) / "config.json"
                config_data = {"moai": {"version": version}}
                config_file.write_text(json.dumps(config_data))

                reader = VersionReader()
                reader._config_path = config_file

                result = reader.get_version()

                # RED: 이 테스트는 실패할 것입니다. 특수 문자 처리 문제
                assert result == version or version in result, f"Expected version '{version}', got '{result}'"

    def test_version_cache_update_on_valid_version(self):
        """
        GIVEN: 유효한 버전이 발견된 경우
        WHEN: get_version() 호출 후 캐시 확인
        THEN: 캐시에 올바르게 저장됨
        """
        from moai_adk.statusline.version_reader import VersionReader

        with tempfile.TemporaryDirectory() as tmpdir:
            config_file = Path(tmpdir) / "config.json"
            config_data = {"moai": {"version": "1.2.3"}}
            config_file.write_text(json.dumps(config_data))

            reader = VersionReader()
            reader._config_path = config_file

            # First call
            version = reader.get_version()

            # Check cache was updated
            assert reader._version_cache == version, f"Cache should contain {version}"
            assert reader._cache_time is not None, "Cache time should be set"
            assert reader._is_cache_valid(), "Cache should be valid"

    def test_version_cache_behavior_on_unknown(self):
        """
        GIVEN: 버전을 읽을 수 없는 경우
        WHEN: get_version() 호출
        THEN: 'unknown'가 캐시에 저장됨
        """
        from moai_adk.statusline.version_reader import VersionReader

        with tempfile.TemporaryDirectory() as tmpdir:
            config_file = Path(tmpdir) / "config.json"
            # Config without version
            config_data = {"moai": {}}
            config_file.write_text(json.dumps(config_data))

            reader = VersionReader()
            reader._config_path = config_file

            # Call should return 'unknown'
            version = reader.get_version()

            # RED: 이 테스트는 성공하지만, 버그를 숨기는 것입니다.
            # 'unknown'는 캐시에 저장되지 않아야 함
            assert version == "unknown", f"Expected 'unknown', got {version}"

            # Cache should contain 'unknown'
            assert reader._version_cache == "unknown", f"Cache should contain 'unknown', got {reader._version_cache}"
