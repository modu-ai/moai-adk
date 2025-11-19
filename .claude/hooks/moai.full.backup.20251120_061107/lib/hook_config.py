#!/usr/bin/env python3
# Hook Configuration Utilities

"""
Hook-related configuration and utility functions
- Load Hook timeout configuration
- Check graceful degradation settings
"""

import json
import logging
from pathlib import Path

logger = logging.getLogger(__name__)


def load_hook_timeout() -> int:
    """
    Load Hook timeout configuration from .moai/config/config.json

    Returns:
        int: timeout value (milliseconds), returns default 5000 if configuration not found
    """
    try:
        config_path = Path(".moai/config/config.json")
        if not config_path.exists():
            return 5000  # Default value

        with open(config_path, 'r', encoding='utf-8') as f:
            config = json.load(f)

        # Get timeout_ms value from hooks section
        hooks_config = config.get("hooks", {})
        timeout_ms = hooks_config.get("timeout_ms", 5000)

        return int(timeout_ms)
    except (json.JSONDecodeError, FileNotFoundError, KeyError, ValueError):
        logger.warning("Failed to load hook timeout from config, using default 5000ms")
        return 5000


def get_graceful_degradation() -> bool:
    """
    Load graceful_degradation configuration from .moai/config/config.json

    Returns:
        bool: graceful_degradation configuration value, returns default True if configuration not found
    """
    try:
        config_path = Path(".moai/config/config.json")
        if not config_path.exists():
            return True  # Default value

        with open(config_path, 'r', encoding='utf-8') as f:
            config = json.load(f)

        # Get graceful_degradation value from hooks section
        hooks_config = config.get("hooks", {})
        return hooks_config.get("graceful_degradation", True)
    except (json.JSONDecodeError, FileNotFoundError, KeyError):
        logger.warning("Failed to load graceful_degradation from config, using default True")
        return True


def get_hook_execution_config() -> dict:
    """
    Load all Hook execution related configuration

    Returns:
        dict: Hook configuration dictionary
    """
    try:
        config_path = Path(".moai/config/config.json")
        if not config_path.exists():
            return {
                "timeout_ms": 5000,
                "graceful_degradation": True
            }

        with open(config_path, 'r', encoding='utf-8') as f:
            config = json.load(f)

        hooks_config = config.get("hooks", {})

        return {
            "timeout_ms": hooks_config.get("timeout_ms", 5000),
            "graceful_degradation": hooks_config.get("graceful_degradation", True)
        }
    except (json.JSONDecodeError, FileNotFoundError):
        logger.warning("Failed to load hook config, using defaults")
        return {
            "timeout_ms": 5000,
            "graceful_degradation": True
        }
