# SPEC: SPEC-CLI-001.md, SPEC-INIT-003.md
# TEST: tests/unit/test_cli_commands.py, tests/unit/test_init_reinit.py
"""MoAI-ADK init command

Project initialization command (interactive/non-interactive):
- Interactive Mode: Ask user for project settings
- Non-Interactive Mode: Use defaults or CLI options

## Skill Invocation Guide (English-Only)

### Related Skills
- **moai-foundation-langs**: For language detection and stack configuration
  - Trigger: When language parameter is not specified (auto-detection)
  - Invocation: Called implicitly during project initialization for language matrix detection

### When to Invoke Skills in Related Workflows
1. **After project initialization**:
   - Run `Skill("moai-foundation-trust")` to verify project structure and toolchain
   - Run `Skill("moai-foundation-langs")` to validate detected language stack

2. **Before first SPEC creation**:
   - Use `Skill("moai-core-language-detection")` to confirm language selection

3. **Project reinitialization** (`--force`):
   - Skills automatically adapt to new project structure
   - No manual intervention required
"""

import json
import os
import sys
from pathlib import Path
from typing import Sequence

import click
import yaml
from rich.console import Console
from rich.progress import BarColumn, Progress, SpinnerColumn, TaskID, TextColumn

from moai_adk import __version__
from moai_adk.cli.prompts import prompt_project_setup
from moai_adk.core.project.initializer import ProjectInitializer
from moai_adk.statusline.version_reader import (
    VersionConfig,
    VersionReader,
)
from moai_adk.utils.banner import print_banner, print_welcome_message

# Force UTF-8 encoding for Windows compatibility
# Windows PowerShell/Console uses 'charmap' by default, which can't encode emojis
if sys.platform == "win32":
    console = Console(force_terminal=True, legacy_windows=False)
else:
    console = Console()


def create_progress_callback(progress: Progress, task_ids: Sequence[TaskID]):
    """Create progress callback

    Args:
        progress: Rich Progress object
        task_ids: List of task IDs (one per phase)

    Returns:
        Progress callback function
    """

    def callback(message: str, current: int, total: int) -> None:
        """Update progress

        Args:
            message: Progress message
            current: Current phase (1-based)
            total: Total phases
        """
        # Complete current phase (1-based index → 0-based)
        if 1 <= current <= len(task_ids):
            progress.update(task_ids[current - 1], completed=1, description=message)

    return callback


