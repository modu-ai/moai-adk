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

try:
    from git import Repo
except ImportError:
    Repo = None  # type: ignore[misc,assignment]

# Initialize Rich console for formatted output
console = Console()


def get_manager(
    repo_path: Path | None = None, worktree_root: Path | None = None, project_name: str | None = None
) -> WorktreeManager:  # noqa: E501
    """Get or create a WorktreeManager instance.

    Args:
        repo_path: Path to Git repository. Defaults to current directory.
        worktree_root: Root directory for worktrees. Auto-detects optimal location.
        project_name: Project name for namespace organization. Auto-detected from repo if not provided.

    Returns:
        WorktreeManager instance.
    """
    # 1. Find Git repository if not specified
    if repo_path is None:
        current_path = Path.cwd()
        # Walk up to find .git directory
        while current_path != current_path.parent:
            if (current_path / ".git").exists():
                repo_path = current_path
                break
            current_path = current_path.parent

        # Fallback to current directory if no .git found
        if repo_path is None:
            repo_path = Path.cwd()

    # 2. Auto-detect worktree root if not specified
    if worktree_root is None:
        worktree_root = _detect_worktree_root(repo_path)

    # 3. Auto-detect project_name if not specified
    if project_name is None:
        project_name = repo_path.name

    return WorktreeManager(repo_path=repo_path, worktree_root=worktree_root, project_name=project_name)


def _detect_worktree_root(repo_path: Path) -> Path:
    """Auto-detect the most appropriate worktree root directory.

    Args:
        repo_path: Path to the Git repository.

    Returns:
        Detected worktree root path.
    """
    # Special handling: if we're in a worktree, find the main repo
    main_repo_path = _find_main_repository(repo_path)

    # Strategy 1: Look for existing registry files in common locations
    potential_roots = [
        Path.home() / "moai" / "worktrees",  # MoAI worktrees directory (highest priority)
        Path.home() / "worktrees",  # User's worktrees directory
        Path.home() / "moai" / "worktrees" / main_repo_path.name,  # MoAI project-specific
        main_repo_path.parent / "worktrees",  # Project-level worktrees
        Path("/Users") / Path.home().name / "worktrees",  # macOS user worktrees
        Path.home() / "worktrees" / main_repo_path.name,  # Alternative organization
    ]

    for root in potential_roots:
        registry_path = root / ".moai-worktree-registry.json"
        if registry_path.exists():
            try:
                with open(registry_path, "r") as f:
                    content = f.read().strip()
                    if content and content != "{}":
                        return root
            except Exception:
                pass

    # Strategy 2: Check if there are actual worktrees in standard locations
    for root in potential_roots:
        if root.exists() and any(root.iterdir()):
            # Look for directories that might be worktrees
            for item in root.iterdir():
                if item.is_dir() and (item / ".git").exists():
                    return root

    # Strategy 3: Create a sensible default based on user's system
    # Prefer ~/moai/worktrees for MoAI projects
    default_root = Path.home() / "moai" / "worktrees"
    if default_root.exists() or default_root.parent.exists():
        return default_root

    # Fallback to user worktrees
    return Path.home() / "worktrees"


def _find_main_repository(start_path: Path) -> Path:
    """Find the main Git repository (not a worktree).

    Args:
        start_path: Starting path to search from.

    Returns:
        Path to the main repository.
    """
    current_path = start_path

    # Walk up to find the main repository (non-worktree)
    while current_path != current_path.parent:
        git_path = current_path / ".git"
        if git_path.exists():
            # Check if this is a worktree or main repo
            if git_path.is_file():
                # This is a worktree - read the main repo path
                try:
                    with open(git_path, "r") as f:
                        for line in f:
                            if line.startswith("gitdir:"):
                                gitdir_path = line[8:].strip()
                                main_git_path = Path(current_path / gitdir_path).parent.parent
                                return main_git_path.resolve()
                except Exception:
                    pass
            elif (git_path / "objects").exists():
                # This is the main repository
                return current_path.resolve()

        current_path = current_path.parent

    # Fallback to the original path
    return start_path.resolve()


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
@click.option("--force", "-f", is_flag=True, help="Force creation even if worktree exists")
def new_worktree(
    spec_id: str,
    branch: str | None,
    base: str,
    repo: str | None,
    worktree_root: str | None,
    force: bool,
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
            force=force,
        )

        console.print("[green]✓[/green] Worktree created successfully")
        console.print(f"  SPEC ID:    {info.spec_id}")
        console.print(f"  Path:       {info.path}")
        console.print(f"  Branch:     {info.branch}")
        console.print(f"  Status:     {info.status}")
        console.print()
        console.print("[yellow]Next steps:[/yellow]")
        console.print(f"  moai-worktree go {spec_id}       # Go to this worktree")

    except WorktreeExistsError as e:
        console.print(f"[red]✗[/red] {e}")
        raise click.Abort()
    except GitOperationError as e:
        console.print(f"[red]✗[/red] {e}")
        raise click.Abort()


