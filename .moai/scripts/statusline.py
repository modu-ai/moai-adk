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
import os
import subprocess
import sys
from pathlib import Path


def get_moai_project_root() -> Path | None:
    """
    Find the MoAI-ADK project root directory.

    This script is at: .moai/scripts/statusline.py
    We need to find the project root that has .moai/config/config.json

    Returns:
        Project root path if found, None otherwise
    """
    try:
        script_path = Path(__file__).resolve()
        # From .moai/scripts/ to project root: go up 3 levels
        # .moai/scripts/statusline.py -> .moai/ -> project_root/
        potential_root = script_path.parent.parent.parent

        # Verify this is a MoAI-ADK project by checking for .moai/config/config.json
        config_path = potential_root / ".moai" / "config" / "config.json"
        if config_path.exists():
            return potential_root

        return None
    except Exception:
        return None


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
    # Get working directory from:
    # 1. Command line argument
    # 2. Environment variable (Claude Code sets this)
    # 3. Current directory
    cwd = sys.argv[1] if len(sys.argv) > 1 else os.environ.get("CLAUDE_PROJECT_DIR", ".")

    # Find MoAI-ADK project root to use its environment
    project_root = get_moai_project_root()

    if project_root:
        # Use uv run with explicit project context to ensure moai-adk package is available
        # This is important when this script is called from a different directory via Claude Code
        result = subprocess.run(
            ["uv", "run", "--project", str(project_root), "python", "-m", "moai_adk.statusline.main"],
            cwd=cwd,
            capture_output=True,
            text=True,
        )
    else:
        # Fall back to direct execution if project root not found
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
