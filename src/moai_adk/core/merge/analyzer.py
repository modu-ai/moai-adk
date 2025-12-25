"""Claude Code Headless-based Merge Analyzer

Analyzes template merge differences using Claude Code headless mode
for intelligent backup vs new template comparison and recommendations.
"""

import json
import logging
import os
import re
import shutil
import subprocess
import sys
from difflib import unified_diff
from pathlib import Path
from typing import Any, Optional

import click
from rich.console import Console
from rich.live import Live
from rich.spinner import Spinner
from rich.table import Table

console = Console()
logger = logging.getLogger(__name__)


class MergeAnalyzer:
    """Merge analyzer using Claude Code for intelligent template merge analysis

    Compares backed-up user configurations with new templates,
    analyzes them using Claude AI, and provides merge recommendations.
    """

    # Primary files to analyze
    ANALYZED_FILES = [
        "CLAUDE.md",
        ".claude/settings.json",
        ".moai/config/config.yaml",  # Updated: JSON → YAML migration
        ".gitignore",
    ]

    # Claude headless execution settings
    CLAUDE_TIMEOUT = 120  # Maximum 2 minutes
    CLAUDE_MODEL = "claude-haiku-4-5-20251001"  # Latest Haiku (cost optimized)
    CLAUDE_TOOLS = ["Read", "Glob", "Grep"]  # Read-only tools

    def __init__(self, project_path: Path):
        """Initialize analyzer with project path."""
        self.project_path = project_path

    def analyze_merge(self, backup_path: Path, template_path: Path) -> dict[str, Any]:
        """Perform merge analysis using Claude Code headless mode

        Args:
            backup_path: Path to backed-up configuration directory
            template_path: Path to new template directory

        Returns:
            Dictionary containing analysis results:
                - files: List of changes by file
                - safe_to_auto_merge: Whether auto-merge is safe
                - user_action_required: Whether user intervention is needed
                - summary: Overall summary
                - error: Error message (if any)
        """
        # 1. Collect files to compare
        diff_files = self._collect_diff_files(backup_path, template_path)

        # 2. Create Claude headless prompt
        prompt = self._create_analysis_prompt(backup_path, template_path, diff_files)

        # 3. Run Claude Code headless (show spinner)
        spinner = Spinner("dots", text="[cyan]Running Claude Code analysis...[/cyan]")

        try:
            with Live(spinner, refresh_per_second=12):
                result = subprocess.run(
                    self._build_claude_command(),
                    input=prompt,
                    capture_output=True,
                    text=True,
                    timeout=self.CLAUDE_TIMEOUT,
                )

            if result.returncode == 0:
                # Use new response parsing method
                analysis = self._parse_claude_response(result.stdout)
                if "error" not in analysis:
                    console.print("[green]✅ Analysis complete[/green]")
                    return analysis
                else:
                    console.print(f"[yellow]⚠️  Analysis failed: {analysis.get('summary', 'Unknown error')}[/yellow]")
                    return self._fallback_analysis(backup_path, template_path, diff_files)
            else:
                # Use improved error detection with full context
                error_msg = self._detect_claude_errors(
                    result.stderr,
                    returncode=result.returncode,
                    stdout=result.stdout,
                )
                logger.warning(
                    f"Claude Code failed: returncode={result.returncode}, "
                    f"stderr={result.stderr[:200] if result.stderr else 'empty'}, "
                    f"stdout_hint={result.stdout[:100] if result.stdout else 'empty'}"
                )
                console.print(f"[yellow]⚠️  Claude execution error: {error_msg}[/yellow]")
                return self._fallback_analysis(backup_path, template_path, diff_files)

        except subprocess.TimeoutExpired:
            console.print("[yellow]⚠️  Claude analysis timeout (exceeded 120 seconds)[/yellow]")
            return self._fallback_analysis(backup_path, template_path, diff_files)
        except FileNotFoundError:
            console.print("[red]❌ Claude Code not found.[/red]")
            console.print("[cyan]   Install Claude Code: https://claude.com/claude-code[/cyan]")
            return self._fallback_analysis(backup_path, template_path, diff_files)

    def ask_user_confirmation(self, analysis: dict[str, Any]) -> bool:
        """Display analysis results and request user confirmation

        Args:
            analysis: Result from analyze_merge()

        Returns:
            True: Proceed, False: Cancel
        """
        # 1. Display analysis results
        self._display_analysis(analysis)

        # 2. User confirmation
        if analysis.get("user_action_required", False):
            console.print(
                "\n⚠️  User intervention required. Please review the following:",
                style="warning",
            )
            for file_info in analysis.get("files", []):
                if file_info.get("conflict_severity") in ["medium", "high"]:
                    console.print(
                        f"   • {file_info['filename']}: {file_info.get('note', '')}",
                    )

        # 3. Confirmation prompt
        proceed = click.confirm(
            "\nProceed with merge?",
            default=analysis.get("safe_to_auto_merge", False),
        )

        return proceed

    def _collect_diff_files(self, backup_path: Path, template_path: Path) -> dict[str, dict[str, Any]]:
        """Collect differences between backup and template files

        Returns:
            Dictionary with diff information per file
        """
        diff_files = {}

        for file_name in self.ANALYZED_FILES:
            backup_file = backup_path / file_name
            template_file = template_path / file_name

            if not backup_file.exists() and not template_file.exists():
                continue

            diff_info = {
                "backup_exists": backup_file.exists(),
                "template_exists": template_file.exists(),
                "has_diff": False,
                "diff_lines": 0,
            }

            if backup_file.exists() and template_file.exists():
                backup_content = backup_file.read_text(encoding="utf-8")
                template_content = template_file.read_text(encoding="utf-8")

                if backup_content != template_content:
                    diff = list(
                        unified_diff(
                            backup_content.splitlines(),
                            template_content.splitlines(),
                            lineterm="",
                        )
                    )
                    diff_info["has_diff"] = True
                    diff_info["diff_lines"] = len(diff)

            diff_files[file_name] = diff_info

        return diff_files

    def _create_analysis_prompt(
        self,
        backup_path: Path,
        template_path: Path,
        diff_files: dict[str, dict[str, Any]],
    ) -> str:
        """Generate Claude headless analysis prompt

        Returns:
            Analysis prompt to send to Claude
        """
        return f"""You are a MoAI-ADK configuration file merge expert.

## Context
- Backed-up user configuration: {backup_path}
- New template: {template_path}
- Files to analyze: {", ".join(self.ANALYZED_FILES)}

## Files to Analyze
{self._format_diff_summary(diff_files)}

## Analysis Tasks
Analyze the following items and provide a JSON response:

1. Identify changes per file
2. Assess conflict risk (low/medium/high)
3. Merge recommendations (use_template/keep_existing/smart_merge)
4. Overall safety assessment

## Response Format (JSON)
{{
  "files": [
    {{
      "filename": "CLAUDE.md",
      "changes": "Description of changes",
      "recommendation": "use_template|keep_existing|smart_merge",
      "conflict_severity": "low|medium|high",
      "note": "Additional notes (optional)"
    }}
  ],
  "safe_to_auto_merge": true/false,
  "user_action_required": true/false,
  "summary": "Whether merge is safe and why",
  "risk_assessment": "Risk assessment"
}}

## Merge Rules Reference
- CLAUDE.md: Preserve Project Information section
- settings.json: Merge env variables, prioritize template permissions.deny
- config.json: Preserve user metadata, update schema
- .gitignore: Additions only (preserve existing items)

## Additional Considerations
- Assess risk of user customization loss
- Determine if force overwriting Alfred infrastructure files
- Review rollback possibilities
"""

    def _display_analysis(self, analysis: dict[str, Any]) -> None:
        """Display analysis results in Rich format"""
        # Title
        console.print("\n📊 Merge Analysis Results (Claude Code)", style="bold")

        # Summary
        summary = analysis.get("summary", "No analysis results")
        console.print(f"\n📝 {summary}")

        # Risk assessment
        risk_assessment = analysis.get("risk_assessment", "")
        if risk_assessment:
            risk_style = "green" if "safe" in risk_assessment.lower() else "yellow"
            console.print(f"⚠️  Risk Level: {risk_assessment}", style=risk_style)

        # Changes by file table
        files_list = analysis.get("files")
        if files_list and isinstance(files_list, list):
            table = Table(title="Changes by File")
            table.add_column("File", style="cyan")
            table.add_column("Changes", style="white")
            table.add_column("Recommendation", style="yellow")
            table.add_column("Risk", style="red")

            for file_info in files_list:
                # Ensure file_info is a dictionary
                if not isinstance(file_info, dict):
                    continue

                severity_style = {
                    "low": "green",
                    "medium": "yellow",
                    "high": "red",
                }.get(file_info.get("conflict_severity", "low"), "white")

                table.add_row(
                    file_info.get("filename", "?"),
                    file_info.get("changes", "")[:30],
                    file_info.get("recommendation", "?"),
                    file_info.get("conflict_severity", "?"),
                    style=severity_style,
                )

            console.print(table)

            # Additional details
            for file_info in files_list:
                # Ensure file_info is a dictionary
                if not isinstance(file_info, dict):
                    continue

                if file_info.get("note"):
                    console.print(
                        f"\n💡 {file_info['filename']}: {file_info['note']}",
                        style="dim",
                    )

    def _parse_claude_response(self, response_text: str) -> dict[str, Any]:
        """Parse Claude Code response supporting both v1.x and v2.0+ formats.

        Args:
            response_text: Raw response text from Claude Code

        Returns:
            Parsed analysis dictionary
        """
        try:
            # First try direct JSON parsing (v1.x format)
            return json.loads(response_text)
        except json.JSONDecodeError:
            # Try v2.0+ wrapped format
            try:
                # Look for JSON in the response
                if '"type":' in response_text and '"result":' in response_text:
                    # Parse the wrapped v2.0+ format
                    response_obj = json.loads(response_text)
                    if "result" in response_obj:
                        result_text = response_obj["result"]

                        # Try to extract JSON from the result field
                        if isinstance(result_text, str):
                            # Look for JSON blocks in the result
                            if "```json" in result_text:
                                # Extract JSON from code block
                                start = result_text.find("```json") + 7
                                end = result_text.find("```", start)
                                if end != -1:
                                    json_text = result_text[start:end].strip()
                                    return json.loads(json_text)
                            elif result_text.strip().startswith("{"):
                                # Try direct JSON parsing
                                return json.loads(result_text)
                            else:
                                # Try to find JSON pattern in text
                                json_match = re.search(r"\{[^{}]*(?:\{[^{}]*\}[^{}]*)*\}", result_text)
                                if json_match:
                                    try:
                                        return json.loads(json_match.group(0))
                                    except json.JSONDecodeError:
                                        pass

                # Fallback: try to find any JSON in the text
                json_match = re.search(r"\{[^{}]*(?:\{[^{}]*\}[^{}]*)*\}", response_text)
                if json_match:
                    return json.loads(json_match.group(0))

            except (json.JSONDecodeError, KeyError, TypeError) as e:
                console.print(f"[yellow]⚠️  Failed to parse Claude v2.0+ response: {e}[/yellow]")
                logger.warning(f"Claude response parsing failed: {e}")

        # If all parsing attempts fail, return error structure
        logger.error(f"Could not parse Claude response. Raw response: {response_text[:500]}...")
        return {
            "files": [],
            "safe_to_auto_merge": False,
            "user_action_required": True,
            "summary": "Failed to parse Claude response",
            "risk_assessment": "High - Response parsing failed",
            "error": "response_parse_failed",
            "raw_response": response_text[:500] if response_text else "",
        }

    def _detect_claude_errors(
        self,
        stderr: str,
        returncode: int = None,
        stdout: str = None,
    ) -> str:
        """Detect and interpret Claude Code specific errors.

        Args:
            stderr: Standard error output from Claude Code
            returncode: Process exit code (optional, for better diagnostics)
            stdout: Standard output (optional, may contain error hints)

        Returns:
            User-friendly error message with actionable guidance
        """
        # Case 1: Process failed but stderr is empty (silent failure)
        if not stderr and returncode is not None and returncode != 0:
            # Try to extract hints from stdout
            if stdout:
                stdout_lower = stdout.lower()
                if "error" in stdout_lower or "failed" in stdout_lower or "exception" in stdout_lower:
                    stdout_hint = stdout[:300] if len(stdout) > 300 else stdout
                    return f"Process failed (exit code {returncode}). Output: {stdout_hint}"
            return (
                f"Process failed silently (exit code {returncode}). "
                "No error details available. Try 'claude --version' to verify installation."
            )

        # Case 2: No stderr and no returncode context
        if not stderr:
            return "No error details available. Try running 'claude --version' to verify installation."

        error_lower = stderr.lower()

        # Pattern matching for specific known errors
        if "model not found" in error_lower or "unknown model" in error_lower:
            return f"Claude model '{self.CLAUDE_MODEL}' not found. Run 'claude --models' to see available models."

        if "permission denied" in error_lower:
            return "Permission denied. Check file permissions and Claude Code access rights."

        if "timeout" in error_lower:
            return f"Claude analysis timed out after {self.CLAUDE_TIMEOUT} seconds. Consider increasing timeout."

        if "file not found" in error_lower:
            return "Required files not found. Check project structure and file paths."

        if "invalid argument" in error_lower or "unknown option" in error_lower:
            return "Invalid Claude Code arguments. This might be a version compatibility issue. Try 'claude --help'."

        if "api key" in error_lower or "authentication" in error_lower or "unauthorized" in error_lower:
            return "Authentication error. Check your Claude API key configuration."

        if "rate limit" in error_lower or "too many requests" in error_lower:
            return "Rate limit exceeded. Please wait a moment and try again."

        if "connection" in error_lower or "network" in error_lower:
            return "Network connection error. Check your internet connection."

        # Return generic error with exit code if available (extended to 300 chars)
        exit_info = f" (exit code {returncode})" if returncode is not None else ""
        return f"Claude Code error{exit_info}: {stderr[:300]}"

    def _find_claude_executable(self) -> Optional[str]:
        """Find Claude Code executable path with Windows compatibility.

        Searches for Claude Code in:
        1. System PATH (shutil.which)
        2. Windows npm global directory
        3. Windows local AppData directory
        4. Common installation paths

        Returns:
            Full path to Claude executable or None if not found
        """
        # First try system PATH
        claude_path = shutil.which("claude")
        if claude_path:
            return claude_path

        # Windows-specific additional paths
        if sys.platform == "win32":
            possible_paths = []

            # npm global installation
            appdata = os.environ.get("APPDATA", "")
            if appdata:
                possible_paths.extend(
                    [
                        Path(appdata) / "npm" / "claude.cmd",
                        Path(appdata) / "npm" / "claude.exe",
                        Path(appdata) / "npm" / "claude",
                    ]
                )

            # Local AppData installation
            localappdata = os.environ.get("LOCALAPPDATA", "")
            if localappdata:
                possible_paths.extend(
                    [
                        Path(localappdata) / "Programs" / "claude" / "claude.exe",
                        Path(localappdata)
                        / "Microsoft"
                        / "WinGet"
                        / "Packages"
                        / "Anthropic.ClaudeCode_*"
                        / "claude.exe",
                    ]
                )

            # User profile paths
            userprofile = os.environ.get("USERPROFILE", "")
            if userprofile:
                possible_paths.extend(
                    [
                        Path(userprofile) / ".claude" / "claude.exe",
                        Path(userprofile) / "AppData" / "Local" / "Programs" / "claude" / "claude.exe",
                    ]
                )

            # Check each possible path
            for p in possible_paths:
                # Handle glob patterns
                if "*" in str(p):
                    parent = p.parent
                    pattern = p.name
                    if parent.exists():
                        matches = list(parent.glob(pattern))
                        if matches:
                            return str(matches[0])
                elif p.exists():
                    return str(p)

        return None

    def _build_claude_command(self) -> list[str]:
        """Build Claude Code headless command (based on official v4.0+)

        Claude Code CLI official options:
        - -p: Non-interactive headless mode
        - --model: Explicit model selection (Haiku)
        - --output-format: JSON response format
        - --tools: Read-only tools only (space-separated - POSIX standard)
        - --permission-mode: Auto-approval (background task)

        Returns:
            List of Claude CLI command arguments

        Raises:
            FileNotFoundError: If Claude Code executable is not found
        """
        # Find Claude executable with Windows compatibility
        claude_path = self._find_claude_executable()
        if not claude_path:
            raise FileNotFoundError("Claude Code executable not found")

        # Tools list space-separated (POSIX standard, officially recommended)
        tools_str = " ".join(self.CLAUDE_TOOLS)

        return [
            claude_path,  # Use full path instead of just "claude"
            "-p",  # Non-interactive headless mode
            "--model",
            self.CLAUDE_MODEL,  # Explicit model specification (Haiku)
            "--output-format",
            "json",  # Single JSON response
            "--tools",
            tools_str,  # Space-separated (Read Glob Grep)
            "--permission-mode",
            "dontAsk",  # Auto-approval (safe, read-only)
        ]

    def _format_diff_summary(self, diff_files: dict[str, dict[str, Any]]) -> str:
        """Format diff_files for prompt"""
        summary = []
        for file_name, info in diff_files.items():
            if info["backup_exists"] and info["template_exists"]:
                status = f"✏️  Modified ({info['diff_lines']} lines)" if info["has_diff"] else "✓ Identical"
            elif info["backup_exists"]:
                status = "❌ Deleted from template"
            else:
                status = "✨ New file (from template)"

            summary.append(f"- {file_name}: {status}")

        return "\n".join(summary)

    def _fallback_analysis(
        self,
        backup_path: Path,
        template_path: Path,
        diff_files: dict[str, dict[str, Any]],
    ) -> dict[str, Any]:
        """Fallback analysis when Claude call fails (difflib-based)

        Returns basic analysis results when Claude is unavailable
        """
        console.print(
            "⚠️  Claude Code unavailable. Using fallback analysis.",
            style="yellow",
        )

        files_analysis = []
        has_high_risk = False

        for file_name, info in diff_files.items():
            if not info["has_diff"]:
                continue

            # Simple risk assessment
            severity = "low"
            if file_name in [".claude/settings.json", ".moai/config/config.json"]:
                severity = "medium" if info["diff_lines"] > 10 else "low"

            files_analysis.append(
                {
                    "filename": file_name,
                    "changes": f"{info['diff_lines']} lines changed",
                    "recommendation": "smart_merge",
                    "conflict_severity": severity,
                }
            )

            if severity == "high":
                has_high_risk = True

        return {
            "files": files_analysis,
            "safe_to_auto_merge": not has_high_risk,
            "user_action_required": has_high_risk,
            "summary": f"{len(files_analysis)} files changed (fallback analysis)",
            "risk_assessment": ("High - Claude unavailable, manual review recommended" if has_high_risk else "Low"),
            "fallback": True,
        }
