#!/usr/bin/env python3
"""
MoAI-ADK Statusline Runner

Wrapper script to run the statusline module.
Executes via: uv run .moai/scripts/statusline.py

Features:
- Tries to use installed moai-adk package for full statusline
- Falls back to .moai/config/config.json if package is not installed
- Displays version and project info even without moai-adk installed
"""

import json
import subprocess
import sys
from pathlib import Path


def fallback_statusline(cwd: str) -> None:
    """Display minimal statusline when moai-adk package is not installed

    Reads from .moai/config/config.json to show version and project info.
    """
    config_path = Path(cwd) / ".moai" / "config" / "config.json"

    try:
        if config_path.exists():
            config = json.loads(config_path.read_text())
            version = config.get("moai", {}).get("version", "unknown")
            project_name = config.get("project", {}).get("name", "MoAI-ADK")

            # Display minimal but informative status
            print(f"📦 Version: {version} (fallback mode)")
            print(f"🏗️  Project: {project_name}")
        else:
            print("⚠️  Config not found - Run moai-adk init first")
            sys.exit(1)
    except Exception as e:
        print(f"⚠️  Error reading config: {e}")
        sys.exit(1)


if __name__ == "__main__":
    # Get working directory from command line argument or use current directory
    cwd = sys.argv[1] if len(sys.argv) > 1 else "."

    # Try to run full statusline with moai-adk module
    result = subprocess.run(
        [sys.executable, "-m", "moai_adk.statusline.main"],
        cwd=cwd,
        capture_output=True,
        text=True,
    )

    if result.returncode != 0:
        # Module not found or error - fall back to config-based display
        if "No module named" in result.stderr or "ModuleNotFoundError" in result.stderr:
            fallback_statusline(cwd)
            sys.exit(0)
        else:
            # Unknown error - print it
            print(result.stderr, file=sys.stderr)
            sys.exit(result.returncode)
    else:
        # Success - output from moai_adk module
        print(result.stdout, end="")
        sys.exit(0)
