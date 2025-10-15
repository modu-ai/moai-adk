# @CODE:CHECKPOINT-EVENT-001 | SPEC: SPEC-CHECKPOINT-EVENT-001.md | TEST: tests/unit/test_branch_manager.py
"""
Branch Manager - Local checkpoint branch 관리.

SPEC: .moai/specs/SPEC-CHECKPOINT-EVENT-001/spec.md
"""

from datetime import datetime

import git


class BranchManager:
    """Local checkpoint branch를 관리하는 매니저."""

    MAX_CHECKPOINTS = 10
    CHECKPOINT_PREFIX = "before-"

    def __init__(self, repo: git.Repo):
        """
        BranchManager 초기화.

        Args:
            repo: GitPython Repo 인스턴스
        """
        self.repo = repo
        self._old_branches: set[str] = set()

    def create_checkpoint_branch(self, operation: str) -> str:
        """
        Checkpoint branch 생성.

        SPEC 요구사항: before-{operation}-{timestamp} 형식으로 local branch 생성

        Args:
            operation: 작업 유형 (delete, refactor, merge 등)

        Returns:
            생성된 branch 이름
        """
        timestamp = datetime.now().strftime("%Y%m%d-%H%M%S")
        branch_name = f"{self.CHECKPOINT_PREFIX}{operation}-{timestamp}"

        # 현재 HEAD에서 branch 생성
        self.repo.create_head(branch_name)

        # FIFO 방식으로 오래된 checkpoint 삭제
        self._enforce_max_checkpoints()

        return branch_name

    def branch_exists(self, branch_name: str) -> bool:
        """
        Branch 존재 여부 확인.

        Args:
            branch_name: 확인할 branch 이름

        Returns:
            존재하면 True, 아니면 False
        """
        return branch_name in [head.name for head in self.repo.heads]

    def has_remote_tracking(self, branch_name: str) -> bool:
        """
        Remote tracking branch 존재 여부 확인.

        SPEC 요구사항: checkpoint는 local branch로만 존재해야 함

        Args:
            branch_name: 확인할 branch 이름

        Returns:
            remote tracking이 있으면 True, 아니면 False
        """
        try:
            branch = self.repo.heads[branch_name]
            return branch.tracking_branch() is not None
        except (IndexError, AttributeError):
            return False

    def list_checkpoint_branches(self) -> list[str]:
        """
        모든 checkpoint branch 목록 조회.

        Returns:
            checkpoint branch 이름 목록
        """
        return [
            head.name
            for head in self.repo.heads
            if head.name.startswith(self.CHECKPOINT_PREFIX)
        ]

    def mark_as_old(self, branch_name: str) -> None:
        """
        Branch를 오래된 것으로 표시 (테스트용).

        Args:
            branch_name: 표시할 branch 이름
        """
        self._old_branches.add(branch_name)

    def cleanup_old_checkpoints(self, max_count: int) -> None:
        """
        오래된 checkpoint 정리.

        SPEC 요구사항: 최대 개수 초과 시 FIFO 방식으로 삭제

        Args:
            max_count: 유지할 최대 checkpoint 개수
        """
        checkpoints = self.list_checkpoint_branches()

        # 오래된 순서대로 정렬 (mark_as_old로 표시된 것 우선)
        sorted_checkpoints = sorted(
            checkpoints,
            key=lambda name: (name not in self._old_branches, name)
        )

        # 초과분 삭제
        to_delete = sorted_checkpoints[: len(sorted_checkpoints) - max_count]
        for branch_name in to_delete:
            if branch_name in [head.name for head in self.repo.heads]:
                self.repo.delete_head(branch_name, force=True)

    def _enforce_max_checkpoints(self) -> None:
        """최대 checkpoint 개수 유지 (내부 메서드)."""
        checkpoints = self.list_checkpoint_branches()

        if len(checkpoints) > self.MAX_CHECKPOINTS:
            # 이름 순서대로 정렬 (timestamp 기준으로 오래된 것이 앞에 옴)
            sorted_checkpoints = sorted(checkpoints)
            to_delete = sorted_checkpoints[: len(sorted_checkpoints) - self.MAX_CHECKPOINTS]

            for branch_name in to_delete:
                self.repo.delete_head(branch_name, force=True)
