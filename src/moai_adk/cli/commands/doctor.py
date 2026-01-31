"""MoAI-ADK doctor command

System diagnostics command:
- Check the Python version
- Verify Git installation
- Validate project structure
- Inspect language-specific tool chains
- Diagnose slash command loading issues (--check-commands)

## Skill Invocation Guide (English-Only)

### Related Skills
- **moai-foundation-langs**: For language toolchain verification and detection
  - Trigger: Use `--verbose` or `--fix` flag to inspect language-specific tools
  - Invocation: `Skill("moai-foundation-langs")` for detailed language stack analysis

- **moai-foundation-trust**: For TRUST 5-principles verification after fixing tools
  - Trigger: After running doctor with `--fix` to validate improvements
  - Invocation: `Skill("moai-foundation-trust")` to verify code quality toolchain

### When to Invoke Skills in Related Workflows
1. **After doctor diagnosis**:
   - Run `Skill("moai-foundation-trust")` to validate that all TRUST tools are properly configured
   - Run `Skill("moai-foundation-langs")` to confirm language-specific toolchains

2. **When tools are missing** (`--fix` flag):
   - Use suggested fixes from doctor command
   - Follow up with `Skill("moai-foundation-langs")` to validate corrections

3. **Debugging slash command issues** (`--check-commands`):
   - Run `Skill("moai-cc-commands")` if commands fail to load
   - Check `.claude/commands/` directory structure and permissions
"""

import json
import sys
from pathlib import Path

import click
import questionary
from rich.console import Console
from rich.table import Table

from moai_adk.core.project.checker import SystemChecker, check_environment
from moai_adk.core.project.detector import detect_project_language

# Force UTF-8 encoding for Windows compatibility
# Windows PowerShell/Console uses 'charmap' by default, which can't encode emojis
if sys.platform == "win32":
    console = Console(force_terminal=True, legacy_windows=False)
else:
    console = Console()


@click.command()
@click.option(
    "--verbose",
    "-v",
    is_flag=True,
    help="Show detailed tool versions and language detection",
)
@click.option("--fix", is_flag=True, help="Suggest fixes for missing tools")
@click.option("--export", type=click.Path(), help="Export diagnostics to JSON file")
@click.option("--check", type=str, help="Check specific tool only")
@click.option("--check-commands", is_flag=True, help="Diagnose slash command loading issues")
@click.option("--shell", is_flag=True, help="Diagnose shell and PATH configuration (WSL/Linux)")
def doctor(
    verbose: bool,
    fix: bool,
    export: str | None,
    check: str | None,
    check_commands: bool,
    shell: bool,
) -> None:
    """Check system requirements and project health

    Verifies:
    - Python version (>= 3.13)
    - Git installation
    - Project structure (.moai directory)
    - Language-specific tool chains (20+ languages)
    """
    try:
        # Handle --shell option first
        if shell:
            _check_shell_and_path(fix=fix)
            return

        # Handle --check-commands option
        if check_commands:
            _check_slash_commands()
            return

        console.print("[cyan]Running system diagnostics...[/cyan]\n")

        # Run basic environment checks
        results = check_environment()
        diagnostics_data: dict = {"basic_checks": results}

        # In verbose mode, verify language-specific toolchains
        if verbose or fix:
            language = detect_project_language()
            diagnostics_data["detected_language"] = language

            if verbose:
                console.print(f"[dim]Detected language: {language or 'Unknown'}[/dim]\n")

            if language:
                checker = SystemChecker()
                language_tools = checker.check_language_tools(language)
                diagnostics_data["language_tools"] = language_tools

                if verbose:
                    _display_language_tools(language, language_tools, checker)

        # Specific tool check
        if check:
            _check_specific_tool(check)
            return

        # Build the base results table
        table = Table(show_header=True, header_style="bold magenta")
        table.add_column("Check", style="dim", width=40)
        table.add_column("Status", justify="center")

        for check_name, status in results.items():
            icon = "✓" if status else "✗"
            color = "green" if status else "red"
            table.add_row(check_name, f"[{color}]{icon}[/{color}]")

        console.print(table)

        # In fix mode, suggest installation commands for missing tools
        if fix and "language_tools" in diagnostics_data:
            _suggest_fixes(
                diagnostics_data["language_tools"],
                diagnostics_data.get("detected_language"),
            )

        # When exporting, write diagnostics to JSON
        if export:
            _export_diagnostics(export, diagnostics_data)

        # Summarize the overall result
        all_passed = all(results.values())
        if all_passed:
            console.print("\n[green]✓ All checks passed[/green]")
        else:
            console.print("\n[yellow]⚠ Some checks failed[/yellow]")
            console.print("[dim]Run [cyan]python -m moai_adk doctor --verbose[/cyan] for detailed diagnostics[/dim]")

    except Exception as e:
        console.print(f"[red]✗ Diagnostic failed: {e}[/red]")
        raise