@click.command()
@click.argument("path", type=click.Path(), default=".")
@click.option(
    "--non-interactive",
    "-y",
    is_flag=True,
    help="Non-interactive mode (use defaults)",
)
@click.option(
    "--mode",
    type=click.Choice(["personal", "team"]),
    default="personal",
    help="Project mode",
)
@click.option(
    "--locale",
    type=click.Choice(["ko", "en", "ja", "zh"]),
    default=None,
    help="Preferred language (ko/en/ja/zh, default: en)",
)
@click.option(
    "--language",
    type=str,
    default=None,
    help="Programming language (auto-detect if not specified)",
)
@click.option(
    "--force",
    is_flag=True,
    help="Force reinitialize without confirmation",
)
def init(
    path: str,
    non_interactive: bool,
    mode: str,
    locale: str,
    language: str | None,
    force: bool,
) -> None:
    """Initialize a new MoAI-ADK project

    Args:
        path: Project directory path (default: current directory)
        non_interactive: Skip prompts and use defaults
        mode: Project mode (personal/team)
        locale: Preferred language (ko/en/ja/zh). Interactive mode supports additional languages.
        language: Programming language
        with_mcp: Install specific MCP servers (can be used multiple times)
        mcp_auto: Auto-install all recommended MCP servers
        force: Force reinitialize without confirmation
    """
    try:
        # 1. Print banner with enhanced version info
        print_banner(__version__)

        # 2. Enhanced version reading with error handling
        try:
            version_config = VersionConfig(
                cache_ttl_seconds=10,  # Very short cache for CLI
                fallback_version=__version__,
                debug_mode=False,
            )
            version_reader = VersionReader(version_config)
            current_version = version_reader.get_version()

            # Log version info for debugging
            console.print(f"[dim]Current MoAI-ADK version: {current_version}[/dim]")
        except Exception as e:
            console.print(f"[yellow]⚠️ Version read error: {e}[/yellow]")

        # 3. Check current directory mode
        is_current_dir = path == "."
        project_path = Path(path).resolve()

        # Initialize variables
        custom_language = None

        # 3. Interactive vs Non-Interactive
        # Default values for new settings (GLM-only simplified flow)
        user_name = ""
        service_type = "glm"  # Always GLM
        pricing_plan = None  # Not used in GLM-only flow
        glm_pricing_plan = "basic"  # Default GLM pricing plan
        anthropic_api_key = None  # Not used in GLM-only flow
        glm_api_key = None
        git_mode = "manual"
        github_username = None
        git_commit_lang = "en"
        code_comment_lang = "en"
        doc_lang = "en"
        tag_enabled = True  # NEW - SPEC-TAG-002
        tag_mode = "warn"  # NEW - SPEC-TAG-002

        if non_interactive:
            # Non-Interactive Mode
            console.print(f"\n[cyan]🚀 Initializing project at {project_path}...[/cyan]\n")
            project_name = project_path.name if is_current_dir else path
            locale = locale or "en"
            # Read GLM API key from environment variable in non-interactive mode
            glm_api_key = os.getenv("MOAI_GLM_API_KEY") or os.getenv("GLM_API_KEY")
            # Language detection happens in /moai:0-project, so default to None here
            # This will become "generic" internally, but Summary will show more helpful message
            if not language:
                language = None
        else:
            # Interactive Mode
            print_welcome_message()

            # Interactive prompt with simplified GLM-only flow
            answers = prompt_project_setup(
                project_name=None if is_current_dir else path,
                is_current_dir=is_current_dir,
                project_path=project_path,
                initial_locale=locale,
            )

            # Extract answers (GLM-only flow)
            locale = answers["locale"]
            user_name = answers["user_name"]
            project_name = answers["project_name"]
            glm_api_key = answers["glm_api_key"]
            git_mode = answers["git_mode"]
            github_username = answers["github_username"]
            git_commit_lang = answers["git_commit_lang"]
            code_comment_lang = answers["code_comment_lang"]
            doc_lang = answers["doc_lang"]
            tag_enabled = answers["tag_enabled"]  # NEW - SPEC-TAG-002
            tag_mode = answers["tag_mode"]  # NEW - SPEC-TAG-002

            # GLM-only defaults (not prompted in simplified flow)
            service_type = "glm"
            glm_pricing_plan = "basic"
            pricing_plan = None
            anthropic_api_key = None

            # Map git_mode to mode for backward compatibility
            mode = "personal" if git_mode in ("personal", "team") else "personal"

            # Language detection happens in /moai:0-project
            language = None

            console.print("\n[cyan]🚀 Starting installation...[/cyan]\n")

        # 4. Check for reinitialization (SPEC-INIT-003 v0.3.0) - DEFAULT TO FORCE MODE
        initializer = ProjectInitializer(project_path)

        if initializer.is_initialized():
            # Always reinitialize without confirmation (force mode by default)
            if non_interactive:
                console.print("\n[green]🔄 Reinitializing project (force mode)...[/green]\n")
            else:
                # Interactive mode: Simple notification
                console.print("\n[cyan]🔄 Reinitializing project...[/cyan]")
                console.print("   Backup will be created at .moai-backups/backup/\n")

        # 5. Initialize project (Progress Bar with 5 phases)
        # Always allow reinit (force mode by default)
        is_reinit = initializer.is_initialized()

        # Reinit mode: update configuration files (v0.3.1+)
        # As of v0.37.0, we use section YAML files (.moai/config/sections/)
        # but still support legacy config.yaml and config.json for backward compatibility
        if is_reinit:
            # Migration: Remove old hook files (Issue #163)
            old_hook_files = [
                ".claude/hooks/alfred/session_start__startup.py",  # v0.8.0 deprecated
            ]
            for old_file in old_hook_files:
                old_path = project_path / old_file
                if old_path.exists():
                    try:
                        old_path.unlink()  # Remove old file
                    except Exception:
                        pass  # Ignore removal failures

            # Support both YAML (v0.32.5+) and JSON (legacy) config files
            # New projects use section files; legacy projects may still have config.yaml/json
            config_yaml_path = project_path / ".moai" / "config" / "config.yaml"
            config_json_path = project_path / ".moai" / "config" / "config.json"

            config_path = config_yaml_path if config_yaml_path.exists() else config_json_path
            is_yaml = config_path.suffix in (".yaml", ".yml")

            if config_path.exists():
                try:
                    with open(config_path, "r", encoding="utf-8") as f:
                        if is_yaml:
                            config_data = yaml.safe_load(f) or {}
                        else:
                            config_data = json.load(f)

                    # Update version and optimization flags
                    if "moai" not in config_data:
                        config_data["moai"] = {}

                    # Use enhanced version reader for consistent version handling
                    try:
                        version_config = VersionConfig(
                            cache_ttl_seconds=5,  # Very short cache for config update
                            fallback_version=__version__,
                            debug_mode=False,
                        )
                        version_reader = VersionReader(version_config)
                        current_version = version_reader.get_version()
                        config_data["moai"]["version"] = current_version
                    except Exception:
                        # Fallback to package version
                        config_data["moai"]["version"] = __version__

                    if "project" not in config_data:
                        config_data["project"] = {}
                    config_data["project"]["optimized"] = False

                    with open(config_path, "w", encoding="utf-8") as f:
                        if is_yaml:
                            yaml.safe_dump(
                                config_data,
                                f,
                                default_flow_style=False,
                                allow_unicode=True,
                                sort_keys=False,
                            )
                        else:
                            json.dump(config_data, f, indent=2, ensure_ascii=False)
                except Exception:
                    # Ignore read/write failures; config is regenerated during initialization
                    pass

        with Progress(
            SpinnerColumn(),
            TextColumn("[progress.description]{task.description}"),
            BarColumn(),
            TextColumn("[progress.percentage]{task.percentage:>3.0f}%"),
            console=console,
        ) as progress:
            # Create 5 phase tasks
            phase_names = [
                "Phase 1: Preparation and backup...",
                "Phase 2: Creating directory structure...",
                "Phase 3: Installing resources...",
                "Phase 4: Generating configurations...",
                "Phase 5: Validation and finalization...",
            ]
            task_ids = [progress.add_task(name, total=1) for name in phase_names]
            callback = create_progress_callback(progress, task_ids)

            result = initializer.initialize(
                mode=mode,
                locale=locale,
                language=language,
                custom_language=custom_language,
                backup_enabled=True,
                progress_callback=callback,
                reinit=True,  # Always allow reinit (force mode by default)
                tag_enabled=tag_enabled,  # NEW - SPEC-TAG-002
                tag_mode=tag_mode,  # NEW - SPEC-TAG-002
            )

        # 5.5. Save additional configuration (both interactive and non-interactive)
        if result.success:
            # In non-interactive mode, only save API keys if provided via environment
            if non_interactive:
                # Only save GLM key if provided via environment variable
                if glm_api_key:
                    _save_glm_key(glm_api_key)
            else:
                # Interactive mode: save all additional config
                _save_additional_config(
                    project_path=project_path,
                    project_name=project_name,
                    locale=locale,
                    user_name=user_name,
                    service_type=service_type,
                    pricing_plan=pricing_plan,
                    glm_pricing_plan=glm_pricing_plan,
                    anthropic_api_key=anthropic_api_key,
                    glm_api_key=glm_api_key,
                    git_mode=git_mode,
                    github_username=github_username,
                    git_commit_lang=git_commit_lang,
                    code_comment_lang=code_comment_lang,
                    doc_lang=doc_lang,
                )

        # 6. Output results
        if result.success:
            separator = "[dim]" + ("─" * 60) + "[/dim]"
            console.print("\n[green bold]✅ Initialization Completed Successfully![/green bold]")
            console.print(separator)
            console.print("\n[cyan]📊 Summary:[/cyan]")
            console.print(f"  [dim]📁 Location:[/dim]  {result.project_path}")
            # Show language more clearly - "generic" means auto-detect
            language_display = "Auto-detect (use /moai:0-project)" if result.language == "generic" else result.language
            console.print(f"  [dim]🌐 Language:[/dim]  {language_display}")
            # Show Git Strategy (default: manual = local-only, no auto-branch)
            console.print("  [dim]🔀 Git:[/dim]       manual (github-flow, branch: manual)")
            console.print(f"  [dim]🌍 Locale:[/dim]    {result.locale}")
            console.print(f"  [dim]📄 Files:[/dim]     {len(result.created_files)} created")
            console.print(f"  [dim]⏱️  Duration:[/dim]  {result.duration}ms")

            # Show backup info if reinitialized
            if is_reinit:
                backup_dir = project_path / ".moai-backups"
                if backup_dir.exists():
                    latest_backup = max(backup_dir.iterdir(), key=lambda p: p.stat().st_mtime)
                    console.print(f"  [dim]💾 Backup:[/dim]    {latest_backup.name}/")

            console.print(f"\n{separator}")

            # Show config merge notice if reinitialized
            if is_reinit:
                console.print("\n[yellow]⚠️  Configuration Status: optimized=false (merge required)[/yellow]")
                console.print()
                console.print("[cyan]What Happened:[/cyan]")
                console.print("  ✅ Template files updated to latest version")
                console.print("  💾 Your previous settings backed up in: [cyan].moai-backups/backup/[/cyan]")
                console.print("  ⏳ Configuration merge required")
                console.print()
                console.print("[cyan]What is optimized=false?[/cyan]")
                console.print("  • Template version changed (you get new features)")
                console.print("  • Your previous settings are safe (backed up)")
                console.print("  • Next: Run /moai:0-project to merge")
                console.print()
                console.print("[cyan]What Happens Next:[/cyan]")
                console.print("  1. Run [bold]/moai:0-project[/bold] in Claude Code")
                console.print("  2. System intelligently merges old settings + new template")
                console.print("  3. After successful merge → optimized becomes true")
                console.print("  4. You're ready to continue developing\n")

            console.print("\n[cyan]🚀 Next Steps:[/cyan]")
            if not is_current_dir:
                console.print(f"  [blue]1.[/blue] Run [bold]cd {project_name}[/bold] to enter the project")
                console.print("  [blue]2.[/blue] Run [bold]/moai:0-project[/bold] in Claude Code for full setup")
                console.print("     (Configure: mode, language, report generation, etc.)")
            else:
                console.print("  [blue]1.[/blue] Run [bold]/moai:0-project[/bold] in Claude Code for full setup")
                console.print("     (Configure: mode, language, report generation, etc.)")

            if not is_current_dir:
                console.print("  [blue]3.[/blue] Start developing with MoAI-ADK!\n")
            else:
                console.print("  [blue]2.[/blue] Start developing with MoAI-ADK!\n")

            # Prompt for MoAI Rank hook installation if eligible
            try:
                from moai_adk.rank.hook import prompt_hook_installation

                prompt_hook_installation(console=console)
            except ImportError:
                pass  # rank module not available
        else:
            console.print("\n[red bold]❌ Initialization Failed![/red bold]")
            if result.errors:
                console.print("\n[red]Errors:[/red]")
                for error in result.errors:
                    console.print(f"  [red]•[/red] {error}")
            console.print()
            raise click.ClickException("Installation failed")

    except KeyboardInterrupt:
        console.print("\n\n[yellow]⚠ Initialization cancelled by user[/yellow]\n")
        raise click.Abort()
    except FileExistsError as e:
        console.print("\n[yellow]⚠ Project already initialized[/yellow]")
        console.print("[dim]  Use 'python -m moai_adk status' to check configuration[/dim]\n")
        raise click.Abort() from e
    except Exception as e:
        console.print(f"\n[red]✗ Initialization failed: {e}[/red]\n")
        raise click.ClickException(str(e)) from e
    finally:
        # Explicitly flush output buffer
        console.file.flush()


