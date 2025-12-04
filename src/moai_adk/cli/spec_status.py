#!/usr/bin/env python3
"""Spec Status Manager CLI Tool

Provides command-line interface for managing SPEC status tracking.

Usage:
    python3 -m moai_adk.cli.spec_status <command> <spec_id> [options]
    moai-spec <command> <spec_id> [options]

Commands:
    - status_update: Update SPEC status
    - validate_completion: Validate if SPEC is ready for completion
    - batch_update: Update all completed SPECs
"""

import argparse
import json
import sys
from datetime import datetime
from pathlib import Path
from typing import Any, Dict

try:
    from moai_adk.core.spec_status_manager import SpecStatusManager
except ImportError:
    # Fallback: look in current directory
    import sys
    from pathlib import Path

    sys.path.insert(0, str(Path(__file__).parent.parent))
    try:
        from core.spec_status_manager import SpecStatusManager  # type: ignore[no-redef]
    except ImportError:
        raise ImportError("SpecStatusManager not found")


def update_spec_status(spec_id: str, new_status: str, reason: str = "") -> Dict[str, Any]:
    """Update SPEC status with validation and logging

    Args:
        spec_id: The SPEC identifier
        new_status: New status value
        reason: Reason for status change

    Returns:
        Update result dictionary
    """
    try:
        # Initialize manager
        project_root = Path.cwd()
        manager = SpecStatusManager(project_root)

        # Validate new status
        valid_statuses = ["draft", "in-progress", "completed", "archived"]
        if new_status not in valid_statuses:
            return {
                "success": False,
                "error": f"Invalid status: {new_status}. Valid statuses: {valid_statuses}",
            }

        # Check if SPEC exists
        spec_file = project_root / ".moai" / "specs" / spec_id / "spec.md"
        if not spec_file.exists():
            return {"success": False, "error": f"SPEC file not found: {spec_file}"}

        # Update status
        success = manager.update_spec_status(spec_id, new_status)

        if success:
            # Log the status change
            log_entry = {
                "timestamp": datetime.now().isoformat(),
                "spec_id": spec_id,
                "old_status": "unknown",  # Could be extracted from file before update
                "new_status": new_status,
                "reason": reason,
            }

            # Create status change log
            log_dir = project_root / ".moai" / "logs"
            log_dir.mkdir(parents=True, exist_ok=True)
            log_file = log_dir / "spec_status_changes.jsonl"

            with open(log_file, "a", encoding="utf-8") as f:
                f.write(json.dumps(log_entry, ensure_ascii=False) + "\n")

            return {
                "success": True,
                "spec_id": spec_id,
                "new_status": new_status,
                "reason": reason,
                "timestamp": log_entry["timestamp"],
            }
        else:
            return {
                "success": False,
                "error": f"Failed to update SPEC {spec_id} status",
            }

    except Exception as e:
        return {"success": False, "error": f"Error updating SPEC status: {str(e)}"}


def validate_spec_completion(spec_id: str) -> Dict[str, Any]:
    """Validate if SPEC is ready for completion

    Args:
        spec_id: The SPEC identifier

    Returns:
        Validation result dictionary
    """
    try:
        project_root = Path.cwd()
        manager = SpecStatusManager(project_root)

        # Run validation
        validation = manager.validate_spec_for_completion(spec_id)

        return {"success": True, "spec_id": spec_id, "validation": validation}

    except Exception as e:
        return {
            "success": False,
            "error": f"Error validating SPEC completion: {str(e)}",
        }


def batch_update_completed_specs() -> Dict[str, Any]:
    """Batch update all draft SPECs that have completed implementations

    Returns:
        Batch update result dictionary
    """
    try:
        project_root = Path.cwd()
        manager = SpecStatusManager(project_root)

        # Run batch update
        results = manager.batch_update_completed_specs()

        # Log the batch update
        log_entry = {
            "timestamp": datetime.now().isoformat(),
            "operation": "batch_update_completed",
            "results": results,
        }

        log_dir = project_root / ".moai" / "logs"
        log_dir.mkdir(parents=True, exist_ok=True)
        log_file = log_dir / "spec_status_changes.jsonl"

        with open(log_file, "a", encoding="utf-8") as f:
            f.write(json.dumps(log_entry, ensure_ascii=False) + "\n")

        return {
            "success": True,
            "results": results,
            "timestamp": log_entry["timestamp"],
        }

    except Exception as e:
        return {"success": False, "error": f"Error in batch update: {str(e)}"}


def detect_draft_specs() -> Dict[str, Any]:
    """Detect all draft SPECs in the project

    Returns:
        List of draft SPEC IDs
    """
    try:
        project_root = Path.cwd()
        manager = SpecStatusManager(project_root)

        draft_specs = manager.detect_draft_specs()

        return {
            "success": True,
            "draft_specs": list(draft_specs),
            "count": len(draft_specs),
        }

    except Exception as e:
        return {"success": False, "error": f"Error detecting draft SPECs: {str(e)}"}


def main():
    """Main function"""
    parser = argparse.ArgumentParser(description="Spec Status Manager Hooks")
    parser.add_argument(
        "command",
        choices=[
            "status_update",
            "validate_completion",
            "batch_update",
            "detect_drafts",
        ],
        help="Command to execute",
    )
    parser.add_argument("spec_id", nargs="?", help="SPEC ID (for specific commands)")
    parser.add_argument("--status", help="New status value")
    parser.add_argument("--reason", default="", help="Reason for status change")

    args = parser.parse_args()

    try:
        result = {}

        if args.command == "status_update":
            if not args.spec_id or not args.status:
                print(
                    json.dumps(
                        {
                            "success": False,
                            "error": "status_update requires spec_id and --status",
                        }
                    )
                )
                sys.exit(1)

            result = update_spec_status(args.spec_id, args.status, args.reason)

        elif args.command == "validate_completion":
            if not args.spec_id:
                print(
                    json.dumps(
                        {
                            "success": False,
                            "error": "validate_completion requires spec_id",
                        }
                    )
                )
                sys.exit(1)

            result = validate_spec_completion(args.spec_id)

        elif args.command == "batch_update":
            result = batch_update_completed_specs()

        elif args.command == "detect_drafts":
            result = detect_draft_specs()

        # Add command info
        result["command"] = args.command
        result["execution_time"] = datetime.now().isoformat()

        print(json.dumps(result, ensure_ascii=False, indent=2))

    except Exception as e:
        error_result = {
            "success": False,
            "command": args.command if "args" in locals() else "unknown",
            "error": str(e),
            "execution_time": datetime.now().isoformat(),
        }

        print(json.dumps(error_result, ensure_ascii=False, indent=2))
        sys.exit(1)


if __name__ == "__main__":
    main()
