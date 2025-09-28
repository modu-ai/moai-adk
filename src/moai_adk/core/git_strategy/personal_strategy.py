"""
Personal Git Strategy

개인 모드: main 브랜치에서 직접 작업하는 간단한 전략입니다.

@DESIGN:GIT-STRATEGY-PERSONAL-001 - 개인 모드 전략 분리
@TRUST:SIMPLE - 단일 책임: 개인 모드 Git 작업만 담당
"""

import logging
from collections.abc import Generator
from contextlib import contextmanager

from .base import GitStrategyBase

# 로깅 설정
logger = logging.getLogger(__name__)


class PersonalGitStrategy(GitStrategyBase):
    """개인 모드: main 브랜치에서 직접 작업

    브랜치 생성 없이 현재 브랜치에서 바로 작업을 수행합니다.

    특징:
    - 브랜치 생성/전환 없음
    - 단순한 워크플로우 (50% 간소화 달성)
    - 개인 프로젝트 최적화
    """

    @contextmanager
    def work_context(self, feature_name: str) -> Generator[None, None, None]:
        """작업 컨텍스트 - 현재 브랜치에서 직접 작업

        Args:
            feature_name: 기능명 (개인 모드에서는 로깅용으로만 사용)

        Yields:
            None
        """
        # 기능명 검증
        validated_name = self.validate_feature_name(feature_name)

        # 저장소 상태 확인
        repo_status = self.get_repository_status()

        self.log_git_operation(
            "personal_work_start",
            {
                "feature_name": validated_name,
                "current_branch": repo_status.get("current_branch"),
                "has_changes": repo_status.get("has_changes"),
            },
        )

        # 잠금 획득하여 동시 작업 방지
        with self.lock_manager.acquire_lock():
            try:
                logger.info(f"개인 모드 작업 시작: {validated_name}")
                yield

                logger.info(f"개인 모드 작업 완료: {validated_name}")

            except Exception as e:
                logger.error(f"개인 모드 작업 중 오류: {e}")
                self.log_git_operation(
                    "personal_work_error",
                    {"feature_name": validated_name, "error": str(e)},
                )
                raise

            finally:
                self.log_git_operation(
                    "personal_work_end", {"feature_name": validated_name}
                )