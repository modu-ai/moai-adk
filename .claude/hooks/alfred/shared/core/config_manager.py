#!/usr/bin/env python3
# @CODE:HOOKS-CONFIG-001 | SPEC: HOOKS-STANDARDIZATION-003
"""Enhanced Configuration Manager for Alfred Hooks

Provides centralized configuration management with:
- Thread-safe singleton pattern
- TTL-based caching (5 minutes)
- File change detection with mtime tracking
- Graceful degradation and error handling
- Performance optimization to eliminate 105 duplicate config loads
"""

import json
import os
import threading
import time
import weakref
from pathlib import Path
from typing import Any, Dict, Optional, Union


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
        "tag_rules": {
            "config_file_path": ".moai/tag-rules.json"
        },
        "tag_validation": {
            "default_code_extensions": [
                ".py", ".ts", ".tsx", ".js", ".jsx", ".go", ".rs", ".java", ".kt",
                ".rb", ".php", ".c", ".cpp", ".cs", ".swift", ".scala"
            ]
        },
        "defaults": {
            "timeout_ms": 5000,
            "graceful_degradation": True
        }
    }
}


class ConfigManager:
    """Thread-safe configuration manager with caching and file monitoring."""

    _instance = None
    _lock = threading.Lock()
    _instances = weakref.WeakSet()

    def __init__(self, config_path: Optional[Path] = None, ttl_seconds: int = 300):
        """Initialize configuration manager.

        Args:
            config_path: Path to configuration file (defaults to .moai/config.json)
            ttl_seconds: Time-to-live for cache in seconds (default: 5 minutes)
        """
        self.config_path = config_path or Path.cwd() / ".moai" / "config.json"
        self._ttl_seconds = ttl_seconds

        # Cache state
        self._config = None
        self._cache_timestamp = 0
        self._file_mtime = None

        # Thread safety
        self._config_lock = threading.RLock()

        # Track this instance for cleanup
        ConfigManager._instances.add(self)

    @classmethod
    def get_instance(cls, config_path: Optional[Path] = None) -> 'ConfigManager':
        """Get thread-safe singleton instance."""
        with cls._lock:
            if cls._instance is None:
                cls._instance = cls(config_path)
            # Update config_path if different path is provided
            elif config_path is not None:
                cls._instance.config_path = config_path or Path.cwd() / ".moai" / "config.json"
                cls._instance._config = None  # Reset cache for new path
            return cls._instance

    def __new__(cls, config_path: Optional[Path] = None, ttl_seconds: int = 300):
        """Implement singleton pattern with thread safety."""
        with cls._lock:
            if cls._instance is None:
                cls._instance = super().__new__(cls)
            return cls._instance

    def load_config(self) -> Dict[str, Any]:
        """Load configuration with caching and file change detection."""
        with self._config_lock:
            current_time = time.time()

            # Check if cache is valid
            if self._is_cache_valid(current_time):
                return self._config

            # Load configuration from file
            self._config = self._load_from_file()
            self._cache_timestamp = current_time

            # Update file modification time
            try:
                self._file_mtime = self.config_path.stat().st_mtime if self.config_path.exists() else None
            except (OSError, IOError):
                self._file_mtime = None

            return self._config

    def _is_cache_valid(self, current_time: float) -> bool:
        """Check if cached configuration is still valid."""
        if self._config is None:
            return False

        # Check TTL expiration
        if current_time - self._cache_timestamp > self._ttl_seconds:
            return False

        # Check file modification
        try:
            if self.config_path.exists():
                current_mtime = self.config_path.stat().st_mtime
                if self._file_mtime is None or current_mtime != self._file_mtime:
                    return False
            elif self._file_mtime is not None:
                # File was deleted
                return False
        except (OSError, IOError):
            # On error, assume cache is invalid
            return False

        return True

    def _load_from_file(self) -> Dict[str, Any]:
        """Load configuration from file with graceful error handling."""
        config = DEFAULT_CONFIG.copy()

        if not self.config_path.exists():
            return config

        try:
            with open(self.config_path, 'r', encoding='utf-8') as f:
                file_config = json.load(f)
                # Simple direct merge
                result = config.copy()
                self._deep_merge(result, file_config)
                return result
        except (json.JSONDecodeError, UnicodeDecodeError):
            # Invalid JSON or encoding - return defaults
            return config
        except (IOError, OSError, MemoryError):
            # File I/O errors - return defaults
            return config

    def _deep_merge(self, target: Dict[str, Any], source: Dict[str, Any]) -> None:
        """Recursively merge source into target."""
        for key, value in source.items():
            if key in target and isinstance(target[key], dict) and isinstance(value, dict):
                self._deep_merge(target[key], value)
            else:
                target[key] = value

    def _merge_configs(self, base: Dict[str, Any], updates: Dict[str, Any]) -> Dict[str, Any]:
        """Safely merge configuration dictionaries with validation."""
        result = base.copy()

        def _deep_merge(target: Dict[str, Any], source: Dict[str, Any]) -> None:
            """Recursively merge source into target with type validation."""
            for key, value in source.items():
                # Prevent circular references
                if isinstance(value, dict) and id(value) in [id(v) for v in source.values() if isinstance(v, dict)]:
                    continue

                if key in target and isinstance(target[key], dict) and isinstance(value, dict):
                    _deep_merge(target[key], value)
                else:
                    target[key] = value

        try:
            _deep_merge(result, updates)
        except (ValueError, RecursionError):
            # Circular reference or other merge error - return base
            pass

        return result

    def get(self, key_path: str, default: Any = None) -> Any:
        """Get configuration value using dot notation."""
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
        """Get hooks-specific configuration."""
        return self.get("hooks", {})

    def get_timeout_seconds(self, hook_type: str = "default") -> int:
        """Get timeout seconds for a specific hook type."""
        timeouts = {
            "git": self.get("hooks.git.timeout_seconds", 2),
            "network": self.get("hooks.network.timeout_seconds", 0.1),
            "version_check": self.get("hooks.version_check.timeout_seconds", 1),
        }
        return timeouts.get(hook_type, self.get("hooks.timeout_seconds", 5))

    def get_timeout_ms(self) -> int:
        """Get timeout milliseconds for hooks."""
        return self.get("hooks.timeout_ms", 5000)

    def get_minimum_timeout_seconds(self) -> int:
        """Get minimum allowed timeout seconds."""
        return self.get("hooks.minimum_timeout_seconds", 1)

    def get_graceful_degradation(self) -> bool:
        """Get graceful degradation setting."""
        return self.get("hooks.graceful_degradation", True)

    def get_message(self, category: str, subcategory: str, key: str) -> str:
        """Get localized message from configuration."""
        message = self.get(f"hooks.messages.{category}.{subcategory}.{key}")

        if message is None:
            # Fallback to defaults
            default_messages = DEFAULT_CONFIG["hooks"]["messages"]
            message = default_messages.get(category, {}).get(subcategory, {}).get(key)

        return message or f"Message not found: {category}.{subcategory}.{key}"

    def get_cache_config(self) -> Dict[str, Any]:
        """Get cache configuration."""
        return self.get("hooks.cache", {})

    def get_project_search_config(self) -> Dict[str, Any]:
        """Get project search configuration."""
        return self.get("hooks.project_search", {})

    def get_network_config(self) -> Dict[str, Any]:
        """Get network configuration."""
        return self.get("hooks.network", {})

    def get_git_config(self) -> Dict[str, Any]:
        """Get git configuration."""
        return self.get("hooks.git", {})

    def get_tag_rules_config(self) -> Dict[str, Any]:
        """Get tag rules configuration."""
        return self.get("hooks.tag_rules", {})

    def get_tag_validation_config(self) -> Dict[str, Any]:
        """Get tag validation configuration."""
        return self.get("hooks.tag_validation", {})

    def get_exit_code(self, exit_type: str) -> int:
        """Get exit code for specific exit type."""
        return self.get("hooks.exit_codes", {}).get(exit_type, 0)

    def update_config(self, updates: Dict[str, Any]) -> bool:
        """Update configuration with new values."""
        with self._config_lock:
            try:
                current_config = self.load_config()
                updated_config = self._merge_configs(current_config, updates)

                # Ensure parent directory exists
                self.config_path.parent.mkdir(parents=True, exist_ok=True)

                with open(self.config_path, 'w', encoding='utf-8') as f:
                    json.dump(updated_config, f, indent=2, ensure_ascii=False)

                # Invalidate cache
                self._config = updated_config
                self._cache_timestamp = time.time()
                self._file_mtime = self.config_path.stat().st_mtime

                return True
            except (IOError, OSError, json.JSONEncodeError, PermissionError):
                return False

    def validate_config(self) -> bool:
        """Validate current configuration."""
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

    def get_language_config(self) -> Dict[str, Any]:
        """Get language configuration."""
        return self.get("language", {"conversation_language": "en"})


# Global configuration manager instance
_config_manager = None


def get_config_manager(config_path: Optional[Path] = None) -> ConfigManager:
    """Get global configuration manager instance."""
    global _config_manager
    if _config_manager is None or config_path is not None:
        _config_manager = ConfigManager.get_instance(config_path)
    return _config_manager


def get_config(key_path: str, default: Any = None) -> Any:
    """Get configuration value using dot notation."""
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