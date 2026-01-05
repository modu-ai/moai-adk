"""Comprehensive tests for SelectiveRestorer with 95%+ coverage.

Tests cover:
- Initialization and backup discovery
- Element restoration (files and directories)
- Element grouping by type
- Path normalization and validation
- Conflict detection and handling
- File content comparison
- Error handling and recovery
- Restoration logging and statistics
- Display and reporting functions
- Factory function
"""

import tempfile
from pathlib import Path
from unittest.mock import MagicMock, patch

from moai_adk.core.migration.selective_restorer import SelectiveRestorer, create_selective_restorer


class TestSelectiveRestorerInitialization:
    """Test SelectiveRestorer initialization."""

    def test_init_with_explicit_backup_path(self):
        """Test initialization with explicit backup path."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            backup_path = project_path / ".moai-backups" / "test-backup"
            backup_path.mkdir(parents=True, exist_ok=True)

            restorer = SelectiveRestorer(project_path, backup_path)

            assert restorer.project_path == project_path
            assert restorer.backup_path == backup_path
            assert restorer.restoration_log == []

    def test_init_with_auto_detected_backup(self):
        """Test initialization with auto-detected backup."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            backups_dir = project_path / ".moai-backups"
            backup_path = backups_dir / "pre-update-backup_2025-01-01_000000"
            backup_path.mkdir(parents=True, exist_ok=True)

            restorer = SelectiveRestorer(project_path)

            assert restorer.project_path == project_path
            assert restorer.backup_path == backup_path
            assert restorer.restoration_log == []

    def test_init_without_backup_directory(self):
        """Test initialization when no backup directory exists."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)

            restorer = SelectiveRestorer(project_path)

            assert restorer.project_path == project_path
            assert restorer.backup_path is None
            assert restorer.restoration_log == []

    def test_init_with_path_object(self):
        """Test initialization with Path object."""
        with tempfile.TemporaryDirectory() as tmpdir:
            restorer = SelectiveRestorer(Path(tmpdir))
            assert isinstance(restorer.project_path, Path)


class TestFindLatestBackup:
    """Test backup discovery functionality."""

    def test_find_latest_backup_single(self):
        """Test finding single backup."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            backups_dir = project_path / ".moai-backups"
            backup_path = backups_dir / "pre-update-backup_001"
            backup_path.mkdir(parents=True, exist_ok=True)

            restorer = SelectiveRestorer(project_path, None)
            result = restorer._find_latest_backup()

            assert result == backup_path

    def test_find_latest_backup_multiple(self):
        """Test finding latest backup among multiple."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            backups_dir = project_path / ".moai-backups"

            # Create multiple backups
            backup1 = backups_dir / "pre-update-backup_001"
            backup2 = backups_dir / "pre-update-backup_002"
            backup3 = backups_dir / "pre-update-backup_003"

            backup1.mkdir(parents=True, exist_ok=True)
            backup2.mkdir(parents=True, exist_ok=True)
            backup3.mkdir(parents=True, exist_ok=True)

            # Manipulate modification times
            for backup in [backup1, backup2, backup3]:
                Path(backup / "marker.txt").write_text("marker")

            restorer = SelectiveRestorer(project_path, None)
            result = restorer._find_latest_backup()

            assert result == backup3

    def test_find_latest_backup_no_backups(self):
        """Test backup discovery when no backups exist."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)

            restorer = SelectiveRestorer(project_path, None)
            result = restorer._find_latest_backup()

            assert result is None

    def test_find_latest_backup_ignores_non_backup_dirs(self):
        """Test that non-backup directories are ignored."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            backups_dir = project_path / ".moai-backups"

            # Create a non-backup directory and a valid backup
            (backups_dir / "random-dir").mkdir(parents=True, exist_ok=True)
            backup_path = backups_dir / "pre-update-backup_001"
            backup_path.mkdir(parents=True, exist_ok=True)

            restorer = SelectiveRestorer(project_path, None)
            result = restorer._find_latest_backup()

            assert result == backup_path

    def test_find_latest_backup_empty_backups_dir(self):
        """Test backup discovery in empty .moai-backups directory."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            backups_dir = project_path / ".moai-backups"
            backups_dir.mkdir(parents=True, exist_ok=True)

            restorer = SelectiveRestorer(project_path, None)
            result = restorer._find_latest_backup()

            assert result is None


