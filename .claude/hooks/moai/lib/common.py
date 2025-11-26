"""Common utility functions for MoAI hooks

Consolidated fallback implementations used across multiple hooks.
"""

import re
import statistics
from typing import Any, Dict, List, Optional, Tuple


def format_duration(seconds: float) -> str:
    """Format duration in seconds to readable string.

    Converts seconds to human-readable format (s, m, h).

    Args:
        seconds: Duration in seconds

    Returns:
        Formatted duration string (e.g., "2.5m", "1.3h")
    """
    if seconds < 60:
        return f"{seconds:.1f}s"
    minutes = seconds / 60
    if minutes < 60:
        return f"{minutes:.1f}m"
    hours = minutes / 60
    return f"{hours:.1f}h"


def get_summary_stats(values: List[float]) -> Dict[str, float]:
    """Get summary statistics for a list of values.

    Calculates mean, min, max, and standard deviation.

    Args:
        values: List of numeric values

    Returns:
        Dictionary with keys: mean, min, max, std
    """
    if not values:
        return {"mean": 0, "min": 0, "max": 0, "std": 0}

    return {
        "mean": statistics.mean(values),
        "min": min(values),
        "max": max(values),
        "std": statistics.stdev(values) if len(values) > 1 else 0,
    }


def is_root_whitelisted(filename: str, config: Dict[str, Any]) -> bool:
    """Check if file is allowed in project root.

    Consolidated from pre_tool__document_management.py and session_end__auto_cleanup.py

    Args:
        filename: Name of the file
        config: Configuration dictionary

    Returns:
        True if file is whitelisted for root directory
    """
    whitelist = config.get("document_management", {}).get("root_whitelist", [])

    for pattern in whitelist:
        # Convert glob pattern to regex
        regex = pattern.replace("*", ".*").replace("?", ".")
        if re.match(f"^{regex}$", filename):
            return True

    return False


def get_file_pattern_category(filename: str, config: Dict[str, Any]) -> Optional[Tuple[str, str]]:
    """Match filename against patterns to determine category.

    Consolidated from pre_tool__document_management.py and session_end__auto_cleanup.py

    Args:
        filename: Name of the file to categorize
        config: Configuration dictionary

    Returns:
        Tuple of (directory_type, category) or None if no match
    """
    patterns = config.get("document_management", {}).get("file_patterns", {})

    for dir_type, categories in patterns.items():
        for category, pattern_list in categories.items():
            for pattern in pattern_list:
                # Convert glob pattern to regex
                regex = pattern.replace("*", ".*").replace("?", ".")
                if re.match(f"^{regex}$", filename):
                    return (dir_type, category)

    return None


def suggest_moai_location(filename: str, config: Dict[str, Any]) -> str:
    """Suggest appropriate .moai/ location based on file pattern.

    Consolidated from pre_tool__document_management.py and session_end__auto_cleanup.py

    Args:
        filename: Name of the file
        config: Configuration dictionary

    Returns:
        Suggested .moai/ path
    """
    # Try pattern matching first
    match = get_file_pattern_category(filename, config)

    if match:
        dir_type, category = match
        base_dir = config.get("document_management", {}).get("directories", {}).get(dir_type, {}).get("base", "")
        if base_dir:
            return f"{base_dir}{category}/"

    # Default fallback suggestions
    if filename.endswith(".md"):
        return ".moai/temp/work/"
    elif filename.endswith((".sh", ".py", ".js")):
        return ".moai/scripts/dev/"
    elif filename.endswith((".tmp", ".temp", ".bak")):
        return ".moai/temp/work/"

    # Ultimate fallback
    return ".moai/temp/work/"
