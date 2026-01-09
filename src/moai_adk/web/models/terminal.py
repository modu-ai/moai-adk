"""Terminal Models

Pydantic models for terminal session management including
session state, message types, and API request/response schemas.
"""

from datetime import datetime
from typing import Literal

from pydantic import BaseModel, ConfigDict, Field


class TerminalSession(BaseModel):
    """Complete terminal session model with all fields

    Represents a PTY terminal session that can be used for
    parallel Claude Code execution in worktrees.
    """

    model_config = ConfigDict(from_attributes=True)

    id: str = Field(..., description="Unique terminal session identifier")
    spec_id: str | None = Field(
        default=None,
        description="Associated SPEC identifier (e.g., SPEC-001)",
    )
    worktree_path: str | None = Field(
        default=None,
        description="Git worktree path for this terminal session",
    )
    pid: int | None = Field(
        default=None,
        description="Process ID of the PTY subprocess",
    )
    status: Literal["pending", "running", "completed", "error"] = Field(
        default="pending",
        description="Terminal session status",
    )
    created_at: datetime = Field(..., description="Session creation timestamp")
    completed_at: datetime | None = Field(
        default=None,
        description="Session completion timestamp",
    )


class TerminalMessage(BaseModel):
    """Terminal WebSocket message model

    Represents messages exchanged between the client and server
    over the terminal WebSocket connection.
    """

    type: Literal["input", "output", "resize", "heartbeat", "close"] = Field(
        ...,
        description="Message type",
    )
    data: str | None = Field(
        default=None,
        description="Message data (input/output content)",
    )
    cols: int | None = Field(
        default=None,
        description="Terminal columns for resize messages",
    )
    rows: int | None = Field(
        default=None,
        description="Terminal rows for resize messages",
    )


class TerminalCreate(BaseModel):
    """Schema for creating a new terminal session

    Used as request body for POST /api/terminals endpoint.
    """

    spec_id: str | None = Field(
        default=None,
        description="Associated SPEC identifier",
    )
    worktree_path: str | None = Field(
        default=None,
        description="Git worktree path for the terminal",
    )
    initial_command: str | None = Field(
        default=None,
        description="Initial command to execute (e.g., 'claude /moai:2-run SPEC-001')",
    )
    cols: int = Field(
        default=80,
        description="Terminal width in columns",
    )
    rows: int = Field(
        default=24,
        description="Terminal height in rows",
    )


class TerminalResize(BaseModel):
    """Schema for resizing a terminal

    Used as request body for POST /api/terminals/{terminal_id}/resize endpoint.
    """

    cols: int = Field(..., description="New terminal width in columns")
    rows: int = Field(..., description="New terminal height in rows")


class TerminalList(BaseModel):
    """Response model for listing terminal sessions"""

    terminals: list[TerminalSession] = Field(
        default_factory=list,
        description="List of terminal sessions",
    )
    total: int = Field(default=0, description="Total number of terminals")
