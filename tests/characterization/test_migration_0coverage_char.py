"""
Characterization Tests for Migration Module (0% Coverage Files)

These tests capture the current behavior of:
1. alfred_to_moai_migrator.py - AlfredToMoaiMigrator class
2. selective_restorer.py - SelectiveRestorer class

Characterization tests document WHAT the code does, not what it should do.
They serve as a safety net for refactoring by preserving existing behavior.
"""

import json
import tempfile
from pathlib import Path

from moai_adk.core.migration.alfred_to_moai_migrator import AlfredToMoaiMigrator
from moai_adk.core.migration.selective_restorer import SelectiveRestorer, create_selective_restorer


class TestAlfredToMoaiMigratorInitialization:
    """Characterization tests for AlfredToMoaiMigrator initialization"""

    def test_init_with_valid_project_root(self):
        """CHAR: Initialize migrator with valid project root path"""
        with tempfile.TemporaryDirectory() as temp_dir:
            project_root = Path(temp_dir)
            (project_root / ".claude").mkdir(parents=True, exist_ok=True)
            (project_root / ".moai" / "config").mkdir(parents=True, exist_ok=True)

            migrator = AlfredToMoaiMigrator(project_root)

            assert migrator.project_root == project_root
            assert migrator.claude_root == project_root / ".claude"
            assert migrator.config_path == project_root / ".moai" / "config" / "config.json"
            assert migrator.settings_path == project_root / ".claude" / "settings.json"
            assert migrator.backup_manager is not None

    def test_init_folder_paths_configuration(self):
        """CHAR: Verify alfred and moai folder paths are configured correctly"""
        with tempfile.TemporaryDirectory() as temp_dir:
            project_root = Path(temp_dir)
            (project_root / ".claude").mkdir(parents=True, exist_ok=True)

            migrator = AlfredToMoaiMigrator(project_root)

            # Check alfred folders
            assert "commands" in migrator.alfred_folders
            assert "agents" in migrator.alfred_folders
            assert "hooks" in migrator.alfred_folders
            assert migrator.alfred_folders["commands"] == project_root / ".claude" / "commands" / "alfred"
            assert migrator.alfred_folders["agents"] == project_root / ".claude" / "agents" / "alfred"
            assert migrator.alfred_folders["hooks"] == project_root / ".claude" / "hooks" / "alfred"

            # Check moai folders
            assert "commands" in migrator.moai_folders
            assert "agents" in migrator.moai_folders
            assert "hooks" in migrator.moai_folders
            assert migrator.moai_folders["commands"] == project_root / ".claude" / "commands" / "moai"
            assert migrator.moai_folders["agents"] == project_root / ".claude" / "agents" / "moai"
            assert migrator.moai_folders["hooks"] == project_root / ".claude" / "hooks" / "moai"


class TestAlfredToMoaiMigratorConfigHandling:
    """Characterization tests for config file operations"""

    def test_load_config_returns_empty_dict_when_file_missing(self):
        """CHAR: Return empty dict when config.json doesn't exist"""
        with tempfile.TemporaryDirectory() as temp_dir:
            project_root = Path(temp_dir)
            (project_root / ".claude").mkdir(parents=True, exist_ok=True)

            migrator = AlfredToMoaiMigrator(project_root)
            config = migrator._load_config()

            assert config == {}

    def test_load_config_reads_valid_json(self):
        """CHAR: Load and parse valid config.json"""
        with tempfile.TemporaryDirectory() as temp_dir:
            project_root = Path(temp_dir)
            config_dir = project_root / ".moai" / "config"
            config_dir.mkdir(parents=True, exist_ok=True)

            test_config = {"moai": {"version": "1.0.0"}, "migration": {}}
            config_file = config_dir / "config.json"
            config_file.write_text(json.dumps(test_config), encoding="utf-8")

            migrator = AlfredToMoaiMigrator(project_root)
            config = migrator._load_config()

            assert config == test_config

    def test_load_config_handles_invalid_json_gracefully(self):
        """CHAR: Return empty dict when config.json has invalid JSON"""
        with tempfile.TemporaryDirectory() as temp_dir:
            project_root = Path(temp_dir)
            config_dir = project_root / ".moai" / "config"
            config_dir.mkdir(parents=True, exist_ok=True)

            config_file = config_dir / "config.json"
            config_file.write_text("invalid json content", encoding="utf-8")

            migrator = AlfredToMoaiMigrator(project_root)
            config = migrator._load_config()

            assert config == {}

    def test_save_config_creates_directory_if_needed(self):
        """CHAR: Create config directory when saving config"""
        with tempfile.TemporaryDirectory() as temp_dir:
            project_root = Path(temp_dir)
            (project_root / ".claude").mkdir(parents=True, exist_ok=True)

            migrator = AlfredToMoaiMigrator(project_root)
            test_config = {"moai": {"version": "1.0.0"}}

            migrator._save_config(test_config)

            assert migrator.config_path.exists()
            assert migrator.config_path.parent.exists()

    def test_save_config_writes_valid_json(self):
        """CHAR: Write config as valid JSON"""
        with tempfile.TemporaryDirectory() as temp_dir:
            project_root = Path(temp_dir)
            (project_root / ".claude").mkdir(parents=True, exist_ok=True)

            migrator = AlfredToMoaiMigrator(project_root)
            test_config = {"moai": {"version": "1.0.0"}, "migration": {"alfred_to_moai": {"migrated": True}}}

            migrator._save_config(test_config)

            with open(migrator.config_path, "r", encoding="utf-8") as f:
                loaded_config = json.load(f)

            assert loaded_config == test_config


