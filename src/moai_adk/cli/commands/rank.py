"""MoAI Rank CLI commands.

Commands for interacting with the MoAI Rank leaderboard service:
- register: Connect GitHub account via OAuth
- status: Show current rank, statistics, and hook status
- logout: Remove stored credentials
- exclude: Exclude a project from session tracking (use --list to view)
- include: Re-include a previously excluded project

For the full leaderboard and detailed statistics, visit https://rank.mo.ai.kr
"""

import sys

import click
from rich.columns import Columns
from rich.console import Console
from rich.panel import Panel

# Force UTF-8 encoding for Windows compatibility
if sys.platform == "win32":
    console = Console(force_terminal=True, legacy_windows=False)
else:
    console = Console()


def format_tokens(tokens: int) -> str:
    """Format token count with K/M suffix."""
    if tokens >= 1_000_000:
        return f"{tokens / 1_000_000:.1f}M"
    elif tokens >= 1_000:
        return f"{tokens / 1_000:.1f}K"
    return str(tokens)


def format_rank_position(position: int, total: int) -> str:
    """Format rank position with total participants."""
    if position <= 3:
        medals = {1: "[gold1]1st[/gold1]", 2: "[grey70]2nd[/grey70]", 3: "[orange3]3rd[/orange3]"}
        return f"{medals[position]} / {total}"
    return f"#{position} / {total}"


@click.group()
def rank() -> None:
    """MoAI Rank - Token usage leaderboard.

    Track your Claude Code token usage and compete on the leaderboard.
    Visit https://rank.mo.ai.kr for the web dashboard.
    """
    pass


@rank.command()
@click.option(
    "--no-sync",
    is_flag=True,
    help="Skip syncing existing sessions after registration",
)
@click.option(
    "--background-sync",
    "-b",
    is_flag=True,
    help="Sync existing sessions in background after registration",
)
def register(no_sync: bool, background_sync: bool) -> None:
    """Register with MoAI Rank via GitHub OAuth.

    Opens your browser to authorize with GitHub.
    Your API key will be stored securely in ~/.moai/rank/credentials.json
    """
    import subprocess
    import sys

    from moai_adk.rank.auth import OAuthHandler
    from moai_adk.rank.config import RankConfig

    # Check if already registered
    if RankConfig.has_credentials():
        creds = RankConfig.load_credentials()
        if creds:
            console.print(f"[yellow]Already registered as [bold]{creds.username}[/bold][/yellow]")
            if not click.confirm("Do you want to re-register?"):
                return

    console.print()
    console.print(
        Panel(
            "[cyan]MoAI Rank Registration[/cyan]\n\n"
            "This will open your browser to authorize with GitHub.\n"
            "After authorization, your API key will be stored securely.",
            title="[bold]Registration[/bold]",
            border_style="cyan",
        )
    )
    console.print()

    with console.status("[bold cyan]Starting OAuth flow...[/bold cyan]"):
        handler = OAuthHandler()

    def on_success(creds):
        from moai_adk.rank.hook import install_hook

        console.print()
        console.print(
            Panel(
                f"[green]Successfully registered as [bold]{creds.username}[/bold][/green]\n\n"
                f"API Key: [dim]{creds.api_key[:20]}...[/dim]\n"
                f"Stored in: [dim]~/.moai/rank/credentials.json[/dim]",
                title="[bold green]Registration Complete[/bold green]",
                border_style="green",
            )
        )

        # Install global hook automatically
        if install_hook():
            console.print()
            console.print(
                Panel(
                    "[cyan]Session tracking hook installed globally.[/cyan]\n\n"
                    "Your Claude Code sessions will be automatically tracked.\n"
                    "Hook location: [dim]~/.claude/hooks/moai/session_end__rank_submit.py[/dim]\n\n"
                    "[dim]To exclude specific projects:[/dim]\n"
                    "  [cyan]moai rank exclude /path/to/project[/cyan]",
                    title="[bold]Global Hook Installed[/bold]",
                    border_style="cyan",
                )
            )

            # Handle sync options
            if no_sync:
                console.print()
                console.print(
                    "[dim]Sync skipped. Run [cyan]moai rank sync[/cyan] later to sync existing sessions.[/dim]"
                )
            elif background_sync:
                console.print()
                console.print("[cyan]Starting background sync...[/cyan]")
                subprocess.Popen(
                    [sys.executable, "-m", "moai_adk.cli.commands.rank", "sync", "--background"],
                    start_new_session=True,
                )
            else:
                console.print()
                console.print(
                    "[dim]To sync existing sessions, run:[/dim]\n"
                    "  [cyan]moai rank sync[/cyan]           # Foreground\n"
                    "  [cyan]moai rank sync --background[/cyan]  # Background (~3min)"
                )
        else:
            console.print("[yellow]Warning: Failed to install session tracking hook.[/yellow]")

        console.print()
        console.print("[dim]Run [cyan]moai rank status[/cyan] to see your stats.[/dim]")

    def on_error(error):
        console.print(f"\n[red]Registration failed: {error}[/red]")

    console.print("[cyan]Opening browser for GitHub authorization...[/cyan]")
    console.print("[dim]Waiting for authorization (timeout: 5 minutes)...[/dim]")
    console.print()

    handler.start_oauth_flow(on_success=on_success, on_error=on_error, timeout=300)


