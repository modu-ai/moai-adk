"""CLI commands for Git worktree management."""

from datetime import datetime
from pathlib import Path

import click
from rich.console import Console
from rich.table import Table

from moai_adk.cli.worktree.exceptions import (
    GitOperationError,
    MergeConflictError,
    UncommittedChangesError,
    WorktreeExistsError,
    WorktreeNotFoundError,
)
from moai_adk.cli.worktree.manager import WorktreeManager

# Initialize Rich console for formatted output
console = Console()


def get_manager(repo_path: Path | None = None, worktree_root: Path | None = None) -> WorktreeManager:
    """Get or create a WorktreeManager instance.

    Args:
        repo_path: Path to Git repository. Defaults to current directory.
        worktree_root: Root directory for worktrees. Defaults to ~/worktrees/{project_name}/.

    Returns:
        WorktreeManager instance.
    """
    if repo_path is None:
        repo_path = Path.cwd()

    if worktree_root is None:
        # Default to ~/worktrees/{project_name}/
        project_name = repo_path.name
        worktree_root = Path.home() / "worktrees" / project_name

    return WorktreeManager(repo_path=repo_path, worktree_root=worktree_root)


@click.group()
def worktree() -> None:
    """Manage Git worktrees for parallel SPEC development."""
    pass


@worktree.command(name="new")
@click.argument("spec_id")
@click.option("--branch", "-b", default=None, help="Custom branch name")
@click.option("--base", default="main", help="Base branch to create from")
@click.option("--repo", type=click.Path(), default=None, help="Repository path")
@click.option("--worktree-root", type=click.Path(), default=None, help="Worktree root directory")
def new_worktree(
    spec_id: str,
    branch: str | None,
    base: str,
    repo: str | None,
    worktree_root: str | None,
) -> None:
    """Create a new worktree for a SPEC.

    Args:
        spec_id: SPEC ID (e.g., SPEC-AUTH-001)
        branch: Custom branch name (optional)
        base: Base branch to create from (default: main)
        repo: Repository path (optional)
        worktree_root: Worktree root directory (optional)
    """
    try:
        repo_path = Path(repo) if repo else Path.cwd()
        wt_root = Path(worktree_root) if worktree_root else None

        manager = get_manager(repo_path, wt_root)
        info = manager.create(
            spec_id=spec_id,
            branch_name=branch,
            base_branch=base,
        )

        console.print("[green]✓[/green] Worktree created successfully")
        console.print(f"  SPEC ID:    {info.spec_id}")
        console.print(f"  Path:       {info.path}")
        console.print(f"  Branch:     {info.branch}")
        console.print(f"  Status:     {info.status}")
        console.print()
        console.print("[yellow]Next steps:[/yellow]")
        console.print(f"  moai-worktree switch {spec_id}   # Switch to this worktree")
        console.print(f"  moai-worktree go {spec_id}       # Get cd command")

    except WorktreeExistsError as e:
        console.print(f"[red]✗[/red] {e}")
        raise click.Abort()
    except GitOperationError as e:
        console.print(f"[red]✗[/red] {e}")
        raise click.Abort()


@worktree.command(name="list")
@click.option("--format", type=click.Choice(["table", "json"]), default="table", help="Output format")
@click.option("--repo", type=click.Path(), default=None, help="Repository path")
@click.option("--worktree-root", type=click.Path(), default=None, help="Worktree root directory")
def list_worktrees(format: str, repo: str | None, worktree_root: str | None) -> None:
    """List all active worktrees.

    Args:
        format: Output format (table or json)
        repo: Repository path (optional)
        worktree_root: Worktree root directory (optional)
    """
    try:
        repo_path = Path(repo) if repo else Path.cwd()
        wt_root = Path(worktree_root) if worktree_root else None

        manager = get_manager(repo_path, wt_root)
        worktrees = manager.list()

        if not worktrees:
            console.print("[yellow]No worktrees found[/yellow]")
            return

        if format == "json":
            data = [w.to_dict() for w in worktrees]
            console.print_json(data=data)
        else:  # table
            table = Table(title="Git Worktrees")
            table.add_column("SPEC ID", style="cyan")
            table.add_column("Branch", style="magenta")
            table.add_column("Path", style="green")
            table.add_column("Status", style="yellow")
            table.add_column("Created", style="blue")

            for info in worktrees:
                created = datetime.fromisoformat(info.created_at.replace("Z", "+00:00"))
                table.add_row(
                    info.spec_id,
                    info.branch,
                    str(info.path),
                    info.status,
                    created.strftime("%Y-%m-%d %H:%M:%S"),
                )

            console.print(table)

    except Exception as e:
        console.print(f"[red]✗[/red] Error listing worktrees: {e}")
        raise click.Abort()


