"""
Git 전략 간소화 및 워크플로우 개선 테스트 모듈

@TEST:UNIT-GIT-LOCK - Git 잠금 시스템 테스트
@TEST:UNIT-GIT-MODE - 개인 모드 Git 전략 테스트
@TEST:UNIT-WORKFLOW - 워크플로우 커맨드 개선 테스트
@TEST:UNIT-MANAGER - git-manager 업데이트 테스트

TDD RED 단계: 모든 테스트는 현재 구현되지 않은 기능으로 인해 실패합니다.
"""

import os
import tempfile
from pathlib import Path
from typing import Generator
from unittest.mock import Mock, patch
import pytest


class TestGitLockSystem:
    """@TEST:UNIT-GIT-LOCK - Git 잠금 시스템 테스트"""

    def test_git_lock_creation(self, temp_dir: Path):
        """Git 잠금 파일이 생성되는지 테스트

        Given: 잠금 파일이 없는 상태
        When: Git 작업을 시작할 때
        Then: .moai/locks/git.lock 파일이 생성되어야 함
        """
        from moai_adk.core.git_lock_manager import GitLockManager

        # Given: 잠금 파일이 없는 상태
        project_dir = temp_dir / "test_project"
        project_dir.mkdir()
        moai_dir = project_dir / ".moai"
        moai_dir.mkdir()
        locks_dir = moai_dir / "locks"

        lock_manager = GitLockManager(project_dir)
        lock_file = locks_dir / "git.lock"

        assert not lock_file.exists()

        # When: Git 작업을 시작할 때
        with lock_manager.acquire_lock():
            # Then: .moai/locks/git.lock 파일이 생성되어야 함
            assert lock_file.exists()
            assert lock_file.read_text().strip()

        # 작업 완료 후에는 잠금 해제
        assert not lock_file.exists()

    def test_git_lock_prevents_concurrent_work(self, temp_dir: Path):
        """잠금 파일이 동시 작업을 방지하는지 테스트

        Given: 이미 잠금 파일이 존재할 때
        When: 다른 Git 작업을 시도할 때
        Then: 작업이 차단되고 적절한 메시지가 표시되어야 함
        """
        from moai_adk.core.git_lock_manager import GitLockManager
        from moai_adk.core.exceptions import GitLockedException

        # Given: 이미 잠금 파일이 존재할 때
        project_dir = temp_dir / "test_project"
        project_dir.mkdir()
        moai_dir = project_dir / ".moai"
        moai_dir.mkdir()

        lock_manager1 = GitLockManager(project_dir)
        lock_manager2 = GitLockManager(project_dir)

        # 첫 번째 작업이 잠금을 획득
        with lock_manager1.acquire_lock():
            # When: 다른 Git 작업을 시도할 때
            # Then: 작업이 차단되고 적절한 메시지가 표시되어야 함
            with pytest.raises(GitLockedException) as exc_info:
                lock_manager2.acquire_lock(wait=False)

            assert "Git 작업이 이미 진행 중입니다" in str(exc_info.value)

    def test_git_lock_cleanup(self, temp_dir: Path):
        """Git 작업 완료 후 잠금 파일이 정리되는지 테스트

        Given: Git 작업이 진행 중인 상태
        When: Git 작업이 완료될 때
        Then: 잠금 파일이 자동으로 삭제되어야 함
        """
        from moai_adk.core.git_lock_manager import GitLockManager

        # Given: Git 작업이 진행 중인 상태
        project_dir = temp_dir / "test_project"
        project_dir.mkdir()
        moai_dir = project_dir / ".moai"
        moai_dir.mkdir()

        lock_manager = GitLockManager(project_dir)
        lock_file = moai_dir / "locks" / "git.lock"

        # 작업 시작 시 잠금 파일 생성 확인
        with lock_manager.acquire_lock():
            assert lock_file.exists()

        # When: Git 작업이 완료될 때
        # Then: 잠금 파일이 자동으로 삭제되어야 함
        assert not lock_file.exists()


class TestPersonalModeGitStrategy:
    """@TEST:UNIT-GIT-MODE - 개인 모드 Git 전략 테스트"""

    def test_personal_mode_direct_commit(self, temp_dir: Path):
        """개인 모드에서 main 브랜치 직접 커밋 테스트

        Given: 개인 모드 설정이 활성화된 상태
        When: spec/build 작업을 수행할 때
        Then: 별도 브랜치 없이 main에서 직접 커밋되어야 함
        """
        from moai_adk.core.git_strategy import PersonalGitStrategy
        from moai_adk.core.config_manager import ConfigManager

        # Given: 개인 모드 설정이 활성화된 상태
        project_dir = temp_dir / "test_project"
        project_dir.mkdir()

        config = ConfigManager(project_dir)
        config.set_mode("personal")

        git_strategy = PersonalGitStrategy(project_dir, config)
        current_branch = git_strategy.get_current_branch()

        # When: spec/build 작업을 수행할 때
        with git_strategy.work_context("test-feature"):
            # Then: 별도 브랜치 없이 main에서 직접 커밋되어야 함
            branch_after_work = git_strategy.get_current_branch()
            assert branch_after_work == current_branch

    def test_team_mode_feature_branch(self, temp_dir: Path):
        """팀 모드에서 feature 브랜치 생성 테스트

        Given: 팀 모드 설정이 활성화된 상태
        When: spec/build 작업을 수행할 때
        Then: feature 브랜치가 생성되고 해당 브랜치에서 작업되어야 함
        """
        from moai_adk.core.git_strategy import TeamGitStrategy
        from moai_adk.core.config_manager import ConfigManager

        # Given: 팀 모드 설정이 활성화된 상태
        project_dir = temp_dir / "test_project"
        project_dir.mkdir()

        config = ConfigManager(project_dir)
        config.set_mode("team")

        git_strategy = TeamGitStrategy(project_dir, config)
        main_branch = git_strategy.get_current_branch()

        # When: spec/build 작업을 수행할 때
        with git_strategy.work_context("test-feature"):
            # Then: feature 브랜치가 생성되고 해당 브랜치에서 작업되어야 함
            current_branch = git_strategy.get_current_branch()
            assert current_branch.startswith("feature/")
            assert "test-feature" in current_branch
            assert current_branch != main_branch


