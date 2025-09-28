"""
Team Git Strategy

팀 모드: feature 브랜치 생성 후 작업하는 협업 전략입니다.

@DESIGN:GIT-STRATEGY-TEAM-001 - 팀 모드 전략 분리
@TRUST:UNIFIED - 단일 책임: 팀 모드 Git 작업만 담당
"""

import logging
import subprocess
from collections.abc import Generator
from contextlib import contextmanager
from pathlib import Path

from .base import GitStrategyBase
from .branch_utils import branch_exists, get_base_branch, pull_latest_changes

# 로깅 설정
logger = logging.getLogger(__name__)


class TeamGitStrategy(GitStrategyBase):
    """팀 모드: feature 브랜치 생성 후 작업

    feature 브랜치를 생성하고 해당 브랜치에서 작업을 수행합니다.

    특징:
    - 체계적인 브랜치 관리
    - 협업 친화적 워크플로우
    - PR/MR 준비 자동화
    """

    def __init__(self, project_dir: Path, config=None):
        super().__init__(project_dir, config)
        self._current_branch: str | None = None
        self._feature_branch: str | None = None

    def _get_base_branch(self) -> str:
        """베이스 브랜치 확인 (위임)"""
        return get_base_branch(self.project_dir, self.config)

    def _create_feature_branch(self, feature_name: str) -> str:
        """feature 브랜치 생성

        Args:
            feature_name: 기능명

        Returns:
            생성된 브랜치명
        """
        base_branch = self._get_base_branch()
        feature_branch = f"feature/{feature_name}"

        try:
            if branch_exists(self.project_dir, base_branch):
                subprocess.run(
                    ["git", "checkout", base_branch],
                    cwd=self.project_dir,
                    capture_output=True,
                    check=True,
                    timeout=30,
                )
                pull_latest_changes(self.project_dir, base_branch)

            if branch_exists(self.project_dir, feature_branch):
                subprocess.run(
                    ["git", "checkout", feature_branch],
                    cwd=self.project_dir,
                    capture_output=True,
                    check=True,
                    timeout=30,
                )
                logger.info(f"기존 feature 브랜치로 전환: {feature_branch}")
            else:
                subprocess.run(
                    ["git", "checkout", "-b", feature_branch],
                    cwd=self.project_dir,
                    capture_output=True,
                    check=True,
                    timeout=30,
                )
                logger.info(f"새 feature 브랜치 생성: {feature_branch}")

            return feature_branch

        except subprocess.CalledProcessError as e:
            logger.error(f"브랜치 생성/전환 실패: {e}")
            return feature_branch


    @contextmanager
    def work_context(self, feature_name: str) -> Generator[None, None, None]:
        """작업 컨텍스트 - feature 브랜치 생성

        Args:
            feature_name: 기능명 (브랜치명에 사용)

        Yields:
            None
        """
        validated_name = self.validate_feature_name(feature_name)
        original_branch = self.get_current_branch()
        self._current_branch = original_branch

        repo_status = self.get_repository_status()

        self.log_git_operation(
            "team_work_start",
            {
                "feature_name": validated_name,
                "original_branch": original_branch,
                "has_changes": repo_status.get("has_changes"),
            },
        )

        with self.lock_manager.acquire_lock():
            feature_branch = None

            try:
                feature_branch = self._create_feature_branch(validated_name)
                self._feature_branch = feature_branch

                logger.info(
                    f"팀 모드 작업 시작: {validated_name} (브랜치: {feature_branch})"
                )

                self._current_branch = feature_branch
                yield

                logger.info(f"팀 모드 작업 완료: {validated_name}")

            except Exception as e:
                logger.error(f"팀 모드 작업 중 오류: {e}")
                self.log_git_operation(
                    "team_work_error",
                    {
                        "feature_name": validated_name,
                        "feature_branch": feature_branch,
                        "error": str(e),
                    },
                )
                raise

            finally:
                self.log_git_operation(
                    "team_work_end",
                    {
                        "feature_name": validated_name,
                        "feature_branch": feature_branch,
                        "stayed_on_feature": True,
                    },
                )

    def get_feature_branch_info(self) -> dict:
        """현재 feature 브랜치 정보 반환

        Returns:
            feature 브랜치 정보 딕셔너리
        """
        return {
            "current_branch": self._current_branch,
            "feature_branch": self._feature_branch,
            "base_branch": self._get_base_branch(),
            "is_feature_branch": self._feature_branch
            and self._current_branch == self._feature_branch,
        }