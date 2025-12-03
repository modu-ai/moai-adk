"""Language management commands for MoAI-ADK.

Provides commands for language configuration, template processing,
and multilingual content generation.
"""

import json
from pathlib import Path

import click
from rich.console import Console
from rich.table import Table

from ...core.claude_integration import ClaudeCLIIntegration
from ...core.language_config import (
    LANGUAGE_CONFIG,
    get_all_supported_codes,
    get_native_name,
    get_optimal_model,
)
from ...core.template_engine import TemplateEngine

console = Console()


@click.group()
def language():
    """Language management and multilingual support."""
    pass


@language.command()
@click.option("--json-output", is_flag=True, help="Output as JSON")
def list(json_output):
    """List all supported languages."""
    if json_output:
        console.print(json.dumps(LANGUAGE_CONFIG, indent=2, ensure_ascii=False))
        return

    table = Table(title="Supported Languages")
    table.add_column("Code", style="cyan", no_wrap=True)
    table.add_column("English Name", style="green")
    table.add_column("Native Name", style="yellow")
    table.add_column("Family", style="blue")

    for code, info in LANGUAGE_CONFIG.items():
        table.add_row(code, info["name"], info["native_name"], info["family"])

    console.print(table)


@language.command()
@click.argument("language_code")
@click.option("--detail", is_flag=True, help="Show detailed information")
def info(language_code, detail):
    """Show information about a specific language."""
    lang_info = LANGUAGE_CONFIG.get(language_code.lower())

    if not lang_info:
        console.print(f"[red]Language code '{language_code}' not found.[/red]")
        console.print(f"Available codes: {', '.join(get_all_supported_codes())}")
        return

    console.print("[bold]Language Information:[/bold]")
    console.print(f"Code: {language_code}")
    console.print(f"English Name: {lang_info['name']}")
    console.print(f"Native Name: {lang_info['native_name']}")
    console.print(f"Family: {lang_info['family']}")

    if detail:
        optimal_model = get_optimal_model(language_code)
        console.print(f"Optimal Claude Model: {optimal_model}")


@language.command()
@click.argument("template_path", type=click.Path(exists=True))
@click.argument("variables_file", type=click.Path(exists=True))
@click.option("--output", "-o", type=click.Path(), help="Output file path")
@click.option("--language", "-l", help="Target language code")
def render_template(template_path, variables_file, output, language):
    """Render template with variables and language support."""
    try:
        # Load variables
        with open(variables_file, "r", encoding="utf-8") as f:
            variables = json.load(f)

        # Add language info if specified
        if language:
            variables["CONVERSATION_LANGUAGE"] = language
            variables["CONVERSATION_LANGUAGE_NAME"] = get_native_name(language)

        # Render template
        template_engine = TemplateEngine()
        template_path_obj = Path(template_path)
        output_path_obj = Path(output) if output else None

        rendered = template_engine.render_file(template_path_obj, variables, output_path_obj)

        if not output:
            console.print("[bold]Rendered Template:[/bold]")
            console.print(rendered)
        else:
            console.print(f"[green]Template rendered to: {output}[/green]")

    except Exception as e:
        console.print(f"[red]Error rendering template: {e}[/red]")


@language.command()
@click.argument("base_description")
@click.option("--target-languages", "-t", help="Comma-separated target language codes")
@click.option("--output", "-o", type=click.Path(), help="Output JSON file")
def translate_descriptions(base_description, target_languages, output):
    """Generate multilingual descriptions using Claude CLI."""
    try:
        if target_languages:
            languages = [lang.strip() for lang in target_languages.split(",")]
        else:
            languages = ["en", "ko", "ja", "es", "fr", "de"]

        claude_integration = ClaudeCLIIntegration()
        descriptions = claude_integration.generate_multilingual_descriptions({"base": base_description}, languages)

        if output:
            with open(output, "w", encoding="utf-8") as f:
                json.dump(descriptions, f, indent=2, ensure_ascii=False)
            console.print(f"[green]Descriptions saved to: {output}[/green]")
        else:
            console.print("[bold]Multilingual Descriptions:[/bold]")
            console.print(json.dumps(descriptions, indent=2, ensure_ascii=False))

    except Exception as e:
        console.print(f"[red]Error generating descriptions: {e}[/red]")


