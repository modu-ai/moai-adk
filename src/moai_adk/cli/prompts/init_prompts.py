"""Project initialization prompts

Collect interactive project settings with modern UI.
Supports multilingual prompts based on user's language selection.
"""

from pathlib import Path
from typing import TypedDict

from rich.console import Console

from moai_adk.cli.prompts.translations import get_translation

console = Console()


class ProjectSetupAnswers(TypedDict):
    """Project setup answers."""

    # Core settings
    project_name: str
    locale: str  # ko | en | ja | zh

    # Service settings
    service_type: str  # claude_subscription | claude_api | glm | hybrid
    claude_auth_type: str | None  # subscription | api_key (for hybrid mode)
    pricing_plan: str | None  # pro | max5 | max20 | basic | glm_pro | enterprise
    anthropic_api_key: str | None
    glm_api_key: str | None

    # Git settings
    git_mode: str  # manual | personal | team
    github_username: str | None

    # Output language settings
    git_commit_lang: str  # ko | en | ja | zh
    code_comment_lang: str  # ko | en | ja | zh
    doc_lang: str  # ko | en | ja | zh


def prompt_project_setup(
    project_name: str | None = None,
    is_current_dir: bool = False,
    project_path: Path | None = None,
    initial_locale: str | None = None,
) -> ProjectSetupAnswers:
    """Project setup prompt with modern UI.

    Implements 8-question flow:
    1. Conversation language selection
    2. Service selection (Claude subscription/API, GLM, Hybrid)
    3. Pricing plan (if API selected)
    4. API key input (if API selected)
    5. Project name
    6. Git mode
    7. GitHub username (if needed)
    8. Output language settings (commit, comment, docs)

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
        "locale": "en",
        "service_type": "claude_subscription",
        "claude_auth_type": None,
        "pricing_plan": None,
        "anthropic_api_key": None,
        "glm_api_key": None,
        "git_mode": "manual",
        "github_username": None,
        "git_commit_lang": "en",
        "code_comment_lang": "en",
        "doc_lang": "en",
    }

    try:
        # ========================================
        # Q1: Language Selection (always in English first)
        # ========================================
        console.print("\n[blue]🌐 Language Selection[/blue]")

        language_choices = [
            {"name": "Korean (한국어)", "value": "ko"},
            {"name": "English", "value": "en"},
            {"name": "Japanese (日本語)", "value": "ja"},
            {"name": "Chinese (中文)", "value": "zh"},
        ]

        language_values = ["ko", "en", "ja", "zh"]
        default_locale = initial_locale or "en"
        default_value = default_locale if default_locale in language_values else "en"

        language_choice = _prompt_select(
            "Select your conversation language:",
            choices=language_choices,
            default=default_value,
        )

        if language_choice is None:
            raise KeyboardInterrupt

        answers["locale"] = language_choice
        t = get_translation(language_choice)

        language_names = {
            "ko": "Korean (한국어)",
            "en": "English",
            "ja": "Japanese (日本語)",
            "zh": "Chinese (中文)",
        }
        console.print(f"[#DA7756]🌐 Selected:[/#DA7756] {language_names.get(language_choice, language_choice)}")

        # ========================================
        # Q2: Service Selection
        # ========================================
        console.print(f"\n[blue]{t['service_selection']}[/blue]")

        service_choices = [
            {
                "name": f"{t['opt_claude_subscription']} - {t['desc_claude_subscription']}",
                "value": "claude_subscription",
            },
            {
                "name": f"{t['opt_claude_api']} - {t['desc_claude_api']}",
                "value": "claude_api",
            },
            {"name": f"{t['opt_glm']} - {t['desc_glm']}", "value": "glm"},
            {"name": f"{t['opt_hybrid']} - {t['desc_hybrid']}", "value": "hybrid"},
        ]

        service_choice = _prompt_select(
            t["q_service"],
            choices=service_choices,
            default="claude_subscription",
        )

        if service_choice is None:
            raise KeyboardInterrupt

        answers["service_type"] = service_choice

        # ========================================
        # Q2.5: Claude Auth Type (for hybrid only)
        # ========================================
        if service_choice == "hybrid":
            console.print(f"\n[blue]{t['claude_auth_selection']}[/blue]")

            claude_auth_choices = [
                {
                    "name": f"{t['opt_claude_sub']} - {t['desc_claude_sub']}",
                    "value": "subscription",
                },
                {
                    "name": f"{t['opt_claude_api_key']} - {t['desc_claude_api_key']}",
                    "value": "api_key",
                },
            ]

            claude_auth_choice = _prompt_select(
                t["q_claude_auth_type"],
                choices=claude_auth_choices,
                default="subscription",
            )

            if claude_auth_choice is None:
                raise KeyboardInterrupt

            answers["claude_auth_type"] = claude_auth_choice

        # ========================================
        # Q3: Pricing Plan (conditional)
        # ========================================
        if service_choice in ("claude_api", "claude_subscription", "hybrid"):
            console.print(f"\n[blue]{t['pricing_selection']}[/blue]")

            claude_pricing_choices = [
                {"name": f"{t['opt_pro']} - {t['desc_pro']}", "value": "pro"},
                {"name": f"{t['opt_max5']} - {t['desc_max5']}", "value": "max5"},
                {"name": f"{t['opt_max20']} - {t['desc_max20']}", "value": "max20"},
            ]

            pricing_choice = _prompt_select(
                t["q_pricing_claude"],
                choices=claude_pricing_choices,
                default="pro",
            )

            if pricing_choice is None:
                raise KeyboardInterrupt

            answers["pricing_plan"] = pricing_choice

        if service_choice == "glm":
            console.print(f"\n[blue]{t['pricing_selection']}[/blue]")

            glm_pricing_choices = [
                {"name": f"{t['opt_basic']} - {t['desc_basic']}", "value": "basic"},
                {
                    "name": f"{t['opt_glm_pro']} - {t['desc_glm_pro']}",
                    "value": "glm_pro",
                },
                {
                    "name": f"{t['opt_enterprise']} - {t['desc_enterprise']}",
                    "value": "enterprise",
                },
            ]

            pricing_choice = _prompt_select(
                t["q_pricing_glm"],
                choices=glm_pricing_choices,
                default="basic",
            )

            if pricing_choice is None:
                raise KeyboardInterrupt

            answers["pricing_plan"] = pricing_choice

        # ========================================
        # Q4: API Key Input (conditional)
        # ========================================
        # Anthropic API key: only for claude_api, or hybrid with api_key auth
        needs_anthropic_key = service_choice == "claude_api" or (
            service_choice == "hybrid" and answers["claude_auth_type"] == "api_key"
        )

        if needs_anthropic_key:
            console.print(f"\n[blue]{t['api_key_input']}[/blue]")

            api_key = _prompt_password(t["q_api_key_anthropic"])

            if api_key is None:
                raise KeyboardInterrupt

            answers["anthropic_api_key"] = api_key
            console.print(f"[dim]{t['msg_api_key_stored']}[/dim]")

        # GLM API key: for glm or hybrid
        if service_choice in ("glm", "hybrid"):
            if service_choice == "glm" or not needs_anthropic_key:
                console.print(f"\n[blue]{t['api_key_input']}[/blue]")

            glm_key = _prompt_password(t["q_api_key_glm"])

            if glm_key is None:
                raise KeyboardInterrupt

            answers["glm_api_key"] = glm_key
            console.print(f"[dim]{t['msg_api_key_stored']}[/dim]")

        # ========================================
        # Q5: Project Name (editable for both cases)
        # ========================================
        console.print(f"\n[blue]{t['project_setup']}[/blue]")

        if not is_current_dir:
            # New project directory - use provided name or prompt
            default_name = project_name if project_name else "my-moai-project"
        else:
            # Current directory - use folder name as default but allow editing
            default_name = project_path.name if project_path else Path.cwd().name

        result = _prompt_text(
            t["q_project_name"],
            default=default_name,
            required=True,
        )
        if result is None:
            raise KeyboardInterrupt
        answers["project_name"] = result

        # ========================================
        # Q6: Git Mode
        # ========================================
        console.print(f"\n[blue]{t['git_setup']}[/blue]")

        git_choices = [
            {"name": f"{t['opt_manual']} - {t['desc_manual']}", "value": "manual"},
            {
                "name": f"{t['opt_personal']} - {t['desc_personal']}",
                "value": "personal",
            },
            {"name": f"{t['opt_team']} - {t['desc_team']}", "value": "team"},
        ]

        git_choice = _prompt_select(
            t["q_git_mode"],
            choices=git_choices,
            default="manual",
        )

        if git_choice is None:
            raise KeyboardInterrupt

        answers["git_mode"] = git_choice

        # ========================================
        # Q7: GitHub Username (conditional)
        # ========================================
        if git_choice in ("personal", "team"):
            github_username = _prompt_text(
                t["q_github_username"],
                required=True,
            )

            if github_username is None:
                raise KeyboardInterrupt

            answers["github_username"] = github_username

        # ========================================
        # Q8: Output Language Settings
        # ========================================
        console.print(f"\n[blue]{t['output_language']}[/blue]")

        lang_choices = [
            {"name": "English", "value": "en"},
            {"name": "Korean (한국어)", "value": "ko"},
            {"name": "Japanese (日本語)", "value": "ja"},
            {"name": "Chinese (中文)", "value": "zh"},
        ]

        # Default to conversation language
        default_output_lang = answers["locale"]

        # Commit message language
        commit_lang = _prompt_select(
            t["q_commit_lang"],
            choices=lang_choices,
            default=default_output_lang,
        )

        if commit_lang is None:
            raise KeyboardInterrupt

        answers["git_commit_lang"] = commit_lang

        # Code comment language
        comment_lang = _prompt_select(
            t["q_comment_lang"],
            choices=lang_choices,
            default=default_output_lang,
        )

        if comment_lang is None:
            raise KeyboardInterrupt

        answers["code_comment_lang"] = comment_lang

        # Documentation language
        doc_lang = _prompt_select(
            t["q_doc_lang"],
            choices=lang_choices,
            default=default_output_lang,
        )

        if doc_lang is None:
            raise KeyboardInterrupt

        answers["doc_lang"] = doc_lang

        console.print(f"\n[green]{t['msg_setup_complete']}[/green]")

        return answers

    except KeyboardInterrupt:
        t = get_translation(answers.get("locale", "en"))
        console.print(f"\n[yellow]{t['msg_cancelled']}[/yellow]")
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


def _prompt_password(
    message: str,
) -> str | None:
    """Display password input prompt.

    Args:
        message: Prompt message

    Returns:
        User input or None if cancelled
    """
    try:
        from moai_adk.cli.ui.prompts import styled_password

        return styled_password(message)
    except ImportError:
        import questionary

        result = questionary.password(
            message,
            validate=lambda text: len(text) > 0 or "API key is required",
        ).ask()
        return result
