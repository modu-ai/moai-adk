"""TAG-003 & TAG-004: Test Terminal REST API and WebSocket

RED Phase: Tests for terminal router endpoints including
REST API for terminal management and WebSocket for terminal I/O.
"""

import sys

import pytest
import pytest_asyncio
from httpx import ASGITransport, AsyncClient


class TestTerminalRouterStructure:
    """Test cases for terminal router module structure"""

    def test_terminal_router_exists(self):
        """Test that terminal router module can be imported"""
        from moai_adk.web.routers import terminal

        assert terminal is not None

    def test_terminal_router_has_router(self):
        """Test that terminal module has router object"""
        from moai_adk.web.routers.terminal import router

        assert router is not None


class TestTerminalRESTAPI:
    """Test cases for Terminal REST API endpoints"""

    @pytest_asyncio.fixture
    async def client(self, tmp_path):
        """Create a test client with terminal router"""
        from moai_adk.web.config import WebConfig
        from moai_adk.web.database import close_database, init_database
        from moai_adk.web.routers.terminal import get_terminal_manager
        from moai_adk.web.server import create_app

        config = WebConfig(database_path=tmp_path / "test.db")
        await init_database(config)

        # Clean up any existing terminals from previous tests
        manager = get_terminal_manager()
        await manager.close_all()

        app = create_app(config)

        async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as client:
            yield client

        # Cleanup after test
        await manager.close_all()
        await close_database()

    @pytest.mark.asyncio
    @pytest.mark.skipif(sys.platform == "win32", reason="PTY not supported on Windows")
    async def test_create_terminal(self, client):
        """Test POST /api/terminals creates a new terminal"""
        response = await client.post("/api/terminals")

        assert response.status_code == 201
        data = response.json()
        assert "id" in data
        assert data["status"] == "running"

    @pytest.mark.asyncio
    @pytest.mark.skipif(sys.platform == "win32", reason="PTY not supported on Windows")
    async def test_create_terminal_with_spec_id(self, client):
        """Test POST /api/terminals with spec_id"""
        response = await client.post(
            "/api/terminals",
            json={"spec_id": "SPEC-001"},
        )

        assert response.status_code == 201
        data = response.json()
        assert data["spec_id"] == "SPEC-001"

    @pytest.mark.asyncio
    @pytest.mark.skipif(sys.platform == "win32", reason="PTY not supported on Windows")
    async def test_create_terminal_with_dimensions(self, client):
        """Test POST /api/terminals with custom dimensions"""
        response = await client.post(
            "/api/terminals",
            json={"cols": 120, "rows": 40},
        )

        assert response.status_code == 201
        data = response.json()
        assert data["status"] == "running"

    @pytest.mark.asyncio
    @pytest.mark.skipif(sys.platform == "win32", reason="PTY not supported on Windows")
    async def test_list_terminals(self, client):
        """Test GET /api/terminals returns list"""
        # Create a terminal first
        await client.post("/api/terminals")

        response = await client.get("/api/terminals")

        assert response.status_code == 200
        data = response.json()
        assert "terminals" in data
        assert "total" in data
        assert data["total"] >= 1

    @pytest.mark.asyncio
    async def test_list_terminals_empty(self, client):
        """Test GET /api/terminals with no terminals"""
        response = await client.get("/api/terminals")

        assert response.status_code == 200
        data = response.json()
        assert data["terminals"] == []
        assert data["total"] == 0

    @pytest.mark.asyncio
    @pytest.mark.skipif(sys.platform == "win32", reason="PTY not supported on Windows")
    async def test_get_terminal(self, client):
        """Test GET /api/terminals/{terminal_id}"""
        # Create a terminal first
        create_response = await client.post("/api/terminals")
        terminal_id = create_response.json()["id"]

        response = await client.get(f"/api/terminals/{terminal_id}")

        assert response.status_code == 200
        data = response.json()
        assert data["id"] == terminal_id

    @pytest.mark.asyncio
    async def test_get_terminal_not_found(self, client):
        """Test GET /api/terminals/{terminal_id} with invalid ID"""
        response = await client.get("/api/terminals/nonexistent-id")

        assert response.status_code == 404

    @pytest.mark.asyncio
    @pytest.mark.skipif(sys.platform == "win32", reason="PTY not supported on Windows")
    async def test_delete_terminal(self, client):
        """Test DELETE /api/terminals/{terminal_id}"""
        # Create a terminal first
        create_response = await client.post("/api/terminals")
        terminal_id = create_response.json()["id"]

        response = await client.delete(f"/api/terminals/{terminal_id}")

        assert response.status_code == 204

        # Verify terminal is gone
        get_response = await client.get(f"/api/terminals/{terminal_id}")
        assert get_response.status_code == 404

    @pytest.mark.asyncio
    async def test_delete_terminal_not_found(self, client):
        """Test DELETE /api/terminals/{terminal_id} with invalid ID"""
        response = await client.delete("/api/terminals/nonexistent-id")

        assert response.status_code == 404

    @pytest.mark.asyncio
    @pytest.mark.skipif(sys.platform == "win32", reason="PTY not supported on Windows")
    async def test_resize_terminal(self, client):
        """Test POST /api/terminals/{terminal_id}/resize"""
        # Create a terminal first
        create_response = await client.post("/api/terminals")
        terminal_id = create_response.json()["id"]

        response = await client.post(
            f"/api/terminals/{terminal_id}/resize",
            json={"cols": 120, "rows": 40},
        )

        assert response.status_code == 200

    @pytest.mark.asyncio
    async def test_resize_terminal_not_found(self, client):
        """Test POST /api/terminals/{terminal_id}/resize with invalid ID"""
        response = await client.post(
            "/api/terminals/nonexistent-id/resize",
            json={"cols": 120, "rows": 40},
        )

        assert response.status_code == 404