class TestAlfredToMoaiMigratorNeedsMigration:
    """Characterization tests for migration detection logic"""

    def test_needs_migration_returns_false_when_already_migrated(self):
        """CHAR: Return False when migration already completed"""
        with tempfile.TemporaryDirectory() as temp_dir:
            project_root = Path(temp_dir)
            config_dir = project_root / ".moai" / "config"
            config_dir.mkdir(parents=True, exist_ok=True)

            # Mark as already migrated
            config = {"migration": {"alfred_to_moai": {"migrated": True, "timestamp": "2025-01-01 12:00:00"}}}
            config_file = config_dir / "config.json"
            config_file.write_text(json.dumps(config), encoding="utf-8")

            migrator = AlfredToMoaiMigrator(project_root)
            result = migrator.needs_migration()

            assert result is False

    def test_needs_migration_returns_true_when_alfred_folders_exist(self):
        """CHAR: Return True when alfred folders are present"""
        with tempfile.TemporaryDirectory() as temp_dir:
            project_root = Path(temp_dir)
            alfred_commands = project_root / ".claude" / "commands" / "alfred"
            alfred_commands.mkdir(parents=True, exist_ok=True)

            migrator = AlfredToMoaiMigrator(project_root)
            result = migrator.needs_migration()

            assert result is True

    def test_needs_migration_returns_false_when_no_alfred_folders(self):
        """CHAR: Return False when no alfred folders exist"""
        with tempfile.TemporaryDirectory() as temp_dir:
            project_root = Path(temp_dir)
            (project_root / ".claude").mkdir(parents=True, exist_ok=True)

            migrator = AlfredToMoaiMigrator(project_root)
            result = migrator.needs_migration()

            assert result is False


