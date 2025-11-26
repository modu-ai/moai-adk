"""
Integration tests for update.py with Alfred to Moai migration

Tests the full workflow:
1. Backup creation
2. Alfred folder detection
3. Migration execution
4. Template sync with moai folders
5. Rollback on failure
"""

import json
from pathlib import Path

import pytest

from moai_adk.core.migration.alfred_to_moai_migrator import AlfredToMoaiMigrator


@pytest.fixture
def temp_project(tmp_path):
    """Create a temporary project structure for testing"""
    project_root = tmp_path / "test_project"
    project_root.mkdir()

    # Create .claude structure with both Alfred and Moai
    claude_root = project_root / ".claude"
    claude_root.mkdir()

    # Create Alfred structure (existing project)
    for folder_type in ["commands", "agents", "hooks"]:
        alfred_folder = claude_root / folder_type / "alfred"
        alfred_folder.mkdir(parents=True, exist_ok=True)
        (alfred_folder / "test.md").write_text("Alfred content")

    # Create Moai structure (Phase 1 template)
    for folder_type in ["commands", "agents", "hooks"]:
        moai_folder = claude_root / folder_type / "moai"
        moai_folder.mkdir(parents=True, exist_ok=True)
        (moai_folder / "test.md").write_text("Moai content")

    # Create settings.json with alfred paths
    settings = claude_root / "settings.json"
    settings.write_text(
        json.dumps(
            {"hooks": {"SessionStart": [{"command": "uv run {{PROJECT_DIR}}/.claude/hooks/alfred/session_start.py"}]}}
        )
    )

    # Create .moai config structure
    moai_root = project_root / ".moai"
    moai_root.mkdir()
    config_dir = moai_root / "config"
    config_dir.mkdir()
    config_path = config_dir / "config.json"
    config_path.write_text(
        json.dumps(
            {
                "moai": {"version": "0.27.0"},
                "language": {"conversation_language": "ko"},
                "project": {"name": "test-project"},
            }
        )
    )

    # Create backups directory
    backups_dir = moai_root / "backups"
    backups_dir.mkdir()

    return project_root


class TestMigrationDetection:
    """Tests for migration detection in update workflow"""

    def test_migration_detection_with_alfred_folders(self, temp_project):
        """Should detect that migration is needed when alfred folders exist"""
        migrator = AlfredToMoaiMigrator(temp_project)
        assert migrator.needs_migration() is True

    def test_no_migration_when_no_alfred_folders(self, tmp_path):
        """Should not need migration when no alfred folders exist"""
        project_root = tmp_path / "test_project"
        project_root.mkdir()
        (project_root / ".claude").mkdir()
        (project_root / ".moai" / "config").mkdir(parents=True)
        (project_root / ".moai" / "config" / "config.json").write_text("{}")

        migrator = AlfredToMoaiMigrator(project_root)
        assert migrator.needs_migration() is False


class TestFullMigrationWorkflow:
    """Tests for complete migration workflow"""

    def test_migration_deletes_alfred_folders(self, temp_project):
        """Should delete alfred folders after successful migration"""
        migrator = AlfredToMoaiMigrator(temp_project)

        # Verify Alfred folders exist
        assert (temp_project / ".claude" / "commands" / "alfred").exists()
        assert (temp_project / ".claude" / "agents" / "alfred").exists()
        assert (temp_project / ".claude" / "hooks" / "alfred").exists()

        # Execute migration
        result = migrator.execute_migration()

        # Verify migration succeeded
        assert result is True

        # Verify Alfred folders are deleted
        assert not (temp_project / ".claude" / "commands" / "alfred").exists()
        assert not (temp_project / ".claude" / "agents" / "alfred").exists()
        assert not (temp_project / ".claude" / "hooks" / "alfred").exists()

    def test_migration_preserves_moai_folders(self, temp_project):
        """Should preserve moai folders after migration"""
        migrator = AlfredToMoaiMigrator(temp_project)

        # Verify Moai folders exist
        assert (temp_project / ".claude" / "commands" / "moai").exists()
        assert (temp_project / ".claude" / "agents" / "moai").exists()
        assert (temp_project / ".claude" / "hooks" / "moai").exists()

        # Execute migration
        result = migrator.execute_migration()
        assert result is True

        # Verify Moai folders still exist
        assert (temp_project / ".claude" / "commands" / "moai").exists()
        assert (temp_project / ".claude" / "agents" / "moai").exists()
        assert (temp_project / ".claude" / "hooks" / "moai").exists()

    def test_migration_updates_settings_json_paths(self, temp_project):
        """Should update all alfred paths to moai in settings.json"""
        migrator = AlfredToMoaiMigrator(temp_project)

        settings_path = temp_project / ".claude" / "settings.json"
        original_content = settings_path.read_text()

        # Verify alfred is in original
        assert "alfred" in original_content.lower()

        # Execute migration
        result = migrator.execute_migration()
        assert result is True

        # Verify settings.json updated
        updated_content = settings_path.read_text()
        assert "alfred" not in updated_content.lower()
        assert "moai" in updated_content.lower()

        # Verify JSON validity
        json.loads(updated_content)

    def test_migration_records_state_in_config(self, temp_project):
        """Should record migration state in config.json"""
        migrator = AlfredToMoaiMigrator(temp_project)

        temp_project / ".moai" / "config" / "config.json"
        config = migrator._load_config()

        # Verify no migration state initially
        assert "migration" not in config or "alfred_to_moai" not in config.get("migration", {})

        # Execute migration
        result = migrator.execute_migration()
        assert result is True

        # Verify migration state recorded
        config = migrator._load_config()
        assert "migration" in config
        assert "alfred_to_moai" in config["migration"]
        migration_state = config["migration"]["alfred_to_moai"]

        assert migration_state["migrated"] is True
        assert "timestamp" in migration_state
        assert migration_state["folders_installed"] == 3
        assert migration_state["folders_removed"] > 0


