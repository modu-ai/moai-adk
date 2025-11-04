#!/usr/bin/env python3
"""User interaction handlers

Handling the UserPromptSubmit event
"""

import json
from datetime import datetime
from pathlib import Path

from core import HookPayload, HookResult
from core.context import get_jit_context
from core.project import find_project_root


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

    # OPTIONAL: Language info - skip if timeout/failure
    conversation_language = "en"
    language_hint = ""
    try:
        project_root = find_project_root(cwd)
        config_path = project_root / ".moai" / "config.json"
        if config_path.exists():
            config = json.loads(config_path.read_text(encoding="utf-8"))
            lang_config = config.get("language", {})
            conversation_language = lang_config.get("conversation_language", "en")
            # Add language hint for user-facing responses
            if conversation_language != "en":
                language_hint = f" [Language: {conversation_language}]"
    except Exception:
        # Graceful degradation - continue without language hint
        pass

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

    system_message = f"📎 Loaded {len(context_files)} context file(s){language_hint}" if context_files else language_hint if language_hint else None

    return HookResult(system_message=system_message, context_files=context_files)


__all__ = ["handle_user_prompt_submit"]