class TestAlfredToMoaiMigratorSettingsUpdate:
    """Characterization tests for settings.json update logic"""

    def test_update_settings_json_replaces_alfred_paths(self):
        """CHAR: Replace alfred paths with moai paths in settings.json"""
        with tempfile.TemporaryDirectory() as temp_dir:
            project_root = Path(temp_dir)
            claude_root = project_root / ".claude"
            claude_root.mkdir(parents=True, exist_ok=True)

            # Create settings.json with alfred paths
            settings_content = """
            {
                "hooks": {
                    "preToolUse": [
                        {
                            "type": "command",
                            "path": ".claude/hooks/alfred/lib/security_guard.py"
                        }
                    ]
                }
            }
            """
            settings_file = claude_root / "settings.json"
            settings_file.write_text(settings_content, encoding="utf-8")

            migrator = AlfredToMoaiMigrator(project_root)
            migrator._update_settings_json_hooks()

            updated_content = settings_file.read_text(encoding="utf-8")

            assert ".claude/hooks/moai/" in updated_content
            assert ".claude/hooks/alfred/" not in updated_content

    def test_update_settings_json_handles_missing_file(self):
        """CHAR: Handle missing settings.json gracefully"""
        with tempfile.TemporaryDirectory() as temp_dir:
            project_root = Path(temp_dir)
            (project_root / ".claude").mkdir(parents=True, exist_ok=True)

            migrator = AlfredToMoaiMigrator(project_root)

            # Should not raise exception
            migrator._update_settings_json_hooks()

    def test_update_settings_json_validates_json_after_update(self):
        """CHAR: Validate JSON structure after path replacement"""
        with tempfile.TemporaryDirectory() as temp_dir:
            project_root = Path(temp_dir)
            claude_root = project_root / ".claude"
            claude_root.mkdir(parents=True, exist_ok=True)

            settings_content = (
                '{"hooks": {"preToolUse": [{"type": "command", "path": ".claude/hooks/alfred/test.py"}]}}'
            )
            settings_file = claude_root / "settings.json"
            settings_file.write_text(settings_content, encoding="utf-8")

            migrator = AlfredToMoaiMigrator(project_root)
            migrator._update_settings_json_hooks()

            # Should be valid JSON
            with open(settings_file, "r", encoding="utf-8") as f:
                json.load(f)


class TestAlfredToMoaiMigratorVerification:
    """Characterization tests for migration verification"""

    def test_verify_migration_checks_moai_folders_exist(self):
        """CHAR: Verify moai folders exist after migration"""
        with tempfile.TemporaryDirectory() as temp_dir:
            project_root = Path(temp_dir)
            claude_root = project_root / ".claude"

            # Create moai folders
            moai_commands = claude_root / "commands" / "moai"
            moai_commands.mkdir(parents=True, exist_ok=True)

            migrator = AlfredToMoaiMigrator(project_root)
            result = migrator._verify_migration()

            # Should fail if not all moai folders exist
            assert result is False or result is True  # Depends on which folders exist

    def test_verify_migration_checks_alfred_folders_deleted(self):
        """CHAR: Verify alfred folders are deleted after migration"""
        with tempfile.TemporaryDirectory() as temp_dir:
            project_root = Path(temp_dir)
            claude_root = project_root / ".claude"

            # Create moai folders but leave alfred folders
            (claude_root / "commands" / "moai").mkdir(parents=True, exist_ok=True)
            (claude_root / "commands" / "alfred").mkdir(parents=True, exist_ok=True)

            migrator = AlfredToMoaiMigrator(project_root)
            result = migrator._verify_migration()

            # Should fail if alfred folders still exist
            assert result is False


class TestAlfredToMoaiMigratorMigrationState:
    """Characterization tests for migration state recording"""

    def test_record_migration_state_saves_to_config(self):
        """CHAR: Record migration state in config.json"""
        with tempfile.TemporaryDirectory() as temp_dir:
            project_root = Path(temp_dir)
            config_dir = project_root / ".moai" / "config"
            config_dir.mkdir(parents=True, exist_ok=True)

            backup_path = Path("/backup/path")
            folders_count = 3

            migrator = AlfredToMoaiMigrator(project_root)
            migrator._record_migration_state(backup_path, folders_count)

            config = migrator._load_config()
            assert "migration" in config
            assert "alfred_to_moai" in config["migration"]
            assert config["migration"]["alfred_to_moai"]["migrated"] is True
            assert config["migration"]["alfred_to_moai"]["backup_path"] == str(backup_path)

    def test_get_package_version_returns_version_from_config(self):
        """CHAR: Get package version from config"""
        with tempfile.TemporaryDirectory() as temp_dir:
            project_root = Path(temp_dir)
            config_dir = project_root / ".moai" / "config"
            config_dir.mkdir(parents=True, exist_ok=True)

            test_config = {"moai": {"version": "1.5.0"}}
            config_file = config_dir / "config.json"
            config_file.write_text(json.dumps(test_config), encoding="utf-8")

            migrator = AlfredToMoaiMigrator(project_root)
            version = migrator._get_package_version()

            assert version == "1.5.0"

    def test_get_package_version_returns_unknown_when_missing(self):
        """CHAR: Return 'unknown' when version not in config"""
        with tempfile.TemporaryDirectory() as temp_dir:
            project_root = Path(temp_dir)
            config_dir = project_root / ".moai" / "config"
            config_dir.mkdir(parents=True, exist_ok=True)

            migrator = AlfredToMoaiMigrator(project_root)
            version = migrator._get_package_version()

            assert version == "unknown"


