"""
MoAI Menu Project - Configuration Migration Manager

This module handles migration from legacy configuration formats to the new
unified configuration system, supporting rollback and history tracking.
"""

import json
import logging
from dataclasses import dataclass
from pathlib import Path
from typing import Any, Dict, List, Optional, Union

from .config_manager import UnifiedConfigManager

logger = logging.getLogger(__name__)


@dataclass
class MigrationStep:
    """Single migration step definition."""

    from_version: str
    to_version: str
    description: str
    migration_func: str
    rollback_func: Optional[str] = None


@dataclass
class LegacyConfigInfo:
    """Information about legacy configuration file."""

    path: Path
    module_name: str
    version: str
    format: str  # 'json', 'yaml', 'ini', etc.
    priority: int  # Migration priority (lower = higher priority)


class LegacyConfigDetector:
    """Detects and analyzes legacy configuration files."""

    def __init__(self, search_dirs: List[Path]):
        """
        Initialize detector with search directories.

        Args:
            search_dirs: List of directories to search for legacy configs
        """
        self.search_dirs = search_dirs
        self.legacy_patterns = {
            "batch_questions": [
                "batch_questions_config.json",
                "batch-config.json",
                ".batch_questions.json",
            ],
            "documentation": [
                "docs_config.json",
                "documentation.json",
                ".docs_config.json",
            ],
            "language_config": [
                "language_settings.json",
                "i18n_config.json",
                ".language_config.json",
            ],
            "template_optimizer": [
                "template_config.json",
                "optimizer_config.json",
                ".template_config.json",
            ],
            "project_initializer": [
                "init_config.json",
                "project_config.json",
                ".init_config.json",
            ],
        }

    def find_legacy_configs(self) -> List[LegacyConfigInfo]:
        """
        Find all legacy configuration files.

        Returns:
            List of legacy configuration information
        """
        legacy_configs = []

        for search_dir in self.search_dirs:
            if not search_dir.exists():
                continue

            for module_name, patterns in self.legacy_patterns.items():
                for pattern in patterns:
                    for file_path in search_dir.glob(f"**/{pattern}"):
                        try:
                            config_info = self._analyze_config_file(file_path, module_name)
                            if config_info:
                                legacy_configs.append(config_info)
                        except Exception as e:
                            logger.warning(f"Failed to analyze {file_path}: {e}")

        # Sort by priority and path
        legacy_configs.sort(key=lambda x: (x.priority, str(x.path)))
        return legacy_configs

    def _analyze_config_file(self, file_path: Path, module_name: str) -> Optional[LegacyConfigInfo]:
        """Analyze individual configuration file."""
        if not file_path.is_file():
            return None

        # Determine format
        suffix = file_path.suffix.lower()
        if suffix == ".json":
            format_type = "json"
        elif suffix in [".yml", ".yaml"]:
            format_type = "yaml"
        elif suffix == ".ini":
            format_type = "ini"
        else:
            format_type = "unknown"

        # Try to load and analyze
        try:
            with open(file_path, "r", encoding="utf-8") as f:
                if format_type == "json":
                    data = json.load(f)
                elif format_type == "yaml":
                    try:
                        import yaml

                        data = yaml.safe_load(f)
                    except ImportError:
                        logger.warning(f"PyYAML not available, skipping {file_path}")
                        return None
                else:
                    return None

            # Extract version if available
            version = data.get("version", "1.0.0") if isinstance(data, dict) else "1.0.0"

            # Determine priority based on file location and naming
            priority = self._calculate_priority(file_path, module_name)

            return LegacyConfigInfo(
                path=file_path,
                module_name=module_name,
                version=version,
                format=format_type,
                priority=priority,
            )

        except Exception as e:
            logger.warning(f"Failed to load config file {file_path}: {e}")
            return None

    def _calculate_priority(self, file_path: Path, module_name: str) -> int:
        """Calculate migration priority for config file."""
        priority = 100  # Default priority

        # Hidden files have higher priority
        if file_path.name.startswith("."):
            priority -= 30

        # Files in .moai or .claude directories have higher priority
        if ".moai" in str(file_path) or ".claude" in str(file_path):
            priority -= 20

        # Files with exact module name have higher priority
        if module_name in file_path.name:
            priority -= 10

        # Files closer to root have higher priority
        depth = len(file_path.relative_to(Path.cwd()).parts)
        priority += depth * 5

        return priority


