"""
Claude Code CLI integration for advanced variable substitution and automation.

Enables headless operation with template variable processing, JSON streaming,
and multi-language support for commands, agents, and output styles.
"""

import json
import logging
import subprocess
import tempfile
from pathlib import Path
from typing import Any, Dict, List, Optional, Union

from .language_config import get_language_info
from .template_engine import TemplateEngine

logger = logging.getLogger(__name__)

# Constants for subprocess output limits
MAX_LINES = 1000  # Default maximum lines for output truncation


def _safe_run_subprocess(
    cmd: list[str],
    *,
    timeout: int = 60,
    max_output_size: int = 1024 * 1024,  # 1MB
    capture_output: bool = True,
    check: bool = False,
    cwd: Optional[Path] = None,
    **kwargs: Any,
) -> subprocess.CompletedProcess:
    """Safely run subprocess command with memory protection.

    Args:
        cmd: Command list to execute
        timeout: Maximum execution time in seconds (default: 60s)
        max_output_size: Maximum output size in bytes (default: 1MB)
        capture_output: Whether to capture stdout/stderr
        check: Whether to raise CalledProcessError on non-zero exit
        cwd: Working directory
        **kwargs: Additional arguments passed to subprocess.run

    Returns:
        CompletedProcess with captured output (possibly truncated)

    Raises:
        subprocess.TimeoutExpired: If command exceeds timeout
        subprocess.CalledProcessError: If check=True and command fails
    """
    if capture_output:
        kwargs.setdefault("text", True)
        kwargs.setdefault("encoding", "utf-8")

    try:
        result = subprocess.run(cmd, capture_output=capture_output, timeout=timeout, cwd=cwd, **kwargs)

        if capture_output and result.stdout:
            result.stdout = _truncate_output(result.stdout, max_output_size, MAX_LINES, "stdout")

        if capture_output and result.stderr:
            result.stderr = _truncate_output(result.stderr, max_output_size, MAX_LINES, "stderr")

        return result

    except subprocess.TimeoutExpired as e:
        logger.warning(f"Subprocess timeout after {timeout}s: {cmd}")
        if e.stderr:
            e.stderr = _truncate_output(str(e.stderr) if e.stderr else "", max_output_size, MAX_LINES, "stderr")  # type: ignore[assignment]
        raise


def _truncate_output(output: str, max_size: int, max_lines: int = 1000, source: str = "output") -> str:
    """Truncate output string if it exceeds max size.

    Args:
        output: Output string to truncate
        max_size: Maximum allowed size in bytes
        max_lines: Maximum allowed lines (default: 1000)
        source: Source name for logging (e.g., "stdout", "stderr")

    Returns:
        Truncated output with warning if truncation occurred
    """
    try:
        encoded_size = output.encode("utf-8")
    except UnicodeEncodeError:
        encoded_size = output.encode("utf-8", errors="replace")

    if len(encoded_size) > max_size or output.count("\n") >= max_lines:
        truncated = encoded_size[:max_size].decode("utf-8", errors="replace")
        line_count = output.count("\n") + 1
        warning = f"\n\n[WARNING: Output from {source} truncated to {max_size} bytes ({line_count} lines)]"
        logger.warning(f"Output from {source} truncated (size: {len(encoded_size)}, limit: {max_size})")
        return truncated + warning

    return output


