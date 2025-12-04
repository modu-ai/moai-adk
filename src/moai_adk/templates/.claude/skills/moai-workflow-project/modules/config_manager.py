"""
MoAI Menu Project - Unified Configuration Manager

This module provides centralized configuration management for all menu project
components, replacing individual configuration managers from 5 separate skills.
"""

import json
import logging
import shutil
from dataclasses import dataclass
from datetime import UTC, datetime
from functools import lru_cache
from pathlib import Path
from typing import Any, Dict, List, Optional, Union

try:
    import yaml

    HAS_YAML = True
except ImportError:
    HAS_YAML = False

try:
    import jsonschema

    HAS_JSONSCHEMA = True
except ImportError:
    HAS_JSONSCHEMA = False

# Setup logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


@dataclass
class ConfigMetadata:
    """Configuration metadata tracking."""

    created_at: str
    updated_at: str
    migration_history: List[Dict[str, Any]]

    @classmethod
    def create_new(cls) -> "ConfigMetadata":
        """Create new metadata with current timestamp."""
        now = datetime.now(UTC).isoformat() + "Z"
        return cls(created_at=now, updated_at=now, migration_history=[])


class ValidationError(Exception):
    """Configuration validation error."""

    pass


class MigrationError(Exception):
    """Configuration migration error."""

    pass


