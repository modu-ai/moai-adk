# @REQ:CONFIG-SYSTEM-011
"""
@FEATURE:CONFIG-001 Configuration management utilities for MoAI-ADK

Handles creation and management of various configuration files including
Claude Code settings, MoAI config, package.json, and other project configs.
@DESIGN:REFACTORED-CONFIG-001 Rebuilt from 564 LOC to modular architecture (< 200 LOC)
"""

from pathlib import Path
from typing import Any

from ..config import Config
from ..utils.logger import get_logger
from .config_claude import ClaudeConfigManager
from .config_project import ProjectConfigManager
from .config_utils import ConfigUtils
from .security import SecurityManager

logger = get_logger(__name__)


class ConfigManager:
    """@TASK:CONFIG-MANAGER-001 Configuration management orchestrator for MoAI-ADK installation."""

    def __init__(self, project_dir: Path = None, security_manager: SecurityManager = None):
        """
        Initialize configuration manager orchestrator.

        Args:
            project_dir: Project directory (added for new functionality)
            security_manager: Security manager instance for validation
        """
        self.project_dir = project_dir or Path.cwd()
        self.security_manager = security_manager or SecurityManager()

        # Initialize specialized modules
        self.claude_config = ClaudeConfigManager(self.security_manager)
        self.project_config = ProjectConfigManager(project_dir)
        self.config_utils = ConfigUtils(self.security_manager)

    # Claude Config delegate methods
    def create_claude_settings(self, settings_path: Path, config: Config) -> bool:
        """@TASK:CONFIG-CLAUDE-001 Create Claude Code settings.json file"""
        return self.claude_config.create_claude_settings(settings_path, config)

    def create_claude_settings_file(self, project_path: Path, config: Config) -> Path:
        """Create Claude settings file and return its path"""
        return self.claude_config.create_claude_settings_file(project_path, config)

    # Project Config delegate methods
    def create_moai_config(self, config_path: Path, config: Config) -> bool:
        """@TASK:CONFIG-MOAI-001 Create .moai/config.json file"""
        return self.project_config.create_moai_config(config_path, config)

    def create_moai_config_file(self, project_path: Path, config: Config) -> Path:
        """Create MoAI config file and return its path"""
        return self.project_config.create_moai_config_file(project_path, config)

    def create_package_json(self, project_path: Path, config: Config) -> Path:
        """Create package.json if applicable"""
        return self.project_config.create_package_json(project_path, config)

    def create_initial_indexes(self, project_path: Path, config: Config) -> list[Path]:
        """Create initial index files for TAG system"""
        return self.project_config.create_initial_indexes(project_path, config)

    def setup_steering_config(self, project_path: Path) -> Path:
        """Setup steering configuration for project governance"""
        return self.project_config.setup_steering_config(project_path)

    def set_mode(self, mode: str):
        """Set project mode"""
        return self.project_config.set_mode(mode)

    def get_mode(self) -> str:
        """Get current project mode"""
        return self.project_config.get_mode()

    def set_option(self, key: str, value: Any):
        """Set configuration option"""
        return self.project_config.set_option(key, value)

    def get_option(self, key: str, default: Any = None) -> Any:
        """Get configuration option"""
        return self.project_config.get_option(key, default)

    # Utils delegate methods
    def _write_json_file(self, file_path: Path, data: dict[str, Any]) -> Path:
        """Write JSON data to file with error handling and logging"""
        return self.config_utils.write_json_file(file_path, data)

    def validate_config_file(self, file_path: Path) -> bool:
        """Validate configuration file format and content"""
        return self.config_utils.validate_config_file(file_path)

    def backup_config_file(self, file_path: Path) -> Path:
        """Create backup of configuration file"""
        return self.config_utils.backup_config_file(file_path)

    # Additional convenience methods
    def setup_full_project_config(self, project_path: Path, config: Config) -> dict[str, Path]:
        """Complete project configuration setup"""
        created_files = {}

        try:
            # Claude settings
            claude_settings = self.create_claude_settings_file(project_path, config)
            if claude_settings:
                created_files['claude_settings'] = claude_settings

            # MoAI config
            moai_config = self.create_moai_config_file(project_path, config)
            if moai_config:
                created_files['moai_config'] = moai_config

            # Package.json (if applicable)
            package_json = self.create_package_json(project_path, config)
            if package_json:
                created_files['package_json'] = package_json

            # Initial indexes
            indexes = self.create_initial_indexes(project_path, config)
            if indexes:
                created_files['indexes'] = indexes

            # Steering config
            steering_config = self.setup_steering_config(project_path)
            if steering_config:
                created_files['steering_config'] = steering_config

            logger.info(f"Full project configuration completed: {len(created_files)} components")
            return created_files

        except Exception as e:
            logger.error(f"Error during full project configuration: {e}")
            raise RuntimeError(f"Project configuration failed: {e}")

    def get_project_config_status(self, project_path: Path) -> dict[str, Any]:
        """Check project configuration status"""
        status = {}

        # Check Claude settings
        claude_settings_path = project_path / ".claude" / "settings.json"
        status['claude_settings'] = self.config_utils.get_config_summary(claude_settings_path)

        # Check MoAI config
        moai_config_path = project_path / ".moai" / "config.json"
        status['moai_config'] = self.config_utils.get_config_summary(moai_config_path)

        # Check package.json
        package_json_path = project_path / "package.json"
        status['package_json'] = self.config_utils.get_config_summary(package_json_path)

        # Check indexes
        tags_index_path = project_path / ".moai" / "indexes" / "tags.db"
        status['tags_index'] = self.config_utils.get_config_summary(tags_index_path)

        return status