@worktree.command(name="list")
@click.option(
    "--format",
    type=click.Choice(["table", "json"]),
    default="table",
    help="Output format",
)
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


@worktree.command(name="go")
@click.argument("spec_id")
@click.option("--repo", type=click.Path(), default=None, help="Repository path")
@click.option("--worktree-root", type=click.Path(), default=None, help="Worktree root directory")
def go_worktree(spec_id: str, repo: str | None, worktree_root: str | None) -> None:
    """Go to a worktree (opens new shell).

    Args:
        spec_id: SPEC ID to go to
        repo: Repository path (optional)
        worktree_root: Worktree root directory (optional)
    """
    try:
        repo_path = Path(repo) if repo else Path.cwd()
        wt_root = Path(worktree_root) if worktree_root else None

        manager = get_manager(repo_path, wt_root)
        info = manager.registry.get(spec_id, project_name=manager.project_name)

        if not info:
            console.print(f"[red]✗[/red] Worktree not found: {spec_id}")
            raise click.Abort()

        import os
        import subprocess

        shell = os.environ.get("SHELL", "/bin/bash")
        console.print(f"[green]→[/green] Opening new shell in {info.path}")
        subprocess.call([shell], cwd=str(info.path))

    except Exception as e:
        console.print(f"[red]✗[/red] Error: {e}")
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


