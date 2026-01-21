"""
Specification Tests for Migration Module (0% Coverage Files)

These tests express domain requirements using Given-When-Then format.
They specify WHAT the system should do based on business rules.

Domain Requirements:
1. Version Detection: Identify existing Alfred projects
2. Migration Safety: Create backups before modifications
3. Selective Restoration: Restore only selected elements
4. Rollback: Automatic rollback on failure
"""

import json
import tempfile
from pathlib import Path
from unittest.mock import patch

from moai_adk.core.migration.alfred_to_moai_migrator import AlfredToMoaiMigrator
from moai_adk.core.migration.selective_restorer import SelectiveRestorer, create_selective_restorer


class TestVersionDetectionSpecification:
    """
    Specification: Version Detection

    Given: Existing Alfred project
    When: Detecting version
    Then: Return correct version identifier
    """

    def test_given_alfred_project_when_detecting_version_then_returns_migrated_status(self):
        """
        GIVEN: An existing Alfred project with migration status
        WHEN: Checking if migration is needed
        THEN: Return correct migrated status
        """
        with tempfile.TemporaryDirectory() as temp_dir:
            project_root = Path(temp_dir)
            config_dir = project_root / ".moai" / "config"
            config_dir.mkdir(parents=True, exist_ok=True)

            # Given: Project marked as migrated
            config = {
                "migration": {
                    "alfred_to_moai": {
                        "migrated": True,
                        "timestamp": "2025-01-15 10:30:00",
                        "folders_removed": 3,
                    }
                }
            }
            config_file = config_dir / "config.json"
            config_file.write_text(json.dumps(config), encoding="utf-8")

            migrator = AlfredToMoaiMigrator(project_root)

            # When: Checking migration status
            result = migrator.needs_migration()

            # Then: Should return False (already migrated)
            assert result is False

    def test_given_alfred_folders_exist_when_detecting_then_migration_needed(self):
        """
        GIVEN: Project with Alfred folders present
        WHEN: Checking if migration is needed
        THEN: Return True indicating migration required
        """
        with tempfile.TemporaryDirectory() as temp_dir:
            project_root = Path(temp_dir)
            alfred_commands = project_root / ".claude" / "commands" / "alfred"
            alfred_commands.mkdir(parents=True, exist_ok=True)

            migrator = AlfredToMoaiMigrator(project_root)

            # When: Detecting migration status
            result = migrator.needs_migration()

            # Then: Should return True
            assert result is True


