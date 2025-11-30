#!/usr/bin/env python3

"""PreToolUse Hook: Document Management - File Location Validation

Claude Code Event: PreToolUse
Purpose: Validate file locations before Write/Edit operations to prevent root pollution
Execution: Triggered before Write, Edit, or MultiEdit tools are used
Matcher: Write|Edit|MultiEdit

Output: System message with validation results and suggestions

Validation Rules:
- Check if file path is in project root
- Validate against root_whitelist from config.json
- If not whitelisted: warn or block creation, suggest correct .moai/ path
- Use pattern matching for auto-categorization
"""

import json
import re
import sys
from pathlib import Path
from typing import Any, Dict, Optional, Tuple

# Setup import path for shared modules
HOOKS_DIR = Path(__file__).parent
LIB_DIR = HOOKS_DIR / "lib"
if str(LIB_DIR) not in sys.path:
    sys.path.insert(0, str(LIB_DIR))

from lib.path_utils import find_project_root

try:
    from lib.unified_timeout_manager import (
        get_timeout_manager, hook_timeout_context, HookTimeoutConfig,
        TimeoutPolicy, HookTimeoutError
    )
    from lib.config_validator import get_config_validator, ValidationIssue
    from lib.common import (  # noqa: E402
        get_file_pattern_category,
        is_root_whitelisted,
        suggest_moai_location,
    )
    from lib.config_manager import ConfigManager  # noqa: E402
    from lib.timeout import TimeoutError as PlatformTimeoutError  # noqa: E402
except ImportError:
    # Fallback for timeout if shared module unavailable
    import signal

    def get_timeout_manager():
        return None

    def hook_timeout_context(hook_name, config=None):
        import contextlib
        @contextlib.contextmanager
        def dummy_context():
            yield
        return dummy_context()

    class HookTimeoutConfig:
        def __init__(self, **kwargs):
            pass

    class TimeoutPolicy:
        FAST = "fast"
        NORMAL = "normal"
        SLOW = "slow"

    class HookTimeoutError(Exception):
        pass

    def get_config_validator():
        return None

    class ValidationIssue:
        pass

    class PlatformTimeoutError(Exception):  # type: ignore[no-redef]
        pass

    ConfigManager = None  # type: ignore

    # Fallback implementations if lib.common unavailable
    def is_root_whitelisted(filename: str, config: Dict) -> bool:
        """Fallback: Check if file is allowed in project root"""
        whitelist = config.get("document_management", {}).get("root_whitelist", [])
        for pattern in whitelist:
            regex = pattern.replace("*", ".*").replace("?", ".")
            if re.match(f"^{regex}$", filename):
                return True
        return False

    def get_file_pattern_category(filename: str, config: Dict) -> Optional[Tuple[str, str]]:
        """Fallback: Match filename against patterns"""
        patterns = config.get("document_management", {}).get("file_patterns", {})
        for dir_type, categories in patterns.items():
            for category, pattern_list in categories.items():
                for pattern in pattern_list:
                    regex = pattern.replace("*", ".*").replace("?", ".")
                    if re.match(f"^{regex}$", filename):
                        return (dir_type, category)
        return None

    def suggest_moai_location(filename: str, config: Dict) -> str:
        """Fallback: Suggest .moai/ location"""
        match = get_file_pattern_category(filename, config)
        if match:
            dir_type, category = match
            base_dir = config.get("document_management", {}).get("directories", {}).get(dir_type, {}).get("base", "")
            if base_dir:
                return f"{base_dir}{category}/"
        if filename.endswith(".md"):
            return ".moai/temp/work/"
        elif filename.endswith((".sh", ".py", ".js")):
            return ".moai/scripts/dev/"
        elif filename.endswith((".tmp", ".temp", ".bak")):
            return ".moai/temp/work/"
        return ".moai/temp/work/"