def _get_rank_medal(position: int) -> str:
    """Get medal emoji for rank position."""
    medals = {1: "[gold1]1st[/gold1]", 2: "[grey70]2nd[/grey70]", 3: "[orange3]3rd[/orange3]"}
    return medals.get(position, f"#{position}")


def _create_progress_bar(value: int, total: int, width: int = 20) -> str:
    """Create a simple text-based progress bar."""
    if total == 0:
        return "[dim]" + "-" * width + "[/dim]"
    ratio = min(value / total, 1.0)
    filled = int(width * ratio)
    return f"[cyan]{'█' * filled}[/cyan][dim]{'░' * (width - filled)}[/dim]"


@rank.command()
def status() -> None:
    """Show your current rank and statistics.

    Displays your ranking position across different time periods,
    cumulative token usage statistics, and hook installation status.

    For the full leaderboard, visit https://rank.mo.ai.kr
    """

    from moai_adk.rank.client import AuthenticationError, RankClient, RankClientError
    from moai_adk.rank.config import RankConfig
    from moai_adk.rank.hook import is_hook_installed

    if not RankConfig.has_credentials():
        console.print("[yellow]Not registered with MoAI Rank.[/yellow]")
        console.print("[dim]Run [cyan]moai rank register[/cyan] to connect your account.[/dim]")
        return

    try:
        with console.status("[bold cyan]Fetching your rank...[/bold cyan]"):
            client = RankClient()
            user_rank = client.get_user_rank()

        console.print()

        # === Header Panel with main stats ===
        weekly = user_rank.weekly
        if weekly:
            rank_text = _get_rank_medal(weekly.position)
            header_content = (
                f"[bold cyan]{user_rank.username}[/bold cyan]\n\n"
                f"[dim]Weekly Rank[/dim]  {rank_text} [dim]/ {weekly.total_participants}[/dim]\n"
                f"[dim]Score[/dim]        [bold]{weekly.composite_score:,.0f}[/bold]"
            )
        else:
            header_content = f"[bold cyan]{user_rank.username}[/bold cyan]\n\n[dim]No ranking data[/dim]"

        console.print(Panel(header_content, title="[bold]MoAI Rank[/bold]", border_style="cyan"))

        # === Rankings Grid (2x2) ===
        rank_panels = []
        periods = [
            ("Daily", user_rank.daily, "yellow"),
            ("Weekly", user_rank.weekly, "cyan"),
            ("Monthly", user_rank.monthly, "green"),
            ("All Time", user_rank.all_time, "magenta"),
        ]

        for period_name, rank_info, color in periods:
            if rank_info:
                pos = _get_rank_medal(rank_info.position)
                content = (
                    f"{pos} [dim]/ {rank_info.total_participants}[/dim]\n[dim]{rank_info.composite_score:,.0f}[/dim]"
                )
            else:
                content = "[dim]-[/dim]"
            rank_panels.append(Panel(content, title=f"[{color}]{period_name}[/{color}]", border_style=color, width=20))

        console.print(Columns(rank_panels, equal=True, expand=True))

        # === Token Statistics with Progress Bars ===
        total = user_rank.total_tokens
        input_pct = (user_rank.input_tokens / total * 100) if total > 0 else 0
        output_pct = (user_rank.output_tokens / total * 100) if total > 0 else 0

        stats_content = (
            f"[bold]{format_tokens(total)}[/bold] [dim]total tokens[/dim]\n\n"
            f"[dim]Input[/dim]  {_create_progress_bar(user_rank.input_tokens, total, 15)} "
            f"[bold]{format_tokens(user_rank.input_tokens)}[/bold] [dim]({input_pct:.0f}%)[/dim]\n"
            f"[dim]Output[/dim] {_create_progress_bar(user_rank.output_tokens, total, 15)} "
            f"[bold]{format_tokens(user_rank.output_tokens)}[/bold] [dim]({output_pct:.0f}%)[/dim]\n\n"
            f"[dim]Sessions:[/dim] [bold]{user_rank.total_sessions}[/bold]"
        )

        console.print(Panel(stats_content, title="[bold]Token Usage[/bold]", border_style="green"))

        # === Footer: Hook + Links ===
        hook_icon = "[green]●[/green]" if is_hook_installed() else "[yellow]○[/yellow]"
        hook_text = "Installed" if is_hook_installed() else "Not installed"

        footer = f"{hook_icon} Hook: {hook_text}  [dim]|[/dim]  [cyan]https://rank.mo.ai.kr[/cyan]"
        console.print()
        console.print(footer)

        if not is_hook_installed():
            console.print("[dim]Run [cyan]moai rank register[/cyan] to install the hook.[/dim]")

    except AuthenticationError as e:
        console.print(f"[red]Authentication failed: {e}[/red]")
        console.print("[dim]Your API key may be invalid. Try [cyan]moai rank register[/cyan] again.[/dim]")
    except RankClientError as e:
        console.print(f"[red]Failed to fetch status: {e}[/red]")


