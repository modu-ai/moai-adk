"""Agent Service

Claude Anthropic API wrapper service providing
chat functionality and streaming response handling.
"""

import os
import re
from typing import AsyncGenerator, Optional

from moai_adk.web.services.provider_service import ProviderService

# Import anthropic conditionally to handle missing dependency gracefully
try:
    import anthropic

    ANTHROPIC_AVAILABLE = True
except ImportError:
    ANTHROPIC_AVAILABLE = False


# MoAI command pattern detection
MOAI_COMMAND_PATTERN = re.compile(r"/moai:(\w+(?:-\w+)*)\s*(.*)", re.IGNORECASE)

# System prompt for MoAI Web Assistant
MOAI_SYSTEM_PROMPT = """You are the MoAI Web Assistant, an AI development companion.

Your capabilities include:
- Creating SPECs using /moai:1-plan command with --worktree option
- Helping users understand the Plan -> Run -> Sync workflow
- Providing guidance on TDD implementation

When a user wants to create a new feature or SPEC:
1. Ask clarifying questions about the feature requirements
2. Suggest using /moai:1-plan with --worktree option
3. Guide them through the SPEC creation process

When users mention /moai:* commands, recognize them:
- /moai:1-plan: Creates SPEC documents with EARS format
- /moai:2-run: Executes TDD implementation
- /moai:3-sync: Synchronizes documentation
- /moai:all-is-well: One-shot automation (Plan -> Run -> Sync)

Respond in the user's language. Be concise and helpful.
"""


class AgentService:
    """Service for interacting with Claude Anthropic API

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
        self._client: Optional["anthropic.Anthropic"] = None
        self._async_client: Optional["anthropic.AsyncAnthropic"] = None
        self._initialize_client()

    def _initialize_client(self) -> None:
        """Initialize Anthropic client with API key from environment"""
        if not ANTHROPIC_AVAILABLE:
            return

        api_key = os.environ.get("ANTHROPIC_API_KEY")
        if api_key:
            self._client = anthropic.Anthropic(api_key=api_key)
            self._async_client = anthropic.AsyncAnthropic(api_key=api_key)

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

    def detect_moai_command(self, message: str) -> Optional[dict]:
        """Detect MoAI commands in the message

        Args:
            message: The user message

        Returns:
            Dict with command info if detected, None otherwise
        """
        match = MOAI_COMMAND_PATTERN.search(message)
        if match:
            command = match.group(1)
            args = match.group(2).strip()
            return {
                "command": command,
                "args": args,
                "full_match": match.group(0),
            }
        return None

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

        # Check for MoAI command in message
        command_info = self.detect_moai_command(message)

        # Use Anthropic API if available, otherwise fallback to simulation
        if ANTHROPIC_AVAILABLE and self._async_client:
            response = ""
            async for chunk in self._stream_anthropic_response(session_id, message, provider, model, command_info):
                response += chunk
                yield chunk

            # Add assistant response to context
            self.add_to_context(session_id, "assistant", response)
        else:
            # Fallback to simulated response
            response = await self._generate_fallback_response(message, provider, model, command_info)

            # Stream response in chunks
            chunk_size = 50
            for i in range(0, len(response), chunk_size):
                chunk = response[i : i + chunk_size]
                yield chunk

            # Add assistant response to context
            self.add_to_context(session_id, "assistant", response)

    async def _stream_anthropic_response(
        self,
        session_id: str,
        message: str,
        provider: str,
        model: str,
        command_info: Optional[dict] = None,
    ) -> AsyncGenerator[str, None]:
        """Stream response from Anthropic API

        Args:
            session_id: The session ID
            message: The user message
            provider: The provider name
            model: The model identifier
            command_info: Detected MoAI command info

        Yields:
            Response text chunks
        """
        # Build messages from context
        messages = self.get_context(session_id)

        # Determine which model to use
        # Map provider model names to Anthropic model IDs
        model_mapping = {
            "opus": "claude-sonnet-4-20250514",  # Use claude-sonnet-4 for now as Opus 4.5 is expensive
            "sonnet": "claude-sonnet-4-20250514",
            "haiku": "claude-3-5-haiku-20241022",
        }
        anthropic_model = model_mapping.get(model, "claude-sonnet-4-20250514")

        # Add command context if detected
        enhanced_system = MOAI_SYSTEM_PROMPT
        if command_info:
            enhanced_system += f"\n\nThe user mentioned a MoAI command: /{command_info['command']}"
            if command_info.get("args"):
                enhanced_system += f" with arguments: {command_info['args']}"

        try:
            async with self._async_client.messages.stream(
                model=anthropic_model,
                max_tokens=4096,
                system=enhanced_system,
                messages=messages,
            ) as stream:
                async for text in stream.text_stream:
                    yield text
        except anthropic.APIError as e:
            yield f"\n\n[API Error: {e.message}]"
        except Exception as e:
            yield f"\n\n[Error: {str(e)}]"

    async def _generate_fallback_response(
        self,
        message: str,
        provider: str,
        model: str,
        command_info: Optional[dict] = None,
    ) -> str:
        """Generate a fallback response when API is unavailable

        Args:
            message: The user message
            provider: The AI provider name
            model: The model identifier
            command_info: Detected MoAI command info

        Returns:
            The generated response
        """
        if not ANTHROPIC_AVAILABLE:
            return (
                "[Anthropic SDK not installed]\n\n"
                "To enable chat functionality, install the web extras:\n"
                "```\npip install moai-adk[web]\n```\n\n"
                f"Your message: '{message[:100]}...'"
            )

        if not self._async_client:
            return (
                "[ANTHROPIC_API_KEY not set]\n\n"
                "To enable chat functionality, set your Anthropic API key:\n"
                "```\nexport ANTHROPIC_API_KEY=your_key_here\n```\n\n"
                f"Your message: '{message[:100]}...'"
            )

        # Command acknowledgment if detected
        if command_info:
            return (
                f"[Command Detected: /moai:{command_info['command']}]\n\n"
                f"I detected the MoAI command but API connection is unavailable.\n"
                f"Arguments: {command_info.get('args', 'none')}\n\n"
                f"Provider: {provider}/{model}"
            )

        return (
            f"[MoAI Web Backend - {provider}/{model}] "
            f"Received your message: '{message[:100]}...'. "
            "API connection unavailable."
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
