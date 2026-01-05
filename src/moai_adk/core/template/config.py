"""Configuration Manager (Migrated to Unified Config).

⚠️ DEPRECATED: This module is deprecated and will be removed in a future version.

New code should use: from moai_adk.core.config.unified import UnifiedConfigManager, get_unified_config

Migration Guide:
    OLD:
        >>> from moai_adk.core.template.config import ConfigManager
        >>> config = ConfigManager()

    NEW:
        >>> from moai_adk.core.config.unified import UnifiedConfigManager
        >>> config = UnifiedConfigManager()
        >>> data = config.get_all()

Configuration sources (in priority order):
    1. Section YAML files: .moai/config/sections/*.yaml (new approach)
       - language.yaml: Language settings
       - user.yaml: User settings
       - project.yaml: Project metadata
       - system.yaml: MoAI system settings
       - git-strategy.yaml: Git workflow settings
       - quality.yaml: Quality and TDD settings
    2. Monolithic config.yaml: .moai/config/config.yaml (legacy)
    3. Monolithic config.json: .moai/config/config.json (legacy fallback)

Migration Path:
    OLD:
        >>> from moai_adk.core.template.config import ConfigManager
        >>> config_mgr = ConfigManager(Path(".moai/config/config.json"))
        >>> data = config_mgr.load()

    NEW:
        >>> from moai_adk.core.config.unified import get_unified_config
        >>> config = get_unified_config()
        >>> data = config.get_all()

Status: Deprecated - Maintained for backward compatibility only
Removal Target: v10.0.0
Last Updated: 2026-01-06
"""

import json
from pathlib import Path
from typing import Any

# Import unified config manager
try:
    from moai_adk.core.config.unified import UnifiedConfigManager, get_unified_config

    UNIFIED_AVAILABLE = True
except ImportError:
    UNIFIED_AVAILABLE = False


class ConfigManager:
    """
    Read and write MoAI configuration.

    DEPRECATED: Use moai_adk.core.config.unified.UnifiedConfigManager instead.
    This class is now a thin wrapper around UnifiedConfigManager for backward compatibility.

    Supports both:
    - New: Section YAML files (.moai/config/sections/*.yaml)
    - Legacy: Monolithic config.json (.moai/config/config.json)
    """

    DEFAULT_CONFIG = {"mode": "personal", "locale": "ko", "moai": {"version": "0.3.0"}}

    def __init__(self, config_path: Path) -> None:
        """
        Initialize the ConfigManager.

        Args:
            config_path: Path to config.json.
        """
        self.config_path = config_path

        # Use unified config if available
        if UNIFIED_AVAILABLE:
            self._unified_config = UnifiedConfigManager(config_path)
        else:
            self._unified_config = None

    def load(self) -> dict[str, Any]:
        """
        Load the configuration file.

        Returns default values when the file is missing.

        Returns:
            Configuration dictionary.
        """
        # Use unified config if available
        if self._unified_config:
            return self._unified_config.get_all()

        # Fallback to direct file access
        if not self.config_path.exists():
            return self.DEFAULT_CONFIG.copy()

        with open(self.config_path, encoding="utf-8") as f:
            data: dict[str, Any] = json.load(f)
            return data

    def save(self, config: dict[str, Any]) -> None:
        """
        Persist the configuration file.

        Creates directories when missing and preserves UTF-8 content.

        Args:
            config: Configuration dictionary to save.
        """
        # Use unified config if available
        if self._unified_config:
            self._unified_config.update(config, deep_merge=False)
            self._unified_config.save(backup=True)
            return

        # Fallback to direct file access
        self.config_path.parent.mkdir(parents=True, exist_ok=True)
        with open(self.config_path, "w", encoding="utf-8") as f:
            json.dump(config, f, ensure_ascii=False, indent=2)

    def update(self, updates: dict[str, Any]) -> None:
        """
        Update the configuration using a deep merge.

        Args:
            updates: Dictionary of updates to apply.
        """
        # Use unified config if available
        if self._unified_config:
            self._unified_config.update(updates, deep_merge=True)
            self._unified_config.save(backup=True)
            return

        # Fallback to direct implementation
        current = self.load()
        merged = self._deep_merge(current, updates)
        self.save(merged)

    def _deep_merge(self, base: dict[str, Any], updates: dict[str, Any]) -> dict[str, Any]:
        """
        Recursively deep-merge dictionaries.

        Args:
            base: Base dictionary.
            updates: Dictionary with updates.

        Returns:
            Merged dictionary.
        """
        result = base.copy()
        for key, value in updates.items():
            if key in result and isinstance(result[key], dict) and isinstance(value, dict):
                result[key] = self._deep_merge(result[key], value)
            else:
                result[key] = value
        return result

    @staticmethod
    def set_optimized_field(config_path: Path, field_path: str, value: Any) -> None:
        """
        Optimized method for setting a single configuration field.

        Reads only once, modifies in memory, writes atomically with minimal overhead.
        Used by optimization processes to update fields like project.optimized.

        Args:
            config_path: Path to config file.
            field_path: Dot-separated field path (e.g., "project.optimized").
            value: Value to set.

        Example:
            >>> ConfigManager.set_optimized_field(
            ...     Path(".moai/config/config.json"),
            ...     "project.optimized",
            ...     True
            ... )
        """
        # Use unified config if available
        if UNIFIED_AVAILABLE:
            unified_config = get_unified_config(config_path)
            unified_config.set(field_path, value)
            unified_config.save(backup=False)  # Skip backup for optimization flag
            return

        # Fallback implementation
        config_path.parent.mkdir(parents=True, exist_ok=True)

        # Read current config
        config_dict: dict[str, Any]
        if config_path.exists():
            with open(config_path, "r", encoding="utf-8") as f:
                config_dict = json.load(f)
        else:
            config_dict = {}

        # Navigate to target field
        keys = field_path.split(".")
        current: dict[str, Any] = config_dict
        for key in keys[:-1]:
            if key not in current:
                current[key] = {}
            current = current[key]

        # Set value
        current[keys[-1]] = value

        # Atomic write
        with open(config_path, "w", encoding="utf-8") as f:
            json.dump(config_dict, f, ensure_ascii=False, indent=2)
