"""MoAI-ADK status command

Project status display:
- Read project information from config.yaml
- Show the number of SPEC documents
- Summarize the Git status

## Skill Invocation Guide (English-Only)

### Related Skills

- **moai-foundation-trust**: For comprehensive TRUST 4-principles verification
  - Trigger: After status shows SPECs exist, to validate code quality
  - Invocation: `Skill("moai-foundation-trust")` to verify all quality gates

- **moai-foundation-git**: For detailed Git workflow information
  - Trigger: When Git status shows "Modified" and you need workflow guidance
  - Invocation: `Skill("moai-foundation-git")` for GitFlow automation details

### When to Invoke Skills in Related Workflows
1. **Before starting new SPEC creation**:
   - Check the SPEC count from status command

2. **After modifications to code/docs**:
   - If status shows "Modified", run `Skill("moai-foundation-git")` for commit strategy
   - Follow up with `Skill("moai-foundation-trust")` to validate code quality

3. **Periodic health checks**:
   - Run status command regularly
   - When SPEC count grows, verify with `Skill("moai-foundation-trust")`
"""

from pathlib import Path

import click
import yaml
from rich.console import Console
from rich.panel import Panel
from rich.table import Table

console = Console()


@click.command()
def status() -> None:
    """Show current project status

    Displays:
    - Project mode (personal/team)
    - Locale setting
    - Number of SPEC documents
    - Git branch and status
    """
    try:
        # Read config.yaml
        config_path = Path.cwd() / ".moai" / "config" / "config.yaml"
        if not config_path.exists():
            console.print("[yellow]⚠ No .moai/config/config.yaml found[/yellow]")
            console.print("[dim]Run [cyan]python -m moai_adk init .[/cyan] to initialize the project[/dim]")
            raise click.Abort()

        with open(config_path) as f:
            config = yaml.safe_load(f) or {}

        # Count SPEC documents
        specs_dir = Path.cwd() / ".moai" / "specs"
        spec_count = len(list(specs_dir.glob("SPEC-*/spec.md"))) if specs_dir.exists() else 0

        # Build the status table
        table = Table(show_header=False, box=None, padding=(0, 2))
        table.add_column("Key", style="cyan")
        table.add_column("Value", style="bold")

        # Read from project section (with legacy fallback)
        project = config.get("project", {})
        table.add_row("Mode", project.get("mode") or config.get("mode", "unknown"))
        table.add_row("Locale", project.get("locale") or config.get("locale", "unknown"))
        table.add_row("SPECs", str(spec_count))

        # Optionally include Git information
        try:
            from git import Repo

            repo = Repo(Path.cwd())
            table.add_row("Branch", repo.active_branch.name)
            table.add_row("Git Status", "Clean" if not repo.is_dirty() else "Modified")
        except Exception:
            pass

        # Render as a panel
        panel = Panel(
            table,
            title="[bold]Project Status[/bold]",
            border_style="cyan",
            expand=False,
        )

        console.print(panel)

    except click.Abort:
        raise
    except Exception as e:
        console.print(f"[red]✗ Failed to get status: {e}[/red]")
        raise
