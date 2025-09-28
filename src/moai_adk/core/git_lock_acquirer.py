"""
Git Lock Acquirer

Git 잠금 획득/해제 로직 전담 모듈

@FEATURE:GIT-LOCK-001 - 잠금 획득/해제 로직
@PERF:LOCK-100MS - 최적화된 대기 메커니즘
@SEC:LOCK-MED - 안전한 잠금 처리
"""

import logging
import time
from collections.abc import Generator
from contextlib import contextmanager

from .exceptions import GitLockedException
from .git_lock_file_handler import LockFileHandler
from .git_lock_validator import LockValidator

# 로깅 설정 (@TASK:LOG-001)
logger = logging.getLogger(__name__)


class LockAcquirer:
    """Git 잠금 획득/해제 전담 클래스

    Features:
    - 성능 최적화된 대기 메커니즘 (@PERF:LOCK-100MS)
    - 컨텍스트 매니저 지원
    - 타임아웃 및 재시도 로직
    """

    def __init__(self, file_handler: LockFileHandler, validator: LockValidator):
        """Initialize LockAcquirer

        Args:
            file_handler: 잠금 파일 핸들러
            validator: 잠금 검증기
        """
        self.file_handler = file_handler
        self.validator = validator
        logger.debug("LockAcquirer 초기화 완료")

    def acquire_lock(self, wait: bool = True, timeout: int = 30):
        """잠금 획득 (컨텍스트 매니저 반환)

        Args:
            wait: 잠금 대기 여부
            timeout: 대기 시간 (초)

        Returns:
            contextmanager

        Raises:
            GitLockedException: 잠금 획득에 실패한 경우
        """
        self.file_handler.ensure_lock_dir()

        # wait=False인 경우 즉시 검사하고 예외 발생
        if not wait and self._is_currently_locked():
            raise GitLockedException("Git 작업이 이미 진행 중입니다")

        return self._acquire_lock_context(wait, timeout)

    def acquire_lock_direct(self, wait: bool = True, timeout: int = 30) -> bool:
        """잠금 직접 획득 (컨텍스트 매니저 없이)

        Args:
            wait: 잠금 대기 여부
            timeout: 대기 시간 (초)

        Returns:
            bool: 잠금 획득 성공 여부

        Raises:
            GitLockedException: 잠금 획득에 실패한 경우
        """
        self.file_handler.ensure_lock_dir()

        # wait=False인 경우 즉시 검사하고 예외 발생
        if not wait and self._is_currently_locked():
            raise GitLockedException("Git 작업이 이미 진행 중입니다")

        # wait=True인 경우 잠금 대기 로직
        if wait:
            self._wait_for_lock_release(timeout)

        # 잠금 파일 생성
        lock_info = self.file_handler.create_lock_info()
        self.file_handler.write_lock_info(lock_info)

        logger.debug(f"직접 잠금 획득 완료: PID={lock_info['pid']}")
        return True

    def release_lock(self):
        """잠금 해제

        강화된 잠금 해제 메커니즘 (@SEC:LOCK-MED)
        """
        if not self.file_handler.lock_file_exists():
            return

        # 잠금 소유권 확인
        lock_info = self.file_handler.read_lock_info()
        if lock_info and not self.validator.check_ownership(lock_info):
            logger.warning("다른 프로세스의 잠금 파일 해제 시도")

        success = self.file_handler.delete_lock_file()
        if success:
            logger.debug("잠금 해제 완료")

    def cleanup_stale_lock(self):
        """무효한 잠금 파일 정리

        자동 정리 기능 (@FEATURE:AUTO-CLEANUP-001)
        """
        lock_info = self.file_handler.read_lock_info()
        if lock_info and not self.validator.is_lock_valid(lock_info):
            success = self.file_handler.delete_lock_file()
            if success:
                logger.info("무효한 잠금 파일 정리 완료")
            else:
                logger.error("무효한 잠금 파일 정리 실패")

    def _is_currently_locked(self) -> bool:
        """현재 잠금 상태 확인

        Returns:
            잠금이 유효하면 True
        """
        if not self.file_handler.lock_file_exists():
            return False

        lock_info = self.file_handler.read_lock_info()
        return self.validator.is_lock_valid(lock_info)

    def _wait_for_lock_release(self, timeout: int):
        """잠금 해제 대기

        성능 최적화된 대기 메커니즘 (@PERF:LOCK-100MS)

        Args:
            timeout: 대기 시간 (초)

        Raises:
            GitLockedException: 타임아웃 발생
        """
        start_time = time.time()
        check_interval = 0.05  # 50ms 간격으로 체크

        while self._is_currently_locked():
            if time.time() - start_time > timeout:
                raise GitLockedException(f"Git 작업 대기 시간 초과 ({timeout}초)")

            time.sleep(check_interval)
            # 점진적으로 체크 간격 증가 (최대 200ms)
            check_interval = min(check_interval * 1.2, 0.2)

    @contextmanager
    def _acquire_lock_context(
        self, wait: bool = True, timeout: int = 30
    ) -> Generator[None, None, None]:
        """실제 잠금 획득 컨텍스트 매니저

        성능 최적화된 대기 메커니즘 (@PERF:LOCK-100MS)

        Args:
            wait: 잠금 대기 여부
            timeout: 대기 시간 (초)

        Yields:
            None

        Raises:
            GitLockedException: 잠금 획득 실패
        """
        # wait=True인 경우 잠금 대기 로직
        if wait:
            self._wait_for_lock_release(timeout)

        try:
            # 잠금 파일 생성
            lock_info = self.file_handler.create_lock_info()
            self.file_handler.write_lock_info(lock_info)

            logger.debug(f"잠금 획득 완료: PID={lock_info['pid']}")
            yield

        finally:
            # 잠금 해제
            self.release_lock()

    def get_current_lock_info(self) -> dict | None:
        """현재 잠금 정보 조회

        Returns:
            잠금 정보 딕셔너리 또는 None
        """
        return self.file_handler.read_lock_info()