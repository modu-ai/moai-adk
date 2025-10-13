# @CODE:CORE-GIT-001 | SPEC: SPEC-CORE-GIT-001.md
"""GitHub Pull Request 유틸리티"""
from typing import Dict, Optional
import os
import subprocess
from pathlib import Path

def create_draft_pr(
    spec_id: str,
    base_branch: str = 'develop',
    title: Optional[str] = None,
    body: Optional[str] = None
) -> str:
    """
    GitHub Draft Pull Request 생성

    :param spec_id: SPEC ID (예: CORE-GIT-001)
    :param base_branch: 기본 브랜치 (기본값: develop)
    :param title: PR 제목 (기본값: SPEC ID 기반)
    :param body: PR 본문 (기본값: 자동 생성)
    :return: 생성된 PR의 URL
    """
    title = title or f"SPEC-{spec_id} 구현"
    body = body or f"@SPEC:{spec_id} 구현을 위한 Pull Request"

    try:
        cmd = [
            'gh', 'pr', 'create',
            '--draft',
            '--base', base_branch,
            '--title', title,
            '--body', body
        ]
        result = subprocess.run(cmd, capture_output=True, text=True, check=True)
        return result.stdout.strip()
    except subprocess.CalledProcessError as e:
        raise RuntimeError(f"Draft PR 생성 실패: {e.stderr}")

def get_repo_status() -> Dict[str, str]:
    """
    현재 Git 저장소의 상태 반환

    :return: 저장소 상태 정보를 담은 딕셔너리
    """
    try:
        result = subprocess.run(
            ['git', 'status', '--porcelain', '--branch'],
            capture_output=True, text=True, check=True
        )
        status = result.stdout.strip()

        changes_staged = len([line for line in status.split('\n') if line.startswith('M ')])
        changes_unstaged = len([line for line in status.split('\n') if line.startswith(' M ')])
        untracked_files = len([line for line in status.split('\n') if line.startswith('??')])

        return {
            'branch': status.split('\n')[0].split('...')[0].replace('## ', ''),
            'staged_changes': changes_staged,
            'unstaged_changes': changes_unstaged,
            'untracked_files': untracked_files
        }
    except subprocess.CalledProcessError as e:
        raise RuntimeError(f"저장소 상태 조회 실패: {e.stderr}")