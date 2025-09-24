#!/usr/bin/env python3
"""
MoAI Safe Git Operations - 잠금 기반 안전한 Git 작업

모든 MoAI 에이전트가 Git 작업을 수행할 때 사용하는 안전한 래퍼 함수들

@TASK:SAFE-GIT-001
"""

import os
import subprocess
import json
import logging
from pathlib import Path
from contextlib import contextmanager
from .git_lock_manager import GitLockManager

logger = logging.getLogger(__name__)

class SafeGitOperations:
    """잠금 기반 안전한 Git 작업 클래스"""

    def __init__(self, agent_name: str, project_root: str = None):
        self.agent_name = agent_name
        self.project_root = Path(project_root or os.getcwd())
        self.lock_manager = GitLockManager(project_root)

    @contextmanager
    def git_operation(self, operation: str, description: str = ""):
        """
        Git 작업 컨텍스트 매니저

        사용 예시:
        with safe_git.git_operation("commit", "TDD RED 커밋"):
            # Git 작업 수행
            pass
        """
        lock_acquired = False
        try:
            # 잠금 획득 시도
            lock_acquired = self.lock_manager.acquire_git_lock(
                self.agent_name, operation, description
            )

            if not lock_acquired:
                raise RuntimeError(f"Git 잠금 획득 실패: {operation}")

            logger.info(f"🔒 {self.agent_name}: Git {operation} 시작")
            yield

        except Exception as e:
            logger.error(f"❌ Git {operation} 실패: {e}")
            raise

        finally:
            if lock_acquired:
                self.lock_manager.release_git_lock(self.agent_name)
                logger.info(f"🔓 {self.agent_name}: Git {operation} 완료")

    def safe_branch_create(self, branch_name: str, base_branch: str = None) -> bool:
        """안전한 브랜치 생성"""
        with self.git_operation("branch_create", f"브랜치 생성: {branch_name}"):
            try:
                # 기준 브랜치 체크아웃
                if base_branch:
                    self._run_git_command(["checkout", base_branch])

                # 브랜치 생성 및 체크아웃
                self._run_git_command(["checkout", "-b", branch_name])

                logger.info(f"✅ 브랜치 생성 완료: {branch_name}")
                return True

            except subprocess.CalledProcessError as e:
                logger.error(f"❌ 브랜치 생성 실패: {e}")
                return False

    def safe_commit(self, message: str, files: list = None) -> bool:
        """안전한 커밋"""
        with self.git_operation("commit", f"커밋: {message[:50]}..."):
            try:
                # 파일 스테이징
                if files:
                    for file in files:
                        self._run_git_command(["add", str(file)])
                else:
                    self._run_git_command(["add", "."])

                # 변경사항 확인
                result = subprocess.run(
                    ["git", "status", "--porcelain"],
                    cwd=self.project_root,
                    capture_output=True,
                    text=True
                )

                if not result.stdout.strip():
                    logger.info("ℹ️ 커밋할 변경사항이 없습니다")
                    return True

                # 커밋 실행
                self._run_git_command(["commit", "-m", message])

                logger.info(f"✅ 커밋 완료: {message}")
                return True

            except subprocess.CalledProcessError as e:
                logger.error(f"❌ 커밋 실패: {e}")
                return False

    def safe_tag_create(self, tag_name: str, message: str = "") -> bool:
        """안전한 태그 생성"""
        with self.git_operation("tag_create", f"태그 생성: {tag_name}"):
            try:
                if message:
                    self._run_git_command(["tag", "-a", tag_name, "-m", message])
                else:
                    self._run_git_command(["tag", tag_name])

                logger.info(f"✅ 태그 생성 완료: {tag_name}")
                return True

            except subprocess.CalledProcessError as e:
                logger.error(f"❌ 태그 생성 실패: {e}")
                return False

    def safe_checkpoint_create(self, checkpoint_message: str) -> bool:
        """안전한 체크포인트 생성"""
        import datetime

        timestamp = datetime.datetime.now().strftime("%Y%m%d_%H%M%S")
        checkpoint_tag = f"moai_cp/{timestamp}"

        with self.git_operation("checkpoint", checkpoint_message):
            try:
                # 현재 변경사항 커밋 (있는 경우)
                status_result = subprocess.run(
                    ["git", "status", "--porcelain"],
                    cwd=self.project_root,
                    capture_output=True,
                    text=True
                )

                if status_result.stdout.strip():
                    commit_message = f"📍 Checkpoint: {checkpoint_message}"
                    if not self.safe_commit(commit_message):
                        return False

                # 체크포인트 태그 생성
                tag_message = f"MoAI Checkpoint: {checkpoint_message}"
                if not self.safe_tag_create(checkpoint_tag, tag_message):
                    return False

                logger.info(f"✅ 체크포인트 생성 완료: {checkpoint_tag}")
                return True

            except Exception as e:
                logger.error(f"❌ 체크포인트 생성 실패: {e}")
                return False

    def safe_sync(self, remote: str = "origin", branch: str = None) -> bool:
        """안전한 원격 동기화"""
        current_branch = branch or self._get_current_branch()

        with self.git_operation("sync", f"동기화: {remote}/{current_branch}"):
            try:
                # 원격 변경사항 가져오기
                self._run_git_command(["fetch", remote])

                # 충돌 검사
                if self._has_diverged(remote, current_branch):
                    logger.warning("⚠️ 브랜치가 분기되었습니다 - 수동 병합 필요")
                    return False

                # Fast-forward 병합 시도
                remote_branch = f"{remote}/{current_branch}"
                self._run_git_command(["merge", "--ff-only", remote_branch])

                logger.info(f"✅ 동기화 완료: {remote}/{current_branch}")
                return True

            except subprocess.CalledProcessError as e:
                logger.error(f"❌ 동기화 실패: {e}")
                return False

    def check_git_lock_status(self) -> dict:
        """현재 Git 잠금 상태 확인"""
        return self.lock_manager.get_lock_status()

    def wait_for_git_unlock(self, timeout: int = 300) -> bool:
        """Git 잠금 해제 대기"""
        import time

        start_time = time.time()
        while time.time() - start_time < timeout:
            if not self.lock_manager.is_git_locked():
                return True

            status = self.lock_manager.get_lock_status()
            if status.get("locked"):
                logger.info(
                    f"⏳ Git 잠금 대기 중... "
                    f"현재: {status['agent']} - {status['operation']}"
                )

            time.sleep(2)  # 2초 대기

        logger.error(f"❌ Git 잠금 대기 타임아웃 ({timeout}초)")
        return False

    def _run_git_command(self, args: list) -> subprocess.CompletedProcess:
        """Git 명령어 실행"""
        cmd = ["git"] + args
        logger.debug(f"Git 명령어 실행: {' '.join(cmd)}")

        result = subprocess.run(
            cmd,
            cwd=self.project_root,
            capture_output=True,
            text=True,
            check=True
        )

        return result

    def _get_current_branch(self) -> str:
        """현재 브랜치 이름 조회"""
        try:
            result = self._run_git_command(["branch", "--show-current"])
            return result.stdout.strip()
        except Exception:
            return "main"  # 기본값

    def _has_diverged(self, remote: str, branch: str) -> bool:
        """브랜치가 분기되었는지 확인"""
        try:
            # ahead/behind 개수 확인
            remote_branch = f"{remote}/{branch}"
            result = self._run_git_command([
                "rev-list", "--count", "--left-right",
                f"{branch}...{remote_branch}"
            ])

            ahead, behind = result.stdout.strip().split('\t')
            return int(ahead) > 0 and int(behind) > 0

        except Exception:
            return False  # 확인 불가시 안전하게 False

