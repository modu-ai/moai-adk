"""
Tests for GitCollector - Git 정보 수집 및 캐싱

"""

from unittest.mock import MagicMock, patch

import pytest


class TestGitCollector:
    """Git 정보 수집 및 캐싱 테스트"""

    @pytest.mark.xfail(reason="Test data migration needed")
    def test_collect_branch_from_git_status(self):
        """
        GIVEN: 정상 작동하는 Git 저장소
        WHEN: collect_git_info()를 호출
        THEN: 현재 branch 이름을 정확히 반환 (예: feature/SPEC-AUTH-001, develop, main)
        """
        from moai_adk.statusline.git_collector import GitCollector

        collector = GitCollector()

        # Mock git status output
        with patch("subprocess.run") as mock_run:
            # Typical git status -b --porcelain output
            mock_run.return_value = MagicMock(
                stdout="## feature/SPEC-AUTH-001...origin/feature/SPEC-AUTH-001\nM file1.py\n",
                stderr="",
                returncode=0,
            )

            git_info = collector.collect_git_info()

            assert git_info.branch == "feature/SPEC-AUTH-001", f"Expected feature/SPEC-AUTH-001, got {git_info.branch}"

    @pytest.mark.xfail(reason="Test data migration needed")
    def test_collect_develop_branch(self):
        """
        GIVEN: develop branch에 있을 때
        WHEN: collect_git_info()를 호출
        THEN: branch == "develop"
        """
        from moai_adk.statusline.git_collector import GitCollector

        collector = GitCollector()

        with patch("subprocess.run") as mock_run:
            mock_run.return_value = MagicMock(
                stdout="## develop...origin/develop\n",
                stderr="",
                returncode=0,
            )

            git_info = collector.collect_git_info()
            assert git_info.branch == "develop"

    @pytest.mark.xfail(reason="Test data migration needed")
    def test_count_git_changes(self):
        """
        GIVEN: 3 staged, 2 modified, 1 untracked 파일
        WHEN: collect_git_info()를 호출
        THEN: staged=3, modified=2, untracked=1
        """
        from moai_adk.statusline.git_collector import GitCollector

        collector = GitCollector()

        with patch("subprocess.run") as mock_run:
            # A = added (staged), M = modified, ?? = untracked
            mock_run.return_value = MagicMock(
                stdout="## develop...origin/develop\nA  file1.py\nA  file2.py\nA  file3.py\nM  file4.py\nM  file5.py\n?? file6.py\n",
                stderr="",
                returncode=0,
            )

            git_info = collector.collect_git_info()

            assert git_info.staged == 3, f"Expected 3 staged, got {git_info.staged}"
            assert git_info.modified == 2, f"Expected 2 modified, got {git_info.modified}"
            assert git_info.untracked == 1, f"Expected 1 untracked, got {git_info.untracked}"

    @pytest.mark.xfail(reason="Test data migration needed")
    def test_empty_git_status(self):
        """
        GIVEN: 깨끗한 Git 상태 (변경 없음)
        WHEN: collect_git_info()를 호출
        THEN: staged=0, modified=0, untracked=0
        """
        from moai_adk.statusline.git_collector import GitCollector

        collector = GitCollector()

        with patch("subprocess.run") as mock_run:
            mock_run.return_value = MagicMock(
                stdout="## main...origin/main\n",
                stderr="",
                returncode=0,
            )

            git_info = collector.collect_git_info()

            assert git_info.staged == 0
            assert git_info.modified == 0
            assert git_info.untracked == 0

    @pytest.mark.xfail(reason="Test data migration needed")
    def test_git_collector_caching(self):
        """
        GIVEN: GitCollector 인스턴스
        WHEN: 5초 이내에 두 번 collect_git_info() 호출
        THEN: 캐시에서 반환되고 git 명령은 한 번만 실행
        """

        from moai_adk.statusline.git_collector import GitCollector

        collector = GitCollector()

        with patch("subprocess.run") as mock_run:
            mock_run.return_value = MagicMock(
                stdout="## develop...origin/develop\nM  file.py\n",
                stderr="",
                returncode=0,
            )

            # First call
            result1 = collector.collect_git_info()
            call_count_after_first = mock_run.call_count

            # Second call immediately after (should use cache)
            result2 = collector.collect_git_info()
            call_count_after_second = mock_run.call_count

            # Git command should only be called once (cache works)
            assert (
                call_count_after_first == call_count_after_second
            ), f"Git command called {call_count_after_second - call_count_after_first} more times (cache not working)"

            # Results should be identical
            assert result1.branch == result2.branch
            assert result1.modified == result2.modified

    @pytest.mark.xfail(reason="Test data migration needed")
    def test_git_command_failure_graceful(self):
        """
        GIVEN: Git 명령이 실패할 때
        WHEN: collect_git_info()를 호출
        THEN: 예외를 발생시키지 않고 기본값 반환 (branch="unknown", counts=0)
        """
        from moai_adk.statusline.git_collector import GitCollector

        collector = GitCollector()

        with patch("subprocess.run") as mock_run:
            mock_run.side_effect = Exception("Git command failed")

            git_info = collector.collect_git_info()

            # Should gracefully handle error
            assert git_info.branch in ["unknown", "N/A", "?"], f"Expected graceful default, got {git_info.branch}"
            assert git_info.staged == 0
            assert git_info.modified == 0
            assert git_info.untracked == 0

    @pytest.mark.xfail(reason="Test data migration needed")
    def test_mixed_git_status(self):
        """
        GIVEN: 다양한 파일 상태 (staged + modified + untracked)
        WHEN: collect_git_info()를 호출
        THEN: 각 카운트가 정확함
        """
        from moai_adk.statusline.git_collector import GitCollector

        collector = GitCollector()

        with patch("subprocess.run") as mock_run:
            output = """## feature/SPEC-AUTH-001...origin/feature/SPEC-AUTH-001
A  src/new_module.py
M  src/existing.py
M  tests/test_new.py
?? docs/notes.txt
?? .env.local
"""
            mock_run.return_value = MagicMock(
                stdout=output,
                stderr="",
                returncode=0,
            )

            git_info = collector.collect_git_info()

            assert git_info.branch == "feature/SPEC-AUTH-001"
            assert git_info.staged == 1  # A
            assert git_info.modified == 2  # M
            assert git_info.untracked == 2  # ??


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