class TestWorkflowCommandImprovement:
    """@TEST:UNIT-WORKFLOW - 워크플로우 커맨드 개선 테스트"""

    def test_moai_1_spec_branch_skip_option(self, temp_dir: Path):
        """/moai:1-spec 브랜치 생성 스킵 옵션 테스트

        Given: 개인 모드에서 브랜치 스킵 옵션이 설정된 상태
        When: /moai:1-spec을 실행할 때
        Then: 브랜치 생성이 스킵되고 현재 브랜치에서 작업되어야 함
        """
        from moai_adk.commands.spec_command import SpecCommand
        from moai_adk.core.config_manager import ConfigManager

        # Given: 개인 모드에서 브랜치 스킵 옵션이 설정된 상태
        project_dir = temp_dir / "test_project"
        project_dir.mkdir()
        moai_dir = project_dir / ".moai"
        moai_dir.mkdir()

        config = ConfigManager(project_dir)
        config.set_mode("personal")
        config.set_option("skip_branch_creation", True)

        spec_command = SpecCommand(project_dir, config)

        # When: /moai:1-spec을 실행할 때
        spec_command.execute(
            spec_name="test-spec", description="테스트 명세", skip_branch=True
        )

        # Then: SPEC 파일이 생성되었는지 확인
        spec_file = moai_dir / "specs" / "test-spec.md"
        assert spec_file.exists()

    def test_moai_2_build_lock_check(self, temp_dir: Path):
        """/moai:2-build Git 작업 전 잠금 확인 테스트

        Given: 다른 Git 작업이 진행 중인 상태 (잠금 파일 존재)
        When: /moai:2-build를 실행할 때
        Then: 잠금 확인 후 대기하거나 적절한 메시지를 표시해야 함
        """
        from moai_adk.commands.build_command import BuildCommand
        from moai_adk.core.git_lock_manager import GitLockManager
        from moai_adk.core.exceptions import GitLockedException

        # Given: 다른 Git 작업이 진행 중인 상태
        project_dir = temp_dir / "test_project"
        project_dir.mkdir()
        moai_dir = project_dir / ".moai"
        moai_dir.mkdir()

        lock_manager = GitLockManager(project_dir)
        build_command = BuildCommand(project_dir)

        # 다른 작업이 잠금을 획득한 상태
        with lock_manager.acquire_lock():
            # When: /moai:2-build를 실행할 때
            # Then: 잠금 확인 후 대기하거나 적절한 메시지를 표시해야 함
            with pytest.raises(GitLockedException) as exc_info:
                build_command.execute(spec_name="test-spec", wait_for_lock=False)

            assert "잠금 파일이 감지되었습니다" in str(exc_info.value)


class TestGitManagerUpdate:
    """@TEST:UNIT-MANAGER - git-manager 업데이트 테스트"""

    def test_git_manager_lock_integration(self, temp_dir: Path):
        """git-manager의 잠금 시스템 통합 테스트

        Given: git-manager가 새로운 잠금 전략을 지원하는 상태
        When: Git 작업을 요청할 때
        Then: 자동으로 잠금을 획득하고 작업 완료 후 해제해야 함
        """
        from moai_adk.core.git_manager import GitManager

        # Given: git-manager가 새로운 잠금 전략을 지원하는 상태
        project_dir = temp_dir / "test_project"
        project_dir.mkdir()
        moai_dir = project_dir / ".moai"
        moai_dir.mkdir()

        git_manager = GitManager(project_dir)
        lock_file = moai_dir / "locks" / "git.lock"

        # When: Git 작업을 요청할 때
        git_manager.commit_with_lock("Add test file", ["test.txt"])

        # Then: 작업 완료 후 잠금이 해제되어야 함
        assert not lock_file.exists()

    def test_git_manager_mode_detection(self, temp_dir: Path):
        """git-manager의 모드별 전략 감지 테스트

        Given: 개인/팀 모드 설정이 있는 상태
        When: git-manager가 실행될 때
        Then: 올바른 모드를 감지하고 해당 전략을 적용해야 함
        """
        from moai_adk.core.git_manager import GitManager
        from moai_adk.core.config_manager import ConfigManager

        # Given: 개인/팀 모드 설정이 있는 상태
        project_dir = temp_dir / "test_project"
        project_dir.mkdir()

        config = ConfigManager(project_dir)
        config.set_mode("personal")

        # When: git-manager가 실행될 때
        git_manager = GitManager(project_dir, config)

        # Then: 올바른 모드를 감지하고 해당 전략을 적용해야 함
        assert git_manager.get_mode() == "personal"
        assert git_manager.strategy.__class__.__name__ == "PersonalGitStrategy"
