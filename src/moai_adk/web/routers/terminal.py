"""Terminal Router

REST API and WebSocket endpoints for terminal management including
create, list, resize, and delete operations, plus real-time terminal I/O.
"""

import asyncio
import json
from typing import Optional

from fastapi import APIRouter, HTTPException, WebSocket, WebSocketDisconnect, status

from moai_adk.web.models.terminal import (
    TerminalCreate,
    TerminalList,
    TerminalResize,
    TerminalSession,
)
from moai_adk.web.services.terminal_service import TerminalManager

router = APIRouter()

# Singleton terminal manager instance
_terminal_manager: Optional[TerminalManager] = None


def get_terminal_manager() -> TerminalManager:
    """Get or create the singleton TerminalManager instance.

    Returns:
        The global TerminalManager instance
    """
    global _terminal_manager
    if _terminal_manager is None:
        _terminal_manager = TerminalManager(max_terminals=6)
    return _terminal_manager


@router.get("/terminals", response_model=TerminalList)
async def list_terminals() -> TerminalList:
    """List all active terminal sessions.

    Returns:
        TerminalList with terminals and total count
    """
    manager = get_terminal_manager()
    terminals = await manager.list_sessions()

    return TerminalList(terminals=terminals, total=len(terminals))


@router.post("/terminals", response_model=TerminalSession, status_code=status.HTTP_201_CREATED)
async def create_terminal(terminal_data: Optional[TerminalCreate] = None) -> TerminalSession:
    """Create a new terminal session.

    Args:
        terminal_data: Optional terminal creation data

    Returns:
        The created terminal session

    Raises:
        HTTPException: If maximum terminals reached or PTY unavailable
    """
    if terminal_data is None:
        terminal_data = TerminalCreate()

    manager = get_terminal_manager()

    try:
        session = await manager.create_terminal(
            spec_id=terminal_data.spec_id,
            worktree_path=terminal_data.worktree_path,
            initial_command=terminal_data.initial_command,
            cols=terminal_data.cols,
            rows=terminal_data.rows,
        )
        return session
    except ValueError as e:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=str(e),
        )
    except RuntimeError as e:
        raise HTTPException(
            status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
            detail=str(e),
        )


@router.get("/terminals/{terminal_id}", response_model=TerminalSession)
async def get_terminal(terminal_id: str) -> TerminalSession:
    """Get a specific terminal session by ID.

    Args:
        terminal_id: The terminal session ID

    Returns:
        The requested terminal session

    Raises:
        HTTPException: If terminal not found
    """
    manager = get_terminal_manager()
    session = await manager.get_session(terminal_id)

    if session is None:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Terminal {terminal_id} not found",
        )

    return session


@router.delete("/terminals/{terminal_id}", status_code=status.HTTP_204_NO_CONTENT)
async def delete_terminal(terminal_id: str) -> None:
    """Close and delete a terminal session.

    Args:
        terminal_id: The terminal session ID

    Raises:
        HTTPException: If terminal not found
    """
    manager = get_terminal_manager()
    session = await manager.get_session(terminal_id)

    if session is None:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Terminal {terminal_id} not found",
        )

    await manager.close(terminal_id)


@router.post("/terminals/{terminal_id}/resize")
async def resize_terminal(terminal_id: str, resize_data: TerminalResize) -> dict:
    """Resize a terminal session.

    Args:
        terminal_id: The terminal session ID
        resize_data: New terminal dimensions

    Returns:
        Success confirmation

    Raises:
        HTTPException: If terminal not found
    """
    manager = get_terminal_manager()
    session = await manager.get_session(terminal_id)

    if session is None:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Terminal {terminal_id} not found",
        )

    try:
        await manager.resize(terminal_id, resize_data.cols, resize_data.rows)
        return {"status": "ok", "cols": resize_data.cols, "rows": resize_data.rows}
    except KeyError:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Terminal {terminal_id} not found",
        )


@router.websocket("/terminal/{terminal_id}")
async def websocket_terminal(websocket: WebSocket, terminal_id: str) -> None:
    """WebSocket endpoint for terminal I/O.

    Handles bidirectional communication between client and PTY:
    - Client sends input messages (keystrokes)
    - Server sends output messages (terminal output)
    - Supports resize and heartbeat messages

    Message format:
    {
        "type": "input" | "output" | "resize" | "heartbeat" | "close",
        "data": "string content for input/output",
        "cols": int (for resize),
        "rows": int (for resize)
    }

    Args:
        websocket: The WebSocket connection
        terminal_id: The terminal session ID to connect to
    """
    manager = get_terminal_manager()
    session = await manager.get_session(terminal_id)

    if session is None:
        await websocket.close(code=status.WS_1008_POLICY_VIOLATION)
        return

    await websocket.accept()

    # Background task for reading PTY output
    async def read_pty_output():
        """Read output from PTY and send to client."""
        try:
            while True:
                if not manager.is_alive(terminal_id):
                    await websocket.send_json({"type": "close", "data": "Terminal exited"})
                    break

                try:
                    output = await manager.read(terminal_id)
                    if output:
                        await websocket.send_json(
                            {
                                "type": "output",
                                "data": output.decode("utf-8", errors="replace"),
                            }
                        )
                except KeyError:
                    # Terminal closed
                    break
                except Exception:
                    # Ignore read errors, continue trying
                    pass

                await asyncio.sleep(0.01)  # Small delay to prevent busy loop
        except Exception:
            pass  # Connection closed

    # Start PTY output reader
    reader_task = asyncio.create_task(read_pty_output())

    try:
        while True:
            # Receive message from client
            data = await websocket.receive_text()

            try:
                message = json.loads(data)
                msg_type = message.get("type", "input")

                if msg_type == "input":
                    # Write input to PTY
                    input_data = message.get("data", "")
                    if input_data:
                        await manager.write(terminal_id, input_data)

                elif msg_type == "resize":
                    # Resize terminal
                    cols = message.get("cols", 80)
                    rows = message.get("rows", 24)
                    await manager.resize(terminal_id, cols, rows)

                elif msg_type == "heartbeat":
                    # Respond to heartbeat
                    await websocket.send_json({"type": "heartbeat"})

                elif msg_type == "close":
                    # Close terminal
                    await manager.close(terminal_id)
                    break

            except json.JSONDecodeError:
                await websocket.send_json(
                    {
                        "type": "error",
                        "data": "Invalid JSON message",
                    }
                )

    except WebSocketDisconnect:
        pass
    finally:
        # Cancel the reader task
        reader_task.cancel()
        try:
            await reader_task
        except asyncio.CancelledError:
            pass
