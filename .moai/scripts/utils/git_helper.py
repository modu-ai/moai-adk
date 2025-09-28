#!/usr/bin/env python3
"""
MoAI-ADK Git 관련 헬퍼 유틸리티

@REQ:GIT-UTILS-001
@FEATURE:GIT-ABSTRACTION-001
@API:GET-GIT-STATUS
@DESIGN:GIT-WRAPPER-001
"""

import logging
import os
import re
import subprocess
from pathlib import Path

from constants import ERROR_MESSAGES, GIT_COMMAND_TIMEOUT, REGEX_PATTERNS

logger = logging.getLogger(__name__)


class GitCommandError(Exception):
    """Git 명령어 실행 오류"""

    def __init__(self, command: list[str], returncode: int, stderr: str):
        self.command = command
        self.returncode = returncode
        self.stderr = stderr
        super().__init__(f"Git command failed: {' '.join(command)}")


class GitHelper:
    """Git 명령어 헬퍼 클래스"""

    def __init__(self, project_root: Path | None = None):
        self.project_root = project_root or Path.cwd()
        self.git_env = os.environ.copy()

    def run_command(
        self, cmd: list[str], check: bool = True, timeout: int | None = None
    ) -> subprocess.CompletedProcess:
        """Git 명령어 실행"""
        if timeout is None:
            timeout = GIT_COMMAND_TIMEOUT

        try:
            result = subprocess.run(
                cmd,
                capture_output=True,
                text=True,
                check=False,
                timeout=timeout,
                cwd=self.project_root,
                env=self.git_env,
            )

            if check and result.returncode != 0:
                raise GitCommandError(cmd, result.returncode, result.stderr or "")
            return result
        except FileNotFoundError:
            raise FileNotFoundError(ERROR_MESSAGES["git_not_found"])

    def is_git_repo(self) -> bool:
        """Git 저장소인지 확인"""
        try:
            result = self.run_command(["git", "rev-parse", "--git-dir"])
            return result.returncode == 0
        except (GitCommandError, FileNotFoundError):
            return False

    def get_current_branch(self) -> str:
        """현재 브랜치명 반환"""
        result = self.run_command(["git", "branch", "--show-current"])
        return result.stdout.strip()

    def get_local_branches(self) -> list[str]:
        """로컬 브랜치 목록 반환"""
        result = self.run_command(["git", "branch"])
        branches = []
        for line in result.stdout.splitlines():
            branch = line.strip().lstrip("* ")
            if branch and not branch.startswith("("):
                branches.append(branch)
        return branches

    def has_uncommitted_changes(self) -> bool:
        """커밋되지 않은 변경사항 확인"""
        result = self.run_command(["git", "status", "--porcelain"])
        return bool(result.stdout.strip())

    def stage_all_changes(self) -> None:
        """모든 변경사항 스테이징"""
        self.run_command(["git", "add", "-A"])

    def commit(self, message: str, allow_empty: bool = False) -> str:
        """커밋 생성"""
        cmd = ["git", "commit", "-m", message]
        if allow_empty:
            cmd.append("--allow-empty")

        self.run_command(cmd)
        result = self.run_command(["git", "rev-parse", "HEAD"])
        return result.stdout.strip()

    def create_branch(self, branch_name: str) -> None:
        """브랜치 생성"""
        if not re.match(REGEX_PATTERNS["branch_name"], branch_name):
            raise ValueError(f"유효하지 않은 브랜치명: {branch_name}")
        self.run_command(["git", "checkout", "-b", branch_name])

    def switch_branch(self, branch_name: str) -> None:
        """브랜치 전환"""
        self.run_command(["git", "checkout", branch_name])

    def delete_branch(self, branch_name: str, force: bool = False) -> None:
        """브랜치 삭제"""
        flag = "-D" if force else "-d"
        self.run_command(["git", "branch", flag, branch_name])

    def stash_push(self, message: str | None = None) -> str:
        """Stash 생성"""
        cmd = ["git", "stash", "push"]
        if message:
            cmd.extend(["-m", message])
        self.run_command(cmd)

        result = self.run_command(["git", "stash", "list", "-1", "--format=%gd"])
        return result.stdout.strip()

    def stash_pop(self) -> None:
        """Stash 복원"""
        self.run_command(["git", "stash", "pop"])

    def create_tag(self, tag_name: str, message: str | None = None) -> None:
        """태그 생성"""
        cmd = ["git", "tag"]
        if message:
            cmd.extend(["-a", tag_name, "-m", message])
        else:
            cmd.append(tag_name)
        self.run_command(cmd)

    def push(
        self,
        remote: str = "origin",
        branch: str | None = None,
        set_upstream: bool = False,
    ) -> None:
        """푸시"""
        if branch is None:
            branch = self.get_current_branch()

        cmd = ["git", "push"]
        if set_upstream:
            cmd.extend(["-u", remote, branch])
        else:
            cmd.extend([remote, branch])
        self.run_command(cmd)

    def pull(self, remote: str = "origin") -> None:
        """풀"""
        self.run_command(["git", "pull", remote])

    def fetch(self, remote: str = "origin") -> None:
        """페치"""
        self.run_command(["git", "fetch", remote])

    def get_commit_info(self, commit: str = "HEAD") -> dict[str, str]:
        """커밋 정보 반환"""
        result = self.run_command(
            ["git", "show", "--format=%H|%an|%ad|%s", "--no-patch", commit]
        )

        parts = result.stdout.strip().split("|", 3)
        if len(parts) >= 4:
            return {
                "hash": parts[0],
                "author": parts[1],
                "date": parts[2],
                "message": parts[3],
            }
        return {}

    def is_clean_working_tree(self) -> bool:
        """작업 트리가 깨끗한지 확인"""
        return not self.has_uncommitted_changes()

    def has_remote(self, remote: str = "origin") -> bool:
        """원격 저장소 설정 확인"""
        try:
            result = self.run_command(["git", "remote"])
            return remote in result.stdout.splitlines()
        except GitCommandError:
            return False
