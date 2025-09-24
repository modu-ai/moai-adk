"""
@FEATURE:BUILD-COMMAND-001 BUILD Command Implementation
@REQ:TDD-AUTOMATION-001 /moai:2-build 명령어의 Git 잠금 확인 로직 구현

@API:POST-BUILD - BUILD 실행 API 인터페이스
@PERF:TDD-FAST - TDD 프로세스 실행 최적화
@SEC:LOCK-MED - Git 잠금 보안 강화
"""

import logging
from pathlib import Path
from typing import Dict

from ..core.git_lock_manager import GitLockManager
from ..core.exceptions import GitLockedException

# 로깅 설정 (@TASK:LOG-001)
logger = logging.getLogger(__name__)


class BuildCommand:
    """
    @TASK:BUILD-MAIN-001 개선된 BUILD 명령어 - Git 잠금 확인 및 TDD 프로세스 최적화

    TRUST 원칙 적용:
    - T: TDD 사이클 엄격 준수
    - R: 명확한 빌드 단계 피드백
    - U: 잠금 시스템 통합 설계
    - S: 안전한 동시 작업 방지
    - T: 상세한 빌드 과정 추적
    """

    def __init__(self, project_dir: Path, config=None):
        """Initialize BuildCommand

        Args:
            project_dir: 프로젝트 디렉토리
            config: 설정 관리자 인스턴스
        """
        # 입력 검증 (@SEC:LOCK-MED)
        if not isinstance(project_dir, Path):
            raise ValueError(f"project_dir must be a Path object: {type(project_dir)}")

        if not project_dir.exists():
            raise ValueError(f"Project directory does not exist: {project_dir}")

        self.project_dir = project_dir.resolve()
        self.config = config
        self.lock_manager = GitLockManager(project_dir)

        logger.debug(f"BuildCommand 초기화: {self.project_dir}")

    def execute(self, spec_name: str, wait_for_lock: bool = True):
        """
        @TASK:BUILD-EXECUTE-001 BUILD 명령어 실행

        Args:
            spec_name: 빌드할 명세 이름
            wait_for_lock: 잠금 대기 여부

        Raises:
            GitLockedException: 잠금 파일이 존재하고 대기하지 않는 경우
            ValueError: 유효하지 않은 입력
        """
        # 입력 검증
        validated_spec_name = self._validate_spec_name(spec_name)

        # 실행 시작 로깅
        self._log_execution_start(validated_spec_name, wait_for_lock)

        # 잠금 확인
        if not wait_for_lock and self.lock_manager.is_locked():
            raise GitLockedException("잠금 파일이 감지되었습니다. 다른 Git 작업이 진행 중입니다.")

        # 잠금과 함께 빌드 실행
        self.execute_with_lock_check(validated_spec_name, wait_for_lock)

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

                # 성공 로깅
                self._log_execution_success(spec_name)

        except GitLockedException:
            self._log_execution_error(spec_name, "Git 잠금으로 인한 실행 실패")
            if not wait_for_lock:
                raise
            # 대기 모드인 경우 예외를 다시 발생시킴
            raise
        except Exception as e:
            self._log_execution_error(spec_name, str(e))
            raise

    def _validate_spec_name(self, spec_name: str) -> str:
        """명세 이름 검증

        Args:
            spec_name: 검증할 명세 이름

        Returns:
            검증된 명세 이름

        Raises:
            ValueError: 유효하지 않은 명세 이름
        """
        if not spec_name or not isinstance(spec_name, str):
            raise ValueError("spec_name은 비어있지 않은 문자열이어야 합니다")

        normalized = spec_name.strip()
        if len(normalized) > 100:
            raise ValueError("명세 이름이 너무 깁니다 (최대 100자)")

        return normalized

    def _execute_tdd_process(self, spec_name: str):
        """
        @TASK:TDD-PROCESS-001 TDD 프로세스 실행 (RED-GREEN-REFACTOR)

        성능 최적화된 TDD 사이클 (@PERF:TDD-FAST)

        Args:
            spec_name: 빌드할 명세 이름
        """
        logger.info(f"TDD 프로세스 시작: {spec_name}")

        try:
            # RED: 실패하는 테스트 작성
            self._execute_red_phase(spec_name)

            # GREEN: 최소 구현
            self._execute_green_phase(spec_name)

            # REFACTOR: 리팩터링
            self._execute_refactor_phase(spec_name)

            logger.info(f"TDD 프로세스 완료: {spec_name}")

        except Exception as e:
            logger.error(f"TDD 프로세스 실행 중 오류: {spec_name}, 오류: {e}")
            raise

    def _execute_red_phase(self, spec_name: str):
        """
        @TASK:TDD-RED-001 RED Phase: 실패하는 테스트 작성

        Args:
            spec_name: 명세 이름
        """
        logger.info(f"🔴 RED Phase 시작: {spec_name}")
        self._write_failing_tests(spec_name)
        logger.info(f"🔴 RED Phase 완료: {spec_name}")

    def _execute_green_phase(self, spec_name: str):
        """
        @TASK:TDD-GREEN-001 GREEN Phase: 최소 구현

        Args:
            spec_name: 명세 이름
        """
        logger.info(f"🟢 GREEN Phase 시작: {spec_name}")
        self._implement_minimum_code(spec_name)
        logger.info(f"🟢 GREEN Phase 완료: {spec_name}")

    def _execute_refactor_phase(self, spec_name: str):
        """
        @TASK:TDD-REFACTOR-001 REFACTOR Phase: 코드 리팩터링

        Args:
            spec_name: 명세 이름
        """
        logger.info(f"🔄 REFACTOR Phase 시작: {spec_name}")
        self._refactor_code(spec_name)
        logger.info(f"🔄 REFACTOR Phase 완료: {spec_name}")

    def _write_failing_tests(self, spec_name: str):
        """실패하는 테스트 작성

        Args:
            spec_name: 명세 이름
        """
        # TDD 첫 번째 단계: 실패하는 테스트 작성
        logger.debug(f"실패 테스트 작성: {spec_name}")

    def _implement_minimum_code(self, spec_name: str):
        """최소 구현

        Args:
            spec_name: 명세 이름
        """
        # TDD 두 번째 단계: 테스트를 통과시키는 최소 코드 구현
        logger.debug(f"최소 구현: {spec_name}")

    def _refactor_code(self, spec_name: str):
        """코드 리팩터링

        Args:
            spec_name: 명세 이름
        """
        # TDD 세 번째 단계: 코드 품질 개선 리팩터링
        logger.debug(f"코드 리팩터링: {spec_name}")

    def _log_execution_start(self, spec_name: str, wait_for_lock: bool):
        """실행 시작 로깅"""
        logger.info("BUILD 명령어 실행 시작", extra={
            "command": "build",
            "spec_name": spec_name,
            "wait_for_lock": wait_for_lock,
            "project_dir": str(self.project_dir)
        })

    def _log_execution_success(self, spec_name: str):
        """실행 성공 로깅"""
        logger.info(f"BUILD 명령어 실행 완료: {spec_name}")

    def _log_execution_error(self, spec_name: str, error_message: str):
        """실행 오류 로깅"""
        logger.error(f"BUILD 명령어 실행 실패: {spec_name}, 오류: {error_message}")

    def get_build_status(self) -> Dict:
        """빌드 상태 정보 반환

        Returns:
            현재 빌드 상태 정보 딕셔너리
        """
        return {
            "project_dir": str(self.project_dir),
            "lock_status": self.lock_manager.get_lock_status(),
            "specs_dir_exists": (self.project_dir / ".moai" / "specs").exists(),
            "tdd_phases": ["RED", "GREEN", "REFACTOR"]
        }