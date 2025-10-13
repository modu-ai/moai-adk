#!/usr/bin/env python3
# /// script
# requires-python = ">=3.13"
# dependencies = []
# ///
# @CODE:HOOKS-002 | SPEC: SPEC-HOOKS-002.md | TEST: tests/unit/test_moai_hooks.py
"""
MoAI-ADK Claude Code Hooks (Self-contained)

600 LOC self-contained Python script for Claude Code hook system.
Zero external dependencies - uses only Python standard library.

Usage:
    uv run --python 3.13 .claude/hooks/moai_hooks.py {event}

Events:
    SessionStart, SessionEnd, PreToolUse, PostToolUse,
    UserPromptSubmit, Notification, Stop, SubagentStop, PreCompact

Input: JSON payload via stdin
Output: JSON result via stdout
"""

import json
import subprocess
import sys
from dataclasses import dataclass, field
from pathlib import Path
from typing import Any, Literal, TypedDict, cast

# ============================================================================
# Data Structures
# ============================================================================


# Hook Event Types
HookEvent = Literal[
    "SessionStart",
    "SessionEnd",
    "PreToolUse",
    "PostToolUse",
    "UserPromptSubmit",
    "Notification",
    "Stop",
    "SubagentStop",
    "PreCompact",
]


# Hook Payload
class HookPayload(TypedDict, total=False):
    event: str
    cwd: str
    toolName: str | None
    toolArgs: dict[str, Any] | None
    userPrompt: str | None
    notificationMessage: str | None


# Hook Result
@dataclass
class HookResult:
    message: str | None = None
    blocked: bool = False
    suggestions: list[str] = field(default_factory=list)
    contextFiles: list[str] = field(default_factory=list)  # noqa: N815

    def to_json(self) -> str:
        return json.dumps(
            {
                "message": self.message,
                "blocked": self.blocked,
                "suggestions": self.suggestions,
                "contextFiles": self.contextFiles,
            },
            indent=2,
        )


# ============================================================================
# Language Detection
# ============================================================================

# Language detection patterns
LANGUAGE_PATTERNS = {
    "python": ["pyproject.toml", "setup.py", "requirements.txt", "*.py"],
    "typescript": ["tsconfig.json", "*.ts", "*.tsx"],
    "javascript": ["package.json", "*.js", "*.jsx"],
    "java": ["pom.xml", "build.gradle", "*.java"],
    "go": ["go.mod", "go.sum", "*.go"],
    "rust": ["Cargo.toml", "Cargo.lock", "*.rs"],
    "dart": ["pubspec.yaml", "*.dart"],
    "swift": ["Package.swift", "*.swift"],
    "kotlin": ["build.gradle.kts", "*.kt", "*.kts"],
    "csharp": ["*.csproj", "*.sln", "*.cs"],
    "php": ["composer.json", "*.php"],
    "ruby": ["Gemfile", "Gemfile.lock", "*.rb"],
    "elixir": ["mix.exs", "*.ex", "*.exs"],
    "scala": ["build.sbt", "*.scala"],
    "clojure": ["project.clj", "deps.edn", "*.clj"],
    "haskell": ["stack.yaml", "*.cabal", "*.hs"],
    "cpp": ["CMakeLists.txt", "*.cpp", "*.hpp"],
    "c": ["Makefile", "*.c", "*.h"],
    "shell": ["*.sh", "*.bash"],
    "lua": ["*.lua"],
}


def detect_language(cwd: str) -> str:
    """Detect project language from file patterns"""
    project_path = Path(cwd)

    for language, patterns in LANGUAGE_PATTERNS.items():
        for pattern in patterns:
            if "*" in pattern:
                # Glob pattern
                if list(project_path.rglob(pattern)):
                    return language
            else:
                # Direct file check
                if (project_path / pattern).exists():
                    return language

    return "Unknown Language"


def get_git_info(cwd: str) -> dict[str, Any]:
    """Collect Git repository information with 2s timeout"""
    try:
        # Check if Git repository
        result = subprocess.run(
            ["git", "rev-parse", "--git-dir"],
            cwd=cwd,
            capture_output=True,
            text=True,
            timeout=2,
        )

        if result.returncode != 0:
            return {}

        # Get current branch
        branch_result = subprocess.run(
            ["git", "branch", "--show-current"],
            cwd=cwd,
            capture_output=True,
            text=True,
            timeout=2,
        )
        branch = branch_result.stdout.strip()

        # Get latest commit
        commit_result = subprocess.run(
            ["git", "rev-parse", "--short", "HEAD"],
            cwd=cwd,
            capture_output=True,
            text=True,
            timeout=2,
        )
        commit = commit_result.stdout.strip()

        # Get change status
        status_result = subprocess.run(
            ["git", "status", "--short"],
            cwd=cwd,
            capture_output=True,
            text=True,
            timeout=2,
        )
        status = status_result.stdout.strip()

        changes = len(status.split("\n")) if status else 0

        return {"branch": branch, "commit": commit, "changes": changes}

    except (subprocess.TimeoutExpired, FileNotFoundError):
        return {}


