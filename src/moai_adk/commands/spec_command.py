"""
@FEATURE:SPEC-COMMAND-001 SPEC Command Implementation
@REQ:SPEC-CREATION-001 → @DESIGN:SPEC-ARCHITECTURE-001 → @TASK:SPEC-MAIN-001 → @TEST:SPEC-EXECUTION-001

@REQ:SPEC-CREATION-001 /moai:1-spec command implementation including branch skip options
@DESIGN:SPEC-ARCHITECTURE-001 Git strategy pattern integration

@API:POST-SPEC SPEC creation API interface
@PERF:CMD-FAST Command execution optimization
@SEC:INPUT-MED Input validation security enhancement
"""

import logging
from pathlib import Path

from ..core.exceptions import GitLockedException
from ..core.git_strategy import GitStrategyBase, PersonalGitStrategy, TeamGitStrategy
from .spec_validator import SpecValidator
from .spec_file_generator import SpecFileGenerator

# Logging setup (@TASK:LOG-001)
logger = logging.getLogger(__name__)


class SpecCommand:
    """
    @TASK:SPEC-MAIN-001 Enhanced SPEC command - branch skip options and user experience improvement

    TRUST principles applied:
    - T: Testable structure design
    - R: Clear user feedback
    - U: Git strategy pattern utilization
    - S: Input validation and error handling
    - T: Detailed execution tracking
    """

    def __init__(self, project_dir: Path, config=None, skip_branch: bool = False):
        """Initialize SpecCommand

        Args:
            project_dir: 프로젝트 디렉토리
            config: 설정 관리자 인스턴스
            skip_branch: 브랜치 생성 스킵 옵션
        """
        # 입력 검증 (@SEC:INPUT-MED)
        if not isinstance(project_dir, Path):
            raise ValueError(f"project_dir must be a Path object: {type(project_dir)}")

        if not project_dir.exists():
            raise ValueError(f"Project directory does not exist: {project_dir}")

        self.project_dir = project_dir.resolve()
        self.config = config
        self.skip_branch = skip_branch

        # Git 전략 초기화
        self._git_strategy: GitStrategyBase | None = None

        # 검증 및 파일 생성 모듈 초기화
        self.validator = SpecValidator()
        self.file_generator = SpecFileGenerator(self.project_dir, self._get_current_mode())

        logger.debug(
            f"SpecCommand 초기화: {self.project_dir}, skip_branch={skip_branch}"
        )

    def _get_git_strategy(self) -> GitStrategyBase:
        """Git 전략 획득

        Returns:
            현재 모드에 적합한 Git 전략 인스턴스
        """
        if self._git_strategy is None:
            mode = self._get_current_mode()

            if mode == "team":
                self._git_strategy = TeamGitStrategy(self.project_dir, self.config)
            else:
                self._git_strategy = PersonalGitStrategy(self.project_dir, self.config)

            logger.debug(
                f"Git 전략 설정: {mode} -> {self._git_strategy.__class__.__name__}"
            )

        return self._git_strategy

    def _get_current_mode(self) -> str:
        """현재 작업 모드 확인

        Returns:
            작업 모드 ('personal' 또는 'team')
        """
        if self.config:
            try:
                # ConfigManager 객체인 경우
                if hasattr(self.config, "get_mode"):
                    return self.config.get_mode()
                # 딕셔너리인 경우
                elif isinstance(self.config, dict):
                    return self.config.get("mode", "personal")
            except (AttributeError, KeyError):
                pass

        return "personal"

    def execute(
        self, spec_name: str, description: str, skip_branch: bool | None = None
    ):
        """
        @TASK:SPEC-EXECUTE-001 SPEC 명령어 실행

        Args:
            spec_name: 명세 이름
            description: 명세 설명
            skip_branch: 브랜치 생성 스킵 여부 (None이면 기본값 사용)

        Raises:
            ValueError: 유효하지 않은 입력
            GitLockedException: Git 작업 충돌
        """
        # 입력 검증
        validated_spec_name = self.validator.validate_spec_name(spec_name)
        validated_description = self.validator.validate_description(description)

        # 스킵 옵션 설정
        if skip_branch is not None:
            self.skip_branch = skip_branch

        # 실행 정보 로깅
        self._log_execution_start(validated_spec_name, validated_description)

        try:
            # SPEC 파일 생성
            self.file_generator.create_spec_file(validated_spec_name, validated_description)

            # Git 작업 수행 (필요한 경우)
            if not self.skip_branch and self._should_create_branch():
                self._execute_git_workflow(validated_spec_name)

            # 성공 메시지
            self._log_execution_success(validated_spec_name)

        except Exception as e:
            self._log_execution_error(validated_spec_name, e)
            raise

    def execute_with_mode(
        self, mode: str, spec_name: str = "test-spec", description: str = "테스트 명세"
    ):
        """모드별 실행 전략

        Args:
            mode: 실행 모드 (personal/team)
            spec_name: 명세 이름
            description: 명세 설명
        """
        # 모드 검증
        if mode not in ["personal", "team"]:
            raise ValueError(f"지원하지 않는 모드입니다: {mode}")

        logger.info(f"모드별 실행: {mode}")

        if mode == "personal" and self.skip_branch:
            # 개인 모드에서는 브랜치 스킵 가능
            self.execute(spec_name, description, skip_branch=True)
        elif mode == "team":
            # 팀 모드에서는 항상 브랜치 생성
            self.execute(spec_name, description, skip_branch=False)
        else:
            # 기본 실행
            self.execute(spec_name, description)


    def _should_create_branch(self) -> bool:
        """브랜치 생성 여부 결정

        Returns:
            브랜치를 생성해야 하면 True, 아니면 False
        """
        mode = self._get_current_mode()

        # 팀 모드에서만 브랜치 생성
        should_create = mode == "team"

        logger.debug(f"브랜치 생성 여부: {should_create} (모드: {mode})")
        return should_create

    def _execute_git_workflow(self, spec_name: str):
        """
        @TASK:SPEC-GIT-WORKFLOW-001 Git 워크플로우 실행

        Args:
            spec_name: 명세 이름 (브랜치명에 사용)
        """
        try:
            git_strategy = self._get_git_strategy()

            # Git 작업 컨텍스트에서 실행
            with git_strategy.work_context(f"spec-{spec_name.lower()}"):
                logger.info(f"Git 워크플로우 실행: {spec_name}")

        except GitLockedException:
            logger.warning("Git 잠금으로 인해 브랜치 작업을 건너뜁니다")
            # 잠금 상황에서는 SPEC 파일만 생성하고 계속 진행
        except Exception as e:
            logger.error(f"Git 워크플로우 실행 중 오류: {e}")
            # Git 작업 실패 시에도 SPEC 생성은 완료된 상태로 유지
            raise

    def _log_execution_start(self, spec_name: str, description: str):
        """실행 시작 로깅"""
        logger.info(
            "SPEC 명령어 실행 시작",
            extra={
                "command": "spec",
                "spec_name": spec_name,
                "description": description[:50] + "..."
                if len(description) > 50
                else description,
                "mode": self._get_current_mode(),
                "skip_branch": self.skip_branch,
            },
        )

    def _log_execution_success(self, spec_name: str):
        """실행 성공 로깅"""
        logger.info(f"SPEC 명령어 실행 완료: {spec_name}")

    def _log_execution_error(self, spec_name: str, error: Exception):
        """실행 오류 로깅"""
        logger.error(f"SPEC 명령어 실행 실패: {spec_name}, 오류: {error}")

    def get_command_status(self) -> dict:
        """명령어 상태 정보 반환

        Returns:
            현재 명령어 상태 정보 딕셔너리
        """
        git_strategy = self._get_git_strategy()

        return {
            "project_dir": str(self.project_dir),
            "mode": self._get_current_mode(),
            "skip_branch": self.skip_branch,
            "git_strategy": git_strategy.__class__.__name__,
            "repository_status": git_strategy.get_repository_status(),
            "specs_dir_exists": (self.project_dir / ".moai" / "specs").exists(),
        }
