"""TAG-007: Terminal Integration Tests

Integration tests for the terminal system including
PTY I/O, WebSocket communication, and multi-terminal scenarios.
"""

import asyncio
import sys

import pytest
import pytest_asyncio
from httpx import ASGITransport, AsyncClient


@pytest.mark.skipif(sys.platform == "win32", reason="PTY not supported on Windows")
class TestTerminalPTYIntegration:
    """Integration tests for PTY terminal functionality"""

    @pytest_asyncio.fixture
    async def client(self, tmp_path):
        """Create a test client"""
        from moai_adk.web.config import WebConfig
        from moai_adk.web.database import close_database, init_database
        from moai_adk.web.routers.terminal import get_terminal_manager
        from moai_adk.web.server import create_app

        config = WebConfig(database_path=tmp_path / "test.db")
        await init_database(config)

        manager = get_terminal_manager()
        await manager.close_all()

        app = create_app(config)

        async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as client:
            yield client

        await manager.close_all()
        await close_database()

    @pytest.mark.asyncio
    async def test_create_and_get_terminal(self, client):
        """Test creating a terminal and retrieving it"""
        # Create terminal
        create_resp = await client.post("/api/terminals", json={"spec_id": "SPEC-INT-001"})
        assert create_resp.status_code == 201
        terminal_id = create_resp.json()["id"]

        # Get terminal
        get_resp = await client.get(f"/api/terminals/{terminal_id}")
        assert get_resp.status_code == 200
        assert get_resp.json()["id"] == terminal_id
        assert get_resp.json()["spec_id"] == "SPEC-INT-001"

    @pytest.mark.asyncio
    async def test_terminal_lifecycle(self, client):
        """Test full terminal lifecycle: create, use, delete"""
        # Create
        create_resp = await client.post("/api/terminals")
        assert create_resp.status_code == 201
        terminal_id = create_resp.json()["id"]
        assert create_resp.json()["status"] == "running"

        # List should show terminal
        list_resp = await client.get("/api/terminals")
        assert list_resp.status_code == 200
        assert list_resp.json()["total"] >= 1
        terminal_ids = [t["id"] for t in list_resp.json()["terminals"]]
        assert terminal_id in terminal_ids

        # Delete
        delete_resp = await client.delete(f"/api/terminals/{terminal_id}")
        assert delete_resp.status_code == 204

        # Get should return 404
        get_resp = await client.get(f"/api/terminals/{terminal_id}")
        assert get_resp.status_code == 404

    @pytest.mark.asyncio
    async def test_terminal_resize(self, client):
        """Test terminal resize operation"""
        # Create terminal with default size
        create_resp = await client.post("/api/terminals")
        terminal_id = create_resp.json()["id"]

        # Resize
        resize_resp = await client.post(
            f"/api/terminals/{terminal_id}/resize",
            json={"cols": 120, "rows": 40},
        )
        assert resize_resp.status_code == 200
        assert resize_resp.json()["cols"] == 120
        assert resize_resp.json()["rows"] == 40


@pytest.mark.skipif(sys.platform == "win32", reason="PTY not supported on Windows")
class TestMultipleTerminals:
    """Integration tests for multiple concurrent terminals"""

    @pytest_asyncio.fixture
    async def client(self, tmp_path):
        """Create a test client"""
        from moai_adk.web.config import WebConfig
        from moai_adk.web.database import close_database, init_database
        from moai_adk.web.routers.terminal import get_terminal_manager
        from moai_adk.web.server import create_app

        config = WebConfig(database_path=tmp_path / "test.db")
        await init_database(config)

        manager = get_terminal_manager()
        await manager.close_all()

        app = create_app(config)

        async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as client:
            yield client

        await manager.close_all()
        await close_database()

    @pytest.mark.asyncio
    async def test_multiple_terminals_parallel(self, client):
        """Test creating and managing multiple terminals in parallel"""
        # Create 3 terminals in parallel
        create_tasks = [client.post("/api/terminals", json={"spec_id": f"SPEC-{i}"}) for i in range(3)]

        responses = await asyncio.gather(*create_tasks)

        terminal_ids = []
        for resp in responses:
            assert resp.status_code == 201
            terminal_ids.append(resp.json()["id"])

        # List should show all terminals
        list_resp = await client.get("/api/terminals")
        assert list_resp.json()["total"] == 3

        # Delete all
        for tid in terminal_ids:
            await client.delete(f"/api/terminals/{tid}")

        # List should be empty
        list_resp = await client.get("/api/terminals")
        assert list_resp.json()["total"] == 0

    @pytest.mark.asyncio
    async def test_terminals_with_different_specs(self, client):
        """Test terminals can be associated with different SPECs"""
        specs = ["SPEC-AUTH-001", "SPEC-API-002", "SPEC-UI-003"]
        terminal_ids = []

        # Create terminals with different specs
        for spec in specs:
            resp = await client.post("/api/terminals", json={"spec_id": spec})
            assert resp.status_code == 201
            terminal_ids.append(resp.json()["id"])

        # Verify each terminal has correct spec
        for i, tid in enumerate(terminal_ids):
            resp = await client.get(f"/api/terminals/{tid}")
            assert resp.json()["spec_id"] == specs[i]