class ClaudeCLIIntegration:
    """
    Advanced Claude CLI integration with template variable processing.

    Features:
    - Template variable substitution using MoAI-ADK's TemplateEngine
    - JSON streaming input/output support
    - Multi-language description processing
    - Headless automation capabilities
    - Configuration file generation and management
    """

    def __init__(self, template_engine: Optional[TemplateEngine] = None):
        """Initialize Claude CLI integration.

        Args:
            template_engine: TemplateEngine instance for variable processing
        """
        self.template_engine = template_engine or TemplateEngine()

    def generate_claude_settings(self, variables: Dict[str, Any], output_path: Optional[Path] = None) -> Path:
        """Generate Claude settings JSON file with variables.

        Args:
            variables: Template variables to include
            output_path: Path for settings file (auto-generated if None)

        Returns:
            Path to generated settings file
        """
        if output_path is None:
            temp_file = tempfile.NamedTemporaryFile(mode="w", suffix=".json", delete=False)
            output_path = Path(temp_file.name)
            temp_file.close()

        settings = {
            "variables": variables,
            "template_context": {
                "conversation_language": variables.get("CONVERSATION_LANGUAGE", "en"),
                "conversation_language_name": variables.get("CONVERSATION_LANGUAGE_NAME", "English"),
                "project_name": variables.get("PROJECT_NAME", ""),
                "codebase_language": variables.get("CODEBASE_LANGUAGE", "python"),
            },
        }

        output_path.write_text(json.dumps(settings, indent=2, ensure_ascii=False), encoding="utf-8")
        return output_path

    def process_template_command(
        self,
        command_template: str,
        variables: Dict[str, Any],
        print_mode: bool = True,
        output_format: str = "json",
    ) -> Dict[str, Any]:
        """Process Claude command with template variables.

        Args:
            command_template: Command template with {{VARIABLE}} placeholders
            variables: Variables for substitution
            print_mode: Use --print flag for non-interactive execution
            output_format: Output format (text, json, stream-json)

        Returns:
            Process result dictionary
        """
        try:
            # Process template variables
            processed_command = self.template_engine.render_string(command_template, variables)

            # Build Claude CLI command
            cmd_parts = ["claude"]

            if print_mode:
                cmd_parts.extend(["--print"])
                cmd_parts.extend(["--output-format", output_format])

            # Add variable settings
            settings_file = self.generate_claude_settings(variables)
            cmd_parts.extend(["--settings", str(settings_file)])

            # Add processed command
            cmd_parts.append(processed_command)

            # Execute Claude CLI
            result = _safe_run_subprocess(
                cmd_parts, timeout=60, max_output_size=1024 * 1024, max_lines=1000, capture_output=True
            )

            # Cleanup settings file
            try:
                settings_file.unlink()
            except OSError:
                pass  # Ignore cleanup errors

            return {
                "success": result.returncode == 0,
                "stdout": result.stdout,
                "stderr": result.stderr,
                "returncode": result.returncode,
                "processed_command": processed_command,
                "variables_used": variables,
            }

        except Exception as e:
            return {
                "success": False,
                "error": str(e),
                "processed_command": command_template,
                "variables_used": variables,
            }

    def generate_multilingual_descriptions(
        self,
        base_descriptions: Dict[str, str],
        target_languages: Optional[list[str]] = None,
    ) -> Dict[str, Dict[str, str]]:
        """Generate multilingual descriptions for commands/agents.

        Args:
            base_descriptions: English base descriptions
            target_languages: Target language codes (auto-detected if None)

        Returns:
            Multilingual descriptions dictionary
        """
        if target_languages is None:
            # Auto-detect from variables or use common languages
            # System provides 4 languages: en, ko, ja, zh
            target_languages = ["en", "ko", "ja", "zh"]

        multilingual = {}

        for item_id, base_desc in base_descriptions.items():
            multilingual[item_id] = {"en": base_desc}

            # Generate descriptions for target languages
            for lang_code in target_languages:
                if lang_code == "en":
                    continue  # Already have base

                lang_info = get_language_info(lang_code)
                if not lang_info:
                    continue

                # Create language-specific description prompt
                translation_prompt = f"""Translate the following Claude Code description to {lang_info["native_name"]}.
Keep technical terms in English. Provide only the translation without explanation:

Original: {base_desc}

Translation:"""

                # Use Claude CLI for translation
                translation_result = self.process_template_command(
                    translation_prompt,
                    {"CONVERSATION_LANGUAGE": lang_code},
                    print_mode=True,
                    output_format="text",
                )

                if translation_result["success"]:
                    # Extract translation from output
                    translation = translation_result["stdout"].strip()
                    if translation:
                        multilingual[item_id][lang_code] = translation

        return multilingual

    def create_agent_with_multilingual_support(
        self,
        agent_name: str,
        base_description: str,
        tools: list[str],
        model: str = "sonnet",
        target_languages: Optional[list[str]] = None,
    ) -> Dict[str, Any]:
        """Create Claude agent with multilingual description support.

        Args:
            agent_name: Agent name (kebab-case)
            base_description: English base description
            tools: List of required tools
            model: Claude model to use
            target_languages: Target languages for descriptions

        Returns:
            Agent configuration dictionary
        """
        if target_languages is None:
            # System provides 4 languages: en, ko, ja, zh
            target_languages = ["en", "ko", "ja", "zh"]

        # Generate multilingual descriptions
        descriptions = self.generate_multilingual_descriptions({agent_name: base_description}, target_languages)

        agent_config = {
            "name": agent_name,
            "description": descriptions[agent_name]["en"],  # Primary English description
            "tools": tools,
            "model": model,
            "descriptions": descriptions[agent_name],  # All language versions
            "multilingual_support": True,
        }

        return agent_config

    def create_command_with_multilingual_support(
        self,
        command_name: str,
        base_description: str,
        argument_hint: list[str],
        tools: list[str],
        model: str = "haiku",
        target_languages: Optional[list[str]] = None,
    ) -> Dict[str, Any]:
        """Create Claude command with multilingual description support.

        Args:
            command_name: Command name (kebab-case)
            base_description: English base description
            argument_hint: List of argument hints
            tools: List of required tools
            model: Claude model to use
            target_languages: Target languages for descriptions

        Returns:
            Command configuration dictionary
        """
        if target_languages is None:
            # System provides 4 languages: en, ko, ja, zh
            target_languages = ["en", "ko", "ja", "zh"]

        # Generate multilingual descriptions
        descriptions = self.generate_multilingual_descriptions({command_name: base_description}, target_languages)

        command_config = {
            "name": command_name,
            "description": descriptions[command_name]["en"],  # Primary English description
            "argument-hint": argument_hint,
            "tools": tools,
            "model": model,
            "descriptions": descriptions[command_name],  # All language versions
            "multilingual_support": True,
        }

        return command_config

    def process_json_stream_input(
        self,
        input_data: Union[Dict[str, Any], str],
        variables: Optional[Dict[str, Any]] = None,
    ) -> Dict[str, Any]:
        """Process JSON stream input for Claude CLI.

        Args:
            input_data: JSON data as dict or JSON string
            variables: Additional variables for processing

        Returns:
            Processed input data
        """
        # Convert string to dict if needed
        processed_data: Dict[str, Any]
        if isinstance(input_data, str):
            try:
                processed_data = json.loads(input_data)
            except json.JSONDecodeError as e:
                raise ValueError(f"Invalid JSON input: {e}")
        else:
            processed_data = input_data

        if variables:
            # Process any string values in processed_data with variables
            for key, value in processed_data.items():
                if isinstance(value, str) and "{{" in value and "}}" in value:
                    processed_data[key] = self.template_engine.render_string(value, variables)

        return processed_data

    def execute_headless_command(
        self,
        prompt_template: str,
        variables: Dict[str, Any],
        input_format: str = "stream-json",
        output_format: str = "stream-json",
        additional_options: Optional[list[str]] = None,
    ) -> Dict[str, Any]:
        """Execute Claude command in headless mode with full variable processing.

        Args:
            prompt_template: Prompt template with variables
            variables: Variables for substitution
            input_format: Input format (text, stream-json)
            output_format: Output format (text, json, stream-json)
            additional_options: Additional CLI options

        Returns:
            Command execution result
        """
        try:
            # Process prompt template
            processed_prompt = self.template_engine.render_string(prompt_template, variables)

            # Build Claude command
            cmd_parts = ["claude", "--print"]
            cmd_parts.extend(["--input-format", input_format])
            cmd_parts.extend(["--output-format", output_format])

            # Add settings
            settings_file = self.generate_claude_settings(variables)
            cmd_parts.extend(["--settings", str(settings_file)])

            # Add additional options
            if additional_options:
                cmd_parts.extend(additional_options)

            # Add processed prompt
            cmd_parts.append(processed_prompt)

            # Execute with streaming support
            if output_format == "stream-json":
                # For streaming, use subprocess with real-time output
                process = subprocess.Popen(
                    cmd_parts,
                    stdout=subprocess.PIPE,
                    stderr=subprocess.PIPE,
                    text=True,
                    encoding="utf-8",
                    errors="replace",
                )

                stdout_lines: List[str] = []
                stderr_lines: List[str] = []

                # Stream output in real-time
                max_lines = 1000  # Limit to prevent memory exhaustion
                while True:
                    stdout_line = process.stdout.readline()
                    stderr_line = process.stderr.readline()

                    if stdout_line == "" and stderr_line == "" and process.poll() is not None:
                        break

                    if stdout_line:
                        if len(stdout_lines) < max_lines:
                            stdout_lines.append(stdout_line.strip())
                        # Truncate if limit reached

                    if stderr_line:
                        if len(stderr_lines) < max_lines:
                            stderr_lines.append(stderr_line.strip())
                        # Truncate if limit reached

                returncode = process.poll()

            else:
                # Non-streaming execution
                result = _safe_run_subprocess(
                    cmd_parts, timeout=60, max_output_size=1024 * 1024, max_lines=1000, capture_output=True
                )

                stdout_lines = result.stdout.strip().split("\n") if result.stdout else []
                stderr_lines = result.stderr.strip().split("\n") if result.stderr else []
                returncode = result.returncode

            # Cleanup
            try:
                settings_file.unlink()
            except OSError:
                pass

            return {
                "success": returncode == 0,
                "stdout": stdout_lines,
                "stderr": stderr_lines,
                "returncode": returncode,
                "processed_prompt": processed_prompt,
                "variables": variables,
            }

        except Exception as e:
            return {
                "success": False,
                "error": str(e),
                "prompt_template": prompt_template,
                "variables": variables,
            }