@worktree.command(name="switch")
@click.argument("spec_id")
@click.option("--repo", type=click.Path(), default=None, help="Repository path")
@click.option("--worktree-root", type=click.Path(), default=None, help="Worktree root directory")
def switch_worktree(spec_id: str, repo: str | None, worktree_root: str | None) -> None:
    """Switch to a worktree (opens new shell).

    Args:
        spec_id: SPEC ID to switch to
        repo: Repository path (optional)
        worktree_root: Worktree root directory (optional)
    """
    try:
        repo_path = Path(repo) if repo else Path.cwd()
        wt_root = Path(worktree_root) if worktree_root else None

        manager = get_manager(repo_path, wt_root)
        info = manager.registry.get(spec_id)

        if not info:
            console.print(f"[red]✗[/red] Worktree not found: {spec_id}")
            raise click.Abort()

        import os
        import subprocess

        shell = os.environ.get("SHELL", "/bin/bash")
        console.print(f"[green]→[/green] Opening new shell in {info.path}")
        subprocess.call([shell], cwd=str(info.path))

    except Exception as e:
        console.print(f"[red]✗[/red] Error switching worktree: {e}")
        raise click.Abort()


@worktree.command(name="remove")
@click.argument("spec_id")
@click.option("--force", "-f", is_flag=True, help="Force remove with uncommitted changes")
@click.option("--repo", type=click.Path(), default=None, help="Repository path")
@click.option("--worktree-root", type=click.Path(), default=None, help="Worktree root directory")
def remove_worktree(spec_id: str, force: bool, repo: str | None, worktree_root: str | None) -> None:
    """Remove a worktree.

    Args:
        spec_id: SPEC ID to remove
        force: Force removal even with uncommitted changes
        repo: Repository path (optional)
        worktree_root: Worktree root directory (optional)
    """
    try:
        repo_path = Path(repo) if repo else Path.cwd()
        wt_root = Path(worktree_root) if worktree_root else None

        manager = get_manager(repo_path, wt_root)
        manager.remove(spec_id=spec_id, force=force)

        console.print(f"[green]✓[/green] Worktree removed: {spec_id}")

    except WorktreeNotFoundError as e:
        console.print(f"[red]✗[/red] {e}")
        raise click.Abort()
    except UncommittedChangesError as e:
        console.print(f"[red]✗[/red] {e}")
        raise click.Abort()
    except GitOperationError as e:
        console.print(f"[red]✗[/red] {e}")
        raise click.Abort()


@worktree.command(name="status")
@click.option("--repo", type=click.Path(), default=None, help="Repository path")
@click.option("--worktree-root", type=click.Path(), default=None, help="Worktree root directory")
def status_worktrees(repo: str | None, worktree_root: str | None) -> None:
    """Show worktree status and sync registry.

    Args:
        repo: Repository path (optional)
        worktree_root: Worktree root directory (optional)
    """
    try:
        repo_path = Path(repo) if repo else Path.cwd()
        wt_root = Path(worktree_root) if worktree_root else None

        manager = get_manager(repo_path, wt_root)

        # Sync registry with Git
        manager.registry.sync_with_git(manager.repo)

        worktrees = manager.list()

        if not worktrees:
            console.print("[yellow]No worktrees found[/yellow]")
            return

        console.print(f"[cyan]Total worktrees: {len(worktrees)}[/cyan]")
        console.print()

        for info in worktrees:
            status_color = "green" if info.status == "active" else "yellow"
            console.print(f"[{status_color}]{info.spec_id}[/{status_color}]")
            console.print(f"  Branch: {info.branch}")
            console.print(f"  Path:   {info.path}")
            console.print(f"  Status: {info.status}")
            console.print()

    except Exception as e:
        console.print(f"[red]✗[/red] Error getting status: {e}")
        raise click.Abort()


@worktree.command(name="go")
@click.argument("spec_id")
@click.option("--repo", type=click.Path(), default=None, help="Repository path")
@click.option("--worktree-root", type=click.Path(), default=None, help="Worktree root directory")
def go_worktree(spec_id: str, repo: str | None, worktree_root: str | None) -> None:
    """Print cd command for shell eval.

    Usage: eval $(moai-worktree go SPEC-001)

    Args:
        spec_id: SPEC ID to navigate to
        repo: Repository path (optional)
        worktree_root: Worktree root directory (optional)
    """
    try:
        repo_path = Path(repo) if repo else Path.cwd()
        wt_root = Path(worktree_root) if worktree_root else None

        manager = get_manager(repo_path, wt_root)
        info = manager.registry.get(spec_id)

        if not info:
            console.print(f"[red]✗[/red] Worktree not found: {spec_id}")
            raise click.Abort()

        # Print cd command that can be eval'd
        click.echo(f"cd {info.path}")

    except Exception as e:
        console.print(f"[red]✗[/red] Error: {e}")
        raise click.Abort()


