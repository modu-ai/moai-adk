# -*- coding: utf-8 -*-
"""LSP Environment Setup for Claude Code

Automatically configures Claude Code environment to find LSP servers
installed via nvm, brew, and other package managers.

## Functionality

- Detects common LSP server installation paths
- Updates ~/.claude.json with proper PATH configuration
- Supports nvm (Node.js), pyenv (Python), and other version managers

## Usage

Called automatically by:
- moai init (after project initialization)
- moai update (after template update)

Can also be called standalone:
```bash
moai-lsp-setup
```
"""

import json
import os
import platform
import sys
from pathlib import Path
from typing import List

from rich.console import Console

# Force UTF-8 encoding for Windows compatibility
# Windows PowerShell/Console uses 'charmap' by default, which can't encode emojis
if sys.platform == "win32":
    console = Console(force_terminal=True, legacy_windows=False)
else:
    console = Console()


def get_nvm_paths() -> List[str]:
    """Get nvm (Node.js) installation paths.

    Returns:
        List of nvm bin directories to add to PATH
    """
    paths = []

    # Check NVM_DIR environment variable
    nvm_dir = os.environ.get("NVM_DIR", os.path.expanduser("~/.nvm"))

    # Default nvm versions directory
    nvm_versions = Path(nvm_dir) / "versions" / "node"
    if nvm_versions.exists():
        # Add all installed Node.js versions
        for version_dir in nvm_versions.iterdir():
            if version_dir.is_dir() and (version_dir / "bin").exists():
                paths.append(str(version_dir / "bin"))

    # Also add user bin directory
    user_bin = Path.home() / "bin"
    if user_bin.exists():
        paths.append(str(user_bin))

    return paths


def get_pyenv_paths() -> List[str]:
    """Get pyenv (Python) installation paths.

    Returns:
        List of pyenv shims directory to add to PATH
    """
    paths = []

    # Check PYENV_ROOT environment variable
    pyenv_root = os.environ.get("PYENV_ROOT", os.path.expanduser("~/.pyenv"))
    pyenv_shims = Path(pyenv_root) / "shims"

    if pyenv_shims.exists():
        paths.append(str(pyenv_shims))

    return paths


def get_brew_paths() -> List[str]:
    """Get Homebrew installation paths (macOS/Linux).

    Returns:
        List of Homebrew bin directories
    """
    paths = []

    if platform.system() == "Darwin":  # macOS
        # Apple Silicon Homebrew
        brew_arm = Path("/opt/homebrew/bin")
        if brew_arm.exists():
            paths.append(str(brew_arm))

        # Intel Homebrew
        brew_intel = Path("/usr/local/bin")
        if brew_intel.exists():
            paths.append(str(brew_intel))
    elif platform.system() == "Linux":
        # Linux Homebrew
        brew_linux = Path("/home/linuxbrew/.linuxbrew/bin")
        if brew_linux.exists():
            paths.append(str(brew_linux))

    return paths


def get_cargo_paths() -> List[str]:
    """Get Cargo (Rust) installation paths.

    Returns:
        List of Cargo bin directories
    """
    paths = []

    cargo_bin = Path.home() / ".cargo" / "bin"
    if cargo_bin.exists():
        paths.append(str(cargo_bin))

    return paths


def get_gem_paths() -> List[str]:
    """Get RubyGems installation paths.

    Returns:
        List of gem bin directories
    """
    paths = []

    # Check GEM_HOME environment variable
    gem_home = os.environ.get("GEM_HOME")
    if gem_home:
        gem_bin = Path(gem_home) / "bin"
        if gem_bin.exists():
            paths.append(str(gem_bin))

    # Default user gem path
    user_gem_bin = Path.home() / ".gem" / "ruby" / "*" / "bin"
    if user_gem_bin.parent.exists():
        for gem_bin in user_gem_bin.parent.glob("*/bin"):
            if gem_bin.exists():
                paths.append(str(gem_bin))

    return paths


