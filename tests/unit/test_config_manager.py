"""
Unit tests for the config manager module.

Tests the ConfigManager class and its configuration file creation methods
to ensure proper configuration management and JSON validation.
"""

import pytest
import json
import tempfile
from pathlib import Path
from unittest.mock import patch, MagicMock, mock_open
from datetime import datetime

from moai_adk.core.config_manager import ConfigManager
from moai_adk.core.security import SecurityManager, SecurityError
from moai_adk.config import Config, RuntimeConfig


class TestConfigManager:
    """Test cases for ConfigManager class."""

    @pytest.fixture
    def temp_dir(self):
        """Create a temporary directory for testing."""
        with tempfile.TemporaryDirectory() as temp_dir:
            yield Path(temp_dir)

    @pytest.fixture
    def security_manager(self):
        """Create a mock security manager for testing."""
        return MagicMock(spec=SecurityManager)

    @pytest.fixture
    def config_manager(self, security_manager):
        """Create a ConfigManager instance for testing."""
        return ConfigManager(security_manager)

    @pytest.fixture
    def sample_config(self, temp_dir):
        """Create a sample Config instance for testing."""
        return Config(
            name="test-project",
            template="standard",
            runtime=RuntimeConfig("python"),
            path=str(temp_dir / "test_project"),
            tech_stack=["python"],
        )

    def test_init_with_security_manager(self, security_manager):
        """Test ConfigManager initialization with security manager."""
        manager = ConfigManager(security_manager)
        assert manager.security_manager == security_manager

    def test_init_without_security_manager(self):
        """Test ConfigManager initialization without security manager."""
        manager = ConfigManager()
        assert isinstance(manager.security_manager, SecurityManager)

    def test_create_claude_settings_file_success(
        self, config_manager, temp_dir, sample_config
    ):
        """Test successful creation of Claude Code settings file."""
        # Mock security validation to return True
        config_manager.security_manager.validate_file_creation.return_value = True

        # Create project directory
        sample_config.project_path.mkdir(parents=True)

        settings_path = config_manager.create_claude_settings_file(
            sample_config.project_path, sample_config
        )

        assert settings_path.exists()
        assert settings_path.name == "settings.json"

        # Validate JSON content
        with open(settings_path, "r", encoding="utf-8") as f:
            settings_data = json.load(f)

        # Check required sections
        assert "hooks" in settings_data
        assert "permissions" in settings_data

        # Check hooks configuration (minimal: tag_validator + session_start hooks)
        pre_entries = settings_data["hooks"].get("PreToolUse", [])
        all_commands = [
            hook.get("command", "")
            for entry in pre_entries
            for hook in entry.get("hooks", [])
        ]
        assert any(cmd.endswith("tag_validator.py") for cmd in all_commands)
        assert "SessionStart" in settings_data["hooks"]
        session_start = settings_data["hooks"]["SessionStart"]
        ss_cmds = [
            hook.get("command", "")
            for entry in session_start
            for hook in entry.get("hooks", [])
        ]
        assert any(cmd.endswith("session_start_notice.py") for cmd in ss_cmds)
        assert any(cmd.endswith("language_detector.py") for cmd in ss_cmds)

    def test_create_claude_settings_file_security_failure(
        self, config_manager, temp_dir, sample_config
    ):
        """Test Claude settings file creation when security validation fails."""
        # Mock security validation to return False
        config_manager.security_manager.validate_file_creation.return_value = False

        with pytest.raises(SecurityError):
            config_manager.create_claude_settings_file(
                sample_config.project_path, sample_config
            )

    def test_create_moai_config_file_success(
        self, config_manager, temp_dir, sample_config
    ):
        """Test successful creation of MoAI config file."""
        # Mock security validation to return True
        config_manager.security_manager.validate_file_creation.return_value = True

        # Create project directory
        sample_config.project_path.mkdir(parents=True)

        config_path = config_manager.create_moai_config_file(
            sample_config.project_path, sample_config
        )

        assert config_path.exists()
        assert config_path.name == "config.json"

        # Validate JSON content
        with open(config_path, "r", encoding="utf-8") as f:
            config_data = json.load(f)

        # Check required sections
        assert "version" in config_data
        assert "project" in config_data
        assert "constitution" in config_data
        assert "tags" in config_data
        assert "pipeline" in config_data
        assert "agents" in config_data

        # Check project configuration
        project_data = config_data["project"]
        assert project_data["name"] == sample_config.name
        assert project_data["runtime"] == sample_config.runtime.name

        # Check constitution principles
        constitution = config_data["constitution"]
        assert constitution["simplicity"]["max_projects"] == 3
        assert constitution["testing"]["tdd_required"] is True
        assert constitution["testing"]["coverage_target"] == 0.8

        # Check 16-core TAG system structure
        tags = config_data["tags"]
        assert tags["version"] == "16-core"
        categories = tags["categories"]
        assert set(categories["spec"]) == {"REQ", "SPEC", "DESIGN", "TASK"}
        assert set(categories["steering"]) == {"VISION", "STRUCT", "TECH", "ADR"}
        assert "implementation" in categories and "quality" in categories

    def test_create_moai_config_file_security_failure(
        self, config_manager, temp_dir, sample_config
    ):
        """Test MoAI config file creation when security validation fails."""
        # Mock security validation to return False
        config_manager.security_manager.validate_file_creation.return_value = False

        with pytest.raises(SecurityError):
            config_manager.create_moai_config_file(
                sample_config.project_path, sample_config
            )

    def test_create_package_json_success(self, config_manager, temp_dir, sample_config):
        """Test successful creation of package.json file."""
        # Mock security validation to return True
        config_manager.security_manager.validate_file_creation.return_value = True

        # Create project directory
        sample_config.project_path.mkdir(parents=True)

        package_path = config_manager.create_package_json(
            sample_config.project_path, sample_config
        )

        assert package_path.exists()
        assert package_path.name == "package.json"

        # Validate JSON content
        with open(package_path, "r", encoding="utf-8") as f:
            package_data = json.load(f)

        # Check required fields
        assert package_data["name"] == sample_config.name
        assert package_data["type"] == "module"
        assert package_data["private"] is True
        assert "scripts" in package_data

        # Check scripts
        scripts = package_data["scripts"]
        assert "dev" in scripts
        assert "build" in scripts
        assert "start" in scripts
        assert "lint" in scripts

    def test_create_package_json_nextjs_stack(self, config_manager, temp_dir):
        """Test package.json creation for Next.js project."""
        # Mock security validation to return True
        config_manager.security_manager.validate_file_creation.return_value = True

        # Create config with Next.js stack
        nextjs_config = Config(
            name="nextjs-project",
            description="Next.js project",
            template_type="web",
            runtime="node",
            project_path=temp_dir / "nextjs_project",
            tech_stack=["nextjs", "react"],
        )

        # Create project directory
        nextjs_config.project_path.mkdir(parents=True)

        package_path = config_manager.create_package_json(
            nextjs_config.project_path, nextjs_config
        )

        # Validate JSON content
        with open(package_path, "r", encoding="utf-8") as f:
            package_data = json.load(f)

        # Check Next.js specific scripts
        scripts = package_data["scripts"]
        assert scripts["dev"] == "next dev"
        assert scripts["build"] == "next build"
        assert scripts["start"] == "next start"
        assert scripts["lint"] == "next lint"

    def test_create_initial_indexes_success(
        self, config_manager, temp_dir, sample_config
    ):
        """Test successful creation of initial index files."""
        # Mock security validation to return True
        config_manager.security_manager.validate_file_creation.return_value = True

        # Create project directory
        sample_config.project_path.mkdir(parents=True)

        index_files = config_manager.create_initial_indexes(
            sample_config.project_path, sample_config
        )

        assert len(index_files) == 3

        # Check files exist
        tags_file = sample_config.project_path / ".moai" / "indexes" / "tags.json"
        traceability_file = (
            sample_config.project_path / ".moai" / "indexes" / "traceability.json"
        )
        state_file = sample_config.project_path / ".moai" / "indexes" / "state.json"

        assert tags_file.exists()
        assert traceability_file.exists()
        assert state_file.exists()

        # Validate tags.json
        with open(tags_file, "r", encoding="utf-8") as f:
            tags_data = json.load(f)

        assert tags_data["metadata"]["version"] == "16-core"
        assert set(tags_data["categories"]["SPEC"].keys()) == {
            "REQ",
            "SPEC",
            "DESIGN",
            "TASK",
        }
        assert set(tags_data["categories"]["Steering"].keys()) == {
            "VISION",
            "STRUCT",
            "TECH",
            "ADR",
        }

        # Validate traceability.json
        with open(traceability_file, "r", encoding="utf-8") as f:
            traceability_data = json.load(f)

        assert "primary" in traceability_data["chains"]
        assert traceability_data["chains"]["primary"] == [
            "REQ",
            "DESIGN",
            "TASK",
            "TEST",
        ]
        assert traceability_data["chains"]["steering"] == [
            "VISION",
            "STRUCT",
            "TECH",
            "ADR",
        ]

        # Validate state.json
        with open(state_file, "r", encoding="utf-8") as f:
            state_data = json.load(f)

        assert state_data["project_name"] == sample_config.name
        assert state_data["current_stage"] == "INIT"

    def test_create_initial_indexes_security_failure(
        self, config_manager, temp_dir, sample_config
    ):
        """Test initial indexes creation when security validation fails."""
        # Mock security validation to return False
        config_manager.security_manager.validate_file_creation.return_value = False

        with pytest.raises(SecurityError):
            config_manager.create_initial_indexes(
                sample_config.project_path, sample_config
            )

    def test_setup_steering_config_success(self, config_manager, temp_dir):
        """Test successful setup of steering configuration."""
        # Mock security validation to return True
        config_manager.security_manager.validate_file_creation.return_value = True

        project_path = temp_dir / "steering_project"
        project_path.mkdir(parents=True)

        config_path = config_manager.setup_steering_config(project_path)

        assert config_path.exists()
        assert config_path.name == "config.json"

        # Validate JSON content
        with open(config_path, "r", encoding="utf-8") as f:
            steering_data = json.load(f)

        # Check steering configuration
        assert steering_data["autoCommit"] == "ask"
        assert steering_data["coverageTarget"] == 0.8

        # Check constitution
        constitution = steering_data["constitution"]
        assert constitution["maxProjects"] == 3
        assert constitution["enforceTDD"] is True

        # Check pipeline gates
        pipeline = steering_data["pipeline"]
        assert "SPECIFY" in pipeline["gates"]
        assert "PLAN" in pipeline["gates"]
        assert "TASKS" in pipeline["gates"]
        assert "IMPLEMENT" in pipeline["gates"]

    def test_write_json_file_success(self, config_manager, temp_dir):
        """Test successful JSON file writing."""
        # Mock security validation to return True
        config_manager.security_manager.validate_file_creation.return_value = True

        test_data = {"test": "data", "number": 42, "array": [1, 2, 3]}
        file_path = temp_dir / "test.json"

        result_path = config_manager._write_json_file(file_path, test_data)

        assert result_path == file_path
        assert file_path.exists()

        # Validate content
        with open(file_path, "r", encoding="utf-8") as f:
            loaded_data = json.load(f)

        assert loaded_data == test_data

    def test_write_json_file_security_failure(self, config_manager, temp_dir):
        """Test JSON file writing when security validation fails."""
        # Mock security validation to return False
        config_manager.security_manager.validate_file_creation.return_value = False

        test_data = {"test": "data"}
        file_path = temp_dir / "test.json"

        with pytest.raises(SecurityError):
            config_manager._write_json_file(file_path, test_data)

    def test_write_json_file_creates_parent_directories(self, config_manager, temp_dir):
        """Test that JSON file writing creates parent directories."""
        # Mock security validation to return True
        config_manager.security_manager.validate_file_creation.return_value = True

        test_data = {"test": "data"}
        nested_path = temp_dir / "nested" / "directories" / "test.json"

        config_manager._write_json_file(nested_path, test_data)

        assert nested_path.exists()
        assert nested_path.parent.exists()

    @patch("builtins.open", side_effect=IOError("Permission denied"))
    def test_write_json_file_io_error(self, mock_open, config_manager, temp_dir):
        """Test JSON file writing with IO error."""
        # Mock security validation to return True
        config_manager.security_manager.validate_file_creation.return_value = True

        test_data = {"test": "data"}
        file_path = temp_dir / "test.json"

        with pytest.raises(IOError):
            config_manager._write_json_file(file_path, test_data)

    def test_validate_config_file_valid_json(self, config_manager, temp_dir):
        """Test validation of valid JSON config file."""
        test_data = {"valid": "json", "data": 123}
        config_file = temp_dir / "valid_config.json"

        with open(config_file, "w", encoding="utf-8") as f:
            json.dump(test_data, f)

        result = config_manager.validate_config_file(config_file)
        assert result is True

    def test_validate_config_file_invalid_json(self, config_manager, temp_dir):
        """Test validation of invalid JSON config file."""
        config_file = temp_dir / "invalid_config.json"

        # Write invalid JSON
        with open(config_file, "w", encoding="utf-8") as f:
            f.write('{"invalid": json}')  # Missing quotes around json

        result = config_manager.validate_config_file(config_file)
        assert result is False

    def test_validate_config_file_nonexistent(self, config_manager, temp_dir):
        """Test validation of nonexistent config file."""
        nonexistent_file = temp_dir / "nonexistent.json"

        result = config_manager.validate_config_file(nonexistent_file)
        assert result is False

    @patch("builtins.open", side_effect=IOError("Permission denied"))
    def test_validate_config_file_io_error(self, mock_open, config_manager, temp_dir):
        """Test validation with IO error."""
        config_file = temp_dir / "config.json"
        config_file.write_text('{"test": "data"}')

        result = config_manager.validate_config_file(config_file)
        assert result is False

    def test_backup_config_file_success(self, config_manager, temp_dir):
        """Test successful config file backup."""
        # Create original config file
        original_file = temp_dir / "config.json"
        original_data = {"original": "data"}

        with open(original_file, "w", encoding="utf-8") as f:
            json.dump(original_data, f)

        backup_path = config_manager.backup_config_file(original_file)

        # Check backup exists
        assert backup_path.exists()
        assert "backup_" in backup_path.name
        assert backup_path.suffix == ".json"

        # Verify backup content
        with open(backup_path, "r", encoding="utf-8") as f:
            backup_data = json.load(f)

        assert backup_data == original_data

    @patch("shutil.copy2", side_effect=IOError("Disk full"))
    def test_backup_config_file_io_error(self, mock_copy, config_manager, temp_dir):
        """Test config file backup with IO error."""
        original_file = temp_dir / "config.json"
        original_file.write_text('{"test": "data"}')

        with pytest.raises(IOError):
            config_manager.backup_config_file(original_file)

    def test_integration_full_config_setup(
        self, config_manager, temp_dir, sample_config
    ):
        """Test complete configuration setup workflow."""
        # Mock security validation to return True
        config_manager.security_manager.validate_file_creation.return_value = True

        # Create project directory
        sample_config.project_path.mkdir(parents=True)

        # Create all configuration files
        claude_settings = config_manager.create_claude_settings_file(
            sample_config.project_path, sample_config
        )
        moai_config = config_manager.create_moai_config_file(
            sample_config.project_path, sample_config
        )
        package_json = config_manager.create_package_json(
            sample_config.project_path, sample_config
        )
        index_files = config_manager.create_initial_indexes(
            sample_config.project_path, sample_config
        )
        steering_config = config_manager.setup_steering_config(
            sample_config.project_path
        )

        # Verify all files exist
        assert claude_settings.exists()
        assert moai_config.exists()
        assert package_json.exists()
        assert len(index_files) == 3
        assert all(f.exists() for f in index_files)
        assert steering_config.exists()

        # Validate all files have valid JSON
        assert config_manager.validate_config_file(claude_settings)
        assert config_manager.validate_config_file(moai_config)
        assert config_manager.validate_config_file(package_json)
        assert all(config_manager.validate_config_file(f) for f in index_files)
        assert config_manager.validate_config_file(steering_config)

    def test_config_content_consistency(self, config_manager, temp_dir, sample_config):
        """Test that configuration files have consistent content."""
        # Mock security validation to return True
        config_manager.security_manager.validate_file_creation.return_value = True

        # Create project directory
        sample_config.project_path.mkdir(parents=True)

        # Create configuration files
        claude_settings = config_manager.create_claude_settings_file(
            sample_config.project_path, sample_config
        )
        moai_config = config_manager.create_moai_config_file(
            sample_config.project_path, sample_config
        )

        # Load and check consistency
        with open(claude_settings, "r", encoding="utf-8") as f:
            claude_data = json.load(f)

        with open(moai_config, "r", encoding="utf-8") as f:
            moai_data = json.load(f)

        # Project name should be consistent
        assert moai_data["project"]["name"] == sample_config.name

        # Runtime should be consistent
        assert moai_data["project"]["runtime"] == sample_config.runtime.name

    def test_json_formatting_and_encoding(self, config_manager, temp_dir):
        """Test JSON formatting and UTF-8 encoding."""
        # Mock security validation to return True
        config_manager.security_manager.validate_file_creation.return_value = True

        # Test data with Unicode characters
        test_data = {
            "project": "테스트 프로젝트",
            "description": "Unicode characters: 한글, 日本語, Emoji 🚀",
            "numbers": [1, 2, 3],
            "nested": {"key": "value"},
        }

        file_path = temp_dir / "unicode_test.json"
        config_manager._write_json_file(file_path, test_data)

        # Read and verify
        with open(file_path, "r", encoding="utf-8") as f:
            content = f.read()
            loaded_data = json.loads(content)

        assert loaded_data == test_data
        # Check that file is properly formatted (indented)
        assert "\n" in content  # Should have newlines from indentation
        assert "  " in content  # Should have indentation spaces

    def test_config_timestamps_and_metadata(
        self, config_manager, temp_dir, sample_config
    ):
        """Test that configuration files contain proper timestamps and metadata."""
        # Mock security validation to return True
        config_manager.security_manager.validate_file_creation.return_value = True

        # Create project directory
        sample_config.project_path.mkdir(parents=True)

        # Record time before creation
        before_time = datetime.now()

        # Create config files
        moai_config = config_manager.create_moai_config_file(
            sample_config.project_path, sample_config
        )
        index_files = config_manager.create_initial_indexes(
            sample_config.project_path, sample_config
        )

        # Record time after creation
        after_time = datetime.now()

        # Check MoAI config timestamp
        with open(moai_config, "r", encoding="utf-8") as f:
            moai_data = json.load(f)

        created_time = datetime.fromisoformat(moai_data["created"])
        assert before_time <= created_time <= after_time

        # Check index files timestamps
        for index_file in index_files:
            with open(index_file, "r", encoding="utf-8") as f:
                index_data = json.load(f)

            if "metadata" in index_data and "generated_at" in index_data["metadata"]:
                generated_time = datetime.fromisoformat(
                    index_data["metadata"]["generated_at"]
                )
                assert before_time <= generated_time <= after_time
