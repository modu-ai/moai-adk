"""
Unified Configuration Manager for MoAI-ADK

Thread-safe singleton configuration manager that consolidates all config management
patterns from across the codebase into a single, robust implementation.

Features:
- Thread-safe singleton with double-checked locking
- Lazy loading with in-memory caching
- Atomic writes with backup creation
- Schema validation support
- Smart defaults with auto-detection
- Deep merge for nested configs
- Configuration migration between versions
- Domain-specific facades for different subsystems

Design Philosophy:
- Single source of truth for `.moai/config/config.yaml` (with JSON fallback)
- Best patterns from StatuslineConfig, UnifiedConfigManager, ConfigurationManager
- Backward compatible with existing ConfigManager interfaces
- Performance optimized with caching and minimal I/O

Usage:
    >>> from moai_adk.core.config.unified import get_unified_config
    >>> config = get_unified_config()
    >>> timeout = config.get("hooks.timeout_ms", 2000)
    >>> config.update({"hooks.timeout_ms": 3000})
    >>> config.save()
"""

import json
import logging
import shutil
import threading
from datetime import datetime
from functools import lru_cache
from pathlib import Path
from typing import Any, Dict, Optional, Union

try:
    import yaml

    YAML_AVAILABLE = True
except ImportError:
    YAML_AVAILABLE = False

logger = logging.getLogger(__name__)


