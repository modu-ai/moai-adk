#!/usr/bin/env python3
"""PreToolUse Hook: TDD Enforcement - Check test exists before code write.

Enforces test-first development by checking for test files before allowing
code writes for Python, TypeScript, and JavaScript.

Exit Codes:
- 0: Allow write (test exists or warning mode)
- 2: Ask user confirmation (enforce mode, no test)
"""

import json
import sys
from pathlib import Path

# Setup import path for shared modules
HOOKS_DIR = Path(__file__).parent
LIB_DIR = HOOKS_DIR / "lib"
if str(LIB_DIR) not in sys.path:
    sys.path.insert(0, str(LIB_DIR))

try:
    from language_detector import detect_language, get_test_file_path
except ImportError:
    # Fallback implementations
    def detect_language(file_path: Path) -> str | None:
        ext = file_path.suffix.lower()
        lang_map = {
            ".py": "Python",
            ".ts": "TypeScript",
            ".tsx": "TypeScript",
            ".js": "JavaScript",
            ".jsx": "JavaScript",
        }
        return lang_map.get(ext)

    def get_test_file_path(source_file: Path) -> Path | None:
        stem = source_file.stem
        ext = source_file.suffix

        if ext == ".py":
            return Path("tests") / f"test_{stem}.py"
        elif ext in (".ts", ".tsx", ".js", ".jsx"):
            return source_file.parent / f"{stem}.test{ext}"
        return None


# File patterns to skip (no test needed)
SKIP_PATTERNS = [
    "/test/",
    "_test.",
    "_spec.",
    ".test.",
    ".spec.",
    "config/",
    "configs/",
    "docs/",
    "doc/",
    ".md",
    ".json",
    ".yaml",
    ".yml",
    ".lock",
    ".min.",
    "__tests__/",
    "__mocks__/",
]

# Mode configuration
MODES = ["warn", "enforce", "off"]


def get_tdd_mode() -> str:
    """Get TDD enforcement mode from quality.yaml."""
    config_path = Path(".moai/config/sections/quality.yaml")
    if config_path.exists():
        try:
            import yaml

            with open(config_path) as f:
                config = yaml.safe_load(f)
                mode = config.get("constitution", {}).get("tdd_mode", "warn")
                if mode in MODES:
                    return mode
        except Exception:
            pass
    return "warn"


def should_skip_file(file_path: str) -> bool:
    """Check if file should skip TDD enforcement.

    Args:
        file_path: Path to the file

    Returns:
        True if should skip, False otherwise
    """
    path_lower = file_path.lower()

    for pattern in SKIP_PATTERNS:
        if pattern.lower() in path_lower:
            return True

    return False


def main():
    """Main hook logic."""
    # Read hook input
    try:
        input_data = json.loads(sys.stdin.read())
    except json.JSONDecodeError:
        # No input or invalid JSON - allow write
        print(json.dumps({}))
        sys.exit(0)

    tool_name = input_data.get("tool_name", "")
    tool_input = input_data.get("tool_input", {})

    # Only check Write operations
    if tool_name != "Write":
        print(json.dumps({}))
        sys.exit(0)

    file_path = tool_input.get("file_path", "")
    if not file_path:
        print(json.dumps({}))
        sys.exit(0)

    # Check if should skip
    if should_skip_file(file_path):
        print(json.dumps({}))
        sys.exit(0)

    # Detect language
    language = detect_language(Path(file_path))
    if not language:
        # Unknown language - skip silently
        print(json.dumps({}))
        sys.exit(0)

    # Only support Python, TypeScript, JavaScript
    if language not in ("Python", "TypeScript", "JavaScript"):
        print(json.dumps({}))
        sys.exit(0)

    # Get expected test file path
    test_file = get_test_file_path(Path(file_path))
    if not test_file:
        print(json.dumps({}))
        sys.exit(0)

    # Check if test file exists
    test_path = Path(test_file)
    test_exists = test_path.exists()

    # Get TDD mode
    mode = get_tdd_mode()

    if mode == "off":
        print(json.dumps({}))
        sys.exit(0)

    output = {
        "hookSpecificOutput": {
            "hookEventName": "PreToolUse",
            "permissionDecision": "allow",  # Default: allow write
        }
    }

    if test_exists:
        # Test file exists - allow write
        output["hookSpecificOutput"]["additionalContext"] = f"TDD check passed for {language}: {test_file} exists"
        print(json.dumps(output))
        sys.exit(0)

    # No test file - handle based on mode
    if mode == "warn":
        output["hookSpecificOutput"]["permissionDecision"] = "allow"
        output["hookSpecificOutput"]["permissionDecisionReason"] = (
            f"TDD Reminder ({language}): Test file not found.\n"
            f"Expected: {test_file}\n"
            f"Best practice: Write test first, then implement code.\n"
            f"Continuing with code write..."
        )
        print(json.dumps(output))
        sys.exit(0)

    elif mode == "enforce":
        output["hookSpecificOutput"]["permissionDecision"] = "ask"
        output["hookSpecificOutput"]["permissionDecisionReason"] = (
            f"TDD Enforcement ({language}): Test file required.\n\n"
            f"Please create test file first:\n"
            f"1. Create {test_file}\n"
            f"2. Write test cases\n"
            f"3. Then implement code\n\n"
            f"Or configure tdd_mode: warn in quality.yaml to disable enforcement."
        )
        print(json.dumps(output))
        sys.exit(2)  # Ask user for confirmation

    print(json.dumps({}))
    sys.exit(0)


if __name__ == "__main__":
    main()
