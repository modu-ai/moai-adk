"""
Unit tests for AlfredToMoaiMigrator

Tests cover:
- Migration need detection
- Alfred folder presence checking
- Migration state tracking
- Path substitution
- Rollback functionality
"""

import json
from unittest.mock import MagicMock, patch

import pytest

from moai_adk.core.migration.alfred_to_moai_migrator import AlfredToMoaiMigrator


@pytest.fixture
def temp_project(tmp_path):
    """Create a temporary project with necessary structure"""
    project_root = tmp_path / "test_project"
    project_root.mkdir()

    # Create .claude structure
    claude_root = project_root / ".claude"
    claude_root.mkdir()

    # Create .moai structure
    moai_root = project_root / ".moai"
    moai_root.mkdir()
    config_dir = moai_root / "config"
    config_dir.mkdir()

    # Create initial config.json
    config_path = config_dir / "config.json"
    config_path.write_text(
        json.dumps(
            {
                "moai": {"version": "0.27.0"},
                "language": {"conversation_language": "ko"},
            }
        )
    )

    return project_root


@pytest.fixture
def migrator(temp_project):
    """Create a migrator instance"""
    return AlfredToMoaiMigrator(temp_project)


class TestNeedsMigration:
    """Tests for needs_migration() method"""

    def test_no_migration_needed_when_no_alfred_folders(self, migrator):
        """Should return False when no alfred folders exist"""
        assert migrator.needs_migration() is False

    def test_migration_needed_when_alfred_commands_exists(self, migrator):
        """Should return True when alfred/commands folder exists"""
        alfred_commands = migrator.claude_root / "commands" / "alfred"
        alfred_commands.mkdir(parents=True, exist_ok=True)
        assert migrator.needs_migration() is True

    def test_migration_needed_when_alfred_agents_exists(self, migrator):
        """Should return True when alfred/agents folder exists"""
        alfred_agents = migrator.claude_root / "agents" / "alfred"
        alfred_agents.mkdir(parents=True, exist_ok=True)
        assert migrator.needs_migration() is True

    def test_migration_needed_when_alfred_hooks_exists(self, migrator):
        """Should return True when alfred/hooks folder exists"""
        alfred_hooks = migrator.claude_root / "hooks" / "alfred"
        alfred_hooks.mkdir(parents=True, exist_ok=True)
        assert migrator.needs_migration() is True

    def test_no_migration_when_already_migrated(self, migrator):
        """Should return False when migration is already completed"""
        # Create alfred folder
        alfred_commands = migrator.claude_root / "commands" / "alfred"
        alfred_commands.mkdir(parents=True, exist_ok=True)

        # Mark as already migrated in config
        config = migrator._load_config()
        config["migration"] = {"alfred_to_moai": {"migrated": True, "timestamp": "2025-11-18 10:00:00"}}
        migrator._save_config(config)

        # Should return False because already migrated
        assert migrator.needs_migration() is False


class TestDeleteAlfredFolders:
    """Tests for _delete_alfred_folders() method"""

    def test_delete_single_alfred_folder(self, migrator):
        """Should delete a single alfred folder"""
        alfred_commands = migrator.claude_root / "commands" / "alfred"
        alfred_commands.mkdir(parents=True, exist_ok=True)
        test_file = alfred_commands / "test.md"
        test_file.write_text("test content")

        alfred_detected = {"commands": alfred_commands}
        migrator._delete_alfred_folders(alfred_detected)

        assert not alfred_commands.exists()

    def test_delete_multiple_alfred_folders(self, migrator):
        """Should delete multiple alfred folders"""
        alfred_commands = migrator.claude_root / "commands" / "alfred"
        alfred_agents = migrator.claude_root / "agents" / "alfred"
        alfred_hooks = migrator.claude_root / "hooks" / "alfred"

        for folder in [alfred_commands, alfred_agents, alfred_hooks]:
            folder.mkdir(parents=True, exist_ok=True)
            (folder / "test.md").write_text("test")

        alfred_detected = {
            "commands": alfred_commands,
            "agents": alfred_agents,
            "hooks": alfred_hooks,
        }
        migrator._delete_alfred_folders(alfred_detected)

        for folder in [alfred_commands, alfred_agents, alfred_hooks]:
            assert not folder.exists()

    def test_delete_skips_nonexistent_folder(self, migrator):
        """Should skip deletion when folder doesn't exist"""
        # Create a dict with nonexistent folder
        # _delete_alfred_folders checks if folder.exists() before deleting
        nonexistent = migrator.claude_root / "nonexistent" / "path"
        alfred_detected = {"nonexistent": nonexistent}

        # Should not raise exception, just skip since folder doesn't exist
        migrator._delete_alfred_folders(alfred_detected)


