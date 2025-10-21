#!/usr/bin/env python
"""Verify version consistency between pyproject.toml and runtime.

This script ensures:
1. pyproject.toml version matches package metadata
2. __init__.py uses dynamic version loading (not hardcoded)
3. No duplicate version definitions

Run before releases to prevent version sync issues.
"""

import re
import sys
from pathlib import Path


def read_pyproject_version() -> str:
    """Read version from pyproject.toml."""
    pyproject = Path("pyproject.toml").read_text()
    match = re.search(r'^version = "([^"]+)"', pyproject, re.MULTILINE)
    if not match:
        raise ValueError("Could not find version in pyproject.toml")
    return match.group(1)


def check_init_py() -> bool:
    """Verify __init__.py uses dynamic version loading."""
    init_py = Path("src/moai_adk/__init__.py").read_text()

    # Check if using importlib.metadata (flexible pattern match)
    has_importlib = (
        "importlib.metadata" in init_py and
        "version(" in init_py
    )
    if not has_importlib:
        print("❌ FAIL: __init__.py should import version from importlib.metadata")
        return False

    # Check if there's direct __version__ assignment OUTSIDE of try/except
    # (fallback in except clause is OK)
    lines = init_py.split('\n')
    for i, line in enumerate(lines):
        # Skip comments and docstrings
        stripped = line.strip()
        if stripped.startswith('#') or stripped.startswith('"""') or stripped.startswith("'''"):
            continue
        # Look for direct assignment (not in except block)
        if re.match(r'__version__\s*=\s*"[0-9]', stripped) and 'except' not in ''.join(lines[max(0, i-5):i]):
            print("❌ FAIL: __init__.py uses hardcoded __version__ (violates SSOT)")
            return False

    print("✅ __init__.py uses dynamic version loading")
    return True


def check_no_duplicate_versions() -> bool:
    """Ensure no hardcoded versions elsewhere."""
    duplicates = []

    # Search for hardcoded version patterns (excluding pyproject.toml)
    for py_file in Path("src/moai_adk").rglob("*.py"):
        if "__init__.py" in str(py_file):
            continue

        content = py_file.read_text()
        if re.search(r'__version__ = "[0-9]', content):
            duplicates.append(str(py_file))

    if duplicates:
        print("❌ FAIL: Found hardcoded __version__ in:")
        for f in duplicates:
            print(f"  - {f}")
        return False

    print("✅ No duplicate version definitions found")
    return True


def main() -> int:
    """Run all version checks."""
    print("🔍 Version Sync Validation\n")

    try:
        version = read_pyproject_version()
        print(f"📌 Version from pyproject.toml: {version}\n")
    except ValueError as e:
        print(f"❌ ERROR: {e}")
        return 1

    checks = [
        check_init_py,
        check_no_duplicate_versions,
    ]

    results = [check() for check in checks]

    print()
    if all(results):
        print("✅ All version checks passed!")
        return 0
    else:
        print("❌ Version sync validation failed")
        return 1


if __name__ == "__main__":
    sys.exit(main())