@rank.command()
def logout() -> None:
    """Remove stored MoAI Rank credentials.

    This will delete your API key from ~/.moai/rank/credentials.json
    """
    from moai_adk.rank.config import RankConfig

    if not RankConfig.has_credentials():
        console.print("[yellow]No credentials stored.[/yellow]")
        return

    creds = RankConfig.load_credentials()
    username = creds.username if creds else "unknown"

    if click.confirm(f"Remove credentials for {username}?"):
        RankConfig.delete_credentials()
        console.print("[green]Credentials removed successfully.[/]")
    else:
        console.print("[dim]Cancelled.[/dim]")


@rank.command()
@click.option("--background", "-b", is_flag=True, help="Run sync in background")
def sync(background: bool) -> None:
    """Sync all existing Claude Code sessions to MoAI Rank.

    Scans ~/.claude/projects/ for all session transcripts and submits them
    to the rank server.

    Examples:
        moai rank sync              # Sync in foreground
        moai rank sync --background  # Sync in background (detached)
    """
    import os
    import subprocess
    import sys

    from moai_adk.rank.config import RankConfig

    if not RankConfig.has_credentials():
        console.print("[yellow]Not registered with MoAI Rank.[/yellow]")
        console.print("[dim]Run [cyan]moai rank register[/cyan] first.[/dim]")
        return

    if background:
        # Run in background
        console.print("[cyan]Starting background sync...[/cyan]")

        python_exe = sys.executable

        # Create the sync script
        sync_script = f'''
import sys
sys.path.insert(0, "{os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))}")
from moai_adk.rank.hook import sync_all_sessions
from rich.console import Console

sync_all_sessions(Console())
'''

        # Write to temp file
        import tempfile

        with tempfile.NamedTemporaryFile(mode="w", suffix=".py", delete=False) as f:
            f.write(sync_script)
            script_path = f.name

        # Run in background
        log_file = RankConfig.CONFIG_DIR / "sync.log"
        null_dev = "/dev/null" if os.name != "nt" else "NUL"

        try:
            subprocess.Popen(
                [python_exe, script_path],
                stdout=(open(log_file, "a") if log_file.parent.exists() else open(null_dev, "w")),
                stderr=subprocess.STDOUT,
                start_new_session=True,
            )
            console.print(f"[green]Background sync started[/green]\n[dim]Log file: [cyan]{log_file}[/cyan][/dim]")
            console.print("[dim]Check sync status with: [cyan]tail -f ~/.moai/rank/sync.log[/cyan][/dim]")
        finally:
            os.unlink(script_path)
    else:
        from moai_adk.rank.hook import sync_all_sessions

        sync_all_sessions(console)


