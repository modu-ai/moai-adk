"""
Git Branch Utility Functions

브랜치 관리를 위한 유틸리티 함수들입니다.

@DESIGN:GIT-BRANCH-UTILS-001 - 브랜치 관리 로직 분리
@TRUST:SIMPLE - 단일 책임: 브랜치 작업만 담당
"""

import logging
import subprocess
from pathlib import Path

# 로깅 설정
logger = logging.getLogger(__name__)


def get_base_branch(project_dir: Path, config=None) -> str:
    """베이스 브랜치 확인

    Args:
        project_dir: 프로젝트 디렉토리
        config: 설정 정보

    Returns:
        베이스 브랜치명 (main, master, develop 등)
    """
    # 설정에서 베이스 브랜치 확인
    if config:
        try:
            if hasattr(config, "get"):
                base_branch = config.get("base_branch")
                if base_branch:
                    return base_branch
            elif isinstance(config, dict) and "base_branch" in config:
                return config["base_branch"]
        except (AttributeError, KeyError):
            pass

    common_bases = ["main", "master", "develop"]

    # 원격 브랜치 확인
    try:
        result = subprocess.run(
            ["git", "branch", "-r"],
            cwd=project_dir,
            capture_output=True,
            text=True,
            timeout=10,
        )

        if result.returncode == 0:
            remote_branches = result.stdout
            for base in common_bases:
                if f"origin/{base}" in remote_branches:
                    return base

    except (subprocess.CalledProcessError, subprocess.TimeoutExpired):
        pass

    # 로컬 브랜치 확인
    try:
        result = subprocess.run(
            ["git", "branch", "--list"] + common_bases,
            cwd=project_dir,
            capture_output=True,
            text=True,
            timeout=10,
        )

        if result.returncode == 0:
            branches = result.stdout
            for base in common_bases:
                if base in branches:
                    return base

    except (subprocess.CalledProcessError, subprocess.TimeoutExpired):
        pass

    return "main"


def branch_exists(project_dir: Path, branch_name: str) -> bool:
    """브랜치 존재 여부 확인

    Args:
        project_dir: 프로젝트 디렉토리
        branch_name: 확인할 브랜치명

    Returns:
        브랜치가 존재하면 True
    """
    try:
        result = subprocess.run(
            ["git", "show-ref", "--verify", "--quiet", f"refs/heads/{branch_name}"],
            cwd=project_dir,
            capture_output=True,
            timeout=10,
        )
        return result.returncode == 0
    except (subprocess.CalledProcessError, subprocess.TimeoutExpired):
        return False


def pull_latest_changes(project_dir: Path, branch_name: str):
    """최신 변경사항 가져오기

    Args:
        project_dir: 프로젝트 디렉토리
        branch_name: 업데이트할 브랜치명
    """
    try:
        result = subprocess.run(
            ["git", "remote"],
            cwd=project_dir,
            capture_output=True,
            text=True,
            timeout=10,
        )

        if result.returncode == 0 and result.stdout.strip():
            subprocess.run(
                ["git", "pull", "--ff-only"],
                cwd=project_dir,
                capture_output=True,
                timeout=30,
            )
            logger.debug(f"브랜치 업데이트 완료: {branch_name}")

    except (subprocess.CalledProcessError, subprocess.TimeoutExpired):
        logger.debug("원격 업데이트 건너뜀")