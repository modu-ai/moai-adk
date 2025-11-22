"""
EnvKeyManager 테스트

API 키 관리 기능 검증
"""

import os
import tempfile
from pathlib import Path
import pytest
from modules.env_key_manager import EnvKeyManager


class TestEnvKeyManager:
    """EnvKeyManager 테스트 클래스"""

    @pytest.fixture(autouse=True)
    def setup_teardown(self):
        """각 테스트 전후 설정"""
        # 기존 환경 변수 저장
        self.original_env_key = os.environ.get(EnvKeyManager.ENV_KEY_NAME)
        self.original_env_file = EnvKeyManager.ENV_FILE

        # 테스트용 임시 .env 파일
        self.test_env_file = ".env.test"
        EnvKeyManager.ENV_FILE = self.test_env_file

        yield

        # 정리
        if self.original_env_key:
            os.environ[EnvKeyManager.ENV_KEY_NAME] = self.original_env_key
        else:
            os.environ.pop(EnvKeyManager.ENV_KEY_NAME, None)

        EnvKeyManager.ENV_FILE = self.original_env_file

        # 테스트 파일 삭제
        Path(self.test_env_file).unlink(missing_ok=True)

    def test_validate_api_key_valid(self):
        """유효한 API 키 검증"""
        valid_key = "gsk_" + "a" * 50
        assert EnvKeyManager.validate_api_key(valid_key) is True

    def test_validate_api_key_invalid_prefix(self):
        """잘못된 프리픽스 키 검증"""
        invalid_key = "sk_" + "a" * 50
        assert EnvKeyManager.validate_api_key(invalid_key) is False

    def test_validate_api_key_too_short(self):
        """너무 짧은 키 검증"""
        short_key = "gsk_short"
        assert EnvKeyManager.validate_api_key(short_key) is False

    def test_validate_api_key_empty(self):
        """빈 키 검증"""
        assert EnvKeyManager.validate_api_key("") is False
        assert EnvKeyManager.validate_api_key(None) is False

    def test_set_api_key(self):
        """API 키 설정"""
        test_key = "gsk_" + "a" * 50
        result = EnvKeyManager.set_api_key(test_key)

        assert result is True
        assert Path(self.test_env_file).exists()

        # 파일 내용 확인
        with open(self.test_env_file, 'r') as f:
            content = f.read()
            assert f"{EnvKeyManager.ENV_KEY_NAME}={test_key}" in content

    def test_set_api_key_invalid(self):
        """유효하지 않은 API 키 설정"""
        invalid_key = "invalid_key"

        with pytest.raises(ValueError):
            EnvKeyManager.set_api_key(invalid_key)

    def test_get_api_key_from_env(self):
        """환경 변수에서 API 키 로드"""
        test_key = "gsk_" + "b" * 50
        os.environ[EnvKeyManager.ENV_KEY_NAME] = test_key

        result = EnvKeyManager.get_api_key()
        assert result == test_key

    def test_get_api_key_from_file(self):
        """파일에서 API 키 로드"""
        test_key = "gsk_" + "c" * 50

        # 환경 변수 제거
        os.environ.pop(EnvKeyManager.ENV_KEY_NAME, None)

        # .env 파일에 키 작성
        with open(self.test_env_file, 'w') as f:
            f.write(f"{EnvKeyManager.ENV_KEY_NAME}={test_key}\n")

        result = EnvKeyManager.get_api_key()
        assert result == test_key

    def test_get_api_key_not_configured(self):
        """API 키가 설정되지 않은 경우"""
        os.environ.pop(EnvKeyManager.ENV_KEY_NAME, None)
        Path(self.test_env_file).unlink(missing_ok=True)

        result = EnvKeyManager.get_api_key()
        assert result is None

    def test_is_configured_true(self):
        """API 키 설정 상태 확인 - True"""
        test_key = "gsk_" + "d" * 50
        EnvKeyManager.set_api_key(test_key)

        assert EnvKeyManager.is_configured() is True

    def test_is_configured_false(self):
        """API 키 설정 상태 확인 - False"""
        os.environ.pop(EnvKeyManager.ENV_KEY_NAME, None)
        Path(self.test_env_file).unlink(missing_ok=True)

        assert EnvKeyManager.is_configured() is False

    def test_update_existing_api_key(self):
        """기존 API 키 업데이트"""
        old_key = "gsk_" + "a" * 50
        new_key = "gsk_" + "b" * 50

        EnvKeyManager.set_api_key(old_key)
        EnvKeyManager.set_api_key(new_key)

        with open(self.test_env_file, 'r') as f:
            content = f.read()

        # 새 키만 포함되어야 함
        assert new_key in content
        assert old_key not in content


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
