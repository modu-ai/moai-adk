#!/usr/bin/env python3
"""PreToolUse Hook: Security Guard

Claude Code Event: PreToolUse
Matcher: Write|Edit|Bash
Purpose: Protect sensitive files and prevent dangerous operations

Security Features:
- Block modifications to secret/credential files
- Protect lock files (package-lock.json, yarn.lock)
- Guard .git directory
- Prevent accidental overwrites of critical configs
- Block dangerous database deletion commands (Supabase, Neon, etc.)
- Block dangerous file deletion commands (rm -rf critical paths)
- Support Claude Code Plan Mode with configurable plans directory

Exit Codes:
- 0: Success (with JSON output for permission decision)
- 2: Error (blocking the operation)

Permission Decisions:
- "allow": Auto-approve safe operations
- "deny": Block dangerous operations (reason shown to Claude)
- "ask": Request user confirmation
"""

import json
import re
import sys
from pathlib import Path
from typing import Any, List, Tuple

# Patterns for files that should NEVER be modified
DENY_PATTERNS = [
    # Secrets and credentials (NOT .env - developers need to edit these)
    r"secrets?\.(json|ya?ml|toml)$",
    r"credentials?\.(json|ya?ml|toml)$",
    r"\.secrets/.*",
    r"secrets/.*",
    # SSH and certificates
    r"\.ssh/.*",
    r"id_rsa.*",
    r"id_ed25519.*",
    r"\.pem$",
    r"\.key$",
    r"\.crt$",
    # Git internals
    r"\.git/.*",
    # Cloud credentials
    r"\.aws/.*",
    r"\.gcloud/.*",
    r"\.azure/.*",
    r"\.kube/.*",
    # Token files
    r"\.token$",
    r"\.tokens/.*",
    r"auth\.json$",
]

# Dangerous Bash commands that should NEVER be executed (Unix + Windows)
DANGEROUS_BASH_PATTERNS = [
    # Database deletion commands - Supabase
    r"supabase\s+db\s+reset",
    r"supabase\s+projects?\s+delete",
    r"supabase\s+functions?\s+delete",
    # Database deletion commands - Neon
    r"neon\s+database\s+delete",
    r"neon\s+projects?\s+delete",
    r"neon\s+branch\s+delete",
    # Database deletion commands - PlanetScale
    r"pscale\s+database\s+delete",
    r"pscale\s+branch\s+delete",
    # Database deletion commands - Railway
    r"railway\s+delete",
    r"railway\s+environment\s+delete",
    # Database deletion commands - Vercel
    r"vercel\s+env\s+rm",
    r"vercel\s+projects?\s+rm",
    # SQL dangerous commands
    r"DROP\s+DATABASE",
    r"DROP\s+SCHEMA",
    r"TRUNCATE\s+TABLE",
    # === Unix dangerous file operations ===
    r"rm\s+-rf\s+/",  # rm -rf /
    r"rm\s+-rf\s+~",  # rm -rf ~
    r"rm\s+-rf\s+\*",  # rm -rf *
    r"rm\s+-rf\s+\.\*",  # rm -rf .*
    r"rm\s+-rf\s+\.git\b",  # rm -rf .git
    r"rm\s+-rf\s+node_modules\s*$",  # rm -rf node_modules (without reinstall intent)
    # === Windows dangerous file operations (CMD) ===
    r"rd\s+/s\s+/q\s+[A-Za-z]:\\",  # rd /s /q C:\
    r"rmdir\s+/s\s+/q\s+[A-Za-z]:\\",  # rmdir /s /q C:\
    r"del\s+/f\s+/q\s+[A-Za-z]:\\",  # del /f /q C:\
    r"rd\s+/s\s+/q\s+\\\\",  # rd /s /q \\ (network paths)
    r"rd\s+/s\s+/q\s+\.git\b",  # rd /s /q .git
    r"del\s+/s\s+/q\s+\*\.\*",  # del /s /q *.*
    r"format\s+[A-Za-z]:",  # format C:
    # === Windows dangerous file operations (PowerShell) ===
    r"Remove-Item\s+.*-Recurse\s+.*-Force\s+[A-Za-z]:\\",  # Remove-Item -Recurse -Force C:\
    r"Remove-Item\s+.*-Recurse\s+.*-Force\s+~",  # Remove-Item -Recurse -Force ~
    r"Remove-Item\s+.*-Recurse\s+.*-Force\s+\$env:",  # Remove-Item with env vars
    r"Remove-Item\s+.*-Recurse\s+.*-Force\s+\.git\b",  # Remove-Item .git
    r"Clear-Content\s+.*-Force",  # Clear-Content -Force (file wipe)
    # Git dangerous commands (cross-platform)
    r"git\s+push\s+.*--force\s+origin\s+(main|master)",
    r"git\s+branch\s+-D\s+(main|master)",
    # Cloud infrastructure deletion
    r"terraform\s+destroy",
    r"pulumi\s+destroy",
    r"aws\s+.*\s+delete-",
    r"gcloud\s+.*\s+delete\b",
    # Azure CLI dangerous commands
    r"az\s+group\s+delete",
    r"az\s+storage\s+account\s+delete",
    r"az\s+sql\s+server\s+delete",
    # Docker dangerous commands
    r"docker\s+system\s+prune\s+(-a|--all)",  # docker system prune -a
    r"docker\s+image\s+prune\s+(-a|--all)",  # docker image prune -a
    r"docker\s+container\s+prune",  # docker container prune
    r"docker\s+volume\s+prune",  # docker volume prune (data loss risk)
    r"docker\s+network\s+prune",  # docker network prune
    r"docker\s+builder\s+prune\s+(-a|--all)",  # docker builder prune -a
]

