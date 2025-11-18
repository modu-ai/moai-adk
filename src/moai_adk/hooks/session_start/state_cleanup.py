"""State cleanup module for session_start hook

Manages configuration loading, state validation, and cleanup scheduling.

Responsibilities:
- Load hook timeout from config
- Get graceful degradation setting
- Load complete configuration
- Determine cleanup schedule
"""

import json
import logging
from datetime import datetime, timedelta
from pathlib import Path
from typing import Any, Dict, Optional

logger = logging.getLogger(__name__)


class StateError(Exception):
    """Exception raised for state-related errors"""

    pass


def load_hook_timeout() -> int:
    """Load hook timeout from config.json (default: 3000ms)

    Returns:
        Timeout in milliseconds (default: 3000)

    Raises:
        StateError: If config file is corrupted
    """
    try:
        config_file = Path(".moai/config/config.json")
        if config_file.exists():
            with open(config_file, "r", encoding="utf-8") as f:
                config = json.load(f)
                return config.get("hooks", {}).get("timeout_ms", 3000)
    except json.JSONDecodeError as e:
        logger.warning(f"Config file corrupted, using default timeout: {e}")
    except Exception as e:
        logger.warning(f"Failed to load timeout from config: {e}")

    return 3000


def get_graceful_degradation() -> bool:
    """Load graceful_degradation setting from config.json (default: True)

    Graceful degradation allows hook to continue even if execution fails.

    Returns:
        True if graceful degradation enabled (default: True)

    Raises:
        StateError: If config file is corrupted
    """
    try:
        config_file = Path(".moai/config/config.json")
        if config_file.exists():
            with open(config_file, "r", encoding="utf-8") as f:
                config = json.load(f)
                return config.get("hooks", {}).get("graceful_degradation", True)
    except json.JSONDecodeError as e:
        logger.warning(f"Config file corrupted, enabling graceful degradation: {e}")
    except Exception as e:
        logger.warning(f"Failed to load graceful_degradation from config: {e}")

    return True


def load_config() -> Dict[str, Any]:
    """Load configuration from .moai/config/config.json

    Returns:
        Configuration dictionary (empty dict if file not found)

    Raises:
        StateError: If config file is corrupted
    """
    try:
        config_file = Path(".moai/config/config.json")
        if config_file.exists():
            with open(config_file, "r", encoding="utf-8") as f:
                return json.load(f)
    except json.JSONDecodeError as e:
        logger.error(f"Config file corrupted: {e}")
        raise StateError(f"Failed to parse config.json: {e}") from e
    except Exception as e:
        logger.warning(f"Failed to load config: {e}")

    return {}


def should_cleanup_today(
    last_cleanup: Optional[str], cleanup_days: int = 7
) -> bool:
    """Determine if cleanup is needed today

    Args:
        last_cleanup: Last cleanup date in YYYY-MM-DD format (None for first run)
        cleanup_days: Cleanup interval in days (default: 7)

    Returns:
        True if cleanup is needed, False otherwise

    Raises:
        StateError: If date parsing fails
    """
    if not last_cleanup:
        return True

    try:
        last_date = datetime.strptime(last_cleanup, "%Y-%m-%d")
        next_cleanup = last_date + timedelta(days=cleanup_days)
        return datetime.now() >= next_cleanup
    except ValueError as e:
        logger.warning(f"Invalid date format: {last_cleanup}: {e}")
        return True


def validate_cleanup_config(config: Dict[str, Any]) -> bool:
    """Validate cleanup configuration

    Args:
        config: Configuration dictionary

    Returns:
        True if configuration is valid

    Raises:
        StateError: If configuration is invalid
    """
    cleanup_config = config.get("auto_cleanup", {})

    # Check required fields
    if not isinstance(cleanup_config.get("cleanup_days"), (int, float)):
        logger.warning("cleanup_days not configured, using default 7")

    if not isinstance(cleanup_config.get("max_reports"), (int, float)):
        logger.warning("max_reports not configured, using default 10")

    return True
