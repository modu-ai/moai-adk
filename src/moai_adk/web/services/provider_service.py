"""Provider Service

AI provider management service for switching between
different AI providers and models.
"""

from typing import Optional


class ProviderService:
    """Service for managing AI providers

    Provides methods for listing, switching, and configuring
    AI providers and their models.
    """

    # Available providers and their models
    PROVIDERS: dict[str, list[str]] = {
        "claude": [
            "claude-sonnet-4-20250514",
            "claude-opus-4-20250514",
            "claude-3-5-sonnet-20241022",
            "claude-3-5-haiku-20241022",
        ],
        "openai": [
            "gpt-4o",
            "gpt-4o-mini",
            "gpt-4-turbo",
            "gpt-3.5-turbo",
        ],
        "gemini": [
            "gemini-2.0-flash",
            "gemini-1.5-pro",
            "gemini-1.5-flash",
        ],
    }

    # Default provider and model
    DEFAULT_PROVIDER = "claude"
    DEFAULT_MODEL = "claude-sonnet-4-20250514"

    # Class-level state (shared across instances)
    _active_provider: str = DEFAULT_PROVIDER
    _active_model: str = DEFAULT_MODEL

    def __init__(self):
        """Initialize the provider service"""
        pass

    def get_available_providers(self) -> dict[str, list[str]]:
        """Get all available providers and their models

        Returns:
            Dictionary mapping provider names to lists of models
        """
        return self.PROVIDERS.copy()

    def get_active_provider(self) -> str:
        """Get the currently active provider

        Returns:
            The active provider name
        """
        return self._active_provider

    def get_active_model(self) -> str:
        """Get the currently active model

        Returns:
            The active model identifier
        """
        return self._active_model

    def switch_provider(
        self,
        provider: str,
        model: Optional[str] = None,
    ) -> bool:
        """Switch to a different provider

        Args:
            provider: The provider name to switch to
            model: Optional specific model to use

        Returns:
            True if switch was successful, False otherwise
        """
        if provider not in self.PROVIDERS:
            return False

        # Get model to use
        if model is None:
            # Use first available model for provider
            model = self.PROVIDERS[provider][0]
        elif model not in self.PROVIDERS[provider]:
            return False

        # Update class-level state
        ProviderService._active_provider = provider
        ProviderService._active_model = model

        return True

    def reset_to_default(self) -> None:
        """Reset to default provider and model"""
        ProviderService._active_provider = self.DEFAULT_PROVIDER
        ProviderService._active_model = self.DEFAULT_MODEL

    def is_provider_available(self, provider: str) -> bool:
        """Check if a provider is available

        Args:
            provider: The provider name to check

        Returns:
            True if provider is available
        """
        return provider in self.PROVIDERS

    def is_model_available(self, provider: str, model: str) -> bool:
        """Check if a model is available for a provider

        Args:
            provider: The provider name
            model: The model identifier

        Returns:
            True if model is available for the provider
        """
        return provider in self.PROVIDERS and model in self.PROVIDERS[provider]
