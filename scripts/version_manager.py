#!/usr/bin/env python3
"""
MoAI-ADK Unified Version Management System
Consolidated version management tool replacing bump_version.py, check_version_consistency.py, and update_version.py

Usage:
    python scripts/version_manager.py check                    # Check version consistency
    python scripts/version_manager.py bump patch              # Bump patch version (0.1.24 -> 0.1.25)
    python scripts/version_manager.py bump minor              # Bump minor version (0.1.24 -> 0.2.0)
    python scripts/version_manager.py bump major              # Bump major version (0.1.24 -> 1.0.0)
    python scripts/version_manager.py set 1.2.3               # Set specific version
    python scripts/version_manager.py status                  # Show current version status
    python scripts/version_manager.py sync                    # Synchronize all version files
    python scripts/version_manager.py --dry-run set 1.2.3     # Preview changes without applying
    python scripts/version_manager.py --verify set 1.2.3      # Apply and verify consistency
"""

import argparse
import json
import re
import sys
from datetime import datetime
from pathlib import Path
from typing import Dict, List, Optional, Tuple


class Colors:
    """ANSI color codes for terminal output"""
    RED = '\033[0;31m'
    GREEN = '\033[0;32m'
    YELLOW = '\033[1;33m'
    BLUE = '\033[0;34m'
    NC = '\033[0m'  # No Color

    @classmethod
    def disable_on_windows(cls):
        """Disable colors on Windows if colorama not available"""
        if sys.platform == 'win32':
            try:
                import colorama
                colorama.init()
            except ImportError:
                cls.RED = cls.GREEN = cls.YELLOW = cls.BLUE = cls.NC = ''


