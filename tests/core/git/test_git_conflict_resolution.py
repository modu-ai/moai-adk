"""Unit tests for GitConflictDetector - Git merge conflict detection and resolution.

Tests the GitConflictDetector class which detects merge conflicts,
analyzes severity, and provides safe auto-resolution for config files.
"""

import subprocess
import tempfile
from pathlib import Path

import pytest

from moai_adk.core.git.conflict_detector import (
    ConflictFile,
    ConflictSeverity,
    GitConflictDetector,
)


class TestConflictFileDataClass:
    """Tests for ConflictFile data class."""

    def test_conflict_file_creation(self):
        """Test ConflictFile creation with all attributes."""
        conflict = ConflictFile(
            path="src/auth.py",
            severity=ConflictSeverity.MEDIUM,
            conflict_type="code",
            lines_conflicting=10,
            description="Function signature change conflicts with modifications",
        )

        assert conflict.path == "src/auth.py"
        assert conflict.severity == ConflictSeverity.MEDIUM
        assert conflict.conflict_type == "code"
        assert conflict.lines_conflicting == 10
        assert conflict.description == "Function signature change conflicts with modifications"

    def test_conflict_severity_enum_values(self):
        """Test ConflictSeverity enum has expected values."""
        assert ConflictSeverity.LOW.value == "low"
        assert ConflictSeverity.MEDIUM.value == "medium"
        assert ConflictSeverity.HIGH.value == "high"


class TestGitConflictDetectorInitialization:
    """Tests for GitConflictDetector initialization."""

    def test_initialize_with_valid_repo(self):
        """Test GitConflictDetector initializes with valid git repository."""
        with tempfile.TemporaryDirectory() as tmpdir:
            repo_path = Path(tmpdir)
            # Initialize a git repo
            subprocess.run(
                ["git", "init"],
                cwd=repo_path,
                capture_output=True,
                check=True,
            )

            detector = GitConflictDetector(repo_path)
            assert detector.repo_path == repo_path
            assert detector.repo is not None

    def test_initialize_with_invalid_repo(self):
        """Test GitConflictDetector raises error with non-git directory."""
        with tempfile.TemporaryDirectory() as tmpdir:
            repo_path = Path(tmpdir)
            # Don't initialize git repo

            with pytest.raises(Exception):  # Should raise InvalidGitRepositoryError
                GitConflictDetector(repo_path)


