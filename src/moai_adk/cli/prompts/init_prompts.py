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
    mode: str  # personal | team
    locale: str  # ko | en
    language: str | None
    author: str


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
        "mode": "personal",
        "locale": "ko",
        "language": None,
        "author": "",
    }

    try:
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

        # 2. Project mode
        result = questionary.select(
            "🔧 Project Mode:",
            choices=[
                questionary.Choice("Personal (single developer)", value="personal"),
                questionary.Choice("Team (collaborative)", value="team"),
            ],
            default="personal",
        ).ask()
        if result is None:
            raise KeyboardInterrupt
        answers["mode"] = result
        answers["locale"] = initial_locale or "en"
        if initial_locale:
            console.print(
                f"[cyan]🌐 Preferred Language:[/cyan] {answers['locale']} (specified via CLI option)"
            )
        else:
            console.print(
                "[cyan]🌐 Preferred Language:[/cyan] en (default, changeable in /alfred:0-project)"
            )

        # 3. Programming language (auto-detect or manual)
        result = questionary.confirm(
            "🔍 Auto-detect programming language?",
            default=True,
        ).ask()
        if result is None:
            raise KeyboardInterrupt
        detect_language = result

        if not detect_language:
            result = questionary.select(
                "💻 Select programming language:",
                choices=[
                    "Python",
                    "TypeScript",
                    "JavaScript",
                    "Java",
                    "Go",
                    "Rust",
                    "Ruby",
                    "Dart",
                    "Swift",
                    "Kotlin",
                    "PHP",
                    "C#",
                    "C",
                    "C++",
                    "Elixir",
                    "Scala",
                    "Clojure",
                    "Haskell",
                    "Lua",
                    "OCaml",
                    "Generic",
                ],
            ).ask()
            if result is None:
                raise KeyboardInterrupt
            answers["language"] = result

        # 4. Author information (optional)
        result = questionary.confirm(
            "👤 Add author information? (optional)",
            default=False,
        ).ask()
        if result is None:
            raise KeyboardInterrupt
        add_author = result

        if add_author:
            result = questionary.text(
                "Author (GitHub ID):",
                default="",
                validate=lambda text: text.startswith("@") or "Must start with @",
            ).ask()
            if result is None:
                raise KeyboardInterrupt
            answers["author"] = result

        return answers

    except KeyboardInterrupt:
        console.print("\n[yellow]Setup cancelled by user[/yellow]")
        raise
