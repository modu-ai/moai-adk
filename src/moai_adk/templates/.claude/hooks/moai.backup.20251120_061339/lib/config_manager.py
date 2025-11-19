#!/usr/bin/env python3
"""Configuration Manager for Alfred Hooks

Provides centralized configuration management with fallbacks and validation.
"""

import json
from pathlib import Path
from typing import Any, Dict, Optional, cast

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
            "config_error": 3
        },
        "messages": {
            "timeout": {
                "post_tool_use": "⚠️ PostToolUse timeout - continuing",
                "session_end": "⚠️ SessionEnd cleanup timeout - session ending anyway",
                "session_start": "⚠️ Session start timeout - continuing without project info"
            },
            "stderr": {
                "timeout": {
                    "post_tool_use": "PostToolUse hook timeout after 5 seconds",
                    "session_end": "SessionEnd hook timeout after 5 seconds",
                    "session_start": "SessionStart hook timeout after 5 seconds"
                }
            },
            "config": {
                "missing": "❌ Project configuration not found - run /alfred:0-project",
                "missing_fields": "⚠️ Missing configuration:"
            }
        },
        "cache": {
            "directory": ".moai/cache",
            "version_ttl_seconds": 1800,
            "git_ttl_seconds": 10
        },
        "project_search": {
            "max_depth": 10
        },
        "network": {
            "test_host": "8.8.8.8",
            "test_port": 53,
            "timeout_seconds": 0.1
        },
        "version_check": {
            "pypi_url": "https://pypi.org/pypi/moai-adk/json",
            "timeout_seconds": 1
        },
        "git": {
            "timeout_seconds": 2
        },
                "defaults": {
            "timeout_ms": 5000,
            "graceful_degradation": True
        }
    }
}


class ConfigManager:
    """Configuration manager for Alfred hooks with validation and fallbacks."""

    def __init__(self, config_path: Optional[Path] = None):
        """Initialize configuration manager.

        Args:
            config_path: Path to configuration file (defaults to .moai/config/config.json)
        """
        self.config_path = config_path or Path.cwd() / ".moai" / "config.json"
        self._config: Optional[Dict[str, Any]] = None

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
                with open(self.config_path, 'r', encoding='utf-8') as f:
                    file_config = json.load(f)
                    config = self._merge_configs(DEFAULT_CONFIG, file_config)
            except (json.JSONDecodeError, IOError, OSError):
                # Use defaults if file is corrupted or unreadable
                config = DEFAULT_CONFIG.copy()
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
        config = self.load_config()
        keys = key_path.split('.')
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
            Whether graceful degradation is enabled
        """
        return self.get("hooks.graceful_degradation", True)

    def get_message(self, category: str, subcategory: str, key: str) -> str:
        """Get localized message from configuration.

        Args:
            category: Message category (timeout, stderr, config)
            subcategory: Subcategory within category
            key: Message key

        Returns:
            Localized message
        """
        default_messages = cast(Dict[str, Any], DEFAULT_CONFIG["hooks"]["messages"])
        message = self.get(f"hooks.messages.{category}.{subcategory}.{key}")

        if message is None and category in default_messages:
            category_messages = default_messages[category]
            if isinstance(category_messages, dict) and subcategory in category_messages:
                subcategory_messages = category_messages[subcategory]
                if isinstance(subcategory_messages, dict):
                    message = subcategory_messages.get(key)

        if message is None:
            # Fallback to English default
            timeout_messages = default_messages.get("timeout", {})
            if isinstance(timeout_messages, dict):
                subcategory_messages = timeout_messages.get(subcategory, {})
                if isinstance(subcategory_messages, dict):
                    fallback = subcategory_messages.get(key)
                    message = fallback or f"Message not found: {category}.{subcategory}.{key}"
                else:
                    message = f"Message not found: {category}.{subcategory}.{key}"
            else:
                message = f"Message not found: {category}.{subcategory}.{key}"

        return message

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


    def get_exit_code(self, exit_type: str) -> int:
        """Get exit code for specific exit type.

        Args:
            exit_type: Type of exit (success, error, critical_error, config_error)

        Returns:
            Exit code
        """
        return self.get("hooks.exit_codes", {}).get(exit_type, 0)

    def update_config(self, updates: Dict[str, Any]) -> bool:
        """Update configuration with new values.

        Args:
            updates: Dictionary with configuration updates

        Returns:
            True if update was successful, False otherwise
        """
        try:
            current_config = self.load_config()
            updated_config = self._merge_configs(current_config, updates)

            # Ensure parent directory exists
            self.config_path.parent.mkdir(parents=True, exist_ok=True)

            with open(self.config_path, 'w', encoding='utf-8') as f:
                json.dump(updated_config, f, indent=2, ensure_ascii=False)

            self._config = updated_config
            return True
        except (IOError, OSError, json.JSONDecodeError):
            return False

    def validate_config(self) -> bool:
        """Validate current configuration.

        Returns:
            True if configuration is valid, False otherwise
        """
        try:
            config = self.load_config()

            # Check required top-level keys
            required_keys = ["hooks"]
            for key in required_keys:
                if key not in config:
                    return False

            # Check hooks structure
            hooks = config.get("hooks", {})
            if not isinstance(hooks, dict):
                return False

            return True
        except Exception:
            return False

    def _merge_configs(self, base: Dict[str, Any], updates: Dict[str, Any]) -> Dict[str, Any]:
        """Recursively merge configuration dictionaries.

        Args:
            base: Base configuration dictionary
            updates: Updates to apply

        Returns:
            Merged configuration dictionary
        """
        result = base.copy()

        for key, value in updates.items():
            if key in result and isinstance(result[key], dict) and isinstance(value, dict):
                result[key] = self._merge_configs(result[key], value)
            else:
                result[key] = value

        return result

    def get_language_config(self) -> Dict[str, Any]:
        """Get language configuration.

        Returns:
            Language configuration dictionary
        """
        return self.get("language", {"conversation_language": "en"})


# Global configuration manager instance
_config_manager = None


def get_config_manager(config_path: Optional[Path] = None) -> ConfigManager:
    """Get global configuration manager instance.

    Args:
        config_path: Path to configuration file

    Returns:
        Configuration manager instance
    """
    global _config_manager
    if _config_manager is None or config_path is not None:
        _config_manager = ConfigManager(config_path)
    return _config_manager


def get_config(key_path: str, default: Any = None) -> Any:
    """Get configuration value using dot notation.

    Args:
        key_path: Dot-separated path to configuration value
        default: Default value if key not found

    Returns:
        Configuration value or default
    """
    return get_config_manager().get(key_path, default)


# Convenience functions for common configuration values
def get_timeout_seconds(hook_type: str = "default") -> int:
    """Get timeout seconds for a specific hook type."""
    return get_config_manager().get_timeout_seconds(hook_type)


def get_graceful_degradation() -> bool:
    """Get graceful degradation setting."""
    return get_config_manager().get_graceful_degradation()


def get_exit_code(exit_type: str) -> int:
    """Get exit code for specific exit type."""
    return get_config_manager().get_exit_code(exit_type)
