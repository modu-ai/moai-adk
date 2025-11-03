#!/usr/bin/env python3
"""Re-export handlers from shared module

This module provides backward compatibility by re-exporting handlers
from the shared.handlers module to allow alfred_hooks.py to import
directly from handlers instead of shared.handlers.
"""

from shared.handlers import (
    handle_notification,
    handle_post_tool_use,
    handle_pre_tool_use,
    handle_session_end,
    handle_session_start,
    handle_stop,
    handle_subagent_stop,
    handle_user_prompt_submit,
)

__all__ = [
    "handle_session_start",
    "handle_session_end",
    "handle_user_prompt_submit",
    "handle_pre_tool_use",
    "handle_post_tool_use",
    "handle_notification",
    "handle_stop",
    "handle_subagent_stop",
]