@worktree.command(name="sync")
@click.argument("spec_id", required=False)
@click.option("--base", default="main", help="Base branch to sync from")
@click.option("--rebase", is_flag=True, help="Use rebase instead of merge")
@click.option("--ff-only", is_flag=True, help="Only sync if fast-forward is possible")
@click.option("--all", "sync_all", is_flag=True, help="Sync all worktrees")
@click.option("--auto-resolve", is_flag=True, help="Automatically resolve conflicts")
@click.option("--repo", type=click.Path(), default=None, help="Repository path")
@click.option("--worktree-root", type=click.Path(), default=None, help="Worktree root directory")
def sync_worktree(
    spec_id: str | None,
    base: str,
    rebase: bool,
    ff_only: bool,
    sync_all: bool,
    auto_resolve: bool,
    repo: str | None,
    worktree_root: str | None,
) -> None:
    """Sync worktree with base branch.

    Args:
        spec_id: SPEC ID to sync (optional with --all)
        base: Base branch to sync from (default: main)
        rebase: Use rebase instead of merge
        ff_only: Only sync if fast-forward is possible
        sync_all: Sync all worktrees
        auto_resolve: Automatically resolve conflicts
        repo: Repository path (optional)
        worktree_root: Worktree root directory (optional)
    """
    if not spec_id and not sync_all:
        console.print("[red]✗[/red] Either SPEC_ID or --all option is required")
        raise click.Abort()

    if spec_id and sync_all:
        console.print("[red]✗[/red] Cannot use both SPEC_ID and --all option")
        raise click.Abort()

    try:
        repo_path = Path(repo) if repo else Path.cwd()
        wt_root = Path(worktree_root) if worktree_root else None

        manager = get_manager(repo_path, wt_root)

        if sync_all:
            # Sync all worktrees
            worktrees = manager.list()
            if not worktrees:
                console.print("[yellow]No worktrees found to sync[/yellow]")
                return

            console.print(f"[cyan]Syncing {len(worktrees)} worktrees...[/cyan]")
            success_count = 0
            conflict_count = 0

            for info in worktrees:
                try:
                    manager.sync(
                        spec_id=info.spec_id,
                        base_branch=base,
                        rebase=rebase,
                        ff_only=ff_only,
                        auto_resolve=auto_resolve,
                    )
                    sync_method = "rebase" if rebase else ("fast-forward" if ff_only else "merge")
                    console.print(f"[green]✓[/green] {info.spec_id} ({sync_method})")
                    success_count += 1
                except MergeConflictError:
                    if auto_resolve:
                        # Try to auto-resolve conflicts
                        try:
                            worktree_repo = Repo(info.path)
                            conflicted_files = [info.spec_id]  # This will be handled by the method
                            manager.auto_resolve_conflicts(worktree_repo, info.spec_id, conflicted_files)
                            console.print(f"[yellow]![/yellow] {info.spec_id} (auto-resolved)")
                            success_count += 1
                        except Exception as e:
                            console.print(f"[red]✗[/red] {info.spec_id} (auto-resolve failed: {e})")
                            conflict_count += 1
                    else:
                        console.print(f"[red]✗[/red] {info.spec_id} (conflicts)")
                        conflict_count += 1
                except Exception as e:
                    console.print(f"[red]✗[/red] {info.spec_id} (failed: {e})")
                    conflict_count += 1

            console.print()
            console.print(f"[green]Summary:[/green] {success_count} synced, {conflict_count} failed")
        else:
            # Sync single worktree
            manager.sync(
                spec_id=spec_id,
                base_branch=base,
                rebase=rebase,
                ff_only=ff_only,
                auto_resolve=auto_resolve,
            )
            sync_method = "rebase" if rebase else ("fast-forward" if ff_only else "merge")
            console.print(f"[green]✓[/green] Worktree synced: {spec_id} ({sync_method})")

    except WorktreeNotFoundError as e:
        console.print(f"[red]✗[/red] {e}")
        raise click.Abort()
    except GitOperationError as e:
        console.print(f"[red]✗[/red] {e}")
        raise click.Abort()


@worktree.command(name="clean")
@click.option("--merged-only", is_flag=True, help="Only remove merged branch worktrees")
@click.option("--interactive", is_flag=True, help="Interactive cleanup with confirmation prompts")
@click.option("--repo", type=click.Path(), default=None, help="Repository path")
@click.option("--worktree-root", type=click.Path(), default=None, help="Worktree root directory")
def clean_worktrees(merged_only: bool, interactive: bool, repo: str | None, worktree_root: str | None) -> None:
    """Remove worktrees for merged branches.

    Args:
        merged_only: Only remove worktrees for merged branches
        interactive: Interactive cleanup with confirmation prompts
        repo: Repository path (optional)
        worktree_root: Worktree root directory (optional)
    """
    try:
        repo_path = Path(repo) if repo else Path.cwd()
        wt_root = Path(worktree_root) if worktree_root else None

        manager = get_manager(repo_path, wt_root)

        if merged_only:
            # Only clean merged branches (default behavior)
            cleaned = manager.clean_merged()
        elif interactive:
            # Interactive cleanup
            worktrees = manager.list()
            if not worktrees:
                console.print("[yellow]No worktrees found to clean[/yellow]")
                return

            console.print(f"[cyan]Found {len(worktrees)} worktrees:[/cyan]")
            for i, info in enumerate(worktrees, 1):
                console.print(f"  {i}. [cyan]{info.spec_id}[/cyan] ({info.branch}) - {info.status}")

            console.print()
            console.print("[yellow]Select worktrees to remove (comma-separated numbers, or 'all'):[/yellow]")

            try:
                selection = input("> ").strip()
                if selection.lower() == "all":
                    cleaned = [info.spec_id for info in worktrees]
                else:
                    indices = [int(x.strip()) for x in selection.split(",") if x.strip().isdigit()]
                    for idx in indices:
                        if 1 <= idx <= len(worktrees):
                            cleaned.append(worktrees[idx - 1].spec_id)
                        else:
                            console.print(f"[red]✗[/red] Invalid index: {idx}")
                            raise click.Abort()

                if not cleaned:
                    console.print("[yellow]No worktrees selected for cleanup[/yellow]")
                    return

                # Final confirmation
                console.print(f"[yellow]About to remove {len(cleaned)} worktrees: {', '.join(cleaned)}[/yellow]")
                if input("Continue? [y/N]: ").strip().lower() in ["y", "yes"]:
                    for spec_id in cleaned:
                        try:
                            manager.remove(spec_id, force=True)
                        except Exception as e:
                            console.print(f"[red]✗[/red] Failed to remove {spec_id}: {e}")

                console.print("[green]✓[/green] Interactive cleanup completed")

            except (ValueError, KeyboardInterrupt):
                console.print("[yellow]Interactive cleanup cancelled[/yellow]")
                raise click.Abort()
        else:
            # Default: clean all (legacy behavior)
            worktrees = manager.list()
            if not worktrees:
                console.print("[yellow]No worktrees found to clean[/yellow]")
                return

            cleaned = [info.spec_id for info in worktrees]
            console.print(
                "[yellow]Removing all worktrees. "
                "Use --merged-only for merged branches only "
                "or --interactive for selective cleanup.[/yellow]"
            )

            for spec_id in cleaned:
                try:
                    manager.remove(spec_id, force=True)
                except Exception as e:
                    console.print(f"[red]✗[/red] Failed to remove {spec_id}: {e}")
                    cleaned.remove(spec_id)

        if cleaned:
            console.print(f"[green]✓[/green] Cleaned {len(cleaned)} worktree(s)")
            for spec_id in cleaned:
                console.print(f"  - {spec_id}")
        else:
            console.print("[yellow]No worktrees were cleaned[/yellow]")

    except Exception as e:
        console.print(f"[red]✗[/red] Error cleaning worktrees: {e}")
        raise click.Abort()


