"""
Version detection module for MoAI-ADK projects

Detects the current version of a project and determines
which migrations are needed.
"""

import json
import logging
from pathlib import Path
from typing import Any, Dict

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
        self.old_config = self.project_root / ".moai" / "config.json"
        self.new_config = self.project_root / ".moai" / "config" / "config.json"
        self.old_statusline = self.project_root / ".claude" / "statusline-config.yaml"
        self.new_statusline = self.project_root / ".moai" / "config" / "statusline-config.yaml"

    def detect_version(self) -> str:
        """
        Detect current project version based on file structure

        Returns:
            Version string (e.g., "0.23.0", "0.24.0+", "unknown")
        """
        # Check if already migrated to v0.24.0+
        if self.new_config.exists():
            return "0.24.0+"

        # Check if v0.23.0 or earlier
        if self.old_config.exists():
            try:
                with open(self.old_config, "r") as f:
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

        if version == "0.24.0+":
            logger.info("Project already on v0.24.0+")
            return False

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
                "cleanup": ["old_file1", "old_file2"]
            }
        """
        plan: Dict[str, Any] = {"move": [], "create": [], "cleanup": []}

        if not self.needs_migration():
            return plan

        # Create config directory
        plan["create"].append(str(self.project_root / ".moai" / "config"))

        # Move config.json
        if self.old_config.exists() and not self.new_config.exists():
            plan["move"].append(
                {
                    "from": str(self.old_config.relative_to(self.project_root)),
                    "to": str(self.new_config.relative_to(self.project_root)),
                    "description": "Main configuration file",
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
            "has_new_config": self.new_config.exists(),
            "has_old_statusline": self.old_statusline.exists(),
            "has_new_statusline": self.new_statusline.exists(),
        }
