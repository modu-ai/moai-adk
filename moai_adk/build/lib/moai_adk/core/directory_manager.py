"""
Directory management utilities for MoAI-ADK.

Handles directory creation, structure setup, and safe directory operations
with security validation and proper error handling.
"""

import tempfile
import os
import shutil
from pathlib import Path
from typing import List

from ..logger import get_logger
from .security import SecurityManager, SecurityError
from ..config import Config

logger = get_logger(__name__)


class DirectoryManager:
    """Manages directory operations for MoAI-ADK installation."""

    def __init__(self, security_manager: SecurityManager = None):
        """
        Initialize directory manager.

        Args:
            security_manager: Security manager instance for validation
        """
        self.security_manager = security_manager or SecurityManager()

    def create_project_directory(self, config: Config) -> None:
        """
        Create the main project directory, preserving existing files when safe.

        Args:
            config: Project configuration containing path and settings
        """
        project_path = config.project_path

        # 현재 디렉토리인지 확인
        is_current_dir = (
            project_path.resolve() == Path.cwd().resolve() or
            str(project_path) == "."
        )

        if project_path.exists():
            # 현재 디렉토리나 기존 프로젝트는 삭제하지 않음
            if is_current_dir or (config.is_existing_project and not config.force_overwrite):
                logger.info("Using existing directory (preserving all files): %s", project_path)
                return

            # 새 프로젝트이고 강제 덮어쓰기가 요청된 경우에만 기존 처리
            if config.force_overwrite:
                self._handle_force_overwrite(project_path)
            else:
                # 기존 디렉토리가 있지만 강제 옵션이 없는 경우
                logger.info("Directory exists, preserving files: %s", project_path)
        else:
            # 디렉토리가 존재하지 않는 경우에만 생성
            project_path.mkdir(parents=True, exist_ok=True)
            logger.info("Created new project directory: %s", project_path)

    def _handle_force_overwrite(self, project_path: Path) -> None:
        """
        Handle force overwrite operation with Git directory preservation.

        Args:
            project_path: Path to project directory
        """
        # Always preserve .git directory if it exists (safer approach)
        git_dir = project_path / ".git"
        git_backup = None

        if git_dir.exists():
            # 🔒 보안 강화: 임시 디렉토리 권한 0o700으로 제한
            temp_dir = tempfile.mkdtemp()
            os.chmod(temp_dir, 0o700)  # 소유자만 접근 가능
            git_backup = Path(temp_dir) / ".git"
            shutil.move(str(git_dir), str(git_backup))
            logger.info("Preserving existing .git directory")

        # Remove the project directory only when force is explicitly requested
        logger.warning("Force overwrite enabled - removing existing directory: %s", project_path)
        if not self.security_manager.safe_rmtree(project_path):
            raise SecurityError(f"Failed to safely delete directory: {project_path}")

        # Recreate project directory
        project_path.mkdir(parents=True, exist_ok=True)

        # Restore .git directory if it was preserved
        if git_backup and git_backup.exists():
            shutil.move(str(git_backup), str(git_dir))
            git_backup.parent.rmdir()  # Clean up temp directory
            logger.info("Restored .git directory")

    def create_directory_structure(self, base_path: Path) -> List[Path]:
        """
        Create the complete MoAI-ADK directory structure.

        Args:
            base_path: Base project path

        Returns:
            List[Path]: List of created directories
        """
        directories = [
            # Claude Code 표준 디렉토리
            base_path / ".claude",
            base_path / ".claude" / "commands" / "moai",
            base_path / ".claude" / "agents" / "moai",
            base_path / ".claude" / "hooks" / "moai",
            base_path / ".claude" / "memory",
            base_path / ".claude" / "logs",
            base_path / ".claude" / "output-styles",

            # MoAI 문서 시스템
            base_path / ".moai",
            base_path / ".moai" / "templates",
            base_path / ".moai" / "steering",
            base_path / ".moai" / "memory",
            base_path / ".moai" / "memory" / "decisions",
            base_path / ".moai" / "specs",
            base_path / ".moai" / "indexes",
            base_path / ".moai" / "reports",
            base_path / ".moai" / "scripts",

            # GitHub Actions (옵션)
            base_path / ".github" / "workflows",
        ]

        created_dirs = []
        for directory in directories:
            try:
                # Security validation for each directory
                if not self.security_manager.validate_path_safety(directory, base_path):
                    logger.error("Security validation failed for directory: %s", directory)
                    continue

                directory.mkdir(parents=True, exist_ok=True)
                created_dirs.append(directory)
                logger.debug("Created directory: %s", directory)

            except Exception as e:
                logger.error("Failed to create directory %s: %s", directory, e)

        logger.info("Created %d directories in project structure", len(created_dirs))
        return created_dirs

    def ensure_directory_exists(self, directory: Path, base_path: Path = None) -> bool:
        """
        Ensure a directory exists, creating it if necessary with security validation.

        Args:
            directory: Directory path to ensure exists
            base_path: Base path for security validation (optional)

        Returns:
            bool: True if directory exists or was created successfully
        """
        try:
            # Security validation if base path provided
            if base_path and not self.security_manager.validate_path_safety(directory, base_path):
                logger.error("Security validation failed for directory: %s", directory)
                return False

            directory.mkdir(parents=True, exist_ok=True)
            logger.debug("Ensured directory exists: %s", directory)
            return True

        except Exception as e:
            logger.error("Failed to ensure directory exists %s: %s", directory, e)
            return False

    def get_directory_info(self, directory: Path) -> dict:
        """
        Get information about a directory.

        Args:
            directory: Directory to analyze

        Returns:
            dict: Directory information including size, file counts, etc.
        """
        if not directory.exists():
            return {
                'exists': False,
                'is_directory': False,
                'file_count': 0,
                'subdirectory_count': 0,
                'total_size': 0
            }

        if not directory.is_dir():
            return {
                'exists': True,
                'is_directory': False,
                'file_count': 0,
                'subdirectory_count': 0,
                'total_size': directory.stat().st_size if directory.exists() else 0
            }

        try:
            file_count = 0
            subdirectory_count = 0
            total_size = 0

            for item in directory.rglob('*'):
                if item.is_file():
                    file_count += 1
                    total_size += item.stat().st_size
                elif item.is_dir():
                    subdirectory_count += 1

            return {
                'exists': True,
                'is_directory': True,
                'file_count': file_count,
                'subdirectory_count': subdirectory_count,
                'total_size': total_size,
                'size_mb': round(total_size / (1024 * 1024), 2)
            }

        except Exception as e:
            logger.error("Error getting directory info for %s: %s", directory, e)
            return {
                'exists': True,
                'is_directory': True,
                'file_count': 0,
                'subdirectory_count': 0,
                'total_size': 0,
                'error': str(e)
            }

    def clean_directory(self, directory: Path, preserve_patterns: List[str] = None) -> bool:
        """
        Clean directory contents while preserving specified patterns.

        Args:
            directory: Directory to clean
            preserve_patterns: List of glob patterns to preserve

        Returns:
            bool: True if successful
        """
        if not directory.exists() or not directory.is_dir():
            return False

        preserve_patterns = preserve_patterns or []

        try:
            for item in directory.iterdir():
                # Check if item should be preserved
                should_preserve = False
                for pattern in preserve_patterns:
                    if item.match(pattern):
                        should_preserve = True
                        break

                if should_preserve:
                    logger.debug("Preserving item: %s", item)
                    continue

                # Remove item
                if item.is_file():
                    item.unlink()
                    logger.debug("Removed file: %s", item)
                elif item.is_dir():
                    if not self.security_manager.safe_rmtree(item):
                        logger.error("Failed to safely remove directory: %s", item)
                        return False
                    logger.debug("Removed directory: %s", item)

            logger.info("Cleaned directory: %s", directory)
            return True

        except Exception as e:
            logger.error("Error cleaning directory %s: %s", directory, e)
            return False

    def create_backup_directory(self, source_path: Path, backup_base: Path = None) -> Path:
        """
        Create a backup directory with timestamp.

        Args:
            source_path: Source directory to backup
            backup_base: Base directory for backups (default: parent of source)

        Returns:
            Path: Path to created backup directory
        """
        import datetime

        if backup_base is None:
            backup_base = source_path.parent

        timestamp = datetime.datetime.now().strftime("%Y%m%d_%H%M%S")
        backup_name = f"{source_path.name}_backup_{timestamp}"
        backup_path = backup_base / backup_name

        try:
            # Security validation
            if not self.security_manager.validate_path_safety(backup_path, backup_base):
                raise SecurityError(f"Security validation failed for backup path: {backup_path}")

            # Create backup
            if source_path.is_file():
                backup_path.parent.mkdir(parents=True, exist_ok=True)
                shutil.copy2(source_path, backup_path)
            elif source_path.is_dir():
                shutil.copytree(source_path, backup_path, dirs_exist_ok=True)
            else:
                raise ValueError(f"Source path does not exist or is not a file/directory: {source_path}")

            logger.info("Created backup: %s -> %s", source_path, backup_path)
            return backup_path

        except Exception as e:
            logger.error("Failed to create backup of %s: %s", source_path, e)
            raise