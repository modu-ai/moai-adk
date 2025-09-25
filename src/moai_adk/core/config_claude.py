"""
Claude Code configuration management.

@FEATURE:CONFIG-CLAUDE-001 Claude Code settings.json management
@DESIGN:SEPARATED-CLAUDE-001 Extracted from oversized config_manager.py (564 LOC)
"""

import json
from pathlib import Path
from typing import Dict, Any

from ..utils.logger import get_logger
from ..config import Config
from .security import SecurityManager

logger = get_logger(__name__)


class ClaudeConfigManager:
    """Claude Code 설정 관리자"""

    def __init__(self, security_manager: SecurityManager = None):
        """Claude 설정 관리자 초기화"""
        self.security_manager = security_manager or SecurityManager()

    def create_claude_settings(self, settings_path: Path, config: Config) -> bool:
        """
        @TASK:CONFIG-CLAUDE-001 Create Claude Code settings.json file

        Args:
            settings_path: Full path to settings.json file
            config: Project configuration

        Returns:
            bool: True if settings file was created successfully
        """
        if settings_path.exists() and not getattr(config, 'force_overwrite', False):
            # keep existing settings (from templates)
            return True

        try:
            settings_data = self._create_claude_settings_data(config)

            # Security validation
            if not self.security_manager.validate_claude_settings(settings_data):
                logger.error("Claude settings validation failed")
                return False

            # Create directory if needed
            settings_path.parent.mkdir(parents=True, exist_ok=True)

            # Write settings file
            with open(settings_path, 'w', encoding='utf-8') as f:
                json.dump(settings_data, f, indent=2)

            logger.info(f"Created Claude settings at {settings_path}")
            return True

        except Exception as e:
            logger.error(f"Failed to create Claude settings: {e}")
            return False

    def _create_claude_settings_data(self, config: Config) -> Dict[str, Any]:
        """Create Claude settings data structure"""
        settings = {
            "defaultMode": "acceptEdits",
            "permissions": {
                "defaultBehavior": "acceptEdits",
                "overrides": {
                    ".claude/hooks/moai/policy_block.py": "ask",
                    ".claude/hooks/moai/pre_write_guard.py": "ask",
                    ".claude/hooks/moai/session_start.py": "acceptAll",
                    ".claude/hooks/moai/critical_security.py": "deny",
                    "src/": "acceptEdits",
                    "tests/": "acceptEdits",
                    "docs/": "acceptEdits",
                    ".moai/": "acceptEdits",
                    "pyproject.toml": "ask",
                    "package.json": "ask",
                    "Makefile": "ask",
                    ".env*": "deny",
                    "*.key": "deny",
                    "*.pem": "deny",
                    "*.p12": "deny"
                }
            }
        }

        # Mode-specific adjustments
        if hasattr(config, 'runtime') and config.runtime:
            runtime_name = getattr(config.runtime, 'name', 'python')

            # Language-specific permissions
            if runtime_name == 'python':
                settings["permissions"]["overrides"].update({
                    "requirements.txt": "ask",
                    "setup.py": "ask",
                    "*.py": "acceptEdits"
                })
            elif runtime_name == 'javascript' or runtime_name == 'typescript':
                settings["permissions"]["overrides"].update({
                    "package.json": "ask",
                    "package-lock.json": "ask",
                    "*.js": "acceptEdits",
                    "*.ts": "acceptEdits",
                    "*.jsx": "acceptEdits",
                    "*.tsx": "acceptEdits"
                })
            elif runtime_name == 'java':
                settings["permissions"]["overrides"].update({
                    "pom.xml": "ask",
                    "build.gradle": "ask",
                    "*.java": "acceptEdits"
                })

        # Project-specific permissions
        project_name = getattr(config, 'name', 'project')
        settings["permissions"]["overrides"].update({
            f"{project_name}/": "acceptEdits",
            "README.md": "acceptEdits",
            "CHANGELOG.md": "acceptEdits",
            "LICENSE": "ask"
        })

        return settings

    def create_claude_settings_file(self, project_path: Path, config: Config) -> Path:
        """Create Claude settings file and return its path"""
        claude_dir = project_path / ".claude"
        settings_path = claude_dir / "settings.json"

        success = self.create_claude_settings(settings_path, config)

        if success:
            return settings_path
        else:
            raise RuntimeError(f"Failed to create Claude settings at {settings_path}")

    def validate_claude_settings(self, settings_path: Path) -> bool:
        """Claude 설정 파일 검증"""
        try:
            if not settings_path.exists():
                logger.warning(f"Claude settings file not found: {settings_path}")
                return False

            with open(settings_path, 'r', encoding='utf-8') as f:
                settings_data = json.load(f)

            # 필수 키 검증
            required_keys = ["defaultMode", "permissions"]
            for key in required_keys:
                if key not in settings_data:
                    logger.error(f"Missing required key in Claude settings: {key}")
                    return False

            # permissions 구조 검증
            permissions = settings_data.get("permissions", {})
            if "defaultBehavior" not in permissions:
                logger.error("Missing defaultBehavior in permissions")
                return False

            # Security validation
            if not self.security_manager.validate_claude_settings(settings_data):
                logger.error("Security validation failed for Claude settings")
                return False

            logger.info(f"Claude settings validated successfully: {settings_path}")
            return True

        except json.JSONDecodeError as e:
            logger.error(f"Invalid JSON in Claude settings: {e}")
            return False
        except Exception as e:
            logger.error(f"Error validating Claude settings: {e}")
            return False

    def update_claude_permissions(self, settings_path: Path, new_permissions: Dict[str, str]) -> bool:
        """Claude 권한 설정 업데이트"""
        try:
            if not settings_path.exists():
                logger.error(f"Claude settings file not found: {settings_path}")
                return False

            with open(settings_path, 'r', encoding='utf-8') as f:
                settings_data = json.load(f)

            # 권한 업데이트
            if "permissions" not in settings_data:
                settings_data["permissions"] = {"defaultBehavior": "acceptEdits"}

            if "overrides" not in settings_data["permissions"]:
                settings_data["permissions"]["overrides"] = {}

            settings_data["permissions"]["overrides"].update(new_permissions)

            # Security validation
            if not self.security_manager.validate_claude_settings(settings_data):
                logger.error("Security validation failed after permission update")
                return False

            # 파일 저장
            with open(settings_path, 'w', encoding='utf-8') as f:
                json.dump(settings_data, f, indent=2)

            logger.info(f"Updated Claude permissions: {list(new_permissions.keys())}")
            return True

        except Exception as e:
            logger.error(f"Error updating Claude permissions: {e}")
            return False

    def get_claude_permissions(self, settings_path: Path) -> Dict[str, str]:
        """Claude 권한 설정 조회"""
        try:
            if not settings_path.exists():
                return {}

            with open(settings_path, 'r', encoding='utf-8') as f:
                settings_data = json.load(f)

            permissions = settings_data.get("permissions", {})
            overrides = permissions.get("overrides", {})

            return overrides

        except Exception as e:
            logger.error(f"Error getting Claude permissions: {e}")
            return {}