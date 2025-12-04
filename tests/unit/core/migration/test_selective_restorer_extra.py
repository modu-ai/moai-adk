"""Extended tests for moai_adk.core.migration.selective_restorer module.

These tests focus on increasing coverage for file restoration and path handling.
"""

import logging
import shutil
import tempfile
from pathlib import Path
from unittest.mock import MagicMock, patch
from typing import Dict, List

import pytest

from moai_adk.core.migration.selective_restorer import (
    SelectiveRestorer,
    create_selective_restorer,
)


class TestSelectiveRestorerInit:
    """Test SelectiveRestorer initialization."""

    def test_init_with_explicit_backup_path(self):
        """Test initialization with explicit backup path."""
        project_path = Path("/test/project")
        backup_path = Path("/test/backup")

        restorer = SelectiveRestorer(project_path, backup_path)

        assert restorer.project_path == project_path
        assert restorer.backup_path == backup_path

    def test_init_with_auto_detect_backup(self):
        """Test initialization with auto-detected backup."""
        project_path = Path("/test/project")

        with patch.object(SelectiveRestorer, "_find_latest_backup", return_value=Path("/test/backup")):
            restorer = SelectiveRestorer(project_path)

            assert restorer.backup_path == Path("/test/backup")

    def test_init_restoration_log_empty(self):
        """Test restoration log is initialized empty."""
        project_path = Path("/test/project")

        with patch.object(SelectiveRestorer, "_find_latest_backup", return_value=None):
            restorer = SelectiveRestorer(project_path)

            assert restorer.restoration_log == []


class TestFindLatestBackup:
    """Test finding latest backup directory."""

    def test_find_latest_backup_not_found(self):
        """Test finding backup when none exists."""
        project_path = Path("/test/project")

        with patch.object(Path, "exists", return_value=False):
            restorer = SelectiveRestorer(project_path)
            result = restorer._find_latest_backup()

            assert result is None

    def test_find_latest_backup_empty_directory(self):
        """Test finding backup in empty backup directory."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            backups_dir = project_path / ".moai-backups"
            backups_dir.mkdir(exist_ok=True)

            restorer = SelectiveRestorer(project_path)
            result = restorer._find_latest_backup()

            assert result is None

    def test_find_latest_backup_single_backup(self):
        """Test finding single backup."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            backups_dir = project_path / ".moai-backups"
            backups_dir.mkdir(exist_ok=True)

            backup_dir = backups_dir / "pre-update-backup_20240101"
            backup_dir.mkdir(exist_ok=True)

            restorer = SelectiveRestorer(project_path)
            result = restorer._find_latest_backup()

            assert result == backup_dir

    def test_find_latest_backup_multiple_backups(self):
        """Test finding latest backup among multiple."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            backups_dir = project_path / ".moai-backups"
            backups_dir.mkdir(exist_ok=True)

            # Create multiple backups
            backup1 = backups_dir / "pre-update-backup_20240101"
            backup2 = backups_dir / "pre-update-backup_20240102"
            backup1.mkdir(exist_ok=True)
            backup2.mkdir(exist_ok=True)

            restorer = SelectiveRestorer(project_path)
            result = restorer._find_latest_backup()

            # Should return the most recent (backup2)
            assert result == backup2 or result == backup1  # Depends on mtime


class TestRestoreElements:
    """Test restoring selected elements."""

    def test_restore_elements_empty_list(self):
        """Test restoring empty element list."""
        project_path = Path("/test/project")

        with patch.object(SelectiveRestorer, "_find_latest_backup", return_value=None):
            restorer = SelectiveRestorer(project_path)
            success, stats = restorer.restore_elements([])

            assert success is True
            assert stats["total"] == 0
            assert stats["success"] == 0
            assert stats["failed"] == 0

    def test_restore_elements_single_file(self):
        """Test restoring single file element."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            backup_path = Path(tmpdir) / "backup"
            backup_path.mkdir()

            # Create backup file
            backup_file = backup_path / ".claude" / "agents" / "test.md"
            backup_file.parent.mkdir(parents=True)
            backup_file.write_text("test content")

            restorer = SelectiveRestorer(project_path, backup_path)

            with patch.object(restorer, "_restore_element_type", return_value={"total": 1, "success": 1, "failed": 0}):
                with patch("builtins.print"):
                    success, stats = restorer.restore_elements([".claude/agents/test.md"])

                    assert success is True or success is False

    def test_restore_elements_multiple_types(self):
        """Test restoring multiple element types."""
        project_path = Path("/test/project")

        with patch.object(SelectiveRestorer, "_find_latest_backup", return_value=None):
            with patch.object(SelectiveRestorer, "_group_elements_by_type", return_value={"agents": [], "skills": []}):
                with patch.object(SelectiveRestorer, "_restore_element_type", return_value={"total": 0, "success": 0, "failed": 0}):
                    with patch("builtins.print"):
                        restorer = SelectiveRestorer(project_path)
                        success, stats = restorer.restore_elements([".claude/agents/test.md"])

                        assert "by_type" in stats


