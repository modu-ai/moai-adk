"""Terminal Manager Service

Manages multiple PTY sessions for parallel terminal execution.
Provides session lifecycle management, I/O operations, and resize functionality.
"""

import asyncio
import os
import uuid
from datetime import UTC, datetime
from typing import Any

from moai_adk.web.models.terminal import TerminalSession

# PTY support is platform-dependent
try:
    import ptyprocess

    PTY_AVAILABLE = True
except ImportError:
    PTY_AVAILABLE = False


class TerminalManager:
    """Manages multiple PTY sessions for parallel execution.

    This class provides functionality to:
    - Create new terminal sessions with PTY
    - Read from and write to terminal sessions
    - Resize terminal dimensions
    - Close individual or all sessions

    Attributes:
        sessions: Dictionary mapping session IDs to TerminalSession objects
        pty_processes: Dictionary mapping session IDs to PTY process objects
        max_terminals: Maximum number of concurrent terminals allowed
    """

    def __init__(self, max_terminals: int = 6):
        """Initialize the TerminalManager.

        Args:
            max_terminals: Maximum number of concurrent terminals allowed.
                          Defaults to 6 per SPEC requirements.
        """
        self.sessions: dict[str, TerminalSession] = {}
        self.pty_processes: dict[str, Any] = {}
        self.max_terminals = max_terminals

    async def create_terminal(
        self,
        spec_id: str | None = None,
        worktree_path: str | None = None,
        initial_command: str | None = None,
        cols: int = 80,
        rows: int = 24,
    ) -> TerminalSession:
        """Create a new terminal session with PTY.

        Args:
            spec_id: Optional SPEC identifier associated with this terminal
            worktree_path: Optional git worktree path for the terminal's cwd
            initial_command: Optional command to execute after terminal starts
            cols: Terminal width in columns (default: 80)
            rows: Terminal height in rows (default: 24)

        Returns:
            The created TerminalSession object

        Raises:
            ValueError: If maximum number of terminals has been reached
            RuntimeError: If PTY is not available on this platform
        """
        # Check terminal limit
        if len(self.sessions) >= self.max_terminals:
            raise ValueError(
                f"Maximum {self.max_terminals} terminals reached. Close existing terminals before creating new ones."
            )

        if not PTY_AVAILABLE:
            raise RuntimeError("PTY support not available. Install ptyprocess: pip install ptyprocess")

        # Generate unique session ID
        session_id = str(uuid.uuid4())

        # Determine working directory
        cwd = worktree_path if worktree_path and os.path.isdir(worktree_path) else None

        # Create PTY process
        shell = os.environ.get("SHELL", "/bin/bash")
        pty = ptyprocess.PtyProcess.spawn(
            [shell],
            dimensions=(rows, cols),
            cwd=cwd,
        )

        # Create session object
        session = TerminalSession(
            id=session_id,
            spec_id=spec_id,
            worktree_path=worktree_path,
            pid=pty.pid,
            status="running",
            created_at=datetime.now(UTC),
        )

        # Store session and PTY process
        self.sessions[session_id] = session
        self.pty_processes[session_id] = pty

        # Execute initial command if provided
        if initial_command:
            await self.write(session_id, f"{initial_command}\n")

        return session

    async def read(self, session_id: str, size: int = 4096) -> bytes:
        """Read output from a terminal session.

        Args:
            session_id: The terminal session ID
            size: Maximum number of bytes to read (default: 4096)

        Returns:
            The output bytes from the terminal

        Raises:
            KeyError: If session_id does not exist
        """
        if session_id not in self.pty_processes:
            raise KeyError(f"Terminal session not found: {session_id}")

        pty = self.pty_processes[session_id]

        # Non-blocking read with timeout
        loop = asyncio.get_event_loop()
        try:
            # Use executor for non-blocking PTY read
            output = await asyncio.wait_for(
                loop.run_in_executor(None, lambda: pty.read(size)),
                timeout=0.1,
            )
            return output
        except asyncio.TimeoutError:
            return b""
        except EOFError:
            # Terminal closed
            return b""

    async def write(self, session_id: str, data: str) -> None:
        """Write input to a terminal session.

        Args:
            session_id: The terminal session ID
            data: The input string to write

        Raises:
            KeyError: If session_id does not exist
        """
        if session_id not in self.pty_processes:
            raise KeyError(f"Terminal session not found: {session_id}")

        pty = self.pty_processes[session_id]

        # Write to PTY (ptyprocess expects bytes)
        loop = asyncio.get_event_loop()
        data_bytes = data.encode("utf-8") if isinstance(data, str) else data
        await loop.run_in_executor(None, lambda: pty.write(data_bytes))

    async def resize(self, session_id: str, cols: int, rows: int) -> None:
        """Resize a terminal session.

        Args:
            session_id: The terminal session ID
            cols: New width in columns
            rows: New height in rows

        Raises:
            KeyError: If session_id does not exist
        """
        if session_id not in self.pty_processes:
            raise KeyError(f"Terminal session not found: {session_id}")

        pty = self.pty_processes[session_id]
        pty.setwinsize(rows, cols)

    async def close(self, session_id: str) -> None:
        """Close a terminal session.

        This terminates the PTY process and removes the session from tracking.
        Safe to call on non-existent sessions (no-op).

        Args:
            session_id: The terminal session ID
        """
        if session_id not in self.sessions:
            return

        # Terminate PTY process if exists
        if session_id in self.pty_processes:
            pty = self.pty_processes[session_id]
            try:
                if pty.isalive():
                    pty.terminate(force=True)
            except Exception:
                pass  # Ignore termination errors
            del self.pty_processes[session_id]

        # Update session status
        session = self.sessions[session_id]
        session.status = "completed"
        session.completed_at = datetime.now(UTC)

        # Remove session
        del self.sessions[session_id]

    async def close_all(self) -> None:
        """Close all terminal sessions.

        Terminates all PTY processes and clears session tracking.
        """
        session_ids = list(self.sessions.keys())
        for session_id in session_ids:
            await self.close(session_id)

    async def get_session(self, session_id: str) -> TerminalSession | None:
        """Get a terminal session by ID.

        Args:
            session_id: The terminal session ID

        Returns:
            The TerminalSession object or None if not found
        """
        return self.sessions.get(session_id)

    async def list_sessions(self) -> list[TerminalSession]:
        """List all active terminal sessions.

        Returns:
            List of all active TerminalSession objects
        """
        return list(self.sessions.values())

    def is_alive(self, session_id: str) -> bool:
        """Check if a terminal session's PTY process is still alive.

        Args:
            session_id: The terminal session ID

        Returns:
            True if the PTY process is alive, False otherwise
        """
        if session_id not in self.pty_processes:
            return False

        pty = self.pty_processes[session_id]
        return pty.isalive()
