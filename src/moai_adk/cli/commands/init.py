# @CODE:CLI-001 | @CODE:INIT-003:CLI
# SPEC: SPEC-CLI-001.md, SPEC-INIT-003.md
# TEST: tests/unit/test_cli_commands.py, tests/unit/test_init_reinit.py
"""MoAI-ADK init command

Project initialization command (interactive/non-interactive):
- Interactive Mode: Ask user for project settings
- Non-Interactive Mode: Use defaults or CLI options

## Skill Invocation Guide (English-Only)

### Related Skills
- **moai-foundation-langs**: For language detection and stack configuration
  - Trigger: When language parameter is not specified (auto-detection)
  - Invocation: Called implicitly during project initialization for language matrix detection

### When to Invoke Skills in Related Workflows
1. **After project initialization**:
   - Run `Skill("moai-foundation-trust")` to verify project structure and toolchain
   - Run `Skill("moai-foundation-langs")` to validate detected language stack

2. **Before first SPEC creation**:
   - Use `Skill("moai-alfred-language-detection")` to confirm language selection

3. **Project reinitialization** (`--force`):
   - Skills automatically adapt to new project structure
   - No manual intervention required
"""

import json
import platform
from pathlib import Path
from typing import Sequence

import click
from rich.console import Console
from rich.progress import BarColumn, Progress, SpinnerColumn, TaskID, TextColumn

from moai_adk import __version__
from moai_adk.cli.prompts import prompt_project_setup
from moai_adk.core.project.initializer import ProjectInitializer
from moai_adk.utils.banner import print_banner, print_welcome_message

console = Console()


def create_progress_callback(progress: Progress, task_ids: Sequence[TaskID]):
    """Create progress callback

    Args:
        progress: Rich Progress object
        task_ids: List of task IDs (one per phase)

    Returns:
        Progress callback function
    """

    def callback(message: str, current: int, total: int) -> None:
        """Update progress

        Args:
            message: Progress message
            current: Current phase (1-based)
            total: Total phases
        """
        # Complete current phase (1-based index → 0-based)
        if 1 <= current <= len(task_ids):
            progress.update(task_ids[current - 1], completed=1, description=message)

    return callback


