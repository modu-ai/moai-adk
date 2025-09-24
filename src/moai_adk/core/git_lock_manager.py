"""
Git Lock Manager

Git 작업 동시 실행 방지를 위한 잠금 관리 시스템
"""

import os
import time
from contextlib import contextmanager
from pathlib import Path
from typing import ContextManager, Generator, Optional

from .exceptions import GitLockedException


class GitLockManager:
    """Git 작업 동시 실행 방지를 위한 잠금 관리자

    .moai/locks/git.lock 파일을 사용하여 Git 작업의 동시 실행을 방지합니다.
    """

    def __init__(self, project_dir: Optional[Path] = None, lock_dir: str = ".moai/locks"):
        """Initialize GitLockManager

        Args:
            project_dir: 프로젝트 루트 디렉토리
            lock_dir: 잠금 파일이 저장될 디렉토리 경로 (project_dir 기준)
        """
        if project_dir is None:
            project_dir = Path.cwd()
        elif isinstance(project_dir, str):
            project_dir = Path(project_dir)

        self.project_dir = project_dir
        self.lock_dir = project_dir / lock_dir
        self.lock_file = self.lock_dir / "git.lock"

    def _ensure_lock_dir(self):
        """잠금 디렉토리가 존재하지 않으면 생성"""
        self.lock_dir.mkdir(parents=True, exist_ok=True)

    def is_locked(self) -> bool:
        """현재 잠금 상태 확인

        Returns:
            잠금 파일이 존재하면 True, 아니면 False
        """
        return self.lock_file.exists()

    def release_lock(self):
        """잠금 파일 삭제

        잠금 파일이 존재하지 않아도 오류를 발생시키지 않습니다.
        """
        try:
            if self.lock_file.exists():
                self.lock_file.unlink()
        except OSError:
            # 파일이 이미 삭제되었거나 다른 프로세스에서 삭제한 경우
            pass

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
        """실제 잠금 획득 컨텍스트 매니저"""
        # wait=True인 경우 잠금 대기 로직
        if wait:
            start_time = time.time()
            while self.is_locked():
                if time.time() - start_time > timeout:
                    raise GitLockedException(f"Git 작업 대기 시간 초과 ({timeout}초)")
                time.sleep(0.1)

        try:
            # 잠금 파일 생성
            with self.lock_file.open("w") as f:
                f.write(f"PID: {os.getpid()}\nTime: {time.ctime()}\n")

            yield

        finally:
            # 잠금 해제
            self.release_lock()

    def acquire_lock_direct(self, wait: bool = True, timeout: int = 30):
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
            while self.is_locked():
                if time.time() - start_time > timeout:
                    raise GitLockedException(f"Git 작업 대기 시간 초과 ({timeout}초)")
                time.sleep(0.1)

        # 잠금 파일 생성
        with self.lock_file.open("w") as f:
            f.write(f"PID: {os.getpid()}\nTime: {time.ctime()}\n")

        return True