# 에이전트별 팩토리 함수
def create_safe_git_for_agent(agent_name: str) -> SafeGitOperations:
    """에이전트별 SafeGitOperations 인스턴스 생성"""
    return SafeGitOperations(agent_name)

# 편의 함수들
def spec_builder_git() -> SafeGitOperations:
    """spec-builder용 Git 작업"""
    return create_safe_git_for_agent("spec-builder")

def code_builder_git() -> SafeGitOperations:
    """code-builder용 Git 작업"""
    return create_safe_git_for_agent("code-builder")

def git_manager_git() -> SafeGitOperations:
    """git-manager용 Git 작업"""
    return create_safe_git_for_agent("git-manager")

def doc_syncer_git() -> SafeGitOperations:
    """doc-syncer용 Git 작업"""
    return create_safe_git_for_agent("doc-syncer")

# CLI 테스트용
if __name__ == "__main__":
    import sys

    if len(sys.argv) < 3:
        print("사용법: safe_git_operations.py <agent_name> <operation> [args...]")
        sys.exit(1)

    agent_name = sys.argv[1]
    operation = sys.argv[2]

    safe_git = SafeGitOperations(agent_name)

    if operation == "status":
        status = safe_git.check_git_lock_status()
        print(json.dumps(status, indent=2))

    elif operation == "checkpoint":
        message = sys.argv[3] if len(sys.argv) > 3 else "Test checkpoint"
        success = safe_git.safe_checkpoint_create(message)
        sys.exit(0 if success else 1)

    elif operation == "commit":
        message = sys.argv[3] if len(sys.argv) > 3 else "Test commit"
        success = safe_git.safe_commit(message)
        sys.exit(0 if success else 1)

    else:
        print(f"지원하지 않는 작업: {operation}")
        sys.exit(1)