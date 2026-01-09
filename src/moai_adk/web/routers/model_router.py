"""Model Router REST API

FastAPI router for Multi-Model Router endpoints
providing task classification and tier switching.
"""

from fastapi import APIRouter, HTTPException
from pydantic import BaseModel

from moai_adk.web.models.model_config import ModelTier
from moai_adk.web.services.model_router import ModelRouter

router = APIRouter(prefix="/router")


# Singleton instance for shared state
_model_router: ModelRouter | None = None


def get_model_router() -> ModelRouter:
    """Get the singleton ModelRouter instance"""
    global _model_router
    if _model_router is None:
        _model_router = ModelRouter()
    return _model_router


# Request/Response Models
class ClassifyTaskRequest(BaseModel):
    """Request body for task classification"""

    task_description: str


class ClassifyTaskResponse(BaseModel):
    """Response body for task classification"""

    task_type: str
    recommended_tier: str
    reason: str


class SwitchTierRequest(BaseModel):
    """Request body for tier switching"""

    tier: str


class SwitchTierResponse(BaseModel):
    """Response body for tier switching"""

    success: bool
    active_tier: str
    message: str


class ModelConfigResponse(BaseModel):
    """Response body for model configuration"""

    tier: str
    model_id: str
    provider: str
    cost_per_1k_input: float
    cost_per_1k_output: float
    base_url: str | None


class RouterConfigResponse(BaseModel):
    """Response body for router configuration"""

    active_tier: str
    models: dict[str, ModelConfigResponse]


class EnvironmentResponse(BaseModel):
    """Response body for environment variables"""

    tier: str
    environment: dict[str, str]


@router.post("/classify", response_model=ClassifyTaskResponse)
async def classify_task(request: ClassifyTaskRequest) -> ClassifyTaskResponse:
    """Classify a task and get recommended model tier

    Args:
        request: Task description to classify

    Returns:
        Classification result with recommended tier
    """
    model_router = get_model_router()
    classification = model_router.classify_task(request.task_description)

    return ClassifyTaskResponse(
        task_type=classification.task_type,
        recommended_tier=classification.recommended_tier.value,
        reason=classification.reason,
    )


@router.get("/config", response_model=RouterConfigResponse)
async def get_config() -> RouterConfigResponse:
    """Get current model router configuration

    Returns:
        Router configuration with all model tiers
    """
    model_router = get_model_router()

    models = {}
    for tier, config in model_router.MODELS.items():
        models[tier.value] = ModelConfigResponse(
            tier=config.tier.value,
            model_id=config.model_id,
            provider=config.provider,
            cost_per_1k_input=config.cost_per_1k_input,
            cost_per_1k_output=config.cost_per_1k_output,
            base_url=config.base_url,
        )

    return RouterConfigResponse(
        active_tier=model_router.get_active_tier().value,
        models=models,
    )


@router.post("/switch", response_model=SwitchTierResponse)
async def switch_tier(request: SwitchTierRequest) -> SwitchTierResponse:
    """Switch to a specific model tier

    Args:
        request: Tier to switch to

    Returns:
        Switch result with new active tier
    """
    model_router = get_model_router()

    # Validate tier
    try:
        tier = ModelTier(request.tier)
    except ValueError:
        raise HTTPException(
            status_code=400,
            detail=f"Invalid tier: {request.tier}. Must be 'planner' or 'implementer'",
        )

    model_router.switch_to_tier(tier)

    return SwitchTierResponse(
        success=True,
        active_tier=tier.value,
        message=f"Successfully switched to {tier.value} tier",
    )


@router.get("/environment/{tier}", response_model=EnvironmentResponse)
async def get_environment(tier: str) -> EnvironmentResponse:
    """Get environment variables for a specific tier

    Args:
        tier: The tier to get environment for

    Returns:
        Environment variables for the tier
    """
    model_router = get_model_router()

    # Validate tier
    try:
        model_tier = ModelTier(tier)
    except ValueError:
        raise HTTPException(
            status_code=400,
            detail=f"Invalid tier: {tier}. Must be 'planner' or 'implementer'",
        )

    config = model_router.get_model_for_tier(model_tier)
    env = model_router.get_environment_for_model(config)

    return EnvironmentResponse(
        tier=tier,
        environment=env,
    )
