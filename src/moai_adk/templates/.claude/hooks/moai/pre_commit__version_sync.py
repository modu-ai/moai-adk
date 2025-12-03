#!/usr/bin/env python3
"""
Pre-commit Hook: Version Synchronization Check

Ensures version consistency across all configuration files before commit.
Runs automatically before each commit to prevent version drift.

Features:
- Checks version consistency across pyproject.toml, config.json, and __init__.py
- Auto-synchronizes versions if inconsistencies detected
- Clears caches to ensure statusline shows correct version
- Provides detailed feedback on version synchronization status
"""

import sys
from pathlib import Path

# Add src to path to import moai_adk modules
sys.path.insert(0, str(Path(__file__).parent.parent.parent.parent / "src"))

try:
    from moai_adk.core.version_sync import (  # noqa: F401 - check_project_versions for availability
        VersionSynchronizer,
        check_project_versions,
    )
except ImportError as e:
    print(f"❌ Error importing version sync module: {e}")
    print("Make sure you're running from the project root directory")
    sys.exit(1)


def main():
    """Main pre-commit hook execution"""
    print("🔍 Checking version consistency...")

    # Get project directory (parent of .git directory)
    current_file = Path(__file__).resolve()
    project_dir = current_file.parent.parent.parent.parent

    try:
        # Initialize synchronizer
        synchronizer = VersionSynchronizer(project_dir)

        # Check consistency
        is_consistent, version_infos = synchronizer.check_consistency()

        if not version_infos:
            print("❌ No version information found in project files")
            print("   Please initialize version in pyproject.toml first")
            sys.exit(1)

        if is_consistent:
            version = version_infos[0].version
            print(f"✅ Version consistency check passed: v{version}")

            # Clear caches to ensure statusline updates
            synchronizer._clear_caches()
            print("🧹 Caches cleared for statusline update")
            sys.exit(0)

        # Version inconsistency detected
        print("⚠️  Version inconsistency detected:")
        for info in version_infos:
            status = "✓" if info.is_valid else "✗"
            print(f"   {status} {info.source.value}: v{info.version} ({info.file_path.name})")

        # Get master version from pyproject.toml
        master_info = synchronizer.get_master_version()
        if not master_info:
            print("❌ Cannot determine master version from pyproject.toml")
            sys.exit(1)

        print(f"\n🔄 Auto-synchronizing to master version: v{master_info.version}")

        # Synchronize all files
        success = synchronizer.synchronize_all()

        if success:
            print("✅ Version synchronization completed successfully")
            print("🧹 Caches cleared for statusline update")
            print("📝 Please review the changes and stage them before committing")
            sys.exit(0)
        else:
            print("❌ Version synchronization failed")
            print("   Please manually fix version inconsistencies")
            sys.exit(1)

    except Exception as e:
        print(f"❌ Error during version check: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main()
