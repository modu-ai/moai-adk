"""Message Model

Pydantic models for chat message data including
message creation and response schemas.
"""

from datetime import datetime
from enum import Enum
from typing import Optional

from pydantic import BaseModel, ConfigDict, Field


class MessageRole(str, Enum):
    """Valid message roles"""

    USER = "user"
    ASSISTANT = "assistant"
    SYSTEM = "system"


class MessageBase(BaseModel):
    """Base message model with common fields"""

    role: MessageRole = Field(..., description="Message role (user/assistant/system)")
    content: str = Field(..., description="Message content")


class MessageCreate(MessageBase):
    """Schema for creating a new message"""

    session_id: str = Field(..., description="Parent session ID")


class Message(MessageBase):
    """Complete message model with all fields"""

    model_config = ConfigDict(from_attributes=True)

    id: str = Field(..., description="Unique message identifier")
    session_id: str = Field(..., description="Parent session ID")
    created_at: datetime = Field(..., description="Message creation timestamp")


class MessageList(BaseModel):
    """Response model for listing messages"""

    messages: list[Message] = Field(default_factory=list, description="List of messages")
    total: int = Field(default=0, description="Total number of messages")


class WebSocketMessage(BaseModel):
    """WebSocket message protocol"""

    type: str = Field(..., description="Message type (chat, system, error)")
    content: str = Field(..., description="Message content")
    timestamp: Optional[datetime] = Field(default=None, description="Message timestamp")
    session_id: Optional[str] = Field(default=None, description="Session ID")
