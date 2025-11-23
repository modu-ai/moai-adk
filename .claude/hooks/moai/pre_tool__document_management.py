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

try:
    from lib.config_manager import ConfigManager  # noqa: E402
    from lib.timeout import CrossPlatformTimeout  # noqa: E402
    from lib.timeout import TimeoutError as PlatformTimeoutError  # noqa: E402
except ImportError:
    # Fallback for timeout if shared module unavailable
    import signal

    class PlatformTimeoutError(Exception):  # type: ignore[no-redef]
        pass

    class CrossPlatformTimeout:  # type: ignore[no-redef]
        def __init__(self, seconds: int) -> None:
            self.seconds = seconds

        def start(self) -> None:
            signal.alarm(int(self.seconds))

        def cancel(self) -> None:
            signal.alarm(0)

    ConfigManager = None  # type: ignore


def get_file_pattern_category(filename: str, config: Dict) -> Optional[Tuple[str, str]]:
    """Match filename against patterns to determine category

    Args:
        filename: Name of the file to categorize
        config: Configuration dictionary

    Returns:
        Tuple of (directory_type, category) or None if no match
        Example: ("reports", "inspection") or ("scripts", "conversion")
    """
    patterns = config.get("document_management", {}).get("file_patterns", {})

    for dir_type, categories in patterns.items():
        for category, pattern_list in categories.items():
            for pattern in pattern_list:
                # Convert glob pattern to regex
                regex = pattern.replace("*", ".*").replace("?", ".")
                if re.match(f"^{regex}$", filename):
                    return (dir_type, category)

    return None


def suggest_moai_location(filename: str, config: Dict) -> str:
    """Suggest appropriate .moai/ location based on file pattern

    Args:
        filename: Name of the file
        config: Configuration dictionary

    Returns:
        Suggested .moai/ path
    """
    # Try pattern matching first
    match = get_file_pattern_category(filename, config)

    if match:
        dir_type, category = match
        base_dir = config.get("document_management", {}).get("directories", {}).get(dir_type, {}).get("base", "")
        if base_dir:
            return f"{base_dir}{category}/"

    # Default fallback suggestions
    if filename.endswith(".md"):
        return ".moai/temp/work/"
    elif filename.endswith((".sh", ".py", ".js")):
        return ".moai/scripts/dev/"
    elif filename.endswith((".tmp", ".temp", ".bak")):
        return ".moai/temp/work/"

    # Ultimate fallback
    return ".moai/temp/work/"


def is_root_whitelisted(filename: str, config: Dict) -> bool:
    """Check if file is allowed in project root

    Args:
        filename: Name of the file
        config: Configuration dictionary

    Returns:
        True if file is whitelisted for root directory
    """
    whitelist = config.get("document_management", {}).get("root_whitelist", [])

    for pattern in whitelist:
        # Convert glob pattern to regex
        regex = pattern.replace("*", ".*").replace("?", ".")
        if re.match(f"^{regex}$", filename):
            return True

    return False


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
        project_root = Path.cwd()

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
        "should_block": False
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
        f"   Tip: Use Skill(\"moai-core-document-management\") for guidance"
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
    response = {
        "continue": not validation["should_block"],
        "systemMessage": validation["warning"]
    }

    return response


def main() -> None:
    """Main entry point for PreToolUse hook

    Validates file locations before Write/Edit/MultiEdit operations:
    1. Load document management configuration
    2. Extract file path from tool parameters
    3. Validate against root whitelist
    4. Suggest correct .moai/ location if violation detected
    5. Warn or block operation based on config

    Exit Codes:
        0: Success (validation complete)
        1: Error (timeout, JSON parse failure, handler exception)
    """
    # Set 2-second timeout (optimized for performance)
    timeout = CrossPlatformTimeout(2)
    timeout.start()

    try:
        # Read JSON payload from stdin
        # Handle Docker/non-interactive environments by checking TTY
        input_data = sys.stdin.read() if not sys.stdin.isatty() else "{}"
        data = json.loads(input_data) if input_data.strip() else {}

        # Call handler
        result = handle_pre_tool_use(data)

        # Output result as JSON
        print(json.dumps(result, ensure_ascii=False))
        sys.exit(0)

    except PlatformTimeoutError:
        # Timeout - allow operation to continue
        timeout_response: Dict[str, Any] = {
            "continue": True,
            "systemMessage": "⚠️ Document validation timeout - operation proceeding",
        }
        print(json.dumps(timeout_response, ensure_ascii=False))
        print("PreToolUse document management hook timeout after 2 seconds", file=sys.stderr)
        sys.exit(1)

    except json.JSONDecodeError as e:
        # JSON parse error - allow operation to continue
        json_error_response: Dict[str, Any] = {
            "continue": True,
            "hookSpecificOutput": {"error": f"JSON parse error: {e}"},
        }
        print(json.dumps(json_error_response, ensure_ascii=False))
        print(f"PreToolUse document management JSON parse error: {e}", file=sys.stderr)
        sys.exit(1)

    except Exception as e:
        # Unexpected error - allow operation to continue
        unexpected_error_response: Dict[str, Any] = {
            "continue": True,
            "hookSpecificOutput": {"error": f"Document management error: {e}"},
        }
        print(json.dumps(unexpected_error_response, ensure_ascii=False))
        print(f"PreToolUse document management unexpected error: {e}", file=sys.stderr)
        sys.exit(1)

    finally:
        # Always cancel alarm
        timeout.cancel()


if __name__ == "__main__":
    main()