class TestTerminalWebSocket:
    """Test cases for Terminal WebSocket endpoint"""

    def test_terminal_manager_singleton_exists(self):
        """Test that terminal_manager singleton is accessible"""
        from moai_adk.web.routers.terminal import get_terminal_manager

        manager = get_terminal_manager()
        assert manager is not None

    @pytest.mark.asyncio
    async def test_websocket_endpoint_registered(self):
        """Test that WebSocket endpoint exists in router"""
        from moai_adk.web.routers.terminal import router

        # Check that there's a WebSocket route
        ws_routes = [route for route in router.routes if hasattr(route, "path") and "ws" in str(route.path).lower()]
        # The WebSocket route should exist
        assert len(ws_routes) >= 1 or any("/terminal/" in str(route.path) for route in router.routes)


class TestTerminalMaxLimit:
    """Test cases for terminal maximum limit enforcement"""

    @pytest_asyncio.fixture
    async def client(self, tmp_path):
        """Create a test client with terminal router"""
        from moai_adk.web.config import WebConfig
        from moai_adk.web.database import close_database, init_database
        from moai_adk.web.routers.terminal import get_terminal_manager
        from moai_adk.web.server import create_app

        config = WebConfig(database_path=tmp_path / "test.db")
        await init_database(config)

        # Reset terminal manager for clean state
        manager = get_terminal_manager()
        await manager.close_all()

        app = create_app(config)

        async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as client:
            yield client

        # Cleanup
        await manager.close_all()
        await close_database()

    @pytest.mark.asyncio
    @pytest.mark.skipif(sys.platform == "win32", reason="PTY not supported on Windows")
    async def test_max_terminals_enforced(self, client):
        """Test that max terminal limit (6) is enforced"""
        from moai_adk.web.routers.terminal import get_terminal_manager

        manager = get_terminal_manager()

        # Create 6 terminals (the limit)
        for _ in range(6):
            response = await client.post("/api/terminals")
            assert response.status_code == 201

        # 7th terminal should fail
        response = await client.post("/api/terminals")
        assert response.status_code == 400
        assert "Maximum" in response.json()["detail"]

        # Cleanup
        await manager.close_all()