def validate_file_location(file_path: str, config: Dict) -> Dict[str, Any]:
    """Validate file location according to document management rules

    Args:
        file_path: Path to file being created/modified
        config: Configuration dictionary

    Returns:
        Validation result dictionary
    """
    path_obj = Path(file_path)
    filename = path_obj.name

    # Get project root (assuming .moai/config exists)
    try:
        project_root = Path(".moai/config/config.json").parent.parent.resolve()
    except Exception:
        project_root = find_project_root()

    # Get absolute path
    try:
        abs_path = path_obj.resolve()
    except Exception:
        abs_path = path_obj

    # Check if file is in project root
    try:
        is_in_root = abs_path.parent == project_root
    except Exception:
        # Fallback: check if path has only one component (filename only)
        is_in_root = str(path_obj.parent) in [".", ""]

    result: Dict[str, Any] = {
        "valid": True,
        "is_root": is_in_root,
        "whitelisted": False,
        "suggested_location": None,
        "warning": None,
        "should_block": False,
    }

    # If not in root, validation passes
    if not is_in_root:
        result["valid"] = True
        return result

    # File is in root - check whitelist
    if is_root_whitelisted(filename, config):
        result["valid"] = True
        result["whitelisted"] = True
        return result

    # File is in root and NOT whitelisted - violation
    doc_mgmt = config.get("document_management", {})
    block_violations = doc_mgmt.get("validation", {}).get("block_violations", False)

    suggested = suggest_moai_location(filename, config)

    result["valid"] = False
    result["suggested_location"] = suggested
    result["warning"] = (
        f"⚠️ Root directory pollution detected\n"
        f"   File: {filename}\n"
        f"   Reason: Not in root_whitelist\n"
        f"   ✅ Suggested: {suggested}{filename}\n"
        f"\n"
        f'   Tip: Use Skill("moai-core-document-management") for guidance'
    )

    if block_violations:
        result["should_block"] = True
        result["warning"] = (
            f"❌ Root directory pollution BLOCKED\n"
            f"   File: {filename}\n"
            f"   Reason: Not in root_whitelist\n"
            f"   ✅ Required: {suggested}{filename}\n"
            f"\n"
            f"   Config: document_management.block_root_pollution = true\n"
            f"   To disable: Set block_root_pollution to false in .moai/config/config.json"
        )

    return result


def handle_pre_tool_use(payload: Dict) -> Dict[str, Any]:
    """Handle PreToolUse event for document management

    Args:
        payload: Hook payload containing tool name and parameters

    Returns:
        Hook response dictionary
    """
    # Load configuration
    if ConfigManager:
        config = ConfigManager().load_config()
    else:
        config = {}

    # Check if document management is enabled
    doc_mgmt = config.get("document_management", {})
    if not doc_mgmt.get("enabled", True):
        return {"continue": True}

    # Get tool name and parameters
    tool_name = payload.get("tool", {}).get("name", "")
    parameters = payload.get("tool", {}).get("parameters", {})

    # Only validate Write, Edit, MultiEdit operations
    if tool_name not in ["Write", "Edit", "MultiEdit"]:
        return {"continue": True}

    # Extract file path
    file_path = None
    if tool_name == "Write":
        file_path = parameters.get("file_path")
    elif tool_name == "Edit":
        file_path = parameters.get("file_path")
    elif tool_name == "MultiEdit":
        # MultiEdit has edits array
        edits = parameters.get("edits", [])
        if edits and len(edits) > 0:
            file_path = edits[0].get("file_path")

    # If no file path, allow operation
    if not file_path:
        return {"continue": True}

    # Validate file location
    validation = validate_file_location(file_path, config)

    # If validation passed, allow operation
    if validation["valid"]:
        return {"continue": True}

    # Validation failed
    response = {"continue": not validation["should_block"], "systemMessage": validation["warning"]}

    return response


def execute_pre_tool_validation(data: Dict[str, Any]) -> Dict[str, Any]:
    """Execute pre-tool validation with performance optimization"""
    start_time = time.time() if 'time' in globals() else 0

    # Call handler
    result = handle_pre_tool_use(data)

    # Add performance metrics
    if 'time' in globals():
        result["performance"] = {
            "execution_time_seconds": round(time.time() - start_time, 3),
            "timeout_manager_used": get_timeout_manager() is not None,
            "config_validator_used": get_config_validator() is not None
        }

    return result