@language.command()
@click.argument("prompt_template")
@click.option("--variables", "-v", type=click.Path(exists=True), help="Variables JSON file")
@click.option("--language", "-l", help="Target language code")
@click.option("--output-format", default="json", help="Output format (text, json, stream-json)")
@click.option("--dry-run", is_flag=True, help="Show command without executing")
def execute(prompt_template, variables, language, output_format, dry_run):
    """Execute Claude CLI with template variables and language support."""
    try:
        # Load or create variables
        template_vars = {}
        if variables:
            with open(variables, "r", encoding="utf-8") as f:
                template_vars = json.load(f)

        if language:
            template_vars["CONVERSATION_LANGUAGE"] = language
            template_vars["CONVERSATION_LANGUAGE_NAME"] = get_native_name(language)

        # Process template
        template_engine = TemplateEngine()
        processed_prompt = template_engine.render_string(prompt_template, template_vars)

        if dry_run:
            console.print("[bold]Dry Run - Command that would be executed:[/bold]")
            console.print(f"claude --print --output-format {output_format} '{processed_prompt}'")
            console.print("[bold]Variables used:[/bold]")
            console.print(json.dumps(template_vars, indent=2, ensure_ascii=False))
            return

        # Execute with Claude integration
        claude_integration = ClaudeCLIIntegration()
        result = claude_integration.process_template_command(
            prompt_template, template_vars, print_mode=True, output_format=output_format
        )

        if result["success"]:
            console.print("[green]✓ Command executed successfully[/green]")
            if output_format == "json" and result["stdout"]:
                try:
                    output_data = json.loads(result["stdout"])
                    console.print(json.dumps(output_data, indent=2, ensure_ascii=False))
                except json.JSONDecodeError:
                    console.print(result["stdout"])
            else:
                console.print(result["stdout"])
        else:
            console.print("[red]✗ Command execution failed[/red]")
            if "error" in result:
                console.print(f"Error: {result['error']}")
            if result["stderr"]:
                console.print(f"Stderr: {result['stderr']}")

    except Exception as e:
        console.print(f"[red]Error executing command: {e}[/red]")


@language.command()
@click.argument("config_file", type=click.Path(exists=True))
@click.option("--validate-languages", is_flag=True, help="Validate language codes in config")
def validate_config(config_file, validate_languages):
    """Validate language configuration in MoAI-ADK config file."""
    try:
        with open(config_file, "r", encoding="utf-8") as f:
            config = json.load(f)

        console.print(f"[bold]Validating config: {config_file}[/bold]")

        # Basic structure validation
        if "language" not in config:
            console.print("[yellow]⚠ No 'language' section found in config[/yellow]")
        else:
            lang_config = config["language"]
            if not isinstance(lang_config, dict):
                console.print("[red]✗ 'language' section must be an object[/red]")
            else:
                console.print("[green]✓ Language section structure is valid[/green]")

                # Check conversation_language
                conv_lang = lang_config.get("conversation_language")
                if conv_lang:
                    if conv_lang in get_all_supported_codes():
                        console.print(f"[green]✓ conversation_language '{conv_lang}' is supported[/green]")
                    else:
                        console.print(f"[red]✗ conversation_language '{conv_lang}' is not supported[/red]")
                else:
                    console.print("[yellow]⚠ No conversation_language specified[/yellow]")

                # Check conversation_language_name
                conv_lang_name = lang_config.get("conversation_language_name")
                if conv_lang_name and conv_lang:
                    expected_name = get_native_name(conv_lang)
                    if conv_lang_name == expected_name:
                        console.print("[green]✓ conversation_language_name matches[/green]")
                    else:
                        warning_msg = (
                            f"⚠ conversation_language_name '{conv_lang_name}' doesn't match expected '{expected_name}'"
                        )
                        console.print(f"[yellow]{warning_msg}[/yellow]")

        if validate_languages:
            # Scan entire config for language codes
            config_str = json.dumps(config)
            found_codes = []
            for code in get_all_supported_codes():
                if code in config_str:
                    found_codes.append(code)

            if found_codes:
                console.print(f"[blue]Found language codes in config: {', '.join(found_codes)}[/blue]")

    except Exception as e:
        console.print(f"[red]Error validating config: {e}[/red]")