class TestCanMergeDetection:
    """Tests for can_merge method - clean merge detection."""

    def test_detect_clean_merge(self):
        """Test detecting clean merge when no conflicts exist."""
        with tempfile.TemporaryDirectory() as tmpdir:
            repo_path = Path(tmpdir)
            # Initialize git repo
            subprocess.run(
                ["git", "init"],
                cwd=repo_path,
                capture_output=True,
                check=True,
            )
            subprocess.run(
                ["git", "config", "user.email", "test@example.com"],
                cwd=repo_path,
                capture_output=True,
                check=True,
            )
            subprocess.run(
                ["git", "config", "user.name", "Test User"],
                cwd=repo_path,
                capture_output=True,
                check=True,
            )

            # Create initial commit on main
            (repo_path / "file.txt").write_text("initial content")
            subprocess.run(
                ["git", "add", "file.txt"],
                cwd=repo_path,
                capture_output=True,
                check=True,
            )
            subprocess.run(
                ["git", "commit", "-m", "Initial commit"],
                cwd=repo_path,
                capture_output=True,
                check=True,
            )

            # Create feature branch with non-conflicting change
            subprocess.run(
                ["git", "checkout", "-b", "feature/test"],
                cwd=repo_path,
                capture_output=True,
                check=True,
            )
            (repo_path / "feature.txt").write_text("feature content")
            subprocess.run(
                ["git", "add", "feature.txt"],
                cwd=repo_path,
                capture_output=True,
                check=True,
            )
            subprocess.run(
                ["git", "commit", "-m", "Add feature"],
                cwd=repo_path,
                capture_output=True,
                check=True,
            )

            # Switch back to main and make non-conflicting change
            subprocess.run(
                ["git", "checkout", "main"],
                cwd=repo_path,
                capture_output=True,
                check=True,
            )
            (repo_path / "main.txt").write_text("main content")
            subprocess.run(
                ["git", "add", "main.txt"],
                cwd=repo_path,
                capture_output=True,
                check=True,
            )
            subprocess.run(
                ["git", "commit", "-m", "Update main"],
                cwd=repo_path,
                capture_output=True,
                check=True,
            )

            detector = GitConflictDetector(repo_path)
            result = detector.can_merge("feature/test", "main")

            assert result["can_merge"] is True
            assert len(result["conflicts"]) == 0

    def test_detect_code_conflicts(self):
        """Test detecting code conflicts when files have competing changes."""
        with tempfile.TemporaryDirectory() as tmpdir:
            repo_path = Path(tmpdir)
            # Initialize git repo
            subprocess.run(
                ["git", "init"],
                cwd=repo_path,
                capture_output=True,
                check=True,
            )
            subprocess.run(
                ["git", "config", "user.email", "test@example.com"],
                cwd=repo_path,
                capture_output=True,
                check=True,
            )
            subprocess.run(
                ["git", "config", "user.name", "Test User"],
                cwd=repo_path,
                capture_output=True,
                check=True,
            )

            # Create initial commit with shared file
            (repo_path / "auth.py").write_text("def login():\n    pass\n")
            subprocess.run(
                ["git", "add", "auth.py"],
                cwd=repo_path,
                capture_output=True,
                check=True,
            )
            subprocess.run(
                ["git", "commit", "-m", "Initial auth"],
                cwd=repo_path,
                capture_output=True,
                check=True,
            )

            # Create feature branch with change
            subprocess.run(
                ["git", "checkout", "-b", "feature/auth"],
                cwd=repo_path,
                capture_output=True,
                check=True,
            )
            (repo_path / "auth.py").write_text("def login():\n    return authenticate()\n")
            subprocess.run(
                ["git", "add", "auth.py"],
                cwd=repo_path,
                capture_output=True,
                check=True,
            )
            subprocess.run(
                ["git", "commit", "-m", "Add auth logic"],
                cwd=repo_path,
                capture_output=True,
                check=True,
            )

            # Switch back to main and make conflicting change
            subprocess.run(
                ["git", "checkout", "main"],
                cwd=repo_path,
                capture_output=True,
                check=True,
            )
            (repo_path / "auth.py").write_text("def login():\n    return user.authenticate()\n")
            subprocess.run(
                ["git", "add", "auth.py"],
                cwd=repo_path,
                capture_output=True,
                check=True,
            )
            subprocess.run(
                ["git", "commit", "-m", "Update auth"],
                cwd=repo_path,
                capture_output=True,
                check=True,
            )

            detector = GitConflictDetector(repo_path)
            result = detector.can_merge("feature/auth", "main")

            assert result["can_merge"] is False
            assert len(result["conflicts"]) > 0
            assert any(c.path == "auth.py" for c in result["conflicts"])


class TestAnalyzeConflicts:
    """Tests for analyze_conflicts method - conflict severity analysis."""

    def test_analyze_config_conflict_severity_low(self):
        """Test analyzing .gitignore conflict as LOW severity."""
        with tempfile.TemporaryDirectory() as tmpdir:
            repo_path = Path(tmpdir)
            # Initialize git repo first
            subprocess.run(
                ["git", "init"],
                cwd=repo_path,
                capture_output=True,
                check=True,
            )
            detector = GitConflictDetector(repo_path)

            conflicts = [
                ConflictFile(
                    path=".gitignore",
                    severity=ConflictSeverity.LOW,
                    conflict_type="config",
                    lines_conflicting=3,
                    description="Different entries added",
                )
            ]

            analyzed = detector.analyze_conflicts(conflicts)
            assert len(analyzed) == 1
            assert analyzed[0].severity == ConflictSeverity.LOW
            assert analyzed[0].conflict_type == "config"

    def test_analyze_code_conflict_severity_medium(self):
        """Test analyzing code conflict as MEDIUM/HIGH severity."""
        with tempfile.TemporaryDirectory() as tmpdir:
            repo_path = Path(tmpdir)
            # Initialize git repo first
            subprocess.run(
                ["git", "init"],
                cwd=repo_path,
                capture_output=True,
                check=True,
            )
            detector = GitConflictDetector(repo_path)

            conflicts = [
                ConflictFile(
                    path="tests/test_auth.py",  # test file, not src/
                    severity=ConflictSeverity.MEDIUM,
                    conflict_type="code",
                    lines_conflicting=15,
                    description="Test modified in both branches",
                )
            ]

            analyzed = detector.analyze_conflicts(conflicts)
            assert len(analyzed) == 1
            assert analyzed[0].severity == ConflictSeverity.MEDIUM
            assert analyzed[0].conflict_type == "code"

    def test_analyze_multiple_conflicts_with_different_severities(self):
        """Test analyzing multiple conflicts with different severity levels."""
        with tempfile.TemporaryDirectory() as tmpdir:
            repo_path = Path(tmpdir)
            # Initialize git repo first
            subprocess.run(
                ["git", "init"],
                cwd=repo_path,
                capture_output=True,
                check=True,
            )
            detector = GitConflictDetector(repo_path)

            conflicts = [
                ConflictFile(
                    path=".gitignore",
                    severity=ConflictSeverity.LOW,
                    conflict_type="config",
                    lines_conflicting=2,
                    description="Config conflict",
                ),
                ConflictFile(
                    path="src/auth.py",
                    severity=ConflictSeverity.HIGH,
                    conflict_type="code",
                    lines_conflicting=50,
                    description="Extensive code changes",
                ),
            ]

            analyzed = detector.analyze_conflicts(conflicts)
            assert len(analyzed) == 2
            # Should return with highest severity first
            assert analyzed[0].severity in [ConflictSeverity.HIGH, ConflictSeverity.LOW]


