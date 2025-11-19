"""Claude Code Headless-based Merge Analyzer

Analyzes template merge differences using Claude Code headless mode
for intelligent backup vs new template comparison and recommendations.
"""

import json
import subprocess
from difflib import unified_diff
from pathlib import Path
from typing import Any

import click
from rich.console import Console
from rich.live import Live
from rich.spinner import Spinner
from rich.table import Table
from rich.panel import Panel # Added for rich.Panel

console = Console()


class MergeAnalyzer:
    """
    Intelligent merge analyzer using Claude Code
    
    Analyzes the differences between the user's current project and the new template,
    and suggests an optimal merge strategy.
    """

    # List of key files to analyze
    ANALYZED_FILES = [
        "CLAUDE.md",
        ".claude/settings.json",
        ".moai/config/config.json",
        ".gitignore",
    ]

    # Claude headless execution settings
    CLAUDE_TIMEOUT = 120  # Max 2 minutes
    CLAUDE_MODEL = "claude-sonnet-4-5-20250929"  # Latest Sonnet
    CLAUDE_TOOLS = ["Read", "Glob", "Grep"]  # Read-only

    def __init__(self, project_path: Path):
        """Initialize analyzer with project path."""
        self.project_path = project_path

    def analyze_merge(
        self, backup_path: Path, template_path: Path
    ) -> dict[str, Any]:
        """
        Analyze merge conflicts and suggest strategy
        
        Args:
            backup_path: Path to the current project (backup)
            template_path: Path to the new template
            
        Returns:
            Analysis result dictionary
                - files: List of changes per file
                - safe_to_auto_merge: Whether auto-merge is safe
                - user_action_required: Whether user intervention is needed
                - summary: Overall summary
                - error: Error message (if any)
        """
        # 1. Collect files to compare
        diff_files = self._collect_diff_files(backup_path, template_path)
        diff_summary = self._format_diff_summary(diff_files) # Prepare diff summary for prompt

        # 2. Generate Claude headless prompt
        prompt = self._create_analysis_prompt(diff_summary)

        # 3. Execute Claude Code headless (show spinner)
        spinner = Spinner("dots", text="[cyan]Claude Code analysis in progress...[/cyan]")

        try:
            with Live(spinner, refresh_per_second=12):
                # Use headless mode to get JSON output
                cmd = [
                    "claude",
                    "code",
                    "--print",  # Output to stdout
                    prompt
                ]
                result = subprocess.run(
                    cmd,
                    capture_output=True,
                    text=True,
                    timeout=self.CLAUDE_TIMEOUT,
                )

            output = result.stdout # Capture output for parsing

            if result.returncode == 0:
                # 4. Parse result
                try:
                    # Find JSON block
                    json_match = re.search(r'```json\s*({.*?})\s*```', output, re.DOTALL)
                    if json_match:
                        analysis_json = json_match.group(1)
                        analysis_result = json.loads(analysis_json)
                    else:
                        # If no JSON block, try parsing the whole output
                        analysis_result = json.loads(output)

                    console.print("[green]✅ Analysis complete[/green]")
                    return analysis_result
                except json.JSONDecodeError as e:
                    console.print(
                        f"[yellow]⚠️  Claude response parsing error: {e}[/yellow]"
                    )
                    # Fallback if JSON parsing fails
                    return {
                        "summary": "Failed to parse analysis result from Claude.",
                        "risk_level": "high",
                        "conflicts": [],
                        "recommendation": "manual",
                        "error": str(e),
                        "raw_output": output # Include raw output for debugging
                    }
            else:
                console.print(
                    f"[yellow]⚠️  Claude execution error: {result.stderr[:200]}[/yellow]"
                )
                return self._fallback_analysis(
                    backup_path, template_path, diff_files
                )

        except subprocess.TimeoutExpired:
            console.print(
                "[yellow]⚠️  Claude analysis timeout (exceeded 120 seconds)[/yellow]"
            )
            return self._fallback_analysis(
                backup_path, template_path, diff_files
            )
        except FileNotFoundError:
            console.print(
                "[red]❌ Claude Code not found.[/red]"
            )
            console.print(
                "[cyan]   Install Claude Code: https://claude.com/claude-code[/cyan]"
            )
            return self._fallback_analysis(
                backup_path, template_path, diff_files
            )

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
                "\n⚠️  User intervention is required. Please review the following:",
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

    def _collect_diff_files(
        self, backup_path: Path, template_path: Path
    ) -> dict[str, dict[str, Any]]:
        """Collect differing files between backup and template

        Returns:
            Dictionary of diff information per file
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
        diff_summary: str,
    ) -> str:
        """
        Generate prompt for Claude Code
        """
        return f"""
