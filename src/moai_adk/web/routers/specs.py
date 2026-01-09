"""SPEC Status Router

REST API endpoints for SPEC status management
including listing and retrieving SPEC information.
Supports both directory-based SPECs (new format) and single-file SPECs (legacy).
"""

from datetime import datetime
from pathlib import Path
from typing import Optional

import yaml
from fastapi import APIRouter, HTTPException, status
from pydantic import BaseModel, Field

router = APIRouter()


class SpecStatus(BaseModel):
    """SPEC status information model"""

    spec_id: str = Field(..., description="SPEC identifier")
    title: str = Field(..., description="SPEC title")
    description: str = Field(default="", description="SPEC description")
    status: str = Field(..., description="Current status (draft, approved, in_progress, completed)")
    priority: str = Field(default="medium", description="Priority level")
    progress: int = Field(default=0, description="Progress percentage 0-100")
    tags: list[str] = Field(default_factory=list, description="Tags for categorization")
    created_at: Optional[datetime] = Field(default=None, description="Creation timestamp")
    updated_at: Optional[datetime] = Field(default=None, description="Last update timestamp")
    file_path: Optional[str] = Field(default=None, description="Path to SPEC file")
    worktree_path: Optional[str] = Field(default=None, description="Path to worktree if exists")


class SpecList(BaseModel):
    """Response model for listing SPECs"""

    specs: list[SpecStatus] = Field(default_factory=list, description="List of SPECs")
    total: int = Field(default=0, description="Total number of SPECs")


def _find_spec_entries(base_path: Path) -> list[tuple[Path, str]]:
    """Find all SPEC entries in the project

    Supports both directory-based SPECs (SPEC-XXX/) and file-based (SPEC-XXX.md)

    Args:
        base_path: Base path to search from

    Returns:
        List of tuples (path, format_type) where format_type is 'directory' or 'file'
    """
    spec_entries = []

    # Check .moai/specs directory
    moai_specs_dir = base_path / ".moai" / "specs"
    if moai_specs_dir.exists():
        # Directory-based SPECs (new format)
        for item in moai_specs_dir.iterdir():
            if item.is_dir() and item.name.startswith("SPEC-"):
                spec_file = item / "spec.md"
                if spec_file.exists():
                    spec_entries.append((item, "directory"))

        # File-based SPECs (legacy format)
        for item in moai_specs_dir.glob("SPEC-*.md"):
            if item.is_file():
                spec_entries.append((item, "file"))

    # Check docs/specs directory (legacy)
    docs_specs_dir = base_path / "docs" / "specs"
    if docs_specs_dir.exists():
        for item in docs_specs_dir.glob("SPEC-*.md"):
            if item.is_file():
                spec_entries.append((item, "file"))

    return sorted(spec_entries, key=lambda x: x[0].name)


def _parse_yaml_frontmatter(content: str) -> dict:
    """Parse YAML frontmatter from markdown content

    Args:
        content: Markdown content with optional YAML frontmatter

    Returns:
        Dict of frontmatter values
    """
    if not content.startswith("---"):
        return {}

    try:
        # Find the closing ---
        end_idx = content.find("---", 3)
        if end_idx == -1:
            return {}

        yaml_content = content[3:end_idx].strip()
        return yaml.safe_load(yaml_content) or {}
    except Exception:
        return {}


def _calculate_progress(spec_status: str, plan_path: Optional[Path] = None) -> int:
    """Calculate progress percentage based on status

    Args:
        spec_status: Current SPEC status
        plan_path: Path to plan.md if exists

    Returns:
        Progress percentage 0-100
    """
    status_progress = {
        "draft": 10,
        "approved": 20,
        "planned": 30,
        "in_progress": 50,
        "implementing": 60,
        "testing": 80,
        "completed": 100,
        "blocked": 0,
    }
    return status_progress.get(spec_status.lower(), 0)


def _check_worktree(spec_id: str) -> Optional[str]:
    """Check if a worktree exists for this SPEC

    Args:
        spec_id: The SPEC identifier

    Returns:
        Worktree path if exists, None otherwise
    """
    import os

    home = Path.home()
    worktree_base = home / "worktrees"

    if not worktree_base.exists():
        # Try alternative location from env
        worktree_base = Path(os.environ.get("MOAI_WORKTREE_ROOT", "")) / "worktrees"

    if not worktree_base.exists():
        return None

    # Search for worktree matching SPEC ID
    for project_dir in worktree_base.iterdir():
        if project_dir.is_dir():
            spec_worktree = project_dir / spec_id
            if spec_worktree.exists():
                return str(spec_worktree)

    return None


