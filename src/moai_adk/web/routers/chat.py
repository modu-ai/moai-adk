"""Chat WebSocket Router

WebSocket endpoint for real-time chat communication
with Claude agent integration and streaming responses.
"""

import json
import uuid
from datetime import UTC, datetime
from typing import Optional

from fastapi import APIRouter, WebSocket, WebSocketDisconnect, status

from moai_adk.web.database import get_database
from moai_adk.web.models.message import MessageRole, WebSocketMessage
from moai_adk.web.services.agent_service import AgentService

router = APIRouter()


def _utcnow() -> datetime:
    """Get current UTC time (timezone-aware)"""
    return datetime.now(UTC)


class ConnectionManager:
    """Manages WebSocket connections for chat sessions"""

    def __init__(self):
        self.active_connections: dict[str, WebSocket] = {}

    async def connect(self, websocket: WebSocket, session_id: str) -> bool:
        """Accept and register a WebSocket connection

        Args:
            websocket: The WebSocket connection
            session_id: The session ID for this connection

        Returns:
            True if connection was accepted, False otherwise
        """
        # Verify session exists
        try:
            db = await get_database()
            cursor = await db.execute("SELECT id FROM sessions WHERE id = ?", (session_id,))
            if await cursor.fetchone() is None:
                await websocket.close(code=status.WS_1008_POLICY_VIOLATION)
                return False
        except Exception:
            await websocket.close(code=status.WS_1011_INTERNAL_ERROR)
            return False

        await websocket.accept()
        self.active_connections[session_id] = websocket
        return True

    def disconnect(self, session_id: str) -> None:
        """Remove a WebSocket connection

        Args:
            session_id: The session ID to disconnect
        """
        self.active_connections.pop(session_id, None)

    async def send_message(self, session_id: str, message: WebSocketMessage) -> None:
        """Send a message to a specific session

        Args:
            session_id: The target session ID
            message: The message to send
        """
        if session_id in self.active_connections:
            await self.active_connections[session_id].send_json(message.model_dump(mode="json"))

    def get_connection(self, session_id: str) -> Optional[WebSocket]:
        """Get the WebSocket connection for a session

        Args:
            session_id: The session ID

        Returns:
            The WebSocket connection or None
        """
        return self.active_connections.get(session_id)


# Global connection manager
manager = ConnectionManager()


async def save_message(
    session_id: str,
    role: MessageRole,
    content: str,
) -> str:
    """Save a message to the database

    Args:
        session_id: The session ID
        role: The message role
        content: The message content

    Returns:
        The created message ID
    """
    db = await get_database()
    message_id = str(uuid.uuid4())

    await db.execute(
        """
        INSERT INTO messages (id, session_id, role, content, created_at)
        VALUES (?, ?, ?, ?, ?)
        """,
        (message_id, session_id, role.value, content, _utcnow().isoformat()),
    )

    # Update session's updated_at timestamp
    await db.execute(
        "UPDATE sessions SET updated_at = ? WHERE id = ?",
        (_utcnow().isoformat(), session_id),
    )

    await db.commit()
    return message_id


@router.websocket("/chat/{session_id}")
async def websocket_chat(websocket: WebSocket, session_id: str) -> None:
    """WebSocket endpoint for chat communication

    Args:
        websocket: The WebSocket connection
        session_id: The session ID for this chat
    """
    # Attempt to connect
    if not await manager.connect(websocket, session_id):
        return

    agent_service = AgentService()

    try:
        # Send connection confirmation
        await manager.send_message(
            session_id,
            WebSocketMessage(
                type="system",
                content="Connected to chat session",
                timestamp=_utcnow(),
                session_id=session_id,
            ),
        )

        while True:
            # Receive message from client
            data = await websocket.receive_text()

            try:
                message_data = json.loads(data)
                user_content = message_data.get("content", "")

                if not user_content:
                    continue

                # Save user message
                await save_message(session_id, MessageRole.USER, user_content)

                # Get AI response (streaming)
                full_response = ""
                async for chunk in agent_service.stream_response(
                    session_id=session_id,
                    message=user_content,
                ):
                    full_response += chunk
                    await manager.send_message(
                        session_id,
                        WebSocketMessage(
                            type="chat",
                            content=chunk,
                            timestamp=_utcnow(),
                            session_id=session_id,
                        ),
                    )

                # Save assistant response
                if full_response:
                    await save_message(session_id, MessageRole.ASSISTANT, full_response)

                # Send completion signal
                await manager.send_message(
                    session_id,
                    WebSocketMessage(
                        type="system",
                        content="[DONE]",
                        timestamp=_utcnow(),
                        session_id=session_id,
                    ),
                )

            except json.JSONDecodeError:
                await manager.send_message(
                    session_id,
                    WebSocketMessage(
                        type="error",
                        content="Invalid message format",
                        timestamp=_utcnow(),
                        session_id=session_id,
                    ),
                )

    except WebSocketDisconnect:
        manager.disconnect(session_id)
    except Exception as e:
        await manager.send_message(
            session_id,
            WebSocketMessage(
                type="error",
                content=str(e),
                timestamp=_utcnow(),
                session_id=session_id,
            ),
        )
        manager.disconnect(session_id)
