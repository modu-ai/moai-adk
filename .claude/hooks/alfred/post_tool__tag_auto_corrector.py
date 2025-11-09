#!/usr/bin/env python3
# @CODE:HOOK-POST-TAG-001 | SPEC: TAG-POST-HOOK-001 | TEST: tests/hooks/test_post_tool_tag_corrector.py
"""PostToolUse Hook: Auto TAG correction and monitoring

Validates and auto-corrects TAG status after Edit/Write/MultiEdit execution.
Real-time monitoring detects missing TAGs and suggests fixes.

Features:
- Validate TAG status after file modification
- Auto-suggest missing TAG creation
- Check TAG chain integrity
- Auto-fix capability (configuration-based)

Usage:
    python3 post_tool__tag_auto_corrector.py <tool_name> <tool_args_json> <result_json>
"""

import json
import sys
import time
from pathlib import Path
from typing import Any, Dict, List, Optional

# Add module path
sys.path.insert(0, str(Path(__file__).parent.parent.parent.parent / "src"))

from moai_adk.core.tags.auto_corrector import AutoCorrection, AutoCorrectionConfig, TagAutoCorrector
from moai_adk.core.tags.policy_validator import PolicyValidationConfig, PolicyViolation, TagPolicyValidator
from moai_adk.core.tags.rollback_manager import RollbackConfig, RollbackManager

from ..utils.hook_config import get_graceful_degradation, load_hook_timeout


def load_config() -> Dict[str, Any]:
    """Load configuration file

    Returns:
        Configuration dictionary
    """
    try:
        config_file = Path(".moai/config.json")
        if config_file.exists():
            with open(config_file, 'r', encoding='utf-8') as f:
                return json.load(f)
    except Exception:
        pass

    return {}


def create_policy_validator() -> TagPolicyValidator:
    """Create TAG policy validator

    Returns:
        TagPolicyValidator instance
    """
    config_data = load_config()
    tag_policy_config = config_data.get("tags", {}).get("policy", {})

    policy_config = PolicyValidationConfig(
        strict_mode=tag_policy_config.get("enforcement_mode", "strict") == "strict",
        require_spec_before_code=tag_policy_config.get("require_spec_before_code", True),
        require_test_for_code=tag_policy_config.get("require_test_for_code", True),
        allow_duplicate_tags=not tag_policy_config.get("enforce_chains", True),
        validation_timeout=tag_policy_config.get("realtime_validation", {}).get("validation_timeout", 5),
        auto_fix_enabled=tag_policy_config.get("auto_correction", {}).get("enabled", False)
    )

    return TagPolicyValidator(config=policy_config)


def create_auto_corrector() -> TagAutoCorrector:
    """Create TAG auto-corrector

    Returns:
        TagAutoCorrector instance
    """
    config_data = load_config()
    auto_correction_config = config_data.get("tags", {}).get("policy", {}).get("auto_correction", {})

    correction_config = AutoCorrectionConfig(
        enable_auto_fix=auto_correction_config.get("enabled", False),
        confidence_threshold=auto_correction_config.get("confidence_threshold", 0.8),
        create_missing_specs=auto_correction_config.get("create_missing_specs", False),
        create_missing_tests=auto_correction_config.get("create_missing_tests", False),
        remove_duplicates=auto_correction_config.get("remove_duplicates", True),
        backup_before_fix=auto_correction_config.get("backup_before_fix", True)
    )

    return TagAutoCorrector(config=correction_config)


def create_rollback_manager() -> RollbackManager:
    """Create rollback manager

    Returns:
        RollbackManager instance
    """
    config_data = load_config()
    rollback_config = config_data.get("tags", {}).get("policy", {}).get("rollback", {})

    config = RollbackConfig(
        checkpoints_dir=rollback_config.get("checkpoints_dir", ".moai/checkpoints"),
        max_checkpoints=rollback_config.get("max_checkpoints", 10),
        auto_cleanup=rollback_config.get("auto_cleanup", True),
        backup_before_rollback=rollback_config.get("backup_before_rollback", True),
        rollback_timeout=rollback_config.get("rollback_timeout", 30)
    )

    return RollbackManager(config=config)


