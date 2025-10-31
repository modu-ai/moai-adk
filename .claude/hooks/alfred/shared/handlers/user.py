#!/usr/bin/env python3
"""User interaction handlers

Handling the UserPromptSubmit event
"""

from datetime import datetime
from pathlib import Path

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
            system_message=Number of Files loaded (or None),
            context_files=Recommended document path list
        )

    TDD History:
        - RED: JIT document loading scenario testing
        - GREEN: Recommend documents by calling get_jit_context()
        - REFACTOR: Message conditional display (only when there is a file)
        - UPDATE: Migrated to Claude Code standard Hook schema with snake_case fields
        - FEATURE: Command execution logging for tracking double-run debugging
    """
    user_prompt = payload.get("userPrompt", "")
    cwd = payload.get("cwd", ".")
    context_files = get_jit_context(user_prompt, cwd)

    # Command execution logging (DEBUG feature for tracking invocations)
    if user_prompt.startswith("/alfred:"):
        try:
            log_dir = Path(cwd) / ".moai" / "logs"
            log_dir.mkdir(parents=True, exist_ok=True)

            log_file = log_dir / "command-invocations.log"
            timestamp = datetime.now().isoformat()

            with open(log_file, "a", encoding="utf-8") as f:
                f.write(f"{timestamp} | {user_prompt}\n")
        except Exception:
            # Silently fail if logging fails (don't interrupt main flow)
            pass

    system_message = f"📎 Loaded {len(context_files)} context file(s)" if context_files else None

    return HookResult(system_message=system_message, context_files=context_files)


__all__ = ["handle_user_prompt_submit"]
