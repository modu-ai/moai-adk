"""MoAI Rank CLI commands.

Commands for interacting with the MoAI Rank leaderboard service:
- register: Connect GitHub account via OAuth
- status: Show current rank and statistics
- leaderboard: Display the top users
- logout: Remove stored credentials
"""

import click
from rich.console import Console
from rich.panel import Panel
from rich.table import Table
from rich.text import Text

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
def register() -> None:
    """Register with MoAI Rank via GitHub OAuth.

    Opens your browser to authorize with GitHub.
    Your API key will be stored securely in ~/.moai/rank/credentials.json
    """
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
        console.print()
        console.print("[dim]Run [cyan]moai rank status[/cyan] to see your stats.[/dim]")

    def on_error(error):
        console.print(f"\n[red]Registration failed: {error}[/red]")

    console.print("[cyan]Opening browser for GitHub authorization...[/cyan]")
    console.print("[dim]Waiting for authorization (timeout: 5 minutes)...[/dim]")
    console.print()

    handler.start_oauth_flow(on_success=on_success, on_error=on_error, timeout=300)


@rank.command()
def status() -> None:
    """Show your current rank and statistics.

    Displays your ranking position across different time periods
    and your cumulative token usage statistics.
    """
    from moai_adk.rank.client import AuthenticationError, RankClient, RankClientError
    from moai_adk.rank.config import RankConfig

    if not RankConfig.has_credentials():
        console.print("[yellow]Not registered with MoAI Rank.[/yellow]")
        console.print("[dim]Run [cyan]moai rank register[/cyan] to connect your account.[/dim]")
        return

    try:
        with console.status("[bold cyan]Fetching your rank...[/bold cyan]"):
            client = RankClient()
            user_rank = client.get_user_rank()

        # Build status display
        console.print()

        # Header with username
        header = Text()
        header.append("MoAI Rank Status: ", style="bold")
        header.append(user_rank.username, style="cyan bold")
        console.print(Panel(header, border_style="cyan"))

        # Rankings table
        rankings_table = Table(show_header=True, header_style="bold cyan", box=None)
        rankings_table.add_column("Period", style="dim")
        rankings_table.add_column("Rank", justify="right")
        rankings_table.add_column("Score", justify="right")

        periods = [
            ("Daily", user_rank.daily),
            ("Weekly", user_rank.weekly),
            ("Monthly", user_rank.monthly),
            ("All Time", user_rank.all_time),
        ]

        for period_name, rank_info in periods:
            if rank_info:
                rank_display = format_rank_position(rank_info.position, rank_info.total_participants)
                score = f"{rank_info.composite_score:,.0f}"
            else:
                rank_display = "[dim]-[/dim]"
                score = "[dim]-[/dim]"
            rankings_table.add_row(period_name, rank_display, score)

        console.print(Panel(rankings_table, title="[bold]Rankings[/bold]", border_style="blue"))

        # Statistics table
        stats_table = Table(show_header=False, box=None, padding=(0, 2))
        stats_table.add_column("Stat", style="dim")
        stats_table.add_column("Value", style="bold", justify="right")

        stats_table.add_row("Total Tokens", format_tokens(user_rank.total_tokens))
        stats_table.add_row("Input Tokens", format_tokens(user_rank.input_tokens))
        stats_table.add_row("Output Tokens", format_tokens(user_rank.output_tokens))
        stats_table.add_row("Sessions", str(user_rank.total_sessions))

        console.print(Panel(stats_table, title="[bold]Statistics[/bold]", border_style="green"))

        console.print()
        console.print(f"[dim]Last updated: {user_rank.last_updated}[/dim]")

    except AuthenticationError as e:
        console.print(f"[red]Authentication failed: {e}[/red]")
        console.print("[dim]Your API key may be invalid. Try [cyan]moai rank register[/cyan] again.[/dim]")
    except RankClientError as e:
        console.print(f"[red]Failed to fetch status: {e}[/red]")


@rank.command()
@click.option(
    "--period",
    "-p",
    type=click.Choice(["daily", "weekly", "monthly", "all_time"]),
    default="weekly",
    help="Leaderboard period (default: weekly)",
)
@click.option(
    "--limit",
    "-n",
    type=int,
    default=10,
    help="Number of entries to show (default: 10, max: 100)",
)
def leaderboard(period: str, limit: int) -> None:
    """Display the MoAI Rank leaderboard.

    Shows the top users ranked by composite score for the selected period.
    """
    from moai_adk.rank.client import RankClient, RankClientError

    limit = min(max(1, limit), 100)

    try:
        with console.status(f"[bold cyan]Fetching {period} leaderboard...[/bold cyan]"):
            client = RankClient()
            entries = client.get_leaderboard(period=period, limit=limit)

        if not entries:
            console.print(f"[yellow]No entries found for {period} leaderboard.[/yellow]")
            return

        console.print()

        # Build leaderboard table
        table = Table(
            title=f"[bold cyan]MoAI Rank - {period.replace('_', ' ').title()} Leaderboard[/bold cyan]",
            show_header=True,
            header_style="bold",
        )

        table.add_column("#", style="dim", width=4)
        table.add_column("User", style="cyan")
        table.add_column("Score", justify="right")
        table.add_column("Tokens", justify="right")
        table.add_column("Sessions", justify="right")

        for entry in entries:
            # Rank with medal for top 3
            if entry.rank == 1:
                rank_str = "[gold1]1[/gold1]"
            elif entry.rank == 2:
                rank_str = "[grey70]2[/grey70]"
            elif entry.rank == 3:
                rank_str = "[orange3]3[/orange3]"
            else:
                rank_str = str(entry.rank)

            # Username (anonymized if private)
            username = entry.username if not entry.is_private else f"[dim]{entry.username}[/dim]"

            table.add_row(
                rank_str,
                username,
                f"{entry.composite_score:,.0f}",
                format_tokens(entry.total_tokens),
                str(entry.session_count),
            )

        console.print(table)
        console.print()
        console.print("[dim]Visit [cyan]https://rank.mo.ai.kr[/cyan] for the full leaderboard.[/dim]")

    except RankClientError as e:
        console.print(f"[red]Failed to fetch leaderboard: {e}[/red]")


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
        console.print("[green]Credentials removed successfully.[/green]")
    else:
        console.print("[dim]Cancelled.[/dim]")


@rank.command()
def verify() -> None:
    """Verify your API key is valid.

    Makes a test request to verify your stored credentials work.
    """
    from moai_adk.rank.auth import verify_api_key
    from moai_adk.rank.config import RankConfig

    if not RankConfig.has_credentials():
        console.print("[yellow]No credentials stored.[/yellow]")
        console.print("[dim]Run [cyan]moai rank register[/cyan] to connect your account.[/dim]")
        return

    creds = RankConfig.load_credentials()
    if not creds:
        console.print("[red]Failed to load credentials.[/red]")
        return

    with console.status("[bold cyan]Verifying API key...[/bold cyan]"):
        is_valid = verify_api_key(creds.api_key)

    if is_valid:
        console.print(f"[green]API key for [bold]{creds.username}[/bold] is valid.[/green]")
    else:
        console.print("[red]API key is invalid or expired.[/red]")
        console.print("[dim]Run [cyan]moai rank register[/cyan] to get a new key.[/dim]")