def _display_language_tools(language: str, tools: dict[str, bool], checker: SystemChecker) -> None:
    """Display a table of language-specific tools (helper)"""
    table = Table(show_header=True, header_style="bold cyan", title=f"{language.title()} Tools")
    table.add_column("Tool", style="dim")
    table.add_column("Status", justify="center")
    table.add_column("Version", style="blue")

    for tool, available in tools.items():
        icon = "✓" if available else "✗"
        color = "green" if available else "red"
        version = checker.get_tool_version(tool) if available else "not installed"

        table.add_row(tool, f"[{color}]{icon}[/{color}]", version or "")

    console.print(table)
    console.print()


def _check_specific_tool(tool: str) -> None:
    """Check only a specific tool (helper)"""
    checker = SystemChecker()
    available = checker._is_tool_available(tool)
    version = checker.get_tool_version(tool) if available else None

    if available:
        console.print(f"[green]✓ {tool} is installed[/green]")
        if version:
            console.print(f"  Version: {version}")
    else:
        console.print(f"[red]✗ {tool} is not installed[/red]")


def _suggest_fixes(tools: dict[str, bool], language: str | None) -> None:
    """Suggest installation commands for missing tools (helper)"""
    missing_tools = [tool for tool, available in tools.items() if not available]

    if not missing_tools:
        console.print("\n[green]✓ All tools are installed[/green]")
        return

    console.print(f"\n[yellow]⚠ Missing {len(missing_tools)} tool(s)[/yellow]")

    try:
        proceed = questionary.confirm(
            "Would you like to see install suggestions for missing tools?",
            default=True,
        ).ask()
    except Exception:
        proceed = True

    if not proceed:
        console.print("[yellow]User skipped install suggestions[/yellow]")
        return

    for tool in missing_tools:
        install_cmd = _get_install_command(tool, language)
        console.print(f"  [red]✗[/red] {tool}")
        if install_cmd:
            console.print(f"    Install: [cyan]{install_cmd}[/cyan]")


def _get_install_command(tool: str, language: str | None) -> str:
    """Return the install command for a given tool (helper)"""
    # Common tools with preferred package managers
    install_commands = {
        # Python tools (prefer uv)
        "pytest": "uv pip install pytest",
        "mypy": "uv pip install mypy",
        "ruff": "uv pip install ruff",
        # JavaScript tools
        "vitest": "npm install -D vitest",
        "biome": "npm install -D @biomejs/biome",
        "eslint": "npm install -D eslint",
        "jest": "npm install -D jest",
    }

    return install_commands.get(tool, f"# Install {tool} for {language}")


def _export_diagnostics(export_path: str, data: dict) -> None:
    """Export diagnostic results to a JSON file (helper)"""
    try:
        output = Path(export_path)
        output.write_text(json.dumps(data, indent=2, ensure_ascii=False), encoding="utf-8")
        console.print(f"\n[green]✓ Diagnostics exported to {export_path}[/green]")
    except Exception as e:
        console.print(f"\n[red]✗ Failed to export diagnostics: {e}[/red]")


