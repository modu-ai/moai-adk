"""TAG-003: Test session management

RED Phase: Tests for verifying SQLite database and session CRUD operations.
Uses pytest-asyncio for async test execution.
"""

import pytest
import pytest_asyncio
from httpx import ASGITransport, AsyncClient


class TestDatabaseSetup:
    """Test cases for database initialization"""

    @pytest.mark.asyncio
    async def test_init_database_creates_connection(self, tmp_path):
        """Test that init_database creates a database connection"""
        from moai_adk.web.config import WebConfig
        from moai_adk.web.database import close_database, init_database

        config = WebConfig(database_path=tmp_path / "test.db")
        db = await init_database(config)

        assert db is not None
        await close_database()

    @pytest.mark.asyncio
    async def test_init_database_creates_sessions_table(self, tmp_path):
        """Test that init_database creates the sessions table"""
        from moai_adk.web.config import WebConfig
        from moai_adk.web.database import close_database, init_database

        config = WebConfig(database_path=tmp_path / "test.db")
        db = await init_database(config)

        cursor = await db.execute("SELECT name FROM sqlite_master WHERE type='table' AND name='sessions'")
        row = await cursor.fetchone()

        assert row is not None
        assert row[0] == "sessions"
        await close_database()

    @pytest.mark.asyncio
    async def test_init_database_creates_messages_table(self, tmp_path):
        """Test that init_database creates the messages table"""
        from moai_adk.web.config import WebConfig
        from moai_adk.web.database import close_database, init_database

        config = WebConfig(database_path=tmp_path / "test.db")
        db = await init_database(config)

        cursor = await db.execute("SELECT name FROM sqlite_master WHERE type='table' AND name='messages'")
        row = await cursor.fetchone()

        assert row is not None
        assert row[0] == "messages"
        await close_database()

    @pytest.mark.asyncio
    async def test_get_database_returns_connection(self, tmp_path):
        """Test that get_database returns the active connection"""
        from moai_adk.web.config import WebConfig
        from moai_adk.web.database import close_database, get_database, init_database

        config = WebConfig(database_path=tmp_path / "test.db")
        await init_database(config)

        db = await get_database()
        assert db is not None
        await close_database()

    @pytest.mark.asyncio
    async def test_get_database_raises_if_not_initialized(self):
        """Test that get_database raises RuntimeError if not initialized"""
        from moai_adk.web.database import close_database, get_database

        # Ensure database is closed
        await close_database()

        with pytest.raises(RuntimeError, match="Database not initialized"):
            await get_database()


class TestSessionsCRUD:
    """Test cases for session CRUD operations via API"""

    @pytest_asyncio.fixture
    async def client(self, tmp_path):
        """Create a test client with a temporary database"""
        from moai_adk.web.config import WebConfig
        from moai_adk.web.database import close_database, init_database
        from moai_adk.web.server import create_app

        config = WebConfig(database_path=tmp_path / "test.db")
        await init_database(config)

        app = create_app(config)

        async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as client:
            yield client

        await close_database()

    @pytest.mark.asyncio
    async def test_create_session_returns_201(self, client):
        """Test that POST /api/sessions returns 201 status"""
        response = await client.post("/api/sessions")
        assert response.status_code == 201

    @pytest.mark.asyncio
    async def test_create_session_returns_session_id(self, client):
        """Test that created session has an ID"""
        response = await client.post("/api/sessions")
        data = response.json()

        assert "id" in data
        assert len(data["id"]) > 0

    @pytest.mark.asyncio
    async def test_create_session_with_title(self, client):
        """Test that session can be created with custom title"""
        response = await client.post("/api/sessions", json={"title": "My Custom Session"})
        data = response.json()

        assert data["title"] == "My Custom Session"

    @pytest.mark.asyncio
    async def test_list_sessions_returns_200(self, client):
        """Test that GET /api/sessions returns 200 status"""
        response = await client.get("/api/sessions")
        assert response.status_code == 200

    @pytest.mark.asyncio
    async def test_list_sessions_returns_sessions_array(self, client):
        """Test that list sessions returns sessions array"""
        # Create a session first
        await client.post("/api/sessions")

        response = await client.get("/api/sessions")
        data = response.json()

        assert "sessions" in data
        assert "total" in data
        assert len(data["sessions"]) >= 1

    @pytest.mark.asyncio
    async def test_get_session_returns_200(self, client):
        """Test that GET /api/sessions/{id} returns 200 for existing session"""
        # Create a session first
        create_response = await client.post("/api/sessions")
        session_id = create_response.json()["id"]

        response = await client.get(f"/api/sessions/{session_id}")
        assert response.status_code == 200

    @pytest.mark.asyncio
    async def test_get_session_returns_404_for_unknown(self, client):
        """Test that GET /api/sessions/{id} returns 404 for unknown session"""
        response = await client.get("/api/sessions/unknown-id")
        assert response.status_code == 404

    @pytest.mark.asyncio
    async def test_update_session_title(self, client):
        """Test that PATCH /api/sessions/{id} updates the title"""
        # Create a session first
        create_response = await client.post("/api/sessions")
        session_id = create_response.json()["id"]

        # Update the title
        response = await client.patch(f"/api/sessions/{session_id}", json={"title": "Updated Title"})
        data = response.json()

        assert response.status_code == 200
        assert data["title"] == "Updated Title"

    @pytest.mark.asyncio
    async def test_delete_session_returns_204(self, client):
        """Test that DELETE /api/sessions/{id} returns 204"""
        # Create a session first
        create_response = await client.post("/api/sessions")
        session_id = create_response.json()["id"]

        response = await client.delete(f"/api/sessions/{session_id}")
        assert response.status_code == 204

    @pytest.mark.asyncio
    async def test_delete_session_removes_session(self, client):
        """Test that deleted session is no longer accessible"""
        # Create a session first
        create_response = await client.post("/api/sessions")
        session_id = create_response.json()["id"]

        # Delete the session
        await client.delete(f"/api/sessions/{session_id}")

        # Try to get the deleted session
        response = await client.get(f"/api/sessions/{session_id}")
        assert response.status_code == 404


class TestHealthEndpoint:
    """Test cases for health check endpoint"""

    @pytest_asyncio.fixture
    async def client(self, tmp_path):
        """Create a test client with a temporary database"""
        from moai_adk.web.config import WebConfig
        from moai_adk.web.database import close_database, init_database
        from moai_adk.web.server import create_app

        config = WebConfig(database_path=tmp_path / "test.db")
        await init_database(config)

        app = create_app(config)

        async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as client:
            yield client

        await close_database()

    @pytest.mark.asyncio
    async def test_health_returns_200(self, client):
        """Test that GET /api/health returns 200"""
        response = await client.get("/api/health")
        assert response.status_code == 200

    @pytest.mark.asyncio
    async def test_health_returns_healthy_status(self, client):
        """Test that health check returns healthy status"""
        response = await client.get("/api/health")
        data = response.json()

        assert data["status"] == "healthy"
        assert "version" in data