def should_monitor_tool(tool_name: str, tool_args: Dict[str, Any], result: Dict[str, Any]) -> bool:
    """Check if tool should be monitored

    Args:
        tool_name: Tool name
        tool_args: Tool arguments
        result: Tool execution result

    Returns:
        True if should be monitored
    """
    # Monitor only file manipulation tools
    monitoring_tools = {"Edit", "Write", "MultiEdit"}

    # Monitor only on success
    if tool_name not in monitoring_tools:
        return False

    if result.get("success", True):  # Default is True
        return True

    return False


def extract_modified_files(tool_name: str, tool_args: Dict[str, Any]) -> List[str]:
    """Extract modified file paths

    Args:
        tool_name: Tool name
        tool_args: Tool arguments

    Returns:
        List of file paths
    """
    file_paths = []

    if tool_name in {"Edit", "Write"}:
        file_path = tool_args.get("file_path", "")
        if file_path:
            file_paths.append(file_path)

    elif tool_name == "MultiEdit":
        edits = tool_args.get("edits", [])
        for edit in edits:
            file_path = edit.get("file_path", "")
            if file_path:
                file_paths.append(file_path)

    return file_paths


def get_current_file_content(file_path: str) -> str:
    """Get current file content

    Args:
        file_path: File path

    Returns:
        File content
    """
    try:
        path = Path(file_path)
        if path.exists():
            return path.read_text(encoding="utf-8", errors="ignore")
    except Exception:
        pass

    return ""


def create_checkpoint_if_needed(rollback_manager: RollbackManager, file_paths: List[str]) -> Optional[str]:
    """Create checkpoint if needed

    Args:
        rollback_manager: Rollback manager
        file_paths: List of file paths

    Returns:
        Checkpoint ID or None
    """
    try:
        # Create checkpoint only if important files changed
        important_files = [fp for fp in file_paths if any(
            pattern in fp for pattern in ["src/", "tests/", ".moai/", ".claude/"]
        )]

        if important_files:
            description = f"TAG system checkpoint: {len(important_files)} files modified"
            return rollback_manager.create_checkpoint(
                description=description,
                files=important_files,
                metadata={"tool": "post_tool_tag_corrector"}
            )
    except Exception:
        pass

    return None


def create_corrections_summary(corrections: List[AutoCorrection]) -> Dict[str, Any]:
    """Create corrections summary

    Args:
        corrections: List of corrections

    Returns:
        Corrections summary dictionary
    """
    if not corrections:
        return {
            "total_corrections": 0,
            "applied_corrections": 0,
            "corrections": []
        }

    applied_corrections = [c for c in corrections if c.confidence >= 0.8]

    summary = {
        "total_corrections": len(corrections),
        "applied_corrections": len(applied_corrections),
        "high_confidence_corrections": len([c for c in corrections if c.confidence >= 0.9]),
        "corrections": []
    }

    for correction in corrections:
        summary["corrections"].append({
            "file_path": correction.file_path,
            "description": correction.description,
            "confidence": correction.confidence,
            "requires_review": correction.requires_review,
            "applied": correction.confidence >= 0.8
        })

    return summary


def create_monitoring_response(
    violations: List[PolicyViolation],
    corrections: List[AutoCorrection],
    checkpoint_id: Optional[str] = None
) -> Dict[str, Any]:
    """Create monitoring result response

    Args:
        violations: List of policy violations
        corrections: List of corrections
        checkpoint_id: Checkpoint ID

    Returns:
        Monitoring response dictionary
    """
    response = {
        "monitoring_completed": True,
        "timestamp": time.time(),
        "violations_found": len(violations),
        "corrections_available": len(corrections),
        "checkpoint_created": checkpoint_id is not None,
        "checkpoint_id": checkpoint_id
    }

    # Add violation information
    if violations:
        response["violations"] = [v.to_dict() for v in violations]
        response["violation_summary"] = {
            "critical": len([v for v in violations if v.level.value == "critical"]),
            "high": len([v for v in violations if v.level.value == "high"]),
            "medium": len([v for v in violations if v.level.value == "medium"]),
            "low": len([v for v in violations if v.level.value == "low"])
        }

    # Add correction information
    if corrections:
        response["corrections"] = create_corrections_summary(corrections)

    return response