def main() -> None:
    """Main entry point for PreToolUse hook

    Validates file locations before Write/Edit/MultiEdit operations:
    1. Load document management configuration
    2. Extract file path from tool parameters
    3. Validate against root whitelist
    4. Suggest correct .moai/ location if violation detected
    5. Warn or block operation based on config

    Features:
    - Optimized timeout handling with unified manager
    - Enhanced error handling with graceful degradation
    - Performance monitoring and validation
    - Configuration validation

    Exit Codes:
        0: Success (validation complete)
        1: Error (timeout, JSON parse failure, handler exception)
    """
    import time

    # Configure timeout for pre_tool hook (fast policy)
    timeout_config = HookTimeoutConfig(
        policy=TimeoutPolicy.FAST,
        custom_timeout_ms=2000,  # 2 seconds for fast validation
        retry_count=1,
        retry_delay_ms=100,
        graceful_degradation=True,
        memory_limit_mb=50  # Low memory limit for validation
    )

    def read_input_data() -> Dict[str, Any]:
        """Read and parse input JSON data"""
        # Read JSON payload from stdin
        # Handle Docker/non-interactive environments by checking TTY
        input_data = sys.stdin.read() if not sys.stdin.isatty() else "{}"
        return json.loads(input_data) if input_data.strip() else {}

    # Use unified timeout manager if available
    timeout_manager = get_timeout_manager()
    if timeout_manager:
        try:
            result = timeout_manager.execute_with_timeout(
                "pre_tool__document_management",
                lambda: execute_pre_tool_validation(read_input_data()),
                config=timeout_config
            )

            # Output result as JSON
            print(json.dumps(result, ensure_ascii=False))
            sys.exit(0)

        except HookTimeoutError as e:
            # Enhanced timeout error handling
            timeout_response = {
                "continue": True,
                "systemMessage": "⚠️ Document validation timeout - operation proceeding",
                "error_details": {
                    "hook_id": e.hook_id,
                    "timeout_seconds": e.timeout_seconds,
                    "execution_time": e.execution_time,
                    "will_retry": e.will_retry
                },
                "performance": {
                    "timeout_manager_used": True,
                    "graceful_degradation": True
                }
            }
            print(json.dumps(timeout_response, ensure_ascii=False))
            print(f"PreToolUse document management hook timeout: {e}", file=sys.stderr)
            sys.exit(1)

        except Exception as e:
            # Enhanced error handling with context
            error_response = {
                "continue": True,
                "systemMessage": "⚠️ Document validation encountered an error - operation proceeding",
                "hookSpecificOutput": {"error": f"Document management error: {e}"},
                "error_details": {
                    "error_type": type(e).__name__,
                    "message": str(e),
                    "graceful_degradation": True
                },
                "performance": {
                    "timeout_manager_used": True,
                    "graceful_degradation": True
                }
            }
            print(json.dumps(error_response, ensure_ascii=False))
            print(f"PreToolUse document management error: {e}", file=sys.stderr)
            sys.exit(1)

    else:
        # Fallback to legacy timeout handling
        try:
            from lib.timeout import CrossPlatformTimeout

            # Set 2-second timeout (optimized for performance)
            timeout = CrossPlatformTimeout(2)
            timeout.start()

            try:
                result = execute_pre_tool_validation(read_input_data())

                # Output result as JSON
                print(json.dumps(result, ensure_ascii=False))
                sys.exit(0)

            finally:
                # Always cancel timeout
                timeout.cancel()

        except ImportError:
            # No timeout handling available
            try:
                result = execute_pre_tool_validation(read_input_data())
                print(json.dumps(result, ensure_ascii=False))
                sys.exit(0)
            except Exception as e:
                print(json.dumps({
                    "continue": True,
                    "systemMessage": "⚠️ Document validation completed with errors - operation proceeding",
                    "hookSpecificOutput": {"error": str(e)}
                }, ensure_ascii=False))
                sys.exit(0)

        except PlatformTimeoutError:
            # Timeout - allow operation to continue
            timeout_response = {
                "continue": True,
                "systemMessage": "⚠️ Document validation timeout - operation proceeding",
                "performance": {
                    "timeout_manager_used": False,
                    "graceful_degradation": True
                }
            }
            print(json.dumps(timeout_response, ensure_ascii=False))
            print("PreToolUse document management hook timeout after 2 seconds", file=sys.stderr)
            sys.exit(1)

        except json.JSONDecodeError as e:
            # JSON parse error - allow operation to continue
            json_error_response = {
                "continue": True,
                "hookSpecificOutput": {"error": f"JSON parse error: {e}"},
                "performance": {
                    "timeout_manager_used": False,
                    "graceful_degradation": True
                }
            }
            print(json.dumps(json_error_response, ensure_ascii=False))
            print(f"PreToolUse document management JSON parse error: {e}", file=sys.stderr)
            sys.exit(1)

        except Exception as e:
            # Unexpected error - allow operation to continue
            unexpected_error_response = {
                "continue": True,
                "hookSpecificOutput": {"error": f"Document management error: {e}"},
                "performance": {
                    "timeout_manager_used": False,
                    "graceful_degradation": True
                }
            }
            print(json.dumps(unexpected_error_response, ensure_ascii=False))
            print(f"PreToolUse document management unexpected error: {e}", file=sys.stderr)
            sys.exit(1)


if __name__ == "__main__":
    main()
