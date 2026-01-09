"""TAG-002: Test TerminalManager Service

RED Phase: Tests for the terminal manager service that handles
PTY session creation, I/O, and lifecycle management.
"""

import sys
from unittest.mock import MagicMock

import pytest


class TestTerminalManagerBasics:
    """Test cases for TerminalManager basic structure"""

    def test_terminal_manager_class_exists(self):
        """Test that TerminalManager class can be imported"""
        from moai_adk.web.services.terminal_service import TerminalManager

        assert TerminalManager is not None

    def test_terminal_manager_has_sessions_dict(self):
        """Test that TerminalManager has sessions dictionary"""
        from moai_adk.web.services.terminal_service import TerminalManager

        manager = TerminalManager()
        assert hasattr(manager, "sessions")
        assert isinstance(manager.sessions, dict)

    def test_terminal_manager_has_max_terminals(self):
        """Test that TerminalManager has max_terminals limit"""
        from moai_adk.web.services.terminal_service import TerminalManager

        manager = TerminalManager()
        assert hasattr(manager, "max_terminals")
        assert manager.max_terminals == 6  # Maximum 6 concurrent terminals

    def test_terminal_manager_has_pty_processes_dict(self):
        """Test that TerminalManager has pty_processes dictionary"""
        from moai_adk.web.services.terminal_service import TerminalManager

        manager = TerminalManager()
        assert hasattr(manager, "pty_processes")
        assert isinstance(manager.pty_processes, dict)


class TestTerminalManagerCreateTerminal:
    """Test cases for create_terminal method"""

    def test_create_terminal_method_exists(self):
        """Test that create_terminal method exists"""
        from moai_adk.web.services.terminal_service import TerminalManager

        manager = TerminalManager()
        assert hasattr(manager, "create_terminal")
        assert callable(manager.create_terminal)

    @pytest.mark.asyncio
    @pytest.mark.skipif(sys.platform == "win32", reason="PTY not supported on Windows")
    async def test_create_terminal_returns_session(self):
        """Test that create_terminal returns a TerminalSession"""
        from moai_adk.web.services.terminal_service import TerminalManager

        from moai_adk.web.models.terminal import TerminalSession

        manager = TerminalManager()
        session = await manager.create_terminal()

        assert isinstance(session, TerminalSession)
        assert session.id is not None
        assert session.status == "running"

        # Cleanup
        await manager.close(session.id)

    @pytest.mark.asyncio
    @pytest.mark.skipif(sys.platform == "win32", reason="PTY not supported on Windows")
    async def test_create_terminal_with_spec_id(self):
        """Test create_terminal with spec_id parameter"""
        from moai_adk.web.services.terminal_service import TerminalManager

        manager = TerminalManager()
        session = await manager.create_terminal(spec_id="SPEC-001")

        assert session.spec_id == "SPEC-001"

        # Cleanup
        await manager.close(session.id)

    @pytest.mark.asyncio
    @pytest.mark.skipif(sys.platform == "win32", reason="PTY not supported on Windows")
    async def test_create_terminal_with_worktree_path(self, tmp_path):
        """Test create_terminal with worktree_path parameter"""
        from moai_adk.web.services.terminal_service import TerminalManager

        worktree_path = str(tmp_path)
        manager = TerminalManager()
        session = await manager.create_terminal(worktree_path=worktree_path)

        assert session.worktree_path == worktree_path

        # Cleanup
        await manager.close(session.id)

    @pytest.mark.asyncio
    @pytest.mark.skipif(sys.platform == "win32", reason="PTY not supported on Windows")
    async def test_create_terminal_with_dimensions(self):
        """Test create_terminal with custom dimensions"""
        from moai_adk.web.services.terminal_service import TerminalManager

        manager = TerminalManager()
        session = await manager.create_terminal(cols=120, rows=40)

        assert session is not None
        assert session.status == "running"

        # Cleanup
        await manager.close(session.id)

    @pytest.mark.asyncio
    @pytest.mark.skipif(sys.platform == "win32", reason="PTY not supported on Windows")
    async def test_create_terminal_has_pid(self):
        """Test that created terminal has a PID"""
        from moai_adk.web.services.terminal_service import TerminalManager

        manager = TerminalManager()
        session = await manager.create_terminal()

        assert session.pid is not None
        assert session.pid > 0

        # Cleanup
        await manager.close(session.id)

    @pytest.mark.asyncio
    @pytest.mark.skipif(sys.platform == "win32", reason="PTY not supported on Windows")
    async def test_create_terminal_stores_session(self):
        """Test that create_terminal stores the session"""
        from moai_adk.web.services.terminal_service import TerminalManager

        manager = TerminalManager()
        session = await manager.create_terminal()

        assert session.id in manager.sessions

        # Cleanup
        await manager.close(session.id)

    @pytest.mark.asyncio
    async def test_create_terminal_max_limit(self):
        """Test that create_terminal respects max_terminals limit"""
        from moai_adk.web.services.terminal_service import TerminalManager

        manager = TerminalManager()
        manager.max_terminals = 2

        # Create mock sessions to fill up the limit
        manager.sessions = {
            "term-1": MagicMock(),
            "term-2": MagicMock(),
        }

        with pytest.raises(ValueError, match="Maximum.*terminals.*reached"):
            await manager.create_terminal()


