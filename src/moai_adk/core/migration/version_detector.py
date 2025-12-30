"""
Version detection module for MoAI-ADK projects

Detects the current version of a project and determines
which migrations are needed.
"""

import json
import logging
from pathlib import Path
from typing import Any, Dict, Optional

import yaml

logger = logging.getLogger(__name__)


class VersionDetector:
    """Detects project version and migration requirements"""

    def __init__(self, project_root: Path):
        """
        Initialize version detector

        Args:
            project_root: Root directory of the project
        """
        self.project_root = Path(project_root)
        # Legacy config paths (pre-v0.24.0)
        self.old_config = self.project_root / ".moai" / "config.json"
        # Intermediate config path (v0.24.0)
        self.intermediate_config = self.project_root / ".moai" / "config" / "config.json"
        # New section YAML paths (v0.36.0+)
        self.sections_dir = self.project_root / ".moai" / "config" / "sections"
        self.system_yaml = self.sections_dir / "system.yaml"
        # Statusline config paths
        self.old_statusline = self.project_root / ".claude" / "statusline-config.yaml"
        self.new_statusline = self.project_root / ".moai" / "config" / "statusline-config.yaml"

    def _read_yaml_file(self, file_path: Path) -> Optional[Dict[str, Any]]:
        """
        Read a YAML file safely

        Args:
            file_path: Path to the YAML file

        Returns:
            Dictionary with file contents, or None if read failed
        """
        try:
            if file_path.exists():
                with open(file_path, "r", encoding="utf-8") as f:
                    return yaml.safe_load(f)
        except Exception as e:
            logger.warning(f"Failed to read YAML file {file_path}: {e}")
        return None

    def _get_version_from_section_yaml(self) -> Optional[str]:
        """
        Get version from section YAML files (new format)

        Returns:
            Version string from system.yaml, or None if not found
        """
        system_data = self._read_yaml_file(self.system_yaml)
        if system_data and "moai" in system_data:
            version = system_data["moai"].get("version")
            if version:
                return str(version)
        return None

    def detect_version(self) -> str:
        """
        Detect current project version based on file structure

        Returns:
            Version string (e.g., "0.23.0", "0.24.0+", "0.36.0+", "unknown")
        """
        # Check if using new section YAML format (v0.36.0+)
        if self.sections_dir.exists() and self.system_yaml.exists():
            version = self._get_version_from_section_yaml()
            if version:
                return version
            return "0.36.0+"

        # Check if already migrated to v0.24.0+ (intermediate config.json)
        if self.intermediate_config.exists():
            try:
                with open(self.intermediate_config, "r", encoding="utf-8") as f:
                    config_data = json.load(f)
                    if "moai_version" in config_data:
                        return config_data["moai_version"]
                    return "0.24.0+"
            except Exception as e:
                logger.warning(f"Failed to read intermediate config: {e}")
                return "0.24.0+"

        # Check if v0.23.0 or earlier (old config.json at .moai/config.json)
        if self.old_config.exists():
            try:
                with open(self.old_config, "r", encoding="utf-8") as f:
                    config_data = json.load(f)
                    # Try to get version from config
                    if "moai_version" in config_data:
                        return config_data["moai_version"]
                    return "0.23.0"
            except Exception as e:
                logger.warning(f"Failed to read old config: {e}")
                return "0.23.0"

        return "unknown"

    def needs_migration(self) -> bool:
        """
        Check if project needs migration

        Returns:
            True if migration is needed, False otherwise
        """
        version = self.detect_version()
        if version == "unknown":
            logger.info("Unknown version, assuming no migration needed")
            return False

        # Check if using new section YAML format (v0.36.0+)
        if self.sections_dir.exists() and self.system_yaml.exists():
            logger.info("Project already using section YAML format (v0.36.0+)")
            return False

        if version == "0.24.0+":
            logger.info("Project on v0.24.0+, may need migration to section YAML format")
            # Could still need migration to section YAML format
            return not self.sections_dir.exists()

        # Version 0.23.0 or earlier needs migration
        return True

    def get_migration_plan(self) -> Dict[str, Any]:
        """
        Get detailed migration plan

        Returns:
            Dictionary with migration actions:
            {
                "move": [{"from": "...", "to": "..."}],
                "create": ["directory1", "directory2"],
                "cleanup": ["old_file1", "old_file2"],
                "convert": [{"from": "...", "to_dir": "...", "description": "..."}]
            }
        """
        plan: Dict[str, Any] = {"move": [], "create": [], "cleanup": [], "convert": []}

        if not self.needs_migration():
            return plan

        # Create config directory structure
        config_dir = self.project_root / ".moai" / "config"
        plan["create"].append(str(config_dir))

        # Create sections directory for new YAML format
        if not self.sections_dir.exists():
            plan["create"].append(str(self.sections_dir))

        # Handle old config.json migration (pre-v0.24.0)
        if self.old_config.exists() and not self.intermediate_config.exists():
            plan["move"].append(
                {
                    "from": str(self.old_config.relative_to(self.project_root)),
                    "to": str(self.intermediate_config.relative_to(self.project_root)),
                    "description": "Main configuration file (legacy)",
                }
            )

        # Mark config.json files for conversion to section YAML
        config_json_to_convert = None
        if self.intermediate_config.exists():
            config_json_to_convert = self.intermediate_config
        elif self.old_config.exists():
            config_json_to_convert = self.old_config

        if config_json_to_convert and not self.system_yaml.exists():
            plan["convert"].append(
                {
                    "from": str(config_json_to_convert.relative_to(self.project_root)),
                    "to_dir": str(self.sections_dir.relative_to(self.project_root)),
                    "description": "Convert config.json to section YAML files",
                }
            )

        # Move statusline-config.yaml
        if self.old_statusline.exists() and not self.new_statusline.exists():
            plan["move"].append(
                {
                    "from": str(self.old_statusline.relative_to(self.project_root)),
                    "to": str(self.new_statusline.relative_to(self.project_root)),
                    "description": "Statusline configuration",
                }
            )

        # Cleanup old files (after successful migration)
        if self.old_config.exists():
            plan["cleanup"].append(str(self.old_config.relative_to(self.project_root)))
        if self.old_statusline.exists():
            plan["cleanup"].append(str(self.old_statusline.relative_to(self.project_root)))
        # Also cleanup intermediate config.json after conversion to YAML
        if self.intermediate_config.exists() and self.sections_dir.exists():
            plan["cleanup"].append(str(self.intermediate_config.relative_to(self.project_root)))

        return plan

    def get_version_info(self) -> Dict[str, Any]:
        """
        Get detailed version information

        Returns:
            Dictionary with version details
        """
        return {
            "detected_version": self.detect_version(),
            "needs_migration": self.needs_migration(),
            "has_old_config": self.old_config.exists(),
            "has_intermediate_config": self.intermediate_config.exists(),
            "has_sections_dir": self.sections_dir.exists(),
            "has_system_yaml": self.system_yaml.exists(),
            "has_old_statusline": self.old_statusline.exists(),
            "has_new_statusline": self.new_statusline.exists(),
            "config_format": self._get_config_format(),
        }

    def _get_config_format(self) -> str:
        """
        Determine the current configuration format

        Returns:
            One of: "section_yaml", "intermediate_json", "legacy_json", "none"
        """
        if self.sections_dir.exists() and self.system_yaml.exists():
            return "section_yaml"
        elif self.intermediate_config.exists():
            return "intermediate_json"
        elif self.old_config.exists():
            return "legacy_json"
        return "none"
