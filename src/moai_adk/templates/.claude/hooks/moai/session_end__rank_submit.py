#!/usr/bin/env python3
"""MoAI Rank Session Hook

This hook submits Claude Code session token usage to the MoAI Rank service.
It is triggered automatically when a session ends.

The hook reads session data from stdin and submits it to the rank service
if the user has registered with MoAI Rank.

Requirements:
- User must be registered: Run `moai-adk rank register` to connect GitHub account
- API key stored securely in ~/.moai/rank/credentials.json

Privacy:
- Only token counts are submitted (input, output, cache tokens)
- Project paths are anonymized using one-way hashing
- No code or conversation content is transmitted
"""

import json
import sys


def main():
    """Main hook entry point."""
    try:
        # Read session data from stdin
        input_data = sys.stdin.read()
        if not input_data:
            return

        session_data = json.loads(input_data)

        # Skip if no token usage data
        input_tokens = session_data.get("inputTokens", 0)
        output_tokens = session_data.get("outputTokens", 0)
        if input_tokens == 0 and output_tokens == 0:
            return

        # Lazy import to avoid startup delay
        try:
            from moai_adk.rank.hook import submit_session_hook
        except ImportError:
            # moai-adk not installed or rank module not available
            return

        result = submit_session_hook(session_data)

        if result["success"]:
            print("Session submitted to MoAI Rank", file=sys.stderr)
        elif result["message"] and result["message"] != "Not registered with MoAI Rank":
            print(f"MoAI Rank: {result['message']}", file=sys.stderr)

    except json.JSONDecodeError:
        # Invalid JSON input, silently skip
        pass
    except Exception as e:
        # Log errors but don't fail the hook
        print(f"MoAI Rank hook error: {e}", file=sys.stderr)


if __name__ == "__main__":
    main()
