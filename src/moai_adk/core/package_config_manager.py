"""
Package configuration manager for language-specific project files.

@TASK:PACKAGE-CONFIG-001 Extracted from config_project.py for TRUST compliance
@DESIGN:SEPARATED-PACKAGE-001 Single responsibility: package.json and similar files
"""

import json
from pathlib import Path
from typing import Any

from ..config import Config
from ..utils.logger import get_logger

logger = get_logger(__name__)


class PackageConfigManager:
    """Manages language-specific package configuration files"""

    def create_package_json(self, project_path: Path, config: Config) -> Path | None:
        """
        Create package.json for Node.js/TypeScript projects.

        Args:
            project_path: Target project directory
            config: Project configuration

        Returns:
            Path to created package.json or None if not applicable
        """
        if not self._is_nodejs_project(config):
            logger.debug("Skipping package.json creation (not a Node.js project)")
            return None

        package_path = project_path / "package.json"

        if package_path.exists():
            logger.info(f"package.json already exists at {package_path}")
            return package_path

        try:
            runtime_name = self._get_runtime_name(config)
            package_data = self._build_package_data(config, runtime_name)

            with open(package_path, "w", encoding="utf-8") as f:
                json.dump(package_data, f, indent=2)

            logger.info(f"Created package.json at {package_path}")
            return package_path

        except Exception as e:
            logger.error(f"Failed to create package.json: {e}")
            return None

    def _is_nodejs_project(self, config: Config) -> bool:
        """Check if this is a Node.js/TypeScript project"""
        if not hasattr(config, "runtime") or not config.runtime:
            return False

        runtime_name = getattr(config.runtime, "name", "python")
        return runtime_name in ["javascript", "typescript", "node"]

    def _get_runtime_name(self, config: Config) -> str:
        """Get runtime name from config"""
        if not hasattr(config, "runtime") or not config.runtime:
            return "javascript"
        return getattr(config.runtime, "name", "javascript")

    def _build_package_data(self, config: Config, runtime_name: str) -> dict[str, Any]:
        """Build package.json data structure"""
        base_data = self._build_base_package_data(config)

        if runtime_name == "typescript":
            self._add_typescript_dependencies(base_data)

        return base_data

    def _build_base_package_data(self, config: Config) -> dict[str, Any]:
        """Build base package.json structure"""
        return {
            "name": getattr(config, "name", "moai-project"),
            "version": "0.1.0",
            "description": f"MoAI-ADK project: {getattr(config, 'name', 'project')}",
            "private": True,
            "scripts": {
                "test": "jest",
                "build": "npm run compile",
                "dev": "npm run watch",
                "lint": "eslint .",
                "format": "prettier --write .",
            },
            "devDependencies": {
                "jest": "^29.0.0",
                "eslint": "^8.0.0",
                "prettier": "^3.0.0",
            },
            "keywords": ["moai-adk", "development"],
            "author": "MoAI-ADK",
            "license": "MIT",
        }

    def _add_typescript_dependencies(self, package_data: dict[str, Any]) -> None:
        """Add TypeScript-specific dependencies and scripts"""
        package_data["devDependencies"].update(
            {
                "typescript": "^5.0.0",
                "@types/node": "^20.0.0",
                "@types/jest": "^29.0.0",
            }
        )
        package_data["scripts"]["compile"] = "tsc"