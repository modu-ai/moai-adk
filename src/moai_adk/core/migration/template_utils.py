"""Template utilities for MoAI-ADK migration system.

This module provides utility functions for working with MoAI-ADK templates,
including functions to get template element names for comparison with user-created elements.
"""

from pathlib import Path
from typing import Set


def _get_template_skill_names() -> Set[str]:
    """Get skill names from the fresh installation template.

    Returns:
        Set of skill directory names found in the template
    """
    template_path = Path(__file__).parent.parent.parent / "templates" / ".claude" / "skills"
    skill_names = set()

    if template_path.exists():
        for skill_dir in template_path.iterdir():
            if skill_dir.is_dir() and skill_dir.name.startswith("moai-"):
                skill_names.add(skill_dir.name)

    return skill_names


def _get_template_command_names() -> Set[str]:
    """Get command names from the fresh installation template.

    Returns:
        Set of command file names found in the template
    """
    template_path = Path(__file__).parent.parent.parent / "templates" / ".claude" / "commands" / "moai"
    command_names = set()

    if template_path.exists():
        for cmd_file in template_path.glob("*.md"):
            command_names.add(cmd_file.name)

    return command_names


def _get_template_agent_names() -> Set[str]:
    """Get agent names from the fresh installation template.

    Returns:
        Set of agent file names found in the template
    """
    template_path = Path(__file__).parent.parent.parent / "templates" / ".claude" / "agents"
    agent_names = set()

    if template_path.exists():
        for agent_file in template_path.rglob("*.md"):
            if agent_file.parent.name == "moai":
                agent_names.add(agent_file.name)

    return agent_names


def _get_template_hook_names() -> Set[str]:
    """Get hook names from the fresh installation template.

    Returns:
        Set of hook file names found in the template
    """
    template_path = Path(__file__).parent.parent.parent / "templates" / ".claude" / "hooks" / "moai"
    hook_names = set()

    if template_path.exists():
        for hook_file in template_path.rglob("*.py"):
            hook_names.add(hook_file.name)

    return hook_names
