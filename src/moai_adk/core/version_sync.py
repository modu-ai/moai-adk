"""
Version Synchronization System for MoAI-ADK

Ensures version consistency across multiple configuration files and provides
automatic synchronization capabilities.

Key Features:
- Version consistency validation across pyproject.toml, __init__.py, and config.json
- Automatic synchronization when inconsistencies detected
- Cache invalidation for statusline version readers
- Version change detection and notification system
- Fallback strategies for version resolution
"""

import json
import logging
import re
from dataclasses import dataclass, field
from datetime import datetime
from enum import Enum
from pathlib import Path
from typing import Any, Dict, List, Optional, Tuple

logger = logging.getLogger(__name__)


class VersionSource(Enum):
    """Enum for version source tracking"""

    PYPROJECT_TOML = "pyproject_toml"
    CONFIG_JSON = "config_json"
    PACKAGE_METADATA = "package_metadata"
    FALLBACK = "fallback"


@dataclass
class VersionInfo:
    """Version information with metadata"""

    version: str
    source: VersionSource
    file_path: Path
    raw_content: str
    timestamp: datetime = field(default_factory=datetime.now)
    is_valid: bool = True

    def __post_init__(self):
        """Validate version format after initialization"""
        self.is_valid = self._is_valid_version(self.version)

    def _is_valid_version(self, version: str) -> bool:
        """Check if version follows semantic versioning"""
        pattern = r"^\d+\.\d+\.\d+(?:-[a-zA-Z0-9]+)?$"
        return bool(re.match(pattern, version))


