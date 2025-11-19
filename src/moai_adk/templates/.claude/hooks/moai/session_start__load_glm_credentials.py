#!/usr/bin/env python3
# SessionStart Hook: Load GLM Credentials
"""SessionStart Hook: Load GLM Credentials

Claude Code Event: SessionStart
Purpose: Load GLM API credentials from .env.glm and environment variables from settings.local.json
Execution: Triggered automatically when Claude Code session begins

Behavior:
1. Check if .env.glm exists (GLM API token)
2. Check if .claude/settings.local.json has env section (GLM model configuration)
3. Load all environment variables into the session
4. Silently continue if GLM is not configured
"""

import json
import os
import sys
from pathlib import Path
from typing import Any


def load_glm_credentials() -> None:
    """Load GLM credentials from .env.glm and settings.local.json into environment"""

    project_root = Path.cwd()

    try:
        # 1. Load .env.glm (API token)
        env_glm_path = project_root / ".env.glm"
        if env_glm_path.exists():
            try:
                content = env_glm_path.read_text().strip()
                for line in content.split("\n"):
                    if "=" in line and not line.startswith("#"):
                        key, value = line.split("=", 1)
                        os.environ[key.strip()] = value.strip()
            except Exception:
                # Silently ignore if .env.glm read fails
                pass

        # 2. Load environment variables from settings.local.json
        settings_path = project_root / ".claude" / "settings.local.json"
        if settings_path.exists():
            try:
                with open(settings_path, "r", encoding="utf-8") as f:
                    settings = json.load(f)

                # Load env section if present
                if "env" in settings:
                    for key, value in settings["env"].items():
                        # Skip placeholder values
                        if value and "{{" not in value:
                            os.environ[key] = value
            except Exception:
                # Silently ignore if settings.local.json read fails
                pass

    except Exception:
        # Silently ignore any unexpected errors
        pass


def main() -> None:
    """Main entry point for GLM credentials hook

    Exit Codes:
        0: Success (with or without GLM configured)
        1: Unexpected error (shouldn't happen)
    """
    try:
        # Read JSON payload from stdin (for compatibility)
        input_data = sys.stdin.read()
        data = json.loads(input_data) if input_data.strip() else {}

        # Load GLM credentials silently
        load_glm_credentials()

        # Return success response (continue without system message)
        result: dict[str, Any] = {
            "continue": True
        }

        print(json.dumps(result))
        sys.exit(0)

    except Exception as e:
        # Return success anyway (don't block session on hook error)
        result: dict[str, Any] = {
            "continue": True
        }

        print(json.dumps(result))
        print(f"GLM credentials hook error (non-blocking): {e}", file=sys.stderr)
        sys.exit(0)


if __name__ == "__main__":
    main()
