"""LLM Backend Switch Command

Switch between Claude and GLM backends for hybrid mode.

Usage:
    moai glm      # Switch to GLM backend
    moai claude   # Switch to Claude backend
"""

import json
import os
import re
import sys
from pathlib import Path

import click
from rich.console import Console
from rich.panel import Panel

# Force UTF-8 encoding for Windows compatibility
# Windows PowerShell/Console uses 'charmap' by default, which can't encode emojis
if sys.platform == "win32":
    console = Console(force_terminal=True, legacy_windows=False)
else:
    console = Console()


def _get_credential_value(var_name: str) -> str | None:
    """Get credential value from multiple sources.

    Priority for GLM keys:
    1. ~/.moai/.env.glm (explicit user config takes precedence)
    2. ~/.moai/credentials.yaml (legacy)
    3. Environment variable (fallback for CI/CD)

    Priority for other keys:
    1. Environment variable
    2. ~/.moai/credentials.yaml

    Note: .env.glm file is prioritized over environment variable because
    users explicitly save keys there via 'moai glm' command. Environment
    variables may contain stale values from previous sessions.

    Args:
        var_name: Variable name (e.g., GLM_API_KEY, ANTHROPIC_API_KEY)

    Returns:
        Credential value or None if not found
    """
    # For GLM keys, check .env.glm file first (user's explicit choice)
    if var_name in ("GLM_API_KEY", "GLM_API_TOKEN"):
        from moai_adk.core.credentials import load_glm_key_from_env

        glm_key = load_glm_key_from_env()
        if glm_key:
            return glm_key

    # Fall back to credentials file (legacy)
    from moai_adk.core.credentials import load_credentials

    creds = load_credentials()

    # Map variable names to credential keys
    var_to_cred = {
        "GLM_API_KEY": "glm_api_key",
        "GLM_API_TOKEN": "glm_api_key",  # Alias
        "ANTHROPIC_API_KEY": "anthropic_api_key",
    }

    cred_key = var_to_cred.get(var_name)
    if cred_key:
        creds_value = creds.get(cred_key)
        if creds_value:
            return creds_value

    # Finally check environment variable (fallback for CI/CD or manual export)
    env_value = os.environ.get(var_name)
    if env_value:
        return env_value

    return None


def _substitute_env_vars(value: str) -> tuple[str, list[str]]:
    """Substitute ${VAR} patterns with credential or environment values.

    Args:
        value: String that may contain ${VAR} patterns.

    Returns:
        Tuple of (substituted_value, list_of_missing_vars).
    """
    missing_vars: list[str] = []
    pattern = r"\$\{(\w+)\}"

    def replace(match: re.Match[str]) -> str:
        var_name = match.group(1)
        cred_value = _get_credential_value(var_name)
        if cred_value is None:
            missing_vars.append(var_name)
            return match.group(0)  # Keep original if not found
        return cred_value

    result = re.sub(pattern, replace, value)
    return result, missing_vars


# GLM environment variables to add/remove
GLM_ENV_KEYS = [
    "ANTHROPIC_AUTH_TOKEN",
    "ANTHROPIC_BASE_URL",
    "ANTHROPIC_DEFAULT_HAIKU_MODEL",
    "ANTHROPIC_DEFAULT_SONNET_MODEL",
    "ANTHROPIC_DEFAULT_OPUS_MODEL",
]


def _get_paths() -> tuple[Path, Path, Path]:
    """Get common paths for switch operations."""
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

    return claude_dir, settings_local, glm_config_source