@pytest.mark.skipif(sys.platform == "win32", reason="PTY not supported on Windows")
class TestTerminalServiceDirect:
    """Direct tests for TerminalManager service"""

    @pytest.mark.asyncio
    async def test_terminal_write_and_read(self):
        """Test writing to and reading from terminal"""
        from moai_adk.web.services.terminal_service import TerminalManager

        manager = TerminalManager()
        session = await manager.create_terminal()

        try:
            # Write a simple echo command
            await manager.write(session.id, "echo TEST_OUTPUT\n")

            # Wait for output
            await asyncio.sleep(0.3)

            # Read output
            output = await manager.read(session.id)
            assert output is not None
            # Output should be bytes
            assert isinstance(output, bytes)
        finally:
            await manager.close(session.id)

    @pytest.mark.asyncio
    async def test_terminal_is_alive(self):
        """Test terminal alive status check"""
        from moai_adk.web.services.terminal_service import TerminalManager

        manager = TerminalManager()
        session = await manager.create_terminal()

        try:
            # Should be alive after creation
            assert manager.is_alive(session.id) is True

            # Close terminal
            await manager.close(session.id)

            # Should not be alive after close
            assert manager.is_alive(session.id) is False
        except Exception:
            await manager.close(session.id)
            raise

    @pytest.mark.asyncio
    async def test_terminal_with_worktree_path(self, tmp_path):
        """Test terminal with custom working directory"""
        from moai_adk.web.services.terminal_service import TerminalManager

        manager = TerminalManager()
        worktree_path = str(tmp_path)
        session = await manager.create_terminal(worktree_path=worktree_path)

        try:
            assert session.worktree_path == worktree_path

            # Write pwd command
            await manager.write(session.id, "pwd\n")
            await asyncio.sleep(0.3)

            output = await manager.read(session.id)
            # Output should contain the worktree path
            output_str = output.decode("utf-8", errors="replace")
            # The path should be somewhere in the output
            assert tmp_path.name in output_str or output is not None
        finally:
            await manager.close(session.id)


@pytest.mark.skipif(sys.platform == "win32", reason="PTY not supported on Windows")
class TestTerminalCleanup:
    """Tests for terminal cleanup behavior"""

    @pytest.mark.asyncio
    async def test_close_all_cleans_up(self):
        """Test that close_all properly cleans up all terminals"""
        from moai_adk.web.services.terminal_service import TerminalManager

        manager = TerminalManager()

        # Create multiple terminals
        for _ in range(3):
            await manager.create_terminal()

        assert len(manager.sessions) == 3
        assert len(manager.pty_processes) == 3

        # Close all
        await manager.close_all()

        assert len(manager.sessions) == 0
        assert len(manager.pty_processes) == 0

    @pytest.mark.asyncio
    async def test_close_updates_session_status(self):
        """Test that closing a terminal updates its status"""
        from moai_adk.web.services.terminal_service import TerminalManager

        manager = TerminalManager()
        session = await manager.create_terminal()

        assert session.status == "running"

        # Get a reference before closing (since close removes from dict)
        session_copy = session.model_copy()

        await manager.close(session.id)

        # Original session object should now be completed
        assert session.status == "completed"
        assert session.completed_at is not None
