"""
Configuration utilities: validation, backup, and file operations.

@FEATURE:CONFIG-UTILS-001 Configuration validation and utility functions
@DESIGN:SEPARATED-UTILS-001 Extracted from oversized config_manager.py (564 LOC)
"""

import json
import shutil
from datetime import datetime
from pathlib import Path
from typing import Any

from ..utils.logger import get_logger
from .security import SecurityManager

logger = get_logger(__name__)


class ConfigUtils:
    """설정 유틸리티 관리자"""

    def __init__(self, security_manager: SecurityManager = None):
        """유틸리티 관리자 초기화"""
        self.security_manager = security_manager or SecurityManager()

    def write_json_file(self, file_path: Path, data: dict[str, Any]) -> Path:
        """
        Write JSON data to file with error handling and logging.

        Args:
            file_path: Path to the JSON file
            data: Data to write

        Returns:
            Path: The written file path

        Raises:
            RuntimeError: If writing fails
        """
        try:
            # Ensure parent directory exists
            file_path.parent.mkdir(parents=True, exist_ok=True)

            # Security validation
            if not self.security_manager.validate_json_content(data):
                raise RuntimeError("Security validation failed for JSON content")

            # Write file with proper formatting
            with open(file_path, "w", encoding="utf-8") as f:
                json.dump(data, f, indent=2, ensure_ascii=False)

            logger.info(f"Successfully wrote JSON file: {file_path}")
            return file_path

        except json.JSONEncodeError as e:
            error_msg = f"JSON encoding error for {file_path}: {e}"
            logger.error(error_msg)
            raise RuntimeError(error_msg)

        except OSError as e:
            error_msg = f"File I/O error for {file_path}: {e}"
            logger.error(error_msg)
            raise RuntimeError(error_msg)

        except Exception as e:
            error_msg = f"Unexpected error writing {file_path}: {e}"
            logger.error(error_msg)
            raise RuntimeError(error_msg)

    def validate_config_file(self, file_path: Path) -> bool:
        """
        Validate configuration file format and content.

        Args:
            file_path: Path to configuration file

        Returns:
            bool: True if valid, False otherwise
        """
        try:
            if not file_path.exists():
                logger.warning(f"Config file does not exist: {file_path}")
                return False

            if not file_path.is_file():
                logger.error(f"Path is not a file: {file_path}")
                return False

            # Check file extension
            if file_path.suffix.lower() != ".json":
                logger.warning(f"Config file is not JSON: {file_path}")
                # Don't return False here - allow non-JSON config files

            # Try to parse JSON content
            with open(file_path, encoding="utf-8") as f:
                config_data = json.load(f)

            # Basic structure validation
            if not isinstance(config_data, dict):
                logger.error(f"Config file must contain JSON object: {file_path}")
                return False

            # Security validation
            if not self.security_manager.validate_json_content(config_data):
                logger.error(f"Security validation failed: {file_path}")
                return False

            # File-specific validation based on filename
            if file_path.name == "settings.json":
                return self._validate_claude_settings_structure(config_data)
            elif file_path.name == "config.json":
                return self._validate_moai_config_structure(config_data)
            elif file_path.name == "package.json":
                return self._validate_package_json_structure(config_data)

            logger.info(f"Config file validated successfully: {file_path}")
            return True

        except json.JSONDecodeError as e:
            logger.error(f"Invalid JSON in config file {file_path}: {e}")
            return False

        except Exception as e:
            logger.error(f"Error validating config file {file_path}: {e}")
            return False

    def _validate_claude_settings_structure(self, data: dict[str, Any]) -> bool:
        """Claude settings 구조 검증"""
        required_keys = ["defaultMode", "permissions"]
        for key in required_keys:
            if key not in data:
                logger.error(f"Missing required key in Claude settings: {key}")
                return False

        permissions = data.get("permissions", {})
        if "defaultBehavior" not in permissions:
            logger.error("Missing defaultBehavior in permissions")
            return False

        return True

    def _validate_moai_config_structure(self, data: dict[str, Any]) -> bool:
        """MoAI config 구조 검증"""
        required_keys = ["version", "project"]
        for key in required_keys:
            if key not in data:
                logger.error(f"Missing required key in MoAI config: {key}")
                return False

        project = data.get("project", {})
        project_required_keys = ["name", "path"]
        for key in project_required_keys:
            if key not in project:
                logger.error(f"Missing required key in project config: {key}")
                return False

        return True

    def _validate_package_json_structure(self, data: dict[str, Any]) -> bool:
        """package.json 구조 검증"""
        required_keys = ["name", "version"]
        for key in required_keys:
            if key not in data:
                logger.error(f"Missing required key in package.json: {key}")
                return False

        return True

    def backup_config_file(self, file_path: Path) -> Path:
        """
        Create backup of configuration file.

        Args:
            file_path: Path to config file to backup

        Returns:
            Path: Path to backup file

        Raises:
            RuntimeError: If backup creation fails
        """
        try:
            if not file_path.exists():
                raise RuntimeError(f"Source file does not exist: {file_path}")

            # Create backup filename with timestamp
            timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
            backup_name = f"{file_path.stem}_{timestamp}{file_path.suffix}"
            backup_path = file_path.parent / f"backup_{backup_name}"

            # Ensure backup directory exists
            backup_path.parent.mkdir(parents=True, exist_ok=True)

            # Create backup
            shutil.copy2(file_path, backup_path)

            logger.info(f"Created config backup: {file_path} -> {backup_path}")
            return backup_path

        except Exception as e:
            error_msg = f"Failed to backup config file {file_path}: {e}"
            logger.error(error_msg)
            raise RuntimeError(error_msg)

    def restore_config_from_backup(self, backup_path: Path, target_path: Path) -> bool:
        """백업으로부터 설정 파일 복원"""
        try:
            if not backup_path.exists():
                logger.error(f"Backup file does not exist: {backup_path}")
                return False

            # Validate backup file first
            if not self.validate_config_file(backup_path):
                logger.error(f"Backup file validation failed: {backup_path}")
                return False

            # Create current backup before restore
            if target_path.exists():
                current_backup = self.backup_config_file(target_path)
                logger.info(f"Created current backup before restore: {current_backup}")

            # Restore from backup
            shutil.copy2(backup_path, target_path)
            logger.info(f"Restored config from backup: {backup_path} -> {target_path}")

            return True

        except Exception as e:
            logger.error(f"Failed to restore config from backup: {e}")
            return False

    def merge_config_files(
        self,
        base_config_path: Path,
        overlay_config_path: Path,
        output_path: Path = None,
    ) -> dict[str, Any]:
        """두 설정 파일 병합"""
        try:
            # Load base configuration
            with open(base_config_path, encoding="utf-8") as f:
                base_config = json.load(f)

            # Load overlay configuration
            with open(overlay_config_path, encoding="utf-8") as f:
                overlay_config = json.load(f)

            # Deep merge configurations
            merged_config = self._deep_merge_dicts(base_config, overlay_config)

            # Save to output path if specified
            if output_path:
                self.write_json_file(output_path, merged_config)
                logger.info(f"Merged config saved to: {output_path}")

            return merged_config

        except Exception as e:
            logger.error(f"Failed to merge config files: {e}")
            raise RuntimeError(f"Config merge failed: {e}")

    def _deep_merge_dicts(
        self, base: dict[str, Any], overlay: dict[str, Any]
    ) -> dict[str, Any]:
        """딕셔너리 깊은 병합"""
        merged = base.copy()

        for key, value in overlay.items():
            if (
                key in merged
                and isinstance(merged[key], dict)
                and isinstance(value, dict)
            ):
                merged[key] = self._deep_merge_dicts(merged[key], value)
            else:
                merged[key] = value

        return merged

    def get_config_summary(self, config_path: Path) -> dict[str, Any]:
        """설정 파일 요약 정보"""
        try:
            if not config_path.exists():
                return {"status": "not_found"}

            stat = config_path.stat()
            summary = {
                "status": "exists",
                "size": stat.st_size,
                "modified": datetime.fromtimestamp(stat.st_mtime).isoformat(),
                "valid": self.validate_config_file(config_path),
            }

            # Add content summary if it's a JSON file
            if config_path.suffix.lower() == ".json":
                try:
                    with open(config_path, encoding="utf-8") as f:
                        data = json.load(f)
                    summary["keys"] = (
                        list(data.keys()) if isinstance(data, dict) else []
                    )
                except:
                    summary["keys"] = []

            return summary

        except Exception as e:
            return {"status": "error", "error": str(e)}

    def cleanup_old_backups(
        self, config_dir: Path, max_age_days: int = 30
    ) -> list[Path]:
        """오래된 백업 파일 정리"""
        deleted_files = []

        try:
            if not config_dir.exists():
                return deleted_files

            cutoff_time = datetime.now().timestamp() - (max_age_days * 24 * 60 * 60)

            for backup_file in config_dir.glob("backup_*"):
                if backup_file.is_file() and backup_file.stat().st_mtime < cutoff_time:
                    backup_file.unlink()
                    deleted_files.append(backup_file)
                    logger.info(f"Deleted old backup: {backup_file}")

        except Exception as e:
            logger.error(f"Error cleaning up old backups: {e}")

        return deleted_files
