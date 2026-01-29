# type: ignore
"""
Statusline configuration loader for Claude Code

Loads and manages statusline configuration from .moai/config/statusline-config.yaml

Performance: Thread-safe singleton with double-checked locking pattern
"""

import logging
import threading
from dataclasses import dataclass
from pathlib import Path
from typing import Any, Dict, Optional

logger = logging.getLogger(__name__)


@dataclass
class CacheConfig:
    """Cache TTL configuration"""

    git_ttl_seconds: int = 5
    metrics_ttl_seconds: int = 10
    alfred_ttl_seconds: int = 1
    todo_ttl_seconds: int = 3
    memory_ttl_seconds: int = 5
    output_style_ttl_seconds: int = 60
    version_ttl_seconds: int = 60
    update_ttl_seconds: int = 300


@dataclass
class ColorConfig:
    """Color configuration"""

    enabled: bool = True
    theme: str = "auto"  # auto | light | dark | high-contrast
    palette: Dict[str, str] = None

    def __post_init__(self):
        if self.palette is None:
            self.palette = {
                "model": "38;5;33",  # Blue
                "output_style": "38;5;219",  # Pink/Magenta
                "feature_branch": "38;5;226",  # Yellow
                "develop_branch": "38;5;51",  # Cyan
                "main_branch": "38;5;46",  # Green
                "staged": "38;5;46",  # Green
                "modified": "38;5;208",  # Orange
                "untracked": "38;5;196",  # Red
                "update_available": "38;5;208",  # Orange
                "memory_usage": "38;5;172",  # Brown/Orange
                "duration": "38;5;240",  # Gray
                "todo_count": "38;5;123",  # Purple
                "separator": "38;5;250",  # Light Gray
            }


@dataclass
class DisplayConfig:
    """Information display settings - Custom ordered status bar"""

    model: bool = True  # 🤖 Model name (glm-4.6, claude-3.5-sonnet, etc.)
    version: bool = True  # 🗿 MoAI-ADK version (0.23.0, etc.)
    output_style: bool = True  # ✍️ Output style (Explanatory, Concise, etc.)
    memory_usage: bool = True  # 💾 Session memory usage
    todo_count: bool = True  # 📋 Active TODO items count
    branch: bool = True  # 🔀 Git branch
    git_status: bool = True  # ✅2 M1 ?10 Git changes status
    duration: bool = True  # ⏱️ Session elapsed time
    directory: bool = True  # 📁 Project name/directory
    active_task: bool = True  # 🎯 Alfred active task
    update_indicator: bool = True  # 🔄 Update notification


@dataclass
class FormatConfig:
    """Format configuration"""

    max_branch_length: int = 20
    truncate_with: str = "..."
    separator: str = " | "

    # Icon configuration for better visual recognition
    icons: Dict[str, str] = None

    def __post_init__(self):
        if self.icons is None:
            self.icons = {
                "git": "🔀",  # Git branch icon (more intuitive than 📊)
                "staged": "✅",  # Staged files
                "modified": "📝",  # Modified files
                "untracked": "❓",  # Untracked files
                "added": "➕",  # Added files
                "deleted": "➖",  # Deleted files
                "renamed": "🔄",  # Renamed files
                "conflict": "⚠️",  # Conflict files
                "model": "🤖",  # AI model
                "output_style": "✍️",  # Writing style
                "duration": "⏱️",  # Time duration
                "memory": "💾",  # Memory usage
                "todo": "📋",  # TODO items
                "update": "🔄",  # Update available
                "project": "📁",  # Project directory
            }


@dataclass
class ErrorHandlingConfig:
    """Error handling configuration"""

    graceful_degradation: bool = True
    log_level: str = "warning"  # warning | error
    fallback_text: str = ""


