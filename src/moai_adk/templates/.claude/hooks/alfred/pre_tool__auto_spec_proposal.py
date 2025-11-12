#!/usr/bin/env python3
"""PreToolUse Hook: Automatic SPEC Proposal on Code File Changes

Claude Code Event: PreToolUse
Purpose: Automatically detect code file creation/modification and propose SPEC generation
Execution: Triggered before Write/Edit/MultiEdit tools on Python/JavaScript/Go files

Features:
- Auto-detects code file operations (Create/Modify)
- Analyzes code structure and infers domain
- Proposes SPEC generation with confidence scoring
- Shows editing guidance for better SPEC quality
- Non-intrusive: Displays info without blocking operations

Usage:
    Hook is automatically triggered by Claude Code on applicable tool operations
    No manual invocation required.
"""

import json
import sys
from pathlib import Path
from typing import Any, Dict, Optional

# Setup import path for MoAI-ADK modules
sys.path.insert(0, str(Path(__file__).parent.parent.parent.parent / "src"))



def get_hook_config() -> Dict[str, Any]:
    """Load hook configuration from .moai/config/config.json.

    Returns:
        Hook configuration dictionary with defaults.
    """
    try:
        config_file = Path(".moai/config/config.json")
        if config_file.exists():
            with open(config_file, "r", encoding="utf-8") as f:
                config = json.load(f)
                # Return hook-specific settings
                hooks_config = config.get("hooks", {})
                return hooks_config.get("auto_spec_proposal", {})
    except Exception:
        pass

    # Default configuration
    return {
        "enabled": True,
        "min_confidence": 0.3,
        "show_suggestions": True,
    }


def should_analyze_file(file_path: str) -> bool:
    """Determine if file should be analyzed for SPEC generation.

    Args:
        file_path: Path to the file being created/modified.

    Returns:
        True if file should be analyzed, False otherwise.
    """
    # Only analyze code files
    supported_extensions = {".py", ".js", ".jsx", ".ts", ".tsx", ".go"}

    # Skip test files, configuration files, and special files
    skip_patterns = {
        "test_", "_test.py", "/test", "/tests",
        "spec.md", "SPEC-",
        ".env", ".gitignore",
        "package.json", "pyproject.toml"
    }

    path = Path(file_path)

    # Check file extension
    if path.suffix.lower() not in supported_extensions:
        return False

    # Skip test files and test directories
    file_str = file_path.lower()
    if (path.name.startswith("test_") or
        "_test" in path.name or
        "/tests/" in file_str or
        "/test/" in file_str or
        file_str.startswith("tests/") or
        file_str.startswith("test/")):
        return False

    # Skip SPEC and config files
    if any(pattern in file_path for pattern in skip_patterns):
        return False

    return True


def extract_tool_context(tool_name: str, tool_args: Dict[str, Any]) -> Optional[Dict[str, Any]]:
    """Extract file path and operation context from tool arguments.

    Args:
        tool_name: Tool being executed (Write, Edit, MultiEdit).
        tool_args: Tool arguments dictionary.

    Returns:
        Dictionary with file_path and operation type, or None if not applicable.
    """
    try:
        if tool_name == "Write":
            file_path = tool_args.get("file_path")
            return {
                "file_path": file_path,
                "operation": "create",
                "tool": "Write"
            }

        elif tool_name == "Edit":
            file_path = tool_args.get("file_path")
            return {
                "file_path": file_path,
                "operation": "modify",
                "tool": "Edit"
            }

        elif tool_name == "MultiEdit":
            # MultiEdit modifies multiple files
            # For simplicity, we skip this for now (can enhance later)
            return None

    except Exception:
        pass

    return None


def generate_spec_proposal(file_path: str) -> Optional[Dict[str, Any]]:
    """Generate SPEC proposal for the code file.

    Args:
        file_path: Path to the code file.

    Returns:
        Dictionary with SPEC proposal info, or None if generation failed.
    """
    try:
        # Check if file exists (for Edit operation)
        path = Path(file_path)
        if not path.exists():
            # For Write operation, file doesn't exist yet - return None
            return None

        # Generate SPEC template
        generator = SpecGenerator()
        result = generator.generate_spec_template(path)

        if result.get("success"):
            return {
                "domain": result["domain"],
                "confidence": result["confidence"],
                "spec_path": result["spec_path"],
                "editing_guide": result["editing_guide"],
            }

    except Exception:
        pass

    return None


