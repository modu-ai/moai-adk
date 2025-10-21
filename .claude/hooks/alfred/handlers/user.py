#!/usr/bin/env python3
"""User interaction handlers

Handling the UserPromptSubmit event
"""

from core import HookPayload, HookResult
from core.context import get_jit_context


def handle_user_prompt_submit(payload: HookPayload) -> HookResult:
    """UserPromptSubmit event handler

    Analyze user prompts and automatically add relevant documents into context.
    Follow the just-in-time (JIT) retrieval principle to load only the documents you need.

    Args:
        payload: Claude Code event payload
                 (includes userPrompt, cwd keys)

    Returns:
        HookResult(
            message=Number of Files loaded (or None),
            contextFiles=Recommended document path list
        )

    TDD History:
        - RED: JIT document loading scenario testing
        - GREEN: Recommend documents by calling get_jit_context()
        - REFACTOR: Message conditional display (only when there is a file)
    """
    user_prompt = payload.get("userPrompt", "")
    cwd = payload.get("cwd", ".")
    context_files = get_jit_context(user_prompt, cwd)

    message = f"📎 Loaded {len(context_files)} context file(s)" if context_files else None

    return HookResult(message=message, contextFiles=context_files)


__all__ = ["handle_user_prompt_submit"]
