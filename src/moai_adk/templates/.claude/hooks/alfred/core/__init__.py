#!/usr/bin/env python3
"""Core module for Alfred Hooks

Common type definitions and utility functions
"""

from dataclasses import dataclass, field
from typing import Any, Literal, NotRequired, TypedDict


class HookPayload(TypedDict):
    """Claude Code Hook event payload type definition

    Data structure that Claude Code passes to the Hook script.
    Use NotRequired because fields may vary depending on the event.
    """

    cwd: str
    userPrompt: NotRequired[str]  # Includes only UserPromptSubmit events
    tool: NotRequired[str]  # PreToolUse/PostToolUse events
    arguments: NotRequired[dict[str, Any]]  # Tool arguments


@dataclass
class HookResult:
    """Hook execution result following Claude Code standard schema.

    Attributes conform to Claude Code Hook output specification:
    https://docs.claude.com/en/docs/claude-code/hooks

    Standard Fields (Claude Code schema):
        continue_execution: Allow execution to continue (default True)
        suppress_output: Suppress hook output display (default False)
        decision: "approve" or "block" operation (optional)
        reason: Explanation for decision (optional)
        permission_decision: "allow", "deny", or "ask" (optional)

    MoAI-ADK Fields (wrapped in hookSpecificOutput):
        system_message: Message displayed to user
        context_files: List of context files to load
        suggestions: Suggestions for user
        exit_code: Exit code (for diagnostics)
    """

    # Claude Code standard fields
    continue_execution: bool = True
    suppress_output: bool = False
    decision: Literal["approve", "block"] | None = None
    reason: str | None = None
    permission_decision: Literal["allow", "deny", "ask"] | None = None

    # MoAI-ADK custom fields (wrapped in hookSpecificOutput)
    system_message: str | None = None
    context_files: list[str] = field(default_factory=list)
    suggestions: list[str] = field(default_factory=list)
    exit_code: int = 0

    def to_dict(self) -> dict[str, Any]:
        """Convert to Claude Code standard Hook output schema.

        Returns:
            Dictionary conforming to Claude Code Hook specification with:
            - Top-level fields: continue, suppressOutput, decision, reason, permissionDecision
            - Nested field: hookSpecificOutput containing MoAI-ADK-specific data

        Examples:
            >>> result = HookResult(continue_execution=True)
            >>> result.to_dict()
            {'continue': True}

            >>> result = HookResult(decision="block", reason="Dangerous")
            >>> result.to_dict()
            {'decision': 'block', 'reason': 'Dangerous'}

            >>> result = HookResult(system_message="Test", context_files=["a.txt"])
            >>> result.to_dict()
            {'continue': True, 'hookSpecificOutput': {'systemMessage': 'Test', 'contextFiles': ['a.txt']}}
        """
        output: dict[str, Any] = {}

        # Add decision or continue flag
        if self.decision:
            output["decision"] = self.decision
            if self.reason:
                output["reason"] = self.reason
        else:
            output["continue"] = self.continue_execution

        # Add suppressOutput if True
        if self.suppress_output:
            output["suppressOutput"] = True

        # Add permissionDecision if set
        if self.permission_decision:
            output["permissionDecision"] = self.permission_decision

        # Wrap MoAI-ADK custom fields in hookSpecificOutput
        hook_output: dict[str, Any] = {}

        if self.system_message:
            hook_output["systemMessage"] = self.system_message

        if self.context_files:
            hook_output["contextFiles"] = self.context_files

        if self.suggestions:
            hook_output["suggestions"] = self.suggestions

        if self.exit_code != 0:
            hook_output["exitCode"] = self.exit_code

        # Only add hookSpecificOutput if there's custom data
        if hook_output:
            output["hookSpecificOutput"] = hook_output

        return output

    def to_user_prompt_submit_dict(self) -> dict[str, Any]:
        """UserPromptSubmit Hook-specific output format.

        Claude Code requires a special schema for UserPromptSubmit events.
        The result is wrapped in the standard Hook schema with hookSpecificOutput.

        Returns:
            Claude Code UserPromptSubmit Hook Dictionary matching schema:
            {
                "continue": true,
                "hookSpecificOutput": {
                    "hookEventName": "UserPromptSubmit",
                    "additionalContext": "string"
                }
            }

        Examples:
            >>> result = HookResult(context_files=["tests/"])
            >>> result.to_user_prompt_submit_dict()
            {'continue': True, 'hookSpecificOutput': {'hookEventName': 'UserPromptSubmit', 'additionalContext': '📎 Context: tests/'}}
        """
        # Convert context_files to additionalContext string
        if self.context_files:
            context_str = "\n".join([f"📎 Context: {f}" for f in self.context_files])
        else:
            context_str = ""

        # Add system_message if there is one
        if self.system_message:
            if context_str:
                context_str = f"{self.system_message}\n\n{context_str}"
            else:
                context_str = self.system_message

        return {
            "continue": self.continue_execution,
            "hookSpecificOutput": {
                "hookEventName": "UserPromptSubmit",
                "additionalContext": context_str
            }
        }


__all__ = ["HookPayload", "HookResult"]

# Note: core module exports:
# - HookPayload, HookResult (type definitions)
# - project.py: detect_language, get_git_info, count_specs, get_project_language
# - context.py: get_jit_context
# - checkpoint.py: detect_risky_operation, create_checkpoint, log_checkpoint, list_checkpoints