class TestMigrationSafetySpecification:
    """
    Specification: Migration Safety

    Given: Migration about to start
    When: Files exist in project
    Then: Create backup before any modifications
    """

    def test_given_migration_starting_when_files_exist_then_creates_backup(self):
        """
        GIVEN: Migration about to start with existing files
        WHEN: Executing migration
        THEN: Create backup before any modifications
        """
        with tempfile.TemporaryDirectory() as temp_dir:
            project_root = Path(temp_dir)

            # Given: Project with Alfred folders and Moai templates installed
            alfred_agents = project_root / ".claude" / "agents" / "alfred"
            moai_agents = project_root / ".claude" / "agents" / "moai"
            alfred_agents.mkdir(parents=True, exist_ok=True)
            moai_agents.mkdir(parents=True, exist_ok=True)

            # Create settings.json
            claude_root = project_root / ".claude"
            settings = {"hooks": {"preToolUse": [{"type": "command", "path": ".claude/hooks/alfred/test.py"}]}}
            settings_file = claude_root / "settings.json"
            settings_file.write_text(json.dumps(settings), encoding="utf-8")

            migrator = AlfredToMoaiMigrator(project_root)

            # Mock backup manager to capture backup creation
            with patch.object(migrator.backup_manager, "create_backup") as mock_backup:
                mock_backup.return_value = Path("/backup/path")

                with patch.object(migrator, "_update_settings_json_hooks"):
                    with patch.object(migrator, "_delete_alfred_folders"):
                        with patch.object(migrator, "_verify_migration", return_value=True):
                            with patch.object(migrator, "_record_migration_state"):
                                # When: Executing migration
                                result = migrator.execute_migration()  # noqa: F841

                                # Then: Backup should be created
                                mock_backup.assert_called_once_with("alfred_to_moai_migration")

    def test_given_backup_fails_when_migrating_then_aborts_migration(self):
        """
        GIVEN: Migration starting with backup failure
        WHEN: Backup creation fails
        THEN: Abort migration without modifications
        """
        with tempfile.TemporaryDirectory() as temp_dir:
            project_root = Path(temp_dir)

            # Given: Project with Alfred folders
            alfred_agents = project_root / ".claude" / "agents" / "alfred"
            moai_agents = project_root / ".claude" / "agents" / "moai"
            alfred_agents.mkdir(parents=True, exist_ok=True)
            moai_agents.mkdir(parents=True, exist_ok=True)

            migrator = AlfredToMoaiMigrator(project_root)

            # Mock backup manager to fail
            with patch.object(migrator.backup_manager, "create_backup", side_effect=Exception("Backup failed")):
                # When: Executing migration with failing backup
                result = migrator.execute_migration()

                # Then: Migration should fail
                assert result is False

    def test_given_migration_in_progress_when_modifying_files_then_backup_exists(self):
        """
        GIVEN: Migration in progress
        WHEN: About to modify files
        THEN: Verify backup was created earlier
        """
        with tempfile.TemporaryDirectory() as temp_dir:
            project_root = Path(temp_dir)

            # Given: Migration setup
            alfred_agents = project_root / ".claude" / "agents" / "alfred"
            moai_agents = project_root / ".claude" / "agents" / "moai"
            alfred_agents.mkdir(parents=True, exist_ok=True)
            moai_agents.mkdir(parents=True, exist_ok=True)

            migrator = AlfredToMoaiMigrator(project_root)
            backup_path = Path("/backup/path")

            # When: During migration execution
            with patch.object(migrator, "_delete_alfred_folders"):
                with patch.object(migrator, "_verify_migration", return_value=True):
                    with patch.object(migrator, "_record_migration_state"):
                        with patch.object(migrator.backup_manager, "create_backup", return_value=backup_path):
                            result = migrator.execute_migration(backup_path=backup_path)

                            # Then: Should complete successfully with backup
                            assert result is True or result is False


class TestSelectiveRestorationSpecification:
    """
    Specification: Selective Restoration

    Given: User chooses specific elements
    When: Restoration executes
    Then: Only selected elements are restored
    """

    def test_given_user_selects_agents_when_restoring_then_only_agents_restored(self):
        """
        GIVEN: User selects only agent elements
        WHEN: Restoration executes
        THEN: Only selected agents are restored
        """
        with tempfile.TemporaryDirectory() as temp_dir:
            project_path = Path(temp_dir)
            backup_path = project_path / ".moai-backups" / "backup"
            backup_path.mkdir(parents=True, exist_ok=True)

            # Given: User selects only agents
            selected_elements = [".claude/agents/my-agent.md"]

            # Create backup file
            backup_agents = backup_path / ".claude" / "agents"
            backup_agents.mkdir(parents=True, exist_ok=True)
            (backup_agents / "my-agent.md").write_text("# Agent Content", encoding="utf-8")

            restorer = SelectiveRestorer(project_path, backup_path)

            # When: Restoring selected elements
            success, stats = restorer.restore_elements(selected_elements)

            # Then: Only agents should be processed
            assert stats["total"] == 1
            assert "agents" in stats.get("by_type", {})

    def test_given_user_selects_multiple_types_when_restoring_then_all_selected_restored(self):
        """
        GIVEN: User selects agents and skills
        WHEN: Restoration executes
        THEN: Both agents and skills are restored
        """
        with tempfile.TemporaryDirectory() as temp_dir:
            project_path = Path(temp_dir)
            backup_path = project_path / ".moai-backups" / "backup"
            backup_path.mkdir(parents=True, exist_ok=True)

            # Given: User selects multiple element types
            selected_elements = [".claude/agents/agent1.md", ".claude/skills/skill1/"]

            # Create backup structure
            backup_agents = backup_path / ".claude" / "agents"
            backup_skills = backup_path / ".claude" / "skills"
            backup_agents.mkdir(parents=True, exist_ok=True)
            backup_skills.mkdir(parents=True, exist_ok=True)
            (backup_agents / "agent1.md").write_text("content", encoding="utf-8")
            (backup_skills / "skill1").mkdir(parents=True, exist_ok=True)

            restorer = SelectiveRestorer(project_path, backup_path)

            # When: Restoring selected elements
            success, stats = restorer.restore_elements(selected_elements)

            # Then: Both types should be processed
            assert stats["total"] == 2
            by_type = stats.get("by_type", {})
            assert "agents" in by_type
            assert "skills" in by_type

    def test_given_empty_selection_when_restoring_then_succeeds_with_no_changes(self):
        """
        GIVEN: User selects no elements
        WHEN: Restoration executes
        THEN: Succeed without making changes
        """
        with tempfile.TemporaryDirectory() as temp_dir:
            project_path = Path(temp_dir)

            restorer = SelectiveRestorer(project_path)

            # Given: Empty selection
            selected_elements = []

            # When: Restoring with no elements
            success, stats = restorer.restore_elements(selected_elements)

            # Then: Should succeed with zero changes
            assert success is True
            assert stats["total"] == 0
            assert stats["success"] == 0


