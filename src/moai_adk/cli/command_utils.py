#!/usr/bin/env python3
# @REQ:CLI-UTILS-001
"""
CLI utility functions for MoAI-ADK

Contains utility functions that support CLI commands including
mode configuration and common helper functions.
"""

import json
from datetime import datetime
from pathlib import Path

import click
from colorama import Fore, Style

from .._version import __version__


def create_mode_configuration(
    project_dir: Path, project_mode: str, quiet: bool = False
) -> None:
    """Create mode-specific configuration for MoAI-ADK project."""
    # Create .moai directory if it doesn't exist
    moai_dir = project_dir / ".moai"
    moai_dir.mkdir(exist_ok=True)

    # Create mode-specific configuration
    config = {
        "project": {
            "name": project_dir.name,
            "mode": project_mode,
            "version": __version__,
            "created": datetime.now().isoformat(),
            "constitution_version": "2.1",
        },
        "git_strategy": {
            "personal": {
                "auto_checkpoint": True,
                "checkpoint_interval": 300,  # 5 minutes
                "simplified_commits": True,
                "auto_push": False,
                "backup_before_sync": True,
                "conflict_strategy": "local_priority",
                "max_checkpoints": 50,
                "cleanup_days": 7,
            },
            "team": {
                "auto_checkpoint": False,
                "structured_commits": True,
                "required_reviews": True,
                "auto_pr": True,
                "auto_sync": True,
                "sync_interval": 1800,  # 30 minutes
                "conflict_strategy": "remote_priority",
                "gitflow_strict": True,
            },
        },
        "workflow": {
            "spec_auto_commit": True,
            "build_tdd_commits": True,
            "sync_living_docs": True,
            "constitution_check": True,
            "tag_tracking": True,
        },
        "features": {
            "checkpoint_system": project_mode == "personal",
            "auto_rollback": project_mode == "personal",
            "smart_sync": True,
            "branch_management": True,
            "commit_automation": True,
        },
    }

    # Save configuration
    config_file = moai_dir / "config.json"
    with open(config_file, "w", encoding="utf-8") as f:
        json.dump(config, f, indent=2, ensure_ascii=False)

    if not quiet:
        mode_desc = {
            "personal": "Personal development optimized (checkpoints, simplified workflow)",
            "team": "Team collaboration optimized (GitFlow, structured commits, PR automation)",
        }
        click.echo(
            f"  {Fore.GREEN}✓{Style.RESET_ALL} {project_mode.title()} mode configured: {mode_desc[project_mode]}"
        )

    # Create checkpoints directory for personal mode
    if project_mode == "personal":
        checkpoints_dir = moai_dir / "checkpoints"
        checkpoints_dir.mkdir(exist_ok=True)

        # Initialize checkpoint metadata
        metadata = {"checkpoints": []}
        metadata_file = checkpoints_dir / "metadata.json"
        with open(metadata_file, "w", encoding="utf-8") as f:
            json.dump(metadata, f, indent=2)

        if not quiet:
            click.echo(
                f"  {Fore.GREEN}✓{Style.RESET_ALL} Checkpoint system initialized"
            )