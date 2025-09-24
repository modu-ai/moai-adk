"""
SPEC Command Implementation

/moai:1-spec 명령어의 브랜치 스킵 옵션을 포함한 구현

@API:POST-SPEC - SPEC 생성 API 인터페이스
@PERF:CMD-FAST - 명령어 실행 최적화
@SEC:INPUT-MED - 입력 검증 보안 강화
"""

import logging
from pathlib import Path
from typing import Dict, Optional

from ..core.git_strategy import GitStrategyBase, PersonalGitStrategy, TeamGitStrategy
from ..core.exceptions import GitLockedException

# 로깅 설정 (@TASK:LOG-001)
logger = logging.getLogger(__name__)


class SpecCommand:
    """개선된 SPEC 명령어 - 브랜치 스킵 옵션 및 사용자 경험 향상

    TRUST 원칙 적용:
    - T: 테스트 가능한 구조 설계
    - R: 명확한 사용자 피드백
    - U: Git 전략 패턴 활용
    - S: 입력 검증 및 에러 처리
    - T: 상세한 실행 추적
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
            raise ValueError(f"project_dir은 Path 객체여야 합니다: {type(project_dir)}")

        if not project_dir.exists():
            raise ValueError(f"프로젝트 디렉토리가 존재하지 않습니다: {project_dir}")

        self.project_dir = project_dir.resolve()
        self.config = config
        self.skip_branch = skip_branch

        # Git 전략 초기화
        self._git_strategy: Optional[GitStrategyBase] = None

        logger.debug(f"SpecCommand 초기화: {self.project_dir}, skip_branch={skip_branch}")

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

            logger.debug(f"Git 전략 설정: {mode} -> {self._git_strategy.__class__.__name__}")

        return self._git_strategy

    def _get_current_mode(self) -> str:
        """현재 작업 모드 확인

        Returns:
            작업 모드 ('personal' 또는 'team')
        """
        if self.config:
            try:
                # ConfigManager 객체인 경우
                if hasattr(self.config, 'get_mode'):
                    return self.config.get_mode()
                # 딕셔너리인 경우
                elif isinstance(self.config, dict):
                    return self.config.get('mode', 'personal')
            except (AttributeError, KeyError):
                pass

        return 'personal'

    def execute(self, spec_name: str, description: str, skip_branch: Optional[bool] = None):
        """SPEC 명령어 실행

        Args:
            spec_name: 명세 이름
            description: 명세 설명
            skip_branch: 브랜치 생성 스킵 여부 (None이면 기본값 사용)

        Raises:
            ValueError: 유효하지 않은 입력
            GitLockedException: Git 작업 충돌
        """
        # 입력 검증
        validated_spec_name = self._validate_spec_name(spec_name)
        validated_description = self._validate_description(description)

        # 스킵 옵션 설정
        if skip_branch is not None:
            self.skip_branch = skip_branch

        # 실행 정보 로깅
        self._log_execution_start(validated_spec_name, validated_description)

        try:
            # SPEC 파일 생성
            self._create_spec_file(validated_spec_name, validated_description)

            # Git 작업 수행 (필요한 경우)
            if not self.skip_branch and self._should_create_branch():
                self._execute_git_workflow(validated_spec_name)

            # 성공 메시지
            self._log_execution_success(validated_spec_name)

        except Exception as e:
            self._log_execution_error(validated_spec_name, e)
            raise

    def execute_with_mode(self, mode: str, spec_name: str = "test-spec", description: str = "테스트 명세"):
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

    def _validate_spec_name(self, spec_name: str) -> str:
        """명세 이름 검증 및 정규화

        Args:
            spec_name: 검증할 명세 이름

        Returns:
            정규화된 명세 이름

        Raises:
            ValueError: 유효하지 않은 명세 이름
        """
        if not spec_name or not isinstance(spec_name, str):
            raise ValueError("spec_name은 비어있지 않은 문자열이어야 합니다")

        # 정규화
        normalized = spec_name.strip().upper()

        # 길이 제한
        if len(normalized) > 50:
            raise ValueError("명세 이름이 너무 깁니다 (최대 50자)")

        # 안전하지 않은 문자 확인
        unsafe_chars = ['/', '\\', '<', '>', ':', '"', '|', '?', '*']
        if any(char in normalized for char in unsafe_chars):
            raise ValueError(f"명세 이름에 안전하지 않은 문자가 포함되어 있습니다: {spec_name}")

        return normalized

    def _validate_description(self, description: str) -> str:
        """설명 검증 및 정규화

        Args:
            description: 검증할 설명

        Returns:
            정규화된 설명

        Raises:
            ValueError: 유효하지 않은 설명
        """
        if not description or not isinstance(description, str):
            raise ValueError("description은 비어있지 않은 문자열이어야 합니다")

        # 정규화
        normalized = description.strip()

        # 길이 제한
        if len(normalized) > 500:
            raise ValueError("설명이 너무 깁니다 (최대 500자)")

        return normalized

    def _create_spec_file(self, spec_name: str, description: str):
        """SPEC 파일 생성

        Args:
            spec_name: 명세 이름
            description: 명세 설명
        """
        try:
            specs_dir = self.project_dir / ".moai" / "specs"
            specs_dir.mkdir(parents=True, exist_ok=True)

            spec_file = specs_dir / f"{spec_name}.md"

            # 파일이 이미 존재하면 백업 생성
            if spec_file.exists():
                backup_file = specs_dir / f"{spec_name}.md.backup"
                spec_file.replace(backup_file)
                logger.info(f"기존 SPEC 파일 백업: {backup_file}")

            # SPEC 내용 생성
            spec_content = self._generate_spec_content(spec_name, description)

            # 파일 작성
            spec_file.write_text(spec_content, encoding='utf-8')

            logger.info(f"SPEC 파일 생성 완료: {spec_file}")

        except OSError as e:
            logger.error(f"SPEC 파일 생성 실패: {e}")
            raise ValueError(f"SPEC 파일 생성 실패: {e}")

    def _generate_spec_content(self, spec_name: str, description: str) -> str:
        """SPEC 파일 내용 생성 (간결한 버전)

        Args:
            spec_name: 명세 이름
            description: 명세 설명

        Returns:
            생성된 SPEC 내용
        """
        return f"""# {spec_name}

