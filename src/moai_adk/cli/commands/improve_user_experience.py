"""
User Experience Improvement CLI Command
"""

import argparse
import asyncio
from pathlib import Path

from rich.console import Console
from rich.panel import Panel
from rich.progress import Progress, SpinnerColumn, TextColumn
from rich.table import Table

from moai_adk.utils.user_experience import (
    UserExperienceAnalyzer,
    generate_improvement_plan,
)

console = Console()


def create_parser(subparsers) -> argparse.ArgumentParser:
    """Create user experience improvement parser"""
    parser = subparsers.add_parser(
        "improve-ux",
        help="Analyze user experience improvements",
        description="Analyze user experience of online documentation portal and provide improvement plan",
    )

    parser.add_argument(
        "--url",
        "-u",
        type=str,
        default="https://adk.mo.ai.kr",
        help="URL to analyze (default: https://adk.mo.ai.kr)",
    )

    parser.add_argument(
        "--output", "-o", type=str, help="File path to save analysis results"
    )

    parser.add_argument(
        "--format",
        "-f",
        type=str,
        choices=["json", "markdown", "text"],
        default="markdown",
        help="Output format (default: markdown)",
    )

    parser.add_argument(
        "--verbose", "-v", action="store_true", help="Display detailed progress"
    )

    parser.add_argument(
        "--max-workers",
        "-w",
        type=int,
        default=5,
        help="Maximum number of concurrent tasks (default: 5)",
    )

    return parser


def format_metrics_table(metrics: dict) -> Table:
    """Format metrics table"""
    table = Table(title="User Experience Metrics")
    table.add_column("Category", style="cyan")
    table.add_column("Score", style="magenta")
    table.add_column("Status", style="green")

    # Performance metrics
    perf = metrics.get("performance")
    table.add_row(
        "Performance",
        f"{perf.success_rate if perf else 0:.2f}",
        "✅ Good" if (perf and perf.is_good) else "❌ Needs Improvement",
    )

    # Navigation metrics
    nav = metrics.get("navigation")
    table.add_row(
        "Navigation",
        f"{nav.structure_score if nav else 0:.2f}",
        "✅ Good" if (nav and nav.is_good) else "❌ Needs Improvement",
    )

    # Content metrics
    content = metrics.get("content")
    table.add_row(
        "Content",
        f"{content.accuracy_score if content else 0:.2f}",
        "✅ Good" if (content and content.is_good) else "❌ Needs Improvement",
    )

    # Accessibility metrics
    acc = metrics.get("accessibility")
    table.add_row(
        "Accessibility",
        f"{1.0 if (acc and acc.is_good) else 0.0:.2f}",
        "✅ Good" if (acc and acc.is_good) else "❌ Needs Improvement",
    )

    return table


def format_recommendations(recommendations: list) -> Table:
    """Format recommendations table"""
    table = Table(title="Improvement Recommendations")
    table.add_column("Recommendation", style="yellow")
    table.add_column("Priority", style="red")

    for i, rec in enumerate(recommendations, 1):
        priority = "Medium"  # Default
        if (
            "error" in rec.lower()
            or "keyboard" in rec.lower()
            or "accuracy" in rec.lower()
        ):
            priority = "High"
        elif "performance" in rec.lower() or "structure" in rec.lower():
            priority = "Medium"
        else:
            priority = "Low"

        table.add_row(f"{i}. {rec}", priority)

    return table


def format_timeline(timeline: dict) -> Table:
    """Format timeline table"""
    table = Table(title="Implementation Plan")
    table.add_column("Period", style="cyan")
    table.add_column("Tasks", style="magenta")

    for period, tasks in timeline.items():
        if tasks:
            table.add_row(
                period.replace("_", " ").title(),
                "\n".join(f"• {task}" for task in tasks),
            )

    return table


async def analyze_user_experience(
    url: str, max_workers: int = 5, verbose: bool = False
) -> dict:
    """Perform user experience analysis"""
    if verbose:
        console.print(f"[blue]Starting user experience analysis: {url}[/blue]")

    analyzer = UserExperienceAnalyzer(url, max_workers=max_workers)

    with Progress(
        SpinnerColumn(),
        TextColumn("[progress.description]{task.description}"),
        console=console,
        transient=True,
    ) as progress:
        # Create analysis task
        analysis_task = progress.add_task("Analyzing user experience...", total=None)

        # Execute analysis
        analysis_report = await analyzer.generate_report()

        progress.update(analysis_task, completed=True)

    return analysis_report


