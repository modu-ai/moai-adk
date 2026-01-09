"""SPEC Status Router

REST API endpoints for SPEC status management
including listing and retrieving SPEC information.
"""

from datetime import datetime
from pathlib import Path
from typing import Optional

from fastapi import APIRouter, HTTPException, status
from pydantic import BaseModel, Field

router = APIRouter()


class SpecStatus(BaseModel):
    """SPEC status information model"""

    spec_id: str = Field(..., description="SPEC identifier")
    title: str = Field(..., description="SPEC title")
    status: str = Field(..., description="Current status (draft, approved, in_progress, completed)")
    created_at: Optional[datetime] = Field(default=None, description="Creation timestamp")
    updated_at: Optional[datetime] = Field(default=None, description="Last update timestamp")
    file_path: Optional[str] = Field(default=None, description="Path to SPEC file")


class SpecList(BaseModel):
    """Response model for listing SPECs"""

    specs: list[SpecStatus] = Field(default_factory=list, description="List of SPECs")
    total: int = Field(default=0, description="Total number of SPECs")


def _find_spec_files(base_path: Path) -> list[Path]:
    """Find all SPEC files in the project

    Args:
        base_path: Base path to search from

    Returns:
        List of paths to SPEC files
    """
    spec_paths = []

    # Check .moai/specs directory
    moai_specs_dir = base_path / ".moai" / "specs"
    if moai_specs_dir.exists():
        spec_paths.extend(moai_specs_dir.glob("SPEC-*.md"))

    # Check docs/specs directory
    docs_specs_dir = base_path / "docs" / "specs"
    if docs_specs_dir.exists():
        spec_paths.extend(docs_specs_dir.glob("SPEC-*.md"))

    return sorted(spec_paths)


def _parse_spec_file(file_path: Path) -> Optional[SpecStatus]:
    """Parse a SPEC file to extract status information

    Args:
        file_path: Path to the SPEC file

    Returns:
        SpecStatus if file is valid, None otherwise
    """
    try:
        content = file_path.read_text(encoding="utf-8")
        lines = content.split("\n")

        # Extract SPEC ID from filename
        spec_id = file_path.stem

        # Default values
        title = spec_id
        spec_status = "draft"

        # Parse header for title and status
        for line in lines[:50]:  # Check first 50 lines
            line = line.strip()

            # Check for title (# Title)
            if line.startswith("# ") and title == spec_id:
                title = line[2:].strip()

            # Check for status field
            if line.lower().startswith("status:"):
                spec_status = line.split(":", 1)[1].strip().lower()
            elif line.lower().startswith("- status:"):
                spec_status = line.split(":", 1)[1].strip().lower()

        # Get file timestamps
        stat = file_path.stat()

        return SpecStatus(
            spec_id=spec_id,
            title=title,
            status=spec_status,
            created_at=datetime.fromtimestamp(stat.st_ctime),
            updated_at=datetime.fromtimestamp(stat.st_mtime),
            file_path=str(file_path),
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
    spec_files = _find_spec_files(base_path)

    specs = []
    for file_path in spec_files:
        spec = _parse_spec_file(file_path)
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
    spec_files = _find_spec_files(base_path)

    for file_path in spec_files:
        if file_path.stem == spec_id:
            spec = _parse_spec_file(file_path)
            if spec:
                return spec

    raise HTTPException(
        status_code=status.HTTP_404_NOT_FOUND,
        detail=f"SPEC '{spec_id}' not found",
    )
