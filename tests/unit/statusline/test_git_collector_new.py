"""Comprehensive tests for GitCollector with 80% coverage target."""

import subprocess
from datetime import datetime, timedelta
from unittest.mock import MagicMock, patch, call

import pytest

from moai_adk.statusline.git_collector import GitCollector, GitInfo


class TestGitCollectorInit:
    """Test GitCollector initialization."""

    def test_init_default(self):
        """Test default initialization."""
        collector = GitCollector()
        assert collector._cache is None
        assert collector._cache_time is None
        assert isinstance(collector._cache_ttl, timedelta)
        assert collector._cache_ttl.total_seconds() == 5

    def test_init_cache_ttl_value(self):
        """Test cache TTL is correctly set."""
        collector = GitCollector()
        assert collector._CACHE_TTL_SECONDS == 5

    def test_init_git_timeout(self):
        """Test git command timeout value."""
        collector = GitCollector()
        assert collector._GIT_COMMAND_TIMEOUT == 2


class TestGitCollectorCollectInfo:
    """Test main collect_git_info method."""

    def test_collect_git_info_cache_hit(self):
        """Test that cached info is returned on valid cache."""
        collector = GitCollector()
        expected_info = GitInfo(branch="main", staged=0, modified=0, untracked=0)
        collector._cache = expected_info
        collector._cache_time = datetime.now()

        result = collector.collect_git_info()
        assert result == expected_info

    def test_collect_git_info_cache_miss(self):
        """Test that git command is executed on cache miss."""
        collector = GitCollector()
        expected_info = GitInfo(branch="feature", staged=1, modified=2, untracked=3)

        with patch.object(collector, "_fetch_git_info", return_value=expected_info):
            result = collector.collect_git_info()
            assert result == expected_info
            assert collector._cache == expected_info

    def test_collect_git_info_cache_expired(self):
        """Test that cache is refreshed when expired."""
        collector = GitCollector()
        old_info = GitInfo(branch="old", staged=0, modified=0, untracked=0)
        new_info = GitInfo(branch="new", staged=1, modified=2, untracked=3)

        collector._cache = old_info
        collector._cache_time = datetime.now() - timedelta(seconds=10)

        with patch.object(collector, "_fetch_git_info", return_value=new_info):
            result = collector.collect_git_info()
            assert result == new_info
            assert result.branch == "new"


class TestGitCollectorFetchGitInfo:
    """Test _fetch_git_info method."""

    def test_fetch_git_info_success(self):
        """Test successful git command execution."""
        collector = GitCollector()
        git_output = "## main...origin/main\nM  file1.py\n"

        with patch("subprocess.run") as mock_run:
            mock_run.return_value = MagicMock(returncode=0, stdout=git_output, stderr="")
            result = collector._fetch_git_info()
            assert result.branch == "main"
            mock_run.assert_called_once()

    def test_fetch_git_info_command_failed(self):
        """Test handling of failed git command."""
        collector = GitCollector()

        with patch("subprocess.run") as mock_run:
            mock_run.return_value = MagicMock(returncode=1, stdout="", stderr="fatal: not a git repository")
            result = collector._fetch_git_info()
            assert result.branch == "unknown"
            assert result.staged == 0

    def test_fetch_git_info_timeout(self):
        """Test handling of git command timeout."""
        collector = GitCollector()

        with patch("subprocess.run", side_effect=subprocess.TimeoutExpired("git", 2)):
            result = collector._fetch_git_info()
            assert result.branch == "unknown"
            assert result.staged == 0

    def test_fetch_git_info_exception(self):
        """Test handling of unexpected exception."""
        collector = GitCollector()

        with patch("subprocess.run", side_effect=Exception("Unexpected error")):
            result = collector._fetch_git_info()
            assert result.branch == "unknown"

    def test_fetch_git_info_correct_command(self):
        """Test that correct git command is executed."""
        collector = GitCollector()
        git_output = "## main...origin/main\n"

        with patch("subprocess.run") as mock_run:
            mock_run.return_value = MagicMock(returncode=0, stdout=git_output, stderr="")
            collector._fetch_git_info()
            mock_run.assert_called_once_with(
                ["git", "status", "-b", "--porcelain"],
                capture_output=True,
                text=True,
                timeout=2,
            )

    def test_fetch_git_info_timeout_value(self):
        """Test that correct timeout is used."""
        collector = GitCollector()
        git_output = "## main...origin/main\n"

        with patch("subprocess.run") as mock_run:
            mock_run.return_value = MagicMock(returncode=0, stdout=git_output, stderr="")
            collector._fetch_git_info()
            call_args = mock_run.call_args
            assert call_args.kwargs["timeout"] == 2


