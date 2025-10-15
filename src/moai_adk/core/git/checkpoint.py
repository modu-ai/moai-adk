# @CODE:CHECKPOINT-EVENT-001 | SPEC: SPEC-CHECKPOINT-EVENT-001.md | TEST: tests/unit/test_checkpoint.py
"""
Checkpoint Manager - Event-Driven Checkpoint 시스템.

SPEC: .moai/specs/SPEC-CHECKPOINT-EVENT-001/spec.md
"""

from datetime import datetime
from pathlib import Path
from typing import Optional

import git

from moai_adk.core.git.branch_manager import BranchManager
from moai_adk.core.git.event_detector import EventDetector


class CheckpointManager:
    """Event-Driven Checkpoint 생성 및 복구를 관리하는 매니저."""

    def __init__(self, repo: git.Repo, project_root: Path):
        """
        CheckpointManager 초기화.

        Args:
            repo: GitPython Repo 인스턴스
            project_root: 프로젝트 루트 경로
        """
        self.repo = repo
        self.project_root = project_root
        self.event_detector = EventDetector()
        self.branch_manager = BranchManager(repo)
        self.log_file = project_root / ".moai" / "checkpoints.log"

        # 로그 디렉토리 생성
        self.log_file.parent.mkdir(parents=True, exist_ok=True)

    def create_checkpoint_if_risky(
        self,
        operation: str,
        deleted_files: Optional[list[str]] = None,
        renamed_files: Optional[list[tuple[str, str]]] = None,
        modified_files: Optional[list[Path]] = None,
    ) -> Optional[str]:
        """
        위험한 작업 감지 시 checkpoint 생성.

        SPEC 요구사항: 위험한 작업이 감지되면 자동으로 checkpoint 생성

        Args:
            operation: 작업 유형
            deleted_files: 삭제될 파일 목록
            renamed_files: 이름 변경될 파일 목록
            modified_files: 수정될 파일 목록

        Returns:
            생성된 checkpoint ID (branch 이름), 안전한 작업이면 None
        """
        is_risky = False

        # 대규모 파일 삭제 확인
        if deleted_files and self.event_detector.is_risky_deletion(deleted_files):
            is_risky = True

        # 복잡한 리팩토링 확인
        if renamed_files and self.event_detector.is_risky_refactoring(renamed_files):
            is_risky = True

        # 중요 파일 수정 확인
        if modified_files:
            for file_path in modified_files:
                if self.event_detector.is_critical_file(file_path):
                    is_risky = True
                    break

        if not is_risky:
            return None

        # Checkpoint 생성
        checkpoint_id = self.branch_manager.create_checkpoint_branch(operation)

        # 메타데이터 기록
        self._log_checkpoint(checkpoint_id, operation)

        return checkpoint_id

    def restore_checkpoint(self, checkpoint_id: str) -> None:
        """
        특정 checkpoint로 복구.

        SPEC 요구사항: 복구 전에 현재 상태를 새로운 checkpoint로 저장

        Args:
            checkpoint_id: 복구할 checkpoint ID (branch 이름)
        """
        # 복구 전 현재 상태를 safety checkpoint로 저장
        safety_checkpoint = self.branch_manager.create_checkpoint_branch("restore")
        self._log_checkpoint(safety_checkpoint, "restore", is_safety=True)

        # Checkpoint branch로 체크아웃
        self.repo.git.checkout(checkpoint_id)

    def list_checkpoints(self) -> list[str]:
        """
        모든 checkpoint 목록 조회.

        Returns:
            checkpoint ID 목록
        """
        return self.branch_manager.list_checkpoint_branches()

    def _log_checkpoint(
        self,
        checkpoint_id: str,
        operation: str,
        is_safety: bool = False
    ) -> None:
        """
        Checkpoint 메타데이터를 로그 파일에 기록.

        SPEC 요구사항: .moai/checkpoints.log에 메타데이터 기록

        Args:
            checkpoint_id: Checkpoint ID
            operation: 작업 유형
            is_safety: Safety checkpoint 여부
        """
        timestamp = datetime.now().isoformat()

        log_entry = f"""---
checkpoint_id: {checkpoint_id}
operation: {operation}
timestamp: {timestamp}
is_safety: {is_safety}
---
"""

        # Append mode로 로그 기록
        with open(self.log_file, "a", encoding="utf-8") as f:
            f.write(log_entry)