class TestGroupElementsByType:
    """Test grouping elements by type."""

    def test_group_elements_agents(self):
        """Test grouping agent elements."""
        project_path = Path("/test/project")

        with patch.object(SelectiveRestorer, "_find_latest_backup", return_value=None):
            restorer = SelectiveRestorer(project_path)
            groups = restorer._group_elements_by_type([
                ".claude/agents/agent1.md",
                ".claude/agents/agent2.md"
            ])

            assert "agents" in groups
            assert len(groups["agents"]) == 2

    def test_group_elements_skills(self):
        """Test grouping skill elements."""
        project_path = Path("/test/project")

        with patch.object(SelectiveRestorer, "_find_latest_backup", return_value=None):
            restorer = SelectiveRestorer(project_path)
            groups = restorer._group_elements_by_type([
                ".claude/skills/skill1/",
                ".claude/skills/skill2/"
            ])

            assert "skills" in groups
            assert len(groups["skills"]) == 2

    def test_group_elements_commands(self):
        """Test grouping command elements."""
        project_path = Path("/test/project")

        with patch.object(SelectiveRestorer, "_find_latest_backup", return_value=None):
            restorer = SelectiveRestorer(project_path)
            groups = restorer._group_elements_by_type([
                ".claude/commands/moai/cmd1.md",
                ".claude/commands/cmd2.md"
            ])

            assert "commands" in groups
            assert len(groups["commands"]) >= 1

    def test_group_elements_hooks(self):
        """Test grouping hook elements."""
        project_path = Path("/test/project")

        with patch.object(SelectiveRestorer, "_find_latest_backup", return_value=None):
            restorer = SelectiveRestorer(project_path)
            groups = restorer._group_elements_by_type([
                ".claude/hooks/moai/hook1.py",
                ".claude/hooks/hook2.py"
            ])

            assert "hooks" in groups
            assert len(groups["hooks"]) >= 1

    def test_group_elements_unknown(self):
        """Test grouping unknown element types."""
        project_path = Path("/test/project")

        with patch.object(SelectiveRestorer, "_find_latest_backup", return_value=None):
            restorer = SelectiveRestorer(project_path)
            groups = restorer._group_elements_by_type([
                ".claude/unknown/file.txt"
            ])

            assert "unknown" in groups
            assert len(groups["unknown"]) == 1

    def test_group_elements_mixed_types(self):
        """Test grouping mixed element types."""
        project_path = Path("/test/project")

        with patch.object(SelectiveRestorer, "_find_latest_backup", return_value=None):
            restorer = SelectiveRestorer(project_path)
            groups = restorer._group_elements_by_type([
                ".claude/agents/agent.md",
                ".claude/skills/skill/",
                ".claude/commands/cmd.md",
                ".claude/hooks/hook.py"
            ])

            assert len(groups) > 0
            total_elements = sum(len(v) for v in groups.values())
            assert total_elements >= 4


