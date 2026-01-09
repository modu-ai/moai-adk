"""All Is Well CLI Command

CLI command for executing the /all-is-well workflow which automates
Plan -> Run -> Sync in a single command.

Part of SPEC-CMD-001: /all-is-well Command Implementation
"""

import asyncio
import sys

import click
from rich.console import Console
from rich.panel import Panel
from rich.progress import Progress, SpinnerColumn, TextColumn

from moai_adk.web.models.workflow import WorkflowConfig, WorkflowReport
from moai_adk.web.services.workflow_service import WorkflowOrchestrator

console = Console()


async def run_workflow_async(config: WorkflowConfig) -> WorkflowReport:
    """Run the workflow asynchronously.

    Args:
        config: Workflow configuration

    Returns:
        Final workflow report
    """
    orchestrator = WorkflowOrchestrator()
    report = await orchestrator.start_workflow(config)
    return report


def run_workflow(config: WorkflowConfig) -> dict:
    """Run the workflow synchronously.

    Args:
        config: Workflow configuration

    Returns:
        Workflow result dictionary
    """
    try:
        report = asyncio.run(run_workflow_async(config))
        return {
            "workflow_id": report.workflow_id,
            "status": report.status if isinstance(report.status, str) else report.status.value,
            "phase": report.phase if isinstance(report.phase, str) else report.phase.value,
            "total_tokens": report.total_tokens,
            "total_cost_usd": report.total_cost_usd,
        }
    except Exception as e:
        return {
            "workflow_id": None,
            "status": "failed",
            "error": str(e),
        }


@click.command("all-is-well")
@click.argument("features", nargs=-1, required=True)
@click.option(
    "--worktree",
    is_flag=True,
    default=False,
    help="Use git worktrees for parallel development",
)
@click.option(
    "--parallel",
    "-p",
    type=int,
    default=1,
    help="Number of parallel workers for implementation",
)
@click.option(
    "--no-branch",
    is_flag=True,
    default=False,
    help="Skip creating feature branches",
)
@click.option(
    "--no-pr",
    is_flag=True,
    default=False,
    help="Skip creating pull requests",
)
@click.option(
    "--auto-merge",
    is_flag=True,
    default=False,
    help="Automatically merge PRs after approval",
)
@click.option(
    "--model",
    "-m",
    type=str,
    default="glm",
    help="Default model to use (glm, opus)",
)
def all_is_well(
    features: tuple[str, ...],
    worktree: bool,
    parallel: int,
    no_branch: bool,
    no_pr: bool,
    auto_merge: bool,
    model: str,
) -> None:
    """Execute the /all-is-well workflow.

    This command automates the complete Plan -> Run -> Sync workflow
    for implementing features in a single command.

    FEATURES: One or more feature descriptions to implement.

    Examples:

        moai-adk all-is-well "user authentication"

        moai-adk all-is-well "user auth" "dashboard" --parallel 2

        moai-adk all-is-well "api endpoints" --worktree --auto-merge
    """
    # Validate features
    if not features:
        console.print("[red]Error:[/red] At least one feature is required")
        sys.exit(1)

    # Create configuration
    config = WorkflowConfig(
        features=list(features),
        use_worktree=worktree,
        parallel_workers=parallel,
        create_branch=not no_branch,
        create_pr=not no_pr,
        auto_merge=auto_merge,
        model=model,
    )

    # Display start message
    console.print(
        Panel(
            f"[bold blue]Starting All Is Well Workflow[/bold blue]\n\n"
            f"Features: {', '.join(features)}\n"
            f"Worktree: {worktree}\n"
            f"Parallel Workers: {parallel}\n"
            f"Model: {model}",
            title="MoAI-ADK",
            border_style="blue",
        )
    )

    # Run workflow with progress
    with Progress(
        SpinnerColumn(),
        TextColumn("[progress.description]{task.description}"),
        console=console,
    ) as progress:
        task = progress.add_task("Starting workflow...", total=None)

        try:
            result = run_workflow(config)

            if result.get("status") == "failed":
                progress.update(task, description="[red]Workflow failed[/red]")
                console.print(f"\n[red]Error:[/red] {result.get('error', 'Unknown error')}")
                sys.exit(1)

            progress.update(task, description="[green]Workflow started[/green]")

        except Exception as e:
            progress.update(task, description="[red]Workflow failed[/red]")
            console.print(f"\n[red]Error:[/red] {e}")
            sys.exit(1)

    # Display result
    console.print(
        Panel(
            f"[bold green]Workflow Started Successfully[/bold green]\n\n"
            f"Workflow ID: {result.get('workflow_id', 'N/A')}\n"
            f"Status: {result.get('status', 'N/A')}\n"
            f"Phase: {result.get('phase', 'N/A')}",
            title="Result",
            border_style="green",
        )
    )


# For main CLI integration
if __name__ == "__main__":
    all_is_well()
