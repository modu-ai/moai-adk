"""TAG-004: Test WebSocket chat functionality

RED Phase: Tests for verifying WebSocket endpoint and message handling.
"""

import pytest
import pytest_asyncio
from httpx import ASGITransport, AsyncClient


class TestWebSocketConnection:
    """Test cases for WebSocket connection management"""

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

    @pytest_asyncio.fixture
    async def session_id(self, client):
        """Create a session and return its ID"""
        response = await client.post("/api/sessions")
        return response.json()["id"]

    @pytest.mark.asyncio
    async def test_connection_manager_exists(self):
        """Test that ConnectionManager class exists"""
        from moai_adk.web.routers.chat import ConnectionManager

        manager = ConnectionManager()
        assert manager is not None
        assert hasattr(manager, "active_connections")

    @pytest.mark.asyncio
    async def test_connection_manager_has_connect_method(self):
        """Test that ConnectionManager has connect method"""
        from moai_adk.web.routers.chat import ConnectionManager

        manager = ConnectionManager()
        assert hasattr(manager, "connect")
        assert callable(manager.connect)

    @pytest.mark.asyncio
    async def test_connection_manager_has_disconnect_method(self):
        """Test that ConnectionManager has disconnect method"""
        from moai_adk.web.routers.chat import ConnectionManager

        manager = ConnectionManager()
        assert hasattr(manager, "disconnect")
        assert callable(manager.disconnect)

    @pytest.mark.asyncio
    async def test_connection_manager_has_send_message_method(self):
        """Test that ConnectionManager has send_message method"""
        from moai_adk.web.routers.chat import ConnectionManager

        manager = ConnectionManager()
        assert hasattr(manager, "send_message")
        assert callable(manager.send_message)


class TestWebSocketMessageProtocol:
    """Test cases for WebSocket message protocol"""

    @pytest.mark.asyncio
    async def test_websocket_message_model_exists(self):
        """Test that WebSocketMessage model exists"""
        from moai_adk.web.models.message import WebSocketMessage

        assert WebSocketMessage is not None

    @pytest.mark.asyncio
    async def test_websocket_message_has_type_field(self):
        """Test that WebSocketMessage has type field"""
        from moai_adk.web.models.message import WebSocketMessage

        msg = WebSocketMessage(type="chat", content="Hello")
        assert hasattr(msg, "type")
        assert msg.type == "chat"

    @pytest.mark.asyncio
    async def test_websocket_message_has_content_field(self):
        """Test that WebSocketMessage has content field"""
        from moai_adk.web.models.message import WebSocketMessage

        msg = WebSocketMessage(type="chat", content="Hello World")
        assert hasattr(msg, "content")
        assert msg.content == "Hello World"

    @pytest.mark.asyncio
    async def test_websocket_message_has_timestamp_field(self):
        """Test that WebSocketMessage has timestamp field"""
        from datetime import datetime

        from moai_adk.web.models.message import WebSocketMessage

        now = datetime.now()
        msg = WebSocketMessage(type="chat", content="Hello", timestamp=now)
        assert hasattr(msg, "timestamp")
        assert msg.timestamp == now

    @pytest.mark.asyncio
    async def test_websocket_message_has_session_id_field(self):
        """Test that WebSocketMessage has session_id field"""
        from moai_adk.web.models.message import WebSocketMessage

        msg = WebSocketMessage(type="chat", content="Hello", session_id="test-123")
        assert hasattr(msg, "session_id")
        assert msg.session_id == "test-123"

    @pytest.mark.asyncio
    async def test_websocket_message_serialization(self):
        """Test that WebSocketMessage can be serialized to JSON"""
        from moai_adk.web.models.message import WebSocketMessage

        msg = WebSocketMessage(type="chat", content="Hello")
        json_data = msg.model_dump(mode="json")

        assert "type" in json_data
        assert "content" in json_data
        assert json_data["type"] == "chat"
        assert json_data["content"] == "Hello"


class TestSaveMessageFunction:
    """Test cases for save_message function"""

    @pytest_asyncio.fixture
    async def db_setup(self, tmp_path):
        """Setup database for testing"""
        from moai_adk.web.config import WebConfig
        from moai_adk.web.database import close_database, get_database, init_database

        config = WebConfig(database_path=tmp_path / "test.db")
        await init_database(config)

        # Create a test session
        db = await get_database()
        await db.execute(
            "INSERT INTO sessions (id, title) VALUES (?, ?)",
            ("test-session-id", "Test Session"),
        )
        await db.commit()

        yield

        await close_database()

    @pytest.mark.asyncio
    async def test_save_message_function_exists(self):
        """Test that save_message function exists"""
        from moai_adk.web.routers.chat import save_message

        assert callable(save_message)

    @pytest.mark.asyncio
    async def test_save_message_returns_message_id(self, db_setup):
        """Test that save_message returns a message ID"""
        from moai_adk.web.models.message import MessageRole
        from moai_adk.web.routers.chat import save_message

        message_id = await save_message(
            session_id="test-session-id",
            role=MessageRole.USER,
            content="Test message",
        )

        assert message_id is not None
        assert len(message_id) > 0

    @pytest.mark.asyncio
    async def test_save_message_stores_in_database(self, db_setup):
        """Test that save_message stores message in database"""
        from moai_adk.web.database import get_database
        from moai_adk.web.models.message import MessageRole
        from moai_adk.web.routers.chat import save_message

        message_id = await save_message(
            session_id="test-session-id",
            role=MessageRole.USER,
            content="Stored message",
        )

        db = await get_database()
        cursor = await db.execute("SELECT id, content, role FROM messages WHERE id = ?", (message_id,))
        row = await cursor.fetchone()

        assert row is not None
        assert row["content"] == "Stored message"
        assert row["role"] == "user"
