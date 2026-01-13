#!/usr/bin/env python3
"""Configuration Manager for Alfred Hooks

Provides centralized configuration management with fallbacks and validation.
"""

import json
from pathlib import Path
from typing import Any, Dict

try:
    import yaml

    YAML_AVAILABLE = True
except ImportError:
    YAML_AVAILABLE = False

from .atomic_write import atomic_write_json
from .path_utils import find_project_root

# Default configuration
DEFAULT_CONFIG = {
    "hooks": {
        "timeout_seconds": 5,
        "timeout_ms": 5000,
        "minimum_timeout_seconds": 1,
        "graceful_degradation": True,
        "exit_codes": {
            "success": 0,
            "error": 1,
            "critical_error": 2,
            "config_error": 3,
        },
        "messages": {
            "timeout": {
                "post_tool_use": "⚠️ PostToolUse timeout - continuing",
                "session_end": "⚠️ SessionEnd cleanup timeout - session ending anyway",
                "session_start": "⚠️ Session start timeout - continuing without project info",
            },
            "stderr": {
                "timeout": {
                    "post_tool_use": "PostToolUse hook timeout after 5 seconds",
                    "session_end": "SessionEnd hook timeout after 5 seconds",
                    "session_start": "SessionStart hook timeout after 5 seconds",
                }
            },
            "config": {
                "missing": "❌ Project configuration not found - run /moai:0-project",
                "missing_fields": "⚠️ Missing configuration:",
            },
        },
        "cache": {
            "directory": ".moai/cache",
            "version_ttl_seconds": 1800,
            "git_ttl_seconds": 10,
        },
        "git": {
            "timeout_seconds": 2,
        },
        "project_search": {"max_depth": 10},
        "network": {"test_host": "8.8.8.8", "test_port": 53, "timeout_seconds": 0.1},
        "version_check": {
            "pypi_url": "https://pypi.org/pypi/moai-adk/json",
            "timeout_seconds": 1,
            "cache_ttl_seconds": 86400,
        },
    },
    "language": {
        "conversation_language": "en",
    },
}


