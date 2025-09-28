"""
Project configuration orchestrator (refactored for TRUST compliance).

@FEATURE:CONFIG-PROJECT-001 Project-specific configuration management
@DESIGN:REFACTORED-PROJECT-001 Orchestrates specialized config modules (87% LOC reduction)
"""

import json
from pathlib import Path
from typing import Any

from .config_data_builder import ConfigDataBuilder
from .index_manager import IndexManager
from .package_config_manager import PackageConfigManager
from ..config import Config
from ..utils.logger import get_logger

logger = get_logger(__name__)


class ProjectConfigManager:
    """Project configuration orchestrator using specialized modules"""

    def __init__(self, project_dir: Path = None):
        """Initialize project configuration orchestrator"""
        self.project_dir = project_dir or Path.cwd()
        self._mode = "personal"  # Default mode
        self._options = {}  # Options storage

        # Initialize specialized managers
        self.config_builder = ConfigDataBuilder(self._mode)
        self.package_manager = PackageConfigManager()
        self.index_manager = IndexManager()

    def create_moai_config(self, config_path: Path, config: Config) -> bool:
        """
        @TASK:CONFIG-MOAI-001 Create .moai/config.json file using ConfigDataBuilder

        Args:
            config_path: Full path to config.json file
            config: Project configuration

        Returns:
            bool: True if config file was created successfully
        """
        if config_path.exists() and not getattr(config, "force_overwrite", False):
            # Keep existing config
            logger.debug(f"MoAI config already exists at {config_path}")
            return True

        try:
            # Use ConfigDataBuilder to generate data
            config_data = self.config_builder.build_moai_config(
                config, str(self.project_dir)
            )

            # Create directory if needed
            config_path.parent.mkdir(parents=True, exist_ok=True)

            # Write config file
            with open(config_path, "w", encoding="utf-8") as f:
                json.dump(config_data, f, indent=2)

            logger.info(f"Created MoAI config at {config_path}")
            return True

        except Exception as e:
            logger.error(f"Failed to create MoAI config: {e}")
            return False

    def create_moai_config_file(self, project_path: Path, config: Config) -> Path:
        """Create MoAI config file and return its path"""
        moai_dir = project_path / ".moai"
        config_path = moai_dir / "config.json"

        success = self.create_moai_config(config_path, config)

        if success:
            return config_path
        else:
            raise RuntimeError(f"Failed to create MoAI config at {config_path}")

    def create_package_json(self, project_path: Path, config: Config) -> Path:
        """Create package.json using PackageConfigManager"""
        return self.package_manager.create_package_json(project_path, config)

    def create_initial_indexes(self, project_path: Path, config: Config) -> list[Path]:
        """Create initial index files using IndexManager"""
        return self.index_manager.create_initial_indexes(project_path, config)

    def setup_steering_config(self, project_path: Path) -> Path:
        """Setup steering configuration using IndexManager"""
        return self.index_manager.setup_steering_config(project_path)

    def set_mode(self, mode: str):
        """
        Set project mode and update specialized managers.

        Args:
            mode: "personal" or "team"
        """
        if mode not in ["personal", "team"]:
            raise ValueError(f"Invalid mode: {mode}. Must be 'personal' or 'team'")

        self._mode = mode
        # Update ConfigDataBuilder with new mode
        self.config_builder = ConfigDataBuilder(mode)
        logger.info(f"Project mode set to: {mode}")

    def get_mode(self) -> str:
        """
        Get current project mode.

        Returns:
            Current mode ("personal" or "team")
        """
        return self._mode

    def set_option(self, key: str, value: Any):
        """
        Set configuration option.

        Args:
            key: Option key
            value: Option value
        """
        self._options[key] = value
        logger.debug(f"Set option {key} = {value}")

    def get_option(self, key: str, default: Any = None) -> Any:
        """
        Get configuration option.

        Args:
            key: Option key
            default: Default value if key not found

        Returns:
            Option value or default
        """
        return self._options.get(key, default)