@worktree.command(name="recover")
@click.option("--repo", type=click.Path(), default=None, help="Repository path")
@click.option("--worktree-root", type=click.Path(), default=None, help="Worktree root directory")
def recover_worktrees(repo: str | None, worktree_root: str | None) -> None:
    """Recover worktree registry from existing directories.

    Scans the worktree root directory for existing worktrees and
    re-registers them in the registry file.

    Args:
        repo: Repository path (optional)
        worktree_root: Worktree root directory (optional)
    """
    try:
        repo_path = Path(repo) if repo else Path.cwd()
        wt_root = Path(worktree_root) if worktree_root else None

        manager = get_manager(repo_path, wt_root)

        console.print(f"[cyan]Scanning {manager.worktree_root} for worktrees...[/cyan]")
        recovered = manager.registry.recover_from_disk()

        if recovered > 0:
            console.print(f"[green]✓[/green] Recovered {recovered} worktree(s)")

            # Show recovered worktrees
            worktrees = manager.list()
            for info in worktrees:
                if info.status == "recovered":
                    console.print(f"  - {info.spec_id} ({info.branch})")
        else:
            console.print("[yellow]No new worktrees found to recover[/yellow]")

    except Exception as e:
        console.print(f"[red]✗[/red] Error recovering worktrees: {e}")
        raise click.Abort()


@worktree.command(name="config")
@click.argument("key", required=False)
@click.argument("value", required=False)
@click.option("--repo", type=click.Path(), default=None, help="Repository path")
@click.option("--worktree-root", type=click.Path(), default=None, help="Worktree root directory")
def config_worktree(key: str | None, value: str | None, repo: str | None, worktree_root: str | None) -> None:
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

        if key is None:
            # No arguments - show all configuration
            console.print("[cyan]Configuration:[/cyan]")
            console.print(f"  root:      {manager.worktree_root}")
            console.print(f"  registry:  {manager.registry.registry_path}")
            console.print()
            console.print("[yellow]Available commands:[/yellow]")
            console.print("  moai-worktree config all           # Show all config")
            console.print("  moai-worktree config root         # Show worktree root")
            console.print("  moai-worktree config registry     # Show registry path")
        elif value is None:
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
