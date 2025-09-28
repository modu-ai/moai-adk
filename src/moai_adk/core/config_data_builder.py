"""
Configuration data structure builder for MoAI projects.

@TASK:CONFIG-DATA-BUILDER-001 Extracted from config_project.py for TRUST compliance
@DESIGN:SEPARATED-BUILDER-001 Single responsibility: MoAI config data generation
"""

from datetime import datetime
from typing import Any

from .._version import get_version
from ..config import Config
from ..utils.logger import get_logger

logger = get_logger(__name__)


class ConfigDataBuilder:
    """Builds MoAI configuration data structures"""

    def __init__(self, mode: str = "personal"):
        """
        Initialize configuration data builder.

        Args:
            mode: Project mode ("personal" or "team")
        """
        if mode not in ["personal", "team"]:
            raise ValueError(f"Invalid mode: {mode}. Must be 'personal' or 'team'")

        self.mode = mode

    def build_moai_config(self, config: Config, project_dir: str = None) -> dict[str, Any]:
        """
        Build complete MoAI configuration data structure.

        Args:
            config: Project configuration object
            project_dir: Project directory path

        Returns:
            Complete MoAI configuration dictionary
        """
        base_config = self._build_base_config(config, project_dir)
        git_strategy = self._build_git_strategy()
        workflows = self._build_workflows_config()
        runtime_config = self._build_runtime_config(config)

        # Merge all configurations
        moai_config = {**base_config, **git_strategy, **workflows}

        if runtime_config:
            moai_config["runtime"] = runtime_config

        return moai_config

    def _build_base_config(self, config: Config, project_dir: str = None) -> dict[str, Any]:
        """Build base configuration section"""
        return {
            "version": "0.1.18",
            "created": datetime.now().isoformat(),
            "constitution_version": "2.1",
            "project": {
                "name": getattr(config, "name", "project"),
                "path": str(getattr(config, "project_path", project_dir or ".")),
                "template": getattr(config, "template", "standard"),
                "mode": self.mode,
            },
            "created_at": datetime.now().isoformat(),
            "moai_adk_version": get_version(),
        }

    def _build_git_strategy(self) -> dict[str, Any]:
        """Build git strategy configuration"""
        return {
            "git_strategy": {
                "personal": {
                    "auto_commit": True,
                    "auto_pr": False,
                    "develop_branch": "main",
                    "feature_prefix": "feature/",
                    "use_gitflow": False,
                },
                "team": {
                    "auto_commit": False,
                    "auto_pr": True,
                    "develop_branch": "develop",
                    "feature_prefix": "feature/SPEC-",
                    "use_gitflow": True,
                },
            }
        }

    def _build_workflows_config(self) -> dict[str, Any]:
        """Build workflows configuration"""
        return {
            "workflows": {
                "moai:0-project": {
                    "enabled": True,
                    "description": "프로젝트 초기화",
                    "dependencies": [],
                },
                "moai:1-spec": {
                    "enabled": True,
                    "description": "SPEC 명세 작성",
                    "dependencies": ["moai:0-project"],
                },
                "moai:2-build": {
                    "enabled": True,
                    "description": "TDD 구현",
                    "dependencies": ["moai:1-spec"],
                },
                "moai:3-sync": {
                    "enabled": True,
                    "description": "문서 동기화",
                    "dependencies": ["moai:2-build"],
                },
            }
        }

    def _build_runtime_config(self, config: Config) -> dict[str, Any] | None:
        """Build runtime configuration if available"""
        if not hasattr(config, "runtime") or not config.runtime:
            return None

        return {
            "language": getattr(config.runtime, "name", "python"),
            "version": getattr(config.runtime, "version", "latest"),
        }