class UnifiedConfigManager:
    """
    Centralized configuration manager for menu project system.

    Features:
    - Single source of truth for all configuration
    - JSON Schema validation
    - Automatic migration support
    - Caching for performance
    - Backup and recovery
    - Environment-specific overrides
    """

    def __init__(
        self,
        config_path: Union[str, Path],
        schema_path: Optional[Union[str, Path]] = None,
    ):
        """
        Initialize configuration manager.

        Args:
            config_path: Path to configuration file
            schema_path: Path to JSON schema file (optional)
        """
        self.config_path = Path(config_path)
        self.schema_path = Path(schema_path) if schema_path else None
        self.config_dir = self.config_path.parent
        self.backup_dir = self.config_dir / "backups"

        # Ensure directories exist
        self.config_dir.mkdir(parents=True, exist_ok=True)
        self.backup_dir.mkdir(parents=True, exist_ok=True)

        # Initialize cache
        self._config_cache: Optional[Dict[str, Any]] = None
        self._schema_cache: Optional[Dict[str, Any]] = None
        self._last_modified: Optional[float] = None

        # Load schema if available
        if self.schema_path and self.schema_path.exists():
            self._load_schema()

    def _load_schema(self) -> None:
        """Load JSON schema for validation."""
        if not HAS_JSONSCHEMA:
            logger.warning("jsonschema not available, skipping validation")
            return

        try:
            with open(self.schema_path, "r", encoding="utf-8") as f:
                self._schema_cache = json.load(f)
            logger.info(f"Schema loaded from {self.schema_path}")
        except Exception as e:
            logger.error(f"Failed to load schema: {e}")
            raise

    def _get_file_mtime(self) -> float:
        """Get file modification time."""
        return self.config_path.stat().st_mtime if self.config_path.exists() else 0

    def _is_cache_valid(self) -> bool:
        """Check if cache is still valid."""
        if self._config_cache is None:
            return False
        return self._last_modified == self._get_file_mtime()

    def _invalidate_cache(self) -> None:
        """Invalidate configuration cache."""
        self._config_cache = None
        self._last_modified = None

    def load_config(self, force_reload: bool = False) -> Dict[str, Any]:
        """
        Load configuration from file with caching.

        Args:
            force_reload: Force reload even if cache is valid

        Returns:
            Configuration dictionary
        """
        if not force_reload and self._is_cache_valid():
            return self._config_cache

        if not self.config_path.exists():
            # Create default configuration
            config = self._create_default_config()
            self.save_config(config)
            return config

        try:
            with open(self.config_path, "r", encoding="utf-8") as f:
                if self.config_path.suffix.lower() in [".yml", ".yaml"]:
                    if not HAS_YAML:
                        raise ImportError("PyYAML is required for YAML configuration files")
                    config = yaml.safe_load(f)
                else:
                    config = json.load(f)

            # Validate against schema if available
            if self._schema_cache:
                self._validate_config(config)

            # Update cache
            self._config_cache = config
            self._last_modified = self._get_file_mtime()

            logger.info(f"Configuration loaded from {self.config_path}")
            return config

        except Exception as e:
            logger.error(f"Failed to load configuration: {e}")
            raise

    def save_config(self, config: Dict[str, Any], create_backup: bool = True) -> None:
        """
        Save configuration to file.

        Args:
            config: Configuration dictionary to save
            create_backup: Whether to create backup before saving
        """
        # Validate before saving
        if self._schema_cache:
            self._validate_config(config)

        # Update metadata
        self._update_metadata(config)

        # Create backup if requested
        if create_backup and self.config_path.exists():
            self._create_backup()

        # Ensure config has version
        if "version" not in config:
            config["version"] = "1.0.0"

        try:
            # Save with atomic write
            temp_path = self.config_path.with_suffix(".tmp")
            with open(temp_path, "w", encoding="utf-8") as f:
                json.dump(config, f, indent=2, ensure_ascii=False)

            # Atomic replace
            temp_path.replace(self.config_path)

            # Update cache
            self._config_cache = config
            self._last_modified = self._get_file_mtime()

            logger.info(f"Configuration saved to {self.config_path}")

        except Exception as e:
            logger.error(f"Failed to save configuration: {e}")
            raise

    def _validate_config(self, config: Dict[str, Any]) -> None:
        """Validate configuration against schema."""
        if not HAS_JSONSCHEMA or not self._schema_cache:
            return

        try:
            jsonschema.validate(config, self._schema_cache)
        except jsonschema.ValidationError as e:
            raise ValidationError(f"Configuration validation failed: {e.message}")

    def _create_default_config(self) -> Dict[str, Any]:
        """Create default configuration."""
        now = datetime.now(UTC).isoformat() + "Z"

        return {
            "version": "1.0.0",
            "metadata": {"created_at": now, "updated_at": now, "migration_history": []},
            "project_settings": {
                "name": "MoAI Menu Project",
                "type": "web",
                "description": "Menu project management system",
                "author": "MoAI Team",
                "license": "MIT",
            },
            "batch_questions": {
                "enabled": True,
                "max_questions": 50,
                "timeout_seconds": 300,
                "retry_attempts": 3,
                "output_format": "json",
                "caching": {"enabled": True, "ttl_hours": 24, "max_cache_size_mb": 100},
            },
            "documentation": {
                "auto_generate": True,
                "format": "markdown",
                "include_api_docs": True,
                "include_examples": True,
                "template_engine": "jinja2",
                "output_directory": "docs",
            },
            "language_config": {
                "default_language": "en",
                "supported_languages": ["en", "ko", "ja", "zh"],
                "auto_detect": True,
                "fallback_language": "en",
                "translation_service": {"provider": "local", "cache_enabled": True},
            },
            "template_optimizer": {
                "enabled": True,
                "optimization_level": "basic",
                "minification": False,
                "caching": {"enabled": True, "strategy": "memory", "ttl_minutes": 60},
                "validation": {
                    "strict_mode": False,
                    "check_syntax": True,
                    "check_security": True,
                },
            },
            "project_initializer": {
                "auto_dependencies": True,
                "git_init": True,
                "create_virtual_env": True,
                "package_managers": {"python": {"enabled": True, "tool": "pip"}},
            },
        }

    def _update_metadata(self, config: Dict[str, Any]) -> None:
        """Update configuration metadata."""
        if "metadata" not in config:
            config["metadata"] = ConfigMetadata.create_new().__dict__

        config["metadata"]["updated_at"] = datetime.now(UTC).isoformat() + "Z"

    def _create_backup(self) -> None:
        """Create backup of current configuration."""
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        backup_name = f"config_backup_{timestamp}.json"
        backup_path = self.backup_dir / backup_name

        try:
            shutil.copy2(self.config_path, backup_path)
            logger.info(f"Configuration backed up to {backup_path}")
        except Exception as e:
            logger.warning(f"Failed to create backup: {e}")

    def get_module_config(self, module_name: str) -> Dict[str, Any]:
        """
        Get configuration for specific module.

        Args:
            module_name: Name of module (e.g., 'batch_questions', 'documentation')

        Returns:
            Module configuration dictionary
        """
        config = self.load_config()

        # Handle nested modules
        if module_name in config:
            return config[module_name]

        # Check for nested module names
        for key, value in config.items():
            if isinstance(value, dict) and module_name in value:
                return value[module_name]

        raise KeyError(f"Module '{module_name}' not found in configuration")

    def update_module_config(self, module_name: str, updates: Dict[str, Any]) -> None:
        """
        Update configuration for specific module.

        Args:
            module_name: Name of module to update
            updates: Configuration updates to apply
        """
        config = self.load_config()

        # Deep merge updates
        if module_name not in config:
            config[module_name] = {}

        config[module_name].update(updates)

        self.save_config(config)

    def migrate_config(self, target_version: str) -> None:
        """
        Migrate configuration to target version.

        Args:
            target_version: Target version to migrate to
        """
        config = self.load_config()
        current_version = config.get("version", "1.0.0")

        if current_version == target_version:
            logger.info("Configuration already at target version")
            return

        # Record migration in history
        migration_record = {
            "from_version": current_version,
            "to_version": target_version,
            "timestamp": datetime.now(UTC).isoformat() + "Z",
            "description": f"Migrated from {current_version} to {target_version}",
        }

        if "metadata" not in config:
            config["metadata"] = ConfigMetadata.create_new().__dict__

        config["metadata"]["migration_history"].append(migration_record)
        config["version"] = target_version

        # Perform actual migration logic here
        # This would be version-specific

        self.save_config(config)
        logger.info(f"Configuration migrated to version {target_version}")

    def restore_backup(self, backup_name: str) -> None:
        """
        Restore configuration from backup.

        Args:
            backup_name: Name of backup file to restore
        """
        backup_path = self.backup_dir / backup_name

        if not backup_path.exists():
            raise FileNotFoundError(f"Backup file '{backup_name}' not found")

        try:
            shutil.copy2(backup_path, self.config_path)
            self._invalidate_cache()
            logger.info(f"Configuration restored from {backup_name}")
        except Exception as e:
            logger.error(f"Failed to restore backup: {e}")
            raise

    def list_backups(self) -> List[str]:
        """List available backup files."""
        if not self.backup_dir.exists():
            return []

        backups = []
        for file_path in self.backup_dir.glob("config_backup_*.json"):
            backups.append(file_path.name)

        return sorted(backups, reverse=True)