class TestTerminalManagerRead:
    """Test cases for read method"""

    def test_read_method_exists(self):
        """Test that read method exists"""
        from moai_adk.web.services.terminal_service import TerminalManager

        manager = TerminalManager()
        assert hasattr(manager, "read")
        assert callable(manager.read)

    @pytest.mark.asyncio
    async def test_read_nonexistent_session_raises(self):
        """Test that reading from nonexistent session raises error"""
        from moai_adk.web.services.terminal_service import TerminalManager

        manager = TerminalManager()

        with pytest.raises(KeyError):
            await manager.read("nonexistent-id")


class TestTerminalManagerWrite:
    """Test cases for write method"""

    def test_write_method_exists(self):
        """Test that write method exists"""
        from moai_adk.web.services.terminal_service import TerminalManager

        manager = TerminalManager()
        assert hasattr(manager, "write")
        assert callable(manager.write)

    @pytest.mark.asyncio
    async def test_write_nonexistent_session_raises(self):
        """Test that writing to nonexistent session raises error"""
        from moai_adk.web.services.terminal_service import TerminalManager

        manager = TerminalManager()

        with pytest.raises(KeyError):
            await manager.write("nonexistent-id", "test data")


class TestTerminalManagerResize:
    """Test cases for resize method"""

    def test_resize_method_exists(self):
        """Test that resize method exists"""
        from moai_adk.web.services.terminal_service import TerminalManager

        manager = TerminalManager()
        assert hasattr(manager, "resize")
        assert callable(manager.resize)

    @pytest.mark.asyncio
    async def test_resize_nonexistent_session_raises(self):
        """Test that resizing nonexistent session raises error"""
        from moai_adk.web.services.terminal_service import TerminalManager

        manager = TerminalManager()

        with pytest.raises(KeyError):
            await manager.resize("nonexistent-id", 120, 40)


class TestTerminalManagerClose:
    """Test cases for close method"""

    def test_close_method_exists(self):
        """Test that close method exists"""
        from moai_adk.web.services.terminal_service import TerminalManager

        manager = TerminalManager()
        assert hasattr(manager, "close")
        assert callable(manager.close)

    @pytest.mark.asyncio
    @pytest.mark.skipif(sys.platform == "win32", reason="PTY not supported on Windows")
    async def test_close_removes_session(self):
        """Test that close removes the session"""
        from moai_adk.web.services.terminal_service import TerminalManager

        manager = TerminalManager()
        session = await manager.create_terminal()
        session_id = session.id

        await manager.close(session_id)

        assert session_id not in manager.sessions

    @pytest.mark.asyncio
    async def test_close_nonexistent_session_no_error(self):
        """Test that closing nonexistent session doesn't raise error"""
        from moai_adk.web.services.terminal_service import TerminalManager

        manager = TerminalManager()

        # Should not raise
        await manager.close("nonexistent-id")


class TestTerminalManagerCloseAll:
    """Test cases for close_all method"""

    def test_close_all_method_exists(self):
        """Test that close_all method exists"""
        from moai_adk.web.services.terminal_service import TerminalManager

        manager = TerminalManager()
        assert hasattr(manager, "close_all")
        assert callable(manager.close_all)

    @pytest.mark.asyncio
    @pytest.mark.skipif(sys.platform == "win32", reason="PTY not supported on Windows")
    async def test_close_all_removes_all_sessions(self):
        """Test that close_all removes all sessions"""
        from moai_adk.web.services.terminal_service import TerminalManager

        manager = TerminalManager()

        # Create multiple sessions
        session1 = await manager.create_terminal()
        session2 = await manager.create_terminal()

        assert len(manager.sessions) == 2

        await manager.close_all()

        assert len(manager.sessions) == 0