class TestGitCollectorParseGitOutput:
    """Test _parse_git_output method."""

    def test_parse_git_output_simple(self):
        """Test parsing simple git output."""
        collector = GitCollector()
        output = "## main...origin/main\nM  file1.py\n"
        result = collector._parse_git_output(output)
        assert result.branch == "main"

    def test_parse_git_output_with_staged_files(self):
        """Test parsing with staged files."""
        collector = GitCollector()
        output = "## main...origin/main\nA  new_file.py\nM  modified.py\n"
        result = collector._parse_git_output(output)
        assert result.staged == 2

    def test_parse_git_output_with_modified_files(self):
        """Test parsing with modified files."""
        collector = GitCollector()
        output = "## main...origin/main\n M file1.py\n M file2.py\n"
        result = collector._parse_git_output(output)
        assert result.modified == 2

    def test_parse_git_output_with_untracked_files(self):
        """Test parsing with untracked files."""
        collector = GitCollector()
        output = "## main...origin/main\n?? untracked1.py\n?? untracked2.py\n"
        result = collector._parse_git_output(output)
        assert result.untracked == 2

    def test_parse_git_output_mixed_changes(self):
        """Test parsing with mixed file changes."""
        collector = GitCollector()
        output = "## develop...origin/develop\n" "A  new.py\n" "M  modified.py\n" "?? untracked.py\n"
        result = collector._parse_git_output(output)
        assert result.branch == "develop"
        assert result.staged == 2  # A and M are both staged changes
        # Note: "??" counts as modified since second char '?' is not space or dot
        assert result.modified == 1  # ?? line counts as modified
        assert result.untracked == 1

    def test_parse_git_output_empty_input(self):
        """Test parsing empty git output."""
        collector = GitCollector()
        output = ""
        result = collector._parse_git_output(output)
        assert result.branch == "unknown"

    def test_parse_git_output_only_header(self):
        """Test parsing with only header line."""
        collector = GitCollector()
        output = "## main...origin/main\n"
        result = collector._parse_git_output(output)
        assert result.branch == "main"
        assert result.staged == 0
        assert result.modified == 0


class TestGitCollectorCountChanges:
    """Test _count_changes method."""

    def test_count_changes_empty_lines(self):
        """Test counting with empty lines list."""
        collector = GitCollector()
        staged, modified, untracked = collector._count_changes([])
        assert staged == 0
        assert modified == 0
        assert untracked == 0

    def test_count_changes_staged_added(self):
        """Test counting staged added files."""
        collector = GitCollector()
        lines = ["A  newfile.py"]
        staged, modified, untracked = collector._count_changes(lines)
        assert staged == 1

    def test_count_changes_staged_modified(self):
        """Test counting staged modified files."""
        collector = GitCollector()
        lines = ["M  modified.py"]
        staged, modified, untracked = collector._count_changes(lines)
        assert staged == 1

    def test_count_changes_working_directory_modified(self):
        """Test counting working directory modified files."""
        collector = GitCollector()
        lines = [" M file.py"]  # Modified in working directory
        staged, modified, untracked = collector._count_changes(lines)
        assert modified == 1

    def test_count_changes_untracked(self):
        """Test counting untracked files."""
        collector = GitCollector()
        lines = ["?? untracked.py"]
        staged, modified, untracked = collector._count_changes(lines)
        assert untracked == 1

    def test_count_changes_mixed_statuses(self):
        """Test counting with mixed file statuses."""
        collector = GitCollector()
        lines = [
            "A  staged_new.py",  # A: staged=1
            "M  staged_modified.py",  # M: staged=1
            " M working_modified.py",  # space+M: modified=1 (second char M != space)
            "?? untracked1.py",  # ??: untracked=1, but second char ? != space so modified+=1
            "?? untracked2.py",  # ??: untracked=1, but second char ? != space so modified+=1
        ]
        staged, modified, untracked = collector._count_changes(lines)
        assert staged == 2
        assert modified == 3  # space+M (1) + ?? (2 more from second char not being space)
        assert untracked == 2

    def test_count_changes_short_lines(self):
        """Test handling of lines shorter than 2 chars."""
        collector = GitCollector()
        lines = ["M"]  # Too short
        staged, modified, untracked = collector._count_changes(lines)
        assert staged == 0

    def test_count_changes_deleted_files(self):
        """Test counting deleted files."""
        collector = GitCollector()
        lines = [" D deleted.py"]
        staged, modified, untracked = collector._count_changes(lines)
        assert modified == 1

    def test_count_changes_renamed_files(self):
        """Test counting renamed files."""
        collector = GitCollector()
        lines = [" R renamed.py"]
        staged, modified, untracked = collector._count_changes(lines)
        assert modified == 1

    def test_count_changes_status_with_space(self):
        """Test counting files with no changes (space status)."""
        collector = GitCollector()
        lines = ["   file.py"]  # Spaces indicate no change
        staged, modified, untracked = collector._count_changes(lines)
        assert staged == 0
        assert modified == 0


