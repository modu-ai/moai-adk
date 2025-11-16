#!/usr/bin/env python3
"""
Verify package integrity before building and distribution.

This script ensures all critical files are included in built packages
(wheel and tarball) before deployment.
"""

import sys
import tarfile
import zipfile
from pathlib import Path

# Project root directory
PROJECT_ROOT = Path(__file__).parent.parent

# Required source files that must exist in src/moai_adk/templates/
REQUIRED_SOURCE_FILES = [
    PROJECT_ROOT / "src/moai_adk/templates/.claude/output-styles/moai/r2d2.md",
    PROJECT_ROOT / "src/moai_adk/templates/.claude/output-styles/moai/yoda.md",
    PROJECT_ROOT / "src/moai_adk/templates/.claude/skills",
    PROJECT_ROOT / "src/moai_adk/templates/.claude/agents",
    PROJECT_ROOT / "src/moai_adk/templates/.moai/config/config.json",
]

# Required patterns in wheel/tarball packages
REQUIRED_WHEEL_PATTERNS = [
    "moai_adk/templates/.claude/output-styles/moai/r2d2.md",
    "moai_adk/templates/.claude/output-styles/moai/yoda.md",
    "moai_adk/templates/.claude/skills",
    "moai_adk/templates/.claude/agents",
    "moai_adk/templates/.moai/config/config.json",
]


def verify_source_files() -> bool:
    """
    Verify that all required source files exist.

    Returns:
        bool: True if all files exist, False otherwise.
    """
    all_exist = True
    for file_path in REQUIRED_SOURCE_FILES:
        if not file_path.exists():
            print(f"✗ Missing: {file_path.relative_to(PROJECT_ROOT)}")
            all_exist = False

    if all_exist:
        print("✓ Source files verified: All critical files present")

    return all_exist


def verify_wheel_contents(wheel_path: str) -> bool:
    """
    Verify that all required patterns exist in wheel file.

    Args:
        wheel_path: Path to the wheel file.

    Returns:
        bool: True if all patterns found, False otherwise.
    """
    wheel_file = Path(wheel_path)

    if not wheel_file.exists():
        print(f"✗ Wheel file not found: {wheel_path}")
        return False

    try:
        with zipfile.ZipFile(wheel_file, "r") as whl:
            wheel_contents = set(whl.namelist())

        # Check each required pattern
        missing_patterns = []
        for pattern in REQUIRED_WHEEL_PATTERNS:
            pattern_found = any(
                pattern in name for name in wheel_contents
            )
            if not pattern_found:
                missing_patterns.append(pattern)

        if missing_patterns:
            for pattern in missing_patterns:
                print(f"✗ Missing pattern in wheel: {pattern}")
            return False

        # Verify critical files are present (stricter check)
        critical_files = [
            "moai_adk/templates/.claude/output-styles/moai/r2d2.md",
            "moai_adk/templates/.claude/output-styles/moai/yoda.md",
            "moai_adk/templates/.moai/config/config.json",
        ]

        for file_path in critical_files:
            if file_path not in wheel_contents:
                print(f"✗ Critical file missing in wheel: {file_path}")
                return False

        print("✓ Wheel contents verified: All patterns found")
        return True

    except zipfile.BadZipFile:
        print(f"✗ Invalid wheel file format: {wheel_path}")
        return False
    except Exception as e:
        print(f"✗ Error reading wheel file: {e}")
        return False


def verify_tarball_contents(tarball_path: str) -> bool:
    """
    Verify that all required patterns exist in tarball file.

    Args:
        tarball_path: Path to the tarball file.

    Returns:
        bool: True if all patterns found, False otherwise.
    """
    tarball_file = Path(tarball_path)

    if not tarball_file.exists():
        print(f"✗ Tarball file not found: {tarball_path}")
        return False

    try:
        with tarfile.open(tarball_file, "r:gz") as tar:
            tarball_contents = set(tar.getnames())

        # Check each required pattern
        missing_patterns = []
        for pattern in REQUIRED_WHEEL_PATTERNS:
            pattern_found = any(
                pattern in name for name in tarball_contents
            )
            if not pattern_found:
                missing_patterns.append(pattern)

        if missing_patterns:
            for pattern in missing_patterns:
                print(f"✗ Missing pattern in tarball: {pattern}")
            return False

        # Verify critical files are present (stricter check)
        found_critical_files = 0
        for file_pattern in ["r2d2.md", "yoda.md", "config.json"]:
            if any(file_pattern in name for name in tarball_contents):
                found_critical_files += 1

        if found_critical_files < 3:
            print("✗ Some critical files missing in tarball")
            return False

        print("✓ Tarball contents verified: All patterns found")
        return True

    except tarfile.ReadError:
        print(f"✗ Invalid tarball file format: {tarball_path}")
        return False
    except Exception as e:
        print(f"✗ Error reading tarball file: {e}")
        return False


def main() -> int:
    """
    Main entry point for package integrity verification.

    Usage:
        python3 scripts/verify_package_integrity.py [wheel_or_tarball_path]

    Returns:
        int: 0 if all checks pass, non-zero on failure.
    """
    # First, always verify source files
    source_ok = verify_source_files()

    # If arguments provided, verify package contents
    if len(sys.argv) > 1:
        package_path = sys.argv[1]
        package_file = Path(package_path)

        if not package_file.exists():
            # Try to find the file using glob if it's a pattern
            import glob

            matches = glob.glob(str(PROJECT_ROOT / package_path))
            if not matches:
                print(
                    f"✗ Package file not found: {package_path}",
                    file=sys.stderr,
                )
                return 3
            package_path = matches[0]
            package_file = Path(package_path)

        if package_path.endswith(".whl"):
            package_ok = verify_wheel_contents(package_path)
        elif package_path.endswith(".tar.gz"):
            package_ok = verify_tarball_contents(package_path)
        else:
            print(
                f"✗ Unknown package format: {package_path}",
                file=sys.stderr,
            )
            return 2

        if not source_ok or not package_ok:
            return 1
    else:
        # Only source files verification
        if not source_ok:
            return 1

    print("✓ Package integrity: 100% complete")
    return 0


if __name__ == "__main__":
    sys.exit(main())
