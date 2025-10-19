# @CODE:PY314-001 | SPEC: SPEC-PY314-001.md | TEST: tests/unit/test_config_manager.py
"""Configuration Manager

Manage .moai/config.json:
- Read and write configuration files
- Support deep merges
- Preserve UTF-8 content
- Create directories automatically
"""

import json
from pathlib import Path
from typing import Any


class ConfigManager:
    """Read and write .moai/config.json."""

    DEFAULT_CONFIG = {
        "mode": "personal",
        "locale": "ko",
        "moai": {
            "version": "0.3.0"
        }
    }

    def __init__(self, config_path: Path) -> None:
        """Initialize the ConfigManager.

        Args:
            config_path: Path to config.json.
        """
        self.config_path = config_path

    def load(self) -> dict[str, Any]:
        """Load the configuration file.

        Returns default values when the file is missing.

        Returns:
            Configuration dictionary.
        """
        if not self.config_path.exists():
            return self.DEFAULT_CONFIG.copy()

        with open(self.config_path, encoding="utf-8") as f:
            data: dict[str, Any] = json.load(f)
            return data

    def save(self, config: dict[str, Any]) -> None:
        """Persist the configuration file.

        Creates directories when missing and preserves UTF-8 content.

        Args:
            config: Configuration dictionary to save.
        """
        # Ensure the directory exists
        self.config_path.parent.mkdir(parents=True, exist_ok=True)

        # Write while preserving UTF-8 characters
        with open(self.config_path, "w", encoding="utf-8") as f:
            json.dump(config, f, ensure_ascii=False, indent=2)

    def update(self, updates: dict[str, Any]) -> None:
        """Update the configuration using a deep merge.

        Args:
            updates: Dictionary of updates to apply.
        """
        current = self.load()
        merged = self._deep_merge(current, updates)
        self.save(merged)

    def _deep_merge(self, base: dict[str, Any], updates: dict[str, Any]) -> dict[str, Any]:
        """Recursively deep-merge dictionaries.

        Args:
            base: Base dictionary.
            updates: Dictionary with updates.

        Returns:
            Merged dictionary.
        """
        result = base.copy()

        for key, value in updates.items():
            if key in result and isinstance(result[key], dict) and isinstance(value, dict):
                # When both sides are dicts, merge recursively
                result[key] = self._deep_merge(result[key], value)
            else:
                # Otherwise, overwrite the value
                result[key] = value

        return result

    @staticmethod
    def set_optimized(config_path: Path, value: bool) -> None:
        """Set the optimized field in config.json.

        Args:
            config_path: Path to config.json.
            value: Value to set (True or False).
        """
        if not config_path.exists():
            return

        try:
            with open(config_path, encoding="utf-8") as f:
                config = json.load(f)

            config.setdefault("project", {})["optimized"] = value

            with open(config_path, "w", encoding="utf-8") as f:
                json.dump(config, f, ensure_ascii=False, indent=2)
                f.write("\n")  # Add trailing newline
        except (json.JSONDecodeError, KeyError, OSError):
            # Ignore errors if config.json is invalid or inaccessible
            pass
