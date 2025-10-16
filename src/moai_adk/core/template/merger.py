# @CODE:TEMPLATE-001 | SPEC: SPEC-INIT-003.md | Chain: TEMPLATE-001
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
        - Preserve the existing "## 프로젝트 정보" section.

        Args:
            template_path: Template CLAUDE.md.
            existing_path: Existing CLAUDE.md.
        """
        # Extract the existing "## 프로젝트 정보" section
        existing_content = existing_path.read_text(encoding="utf-8")
        project_info_start = existing_content.find("## 프로젝트 정보")
        project_info = ""
        if project_info_start != -1:
            # Extract until EOF
            project_info = existing_content[project_info_start:]

        # Load template content
        template_content = template_path.read_text(encoding="utf-8")

        # Merge when project info exists
        if project_info:
            # Remove the project info section from the template
            template_project_start = template_content.find("## 프로젝트 정보")
            if template_project_start != -1:
                template_content = template_content[:template_project_start].rstrip()

            # Merge template content with the preserved section
            merged_content = f"{template_content}\n\n{project_info}"
            existing_path.write_text(merged_content, encoding="utf-8")
        else:
            # No project info; copy the template as-is
            shutil.copy2(template_path, existing_path)

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
        merged_lines = existing_lines + [
            line for line in template_lines if line not in existing_lines
        ]

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
        config_path = self.target_path / ".moai" / "config.json"

        # Load existing config if present
        existing_config: dict[str, Any] = {}
        if config_path.exists():
            with open(config_path, encoding="utf-8") as f:
                existing_config = json.load(f)

        # Build new config while preferring existing values
        new_config: dict[str, str] = {
            "projectName": existing_config.get(
                "projectName", self.target_path.name
            ),
            "mode": existing_config.get("mode", "personal"),
            "locale": existing_config.get("locale", "ko"),
            "language": existing_config.get(
                "language", detected_language or "generic"
            ),
        }

        return new_config