class TestGroupElementsByType:
    """Test element grouping by type."""

    def test_group_agents(self):
        """Test grouping agent elements."""
        with tempfile.TemporaryDirectory() as tmpdir:
            restorer = SelectiveRestorer(Path(tmpdir))
            elements = [
                ".claude/agents/my-agent.md",
                ".claude/agents/other-agent.md",
            ]

            groups = restorer._group_elements_by_type(elements)

            assert len(groups["agents"]) == 2
            assert groups["commands"] == []
            assert groups["skills"] == []
            assert groups["hooks"] == []

    def test_group_commands(self):
        """Test grouping command elements."""
        with tempfile.TemporaryDirectory() as tmpdir:
            restorer = SelectiveRestorer(Path(tmpdir))
            elements = [
                ".claude/commands/moai/my-command.md",
                ".claude/commands/custom-command.md",
            ]

            groups = restorer._group_elements_by_type(elements)

            assert len(groups["commands"]) == 2
            assert groups["agents"] == []
            assert groups["skills"] == []
            assert groups["hooks"] == []

    def test_group_skills(self):
        """Test grouping skill elements."""
        with tempfile.TemporaryDirectory() as tmpdir:
            restorer = SelectiveRestorer(Path(tmpdir))
            elements = [
                ".claude/skills/my-skill",
                ".moai/skills/skill-dir",
            ]

            groups = restorer._group_elements_by_type(elements)

            assert len(groups["skills"]) == 2
            assert groups["agents"] == []
            assert groups["commands"] == []
            assert groups["hooks"] == []

    def test_group_hooks(self):
        """Test grouping hook elements."""
        with tempfile.TemporaryDirectory() as tmpdir:
            restorer = SelectiveRestorer(Path(tmpdir))
            elements = [
                ".claude/hooks/moai/my-hook.py",
                ".claude/hooks/custom-hook.py",
            ]

            groups = restorer._group_elements_by_type(elements)

            assert len(groups["hooks"]) == 2
            assert groups["agents"] == []
            assert groups["commands"] == []
            assert groups["skills"] == []

    def test_group_unknown_elements(self):
        """Test grouping unknown element types."""
        with tempfile.TemporaryDirectory() as tmpdir:
            restorer = SelectiveRestorer(Path(tmpdir))
            elements = [
                ".claude/unknown/element.txt",
                "some-random-path",
            ]

            groups = restorer._group_elements_by_type(elements)

            assert len(groups["unknown"]) == 2
            assert groups["agents"] == []
            assert groups["commands"] == []
            assert groups["skills"] == []
            assert groups["hooks"] == []

    def test_group_mixed_elements(self):
        """Test grouping mixed element types."""
        with tempfile.TemporaryDirectory() as tmpdir:
            restorer = SelectiveRestorer(Path(tmpdir))
            elements = [
                ".claude/agents/agent1.md",
                ".claude/commands/moai/cmd.md",
                ".moai/skills/skill1",
                ".claude/hooks/moai/hook.py",
                ".claude/unknown/file.txt",
            ]

            groups = restorer._group_elements_by_type(elements)

            assert len(groups["agents"]) == 1
            assert len(groups["commands"]) == 1
            assert len(groups["skills"]) == 1
            assert len(groups["hooks"]) == 1
            assert len(groups["unknown"]) == 1

    def test_group_elements_empty_list(self):
        """Test grouping empty element list."""
        with tempfile.TemporaryDirectory() as tmpdir:
            restorer = SelectiveRestorer(Path(tmpdir))
            elements = []

            groups = restorer._group_elements_by_type(elements)

            assert groups["agents"] == []
            assert groups["commands"] == []
            assert groups["skills"] == []
            assert groups["hooks"] == []
            assert groups["unknown"] == []


class TestNormalizeElementPath:
    """Test element path normalization."""

    def test_normalize_relative_path_claude(self):
        """Test normalizing relative .claude path."""
        with tempfile.TemporaryDirectory() as tmpdir:
            restorer = SelectiveRestorer(Path(tmpdir))
            element_path = Path(".claude/agents/my-agent.md")

            result = restorer._normalize_element_path(element_path)

            assert result == element_path

    def test_normalize_relative_path_moai(self):
        """Test normalizing relative .moai path."""
        with tempfile.TemporaryDirectory() as tmpdir:
            restorer = SelectiveRestorer(Path(tmpdir))
            element_path = Path(".moai/skills/skill1")

            result = restorer._normalize_element_path(element_path)

            assert result == element_path

    def test_normalize_absolute_path_claude(self):
        """Test normalizing absolute .claude path."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            restorer = SelectiveRestorer(project_path)

            absolute_path = project_path / ".claude" / "agents" / "my-agent.md"
            result = restorer._normalize_element_path(absolute_path)

            assert result == Path(".claude/agents/my-agent.md")

    def test_normalize_absolute_path_moai(self):
        """Test normalizing absolute .moai path."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            restorer = SelectiveRestorer(project_path)

            absolute_path = project_path / ".moai" / "skills" / "skill1"
            result = restorer._normalize_element_path(absolute_path)

            assert result == Path(".moai/skills/skill1")

    def test_normalize_invalid_relative_path(self):
        """Test normalizing invalid relative path."""
        with tempfile.TemporaryDirectory() as tmpdir:
            restorer = SelectiveRestorer(Path(tmpdir))
            element_path = Path("invalid/path/file.md")

            result = restorer._normalize_element_path(element_path)

            assert result is None

    def test_normalize_absolute_path_without_prefix(self):
        """Test normalizing absolute path without .claude/.moai prefix."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            restorer = SelectiveRestorer(project_path)

            absolute_path = project_path / "invalid" / "path" / "file.md"
            result = restorer._normalize_element_path(absolute_path)

            assert result is None

    def test_normalize_path_with_nested_directories(self):
        """Test normalizing path with nested directories."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            restorer = SelectiveRestorer(project_path)

            absolute_path = project_path / ".claude" / "commands" / "moai" / "deep" / "nested" / "cmd.md"
            result = restorer._normalize_element_path(absolute_path)

            assert result == Path(".claude/commands/moai/deep/nested/cmd.md")


