"""Providers Router

REST API endpoints for AI provider management
including provider switching and configuration.
"""

from fastapi import APIRouter, HTTPException, status
from pydantic import BaseModel, Field

from moai_adk.web.services.provider_service import ProviderService

router = APIRouter()


class ProviderInfo(BaseModel):
    """Provider information model"""

    name: str = Field(..., description="Provider name")
    models: list[str] = Field(default_factory=list, description="Available models")
    is_active: bool = Field(default=False, description="Whether provider is currently active")


class ProviderList(BaseModel):
    """Response model for listing providers"""

    providers: list[ProviderInfo] = Field(default_factory=list, description="List of available providers")
    active_provider: str = Field(..., description="Currently active provider")


class SwitchProviderRequest(BaseModel):
    """Request model for switching providers"""

    provider: str = Field(..., description="Provider name to switch to")
    model: str | None = Field(default=None, description="Optional model to use")


class SwitchProviderResponse(BaseModel):
    """Response model for provider switch"""

    success: bool = Field(..., description="Whether switch was successful")
    provider: str = Field(..., description="Current provider after switch")
    model: str = Field(..., description="Current model after switch")


@router.get("/providers", response_model=ProviderList)
async def list_providers() -> ProviderList:
    """List all available AI providers

    Returns:
        ProviderList with available providers and active provider
    """
    service = ProviderService()
    providers = service.get_available_providers()
    active = service.get_active_provider()

    provider_infos = [
        ProviderInfo(
            name=name,
            models=models,
            is_active=(name == active),
        )
        for name, models in providers.items()
    ]

    return ProviderList(providers=provider_infos, active_provider=active)


@router.post("/providers/switch", response_model=SwitchProviderResponse)
async def switch_provider(request: SwitchProviderRequest) -> SwitchProviderResponse:
    """Switch to a different AI provider

    Args:
        request: Provider switch request with target provider and optional model

    Returns:
        SwitchProviderResponse with switch result

    Raises:
        HTTPException: If provider or model is not available
    """
    service = ProviderService()

    # Validate provider exists
    providers = service.get_available_providers()
    if request.provider not in providers:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=f"Provider '{request.provider}' not available. Available providers: {list(providers.keys())}",
        )

    # Validate model if specified
    if request.model and request.model not in providers[request.provider]:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=f"Model '{request.model}' not available for provider '{request.provider}'. "
            f"Available models: {providers[request.provider]}",
        )

    # Switch provider
    success = service.switch_provider(request.provider, request.model)

    return SwitchProviderResponse(
        success=success,
        provider=service.get_active_provider(),
        model=service.get_active_model(),
    )


@router.get("/providers/current", response_model=SwitchProviderResponse)
async def get_current_provider() -> SwitchProviderResponse:
    """Get the current active provider and model

    Returns:
        SwitchProviderResponse with current provider and model
    """
    service = ProviderService()

    return SwitchProviderResponse(
        success=True,
        provider=service.get_active_provider(),
        model=service.get_active_model(),
    )
