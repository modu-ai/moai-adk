#!/usr/bin/env python3
"""
Analyze command for MoAI-ADK

Analyze Claude Code sessions and generate improvement suggestions.
"""

from pathlib import Path
from typing import Optional

import click
from rich.console import Console
from rich.table import Table

from moai_adk.core.analysis.session_analyzer import SessionAnalyzer

console = Console()


@click.command()
@click.option("--days", "-d", default=7, help="Number of days to analyze (default: 7)")
@click.option("--output", "-o", type=click.Path(), help="Output file path")
@click.option("--verbose", "-v", is_flag=True, help="Verbose output")
@click.option("--report-only", "-r", is_flag=True, help="Generate report only (no console output)")
@click.option(
    "--project-path",
    "-p",
    type=click.Path(),
    help="Project root path (default: current directory)",
)
def analyze(
    days: int,
    output: Optional[Path],
    verbose: bool,
    report_only: bool,
    project_path: Optional[Path],
):
    """
    Analyze Claude Code sessions from the last N days

    Analyzes Claude Code session logs to identify usage patterns,
    error frequencies, and generate improvement suggestions.

    Examples:
        moai-adk analyze session
        moai-adk analyze session --days 14 --verbose
        moai-adk analyze session --output /path/to/report.md
    """
    if project_path is None:
        project_path = Path.cwd()

    # Validate project path
    if not (project_path / ".moai").exists():
        console.print("[red]Error:[/red] Not a MoAI-ADK project (missing .moai directory)")
        console.print(f"[blue]Current path:[/blue] {project_path}")
        return

    # Initialize analyzer
    analyzer = SessionAnalyzer(days_back=days, verbose=verbose)

    if not report_only:
        console.print(f"[blue]📊 Analyzing sessions from last {days} days...[/blue]")

    # Parse sessions
    patterns = analyzer.parse_sessions()

    if not report_only:
        console.print(f"[green]✅ Analyzed {patterns['total_sessions']} sessions[/green]")
        console.print(f"[blue]   Total events: {patterns['total_events']}[/blue]")

        # Display summary table
        table = Table(title="Session Summary")
        table.add_column("Metric", style="cyan")
        table.add_column("Value", style="green")

        table.add_row("Total Sessions", str(patterns["total_sessions"]))
        table.add_row("Total Events", str(patterns["total_events"]))
        table.add_row("Failed Sessions", str(patterns["failed_sessions"]))
        table.add_row("Success Rate", f"{patterns.get('success_rate', 0):.1f}%")

        console.print(table)

        # Show top tools
        if patterns["tool_usage"]:
            console.print("\n[bold]🔧 Top Tools Used:[/bold]")
            top_tools = sorted(patterns["tool_usage"].items(), key=lambda x: x[1], reverse=True)[:10]

            tools_table = Table()
            tools_table.add_column("Tool", style="cyan")
            tools_table.add_column("Usage", style="green")

            for tool, count in top_tools:
                tools_table.add_row(f"`{tool}`", str(count))

            console.print(tools_table)

    # Save report
    output_path_obj = Path(output) if output else None
    report_path = analyzer.save_report(output_path_obj, project_path)

    if not report_only:
        console.print(f"\n[green]📄 Report saved: {report_path}[/green]")

        # Show key suggestions
        report_content = analyzer.generate_report()
        if "💡 Improvement Suggestions" in report_content:
            console.print("\n[bold yellow]💡 Key Suggestions:[/bold yellow]")
            # Extract suggestions section
            suggestions_start = report_content.find("💡 Improvement Suggestions")
            if suggestions_start != -1:
                suggestions_section = report_content[suggestions_start:]
                # Extract first few suggestions
                lines = suggestions_section.split("\n")[2:]  # Skip header and empty line
                for line in lines[:10]:  # Show first 10 lines
                    if line.strip() and not line.startswith("---"):
                        console.print(line)
