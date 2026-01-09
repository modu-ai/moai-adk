"""Cost Tracking REST API

FastAPI router for Cost Tracking endpoints
providing usage recording and cost summaries.
"""

from datetime import date, datetime

from fastapi import APIRouter, HTTPException, Query
from pydantic import BaseModel

from moai_adk.web.models.model_config import ModelTier
from moai_adk.web.services.cost_tracker import CostTracker

router = APIRouter(prefix="/cost")


# Singleton instance for shared state
_cost_tracker: CostTracker | None = None


def get_cost_tracker() -> CostTracker:
    """Get the singleton CostTracker instance"""
    global _cost_tracker
    if _cost_tracker is None:
        _cost_tracker = CostTracker()
    return _cost_tracker


# Request/Response Models
class RecordUsageRequest(BaseModel):
    """Request body for recording token usage"""

    session_id: str
    model_id: str
    provider: str
    input_tokens: int
    output_tokens: int


class TokenUsageResponse(BaseModel):
    """Response body for token usage"""

    input_tokens: int
    output_tokens: int
    total_tokens: int


class CostRecordResponse(BaseModel):
    """Response body for cost record"""

    id: str
    session_id: str
    model_id: str
    provider: str
    usage: TokenUsageResponse
    cost_usd: float
    created_at: datetime


class CostSummaryResponse(BaseModel):
    """Response body for cost summary"""

    total_cost: float
    by_provider: dict[str, float]
    by_model: dict[str, float]
    token_counts: TokenUsageResponse


class CostEstimateResponse(BaseModel):
    """Response body for cost estimate"""

    tier: str
    estimated_tokens: int
    estimated_cost: float


@router.post("/record", response_model=CostRecordResponse, status_code=201)
async def record_usage(request: RecordUsageRequest) -> CostRecordResponse:
    """Record token usage and calculate cost

    Args:
        request: Token usage data

    Returns:
        Cost record with calculated cost
    """
    tracker = get_cost_tracker()
    record = tracker.record_usage(
        session_id=request.session_id,
        model_id=request.model_id,
        provider=request.provider,
        input_tokens=request.input_tokens,
        output_tokens=request.output_tokens,
    )

    return CostRecordResponse(
        id=record.id,
        session_id=record.session_id,
        model_id=record.model_id,
        provider=record.provider,
        usage=TokenUsageResponse(
            input_tokens=record.usage.input_tokens,
            output_tokens=record.usage.output_tokens,
            total_tokens=record.usage.total_tokens,
        ),
        cost_usd=record.cost_usd,
        created_at=record.created_at,
    )


@router.get("/session/{session_id}", response_model=CostSummaryResponse)
async def get_session_cost(session_id: str) -> CostSummaryResponse:
    """Get cost summary for a specific session

    Args:
        session_id: Session identifier

    Returns:
        Cost summary for the session
    """
    tracker = get_cost_tracker()
    summary = tracker.get_session_cost(session_id)

    return CostSummaryResponse(
        total_cost=summary.total_cost,
        by_provider=summary.by_provider,
        by_model=summary.by_model,
        token_counts=TokenUsageResponse(
            input_tokens=summary.token_counts.input_tokens,
            output_tokens=summary.token_counts.output_tokens,
            total_tokens=summary.token_counts.total_tokens,
        ),
    )


@router.get("/daily", response_model=CostSummaryResponse)
async def get_daily_cost(
    target_date: str | None = Query(None, alias="date"),
) -> CostSummaryResponse:
    """Get cost summary for a specific date

    Args:
        target_date: Date in YYYY-MM-DD format (defaults to today)

    Returns:
        Cost summary for the date
    """
    tracker = get_cost_tracker()

    if target_date:
        try:
            parsed_date = date.fromisoformat(target_date)
        except ValueError:
            raise HTTPException(
                status_code=400,
                detail=f"Invalid date format: {target_date}. Use YYYY-MM-DD.",
            )
    else:
        parsed_date = date.today()

    summary = tracker.get_daily_cost(parsed_date)

    return CostSummaryResponse(
        total_cost=summary.total_cost,
        by_provider=summary.by_provider,
        by_model=summary.by_model,
        token_counts=TokenUsageResponse(
            input_tokens=summary.token_counts.input_tokens,
            output_tokens=summary.token_counts.output_tokens,
            total_tokens=summary.token_counts.total_tokens,
        ),
    )


@router.get("/total", response_model=CostSummaryResponse)
async def get_total_cost() -> CostSummaryResponse:
    """Get total cost summary across all records

    Returns:
        Total cost summary
    """
    tracker = get_cost_tracker()
    summary = tracker.get_total_cost()

    return CostSummaryResponse(
        total_cost=summary.total_cost,
        by_provider=summary.by_provider,
        by_model=summary.by_model,
        token_counts=TokenUsageResponse(
            input_tokens=summary.token_counts.input_tokens,
            output_tokens=summary.token_counts.output_tokens,
            total_tokens=summary.token_counts.total_tokens,
        ),
    )


@router.get("/estimate", response_model=CostEstimateResponse)
async def estimate_cost(
    tier: str = Query(..., description="Model tier (planner or implementer)"),
    estimated_tokens: int = Query(..., description="Estimated total tokens"),
) -> CostEstimateResponse:
    """Estimate cost for a task based on tier

    Args:
        tier: Model tier to estimate for
        estimated_tokens: Estimated total tokens

    Returns:
        Cost estimate
    """
    tracker = get_cost_tracker()

    # Validate tier
    try:
        model_tier = ModelTier(tier)
    except ValueError:
        raise HTTPException(
            status_code=400,
            detail=f"Invalid tier: {tier}. Must be 'planner' or 'implementer'",
        )

    estimated_cost = tracker.estimate_cost(model_tier, estimated_tokens)

    return CostEstimateResponse(
        tier=tier,
        estimated_tokens=estimated_tokens,
        estimated_cost=estimated_cost,
    )
