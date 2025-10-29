#!/usr/bin/env python3
"""Notification and control handlers

Notification, Stop, SubagentStop event handling
"""

from core import HookPayload, HookResult


def handle_notification(payload: HookPayload) -> HookResult:
    """Notification event handler (default implementation)"""
    return HookResult()


def handle_stop(payload: HookPayload) -> HookResult:
    """Stop event handler (default implementation)"""
    return HookResult()


def handle_subagent_stop(payload: HookPayload) -> HookResult:
    """SubagentStop event handler (default implementation)"""
    return HookResult()


__all__ = ["handle_notification", "handle_stop", "handle_subagent_stop"]