class StatuslineConfig:
    """
    Thread-safe singleton configuration loader for statusline

    Loads configuration from .moai/config/statusline-config.yaml
    Falls back to default values if file not found or parsing fails

    Performance: Double-checked locking pattern for thread safety
    """

    _instance: Optional["StatuslineConfig"] = None
    _config: Dict[str, Any] = {}
    _lock = threading.Lock()  # Thread-safe singleton lock

    def __new__(cls):
        # Double-checked locking pattern for thread-safe singleton
        if cls._instance is None:
            with cls._lock:
                # Double-check after acquiring lock
                if cls._instance is None:
                    cls._instance = super().__new__(cls)
                    cls._instance._load_config()
        return cls._instance

    def _load_config(self) -> None:
        """Load configuration from YAML file"""
        config_path = self._find_config_file()

        if config_path and config_path.exists():
            try:
                self._config = self._parse_yaml(config_path)
                logger.debug(f"Loaded statusline config from {config_path}")
            except Exception as e:
                logger.warning(f"Failed to load config from {config_path}: {e}")
                self._config = self._get_default_config()
        else:
            logger.debug("Statusline config file not found, using defaults")
            self._config = self._get_default_config()

    @staticmethod
    def _find_config_file() -> Optional[Path]:
        """
        Find statusline config file starting from current directory up to project root

        Returns:
            Path to config file or None if not found
        """
        # Try common locations
        locations = [
            Path.cwd() / ".moai" / "config" / "statusline-config.yaml",
            Path.cwd() / ".moai" / "config" / "statusline-config.yml",
            Path.home() / ".moai" / "config" / "statusline-config.yaml",
        ]

        for path in locations:
            if path.exists():
                return path

        return None

    @staticmethod
    def _parse_yaml(path: Path) -> Dict[str, Any]:
        """
        Parse YAML file

        Args:
            path: Path to YAML file

        Returns:
            Parsed configuration dictionary
        """
        try:
            import yaml

            with open(path, "r", encoding="utf-8", errors="replace") as f:
                data = yaml.safe_load(f)
            return data or {}
        except ImportError:
            logger.debug("PyYAML not available, attempting JSON fallback")
            return StatuslineConfig._parse_json_fallback(path)

    @staticmethod
    def _parse_json_fallback(path: Path) -> Dict[str, Any]:
        """
        Parse YAML as JSON fallback (limited support)

        Args:
            path: Path to file

        Returns:
            Parsed configuration dictionary
        """
        import json

        try:
            with open(path, "r", encoding="utf-8", errors="replace") as f:
                return json.load(f)
        except Exception as e:
            logger.debug(f"JSON fallback failed: {e}")
            return {}

    @staticmethod
    def _get_default_config() -> Dict[str, Any]:
        """
        Get default configuration

        Returns:
            Default configuration dictionary
        """
        return {
            "statusline": {
                "enabled": True,
                "mode": "extended",
                "refresh_interval_ms": 300,
                "colors": {
                    "enabled": True,
                    "theme": "auto",
                    "palette": {
                        "model": "38;5;33",
                        "feature_branch": "38;5;226",
                        "develop_branch": "38;5;51",
                        "main_branch": "38;5;46",
                        "staged": "38;5;46",
                        "modified": "38;5;208",
                        "untracked": "38;5;196",
                        "update_available": "38;5;208",
                    },
                },
                "cache": {
                    "git_ttl_seconds": 5,
                    "metrics_ttl_seconds": 10,
                    "alfred_ttl_seconds": 1,
                    "todo_ttl_seconds": 3,
                    "memory_ttl_seconds": 5,
                    "output_style_ttl_seconds": 60,
                    "version_ttl_seconds": 60,
                    "update_ttl_seconds": 300,
                },
                "display": {
                    "model": True,  # 🤖 Model name (glm-4.6, claude-3.5-sonnet, etc.)
                    "version": True,  # 🗿 MoAI-ADK version (0.23.0, etc.)
                    "output_style": True,  # ✍️ Output style (Explanatory, Concise, etc.)
                    "memory_usage": True,  # 💾 Session memory usage
                    "todo_count": True,  # 📋 Active TODO items count
                    "branch": True,  # 🔀 Git branch
                    "git_status": True,  # ✅2 M1 ?10 Git changes status
                    "duration": True,  # ⏱️ Session elapsed time
                    "directory": True,  # 📁 Project name/directory
                    "active_task": True,  # 🎯 Alfred active task
                    "update_indicator": True,  # 🔄 Update notification
                },
                "error_handling": {
                    "graceful_degradation": True,
                    "log_level": "warning",
                    "fallback_text": "",
                },
                "format": {
                    "max_branch_length": 20,
                    "truncate_with": "...",
                    "separator": " | ",
                    "icons": {
                        "git": "🔀",  # Git branch icon (more intuitive than 📊)
                        "staged": "✅",  # Staged files
                        "modified": "📝",  # Modified files
                        "untracked": "❓",  # Untracked files
                        "added": "➕",  # Added files
                        "deleted": "➖",  # Deleted files
                        "renamed": "🔄",  # Renamed files
                        "conflict": "⚠️",  # Conflict files
                        "model": "🤖",  # AI model
                        "output_style": "✍️",  # Writing style
                        "duration": "⏱️",  # Time duration
                        "memory": "💾",  # Memory usage
                        "todo": "📋",  # TODO items
                        "update": "🔄",  # Update available
                        "project": "📁",  # Project directory
                    },
                },
            }
        }

    def get(self, key: str, default: Any = None) -> Any:
        """
        Get configuration value by dot-notation key

        Args:
            key: Configuration key (e.g., "statusline.mode", "statusline.cache.git_ttl_seconds")
            default: Default value if key not found

        Returns:
            Configuration value or default
        """
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

    def get_cache_config(self) -> CacheConfig:
        """Get cache configuration"""
        cache_data = self.get("statusline.cache", {})
        return CacheConfig(
            git_ttl_seconds=cache_data.get("git_ttl_seconds", 5),
            metrics_ttl_seconds=cache_data.get("metrics_ttl_seconds", 10),
            alfred_ttl_seconds=cache_data.get("alfred_ttl_seconds", 1),
            version_ttl_seconds=cache_data.get("version_ttl_seconds", 60),
            update_ttl_seconds=cache_data.get("update_ttl_seconds", 300),
        )

    def get_color_config(self) -> ColorConfig:
        """Get color configuration"""
        color_data = self.get("statusline.colors", {})
        return ColorConfig(
            enabled=color_data.get("enabled", True),
            theme=color_data.get("theme", "auto"),
            palette=color_data.get("palette", {}),
        )

    def get_display_config(self) -> DisplayConfig:
        """Get display configuration"""
        display_data = self.get("statusline.display", {})
        return DisplayConfig(
            model=display_data.get("model", True),
            version=display_data.get("version", True),
            output_style=display_data.get("output_style", True),
            memory_usage=display_data.get("memory_usage", True),
            todo_count=display_data.get("todo_count", True),
            branch=display_data.get("branch", True),
            git_status=display_data.get("git_status", True),
            duration=display_data.get("duration", True),
            directory=display_data.get("directory", True),
            active_task=display_data.get("active_task", True),
            update_indicator=display_data.get("update_indicator", True),
        )

    def get_format_config(self) -> FormatConfig:
        """Get format configuration"""
        format_data = self.get("statusline.format", {})
        return FormatConfig(
            max_branch_length=format_data.get("max_branch_length", 20),
            truncate_with=format_data.get("truncate_with", "..."),
            separator=format_data.get("separator", " | "),
        )

    def get_error_handling_config(self) -> ErrorHandlingConfig:
        """Get error handling configuration"""
        error_data = self.get("statusline.error_handling", {})
        return ErrorHandlingConfig(
            graceful_degradation=error_data.get("graceful_degradation", True),
            log_level=error_data.get("log_level", "warning"),
            fallback_text=error_data.get("fallback_text", ""),
        )