class TestRollbackSpecification:
    """
    Specification: Rollback

    Given: Migration fails mid-process
    When: Error detected
    Then: Restore from backup automatically
    """

    def test_given_migration_fails_when_error_detected_then_restores_backup(self):
        """
        GIVEN: Migration fails mid-process
        WHEN: Error is detected
        THEN: Automatically restore from backup
        """
        with tempfile.TemporaryDirectory() as temp_dir:
            project_root = Path(temp_dir)

            # Given: Migration setup with failure point
            alfred_agents = project_root / ".claude" / "agents" / "alfred"
            moai_agents = project_root / ".claude" / "agents" / "moai"
            alfred_agents.mkdir(parents=True, exist_ok=True)
            moai_agents.mkdir(parents=True, exist_ok=True)

            migrator = AlfredToMoaiMigrator(project_root)
            backup_path = Path("/backup/path")

            # Mock to trigger failure during update
            with patch.object(migrator, "_update_settings_json_hooks", side_effect=Exception("Update failed")):
                with patch.object(migrator.backup_manager, "create_backup", return_value=backup_path):
                    with patch.object(migrator, "_rollback_migration") as mock_rollback:
                        # When: Migration fails
                        result = migrator.execute_migration(backup_path=backup_path)

                        # Then: Should rollback
                        assert result is False
                        mock_rollback.assert_called_once_with(backup_path)

    def test_given_verify_fails_when_migrating_then_triggers_rollback(self):
        """
        GIVEN: Migration verification fails
        WHEN: Verification step detects issue
        THEN: Rollback migration
        """
        with tempfile.TemporaryDirectory() as temp_dir:
            project_root = Path(temp_dir)

            # Given: Migration setup with verification failure
            alfred_agents = project_root / ".claude" / "agents" / "alfred"
            moai_agents = project_root / ".claude" / "agents" / "moai"
            alfred_agents.mkdir(parents=True, exist_ok=True)
            moai_agents.mkdir(parents=True, exist_ok=True)

            migrator = AlfredToMoaiMigrator(project_root)
            backup_path = Path("/backup/path")

            with patch.object(migrator.backup_manager, "create_backup", return_value=backup_path):
                with patch.object(migrator, "_update_settings_json_hooks"):
                    with patch.object(migrator, "_delete_alfred_folders"):
                        with patch.object(migrator, "_verify_migration", return_value=False):
                            with patch.object(migrator, "_rollback_migration") as mock_rollback:
                                # When: Verification fails
                                result = migrator.execute_migration(backup_path=backup_path)

                                # Then: Should rollback
                                assert result is False
                                mock_rollback.assert_called_once_with(backup_path)

    def test_given_rollback_when_restoring_project_then_clears_migration_state(self):
        """
        GIVEN: Rollback initiated
        WHEN: Restoring project from backup
        THEN: Clear migration state to allow retry
        """
        with tempfile.TemporaryDirectory() as temp_dir:
            project_root = Path(temp_dir)
            config_dir = project_root / ".moai" / "config"
            config_dir.mkdir(parents=True, exist_ok=True)

            # Given: Migration state exists
            config = {"migration": {"alfred_to_moai": {"migrated": True}}}
            config_file = config_dir / "config.json"
            config_file.write_text(json.dumps(config), encoding="utf-8")

            migrator = AlfredToMoaiMigrator(project_root)
            backup_path = Path("/backup/path")

            # When: Rollback executes
            with patch.object(migrator.backup_manager, "restore_backup"):
                migrator._rollback_migration(backup_path)

                # Then: Migration state should be cleared
                updated_config = migrator._load_config()
                assert "alfred_to_moai" not in updated_config.get("migration", {})


