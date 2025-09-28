"""
Tests for refactored config project modules.

@TEST:CONFIG-MODULES-001 Test suite for TRUST-compliant config modules
@DESIGN:MODULAR-TEST-001 Tests for ConfigDataBuilder, PackageConfigManager, IndexManager
"""

import json
import sqlite3
import tempfile
from datetime import datetime
from pathlib import Path
from unittest.mock import Mock, patch

import pytest

from src.moai_adk.config import Config
from src.moai_adk.core.config_data_builder import ConfigDataBuilder
from src.moai_adk.core.config_project import ProjectConfigManager
from src.moai_adk.core.index_manager import IndexManager
from src.moai_adk.core.package_config_manager import PackageConfigManager


class TestConfigDataBuilder:
    """Test ConfigDataBuilder functionality"""

    def test_init_with_valid_mode(self):
        """Test initialization with valid mode"""
        builder = ConfigDataBuilder("personal")
        assert builder.mode == "personal"

        builder = ConfigDataBuilder("team")
        assert builder.mode == "team"

    def test_init_with_invalid_mode(self):
        """Test initialization with invalid mode raises error"""
        with pytest.raises(ValueError, match="Invalid mode"):
            ConfigDataBuilder("invalid")

    def test_build_moai_config_personal_mode(self):
        """Test building MoAI config for personal mode"""
        builder = ConfigDataBuilder("personal")
        config = Mock()
        config.name = "test-project"
        config.template = "standard"

        result = builder.build_moai_config(config, "/test/path")

        assert result["project"]["mode"] == "personal"
        assert result["project"]["name"] == "test-project"
        assert result["git_strategy"]["personal"]["auto_commit"] is True
        assert result["git_strategy"]["personal"]["auto_pr"] is False

    def test_build_moai_config_team_mode(self):
        """Test building MoAI config for team mode"""
        builder = ConfigDataBuilder("team")
        config = Mock()
        config.name = "team-project"

        result = builder.build_moai_config(config)

        assert result["project"]["mode"] == "team"
        assert result["git_strategy"]["team"]["auto_commit"] is False
        assert result["git_strategy"]["team"]["auto_pr"] is True
        assert result["git_strategy"]["team"]["use_gitflow"] is True

    def test_build_moai_config_with_runtime(self):
        """Test building MoAI config with runtime information"""
        builder = ConfigDataBuilder("personal")
        config = Mock()
        config.name = "runtime-project"
        config.runtime = Mock()
        config.runtime.name = "typescript"
        config.runtime.version = "5.0.0"

        result = builder.build_moai_config(config)

        assert "runtime" in result
        assert result["runtime"]["language"] == "typescript"
        assert result["runtime"]["version"] == "5.0.0"

    def test_build_workflows_config(self):
        """Test workflows configuration structure"""
        builder = ConfigDataBuilder("personal")
        config = Mock()
        config.name = "workflow-test"

        result = builder.build_moai_config(config)

        workflows = result["workflows"]
        assert "moai:0-project" in workflows
        assert "moai:1-spec" in workflows
        assert "moai:2-build" in workflows
        assert "moai:3-sync" in workflows
        assert workflows["moai:1-spec"]["dependencies"] == ["moai:0-project"]


class TestPackageConfigManager:
    """Test PackageConfigManager functionality"""

    def test_create_package_json_javascript_project(self):
        """Test creating package.json for JavaScript project"""
        with tempfile.TemporaryDirectory() as tmp_dir:
            project_path = Path(tmp_dir)
            manager = PackageConfigManager()

            config = Mock()
            config.name = "js-project"
            config.runtime = Mock()
            config.runtime.name = "javascript"

            result = manager.create_package_json(project_path, config)

            assert result is not None
            assert result.exists()

            with open(result) as f:
                package_data = json.load(f)

            assert package_data["name"] == "js-project"
            assert package_data["scripts"]["test"] == "jest"
            assert "jest" in package_data["devDependencies"]

    def test_create_package_json_typescript_project(self):
        """Test creating package.json for TypeScript project"""
        with tempfile.TemporaryDirectory() as tmp_dir:
            project_path = Path(tmp_dir)
            manager = PackageConfigManager()

            config = Mock()
            config.name = "ts-project"
            config.runtime = Mock()
            config.runtime.name = "typescript"

            result = manager.create_package_json(project_path, config)

            assert result is not None
            assert result.exists()

            with open(result) as f:
                package_data = json.load(f)

            assert "typescript" in package_data["devDependencies"]
            assert "@types/node" in package_data["devDependencies"]
            assert package_data["scripts"]["compile"] == "tsc"

    def test_create_package_json_non_nodejs_project(self):
        """Test package.json creation skipped for non-Node.js projects"""
        with tempfile.TemporaryDirectory() as tmp_dir:
            project_path = Path(tmp_dir)
            manager = PackageConfigManager()

            config = Mock()
            config.name = "python-project"
            config.runtime = Mock()
            config.runtime.name = "python"

            result = manager.create_package_json(project_path, config)

            assert result is None

    def test_create_package_json_existing_file(self):
        """Test behavior when package.json already exists"""
        with tempfile.TemporaryDirectory() as tmp_dir:
            project_path = Path(tmp_dir)
            package_path = project_path / "package.json"

            # Create existing package.json
            with open(package_path, "w") as f:
                json.dump({"name": "existing"}, f)

            manager = PackageConfigManager()
            config = Mock()
            config.runtime = Mock()
            config.runtime.name = "javascript"

            result = manager.create_package_json(project_path, config)

            assert result == package_path
            # Verify file wasn't overwritten
            with open(package_path) as f:
                data = json.load(f)
            assert data["name"] == "existing"