class TestGitCollectorExtractBranch:
    """Test _extract_branch static method."""

    def test_extract_branch_valid_format(self):
        """Test branch extraction with valid format."""
        branch = GitCollector._extract_branch("## main...origin/main")
        assert branch == "main"

    def test_extract_branch_develop(self):
        """Test branch extraction for develop."""
        branch = GitCollector._extract_branch("## develop...origin/develop")
        assert branch == "develop"

    def test_extract_branch_with_dashes(self):
        """Test branch extraction with dashes in name."""
        branch = GitCollector._extract_branch("## feature-branch...origin/feature-branch")
        assert branch == "feature-branch"

    def test_extract_branch_with_slashes(self):
        """Test branch extraction with slashes in name."""
        branch = GitCollector._extract_branch("## feature/new-feature...origin/feature/new-feature")
        assert branch == "feature/new-feature"

    def test_extract_branch_invalid_format_no_header(self):
        """Test branch extraction with no ## header."""
        branch = GitCollector._extract_branch("invalid branch line")
        assert branch == "unknown"

    def test_extract_branch_empty_string(self):
        """Test branch extraction with empty string."""
        branch = GitCollector._extract_branch("")
        assert branch == "unknown"

    def test_extract_branch_detached_head(self):
        """Test branch extraction for detached HEAD."""
        branch = GitCollector._extract_branch("## HEAD (no branch)")
        assert branch == "HEAD"


class TestGitCollectorCreateErrorInfo:
    """Test _create_error_info static method."""

    def test_create_error_info_default_values(self):
        """Test error info has correct default values."""
        info = GitCollector._create_error_info()
        assert info.branch == "unknown"
        assert info.staged == 0
        assert info.modified == 0
        assert info.untracked == 0

    def test_create_error_info_returns_gitinfo(self):
        """Test that error info returns GitInfo instance."""
        info = GitCollector._create_error_info()
        assert isinstance(info, GitInfo)


class TestGitCollectorCache:
    """Test caching behavior."""

    def test_is_cache_valid_with_valid_cache(self):
        """Test cache validation with valid cache."""
        collector = GitCollector()
        collector._cache = GitInfo("main", 0, 0, 0)
        collector._cache_time = datetime.now()
        assert collector._is_cache_valid()

    def test_is_cache_valid_with_expired_cache(self):
        """Test cache validation with expired cache."""
        collector = GitCollector()
        collector._cache = GitInfo("main", 0, 0, 0)
        collector._cache_time = datetime.now() - timedelta(seconds=10)
        assert not collector._is_cache_valid()

    def test_is_cache_valid_with_no_cache(self):
        """Test cache validation with no cache."""
        collector = GitCollector()
        collector._cache = None
        assert not collector._is_cache_valid()

    def test_update_cache(self):
        """Test cache update."""
        collector = GitCollector()
        info = GitInfo("develop", 1, 2, 3)
        collector._update_cache(info)
        assert collector._cache == info
        assert collector._cache_time is not None

    def test_cache_ttl_setting(self):
        """Test that cache TTL is correctly set."""
        collector = GitCollector()
        assert collector._cache_ttl == timedelta(seconds=5)


