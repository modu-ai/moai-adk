"""
Checkpoint 시스템 테스트.

Event-Driven Checkpoint 생성 및 복구 기능을 테스트합니다.

SPEC: .moai/specs/SPEC-CHECKPOINT-EVENT-001/spec.md
"""

import pytest

from moai_adk.core.git.checkpoint import CheckpointManager


class TestCheckpointManager:
    """CheckpointManager event-driven checkpoint 테스트."""

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
    def manager(self, temp_git_repo, tmp_path):
        """CheckpointManager 인스턴스."""
        return CheckpointManager(temp_git_repo, tmp_path)

    # TEST-CHECKPOINT-EVENT-007: 위험 작업 감지 시 checkpoint 생성
    def test_should_create_checkpoint_on_risky_operation(self, manager):
        """위험한 작업 감지 시 자동으로 checkpoint를 생성해야 한다."""
        # 대규모 파일 삭제 시뮬레이션
        deleted_files = [f"file{i}.py" for i in range(10)]
        checkpoint_id = manager.create_checkpoint_if_risky(operation="delete", deleted_files=deleted_files)
        assert checkpoint_id is not None
        assert checkpoint_id.startswith("before-delete-")

    def test_should_not_create_checkpoint_on_safe_operation(self, manager):
        """안전한 작업은 checkpoint를 생성하지 않아야 한다."""
        # 소규모 파일 삭제 시뮬레이션
        deleted_files = ["file1.py"]
        checkpoint_id = manager.create_checkpoint_if_risky(operation="delete", deleted_files=deleted_files)
        assert checkpoint_id is None

    # TEST-CHECKPOINT-EVENT-008: Checkpoint 메타데이터 기록
    def test_should_log_checkpoint_metadata(self, manager, tmp_path):
        """checkpoint 생성 시 메타데이터를 .moai/checkpoints.log에 기록해야 한다."""
        deleted_files = [f"file{i}.py" for i in range(10)]
        manager.create_checkpoint_if_risky(operation="delete", deleted_files=deleted_files)

        log_file = tmp_path / ".moai" / "checkpoints.log"
        assert log_file.exists()

        log_content = log_file.read_text()
        assert "before-delete-" in log_content
        assert "operation: delete" in log_content

    def test_should_include_timestamp_in_metadata(self, manager, tmp_path):
        """checkpoint 메타데이터에 timestamp가 포함되어야 한다."""
        deleted_files = [f"file{i}.py" for i in range(10)]
        manager.create_checkpoint_if_risky(operation="delete", deleted_files=deleted_files)

        log_file = tmp_path / ".moai" / "checkpoints.log"
        log_content = log_file.read_text()
        # ISO 8601 형식 timestamp 확인
        assert "timestamp:" in log_content

    # TEST-CHECKPOINT-EVENT-009: Checkpoint 복구
    def test_should_restore_from_checkpoint(self, manager, temp_git_repo, tmp_path):
        """특정 checkpoint로 복구 기능이 동작해야 한다."""
        # 초기 상태 저장
        test_file = tmp_path / "test.txt"
        original_content = test_file.read_text()

        # Checkpoint 생성
        deleted_files = [f"file{i}.py" for i in range(10)]
        checkpoint_id = manager.create_checkpoint_if_risky(operation="delete", deleted_files=deleted_files)

        # 파일 수정
        test_file.write_text("modified")
        temp_git_repo.index.add([str(test_file)])
        temp_git_repo.index.commit("Modify file")

        # Checkpoint로 복구
        manager.restore_checkpoint(checkpoint_id)

        # 원본 내용으로 복구되었는지 확인
        assert test_file.read_text() == original_content

    def test_should_create_safety_checkpoint_before_restore(self, manager):
        """복구 전에 현재 상태를 새로운 checkpoint로 저장해야 한다."""
        # 첫 번째 checkpoint 생성
        deleted_files = [f"file{i}.py" for i in range(10)]
        checkpoint_id = manager.create_checkpoint_if_risky(operation="delete", deleted_files=deleted_files)

        # 복구 실행
        manager.restore_checkpoint(checkpoint_id)

        # 복구 전 상태를 저장한 checkpoint가 생성되었는지 확인
        checkpoints = manager.list_checkpoints()
        safety_checkpoint = [c for c in checkpoints if "before-restore-" in c]
        assert len(safety_checkpoint) > 0