class ConfigurationMigrator:
    """
    Handles migration from legacy configurations to unified format.
    """

    def __init__(self, config_manager: UnifiedConfigManager, dry_run: bool = False):
        """
        Initialize migrator.

        Args:
            config_manager: Target configuration manager
            dry_run: If True, only show what would be migrated
        """
        self.config_manager = config_manager
        self.dry_run = dry_run
        self.detector = LegacyConfigDetector([Path.cwd()])

        # Migration steps registry
        self.migration_steps: Dict[str, List[MigrationStep]] = {
            "batch_questions": self._get_batch_questions_migrations(),
            "documentation": self._get_documentation_migrations(),
            "language_config": self._get_language_config_migrations(),
            "template_optimizer": self._get_template_optimizer_migrations(),
            "project_initializer": self._get_project_initializer_migrations(),
        }

    def detect_and_migrate(self, backup: bool = True) -> Dict[str, Any]:
        """
        Detect legacy configurations and migrate them.

        Args:
            backup: Whether to create backup before migration

        Returns:
            Migration result summary
        """
        logger.info("Detecting legacy configurations...")
        legacy_configs = self.detector.find_legacy_configs()

        if not legacy_configs:
            logger.info("No legacy configurations found")
            return {"migrated": 0, "failed": 0, "skipped": 0, "details": []}

        logger.info(f"Found {len(legacy_configs)} legacy configuration files")

        # Group by module
        module_configs = {}
        for config_info in legacy_configs:
            if config_info.module_name not in module_configs:
                module_configs[config_info.module_name] = []
            module_configs[config_info.module_name].append(config_info)

        # Migrate each module
        results = {"migrated": 0, "failed": 0, "skipped": 0, "details": []}

        for module_name, configs in module_configs.items():
            try:
                module_result = self._migrate_module(module_name, configs, backup)
                results["details"].append(module_result)

                if module_result["status"] == "success":
                    results["migrated"] += len(configs)
                elif module_result["status"] == "failed":
                    results["failed"] += len(configs)
                else:
                    results["skipped"] += len(configs)

            except Exception as e:
                logger.error(f"Failed to migrate module {module_name}: {e}")
                results["failed"] += len(configs)
                results["details"].append(
                    {
                        "module": module_name,
                        "status": "failed",
                        "error": str(e),
                        "files": [str(c.path) for c in configs],
                    }
                )

        return results

    def _migrate_module(self, module_name: str, configs: List[LegacyConfigInfo], backup: bool) -> Dict[str, Any]:
        """Migrate configurations for a specific module."""
        logger.info(f"Migrating {len(configs)} configuration files for module {module_name}")

        # Load existing unified config
        unified_config = self.config_manager.load_config()

        # Merge legacy configurations
        merged_legacy_config = {}

        for config_info in configs:
            try:
                legacy_data = self._load_legacy_config(config_info)

                # Apply transformations based on version
                transformed_data = self._transform_legacy_data(
                    legacy_data, config_info.module_name, config_info.version
                )

                # Merge into unified format
                merged_legacy_config.update(transformed_data)

                logger.debug(f"Merged {config_info.path}")

            except Exception as e:
                logger.error(f"Failed to load {config_info.path}: {e}")
                raise

        if not merged_legacy_config:
            return {
                "module": module_name,
                "status": "skipped",
                "message": "No valid configuration data found",
                "files": [str(c.path) for c in configs],
            }

        # Create backup if requested
        if backup and not self.dry_run:
            self.config_manager._create_backup()

        # Update unified configuration
        if module_name in unified_config:
            # Merge with existing configuration
            unified_config[module_name].update(merged_legacy_config)
        else:
            unified_config[module_name] = merged_legacy_config

        # Save updated configuration
        if not self.dry_run:
            self.config_manager.save_config(unified_config)
            logger.info(f"Successfully migrated {module_name} configuration")

        return {
            "module": module_name,
            "status": "success",
            "merged_keys": list(merged_legacy_config.keys()),
            "files": [str(c.path) for c in configs],
            "dry_run": self.dry_run,
        }

    def _load_legacy_config(self, config_info: LegacyConfigInfo) -> Dict[str, Any]:
        """Load legacy configuration file."""
        try:
            with open(config_info.path, "r", encoding="utf-8") as f:
                if config_info.format == "json":
                    return json.load(f)
                elif config_info.format == "yaml":
                    try:
                        import yaml

                        return yaml.safe_load(f)
                    except ImportError:
                        raise ImportError("PyYAML is required for YAML migration")
                else:
                    raise ValueError(f"Unsupported format: {config_info.format}")

        except Exception as e:
            logger.error(f"Failed to load {config_info.path}: {e}")
            raise

    def _transform_legacy_data(self, data: Dict[str, Any], module_name: str, version: str) -> Dict[str, Any]:
        """Transform legacy data to unified format."""
        if module_name not in self.migration_steps:
            logger.warning(f"No migration steps defined for {module_name}")
            return data

        # Apply version-specific transformations
        for step in self.migration_steps[module_name]:
            if self._version_matches(version, step.from_version):
                try:
                    migration_func = getattr(self, step.migration_func)
                    data = migration_func(data)
                    logger.debug(f"Applied migration {step.from_version}->{step.to_version} for {module_name}")
                except Exception as e:
                    logger.error(f"Failed to apply migration {step.from_version}->{step.to_version}: {e}")
                    raise

        return data

    def _version_matches(self, current_version: str, target_version: str) -> bool:
        """Check if current version matches target pattern."""
        # Simple version matching - can be enhanced
        return current_version.startswith(target_version.rstrip("x"))

    def _get_batch_questions_migrations(self) -> List[MigrationStep]:
        """Get migration steps for batch questions module."""
        return [
            MigrationStep(
                from_version="1.0",
                to_version="1.0.0",
                description="Initial format standardization",
                migration_func="_migrate_batch_questions_1_0",
            )
        ]

    def _get_documentation_migrations(self) -> List[MigrationStep]:
        """Get migration steps for documentation module."""
        return [
            MigrationStep(
                from_version="1.0",
                to_version="1.0.0",
                description="Initial format standardization",
                migration_func="_migrate_documentation_1_0",
            )
        ]

    def _get_language_config_migrations(self) -> List[MigrationStep]:
        """Get migration steps for language configuration module."""
        return [
            MigrationStep(
                from_version="1.0",
                to_version="1.0.0",
                description="Initial format standardization",
                migration_func="_migrate_language_config_1_0",
            )
        ]

    def _get_template_optimizer_migrations(self) -> List[MigrationStep]:
        """Get migration steps for template optimizer module."""
        return [
            MigrationStep(
                from_version="1.0",
                to_version="1.0.0",
                description="Initial format standardization",
                migration_func="_migrate_template_optimizer_1_0",
            )
        ]

    def _get_project_initializer_migrations(self) -> List[MigrationStep]:
        """Get migration steps for project initializer module."""
        return [
            MigrationStep(
                from_version="1.0",
                to_version="1.0.0",
                description="Initial format standardization",
                migration_func="_migrate_project_initializer_1_0",
            )
        ]

    # Migration functions
    def _migrate_batch_questions_1_0(self, data: Dict[str, Any]) -> Dict[str, Any]:
        """Migrate batch questions from 1.0 format."""
        # Example transformation logic
        if "max_batch_size" in data:
            data["max_questions"] = data.pop("max_batch_size")

        if "timeout" in data:
            data["timeout_seconds"] = data.pop("timeout")

        # Ensure default values
        data.setdefault("enabled", True)
        data.setdefault("max_questions", 50)
        data.setdefault("timeout_seconds", 300)
        data.setdefault("retry_attempts", 3)
        data.setdefault("output_format", "json")

        return data

    def _migrate_documentation_1_0(self, data: Dict[str, Any]) -> Dict[str, Any]:
        """Migrate documentation from 1.0 format."""
        # Example transformation logic
        if "auto_docs" in data:
            data["auto_generate"] = data.pop("auto_docs")

        if "doc_format" in data:
            data["format"] = data.pop("doc_format")

        # Ensure defaults
        data.setdefault("auto_generate", True)
        data.setdefault("format", "markdown")
        data.setdefault("include_api_docs", True)
        data.setdefault("include_examples", True)
        data.setdefault("template_engine", "jinja2")
        data.setdefault("output_directory", "docs")

        return data

    def _migrate_language_config_1_0(self, data: Dict[str, Any]) -> Dict[str, Any]:
        """Migrate language configuration from 1.0 format."""
        # Example transformation logic
        if "default_lang" in data:
            data["default_language"] = data.pop("default_lang")

        if "supported_langs" in data:
            data["supported_languages"] = data.pop("supported_langs")

        # Ensure defaults
        data.setdefault("default_language", "en")
        data.setdefault("supported_languages", ["en"])
        data.setdefault("auto_detect", True)
        data.setdefault("fallback_language", "en")

        return data

    def _migrate_template_optimizer_1_0(self, data: Dict[str, Any]) -> Dict[str, Any]:
        """Migrate template optimizer from 1.0 format."""
        # Example transformation logic
        if "optimize_level" in data:
            data["optimization_level"] = data.pop("optimize_level")

        # Ensure defaults
        data.setdefault("enabled", True)
        data.setdefault("optimization_level", "basic")
        data.setdefault("minification", False)

        return data

    def _migrate_project_initializer_1_0(self, data: Dict[str, Any]) -> Dict[str, Any]:
        """Migrate project initializer from 1.0 format."""
        # Example transformation logic
        if "auto_install" in data:
            data["auto_dependencies"] = data.pop("auto_install")

        # Ensure defaults
        data.setdefault("auto_dependencies", True)
        data.setdefault("git_init", True)
        data.setdefault("create_virtual_env", True)

        return data

    def rollback_migration(self, target_timestamp: Optional[str] = None) -> bool:
        """
        Rollback migration to specific timestamp or latest backup.

        Args:
            target_timestamp: Target timestamp for rollback (optional)

        Returns:
            True if rollback successful, False otherwise
        """
        try:
            backups = self.config_manager.list_backups()

            if not backups:
                logger.error("No backups available for rollback")
                return False

            # Find target backup
            if target_timestamp:
                target_backup = None
                for backup in backups:
                    if target_timestamp in backup:
                        target_backup = backup
                        break

                if not target_backup:
                    logger.error(f"Backup with timestamp {target_timestamp} not found")
                    return False
            else:
                target_backup = backups[0]  # Use latest backup

            # Restore backup
            self.config_manager.restore_backup(target_backup)
            logger.info(f"Successfully rolled back to {target_backup}")
            return True

        except Exception as e:
            logger.error(f"Rollback failed: {e}")
            return False

    def preview_migration(self) -> Dict[str, Any]:
        """
        Preview what would be migrated without actually doing it.

        Returns:
            Migration preview information
        """
        # Create temporary migrator in dry-run mode
        preview_migrator = ConfigurationMigrator(self.config_manager, dry_run=True)

        # Run migration detection
        return preview_migrator.detect_and_migrate(backup=False)


