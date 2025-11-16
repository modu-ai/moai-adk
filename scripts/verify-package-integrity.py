#!/usr/bin/env python3
"""
Package Integrity Verification Script

Validates the integrity of MoAI-ADK package distribution files (wheel and tarball).
Checks for required files, correct structure, and proper metadata.
"""

import os
import sys
import json
import tarfile
import zipfile
from pathlib import Path
from typing import Optional, Tuple, List


def colored(text: str, color: str) -> str:
    """Add ANSI color codes to text."""
    colors = {
        "green": "\033[92m",
        "red": "\033[91m",
        "yellow": "\033[93m",
        "blue": "\033[94m",
        "reset": "\033[0m",
    }
    return f"{colors.get(color, '')}{text}{colors['reset']}"


class PackageIntegrityVerifier:
    """Verifies package integrity for distribution files."""

    def __init__(self):
        self.errors: List[str] = []
        self.warnings: List[str] = []
        self.checks_passed = 0
        self.checks_failed = 0

    def verify_wheel(self, wheel_path: str) -> bool:
        """Verify wheel package integrity."""
        print(f"\n{colored('Verifying Wheel Package', 'blue')}:")
        print(f"Path: {wheel_path}")

        if not os.path.exists(wheel_path):
            self.errors.append(f"Wheel file not found: {wheel_path}")
            self.checks_failed += 1
            return False

        try:
            with zipfile.ZipFile(wheel_path, 'r') as wheel:
                files = wheel.namelist()
                self._check_wheel_files(files, wheel_path)
                self._check_wheel_metadata(wheel, wheel_path)
                self.checks_passed += 1
                return len(self.errors) == 0
        except zipfile.BadZipFile:
            self.errors.append(f"Invalid wheel file: {wheel_path}")
            self.checks_failed += 1
            return False
        except Exception as e:
            self.errors.append(f"Error reading wheel: {str(e)}")
            self.checks_failed += 1
            return False

    def verify_tarball(self, tarball_path: str) -> bool:
        """Verify tarball package integrity."""
        print(f"\n{colored('Verifying Tarball Package', 'blue')}:")
        print(f"Path: {tarball_path}")

        if not os.path.exists(tarball_path):
            self.errors.append(f"Tarball file not found: {tarball_path}")
            self.checks_failed += 1
            return False

        try:
            with tarfile.open(tarball_path, 'r:*') as tar:
                members = tar.getmembers()
                self._check_tarball_files(members, tarball_path)
                self._check_tarball_metadata(tar, tarball_path)
                self.checks_passed += 1
                return len(self.errors) == 0
        except tarfile.ReadError:
            self.errors.append(f"Invalid tarball file: {tarball_path}")
            self.checks_failed += 1
            return False
        except Exception as e:
            self.errors.append(f"Error reading tarball: {str(e)}")
            self.checks_failed += 1
            return False

    def verify_source_files(self, project_root: str = ".") -> bool:
        """Verify source file structure."""
        print(f"\n{colored('Verifying Source Files', 'blue')}:")

        required_files = [
            "pyproject.toml",
            "README.md",
            "LICENSE",
            "src/moai_adk/__init__.py",
            "src/moai_adk/cli/main.py",
        ]

        required_dirs = [
            "src/moai_adk",
            "src/moai_adk/templates",
            "src/moai_adk/templates/.claude/output-styles/moai",
        ]

        all_exist = True

        # Check files
        for file_path in required_files:
            full_path = os.path.join(project_root, file_path)
            if os.path.exists(full_path):
                print(f"  {colored('✓', 'green')} {file_path}")
            else:
                print(f"  {colored('✗', 'red')} {file_path}")
                self.errors.append(f"Required file missing: {file_path}")
                all_exist = False

        # Check directories
        for dir_path in required_dirs:
            full_path = os.path.join(project_root, dir_path)
            if os.path.isdir(full_path):
                print(f"  {colored('✓', 'green')} {dir_path}/")
            else:
                print(f"  {colored('✗', 'red')} {dir_path}/")
                self.errors.append(f"Required directory missing: {dir_path}")
                all_exist = False

        # Check output-styles files
        output_styles_path = os.path.join(
            project_root, "src/moai_adk/templates/.claude/output-styles/moai"
        )
        if os.path.isdir(output_styles_path):
            files = os.listdir(output_styles_path)
            if len(files) >= 2:
                print(f"  {colored('✓', 'green')} output-styles contains {len(files)} files")
                for f in files:
                    print(f"    - {f}")
            else:
                self.warnings.append(
                    f"output-styles contains only {len(files)} files (expected >= 2)"
                )

        if all_exist:
            self.checks_passed += 1
        else:
            self.checks_failed += 1

        return all_exist and len(self.errors) == 0

    def _check_wheel_files(self, files: List[str], wheel_path: str) -> None:
        """Check required files in wheel."""
        required_patterns = [
            "moai_adk/",
            "moai_adk/__init__.py",
            "moai_adk/cli/",
            "moai_adk-",  # dist-info directory
        ]

        print("\n  Required files in wheel:")
        for pattern in required_patterns:
            found = any(pattern in f for f in files)
            status = colored("✓", "green") if found else colored("✗", "red")
            print(f"    {status} {pattern}")
            if not found:
                self.errors.append(f"Wheel missing expected files matching: {pattern}")

    def _check_wheel_metadata(self, wheel: zipfile.ZipFile, wheel_path: str) -> None:
        """Check wheel metadata."""
        print("\n  Metadata files:")
        metadata_files = [
            f for f in wheel.namelist() if "dist-info/METADATA" in f or "dist-info/WHEEL" in f
        ]

        for meta in metadata_files:
            try:
                content = wheel.read(meta).decode("utf-8")
                if "Name: moai-adk" in content or "moai-adk" in meta:
                    print(f"    {colored('✓', 'green')} {os.path.basename(meta)}")
                else:
                    self.warnings.append(f"Metadata may be incomplete: {meta}")
            except Exception as e:
                self.errors.append(f"Error reading metadata {meta}: {str(e)}")

    def _check_tarball_files(self, members: List[tarfile.TarInfo], tarball_path: str) -> None:
        """Check required files in tarball."""
        print("\n  Required files in tarball:")

        file_paths = {m.name for m in members if m.isfile()}

        required_patterns = [
            "pyproject.toml",
            "README.md",
            "LICENSE",
            "src/moai_adk/__init__.py",
        ]

        for pattern in required_patterns:
            found = any(pattern in p for p in file_paths)
            status = colored("✓", "green") if found else colored("✗", "red")
            print(f"    {status} {pattern}")
            if not found:
                self.errors.append(f"Tarball missing expected files matching: {pattern}")

    def _check_tarball_metadata(self, tar: tarfile.TarFile, tarball_path: str) -> None:
        """Check tarball metadata."""
        print("\n  Metadata files:")
        metadata_files = [m.name for m in tar.getmembers() if "PKG-INFO" in m.name]

        if metadata_files:
            for meta in metadata_files[:1]:  # Check first one
                try:
                    member = tar.getmember(meta)
                    f = tar.extractfile(member)
                    if f:
                        content = f.read().decode("utf-8")
                        if "Name: moai-adk" in content:
                            print(f"    {colored('✓', 'green')} PKG-INFO")
                        else:
                            self.warnings.append("PKG-INFO may be incomplete")
                except Exception as e:
                    self.errors.append(f"Error reading PKG-INFO: {str(e)}")
        else:
            self.warnings.append("No PKG-INFO found in tarball")

    def print_summary(self) -> None:
        """Print verification summary."""
        print(f"\n{colored('=' * 60, 'blue')}")
        print(f"{colored('Verification Summary', 'blue')}")
        print(f"{colored('=' * 60, 'blue')}")

        if self.checks_passed > 0:
            print(f"{colored('✓', 'green')} Checks Passed: {self.checks_passed}")
        if self.checks_failed > 0:
            print(f"{colored('✗', 'red')} Checks Failed: {self.checks_failed}")

        if self.warnings:
            print(f"\n{colored('Warnings:', 'yellow')}")
            for warning in self.warnings:
                print(f"  ⚠ {warning}")

        if self.errors:
            print(f"\n{colored('Errors:', 'red')}")
            for error in self.errors:
                print(f"  ✗ {error}")
        else:
            print(f"\n{colored('All checks passed!', 'green')}")

    def has_errors(self) -> bool:
        """Check if verification has errors."""
        return len(self.errors) > 0

    def exit_code(self) -> int:
        """Return appropriate exit code."""
        return 1 if self.has_errors() else 0


def main():
    """Main entry point."""
    if len(sys.argv) == 1:
        # No arguments - verify source files
        print(f"{colored('MoAI-ADK Package Integrity Verification', 'blue')}\n")
        verifier = PackageIntegrityVerifier()
        verifier.verify_source_files()
        verifier.print_summary()
        sys.exit(verifier.exit_code())

    else:
        # Verify specific package files
        print(f"{colored('MoAI-ADK Package Integrity Verification', 'blue')}\n")
        verifier = PackageIntegrityVerifier()

        for arg in sys.argv[1:]:
            if arg.endswith(".whl"):
                verifier.verify_wheel(arg)
            elif arg.endswith((".tar.gz", ".tar", ".tgz")):
                verifier.verify_tarball(arg)
            else:
                print(f"Unknown file type: {arg}")

        verifier.print_summary()
        sys.exit(verifier.exit_code())


if __name__ == "__main__":
    main()