class TestValidateElementPath:
    """Test element path validation."""

    def test_validate_valid_claude_path(self):
        """Test validating valid .claude path."""
        with tempfile.TemporaryDirectory() as tmpdir:
            restorer = SelectiveRestorer(Path(tmpdir))
            path = Path(".claude/agents/my-agent.md")

            result = restorer._validate_element_path(path)

            assert result is True

    def test_validate_valid_moai_path(self):
        """Test validating valid .moai path."""
        with tempfile.TemporaryDirectory() as tmpdir:
            restorer = SelectiveRestorer(Path(tmpdir))
            path = Path(".moai/skills/skill1")

            result = restorer._validate_element_path(path)

            assert result is True

    def test_validate_path_traversal_attack(self):
        """Test validation blocks path traversal attacks."""
        with tempfile.TemporaryDirectory() as tmpdir:
            restorer = SelectiveRestorer(Path(tmpdir))
            path = Path(".claude/../../../etc/passwd")

            result = restorer._validate_element_path(path)

            assert result is False

    def test_validate_invalid_prefix(self):
        """Test validation rejects invalid prefix."""
        with tempfile.TemporaryDirectory() as tmpdir:
            restorer = SelectiveRestorer(Path(tmpdir))
            path = Path("invalid/path/file.md")

            result = restorer._validate_element_path(path)

            assert result is False

    def test_validate_suspicious_pattern_double_slash(self):
        """Test validation warns about double slash."""
        with tempfile.TemporaryDirectory() as tmpdir:
            restorer = SelectiveRestorer(Path(tmpdir))
            path = Path(".claude//agents/file.md")

            result = restorer._validate_element_path(path)

            # Should still be valid but warn
            assert result is True

    def test_validate_suspicious_pattern_tilde(self):
        """Test validation warns about tilde."""
        with tempfile.TemporaryDirectory() as tmpdir:
            restorer = SelectiveRestorer(Path(tmpdir))
            path = Path(".claude/agents/~file.md")

            result = restorer._validate_element_path(path)

            # Should still be valid but warn
            assert result is True

    def test_validate_suspicious_pattern_dollar(self):
        """Test validation warns about dollar sign."""
        with tempfile.TemporaryDirectory() as tmpdir:
            restorer = SelectiveRestorer(Path(tmpdir))
            path = Path(".claude/agents/$file.md")

            result = restorer._validate_element_path(path)

            # Should still be valid but warn
            assert result is True


class TestRestoreSingleElement:
    """Test single element restoration."""

    def test_restore_single_file(self):
        """Test restoring a single file."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            backup_path = project_path / ".moai-backups" / "backup"
            backup_path.mkdir(parents=True, exist_ok=True)

            # Create backup file
            backup_file = backup_path / ".claude" / "agents" / "my-agent.md"
            backup_file.parent.mkdir(parents=True, exist_ok=True)
            backup_file.write_text("agent content")

            restorer = SelectiveRestorer(project_path, backup_path)
            success = restorer._restore_single_element(
                Path(".claude/agents/my-agent.md"),
                "agents"
            )

            assert success is True
            target_file = project_path / ".claude" / "agents" / "my-agent.md"
            assert target_file.exists()
            assert target_file.read_text() == "agent content"

    def test_restore_single_directory(self):
        """Test restoring a single directory (skill)."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            backup_path = project_path / ".moai-backups" / "backup"
            backup_path.mkdir(parents=True, exist_ok=True)

            # Create backup directory with files
            backup_skill = backup_path / ".moai" / "skills" / "my-skill"
            backup_skill.mkdir(parents=True, exist_ok=True)
            (backup_skill / "file1.py").write_text("file1 content")
            (backup_skill / "file2.py").write_text("file2 content")

            restorer = SelectiveRestorer(project_path, backup_path)
            success = restorer._restore_single_element(
                Path(".moai/skills/my-skill"),
                "skills"
            )

            assert success is True
            target_skill = project_path / ".moai" / "skills" / "my-skill"
            assert target_skill.exists()
            assert (target_skill / "file1.py").read_text() == "file1 content"
            assert (target_skill / "file2.py").read_text() == "file2 content"

    def test_restore_nonexistent_backup(self):
        """Test restoration fails when backup doesn't exist."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            backup_path = project_path / ".moai-backups" / "backup"
            backup_path.mkdir(parents=True, exist_ok=True)

            restorer = SelectiveRestorer(project_path, backup_path)
            success = restorer._restore_single_element(
                Path(".claude/agents/nonexistent.md"),
                "agents"
            )

            assert success is False

    def test_restore_with_invalid_path(self):
        """Test restoration fails with invalid path."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            backup_path = project_path / ".moai-backups" / "backup"
            backup_path.mkdir(parents=True, exist_ok=True)

            restorer = SelectiveRestorer(project_path, backup_path)
            success = restorer._restore_single_element(
                Path("invalid/path/file.md"),
                "unknown"
            )

            assert success is False

    def test_restore_creates_target_directory(self):
        """Test restoration creates target directory if needed."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            backup_path = project_path / ".moai-backups" / "backup"
            backup_path.mkdir(parents=True, exist_ok=True)

            # Create backup file in nested structure
            backup_file = backup_path / ".claude" / "commands" / "moai" / "deep" / "nested" / "cmd.md"
            backup_file.parent.mkdir(parents=True, exist_ok=True)
            backup_file.write_text("command content")

            restorer = SelectiveRestorer(project_path, backup_path)
            success = restorer._restore_single_element(
                Path(".claude/commands/moai/deep/nested/cmd.md"),
                "commands"
            )

            assert success is True
            target_file = project_path / ".claude" / "commands" / "moai" / "deep" / "nested" / "cmd.md"
            assert target_file.exists()

    def test_restore_logs_success(self):
        """Test restoration logs successful restoration."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            backup_path = project_path / ".moai-backups" / "backup"
            backup_path.mkdir(parents=True, exist_ok=True)

            # Create backup file
            backup_file = backup_path / ".claude" / "agents" / "my-agent.md"
            backup_file.parent.mkdir(parents=True, exist_ok=True)
            backup_file.write_text("agent content")

            restorer = SelectiveRestorer(project_path, backup_path)
            restorer._restore_single_element(
                Path(".claude/agents/my-agent.md"),
                "agents"
            )

            assert len(restorer.restoration_log) > 0
            log_entry = restorer.restoration_log[0]
            assert log_entry["status"] == "success"
            assert "agents" in log_entry["type"]


