#!/usr/bin/env python3
"""Notification and control handlers

Notification, Stop, SubagentStop 이벤트 처리
"""

from core import HookPayload, HookResult


def handle_notification(payload: HookPayload) -> HookResult:
    """Notification 이벤트 핸들러 (기본 구현)"""
    return HookResult()


def handle_stop(payload: HookPayload) -> HookResult:
    """Stop 이벤트 핸들러 (기본 구현)"""
    return HookResult()


def handle_subagent_stop(payload: HookPayload) -> HookResult:
    """SubagentStop 이벤트 핸들러 (기본 구현)"""
    return HookResult()


__all__ = ["handle_notification", "handle_stop", "handle_subagent_stop"]
