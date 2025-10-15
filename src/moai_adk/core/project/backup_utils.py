# @CODE:INIT-003:BACKUP | SPEC: .moai/specs/SPEC-INIT-003/spec.md | TEST: tests/unit/test_backup_utils.py
"""백업 유틸리티 모듈 (SPEC-INIT-003 v0.3.0)

Selective Backup 전략:
- 필요한 파일만 백업 (OR 조건)
- 백업 경로: .moai/backups/{timestamp}/ (v0.3.0)
"""

from datetime import datetime
from pathlib import Path


# 백업 대상 파일/디렉토리 (OR 조건 - 하나라도 있으면 백업)
BACKUP_TARGETS = [
    ".moai/config.json",
    ".moai/project/",
    ".moai/memory/",
    ".claude/",
    "CLAUDE.md",
]

# 사용자 데이터 보호 경로 (백업에서 제외)
PROTECTED_PATHS = [
    ".moai/specs/",
    ".moai/reports/",
]


def has_any_moai_files(project_path: Path) -> bool:
    """MoAI-ADK 파일 존재 여부 확인 (OR 조건)

    Args:
        project_path: 프로젝트 경로

    Returns:
        하나라도 존재하면 True
    """
    for target in BACKUP_TARGETS:
        target_path = project_path / target
        if target_path.exists():
            return True
    return False


def get_backup_targets(project_path: Path) -> list[str]:
    """백업 대상 파일/디렉토리 목록 반환

    Args:
        project_path: 프로젝트 경로

    Returns:
        존재하는 백업 대상 목록
    """
    targets: list[str] = []
    for target in BACKUP_TARGETS:
        target_path = project_path / target
        if target_path.exists():
            targets.append(target)
    return targets


def generate_backup_dir_name() -> str:
    """타임스탬프 기반 백업 디렉토리명 생성 (v0.3.0)

    Returns:
        YYYYMMDD-HHMMSS 형식 (순수 타임스탬프)
        Note: .moai/backups/ 경로는 호출자가 추가
    """
    timestamp = datetime.now().strftime("%Y%m%d-%H%M%S")
    return timestamp


def is_protected_path(rel_path: Path) -> bool:
    """보호 경로 여부 확인

    Args:
        rel_path: 상대 경로

    Returns:
        보호 경로이면 True
    """
    rel_str = str(rel_path).replace("\\", "/")
    return any(
        rel_str.startswith(p.lstrip("./").rstrip("/")) for p in PROTECTED_PATHS
    )