def format_spec_proposal_message(proposal: Dict[str, Any], operation: str) -> str:
    """Format SPEC proposal as interactive user-friendly message.

    Args:
        proposal: SPEC proposal dictionary.
        operation: File operation type (create/modify).

    Returns:
        Formatted message for display prompting user action.
    """
    domain = proposal["domain"]
    confidence = proposal["confidence"]
    confidence_pct = f"{confidence * 100:.0f}%"
    spec_path = proposal["spec_path"]

    # Determine confidence level description
    if confidence >= 0.8:
        confidence_desc = "Very High"
        confidence_emoji = "🟢"
    elif confidence >= 0.6:
        confidence_desc = "High"
        confidence_emoji = "🟡"
    elif confidence >= 0.4:
        confidence_desc = "Medium"
        confidence_emoji = "🟠"
    else:
        confidence_desc = "Low"
        confidence_emoji = "🔴"

    # Build message
    message_lines = [
        "",
        "💡 Auto SPEC Proposal",
        f"📊 Domain: {domain} | Confidence: {confidence_emoji} {confidence_desc} ({confidence_pct})",
        f"📁 Suggested SPEC path: {spec_path}",
    ]

    # Add editing guide suggestions
    if proposal.get("editing_guide"):
        message_lines.append("")
        message_lines.append("📝 Recommended SPEC Checklist:")
        for i, suggestion in enumerate(proposal["editing_guide"][:4], 1):  # Show first 4
            message_lines.append(f"   {i}. {suggestion}")

    # Add action prompt
    message_lines.append("")
    if confidence >= 0.6:
        message_lines.append("✅ Would you like me to generate the SPEC file now?")
        message_lines.append("   Just confirm: 'Yes, create SPEC' or 'Generate SPEC file'")
    elif confidence >= 0.4:
        message_lines.append("❓ Would you like me to generate a SPEC template for this domain?")
        message_lines.append("   Note: You may need to adjust the domain inference afterward.")
        message_lines.append("   Confirm: 'Yes, create SPEC' or 'Skip'")
    else:
        message_lines.append("⚠️ SPEC proposal confidence is low. Consider manually creating SPEC.")
        message_lines.append("   Or confirm: 'Yes, create SPEC' to proceed with auto-generation")

    return "\n".join(message_lines)


def main() -> None:
    """Main entry point for auto SPEC proposal hook.

    Reads PreToolUse event data, determines if code file is being created/modified,
    and proposes automatic SPEC generation with confidence scoring.

    Exit Codes:
        0: Success
        1: Error (invalid input, timeout, etc.)
    """
    try:
        # Read JSON payload from stdin
        input_data = sys.stdin.read()
        data = json.loads(input_data) if input_data.strip() else {}

        # Load hook configuration
        hook_config = get_hook_config()
        if not hook_config.get("enabled", True):
            # Hook is disabled, continue without processing
            result: Dict[str, Any] = {"continue": True}
            print(json.dumps(result))
            return

        # Extract tool context
        tool_name = data.get("toolName", "")
        tool_args = data.get("toolArguments", {})

        context = extract_tool_context(tool_name, tool_args)
        if not context:
            # Not a relevant tool operation
            result = {"continue": True}
            print(json.dumps(result))
            return

        file_path = context["file_path"]
        operation = context["operation"]

        # Check if file should be analyzed
        if not should_analyze_file(file_path):
            # Not a code file we should analyze
            result = {"continue": True}
            print(json.dumps(result))
            return

        # Generate SPEC proposal
        proposal = generate_spec_proposal(file_path)
        if not proposal:
            # Could not generate proposal
            result = {"continue": True}
            print(json.dumps(result))
            return

        # Check confidence threshold
        min_confidence = hook_config.get("min_confidence", 0.3)
        if proposal["confidence"] < min_confidence:
            # Confidence below threshold, don't show proposal
            result = {"continue": True}
            print(json.dumps(result))
            return

        # Format and return proposal message
        message = format_spec_proposal_message(proposal, operation)

        result = {
            "continue": True,
            "systemMessage": message,
            "hookSpecificOutput": {
                "spec_proposal": proposal,
                "operation": operation,
            }
        }

        print(json.dumps(result))

    except json.JSONDecodeError as e:
        # JSON parse error
        error_result: Dict[str, Any] = {
            "continue": True,
            "hookSpecificOutput": {"error": f"JSON parse error: {e}"},
        }
        print(json.dumps(error_result))
        print(f"Auto SPEC Proposal JSON error: {e}", file=sys.stderr)
        sys.exit(1)

    except Exception as e:
        # Unexpected error - continue without proposal
        error_result: Dict[str, Any] = {
            "continue": True,
            "hookSpecificOutput": {"error": f"Hook error: {e}"},
        }
        print(json.dumps(error_result))
        print(f"Auto SPEC Proposal error: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
