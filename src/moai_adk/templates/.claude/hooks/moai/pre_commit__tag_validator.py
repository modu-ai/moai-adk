#!/usr/bin/env python3
"""Pre-commit hook for TAG validation (T3: Pre-commit Validation Hook).

Validates TAG format and SPEC existence before Git commit.
Runs in warn mode by default (allows commit with warnings).
"""

from __future__ import annotations

import logging
import sys
from pathlib import Path

try:
    from .lib import (
        extract_tags_from_file,
        get_config_manager,
        spec_document_exists,
        validate_tag,
    )
    from .lib.tag_validator import TAG
except ImportError:
    # Fallback for direct execution from hooks directory
    try:
        from lib import (
            extract_tags_from_file,
            get_config_manager,
            spec_document_exists,
            validate_tag,
        )
        from lib.tag_validator import TAG
    except ImportError:
        # Final fallback - try relative import from parent
        import sys
        from pathlib import Path

        sys.path.insert(0, str(Path(__file__).parent.parent))
        from hooks.moai.lib import (
            extract_tags_from_file,
            get_config_manager,
            spec_document_exists,
            validate_tag,
        )
        from hooks.moai.lib.tag_validator import TAG

logger = logging.getLogger(__name__)


def get_file_context(file_path: Path) -> dict:
    """Get file context for flexible TAG validation.

    Args:
        file_path: Path to analyze

    Returns:
        Context dict with file type information
    """
    import re

    file_str = str(file_path)

    return {
        "is_test_file": bool(re.search(r"test_\w+\.py|_\w+_test\.py|tests?/", file_str)),
        "is_impl_file": not bool(re.search(r"test_\w+\.py|_\w+_test\.py|tests?/", file_str)),
    }


def load_tag_validation_config() -> dict:
    """Load TAG validation configuration from quality.yaml.

    Returns:
        Configuration dict with enabled, mode, and other settings
    """
    try:
        config_manager = get_config_manager()
        quality_config = config_manager.config

        # Get TAG validation settings
        tag_config = quality_config.get("tag_validation", {})

        # Apply defaults if not configured (T5.2)
        defaults = {
            "enabled": True,
            "mode": "warn",  # warn | enforce | off
        }

        # Merge defaults with config
        for key, value in defaults.items():
            if key not in tag_config:
                tag_config[key] = value

        return tag_config

    except Exception as e:
        logger.warning(f"Error loading TAG validation config: {e}")
        # Return safe defaults
        return {"enabled": True, "mode": "warn"}


def validate_tag_format(tag: TAG) -> tuple[bool, list[str]]:
    """Validate TAG format.

    Args:
        tag: TAG to validate

    Returns:
        Tuple of (is_valid, error_messages)
    """
    is_valid, errors = validate_tag(tag)
    return is_valid, errors


def check_spec_exists(spec_id: str) -> bool:
    """Check if SPEC document exists.

    Args:
        spec_id: SPEC identifier to check

    Returns:
        True if SPEC exists, False otherwise
    """
    return spec_document_exists(spec_id)


def validate_file_tags(file_path: Path, config: dict) -> tuple[int, list[str]]:
    """Validate all TAGs in a file.

    Args:
        file_path: Path to Python file
        config: TAG validation configuration

    Returns:
        Tuple of (exit_code, error_messages)
    """
    errors = []
    file_context = get_file_context(file_path)

    # Extract TAGs from file
    tags = extract_tags_from_file(file_path)

    if not tags:
        return 0, []  # No TAGs, no errors

    # Validate each TAG
    for tag in tags:
        # Validate TAG format
        is_valid, format_errors = validate_tag_format(tag)

        if not is_valid:
            for error in format_errors:
                errors.append(f"{file_path}:{tag.line}: {error}")

        # Check SPEC existence with context-aware validation
        if not check_spec_exists(tag.spec_id):
            mode = config.get("mode", "warn")

            # `related` verb: always allow, just show hint
            if tag.verb == "related":
                errors.append(
                    f"{file_path}:{tag.line}: Hint: Consider creating SPEC for {tag.spec_id} "
                    "(related TAG - no SPEC required)"
                )
                continue  # Allow commit

            # TEST file with `verify` verb: allow with warning
            if file_context["is_test_file"] and tag.verb == "verify":
                errors.append(
                    f"{file_path}:{tag.line}: Warning: TEST references missing SPEC: {tag.spec_id} "
                    "(verify TAG in test file - commit allowed)"
                )
                continue  # Allow commit

            # Implementation file with `impl` verb: strict validation
            if file_context["is_impl_file"] and tag.verb == "impl":
                if mode == "enforce":
                    errors.append(
                        f"{file_path}:{tag.line}: Error: Implementation references missing SPEC: {tag.spec_id} "
                        "(commit blocked in enforce mode)"
                    )
                else:
                    errors.append(
                        f"{file_path}:{tag.line}: Warning: Implementation references missing SPEC: {tag.spec_id} "
                        "(commit allowed in warn mode)"
                    )
                continue

            # Default: warn but allow
            errors.append(
                f"{file_path}:{tag.line}: Warning: SPEC document not found: {tag.spec_id} (commit allowed in warn mode)"
            )

    # Determine exit code
    if errors:
        mode = config.get("mode", "warn")
        if mode == "enforce":
            # Check if any errors are blocking (not hints/warnings)
            blocking_errors = [e for e in errors if "Error:" in e]
            if blocking_errors:
                return 1, errors  # Block commit
        return 1, errors  # Warning code but commit allowed
    else:
        return 0, []  # Success


def main() -> int:
    """Main pre-commit hook entry point.

    Returns:
        Exit code (0 = success, 1 = warning/error)
    """
    # Load configuration
    config = load_tag_validation_config()

    # Check if TAG validation is enabled
    if not config.get("enabled", True):
        print("TAG validation disabled")
        return 0

    # Check if validation mode is 'off'
    mode = config.get("mode", "warn")
    if mode == "off":
        print("TAG validation mode: off")
        return 0

    # Get staged files from git
    import subprocess

    try:
        # Get list of staged Python files
        result = subprocess.run(
            ["git", "diff", "--cached", "--name-only", "--diff-filter=ACM"],
            capture_output=True,
            text=True,
            check=False,
        )

        if result.returncode != 0:
            # Not in git repository or no staged files
            return 0

        staged_files = result.stdout.strip().splitlines()

        # Filter for Python files only
        python_files = [Path(f) for f in staged_files if f.endswith(".py") and Path(f).exists()]

        if not python_files:
            # No Python files staged
            return 0

    except Exception as e:
        logger.warning(f"Error getting staged files: {e}")
        return 0  # Don't block commit on git error

    # Validate TAGs in all staged Python files
    all_errors = []

    for file_path in python_files:
        _, file_errors = validate_file_tags(file_path, config)

        if file_errors:
            all_errors.extend(file_errors)

    # Display errors
    if all_errors:
        print("\nTAG Validation Results:")
        print("=" * 60)

        for error in all_errors:
            print(f"  {error}")

        print("=" * 60)

        if mode == "enforce":
            print("\nCommit blocked due to TAG validation errors (enforce mode)")
            print("Fix the errors and try again, or use --no-verify to bypass")
        else:
            print("\nCommit allowed with warnings (warn mode)")
            print("Consider fixing the TAG validation errors")

        return 1  # Warning code
    else:
        print("TAG validation: All TAGs valid")
        return 0


if __name__ == "__main__":
    sys.exit(main())
