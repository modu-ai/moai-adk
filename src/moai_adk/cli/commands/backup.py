"""Backup command"""
from pathlib import Path

import click
from rich.console import Console

from moai_adk.core.template.processor import TemplateProcessor

console = Console()


@click.command()
@click.option(
    "--path",
    type=click.Path(exists=True),
    default=".",
    help="Project path (default: current directory)"
)
def backup(path: str) -> None:
    """Create a backup of the current project.

    Includes:
    - .claude/ (entire directory)
    - .moai/ (excluding specs and reports)
    - CLAUDE.md

    Backup location: .moai-backup/YYYYMMDD-HHMMSS/
    """
    try:
        project_path = Path(path).resolve()

        # Verify the project has been initialized
        if not (project_path / ".moai").exists():
            console.print("[yellow]⚠ Project not initialized[/yellow]")
            raise click.Abort()

        # Create the backup
        console.print("[cyan]💾 Creating backup...[/cyan]")
        processor = TemplateProcessor(project_path)
        backup_path = processor.create_backup()

        # Success message
        console.print(f"[green]✓ Backup completed: {backup_path.relative_to(project_path)}[/green]")

        # Show backup contents
        backup_items = list(backup_path.iterdir())
        for item in backup_items:
            if item.is_dir():
                file_count = len(list(item.rglob("*")))
                console.print(f"   ├─ {item.name}/ ({file_count} files)")
            else:
                console.print(f"   └─ {item.name}")

    except Exception as e:
        console.print(f"[red]✗ Backup failed: {e}[/red]")
        raise click.ClickException(str(e)) from e
