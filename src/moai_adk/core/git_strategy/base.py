"""
Git Strategy Base Class

기본 Git 전략 인터페이스 및 공통 기능을 제공합니다.

@DESIGN:GIT-STRATEGY-BASE-001 - 추상 기본 클래스 분리
@TRUST:READABLE - 단일 책임: 기본 Git 작업만 담당
"""

import logging
import subprocess
from abc import ABC, abstractmethod
from collections.abc import Generator
from pathlib import Path

from ..git_lock_manager import GitLockManager

# 로깅 설정
logger = logging.getLogger(__name__)


class GitStrategyBase(ABC):
    """Git 전략 기본 클래스

    TRUST 원칙 적용:
    - T: 추상화를 통한 테스트 가능성 향상
    - R: 명확한 인터페이스 정의
    - U: 책임 분리 (전략별 구현)
    - S: 보안 검증 및 로깅
    - T: 작업 추적 가능성
    """

    def __init__(self, project_dir: Path, config=None):
        """Initialize Git Strategy

        Args:
            project_dir: 프로젝트 루트 디렉토리
            config: 설정 정보 (optional)
        """
        # 입력 검증
        if not isinstance(project_dir, Path):
            raise ValueError(f"project_dir은 Path 객체여야 합니다: {type(project_dir)}")

        if not project_dir.exists():
            raise ValueError(f"프로젝트 디렉토리가 존재하지 않습니다: {project_dir}")

        self.project_dir = project_dir.resolve()
        self.config = config or {}
        self.lock_manager = GitLockManager(project_dir)

        logger.debug(
            f"GitStrategy 초기화: {self.__class__.__name__}, 프로젝트: {self.project_dir}"
        )

    def get_current_branch(self) -> str:
        """현재 브랜치명 반환

        강화된 에러 처리 및 fallback 메커니즘

        Returns:
            현재 브랜치명
        """
        try:
            result = subprocess.run(
                ["git", "branch", "--show-current"],
                cwd=self.project_dir,
                capture_output=True,
                text=True,
                check=True,
                timeout=10,
            )
            branch = result.stdout.strip()

            if not branch:
                result = subprocess.run(
                    ["git", "rev-parse", "--short", "HEAD"],
                    cwd=self.project_dir,
                    capture_output=True,
                    text=True,
                    timeout=10,
                )
                if result.returncode == 0:
                    branch = f"HEAD@{result.stdout.strip()}"
                else:
                    branch = self._get_fallback_branch()

            logger.debug(f"현재 브랜치: {branch}")
            return branch

        except subprocess.TimeoutExpired:
            logger.error("Git 브랜치 확인 타임아웃")
            return self._get_fallback_branch()

        except (subprocess.CalledProcessError, FileNotFoundError) as e:
            logger.warning(f"Git 브랜치 확인 실패: {e}")
            return self._get_fallback_branch()

    def _get_fallback_branch(self) -> str:
        """Fallback 브랜치명 반환

        Returns:
            전략별 기본 브랜치명
        """
        from .team_strategy import TeamGitStrategy

        if isinstance(self, TeamGitStrategy):
            current = getattr(self, "_current_branch", None)
            return current if current else "feature/unknown"
        return "main"

    def is_git_repository(self) -> bool:
        """Git 저장소인지 확인

        Returns:
            Git 저장소이면 True
        """
        try:
            result = subprocess.run(
                ["git", "rev-parse", "--git-dir"],
                cwd=self.project_dir,
                capture_output=True,
                timeout=5,
            )
            return result.returncode == 0
        except (
            subprocess.CalledProcessError,
            subprocess.TimeoutExpired,
            FileNotFoundError,
        ):
            return False

    def get_repository_status(self) -> dict:
        """저장소 상태 정보 반환

        Returns:
            저장소 상태 정보 딕셔너리
        """
        status = {
            "is_git_repo": self.is_git_repository(),
            "current_branch": None,
            "has_changes": False,
            "is_locked": self.lock_manager.is_locked(),
        }

        if status["is_git_repo"]:
            try:
                status["current_branch"] = self.get_current_branch()

                result = subprocess.run(
                    ["git", "status", "--porcelain"],
                    cwd=self.project_dir,
                    capture_output=True,
                    text=True,
                    timeout=10,
                )
                status["has_changes"] = bool(result.stdout.strip())

            except (subprocess.CalledProcessError, subprocess.TimeoutExpired) as e:
                logger.warning(f"Git 상태 확인 실패: {e}")

        return status

    @abstractmethod
    def work_context(self, feature_name: str) -> Generator[None, None, None]:
        """작업 컨텍스트 - 서브클래스에서 구현

        Args:
            feature_name: 기능명

        Yields:
            None
        """

    def validate_feature_name(self, feature_name: str) -> str:
        """기능명 검증 및 정규화

        Args:
            feature_name: 검증할 기능명

        Returns:
            정규화된 기능명

        Raises:
            ValueError: 유효하지 않은 기능명
        """
        if not feature_name or not isinstance(feature_name, str):
            raise ValueError("feature_name은 비어있지 않은 문자열이어야 합니다")

        normalized = feature_name.strip().replace(" ", "-").lower()

        unsafe_chars = ["..", "/", "\\", "~", "^", ":", "[", "]", "*", "?"]
        if any(char in normalized for char in unsafe_chars):
            raise ValueError(
                f"기능명에 안전하지 않은 문자가 포함되어 있습니다: {feature_name}"
            )

        if len(normalized) > 100:
            raise ValueError("기능명이 너무 깁니다 (최대 100자)")

        return normalized

    def log_git_operation(self, operation: str, details: dict):
        """Git 작업 로깅

        Args:
            operation: 작업명
            details: 작업 세부사항
        """
        logger.info(
            f"Git 작업: {operation}",
            extra={
                "operation": operation,
                "strategy": self.__class__.__name__,
                "project_dir": str(self.project_dir),
                **details,
            },
        )