def _save_glm_key(glm_api_key: str) -> None:
    """Save GLM API key to .env.glm file.

    Args:
        glm_api_key: The GLM API key to save
    """
    import os

    from moai_adk.core.credentials import (
        get_env_glm_path,
        remove_glm_key_from_shell_config,
        save_glm_key_to_env,
    )

    save_glm_key_to_env(glm_api_key)
    console.print(f"[green]✓[/green] GLM API key saved to [cyan]{get_env_glm_path()}[/cyan]")

    # Automatically remove GLM_API_KEY from shell config if present
    if os.environ.get("GLM_API_KEY"):
        results = remove_glm_key_from_shell_config()
        modified = [name for name, mod in results.items() if mod]
        if modified:
            files_str = ", ~/.".join(modified)
            console.print(f"[green]✓[/green] Removed GLM_API_KEY from: [cyan]~/.{files_str}[/cyan]")
            console.print("[dim]Backup: .moai-backup | Run 'source ~/.zshrc' to apply[/dim]")


def _save_additional_config(
    project_path: Path,
    project_name: str,
    locale: str,
    user_name: str,
    service_type: str,
    pricing_plan: str | None,
    glm_pricing_plan: str | None,
    anthropic_api_key: str | None,
    glm_api_key: str | None,
    git_mode: str,
    github_username: str | None,
    git_commit_lang: str,
    code_comment_lang: str,
    doc_lang: str,
) -> None:
    """Save additional configuration from interactive mode.

    Args:
        project_path: Project directory path
        project_name: Project name
        locale: Conversation language (ko, en, ja, zh)
        user_name: User name for personalization (can be empty)
        service_type: Service type (claude_subscription, claude_api, glm, hybrid)
        pricing_plan: Claude pricing plan (pro, max5, max20)
        glm_pricing_plan: GLM pricing plan (basic, glm_pro, enterprise)
        anthropic_api_key: Anthropic API key (if provided)
        glm_api_key: GLM API key (if provided)
        git_mode: Git mode (manual, personal, team)
        github_username: GitHub username (if provided)
        git_commit_lang: Commit message language
        code_comment_lang: Code comment language
        doc_lang: Documentation language
    """
    sections_dir = project_path / ".moai" / "config" / "sections"
    # Ensure sections directory exists
    sections_dir.mkdir(parents=True, exist_ok=True)

    # 1. Save API keys to global locations
    # Anthropic key → ~/.moai/credentials.yaml
    # GLM key → ~/.moai/.env.glm (new dotenv format)
    if anthropic_api_key:
        from moai_adk.core.credentials import save_credentials

        save_credentials(
            anthropic_api_key=anthropic_api_key,
            merge=True,  # Preserve existing keys
        )

    if glm_api_key:
        import os

        from moai_adk.core.credentials import (
            get_env_glm_path,
            remove_glm_key_from_shell_config,
            save_glm_key_to_env,
        )

        save_glm_key_to_env(glm_api_key)
        console.print(f"[green]✓[/green] GLM API key saved to [cyan]{get_env_glm_path()}[/cyan]")

        # Automatically remove GLM_API_KEY from shell config if present
        if os.environ.get("GLM_API_KEY"):
            results = remove_glm_key_from_shell_config()
            modified = [name for name, mod in results.items() if mod]
            if modified:
                files_str = ", ~/.".join(modified)
                console.print(f"[green]✓[/green] Removed GLM_API_KEY from: [cyan]~/.{files_str}[/cyan]")
                console.print("[dim]Backup: .moai-backup | Run 'source ~/.zshrc' to apply[/dim]")

    # 2. Save service/pricing to pricing.yaml
    pricing_path = sections_dir / "pricing.yaml"
    if pricing_path.exists():
        try:
            pricing_data = yaml.safe_load(pricing_path.read_text()) or {}
        except Exception:
            pricing_data = {}
    else:
        pricing_data = {}

    if "service" not in pricing_data:
        pricing_data["service"] = {}

    pricing_data["service"]["type"] = service_type
    if pricing_plan:
        pricing_data["service"]["claude_pricing_plan"] = pricing_plan
    if glm_pricing_plan:
        pricing_data["service"]["glm_pricing_plan"] = glm_pricing_plan

    with open(pricing_path, "w", encoding="utf-8") as f:
        yaml.safe_dump(
            pricing_data,
            f,
            default_flow_style=False,
            allow_unicode=True,
            sort_keys=False,
        )

    # 3. Update language.yaml with output language settings
    language_path = sections_dir / "language.yaml"
    if language_path.exists():
        try:
            lang_data = yaml.safe_load(language_path.read_text()) or {}
        except Exception:
            lang_data = {}
    else:
        lang_data = {}

    if "language" not in lang_data:
        lang_data["language"] = {}

    # Conversation language mapping
    locale_names = {
        "ko": "Korean (한국어)",
        "en": "English",
        "ja": "Japanese (日本語)",
        "zh": "Chinese (中文)",
    }

    lang_data["language"]["conversation_language"] = locale
    lang_data["language"]["conversation_language_name"] = locale_names.get(locale, "English")
    lang_data["language"]["git_commit_messages"] = git_commit_lang
    lang_data["language"]["code_comments"] = code_comment_lang
    lang_data["language"]["documentation"] = doc_lang

    with open(language_path, "w", encoding="utf-8") as f:
        yaml.safe_dump(lang_data, f, default_flow_style=False, allow_unicode=True, sort_keys=False)

    # 4. Update git-strategy.yaml with git mode
    git_path = sections_dir / "git-strategy.yaml"
    if git_path.exists():
        try:
            git_data = yaml.safe_load(git_path.read_text()) or {}
        except Exception:
            git_data = {}
    else:
        git_data = {}

    if "git_strategy" not in git_data:
        git_data["git_strategy"] = {}

    git_data["git_strategy"]["mode"] = git_mode

    with open(git_path, "w", encoding="utf-8") as f:
        yaml.safe_dump(git_data, f, default_flow_style=False, allow_unicode=True, sort_keys=False)

    # 5. Update project.yaml with project name and GitHub username
    project_yaml_path = sections_dir / "project.yaml"
    if project_yaml_path.exists():
        try:
            project_data = yaml.safe_load(project_yaml_path.read_text()) or {}
        except Exception:
            project_data = {}
    else:
        project_data = {}

    # Save project name
    if "project" not in project_data:
        project_data["project"] = {}
    project_data["project"]["name"] = project_name

    # Save GitHub username if provided
    if github_username:
        if "github" not in project_data:
            project_data["github"] = {}
        project_data["github"]["profile_name"] = github_username

    with open(project_yaml_path, "w", encoding="utf-8") as f:
        yaml.safe_dump(
            project_data,
            f,
            default_flow_style=False,
            allow_unicode=True,
            sort_keys=False,
        )

    # 6. Update user.yaml with user name
    user_yaml_path = sections_dir / "user.yaml"
    if user_yaml_path.exists():
        try:
            user_data = yaml.safe_load(user_yaml_path.read_text()) or {}
        except Exception:
            user_data = {}
    else:
        user_data = {}

    if "user" not in user_data:
        user_data["user"] = {}

    user_data["user"]["name"] = user_name

    with open(user_yaml_path, "w", encoding="utf-8") as f:
        yaml.safe_dump(
            user_data,
            f,
            default_flow_style=False,
            allow_unicode=True,
            sort_keys=False,
        )
