# @CODE:TEMPLATE-001 | SPEC: SPEC-INIT-003.md | Chain: TEMPLATE-001
"""템플릿 백업 관리 클래스 (SPEC-INIT-003 v0.3.0).

템플릿 업데이트 시 사용자 데이터를 보호하기 위한 백업 생성 및 관리.
"""

from __future__ import annotations

import shutil
from datetime import datetime
from pathlib import Path


class TemplateBackup:
    """템플릿 백업 생성 및 관리."""

    # 백업 제외 경로 (사용자 데이터 보호)
    BACKUP_EXCLUDE_DIRS = [
        "specs",  # 사용자 SPEC 문서
        "reports",  # 사용자 리포트
    ]

    def __init__(self, target_path: Path) -> None:
        """초기화.

        Args:
            target_path: 프로젝트 경로 (절대 경로)
        """
        self.target_path = target_path.resolve()

    def has_existing_files(self) -> bool:
        """백업 필요한 기존 파일 존재 여부 확인.

        백업 정책:
        - .moai/, .claude/, CLAUDE.md 중 **1개라도 존재하면 백업 생성**
        - 백업 경로: .moai-backup/YYYYMMDD-HHMMSS/

        Returns:
            True if 백업 필요 (파일 1개 이상 존재)
        """
        return any(
            (self.target_path / item).exists()
            for item in [".moai", ".claude", "CLAUDE.md"]
        )

    def create_backup(self) -> Path:
        """타임스탬프 기반 백업 생성.

        백업 대상:
        - .moai/ (specs/, reports/ 제외)
        - .claude/
        - CLAUDE.md

        Returns:
            백업 경로 (예: .moai-backup/20250110-143025/)
        """
        timestamp = datetime.now().strftime("%Y%m%d-%H%M%S")
        backup_path = self.target_path / ".moai-backup" / timestamp
        backup_path.mkdir(parents=True, exist_ok=True)

        # 백업 대상 복사
        for item in [".moai", ".claude", "CLAUDE.md"]:
            src = self.target_path / item
            if not src.exists():
                continue

            dst = backup_path / item

            if item == ".moai":
                # 보호 경로 제외하고 복사
                self._copy_exclude_protected(src, dst)
            elif src.is_dir():
                shutil.copytree(src, dst, dirs_exist_ok=True)
            else:
                shutil.copy2(src, dst)

        return backup_path

    def _copy_exclude_protected(self, src: Path, dst: Path) -> None:
        """보호 경로를 제외하고 백업 복사.

        백업 제외 경로:
        - .moai/specs/ (사용자 SPEC)
        - .moai/reports/ (사용자 리포트)

        Args:
            src: 소스 디렉토리
            dst: 대상 디렉토리
        """
        dst.mkdir(parents=True, exist_ok=True)

        for item in src.rglob("*"):
            rel_path = item.relative_to(src)
            rel_path_str = str(rel_path)

            # 백업 제외 경로 검사
            if any(
                rel_path_str.startswith(exclude_dir)
                for exclude_dir in self.BACKUP_EXCLUDE_DIRS
            ):
                continue

            dst_item = dst / rel_path
            if item.is_file():
                dst_item.parent.mkdir(parents=True, exist_ok=True)
                shutil.copy2(item, dst_item)
            elif item.is_dir():
                dst_item.mkdir(parents=True, exist_ok=True)