You are an expert in project file merging.
Analyze the differences between the current project and the new template to suggest a merge strategy.

Diff summary:
{diff_summary}

Please analyze in the following format and output ONLY JSON:

{{
  "summary": "Brief summary of changes (1-2 sentences)",
  "risk_level": "low|medium|high",
  "conflicts": [
    {{
      "file": "File path",
      "type": "modify|delete|create",
      "description": "Description of change",
      "recommendation": "keep_current|use_template|merge"
    }}
  ],
  "recommendation": "auto|manual"
}}

"risk_level" criteria:
- low: Only new files added or simple config changes
- medium: Code logic changes or config structure changes
- high: User custom code deletion risk or complex conflicts

"recommendation" criteria:
- auto: Low risk, safe to overwrite
- manual: High risk, user verification needed
"""

    def _display_analysis(self, analysis: dict[str, Any]) -> None:
        """
        Display analysis result to user
        """
        # 1. Summary and Risk Level
        risk_color = {
            "low": "green",
            "medium": "yellow",
            "high": "red"
        }.get(analysis.get("risk_level", "high"), "red")

        console.print(Panel(
            f"[bold]Analysis Summary:[/bold] {analysis.get('summary')}\n"
            f"[bold]Risk Level:[/bold] [{risk_color}]{analysis.get('risk_level', 'unknown').upper()}[/{risk_color}]",
            title="📋 Merge Analysis Result",
            border_style=risk_color
        ))

        # File-specific changes table
        if analysis.get("files"): # Original code used 'files', new uses 'conflicts'
            table = Table(title="File-specific Changes")
            table.add_column("File", style="cyan")
            table.add_column("Changes", style="white")
            table.add_column("Recommendation", style="yellow")
            table.add_column("Severity", style="red")

            for file_info in analysis["files"]: # Assuming 'files' key is still used for display
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

            # Additional notes
            for file_info in analysis["files"]:
                if file_info.get("note"):
                    console.print(
                        f"\n💡 {file_info['filename']}: {file_info['note']}",
                        style="dim",
                    )

    def _build_claude_command(self) -> list[str]:
        """Build Claude Code headless command"""
        # This method is now effectively replaced by the direct `cmd` construction in analyze_merge
        # but keeping it for consistency if other parts still call it.
        return [
            "claude",
            "-p",
            "--output-format",
            "json",
            "--tools",
            ",".join(self.CLAUDE_TOOLS),
        ]

    def _format_diff_summary(
        self, diff_files: dict[str, dict[str, Any]]
    ) -> str:
        """Format diff_files into a prompt-friendly string"""
        summary = []
        for file_name, info in diff_files.items():
            if info["backup_exists"] and info["template_exists"]:
                status = (
                    f"✏️  Modified ({info['diff_lines']} lines)"
                    if info["has_diff"]
                    else "✓ Identical"
                )
            elif info["backup_exists"]:
                status = "❌ Deleted from template"
            else:
                status = "✨ New file (template)"

            summary.append(f"- {file_name}: {status}")

        return "\n".join(summary)

    def _fallback_analysis(
        self,
        backup_path: Path,
        template_path: Path,
        diff_files: dict[str, dict[str, Any]],
    ) -> dict[str, Any]:
        """Fallback analysis if Claude call fails (difflib-based)

        Returns basic analysis results when Claude is unavailable
        """
        console.print(
            "⚠️  Claude Code is unavailable. Using basic analysis.",
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

            files_analysis.append({
                "filename": file_name,
                "changes": f"{info['diff_lines']} lines changed",
                "recommendation": "smart_merge",
                "conflict_severity": severity,
            })

            if severity == "high":
                has_high_risk = True

        return {
            "files": files_analysis,
            "safe_to_auto_merge": not has_high_risk,
            "user_action_required": has_high_risk,
            "summary": f"{len(files_analysis)} files with changes detected (basic analysis)",
            "risk_assessment": "High - Claude analysis unavailable, manual review recommended" if has_high_risk else "Low",
            "fallback": True,
        }

