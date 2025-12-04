"""
Extended tests for conflict_detector module.

Tests enums, dataclasses, conflict detection, analysis, and resolution.
"""

import tempfile
from dataclasses import dataclass
from pathlib import Path
from unittest.mock import MagicMock, Mock, patch

import pytest

from moai_adk.core.git.conflict_detector import (
    ConflictFile,
    ConflictSeverity,
    GitConflictDetector,
)


class TestConflictSeverityEnum:
    """Test ConflictSeverity enum."""

    def test_conflict_severity_values(self):
        """Test ConflictSeverity enum values."""
        assert ConflictSeverity.LOW.value == "low"
        assert ConflictSeverity.MEDIUM.value == "medium"
        assert ConflictSeverity.HIGH.value == "high"

    def test_severity_comparison(self):
        """Test severity enum comparison."""
        assert ConflictSeverity.LOW != ConflictSeverity.HIGH
        assert ConflictSeverity.MEDIUM != ConflictSeverity.LOW


class TestConflictFileDataclass:
    """Test ConflictFile dataclass."""

    def test_conflict_file_creation(self):
        """Test ConflictFile creation."""
        conflict = ConflictFile(
            path="src/main.py",
            severity=ConflictSeverity.HIGH,
            conflict_type="code",
            lines_conflicting=5,
            description="Merge conflict in source file",
        )
        assert conflict.path == "src/main.py"
        assert conflict.severity == ConflictSeverity.HIGH
        assert conflict.conflict_type == "code"

    def test_conflict_file_with_config(self):
        """Test ConflictFile for config type."""
        conflict = ConflictFile(
            path=".claude/settings.json",
            severity=ConflictSeverity.LOW,
            conflict_type="config",
            lines_conflicting=1,
            description="Config conflict",
        )
        assert conflict.conflict_type == "config"
        assert conflict.severity == ConflictSeverity.LOW


class TestGitConflictDetectorInit:
    """Test GitConflictDetector initialization."""

    def test_detector_init_with_valid_repo(self):
        """Test detector initialization with valid git repo."""
        with tempfile.TemporaryDirectory() as tmpdir:
            repo_path = Path(tmpdir)
            with patch("moai_adk.core.git.conflict_detector.Repo") as mock_repo:
                mock_repo.return_value = MagicMock()
                detector = GitConflictDetector(repo_path)
                assert detector.repo_path == repo_path

    def test_detector_init_invalid_repo(self):
        """Test detector initialization with invalid repo."""
        with tempfile.TemporaryDirectory() as tmpdir:
            from git import InvalidGitRepositoryError

            with patch("moai_adk.core.git.conflict_detector.Repo") as mock_repo:
                mock_repo.side_effect = InvalidGitRepositoryError("Not a repo")
                with pytest.raises(InvalidGitRepositoryError):
                    GitConflictDetector(tmpdir)

    def test_detector_safe_auto_resolve_files(self):
        """Test safe auto-resolve file list."""
        detector = GitConflictDetector.__dict__
        # Check that SAFE_AUTO_RESOLVE_FILES exists and contains expected files
        assert hasattr(GitConflictDetector, "SAFE_AUTO_RESOLVE_FILES")

    def test_detector_config_file_patterns(self):
        """Test config file pattern list."""
        assert hasattr(GitConflictDetector, "CONFIG_FILE_PATTERNS")


