"""Session Model

Pydantic models for chat session data including
session creation, update, and response schemas.
"""

from datetime import datetime
from typing import Optional

from pydantic import BaseModel, ConfigDict, Field


class SessionBase(BaseModel):
    """Base session model with common fields"""

    title: str = Field(default="New Session", description="Session title")
    provider: str = Field(default="claude", description="AI provider name")
    model: str = Field(default="claude-sonnet-4-20250514", description="Model identifier")


class SessionCreate(SessionBase):
    """Schema for creating a new session"""

    pass


class SessionUpdate(BaseModel):
    """Schema for updating an existing session"""

    title: Optional[str] = Field(default=None, description="Updated session title")
    provider: Optional[str] = Field(default=None, description="Updated AI provider")
    model: Optional[str] = Field(default=None, description="Updated model identifier")


class Session(SessionBase):
    """Complete session model with all fields"""

    model_config = ConfigDict(from_attributes=True)

    id: str = Field(..., description="Unique session identifier")
    created_at: datetime = Field(..., description="Session creation timestamp")
    updated_at: datetime = Field(..., description="Last update timestamp")


class SessionList(BaseModel):
    """Response model for listing sessions"""

    sessions: list[Session] = Field(default_factory=list, description="List of sessions")
    total: int = Field(default=0, description="Total number of sessions")
