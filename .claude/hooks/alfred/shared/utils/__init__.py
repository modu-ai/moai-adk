#!/usr/bin/env python3
"""Utility modules for Alfred Hooks

This package provides shared utilities for hook execution,
state tracking, and deduplication.
"""

from .state_tracking import (
    HookStateManager,
    get_state_manager,
    track_hook_execution,
    deduplicate_command,
    mark_command_complete
)

__all__ = [
    "HookStateManager",
    "get_state_manager",
    "track_hook_execution",
    "deduplicate_command",
    "mark_command_complete"
]