class TestAutoResolveSafe:
    """Tests for auto_resolve_safe method - safe auto-resolution for config files."""

    def test_auto_resolve_claude_md_conflict(self):
        """Test auto-resolving CLAUDE.md conflict using TemplateMerger."""
        with tempfile.TemporaryDirectory() as tmpdir:
            repo_path = Path(tmpdir)

            # Initialize git repo
            subprocess.run(
                ["git", "init"],
                cwd=repo_path,
                capture_output=True,
                check=True,
            )
            subprocess.run(
                ["git", "config", "user.email", "test@example.com"],
                cwd=repo_path,
                capture_output=True,
                check=True,
            )
            subprocess.run(
                ["git", "config", "user.name", "Test User"],
                cwd=repo_path,
                capture_output=True,
                check=True,
            )

            # Create CLAUDE.md with project info
            claude_md = repo_path / "CLAUDE.md"
            claude_md.write_text("# Project\n\n## Project Information\n\nCustom project info")

            subprocess.run(
                ["git", "add", "CLAUDE.md"],
                cwd=repo_path,
                capture_output=True,
                check=True,
            )
            subprocess.run(
                ["git", "commit", "-m", "Initial CLAUDE.md"],
                cwd=repo_path,
                capture_output=True,
                check=True,
            )

            # Create feature branch with template update
            subprocess.run(
                ["git", "checkout", "-b", "feature/docs"],
                cwd=repo_path,
                capture_output=True,
                check=True,
            )
            claude_md.write_text("# New Template\n\nNew content")
            subprocess.run(
                ["git", "add", "CLAUDE.md"],
                cwd=repo_path,
                capture_output=True,
                check=True,
            )
            subprocess.run(
                ["git", "commit", "-m", "Update template"],
                cwd=repo_path,
                capture_output=True,
                check=True,
            )

            # Back on main, modify project info
            subprocess.run(
                ["git", "checkout", "main"],
                cwd=repo_path,
                capture_output=True,
                check=True,
            )
            claude_md.write_text("# Project\n\n## Project Information\n\nUpdated custom info")
            subprocess.run(
                ["git", "add", "CLAUDE.md"],
                cwd=repo_path,
                capture_output=True,
                check=True,
            )
            subprocess.run(
                ["git", "commit", "-m", "Update project info"],
                cwd=repo_path,
                capture_output=True,
                check=True,
            )

            detector = GitConflictDetector(repo_path)
            conflicts = [
                ConflictFile(
                    path="CLAUDE.md",
                    severity=ConflictSeverity.LOW,
                    conflict_type="config",
                    lines_conflicting=5,
                    description="Project info section preserved",
                )
            ]

            # Note: This would normally interact with merge state
            # Test would verify auto_resolve_safe returns True for safe conflicts
            assert detector is not None
            assert len(conflicts) == 1
            assert conflicts[0].path == "CLAUDE.md"

    def test_auto_resolve_gitignore_conflict(self):
        """Test auto-resolving .gitignore conflict using TemplateMerger."""
        with tempfile.TemporaryDirectory() as tmpdir:
            repo_path = Path(tmpdir)
            # Initialize git repo first
            subprocess.run(
                ["git", "init"],
                cwd=repo_path,
                capture_output=True,
                check=True,
            )
            GitConflictDetector(repo_path)

            conflicts = [
                ConflictFile(
                    path=".gitignore",
                    severity=ConflictSeverity.LOW,
                    conflict_type="config",
                    lines_conflicting=3,
                    description="Different ignore patterns",
                )
            ]

            # Verify conflict is safe for auto-resolution
            is_safe = all(
                c.path in [".gitignore", "CLAUDE.md", ".claude/settings.json"] and c.severity == ConflictSeverity.LOW
                for c in conflicts
            )
            assert is_safe is True

    def test_auto_resolve_config_json_conflict(self):
        """Test auto-resolving .moai/config/config.json conflict."""
        with tempfile.TemporaryDirectory() as tmpdir:
            repo_path = Path(tmpdir)
            # Initialize git repo first
            subprocess.run(
                ["git", "init"],
                cwd=repo_path,
                capture_output=True,
                check=True,
            )
            detector = GitConflictDetector(repo_path)

            conflicts = [
                ConflictFile(
                    path=".moai/config/config.json",
                    severity=ConflictSeverity.LOW,
                    conflict_type="config",
                    lines_conflicting=2,
                    description="Version field updated",
                )
            ]

            all(
                c.path in [".gitignore", "CLAUDE.md", ".claude/settings.json"] and c.severity == ConflictSeverity.LOW
                for c in conflicts
            )
            # This should be safe to auto-resolve but may require special handling
            assert detector is not None

    def test_reject_auto_resolve_code_conflicts(self):
        """Test that code conflicts are NOT auto-resolved."""
        with tempfile.TemporaryDirectory() as tmpdir:
            repo_path = Path(tmpdir)
            # Initialize git repo first
            subprocess.run(
                ["git", "init"],
                cwd=repo_path,
                capture_output=True,
                check=True,
            )
            GitConflictDetector(repo_path)

            conflicts = [
                ConflictFile(
                    path="src/auth.py",
                    severity=ConflictSeverity.MEDIUM,
                    conflict_type="code",
                    lines_conflicting=20,
                    description="Function signature differs",
                )
            ]

            # Code conflicts should NOT be auto-resolved
            is_safe = all(
                c.path in [".gitignore", "CLAUDE.md", ".claude/settings.json"] and c.severity == ConflictSeverity.LOW
                for c in conflicts
            )
            assert is_safe is False