class ConfigManager:
    """Configuration manager for Alfred hooks with validation and fallbacks."""

    def __init__(self, config_path: Path | None = None):
        """Initialize configuration manager.

        Args:
            config_path: Path to configuration file (defaults to .moai/config/config.yaml or config.json)
        """
        if config_path:
            self.config_path = config_path
        else:
            # Auto-detect YAML (preferred) or JSON (fallback)
            project_root = find_project_root()
            yaml_path = project_root / ".moai" / "config" / "config.yaml"
            json_path = project_root / ".moai" / "config" / "config.json"

            if YAML_AVAILABLE and yaml_path.exists():
                self.config_path = yaml_path
            elif json_path.exists():
                self.config_path = json_path
            else:
                # Default to YAML for new projects
                self.config_path = yaml_path if YAML_AVAILABLE else json_path

        self._config: Dict[str, Any] | None = None

    def load_config(self) -> Dict[str, Any]:
        """Load configuration from file with fallback to defaults.

        Returns:
            Merged configuration dictionary
        """
        if self._config is not None:
            return self._config

        # Load from file if exists
        config = {}
        if self.config_path.exists():
            try:
                with open(self.config_path, "r", encoding="utf-8") as f:
                    if self.config_path.suffix in [".yaml", ".yml"]:
                        if not YAML_AVAILABLE:
                            # Fall back to defaults if YAML not available
                            config = DEFAULT_CONFIG.copy()
                        else:
                            file_config = yaml.safe_load(f) or {}
                            config = self._merge_configs(DEFAULT_CONFIG.copy(), file_config)
                    else:
                        file_config = json.load(f)
                        config = self._merge_configs(DEFAULT_CONFIG.copy(), file_config)
            except (json.JSONDecodeError, IOError, OSError):
                # Use defaults if file is corrupted or unreadable
                config = DEFAULT_CONFIG.copy()
            except Exception as e:
                # Handle YAML errors or other parsing issues
                if YAML_AVAILABLE and isinstance(e, yaml.YAMLError):
                    config = DEFAULT_CONFIG.copy()
                else:
                    # Re-raise unexpected exceptions
                    raise
        else:
            # Use defaults if file doesn't exist
            config = DEFAULT_CONFIG.copy()

        self._config = config
        return config

    def get(self, key_path: str, default: Any = None) -> Any:
        """Get configuration value using dot notation.

        Args:
            key_path: Dot-separated path to configuration value
            default: Default value if key not found

        Returns:
            Configuration value or default
        """
        if not key_path:
            return default

        config = self.load_config()
        keys = key_path.split(".")
        current = config

        for key in keys:
            if isinstance(current, dict) and key in current:
                current = current[key]
            else:
                return default

        return current

    def get_hooks_config(self) -> Dict[str, Any]:
        """Get hooks-specific configuration.

        Returns:
            Hooks configuration dictionary
        """
        return self.get("hooks", {})

    def get_cache_config(self) -> Dict[str, Any]:
        """Get cache configuration.

        Returns:
            Cache configuration dictionary
        """
        return self.get("hooks.cache", {})

    def get_project_search_config(self) -> Dict[str, Any]:
        """Get project search configuration.

        Returns:
            Project search configuration dictionary
        """
        return self.get("hooks.project_search", {})

    def get_network_config(self) -> Dict[str, Any]:
        """Get network configuration.

        Returns:
            Network configuration dictionary
        """
        return self.get("hooks.network", {})

    def get_git_config(self) -> Dict[str, Any]:
        """Get git configuration.

        Returns:
            Git configuration dictionary
        """
        return self.get("hooks.git", {})

    def get_language_config(self) -> Dict[str, Any]:
        """Get language configuration.

        Returns:
            Language configuration dictionary
        """
        return self.get("language", {})

    def get_message(self, category: str, subcategory: str = "", key: str = "") -> str:
        """Get message from messages configuration.

        Args:
            category: Message category (e.g., 'stderr', 'timeout')
            subcategory: Sub-category (e.g., 'timeout', 'config')
            key: Message key (e.g., 'post_tool_use')

        Returns:
            Message string
        """
        if key:
            message = self.get(f"hooks.messages.{category}.{subcategory}.{key}")
            if message:
                return str(message)

        if subcategory:
            message = self.get(f"hooks.messages.{category}.{subcategory}")
            if isinstance(message, dict):
                # Try to find any string value in the dict
                for v in message.values():
                    if isinstance(v, str):
                        return v
            elif message:
                return str(message)

        message = self.get(f"hooks.messages.{category}")
        if message:
            return str(message)

        return "An error occurred"

    def get_timeout_seconds(self, hook_type: str = "default") -> int:
        """Get timeout seconds for a specific hook type.

        Args:
            hook_type: Type of hook (default, git, network, version_check)

        Returns:
            Timeout seconds
        """
        if hook_type == "git":
            return self.get("hooks.git.timeout_seconds", 2)
        elif hook_type == "network":
            return self.get("hooks.network.timeout_seconds", 0.1)
        elif hook_type == "version_check":
            return self.get("hooks.version_check.timeout_seconds", 1)
        else:
            return self.get("hooks.timeout_seconds", 5)

    def get_timeout_ms(self) -> int:
        """Get timeout milliseconds for hooks.

        Returns:
            Timeout milliseconds
        """
        return self.get("hooks.timeout_ms", 5000)

    def get_minimum_timeout_seconds(self) -> int:
        """Get minimum allowed timeout seconds.

        Returns:
            Minimum timeout seconds
        """
        return self.get("hooks.minimum_timeout_seconds", 1)

    def get_graceful_degradation(self) -> bool:
        """Get graceful degradation setting.

        Returns:
            True if graceful degradation is enabled
        """
        return self.get("hooks.graceful_degradation", True)

    def get_exit_code(self, code_type: str = "success") -> int:
        """Get exit code for a specific type.

        Args:
            code_type: Type of exit code (success, error, critical_error, config_error)

        Returns:
            Exit code integer
        """
        exit_codes = self.get("hooks.exit_codes", {})
        return exit_codes.get(code_type, 0)

    def update_config(self, updates: dict[str, Any]) -> bool:
        """Update configuration and save to file.

        Args:
            updates: Dictionary of updates to merge

        Returns:
            True if successful, False otherwise
        """
        try:
            # Load current config
            current = self.load_config()

            # Merge updates
            updated = self._merge_configs(current, updates)

            # Write updated config using atomic write (H3)
            # atomic_write_json will create parent directories if needed
            atomic_write_json(self.config_path, updated, indent=2)

            # Clear cache to force reload
            self._config = None

            return True
        except (IOError, OSError, json.JSONDecodeError):
            return False

    def validate_config(self) -> bool:
        """Validate configuration structure.

        Returns:
            True if configuration is valid, False otherwise
        """
        config = self.load_config()

        # Check that hooks exists and is a dict
        if not isinstance(config.get("hooks"), dict):
            return False

        return True

    def _merge_configs(self, base: dict[str, Any], override: dict[str, Any]) -> Dict[str, Any]:
        """Recursively merge two configuration dictionaries.

        This method delegates to common.merge_configs() for consistent behavior.

        Args:
            base: Base configuration dictionary
            override: Override configuration dictionary

        Returns:
            Merged configuration dictionary
        """
        from .common import merge_configs

        return merge_configs(base, override)


# Module-level helper functions
_config_manager: ConfigManager | None = None


def get_config_manager(config_path: Path | None = None) -> ConfigManager:
    """Get or create a ConfigManager instance.

    Args:
        config_path: Optional path to configuration file

    Returns:
        ConfigManager instance
    """
    global _config_manager
    if _config_manager is None or config_path is not None:
        _config_manager = ConfigManager(config_path)
    return _config_manager


def get_config(key_path: str = "", config_path: Path | None = None) -> Any:
    """Get configuration value.

    Args:
        key_path: Optional dot-separated path to configuration value
        config_path: Optional path to configuration file

    Returns:
        Configuration value or dictionary
    """
    cm = get_config_manager(config_path)
    if key_path:
        return cm.get(key_path)
    return cm.load_config()


def get_timeout_seconds(hook_type: str = "default") -> int:
    """Get timeout seconds for a hook type.

    Args:
        hook_type: Type of hook

    Returns:
        Timeout seconds
    """
    return get_config_manager().get_timeout_seconds(hook_type)


def get_timeout_ms() -> int:
    """Get timeout milliseconds.

    Returns:
        Timeout milliseconds
    """
    return get_config_manager().get_timeout_ms()


def get_minimum_timeout_seconds() -> int:
    """Get minimum allowed timeout seconds.

    Returns:
        Minimum timeout seconds
    """
    return get_config_manager().get_minimum_timeout_seconds()


def get_graceful_degradation() -> bool:
    """Get graceful degradation setting.

    Returns:
        True if enabled
    """
    return get_config_manager().get_graceful_degradation()


def get_exit_code(code_type: str = "success") -> int:
    """Get exit code for a type.

    Args:
        code_type: Type of exit code

    Returns:
        Exit code
    """
    return get_config_manager().get_exit_code(code_type)
