#!/usr/bin/env python3
"""
MoAI-ADK Statusline Runner

Wrapper script to run the statusline module.
Executes via: uv run .moai/scripts/statusline.py

"""

import subprocess
import sys

if __name__ == "__main__":
    # Get working directory from command line argument or use current directory
    cwd = sys.argv[1] if len(sys.argv) > 1 else "."
    result = subprocess.run(
        [sys.executable, "-m", "moai_adk.statusline.main"],
        cwd=cwd,
    )
    sys.exit(result.returncode)
