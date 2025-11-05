#!/usr/bin/env python3
"""
MoAI-ADK Documentation Build Script
UV-based documentation building with multilingual support
"""

import subprocess
import sys
import os
from pathlib import Path

def run_command(cmd, check=True):
    """Run command with proper error handling"""
    print(f"🔧 Running: {cmd}")
    result = subprocess.run(cmd, shell=True, capture_output=True, text=True)

    if check and result.returncode != 0:
        print(f"❌ Error: {result.stderr}")
        sys.exit(1)

    return result

def build_docs():
    """Build documentation using UV and MkDocs"""
    print("🚀 Building MoAI-ADK Documentation...")

    # Check if UV is available
    try:
        run_command("uv --version")
    except:
        print("❌ UV not found. Please install UV: pip install uv")
        sys.exit(1)

    # Install dependencies with UV
    print("📦 Installing dependencies with UV...")
    run_command("uv sync")

    # Build documentation
    print("📚 Building MkDocs site...")
    run_command("uv run mkdocs build --strict")

    # Check if build was successful
    site_dir = Path("site")
    if site_dir.exists():
        print(f"✅ Documentation built successfully!")
        print(f"📁 Output directory: {site_dir.absolute()}")

        # Count files
        html_files = list(site_dir.rglob("*.html"))
        print(f"📊 Generated {len(html_files)} HTML files")

        # Show directory structure
        print("\n📂 Build structure:")
        for item in sorted(site_dir.iterdir()):
            if item.is_dir():
                files_count = len(list(item.rglob("*")))
                print(f"  📁 {item.name}/ ({files_count} files)")
            else:
                size_mb = item.stat().st_size / (1024 * 1024)
                print(f"  📄 {item.name} ({size_mb:.1f} MB)")
    else:
        print("❌ Build failed - site directory not found")
        sys.exit(1)

def serve_docs():
    """Serve documentation locally"""
    print("🌐 Starting local documentation server...")

    # Install dependencies with UV
    run_command("uv sync")

    # Start development server
    print("🚀 Starting MkDocs development server...")
    print("📖 Documentation will be available at: http://127.0.0.1:8000")
    run_command("uv run mkdocs serve --dirtyreload", check=False)

def deploy_docs():
    """Deploy documentation to Vercel"""
    print("🚀 Preparing documentation for Vercel deployment...")

    # Build first
    build_docs()

    # Check for Vercel configuration
    vercel_config = Path("vercel.json")
    if not vercel_config.exists():
        print("❌ Vercel configuration not found")
        sys.exit(1)

    print("✅ Ready for Vercel deployment!")
    print("📝 Next steps:")
    print("  1. Install Vercel CLI: npm i -g vercel")
    print("  2. Run: vercel --prod")
    print("  3. Follow the prompts to deploy")

if __name__ == "__main__":
    if len(sys.argv) > 1:
        command = sys.argv[1]
        if command == "serve":
            serve_docs()
        elif command == "deploy":
            deploy_docs()
        else:
            print("Usage: python build_docs.py [serve|deploy]")
            sys.exit(1)
    else:
        build_docs()