#!/usr/bin/env python3
"""
Universal Python wrapper for MoAI-ADK statusline.
This works on all platforms with Python 3.6+.
"""

import json
import os
import subprocess
import sys
from pathlib import Path


def find_project_root():
    """Find the project root directory."""
    current = Path(__file__).parent.parent.parent.resolve()
    return current


def try_uv_execution(project_root, input_json):
    """Try to execute with uv."""
    try:
        result = subprocess.run(
            ["uv", "run", "python", "-m", "moai_adk.statusline.main"],
            input=input_json,
            capture_output=True,
            text=True,
            cwd=str(project_root),
            timeout=2
        )
        if result.returncode == 0 and result.stdout:
            return result.stdout
    except (subprocess.SubprocessError, FileNotFoundError, OSError):
        pass
    return None


def try_module_execution(input_json):
    """Try to execute as Python module."""
    try:
        result = subprocess.run(
            [sys.executable, "-m", "moai_adk.statusline.main"],
            input=input_json,
            capture_output=True,
            text=True,
            timeout=2
        )
        if result.returncode == 0 and result.stdout:
            return result.stdout
    except (subprocess.SubprocessError, FileNotFoundError, OSError):
        pass
    return None


def try_direct_execution(project_root, input_json):
    """Try direct script execution."""
    main_py = project_root / "src" / "moai_adk" / "statusline" / "main.py"
    if main_py.exists():
        try:
            env = os.environ.copy()
            pythonpath = str(project_root / "src")
            if "PYTHONPATH" in env:
                pythonpath = f"{pythonpath}{os.pathsep}{env['PYTHONPATH']}"
            env["PYTHONPATH"] = pythonpath

            result = subprocess.run(
                [sys.executable, str(main_py)],
                input=input_json,
                capture_output=True,
                text=True,
                env=env,
                timeout=2
            )
            if result.returncode == 0 and result.stdout:
                return result.stdout
        except (subprocess.SubprocessError, FileNotFoundError, OSError):
            pass
    return None


def fallback_statusline(input_json, project_root):
    """Generate a minimal fallback statusline."""
    try:
        data = json.loads(input_json) if input_json else {}
    except json.JSONDecodeError:
        data = {}

    # Extract model
    model = "Unknown"
    if "model" in data:
        model_info = data["model"]
        if isinstance(model_info, dict):
            model = model_info.get("display_name") or model_info.get("name") or "Unknown"
        else:
            model = str(model_info)

    # Try to get version from config
    version = "unknown"
    config_file = project_root / ".moai" / "config" / "config.json"
    if config_file.exists():
        try:
            with open(config_file, 'r') as f:
                config = json.load(f)
                version = config.get("moai", {}).get("version", "unknown")
        except (json.JSONDecodeError, IOError):
            pass

    # Try to get git branch
    branch = ""
    try:
        result = subprocess.run(
            ["git", "branch", "--show-current"],
            capture_output=True,
            text=True,
            cwd=str(project_root),
            timeout=1
        )
        if result.returncode == 0:
            branch = result.stdout.strip()
    except:
        pass

    # Build statusline
    parts = [f"🤖 {model}", f"🗿 Ver {version}"]
    if branch:
        parts.append(f"🔀 {branch}")

    return " | ".join(parts)


def main():
    """Main entry point."""
    # Read input from stdin
    try:
        input_json = sys.stdin.read()
    except:
        input_json = ""

    # Find project root
    project_root = find_project_root()

    # Try different execution methods in order
    result = None

    # 1. Try uv
    result = try_uv_execution(project_root, input_json)

    # 2. Try Python module
    if not result:
        result = try_module_execution(input_json)

    # 3. Try direct execution
    if not result:
        result = try_direct_execution(project_root, input_json)

    # 4. Fallback
    if not result:
        result = fallback_statusline(input_json, project_root)

    # Output result
    if result:
        print(result, end="")


if __name__ == "__main__":
    main()