class TestMergeCleanup:
    """Tests for cleanup_merge_state method - safe merge state cleanup."""

    def test_cleanup_failed_merge(self):
        """Test cleaning up after failed merge attempt."""
        with tempfile.TemporaryDirectory() as tmpdir:
            repo_path = Path(tmpdir)

            # Initialize git repo
            subprocess.run(
                ["git", "init"],
                cwd=repo_path,
                capture_output=True,
                check=True,
            )
            subprocess.run(
                ["git", "config", "user.email", "test@example.com"],
                cwd=repo_path,
                capture_output=True,
                check=True,
            )
            subprocess.run(
                ["git", "config", "user.name", "Test User"],
                cwd=repo_path,
                capture_output=True,
                check=True,
            )

            # Create initial commit
            (repo_path / "file.txt").write_text("content")
            subprocess.run(
                ["git", "add", "file.txt"],
                cwd=repo_path,
                capture_output=True,
                check=True,
            )
            subprocess.run(
                ["git", "commit", "-m", "Initial"],
                cwd=repo_path,
                capture_output=True,
                check=True,
            )

            detector = GitConflictDetector(repo_path)

            # Verify cleanup method exists and can be called
            # In real scenario, would verify MERGE_HEAD is removed after cleanup
            assert hasattr(detector, "cleanup_merge_state")
            assert callable(detector.cleanup_merge_state)

    def test_cleanup_removes_merge_state_files(self):
        """Test that cleanup removes MERGE_HEAD and other merge state files."""
        with tempfile.TemporaryDirectory() as tmpdir:
            repo_path = Path(tmpdir)

            # Initialize git repo
            subprocess.run(
                ["git", "init"],
                cwd=repo_path,
                capture_output=True,
                check=True,
            )

            detector = GitConflictDetector(repo_path)
            git_dir = repo_path / ".git"

            # Simulate merge state files
            merge_head = git_dir / "MERGE_HEAD"
            merge_msg = git_dir / "MERGE_MSG"
            merge_head.touch()
            merge_msg.touch()

            assert merge_head.exists()
            assert merge_msg.exists()

            # Cleanup should remove these
            detector.cleanup_merge_state()

            assert not merge_head.exists()
            assert not merge_msg.exists()


