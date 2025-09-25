#!/usr/bin/env python3
"""
MoAI-ADK 통합 Git 워크플로우 시스템

@REQ:GIT-WORKFLOW-001
@FEATURE:GIT-MANAGEMENT-001
@API:GET-GIT-WORKFLOW
@DESIGN:UNIFIED-GIT-001
"""

import logging
import re
from pathlib import Path
from typing import Any

from checkpoint_system import CheckpointSystem
from constants import (
    DEFAULT_BRANCH_NAME,
    FEATURE_BRANCH_PREFIX,
    HOTFIX_BRANCH_PREFIX,
    PERSONAL_MODE,
    TEAM_MODE,
)
from git_helper import GitCommandError, GitHelper
from project_helper import ProjectHelper

logger = logging.getLogger(__name__)


class GitWorkflowError(Exception):
    """Git 워크플로우 관련 오류"""


class GitWorkflow:
    """통합 Git 워크플로우 관리"""

    def __init__(self, project_root: Path | None = None):
        self.project_root = project_root or ProjectHelper.find_project_root()
        self.git = GitHelper(self.project_root)
        self.checkpoint_system = CheckpointSystem(self.project_root)
        self.config = ProjectHelper.load_config(self.project_root)
        self.mode = self.config.get("mode", PERSONAL_MODE)

    def create_feature_branch(
        self, feature_name: str, from_branch: str | None = None
    ) -> str:
        """기능 브랜치 생성"""
        try:
            if not self._is_valid_branch_name(feature_name):
                raise GitWorkflowError(f"유효하지 않은 브랜치명: {feature_name}")

            base_branch = from_branch or self._get_default_branch()
            branch_name = f"{FEATURE_BRANCH_PREFIX}{feature_name}"

            if self.git.has_uncommitted_changes():
                self.checkpoint_system.create_checkpoint(
                    f"Pre-branch creation: {branch_name}", is_auto=True
                )

            if base_branch != self.git.get_current_branch():
                self.git.switch_branch(base_branch)

            if self.mode == TEAM_MODE and self.git.has_remote():
                self.git.pull()

            self.git.create_branch(branch_name)
            logger.info(f"기능 브랜치 생성 완료: {branch_name}")
            return branch_name

        except GitCommandError as e:
            raise GitWorkflowError(f"브랜치 생성 실패: {e}")

    def create_constitution_commit(
        self, message: str, files: list[str] | None = None
    ) -> str:
        """개발 가이드 기반 커밋 생성"""
        try:
            if not message.strip():
                raise GitWorkflowError("커밋 메시지가 비어있습니다.")

            if files:
                for file_path in files:
                    self.git.run_command(["git", "add", file_path])
            else:
                self.git.stage_all_changes()

            commit_hash = self.git.commit(self._format_commit_message(message))

            if self.mode == PERSONAL_MODE:
                self.checkpoint_system.create_checkpoint(
                    f"Commit: {message[:50]}", is_auto=True
                )

            logger.info(f"개발 가이드 커밋 생성 완료: {commit_hash[:8]}")
            return commit_hash

        except GitCommandError as e:
            raise GitWorkflowError(f"커밋 생성 실패: {e}")

    def sync_with_remote(self, push: bool = True, branch: str | None = None) -> bool:
        """원격 저장소와 동기화"""
        try:
            if not self.git.has_remote():
                logger.warning("원격 저장소가 설정되지 않았습니다.")
                return False

            current_branch = branch or self.git.get_current_branch()

            if self.mode == TEAM_MODE:
                try:
                    self.git.pull()
                    logger.info("원격 변경사항 가져오기 완료")
                except GitCommandError as e:
                    logger.warning(f"Pull 실패: {e}")

            if push and not self.git.has_uncommitted_changes():
                try:
                    self.git.push(branch=current_branch, set_upstream=True)
                    logger.info(f"원격 저장소 푸시 완료: {current_branch}")
                except GitCommandError as e:
                    logger.warning(f"Push 실패: {e}")
                    return False

            return True

        except Exception as e:
            logger.error(f"동기화 실패: {e}")
            return False

    def create_hotfix_branch(self, fix_name: str) -> str:
        """핫픽스 브랜치 생성"""
        try:
            if not self._is_valid_branch_name(fix_name):
                raise GitWorkflowError(f"유효하지 않은 핫픽스명: {fix_name}")

            branch_name = f"{HOTFIX_BRANCH_PREFIX}{fix_name}"
            main_branch = self._get_default_branch()

            if self.git.has_uncommitted_changes():
                self.checkpoint_system.create_checkpoint(
                    f"Pre-hotfix: {fix_name}", is_auto=True
                )

            if main_branch != self.git.get_current_branch():
                self.git.switch_branch(main_branch)

            if self.mode == TEAM_MODE and self.git.has_remote():
                self.git.pull()

            self.git.create_branch(branch_name)
            logger.info(f"핫픽스 브랜치 생성 완료: {branch_name}")
            return branch_name

        except GitCommandError as e:
            raise GitWorkflowError(f"핫픽스 브랜치 생성 실패: {e}")

    def get_branch_status(self) -> dict[str, Any]:
        """브랜치 상태 조회"""
        try:
            current_branch = self.git.get_current_branch()
            local_branches = self.git.get_local_branches()
            has_uncommitted = self.git.has_uncommitted_changes()

            status = {
                "current_branch": current_branch,
                "local_branches": local_branches,
                "has_uncommitted_changes": has_uncommitted,
                "has_remote": self.git.has_remote(),
                "mode": self.mode,
                "clean_working_tree": self.git.is_clean_working_tree(),
            }

            return status

        except Exception as e:
            logger.error(f"브랜치 상태 조회 실패: {e}")
            return {"error": str(e)}

    def cleanup_merged_branches(self, dry_run: bool = True) -> list[str]:
        """병합된 브랜치 정리"""
        try:
            result = self.git.run_command(
                ["git", "branch", "--merged", self._get_default_branch()]
            )

            merged_branches = []
            for line in result.stdout.splitlines():
                branch = line.strip().lstrip("* ")
                if (
                    branch
                    and not branch.startswith("(")
                    and branch != self._get_default_branch()
                ):
                    merged_branches.append(branch)

            if not dry_run:
                for branch in merged_branches:
                    try:
                        self.git.delete_branch(branch)
                        logger.info(f"병합된 브랜치 삭제: {branch}")
                    except GitCommandError as e:
                        logger.warning(f"브랜치 삭제 실패 {branch}: {e}")

            return merged_branches

        except GitCommandError as e:
            logger.error(f"병합된 브랜치 조회 실패: {e}")
            return []

    def _get_default_branch(self) -> str:
        """기본 브랜치 반환"""
        return self.config.get("git", {}).get("default_branch", DEFAULT_BRANCH_NAME)

    def _is_valid_branch_name(self, name: str) -> bool:
        """브랜치명 유효성 검사"""
        if not name or len(name) > 50:
            return False
        return re.match(r"^[a-zA-Z0-9._/-]+$", name) is not None

    def _format_commit_message(self, message: str) -> str:
        """개발 가이드 기반 커밋 메시지 포맷팅"""
        if not message.strip():
            return message

        formatted = message.strip()
        if not any(
            formatted.startswith(prefix)
            for prefix in ["🔧", "✨", "🐛", "📚", "🧪", "♻️"]
        ):
            if "feat" in message.lower() or "feature" in message.lower():
                formatted = f"✨ {formatted}"
            elif "fix" in message.lower() or "bug" in message.lower():
                formatted = f"🐛 {formatted}"
            elif "doc" in message.lower():
                formatted = f"📚 {formatted}"
            elif "test" in message.lower():
                formatted = f"🧪 {formatted}"
            elif "refactor" in message.lower():
                formatted = f"♻️ {formatted}"
            else:
                formatted = f"🔧 {formatted}"

        return formatted


def create_feature_branch(feature_name: str, project_root: Path | None = None) -> str:
    """기능 브랜치 생성 편의 함수"""
    workflow = GitWorkflow(project_root)
    return workflow.create_feature_branch(feature_name)


def create_constitution_commit(message: str, project_root: Path | None = None) -> str:
    """개발 가이드 커밋 생성 편의 함수"""
    workflow = GitWorkflow(project_root)
    return workflow.create_constitution_commit(message)


def sync_with_remote(project_root: Path | None = None, push: bool = True) -> bool:
    """원격 동기화 편의 함수"""
    workflow = GitWorkflow(project_root)
    return workflow.sync_with_remote(push=push)
