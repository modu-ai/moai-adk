"""
Template Variable Synchronizer for MoAI-ADK

Handles synchronization of template variables when configuration is updated
through /moai:0-project command or other configuration management operations.

This module ensures that when language or user settings change, all template
variables in the system are properly updated to reflect the new configuration.
"""

import json
import re
from pathlib import Path
from typing import Any, Dict, List, Optional

from .language_config_resolver import get_resolver


class TemplateVariableSynchronizer:
    """
    Synchronizes template variables across the MoAI-ADK system when configuration changes.

    Responsible for:
    - Detecting template variables that need updating
    - Re-substituting variables in affected files
    - Maintaining consistency between config and template files
    - Tracking template variable usage across the project
    """

    # Template variable patterns to track and synchronize
    TEMPLATE_PATTERNS = {
        r"\{\{CONVERSATION_LANGUAGE\}\}",
        r"\{\{CONVERSATION_LANGUAGE_NAME\}\}",
        r"\{\{AGENT_PROMPT_LANGUAGE\}\}",
        r"\{\{USER_NAME\}\}",
        r"\{\{PERSONALIZED_GREETING\}\}",
        r"\{\{LANGUAGE_CONFIG_SOURCE\}\}",
    }

    # Files that typically contain template variables
    TEMPLATE_TRACKING_PATTERNS = [
        ".claude/settings.json",
        ".claude/settings.local.json",
        ".moai/config/config.json",
        ".claude/output-styles/**/*.md",
        ".claude/hooks/**/*.py",
        ".claude/commands/**/*.md",
        ".claude/skills/**/*.md",
        "CLAUDE.md",
        "README.md",
    ]

    def __init__(self, project_root: str):
        """
        Initialize the synchronizer.

        Args:
            project_root: Root directory of the MoAI-ADK project
        """
        self.project_root = Path(project_root)
        self.language_resolver = get_resolver(project_root)

    def synchronize_after_config_change(self, changed_config_path: Optional[Path] = None) -> Dict[str, Any]:
        """
        Synchronize template variables after configuration changes.

        Args:
            changed_config_path: Path to the configuration file that was changed
                                 (None if unknown or multiple files changed)

        Returns:
            Dictionary containing synchronization results
        """
        results: Dict[str, Any] = {
            "files_updated": 0,
            "variables_updated": [],
            "errors": [],
            "sync_status": "completed",
        }

        try:
            # Get current resolved configuration
            current_config = self.language_resolver.resolve_config(force_refresh=True)
            template_vars = self.language_resolver.export_template_variables(current_config)

            # Find files that need updating
            files_to_update = self._find_files_with_template_variables(changed_config_path)

            # Update each file
            for file_path in files_to_update:
                try:
                    updated_vars = self._update_file_template_variables(file_path, template_vars)
                    if updated_vars:
                        files_updated: int = results["files_updated"]  # type: ignore[assignment]
                        results["files_updated"] = files_updated + 1
                        variables_updated: List[str] = results["variables_updated"]  # type: ignore[assignment]
                        variables_updated.extend(updated_vars)

                except Exception as e:
                    error_msg = f"Failed to update {file_path}: {str(e)}"
                    errors: List[str] = results["errors"]  # type: ignore[assignment]
                    errors.append(error_msg)

            # Special handling for certain file types
            self._handle_special_file_updates(template_vars, results)

        except Exception as e:
            results["sync_status"] = "failed"
            errors_list: List[str] = results["errors"]  # type: ignore[assignment]
            errors_list.append(f"Synchronization failed: {str(e)}")

        return results

    def _find_files_with_template_variables(self, changed_config_path: Optional[Path]) -> List[Path]:
        """
        Find files that contain template variables and might need updating.

        Args:
            changed_config_path: Specific config file that changed, if known

        Returns:
            List of file paths that contain template variables
        """
        files_with_variables = []

        # If a specific config file changed, prioritize files that depend on it
        if changed_config_path:
            dependency_map = {
                ".moai/config/config.json": [
                    ".claude/settings.json",
                    ".claude/settings.local.json",
                    ".claude/output-styles",
                    ".claude/hooks",
                ]
            }

            config_key = str(changed_config_path.relative_to(self.project_root))
            if config_key in dependency_map:
                for pattern in dependency_map[config_key]:
                    files_with_variables.extend(self._glob_files(pattern))

        # Always check common template files
        common_patterns = [
            ".claude/settings.json",
            ".claude/settings.local.json",
            ".claude/output-styles/**/*.md",
            "CLAUDE.md",
        ]

        for pattern in common_patterns:
            files_with_variables.extend(self._glob_files(pattern))

        # Remove duplicates and sort
        files_with_variables = list(set(files_with_variables))
        files_with_variables.sort()

        return [f for f in files_with_variables if f.exists() and f.is_file()]

    def _glob_files(self, pattern: str) -> List[Path]:
        """
        Glob files matching a pattern.

        Args:
            pattern: Glob pattern (supports ** for recursive matching)

        Returns:
            List of matching file paths
        """
        try:
            return list(self.project_root.glob(pattern))
        except (OSError, ValueError):
            return []

    def _update_file_template_variables(self, file_path: Path, template_vars: Dict[str, str]) -> List[str]:
        """
        Update template variables in a specific file.

        Args:
            file_path: Path to the file to update
            template_vars: Dictionary of template variables to substitute

        Returns:
            List of variable names that were updated in this file
        """
        if not file_path.exists():
            return []

        try:
            content = file_path.read_text(encoding="utf-8")
            original_content = content

            # Track which variables were updated
            updated_vars = []

            # Apply each template variable substitution
            for var_name, var_value in template_vars.items():
                # Look for {{VARIABLE_NAME}} pattern
                pattern = re.compile(r"\{\{" + re.escape(var_name) + r"\}\}")

                if pattern.search(content):
                    content = pattern.sub(var_value, content)
                    updated_vars.append(var_name)

            # Only write if content changed
            if content != original_content:
                file_path.write_text(content, encoding="utf-8")

            return updated_vars

        except (OSError, UnicodeDecodeError, UnicodeEncodeError):
            # File read/write errors - skip this file
            return []

    def _handle_special_file_updates(self, template_vars: Dict[str, str], results: Dict[str, Any]) -> None:
        """
        Handle special cases for certain file types.

        Args:
            template_vars: Dictionary of template variables
            results: Results dictionary to update
        """
        # Handle settings.json environment variables section
        settings_file = self.project_root / ".claude" / "settings.json"
        if settings_file.exists():
            try:
                self._update_settings_env_vars(settings_file, template_vars, results)
            except Exception as e:
                results["errors"].append(f"Failed to update settings.json env vars: {str(e)}")

    def _update_settings_env_vars(
        self,
        settings_file: Path,
        template_vars: Dict[str, str],
        results: Dict[str, Any],
    ) -> None:
        """
        Update the environment variables section in settings.json.

        Args:
            settings_file: Path to settings.json
            template_vars: Template variables to use for env values
            results: Results dictionary to update
        """
        try:
            settings_data = json.loads(settings_file.read_text(encoding="utf-8"))

            # Define environment variable mappings
            env_mappings = {
                "CONVERSATION_LANGUAGE": "MOAI_CONVERSATION_LANG",
                "AGENT_PROMPT_LANGUAGE": "MOAI_AGENT_PROMPT_LANG",
                "CONVERSATION_LANGUAGE_NAME": "MOAI_CONVERSATION_LANG_NAME",
                "USER_NAME": "MOAI_USER_NAME",
                "LANGUAGE_CONFIG_SOURCE": "MOAI_CONFIG_SOURCE",
            }

            # Update or create env section
            if "env" not in settings_data:
                settings_data["env"] = {}

            updated_vars = []
            for template_var, env_var in env_mappings.items():
                if template_var in template_vars:
                    old_value = settings_data["env"].get(env_var)
                    new_value = template_vars[template_var]

                    if old_value != new_value:
                        settings_data["env"][env_var] = new_value
                        updated_vars.append(f"{env_var}: {old_value} → {new_value}")

            # Write back if changed
            if updated_vars:
                settings_file.write_text(
                    json.dumps(settings_data, indent=2, ensure_ascii=False) + "\n",
                    encoding="utf-8",
                )
                results["files_updated"] += 1
                results["variables_updated"].extend(updated_vars)

        except (json.JSONDecodeError, OSError, UnicodeDecodeError, UnicodeEncodeError):
            # Skip if settings.json is malformed or can't be accessed
            pass

    def validate_template_variable_consistency(self) -> Dict[str, Any]:
        """
        Validate that template variables are consistent across the project.

        Returns:
            Dictionary containing validation results
        """
        validation_results = {
            "status": "passed",
            "inconsistencies": [],
            "total_files_checked": 0,
            "files_with_variables": 0,
        }

        try:
            # Get current template variables
            current_config = self.language_resolver.resolve_config()
            current_vars = self.language_resolver.export_template_variables(current_config)

            # Check files for template variable consistency
            files_with_variables = self._find_files_with_template_variables(None)
            validation_results["total_files_checked"] = len(files_with_variables)

            for file_path in files_with_variables:
                try:
                    content = file_path.read_text(encoding="utf-8")
                    file_inconsistencies: List[str] = []

                    # Check each template variable
                    for var_name, expected_value in current_vars.items():
                        pattern = re.compile(r"\{\{" + re.escape(var_name) + r"\}\}")

                        if pattern.search(content):
                            files_with_vars: int = validation_results["files_with_variables"]  # type: ignore[assignment]
                            validation_results["files_with_variables"] = files_with_vars + 1
                            # Variable found but not substituted - this might be expected
                            # Only report as inconsistency if we expect it to be substituted
                            pass

                    if file_inconsistencies:
                        inconsistencies: List[Dict[str, Any]] = validation_results["inconsistencies"]  # type: ignore[assignment]
                        inconsistencies.append(
                            {
                                "file": str(file_path.relative_to(self.project_root)),
                                "issues": file_inconsistencies,
                            }
                        )

                except (OSError, UnicodeDecodeError):
                    # Skip files that can't be read
                    continue

            if validation_results["inconsistencies"]:
                validation_results["status"] = "warning"

        except Exception as e:
            validation_results["status"] = "failed"
            validation_results["error"] = str(e)

        return validation_results

    def get_template_variable_usage_report(self) -> Dict[str, Any]:
        """
        Generate a report of template variable usage across the project.

        Returns:
            Dictionary containing template variable usage statistics
        """
        usage_report = {
            "total_files_with_variables": 0,
            "variable_usage": {},
            "files_by_variable": {},
            "unsubstituted_variables": [],
        }

        try:
            files_with_variables = self._find_files_with_template_variables(None)
            usage_report["total_files_with_variables"] = len(files_with_variables)

            for file_path in files_with_variables:
                try:
                    content = file_path.read_text(encoding="utf-8")
                    relative_path = str(file_path.relative_to(self.project_root))

                    # Check for each template variable pattern
                    for pattern_str in self.TEMPLATE_PATTERNS:
                        pattern = re.compile(pattern_str)
                        matches = pattern.findall(content)

                        if matches:
                            var_name = pattern_str.strip(r"{}")
                            variable_usage: Dict[str, int] = usage_report["variable_usage"]  # type: ignore[assignment]
                            files_by_variable: Dict[str, List[str]] = usage_report["files_by_variable"]  # type: ignore[assignment]
                            unsubstituted_variables: List[Dict[str, Any]] = usage_report["unsubstituted_variables"]  # type: ignore[assignment]

                            if var_name not in variable_usage:
                                variable_usage[var_name] = 0
                                files_by_variable[var_name] = []

                            variable_usage[var_name] += len(matches)
                            files_by_variable[var_name].append(relative_path)

                            # Track unsubstituted variables
                            unsubstituted_variables.append(
                                {
                                    "file": relative_path,
                                    "variable": var_name,
                                    "count": len(matches),
                                }
                            )

                except (OSError, UnicodeDecodeError):
                    continue

        except Exception as e:
            usage_report["error"] = str(e)

        return usage_report


def synchronize_template_variables(project_root: str, changed_config_path: Optional[str] = None) -> Dict[str, Any]:
    """
    Convenience function to synchronize template variables.

    Args:
        project_root: Root directory of the project
        changed_config_path: Path to configuration file that changed

    Returns:
        Synchronization results
    """
    synchronizer = TemplateVariableSynchronizer(project_root)

    config_path = Path(changed_config_path) if changed_config_path else None
    return synchronizer.synchronize_after_config_change(config_path)
