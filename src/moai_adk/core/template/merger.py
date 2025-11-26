"""Template file merger (SPEC-INIT-003 v0.3.0).

Intelligently merges existing user files with new templates.
"""

from __future__ import annotations

import json
import shutil
from pathlib import Path
from typing import Any


class TemplateMerger:
    """Encapsulate template merging logic."""

    PROJECT_INFO_HEADERS = (
        "## Project Information",
        "## Project Information",  # Keep for backward compatibility
        "## Project Info",
    )

    def __init__(self, target_path: Path) -> None:
        """Initialize the merger.

        Args:
            target_path: Project path (absolute).
        """
        self.target_path = target_path.resolve()

    def merge_claude_md(self, template_path: Path, existing_path: Path) -> None:
        """Smart merge for CLAUDE.md.

        Rules:
        - Use the latest template structure/content.
        - Preserve the existing "## Project Information" section.

        Args:
            template_path: Template CLAUDE.md.
            existing_path: Existing CLAUDE.md.
        """
        # Extract the existing project information section
        existing_content = existing_path.read_text(encoding="utf-8")
        project_info_start, _ = self._find_project_info_section(existing_content)
        project_info = ""
        if project_info_start != -1:
            # Extract until EOF
            project_info = existing_content[project_info_start:]

        # Load template content
        template_content = template_path.read_text(encoding="utf-8")

        # Merge when project info exists
        if project_info:
            # Remove the project info section from the template
            template_project_start, _ = self._find_project_info_section(template_content)
            if template_project_start != -1:
                template_content = template_content[:template_project_start].rstrip()

            # Merge template content with the preserved section
            merged_content = f"{template_content}\n\n{project_info}"
            existing_path.write_text(merged_content, encoding="utf-8")
        else:
            # No project info; copy the template as-is
            shutil.copy2(template_path, existing_path)

    def _find_project_info_section(self, content: str) -> tuple[int, str | None]:
        """Find the project information header in the given content."""
        for header in self.PROJECT_INFO_HEADERS:
            index = content.find(header)
            if index != -1:
                return index, header
        return -1, None

    def merge_gitignore(self, template_path: Path, existing_path: Path) -> None:
        """.gitignore merge.

        Rules:
        - Keep existing entries.
        - Add new entries from the template.
        - Remove duplicates.

        Args:
            template_path: Template .gitignore file.
            existing_path: Existing .gitignore file.
        """
        template_lines = set(template_path.read_text(encoding="utf-8").splitlines())
        existing_lines = existing_path.read_text(encoding="utf-8").splitlines()

        # Merge while removing duplicates
        merged_lines = existing_lines + [line for line in template_lines if line not in existing_lines]

        existing_path.write_text("\n".join(merged_lines) + "\n", encoding="utf-8")

    def merge_config(self, detected_language: str | None = None) -> dict[str, str]:
        """Smart merge for config.json.

        Rules:
        - Prefer existing settings.
        - Use detected language plus defaults for new projects.

        Args:
            detected_language: Detected language.

        Returns:
            Merged configuration dictionary.
        """
        config_path = self.target_path / ".moai" / "config" / "config.json"

        # Load existing config if present
        existing_config: dict[str, Any] = {}
        if config_path.exists():
            with open(config_path, encoding="utf-8") as f:
                existing_config = json.load(f)

        # Build new config while preferring existing values
        new_config: dict[str, str] = {
            "projectName": existing_config.get("projectName", self.target_path.name),
            "mode": existing_config.get("mode", "personal"),
            "locale": existing_config.get("locale", "ko"),
            "language": existing_config.get("language", detected_language or "generic"),
        }

        return new_config

    def merge_settings_json(self, template_path: Path, existing_path: Path, backup_path: Path | None = None) -> None:
        """Smart merge for .claude/settings.json.

        Rules:
        - env: shallow merge (user variables preserved)
        - permissions.allow: array merge (deduplicated)
        - permissions.deny: template priority (security)
        - hooks: template priority

        Args:
            template_path: Template settings.json.
            existing_path: Existing settings.json.
            backup_path: Backup settings.json (optional, for user settings extraction).
        """
        # Load template
        template_data = json.loads(template_path.read_text(encoding="utf-8"))

        # Load backup or existing for user settings
        user_data: dict[str, Any] = {}
        if backup_path and backup_path.exists():
            user_data = json.loads(backup_path.read_text(encoding="utf-8"))
        elif existing_path.exists():
            user_data = json.loads(existing_path.read_text(encoding="utf-8"))

        # Merge env (shallow merge, user variables preserved)
        merged_env = {**template_data.get("env", {}), **user_data.get("env", {})}

        # Merge permissions.allow (deduplicated array merge)
        template_allow = set(template_data.get("permissions", {}).get("allow", []))
        user_allow = set(user_data.get("permissions", {}).get("allow", []))
        merged_allow = sorted(template_allow | user_allow)

        # permissions.deny: template priority (security)
        merged_deny = template_data.get("permissions", {}).get("deny", [])

        # permissions.ask: template priority + user additions
        template_ask = set(template_data.get("permissions", {}).get("ask", []))
        user_ask = set(user_data.get("permissions", {}).get("ask", []))
        merged_ask = sorted(template_ask | user_ask)

        # Start with full template (include all fields from template)
        merged = template_data.copy()

        # Override with merged values
        merged["env"] = merged_env
        merged["permissions"] = {
            "defaultMode": template_data.get("permissions", {}).get("defaultMode", "default"),
            "allow": merged_allow,
            "ask": merged_ask,
            "deny": merged_deny,
        }

        # Preserve user customizations for specific fields (if exist in backup/existing)
        preserve_fields = ["outputStyle", "spinnerTipsEnabled"]
        for field in preserve_fields:
            if field in user_data:
                merged[field] = user_data[field]

        existing_path.write_text(json.dumps(merged, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")

    def merge_github_workflows(self, template_dir: Path, existing_dir: Path) -> None:
        """Smart merge for .github/workflows/ directory.

        Rules:
        - Preserve existing user workflows (never delete)
        - Add/update only MoAI-ADK managed workflows (moai-*.yml)
        - Copy other template directories (ISSUE_TEMPLATE/, PULL_REQUEST_TEMPLATE.md)

        Args:
            template_dir: Template .github directory.
            existing_dir: Existing .github directory.
        """
        import shutil

        # Ensure workflows directory exists
        workflows_dir = existing_dir / "workflows"
        workflows_dir.mkdir(exist_ok=True)

        # Track existing user workflows for preservation
        user_workflows = set()
        if workflows_dir.exists():
            for workflow_file in workflows_dir.glob("*.yml"):
                user_workflows.add(workflow_file.name)

        # Copy template contents with smart merge for workflows
        for item in template_dir.rglob("*"):
            if item.is_file():
                rel_path = item.relative_to(template_dir)
                dst_item = existing_dir / rel_path

                # Handle workflow files specially
                if rel_path.parent.name == "workflows" and rel_path.name.endswith(".yml"):
                    # Only update MoAI-ADK managed workflows (moai-*.yml)
                    if rel_path.name.startswith("moai-"):
                        dst_item.parent.mkdir(parents=True, exist_ok=True)
                        shutil.copy2(item, dst_item)
                    # Skip non-moai workflows to preserve user custom workflows
                    continue

                # Copy non-workflow files normally
                dst_item.parent.mkdir(parents=True, exist_ok=True)
                shutil.copy2(item, dst_item)

            elif item.is_dir():
                # Create directories as needed
                rel_path = item.relative_to(template_dir)
                dst_item = existing_dir / rel_path
                dst_item.mkdir(parents=True, exist_ok=True)
