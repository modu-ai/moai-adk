#!/usr/bin/env python3
"""User interaction handlers

Handling the UserPromptSubmit event with enhanced agent delegation and skills JIT context
"""

from datetime import datetime
from pathlib import Path

from core import HookPayload, HookResult
from core.context import get_jit_context
from core.agent_context import get_enhanced_jit_context


def handle_user_prompt_submit(payload: HookPayload) -> HookResult:
    """Enhanced UserPromptSubmit event handler

    Analyze user prompts and automatically add relevant documents into context.
    Features expert agent delegation and skills JIT context loading.

    Args:
        payload: Claude Code event payload
                 (includes userPrompt, cwd keys)

    Returns:
        HookResult(
            system_message=Enhanced context and agent recommendation message,
            context_files=Recommended document path list including skill references
        )

    Enhanced Features:
        - Expert agent delegation based on prompt intent analysis
        - Skills JIT context loading with agent-specific skill recommendations
        - Traditional document context loading (backward compatible)
        - Agent collaboration suggestions for complex tasks

    TDD History:
        - RED: JIT document loading scenario testing
        - GREEN: Recommend documents by calling get_jit_context()
        - REFACTOR: Message conditional display (only when there is a file)
        - UPDATE: Migrated to Claude Code standard Hook schema with snake_case fields
        - FEATURE: Command execution logging for tracking double-run debugging
        - ENHANCE: Expert agent delegation and skills JIT context loading
    """
    user_prompt = payload.get("userPrompt", "")
    cwd = payload.get("cwd", ".")

    # Enhanced context with agent delegation and skills JIT loading
    context_files, system_message = get_enhanced_jit_context(user_prompt, cwd)

    # Command execution logging (DEBUG feature for tracking invocations)
    if user_prompt.startswith("/alfred:"):
        try:
            log_dir = Path(cwd) / ".moai" / "logs"
            log_dir.mkdir(parents=True, exist_ok=True)

            log_file = log_dir / "command-invocations.log"
            timestamp = datetime.now().isoformat()

            # Enhanced logging with agent information
            log_entry = f"{timestamp} | {user_prompt}"
            if system_message and "Expert Agent" in system_message:
                log_entry += " | AGENT_DELEGATION"
            if context_files and any("skills/" in str(f) for f in context_files):
                log_entry += " | SKILLS_JIT"

            with open(log_file, "a", encoding="utf-8") as f:
                f.write(f"{log_entry}\n")
        except Exception:
            # Silently fail if logging fails (don't interrupt main flow)
            pass

    # Fallback system message for backward compatibility
    if not system_message and context_files:
        system_message = f"📎 Loaded {len(context_files)} context file(s)"

    return HookResult(system_message=system_message, context_files=context_files)


__all__ = ["handle_user_prompt_submit"]