# Utility functions
def run_migration(config_dir: Union[str, Path], backup: bool = True, dry_run: bool = False) -> Dict[str, Any]:
    """
    Run complete migration process.

    Args:
        config_dir: Directory containing configuration files
        backup: Whether to create backup before migration
        dry_run: If True, only preview what would be migrated

    Returns:
        Migration result summary
    """
    config_dir = Path(config_dir)
    config_manager = UnifiedConfigManager(config_dir / "config.json")
    migrator = ConfigurationMigrator(config_manager, dry_run=dry_run)

    return migrator.detect_and_migrate(backup=backup)


def check_migration_status(config_dir: Union[str, Path]) -> Dict[str, Any]:
    """
    Check migration status for configuration directory.

    Args:
        config_dir: Directory to check

    Returns:
        Migration status information
    """
    config_dir = Path(config_dir)

    # Check for unified config
    unified_config_path = config_dir / "config.json"
    has_unified_config = unified_config_path.exists()

    # Check for legacy configs
    detector = LegacyConfigDetector([config_dir])
    legacy_configs = detector.find_legacy_configs()

    return {
        "has_unified_config": has_unified_config,
        "legacy_configs_found": len(legacy_configs),
        "legacy_configs": [
            {
                "path": str(c.path),
                "module": c.module_name,
                "version": c.version,
                "format": c.format,
            }
            for c in legacy_configs
        ],
        "migration_needed": len(legacy_configs) > 0,
    }