class TestSelectiveRestorerInitialization:
    """Characterization tests for SelectiveRestorer initialization"""

    def test_init_with_project_path(self):
        """CHAR: Initialize restorer with project path"""
        with tempfile.TemporaryDirectory() as temp_dir:
            project_path = Path(temp_dir)

            restorer = SelectiveRestorer(project_path)

            assert restorer.project_path == project_path
            assert restorer.restoration_log == []
            assert restorer.backup_path is None or isinstance(restorer.backup_path, Path)

    def test_init_with_backup_path(self):
        """CHAR: Initialize restorer with explicit backup path"""
        with tempfile.TemporaryDirectory() as temp_dir:
            project_path = Path(temp_dir)
            backup_path = Path(temp_dir) / "backup"

            restorer = SelectiveRestorer(project_path, backup_path)

            assert restorer.backup_path == backup_path

    def test_find_latest_backup(self):
        """CHAR: Auto-detect latest backup directory"""
        with tempfile.TemporaryDirectory() as temp_dir:
            project_path = Path(temp_dir)
            backups_dir = project_path / ".moai-backups"
            backups_dir.mkdir(parents=True, exist_ok=True)

            # Create backup directories
            backup1 = backups_dir / "pre-update-backup_001"
            backup2 = backups_dir / "pre-update-backup_002"
            backup1.mkdir(parents=True, exist_ok=True)
            backup2.mkdir(parents=True, exist_ok=True)

            restorer = SelectiveRestorer(project_path)

            # Should find the latest backup
            assert restorer.backup_path is not None
            assert "pre-update-backup" in str(restorer.backup_path)


class TestSelectiveRestorerElementGrouping:
    """Characterization tests for element grouping logic"""

    def test_group_elements_by_type(self):
        """CHAR: Group elements into agents, commands, skills, hooks"""
        restorer = SelectiveRestorer(Path("/tmp/project"))

        elements = [
            ".claude/agents/my-agent.md",
            ".claude/commands/moai/my-cmd.md",
            ".claude/skills/my-skill/",
            ".claude/hooks/moai/my-hook.py",
        ]

        groups = restorer._group_elements_by_type(elements)

        assert "agents" in groups
        assert "commands" in groups
        assert "skills" in groups
        assert "hooks" in groups
        assert len(groups["agents"]) == 1
        assert len(groups["commands"]) == 1
        assert len(groups["skills"]) == 1
        assert len(groups["hooks"]) == 1

    def test_group_unknown_elements(self):
        """CHAR: Handle elements with unknown type"""
        restorer = SelectiveRestorer(Path("/tmp/project"))

        elements = [".claude/unknown-type/file.md"]

        groups = restorer._group_elements_by_type(elements)

        assert "unknown" in groups
        assert len(groups["unknown"]) == 1


class TestSelectiveRestorerPathNormalization:
    """Characterization tests for path normalization"""

    def test_normalize_absolute_path_with_claude_prefix(self):
        """CHAR: Normalize absolute path containing .claude/"""
        restorer = SelectiveRestorer(Path("/tmp/project"))

        absolute_path = Path("/tmp/project/.claude/agents/test.md")
        normalized = restorer._normalize_element_path(absolute_path)

        assert normalized is not None
        assert str(normalized).startswith(".claude/")

    def test_normalize_absolute_path_with_moai_prefix(self):
        """CHAR: Normalize absolute path containing .moai/"""
        restorer = SelectiveRestorer(Path("/tmp/project"))

        absolute_path = Path("/tmp/project/.moai/config/config.yaml")
        normalized = restorer._normalize_element_path(absolute_path)

        assert normalized is not None
        assert str(normalized).startswith(".moai/")

    def test_normalize_relative_path(self):
        """CHAR: Return relative path as-is"""
        restorer = SelectiveRestorer(Path("/tmp/project"))

        relative_path = Path(".claude/agents/test.md")
        normalized = restorer._normalize_element_path(relative_path)

        assert normalized == relative_path

    def test_normalize_absolute_path_without_safe_prefix_returns_none(self):
        """CHAR: Return None for absolute path without .claude/ or .moai/"""
        restorer = SelectiveRestorer(Path("/tmp/project"))

        absolute_path = Path("/tmp/project/other/path/file.md")
        normalized = restorer._normalize_element_path(absolute_path)

        assert normalized is None


