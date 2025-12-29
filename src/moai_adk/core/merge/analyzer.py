"""Pure Python Merge Analyzer

Analyzes template merge differences using semantic heuristics
for intelligent backup vs new template comparison and recommendations.

This module replaces the previous Claude Code headless implementation
with a pure Python approach that works without API keys.
"""

import json
import logging
import re
from difflib import unified_diff
from pathlib import Path
from typing import Any

import click
import yaml
from rich.console import Console
from rich.table import Table

console = Console()
logger = logging.getLogger(__name__)


class MergeAnalyzer:
    """Merge analyzer using semantic heuristics for intelligent template merge analysis

    Compares backed-up user configurations with new templates,
    analyzes them using Python-based semantic analysis, and provides merge recommendations.
    """

    # Primary files to analyze
    ANALYZED_FILES = [
        "CLAUDE.md",
        ".claude/settings.json",
        ".moai/config/config.yaml",
        ".gitignore",
    ]

    # Risk factors for scoring
    RISK_FACTORS = {
        "user_section_modified": 3,
        "permission_change": 2,
        "env_variable_removed": 2,
        "large_diff": 1,
        "config_key_removed": 2,
        "breaking_schema_change": 3,
        "custom_content_loss": 3,
    }

    # User-customizable sections in CLAUDE.md (should be preserved)
    CLAUDE_MD_USER_SECTIONS = [
        "Project Information",
        "User Personalization",
        "Custom Rules",
        "Local Configuration",
    ]

    # Critical settings.json keys that require careful handling
    SETTINGS_CRITICAL_KEYS = [
        "permissions.deny",
        "permissions.allow",
        "env",
        "hooks",
    ]

    def __init__(self, project_path: Path):
        """Initialize analyzer with project path."""
        self.project_path = project_path

    def analyze_merge(self, backup_path: Path, template_path: Path) -> dict[str, Any]:
        """Perform merge analysis using semantic heuristics

        Args:
            backup_path: Path to backed-up configuration directory
            template_path: Path to new template directory

        Returns:
            Dictionary containing analysis results
        """
        # 1. Collect files to compare
        diff_files = self._collect_diff_files(backup_path, template_path)

        # 2. Perform semantic analysis for each file
        files_analysis = []
        total_risk_score = 0

        for file_name, info in diff_files.items():
            if not info["has_diff"] and info["backup_exists"] and info["template_exists"]:
                # No changes, skip
                continue

            # Analyze based on file type
            analysis = self._analyze_file_semantic(file_name, backup_path, template_path, info)
            files_analysis.append(analysis)
            total_risk_score += analysis.get("risk_score", 0)

        # 3. Determine overall safety
        risk_level = self._calculate_risk_level(total_risk_score)
        safe_to_merge = risk_level == "low"
        user_action_required = risk_level == "high"

        return {
            "files": files_analysis,
            "safe_to_auto_merge": safe_to_merge,
            "user_action_required": user_action_required,
            "summary": f"{len(files_analysis)} files need attention",
            "risk_assessment": risk_level.capitalize(),
            "total_risk_score": total_risk_score,
            "analysis_method": "pure_python_semantic",
        }

    def ask_user_confirmation(self, analysis: dict[str, Any]) -> bool:
        """Display analysis results and request user confirmation

        Args:
            analysis: Result from analyze_merge()

        Returns:
            True: Proceed, False: Cancel
        """
        # 1. Display analysis results
        self._display_analysis(analysis)

        # 2. Show warnings for high-risk items
        if analysis.get("user_action_required", False):
            console.print(
                "\n[yellow]User intervention may be required. Please review:[/yellow]",
            )
            for file_info in analysis.get("files", []):
                if file_info.get("conflict_severity") in ["medium", "high"]:
                    console.print(
                        f"   [yellow]{file_info['filename']}: {file_info.get('note', '')}[/yellow]",
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

            # Read file contents
            backup_content: str | None = None
            template_content: str | None = None

            if backup_file.exists():
                try:
                    backup_content = backup_file.read_text(encoding="utf-8")
                except Exception as e:
                    logger.warning(f"Failed to read backup file {backup_file}: {e}")

            if template_file.exists():
                try:
                    template_content = template_file.read_text(encoding="utf-8")
                except Exception as e:
                    logger.warning(f"Failed to read template file {template_file}: {e}")

            # Calculate diff
            has_diff = False
            diff_lines = 0

            if backup_content and template_content:
                if backup_content != template_content:
                    diff = list(
                        unified_diff(
                            backup_content.splitlines(),
                            template_content.splitlines(),
                            lineterm="",
                        )
                    )
                    has_diff = True
                    diff_lines = len(diff)

            diff_info: dict[str, Any] = {
                "backup_exists": backup_file.exists(),
                "template_exists": template_file.exists(),
                "has_diff": has_diff,
                "diff_lines": diff_lines,
                "backup_content": backup_content,
                "template_content": template_content,
            }

            diff_files[file_name] = diff_info

        return diff_files

    def _analyze_file_semantic(
        self,
        file_name: str,
        backup_path: Path,
        template_path: Path,
        info: dict[str, Any],
    ) -> dict[str, Any]:
        """Perform semantic analysis based on file type

        Returns:
            Analysis result for the specific file
        """
        # Route to appropriate analyzer based on file type
        if file_name == "CLAUDE.md":
            return self._analyze_claude_md(file_name, info)
        elif file_name == ".claude/settings.json":
            return self._analyze_settings_json(file_name, info)
        elif file_name == ".moai/config/config.yaml":
            return self._analyze_config_yaml(file_name, info)
        elif file_name == ".gitignore":
            return self._analyze_gitignore(file_name, info)
        else:
            return self._analyze_generic(file_name, info)

    def _analyze_claude_md(self, file_name: str, info: dict[str, Any]) -> dict[str, Any]:
        """Analyze CLAUDE.md with section-aware logic

        Detects user-customized sections and determines merge strategy.
        """
        risk_score = 0
        notes = []
        recommendation = "use_template"

        backup_content = info.get("backup_content", "")
        template_content = info.get("template_content", "")

        if not info["backup_exists"]:
            return {
                "filename": file_name,
                "changes": "New file from template",
                "recommendation": "use_template",
                "conflict_severity": "low",
                "risk_score": 0,
                "note": "New file will be created",
            }

        if not info["template_exists"]:
            return {
                "filename": file_name,
                "changes": "File removed in template",
                "recommendation": "keep_existing",
                "conflict_severity": "medium",
                "risk_score": self.RISK_FACTORS["custom_content_loss"],
                "note": "Template no longer includes this file",
            }

        # Check for user-customized sections
        user_customizations = []
        for section in self.CLAUDE_MD_USER_SECTIONS:
            pattern = rf"#+\s*{re.escape(section)}"
            backup_has = bool(re.search(pattern, backup_content, re.IGNORECASE))
            template_has = bool(re.search(pattern, template_content, re.IGNORECASE))

            if backup_has and not template_has:
                user_customizations.append(section)
                risk_score += self.RISK_FACTORS["user_section_modified"]

        # Analyze diff size
        if info["diff_lines"] > 100:
            risk_score += self.RISK_FACTORS["large_diff"]
            notes.append("Large changes detected")

        # Determine recommendation
        if user_customizations:
            recommendation = "smart_merge"
            notes.append(f"User sections detected: {', '.join(user_customizations)}")
        elif info["diff_lines"] > 50:
            recommendation = "smart_merge"
        else:
            recommendation = "use_template"

        severity = self._risk_score_to_severity(risk_score)

        return {
            "filename": file_name,
            "changes": f"{info['diff_lines']} lines changed",
            "recommendation": recommendation,
            "conflict_severity": severity,
            "risk_score": risk_score,
            "note": "; ".join(notes) if notes else "Standard template update",
            "user_customizations": user_customizations,
        }

    def _analyze_settings_json(self, file_name: str, info: dict[str, Any]) -> dict[str, Any]:
        """Analyze settings.json with JSON structure comparison

        Focuses on permissions, env variables, and hooks.
        """
        risk_score = 0
        notes = []
        recommendation = "use_template"

        if not info["backup_exists"]:
            return {
                "filename": file_name,
                "changes": "New file from template",
                "recommendation": "use_template",
                "conflict_severity": "low",
                "risk_score": 0,
                "note": "New settings file will be created",
            }

        if not info["template_exists"]:
            return {
                "filename": file_name,
                "changes": "File removed in template",
                "recommendation": "keep_existing",
                "conflict_severity": "high",
                "risk_score": self.RISK_FACTORS["breaking_schema_change"],
                "note": "Critical: settings.json removed from template",
            }

        # Parse JSON
        try:
            backup_json = json.loads(info["backup_content"]) if info["backup_content"] else {}
            template_json = json.loads(info["template_content"]) if info["template_content"] else {}
        except json.JSONDecodeError as e:
            logger.warning(f"JSON parse error in {file_name}: {e}")
            return {
                "filename": file_name,
                "changes": "JSON parse error",
                "recommendation": "smart_merge",
                "conflict_severity": "high",
                "risk_score": self.RISK_FACTORS["breaking_schema_change"],
                "note": f"JSON parse error: {e}",
            }

        # Compare critical keys
        critical_changes = []

        # Check env variables
        backup_env = backup_json.get("env", {})
        template_env = template_json.get("env", {})

        removed_env = set(backup_env.keys()) - set(template_env.keys())
        if removed_env:
            critical_changes.append(f"Env vars removed: {', '.join(removed_env)}")
            risk_score += self.RISK_FACTORS["env_variable_removed"] * len(removed_env)

        # Check permissions
        backup_deny = set(backup_json.get("permissions", {}).get("deny", []))
        template_deny = set(template_json.get("permissions", {}).get("deny", []))

        if backup_deny != template_deny:
            critical_changes.append("Permission deny list changed")
            risk_score += self.RISK_FACTORS["permission_change"]

        # Check hooks
        backup_hooks = backup_json.get("hooks", {})
        template_hooks = template_json.get("hooks", {})

        if backup_hooks != template_hooks:
            critical_changes.append("Hooks configuration changed")
            risk_score += 1

        # Determine recommendation
        if critical_changes:
            recommendation = "smart_merge"
            notes.extend(critical_changes)
        elif info["diff_lines"] > 20:
            recommendation = "smart_merge"
        else:
            recommendation = "use_template"

        severity = self._risk_score_to_severity(risk_score)

        return {
            "filename": file_name,
            "changes": f"{info['diff_lines']} lines changed",
            "recommendation": recommendation,
            "conflict_severity": severity,
            "risk_score": risk_score,
            "note": "; ".join(notes) if notes else "Settings update available",
            "critical_changes": critical_changes,
        }

    def _analyze_config_yaml(self, file_name: str, info: dict[str, Any]) -> dict[str, Any]:
        """Analyze config.yaml with YAML structure comparison

        Preserves user settings while updating schema.
        """
        risk_score = 0
        notes = []
        recommendation = "use_template"

        if not info["backup_exists"]:
            return {
                "filename": file_name,
                "changes": "New file from template",
                "recommendation": "use_template",
                "conflict_severity": "low",
                "risk_score": 0,
                "note": "New config file will be created",
            }

        if not info["template_exists"]:
            return {
                "filename": file_name,
                "changes": "File removed in template",
                "recommendation": "keep_existing",
                "conflict_severity": "medium",
                "risk_score": self.RISK_FACTORS["config_key_removed"],
                "note": "Config file removed from template",
            }

        # Parse YAML
        try:
            backup_yaml = yaml.safe_load(info["backup_content"]) if info["backup_content"] else {}
            template_yaml = yaml.safe_load(info["template_content"]) if info["template_content"] else {}

            # Handle None results
            backup_yaml = backup_yaml or {}
            template_yaml = template_yaml or {}
        except yaml.YAMLError as e:
            logger.warning(f"YAML parse error in {file_name}: {e}")
            return {
                "filename": file_name,
                "changes": "YAML parse error",
                "recommendation": "smart_merge",
                "conflict_severity": "high",
                "risk_score": self.RISK_FACTORS["breaking_schema_change"],
                "note": f"YAML parse error: {e}",
            }

        # User-specific keys to preserve
        user_keys = ["user", "language", "project"]
        user_customizations = []

        for key in user_keys:
            backup_val = backup_yaml.get(key)
            template_val = template_yaml.get(key)

            if backup_val and backup_val != template_val:
                user_customizations.append(key)

        # Check for removed keys
        backup_keys = set(self._flatten_keys(backup_yaml))
        template_keys = set(self._flatten_keys(template_yaml))
        removed_keys = backup_keys - template_keys

        if removed_keys:
            risk_score += self.RISK_FACTORS["config_key_removed"]
            notes.append(f"Keys removed: {len(removed_keys)}")

        # Determine recommendation
        if user_customizations:
            recommendation = "smart_merge"
            notes.append(f"User settings: {', '.join(user_customizations)}")
            risk_score += 1
        elif info["diff_lines"] > 30:
            recommendation = "smart_merge"
        else:
            recommendation = "use_template"

        severity = self._risk_score_to_severity(risk_score)

        return {
            "filename": file_name,
            "changes": f"{info['diff_lines']} lines changed",
            "recommendation": recommendation,
            "conflict_severity": severity,
            "risk_score": risk_score,
            "note": "; ".join(notes) if notes else "Config update available",
            "user_customizations": user_customizations,
        }

    def _analyze_gitignore(self, file_name: str, info: dict[str, Any]) -> dict[str, Any]:
        """Analyze .gitignore with line-based comparison

        Preserves user additions, only adds new template entries.
        """
        risk_score = 0
        notes = []

        if not info["backup_exists"]:
            return {
                "filename": file_name,
                "changes": "New file from template",
                "recommendation": "use_template",
                "conflict_severity": "low",
                "risk_score": 0,
                "note": "New .gitignore will be created",
            }

        if not info["template_exists"]:
            return {
                "filename": file_name,
                "changes": "File removed in template",
                "recommendation": "keep_existing",
                "conflict_severity": "low",
                "risk_score": 0,
                "note": "Keep existing .gitignore",
            }

        # Line-based analysis
        backup_lines = set(info["backup_content"].splitlines()) if info["backup_content"] else set()
        template_lines = set(info["template_content"].splitlines()) if info["template_content"] else set()

        # Filter out comments and empty lines
        backup_entries = {line.strip() for line in backup_lines if line.strip() and not line.startswith("#")}
        template_entries = {line.strip() for line in template_lines if line.strip() and not line.startswith("#")}

        new_in_template = template_entries - backup_entries
        removed_in_template = backup_entries - template_entries

        if new_in_template:
            notes.append(f"{len(new_in_template)} new entries in template")

        if removed_in_template:
            notes.append(f"{len(removed_in_template)} user entries to preserve")
            risk_score += 1

        # gitignore is typically safe to merge (additions only)
        recommendation = "smart_merge" if removed_in_template else "use_template"
        severity = "low"

        return {
            "filename": file_name,
            "changes": f"{info['diff_lines']} lines changed",
            "recommendation": recommendation,
            "conflict_severity": severity,
            "risk_score": risk_score,
            "note": "; ".join(notes) if notes else "gitignore update",
            "new_entries": list(new_in_template),
            "user_entries": list(removed_in_template),
        }

    def _analyze_generic(self, file_name: str, info: dict[str, Any]) -> dict[str, Any]:
        """Generic file analysis for unknown file types"""
        risk_score = 0

        if not info["backup_exists"]:
            return {
                "filename": file_name,
                "changes": "New file from template",
                "recommendation": "use_template",
                "conflict_severity": "low",
                "risk_score": 0,
            }

        if not info["template_exists"]:
            return {
                "filename": file_name,
                "changes": "File removed in template",
                "recommendation": "keep_existing",
                "conflict_severity": "medium",
                "risk_score": self.RISK_FACTORS["custom_content_loss"],
            }

        if info["diff_lines"] > 100:
            risk_score += self.RISK_FACTORS["large_diff"]

        severity = self._risk_score_to_severity(risk_score)

        return {
            "filename": file_name,
            "changes": f"{info['diff_lines']} lines changed",
            "recommendation": "smart_merge" if info["diff_lines"] > 50 else "use_template",
            "conflict_severity": severity,
            "risk_score": risk_score,
        }

    def _flatten_keys(self, d: dict, parent_key: str = "") -> list[str]:
        """Flatten nested dictionary keys for comparison"""
        keys = []
        for k, v in d.items():
            new_key = f"{parent_key}.{k}" if parent_key else k
            keys.append(new_key)
            if isinstance(v, dict):
                keys.extend(self._flatten_keys(v, new_key))
        return keys

    def _risk_score_to_severity(self, score: int) -> str:
        """Convert risk score to severity level"""
        if score <= 2:
            return "low"
        elif score <= 5:
            return "medium"
        else:
            return "high"

    def _calculate_risk_level(self, total_score: int) -> str:
        """Calculate overall risk level from total score"""
        if total_score <= 3:
            return "low"
        elif total_score <= 8:
            return "medium"
        else:
            return "high"

    def _display_analysis(self, analysis: dict[str, Any]) -> None:
        """Display analysis results in Rich format"""
        # Title
        method = analysis.get("analysis_method", "semantic")
        console.print(f"\n[bold]Merge Analysis Results ({method})[/bold]")

        # Summary
        summary = analysis.get("summary", "No analysis results")
        console.print(f"\n{summary}")

        # Risk assessment
        risk_level = analysis.get("risk_assessment", "Unknown")
        risk_style = {
            "Low": "green",
            "Medium": "yellow",
            "High": "red",
        }.get(risk_level, "white")
        console.print(f"Risk Level: [{risk_style}]{risk_level}[/{risk_style}]")

        # Changes by file table
        files_list = analysis.get("files")
        if files_list and isinstance(files_list, list):
            table = Table(title="Changes by File")
            table.add_column("File", style="cyan")
            table.add_column("Changes", style="white")
            table.add_column("Recommendation", style="yellow")
            table.add_column("Risk", style="red")

            for file_info in files_list:
                if not isinstance(file_info, dict):
                    continue

                severity = file_info.get("conflict_severity", "low")
                severity_style = {
                    "low": "green",
                    "medium": "yellow",
                    "high": "red",
                }.get(severity, "white")

                table.add_row(
                    file_info.get("filename", "?"),
                    str(file_info.get("changes", ""))[:40],
                    file_info.get("recommendation", "?"),
                    severity,
                    style=severity_style,
                )

            console.print(table)

            # Additional notes
            for file_info in files_list:
                if not isinstance(file_info, dict):
                    continue
                if file_info.get("note"):
                    console.print(
                        f"\n[dim]{file_info['filename']}: {file_info['note']}[/dim]",
                    )