def count_specs(cwd: str) -> dict[str, int]:
    """Count SPEC files in .moai/specs/ directory"""
    specs_dir = Path(cwd) / ".moai" / "specs"

    if not specs_dir.exists():
        return {"completed": 0, "total": 0, "percentage": 0}

    total = 0
    completed = 0

    for spec_dir in specs_dir.iterdir():
        if spec_dir.is_dir() and spec_dir.name.startswith("SPEC-"):
            total += 1
            spec_file = spec_dir / "spec.md"
            if spec_file.exists():
                content = spec_file.read_text(encoding="utf-8")
                if "status: completed" in content:
                    completed += 1

    percentage = int(completed / total * 100) if total > 0 else 0

    return {"completed": completed, "total": total, "percentage": percentage}


def get_jit_context(user_prompt: str, cwd: str) -> list[str]:
    """Just-in-Time context retrieval based on command patterns"""
    context_files = []

    # Pattern matching for commands
    if "/alfred:1-spec" in user_prompt:
        context_files.append(".moai/memory/spec-metadata.md")

    if "/alfred:2-build" in user_prompt:
        context_files.append(".moai/memory/development-guide.md")

    if any(word in user_prompt.lower() for word in ["test", "pytest", "jest"]):
        tests_dir = Path(cwd) / "tests"
        if tests_dir.exists():
            context_files.append("tests/")

    # Filter existing files
    existing_files = []
    for file_path in context_files:
        full_path = Path(cwd) / file_path
        if full_path.exists():
            existing_files.append(file_path)

    return existing_files


# ============================================================================
# Hook Handlers
# ============================================================================


def handle_session_start(payload: HookPayload) -> HookResult:
    """Handle SessionStart event"""
    cwd = payload.get("cwd", ".")

    # Collect information
    language = detect_language(cwd)
    git_info = get_git_info(cwd)
    spec_count = count_specs(cwd)

    # Build message
    parts = []
    parts.append("🚀 MoAI-ADK Session Started")
    parts.append(f"Language: {language}")

    if git_info:
        parts.append(f"Git: {git_info['branch']} @ {git_info['commit']}")
        if git_info.get("changes", 0) > 0:
            parts.append(f"Changes: {git_info['changes']} files")

    if spec_count["total"] > 0:
        parts.append(
            f"SPECs: {spec_count['completed']}/{spec_count['total']} "
            f"({spec_count['percentage']}%)"
        )

    return HookResult(message="\n".join(parts))


def handle_user_prompt_submit(payload: HookPayload) -> HookResult:
    """Handle UserPromptSubmit event with JIT context"""
    cwd = payload.get("cwd", ".")
    user_prompt = payload.get("userPrompt") or ""

    context_files = get_jit_context(user_prompt, cwd)

    if context_files:
        return HookResult(
            contextFiles=context_files,
            message=f"📚 Loaded {len(context_files)} context file(s)",
        )

    return HookResult()


def handle_pre_compact(payload: HookPayload) -> HookResult:
    """Handle PreCompact event"""
    return HookResult(
        message="💡 Tip: Use `/clear` or `/new` to start fresh session",
        suggestions=[
            "Summarize current session decisions",
            "Save important context to .moai/memory/",
            "Continue with clean context",
        ],
    )


def handle_session_end(payload: HookPayload) -> HookResult:
    """Handle SessionEnd event (no-op)"""
    return HookResult()


def handle_pre_tool_use(payload: HookPayload) -> HookResult:
    """Handle PreToolUse event (no-op)"""
    return HookResult()


def handle_post_tool_use(payload: HookPayload) -> HookResult:
    """Handle PostToolUse event (no-op)"""
    return HookResult()


def handle_notification(payload: HookPayload) -> HookResult:
    """Handle Notification event (no-op)"""
    return HookResult()


def handle_stop(payload: HookPayload) -> HookResult:
    """Handle Stop event (no-op)"""
    return HookResult()


def handle_subagent_stop(payload: HookPayload) -> HookResult:
    """Handle SubagentStop event (no-op)"""
    return HookResult()


# ============================================================================
# Main Entry Point
# ============================================================================


def main() -> None:
    """Main entry point"""
    try:
        # Parse command line arguments
        if len(sys.argv) < 2:
            print("Usage: moai_hooks.py {event}", file=sys.stderr)
            sys.exit(1)

        event = sys.argv[1]

        # Read JSON payload from stdin
        payload_json = sys.stdin.read()
        payload: HookPayload = (
            cast(HookPayload, json.loads(payload_json)) if payload_json else {}
        )

        # Route to appropriate handler
        handlers = {
            "SessionStart": handle_session_start,
            "SessionEnd": handle_session_end,
            "PreToolUse": handle_pre_tool_use,
            "PostToolUse": handle_post_tool_use,
            "UserPromptSubmit": handle_user_prompt_submit,
            "Notification": handle_notification,
            "Stop": handle_stop,
            "SubagentStop": handle_subagent_stop,
            "PreCompact": handle_pre_compact,
        }

        handler = handlers.get(event)
        # Unknown event: return empty result (no-op)
        result = HookResult() if not handler else handler(payload)

        # Output JSON result
        print(result.to_json())
        sys.exit(0)

    except json.JSONDecodeError as e:
        print(f"JSON parse error: {e}", file=sys.stderr)
        sys.exit(1)

    except Exception as e:
        print(f"Hook execution error: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