def update_glm_key(api_key: str) -> None:
    """Update GLM API key without switching backend.

    Args:
        api_key: GLM API key to save to ~/.moai/.env.glm.
    """
    from moai_adk.core.credentials import (
        get_env_glm_path,
        save_glm_key_to_env,
    )

    # Save provided key to .env.glm
    save_glm_key_to_env(api_key)
    console.print(f"[green]✓[/green] GLM API key updated: [cyan]{get_env_glm_path()}[/cyan]")
    console.print("[dim]Use 'moai glm' to switch to GLM backend.[/dim]")

    # Check if GLM_API_KEY is set in environment (likely from ~/.zshrc or ~/.bashrc)
    # If so, automatically remove it to avoid duplicate configuration
    if os.environ.get("GLM_API_KEY"):
        from moai_adk.core.credentials import remove_glm_key_from_shell_config

        results = remove_glm_key_from_shell_config()
        modified_files = [name for name, modified in results.items() if modified]

        if modified_files:
            files_str = ", ~/.".join(modified_files)
            console.print(f"\n[green]✓[/green] Removed GLM_API_KEY from: [cyan]~/.{files_str}[/cyan]")
            console.print("[dim]Backup created with .moai-backup extension[/dim]")
            console.print("[yellow]⚠️  Run 'source ~/.zshrc' or restart terminal[/yellow]\n")
        else:
            console.print("\n[yellow]⚠️  GLM_API_KEY env var detected but not in shell config.[/yellow]")
            console.print("[dim]You may need to manually check your shell configuration.[/dim]\n")

        console.print(f"[dim]MoAI-ADK will now use: [cyan]{get_env_glm_path()}[/cyan][/dim]\n")


def switch_to_glm() -> None:
    """Switch to GLM backend. Called by 'moai glm' command.

    Does not prompt for API key. Use 'moai glm <key>' to update key first.
    """
    from moai_adk.core.credentials import (
        glm_env_exists,
        load_glm_key_from_env,
    )

    # Verify GLM key exists before switching
    if not glm_env_exists():
        console.print(
            "[red]Error:[/red] GLM API key not found.\n"
            "[dim]Please run: [cyan]moai glm <your-api-key>[/cyan] to set your key first.[/dim]"
        )
        raise click.Abort()

    # Show existing key info
    existing_key = load_glm_key_from_env()
    masked_key = existing_key[:8] + "..." if existing_key else "..."
    console.print(f"[green]✓[/green] Using GLM API key: [cyan]{masked_key}[/cyan]")

    # Continue with backend switch
    _, settings_local, glm_config_source = _get_paths()
    _switch_to_glm(settings_local, glm_config_source)


def switch_to_claude() -> None:
    """Switch to Claude backend. Called by 'moai claude' command."""
    _, settings_local, _ = _get_paths()
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

    # Substitute environment variables in GLM env values
    all_missing_vars: list[str] = []
    substituted_env: dict[str, str] = {}
    for key, value in glm_env.items():
        if isinstance(value, str):
            substituted_value, missing = _substitute_env_vars(value)
            substituted_env[key] = substituted_value
            all_missing_vars.extend(missing)
        else:
            substituted_env[key] = value

    # Warn if any credentials are missing
    if all_missing_vars:
        unique_missing = list(set(all_missing_vars))
        console.print(
            f"[red]Error:[/red] Missing credential(s): "
            f"[cyan]{', '.join(unique_missing)}[/cyan]\n\n"
            "[dim]Please run 'moai init' or 'moai update' to configure your API key,[/dim]\n"
            "[dim]or set it manually in ~/.moai/credentials.yaml:[/dim]\n\n"
            "[yellow]  glm_api_key: your-api-token-here[/yellow]\n",
            style="red",
        )
        raise click.Abort()

    # Load or create settings.local.json
    if settings_local.exists():
        try:
            local_data = json.loads(settings_local.read_text())
        except json.JSONDecodeError:
            local_data = {}
    else:
        local_data = {}

    # Merge substituted GLM env into settings.local.json
    if "env" not in local_data:
        local_data["env"] = {}
    local_data["env"].update(substituted_env)

    # Write back
    settings_local.write_text(json.dumps(local_data, indent=2) + "\n")

    console.print(
        Panel(
            "[cyan bold]Switched to GLM[/cyan bold] backend\n\n"
            "[dim]Added GLM env to: .claude/settings.local.json[/dim]\n"
            "[dim]Environment variables have been substituted.[/dim]",
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