class VersionManager:
    def __init__(self, project_root: Optional[Path] = None):
        self.project_root = project_root or Path(__file__).parent.parent
        self.version_files = {
            "pyproject.toml": self.project_root / "pyproject.toml",
            "_version.py": self.project_root / "src" / "moai_adk" / "_version.py",
            "version.json": self.project_root / ".moai" / "version.json"
        }

        # Initialize colors
        Colors.disable_on_windows()

    def print_success(self, message: str) -> None:
        """Print success message"""
        print(f"{Colors.GREEN}✅ {message}{Colors.NC}")

    def print_error(self, message: str) -> None:
        """Print error message"""
        print(f"{Colors.RED}❌ {message}{Colors.NC}")

    def print_warning(self, message: str) -> None:
        """Print warning message"""
        print(f"{Colors.YELLOW}⚠️ {message}{Colors.NC}")

    def print_info(self, message: str) -> None:
        """Print info message"""
        print(f"{Colors.BLUE}ℹ️ {message}{Colors.NC}")

    def _get_pyproject_version(self) -> Optional[str]:
        """Extract version from pyproject.toml"""
        pyproject_path = self.version_files["pyproject.toml"]
        if not pyproject_path.exists():
            return None

        try:
            content = pyproject_path.read_text(encoding='utf-8')
            match = re.search(r'version = "([^"]+)"', content)
            return match.group(1) if match else None
        except Exception as e:
            self.print_error(f"Failed to read pyproject.toml: {e}")
            return None

    def _get_version_py_version(self) -> Optional[str]:
        """Extract version from _version.py"""
        version_py_path = self.version_files["_version.py"]
        if not version_py_path.exists():
            return None

        try:
            content = version_py_path.read_text(encoding='utf-8')
            match = re.search(r'__version__ = ["\']([^"\']+)["\']', content)
            return match.group(1) if match else None
        except Exception as e:
            self.print_error(f"Failed to read _version.py: {e}")
            return None

    def _get_version_json_version(self) -> Optional[str]:
        """Extract version from .moai/version.json"""
        version_json_path = self.version_files["version.json"]
        if not version_json_path.exists():
            return None

        try:
            with open(version_json_path, 'r', encoding='utf-8') as f:
                data = json.load(f)
                return data.get('version')
        except Exception as e:
            self.print_error(f"Failed to read version.json: {e}")
            return None

    def get_current_versions(self) -> Dict[str, Optional[str]]:
        """Get versions from all sources"""
        return {
            "pyproject.toml": self._get_pyproject_version(),
            "_version.py": self._get_version_py_version(),
            "version.json": self._get_version_json_version()
        }

    def check_consistency(self) -> Tuple[bool, Dict[str, Optional[str]]]:
        """Check if all version sources are consistent"""
        versions = self.get_current_versions()

        # Filter out None values
        valid_versions = {k: v for k, v in versions.items() if v is not None}

        if not valid_versions:
            self.print_error("No version files found!")
            return False, versions

        # Check if all versions are the same
        unique_versions = set(valid_versions.values())
        consistent = len(unique_versions) == 1

        return consistent, versions

    def _parse_version(self, version_str: str) -> Tuple[int, int, int]:
        """Parse semantic version string into (major, minor, patch) tuple"""
        match = re.match(r'^(\d+)\.(\d+)\.(\d+)', version_str)
        if not match:
            raise ValueError(f"Invalid version format: {version_str}")

        return tuple(map(int, match.groups()))

    def _format_version(self, major: int, minor: int, patch: int) -> str:
        """Format version tuple into string"""
        return f"{major}.{minor}.{patch}"

    def bump_version(self, current_version: str, bump_type: str) -> str:
        """Bump version according to type (major, minor, patch)"""
        major, minor, patch = self._parse_version(current_version)

        if bump_type == "major":
            return self._format_version(major + 1, 0, 0)
        elif bump_type == "minor":
            return self._format_version(major, minor + 1, 0)
        elif bump_type == "patch":
            return self._format_version(major, minor, patch + 1)
        else:
            raise ValueError(f"Invalid bump type: {bump_type}")

    def _update_pyproject_version(self, new_version: str) -> bool:
        """Update version in pyproject.toml"""
        pyproject_path = self.version_files["pyproject.toml"]
        if not pyproject_path.exists():
            return False

        try:
            content = pyproject_path.read_text(encoding='utf-8')
            new_content = re.sub(
                r'version = "[^"]+"',
                f'version = "{new_version}"',
                content
            )
            pyproject_path.write_text(new_content, encoding='utf-8')
            return True
        except Exception as e:
            self.print_error(f"Failed to update pyproject.toml: {e}")
            return False

    def _update_version_py_version(self, new_version: str) -> bool:
        """Update version in _version.py"""
        version_py_path = self.version_files["_version.py"]
        if not version_py_path.exists():
            return False

        try:
            content = version_py_path.read_text(encoding='utf-8')
            new_content = re.sub(
                r'__version__ = ["\'][^"\']+["\']',
                f'__version__ = "{new_version}"',
                content
            )
            version_py_path.write_text(new_content, encoding='utf-8')
            return True
        except Exception as e:
            self.print_error(f"Failed to update _version.py: {e}")
            return False

    def _update_version_json_version(self, new_version: str) -> bool:
        """Update version in .moai/version.json"""
        version_json_path = self.version_files["version.json"]

        try:
            # Create directory if it doesn't exist
            version_json_path.parent.mkdir(exist_ok=True)

            # Load existing data or create new
            if version_json_path.exists():
                with open(version_json_path, 'r', encoding='utf-8') as f:
                    data = json.load(f)
            else:
                data = {}

            # Update version and timestamp
            data['version'] = new_version
            data['updated_at'] = datetime.now().isoformat()

            # Write back to file
            with open(version_json_path, 'w', encoding='utf-8') as f:
                json.dump(data, f, indent=2, ensure_ascii=False)

            return True
        except Exception as e:
            self.print_error(f"Failed to update version.json: {e}")
            return False

    def update_all_versions(self, new_version: str, dry_run: bool = False) -> bool:
        """Update version in all files"""
        if dry_run:
            self.print_info(f"DRY RUN: Would update all version files to {new_version}")
            return True

        success = True

        # Update pyproject.toml
        if self.version_files["pyproject.toml"].exists():
            if self._update_pyproject_version(new_version):
                self.print_success(f"Updated pyproject.toml to {new_version}")
            else:
                self.print_error("Failed to update pyproject.toml")
                success = False

        # Update _version.py
        if self.version_files["_version.py"].exists():
            if self._update_version_py_version(new_version):
                self.print_success(f"Updated _version.py to {new_version}")
            else:
                self.print_error("Failed to update _version.py")
                success = False

        # Update version.json
        if self._update_version_json_version(new_version):
            self.print_success(f"Updated version.json to {new_version}")
        else:
            self.print_error("Failed to update version.json")
            success = False

        return success

    def show_status(self) -> None:
        """Show current version status"""
        self.print_info("MoAI-ADK Version Status")
        print("=" * 50)

        versions = self.get_current_versions()
        consistent, _ = self.check_consistency()

        for source, version in versions.items():
            status = f"✅ {version}" if version else "❌ Not found"
            print(f"{source:<15}: {status}")

        print("=" * 50)
        if consistent:
            self.print_success("All versions are consistent")
        else:
            self.print_error("Version inconsistency detected!")

    def sync_versions(self) -> bool:
        """Synchronize all version files to match primary source (pyproject.toml)"""
        primary_version = self._get_pyproject_version()
        if not primary_version:
            self.print_error("Primary version source (pyproject.toml) not found!")
            return False

        self.print_info(f"Synchronizing all versions to {primary_version}")
        return self.update_all_versions(primary_version)

    def cmd_check(self) -> bool:
        """Command: Check version consistency"""
        consistent, versions = self.check_consistency()

        print("🔍 Version Consistency Check")
        print("=" * 40)

        for source, version in versions.items():
            if version:
                print(f"  {source}: {version}")
            else:
                print(f"  {source}: ❌ Not found")

        print("=" * 40)

        if consistent:
            self.print_success("All versions are consistent!")
            return True
        else:
            self.print_error("Version inconsistency detected!")
            self.print_info("Run 'python scripts/version_manager.py sync' to fix")
            return False

    def cmd_bump(self, bump_type: str, dry_run: bool = False, verify: bool = False) -> bool:
        """Command: Bump version"""
        current_version = self._get_pyproject_version()
        if not current_version:
            self.print_error("Current version not found in pyproject.toml")
            return False

        try:
            new_version = self.bump_version(current_version, bump_type)
            self.print_info(f"Bumping {bump_type} version: {current_version} → {new_version}")

            success = self.update_all_versions(new_version, dry_run)

            if success and verify and not dry_run:
                return self.cmd_check()

            return success
        except ValueError as e:
            self.print_error(str(e))
            return False

    def cmd_set(self, new_version: str, dry_run: bool = False, verify: bool = False) -> bool:
        """Command: Set specific version"""
        try:
            # Validate version format
            self._parse_version(new_version)
        except ValueError as e:
            self.print_error(f"Invalid version format: {e}")
            return False

        current_versions = self.get_current_versions()
        primary_version = self._get_pyproject_version() or "unknown"

        self.print_info(f"Setting version: {primary_version} → {new_version}")

        success = self.update_all_versions(new_version, dry_run)

        if success and verify and not dry_run:
            return self.cmd_check()

        return success

    def cmd_status(self) -> bool:
        """Command: Show status"""
        self.show_status()
        return True

    def cmd_sync(self) -> bool:
        """Command: Synchronize versions"""
        return self.sync_versions()


