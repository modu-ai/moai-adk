#!/usr/bin/env python3
"""Core module for Alfred Hooks

Common type definitions and utility functions
"""

from dataclasses import asdict, dataclass, field
from typing import Any, NotRequired, TypedDict


class HookPayload(TypedDict):
    """Claude Code Hook event payload type definition

    Data structure that Claude Code passes to the Hook script.
    Use NotRequired because fields may vary depending on the event.
    """

    cwd: str
    userPrompt: NotRequired[str] # Includes only UserPromptSubmit events
    tool: NotRequired[str]  # PreToolUse/PostToolUse events
    arguments: NotRequired[dict[str, Any]]  # Tool arguments


@dataclass
class HookResult:
    """Hook execution result"""

    message: str | None = None
    systemMessage: str | None = None  # Message displayed directly to the user  # noqa: N815
    blocked: bool = False
    contextFiles: list[str] = field(default_factory=list)  # noqa: N815
    suggestions: list[str] = field(default_factory=list)
    exitCode: int = 0  # noqa: N815

    def to_dict(self) -> dict[str, Any]:
        """Dictionary conversion for general Hook"""
        return asdict(self)

    def to_user_prompt_submit_dict(self) -> dict[str, Any]:
        """UserPromptSubmit Hook-specific output format

        Claude Code requires a special schema for UserPromptSubmit:
        {
            "hookEventName": "UserPromptSubmit",
            "additionalContext": "string (required)"
        }

        Returns:
            Claude Code UserPromptSubmit Hook Dictionary matching schema

        Examples:
            >>> result = HookResult(contextFiles=["tests/"])
            >>> result.to_user_prompt_submit_dict()
            {'hookEventName': 'UserPromptSubmit', 'additionalContext': '📎 Context: tests/'}
        """
        # Convert contextFiles to additionalContext string
        if self.contextFiles:
            context_str = "\n".join([f"📎 Context: {f}" for f in self.contextFiles])
        else:
            context_str = ""

        # Add message if there is one
        if self.message:
            if context_str:
                context_str = f"{self.message}\n\n{context_str}"
            else:
                context_str = self.message

        # If the string is empty, use default
        if not context_str:
            context_str = ""

        return {
            "hookEventName": "UserPromptSubmit",
            "additionalContext": context_str
        }


__all__ = ["HookPayload", "HookResult"]

# Note: core module exports:
# - HookPayload, HookResult (type definitions)
# - project.py: detect_language, get_git_info, count_specs, get_project_language
# - context.py: get_jit_context
# - checkpoint.py: detect_risky_operation, create_checkpoint, log_checkpoint, list_checkpoints
