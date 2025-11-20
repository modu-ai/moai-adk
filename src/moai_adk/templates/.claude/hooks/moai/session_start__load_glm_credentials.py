#!/usr/bin/env python3
"""GLM Credentials Loader Hook for Claude Code Session Start

Automatically loads GLM API credentials from .env.glm file at session start.
Enables seamless GLM integration without manual token re-entry.

Execution: Triggered on every SessionStart hook
Environment: Reads from .env.glm (created by setup-glm.py)
Output: Loads ANTHROPIC_AUTH_TOKEN into Claude Code environment
"""

import os
from pathlib import Path


def load_glm_credentials() -> bool:
    """Load GLM credentials from .env.glm file

    Returns:
        True if credentials loaded successfully, False otherwise
    """
    env_glm_path = Path(".env.glm")

    # Check if .env.glm exists
    if not env_glm_path.exists():
        return False

    try:
        # Read API token from .env.glm
        with open(env_glm_path, "r", encoding="utf-8") as f:
            content = f.read().strip()

        # Parse ANTHROPIC_AUTH_TOKEN=<token>
        if "ANTHROPIC_AUTH_TOKEN=" in content:
            token = content.split("ANTHROPIC_AUTH_TOKEN=")[1].strip()

            # Set environment variable for Claude Code
            os.environ["ANTHROPIC_AUTH_TOKEN"] = token

            # Optional: Log successful load (can be removed for silent operation)
            # print(f"✅ GLM credentials loaded from {env_glm_path}")

            return True

        return False

    except Exception:
        # Silent failure: GLM not configured, use default Claude models
        # print(f"⚠️  Error loading GLM credentials: {e}")
        return False


def main() -> None:
    """Main entry point for SessionStart hook"""
    load_glm_credentials()


if __name__ == "__main__":
    main()
