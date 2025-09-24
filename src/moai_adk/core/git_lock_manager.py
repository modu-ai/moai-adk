"""
Git Lock Manager

Git 작업 동시 실행 방지를 위한 잠금 관리 시스템

@FEATURE:GIT-LOCK-001 - 동시 Git 작업 방지
@PERF:LOCK-100MS - 잠금 성능 최적화 (100ms 이내 응답)
@SEC:LOCK-MED - 잠금 파일 보안 강화
"""

import json
import logging
import os
import time
from contextlib import contextmanager
from pathlib import Path
from typing import Dict, Generator, Optional

from .exceptions import GitLockedException

# 로깅 설정 (@TASK:LOG-001)
logger = logging.getLogger(__name__)


class GitLockManager:
    """Git 작업 동시 실행 방지를 위한 잠금 관리자

    .moai/locks/git.lock 파일을 사용하여 Git 작업의 동시 실행을 방지합니다.

    Features:
    - 성능 최적화된 잠금 메커니즘 (@PERF:LOCK-100MS)
    - 강화된 에러 처리 및 복구 (@SEC:LOCK-MED)
    - 구조화된 로깅 (@TASK:LOG-001)
    - 자동 정리 및 모니터링 (@FEATURE:AUTO-CLEANUP-001)
    """

    def __init__(self, project_dir: Optional[Path] = None, lock_dir: str = ".moai/locks"):
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
            raise ValueError(f"project_dir must be a Path or str type: {type(project_dir)}")

        # Directory validation - allow non-existent directories during initialization
        if not project_dir.exists():
            logger.warning(f"Project directory will be created during installation: {project_dir}")
            # Directory will be created by installer - defer validation

        self.project_dir = project_dir
        self.lock_dir = project_dir / lock_dir
        self.lock_file = self.lock_dir / "git.lock"

        # 성능 최적화를 위한 캐시
        self._lock_info_cache: Optional[Dict] = None
        self._last_check_time = 0.0

        logger.debug(f"GitLockManager 초기화: {self.project_dir}, 잠금파일: {self.lock_file}")

    def _ensure_lock_dir(self):
        """잠금 디렉토리가 존재하지 않으면 생성

        강화된 에러 처리 및 권한 검사 포함 (@SEC:LOCK-MED)
        """
        try:
            self.lock_dir.mkdir(parents=True, exist_ok=True)
            # 권한 검사
            if not os.access(self.lock_dir, os.W_OK):
                raise GitLockedException(f"잠금 디렉토리에 쓰기 권한이 없습니다: {self.lock_dir}")
            logger.debug(f"잠금 디렉토리 확보: {self.lock_dir}")
        except OSError as e:
            logger.error(f"잠금 디렉토리 생성 실패: {self.lock_dir}, 오류: {e}")
            raise GitLockedException(f"잠금 디렉토리 생성 실패: {e}")

    def _get_lock_info(self) -> Optional[Dict]:
        """잠금 파일 정보 획득

        캐시된 정보를 사용하여 성능 최적화 (@PERF:LOCK-100MS)

        Returns:
            잠금 파일 정보 딕셔너리 또는 None
        """
        if not self.lock_file.exists():
            return None

        current_time = time.time()
        # 캐시가 유효한지 확인 (1초 캐시)
        if (self._lock_info_cache and
            current_time - self._last_check_time < 1.0):
            return self._lock_info_cache

        try:
            with self.lock_file.open('r') as f:
                content = f.read().strip()

            # JSON 형식 시도
            try:
                lock_info = json.loads(content)
            except json.JSONDecodeError:
                # 레거시 형식 파싱
                lock_info = self._parse_legacy_lock_format(content)

            self._lock_info_cache = lock_info
            self._last_check_time = current_time
            return lock_info

        except (OSError, ValueError) as e:
            logger.warning(f"잠금 파일 읽기 실패: {e}")
            return None

    def _parse_legacy_lock_format(self, content: str) -> Dict:
        """레거시 잠금 파일 형식 파싱

        Args:
            content: 잠금 파일 내용

        Returns:
            파싱된 정보 딕셔너리
        """
        info = {"pid": None, "timestamp": None, "legacy": True}

        for line in content.split('\n'):
            if line.startswith('PID:'):
                try:
                    info["pid"] = int(line.split(':', 1)[1].strip())
                except ValueError:
                    pass
            elif line.startswith('Time:'):
                info["timestamp"] = line.split(':', 1)[1].strip()

        return info

    def _is_process_running(self, pid: int) -> bool:
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

    def _is_lock_valid(self, lock_info: Dict) -> bool:
        """잠금 파일 유효성 검사

        Args:
            lock_info: 잠금 파일 정보

        Returns:
            유효한 잠금이면 True
        """
        if not lock_info:
            return False

        pid = lock_info.get("pid")
        if not pid:
            return False

        # 프로세스가 실행 중인지 확인
        if not self._is_process_running(pid):
            logger.info(f"잠금 소유 프로세스가 종료됨: PID={pid}")
            return False

        # 타임스탬프 검사 (1시간 초과 시 무효)
        timestamp = lock_info.get("created_at")
        if timestamp:
            try:
                lock_time = float(timestamp)
                if time.time() - lock_time > 3600:  # 1시간
                    logger.warning(f"오래된 잠금 파일: {time.time() - lock_time}초 경과")
                    return False
            except ValueError:
                pass

        return True

    def _cleanup_stale_lock(self):
        """무효한 잠금 파일 정리

        자동 정리 기능 (@FEATURE:AUTO-CLEANUP-001)
        """
        try:
            if self.lock_file.exists():
                self.lock_file.unlink()
                logger.info("무효한 잠금 파일 정리 완료")
        except OSError as e:
            logger.error(f"무효한 잠금 파일 정리 실패: {e}")

        # 캐시 무효화
        self._lock_info_cache = None

    def is_locked(self) -> bool:
        """현재 잠금 상태 확인

        성능 최적화된 잠금 상태 검사 (@PERF:LOCK-100MS)

        Returns:
            잠금 파일이 존재하고 유효하면 True, 아니면 False
        """
        if not self.lock_file.exists():
            return False

        # 잠금 파일 유효성 검증 (좀비 프로세스 확인)
        try:
            lock_info = self._get_lock_info()
            if lock_info and self._is_lock_valid(lock_info):
                return True
            else:
                # 무효한 잠금 파일 정리
                logger.info("무효한 잠금 파일 발견, 자동 정리 중")
                self._cleanup_stale_lock()
                return False
        except Exception as e:
            logger.error(f"잠금 상태 확인 중 오류: {e}")
            return True  # 안전을 위해 잠금된 것으로 간주

    def release_lock(self):
        """잠금 파일 삭제

        강화된 잠금 해제 메커니즘 (@SEC:LOCK-MED)
        """
        try:
            if self.lock_file.exists():
                # 잠금 소유권 확인
                lock_info = self._get_lock_info()
                current_pid = os.getpid()

                if lock_info and lock_info.get('pid') != current_pid:
                    logger.warning(f"다른 프로세스의 잠금 파일 해제 시도: "
                                 f"소유자 PID={lock_info.get('pid')}, 현재 PID={current_pid}")

                self.lock_file.unlink()
                self._lock_info_cache = None  # 캐시 무효화
                logger.debug(f"잠금 해제 완료: PID={current_pid}")

        except OSError as e:
            logger.warning(f"잠금 파일 삭제 중 오류 (무시됨): {e}")
            # 파일이 이미 삭제되었거나 다른 프로세스에서 삭제한 경우 무시

    def _create_lock_info(self) -> Dict:
        """잠금 정보 생성

        Returns:
            잠금 파일에 저장할 정보 딕셔너리
        """
        return {
            "pid": os.getpid(),
            "created_at": time.time(),
            "timestamp": time.ctime(),
            "version": "2.0"
        }

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
        self._ensure_lock_dir()

        # wait=False인 경우 즉시 검사하고 예외 발생
        if not wait and self.is_locked():
            raise GitLockedException("Git 작업이 이미 진행 중입니다")

        return self._acquire_lock_context(wait, timeout)

    @contextmanager
    def _acquire_lock_context(self, wait: bool = True, timeout: int = 30) -> Generator[None, None, None]:
        """실제 잠금 획득 컨텍스트 매니저

        성능 최적화된 대기 메커니즘 (@PERF:LOCK-100MS)
        """
        # wait=True인 경우 잠금 대기 로직
        if wait:
            start_time = time.time()
            check_interval = 0.05  # 50ms 간격으로 체크

            while self.is_locked():
                if time.time() - start_time > timeout:
                    raise GitLockedException(f"Git 작업 대기 시간 초과 ({timeout}초)")

                time.sleep(check_interval)
                # 점진적으로 체크 간격 증가 (최대 200ms)
                check_interval = min(check_interval * 1.2, 0.2)

        try:
            # 잠금 파일 생성 (JSON 형식)
            lock_info = self._create_lock_info()
            with self.lock_file.open("w") as f:
                json.dump(lock_info, f, indent=2)

            # 캐시 업데이트
            self._lock_info_cache = lock_info
            self._last_check_time = time.time()

            logger.debug(f"잠금 획득 완료: PID={lock_info['pid']}")
            yield

        finally:
            # 잠금 해제
            self.release_lock()

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
        self._ensure_lock_dir()

        # wait=False인 경우 즉시 검사하고 예외 발생
        if not wait and self.is_locked():
            raise GitLockedException("Git 작업이 이미 진행 중입니다")

        # wait=True인 경우 잠금 대기 로직
        if wait:
            start_time = time.time()
            check_interval = 0.05  # 50ms 간격으로 체크

            while self.is_locked():
                if time.time() - start_time > timeout:
                    raise GitLockedException(f"Git 작업 대기 시간 초과 ({timeout}초)")

                time.sleep(check_interval)
                check_interval = min(check_interval * 1.2, 0.2)

        # 잠금 파일 생성
        lock_info = self._create_lock_info()
        with self.lock_file.open("w") as f:
            json.dump(lock_info, f, indent=2)

        # 캐시 업데이트
        self._lock_info_cache = lock_info
        self._last_check_time = time.time()

        logger.debug(f"직접 잠금 획득 완료: PID={lock_info['pid']}")
        return True

    def get_lock_status(self) -> Dict:
        """잠금 상태 정보 반환

        모니터링 및 디버깅을 위한 상세 정보 제공

        Returns:
            잠금 상태 정보 딕셔너리
        """
        status = {
            "is_locked": self.is_locked(),
            "lock_file_exists": self.lock_file.exists(),
            "lock_dir_exists": self.lock_dir.exists(),
            "lock_dir_writable": self.lock_dir.exists() and os.access(self.lock_dir, os.W_OK)
        }

        if status["lock_file_exists"]:
            lock_info = self._get_lock_info()
            if lock_info:
                status.update({
                    "lock_info": lock_info,
                    "process_running": self._is_process_running(lock_info.get("pid", 0)),
                    "lock_age_seconds": time.time() - lock_info.get("created_at", 0)
                })

        return status