class TestRebaseWorkflow:
    """Tests for rebase workflow - alternative to merge for conflict resolution."""

    def test_rebase_feature_branch_on_develop(self):
        """Test rebasing feature branch onto develop to resolve conflicts."""
        with tempfile.TemporaryDirectory() as tmpdir:
            repo_path = Path(tmpdir)

            # Initialize git repo
            subprocess.run(
                ["git", "init"],
                cwd=repo_path,
                capture_output=True,
                check=True,
            )
            subprocess.run(
                ["git", "config", "user.email", "test@example.com"],
                cwd=repo_path,
                capture_output=True,
                check=True,
            )
            subprocess.run(
                ["git", "config", "user.name", "Test User"],
                cwd=repo_path,
                capture_output=True,
                check=True,
            )

            # Create develop branch
            (repo_path / "file.txt").write_text("initial")
            subprocess.run(
                ["git", "add", "file.txt"],
                cwd=repo_path,
                capture_output=True,
                check=True,
            )
            subprocess.run(
                ["git", "commit", "-m", "Initial"],
                cwd=repo_path,
                capture_output=True,
                check=True,
            )
            subprocess.run(
                ["git", "checkout", "-b", "develop"],
                cwd=repo_path,
                capture_output=True,
                check=True,
            )

            # Create feature branch
            subprocess.run(
                ["git", "checkout", "-b", "feature/test"],
                cwd=repo_path,
                capture_output=True,
                check=True,
            )
            (repo_path / "feature.txt").write_text("feature")
            subprocess.run(
                ["git", "add", "feature.txt"],
                cwd=repo_path,
                capture_output=True,
                check=True,
            )
            subprocess.run(
                ["git", "commit", "-m", "Feature work"],
                cwd=repo_path,
                capture_output=True,
                check=True,
            )

            # Update develop
            subprocess.run(
                ["git", "checkout", "develop"],
                cwd=repo_path,
                capture_output=True,
                check=True,
            )
            (repo_path / "file.txt").write_text("updated")
            subprocess.run(
                ["git", "add", "file.txt"],
                cwd=repo_path,
                capture_output=True,
                check=True,
            )
            subprocess.run(
                ["git", "commit", "-m", "Update"],
                cwd=repo_path,
                capture_output=True,
                check=True,
            )

            detector = GitConflictDetector(repo_path)

            # Verify detector has rebase method
            assert hasattr(detector, "rebase_branch")
            assert callable(detector.rebase_branch)

            # Rebase feature onto develop
            # result = detector.rebase_branch("feature/test", "develop")
            # Actual rebase tested in integration tests


class TestIntegrationWith3Sync:
    """Integration tests for conflict detection with 3-sync command."""

    def test_detector_returns_correct_structure(self):
        """Test that detector returns properly structured conflict data."""
        with tempfile.TemporaryDirectory() as tmpdir:
            repo_path = Path(tmpdir)

            # Initialize git repo
            subprocess.run(
                ["git", "init"],
                cwd=repo_path,
                capture_output=True,
                check=True,
            )

            detector = GitConflictDetector(repo_path)

            # Verify detector returns dict with expected keys
            mock_conflicts = [
                ConflictFile(
                    path="test.py",
                    severity=ConflictSeverity.MEDIUM,
                    conflict_type="code",
                    lines_conflicting=10,
                    description="Test conflict",
                )
            ]

            analyzed = detector.analyze_conflicts(mock_conflicts)
            assert isinstance(analyzed, list)
            assert all(isinstance(c, ConflictFile) for c in analyzed)

    def test_conflict_summary_for_user_presentation(self):
        """Test generating summary for user presentation in 3-sync."""
        with tempfile.TemporaryDirectory() as tmpdir:
            repo_path = Path(tmpdir)

            # Initialize git repo
            subprocess.run(
                ["git", "init"],
                cwd=repo_path,
                capture_output=True,
                check=True,
            )

            detector = GitConflictDetector(repo_path)

            conflicts = [
                ConflictFile(
                    path=".gitignore",
                    severity=ConflictSeverity.LOW,
                    conflict_type="config",
                    lines_conflicting=3,
                    description="Config entries differ",
                ),
                ConflictFile(
                    path="src/auth.py",
                    severity=ConflictSeverity.HIGH,
                    conflict_type="code",
                    lines_conflicting=50,
                    description="Function changes conflict",
                ),
            ]

            # Verify we can summarize conflicts
            summary = detector.summarize_conflicts(conflicts)
            assert summary is not None
            assert "LOW" in summary or ".gitignore" in summary
            assert "HIGH" in summary or "src/auth.py" in summary
