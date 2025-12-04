"""Custom Element Scanner for MoAI-ADK

This module provides functionality to scan for user-created custom elements
(agents, commands, skills, hooks) that are not part of the official MoAI-ADK template
installation. Unlike the existing detection functions that use template-based filtering,
this scanner compares against fresh installation file lists to identify truly custom elements.

The scanner is used during updates to provide users with a list of their custom
elements that can be selectively restored from backup.
"""

import logging
from pathlib import Path
from typing import Any, Dict, List, Set

from moai_adk.core.migration.template_utils import (
    _get_template_agent_names,
    _get_template_command_names,
    _get_template_hook_names,
    _get_template_skill_names,
)

logger = logging.getLogger(__name__)


class TemplateSkill:
    """Represents a template skill directory."""

    def __init__(self, name: str, path: Path, has_skill_md: bool, is_template: bool = False):
        """Initialize a TemplateSkill.

        Args:
            name: Name of the skill directory
            path: Path to the skill directory
            has_skill_md: Whether the skill has a SKILL.md file
            is_template: Whether this skill is part of the MoAI-ADK template
        """
        self.name = name
        self.path = path
        self.has_skill_md = has_skill_md
        self.is_template = is_template


class CustomElementScanner:
    """Scanner for detecting user-created custom elements in MoAI-ADK projects.

    This scanner identifies elements that are not part of the official MoAI-ADK
    template by comparing against fresh installation file lists, providing a more
    accurate detection of user-created content.
    """

    def __init__(self, project_path: Path):
        """Initialize the scanner.

        Args:
            project_path: Path to the MoAI-ADK project directory
        """
        self.project_path = project_path
        self.template_elements = self._get_template_elements()

    def _get_template_elements(self) -> Dict[str, Set[str]]:
        """Get template element names from fresh installation.

        Returns:
            Dictionary with element types as keys and sets of element names as values
        """
        return {
            "skills": _get_template_skill_names(),
            "commands": _get_template_command_names(),
            "agents": _get_template_agent_names(),
            "hooks": _get_template_hook_names(),
        }

    def scan_custom_elements(self) -> Dict[str, Any]:
        """Scan for user-created custom elements only.

        Scans the project directory for agents, commands, skills, and hooks that are
        NOT part of the official MoAI-ADK template installation. Template elements
        in moai/ subdirectories and moai-* prefixed skills are excluded because
        they are automatically overwritten during updates.

        Exclusions:
            - agents/moai/* - Template agents (overwritten during update)
            - commands/moai/* - Template commands (overwritten during update)
            - skills/moai-* - Template skills (overwritten during update)
            - hooks/moai/* - Template hooks (overwritten during update)

        Returns:
            Dictionary with element types as keys and lists of custom element paths as values

        Example:
            {
                "agents": [Path(".claude/agents/custom/my-agent.md")],
                "commands": [Path(".claude/commands/custom/my-command.md")],
                "skills": [Path(".claude/skills/my-custom-skill/")],
                "hooks": [Path(".claude/hooks/custom/my-hook.py")]
            }
        """
        custom_elements: Dict[str, Any] = {}

        # Scan agents (all files in .claude/agents/, excluding template files)
        custom_elements["agents"] = self._scan_custom_agents()

        # Scan commands (all files in .claude/commands/moai/, excluding template files)
        custom_elements["commands"] = self._scan_custom_commands()

        # Scan skills (all directories in .claude/skills/, excluding moai-* patterns)
        custom_elements["skills"] = self._scan_custom_skills()

        # Scan hooks (all files in .claude/hooks/moai/, excluding template files)
        custom_elements["hooks"] = self._scan_custom_hooks()

        return custom_elements

    def _scan_custom_agents(self) -> List[Path]:
        """Scan for custom agents only (excluding agents/moai/* template agents).

        Template agents in the moai/ subdirectory are automatically overwritten
        during updates. Only user-created custom agents should be available
        for restoration.

        Returns:
            List of paths to custom agent files (excluding moai/ subdirectory)
        """
        agents_dir = self.project_path / ".claude" / "agents"

        if not agents_dir.exists():
            return []

        custom_agents = []

        # Scan all subdirectories under agents EXCEPT moai/
        for item in agents_dir.iterdir():
            if item.is_dir():
                # Skip moai/ subdirectory - template agents should be overwritten
                if item.name == "moai":
                    continue
                # Scan custom subdirectories
                for file_path in item.rglob("*.md"):
                    relative_path = file_path.relative_to(self.project_path)
                    custom_agents.append(relative_path)
            elif item.is_file() and item.suffix == ".md":
                # Include root-level agent files (not in any subdirectory)
                relative_path = item.relative_to(self.project_path)
                custom_agents.append(relative_path)

        return custom_agents

    def _scan_custom_commands(self) -> List[Path]:
        """Scan for custom commands only (excluding commands/moai/* template commands).

        Template commands in the moai/ subdirectory are automatically overwritten
        during updates. Only user-created custom commands should be available
        for restoration.

        Returns:
            List of paths to custom command files (excluding moai/ subdirectory)
        """
        commands_dir = self.project_path / ".claude" / "commands"

        if not commands_dir.exists():
            return []

        custom_commands = []

        # Scan all subdirectories under commands EXCEPT moai/
        # moai/ subdirectory contains template commands that should be overwritten
        for subdir in commands_dir.iterdir():
            if subdir.is_dir():
                # Skip moai/ subdirectory - template commands should be overwritten
                if subdir.name == "moai":
                    continue
                # Scan custom subdirectories
                for file_path in subdir.rglob("*.md"):
                    relative_path = file_path.relative_to(self.project_path)
                    custom_commands.append(relative_path)
            elif subdir.is_file() and subdir.suffix == ".md":
                # Include root-level command files (not in any subdirectory)
                relative_path = subdir.relative_to(self.project_path)
                custom_commands.append(relative_path)

        return custom_commands

    def _scan_custom_skills(self) -> List[TemplateSkill]:
        """Scan for custom skills only (excluding moai-* template skills).

        Template skills (moai-* prefix) are automatically overwritten during updates.
        Only user-created custom skills should be available for restoration.

        Returns:
            List of TemplateSkill objects representing custom skill directories only
        """
        skills_dir = self.project_path / ".claude" / "skills"

        if not skills_dir.exists():
            return []

        custom_skills = []
        template_skills = self.template_elements["skills"]

        for skill_dir in skills_dir.iterdir():
            if skill_dir.is_dir():
                # Skip moai-* template skills - they should be overwritten, not restored
                # Only include user-created custom skills (non-moai-* prefix)
                if skill_dir.name.startswith("moai-"):
                    continue

                relative_path = skill_dir.relative_to(self.project_path)
                is_template_skill = skill_dir.name in template_skills

                custom_skills.append(
                    TemplateSkill(
                        name=skill_dir.name,
                        path=relative_path,
                        has_skill_md=(skill_dir / "SKILL.md").exists(),
                        is_template=is_template_skill,
                    )
                )

        return custom_skills

    def _scan_custom_hooks(self) -> List[Path]:
        """Scan for custom hooks only (excluding hooks/moai/* template hooks).

        Template hooks in the moai/ subdirectory are automatically overwritten
        during updates. Only user-created custom hooks should be available
        for restoration.

        Returns:
            List of paths to custom hook files (excluding moai/ subdirectory)
        """
        hooks_dir = self.project_path / ".claude" / "hooks"

        if not hooks_dir.exists():
            return []

        custom_hooks = []

        # Scan all subdirectories under hooks EXCEPT moai/
        # moai/ subdirectory contains template hooks that should be overwritten
        for subdir in hooks_dir.iterdir():
            if subdir.is_dir():
                # Skip moai/ subdirectory - template hooks should be overwritten
                if subdir.name == "moai":
                    continue
                # Scan custom subdirectories
                for file_path in subdir.rglob("*.py"):
                    relative_path = file_path.relative_to(self.project_path)
                    custom_hooks.append(relative_path)
            elif subdir.is_file() and subdir.suffix == ".py":
                # Include root-level hook files (not in any subdirectory)
                relative_path = subdir.relative_to(self.project_path)
                custom_hooks.append(relative_path)

        return custom_hooks

    def get_custom_elements_display_list(self) -> List[Dict[str, Any]]:
        """Get formatted list of custom elements for display.

        Returns:
            List of dictionaries containing element info for CLI display
        """
        custom_elements = self.scan_custom_elements()
        display_list = []
        index = 1

        # Add agents
        for agent_path in custom_elements["agents"]:
            display_list.append(
                {
                    "index": index,
                    "type": "agent",
                    "name": agent_path.stem,
                    "path": str(agent_path),
                    "display_name": f"{agent_path.stem} (agent)",
                }
            )
            index += 1

        # Add commands
        for command_path in custom_elements["commands"]:
            display_list.append(
                {
                    "index": index,
                    "type": "command",
                    "name": command_path.stem,
                    "path": str(command_path),
                    "display_name": f"{command_path.stem} (command)",
                }
            )
            index += 1

        # Add skills
        skills_list: List[TemplateSkill] = custom_elements["skills"]
        for skill in skills_list:
            skill_display_name = skill.name
            # Add indicator for template vs custom skills
            if hasattr(skill, "is_template") and skill.is_template:
                skill_display_name = f"{skill.name} (template)"
            else:
                skill_display_name = f"{skill.name} (custom)"

            display_list.append(
                {
                    "index": index,
                    "type": "skill",
                    "name": skill.name,
                    "path": str(skill.path),
                    "display_name": skill_display_name,
                }
            )
            index += 1

        # Add hooks
        for hook_path in custom_elements["hooks"]:
            display_list.append(
                {
                    "index": index,
                    "type": "hook",
                    "name": hook_path.stem,
                    "path": str(hook_path),
                    "display_name": f"{hook_path.stem} (hook)",
                }
            )
            index += 1

        return display_list

    def get_element_count(self) -> int:
        """Get total count of custom elements.

        Returns:
            Total number of custom elements found
        """
        custom_elements = self.scan_custom_elements()
        return (
            len(custom_elements["agents"])
            + len(custom_elements["commands"])
            + len(custom_elements["skills"])
            + len(custom_elements["hooks"])
        )


def create_custom_element_scanner(project_path: str | Path) -> CustomElementScanner:
    """Factory function to create a CustomElementScanner.

    Args:
        project_path: Path to the MoAI-ADK project directory

    Returns:
        Configured CustomElementScanner instance

    Example:
        >>> scanner = create_custom_element_scanner("/path/to/project")
        >>> elements = scanner.scan_custom_elements()
        >>> print(f"Found {len(elements['agents']) custom agents")
    """
    return CustomElementScanner(Path(project_path).resolve())