@click.command()
@click.argument("path", type=click.Path(), default=".")
@click.option(
    "--non-interactive",
    "-y",
    is_flag=True,
    help="Non-interactive mode (use defaults)",
)
@click.option(
    "--mode",
    type=click.Choice(["personal", "team"]),
    default="personal",
    help="Project mode",
)
@click.option(
    "--locale",
    type=click.Choice(["ko", "en"]),
    default=None,
    help="Preferred language (default: en)",
)
@click.option(
    "--language",
    type=str,
    default=None,
    help="Programming language (auto-detect if not specified)",
)
@click.option(
    "--force",
    is_flag=True,
    help="Force reinitialize without confirmation",
)
def init(
    path: str,
    non_interactive: bool,
    mode: str,
    locale: str,
    language: str | None,
    force: bool,
) -> None:
    """Initialize a new MoAI-ADK project

    Args:
        path: Project directory path (default: current directory)
        non_interactive: Skip prompts and use defaults
        mode: Project mode (personal/team)
        locale: Preferred language (ko/en). When omitted, defaults to en.
        language: Programming language
        force: Force reinitialize without confirmation
    """
    try:
        # 1. Print banner
        print_banner(__version__)

        # 2. Check current directory mode
        is_current_dir = path == "."
        project_path = Path(path).resolve()

        # 3. Interactive vs Non-Interactive
        if non_interactive:
            # Non-Interactive Mode
            console.print(
                f"\n[cyan]🚀 Initializing project at {project_path}...[/cyan]\n"
            )
            project_name = project_path.name if is_current_dir else path
            locale = locale or "en"
            # Language detection happens in /alfred:0-project, so default to None here
            # This will become "generic" internally, but Summary will show more helpful message
            if not language:
                language = None
        else:
            # Interactive Mode
            print_welcome_message()

            # Interactive prompt
            answers = prompt_project_setup(
                project_name=None if is_current_dir else path,
                is_current_dir=is_current_dir,
                project_path=project_path,
                initial_locale=locale,
            )

            # Override with prompt answers
            mode = answers["mode"]
            locale = answers["locale"]
            language = answers["language"]
            project_name = answers["project_name"]

            console.print("\n[cyan]🚀 Starting installation...[/cyan]\n")

            if locale is None:
                locale = answers["locale"]

        # 4. Check for reinitialization (SPEC-INIT-003 v0.3.0) - DEFAULT TO FORCE MODE
        initializer = ProjectInitializer(project_path)

        if initializer.is_initialized():
            # Always reinitialize without confirmation (force mode by default)
            if non_interactive:
                console.print("\n[green]🔄 Reinitializing project (force mode)...[/green]\n")
            else:
                # Interactive mode: Simple notification
                console.print("\n[cyan]🔄 Reinitializing project...[/cyan]")
                console.print("   Backup will be created at .moai-backups/backup/\n")

        # 5. Initialize project (Progress Bar with 5 phases)
        # Always allow reinit (force mode by default)
        is_reinit = initializer.is_initialized()

        # Reinit mode: set config.json optimized to false (v0.3.1+)
        if is_reinit:
            # Migration: Remove old hook files (Issue #163)
            old_hook_files = [
                ".claude/hooks/alfred/session_start__startup.py",  # v0.8.0 deprecated
            ]
            for old_file in old_hook_files:
                old_path = project_path / old_file
                if old_path.exists():
                    try:
                        old_path.unlink()  # Remove old file
                    except Exception:
                        pass  # Ignore removal failures

            config_path = project_path / ".moai" / "config.json"
            if config_path.exists():
                try:
                    with open(config_path, "r", encoding="utf-8") as f:
                        config_data = json.load(f)

                    # Update version and optimization flags
                    if "moai" not in config_data:
                        config_data["moai"] = {}
                    config_data["moai"]["version"] = __version__

                    if "project" not in config_data:
                        config_data["project"] = {}
                    config_data["project"]["optimized"] = False

                    with open(config_path, "w", encoding="utf-8") as f:
                        json.dump(config_data, f, indent=2, ensure_ascii=False)
                except Exception:
                    # Ignore read/write failures; config.json is regenerated during initialization
                    pass

        with Progress(
            SpinnerColumn(),
            TextColumn("[progress.description]{task.description}"),
            BarColumn(),
            TextColumn("[progress.percentage]{task.percentage:>3.0f}%"),
            console=console,
        ) as progress:
            # Create 5 phase tasks
            phase_names = [
                "Phase 1: Preparation and backup...",
                "Phase 2: Creating directory structure...",
                "Phase 3: Installing resources...",
                "Phase 4: Generating configurations...",
                "Phase 5: Validation and finalization...",
            ]
            task_ids = [progress.add_task(name, total=1) for name in phase_names]
            callback = create_progress_callback(progress, task_ids)

            result = initializer.initialize(
                mode=mode,
                locale=locale,
                language=language,
                backup_enabled=True,
                progress_callback=callback,
                reinit=True,  # Always allow reinit (force mode by default)
            )

        # 6. Output results
        if result.success:
            separator = "[dim]" + ("─" * 60) + "[/dim]"
            console.print(
                "\n[green bold]✅ Initialization Completed Successfully![/green bold]"
            )
            console.print(separator)
            console.print("\n[cyan]📊 Summary:[/cyan]")
            console.print(f"  [dim]📁 Location:[/dim]  {result.project_path}")
            # Show language more clearly - "generic" means auto-detect
            language_display = (
                "Auto-detect (use /alfred:0-project)"
                if result.language == "generic"
                else result.language
            )
            console.print(f"  [dim]🌐 Language:[/dim]  {language_display}")
            console.print(f"  [dim]🔧 Mode:[/dim]      {result.mode}")
            console.print(
                f"  [dim]🌍 Locale:[/dim]    {result.locale}"
            )
            console.print(
                f"  [dim]📄 Files:[/dim]     {len(result.created_files)} created"
            )
            console.print(f"  [dim]⏱️  Duration:[/dim]  {result.duration}ms")

            # Show backup info if reinitialized
            if is_reinit:
                backup_dir = project_path / ".moai-backups"
                if backup_dir.exists():
                    latest_backup = max(backup_dir.iterdir(), key=lambda p: p.stat().st_mtime)
                    console.print(f"  [dim]💾 Backup:[/dim]    {latest_backup.name}/")

            console.print(f"\n{separator}")

            # Show config merge notice if reinitialized
            if is_reinit:
                console.print("\n[yellow]⚠️  Configuration Notice:[/yellow]")
                console.print("  All template files have been [bold]force overwritten[/bold]")
                console.print("  Previous files are backed up in [cyan].moai-backups/backup/[/cyan]")
                console.print("\n  [cyan]To merge your previous config:[/cyan]")
                console.print("  Run [bold]/alfred:0-project[/bold] command in Claude Code")
                console.print("  It will merge backup config when [dim]optimized=false[/dim]\n")

            console.print("\n[cyan]🚀 Next Steps:[/cyan]")
            if not is_current_dir:
                console.print(
                    f"  [blue]1.[/blue] Run [bold]cd {project_name}[/bold] to enter the project"
                )
                console.print(
                    "  [blue]2.[/blue] Run [bold]/alfred:0-project[/bold] in Claude Code for full setup"
                )
                console.print(
                    "     (Configure: mode, language, report generation, etc.)\n"
                )
            else:
                console.print(
                    "  [blue]1.[/blue] Run [bold]/alfred:0-project[/bold] in Claude Code for full setup"
                )
                console.print(
                    "     (Configure: mode, language, report generation, etc.)\n"
                )
        else:
            console.print("\n[red bold]❌ Initialization Failed![/red bold]")
            if result.errors:
                console.print("\n[red]Errors:[/red]")
                for error in result.errors:
                    console.print(f"  [red]•[/red] {error}")
            console.print()
            raise click.ClickException("Installation failed")

    except KeyboardInterrupt:
        console.print("\n\n[yellow]⚠ Initialization cancelled by user[/yellow]\n")
        raise click.Abort()
    except FileExistsError as e:
        console.print("\n[yellow]⚠ Project already initialized[/yellow]")
        console.print("[dim]  Use 'python -m moai_adk status' to check configuration[/dim]\n")
        raise click.Abort() from e
    except Exception as e:
        console.print(f"\n[red]✗ Initialization failed: {e}[/red]\n")
        raise click.ClickException(str(e)) from e
    finally:
        # Explicitly flush output buffer
        console.file.flush()
