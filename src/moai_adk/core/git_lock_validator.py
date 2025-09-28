"""
Git Lock Validator

Git 잠금 유효성 검증 전담 모듈

@FEATURE:GIT-LOCK-001 - 잠금 유효성 검증
@PERF:LOCK-100MS - 빠른 프로세스 검증
@SEC:LOCK-MED - 보안 강화된 검증
"""

import logging
import os
import time

# 로깅 설정 (@TASK:LOG-001)
logger = logging.getLogger(__name__)


class LockValidator:
    """Git 잠금 유효성 검증 전담 클래스

    Features:
    - 프로세스 실행 상태 검증
    - 타임스탬프 기반 만료 검사
    - 소유권 검증 (@SEC:LOCK-MED)
    """

    def __init__(self, max_lock_age_seconds: int = 3600):
        """Initialize LockValidator

        Args:
            max_lock_age_seconds: 최대 잠금 유지 시간 (기본: 1시간)
        """
        self.max_lock_age = max_lock_age_seconds
        logger.debug(f"LockValidator 초기화: 최대 유지시간={max_lock_age_seconds}초")

    def is_process_running(self, pid: int) -> bool:
        """프로세스 실행 상태 확인

        Args:
            pid: 프로세스 ID

        Returns:
            프로세스가 실행 중이면 True
        """
        if not pid:
            return False

        try:
            os.kill(pid, 0)  # 신호 0은 프로세스 존재 확인용
            return True
        except (OSError, ProcessLookupError):
            return False

    def is_lock_expired(self, lock_info: dict) -> bool:
        """잠금 만료 여부 확인

        Args:
            lock_info: 잠금 정보

        Returns:
            만료되었으면 True
        """
        if not lock_info:
            return True

        timestamp = lock_info.get("created_at")
        if not timestamp:
            return True

        try:
            lock_time = float(timestamp)
            age_seconds = time.time() - lock_time

            if age_seconds > self.max_lock_age:
                logger.warning(f"오래된 잠금 파일: {age_seconds}초 경과")
                return True

            return False

        except (ValueError, TypeError):
            logger.warning("잠금 타임스탬프 형식 오류")
            return True

    def is_lock_valid(self, lock_info: dict) -> bool:
        """잠금 유효성 종합 검사

        Args:
            lock_info: 잠금 정보

        Returns:
            유효한 잠금이면 True
        """
        if not lock_info:
            return False

        pid = lock_info.get("pid")
        if not pid:
            logger.debug("잠금 정보에 PID가 없음")
            return False

        # 프로세스 실행 상태 확인
        if not self.is_process_running(pid):
            logger.info(f"잠금 소유 프로세스가 종료됨: PID={pid}")
            return False

        # 만료 시간 확인
        if self.is_lock_expired(lock_info):
            return False

        return True

    def check_ownership(self, lock_info: dict, current_pid: int = None) -> bool:
        """잠금 소유권 확인

        Args:
            lock_info: 잠금 정보
            current_pid: 현재 프로세스 ID (기본: os.getpid())

        Returns:
            현재 프로세스가 잠금을 소유하면 True
        """
        if not lock_info:
            return False

        if current_pid is None:
            current_pid = os.getpid()

        lock_pid = lock_info.get("pid")
        is_owner = lock_pid == current_pid

        if not is_owner:
            logger.warning(
                f"잠금 소유권 불일치: 소유자 PID={lock_pid}, 현재 PID={current_pid}"
            )

        return is_owner

    def get_lock_age_seconds(self, lock_info: dict) -> float:
        """잠금 경과 시간 계산

        Args:
            lock_info: 잠금 정보

        Returns:
            잠금 생성 후 경과 시간 (초)
        """
        if not lock_info:
            return 0.0

        timestamp = lock_info.get("created_at")
        if not timestamp:
            return 0.0

        try:
            lock_time = float(timestamp)
            return time.time() - lock_time
        except (ValueError, TypeError):
            return 0.0

    def validate_lock_info_format(self, lock_info: dict) -> bool:
        """잠금 정보 형식 검증

        Args:
            lock_info: 검증할 잠금 정보

        Returns:
            형식이 올바르면 True
        """
        if not isinstance(lock_info, dict):
            return False

        required_fields = ["pid", "created_at"]
        for field in required_fields:
            if field not in lock_info:
                logger.warning(f"필수 필드 누락: {field}")
                return False

        # PID 타입 검증
        try:
            int(lock_info["pid"])
        except (ValueError, TypeError):
            logger.warning("잘못된 PID 형식")
            return False

        # 타임스탬프 타입 검증
        try:
            float(lock_info["created_at"])
        except (ValueError, TypeError):
            logger.warning("잘못된 타임스탬프 형식")
            return False

        return True