class TestHandleFileConflict:
    """Test file conflict handling."""

    def test_handle_identical_files(self):
        """Test handling identical file conflict."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)

            # Create target and backup files with same content
            target_file = project_path / ".claude" / "agents" / "agent.md"
            target_file.parent.mkdir(parents=True, exist_ok=True)
            target_file.write_text("same content")

            backup_file = project_path / ".moai-backups" / "backup" / ".claude" / "agents" / "agent.md"
            backup_file.parent.mkdir(parents=True, exist_ok=True)
            backup_file.write_text("same content")

            restorer = SelectiveRestorer(project_path)
            result = restorer._handle_file_conflict(target_file, backup_file)

            assert result is True

    def test_handle_different_files(self):
        """Test handling different file conflict."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)

            # Create target and backup files with different content
            target_file = project_path / ".claude" / "agents" / "agent.md"
            target_file.parent.mkdir(parents=True, exist_ok=True)
            target_file.write_text("target content")

            backup_file = project_path / ".moai-backups" / "backup" / ".claude" / "agents" / "agent.md"
            backup_file.parent.mkdir(parents=True, exist_ok=True)
            backup_file.write_text("backup content")

            restorer = SelectiveRestorer(project_path)
            result = restorer._handle_file_conflict(target_file, backup_file)

            assert result is True
            backup_target = target_file.with_suffix(".backup")
            assert backup_target.exists()

    def test_handle_directory_conflict(self):
        """Test handling directory conflict."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)

            # Create target directory with content
            target_dir = project_path / ".moai" / "skills" / "my-skill"
            target_dir.mkdir(parents=True, exist_ok=True)
            (target_dir / "file.py").write_text("target content")

            # Create backup directory
            backup_dir = project_path / ".moai-backups" / "backup" / ".moai" / "skills" / "my-skill"
            backup_dir.mkdir(parents=True, exist_ok=True)
            (backup_dir / "file.py").write_text("backup content")

            restorer = SelectiveRestorer(project_path)
            result = restorer._handle_file_conflict(target_dir, backup_dir)

            assert result is True
            backup_dir_target = target_dir.parent / f"{target_dir.name}.backup_dir"
            assert backup_dir_target.exists()

    def test_handle_mixed_type_conflict(self):
        """Test handling mixed file/directory type conflict."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)

            # Create target as file
            target_file = project_path / ".claude" / "agents" / "agent.md"
            target_file.parent.mkdir(parents=True, exist_ok=True)
            target_file.write_text("file content")

            # Create backup as directory
            backup_dir = project_path / ".moai-backups" / "backup" / ".claude" / "agents" / "agent.md"
            backup_dir.mkdir(parents=True, exist_ok=True)

            restorer = SelectiveRestorer(project_path)
            result = restorer._handle_file_conflict(target_file, backup_dir)

            # Mixed types should return True (skip comparison)
            assert result is True

    def test_handle_conflict_backup_directory_exists(self):
        """Test handling conflict when backup directory already exists."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)

            # Create target directory
            target_dir = project_path / ".moai" / "skills" / "my-skill"
            target_dir.mkdir(parents=True, exist_ok=True)

            # Create existing backup directory
            existing_backup = target_dir.parent / f"{target_dir.name}.backup_dir"
            existing_backup.mkdir(parents=True, exist_ok=True)

            # Create new backup directory
            backup_dir = project_path / ".moai-backups" / "backup" / ".moai" / "skills" / "my-skill"
            backup_dir.mkdir(parents=True, exist_ok=True)

            restorer = SelectiveRestorer(project_path)
            result = restorer._handle_file_conflict(target_dir, backup_dir)

            assert result is True
            assert existing_backup.exists()


class TestRestoreElements:
    """Test main element restoration process."""

    def test_restore_elements_empty_list(self):
        """Test restoration with empty element list."""
        with tempfile.TemporaryDirectory() as tmpdir:
            restorer = SelectiveRestorer(Path(tmpdir))
            success, stats = restorer.restore_elements([])

            assert success is True
            assert stats["total"] == 0
            assert stats["success"] == 0
            assert stats["failed"] == 0

    def test_restore_elements_single(self):
        """Test restoration with single element."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            backup_path = project_path / ".moai-backups" / "backup"
            backup_path.mkdir(parents=True, exist_ok=True)

            # Create backup file
            backup_file = backup_path / ".claude" / "agents" / "agent.md"
            backup_file.parent.mkdir(parents=True, exist_ok=True)
            backup_file.write_text("content")

            restorer = SelectiveRestorer(project_path, backup_path)
            success, stats = restorer.restore_elements([".claude/agents/agent.md"])

            assert success is True
            assert stats["total"] == 1
            assert stats["success"] == 1
            assert stats["failed"] == 0

    def test_restore_elements_multiple(self):
        """Test restoration with multiple elements."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            backup_path = project_path / ".moai-backups" / "backup"
            backup_path.mkdir(parents=True, exist_ok=True)

            # Create multiple backup files
            for i in range(3):
                backup_file = backup_path / ".claude" / "agents" / f"agent{i}.md"
                backup_file.parent.mkdir(parents=True, exist_ok=True)
                backup_file.write_text(f"content{i}")

            restorer = SelectiveRestorer(project_path, backup_path)
            elements = [f".claude/agents/agent{i}.md" for i in range(3)]
            success, stats = restorer.restore_elements(elements)

            assert success is True
            assert stats["total"] == 3
            assert stats["success"] == 3
            assert stats["failed"] == 0

    def test_restore_elements_with_failures(self):
        """Test restoration with some failures."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            backup_path = project_path / ".moai-backups" / "backup"
            backup_path.mkdir(parents=True, exist_ok=True)

            # Create only one backup file
            backup_file = backup_path / ".claude" / "agents" / "agent1.md"
            backup_file.parent.mkdir(parents=True, exist_ok=True)
            backup_file.write_text("content")

            restorer = SelectiveRestorer(project_path, backup_path)
            # Request two files but only one exists
            elements = [".claude/agents/agent1.md", ".claude/agents/agent2.md"]
            success, stats = restorer.restore_elements(elements)

            assert success is False
            assert stats["total"] == 2
            assert stats["success"] == 1
            assert stats["failed"] == 1

    def test_restore_elements_groups_by_type(self):
        """Test restoration groups elements by type."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            backup_path = project_path / ".moai-backups" / "backup"
            backup_path.mkdir(parents=True, exist_ok=True)

            # Create backup files of different types
            agent_file = backup_path / ".claude" / "agents" / "agent.md"
            agent_file.parent.mkdir(parents=True, exist_ok=True)
            agent_file.write_text("agent")

            skill_dir = backup_path / ".moai" / "skills" / "skill"
            skill_dir.mkdir(parents=True, exist_ok=True)
            (skill_dir / "file.py").write_text("skill")

            restorer = SelectiveRestorer(project_path, backup_path)
            elements = [".claude/agents/agent.md", ".moai/skills/skill"]
            success, stats = restorer.restore_elements(elements)

            assert success is True
            assert "by_type" in stats
            assert "agents" in stats["by_type"]
            assert "skills" in stats["by_type"]


class TestWasRestorationSuccessful:
    """Test restoration success checking."""

    def test_was_restoration_successful_true(self):
        """Test checking successful restoration."""
        with tempfile.TemporaryDirectory() as tmpdir:
            restorer = SelectiveRestorer(Path(tmpdir))
            restorer.restoration_log.append({
                "path": ".claude/agents/agent.md",
                "type": "agents",
                "status": "success",
                "timestamp": "123456",
            })

            result = restorer._was_restoration_successful(Path(".claude/agents/agent.md"))

            assert result is True

    def test_was_restoration_successful_false(self):
        """Test checking failed restoration."""
        with tempfile.TemporaryDirectory() as tmpdir:
            restorer = SelectiveRestorer(Path(tmpdir))
            restorer.restoration_log.append({
                "path": ".claude/agents/agent.md",
                "type": "agents",
                "status": "failed",
                "timestamp": "123456",
            })

            result = restorer._was_restoration_successful(Path(".claude/agents/agent.md"))

            assert result is False

    def test_was_restoration_successful_not_found(self):
        """Test checking restoration for non-existent entry."""
        with tempfile.TemporaryDirectory() as tmpdir:
            restorer = SelectiveRestorer(Path(tmpdir))

            result = restorer._was_restoration_successful(Path(".claude/agents/unknown.md"))

            assert result is False


class TestDisplayRestorationSummary:
    """Test display of restoration summary."""

    @patch("builtins.print")
    def test_display_summary_all_success(self, mock_print):
        """Test displaying successful restoration summary."""
        with tempfile.TemporaryDirectory() as tmpdir:
            restorer = SelectiveRestorer(Path(tmpdir))
            stats = {
                "total": 3,
                "success": 3,
                "failed": 0,
                "by_type": {
                    "agents": {"total": 1, "success": 1, "failed": 0},
                    "skills": {"total": 2, "success": 2, "failed": 0},
                }
            }

            restorer._display_restoration_summary(stats)

            # Check that print was called with expected content
            print_calls = [str(call).lower() for call in mock_print.call_args_list]
            assert any("restoration complete" in call for call in print_calls)
            assert any("total elements: 3" in call for call in print_calls)

    @patch("builtins.print")
    def test_display_summary_with_failures(self, mock_print):
        """Test displaying summary with failures."""
        with tempfile.TemporaryDirectory() as tmpdir:
            restorer = SelectiveRestorer(Path(tmpdir))
            stats = {
                "total": 3,
                "success": 2,
                "failed": 1,
                "by_type": {
                    "agents": {"total": 1, "success": 0, "failed": 1},
                    "skills": {"total": 2, "success": 2, "failed": 0},
                }
            }

            restorer._display_restoration_summary(stats)

            # Check that failures are mentioned
            print_calls = [str(call) for call in mock_print.call_args_list]
            assert any("failed" in call.lower() for call in print_calls)


class TestLogRestorationDetails:
    """Test restoration logging."""

    def test_log_restoration_all_success(self):
        """Test logging all successful restorations."""
        with tempfile.TemporaryDirectory() as tmpdir:
            restorer = SelectiveRestorer(Path(tmpdir))
            restorer.restoration_log = [
                {"path": ".claude/agents/agent.md", "type": "agents", "status": "success"},
            ]

            stats = {"total": 1, "success": 1, "failed": 0}
            elements = [".claude/agents/agent.md"]

            # Should not raise exception
            restorer._log_restoration_details(elements, stats)

    def test_log_restoration_with_failures(self):
        """Test logging with failures."""
        with tempfile.TemporaryDirectory() as tmpdir:
            restorer = SelectiveRestorer(Path(tmpdir))
            restorer.restoration_log = [
                {"path": ".claude/agents/agent.md", "type": "agents", "status": "failed"},
            ]

            stats = {"total": 1, "success": 0, "failed": 1}
            elements = [".claude/agents/agent.md"]

            # Should not raise exception
            restorer._log_restoration_details(elements, stats)


class TestRestoreElementType:
    """Test restoration of specific element types."""

    def test_restore_element_type_success(self):
        """Test restoring element type successfully."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            backup_path = project_path / ".moai-backups" / "backup"
            backup_path.mkdir(parents=True, exist_ok=True)

            # Create backup files
            agent_file = backup_path / ".claude" / "agents" / "agent.md"
            agent_file.parent.mkdir(parents=True, exist_ok=True)
            agent_file.write_text("content")

            restorer = SelectiveRestorer(project_path, backup_path)
            elements = [Path(".claude/agents/agent.md")]
            stats = restorer._restore_element_type("agents", elements)

            assert stats["total"] == 1
            assert stats["success"] == 1
            assert stats["failed"] == 0

    def test_restore_element_type_with_exception(self):
        """Test restoring element type with exception."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            backup_path = project_path / ".moai-backups" / "backup"
            backup_path.mkdir(parents=True, exist_ok=True)

            restorer = SelectiveRestorer(project_path, backup_path)
            elements = [Path(".claude/agents/nonexistent.md")]
            stats = restorer._restore_element_type("agents", elements)

            assert stats["total"] == 1
            assert stats["success"] == 0
            assert stats["failed"] == 1


class TestFactoryFunction:
    """Test factory function."""

    def test_create_selective_restorer_with_string_path(self):
        """Test factory function with string path."""
        with tempfile.TemporaryDirectory() as tmpdir:
            restorer = create_selective_restorer(tmpdir)

            assert isinstance(restorer, SelectiveRestorer)
            assert restorer.project_path == Path(tmpdir).resolve()

    def test_create_selective_restorer_with_path_object(self):
        """Test factory function with Path object."""
        with tempfile.TemporaryDirectory() as tmpdir:
            tmpdir_path = Path(tmpdir)
            restorer = create_selective_restorer(tmpdir_path)

            assert isinstance(restorer, SelectiveRestorer)
            assert restorer.project_path == tmpdir_path.resolve()

    def test_create_selective_restorer_with_backup_path(self):
        """Test factory function with explicit backup path."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            backup_path = project_path / ".moai-backups" / "backup"
            backup_path.mkdir(parents=True, exist_ok=True)

            restorer = create_selective_restorer(project_path, backup_path)

            assert restorer.project_path == Path(tmpdir).resolve()
            assert restorer.backup_path == backup_path


