"""Health Check Router

Provides health check endpoint for monitoring
the status of the MoAI Web Backend.
"""

from fastapi import APIRouter
from pydantic import BaseModel

router = APIRouter()


class HealthResponse(BaseModel):
    """Health check response model"""

    status: str
    version: str


@router.get("/health", response_model=HealthResponse)
async def health_check() -> HealthResponse:
    """Check the health status of the API

    Returns:
        HealthResponse with status and version information
    """
    return HealthResponse(status="healthy", version="0.1.0")
