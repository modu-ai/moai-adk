#!/usr/bin/env python3
"""
MoAI-ADK Build Script
Cross-platform Python replacement for build.sh
"""

import os
import sys
import shutil
import subprocess
from pathlib import Path


def print_section(message: str, char: str = "=", length: int = 50) -> None:
    """Print formatted section header"""
    print(f"🗿 {message}")
    print(char * length)


def run_command(cmd: list[str], description: str) -> bool:
    """Run command and handle errors"""
    try:
        result = subprocess.run(cmd, check=True, capture_output=True, text=True)
        if result.stdout.strip():
            print(result.stdout.strip())
        return True
    except subprocess.CalledProcessError as e:
        print(f"❌ {description} failed!")
        if e.stdout:
            print("STDOUT:", e.stdout)
        if e.stderr:
            print("STDERR:", e.stderr)
        return False


def clean_build_artifacts(project_root: Path) -> bool:
    """Clean previous build artifacts"""
    print("🧹 Step 2: Cleaning previous build artifacts...")

    artifacts = ["dist", "build"]
    egg_info_patterns = list(project_root.glob("*.egg-info"))

    cleaned = []

    for artifact_name in artifacts:
        artifact_path = project_root / artifact_name
        if artifact_path.exists():
            shutil.rmtree(artifact_path)
            cleaned.append(artifact_name)

    for egg_info in egg_info_patterns:
        shutil.rmtree(egg_info)
        cleaned.append(egg_info.name)

    if cleaned:
        print(f"Cleaned: {', '.join(cleaned)}")
    else:
        print("No artifacts to clean")

    return True


def verify_build_artifacts(project_root: Path) -> bool:
    """Verify build artifacts were created"""
    print("✅ Step 4: Verifying build artifacts...")

    dist_dir = project_root / "dist"
    if not dist_dir.exists():
        print("❌ No dist directory found!")
        return False

    artifacts = list(dist_dir.iterdir())
    if not artifacts:
        print("❌ No build artifacts found!")
        return False

    print("Build artifacts created:")
    for artifact in sorted(artifacts):
        size = artifact.stat().st_size
        print(f"  {artifact.name} ({size:,} bytes)")

    return True


def show_next_steps(project_root: Path) -> None:
    """Show optional next steps"""
    print("=" * 50)
    print("🗿 MoAI-ADK build completed successfully!")
    print("")
    print("📦 Next steps (optional):")

    # Find wheel file for install example
    dist_dir = project_root / "dist"
    wheel_files = list(dist_dir.glob("*.whl"))
    if wheel_files:
        wheel_name = wheel_files[0].name
        print(f"  • Test install: pip install dist/{wheel_name}")
    else:
        print("  • Test install: pip install dist/*.whl")

    print("  • Upload to PyPI: python -m twine upload dist/*")

    # Try to get version for tag example
    try:
        version_cmd = [
            sys.executable, "-c",
            "from src.moai_adk._version import __version__; print(__version__)"
        ]
        result = subprocess.run(version_cmd, capture_output=True, text=True, cwd=project_root)
        if result.returncode == 0:
            version = result.stdout.strip()
            print(f"  • Create Git tag: git tag v{version}")
        else:
            print("  • Create Git tag: git tag vX.Y.Z")
    except Exception:
        print("  • Create Git tag: git tag vX.Y.Z")


def main():
    """Main build function"""
    print_section("MoAI-ADK Build Script")

    # Get project root (parent of scripts directory)
    script_dir = Path(__file__).parent
    project_root = script_dir.parent

    # Change to project root
    os.chdir(project_root)
    print(f"Working directory: {project_root}")

    # Step 1: Pre-build version synchronization
    print("🔄 Step 1: Pre-build version synchronization...")
    build_hooks_path = project_root / "build_hooks.py"
    if build_hooks_path.exists():
        if not run_command([sys.executable, "build_hooks.py", "--pre-build"],
                          "Pre-build version sync"):
            return False
    else:
        print("⚠️  build_hooks.py not found, skipping version sync")

    # Step 2: Clean previous build artifacts
    if not clean_build_artifacts(project_root):
        return False

    # Step 3: Build package
    print("📦 Step 3: Building package...")
    if not run_command([sys.executable, "-m", "build"], "Package build"):
        return False

    # Step 4: Verify build artifacts
    if not verify_build_artifacts(project_root):
        return False

    # Show next steps
    show_next_steps(project_root)

    return True


if __name__ == "__main__":
    success = main()
    sys.exit(0 if success else 1)