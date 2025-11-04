# @CODE:CLI-PROMPTS-001 | SPEC: SPEC-CLI-001/spec.md
"""Project initialization prompts

Collect interactive project settings
"""

from pathlib import Path
from typing import TypedDict

import questionary
from rich.console import Console

console = Console()


class ProjectSetupAnswers(TypedDict):
    """Project setup answers"""

    project_name: str
    mode: str  # personal | team (default from init)
    locale: str  # ko | en (default from init)
    language: str | None  # Will be set in /alfred:0-project
    author: str  # Will be set in /alfred:0-project


def prompt_project_setup(
    project_name: str | None = None,
    is_current_dir: bool = False,
    project_path: Path | None = None,
    initial_locale: str | None = None,
) -> ProjectSetupAnswers:
    """Project setup prompt

    Args:
        project_name: Project name (asks when None)
        is_current_dir: Whether the current directory is being used
        project_path: Project path (used to derive the name)
        initial_locale: Preferred locale provided via CLI (optional)

    Returns:
        Project setup answers

    Raises:
        KeyboardInterrupt: When user cancels the prompt (Ctrl+C)
    """
    answers: ProjectSetupAnswers = {
        "project_name": "",
        "mode": "personal",  # Default: will be configurable in /alfred:0-project
        "locale": "en",      # Default: will be configurable in /alfred:0-project
        "language": None,    # Will be detected in /alfred:0-project
        "author": "",        # Will be set in /alfred:0-project
    }

    try:
        # SIMPLIFIED: Only ask for project name
        # All other settings (mode, locale, language, author) are now configured in /alfred:0-project

        # 1. Project name (only when not using the current directory)
        if not is_current_dir:
            if project_name:
                answers["project_name"] = project_name
                console.print(f"[cyan]📦 Project Name:[/cyan] {project_name}")
            else:
                result = questionary.text(
                    "📦 Project Name:",
                    default="my-moai-project",
                    validate=lambda text: len(text) > 0 or "Project name is required",
                ).ask()
                if result is None:
                    raise KeyboardInterrupt
                answers["project_name"] = result
        else:
            # Use the current directory name
            # Note: Path.cwd() reflects the process working directory (Codex CLI cwd)
            # Prefer project_path when provided (user execution location)
            if project_path:
                answers["project_name"] = project_path.name
            else:
                answers["project_name"] = Path.cwd().name  # fallback
            console.print(
                f"[cyan]📦 Project Name:[/cyan] {answers['project_name']} [dim](current directory)[/dim]"
            )

        # NOTE: All other configuration (mode, language, author) is now handled in /alfred:0-project
        # This significantly reduces init time and improves UX
        answers["locale"] = initial_locale or "en"

        return answers

    except KeyboardInterrupt:
        console.print("\n[yellow]Setup cancelled by user[/yellow]")
        raise
