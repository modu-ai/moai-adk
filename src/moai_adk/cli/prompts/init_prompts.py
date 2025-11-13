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
    locale: str  # ko | en | ja | zh | other (default from init)
    language: str | None  # Will be set in /alfred:0-project
    author: str  # Will be set in /alfred:0-project
    mcp_servers: list[str]  # Selected MCP servers to install
    custom_language: str | None  # User input for "other" language option


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
        "locale": "en",  # Default: will be configurable in /alfred:0-project
        "language": None,  # Will be detected in /alfred:0-project
        "author": "",  # Will be set in /alfred:0-project
        "mcp_servers": [],  # Selected MCP servers
        "custom_language": None,  # User input for other language
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

        # 2. Language selection - Korean, English, Japanese, Chinese, Other
        console.print("\n[blue]🌐 Language Selection[/blue]")

        language_choice = questionary.select(
            "Select your conversation language:",
            choices=[
                {"name": "한국어 (Korean)", "value": "ko"},
                {"name": "English", "value": "en"},
                {"name": "日本語 (Japanese)", "value": "ja"},
                {"name": "中文 (Chinese)", "value": "zh"},
                {"name": "Other - Manual input", "value": "other"},
            ],
            default=initial_locale or "en",
        ).ask()

        if language_choice is None:
            raise KeyboardInterrupt

        if language_choice == "other":
            # Prompt for manual input
            custom_lang = questionary.text(
                "Enter your language:",
                validate=lambda text: len(text) > 0 or "Language is required",
            ).ask()

            if custom_lang is None:
                raise KeyboardInterrupt

            answers["custom_language"] = custom_lang
            answers["locale"] = "other"  # When ISO code is not available
            console.print(f"[cyan]🌐 Selected Language:[/cyan] {custom_lang}")
        else:
            answers["locale"] = language_choice
            language_names = {
                "ko": "한국어 (Korean)",
                "en": "English",
                "ja": "日本語 (Japanese)",
                "zh": "中文 (Chinese)",
            }
            console.print(
                f"[cyan]🌐 Selected Language:[/cyan] {language_names.get(language_choice, language_choice)}"
            )

        # Auto-install MCP servers
        mcp_servers = ["context7", "playwright", "sequential-thinking"]
        answers["mcp_servers"] = mcp_servers
        console.print("\n[blue]🔧 MCP (Model Context Protocol) Configuration[/blue]")
        console.print(
            "[dim]Enhance AI capabilities with MCP servers (auto-installing recommended servers)[/dim]\n"
        )
        console.print(
            f"[green]✅ MCP servers auto-installed: {', '.join(mcp_servers)}[/green]"
        )

        return answers

    except KeyboardInterrupt:
        console.print("\n[yellow]Setup cancelled by user[/yellow]")
        raise