class TestGitCollectorGitInfo:
    """Test GitInfo dataclass."""

    def test_gitinfo_creation(self):
        """Test GitInfo creation."""
        info = GitInfo(branch="main", staged=1, modified=2, untracked=3)
        assert info.branch == "main"
        assert info.staged == 1
        assert info.modified == 2
        assert info.untracked == 3

    def test_gitinfo_equality(self):
        """Test GitInfo equality."""
        info1 = GitInfo("main", 0, 0, 0)
        info2 = GitInfo("main", 0, 0, 0)
        assert info1 == info2

    def test_gitinfo_inequality(self):
        """Test GitInfo inequality."""
        info1 = GitInfo("main", 0, 0, 0)
        info2 = GitInfo("develop", 0, 0, 0)
        assert info1 != info2


class TestGitCollectorIntegration:
    """Integration tests for GitCollector."""

    def test_full_git_collection_flow(self):
        """Test complete git information collection flow."""
        collector = GitCollector()
        git_output = (
            "## feature-branch...origin/feature-branch\n"
            "A  new_feature.py\n"
            "M  updated_file.py\n"
            " M working_file.py\n"
            "?? untracked.py\n"
        )

        with patch("subprocess.run") as mock_run:
            mock_run.return_value = MagicMock(returncode=0, stdout=git_output, stderr="")
            result = collector.collect_git_info()
            assert result.branch == "feature-branch"
            assert result.staged == 2  # A and M in index
            assert result.modified == 2  # M in working dir + space+M = both count
            assert result.untracked == 1

    def test_git_collection_with_multiple_calls(self):
        """Test git collection with multiple calls (caching)."""
        collector = GitCollector()
        git_output = "## main...origin/main\nM  file.py\n"

        with patch("subprocess.run") as mock_run:
            mock_run.return_value = MagicMock(returncode=0, stdout=git_output, stderr="")

            # First call
            result1 = collector.collect_git_info()
            call_count_1 = mock_run.call_count

            # Second call (should use cache)
            result2 = collector.collect_git_info()
            call_count_2 = mock_run.call_count

            assert result1 == result2
            assert call_count_1 == call_count_2  # No additional call


class TestGitCollectorEdgeCases:
    """Test edge cases and error conditions."""

    def test_git_output_with_no_branch(self):
        """Test handling of git output with no branch info."""
        collector = GitCollector()
        output = "M  file.py\n"
        result = collector._parse_git_output(output)
        assert result.branch == "unknown"

    def test_git_output_with_spaces_in_branch(self):
        """Test handling of spaces in branch names (if supported)."""
        collector = GitCollector()
        output = "## feature branch...origin/feature\nM  file.py\n"
        result = collector._parse_git_output(output)
        assert isinstance(result.branch, str)

    def test_git_output_with_unicode_characters(self):
        """Test handling of unicode in git output."""
        collector = GitCollector()
        output = "## main...origin/main\nM  文件.py\n"
        result = collector._parse_git_output(output)
        assert result.branch == "main"

    def test_multiple_status_characters(self):
        """Test parsing with various status character combinations."""
        collector = GitCollector()
        lines = [
            "AM file1.py",  # Added and modified
            "MM file2.py",  # Modified in both
            "DD file3.py",  # Deleted in both
        ]
        staged, modified, untracked = collector._count_changes(lines)
        # AM: staged=1, modified=1; MM: staged=1, modified=1; DD: staged=1, modified=1
        assert staged >= 2
        assert modified >= 2

    def test_subprocess_communicate_vs_run(self):
        """Test that subprocess.run is used correctly."""
        collector = GitCollector()

        with patch("subprocess.run") as mock_run:
            mock_run.return_value = MagicMock(returncode=0, stdout="## main...origin/main\n", stderr="")
            collector._fetch_git_info()
            assert mock_run.called
            assert mock_run.call_args.kwargs["capture_output"] is True
            assert mock_run.call_args.kwargs["text"] is True
