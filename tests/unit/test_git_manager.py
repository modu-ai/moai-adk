# @TEST:CORE-GIT-001 | SPEC: SPEC-CORE-GIT-001.md
"""GitManager 클래스 테스트"""

import pytest
from pathlib import Path
from git import Repo
from moai_adk.core.git.manager import GitManager


@pytest.fixture
def temp_git_repo(tmp_path):
    """임시 Git 저장소 fixture"""
    repo_path = tmp_path / "test_repo"
    repo_path.mkdir()
    repo = Repo.init(repo_path)

    # 초기 커밋 생성 (Git이 HEAD를 가지려면 필요)
    (repo_path / "README.md").write_text("# Test Repo")
    repo.index.add(["README.md"])
    repo.index.commit("Initial commit")

    # develop 브랜치 생성
    develop = repo.create_head("develop")
    develop.checkout()

    return repo_path


class TestGitManager:
    """GitManager 클래스 테스트"""

    def test_should_detect_git_repository(self, temp_git_repo):
        """TEST-GIT-001: Git 저장소 여부 확인"""
        manager = GitManager(str(temp_git_repo))
        assert manager.is_repo() is True

    def test_should_return_false_for_non_git_directory(self, tmp_path):
        """TEST-GIT-002: 비 Git 디렉토리에서 False 반환"""
        non_git_dir = tmp_path / "not_a_repo"
        non_git_dir.mkdir()
        manager = GitManager(str(non_git_dir))
        assert manager.is_repo() is False

    def test_should_get_current_branch_name(self, temp_git_repo):
        """TEST-GIT-003: 현재 브랜치명 반환"""
        manager = GitManager(str(temp_git_repo))
        assert manager.current_branch() == "develop"

    def test_should_detect_dirty_state(self, temp_git_repo):
        """TEST-GIT-004: 작업 디렉토리 변경사항 감지"""
        manager = GitManager(str(temp_git_repo))

        # 초기 상태는 clean
        assert manager.is_dirty() is False

        # 파일 수정
        (temp_git_repo / "README.md").write_text("# Modified")
        assert manager.is_dirty() is True

    def test_should_create_and_checkout_new_branch(self, temp_git_repo):
        """TEST-GIT-005: 새 브랜치 생성 및 전환"""
        manager = GitManager(str(temp_git_repo))

        manager.create_branch("feature/test-branch", from_branch="develop")

        assert manager.current_branch() == "feature/test-branch"

    def test_should_commit_changes(self, temp_git_repo):
        """TEST-GIT-006: 파일 스테이징 및 커밋"""
        manager = GitManager(str(temp_git_repo))

        # 파일 생성
        test_file = temp_git_repo / "test.txt"
        test_file.write_text("test content")

        # 커밋
        manager.commit("Test commit", files=["test.txt"])

        # 커밋 확인
        last_commit = manager.repo.head.commit
        assert last_commit.message == "Test commit"

    def test_should_commit_all_changes_when_no_files_specified(self, temp_git_repo):
        """TEST-GIT-007: 파일 미지정 시 모든 변경사항 커밋"""
        manager = GitManager(str(temp_git_repo))

        # 여러 파일 생성
        (temp_git_repo / "file1.txt").write_text("content1")
        (temp_git_repo / "file2.txt").write_text("content2")

        # files 인자 없이 커밋
        manager.commit("Commit all")

        # 모든 파일이 커밋되었는지 확인
        assert manager.is_dirty() is False

    def test_should_push_to_remote_with_upstream(self, temp_git_repo):
        """TEST-GIT-008: 원격 저장소에 푸시 (upstream 설정)"""
        # 이 테스트는 실제 원격 저장소 없이는 실행 불가
        # 모킹을 통한 검증으로 대체 가능
        manager = GitManager(str(temp_git_repo))

        # GitPython의 push 메서드를 모킹해야 함
        # 실제 원격 없이는 테스트 불가하므로 skip
        pytest.skip("Requires remote repository setup")
