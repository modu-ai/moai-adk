# @TEST:CORE-GIT-001 | SPEC: SPEC-CORE-GIT-001.md
"""
GitPython 기반 Git 관리 테스트.

SPEC: .moai/specs/SPEC-CORE-GIT-001/spec.md
"""

import pytest
from pathlib import Path
import git
from unittest.mock import MagicMock, patch

from moai_adk.core.git import GitManager, generate_branch_name, format_commit_message


@pytest.fixture
def git_manager(tmp_path):
    """임시 Git 저장소 생성 및 GitManager 반환."""
    # 임시 Git 저장소 초기화
    repo = git.Repo.init(tmp_path)

    # 초기 커밋 생성 (필수 - Git 저장소는 최소 1개 커밋 필요)
    readme_path = tmp_path / "README.md"
    readme_path.write_text("# Test Repository")
    repo.index.add(["README.md"])
    repo.index.commit("Initial commit")

    # GitManager 인스턴스 반환
    return GitManager(str(tmp_path))


class TestGitManager:
    """GitManager 클래스 테스트."""

    def test_is_repo(self, git_manager):
        """Git 저장소 여부 확인 테스트."""
        # GIVEN: Git 저장소가 초기화된 상태
        # WHEN: is_repo() 호출
        result = git_manager.is_repo()

        # THEN: True 반환
        assert result is True

    def test_current_branch(self, git_manager):
        """현재 브랜치명 조회 테스트."""
        # GIVEN: Git 저장소가 초기화된 상태 (기본 브랜치 존재)
        # WHEN: current_branch() 호출
        branch_name = git_manager.current_branch()

        # THEN: 브랜치명 반환 (master 또는 main)
        assert branch_name in ["master", "main"]

    def test_is_dirty_clean(self, git_manager):
        """Clean 상태 확인 테스트."""
        # GIVEN: 변경사항이 없는 상태
        # WHEN: is_dirty() 호출
        result = git_manager.is_dirty()

        # THEN: False 반환 (clean 상태)
        assert result is False

    def test_is_dirty_modified(self, git_manager, tmp_path):
        """Modified 상태 확인 테스트."""
        # GIVEN: 파일 수정
        test_file = tmp_path / "test.txt"
        test_file.write_text("Modified content")
        git_manager.repo.index.add(["test.txt"])

        # WHEN: is_dirty() 호출
        result = git_manager.is_dirty()

        # THEN: True 반환 (dirty 상태)
        assert result is True

    def test_create_branch(self, git_manager):
        """브랜치 생성 테스트."""
        # GIVEN: Git 저장소가 초기화된 상태
        new_branch = "feature/SPEC-TEST-001"

        # WHEN: create_branch() 호출
        git_manager.create_branch(new_branch, from_branch=None)

        # THEN: 새 브랜치가 생성되고 전환됨
        assert git_manager.current_branch() == new_branch

    def test_commit(self, git_manager, tmp_path):
        """커밋 생성 테스트."""
        # GIVEN: 새 파일 생성
        test_file = tmp_path / "new_file.txt"
        test_file.write_text("New content")

        # WHEN: commit() 호출
        commit_message = "test: add new file"
        git_manager.commit(commit_message, files=["new_file.txt"])

        # THEN: 커밋이 생성됨
        latest_commit = git_manager.repo.head.commit
        assert latest_commit.message.strip() == commit_message

    def test_commit_all_changes(self, git_manager, tmp_path):
        """모든 변경사항 커밋 테스트 (files=None)."""
        # GIVEN: 여러 파일 변경
        file1 = tmp_path / "file1.txt"
        file2 = tmp_path / "file2.txt"
        file1.write_text("Content 1")
        file2.write_text("Content 2")

        # WHEN: commit() 호출 (files=None)
        commit_message = "test: add multiple files"
        git_manager.commit(commit_message, files=None)

        # THEN: 모든 파일이 커밋됨
        latest_commit = git_manager.repo.head.commit
        assert latest_commit.message.strip() == commit_message

    def test_create_branch_from_specific_branch(self, git_manager):
        """특정 브랜치로부터 새 브랜치 생성 테스트."""
        # GIVEN: develop 브랜치 생성
        git_manager.create_branch("develop", from_branch=None)

        # 원래 브랜치로 돌아가기
        original_branch = "main" if "main" in [b.name for b in git_manager.repo.branches] else "master"
        git_manager.git.checkout(original_branch)

        # WHEN: develop으로부터 feature 브랜치 생성
        git_manager.create_branch("feature/TEST", from_branch="develop")

        # THEN: 새 브랜치가 생성됨
        assert git_manager.current_branch() == "feature/TEST"

    def test_is_repo_exception_handling(self, tmp_path):
        """Git 저장소가 아닐 때 예외 처리 테스트."""
        # GIVEN: Git 저장소가 아닌 디렉토리
        non_git_dir = tmp_path / "non_git"
        non_git_dir.mkdir()

        # WHEN/THEN: InvalidGitRepositoryError 발생
        with pytest.raises(git.exc.InvalidGitRepositoryError):
            GitManager(str(non_git_dir))