@worktree.command(name="sync")
@click.argument("spec_id")
@click.option("--base", default="main", help="Base branch to sync from")
@click.option("--repo", type=click.Path(), default=None, help="Repository path")
@click.option("--worktree-root", type=click.Path(), default=None, help="Worktree root directory")
def sync_worktree(spec_id: str, base: str, repo: str | None, worktree_root: str | None) -> None:
    """Sync worktree with base branch.

    Args:
        spec_id: SPEC ID to sync
        base: Base branch to sync from (default: main)
        repo: Repository path (optional)
        worktree_root: Worktree root directory (optional)
    """
    try:
        repo_path = Path(repo) if repo else Path.cwd()
        wt_root = Path(worktree_root) if worktree_root else None

        manager = get_manager(repo_path, wt_root)
        manager.sync(spec_id=spec_id, base_branch=base)

        console.print(f"[green]✓[/green] Worktree synced: {spec_id}")

    except WorktreeNotFoundError as e:
        console.print(f"[red]✗[/red] {e}")
        raise click.Abort()
    except MergeConflictError as e:
        console.print(f"[red]✗[/red] {e}")
        raise click.Abort()
    except GitOperationError as e:
        console.print(f"[red]✗[/red] {e}")
        raise click.Abort()


@worktree.command(name="clean")
@click.option("--repo", type=click.Path(), default=None, help="Repository path")
@click.option("--worktree-root", type=click.Path(), default=None, help="Worktree root directory")
def clean_worktrees(repo: str | None, worktree_root: str | None) -> None:
    """Remove worktrees for merged branches.

    Args:
        repo: Repository path (optional)
        worktree_root: Worktree root directory (optional)
    """
    try:
        repo_path = Path(repo) if repo else Path.cwd()
        wt_root = Path(worktree_root) if worktree_root else None

        manager = get_manager(repo_path, wt_root)
        cleaned = manager.clean_merged()

        if not cleaned:
            console.print("[yellow]No merged worktrees to clean[/yellow]")
            return

        console.print(f"[green]✓[/green] Cleaned {len(cleaned)} worktree(s)")
        for spec_id in cleaned or []:
            console.print(f"  - {spec_id}")

    except Exception as e:
        console.print(f"[red]✗[/red] Error cleaning worktrees: {e}")
        raise click.Abort()


@worktree.command(name="config")
@click.argument("key")
@click.argument("value", required=False)
@click.option("--repo", type=click.Path(), default=None, help="Repository path")
@click.option("--worktree-root", type=click.Path(), default=None, help="Worktree root directory")
def config_worktree(key: str, value: str | None, repo: str | None, worktree_root: str | None) -> None:
    """Get or set worktree configuration.

    Supported configuration keys:
    - root: Worktree root directory
    - auto-sync: Enable automatic sync on worktree creation

    Args:
        key: Configuration key
        value: Configuration value (optional for get)
        repo: Repository path (optional)
        worktree_root: Worktree root directory (optional)
    """
    try:
        repo_path = Path(repo) if repo else Path.cwd()
        wt_root = Path(worktree_root) if worktree_root else None

        manager = get_manager(repo_path, wt_root)

        if value is None:
            # Get configuration
            if key == "root":
                console.print(f"[cyan]Worktree root:[/cyan] {manager.worktree_root}")
            elif key == "registry":
                console.print(f"[cyan]Registry path:[/cyan] {manager.registry.registry_path}")
            elif key == "all":
                console.print("[cyan]Configuration:[/cyan]")
                console.print(f"  root:      {manager.worktree_root}")
                console.print(f"  registry:  {manager.registry.registry_path}")
            else:
                console.print(f"[yellow]Unknown config key: {key}[/yellow]")
                console.print("[yellow]Available keys: root, registry, all[/yellow]")
        else:
            # Set configuration (limited support)
            if key == "root":
                console.print("[yellow]Use --worktree-root option to change root directory[/yellow]")
            else:
                console.print(f"[yellow]Cannot set configuration key: {key}[/yellow]")

    except Exception as e:
        console.print(f"[red]✗[/red] Error: {e}")
        raise click.Abort()
