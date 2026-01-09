"""Agent Service

Claude Agent SDK wrapper service providing
chat functionality and streaming response handling.
"""

from typing import AsyncGenerator, Optional

from moai_adk.web.services.provider_service import ProviderService


class AgentService:
    """Service for interacting with Claude Agent SDK

    Provides methods for sending messages and streaming responses
    from the AI agent.
    """

    def __init__(self, provider_service: Optional[ProviderService] = None):
        """Initialize the agent service

        Args:
            provider_service: Optional ProviderService instance
        """
        self._provider_service = provider_service or ProviderService()
        self._context: dict[str, list[dict]] = {}

    def get_context(self, session_id: str) -> list[dict]:
        """Get the conversation context for a session

        Args:
            session_id: The session ID

        Returns:
            List of context messages
        """
        return self._context.get(session_id, [])

    def add_to_context(self, session_id: str, role: str, content: str) -> None:
        """Add a message to the session context

        Args:
            session_id: The session ID
            role: Message role (user/assistant)
            content: Message content
        """
        if session_id not in self._context:
            self._context[session_id] = []

        self._context[session_id].append({"role": role, "content": content})

        # Keep context manageable (last 20 messages)
        if len(self._context[session_id]) > 20:
            self._context[session_id] = self._context[session_id][-20:]

    def clear_context(self, session_id: str) -> None:
        """Clear the context for a session

        Args:
            session_id: The session ID
        """
        self._context.pop(session_id, None)

    async def stream_response(
        self,
        session_id: str,
        message: str,
    ) -> AsyncGenerator[str, None]:
        """Stream a response from the AI agent

        Args:
            session_id: The session ID for context
            message: The user message

        Yields:
            Response chunks as they are generated
        """
        # Add user message to context
        self.add_to_context(session_id, "user", message)

        # Get provider configuration
        provider = self._provider_service.get_active_provider()
        model = self._provider_service.get_active_model()

        # For now, provide a simulated response
        # TODO: Integrate with actual Claude Agent SDK when available
        response = await self._generate_response(message, provider, model)

        # Stream response in chunks
        chunk_size = 50
        for i in range(0, len(response), chunk_size):
            chunk = response[i : i + chunk_size]
            yield chunk

        # Add assistant response to context
        self.add_to_context(session_id, "assistant", response)

    async def _generate_response(
        self,
        message: str,
        provider: str,
        model: str,
    ) -> str:
        """Generate a response using the AI provider

        Args:
            message: The user message
            provider: The AI provider name
            model: The model identifier

        Returns:
            The generated response
        """
        # Placeholder implementation
        # Will be replaced with actual Claude SDK integration
        return (
            f"[MoAI Web Backend - {provider}/{model}] "
            f"Received your message: '{message[:100]}...'. "
            "This is a placeholder response. "
            "Full Claude Agent SDK integration coming soon."
        )

    async def send_message(
        self,
        session_id: str,
        message: str,
    ) -> str:
        """Send a message and get a complete response

        Args:
            session_id: The session ID
            message: The user message

        Returns:
            The complete response
        """
        response_parts = []
        async for chunk in self.stream_response(session_id, message):
            response_parts.append(chunk)

        return "".join(response_parts)