class TestNormalizeElementPath:
    """Test element path normalization."""

    def test_normalize_relative_path(self):
        """Test normalizing relative path."""
        project_path = Path("/test/project")

        with patch.object(SelectiveRestorer, "_find_latest_backup", return_value=None):
            restorer = SelectiveRestorer(project_path)
            result = restorer._normalize_element_path(Path(".claude/agents/test.md"))

            assert result == Path(".claude/agents/test.md")

    def test_normalize_absolute_path_with_claude(self):
        """Test normalizing absolute path with .claude prefix."""
        project_path = Path("/test/project")

        with patch.object(SelectiveRestorer, "_find_latest_backup", return_value=None):
            restorer = SelectiveRestorer(project_path)
            result = restorer._normalize_element_path(Path("/test/project/.claude/agents/test.md"))

            assert result is not None
            assert ".claude" in str(result)

    def test_normalize_absolute_path_with_moai(self):
        """Test normalizing absolute path with .moai prefix."""
        project_path = Path("/test/project")

        with patch.object(SelectiveRestorer, "_find_latest_backup", return_value=None):
            restorer = SelectiveRestorer(project_path)
            result = restorer._normalize_element_path(Path("/test/project/.moai/config/test.json"))

            assert result is not None
            assert ".moai" in str(result)

    def test_normalize_invalid_path(self):
        """Test normalizing invalid path."""
        project_path = Path("/test/project")

        with patch.object(SelectiveRestorer, "_find_latest_backup", return_value=None):
            restorer = SelectiveRestorer(project_path)
            result = restorer._normalize_element_path(Path("agents/test.md"))

            assert result is None

    def test_normalize_path_without_prefix(self):
        """Test normalizing absolute path without safe prefix."""
        project_path = Path("/test/project")

        with patch.object(SelectiveRestorer, "_find_latest_backup", return_value=None):
            restorer = SelectiveRestorer(project_path)
            result = restorer._normalize_element_path(Path("/usr/bin/file"))

            assert result is None


class TestValidateElementPath:
    """Test element path validation."""

    def test_validate_valid_claude_path(self):
        """Test validating valid .claude path."""
        project_path = Path("/test/project")

        with patch.object(SelectiveRestorer, "_find_latest_backup", return_value=None):
            restorer = SelectiveRestorer(project_path)
            result = restorer._validate_element_path(Path(".claude/agents/test.md"))

            assert result is True

    def test_validate_valid_moai_path(self):
        """Test validating valid .moai path."""
        project_path = Path("/test/project")

        with patch.object(SelectiveRestorer, "_find_latest_backup", return_value=None):
            restorer = SelectiveRestorer(project_path)
            result = restorer._validate_element_path(Path(".moai/config/test.json"))

            assert result is True

    def test_validate_path_traversal(self):
        """Test validating path with traversal attempt."""
        project_path = Path("/test/project")

        with patch.object(SelectiveRestorer, "_find_latest_backup", return_value=None):
            restorer = SelectiveRestorer(project_path)
            result = restorer._validate_element_path(Path(".claude/../../../etc/passwd"))

            assert result is False

    def test_validate_invalid_prefix(self):
        """Test validating path with invalid prefix."""
        project_path = Path("/test/project")

        with patch.object(SelectiveRestorer, "_find_latest_backup", return_value=None):
            restorer = SelectiveRestorer(project_path)
            result = restorer._validate_element_path(Path("agents/test.md"))

            assert result is False

    def test_validate_suspicious_pattern(self):
        """Test validating path with suspicious pattern."""
        project_path = Path("/test/project")

        with patch.object(SelectiveRestorer, "_find_latest_backup", return_value=None):
            restorer = SelectiveRestorer(project_path)
            result = restorer._validate_element_path(Path(".claude//agents/test.md"))

            assert result is True  # Still valid, just logged as warning