def detect_lsp_paths() -> List[str]:
    """Detect all LSP server installation paths.

    Returns:
        List of directories to add to PATH

    Priority:
        1. nvm (Node.js) - typescript-language-server, etc.
        2. pyenv (Python) - pyright, pylsp
        3. Homebrew - gopls, rust-analyzer, clangd
        4. Cargo (Rust) - rust-analyzer
        5. RubyGems - solargraph
    """
    all_paths = []

    # Detect paths from various package managers
    all_paths.extend(get_nvm_paths())
    all_paths.extend(get_pyenv_paths())
    all_paths.extend(get_brew_paths())
    all_paths.extend(get_cargo_paths())
    all_paths.extend(get_gem_paths())

    # Remove duplicates while preserving order
    seen = set()
    unique_paths = []
    for path in all_paths:
        if path not in seen:
            seen.add(path)
            unique_paths.append(path)

    return unique_paths


def update_claude_env(paths: List[str]) -> bool:
    """Update ~/.claude.json with LSP server paths.

    Args:
        paths: List of directories to add to PATH

    Returns:
        True if successful, False otherwise
    """
    claude_json = Path.home() / ".claude.json"

    if not claude_json.exists():
        console.print("[yellow]⚠️  ~/.claude.json not found. Claude Code may not be installed.[/yellow]")
        return False

    try:
        # Read current configuration
        with open(claude_json, "r", encoding="utf-8") as f:
            config = json.load(f)

        # Initialize env section if it doesn't exist
        if "env" not in config:
            config["env"] = {}

        # Get current PATH or initialize with $PATH
        current_path = config["env"].get("PATH", "$PATH")

        # Parse current PATH entries
        current_entries = set(current_path.split(":"))

        # Add new paths if not already present
        new_entries = set(paths)
        entries_to_add = new_entries - current_entries

        if entries_to_add:
            # Build new PATH string
            new_path = "$PATH:" + ":".join(sorted(entries_to_add))
            config["env"]["PATH"] = new_path

            # Write updated configuration
            with open(claude_json, "w", encoding="utf-8") as f:
                json.dump(config, f, indent=2, ensure_ascii=False)

            return True
        else:
            return False

    except (json.JSONDecodeError, IOError) as e:
        console.print(f"[red]✗ Failed to update ~/.claude.json: {e}[/red]")
        return False


def setup_lsp_environment(verbose: bool = True) -> bool:
    """Setup Claude Code environment for LSP servers.

    Detects LSP server installation paths and updates ~/.claude.json
    to include them in PATH.

    Args:
        verbose: Print detailed messages

    Returns:
        True if configuration was updated, False otherwise
    """
    if verbose:
        console.print("[cyan]🔍 Detecting LSP server installation paths...[/cyan]")

    # Detect all LSP paths
    lsp_paths = detect_lsp_paths()

    if not lsp_paths:
        if verbose:
            console.print("[yellow]⚠️  No LSP server paths detected.[/yellow]")
            console.print("[dim]   This is normal if you haven't installed LSP servers yet.[/dim]")
        return False

    if verbose:
        console.print("[green]✓[/green] Found paths:")
        for path in lsp_paths:
            console.print(f"  [dim]•[/dim] {path}")

    # Update Claude Code environment
    updated = update_claude_env(lsp_paths)

    if updated:
        if verbose:
            console.print("[green]✓[/green] Updated [cyan]~/.claude.json[/cyan] with LSP paths")
            console.print("[dim]   Please restart Claude Code for changes to take effect.[/dim]")
        return True
    else:
        if verbose:
            console.print("[dim]✓[/dim] Claude Code environment already configured")
        return False


def main() -> None:
    """CLI entry point for standalone execution."""
    try:
        import click

        @click.command()
        @click.option(
            "--verbose",
            "-v",
            is_flag=True,
            default=True,
            help="Show detailed messages",
        )
        def lsp_setup_cli(verbose: bool) -> None:
            """Setup Claude Code environment for LSP servers."""
            if setup_lsp_environment(verbose=verbose):
                console.print("\n[green]✅ LSP environment configured successfully![/green]\n")
            else:
                console.print("\n[dim]ℹ️  LSP environment already up to date[/dim]\n")

        lsp_setup_cli()

    except ImportError:
        # Click not available, run directly
        if setup_lsp_environment(verbose=True):
            console.print("\n[green]✅ LSP environment configured successfully![/green]\n")
        else:
            console.print("\n[dim]ℹ️  LSP environment already up to date[/dim]\n")


if __name__ == "__main__":
    main()
