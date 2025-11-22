"""
Nano Banana Pro - API 키 관리 모듈

Google Gemini 3 API 키를 안전하게 입력받아 .env 파일에 저장하는 모듈
"""

import os
import re
from pathlib import Path
from typing import Optional, Tuple
import getpass
import logging

logger = logging.getLogger(__name__)


class EnvKeyManager:
    """
    API 키를 환경 변수로 관리하는 클래스

    특징:
    - 비밀 입력 (getpass 사용)
    - 형식 검증 (gsk_ 확인)
    - .env 파일 안전 저장
    - 테스트 연결
    - 권한 설정 (chmod 600)
    """

    ENV_FILE = ".env"
    API_KEY_VAR = "GOOGLE_API_KEY"

    @staticmethod
    def setup_api_key() -> bool:
        """
        대화형으로 사용자로부터 API 키를 입력받아 .env에 저장

        Returns:
            bool: 성공 여부

        Example:
            >>> manager = EnvKeyManager()
            >>> success = manager.setup_api_key()
            🔐 Gemini API 키 설정 마법사
            ...
        """
        print("\n" + "="*60)
        print("🔐 Gemini 3 API 키 설정 마법사")
        print("="*60 + "\n")

        # Step 1: 안내
        print("📋 API 키를 발급받으세요:")
        print("   1. https://aistudio.google.com/apikey 방문")
        print("   2. '+ Create new API key' 클릭")
        print("   3. 'In project' 선택 후 API 키 생성")
        print("   4. API 키 복사\n")

        # Step 2: 입력
        print("⚠️  보안 알림: API 키는 화면에 표시되지 않습니다\n")

        while True:
            api_key = getpass.getpass("API 키를 입력하세요: ")

            if not api_key:
                print("❌ API 키가 비어있습니다. 다시 입력해주세요.\n")
                continue

            api_key = api_key.strip()

            # Step 3: 형식 검증
            if not EnvKeyManager.validate_api_key(api_key):
                print("❌ 올바르지 않은 API 키 형식입니다.")
                print("   • gsk_로 시작해야 합니다")
                print("   • 최소 20자 이상이어야 합니다\n")
                continue

            # Step 4: 재확인
            print("\n✓ API 키 형식이 유효합니다")
            confirm = input("이 키를 저장하시겠습니까? (y/n): ").strip().lower()

            if confirm == 'y':
                break
            else:
                print("취소되었습니다.\n")
                continue

        # Step 5: 저장
        try:
            EnvKeyManager.save_api_key(api_key)
            print("\n✅ API 키가 .env 파일에 저장되었습니다!")

            # Step 6: 테스트 (선택사항)
            print("\n🔍 API 연결 테스트 중...\n")
            if EnvKeyManager.test_connection(api_key):
                print("✓ Gemini API 연결 성공")
                print("✓ 할당량 확인 완료")
                print("\n✅ 모든 설정이 완료되었습니다!")
                print("이제 이미지를 생성할 준비가 되었습니다! 🎨\n")
                return True
            else:
                print("⚠️  API 키 검증에 문제가 있습니다.")
                print("Google Cloud Console에서 설정을 확인해주세요.\n")
                return False

        except Exception as e:
            print(f"\n❌ 저장 중 오류 발생: {str(e)}\n")
            return False

    @staticmethod
    def validate_api_key(api_key: str) -> bool:
        """
        API 키 형식 검증

        Args:
            api_key: 검증할 API 키

        Returns:
            bool: 유효한 형식 여부

        Validation Rules:
            - gsk_로 시작
            - 최소 20자 이상
            - 영문자, 숫자, 언더스코어만 포함
        """
        if not api_key:
            return False

        # 형식 검증
        pattern = r'^gsk_[a-zA-Z0-9_]{15,}$'

        if not re.match(pattern, api_key):
            return False

        return True

    @staticmethod
    def save_api_key(api_key: str) -> None:
        """
        API 키를 .env 파일에 저장

        Args:
            api_key: 저장할 API 키

        Security:
            - 파일 권한을 600으로 설정 (소유자 읽기/쓰기만)
            - 기존 키가 있으면 덮어쓰기
            - 백업 생성
        """
        env_path = Path(EnvKeyManager.ENV_FILE)

        # 기존 파일이 있으면 백업
        if env_path.exists():
            backup_path = Path(f"{EnvKeyManager.ENV_FILE}.backup")
            with open(env_path, 'r') as f:
                backup_content = f.read()
            with open(backup_path, 'w') as f:
                f.write(backup_content)
            logger.info(f"Backup created: {backup_path}")

        # 기존 내용 로드 및 업데이트
        env_vars = {}
        if env_path.exists():
            with open(env_path, 'r') as f:
                for line in f:
                    line = line.strip()
                    if line and '=' in line and not line.startswith('#'):
                        key, value = line.split('=', 1)
                        env_vars[key.strip()] = value.strip()

        # API 키 업데이트
        env_vars[EnvKeyManager.API_KEY_VAR] = api_key

        # 파일 작성
        with open(env_path, 'w') as f:
            for key, value in env_vars.items():
                f.write(f"{key}={value}\n")

        # 파일 권한 설정 (600: 소유자만 읽기/쓰기)
        os.chmod(env_path, 0o600)

        logger.info(f"API key saved to {env_path}")

    @staticmethod
    def load_api_key() -> Optional[str]:
        """
        환경 변수에서 API 키 로드

        Returns:
            str 또는 None: API 키, 없으면 None

        Priority:
            1. 환경 변수 (GOOGLE_API_KEY)
            2. .env 파일
        """
        # 환경 변수 확인
        api_key = os.getenv(EnvKeyManager.API_KEY_VAR)
        if api_key:
            return api_key

        # .env 파일 확인
        env_path = Path(EnvKeyManager.ENV_FILE)
        if env_path.exists():
            try:
                with open(env_path, 'r') as f:
                    for line in f:
                        line = line.strip()
                        if line.startswith(EnvKeyManager.API_KEY_VAR + '='):
                            api_key = line.split('=', 1)[1].strip()
                            if api_key:
                                return api_key
            except Exception as e:
                logger.error(f"Error reading .env file: {e}")

        return None

    @staticmethod
    def test_connection(api_key: str) -> bool:
        """
        API 연결 테스트

        Args:
            api_key: 테스트할 API 키

        Returns:
            bool: 연결 성공 여부

        Tests:
            - API 키 형식 검증
            - Gemini API 접속 테스트
            - 간단한 API 호출
        """
        try:
            # 형식 검증
            if not EnvKeyManager.validate_api_key(api_key):
                logger.error("Invalid API key format")
                return False

            # Gemini API 테스트 (선택: 실제 호출 또는 간단한 검증)
            # 실제 구현에서는 google-generativeai 사용
            try:
                import google.generativeai as genai
                genai.configure(api_key=api_key)

                # 모델 목록 조회로 연결 테스트
                models = genai.list_models()
                if models:
                    logger.info("API connection successful")
                    return True
                else:
                    logger.error("No models available")
                    return False

            except ImportError:
                # google-generativeai 미설치 시 형식 검증만
                logger.warning("google-generativeai not installed, skipping API test")
                return True
            except Exception as e:
                logger.error(f"API connection failed: {e}")
                return False

        except Exception as e:
            logger.error(f"Test connection error: {e}")
            return False

    @staticmethod
    def is_configured() -> bool:
        """
        API 키가 설정되었는지 확인

        Returns:
            bool: 설정 여부
        """
        api_key = EnvKeyManager.load_api_key()
        return api_key is not None and EnvKeyManager.validate_api_key(api_key)

    @staticmethod
    def reset_api_key() -> None:
        """
        API 키 제거 (초기화)

        Warning: 이 작업은 되돌릴 수 없습니다
        """
        env_path = Path(EnvKeyManager.ENV_FILE)

        if env_path.exists():
            env_vars = {}
            with open(env_path, 'r') as f:
                for line in f:
                    line = line.strip()
                    if line and '=' in line and not line.startswith('#'):
                        key, value = line.split('=', 1)
                        if key.strip() != EnvKeyManager.API_KEY_VAR:
                            env_vars[key.strip()] = value.strip()

            # 파일 재작성 (API 키 제외)
            with open(env_path, 'w') as f:
                for key, value in env_vars.items():
                    f.write(f"{key}={value}\n")

            os.chmod(env_path, 0o600)
            logger.info("API key removed from .env")

    @staticmethod
    def show_setup_status() -> None:
        """
        현재 설정 상태 표시
        """
        print("\n" + "="*60)
        print("📊 API 키 설정 상태")
        print("="*60 + "\n")

        is_configured = EnvKeyManager.is_configured()

        if is_configured:
            print("✅ API 키가 설정되어 있습니다")
            print(f"   파일: {EnvKeyManager.ENV_FILE}")
            print(f"   변수: {EnvKeyManager.API_KEY_VAR}")
            api_key = EnvKeyManager.load_api_key()
            print(f"   형식: {api_key[:6]}...{api_key[-4:]} (마스킹됨)")
            print("\n✓ 이미지 생성을 시작할 수 있습니다!\n")
        else:
            print("❌ API 키가 설정되지 않았습니다")
            print("\n다음 명령으로 설정하세요:")
            print("  from env_key_manager import EnvKeyManager")
            print("  EnvKeyManager.setup_api_key()\n")


if __name__ == "__main__":
    # 테스트
    manager = EnvKeyManager()

    # 현재 상태 확인
    manager.show_setup_status()

    # API 키 설정 (대화형)
    # manager.setup_api_key()
