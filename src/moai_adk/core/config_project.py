"""
Project configuration management (MoAI, package.json, etc.).

@FEATURE:CONFIG-PROJECT-001 Project-specific configuration management
@DESIGN:SEPARATED-PROJECT-001 Extracted from oversized config_manager.py (564 LOC)
"""

import json
from datetime import datetime
from pathlib import Path
from typing import Any

from .._version import get_version
from ..config import Config
from ..utils.logger import get_logger

logger = get_logger(__name__)


class ProjectConfigManager:
    """프로젝트 설정 관리자"""

    def __init__(self, project_dir: Path = None):
        """프로젝트 설정 관리자 초기화"""
        self.project_dir = project_dir or Path.cwd()
        self._mode = "personal"  # 기본값
        self._options = {}  # 옵션 저장소

    def create_moai_config(self, config_path: Path, config: Config) -> bool:
        """
        @TASK:CONFIG-MOAI-001 Create .moai/config.json file

        Args:
            config_path: Full path to config.json file
            config: Project configuration

        Returns:
            bool: True if config file was created successfully
        """
        if config_path.exists() and not getattr(config, 'force_overwrite', False):
            # keep existing config
            return True

        try:
            config_data = self._create_moai_config_data(config)

            # Create directory if needed
            config_path.parent.mkdir(parents=True, exist_ok=True)

            # Write config file
            with open(config_path, 'w', encoding='utf-8') as f:
                json.dump(config_data, f, indent=2)

            logger.info(f"Created MoAI config at {config_path}")
            return True

        except Exception as e:
            logger.error(f"Failed to create MoAI config: {e}")
            return False

    def _create_moai_config_data(self, config: Config) -> dict[str, Any]:
        """Create MoAI configuration data structure"""
        moai_config = {
            "version": "0.1.9",
            "created": datetime.now().isoformat(),
            "constitution_version": "2.1",
            "project": {
                "name": getattr(config, 'name', 'project'),
                "path": str(getattr(config, 'project_path', self.project_dir)),
                "template": getattr(config, 'template', 'standard'),
                "mode": self._mode
            },
            "git_strategy": {
                "personal": {
                    "auto_commit": True,
                    "auto_pr": False,
                    "develop_branch": "main",
                    "feature_prefix": "feature/",
                    "use_gitflow": False
                },
                "team": {
                    "auto_commit": False,
                    "auto_pr": True,
                    "develop_branch": "develop",
                    "feature_prefix": "feature/SPEC-",
                    "use_gitflow": True
                }
            },
            "workflows": {
                "moai:0-project": {
                    "enabled": True,
                    "description": "프로젝트 초기화",
                    "dependencies": []
                },
                "moai:1-spec": {
                    "enabled": True,
                    "description": "SPEC 명세 작성",
                    "dependencies": ["moai:0-project"]
                },
                "moai:2-build": {
                    "enabled": True,
                    "description": "TDD 구현",
                    "dependencies": ["moai:1-spec"]
                },
                "moai:3-sync": {
                    "enabled": True,
                    "description": "문서 동기화",
                    "dependencies": ["moai:2-build"]
                }
            },
            "created_at": datetime.now().isoformat(),
            "moai_adk_version": get_version()
        }

        # Runtime-specific configuration
        if hasattr(config, 'runtime') and config.runtime:
            runtime_name = getattr(config.runtime, 'name', 'python')
            moai_config["runtime"] = {
                "language": runtime_name,
                "version": getattr(config.runtime, 'version', 'latest')
            }

        return moai_config

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
        """Create package.json if applicable"""
        # Check if this is a Node.js project
        if hasattr(config, 'runtime') and config.runtime:
            runtime_name = getattr(config.runtime, 'name', 'python')
            if runtime_name not in ['javascript', 'typescript', 'node']:
                return None

        package_path = project_path / "package.json"

        if package_path.exists():
            logger.info(f"package.json already exists at {package_path}")
            return package_path

        try:
            package_data = {
                "name": getattr(config, 'name', 'moai-project'),
                "version": "0.1.0",
                "description": f"MoAI-ADK project: {getattr(config, 'name', 'project')}",
                "private": True,
                "scripts": {
                    "test": "jest",
                    "build": "npm run compile",
                    "dev": "npm run watch",
                    "lint": "eslint .",
                    "format": "prettier --write ."
                },
                "devDependencies": {
                    "jest": "^29.0.0",
                    "eslint": "^8.0.0",
                    "prettier": "^3.0.0"
                },
                "keywords": ["moai-adk", "development"],
                "author": "MoAI-ADK",
                "license": "MIT"
            }

            # TypeScript-specific dependencies
            if runtime_name == 'typescript':
                package_data["devDependencies"].update({
                    "typescript": "^5.0.0",
                    "@types/node": "^20.0.0",
                    "@types/jest": "^29.0.0"
                })
                package_data["scripts"]["compile"] = "tsc"

            with open(package_path, 'w', encoding='utf-8') as f:
                json.dump(package_data, f, indent=2)

            logger.info(f"Created package.json at {package_path}")
            return package_path

        except Exception as e:
            logger.error(f"Failed to create package.json: {e}")
            return None

    def create_initial_indexes(self, project_path: Path, config: Config) -> list[Path]:
        """Create initial index files for TAG system"""
        created_files = []

        try:
            indexes_dir = project_path / ".moai" / "indexes"
            indexes_dir.mkdir(parents=True, exist_ok=True)

            # tags.json
            tags_index_path = indexes_dir / "tags.json"
            if not tags_index_path.exists():
                initial_tags_data = {
                    "version": "1.0.0",
                    "updated": datetime.now().isoformat(),
                    "statistics": {
                        "total_tags": 0,
                        "categories": {
                            "Primary": 0,
                            "Steering": 0,
                            "Implementation": 0,
                            "Quality": 0
                        }
                    },
                    "index": {},
                    "references": {}
                }

                with open(tags_index_path, 'w', encoding='utf-8') as f:
                    json.dump(initial_tags_data, f, indent=2)

                created_files.append(tags_index_path)
                logger.info(f"Created tags index at {tags_index_path}")

            # reports/sync-report.md
            reports_dir = project_path / ".moai" / "reports"
            reports_dir.mkdir(parents=True, exist_ok=True)

            sync_report_path = reports_dir / "sync-report.md"
            if not sync_report_path.exists():
                initial_report = f"""# MoAI-ADK Sync Report

## Project Information
- **Project Name**: {getattr(config, 'name', 'project')}
- **Created**: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}
- **MoAI Version**: {get_version()}

## Sync Status
- **Last Sync**: {datetime.now().isoformat()}
- **TAG Count**: 0
- **Document Status**: Ready

## Next Steps
1. Run `/moai:1-spec` to create your first specification
2. Use `/moai:2-build` for TDD implementation
3. Execute `/moai:3-sync` to update this report
"""

                with open(sync_report_path, 'w', encoding='utf-8') as f:
                    f.write(initial_report)

                created_files.append(sync_report_path)
                logger.info(f"Created sync report at {sync_report_path}")

        except Exception as e:
            logger.error(f"Failed to create initial indexes: {e}")

        return created_files

    def setup_steering_config(self, project_path: Path) -> Path:
        """Setup steering configuration for project governance"""
        steering_dir = project_path / ".moai" / "steering"
        steering_dir.mkdir(parents=True, exist_ok=True)

        config_path = steering_dir / "governance.json"

        if config_path.exists():
            logger.info(f"Steering config already exists at {config_path}")
            return config_path

        try:
            steering_config = {
                "version": "1.0",
                "governance": {
                    "decision_making": "consensus",
                    "review_required": ["SPEC", "ADR"],
                    "approval_threshold": 1
                },
                "policies": {
                    "trust_principles": {
                        "test_first": True,
                        "readable_code": True,
                        "unified_architecture": True,
                        "secured_development": True,
                        "trackable_changes": True
                    },
                    "code_standards": {
                        "max_function_lines": 50,
                        "max_file_lines": 300,
                        "max_parameters": 5,
                        "max_complexity": 10
                    }
                },
                "workflows": {
                    "spec_first": True,
                    "tdd_required": True,
                    "documentation_sync": True
                },
                "created": datetime.now().isoformat(),
                "moai_version": get_version()
            }

            with open(config_path, 'w', encoding='utf-8') as f:
                json.dump(steering_config, f, indent=2)

            logger.info(f"Created steering config at {config_path}")
            return config_path

        except Exception as e:
            logger.error(f"Failed to create steering config: {e}")
            return None

    def set_mode(self, mode: str):
        """
        Set project mode.

        Args:
            mode: "personal" or "team"
        """
        if mode not in ["personal", "team"]:
            raise ValueError(f"Invalid mode: {mode}. Must be 'personal' or 'team'")

        self._mode = mode
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