def _parse_directory_spec(dir_path: Path) -> Optional[SpecStatus]:
    """Parse a directory-based SPEC

    Args:
        dir_path: Path to the SPEC directory

    Returns:
        SpecStatus if valid, None otherwise
    """
    try:
        spec_file = dir_path / "spec.md"
        if not spec_file.exists():
            return None

        content = spec_file.read_text(encoding="utf-8")
        frontmatter = _parse_yaml_frontmatter(content)

        # Extract SPEC ID from directory name
        spec_id = dir_path.name

        # Get values from frontmatter or defaults
        title = frontmatter.get("title", spec_id)
        description = frontmatter.get("description", "")
        spec_status = frontmatter.get("status", "draft")
        priority = frontmatter.get("priority", "medium")
        tags = frontmatter.get("tags", [])

        # If no title in frontmatter, try to get from first heading
        if title == spec_id:
            lines = content.split("\n")
            for line in lines:
                if line.startswith("# "):
                    title = line[2:].strip()
                    break

        # Calculate progress
        plan_path = dir_path / "plan.md"
        progress = _calculate_progress(spec_status, plan_path if plan_path.exists() else None)

        # Check for worktree
        worktree_path = _check_worktree(spec_id)

        # Get file timestamps
        stat = spec_file.stat()

        return SpecStatus(
            spec_id=spec_id,
            title=title,
            description=description,
            status=spec_status,
            priority=priority,
            progress=progress,
            tags=tags if isinstance(tags, list) else [],
            created_at=datetime.fromtimestamp(stat.st_ctime),
            updated_at=datetime.fromtimestamp(stat.st_mtime),
            file_path=str(spec_file),
            worktree_path=worktree_path,
        )

    except Exception:
        return None


def _parse_file_spec(file_path: Path) -> Optional[SpecStatus]:
    """Parse a file-based SPEC (legacy format)

    Args:
        file_path: Path to the SPEC file

    Returns:
        SpecStatus if file is valid, None otherwise
    """
    try:
        content = file_path.read_text(encoding="utf-8")
        frontmatter = _parse_yaml_frontmatter(content)
        lines = content.split("\n")

        # Extract SPEC ID from filename
        spec_id = file_path.stem

        # Default values
        title = frontmatter.get("title", spec_id)
        description = frontmatter.get("description", "")
        spec_status = frontmatter.get("status", "draft")
        priority = frontmatter.get("priority", "medium")
        tags = frontmatter.get("tags", [])

        # Parse header for title if not in frontmatter
        if title == spec_id:
            for line in lines[:50]:
                line = line.strip()
                if line.startswith("# ") and title == spec_id:
                    title = line[2:].strip()
                    break

        # Legacy status parsing
        if spec_status == "draft":
            for line in lines[:50]:
                line = line.strip()
                if line.lower().startswith("status:"):
                    spec_status = line.split(":", 1)[1].strip().lower()
                    break
                elif line.lower().startswith("- status:"):
                    spec_status = line.split(":", 1)[1].strip().lower()
                    break

        # Calculate progress
        progress = _calculate_progress(spec_status)

        # Check for worktree
        worktree_path = _check_worktree(spec_id)

        # Get file timestamps
        stat = file_path.stat()

        return SpecStatus(
            spec_id=spec_id,
            title=title,
            description=description,
            status=spec_status,
            priority=priority,
            progress=progress,
            tags=tags if isinstance(tags, list) else [],
            created_at=datetime.fromtimestamp(stat.st_ctime),
            updated_at=datetime.fromtimestamp(stat.st_mtime),
            file_path=str(file_path),
            worktree_path=worktree_path,
        )

    except Exception:
        return None


@router.get("/specs", response_model=SpecList)
async def list_specs() -> SpecList:
    """List all SPECs in the project

    Returns:
        SpecList with all discovered SPECs
    """
    base_path = Path.cwd()
    spec_entries = _find_spec_entries(base_path)

    specs = []
    for path, format_type in spec_entries:
        if format_type == "directory":
            spec = _parse_directory_spec(path)
        else:
            spec = _parse_file_spec(path)

        if spec:
            specs.append(spec)

    return SpecList(specs=specs, total=len(specs))


@router.get("/specs/{spec_id}", response_model=SpecStatus)
async def get_spec(spec_id: str) -> SpecStatus:
    """Get a specific SPEC by ID

    Args:
        spec_id: The SPEC identifier (e.g., SPEC-001)

    Returns:
        SpecStatus for the requested SPEC

    Raises:
        HTTPException: If SPEC not found
    """
    base_path = Path.cwd()
    spec_entries = _find_spec_entries(base_path)

    for path, format_type in spec_entries:
        entry_id = path.stem if format_type == "file" else path.name
        if entry_id == spec_id:
            if format_type == "directory":
                spec = _parse_directory_spec(path)
            else:
                spec = _parse_file_spec(path)

            if spec:
                return spec

    raise HTTPException(
        status_code=status.HTTP_404_NOT_FOUND,
        detail=f"SPEC '{spec_id}' not found",
    )


@router.post("/specs/{spec_id}/run")
async def run_spec(spec_id: str) -> dict:
    """Trigger execution of a SPEC via terminal

    Args:
        spec_id: The SPEC identifier to execute

    Returns:
        Dict with execution status and terminal info
    """
    # Verify SPEC exists
    base_path = Path.cwd()
    spec_entries = _find_spec_entries(base_path)

    spec_found = False
    worktree_path = None

    for path, format_type in spec_entries:
        entry_id = path.stem if format_type == "file" else path.name
        if entry_id == spec_id:
            spec_found = True
            worktree_path = _check_worktree(spec_id)
            break

    if not spec_found:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"SPEC '{spec_id}' not found",
        )

    # Return info for terminal spawning
    # Actual terminal creation handled by terminal router
    return {
        "spec_id": spec_id,
        "worktree_path": worktree_path,
        "command": f"claude /moai:all-is-well {spec_id}",
        "status": "ready",
        "message": f"Ready to execute {spec_id}. Use terminal endpoint to spawn.",
    }