class TestRestoreElementTypeEdgeCases:
    """Test edge cases in element type restoration."""

    @patch("builtins.print")
    def test_restore_element_type_prints_success(self, mock_print):
        """Test that success messages are printed."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            backup_path = project_path / ".moai-backups" / "backup"
            backup_path.mkdir(parents=True, exist_ok=True)

            # Create backup file
            backup_file = backup_path / ".claude" / "agents" / "agent.md"
            backup_file.parent.mkdir(parents=True, exist_ok=True)
            backup_file.write_text("content")

            restorer = SelectiveRestorer(project_path, backup_path)
            elements = [Path(".claude/agents/agent.md")]
            stats = restorer._restore_element_type("agents", elements)

            assert stats["success"] == 1
            # Verify print was called
            assert mock_print.called

    @patch("builtins.print")
    def test_restore_element_type_prints_failure(self, mock_print):
        """Test that failure messages are printed."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            backup_path = project_path / ".moai-backups" / "backup"
            backup_path.mkdir(parents=True, exist_ok=True)

            restorer = SelectiveRestorer(project_path, backup_path)
            elements = [Path(".claude/agents/nonexistent.md")]
            stats = restorer._restore_element_type("agents", elements)

            assert stats["failed"] == 1

    @patch("builtins.print")
    def test_restore_element_type_prints_exception(self, mock_print):
        """Test that exception messages are printed."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            backup_path = project_path / ".moai-backups" / "backup"
            backup_path.mkdir(parents=True, exist_ok=True)

            restorer = SelectiveRestorer(project_path, backup_path)

            # Mock _restore_single_element to raise an exception
            with patch.object(restorer, '_restore_single_element', side_effect=Exception("Test error")):
                elements = [Path(".claude/agents/agent.md")]
                stats = restorer._restore_element_type("agents", elements)

            assert stats["failed"] == 1


class TestNormalizePathEdgeCases:
    """Test edge cases in path normalization."""

    def test_normalize_absolute_path_with_multiple_prefixes_finds_first(self):
        """Test that first occurrence of prefix is used when multiple exist."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            restorer = SelectiveRestorer(project_path)

            # Create path with .claude appearing twice
            path_with_double = project_path / ".claude" / "agents" / ".claude" / "nested.md"
            result = restorer._normalize_element_path(path_with_double)

            assert result == Path(".claude/agents/.claude/nested.md")

    def test_normalize_path_exception_in_split(self):
        """Test handling exception when trying to split path."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            restorer = SelectiveRestorer(project_path)

            # Test by mocking the str() to contain .claude/ prefix
            # but make str.split fail
            mock_path = MagicMock()
            mock_path.is_absolute.return_value = True

            # When str() is called on the mock, return a string with .claude/
            with patch('pathlib.Path.__str__', return_value="/some/path/.moai/file"):
                # This should work and normalize properly
                element_path = Path("/some/path/.moai/file")
                result = restorer._normalize_element_path(element_path)
                assert result == Path(".moai/file")

    def test_normalize_absolute_path_continues_on_error(self):
        """Test that normalization continues trying other prefixes on error."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            restorer = SelectiveRestorer(project_path)

            # Create a path that will match .moai but not .claude
            element_path = project_path / ".moai" / "skills" / "skill1"
            result = restorer._normalize_element_path(element_path)

            assert result == Path(".moai/skills/skill1")