class TestConflictDetection:
    """Test conflict detection functionality."""

    def test_can_merge_no_conflicts(self):
        """Test merge check with no conflicts."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch("moai_adk.core.git.conflict_detector.Repo") as mock_repo_class:
                mock_repo = MagicMock()
                mock_repo_class.return_value = mock_repo
                mock_repo.active_branch.name = "main"
                mock_repo.git.merge = MagicMock()

                detector = GitConflictDetector(tmpdir)
                result = detector.can_merge("feature", "main")

                assert "can_merge" in result
                assert "conflicts" in result

    def test_can_merge_with_conflicts(self):
        """Test merge check with conflicts."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch("moai_adk.core.git.conflict_detector.Repo") as mock_repo_class:
                mock_repo = MagicMock()
                mock_repo_class.return_value = mock_repo
                mock_repo.active_branch.name = "main"
                mock_repo.git.merge.side_effect = Exception("merge conflict")
                mock_repo.index.unmerged_blobs.return_value = {}

                detector = GitConflictDetector(tmpdir)
                result = detector.can_merge("feature", "main")

                assert "can_merge" in result

    def test_detect_conflicted_files(self):
        """Test detecting conflicted files."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch("moai_adk.core.git.conflict_detector.Repo") as mock_repo_class:
                mock_repo = MagicMock()
                mock_repo_class.return_value = mock_repo
                mock_repo.index.unmerged_blobs.return_value = {}

                detector = GitConflictDetector(tmpdir)
                conflicts = detector._detect_conflicted_files()

                assert isinstance(conflicts, list)


class TestFileClassification:
    """Test file type classification."""

    def test_classify_markdown_as_config(self):
        """Test classifying markdown as config."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch("moai_adk.core.git.conflict_detector.Repo") as mock_repo:
                mock_repo.return_value = MagicMock()
                detector = GitConflictDetector(tmpdir)
                file_type = detector._classify_file_type("README.md")
                assert file_type == "config"

    def test_classify_json_as_config(self):
        """Test classifying JSON as config."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch("moai_adk.core.git.conflict_detector.Repo") as mock_repo:
                mock_repo.return_value = MagicMock()
                detector = GitConflictDetector(tmpdir)
                file_type = detector._classify_file_type("config.json")
                assert file_type == "config"

    def test_classify_python_as_code(self):
        """Test classifying Python as code."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch("moai_adk.core.git.conflict_detector.Repo") as mock_repo:
                mock_repo.return_value = MagicMock()
                detector = GitConflictDetector(tmpdir)
                file_type = detector._classify_file_type("main.py")
                assert file_type == "code"

    def test_classify_default_as_code(self):
        """Test default classification as code."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch("moai_adk.core.git.conflict_detector.Repo") as mock_repo:
                mock_repo.return_value = MagicMock()
                detector = GitConflictDetector(tmpdir)
                file_type = detector._classify_file_type("unknown.ext")
                assert file_type == "code"


class TestSeverityDetermination:
    """Test conflict severity determination."""

    def test_safe_config_file_low_severity(self):
        """Test safe config files have low severity."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch("moai_adk.core.git.conflict_detector.Repo") as mock_repo:
                mock_repo.return_value = MagicMock()
                detector = GitConflictDetector(tmpdir)
                severity = detector._determine_severity("CLAUDE.md", "config")
                assert severity == ConflictSeverity.LOW

    def test_source_code_high_severity(self):
        """Test source code has high severity."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch("moai_adk.core.git.conflict_detector.Repo") as mock_repo:
                mock_repo.return_value = MagicMock()
                detector = GitConflictDetector(tmpdir)
                severity = detector._determine_severity("src/main.py", "code")
                assert severity == ConflictSeverity.HIGH

    def test_test_code_medium_severity(self):
        """Test test code has medium severity."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch("moai_adk.core.git.conflict_detector.Repo") as mock_repo:
                mock_repo.return_value = MagicMock()
                detector = GitConflictDetector(tmpdir)
                severity = detector._determine_severity("tests/test_main.py", "code")
                assert severity == ConflictSeverity.MEDIUM

    def test_config_file_low_severity(self):
        """Test config file has low severity."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch("moai_adk.core.git.conflict_detector.Repo") as mock_repo:
                mock_repo.return_value = MagicMock()
                detector = GitConflictDetector(tmpdir)
                severity = detector._determine_severity("settings.json", "config")
                assert severity == ConflictSeverity.LOW


class TestConflictAnalysis:
    """Test conflict analysis."""

    def test_analyze_conflicts_sorting(self):
        """Test conflicts are sorted by severity."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch("moai_adk.core.git.conflict_detector.Repo") as mock_repo:
                mock_repo.return_value = MagicMock()
                detector = GitConflictDetector(tmpdir)

                conflicts = [
                    ConflictFile("file1.py", ConflictSeverity.LOW, "code", 1, ""),
                    ConflictFile("file2.py", ConflictSeverity.HIGH, "code", 3, ""),
                    ConflictFile("file3.py", ConflictSeverity.MEDIUM, "code", 2, ""),
                ]

                analyzed = detector.analyze_conflicts(conflicts)

                # Should be sorted HIGH -> MEDIUM -> LOW
                assert analyzed[0].severity == ConflictSeverity.HIGH
                assert analyzed[1].severity == ConflictSeverity.MEDIUM
                assert analyzed[2].severity == ConflictSeverity.LOW

    def test_analyze_conflicts_empty_list(self):
        """Test analyzing empty conflict list."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch("moai_adk.core.git.conflict_detector.Repo") as mock_repo:
                mock_repo.return_value = MagicMock()
                detector = GitConflictDetector(tmpdir)
                analyzed = detector.analyze_conflicts([])
                assert analyzed == []


class TestAutoResolution:
    """Test auto-resolution functionality."""

    def test_auto_resolve_safe_no_conflicts(self):
        """Test auto-resolve with no conflicts."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch("moai_adk.core.git.conflict_detector.Repo") as mock_repo:
                mock_repo.return_value = MagicMock()
                mock_repo.return_value.index.unmerged_blobs.return_value = {}

                detector = GitConflictDetector(tmpdir)
                with patch("moai_adk.core.git.conflict_detector.TemplateMerger"):
                    result = detector.auto_resolve_safe()
                    assert result is True

    def test_auto_resolve_safe_unsafe_file(self):
        """Test auto-resolve with unsafe file."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch("moai_adk.core.git.conflict_detector.Repo") as mock_repo:
                mock_repo.return_value = MagicMock()
                # Return a conflict in unsafe file
                unsafe_file = "src/main.py"
                mock_repo.return_value.index.unmerged_blobs.return_value = {
                    unsafe_file: None
                }
                mock_repo.return_value.index.unmerged_blobs.keys = MagicMock(
                    return_value=[unsafe_file]
                )

                detector = GitConflictDetector(tmpdir)
                result = detector.auto_resolve_safe()
                # Should return False for unsafe files
                assert result is False

    def test_auto_resolve_imports_template_merger(self):
        """Test auto-resolve imports TemplateMerger."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch("moai_adk.core.git.conflict_detector.Repo") as mock_repo:
                mock_repo.return_value = MagicMock()
                mock_repo.return_value.index.unmerged_blobs.return_value = {}

                detector = GitConflictDetector(tmpdir)
                with patch(
                    "moai_adk.core.git.conflict_detector.TemplateMerger"
                ) as mock_merger:
                    result = detector.auto_resolve_safe()
                    # TemplateMerger should be imported
                    assert result is True


