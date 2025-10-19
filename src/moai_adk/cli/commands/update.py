"""Update command - Upgrade moai-adk package to the latest version"""
import json
import subprocess
import sys

import click
from packaging import version
from rich.console import Console

from moai_adk import __version__

console = Console()


def get_latest_version() -> str | None:
    """Get the latest version from PyPI.

    Returns:
        Latest version string, or None if fetch fails.
    """
    try:
        import urllib.error
        import urllib.request
        from typing import cast

        url = "https://pypi.org/pypi/moai-adk/json"
        with urllib.request.urlopen(url, timeout=5) as response:  # nosec B310 - URL is hardcoded HTTPS to PyPI API, no user input
            data = json.loads(response.read().decode("utf-8"))
            version_str: str = cast(str, data["info"]["version"])
            return version_str
    except (urllib.error.URLError, json.JSONDecodeError, KeyError, TimeoutError):
        # Return None if PyPI check fails
        return None


def detect_install_method() -> str:
    """Detect how moai-adk was installed.

    Returns:
        'uv-tool', 'uv-pip', or 'pip'
    """
    # Check if installed via uv tool
    try:
        result = subprocess.run(
            ["uv", "tool", "list"],
            capture_output=True,
            text=True,
            timeout=5,
            check=False
        )
        if result.returncode == 0 and "moai-adk" in result.stdout:
            return "uv-tool"
    except (FileNotFoundError, subprocess.TimeoutExpired):
        pass

    # Check if uv is available (for uv pip)
    try:
        subprocess.run(
            ["uv", "--version"],
            capture_output=True,
            timeout=5,
            check=False
        )
        return "uv-pip"
    except (FileNotFoundError, subprocess.TimeoutExpired):
        pass

    # Default to pip
    return "pip"


def upgrade_package(install_method: str, target_version: str) -> bool:
    """Upgrade moai-adk package.

    Args:
        install_method: 'uv-tool', 'uv-pip', or 'pip'
        target_version: Target version to upgrade to

    Returns:
        True if successful, False otherwise
    """
    commands = {
        "uv-tool": ["uv", "tool", "upgrade", "moai-adk"],
        "uv-pip": ["uv", "pip", "install", "--upgrade", "moai-adk"],
        "pip": [sys.executable, "-m", "pip", "install", "--upgrade", "moai-adk"],
    }

    cmd = commands.get(install_method)
    if not cmd:
        return False

    try:
        console.print(f"\n[cyan]📦 Upgrading via {install_method}...[/cyan]")
        console.print(f"[dim]   Command: {' '.join(cmd)}[/dim]")

        result = subprocess.run(
            cmd,
            capture_output=True,
            text=True,
            timeout=120,
            check=False
        )

        if result.returncode == 0:
            console.print(f"[green]✓ Upgraded to version {target_version}[/green]")
            return True
        else:
            console.print(f"[red]✗ Upgrade failed[/red]")
            if result.stderr:
                console.print(f"[dim]{result.stderr.strip()}[/dim]")
            return False

    except subprocess.TimeoutExpired:
        console.print("[red]✗ Upgrade timeout[/red]")
        return False
    except Exception as e:
        console.print(f"[red]✗ Upgrade error: {e}[/red]")
        return False


@click.command()
@click.option(
    "--check",
    is_flag=True,
    help="Only check version (do not upgrade)"
)
def update(check: bool) -> None:
    """Upgrade moai-adk package to the latest version.

    This command automatically detects the installation method
    (uv tool, uv pip, or pip) and upgrades the package accordingly.

    For template updates, use 'moai-adk init .' instead.

    Examples:
        moai-adk update              # Upgrade to latest version
        moai-adk update --check      # Check version only
    """
    try:
        # Phase 1: Check versions
        console.print("[cyan]🔍 Checking versions...[/cyan]")
        current_version = __version__
        latest_version = get_latest_version()

        # Handle PyPI fetch failure
        if latest_version is None:
            console.print(f"   Current version: {current_version}")
            console.print("   Latest version:  [yellow]Unable to fetch from PyPI[/yellow]")
            console.print("[yellow]⚠ Cannot check for updates[/yellow]")
            return

        console.print(f"   Current version: {current_version}")
        console.print(f"   Latest version:  {latest_version}")

        # Parse versions
        current_ver = version.parse(current_version)
        latest_ver = version.parse(latest_version)

        # Check mode
        if check:
            if current_ver < latest_ver:
                console.print("[yellow]⚠ Update available[/yellow]")
            elif current_ver > latest_ver:
                console.print("[green]✓ Development version (newer than PyPI)[/green]")
            else:
                console.print("[green]✓ Already up to date[/green]")
            return

        # Check if upgrade needed
        if current_ver >= latest_ver:
            console.print("[green]✓ Already up to date[/green]")
            return

        # Phase 2: Detect install method
        install_method = detect_install_method()
        console.print(f"\n[cyan]🔎 Detected installation method: {install_method}[/cyan]")

        # Phase 3: Upgrade package
        success = upgrade_package(install_method, latest_version)

        if success:
            console.print("\n[green]✓ Update complete![/green]")
            console.print("\n[dim]💡 For template updates, run: moai-adk init .[/dim]")
        else:
            console.print("\n[yellow]⚠ Upgrade failed. Please try manually:[/yellow]")
            if install_method == "uv-tool":
                console.print("   uv tool upgrade moai-adk")
            elif install_method == "uv-pip":
                console.print("   uv pip install --upgrade moai-adk")
            else:
                console.print("   pip install --upgrade moai-adk")
            raise click.Abort()

    except click.Abort:
        raise
    except Exception as e:
        console.print(f"[red]✗ Update failed: {e}[/red]")
        raise click.ClickException(str(e)) from e
