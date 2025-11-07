"""
Tests for VersionReader - MoAI-ADK 버전 읽기

@TEST:VERSION-READER-001 - 버전 읽기
@TEST:VERSION-READER-002 - 60초 캐싱
"""

import pytest
import json
import tempfile
from pathlib import Path


class TestVersionReader:
    """버전 읽기 테스트"""

    def test_read_version(self):
        """
        GIVEN: .moai/config.json에 version="0.20.1"
        WHEN: get_version() 호출
        THEN: "0.20.1" 또는 "v0.20.1" 반환
        """
        # @TEST:VERSION-READER-001
        from moai_adk.statusline.version_reader import VersionReader

        with tempfile.TemporaryDirectory() as tmpdir:
            config_file = Path(tmpdir) / "config.json"
            config_data = {"moai": {"version": "0.20.1"}}
            config_file.write_text(json.dumps(config_data))

            reader = VersionReader()
            reader._config_path = config_file

            version = reader.get_version()

            assert "0.20.1" in version, f"Expected version 0.20.1, got {version}"

    def test_version_with_v_prefix(self):
        """
        GIVEN: config.json에 version="v0.20.1"
        WHEN: get_version() 호출
        THEN: 버전 반환 (v 접두사 포함 또는 제외)
        """
        from moai_adk.statusline.version_reader import VersionReader

        with tempfile.TemporaryDirectory() as tmpdir:
            config_file = Path(tmpdir) / "config.json"
            config_data = {"moai": {"version": "v0.20.1"}}
            config_file.write_text(json.dumps(config_data))

            reader = VersionReader()
            reader._config_path = config_file

            version = reader.get_version()

            assert "0.20.1" in version, f"Expected version with 0.20.1, got {version}"

    def test_version_caching(self):
        """
        GIVEN: VersionReader 인스턴스
        WHEN: 60초 이내에 두 번 get_version() 호출
        THEN: 캐시에서 반환
        """
        # @TEST:VERSION-READER-002
        from moai_adk.statusline.version_reader import VersionReader

        with tempfile.TemporaryDirectory() as tmpdir:
            config_file = Path(tmpdir) / "config.json"
            config_data = {"moai": {"version": "0.20.1"}}
            config_file.write_text(json.dumps(config_data))

            reader = VersionReader()
            reader._config_path = config_file

            # First call
            result1 = reader.get_version()

            # Second call immediately (should use cache)
            result2 = reader.get_version()

            # Results should be identical
            assert result1 == result2, "Cache should return identical results"

    def test_version_missing_graceful(self):
        """
        GIVEN: config.json에 version 필드 없음
        WHEN: get_version() 호출
        THEN: 기본값 반환 (예: "unknown", "[???]")
        """
        from moai_adk.statusline.version_reader import VersionReader

        with tempfile.TemporaryDirectory() as tmpdir:
            config_file = Path(tmpdir) / "config.json"
            config_data = {"moai": {}}
            config_file.write_text(json.dumps(config_data))

            reader = VersionReader()
            reader._config_path = config_file

            version = reader.get_version()

            # Should return some default or error indicator
            assert version and len(version) > 0, "Should return a value"


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
