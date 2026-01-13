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
    """Project setup answers (GLM-only simplified flow)."""

    # Core settings
    project_name: str
    locale: str  # ko | en | ja | zh
    user_name: str  # User name for personalization (optional, can be empty)

    # GLM API key (optional)
    glm_api_key: str | None

    # Git settings
    git_mode: str  # manual | personal | team
    github_username: str | None

    # Output language settings
    git_commit_lang: str  # ko | en | ja | zh
    code_comment_lang: str  # ko | en | ja | zh
    doc_lang: str  # ko | en | ja | zh

    # TAG System (NEW - SPEC-TAG-002)
    tag_enabled: bool  # Q7.1: Enable TAG validation
    tag_mode: str  # Q7.2: TAG validation mode (warn | enforce | off)


def prompt_project_setup(
    project_name: str | None = None,
    is_current_dir: bool = False,
    project_path: Path | None = None,
    initial_locale: str | None = None,
) -> ProjectSetupAnswers:
    """Project setup prompt with modern UI (GLM-only simplified flow).

    Implements simplified 5-question flow:
    1. Conversation language selection
    2. User name (optional)
    3. GLM API key input (optional, can skip)
    4. Project name
    5. Git mode
    6. GitHub username (if needed)
    7. Output language settings (commit, comment, docs)

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
        "user_name": "",
        "glm_api_key": None,
        "git_mode": "manual",
        "github_username": None,
        "git_commit_lang": "en",
        "code_comment_lang": "en",
        "doc_lang": "en",
        "tag_enabled": True,
        "tag_mode": "warn",
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
        # Q1.5: User Name (optional)
        # ========================================
        console.print(f"\n[blue]{t['user_setup']}[/blue]")

        user_name = _prompt_text(
            t["q_user_name"],
            default="",
            required=False,
        )

        if user_name is None:
            raise KeyboardInterrupt

        answers["user_name"] = user_name.strip()
        if answers["user_name"]:
            console.print(f"[#DA7756]👤 Welcome:[/#DA7756] {answers['user_name']}")

        # ========================================
        # Q2: GLM API Key Input (optional)
        # ========================================
        from moai_adk.core.credentials import (
            glm_env_exists,
            load_glm_key_from_env,
            save_glm_key_to_env,
        )

        console.print(f"\n[blue]{t['api_key_input']}[/blue]")
        console.print("[dim]GLM CodePlan API key (optional - press Enter to skip)[/dim]\n")

        has_existing_key = glm_env_exists()
        existing_key = load_glm_key_from_env() if has_existing_key else None

        if has_existing_key and existing_key:
            masked_key = existing_key[:8] + "..." if existing_key else "..."
            console.print(f"[green]✓[/green] {t['msg_glm_key_found']} [cyan]{masked_key}[/cyan]")
            console.print(f"[dim]{t['msg_glm_key_keep_prompt']}[/dim]\n")
        else:
            console.print(f"[dim]{t['msg_glm_key_skip_guidance']}[/dim]\n")

        # Prompt for GLM API key (optional, can skip)
        glm_key = _prompt_password_optional(t["q_api_key_glm"])

        if glm_key is None:
            raise KeyboardInterrupt

        # Save new key if provided
        if glm_key:
            save_glm_key_to_env(glm_key)
            answers["glm_api_key"] = glm_key
            console.print(f"[dim]{t['msg_api_key_stored']}[/dim]")
        elif has_existing_key and existing_key:
            # User pressed Enter, keeping existing key
            answers["glm_api_key"] = existing_key
            console.print("[dim]Keeping existing GLM API key.[/dim]")
        else:
            # User skipped, no key provided - this is allowed
            answers["glm_api_key"] = None
            console.print("[yellow]⚠️  GLM API key skipped.[/yellow]")
            console.print(f"[dim]{t['msg_glm_key_skip_guidance']}[/dim]")

        # ========================================
        # Q3: Project Name (editable for both cases)
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
        # Q4: Git Mode
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
        # Q5: GitHub Username (conditional)
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
        # Q6: Output Language Settings
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

        # ========================================
        # Q7: TAG System Configuration (NEW - SPEC-TAG-002)
        # ========================================
        console.print(f"\n[blue]{t['tag_setup']}[/blue]")

        # Display TAG system introduction with TDD purpose
        console.print(f"\n[dim]{t['tag_system_intro']}[/dim]\n")

        # Display recommendation
        console.print(f"[cyan]💡 {t['tag_yes_recommendation']}[/cyan]\n")

        # Q7.1: Enable TAG validation
        tag_enabled_choice = _prompt_confirm(
            t["q_tag_enable"],
            default=True,
        )

        if tag_enabled_choice is None:
            raise KeyboardInterrupt

        answers["tag_enabled"] = tag_enabled_choice

        # Display result
        if tag_enabled_choice:
            console.print(f"[green]{t['msg_tag_enabled']}[/green]")
        else:
            console.print(f"[dim]{t['msg_tag_disabled']}[/dim]")

        # Conditional Q7.2: TAG validation mode
        if tag_enabled_choice:
            # Display mode guide
            console.print(f"\n[dim]{t['tag_mode_guide_title']}[/dim]")
            console.print(f"[dim]{t['tag_mode_guide_subtitle']}[/dim]\n")

            tag_mode_choices = [
                {
                    "name": f"{t['opt_tag_warn']} - {t['desc_tag_warn']}",
                    "value": "warn",
                },
                {
                    "name": f"{t['opt_tag_enforce']} - {t['desc_tag_enforce']}",
                    "value": "enforce",
                },
                {
                    "name": f"{t['opt_tag_off']} - {t['desc_tag_off']}",
                    "value": "off",
                },
            ]

            tag_mode_choice = _prompt_select(
                t["q_tag_mode"],
                choices=tag_mode_choices,
                default="warn",
            )

            if tag_mode_choice is None:
                raise KeyboardInterrupt

            answers["tag_mode"] = tag_mode_choice
            console.print(f"[#DA7756]TAG Mode:[/#DA7756] {tag_mode_choice}")
        else:
            answers["tag_mode"] = "off"
            console.print(f"[dim]{t['tag_no_warning']}[/dim]")

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
    except (ImportError, OSError, Exception):
        # Fallback to questionary on any error (including macOS OSError: Invalid argument)
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


def _prompt_confirm(
    message: str,
    default: bool = True,
) -> bool | None:
    """Display confirmation prompt (yes/no) with fallback.

    Args:
        message: Prompt message
        default: Default value (True for yes, False for no)

    Returns:
        True for yes, False for no, or None if cancelled
    """
    import questionary

    result = questionary.confirm(
        message,
        default=default,
    ).ask()

    # Returns True/False or None if cancelled
    return result


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


def _prompt_password_optional(
    message: str,
) -> str | None:
    """Display password input prompt with optional empty input.

    Allows pressing Enter to submit empty string (for keeping existing key).

    Args:
        message: Prompt message

    Returns:
        User input (empty string allowed), or None if cancelled
    """
    import questionary

    result = questionary.password(message).ask()
    # Return empty string as-is (user pressed Enter to keep existing key)
    # Return None only if user cancelled with Ctrl+C
    return result if result is not None else None