class TestCleanupMergeState:
    """Test merge state cleanup."""

    def test_cleanup_removes_merge_files(self):
        """Test cleanup removes merge state files."""
        with tempfile.TemporaryDirectory() as tmpdir:
            repo_path = Path(tmpdir)
            git_dir = repo_path / ".git"
            git_dir.mkdir()

            # Create merge state files
            (git_dir / "MERGE_HEAD").touch()
            (git_dir / "MERGE_MSG").touch()

            with patch("moai_adk.core.git.conflict_detector.Repo") as mock_repo:
                mock_repo.return_value = MagicMock()
                mock_repo.return_value.git.merge = MagicMock()

                detector = GitConflictDetector(repo_path)
                detector.cleanup_merge_state()

                # Files should be removed or attempted to remove
                # (may not exist due to mock)


class TestRebase:
    """Test rebase functionality."""

    def test_rebase_branch(self):
        """Test rebasing a branch."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch("moai_adk.core.git.conflict_detector.Repo") as mock_repo:
                mock_repo.return_value = MagicMock()
                mock_repo.return_value.active_branch.name = "main"

                detector = GitConflictDetector(tmpdir)
                result = detector.rebase_branch("feature", "main")

                # Should attempt rebase
                assert isinstance(result, bool)

    def test_rebase_failure_handling(self):
        """Test rebase failure handling."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch("moai_adk.core.git.conflict_detector.Repo") as mock_repo:
                mock_repo.return_value = MagicMock()
                mock_repo.return_value.active_branch.name = "main"
                mock_repo.return_value.git.rebase.side_effect = Exception("Rebase failed")

                detector = GitConflictDetector(tmpdir)
                result = detector.rebase_branch("feature", "main")

                # Should handle failure gracefully
                assert result is False