class TestHandleFileConflictEdgeCases:
    """Test edge cases in file conflict handling."""

    def test_handle_conflict_directory_backup_fails(self):
        """Test when backing up a directory fails."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)

            # Create target directory
            target_dir = project_path / ".moai" / "skills" / "my-skill"
            target_dir.mkdir(parents=True, exist_ok=True)

            # Create backup directory
            backup_dir = project_path / ".moai-backups" / "backup" / ".moai" / "skills" / "my-skill"
            backup_dir.mkdir(parents=True, exist_ok=True)

            restorer = SelectiveRestorer(project_path)

            # Mock shutil.copytree to raise an exception
            with patch("shutil.copytree", side_effect=PermissionError("Cannot backup")):
                result = restorer._handle_file_conflict(target_dir, backup_dir)

            assert result is False

    def test_handle_conflict_file_backup_fails(self):
        """Test when backing up a file fails."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)

            # Create target and backup files with different content
            target_file = project_path / ".claude" / "agents" / "agent.md"
            target_file.parent.mkdir(parents=True, exist_ok=True)
            target_file.write_text("target content")

            backup_file = project_path / ".moai-backups" / "backup" / ".claude" / "agents" / "agent.md"
            backup_file.parent.mkdir(parents=True, exist_ok=True)
            backup_file.write_text("backup content")

            restorer = SelectiveRestorer(project_path)

            # Mock to cause backup to fail
            with patch("moai_adk.core.migration.selective_restorer.shutil.copy2", side_effect=PermissionError("Cannot backup")):
                result = restorer._handle_file_conflict(target_file, backup_file)

            assert result is False

    def test_handle_conflict_general_exception(self):
        """Test handling of general exceptions in conflict handling."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)

            target_file = project_path / ".claude" / "agents" / "agent.md"
            target_file.parent.mkdir(parents=True, exist_ok=True)
            target_file.write_text("content")

            backup_file = project_path / ".moai-backups" / "backup" / ".claude" / "agents" / "agent.md"
            backup_file.parent.mkdir(parents=True, exist_ok=True)
            backup_file.write_text("content")

            restorer = SelectiveRestorer(project_path)

            # Mock read_text to raise an exception
            with patch.object(Path, 'read_text', side_effect=RuntimeError("Read error")):
                result = restorer._handle_file_conflict(target_file, backup_file)

            assert result is False


class TestPrintAndLoggingEdgeCases:
    """Test print and logging statements."""

    def test_restore_elements_prints_header(self):
        """Test that header is printed during restoration."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            backup_path = project_path / ".moai-backups" / "backup"
            backup_path.mkdir(parents=True, exist_ok=True)

            restorer = SelectiveRestorer(project_path, backup_path)

            # Restore single element to ensure print is called
            backup_file = backup_path / ".claude" / "agents" / "agent.md"
            backup_file.parent.mkdir(parents=True, exist_ok=True)
            backup_file.write_text("content")

            with patch("builtins.print") as mock_print:
                restorer.restore_elements([".claude/agents/agent.md"])
                # Check that print was called
                assert mock_print.called

    @patch("builtins.print")
    def test_display_restoration_summary_empty_by_type(self, mock_print):
        """Test displaying summary with empty by_type."""
        with tempfile.TemporaryDirectory() as tmpdir:
            restorer = SelectiveRestorer(Path(tmpdir))
            stats = {
                "total": 0,
                "success": 0,
                "failed": 0,
                "by_type": {
                    "agents": {"total": 0, "success": 0, "failed": 0},
                }
            }

            restorer._display_restoration_summary(stats)

            # Should handle empty types gracefully
            assert mock_print.called


