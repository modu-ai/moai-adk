"""LLM Backend Switch Command

Switch between Claude and GLM backends for hybrid mode.

Usage:
    moai switch glm      # Switch to GLM backend
    moai switch claude   # Switch to Claude backend
    moai switch status   # Show current backend status
"""

import json
from pathlib import Path

import click
from rich.console import Console
from rich.panel import Panel

console = Console()

# GLM environment variables to add/remove
GLM_ENV_KEYS = [
    "ANTHROPIC_AUTH_TOKEN",
    "ANTHROPIC_BASE_URL",
    "ANTHROPIC_DEFAULT_HAIKU_MODEL",
    "ANTHROPIC_DEFAULT_SONNET_MODEL",
    "ANTHROPIC_DEFAULT_OPUS_MODEL",
]


@click.command()
@click.argument("backend", type=click.Choice(["glm", "claude", "status"]), default="status")
def switch(backend: str) -> None:
    """Switch LLM backend between Claude and GLM.

    \b
    BACKEND options:
      glm     - Switch to GLM backend (cost-effective)
      claude  - Switch to Claude backend (default)
      status  - Show current backend status
    """
    project_path = Path.cwd()
    claude_dir = project_path / ".claude"
    settings_local = claude_dir / "settings.local.json"
    glm_config_source = project_path / ".moai" / "llm-configs" / "glm.json"

    # Check if project is initialized
    if not claude_dir.exists():
        console.print(
            "[red]Error:[/red] Not a MoAI project. Run 'moai init' first.",
            style="red",
        )
        raise click.Abort()

    if backend == "status":
        _show_status(settings_local)
    elif backend == "glm":
        _switch_to_glm(settings_local, glm_config_source)
    elif backend == "claude":
        _switch_to_claude(settings_local)


def _has_glm_env(settings_local: Path) -> bool:
    """Check if settings.local.json has GLM environment variables."""
    if not settings_local.exists():
        return False
    try:
        data = json.loads(settings_local.read_text())
        env = data.get("env", {})
        return "ANTHROPIC_BASE_URL" in env
    except (json.JSONDecodeError, OSError):
        return False


def _show_status(settings_local: Path) -> None:
    """Show current backend status."""
    if _has_glm_env(settings_local):
        console.print(
            Panel(
                "[cyan bold]GLM[/cyan bold] backend is active\n\n"
                "[dim]Using: .claude/settings.local.json[/dim]\n"
                "[dim]API: GLM CodePlan (cost-effective)[/dim]",
                title="[yellow]Current Backend[/yellow]",
                border_style="cyan",
            )
        )
        console.print("\n[dim]To switch to Claude: [cyan]moai switch claude[/cyan][/dim]")
    else:
        console.print(
            Panel(
                "[green bold]Claude[/green bold] backend is active\n\n[dim]API: Anthropic Claude (default)[/dim]",
                title="[yellow]Current Backend[/yellow]",
                border_style="green",
            )
        )
        console.print("\n[dim]To switch to GLM: [cyan]moai switch glm[/cyan][/dim]")


def _switch_to_glm(settings_local: Path, glm_config_source: Path) -> None:
    """Switch to GLM backend by merging GLM env into settings.local.json."""
    # Check if already using GLM
    if _has_glm_env(settings_local):
        console.print("[yellow]Already using GLM backend.[/yellow]")
        return

    # Check if GLM config template exists
    if not glm_config_source.exists():
        console.print(
            "[red]Error:[/red] GLM config not found at .moai/llm-configs/glm.json\n"
            "[dim]Run 'moai init' or create the config manually.[/dim]",
            style="red",
        )
        raise click.Abort()

    # Load GLM config
    glm_data = json.loads(glm_config_source.read_text())
    glm_env = glm_data.get("env", {})

    # Load or create settings.local.json
    if settings_local.exists():
        try:
            local_data = json.loads(settings_local.read_text())
        except json.JSONDecodeError:
            local_data = {}
    else:
        local_data = {}

    # Merge GLM env into settings.local.json
    if "env" not in local_data:
        local_data["env"] = {}
    local_data["env"].update(glm_env)

    # Write back
    settings_local.write_text(json.dumps(local_data, indent=2) + "\n")

    console.print(
        Panel(
            "[cyan bold]Switched to GLM[/cyan bold] backend\n\n"
            "[dim]Added GLM env to: .claude/settings.local.json[/dim]",
            title="[green]Backend Switched[/green]",
            border_style="cyan",
        )
    )
    console.print("\n[yellow]Restart Claude Code to apply changes.[/yellow]")


def _switch_to_claude(settings_local: Path) -> None:
    """Switch to Claude backend by removing GLM env from settings.local.json."""
    # Check if already using Claude
    if not _has_glm_env(settings_local):
        console.print("[yellow]Already using Claude backend.[/yellow]")
        return

    # Load settings.local.json
    try:
        local_data = json.loads(settings_local.read_text())
    except (json.JSONDecodeError, OSError):
        console.print("[yellow]Already using Claude backend.[/yellow]")
        return

    # Remove GLM env keys
    if "env" in local_data:
        for key in GLM_ENV_KEYS:
            local_data["env"].pop(key, None)

        # Remove env section if empty
        if not local_data["env"]:
            del local_data["env"]

    # Write back
    settings_local.write_text(json.dumps(local_data, indent=2) + "\n")

    console.print(
        Panel(
            "[green bold]Switched to Claude[/green bold] backend\n\n"
            "[dim]Removed GLM env from: .claude/settings.local.json[/dim]",
            title="[green]Backend Switched[/green]",
            border_style="green",
        )
    )
    console.print("\n[yellow]Restart Claude Code to apply changes.[/yellow]")