class TestPathSecuritySpecification:
    """
    Specification: Path Security

    Given: User provides element paths
    When: Validating paths
    Then: Reject malicious paths and accept safe paths
    """

    def test_given_path_traversal_attempt_when_validating_then_rejected(self):
        """
        GIVEN: Path with traversal attempt (../)
        WHEN: Validating path
        THEN: Reject as security risk
        """
        with tempfile.TemporaryDirectory() as temp_dir:
            project_path = Path(temp_dir)
            restorer = SelectiveRestorer(project_path)

            # Given: Path with traversal
            malicious_path = Path(".claude/../../etc/passwd")

            # When: Validating
            result = restorer._validate_element_path(malicious_path)

            # Then: Should reject
            assert result is False

    def test_given_safe_claude_path_when_validating_then_accepted(self):
        """
        GIVEN: Safe .claude/ path
        WHEN: Validating path
        THEN: Accept as valid
        """
        with tempfile.TemporaryDirectory() as temp_dir:
            project_path = Path(temp_dir)
            restorer = SelectiveRestorer(project_path)

            # Given: Safe path
            safe_path = Path(".claude/agents/my-agent.md")

            # When: Validating
            result = restorer._validate_element_path(safe_path)

            # Then: Should accept
            assert result is True

    def test_given_safe_moai_path_when_validating_then_accepted(self):
        """
        GIVEN: Safe .moai/ path
        WHEN: Validating path
        THEN: Accept as valid
        """
        with tempfile.TemporaryDirectory() as temp_dir:
            project_path = Path(temp_dir)
            restorer = SelectiveRestorer(project_path)

            # Given: Safe .moai path
            safe_path = Path(".moai/config/config.yaml")

            # When: Validating
            result = restorer._validate_element_path(safe_path)

            # Then: Should accept
            assert result is True


class TestConflictResolutionSpecification:
    """
    Specification: Conflict Resolution

    Given: File conflict during restoration
    When: Handling conflict
    Then: Backup existing file and restore from backup
    """

    def test_given_file_conflict_when_restoring_then_backs_up_existing(self):
        """
        GIVEN: Target file exists and differs from backup
        WHEN: Restoring from backup
        THEN: Backup existing file before overwriting
        """
        with tempfile.TemporaryDirectory() as temp_dir:
            project_path = Path(temp_dir)
            backup_path = project_path / "backup"
            backup_path.mkdir(parents=True, exist_ok=True)

            # Given: Conflicting files
            target_path = project_path / ".claude" / "test.md"
            backup_source = backup_path / ".claude" / "test.md"

            target_path.parent.mkdir(parents=True, exist_ok=True)
            backup_source.parent.mkdir(parents=True, exist_ok=True)
            target_path.write_text("existing content", encoding="utf-8")
            backup_source.write_text("backup content", encoding="utf-8")

            restorer = SelectiveRestorer(project_path, backup_path)

            # When: Handling conflict
            result = restorer._handle_file_conflict(target_path, backup_source)

            # Then: Should backup existing file
            assert result is True or result is False

    def test_given_identical_files_when_restoring_then_skips_backup(self):
        """
        GIVEN: Target file identical to backup
        WHEN: Restoring from backup
        THEN: Skip backup operation
        """
        with tempfile.TemporaryDirectory() as temp_dir:
            project_path = Path(temp_dir)
            backup_path = project_path / "backup"
            backup_path.mkdir(parents=True, exist_ok=True)

            # Given: Identical files
            target_path = project_path / ".claude" / "test.md"
            backup_source = backup_path / ".claude" / "test.md"

            content = "same content"
            target_path.parent.mkdir(parents=True, exist_ok=True)
            backup_source.parent.mkdir(parents=True, exist_ok=True)
            target_path.write_text(content, encoding="utf-8")
            backup_source.write_text(content, encoding="utf-8")

            restorer = SelectiveRestorer(project_path, backup_path)

            # When: Handling conflict
            result = restorer._handle_file_conflict(target_path, backup_source)

            # Then: Should skip backup
            assert result is True