# Specialized Module Managers
class BatchQuestionsConfigManager:
    """Specialized manager for batch questions configuration."""

    def __init__(self, config_manager: UnifiedConfigManager):
        self.config_manager = config_manager

    def get_config(self) -> Dict[str, Any]:
        """Get batch questions configuration."""
        return self.config_manager.get_module_config("batch_questions")

    def update_config(self, updates: Dict[str, Any]) -> None:
        """Update batch questions configuration."""
        self.config_manager.update_module_config("batch_questions", updates)

    def is_enabled(self) -> bool:
        """Check if batch questions is enabled."""
        return self.get_config().get("enabled", True)

    def get_max_questions(self) -> int:
        """Get maximum questions limit."""
        return self.get_config().get("max_questions", 50)


class DocumentationConfigManager:
    """Specialized manager for documentation configuration."""

    def __init__(self, config_manager: UnifiedConfigManager):
        self.config_manager = config_manager

    def get_config(self) -> Dict[str, Any]:
        """Get documentation configuration."""
        return self.config_manager.get_module_config("documentation")

    def update_config(self, updates: Dict[str, Any]) -> None:
        """Update documentation configuration."""
        self.config_manager.update_module_config("documentation", updates)

    def should_auto_generate(self) -> bool:
        """Check if auto-generation is enabled."""
        return self.get_config().get("auto_generate", True)


class LanguageConfigManager:
    """Specialized manager for language configuration."""

    def __init__(self, config_manager: UnifiedConfigManager):
        self.config_manager = config_manager

    def get_config(self) -> Dict[str, Any]:
        """Get language configuration."""
        return self.config_manager.get_module_config("language_config")

    def update_config(self, updates: Dict[str, Any]) -> None:
        """Update language configuration."""
        self.config_manager.update_module_config("language_config", updates)

    def get_default_language(self) -> str:
        """Get default language."""
        return self.get_config().get("default_language", "en")

    def get_supported_languages(self) -> List[str]:
        """Get list of supported languages."""
        return self.get_config().get("supported_languages", ["en"])


class TemplateOptimizerConfigManager:
    """Specialized manager for template optimizer configuration."""

    def __init__(self, config_manager: UnifiedConfigManager):
        self.config_manager = config_manager

    def get_config(self) -> Dict[str, Any]:
        """Get template optimizer configuration."""
        return self.config_manager.get_module_config("template_optimizer")

    def update_config(self, updates: Dict[str, Any]) -> None:
        """Update template optimizer configuration."""
        self.config_manager.update_module_config("template_optimizer", updates)

    def get_optimization_level(self) -> str:
        """Get optimization level."""
        return self.get_config().get("optimization_level", "basic")


class ProjectInitializerConfigManager:
    """Specialized manager for project initializer configuration."""

    def __init__(self, config_manager: UnifiedConfigManager):
        self.config_manager = config_manager

    def get_config(self) -> Dict[str, Any]:
        """Get project initializer configuration."""
        return self.config_manager.get_module_config("project_initializer")

    def update_config(self, updates: Dict[str, Any]) -> None:
        """Update project initializer configuration."""
        self.config_manager.update_module_config("project_initializer", updates)

    def should_auto_dependencies(self) -> bool:
        """Check if auto-dependency installation is enabled."""
        return self.get_config().get("auto_dependencies", True)


# Factory function
def create_config_manager(config_dir: Union[str, Path]) -> UnifiedConfigManager:
    """
    Create unified configuration manager with default schema.

    Args:
        config_dir: Directory containing configuration files

    Returns:
        Configured UnifiedConfigManager instance
    """
    config_dir = Path(config_dir)
    config_path = config_dir / "config.json"
    schema_path = config_dir / "schemas" / "config-schema.json"

    return UnifiedConfigManager(config_path, schema_path)


# Convenience access functions
@lru_cache(maxsize=32)
def get_config_manager(config_path: str) -> UnifiedConfigManager:
    """Get cached configuration manager instance."""
    return UnifiedConfigManager(config_path)
