"""Sessions Router

REST API endpoints for session management including
create, read, update, and delete operations.
"""

import uuid
from datetime import UTC, datetime
from typing import Optional

from fastapi import APIRouter, HTTPException, status

from moai_adk.web.database import get_database
from moai_adk.web.models.session import (
    Session,
    SessionCreate,
    SessionList,
    SessionUpdate,
)

router = APIRouter()


def _utcnow() -> datetime:
    """Get current UTC time (timezone-aware)"""
    return datetime.now(UTC)


@router.get("/sessions", response_model=SessionList)
async def list_sessions(
    limit: int = 50,
    offset: int = 0,
) -> SessionList:
    """List all sessions with pagination

    Args:
        limit: Maximum number of sessions to return
        offset: Number of sessions to skip

    Returns:
        SessionList with sessions and total count
    """
    db = await get_database()

    # Get total count
    cursor = await db.execute("SELECT COUNT(*) FROM sessions")
    row = await cursor.fetchone()
    total = row[0] if row else 0

    # Get sessions with pagination
    cursor = await db.execute(
        """
        SELECT id, title, provider, model, created_at, updated_at
        FROM sessions
        ORDER BY updated_at DESC
        LIMIT ? OFFSET ?
        """,
        (limit, offset),
    )
    rows = await cursor.fetchall()

    sessions = [
        Session(
            id=row["id"],
            title=row["title"],
            provider=row["provider"],
            model=row["model"],
            created_at=datetime.fromisoformat(row["created_at"]),
            updated_at=datetime.fromisoformat(row["updated_at"]),
        )
        for row in rows
    ]

    return SessionList(sessions=sessions, total=total)


@router.post("/sessions", response_model=Session, status_code=status.HTTP_201_CREATED)
async def create_session(session_data: Optional[SessionCreate] = None) -> Session:
    """Create a new session

    Args:
        session_data: Optional session creation data

    Returns:
        The created session
    """
    if session_data is None:
        session_data = SessionCreate()

    db = await get_database()
    session_id = str(uuid.uuid4())
    now = _utcnow()

    await db.execute(
        """
        INSERT INTO sessions (id, title, provider, model, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?)
        """,
        (
            session_id,
            session_data.title,
            session_data.provider,
            session_data.model,
            now.isoformat(),
            now.isoformat(),
        ),
    )
    await db.commit()

    return Session(
        id=session_id,
        title=session_data.title,
        provider=session_data.provider,
        model=session_data.model,
        created_at=now,
        updated_at=now,
    )


@router.get("/sessions/{session_id}", response_model=Session)
async def get_session(session_id: str) -> Session:
    """Get a specific session by ID

    Args:
        session_id: The session ID to retrieve

    Returns:
        The requested session

    Raises:
        HTTPException: If session not found
    """
    db = await get_database()

    cursor = await db.execute(
        """
        SELECT id, title, provider, model, created_at, updated_at
        FROM sessions
        WHERE id = ?
        """,
        (session_id,),
    )
    row = await cursor.fetchone()

    if row is None:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Session {session_id} not found",
        )

    return Session(
        id=row["id"],
        title=row["title"],
        provider=row["provider"],
        model=row["model"],
        created_at=datetime.fromisoformat(row["created_at"]),
        updated_at=datetime.fromisoformat(row["updated_at"]),
    )


@router.patch("/sessions/{session_id}", response_model=Session)
async def update_session(session_id: str, session_data: SessionUpdate) -> Session:
    """Update an existing session

    Args:
        session_id: The session ID to update
        session_data: Fields to update

    Returns:
        The updated session

    Raises:
        HTTPException: If session not found
    """
    db = await get_database()

    # Check if session exists
    cursor = await db.execute("SELECT id FROM sessions WHERE id = ?", (session_id,))
    if await cursor.fetchone() is None:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Session {session_id} not found",
        )

    # Build update query dynamically
    updates = []
    params = []

    if session_data.title is not None:
        updates.append("title = ?")
        params.append(session_data.title)

    if session_data.provider is not None:
        updates.append("provider = ?")
        params.append(session_data.provider)

    if session_data.model is not None:
        updates.append("model = ?")
        params.append(session_data.model)

    if updates:
        updates.append("updated_at = ?")
        params.append(_utcnow().isoformat())
        params.append(session_id)

        await db.execute(
            f"UPDATE sessions SET {', '.join(updates)} WHERE id = ?",
            params,
        )
        await db.commit()

    return await get_session(session_id)


@router.delete("/sessions/{session_id}", status_code=status.HTTP_204_NO_CONTENT)
async def delete_session(session_id: str) -> None:
    """Delete a session and all its messages

    Args:
        session_id: The session ID to delete

    Raises:
        HTTPException: If session not found
    """
    db = await get_database()

    # Check if session exists
    cursor = await db.execute("SELECT id FROM sessions WHERE id = ?", (session_id,))
    if await cursor.fetchone() is None:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Session {session_id} not found",
        )

    # Delete session (messages will be deleted via CASCADE)
    await db.execute("DELETE FROM sessions WHERE id = ?", (session_id,))
    await db.commit()