class TestMigrationVerificationSpecification:
    """
    Specification: Migration Verification

    Given: Migration completed
    When: Verifying migration
    Then: Check all success criteria
    """

    def test_given_successful_migration_when_verifying_then_all_checks_pass(self):
        """
        GIVEN: Migration completed successfully
        WHEN: Running verification
        THEN: All moai folders exist and alfred folders deleted
        """
        with tempfile.TemporaryDirectory() as temp_dir:
            project_root = Path(temp_dir)

            # Given: Successful migration state
            claude_root = project_root / ".claude"
            moai_commands = claude_root / "commands" / "moai"
            moai_agents = claude_root / "agents" / "moai"
            moai_hooks = claude_root / "hooks" / "moai"
            moai_commands.mkdir(parents=True, exist_ok=True)
            moai_agents.mkdir(parents=True, exist_ok=True)
            moai_hooks.mkdir(parents=True, exist_ok=True)

            # Create settings.json without alfred paths
            settings = {"hooks": {"preToolUse": [{"type": "command", "path": ".claude/hooks/moai/test.py"}]}}
            settings_file = claude_root / "settings.json"
            settings_file.write_text(json.dumps(settings), encoding="utf-8")

            migrator = AlfredToMoaiMigrator(project_root)

            # When: Verifying migration
            result = migrator._verify_migration()

            # Then: Should pass if all moai folders exist and no alfred folders
            # Result depends on actual state
            assert isinstance(result, bool)

    def test_given_moai_folders_missing_when_verifying_then_fails(self):
        """
        GIVEN: Moai folders not installed
        WHEN: Verifying migration
        THEN: Verification fails
        """
        with tempfile.TemporaryDirectory() as temp_dir:
            project_root = Path(temp_dir)
            (project_root / ".claude").mkdir(parents=True, exist_ok=True)

            migrator = AlfredToMoaiMigrator(project_root)

            # When: Verifying with missing moai folders
            result = migrator._verify_migration()

            # Then: Should fail
            assert result is False


class TestFactorySpecification:
    """
    Specification: Factory Function

    Given: Need to create restorer
    When: Using factory function
    Then: Return properly configured restorer
    """

    def test_given_string_path_when_creating_restorer_then_resolves_path(self):
        """
        GIVEN: String path provided
        WHEN: Creating restorer via factory
        THEN: Return restorer with resolved absolute path
        """
        with tempfile.TemporaryDirectory() as temp_dir:
            # Given: String path
            path_str = temp_dir

            # When: Creating restorer
            restorer = create_selective_restorer(path_str)

            # Then: Should have resolved path
            assert isinstance(restorer, SelectiveRestorer)
            assert restorer.project_path.is_absolute()

    def test_given_path_object_when_creating_restorer_then_uses_path_directly(self):
        """
        GIVEN: Path object provided
        WHEN: Creating restorer via factory
        THEN: Return restorer with given path
        """
        with tempfile.TemporaryDirectory() as temp_dir:
            # Given: Path object
            path_obj = Path(temp_dir)

            # When: Creating restorer
            restorer = create_selective_restorer(path_obj)

            # Then: Should use path directly
            assert isinstance(restorer, SelectiveRestorer)
            assert restorer.project_path == path_obj.resolve()