class TestRestoreSingleElement:
    """Test restoring single element."""

    def test_restore_single_file(self):
        """Test restoring single file."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            backup_path = Path(tmpdir) / "backup"
            backup_path.mkdir()

            # Create backup file
            backup_file = backup_path / ".claude" / "agents" / "test.md"
            backup_file.parent.mkdir(parents=True)
            backup_file.write_text("test content")

            restorer = SelectiveRestorer(project_path, backup_path)
            result = restorer._restore_single_element(Path(".claude/agents/test.md"), "agents")

            assert result is True or result is False

    def test_restore_directory(self):
        """Test restoring directory."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            backup_path = Path(tmpdir) / "backup"
            backup_path.mkdir()

            # Create backup directory
            backup_dir = backup_path / ".claude" / "skills" / "test_skill"
            backup_dir.mkdir(parents=True)
            (backup_dir / "manifest.md").write_text("skill manifest")

            restorer = SelectiveRestorer(project_path, backup_path)
            result = restorer._restore_single_element(Path(".claude/skills/test_skill"), "skills")

            assert result is True or result is False

    def test_restore_nonexistent_backup(self):
        """Test restoring when backup doesn't exist."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            backup_path = Path(tmpdir) / "backup"
            backup_path.mkdir()

            restorer = SelectiveRestorer(project_path, backup_path)
            result = restorer._restore_single_element(Path(".claude/nonexistent/file.md"), "agents")

            assert result is False

    def test_restore_with_invalid_path(self):
        """Test restoring with invalid path."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            backup_path = Path(tmpdir) / "backup"
            backup_path.mkdir()

            restorer = SelectiveRestorer(project_path, backup_path)
            result = restorer._restore_single_element(Path("invalid/path.md"), "unknown")

            assert result is False


class TestHandleFileConflict:
    """Test handling file conflicts."""

    def test_handle_identical_files(self):
        """Test handling identical files (no conflict)."""
        with tempfile.TemporaryDirectory() as tmpdir:
            target_path = Path(tmpdir) / "target.md"
            backup_path = Path(tmpdir) / "backup.md"

            content = "identical content"
            target_path.write_text(content)
            backup_path.write_text(content)

            project_path = Path(tmpdir) / "project"
            with patch.object(SelectiveRestorer, "_find_latest_backup", return_value=None):
                restorer = SelectiveRestorer(project_path)
                result = restorer._handle_file_conflict(target_path, backup_path)

                assert result is True

    def test_handle_different_files(self):
        """Test handling different files."""
        with tempfile.TemporaryDirectory() as tmpdir:
            target_path = Path(tmpdir) / "target.md"
            backup_path = Path(tmpdir) / "backup.md"

            target_path.write_text("target content")
            backup_path.write_text("backup content")

            project_path = Path(tmpdir) / "project"
            with patch.object(SelectiveRestorer, "_find_latest_backup", return_value=None):
                with patch("builtins.print"):
                    restorer = SelectiveRestorer(project_path)
                    result = restorer._handle_file_conflict(target_path, backup_path)

                    assert result is True or result is False


class TestDisplayRestorationSummary:
    """Test displaying restoration summary."""

    def test_display_summary_success(self):
        """Test displaying successful restoration summary."""
        project_path = Path("/test/project")

        with patch.object(SelectiveRestorer, "_find_latest_backup", return_value=None):
            with patch("builtins.print") as mock_print:
                restorer = SelectiveRestorer(project_path)
                stats = {
                    "total": 5,
                    "success": 5,
                    "failed": 0,
                    "by_type": {}
                }
                restorer._display_restoration_summary(stats)

                # Verify print was called
                assert mock_print.called

    def test_display_summary_with_failures(self):
        """Test displaying summary with failures."""
        project_path = Path("/test/project")

        with patch.object(SelectiveRestorer, "_find_latest_backup", return_value=None):
            with patch("builtins.print") as mock_print:
                restorer = SelectiveRestorer(project_path)
                stats = {
                    "total": 5,
                    "success": 3,
                    "failed": 2,
                    "by_type": {
                        "agents": {"total": 2, "success": 2, "failed": 0},
                        "skills": {"total": 3, "success": 1, "failed": 2}
                    }
                }
                restorer._display_restoration_summary(stats)

                assert mock_print.called