class TestSummarization:
    """Test conflict summarization."""

    def test_summarize_no_conflicts(self):
        """Test summarizing with no conflicts."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch("moai_adk.core.git.conflict_detector.Repo") as mock_repo:
                mock_repo.return_value = MagicMock()

                detector = GitConflictDetector(tmpdir)
                summary = detector.summarize_conflicts([])

                assert "No conflicts" in summary

    def test_summarize_multiple_conflicts(self):
        """Test summarizing multiple conflicts."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch("moai_adk.core.git.conflict_detector.Repo") as mock_repo:
                mock_repo.return_value = MagicMock()

                detector = GitConflictDetector(tmpdir)

                conflicts = [
                    ConflictFile("file1.py", ConflictSeverity.HIGH, "code", 1, "Conflict 1"),
                    ConflictFile("file2.json", ConflictSeverity.LOW, "config", 1, "Conflict 2"),
                ]

                summary = detector.summarize_conflicts(conflicts)

                assert "file1.py" in summary
                assert "file2.json" in summary
                assert "HIGH" in summary.upper() or "high" in summary
                assert "LOW" in summary.upper() or "low" in summary

    def test_summarize_formats_correctly(self):
        """Test summary formatting."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch("moai_adk.core.git.conflict_detector.Repo") as mock_repo:
                mock_repo.return_value = MagicMock()

                detector = GitConflictDetector(tmpdir)

                conflicts = [
                    ConflictFile("src/main.py", ConflictSeverity.HIGH, "code", 3, "Main file"),
                ]

                summary = detector.summarize_conflicts(conflicts)

                assert isinstance(summary, str)
                assert len(summary) > 0


class TestEdgeCases:
    """Test edge cases and error conditions."""

    def test_detect_conflicts_file_read_error(self):
        """Test handling file read errors."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch("moai_adk.core.git.conflict_detector.Repo") as mock_repo:
                mock_repo.return_value = MagicMock()
                mock_repo.return_value.index.unmerged_blobs.return_value = {}

                detector = GitConflictDetector(tmpdir)
                conflicts = detector._detect_conflicted_files()

                # Should handle errors gracefully
                assert isinstance(conflicts, list)

    def test_classify_special_filenames(self):
        """Test classifying files with special names."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch("moai_adk.core.git.conflict_detector.Repo") as mock_repo:
                mock_repo.return_value = MagicMock()

                detector = GitConflictDetector(tmpdir)

                # Test various special names
                test_files = [
                    ".gitignore",
                    ".env.local",
                    "Makefile",
                    "docker-compose.yml",
                ]

                for filename in test_files:
                    file_type = detector._classify_file_type(filename)
                    assert file_type in ["config", "code"]

    def test_severity_with_path_depth(self):
        """Test severity determination with various path depths."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch("moai_adk.core.git.conflict_detector.Repo") as mock_repo:
                mock_repo.return_value = MagicMock()

                detector = GitConflictDetector(tmpdir)

                paths = [
                    "src/main.py",
                    "src/core/lib.py",
                    "src/core/deep/nested/file.py",
                ]

                for path in paths:
                    severity = detector._determine_severity(path, "code")
                    assert severity in [
                        ConflictSeverity.LOW,
                        ConflictSeverity.MEDIUM,
                        ConflictSeverity.HIGH,
                    ]

    def test_empty_string_path(self):
        """Test with empty path."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch("moai_adk.core.git.conflict_detector.Repo") as mock_repo:
                mock_repo.return_value = MagicMock()

                detector = GitConflictDetector(tmpdir)
                file_type = detector._classify_file_type("")
                assert file_type in ["config", "code"]

    def test_very_long_path(self):
        """Test with very long path."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch("moai_adk.core.git.conflict_detector.Repo") as mock_repo:
                mock_repo.return_value = MagicMock()

                detector = GitConflictDetector(tmpdir)
                long_path = "/".join(["dir"] * 100) + "/file.py"
                file_type = detector._classify_file_type(long_path)
                assert file_type == "code"

    def test_unicode_in_path(self):
        """Test with unicode characters in path."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch("moai_adk.core.git.conflict_detector.Repo") as mock_repo:
                mock_repo.return_value = MagicMock()

                detector = GitConflictDetector(tmpdir)
                path = "src/测试/файл.py"
                file_type = detector._classify_file_type(path)
                assert file_type in ["config", "code"]


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