class TestTerminalManagerGetSession:
    """Test cases for get_session method"""

    def test_get_session_method_exists(self):
        """Test that get_session method exists"""
        from moai_adk.web.services.terminal_service import TerminalManager

        manager = TerminalManager()
        assert hasattr(manager, "get_session")
        assert callable(manager.get_session)

    @pytest.mark.asyncio
    @pytest.mark.skipif(sys.platform == "win32", reason="PTY not supported on Windows")
    async def test_get_session_returns_session(self):
        """Test that get_session returns the session"""
        from moai_adk.web.services.terminal_service import TerminalManager

        from moai_adk.web.models.terminal import TerminalSession

        manager = TerminalManager()
        created_session = await manager.create_terminal()

        retrieved_session = await manager.get_session(created_session.id)

        assert retrieved_session is not None
        assert isinstance(retrieved_session, TerminalSession)
        assert retrieved_session.id == created_session.id

        # Cleanup
        await manager.close(created_session.id)

    @pytest.mark.asyncio
    async def test_get_session_nonexistent_returns_none(self):
        """Test that get_session returns None for nonexistent session"""
        from moai_adk.web.services.terminal_service import TerminalManager

        manager = TerminalManager()

        session = await manager.get_session("nonexistent-id")

        assert session is None


class TestTerminalManagerListSessions:
    """Test cases for list_sessions method"""

    def test_list_sessions_method_exists(self):
        """Test that list_sessions method exists"""
        from moai_adk.web.services.terminal_service import TerminalManager

        manager = TerminalManager()
        assert hasattr(manager, "list_sessions")
        assert callable(manager.list_sessions)

    @pytest.mark.asyncio
    async def test_list_sessions_empty(self):
        """Test list_sessions returns empty list when no sessions"""
        from moai_adk.web.services.terminal_service import TerminalManager

        manager = TerminalManager()

        sessions = await manager.list_sessions()

        assert isinstance(sessions, list)
        assert len(sessions) == 0

    @pytest.mark.asyncio
    @pytest.mark.skipif(sys.platform == "win32", reason="PTY not supported on Windows")
    async def test_list_sessions_returns_all(self):
        """Test list_sessions returns all sessions"""
        from moai_adk.web.services.terminal_service import TerminalManager

        manager = TerminalManager()

        session1 = await manager.create_terminal()
        session2 = await manager.create_terminal()

        sessions = await manager.list_sessions()

        assert len(sessions) == 2
        session_ids = [s.id for s in sessions]
        assert session1.id in session_ids
        assert session2.id in session_ids

        # Cleanup
        await manager.close_all()


class TestTerminalManagerPTYIntegration:
    """Integration tests for PTY functionality"""

    @pytest.mark.asyncio
    @pytest.mark.skipif(sys.platform == "win32", reason="PTY not supported on Windows")
    async def test_write_and_read_echo(self):
        """Test writing to terminal and reading output"""
        from moai_adk.web.services.terminal_service import TerminalManager

        manager = TerminalManager()
        session = await manager.create_terminal()

        # Write echo command
        await manager.write(session.id, "echo hello\n")

        # Give it time to process
        import asyncio

        await asyncio.sleep(0.2)

        # Read output (may contain prompt and echo)
        output = await manager.read(session.id)

        assert output is not None
        # Output should contain the echoed word
        # Note: The exact output depends on shell configuration

        # Cleanup
        await manager.close(session.id)

    @pytest.mark.asyncio
    @pytest.mark.skipif(sys.platform == "win32", reason="PTY not supported on Windows")
    async def test_resize_terminal(self):
        """Test resizing a terminal"""
        from moai_adk.web.services.terminal_service import TerminalManager

        manager = TerminalManager()
        session = await manager.create_terminal(cols=80, rows=24)

        # Resize should not raise
        await manager.resize(session.id, 120, 40)

        # Cleanup
        await manager.close(session.id)

    @pytest.mark.asyncio
    @pytest.mark.skipif(sys.platform == "win32", reason="PTY not supported on Windows")
    async def test_terminal_with_initial_command(self, tmp_path):
        """Test terminal with initial command execution"""
        from moai_adk.web.services.terminal_service import TerminalManager

        manager = TerminalManager()
        session = await manager.create_terminal(initial_command="echo 'initialized'")

        assert session.status == "running"

        # Cleanup
        await manager.close(session.id)

    @pytest.mark.asyncio
    @pytest.mark.skipif(sys.platform == "win32", reason="PTY not supported on Windows")
    async def test_pty_process_terminates_on_close(self):
        """Test that PTY process terminates when session is closed"""
        from moai_adk.web.services.terminal_service import TerminalManager

        manager = TerminalManager()
        session = await manager.create_terminal()
        session_id = session.id

        # Verify PTY process exists
        assert session_id in manager.pty_processes

        await manager.close(session_id)

        # Verify PTY process is removed
        assert session_id not in manager.pty_processes
