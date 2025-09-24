"""
BUILD Command Implementation

/moai:2-build 명령어의 Git 잠금 확인 로직 구현
"""

from pathlib import Path
from typing import Optional

from ..core.git_lock_manager import GitLockManager
from ..core.exceptions import GitLockedException


class BuildCommand:
    """개선된 BUILD 명령어 - 잠금 확인 로직"""

    def __init__(self, project_dir: Path, config=None):
        """Initialize BuildCommand

        Args:
            project_dir: 프로젝트 디렉토리
            config: 설정 관리자 인스턴스
        """
        self.project_dir = project_dir
        self.config = config
        self.lock_manager = GitLockManager(project_dir)

    def execute(self, spec_name: str, wait_for_lock: bool = True):
        """BUILD 명령어 실행

        Args:
            spec_name: 빌드할 명세 이름
            wait_for_lock: 잠금 대기 여부

        Raises:
            GitLockedException: 잠금 파일이 존재하고 대기하지 않는 경우
        """
        # 잠금 확인
        if not wait_for_lock and self.lock_manager.is_locked():
            raise GitLockedException("잠금 파일이 감지되었습니다. 다른 Git 작업이 진행 중입니다.")

        # 잠금과 함께 빌드 실행
        self.execute_with_lock_check(spec_name, wait_for_lock)

    def execute_with_lock_check(self, spec_name: str = "test-spec", wait_for_lock: bool = True):
        """잠금 확인 후 실행

        Args:
            spec_name: 빌드할 명세 이름
            wait_for_lock: 잠금 대기 여부
        """
        try:
            with self.lock_manager.acquire_lock(wait=wait_for_lock):
                # TDD 빌드 프로세스 실행
                self._execute_tdd_process(spec_name)

        except GitLockedException:
            if not wait_for_lock:
                raise
            # 대기 모드인 경우 예외를 다시 발생시킴
            raise

    def _execute_tdd_process(self, spec_name: str):
        """TDD 프로세스 실행 (RED-GREEN-REFACTOR)

        Args:
            spec_name: 빌드할 명세 이름
        """
        # 최소 구현: 실제 TDD 프로세스는 생략
        # 테스트 통과를 위한 더미 구현

        # RED: 테스트 작성
        self._write_failing_tests(spec_name)

        # GREEN: 최소 구현
        self._implement_minimum_code(spec_name)

        # REFACTOR: 리팩터링
        self._refactor_code(spec_name)

    def _write_failing_tests(self, spec_name: str):
        """실패하는 테스트 작성"""
        # 더미 구현
        pass

    def _implement_minimum_code(self, spec_name: str):
        """최소 구현"""
        # 더미 구현
        pass

    def _refactor_code(self, spec_name: str):
        """코드 리팩터링"""
        # 더미 구현
        pass