def main():
    """Main function"""
    parser = argparse.ArgumentParser(
        description="MoAI-ADK Unified Version Management System",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog=__doc__
    )

    parser.add_argument('--dry-run', action='store_true',
                       help='Preview changes without applying them')
    parser.add_argument('--verify', action='store_true',
                       help='Verify consistency after making changes')

    subparsers = parser.add_subparsers(dest='command', help='Available commands')

    # Check command
    subparsers.add_parser('check', help='Check version consistency')

    # Bump command
    bump_parser = subparsers.add_parser('bump', help='Bump version')
    bump_parser.add_argument('type', choices=['major', 'minor', 'patch'],
                           help='Version component to bump')

    # Set command
    set_parser = subparsers.add_parser('set', help='Set specific version')
    set_parser.add_argument('version', help='New version (e.g., 1.2.3)')

    # Status command
    subparsers.add_parser('status', help='Show current version status')

    # Sync command
    subparsers.add_parser('sync', help='Synchronize all version files')

    args = parser.parse_args()

    if not args.command:
        parser.print_help()
        return 1

    manager = VersionManager()

    # Execute command
    try:
        if args.command == 'check':
            success = manager.cmd_check()
        elif args.command == 'bump':
            success = manager.cmd_bump(args.type, args.dry_run, args.verify)
        elif args.command == 'set':
            success = manager.cmd_set(args.version, args.dry_run, args.verify)
        elif args.command == 'status':
            success = manager.cmd_status()
        elif args.command == 'sync':
            success = manager.cmd_sync()
        else:
            parser.print_help()
            return 1

        return 0 if success else 1

    except KeyboardInterrupt:
        print("\n🛑 Operation cancelled by user")
        return 1
    except Exception as e:
        print(f"❌ Unexpected error: {e}")
        return 1


if __name__ == "__main__":
    sys.exit(main())