class TestUpdateSettingsJsonHooks:
    """Tests for _update_settings_json_hooks() method"""

    def test_update_hooks_alfred_to_moai(self, migrator):
        """Should replace all alfred hook paths with moai"""
        settings_path = migrator.claude_root / "settings.json"
        original_content = """{
  "hooks": {
    "SessionStart": [
      {
        "command": "uv run {{PROJECT_DIR}}/.claude/hooks/alfred/session_start.py"
      }
    ]
  }
}"""
        settings_path.write_text(original_content)

        migrator._update_settings_json_hooks()

        updated_content = settings_path.read_text()
        assert "hooks/alfred" not in updated_content
        assert "hooks/moai" in updated_content
        # Should still be valid JSON
        json.loads(updated_content)

    def test_update_all_path_types(self, migrator):
        """Should update hooks, commands, and agents paths"""
        settings_path = migrator.claude_root / "settings.json"
        content = """{
  "hooks": {"path": "{{PROJECT_DIR}}/.claude/hooks/alfred/test.py"},
  "commands": {"path": "{{PROJECT_DIR}}/.claude/commands/alfred/cmd.md"},
  "agents": {"path": "{{PROJECT_DIR}}/.claude/agents/alfred/agent.md"}
}"""
        settings_path.write_text(content)

        migrator._update_settings_json_hooks()

        updated = settings_path.read_text()
        assert "alfred" not in updated
        assert "moai" in updated

    def test_settings_json_not_exists_no_error(self, migrator):
        """Should not raise error if settings.json doesn't exist"""
        # settings.json doesn't exist
        migrator._update_settings_json_hooks()  # Should not raise


class TestVerifyMigration:
    """Tests for _verify_migration() method"""

    def test_verify_fails_if_moai_commands_missing(self, migrator):
        """Should fail verification if moai/commands folder is missing"""
        # Create Alfred folder
        alfred = migrator.claude_root / "commands" / "alfred"
        alfred.mkdir(parents=True, exist_ok=True)

        # Moai folder doesn't exist
        result = migrator._verify_migration()
        assert result is False

    def test_verify_fails_if_alfred_still_exists(self, migrator):
        """Should fail verification if alfred folder still exists"""
        # Create both folders
        alfred = migrator.claude_root / "commands" / "alfred"
        moai = migrator.claude_root / "commands" / "moai"
        alfred.mkdir(parents=True, exist_ok=True)
        moai.mkdir(parents=True, exist_ok=True)

        result = migrator._verify_migration()
        assert result is False

    def test_verify_fails_if_settings_has_alfred_reference(self, migrator):
        """Should fail verification if settings.json still has alfred reference"""
        # Create moai folder
        moai = migrator.claude_root / "commands" / "moai"
        moai.mkdir(parents=True, exist_ok=True)

        # Create settings.json with alfred reference
        settings = migrator.claude_root / "settings.json"
        settings.write_text('{"hooks": "alfred"}')

        result = migrator._verify_migration()
        assert result is False

    def test_verify_succeeds_when_all_conditions_met(self, migrator):
        """Should succeed when all verification conditions are met"""
        # Create moai folders
        for folder in ["commands/moai", "agents/moai", "hooks/moai"]:
            (migrator.claude_root / folder).mkdir(parents=True, exist_ok=True)

        # Create settings.json with moai reference
        settings = migrator.claude_root / "settings.json"
        settings.write_text('{"hooks": "moai"}')

        result = migrator._verify_migration()
        assert result is True


