"""Common utility functions for MoAI hooks

Consolidated fallback implementations used across multiple hooks.
"""

import statistics
from typing import Dict, List


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
        "std": statistics.stdev(values) if len(values) > 1 else 0
    }
