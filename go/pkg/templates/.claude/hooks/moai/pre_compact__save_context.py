#!/usr/bin/env python3
"""PreCompact Hook: Save Context Before Clear or Auto Compact

Claude Code Event: PreCompact
Purpose: Placeholder for future context preservation features
Execution: Triggered automatically before context compaction

Note: Session context storage has been removed as hooks cannot access
Claude's internal TaskList state. This hook is kept as a placeholder
for potential future features.
"""

from __future__ import annotations

import json
import sys
from typing import Any

# =============================================================================
# Windows UTF-8 Encoding Fix (Issue #249)
# =============================================================================
if sys.platform == "win32":
    try:
        sys.stdout.reconfigure(encoding="utf-8", errors="replace")
        sys.stderr.reconfigure(encoding="utf-8", errors="replace")
    except (AttributeError, OSError):
        pass


def execute_pre_compact() -> dict[str, Any]:
    """Execute the pre-compact workflow.

    Currently a placeholder - no context is saved.
    """
    # Read hook payload from stdin (consume input)
    _ = json.loads(sys.stdin.read() if not sys.stdin.isatty() else "{}")

    return {
        "continue": True,
        "systemMessage": "Pre-compact: proceeding without save",
    }


def main() -> None:
    """Main entry point for PreCompact hook.

    Exit Codes:
        0: Success
        1: Error (exception)
    """
    try:
        result = execute_pre_compact()
        print(json.dumps(result, ensure_ascii=False))
        sys.exit(0)
    except Exception as e:
        error_response = {
            "continue": True,
            "systemMessage": f"Pre-compact hook error: {e}",
        }
        print(json.dumps(error_response, ensure_ascii=False))
        sys.exit(1)


if __name__ == "__main__":
    main()