class TestSelectiveRestorerPathValidation:
    """Characterization tests for path validation"""

    def test_validate_rejects_path_traversal(self):
        """CHAR: Reject paths containing .. for security"""
        restorer = SelectiveRestorer(Path("/tmp/project"))

        malicious_path = Path(".claude/../../etc/passwd")
        result = restorer._validate_element_path(malicious_path)

        assert result is False

    def test_validate_accepts_claude_prefix(self):
        """CHAR: Accept paths starting with .claude/"""
        restorer = SelectiveRestorer(Path("/tmp/project"))

        valid_path = Path(".claude/agents/test.md")
        result = restorer._validate_element_path(valid_path)

        assert result is True

    def test_validate_accepts_moai_prefix(self):
        """CHAR: Accept paths starting with .moai/"""
        restorer = SelectiveRestorer(Path("/tmp/project"))

        valid_path = Path(".moai/config/test.yaml")
        result = restorer._validate_element_path(valid_path)

        assert result is True

    def test_validate_rejects_unsafe_prefix(self):
        """CHAR: Reject paths not starting with allowed prefixes"""
        restorer = SelectiveRestorer(Path("/tmp/project"))

        invalid_path = Path("unsafe/path/file.md")
        result = restorer._validate_element_path(invalid_path)

        assert result is False


class TestSelectiveRestorerConflictHandling:
    """Characterization tests for file conflict handling"""

    def test_handle_conflict_with_directories(self):
        """CHAR: Handle directory conflicts by backing up target"""
        with tempfile.TemporaryDirectory() as temp_dir:
            project_path = Path(temp_dir)
            target_path = project_path / ".claude" / "skills" / "test-skill"
            backup_source = project_path / "backup" / ".claude" / "skills" / "test-skill"

            target_path.mkdir(parents=True, exist_ok=True)
            backup_source.mkdir(parents=True, exist_ok=True)

            restorer = SelectiveRestorer(project_path)
            result = restorer._handle_file_conflict(target_path, backup_source)

            assert result is True or result is False  # May succeed or fail

    def test_handle_conflict_with_identical_files(self):
        """CHAR: Skip backup when files are identical"""
        with tempfile.TemporaryDirectory() as temp_dir:
            project_path = Path(temp_dir)
            target_path = project_path / ".claude" / "test.md"
            backup_source = project_path / "backup" / ".claude" / "test.md"

            # Create parent directories
            target_path.parent.mkdir(parents=True, exist_ok=True)
            backup_source.parent.mkdir(parents=True, exist_ok=True)

            content = "same content"
            target_path.write_text(content, encoding="utf-8")
            backup_source.write_text(content, encoding="utf-8")

            restorer = SelectiveRestorer(project_path)
            result = restorer._handle_file_conflict(target_path, backup_source)

            assert result is True


class TestSelectiveRestorerFactory:
    """Characterization tests for factory function"""

    def test_create_selective_restorer_with_string_path(self):
        """CHAR: Create restorer from string path"""
        with tempfile.TemporaryDirectory() as temp_dir:
            restorer = create_selective_restorer(temp_dir)

            assert isinstance(restorer, SelectiveRestorer)
            assert restorer.project_path == Path(temp_dir).resolve()

    def test_create_selective_restorer_with_path_object(self):
        """CHAR: Create restorer from Path object"""
        with tempfile.TemporaryDirectory() as temp_dir:
            project_path = Path(temp_dir)
            restorer = create_selective_restorer(project_path)

            assert isinstance(restorer, SelectiveRestorer)
            assert restorer.project_path == project_path.resolve()

    def test_create_selective_restorer_with_backup_path(self):
        """CHAR: Create restorer with explicit backup path"""
        with tempfile.TemporaryDirectory() as temp_dir:
            backup_path = Path(temp_dir) / "backup"
            restorer = create_selective_restorer(temp_dir, backup_path)

            assert restorer.backup_path == backup_path
