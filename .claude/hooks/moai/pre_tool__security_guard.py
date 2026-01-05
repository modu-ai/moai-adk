#!/usr/bin/env python3
"""PreToolUse Hook: Security Guard

Claude Code Event: PreToolUse
Matcher: Write|Edit
Purpose: Protect sensitive files and prevent dangerous modifications

Security Features:
- Block modifications to .env and secret files
- Protect lock files (package-lock.json, yarn.lock)
- Guard .git directory
- Prevent accidental overwrites of critical configs

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
from typing import Any, Dict, List, Tuple

# Patterns for files that should NEVER be modified
DENY_PATTERNS = [
    # Environment and secrets
    r"\.env$",
    r"\.env\.[^/]+$",  # .env.local, .env.production, etc.
    r"\.envrc$",
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


def compile_patterns(patterns: List[str]) -> List[re.Pattern]:
    """Compile regex patterns for efficient matching."""
    return [re.compile(p, re.IGNORECASE) for p in patterns]


DENY_COMPILED = compile_patterns(DENY_PATTERNS)
ASK_COMPILED = compile_patterns(ASK_PATTERNS)
SENSITIVE_COMPILED = compile_patterns(SENSITIVE_CONTENT_PATTERNS)


def check_file_path(file_path: str) -> Tuple[str, str]:
    """Check if file path matches any security patterns.

    Args:
        file_path: Path to check

    Returns:
        Tuple of (decision, reason)
        decision: "allow", "deny", or "ask"
    """
    # Normalize path for matching
    normalized = file_path.replace("\\", "/")

    # Check deny patterns first
    for pattern in DENY_COMPILED:
        if pattern.search(normalized):
            return "deny", "Protected file: access denied for security reasons"

    # Check ask patterns
    for pattern in ASK_COMPILED:
        if pattern.search(normalized):
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

    # Only process Write and Edit tools
    if tool_name not in ("Write", "Edit"):
        sys.exit(0)

    # Get file path from tool input
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
    output: Dict[str, Any] = {}

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