class TestBranchNaming:
    """브랜치 네이밍 유틸리티 테스트."""

    def test_generate_branch_name(self):
        """브랜치명 생성 유틸리티 테스트."""
        # GIVEN: SPEC ID
        spec_id = "AUTH-001"

        # WHEN: generate_branch_name() 호출
        branch_name = generate_branch_name(spec_id)

        # THEN: feature/SPEC-XXX 형식 반환
        assert branch_name == "feature/SPEC-AUTH-001"


class TestCommitMessage:
    """커밋 메시지 포맷팅 테스트."""

    def test_format_commit_message_ko(self):
        """한국어 커밋 메시지 테스트."""
        # GIVEN: 한국어 locale
        stage = "red"
        description = "사용자 인증 테스트 작성"

        # WHEN: format_commit_message() 호출
        message = format_commit_message(stage, description, locale="ko")

        # THEN: 한국어 템플릿 적용
        assert message == "🔴 RED: 사용자 인증 테스트 작성"

    def test_format_commit_message_en(self):
        """영어 커밋 메시지 테스트."""
        # GIVEN: 영어 locale
        stage = "green"
        description = "Implement user authentication"

        # WHEN: format_commit_message() 호출
        message = format_commit_message(stage, description, locale="en")

        # THEN: 영어 템플릿 적용
        assert message == "🟢 GREEN: Implement user authentication"

    def test_format_commit_message_refactor(self):
        """REFACTOR 단계 커밋 메시지 테스트."""
        # GIVEN: REFACTOR 단계
        stage = "refactor"
        description = "코드 구조 개선"

        # WHEN: format_commit_message() 호출
        message = format_commit_message(stage, description, locale="ko")

        # THEN: REFACTOR 이모지 포함
        assert message == "♻️ REFACTOR: 코드 구조 개선"

    def test_format_commit_message_docs(self):
        """DOCS 단계 커밋 메시지 테스트."""
        # GIVEN: DOCS 단계
        stage = "docs"
        description = "API 문서 업데이트"

        # WHEN: format_commit_message() 호출
        message = format_commit_message(stage, description, locale="ko")

        # THEN: DOCS 이모지 포함
        assert message == "📝 DOCS: API 문서 업데이트"

    def test_format_commit_message_invalid_stage(self):
        """잘못된 stage 입력 테스트."""
        # GIVEN: 잘못된 stage
        stage = "invalid"
        description = "Test"

        # WHEN/THEN: ValueError 발생
        with pytest.raises(ValueError, match="Invalid stage"):
            format_commit_message(stage, description)

    def test_format_commit_message_unsupported_locale(self):
        """지원하지 않는 locale 입력 시 기본값(en) 사용 테스트."""
        # GIVEN: 지원하지 않는 locale
        stage = "red"
        description = "Test description"

        # WHEN: format_commit_message() 호출 (unsupported locale)
        message = format_commit_message(stage, description, locale="unsupported")

        # THEN: 영어 템플릿 사용 (fallback)
        assert message == "🔴 RED: Test description"