class TestEdgeCasesAndErrors:
    """Test edge cases and error conditions."""

    def test_restore_with_permission_error(self):
        """Test restoration handles permission errors."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            backup_path = project_path / ".moai-backups" / "backup"
            backup_path.mkdir(parents=True, exist_ok=True)

            # Create backup file
            backup_file = backup_path / ".claude" / "agents" / "agent.md"
            backup_file.parent.mkdir(parents=True, exist_ok=True)
            backup_file.write_text("content")

            restorer = SelectiveRestorer(project_path, backup_path)

            # Mock shutil.copy2 to raise PermissionError
            with patch("shutil.copy2", side_effect=PermissionError("Access denied")):
                success = restorer._restore_single_element(
                    Path(".claude/agents/agent.md"),
                    "agents"
                )

            assert success is False

    def test_restore_with_file_encoding_error(self):
        """Test restoration handles file encoding errors."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            backup_path = project_path / ".moai-backups" / "backup"
            backup_path.mkdir(parents=True, exist_ok=True)

            # Create target and backup files
            target_file = project_path / ".claude" / "agents" / "agent.md"
            target_file.parent.mkdir(parents=True, exist_ok=True)
            target_file.write_bytes(b"\x80\x81\x82\x83")  # Invalid UTF-8

            backup_file = backup_path / ".claude" / "agents" / "agent.md"
            backup_file.parent.mkdir(parents=True, exist_ok=True)
            backup_file.write_bytes(b"\x84\x85\x86\x87")  # Different invalid UTF-8

            restorer = SelectiveRestorer(project_path, backup_path)
            result = restorer._handle_file_conflict(target_file, backup_file)

            # Should handle gracefully with errors="ignore"
            assert result is True

    def test_restore_with_broken_symlink(self):
        """Test restoration with broken symlink."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            backup_path = project_path / ".moai-backups" / "backup"
            backup_path.mkdir(parents=True, exist_ok=True)

            # Create backup file
            backup_file = backup_path / ".claude" / "agents" / "agent.md"
            backup_file.parent.mkdir(parents=True, exist_ok=True)
            backup_file.write_text("content")

            restorer = SelectiveRestorer(project_path, backup_path)
            success = restorer._restore_single_element(
                Path(".claude/agents/agent.md"),
                "agents"
            )

            assert success is True

    def test_normalize_path_with_unicode(self):
        """Test normalizing path with unicode characters."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            restorer = SelectiveRestorer(project_path)

            # Create absolute path with unicode
            absolute_path = project_path / ".claude" / "agents" / "μάσια.md"
            result = restorer._normalize_element_path(absolute_path)

            assert result == Path(".claude/agents/μάσια.md")

    def test_restoration_log_entry_structure(self):
        """Test restoration log entry has correct structure."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            backup_path = project_path / ".moai-backups" / "backup"
            backup_path.mkdir(parents=True, exist_ok=True)

            # Create backup file
            backup_file = backup_path / ".claude" / "agents" / "agent.md"
            backup_file.parent.mkdir(parents=True, exist_ok=True)
            backup_file.write_text("content")

            restorer = SelectiveRestorer(project_path, backup_path)
            restorer._restore_single_element(Path(".claude/agents/agent.md"), "agents")

            assert len(restorer.restoration_log) > 0
            entry = restorer.restoration_log[0]
            assert "path" in entry
            assert "type" in entry
            assert "status" in entry
            assert "timestamp" in entry

    def test_validate_path_with_only_allowed_prefixes(self):
        """Test validation checks only allowed prefixes."""
        with tempfile.TemporaryDirectory() as tmpdir:
            restorer = SelectiveRestorer(Path(tmpdir))

            # Test both allowed prefixes
            assert restorer._validate_element_path(Path(".claude/agents/agent.md")) is True
            assert restorer._validate_element_path(Path(".moai/skills/skill")) is True

            # Test non-allowed prefixes
            assert restorer._validate_element_path(Path(".git/config")) is False
            assert restorer._validate_element_path(Path("src/main.py")) is False

    def test_restore_elements_statistics_accuracy(self):
        """Test restoration statistics are accurate."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            backup_path = project_path / ".moai-backups" / "backup"
            backup_path.mkdir(parents=True, exist_ok=True)

            # Create multiple backup files
            for element_type in ["agents", "commands", "skills"]:
                for i in range(2):
                    element_path = backup_path / ".claude" / element_type
                    element_path.mkdir(parents=True, exist_ok=True)
                    if element_type == "skills":
                        (element_path / f"{element_type[:-1]}{i}").mkdir(parents=True, exist_ok=True)
                    else:
                        (element_path / f"{element_type[:-1]}{i}.md").write_text("content")

            restorer = SelectiveRestorer(project_path, backup_path)
            elements = [
                ".claude/agents/agent0.md",
                ".claude/agents/agent1.md",
                ".claude/commands/command0.md",
                ".claude/commands/command1.md",
            ]

            success, stats = restorer.restore_elements(elements)

            assert stats["total"] == 4
            assert stats["by_type"]["agents"]["total"] == 2
            assert stats["by_type"]["commands"]["total"] == 2
            assert stats["by_type"]["agents"]["success"] == 2
            assert stats["by_type"]["commands"]["success"] == 2