# Bash commands that require user confirmation
ASK_BASH_PATTERNS = [
    # Database reset/migration (may be intentional)
    r"prisma\s+migrate\s+reset",
    r"prisma\s+db\s+push\s+--force",
    r"drizzle-kit\s+push",
    # Git force operations (non-main branches)
    r"git\s+push\s+.*--force",
    r"git\s+reset\s+--hard",
    r"git\s+clean\s+-fd",
    # Package manager cache clear
    r"npm\s+cache\s+clean",
    r"yarn\s+cache\s+clean",
    r"pnpm\s+store\s+prune",
]

# Patterns for files that require user confirmation
ASK_PATTERNS = [
    # Lock files
    r"package-lock\.json$",
    r"yarn\.lock$",
    r"pnpm-lock\.ya?ml$",
    r"Gemfile\.lock$",
    r"Cargo\.lock$",
    r"poetry\.lock$",
    r"composer\.lock$",
    r"Pipfile\.lock$",
    r"uv\.lock$",
    # Critical configs
    r"tsconfig\.json$",
    r"pyproject\.toml$",
    r"Cargo\.toml$",
    r"package\.json$",
    r"docker-compose\.ya?ml$",
    r"Dockerfile$",
    r"\.dockerignore$",
    # CI/CD configs
    r"\.github/workflows/.*\.ya?ml$",
    r"\.gitlab-ci\.ya?ml$",
    r"\.circleci/.*",
    r"Jenkinsfile$",
    # Infrastructure
    r"terraform/.*\.tf$",
    r"\.terraform/.*",
    r"kubernetes/.*\.ya?ml$",
    r"k8s/.*\.ya?ml$",
]

# Content patterns that indicate sensitive data
SENSITIVE_CONTENT_PATTERNS = [
    r"-----BEGIN\s+(RSA\s+)?PRIVATE\s+KEY-----",
    r"-----BEGIN\s+CERTIFICATE-----",
    r"sk-[a-zA-Z0-9]{32,}",  # OpenAI API keys
    r"ghp_[a-zA-Z0-9]{36}",  # GitHub tokens
    r"gho_[a-zA-Z0-9]{36}",  # GitHub OAuth tokens
    r"glpat-[a-zA-Z0-9\-]{20}",  # GitLab tokens
    r"xox[baprs]-[a-zA-Z0-9\-]+",  # Slack tokens
    r"AKIA[0-9A-Z]{16}",  # AWS access keys
    r"ya29\.[a-zA-Z0-9_\-]+",  # Google OAuth tokens
]


def compile_patterns(patterns: list[str]) -> List[re.Pattern]:
    """Compile regex patterns for efficient matching."""
    return [re.compile(p, re.IGNORECASE) for p in patterns]


DENY_COMPILED = compile_patterns(DENY_PATTERNS)
ASK_COMPILED = compile_patterns(ASK_PATTERNS)
SENSITIVE_COMPILED = compile_patterns(SENSITIVE_CONTENT_PATTERNS)
DANGEROUS_BASH_COMPILED = compile_patterns(DANGEROUS_BASH_PATTERNS)
ASK_BASH_COMPILED = compile_patterns(ASK_BASH_PATTERNS)


def get_project_root() -> Path:
    """Get the project root directory from environment or current working directory.

    Returns:
        Path to the project root directory.
    """
    import os

    project_dir = os.environ.get("CLAUDE_PROJECT_DIR", os.getcwd())
    return Path(project_dir).resolve()


def get_plans_directory() -> Path | None:
    """Get the configured plans directory from Claude Code settings.

    Reads from .claude/settings.json in the project root.
    Returns None if not configured or file doesn't exist.

    Returns:
        Path to plans directory if configured, None otherwise
    """

    project_root = get_project_root()
    settings_file = project_root / ".claude" / "settings.json"

    if not settings_file.exists():
        return None

    try:
        with open(settings_file, "r") as f:
            settings = json.load(f)
            plans_dir = settings.get("plansDirectory")
            if plans_dir:
                # Resolve relative to project root
                return (project_root / plans_dir).resolve()
    except (json.JSONDecodeError, OSError):
        pass

    return None


