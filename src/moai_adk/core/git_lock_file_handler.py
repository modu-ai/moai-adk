"""
Git Lock File Handler

Git 잠금 파일의 읽기/쓰기/파싱 담당 모듈

@FEATURE:GIT-LOCK-001 - 잠금 파일 I/O 전담
@PERF:LOCK-100MS - 파일 작업 최적화
@SEC:LOCK-MED - 파일 보안 검증
"""

import json
import logging
import os
import time
from pathlib import Path

# 로깅 설정 (@TASK:LOG-001)
logger = logging.getLogger(__name__)


class LockFileHandler:
    """Git 잠금 파일 읽기/쓰기 전담 핸들러

    Features:
    - JSON/레거시 형식 파싱 지원
    - 캐시 기반 성능 최적화 (@PERF:LOCK-100MS)
    - 강화된 파일 보안 검증 (@SEC:LOCK-MED)
    """

    def __init__(self, lock_file_path: Path):
        """Initialize LockFileHandler

        Args:
            lock_file_path: 잠금 파일 경로
        """
        self.lock_file = lock_file_path
        self.lock_dir = lock_file_path.parent

        # 성능 최적화를 위한 캐시
        self._lock_info_cache: dict | None = None
        self._last_check_time = 0.0

        logger.debug(f"LockFileHandler 초기화: {self.lock_file}")

    def ensure_lock_dir(self):
        """잠금 디렉토리가 존재하지 않으면 생성

        강화된 에러 처리 및 권한 검사 포함 (@SEC:LOCK-MED)

        Raises:
            OSError: 디렉토리 생성 실패 또는 권한 부족
        """
        try:
            self.lock_dir.mkdir(parents=True, exist_ok=True)

            # 권한 검사
            if not os.access(self.lock_dir, os.W_OK):
                raise OSError(f"잠금 디렉토리에 쓰기 권한이 없습니다: {self.lock_dir}")

            logger.debug(f"잠금 디렉토리 확보: {self.lock_dir}")

        except OSError as e:
            logger.error(f"잠금 디렉토리 생성 실패: {self.lock_dir}, 오류: {e}")
            raise

    def read_lock_info(self) -> dict | None:
        """잠금 파일 정보 읽기

        캐시된 정보를 사용하여 성능 최적화 (@PERF:LOCK-100MS)

        Returns:
            잠금 파일 정보 딕셔너리 또는 None
        """
        if not self.lock_file.exists():
            return None

        current_time = time.time()
        # 캐시가 유효한지 확인 (1초 캐시)
        if self._lock_info_cache and current_time - self._last_check_time < 1.0:
            return self._lock_info_cache

        try:
            with self.lock_file.open("r") as f:
                content = f.read().strip()

            # JSON 형식 시도
            try:
                lock_info = json.loads(content)
            except json.JSONDecodeError:
                # 레거시 형식 파싱
                lock_info = self._parse_legacy_format(content)

            self._lock_info_cache = lock_info
            self._last_check_time = current_time
            return lock_info

        except (OSError, ValueError) as e:
            logger.warning(f"잠금 파일 읽기 실패: {e}")
            return None

    def _parse_legacy_format(self, content: str) -> dict:
        """레거시 잠금 파일 형식 파싱

        Args:
            content: 잠금 파일 내용

        Returns:
            파싱된 정보 딕셔너리
        """
        info = {"pid": None, "timestamp": None, "legacy": True}

        for line in content.split("\n"):
            if line.startswith("PID:"):
                try:
                    info["pid"] = int(line.split(":", 1)[1].strip())
                except ValueError:
                    pass
            elif line.startswith("Time:"):
                info["timestamp"] = line.split(":", 1)[1].strip()

        return info

    def write_lock_info(self, lock_info: dict):
        """잠금 정보를 파일에 저장

        Args:
            lock_info: 저장할 잠금 정보

        Raises:
            OSError: 파일 쓰기 실패
        """
        try:
            with self.lock_file.open("w") as f:
                json.dump(lock_info, f, indent=2)

            # 캐시 업데이트
            self._lock_info_cache = lock_info
            self._last_check_time = time.time()

            logger.debug(f"잠금 정보 저장 완료: PID={lock_info.get('pid')}")

        except OSError as e:
            logger.error(f"잠금 파일 저장 실패: {e}")
            raise

    def delete_lock_file(self):
        """잠금 파일 삭제

        Returns:
            bool: 삭제 성공 여부
        """
        try:
            if self.lock_file.exists():
                self.lock_file.unlink()
                self._lock_info_cache = None  # 캐시 무효화
                logger.debug("잠금 파일 삭제 완료")
                return True
        except OSError as e:
            logger.warning(f"잠금 파일 삭제 중 오류: {e}")

        return False

    def lock_file_exists(self) -> bool:
        """잠금 파일 존재 여부 확인

        Returns:
            잠금 파일이 존재하면 True
        """
        return self.lock_file.exists()

    def create_lock_info(self) -> dict:
        """새로운 잠금 정보 생성

        Returns:
            잠금 파일에 저장할 정보 딕셔너리
        """
        return {
            "pid": os.getpid(),
            "created_at": time.time(),
            "timestamp": time.ctime(),
            "version": "2.0",
        }

    def invalidate_cache(self):
        """캐시 무효화"""
        self._lock_info_cache = None
        self._last_check_time = 0.0