class TestIndexManager:
    """Test IndexManager functionality"""

    def test_create_initial_indexes(self):
        """Test creating initial index files"""
        with tempfile.TemporaryDirectory() as tmp_dir:
            project_path = Path(tmp_dir)
            manager = IndexManager()

            config = Mock()
            config.name = "index-test"

            result = manager.create_initial_indexes(project_path, config)

            assert len(result) == 2  # tags.db and sync-report.md

            # Check tags database
            tags_db = project_path / ".moai" / "indexes" / "tags.db"
            assert tags_db.exists()

            # Verify database structure
            conn = sqlite3.connect(tags_db)
            cursor = conn.cursor()
            cursor.execute("SELECT name FROM sqlite_master WHERE type='table'")
            tables = [row[0] for row in cursor.fetchall()]
            assert "tags" in tables
            assert "statistics" in tables
            assert "tag_references" in tables
            conn.close()

            # Check sync report
            sync_report = project_path / ".moai" / "reports" / "sync-report.md"
            assert sync_report.exists()

            content = sync_report.read_text()
            assert "index-test" in content
            assert "MoAI-ADK Sync Report" in content

    def test_setup_steering_config(self):
        """Test setting up steering configuration"""
        with tempfile.TemporaryDirectory() as tmp_dir:
            project_path = Path(tmp_dir)
            manager = IndexManager()

            result = manager.setup_steering_config(project_path)

            assert result is not None
            assert result.exists()

            with open(result) as f:
                config_data = json.load(f)

            assert config_data["version"] == "1.0"
            assert "governance" in config_data
            assert "policies" in config_data
            assert config_data["policies"]["trust_principles"]["test_first"] is True

    def test_setup_steering_config_existing_file(self):
        """Test steering config setup when file already exists"""
        with tempfile.TemporaryDirectory() as tmp_dir:
            project_path = Path(tmp_dir)
            steering_dir = project_path / ".moai" / "steering"
            steering_dir.mkdir(parents=True)

            config_path = steering_dir / "governance.json"
            with open(config_path, "w") as f:
                json.dump({"existing": True}, f)

            manager = IndexManager()
            result = manager.setup_steering_config(project_path)

            assert result == config_path
            # Verify file wasn't overwritten
            with open(config_path) as f:
                data = json.load(f)
            assert data["existing"] is True


class TestProjectConfigManager:
    """Test refactored ProjectConfigManager orchestration"""

    def test_init_creates_specialized_managers(self):
        """Test initialization creates all specialized managers"""
        manager = ProjectConfigManager()

        assert hasattr(manager, "config_builder")
        assert hasattr(manager, "package_manager")
        assert hasattr(manager, "index_manager")
        assert isinstance(manager.config_builder, ConfigDataBuilder)
        assert isinstance(manager.package_manager, PackageConfigManager)
        assert isinstance(manager.index_manager, IndexManager)

    def test_create_moai_config_delegates_to_builder(self):
        """Test MoAI config creation delegates to ConfigDataBuilder"""
        with tempfile.TemporaryDirectory() as tmp_dir:
            config_path = Path(tmp_dir) / "config.json"
            manager = ProjectConfigManager()

            # Create a proper config object instead of Mock
            config = type('Config', (), {})()
            config.name = "delegation-test"
            config.template = "standard"

            result = manager.create_moai_config(config_path, config)

            assert result is True
            assert config_path.exists()

            with open(config_path) as f:
                config_data = json.load(f)

            assert config_data["project"]["name"] == "delegation-test"
            assert config_data["project"]["mode"] == "personal"

    def test_create_package_json_delegates_to_package_manager(self):
        """Test package.json creation delegates to PackageConfigManager"""
        with tempfile.TemporaryDirectory() as tmp_dir:
            project_path = Path(tmp_dir)
            manager = ProjectConfigManager()

            config = Mock()
            config.name = "package-delegation"
            config.runtime = Mock()
            config.runtime.name = "javascript"

            result = manager.create_package_json(project_path, config)

            assert result is not None
            assert result.exists()

    def test_create_initial_indexes_delegates_to_index_manager(self):
        """Test index creation delegates to IndexManager"""
        with tempfile.TemporaryDirectory() as tmp_dir:
            project_path = Path(tmp_dir)
            manager = ProjectConfigManager()

            # Create a proper config object instead of Mock
            config = type('Config', (), {})()
            config.name = "index-delegation"

            result = manager.create_initial_indexes(project_path, config)

            assert len(result) == 2
            assert all(path.exists() for path in result)

    def test_set_mode_updates_config_builder(self):
        """Test mode change updates ConfigDataBuilder"""
        manager = ProjectConfigManager()

        # Initial mode
        assert manager.get_mode() == "personal"
        assert manager.config_builder.mode == "personal"

        # Change mode
        manager.set_mode("team")
        assert manager.get_mode() == "team"
        assert manager.config_builder.mode == "team"

    def test_set_mode_invalid_raises_error(self):
        """Test invalid mode raises ValueError"""
        manager = ProjectConfigManager()

        with pytest.raises(ValueError, match="Invalid mode"):
            manager.set_mode("invalid")

    def test_options_management(self):
        """Test option setting and getting"""
        manager = ProjectConfigManager()

        # Test setting and getting options
        manager.set_option("test_key", "test_value")
        assert manager.get_option("test_key") == "test_value"

        # Test default value
        assert manager.get_option("nonexistent", "default") == "default"
        assert manager.get_option("nonexistent") is None