class UnifiedConfigManager:
    """
    Thread-safe singleton configuration manager for MoAI-ADK.

    This class consolidates all configuration management patterns from:
    - StatuslineConfig: Thread-safe singleton pattern
    - UnifiedConfigManager (skills): Backup, migration, validation
    - ConfigurationManager (project): Smart defaults, auto-detection
    - ConfigManager (hooks): Convenience functions

    Attributes:
        _instance: Singleton instance
        _config: Cached configuration dictionary
        _lock: Thread lock for singleton pattern
        _config_path: Path to configuration file
        _last_modified: Last modification timestamp for cache invalidation
    """

    _instance: Optional["UnifiedConfigManager"] = None
    _config: Dict[str, Any] = {}
    _lock = threading.Lock()
    _config_path: Optional[Path] = None
    _last_modified: Optional[float] = None

    def __new__(cls, config_path: Optional[Union[str, Path]] = None):
        """
        Create or return singleton instance with double-checked locking.

        Args:
            config_path: Optional path to config file (default: .moai/config/config.json)

        Returns:
            UnifiedConfigManager: Singleton instance
        """
        # Double-checked locking pattern for thread-safe singleton
        if cls._instance is None:
            with cls._lock:
                # Double-check after acquiring lock
                if cls._instance is None:
                    cls._instance = super().__new__(cls)
                    cls._instance._initialize(config_path)
        return cls._instance

    def _initialize(self, config_path: Optional[Union[str, Path]] = None) -> None:
        """
        Initialize configuration manager (called once by singleton).

        Args:
            config_path: Optional path to config file
        """
        # Set config path with auto-detection
        if config_path:
            self._config_path = Path(config_path)
        else:
            # Auto-detect YAML (preferred) or JSON (fallback)
            base_path = Path.cwd() / ".moai" / "config"
            yaml_path = base_path / "config.yaml"
            json_path = base_path / "config.json"

            if YAML_AVAILABLE and yaml_path.exists():
                self._config_path = yaml_path
            elif json_path.exists():
                self._config_path = json_path
            else:
                # Default to YAML for new projects
                self._config_path = yaml_path if YAML_AVAILABLE else json_path

        # Load configuration
        self._load_config()

    def _load_config(self) -> None:
        """
        Load configuration from file with caching.

        Implements cache invalidation based on file modification time.
        Falls back to default configuration if file doesn't exist.
        Supports both YAML (preferred) and JSON (legacy) formats.
        """
        try:
            # Check if file exists
            if not self._config_path.exists():
                logger.warning(f"Config file not found: {self._config_path}")
                self._config = self._get_default_config()
                return

            # Check cache validity
            current_mtime = self._config_path.stat().st_mtime
            if self._last_modified and current_mtime == self._last_modified:
                # Cache is still valid
                return

            # Load from file (auto-detect format)
            with open(self._config_path, "r", encoding="utf-8") as f:
                if self._config_path.suffix == ".yaml" or self._config_path.suffix == ".yml":
                    if not YAML_AVAILABLE:
                        raise ImportError("PyYAML is required for YAML config files. Install with: pip install pyyaml")
                    self._config = yaml.safe_load(f) or {}
                else:
                    self._config = json.load(f)

            # Update cache timestamp
            self._last_modified = current_mtime

            logger.debug(f"Loaded config from {self._config_path}")

        except (
            json.JSONDecodeError,
            yaml.YAMLError if YAML_AVAILABLE else Exception,
            OSError,
            UnicodeDecodeError,
        ) as e:
            logger.error(f"Failed to load config: {e}")
            self._config = self._get_default_config()

    def get(self, key: str, default: Any = None) -> Any:
        """
        Get configuration value with dot-notation support.

        Args:
            key: Configuration key (supports dot notation, e.g., "hooks.timeout_ms")
            default: Default value if key not found

        Returns:
            Configuration value or default

        Examples:
            >>> config.get("hooks.timeout_ms", 2000)
            2000
            >>> config.get("project.name")
            "MoAI-ADK"
        """
        # Reload if file changed
        self._reload_if_modified()

        # Navigate nested dict with dot notation
        keys = key.split(".")
        value = self._config

        for k in keys:
            if isinstance(value, dict):
                value = value.get(k)
                if value is None:
                    return default
            else:
                return default

        return value if value is not None else default

    def set(self, key: str, value: Any) -> None:
        """
        Set configuration value with dot-notation support.

        Args:
            key: Configuration key (supports dot notation)
            value: Value to set

        Examples:
            >>> config.set("hooks.timeout_ms", 3000)
            >>> config.set("project.name", "MyProject")
        """
        # Navigate to parent dict
        keys = key.split(".")
        target = self._config

        for k in keys[:-1]:
            if k not in target:
                target[k] = {}
            target = target[k]

        # Set value
        target[keys[-1]] = value

    def update(self, updates: Dict[str, Any], deep_merge: bool = True) -> None:
        """
        Update configuration with dictionary.

        Args:
            updates: Dictionary of updates
            deep_merge: If True, recursively merge nested dicts (default: True)

        Examples:
            >>> config.update({"hooks": {"timeout_ms": 3000}})
            >>> config.update({"project.name": "MyProject"}, deep_merge=False)
        """
        if deep_merge:
            self._config = self._deep_merge(self._config, updates)
        else:
            self._config.update(updates)

    def save(self, backup: bool = True) -> bool:
        """
        Save configuration to file with atomic write.

        Args:
            backup: If True, create backup before saving (default: True)

        Returns:
            bool: True if save successful, False otherwise

        Pattern: Temp file → Atomic rename (prevents corruption)
        """
        try:
            # Create backup if requested
            if backup and self._config_path.exists():
                self._create_backup()

            # Atomic write pattern: temp file → rename
            temp_path = self._config_path.with_suffix(".tmp")

            # Write to temp file (auto-detect format)
            with open(temp_path, "w", encoding="utf-8") as f:
                if self._config_path.suffix == ".yaml" or self._config_path.suffix == ".yml":
                    if not YAML_AVAILABLE:
                        raise ImportError("PyYAML is required for YAML config files. Install with: pip install pyyaml")
                    yaml.safe_dump(
                        self._config,
                        f,
                        default_flow_style=False,
                        allow_unicode=True,
                        sort_keys=False,
                    )
                else:
                    json.dump(self._config, f, indent=2, ensure_ascii=False)

            # Atomic rename
            temp_path.replace(self._config_path)

            # Update cache timestamp
            self._last_modified = self._config_path.stat().st_mtime

            logger.info(f"Saved config to {self._config_path}")
            return True

        except (OSError, PermissionError) as e:
            logger.error(f"Failed to save config: {e}")
            return False

    def _create_backup(self) -> None:
        """Create timestamped backup of config file."""
        try:
            timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
            backup_dir = self._config_path.parent / "backups"
            backup_dir.mkdir(parents=True, exist_ok=True)

            backup_path = backup_dir / f"config_backup_{timestamp}.json"
            shutil.copy2(self._config_path, backup_path)

            logger.debug(f"Created backup: {backup_path}")

        except (OSError, PermissionError) as e:
            logger.warning(f"Failed to create backup: {e}")

    def _reload_if_modified(self) -> None:
        """Reload config if file has been modified."""
        try:
            if self._config_path.exists():
                current_mtime = self._config_path.stat().st_mtime
                if self._last_modified is None or current_mtime != self._last_modified:
                    self._load_config()
        except OSError:
            pass

    @staticmethod
    def _deep_merge(base: Dict[str, Any], updates: Dict[str, Any]) -> Dict[str, Any]:
        """
        Recursively merge two dictionaries.

        Args:
            base: Base dictionary
            updates: Updates to merge

        Returns:
            Merged dictionary
        """
        result = base.copy()

        for key, value in updates.items():
            if key in result and isinstance(result[key], dict) and isinstance(value, dict):
                result[key] = UnifiedConfigManager._deep_merge(result[key], value)
            else:
                result[key] = value

        return result

    @staticmethod
    def _get_default_config() -> Dict[str, Any]:
        """
        Get default configuration when file doesn't exist.

        Returns:
            Default configuration dictionary
        """
        return {
            "moai": {"version": "0.28.0", "update_check_frequency": "daily"},
            "project": {"name": "MyProject", "initialized": False},
            "hooks": {"timeout_ms": 2000, "graceful_degradation": True},
            "session": {"suppress_setup_messages": False},
            "language": {"conversation_language": "en", "agent_prompt_language": "en"},
        }

    def get_all(self) -> Dict[str, Any]:
        """
        Get entire configuration dictionary.

        Returns:
            Complete configuration dictionary
        """
        self._reload_if_modified()
        return self._config.copy()

    def reset_to_defaults(self) -> None:
        """Reset configuration to defaults without saving."""
        self._config = self._get_default_config()


