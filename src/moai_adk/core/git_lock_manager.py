"""
Git Lock Manager

Git 작업 동시 실행 방지를 위한 잠금 관리 시스템

@FEATURE:GIT-LOCK-001 - 동시 Git 작업 방지
@PERF:LOCK-100MS - 잠금 성능 최적화 (100ms 이내 응답)
@SEC:LOCK-MED - 잠금 파일 보안 강화
"""

import logging
import os
import time
from contextlib import contextmanager
from pathlib import Path
from typing import Generator

from .exceptions import GitLockedException
from .git_lock_acquirer import LockAcquirer
from .git_lock_file_handler import LockFileHandler
from .git_lock_validator import LockValidator

# 로깅 설정 (@TASK:LOG-001)
logger = logging.getLogger(__name__)


class GitLockManager:
    """Git 작업 동시 실행 방지를 위한 잠금 관리자

    .moai/locks/git.lock 파일을 사용하여 Git 작업의 동시 실행을 방지합니다.

    Features:
    - 모듈화된 아키텍처 (TRUST-U 원칙)
    - 성능 최적화된 잠금 메커니즘 (@PERF:LOCK-100MS)
    - 강화된 에러 처리 및 복구 (@SEC:LOCK-MED)
    - 구조화된 로깅 (@TASK:LOG-001)
    - 자동 정리 및 모니터링 (@FEATURE:AUTO-CLEANUP-001)
    """

    def __init__(self, project_dir: Path | None = None, lock_dir: str = ".moai/locks"):
        """Initialize GitLockManager

        Args:
            project_dir: 프로젝트 루트 디렉토리
            lock_dir: 잠금 파일이 저장될 디렉토리 경로 (project_dir 기준)
        """
        # 입력 검증 강화 (@SEC:LOCK-MED)
        if project_dir is None:
            project_dir = Path.cwd()
        elif isinstance(project_dir, str):
            project_dir = Path(project_dir).resolve()  # 절대 경로로 변환
        elif isinstance(project_dir, Path):
            project_dir = project_dir.resolve()
        else:
            raise ValueError(
                f"project_dir must be a Path or str type: {type(project_dir)}"
            )

        # Directory validation - allow non-existent directories during initialization
        if not project_dir.exists():
            logger.warning(
                f"Project directory will be created during installation: {project_dir}"
            )
            # Directory will be created by installer - defer validation

        self.project_dir = project_dir
        self.lock_dir = project_dir / lock_dir
        self.lock_file = self.lock_dir / "git.lock"

        # 모듈화된 컴포넌트 초기화
        self.file_handler = LockFileHandler(self.lock_file)
        self.validator = LockValidator()
        self.acquirer = LockAcquirer(self.file_handler, self.validator)

        logger.debug(
            f"GitLockManager 초기화: {self.project_dir}, 잠금파일: {self.lock_file}"
        )

    # API 호환성을 위한 위임 메서드들

    def is_locked(self) -> bool:
        """현재 잠금 상태 확인

        성능 최적화된 잠금 상태 검사 (@PERF:LOCK-100MS)

        Returns:
            잠금 파일이 존재하고 유효하면 True, 아니면 False
        """
        if not self.file_handler.lock_file_exists():
            return False

        # 잠금 파일 유효성 검증 (좀비 프로세스 확인)
        try:
            lock_info = self.file_handler.read_lock_info()
            if lock_info and self.validator.is_lock_valid(lock_info):
                return True
            else:
                # 무효한 잠금 파일 정리
                logger.info("무효한 잠금 파일 발견, 자동 정리 중")
                self.acquirer.cleanup_stale_lock()
                return False
        except Exception as e:
            logger.error(f"잠금 상태 확인 중 오류: {e}")
            return True  # 안전을 위해 잠금된 것으로 간주

    def release_lock(self):
        """잠금 파일 삭제

        강화된 잠금 해제 메커니즘 (@SEC:LOCK-MED)
        """
        self.acquirer.release_lock()

    def acquire_lock(self, wait: bool = True, timeout: int = 30):
        """잠금 획득

        컨텍스트 매니저로 사용하면 자동으로 잠금이 해제되고,
        직접 호출하면 검증만 수행합니다.

        Args:
            wait: 잠금 대기 여부
            timeout: 대기 시간 (초)

        Returns:
            contextmanager 또는 None

        Raises:
            GitLockedException: 잠금 획득에 실패한 경우
        """
        return self.acquirer.acquire_lock(wait, timeout)

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
        return self.acquirer.acquire_lock_direct(wait, timeout)

    def get_lock_status(self) -> dict:
        """잠금 상태 정보 반환

        모니터링 및 디버깅을 위한 상세 정보 제공

        Returns:
            잠금 상태 정보 딕셔너리
        """
        status = {
            "is_locked": self.is_locked(),
            "lock_file_exists": self.file_handler.lock_file_exists(),
            "lock_dir_exists": self.lock_dir.exists(),
            "lock_dir_writable": self.lock_dir.exists()
            and os.access(self.lock_dir, os.W_OK),
        }

        if status["lock_file_exists"]:
            lock_info = self.file_handler.read_lock_info()
            if lock_info:
                status.update(
                    {
                        "lock_info": lock_info,
                        "process_running": self.validator.is_process_running(
                            lock_info.get("pid", 0)
                        ),
                        "lock_age_seconds": self.validator.get_lock_age_seconds(lock_info),
                    }
                )

        return status