## 개요

{description}

## 요구사항

### 기능 요구사항

- [ ] 핵심 기능 구현
- [ ] 사용자 인터페이스 개발
- [ ] 데이터 처리 로직 구현

### 비기능 요구사항

- [ ] 성능: 응답 시간 < 1초
- [ ] 보안: 입력 검증 및 인증
- [ ] 안정성: 99% 가용성

## 수락 기준

- [ ] 모든 주요 기능이 정상 동작
- [ ] 테스트 커버리지 ≥ 85%
- [ ] 코드 리뷰 완료

## 태그 체계

@REQ:{spec_name}-001 - 주요 요구사항
@DESIGN:{spec_name}-ARCH-001 - 아키텍처 설계
@TASK:{spec_name}-IMPL-001 - 구현 작업
@TEST:{spec_name}-UNIT-001 - 단위 테스트

---

생성 일시: {self._get_current_timestamp()}
모드: {self._get_current_mode()}
"""

    def _get_current_timestamp(self) -> str:
        """현재 타임스탬프 반환

        Returns:
            포맷된 현재 시간
        """
        import datetime
        return datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S")

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
        """Git 워크플로우 실행

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
        logger.info("SPEC 명령어 실행 시작", extra={
            "command": "spec",
            "spec_name": spec_name,
            "description": description[:50] + "..." if len(description) > 50 else description,
            "mode": self._get_current_mode(),
            "skip_branch": self.skip_branch
        })

    def _log_execution_success(self, spec_name: str):
        """실행 성공 로깅"""
        logger.info(f"SPEC 명령어 실행 완료: {spec_name}")

    def _log_execution_error(self, spec_name: str, error: Exception):
        """실행 오류 로깅"""
        logger.error(f"SPEC 명령어 실행 실패: {spec_name}, 오류: {error}")

    def get_command_status(self) -> Dict:
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
            "specs_dir_exists": (self.project_dir / ".moai" / "specs").exists()
        }