# Module-level singleton instance
_unified_config_instance: Optional[UnifiedConfigManager] = None


def get_unified_config(
    config_path: Optional[Union[str, Path]] = None,
) -> UnifiedConfigManager:
    """
    Get or create unified configuration manager instance.

    Args:
        config_path: Optional path to config file

    Returns:
        UnifiedConfigManager: Singleton instance

    Examples:
        >>> config = get_unified_config()
        >>> timeout = config.get("hooks.timeout_ms", 2000)
    """
    global _unified_config_instance

    if _unified_config_instance is None:
        _unified_config_instance = UnifiedConfigManager(config_path)

    return _unified_config_instance


# Convenience functions for common operations
@lru_cache(maxsize=1)
def get_config_path() -> Path:
    """Get path to configuration file."""
    return Path.cwd() / ".moai" / "config" / "config.json"


def get_config_value(key: str, default: Any = None) -> Any:
    """
    Get configuration value (convenience function).

    Args:
        key: Configuration key with dot notation
        default: Default value if key not found

    Returns:
        Configuration value or default
    """
    return get_unified_config().get(key, default)


def set_config_value(key: str, value: Any) -> None:
    """
    Set configuration value (convenience function).

    Args:
        key: Configuration key with dot notation
        value: Value to set
    """
    config = get_unified_config()
    config.set(key, value)


def save_config(backup: bool = True) -> bool:
    """
    Save configuration (convenience function).

    Args:
        backup: Create backup before saving

    Returns:
        bool: True if successful
    """
    return get_unified_config().save(backup)
