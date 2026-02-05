"""Backup utility module (SPEC-INIT-003 v0.3.0)

Selective backup strategy:
- Back up only the required files (OR condition)
- Backup path: .moai-backups/backup/ (v0.4.2)
"""

from pathlib import Path

# Backup targets (OR condition - back up when any exist)
# Note: Both YAML and JSON config files are checked for backward compatibility
BACKUP_TARGETS = [
    ".moai/config/config.yaml",
    ".moai/config/config.json",  # Legacy support
    ".moai/config/sections/",  # Section-based config files
    ".moai/project/",
    ".moai/memory/",
    ".claude/",
    ".github/",
    "CLAUDE.md",
    # MCP/LSP configuration (v2.0.0)
    ".mcp.json",  # MCP server configuration
    ".lsp.json",  # LSP server configuration
    # Git hooks (v2.0.0)
    ".git-hooks/",  # Custom git hooks
]

# User data protection paths (excluded from backups) - MUST match TemplateProcessor.PROTECTED_PATHS
PROTECTED_PATHS = [
    ".moai/specs/",
    ".moai/reports/",
    ".moai/project/",
    ".moai/config/sections/",
]


def has_any_moai_files(project_path: Path) -> bool:
    """Check whether any MoAI-ADK files exist (OR condition).

    Args:
        project_path: Project path.

    Returns:
        True when any backup target exists.
    """
    for target in BACKUP_TARGETS:
        target_path = project_path / target
        if target_path.exists():
            return True
    return False


def get_backup_targets(project_path: Path) -> list[str]:
    """Return existing backup targets.

    Args:
        project_path: Project path.

    Returns:
        List of backup targets that exist.
    """
    targets: list[str] = []
    for target in BACKUP_TARGETS:
        target_path = project_path / target
        if target_path.exists():
            targets.append(target)
    return targets


def is_protected_path(rel_path: Path) -> bool:
    """Check whether the path is protected.

    Args:
        rel_path: Relative path.

    Returns:
        True when the path should be excluded from backups.
    """
    rel_str = str(rel_path).replace("\\", "/")
    return any(rel_str.startswith(p.lstrip("./").rstrip("/")) for p in PROTECTED_PATHS)