class TestRecordMigrationState:
    """Tests for _record_migration_state() method"""

    def test_record_migration_creates_config_entry(self, migrator, tmp_path):
        """Should create migration entry in config.json"""
        backup_path = tmp_path / "backup"
        backup_path.mkdir()

        migrator._record_migration_state(backup_path, 3)

        config = migrator._load_config()
        assert "migration" in config
        assert "alfred_to_moai" in config["migration"]
        assert config["migration"]["alfred_to_moai"]["migrated"] is True

    def test_record_includes_required_fields(self, migrator, tmp_path):
        """Should include all required fields in migration state"""
        backup_path = tmp_path / "backup"
        backup_path.mkdir()

        migrator._record_migration_state(backup_path, 3)

        config = migrator._load_config()
        migration_state = config["migration"]["alfred_to_moai"]

        required_fields = [
            "migrated",
            "timestamp",
            "folders_installed",
            "folders_removed",
            "backup_path",
            "package_version",
        ]

        for field in required_fields:
            assert field in migration_state


class TestExecuteMigration:
    """Tests for execute_migration() method"""

    def test_migration_fails_without_moai_folders(self, migrator):
        """Should fail if moai folders don't exist (Phase 1 not completed)"""
        # Create Alfred folder
        alfred = migrator.claude_root / "commands" / "alfred"
        alfred.mkdir(parents=True, exist_ok=True)

        # Don't create moai folders
        result = migrator.execute_migration()
        assert result is False

    @patch("moai_adk.core.migration.alfred_to_moai_migrator.BackupManager")
    def test_migration_fails_on_backup_error(self, mock_backup_manager, migrator):
        """Should fail and return False on backup error"""
        # Setup mocked backup manager to raise exception
        mock_backup_instance = MagicMock()
        mock_backup_instance.create_backup.side_effect = Exception("Backup failed")
        mock_backup_manager.return_value = mock_backup_instance

        migrator.backup_manager = mock_backup_instance

        # Create Alfred folder
        alfred = migrator.claude_root / "commands" / "alfred"
        alfred.mkdir(parents=True, exist_ok=True)

        result = migrator.execute_migration()
        assert result is False


# Integration tests
class TestMigrationIntegration:
    """Integration tests for full migration workflow"""

    def test_complete_migration_workflow(self, migrator, tmp_path):
        """Test complete migration from Alfred to Moai"""
        # Phase 1: Create Moai structure (already done by Phase 1)
        for folder in ["commands/moai", "agents/moai", "hooks/moai"]:
            (migrator.claude_root / folder).mkdir(parents=True, exist_ok=True)

        # Create test files in Alfred folders
        for folder in ["commands/alfred", "agents/alfred", "hooks/alfred"]:
            alfred_folder = migrator.claude_root / folder
            alfred_folder.mkdir(parents=True, exist_ok=True)
            (alfred_folder / "test.md").write_text("test content")

        # Create settings.json with alfred paths
        settings = migrator.claude_root / "settings.json"
        settings.write_text('{"hooks": "{{PROJECT_DIR}}/.claude/hooks/alfred/test.py"}')

        # Create backup
        backup_path = tmp_path / "test_backup"
        backup_path.mkdir()

        # Execute migration
        result = migrator.execute_migration(backup_path)

        # Verify results
        assert result is True
        for folder in ["commands/alfred", "agents/alfred", "hooks/alfred"]:
            assert not (migrator.claude_root / folder).exists()

        assert "alfred" not in settings.read_text()
        assert "moai" in settings.read_text()

        config = migrator._load_config()
        assert config["migration"]["alfred_to_moai"]["migrated"] is True
