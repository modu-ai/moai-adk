"""Project initialization prompts

Collect interactive project settings with modern UI.
"""

from pathlib import Path
from typing import TypedDict

from rich.console import Console

console = Console()


class ProjectSetupAnswers(TypedDict):
    """Project setup answers"""

    project_name: str
    mode: str  # personal | team (default from init)
    locale: str  # ko | en | ja | zh | other (default from init)
    language: str | None  # Will be set in /moai:0-project
    author: str  # Will be set in /moai:0-project
    custom_language: str | None  # User input for "other" language option


def prompt_project_setup(
    project_name: str | None = None,
    is_current_dir: bool = False,
    project_path: Path | None = None,
    initial_locale: str | None = None,
) -> ProjectSetupAnswers:
    """Project setup prompt with modern UI.

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
        "mode": "personal",  # Default: will be configurable in /moai:0-project
        "locale": "en",  # Default: will be configurable in /moai:0-project
        "language": None,  # Will be detected in /moai:0-project
        "author": "",  # Will be set in /moai:0-project
        "custom_language": None,  # User input for other language
    }

    try:
        # SIMPLIFIED: Only ask for project name
        # All other settings (mode, locale, language, author) are now configured in /moai:0-project

        # 1. Project name (only when not using the current directory)
        if not is_current_dir:
            if project_name:
                answers["project_name"] = project_name
                console.print(f"[#DA7756]📦 Project Name:[/#DA7756] {project_name}")
            else:
                # Try new UI, fallback to questionary
                result = _prompt_text(
                    "📦 Project Name:",
                    default="my-moai-project",
                    required=True,
                )
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
                f"[#DA7756]📦 Project Name:[/#DA7756] {answers['project_name']} [dim](current directory)[/dim]"
            )

        # 2. Language selection - Korean, English, Japanese, Chinese, Other
        console.print("\n[blue]🌐 Language Selection[/blue]")

        # Build choices list
        language_choices = [
            {"name": "Korean (한국어)", "value": "ko"},
            {"name": "English", "value": "en"},
            {"name": "Japanese (日本語)", "value": "ja"},
            {"name": "Chinese (中文)", "value": "zh"},
            {"name": "Other - Manual input", "value": "other"},
        ]

        # Determine default
        language_values = ["ko", "en", "ja", "zh", "other"]
        default_locale = initial_locale or "en"
        default_value = default_locale if default_locale in language_values else "en"

        language_choice = _prompt_select(
            "Select your conversation language:",
            choices=language_choices,
            default=default_value,
        )

        if language_choice is None:
            raise KeyboardInterrupt

        if language_choice == "other":
            # Prompt for manual input
            custom_lang = _prompt_text(
                "Enter your language:",
                required=True,
            )

            if custom_lang is None:
                raise KeyboardInterrupt

            answers["custom_language"] = custom_lang
            answers["locale"] = "other"  # When ISO code is not available
            console.print(f"[#DA7756]🌐 Selected Language:[/#DA7756] {custom_lang}")
        else:
            answers["locale"] = language_choice
            language_names = {
                "ko": "Korean (한국어)",
                "en": "English",
                "ja": "Japanese (日本語)",
                "zh": "Chinese (中文)",
            }
            console.print(
                f"[#DA7756]🌐 Selected Language:[/#DA7756] {language_names.get(language_choice, language_choice)}"
            )

        return answers

    except KeyboardInterrupt:
        console.print("\n[yellow]Setup cancelled by user[/yellow]")
        raise


def _prompt_text(
    message: str,
    default: str = "",
    required: bool = False,
) -> str | None:
    """Display text input prompt with modern UI fallback.

    Args:
        message: Prompt message
        default: Default value
        required: Whether input is required

    Returns:
        User input or None if cancelled
    """
    try:
        from moai_adk.cli.ui.prompts import styled_input

        return styled_input(message, default=default, required=required)
    except ImportError:
        import questionary

        if required:
            result = questionary.text(
                message,
                default=default,
                validate=lambda text: len(text) > 0 or "This field is required",
            ).ask()
        else:
            result = questionary.text(message, default=default).ask()
        return result


def _prompt_select(
    message: str,
    choices: list[dict[str, str]],
    default: str | None = None,
) -> str | None:
    """Display select prompt with modern UI fallback.

    Args:
        message: Prompt message
        choices: List of choices with name and value
        default: Default value

    Returns:
        Selected value or None if cancelled
    """
    try:
        from moai_adk.cli.ui.prompts import styled_select

        return styled_select(message, choices=choices, default=default)
    except ImportError:
        import questionary

        # Map choices for questionary format
        choice_names = [c["name"] for c in choices]
        value_map = {c["name"]: c["value"] for c in choices}

        # Find default name
        default_name = None
        if default:
            for c in choices:
                if c["value"] == default:
                    default_name = c["name"]
                    break

        result_name = questionary.select(
            message,
            choices=choice_names,
            default=default_name,
        ).ask()

        if result_name is None:
            return None

        return value_map.get(result_name)