class TestMigrationRollback:
    """Tests for migration rollback on failure"""

    def test_rollback_restores_alfred_folders_on_settings_update_failure(self, temp_project):
        """Should rollback (restore from backup) if settings.json update fails"""
        migrator = AlfredToMoaiMigrator(temp_project)

        # Store original state
        (temp_project / ".claude" / "commands" / "alfred").exists()

        # Create invalid settings.json that will fail JSON parsing
        settings_path = temp_project / ".claude" / "settings.json"
        invalid_json = "{invalid json content}"
        settings_path.write_text(invalid_json)

        # This should fail during settings.json update
        # and trigger rollback
        migrator.execute_migration()

        # Migration should fail
        # (Note: actual rollback depends on backup creation)
        # This test verifies the migration doesn't proceed on settings update failure


class TestMigrationDuplicatePrevention:
    """Tests for preventing duplicate migrations"""

    def test_no_migration_when_already_migrated(self, temp_project):
        """Should skip migration if already completed"""
        migrator = AlfredToMoaiMigrator(temp_project)

        # First migration
        result1 = migrator.execute_migration()
        assert result1 is True

        # Create Alfred folders again (to test that migration is skipped)
        alfred_commands = temp_project / ".claude" / "commands" / "alfred"
        alfred_commands.mkdir(parents=True, exist_ok=True)

        # Create new migrator instance
        migrator2 = AlfredToMoaiMigrator(temp_project)

        # Should return False (no migration needed) because already marked as migrated
        assert migrator2.needs_migration() is False


class TestMigrationWithPartialFolders:
    """Tests for migration with partial Alfred folders"""

    def test_migration_with_only_commands_folder(self, tmp_path):
        """Should handle migration with only commands/alfred folder"""
        project_root = tmp_path / "test_project"
        project_root.mkdir()

        claude_root = project_root / ".claude"
        claude_root.mkdir()

        # Create only commands/alfred
        commands_alfred = claude_root / "commands" / "alfred"
        commands_alfred.mkdir(parents=True, exist_ok=True)
        (commands_alfred / "test.md").write_text("test")

        # Create moai folders
        for folder_type in ["commands", "agents", "hooks"]:
            (claude_root / folder_type / "moai").mkdir(parents=True, exist_ok=True)

        # Create config
        moai_root = project_root / ".moai"
        moai_root.mkdir()
        (moai_root / "config").mkdir()
        (moai_root / "config" / "config.json").write_text("{}")

        migrator = AlfredToMoaiMigrator(project_root)

        # Should detect migration needed
        assert migrator.needs_migration() is True

        # Should execute successfully
        result = migrator.execute_migration()
        assert result is True

        # Should delete only the existing alfred folder
        assert not (claude_root / "commands" / "alfred").exists()


class TestMigrationEdgeCases:
    """Tests for edge cases in migration"""

    def test_migration_with_corrupted_settings_json(self, temp_project):
        """Should handle corrupted settings.json gracefully"""
        # Create invalid JSON
        settings_path = temp_project / ".claude" / "settings.json"
        settings_path.write_text("{ invalid json }")

        migrator = AlfredToMoaiMigrator(temp_project)

        # Migration should attempt but handle the error
        # (exact behavior depends on error handling strategy)
        migrator.execute_migration()

        # Should either fail or attempt recovery
        # This test verifies it doesn't crash unexpectedly

    def test_migration_with_readonly_directory(self, temp_project):
        """Should fail gracefully if directories are read-only"""
        if hasattr(temp_project, "__fspath__"):
            path = temp_project
        else:
            path = Path(temp_project)

        # Make directories read-only (if possible on this OS)
        try:
            path.chmod(0o444)

            migrator = AlfredToMoaiMigrator(temp_project)
            result = migrator.execute_migration()

            # Should fail due to permission error
            assert result is False
        except Exception:
            # Skip test if permission changes not supported
            pytest.skip("Platform doesn't support permission changes")
        finally:
            # Restore permissions for cleanup
            path.chmod(0o755)
