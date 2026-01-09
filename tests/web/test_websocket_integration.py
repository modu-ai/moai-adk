"""Test WebSocket integration

Integration tests for WebSocket chat functionality.
"""

import pytest
import pytest_asyncio


class TestConnectionManager:
    """Test ConnectionManager class directly"""

    @pytest.mark.asyncio
    async def test_disconnect_removes_connection(self):
        """Test that disconnect removes the connection"""
        from moai_adk.web.routers.chat import ConnectionManager

        manager = ConnectionManager()

        # Add a fake connection
        manager.active_connections["test-session"] = "fake-websocket"

        # Disconnect
        manager.disconnect("test-session")

        assert "test-session" not in manager.active_connections

    @pytest.mark.asyncio
    async def test_disconnect_nonexistent_session(self):
        """Test disconnect with non-existent session doesn't raise"""
        from moai_adk.web.routers.chat import ConnectionManager

        manager = ConnectionManager()

        # Should not raise
        manager.disconnect("nonexistent-session")

    @pytest.mark.asyncio
    async def test_get_connection_returns_none_for_unknown(self):
        """Test get_connection returns None for unknown session"""
        from moai_adk.web.routers.chat import ConnectionManager

        manager = ConnectionManager()

        result = manager.get_connection("unknown-session")
        assert result is None

    @pytest.mark.asyncio
    async def test_get_connection_returns_websocket(self):
        """Test get_connection returns the WebSocket"""
        from moai_adk.web.routers.chat import ConnectionManager

        manager = ConnectionManager()
        manager.active_connections["test-session"] = "fake-websocket"

        result = manager.get_connection("test-session")
        assert result == "fake-websocket"


class TestChatFunctions:
    """Test chat module functions"""

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
            ("chat-test-session", "Chat Test Session"),
        )
        await db.commit()

        yield

        await close_database()

    @pytest.mark.asyncio
    async def test_save_message_user_role(self, db_setup):
        """Test saving a user message"""
        from moai_adk.web.database import get_database
        from moai_adk.web.models.message import MessageRole
        from moai_adk.web.routers.chat import save_message

        message_id = await save_message(
            session_id="chat-test-session",
            role=MessageRole.USER,
            content="User test message",
        )

        db = await get_database()
        cursor = await db.execute("SELECT role, content FROM messages WHERE id = ?", (message_id,))
        row = await cursor.fetchone()

        assert row["role"] == "user"
        assert row["content"] == "User test message"

    @pytest.mark.asyncio
    async def test_save_message_assistant_role(self, db_setup):
        """Test saving an assistant message"""
        from moai_adk.web.database import get_database
        from moai_adk.web.models.message import MessageRole
        from moai_adk.web.routers.chat import save_message

        message_id = await save_message(
            session_id="chat-test-session",
            role=MessageRole.ASSISTANT,
            content="Assistant test message",
        )

        db = await get_database()
        cursor = await db.execute("SELECT role, content FROM messages WHERE id = ?", (message_id,))
        row = await cursor.fetchone()

        assert row["role"] == "assistant"
        assert row["content"] == "Assistant test message"

    @pytest.mark.asyncio
    async def test_save_message_updates_session_timestamp(self, db_setup):
        """Test that saving a message updates session timestamp"""
        import asyncio

        from moai_adk.web.database import get_database
        from moai_adk.web.models.message import MessageRole
        from moai_adk.web.routers.chat import save_message

        db = await get_database()

        # Get initial session timestamp
        cursor = await db.execute("SELECT updated_at FROM sessions WHERE id = ?", ("chat-test-session",))
        initial = await cursor.fetchone()
        initial_time = initial["updated_at"]

        # Wait a bit
        await asyncio.sleep(0.1)

        # Save a message
        await save_message(
            session_id="chat-test-session",
            role=MessageRole.USER,
            content="Update timestamp test",
        )

        # Check updated timestamp
        cursor = await db.execute("SELECT updated_at FROM sessions WHERE id = ?", ("chat-test-session",))
        updated = await cursor.fetchone()
        updated_time = updated["updated_at"]

        assert updated_time != initial_time

    @pytest.mark.asyncio
    async def test_utcnow_returns_datetime(self):
        """Test _utcnow helper returns datetime"""
        from datetime import datetime

        from moai_adk.web.routers.chat import _utcnow

        result = _utcnow()
        assert isinstance(result, datetime)


class TestProviderServiceAdditional:
    """Additional tests for ProviderService"""

    @pytest.mark.asyncio
    async def test_is_provider_available_true(self):
        """Test is_provider_available returns True for valid provider"""
        from moai_adk.web.services.provider_service import ProviderService

        service = ProviderService()
        assert service.is_provider_available("claude") is True

    @pytest.mark.asyncio
    async def test_is_provider_available_false(self):
        """Test is_provider_available returns False for invalid provider"""
        from moai_adk.web.services.provider_service import ProviderService

        service = ProviderService()
        assert service.is_provider_available("invalid") is False

    @pytest.mark.asyncio
    async def test_is_model_available_true(self):
        """Test is_model_available returns True for valid model"""
        from moai_adk.web.services.provider_service import ProviderService

        service = ProviderService()
        assert service.is_model_available("claude", "claude-sonnet-4-20250514") is True

    @pytest.mark.asyncio
    async def test_is_model_available_false_wrong_provider(self):
        """Test is_model_available returns False for wrong provider"""
        from moai_adk.web.services.provider_service import ProviderService

        service = ProviderService()
        assert service.is_model_available("invalid", "claude-sonnet-4-20250514") is False

    @pytest.mark.asyncio
    async def test_is_model_available_false_wrong_model(self):
        """Test is_model_available returns False for wrong model"""
        from moai_adk.web.services.provider_service import ProviderService

        service = ProviderService()
        assert service.is_model_available("claude", "invalid-model") is False

    @pytest.mark.asyncio
    async def test_switch_provider_with_specific_model(self):
        """Test switching provider with a specific model"""
        from moai_adk.web.services.provider_service import ProviderService

        service = ProviderService()
        result = service.switch_provider("openai", "gpt-4o")

        assert result is True
        assert service.get_active_provider() == "openai"
        assert service.get_active_model() == "gpt-4o"

        # Reset
        service.reset_to_default()

    @pytest.mark.asyncio
    async def test_switch_provider_with_invalid_model(self):
        """Test switching provider with invalid model returns False"""
        from moai_adk.web.services.provider_service import ProviderService

        service = ProviderService()
        result = service.switch_provider("claude", "invalid-model")

        assert result is False


class TestAgentServiceAdditional:
    """Additional tests for AgentService"""

    @pytest.mark.asyncio
    async def test_context_limit_enforced(self):
        """Test that context is limited to 20 messages"""
        from moai_adk.web.services.agent_service import AgentService

        service = AgentService()

        # Add 25 messages
        for i in range(25):
            service.add_to_context("test-session", "user", f"Message {i}")

        context = service.get_context("test-session")
        assert len(context) == 20
        # Should have the last 20 messages
        assert context[0]["content"] == "Message 5"
        assert context[-1]["content"] == "Message 24"