def check_file_path(file_path: str) -> Tuple[str, str]:
    """Check if file path matches any security patterns.

    Security measures:
    - Resolves symlinks and '..' components to prevent path traversal
    - Checks both original and resolved paths against patterns
    - Validates path is within project boundaries or configured plans directory

    Args:
        file_path: Path to check

    Returns:
        Tuple of (decision, reason)
        decision: "allow", "deny", or "ask"
    """
    # Resolve path to prevent path traversal attacks
    # This handles: symlinks, '..' components, and normalizes the path
    try:
        resolved_path = Path(file_path).resolve()
        resolved_str = str(resolved_path)
    except (OSError, ValueError):
        # If path resolution fails, deny for safety
        return "deny", "Invalid file path: cannot resolve"

    # Normalize original path for pattern matching (keeps relative structure visible)
    normalized_original = file_path.replace("\\", "/")
    normalized_resolved = resolved_str.replace("\\", "/")

    # Check project boundary and plans directory
    project_root = get_project_root()
    is_within_project = False
    is_within_plans_dir = False

    # Check if path is within project directory
    try:
        resolved_path.relative_to(project_root)
        is_within_project = True
    except ValueError:
        pass

    # Check if path is within configured plans directory
    plans_dir = get_plans_directory()
    if plans_dir:
        try:
            resolved_path.relative_to(plans_dir)
            is_within_plans_dir = True
        except ValueError:
            pass

    # If path is outside both project and plans directory, deny
    if not is_within_project and not is_within_plans_dir:
        return "deny", "Path traversal detected: file is outside project directory"

    # Check deny patterns against BOTH original and resolved paths
    # This catches attempts like ".env/../other.txt" (matches .env in original)
    # AND "/absolute/path/to/.env" (matches in resolved)
    for pattern in DENY_COMPILED:
        if pattern.search(normalized_original) or pattern.search(normalized_resolved):
            return "deny", "Protected file: access denied for security reasons"

    # Check ask patterns against both paths
    for pattern in ASK_COMPILED:
        if pattern.search(normalized_original) or pattern.search(normalized_resolved):
            return "ask", f"Critical config file: {Path(file_path).name}"

    return "allow", ""


def check_content_for_secrets(content: str) -> Tuple[bool, str]:
    """Check if content contains sensitive data patterns.

    Args:
        content: Content to check

    Returns:
        Tuple of (has_secrets, description)
    """
    for pattern in SENSITIVE_COMPILED:
        match = pattern.search(content)
        if match:
            # Don't reveal the actual secret or pattern
            return True, "Detected sensitive data (credentials, API keys, or certificates)"

    return False, ""


def check_bash_command(command: str) -> Tuple[str, str]:
    """Check if Bash command is dangerous or requires confirmation.

    Args:
        command: Bash command to check

    Returns:
        Tuple of (decision, reason)
        decision: "allow", "deny", or "ask"
    """
    # Check for dangerous commands that should be blocked
    for pattern in DANGEROUS_BASH_COMPILED:
        if pattern.search(command):
            return "deny", f"Dangerous command blocked: {pattern.pattern}"

    # Check for commands that require user confirmation
    for pattern in ASK_BASH_COMPILED:
        if pattern.search(command):
            return "ask", "This command may have significant effects. Please confirm."

    return "allow", ""


def main() -> None:
    """Main entry point for PreToolUse security guard hook.

    Reads JSON input from stdin, checks for security concerns,
    and outputs permission decision.
    """
    try:
        # Read JSON input from stdin
        input_data = json.load(sys.stdin)
    except json.JSONDecodeError:
        # Invalid JSON input - allow by default
        sys.exit(0)

    # Extract tool information
    tool_name = input_data.get("tool_name", "")
    tool_input = input_data.get("tool_input", {})

    # Only process Write, Edit, and Bash tools
    if tool_name not in ("Write", "Edit", "Bash"):
        sys.exit(0)

    decision = "allow"
    reason = ""

    # Handle Bash commands
    if tool_name == "Bash":
        command = tool_input.get("command", "")
        if command:
            decision, reason = check_bash_command(command)
    else:
        # Handle Write and Edit tools
        file_path = tool_input.get("file_path", "")
        if not file_path:
            sys.exit(0)

        # Check file path against patterns
        decision, reason = check_file_path(file_path)

        # For Write operations, also check content for secrets
        if tool_name == "Write" and decision == "allow":
            content = tool_input.get("content", "")
            if content:
                has_secrets, secret_reason = check_content_for_secrets(content)
                if has_secrets:
                    decision = "deny"
                    reason = f"Content contains secrets: {secret_reason}"

    # Build output based on decision
    output: dict[str, Any] = {}

    if decision == "deny":
        output = {
            "hookSpecificOutput": {
                "hookEventName": "PreToolUse",
                "permissionDecision": "deny",
                "permissionDecisionReason": reason,
            }
        }
        print(json.dumps(output))
        sys.exit(0)

    elif decision == "ask":
        output = {
            "hookSpecificOutput": {
                "hookEventName": "PreToolUse",
                "permissionDecision": "ask",
                "permissionDecisionReason": reason,
            }
        }
        print(json.dumps(output))
        sys.exit(0)

    else:
        # Allow - no output needed (or suppress)
        output = {"suppressOutput": True}
        print(json.dumps(output))
        sys.exit(0)


if __name__ == "__main__":
    main()
