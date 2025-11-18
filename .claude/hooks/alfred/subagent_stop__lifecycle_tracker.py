#!/usr/bin/env python3
"""Subagent Stop Lifecycle Tracker Hook

Tracks lifecycle events when subagents complete execution.
Handles cleanup, state persistence, and transition management.
"""

import json
import sys
from datetime import datetime
from pathlib import Path


def get_lifecycle_state_file(cwd: str) -> Path:
    """Get the path to lifecycle state tracking file"""
    state_dir = Path(cwd) / ".moai" / "memory"
    state_dir.mkdir(parents=True, exist_ok=True)
    return state_dir / "subagent-lifecycle-state.json"


def load_lifecycle_state(cwd: str) -> dict:
    """Load current subagent lifecycle state"""
    try:
        state_file = get_lifecycle_state_file(cwd)
        if state_file.exists():
            with open(state_file, "r", encoding="utf-8") as f:
                return json.load(f)
    except Exception:
        pass
    return {
        "last_subagent": None,
        "last_completion": None,
        "total_executions": 0,
        "session_start": datetime.now().isoformat(),
    }


def save_lifecycle_state(cwd: str, state: dict) -> None:
    """Save subagent lifecycle state"""
    try:
        state_file = get_lifecycle_state_file(cwd)
        with open(state_file, "w", encoding="utf-8") as f:
            json.dump(state, f, indent=2)
    except Exception:
        pass


def handle_subagent_stop(payload: dict) -> dict:
    """Handle subagent stop event

    Tracks when a subagent completes execution and updates lifecycle state.
    """
    try:
        cwd = payload.get("cwd", ".")
        subagent_type = payload.get("subagent_type", "unknown")
        status = payload.get("status", "unknown")

        # Load current state
        state = load_lifecycle_state(cwd)

        # Update with new information
        state["last_subagent"] = subagent_type
        state["last_completion"] = datetime.now().isoformat()
        state["last_status"] = status
        state["total_executions"] = state.get("total_executions", 0) + 1

        # Save updated state
        save_lifecycle_state(cwd, state)

        return {
            "status": "success",
            "message": f"Tracked subagent stop: {subagent_type}",
            "state": state,
        }
    except Exception as e:
        # Graceful degradation: log but don't fail
        return {
            "status": "warning",
            "message": f"Error tracking subagent lifecycle: {str(e)}",
        }


def main():
    """Main hook entry point"""
    try:
        # Read hook payload from stdin
        payload = json.loads(sys.stdin.read())

        # Process subagent stop event
        result = handle_subagent_stop(payload)

        # Output result
        print(json.dumps(result), file=sys.stdout)
        sys.exit(0)

    except Exception as e:
        # Graceful error handling
        error_result = {
            "status": "error",
            "message": f"Hook execution failed: {str(e)}",
        }
        print(json.dumps(error_result), file=sys.stdout)
        sys.exit(0)  # Exit 0 to prevent hook failure from blocking execution


if __name__ == "__main__":
    main()
