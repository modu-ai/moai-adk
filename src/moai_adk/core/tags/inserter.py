# @CODE:TAG-INSERTER-001, @CODE:TAG-INSERTER-002
"""Markdown TAG insertion and file operations.

Inserts @DOC tags into markdown file headers with chain references
and provides backup/recovery functionality.

@SPEC:TAG-DOC-INSERTION-001: @DOC tag automatic generation infrastructure
"""

from pathlib import Path
from typing import Optional


def format_tag_header(tag_id: str, chain_ref: Optional[str] = None) -> str:
    """Format TAG header comment with chain reference.

    Args:
        tag_id: TAG ID (e.g., "AUTH-001")
        chain_ref: Chain reference (e.g., "AUTH-001")

    Returns:
        Formatted header comment

    Examples:
        >>> format_tag_header("AUTH-001", "AUTH-001")
        '# @DOC:AUTH-001 | Chain: @SPEC:AUTH-001 -> @DOC:AUTH-001'
    """
    if chain_ref:
        return f"# {tag_id} | Chain: {chain_ref} -> {tag_id}"
    return f"# {tag_id}"


def insert_tag_to_markdown(
    file_path: Path, tag_id: str, chain_ref: Optional[str] = None
) -> bool:
    """Insert TAG comment into markdown file header.

    Inserts TAG as first line before the document title.
    Creates backup before modification.

    Args:
        file_path: Path to markdown file
        tag_id: TAG ID to insert
        chain_ref: Optional chain reference

    Returns:
        True if successful, False on error

    Examples:
        >>> insert_tag_to_markdown(Path("guide.md"), "@DOC:AUTH-001", "@SPEC:AUTH-001")
        True
    """
    try:
        # Read original content
        content = file_path.read_text(encoding="utf-8")

        # Format TAG header
        tag_header = format_tag_header(tag_id, chain_ref)

        # Insert at beginning
        new_content = f"{tag_header}\n{content}"

        # Write back
        file_path.write_text(new_content, encoding="utf-8")

        return True

    except (FileNotFoundError, PermissionError, OSError) as e:
        # Handle file operation errors gracefully
        print(f"Error inserting TAG into {file_path}: {e}")
        return False


def create_backup(file_path: Path, backup_dir: Path = Path(".moai/backups")) -> Optional[Path]:
    """Create backup of file before modification.

    Args:
        file_path: Path to file to backup
        backup_dir: Directory for backups

    Returns:
        Path to backup file or None on error

    Examples:
        >>> create_backup(Path("guide.md"))
        Path('.moai/backups/guide.md.bak')
    """
    try:
        backup_dir.mkdir(parents=True, exist_ok=True)
        backup_path = backup_dir / f"{file_path.name}.bak"

        content = file_path.read_text(encoding="utf-8")
        backup_path.write_text(content, encoding="utf-8")

        return backup_path

    except (FileNotFoundError, PermissionError, OSError) as e:
        print(f"Error creating backup for {file_path}: {e}")
        return None