def generate_markdown_report(analysis_report: dict, improvement_plan: dict) -> str:
    """Generate markdown report"""
    report = []

    # Header
    report.append("# User Experience Improvement Report")
    report.append("")
    report.append(f"**Analysis Target**: {analysis_report.get('base_url', 'N/A')}")
    report.append(f"**Analysis Time**: {analysis_report.get('generated_at', 'N/A')}")
    report.append("")

    # Overall score
    overall_score = analysis_report.get("overall_score", 0)
    report.append(f"## Overall Score: {overall_score:.2f}")
    report.append("")

    # Metrics details
    report.append("### Metrics Details")
    report.append("")

    # Performance metrics
    performance = analysis_report.get("performance")
    report.append("#### Performance")
    report.append(
        f"- Success Rate: {performance.success_rate if performance else 0:.2f}"
    )
    report.append(
        f"- Average Load Time: {performance.load_time if performance else 0:.2f}s"
    )
    report.append(
        f"- Response Time: {performance.response_time if performance else 0:.2f}s"
    )
    report.append(f"- Error Rate: {performance.error_rate if performance else 0:.2f}")
    report.append("")

    # Navigation metrics
    navigation = analysis_report.get("navigation")
    report.append("#### Navigation")
    report.append(
        f"- Structure Score: {navigation.structure_score if navigation else 0:.2f}"
    )
    report.append(f"- Link Count: {navigation.link_count if navigation else 0}")
    report.append(f"- Depth: {navigation.depth if navigation else 0}")
    report.append(f"- Completeness: {navigation.completeness if navigation else 0:.2f}")
    report.append("")

    # Content metrics
    content = analysis_report.get("content")
    report.append("#### Content")
    report.append(f"- Accuracy: {content.accuracy_score if content else 0:.2f}")
    report.append(f"- Completeness: {content.completeness_score if content else 0:.2f}")
    report.append(f"- Organization: {content.organization_score if content else 0:.2f}")
    report.append(f"- Readability: {content.readability_score if content else 0:.2f}")
    report.append("")

    # Accessibility metrics
    accessibility = analysis_report.get("accessibility")
    report.append("#### Accessibility")
    report.append(
        f"- Keyboard Navigation: {'✅' if (accessibility and accessibility.keyboard_navigation) else '❌'}"
    )
    report.append(
        f"- Screen Reader Support: {'✅' if (accessibility and accessibility.screen_reader_support) else '❌'}"
    )
    report.append(
        f"- Color Contrast: {'✅' if (accessibility and accessibility.color_contrast) else '❌'}"
    )
    report.append(
        f"- Responsive Design: {'✅' if (accessibility and accessibility.responsive_design) else '❌'}"
    )
    report.append(
        f"- ARIA Labels: {'✅' if (accessibility and accessibility.aria_labels) else '❌'}"
    )
    report.append("")

    # Recommendations
    report.append("### Improvement Recommendations")
    report.append("")
    for rec in analysis_report.get("recommendations", []):
        report.append(f"- {rec}")
    report.append("")

    # Implementation plan
    report.append("### Implementation Plan")
    report.append("")
    report.append(
        f"**Estimated Duration**: {improvement_plan.get('estimated_duration', 'N/A')}"
    )
    report.append("")

    timeline = improvement_plan.get("timeline", {})
    for period, tasks in timeline.items():
        if tasks:
            report.append(f"#### {period.replace('_', ' ').title()}")
            for task in tasks:
                report.append(f"- {task}")
            report.append("")

    return "\n".join(report)


def run_command(args) -> int:
    """Run user experience improvement command"""
    try:
        # Execute analysis
        analysis_report = asyncio.run(
            analyze_user_experience(args.url, args.max_workers, args.verbose)
        )

        # Generate improvement plan
        improvement_plan = generate_improvement_plan(analysis_report)

        # Display results
        if args.verbose:
            console.print(
                Panel.fit(
                    f"Overall Score: {analysis_report['overall_score']:.2f}",
                    title="Analysis Results",
                    style="blue",
                )
            )

        # Display metrics table
        metrics_table = format_metrics_table(analysis_report)
        console.print(metrics_table)

        # Display recommendations
        if analysis_report.get("recommendations"):
            recommendations_table = format_recommendations(
                analysis_report["recommendations"]
            )
            console.print(recommendations_table)

        # Display implementation plan
        if improvement_plan.get("timeline"):
            timeline_table = format_timeline(improvement_plan["timeline"])
            console.print(timeline_table)

        # Save results
        if args.output:
            if args.format == "json":
                import json

                result = {
                    "analysis": analysis_report,
                    "improvement_plan": improvement_plan,
                }
                output_content = json.dumps(result, indent=2, ensure_ascii=False)
            else:
                output_content = generate_markdown_report(
                    analysis_report, improvement_plan
                )

            output_path = Path(args.output)
            output_path.write_text(output_content, encoding="utf-8")
            console.print(f"\n[green]Results saved:[/green] {output_path}")

        # Return exit code
        if analysis_report["overall_score"] >= 0.85:
            console.print(
                f"\n[green]✅ User experience is excellent (score: {analysis_report['overall_score']:.2f})[/green]"
            )
            return 0
        else:
            console.print(
                f"\n[yellow]⚠️  User experience needs improvement (score: {analysis_report['overall_score']:.2f})[/yellow]"
            )
            return 1

    except KeyboardInterrupt:
        console.print("\n[yellow]Analysis interrupted by user.[/yellow]")
        return 1
    except Exception as e:
        console.print(f"[red]Error occurred:[/red] {e}")
        return 1
