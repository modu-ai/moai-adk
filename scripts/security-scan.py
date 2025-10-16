#!/usr/bin/env python3
# @CODE:SECURITY-001 | Cross-platform security scan script

"""
MoAI-ADK Security Scanner

Runs security scans using pip-audit and bandit.
Works on Windows, macOS, and Linux.
"""

import argparse
import subprocess
import sys
from pathlib import Path


def print_header(text: str) -> None:
    """Print a formatted header"""
    print(f"\n{text}")
    print("=" * len(text))


def print_step(step: str, description: str) -> None:
    """Print a step header"""
    print(f"\n🔍 {step}: {description}")
    print("-" * 70)


def check_tool_installed(tool: str) -> bool:
    """Check if a security tool is installed"""
    try:
        subprocess.run(
            [sys.executable, "-m", tool, "--version"],
            capture_output=True,
            check=True,
        )
        return True
    except (subprocess.CalledProcessError, FileNotFoundError):
        return False


def install_tool(tool: str) -> bool:
    """Install a security tool using pip"""
    print(f"Installing {tool}...")
    try:
        subprocess.run(
            [sys.executable, "-m", "pip", "install", tool],
            check=True,
        )
        return True
    except subprocess.CalledProcessError:
        print(f"❌ Failed to install {tool}")
        return False


def run_pip_audit(ignore_vulns: list[str] | None = None) -> bool:
    """Run pip-audit for dependency vulnerability scan

    Args:
        ignore_vulns: List of vulnerability IDs to ignore (e.g., ['GHSA-xxxx-xxxx-xxxx'])
    """
    print_step("Step 1", "Running pip-audit (dependency vulnerability scan)")

    cmd = [sys.executable, "-m", "pip_audit"]

    # Add ignore-vuln options
    if ignore_vulns:
        for vuln_id in ignore_vulns:
            cmd.extend(["--ignore-vuln", vuln_id])
            print(f"ℹ️ Ignoring vulnerability: {vuln_id}")

    try:
        subprocess.run(cmd, check=True)
        print("✅ No vulnerabilities found")
        return True
    except subprocess.CalledProcessError:
        print("⚠️ Vulnerabilities detected. Please review above.")
        return False


def run_bandit() -> bool:
    """Run bandit for code security scan"""
    print_step("Step 2", "Running bandit (code security scan)")

    # Find src directory
    script_dir = Path(__file__).parent
    project_root = script_dir.parent
    src_dir = project_root / "src"

    if not src_dir.exists():
        print(f"❌ Source directory not found: {src_dir}")
        return False

    try:
        # Run bandit with low-level severity filter (-ll)
        subprocess.run(
            [sys.executable, "-m", "bandit", "-r", str(src_dir), "-ll"],
            check=True,
        )
        print("✅ No high/medium security issues found")
        return True
    except subprocess.CalledProcessError:
        print("❌ Security issues detected. Please review above.")
        return False


def main() -> int:
    """Main security scan routine"""
    # Parse command-line arguments
    parser = argparse.ArgumentParser(
        description="MoAI-ADK Security Scanner",
        formatter_class=argparse.RawDescriptionHelpFormatter,
    )
    parser.add_argument(
        "--ignore-vuln",
        action="append",
        dest="ignore_vulns",
        metavar="ID",
        help="Ignore specific vulnerability ID (can be specified multiple times)",
    )
    args = parser.parse_args()

    print_header("🔍 MoAI-ADK Security Scan")

    # Check and install security tools
    print("\n📦 Checking security tools...")

    tools = ["pip_audit", "bandit"]
    for tool in tools:
        if not check_tool_installed(tool):
            if not install_tool(tool):
                return 1

    # Run security scans
    pip_audit_passed = run_pip_audit(ignore_vulns=args.ignore_vulns)
    bandit_passed = run_bandit()

    # Summary
    print("\n" + "=" * 70)
    if pip_audit_passed and bandit_passed:
        print("✅ All security scans passed!")
        return 0
    else:
        print("⚠️ Security scan completed with warnings/errors")
        print("   Please review the issues above and fix them.")
        return 1


if __name__ == "__main__":
    sys.exit(main())
