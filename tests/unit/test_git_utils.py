# @TEST:CORE-GIT-001 | SPEC: SPEC-CORE-GIT-001.md
"""Git 유틸리티 함수 테스트"""

import pytest
import subprocess
from unittest.mock import patch, MagicMock
from pathlib import Path
from moai_adk.core.git.branch import generate_branch_name
from moai_adk.core.git.commit import format_commit_message
from moai_adk.core.git.pr import create_draft_pr, get_repo_status
from moai_adk.core.git.manager import GitManager


class TestBranchNaming:
    """브랜치 네이밍 테스트"""

    def test_should_generate_branch_name_from_spec_id(self):
        """TEST-GIT-101: SPEC ID로부터 브랜치명 생성"""
        branch_name = generate_branch_name("AUTH-001")
        assert branch_name == "feature/SPEC-AUTH-001"

    def test_should_handle_different_spec_formats(self):
        """TEST-GIT-102: 다양한 SPEC ID 형식 처리"""
        assert generate_branch_name("CORE-GIT-001") == "feature/SPEC-CORE-GIT-001"
        assert generate_branch_name("PY314-001") == "feature/SPEC-PY314-001"


class TestCommitMessage:
    """커밋 메시지 포맷팅 테스트"""

    def test_should_format_red_commit_message_in_korean(self):
        """TEST-GIT-201: RED 단계 커밋 메시지 (한국어)"""
        msg = format_commit_message("red", "사용자 인증 테스트 작성", locale="ko")
        assert msg == "🔴 RED: 사용자 인증 테스트 작성"

    def test_should_format_green_commit_message_in_korean(self):
        """TEST-GIT-202: GREEN 단계 커밋 메시지 (한국어)"""
        msg = format_commit_message("green", "사용자 인증 구현 완료", locale="ko")
        assert msg == "🟢 GREEN: 사용자 인증 구현 완료"

    def test_should_format_refactor_commit_message_in_korean(self):
        """TEST-GIT-203: REFACTOR 단계 커밋 메시지 (한국어)"""
        msg = format_commit_message("refactor", "인증 로직 리팩토링", locale="ko")
        assert msg == "♻️ REFACTOR: 인증 로직 리팩토링"

    def test_should_format_commit_message_in_english(self):
        """TEST-GIT-204: 영어 로케일 커밋 메시지"""
        msg = format_commit_message("red", "Add authentication test", locale="en")
        assert msg == "🔴 RED: Add authentication test"

    def test_should_fallback_to_english_for_unknown_locale(self):
        """TEST-GIT-205: 알 수 없는 로케일은 영어로 폴백"""
        msg = format_commit_message("green", "Implementation", locale="ja")
        assert msg == "🟢 GREEN: Implementation"


class TestDraftPR:
    """Draft PR 생성 테스트"""

    @patch("subprocess.run")
    def test_should_create_draft_pr_with_gh_cli(self, mock_run):
        """TEST-GIT-301: gh CLI로 Draft PR 생성"""
        # subprocess.run 모킹
        mock_run.return_value = MagicMock(
            stdout="https://github.com/user/repo/pull/123",
            returncode=0
        )

        pr_url = create_draft_pr(
            title="Test PR",
            body="Test description",
            base="develop"
        )

        # gh CLI 호출 확인
        mock_run.assert_called_once()
        call_args = mock_run.call_args[0][0]

        assert call_args[0] == "gh"
        assert call_args[1] == "pr"
        assert call_args[2] == "create"
        assert "--draft" in call_args
        assert pr_url == "https://github.com/user/repo/pull/123"

    @patch("subprocess.run")
    def test_should_raise_error_when_gh_cli_fails(self, mock_run):
        """TEST-GIT-302: gh CLI 실패 시 에러 발생"""
        mock_run.side_effect = subprocess.CalledProcessError(1, "gh")

        with pytest.raises(subprocess.CalledProcessError):
            create_draft_pr(
                title="Test PR",
                body="Test description"
            )


class TestRepoStatus:
    """저장소 상태 조회 테스트"""

    def test_should_return_repo_status(self, tmp_path):
        """TEST-GIT-401: 저장소 상태 정보 반환"""
        # 임시 Git 저장소 생성
        repo_path = tmp_path / "test_repo"
        repo_path.mkdir()
        from git import Repo
        repo = Repo.init(repo_path)

        # 초기 커밋
        (repo_path / "README.md").write_text("# Test")
        repo.index.add(["README.md"])
        repo.index.commit("Initial commit")

        # GitManager로 상태 조회
        manager = GitManager(str(repo_path))
        status = get_repo_status(manager)

        assert status["is_repo"] is True
        # Git 2.28+ 기본 브랜치는 main 또는 master
        assert status["current_branch"] in ["main", "master"]
        assert status["is_dirty"] is False
        assert isinstance(status["untracked_files"], list)
        assert isinstance(status["modified_files"], list)

    def test_should_detect_untracked_files(self, tmp_path):
        """TEST-GIT-402: 추적되지 않은 파일 감지"""
        # 임시 Git 저장소 생성
        repo_path = tmp_path / "test_repo"
        repo_path.mkdir()
        from git import Repo
        repo = Repo.init(repo_path)

        # 초기 커밋
        (repo_path / "README.md").write_text("# Test")
        repo.index.add(["README.md"])
        repo.index.commit("Initial commit")

        # 추적되지 않은 파일 생성
        (repo_path / "untracked.txt").write_text("new file")

        manager = GitManager(str(repo_path))
        status = get_repo_status(manager)

        assert "untracked.txt" in status["untracked_files"]