class VersionSynchronizer:
    """
    Version synchronization system for MoAI-ADK

    Ensures version consistency across multiple files and provides
    automatic synchronization capabilities.
    """

    def __init__(self, working_dir: Optional[Path] = None):
        """
        Initialize version synchronizer

        Args:
            working_dir: Project directory. If None, uses current working directory
        """
        if working_dir is None:
            working_dir = Path.cwd()

        self.working_dir = working_dir
        self.pyproject_path = working_dir / "pyproject.toml"
        self.config_path = working_dir / ".moai" / "config" / "config.json"
        self.init_path = working_dir / "src" / "moai_adk" / "__init__.py"

        # Version file paths in order of priority
        # Note: __init__.py reads version dynamically, so we don't synchronize it
        self.version_files = [
            (self.pyproject_path, VersionSource.PYPROJECT_TOML),
            (self.config_path, VersionSource.CONFIG_JSON),
        ]

        # Statusline cache directories to invalidate
        self.cache_directories = [
            working_dir / ".moai" / "cache",
            working_dir / ".claude" / "cache",
        ]

        logger.info(f"VersionSynchronizer initialized for: {working_dir}")

    def check_consistency(self) -> Tuple[bool, List[VersionInfo]]:
        """
        Check version consistency across all version files

        Returns:
            Tuple of (is_consistent, version_info_list)
        """
        version_infos = []

        for file_path, source in self.version_files:
            try:
                version_info = self._extract_version(file_path, source)
                if version_info:
                    version_infos.append(version_info)
            except Exception as e:
                logger.warning(f"Failed to extract version from {file_path}: {e}")
                continue

        if not version_infos:
            logger.error("No version information found in any file")
            return False, []

        # Check if all valid versions are the same
        valid_versions = [info.version for info in version_infos if info.is_valid]
        is_consistent = len(set(valid_versions)) <= 1

        if not is_consistent:
            logger.warning(f"Version inconsistency detected: {valid_versions}")

        return is_consistent, version_infos

    def get_master_version(self) -> Optional[VersionInfo]:
        """
        Get master version from pyproject.toml (primary source)

        Returns:
            VersionInfo from pyproject.toml or None if not found
        """
        try:
            version_info = self._extract_version(self.pyproject_path, VersionSource.PYPROJECT_TOML)
            if version_info and version_info.is_valid:
                logger.info(f"Master version from pyproject.toml: {version_info.version}")
                return version_info
        except Exception as e:
            logger.error(f"Failed to get master version from pyproject.toml: {e}")

        return None

    def synchronize_all(self, target_version: Optional[str] = None, dry_run: bool = False) -> bool:
        """
        Synchronize all version files to match master version

        Args:
            target_version: Target version to synchronize to. If None, uses pyproject.toml version
            dry_run: If True, only shows what would be changed without making changes

        Returns:
            True if synchronization successful
        """
        if target_version is None:
            master_info = self.get_master_version()
            if not master_info:
                logger.error("No master version found for synchronization")
                return False
            target_version = master_info.version

        logger.info(f"Synchronizing all files to version: {target_version}")

        sync_results = []

        for file_path, source in self.version_files:
            if source == VersionSource.PYPROJECT_TOML:
                continue  # Skip pyproject.toml as it's the source

            try:
                success = self._synchronize_file(file_path, source, target_version, dry_run)
                sync_results.append((file_path, success))

                if not dry_run and success:
                    logger.info(f"Synchronized {file_path} to version {target_version}")
                elif dry_run:
                    logger.info(f"[DRY RUN] Would synchronize {file_path} to version {target_version}")

            except Exception as e:
                logger.error(f"Failed to synchronize {file_path}: {e}")
                sync_results.append((file_path, False))

        # Clear caches after successful synchronization
        if not dry_run and all(success for _, success in sync_results):
            self._clear_caches()
            logger.info("All version files synchronized successfully")
            return True

        logger.error("Some files failed to synchronize")
        return False

    def _extract_version(self, file_path: Path, source: VersionSource) -> Optional[VersionInfo]:
        """Extract version from specific file based on source type"""
        if not file_path.exists():
            logger.debug(f"File not found: {file_path}")
            return None

        try:
            with open(file_path, "r", encoding="utf-8") as f:
                content = f.read()

            if source == VersionSource.PYPROJECT_TOML:
                version = self._extract_from_pyproject(content)
            elif source == VersionSource.CONFIG_JSON:
                version = self._extract_from_config(content)
            elif source == VersionSource.PACKAGE_METADATA:
                version = self._extract_from_init(content)
            else:
                return None

            if version:
                return VersionInfo(
                    version=version,
                    source=source,
                    file_path=file_path,
                    raw_content=content,
                )

        except Exception as e:
            logger.error(f"Error reading {file_path}: {e}")

        return None

    def _extract_from_pyproject(self, content: str) -> Optional[str]:
        """Extract version from pyproject.toml"""
        pattern = r'^version\s*=\s*["\']([^"\']+)["\']'
        match = re.search(pattern, content, re.MULTILINE)
        return match.group(1) if match else None

    def _extract_from_config(self, content: str) -> Optional[str]:
        """Extract version from .moai/config/config.json"""
        try:
            config_data = json.loads(content)
            moai_config = config_data.get("moai", {})
            return moai_config.get("version")
        except json.JSONDecodeError:
            logger.error("Invalid JSON in config.json")
            return None

    def _extract_from_init(self, content: str) -> Optional[str]:
        """Extract version from __init__.py"""
        pattern = r'__version__\s*=\s*["\']([^"\']+)["\']'
        match = re.search(pattern, content)
        return match.group(1) if match else None

    def _synchronize_file(self, file_path: Path, source: VersionSource, target_version: str, dry_run: bool) -> bool:
        """Synchronize specific file to target version"""
        current_info = self._extract_version(file_path, source)
        if not current_info:
            logger.warning(f"No current version found in {file_path}")
            return True  # Nothing to synchronize

        if current_info.version == target_version:
            logger.debug(f"Version already synchronized in {file_path}")
            return True

        if dry_run:
            return True

        try:
            new_content = self._update_version_in_content(current_info.raw_content, source, target_version)

            with open(file_path, "w", encoding="utf-8") as f:
                f.write(new_content)

            return True

        except Exception as e:
            logger.error(f"Failed to update {file_path}: {e}")
            return False

    def _update_version_in_content(self, content: str, source: VersionSource, target_version: str) -> str:
        """Update version in content based on source type"""
        if source == VersionSource.CONFIG_JSON:
            # Update moai.version in JSON
            try:
                config_data = json.loads(content)
                if "moai" not in config_data:
                    config_data["moai"] = {}
                config_data["moai"]["version"] = target_version
                return json.dumps(config_data, indent=2)
            except json.JSONDecodeError:
                # Fallback: regex replacement
                pattern = r'("moai"\s*:\s*\{[^}]*"version"\s*:\s*")[^"\']+(")'
                return re.sub(pattern, f"\\1{target_version}\\2", content, flags=re.DOTALL)

        elif source == VersionSource.PACKAGE_METADATA:
            # Update __version__ in Python (fallback version)
            pattern = r'(__version__\s*=\s*["\'])([^"\']+)(["\'])'
            return re.sub(pattern, f"\\1{target_version}\\3", content)

        return content

    def _clear_caches(self) -> None:
        """Clear all relevant caches"""
        cleared_count = 0

        for cache_dir in self.cache_directories:
            if cache_dir.exists():
                try:
                    import shutil

                    cache_size = sum(1 for _ in cache_dir.rglob("*") if _.is_file())
                    shutil.rmtree(cache_dir)
                    cache_dir.mkdir(exist_ok=True)
                    cleared_count += cache_size
                    logger.info(f"Cleared cache directory: {cache_dir} ({cache_size} files)")
                except Exception as e:
                    logger.warning(f"Failed to clear cache directory {cache_dir}: {e}")

        # Clear version reader cache if available
        try:
            from ..statusline.version_reader import VersionReader

            reader = VersionReader(working_dir=self.working_dir)
            reader.clear_cache()
            logger.info("Cleared VersionReader cache")
            cleared_count += 1
        except ImportError:
            pass
        except Exception as e:
            logger.warning(f"Failed to clear VersionReader cache: {e}")

        logger.info(f"Cache clearing completed: {cleared_count} cache entries cleared")

    def validate_version_format(self, version: str) -> Tuple[bool, Optional[str]]:
        """
        Validate version format and return normalized version

        Args:
            version: Version string to validate

        Returns:
            Tuple of (is_valid, normalized_version)
        """
        if not version:
            return False, None

        # Remove 'v' prefix if present
        normalized = version[1:] if version.startswith("v") else version

        # Validate semantic versioning
        pattern = r"^(\d+)\.(\d+)\.(\d+)(?:[-.]([a-zA-Z0-9]+))?$"
        match = re.match(pattern, normalized)

        if match:
            return True, normalized

        return False, None

    def get_version_report(self) -> Dict[str, Any]:
        """
        Generate comprehensive version report

        Returns:
            Dictionary containing version information and status
        """
        is_consistent, version_infos = self.check_consistency()

        report: Dict[str, Any] = {
            "timestamp": datetime.now().isoformat(),
            "working_directory": str(self.working_dir),
            "is_consistent": is_consistent,
            "version_count": len(version_infos),
            "versions": [],
            "issues": [],
            "recommendations": [],
        }

        # Add version information
        versions_list: List[Dict[str, Any]] = report["versions"]  # type: ignore[assignment]
        for info in version_infos:
            versions_list.append(
                {
                    "version": info.version,
                    "source": info.source.value,
                    "file": str(info.file_path),
                    "is_valid": info.is_valid,
                    "last_modified": info.timestamp.isoformat(),
                }
            )

        # Add issues and recommendations
        issues_list: List[str] = report["issues"]  # type: ignore[assignment]
        recommendations_list: List[str] = report["recommendations"]  # type: ignore[assignment]

        if not version_infos:
            issues_list.append("No version information found in any file")
            recommendations_list.append("Initialize version in pyproject.toml")

        if not is_consistent:
            issues_list.append("Version inconsistency detected across files")
            recommendations_list.append("Run synchronize_all() to fix inconsistencies")

        invalid_versions = [info for info in version_infos if not info.is_valid]
        if invalid_versions:
            issues_list.append(f"Invalid version format in {len(invalid_versions)} files")
            recommendations_list.append("Fix version format to follow semantic versioning")

        return report


# Convenience functions for common operations
def check_project_versions(working_dir: Optional[Path] = None) -> Dict[str, Any]:
    """
    Check version consistency for the project

    Args:
        working_dir: Project directory (defaults to current directory)

    Returns:
        Version report dictionary
    """
    synchronizer = VersionSynchronizer(working_dir)
    return synchronizer.get_version_report()


def synchronize_project_versions(working_dir: Optional[Path] = None, dry_run: bool = False) -> bool:
    """
    Synchronize all version files in the project

    Args:
        working_dir: Project directory (defaults to current directory)
        dry_run: If True, only show what would be changed

    Returns:
        True if synchronization successful
    """
    synchronizer = VersionSynchronizer(working_dir)
    return synchronizer.synchronize_all(dry_run=dry_run)
