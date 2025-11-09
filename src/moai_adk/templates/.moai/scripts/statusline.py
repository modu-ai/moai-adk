#!/usr/bin/env python3
"""
MoAI-ADK Statusline Runner

Wrapper script to run the statusline module.
Executes via: uv run .moai/scripts/statusline.py

@CODE:STATUSLINE-RUNNER-001
"""

import subprocess
import sys

if __name__ == "__main__":
    result = subprocess.run(
        [sys.executable, "-m", "moai_adk.statusline.main"],
        cwd=sys.argv[1] if len(sys.argv) > 1 else ".",
    )
    sys.exit(result.returncode)
