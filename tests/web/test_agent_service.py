"""TAG-005: Test Claude Agent integration

RED Phase: Tests for verifying Claude Agent SDK wrapper service.
"""

import pytest


class TestAgentServiceExists:
    """Test cases for AgentService existence and structure"""

    @pytest.mark.asyncio
    async def test_agent_service_class_exists(self):
        """Test that AgentService class exists"""
        from moai_adk.web.services.agent_service import AgentService

        assert AgentService is not None

    @pytest.mark.asyncio
    async def test_agent_service_can_be_instantiated(self):
        """Test that AgentService can be instantiated"""
        from moai_adk.web.services.agent_service import AgentService

        service = AgentService()
        assert service is not None

    @pytest.mark.asyncio
    async def test_agent_service_has_stream_response_method(self):
        """Test that AgentService has stream_response method"""
        from moai_adk.web.services.agent_service import AgentService

        service = AgentService()
        assert hasattr(service, "stream_response")
        assert callable(service.stream_response)

    @pytest.mark.asyncio
    async def test_agent_service_has_send_message_method(self):
        """Test that AgentService has send_message method"""
        from moai_adk.web.services.agent_service import AgentService

        service = AgentService()
        assert hasattr(service, "send_message")
        assert callable(service.send_message)


class TestAgentServiceContextManagement:
    """Test cases for AgentService context management"""

    @pytest.mark.asyncio
    async def test_agent_service_has_context(self):
        """Test that AgentService maintains context"""
        from moai_adk.web.services.agent_service import AgentService

        service = AgentService()
        assert hasattr(service, "_context")

    @pytest.mark.asyncio
    async def test_get_context_returns_list(self):
        """Test that get_context returns a list"""
        from moai_adk.web.services.agent_service import AgentService

        service = AgentService()
        context = service.get_context("test-session")
        assert isinstance(context, list)

    @pytest.mark.asyncio
    async def test_add_to_context_stores_message(self):
        """Test that add_to_context stores a message"""
        from moai_adk.web.services.agent_service import AgentService

        service = AgentService()
        service.add_to_context("test-session", "user", "Hello")

        context = service.get_context("test-session")
        assert len(context) == 1
        assert context[0]["role"] == "user"
        assert context[0]["content"] == "Hello"

    @pytest.mark.asyncio
    async def test_clear_context_removes_messages(self):
        """Test that clear_context removes all messages"""
        from moai_adk.web.services.agent_service import AgentService

        service = AgentService()
        service.add_to_context("test-session", "user", "Hello")
        service.clear_context("test-session")

        context = service.get_context("test-session")
        assert len(context) == 0


class TestAgentServiceStreamResponse:
    """Test cases for AgentService stream_response method"""

    @pytest.mark.asyncio
    async def test_stream_response_is_async_generator(self):
        """Test that stream_response returns an async generator"""
        import inspect

        from moai_adk.web.services.agent_service import AgentService

        service = AgentService()
        result = service.stream_response(session_id="test", message="Hello")

        assert inspect.isasyncgen(result)

    @pytest.mark.asyncio
    async def test_stream_response_yields_strings(self):
        """Test that stream_response yields string chunks"""
        from moai_adk.web.services.agent_service import AgentService

        service = AgentService()
        chunks = []

        async for chunk in service.stream_response(session_id="test", message="Hello"):
            chunks.append(chunk)

        assert len(chunks) > 0
        assert all(isinstance(chunk, str) for chunk in chunks)

    @pytest.mark.asyncio
    async def test_stream_response_adds_to_context(self):
        """Test that stream_response adds messages to context"""
        from moai_adk.web.services.agent_service import AgentService

        service = AgentService()

        # Consume the generator
        async for _ in service.stream_response(session_id="test", message="Hello"):
            pass

        context = service.get_context("test")
        # Should have user message and assistant response
        assert len(context) == 2
        assert context[0]["role"] == "user"
        assert context[1]["role"] == "assistant"


class TestAgentServiceSendMessage:
    """Test cases for AgentService send_message method"""

    @pytest.mark.asyncio
    async def test_send_message_returns_string(self):
        """Test that send_message returns a string response"""
        from moai_adk.web.services.agent_service import AgentService

        service = AgentService()
        response = await service.send_message(session_id="test", message="Hello")

        assert isinstance(response, str)
        assert len(response) > 0

    @pytest.mark.asyncio
    async def test_send_message_adds_to_context(self):
        """Test that send_message adds messages to context"""
        from moai_adk.web.services.agent_service import AgentService

        service = AgentService()
        await service.send_message(session_id="test", message="Hello")

        context = service.get_context("test")
        assert len(context) == 2


class TestAgentServiceProviderIntegration:
    """Test cases for AgentService provider integration"""

    @pytest.mark.asyncio
    async def test_agent_service_uses_provider_service(self):
        """Test that AgentService uses ProviderService"""
        from moai_adk.web.services.agent_service import AgentService

        service = AgentService()
        assert hasattr(service, "_provider_service")

    @pytest.mark.asyncio
    async def test_agent_service_can_accept_custom_provider_service(self):
        """Test that AgentService can accept a custom ProviderService"""
        from moai_adk.web.services.agent_service import AgentService
        from moai_adk.web.services.provider_service import ProviderService

        custom_provider = ProviderService()
        service = AgentService(provider_service=custom_provider)

        assert service._provider_service is custom_provider