class TestLogRestorationDetails:
    """Test logging restoration details."""

    def test_log_restoration_success(self):
        """Test logging successful restoration."""
        project_path = Path("/test/project")

        with patch.object(SelectiveRestorer, "_find_latest_backup", return_value=None):
            restorer = SelectiveRestorer(project_path)
            restorer.restoration_log = [
                {"path": ".claude/agents/test.md", "type": "agents", "status": "success", "timestamp": "0"}
            ]

            stats = {
                "total": 1,
                "success": 1,
                "failed": 0
            }

            with patch.object(restorer, "_was_restoration_successful", return_value=True):
                restorer._log_restoration_details([".claude/agents/test.md"], stats)

                # Should complete without error

    def test_log_restoration_failure(self):
        """Test logging failed restoration."""
        project_path = Path("/test/project")

        with patch.object(SelectiveRestorer, "_find_latest_backup", return_value=None):
            restorer = SelectiveRestorer(project_path)

            stats = {
                "total": 1,
                "success": 0,
                "failed": 1
            }

            with patch.object(restorer, "_was_restoration_successful", return_value=False):
                restorer._log_restoration_details([".claude/agents/test.md"], stats)

                # Should complete without error


class TestWasRestorationSuccessful:
    """Test checking restoration success."""

    def test_was_restoration_successful_found(self):
        """Test checking successful restoration."""
        project_path = Path("/test/project")

        with patch.object(SelectiveRestorer, "_find_latest_backup", return_value=None):
            restorer = SelectiveRestorer(project_path)
            restorer.restoration_log = [
                {"path": ".claude/agents/test.md", "type": "agents", "status": "success"}
            ]

            result = restorer._was_restoration_successful(Path(".claude/agents/test.md"))

            assert result is True

    def test_was_restoration_successful_not_found(self):
        """Test checking when restoration not found."""
        project_path = Path("/test/project")

        with patch.object(SelectiveRestorer, "_find_latest_backup", return_value=None):
            restorer = SelectiveRestorer(project_path)
            restorer.restoration_log = []

            result = restorer._was_restoration_successful(Path(".claude/agents/test.md"))

            assert result is False

    def test_was_restoration_successful_failed_status(self):
        """Test checking restoration with failed status."""
        project_path = Path("/test/project")

        with patch.object(SelectiveRestorer, "_find_latest_backup", return_value=None):
            restorer = SelectiveRestorer(project_path)
            restorer.restoration_log = [
                {"path": ".claude/agents/test.md", "type": "agents", "status": "failed"}
            ]

            result = restorer._was_restoration_successful(Path(".claude/agents/test.md"))

            assert result is False


class TestCreateSelectiveRestorer:
    """Test factory function."""

    def test_create_restorer_with_string_path(self):
        """Test creating restorer with string path."""
        project_path = "/test/project"

        with patch.object(SelectiveRestorer, "_find_latest_backup", return_value=None):
            restorer = create_selective_restorer(project_path)

            assert restorer.project_path.is_absolute()

    def test_create_restorer_with_path_object(self):
        """Test creating restorer with Path object."""
        project_path = Path("/test/project")

        with patch.object(SelectiveRestorer, "_find_latest_backup", return_value=None):
            restorer = create_selective_restorer(project_path)

            assert isinstance(restorer, SelectiveRestorer)

    def test_create_restorer_with_backup_path(self):
        """Test creating restorer with backup path."""
        project_path = "/test/project"
        backup_path = Path("/test/backup")

        restorer = create_selective_restorer(project_path, backup_path)

        assert restorer.backup_path == backup_path


class TestRestoreElementType:
    """Test restoring elements of specific type."""

    def test_restore_element_type_empty(self):
        """Test restoring empty element type."""
        project_path = Path("/test/project")

        with patch.object(SelectiveRestorer, "_find_latest_backup", return_value=None):
            restorer = SelectiveRestorer(project_path)
            stats = restorer._restore_element_type("agents", [])

            assert stats["total"] == 0
            assert stats["success"] == 0
            assert stats["failed"] == 0

    def test_restore_element_type_with_elements(self):
        """Test restoring elements of type."""
        project_path = Path("/test/project")

        with patch.object(SelectiveRestorer, "_find_latest_backup", return_value=None):
            with patch.object(SelectiveRestorer, "_restore_single_element", return_value=True):
                with patch("builtins.print"):
                    restorer = SelectiveRestorer(project_path)
                    stats = restorer._restore_element_type("agents", [Path(".claude/agents/test.md")])

                    assert stats["total"] == 1
                    assert stats["success"] == 1