@rank.command()
@click.option("--list", "-l", "show_list", is_flag=True, help="List all excluded projects")
@click.argument("path", required=False)
def exclude(path: str | None, show_list: bool) -> None:
    """Exclude a project from session tracking.

    Adds the specified project path (or current directory) to the exclusion list.
    Sessions from excluded projects will not be submitted to MoAI Rank.

    Examples:
        moai rank exclude                    # Exclude current directory
        moai rank exclude /path/to/project   # Exclude specific project
        moai rank exclude "*/private/*"      # Exclude with wildcard pattern
        moai rank exclude --list             # List all excluded projects
    """
    from pathlib import Path as PathLib

    from moai_adk.rank.hook import add_project_exclusion, load_rank_config

    # Handle --list option
    if show_list:
        config = load_rank_config()
        exclusions = config.get("exclude_projects", [])

        if not exclusions:
            console.print("[dim]No projects are excluded from tracking.[/dim]")
            console.print("[dim]Use [cyan]moai rank exclude <path>[/cyan] to exclude a project.[/dim]")
            return

        console.print()
        console.print("[bold]Excluded Projects:[/bold]")
        for exc_path in exclusions:
            console.print(f"  [dim]•[/dim] {exc_path}")
        console.print()
        console.print(f"[dim]Total: {len(exclusions)} project(s) excluded[/dim]")
        return

    # Use current directory if no path specified
    if path is None:
        path = str(PathLib.cwd())

    if add_project_exclusion(path):
        console.print(f"[green]Excluded: [bold]{path}[/bold][/green]")
        console.print("[dim]Sessions from this project will not be tracked.[/dim]")

        # Show current exclusions
        config = load_rank_config()
        exclusions = config.get("exclude_projects", [])
        if len(exclusions) > 1:
            console.print(f"\n[dim]Total excluded projects: {len(exclusions)}[/dim]")
    else:
        console.print(f"[red]Failed to exclude: {path}[/red]")


@rank.command()
@click.argument("path", required=False)
def include(path: str | None) -> None:
    """Re-include a previously excluded project.

    Removes the specified project path (or current directory) from the exclusion list.
    Sessions from this project will be submitted to MoAI Rank again.

    Examples:
        moai rank include                    # Include current directory
        moai rank include /path/to/project   # Include specific project
    """
    from pathlib import Path as PathLib

    from moai_adk.rank.hook import load_rank_config, remove_project_exclusion

    # Use current directory if no path specified
    if path is None:
        path = str(PathLib.cwd())

    if remove_project_exclusion(path):
        console.print(f"[green]Included: [bold]{path}[/bold][/green]")
        console.print("[dim]Sessions from this project will now be tracked.[/dim]")

        # Show remaining exclusions
        config = load_rank_config()
        exclusions = config.get("exclude_projects", [])
        if exclusions:
            console.print(f"\n[dim]Remaining excluded projects: {len(exclusions)}[/dim]")
    else:
        console.print(f"[red]Failed to include: {path}[/red]")
