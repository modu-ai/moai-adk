"""
Branch Manager 테스트.

Local checkpoint branch 관리 기능을 테스트합니다.

SPEC: .moai/specs/SPEC-CHECKPOINT-EVENT-001/spec.md
"""

import pytest

from moai_adk.core.git.branch_manager import BranchManager


class TestBranchManager:
    """BranchManager local branch 관리 테스트."""

    @pytest.fixture
    def temp_git_repo(self, tmp_path):
        """임시 Git 저장소 생성."""
        import git

        repo = git.Repo.init(tmp_path)
        # 초기 커밋 생성
        test_file = tmp_path / "test.txt"
        test_file.write_text("initial")
        repo.index.add([str(test_file)])
        repo.index.commit("Initial commit")
        return repo

    @pytest.fixture
    def manager(self, temp_git_repo):
        """BranchManager 인스턴스."""
        return BranchManager(temp_git_repo)

    # TEST-CHECKPOINT-EVENT-004: Checkpoint branch 생성
    def test_should_create_checkpoint_branch(self, manager):
        """checkpoint branch를 before-{operation}-{timestamp} 형식으로 생성해야 한다."""
        branch_name = manager.create_checkpoint_branch("delete")
        assert branch_name.startswith("before-delete-")
        assert manager.branch_exists(branch_name) is True

    def test_should_prevent_remote_push(self, manager):
        """checkpoint branch는 local branch로만 존재하며 원격 push가 차단되어야 한다."""
        branch_name = manager.create_checkpoint_branch("merge")
        # remote tracking branch가 없어야 함
        assert manager.has_remote_tracking(branch_name) is False

    # TEST-CHECKPOINT-EVENT-005: Checkpoint branch 관리
    def test_should_maintain_max_checkpoints(self, manager):
        """최대 10개의 checkpoint를 유지하고 초과 시 FIFO 방식으로 삭제해야 한다."""
        # 11개 checkpoint 생성
        for i in range(11):
            manager.create_checkpoint_branch(f"operation-{i}")

        # 10개만 남아있어야 함
        checkpoint_branches = manager.list_checkpoint_branches()
        assert len(checkpoint_branches) == 10

    def test_should_delete_old_checkpoints_first(self, manager):
        """오래된 checkpoint부터 삭제되어야 한다 (FIFO)."""
        # 2개 생성 후 시간 차이 시뮬레이션
        first = manager.create_checkpoint_branch("first")
        second = manager.create_checkpoint_branch("second")

        # 첫 번째가 가장 오래된 것으로 표시
        manager.mark_as_old(first)

        # cleanup 실행
        manager.cleanup_old_checkpoints(max_count=1)

        # 두 번째만 남아있어야 함
        assert manager.branch_exists(first) is False
        assert manager.branch_exists(second) is True

    # TEST-CHECKPOINT-EVENT-006: Branch 이름 규칙
    def test_should_follow_naming_convention(self, manager):
        """checkpoint branch는 'before-*' 접두사를 가져야 한다."""
        operations = ["delete", "refactor", "merge"]
        for op in operations:
            branch_name = manager.create_checkpoint_branch(op)
            assert branch_name.startswith("before-")
            assert op in branch_name

    def test_should_include_timestamp_in_branch_name(self, manager):
        """branch 이름에 timestamp가 포함되어야 한다."""
        branch_name = manager.create_checkpoint_branch("test")
        # YYYYMMDD-HHMMSS 형식 확인
        timestamp_part = branch_name.split("-")[-2:]
        assert len(timestamp_part) == 2
        assert len(timestamp_part[0]) == 8  # YYYYMMDD
        assert len(timestamp_part[1]) == 6  # HHMMSS