def main() -> None:
    """Main function"""
    try:
        # Load timeout from config (milliseconds to seconds)
        timeout_seconds = load_hook_timeout() / 1000
        graceful_degradation = get_graceful_degradation()

        # Parse arguments
        if len(sys.argv) < 4:
            print(json.dumps({
                "monitoring_completed": False,
                "error": "Invalid arguments. Usage: python3 post_tool__tag_auto_corrector.py <tool_name> <tool_args_json> <result_json>"
            }))
            sys.exit(0)

        tool_name = sys.argv[1]
        try:
            tool_args = json.loads(sys.argv[2])
            tool_result = json.loads(sys.argv[3])
        except json.JSONDecodeError as e:
            print(json.dumps({
                "monitoring_completed": False,
                "error": f"Invalid JSON: {str(e)}"
            }))
            sys.exit(0)

        # Record start time
        start_time = time.time()

        # Check if should monitor
        if not should_monitor_tool(tool_name, tool_args, tool_result):
            print(json.dumps({
                "monitoring_completed": True,
                "message": "Tool execution result not a monitoring target"
            }))
            sys.exit(0)

        # Extract modified file paths
        file_paths = extract_modified_files(tool_name, tool_args)
        if not file_paths:
            print(json.dumps({
                "monitoring_completed": True,
                "message": "No modified files"
            }))
            sys.exit(0)

        # Create components
        policy_validator = create_policy_validator()
        auto_corrector = create_auto_corrector()
        rollback_manager = create_rollback_manager()

        # Create checkpoint if important files
        checkpoint_id = create_checkpoint_if_needed(rollback_manager, file_paths)

        # Validate and correct all files
        all_violations = []
        all_corrections = []

        for file_path in file_paths:
            # Timeout check
            if time.time() - start_time > timeout_seconds:
                break

            # Get current file content
            content = get_current_file_content(file_path)
            if not content:
                continue

            # Validate policy violations
            violations = policy_validator.validate_after_modification(file_path, content)
            all_violations.extend(violations)

            # Generate auto corrections
            if violations:
                corrections = auto_corrector.generate_corrections(violations)
                all_corrections.extend(corrections)

        # Apply auto corrections
        applied_corrections = []
        if all_corrections and auto_corrector.config.enable_auto_fix:
            success = auto_corrector.apply_corrections(all_corrections)
            if success:
                applied_corrections = [c for c in all_corrections
                                     if c.confidence >= auto_corrector.config.confidence_threshold]

        # Create response
        response = create_monitoring_response(all_violations, all_corrections, checkpoint_id)

        # Additional information
        if applied_corrections:
            response["auto_corrections_applied"] = len(applied_corrections)
            response["message"] = f"✅ {len(applied_corrections)} auto-corrections applied"
        elif all_corrections:
            response["message"] = f"💡 {len(all_corrections)} correction suggestions generated (auto-apply disabled)"
        elif all_violations:
            response["message"] = f"⚠️ {len(all_violations)} TAG policy violations found"
        else:
            response["message"] = "✅ TAG policy compliance verified"

        print(json.dumps(response, ensure_ascii=False, indent=2))

    except Exception as e:
        # Log exception and continue
        error_response = {
            "monitoring_completed": False,
            "error": f"Hook execution error: {str(e)}",
            "message": "Hook execution error occurred but processing normally"
        }

        if graceful_degradation:
            error_response["graceful_degradation"] = True
            error_response["message"] = "Hook failed but continuing due to graceful degradation"

        print(json.dumps(error_response, ensure_ascii=False))


if __name__ == "__main__":
    main()