def _check_shell_and_path(fix: bool = False) -> None:
    """Check shell environment and PATH configuration (helper)

    Diagnoses common issues:
    - WSL2 PATH not loading from ~/.bashrc
    - ~/.local/bin not in PATH
    - Shell configuration file issues
    """
    from moai_adk.utils.shell_validator import (
        auto_fix_path,
        diagnose_path,
        get_shell_info,
        is_wsl,
    )

    console.print("[cyan]Running shell and PATH diagnostics...[/cyan]\n")

    # Get shell info
    shell_info = get_shell_info()
    console.print(f"[dim]{shell_info}[/dim]\n")

    # Run PATH diagnostics
    diag = diagnose_path()

    # Build results table
    table = Table(show_header=True, header_style="bold cyan", title="PATH Configuration")
    table.add_column("Check", style="dim", width=35)
    table.add_column("Status", justify="center", width=10)
    table.add_column("Details", style="blue")

    # ~/.local/bin exists
    icon = "✓" if diag.local_bin_exists else "✗"
    color = "green" if diag.local_bin_exists else "yellow"
    table.add_row("~/.local/bin exists", f"[{color}]{icon}[/{color}]", "Required for MCP servers")

    # ~/.local/bin in current PATH
    icon = "✓" if diag.local_bin_in_path else "✗"
    color = "green" if diag.local_bin_in_path else "red"
    table.add_row("~/.local/bin in PATH", f"[{color}]{icon}[/{color}]", diag.shell_type)

    # Login shell PATH (for WSL/Linux)
    if diag.is_login_shell_env:
        icon = "✓" if diag.in_login_shell_path else "✗"
        color = "green" if diag.in_login_shell_path else "red"
        table.add_row("Login shell PATH configured", f"[{color}]{icon}[/{color}]", diag.config_file)

    console.print(table)
    console.print()

    # Show problem details if PATH is not configured
    if not diag.local_bin_in_path:
        console.print("[yellow]⚠ Problem Detected[/yellow]")
        console.print("~/.local/bin is NOT in your PATH.")
        console.print("This can cause MCP servers and CLI tools to fail.\n")

        if is_wsl() and diag.shell_type == "bash":
            console.print("[dim]WSL Explanation:[/dim]")
            console.print("[dim]WSL uses login shell which reads ~/.profile, not ~/.bashrc.[/dim]")
            console.print("[dim]Even if ~/.bashrc has PATH configured, it may not be loaded.[/dim]\n")

        console.print("[cyan]Recommended Fix:[/cyan]")
        console.print(f"  {diag.recommended_fix}\n")

        # Offer auto-fix if available
        if fix and diag.auto_fixable:
            try:
                proceed = questionary.confirm(
                    "Would you like to automatically fix the PATH configuration?",
                    default=True,
                ).ask()
            except Exception:
                proceed = False

            if proceed:
                success, message = auto_fix_path()
                if success:
                    console.print(f"[green]✓ {message}[/green]")
                else:
                    console.print(f"[red]✗ {message}[/red]")
        elif not fix:
            console.print("[dim]Run with --fix to automatically configure PATH[/dim]")
    else:
        console.print("[green]✓ PATH configuration looks good![/green]")


def _check_slash_commands() -> None:
    """Check slash command loading issues (helper)"""
    from moai_adk.core.diagnostics.slash_commands import diagnose_slash_commands

    console.print("[cyan]Running slash command diagnostics...[/cyan]\n")

    result = diagnose_slash_commands()

    # Handle error case
    if "error" in result:
        console.print(f"[red]✗ {result['error']}[/red]")
        return

    # Build results table
    table = Table(show_header=True, header_style="bold magenta")
    table.add_column("Command File", style="dim", width=40)
    table.add_column("Status", justify="center", width=10)
    table.add_column("Issues", style="yellow")

    for detail in result["details"]:
        icon = "✓" if detail["valid"] else "✗"
        color = "green" if detail["valid"] else "red"
        issues = ", ".join(detail["errors"]) if detail["errors"] else "-"

        table.add_row(detail["file"], f"[{color}]{icon}[/{color}]", issues)

    console.print(table)
    console.print()

    # Summary
    total = result["total_files"]
    valid = result["valid_commands"]

    if valid == total and total > 0:
        console.print(f"[green]✓ {valid}/{total} command files are valid[/green]")
    elif total == 0:
        console.print("[yellow]⚠ No command files found in .claude/commands/[/yellow]")
    else:
        console.print(f"[yellow]⚠ Only {valid}/{total} command files are valid[/yellow]")
        console.print("[dim]Fix the issues above to enable slash commands[/dim]")
